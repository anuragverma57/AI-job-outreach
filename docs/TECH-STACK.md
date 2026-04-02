# Technology Stack

## Overview


| Layer            | Technology              | Purpose                                           |
| ---------------- | ----------------------- | ------------------------------------------------- |
| API Gateway      | Go                      | Core backend API, routing, business logic         |
| AI Service       | Python + FastAPI        | LLM integration, resume parsing, email generation |
| Worker Service   | Go                      | Background jobs, email sending, scheduling        |
| Frontend         | Next.js (React)         | Dashboard UI, application management              |
| Database         | PostgreSQL              | Primary data store                                |
| Queue & Cache    | Redis                   | Job queues, caching, pub/sub                      |
| Containerization | Docker + Docker Compose | **Postgres + Redis** in dev; app processes via Makefile |


---

## Backend: Go

**Version:** 1.22+

**Used for:** API Gateway and **Worker** (second `main` in `api-gateway/cmd/worker`)

**Why Go:**

- Excellent performance for HTTP servers and concurrent workloads
- Goroutines make background worker implementation natural
- Strong standard library reduces dependency on third-party packages
- Compiles to a single static binary — simple Docker images
- Great fit for building API servers and long-running background processes

**Key Libraries:**


| Library                  | Purpose                                       |
| ------------------------ | --------------------------------------------- |
| `gofiber/fiber`          | HTTP framework (fast, Express-inspired)       |
| `pgx`                    | PostgreSQL driver (high performance, pure Go) |
| `go-redis/redis`         | Redis client                                  |
| `golang-jwt/jwt`         | JWT authentication                            |
| `golang-migrate/migrate` | Database migrations                           |
| `viper`                  | Configuration management                      |
| `zap` or `zerolog`       | Structured logging                            |
| `gomail` or `net/smtp`   | Email sending                                 |
| `testify`                | Testing assertions                            |


---

## Backend: Python

**Version:** 3.11+

**Used for:** AI Service

**Why Python:**

- FastAPI for HTTP APIs and auto-generated docs
- `pdfplumber` for PDF text extraction
- `httpx` to call **any OpenAI-compatible** chat completions endpoint (local Ollama shim, vLLM, cloud APIs)

**Key Libraries:**


| Library                  | Purpose                                       |
| ------------------------ | --------------------------------------------- |
| `fastapi`                | HTTP framework                                |
| `uvicorn`                | ASGI server                                   |
| `pydantic`               | Request/response validation and serialization |
| `pdfplumber`             | PDF resume parsing                            |
| `python-dotenv`          | Environment variable management               |
| `httpx`                  | Sync HTTP client → `LLM_BASE_URL/chat/completions` |
| `pytest`                 | Testing (optional)                            |


---

## Frontend: Next.js

**Version:** 14+ (App Router)

**Why Next.js:**

- React-based with SSR and file-based routing out of the box
- API routes available if any frontend-side API logic is needed
- Great developer experience with TypeScript support
- Large ecosystem for UI components

**Key Libraries:**


| Library                  | Purpose                         |
| ------------------------ | ------------------------------- |
| `next`                   | Framework                       |
| `typescript`             | Type safety                     |
| `tailwindcss`            | Utility-first CSS               |
| `shadcn/ui`              | Pre-built accessible components |
| `react-query` / `swr`    | Data fetching and caching       |
| `recharts` or `chart.js` | Dashboard charts and analytics  |
| `react-hook-form`        | Form management                 |
| `zod`                    | Schema validation               |
| `axios` or `fetch`       | HTTP client                     |


---

## Database: PostgreSQL

**Version:** 16+

**Why PostgreSQL:**

- Robust relational database with strong ACID compliance
- JSON/JSONB support for flexible fields (parsed resume data, AI responses)
- Full-text search capabilities (useful for JD analysis)
- Widely used in production — transferable knowledge
- Excellent tooling and monitoring

**Usage:**

- Primary data store for users, applications, emails, resumes, analytics
- Stores parsed resume text and AI-generated email drafts
- Handles application status tracking and history

---

## Queue & Cache: Redis

**Version:** 7+

**Why Redis:**

- In-memory store with sub-millisecond latency
- Reliable job queue using sorted sets (for delayed/scheduled jobs) or Redis Streams
- Caching layer to reduce redundant AI calls
- Simple pub/sub if real-time features are added later

**Usage:**


| Use Case                  | Redis Feature                             |
| ------------------------- | ----------------------------------------- |
| Email job queue           | Sorted sets (score = scheduled timestamp) |
| Background job processing | List-based queue (LPUSH/BRPOP)            |
| Caching AI responses      | Key-value with TTL                        |
| Rate limiting             | Sliding window counters                   |


---

## Containerization: Docker + Docker Compose

**Current repo:** `docker-compose.yml` runs **postgres** (published as **5433** on the host) and **redis** (**6379**). API gateway, worker, AI service, and frontend are started with **`Makefile`** targets on the host (no Dockerfiles required for daily dev).

**Possible future layout** (not all wired in compose today):


| Service       | Typical image        | Port  |
| ------------- | -------------------- | ----- |
| `postgres`    | `postgres:16-alpine` | 5433→5432 |
| `redis`       | `redis:7-alpine`     | 6379  |
| `api-gateway` | Go build             | 8080  |
| `ai-service`  | Python slim          | 8000  |
| `worker`      | Go build (same module as gateway) | — |
| `frontend`    | Node                 | 3000  |


---

## Email Sending

**Options (pick one):**


| Option                        | Pros                                        | Cons                          |
| ----------------------------- | ------------------------------------------- | ----------------------------- |
| **SMTP (Gmail App Password)** | Free, simple setup                          | Rate limits, may land in spam |
| **Resend**                    | Developer-friendly API, good deliverability | Paid after free tier          |
| **SendGrid**                  | Industry standard, analytics built-in       | More complex setup            |
| **Amazon SES**                | Cheapest at scale                           | AWS account required          |


**Recommendation:** Start with **Resend** (100 emails/day free) or Gmail SMTP for development, switch to SendGrid/SES for production.

---

## AI / LLM

**Configured in this project:** any server that implements **OpenAI Chat Completions** JSON over HTTP:

- `POST {LLM_BASE_URL}/chat/completions`
- Response: `choices[0].message.content` (string the model fills with JSON for the email payload)

Environment variables: `LLM_BASE_URL` (no trailing slash; usually ends in `/v1`), `LLM_MODEL`, optional `LLM_API_KEY`.

**Examples:** OpenAI-compatible proxy in front of **Ollama**, **vLLM**, or hosted APIs (OpenAI, Groq, etc.).

**Operational note:** Smaller local models may return invalid or empty JSON; the AI service includes parsing helpers and clear errors — tune prompts or model size for reliability.

---

## Development Tools


| Tool                   | Purpose                                       |
| ---------------------- | --------------------------------------------- |
| `air` (Go)             | Hot reload for Go services during development |
| `uvicorn --reload`     | Hot reload for FastAPI                        |
| `docker compose watch` | Auto-rebuild containers on code changes       |
| `pgAdmin` or `DBeaver` | Database GUI                                  |
| `RedisInsight`         | Redis GUI                                     |
| `Postman` or `Bruno`   | API testing                                   |
| `golangci-lint`        | Go linting                                    |
| `ruff`                 | Python linting                                |
| `eslint` + `prettier`  | Frontend linting and formatting               |


