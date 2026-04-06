"use client";

import { useMemo } from "react";
import { AnalyticsSummaryCards } from "@/components/analytics/analytics-summary-cards";
import { useAnalyticsSummary } from "@/hooks/use-analytics-summary";
import { APPLICATION_PIPELINE_STATUS_OPTIONS } from "@/types/application";
import type { AnalyticsSummary } from "@/types/analytics";

function formatPercent(rate: number | undefined): string {
  if (rate == null || !Number.isFinite(rate)) return "—";
  return `${Math.round(rate * 100)}%`;
}

function maxStatusCount(summary: AnalyticsSummary | null): number {
  if (!summary?.applications_by_status) return 1;
  const values = APPLICATION_PIPELINE_STATUS_OPTIONS.map(
    (o) => summary.applications_by_status[o.value] ?? 0
  );
  return Math.max(1, ...values);
}

export default function AnalyticsPage() {
  const { summary, isLoading, error } = useAnalyticsSummary();

  const statusMax = useMemo(() => maxStatusCount(summary), [summary]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Analytics</h1>
        <p className="text-sm text-muted-foreground">
          Summary of your applications and outreach activity.
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
        <>
          <AnalyticsSummaryCards summary={summary} />

          <div className="grid gap-6 lg:grid-cols-2">
            <div className="space-y-4 rounded-xl border bg-card p-5 text-card-foreground">
              <h2 className="text-sm font-semibold">Applications by status</h2>
              <div className="space-y-3">
                {APPLICATION_PIPELINE_STATUS_OPTIONS.map((opt) => {
                  const count =
                    summary?.applications_by_status?.[opt.value] ?? 0;
                  const widthPct = Math.min(100, (count / statusMax) * 100);
                  return (
                    <div key={opt.value} className="space-y-1">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">{opt.label}</span>
                        <span className="tabular-nums font-medium">{count}</span>
                      </div>
                      <div className="h-2 overflow-hidden rounded-full bg-muted">
                        <div
                          className="h-full rounded-full bg-primary/80 transition-[width]"
                          style={{ width: `${widthPct}%` }}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            <div className="space-y-4 rounded-xl border bg-card p-5 text-card-foreground">
              <h2 className="text-sm font-semibold">Emails</h2>
              <dl className="grid grid-cols-2 gap-3 text-sm">
                <div className="rounded-lg bg-muted/50 p-3">
                  <dt className="text-muted-foreground">Sent</dt>
                  <dd className="mt-1 text-lg font-semibold tabular-nums">
                    {summary?.emails?.sent ?? 0}
                  </dd>
                </div>
                <div className="rounded-lg bg-muted/50 p-3">
                  <dt className="text-muted-foreground">Scheduled</dt>
                  <dd className="mt-1 text-lg font-semibold tabular-nums">
                    {summary?.emails?.scheduled ?? 0}
                  </dd>
                </div>
                <div className="rounded-lg bg-muted/50 p-3">
                  <dt className="text-muted-foreground">Failed</dt>
                  <dd className="mt-1 text-lg font-semibold tabular-nums">
                    {summary?.emails?.failed ?? 0}
                  </dd>
                </div>
                <div className="rounded-lg bg-muted/50 p-3">
                  <dt className="text-muted-foreground">Draft</dt>
                  <dd className="mt-1 text-lg font-semibold tabular-nums">
                    {summary?.emails?.draft ?? 0}
                  </dd>
                </div>
              </dl>

              {summary?.rates &&
              (summary.rates.reply_rate != null ||
                summary.rates.interview_rate != null) ? (
                <div className="border-t pt-4">
                  <h3 className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    Rates
                  </h3>
                  <div className="flex flex-wrap gap-4 text-sm">
                    {summary.rates.reply_rate != null ? (
                      <div>
                        <span className="text-muted-foreground">Reply rate </span>
                        <span className="font-medium tabular-nums">
                          {formatPercent(summary.rates.reply_rate)}
                        </span>
                      </div>
                    ) : null}
                    {summary.rates.interview_rate != null ? (
                      <div>
                        <span className="text-muted-foreground">
                          Interview rate{" "}
                        </span>
                        <span className="font-medium tabular-nums">
                          {formatPercent(summary.rates.interview_rate)}
                        </span>
                      </div>
                    ) : null}
                  </div>
                </div>
              ) : null}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
