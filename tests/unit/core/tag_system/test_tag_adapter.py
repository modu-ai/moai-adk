"""
@TEST:SPEC-009-TAG-ADAPTER-001 - TAG JSON API 호환성 어댑터 실패 테스트

RED 단계: SQLite 백엔드와 기존 JSON API 100% 호환성 실패 테스트
"""

import pytest
import json
import tempfile
from pathlib import Path
from typing import Dict, List, Any
from unittest.mock import Mock, patch, MagicMock

# 아직 구현되지 않은 모듈들 - 실패할 예정
from moai_adk.core.tag_system.adapter import (
    TagIndexAdapter,
    ApiCompatibilityError,
    AdapterConfiguration
)
from moai_adk.core.tag_system.database import TagDatabaseManager


class TestTagIndexAdapter:
    """TAG JSON API 호환성 어댑터 테스트 스위트"""

    def setup_method(self):
        """각 테스트 전 초기화"""
        self.temp_db = Path(tempfile.mktemp(suffix='.db'))
        self.temp_json = Path(tempfile.mktemp(suffix='.json'))

        # 어댑터 초기화 (SQLite 백엔드 사용)
        self.adapter = TagIndexAdapter(
            database_path=self.temp_db,
            json_fallback_path=self.temp_json
        )

    def teardown_method(self):
        """각 테스트 후 정리"""
        self.adapter.close()
        if self.temp_db.exists():
            self.temp_db.unlink()
        if self.temp_json.exists():
            self.temp_json.unlink()

    def test_should_maintain_exact_json_api_compatibility(self):
        """
        Given: 기존 TagIndexManager의 JSON API 구조
        When: TagIndexAdapter를 통해 동일한 메서드를 호출할 때
        Then: 완전히 동일한 JSON 구조를 반환해야 함
        """
        # GIVEN: 기존 JSON API 호환 데이터 준비
        self.adapter.initialize()

        # 기존 API와 동일한 메서드 존재 확인
        expected_methods = [
            'load_index',
            'save_index',
            'initialize_index',
            'validate_index_schema',
            'process_file_change',
            'start_watching',
            'stop_watching'
        ]

        # WHEN: 어댑터 메서드 존재 확인
        for method_name in expected_methods:
            assert hasattr(self.adapter, method_name), f"Missing method: {method_name}"

        # THEN: load_index 반환 구조가 기존과 동일해야 함
        index_data = self.adapter.load_index()

        # 기존 JSON 구조와 정확히 일치해야 함
        expected_structure_keys = [
            "metadata", "categories", "chains", "files"
        ]
        for key in expected_structure_keys:
            assert key in index_data, f"Missing key: {key}"

        # metadata 구조 검증
        metadata = index_data["metadata"]
        expected_metadata_keys = ["created_at", "updated_at", "version", "total_tags"]
        for key in expected_metadata_keys:
            assert key in metadata, f"Missing metadata key: {key}"

        # categories 구조 검증 (16-Core TAG 시스템)
        categories = index_data["categories"]
        expected_category_groups = ["PRIMARY", "STEERING", "IMPLEMENTATION", "QUALITY"]
        for group in expected_category_groups:
            assert group in categories, f"Missing category group: {group}"

    def test_should_process_file_change_like_original_api(self):
        """
        Given: 기존 TagIndexManager.process_file_change와 동일한 시그니처
        When: 파일 변경을 처리할 때
        Then: 동일한 방식으로 인덱스가 업데이트되어야 함
        """
        # GIVEN: 어댑터 초기화 및 테스트 파일
        self.adapter.initialize()
        test_file = Path(tempfile.mktemp(suffix='.md'))
        test_content = """
        # 요구사항 문서

        @REQ:USER-AUTH-001 사용자 인증 기능
        @DESIGN:JWT-SYSTEM-001 JWT 기반 인증 시스템
        """
        test_file.write_text(test_content)

        # WHEN: 파일 변경 처리 (기존 API와 동일한 시그니처)
        self.adapter.process_file_change(test_file, "created")

        # THEN: 기존 JSON 형식과 동일한 결과 생성
        index_data = self.adapter.load_index()

        # 총 TAG 수 확인
        assert index_data["metadata"]["total_tags"] == 2

        # categories 구조가 기존과 동일해야 함
        primary_categories = index_data["categories"]["PRIMARY"]

        # REQ 카테고리 확인
        assert "REQ" in primary_categories
        req_tags = primary_categories["REQ"]
        assert "USER-AUTH-001" in req_tags
        assert req_tags["USER-AUTH-001"]["description"] == "사용자 인증 기능"
        assert req_tags["USER-AUTH-001"]["file"] == str(test_file)

        # DESIGN 카테고리 확인
        assert "DESIGN" in primary_categories
        design_tags = primary_categories["DESIGN"]
        assert "JWT-SYSTEM-001" in design_tags

        # files 구조 확인 (기존과 동일)
        files = index_data["files"]
        assert str(test_file) in files
        file_tags = files[str(test_file)]
        assert len(file_tags) == 2
        assert file_tags[0]["category"] == "REQ"
        assert file_tags[1]["category"] == "DESIGN"

        # 정리
        test_file.unlink()

    def test_should_support_callback_system_like_original(self):
        """
        Given: 기존 TagIndexManager의 콜백 시스템
        When: on_file_changed 콜백을 설정할 때
        Then: 기존과 동일하게 이벤트가 호출되어야 함
        """
        # GIVEN: 콜백 함수 설정
        self.adapter.initialize()
        received_events = []

        def callback_handler(event):
            received_events.append(event)

        self.adapter.on_file_changed = callback_handler

        # WHEN: 파일 변경 이벤트 발생
        test_file = Path(tempfile.mktemp(suffix='.md'))
        test_file.write_text("@REQ:CALLBACK-TEST-001 콜백 테스트")

        self.adapter.process_file_change(test_file, "created")

        # THEN: 콜백이 기존 형식으로 호출됨
        assert len(received_events) == 1

        event = received_events[0]
        # 기존 IndexUpdateEvent와 동일한 구조
        assert hasattr(event, 'event_type')
        assert hasattr(event, 'file_path')
        assert hasattr(event, 'timestamp')

        assert event.event_type == "created"
        assert event.file_path == test_file

        # 정리
        test_file.unlink()

    def test_should_validate_schema_exactly_like_original(self):
        """
        Given: 기존 TagIndexManager의 스키마 검증 로직
        When: validate_index_schema를 호출할 때
        Then: 동일한 스키마 검증 결과를 반환해야 함
        """
        # GIVEN: 올바른 스키마와 잘못된 스키마
        valid_schema = {
            "metadata": {
                "created_at": "2024-01-01T00:00:00Z",
                "updated_at": "2024-01-01T00:00:00Z",
                "version": "1.0",
                "total_tags": 0
            },
            "categories": {
                "PRIMARY": {},
                "STEERING": {},
                "IMPLEMENTATION": {},
                "QUALITY": {}
            },
            "chains": [],
            "files": {}
        }

        invalid_schema = {
            "metadata": {
                "version": "1.0"  # 필수 필드 누락
            },
            "categories": "invalid_type"  # 잘못된 타입
        }

        # WHEN & THEN: 스키마 검증이 기존과 동일해야 함
        assert self.adapter.validate_index_schema(valid_schema) is True
        assert self.adapter.validate_index_schema(invalid_schema) is False

    def test_should_handle_json_fallback_mode(self):
        """
        Given: SQLite 데이터베이스가 사용 불가능한 상황
        When: JSON fallback 모드로 동작할 때
        Then: 기존 JSON 파일 기반 동작과 완전히 동일해야 함
        """
        # GIVEN: SQLite 사용 불가능 시뮬레이션
        with patch.object(self.adapter, '_sqlite_available', False):
            self.adapter.initialize()

            # WHEN: JSON fallback 모드로 동작
            test_file = Path(tempfile.mktemp(suffix='.md'))
            test_file.write_text("@REQ:FALLBACK-001 fallback 모드 테스트")

            self.adapter.process_file_change(test_file, "created")

            # THEN: JSON 파일이 생성되고 기존 형식과 동일
            assert self.temp_json.exists()

            with open(self.temp_json) as f:
                json_data = json.load(f)

            # 기존 JSON 구조와 완전 동일
            assert "metadata" in json_data
            assert "categories" in json_data
            assert "chains" in json_data
            assert "files" in json_data

            assert json_data["metadata"]["total_tags"] == 1

            # 정리
            test_file.unlink()

    def test_should_maintain_watching_interface_compatibility(self):
        """
        Given: 기존 파일 감시 인터페이스
        When: start_watching/stop_watching을 호출할 때
        Then: 기존과 동일한 동작을 수행해야 함
        """
        # GIVEN: 어댑터 초기화
        self.adapter.initialize()

        # WHEN: 감시 상태 확인 (기존 API와 동일)
        assert hasattr(self.adapter, 'is_watching')
        assert self.adapter.is_watching is False

        # 감시 시작
        self.adapter.start_watching()
        assert self.adapter.is_watching is True

        # 감시 중지
        self.adapter.stop_watching()
        assert self.adapter.is_watching is False

    def test_should_preserve_performance_characteristics(self):
        """
        Given: 성능이 중요한 대용량 TAG 처리
        When: SQLite 백엔드를 통해 처리할 때
        Then: 기존 JSON 방식보다 빠르면서도 API 호환성 유지해야 함
        """
        import time

        # GIVEN: 대용량 TAG 데이터 준비
        self.adapter.initialize()

        # 1000개 TAG가 포함된 파일 생성
        large_content = []
        for i in range(1000):
            category = ['REQ', 'DESIGN', 'TASK', 'TEST'][i % 4]
            large_content.append(f"@{category}:PERF-TEST-{i:04d} 성능 테스트 TAG {i}")

        test_file = Path(tempfile.mktemp(suffix='.md'))
        test_file.write_text('\n'.join(large_content))

        # WHEN: 대용량 파일 처리 성능 측정
        start_time = time.time()
        self.adapter.process_file_change(test_file, "created")
        processing_time = time.time() - start_time

        # THEN: 성능 요구사항 만족
        # JSON 방식 대비 10x 성능 개선 목표 (2초 → 0.2초)
        assert processing_time < 0.5, f"처리 시간 {processing_time:.3f}초가 목표를 초과"

        # API 호환성 확인 - 정확한 결과 반환
        index_data = self.adapter.load_index()
        assert index_data["metadata"]["total_tags"] == 1000

        # 기존 형식과 동일한 구조
        assert "categories" in index_data
        assert "files" in index_data
        assert str(test_file) in index_data["files"]

        # 정리
        test_file.unlink()

    def test_should_support_migration_between_backends(self):
        """
        Given: 기존 JSON 데이터와 새로운 SQLite 백엔드
        When: 백엔드 간 전환이 필요할 때
        Then: 데이터 손실 없이 전환되고 API 호환성 유지해야 함
        """
        # GIVEN: 기존 JSON 형식 데이터 준비
        json_data = {
            "metadata": {
                "created_at": "2024-01-01T00:00:00Z",
                "updated_at": "2024-01-01T00:00:00Z",
                "version": "1.0",
                "total_tags": 2
            },
            "categories": {
                "PRIMARY": {
                    "REQ": {
                        "USER-MIGRATION-001": {
                            "description": "마이그레이션 테스트",
                            "file": "/test/migration.md"
                        }
                    },
                    "DESIGN": {
                        "ARCH-MIGRATION-001": {
                            "description": "아키텍처 마이그레이션",
                            "file": "/test/migration.md"
                        }
                    }
                }
            },
            "chains": [],
            "files": {
                "/test/migration.md": [
                    {
                        "category": "REQ",
                        "identifier": "USER-MIGRATION-001",
                        "description": "마이그레이션 테스트"
                    },
                    {
                        "category": "DESIGN",
                        "identifier": "ARCH-MIGRATION-001",
                        "description": "아키텍처 마이그레이션"
                    }
                ]
            }
        }

        with open(self.temp_json, 'w', encoding='utf-8') as f:
            json.dump(json_data, f, indent=2, ensure_ascii=False)

        # WHEN: JSON에서 SQLite로 마이그레이션
        self.adapter.migrate_from_json(self.temp_json)

        # THEN: 데이터 손실 없이 마이그레이션 완료
        migrated_data = self.adapter.load_index()

        # 기존 데이터와 완전히 동일
        assert migrated_data["metadata"]["total_tags"] == 2
        assert "USER-MIGRATION-001" in migrated_data["categories"]["PRIMARY"]["REQ"]
        assert "ARCH-MIGRATION-001" in migrated_data["categories"]["PRIMARY"]["DESIGN"]

        # 역방향 마이그레이션 테스트
        export_json = Path(tempfile.mktemp(suffix='.json'))
        self.adapter.export_to_json(export_json)

        with open(export_json, 'r', encoding='utf-8') as f:
            exported_data = json.load(f)

        # 원본 데이터와 동일해야 함
        assert exported_data["metadata"]["total_tags"] == 2
        assert exported_data["categories"]["PRIMARY"]["REQ"]["USER-MIGRATION-001"]["description"] == "마이그레이션 테스트"

        # 정리
        export_json.unlink()

    def test_should_handle_concurrent_access_transparently(self):
        """
        Given: 기존 API 사용 코드에서 동시 접근
        When: 여러 스레드에서 어댑터를 사용할 때
        Then: 기존 API 동작 방식과 동일하게 동시성 처리해야 함
        """
        import threading
        import queue

        # GIVEN: 동시 접근 테스트 준비
        self.adapter.initialize()
        results = queue.Queue()
        errors = queue.Queue()

        def worker_thread(thread_id: int):
            try:
                # 각 스레드에서 파일 생성 및 처리
                for i in range(50):
                    test_file = Path(tempfile.mktemp(suffix='.md'))
                    test_file.write_text(f"@REQ:CONCURRENT-{thread_id:02d}-{i:02d} 동시 접근 테스트")

                    # 기존 API와 동일한 방식으로 호출
                    self.adapter.process_file_change(test_file, "created")

                    results.put((thread_id, i))
                    test_file.unlink()

            except Exception as e:
                errors.put(e)

        # WHEN: 5개 스레드에서 동시 실행
        threads = []
        for thread_id in range(5):
            t = threading.Thread(target=worker_thread, args=(thread_id,))
            threads.append(t)
            t.start()

        for t in threads:
            t.join()

        # THEN: 기존 API처럼 안전한 동시성 처리
        assert errors.qsize() == 0, f"동시성 오류 발생: {list(errors.queue)}"
        assert results.qsize() == 250  # 5 * 50

        # 최종 상태 확인
        final_index = self.adapter.load_index()
        # 모든 스레드 완료 후 일관된 상태여야 함
        assert "metadata" in final_index
        assert "categories" in final_index

    def test_should_provide_debugging_and_introspection_like_original(self):
        """
        Given: 기존 디버깅 및 내부 상태 조회 기능
        When: 어댑터 내부 상태를 조회할 때
        Then: 기존과 동일한 디버깅 인터페이스를 제공해야 함
        """
        # GIVEN: 어댑터 초기화 및 데이터 준비
        self.adapter.initialize()

        test_file = Path(tempfile.mktemp(suffix='.md'))
        test_file.write_text("@REQ:DEBUG-001 디버깅 테스트")
        self.adapter.process_file_change(test_file, "created")

        # WHEN: 내부 상태 조회 (기존 API 메서드들)
        # 프로퍼티 기반 상태 확인
        assert hasattr(self.adapter, 'is_watching')

        # 설정 정보 조회 (새로운 디버깅 기능)
        config_info = self.adapter.get_configuration_info()

        # THEN: 유용한 디버깅 정보 제공
        assert 'backend_type' in config_info  # 'sqlite' 또는 'json'
        assert 'database_path' in config_info
        assert 'performance_stats' in config_info

        # 성능 통계 확인
        perf_stats = config_info['performance_stats']
        assert 'total_tags' in perf_stats
        assert 'query_count' in perf_stats
        assert 'avg_query_time' in perf_stats

        # 정리
        test_file.unlink()

    def test_should_gracefully_degrade_on_errors(self):
        """
        Given: SQLite 백엔드에서 오류 발생 상황
        When: 데이터베이스 연결 실패나 손상이 발생할 때
        Then: 기존 API처럼 JSON fallback으로 graceful degradation 수행해야 함
        """
        # GIVEN: SQLite 오류 시뮬레이션
        self.adapter.initialize()

        # WHEN: 데이터베이스 손상 시뮬레이션
        with patch.object(self.adapter, '_database') as mock_db:
            mock_db.insert_tag.side_effect = Exception("Database connection lost")

            # 오류 발생 시에도 JSON fallback으로 동작해야 함
            test_file = Path(tempfile.mktemp(suffix='.md'))
            test_file.write_text("@REQ:ERROR-HANDLING-001 오류 처리 테스트")

            # 예외 발생하지 않고 fallback으로 처리
            self.adapter.process_file_change(test_file, "created")

        # THEN: 기본 기능은 계속 동작
        # JSON fallback이 작동했는지 확인
        assert self.temp_json.exists()

        index_data = self.adapter.load_index()
        assert "metadata" in index_data
        assert index_data["metadata"]["total_tags"] >= 0  # 최소한 구조는 유지

        # 정리
        test_file.unlink()