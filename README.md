# AI Job Outreach Automation Platform

An intelligent job application automation system that helps candidates streamline their job hunting workflow — from analyzing job descriptions and generating personalized cold emails to scheduling sends and tracking application progress.

## What This Project Does

Job hunting involves a lot of repetitive manual work: reading job descriptions, crafting emails, sending them at the right time, and tracking responses. This platform automates most of that pipeline.

**Input a job opportunity** → **AI generates a personalized cold email** → **Review & schedule** → **Auto-send** → **Track progress on a dashboard**

## Key Capabilities

- **Resume & JD Analysis** — AI compares your resume against the job description to identify relevant skills and experience
- **Cold Email Generation** — Generates personalized, professional outreach emails using LLMs
- **Review & Edit** — Full control to review, edit, or regenerate emails before sending
- **Scheduled Sending** — Queue emails to be sent at optimal times
- **Application Tracking** — Track statuses: sent, replied, interview scheduled, rejected
- **Analytics Dashboard** — Visual overview of your job search pipeline

## Architecture Overview

The system follows a **modular microservices architecture** with three core services:

| Service | Language | Responsibility |
|---------|----------|----------------|
| **API Gateway** | Go | REST API, auth, routing, scheduling, application tracking |
| **AI Service** | Python (FastAPI) | Resume parsing, JD analysis, email generation via LLM |
| **Worker Service** | Go | Background job processing — sends scheduled emails |

Supporting infrastructure: **PostgreSQL** (persistence), **Redis** (job queue + caching), **React/Next.js** (frontend dashboard).

All services are containerized with **Docker** and orchestrated via **Docker Compose**.

## Documentation

| Document | Description |
|----------|-------------|
| [Core Idea](docs/CORE-IDEA.md) | Problem statement, goals, and vision |
| [Architecture](docs/ARCHITECTURE.md) | System design, service communication, data flow |
| [Tech Stack](docs/TECH-STACK.md) | Technologies used and rationale |
| [Features](docs/FEATURES.md) | Complete feature list with priorities |
| [Project Structure](docs/PROJECT-STRUCTURE.md) | Repository layout and folder organization |
| [Phases](docs/PHASES.md) | Development roadmap broken into phases |
| [Implementation Guide](docs/IMPLEMENTATION-GUIDE.md) | Step-by-step build guide |

## Quick Start

> Implementation details coming soon. See [Phases](docs/PHASES.md) for the development roadmap.

```bash
# Clone the repository
git clone <repo-url>
cd AI-job-outreach

# Start all services
docker-compose up --build

# API Gateway:    http://localhost:8080
# AI Service:     http://localhost:8000
# Frontend:       http://localhost:3000
```

## Project Status

🟡 **Planning & Documentation Phase** — Architecture and documentation are being finalized before implementation begins.

## License

MIT
