"""Config Router

FastAPI router for configuration management endpoints.
Provides REST API for 0-Project configuration dialog.

Endpoints:
- GET /api/config/schema - Get tab schema v3.0.0
- GET /api/config - Get current configuration
- POST /api/config - Save configuration with backup
- POST /api/config/validate - Validate configuration changes
"""

import shutil
from pathlib import Path
from typing import Any, Dict

from fastapi import APIRouter, HTTPException, status

from moai_adk.project.configuration import ConfigurationManager
from moai_adk.project.schema import load_tab_schema
from moai_adk.web.models.config import (
    SaveConfigResponse,
    TabSchemaResponse,
    ValidateConfigResponse,
)

# Configuration directory
CONFIG_DIR = Path(".moai/config")

# Create router
router = APIRouter()

# Initialize config manager
_config_manager: ConfigurationManager | None = None


def get_config_manager() -> ConfigurationManager:
    """Get or create ConfigurationManager instance."""
    global _config_manager
    # Always create new manager to use current CONFIG_DIR
    config_path = CONFIG_DIR / "config.yaml"
    _config_manager = None  # Clear cache to ensure fresh instance
    _config_manager = ConfigurationManager(config_path=config_path)
    return _config_manager


@router.get("/schema", response_model=TabSchemaResponse)
async def get_schema() -> TabSchemaResponse:
    """Get tab schema v3.0.0.

    Returns the complete tab schema with 3 tabs:
    - Tab 1: Essential Setup (Quick Start)
    - Tab 2: Documentation
    - Tab 3: Git Automation

    Each tab contains batches with questions.
    """
    schema = load_tab_schema()
    return TabSchemaResponse(**schema)


@router.get("/config", response_model=Dict[str, Any])
async def get_config() -> Dict[str, Any]:
    """Get current configuration.

    Returns the current project configuration from config.yaml.
    If no config file exists, returns empty dict with smart defaults applied.
    """
    manager = get_config_manager()

    # Load existing config or get empty dict
    config = manager.load()

    # Apply smart defaults if config is empty
    if not config:
        config = {
            "user": {},
            "language": {},
            "project": {},
            "git_strategy": {
                "personal": {},
                "team": {},
            },
            "constitution": {},
            "moai": {},
        }

    return config


@router.post("/config", response_model=SaveConfigResponse)
async def save_config(config_data: Dict[str, Any]) -> SaveConfigResponse:
    """Save configuration with automatic backup.

    Args:
        config_data: Configuration data to save

    Returns:
        SaveConfigResponse with success status and backup path

    Raises:
        HTTPException: If validation fails or save error occurs
    """
    manager = get_config_manager()

    # Ensure config directory exists
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)

    # Create backup if existing config
    config_path = CONFIG_DIR / "config.yaml"
    backup_path = None

    if config_path.exists():
        backup_path = str(config_path) + ".backup"
        try:
            shutil.copy2(config_path, backup_path)
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                detail=f"Failed to create backup: {str(e)}",
            ) from e

    # Build complete config with smart defaults
    try:
        complete_config = manager.build_from_responses(config_data)
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Failed to build configuration: {str(e)}",
        ) from e

    # Validate required fields
    required_fields = [
        "user.name",
        "language.conversation_language",
        "language.agent_prompt_language",
        "project.name",
        "git_strategy.mode",
        "constitution.test_coverage_target",
        "constitution.enforce_tdd",
        "project.documentation_mode",
    ]

    flat_config = _flatten_config(complete_config)
    missing_fields = [f for f in required_fields if f not in flat_config]

    if missing_fields:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
            detail={
                "message": "Missing required fields",
                "missing_fields": missing_fields,
            },
        )

    # Save configuration
    try:
        manager.save(complete_config)
        return SaveConfigResponse(
            success=True,
            message="Configuration saved successfully",
            backup_path=backup_path,
        )
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_CONTENT,
            detail=str(e),
        ) from e
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to save configuration: {str(e)}",
        ) from e


@router.post("/validate", response_model=ValidateConfigResponse)
async def validate_config(config_data: Dict[str, Any]) -> ValidateConfigResponse:
    """Validate configuration without saving.

    Checks if all required fields are present and valid.
    Returns list of missing fields if any.

    Args:
        config_data: Configuration data to validate

    Returns:
        ValidateConfigResponse with validation result
    """
    required_fields = [
        "user.name",
        "language.conversation_language",
        "language.agent_prompt_language",
        "project.name",
        "git_strategy.mode",
        "constitution.test_coverage_target",
        "constitution.enforce_tdd",
        "project.documentation_mode",
    ]

    flat_config = _flatten_config(config_data)
    missing_fields = [f for f in required_fields if f not in flat_config]

    # Additional validations
    errors = []

    # Validate test coverage target range
    if "constitution.test_coverage_target" in flat_config:
        coverage = flat_config["constitution.test_coverage_target"]
        if not isinstance(coverage, int) or not (0 <= coverage <= 100):
            errors.append("constitution.test_coverage_target must be between 0 and 100")

    return ValidateConfigResponse(
        valid=len(missing_fields) == 0 and len(errors) == 0,
        missing_fields=missing_fields,
        errors=errors,
    )


def _flatten_config(config: Dict[str, Any], prefix: str = "") -> Dict[str, Any]:
    """Flatten nested config dict for validation.

    Args:
        config: Configuration dict (potentially nested)
        prefix: Current prefix for nested keys

    Returns:
        Flattened dict with dot-notation keys
    """
    result = {}

    for key, value in config.items():
        new_key = f"{prefix}.{key}" if prefix else key
        if isinstance(value, dict):
            result.update(_flatten_config(value, new_key))
        else:
            result[new_key] = value

    return result
