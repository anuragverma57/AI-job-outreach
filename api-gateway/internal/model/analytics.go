package model

// AnalyticsSummary is the inner object for GET /api/analytics/summary (field "summary").
type AnalyticsSummary struct {
	TotalApplications    int                 `json:"total_applications"`
	ApplicationsByStatus map[string]int      `json:"applications_by_status"`
	Emails               AnalyticsEmailStats `json:"emails"`
	Rates                AnalyticsRates      `json:"rates"`
}

// AnalyticsEmailStats counts emails linked to the user's applications, by email.status.
type AnalyticsEmailStats struct {
	Sent      int `json:"sent"`
	Scheduled int `json:"scheduled"`
	Failed    int `json:"failed"`
	Draft     int `json:"draft"`
}

// AnalyticsRates are derived metrics; see service layer comments for exact formulas.
type AnalyticsRates struct {
	ReplyRate     float64 `json:"reply_rate"`
	InterviewRate float64 `json:"interview_rate"`
}
