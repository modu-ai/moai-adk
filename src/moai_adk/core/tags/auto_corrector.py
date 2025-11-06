#!/usr/bin/env python3
# @CODE:TAG-AUTO-CORRECTOR-001 | @SPEC:TAG-AUTO-001
"""TAG 오류 자동 수정 시스템

TAG 정책 위반을 자동으로 수정하고 스마트한 TAG를 생성하는 시스템.
Post-Tool-Use 훅과 통합하여 실시간으로 TAG 오류를 교정.

주요 기능:
- 누락된 TAG 자동 생성
- 중복 TAG 자동 제거
- TAG 체인 연결 자동 복구
- 스마트 TAG 제안 시스템
"""

import re
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple

from .policy_validator import PolicyViolation, PolicyViolationType


@dataclass
class AutoCorrection:
    """자동 수정 정보

    Attributes:
        file_path: 수정할 파일 경로
        original_content: 원본 내용
        corrected_content: 수정된 내용
        description: 수정 설명
        confidence: 수정 신뢰도 (0.0-1.0)
        requires_review: 수동 검토 필요 여부
    """
    file_path: str
    original_content: str
    corrected_content: str
    description: str
    confidence: float
    requires_review: bool = False


@dataclass
class AutoCorrectionConfig:
    """자동 수정 설정

    Attributes:
        enable_auto_fix: 자동 수정 활성화
        confidence_threshold: 자동 적용할 최소 신뢰도
        create_missing_specs: 누락된 SPEC 자동 생성
        create_missing_tests: 누락된 TEST 자동 생성
        remove_duplicates: 중복 TAG 자동 제거
        backup_before_fix: 수정 전 백업 생성
    """
    enable_auto_fix: bool = False
    confidence_threshold: float = 0.8
    create_missing_specs: bool = False
    create_missing_tests: bool = False
    remove_duplicates: bool = True
    backup_before_fix: bool = True


class TagAutoCorrector:
    """TAG 오류 자동 수정기

    Post-Tool-Use 훅에서 호출되어 TAG 정책 위반을 자동으로 수정.
    스마트한 알고리즘으로 최적의 TAG를 생성하고 제안.

    Usage:
        config = AutoCorrectionConfig(enable_auto_fix=True)
        corrector = TagAutoCorrector(config=config)

        violations = [...]
        corrections = corrector.generate_corrections(violations)

        if corrections:
            corrector.apply_corrections(corrections)
    """

    # TAG 정규식 패턴
    TAG_PATTERN = re.compile(r"@(SPEC|CODE|TEST|DOC):([A-Z0-9-]+-\d{3})")
    SHEBANG_PATTERN = re.compile(r"^#!.*\n")

    def __init__(self, config: Optional[AutoCorrectionConfig] = None):
        """초기화

        Args:
            config: 자동 수정 설정 (기본: AutoCorrectionConfig())
        """
        self.config = config or AutoCorrectionConfig()

    def generate_corrections(self, violations: List[PolicyViolation]) -> List[AutoCorrection]:
        """정책 위반에 대한 자동 수정 생성

        Args:
            violations: 정책 위반 목록

        Returns:
            AutoCorrection 목록
        """
        corrections = []

        # 파일별로 위반 그룹화
        violations_by_file = self._group_violations_by_file(violations)

        for file_path, file_violations in violations_by_file.items():
            try:
                # 파일 내용 읽기
                path = Path(file_path)
                if not path.exists():
                    continue

                original_content = path.read_text(encoding="utf-8", errors="ignore")
                corrected_content = original_content

                # 각 위반에 대해 수정 적용
                for violation in file_violations:
                    correction = self._generate_single_correction(
                        file_path, corrected_content, violation
                    )
                    if correction:
                        corrected_content = correction.corrected_content
                        corrections.append(correction)

            except Exception:
                # 파일 읽기 실패 시 건너뛰기
                continue

        return corrections

    def apply_corrections(self, corrections: List[AutoCorrection]) -> bool:
        """자동 수정 적용

        Args:
            corrections: 적용할 수정 목록

        Returns:
            성공 여부
        """
        if not self.config.enable_auto_fix:
            return False

        success_count = 0
        for correction in corrections:
            if correction.confidence >= self.config.confidence_threshold:
                try:
                    # 백업 생성
                    if self.config.backup_before_fix:
                        self._create_backup(correction.file_path, correction.original_content)

                    # 수정 적용
                    path = Path(correction.file_path)
                    path.write_text(correction.corrected_content, encoding="utf-8")
                    success_count += 1

                except Exception:
                    # 수정 실패 시 건너뛰기
                    continue

        return success_count == len(corrections)

    def suggest_tag_for_code_file(self, file_path: str) -> Optional[Tuple[str, float]]:
        """코드 파일에 대한 TAG 제안

        Args:
            file_path: 코드 파일 경로

        Returns:
            (TAG, 신뢰도) 튜플 또는 None
        """
        path = Path(file_path)

        # 파일 경로에서 도메인 추출
        domain = self._extract_domain_from_path(path)
        if not domain:
            return None

        # 기존 TAG 확인
        existing_tags = self._find_existing_tags_in_project(domain)

        # 다음 번호 계산
        next_number = self._calculate_next_number(existing_tags, domain)

        tag = f"@CODE:{domain}-{next_number:03d}"
        confidence = self._calculate_tag_confidence(domain, file_path)

        return tag, confidence

    def _group_violations_by_file(self, violations: List[PolicyViolation]) -> Dict[str, List[PolicyViolation]]:
        """파일별로 정책 위반 그룹화

        Args:
            violations: 정책 위반 목록

        Returns:
            {file_path: [violations]} 딕셔너리
        """
        grouped = {}
        for violation in violations:
            if violation.file_path:
                if violation.file_path not in grouped:
                    grouped[violation.file_path] = []
                grouped[violation.file_path].append(violation)
        return grouped

    def _generate_single_correction(self, file_path: str, content: str,
                                  violation: PolicyViolation) -> Optional[AutoCorrection]:
        """단일 정책 위반에 대한 수정 생성

        Args:
            file_path: 파일 경로
            content: 현재 파일 내용
            violation: 정책 위반

        Returns:
            AutoCorrection 또는 None
        """
        if violation.type == PolicyViolationType.MISSING_TAGS:
            return self._fix_missing_tags(file_path, content, violation)
        elif violation.type == PolicyViolationType.DUPLICATE_TAGS:
            return self._fix_duplicate_tags(file_path, content, violation)
        elif violation.type == PolicyViolationType.NO_SPEC_REFERENCE:
            return self._fix_missing_spec_reference(file_path, content, violation)
        elif violation.type == PolicyViolationType.CHAIN_BREAK:
            return self._fix_chain_break(file_path, content, violation)

        return None

    def _fix_missing_tags(self, file_path: str, content: str,
                         violation: PolicyViolation) -> Optional[AutoCorrection]:
        """누락된 TAG 수정

        Args:
            file_path: 파일 경로
            content: 파일 내용
            violation: 정책 위반

        Returns:
            AutoCorrection 또는 None
        """
        # 코드 파일에 대한 TAG 제안
        tag_suggestion = self.suggest_tag_for_code_file(file_path)
        if not tag_suggestion:
            return None

        tag, confidence = tag_suggestion

        # TAG 삽입 위치 찾기
        insert_position = self._find_tag_insert_position(content)

        # TAG 주석 생성
        tag_comment = f"# {tag}\n"

        # 내용에 TAG 삽입
        lines = content.splitlines()
        if insert_position is not None:
            lines.insert(insert_position, tag_comment.strip())
        else:
            # 파일 시작에 삽입 (shebang 다음)
            if self.SHEBANG_PATTERN.match(content):
                shebang_line = lines[0]
                lines = [shebang_line, "", tag_comment.strip()] + lines[1:]
            else:
                lines = [tag_comment.strip()] + lines

        corrected_content = "\n".join(lines) + "\n"

        return AutoCorrection(
            file_path=file_path,
            original_content=content,
            corrected_content=corrected_content,
            description=f"@TAG 자동 추가: {tag}",
            confidence=confidence,
            requires_review=confidence < 0.9
        )

    def _fix_duplicate_tags(self, file_path: str, content: str,
                           violation: PolicyViolation) -> Optional[AutoCorrection]:
        """중복 TAG 수정

        Args:
            file_path: 파일 경로
            content: 파일 내용
            violation: 정책 위반

        Returns:
            AutoCorrection 또는 None
        """
        if not violation.tag:
            return None

        # 중복 TAG 제거 (첫 번째만 유지)
        tag = violation.tag
        corrected_content = content
        found_first = False

        # 정규식으로 모든 TAG 찾기
        pattern = re.compile(re.escape(tag))
        matches = list(pattern.finditer(content))

        if len(matches) <= 1:
            return None

        # 첫 번째를 제외한 모든 TAG 제거
        # 역순으로 처리하여 인덱스 변화 문제 방지
        for match in reversed(matches[1:]):
            start, end = match.span()
            line_start = content.rfind('\n', 0, start) + 1
            line_end = content.find('\n', end)
            if line_end == -1:
                line_end = len(content)

            line = content[line_start:line_end]
            if line.strip() == f"#{tag}":
                # TAG만 있는 라인 제거
                corrected_content = corrected_content[:line_start] + corrected_content[line_end:]
            else:
                # 라인의 일부로 있는 TAG 제거
                corrected_content = (corrected_content[:start] +
                                  corrected_content[end:])

        return AutoCorrection(
            file_path=file_path,
            original_content=content,
            corrected_content=corrected_content,
            description=f"중복 TAG 제거: {tag}",
            confidence=0.95,
            requires_review=False
        )

    def _fix_missing_spec_reference(self, file_path: str, content: str,
                                   violation: PolicyViolation) -> Optional[AutoCorrection]:
        """누락된 SPEC 참조 수정

        Args:
            file_path: 파일 경로
            content: 파일 내용
            violation: 정책 위반

        Returns:
            AutoCorrection 또는 None
        """
        if not violation.tag or not self.config.create_missing_specs:
            return None

        # CODE TAG에서 도메인 추출
        match = self.TAG_PATTERN.search(violation.tag)
        if not match:
            return None

        domain = match.group(2)

        # SPEC 파일 자동 생성
        spec_created = self._create_spec_file(domain)
        if not spec_created:
            return None

        # 기존 내용에 SPEC 참조 추가
        corrected_content = self._add_spec_reference_to_content(content, domain)

        return AutoCorrection(
            file_path=file_path,
            original_content=content,
            corrected_content=corrected_content,
            description=f"SPEC 참조 추가: {domain}",
            confidence=0.8,
            requires_review=True
        )

    def _fix_chain_break(self, file_path: str, content: str,
                        violation: PolicyViolation) -> Optional[AutoCorrection]:
        """체인 끊김 수정

        Args:
            file_path: 파일 경로
            content: 파일 내용
            violation: 정책 위반

        Returns:
            AutoCorrection 또는 None
        """
        if violation.type == PolicyViolationType.CHAIN_BREAK and violation.tag:
            # CODE에 대한 TEST 생성
            if "@CODE:" in violation.tag and self.config.create_missing_tests:
                return self._create_missing_test_file(violation.tag)

        return None

    def _find_tag_insert_position(self, content: str) -> Optional[int]:
        """TAG 삽입 위치 찾기

        Args:
            content: 파일 내용

        Returns:
            라인 번호 (0-based) 또는 None
        """
        lines = content.splitlines()

        # shebang 다음 위치 찾기
        for i, line in enumerate(lines):
            if line.startswith('#!'):
                return i + 1

        # docstring 다음 위치 찾기 (Python)
        for i, line in enumerate(lines):
            if line.strip().startswith('"""') or line.strip().startswith("'''"):
                # docstring 끝 찾기
                quote_char = '"""' if line.strip().startswith('"""') else "'''"
                if quote_char in line[line.find(quote_char)+3:]:
                    return i + 1
                else:
                    # 여러 줄 docstring
                    for j in range(i + 1, len(lines)):
                        if quote_char in lines[j]:
                            return j + 1

        # 첫 번째 빈 라인 또는 주석 다음
        for i, line in enumerate(lines):
            if not line.strip() or line.strip().startswith('#'):
                return i

        return 0

    def _extract_domain_from_path(self, path: Path) -> Optional[str]:
        """파일 경로에서 도메인 추출

        Extract domain from file path:
        - test files (tests/ or test_*.py) → None
        - src/domain/... → domain (uppercase)
        - lib/domain/... → domain (uppercase)
        - filename (no parent dir) → filename stem (uppercase)

        Args:
            path: 파일 경로

        Returns:
            도메인 문자열 또는 None
        """
        parts = path.parts
        filename = path.name

        # Test files should not have domain extraction
        if "tests" in parts or filename.startswith("test_"):
            return None

        # For src/ paths, extract first directory after src
        if "src" in parts:
            src_index = parts.index("src")
            if src_index + 1 < len(parts):
                domain_part = parts[src_index + 1]
                return domain_part.upper().replace("_", "-")

        # For lib/ paths or other paths, extract parent directory
        if len(parts) > 1:
            parent_dir = parts[-2]  # Parent directory
            return parent_dir.upper().replace("_", "-")

        # For single files with no parent directory, use filename stem
        stem = path.stem.upper()
        if stem and re.match(r'^[A-Z-]+$', stem.replace("-", "")):
            return stem.replace("_", "-")

        return None

    def _find_existing_tags_in_project(self, domain: str) -> Set[str]:
        """프로젝트에서 기존 TAG 찾기

        Args:
            domain: 도메인

        Returns:
            기존 TAG 번호 집합
        """
        existing_numbers = set()

        # 프로젝트 루트에서 TAG 검색
        for pattern in ["**/*.py", "**/*.js", "**/*.ts", "**/*.md"]:
            for path in Path(".").glob(pattern):
                if path.is_file():
                    try:
                        content = path.read_text(encoding="utf-8", errors="ignore")
                        matches = self.TAG_PATTERN.findall(content)
                        for tag_type, tag_domain in matches:
                            if tag_domain.startswith(domain):
                                # 도메인-번호 형식에서 번호 추출
                                if f"{domain}-" in tag_domain:
                                    number_part = tag_domain.replace(f"{domain}-", "")
                                    if number_part.isdigit():
                                        existing_numbers.add(int(number_part))
                    except Exception:
                        continue

        return existing_numbers

    def _calculate_next_number(self, existing_numbers: Set[int], domain: str) -> int:
        """다음 TAG 번호 계산

        Args:
            existing_numbers: 기존 TAG 번호 집합
            domain: 도메인

        Returns:
            다음 번호
        """
        if not existing_numbers:
            return 1

        # 1부터 999까지 중 비어있는 가장 작은 번호 찾기
        for i in range(1, 1000):
            if i not in existing_numbers:
                return i

        # 모두 사용된 경우 마지막 번호 + 1
        return max(existing_numbers) + 1

    def _calculate_tag_confidence(self, domain: str, file_path: str) -> float:
        """TAG 신뢰도 계산

        Args:
            domain: 도메인
            file_path: 파일 경로

        Returns:
            신뢰도 (0.0-1.0)
        """
        confidence = 0.5  # 기본 신뢰도

        path = Path(file_path)

        # 경로 기반 신뢰도 증가
        if "src" in str(path):
            confidence += 0.2

        # 도메인 일치 여부
        if domain.lower() in path.stem.lower():
            confidence += 0.2

        # 파일 구조 정확성
        if path.suffix in ['.py', '.js', '.ts']:
            confidence += 0.1

        return min(confidence, 1.0)

    def _create_backup(self, file_path: str, content: str) -> None:
        """백업 파일 생성

        Args:
            file_path: 원본 파일 경로
            content: 원본 내용
        """
        try:
            backup_path = Path(f"{file_path}.backup")
            backup_path.write_text(content, encoding="utf-8")
        except Exception:
            pass

    def _create_spec_file(self, domain: str) -> bool:
        """SPEC 파일 자동 생성

        Args:
            domain: 도메인

        Returns:
            성공 여부
        """
        try:
            spec_dir = Path(f".moai/specs/SPEC-{domain}")
            spec_dir.mkdir(parents=True, exist_ok=True)

            spec_file = spec_dir / "spec.md"
            if not spec_file.exists():
                spec_content = f"""# SPEC: {domain}

## 요구사항

- [요구사항 상세 내용]

## 구현 가이드

### TAG 연결
- @SPEC:{domain} (현재 문서)
- @CODE:{domain} (구현 파일)
- @TEST:{domain} (테스트 파일)

### 인수 조건
- [ ] 기능 구현 완료
- [ ] 테스트 통과
- [ ] 문서화 완료

## 기록

- 생성일: {Path('.').absolute().name}
- 상태: 작성 중
"""
                spec_file.write_text(spec_content, encoding="utf-8")
                return True

        except Exception:
            pass

        return False

    def _add_spec_reference_to_content(self, content: str, domain: str) -> str:
        """내용에 SPEC 참조 추가

        Args:
            content: 원본 내용
            domain: 도메인

        Returns:
            수정된 내용
        """
        lines = content.splitlines()

        # 기존 CODE TAG 찾기
        for i, line in enumerate(lines):
            if f"@CODE:{domain}" in line:
                # SPEC 참조 추가
                spec_ref = f" | SPEC: .moai/specs/SPEC-{domain}/spec.md"
                if spec_ref not in line:
                    lines[i] = line + spec_ref
                break

        return "\n".join(lines) + "\n"

    def _create_missing_test_file(self, code_tag: str) -> Optional[AutoCorrection]:
        """누락된 테스트 파일 생성

        Args:
            code_tag: CODE TAG

        Returns:
            AutoCorrection 또는 None
        """
        match = self.TAG_PATTERN.search(code_tag)
        if not match:
            return None

        domain = match.group(2)

        try:
            test_dir = Path("tests")
            test_dir.mkdir(exist_ok=True)

            test_file = test_dir / f"test_{domain.lower()}.py"

            if not test_file.exists():
                test_content = f'''#!/usr/bin/env python3
# @TEST:{domain} | SPEC: .moai/specs/SPEC-{domain}/spec.md | CODE: @CODE:{domain}
"""Test cases for {domain} functionality.

This module tests the implementation defined in @CODE:{domain}.
"""

import pytest


class Test{domain.replace('-', '')}:
    """Test class for {domain} functionality."""

    def test_basic_functionality(self):
        """Test basic functionality."""
        # TODO: Implement test cases
        assert True

    def test_edge_cases(self):
        """Test edge cases."""
        # TODO: Implement edge case tests
        assert True

    def test_error_conditions(self):
        """Test error conditions."""
        # TODO: Implement error condition tests
        assert True
'''
                test_file.write_text(test_content, encoding="utf-8")

                return AutoCorrection(
                    file_path=str(test_file),
                    original_content="",
                    corrected_content=test_content,
                    description=f"테스트 파일 생성: {domain}",
                    confidence=0.8,
                    requires_review=True
                )

        except Exception:
            pass

        return None
