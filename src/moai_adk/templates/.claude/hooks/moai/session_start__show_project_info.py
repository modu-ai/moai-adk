#!/usr/bin/env python3
# SessionStart Hook: Enhanced Project Information
"""SessionStart Hook: Enhanced Project Information

Claude Code Event: SessionStart
Purpose: Display enhanced project status with Git info, test status, and SPEC progress
Execution: Triggered automatically when Claude Code session begins

Enhanced Features:
- Optimized timeout handling with unified manager
- Efficient Git operations with connection pooling and caching
- Enhanced error handling with graceful degradation
- Resource monitoring and cleanup
- Risk assessment with performance metrics
"""

from __future__ import annotations

import json
import logging
import sys
from datetime import datetime
from pathlib import Path
from typing import Any

# =============================================================================
# Windows UTF-8 Encoding Fix (Issue #249)
# Ensures emoji characters are properly displayed on Windows terminals
# =============================================================================
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except (AttributeError, OSError):
        # Python < 3.7 or reconfigure not available
        pass

# =============================================================================
# Constants - Risk Assessment Thresholds
# =============================================================================
# These thresholds determine project risk level based on various factors
RISK_SCORE_HIGH = 20  # Score >= this is HIGH risk
RISK_SCORE_MEDIUM = 10  # Score >= this (and < HIGH) is MEDIUM risk

# Git changes thresholds for risk calculation
GIT_CHANGES_HIGH_THRESHOLD = 20  # Adds 10 to risk score
GIT_CHANGES_MEDIUM_THRESHOLD = 10  # Adds 5 to risk score

# SPEC progress thresholds
SPEC_PROGRESS_LOW = 50  # Below this adds 15 to risk score
SPEC_PROGRESS_MEDIUM = 80  # Below this adds 8 to risk score

# Risk score contributions
RISK_GIT_CHANGES_HIGH = 10
RISK_GIT_CHANGES_MEDIUM = 5
RISK_SPEC_LOW = 15
RISK_SPEC_MEDIUM = 8
RISK_TEST_FAILED = 12
RISK_COVERAGE_UNKNOWN = 5

# Setup message suppression period (days)
SETUP_MESSAGE_RESCAN_DAYS = 7

# =============================================================================
# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

# Import path utils for project root resolution
from lib.file_utils import check_file_size  # noqa: E402
from lib.path_utils import find_project_root  # noqa: E402

# Import unified timeout manager and Git operations manager
try:
    from lib.git_operations_manager import GitOperationType, get_git_manager
    from lib.timeout import TimeoutError as PlatformTimeoutError
    from lib.unified_timeout_manager import (
        HookTimeoutConfig,
        HookTimeoutError,
        TimeoutPolicy,
        get_timeout_manager,
        hook_timeout_context,
    )
except ImportError:
    # Fallback implementations if new modules not available

    def get_timeout_manager():
        return None

    def hook_timeout_context(hook_name, config=None):
        import contextlib

        @contextlib.contextmanager
        def dummy_context():
            yield

        return dummy_context()

    class HookTimeoutConfig:  # type: ignore[no-redef]
        def __init__(self, **kwargs):
            pass

    class TimeoutPolicy:  # type: ignore[no-redef]
        FAST = "fast"
        NORMAL = "normal"
        SLOW = "slow"

    class HookTimeoutError(Exception):  # type: ignore[no-redef]
        pass

    def get_git_manager():
        return None

    class GitOperationType:  # type: ignore[no-redef]
        BRANCH = "branch"
        LOG = "log"
        STATUS = "status"

    class PlatformTimeoutError(Exception):  # type: ignore[no-redef]
        pass


# Import config cache
try:
    from core.config_cache import get_cached_config, get_cached_spec_progress
except ImportError:
    # Fallback to direct functions if cache not available
    # Try PyYAML first, then use simple parser
    try:
        import yaml as yaml_fallback

        HAS_YAML_FALLBACK = True
    except ImportError:
        HAS_YAML_FALLBACK = False

    def _simple_yaml_parse(content: str) -> dict:
        """Simple YAML parser for basic key-value configs without PyYAML dependency.

        Handles:
        - Top-level keys with nested values
        - String values (quoted or unquoted, including empty strings)
        - Boolean values (true/false)
        - Numeric values
        - Comments (lines starting with # or inline after values)

        Does NOT handle:
        - Lists
        - Complex nested structures beyond 2 levels
        - Multi-line strings
        """
        result = {}
        current_section = None
        lines = content.split("\n")

        for line in lines:
            # Skip empty lines and comments
            stripped = line.strip()
            if not stripped or stripped.startswith("#"):
                continue

            # Count leading spaces for indentation
            indent = len(line) - len(line.lstrip())

            # Check if this is a key-value pair
            if ":" in stripped:
                key_part, _, value_part = stripped.partition(":")
                key = key_part.strip()
                value = value_part.strip()

                # Track if value was explicitly quoted (including empty strings)
                was_quoted = False

                # Handle quoted strings first - extract value within quotes
                if value.startswith('"'):
                    # Find the closing quote
                    close_quote = value.find('"', 1)
                    if close_quote > 0:
                        value = value[1:close_quote]
                        was_quoted = True
                    elif value == '""':
                        # Handle explicit empty string ""
                        value = ""
                        was_quoted = True
                elif value.startswith("'"):
                    # Find the closing quote
                    close_quote = value.find("'", 1)
                    if close_quote > 0:
                        value = value[1:close_quote]
                        was_quoted = True
                    elif value == "''":
                        # Handle explicit empty string ''
                        value = ""
                        was_quoted = True
                else:
                    # Remove inline comments for unquoted values
                    if "#" in value:
                        value = value.split("#")[0].strip()

                # Top-level key (no indentation or minimal indentation)
                if indent == 0:
                    if value or was_quoted:
                        # Simple key: value (including empty quoted strings)
                        result[key] = _parse_simple_value(value)
                    else:
                        # Section header (e.g., "user:", "language:")
                        current_section = key
                        result[current_section] = {}
                elif current_section and indent > 0:
                    # Nested key under current section
                    if value or was_quoted:
                        # Store value (including empty quoted strings)
                        result[current_section][key] = _parse_simple_value(value)
                    else:
                        # Nested section (2-level nesting) - only when no value at all
                        result[current_section][key] = {}

        return result

    def _parse_simple_value(value: str):
        """Parse a simple value string into appropriate Python type."""
        if not value:
            return ""

        # Boolean
        if value.lower() == "true":
            return True
        if value.lower() == "false":
            return False

        # Numeric
        try:
            if "." in value:
                return float(value)
            return int(value)
        except ValueError:
            pass

        return value

    def _merge_configs(base: dict, override: dict) -> dict:
        """Recursively merge two configuration dictionaries."""
        result = base.copy()
        for key, value in override.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = _merge_configs(result[key], value)
            else:
                result[key] = value
        return result

    def _load_yaml_file(file_path: Path) -> dict:
        """Load a YAML file using PyYAML or simple parser."""
        if not file_path.exists():
            return {}

        # Check file size before reading (H2: 10MB limit)
        is_safe, error_msg = check_file_size(file_path)
        if not is_safe:
            # File too large or other error, skip loading
            return {}

        try:
            content = file_path.read_text(encoding="utf-8")
            if HAS_YAML_FALLBACK:
                return yaml_fallback.safe_load(content) or {}
            else:
                return _simple_yaml_parse(content)
        except Exception:
            return {}

    def get_cached_config():
        """Load config with section file merging for complete configuration.

        FIX #245/#243: Properly loads both config.yaml AND section files,
        merging them to provide complete configuration data.

        Priority (highest to lowest):
        1. Section files (.moai/config/sections/*.yaml)
        2. Main config file (.moai/config/config.yaml)
        """
        project_root = find_project_root()
        config_dir = project_root / ".moai" / "config"

        # Start with main config file
        main_config_path = config_dir / "config.yaml"
        config = _load_yaml_file(main_config_path)

        # If main config failed, try JSON fallback
        if not config:
            json_config_path = config_dir / "config.json"
            if json_config_path.exists():
                # Check file size before reading (H2: 10MB limit)
                is_safe, _ = check_file_size(json_config_path)
                if is_safe:
                    try:
                        config = json.loads(json_config_path.read_text(encoding="utf-8"))
                    except (json.JSONDecodeError, OSError):
                        config = {}
                else:
                    config = {}

        # Merge section files (they take priority for their specific keys)
        sections_dir = config_dir / "sections"
        if sections_dir.exists():
            section_files = [
                ("user.yaml", "user"),
                ("language.yaml", "language"),
                ("git-strategy.yaml", "git_strategy"),
                ("project.yaml", "project"),
                ("quality.yaml", "quality"),
                ("system.yaml", "system"),
            ]

            for filename, key in section_files:
                section_path = sections_dir / filename
                section_data = _load_yaml_file(section_path)
                if section_data:
                    # Merge section data into config
                    config = _merge_configs(config, section_data)

        return config if config else None

    def get_cached_spec_progress():
        """Get SPEC progress information - FIXED to use YAML frontmatter parsing"""
        # FIX #3: Use absolute path from find_project_root() to ensure current project only
        project_root = find_project_root()
        specs_dir = project_root / ".moai" / "specs"

        if not specs_dir.exists():
            return {"completed": 0, "total": 0, "percentage": 0}
        try:
            # Only scan SPEC folders in THIS project's .moai/specs/ directory
            spec_folders = [d for d in specs_dir.iterdir() if d.is_dir() and d.name.startswith("SPEC-")]
            total = len(spec_folders)

            # FIX: Parse YAML frontmatter to check for status: completed
            completed = 0
            for folder in spec_folders:
                spec_file = folder / "spec.md"
                if not spec_file.exists():
                    continue

                try:
                    # Read spec.md content
                    content = spec_file.read_text(encoding="utf-8")

                    # Parse YAML frontmatter (between --- delimiters)
                    if content.startswith("---"):
                        yaml_end = content.find("---", 3)
                        if yaml_end > 0:
                            yaml_content = content[3:yaml_end]
                            # Check for status: completed (with or without quotes)
                            if "status: completed" in yaml_content or 'status: "completed"' in yaml_content:
                                completed += 1
                except (OSError, UnicodeDecodeError):
                    # File read failure or encoding error - considered incomplete
                    pass

            percentage = (completed / total * 100) if total > 0 else 0
            return {
                "completed": completed,
                "total": total,
                "percentage": round(percentage, 0),
            }
        except (OSError, PermissionError):
            # Directory access or permission errors
            return {"completed": 0, "total": 0, "percentage": 0}


def should_show_setup_messages() -> bool:
    """Determine whether to show setup completion messages (cached version).

    Logic:
    1. Read .moai/config/config.yaml (using cache)
    2. Check session.suppress_setup_messages flag
    3. If suppress_setup_messages is False, always show messages
    4. If suppress_setup_messages is True:
       - Check if more than 7 days have passed since suppression
       - Show messages if time threshold exceeded

    Uses ConfigCache to avoid repeated config file reads.

    Returns:
        bool: True if messages should be shown, False otherwise
    """
    config = get_cached_config()

    # If config doesn't exist, show messages
    if not config:
        return True

    # Check project initialization status
    if not config.get("project", {}).get("initialized", False):
        return True

    # Check suppress_setup_messages flag
    session_config = config.get("session", {})
    suppress = session_config.get("suppress_setup_messages", False)

    if not suppress:
        # Flag is False, show messages
        return True

    # Flag is True, check time threshold
    suppressed_at_str = session_config.get("setup_messages_suppressed_at")
    if not suppressed_at_str:
        # No timestamp recorded, show messages
        return True

    try:
        suppressed_at = datetime.fromisoformat(suppressed_at_str)
        now = datetime.now(suppressed_at.tzinfo) if suppressed_at.tzinfo else datetime.now()
        days_passed = (now - suppressed_at).days

        # Show messages if threshold exceeded
        return days_passed >= SETUP_MESSAGE_RESCAN_DAYS
    except (ValueError, TypeError):
        # If timestamp is invalid, show messages
        return True


def check_git_initialized() -> bool:
    """Check if git repository is initialized

    Returns:
        bool: True if .git directory exists, False otherwise
    """
    try:
        project_root = find_project_root()
        git_dir = project_root / ".git"
        return git_dir.exists() and git_dir.is_dir()
    except Exception:
        return False


def get_git_info() -> dict[str, Any]:
    """Get comprehensive git information using optimized Git operations manager

    FIXED: Handles git not initialized state properly
    - Branch: Shows helpful message if git not initialized
    - Last Commit: Shows helpful message if git not initialized or no commits

    Uses connection pooling, caching, and parallel execution for optimal performance.
    Falls back to basic implementation if Git manager unavailable.
    """
    # FIX #1 and #4: Check if git is initialized first
    if not check_git_initialized():
        return {
            "branch": "Git not initialized → Run 'moai-adk init' to set up Git repository",
            "last_commit": "Git not initialized → Run 'moai-adk init' to set up Git repository",
            "commit_time": "",
            "changes": 0,
            "git_initialized": False,
        }

    git_manager = get_git_manager()
    if git_manager:
        try:
            # Use optimized Git manager
            project_info = git_manager.get_project_info(use_cache=True)
            branch = project_info.get("branch", "unknown")
            last_commit = project_info.get("last_commit", "unknown")

            # FIX #1: Handle empty branch (no commits yet)
            if not branch or branch == "unknown":
                branch = "No commits yet → Make your first commit"

            # FIX #4: Handle no commits case
            if not last_commit or last_commit == "unknown":
                last_commit = "No commits yet"

            return {
                "branch": branch,
                "last_commit": last_commit,
                "commit_time": project_info.get("commit_time", "unknown"),
                "changes": project_info.get("changes", 0),
                "fetch_time": project_info.get("fetch_time", ""),
                "git_initialized": True,
            }
        except Exception as e:
            logging.warning(f"Git manager failed, falling back: {e}")

    # Fallback to basic Git operations
    try:
        import concurrent.futures
        from concurrent.futures import ThreadPoolExecutor, as_completed

        # Define git commands to run in parallel
        git_commands = [
            (["git", "branch", "--show-current"], "branch"),
            (["git", "rev-parse", "--abbrev-ref", "HEAD"], "head_ref"),
            (["git", "rev-parse", "--short", "HEAD"], "head_commit"),
            (["git", "log", "--pretty=format:%h %s", "-1"], "last_commit"),
            (["git", "log", "--pretty=format:%ar", "-1"], "commit_time"),
            (["git", "status", "--porcelain"], "changes_raw"),
        ]

        # Execute git commands in parallel
        results = {}
        with ThreadPoolExecutor(max_workers=4) as executor:
            # Submit all tasks
            futures = {executor.submit(_run_git_command_fallback, cmd): key for cmd, key in git_commands}

            # Collect results as they complete with overall timeout
            # FIX #254: Add timeout to prevent infinite waiting on stuck git operations
            try:
                for future in as_completed(futures, timeout=8):
                    key = futures[future]
                    try:
                        results[key] = future.result()
                    except (TimeoutError, RuntimeError):
                        # Future execution timeout or runtime errors
                        results[key] = ""
            except concurrent.futures.TimeoutError:
                # Overall timeout exceeded - use whatever results we have
                logging.warning("Git operations timeout after 8 seconds - using partial results")
                # Collect any completed futures before timeout
                for future, key in futures.items():
                    if future.done():
                        try:
                            if key not in results:
                                results[key] = future.result()
                        except (TimeoutError, RuntimeError):
                            results[key] = ""

        # Process results with proper handling for empty values
        branch = results.get("branch", "")
        head_ref = results.get("head_ref", "")
        head_commit = results.get("head_commit", "")
        last_commit = results.get("last_commit", "")

        # FIX: Detect detached HEAD state
        if not branch and head_ref == "HEAD":
            # Detached HEAD state - show commit hash
            branch = f"HEAD detached at {head_commit}"
        elif not branch:
            # No commits yet
            branch = "No commits yet → Make your first commit"

        # FIX #4: Handle no commits case
        if not last_commit:
            last_commit = "No commits yet"

        return {
            "branch": branch,
            "last_commit": last_commit,
            "commit_time": results.get("commit_time", ""),
            "changes": (len(results.get("changes_raw", "").splitlines()) if results.get("changes_raw") else 0),
            "git_initialized": True,
        }

    except (RuntimeError, OSError, TimeoutError):
        # ThreadPoolExecutor, git command, or timeout errors
        return {
            "branch": "Error reading git info",
            "last_commit": "Error reading git info",
            "commit_time": "",
            "changes": 0,
            "git_initialized": True,
        }


def _run_git_command_fallback(cmd: list[str]) -> str:
    """Fallback git command execution"""
    try:
        import subprocess

        result = subprocess.run(cmd, capture_output=True, text=True, timeout=3)
        return result.stdout.strip() if result.returncode == 0 else ""
    except (
        subprocess.TimeoutExpired,
        subprocess.SubprocessError,
        FileNotFoundError,
        OSError,
    ):
        # Git command timeout, subprocess error, or git not found
        return ""


def get_git_strategy_info(config: dict) -> dict:
    """Get git strategy information from config

    FIX #2: NEW FEATURE - Display git strategy information

    Args:
        config: Configuration dictionary

    Returns:
        Dictionary with git_flow and auto_branch information
    """
    if not config:
        return {"git_flow": "unknown", "auto_branch": "unknown"}

    git_strategy = config.get("git_strategy", {})
    mode = git_strategy.get("mode", "manual")

    # Get auto_branch setting from branch_creation config
    branch_creation = git_strategy.get("branch_creation", {})
    auto_enabled = branch_creation.get("auto_enabled", False)

    # Determine auto_branch display
    auto_branch_display = "Yes" if auto_enabled else "No"

    return {"git_flow": mode, "auto_branch": auto_branch_display}


def _parse_version(version_str: str) -> tuple[int, ...]:
    """Parse version string to comparable tuple

    Args:
        version_str: Version string (e.g., "0.25.4")

    Returns:
        Tuple of integers for comparison (e.g., (0, 25, 4))
    """
    try:
        import re

        clean = version_str.lstrip("v")
        parts = [int(x) for x in re.split(r"[^\d]+", clean) if x.isdigit()]
        return tuple(parts) if parts else (0,)
    except (ValueError, AttributeError, TypeError):
        # Version parsing errors (invalid format, None input, type mismatch)
        return (0,)


def _is_newer_version(newer: str, older: str) -> bool:
    """Compare two versions (semantic versioning)

    Args:
        newer: Version that might be newer
        older: Version that might be older

    Returns:
        True if newer > older
    """
    newer_parts = _parse_version(newer)
    older_parts = _parse_version(older)
    return newer_parts > older_parts


def check_version_update() -> tuple[str, bool]:
    """Check if version update is available (fast version using cached data)

    Reuses PyPI cache from Phase 1 (config_health_check.py).
    Falls back to importlib.metadata for installed version.

    Returns:
        (status_indicator, has_update)
        - status_indicator: "(latest)", "(dev)" or "⬆️ X.X.X available"
        - has_update: True if update available
    """
    try:
        import importlib.metadata

        # Get installed version (fast, ~6ms)
        try:
            installed_version = importlib.metadata.version("moai-adk")
        except importlib.metadata.PackageNotFoundError:
            return "(latest)", False

        # Try to load cached PyPI version from Phase 1
        version_cache_file = find_project_root() / ".moai" / "cache" / "version-check.json"
        latest_version = None

        if version_cache_file.exists():
            try:
                cache_data = json.loads(version_cache_file.read_text())
                latest_version = cache_data.get("latest")
            except (json.JSONDecodeError, OSError, UnicodeDecodeError):
                # Cache file read or JSON parsing errors
                pass

        # If no cache or cache is stale, skip check (avoid slow subprocess)
        if not latest_version:
            return "(latest)", False

        # Compare versions with semantic versioning
        if _is_newer_version(latest_version, installed_version):
            # PyPI has newer version (use update icon instead of warning)
            return f"⬆️ {latest_version} available", True
        elif _is_newer_version(installed_version, latest_version):
            # Local version is newer (development version)
            return "(dev)", False
        else:
            # Same version
            return "(latest)", False

    except (ImportError, AttributeError, TypeError):
        # Import errors or unexpected type/attribute errors
        return "(latest)", False


def get_test_info() -> dict[str, Any]:
    """Get test coverage and status information

    NOTE: SessionStart hook must complete quickly (<0.5s).
    Running pytest is too slow (5+ seconds), so we skip it and return unknown status.
    Users can run tests manually with: pytest --cov

    To check test status, use: /moai:test-status (future feature)
    """
    # Skip pytest execution - it's too slow for SessionStart
    return {"coverage": "unknown", "status": "❓"}


def get_spec_progress() -> dict[str, Any]:
    """Get SPEC progress information (cached version)

    Uses ConfigCache to avoid repeated filesystem scans.
    Cache is valid for 5 minutes or until .moai/specs/ is modified.

    Returns:
        Dict with keys: completed, total, percentage
    """
    return get_cached_spec_progress()


def calculate_risk(git_info: dict, spec_progress: dict, test_info: dict) -> str:
    """Calculate overall project risk level using defined thresholds."""
    risk_score = 0

    # Git changes contribute to risk
    if git_info["changes"] > GIT_CHANGES_HIGH_THRESHOLD:
        risk_score += RISK_GIT_CHANGES_HIGH
    elif git_info["changes"] > GIT_CHANGES_MEDIUM_THRESHOLD:
        risk_score += RISK_GIT_CHANGES_MEDIUM

    # SPEC progress contributes to risk
    if spec_progress["percentage"] < SPEC_PROGRESS_LOW:
        risk_score += RISK_SPEC_LOW
    elif spec_progress["percentage"] < SPEC_PROGRESS_MEDIUM:
        risk_score += RISK_SPEC_MEDIUM

    # Test status contributes to risk
    if test_info["status"] != "✅":
        risk_score += RISK_TEST_FAILED
    elif test_info["coverage"] == "unknown":
        risk_score += RISK_COVERAGE_UNKNOWN

    # Determine risk level
    if risk_score >= RISK_SCORE_HIGH:
        return "HIGH"
    elif risk_score >= RISK_SCORE_MEDIUM:
        return "MEDIUM"
    else:
        return "LOW"


def format_project_metadata() -> str:
    """Format project metadata information as a string.

    Returns:
        Formatted project metadata string with version and Git info
    """
    moai_version = "unknown"
    config = get_cached_config()
    if config:
        moai_version = config.get("moai", {}).get("version", "unknown")

    version_status, _has_update = check_version_update()
    return f"📦 Version: {moai_version} {version_status}"


def get_language_info(config: dict) -> dict:
    """Get language configuration information

    Args:
        config: Configuration dictionary

    Returns:
        Dictionary with language info including display name and status
    """
    if not config:
        return {
            "conversation_language": "en",
            "language_name": "English",
            "status": "⚠️ No config",
        }

    lang_config = config.get("language", {})
    conversation_lang = lang_config.get("conversation_language", "en")
    lang_name = lang_config.get("conversation_language_name", "Unknown")

    # Language status indicator (removed Active indicator for cleaner output)
    return {"conversation_language": conversation_lang, "language_name": lang_name}


def load_user_personalization() -> dict:
    """Load user personalization settings using centralized language configuration resolver

    FIX #5: Check for template variables and provide setup guidance

    Uses the new LanguageConfigResolver which provides:
    - Environment variable priority handling
    - Configuration file integration
    - Consistency validation and auto-correction
    - Template variable export capabilities

    Returns:
        Dictionary with user personalization information
    """
    try:
        # Import the centralized language configuration resolver
        from src.moai_adk.core.language_config_resolver import get_resolver

        # Get resolver instance and resolve configuration
        resolver = get_resolver(str(find_project_root()))
        config = resolver.resolve_config()

        # FIX #5: Check if USER_NAME is a template variable or empty
        user_name = config.get("user_name", "")
        has_valid_name = user_name and not user_name.startswith("{{") and not user_name.endswith("}}")

        # Build personalization info using resolved configuration
        personalization = {
            "user_name": user_name if has_valid_name else "",
            "conversation_language": config.get("conversation_language", "en"),
            "conversation_language_name": config.get("conversation_language_name", "English"),
            "agent_prompt_language": config.get("agent_prompt_language", "en"),
            "is_korean": config.get("conversation_language") == "ko",
            "has_personalization": has_valid_name,
            "config_source": config.get("config_source", "default"),
            "personalized_greeting": (resolver.get_personalized_greeting(config) if has_valid_name else ""),
            "needs_setup": not has_valid_name,  # FIX #5: Flag for setup guidance
        }

        # Export template variables for other system components
        template_vars = resolver.export_template_variables(config)

        # Store resolved configuration for session-wide access
        personalization_cache_file = find_project_root() / ".moai" / "cache" / "personalization.json"
        try:
            personalization_cache_file.parent.mkdir(parents=True, exist_ok=True)

            # Store both personalization info and template variables
            cache_data = {
                "personalization": personalization,
                "template_variables": template_vars,
                "resolved_at": datetime.now().isoformat(),
                "config_source": config.get("config_source", "default"),
            }
            personalization_cache_file.write_text(json.dumps(cache_data, ensure_ascii=False, indent=2))

        except (OSError, PermissionError):
            # Cache write errors are non-critical
            pass

        return personalization

    except ImportError:
        # Fallback to basic implementation if resolver not available
        import os

        # Load config from cache or direct file
        config = get_cached_config()

        # Environment variables take priority
        user_name = os.getenv("MOAI_USER_NAME")
        conversation_lang = os.getenv("MOAI_CONVERSATION_LANG")

        # Fallback to config file if environment variables not set
        if user_name is None and config:
            user_name = config.get("user", {}).get("name", "")

        if conversation_lang is None and config:
            conversation_lang = config.get("language", {}).get("conversation_language", "en")

        # FIX #5: Check if USER_NAME is a template variable or empty
        has_valid_name = user_name and not user_name.startswith("{{") and not user_name.endswith("}}")

        # Get language name
        lang_name_map = {
            "ko": "Korean",
            "en": "English",
            "ja": "Japanese",
            "zh": "Chinese",
            "es": "Spanish",
            "fr": "French",
            "de": "German",
            "ru": "Russian",
        }
        lang_name = lang_name_map.get(conversation_lang, "Unknown")

        # Build personalization info
        personalization = {
            "user_name": user_name if has_valid_name else "",
            "conversation_language": conversation_lang or "en",
            "conversation_language_name": lang_name,
            "is_korean": conversation_lang == "ko",
            "has_personalization": has_valid_name,
            "config_source": "fallback",
            "personalized_greeting": (
                f"{user_name}님"
                if has_valid_name and conversation_lang == "ko"
                else user_name
                if has_valid_name
                else ""
            ),
            "needs_setup": not has_valid_name,  # FIX #5: Flag for setup guidance
        }

        # Store for session-wide access
        personalization_cache_file = find_project_root() / ".moai" / "cache" / "personalization.json"
        try:
            personalization_cache_file.parent.mkdir(parents=True, exist_ok=True)
            personalization_cache_file.write_text(json.dumps(personalization, ensure_ascii=False, indent=2))
        except (OSError, PermissionError):
            # Cache write errors are non-critical
            pass

        return personalization


def format_session_output() -> str:
    """Format the complete session start output with proper line alignment (optimized).

    Uses caches for config and SPEC progress to minimize file I/O.
    Parallel git command execution for fast data gathering.
    """
    # Gather information (in parallel for git, cached for config/SPEC)
    git_info = get_git_info()

    # Get config for language and version info
    config = get_cached_config()

    # Load user personalization settings
    personalization = load_user_personalization()

    # Get MoAI version from installed package (not config.json)
    try:
        from moai_adk import __version__ as installed_version

        moai_version = installed_version
    except ImportError:
        # Fallback to config version if package import fails
        moai_version = "unknown"
        if config:
            moai_version = config.get("moai", {}).get("version", "unknown")

    # Get language info
    lang_info = get_language_info(config)

    # FIX #2: Get git strategy info
    git_strategy = get_git_strategy_info(config)

    # Check for version updates (uses Phase 1 cache)
    version_status, _has_update = check_version_update()

    # Format output with each item on separate line (reordered per user request)
    output = [
        "🚀 MoAI-ADK Session Started",
        f"   📦 Version: {moai_version} {version_status}",
        f"   🔄 Changes: {git_info['changes']}",
        f"   🌿 Branch: {git_info['branch']}",
        # FIX #2: Add Git Strategy information
        f"   🔧 Github-Flow: {git_strategy['git_flow']} | Auto Branch: {git_strategy['auto_branch']}",
        f"   🔨 Last Commit: {git_info['last_commit']}",
        f"   🌐 Language: {lang_info['language_name']} ({lang_info['conversation_language']})",
    ]

    # FIX #5: Add personalization or setup guidance (never show template variables)
    # Multilingual support: ko, en, ja, zh
    conv_lang = personalization.get("conversation_language", "en")

    if personalization.get("needs_setup", False):
        # Show setup guidance (based on conversation_language)
        # Guide user to generate project documentation with /moai:0-project
        setup_messages = {
            "ko": "   👋 환영합니다! '/moai:0-project' 명령어로 프로젝트 문서를 생성해주세요",
            "ja": "   👋 ようこそ！'/moai:0-project' コマンドでプロジェクトドキュメントを生成してください",
            "zh": "   👋 欢迎！请运行 '/moai:0-project' 命令生成项目文档",
            "en": "   👋 Welcome! Please run '/moai:0-project' to generate project documentation",
        }
        output.append(setup_messages.get(conv_lang, setup_messages["en"]))
    elif personalization["has_personalization"]:
        user_greeting = personalization.get("personalized_greeting", "")
        user_name = personalization.get("user_name", "")
        display_name = user_greeting if user_greeting else user_name

        # Prevent duplicate honorifics (e.g., "님님" in Korean, "さんさん" in Japanese)
        ko_suffix = "" if display_name.endswith("님") else "님"
        ja_suffix = "" if display_name.endswith("さん") else "さん"

        welcome_back_messages = {
            "ko": f"   👋 다시 오신 것을 환영합니다, {display_name}{ko_suffix}!",
            "ja": f"   👋 おかえりなさい、{display_name}{ja_suffix}！",
            "zh": f"   👋 欢迎回来，{display_name}！",
            "en": f"   👋 Welcome back, {display_name}!",
        }
        output.append(welcome_back_messages.get(conv_lang, welcome_back_messages["en"]))

    # Configuration source is now handled silently for cleaner output
    # Users can check configuration using dedicated tools if needed

    return "\n".join(output)


def main() -> None:
    """Main entry point for enhanced SessionStart hook

    Displays enhanced project information including:
    - Programming language and version
    - Git branch, changes, and last commit with time
    - Git strategy (mode and auto_branch setting)
    - SPEC progress (completed/total)
    - Test coverage and status
    - Risk assessment

    Features:
    - Optimized timeout handling with unified manager
    - Enhanced error handling with graceful degradation
    - Resource monitoring and cleanup
    - Retry mechanisms for transient failures

    Exit Codes:
        0: Success
        1: Error (timeout, JSON parse failure, handler exception)
    """
    # Configure timeout for session start hook
    timeout_config = HookTimeoutConfig(
        policy=TimeoutPolicy.NORMAL,
        custom_timeout_ms=5000,  # 5 seconds
        retry_count=1,
        retry_delay_ms=200,
        graceful_degradation=True,
        memory_limit_mb=100,  # Optional memory limit
    )

    def execute_session_start():
        """Execute session start logic with proper error handling"""
        # Read JSON payload from stdin (for compatibility)
        # Handle Docker/non-interactive environments by checking TTY
        # Note: SessionStart hook receives session info but we don't need it currently
        input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
        _ = json.loads(input_data) if input_data.strip() else {}  # Explicitly ignore

        # Check if setup messages should be shown
        show_messages = should_show_setup_messages()

        # Generate enhanced session output (conditionally)
        session_output = format_session_output() if show_messages else ""

        # Return as system message
        result: dict[str, Any] = {
            "continue": True,
            "systemMessage": session_output,
            "performance": {
                "git_manager_used": get_git_manager() is not None,
                "timeout_manager_used": get_timeout_manager() is not None,
            },
        }

        return result

    # Use unified timeout manager if available
    timeout_manager = get_timeout_manager()
    if timeout_manager:
        try:
            result = timeout_manager.execute_with_timeout(
                "session_start__show_project_info",
                execute_session_start,
                config=timeout_config,
            )

            print(json.dumps(result, ensure_ascii=False))
            sys.exit(0)

        except HookTimeoutError as e:
            # Enhanced timeout error handling
            timeout_response: dict[str, Any] = {
                "continue": True,
                "systemMessage": "⚠️ Session start timeout - continuing without project info",
                "error_details": {
                    "hook_id": e.hook_id,
                    "timeout_seconds": e.timeout_seconds,
                    "execution_time": e.execution_time,
                    "will_retry": e.will_retry,
                },
            }
            print(json.dumps(timeout_response, ensure_ascii=False))
            print(f"SessionStart hook timeout: {e}", file=sys.stderr)
            sys.exit(1)

        except Exception as e:
            # Enhanced error handling with context
            error_response: dict[str, Any] = {
                "continue": True,
                "systemMessage": "⚠️ Session start encountered an error - continuing",
                "error_details": {
                    "error_type": type(e).__name__,
                    "message": str(e),
                    "graceful_degradation": True,
                },
            }
            print(json.dumps(error_response, ensure_ascii=False))
            print(f"SessionStart error: {e}", file=sys.stderr)
            sys.exit(1)

    else:
        # Fallback to legacy timeout handling
        try:
            from lib.timeout import CrossPlatformTimeout
            from lib.timeout import TimeoutError as PlatformTimeoutError

            # Set 5-second timeout
            timeout = CrossPlatformTimeout(5)
            timeout.start()

            try:
                result = execute_session_start()
                print(json.dumps(result))
                sys.exit(0)

            except PlatformTimeoutError:
                # Timeout - return minimal valid response
                timeout_response_legacy: dict[str, Any] = {
                    "continue": True,
                    "systemMessage": "⚠️ Session start timeout - continuing without project info",
                }
                print(json.dumps(timeout_response_legacy))
                print("SessionStart hook timeout after 5 seconds", file=sys.stderr)
                sys.exit(1)

            finally:
                # Always cancel timeout
                timeout.cancel()

        except ImportError:
            # No timeout handling available
            try:
                result = execute_session_start()
                print(json.dumps(result))
                sys.exit(0)
            except Exception as e:
                print(
                    json.dumps(
                        {
                            "continue": True,
                            "systemMessage": "⚠️ Session start completed with errors",
                            "error": str(e),
                        }
                    )
                )
                sys.exit(0)

        except json.JSONDecodeError as e:
            # JSON parse error
            json_error_response: dict[str, Any] = {
                "continue": True,
                "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
            }
            print(json.dumps(json_error_response))
            print(f"SessionStart JSON parse error: {e}", file=sys.stderr)
            sys.exit(1)

        except Exception as e:
            # Unexpected error
            general_error_response: dict[str, Any] = {
                "continue": True,
                "hookSpecificOutput": {"error": f"SessionStart error: {e}"},
            }
            print(json.dumps(general_error_response))
            print(f"SessionStart unexpected error: {e}", file=sys.stderr)
            sys.exit(1)


if __name__ == "__main__":
    main()
