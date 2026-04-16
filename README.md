# AI Job Outreach Automation Platform

An intelligent job application automation system that helps candidates streamline job hunting — from resumes and job applications to AI-generated cold emails, scheduled sending, and (planned) pipeline analytics.

## What this project does

**Resume & applications** → **AI draft email** → **Edit** → **Schedule** → **Worker sends via SMTP** → *(planned)* **Track & dashboard**

## Key capabilities (today)

| Capability | Status |
|------------|--------|
| Auth (register, login, refresh, cookies) | Implemented |
| Resume PDF upload, parse text, list/delete | Implemented |
| Job applications CRUD | Implemented |
| AI email generate / regenerate / edit | Implemented (depends on external LLM latency & quality) |
| Schedule / cancel / reschedule; list scheduled | Implemented |
| Background worker + SMTP send | Implemented (`make run-worker`) |
| Real dashboard analytics | Not yet (placeholder UI) |
| `/analytics` page | Not yet (sidebar link may 404) |

## Architecture (actual layout)

| Piece | Stack | Role |
|-------|--------|------|
| **API Gateway** | Go (Fiber) | REST API, auth, Postgres, Redis enqueue |
| **Worker** | Go (same repo, `cmd/worker`) | Dequeue, SMTP send, status updates |
| **AI Service** | Python (FastAPI) | `/ai/parse-resume`, `/ai/generate-email` → OpenAI-compatible LLM |
| **Frontend** | Next.js (App Router) | Dashboard UI |
| **PostgreSQL** | Docker (`make up`) | Primary store |
| **Redis** | Docker (`make up`) | Scheduled email queue |

**Docker today:** only **Postgres + Redis**. The app processes run locally (or you can containerize them later).

## Documentation

| Document | Description |
|----------|-------------|
| [Current state](docs/CURRENT-STATE.md) | What is implemented vs planned |
| [Core idea](docs/CORE-IDEA.md) | Problem, goals, vision |
| [Architecture](docs/ARCHITECTURE.md) | Services, data flow, DB sketch, stack snapshot, repo map |
| [Phases](docs/PHASES.md) | Current roadmap with done/in-progress/pending |
| [Implementation guide](docs/IMPLEMENTATION-GUIDE.md) | Build patterns + implementation playbooks (status/analytics) |

## Run locally

### 0) Environment

```bash
cp .env.example .env
```

Edit `.env`: Postgres (note **5433** on host), Redis, `AI_SERVICE_URL`, `LLM_BASE_URL` / `LLM_MODEL`, SMTP for the worker.

### 1) Infrastructure

```bash
make up
```

- Postgres: `localhost:5433`
- Redis: `localhost:6379`

### 2) Migrations

```bash
make migrate-up
```

### 3) Services (separate terminals as needed)

| Service | Command | URL |
|---------|---------|-----|
| API Gateway | `make run-api` | http://localhost:8080 |
| AI Service | `make setup-ai` once, then `make run-ai` | http://localhost:8000 |
| Worker | `make run-worker` | — |
| Frontend | `cd frontend && npm install && npm run dev` | http://localhost:3000 |

### Convenience

`make dev` runs: Docker infra → migrations → AI service (background) → API gateway. Start the **frontend** and **worker** yourself when you need them.

### Quick health checks

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8000/health
```

## Project status

**Active development** — Phases 1–4 largely implemented in code; Phase 5 (tracking & analytics API/UI) and full containerization of all apps are still open.

## License

MIT
