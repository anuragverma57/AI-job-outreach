# Current implementation state

This file is a **snapshot** of what the repository implements today. For the original roadmap, see [PHASES.md](PHASES.md).

## Shipped in code

| Area | Notes |
|------|--------|
| **Infra (Docker Compose)** | `postgres` (host port **5433** → container 5432) and `redis` only. API, AI, worker, and frontend run **on the host** via `Makefile` / manual commands. |
| **API Gateway (Go / Fiber)** | `GET /health`; auth (`/api/auth/*`, `/api/me`); resumes CRUD upload/list/delete; applications CRUD; email generate/regenerate/get/update; email schedule/cancel/reschedule/list by status. Uses **JWT access** + **HttpOnly refresh cookies** (see handlers). |
| **Redis queue** | `internal/queue` — used when scheduling sends (gateway enqueues; worker claims). |
| **Worker** | `api-gateway/cmd/worker` — same Go module as gateway; polls Redis, sends via **SMTP** (`internal/sender`), updates DB. Run: `make run-worker`. |
| **AI service (FastAPI)** | `GET /health`; `POST /ai/parse-resume`; `POST /ai/generate-email`. Calls an **OpenAI-compatible** LLM at `LLM_BASE_URL` + `/chat/completions` using **httpx** (see `ai-service/app/config.py`). |
| **Frontend (Next.js)** | Auth pages; dashboard shell (placeholder stats); resumes; applications list/new/detail (email UI + scheduling); **Email outreach** list (`/emails`). Sidebar **Analytics** → route not implemented yet. |

## Not implemented yet (planned)

- **PATCH** application status API + history + dashboard analytics (Phase 5).
- **Full Docker** for api-gateway, ai-service, worker, frontend (optional future).
- **Production** hardening, tests, OpenAPI docs (Phase 6).

## Environment highlights

See root `.env.example`: `DATABASE_URL` / Postgres vars, `REDIS_URL` or host/port, `AI_SERVICE_URL`, `LLM_BASE_URL` / `LLM_MODEL`, SMTP for worker.

**Important:** `LLM_BASE_URL` must point at the **LLM HTTP API** (e.g. `http://host:port/v1`), not at the FastAPI `ai-service` URL.
