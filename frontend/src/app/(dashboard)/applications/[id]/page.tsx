"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  Calendar,
  Clock,
  ExternalLink,
  Loader2,
  RefreshCw,
  Save,
  Sparkles,
  Trash2,
  X,
} from "lucide-react";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { ApplicationStatusSelect } from "@/components/applications/application-status-select";
import { StatusBadge } from "@/components/applications/status-badge";
import { api, API_BASE_URL, ApiClientError } from "@/lib/api";
import { streamGenerateEmail } from "@/lib/generate-email-stream";
import type { Application } from "@/types/application";
import type { Resume } from "@/types/resume";
import type { Email, EmailTone } from "@/types/email";

const toneOptions: { value: EmailTone; label: string }[] = [
  { value: "formal", label: "Formal" },
  { value: "friendly", label: "Friendly" },
  { value: "concise", label: "Concise" },
];

export default function ApplicationDetailPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const [application, setApplication] = useState<Application | null>(null);
  const [email, setEmail] = useState<Email | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    try {
      const appRes = await api.getApplication(params.id);
      setApplication(appRes.application);

      try {
        const emailRes = await api.getEmail(params.id);
        setEmail(emailRes.email);
      } catch {
        /* no email yet */
      }
    } catch {
      setError("Failed to load application.");
    } finally {
      setIsLoading(false);
    }
  }, [params.id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  async function handleDelete() {
    if (!confirm("Delete this application?")) return;
    try {
      await api.deleteApplication(params.id);
      router.push("/applications");
    } catch {
      setError("Failed to delete application.");
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
      </div>
    );
  }

  if (error || !application) {
    return (
      <div className="space-y-4">
        <div className="rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error || "Application not found."}
        </div>
        <Link
          href="/applications"
          className={buttonVariants({ variant: "outline" })}
        >
          <ArrowLeft className="mr-2 size-4" />
          Back to applications
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <Link
            href="/applications"
            className={buttonVariants({ variant: "ghost", size: "icon" })}
          >
            <ArrowLeft className="size-4" />
          </Link>
          <div>
            <div className="flex flex-wrap items-center gap-x-2 gap-y-2">
              <h1 className="text-2xl font-semibold tracking-tight">
                {application.company_name}
              </h1>
              <StatusBadge status={application.status} />
              <ApplicationStatusSelect
                applicationId={application.id}
                status={application.status}
                onUpdated={setApplication}
              />
            </div>
            <p className="text-sm text-muted-foreground">{application.role}</p>
          </div>
        </div>
        <Button
          variant="ghost"
          size="icon"
          onClick={handleDelete}
          className="shrink-0 text-muted-foreground hover:text-destructive"
        >
          <Trash2 className="size-4" />
        </Button>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="space-y-4 rounded-xl border bg-card p-5">
          <h2 className="text-sm font-semibold">Application Details</h2>
          <div className="grid gap-3 text-sm">
            <div>
              <span className="text-muted-foreground">Recruiter Email</span>
              <p className="font-medium">{application.recruiter_email}</p>
            </div>
            {application.job_link && (
              <div>
                <span className="text-muted-foreground">Job Link</span>
                <p>
                  <a
                    href={application.job_link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 font-medium text-primary underline-offset-2 hover:underline"
                  >
                    View posting
                    <ExternalLink className="size-3" />
                  </a>
                </p>
              </div>
            )}
            <div>
              <span className="text-muted-foreground">Job Description</span>
              <p className="mt-1 max-h-60 overflow-y-auto whitespace-pre-wrap rounded-lg bg-muted/50 p-3 text-xs leading-relaxed">
                {application.job_description}
              </p>
            </div>
          </div>
        </div>

        <EmailPanel
          application={application}
          email={email}
          onEmailUpdate={setEmail}
        />
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Email Panel
// ---------------------------------------------------------------------------

type StreamPhase =
  | "idle"
  | "connecting"
  | "streaming"
  | "persisting"
  | "error";

function EmailPanel({
  application,
  email,
  onEmailUpdate,
}: {
  application: Application;
  email: Email | null;
  onEmailUpdate: (email: Email) => void;
}) {
  const [tone, setTone] = useState<EmailTone>("formal");
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [resumesLoading, setResumesLoading] = useState(true);
  const [streamPhase, setStreamPhase] = useState<StreamPhase>("idle");
  const [streamPreview, setStreamPreview] = useState("");
  const [isFallbackGenerating, setIsFallbackGenerating] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [subject, setSubject] = useState(email?.subject ?? "");
  const [body, setBody] = useState(email?.body ?? "");
  const [error, setError] = useState("");
  const [saveSuccess, setSaveSuccess] = useState(false);
  const streamAbortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function loadResumes() {
      try {
        const res = await api.listResumes();
        if (!cancelled) {
          setResumes(res.resumes);
        }
      } catch {
        /* non-blocking */
      } finally {
        if (!cancelled) {
          setResumesLoading(false);
        }
      }
    }
    loadResumes();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    return () => {
      streamAbortRef.current?.abort();
    };
  }, []);

  useEffect(() => {
    if (email) {
      setSubject(email.subject);
      setBody(email.body);
    }
  }, [email]);

  const isStreamBusy =
    streamPhase === "connecting" ||
    streamPhase === "streaming" ||
    streamPhase === "persisting";
  const isBusy = isStreamBusy || isFallbackGenerating;

  function cancelStream() {
    streamAbortRef.current?.abort();
    streamAbortRef.current = null;
    setStreamPhase("idle");
    setStreamPreview("");
  }

  async function handleGenerateStream() {
    setError("");
    setStreamPreview("");
    setStreamPhase("connecting");

    const resume = resumes.find((r) => r.id === application.resume_id);
    const parsed = resume?.parsed_text?.trim();
    if (!application.resume_id || !parsed) {
      setStreamPhase("error");
      setError(
        "This application needs a linked resume with parsed text. Edit the application or upload a resume."
      );
      return;
    }

    const abort = new AbortController();
    streamAbortRef.current = abort;

    try {
      const draft = await streamGenerateEmail(
        API_BASE_URL,
        {
          resume_text: parsed,
          job_description: application.job_description,
          company_name: application.company_name,
          role: application.role,
          recruiter_email: application.recruiter_email ?? "",
          job_link: application.job_link ?? "",
          tone,
        },
        {
          onPhase: (p) => {
            if (p === "connecting") {
              setStreamPhase("connecting");
            }
            if (p === "streaming") {
              setStreamPhase("streaming");
            }
          },
          onDelta: (acc) => setStreamPreview(acc),
          signal: abort.signal,
        }
      );

      setStreamPhase("persisting");
      const res = email
        ? await api.regenerateEmail(application.id, tone, draft)
        : await api.generateEmail(application.id, tone, draft);
      onEmailUpdate(res.email);
      setStreamPreview("");
      setStreamPhase("idle");
    } catch (err: unknown) {
      if (err instanceof Error && err.name === "AbortError") {
        setStreamPhase("idle");
        setStreamPreview("");
        return;
      }
      setStreamPhase("error");
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Streaming failed.");
      }
    } finally {
      streamAbortRef.current = null;
    }
  }

  async function handleGenerateFallback() {
    setIsFallbackGenerating(true);
    setError("");
    setStreamPreview("");
    setStreamPhase("idle");
    try {
      const res = email
        ? await api.regenerateEmail(application.id, tone)
        : await api.generateEmail(application.id, tone);
      onEmailUpdate(res.email);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError("Failed to generate email.");
      }
    } finally {
      setIsFallbackGenerating(false);
    }
  }

  async function handleSave() {
    if (!email) return;
    setIsSaving(true);
    setError("");
    setSaveSuccess(false);
    try {
      const res = await api.updateEmail(email.id, { subject, body });
      onEmailUpdate(res.email);
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 2000);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError("Failed to save email.");
      }
    } finally {
      setIsSaving(false);
    }
  }

  const hasChanges =
    email && (subject !== email.subject || body !== email.body);

  const isEditable = email && (email.status === "draft" || email.status === "scheduled");

  return (
    <div className="space-y-4 rounded-xl border bg-card p-5">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold">Outreach Email</h2>
        {email && <EmailStatusBadge status={email.status} />}
      </div>

      {error && (
        <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      )}

      {(streamPhase !== "idle" || streamPreview) && (
        <div
          className="space-y-2 rounded-lg border bg-muted/30 p-3"
          aria-live="polite"
        >
          <p className="text-xs text-muted-foreground">
            {streamPhase === "connecting" && "Connecting to AI…"}
            {streamPhase === "streaming" &&
              "Streaming model output (live preview — JSON is parsed when complete)…"}
            {streamPhase === "persisting" && "Saving draft to your application…"}
            {streamPhase === "error" && "Streaming failed. Retry or use non-streaming generation."}
          </p>
          {(streamPhase === "streaming" ||
            streamPhase === "error" ||
            (streamPreview && streamPhase !== "persisting")) && (
            <Textarea
              readOnly
              value={streamPreview}
              className="min-h-28 font-mono text-xs"
              aria-label="Streaming draft preview"
            />
          )}
          {streamPhase === "persisting" && (
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <Loader2 className="size-4 animate-spin" />
              Saving…
            </div>
          )}
        </div>
      )}

      {/* Tone selector + generate button */}
      {(!email || isEditable) && (
        <div className="flex flex-col gap-3">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
            <div className="grid flex-1 gap-1.5">
              <Label htmlFor="tone" className="text-xs">
                Tone
              </Label>
              <select
                id="tone"
                value={tone}
                onChange={(e) => setTone(e.target.value as EmailTone)}
                disabled={isBusy || resumesLoading}
                className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-50 dark:bg-input/30"
              >
                {toneOptions.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Button
                onClick={handleGenerateStream}
                disabled={isBusy || resumesLoading}
                className="shrink-0"
              >
                {streamPhase === "connecting" || streamPhase === "streaming" ? (
                  <>
                    <Loader2 className="mr-2 size-4 animate-spin" />
                    {streamPhase === "connecting" ? "Connecting…" : "Streaming…"}
                  </>
                ) : streamPhase === "persisting" ? (
                  <>
                    <Loader2 className="mr-2 size-4 animate-spin" />
                    Saving…
                  </>
                ) : email ? (
                  <>
                    <RefreshCw className="mr-2 size-4" />
                    Regenerate
                  </>
                ) : (
                  <>
                    <Sparkles className="mr-2 size-4" />
                    Generate Email
                  </>
                )}
              </Button>
              {(streamPhase === "connecting" ||
                streamPhase === "streaming") && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={cancelStream}
                >
                  Cancel
                </Button>
              )}
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleGenerateFallback}
              disabled={isBusy || resumesLoading}
            >
              {isFallbackGenerating ? (
                <>
                  <Loader2 className="mr-2 size-3 animate-spin" />
                  Generating…
                </>
              ) : (
                "Generate without streaming"
              )}
            </Button>
            {streamPhase === "error" && (
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={handleGenerateStream}
                disabled={isBusy || resumesLoading}
              >
                Retry streaming
              </Button>
            )}
          </div>
          {resumesLoading && (
            <p className="text-xs text-muted-foreground">Loading resumes…</p>
          )}
        </div>
      )}

      {/* Match score + key points */}
      {email && email.match_score != null && (
        <div className="rounded-lg bg-muted/50 p-3">
          <div className="flex items-center gap-2 text-xs font-medium">
            <span>Match Score</span>
            <Badge variant="outline" className="text-xs">
              {Math.round(email.match_score * 100)}%
            </Badge>
          </div>
          {email.key_points && email.key_points.length > 0 && (
            <ul className="mt-2 space-y-1 text-xs text-muted-foreground">
              {email.key_points.map((point, i) => (
                <li key={i} className="flex gap-1.5">
                  <span className="text-foreground">•</span>
                  {point}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      {/* Email editor */}
      {email && (
        <div className="space-y-3">
          <div className="grid gap-1.5">
            <Label htmlFor="email-subject" className="text-xs">
              Subject
            </Label>
            <Input
              id="email-subject"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              disabled={!isEditable}
            />
          </div>
          <div className="grid gap-1.5">
            <Label htmlFor="email-body" className="text-xs">
              Body
            </Label>
            <Textarea
              id="email-body"
              value={body}
              onChange={(e) => setBody(e.target.value)}
              className="min-h-52"
              disabled={!isEditable}
            />
          </div>

          {isEditable && (
            <div className="flex items-center justify-between">
              <div>
                {saveSuccess && (
                  <p className="text-xs text-muted-foreground">Saved!</p>
                )}
              </div>
              <Button
                onClick={handleSave}
                disabled={isSaving || !hasChanges}
                variant={hasChanges ? "default" : "secondary"}
              >
                {isSaving ? (
                  <>
                    <Loader2 className="mr-2 size-4 animate-spin" />
                    Saving…
                  </>
                ) : (
                  <>
                    <Save className="mr-2 size-4" />
                    Save Email
                  </>
                )}
              </Button>
            </div>
          )}

          <Separator />

          {/* Scheduling section */}
          <ScheduleSection email={email} onEmailUpdate={onEmailUpdate} />
        </div>
      )}

      {!email &&
        streamPhase === "idle" &&
        !isFallbackGenerating &&
        !streamPreview && (
          <p className="py-6 text-center text-sm text-muted-foreground">
            Generate an email to get started with your outreach.
          </p>
        )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Email Status Badge (color-coded for email lifecycle)
// ---------------------------------------------------------------------------

function EmailStatusBadge({ status }: { status: string }) {
  const config: Record<string, { variant: "default" | "secondary" | "destructive" | "outline" }> = {
    draft: { variant: "secondary" },
    scheduled: { variant: "outline" },
    sending: { variant: "outline" },
    sent: { variant: "default" },
    failed: { variant: "destructive" },
  };
  const c = config[status] ?? config.draft;
  return (
    <Badge variant={c.variant} className="text-xs">
      {status}
    </Badge>
  );
}

// ---------------------------------------------------------------------------
// Schedule Section
// ---------------------------------------------------------------------------

function ScheduleSection({
  email,
  onEmailUpdate,
}: {
  email: Email;
  onEmailUpdate: (email: Email) => void;
}) {
  const [scheduleDate, setScheduleDate] = useState("");
  const [isScheduling, setIsScheduling] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const [error, setError] = useState("");

  const canSchedule = email.status === "draft";
  const isScheduled = email.status === "scheduled";
  const isSentOrSending = email.status === "sent" || email.status === "sending";

  function getQuickTime(hoursFromNow: number): string {
    const d = new Date(Date.now() + hoursFromNow * 3600_000);
    return d.toISOString();
  }

  function getTomorrow9AM(): string {
    const d = new Date();
    d.setDate(d.getDate() + 1);
    d.setHours(9, 0, 0, 0);
    return d.toISOString();
  }

  async function handleSchedule(sendAt: string) {
    setIsScheduling(true);
    setError("");
    try {
      const res = isScheduled
        ? await api.rescheduleEmail(email.id, sendAt)
        : await api.scheduleEmail(email.id, sendAt);
      onEmailUpdate(res.email);
      setScheduleDate("");
    } catch (err) {
      if (err instanceof ApiClientError) setError(err.message);
      else setError("Failed to schedule email.");
    } finally {
      setIsScheduling(false);
    }
  }

  async function handleCancel() {
    setIsCancelling(true);
    setError("");
    try {
      const res = await api.cancelSchedule(email.id);
      onEmailUpdate(res.email);
    } catch (err) {
      if (err instanceof ApiClientError) setError(err.message);
      else setError("Failed to cancel schedule.");
    } finally {
      setIsCancelling(false);
    }
  }

  function handleCustomSchedule() {
    if (!scheduleDate) return;
    handleSchedule(new Date(scheduleDate).toISOString());
  }

  // Already sent — show sent time
  if (isSentOrSending) {
    return (
      <div className="rounded-lg bg-muted/50 p-3 text-sm">
        <div className="flex items-center gap-2 text-muted-foreground">
          <Clock className="size-4" />
          {email.status === "sent" && email.sent_at
            ? `Sent on ${formatDateTime(email.sent_at)}`
            : "Sending…"}
        </div>
      </div>
    );
  }

  // Failed — show retry hint
  if (email.status === "failed") {
    return (
      <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
        Email failed to send (attempted {email.retry_count} times). Regenerate
        and try scheduling again.
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {error && (
        <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* Currently scheduled — show info + actions */}
      {isScheduled && email.scheduled_at && (
        <div className="flex flex-col gap-3 rounded-lg bg-muted/50 p-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-2 text-sm">
            <Calendar className="size-4 text-muted-foreground" />
            <span>
              Scheduled for{" "}
              <span className="font-medium">
                {formatDateTime(email.scheduled_at)}
              </span>
            </span>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleCancel}
              disabled={isCancelling}
            >
              {isCancelling ? (
                <Loader2 className="mr-1.5 size-3 animate-spin" />
              ) : (
                <X className="mr-1.5 size-3" />
              )}
              Cancel
            </Button>
          </div>
        </div>
      )}

      {/* Schedule controls — shown for draft and scheduled (reschedule) */}
      {(canSchedule || isScheduled) && (
        <>
          <Label className="text-xs">
            {isScheduled ? "Reschedule" : "Schedule Send"}
          </Label>

          {/* Quick-schedule buttons */}
          <div className="flex flex-wrap gap-2">
            {[
              { label: "In 1 hour", hours: 1 },
              { label: "In 3 hours", hours: 3 },
              { label: "Tomorrow 9 AM", hours: 0 },
            ].map((opt) => (
              <Button
                key={opt.label}
                variant="outline"
                size="sm"
                disabled={isScheduling}
                onClick={() =>
                  handleSchedule(
                    opt.hours > 0 ? getQuickTime(opt.hours) : getTomorrow9AM()
                  )
                }
              >
                <Clock className="mr-1.5 size-3" />
                {opt.label}
              </Button>
            ))}
          </div>

          {/* Custom date/time picker */}
          <div className="flex flex-col gap-2 sm:flex-row">
            <Input
              type="datetime-local"
              value={scheduleDate}
              onChange={(e) => setScheduleDate(e.target.value)}
              min={new Date().toISOString().slice(0, 16)}
              className="flex-1"
            />
            <Button
              onClick={handleCustomSchedule}
              disabled={isScheduling || !scheduleDate}
              className="shrink-0"
            >
              {isScheduling ? (
                <Loader2 className="mr-2 size-4 animate-spin" />
              ) : (
                <Calendar className="mr-2 size-4" />
              )}
              {isScheduled ? "Reschedule" : "Schedule"}
            </Button>
          </div>
        </>
      )}
    </div>
  );
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  });
}
