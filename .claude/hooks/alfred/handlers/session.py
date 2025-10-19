#!/usr/bin/env python3
"""Session event handlers

SessionStart, SessionEnd 이벤트 처리
"""

import json
from pathlib import Path

from core import HookPayload, HookResult
from core.checkpoint import list_checkpoints
from core.project import count_specs, detect_language, get_git_info


def _get_locale(cwd: str) -> str:
    """프로젝트 locale 감지

    .moai/config.json의 project.locale 또는 CLAUDE.md 파싱하여 locale을 추출합니다.

    Args:
        cwd: 프로젝트 루트 디렉토리 경로

    Returns:
        locale 문자열 ('ko', 'en' 등). 기본값 'en'

    Examples:
        >>> _get_locale(".")
        'ko'

    Notes:
        - 우선순위: .moai/config.json > CLAUDE.md
        - 파싱 실패 시 기본값 'en' 반환
    """
    # 1. .moai/config.json에서 읽기
    config_path = Path(cwd) / ".moai" / "config.json"
    if config_path.exists():
        try:
            config = json.loads(config_path.read_text())
            locale = config.get("project", {}).get("locale")
            if locale:
                return locale
        except (OSError, json.JSONDecodeError):
            pass

    # 2. CLAUDE.md에서 읽기 (fallback)
    claude_md_path = Path(cwd) / ".moai" / "CLAUDE.md"
    if claude_md_path.exists():
        try:
            content = claude_md_path.read_text()
            # locale: ko 형식 찾기
            for line in content.splitlines():
                if "locale:" in line.lower():
                    locale = line.split(":")[-1].strip()
                    if locale:
                        return locale
        except (OSError, UnicodeDecodeError):
            pass

    return "en"  # 기본값


def _get_labels(locale: str) -> dict[str, str]:
    """다국어 라벨 반환

    Args:
        locale: 언어 코드 ('ko', 'en')

    Returns:
        라벨 딕셔너리

    Examples:
        >>> _get_labels('ko')
        {'based': '기반 프로젝트', 'specs': 'SPEC 진행도', ...}
    """
    labels = {
        "ko": {
            "based": "기반 프로젝트",
            "specs": "SPEC 진행도",
            "changes": "미커밋 변경사항",
            "branch": "브랜치",
            "checkpoints": "사용 가능한 체크포인트",
            "files": "개",
        },
        "en": {
            "based": "Based project",
            "specs": "SPEC progress",
            "changes": "Uncommitted changes",
            "branch": "Branch",
            "checkpoints": "Available checkpoints",
            "files": " files",  # 영어는 공백 포함
        },
    }

    return labels.get(locale, labels["en"])


def handle_session_start(payload: HookPayload) -> HookResult:
    """SessionStart 이벤트 핸들러 (프로젝트 상태 요약)

    Claude Code 세션 시작 시 프로젝트 상태 정보를 표시합니다.
    5가지 핵심 정보를 간결하게 제공하여 사용자가 현재 상태를 빠르게 파악하도록 돕습니다.

    Args:
        payload: Claude Code 이벤트 페이로드 (cwd 키 필수)

    Returns:
        HookResult(message=상태 요약, systemMessage=사용자 표시용)

    Message Format (한국어 예시):
        🚀 MoAI-ADK Session Started
           🐍 Python 기반 프로젝트
           📊 SPEC 진행도: 28/31 (90%)
           🔄 미커밋 변경사항: 2개
           📍 브랜치: feature/update-0.4.0
           💾 사용 가능한 체크포인트: 3개 (before-delete-..., before-merge-..., ...)

    Message Format (English):
        🚀 MoAI-ADK Session Started
           🐍 Python Based project
           📊 SPEC progress: 28/31 (90%)
           🔄 Uncommitted changes: 2 files
           📍 Branch: feature/update-0.4.0
           💾 Available checkpoints: 3 (before-delete-..., before-merge-..., ...)

    Note:
        - Claude Code는 SessionStart를 여러 단계로 처리 (clear → compact)
        - 중복 출력 방지를 위해 "compact" 단계에서만 메시지 표시
        - "clear" 단계는 빈 결과 반환 (사용자에게 보이지 않음)
        - 다국어 지원: .moai/config.json의 project.locale 기준

    Performance:
        - 언어 감지: detect_language() < 10ms
        - SPEC 카운트: count_specs() < 50ms
        - Git 정보: get_git_info() < 100ms
        - Checkpoint 목록: list_checkpoints() < 20ms
        - 총 예상 시간: < 200ms

    TDD History:
        - RED: 세션 시작 메시지 형식 테스트
        - GREEN: helper 함수 조합하여 상태 메시지 생성
        - REFACTOR: 메시지 포맷 개선, 가독성 향상, checkpoint 목록 추가
        - FIX: clear 단계 중복 출력 방지 (compact 단계만 표시)
        - SIMPLIFY: 복잡한 상태 정보 제거, 3줄 간단 인사로 변경
        - ENHANCE: 5가지 핵심 정보 추가, 다국어 지원 (ko/en)

    @TAG:CHECKPOINT-EVENT-001
    """
    # Claude Code SessionStart는 여러 단계로 실행됨 (clear, compact 등)
    # "clear" 단계는 무시하고 "compact" 단계에서만 메시지 출력
    event_phase = payload.get("phase", "")
    if event_phase == "clear":
        return HookResult()  # 빈 결과 반환 (중복 출력 방지)

    cwd = payload.get("cwd", ".")

    # 1. Locale 감지 (다국어 지원)
    locale = _get_locale(cwd)
    labels = _get_labels(locale)

    # 2. 언어 감지
    language = detect_language(cwd)
    language_emoji = {
        "python": "🐍",
        "typescript": "📘",
        "javascript": "📜",
        "java": "☕",
        "go": "🐹",
        "rust": "🦀",
        "dart": "🎯",
        "swift": "🍎",
        "kotlin": "🅺",
        "ruby": "💎",
    }.get(language.lower(), "📦")

    language_display = language.capitalize() if language != "Unknown Language" else "Unknown"

    # 3. SPEC 진행도
    spec_info = count_specs(cwd)
    spec_text = f"{spec_info['completed']}/{spec_info['total']} ({spec_info['percentage']}%)"

    # 4. Git 정보 (브랜치, 미커밋 변경사항)
    git_info = get_git_info(cwd)
    branch = git_info.get("branch", "N/A")
    changes = git_info.get("changes", 0)

    # 5. Checkpoint 목록 (최근 3개)
    checkpoints = list_checkpoints(cwd, max_count=3)
    checkpoint_count = len(checkpoints)
    if checkpoint_count > 0:
        checkpoint_names = ", ".join([cp["branch"] for cp in checkpoints])
        checkpoint_text = f"{checkpoint_count} ({checkpoint_names})"
    else:
        checkpoint_text = "0"

    # 메시지 조립
    system_message = (
        "🚀 MoAI-ADK Session Started\n"
        f"   {language_emoji} {language_display} {labels['based']}\n"
        f"   📊 {labels['specs']}: {spec_text}\n"
        f"   🔄 {labels['changes']}: {changes}{labels['files']}\n"
        f"   📍 {labels['branch']}: {branch}\n"
        f"   💾 {labels['checkpoints']}: {checkpoint_text}"
    )

    return HookResult(
        message=system_message,  # Claude 컨텍스트용
        systemMessage=system_message,  # 사용자 표시용
    )


def handle_session_end(payload: HookPayload) -> HookResult:
    """SessionEnd 이벤트 핸들러 (기본 구현)"""
    return HookResult()


__all__ = ["handle_session_start", "handle_session_end"]
