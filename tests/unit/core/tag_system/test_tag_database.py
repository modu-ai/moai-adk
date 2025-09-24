"""
@TEST:SPEC-009-TAG-DATABASE-001 - TAG SQLite 데이터베이스 관리자 실패 테스트

RED 단계: SQLite 기반 TAG 저장소 실패 테스트
"""

import pytest
import sqlite3
import tempfile
import time
from pathlib import Path
from typing import List, Dict, Any
from unittest.mock import Mock, patch

# 아직 구현되지 않은 모듈들 - 실패할 예정
from moai_adk.core.tag_system.database import (
    TagDatabaseManager,
    TagDatabase,
    DatabaseConnection,
    TagSearchResult,
    DatabaseError,
    TransactionError
)


class TestTagDatabaseManager:
    """TAG SQLite 데이터베이스 관리자 테스트 스위트"""

    def setup_method(self):
        """각 테스트 전 초기화"""
        self.temp_db = Path(tempfile.mktemp(suffix='.db'))
        self.db_manager = TagDatabaseManager(self.temp_db)

    def teardown_method(self):
        """각 테스트 후 정리"""
        self.db_manager.close()
        if self.temp_db.exists():
            self.temp_db.unlink()

    def test_should_create_database_schema_on_initialization(self):
        """
        Given: 새로운 SQLite 데이터베이스 파일
        When: TagDatabaseManager를 초기화할 때
        Then: SPEC-009에 정의된 스키마가 생성되어야 함
        """
        # WHEN: 데이터베이스 초기화
        self.db_manager.initialize()

        # THEN: 스키마 구조 검증
        schema = self.db_manager.get_schema()

        # tags 테이블 검증
        assert 'tags' in schema
        tags_columns = schema['tags']
        expected_tags_columns = [
            ('id', 'INTEGER PRIMARY KEY'),
            ('category', 'TEXT NOT NULL'),
            ('identifier', 'TEXT NOT NULL'),
            ('description', 'TEXT'),
            ('file_path', 'TEXT NOT NULL'),
            ('line_number', 'INTEGER'),
            ('created_at', 'TIMESTAMP DEFAULT CURRENT_TIMESTAMP'),
            ('updated_at', 'TIMESTAMP DEFAULT CURRENT_TIMESTAMP')
        ]
        for col_name, col_type in expected_tags_columns:
            assert col_name in tags_columns
            assert tags_columns[col_name] == col_type

        # tag_references 테이블 검증
        assert 'tag_references' in schema
        references_columns = schema['tag_references']
        expected_ref_columns = [
            ('id', 'INTEGER PRIMARY KEY'),
            ('source_tag_id', 'INTEGER NOT NULL'),
            ('target_tag_id', 'INTEGER NOT NULL'),
            ('reference_type', 'TEXT DEFAULT "chain"'),
            ('created_at', 'TIMESTAMP DEFAULT CURRENT_TIMESTAMP')
        ]
        for col_name, col_type in expected_ref_columns:
            assert col_name in references_columns
            assert references_columns[col_name] == col_type

        # 인덱스 검증
        indexes = self.db_manager.get_indexes()
        expected_indexes = [
            'idx_tags_category_identifier',
            'idx_tags_file_path',
            'idx_tag_references_source',
            'idx_tag_references_target'
        ]
        for index_name in expected_indexes:
            assert index_name in indexes

    def test_should_insert_tag_with_all_fields(self):
        """
        Given: 초기화된 SQLite 데이터베이스
        When: 모든 필드가 포함된 TAG를 삽입할 때
        Then: 올바르게 저장되고 ID가 반환되어야 함
        """
        # GIVEN: 데이터베이스 초기화
        self.db_manager.initialize()

        # WHEN: TAG 삽입
        tag_data = {
            'category': 'REQ',
            'identifier': 'USER-LOGIN-001',
            'description': '사용자 로그인 기능 요구사항',
            'file_path': '/path/to/requirements.md',
            'line_number': 42
        }
        tag_id = self.db_manager.insert_tag(**tag_data)

        # THEN: 올바른 ID 반환 및 데이터 저장 확인
        assert isinstance(tag_id, int)
        assert tag_id > 0

        # 저장된 데이터 조회
        retrieved_tag = self.db_manager.get_tag_by_id(tag_id)
        assert retrieved_tag is not None
        assert retrieved_tag['category'] == 'REQ'
        assert retrieved_tag['identifier'] == 'USER-LOGIN-001'
        assert retrieved_tag['description'] == '사용자 로그인 기능 요구사항'
        assert retrieved_tag['file_path'] == '/path/to/requirements.md'
        assert retrieved_tag['line_number'] == 42
        assert 'created_at' in retrieved_tag
        assert 'updated_at' in retrieved_tag

    def test_should_search_tags_by_category_with_performance(self):
        """
        Given: 다양한 카테고리의 TAG들이 저장된 데이터베이스
        When: 특정 카테고리로 검색을 수행할 때
        Then: 10x 성능 개선 목표를 만족하며 정확한 결과를 반환해야 함
        """
        # GIVEN: 대량 TAG 데이터 삽입
        self.db_manager.initialize()

        # 1000개 TAG 생성 (SPEC-009 성능 테스트 기준)
        categories = ['REQ', 'DESIGN', 'TASK', 'TEST']
        test_tags = []
        for i in range(1000):
            category = categories[i % len(categories)]
            tag_data = {
                'category': category,
                'identifier': f'PERF-TEST-{i:04d}',
                'description': f'성능 테스트용 TAG {i}',
                'file_path': f'/test/file_{i // 100}.md',
                'line_number': i % 100 + 1
            }
            test_tags.append(tag_data)
            self.db_manager.insert_tag(**tag_data)

        # WHEN: 카테고리별 검색 성능 측정
        start_time = time.time()
        req_tags = self.db_manager.search_tags_by_category('REQ')
        search_time = time.time() - start_time

        # THEN: 성능 및 정확성 검증
        # 10x 성능 개선: JSON에서 0.1초 → SQLite에서 0.01초 목표
        assert search_time < 0.05, f"검색 시간 {search_time:.3f}초는 목표 0.05초를 초과"

        # 정확한 결과 개수 확인 (1000개 중 250개가 REQ)
        assert len(req_tags) == 250

        # 결과 구조 검증
        for tag in req_tags:
            assert tag['category'] == 'REQ'
            assert 'identifier' in tag
            assert 'description' in tag
            assert 'file_path' in tag

    def test_should_create_tag_reference_chain(self):
        """
        Given: 두 개의 TAG가 데이터베이스에 저장됨
        When: TAG 간 참조 관계를 생성할 때
        Then: tag_references 테이블에 올바르게 저장되어야 함
        """
        # GIVEN: 두 TAG 생성
        self.db_manager.initialize()

        source_tag = self.db_manager.insert_tag(
            category='REQ',
            identifier='USER-AUTH-001',
            description='사용자 인증 요구사항',
            file_path='/spec/requirements.md',
            line_number=10
        )

        target_tag = self.db_manager.insert_tag(
            category='DESIGN',
            identifier='AUTH-SYSTEM-001',
            description='인증 시스템 설계',
            file_path='/spec/design.md',
            line_number=25
        )

        # WHEN: 참조 관계 생성
        reference_id = self.db_manager.create_reference(
            source_tag_id=source_tag,
            target_tag_id=target_tag,
            reference_type='chain'
        )

        # THEN: 참조 관계 확인
        assert isinstance(reference_id, int)
        assert reference_id > 0

        # 참조 관계 조회
        references = self.db_manager.get_references_by_source(source_tag)
        assert len(references) == 1
        assert references[0]['target_tag_id'] == target_tag
        assert references[0]['reference_type'] == 'chain'

    def test_should_update_existing_tag(self):
        """
        Given: 기존에 저장된 TAG
        When: TAG 정보를 업데이트할 때
        Then: 변경사항이 올바르게 반영되고 updated_at이 갱신되어야 함
        """
        # GIVEN: 기존 TAG 생성
        self.db_manager.initialize()
        original_tag = self.db_manager.insert_tag(
            category='REQ',
            identifier='USER-PROFILE-001',
            description='기존 설명',
            file_path='/old/path.md',
            line_number=10
        )

        # 원본 updated_at 기록
        original_data = self.db_manager.get_tag_by_id(original_tag)
        original_updated_at = original_data['updated_at']

        # 약간의 시간 지연
        time.sleep(0.001)

        # WHEN: TAG 업데이트
        updated_rows = self.db_manager.update_tag(
            tag_id=original_tag,
            description='업데이트된 설명',
            file_path='/new/path.md',
            line_number=20
        )

        # THEN: 업데이트 확인
        assert updated_rows == 1

        updated_data = self.db_manager.get_tag_by_id(original_tag)
        assert updated_data['description'] == '업데이트된 설명'
        assert updated_data['file_path'] == '/new/path.md'
        assert updated_data['line_number'] == 20
        assert updated_data['updated_at'] > original_updated_at

        # 변경되지 않은 필드 확인
        assert updated_data['category'] == 'REQ'
        assert updated_data['identifier'] == 'USER-PROFILE-001'

    def test_should_delete_tag_and_cascade_references(self):
        """
        Given: 참조 관계가 있는 TAG들
        When: 소스 TAG를 삭제할 때
        Then: TAG와 관련된 모든 참조 관계도 함께 삭제되어야 함
        """
        # GIVEN: 참조 관계가 있는 TAG들 생성
        self.db_manager.initialize()

        source_tag = self.db_manager.insert_tag(
            category='REQ',
            identifier='TO-DELETE-001',
            description='삭제될 TAG',
            file_path='/temp.md',
            line_number=1
        )

        target_tag = self.db_manager.insert_tag(
            category='DESIGN',
            identifier='DESIGN-001',
            description='유지될 TAG',
            file_path='/design.md',
            line_number=1
        )

        # 참조 관계 생성
        self.db_manager.create_reference(source_tag, target_tag, 'chain')

        # WHEN: 소스 TAG 삭제
        deleted_rows = self.db_manager.delete_tag(source_tag)

        # THEN: TAG와 참조 관계 모두 삭제 확인
        assert deleted_rows == 1
        assert self.db_manager.get_tag_by_id(source_tag) is None
        assert len(self.db_manager.get_references_by_source(source_tag)) == 0

        # 타겟 TAG는 유지되어야 함
        assert self.db_manager.get_tag_by_id(target_tag) is not None

    def test_should_handle_database_transaction_rollback(self):
        """
        Given: 트랜잭션 내에서 여러 작업 수행
        When: 중간에 오류가 발생할 때
        Then: 모든 변경사항이 롤백되어야 함
        """
        # GIVEN: 데이터베이스 초기화
        self.db_manager.initialize()

        # WHEN: 트랜잭션 내에서 작업 수행 중 오류 발생
        with pytest.raises(TransactionError):
            with self.db_manager.transaction():
                # 첫 번째 TAG 삽입 (성공)
                tag1 = self.db_manager.insert_tag(
                    category='REQ',
                    identifier='TRANSACTION-TEST-001',
                    description='트랜잭션 테스트',
                    file_path='/test.md',
                    line_number=1
                )

                # 잘못된 데이터로 두 번째 TAG 삽입 시도 (실패 예정)
                self.db_manager.insert_tag(
                    category='INVALID',  # 유효하지 않은 카테고리
                    identifier='',  # 빈 식별자
                    description=None,
                    file_path=None,  # 필수 필드 누락
                    line_number=-1  # 잘못된 값
                )

        # THEN: 모든 변경사항이 롤백됨
        # 첫 번째 TAG도 삽입되지 않아야 함
        all_tags = self.db_manager.get_all_tags()
        assert len(all_tags) == 0

    def test_should_perform_complex_search_queries(self):
        """
        Given: 다양한 TAG 데이터가 있는 데이터베이스
        When: 복합 검색 쿼리를 수행할 때
        Then: 정확한 결과를 빠르게 반환해야 함
        """
        # GIVEN: 다양한 TAG 데이터 준비
        self.db_manager.initialize()

        test_data = [
            ('REQ', 'USER-AUTH-001', 'login.md', 10),
            ('REQ', 'USER-AUTH-002', 'login.md', 15),
            ('DESIGN', 'AUTH-SYSTEM-001', 'design.md', 5),
            ('TASK', 'API-IMPL-001', 'api.md', 20),
            ('TEST', 'UNIT-AUTH-001', 'test.md', 30)
        ]

        for category, identifier, filename, line in test_data:
            self.db_manager.insert_tag(
                category=category,
                identifier=identifier,
                description=f'{category} 설명',
                file_path=f'/path/{filename}',
                line_number=line
            )

        # WHEN: 복합 검색 수행
        # 1. 특정 파일의 TAG들
        login_tags = self.db_manager.search_tags_by_file('/path/login.md')

        # 2. 카테고리와 식별자 패턴으로 검색
        auth_tags = self.db_manager.search_tags_by_pattern('AUTH')

        # 3. 줄 번호 범위로 검색
        range_tags = self.db_manager.search_tags_by_line_range(10, 20)

        # THEN: 정확한 검색 결과 확인
        assert len(login_tags) == 2
        assert all(tag['file_path'] == '/path/login.md' for tag in login_tags)

        assert len(auth_tags) >= 3  # USER-AUTH-001, USER-AUTH-002, AUTH-SYSTEM-001
        assert all('AUTH' in tag['identifier'] for tag in auth_tags)

        assert len(range_tags) == 3  # line 10, 15, 20
        assert all(10 <= tag['line_number'] <= 20 for tag in range_tags)

    def test_should_measure_memory_usage_improvement(self):
        """
        Given: 대용량 TAG 데이터셋
        When: SQLite 저장 및 조회를 수행할 때
        Then: 50% 메모리 사용량 감소 목표를 달성해야 함
        """
        # GIVEN: 메모리 사용량 측정을 위한 대용량 데이터
        self.db_manager.initialize()

        import psutil
        import os

        # 초기 메모리 사용량 측정
        process = psutil.Process(os.getpid())
        initial_memory = process.memory_info().rss

        # 대용량 TAG 데이터 생성 (5000개)
        for i in range(5000):
            category = ['REQ', 'DESIGN', 'TASK', 'TEST'][i % 4]
            self.db_manager.insert_tag(
                category=category,
                identifier=f'MEMORY-TEST-{i:05d}',
                description=f'메모리 테스트용 긴 설명 문자열 ' * 10,  # 긴 설명
                file_path=f'/very/long/path/to/file/number_{i}.md',
                line_number=i % 1000 + 1
            )

        # WHEN: 전체 데이터 조회
        all_tags = self.db_manager.get_all_tags()

        # 현재 메모리 사용량 측정
        current_memory = process.memory_info().rss
        memory_increase = current_memory - initial_memory

        # THEN: 메모리 효율성 검증
        # 5000개 TAG에 대해 50MB 미만 메모리 증가 목표
        assert memory_increase < 50 * 1024 * 1024, f"메모리 증가량: {memory_increase / 1024 / 1024:.2f}MB"
        assert len(all_tags) == 5000

    def test_should_handle_concurrent_database_access(self):
        """
        Given: 여러 스레드에서 동시 데이터베이스 접근
        When: 동시에 TAG 삽입/조회를 수행할 때
        Then: 데이터 무결성이 유지되어야 함
        """
        import threading
        import queue

        # GIVEN: 동시 접근을 위한 설정
        self.db_manager.initialize()
        results = queue.Queue()
        errors = queue.Queue()

        def insert_tags(thread_id: int, count: int):
            try:
                for i in range(count):
                    tag_id = self.db_manager.insert_tag(
                        category='TEST',
                        identifier=f'CONCURRENT-{thread_id:02d}-{i:03d}',
                        description=f'스레드 {thread_id}에서 생성된 TAG {i}',
                        file_path=f'/thread_{thread_id}/file_{i}.md',
                        line_number=i + 1
                    )
                    results.put(tag_id)
            except Exception as e:
                errors.put(e)

        # WHEN: 5개 스레드에서 각각 100개 TAG 동시 삽입
        threads = []
        for thread_id in range(5):
            t = threading.Thread(target=insert_tags, args=(thread_id, 100))
            threads.append(t)
            t.start()

        # 모든 스레드 완료 대기
        for t in threads:
            t.join()

        # THEN: 동시성 처리 검증
        assert errors.qsize() == 0, f"오류 발생: {list(errors.queue)}"
        assert results.qsize() == 500  # 5 * 100

        # 모든 TAG가 올바르게 저장되었는지 확인
        all_tags = self.db_manager.get_all_tags()
        assert len(all_tags) == 500

        # 중복 확인
        identifiers = [tag['identifier'] for tag in all_tags]
        assert len(identifiers) == len(set(identifiers))  # 중복 없음

    def test_should_handle_database_connection_failures(self):
        """
        Given: 데이터베이스 연결 문제 상황
        When: 연결 실패나 손상된 데이터베이스에 접근할 때
        Then: 적절한 예외를 발생시키고 복구 방법을 제시해야 함
        """
        # GIVEN: 잘못된 데이터베이스 경로
        invalid_db_path = Path('/invalid/path/database.db')
        invalid_manager = TagDatabaseManager(invalid_db_path)

        # WHEN & THEN: 초기화 실패 처리
        with pytest.raises(DatabaseError) as exc_info:
            invalid_manager.initialize()

        assert "데이터베이스 생성 실패" in str(exc_info.value)

        # 손상된 데이터베이스 파일 시뮬레이션
        corrupted_db = Path(tempfile.mktemp(suffix='.db'))
        corrupted_db.write_bytes(b'invalid sqlite data')

        corrupted_manager = TagDatabaseManager(corrupted_db)
        with pytest.raises(DatabaseError) as exc_info:
            corrupted_manager.initialize()

        assert "데이터베이스 손상" in str(exc_info.value)

    def test_should_optimize_with_prepared_statements(self):
        """
        Given: 반복되는 쿼리 패턴
        When: prepared statement를 사용할 때
        Then: 쿼리 성능이 향상되어야 함
        """
        # GIVEN: 데이터베이스 초기화
        self.db_manager.initialize()

        # WHEN: prepared statement를 사용한 대량 삽입
        categories = ['REQ', 'DESIGN', 'TASK', 'TEST']

        start_time = time.time()
        with self.db_manager.prepared_insert() as inserter:
            for i in range(1000):
                inserter.execute(
                    category=categories[i % 4],
                    identifier=f'PREPARED-{i:04d}',
                    description=f'Prepared statement 테스트 {i}',
                    file_path=f'/batch/file_{i}.md',
                    line_number=i + 1
                )
        batch_time = time.time() - start_time

        # 일반 삽입과 비교
        start_time = time.time()
        for i in range(100):  # 더 적은 수로 테스트
            self.db_manager.insert_tag(
                category=categories[i % 4],
                identifier=f'NORMAL-{i:04d}',
                description=f'일반 삽입 테스트 {i}',
                file_path=f'/normal/file_{i}.md',
                line_number=i + 1
            )
        normal_time = time.time() - start_time

        # THEN: Prepared statement가 더 효율적이어야 함
        # 1000개 대비 100개이므로 10배 보정
        normalized_normal_time = normal_time * 10
        assert batch_time < normalized_normal_time, \
            f"Prepared: {batch_time:.3f}s, Normal: {normalized_normal_time:.3f}s"

        # 결과 개수 확인
        all_tags = self.db_manager.get_all_tags()
        assert len(all_tags) == 1100  # 1000 + 100