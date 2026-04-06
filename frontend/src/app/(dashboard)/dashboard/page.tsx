"use client";

import { useAuth } from "@/hooks/use-auth";
import { useAnalyticsSummary } from "@/hooks/use-analytics-summary";
import { AnalyticsSummaryCards } from "@/components/analytics/analytics-summary-cards";

export default function DashboardPage() {
  const { user } = useAuth();
  const { summary, isLoading, error } = useAnalyticsSummary();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">
          Welcome back, {user?.name?.split(" ")[0]}
        </h1>
        <p className="text-sm text-muted-foreground">
          Here&apos;s an overview of your job outreach pipeline.
        </p>
      </div>

      {error ? (
        <div className="rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      ) : null}

      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
        </div>
      ) : (
        <AnalyticsSummaryCards summary={summary} />
      )}

      <div className="rounded-xl border bg-card p-6 text-card-foreground">
        <p className="text-sm text-muted-foreground">
          Start by uploading your resume, then create your first application.
        </p>
      </div>
    </div>
  );
}
