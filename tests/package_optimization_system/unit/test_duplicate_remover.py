"""
Duplicate Remover Unit Tests

@TEST:UNIT-FEATURE-002 - 중복 파일 제거 로직 테스트
@REQ:OPT-DEDUPE-002 - 중복 제거 자동화 검증
"""

import pytest
import tempfile
import hashlib
from pathlib import Path
from unittest.mock import Mock, patch

from package_optimization_system.core.duplicate_remover import DuplicateRemover


class TestDuplicateRemover:
    """DuplicateRemover 클래스 단위 테스트"""

    def setup_method(self):
        """각 테스트 전 설정"""
        self.temp_dir = tempfile.mkdtemp()
        self.remover = DuplicateRemover(self.temp_dir)

    def teardown_method(self):
        """각 테스트 후 정리"""
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def test_should_initialize_with_valid_directory(self):
        """
        Given: 유효한 디렉터리 경로가 주어졌을 때
        When: DuplicateRemover를 초기화하면
        Then: 정상적으로 초기화되어야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Act & Assert
        remover = DuplicateRemover(self.temp_dir)
        assert remover.target_directory == self.temp_dir
        assert remover.hash_algorithm == "sha256"

    def test_should_calculate_file_hash_correctly(self):
        """
        Given: 파일이 주어졌을 때
        When: 파일 해시를 계산하면
        Then: 정확한 SHA256 해시를 반환해야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Arrange
        test_file = Path(self.temp_dir) / "test.txt"
        content = "Hello, World!"
        test_file.write_text(content)

        expected_hash = hashlib.sha256(content.encode()).hexdigest()

        # Act
        actual_hash = self.remover.calculate_file_hash(str(test_file))

        # Assert
        assert actual_hash == expected_hash

    def test_should_identify_duplicate_files(self):
        """
        Given: 동일한 내용의 파일들이 여러 개 있을 때
        When: 중복 파일을 탐지하면
        Then: 중복 파일 그룹을 정확히 식별해야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Arrange
        content1 = "Duplicate content"
        content2 = "Unique content"

        # 중복 파일들
        (Path(self.temp_dir) / "dup1.txt").write_text(content1)
        (Path(self.temp_dir) / "dup2.txt").write_text(content1)
        (Path(self.temp_dir) / "dup3.txt").write_text(content1)

        # 유니크 파일
        (Path(self.temp_dir) / "unique.txt").write_text(content2)

        # Act
        duplicates = self.remover.find_duplicates()

        # Assert
        assert len(duplicates) == 1  # 하나의 중복 그룹
        duplicate_group = duplicates[0]
        assert len(duplicate_group["files"]) == 3  # 3개의 중복 파일
        assert duplicate_group["hash"] == hashlib.sha256(content1.encode()).hexdigest()

    def test_should_preserve_one_file_per_duplicate_group(self):
        """
        Given: 중복 파일 그룹이 있을 때
        When: 중복 제거를 실행하면
        Then: 각 그룹에서 하나의 파일만 보존되어야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Arrange
        content = "Duplicate content"
        files = []
        for i in range(5):
            file_path = Path(self.temp_dir) / f"dup{i}.txt"
            file_path.write_text(content)
            files.append(str(file_path))

        # Act
        result = self.remover.remove_duplicates()

        # Assert
        remaining_files = list(Path(self.temp_dir).glob("*.txt"))
        assert len(remaining_files) == 1  # 하나만 남아야 함
        assert result["removed_count"] == 4  # 4개 제거
        assert result["saved_bytes"] > 0

    def test_should_choose_best_file_to_preserve(self):
        """
        Given: 중복 파일들이 있을 때
        When: 중복 제거를 실행하면
        Then: 가장 적합한 파일을 보존해야 한다 (예: 가장 짧은 경로)
        @TEST:UNIT-FEATURE-002
        """
        # Arrange
        content = "Duplicate content"

        # 다양한 경로 길이의 파일들
        (Path(self.temp_dir) / "a.txt").write_text(content)
        subdir = Path(self.temp_dir) / "subdir"
        subdir.mkdir()
        (subdir / "very_long_filename.txt").write_text(content)

        # Act
        self.remover.remove_duplicates()

        # Assert
        remaining_files = list(Path(self.temp_dir).rglob("*.txt"))
        assert len(remaining_files) == 1
        # 더 짧은 경로의 파일이 보존되어야 함
        assert remaining_files[0].name == "a.txt"

    def test_should_handle_empty_directory(self):
        """
        Given: 빈 디렉터리가 주어졌을 때
        When: 중복 탐지를 실행하면
        Then: 빈 결과를 반환해야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Act
        duplicates = self.remover.find_duplicates()
        result = self.remover.remove_duplicates()

        # Assert
        assert len(duplicates) == 0
        assert result["removed_count"] == 0
        assert result["saved_bytes"] == 0

    def test_should_ignore_specified_file_types(self):
        """
        Given: 제외할 파일 타입이 설정되었을 때
        When: 중복 탐지를 실행하면
        Then: 지정된 파일 타입은 무시되어야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Arrange
        content = "Duplicate content"

        # 제외할 파일 타입 설정
        self.remover.exclude_extensions = [".log", ".tmp"]

        # 다양한 확장자 파일 생성
        (Path(self.temp_dir) / "dup1.txt").write_text(content)
        (Path(self.temp_dir) / "dup2.txt").write_text(content)
        (Path(self.temp_dir) / "ignore1.log").write_text(content)
        (Path(self.temp_dir) / "ignore2.tmp").write_text(content)

        # Act
        duplicates = self.remover.find_duplicates()

        # Assert
        assert len(duplicates) == 1  # .txt 파일들만 중복으로 감지
        duplicate_group = duplicates[0]
        assert len(duplicate_group["files"]) == 2  # .txt 파일 2개만

        # .log, .tmp 파일이 포함되지 않았는지 확인
        all_files = [f for group in duplicates for f in group["files"]]
        assert not any(f.endswith(('.log', '.tmp')) for f in all_files)

    def test_should_handle_large_files_efficiently(self):
        """
        Given: 큰 파일들이 있을 때
        When: 중복 탐지를 실행하면
        Then: 메모리 효율적으로 처리되어야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Arrange
        # 큰 파일 시뮬레이션을 위한 mock
        with patch.object(self.remover, 'calculate_file_hash') as mock_hash:
            mock_hash.side_effect = ["hash1", "hash1", "hash2"]

            # 가상의 큰 파일들
            large_files = [
                Path(self.temp_dir) / "large1.bin",
                Path(self.temp_dir) / "large2.bin",
                Path(self.temp_dir) / "large3.bin"
            ]

            for file_path in large_files:
                file_path.write_text("dummy")

            # Act
            duplicates = self.remover.find_duplicates()

            # Assert
            assert len(duplicates) == 1  # 하나의 중복 그룹
            assert mock_hash.call_count == 3  # 각 파일마다 해시 계산

    def test_should_report_detailed_metrics(self):
        """
        Given: 중복 제거 프로세스가 실행될 때
        When: 메트릭 수집이 활성화되어 있으면
        Then: 상세한 중복 제거 메트릭이 보고되어야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Arrange
        content = "Duplicate content"
        for i in range(3):
            (Path(self.temp_dir) / f"dup{i}.txt").write_text(content)

        # Act
        result = self.remover.remove_duplicates()

        # Assert
        assert "removed_count" in result
        assert "saved_bytes" in result
        assert "duplicate_groups" in result
        assert "processing_time" in result
        assert isinstance(result["processing_time"], float)

    def test_should_handle_permission_errors_gracefully(self):
        """
        Given: 권한이 없는 파일이 있을 때
        When: 중복 제거를 실행하면
        Then: 에러를 적절히 처리하고 계속 진행해야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Arrange
        content = "Duplicate content"
        (Path(self.temp_dir) / "file1.txt").write_text(content)
        (Path(self.temp_dir) / "file2.txt").write_text(content)

        # 권한 에러 시뮬레이션
        with patch('os.remove', side_effect=PermissionError("Permission denied")):
            # Act
            result = self.remover.remove_duplicates()

            # Assert
            assert "errors" in result
            assert len(result["errors"]) > 0
            assert "permission" in result["errors"][0].lower()

    def test_should_maintain_file_integrity(self):
        """
        Given: 중복 파일 제거 후
        When: 남은 파일을 검증하면
        Then: 파일 내용이 손상되지 않았어야 한다
        @TEST:UNIT-FEATURE-002
        """
        # Arrange
        original_content = "Original content that must be preserved"
        original_hash = hashlib.sha256(original_content.encode()).hexdigest()

        # 중복 파일들 생성
        for i in range(3):
            (Path(self.temp_dir) / f"dup{i}.txt").write_text(original_content)

        # Act
        self.remover.remove_duplicates()

        # Assert
        remaining_files = list(Path(self.temp_dir).glob("*.txt"))
        assert len(remaining_files) == 1

        # 남은 파일의 내용 검증
        remaining_content = remaining_files[0].read_text()
        remaining_hash = hashlib.sha256(remaining_content.encode()).hexdigest()
        assert remaining_hash == original_hash
        assert remaining_content == original_content