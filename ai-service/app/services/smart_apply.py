import json

import httpx

from app.config import LLM_MODEL, LLM_STREAM, LLM_STREAM_MAX_SECONDS
from app.models.requests import SmartApplyExtractRequest
from app.models.responses import SmartApplyExtractResponse
from app.services.llm_response import extract_message_content, parse_email_json
from app.services.llm_stream import (
    accumulate_assistant_text_from_stream,
    chat_completions_url,
    default_headers,
    httpx_timeout_for_json,
    httpx_timeout_for_stream,
    post_chat_completions_json,
)

SYSTEM_PROMPT_EXTRACT = """
You extract job details and pick the best resume for that job.
Return ONLY valid JSON with this exact schema:
{
  "company_name": "string",
  "role": "string",
  "recruiter_email": "string|null",
  "job_link": "string|null",
  "job_description": "string",
  "selected_resume_id": "string",
  "extraction_confidence": "low|medium|high"
}
Use null for missing optional values. Do not add extra keys.
""".strip()


def _user_content_json(req: SmartApplyExtractRequest) -> str:
    resumes_payload = [
        {"resume_id": r.resume_id, "parsed_text": r.parsed_text[:4000]} for r in req.resumes
    ]
    return json.dumps({"raw_text": req.raw_text, "resumes": resumes_payload})


def _user_content_plain(req: SmartApplyExtractRequest) -> str:
    """
    Same information as _user_content_json but without nested JSON-in-JSON.
    Some OpenAI-compatible proxies return 500 on large or double-encoded payloads.
    """
    parts: list[str] = [
        "Raw job text:\n",
        req.raw_text.strip(),
        "\n\nResumes (pick selected_resume_id from these resume_id values only):\n",
    ]
    for r in req.resumes:
        parts.append(f"\n--- resume_id: {r.resume_id} ---\n")
        parts.append((r.parsed_text or "")[:4000])
        parts.append("\n")
    return "".join(parts)


def _is_upstream_llm_error(err: ValueError) -> bool:
    return str(err).startswith("LLM HTTP")


def _call_llm_for_raw_text(payload: dict) -> str:
    """POST chat/completions; return assistant message string."""
    headers = default_headers()
    url = chat_completions_url()
    max_wall = None if LLM_STREAM_MAX_SECONDS <= 0 else LLM_STREAM_MAX_SECONDS

    if LLM_STREAM:
        try:
            with httpx.Client(timeout=httpx_timeout_for_stream()) as client:
                return accumulate_assistant_text_from_stream(
                    client, url, payload, headers, max_wall_seconds=max_wall
                )
        except Exception:
            with httpx.Client(timeout=httpx_timeout_for_json()) as client:
                data = post_chat_completions_json(client, url, payload, headers)
                return extract_message_content(data)

    with httpx.Client(timeout=httpx_timeout_for_json()) as client:
        data = post_chat_completions_json(client, url, payload, headers)
        return extract_message_content(data)


def _build_extract_payload(user_content: str) -> dict:
    return {
        "model": LLM_MODEL,
        "messages": [
            {"role": "system", "content": SYSTEM_PROMPT_EXTRACT},
            {"role": "user", "content": user_content},
        ],
        "temperature": 0.2,
        # Keep moderate; some local stacks error on very high max_tokens.
        "max_tokens": 2048,
    }


def smart_apply_extract_and_match(req: SmartApplyExtractRequest) -> SmartApplyExtractResponse:
    """
    Call remote OpenAI-compatible /chat/completions. If the upstream returns an HTTP
    error (wrapped as ValueError starting with 'LLM HTTP'), retry once with a plain
    user message (no JSON-in-JSON) — some proxies return 500 on nested payloads.
    """
    payload_json = _build_extract_payload(_user_content_json(req))
    try:
        raw_text = _call_llm_for_raw_text(payload_json)
    except ValueError as e:
        if _is_upstream_llm_error(e):
            payload_plain = _build_extract_payload(_user_content_plain(req))
            raw_text = _call_llm_for_raw_text(payload_plain)
        else:
            raise
    except Exception as e:
        raise ValueError(f"LLM call failed: {e}") from e

    try:
        parsed = parse_email_json(raw_text)
    except ValueError:
        payload_plain = _build_extract_payload(_user_content_plain(req))
        try:
            raw_text = _call_llm_for_raw_text(payload_plain)
            parsed = parse_email_json(raw_text)
        except ValueError as e:
            raise ValueError(
                "Could not parse extraction JSON from LLM output. "
                f"Check LLM_MODEL matches your server (same name as working curl). Detail: {e}"
            ) from e

    confidence = str(parsed.get("extraction_confidence", "medium")).strip().lower()
    if confidence not in {"low", "medium", "high"}:
        confidence = "medium"

    recruiter_email = parsed.get("recruiter_email")
    if recruiter_email == "":
        recruiter_email = None
    job_link = parsed.get("job_link")
    if job_link == "":
        job_link = None

    return SmartApplyExtractResponse(
        company_name=str(parsed.get("company_name", "")).strip(),
        role=str(parsed.get("role", "")).strip(),
        recruiter_email=recruiter_email,
        job_link=job_link,
        job_description=str(parsed.get("job_description", "")).strip(),
        selected_resume_id=str(parsed.get("selected_resume_id", "")).strip(),
        extraction_confidence=confidence,
    )
