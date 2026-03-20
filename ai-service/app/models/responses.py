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
