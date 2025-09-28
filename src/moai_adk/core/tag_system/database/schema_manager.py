"""
Database Schema Manager

TAG 데이터베이스의 스키마와 인덱스 관리를 담당합니다.

@DESIGN:TAG-SCHEMA-MANAGER-001 - 스키마 관리 분리
@TRUST:UNIFIED - 단일 책임: 스키마 및 인덱스 관리만 담당
"""

import logging
from pathlib import Path

from .models import DatabaseError

# 로깅 설정
logger = logging.getLogger(__name__)


class SchemaManager:
    """
    TAG 데이터베이스 스키마 관리자

    TRUST 원칙 적용:
    - Test First: 테스트가 요구하는 최소 기능만 구현
    - Readable: 명확한 스키마 관리 로직
    - Unified: 단일 책임 - 스키마 관리만 담당
    - Secured: 스키마 검증 및 오류 처리
    - Trackable: 구조화 로깅
    """

    # SPEC-009 스키마 정의
    SCHEMA_SQL = {
        "tags": """
        CREATE TABLE IF NOT EXISTS tags (
            id INTEGER PRIMARY KEY,
            category TEXT NOT NULL,
            identifier TEXT NOT NULL,
            description TEXT,
            file_path TEXT NOT NULL,
            line_number INTEGER,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
        """,
        "tag_references": """
        CREATE TABLE IF NOT EXISTS tag_references (
            id INTEGER PRIMARY KEY,
            source_tag_id INTEGER NOT NULL,
            target_tag_id INTEGER NOT NULL,
            reference_type TEXT DEFAULT "chain",
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (source_tag_id) REFERENCES tags(id) ON DELETE CASCADE,
            FOREIGN KEY (target_tag_id) REFERENCES tags(id) ON DELETE CASCADE
        )
        """,
    }

    # 인덱스 정의
    INDEX_SQL = {
        "idx_tags_category_identifier": """
        CREATE INDEX IF NOT EXISTS idx_tags_category_identifier
        ON tags(category, identifier)
        """,
        "idx_tags_file_path": """
        CREATE INDEX IF NOT EXISTS idx_tags_file_path
        ON tags(file_path)
        """,
        "idx_tag_references_source": """
        CREATE INDEX IF NOT EXISTS idx_tag_references_source
        ON tag_references(source_tag_id)
        """,
        "idx_tag_references_target": """
        CREATE INDEX IF NOT EXISTS idx_tag_references_target
        ON tag_references(target_tag_id)
        """,
    }

    def __init__(self, database):
        """스키마 관리자 초기화"""
        self._database = database
        logger.debug("SchemaManager 초기화")

    def initialize_schema(self, db_path: Path) -> None:
        """데이터베이스 스키마 초기화"""
        try:
            # 디렉토리 생성
            self._create_directory(db_path)

            # Foreign Key 제약 조건 활성화
            self._database.execute("PRAGMA foreign_keys = ON")

            # 스키마 생성
            for table_name, schema in self.SCHEMA_SQL.items():
                self._database.execute(schema)
                logger.debug(f"테이블 생성: {table_name}")

            # 업데이트 트리거 생성
            self._create_update_trigger()

            logger.info(f"데이터베이스 스키마 초기화 완료: {db_path}")

        except Exception as e:
            logger.error(f"스키마 초기화 실패: {e}")
            self._handle_schema_error(e)

    def create_indexes(self) -> None:
        """인덱스 생성"""
        try:
            for index_name, index_sql in self.INDEX_SQL.items():
                self._database.execute(index_sql)
                logger.debug(f"인덱스 생성: {index_name}")

            logger.info("모든 인덱스 생성 완료")

        except Exception as e:
            logger.error(f"인덱스 생성 실패: {e}")
            raise DatabaseError(f"인덱스 생성 실패: {e}")

    def get_schema(self) -> dict[str, dict[str, str]]:
        """테이블 스키마 정보 반환"""
        schema = {}

        # tags 테이블 스키마
        schema["tags"] = {
            "id": "INTEGER PRIMARY KEY",
            "category": "TEXT NOT NULL",
            "identifier": "TEXT NOT NULL",
            "description": "TEXT",
            "file_path": "TEXT NOT NULL",
            "line_number": "INTEGER",
            "created_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
            "updated_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
        }

        # tag_references 테이블 스키마
        schema["tag_references"] = {
            "id": "INTEGER PRIMARY KEY",
            "source_tag_id": "INTEGER NOT NULL",
            "target_tag_id": "INTEGER NOT NULL",
            "reference_type": 'TEXT DEFAULT "chain"',
            "created_at": "TIMESTAMP DEFAULT CURRENT_TIMESTAMP",
        }

        return schema

    def get_indexes(self) -> list[str]:
        """생성된 인덱스 목록 반환"""
        try:
            cursor = self._database.execute(
                "SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%'"
            )
            return [row[0] for row in cursor.fetchall()]
        except Exception as e:
            logger.error(f"인덱스 조회 실패: {e}")
            return []

    def _create_directory(self, db_path: Path) -> None:
        """데이터베이스 디렉토리 생성"""
        try:
            db_path.parent.mkdir(parents=True, exist_ok=True)
        except (OSError, PermissionError) as e:
            if "invalid" in str(db_path).lower():
                raise DatabaseError(f"데이터베이스 생성 실패: 잘못된 경로 {db_path}")
            raise DatabaseError(f"디렉토리 생성 실패: {e}")

    def _create_update_trigger(self) -> None:
        """업데이트 시 updated_at 필드 자동 갱신 트리거 생성"""
        trigger_sql = """
        CREATE TRIGGER IF NOT EXISTS update_tags_timestamp
        AFTER UPDATE ON tags
        FOR EACH ROW
        BEGIN
            UPDATE tags SET updated_at = datetime('now', 'localtime') WHERE id = NEW.id;
        END
        """
        self._database.execute(trigger_sql)
        logger.debug("updated_at 트리거 생성")

    def _handle_schema_error(self, error: Exception) -> None:
        """스키마 관련 오류 처리"""
        error_msg = str(error).lower()
        if "not a database" in error_msg or "malformed" in error_msg:
            raise DatabaseError(f"데이터베이스 손상: {error}")
        else:
            raise DatabaseError(f"스키마 초기화 실패: {error}")