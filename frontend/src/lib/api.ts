import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  ApiError,
} from "@/types/user";
import type { ResumeListResponse, ResumeUploadResponse } from "@/types/resume";

const API_BASE_URL =
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
