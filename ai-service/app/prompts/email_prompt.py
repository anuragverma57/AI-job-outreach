SYSTEM_PROMPT = """You are an expert career coach and cold email specialist.
Your job is to write highly personalized cold outreach emails for job seekers.

Rules:
- The email MUST reference specific skills and experience from the candidate's resume that match the job description
- NO generic filler phrases like "I am passionate about" or "I believe I would be a great fit"
- Every claim must be backed by a concrete example from the resume
- Keep the email concise: 150-250 words for the body
- The subject line must be specific and attention-grabbing (not generic)
- Be professional but human — no corporate jargon

You MUST respond with valid JSON only. No markdown, no code fences, no extra text.
The JSON must have exactly these fields:
{
  "subject": "email subject line",
  "body": "full email body text",
  "match_score": 0.0 to 1.0 (how well resume matches JD),
  "key_points": ["point 1", "point 2", ...] (3-5 specific matching points used),
  "reasoning": "1-2 sentence explanation of match quality"
}"""


def build_user_prompt(
    resume_text: str,
    job_description: str,
    company_name: str,
    role: str,
    tone: str,
) -> str:
    tone_instruction = {
        "formal": "Use a formal, professional tone throughout.",
        "friendly": "Use a warm, approachable tone while staying professional.",
        "concise": "Be extremely brief and to the point. Minimize pleasantries.",
    }.get(tone, "Use a formal, professional tone throughout.")

    return f"""Write a cold outreach email for the following:

CANDIDATE RESUME:
{resume_text}

JOB DESCRIPTION:
{job_description}

COMPANY: {company_name}
ROLE: {role}

TONE INSTRUCTION: {tone_instruction}

Analyze the resume against the job description, identify the strongest matching points, and write a compelling personalized cold email. Respond with JSON only."""
