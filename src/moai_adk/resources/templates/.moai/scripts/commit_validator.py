#!/usr/bin/env python3
"""
Commit Validator Module for MoAI Commit Helper
Handles commit message validation with TRUST principles

@TASK:VALIDATION-001
@FEATURE:COMMIT-VALIDATION-001
@API:VALIDATE-COMMIT
@TRUST:SECURED
"""

import re
from typing import Any


class CommitValidator:
    """Commit message validator ensuring TRUST compliance"""

    # Constants for validation rules
    MIN_MESSAGE_LENGTH = 10
    MAX_MESSAGE_LENGTH = 200
    EMOJI_PATTERN = re.compile(
        r"[\U0001F600-\U0001F64F\U0001F300-\U0001F5FF\U0001F680-\U0001F6FF\U0001F1E0-\U0001F1FF]"
    )

    def validate_commit_message(self, message: str) -> dict[str, Any]:
        """Validate commit message according to TRUST principles"""
        # Guard clause: Empty message check
        if not message or not message.strip():
            return self._create_validation_result(
                False, "메시지가 비어있습니다", message
            )

        # Guard clause: Length validation
        length_validation = self._validate_length(message)
        if not length_validation["valid"]:
            return length_validation

        # Additional validations
        emoji_check = self._check_emoji_presence(message)

        return self._create_validation_result(
            True, "검증 통과", message, emoji_check
        )

    def validate_file_list(self, files: list[str] | None) -> dict[str, Any]:
        """Validate file list for commit"""
        if files is None:
            return {"valid": True, "reason": "전체 파일 커밋"}

        if not files:
            return {"valid": False, "reason": "파일 목록이 비어있습니다"}

        # Check for dangerous patterns
        dangerous_patterns = [".env", "secret", "password", "key", ".pem"]
        dangerous_files = []

        for file in files:
            if any(pattern in file.lower() for pattern in dangerous_patterns):
                dangerous_files.append(file)

        if dangerous_files:
            return {
                "valid": False,
                "reason": f"민감한 파일이 포함되어 있습니다: {', '.join(dangerous_files)}",
                "dangerous_files": dangerous_files,
            }

        return {
            "valid": True,
            "reason": f"{len(files)}개 파일 검증 통과",
            "file_count": len(files),
        }

    def validate_change_context(self, changes: dict[str, Any]) -> dict[str, Any]:
        """Validate change context for commit safety"""
        if not changes.get("success", False):
            return {"valid": False, "reason": "변경사항 조회 실패"}

        if not changes.get("has_changes", False):
            return {"valid": False, "reason": "커밋할 변경사항이 없습니다"}

        file_count = changes.get("count", 0)
        if file_count > 50:
            return {
                "valid": False,
                "reason": f"너무 많은 파일이 변경되었습니다 ({file_count}개). 작은 단위로 커밋하세요.",
                "file_count": file_count,
            }

        return {
            "valid": True,
            "reason": f"{file_count}개 파일 변경 검증 통과",
            "file_count": file_count,
        }

    def _validate_length(self, message: str) -> dict[str, Any]:
        """Validate message length constraints"""
        length = len(message)

        if length < self.MIN_MESSAGE_LENGTH:
            return self._create_validation_result(
                False, f"메시지가 너무 짧습니다 (최소 {self.MIN_MESSAGE_LENGTH}자)", message
            )

        if length > self.MAX_MESSAGE_LENGTH:
            return self._create_validation_result(
                False, f"메시지가 너무 깁니다 (최대 {self.MAX_MESSAGE_LENGTH}자)", message
            )

        return {"valid": True, "length": length}

    def _check_emoji_presence(self, message: str) -> bool:
        """Check if message contains emoji"""
        return bool(self.EMOJI_PATTERN.search(message))

    def _create_validation_result(
        self, valid: bool, reason: str, message: str, has_emoji: bool = False
    ) -> dict[str, Any]:
        """Create standardized validation result"""
        return {
            "valid": valid,
            "reason": reason,
            "has_emoji": has_emoji or self._check_emoji_presence(message),
            "length": len(message),
        }