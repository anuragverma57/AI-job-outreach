# Project Structure

Monorepo layout **as implemented** in this repository. For a feature-level snapshot see [CURRENT-STATE.md](CURRENT-STATE.md).

```
AI-job-outreach/
├── README.md
├── Makefile                    # up, migrate-*, run-api, run-worker, run-ai, dev
├── docker-compose.yml          # postgres + redis only (dev)
├── .env.example
├── .gitignore
│
├── docs/                       # Architecture & roadmap docs
│
├── api-gateway/                # Go module: HTTP API + shared internal packages
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   ├── server/
│   │   │   └── main.go         # API gateway entry
│   │   └── worker/
│   │       └── main.go         # Background worker entry (same module)
│   ├── internal/
│   │   ├── config/
│   │   ├── database/
│   │   ├── middleware/         # e.g. auth
│   │   ├── handler/            # auth, health, resume, application, email
│   │   ├── service/
│   │   ├── repository/
│   │   ├── model/
│   │   ├── client/             # ai_client.go → FastAPI AI service
│   │   ├── queue/              # Redis scheduled-email queue
│   │   ├── sender/             # SMTP (worker)
│   │   └── router/
│   └── migrations/
│       ├── 000001_create_users.up.sql
│       ├── 000002_create_refresh_tokens.up.sql
│       ├── 000003_create_resumes.up.sql
│       ├── 000004_create_applications.up.sql
│       └── 000005_create_emails.up.sql
│
├── ai-service/                 # Python FastAPI
│   ├── requirements.txt
│   └── app/
│       ├── main.py
│       ├── config.py           # LLM_BASE_URL, LLM_MODEL, …
│       ├── routers/
│       │   ├── health.py       # GET /health
│       │   └── generation.py   # /ai/parse-resume, /ai/generate-email
│       ├── services/
│       │   ├── resume_parser.py
│       │   ├── email_generator.py
│       │   └── llm_response.py # parse OpenAI-shaped responses + JSON email payload
│       ├── prompts/
│       │   └── email_prompt.py
│       └── models/
│           ├── requests.py
│           └── responses.py
│
└── frontend/                   # Next.js (App Router)
    ├── package.json
    └── src/
        ├── app/
        │   ├── page.tsx        # redirect to /dashboard
        │   ├── layout.tsx
        │   ├── (auth)/         # login, register
        │   └── (dashboard)/    # layout, dashboard, resumes, applications, emails
        ├── components/         # ui (shadcn), layout, applications/…
        ├── hooks/              # use-auth, …
        ├── lib/
        │   └── api.ts
        └── types/
```

There is **no** top-level `worker/` Go module — the worker lives under **`api-gateway/cmd/worker`**.

**Not in tree above:** optional `scripts/`, Postman exports, local `venv/`, `uploads/`, build artifacts (`api-gateway/bin/`, compiled binaries — gitignored).

## Conventions

- **Go:** `handler` → `service` → `repository`; shared DB pool in `server` and `worker` mains.
- **Python:** Stateless AI service; secrets only via env.
- **Frontend:** `credentials: "include"` for cookie-based auth; `NEXT_PUBLIC_API_URL` → gateway.

## Docker

Root Compose file does **not** build the Go/Python/Node apps yet. Use `make up` for data stores, then `make run-api`, `make run-ai`, `make run-worker`, and `npm run dev` in `frontend/`.
