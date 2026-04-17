import json
from collections.abc import Iterator

import httpx

from app.config import LLM_MODEL, LLM_STREAM, LLM_STREAM_MAX_SECONDS
from app.models.requests import GenerateEmailRequest
from app.models.responses import GenerateEmailResponse
from app.prompts.email_prompt import SYSTEM_PROMPT, build_user_prompt
from app.services.llm_response import extract_message_content, parse_email_json
from app.services.llm_stream import (
    accumulate_assistant_text_from_stream,
    chat_completions_url,
    default_headers,
    httpx_timeout_for_json,
    httpx_timeout_for_stream,
    iter_stream_assistant_chunks,
    post_chat_completions_json,
)


def build_generate_email_llm_payload(req: GenerateEmailRequest) -> dict:
    user_prompt = build_user_prompt(
        resume_text=req.resume_text,
        job_description=req.job_description,
        company_name=req.company_name,
        role=req.role,
        tone=req.tone,
    )
    return {
        "model": LLM_MODEL,
        "messages": [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": user_prompt},
        ],
        "temperature": 0.7,
        "max_tokens": 2048,
    }


def parse_email_from_llm_text(raw_text: str) -> GenerateEmailResponse:
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


def _generate_email_non_stream(req: GenerateEmailRequest) -> GenerateEmailResponse:
    """Baseline: full JSON response from LLM (no SSE)."""
    payload = build_generate_email_llm_payload(req)
    headers = default_headers()
    url = chat_completions_url()

    with httpx.Client(timeout=httpx_timeout_for_json()) as client:
        data = post_chat_completions_json(client, url, payload, headers)

    try:
        raw_text = extract_message_content(data)
    except ValueError as e:
        raise ValueError(f"Unexpected LLM response shape: {data!r}") from e

    return parse_email_from_llm_text(raw_text)


def _generate_email_stream_then_parse(req: GenerateEmailRequest) -> GenerateEmailResponse:
    """Stream SSE from LLM; buffer assistant text; parse JSON only after stream completes."""
    payload = build_generate_email_llm_payload(req)
    headers = default_headers()
    url = chat_completions_url()
    max_wall = None if LLM_STREAM_MAX_SECONDS <= 0 else LLM_STREAM_MAX_SECONDS

    with httpx.Client(timeout=httpx_timeout_for_stream()) as client:
        raw_text = accumulate_assistant_text_from_stream(
            client, url, payload, headers, max_wall_seconds=max_wall
        )

    return parse_email_from_llm_text(raw_text)


def generate_email(req: GenerateEmailRequest) -> GenerateEmailResponse:
    """
    Generate email JSON via LLM. If LLM_STREAM is true, uses streaming completion first.
    On stream failure, retries once with non-stream (streaming improves UX; it is not a reliability fix).
    """
    if not LLM_STREAM:
        return _generate_email_non_stream(req)

    try:
        return _generate_email_stream_then_parse(req)
    except Exception as e:
        try:
            return _generate_email_non_stream(req)
        except Exception:
            raise e


def iter_generate_email_sse_events(req: GenerateEmailRequest) -> Iterator[str]:
    """
    Stream NDJSON-like SSE events to the client: delta lines while the LLM streams,
    then a final 'done' event with the parsed GenerateEmailResponse.
    Upstream always uses stream:true (this route exists for progressive UX).
    """
    payload = build_generate_email_llm_payload(req)
    headers = default_headers()
    url = chat_completions_url()
    max_wall = None if LLM_STREAM_MAX_SECONDS <= 0 else LLM_STREAM_MAX_SECONDS
    parts: list[str] = []

    with httpx.Client(timeout=httpx_timeout_for_stream()) as client:
        for chunk in iter_stream_assistant_chunks(
            client, url, payload, headers, max_wall_seconds=max_wall
        ):
            parts.append(chunk)
            yield f"data: {json.dumps({'type': 'delta', 'text': chunk})}\n\n"

    full_text = "".join(parts)
    result = parse_email_from_llm_text(full_text)
    yield f"data: {json.dumps({'type': 'done', 'result': result.model_dump()})}\n\n"
