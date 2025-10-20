# @TEST:I18N-001 | SPEC: SPEC-I18N-001.md
"""
Simple integration tests for session i18n.
세션 i18n 간단한 통합 테스트.
"""
import sys
from pathlib import Path

import pytest

# Add .claude/hooks/alfred to sys.path
hooks_path = Path(__file__).parent.parent / ".claude" / "hooks" / "alfred"
sys.path.insert(0, str(hooks_path))

from handlers.session import handle_session_start  # noqa: E402


def test_session_start_with_actual_project():
    """Test session start with actual project directory"""
    project_root = Path(__file__).parent.parent

    payload = {
        "cwd": str(project_root),
        "phase": "compact",
    }

    result = handle_session_start(payload)

    # Verify message exists
    assert result.systemMessage is not None
    assert "🚀 MoAI-ADK" in result.systemMessage
    assert "python" in result.systemMessage.lower()
