#!/usr/bin/env python3
"""Context Engineering utilities

JIT (Just-in-Time) Retrieval
"""

from pathlib import Path


def get_jit_context(prompt: str, cwd: str) -> list[str]:
    """JIT Context Retrieval based on prompt.

    Analyze user prompts and automatically recommend relevant documents.
    Alfred commands and keyword-based pattern matching load only the documents you need.

    Args:
        prompt: Prompt for user input (case is irrelevant)
        cwd: Project root directory path

    Returns:
        List of recommended document paths (relative paths).
        If there is no matching pattern or file, an empty list []

    Patterns:
        - "/alfred:1-plan" → .claude/skills/moai-alfred-spec-metadata-extended/reference.md
        - "/alfred:2-run" → .claude/skills/moai-alfred-dev-guide/reference.md
        - "test" → tests/ (if directory exists)

    Examples:
        >>> get_jit_context("/alfred:1-plan", "/project")
        ['.claude/skills/moai-alfred-spec-metadata-extended/reference.md']
        >>> get_jit_context("implement test", "/project")
        ['tests/']
        >>> get_jit_context("unknown", "/project")
        []

    Notes:
        - Context Engineering: Compliance with JIT Retrieval principles
        - Minimize initial context burden by loading only necessary documents
        - Return after checking whether file exists

    TDD History:
        - RED: 18 items scenario testing (command matching, keywords, empty results)
        - GREEN: Pattern matching dictionary-based implementation
        - REFACTOR: Expandable pattern structure, file existence validation added
    """
    context_files = []
    cwd_path = Path(cwd)

    # Pattern matching
    patterns = {
        "/alfred:1-plan": [".claude/skills/moai-alfred-spec-metadata-extended/reference.md"],
        "/alfred:2-run": [".claude/skills/moai-alfred-dev-guide/reference.md"],
        "test": ["tests/"],
    }

    for pattern, files in patterns.items():
        if pattern in prompt.lower():
            for file in files:
                file_path = cwd_path / file
                if file_path.exists():
                    context_files.append(file)

    return context_files


__all__ = ["get_jit_context"]
