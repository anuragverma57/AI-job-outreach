# Development Phases

The project is broken into 6 phases. Each phase produces a working, testable increment. No phase depends on future phases, so you can stop at any point and have a functional system.

## Where the code is today

See **[CURRENT-STATE.md](CURRENT-STATE.md)** for an accurate checklist. In short: **Phases 1–4 are largely implemented** (with Compose limited to Postgres + Redis and apps run via Makefile); **Phases 5–6 are still open**. The unchecked items below are kept as the original roadmap — use `CURRENT-STATE.md` when they conflict with reality.

---

## Phase 1: Foundation & Infrastructure

**Goal:** Set up the project skeleton, Docker environment, database, and verify all services can start and communicate.

**Duration estimate:** 2–3 days

### Tasks

- [ ] Initialize the monorepo structure (directories, go.mod, requirements.txt, package.json)
- [ ] Create `docker-compose.yml` with all 6 containers (api-gateway, ai-service, worker, frontend, postgres, redis)
- [ ] Create Dockerfiles for each custom service
- [ ] Set up PostgreSQL with initial migration (users table)
- [ ] Set up Redis and verify connectivity
- [ ] API Gateway: basic HTTP server with health check endpoint (`GET /health`)
- [ ] AI Service: basic FastAPI app with health check endpoint (`GET /health`)
- [ ] Worker: basic Go process that connects to Redis and logs "worker started"
- [ ] Frontend: Next.js app with a landing page
- [ ] Verify all services start with `docker-compose up` and can reach each other
- [ ] Create `.env.example` with all required environment variables
- [ ] Create `Makefile` with common commands (`make up`, `make down`, `make logs`, `make migrate`)

### Deliverables

- All services start and respond to health checks
- Database is running with migrations applied
- Redis is running and accessible
- `docker-compose up --build` works end to end

---

## Phase 2: Authentication & Resume Upload

**Goal:** Users can register, log in, and upload resumes. The backend handles auth properly with JWT.

**Duration estimate:** 3–4 days

### Tasks

- [ ] Database migrations: `users` table, `resumes` table
- [ ] API Gateway: user registration endpoint (`POST /api/auth/register`)
- [ ] API Gateway: user login endpoint (`POST /api/auth/login`)
- [ ] API Gateway: JWT middleware for protected routes
- [ ] API Gateway: token refresh endpoint (`POST /api/auth/refresh`)
- [ ] API Gateway: resume upload endpoint (`POST /api/resumes`) — accepts PDF
- [ ] API Gateway: resume list/delete endpoints
- [ ] AI Service: resume parsing endpoint (`POST /ai/parse-resume`) — extracts text from PDF
- [ ] API Gateway calls AI Service to parse resume on upload, stores parsed text
- [ ] Frontend: registration page
- [ ] Frontend: login page with token storage
- [ ] Frontend: resume upload page with file picker
- [ ] Frontend: display list of uploaded resumes

### Deliverables

- End-to-end auth flow working (register → login → access protected routes)
- Resume upload, parsing, and storage working
- Frontend connected to backend for auth and resume management

---

## Phase 3: Job Applications & AI Email Generation

**Goal:** Users can create job applications and get AI-generated cold emails.

**Duration estimate:** 4–5 days

### Tasks

- [ ] Database migration: `applications` table, `emails` table
- [ ] API Gateway: application CRUD endpoints
  - `POST /api/applications` — create
  - `GET /api/applications` — list (with filters)
  - `GET /api/applications/:id` — detail
  - `PUT /api/applications/:id` — update
  - `DELETE /api/applications/:id` — delete
- [ ] AI Service: email generation endpoint (`POST /ai/generate-email`)
  - Accepts: resume text, job description, company name, role, tone preference
  - Returns: subject, body, match score, key points
- [ ] AI Service: implement prompt engineering for email generation
  - Design system prompt for cold email generation
  - Handle different tones (formal, friendly, concise)
  - Ensure emails are personalized, not generic
- [ ] API Gateway: email generation trigger (`POST /api/applications/:id/generate-email`)
  - Fetches resume and application data
  - Calls AI Service
  - Stores generated email in database
- [ ] API Gateway: email endpoints
  - `GET /api/applications/:id/email` — get generated email
  - `PUT /api/emails/:id` — update/edit email
  - `POST /api/applications/:id/regenerate-email` — regenerate
- [ ] Frontend: new application form (company, role, email, JD, link)
- [ ] Frontend: application list page with status badges
- [ ] Frontend: application detail page showing generated email
- [ ] Frontend: email editor with inline editing
- [ ] Frontend: regenerate button with tone selector

### Deliverables

- Full application CRUD working
- AI generates personalized emails based on resume + JD
- User can review, edit, and regenerate emails
- Frontend displays the full workflow

---

## Phase 4: Email Scheduling & Background Worker

**Goal:** Users can schedule emails. The worker sends them at the right time.

**Duration estimate:** 3–4 days

### Tasks

- [ ] API Gateway: schedule endpoint (`POST /api/emails/:id/schedule`)
  - Accepts: `send_at` timestamp or relative delay
  - Updates email status to "scheduled"
  - Enqueues job in Redis sorted set (score = send_at unix timestamp)
- [ ] API Gateway: cancel schedule (`DELETE /api/emails/:id/schedule`)
- [ ] API Gateway: reschedule (`PUT /api/emails/:id/schedule`)
- [ ] API Gateway: list scheduled emails (`GET /api/emails?status=scheduled`)
- [ ] Worker: implement Redis queue consumer
  - Poll sorted set for jobs where score <= now
  - Process jobs concurrently with worker pool
- [ ] Worker: implement email sender
  - SMTP sender implementation
  - Alternative: Resend API sender
  - Configurable via environment variable
- [ ] Worker: implement retry logic
  - Max 3 retries with exponential backoff
  - Move to dead letter queue after max retries
  - Update email status (sent / failed)
- [ ] Worker: update application status in DB after successful send
- [ ] Frontend: schedule picker component (date/time picker + quick options)
- [ ] Frontend: scheduled emails list view
- [ ] Frontend: cancel/reschedule functionality

### Deliverables

- Emails can be scheduled and are sent automatically at the correct time
- Failed sends are retried
- Email and application statuses update correctly
- Frontend provides scheduling UI

---

## Phase 5: Application Tracking & Dashboard

**Goal:** Users can track application progress and see analytics on a dashboard.

**Duration estimate:** 3–4 days

### Tasks

- [ ] API Gateway: status update endpoint (`PATCH /api/applications/:id/status`)
  - Valid transitions: draft → applied → replied → interview → offer/rejected/ghosted
- [ ] API Gateway: status history (store each status change with timestamp)
- [ ] API Gateway: analytics endpoints
  - `GET /api/analytics/summary` — total counts by status
  - `GET /api/analytics/timeline` — applications over time
  - `GET /api/analytics/rates` — response rate, interview rate
  - `GET /api/analytics/recent` — recent activity
- [ ] Frontend: status update dropdown on application detail page
- [ ] Frontend: dashboard page
  - Stats cards (total applications, response rate, interviews, etc.)
  - Status breakdown chart (pie or bar)
  - Applications over time chart (line)
  - Recent activity feed
- [ ] Frontend: filter applications by status on list page
- [ ] Frontend: search applications by company/role

### Deliverables

- Application status tracking with full lifecycle
- Analytics dashboard with charts and stats
- Filtering and search on application list

---

## Phase 6: Polish, Testing & Hardening

**Goal:** Production-ready quality — error handling, tests, validation, documentation.

**Duration estimate:** 3–5 days

### Tasks

- [ ] API Gateway: input validation on all endpoints
- [ ] API Gateway: proper error responses (consistent error format)
- [ ] API Gateway: rate limiting middleware
- [ ] API Gateway: request logging with structured logs
- [ ] API Gateway: unit tests for services and handlers
- [ ] API Gateway: integration tests for key flows
- [ ] AI Service: unit tests for parser, matcher, generator
- [ ] AI Service: mock LLM responses for testing
- [ ] Worker: unit tests for consumer and sender
- [ ] Worker: graceful shutdown handling
- [ ] Frontend: loading states, error states, empty states
- [ ] Frontend: responsive design review
- [ ] Frontend: toast notifications for actions
- [ ] Docker: production-optimized Dockerfiles (multi-stage builds)
- [ ] Documentation: API documentation (OpenAPI/Swagger)
- [ ] Documentation: update README with final setup instructions
- [ ] Security review: ensure no secrets in code, proper CORS, input sanitization

### Deliverables

- Comprehensive test coverage on critical paths
- Polished UI with proper error handling
- Production-ready Docker setup
- Complete documentation

---

## Phase Summary

| Phase | Focus | Est. Duration |
|-------|-------|---------------|
| 1 | Foundation & Infrastructure | 2–3 days |
| 2 | Authentication & Resume Upload | 3–4 days |
| 3 | Job Applications & AI Email Generation | 4–5 days |
| 4 | Email Scheduling & Background Worker | 3–4 days |
| 5 | Application Tracking & Dashboard | 3–4 days |
| 6 | Polish, Testing & Hardening | 3–5 days |
| **Total** | | **18–25 days** |

Each phase builds on the previous one. At the end of each phase, the system is functional and demonstrable.
