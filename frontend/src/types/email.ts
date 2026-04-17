export type EmailStatus =
  | "draft"
  | "scheduled"
  | "sending"
  | "sent"
  | "failed";

export type EmailTone = "formal" | "friendly" | "concise";

export interface Email {
  id: string;
  application_id: string;
  subject: string;
  body: string;
  status: EmailStatus;
  match_score: number | null;
  key_points: string[] | null;
  reasoning?: string | null;
  scheduled_at: string | null;
  sent_at: string | null;
  retry_count: number;
  created_at: string;
  updated_at: string;
}

export interface EmailResponse {
  email: Email;
}

/** Payload for POST /api/applications/:id/generate-email — optional draft skips LLM (after streaming). */
export interface GeneratedEmailDraft {
  subject: string;
  body: string;
  match_score: number;
  key_points: string[];
  reasoning: string;
}

export interface GenerateEmailRequest {
  tone: EmailTone;
  draft?: GeneratedEmailDraft;
}

export interface UpdateEmailRequest {
  subject: string;
  body: string;
}

export interface ScheduleEmailRequest {
  send_at: string;
}

export interface ScheduledEmail extends Email {
  application?: {
    id: string;
    company_name: string;
    role: string;
    recruiter_email: string;
  };
}

export interface ScheduledEmailListResponse {
  emails: ScheduledEmail[];
}
