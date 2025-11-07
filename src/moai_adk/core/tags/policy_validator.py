#!/usr/bin/env python3
# @CODE:TAG-POLICY-VALIDATOR-001 | @SPEC:TAG-POLICY-001
"""TAG 정책 위반 실시간 검증 시스템

MoAI-ADK의 SPEC-first 원칙을 강제하는 실시간 TAG 정책 검증기.
Pre-Tool-Use 훅과 통합하여 SPEC-less 코드 생성을 원천적으로 차단.

주요 기능:
- 실시간 TAG 정책 위반 탐지
- SPEC 없이 CODE 생성 시 차단
- TAG 체인 무결성 검증
- 즉각적인 위반 보고 및 수정 가이드

@SPEC:TAG-POLICY-001
"""

import json
import re
import time
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Set

from moai_adk.core.tags.language_dirs import (
    detect_directories,
    get_exclude_patterns,
    is_code_directory,
)


class PolicyViolationLevel(Enum):
    """정책 위반 수준

    CRITICAL: 작업을 즉시 중단해야 하는 치명적 위반
    HIGH: 사용자 확인 후 진행 가능한 높은 수준 위반
    MEDIUM: 경고 수준 위반 (권고 사항)
    LOW: 정보 수준 (권장 사항)
    """
    CRITICAL = "critical"
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


class PolicyViolationType(Enum):
    """정책 위반 유형

    SPECLESS_CODE: SPEC 없이 CODE 생성 (치명적)
    MISSING_TAGS: 필수 TAG 누락 (높음)
    CHAIN_BREAK: TAG 체인 연결 끊김 (높음)
    DUPLICATE_TAGS: 중복 TAG (중간)
    FORMAT_INVALID: TAG 형식 오류 (중간)
    NO_SPEC_REFERENCE: CODE가 SPEC 참조 없음 (낮음)
    """
    SPECLESS_CODE = "specless_code"
    MISSING_TAGS = "missing_tags"
    CHAIN_BREAK = "chain_break"
    DUPLICATE_TAGS = "duplicate_tags"
    FORMAT_INVALID = "format_invalid"
    NO_SPEC_REFERENCE = "no_spec_reference"


@dataclass
class PolicyViolation:
    """TAG 정책 위반 정보

    Attributes:
        level: 위반 수준 (CRITICAL|HIGH|MEDIUM|LOW)
        type: 위반 유형 (PolicyViolationType)
        tag: 관련 TAG (있는 경우)
        message: 위반 설명
        file_path: 관련 파일 경로
        action: 제안되는 조치 (block|warn|suggest)
        guidance: 수정 안내
        auto_fix_possible: 자동 수정 가능 여부
    """
    level: PolicyViolationLevel
    type: PolicyViolationType
    tag: Optional[str]
    message: str
    file_path: Optional[str]
    action: str  # block|warn|suggest
    guidance: str
    auto_fix_possible: bool = False

    def to_dict(self) -> Dict[str, Any]:
        """딕셔너리로 변환"""
        return {
            "level": self.level.value,
            "type": self.type.value,
            "tag": self.tag,
            "message": self.message,
            "file_path": self.file_path,
            "action": self.action,
            "guidance": self.guidance,
            "auto_fix_possible": self.auto_fix_possible
        }

    def should_block_operation(self) -> bool:
        """작업을 차단해야 하는지 여부"""
        return self.level == PolicyViolationLevel.CRITICAL or self.action == "block"


@dataclass
class PolicyValidationConfig:
    """TAG 정책 검증 설정

    Attributes:
        strict_mode: 엄격 모드 (모든 위반을 차단)
        require_spec_before_code: CODE 생성 전 SPEC 필수
        require_test_for_code: CODE에 TEST 필수
        allow_duplicate_tags: 중복 TAG 허용 여부
        validation_timeout: 검증 타임아웃 (초)
        auto_fix_enabled: 자동 수정 기능 활성화
        file_types_to_validate: 검증할 파일 확장자
    """
    strict_mode: bool = True
    require_spec_before_code: bool = True
    require_test_for_code: bool = True
    allow_duplicate_tags: bool = False
    validation_timeout: int = 5
    auto_fix_enabled: bool = False
    file_types_to_validate: Set[str] = field(default_factory=lambda: {
        "py", "js", "ts", "jsx", "tsx", "md", "txt", "yml", "yaml"
    })


class TagPolicyValidator:
    """TAG 정책 실시간 검증기

    Pre-Tool-Use 훅과 통합하여 파일 생성/수정 시점에 TAG 정책 위반을 탐지하고 차단.
    SPEC-first 원칙을 강제하여 품질 보증.

    Usage:
        config = PolicyValidationConfig(strict_mode=True)
        validator = TagPolicyValidator(config=config)

        # 파일 생성 전 검증
        violations = validator.validate_before_creation(
            file_path="src/example.py",
            content="def example(): pass"
        )

        # 작업 차단 여부 확인
        should_block = any(v.should_block_operation() for v in violations)
    """

    # TAG 정규식 패턴
    TAG_PATTERN = re.compile(r"@(SPEC|CODE|TEST|DOC):([A-Z0-9-]+-\d{3})")

    def __init__(self, config: Optional[PolicyValidationConfig] = None, project_config: Optional[Dict] = None):
        """초기화

        Args:
            config: 정책 검증 설정 (기본: PolicyValidationConfig())
            project_config: 프로젝트 설정 (.moai/config.json에서 로드됨, 선택적)
        """
        self.config = config or PolicyValidationConfig()
        self.project_config = project_config or self._load_project_config()
        self.code_directories = detect_directories(self.project_config)
        self.exclude_patterns = get_exclude_patterns(self.project_config)
        self._start_time = time.time()

    def validate_before_creation(self, file_path: str, content: str) -> List[PolicyViolation]:
        """파일 생성 전 TAG 정책 검증

        Pre-Tool-Use 훅에서 호출. 파일 생성 시점에 정책 위반을 탐지.

        Args:
            file_path: 생성/수정할 파일 경로
            content: 파일 내용

        Returns:
            PolicyViolation 목록
        """
        violations: List[PolicyViolation] = []

        # 타임아웃 체크
        if time.time() - self._start_time > self.config.validation_timeout:
            return violations

        # 파일 타입 확인
        if not self._should_validate_file(file_path):
            return violations

        # 기존 파일 TAG 추출
        existing_tags = self._extract_tags_from_content(content)

        # 새 파일인지 확인
        is_new_file = not Path(file_path).exists()

        if is_new_file:
            # 새 파일 생성 시 검증
            violations.extend(self._validate_new_file_creation(file_path, existing_tags))
        else:
            # 기존 파일 수정 시 검증
            violations.extend(self._validate_file_modification(file_path, existing_tags))

        return violations

    def validate_after_modification(self, file_path: str, content: str) -> List[PolicyViolation]:
        """파일 수정 후 TAG 정책 검증

        Post-Tool-Use 훅에서 호출. 수정 후 최종 상태 검증.

        Args:
            file_path: 수정된 파일 경로
            content: 수정된 파일 내용

        Returns:
            PolicyViolation 목록 (주로 경고 수준)
        """
        violations: List[PolicyViolation] = []

        # 타임아웃 체크
        if time.time() - self._start_time > self.config.validation_timeout:
            return violations

        # 파일 타입 확인
        if not self._should_validate_file(file_path):
            return violations

        # TAG 추출
        tags = self._extract_tags_from_content(content)

        # 누락된 TAG 확인
        violations.extend(self._validate_missing_tags(file_path, tags))

        # 체인 무결성 검증
        violations.extend(self._validate_chain_integrity(file_path, tags))

        return violations

    def _should_validate_file(self, file_path: str) -> bool:
        """파일을 검증해야 하는지 확인

        Args:
            file_path: 파일 경로

        Returns:
            검증 대상이면 True
        """
        path = Path(file_path)
        suffix = path.suffix.lstrip(".")

        # 파일 확장자 확인
        if suffix not in self.config.file_types_to_validate:
            return False

        # 선택적 파일 패턴 제외 (TAG 검증 대상 아님)
        optional_patterns = [
            "CLAUDE.md",
            "README.md",
            "CHANGELOG.md",
            "CONTRIBUTING.md",
            ".claude/",
            ".moai/docs/",
            ".moai/reports/",
            ".moai/analysis/",
            "docs/",
            "templates/",
            "examples/",
        ]

        file_path_str = str(path)
        if any(pattern in file_path_str for pattern in optional_patterns):
            return False

        return True

    def _extract_tags_from_content(self, content: str) -> Dict[str, List[str]]:
        """내용에서 TAG 추출

        Args:
            content: 파일 내용

        Returns:
            {tag_type: [domains]} 딕셔너리
        """
        tags: Dict[str, List[str]] = {
            "SPEC": [], "CODE": [], "TEST": [], "DOC": []
        }

        matches = self.TAG_PATTERN.findall(content)
        for tag_type, domain in matches:
            tags[tag_type].append(domain)

        return tags

    def _validate_new_file_creation(self, file_path: str, tags: Dict[str, List[str]]) -> List[PolicyViolation]:
        """새 파일 생성 시 정책 검증

        Args:
            file_path: 파일 경로
            tags: 추출된 TAG

        Returns:
            PolicyViolation 목록
        """
        violations = []

        # CODE 파일 생성 시 SPEC 필수 확인
        if self._is_code_file(file_path) and self.config.require_spec_before_code:
            if not tags.get("CODE"):
                violations.append(PolicyViolation(
                    level=PolicyViolationLevel.CRITICAL,
                    type=PolicyViolationType.SPECLESS_CODE,
                    tag=None,
                    message="CODE 파일에 @TAG가 없습니다",
                    file_path=file_path,
                    action="block",
                    guidance="CODE 파일은 반드시 @CODE:DOMAIN-XXX 형식의 TAG를 가져야 합니다. 먼저 SPEC을 생성하세요.",
                    auto_fix_possible=False
                ))
            else:
                # CODE TAG가 있는 경우 연결된 SPEC 확인
                for domain in tags["CODE"]:
                    spec_file = self._find_spec_file(domain)
                    if not spec_file:
                        spec_path = f".moai/specs/SPEC-{domain}/spec.md"
                        guidance = f"{spec_path} 파일을 생성하거나 기존 SPEC에 추가하세요."
                        violations.append(PolicyViolation(
                            level=PolicyViolationLevel.HIGH,
                            type=PolicyViolationType.NO_SPEC_REFERENCE,
                            tag=f"@CODE:{domain}",
                            message=f"@CODE:{domain}에 연결된 SPEC이 없습니다",
                            file_path=file_path,
                            action="block" if self.config.strict_mode else "warn",
                            guidance=guidance,
                            auto_fix_possible=True
                        ))

        # TEST 파일 생성 시 CODE 필수 확인
        if self._is_test_file(file_path) and tags.get("TEST"):
            for domain in tags["TEST"]:
                code_file = self._find_code_file(domain)
                if not code_file:
                    violations.append(PolicyViolation(
                        level=PolicyViolationLevel.HIGH,
                        type=PolicyViolationType.CHAIN_BREAK,
                        tag=f"@TEST:{domain}",
                        message=f"@TEST:{domain}에 연결된 CODE가 없습니다",
                        file_path=file_path,
                        action="warn",
                        guidance=f"먼저 @CODE:{domain}를 가진 구현 파일을 생성하세요.",
                        auto_fix_possible=False
                    ))

        return violations

    def _validate_file_modification(self, file_path: str, tags: Dict[str, List[str]]) -> List[PolicyViolation]:
        """파일 수정 시 정책 검증

        Args:
            file_path: 파일 경로
            tags: 추출된 TAG

        Returns:
            PolicyViolation 목록
        """
        violations = []

        # 중복 TAG 확인
        if not self.config.allow_duplicate_tags:
            duplicates = self._find_duplicate_tags(file_path, tags)
            for duplicate in duplicates:
                violations.append(PolicyViolation(
                    level=PolicyViolationLevel.MEDIUM,
                    type=PolicyViolationType.DUPLICATE_TAGS,
                    tag=duplicate,
                    message=f"중복된 TAG: {duplicate}",
                    file_path=file_path,
                    action="warn",
                    guidance="중복된 TAG를 제거하세요. 각 TAG는 고유해야 합니다.",
                    auto_fix_possible=True
                ))

        return violations

    def _validate_missing_tags(self, file_path: str, tags: Dict[str, List[str]]) -> List[PolicyViolation]:
        """누락된 TAG 확인

        Args:
            file_path: 파일 경로
            tags: 추출된 TAG

        Returns:
            PolicyViolation 목록
        """
        violations = []

        # CODE 파일인데 TAG가 없는 경우
        if self._is_code_file(file_path) and not tags.get("CODE"):
            violations.append(PolicyViolation(
                level=PolicyViolationLevel.HIGH,
                type=PolicyViolationType.MISSING_TAGS,
                tag=None,
                message="CODE 파일에 @TAG가 누락되었습니다",
                file_path=file_path,
                action="suggest",
                guidance="파일 상단에 @CODE:DOMAIN-XXX 형식의 TAG를 추가하세요.",
                auto_fix_possible=True
            ))

        return violations

    def _validate_chain_integrity(self, file_path: str, tags: Dict[str, List[str]]) -> List[PolicyViolation]:
        """TAG 체인 무결성 검증

        Args:
            file_path: 파일 경로
            tags: 추출된 TAG

        Returns:
            PolicyViolation 목록
        """
        violations = []

        # CODE가 있는데 TEST가 없는 경우
        if tags.get("CODE") and self.config.require_test_for_code:
            for domain in tags["CODE"]:
                test_file = self._find_test_file(domain)
                if not test_file:
                    violations.append(PolicyViolation(
                        level=PolicyViolationLevel.MEDIUM,
                        type=PolicyViolationType.CHAIN_BREAK,
                        tag=f"@CODE:{domain}",
                        message=f"@CODE:{domain}에 연결된 TEST가 없습니다",
                        file_path=file_path,
                        action="suggest",
                        guidance=f"tests/ 디렉토리에 @TEST:{domain}를 가진 테스트 파일을 생성하세요.",
                        auto_fix_possible=True
                    ))

        return violations

    def _is_code_file(self, file_path: str) -> bool:
        """코드 파일인지 확인 (언어별 동적 감지)

        Args:
            file_path: 파일 경로

        Returns:
            코드 파일이면 True
        """
        path = Path(file_path)

        # 파일 확장자 확인 (코드 파일 확장자만)
        code_extensions = {".py", ".js", ".ts", ".jsx", ".tsx", ".go", ".rs", ".kt", ".rb", ".php", ".java", ".cs"}
        if path.suffix not in code_extensions:
            return False

        # language_dirs를 사용한 동적 감지
        return is_code_directory(path, self.project_config)

    def _is_test_file(self, file_path: str) -> bool:
        """테스트 파일인지 확인

        Args:
            file_path: 파일 경로

        Returns:
            테스트 파일이면 True
        """
        path = Path(file_path)
        test_patterns = ["test/", "tests/", "__tests__", "spec/", "_test.", "_spec."]
        return any(pattern in str(path) for pattern in test_patterns)

    def _find_spec_file(self, domain: str) -> Optional[Path]:
        """DOMAIN에 해당하는 SPEC 파일 찾기

        Args:
            domain: TAG 도메인 (예: USER-REG-001)

        Returns:
            SPEC 파일 경로 또는 None
        """
        spec_patterns = [
            f".moai/specs/SPEC-{domain}/spec.md",
            f".moai/specs/SPEC-{domain}.md",
            f"specs/SPEC-{domain}.md"
        ]

        for pattern in spec_patterns:
            path = Path(pattern)
            if path.exists():
                return path

        return None

    def _find_code_file(self, domain: str) -> Optional[Path]:
        """DOMAIN에 해당하는 CODE 파일 찾기

        Args:
            domain: TAG 도메인

        Returns:
            CODE 파일 경로 또는 None
        """
        # 프로젝트 루트에서 CODE TAG 검색
        for pattern in ["src/**/*.py", "lib/**/*.py", "**/*.py", "**/*.js", "**/*.ts"]:
            for path in Path(".").glob(pattern):
                if path.is_file():
                    try:
                        content = path.read_text(encoding="utf-8", errors="ignore")
                        if f"@CODE:{domain}" in content:
                            return path
                    except Exception:
                        continue

        return None

    def _find_test_file(self, domain: str) -> Optional[Path]:
        """DOMAIN에 해당하는 TEST 파일 찾기

        Args:
            domain: TAG 도메인

        Returns:
            TEST 파일 경로 또는 None
        """
        test_patterns = [
            f"tests/**/test_*{domain}*.py",
            f"test/**/test_*{domain}*.py",
            f"tests/**/*{domain}*_test.py",
            f"**/*test*{domain}*.py"
        ]

        for pattern in test_patterns:
            for path in Path(".").glob(pattern):
                if path.is_file():
                    try:
                        content = path.read_text(encoding="utf-8", errors="ignore")
                        if f"@TEST:{domain}" in content:
                            return path
                    except Exception:
                        continue

        return None

    def _find_duplicate_tags(self, file_path: str, tags: Dict[str, List[str]]) -> List[str]:
        """파일 내 중복 TAG 찾기

        Args:
            file_path: 파일 경로
            tags: 추출된 TAG

        Returns:
            중복 TAG 목록
        """
        duplicates: List[str] = []

        try:
            content = Path(file_path).read_text(encoding="utf-8", errors="ignore")

            for tag_type, domains in tags.items():
                for domain in domains:
                    tag = f"@{tag_type}:{domain}"
                    count = content.count(tag)
                    if count > 1:
                        duplicates.append(tag)

        except Exception:
            pass

        return duplicates

    def create_validation_report(self, violations: List[PolicyViolation]) -> str:
        """검증 결과 리포트 생성

        Args:
            violations: 정책 위반 목록

        Returns:
            포맷된 리포트 문자열
        """
        if not violations:
            return "✅ TAG 정책 검증 통과"

        lines = []
        lines.append("❌ TAG 정책 위반 발견")
        lines.append("=" * 50)

        # 수준별 그룹화
        by_level: Dict[PolicyViolationLevel, List[PolicyViolation]] = {
            PolicyViolationLevel.CRITICAL: [],
            PolicyViolationLevel.HIGH: [],
            PolicyViolationLevel.MEDIUM: [],
            PolicyViolationLevel.LOW: []
        }

        for violation in violations:
            by_level[violation.level].append(violation)

        # 수준별 출력
        level_names = {
            PolicyViolationLevel.CRITICAL: "🚨 치명적",
            PolicyViolationLevel.HIGH: "⚠️ 높음",
            PolicyViolationLevel.MEDIUM: "⚡ 중간",
            PolicyViolationLevel.LOW: "ℹ️ 낮음"
        }

        for level in [PolicyViolationLevel.CRITICAL, PolicyViolationLevel.HIGH,
                     PolicyViolationLevel.MEDIUM, PolicyViolationLevel.LOW]:
            level_violations = by_level[level]
            if level_violations:
                lines.append(f"\n{level_names[level]} ({len(level_violations)}개):")
                lines.append("-" * 30)

                for violation in level_violations:
                    tag_info = f" - {violation.tag}" if violation.tag else ""
                    lines.append(f"  {violation.message}{tag_info}")
                    if violation.file_path:
                        lines.append(f"    파일: {violation.file_path}")
                    lines.append(f"    조치: {violation.guidance}")
                    if violation.auto_fix_possible:
                        lines.append("    🤖 자동 수정 가능")
                    lines.append("")

        return "\n".join(lines)

    def _load_project_config(self) -> Dict:
        """프로젝트 설정 로드 (.moai/config.json)

        Returns:
            프로젝트 설정 딕셔너리
        """
        config_path = Path(".moai/config.json")
        if config_path.exists():
            try:
                return json.loads(config_path.read_text(encoding="utf-8"))
            except Exception:
                pass

        # 기본 설정 반환
        return {"project": {"language": "python"}}

    def _fix_duplicate_tags(self, content: str) -> str:
        """중복 TAG 제거

        같은 TAG가 여러 번 나타나는 경우 첫 번째만 유지하고 나머지는 제거.

        Args:
            content: 파일 내용

        Returns:
            수정된 내용
        """
        lines = content.split("\n")
        seen_tags = set()
        result_lines = []

        for line in lines:
            # 이 줄에서 모든 TAG 추출
            tags = self.TAG_PATTERN.findall(line)
            modified_line = line

            for tag_type, domain in tags:
                tag = f"@{tag_type}:{domain}"
                if tag in seen_tags:
                    # 이미 본 TAG - 이 줄에서 제거
                    modified_line = modified_line.replace(f"{tag} | ", "")
                    modified_line = modified_line.replace(f" | {tag}", "")
                    modified_line = modified_line.replace(tag, "")
                else:
                    seen_tags.add(tag)

            result_lines.append(modified_line)

        return "\n".join(result_lines)

    def _fix_format_errors(self, content: str) -> str:
        """TAG 형식 오류 수정

        - 콜론 누락: @CODE AUTH-001 → @CODE:AUTH-001
        - 공백 정규화: @CODE:AUTH-001  |  @SPEC:... → @CODE:AUTH-001 | @SPEC:...

        Args:
            content: 파일 내용

        Returns:
            수정된 내용
        """
        # 콜론 누락 수정 (예: @CODE AUTH-001 → @CODE:AUTH-001)
        content = re.sub(r"@(SPEC|CODE|TEST|DOC)\s+([A-Z0-9-]+-\d{3})", r"@\1:\2", content)

        # 공백 정규화 (파이프 주변)
        content = re.sub(r"\s*\|\s*", " | ", content)

        # 중복 공백 제거
        content = re.sub(r"  +", " ", content)

        return content

    def _apply_auto_fix(self, file_path: str, violations: List[PolicyViolation]) -> Dict[str, Any]:
        """자동 수정 적용

        설정에 따라 SAFE 수준의 위반을 자동으로 수정.

        Args:
            file_path: 파일 경로
            violations: 정책 위반 목록

        Returns:
            수정 결과 딕셔너리
        """
        result: Dict[str, Any] = {
            "success": False,
            "fixed_count": 0,
            "pending_count": 0,
            "fixed_violations": [],
            "pending_violations": []
        }

        try:
            content = Path(file_path).read_text(encoding="utf-8")
            modified = False

            for violation in violations:
                if violation.type == PolicyViolationType.DUPLICATE_TAGS:
                    content = self._fix_duplicate_tags(content)
                    result["fixed_count"] += 1
                    result["fixed_violations"].append(violation)
                    modified = True

                elif violation.type == PolicyViolationType.FORMAT_INVALID:
                    content = self._fix_format_errors(content)
                    result["fixed_count"] += 1
                    result["fixed_violations"].append(violation)
                    modified = True

                else:
                    # 수정 불가능한 위반
                    result["pending_count"] += 1
                    result["pending_violations"].append(violation)

            # 수정된 내용 저장
            if modified:
                Path(file_path).write_text(content, encoding="utf-8")
                result["success"] = True

        except Exception as e:
            result["error"] = str(e)

        return result
