# Implementation Guide: Analytics (MVP)

This phase adds **read-only analytics** built from **PostgreSQL** only: **no Gmail**, no new tables for v1. Data comes from **applications** (pipeline status, per user) and **emails** (draft / scheduled / sent / failed), joined through applications owned by the user.

**Depends on:** User-editable pipeline status (`PATCH /api/applications/:id/status`) so the funnel counts **replied**, **interview**, **offer**, etc. are meaningful.

**Goal:** `GET /api/analytics/summary` returns aggregate numbers for the logged-in user; **`/analytics`** renders them; **dashboard** reuses the same numbers (replace hard-coded zeros).

---

## 1. What we show (MVP scope)

| Metric | Source | Notes |
|--------|--------|--------|
| Total applications | `COUNT(*)` on `applications` where `user_id = $1` | |
| Applications by status | `GROUP BY status` on same filter | Same LOV as pipeline status (`draft`, `applied`, `replied`, …) |
| Emails sent | `COUNT(*)` on `emails` where `status = 'sent'` and application belongs to user | Join `emails` → `applications` on `application_id` |
| Emails scheduled | `status = 'scheduled'` | Same join |
| Emails failed | `status = 'failed'` | Same join |
| Emails draft | `status = 'draft'` | Optional card or omit if noisy |

**Derived rates (optional in v1, cheap to add):**

- **Reply rate** = applications with `status IN ('replied','interview','offer')` / max(1, applications that are “past draft”) — product can define denominator as **total applications** or **applied+**; document the choice in code comments.

**Out of scope for this guide**

- Timeline charts (applications over time) — needs `DATE_TRUNC` query; add as **Phase 5b** if desired.
- `GET /api/analytics/timeline` — defer.
- Heavy charting library — MVP can use **stat cards** + optional simple bar list (CSS or existing UI).

---

## 2. API contract (backend owns)

### `GET /api/analytics/summary`

- **Auth:** Same JWT + cookie pattern as `/api/applications`.
- **Response `200`:** JSON shaped like below (field names can be `snake_case` to match existing API style — stay consistent with `application` JSON).

**Example shape:**

```json
{
  "summary": {
    "total_applications": 12,
    "applications_by_status": {
      "draft": 2,
      "applied": 5,
      "replied": 3,
      "interview": 1,
      "offer": 0,
      "rejected": 1,
      "ghosted": 0
    },
    "emails": {
      "sent": 4,
      "scheduled": 1,
      "failed": 0,
      "draft": 2
    },
    "rates": {
      "reply_rate": 0.25,
      "interview_rate": 0.08
    }
  }
}
```

**Rules:**

- **`applications_by_status`:** Always include **all** canonical statuses (zeros where missing) so the frontend does not guess keys. Align with the same LOV as `ApplicationStatus` / `PATCH .../status` validation.
- **`rates`:** Optional; if implemented, define denominator in one place (e.g. `reply_rate = replied_or_better / total_applications` or `/ max(1, total - draft)` — pick one and document).

**Errors:** `401` if unauthenticated; `500` on DB failure with existing error JSON pattern.

---

## 3. Backend responsibilities (`api-gateway`)

| Task | Details |
|------|---------|
| **Repository** | Add something like `AnalyticsRepository` or methods on `ApplicationRepository` + `EmailRepository` — prefer **one** `GetSummary(ctx, userID)` that runs **minimal queries** (1–2 round-trips: e.g. one aggregated query for apps, one for emails with JOIN). |
| **Service** | `AnalyticsService` with `GetSummary(userID)` building the response struct including zero-filled `applications_by_status`. |
| **Handler** | `AnalyticsHandler.Summary` → `GET /api/analytics/summary`. |
| **Router** | Register under protected `api` group: `api.Get("/analytics/summary", ...)`. |
| **Models** | `model.AnalyticsSummary` (or nested structs) for JSON serialization. |

**SQL sketch (adjust to your query style):**

- Applications: `SELECT status, COUNT(*) FROM applications WHERE user_id = $1 GROUP BY status`
- Emails: `SELECT e.status, COUNT(*) FROM emails e INNER JOIN applications a ON a.id = e.application_id WHERE a.user_id = $1 GROUP BY e.status`

Merge into the response map in Go.

**Performance:** For single-user MVP scale, raw SQL is enough; no Redis cache required.

---

## 4. Frontend responsibilities (`frontend`)

| Task | Details |
|------|---------|
| **Types** | `AnalyticsSummary` / nested types in e.g. `src/types/analytics.ts`. |
| **API client** | `getAnalyticsSummary()` → `GET /api/analytics/summary` with `credentials: "include"`. |
| **`/analytics` page** | New route: `src/app/(dashboard)/analytics/page.tsx`. Fetch on mount; show loading/error/empty states; display stat cards + optional simple breakdown (list or bars) for `applications_by_status`. Match existing dashboard layout tokens (`card`, spacing). |
| **Dashboard** | Replace placeholder `"0"` values with the same summary fetch (reuse a small hook `useAnalyticsSummary()` or fetch in both pages — avoid duplicating large layout logic). |
| **Sidebar** | Link to `/analytics` already exists — ensure the page exists (fixes 404). |

**Out of scope**

- Real-time polling — refresh on navigation is enough for MVP.
- Export CSV.

---

## 5. Integration order

1. **Backend** implements `GET /api/analytics/summary` and manual test (`curl` with session cookie or Bearer token — match how you test other protected routes).
2. **Frontend** adds client + `/analytics` page.
3. **Frontend** wires **dashboard** to the same endpoint.
4. Update **`docs/CURRENT-STATE.md`** and **`docs/PHASES.md`** (Phase 5: analytics partial → done for MVP summary).

---

## 6. Files likely touched

**Backend**

- `api-gateway/internal/repository/` — new file or extend existing repos
- `api-gateway/internal/service/analytics.go` (new)
- `api-gateway/internal/handler/analytics.go` (new)
- `api-gateway/internal/model/` — summary structs
- `api-gateway/internal/router/router.go`

**Frontend**

- `frontend/src/types/analytics.ts` (new)
- `frontend/src/lib/api.ts`
- `frontend/src/app/(dashboard)/analytics/page.tsx` (new)
- `frontend/src/app/(dashboard)/dashboard/page.tsx`
- Optional: `frontend/src/hooks/use-analytics-summary.ts` (new)

---

## 7. Testing checklist

- [ ] Logged-in user A only sees counts for their applications/emails.
- [ ] New application increments totals after refresh.
- [ ] Status change on an application updates `applications_by_status` counts.
- [ ] Sent email increments `emails.sent` (and application may be `applied` per worker).
- [ ] `/analytics` loads without 404; sidebar link works.

---

## 8. Frontend vs backend — who owns what

| Area | Backend only | Frontend only | Both (contract) |
|------|----------------|---------------|-----------------|
| **Auth / scoping** | Every query filtered by `user_id` from JWT — never trust client for user id | Sends cookies / `Authorization` like other `/api/*` calls | Same auth as applications |
| **JSON shape** | Defines canonical response; zero-fill `applications_by_status` | Types mirror API; defensive UI if a key is missing | Field names (`snake_case` vs `camelCase`) — match existing API conventions in this repo |
| **Business meaning of rates** | Implements formula in code + comment | Displays numbers; no client-side recomputation of rates unless documented | Agree on denominator for `reply_rate` / `interview_rate` |
| **UI** | — | `/analytics` page, dashboard cards, loading/error/empty | — |
| **Performance** | Efficient SQL (1–2 queries) | No over-fetching; optional shared hook | — |

---

## 9. Assignment summary

| Role | Delivers |
|------|----------|
| **Backend** | `GET /api/analytics/summary`, correct scoping by `user_id`, zero-filled status map |
| **Frontend** | API types + client, `/analytics` UI, dashboard wired to live data |
| **Integration** | Shared JSON contract; both sides use same LOV keys |

This completes the **analytics MVP** slice described in [PHASES.md](PHASES.md) Phase 5 for summary stats; timeline charts remain optional follow-up.

---

## 10. Copy-paste prompts for agents

**Reference file (single source of truth):** this document —  
`docs/IMPLEMENTATION-GUIDE-ANALYTICS.md`

### Prompt — backend agent

```
Implement the Analytics MVP backend per the repository guide (read the full file):

docs/IMPLEMENTATION-GUIDE-ANALYTICS.md

Scope: Add GET /api/analytics/summary behind the same auth middleware as other /api routes. Implement user-scoped SQL (applications + emails joined via applications), return the JSON shape in section 2 including applications_by_status with ALL canonical LOV keys zero-filled. Add handler, service, repository (or minimal new types), models, register in internal/router/router.go. Match existing error JSON patterns. Do not add Gmail, timeline endpoints, or new migrations unless the guide explicitly requires them. Document rate formulas in code comments if you implement rates.
```

### Prompt — frontend agent

```
Implement the Analytics MVP frontend per:

docs/IMPLEMENTATION-GUIDE-ANALYTICS.md

Scope: Add TypeScript types, getAnalyticsSummary() in src/lib/api.ts (credentials: include). Create src/app/(dashboard)/analytics/page.tsx with stat cards and optional status breakdown; loading and error states. Update src/app/(dashboard)/dashboard/page.tsx to replace hard-coded zeros with the same summary (reuse a hook useAnalyticsSummary in src/hooks/ if it reduces duplication). Match existing UI components and layout. Do not add charting libraries unless already in package.json; keep MVP visual weight low.
```

### Prompt — full-stack (single agent)

```
Implement Analytics MVP end-to-end per docs/IMPLEMENTATION-GUIDE-ANALYTICS.md — backend GET /api/analytics/summary first, then frontend /analytics + dashboard. Update docs/CURRENT-STATE.md when done. Out of scope: timeline charts, Gmail, new DB tables.
```
