"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowLeft, Loader2, Sparkles } from "lucide-react";
import Link from "next/link";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { api, ApiClientError } from "@/lib/api";
import {
  createApplicationSchema,
  type CreateApplicationFormData,
} from "@/lib/validations/application";
import type { Resume } from "@/types/resume";
import type { SmartApplyResponse } from "@/types/smart-apply";

type NewApplicationMode = "manual" | "smart";

interface SmartReviewDraft {
  applicationId: string;
  emailId: string;
  initialResumeId: string;
  companyName: string;
  role: string;
  recruiterEmail: string;
  jobLink: string;
  jobDescription: string;
  resumeId: string;
  emailSubject: string;
  emailBody: string;
  extractionConfidence: string;
}

export default function NewApplicationPage() {
  const router = useRouter();
  const [mode, setMode] = useState<NewApplicationMode>("manual");
  const [serverError, setServerError] = useState("");
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [resumesLoading, setResumesLoading] = useState(true);
  const [rawText, setRawText] = useState("");
  const [smartError, setSmartError] = useState("");
  const [isSmartProcessing, setIsSmartProcessing] = useState(false);
  const [isSavingSmartDraft, setIsSavingSmartDraft] = useState(false);
  const [smartDraft, setSmartDraft] = useState<SmartReviewDraft | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<CreateApplicationFormData>({
    resolver: zodResolver(createApplicationSchema),
    defaultValues: { job_link: "" },
  });

  const fetchResumes = useCallback(async () => {
    try {
      const res = await api.listResumes();
      setResumes(res.resumes);
    } catch {
      /* non-blocking */
    } finally {
      setResumesLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchResumes();
  }, [fetchResumes]);

  async function onSubmit(data: CreateApplicationFormData) {
    setServerError("");

    try {
      const res = await api.createApplication({
        company_name: data.company_name,
        role: data.role,
        recruiter_email: data.recruiter_email,
        job_description: data.job_description,
        job_link: data.job_link || "",
        resume_id: data.resume_id || undefined,
      });
      router.push(`/applications/${res.application.id}`);
    } catch (error) {
      if (error instanceof ApiClientError) {
        setServerError(error.message);
      } else {
        setServerError("Something went wrong. Please try again.");
      }
    }
  }

  function hydrateSmartDraft(res: SmartApplyResponse): SmartReviewDraft {
    return {
      applicationId: res.application.id,
      emailId: res.email.id,
      initialResumeId: res.application.resume_id ?? "",
      companyName: res.application.company_name ?? "",
      role: res.application.role ?? "",
      recruiterEmail: res.application.recruiter_email ?? "",
      jobLink: res.application.job_link ?? "",
      jobDescription: res.application.job_description ?? "",
      resumeId: res.application.resume_id ?? "",
      emailSubject: res.email.subject ?? "",
      emailBody: res.email.body ?? "",
      extractionConfidence: res.meta?.extraction_confidence ?? "unknown",
    };
  }

  async function handleSmartApplyExtract() {
    setSmartError("");
    if (!rawText.trim()) {
      setSmartError("Paste the job content before extracting.");
      return;
    }
    if (!resumesLoading && resumes.length === 0) {
      setSmartError("Upload a resume first to use Smart Apply.");
      return;
    }
    setIsSmartProcessing(true);
    try {
      const res = await api.smartApply(rawText);
      setSmartDraft(hydrateSmartDraft(res));
    } catch (error) {
      if (error instanceof ApiClientError) {
        setSmartError(error.message);
      } else {
        setSmartError("Smart Apply failed. Please retry or switch to manual.");
      }
    } finally {
      setIsSmartProcessing(false);
    }
  }

  const smartMissingRequired =
    !smartDraft ||
    !smartDraft.companyName.trim() ||
    !smartDraft.role.trim() ||
    !smartDraft.jobDescription.trim() ||
    !smartDraft.emailSubject.trim() ||
    !smartDraft.emailBody.trim();
  const hasResumeSelectionChanged =
    !!smartDraft && smartDraft.resumeId !== smartDraft.initialResumeId;

  async function handleSmartDraftContinue() {
    if (!smartDraft) return;
    setSmartError("");
    if (smartMissingRequired) {
      setSmartError("Fill company, role, job description, email subject, and email body.");
      return;
    }
    setIsSavingSmartDraft(true);
    try {
      await api.updateApplication(smartDraft.applicationId, {
        company_name: smartDraft.companyName.trim(),
        role: smartDraft.role.trim(),
        recruiter_email: smartDraft.recruiterEmail.trim(),
        job_description: smartDraft.jobDescription.trim(),
        job_link: smartDraft.jobLink.trim(),
      });
      await api.updateEmail(smartDraft.emailId, {
        subject: smartDraft.emailSubject.trim(),
        body: smartDraft.emailBody.trim(),
      });
      router.push(`/applications/${smartDraft.applicationId}`);
    } catch (error) {
      if (error instanceof ApiClientError) {
        setSmartError(error.message);
      } else {
        setSmartError("Failed to save Smart Apply draft. Try again.");
      }
    } finally {
      setIsSavingSmartDraft(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center gap-3">
        <Link
          href="/applications"
          className={buttonVariants({ variant: "ghost", size: "icon" })}
        >
          <ArrowLeft className="size-4" />
        </Link>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            New Application
          </h1>
          <p className="text-sm text-muted-foreground">
            Choose manual entry or Smart Apply from pasted job content.
          </p>
        </div>
      </div>

      <div className="inline-flex rounded-lg border bg-muted/40 p-1">
        <button
          type="button"
          onClick={() => setMode("manual")}
          className={`rounded-md px-3 py-1.5 text-sm transition ${
            mode === "manual"
              ? "bg-background shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          Manual
        </button>
        <button
          type="button"
          onClick={() => setMode("smart")}
          className={`rounded-md px-3 py-1.5 text-sm transition ${
            mode === "smart"
              ? "bg-background shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          Smart Apply
        </button>
      </div>

      {mode === "manual" ? (
        <form
          onSubmit={handleSubmit(onSubmit)}
          className="space-y-5 rounded-xl border bg-card p-5 sm:p-6"
        >
          {serverError && (
            <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {serverError}
            </div>
          )}

          <div className="grid gap-5 sm:grid-cols-2">
            <div className="grid gap-1.5">
              <Label htmlFor="company_name">Company name</Label>
              <Input
                id="company_name"
                placeholder="Google"
                aria-invalid={!!errors.company_name}
                {...register("company_name")}
              />
              {errors.company_name && (
                <p className="text-xs text-destructive">
                  {errors.company_name.message}
                </p>
              )}
            </div>

            <div className="grid gap-1.5">
              <Label htmlFor="role">Role / Job title</Label>
              <Input
                id="role"
                placeholder="Senior Frontend Engineer"
                aria-invalid={!!errors.role}
                {...register("role")}
              />
              {errors.role && (
                <p className="text-xs text-destructive">{errors.role.message}</p>
              )}
            </div>
          </div>

          <div className="grid gap-5 sm:grid-cols-2">
            <div className="grid gap-1.5">
              <Label htmlFor="recruiter_email">Recruiter email</Label>
              <Input
                id="recruiter_email"
                type="email"
                placeholder="recruiter@company.com"
                aria-invalid={!!errors.recruiter_email}
                {...register("recruiter_email")}
              />
              {errors.recruiter_email && (
                <p className="text-xs text-destructive">
                  {errors.recruiter_email.message}
                </p>
              )}
            </div>

            <div className="grid gap-1.5">
              <Label htmlFor="job_link">Job posting link</Label>
              <Input
                id="job_link"
                type="url"
                placeholder="https://..."
                aria-invalid={!!errors.job_link}
                {...register("job_link")}
              />
              {errors.job_link && (
                <p className="text-xs text-destructive">
                  {errors.job_link.message}
                </p>
              )}
            </div>
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="resume_id">Resume</Label>
            <select
              id="resume_id"
              className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-50 dark:bg-input/30"
              disabled={resumesLoading}
              {...register("resume_id")}
            >
              <option value="">
                {resumesLoading
                  ? "Loading resumes…"
                  : "Select a resume (optional)"}
              </option>
              {resumes.map((r) => (
                <option key={r.id} value={r.id}>
                  {r.file_name}
                </option>
              ))}
            </select>
            {resumes.length === 0 && !resumesLoading && (
              <p className="text-xs text-muted-foreground">
                No resumes uploaded yet.{" "}
                <Link href="/resumes" className="underline underline-offset-2">
                  Upload one
                </Link>
              </p>
            )}
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="job_description">Job description</Label>
            <Textarea
              id="job_description"
              placeholder="Paste the full job description here…"
              className="min-h-40"
              aria-invalid={!!errors.job_description}
              {...register("job_description")}
            />
            {errors.job_description && (
              <p className="text-xs text-destructive">
                {errors.job_description.message}
              </p>
            )}
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <Link
              href="/applications"
              className={buttonVariants({ variant: "outline" })}
            >
              Cancel
            </Link>
            <Button type="submit" size="lg" disabled={isSubmitting}>
              {isSubmitting ? "Creating…" : "Create Application"}
            </Button>
          </div>
        </form>
      ) : (
        <div className="space-y-5 rounded-xl border bg-card p-5 sm:p-6">
          {smartError && (
            <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {smartError}
            </div>
          )}

          {!smartDraft ? (
            <>
              <div className="grid gap-1.5">
                <Label htmlFor="smart_raw_text">Paste job post / JD</Label>
                <Textarea
                  id="smart_raw_text"
                  value={rawText}
                  onChange={(e) => setRawText(e.target.value)}
                  className="min-h-48"
                  placeholder="Paste LinkedIn job post, hiring email, or job description text here..."
                  disabled={isSmartProcessing}
                />
                <p className="text-xs text-muted-foreground">
                  Smart Apply will extract fields, select a resume, and draft your email.
                </p>
              </div>

              {resumes.length === 0 && !resumesLoading && (
                <p className="text-xs text-muted-foreground">
                  Upload at least one resume to use Smart Apply.{" "}
                  <Link href="/resumes" className="underline underline-offset-2">
                    Go to resumes
                  </Link>
                </p>
              )}

              <div className="flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setMode("manual")}
                  className={buttonVariants({ variant: "outline" })}
                >
                  Switch to manual
                </button>
                <Button
                  type="button"
                  onClick={handleSmartApplyExtract}
                  disabled={isSmartProcessing || resumesLoading}
                >
                  {isSmartProcessing ? (
                    <>
                      <Loader2 className="mr-2 size-4 animate-spin" />
                      Processing…
                    </>
                  ) : (
                    <>
                      <Sparkles className="mr-2 size-4" />
                      Extract & Draft
                    </>
                  )}
                </Button>
              </div>
            </>
          ) : (
            <>
              <div className="rounded-lg bg-muted/40 p-3 text-xs text-muted-foreground">
                Extraction confidence:{" "}
                <span className="font-medium text-foreground">
                  {smartDraft.extractionConfidence}
                </span>
              </div>
              {(!smartDraft.recruiterEmail.trim() || !smartDraft.jobLink.trim()) && (
                <div className="rounded-lg bg-amber-500/10 px-3 py-2 text-xs text-amber-700 dark:text-amber-300">
                  Partial extraction: recruiter email and/or job link were not found. Fill them manually if available.
                </div>
              )}
              <div className="grid gap-5 sm:grid-cols-2">
                <div className="grid gap-1.5">
                  <Label htmlFor="smart_company">Company name</Label>
                  <Input
                    id="smart_company"
                    value={smartDraft.companyName}
                    onChange={(e) =>
                      setSmartDraft((prev) =>
                        prev ? { ...prev, companyName: e.target.value } : prev
                      )
                    }
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label htmlFor="smart_role">Role / Job title</Label>
                  <Input
                    id="smart_role"
                    value={smartDraft.role}
                    onChange={(e) =>
                      setSmartDraft((prev) =>
                        prev ? { ...prev, role: e.target.value } : prev
                      )
                    }
                  />
                </div>
              </div>

              <div className="grid gap-5 sm:grid-cols-2">
                <div className="grid gap-1.5">
                  <Label htmlFor="smart_email">Recruiter email</Label>
                  <Input
                    id="smart_email"
                    placeholder="recruiter@company.com"
                    value={smartDraft.recruiterEmail}
                    onChange={(e) =>
                      setSmartDraft((prev) =>
                        prev ? { ...prev, recruiterEmail: e.target.value } : prev
                      )
                    }
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label htmlFor="smart_link">Job posting link</Label>
                  <Input
                    id="smart_link"
                    placeholder="https://..."
                    value={smartDraft.jobLink}
                    onChange={(e) =>
                      setSmartDraft((prev) =>
                        prev ? { ...prev, jobLink: e.target.value } : prev
                      )
                    }
                  />
                </div>
              </div>

              <div className="grid gap-1.5">
                <Label htmlFor="smart_resume_id">Resume</Label>
                <select
                  id="smart_resume_id"
                  value={smartDraft.resumeId}
                  onChange={(e) =>
                    setSmartDraft((prev) =>
                      prev ? { ...prev, resumeId: e.target.value } : prev
                    )
                  }
                  className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
                >
                  <option value="">Select resume</option>
                  {resumes.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.file_name}
                    </option>
                  ))}
                </select>
                {hasResumeSelectionChanged ? (
                  <p className="text-xs text-muted-foreground">
                    Resume selection has changed in review. Current backend uses the originally selected resume for the generated draft email.
                  </p>
                ) : null}
              </div>

              <div className="grid gap-1.5">
                <Label htmlFor="smart_jd">Job description</Label>
                <Textarea
                  id="smart_jd"
                  className="min-h-36"
                  value={smartDraft.jobDescription}
                  onChange={(e) =>
                    setSmartDraft((prev) =>
                      prev ? { ...prev, jobDescription: e.target.value } : prev
                    )
                  }
                />
              </div>

              <div className="space-y-3 rounded-lg border p-4">
                <h2 className="text-sm font-semibold">Draft email</h2>
                <div className="grid gap-1.5">
                  <Label htmlFor="smart_subject">Subject</Label>
                  <Input
                    id="smart_subject"
                    value={smartDraft.emailSubject}
                    onChange={(e) =>
                      setSmartDraft((prev) =>
                        prev ? { ...prev, emailSubject: e.target.value } : prev
                      )
                    }
                  />
                </div>
                <div className="grid gap-1.5">
                  <Label htmlFor="smart_body">Body</Label>
                  <Textarea
                    id="smart_body"
                    className="min-h-44"
                    value={smartDraft.emailBody}
                    onChange={(e) =>
                      setSmartDraft((prev) =>
                        prev ? { ...prev, emailBody: e.target.value } : prev
                      )
                    }
                  />
                </div>
              </div>

              <div className="flex justify-end gap-3 pt-1">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setSmartDraft(null)}
                  disabled={isSavingSmartDraft}
                >
                  Start over
                </Button>
                <Button
                  type="button"
                  onClick={handleSmartDraftContinue}
                  disabled={isSavingSmartDraft || smartMissingRequired}
                >
                  {isSavingSmartDraft ? (
                    <>
                      <Loader2 className="mr-2 size-4 animate-spin" />
                      Saving…
                    </>
                  ) : (
                    "Continue to Schedule/Send"
                  )}
                </Button>
              </div>
            </>
          )}
        </div>
      )}
    </div>
  );
}
