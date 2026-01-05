#!/usr/bin/env python3
"""Project metadata utilities

Project information inquiry (language, Git, SPEC progress, etc.)
"""

import json
import signal
import socket
import subprocess
from contextlib import contextmanager
from pathlib import Path
from typing import Any

import yaml

# Import TimeoutError from lib.timeout (canonical definition)
try:
    from lib.timeout import TimeoutError  # noqa: F401
except ImportError:
    # Fallback if lib.timeout not available
    class TimeoutError(Exception):  # type: ignore[no-redef]
        """Signal-based timeout exception"""

        pass


# Import find_project_root from path_utils (canonical implementation)
# This consolidates project root detection to a single source of truth
try:
    from lib.path_utils import find_project_root as _canonical_find_project_root
except ImportError:
    # Fallback if path_utils not available
    _canonical_find_project_root = None  # type: ignore


# Cache directory for version check results
CACHE_DIR_NAME = ".moai/cache"


def find_project_root(start_path: str | Path = ".") -> Path:
    """Find MoAI-ADK project root by searching upward for project markers.

    This is a wrapper around path_utils.find_project_root() for backward compatibility.
    The canonical implementation in path_utils provides:
    - Environment variable support (MOAI_PROJECT_ROOT, CLAUDE_PROJECT_DIR)
    - Caching for performance
    - Multiple project marker detection

    Args:
        start_path: Starting directory (default: current directory)

    Returns:
        Project root Path. If not found, returns start_path as absolute path.

    Examples:
        >>> find_project_root(".")
        Path("/Users/user/my-project")
        >>> find_project_root(".claude/hooks/alfred")
        Path("/Users/user/my-project")  # Found root 3 levels up

    Note:
        For new code, prefer using path_utils.find_project_root() directly.
    """
    # Use canonical implementation if available
    if _canonical_find_project_root is not None:
        path = Path(start_path).resolve() if start_path != "." else None
        return _canonical_find_project_root(path)

    # Fallback implementation if path_utils not available
    current = Path(start_path).resolve()
    max_depth = 10  # Prevent infinite loop

    for _ in range(max_depth):
        # Check for .moai/config/config.yaml (primary indicator)
        if (current / ".moai" / "config" / "config.yaml").exists():
            return current

        # Check for CLAUDE.md (secondary indicator)
        if (current / "CLAUDE.md").exists():
            return current

        # Move up one level
        parent = current.parent
        if parent == current:  # Reached filesystem root
            break
        current = parent

    # Not found - return start_path as absolute
    return Path(start_path).resolve()


@contextmanager
def timeout_handler(seconds: int):
    """Hard timeout using SIGALRM (works on Unix systems including macOS)

    This uses kernel-level signal to interrupt ANY blocking operation,
    even if subprocess.run() timeout fails on macOS.

    Args:
        seconds: Timeout duration in seconds

    Raises:
        TimeoutError: If operation exceeds timeout
    """

    def _handle_timeout(signum, frame):
        raise TimeoutError(f"Operation timed out after {seconds} seconds")

    # Set the signal handler
    old_handler = signal.signal(signal.SIGALRM, _handle_timeout)
    signal.alarm(seconds)
    try:
        yield
    finally:
        signal.alarm(0)  # Disable alarm
        signal.signal(signal.SIGALRM, old_handler)


def detect_language(cwd: str) -> str:
    """Detect project language (supports 20 items languages)

    Browse the File system to detect your project's main development language.
    First, check configuration files such as pyproject.toml and tsconfig.json.
    Apply TypeScript first principles (if tsconfig.json exists).

    Args:
        cwd: Project root directory path (both absolute and relative paths are possible)

    Returns:
        str: Detected language name (lowercase). If detection fails, "Unknown Language" is returned.
        Supported languages: python, typescript, javascript, java, go, rust,
                  dart, swift, kotlin, php, ruby, elixir, scala,
                  clojure, cpp, c, csharp, haskell, shell, lua

    Examples:
        >>> detect_language("/path/to/python/project")
        'python'
        >>> detect_language("/path/to/typescript/project")
        'typescript'
        >>> detect_language("/path/to/unknown/project")
        'Unknown Language'

    TDD History:
        - RED: Write a 21 items language detection test (20 items language + 1 items unknown)
        - GREEN: 20 items language + unknown implementation, all tests passed
        - REFACTOR: Optimize file inspection order, apply TypeScript priority principle
    """
    cwd_path = Path(cwd)

    # Language detection mapping
    language_files = {
        "pyproject.toml": "python",
        "tsconfig.json": "typescript",
        "package.json": "javascript",
        "pom.xml": "java",
        "go.mod": "go",
        "Cargo.toml": "rust",
        "pubspec.yaml": "dart",
        "Package.swift": "swift",
        "build.gradle.kts": "kotlin",
        "composer.json": "php",
        "Gemfile": "ruby",
        "mix.exs": "elixir",
        "build.sbt": "scala",
        "project.clj": "clojure",
        "CMakeLists.txt": "cpp",
        "Makefile": "c",
    }

    # Check standard language files
    for file_name, language in language_files.items():
        if (cwd_path / file_name).exists():
            # Special handling for package.json - prefer typescript if tsconfig exists
            if file_name == "package.json" and (cwd_path / "tsconfig.json").exists():
                return "typescript"
            return language

    # Check for C# project files (*.csproj)
    if any(cwd_path.glob("*.csproj")):
        return "csharp"

    # Check for Haskell project files (*.cabal)
    if any(cwd_path.glob("*.cabal")):
        return "haskell"

    # Check for Shell scripts (*.sh)
    if any(cwd_path.glob("*.sh")):
        return "shell"

    # Check for Lua files (*.lua)
    if any(cwd_path.glob("*.lua")):
        return "lua"

    return "Unknown Language"


def _run_git_command(args: list[str], cwd: str, timeout: int = 2) -> str:
    """Git command execution with HARD timeout protection

    Safely execute Git commands and return output.
    Uses SIGALRM (kernel-level interrupt) to handle macOS subprocess timeout bug.
    Eliminates code duplication and provides consistent error handling.

    Args:
        args: Git command argument list (git adds automatically)
        cwd: Execution directory path
        timeout: Timeout in seconds (default: 2 seconds)

    Returns:
        str: Git command output (stdout, removing leading and trailing spaces)

    Raises:
        subprocess.TimeoutExpired: Timeout exceeded (via TimeoutError)
        subprocess.CalledProcessError: Git command failed

    Examples:
        >>> _run_git_command(["branch", "--show-current"], ".")
        'main'

    TDD History:
        - RED: Git command hang scenario test
        - GREEN: SIGALRM-based timeout implementation
        - REFACTOR: Exception conversion to subprocess.TimeoutExpired
    """
    try:
        with timeout_handler(timeout):
            result = subprocess.run(
                ["git"] + args,
                cwd=cwd,
                capture_output=True,
                text=True,
                check=False,  # Don't raise on non-zero exit - we'll check manually
            )

            # Check exit code manually
            if result.returncode != 0:
                raise subprocess.CalledProcessError(result.returncode, ["git"] + args, result.stdout, result.stderr)

            return result.stdout.strip()

    except TimeoutError:
        # Convert to subprocess.TimeoutExpired for consistent error handling
        raise subprocess.TimeoutExpired(["git"] + args, timeout)


def get_git_info(cwd: str) -> dict[str, Any]:
    """Gather Git repository information

    View the current status of a Git repository.
    Returns the branch name, commit hash, number of changes, and last commit message.
    If it is not a Git repository, it returns an empty dictionary.

    Args:
        cwd: Project root directory path

    Returns:
        Git information dictionary. Includes the following keys:
        - branch: Current branch name (str)
        - commit: Current commit hash (str, full hash)
        - changes: Number of changed files (int, staged + unstaged)
        - last_commit: Last commit message (str, subject only)

        Empty dictionary {} if it is not a Git repository or the query fails.

    Examples:
        >>> get_git_info("/path/to/git/repo")
        {'branch': 'main', 'commit': 'abc123...', 'changes': 3, 'last_commit': 'Fix bug'}
        >>> get_git_info("/path/to/non-git")
        {}

    Notes:
        - Timeout: 2 seconds for each Git command
        - Security: Safe execution with subprocess.run(shell=False)
        - Error handling: Returns an empty dictionary in case of all exceptions
        - Commit message limited to 50 characters for display purposes

    TDD History:
        - RED: 3 items scenario test (Git repo, non-Git, error)
        - GREEN: Implementation of subprocess-based Git command execution
        - REFACTOR: Add timeout (2 seconds), strengthen exception handling, remove duplicates with helper function
        - UPDATE: Added last_commit message field for SessionStart display
    """
    try:
        # Check if it's a git repository
        _run_git_command(["rev-parse", "--git-dir"], cwd)

        # Get branch name, commit hash, and changes
        branch = _run_git_command(["branch", "--show-current"], cwd)
        commit = _run_git_command(["rev-parse", "HEAD"], cwd)
        status_output = _run_git_command(["status", "--short"], cwd)
        changes = len([line for line in status_output.splitlines() if line])

        # Get last commit message (subject only, limited to 50 chars)
        last_commit = _run_git_command(["log", "-1", "--format=%s"], cwd)
        if len(last_commit) > 50:
            last_commit = last_commit[:47] + "..."

        return {
            "branch": branch,
            "commit": commit,
            "changes": changes,
            "last_commit": last_commit,
        }

    except (
        subprocess.TimeoutExpired,
        subprocess.CalledProcessError,
        FileNotFoundError,
    ):
        return {}


def count_specs(cwd: str) -> dict[str, int]:
    """SPEC File count and progress calculation

    Browse the .moai/specs/ directory to find the number of SPEC Files and
    Counts the number of SPECs with status: completed.

    Args:
        cwd: Project root directory path (or any subdirectory, will search upward)

    Returns:
        SPEC progress dictionary. Includes the following keys:
        - completed: Number of completed SPECs (int)
        - total: total number of SPECs (int)
        - percentage: completion percentage (int, 0~100)

        All 0 if .moai/specs/ directory does not exist

    Examples:
        >>> count_specs("/path/to/project")
        {'completed': 2, 'total': 5, 'percentage': 40}
        >>> count_specs("/path/to/no-specs")
        {'completed': 0, 'total': 0, 'percentage': 0}

    Notes:
        - SPEC File Location: .moai/specs/SPEC-{ID}/spec.md
        - Completion condition: Include "status: completed" in YAML front matter
        - If parsing fails, the SPEC is considered incomplete.
        - Automatically finds project root to locate .moai/specs/

    TDD History:
        - RED: 5 items scenario test (0/0, 2/5, 5/5, no directory, parsing error)
        - GREEN: SPEC search with Path.iterdir(), YAML parsing implementation
        - REFACTOR: Strengthened exception handling, improved percentage calculation safety
        - UPDATE: Add project root detection for consistent path resolution
    """
    # Find project root to ensure we read specs from correct location
    project_root = find_project_root(cwd)
    specs_dir = project_root / ".moai" / "specs"

    if not specs_dir.exists():
        return {"completed": 0, "total": 0, "percentage": 0}

    completed = 0
    total = 0

    for spec_dir in specs_dir.iterdir():
        if not spec_dir.is_dir() or not spec_dir.name.startswith("SPEC-"):
            continue

        spec_file = spec_dir / "spec.md"
        if not spec_file.exists():
            continue

        total += 1

        # Parse YAML front matter
        try:
            content = spec_file.read_text()
            if content.startswith("---"):
                yaml_end = content.find("---", 3)
                if yaml_end > 0:
                    yaml_content = content[3:yaml_end]
                    if "status: completed" in yaml_content:
                        completed += 1
        except (OSError, UnicodeDecodeError):
            # File read failure or encoding error - considered incomplete
            pass

    percentage = int(completed / total * 100) if total > 0 else 0

    return {
        "completed": completed,
        "total": total,
        "percentage": percentage,
    }


def get_project_language(cwd: str) -> str:
    """Determine the primary project language (prefers config.yaml).

    Args:
        cwd: Project root directory (or any subdirectory, will search upward).

    Returns:
        Language string in lower-case.

    Notes:
        - Reads ``.moai/config/config.yaml`` first for a quick answer.
        - Falls back to ``detect_language`` if configuration is missing.
        - Automatically finds project root to locate .moai/config/config.yaml
    """
    # Find project root to ensure we read config from correct location
    project_root = find_project_root(cwd)
    config_path = project_root / ".moai" / "config" / "config.yaml"
    if config_path.exists():
        try:
            config = yaml.safe_load(config_path.read_text()) or {}
            lang = config.get("language", "")
            if lang:
                return lang
        except (OSError, yaml.YAMLError):
            # Fall back to detection on parse errors
            pass

    # Fall back to the original language detection routine (use project root)
    return detect_language(str(project_root))


def _validate_project_structure(cwd: str) -> bool:
    """Validate that project has required MoAI-ADK structure

    Args:
        cwd: Project root directory path

    Returns:
        bool: True if .moai/config/config.yaml exists, False otherwise
    """
    project_root = find_project_root(cwd)
    return (project_root / ".moai" / "config" / "config.yaml").exists()


def get_version_check_config(cwd: str) -> dict[str, Any]:
    """Read version check configuration from .moai/config/config.yaml

    Returns version check settings with sensible defaults.
    Supports frequency-based cache TTL configuration.

    Args:
        cwd: Project root directory path

    Returns:
        dict with keys:
            - "enabled": Boolean (default: True)
            - "frequency": "always" | "daily" | "weekly" | "never" (default: "daily")
            - "cache_ttl_hours": TTL in hours based on frequency

    Frequency to TTL mapping:
        - "always": 0 hours (no caching)
        - "daily": 24 hours
        - "weekly": 168 hours (7 days)
        - "never": infinity (never check)

    TDD History:
        - RED: 8 test scenarios (defaults, custom, disabled, TTL, etc.)
        - GREEN: Minimal config reading with defaults
        - REFACTOR: Add validation and error handling
    """
    # TTL mapping by frequency
    ttl_by_frequency = {"always": 0, "daily": 24, "weekly": 168, "never": float("inf")}

    # Default configuration
    defaults = {"enabled": True, "frequency": "daily", "cache_ttl_hours": 24}

    # Find project root to ensure we read config from correct location
    project_root = find_project_root(cwd)
    config_path = project_root / ".moai" / "config" / "config.yaml"
    if not config_path.exists():
        return defaults

    try:
        config = yaml.safe_load(config_path.read_text()) or {}

        # Extract moai.version_check section
        moai_config = config.get("moai", {})
        version_check_config = moai_config.get("version_check", {})

        # Read enabled flag (default: True)
        enabled = version_check_config.get("enabled", defaults["enabled"])

        # Read frequency (default: "daily")
        frequency = moai_config.get("update_check_frequency", defaults["frequency"])

        # Validate frequency
        if frequency not in ttl_by_frequency:
            frequency = defaults["frequency"]

        # Calculate TTL from frequency
        cache_ttl_hours = ttl_by_frequency[frequency]

        # Allow explicit cache_ttl_hours override
        if "cache_ttl_hours" in version_check_config:
            cache_ttl_hours = version_check_config["cache_ttl_hours"]

        return {
            "enabled": enabled,
            "frequency": frequency,
            "cache_ttl_hours": cache_ttl_hours,
        }

    except (OSError, json.JSONDecodeError, KeyError):
        # Config read or parse error - return defaults
        return defaults


def is_network_available(timeout_seconds: float = 0.1) -> bool:
    """Quick network availability check using socket.

    Does NOT check PyPI specifically, just basic connectivity.
    Returns immediately on success (< 50ms typically).
    Returns False on any error without raising exceptions.

    Args:
        timeout_seconds: Socket timeout in seconds (default 0.1s)

    Returns:
        True if network appears available, False otherwise

    Examples:
        >>> is_network_available()
        True  # Network is available
        >>> is_network_available(timeout_seconds=0.001)
        False  # Timeout too short, returns False

    TDD History:
        - RED: 3 test scenarios (success, failure, timeout)
        - GREEN: Minimal socket.create_connection implementation
        - REFACTOR: Add error handling for all exception types
    """
    try:
        # Try connecting to Google's public DNS server (8.8.8.8:53)
        # This is a reliable host that's typically reachable
        connection = socket.create_connection(("8.8.8.8", 53), timeout=timeout_seconds)
        connection.close()
        return True
    except (socket.timeout, OSError, Exception):
        # Any connection error means network is unavailable
        # This includes: timeout, connection refused, network unreachable, etc.
        return False


def is_major_version_change(current: str, latest: str) -> bool:
    """Detect if version change is a major version bump.

    A major version change is when the first (major) component increases:
    - 0.8.1 → 1.0.0: True (0 → 1)
    - 1.2.3 → 2.0.0: True (1 → 2)
    - 0.8.1 → 0.9.0: False (0 → 0, minor changed)
    - 1.2.3 → 1.3.0: False (1 → 1)

    Args:
        current: Current version string (e.g., "0.8.1")
        latest: Latest version string (e.g., "1.0.0")

    Returns:
        True if major version increased, False otherwise

    Examples:
        >>> is_major_version_change("0.8.1", "1.0.0")
        True
        >>> is_major_version_change("0.8.1", "0.9.0")
        False
        >>> is_major_version_change("dev", "1.0.0")
        False  # Invalid versions return False

    TDD History:
        - RED: 4 test scenarios (0→1, 1→2, minor, invalid)
        - GREEN: Minimal version parsing and comparison
        - REFACTOR: Improve error handling for invalid versions
    """
    try:
        # Parse version strings into integer components
        current_parts = [int(x) for x in current.split(".")]
        latest_parts = [int(x) for x in latest.split(".")]

        # Compare major version (first component)
        if len(current_parts) >= 1 and len(latest_parts) >= 1:
            return latest_parts[0] > current_parts[0]

        # If parsing succeeds but empty, no major change
        return False

    except (ValueError, AttributeError, IndexError):
        # Invalid version format - return False (no exception)
        return False


def get_package_version_info(cwd: str = ".") -> dict[str, Any]:
    """Check MoAI-ADK current and latest version with caching and offline support.

    ⭐ CRITICAL GUARANTEE: This function ALWAYS returns the current installed version.
    Network failures, cache issues, and timeouts NEVER result in "unknown" version.

    Execution flow:
    1. Get current installed version (ALWAYS succeeds) ← CRITICAL
    2. Build minimal result with current version
    3. Try to load from cache (< 50ms) - optional enhancement
    4. If cache valid, return cached latest info
    5. If cache invalid/miss, optionally query PyPI - optional enhancement
    6. Save result to cache for next time - optional

    Args:
        cwd: Project root directory (for cache location)

    Returns:
        dict with keys:
            - "current": Current installed version (ALWAYS valid, never empty)
            - "latest": Latest version available on PyPI (may be "unknown")
            - "update_available": Boolean indicating if update is available
            - "upgrade_command": Recommended upgrade command (if update available)
            - "release_notes_url": URL to release notes
            - "is_major_update": Boolean indicating major version change

    Guarantees:
        - Cache hit (< 24 hours): Returns in ~20ms, no network access ✓
        - Cache miss + online: Query PyPI (1s timeout), cache result ✓
        - Cache miss + offline: Return current version only (~100ms) ✓
        - Network timeout: Returns current + "unknown" latest (~50ms) ✓
        - Any exception: Always returns current version ✓

    TDD History:
        - RED: 5 test scenarios (network detection, cache integration, offline mode)
        - GREEN: Integrate VersionCache with network detection
        - REFACTOR: Extract cache directory constant, improve error handling
    """
    import importlib.util
    import urllib.error
    import urllib.request
    from importlib.metadata import PackageNotFoundError, version

    # Import VersionCache from the same directory (using dynamic import for testing compatibility)
    try:
        version_cache_path = Path(__file__).parent / "version_cache.py"
        spec = importlib.util.spec_from_file_location("version_cache", version_cache_path)
        if spec and spec.loader:
            version_cache_module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(version_cache_module)
            version_cache_class = version_cache_module.VersionCache
        else:
            # Skip caching if module can't be loaded
            version_cache_class = None
    except (ImportError, OSError):
        # Graceful degradation: skip caching on import errors
        version_cache_class = None

    # 1. Find project root (ensure cache is always in correct location)
    # This prevents creating .moai/cache in wrong locations when hooks run
    # from subdirectories like .claude/hooks/alfred/
    project_root = find_project_root(cwd)

    # 2. Initialize cache (skip if VersionCache couldn't be imported)
    cache_dir = project_root / CACHE_DIR_NAME
    version_cache = version_cache_class(cache_dir) if version_cache_class else None

    # 2. Get current installed version first (needed for cache validation)
    current_version = "unknown"
    try:
        current_version = version("moai-adk")
    except PackageNotFoundError:
        current_version = "dev"
        # Dev mode - skip cache and return immediately
        return {
            "current": "dev",
            "latest": "unknown",
            "update_available": False,
            "upgrade_command": "",
        }

    # 3. Try to load from cache (fast path with version validation)
    if version_cache and version_cache.is_valid():
        cached_info = version_cache.load()
        if cached_info:
            # Only use cache if the cached version matches current installed version
            # This prevents stale cache when package is upgraded locally
            if cached_info.get("current") == current_version:
                # Ensure new fields exist for backward compatibility
                if "release_notes_url" not in cached_info:
                    # Add missing fields to old cached data
                    cached_info.setdefault("release_notes_url", None)
                    cached_info.setdefault("is_major_update", False)
                return cached_info
            # else: cache is stale (version changed), fall through to re-check

    # 4. Cache miss or stale - need to query PyPI
    result = {
        "current": current_version,
        "latest": "unknown",
        "update_available": False,
        "upgrade_command": "",
    }

    # 5. Check if version check is enabled in config
    config = get_version_check_config(cwd)
    if not config["enabled"]:
        # Version check disabled - return only current version
        return result

    # 6. Check network before PyPI query
    if not is_network_available():
        # Offline mode - return current version only
        return result

    # 7. Network available - query PyPI
    pypi_data = None
    try:
        with timeout_handler(1):
            url = "https://pypi.org/pypi/moai-adk/json"
            headers = {"Accept": "application/json"}
            req = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(req, timeout=0.8) as response:
                pypi_data = json.load(response)
                result["latest"] = pypi_data.get("info", {}).get("version", "unknown")

                # Extract release notes URL from project_urls
                try:
                    project_urls = pypi_data.get("info", {}).get("project_urls", {})
                    release_url = project_urls.get("Changelog", "")
                    if not release_url:
                        # Fallback to GitHub releases URL pattern
                        release_url = f"https://github.com/modu-ai/moai-adk/releases/tag/v{result['latest']}"
                    result["release_notes_url"] = release_url
                except (KeyError, AttributeError, TypeError):
                    result["release_notes_url"] = None

    except (urllib.error.URLError, TimeoutError, Exception):
        # PyPI query failed - return current version
        result["release_notes_url"] = None
        pass

    # 7. Compare versions (simple comparison)
    if result["current"] != "unknown" and result["latest"] != "unknown":
        try:
            # Parse versions for comparison
            current_str = str(result["current"])
            latest_str = str(result["latest"])

            current_parts = [int(x) for x in current_str.split(".")]
            latest_parts = [int(x) for x in latest_str.split(".")]

            # Pad shorter version with zeros
            max_len = max(len(current_parts), len(latest_parts))
            current_parts.extend([0] * (max_len - len(current_parts)))
            latest_parts.extend([0] * (max_len - len(latest_parts)))

            if latest_parts > current_parts:
                result["update_available"] = True
                result["upgrade_command"] = "uv tool upgrade moai-adk"

                # Detect major version change
                result["is_major_update"] = is_major_version_change(current_str, latest_str)
            else:
                result["is_major_update"] = False
        except (ValueError, AttributeError):
            # Version parsing failed - skip comparison
            result["is_major_update"] = False
            pass

    # 8. Save result to cache (if caching is available)
    if version_cache:
        version_cache.save(result)

    return result


__all__ = [
    "find_project_root",
    "detect_language",
    "get_git_info",
    "count_specs",
    "get_project_language",
    "get_version_check_config",
    "is_network_available",
    "is_major_version_change",
    "get_package_version_info",
]
