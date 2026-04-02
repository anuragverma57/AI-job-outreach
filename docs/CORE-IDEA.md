# Core Idea

## The Problem

Job hunting is a tedious, repetitive process. A typical workflow looks like this:

1. Find a job posting (LinkedIn, Naukri, company career page, referral, etc.)
2. Read the job description carefully
3. Figure out how your experience maps to the role
4. Write a cold email to the recruiter explaining why you're a fit
5. Send the email (ideally at a good time, not 2 AM)
6. Manually track whether you got a reply, interview, or rejection
7. Repeat this 50–200 times

Most of this is mechanical. The creative part — figuring out the match and writing a compelling email — can be assisted by AI. The logistical parts — scheduling, sending, tracking — should be fully automated.

## The Solution

Build a platform that automates the outreach pipeline:

```
Job Opportunity Input
        │
        ▼
┌─────────────────────┐
│  Resume + JD Analysis│  ← AI identifies the match
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Cold Email Generation│  ← AI drafts the email
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Review & Edit       │  ← User stays in control
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Schedule & Send     │  ← System handles timing
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  Track & Analyze     │  ← Dashboard shows pipeline
└─────────────────────┘
```

## Goals

### Primary Goals

1. **Reduce manual effort** — A user should go from "found a job posting" to "email scheduled" in under 2 minutes
2. **Maintain quality** — AI-generated emails should be personalized, not generic templates
3. **Stay organized** — All applications tracked in one place with clear status visibility
4. **Learn by building** — The system should reflect real-world backend architecture patterns: microservices, queues, background workers, API design, containerization

### Non-Goals (for now)

- Automated scraping of job boards (future enhancement)
- Multi-user SaaS with billing
- Mobile app
- Integration with ATS systems

## Who Is This For

A single user (the job seeker) who wants to:
- Apply to many roles efficiently
- Send personalized (not spammy) outreach
- Keep track of their entire pipeline
- Have a portfolio project demonstrating serious backend engineering

## Core Principles

1. **Automation over manual work** — If the system can do it, the user shouldn't have to
2. **User stays in control** — AI generates, user approves. No email goes out without review
3. **Real architecture** — No shortcuts. Proper service boundaries, proper queue, proper database schema
4. **Simplicity where possible** — Microservices where they make sense, not for the sake of it
5. **Extensibility** — Designed so future features (LinkedIn scraping, templates, A/B testing) can plug in cleanly

For what is already built in the repo versus still planned, see [CURRENT-STATE.md](CURRENT-STATE.md).
