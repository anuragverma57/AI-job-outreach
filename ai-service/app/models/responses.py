from pydantic import BaseModel
from typing import List


class ParseResumeResponse(BaseModel):
    parsed_text: str


class GenerateEmailResponse(BaseModel):
    subject: str
    body: str
    match_score: float
    key_points: List[str]
    reasoning: str


class SmartApplyExtractResponse(BaseModel):
    company_name: str = ""
    role: str = ""
    recruiter_email: str | None = None
    job_link: str | None = None
    job_description: str = ""
    selected_resume_id: str = ""
    extraction_confidence: str = "medium"
