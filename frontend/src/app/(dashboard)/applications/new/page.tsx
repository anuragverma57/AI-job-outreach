"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { api, ApiClientError } from "@/lib/api";
import {
  createApplicationSchema,
  type CreateApplicationFormData,
} from "@/lib/validations/application";
import type { Resume } from "@/types/resume";

export default function NewApplicationPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState("");
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [resumesLoading, setResumesLoading] = useState(true);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<CreateApplicationFormData>({
    resolver: zodResolver(createApplicationSchema),
    defaultValues: { job_link: "" },
  });

  const fetchResumes = useCallback(async () => {
    try {
      const res = await api.listResumes();
      setResumes(res.resumes);
    } catch {
      /* non-blocking */
    } finally {
      setResumesLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchResumes();
  }, [fetchResumes]);

  async function onSubmit(data: CreateApplicationFormData) {
    setServerError("");

    try {
      const res = await api.createApplication({
        company_name: data.company_name,
        role: data.role,
        recruiter_email: data.recruiter_email,
        job_description: data.job_description,
        job_link: data.job_link || "",
        resume_id: data.resume_id || undefined,
      });
      router.push(`/applications/${res.application.id}`);
    } catch (error) {
      if (error instanceof ApiClientError) {
        setServerError(error.message);
      } else {
        setServerError("Something went wrong. Please try again.");
      }
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div className="flex items-center gap-3">
        <Link
          href="/applications"
          className={buttonVariants({ variant: "ghost", size: "icon" })}
        >
          <ArrowLeft className="size-4" />
        </Link>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">
            New Application
          </h1>
          <p className="text-sm text-muted-foreground">
            Enter the job details to create an application.
          </p>
        </div>
      </div>

      <form
        onSubmit={handleSubmit(onSubmit)}
        className="space-y-5 rounded-xl border bg-card p-5 sm:p-6"
      >
        {serverError && (
          <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {serverError}
          </div>
        )}

        <div className="grid gap-5 sm:grid-cols-2">
          <div className="grid gap-1.5">
            <Label htmlFor="company_name">Company name</Label>
            <Input
              id="company_name"
              placeholder="Google"
              aria-invalid={!!errors.company_name}
              {...register("company_name")}
            />
            {errors.company_name && (
              <p className="text-xs text-destructive">
                {errors.company_name.message}
              </p>
            )}
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="role">Role / Job title</Label>
            <Input
              id="role"
              placeholder="Senior Frontend Engineer"
              aria-invalid={!!errors.role}
              {...register("role")}
            />
            {errors.role && (
              <p className="text-xs text-destructive">{errors.role.message}</p>
            )}
          </div>
        </div>

        <div className="grid gap-5 sm:grid-cols-2">
          <div className="grid gap-1.5">
            <Label htmlFor="recruiter_email">Recruiter email</Label>
            <Input
              id="recruiter_email"
              type="email"
              placeholder="recruiter@company.com"
              aria-invalid={!!errors.recruiter_email}
              {...register("recruiter_email")}
            />
            {errors.recruiter_email && (
              <p className="text-xs text-destructive">
                {errors.recruiter_email.message}
              </p>
            )}
          </div>

          <div className="grid gap-1.5">
            <Label htmlFor="job_link">Job posting link</Label>
            <Input
              id="job_link"
              type="url"
              placeholder="https://..."
              aria-invalid={!!errors.job_link}
              {...register("job_link")}
            />
            {errors.job_link && (
              <p className="text-xs text-destructive">
                {errors.job_link.message}
              </p>
            )}
          </div>
        </div>

        <div className="grid gap-1.5">
          <Label htmlFor="resume_id">Resume</Label>
          <select
            id="resume_id"
            className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-50 dark:bg-input/30"
            disabled={resumesLoading}
            {...register("resume_id")}
          >
            <option value="">
              {resumesLoading ? "Loading resumes…" : "Select a resume (optional)"}
            </option>
            {resumes.map((r) => (
              <option key={r.id} value={r.id}>
                {r.file_name}
              </option>
            ))}
          </select>
          {resumes.length === 0 && !resumesLoading && (
            <p className="text-xs text-muted-foreground">
              No resumes uploaded yet.{" "}
              <Link href="/resumes" className="underline underline-offset-2">
                Upload one
              </Link>
            </p>
          )}
        </div>

        <div className="grid gap-1.5">
          <Label htmlFor="job_description">Job description</Label>
          <Textarea
            id="job_description"
            placeholder="Paste the full job description here…"
            className="min-h-40"
            aria-invalid={!!errors.job_description}
            {...register("job_description")}
          />
          {errors.job_description && (
            <p className="text-xs text-destructive">
              {errors.job_description.message}
            </p>
          )}
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <Link
            href="/applications"
            className={buttonVariants({ variant: "outline" })}
          >
            Cancel
          </Link>
          <Button type="submit" size="lg" disabled={isSubmitting}>
            {isSubmitting ? "Creating…" : "Create Application"}
          </Button>
        </div>
      </form>
    </div>
  );
}
