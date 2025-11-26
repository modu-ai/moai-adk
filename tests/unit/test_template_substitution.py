"""Unit tests for template variable substitution (processor.py)"""

from moai_adk.core.template.processor import TemplateProcessor


class TestBasicSubstitution:
    """Test basic variable substitution functionality"""

    def test_substitute_single_variable(self, tmp_path):
        """Test substitution of a single variable"""
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "TestProject"})

        content = "# {{PROJECT_NAME}}"
        result, warnings = processor._substitute_variables(content)

        assert "TestProject" in result
        assert "{{PROJECT_NAME}}" not in result
        assert len(warnings) == 0

    def test_substitute_multiple_variables(self, tmp_path):
        """Test substitution of multiple variables"""
        processor = TemplateProcessor(tmp_path)
        processor.set_context(
            {"PROJECT_NAME": "MyProject", "PROJECT_DESCRIPTION": "My awesome project", "AUTHOR": "John Doe"}
        )

        content = """# {{PROJECT_NAME}}
Description: {{PROJECT_DESCRIPTION}}
Author: {{AUTHOR}}"""

        result, warnings = processor._substitute_variables(content)

        assert "MyProject" in result
        assert "My awesome project" in result
        assert "John Doe" in result
        assert "{{" not in result
        assert len(warnings) == 0

    def test_unsubstituted_variable_warning(self, tmp_path):
        """Test warning for unsubstituted variables"""
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "TestProject"})

        content = "# {{PROJECT_NAME}}\nVersion: {{PROJECT_VERSION}}"
        result, warnings = processor._substitute_variables(content)

        assert "TestProject" in result
        assert len(warnings) >= 1
        assert any("PROJECT_VERSION" in warning for warning in warnings)
        assert any("not substituted" in warning for warning in warnings)

    def test_no_context(self, tmp_path):
        """Test substitution with empty context"""
        processor = TemplateProcessor(tmp_path)
        # Don't set context

        content = "# {{PROJECT_NAME}}"
        result, warnings = processor._substitute_variables(content)

        # Without context, variables should remain unchanged
        assert "{{PROJECT_NAME}}" in result


class TestInjectionPrevention:
    """Test security features: injection prevention"""

    def test_recursive_substitution_prevented(self, tmp_path):
        """Test that recursive substitution attacks are prevented"""
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "Test{{ATTACK}}"})

        content = "# {{PROJECT_NAME}}"
        result, warnings = processor._substitute_variables(content)

        # The value should have {{ }} removed by sanitization
        assert "TestATTACK" in result
        assert "{{" not in result
        assert "}}" not in result

    def test_control_character_removal(self, tmp_path):
        """Test that control characters are removed"""
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "Test\x00Project\x01"})

        content = "# {{PROJECT_NAME}}"
        result, warnings = processor._substitute_variables(content)

        # Control characters should be removed
        assert "\x00" not in result
        assert "\x01" not in result
        assert "TestProject" in result

    def test_whitespace_preserved(self, tmp_path):
        """Test that valid whitespace is preserved"""
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "Test\nProject\tWith\rSpaces"})

        content = "# {{PROJECT_NAME}}"
        result, warnings = processor._substitute_variables(content)

        assert "Test\nProject\tWith\rSpaces" in result


class TestFileOperations:
    """Test file copying with substitution"""

    def test_text_file_substitution(self, tmp_path):
        """Test substitution in text files"""
        # Create source file
        src = tmp_path / "template.md"
        src.write_text("# {{PROJECT_NAME}}\n{{PROJECT_DESCRIPTION}}")

        # Create processor and set context
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "MyProject", "PROJECT_DESCRIPTION": "Test project"})

        # Copy with substitution
        dst = tmp_path / "output.md"
        warnings = processor._copy_file_with_substitution(src, dst)

        # Verify destination
        content = dst.read_text()
        assert "MyProject" in content
        assert "Test project" in content
        assert "{{" not in content
        assert len(warnings) == 0

    def test_binary_file_fallback(self, tmp_path):
        """Test that binary files are copied without substitution"""
        # Create binary file
        src = tmp_path / "image.bin"
        src.write_bytes(b"\x89PNG\r\n\x1a\n")

        # Create processor
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "Test"})

        # Copy
        dst = tmp_path / "output.bin"
        processor._copy_file_with_substitution(src, dst)

        # Verify binary content unchanged
        assert dst.read_bytes() == b"\x89PNG\r\n\x1a\n"

    def test_text_file_detection(self, tmp_path):
        """Test that file type is correctly detected"""
        processor = TemplateProcessor(tmp_path)

        # Text files
        assert processor._is_text_file(tmp_path / "file.md") is True
        assert processor._is_text_file(tmp_path / "file.json") is True
        assert processor._is_text_file(tmp_path / "file.yaml") is True
        assert processor._is_text_file(tmp_path / "file.py") is True

        # Binary files
        assert processor._is_text_file(tmp_path / "file.png") is False
        assert processor._is_text_file(tmp_path / "file.bin") is False
        assert processor._is_text_file(tmp_path / "file.exe") is False


class TestContextSetting:
    """Test context setting and management"""

    def test_set_context(self, tmp_path):
        """Test setting context"""
        processor = TemplateProcessor(tmp_path)
        context = {"KEY": "VALUE"}
        processor.set_context(context)

        assert processor.context == context

    def test_context_persistence(self, tmp_path):
        """Test that context persists across operations"""
        processor = TemplateProcessor(tmp_path)
        context = {"PROJECT_NAME": "Test"}
        processor.set_context(context)

        # Use context
        result1, _ = processor._substitute_variables("{{PROJECT_NAME}}")
        # Context should still be there
        result2, _ = processor._substitute_variables("{{PROJECT_NAME}}")

        assert "Test" in result1
        assert "Test" in result2


class TestIntegration:
    """Integration tests for complete workflow"""

    def test_copy_directory_with_substitution(self, tmp_path):
        """Test recursive directory copy with substitution"""
        # Create source directory structure
        src_dir = tmp_path / "src"
        src_dir.mkdir()
        (src_dir / "file1.md").write_text("# {{PROJECT_NAME}}")
        (src_dir / "file2.json").write_text('{"name": "{{PROJECT_NAME}}"}')

        # Create processor
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_NAME": "MyProject"})

        # Copy directory
        dst_dir = tmp_path / "dst"
        processor._copy_dir_with_substitution(src_dir, dst_dir)

        # Verify
        assert (dst_dir / "file1.md").read_text() == "# MyProject"
        assert "MyProject" in (dst_dir / "file2.json").read_text()

    def test_complete_substitution_pipeline(self, tmp_path):
        """Test complete substitution pipeline"""
        # Setup
        processor = TemplateProcessor(tmp_path)
        processor.set_context(
            {"PROJECT_NAME": "AwesomeApp", "PROJECT_DESCRIPTION": "An awesome application", "AUTHOR": "DevTeam"}
        )

        # Create and copy files
        src = tmp_path / "README.md"
        src.write_text(
            """# {{PROJECT_NAME}}

Description: {{PROJECT_DESCRIPTION}}
Author: {{AUTHOR}}
Version: {{VERSION}}"""
        )

        dst = tmp_path / "output.md"
        warnings = processor._copy_file_with_substitution(src, dst)

        # Verify substitution
        content = dst.read_text()
        assert "AwesomeApp" in content
        assert "An awesome application" in content
        assert "DevTeam" in content
        assert "{{VERSION}}" in content  # Unsubstituted
        assert len(warnings) >= 1  # One or more warnings for VERSION


class TestHookProjectDirSubstitution:
    """Test PROJECT_DIR variable substitution for cross-platform support"""

    def test_hook_project_dir_windows(self, tmp_path):
        """Test PROJECT_DIR substitution on Windows"""
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_DIR": "%CLAUDE_PROJECT_DIR%"})

        content = "uv run {{PROJECT_DIR}}/.claude/hooks/session_start.py"
        result, warnings = processor._substitute_variables(content)

        assert "%CLAUDE_PROJECT_DIR%" in result
        assert "{{PROJECT_DIR}}" not in result
        assert len(warnings) == 0

    def test_hook_project_dir_unix(self, tmp_path):
        """Test PROJECT_DIR substitution on Unix-like systems"""
        processor = TemplateProcessor(tmp_path)
        processor.set_context({"PROJECT_DIR": "$CLAUDE_PROJECT_DIR"})

        content = "uv run {{PROJECT_DIR}}/.claude/hooks/session_start.py"
        result, warnings = processor._substitute_variables(content)

        assert "$CLAUDE_PROJECT_DIR" in result
        assert "{{PROJECT_DIR}}" not in result
        assert len(warnings) == 0

    def test_hook_project_dir_missing_context(self, tmp_path):
        """Test PROJECT_DIR without context (should generate warning)"""
        processor = TemplateProcessor(tmp_path)
        # Don't set PROJECT_DIR in context

        content = "uv run {{PROJECT_DIR}}/.claude/hooks/session_start.py"
        result, warnings = processor._substitute_variables(content)

        assert "{{PROJECT_DIR}}" in result
        assert len(warnings) >= 1
        # Check for enhanced error message
        warning_text = " ".join(warnings)
        assert "PROJECT_DIR" in warning_text
        assert "Cross-platform project path" in warning_text

    def test_enhanced_error_messages(self, tmp_path):
        """Test enhanced error messages for common variables"""
        processor = TemplateProcessor(tmp_path)
        # Empty context to trigger warnings

        content = """
# {{PROJECT_NAME}}
Author: {{AUTHOR}}
Language: {{CONVERSATION_LANGUAGE}}
Hook: {{PROJECT_DIR}}
Version: {{MOAI_VERSION}}
Unknown: {{UNKNOWN_VAR}}
"""
        result, warnings = processor._substitute_variables(content)

        # Should have warnings for all unsubstituted variables
        assert len(warnings) >= 2  # At least "Template variables not substituted" and details

        warning_text = " ".join(warnings)
        assert "PROJECT_NAME" in warning_text
        assert "AUTHOR" in warning_text
        assert "CONVERSATION_LANGUAGE" in warning_text
        assert "PROJECT_DIR" in warning_text
        assert "Cross-platform project path" in warning_text
        assert "UNKNOWN_VAR" in warning_text
        assert "Unknown variable" in warning_text
