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
| Containerization | Docker + Docker Compose | Local development and deployment                  |


---

## Backend: Go

**Version:** 1.22+

**Used for:** API Gateway, Worker Service

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

- First-class support for OpenAI SDK and LLM tooling
- FastAPI provides async performance with auto-generated API docs
- Rich ecosystem for text processing (resume parsing, NLP)
- Fastest path to LLM integration

**Key Libraries:**


| Library                  | Purpose                                       |
| ------------------------ | --------------------------------------------- |
| `fastapi`                | Async HTTP framework                          |
| `uvicorn`                | ASGI server                                   |
| `openai`                 | OpenAI API client                             |
| `pydantic`               | Request/response validation and serialization |
| `PyPDF2` or `pdfplumber` | PDF resume parsing                            |
| `python-dotenv`          | Environment variable management               |
| `httpx`                  | Async HTTP client (if needed)                 |
| `pytest`                 | Testing                                       |


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

**Why Docker:**

- Consistent environments across development machines
- Each service runs in isolation with its own dependencies
- Docker Compose orchestrates all services with a single command
- Mirrors production deployment patterns

**Container Layout:**


| Container     | Base Image           | Exposed Port |
| ------------- | -------------------- | ------------ |
| `api-gateway` | `golang:1.22-alpine` | 8080         |
| `ai-service`  | `python:3.11-slim`   | 8000         |
| `worker`      | `golang:1.22-alpine` | — (no port)  |
| `frontend`    | `node:20-alpine`     | 3000         |
| `postgres`    | `postgres:16-alpine` | 5432         |
| `redis`       | `redis:7-alpine`     | 6379         |


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

**Primary:** OpenAI API (GPT-4o or GPT-4o-mini)

**Why:**

- Best-in-class for structured text generation
- Reliable API with good rate limits
- JSON mode for structured responses

**Alternatives (if needed):**

- Anthropic Claude — strong at following instructions
- Local models via Ollama — free, private, but lower quality
- Groq — fast inference for open-source models

**Recommendation:** Start with OpenAI `gpt-4o-mini` (cheap, fast, good enough). Switch to `gpt-4o` or Claude for better quality if needed.

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


