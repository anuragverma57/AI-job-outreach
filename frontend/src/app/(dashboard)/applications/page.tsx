"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { Briefcase, Plus } from "lucide-react";
import { buttonVariants } from "@/components/ui/button";
import { StatusBadge } from "@/components/applications/status-badge";
import { api } from "@/lib/api";
import type { Application } from "@/types/application";

export default function ApplicationsPage() {
  const [applications, setApplications] = useState<Application[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchApplications = useCallback(async () => {
    try {
      const res = await api.listApplications();
      setApplications(res.applications);
    } catch {
      setError("Failed to load applications.");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchApplications();
  }, [fetchApplications]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            Applications
          </h1>
          <p className="text-sm text-muted-foreground">
            Manage your job applications and outreach emails.
          </p>
        </div>
        <div className="shrink-0">
          <Link
            href="/applications/new"
            className={buttonVariants({ size: "lg" })}
          >
            <Plus className="mr-2 size-4" />
            New Application
          </Link>
        </div>
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
      ) : applications.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="grid gap-3">
          {applications.map((app) => (
            <ApplicationCard key={app.id} application={app} />
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
        <Briefcase className="size-6 text-muted-foreground" />
      </div>
      <h3 className="mt-4 text-sm font-medium">No applications yet</h3>
      <p className="mt-1 text-sm text-muted-foreground">
        Create your first application to start generating outreach emails.
      </p>
      <Link
        href="/applications/new"
        className={buttonVariants({ variant: "outline", size: "sm", className: "mt-4" })}
      >
        <Plus className="mr-2 size-4" />
        Create your first application
      </Link>
    </div>
  );
}

function ApplicationCard({ application }: { application: Application }) {
  return (
    <Link
      href={`/applications/${application.id}`}
      className="flex items-center justify-between gap-4 rounded-xl border bg-card px-4 py-4 transition-colors hover:bg-muted/30 sm:px-5"
    >
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <p className="truncate text-sm font-medium">
            {application.company_name}
          </p>
          <StatusBadge status={application.status} />
        </div>
        <p className="mt-0.5 truncate text-sm text-muted-foreground">
          {application.role}
        </p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {new Date(application.created_at).toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
            year: "numeric",
          })}
        </p>
      </div>
    </Link>
  );
}
