"""
Unit tests for the file manager module.

Tests the FileManager class and its file operation methods
to ensure proper template handling and security validation.
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock, mock_open, call

from moai_adk.core.file_manager import FileManager
from moai_adk.core.security import SecurityManager


class TestFileManager:
    """Test cases for FileManager class."""

    @pytest.fixture
    def temp_dir(self):
        """Create a temporary directory for testing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    @pytest.fixture
    def security_manager(self):
        """Create a mock security manager for testing."""
        return MagicMock(spec=SecurityManager)

    @pytest.fixture
    def file_manager(self, temp_dir, security_manager):
        """Create a FileManager instance for testing."""
        template_dir = temp_dir / "templates"
        template_dir.mkdir()
        return FileManager(template_dir, security_manager)

    def test_init_with_security_manager(self, temp_dir, security_manager):
        """Test FileManager initialization with security manager."""
        manager = FileManager(temp_dir, security_manager)

        assert manager.template_dir == temp_dir
        assert manager.security_manager == security_manager

    def test_init_without_security_manager(self, temp_dir):
        """Test FileManager initialization without security manager."""
        manager = FileManager(temp_dir)

        assert manager.template_dir == temp_dir
        assert isinstance(manager.security_manager, SecurityManager)

    def test_copy_template_files_source_not_exists(self, file_manager, temp_dir):
        """Test copy_template_files when source directory doesn't exist."""
        source_dir = temp_dir / "nonexistent"
        target_dir = temp_dir / "target"

        result = file_manager.copy_template_files(source_dir, target_dir, "*.py")

        assert result == []

    def test_copy_template_files_success(self, file_manager, temp_dir):
        """Test successful copying of template files."""
        # Setup source files
        source_dir = temp_dir / "source"
        source_dir.mkdir()
        (source_dir / "file1.py").write_text("print('hello')")
        (source_dir / "file2.py").write_text("print('world')")
        (source_dir / "readme.txt").write_text("readme")

        target_dir = temp_dir / "target"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.copy_template_files(source_dir, target_dir, "*.py")

        assert len(result) == 2
        assert all(f.suffix == ".py" for f in result)
        assert (target_dir / "file1.py").exists()
        assert (target_dir / "file2.py").exists()
        assert not (target_dir / "readme.txt").exists()

    def test_copy_template_files_security_validation_fails(
        self, file_manager, temp_dir
    ):
        """Test copy_template_files when security validation fails."""
        # Setup source files
        source_dir = temp_dir / "source"
        source_dir.mkdir()
        (source_dir / "file1.py").write_text("print('hello')")

        target_dir = temp_dir / "target"

        # Mock security validation to return False
        file_manager.security_manager.validate_file_creation.return_value = False

        result = file_manager.copy_template_files(source_dir, target_dir, "*.py")

        assert result == []
        assert not (target_dir / "file1.py").exists()

    def test_copy_template_files_preserve_permissions(self, file_manager, temp_dir):
        """Test copy_template_files with permission preservation."""
        # Setup source files
        source_dir = temp_dir / "source"
        source_dir.mkdir()
        source_file = source_dir / "script.py"
        source_file.write_text("print('hello')")
        source_file.chmod(0o755)

        target_dir = temp_dir / "target"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.copy_template_files(
            source_dir, target_dir, "*.py", preserve_permissions=True
        )

        assert len(result) == 1
        target_file = target_dir / "script.py"
        assert target_file.exists()
        # Check that permissions were preserved
        assert oct(target_file.stat().st_mode)[-3:] == "755"

    def test_copy_template_files_skips_directories(self, file_manager, temp_dir):
        """Test that copy_template_files skips directories."""
        # Setup source with directory and file
        source_dir = temp_dir / "source"
        source_dir.mkdir()
        (source_dir / "subdir").mkdir()  # This should be skipped
        (source_dir / "file.py").write_text("content")

        target_dir = temp_dir / "target"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.copy_template_files(source_dir, target_dir, "*")

        assert len(result) == 1
        assert result[0].name == "file.py"

    def test_render_template_file_success(self, file_manager, temp_dir):
        """Test successful template rendering."""
        template_path = temp_dir / "template.txt"
        template_path.write_text("Hello $name, version $version")

        context = {"name": "World", "version": "1.0"}

        result = file_manager.render_template_file(template_path, context)

        assert result == "Hello World, version 1.0"

    def test_render_template_file_missing_variable(self, file_manager, temp_dir):
        """Test template rendering with missing variable."""
        template_path = temp_dir / "template.txt"
        template_path.write_text("Hello $name, version $version")

        context = {"name": "World"}  # Missing 'version'

        result = file_manager.render_template_file(template_path, context)

        # safe_substitute should leave missing variables as-is
        assert result == "Hello World, version $version"

    def test_render_template_file_not_found(self, file_manager, temp_dir):
        """Test template rendering when file doesn't exist."""
        template_path = temp_dir / "nonexistent.txt"
        context = {"name": "World"}

        result = file_manager.render_template_file(template_path, context)

        assert result == ""

    @patch("builtins.open", side_effect=IOError("Permission denied"))
    def test_render_template_file_io_error(self, mock_open, file_manager, temp_dir):
        """Test template rendering with IO error."""
        template_path = temp_dir / "template.txt"
        context = {"name": "World"}

        result = file_manager.render_template_file(template_path, context)

        assert result == ""

    def test_copy_and_render_template_success(self, file_manager, temp_dir):
        """Test successful template copying and rendering."""
        source_path = temp_dir / "source.txt"
        source_path.write_text("Project: $project_name")

        target_path = temp_dir / "output" / "result.txt"
        context = {"project_name": "MoAI-ADK"}

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.copy_and_render_template(
            source_path, target_path, context
        )

        assert result is True
        assert target_path.exists()
        assert target_path.read_text() == "Project: MoAI-ADK"

    def test_copy_and_render_template_source_not_exists(self, file_manager, temp_dir):
        """Test copy_and_render_template when source doesn't exist."""
        source_path = temp_dir / "nonexistent.txt"
        target_path = temp_dir / "output.txt"
        context = {"name": "test"}

        result = file_manager.copy_and_render_template(
            source_path, target_path, context
        )

        assert result is False
        assert not target_path.exists()

    def test_copy_and_render_template_security_fail(self, file_manager, temp_dir):
        """Test copy_and_render_template when security validation fails."""
        source_path = temp_dir / "source.txt"
        source_path.write_text("Content")

        target_path = temp_dir / "output.txt"
        context = {}

        # Mock security validation to return False
        file_manager.security_manager.validate_file_creation.return_value = False

        result = file_manager.copy_and_render_template(
            source_path, target_path, context
        )

        assert result is False
        assert not target_path.exists()

    def test_copy_and_render_template_no_create_dirs(self, file_manager, temp_dir):
        """Test copy_and_render_template without creating directories."""
        source_path = temp_dir / "source.txt"
        source_path.write_text("Content")

        # Target in non-existent directory
        target_path = temp_dir / "nonexistent" / "output.txt"
        context = {}

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.copy_and_render_template(
            source_path, target_path, context, create_dirs=False
        )

        # Should fail because parent directory doesn't exist
        assert result is False

    def test_copy_hook_scripts_success(self, file_manager, temp_dir):
        """Test successful copying of hook scripts."""
        # Setup source hook files
        hook_dir = file_manager.template_dir / ".claude" / "hooks" / "moai"
        hook_dir.mkdir(parents=True)

        hook_files = [
            "pre_write_guard.py",
            "policy_block.py",
            "tag_validator.py",
            "steering_guard.py",
            "run_tests_and_report.py",
            "session_start_notice.py",
            "language_detector.py",
        ]

        for hook_file in hook_files:
            (hook_dir / hook_file).write_text(f"# {hook_file} content")

        target_dir = temp_dir / "hooks"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.copy_hook_scripts(target_dir)

        assert len(result) == 7
        for hook_file in hook_files:
            target_file = target_dir / hook_file
            assert target_file.exists()
            # Check that Python files have executable permissions
            assert oct(target_file.stat().st_mode)[-3:] == "755"

    def test_copy_hook_scripts_security_validation_fails(self, file_manager, temp_dir):
        """Test copy_hook_scripts when security validation fails."""
        # Setup source hook files
        hook_dir = file_manager.template_dir / ".claude" / "hooks" / "moai"
        hook_dir.mkdir(parents=True)
        (hook_dir / "policy_block.py").write_text("# content")

        target_dir = temp_dir / "hooks"

        # Mock security validation to return False
        file_manager.security_manager.validate_file_creation.return_value = False

        result = file_manager.copy_hook_scripts(target_dir)

        assert result == []

    def test_copy_hook_scripts_missing_source(self, file_manager, temp_dir):
        """Test copy_hook_scripts when source files are missing."""
        target_dir = temp_dir / "hooks"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.copy_hook_scripts(target_dir)

        assert result == []

    def test_copy_verification_scripts_success(self, file_manager, temp_dir):
        """Test successful copying of verification scripts."""
        # Setup source script files
        script_dir = file_manager.template_dir / "scripts"
        script_dir.mkdir(parents=True)

        script_files = [
            "validate_stage.py",
            "check-secrets.py",
            "check-licenses.py",
            # Note: run-tests.sh removed - use Python test_runner.py instead
        ]

        for script_file in script_files:
            (script_dir / script_file).write_text(f"# {script_file} content")

        target_dir = temp_dir / "scripts"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.copy_verification_scripts(target_dir)

        assert len(result) == 4
        for script_file in script_files:
            target_file = target_dir / script_file
            assert target_file.exists()
            # Check that all files have executable permissions
            assert oct(target_file.stat().st_mode)[-3:] == "755"

    def test_copy_verification_scripts_partial_success(self, file_manager, temp_dir):
        """Test copy_verification_scripts with some missing files."""
        # Setup only some source script files
        script_dir = file_manager.template_dir / "scripts"
        script_dir.mkdir(parents=True)
        (script_dir / "validate_stage.py").write_text("# content")
        # Other files missing

        target_dir = temp_dir / "scripts"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.copy_verification_scripts(target_dir)

        assert len(result) == 1
        assert (target_dir / "validate_stage.py").exists()

    def test_install_output_styles_success(self, file_manager, temp_dir):
        """Test successful installation of output styles."""
        # Setup source style files
        styles_dir = file_manager.template_dir / ".claude" / "output-styles"
        styles_dir.mkdir(parents=True)

        style_files = ["expert.md", "beginner.md", "study.md"]
        for style_file in style_files:
            (styles_dir / style_file).write_text(f"# $project_name {style_file}")

        target_dir = temp_dir / "styles"
        context = {"project_name": "TestProject"}

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.install_output_styles(target_dir, context)

        assert len(result) == 3
        for style_file in style_files:
            target_file = target_dir / style_file
            assert target_file.exists()
            assert target_file.read_text() == f"# TestProject {style_file}"

    def test_install_output_styles_missing_template_dir(self, file_manager, temp_dir):
        """Test install_output_styles when template directory is missing."""
        target_dir = temp_dir / "styles"
        context = {"project_name": "TestProject"}

        result = file_manager.install_output_styles(target_dir, context)

        assert result == []

    def test_install_output_styles_security_validation_fails(
        self, file_manager, temp_dir
    ):
        """Test install_output_styles when security validation fails."""
        # Setup source style files
        styles_dir = file_manager.template_dir / ".claude" / "output-styles"
        styles_dir.mkdir(parents=True)
        (styles_dir / "expert.md").write_text("# Content")

        target_dir = temp_dir / "styles"
        context = {}

        # Mock security validation to return False
        file_manager.security_manager.validate_file_creation.return_value = False

        result = file_manager.install_output_styles(target_dir, context)

        assert result == []

    def test_create_gitignore_success(self, file_manager, temp_dir):
        """Test successful creation of .gitignore file."""
        gitignore_path = temp_dir / ".gitignore"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.create_gitignore(gitignore_path)

        assert result is True
        assert gitignore_path.exists()

        content = gitignore_path.read_text()
        assert "# MoAI-ADK specific" in content
        assert "__pycache__/" in content
        assert ".DS_Store" in content

    def test_create_gitignore_security_validation_fails(self, file_manager, temp_dir):
        """Test create_gitignore when security validation fails."""
        gitignore_path = temp_dir / ".gitignore"

        # Mock security validation to return False
        file_manager.security_manager.validate_file_creation.return_value = False

        result = file_manager.create_gitignore(gitignore_path)

        assert result is False
        assert not gitignore_path.exists()

    @patch("builtins.open", side_effect=IOError("Permission denied"))
    def test_create_gitignore_io_error(self, mock_open, file_manager, temp_dir):
        """Test create_gitignore with IO error."""
        gitignore_path = temp_dir / ".gitignore"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        result = file_manager.create_gitignore(gitignore_path)

        assert result is False

    def test_integration_workflow(self, file_manager, temp_dir):
        """Test complete workflow with multiple operations."""
        # Setup template structure
        claude_dir = file_manager.template_dir / ".claude"
        (claude_dir / "hooks" / "moai").mkdir(parents=True)
        (claude_dir / "output-styles").mkdir(parents=True)
        (file_manager.template_dir / "scripts").mkdir(parents=True)

        # Create template files
        (claude_dir / "hooks" / "moai" / "policy_block.py").write_text("# Policy hook")
        (claude_dir / "output-styles" / "expert.md").write_text(
            "# Expert style for $project"
        )
        # Note: run-tests.sh removed - Python test_runner.py used instead

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        project_dir = temp_dir / "project"
        context = {"project": "MyProject"}

        # Execute workflow
        hook_files = file_manager.copy_hook_scripts(
            project_dir / ".claude" / "hooks" / "moai"
        )
        style_files = file_manager.install_output_styles(
            project_dir / ".claude" / "output-styles", context
        )
        script_files = file_manager.copy_verification_scripts(
            project_dir / ".moai" / "scripts"
        )
        gitignore_created = file_manager.create_gitignore(project_dir / ".gitignore")

        # Verify results
        assert len(hook_files) == 1
        assert len(style_files) == 1
        assert len(script_files) == 1
        assert gitignore_created is True

        # Verify file contents
        assert (
            project_dir / ".claude" / "output-styles" / "expert.md"
        ).read_text() == "# Expert style for MyProject"
        assert (project_dir / ".gitignore").exists()

    def test_error_handling_in_copy_template_files(self, file_manager, temp_dir):
        """Test error handling in copy_template_files."""
        # Setup source files
        source_dir = temp_dir / "source"
        source_dir.mkdir()
        (source_dir / "file.py").write_text("content")

        target_dir = temp_dir / "target"

        # Mock security validation to return True
        file_manager.security_manager.validate_file_creation.return_value = True

        # Mock shutil.copy2 to raise an exception
        with patch("shutil.copy2", side_effect=IOError("Disk full")):
            result = file_manager.copy_template_files(source_dir, target_dir, "*.py")

            # Should handle the error gracefully and return empty list
            assert result == []

    def test_template_rendering_with_complex_content(self, file_manager, temp_dir):
        """Test template rendering with complex content including special characters."""
        template_path = temp_dir / "complex_template.txt"
        template_content = """
Project: $project_name
Version: $version
Special chars: @#$%^&*()
Nested: ${project_name}_${version}
Unicode: 한글 テスト 🚀
"""
        template_path.write_text(template_content)

        context = {"project_name": "MoAI-ADK", "version": "1.0.0"}

        result = file_manager.render_template_file(template_path, context)

        assert "Project: MoAI-ADK" in result
        assert "Version: 1.0.0" in result
        assert "Special chars: @#$%^&*()" in result
        assert "Nested: MoAI-ADK_1.0.0" in result
        assert "Unicode: 한글 テスト 🚀" in result
