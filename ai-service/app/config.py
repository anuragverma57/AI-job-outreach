import os

from dotenv import load_dotenv

load_dotenv()

# OpenAI-compatible endpoint (e.g. Ollama or proxy). No trailing slash.
# Example: http://192.168.29.231:8000/v1 → POST .../v1/chat/completions
LLM_BASE_URL = os.getenv("LLM_BASE_URL", "http://192.168.29.231:8000/v1").rstrip("/")
# Model name your server expects (e.g. llama3.1, mistral, etc.)
LLM_MODEL = os.getenv("LLM_MODEL", "llama3.1")
# Optional; leave empty for local Ollama
LLM_API_KEY = os.getenv("LLM_API_KEY", "")
