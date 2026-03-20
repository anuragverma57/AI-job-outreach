from fastapi import FastAPI

from app.routers import health, generation

app = FastAPI(title="AI Service", version="1.0.0")

app.include_router(health.router)
app.include_router(generation.router)
