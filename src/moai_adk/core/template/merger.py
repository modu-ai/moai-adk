# @CODE:TEMPLATE-001 | SPEC: SPEC-INIT-003.md | Chain: TEMPLATE-001
"""템플릿 파일 병합 클래스 (SPEC-INIT-003 v0.3.0).

기존 사용자 파일과 새 템플릿을 스마트하게 병합.
"""

from __future__ import annotations

import json
import shutil
from pathlib import Path
from typing import Any


class TemplateMerger:
    """템플릿 파일 병합 로직."""

    def __init__(self, target_path: Path) -> None:
        """초기화.

        Args:
            target_path: 프로젝트 경로 (절대 경로)
        """
        self.target_path = target_path.resolve()

    def merge_claude_md(self, template_path: Path, existing_path: Path) -> None:
        """CLAUDE.md 스마트 병합.

        병합 규칙:
        - 템플릿의 최신 구조/내용 사용
        - 기존 "## 프로젝트 정보" 섹션 유지

        Args:
            template_path: 템플릿 CLAUDE.md
            existing_path: 기존 CLAUDE.md
        """
        # 기존 프로젝트 정보 섹션 추출
        existing_content = existing_path.read_text(encoding="utf-8")
        project_info_start = existing_content.find("## 프로젝트 정보")
        project_info = ""
        if project_info_start != -1:
            # EOF까지 추출
            project_info = existing_content[project_info_start:]

        # 템플릿 내용 가져오기
        template_content = template_path.read_text(encoding="utf-8")

        # 프로젝트 정보가 있으면 병합
        if project_info:
            # 템플릿에서 프로젝트 정보 제거
            template_project_start = template_content.find("## 프로젝트 정보")
            if template_project_start != -1:
                template_content = template_content[:template_project_start].rstrip()

            # 병합
            merged_content = f"{template_content}\n\n{project_info}"
            existing_path.write_text(merged_content, encoding="utf-8")
        else:
            # 프로젝트 정보 없으면 템플릿 그대로 복사
            shutil.copy2(template_path, existing_path)

    def merge_gitignore(self, template_path: Path, existing_path: Path) -> None:
        """.gitignore 병합.

        병합 규칙:
        - 기존 항목 유지
        - 템플릿 신규 항목 추가
        - 중복 제거

        Args:
            template_path: 템플릿 .gitignore
            existing_path: 기존 .gitignore
        """
        template_lines = set(template_path.read_text(encoding="utf-8").splitlines())
        existing_lines = existing_path.read_text(encoding="utf-8").splitlines()

        # 중복 제거하고 병합
        merged_lines = existing_lines + [
            line for line in template_lines if line not in existing_lines
        ]

        existing_path.write_text("\n".join(merged_lines) + "\n", encoding="utf-8")

    def merge_config(self, detected_language: str | None = None) -> dict[str, str]:
        """config.json 스마트 병합.

        병합 규칙:
        - 기존 설정값 우선 사용
        - 신규 프로젝트는 감지된 언어 + 기본값 사용

        Args:
            detected_language: 감지된 언어

        Returns:
            병합된 config
        """
        config_path = self.target_path / ".moai" / "config.json"

        # 기존 config 읽기
        existing_config: dict[str, Any] = {}
        if config_path.exists():
            with open(config_path, encoding="utf-8") as f:
                existing_config = json.load(f)

        # 새 config 생성 (기존 값 우선)
        new_config: dict[str, str] = {
            "projectName": existing_config.get(
                "projectName", self.target_path.name
            ),
            "mode": existing_config.get("mode", "personal"),
            "locale": existing_config.get("locale", "ko"),
            "language": existing_config.get(
                "language", detected_language or "generic"
            ),
        }

        return new_config
