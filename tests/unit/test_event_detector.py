"""
Event Detector 테스트.

위험한 작업 감지 기능을 테스트합니다.

SPEC: .moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md
"""

from pathlib import Path

import pytest

from moai_adk.core.git.event_detector import EventDetector


class TestEventDetector:
    """EventDetector 위험 작업 감지 테스트."""

    @pytest.fixture
    def detector(self):
        """EventDetector 인스턴스."""
        return EventDetector()

    # TEST-CHECKPOINT-EVENT-001: 대규모 파일 삭제 감지
    def test_should_detect_large_file_deletion(self, detector):
        """10개 이상 파일 삭제 시 위험 작업으로 감지해야 한다."""
        deleted_files = [f"file{i}.py" for i in range(10)]
        assert detector.is_risky_deletion(deleted_files) is True

    def test_should_not_detect_small_file_deletion(self, detector):
        """9개 이하 파일 삭제는 안전한 작업으로 판단해야 한다."""
        deleted_files = [f"file{i}.py" for i in range(9)]
        assert detector.is_risky_deletion(deleted_files) is False

    # TEST-CHECKPOINT-EVENT-002: 복잡한 리팩토링 감지
    def test_should_detect_large_refactoring(self, detector):
        """10개 이상 파일 이름 변경 시 위험 작업으로 감지해야 한다."""
        renamed_files = [(f"old{i}.py", f"new{i}.py") for i in range(10)]
        assert detector.is_risky_refactoring(renamed_files) is True

    def test_should_not_detect_small_refactoring(self, detector):
        """9개 이하 파일 이름 변경은 안전한 작업으로 판단해야 한다."""
        renamed_files = [(f"old{i}.py", f"new{i}.py") for i in range(9)]
        assert detector.is_risky_refactoring(renamed_files) is False

    # TEST-CHECKPOINT-EVENT-003: 중요 파일 수정 감지
    def test_should_detect_critical_file_modification(self, detector):
        """CLAUDE.md, config.json 등 중요 파일 수정 시 위험 작업으로 감지해야 한다."""
        critical_files = ["CLAUDE.md", ".moai/config.json", ".moai/memory/development-guide.md"]
        for file in critical_files:
            assert detector.is_critical_file(Path(file)) is True

    def test_should_not_detect_normal_file_modification(self, detector):
        """일반 파일 수정은 안전한 작업으로 판단해야 한다."""
        normal_files = ["src/module.py", "tests/test_something.py", "README.md"]
        for file in normal_files:
            assert detector.is_critical_file(Path(file)) is False
