import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  ApiError,
} from "@/types/user";
import type { ResumeListResponse, ResumeUploadResponse } from "@/types/resume";
import type {
  ApplicationResponse,
  ApplicationListResponse,
  CreateApplicationRequest,
  ApplicationStatus,
  UpdateApplicationRequest,
} from "@/types/application";
import type {
  EmailResponse,
  EmailTone,
  GeneratedEmailDraft,
  UpdateEmailRequest,
  ScheduledEmailListResponse,
} from "@/types/email";
import type { AnalyticsSummaryResponse } from "@/types/analytics";
import type { SmartApplyResponse } from "@/types/smart-apply";

export const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const response = await fetch(url, {
      ...options,
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    if (!response.ok) {
      const errorBody: ApiError = await response.json().catch(() => ({
        error: "Something went wrong",
      }));
      throw new ApiClientError(errorBody.error, response.status);
    }

    return response.json();
  }

  private async upload<T>(endpoint: string, formData: FormData): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const response = await fetch(url, {
      method: "POST",
      credentials: "include",
      body: formData,
    });

    if (!response.ok) {
      const errorBody: ApiError = await response.json().catch(() => ({
        error: "Something went wrong",
      }));
      throw new ApiClientError(errorBody.error, response.status);
    }

    return response.json();
  }

  async login(data: LoginRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    return this.request<AuthResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async refresh(): Promise<AuthResponse> {
    return this.request<AuthResponse>("/api/auth/refresh", {
      method: "POST",
    });
  }

  async logout(): Promise<void> {
    await this.request("/api/auth/logout", { method: "POST" });
  }

  async getMe(): Promise<AuthResponse> {
    return this.request<AuthResponse>("/api/me");
  }

  // --- Resume endpoints ---

  async uploadResume(file: File): Promise<ResumeUploadResponse> {
    const formData = new FormData();
    formData.append("file", file);
    return this.upload<ResumeUploadResponse>("/api/resumes/", formData);
  }

  async listResumes(): Promise<ResumeListResponse> {
    return this.request<ResumeListResponse>("/api/resumes/");
  }

  async deleteResume(id: string): Promise<void> {
    await this.request(`/api/resumes/${id}`, { method: "DELETE" });
  }

  // --- Application endpoints ---

  async createApplication(
    data: CreateApplicationRequest
  ): Promise<ApplicationResponse> {
    return this.request<ApplicationResponse>("/api/applications", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async listApplications(): Promise<ApplicationListResponse> {
    return this.request<ApplicationListResponse>("/api/applications");
  }

  async getApplication(id: string): Promise<ApplicationResponse> {
    return this.request<ApplicationResponse>(`/api/applications/${id}`);
  }

  async smartApply(rawText: string): Promise<SmartApplyResponse> {
    return this.request<SmartApplyResponse>("/api/applications/smart-apply", {
      method: "POST",
      body: JSON.stringify({ raw_text: rawText }),
    });
  }

  async updateApplication(
    id: string,
    data: UpdateApplicationRequest
  ): Promise<ApplicationResponse> {
    return this.request<ApplicationResponse>(`/api/applications/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async updateApplicationStatus(
    id: string,
    status: ApplicationStatus
  ): Promise<ApplicationResponse> {
    return this.request<ApplicationResponse>(`/api/applications/${id}/status`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    });
  }

  async deleteApplication(id: string): Promise<void> {
    await this.request(`/api/applications/${id}`, { method: "DELETE" });
  }

  // --- Analytics ---

  async getAnalyticsSummary(): Promise<AnalyticsSummaryResponse> {
    return this.request<AnalyticsSummaryResponse>("/api/analytics/summary");
  }

  // --- Email endpoints ---

  async generateEmail(
    applicationId: string,
    tone: EmailTone,
    draft?: GeneratedEmailDraft
  ): Promise<EmailResponse> {
    const body: { tone: EmailTone; draft?: GeneratedEmailDraft } = { tone };
    if (draft) {
      body.draft = draft;
    }
    return this.request<EmailResponse>(
      `/api/applications/${applicationId}/generate-email`,
      { method: "POST", body: JSON.stringify(body) }
    );
  }

  async getEmail(applicationId: string): Promise<EmailResponse> {
    return this.request<EmailResponse>(
      `/api/applications/${applicationId}/email`
    );
  }

  async regenerateEmail(
    applicationId: string,
    tone: EmailTone,
    draft?: GeneratedEmailDraft
  ): Promise<EmailResponse> {
    const body: { tone: EmailTone; draft?: GeneratedEmailDraft } = { tone };
    if (draft) {
      body.draft = draft;
    }
    return this.request<EmailResponse>(
      `/api/applications/${applicationId}/regenerate-email`,
      { method: "POST", body: JSON.stringify(body) }
    );
  }

  async updateEmail(
    emailId: string,
    data: UpdateEmailRequest
  ): Promise<EmailResponse> {
    return this.request<EmailResponse>(`/api/emails/${emailId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  // --- Schedule endpoints ---

  async scheduleEmail(
    emailId: string,
    sendAt: string
  ): Promise<EmailResponse> {
    return this.request<EmailResponse>(`/api/emails/${emailId}/schedule`, {
      method: "POST",
      body: JSON.stringify({ send_at: sendAt }),
    });
  }

  async rescheduleEmail(
    emailId: string,
    sendAt: string
  ): Promise<EmailResponse> {
    return this.request<EmailResponse>(`/api/emails/${emailId}/schedule`, {
      method: "PUT",
      body: JSON.stringify({ send_at: sendAt }),
    });
  }

  async cancelSchedule(emailId: string): Promise<EmailResponse> {
    return this.request<EmailResponse>(`/api/emails/${emailId}/schedule`, {
      method: "DELETE",
    });
  }

  async listScheduledEmails(): Promise<ScheduledEmailListResponse> {
    return this.request<ScheduledEmailListResponse>(
      "/api/emails?status=scheduled"
    );
  }
}

export class ApiClientError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiClientError";
    this.status = status;
  }
}

export const api = new ApiClient(API_BASE_URL);
