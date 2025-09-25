"""
@FEATURE:TAG-VALIDATOR-001 - Primary Chain 검증 로직 최소 구현

16-Core TAG 체인 유효성 검증 및 무결성 검사
"""

from dataclasses import dataclass

from .parser import TagMatch


@dataclass
class ChainValidationResult:
    """Chain 검증 결과"""

    is_valid: bool
    completeness_score: float
    missing_links: list[str]
    chain_type: str


@dataclass
class ValidationError:
    """검증 오류 정보"""

    message: str
    tag: TagMatch | None = None


@dataclass
class BrokenReference:
    """깨진 참조 정보"""

    source_identifier: str
    broken_reference: str
    reason: str


@dataclass
class ConsistencyViolation:
    """일관성 위반 정보"""

    identifier: str
    category: str
    issue_type: str
    expected: str
    actual: str


class TagValidator:
    """
    Primary Chain 검증 최소 구현

    TRUST 원칙 적용:
    - Test First: 테스트 요구사항만 충족
    - Readable: 명확한 검증 로직
    - Unified: 검증 책임만 담당
    """

    def __init__(self):
        """검증기 초기화"""
        self._primary_chain = ["REQ", "DESIGN", "TASK", "TEST"]
        self._steering_chain = ["VISION", "STRUCT", "TECH", "ADR"]
        self._implementation_chain = ["FEATURE", "API", "UI", "DATA"]
        self._quality_chain = ["PERF", "SEC", "DOCS", "TAG"]

        self._all_categories = {
            "PRIMARY": self._primary_chain,
            "STEERING": self._steering_chain,
            "IMPLEMENTATION": self._implementation_chain,
            "QUALITY": self._quality_chain,
        }

    def validate_primary_chain(self, tags: list[TagMatch]) -> ChainValidationResult:
        """
        Primary Chain 검증

        Args:
            tags: 검증할 TAG 목록

        Returns:
            검증 결과
        """
        primary_tags = [tag for tag in tags if tag.category in self._primary_chain]
        found_categories = set(tag.category for tag in primary_tags)

        missing_links = [
            cat for cat in self._primary_chain if cat not in found_categories
        ]
        completeness_score = len(found_categories) / len(self._primary_chain)

        return ChainValidationResult(
            is_valid=len(missing_links) == 0,
            completeness_score=completeness_score,
            missing_links=missing_links,
            chain_type="PRIMARY",
        )

    def detect_circular_references(self, tags: list[TagMatch]) -> list[list[TagMatch]]:
        """
        순환 참조 검사

        Args:
            tags: 검사할 TAG 목록

        Returns:
            순환 참조가 발견된 TAG 체인 목록
        """
        circular_refs = []

        # 태그별 참조 매핑 구축
        tag_refs = {}
        tag_map = {}

        for tag in tags:
            key = f"{tag.category}:{tag.identifier}"
            tag_map[key] = tag
            tag_refs[key] = tag.references if tag.references else []

        # DFS로 순환 참조 검색
        visited = set()
        recursion_stack = set()

        def dfs(tag_key: str, path: list[TagMatch]) -> None:
            if tag_key in recursion_stack:
                # 순환 참조 발견
                cycle_start = path.index(tag_map[tag_key])
                cycle = path[cycle_start:]
                if cycle:  # 빈 사이클 방지
                    circular_refs.append(cycle)
                return

            if tag_key in visited:
                return

            visited.add(tag_key)
            recursion_stack.add(tag_key)

            if tag_key in tag_map:
                current_path = path + [tag_map[tag_key]]
                for ref in tag_refs.get(tag_key, []):
                    if ref in tag_map:
                        dfs(ref, current_path)

            recursion_stack.discard(tag_key)

        for tag in tags:
            tag_key = f"{tag.category}:{tag.identifier}"
            if tag_key not in visited:
                dfs(tag_key, [])

        return circular_refs

    def find_orphaned_tags(self, tags: list[TagMatch]) -> list[TagMatch]:
        """
        고아 TAG 검색

        Args:
            tags: 검사할 TAG 목록

        Returns:
            고아 TAG 목록
        """
        orphaned_tags = []
        all_references = set()

        # 모든 참조 수집
        for tag in tags:
            if tag.references:
                all_references.update(tag.references)

        # 참조되지 않고 참조도 하지 않는 TAG 찾기
        for tag in tags:
            tag_key = f"{tag.category}:{tag.identifier}"
            has_no_references = not tag.references
            is_not_referenced = tag_key not in all_references

            if has_no_references and is_not_referenced:
                orphaned_tags.append(tag)

        return orphaned_tags

    def check_naming_consistency(
        self, tags: list[TagMatch]
    ) -> list[ConsistencyViolation]:
        """
        명명 일관성 검사

        Args:
            tags: 검사할 TAG 목록

        Returns:
            일관성 위반 목록
        """
        violations = []

        for tag in tags:
            # 대문자-하이픈 패턴 검사
            if not self._is_consistent_naming(tag.identifier):
                violations.append(
                    ConsistencyViolation(
                        identifier=tag.identifier,
                        category=tag.category,
                        issue_type="naming_inconsistency",
                        expected="UPPERCASE-WITH-HYPHENS",
                        actual=tag.identifier,
                    )
                )

        return violations

    def calculate_tag_coverage(self, tags: list[TagMatch]) -> dict[str, float]:
        """
        TAG 커버리지 계산

        Args:
            tags: 커버리지를 계산할 TAG 목록

        Returns:
            카테고리별 커버리지 비율
        """
        coverage = {}

        for group_name, categories in self._all_categories.items():
            found_categories = set()
            for tag in tags:
                if tag.category in categories:
                    found_categories.add(tag.category)

            coverage_ratio = len(found_categories) / len(categories)
            coverage[group_name] = coverage_ratio

        return coverage

    def validate_reference_integrity(
        self, tags: list[TagMatch]
    ) -> list[BrokenReference]:
        """
        참조 무결성 검사

        Args:
            tags: 검사할 TAG 목록

        Returns:
            깨진 참조 목록
        """
        broken_refs = []
        existing_tags = set()

        # 존재하는 TAG 수집
        for tag in tags:
            tag_key = f"{tag.category}:{tag.identifier}"
            existing_tags.add(tag_key)

        # 참조 무결성 검사
        for tag in tags:
            if tag.references:
                for ref in tag.references:
                    if ref not in existing_tags:
                        broken_refs.append(
                            BrokenReference(
                                source_identifier=tag.identifier,
                                broken_reference=ref,
                                reason="Referenced tag does not exist",
                            )
                        )

        return broken_refs

    def _is_consistent_naming(self, identifier: str) -> bool:
        """명명 일관성 검사"""
        import re

        # 대문자-하이픈-숫자 패턴 허용
        pattern = r"^[A-Z][A-Z0-9-]*[A-Z0-9]$"
        return bool(re.match(pattern, identifier))
