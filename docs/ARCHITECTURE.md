# System Architecture

## Architecture Style: Modular Microservices

The system uses a **lightweight microservices architecture** with three backend services, a frontend, and shared infrastructure. This isn't microservices for the sake of complexity — each service has a clear boundary and a distinct runtime requirement:

- **Go** is ideal for the API gateway and worker (concurrency, performance, static binary)
- **Python** is ideal for AI/ML work (LLM libraries, FastAPI, ecosystem)

Splitting them lets each service use the best tool for its job and scale independently.

## High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                          FRONTEND (Next.js)                          │
│                         http://localhost:3000                         │
└──────────────────────────────┬───────────────────────────────────────┘
                               │ REST API calls
                               ▼
┌──────────────────────────────────────────────────────────────────────┐
│                        API GATEWAY (Go)                              │
│                       http://localhost:8080                           │
│                                                                      │
│  Responsibilities:                                                   │
│  ├── Authentication & session management                             │
│  ├── Job application CRUD                                            │
│  ├── Resume upload & storage                                         │
│  ├── Email review & approval                                         │
│  ├── Schedule management                                             │
│  ├── Application status tracking                                     │
│  ├── Analytics aggregation                                           │
│  └── Enqueue jobs to Redis                                           │
└─────────┬───────────────────┬────────────────────┬───────────────────┘
          │                   │                    │
          │ HTTP              │ Redis Queue        │ SQL
          ▼                   ▼                    ▼
┌─────────────────┐  ┌──────────────┐  ┌───────────────────┐
│  AI SERVICE      │  │    REDIS     │  │   POSTGRESQL      │
│  (Python/FastAPI)│  │              │  │                   │
│  :8000           │  │  :6379       │  │  :5432            │
│                  │  │              │  │                   │
│  ├── Resume      │  │  ├── Email   │  │  ├── users        │
│  │   parsing     │  │  │   queue   │  │  ├── resumes      │
│  ├── JD analysis │  │  ├── Job     │  │  ├── jobs         │
│  ├── Email       │  │  │   queue   │  │  ├── emails       │
│  │   generation  │  │  └── Cache   │  │  ├── schedules    │
│  └── LLM calls   │  │              │  │  └── analytics    │
└─────────────────┘  └──────┬───────┘  └───────────────────┘
                            │
                            │ Consume jobs
                            ▼
                   ┌──────────────────┐
                   │  WORKER SERVICE  │
                   │  (Go)            │
                   │                  │
                   │  ├── Email       │
                   │  │   sender      │
                   │  ├── Scheduler   │
                   │  │   (cron)      │
                   │  └── Retry       │
                   │      logic       │
                   └──────────────────┘
```

## Service Details

### 1. API Gateway (Go)

The central backend service. All client requests go through here.

**Why Go:** High-performance HTTP server, excellent concurrency model for handling many simultaneous requests, strong standard library for building APIs.

| Responsibility | Details |
|---|---|
| Authentication | JWT-based auth, user registration and login |
| Job CRUD | Create, read, update, delete job applications |
| Resume Management | Upload, store, and retrieve resumes |
| Email Workflow | Trigger AI generation, store drafts, handle user edits |
| Scheduling | Accept schedule requests, enqueue to Redis at the right time |
| Tracking | Update application statuses (sent, replied, interview, rejected) |
| Analytics | Aggregate stats for the dashboard |

**Key Design Decisions:**
- Uses the `Fiber` framework for HTTP routing
- Talks to PostgreSQL via `pgx` or `sqlx`
- Talks to Redis via `go-redis`
- Calls AI Service over internal HTTP (service-to-service)
- Enqueues background jobs to Redis for the Worker

### 2. AI Service (Python / FastAPI)

Handles all AI and LLM-related work. This is a stateless service — it receives a request, processes it, and returns a result.

**Why Python:** LLM libraries (OpenAI SDK, LangChain, etc.) have first-class Python support. FastAPI provides async performance with clean API design.

| Responsibility | Details |
|---|---|
| Resume Parsing | Extract structured data from uploaded resume (text/PDF) |
| JD Analysis | Parse and extract key requirements from a job description |
| Match Analysis | Compare resume against JD, identify relevant skills/experience |
| Email Generation | Generate a personalized cold email using the match analysis |
| Regeneration | Re-generate with different tone/emphasis if user requests |

**Key Design Decisions:**
- Stateless — no database access, receives all needed data in the request
- Uses OpenAI API (or compatible LLM) for generation
- Prompt engineering handled here — prompts are version-controlled
- Returns structured JSON responses

### 3. Worker Service (Go)

A background process that consumes jobs from Redis queues and executes them.

**Why Go:** Long-running process with concurrent job processing. Go's goroutines make this efficient.

| Responsibility | Details |
|---|---|
| Email Sending | Consume email jobs from Redis, send via SMTP/email API |
| Scheduling | Periodically check for emails that are due to be sent |
| Retry Logic | Retry failed sends with exponential backoff |
| Status Updates | Update email/application status in PostgreSQL after send |

**Key Design Decisions:**
- Runs as a separate container, always on
- Uses a ticker/cron to poll for scheduled emails
- Processes jobs concurrently using goroutines with a worker pool
- Dead letter queue for permanently failed jobs

## Service Communication

```
Frontend  ──HTTP/REST──►  API Gateway  ──HTTP──►  AI Service
                              │
                              │──SQL──►  PostgreSQL
                              │
                              │──Redis──►  Redis (enqueue)
                              │
Worker  ◄──Redis──  Redis (dequeue)
   │
   │──SQL──►  PostgreSQL (update status)
   │
   │──SMTP──►  Email Provider
```

### Communication Patterns

| From | To | Protocol | Pattern |
|---|---|---|---|
| Frontend → API Gateway | REST/HTTP | Synchronous request-response |
| API Gateway → AI Service | HTTP | Synchronous (internal network) |
| API Gateway → PostgreSQL | TCP (SQL) | Direct connection |
| API Gateway → Redis | TCP | Enqueue jobs |
| Worker → Redis | TCP | Dequeue/consume jobs |
| Worker → PostgreSQL | TCP (SQL) | Update statuses |
| Worker → Email Provider | SMTP/API | Send emails |

### Why Not gRPC?

REST over HTTP between Go and Python keeps things simple for a system this size. gRPC adds complexity (protobuf, code generation) that isn't justified when there are only two services talking to each other. If the system grows to 5+ services, gRPC would be worth reconsidering.

## Data Flow: End-to-End

### Flow 1: Create Application & Generate Email

```
1. User fills in job details (company, role, recruiter email, JD, link)
2. Frontend POST /api/applications → API Gateway
3. API Gateway stores the application in PostgreSQL (status: "draft")
4. API Gateway calls AI Service: POST /ai/generate-email
   - Sends: resume text, job description, company name, role
5. AI Service analyzes resume vs JD
6. AI Service generates personalized cold email
7. AI Service returns: { subject, body, match_score, key_points }
8. API Gateway stores generated email in PostgreSQL (status: "draft")
9. API Gateway returns email to Frontend for user review
```

### Flow 2: Schedule & Send Email

```
1. User reviews/edits the email, picks a send time
2. Frontend POST /api/emails/{id}/schedule → API Gateway
3. API Gateway updates email status to "scheduled" with send_at timestamp
4. API Gateway enqueues a delayed job in Redis
5. Worker picks up the job when send_at time arrives
6. Worker sends the email via SMTP
7. Worker updates email status in PostgreSQL to "sent"
8. Worker updates application status to "applied"
```

### Flow 3: Track Application

```
1. User manually updates status (got reply, interview, rejection)
   - Or: future webhook integration auto-detects replies
2. Frontend PATCH /api/applications/{id}/status → API Gateway
3. API Gateway updates status in PostgreSQL
4. Dashboard queries aggregate stats via GET /api/analytics
```

## Database Design (High-Level)

```
users
├── id (UUID)
├── email
├── password_hash
├── name
└── created_at

resumes
├── id (UUID)
├── user_id (FK → users)
├── file_path
├── parsed_text
└── uploaded_at

applications
├── id (UUID)
├── user_id (FK → users)
├── resume_id (FK → resumes)
├── company_name
├── role
├── recruiter_email
├── job_description
├── job_link
├── status (draft | applied | replied | interview | rejected | ghosted)
├── created_at
└── updated_at

emails
├── id (UUID)
├── application_id (FK → applications)
├── subject
├── body
├── status (draft | scheduled | sending | sent | failed)
├── scheduled_at
├── sent_at
├── retry_count
├── created_at
└── updated_at
```

## Error Handling & Resilience

| Scenario | Handling |
|---|---|
| AI Service is down | API Gateway returns 503, frontend shows retry option |
| Email send fails | Worker retries with exponential backoff (max 3 retries) |
| Redis is down | API Gateway falls back to direct DB polling for scheduled jobs |
| LLM API rate limited | AI Service queues internally, returns 429 to gateway |
| Database connection lost | Connection pooling with automatic reconnection |

## Security Considerations

- JWT tokens for authentication with expiry and refresh
- Passwords hashed with bcrypt
- API rate limiting on the gateway
- Environment variables for all secrets (API keys, SMTP credentials)
- Internal services (AI, Worker) not exposed to the public network
- Input validation and sanitization on all endpoints
- CORS configuration for frontend origin only
