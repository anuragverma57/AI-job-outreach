# Feature List

Features are organized by priority. **P0** = must-have for MVP, **P1** = important but can follow, **P2** = nice-to-have / future.

## Implementation status (high level)

| Area | Status |
|------|--------|
| P0 §1–2 Auth & resumes | Shipped (core flows) |
| P0 §3 Applications CRUD | Shipped |
| P0 §4–5 AI email generation & review/edit | Shipped |
| **P0 §6–7 Scheduling & automated send** | **In progress — next milestone (Phase 4)** |
| P0 §8–9 Tracking & dashboard analytics | Not started |

**Next build focus:** §6–7 — Redis-backed scheduling, Go **worker** that sends at `scheduled_at`, SMTP or transactional email API, retries and status updates. See `docs/PHASES.md` Phase 4.

---

## P0 — Core MVP Features

### 1. User Authentication

- User registration with email and password
- Login with JWT token issuance
- Token refresh mechanism
- Protected routes (all features require auth)

### 2. Resume Management

- Upload resume (PDF format)
- Parse and extract text from resume
- Store parsed text for AI analysis
- Support multiple resumes (user can pick which one to use)
- View and delete uploaded resumes

### 3. Job Application Management

- Create a new application with:
  - Company name
  - Role / job title
  - Recruiter email
  - Job description (text)
  - Job posting link
- View all applications (list with filters)
- View single application detail
- Edit application details
- Delete application

### 4. AI-Powered Email Generation

- Analyze resume against job description
- Identify key matching skills, experience, and qualifications
- Generate a personalized cold email with:
  - Subject line
  - Email body
  - Key talking points used
- Return a match/relevance score
- Support regeneration with different tone (formal, friendly, concise)

### 5. Email Review & Editing

- Display generated email for user review
- Inline editing of subject and body
- Regenerate email (full or partial)
- Save edited version as the final draft

### 6. Email Scheduling

- Schedule email for a specific date and time
- Schedule with a relative delay ("send in 2 hours")
- View all scheduled emails
- Cancel a scheduled email before it's sent
- Reschedule a pending email

**Phase 4 implementation notes (backend):**

- Persist `scheduled_at` on `emails` (already in schema) and set `status` to `scheduled` when the user confirms a send time.
- Enqueue work in **Redis** (e.g. sorted set: member = `email_id`, score = send time as Unix ms/seconds) or Redis Streams — pick one and document it.
- Expose authenticated HTTP APIs, for example:
  - `POST /api/emails/:id/schedule` — body: `{ "send_at": "<ISO8601>" }` or `{ "delay_seconds": 7200 }`
  - `DELETE /api/emails/:id/schedule` — cancel (revert to `draft` or `cancelled` per your rules)
  - `PUT /api/emails/:id/schedule` — reschedule
  - `GET /api/emails?status=scheduled` — list due/pending scheduled emails for the current user

### 7. Automated Email Sending

- Background worker consumes scheduled email jobs
- Send email via SMTP or email API (Resend/SendGrid)
- Retry failed sends (up to 3 attempts with backoff)
- Update email status after send (sent / failed)
- Update application status after successful send

**Phase 4 implementation notes (worker + infra):**

- New **Go worker** process (separate from API Gateway): connects to Redis + Postgres, claims due jobs, sends mail, updates `emails` (`sent_at`, `status`, `retry_count`) and related `applications.status` (e.g. to `applied` after successful send).
- Configure SMTP/API via env (reuse patterns from `.env.example`); never commit secrets.
- **Docker Compose:** add `worker` service (and ensure `redis` is used by gateway/worker in dev).

### 8. Application Status Tracking

- Track application through stages:
  - `draft` → `applied` → `replied` → `interview` → `offer` / `rejected` / `ghosted`
- Manual status update by user
- Status history with timestamps
- Filter applications by status

### 9. Dashboard & Analytics

- Total applications count
- Breakdown by status (pie/bar chart)
- Applications over time (line chart)
- Response rate percentage
- Interview conversion rate
- Recent activity feed

---

## P1 — Post-MVP Enhancements

### 10. Email Templates

- Save a generated email as a reusable template
- Apply template to new applications
- Template library with categorization (by role type, industry)

### 11. Follow-up Emails

- Auto-suggest a follow-up if no reply within X days
- Generate follow-up email referencing the original
- Schedule follow-up as a separate email in the queue

### 12. Bulk Operations

- Select multiple applications and update status
- Schedule multiple emails at once
- Export applications to CSV

### 13. Search & Filters

- Full-text search across applications (company, role, JD)
- Filter by date range, status, company
- Sort by date, status, company name

### 14. Email Analytics

- Track if email was opened (if using email API with tracking)
- Link click tracking
- Bounce detection

### 15. Notification System

- In-app notifications for:
  - Email sent successfully
  - Email failed to send
  - Follow-up reminder

---

## P2 — Future / Stretch Features

### 16. LinkedIn Integration

- Detect hiring posts from LinkedIn feed
- Auto-extract job details from LinkedIn post URL
- Store detected opportunities for outreach

### 17. Multi-Channel Outreach

- Support LinkedIn InMail message drafting
- Support Twitter DM drafting
- Channel-aware email generation

### 18. Smart Scheduling

- AI-suggested send times based on timezone and industry
- A/B test different send times
- Optimal time learning from response data

### 19. Resume Tailoring

- AI suggests resume modifications for specific roles
- Highlight which resume sections are most relevant
- Generate role-specific resume summaries

### 20. Team / Multi-User Support

- Multiple user accounts with individual dashboards
- Admin view across all users (for coaching/mentoring scenarios)
- Shared template library

### 21. Webhook Integration

- Incoming webhook for email replies (via email provider)
- Auto-update application status on reply detection
- Slack/Discord notifications

---

## Feature Dependencies

```
Authentication ──► Resume Upload ──► Application CRUD ──► Email Generation
                                                              │
                                                              ▼
                                                        Email Review
                                                              │
                                                              ▼
                                                        Email Scheduling
                                                              │
                                                              ▼
                                                        Email Sending (Worker)
                                                              │
                                                              ▼
                                                        Status Tracking
                                                              │
                                                              ▼
                                                        Dashboard Analytics
```

This dependency chain defines the natural build order. Each feature builds on the previous one.