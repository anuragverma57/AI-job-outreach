import type { Application } from "@/types/application";
import type { Email } from "@/types/email";

export interface SmartApplyRequest {
  raw_text: string;
}

export interface SmartApplyMeta {
  extraction_confidence: string;
}

export interface SmartApplyResponse {
  application: Application;
  email: Email;
  meta: SmartApplyMeta;
}
