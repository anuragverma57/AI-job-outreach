export type ApplicationStatus =
  | "draft"
  | "applied"
  | "replied"
  | "interview"
  | "offer"
  | "rejected"
  | "ghosted";

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
