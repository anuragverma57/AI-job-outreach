import os

from dotenv import load_dotenv

load_dotenv()


def _env_bool(key: str, default: bool) -> bool:
    raw = os.getenv(key)
    if raw is None or raw == "":
        return default
    return raw.strip().lower() in ("1", "true", "yes", "on")


def _env_float(key: str, default: float) -> float:
    raw = os.getenv(key)
    if raw is None or raw == "":
        return default
    try:
        return float(raw)
    except ValueError:
        return default


# OpenAI-compatible endpoint (e.g. Ollama or proxy). No trailing slash.
# Example: http://192.168.29.231:8000/v1 → POST .../v1/chat/completions
LLM_BASE_URL = os.getenv("LLM_BASE_URL", "http://192.168.29.231:8000/v1").rstrip("/")
# Model name your server expects — must match what works in your LLM curl (e.g. llama3).
LLM_MODEL = os.getenv("LLM_MODEL", "llama3")
# Optional; leave empty for local Ollama
LLM_API_KEY = os.getenv("LLM_API_KEY", "")

# When true, use POST .../chat/completions with "stream": true (SSE); else non-stream JSON.
# On stream failure, callers may fall back to non-stream once (see email_generator / smart_apply).
LLM_STREAM = _env_bool("LLM_STREAM", False)

# httpx timeouts: long read for streaming; bounded connect/write.
LLM_HTTP_CONNECT_TIMEOUT = _env_float("LLM_HTTP_CONNECT_TIMEOUT", 10.0)
LLM_HTTP_READ_TIMEOUT = _env_float("LLM_HTTP_READ_TIMEOUT", 600.0)
LLM_HTTP_WRITE_TIMEOUT = _env_float("LLM_HTTP_WRITE_TIMEOUT", 30.0)
LLM_HTTP_POOL_TIMEOUT = _env_float("LLM_HTTP_POOL_TIMEOUT", 5.0)

# Optional max wall-clock for a single stream (seconds). 0 = disabled (read timeout still applies).
LLM_STREAM_MAX_SECONDS = _env_float("LLM_STREAM_MAX_SECONDS", 900.0)
