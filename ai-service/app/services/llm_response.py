"""Helpers to parse OpenAI-compatible chat responses and JSON email payloads."""

from __future__ import annotations

import json
from typing import Any


def extract_message_content(data: dict[str, Any]) -> str:
    """Get assistant text from a /v1/chat/completions JSON body."""
    choices = data.get("choices") or []
    if not choices:
        raise ValueError("LLM response has no 'choices'")

    first = choices[0]
    if not isinstance(first, dict):
        raise ValueError("LLM response choices[0] is not an object")

    msg = first.get("message")
    if isinstance(msg, dict):
        content = msg.get("content")
        if content is not None:
            return str(content).strip()

    # Rare alternate shapes
    if "text" in first:
        return str(first["text"]).strip()

    raise ValueError(f"Cannot find message content in choices[0]: {first!r}")


def parse_email_json(raw_text: str) -> dict[str, Any]:
    """
    Parse JSON object from model output. Handles markdown fences and extra prose.
    """
    text = (raw_text or "").strip()
    if not text:
        raise ValueError(
            "Model returned empty content. Check LLM_BASE_URL, LLM_MODEL, and that "
            "the server is reachable from this machine."
        )

    if text.startswith("```"):
        lines = text.split("\n")
        lines = [ln for ln in lines if not ln.strip().startswith("```")]
        text = "\n".join(lines).strip()

    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass

    # Extract first JSON object substring
    start = text.find("{")
    end = text.rfind("}")
    if start != -1 and end != -1 and end > start:
        try:
            return json.loads(text[start : end + 1])
        except json.JSONDecodeError as e:
            raise ValueError(
                f"Model did not return valid JSON. First 500 chars:\n{text[:500]}"
            ) from e

    raise ValueError(
        f"Model did not return valid JSON. First 500 chars:\n{text[:500]}"
    )
