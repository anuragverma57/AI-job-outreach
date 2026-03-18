"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { FileText, Plus, Trash2, Upload } from "lucide-react";
import { Button } from "@/components/ui/button";
import { api, ApiClientError } from "@/lib/api";
import type { Resume } from "@/types/resume";

export default function ResumesPage() {
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchResumes = useCallback(async () => {
    try {
      const res = await api.listResumes();
      setResumes(res.resumes);
    } catch {
      setError("Failed to load resumes.");
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchResumes();
  }, [fetchResumes]);

  async function handleUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsUploading(true);
    setError("");

    try {
      const res = await api.uploadResume(file);
      setResumes((prev) => [res.resume, ...prev]);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError("Failed to upload resume.");
      }
    } finally {
      setIsUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  async function handleDelete(id: string) {
    try {
      await api.deleteResume(id);
      setResumes((prev) => prev.filter((r) => r.id !== id));
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError("Failed to delete resume.");
      }
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Resumes</h1>
          <p className="text-sm text-muted-foreground">
            Upload and manage your resumes for AI email generation.
          </p>
        </div>
        <div>
          <input
            ref={fileInputRef}
            type="file"
            accept=".pdf"
            className="hidden"
            onChange={handleUpload}
          />
          <Button
            size="lg"
            onClick={() => fileInputRef.current?.click()}
            disabled={isUploading}
          >
            {isUploading ? (
              <>
                <Upload className="mr-2 size-4 animate-spin" />
                Uploading…
              </>
            ) : (
              <>
                <Plus className="mr-2 size-4" />
                Upload Resume
              </>
            )}
          </Button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="flex items-center justify-center py-16">
          <div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
        </div>
      ) : resumes.length === 0 ? (
        <EmptyState onUpload={() => fileInputRef.current?.click()} />
      ) : (
        <div className="grid gap-3">
          {resumes.map((resume) => (
            <ResumeCard
              key={resume.id}
              resume={resume}
              onDelete={handleDelete}
            />
          ))}
        </div>
      )}
    </div>
  );
}

function EmptyState({ onUpload }: { onUpload: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center rounded-xl border border-dashed py-16 text-center">
      <div className="rounded-full bg-muted p-3">
        <FileText className="size-6 text-muted-foreground" />
      </div>
      <h3 className="mt-4 text-sm font-medium">No resumes yet</h3>
      <p className="mt-1 text-sm text-muted-foreground">
        Upload a PDF resume to get started with AI email generation.
      </p>
      <Button variant="outline" size="sm" className="mt-4" onClick={onUpload}>
        <Plus className="mr-2 size-4" />
        Upload your first resume
      </Button>
    </div>
  );
}

function ResumeCard({
  resume,
  onDelete,
}: {
  resume: Resume;
  onDelete: (id: string) => void;
}) {
  const [isDeleting, setIsDeleting] = useState(false);

  async function handleDelete() {
    setIsDeleting(true);
    await onDelete(resume.id);
    setIsDeleting(false);
  }

  return (
    <div className="flex items-center justify-between rounded-xl border bg-card px-5 py-4 transition-colors hover:bg-muted/30">
      <div className="flex items-center gap-4">
        <div className="rounded-lg bg-muted p-2.5">
          <FileText className="size-5 text-muted-foreground" />
        </div>
        <div>
          <p className="text-sm font-medium">{resume.file_name}</p>
          <p className="text-xs text-muted-foreground">
            Uploaded {new Date(resume.created_at).toLocaleDateString("en-US", {
              month: "short",
              day: "numeric",
              year: "numeric",
            })}
          </p>
        </div>
      </div>
      <Button
        variant="ghost"
        size="icon"
        onClick={handleDelete}
        disabled={isDeleting}
        className="text-muted-foreground hover:text-destructive"
      >
        <Trash2 className="size-4" />
      </Button>
    </div>
  );
}
