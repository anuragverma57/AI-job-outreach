# Development Phases

Phased roadmap for the AI Job Outreach platform. Each phase should leave the product **demoable**. Prefer extending existing code over rewrites.

**Current phase:** **Phase 5 — Application tracking & dashboard**  
**Quick reality check:** See [CURRENT-STATE.md](CURRENT-STATE.md) for a short implementation snapshot.

---

## Status legend

| Label | Meaning |
|--------|--------|
| **Done** | Implemented in repo and wired end-to-end for the main path |
| **Partial** | Works but missing some roadmap items or uses a simpler design |
| **Not started** | No meaningful implementation yet |

---

## Phase 1: Foundation & infrastructure

**Goal:** Repo layout, local dev story, Postgres + Redis, health checks, migrations baseline.

| Item | Status | Notes |
|------|--------|--------|
| Monorepo layout (`api-gateway`, `ai-service`, `frontend`, docs) | **Done** | Go module, Python `requirements.txt`, Next.js |
| PostgreSQL + Redis (Docker) | **Done** | `docker-compose.yml`: Postgres host **5433**, Redis **6379** |
| Dockerfiles + Compose for API, AI, worker, frontend | **Not started** | Apps run on host via `Makefile` / manual commands |
| DB migrations (users, tokens, …) | **Done** | `api-gateway/migrations/` |
| API Gateway `/health` | **Done** | Fiber + DB check |
| AI service `/health` | **Done** | FastAPI |
| Worker process | **Done** | `api-gateway/cmd/worker` |
| Frontend shell / landing | **Done** | Next.js App Router |
| `.env.example`, `Makefile` (`up`, migrate, `run-api`, `run-ai`, `run-worker`) | **Done** | `make dev` starts infra + migrate + AI + API |

**Phase 1 summary:** **Partial** — production-style “all services in Compose” is optional later; local dev path is solid.

---

## Phase 2: Authentication & resume upload

**Goal:** Register, login, JWT access + refresh cookies, PDF resume upload, parsed text stored.

| Item | Status | Notes |
|------|--------|--------|
| Users + refresh tokens in DB | **Done** | Migrations `000001`, `000002` |
| `POST /api/auth/register`, `login`, `refresh`, `logout` | **Done** | See `internal/handler/auth.go` |
| JWT middleware, `GET /api/me` | **Done** | |
| Resumes table + upload / list / delete | **Done** | `POST/GET/DELETE /api/resumes` |
| AI `POST /ai/parse-resume` + gateway calls on upload | **Done** | OpenAI-compatible LLM via `LLM_BASE_URL` |
| Frontend auth + resumes pages | **Done** | Login, register, resumes |

**Phase 2 summary:** **Done**

---

## Phase 3: Job applications & AI email generation

**Goal:** Application CRUD, generate/regenerate email from resume + JD, edit draft in UI.

| Item | Status | Notes |
|------|--------|--------|
| Applications + emails tables | **Done** | Migrations `000004`, `000005` |
| Application CRUD API | **Done** | `POST/GET/PUT/DELETE /api/applications` |
| AI `POST /ai/generate-email` | **Done** | Tones: formal / friendly / concise |
| Generate / regenerate / get / update email | **Done** | Routes under `/api/applications/:id/...` and `PUT /api/emails/:id` |
| Frontend: new application, list, detail, email editor | **Done** | Scheduling UI also lives on detail flow |

**Phase 3 summary:** **Done** (editing **application fields** from the UI is optional; backend `PUT` exists if you add a thin client call later).

---

## Phase 4: Email scheduling & background worker

**Goal:** Schedule sends, worker delivers via SMTP, retries, status updates.

| Item | Status | Notes |
|------|--------|--------|
| Schedule / cancel / reschedule | **Done** | `POST/DELETE/PUT /api/emails/:id/schedule` |
| List by status (e.g. scheduled) | **Done** | `GET /api/emails?status=scheduled` |
| Redis queue + gateway enqueue | **Done** | `internal/queue` |
| Worker: claim due jobs, SMTP send | **Done** | `internal/sender/smtp.go` |
| Retries with backoff | **Done** | Max 3 retries in `cmd/worker/main.go` |
| On success: email **sent**, application → **applied** | **Done** | |
| Separate dead-letter Redis queue | **Partial** | Failures → `failed` in DB; no separate DLQ key |
| Frontend: scheduled list (`/emails`) | **Done** | |

**Phase 4 summary:** **Done** (DLQ can stay as DB status unless you outgrow it).

---

## Phase 5: Application tracking & dashboard (current)

**Goal:** User can move applications through the pipeline and see real stats — not placeholder zeros.

| Item | Status | Notes |
|------|--------|--------|
| `status` column on `applications` | **Done** | Default `draft`; worker sets `applied` after send |
| User-facing **status update** API (validated transitions) | **Not started** | `UpdateApplicationRequest` has no `status`; no `PATCH .../status` |
| Status **history** table / audit | **Not started** | |
| Analytics API (`/api/analytics/...` or minimal `/api/stats`) | **Not started** | |
| Dashboard: real numbers from API | **Not started** | `/dashboard` still hard-coded `0` |
| `/analytics` route | **Not started** | Sidebar links to it → 404 today |
| List filters (status, search) | **Not started** | Nice for MVP+ |

**Phase 5 summary:** **In progress** — schema and badges exist; **product value** (track replies/interviews + dashboard) is still to build.

---

## Phase 6: Polish, testing & hardening

**Goal:** Tests, consistent errors, rate limits, OpenAPI, production Docker, UX polish.

| Item | Status | Notes |
|------|--------|--------|
| Automated tests (Go / Python / frontend) | **Not started** | |
| Rate limiting, structured logging | **Not started** | Basic Fiber logger exists |
| OpenAPI / Swagger | **Not started** | |
| Multi-stage Dockerfiles, full Compose stack | **Not started** | |
| Toasts, empty/error states pass | **Partial** | Some pages handle errors; not uniform |

**Phase 6 summary:** **Not started**

---

## Phase summary table

| Phase | Focus | Status |
|-------|--------|--------|
| 1 | Foundation & infrastructure | **Partial** (dev-ready; full containerization later) |
| 2 | Auth & resumes | **Done** |
| 3 | Applications & AI emails | **Done** |
| 4 | Scheduling & worker | **Done** |
| 5 | Tracking & dashboard | **In progress** ← **you are here** |
| 6 | Polish & hardening | **Not started** |

---

## MVP scope (recommended)

**Must ship for a credible MVP**

1. **Status updates** — API + UI so users can record replied / interview / offer / rejected (and keep `applied` from send).
2. **One analytics read model** — e.g. `GET /api/analytics/summary` (counts by status + emails sent) without overbuilding timeline charts on day one.
3. **Dashboard + fix `/analytics`** — either implement the page or point the nav to `/dashboard` until Analytics exists.

**Defer (post-MVP)**

- Full status history timeline UI
- Advanced charts, search, and rate limiting (bring in during Phase 6 as needed)
- Running all services in Docker if local `Makefile` workflow is enough for you

---

## Original time estimates (rough)

| Phase | Est. duration |
|-------|----------------|
| 1 | 2–3 days |
| 2 | 3–4 days |
| 3 | 4–5 days |
| 4 | 3–4 days |
| 5 | 3–4 days |
| 6 | 3–5 days |

Estimates assumed greenfield; much of 1–4 is already done — focus remaining time on **Phase 5** then **Phase 6** slices you actually need.
