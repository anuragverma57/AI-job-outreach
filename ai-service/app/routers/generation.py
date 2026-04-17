import json

from fastapi import APIRouter, UploadFile, File, HTTPException
from fastapi.responses import StreamingResponse

from app.models.requests import GenerateEmailRequest, SmartApplyExtractRequest
from app.models.responses import ParseResumeResponse, GenerateEmailResponse, SmartApplyExtractResponse
from app.services.resume_parser import extract_text_from_pdf
from app.services.email_generator import generate_email, iter_generate_email_sse_events
from app.services.smart_apply import smart_apply_extract_and_match

router = APIRouter(prefix="/ai")


@router.post("/parse-resume", response_model=ParseResumeResponse)
async def parse_resume(file: UploadFile = File(...)):
    if not file.filename or not file.filename.lower().endswith(".pdf"):
        raise HTTPException(status_code=400, detail="Only PDF files are accepted")

    file_bytes = await file.read()

    if len(file_bytes) == 0:
        raise HTTPException(status_code=400, detail="Empty file")

    try:
        parsed_text = extract_text_from_pdf(file_bytes)
    except ValueError as e:
        raise HTTPException(status_code=422, detail=str(e))
    except Exception:
        raise HTTPException(status_code=500, detail="Failed to parse PDF")

    return ParseResumeResponse(parsed_text=parsed_text)


@router.post("/generate-email", response_model=GenerateEmailResponse)
async def generate_email_endpoint(req: GenerateEmailRequest):
    if not req.resume_text.strip():
        raise HTTPException(status_code=400, detail="resume_text is required")
    if not req.job_description.strip():
        raise HTTPException(status_code=400, detail="job_description is required")

    try:
        result = generate_email(req)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Email generation failed: {str(e)}")

    return result


@router.post("/generate-email/stream")
async def generate_email_stream_endpoint(req: GenerateEmailRequest):
    """
    SSE stream: forwards OpenAI-style token deltas as JSON lines, then a final `done` event
    with the parsed GenerateEmailResponse. Upstream LLM always uses stream:true for this route.
    """
    if not req.resume_text.strip():
        raise HTTPException(status_code=400, detail="resume_text is required")
    if not req.job_description.strip():
        raise HTTPException(status_code=400, detail="job_description is required")

    def event_stream():
        try:
            yield from iter_generate_email_sse_events(req)
        except Exception as e:
            yield f"data: {json.dumps({'type': 'error', 'detail': str(e)})}\n\n"

    return StreamingResponse(
        event_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",
        },
    )


@router.post("/smart-apply/extract-match", response_model=SmartApplyExtractResponse)
async def smart_apply_extract_match_endpoint(req: SmartApplyExtractRequest):
    if not req.raw_text.strip():
        raise HTTPException(status_code=400, detail="raw_text is required")
    if not req.resumes:
        raise HTTPException(status_code=400, detail="at least one resume is required")

    try:
        return smart_apply_extract_and_match(req)
    except ValueError as e:
        msg = str(e)
        # Upstream LLM returned 4xx/5xx — not a client body validation issue.
        if msg.startswith("LLM HTTP"):
            raise HTTPException(status_code=502, detail=msg) from e
        raise HTTPException(status_code=422, detail=msg) from e
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Smart apply extraction failed: {str(e)}")
