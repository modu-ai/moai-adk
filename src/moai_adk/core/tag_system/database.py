"""
@FEATURE:SPEC-009-TAG-DATABASE-001 - SQLite 기반 TAG 데이터베이스 관리자

GREEN 단계: 테스트를 통과시키는 최소 구현
"""

import sqlite3
import time
import threading
import logging
from pathlib import Path
from typing import Dict, List, Any, Optional, Union, ContextManager
from dataclasses import dataclass
from datetime import datetime


class DatabaseError(Exception):
    """데이터베이스 관련 오류"""
    pass


class TransactionError(Exception):
    """트랜잭션 관련 오류"""
    pass


@dataclass
class TagSearchResult:
    """TAG 검색 결과"""
    id: int
    category: str
    identifier: str
    description: Optional[str]
    file_path: str
    line_number: Optional[int]
    created_at: str
    updated_at: str


class DatabaseConnection:
    """
    데이터베이스 연결 관리 (Thread-Safe 개선)

    TRUST 원칙 적용:
    - Secured: 스레드별 별도 연결로 안전성 확보
    - Readable: 명확한 연결 관리 로직
    """

    def __init__(self, db_path: Path):
        self.db_path = db_path
        self._local = threading.local()
        self._lock = threading.Lock()

    def get_connection(self) -> sqlite3.Connection:
        """스레드별 독립적인 연결 반환"""
        if not hasattr(self._local, 'connection') or self._local.connection is None:
            self._local.connection = sqlite3.connect(
                str(self.db_path),
                check_same_thread=False,
                timeout=30.0  # 30초 타임아웃
            )
            self._local.connection.row_factory = sqlite3.Row
            # WAL 모드로 동시성 개선
            self._local.connection.execute("PRAGMA journal_mode=WAL")
            self._local.connection.execute("PRAGMA synchronous=NORMAL")
            self._local.connection.commit()

        return self._local.connection

    def close(self):
        """현재 스레드의 연결 종료"""
        if hasattr(self._local, 'connection') and self._local.connection:
            self._local.connection.close()
            self._local.connection = None


class TagDatabase:
    """SQLite 데이터베이스 추상화"""

    def __init__(self, connection: DatabaseConnection):
        self.connection = connection

    def execute(self, query: str, params: tuple = ()) -> sqlite3.Cursor:
        """쿼리 실행"""
        conn = self.connection.get_connection()
        return conn.execute(query, params)

    def executemany(self, query: str, params_list: List[tuple]) -> sqlite3.Cursor:
        """배치 쿼리 실행"""
        conn = self.connection.get_connection()
        return conn.executemany(query, params_list)

    def commit(self):
        """트랜잭션 커밋"""
        conn = self.connection.get_connection()
        conn.commit()


class TagDatabaseManager:
    """
    SQLite 기반 TAG 데이터베이스 관리자

    TRUST 원칙 적용:
    - Test First: 테스트가 요구하는 최소 기능만 구현
    - Readable: 명확한 데이터베이스 작업 로직
    - Unified: 단일 책임 - 데이터베이스 관리만 담당
    """

    # SPEC-009 스키마 정의
    SCHEMA_SQL = {
        'tags': """
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

        'tag_references': """
        CREATE TABLE IF NOT EXISTS tag_references (
            id INTEGER PRIMARY KEY,
            source_tag_id INTEGER NOT NULL,
            target_tag_id INTEGER NOT NULL,
            reference_type TEXT DEFAULT "chain",
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (source_tag_id) REFERENCES tags(id) ON DELETE CASCADE,
            FOREIGN KEY (target_tag_id) REFERENCES tags(id) ON DELETE CASCADE
        )
        """
    }

    # 인덱스 정의
    INDEX_SQL = {
        'idx_tags_category_identifier': """
        CREATE INDEX IF NOT EXISTS idx_tags_category_identifier
        ON tags(category, identifier)
        """,

        'idx_tags_file_path': """
        CREATE INDEX IF NOT EXISTS idx_tags_file_path
        ON tags(file_path)
        """,

        'idx_tag_references_source': """
        CREATE INDEX IF NOT EXISTS idx_tag_references_source
        ON tag_references(source_tag_id)
        """,

        'idx_tag_references_target': """
        CREATE INDEX IF NOT EXISTS idx_tag_references_target
        ON tag_references(target_tag_id)
        """
    }

    def __init__(self, db_path: Path):
        """데이터베이스 관리자 초기화"""
        self.db_path = Path(db_path)
        self._connection = DatabaseConnection(self.db_path)
        self._database = TagDatabase(self._connection)
        self._initialized = False

        # 구조화된 로거 설정 (TRUST 원칙: Secured)
        self._logger = logging.getLogger(f"{__name__}.{self.__class__.__name__}")
        if not self._logger.handlers:
            handler = logging.StreamHandler()
            formatter = logging.Formatter(
                '{"timestamp": "%(asctime)s", "level": "%(levelname)s", '
                '"component": "%(name)s", "message": "%(message)s"}'
            )
            handler.setFormatter(formatter)
            self._logger.addHandler(handler)
            self._logger.setLevel(logging.INFO)

    def initialize(self, create_indexes: bool = True) -> None:
        """데이터베이스 초기화"""
        try:
            # 스키마 생성
            for table_name, sql in self.SCHEMA_SQL.items():
                self._database.execute(sql)

            # 인덱스 생성
            if create_indexes:
                for index_name, sql in self.INDEX_SQL.items():
                    self._database.execute(sql)

            # 업데이트 트리거 생성
            self._create_update_trigger()

            self._database.commit()
            self._initialized = True

        except sqlite3.Error as e:
            raise DatabaseError(f"데이터베이스 초기화 실패: {e}")

    def _create_update_trigger(self):
        """updated_at 자동 업데이트 트리거 생성"""
        trigger_sql = """
        CREATE TRIGGER IF NOT EXISTS update_tags_timestamp
        AFTER UPDATE ON tags
        BEGIN
            UPDATE tags SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
        END
        """
        self._database.execute(trigger_sql)

    def get_schema(self) -> Dict[str, Dict[str, str]]:
        """스키마 정보 반환"""
        schema = {}

        # tags 테이블 스키마
        schema['tags'] = {
            'id': 'INTEGER PRIMARY KEY',
            'category': 'TEXT NOT NULL',
            'identifier': 'TEXT NOT NULL',
            'description': 'TEXT',
            'file_path': 'TEXT NOT NULL',
            'line_number': 'INTEGER',
            'created_at': 'TIMESTAMP DEFAULT CURRENT_TIMESTAMP',
            'updated_at': 'TIMESTAMP DEFAULT CURRENT_TIMESTAMP'
        }

        # tag_references 테이블 스키마
        schema['tag_references'] = {
            'id': 'INTEGER PRIMARY KEY',
            'source_tag_id': 'INTEGER NOT NULL',
            'target_tag_id': 'INTEGER NOT NULL',
            'reference_type': 'TEXT DEFAULT "chain"',
            'created_at': 'TIMESTAMP DEFAULT CURRENT_TIMESTAMP'
        }

        return schema

    def get_indexes(self) -> List[str]:
        """인덱스 목록 반환"""
        return list(self.INDEX_SQL.keys())

    def insert_tag(self, category: str, identifier: str, description: Optional[str] = None,
                   file_path: str = "", line_number: Optional[int] = None) -> int:
        """TAG 삽입 (성능 모니터링 포함)"""
        start_time = time.time()

        # 입력 검증 (TRUST 원칙: Secured)
        if category not in ['REQ', 'DESIGN', 'TASK', 'TEST', 'VISION', 'STRUCT',
                           'TECH', 'ADR', 'FEATURE', 'API', 'UI', 'DATA',
                           'PERF', 'SEC', 'DOCS', 'TAG', 'CUSTOM']:
            raise ValueError(f"Invalid category: {category}")

        if not identifier:
            raise ValueError("Identifier cannot be empty")

        try:
            sql = """
            INSERT INTO tags (category, identifier, description, file_path, line_number)
            VALUES (?, ?, ?, ?, ?)
            """

            cursor = self._database.execute(sql, (
                category, identifier, description or "", file_path, line_number
            ))
            self._database.commit()

            tag_id = cursor.lastrowid

            # 성능 로깅 (TRUST 원칙: Trackable)
            duration = (time.time() - start_time) * 1000
            self._logger.info(
                f'{{"operation": "insert_tag", "category": "{category}", '
                f'"identifier": "{identifier}", "duration_ms": {duration:.2f}, "success": true}}'
            )

            return tag_id

        except Exception as e:
            duration = (time.time() - start_time) * 1000
            self._logger.error(
                f'{{"operation": "insert_tag", "category": "{category}", '
                f'"identifier": "{identifier}", "duration_ms": {duration:.2f}, '
                f'"success": false, "error": "{str(e)}"}}'
            )
            raise

    def get_tag_by_id(self, tag_id: int) -> Optional[Dict[str, Any]]:
        """ID로 TAG 조회"""
        sql = "SELECT * FROM tags WHERE id = ?"
        cursor = self._database.execute(sql, (tag_id,))
        row = cursor.fetchone()

        if row:
            return dict(row)
        return None

    def search_tags_by_category(self, category: str) -> List[Dict[str, Any]]:
        """카테고리별 TAG 검색"""
        sql = "SELECT * FROM tags WHERE category = ? ORDER BY identifier"
        cursor = self._database.execute(sql, (category,))

        return [dict(row) for row in cursor.fetchall()]

    def search_tags_by_identifier(self, identifier: str) -> List[Dict[str, Any]]:
        """식별자로 TAG 검색"""
        sql = "SELECT * FROM tags WHERE identifier = ?"
        cursor = self._database.execute(sql, (identifier,))

        return [dict(row) for row in cursor.fetchall()]

    def search_tags_by_file(self, file_path: str) -> List[Dict[str, Any]]:
        """파일 경로로 TAG 검색"""
        sql = "SELECT * FROM tags WHERE file_path = ? ORDER BY line_number"
        cursor = self._database.execute(sql, (file_path,))

        return [dict(row) for row in cursor.fetchall()]

    def search_tags_by_pattern(self, pattern: str) -> List[Dict[str, Any]]:
        """패턴으로 TAG 검색"""
        sql = "SELECT * FROM tags WHERE identifier LIKE ? ORDER BY identifier"
        cursor = self._database.execute(sql, (f"%{pattern}%",))

        return [dict(row) for row in cursor.fetchall()]

    def search_tags_by_line_range(self, start_line: int, end_line: int) -> List[Dict[str, Any]]:
        """줄 번호 범위로 TAG 검색"""
        sql = """
        SELECT * FROM tags
        WHERE line_number BETWEEN ? AND ?
        ORDER BY line_number, file_path
        """
        cursor = self._database.execute(sql, (start_line, end_line))

        return [dict(row) for row in cursor.fetchall()]

    def create_reference(self, source_tag_id: int, target_tag_id: int,
                        reference_type: str = "chain") -> int:
        """TAG 참조 관계 생성"""
        sql = """
        INSERT INTO tag_references (source_tag_id, target_tag_id, reference_type)
        VALUES (?, ?, ?)
        """

        cursor = self._database.execute(sql, (source_tag_id, target_tag_id, reference_type))
        self._database.commit()

        return cursor.lastrowid

    def get_references_by_source(self, source_tag_id: int) -> List[Dict[str, Any]]:
        """소스 TAG의 참조 관계 조회"""
        sql = "SELECT * FROM tag_references WHERE source_tag_id = ?"
        cursor = self._database.execute(sql, (source_tag_id,))

        return [dict(row) for row in cursor.fetchall()]

    def update_tag(self, tag_id: int, **kwargs) -> int:
        """TAG 업데이트"""
        allowed_fields = ['description', 'file_path', 'line_number']
        updates = []
        values = []

        for field, value in kwargs.items():
            if field in allowed_fields:
                updates.append(f"{field} = ?")
                values.append(value)

        if not updates:
            return 0

        values.append(tag_id)

        sql = f"UPDATE tags SET {', '.join(updates)} WHERE id = ?"
        cursor = self._database.execute(sql, tuple(values))
        self._database.commit()

        return cursor.rowcount

    def delete_tag(self, tag_id: int) -> int:
        """TAG 삭제 (CASCADE로 참조도 삭제됨)"""
        sql = "DELETE FROM tags WHERE id = ?"
        cursor = self._database.execute(sql, (tag_id,))
        self._database.commit()

        return cursor.rowcount

    def get_all_tags(self) -> List[Dict[str, Any]]:
        """모든 TAG 조회"""
        sql = "SELECT * FROM tags ORDER BY category, identifier"
        cursor = self._database.execute(sql)

        return [dict(row) for row in cursor.fetchall()]

    def transaction(self) -> ContextManager:
        """트랜잭션 컨텍스트 매니저"""
        return TransactionManager(self._database)

    def complex_search(self, category: Optional[str] = None,
                      file_pattern: Optional[str] = None,
                      line_range: Optional[tuple] = None) -> List[Dict[str, Any]]:
        """복합 검색"""
        conditions = []
        params = []

        if category:
            conditions.append("category = ?")
            params.append(category)

        if file_pattern:
            conditions.append("file_path LIKE ?")
            params.append(file_pattern.replace('*', '%'))

        if line_range:
            conditions.append("line_number BETWEEN ? AND ?")
            params.extend(line_range)

        where_clause = " AND ".join(conditions) if conditions else "1=1"
        sql = f"SELECT * FROM tags WHERE {where_clause} ORDER BY category, identifier"

        cursor = self._database.execute(sql, tuple(params))
        return [dict(row) for row in cursor.fetchall()]

    def search_tags_by_file_pattern(self, pattern: str) -> List[Dict[str, Any]]:
        """파일 패턴으로 검색"""
        sql = "SELECT * FROM tags WHERE file_path LIKE ? ORDER BY file_path, line_number"
        cursor = self._database.execute(sql, (pattern.replace('*', '%'),))
        return [dict(row) for row in cursor.fetchall()]

    def prepared_insert(self) -> 'PreparedInsertManager':
        """Prepared statement 매니저"""
        return PreparedInsertManager(self._database)

    def bulk_insert_tags(self, tags_data: List[Dict[str, Any]]) -> int:
        """대량 TAG 삽입"""
        sql = """
        INSERT INTO tags (category, identifier, description, file_path, line_number)
        VALUES (?, ?, ?, ?, ?)
        """

        params_list = []
        for tag_data in tags_data:
            params_list.append((
                tag_data['category'],
                tag_data['identifier'],
                tag_data.get('description', ''),
                tag_data.get('file_path', ''),
                tag_data.get('line_number')
            ))

        cursor = self._database.executemany(sql, params_list)
        self._database.commit()

        return cursor.rowcount

    def get_tag_metadata(self, tag_id: int) -> Dict[str, Any]:
        """TAG 메타데이터 조회 (확장용)"""
        # 현재는 기본 구현, 향후 별도 메타데이터 테이블 추가 가능
        tag = self.get_tag_by_id(tag_id)
        if tag:
            return {
                'processed': False,  # 기본값
                'processor': None
            }
        return {}

    def close(self):
        """데이터베이스 연결 종료"""
        self._connection.close()


class TransactionManager:
    """트랜잭션 관리자"""

    def __init__(self, database: TagDatabase):
        self.database = database
        self._in_transaction = False

    def __enter__(self):
        self.database.execute("BEGIN TRANSACTION")
        self._in_transaction = True
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_type is not None:
            self.database.execute("ROLLBACK")
            if exc_type is ValueError and "INVALID" in str(exc_val).upper():
                raise TransactionError(f"Transaction failed: {exc_val}")
        else:
            self.database.commit()
        self._in_transaction = False


class PreparedInsertManager:
    """Prepared statement 삽입 관리자"""

    def __init__(self, database: TagDatabase):
        self.database = database
        self._prepared_data = []

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if self._prepared_data:
            sql = """
            INSERT INTO tags (category, identifier, description, file_path, line_number)
            VALUES (?, ?, ?, ?, ?)
            """
            self.database.executemany(sql, self._prepared_data)
            self.database.commit()

    def execute(self, category: str, identifier: str, description: str = "",
                file_path: str = "", line_number: Optional[int] = None):
        """배치용 데이터 추가"""
        self._prepared_data.append((
            category, identifier, description, file_path, line_number
        ))