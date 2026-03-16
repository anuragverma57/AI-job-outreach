# Project Structure

## Repository Layout

This is a **monorepo** containing all services. Each service is independently buildable and deployable via its own Dockerfile, but they live together for easier development.

```
AI-job-outreach/
│
├── README.md
├── docker-compose.yml
├── .env.example
├── .gitignore
├── Makefile
│
├── docs/
│   ├── CORE-IDEA.md
│   ├── ARCHITECTURE.md
│   ├── TECH-STACK.md
│   ├── FEATURES.md
│   ├── PROJECT-STRUCTURE.md
│   ├── PHASES.md
│   └── IMPLEMENTATION-GUIDE.md
│
├── api-gateway/                    # Go — Core Backend API
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── server/
│   │       └── main.go             # Entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go           # App configuration (env vars, defaults)
│   │   ├── middleware/
│   │   │   ├── auth.go             # JWT authentication middleware
│   │   │   ├── cors.go             # CORS configuration
│   │   │   ├── logger.go           # Request logging middleware
│   │   │   └── ratelimit.go        # Rate limiting
│   │   ├── handler/
│   │   │   ├── auth.go             # Register, login, refresh handlers
│   │   │   ├── application.go      # Application CRUD handlers
│   │   │   ├── resume.go           # Resume upload/management handlers
│   │   │   ├── email.go            # Email generation, review, schedule handlers
│   │   │   └── analytics.go        # Dashboard data handlers
│   │   ├── service/
│   │   │   ├── auth.go             # Auth business logic
│   │   │   ├── application.go      # Application business logic
│   │   │   ├── resume.go           # Resume business logic
│   │   │   ├── email.go            # Email workflow orchestration
│   │   │   └── analytics.go        # Analytics computation
│   │   ├── repository/
│   │   │   ├── user.go             # User DB queries
│   │   │   ├── application.go      # Application DB queries
│   │   │   ├── resume.go           # Resume DB queries
│   │   │   └── email.go            # Email DB queries
│   │   ├── model/
│   │   │   ├── user.go             # User struct and types
│   │   │   ├── application.go      # Application struct and types
│   │   │   ├── resume.go           # Resume struct and types
│   │   │   └── email.go            # Email struct and types
│   │   ├── queue/
│   │   │   └── redis.go            # Redis queue producer (enqueue jobs)
│   │   ├── client/
│   │   │   └── ai_client.go        # HTTP client for AI Service
│   │   └── router/
│   │       └── router.go           # Route definitions
│   ├── migrations/
│   │   ├── 000001_create_users.up.sql
│   │   ├── 000001_create_users.down.sql
│   │   ├── 000002_create_resumes.up.sql
│   │   ├── 000002_create_resumes.down.sql
│   │   ├── 000003_create_applications.up.sql
│   │   ├── 000003_create_applications.down.sql
│   │   ├── 000004_create_emails.up.sql
│   │   └── 000004_create_emails.down.sql
│   └── tests/
│       ├── handler_test.go
│       ├── service_test.go
│       └── repository_test.go
│
├── ai-service/                     # Python — AI & LLM Service
│   ├── Dockerfile
│   ├── requirements.txt
│   ├── pyproject.toml
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py                 # FastAPI entry point
│   │   ├── config.py               # Settings and env vars
│   │   ├── routers/
│   │   │   ├── __init__.py
│   │   │   ├── health.py           # Health check endpoint
│   │   │   └── generation.py       # Email generation endpoints
│   │   ├── services/
│   │   │   ├── __init__.py
│   │   │   ├── resume_parser.py    # PDF text extraction
│   │   │   ├── jd_analyzer.py      # Job description analysis
│   │   │   ├── matcher.py          # Resume-JD matching logic
│   │   │   └── email_generator.py  # LLM-based email generation
│   │   ├── prompts/
│   │   │   ├── __init__.py
│   │   │   ├── email_prompt.py     # Prompt templates for email generation
│   │   │   └── analysis_prompt.py  # Prompt templates for resume/JD analysis
│   │   ├── models/
│   │   │   ├── __init__.py
│   │   │   ├── requests.py         # Pydantic request models
│   │   │   └── responses.py        # Pydantic response models
│   │   └── utils/
│   │       ├── __init__.py
│   │       └── text.py             # Text processing utilities
│   └── tests/
│       ├── __init__.py
│       ├── test_resume_parser.py
│       ├── test_email_generator.py
│       └── test_matcher.py
│
├── worker/                         # Go — Background Worker
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── worker/
│   │       └── main.go             # Entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go           # Worker configuration
│   │   ├── consumer/
│   │   │   └── email_consumer.go   # Redis queue consumer
│   │   ├── sender/
│   │   │   ├── smtp.go             # SMTP email sender
│   │   │   └── resend.go           # Resend API sender (alternative)
│   │   ├── scheduler/
│   │   │   └── scheduler.go        # Polls for due scheduled emails
│   │   ├── repository/
│   │   │   └── email.go            # DB queries for status updates
│   │   └── pool/
│   │       └── worker_pool.go      # Concurrent worker pool management
│   └── tests/
│       ├── consumer_test.go
│       └── sender_test.go
│
├── frontend/                       # Next.js — Dashboard UI
│   ├── Dockerfile
│   ├── package.json
│   ├── tsconfig.json
│   ├── next.config.js
│   ├── tailwind.config.ts
│   ├── postcss.config.js
│   ├── public/
│   │   └── ...
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx            # Dashboard home
│   │   │   ├── login/
│   │   │   │   └── page.tsx
│   │   │   ├── register/
│   │   │   │   └── page.tsx
│   │   │   ├── applications/
│   │   │   │   ├── page.tsx        # Application list
│   │   │   │   ├── new/
│   │   │   │   │   └── page.tsx    # New application form
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx    # Application detail + email
│   │   │   ├── resumes/
│   │   │   │   └── page.tsx        # Resume management
│   │   │   └── analytics/
│   │   │       └── page.tsx        # Analytics dashboard
│   │   ├── components/
│   │   │   ├── ui/                 # shadcn/ui components
│   │   │   ├── layout/
│   │   │   │   ├── sidebar.tsx
│   │   │   │   ├── header.tsx
│   │   │   │   └── footer.tsx
│   │   │   ├── applications/
│   │   │   │   ├── application-card.tsx
│   │   │   │   ├── application-form.tsx
│   │   │   │   └── status-badge.tsx
│   │   │   ├── email/
│   │   │   │   ├── email-preview.tsx
│   │   │   │   ├── email-editor.tsx
│   │   │   │   └── schedule-picker.tsx
│   │   │   └── analytics/
│   │   │       ├── stats-cards.tsx
│   │   │       ├── status-chart.tsx
│   │   │       └── activity-feed.tsx
│   │   ├── lib/
│   │   │   ├── api.ts              # API client
│   │   │   ├── auth.ts             # Auth utilities
│   │   │   └── utils.ts            # General utilities
│   │   ├── hooks/
│   │   │   ├── use-applications.ts
│   │   │   ├── use-auth.ts
│   │   │   └── use-analytics.ts
│   │   └── types/
│   │       ├── application.ts
│   │       ├── email.ts
│   │       └── user.ts
│   └── ...
│
└── scripts/
    ├── seed.sql                    # Sample data for development
    └── setup.sh                    # One-time setup script
```

## Key Conventions

### Go Services (api-gateway, worker)

- Follow standard Go project layout (`cmd/`, `internal/`)
- `cmd/` contains entry points only — minimal code
- `internal/` contains all business logic — not importable by external packages
- Layers: **handler** → **service** → **repository** (clean separation)
- Each layer has its own interface for testability

### Python Service (ai-service)

- FastAPI app structure with routers, services, and models
- Prompts are isolated in their own module for easy iteration
- Pydantic models for all request/response validation
- Stateless design — no database access from this service

### Frontend (Next.js)

- App Router (Next.js 14+)
- Components organized by feature, not by type
- Shared UI components in `components/ui/`
- API client centralized in `lib/api.ts`
- Custom hooks for data fetching patterns

### Shared Patterns

- Environment variables via `.env` files (never committed)
- Each service has its own `Dockerfile`
- `docker-compose.yml` at the root wires everything together
- Database migrations managed by the API Gateway service
- `Makefile` for common development commands
