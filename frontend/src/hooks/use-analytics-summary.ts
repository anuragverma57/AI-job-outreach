"use client";

import { useCallback, useEffect, useState } from "react";
import { api, ApiClientError } from "@/lib/api";
import type { AnalyticsSummary } from "@/types/analytics";

export function useAnalyticsSummary() {
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");

  const refetch = useCallback(async () => {
    setIsLoading(true);
    setError("");
    try {
      const res = await api.getAnalyticsSummary();
      setSummary(res.summary);
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError("Failed to load analytics.");
      }
      setSummary(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void refetch();
  }, [refetch]);

  return { summary, isLoading, error, refetch };
}
