"""
@TEST:UNIT-TAG-VALIDATOR - Primary Chain 검증 테스트

RED 단계: Primary Chain 연결성 검증 실패 테스트
"""

import pytest
from typing import List, Dict, Any, Set
from pathlib import Path

from moai_adk.core.tag_system.validator import TagValidator, ChainValidationResult, ValidationError
from moai_adk.core.tag_system.parser import TagMatch


class TestTagValidator:
    """Primary Chain 검증 테스트 스위트"""

    def setup_method(self):
        """각 테스트 전 초기화"""
        self.validator = TagValidator()

    def test_should_validate_complete_primary_chain(self):
        """
        Given: 완전한 Primary Chain (@REQ → @DESIGN → @TASK → @TEST)
        When: Primary Chain 검증을 실행할 때
        Then: 유효한 체인으로 판정되어야 함
        """
        # GIVEN: 완전한 Primary Chain
        tags = [
            TagMatch(category="REQ", identifier="USER-AUTH-001", description="사용자 인증"),
            TagMatch(category="DESIGN", identifier="JWT-DESIGN-001", description="JWT 토큰 설계"),
            TagMatch(category="TASK", identifier="API-LOGIN-001", description="로그인 API"),
            TagMatch(category="TEST", identifier="UNIT-AUTH-001", description="인증 테스트")
        ]

        # WHEN: Primary Chain 검증
        result = self.validator.validate_primary_chain(tags)

        # THEN: 완전한 체인으로 판정
        assert result.is_valid is True
        assert result.completeness_score == 1.0
        assert len(result.missing_links) == 0
        assert result.chain_type == "PRIMARY"

    def test_should_detect_broken_primary_chain(self):
        """
        Given: 끊어진 Primary Chain (DESIGN 누락)
        When: Primary Chain 검증을 실행할 때
        Then: 불완전한 체인으로 판정하고 누락된 링크를 식별해야 함
        """
        # GIVEN: DESIGN이 누락된 Primary Chain
        incomplete_tags = [
            TagMatch(category="REQ", identifier="USER-AUTH-001", description="사용자 인증"),
            TagMatch(category="TASK", identifier="API-LOGIN-001", description="로그인 API"),
            TagMatch(category="TEST", identifier="UNIT-AUTH-001", description="인증 테스트")
        ]

        # WHEN: Primary Chain 검증
        result = self.validator.validate_primary_chain(incomplete_tags)

        # THEN: 불완전한 체인으로 판정, DESIGN 누락 식별
        assert result.is_valid is False
        assert result.completeness_score < 1.0
        assert "DESIGN" in result.missing_links
        assert len(result.missing_links) == 1

    def test_should_detect_circular_tag_references(self):
        """
        Given: 순환 참조가 있는 TAG 체인
        When: 순환 참조 검사를 실행할 때
        Then: 순환 참조를 감지하고 관련 TAG들을 식별해야 함
        """
        # GIVEN: 순환 참조가 있는 TAG 체인
        circular_tags = [
            TagMatch(category="REQ", identifier="USER-A-001", description="A 참조 B", references=["DESIGN:JWT-B-001"]),
            TagMatch(category="DESIGN", identifier="JWT-B-001", description="B 참조 C", references=["TASK:API-C-001"]),
            TagMatch(category="TASK", identifier="API-C-001", description="C 참조 A", references=["REQ:USER-A-001"])
        ]

        # WHEN: 순환 참조 검사
        circular_refs = self.validator.detect_circular_references(circular_tags)

        # THEN: 순환 참조 감지
        assert len(circular_refs) > 0
        cycle = circular_refs[0]
        involved_identifiers = {tag.identifier for tag in cycle}
        assert "USER-A-001" in involved_identifiers
        assert "JWT-B-001" in involved_identifiers
        assert "API-C-001" in involved_identifiers

    def test_should_identify_orphaned_tags(self):
        """
        Given: 고아 TAG들 (연결되지 않은 TAG)
        When: 고아 TAG 검사를 실행할 때
        Then: 연결되지 않은 TAG들을 식별해야 함
        """
        # GIVEN: 연결된 TAG와 고아 TAG가 섞인 상황
        mixed_tags = [
            # 연결된 Primary Chain
            TagMatch(category="REQ", identifier="USER-AUTH-001", references=["DESIGN:JWT-DESIGN-001"]),
            TagMatch(category="DESIGN", identifier="JWT-DESIGN-001", references=["TASK:API-LOGIN-001"]),
            TagMatch(category="TASK", identifier="API-LOGIN-001", references=["TEST:UNIT-AUTH-001"]),
            TagMatch(category="TEST", identifier="UNIT-AUTH-001"),

            # 고아 TAG들
            TagMatch(category="REQ", identifier="ORPHAN-REQ-001"),  # 참조 없음
            TagMatch(category="FEATURE", identifier="ORPHAN-FEATURE-001")  # 참조 없음
        ]

        # WHEN: 고아 TAG 검사
        orphaned_tags = self.validator.find_orphaned_tags(mixed_tags)

        # THEN: 고아 TAG들 식별
        assert len(orphaned_tags) == 2
        orphaned_identifiers = {tag.identifier for tag in orphaned_tags}
        assert "ORPHAN-REQ-001" in orphaned_identifiers
        assert "ORPHAN-FEATURE-001" in orphaned_identifiers

    def test_should_validate_tag_naming_consistency(self):
        """
        Given: 다양한 명명 규칙의 TAG들
        When: 명명 일관성을 검사할 때
        Then: 일관성 위반을 감지해야 함
        """
        # GIVEN: 일관성 있는 TAG와 위반 TAG가 섞인 상황
        inconsistent_tags = [
            TagMatch(category="REQ", identifier="USER-AUTH-001"),  # 올바른 형식
            TagMatch(category="REQ", identifier="user-auth-002"),  # 소문자 (위반)
            TagMatch(category="DESIGN", identifier="JWT-DESIGN-001"),  # 올바른 형식
            TagMatch(category="DESIGN", identifier="jwt_design_002")  # 언더스코어 (위반)
        ]

        # WHEN: 명명 일관성 검사
        consistency_violations = self.validator.check_naming_consistency(inconsistent_tags)

        # THEN: 일관성 위반 감지
        assert len(consistency_violations) == 2
        violation_identifiers = {v.identifier for v in consistency_violations}
        assert "user-auth-002" in violation_identifiers
        assert "jwt_design_002" in violation_identifiers

    def test_should_measure_tag_coverage_across_categories(self):
        """
        Given: 다양한 카테고리의 TAG들
        When: TAG 커버리지를 측정할 때
        Then: 각 카테고리별 커버리지 비율을 계산해야 함
        """
        # GIVEN: 불균등한 카테고리 분포
        uneven_tags = [
            # Primary: 4개 모두 존재 (100%)
            TagMatch(category="REQ", identifier="USER-001"),
            TagMatch(category="DESIGN", identifier="ARCH-001"),
            TagMatch(category="TASK", identifier="IMPL-001"),
            TagMatch(category="TEST", identifier="UNIT-001"),

            # Implementation: 2개만 존재 (50%)
            TagMatch(category="FEATURE", identifier="LOGIN-001"),
            TagMatch(category="API", identifier="AUTH-001")
            # UI, DATA 누락
        ]

        # WHEN: 커버리지 측정
        coverage_report = self.validator.calculate_tag_coverage(uneven_tags)

        # THEN: 정확한 커버리지 계산
        assert coverage_report["PRIMARY"] == 1.0  # 100% (4/4)
        assert coverage_report["IMPLEMENTATION"] == 0.5  # 50% (2/4)
        assert coverage_report["STEERING"] == 0.0  # 0% (0/4)
        assert coverage_report["QUALITY"] == 0.0  # 0% (0/4)

    def test_should_validate_tag_reference_integrity(self):
        """
        Given: 상호 참조하는 TAG들
        When: 참조 무결성을 검사할 때
        Then: 깨진 참조와 유효한 참조를 구분해야 함
        """
        # GIVEN: 유효한 참조와 깨진 참조가 섞인 상황
        tags_with_refs = [
            TagMatch(category="REQ", identifier="USER-001", references=["DESIGN:ARCH-001"]),  # 유효
            TagMatch(category="DESIGN", identifier="ARCH-001", references=["TASK:IMPL-001"]),  # 유효
            TagMatch(category="TASK", identifier="IMPL-001", references=["TEST:UNIT-999"]),  # 깨진 참조
            TagMatch(category="TEST", identifier="UNIT-001")  # 실제 존재하는 테스트
        ]

        # WHEN: 참조 무결성 검사
        broken_refs = self.validator.validate_reference_integrity(tags_with_refs)

        # THEN: 깨진 참조만 식별
        assert len(broken_refs) == 1
        broken_ref = broken_refs[0]
        assert broken_ref.source_identifier == "IMPL-001"
        assert broken_ref.broken_reference == "TEST:UNIT-999"
        assert broken_ref.reason == "Referenced tag does not exist"