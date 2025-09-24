"""
@FEATURE:TAG-PARSER-001 - TAG 파싱 엔진 최소 구현

16-Core TAG 시스템의 TAG 추출 및 분류 엔진
"""

import re
from enum import Enum
from dataclasses import dataclass
from typing import List, Dict, Any, Optional, Tuple


class TagCategory(Enum):
    """16-Core TAG 카테고리 분류"""
    PRIMARY = "PRIMARY"
    STEERING = "STEERING"
    IMPLEMENTATION = "IMPLEMENTATION"
    QUALITY = "QUALITY"


@dataclass
class TagMatch:
    """TAG 매칭 결과"""
    category: str
    identifier: str
    description: Optional[str] = None
    references: Optional[List[str]] = None

    def __post_init__(self):
        if self.references is None:
            self.references = []


@dataclass
class TagPosition:
    """TAG 위치 정보"""
    line_number: int
    column: int = 0


@dataclass
class TagChain:
    """TAG 체인"""
    links: List[TagMatch]


@dataclass
class DuplicateTagInfo:
    """중복 TAG 정보"""
    category: str
    identifier: str
    positions: List[TagPosition]


class TagParser:
    """
    16-Core TAG 파싱 엔진 최소 구현

    TRUST 원칙 적용:
    - Test First: 테스트가 요구하는 최소 기능만 구현
    - Readable: 명시적이고 이해하기 쉬운 코드
    - Unified: 단일 책임 - TAG 파싱만 담당
    """

    # 상수: 16-Core TAG 체계 정의
    PRIMARY_CATEGORIES = ["REQ", "DESIGN", "TASK", "TEST"]
    STEERING_CATEGORIES = ["VISION", "STRUCT", "TECH", "ADR"]
    IMPLEMENTATION_CATEGORIES = ["FEATURE", "API", "UI", "DATA"]
    QUALITY_CATEGORIES = ["PERF", "SEC", "DOCS", "TAG"]

    # 정규표현식 패턴 상수
    TAG_PATTERN = r'@([A-Z]+):([A-Z0-9-]+)(?:\s+(.+))?'
    CHAIN_PATTERN = r'@[A-Z]+:[A-Z0-9-]+(?:\s*→\s*@[A-Z]+:[A-Z0-9-]+)+'
    INDIVIDUAL_TAG_PATTERN = r'@([A-Z]+):([A-Z0-9-]+)'

    def __init__(self):
        """TAG 파서 초기화"""
        self._tag_categories = {
            TagCategory.PRIMARY: self.PRIMARY_CATEGORIES,
            TagCategory.STEERING: self.STEERING_CATEGORIES,
            TagCategory.IMPLEMENTATION: self.IMPLEMENTATION_CATEGORIES,
            TagCategory.QUALITY: self.QUALITY_CATEGORIES
        }

        # TAG 매칭 정규표현식 컴파일
        self._tag_pattern = re.compile(self.TAG_PATTERN)
        self._chain_pattern = re.compile(self.CHAIN_PATTERN)
        self._individual_tag_pattern = re.compile(self.INDIVIDUAL_TAG_PATTERN)

    def get_tag_categories(self) -> Dict[TagCategory, List[str]]:
        """16-Core TAG 카테고리 반환"""
        return self._tag_categories.copy()

    def extract_tags(self, content: str) -> List[TagMatch]:
        """
        텍스트에서 TAG 추출

        Args:
            content: 분석할 텍스트 콘텐츠

        Returns:
            추출된 TAG 목록
        """
        tags = []

        for match in self._tag_pattern.finditer(content):
            category = match.group(1)
            identifier = match.group(2)
            description = match.group(3) if match.group(3) else None

            # 16-Core TAG 검증
            if self._is_valid_tag_category(category):
                tags.append(TagMatch(
                    category=category,
                    identifier=identifier,
                    description=description
                ))

        return tags

    def parse_tag_chains(self, content: str) -> List[TagChain]:
        """
        TAG 체인 파싱

        Args:
            content: 체인이 포함된 텍스트

        Returns:
            파싱된 TAG 체인 목록
        """
        chains = []

        # → 기호로 연결된 TAG 체인 검색
        for match in self._chain_pattern.finditer(content):
            full_match = match.group(0)
            # 개별 TAG들을 추출
            individual_tags = re.findall(self.INDIVIDUAL_TAG_PATTERN, full_match)

            links = []
            for category, identifier in individual_tags:
                if self._is_valid_tag_category(category):
                    links.append(TagMatch(
                        category=category,
                        identifier=identifier,
                        description=None
                    ))

            if links:
                chains.append(TagChain(links=links))

        return chains

    def validate_tag_format(self, tag_string: str) -> bool:
        """
        TAG 형식 검증

        Args:
            tag_string: 검증할 TAG 문자열

        Returns:
            유효한 형식인지 여부
        """
        match = self._tag_pattern.match(tag_string.strip())
        if not match:
            return False

        category = match.group(1)
        return self._is_valid_tag_category(category)

    def extract_tags_with_positions(self, content: str) -> List[Tuple[TagMatch, TagPosition]]:
        """
        위치 정보와 함께 TAG 추출

        Args:
            content: 분석할 텍스트

        Returns:
            (TAG, 위치) 튜플 목록
        """
        results = []
        lines = content.split('\n')

        for line_num, line in enumerate(lines, 1):
            for match in self._tag_pattern.finditer(line):
                category = match.group(1)
                identifier = match.group(2)
                description = match.group(3) if match.group(3) else None

                if self._is_valid_tag_category(category):
                    tag = TagMatch(
                        category=category,
                        identifier=identifier,
                        description=description
                    )
                    position = TagPosition(line_number=line_num, column=match.start())
                    results.append((tag, position))

        return results

    def find_duplicate_tags(self, content: str) -> List[DuplicateTagInfo]:
        """
        중복 TAG 검색

        Args:
            content: 검색할 텍스트

        Returns:
            중복 TAG 정보 목록
        """
        tag_positions = {}
        tags_with_pos = self.extract_tags_with_positions(content)

        # TAG별 위치 수집
        for tag, position in tags_with_pos:
            key = f"{tag.category}:{tag.identifier}"
            if key not in tag_positions:
                tag_positions[key] = []
            tag_positions[key].append(position)

        # 중복 TAG 식별
        duplicates = []
        for key, positions in tag_positions.items():
            if len(positions) > 1:
                category, identifier = key.split(':', 1)
                duplicates.append(DuplicateTagInfo(
                    category=category,
                    identifier=identifier,
                    positions=positions
                ))

        return duplicates

    def _is_valid_tag_category(self, category: str) -> bool:
        """TAG 카테고리 유효성 검사"""
        all_categories = []
        for category_list in self._tag_categories.values():
            all_categories.extend(category_list)
        return category in all_categories