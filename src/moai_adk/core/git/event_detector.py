# @CODE:CHECKPOINT-EVENT-001 | SPEC: SPEC-CHECKPOINT-EVENT-001.md | TEST: tests/unit/test_event_detector.py
"""
Event Detector - 위험한 작업 감지.

SPEC: .moai/specs/SPEC-CHECKPOINT-EVENT-001/spec.md
"""

from pathlib import Path


class EventDetector:
    """위험한 작업을 감지하는 이벤트 감지기."""

    # @CODE:CHECKPOINT-EVENT-001:DOMAIN - 중요 파일 목록
    CRITICAL_FILES = {
        "CLAUDE.md",
        "config.json",
        ".moai/config.json",
    }

    CRITICAL_DIRS = {
        ".moai/memory",
    }

    def is_risky_deletion(self, deleted_files: list[str]) -> bool:
        """
        대규모 파일 삭제 감지.

        SPEC 요구사항: 10개 이상 파일 삭제 시 위험 작업으로 판단

        Args:
            deleted_files: 삭제될 파일 목록

        Returns:
            10개 이상이면 True, 아니면 False
        """
        return len(deleted_files) >= 10

    def is_risky_refactoring(self, renamed_files: list[tuple[str, str]]) -> bool:
        """
        복잡한 리팩토링 감지.

        SPEC 요구사항: 10개 이상 파일 이름 변경 시 위험 작업으로 판단

        Args:
            renamed_files: (old_name, new_name) 튜플 목록

        Returns:
            10개 이상이면 True, 아니면 False
        """
        return len(renamed_files) >= 10

    def is_critical_file(self, file_path: Path) -> bool:
        """
        중요 파일 수정 감지.

        SPEC 요구사항: CLAUDE.md, config.json, .moai/memory/*.md 수정 시 위험 작업

        Args:
            file_path: 확인할 파일 경로

        Returns:
            중요 파일이면 True, 아니면 False
        """
        # 파일명이 중요 파일 목록에 있는지 확인
        if file_path.name in self.CRITICAL_FILES:
            return True

        # 경로 문자열로 변환하여 확인
        path_str = str(file_path)

        # .moai/config.json 확인
        if ".moai/config.json" in path_str or ".moai\\config.json" in path_str:
            return True

        # .moai/memory/ 디렉토리 내 파일 확인
        for critical_dir in self.CRITICAL_DIRS:
            if critical_dir in path_str or critical_dir.replace("/", "\\") in path_str:
                return True

        return False
