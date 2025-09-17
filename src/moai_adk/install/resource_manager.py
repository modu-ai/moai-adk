"""
MoAI-ADK Resource Manager

패키지 내장 리소스를 관리하는 모듈입니다.
심볼릭 링크 대신 패키지에서 직접 리소스를 복사하여 관리합니다.
"""

import shutil
from pathlib import Path
from string import Template as StrTemplate
from typing import Optional, List, Callable, Dict
from importlib import resources
import logging

logger = logging.getLogger(__name__)


MEMORY_STACK_TEMPLATE_MAP: Dict[str, List[str]] = {
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

    def get_template_content(self, template_name: str) -> Optional[str]:
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
                     exclude_subdirs: Optional[List[str]] = None) -> bool:
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

                    ignore: Optional[Callable] = None
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
            return True

        except Exception as e:
            logger.error(f"Failed to copy template {template_name}: {e}")
            return False

    def copy_claude_resources(self, project_path: Path,
                             overwrite: bool = False) -> List[Path]:
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
                copied_files.append(target_path)

        return copied_files

    def copy_moai_resources(self, project_path: Path,
                           overwrite: bool = False,
                           exclude_templates: bool = False) -> List[Path]:
        """
        MoAI 관련 리소스를 프로젝트에 복사

        Args:
            project_path: 프로젝트 경로
            overwrite: 기존 파일 덮어쓰기 여부

        Returns:
            List[Path]: 복사된 파일 경로들
        """
        copied_files = []
        moai_resources = ['.moai']

        for resource in moai_resources:
            target_path = project_path / resource
            if self.copy_template(
                resource,
                target_path,
                overwrite,
                exclude_subdirs=['_templates'] if exclude_templates else None,
            ):
                copied_files.append(target_path)

        return copied_files

    def copy_github_resources(self, project_path: Path,
                             overwrite: bool = False) -> List[Path]:
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
        tech_stack: List[str],
        context: Dict[str, str],
        overwrite: bool = False
    ) -> List[Path]:
        """Copy stack-specific memory templates into the project."""

        copied_files: List[Path] = []
        memory_dir = project_path / ".moai" / "memory"
        memory_dir.mkdir(parents=True, exist_ok=True)

        # Always include common template
        templates_to_copy: List[str] = ["common"]

        for tech in tech_stack:
            tech_key = tech.lower()
            matches = MEMORY_STACK_TEMPLATE_MAP.get(tech_key)
            if matches:
                templates_to_copy.extend(matches)
            else:
                logger.debug("No memory template mapping for tech stack item: %s", tech)

        # Remove duplicates while preserving order
        seen: set[str] = set()
        unique_templates: List[str] = []
        for template_name in templates_to_copy:
            if template_name not in seen:
                seen.add(template_name)
                unique_templates.append(template_name)

        for template_name in unique_templates:
            template_rel_path = f".moai/_templates/memory/{template_name}.template.md"
            target_path = memory_dir / f"{template_name}.md"

            if not self._render_template_with_context(
                template_rel_path,
                target_path,
                context,
                overwrite
            ):
                logger.warning("Failed to render memory template: %s", template_name)
            else:
                copied_files.append(target_path)

        return copied_files

    def _render_template_with_context(
        self,
        template_name: str,
        target_path: Path,
        context: Dict[str, str],
        overwrite: bool = False
    ) -> bool:
        """Render a text template with context variables to a target file."""

        try:
            content = self.get_template_content(template_name)
            if content is None:
                logger.warning("Template content not found: %s", template_name)
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

    def list_templates(self) -> List[str]:
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
