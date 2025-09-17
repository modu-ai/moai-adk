#!/usr/bin/env python3
"""
MoAI-ADK Context Selector (Top-K recommender)

PreToolUse 훅에서 가볍게 실행되어, 현재 프로젝트의 .moai/memory/에서
핵심 컨텍스트 3~5개만 추천(@ 경로 형태)으로 표준 출력에 제공한다.

설계 제약
- 한 함수 ≤ 50 LOC, 파일 ≤ 300 LOC
- 외부 의존성 없음, I/O는 읽기 전용
"""

from pathlib import Path
from typing import List
import os


def find_project_root() -> Path:
    """프로젝트 루트를 추정한다.
    우선순위: ENV(CLAUDE_PROJECT_DIR) > 현재 디렉토리 상향 탐색(.moai/.claude).
    """
    env_dir = os.getenv("CLAUDE_PROJECT_DIR")
    if env_dir:
        return Path(env_dir)

    cur = Path.cwd()
    for _ in range(10):
        if (cur / ".moai").exists() or (cur / ".claude").exists():
            return cur
        if cur.parent == cur:
            break
        cur = cur.parent
    return Path.cwd()


def suggest_memory_files(project_root: Path, k: int = 5) -> List[Path]:
    """.moai/memory/에서 우선순위 기반으로 최대 k개 추천한다."""
    mem_dir = project_root / ".moai" / "memory"
    if not mem_dir.exists() or not mem_dir.is_dir():
        return []

    candidates: List[Path] = []

    # 1) 최우선 공통 문서
    for name in ("common.md", "constitution.md"):
        p = mem_dir / name
        if p.exists():
            candidates.append(p)

    # 2) 스택별 문서(backend-*, frontend-*) 중 최신 수정 1~2개
    stack_files = list(mem_dir.glob("backend-*.md")) + list(mem_dir.glob("frontend-*.md"))
    stack_files.sort(key=lambda x: x.stat().st_mtime, reverse=True)
    for p in stack_files[:2]:
        if p not in candidates:
            candidates.append(p)

    # 3) 그 외 최근 수정 문서에서 보충
    others = [p for p in mem_dir.glob("*.md") if p not in candidates]
    others.sort(key=lambda x: x.stat().st_mtime, reverse=True)
    for p in others:
        if len(candidates) >= k:
            break
        candidates.append(p)

    return candidates[:k]


def main() -> None:
    root = find_project_root()
    suggestions = suggest_memory_files(root, k=5)
    if not suggestions:
        return

    print("🔎 Context suggestions (Top-K from .moai/memory):")
    for p in suggestions:
        try:
            rel = p.relative_to(root)
        except ValueError:
            rel = p
        # @ 경로 표기 출력
        print(f"  - @{rel}")


if __name__ == "__main__":
    main()

