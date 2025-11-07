#!/usr/bin/env python3
# @CODE:LANGUAGE-DIRS-001 | @SPEC:TAG-LANGUAGE-DETECTION-001 | @DOC:LANGUAGE-DIRS-CONFIG-001
"""Language-specific code directory detection and configuration.

Automatically detects expected code directories for a project based on its language,
merges with user-defined settings, and returns the final directory patterns.

Supported Languages (10):
  - Python, JavaScript, TypeScript
  - Go, Rust
  - Kotlin, Ruby, PHP
  - Java, C#

Features:
  - Auto/manual/hybrid detection modes
  - Custom pattern override support
  - Exclude pattern management
  - File extension to language detection
"""

from typing import Dict, List, Optional, Set
from pathlib import Path


# Language-specific code directory patterns
# Maps each language to its conventional code directory paths
LANGUAGE_DIRECTORY_MAP: Dict[str, List[str]] = {
    "python": [
        "src/",
        "lib/",
        "{package_name}/",  # Package name replaces with actual name
    ],
    "javascript": [
        "src/",
        "lib/",
        "app/",
        "pages/",
        "components/",
    ],
    "typescript": [
        "src/",
        "lib/",
        "app/",
        "pages/",
        "components/",
    ],
    "go": [
        "cmd/",
        "pkg/",
        "internal/",
    ],
    "rust": [
        "src/",
        "crates/",
    ],
    "kotlin": [
        "src/main/kotlin/",
        "src/test/kotlin/",
    ],
    "ruby": [
        "lib/",
        "app/",
    ],
    "php": [
        "src/",
        "app/",
    ],
    "java": [
        "src/main/java/",
        "src/test/java/",
    ],
    "csharp": [
        "src/",
        "App/",
    ],
}

# Common exclude patterns for all languages
# These directories are typically not part of source code
COMMON_EXCLUDE_PATTERNS: List[str] = [
    "tests/",
    "test/",
    "__tests__/",
    "spec/",
    "specs/",
    ".moai/",
    ".claude/",
    "node_modules/",
    "dist/",
    "build/",
    ".next/",
    ".nuxt/",
    "examples/",
    "docs/",
    "documentation/",
    "templates/",
    ".git/",
    ".github/",
    "venv/",
    ".venv/",
    "vendor/",
    "target/",
    "bin/",
    "__pycache__/",
    "*.egg-info/",
]


def detect_directories(
    config: Optional[Dict] = None,
    language: Optional[str] = None,
) -> List[str]:
    """Detect code directory patterns based on project configuration and language.

    Detection logic:
    1. Manual mode: If custom patterns exist → use custom patterns only
    2. Auto mode (default): Use language-specific default patterns
    3. Hybrid mode: Merge language defaults with custom patterns

    Args:
        config: Project configuration dict (e.g., loaded from .moai/config.json).
                Expected structure: {"project": {"language": "python"}, "tags": {...}}
        language: Project language code. Overrides config language if provided.

    Returns:
        List of detected code directory patterns.

    Raises:
        KeyError: If config is malformed (should not occur in normal use).

    Example:
        >>> config = {"project": {"language": "python"}}
        >>> dirs = detect_directories(config)
        >>> assert "src/" in dirs
    """
    config = config or {}

    # Determine language
    if language is None:
        language = config.get("project", {}).get("language", "python").lower()

    # Extract custom configuration
    tags_policy = config.get("tags", {}).get("policy", {})
    code_dirs_config = tags_policy.get("code_directories", {})

    detection_mode = code_dirs_config.get("detection_mode", "auto")
    custom_patterns = code_dirs_config.get("patterns", [])

    # Handle detection modes
    if detection_mode == "manual" and custom_patterns:
        # Manual mode: use custom patterns only
        return custom_patterns

    elif detection_mode == "auto":
        # Auto mode: use language-specific default patterns
        return LANGUAGE_DIRECTORY_MAP.get(language, LANGUAGE_DIRECTORY_MAP["python"])

    elif detection_mode == "hybrid":
        # Hybrid mode: merge language defaults with custom patterns
        base_patterns = LANGUAGE_DIRECTORY_MAP.get(language, LANGUAGE_DIRECTORY_MAP["python"])
        combined = list(set(base_patterns + custom_patterns))
        return sorted(combined)

    # Default: auto-detect
    return LANGUAGE_DIRECTORY_MAP.get(language, LANGUAGE_DIRECTORY_MAP["python"])


def get_exclude_patterns(
    config: Optional[Dict] = None,
) -> List[str]:
    """Get directory patterns to exclude from code detection.

    Exclusion logic:
    1. If custom patterns exist and merge is disabled → use custom only
    2. If custom patterns exist and merge is enabled → merge with common patterns
    3. Otherwise → use common exclude patterns

    Args:
        config: Project configuration dict. Expected structure:
                {"tags": {"policy": {"code_directories": {...}}}}

    Returns:
        List of directory patterns to exclude from code detection.

    Example:
        >>> config = {"tags": {"policy": {"code_directories": {"exclude_patterns": ["vendor/"]}}}}
        >>> patterns = get_exclude_patterns(config)
        >>> assert "tests/" in patterns  # Common patterns included
    """
    config = config or {}

    tags_policy = config.get("tags", {}).get("policy", {})
    code_dirs_config = tags_policy.get("code_directories", {})

    custom_exclude = code_dirs_config.get("exclude_patterns", [])
    merge_with_common = code_dirs_config.get("merge_exclude_patterns", True)

    if custom_exclude and not merge_with_common:
        # Use custom patterns only
        return custom_exclude

    if custom_exclude and merge_with_common:
        # Merge common and custom patterns
        return list(set(COMMON_EXCLUDE_PATTERNS + custom_exclude))

    # Default: common exclude patterns only
    return COMMON_EXCLUDE_PATTERNS


def is_code_directory(
    path: Path,
    config: Optional[Dict] = None,
    language: Optional[str] = None,
) -> bool:
    """Check if the given path is a code directory.

    Checks path against exclude patterns first (early return if excluded),
    then validates against code directory patterns.

    Args:
        path: File or directory path to check (Path object or string).
        config: Project configuration dict.
        language: Project language code.

    Returns:
        True if path matches code directory patterns and is not excluded, False otherwise.

    Example:
        >>> path = Path("src/auth/login.py")
        >>> is_code_directory(path, language="python")
        True
        >>> path = Path("tests/test_auth.py")
        >>> is_code_directory(path, language="python")
        False
    """
    code_dirs = detect_directories(config, language)
    exclude_patterns = get_exclude_patterns(config)

    path_str = str(path)

    # Check exclude patterns first (early return)
    for exclude in exclude_patterns:
        if exclude.endswith("/"):
            if path_str.startswith(exclude) or f"/{exclude}" in path_str:
                return False
        else:
            if exclude in path_str:
                return False

    # Check code directory patterns
    for code_dir in code_dirs:
        if code_dir.endswith("/"):
            if path_str.startswith(code_dir) or f"/{code_dir}" in path_str:
                return True
        else:
            if code_dir in path_str:
                return True

    return False


def get_language_by_file_extension(file_path: Path) -> Optional[str]:
    """Infer language from file extension.

    Args:
        file_path: File path (Path object or string).

    Returns:
        Inferred language code (e.g., "python", "javascript"), or None if not recognized.

    Example:
        >>> get_language_by_file_extension(Path("auth.py"))
        "python"
        >>> get_language_by_file_extension(Path("index.ts"))
        "typescript"
    """
    extension_map = {
        ".py": "python",
        ".js": "javascript",
        ".jsx": "javascript",
        ".ts": "typescript",
        ".tsx": "typescript",
        ".go": "go",
        ".rs": "rust",
        ".kt": "kotlin",
        ".kts": "kotlin",
        ".rb": "ruby",
        ".php": "php",
        ".java": "java",
        ".cs": "csharp",
    }

    suffix = file_path.suffix.lower()
    return extension_map.get(suffix)


def get_all_supported_languages() -> List[str]:
    """Get list of all supported languages.

    Returns:
        Sorted list of language codes currently supported.

    Example:
        >>> langs = get_all_supported_languages()
        >>> assert "python" in langs
        >>> assert len(langs) == 10
    """
    return sorted(LANGUAGE_DIRECTORY_MAP.keys())


def validate_language(language: str) -> bool:
    """Check if language is supported.

    Args:
        language: Language code to validate.

    Returns:
        True if language is supported, False otherwise.

    Example:
        >>> validate_language("python")
        True
        >>> validate_language("cobol")
        False
    """
    return language.lower() in LANGUAGE_DIRECTORY_MAP


def merge_directory_patterns(
    base_patterns: List[str],
    custom_patterns: List[str],
) -> List[str]:
    """Merge base and custom directory patterns.

    Removes duplicates and returns sorted result.

    Args:
        base_patterns: Base language-specific directory patterns.
        custom_patterns: User-defined custom directory patterns.

    Returns:
        Merged and deduplicated patterns, sorted alphabetically.

    Example:
        >>> base = ["src/", "lib/"]
        >>> custom = ["lib/", "app/"]
        >>> result = merge_directory_patterns(base, custom)
        >>> assert result == ["app/", "lib/", "src/"]
    """
    merged = list(set(base_patterns + custom_patterns))
    return sorted(merged)


def expand_package_placeholder(
    patterns: List[str],
    package_name: Optional[str] = None,
) -> List[str]:
    """Expand {package_name} placeholder in patterns with actual package name.

    If package_name is not provided, patterns containing {package_name} are filtered out.

    Args:
        patterns: Original directory patterns (may contain {package_name}).
        package_name: Actual package name for substitution. If None, placeholders are removed.

    Returns:
        Expanded patterns with placeholders replaced or filtered.

    Example:
        >>> patterns = ["src/", "{package_name}/", "lib/"]
        >>> expand_package_placeholder(patterns, "myproject")
        ["src/", "myproject/", "lib/"]
        >>> expand_package_placeholder(patterns, None)
        ["src/", "lib/"]
    """
    if not package_name:
        return [p for p in patterns if "{package_name}" not in p]

    expanded = []
    for pattern in patterns:
        if "{package_name}" in pattern:
            expanded.append(pattern.replace("{package_name}", package_name))
        else:
            expanded.append(pattern)

    return expanded
