import json

from google import genai

from app.config import GEMINI_API_KEY, GEMINI_MODEL
from app.models.requests import GenerateEmailRequest
from app.models.responses import GenerateEmailResponse
from app.prompts.email_prompt import SYSTEM_PROMPT, build_user_prompt


client = genai.Client(api_key=GEMINI_API_KEY)


def generate_email(req: GenerateEmailRequest) -> GenerateEmailResponse:
    user_prompt = build_user_prompt(
        resume_text=req.resume_text,
        job_description=req.job_description,
        company_name=req.company_name,
        role=req.role,
        tone=req.tone,
    )

    response = client.models.generate_content(
        model=GEMINI_MODEL,
        contents=[
            {"role": "user", "parts": [{"text": SYSTEM_PROMPT + "\n\n" + user_prompt}]}
        ],
        config={
            "temperature": 0.7,
            "max_output_tokens": 2048,
        },
    )

    raw_text = response.text.strip()

    # Strip markdown code fences if present
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
