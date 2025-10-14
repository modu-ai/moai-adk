# @CODE:PY314-001 | SPEC: SPEC-PY314-001.md | TEST: tests/unit/test_config_manager.py
"""Configuration Manager

.moai/config.json 파일 관리:
- 설정 파일 읽기/쓰기
- 깊은 병합 (deep merge) 지원
- 한글 UTF-8 보존
- 디렉토리 자동 생성
"""

import json
from pathlib import Path
from typing import Any


class ConfigManager:
    """Configuration Manager 클래스

    .moai/config.json 파일을 읽고 쓰는 관리자입니다.
    """

    DEFAULT_CONFIG = {
        "mode": "personal",
        "locale": "ko",
        "moai": {
            "version": "0.3.0"
        }
    }

    def __init__(self, config_path: Path) -> None:
        """ConfigManager 초기화

        Args:
            config_path: config.json 파일 경로
        """
        self.config_path = config_path

    def load(self) -> dict[str, Any]:
        """설정 파일 로드

        파일이 없으면 기본 설정을 반환합니다.

        Returns:
            설정 딕셔너리
        """
        if not self.config_path.exists():
            return self.DEFAULT_CONFIG.copy()

        with open(self.config_path, encoding="utf-8") as f:
            data: dict[str, Any] = json.load(f)
            return data

    def save(self, config: dict[str, Any]) -> None:
        """설정 파일 저장

        디렉토리가 없으면 자동으로 생성합니다.
        한글을 보존합니다 (ensure_ascii=False).

        Args:
            config: 저장할 설정 딕셔너리
        """
        # 디렉토리 생성
        self.config_path.parent.mkdir(parents=True, exist_ok=True)

        # 한글 보존하여 저장
        with open(self.config_path, "w", encoding="utf-8") as f:
            json.dump(config, f, ensure_ascii=False, indent=2)

    def update(self, updates: dict[str, Any]) -> None:
        """설정 업데이트 (깊은 병합)

        기존 설정에 새 설정을 깊은 병합하여 저장합니다.
        중첩된 딕셔너리는 재귀적으로 병합됩니다.

        Args:
            updates: 업데이트할 설정 딕셔너리
        """
        current = self.load()
        merged = self._deep_merge(current, updates)
        self.save(merged)

    def _deep_merge(self, base: dict[str, Any], updates: dict[str, Any]) -> dict[str, Any]:
        """딕셔너리 깊은 병합 (재귀)

        Args:
            base: 기본 딕셔너리
            updates: 업데이트 딕셔너리

        Returns:
            병합된 딕셔너리
        """
        result = base.copy()

        for key, value in updates.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                # 양쪽 모두 dict면 재귀적으로 병합
                result[key] = self._deep_merge(result[key], value)
            else:
                # 그 외에는 덮어쓰기
                result[key] = value

        return result
