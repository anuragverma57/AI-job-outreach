# Implementation Guide: Pipeline status (LOV) — user-marked opportunities

This feature lets users **record where each job opportunity sits in the funnel** using a **fixed list of values (LOV)** stored in our database. There is **no inbox integration** in this phase: the user explicitly picks a status in the UI.

**Goal:** One canonical set of statuses, validated on the server, visible on list/detail, ready for future analytics (counts by status) without redesign.

---

## 1. Product behavior

1. Every **application** has a `status` field (already on `applications.status` in Postgres).
2. The **background worker** may set status to **`applied`** when an email is successfully sent (existing behavior — do not break).
3. The user can **change status** anytime via the UI (dropdown or select). Changes persist to the DB and refresh badges everywhere that reads `application.status`.
4. Invalid values are **rejected by the API** with a clear error so frontend and backend stay in sync.

---

## 2. Canonical list of values (LOV)

**Single source of truth:** define the same set in **Go** (validation) and **TypeScript** (UI labels). Keep the **string values** identical (snake-case keys below).

| Value (`status`) | User-facing label (example) |
|------------------|-----------------------------|
| `draft`          | Draft                       |
| `applied`        | Applied                     |
| `replied`        | Replied                     |
| `interview`      | Interview                   |
| `offer`          | Offer                       |
| `rejected`       | Rejected                    |
| `ghosted`        | Ghosted                     |

**Notes:**

- **`draft`**: new application default (DB default already `draft`).
- **`applied`**: set automatically by worker after successful send; user may also select it manually if needed (e.g. applied outside the product).
- **No migration required** for this feature if the column is already `VARCHAR` — only API + UI work.

**Transition rules (MVP recommendation):**

- **Option A (simplest):** Any LOV value is allowed from any current value. Easiest to ship; fewer edge-case bugs.
- **Option B (stricter):** Enforce a small state machine later (e.g. cannot go from `offer` to `draft`). Defer unless product asks for it.

This guide assumes **Option A** unless you explicitly choose B.

---

## 3. API contract (backend owns the contract)

### 3.1 Update status

**Endpoint:** `PATCH /api/applications/:id/status`  
**Auth:** Same as other `/api/*` routes (JWT + cookie refresh pattern already used).

**Request body:**

```json
{ "status": "replied" }
```

**Success:** `200 OK`

```json
{ "application": { /* full application object including updated status and updated_at */ } }
```

**Errors:**

| Situation | Status | Body |
|-----------|--------|------|
| Invalid JSON / missing `status` | 400 | `{ "error": "..." }` |
| `status` not in LOV | 400 | `{ "error": "invalid status" }` (or enumerate allowed) |
| Not found or not owned | 404 / 403 | Match existing application handler patterns |

**Why PATCH and not PUT on the whole resource?**

- Keeps **status** separate from `PUT /api/applications/:id` (job fields only), so partial updates stay clear and the frontend can call a tiny method without sending the whole form.

### 3.2 Optional: list allowed values (nice for frontend)

**Endpoint:** `GET /api/applications/meta/statuses` (or `GET /api/meta/application-statuses`)  
**Auth:** Protected (same as applications).

**Response:**

```json
{
  "statuses": [
    { "value": "draft", "label": "Draft" },
    { "value": "applied", "label": "Applied" }
  ]
}
```

**Who needs this?**

- If labels ever change or you add a status in one place, the UI can stay dumb. **MVP shortcut:** skip this endpoint and **duplicate the LOV in TypeScript** to match Go; add the meta endpoint when you want one source of truth for labels.

---

## 4. Responsibilities

### 4.1 Backend (`api-gateway`)

| Responsibility | Details |
|----------------|---------|
| **Validate** `status` against the fixed LOV | Centralize in one place (e.g. package `model` or `service` helper). |
| **Expose** `PATCH /api/applications/:id/status` | New handler method; register in `internal/router/router.go` under the authenticated `/api/applications` group. |
| **Authorize** | Reuse the same ownership checks as `GetByID` / `Update` (user must own the application). |
| **Persist** | Call repository to update `status` and return the full row. **Note:** `ApplicationRepository.UpdateStatus` already exists — prefer extending it to **return** the updated application (or call `FindByID` after update) so the handler can respond with a full `application` object. |
| **Do not** break worker | Worker continues to call `UpdateStatus` / existing path for `applied` after send. Ensure no regression. |
| **Errors** | Consistent JSON `{ "error": "..." }` like other handlers. |

**Out of scope for this feature (backend):**

- Status **history** table / audit log (future).
- Analytics aggregation endpoints (separate small task: `GET /api/analytics/summary` querying `applications` + `emails`).
- Gmail or IMAP.

### 4.2 Frontend (`frontend`)

| Responsibility | Details |
|----------------|---------|
| **Types** | `ApplicationStatus` in `src/types/application.ts` must match backend LOV **exactly**. |
| **API client** | Add e.g. `updateApplicationStatus(id, status)` calling `PATCH /api/applications/${id}/status` with JSON body; use existing `credentials: "include"` pattern. |
| **Application detail page** | Add a **status control** (select or dropdown) next to the existing `StatusBadge`: on change, call API, then update local state or refetch. Show loading/error (toast or inline) per existing UI patterns. |
| **Application list** | Already shows `StatusBadge`; ensure it updates after returning from detail (refetch on focus or after navigation if needed). |
| **Optional** | Filter list by status — **defer** to a follow-up if timeboxed; not required for “mark opportunity” MVP. |

**Out of scope for this feature (frontend):**

- Real dashboard charts (separate task once summary API exists).
- Fixing `/analytics` route unless you bundle it — recommend **either** implement a stub page **or** temporarily hide the nav item until analytics ships.

### 4.3 Integration (who does what to ship the end product)

| Step | Owner | Action |
|------|--------|--------|
| 1 | **Backend** | Implement PATCH + validation + tests (manual `curl` minimum). |
| 2 | **Frontend** | Implement API method + detail UI; test against running API. |
| 3 | **Either / pair** | Agree on **exact** error strings or status codes for invalid `status` so the UI can show a friendly message. |
| 4 | **Either** | Update `docs/CURRENT-STATE.md` and `docs/PHASES.md` (Phase 5) to note “user status updates shipped.” |
| 5 | **QA** | Flow: create application → generate email → schedule/send (optional) → set status to **Replied** / **Interview** → refresh list → badge matches. |

**Contract first:** Backend merges the route and documents the JSON shape in this file (or OpenAPI later). Frontend implements against **this guide**; if the API changes, backend updates the guide.

---

## 5. Files likely touched (reference)

**Backend**

- `api-gateway/internal/router/router.go` — register route.
- `api-gateway/internal/handler/application.go` — new handler (or dedicated small handler).
- `api-gateway/internal/service/application.go` — `UpdateStatus` with validation.
- `api-gateway/internal/model/application.go` — request DTO e.g. `UpdateApplicationStatusRequest`.
- `api-gateway/internal/repository/application.go` — ensure update + read path returns fresh data (adjust `UpdateStatus` if it currently returns only `error`).

**Frontend**

- `frontend/src/lib/api.ts`
- `frontend/src/app/(dashboard)/applications/[id]/page.tsx`
- Optionally `frontend/src/components/applications/` for a small `StatusSelect` component.

---

## 6. Testing checklist

- [ ] Valid status updates return `200` and persisted value matches.
- [ ] Invalid status returns `400`.
- [ ] Wrong user / wrong id returns `403` / `404` consistent with other application routes.
- [ ] Worker still sets `applied` after send; user can still change status afterward.
- [ ] UI shows new status on detail and list without a full page reload (or acceptable refetch).

---

## 7. Follow-ups (not part of this guide’s MVP)

1. **`GET /api/analytics/summary`** — counts by `applications.status`, emails sent/failed/scheduled; power dashboard cards.
2. **Status history** — `application_status_events` table + `GET /api/applications/:id/history`.
3. **List filters** — query param on `GET /api/applications?status=replied`.

---

## Assignment confirmation

**Yes:** this guide describes implementing **pipeline status in our own DB** with a **fixed LOV**, **user-driven updates** in the UI, **backend validation**, and clear **frontend vs backend vs integration** ownership. No email provider inbox features in this slice.
