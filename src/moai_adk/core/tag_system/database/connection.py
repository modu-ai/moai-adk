"""
Database Connection Management

데이터베이스 연결 관리를 담당하는 클래스들입니다.

@DESIGN:TAG-DATABASE-CONNECTION-001 - 연결 관리 분리
@TRUST:SECURED - 단일 책임: 안전한 데이터베이스 연결만 담당
"""

import sqlite3
import threading
from pathlib import Path


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
        if not hasattr(self._local, "connection") or self._local.connection is None:
            self._local.connection = sqlite3.connect(
                str(self.db_path),
                check_same_thread=False,
                timeout=30.0,  # 30초 타임아웃
            )
            self._local.connection.row_factory = sqlite3.Row
            # WAL 모드로 동시성 개선
            self._local.connection.execute("PRAGMA journal_mode=WAL")
            self._local.connection.execute("PRAGMA synchronous=NORMAL")
            self._local.connection.commit()

        return self._local.connection

    def close(self):
        """현재 스레드의 연결 종료"""
        if hasattr(self._local, "connection") and self._local.connection:
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

    def executemany(self, query: str, params_list: list[tuple]) -> sqlite3.Cursor:
        """배치 쿼리 실행"""
        conn = self.connection.get_connection()
        return conn.executemany(query, params_list)

    def commit(self):
        """트랜잭션 커밋"""
        conn = self.connection.get_connection()
        conn.commit()