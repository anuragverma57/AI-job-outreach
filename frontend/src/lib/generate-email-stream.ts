import type { EmailTone, GeneratedEmailDraft } from "@/types/email";

export interface GenerateEmailStreamPayload {
  resume_text: string;
  job_description: string;
  company_name: string;
  role: string;
  recruiter_email: string;
  job_link: string;
  tone: EmailTone;
}

type StreamEvent =
  | { type: "delta"; text: string }
  | { type: "done"; result: GeneratedEmailDraft }
  | { type: "error"; detail: string };

function parseDataLine(line: string): StreamEvent | null {
  const prefix = "data: ";
  if (!line.startsWith(prefix)) {
    return null;
  }
  const jsonStr = line.slice(prefix.length).trim();
  if (!jsonStr) {
    return null;
  }
  try {
    const obj = JSON.parse(jsonStr) as Record<string, unknown>;
    const t = obj.type;
    if (t === "delta" && typeof obj.text === "string") {
      return { type: "delta", text: obj.text };
    }
    if (t === "done" && obj.result && typeof obj.result === "object") {
      const r = obj.result as Record<string, unknown>;
      return {
        type: "done",
        result: {
          subject: String(r.subject ?? ""),
          body: String(r.body ?? ""),
          match_score: Number(r.match_score ?? 0),
          key_points: Array.isArray(r.key_points)
            ? (r.key_points as unknown[]).map(String)
            : [],
          reasoning: String(r.reasoning ?? ""),
        },
      };
    }
    if (t === "error" && typeof obj.detail === "string") {
      return { type: "error", detail: obj.detail };
    }
  } catch {
    return null;
  }
  return null;
}

function processLine(
  line: string,
  onDelta?: (accumulated: string) => void,
  accumulatedRef?: { value: string }
): GeneratedEmailDraft | null {
  const ev = parseDataLine(line);
  if (!ev) {
    return null;
  }
  if (ev.type === "delta") {
    accumulatedRef!.value += ev.text;
    onDelta?.(accumulatedRef!.value);
    return null;
  }
  if (ev.type === "error") {
    throw new Error(ev.detail);
  }
  if (ev.type === "done") {
    if (!ev.result.subject?.trim() || !ev.result.body?.trim()) {
      throw new Error("Stream ended without a valid subject/body.");
    }
    return ev.result;
  }
  return null;
}

/**
 * POST /api/ai/generate-email/stream (gateway-proxied SSE). No artificial fetch timeout;
 * use AbortSignal to cancel. Parses `data: {"type":"delta"|"done"|"error",...}` lines.
 */
export async function streamGenerateEmail(
  baseUrl: string,
  payload: GenerateEmailStreamPayload,
  options: {
    onPhase?: (phase: "connecting" | "streaming") => void;
    onDelta?: (accumulated: string) => void;
    signal?: AbortSignal;
  } = {}
): Promise<GeneratedEmailDraft> {
  const { onPhase, onDelta, signal } = options;
  onPhase?.("connecting");

  const url = `${baseUrl.replace(/\/$/, "")}/api/ai/generate-email/stream`;
  const response = await fetch(url, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
    signal,
  });

  if (!response.ok) {
    const errBody = await response.json().catch(() => ({
      error: "Streaming request failed",
    }));
    const msg =
      typeof errBody === "object" && errBody && "error" in errBody
        ? String((errBody as { error: string }).error)
        : "Streaming request failed";
    throw new Error(msg);
  }

  const reader = response.body?.getReader();
  if (!reader) {
    throw new Error("No response body");
  }

  const decoder = new TextDecoder();
  let buffer = "";
  const accumulatedRef = { value: "" };
  onPhase?.("streaming");

  try {
    for (;;) {
      const { done, value } = await reader.read();
      if (done) {
        break;
      }
      buffer += decoder.decode(value, { stream: true });

      let newlineIdx: number;
      while ((newlineIdx = buffer.indexOf("\n")) >= 0) {
        const line = buffer.slice(0, newlineIdx).replace(/\r$/, "");
        buffer = buffer.slice(newlineIdx + 1);

        const result = processLine(line, onDelta, accumulatedRef);
        if (result) {
          return result;
        }
      }
    }

    const tail = buffer.trim();
    if (tail) {
      const result = processLine(tail, onDelta, accumulatedRef);
      if (result) {
        return result;
      }
    }
  } finally {
    reader.releaseLock();
  }

  throw new Error("Stream ended without a completion event.");
}
