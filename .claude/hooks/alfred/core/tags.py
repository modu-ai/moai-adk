#!/usr/bin/env python3
# @CODE:HOOK-TAG-001 | SPEC: TBD | TEST: tests/hooks/test_tag_validation.py
"""TAG validation helpers for MoAI-ADK hooks

Fast checks used by PreToolUse/PostToolUse to nudge users when
new or modified files are missing required @TAG annotations.

Configurable rules with sensible defaults:
- Load patterns from .moai/tag-rules.json if present.
- Otherwise, apply default glob patterns (folder names are not hard-coded only).

Defaults (order matters; first match wins):
1) SPEC
   - .moai/specs/**
   - **/SPEC-*/spec.md
2) TEST
   - **/*_test.py, **/test_*.py, **/*.test.* (ts,tsx,js,jsx,go,rs)
   - **/*.spec.* (ts,tsx,js,jsx)
   - tests/**
3) DOC
   - docs/**/*.md, **/README.md, **/*.api.md
4) CODE
   - Source extensions: .py,.ts,.tsx,.js,.jsx,.go,.rs,.java,.kt,.rb,.php,.c,.cpp,.cs,.swift,.scala
   - Excluding TEST patterns

Notes:
- Best-effort: skip binary/large files and non-target paths
- Do not block execution; return a list of issues for messaging
"""

from __future__ import annotations

import fnmatch
import json
import subprocess
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable, List, Optional

DEFAULT_CODE_EXTS = (
    ".py", ".ts", ".tsx", ".js", ".jsx", ".go", ".rs",
    ".java", ".kt", ".rb", ".php", ".c", ".cpp", ".cs",
    ".swift", ".scala"
)


@dataclass
class TagIssue:
    path: str
    expected: str  # one of @SPEC, @TEST, @CODE, @DOC
    reason: str


@dataclass
class Rule:
    include: List[str]
    expect: str  # '@SPEC:' | '@TEST:' | '@CODE:' | '@DOC:'
    exclude: List[str]


def _load_rules(cwd: str) -> List[Rule]:
    """Load tag rules from .moai/tag-rules.json or return defaults.

    Schema example:
    {
      "rules": [
        {"include": ["**/*_test.py", "**/*.test.ts"], "expect": "@TEST:", "exclude": []},
        {"include": ["docs/**/*.md", "**/README.md"], "expect": "@DOC:", "exclude": []}
      ]
    }
    """
    cfg = Path(cwd) / ".moai" / "tag-rules.json"
    if cfg.exists():
        try:
            data = json.loads(cfg.read_text(encoding="utf-8"))
            items = data.get("rules", [])
            rules: List[Rule] = []
            for it in items:
                include = list(it.get("include", []))
                expect = str(it.get("expect", ""))
                exclude = list(it.get("exclude", []))
                if include and expect in ("@SPEC:", "@TEST:", "@CODE:", "@DOC:"):
                    rules.append(Rule(include=include, expect=expect, exclude=exclude))
            if rules:
                return rules
        except Exception:
            pass

    # Defaults (ordered)
    return [
        Rule(
            include=[".moai/specs/**", "**/SPEC-*/spec.md"],
            expect="@SPEC:",
            exclude=[]
        ),
        Rule(
            include=[
                "**/*_test.py", "**/test_*.py", "**/*.test.ts",
                "**/*.test.tsx", "**/*.test.js", "**/*.test.jsx",
                "**/*.test.go", "**/*.test.rs", "**/*.spec.ts",
                "**/*.spec.tsx", "tests/**"
            ],
            expect="@TEST:",
            exclude=[".claude/**"]
        ),
        Rule(
            include=["docs/**/*.md", "**/README.md", "**/*.api.md"],
            expect="@DOC:",
            exclude=[".claude/**"]
        ),
        Rule(
            include=["**/*"],
            expect="@CODE:",
            exclude=[
                "tests/**", "docs/**", ".moai/**", ".claude/**",
                "**/*.md", "**/*.json", "**/*.yml", "**/*.yaml",
                "**/*.toml", "**/*.lock", "**/*.svg", "**/*.png",
                "**/*.jpg", "**/*.jpeg", "**/*.gif"
            ]
        ),
    ]


def _match_any(path: str, patterns: List[str]) -> bool:
    return any(fnmatch.fnmatch(path, pat) for pat in patterns)


def _needs_tag_str(path_str: str, rules: List[Rule]) -> Optional[str]:
    p = path_str
    for rule in rules:
        if _match_any(p, rule.include) and not _match_any(p, rule.exclude):
            if rule.expect == "@CODE:":
                # CODE: limit to source-like extensions to reduce noise
                if not any(p.endswith(ext) for ext in DEFAULT_CODE_EXTS):
                    continue
            return rule.expect
    return None


def _has_tag(content: str, expected: str) -> bool:
    return expected in content


def _iter_recent_changes(cwd: str) -> Iterable[Path]:
    root = Path(cwd)
    try:
        # Staged files
        r1 = subprocess.run(
            ["git", "diff", "--name-only", "--cached"],
            cwd=cwd, capture_output=True, text=True, timeout=1
        )
        # Modified (unstaged) tracked files
        r2 = subprocess.run(
            ["git", "ls-files", "-m"],
            cwd=cwd, capture_output=True, text=True, timeout=1
        )
        # Untracked (other) files respecting .gitignore
        r3 = subprocess.run(
            ["git", "ls-files", "-o", "--exclude-standard"],
            cwd=cwd, capture_output=True, text=True, timeout=1
        )
        names = set()
        if r1.returncode == 0:
            names.update([line.strip() for line in r1.stdout.splitlines() if line.strip()])
        if r2.returncode == 0:
            names.update([line.strip() for line in r2.stdout.splitlines() if line.strip()])
        if r3.returncode == 0:
            names.update([line.strip() for line in r3.stdout.splitlines() if line.strip()])
        for n in names:
            p = (root / n).resolve()
            if p.is_file():
                yield p
    except Exception:
        return []


def scan_recent_changes_for_missing_tags(cwd: str) -> list[TagIssue]:
    issues: list[TagIssue] = []
    rules = _load_rules(cwd)
    root = Path(cwd).resolve()
    for path in _iter_recent_changes(cwd):
        try:
            content = path.read_text(encoding="utf-8", errors="ignore")
        except Exception:
            continue
        # compute relative path once and use for matching/excluding
        try:
            rel = path.resolve().relative_to(root)
            rel_s = rel.as_posix()
        except Exception:
            rel_s = path.name

        expected = _needs_tag_str(rel_s, rules)
        if not expected:
            continue
        if not _has_tag(content, expected):
            issues.append(TagIssue(path=rel_s, expected=expected, reason="missing tag"))
    return issues
