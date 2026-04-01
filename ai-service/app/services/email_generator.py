import httpx

from app.config import LLM_API_KEY, LLM_BASE_URL, LLM_MODEL
from app.models.requests import GenerateEmailRequest
from app.models.responses import GenerateEmailResponse
from app.prompts.email_prompt import SYSTEM_PROMPT, build_user_prompt
from app.services.llm_response import extract_message_content, parse_email_json


def _chat_completions_url() -> str:
    return f"{LLM_BASE_URL}/chat/completions"


def generate_email(req: GenerateEmailRequest) -> GenerateEmailResponse:
    user_prompt = build_user_prompt(
        resume_text=req.resume_text,
        job_description=req.job_description,
        company_name=req.company_name,
        role=req.role,
        tone=req.tone,
    )

    payload: dict = {
        "model": LLM_MODEL,
        "messages": [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": user_prompt},
        ],
        "temperature": 0.7,
        "max_tokens": 2048,
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
            raise ValueError(
                f"LLM HTTP {response.status_code} from {url}. Body: {body}"
            ) from e

        try:
            data = response.json()
        except Exception as e:
            raise ValueError(
                f"LLM did not return JSON. URL={url} raw={response.text[:500]!r}"
            ) from e

    try:
        raw_text = extract_message_content(data)
    except ValueError as e:
        raise ValueError(f"Unexpected LLM response shape: {data!r}") from e

    try:
        parsed = parse_email_json(raw_text)
    except ValueError:
        raise

    for key in ("subject", "body"):
        if key not in parsed:
            raise ValueError(
                f"JSON missing '{key}'. Keys present: {list(parsed.keys())}"
            )

    return GenerateEmailResponse(
        subject=parsed["subject"],
        body=parsed["body"],
        match_score=max(0.0, min(1.0, float(parsed.get("match_score", 0.5)))),
        key_points=parsed.get("key_points", []),
        reasoning=parsed.get("reasoning", ""),
    )
