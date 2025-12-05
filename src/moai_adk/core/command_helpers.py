"""
Command Helper Utilities

Provides helper functions for commands to interact with ContextManager
and perform common operations like context extraction and validation.
"""

import json
import os
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional

import yaml

# Conditional import of ContextManager
try:
    from moai_adk.core.context_manager import (
        ContextManager,
        validate_and_convert_path,
        validate_no_template_vars,
    )

    CONTEXT_MANAGER_AVAILABLE = True
except ImportError:
    CONTEXT_MANAGER_AVAILABLE = False


def extract_project_metadata(project_root: str) -> Dict[str, Any]:
    """
    Extract project metadata from config.yaml.

    Args:
        project_root: Root directory of the project

    Returns:
        Dictionary containing project metadata

    Raises:
        FileNotFoundError: If config.yaml doesn't exist
        yaml.YAMLError: If config.yaml is invalid
    """
    config_path = os.path.join(project_root, ".moai", "config", "config.yaml")

    if not os.path.exists(config_path):
        raise FileNotFoundError(f"Config file not found: {config_path}")

    with open(config_path, "r", encoding="utf-8") as f:
        config = yaml.safe_load(f) or {}

    # Extract key metadata
    metadata = {
        "project_name": config.get("project", {}).get("name", "Unknown"),
        "mode": config.get("project", {}).get("mode", "personal"),
        "github_profile": config.get("github", {}).get("profile_name", ""),
        "language": config.get("language", {}).get("conversation_language", "en"),
        "tech_stack": [],  # To be detected separately
    }

    return metadata


def detect_tech_stack(project_root: str) -> List[str]:
    """
    Detect primary tech stack from project structure.

    Checks for common project indicator files.

    Args:
        project_root: Root directory of the project

    Returns:
        List of detected languages/frameworks
    """
    indicators = {
        "pyproject.toml": "python",
        "package.json": "javascript",
        "go.mod": "go",
        "Cargo.toml": "rust",
        "pom.xml": "java",
        "Gemfile": "ruby",
    }

    tech_stack = []

    for indicator_file, language in indicators.items():
        if os.path.exists(os.path.join(project_root, indicator_file)):
            tech_stack.append(language)

    # Default to python if nothing detected
    if not tech_stack:
        tech_stack.append("python")

    return tech_stack


def build_phase_result(
    phase_name: str,
    status: str,
    outputs: Dict[str, Any],
    files_created: List[str],
    next_phase: Optional[str] = None,
) -> Dict[str, Any]:
    """
    Build standardized phase result dictionary.

    Args:
        phase_name: Name of the phase (e.g., "0-project")
        status: Phase status (completed/error/interrupted)
        outputs: Dictionary of phase outputs
        files_created: List of created files (absolute paths)
        next_phase: Optional next phase name

    Returns:
        Standardized phase result dictionary
    """
    phase_result = {
        "phase": phase_name,
        "timestamp": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "status": status,
        "outputs": outputs,
        "files_created": files_created,
    }

    if next_phase:
        phase_result["next_phase"] = next_phase

    return phase_result


def validate_phase_files(relative_paths: List[str], project_root: str) -> List[str]:
    """
    Validate and convert relative file paths to absolute paths.

    Handles errors gracefully by logging warnings and skipping invalid paths.

    Args:
        relative_paths: List of relative file paths
        project_root: Project root directory

    Returns:
        List of validated absolute paths
    """
    if not CONTEXT_MANAGER_AVAILABLE:
        # Fallback: simple absolute path conversion
        return [os.path.abspath(os.path.join(project_root, p)) for p in relative_paths]

    absolute_paths = []

    for rel_path in relative_paths:
        try:
            abs_path = validate_and_convert_path(rel_path, project_root)
            absolute_paths.append(abs_path)
        except (ValueError, FileNotFoundError) as e:
            # Log warning but continue processing
            print(f"Warning: Could not validate path '{rel_path}': {e}")

    return absolute_paths


def _prepare_phase_data(
    phase_name: str,
    status: str,
    outputs: Dict[str, Any],
    absolute_paths: List[str],
    next_phase: Optional[str],
) -> Dict[str, Any]:
    """
    Prepare phase data for saving.

    Args:
        phase_name: Name of the phase
        status: Phase status
        outputs: Phase outputs
        absolute_paths: List of absolute file paths
        next_phase: Optional next phase

    Returns:
        Phase data dictionary ready for saving
    """
    phase_data = build_phase_result(
        phase_name=phase_name,
        status=status,
        outputs=outputs,
        files_created=absolute_paths,
        next_phase=next_phase,
    )

    # Validate no unsubstituted template variables
    phase_json = json.dumps(phase_data)
    validate_no_template_vars(phase_json)

    return phase_data


def _validate_and_save(context_mgr: Any, phase_data: Dict[str, Any]) -> str:
    """
    Validate and save phase data.

    Args:
        context_mgr: ContextManager instance
        phase_data: Phase data to save

    Returns:
        Path to saved file
    """
    saved_path = context_mgr.save_phase_result(phase_data)
    print(f"✓ Phase context saved: {os.path.basename(saved_path)}")
    return saved_path


def save_command_context(
    phase_name: str,
    project_root: str,
    outputs: Dict[str, Any],
    files_created: List[str],
    next_phase: Optional[str] = None,
    status: str = "completed",
) -> Optional[str]:
    """
    Save command phase context using ContextManager.

    This is a convenience wrapper for commands to save phase results
    with proper error handling.

    Args:
        phase_name: Name of the phase (e.g., "0-project")
        project_root: Project root directory
        outputs: Phase-specific outputs
        files_created: List of relative file paths
        next_phase: Optional next phase recommendation
        status: Phase status (default: "completed")

    Returns:
        Path to saved file, or None if save failed
    """
    if not CONTEXT_MANAGER_AVAILABLE:
        print("Warning: ContextManager not available. Phase context not saved.")
        return None

    try:
        context_mgr = ContextManager(project_root)
        absolute_paths = validate_phase_files(files_created, project_root)
        phase_data = _prepare_phase_data(phase_name, status, outputs, absolute_paths, next_phase)
        return _validate_and_save(context_mgr, phase_data)

    except Exception as e:
        print(f"Warning: Failed to save phase context: {e}")
        print("Command execution continues normally.")
        return None


def load_previous_phase(project_root: str) -> Optional[Dict[str, Any]]:
    """
    Load the most recent phase result.

    Args:
        project_root: Project root directory

    Returns:
        Phase result dictionary, or None if unavailable
    """
    if not CONTEXT_MANAGER_AVAILABLE:
        return None

    try:
        context_mgr = ContextManager(project_root)
        return context_mgr.load_latest_phase()
    except Exception as e:
        print(f"Warning: Could not load previous phase: {e}")
        return None
