# Smart Apply Feature Spec

This document is the single focused spec for the **Smart Apply** feature.

## 1) Feature overview

Smart Apply is an additional application-creation mode.  
It does **not** replace manual creation.

User pastes unstructured job text (LinkedIn post, JD, hiring message, etc.).  
System extracts structured fields, selects best resume, generates draft email, and opens a review screen before user schedules/sends.

---

## 2) Product goals

- Reduce friction in application creation.
- Be faster than manual form filling.
- Keep user in control (AI suggests, user confirms).
- Increase number of applications sent per session.

---

## 3) User flow

1. User opens `Applications -> New`.
2. User chooses **Smart Apply** tab/mode.
3. User pastes raw job content and clicks **Extract & Draft**.
4. Backend orchestrates:
   - extract fields from text
   - match best uploaded resume
   - generate personalized email
   - create draft application + draft email
5. User lands on review screen:
   - edit extracted fields
   - edit generated email
   - change selected resume
6. User chooses:
   - save draft
   - schedule
   - send now (if supported in current flow)

---

## 4) Required output fields

From Smart Apply processing:

- `company_name`
- `role`
- `recruiter_email` (nullable)
- `job_link` (nullable)
- `job_description`
- `selected_resume_id`
- `generated_email.subject`
- `generated_email.body`

---

## 5) UX states

- **Idle/input:** large textarea + CTA + “switch to manual” action.
- **Processing:** disable submit, show progress text/spinner.
- **Partial extraction:** show warnings (e.g., recruiter email missing) and let user fill manually.
- **Success/review:** editable form + email editor + resume selector.
- **Error:** clear retry message + manual mode fallback.

---

## 6) Backend responsibilities (API + AI orchestration)

### API Gateway

- Add protected endpoint for Smart Apply creation flow.
- Validate input (`raw_text` required, size limits).
- Fetch user resumes and parsed text from DB.
- Call AI service for extraction and resume matching.
- Call AI email generation using selected resume + extracted JD.
- Persist in DB:
  - create `applications` row (status `draft`)
  - create `emails` row (status `draft`)
- Return response needed by review UI.
- Reuse existing schedule/send endpoints after review.

### AI Service

- Add/extend extraction capability:
  - parse unstructured text into structured job fields.
- Add/extend resume matching:
  - choose best resume from user’s stored resumes.
- Reuse existing generation logic for email draft.
- Return stable JSON contract (no free-form output).

### Data/validation

- Must handle incomplete text gracefully.
- `recruiter_email` and `job_link` can be null.
- Do not fail whole flow if optional fields are missing.

---

## 7) Frontend responsibilities

- Add Smart Apply mode in new-application experience.
- Build raw text input UI and submit action.
- Show processing state and prevent double submits.
- Render review screen populated from API response.
- Allow user edits for all extracted fields.
- Allow changing selected resume before confirm.
- Allow editing generated email before schedule/send.
- Show validation message if required fields are missing before final confirm.
- Keep manual mode unchanged and always accessible.

---

## 8) API design (MVP)

## Endpoint

`POST /api/applications/smart-apply`

### Request

```json
{
  "raw_text": "full unstructured job content pasted by user"
}
```

### Response (example)

```json
{
  "application": {
    "id": "uuid",
    "company_name": "Acme",
    "role": "Backend Engineer",
    "recruiter_email": null,
    "job_link": null,
    "job_description": "....",
    "resume_id": "uuid",
    "status": "draft"
  },
  "email": {
    "id": "uuid",
    "subject": "...",
    "body": "...",
    "status": "draft"
  },
  "meta": {
    "extraction_confidence": "medium"
  }
}
```

### Error behavior

- `400` invalid/missing input
- `401` unauthenticated
- `422` extraction produced insufficient required fields (optional if you prefer partial success)
- `500` AI or server failure

---

## 9) Edge cases and fallbacks

- No recruiter email found -> allow manual entry in review.
- No job link found -> keep null, continue.
- Company/role weak extraction -> highlight field for user confirmation.
- User has no resumes -> block Smart Apply with clear CTA: “Upload a resume first.”
- AI timeout/failure -> retry action + manual mode fallback.

---

## 10) MVP scope vs future

### MVP (now)

- Smart Apply endpoint + UI flow.
- Extraction + resume auto-selection + email generation.
- Review + edit + confirm flow.
- Integration with existing scheduling/sending.

### Future (later)

- Confidence per field shown in UI.
- Preview-only mode (no DB write until confirm).
- Duplicate detection.
- Inbox integration (Gmail/IMAP) and auto-reply detection.
- Browser extension for one-click import from job sites.

---

## 11) Agent assignment prompts

### Backend agent prompt

Implement Smart Apply backend using `docs/FEATURE.md` as source of truth.  
Scope: add `POST /api/applications/smart-apply` in API gateway, orchestrate extraction + resume match + email generation via AI service, create draft application/email in DB, and return review payload. Handle missing optional fields gracefully, enforce auth/validation, and do not break existing manual creation and scheduling flows.

### Frontend agent prompt

Implement Smart Apply frontend using `docs/FEATURE.md` as source of truth.  
Scope: add Smart Apply mode in new application flow with paste input, processing state, review/edit screen (fields + resume selector + email editor), and handoff to existing schedule/send actions. Keep manual form unchanged and available.
