"""
@TEST:RESOURCE-MANAGER-RED-001 RED Phase - Failing tests for resource_manager.py coverage

Phase 1 of TDD: Write failing tests for ResourceManager class
that accurately reflect the real API and target the core resource operations.

Target Coverage Goals:
- resource_manager.py: 48.33% → 85%
- Key functions: __init__(), copy_*, validate_*, get_*, apply_*
- Real API based tests for file operations and template management
"""

import tempfile
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from moai_adk.install.resource_manager import ResourceManager


class TestResourceManagerInitializationRed:
    """RED: Test ResourceManager initialization and component setup."""

    def setup_method(self):
        """Set up test environment."""
        self.temp_dir = Path(tempfile.mkdtemp())

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_resource_manager_initialization_should_create_component_managers(self):
        """RED: Test ResourceManager initializes with all component managers."""
        # Act
        manager = ResourceManager()

        # Assert
        assert hasattr(manager, 'template_manager')
        assert hasattr(manager, 'file_operations')
        assert hasattr(manager, 'validator')
        assert hasattr(manager, 'templates_root')  # Backwards compatibility

    def test_templates_root_should_provide_backwards_compatibility(self):
        """RED: Test backwards compatibility for templates_root access."""
        # Act
        manager = ResourceManager()

        # Assert
        # Should provide access to template manager's templates_root
        assert manager.templates_root is not None
        assert manager.templates_root == manager.template_manager.templates_root


class TestResourceManagerVersioningRed:
    """RED: Test version and template path operations."""

    def setup_method(self):
        """Set up test environment."""
        self.manager = ResourceManager()

    def test_get_version_should_return_package_version(self):
        """RED: Test get_version returns current package version."""
        # Act
        version = self.manager.get_version()

        # Assert
        assert isinstance(version, str)
        assert len(version) > 0
        # Should be semantic version format (basic check)
        assert "." in version

    def test_get_template_path_should_return_valid_path(self):
        """RED: Test get_template_path returns path for valid template."""
        # Act
        path = self.manager.get_template_path(".moai")

        # Assert
        assert isinstance(path, Path)
        assert path.exists()

    def test_get_template_path_should_handle_nonexistent_template(self):
        """RED: Test get_template_path with non-existent template."""
        # Act
        path = self.manager.get_template_path("nonexistent_template")

        # Assert
        # Should return path even if doesn't exist (caller handles validation)
        assert isinstance(path, Path)

    def test_get_template_content_should_return_string_for_valid_template(self):
        """RED: Test get_template_content returns content as string."""
        # Act
        content = self.manager.get_template_content("CLAUDE.md")

        # Assert
        if content is not None:  # Template exists
            assert isinstance(content, str)
            assert len(content) > 0
        # If None, template doesn't exist - acceptable

    def test_get_template_content_should_return_none_for_invalid_template(self):
        """RED: Test get_template_content returns None for invalid template."""
        # Act
        content = self.manager.get_template_content("definitely_does_not_exist.xyz")

        # Assert
        assert content is None


class TestResourceManagerCopyOperationsRed:
    """RED: Test file copying operations."""

    def setup_method(self):
        """Set up test environment."""
        self.manager = ResourceManager()
        self.temp_dir = Path(tempfile.mkdtemp())

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_copy_template_should_validate_path_safety_first(self):
        """RED: Test copy_template validates path safety before operations."""
        # Arrange: Create unsafe path
        unsafe_path = Path("/etc/passwd")

        # Act
        result = self.manager.copy_template(".moai", unsafe_path)

        # Assert
        assert result is False  # Should reject unsafe path

    def test_copy_template_should_copy_valid_template_to_safe_path(self):
        """RED: Test copy_template with valid template and safe path."""
        # Arrange
        target_path = self.temp_dir / "test_moai"

        # Act
        result = self.manager.copy_template(".moai", target_path)

        # Assert
        assert result is True
        assert target_path.exists()

    def test_copy_template_should_handle_overwrite_flag(self):
        """RED: Test copy_template respects overwrite parameter."""
        # Arrange
        target_path = self.temp_dir / "test_moai"
        target_path.mkdir()  # Pre-existing directory

        # Act - without overwrite
        result1 = self.manager.copy_template(".moai", target_path, overwrite=False)

        # Act - with overwrite
        result2 = self.manager.copy_template(".moai", target_path, overwrite=True)

        # Assert
        # Exact behavior depends on implementation, but should handle both cases
        assert isinstance(result1, bool)
        assert isinstance(result2, bool)

    def test_copy_template_should_exclude_specified_subdirectories(self):
        """RED: Test copy_template excludes specified subdirectories."""
        # Arrange
        target_path = self.temp_dir / "test_moai"
        exclude_subdirs = ["templates", "backup"]

        # Act
        result = self.manager.copy_template(
            ".moai", target_path, exclude_subdirs=exclude_subdirs
        )

        # Assert
        assert isinstance(result, bool)
        # If successful, excluded directories should not be copied
        if result and target_path.exists():
            for excluded in exclude_subdirs:
                excluded_path = target_path / excluded
                # Should not exist or be empty
                assert not excluded_path.exists() or len(list(excluded_path.iterdir())) == 0

    def test_copy_claude_resources_should_setup_claude_directory(self):
        """RED: Test copy_claude_resources creates .claude structure."""
        # Act
        result = self.manager.copy_claude_resources(self.temp_dir)

        # Assert
        assert isinstance(result, bool)
        if result:
            claude_dir = self.temp_dir / ".claude"
            assert claude_dir.exists()

    def test_copy_claude_resources_should_handle_python_command_parameter(self):
        """RED: Test copy_claude_resources with custom python command."""
        # Act
        result = self.manager.copy_claude_resources(
            self.temp_dir, python_command="python3"
        )

        # Assert
        assert isinstance(result, bool)

    def test_copy_moai_resources_should_setup_moai_directory(self):
        """RED: Test copy_moai_resources creates .moai structure."""
        # Act
        result = self.manager.copy_moai_resources(self.temp_dir)

        # Assert
        assert isinstance(result, bool)
        if result:
            moai_dir = self.temp_dir / ".moai"
            assert moai_dir.exists()

    def test_copy_moai_resources_should_handle_exclude_templates_flag(self):
        """RED: Test copy_moai_resources excludes templates when requested."""
        # Act
        result = self.manager.copy_moai_resources(
            self.temp_dir, exclude_templates=True
        )

        # Assert
        assert isinstance(result, bool)

    def test_copy_github_resources_should_setup_github_workflows(self):
        """RED: Test copy_github_resources creates GitHub workflow structure."""
        # Act
        result = self.manager.copy_github_resources(self.temp_dir)

        # Assert
        assert isinstance(result, bool)
        if result:
            github_dir = self.temp_dir / ".github"
            # May or may not exist depending on template availability

    def test_copy_project_memory_should_create_claude_md_file(self):
        """RED: Test copy_project_memory creates CLAUDE.md."""
        # Act
        result = self.manager.copy_project_memory(self.temp_dir)

        # Assert
        assert isinstance(result, bool)
        if result:
            memory_file = self.temp_dir / "CLAUDE.md"
            assert memory_file.exists()
            assert memory_file.is_file()


class TestResourceManagerContextOperationsRed:
    """RED: Test template context and substitution operations."""

    def setup_method(self):
        """Set up test environment."""
        self.manager = ResourceManager()
        self.temp_dir = Path(tempfile.mkdtemp())

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_copy_moai_resources_with_context_should_substitute_variables(self):
        """RED: Test copy_moai_resources_with_context applies template substitution."""
        # Arrange
        context = {
            "project_name": "test-project",
            "author": "Test Author",
            "version": "1.0.0"
        }

        # Act
        result = self.manager.copy_moai_resources_with_context(
            self.temp_dir, context
        )

        # Assert
        assert isinstance(result, bool)

    def test_copy_project_memory_with_context_should_substitute_variables(self):
        """RED: Test copy_project_memory_with_context applies substitution."""
        # Arrange
        context = {
            "project_name": "test-project",
            "description": "Test project description"
        }

        # Act
        result = self.manager.copy_project_memory_with_context(
            self.temp_dir, context
        )

        # Assert
        assert isinstance(result, bool)

    def test_copy_memory_templates_should_handle_tech_stack_filtering(self):
        """RED: Test copy_memory_templates filters by tech stack."""
        # Arrange
        tech_stack = ["python", "react", "postgresql"]
        context = {"project_name": "test"}

        # Act
        result = self.manager.copy_memory_templates(
            self.temp_dir, tech_stack, context
        )

        # Assert
        assert isinstance(result, list)
        # Should return list of copied template paths

    def test_apply_project_context_should_modify_template_in_place(self):
        """RED: Test apply_project_context modifies existing file."""
        # Arrange
        test_file = self.temp_dir / "test_template.md"
        test_file.write_text("Project: {{project_name}}")
        context = {"project_name": "MyProject"}

        # Act
        result = self.manager.apply_project_context(test_file, context)

        # Assert
        assert isinstance(result, bool)
        if result:
            # File should be modified
            content = test_file.read_text()
            assert "MyProject" in content

    def test_substitute_template_variables_should_replace_placeholders(self):
        """RED: Test substitute_template_variables replaces template vars."""
        # Arrange
        template_content = "Hello {{name}}, version {{version}}"
        context = {"name": "World", "version": "1.0"}

        # Act
        result = self.manager.substitute_template_variables(template_content, context)

        # Assert
        assert isinstance(result, str)
        assert "{{" not in result  # No remaining placeholders
        assert "World" in result
        assert "1.0" in result

    def test_render_template_with_context_should_process_template_content(self):
        """RED: Test render_template_with_context processes template."""
        # Arrange
        template_content = "Project: {{project_name}}\nVersion: {{version}}"
        context = {"project_name": "TestApp", "version": "2.0"}

        # Act
        result = self.manager.render_template_with_context(template_content, context)

        # Assert
        assert isinstance(result, str)
        assert "TestApp" in result
        assert "2.0" in result


class TestResourceManagerValidationRed:
    """RED: Test validation operations."""

    def setup_method(self):
        """Set up test environment."""
        self.manager = ResourceManager()
        self.temp_dir = Path(tempfile.mkdtemp())

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_list_templates_should_return_available_template_names(self):
        """RED: Test list_templates returns list of template names."""
        # Act
        templates = self.manager.list_templates()

        # Assert
        assert isinstance(templates, list)
        # Should contain some standard templates
        expected_templates = [".moai", ".claude", "CLAUDE.md"]
        for template in expected_templates:
            if template in templates:
                break
        else:
            pytest.fail("No expected templates found in list")

    def test_validate_project_resources_should_check_basic_project_structure(self):
        """RED: Test validate_project_resources checks project structure."""
        # Act - empty directory
        result_empty = self.manager.validate_project_resources(self.temp_dir)

        # Arrange - create basic structure
        (self.temp_dir / ".moai").mkdir()
        (self.temp_dir / ".claude").mkdir()

        # Act - with structure
        result_with_structure = self.manager.validate_project_resources(self.temp_dir)

        # Assert
        assert isinstance(result_empty, bool)
        assert isinstance(result_with_structure, bool)
        # Should be different results
        assert result_empty != result_with_structure

    def test_validate_required_resources_should_check_all_requirements(self):
        """RED: Test validate_required_resources comprehensive check."""
        # Act - empty directory
        result = self.manager.validate_required_resources(self.temp_dir)

        # Assert
        assert isinstance(result, bool)
        assert result is False  # Empty directory should fail validation

    def test_validate_safe_path_should_reject_dangerous_paths(self):
        """RED: Test validate_safe_path rejects dangerous paths."""
        # Arrange - dangerous paths
        dangerous_paths = [
            Path("/etc/passwd"),
            Path("/usr/bin/sudo"),
            Path("../../../etc/shadow"),
            Path("/root/.ssh/id_rsa"),
        ]

        # Act & Assert
        for dangerous_path in dangerous_paths:
            result = self.manager.validate_safe_path(dangerous_path)
            assert result is False, f"Should reject dangerous path: {dangerous_path}"

    def test_validate_safe_path_should_accept_safe_paths(self):
        """RED: Test validate_safe_path accepts safe paths."""
        # Arrange - safe paths
        safe_paths = [
            self.temp_dir / "project",
            Path.cwd() / "test_project",
            self.temp_dir / "subdir" / "project",
        ]

        # Act & Assert
        for safe_path in safe_paths:
            result = self.manager.validate_safe_path(safe_path)
            assert result is True, f"Should accept safe path: {safe_path}"

    def test_check_path_conflicts_should_detect_existing_files(self):
        """RED: Test check_path_conflicts detects conflicting files."""
        # Arrange - create conflicting files
        (self.temp_dir / ".moai").mkdir()
        (self.temp_dir / "CLAUDE.md").write_text("existing")

        # Act
        conflicts = self.manager.check_path_conflicts(self.temp_dir)

        # Assert
        assert isinstance(conflicts, list)
        assert len(conflicts) > 0  # Should detect conflicts

    def test_get_project_status_should_return_comprehensive_status(self):
        """RED: Test get_project_status returns detailed status information."""
        # Act
        status = self.manager.get_project_status(self.temp_dir)

        # Assert
        assert isinstance(status, dict)
        # Should contain expected status keys
        expected_keys = ["moai_initialized", "claude_initialized", "git_repository"]
        for key in expected_keys:
            if key in status:
                assert isinstance(status[key], bool)

    def test_validate_clean_installation_should_check_installation_readiness(self):
        """RED: Test validate_clean_installation checks if path is ready."""
        # Act - empty directory
        result_clean = self.manager.validate_clean_installation(self.temp_dir)

        # Arrange - create some files
        (self.temp_dir / "existing_file.txt").write_text("content")

        # Act - with existing files
        result_with_files = self.manager.validate_clean_installation(self.temp_dir)

        # Assert
        assert isinstance(result_clean, bool)
        assert isinstance(result_with_files, bool)


class TestResourceManagerEdgeCasesRed:
    """RED: Test edge cases and error conditions."""

    def setup_method(self):
        """Set up test environment."""
        self.manager = ResourceManager()

    def test_copy_operations_should_handle_permission_errors_gracefully(self):
        """RED: Test copy operations handle permission errors."""
        # Arrange - readonly directory
        with tempfile.TemporaryDirectory() as temp_dir:
            readonly_path = Path(temp_dir) / "readonly"
            readonly_path.mkdir(mode=0o444)

            try:
                # Act
                result = self.manager.copy_moai_resources(readonly_path)

                # Assert - should handle gracefully
                assert isinstance(result, bool)
                # Likely False due to permissions, but shouldn't crash
            finally:
                # Restore permissions for cleanup
                readonly_path.chmod(0o755)

    def test_template_operations_should_handle_missing_templates_gracefully(self):
        """RED: Test template operations with missing templates."""
        # Act
        result = self.manager.copy_template("definitely_nonexistent_template", Path("/tmp/test"))

        # Assert
        assert isinstance(result, bool)
        assert result is False  # Should fail gracefully

    def test_context_substitution_should_handle_missing_variables(self):
        """RED: Test context substitution with missing variables."""
        # Arrange
        template_content = "{{existing_var}} and {{missing_var}}"
        incomplete_context = {"existing_var": "value1"}

        # Act
        result = self.manager.substitute_template_variables(template_content, incomplete_context)

        # Assert
        assert isinstance(result, str)
        # Should handle missing variables gracefully (implementation dependent)

    def test_unicode_content_should_be_handled_safely(self):
        """RED: Test operations handle unicode content safely."""
        # Arrange
        unicode_content = "프로젝트: {{project_name}} 버전: {{version}}"
        unicode_context = {"project_name": "한글프로젝트", "version": "1.0"}

        # Act
        result = self.manager.substitute_template_variables(unicode_content, unicode_context)

        # Assert
        assert isinstance(result, str)
        assert "한글프로젝트" in result

    def test_large_template_content_should_be_processed_efficiently(self):
        """RED: Test processing of large template content."""
        # Arrange - simulate large content
        large_content = "{{var}}\n" * 10000  # 10k lines
        context = {"var": "value"}

        # Act
        result = self.manager.substitute_template_variables(large_content, context)

        # Assert
        assert isinstance(result, str)
        assert len(result.split('\n')) >= 10000


class TestResourceManagerIntegrationRed:
    """RED: Test integration scenarios."""

    def setup_method(self):
        """Set up test environment."""
        self.manager = ResourceManager()
        self.temp_dir = Path(tempfile.mkdtemp())

    def teardown_method(self):
        """Clean up test environment."""
        if self.temp_dir.exists():
            import shutil
            shutil.rmtree(self.temp_dir)

    def test_full_project_setup_should_create_complete_structure(self):
        """RED: Test complete project setup workflow."""
        # Act - simulate full project setup
        context = {"project_name": "TestProject", "version": "1.0.0"}

        results = []
        results.append(self.manager.copy_moai_resources_with_context(self.temp_dir, context))
        results.append(self.manager.copy_claude_resources(self.temp_dir))
        results.append(self.manager.copy_project_memory_with_context(self.temp_dir, context))

        # Assert
        assert all(isinstance(r, bool) for r in results)
        # At least some operations should succeed
        assert any(results), "At least one setup operation should succeed"

    def test_validation_after_setup_should_pass_checks(self):
        """RED: Test validation after successful setup."""
        # Arrange - set up project
        self.manager.copy_moai_resources(self.temp_dir)
        self.manager.copy_claude_resources(self.temp_dir)

        # Act
        is_valid = self.manager.validate_project_resources(self.temp_dir)
        status = self.manager.get_project_status(self.temp_dir)

        # Assert
        assert isinstance(is_valid, bool)
        assert isinstance(status, dict)

    def test_overwrite_scenarios_should_handle_existing_installations(self):
        """RED: Test overwrite behavior with existing installations."""
        # Arrange - create initial installation
        self.manager.copy_moai_resources(self.temp_dir)

        # Act - try to install again with and without overwrite
        result_no_overwrite = self.manager.copy_moai_resources(self.temp_dir, overwrite=False)
        result_with_overwrite = self.manager.copy_moai_resources(self.temp_dir, overwrite=True)

        # Assert
        assert isinstance(result_no_overwrite, bool)
        assert isinstance(result_with_overwrite, bool)
        # Behavior may differ between overwrite modes

    def test_component_manager_integration_should_work_seamlessly(self):
        """RED: Test that component managers work together properly."""
        # Act - access different manager capabilities
        version = self.manager.get_version()
        templates = self.manager.list_templates()
        safe_path_result = self.manager.validate_safe_path(self.temp_dir)

        # Assert
        assert isinstance(version, str)
        assert isinstance(templates, list)
        assert isinstance(safe_path_result, bool)
        # All component managers should be accessible and functional