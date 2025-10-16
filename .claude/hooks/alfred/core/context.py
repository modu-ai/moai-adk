#!/usr/bin/env python3
"""Context Engineering utilities

JIT (Just-in-Time) Retrieval 및 워크플로우 컨텍스트 관리
"""

import time
from pathlib import Path
from typing import Any

# Workflow context shared across phases
_workflow_context: dict[str, Any] = {}


def get_jit_context(prompt: str, cwd: str) -> list[str]:
    """프롬프트 기반 JIT Context Retrieval

    사용자 프롬프트를 분석하여 관련 문서를 자동으로 추천합니다.
    Alfred 커맨드, 키워드 기반 패턴 매칭으로 필요한 문서만 로드합니다.

    Args:
        prompt: 사용자 입력 프롬프트 (대소문자 무관)
        cwd: 프로젝트 루트 디렉토리 경로

    Returns:
        추천 문서 경로 리스트 (상대 경로).
        매칭되는 패턴이 없거나 파일이 없으면 빈 리스트 []

    Patterns:
        - "/alfred:1-spec" → .moai/memory/spec-metadata.md
        - "/alfred:2-build" → .moai/memory/development-guide.md
        - "test" → tests/ (디렉토리가 존재하는 경우)

    Examples:
        >>> get_jit_context("/alfred:1-spec", "/project")
        ['.moai/memory/spec-metadata.md']
        >>> get_jit_context("implement test", "/project")
        ['tests/']
        >>> get_jit_context("unknown", "/project")
        []

    Notes:
        - Context Engineering: JIT Retrieval 원칙 준수
        - 필요한 문서만 로드하여 초기 컨텍스트 부담 최소화
        - 파일 존재 여부 확인 후 반환

    TDD History:
        - RED: 18개 시나리오 테스트 (커맨드 매칭, 키워드, 빈 결과)
        - GREEN: 패턴 매칭 딕셔너리 기반 구현
        - REFACTOR: 확장 가능한 패턴 구조, 파일 존재 검증 추가
    """
    context_files = []
    cwd_path = Path(cwd)

    # Pattern matching
    patterns = {
        "/alfred:1-spec": [".moai/memory/spec-metadata.md"],
        "/alfred:2-build": [".moai/memory/development-guide.md"],
        "test": ["tests/"],
    }

    for pattern, files in patterns.items():
        if pattern in prompt.lower():
            for file in files:
                file_path = cwd_path / file
                if file_path.exists():
                    context_files.append(file)

    return context_files


def save_phase_context(phase: str, data: dict):
    """Store per-phase workflow context data.

    Args:
        phase: ``"analysis"``, ``"implementation"`` or ``"verification"``.
        data: Payload to share with later phases.

    Notes:
        - Enables reuse of expensive analysis across phases.
        - Entries expire after 10 minutes to avoid stale data.
    """
    _workflow_context[phase] = {"data": data, "timestamp": time.time()}


def load_phase_context(phase: str) -> dict | None:
    """Load previously stored workflow context for a phase.

    Returns:
        Stored context data or ``None`` when expired or missing.
    """
    if phase in _workflow_context:
        ctx = _workflow_context[phase]
        # Only accept entries that are younger than 10 minutes
        if time.time() - ctx["timestamp"] < 600:
            return ctx["data"]
    return None


def clear_workflow_context():
    """Clear all cached workflow context (call at workflow end)."""
    _workflow_context.clear()


__all__ = [
    "get_jit_context",
    "save_phase_context",
    "load_phase_context",
    "clear_workflow_context",
]
