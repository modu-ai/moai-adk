"""
Database Models and Exceptions

데이터베이스 관련 데이터 모델과 예외 클래스들입니다.

@DESIGN:TAG-DATABASE-MODELS-001 - 데이터 모델 분리
@TRUST:SIMPLE - 단일 책임: 데이터 구조 정의만 담당
"""

from dataclasses import dataclass


class DatabaseError(Exception):
    """데이터베이스 관련 오류"""


class TransactionError(Exception):
    """트랜잭션 관련 오류"""


@dataclass
class TagSearchResult:
    """TAG 검색 결과"""

    id: int
    category: str
    identifier: str
    description: str | None
    file_path: str
    line_number: int | None
    created_at: str
    updated_at: str