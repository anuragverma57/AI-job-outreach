import logging

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from app.routers import health, generation

logger = logging.getLogger(__name__)

app = FastAPI(title="AI Service", version="1.0.0")


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    """
    Log validation detail so 422s are debuggable (e.g. missing `resumes` on
    POST /ai/smart-apply/extract-match — this endpoint is NOT /v1/chat/completions).
    """
    logger.warning("request validation failed: %s path=%s", exc.errors(), request.url.path)
    return JSONResponse(status_code=422, content={"detail": exc.errors()})


app.include_router(health.router)
app.include_router(generation.router)
