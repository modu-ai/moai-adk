"""
Database Transaction Management

데이터베이스 트랜잭션 관리를 담당하는 클래스들입니다.

@DESIGN:TAG-DATABASE-TRANSACTION-001 - 트랜잭션 관리 분리
@TRUST:SECURED - 단일 책임: 안전한 트랜잭션 처리만 담당
"""

from .models import TransactionError


class TransactionManager:
    """트랜잭션 관리자"""

    def __init__(self, database):
        self.database = database
        self._in_transaction = False
        # 트랜잭션 매니저를 데이터베이스에 등록
        self.database._transaction_manager = self

    def __enter__(self):
        self.database.execute("BEGIN TRANSACTION")
        self._in_transaction = True
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self._in_transaction = False

        if exc_type is not None:
            try:
                self.database.execute("ROLLBACK")
            except Exception:
                # 롤백 실패는 무시 (이미 에러 상황)
                pass

            # ValueError를 TransactionError로 변환
            if exc_type is ValueError:
                raise TransactionError(f"Transaction failed: {exc_val}") from exc_val

            # 다른 예외는 그대로 전파
            return False
        else:
            self.database.commit()
            return False


class PreparedInsertManager:
    """Prepared statement 삽입 관리자"""

    def __init__(self, database):
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

    def execute(
        self,
        category: str,
        identifier: str,
        description: str = "",
        file_path: str = "",
        line_number: int | None = None,
    ):
        """배치용 데이터 추가"""
        self._prepared_data.append(
            (category, identifier, description, file_path, line_number)
        )