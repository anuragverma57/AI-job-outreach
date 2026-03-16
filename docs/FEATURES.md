# Feature List

Features are organized by priority. **P0** = must-have for MVP, **P1** = important but can follow, **P2** = nice-to-have / future.

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

### 7. Automated Email Sending
- Background worker consumes scheduled email jobs
- Send email via SMTP or email API (Resend/SendGrid)
- Retry failed sends (up to 3 attempts with backoff)
- Update email status after send (sent / failed)
- Update application status after successful send

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
