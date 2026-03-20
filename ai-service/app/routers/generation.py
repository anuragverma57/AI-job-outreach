from fastapi import APIRouter, UploadFile, File, HTTPException

from app.models.requests import GenerateEmailRequest
from app.models.responses import ParseResumeResponse, GenerateEmailResponse
from app.services.resume_parser import extract_text_from_pdf
from app.services.email_generator import generate_email

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
