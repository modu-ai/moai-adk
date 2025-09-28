"""
Database Package for Tag System (Modularized)

TRUST 원칙에 따라 모듈화된 TAG 데이터베이스 시스템입니다.

@DESIGN:TAG-DATABASE-PACKAGE-002 - 모듈화 완료
@TRUST:UNIFIED - 단일 책임 원칙에 따른 모듈 분리
@REFACTOR:MODULARIZATION-001 - 463줄 → 4개 모듈로 분해

Modules:
- connection: 데이터베이스 연결 관리
- models: 데이터 모델 및 예외 클래스
- transaction: 트랜잭션 관리
- schema_manager: 스키마 및 인덱스 관리
- crud_manager: CRUD 작업 처리
- search_manager: 검색 기능 제공
- manager: 통합 관리자 (퍼사드 패턴)
"""

# Core Database Components
from .connection import DatabaseConnection, TagDatabase
from .models import DatabaseError, TransactionError, TagSearchResult
from .transaction import TransactionManager, PreparedInsertManager

# Modularized Managers (TRUST Compliance)
from .schema_manager import SchemaManager
from .crud_manager import TagCrudManager
from .basic_search import BasicTagSearch
from .advanced_search import AdvancedTagSearch
from .search_manager import TagSearchManager
from .manager import TagDatabaseManager

# Backwards compatibility - expose all classes at package level
__all__ = [
    # Connection Management
    "DatabaseConnection",
    "TagDatabase",

    # Models and Exceptions
    "DatabaseError",
    "TransactionError",
    "TagSearchResult",

    # Transaction Management
    "TransactionManager",
    "PreparedInsertManager",

    # Modularized Managers
    "SchemaManager",
    "TagCrudManager",
    "BasicTagSearch",
    "AdvancedTagSearch",
    "TagSearchManager",

    # Main Manager (Facade)
    "TagDatabaseManager",
]