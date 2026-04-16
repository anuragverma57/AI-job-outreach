import json

import httpx

from app.config import LLM_API_KEY, LLM_BASE_URL, LLM_MODEL
from app.models.requests import SmartApplyExtractRequest
from app.models.responses import SmartApplyExtractResponse
from app.services.llm_response import extract_message_content, parse_email_json


def _chat_completions_url() -> str:
    return f"{LLM_BASE_URL}/chat/completions"


def smart_apply_extract_and_match(req: SmartApplyExtractRequest) -> SmartApplyExtractResponse:
    resumes_payload = [
        {"resume_id": r.resume_id, "parsed_text": r.parsed_text[:4000]} for r in req.resumes
    ]
    user_payload = {"raw_text": req.raw_text, "resumes": resumes_payload}

    system_prompt = """
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

    payload = {
        "model": LLM_MODEL,
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": json.dumps(user_payload)},
        ],
        "temperature": 0.2,
        "max_tokens": 1500,
    }

    headers = {"Content-Type": "application/json"}
    if LLM_API_KEY:
        headers["Authorization"] = f"Bearer {LLM_API_KEY}"

    url = _chat_completions_url()
    with httpx.Client(timeout=120.0) as client:
        response = client.post(url, json=payload, headers=headers)
        try:
            response.raise_for_status()
        except httpx.HTTPStatusError as e:
            body = response.text[:2000] if response.text else "(empty body)"
            raise ValueError(f"LLM HTTP {response.status_code} from {url}. Body: {body}") from e
        data = response.json()

    raw_text = extract_message_content(data)
    parsed = parse_email_json(raw_text)

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
