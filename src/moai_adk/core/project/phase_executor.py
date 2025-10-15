# @CODE:INIT-003:PHASE | SPEC: .moai/specs/SPEC-INIT-003/spec.md | TEST: tests/unit/test_init_reinit.py
"""Phase 기반 설치 실행 모듈 (SPEC-INIT-003 v0.3.0)

5단계 Phase로 구성된 프로젝트 초기화 실행:
- Phase 1: Preparation (백업 생성: .moai/backups/{timestamp}/)
- Phase 2: Directory (디렉토리 구조 생성)
- Phase 3: Resource (템플릿 복사, 사용자 콘텐츠 보존)
- Phase 4: Configuration (설정 파일 생성)
- Phase 5: Validation (검증 및 마무리)
"""

import shutil
import subprocess
from collections.abc import Callable
from pathlib import Path

from rich.console import Console

from moai_adk.core.project.backup_utils import (
    generate_backup_dir_name,
    get_backup_targets,
    has_any_moai_files,
    is_protected_path,
)
from moai_adk.core.project.validator import ProjectValidator

console = Console()

# 진행상황 콜백 타입
ProgressCallback = Callable[[str, int, int], None]


class PhaseExecutor:
    """Phase 기반 설치 실행

    5단계 Phase:
    1. Preparation: 백업 및 시스템 검증
    2. Directory: 디렉토리 구조 생성
    3. Resource: 템플릿 리소스 복사
    4. Configuration: 설정 파일 생성
    5. Validation: 검증 및 마무리
    """

    # 필수 디렉토리 구조
    REQUIRED_DIRECTORIES = [
        ".moai/",
        ".moai/project/",
        ".moai/specs/",
        ".moai/reports/",
        ".moai/memory/",
        ".moai/backups/",
        ".claude/",
        ".claude/logs/",
    ]

    def __init__(self, validator: ProjectValidator) -> None:
        """초기화

        Args:
            validator: 검증 인스턴스
        """
        self.validator = validator
        self.total_phases = 5
        self.current_phase = 0

    def execute_preparation_phase(
        self,
        project_path: Path,
        backup_enabled: bool = True,
        progress_callback: ProgressCallback | None = None,
    ) -> None:
        """Phase 1: 준비 및 백업

        Args:
            project_path: 프로젝트 경로
            backup_enabled: 백업 활성화 여부
            progress_callback: 진행상황 콜백
        """
        self.current_phase = 1
        self._report_progress(
            "Phase 1: Preparation and backup...", progress_callback
        )

        # 시스템 요구사항 검증
        self.validator.validate_system_requirements()

        # 프로젝트 경로 검증
        self.validator.validate_project_path(project_path)

        # 백업 생성
        if backup_enabled and has_any_moai_files(project_path):
            self._create_backup(project_path)

    def execute_directory_phase(
        self,
        project_path: Path,
        progress_callback: ProgressCallback | None = None,
    ) -> None:
        """Phase 2: 디렉토리 생성

        Args:
            project_path: 프로젝트 경로
            progress_callback: 진행상황 콜백
        """
        self.current_phase = 2
        self._report_progress(
            "Phase 2: Creating directory structure...", progress_callback
        )

        for directory in self.REQUIRED_DIRECTORIES:
            dir_path = project_path / directory
            dir_path.mkdir(parents=True, exist_ok=True)

    def execute_resource_phase(
        self,
        project_path: Path,
        progress_callback: ProgressCallback | None = None,
    ) -> list[str]:
        """Phase 3: 리소스 설치

        Args:
            project_path: 프로젝트 경로
            progress_callback: 진행상황 콜백

        Returns:
            생성된 파일 목록
        """
        self.current_phase = 3
        self._report_progress(
            "Phase 3: Installing resources...", progress_callback
        )

        # TemplateProcessor를 통해 리소스 복사 (silent 모드)
        from moai_adk.core.template import TemplateProcessor

        processor = TemplateProcessor(project_path)
        processor.copy_templates(backup=False, silent=True)  # Progress Bar 충돌 방지

        # 생성된 파일 목록 반환 (간소화)
        return [
            ".claude/",
            ".moai/",
            "CLAUDE.md",
            ".gitignore",
        ]

    def execute_configuration_phase(
        self,
        project_path: Path,
        config: dict[str, str],
        progress_callback: ProgressCallback | None = None,
    ) -> list[str]:
        """Phase 4: 설정 생성

        Args:
            project_path: 프로젝트 경로
            config: 설정 딕셔너리
            progress_callback: 진행상황 콜백

        Returns:
            생성된 파일 목록
        """
        self.current_phase = 4
        self._report_progress(
            "Phase 4: Generating configurations...", progress_callback
        )

        import json
        from moai_adk import __version__

        # config에 버전 정보 추가 (v0.3.1+)
        config["moai_adk_version"] = __version__
        config["optimized"] = False  # 초기값

        # config.json 생성
        config_path = project_path / ".moai" / "config.json"
        with open(config_path, "w", encoding="utf-8") as f:
            json.dump(config, f, indent=2, ensure_ascii=False)

        return [str(config_path)]

    def execute_validation_phase(
        self,
        project_path: Path,
        mode: str = "personal",
        progress_callback: ProgressCallback | None = None,
    ) -> None:
        """Phase 5: 검증 및 마무리

        Args:
            project_path: 프로젝트 경로
            mode: 프로젝트 모드 (personal/team)
            progress_callback: 진행상황 콜백
        """
        self.current_phase = 5
        self._report_progress(
            "Phase 5: Validation and finalization...", progress_callback
        )

        # 설치 결과 검증
        self.validator.validate_installation(project_path)

        # Team 모드: Git 초기화
        if mode == "team":
            self._initialize_git(project_path)

    def _create_backup(self, project_path: Path) -> None:
        """백업 생성 (Selective) - v0.3.0

        Args:
            project_path: 프로젝트 경로
        """
        # 백업 디렉토리 생성 (v0.3.0: .moai/backups/{timestamp}/)
        timestamp = generate_backup_dir_name()
        backup_path = project_path / ".moai" / "backups" / timestamp
        backup_path.mkdir(parents=True, exist_ok=True)

        # 백업 대상 가져오기
        targets = get_backup_targets(project_path)
        backed_up_files: list[str] = []

        # 백업 실행
        for target in targets:
            src_path = project_path / target
            dst_path = backup_path / target

            if src_path.is_dir():
                self._copy_directory_selective(src_path, dst_path)
                backed_up_files.append(f"{target}/")
            else:
                dst_path.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(src_path, dst_path)
                backed_up_files.append(target)

        # Progress Bar 충돌 방지를 위해 메시지 제거
        # (Phase 1 진행상황 메시지로 충분)

    def _copy_directory_selective(self, src: Path, dst: Path) -> None:
        """보호 경로를 제외하고 디렉토리 복사

        Args:
            src: 소스 디렉토리
            dst: 대상 디렉토리
        """
        dst.mkdir(parents=True, exist_ok=True)

        for item in src.rglob("*"):
            rel_path = item.relative_to(src)

            # 보호 경로 제외
            if is_protected_path(rel_path):
                continue

            dst_item = dst / rel_path
            if item.is_file():
                dst_item.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(item, dst_item)
            elif item.is_dir():
                dst_item.mkdir(parents=True, exist_ok=True)

    def _initialize_git(self, project_path: Path) -> None:
        """Git 저장소 초기화

        Args:
            project_path: 프로젝트 경로
        """
        try:
            subprocess.run(
                ["git", "init"],
                cwd=project_path,
                check=True,
                capture_output=True,
            )
            # Progress Bar 충돌 방지를 위해 메시지 제거
        except subprocess.CalledProcessError:
            # 에러 발생 시에만 로깅 (실패해도 치명적이지 않음)
            pass

    def _report_progress(
        self, message: str, callback: ProgressCallback | None
    ) -> None:
        """진행상황 보고

        Args:
            message: 진행 메시지
            callback: 콜백 함수
        """
        if callback:
            callback(message, self.current_phase, self.total_phases)
        # callback이 없으면 Progress Bar가 없다는 뜻이므로 출력하지 않음
        # (CLI 외부에서 사용 시에는 result만 확인하면 됨)
