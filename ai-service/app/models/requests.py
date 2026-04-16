from pydantic import BaseModel
from typing import Literal


class GenerateEmailRequest(BaseModel):
    resume_text: str
    job_description: str
    company_name: str
    role: str
    recruiter_email: str = ""
    job_link: str = ""
    tone: Literal["formal", "friendly", "concise"] = "formal"


class SmartApplyResumeCandidate(BaseModel):
    resume_id: str
    parsed_text: str


class SmartApplyExtractRequest(BaseModel):
    raw_text: str
    resumes: list[SmartApplyResumeCandidate]
