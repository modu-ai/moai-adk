"""FastAPI Application Factory

Provides the FastAPI application factory for the MoAI Web Backend.
Includes middleware configuration, router registration, and lifespan management.
"""

from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from moai_adk.web.config import WebConfig
from moai_adk.web.database import close_database, init_database
from moai_adk.web.routers import chat, health, providers, sessions, specs


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Application lifespan manager for startup and shutdown events"""
    # Startup: Initialize database
    await init_database()
    yield
    # Shutdown: Close database connection
    await close_database()


def create_app(config: WebConfig | None = None) -> FastAPI:
    """Create and configure the FastAPI application

    Args:
        config: Optional WebConfig instance. Uses defaults if not provided.

    Returns:
        Configured FastAPI application instance
    """
    if config is None:
        config = WebConfig()

    app = FastAPI(
        title="MoAI Web Backend",
        description="Web backend for MoAI-ADK providing chat interface and session management",
        version="0.1.0",
        lifespan=lifespan,
    )

    # Configure CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=config.cors_origins,
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # Register routers
    app.include_router(health.router, prefix="/api", tags=["health"])
    app.include_router(sessions.router, prefix="/api", tags=["sessions"])
    app.include_router(chat.router, prefix="/ws", tags=["chat"])
    app.include_router(providers.router, prefix="/api", tags=["providers"])
    app.include_router(specs.router, prefix="/api", tags=["specs"])

    return app
