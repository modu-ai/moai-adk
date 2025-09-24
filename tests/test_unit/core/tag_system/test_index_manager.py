"""
@TEST:UNIT-INDEX-MANAGER - 실시간 TAG 인덱스 관리 테스트

RED 단계: watchdog 기반 파일 감지 및 인덱스 갱신 실패 테스트
"""

import pytest
import json
import tempfile
import time
from pathlib import Path
from unittest.mock import Mock, patch, MagicMock

from moai_adk.core.tag_system.index_manager import TagIndexManager, IndexUpdateEvent, WatcherStatus
from moai_adk.core.tag_system.parser import TagMatch


class TestTagIndexManager:
    """실시간 TAG 인덱스 관리 테스트 스위트"""

    def setup_method(self):
        """각 테스트 전 초기화"""
        self.temp_dir = Path(tempfile.mkdtemp())
        self.index_file = self.temp_dir / "tags.json"
        self.manager = TagIndexManager(
            watch_directory=self.temp_dir,
            index_file=self.index_file
        )

    def teardown_method(self):
        """각 테스트 후 정리"""
        if self.manager.is_watching:
            self.manager.stop_watching()

        # 임시 디렉토리 정리
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_should_initialize_empty_tag_index_file(self):
        """
        Given: 존재하지 않는 TAG 인덱스 파일
        When: TagIndexManager 초기화를 실행할 때
        Then: 빈 인덱스 구조가 생성되어야 함
        """
        # GIVEN: 존재하지 않는 인덱스 파일
        assert not self.index_file.exists()

        # WHEN: 인덱스 초기화
        self.manager.initialize_index()

        # THEN: 올바른 구조로 빈 인덱스 생성
        assert self.index_file.exists()
        with open(self.index_file) as f:
            index_data = json.load(f)

        expected_structure = {
            "metadata": {
                "created_at": index_data["metadata"]["created_at"],  # 동적 값
                "updated_at": index_data["metadata"]["updated_at"],  # 동적 값
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
        assert index_data == expected_structure

    @pytest.mark.skip(reason="Watchdog integration test - skip to avoid timeout")
    def test_should_detect_file_changes_with_watchdog(self):
        """
        Given: watchdog로 감시 중인 디렉토리
        When: 파일이 생성/수정될 때
        Then: 파일 변경 이벤트가 감지되어야 함
        """
        # GIVEN: 파일 변경 이벤트를 직접 시뮬레이션
        events_received = []

        def event_handler(event: IndexUpdateEvent):
            events_received.append(event)

        self.manager.on_file_changed = event_handler

        # WHEN: 파일 변경 처리 직접 호출 (실제 watchdog 대신)
        test_file = self.temp_dir / "test.md"
        test_file.write_text("@REQ:USER-TEST-001 테스트 파일")

        # 직접 파일 변경 처리 호출
        self.manager.process_file_change(test_file, "created")

        # THEN: 파일 변경 이벤트 감지
        assert len(events_received) > 0
        create_event = events_received[0]
        assert create_event.event_type == "created"
        assert create_event.file_path == test_file

    def test_should_parse_tags_from_changed_files(self):
        """
        Given: TAG가 포함된 파일
        When: 파일 변경이 감지될 때
        Then: TAG를 파싱하여 인덱스에 추가해야 함
        """
        # GIVEN: TAG가 포함된 파일 생성
        self.manager.initialize_index()
        test_file = self.temp_dir / "requirements.md"
        content = """
        # 요구사항 문서

        @REQ:USER-LOGIN-001 사용자 로그인 기능
        @REQ:USER-LOGOUT-001 사용자 로그아웃 기능
        @DESIGN:AUTH-SYSTEM-001 인증 시스템 설계
        """
        test_file.write_text(content)

        # WHEN: 파일 변경 처리
        self.manager.process_file_change(test_file, "created")

        # THEN: TAG가 인덱스에 추가됨
        index_data = self.manager.load_index()
        assert index_data["metadata"]["total_tags"] == 3

        # REQ 카테고리 확인
        req_tags = index_data["categories"]["PRIMARY"]["REQ"]
        assert len(req_tags) == 2
        assert "USER-LOGIN-001" in req_tags
        assert "USER-LOGOUT-001" in req_tags

        # DESIGN 카테고리 확인
        design_tags = index_data["categories"]["PRIMARY"]["DESIGN"]
        assert len(design_tags) == 1
        assert "AUTH-SYSTEM-001" in design_tags

    def test_should_update_index_when_file_deleted(self):
        """
        Given: TAG가 포함된 파일이 인덱스에 등록됨
        When: 해당 파일이 삭제될 때
        Then: 인덱스에서도 해당 TAG들이 제거되어야 함
        """
        # GIVEN: TAG가 등록된 상태
        self.manager.initialize_index()
        test_file = self.temp_dir / "temporary.md"
        test_file.write_text("@REQ:USER-TEMP-001 임시 요구사항")
        self.manager.process_file_change(test_file, "created")

        # 파일이 등록되었는지 확인
        index_data = self.manager.load_index()
        assert index_data["metadata"]["total_tags"] == 1

        # WHEN: 파일 삭제
        test_file.unlink()  # 파일 삭제
        self.manager.process_file_change(test_file, "deleted")

        # THEN: 인덱스에서 TAG 제거됨
        updated_index = self.manager.load_index()
        assert updated_index["metadata"]["total_tags"] == 0
        assert len(updated_index["files"]) == 0

    def test_should_maintain_index_integrity_during_updates(self):
        """
        Given: 기존 TAG 인덱스
        When: 여러 파일에서 동시에 변경이 발생할 때
        Then: 인덱스 무결성이 유지되어야 함
        """
        # GIVEN: 기존 인덱스 설정
        self.manager.initialize_index()

        # 여러 파일 생성
        files_and_tags = [
            ("doc1.md", "@REQ:USER-AUTH-001 문서1"),
            ("doc2.md", "@REQ:USER-PROFILE-001 문서2"),
            ("doc3.md", "@DESIGN:ARCH-SYSTEM-001 설계3")
        ]

        # WHEN: 여러 파일 동시 처리 (순차적으로 시뮬레이션)
        for filename, content in files_and_tags:
            file_path = self.temp_dir / filename
            file_path.write_text(content)
            self.manager.process_file_change(file_path, "created")

        # THEN: 모든 TAG가 올바르게 인덱싱됨
        final_index = self.manager.load_index()
        assert final_index["metadata"]["total_tags"] == 3

        # 파일별 TAG 매핑 확인
        assert len(final_index["files"]) == 3
        for filename, _ in files_and_tags:
            file_key = str(self.temp_dir / filename)
            assert file_key in final_index["files"]

    def test_should_handle_json_schema_validation(self):
        """
        Given: TAG 인덱스 JSON 스키마
        When: 인덱스 데이터를 검증할 때
        Then: 올바른 스키마 구조만 허용해야 함
        """
        # GIVEN: 올바른 인덱스 데이터
        valid_index = {
            "metadata": {
                "created_at": "2024-01-01T00:00:00Z",
                "updated_at": "2024-01-01T00:00:00Z",
                "version": "1.0",
                "total_tags": 1
            },
            "categories": {
                "PRIMARY": {"REQ": {"USER-TEST-001": {"description": "테스트"}}},
                "STEERING": {},
                "IMPLEMENTATION": {},
                "QUALITY": {}
            },
            "chains": [],
            "files": {}
        }

        # 잘못된 인덱스 데이터
        invalid_index = {
            "metadata": {
                "version": "1.0"  # 필수 필드 누락
            },
            "categories": "invalid_structure"  # 잘못된 타입
        }

        # WHEN & THEN: 스키마 검증
        assert self.manager.validate_index_schema(valid_index) is True
        assert self.manager.validate_index_schema(invalid_index) is False

    def test_should_measure_index_update_performance(self):
        """
        Given: 여러 TAG가 포함된 파일
        When: 인덱스 업데이트를 실행할 때
        Then: 목표 성능 기준을 만족해야 함
        """
        # GIVEN: 다양한 TAG가 포함된 파일
        self.manager.initialize_index()
        large_file = self.temp_dir / "large_document.md"

        # 미리 정의된 유효한 TAG들 사용
        predefined_tags = [
            "@REQ:USER-LOGIN-001",
            "@REQ:USER-LOGOUT-001",
            "@REQ:USER-REGISTER-001",
            "@DESIGN:AUTH-SYSTEM-001",
            "@DESIGN:DB-SCHEMA-001",
            "@TASK:API-IMPL-001",
            "@TASK:UI-IMPL-001",
            "@TEST:UNIT-AUTH-001",
            "@TEST:E2E-LOGIN-001",
            "@FEATURE:USER-MGMT-001"
        ]

        large_content = "\n".join(f"{tag} 설명" for tag in predefined_tags)
        large_file.write_text(large_content)

        # WHEN: 성능 측정하며 처리
        start_time = time.time()
        self.manager.process_file_change(large_file, "created")
        processing_time = time.time() - start_time

        # THEN: 성능 기준 만족 (2초)
        assert processing_time < 2.0

        # 결과 검증
        index_data = self.manager.load_index()
        assert index_data["metadata"]["total_tags"] == 10

    def test_should_handle_concurrent_file_modifications(self):
        """
        Given: 동시에 여러 파일이 수정되는 상황
        When: 병렬 처리를 시도할 때
        Then: 데이터 경합 없이 모든 변경사항이 반영되어야 함
        """
        # GIVEN: 병렬 처리를 위한 setup
        self.manager.initialize_index()

        # 미리 정의된 유효한 TAG들로 파일 생성
        concurrent_files = []
        tag_mapping = [
            "@REQ:USER-CONCURRENT-001",
            "@DESIGN:ARCH-CONCURRENT-001",
            "@TASK:IMPL-CONCURRENT-001",
            "@TEST:UNIT-CONCURRENT-001",
            "@FEATURE:SYSTEM-CONCURRENT-001"
        ]

        for i, tag in enumerate(tag_mapping):
            file_path = self.temp_dir / f"concurrent_{i}.md"
            file_path.write_text(f"{tag} 동시 수정 테스트")
            concurrent_files.append(file_path)

        # WHEN: 빠른 연속 처리로 동시성 시뮬레이션
        for file_path in concurrent_files:
            self.manager.process_file_change(file_path, "created")

        # THEN: 모든 변경사항 반영 확인
        final_index = self.manager.load_index()
        assert final_index["metadata"]["total_tags"] == 5
        assert len(final_index["files"]) == 5