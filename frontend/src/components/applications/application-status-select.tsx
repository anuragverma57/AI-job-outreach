"use client";

import { useState } from "react";
import { Loader2 } from "lucide-react";
import { Label } from "@/components/ui/label";
import { api, ApiClientError } from "@/lib/api";
import {
  APPLICATION_PIPELINE_STATUS_OPTIONS,
  type Application,
  type ApplicationStatus,
} from "@/types/application";

const selectClassName =
  "h-8 min-w-[9.5rem] rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-input/30";

export function ApplicationStatusSelect({
  applicationId,
  status,
  onUpdated,
}: {
  applicationId: string;
  status: ApplicationStatus;
  onUpdated: (application: Application) => void;
}) {
  const [isUpdating, setIsUpdating] = useState(false);
  const [error, setError] = useState("");

  async function handleChange(next: ApplicationStatus) {
    if (next === status) return;
    setIsUpdating(true);
    setError("");
    try {
      const res = await api.updateApplicationStatus(applicationId, next);
      onUpdated(res.application);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError("Failed to update status.");
      }
    } finally {
      setIsUpdating(false);
    }
  }

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center gap-2">
        <Label htmlFor="application-pipeline-status" className="sr-only">
          Pipeline status
        </Label>
        <select
          id="application-pipeline-status"
          aria-busy={isUpdating}
          value={status}
          disabled={isUpdating}
          onChange={(e) =>
            handleChange(e.target.value as ApplicationStatus)
          }
          className={selectClassName}
        >
          {APPLICATION_PIPELINE_STATUS_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        {isUpdating ? (
          <Loader2
            className="size-4 shrink-0 animate-spin text-muted-foreground"
            aria-hidden
          />
        ) : null}
      </div>
      {error ? (
        <p className="text-xs text-destructive" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
