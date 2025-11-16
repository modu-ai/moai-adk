#!/usr/bin/env python3
""".gitignore parser utility

Parse .gitignore files to extract patterns for exclusion.

Features:
- Parse .gitignore files
- Exclude comments and empty lines
- Normalize glob patterns
- Handle directory patterns

Usage:
    from .utils.gitignore_parser import load_gitignore_patterns
    patterns = load_gitignore_patterns()
"""

from pathlib import Path
from typing import List, Set


def load_gitignore_patterns(gitignore_path: str = ".gitignore") -> List[str]:
    """Load patterns from .gitignore

    Args:
        gitignore_path: Path to .gitignore file (default: ".gitignore")

    Returns:
        List of exclude patterns
    """
    patterns = []

    try:
        gitignore_file = Path(gitignore_path)
        if not gitignore_file.exists():
            return patterns

        with open(gitignore_file, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()

                # Exclude empty lines or comments
                if not line or line.startswith('#'):
                    continue

                # Exclude negation patterns (!) - not used in validation
                if line.startswith('!'):
                    continue

                # Normalize pattern
                pattern = normalize_pattern(line)
                if pattern:
                    patterns.append(pattern)

    except Exception:
        # Return empty list on file read failure
        pass

    return patterns


def normalize_pattern(pattern: str) -> str:
    """Normalize pattern

    Args:
        pattern: Original pattern

    Returns:
        Normalized pattern
    """
    # Remove leading and trailing whitespace
    pattern = pattern.strip()

    # Absolute path pattern (starts with /)
    if pattern.startswith('/'):
        # Remove / and return
        return pattern[1:]

    # Return regular pattern as-is (including directory slashes)
    return pattern


def is_path_ignored(
    file_path: str,
    ignore_patterns: List[str],
    protected_paths: List[str] = None
) -> bool:
    """Check if file path matches ignore patterns

    Args:
        file_path: File path to check
        ignore_patterns: List of ignore patterns
        protected_paths: Paths that should be protected (not ignored)

    Returns:
        True if matched
    """
    import fnmatch

    # Set default protected paths
    if protected_paths is None:
        protected_paths = [".moai/specs/"]

    # Do not ignore if protected path
    for protected in protected_paths:
        if file_path.startswith(protected):
            return False

    # Normalize path
    path_parts = Path(file_path).parts

    for pattern in ignore_patterns:
        # Wildcard directory patterns (hooks_backup_*/, *_backup_*/)
        if '*' in pattern and pattern.endswith('/'):
            pattern_without_slash = pattern[:-1]
            # Match with each path part
            for part in path_parts:
                if fnmatch.fnmatch(part, pattern_without_slash):
                    return True

        # Wildcard pattern matching (*.ext, *backup*)
        elif '*' in pattern:
            # Match full path
            if fnmatch.fnmatch(file_path, pattern):
                return True
            # Match each path part
            for part in path_parts:
                if fnmatch.fnmatch(part, pattern):
                    return True

        # Directory pattern matching
        elif pattern.endswith('/'):
            dir_name = pattern[:-1]
            if dir_name in path_parts:
                return True

        # Simple string matching (filename or directory name)
        else:
            if pattern in file_path:
                return True

    return False


def get_combined_exclude_patterns(
    base_patterns: List[str],
    gitignore_path: str = ".gitignore"
) -> List[str]:
    """Combine base patterns with .gitignore patterns

    Args:
        base_patterns: Base exclude patterns
        gitignore_path: Path to .gitignore file

    Returns:
        Combined exclude pattern list (deduplicated)
    """
    # Start with base patterns
    patterns_set: Set[str] = set(base_patterns)

    # Add .gitignore patterns
    gitignore_patterns = load_gitignore_patterns(gitignore_path)
    patterns_set.update(gitignore_patterns)

    # Sort and return
    return sorted(list(patterns_set))


if __name__ == "__main__":
    # Test code
    patterns = load_gitignore_patterns()
    print(f"Loaded {len(patterns)} patterns from .gitignore")
    for pattern in patterns[:10]:
        print(f"  - {pattern}")
