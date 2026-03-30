import json

import httpx

from app.config import LLM_API_KEY, LLM_BASE_URL, LLM_MODEL
from app.models.requests import GenerateEmailRequest
from app.models.responses import GenerateEmailResponse
from app.prompts.email_prompt import SYSTEM_PROMPT, build_user_prompt


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

    payload = {
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

    with httpx.Client(timeout=120.0) as client:
        response = client.post(_chat_completions_url(), json=payload, headers=headers)
        response.raise_for_status()
        data = response.json()

    try:
        raw_text = data["choices"][0]["message"]["content"].strip()
    except (KeyError, IndexError, TypeError) as e:
        raise ValueError(f"Unexpected LLM response shape: {data}") from e

    if raw_text.startswith("```"):
        lines = raw_text.split("\n")
        lines = [l for l in lines if not l.strip().startswith("```")]
        raw_text = "\n".join(lines).strip()

    parsed = json.loads(raw_text)

    return GenerateEmailResponse(
        subject=parsed["subject"],
        body=parsed["body"],
        match_score=max(0.0, min(1.0, float(parsed.get("match_score", 0.5)))),
        key_points=parsed.get("key_points", []),
        reasoning=parsed.get("reasoning", ""),
    )
