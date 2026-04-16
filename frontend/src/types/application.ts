export type ApplicationStatus =
  | "draft"
  | "applied"
  | "replied"
  | "interview"
  | "offer"
  | "rejected"
  | "ghosted";

/** Canonical pipeline LOV — string values must match api-gateway validation (see docs/IMPLEMENTATION-GUIDE.md Appendix A). */
export const APPLICATION_PIPELINE_STATUS_OPTIONS: {
  value: ApplicationStatus;
  label: string;
}[] = [
  { value: "draft", label: "Draft" },
  { value: "applied", label: "Applied" },
  { value: "replied", label: "Replied" },
  { value: "interview", label: "Interview" },
  { value: "offer", label: "Offer" },
  { value: "rejected", label: "Rejected" },
  { value: "ghosted", label: "Ghosted" },
];

export interface Application {
  id: string;
  user_id: string;
  resume_id: string | null;
  company_name: string;
  role: string;
  recruiter_email: string;
  job_description: string;
  job_link: string;
  status: ApplicationStatus;
  created_at: string;
  updated_at: string;
}

export interface CreateApplicationRequest {
  company_name: string;
  role: string;
  recruiter_email: string;
  job_description: string;
  job_link: string;
  resume_id?: string;
}

export interface ApplicationResponse {
  application: Application;
}

export interface ApplicationListResponse {
  applications: Application[];
}
