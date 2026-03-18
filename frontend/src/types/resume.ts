export interface Resume {
  id: string;
  user_id: string;
  file_name: string;
  parsed_text?: string | null;
  created_at: string;
}

export interface ResumeListResponse {
  resumes: Resume[];
}

export interface ResumeUploadResponse {
  resume: Resume;
}
