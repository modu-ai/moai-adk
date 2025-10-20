#!/usr/bin/env python3
"""Session event handlers

SessionStart, SessionEnd 이벤트 처리
"""

from core import HookPayload, HookResult
from core.checkpoint import list_checkpoints
from core.project import count_specs, detect_language, get_git_info, get_project_locale
from core.i18n import t


def handle_session_start(payload: HookPayload) -> HookResult:
    """SessionStart 이벤트 핸들러 (Checkpoint 목록 포함)

    Claude Code 세션 시작 시 프로젝트 상태를 요약하여 표시합니다.
    언어, Git 상태, SPEC 진행도, Checkpoint 목록을 한눈에 확인할 수 있습니다.

    Args:
        payload: Claude Code 이벤트 페이로드 (cwd 키 필수)

    Returns:
        HookResult(message=프로젝트 상태 요약 메시지, systemMessage=사용자 표시용)

    Message Format:
        🚀 MoAI-ADK Session Started
           Language: {언어}
           Branch: {브랜치} ({커밋 해시})
           Changes: {변경 파일 수}
           SPEC Progress: {완료}/{전체} ({퍼센트}%)
           Checkpoints: {개수} available (최신 3개 표시)

    Note:
        - Claude Code는 SessionStart를 여러 단계로 처리 (clear → compact)
        - 중복 출력 방지를 위해 "compact" 단계에서만 메시지 표시
        - "clear" 단계는 빈 결과 반환 (사용자에게 보이지 않음)

    TDD History:
        - RED: 세션 시작 메시지 형식 테스트
        - GREEN: helper 함수 조합하여 상태 메시지 생성
        - REFACTOR: 메시지 포맷 개선, 가독성 향상, checkpoint 목록 추가
        - FIX: clear 단계 중복 출력 방지 (compact 단계만 표시)

    @TAG:CHECKPOINT-EVENT-001
    """
    # Claude Code SessionStart는 여러 단계로 실행됨 (clear, compact 등)
    # "clear" 단계는 무시하고 "compact" 단계에서만 메시지 출력
    event_phase = payload.get("phase", "")
    if event_phase == "clear":
        return HookResult()  # 빈 결과 반환 (중복 출력 방지)

    cwd = payload.get("cwd", ".")

    # Get project locale for i18n messages
    # i18n 메시지용 프로젝트 locale 읽기
    locale = get_project_locale(cwd)

    # Collect project metadata
    # 프로젝트 메타데이터 수집
    language = detect_language(cwd)
    git_info = get_git_info(cwd)
    specs = count_specs(cwd)
    checkpoints = list_checkpoints(cwd, max_count=10)

    branch = git_info.get("branch", "N/A")
    commit = git_info.get("commit", "N/A")[:7]
    changes = git_info.get("changes", 0)
    spec_progress = f"{specs['completed']}/{specs['total']}"

    # systemMessage: Localized messages for the user
    # systemMessage: 사용자 locale 기반 메시지
    lines = [
        t("session_start", locale),
        f"   {t('language', locale)}: {language}",
        f"   {t('branch', locale)}: {branch} ({commit})",
        f"   {t('changes', locale)}: {changes}",
        f"   {t('spec_progress', locale)}: {spec_progress} ({specs['percentage']}%)",
    ]

    # Checkpoint list (show latest 3)
    # Checkpoint 목록 추가 (최신 3개만 표시)
    if checkpoints:
        lines.append(f"   {t('checkpoints', locale)}: {len(checkpoints)} available")
        for cp in reversed(checkpoints[-3:]):  # 최신 3개
            branch_short = cp["branch"].replace("before-", "")
            lines.append(f"      - {branch_short}")
        lines.append(f"   {t('restore_hint', locale)}")

    system_message = "\n".join(lines)

    return HookResult(
        message=system_message,  # Claude 컨텍스트용
        systemMessage=system_message,  # 사용자 표시용
    )


def handle_session_end(payload: HookPayload) -> HookResult:
    """SessionEnd 이벤트 핸들러 (기본 구현)"""
    return HookResult()


__all__ = ["handle_session_start", "handle_session_end"]
