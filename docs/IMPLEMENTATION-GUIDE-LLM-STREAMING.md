# LLM Streaming — Implementation Guide

This document defines how to add **OpenAI-compatible SSE streaming** for remote LLM calls (e.g. Ollama on another machine), how it **relates to reliability**, and **who does what** on backend vs frontend.

**Reference server (example):** `http://192.168.29.231:8000` — `POST /v1/chat/completions` with `"stream": true` returns **SSE** when configured correctly. Verify with **curl** before writing integration code (see §2).

---

## 1. Goals and misconceptions

### What streaming does

- Sends the model output **incrementally** (token chunks) so the UI can show **live text** and a **progress feel**.
- Keeps a **long-lived HTTP response** open; users are less likely to assume the app is frozen.

### What streaming does **not** do

- It does **not** prevent timeouts, network errors, or server crashes. Those still require **timeouts**, **retries**, and **clear error handling**.
- **Slow remote hardware** still needs **generous read timeouts** (or stream-based reads with a **max duration**).

### What actually keeps the app “running”

| Layer | Action |
|--------|--------|
| **Client (httpx / fetch)** | Large or streaming read timeout; optional max wall-clock for the whole stream |
| **Retries** | Transient failures only (connection reset, 502/503/504, timeouts) with backoff + cap |
| **Streaming** | Better **UX** and perceived responsiveness |
| **Fallback** | Non-stream retry once if stream fails; user-visible message on failure |

---

## 2. Verify streaming before coding (mandatory)

Use the **same base URL** as `LLM_BASE_URL` in `.env` (e.g. `http://192.168.29.231:8000/v1` — no trailing slash issues).

**OpenAI-compatible streaming (recommended for this app):**

```bash
curl -sS -N --no-buffer \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3",
    "messages": [{"role":"user","content":"Count 1 to 3 in one word per line."}],
    "stream": true
  }' \
  "http://192.168.29.231:8000/v1/chat/completions"
```

**Expected:** Lines starting with `data: `, chunks with `choices[0].delta.content`, and `data: [DONE]` at the end.

**Non-streaming baseline (unchanged behavior):**

```bash
curl -sS \
  -H "Content-Type: application/json" \
  -d '{"model":"llama3","messages":[{"role":"user","content":"Hi"}],"stream":false}' \
  "http://192.168.29.231:8000/v1/chat/completions"
```

**Note:** If your server also exposes `/chat?stream=true` (NDJSON), that is a **different protocol**. For **one** implementation path, standardize on **`/v1/chat/completions` + `stream: true`** in the AI service unless product requires both.

---

## 3. JSON-structured outputs (email generation, Smart Apply)

Today the model is prompted to return **JSON** (subject, body, match_score, …). With streaming:

- **Tokens arrive as deltas**; the full string is only valid **as JSON after the stream ends** (or after you buffer the assistant message).
- **UX:** Show a **live preview** of raw text (or a “streaming…” placeholder) while chunks arrive; **parse JSON only after** the stream completes (or use a dedicated non-stream path for final parse if you split flows).

**Do not** call `JSON.parse` on partial chunks.

---

## 4. Architecture options

| Option | Flow | Pros | Cons |
|--------|------|------|------|
| **A** | Browser → **AI service** (SSE) directly | Fewer hops | CORS, auth, exposing internal URL |
| **B** | Browser → **Go gateway** → AI service (SSE) **proxy** | Single auth, AI internal | Gateway must support streaming proxy |
| **C** | Browser → **Next.js** route → AI service | OK for dev | Extra hop, still need streaming |

**Recommendation:** **Option B** for production alignment with existing gateway auth; **Option A** acceptable only for local dev with strict CORS and no secrets in browser.

This guide does not mandate Option B vs A — **pick one** and document in the PR.

---

## 5. Backend responsibilities (AI service + optional gateway)

### AI service (Python / FastAPI) — primary owner

| Task | Owner | Notes |
|------|--------|--------|
| Shared **SSE client** for `POST {LLM_BASE_URL}/chat/completions` with `"stream": true` | Backend | Use `httpx` `stream=True`; parse `data:` lines; accumulate `delta.content` |
| **Buffer** full assistant text; then **existing** `parse_email_json` / extraction logic | Backend | Same as today, after stream completes |
| New routes **optional** pattern: `POST /ai/generate-email/stream` returning **SSE** or **NDJSON** to the client | Backend | Or extend existing endpoint with `Accept: text/event-stream` / query `?stream=1` |
| **Timeouts** | Backend | Long `timeout` on httpx for read; or separate connect vs read; cap total stream duration in code |
| **Env** | Backend | `LLM_STREAM=true` to toggle stream vs non-stream for rollout |
| **Retries** | Backend | Optional: retry non-stream once on transient failure; streaming retry is trickier — document behavior |

### API gateway (Go) — only if browser talks to gateway

| Task | Owner | Notes |
|------|--------|--------|
| Proxy `POST` to AI service and **forward** `text/event-stream` body to client | Backend | Fiber `Stream` or raw body copy; preserve headers |
| **Auth** same as other `/api` routes | Backend | |
| **Timeout** on upstream | Backend | Must allow long streams (same order of magnitude as AI service) |

### Out of scope for this feature (unless explicitly added)

- Worker / SMTP (unchanged)
- Changing **Redis** or **Postgres** schema for streaming

---

## 6. Frontend responsibilities

| Task | Owner | Notes |
|------|--------|--------|
| Call streaming endpoint with **`fetch`** + `ReadableStream` **or** `EventSource` if SSE URL is GET (prefer POST + fetch for chat) | Frontend | |
| **Decode** SSE lines (`data: ...`), parse JSON chunks, accumulate `content` deltas | Frontend | Match OpenAI chunk shape |
| **UI states:** idle → loading → streaming (live text) → success (parsed fields) → error | Frontend | |
| **Do not** abort early with short timeouts; use configurable **long** timeout or stream-only | Frontend | |
| **Fallback:** button “Retry without streaming” or automatic fallback calling existing non-stream API | Frontend | Optional but valuable |
| **Accessibility:** announce loading/streaming for screen readers | Frontend | Nice-to-have |

---

## 7. Testing checklist

- [ ] curl streaming (§2) succeeds from the machine where **ai-service** runs.
- [ ] Non-stream path still works end-to-end (regression).
- [ ] Stream completes and JSON parse succeeds for email generation.
- [ ] Slow network: UI does not show “failed” at 30s if server is still streaming.
- [ ] Error path: user sees message + retry.

---

## 8. Assignment summary

| Role | Delivers |
|------|----------|
| **Backend (AI)** | SSE stream reader from LLM; optional streaming HTTP endpoints; env flag; timeouts; buffer-then-parse JSON |
| **Backend (Gateway)** | Optional SSE proxy + auth + long timeouts |
| **Frontend** | Stream consumer, progressive UI, final parse + existing forms |
| **Integration** | One chosen path (direct to AI vs via gateway); shared contract for chunk format |

---

## 9. Copy-paste prompts for agents

**Reference file:** `docs/IMPLEMENTATION-GUIDE-LLM-STREAMING.md`

### Prompt — backend agent

```
Implement LLM streaming per docs/IMPLEMENTATION-GUIDE-LLM-STREAMING.md.

Scope:
1. Confirm LLM streaming uses POST {LLM_BASE_URL}/chat/completions with "stream": true (OpenAI SSE). Add a small httpx-based stream reader that accumulates assistant text from deltas, then reuses existing parsers (e.g. parse_email_json) after the stream ends.
2. Add env flag (e.g. LLM_STREAM) to switch stream vs non-stream for safe rollout.
3. Set sensible httpx timeouts (long read timeout; optional max wall-clock for stream).
4. Expose streaming to clients via a new FastAPI route (SSE or NDJSON) OR extend existing generate-email flow with a streaming variant—document the chosen path.
5. If the product requires the browser to talk only to the Go gateway, add a streaming proxy route in the gateway with auth; otherwise document that the frontend calls the AI service directly in dev only.

Do not claim streaming fixes reliability by itself; implement timeouts/retries as described in the doc. Test with curl from §2 before integration tests.
```

### Prompt — frontend agent

```
Implement streaming UX per docs/IMPLEMENTATION-GUIDE-LLM-STREAMING.md.

Scope:
1. Consume the backend streaming endpoint (fetch + ReadableStream or agreed client) and parse SSE chunks to build live assistant text during email generation (and Smart Apply if applicable).
2. Show clear states: idle → connecting → streaming (live preview) → success (parsed subject/body from final JSON) → error with retry.
3. Avoid short fetch timeouts; align with long-running remote LLM.
4. Optional: fallback button to call existing non-streaming generate API.
5. Match existing UI components and layout; do not add heavy charting libs.

Coordinate with backend on exact URL, auth (cookies vs public AI URL), and response format (SSE vs NDJSON).
```

### Prompt — full-stack (single agent)

```
Implement end-to-end LLM streaming for email generation (minimum) per docs/IMPLEMENTATION-GUIDE-LLM-STREAMING.md: verify curl §2, implement AI service stream + buffer + parse, then frontend streaming UI with states and optional non-stream fallback. Document gateway proxy only if the app must not call the AI service URL from the browser.
```

---

## 10. Implemented contract (this repo)

**Chosen path:** **Option B** — browsers call the **Go API gateway** for streaming so JWT/cookie auth stays consistent; the gateway proxies to the AI service. Direct calls to the AI service are possible in local dev only (no gateway auth); ensure CORS if the browser talks to the AI host directly.

| Component | Behavior |
|-----------|----------|
| **Env (`ai-service`)** | `LLM_STREAM` — when `true`, internal `POST /ai/generate-email` and Smart Apply use upstream `stream: true` (SSE), buffer assistant text, then `parse_email_json`. On stream failure, **one non-stream retry** is attempted. `LLM_HTTP_*` and `LLM_STREAM_MAX_SECONDS` control httpx timeouts and optional max wall-clock for a stream. |
| **Upstream LLM** | Always `POST {LLM_BASE_URL}/chat/completions` with OpenAI-style SSE when streaming (`"stream": true`). |
| **AI service — client streaming** | `POST /ai/generate-email/stream` returns **`text/event-stream`**: each line is `data: {"type":"delta","text":"..."}`; final line is `data: {"type":"done","result":{...}}` (same fields as `GenerateEmailResponse`). Errors may yield `data: {"type":"error","detail":"..."}`. |
| **API gateway** | `POST /api/ai/generate-email/stream` (authenticated) proxies the request body to `AI_SERVICE_URL/ai/generate-email/stream` and forwards the SSE body (long-lived client timeout). |

**Regression:** `POST /ai/generate-email` without streaming flag still exists; when `LLM_STREAM=false`, behavior matches the previous non-stream JSON completion path.
