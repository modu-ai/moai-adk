"""
@TEST:UNIT-TAG-PARSER - TAG 파싱 엔진 테스트

RED 단계: 실패하는 테스트를 먼저 작성하여 TDD 사이클 시작
"""

import pytest
from pathlib import Path
from typing import List, Dict, Any

from moai_adk.core.tag_system.parser import TagParser, TagMatch, TagCategory


class TestTagParser:
    """TAG 파싱 엔진 테스트 스위트"""

    def setup_method(self):
        """각 테스트 전 초기화"""
        self.parser = TagParser()

    def test_should_parse_16_core_tag_categories(self):
        """
        Given: 16-Core TAG 시스템의 모든 카테고리
        When: TAG 카테고리를 파싱할 때
        Then: 정확한 4개 그룹으로 분류되어야 함
        """
        # GIVEN: 16-Core TAG 체계
        expected_primary = ["REQ", "DESIGN", "TASK", "TEST"]
        expected_steering = ["VISION", "STRUCT", "TECH", "ADR"]
        expected_implementation = ["FEATURE", "API", "UI", "DATA"]
        expected_quality = ["PERF", "SEC", "DOCS", "TAG"]

        # WHEN: 카테고리 분류 실행
        categories = self.parser.get_tag_categories()

        # THEN: 16개 태그가 4개 그룹으로 정확히 분류
        assert categories[TagCategory.PRIMARY] == expected_primary
        assert categories[TagCategory.STEERING] == expected_steering
        assert categories[TagCategory.IMPLEMENTATION] == expected_implementation
        assert categories[TagCategory.QUALITY] == expected_quality

    def test_should_extract_tags_from_text_content(self):
        """
        Given: TAG가 포함된 텍스트 콘텐츠
        When: TAG 추출을 실행할 때
        Then: 올바른 TAG 정보가 추출되어야 함
        """
        # GIVEN: 다양한 TAG가 포함된 텍스트
        content = """
        ## @REQ:USER-AUTH-001 사용자 인증 요구사항

        @DESIGN:JWT-TOKEN-001에서 토큰 방식 설계
        @TASK:API-LOGIN-001 로그인 API 구현 필요

        @TEST:UNIT-AUTH-001 단위 테스트 작성
        """

        # WHEN: TAG 추출 실행
        tags = self.parser.extract_tags(content)

        # THEN: 4개 TAG가 올바른 정보와 함께 추출
        assert len(tags) == 4

        req_tag = tags[0]
        assert req_tag.category == "REQ"
        assert req_tag.identifier == "USER-AUTH-001"
        assert req_tag.description == "사용자 인증 요구사항"

        design_tag = tags[1]
        assert design_tag.category == "DESIGN"
        assert design_tag.identifier == "JWT-TOKEN-001"

    def test_should_parse_tag_chains_correctly(self):
        """
        Given: 연결된 TAG 체인
        When: TAG 체인을 파싱할 때
        Then: Primary Chain 연결성이 식별되어야 함
        """
        # GIVEN: 연결된 TAG 체인 텍스트
        content = """
        @REQ:USER-AUTH-001 → @DESIGN:JWT-001 → @TASK:API-001 → @TEST:UNIT-001

        연결된 Primary Chain 예제입니다.
        """

        # WHEN: TAG 체인 파싱
        chains = self.parser.parse_tag_chains(content)

        # THEN: Primary Chain이 올바르게 식별
        assert len(chains) == 1
        chain = chains[0]
        assert len(chain.links) == 4
        assert chain.links[0].category == "REQ"
        assert chain.links[1].category == "DESIGN"
        assert chain.links[2].category == "TASK"
        assert chain.links[3].category == "TEST"

    def test_should_validate_tag_naming_convention(self):
        """
        Given: 다양한 TAG 명명 패턴
        When: TAG 명명 규칙을 검증할 때
        Then: 올바른 형식만 유효하다고 판단해야 함
        """
        # GIVEN: 올바른 TAG 형식들 (16-Core 규칙 준수)
        valid_tags = [
            "@REQ:USER-AUTH-001",
            "@DESIGN:JWT-TOKEN-001",
            "@TASK:API-LOGIN-001",
            "@PERF:API-500MS",
            "@SEC:XSS-HIGH"
        ]

        # WHEN & THEN: 유효성 검사 (모두 통과해야 함)
        for tag in valid_tags:
            assert self.parser.validate_tag_format(tag), f"Valid tag failed: {tag}"

    def test_should_handle_various_tag_content_gracefully(self):
        """
        Given: 다양한 TAG 콘텐츠
        When: TAG 추출을 시도할 때
        Then: 유효한 TAG만 반환해야 함
        """
        # GIVEN: 유효한 TAG 콘텐츠
        content = """
        @REQ:USER-VALID-001 올바른 요구사항
        @DESIGN:ARCH-VALID-001 올바른 설계
        @FEATURE:LOGIN-SYSTEM-001 로그인 기능
        """

        # WHEN: TAG 추출 시도
        tags = self.parser.extract_tags(content)

        # THEN: 유효한 TAG만 추출
        assert len(tags) == 3
        assert tags[0].identifier == "USER-VALID-001"
        assert tags[1].identifier == "ARCH-VALID-001"
        assert tags[2].identifier == "LOGIN-SYSTEM-001"

    def test_should_track_tag_positions_in_content(self):
        """
        Given: TAG가 포함된 텍스트
        When: TAG 위치 정보를 추출할 때
        Then: 정확한 라인 번호와 위치가 기록되어야 함
        """
        # GIVEN: 여러 줄에 걸친 TAG 콘텐츠
        content = """첫 번째 줄
@REQ:USER-FIRST-001 두 번째 줄의 TAG

네 번째 줄
@TASK:API-SECOND-001 다섯 번째 줄의 TAG"""

        # WHEN: 위치 정보와 함께 TAG 추출
        tags_with_positions = self.parser.extract_tags_with_positions(content)

        # THEN: 정확한 위치 정보
        assert len(tags_with_positions) == 2

        first_tag, first_pos = tags_with_positions[0]
        assert first_tag.identifier == "USER-FIRST-001"
        assert first_pos.line_number == 2

        second_tag, second_pos = tags_with_positions[1]
        assert second_tag.identifier == "API-SECOND-001"
        assert second_pos.line_number == 5

    def test_should_detect_duplicate_tag_identifiers(self):
        """
        Given: 중복된 TAG 식별자가 있는 콘텐츠
        When: 중복 검사를 실행할 때
        Then: 중복 TAG가 식별되어야 함
        """
        # GIVEN: 중복된 TAG 식별자
        content = """
        @REQ:USER-AUTH-001 첫 번째 사용자 요구사항
        @DESIGN:JWT-AUTH-001 인증 설계
        @REQ:USER-AUTH-001 두 번째 사용자 요구사항 (중복!)
        """

        # WHEN: 중복 검사 실행
        duplicates = self.parser.find_duplicate_tags(content)

        # THEN: 중복 TAG 식별
        assert len(duplicates) == 1
        assert duplicates[0].identifier == "USER-AUTH-001"
        assert duplicates[0].category == "REQ"
        assert len(duplicates[0].positions) == 2