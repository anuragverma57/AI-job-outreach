# Implementation Guide

This document provides a detailed, step-by-step guide for building the AI Job Outreach platform. It covers how to implement each service, how the pieces connect, design decisions to make at each step, and patterns to follow.

---

## Prerequisites

Before starting, ensure you have installed:

- **Go** 1.22+ (`go version`)
- **Python** 3.11+ (`python3 --version`)
- **Node.js** 20+ (`node --version`)
- **Docker** and **Docker Compose** (`docker --version`, `docker compose version`)
- **Git** (`git --version`)
- A code editor (VS Code / Cursor recommended)
- An **LLM endpoint** that speaks **OpenAI Chat Completions** (URL + model name; API key optional), or a hosted key if you point `LLM_BASE_URL` at a cloud provider

---

## Step 1: Project Initialization

### 1.1 Create the monorepo

```bash
mkdir AI-job-outreach && cd AI-job-outreach
git init
```

Create the top-level directories:

```bash
mkdir -p api-gateway/cmd/server api-gateway/cmd/worker api-gateway/internal
mkdir -p ai-service/app
mkdir -p frontend
mkdir -p docs scripts
```

The **worker** is a second `main` package inside **`api-gateway/cmd/worker`** (same Go module as the gateway).

### 1.2 Initialize each service

**Go module (gateway + worker):**

```bash
cd api-gateway && go mod init github.com/<your-username>/ai-job-outreach/api-gateway && cd ..
```

**Python service:**

```bash
cd ai-service
python3 -m venv venv
pip install fastapi uvicorn pydantic pdfplumber python-dotenv httpx python-multipart
pip freeze > requirements.txt
cd ..
```

**Frontend:**

```bash
npx create-next-app@latest frontend --typescript --tailwind --eslint --app --src-dir
```

### 1.3 Create .env.example

Align with the root **`.env.example`** in the repo. Typical local dev:

```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
POSTGRES_USER=outreach
POSTGRES_PASSWORD=outreach_secret
POSTGRES_DB=outreach

REDIS_HOST=localhost
REDIS_PORT=6379

API_PORT=8080
JWT_SECRET=your-jwt-secret-change-in-production
CORS_ORIGINS=http://localhost:3000

AI_SERVICE_URL=http://localhost:8000
LLM_BASE_URL=http://your-llm-host:port/v1
LLM_MODEL=llama3.1
LLM_API_KEY=

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=your-email@gmail.com

NEXT_PUBLIC_API_URL=http://localhost:8080
```

### 1.4 Create .gitignore

```
# Environment
.env
*.env.local

# Go
api-gateway/tmp/
worker/tmp/

# Python
ai-service/venv/
ai-service/__pycache__/
*.pyc

# Node
frontend/node_modules/
frontend/.next/

# IDE
.vscode/
.idea/

# OS
.DS_Store

# Uploads
uploads/
```

---

## Step 2: Docker Setup

### 2.1 Docker Compose

**Current repository:** `docker-compose.yml` defines **postgres** and **redis** only. API gateway, worker, AI service, and frontend are started with **`Makefile`** on the host (`make run-api`, `make run-worker`, `make run-ai`, `npm run dev`).

**Future / production:** you can add Dockerfiles and extra Compose services so `docker compose up` runs everything; use the same env var names as `.env.example`.

Key decisions when you expand Compose:

- **Networking:** Same Docker network; DB/Redis hostnames become `postgres`, `redis` instead of `localhost`.
- **Volumes:** Named volumes for Postgres (and Redis if needed).
- **Health checks:** `depends_on` with `condition: service_healthy` where useful.

### 2.2 Dockerfiles

**Go services (api-gateway, worker):**

Use multi-stage builds:
1. Stage 1: Build the Go binary using `golang:1.22-alpine`
2. Stage 2: Copy binary into `alpine:latest` for a minimal runtime image

For development, use `air` for hot-reload by mounting source code and running `air` instead of the compiled binary.

**Python service (ai-service):**

1. Base: `python:3.11-slim`
2. Install dependencies from `requirements.txt`
3. Run with `uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload`

**Frontend:**

1. Base: `node:20-alpine`
2. Install dependencies, run `next dev` for development

### 2.3 Makefile

Create a `Makefile` at the project root with common commands:

```makefile
up:             docker compose up -d
down:           docker compose down
logs:           docker compose logs -f
restart:        docker compose restart
migrate-up:     run database migrations
migrate-down:   rollback last migration
seed:           load seed data
test-go:        run Go tests
test-python:    run Python tests
test-frontend:  run frontend tests
```

---

## Step 3: Database Design & Migrations

### 3.1 Migration strategy

Use `golang-migrate` for managing database schema. Migrations live in `api-gateway/migrations/`.

Each migration has an `up.sql` and `down.sql` file. Migrations are applied by the API Gateway on startup or via a CLI command.

### 3.2 Schema design principles

- Use **UUIDs** as primary keys (avoid exposing sequential IDs)
- Use **enums** (or check constraints) for status fields
- Add **indexes** on foreign keys and commonly queried columns
- Use **timestamps** (`created_at`, `updated_at`) on all tables
- Use **soft deletes** where appropriate (`deleted_at` nullable timestamp)

### 3.3 Migration order

1. `000001_create_users` — users table
2. `000002_create_refresh_tokens` — refresh token storage
3. `000003_create_resumes` — resumes table (FK → users)
4. `000004_create_applications` — applications table (FK → users, resumes)
5. `000005_create_emails` — emails table (FK → applications)

### 3.4 Key schema notes

**users table:**
- `password_hash` stored using bcrypt
- `email` must be unique

**resumes table:**
- `file_path` stores the path to the uploaded PDF
- `parsed_text` stores the extracted text (can be large, use TEXT type)
- One user can have multiple resumes

**applications table:**
- `status` enum: `draft`, `applied`, `replied`, `interview`, `offer`, `rejected`, `ghosted`
- `job_description` stored as TEXT
- `resume_id` is nullable (user might create application before selecting a resume)

**emails table:**
- `status` enum: `draft`, `scheduled`, `sending`, `sent`, `failed`
- `scheduled_at` is nullable (null = not scheduled)
- `sent_at` is nullable (null = not yet sent)
- `retry_count` defaults to 0, max 3

---

## Step 4: API Gateway Implementation

### 4.1 Architecture pattern

Follow the **Handler → Service → Repository** pattern:

```
HTTP Request
     │
     ▼
  Handler      ← Parses request, validates input, returns HTTP response
     │
     ▼
  Service      ← Business logic, orchestration, calls external services
     │
     ▼
  Repository   ← Database queries only, no business logic
```

Each layer communicates through **interfaces**, making unit testing straightforward.

### 4.2 Router setup

Use `Fiber` to define routes:

```
/api
├── /auth
│   ├── POST /register
│   ├── POST /login
│   └── POST /refresh
├── /resumes          (auth required)
│   ├── GET /
│   ├── POST /
│   └── DELETE /:id
├── /applications     (auth required)
│   ├── GET /
│   ├── POST /
│   ├── GET /:id
│   ├── PUT /:id
│   ├── DELETE /:id
│   ├── POST /:id/generate-email
│   ├── POST /:id/regenerate-email
│   └── PATCH /:id/status
├── /emails           (auth required)
│   ├── GET /:id
│   ├── PUT /:id
│   ├── POST /:id/schedule
│   ├── PUT /:id/schedule
│   └── DELETE /:id/schedule
└── /analytics        (auth required)
    ├── GET /summary
    ├── GET /timeline
    ├── GET /rates
    └── GET /recent
```

### 4.3 Authentication flow

1. **Registration:** Hash password with bcrypt, store user, return JWT
2. **Login:** Verify password against hash, issue JWT + refresh token
3. **Middleware:** Extract JWT from `Authorization: Bearer <token>` header, validate, inject user ID into request context
4. **Refresh:** Accept refresh token, validate, issue new JWT

JWT payload should include: `user_id`, `email`, `exp` (expiry), `iat` (issued at).

### 4.4 AI Service client

Create an HTTP client in `internal/client/ai_client.go` that:
- Calls the AI Service's endpoints
- Handles timeouts (LLM calls can take 10–30 seconds)
- Parses responses into Go structs
- Returns proper errors on failure

Set a generous timeout (30–60 seconds) for AI generation calls.

### 4.5 Redis queue producer

In `internal/queue/redis.go`, implement:

- `EnqueueEmail(emailID string, sendAt time.Time)` — adds to a sorted set with score = send_at unix timestamp
- `DequeueEmail(emailID string)` — removes from the sorted set (for cancellation)
- `RescheduleEmail(emailID string, newSendAt time.Time)` — update the score

Use a Redis sorted set (`ZADD`) where:
- **Key:** `email:scheduled`
- **Score:** Unix timestamp of when to send
- **Member:** Email ID (UUID)

---

## Step 5: AI Service Implementation

### 5.1 Endpoint design

```
POST /ai/parse-resume
  Input: PDF file upload
  Output: { parsed_text: string }

POST /ai/generate-email
  Input: {
    resume_text: string,
    job_description: string,
    company_name: string,
    role: string,
    tone: "formal" | "friendly" | "concise",
    candidate_name: string
  }
  Output: {
    subject: string,
    body: string,
    match_score: float (0-1),
    key_points: string[],
    reasoning: string
  }

GET /health
  Output: { status: "ok" }
```

### 5.2 Prompt engineering

The quality of generated emails depends heavily on the prompts. Key principles:

- **System prompt** defines the AI's role: "You are an expert career coach who writes compelling cold outreach emails..."
- **Provide full context** in the user prompt: resume text, JD, company, role
- **Specify output format** explicitly: ask for JSON with exact fields
- **Include constraints:** word count limits, no generic phrases, must reference specific experience
- **Tone parameter** adjusts the system prompt variation

Store prompts in `app/prompts/` as Python string templates. This makes them easy to iterate on without changing business logic.

### 5.3 Resume parsing

Use `pdfplumber` to extract text from PDF files:
1. Receive the PDF as a file upload
2. Extract all text content
3. Clean up whitespace, headers/footers
4. Return as a single string

The parsed text is stored in PostgreSQL by the API Gateway. Future AI calls use the stored text, not the PDF.

### 5.4 Match analysis

Before generating the email, the service should:
1. Extract key requirements from the JD (skills, experience, qualifications)
2. Identify matching elements in the resume
3. Score the match (0-1 scale)
4. Use this analysis to inform the email generation

This can be a single LLM call with a structured prompt, or two calls (analyze then generate) for better quality.

---

## Step 6: Worker Service Implementation

### 6.1 Architecture

The worker runs as a long-lived process with two main loops:

```
main()
  ├── Start scheduler goroutine (polls for due emails)
  ├── Start consumer goroutine (processes email send jobs)
  └── Block on shutdown signal (SIGINT, SIGTERM)
```

### 6.2 Scheduler

The scheduler runs on a ticker (every 30 seconds):

```
Every 30 seconds:
  1. ZRANGEBYSCORE email:scheduled 0 <current_unix_timestamp>
     → Get all emails where scheduled time <= now
  2. For each email:
     a. ZREM email:scheduled <email_id>  (remove from sorted set)
     b. LPUSH email:send <email_id>      (push to send queue)
```

This two-step process (sorted set → list) separates "when to send" from "what to process."

### 6.3 Consumer

The consumer uses `BRPOP` on the `email:send` list queue:

```
Loop:
  1. BRPOP email:send 5  (block for 5 seconds)
  2. If job received:
     a. Fetch email details from PostgreSQL
     b. Send email via SMTP/API
     c. On success: update status to "sent", update application to "applied"
     d. On failure: increment retry_count
        - If retry_count < 3: re-enqueue with backoff delay
        - If retry_count >= 3: update status to "failed"
```

### 6.4 Worker pool

Use a bounded worker pool (e.g., 5 concurrent workers) to process jobs:

```go
pool := make(chan struct{}, 5) // max 5 concurrent sends
for job := range jobs {
    pool <- struct{}{} // acquire slot
    go func(j Job) {
        defer func() { <-pool }() // release slot
        processJob(j)
    }(job)
}
```

### 6.5 Graceful shutdown

On receiving SIGINT/SIGTERM:
1. Stop accepting new jobs
2. Wait for in-progress jobs to complete (with a timeout)
3. Close database and Redis connections
4. Exit cleanly

---

## Step 7: Frontend Implementation

### 7.1 Page structure

```
/                    → Dashboard (redirects to /login if not authenticated)
/login               → Login form
/register            → Registration form
/applications        → Application list with filters
/applications/new    → New application form
/applications/:id    → Application detail + email preview/edit + scheduling
/resumes             → Resume management (upload, list, delete)
/analytics           → Analytics dashboard with charts
```

### 7.2 State management

- Use **React Query** (TanStack Query) for server state (API data fetching, caching, invalidation)
- Use **React Context** for auth state (JWT token, user info)
- No need for Redux or Zustand — React Query handles most state needs

### 7.3 API client

Create a centralized API client in `lib/api.ts`:
- Base URL from environment variable
- Automatic JWT token injection in headers
- Response interceptor for 401 → redirect to login
- Type-safe request/response with TypeScript interfaces

### 7.4 Component design

Follow the pattern: **Page → Layout → Feature Components → UI Primitives**

- Pages are thin — they compose feature components
- Feature components (e.g., `ApplicationForm`, `EmailEditor`) contain the logic
- UI primitives (shadcn/ui) handle visual presentation
- Custom hooks abstract data fetching from components

### 7.5 Key UI flows

**New Application Flow:**
1. User fills in company name, role, recruiter email, JD, job link
2. User selects which resume to use
3. Submit → API creates application → AI generates email
4. Show loading state during AI generation (can take 10-20 seconds)
5. Display generated email with edit option
6. User reviews, optionally edits, then schedules or sends

**Dashboard:**
- Top row: stat cards (total, sent, replied, interviews, response rate)
- Middle: status breakdown chart + timeline chart
- Bottom: recent activity feed

---

## Step 8: Testing Strategy

### 8.1 Go tests

- **Unit tests** for services (mock repositories and external clients)
- **Integration tests** for repositories (use a test database or testcontainers)
- **Handler tests** using `httptest` to test HTTP endpoints
- Run with: `go test ./...`

### 8.2 Python tests

- **Unit tests** for resume parser, matcher, email generator
- **Mock LLM responses** to avoid API costs during testing
- Run with: `pytest`

### 8.3 Frontend tests

- **Component tests** with React Testing Library
- **Integration tests** for key flows (create application → generate email → schedule)
- Run with: `npm test`

---

## Step 9: Development Workflow

### Daily workflow

```bash
# Start all services
make up

# Watch logs
make logs

# API Gateway runs at http://localhost:8080
# AI Service runs at http://localhost:8000
# Frontend runs at http://localhost:3000

# Run migrations
make migrate-up

# Run tests
make test-go
make test-python

# Stop everything
make down
```

### Adding a new feature

1. Design the API endpoint (request/response schema)
2. Write the database migration (if needed)
3. Implement Repository → Service → Handler (bottom-up)
4. Write tests for the service layer
5. Implement the frontend component
6. Test the full flow end-to-end

### API-first development

Build and test the backend API first using Postman or `curl`. Once the API is solid, build the frontend against it. This prevents rework from API design changes.

---

## Common Pitfalls to Avoid

1. **Don't call the LLM from the frontend directly** — always go through the API Gateway, which controls access and can cache responses
2. **Don't store secrets in code** — use environment variables for everything
3. **Don't skip input validation** — validate on both frontend and backend
4. **Don't make the AI Service stateful** — it should never access the database directly
5. **Don't use polling for real-time updates** — start with manual refresh, add WebSockets later if needed
6. **Don't over-engineer early** — get the happy path working first, then add error handling and edge cases
7. **Don't forget database indexes** — add indexes on foreign keys and frequently queried columns from the start

---

## Appendix A: Pipeline status (LOV) playbook

Use this as the default pattern for user-managed opportunity tracking.

### Canonical statuses

- `draft`
- `applied`
- `replied`
- `interview`
- `offer`
- `rejected`
- `ghosted`

### API contract

- `PATCH /api/applications/:id/status`
- Request: `{ "status": "<one-of-lov>" }`
- Response: `{ "application": { ...updated row... } }`
- Validation: reject unknown statuses with `400`
- Auth/ownership: same ownership checks as application CRUD

### Responsibility split

- **Backend**
  - Validate status against canonical LOV
  - Enforce user ownership
  - Persist update and return updated application
  - Keep worker compatibility (worker may still set `applied`)
- **Frontend**
  - Add status update control in application detail
  - Keep `ApplicationStatus` union aligned with backend LOV
  - Show loading/error state and refresh badge values
- **Integration**
  - Confirm one shared JSON contract and status keys

---

## Appendix B: Analytics summary playbook

Use this for the first analytics iteration. Keep it DB-driven and simple.

### API contract

- `GET /api/analytics/summary` (authenticated)
- Recommended response shape:
  - `total_applications`
  - `applications_by_status` (always include all status keys, zero-filled)
  - `emails` (`sent`, `scheduled`, `failed`, optional `draft`)
  - optional `rates` (`reply_rate`, `interview_rate`)

### Minimum query model

- Applications by user:
  - `SELECT status, COUNT(*) FROM applications WHERE user_id = $1 GROUP BY status`
- Emails by user (joined through applications):
  - `SELECT e.status, COUNT(*) FROM emails e INNER JOIN applications a ON a.id = e.application_id WHERE a.user_id = $1 GROUP BY e.status`

### Responsibility split

- **Backend**
  - Build `summary` endpoint with user-scoped aggregation
  - Zero-fill missing status keys before returning JSON
  - Keep error format consistent with current handlers
- **Frontend**
  - Add `/analytics` page (to avoid dead sidebar route)
  - Replace dashboard placeholder numbers with live summary data
  - Add loading/error/empty states
- **Integration**
  - Reuse one endpoint for both dashboard and analytics page

### Scope guardrails

- In scope: summary cards + status breakdown
- Out of scope (next step): timeline charts, inbox/Gmail sync, new analytics tables
