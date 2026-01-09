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
from moai_adk.web.routers import (
    chat,
    cost,
    health,
    model_router,
    providers,
    sessions,
    specs,
    terminal,
)

# Import config_router with alias to avoid conflict with WebConfig parameter
from moai_adk.web.routers.config import router as config_router


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Application lifespan manager for startup and shutdown events"""
    # Startup: Initialize database
    await init_database()
    yield
    # Shutdown: Close database connection and terminal sessions
    from moai_adk.web.routers.terminal import get_terminal_manager

    await get_terminal_manager().close_all()
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
    app.include_router(terminal.router, prefix="/api", tags=["terminals"])
    app.include_router(terminal.router, prefix="/ws", tags=["terminals"])
    app.include_router(model_router.router, prefix="/api", tags=["model-router"])
    app.include_router(cost.router, prefix="/api", tags=["cost"])
    app.include_router(config_router, prefix="/api", tags=["config"])

    # Import workflow router here to avoid circular import and lint removal
    from moai_adk.web.routers import workflow

    app.include_router(workflow.router, prefix="/api/workflow", tags=["workflow"])

    return app


# Create app instance for uvicorn
app = create_app()
