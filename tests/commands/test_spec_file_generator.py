"""
@TEST:SPEC-FILE-GENERATOR-001 SpecFileGenerator Unit Tests
@REQ:TRUST-COMPLIANCE-001 → @DESIGN:MODULE-SPLIT-001 → @TASK:FILE-GENERATOR-001 → @TEST:SPEC-FILE-GENERATOR-001

Tests for SPEC file generation module following TRUST principles:
- T: Test-first development
- R: Readable test structure
- U: Unified file generation responsibility
- S: Secure file operations
- T: Trackable file creation
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import patch
from src.moai_adk.commands.spec_file_generator import SpecFileGenerator


class TestSpecFileGenerator:
    """Test class for SpecFileGenerator following TRUST principles"""

    def setup_method(self):
        """Setup test environment with temporary directory"""
        self.temp_dir = tempfile.mkdtemp()
        self.project_dir = Path(self.temp_dir)
        self.generator = SpecFileGenerator(self.project_dir)

    def teardown_method(self):
        """Cleanup temporary directory"""
        import shutil
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    # Happy Path Tests
    def test_should_create_spec_file_with_valid_inputs(self):
        """Test SPEC file creation with valid inputs (happy path)"""
        # Given
        spec_name = "USER-AUTH"
        description = "User authentication system"

        # When
        result_path = self.generator.create_spec_file(spec_name, description)

        # Then
        assert result_path.exists()
        assert result_path.name == "USER-AUTH.md"
        assert result_path.parent.name == "specs"

        # Verify content structure
        content = result_path.read_text(encoding="utf-8")
        assert f"# {spec_name}" in content
        assert description in content
        assert "## 요구사항" in content
        assert "## 수락 기준" in content
        assert "## 태그 체계" in content

    def test_should_create_specs_directory_if_not_exists(self):
        """Test automatic creation of specs directory"""
        # Given
        spec_name = "TEST-SPEC"
        description = "Test description"
        specs_dir = self.project_dir / ".moai" / "specs"

        # Ensure directory doesn't exist
        assert not specs_dir.exists()

        # When
        self.generator.create_spec_file(spec_name, description)

        # Then
        assert specs_dir.exists()
        assert specs_dir.is_dir()

    def test_should_generate_proper_content_structure(self):
        """Test proper SPEC content structure generation"""
        # Given
        spec_name = "API-DESIGN"
        description = "REST API design specification"

        # When
        content = self.generator.generate_spec_content(spec_name, description)

        # Then
        assert content.startswith(f"# {spec_name}")
        assert description in content
        assert "## 개요" in content
        assert "## 요구사항" in content
        assert "### 기능 요구사항" in content
        assert "### 비기능 요구사항" in content
        assert "## 수락 기준" in content
        assert "## 태그 체계" in content
        assert f"@REQ:{spec_name}-001" in content
        assert f"@DESIGN:{spec_name}-ARCH-001" in content

    def test_should_backup_existing_spec_file(self):
        """Test backup creation for existing SPEC files"""
        # Given
        spec_name = "EXISTING-SPEC"
        description = "Test description"
        specs_dir = self.project_dir / ".moai" / "specs"
        specs_dir.mkdir(parents=True, exist_ok=True)

        existing_file = specs_dir / f"{spec_name}.md"
        existing_content = "Existing content"
        existing_file.write_text(existing_content, encoding="utf-8")

        # When
        self.generator.create_spec_file(spec_name, description)

        # Then
        backup_file = specs_dir / f"{spec_name}.md.backup"
        assert backup_file.exists()
        assert backup_file.read_text(encoding="utf-8") == existing_content

    # Edge Cases
    def test_should_handle_special_characters_in_spec_name(self):
        """Test handling of special characters in spec name for filename"""
        # Given
        spec_name = "USER-AUTH-V2"
        description = "User authentication version 2"

        # When
        result_path = self.generator.create_spec_file(spec_name, description)

        # Then
        assert result_path.exists()
        assert result_path.name == "USER-AUTH-V2.md"

    @patch('src.moai_adk.commands.spec_file_generator.datetime')
    def test_should_include_timestamp_in_content(self, mock_datetime):
        """Test inclusion of timestamp in generated content"""
        # Given
        mock_datetime.datetime.now.return_value.strftime.return_value = "2024-01-01 12:00:00"
        spec_name = "TIME-TEST"
        description = "Timestamp test"

        # When
        content = self.generator.generate_spec_content(spec_name, description)

        # Then
        assert "생성 일시: 2024-01-01 12:00:00" in content

    def test_should_include_mode_in_content(self):
        """Test inclusion of current mode in generated content"""
        # Given
        spec_name = "MODE-TEST"
        description = "Mode test"
        generator = SpecFileGenerator(self.project_dir, current_mode="team")

        # When
        content = generator.generate_spec_content(spec_name, description)

        # Then
        assert "모드: team" in content

    # Error Cases
    def test_should_handle_file_creation_permission_error(self):
        """Test handling of file creation permission errors"""
        # Given
        spec_name = "PERM-TEST"
        description = "Permission test"

        # Create generator with non-existent/non-writable directory
        invalid_dir = Path("/invalid/path/that/does/not/exist")
        generator = SpecFileGenerator(invalid_dir)

        # When & Then
        with pytest.raises(ValueError, match="SPEC 파일 생성 실패"):
            generator.create_spec_file(spec_name, description)

    def test_should_handle_unicode_content_properly(self):
        """Test proper handling of Unicode content"""
        # Given
        spec_name = "유니코드테스트"
        description = "한글과 🚀 이모지를 포함한 설명"

        # When
        result_path = self.generator.create_spec_file(spec_name, description)

        # Then
        content = result_path.read_text(encoding="utf-8")
        assert spec_name in content
        assert description in content
        assert "🚀" in content

    # File System Tests
    def test_should_use_utf8_encoding_for_file_operations(self):
        """Test UTF-8 encoding for file operations"""
        # Given
        spec_name = "ENCODING-TEST"
        description = "한글 설명 with UTF-8 encoding"

        # When
        result_path = self.generator.create_spec_file(spec_name, description)

        # Then
        # Read with explicit encoding to verify
        with open(result_path, 'r', encoding='utf-8') as f:
            content = f.read()
        assert description in content

    def test_should_return_correct_file_path(self):
        """Test that correct file path is returned"""
        # Given
        spec_name = "PATH-TEST"
        description = "Path test description"

        # When
        result_path = self.generator.create_spec_file(spec_name, description)

        # Then
        expected_path = self.project_dir / ".moai" / "specs" / f"{spec_name}.md"
        # Resolve both paths to handle symlinks on macOS
        assert result_path.resolve() == expected_path.resolve()
        assert result_path.is_absolute()

    def test_should_handle_long_spec_names_in_filename(self):
        """Test handling of long spec names in filename creation"""
        # Given
        spec_name = "VERY-LONG-SPECIFICATION-NAME-THAT-MIGHT-CAUSE-ISSUES"
        description = "Long name test"

        # When
        result_path = self.generator.create_spec_file(spec_name, description)

        # Then
        assert result_path.exists()
        assert spec_name in result_path.name