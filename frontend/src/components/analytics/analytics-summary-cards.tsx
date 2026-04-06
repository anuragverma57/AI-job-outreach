import type { AnalyticsSummary } from "@/types/analytics";

function formatCount(n: number): string {
  return String(Number.isFinite(n) ? n : 0);
}

/**
 * Top-line metrics aligned with the dashboard: totals, outreach sent, and pipeline counts.
 */
export function AnalyticsSummaryCards({
  summary,
}: {
  summary: AnalyticsSummary | null;
}) {
  const total = summary?.total_applications ?? 0;
  const sent = summary?.emails?.sent ?? 0;
  const replies = summary?.applications_by_status?.replied ?? 0;
  const interviews = summary?.applications_by_status?.interview ?? 0;

  const stats = [
    { label: "Total Applications", value: formatCount(total) },
    { label: "Emails Sent", value: formatCount(sent) },
    { label: "Replies", value: formatCount(replies) },
    { label: "Interviews", value: formatCount(interviews) },
  ];

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {stats.map((stat) => (
        <div
          key={stat.label}
          className="rounded-xl border bg-card p-5 text-card-foreground"
        >
          <p className="text-sm text-muted-foreground">{stat.label}</p>
          <p className="mt-1 text-2xl font-semibold tabular-nums">{stat.value}</p>
        </div>
      ))}
    </div>
  );
}
