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
  scheduled_at: string | null;
  sent_at: string | null;
  retry_count: number;
  created_at: string;
  updated_at: string;
}

export interface EmailResponse {
  email: Email;
}

export interface GenerateEmailRequest {
  tone: EmailTone;
}

export interface UpdateEmailRequest {
  subject: string;
  body: string;
}
