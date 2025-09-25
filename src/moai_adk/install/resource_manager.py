"""
@FEATURE:RESOURCE-001 MoAI-ADK Resource Manager

@TASK:RESOURCE-001 패키지 내장 리소스를 관리하는 모듈입니다.
@TASK:RESOURCE-002 심볼릭 링크 대신 패키지에서 직접 리소스를 복사하여 관리합니다.
"""

import logging
import shutil
from collections.abc import Callable
from importlib import resources
from pathlib import Path
from string import Template as StrTemplate

logger = logging.getLogger(__name__)


MEMORY_STACK_TEMPLATE_MAP: dict[str, list[str]] = {
    "python": ["backend-python"],
    "fastapi": ["backend-python", "backend-fastapi"],
    "django": ["backend-python"],
    "flask": ["backend-python"],
    "java": ["backend-spring"],
    "spring": ["backend-spring"],
    "spring boot": ["backend-spring"],
    "springboot": ["backend-spring"],
    "spring-boot": ["backend-spring"],
    "react": ["frontend-react"],
    "nextjs": ["frontend-react", "frontend-next"],
    "vue": ["frontend-vue"],
    "nuxt": ["frontend-vue"],
    "angular": ["frontend-angular"],
    "typescript": ["frontend-react"],
    "javascript": ["frontend-react"],
}


class ResourceManager:
    """
    패키지 내장 리소스 관리 클래스

    pip으로 설치된 패키지에서 템플릿과 설정 파일을 관리합니다.
    심볼릭 링크를 사용하지 않고 직접 복사 방식을 사용합니다.
    """

    def __init__(self):
        """
        ResourceManager 초기화
        """
        try:
            # 패키지 리소스 접근
            self.resources_root = resources.files('moai_adk.resources')
            self.templates_root = self.resources_root / 'templates'
            logger.info("ResourceManager initialized")

        except Exception as e:
            logger.error(f"Failed to initialize ResourceManager: {e}")
            raise

    def get_version(self) -> str:
        """패키지 버전 반환"""
        try:
            version_file = self.resources_root / 'VERSION'
            with resources.as_file(version_file) as version_path:
                return version_path.read_text().strip()
        except Exception as e:
            logger.warning(f"Could not read version: {e}")
            return "unknown"

    def get_template_path(self, template_name: str) -> Path:
        """템플릿 경로 반환 (읽기 전용)"""
        return self.templates_root / template_name

    def get_template_content(self, template_name: str) -> str | None:
        """
        템플릿 내용 반환

        Args:
            template_name: 템플릿 파일명

        Returns:
            str: 템플릿 내용 (없으면 None)
        """
        try:
            template_path = self.get_template_path(template_name)
            with resources.as_file(template_path) as actual_path:
                if actual_path.is_file() and actual_path.suffix in ['.md', '.json', '.yml', '.yaml', '.txt']:
                    return actual_path.read_text(encoding='utf-8')
                return None
        except Exception as e:
            logger.warning(f"Failed to read template content {template_name}: {e}")
            return None

    def copy_template(self, template_name: str, target_path: Path,
                     overwrite: bool = False,
                     exclude_subdirs: list[str] | None = None) -> bool:
        """
        템플릿을 대상 경로로 복사

        Args:
            template_name: 복사할 템플릿 이름 (.claude, .moai 등)
            target_path: 복사 대상 경로
            overwrite: 기존 파일 덮어쓰기 여부

        Returns:
            bool: 복사 성공 여부
        """
        try:
            # 경로 안전성 검증
            if not self._validate_safe_path(target_path):
                raise ValueError(f"Unsafe target path detected: {target_path}")

            # 절대 경로로 변환
            target_path = target_path.resolve()
            template_path = self.get_template_path(template_name)

            # 대상 경로가 이미 존재하는 경우
            if target_path.exists():
                if target_path.is_file():
                    # 파일인 경우 기존 로직 유지
                    if not overwrite:
                        logger.info(f"Target file already exists, skipping: {target_path}")
                        return True
                    else:
                        logger.info(f"Overwriting existing file: {target_path}")
                        target_path.unlink()
                else:
                    # 디렉토리인 경우 병합할 수 있도록 처리
                    logger.info(f"Target directory exists, will merge contents: {target_path}")

            # 부모 디렉토리 생성
            target_path.parent.mkdir(parents=True, exist_ok=True)

            # 패키지에서 복사
            with resources.as_file(template_path) as source_path:
                if source_path.is_dir():
                    # 디렉토리 복사 - 기존 디렉토리와 병합 허용
                    def copy_function(src, dst, **kwargs):
                        # 개별 파일 overwrite 정책 적용
                        if Path(dst).exists() and not overwrite:
                            logger.debug(f"File exists, skipping: {dst}")
                            return dst
                        return shutil.copy2(src, dst, **kwargs)

                    ignore: Callable | None = None
                    if exclude_subdirs:
                        ignore = shutil.ignore_patterns(*exclude_subdirs)

                    shutil.copytree(
                        source_path,
                        target_path,
                        dirs_exist_ok=True,
                        copy_function=copy_function if not overwrite else shutil.copy2,
                        ignore=ignore,
                    )
                else:
                    shutil.copy2(source_path, target_path)

            logger.info(f"Successfully copied {template_name} to {target_path}")

            # 후처리: .claude/hooks/moai/*.py 실행 권한 보장
            try:
                if target_path.is_dir():
                    # template_name이 '.claude'이거나 타겟 경로가 .claude 루트인 경우에만 후처리
                    if template_name.endswith('.claude') or target_path.name == '.claude':
                        self._ensure_hook_permissions(target_path)
                else:
                    # 개별 파일 복사 케이스: .claude 내 파일인 경우 상위에서 처리
                    pass
            except Exception as perm_exc:
                logger.warning(f"Failed to set hook permissions under {target_path}: {perm_exc}")
            return True

        except Exception as e:
            logger.error(f"Failed to copy template {template_name}: {e}")
            return False

    def _validate_safe_path(self, target_path: Path) -> bool:
        """
        경로 안전성 검증

        Args:
            target_path: 검증할 경로

        Returns:
            bool: 안전한 경로 여부
        """
        try:
            resolved_path = target_path.resolve()

            # 경로 순회 공격 방지 (.., 심볼릭 링크 등)
            if ".." in str(resolved_path):
                logger.warning(f"Path traversal detected in: {target_path}")
                return False

            # 시스템 중요 디렉토리 보호
            dangerous_paths = ["/etc", "/usr/bin", "/usr/sbin", "/var", "/boot", "/sys", "/proc"]
            for dangerous in dangerous_paths:
                if str(resolved_path).startswith(dangerous):
                    logger.warning(f"Attempt to write to dangerous system path: {target_path}")
                    return False

            return True

        except Exception as e:
            logger.warning(f"Path validation failed for {target_path}: {e}")
            return False

    def copy_claude_resources(self, project_path: Path,
                             overwrite: bool = False) -> list[Path]:
        """
        Claude Code 관련 리소스를 프로젝트에 복사

        Args:
            project_path: 프로젝트 경로
            overwrite: 기존 파일 덮어쓰기 여부

        Returns:
            List[Path]: 복사된 파일 경로들
        """
        copied_files = []
        claude_resources = ['.claude']

        for resource in claude_resources:
            target_path = project_path / resource
            if self.copy_template(resource, target_path, overwrite):
                # 실행 권한 보장 (이중 안전장치)
                try:
                    self._ensure_hook_permissions(target_path)
                except Exception as e:
                    logger.warning(f"Hook permission ensure skipped: {e}")
                copied_files.append(target_path)

        return copied_files

    def _ensure_hook_permissions(self, claude_root: Path) -> None:
        """Ensure executable permissions for hook python files.

        대상: {claude_root}/hooks/moai/*.py
        Windows에서는 무시(권한 비트 미사용)되지만 호출 자체는 안전합니다.
        """
        try:
            hooks_dir = claude_root / 'hooks' / 'moai'
            if not hooks_dir.exists() or not hooks_dir.is_dir():
                return
            for py_file in hooks_dir.glob('*.py'):
                try:
                    # 0o755: 소유자 실행/읽기/쓰기, 그룹/기타 실행/읽기
                    py_file.chmod(0o755)
                    logger.debug("Set executable permission: %s", py_file)
                except Exception as file_exc:
                    logger.warning("Failed to chmod +x %s: %s", py_file, file_exc)
        except Exception as exc:
            logger.warning("Failed to ensure hook permissions under %s: %s", claude_root, exc)

    def copy_moai_resources(self, project_path: Path,
                           overwrite: bool = False,
                           exclude_templates: bool = False,
                           project_context: dict[str, str] | None = None) -> list[Path]:
        """
        MoAI 관련 리소스를 프로젝트에 복사

        Args:
            project_path: 프로젝트 경로
            overwrite: 기존 파일 덮어쓰기 여부
            exclude_templates: 템플릿 디렉토리 제외 여부
            project_context: 템플릿 변수 치환용 컨텍스트

        Returns:
            List[Path]: 복사된 파일 경로들
        """
        copied_files = []
        moai_resources = ['.moai']

        # @TASK:TEMPLATE-VERIFY-001 Clean template validation
        logger.info("Starting MoAI resources installation...")

        for resource in moai_resources:
            target_path = project_path / resource
            logger.info(f"Installing {resource} to {target_path}")

            if self.copy_template(
                resource,
                target_path,
                overwrite,
                exclude_subdirs=['_templates'] if exclude_templates else None,
            ):
                copied_files.append(target_path)
                self._validate_clean_installation(target_path)
                logger.info(f"Successfully installed {resource}")
            else:
                logger.error(f"Failed to install {resource}")

        # 템플릿 변수 치환 수행
        if project_context and copied_files:
            self._apply_project_context(copied_files, project_context)

        logger.info(f"MoAI resources installation completed. {len(copied_files)} resources installed.")
        return copied_files

    def copy_github_resources(self, project_path: Path,
                             overwrite: bool = False) -> list[Path]:
        """
        GitHub 워크플로우 리소스를 프로젝트에 복사

        Args:
            project_path: 프로젝트 경로
            overwrite: 기존 파일 덮어쓰기 여부

        Returns:
            List[Path]: 복사된 파일 경로들
        """
        copied_files = []
        github_resources = ['.github']

        for resource in github_resources:
            target_path = project_path / resource
            if self.copy_template(resource, target_path, overwrite):
                copied_files.append(target_path)

        return copied_files

    def copy_project_memory(self, project_path: Path,
                           overwrite: bool = False) -> bool:
        """
        프로젝트 메모리 파일(CLAUDE.md) 생성

        Args:
            project_path: 프로젝트 경로
            overwrite: 기존 파일 덮어쓰기 여부

        Returns:
            bool: 생성 성공 여부
        """
        try:
            target_path = project_path / "CLAUDE.md"
            return self.copy_template("CLAUDE.md", target_path, overwrite)
        except Exception as e:
            logger.error(f"Failed to copy project memory: {e}")
            return False

    def copy_memory_templates(
        self,
        project_path: Path,
        tech_stack: list[str],
        context: dict[str, str],
        overwrite: bool = False
    ) -> list[Path]:
        """Copy stack-specific memory templates into the project."""

        copied_files: list[Path] = []
        memory_dir = project_path / ".moai" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)

        # Only copy templates that actually exist in the package
        templates_to_copy: list[str] = ["development-guide"]

        for tech in tech_stack:
            tech_key = tech.lower()
            matches = MEMORY_STACK_TEMPLATE_MAP.get(tech_key)
            if matches:
                templates_to_copy.extend(matches)
            else:
                logger.debug("No memory template mapping for tech stack item: %s", tech)

        # Remove duplicates while preserving order
        seen: set[str] = set()
        unique_templates: list[str] = []
        for template_name in templates_to_copy:
            if template_name not in seen:
                seen.add(template_name)
                unique_templates.append(template_name)

        failed_templates = []

        for template_name in unique_templates:
            # Use actual package file structure
            template_rel_path = f".moai/memory/{template_name}.md"
            target_path = memory_dir / f"{template_name}.md"

            if not self._render_template_with_context(
                template_rel_path,
                target_path,
                context,
                overwrite
            ):
                logger.error("Failed to render memory template: %s", template_name)
                failed_templates.append(template_name)
            else:
                copied_files.append(target_path)

        # Report any template failures to user
        if failed_templates:
            logger.warning(
                "Failed to copy %d memory templates: %s",
                len(failed_templates),
                ", ".join(failed_templates)
            )

        return copied_files


    def _render_template_with_context(
        self,
        template_name: str,
        target_path: Path,
        context: dict[str, str],
        overwrite: bool = False
    ) -> bool:
        """Render a text template with context variables to a target file."""

        try:
            content = self.get_template_content(template_name)
            if content is None:
                logger.error(
                    "Template file not found: %s. Expected location: %s",
                    template_name,
                    f"src/moai_adk/resources/templates/{template_name}"
                )
                return False

            if target_path.exists() and not overwrite:
                logger.info("Memory file already exists, skipping: %s", target_path)
                return True

            target_path.parent.mkdir(parents=True, exist_ok=True)

            rendered = StrTemplate(content).safe_substitute(context)
            target_path.write_text(rendered, encoding="utf-8")
            logger.info("Rendered memory template %s -> %s", template_name, target_path)
            return True

        except Exception as exc:
            logger.error("Failed to render template %s: %s", template_name, exc)
            return False

    def validate_project_resources(self, project_path: Path) -> bool:
        """
        프로젝트 리소스 검증 (validate_required_resources와 동일)

        Args:
            project_path: 프로젝트 경로

        Returns:
            bool: 검증 성공 여부
        """
        return self.validate_required_resources(project_path)

    def list_templates(self) -> list[str]:
        """사용 가능한 템플릿 목록 반환"""
        templates = []
        try:
            with resources.as_file(self.templates_root) as templates_path:
                if templates_path.exists():
                    templates = [item.name for item in templates_path.iterdir()
                               if item.is_dir() or item.suffix in ['.md', '.json', '.yml', '.yaml']]
            return sorted(templates)
        except Exception as e:
            logger.warning(f"Failed to list templates: {e}")
            return []

    def validate_required_resources(self, project_path: Path) -> bool:
        """필수 리소스가 모두 있는지 확인"""
        required_paths = [
            project_path / '.claude',
            project_path / '.moai',
            project_path / 'CLAUDE.md'
        ]

        missing_paths = [path for path in required_paths if not path.exists()]

        if missing_paths:
            logger.warning(f"Missing required resources: {[str(p) for p in missing_paths]}")
            return False

        logger.info("All required resources are present")
        return True

    def _validate_clean_installation(self, target_path: Path) -> bool:
        """
        @TASK:TEMPLATE-VERIFY-001 설치된 리소스가 깨끗한 초기 상태인지 검증

        Args:
            target_path: 검증할 대상 경로

        Returns:
            bool: 깨끗한 설치 여부
        """
        try:
            # 1. specs 디렉토리가 비어있거나 .gitkeep만 있는지 확인
            specs_dir = target_path / 'specs'
            if specs_dir.exists():
                spec_files = [f for f in specs_dir.iterdir() if f.name != '.gitkeep']
                if spec_files:
                    logger.warning(f"Found unexpected spec files in clean installation: {[f.name for f in spec_files]}")
                    return False

            # 2. tags.json이 초기 구조인지 확인 (약 50줄 미만)
            tags_file = target_path / 'indexes' / 'tags.json'
            if tags_file.exists():
                line_count = sum(1 for _ in open(tags_file))
                if line_count > 50:
                    logger.warning(f"tags.json seems to contain development data: {line_count} lines (expected < 50)")
                    return False

            # 3. reports 디렉토리가 비어있거나 .gitkeep만 있는지 확인
            reports_dir = target_path / 'reports'
            if reports_dir.exists():
                report_files = [f for f in reports_dir.iterdir() if f.name != '.gitkeep']
                if report_files:
                    logger.warning(f"Found unexpected report files in clean installation: {[f.name for f in report_files]}")
                    return False

            logger.debug(f"Clean installation validated successfully: {target_path}")
            return True

        except Exception as e:
            logger.error(f"Failed to validate clean installation at {target_path}: {e}")
            return False

    def _apply_project_context(self, copied_paths: list[Path], project_context: dict[str, str]) -> None:
        """
        복사된 파일들에 프로젝트 컨텍스트 변수를 치환합니다.

        Args:
            copied_paths: 복사된 파일 경로들
            project_context: 치환할 변수들 (예: PROJECT_NAME, PROJECT_DESCRIPTION 등)
        """
        try:
            template_files = [
                'config.json',
                'project/product.md',
                'project/structure.md',
                'project/tech.md'
            ]

            for base_path in copied_paths:
                if not base_path.is_dir():
                    continue

                for template_file in template_files:
                    file_path = base_path / template_file
                    if file_path.exists() and file_path.is_file():
                        self._substitute_template_variables(file_path, project_context)

        except Exception as e:
            logger.warning(f"Failed to apply project context: {e}")

    def _substitute_template_variables(self, file_path: Path, context: dict[str, str]) -> None:
        """
        단일 파일의 템플릿 변수를 치환합니다.

        Args:
            file_path: 치환할 파일 경로
            context: 치환 변수들
        """
        try:
            content = file_path.read_text(encoding='utf-8')

            # 더 안전한 템플릿 치환을 위해 정확한 패턴만 치환
            for key, value in context.items():
                pattern = "{{" + key + "}}"
                content = content.replace(pattern, value)

            file_path.write_text(content, encoding='utf-8')
            logger.debug(f"Template variables substituted in {file_path}")

        except Exception as e:
            logger.warning(f"Failed to substitute template variables in {file_path}: {e}")
