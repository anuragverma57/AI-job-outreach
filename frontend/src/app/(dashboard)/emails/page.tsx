"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import {
  Calendar,
  Clock,
  Loader2,
  Mail,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { api, ApiClientError } from "@/lib/api";
import type { ScheduledEmail } from "@/types/email";

export default function ScheduledEmailsPage() {
  const [emails, setEmails] = useState<ScheduledEmail[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchEmails = useCallback(async () => {
    try {
      const res = await api.listScheduledEmails();
      setEmails(res.emails);
    } catch {
      setError("Failed to load scheduled emails.");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchEmails();
  }, [fetchEmails]);

  async function handleCancel(emailId: string) {
    try {
      await api.cancelSchedule(emailId);
      setEmails((prev) => prev.filter((e) => e.id !== emailId));
    } catch (err) {
      if (err instanceof ApiClientError) setError(err.message);
      else setError("Failed to cancel schedule.");
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">
          Email Outreach
        </h1>
        <p className="text-sm text-muted-foreground">
          Scheduled emails waiting to be sent.
        </p>
      </div>

      {error && (
        <div className="rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
        </div>
      ) : emails.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="grid gap-3">
          {emails.map((email) => (
            <ScheduledEmailCard
              key={email.id}
              email={email}
              onCancel={handleCancel}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border border-dashed py-16 text-center">
      <div className="rounded-full bg-muted p-3">
        <Mail className="size-6 text-muted-foreground" />
      </div>
      <h3 className="mt-4 text-sm font-medium">No scheduled emails</h3>
      <p className="mt-1 max-w-xs text-sm text-muted-foreground">
        When you schedule an email from an application, it will appear here.
      </p>
    </div>
  );
}

function ScheduledEmailCard({
  email,
  onCancel,
}: {
  email: ScheduledEmail;
  onCancel: (id: string) => void;
}) {
  const [isCancelling, setIsCancelling] = useState(false);

  async function handleCancel() {
    setIsCancelling(true);
    await onCancel(email.id);
    setIsCancelling(false);
  }

  return (
    <div className="flex flex-col gap-3 rounded-xl border bg-card p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5">
      <div className="min-w-0 space-y-1">
        {email.application && (
          <Link
            href={`/applications/${email.application.id}`}
            className="text-sm font-medium hover:underline"
          >
            {email.application.company_name} — {email.application.role}
          </Link>
        )}
        <p className="truncate text-sm text-muted-foreground">
          {email.subject}
        </p>
        {email.scheduled_at && (
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <Calendar className="size-3" />
            {formatDateTime(email.scheduled_at)}
          </div>
        )}
      </div>

      <div className="flex shrink-0 items-center gap-2">
        <Badge variant="outline" className="text-xs">
          <Clock className="mr-1 size-3" />
          Scheduled
        </Badge>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleCancel}
          disabled={isCancelling}
          className="text-muted-foreground hover:text-destructive"
        >
          {isCancelling ? (
            <Loader2 className="size-4 animate-spin" />
          ) : (
            <X className="size-4" />
          )}
        </Button>
      </div>
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
