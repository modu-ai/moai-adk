"""템플릿 복사 및 백업 프로세서 (SPEC-INIT-003 v0.3.0: 사용자 콘텐츠 보존)."""

from __future__ import annotations

import json
import shutil
from datetime import datetime
from pathlib import Path
from typing import Any

from rich.console import Console

console = Console()


class TemplateProcessor:
    """템플릿 복사 및 백업 관리 클래스."""

    # 사용자 데이터 보호 경로 (절대 건드리지 않음) - SPEC-INIT-003 v0.3.0
    PROTECTED_PATHS = [
        ".moai/specs/",       # 사용자 SPEC 문서
        ".moai/reports/",     # 사용자 리포트
        ".moai/project/",     # 사용자 프로젝트 문서 (product/structure/tech.md)
        ".moai/config.json",  # 사용자 설정 (병합은 /alfred:9-update에서)
    ]

    # 백업 제외 경로
    BACKUP_EXCLUDE = PROTECTED_PATHS

    def __init__(self, target_path: Path) -> None:
        """초기화.

        Args:
            target_path: 프로젝트 경로
        """
        self.target_path = target_path.resolve()
        self.template_root = self._get_template_root()

    def _get_template_root(self) -> Path:
        """템플릿 루트 경로 반환.

        Returns:
            템플릿 루트 경로
        """
        # src/moai_adk/core/template/processor.py → src/moai_adk/templates/
        current_file = Path(__file__).resolve()
        package_root = current_file.parent.parent.parent
        return package_root / "templates"

    def copy_templates(self, backup: bool = True, silent: bool = False) -> None:
        """템플릿 파일을 프로젝트에 복사.

        Args:
            backup: 백업 생성 여부
            silent: 조용한 모드 (로그 출력 최소화)
        """
        # 1. 백업 생성 (기존 파일 있으면)
        if backup and self._has_existing_files():
            backup_path = self.create_backup()
            if not silent:
                console.print(f"💾 백업 생성: {backup_path.name}")

        # 2. 템플릿 복사
        if not silent:
            console.print("📄 템플릿 복사 중...")

        self._copy_claude(silent)
        self._copy_moai(silent)
        self._copy_claude_md(silent)
        self._copy_gitignore(silent)

        if not silent:
            console.print("✅ 템플릿 복사 완료")

    def _has_existing_files(self) -> bool:
        """기존 프로젝트 파일 존재 여부 확인.

        백업 정책:
        - .moai/, .claude/, CLAUDE.md 중 **1개라도 존재하면 백업 생성**
        - 백업 경로: .moai-backup/YYYYMMDD-HHMMSS/
        - 보호 경로: .moai/specs/, .moai/reports/ (백업 제외)

        덮어쓰기 정책:
        - 동일 파일명은 **복사 시 덮어쓰기**
        - .claude/ → 전체 삭제 후 재복사
        - .moai/ → 보호 경로 제외하고 복사 (덮어쓰기)
        - CLAUDE.md → 스마트 병합 (프로젝트 정보 유지)

        Returns:
            True if 백업 필요 (파일 1개 이상 존재)
        """
        return any(
            (self.target_path / item).exists()
            for item in [".moai", ".claude", "CLAUDE.md"]
        )

    def create_backup(self) -> Path:
        """타임스탬프 기반 백업 생성.

        Returns:
            백업 경로
        """
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        backup_path = self.target_path / ".moai-backup" / timestamp
        backup_path.mkdir(parents=True, exist_ok=True)

        for item in [".moai", ".claude", "CLAUDE.md"]:
            src = self.target_path / item
            if not src.exists():
                continue

            dst = backup_path / item

            if item == ".moai":
                # PROTECTED_PATHS 제외하고 복사
                self._copy_exclude_protected(src, dst)
            elif src.is_dir():
                shutil.copytree(src, dst, dirs_exist_ok=True)
            else:
                shutil.copy2(src, dst)

        return backup_path

    def _copy_exclude_protected(self, src: Path, dst: Path) -> None:
        """보호 경로를 제외하고 복사 (SPEC-INIT-003 v0.3.0: 기존 파일 보존).

        Args:
            src: 소스 디렉토리
            dst: 대상 디렉토리
        """
        dst.mkdir(parents=True, exist_ok=True)

        # PROTECTED_PATHS: specs/, reports/만 템플릿 복사 제외
        # project/, config.json은 기존 파일 존재 시에만 보존
        template_protected_paths = [
            "specs",
            "reports",
        ]

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # 템플릿 복사 제외 경로 (specs/, reports/)
            if any(rel_path_str.startswith(p) for p in template_protected_paths):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                # v0.3.0: 기존 파일이 존재하면 건너뛰기 (사용자 콘텐츠 보존)
                # 이렇게 하면 project/, config.json도 자동 보호됨
                if dst_item.exists():
                    continue
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

    def _copy_claude(self, silent: bool = False) -> None:
        """.claude/ 디렉토리 복사."""
        src = self.template_root / ".claude"
        dst = self.target_path / ".claude"

        if not src.exists():
            if not silent:
                console.print("⚠️ .claude/ 템플릿 없음")
            return

        # 전체 복사 (덮어쓰기)
        if dst.exists():
            shutil.rmtree(dst)
        shutil.copytree(src, dst)
        if not silent:
            console.print("   ✅ .claude/ 복사 완료")

    def _copy_moai(self, silent: bool = False) -> None:
        """.moai/ 디렉토리 복사 (보호 경로 제외, SPEC-INIT-003 v0.3.0)."""
        src = self.template_root / ".moai"
        dst = self.target_path / ".moai"

        if not src.exists():
            if not silent:
                console.print("⚠️ .moai/ 템플릿 없음")
            return

        # 템플릿 복사 제외 경로 (specs/, reports/)
        template_protected_paths = [
            "specs",
            "reports",
        ]

        # 보호 경로 제외하고 복사
        for item in src.rglob("*"):
            rel_path = item.relative_to(src)
            rel_path_str = str(rel_path)

            # 템플릿 복사 제외 (specs/, reports/)
            if any(rel_path_str.startswith(p) for p in template_protected_paths):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                # v0.3.0: 기존 파일이 존재하면 건너뛰기 (사용자 콘텐츠 보존)
                if dst_item.exists():
                    continue
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

        if not silent:
            console.print("   ✅ .moai/ 복사 완료 (user content preserved)")

    def _copy_claude_md(self, silent: bool = False) -> None:
        """CLAUDE.md 복사 (스마트 병합)."""
        src = self.template_root / "CLAUDE.md"
        dst = self.target_path / "CLAUDE.md"

        if not src.exists():
            if not silent:
                console.print("⚠️ CLAUDE.md 템플릿 없음")
            return

        # 기존 파일 있으면 프로젝트 정보 유지
        if dst.exists():
            self._merge_claude_md(src, dst)
            if not silent:
                console.print("   🔄 CLAUDE.md 병합 (프로젝트 정보 유지)")
        else:
            shutil.copy2(src, dst)
            if not silent:
                console.print("   ✅ CLAUDE.md 복사 완료")

    def _merge_claude_md(self, src: Path, dst: Path) -> None:
        """CLAUDE.md 스마트 병합.

        Args:
            src: 템플릿 CLAUDE.md
            dst: 프로젝트 CLAUDE.md
        """
        # 기존 프로젝트 정보 섹션 추출
        dst_content = dst.read_text(encoding="utf-8")
        project_info_start = dst_content.find("## 프로젝트 정보")
        project_info = ""
        if project_info_start != -1:
            # EOF까지 추출
            project_info = dst_content[project_info_start:]

        # 템플릿 내용 가져오기
        src_content = src.read_text(encoding="utf-8")

        # 프로젝트 정보가 있으면 병합
        if project_info:
            # 템플릿에서 프로젝트 정보 제거
            template_project_start = src_content.find("## 프로젝트 정보")
            if template_project_start != -1:
                src_content = src_content[:template_project_start].rstrip()

            # 병합
            merged_content = f"{src_content}\n\n{project_info}"
            dst.write_text(merged_content, encoding="utf-8")
        else:
            # 프로젝트 정보 없으면 템플릿 그대로 복사
            shutil.copy2(src, dst)

    def _copy_gitignore(self, silent: bool = False) -> None:
        """.gitignore 복사 (선택)."""
        src = self.template_root / ".gitignore"
        dst = self.target_path / ".gitignore"

        if not src.exists():
            return

        # 기존 .gitignore 있으면 병합
        if dst.exists():
            self._merge_gitignore(src, dst)
            if not silent:
                console.print("   🔄 .gitignore 병합")
        else:
            shutil.copy2(src, dst)
            if not silent:
                console.print("   ✅ .gitignore 복사 완료")

    def _merge_gitignore(self, src: Path, dst: Path) -> None:
        """.gitignore 병합.

        Args:
            src: 템플릿 .gitignore
            dst: 프로젝트 .gitignore
        """
        src_lines = set(src.read_text(encoding="utf-8").splitlines())
        dst_lines = dst.read_text(encoding="utf-8").splitlines()

        # 중복 제거하고 병합
        merged_lines = dst_lines + [
            line for line in src_lines if line not in dst_lines
        ]

        dst.write_text("\n".join(merged_lines) + "\n", encoding="utf-8")

    def merge_config(self, detected_language: str | None = None) -> dict[str, str]:
        """config.json 스마트 병합.

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

        # 새 config 생성
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
