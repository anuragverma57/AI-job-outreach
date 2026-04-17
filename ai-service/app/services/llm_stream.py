"""
OpenAI-compatible SSE streaming for POST {LLM_BASE_URL}/chat/completions with stream: true.

Streaming improves perceived responsiveness; it does not replace timeouts, retries, or error handling.
See docs/IMPLEMENTATION-GUIDE-LLM-STREAMING.md.
"""

from __future__ import annotations

import json
import time
from collections.abc import Iterator
from typing import Any

import httpx

from app.config import (
    LLM_API_KEY,
    LLM_BASE_URL,
    LLM_HTTP_CONNECT_TIMEOUT,
    LLM_HTTP_POOL_TIMEOUT,
    LLM_HTTP_READ_TIMEOUT,
    LLM_HTTP_WRITE_TIMEOUT,
    LLM_STREAM_MAX_SECONDS,
)


def chat_completions_url() -> str:
    return f"{LLM_BASE_URL}/chat/completions"


def default_headers() -> dict[str, str]:
    h: dict[str, str] = {"Content-Type": "application/json"}
    if LLM_API_KEY:
        h["Authorization"] = f"Bearer {LLM_API_KEY}"
    return h


def httpx_timeout_for_stream() -> httpx.Timeout:
    """Long read timeout for token-by-token SSE; connect/write bounded."""
    return httpx.Timeout(
        connect=LLM_HTTP_CONNECT_TIMEOUT,
        read=LLM_HTTP_READ_TIMEOUT,
        write=LLM_HTTP_WRITE_TIMEOUT,
        pool=LLM_HTTP_POOL_TIMEOUT,
    )


def httpx_timeout_for_json() -> httpx.Timeout:
    """Non-streaming completion: shorter read than stream, still generous for remote LLMs."""
    return httpx.Timeout(
        connect=LLM_HTTP_CONNECT_TIMEOUT,
        read=120.0,
        write=LLM_HTTP_WRITE_TIMEOUT,
        pool=LLM_HTTP_POOL_TIMEOUT,
    )


def _parse_sse_data_line_for_delta(data: str) -> str | None:
    """Return assistant content fragment from one SSE data payload, or None."""
    if data == "[DONE]":
        return None
    try:
        obj: dict[str, Any] = json.loads(data)
    except json.JSONDecodeError:
        return None
    choices = obj.get("choices") or []
    if not choices:
        return None
    first = choices[0]
    if not isinstance(first, dict):
        return None
    delta = first.get("delta") or {}
    if not isinstance(delta, dict):
        return None
    piece = delta.get("content")
    if piece is None:
        return None
    return str(piece)


def iter_stream_assistant_chunks(
    client: httpx.Client,
    url: str,
    payload: dict[str, Any],
    headers: dict[str, str],
    *,
    max_wall_seconds: float | None,
) -> Iterator[str]:
    """
    POST with stream:true; yield assistant content fragments from SSE deltas until [DONE].
    """
    stream_payload = {**payload, "stream": True}
    start = time.monotonic()

    with client.stream("POST", url, json=stream_payload, headers=headers) as response:
        try:
            response.raise_for_status()
        except httpx.HTTPStatusError as e:
            body = response.text[:2000] if response.text else "(empty body)"
            raise ValueError(
                f"LLM HTTP {response.status_code} from {url}. Body: {body}"
            ) from e

        for line in response.iter_lines():
            if max_wall_seconds and max_wall_seconds > 0:
                if (time.monotonic() - start) > max_wall_seconds:
                    raise TimeoutError(
                        f"LLM stream exceeded max wall-clock ({max_wall_seconds}s)"
                    )
            if not line:
                continue
            if line.startswith(":"):
                continue
            if not line.startswith("data: "):
                continue
            data = line[6:].strip()
            if data == "[DONE]":
                break
            piece = _parse_sse_data_line_for_delta(data)
            if piece:
                yield piece


def accumulate_assistant_text_from_stream(
    client: httpx.Client,
    url: str,
    payload: dict[str, Any],
    headers: dict[str, str],
    *,
    max_wall_seconds: float | None,
) -> str:
    """Buffer full assistant text from streaming completion; parse JSON only after stream ends."""
    return "".join(
        iter_stream_assistant_chunks(
            client, url, payload, headers, max_wall_seconds=max_wall_seconds
        )
    )


def post_chat_completions_json(
    client: httpx.Client,
    url: str,
    payload: dict[str, Any],
    headers: dict[str, str],
) -> dict[str, Any]:
    """Non-streaming chat completion; returns full JSON body."""
    body = {**payload, "stream": False}
    response = client.post(url, json=body, headers=headers)
    try:
        response.raise_for_status()
    except httpx.HTTPStatusError as e:
        raw = response.text[:2000] if response.text else "(empty body)"
        raise ValueError(f"LLM HTTP {response.status_code} from {url}. Body: {raw}") from e
    try:
        return response.json()
    except Exception as e:
        raise ValueError(
            f"LLM did not return JSON. URL={url} raw={response.text[:500]!r}"
        ) from e
