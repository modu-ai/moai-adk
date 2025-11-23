#!/usr/bin/env python3
# SessionStart Hook: Enhanced Project Information
"""SessionStart Hook: Enhanced Project Information

Claude Code Event: SessionStart
Purpose: Display enhanced project status with Git info, test status, and SPEC progress
Execution: Triggered automatically when Claude Code session begins

Enhanced Features:
- Last commit information with relative time
- Test coverage and status
- Risk assessment
- Formatted output with clear sections
"""

import json
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any

# Setup import path for shared modules
HOOKS_DIR = Path(__file__).parent
LIB_DIR = HOOKS_DIR / "lib"
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

# Try to import existing modules, provide fallbacks if not available
try:
    from lib.timeout import CrossPlatformTimeout
    from lib.timeout import TimeoutError as PlatformTimeoutError
except ImportError:
    # Fallback timeout implementation

    class CrossPlatformTimeout:  # type: ignore[no-redef]
        def __init__(self, seconds):
            self.seconds = seconds

        def start(self):
            pass

        def cancel(self):
            pass

    class PlatformTimeoutError(Exception):  # type: ignore[no-redef]
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
            except Exception:
                return None
        return None

    def get_cached_spec_progress():
        specs_dir = Path.cwd() / ".moai" / "specs"
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
                "percentage": round(percentage, 0)
            }
        except Exception:
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


def _run_git_command(cmd: list[str]) -> str:
    """Run a single git command with timeout"""
    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=3
        )
        return result.stdout.strip() if result.returncode == 0 else ""
    except Exception:
        return ""


def get_git_cache_file() -> Path:
    """Get path to git info cache file"""
    cache_dir = Path.cwd() / ".moai" / "cache"
    cache_dir.mkdir(parents=True, exist_ok=True)
    return cache_dir / "git-info.json"


def load_git_cache() -> dict[str, Any] | None:
    """Load git info cache if valid (< 1 minute old)

    Returns:
        Cached git data if valid, None otherwise
    """
    try:
        cache_file = get_git_cache_file()
        if not cache_file.exists():
            return None

        cache_data = json.loads(cache_file.read_text())
        last_check = cache_data.get("last_check")

        if not last_check:
            return None

        # Cache is valid for 1 minute (git changes are frequent)
        last_check_dt = datetime.fromisoformat(last_check)
        if datetime.now() - last_check_dt < timedelta(minutes=1):
            return cache_data

        return None
    except Exception:
        return None


def save_git_cache(data: dict[str, Any]) -> None:
    """Save git info cache with timestamp"""
    try:
        cache_file = get_git_cache_file()
        cache_file.parent.mkdir(parents=True, exist_ok=True)

        cache_data = {
            **data,
            "last_check": datetime.now().isoformat()
        }
        cache_file.write_text(json.dumps(cache_data, indent=2))
    except Exception:
        pass  # Silently fail on cache write


def get_git_info() -> dict[str, Any]:
    """Get comprehensive git information with parallel execution and caching

    Uses ThreadPoolExecutor to run git commands in parallel (47ms → ~20ms).
    Caches results for 1 minute to avoid redundant git queries.
    """
    # Try cache first
    cached = load_git_cache()
    if cached:
        # Remove cache metadata before returning
        result = {k: v for k, v in cached.items() if k != "last_check"}
        return result

    try:
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
            futures = {
                executor.submit(_run_git_command, cmd): key
                for cmd, key in git_commands
            }

            # Collect results as they complete
            for future in as_completed(futures):
                key = futures[future]
                try:
                    results[key] = future.result()
                except Exception:
                    results[key] = ""

        # Process results
        git_data = {
            "branch": results.get("branch", "unknown"),
            "last_commit": results.get("last_commit", "unknown"),
            "commit_time": results.get("commit_time", "unknown"),
            "changes": len(results.get("changes_raw", "").splitlines()) if results.get("changes_raw") else 0
        }

        # Cache the results
        save_git_cache(git_data)

        return git_data

    except Exception:
        return {
            "branch": "unknown",
            "last_commit": "unknown",
            "commit_time": "unknown",
            "changes": 0
        }


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
    except Exception:
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
        version_cache_file = Path.cwd() / ".moai" / "cache" / "version-check.json"
        latest_version = None

        if version_cache_file.exists():
            try:
                cache_data = json.loads(version_cache_file.read_text())
                latest_version = cache_data.get("latest")
            except Exception:
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

    except Exception:
        return "(latest)", False


def get_test_info() -> dict[str, Any]:
    """Get test coverage and status information

    NOTE: SessionStart hook must complete quickly (<0.5s).
    Running pytest is too slow (5+ seconds), so we skip it and return unknown status.
    Users can run tests manually with: pytest --cov

    To check test status, use: /alfred:test-status (future feature)
    """
    # Skip pytest execution - it's too slow for SessionStart
    return {
        "coverage": "unknown",
        "status": "❓"
    }


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


def display_project_info() -> None:
    """Display project information to stdout.

    Returns:
        None
    """
    # This function is a placeholder for future display logic
    pass


def format_session_output() -> str:
    """Format the complete session start output with proper line alignment (optimized).

    Uses caches for config and SPEC progress to minimize file I/O.
    Parallel git command execution for fast data gathering.
    """
    # Gather information (in parallel for git, cached for config/SPEC)
    git_info = get_git_info()
    spec_progress = get_spec_progress()

    # Get MoAI version from cached config
    moai_version = "unknown"
    config = get_cached_config()
    if config:
        moai_version = config.get("moai", {}).get("version", "unknown")

    # Check for version updates (uses Phase 1 cache)
    version_status, _has_update = check_version_update()

    # Format output with each item on separate line
    output = [
        "🚀 MoAI-ADK Session Started",
        f"📦 Version: {moai_version} {version_status}",
        f"🌿 Branch: {git_info['branch']}",
        f"🔄 Changes: {git_info['changes']}",
        f"🎯 SPEC Progress: {spec_progress['completed']}/{spec_progress['total']} ({int(spec_progress['percentage'])}%)",
        f"🔨 Last Commit: {git_info['last_commit']}"
    ]

    return "\n".join(output)


def main() -> None:
    """Main entry point for enhanced SessionStart hook

    Displays enhanced project information including:
    - Programming language and version
    - Git branch, changes, and last commit with time
    - SPEC progress (completed/total)
    - Test coverage and status
    - Risk assessment

    Exit Codes:
        0: Success
        1: Error (timeout, JSON parse failure, handler exception)
    """
    # Set 5-second timeout
    timeout = CrossPlatformTimeout(5)
    timeout.start()

    try:
        # Read JSON payload from stdin (for compatibility)
        # Handle Docker/non-interactive environments by checking TTY
        input_data = sys.stdin.read() if not sys.stdin.isatty() else "{}"
        _data = json.loads(input_data) if input_data.strip() else {}

        # Check if setup messages should be shown
        show_messages = should_show_setup_messages()

        # Generate enhanced session output (conditionally)
        session_output = format_session_output() if show_messages else ""

        # Return as system message
        result: dict[str, Any] = {
            "continue": True,
            "systemMessage": session_output
        }

        print(json.dumps(result))
        sys.exit(0)

    except PlatformTimeoutError:
        # Timeout - return minimal valid response
        timeout_response: dict[str, Any] = {
            "continue": True,
            "systemMessage": "⚠️ Session start timeout - continuing without project info",
        }
        print(json.dumps(timeout_response))
        print("SessionStart hook timeout after 5 seconds", file=sys.stderr)
        sys.exit(1)

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

    finally:
        # Always cancel timeout
        timeout.cancel()


if __name__ == "__main__":
    main()
