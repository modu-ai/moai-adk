# @CODE:CORE-GIT-001 | SPEC: SPEC-CORE-GIT-001.md
"""커밋 메시지 포맷팅 유틸리티"""
from typing import Literal, Optional

def format_commit_message(
    spec_id: str,
    stage: Literal['RED', 'GREEN', 'REFACTOR'],
    message: Optional[str] = None,
    locale: str = 'ko'
) -> str:
    """
    TDD 커밋 메시지 포맷팅

    :param spec_id: SPEC ID (예: CORE-GIT-001)
    :param stage: TDD 단계 (RED, GREEN, REFACTOR)
    :param message: 선택적 커스텀 메시지
    :param locale: 메시지 로캘 (ko, en, ja, zh)
    :return: 포맷팅된 커밋 메시지
    """
    stage_emojis = {
        'RED': '🔴',
        'GREEN': '🟢',
        'REFACTOR': '♻️'
    }

    stage_messages = {
        'ko': {
            'RED': 'RED 단계: 테스트 작성',
            'GREEN': 'GREEN 단계: 구현 완료',
            'REFACTOR': 'REFACTOR: 코드 개선'
        },
        'en': {
            'RED': 'RED Stage: Writing tests',
            'GREEN': 'GREEN Stage: Implementation complete',
            'REFACTOR': 'REFACTOR: Code improvement'
        },
        # 다른 언어 지원 가능
    }

    emoji = stage_emojis.get(stage, '🏷️')
    default_msg = stage_messages.get(locale, stage_messages['ko']).get(stage, f"{stage} 단계")

    message_text = message or default_msg
    return f"{emoji} {stage.upper()}: {message_text}\n\n@TAG:{spec_id}-{stage}"