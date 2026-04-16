import type { ApplicationStatus } from "./application";

/** Email aggregate counts from GET /api/analytics/summary (see docs/IMPLEMENTATION-GUIDE.md Appendix B). */
export interface AnalyticsEmailsSummary {
  sent: number;
  scheduled: number;
  failed: number;
  draft: number;
}

/**
 * Optional rates from the API; meaning is defined server-side (denominator in backend comments).
 */
export interface AnalyticsRatesSummary {
  reply_rate?: number;
  interview_rate?: number;
}

/**
 * Full analytics payload — `applications_by_status` should include every LOV key (zeros where empty).
 */
export interface AnalyticsSummary {
  total_applications: number;
  applications_by_status: Record<ApplicationStatus, number>;
  emails: AnalyticsEmailsSummary;
  rates?: AnalyticsRatesSummary;
}

export interface AnalyticsSummaryResponse {
  summary: AnalyticsSummary;
}
