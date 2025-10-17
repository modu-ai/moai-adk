#!/usr/bin/env python3
"""TAG search and verification system

TAG 검색, 체인 검증, 라이브러리 버전 캐싱
"""

import json
import subprocess
import time
from pathlib import Path

# TAG search cache: {pattern: (results, mtime_hash, cached_at)}
_tag_cache: dict[str, tuple[list[dict], float, float]] = {}

# Library version cache: {lib_name: (version, timestamp)}
_lib_version_cache: dict[str, tuple[str, float]] = {}


def _get_dir_mtime_hash(paths: list[str]) -> float:
    """Calculate a directory-wide mtime hash used for cache invalidation.

    Args:
        paths: List of directory paths to inspect.

    Returns:
        Highest modification timestamp (float) across all files.

    Notes:
        - Any file change bumps the hash and invalidates the cache.
        - Missing or inaccessible directories are ignored.
    """
    max_mtime = 0.0
    for path in paths:
        path_obj = Path(path)
        if not path_obj.exists():
            continue

        try:
            for file_path in path_obj.rglob("*"):
                if file_path.is_file():
                    max_mtime = max(max_mtime, file_path.stat().st_mtime)
        except (OSError, PermissionError):
            # Skip directories we cannot read
            continue

    return max_mtime


def search_tags(pattern: str, scope: list[str] | None = None, cache_ttl: int = 60) -> list[dict]:
    """Search TAG markers using an in-memory cache.

    Args:
        pattern: Regex pattern (for example ``'@SPEC:AUTH-.*'``).
        scope: List of directories to scan. Defaults to specs/src/tests.
        cache_ttl: Cache time-to-live in seconds (default 60).

    Returns:
        List of matches such as ``{"file": "path", "line": 10, "tag": "...", "content": "..."}``.

    Notes:
        - Cache integrity relies on directory mtimes plus TTL.
        - Cache hits avoid spawning ``rg`` while misses shell out (≈13ms).
        - Uses ``rg --json`` output for structured parsing.
    """
    if scope is None:
        scope = [".moai/specs/", "src/", "tests/"]

    cache_key = f"{pattern}:{':'.join(scope)}"

    current_mtime = _get_dir_mtime_hash(scope)
    cache_entry = _tag_cache.get(cache_key)

    # Serve cached results when still valid
    if cache_entry:
        cached_results, cached_mtime, cached_at = cache_entry
        ttl_valid = time.time() - cached_at < cache_ttl

        # Matching mtime and a fresh TTL means we can reuse the cache
        if current_mtime == cached_mtime and ttl_valid:
            return cached_results

    # Cache miss → invoke ripgrep
    cmd = ["rg", pattern, "--json"] + scope

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=5, check=False)
    except (subprocess.TimeoutExpired, FileNotFoundError):
        # Missing rg or a timeout returns an empty list
        return []

    matches = []
    for line in result.stdout.strip().split("\n"):
        if not line:
            continue
        try:
            data = json.loads(line)
            if data.get("type") == "match":
                matches.append(
                    {
                        "file": data["data"]["path"]["text"],
                        "line": data["data"]["line_number"],
                        "tag": data["data"]["lines"]["text"].strip(),
                        "content": data["data"]["lines"]["text"],
                    }
                )
        except (json.JSONDecodeError, KeyError):
            # Ignore malformed JSON lines
            continue

    # Persist results with the current mtime snapshot
    _tag_cache[cache_key] = (matches, _get_dir_mtime_hash(scope), time.time())

    return matches


def verify_tag_chain(spec_id: str) -> dict:
    """Verify a TAG chain across ``@SPEC`` → ``@TEST`` → ``@CODE``.

    Args:
        spec_id: SPEC identifier (for example ``"AUTH-001"``).

    Returns:
        Dictionary with keys ``complete``, ``spec``, ``test``, ``code`` and ``orphans``.

    Notes:
        - Orphans capture TAGs found in code/tests without a SPEC.
        - A chain is complete only when all three categories contain matches.
        - Relies on ``search_tags`` so cached scans remain inexpensive.
    """
    chain = {
        "spec": search_tags(f"@SPEC:{spec_id}", [".moai/specs/"]),
        "test": search_tags(f"@TEST:{spec_id}", ["tests/"]),
        "code": search_tags(f"@CODE:{spec_id}", ["src/"]),
    }

    orphans = []
    if chain["code"] and not chain["spec"]:
        orphans.extend(chain["code"])
    if chain["test"] and not chain["spec"]:
        orphans.extend(chain["test"])

    return {
        "complete": bool(chain["spec"] and chain["test"] and chain["code"]),
        **chain,
        "orphans": orphans,
    }


def find_all_tags_by_type(tag_type: str = "SPEC") -> dict:
    """Return TAG IDs grouped by domain for the requested type.

    Args:
        tag_type: TAG category (``SPEC``, ``TEST``, ``CODE`` or ``DOC``).

    Returns:
        Dictionary like ``{"AUTH": ["AUTH-001", "AUTH-002"], ...}``.

    Notes:
        - Domain is derived from the ``FOO-123`` prefix.
        - Deduplicates repeated matches within the same domain.
        - Fetches data via ``search_tags`` so caching still applies.
    """
    tags = search_tags(f"@{tag_type}:([A-Z]+-[0-9]{{3}})")

    by_domain = {}
    for tag in tags:
        # @SPEC:AUTH-001 → AUTH
        try:
            tag_id = tag["tag"].split(":")[1]
            domain = "-".join(tag_id.split("-")[:-1])

            if domain not in by_domain:
                by_domain[domain] = []
            if tag_id not in by_domain[domain]:
                by_domain[domain].append(tag_id)
        except IndexError:
            # 파싱 실패 무시
            continue

    return by_domain


def suggest_tag_reuse(keyword: str, tag_type: str = "SPEC") -> list[str]:
    """Suggest existing TAG IDs that match the supplied keyword.

    Args:
        keyword: Keyword used to match domain names (case-insensitive).
        tag_type: TAG category (defaults to ``SPEC``).

    Returns:
        A list of up to five suggested TAG IDs.

    Notes:
        - Encourages reuse to avoid creating duplicate TAGs.
        - Performs a simple substring match against domain names.
    """
    all_tags = find_all_tags_by_type(tag_type)
    suggestions = []

    keyword_lower = keyword.lower()
    for domain, tag_ids in all_tags.items():
        if keyword_lower in domain.lower():
            suggestions.extend(tag_ids)

    return suggestions[:5]  # Cap results at five


def get_library_version(lib_name: str, cache_ttl: int = 86400) -> str | None:
    """Get the cached latest stable version for a library.

    Args:
        lib_name: Package name (for example ``"fastapi"``).
        cache_ttl: Cache TTL in seconds (defaults to 24 hours).

    Returns:
        Cached version string or ``None`` when the cache is cold.

    Notes:
        - Cache hits skip costly web searches (saves 3–5 seconds).
        - Agents should call ``set_library_version`` after fetching live data.
    """
    # Serve cached value when still within TTL
    if lib_name in _lib_version_cache:
        cached_version, cached_time = _lib_version_cache[lib_name]
        if time.time() - cached_time < cache_ttl:
            return cached_version

    # Cache miss → agent needs to perform the web search
    return None


def set_library_version(lib_name: str, version: str):
    """Persist a library version in the cache."""
    _lib_version_cache[lib_name] = (version, time.time())


__all__ = [
    "search_tags",
    "verify_tag_chain",
    "find_all_tags_by_type",
    "suggest_tag_reuse",
    "get_library_version",
    "set_library_version",
]
