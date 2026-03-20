import { z } from "zod";

export const createApplicationSchema = z.object({
  company_name: z.string().min(1, "Company name is required"),
  role: z.string().min(1, "Role is required"),
  recruiter_email: z.string().email("Please enter a valid email"),
  job_description: z.string().min(1, "Job description is required"),
  job_link: z.string().url("Please enter a valid URL").or(z.literal("")),
  resume_id: z.string().optional(),
});

export type CreateApplicationFormData = z.infer<typeof createApplicationSchema>;
