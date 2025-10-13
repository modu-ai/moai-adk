# @CODE:CORE-TEMPLATE-001 | SPEC: SPEC-CORE-TEMPLATE-001.md | TEST: tests/unit/test_config_manager.py
"""ConfigManager - config.json 관리자

.moai/config.json 파일을 읽고, 쓰고, 업데이트하는 기능을 제공합니다.
"""

import json
from pathlib import Path
from typing import Any


class ConfigManager:
    """MoAI-ADK config.json 관리자

    .moai/config.json 파일의 CRUD 작업을 담당합니다.
    깊은 병합(deep merge) 알고리즘을 사용하여 설정을 안전하게 업데이트합니다.

    Attributes:
        config_path: config.json 파일 경로
        DEFAULT_CONFIG: 기본 설정 딕셔너리

    Examples:
        >>> manager = ConfigManager(".moai/config.json")
        >>> config = manager.load()
        >>> manager.update({"locale": "en"})
    """

    DEFAULT_CONFIG: dict[str, Any] = {
        "moai": {"version": "0.3.0"},
        "mode": "personal",
        "projectName": "",
        "features": [],
        "locale": "ko",
        "git": {"enabled": True, "autoCommit": True, "branchPrefix": ""},
        "spec": {"storage": "local", "workflow": "commit", "localPath": ".moai/specs/"},
        "backup": {"enabled": True, "retentionDays": 30},
        "constitution": {
            "enforce_tdd": True,
            "enforce_spec": False,
            "require_tags": True,
            "test_coverage_target": 85,
        },
    }

    def __init__(self, config_path: str | Path = ".moai/config.json") -> None:
        """ConfigManager 초기화

        Args:
            config_path: config.json 파일 경로 (기본값: ".moai/config.json")
        """
        self.config_path = Path(config_path)

    def load(self) -> dict[str, Any]:
        """config.json 읽기

        파일이 없으면 기본 설정을 반환합니다.

        Returns:
            설정 딕셔너리 (파일이 없으면 DEFAULT_CONFIG의 복사본)

        Examples:
            >>> manager = ConfigManager()
            >>> config = manager.load()
            >>> config["locale"]
            'ko'
        """
        if not self.config_path.exists():
            # shallow copy는 dict[str, Any] 타입을 유지
            import copy

            return copy.deepcopy(self.DEFAULT_CONFIG)

        with open(self.config_path, encoding="utf-8") as f:
            loaded: dict[str, Any] = json.load(f)
            return loaded

    def save(self, config: dict[str, Any]) -> None:
        """config.json 저장

        부모 디렉토리가 없으면 자동으로 생성합니다.
        UTF-8 인코딩, 들여쓰기 2칸, 한글 보존으로 저장합니다.

        Args:
            config: 저장할 설정 딕셔너리

        Examples:
            >>> manager = ConfigManager()
            >>> manager.save({"mode": "team"})
        """
        self.config_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.config_path, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=2, ensure_ascii=False)

    def update(self, updates: dict[str, Any]) -> None:
        """config.json 업데이트 (병합)

        기존 설정과 새로운 설정을 깊은 병합(deep merge)합니다.
        중첩된 딕셔너리는 재귀적으로 병합됩니다.

        Args:
            updates: 업데이트할 설정 딕셔너리

        Examples:
            >>> manager = ConfigManager()
            >>> manager.update({"git": {"autoCommit": False}})
            # git.enabled은 유지, git.autoCommit만 변경됨
        """
        config = self.load()
        self._deep_merge(config, updates)
        self.save(config)

    def _deep_merge(self, base: dict[str, Any], updates: dict[str, Any]) -> None:
        """딕셔너리 깊은 병합

        updates의 모든 키-값 쌍을 base에 재귀적으로 병합합니다.
        중첩된 딕셔너리는 덮어쓰지 않고 병합합니다.

        Args:
            base: 기본 딕셔너리 (in-place로 수정됨)
            updates: 업데이트 딕셔너리

        Examples:
            >>> base = {"a": {"b": 1, "c": 2}}
            >>> updates = {"a": {"c": 3, "d": 4}}
            >>> _deep_merge(base, updates)
            >>> base
            {'a': {'b': 1, 'c': 3, 'd': 4}}
        """
        for key, value in updates.items():
            if key in base and isinstance(base[key], dict) and isinstance(value, dict):
                self._deep_merge(base[key], value)
            else:
                base[key] = value
