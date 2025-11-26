"""
Tests for AlfredDetector - Alfred 작업 감지

"""

import json
import tempfile
from pathlib import Path

import pytest


class TestAlfredDetector:
    """Alfred 작업 감지 테스트"""

    def test_detect_active_command(self):
        """
        GIVEN: /moai:2-run SPEC-AUTH-001 실행 중
        WHEN: detect_active_task() 호출
        THEN: command="run", spec_id="AUTH-001"
        """
        from moai_adk.statusline.alfred_detector import AlfredDetector

        detector = AlfredDetector()

        # Mock the session state file
        with tempfile.TemporaryDirectory() as tmpdir:
            state_file = Path(tmpdir) / "last-session-state.json"
            state_data = {"active_task": {"command": "run", "spec_id": "SPEC-AUTH-001"}}
            state_file.write_text(json.dumps(state_data))

            detector._session_state_path = state_file

            task = detector.detect_active_task()

            assert task.spec_id in [
                "SPEC-AUTH-001",
                "AUTH-001",
            ], f"Expected spec_id with AUTH-001, got {task.spec_id}"

    def test_no_active_command(self):
        """
        GIVEN: Alfred 명령 미실행 상태
        WHEN: detect_active_task() 호출
        THEN: spec_id=None
        """
        from moai_adk.statusline.alfred_detector import AlfredDetector

        detector = AlfredDetector()

        # Mock the session state file with no active task
        with tempfile.TemporaryDirectory() as tmpdir:
            state_file = Path(tmpdir) / "last-session-state.json"
            state_data = {"active_task": None}
            state_file.write_text(json.dumps(state_data))

            detector._session_state_path = state_file

            task = detector.detect_active_task()

            assert task.spec_id is None, f"Expected None spec_id, got {task.spec_id}"

    def test_alfred_detector_caching(self):
        """
        GIVEN: AlfredDetector 인스턴스
        WHEN: 1초 이내에 두 번 detect_active_task() 호출
        THEN: 캐시에서 반환
        """
        from moai_adk.statusline.alfred_detector import AlfredDetector

        detector = AlfredDetector()

        with tempfile.TemporaryDirectory() as tmpdir:
            state_file = Path(tmpdir) / "last-session-state.json"
            state_data = {"active_task": {"command": "plan"}}
            state_file.write_text(json.dumps(state_data))

            detector._session_state_path = state_file

            # First call
            result1 = detector.detect_active_task()

            # Second call immediately (should use cache)
            result2 = detector.detect_active_task()

            # Results should be identical
            assert result1.command == result2.command, "Cache should return identical results"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
