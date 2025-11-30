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

import json
import sys
import logging
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Dict, Optional

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

# Import path utils for project root resolution
from lib.path_utils import find_project_root

# Import unified timeout manager and Git operations manager
try:
    from lib.unified_timeout_manager import (
        get_timeout_manager, hook_timeout_context, HookTimeoutConfig,
        TimeoutPolicy, HookTimeoutError
    )
    from lib.git_operations_manager import get_git_manager, GitOperationType
    from lib.timeout import TimeoutError as PlatformTimeoutError
except ImportError:
    # Fallback implementations if new modules not available
    import signal

    def get_timeout_manager():
        return None

    def hook_timeout_context(hook_name, config=None):
        import contextlib
        @contextlib.contextmanager
        def dummy_context():
            yield
        return dummy_context()

    class HookTimeoutConfig:
        def __init__(self, **kwargs):
            pass

    class TimeoutPolicy:
        FAST = "fast"
        NORMAL = "normal"
        SLOW = "slow"

    class HookTimeoutError(Exception):
        pass

    def get_git_manager():
        return None

    class GitOperationType:
        BRANCH = "branch"
        LOG = "log"
        STATUS = "status"

    class PlatformTimeoutError(Exception):
        pass


# Import config cache
try:
    from core.config_cache import get_cached_config, get_cached_spec_progress
except ImportError:
    # Fallback to direct functions if cache not available
    def get_cached_config():
        config_path = Path(".moai/config/config.json")
        if config_path.exists():
            try:
                return json.loads(config_path.read_text())
            except (json.JSONDecodeError, OSError, UnicodeDecodeError):
                # JSON parsing, file read, or encoding errors
                return None
        return None

    def get_cached_spec_progress():
        specs_dir = find_project_root() / ".moai" / "specs"
        if not specs_dir.exists():
            return {"completed": 0, "total": 0, "percentage": 0}
        try:
            spec_folders = [d for d in specs_dir.iterdir() if d.is_dir() and d.name.startswith("SPEC-")]
            total = len(spec_folders)
            completed = sum(1 for folder in spec_folders if (folder / "spec.md").exists())
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
    1. Read .moai/config/config.json (using cache)
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

    # Flag is True, check time threshold (7 days)
    suppressed_at_str = session_config.get("setup_messages_suppressed_at")
    if not suppressed_at_str:
        # No timestamp recorded, show messages
        return True

    try:
        suppressed_at = datetime.fromisoformat(suppressed_at_str)
        now = datetime.now(suppressed_at.tzinfo) if suppressed_at.tzinfo else datetime.now()
        days_passed = (now - suppressed_at).days

        # Show messages if more than 7 days have passed
        return days_passed >= 7
    except (ValueError, TypeError):
        # If timestamp is invalid, show messages
        return True


def get_git_info() -> Dict[str, Any]:
    """Get comprehensive git information using optimized Git operations manager

    Uses connection pooling, caching, and parallel execution for optimal performance.
    Falls back to basic implementation if Git manager unavailable.
    """
    git_manager = get_git_manager()
    if git_manager:
        try:
            # Use optimized Git manager
            project_info = git_manager.get_project_info(use_cache=True)
            return {
                "branch": project_info.get("branch", "unknown"),
                "last_commit": project_info.get("last_commit", "unknown"),
                "commit_time": project_info.get("commit_time", "unknown"),
                "changes": project_info.get("changes", 0),
                "fetch_time": project_info.get("fetch_time", ""),
            }
        except Exception as e:
            logging.warning(f"Git manager failed, falling back: {e}")

    # Fallback to basic Git operations
    try:
        import subprocess
        from concurrent.futures import ThreadPoolExecutor, as_completed

        # Define git commands to run in parallel
        git_commands = [
            (["git", "branch", "--show-current"], "branch"),
            (["git", "log", "--pretty=format:%h %s", "-1"], "last_commit"),
            (["git", "log", "--pretty=format:%ar", "-1"], "commit_time"),
            (["git", "status", "--porcelain"], "changes_raw"),
        ]

        # Execute git commands in parallel
        results = {}
        with ThreadPoolExecutor(max_workers=4) as executor:
            # Submit all tasks
            futures = {executor.submit(_run_git_command_fallback, cmd): key for cmd, key in git_commands}

            # Collect results as they complete
            for future in as_completed(futures):
                key = futures[future]
                try:
                    results[key] = future.result()
                except (TimeoutError, RuntimeError):
                    # Future execution timeout or runtime errors
                    results[key] = ""

        # Process results
        return {
            "branch": results.get("branch", "unknown"),
            "last_commit": results.get("last_commit", "unknown"),
            "commit_time": results.get("commit_time", "unknown"),
            "changes": (len(results.get("changes_raw", "").splitlines()) if results.get("changes_raw") else 0),
        }

    except (RuntimeError, OSError, TimeoutError):
        # ThreadPoolExecutor, git command, or timeout errors
        return {
            "branch": "unknown",
            "last_commit": "unknown",
            "commit_time": "unknown",
            "changes": 0,
        }


def _run_git_command_fallback(cmd: list[str]) -> str:
    """Fallback git command execution"""
    try:
        import subprocess
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=3)
        return result.stdout.strip() if result.returncode == 0 else ""
    except (subprocess.TimeoutExpired, subprocess.SubprocessError, FileNotFoundError, OSError):
        # Git command timeout, subprocess error, or git not found
        return ""


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
    """Calculate overall project risk level"""
    risk_score = 0

    # Git changes contribute to risk
    if git_info["changes"] > 20:
        risk_score += 10
    elif git_info["changes"] > 10:
        risk_score += 5

    # SPEC progress contributes to risk
    if spec_progress["percentage"] < 50:
        risk_score += 15
    elif spec_progress["percentage"] < 80:
        risk_score += 8

    # Test status contributes to risk
    if test_info["status"] != "✅":
        risk_score += 12
    elif test_info["coverage"] == "unknown":
        risk_score += 5

    # Determine risk level
    if risk_score >= 20:
        return "HIGH"
    elif risk_score >= 10:
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
        return {"conversation_language": "en", "language_name": "English", "status": "⚠️ No config"}

    lang_config = config.get("language", {})
    conversation_lang = lang_config.get("conversation_language", "en")
    lang_name = lang_config.get("conversation_language_name", "Unknown")

    # Language status indicator
    status = "✅ Active" if conversation_lang != "en" else "🌐 English"

    return {
        "conversation_language": conversation_lang,
        "language_name": lang_name,
        "status": status
    }


def load_user_personalization() -> dict:
    """Load user personalization settings using centralized language configuration resolver

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

        # Build personalization info using resolved configuration
        personalization = {
            'user_name': config.get('user_name', ''),
            'conversation_language': config.get('conversation_language', 'en'),
            'conversation_language_name': config.get('conversation_language_name', 'English'),
            'agent_prompt_language': config.get('agent_prompt_language', 'en'),
            'is_korean': config.get('conversation_language') == 'ko',
            'has_personalization': bool(config.get('user_name', '').strip()),
            'config_source': config.get('config_source', 'default'),
            'personalized_greeting': resolver.get_personalized_greeting(config)
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
                "config_source": config.get('config_source', 'default')
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
        user_name = os.getenv('MOAI_USER_NAME')
        conversation_lang = os.getenv('MOAI_CONVERSATION_LANG')

        # Fallback to config file if environment variables not set
        if user_name is None and config:
            user_name = config.get('user', {}).get('name', '')

        if conversation_lang is None and config:
            conversation_lang = config.get('language', {}).get('conversation_language', 'en')

        # Get language name
        lang_name_map = {
            'ko': 'Korean',
            'en': 'English',
            'ja': 'Japanese',
            'zh': 'Chinese',
            'es': 'Spanish',
            'fr': 'French',
            'de': 'German',
            'ru': 'Russian'
        }
        lang_name = lang_name_map.get(conversation_lang, 'Unknown')

        # Build personalization info
        personalization = {
            'user_name': user_name or '',
            'conversation_language': conversation_lang or 'en',
            'conversation_language_name': lang_name,
            'is_korean': conversation_lang == 'ko',
            'has_personalization': bool(user_name),
            'config_source': 'fallback',
            'personalized_greeting': f"{user_name}님" if user_name and conversation_lang == 'ko' else user_name or ''
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
    spec_progress = get_spec_progress()

    # Get config for language and version info
    config = get_cached_config()

    # Load user personalization settings
    personalization = load_user_personalization()

    # Get MoAI version
    moai_version = "unknown"
    if config:
        moai_version = config.get("moai", {}).get("version", "unknown")

    # Get language info
    lang_info = get_language_info(config)

    # Check for version updates (uses Phase 1 cache)
    version_status, _has_update = check_version_update()

    # Format output with each item on separate line
    output = [
        "🚀 MoAI-ADK Session Started",
        f"📦 Version: {moai_version} {version_status}",
        f"🌿 Branch: {git_info['branch']}",
        f"🔄 Changes: {git_info['changes']}",
        f"🎯 SPEC Progress: {spec_progress['completed']}/{spec_progress['total']} ({int(spec_progress['percentage'])}%)",
        f"🔨 Last Commit: {git_info['last_commit']}",
        f"🌐 Language: {lang_info['language_name']} ({lang_info['conversation_language']}) {lang_info['status']}",
    ]

    # Add personalization info if available
    if personalization['has_personalization']:
        user_greeting = personalization.get('personalized_greeting', '')
        if user_greeting:
            greeting = f"👋 Welcome back, {user_greeting}!" if personalization['is_korean'] else f"👋 Welcome back, {user_greeting}!"
        else:
            greeting = f"👋 Welcome back, {personalization['user_name']}!" if personalization['is_korean'] else f"👋 Welcome back, {personalization['user_name']}!"
        output.append(greeting)

    # Configuration source is now handled silently for cleaner output
    # Users can check configuration using dedicated tools if needed

    return "\n".join(output)


def main() -> None:
    """Main entry point for enhanced SessionStart hook

    Displays enhanced project information including:
    - Programming language and version
    - Git branch, changes, and last commit with time
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
        memory_limit_mb=100  # Optional memory limit
    )

    def execute_session_start():
        """Execute session start logic with proper error handling"""
        # Read JSON payload from stdin (for compatibility)
        # Handle Docker/non-interactive environments by checking TTY
        input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
        data = json.loads(input_data) if input_data.strip() else {}

        # Check if setup messages should be shown
        show_messages = should_show_setup_messages()

        # Generate enhanced session output (conditionally)
        session_output = format_session_output() if show_messages else ""

        # Return as system message
        result: Dict[str, Any] = {
            "continue": True,
            "systemMessage": session_output,
            "performance": {
                "git_manager_used": get_git_manager() is not None,
                "timeout_manager_used": get_timeout_manager() is not None
            }
        }

        return result

    # Use unified timeout manager if available
    timeout_manager = get_timeout_manager()
    if timeout_manager:
        try:
            result = timeout_manager.execute_with_timeout(
                "session_start__show_project_info",
                execute_session_start,
                config=timeout_config
            )

            print(json.dumps(result, ensure_ascii=False))
            sys.exit(0)

        except HookTimeoutError as e:
            # Enhanced timeout error handling
            timeout_response: Dict[str, Any] = {
                "continue": True,
                "systemMessage": "⚠️ Session start timeout - continuing without project info",
                "error_details": {
                    "hook_id": e.hook_id,
                    "timeout_seconds": e.timeout_seconds,
                    "execution_time": e.execution_time,
                    "will_retry": e.will_retry
                }
            }
            print(json.dumps(timeout_response, ensure_ascii=False))
            print(f"SessionStart hook timeout: {e}", file=sys.stderr)
            sys.exit(1)

        except Exception as e:
            # Enhanced error handling with context
            error_response: Dict[str, Any] = {
                "continue": True,
                "systemMessage": "⚠️ Session start encountered an error - continuing",
                "error_details": {
                    "error_type": type(e).__name__,
                    "message": str(e),
                    "graceful_degradation": True
                }
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
                timeout_response: Dict[str, Any] = {
                    "continue": True,
                    "systemMessage": "⚠️ Session start timeout - continuing without project info",
                }
                print(json.dumps(timeout_response))
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
                print(json.dumps({
                    "continue": True,
                    "systemMessage": "⚠️ Session start completed with errors",
                    "error": str(e)
                }))
                sys.exit(0)

        except json.JSONDecodeError as e:
            # JSON parse error
            json_error_response: Dict[str, Any] = {
                "continue": True,
                "hookSpecificOutput": {"error": f"JSON parse error: {e}"},
            }
            print(json.dumps(json_error_response))
            print(f"SessionStart JSON parse error: {e}", file=sys.stderr)
            sys.exit(1)

        except Exception as e:
            # Unexpected error
            general_error_response: Dict[str, Any] = {
                "continue": True,
                "hookSpecificOutput": {"error": f"SessionStart error: {e}"},
            }
            print(json.dumps(general_error_response))
            print(f"SessionStart unexpected error: {e}", file=sys.stderr)
            sys.exit(1)


if __name__ == "__main__":
    main()
