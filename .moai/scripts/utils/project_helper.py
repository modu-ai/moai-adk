#!/usr/bin/env python3
"""
MoAI-ADK 프로젝트 관련 헬퍼 유틸리티

@REQ:PROJECT-UTILS-001
@FEATURE:PROJECT-MANAGEMENT-001
@API:GET-PROJECT-INFO
@DESIGN:PROJECT-ABSTRACTION-001
"""

import json
import logging
from pathlib import Path
from typing import Any

from constants import (
    CLAUDE_MEMORY_FILE_NAME,
    CONFIG_FILE_NAME,
    DEVELOPMENT_GUIDE_FILE_NAME,
    ERROR_MESSAGES,
    MEMORY_DIR_NAME,
    MOAI_DIR_NAME,
    PERSONAL_MODE,
    VALID_MODES,
)

logger = logging.getLogger(__name__)


class ProjectHelper:
    """프로젝트 관련 헬퍼 클래스"""

    @staticmethod
    def find_project_root(start_path: Path | None = None) -> Path:
        """
        프로젝트 루트 디렉터리를 찾습니다.

        Args:
            start_path: 시작 경로 (기본값: 현재 작업 디렉터리)

        Returns:
            프로젝트 루트 경로

        Raises:
            FileNotFoundError: 프로젝트 루트를 찾을 수 없는 경우
        """
        if start_path is None:
            start_path = Path.cwd()

        current = start_path.resolve()

        # 최대 5단계까지 상위 디렉터리 탐색
        for _ in range(5):
            # .moai 디렉터리가 있으면 프로젝트 루트
            if (current / MOAI_DIR_NAME).exists():
                return current

            # .git 디렉터리가 있으면서 setup.py나 pyproject.toml이 있으면 프로젝트 루트
            if (current / ".git").exists():
                if (current / "setup.py").exists() or (
                    current / "pyproject.toml"
                ).exists():
                    return current

            parent = current.parent
            if parent == current:  # 루트 디렉터리에 도달
                break
            current = parent

        # 찾지 못한 경우, 현재 스크립트 위치 기준으로 추정
        script_path = Path(__file__).resolve()
        if ".moai" in str(script_path):
            # .moai/scripts/utils/project_helper.py -> project_root
            project_root = script_path.parents[3]
            if (project_root / MOAI_DIR_NAME).exists():
                return project_root

        raise FileNotFoundError(
            f"프로젝트 루트를 찾을 수 없습니다. 시작 경로: {start_path}"
        )

    @staticmethod
    def load_config(project_root: Path | None = None) -> dict[str, Any]:
        """
        MoAI 설정 파일을 로드합니다.

        Args:
            project_root: 프로젝트 루트 경로

        Returns:
            설정 딕셔너리

        Raises:
            FileNotFoundError: 설정 파일이 없는 경우
            json.JSONDecodeError: JSON 파싱 오류
        """
        if project_root is None:
            project_root = ProjectHelper.find_project_root()

        config_path = project_root / MOAI_DIR_NAME / CONFIG_FILE_NAME

        if not config_path.exists():
            logger.warning(f"설정 파일이 없습니다: {config_path}")
            return ProjectHelper._get_default_config()

        try:
            with open(config_path, encoding="utf-8") as f:
                config = json.load(f)

            # 기본값과 병합
            default_config = ProjectHelper._get_default_config()
            merged_config = ProjectHelper._merge_configs(default_config, config)

            return merged_config

        except json.JSONDecodeError as e:
            logger.error(f"설정 파일 파싱 오류: {config_path}, {e}")
            raise
        except Exception as e:
            logger.error(f"설정 파일 로드 오류: {config_path}, {e}")
            raise

    @staticmethod
    def save_config(config: dict[str, Any], project_root: Path | None = None) -> None:
        """
        MoAI 설정 파일을 저장합니다.

        Args:
            config: 저장할 설정 딕셔너리
            project_root: 프로젝트 루트 경로
        """
        if project_root is None:
            project_root = ProjectHelper.find_project_root()

        config_path = project_root / MOAI_DIR_NAME / CONFIG_FILE_NAME
        config_path.parent.mkdir(parents=True, exist_ok=True)

        try:
            with open(config_path, "w", encoding="utf-8") as f:
                json.dump(config, f, indent=2, ensure_ascii=False)
            logger.info(f"설정 파일 저장 완료: {config_path}")
        except Exception as e:
            logger.error(f"설정 파일 저장 오류: {config_path}, {e}")
            raise

    @staticmethod
    def get_project_mode(project_root: Path | None = None) -> str:
        """
        현재 프로젝트 모드를 반환합니다.

        Args:
            project_root: 프로젝트 루트 경로

        Returns:
            프로젝트 모드 ('personal' 또는 'team')
        """
        config = ProjectHelper.load_config(project_root)
        return config.get("mode", PERSONAL_MODE)

    @staticmethod
    def set_project_mode(mode: str, project_root: Path | None = None) -> None:
        """
        프로젝트 모드를 설정합니다.

        Args:
            mode: 설정할 모드 ('personal' 또는 'team')
            project_root: 프로젝트 루트 경로

        Raises:
            ValueError: 유효하지 않은 모드
        """
        if mode not in VALID_MODES:
            raise ValueError(ERROR_MESSAGES["invalid_mode"])

        config = ProjectHelper.load_config(project_root)
        config["mode"] = mode
        ProjectHelper.save_config(config, project_root)
        logger.info(f"프로젝트 모드 변경: {mode}")

    @staticmethod
    def is_moai_project(project_root: Path | None = None) -> bool:
        """
        MoAI 프로젝트인지 확인합니다.

        Args:
            project_root: 프로젝트 루트 경로

        Returns:
            MoAI 프로젝트 여부
        """
        if project_root is None:
            try:
                project_root = ProjectHelper.find_project_root()
            except FileNotFoundError:
                return False

        moai_dir = project_root / MOAI_DIR_NAME
        return moai_dir.exists() and moai_dir.is_dir()

    @staticmethod
    def get_project_info(project_root: Path | None = None) -> dict[str, Any]:
        """
        프로젝트 정보를 반환합니다.

        Args:
            project_root: 프로젝트 루트 경로

        Returns:
            프로젝트 정보 딕셔너리
        """
        if project_root is None:
            project_root = ProjectHelper.find_project_root()

        config = ProjectHelper.load_config(project_root)

        info = {
            "name": project_root.name,
            "root": str(project_root),
            "mode": config.get("mode", PERSONAL_MODE),
            "is_moai_project": ProjectHelper.is_moai_project(project_root),
            "has_git": (project_root / ".git").exists(),
            "has_claude": (project_root / ".claude").exists(),
            "config": config,
        }

        return info

    @staticmethod
    def get_development_guide_path(project_root: Path | None = None) -> Path:
        """
        개발 가이드 파일 경로를 반환합니다.

        Args:
            project_root: 프로젝트 루트 경로

        Returns:
            개발 가이드 파일 경로
        """
        if project_root is None:
            project_root = ProjectHelper.find_project_root()

        return (
            project_root / MOAI_DIR_NAME / MEMORY_DIR_NAME / DEVELOPMENT_GUIDE_FILE_NAME
        )

    @staticmethod
    def get_claude_memory_path(project_root: Path | None = None) -> Path:
        """
        Claude 메모리 파일 경로를 반환합니다.

        Args:
            project_root: 프로젝트 루트 경로

        Returns:
            Claude 메모리 파일 경로
        """
        if project_root is None:
            project_root = ProjectHelper.find_project_root()

        return project_root / CLAUDE_MEMORY_FILE_NAME

    @staticmethod
    def list_specs(project_root: Path | None = None) -> list[dict[str, Any]]:
        """
        SPEC 목록을 반환합니다.

        Args:
            project_root: 프로젝트 루트 경로

        Returns:
            SPEC 정보 리스트
        """
        if project_root is None:
            project_root = ProjectHelper.find_project_root()

        specs_dir = project_root / MOAI_DIR_NAME / "specs"
        if not specs_dir.exists():
            return []

        specs = []
        for spec_dir in specs_dir.iterdir():
            if spec_dir.is_dir() and spec_dir.name.startswith("SPEC-"):
                spec_file = spec_dir / "spec.md"
                if spec_file.exists():
                    specs.append(
                        {
                            "id": spec_dir.name,
                            "path": str(spec_dir),
                            "spec_file": str(spec_file),
                            "exists": True,
                        }
                    )

        return sorted(specs, key=lambda x: x["id"])

    @staticmethod
    def _get_default_config() -> dict[str, Any]:
        """기본 설정 반환"""
        return {
            "mode": PERSONAL_MODE,
            "checkpoint": {"auto_create": True, "max_count": 10, "interval_minutes": 5},
            "git": {"auto_push": False, "default_branch": "main"},
            "trust_principles": {"strict_mode": False, "min_coverage": 85},
        }

    @staticmethod
    def _merge_configs(default: dict[str, Any], user: dict[str, Any]) -> dict[str, Any]:
        """설정 병합"""
        merged = default.copy()

        for key, value in user.items():
            if (
                key in merged
                and isinstance(merged[key], dict)
                and isinstance(value, dict)
            ):
                merged[key] = ProjectHelper._merge_configs(merged[key], value)
            else:
                merged[key] = value

        return merged
