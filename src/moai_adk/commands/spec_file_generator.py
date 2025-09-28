"""
@TASK:FILE-GENERATOR-001 SpecFileGenerator Module
@REQ:TRUST-COMPLIANCE-001 → @DESIGN:MODULE-SPLIT-001 → @TASK:FILE-GENERATOR-001

SPEC file generation module following TRUST principles:
- T: Test-driven file operations
- R: Readable template generation
- U: Unified file creation responsibility
- S: Secure file operations with validation
- T: Trackable file creation process
"""

import datetime
import logging
from pathlib import Path

# Logging setup
logger = logging.getLogger(__name__)


class SpecFileGenerator:
    """
    @TASK:SPEC-FILE-CREATE-001 SPEC file generation with security and validation

    TRUST principles applied:
    - T: Testable file operations
    - R: Clear template structure
    - U: Single responsibility (file generation only)
    - S: Secure file operations with backup
    - T: Trackable file creation process
    """

    def __init__(self, project_dir: Path, current_mode: str = "personal"):
        """Initialize SpecFileGenerator

        Args:
            project_dir: Project directory path
            current_mode: Current operation mode (personal/team)
        """
        if not isinstance(project_dir, Path):
            raise ValueError(f"project_dir must be a Path object: {type(project_dir)}")

        self.project_dir = project_dir.resolve()
        self.current_mode = current_mode

        logger.debug(f"SpecFileGenerator initialized: {self.project_dir}, mode={current_mode}")

    def create_spec_file(self, spec_name: str, description: str) -> Path:
        """Create SPEC file with given name and description

        Args:
            spec_name: Name of the specification
            description: Description of the specification

        Returns:
            Path to the created SPEC file

        Raises:
            ValueError: If file creation fails
        """
        try:
            # Ensure specs directory exists
            specs_dir = self.project_dir / ".moai" / "specs"
            specs_dir.mkdir(parents=True, exist_ok=True)

            # Define spec file path
            spec_file = specs_dir / f"{spec_name}.md"

            # Backup existing file if it exists
            if spec_file.exists():
                backup_file = specs_dir / f"{spec_name}.md.backup"
                spec_file.replace(backup_file)
                logger.info(f"기존 SPEC 파일 백업: {backup_file}")

            # Generate content
            spec_content = self.generate_spec_content(spec_name, description)

            # Write file with UTF-8 encoding
            spec_file.write_text(spec_content, encoding="utf-8")

            logger.info(f"SPEC 파일 생성 완료: {spec_file}")
            return spec_file

        except OSError as e:
            logger.error(f"SPEC 파일 생성 실패: {e}")
            raise ValueError(f"SPEC 파일 생성 실패: {e}")

    def generate_spec_content(self, spec_name: str, description: str) -> str:
        """Generate SPEC file content with template

        Args:
            spec_name: Name of the specification
            description: Description of the specification

        Returns:
            Generated SPEC content
        """
        timestamp = self._get_current_timestamp()

        return f"""# {spec_name}

## 개요

{description}

## 요구사항

### 기능 요구사항

- [ ] 핵심 기능 구현
- [ ] 사용자 인터페이스 개발
- [ ] 데이터 처리 로직 구현

### 비기능 요구사항

- [ ] 성능: 응답 시간 < 1초
- [ ] 보안: 입력 검증 및 인증
- [ ] 안정성: 99% 가용성

## 수락 기준

- [ ] 모든 주요 기능이 정상 동작
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 코드 리뷰 완료

## 태그 체계

@REQ:{spec_name}-001 - 주요 요구사항
@DESIGN:{spec_name}-ARCH-001 - 아키텍처 설계
@TASK:{spec_name}-IMPL-001 - 구현 작업
@TEST:{spec_name}-UNIT-001 - 단위 테스트

---

생성 일시: {timestamp}
모드: {self.current_mode}
"""

    def _get_current_timestamp(self) -> str:
        """Get current timestamp in formatted string

        Returns:
            Formatted current timestamp
        """
        return datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")