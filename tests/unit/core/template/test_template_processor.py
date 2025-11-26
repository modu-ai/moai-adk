"""
Template Processor Tests

Test cases for template synchronization and processing functionality.
"""

import tempfile
from pathlib import Path


class TestTemplateProcessor:
    """Test suite for template processor functionality."""

    def test_template_processor_creation(self):
        """Test that template processor can be created successfully."""
        # This test should fail initially as the TemplateProcessor class doesn't exist
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)
            assert processor is not None
            assert processor.target_path == target_path.resolve()
            assert hasattr(processor, "backup")
            assert hasattr(processor, "merger")
            assert hasattr(processor, "context")

    def test_template_processor_with_context(self):
        """Test template processor with variable context."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            context = {"PROJECT_NAME": "test-project", "AUTHOR": "Test Author", "CONVERSATION_LANGUAGE": "ko"}
            processor.set_context(context)
            assert processor.context == context

    def test_template_processor_get_template_root(self):
        """Test that template root path is correctly identified."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)
            template_root = processor._get_template_root()

            assert template_root.name == "templates"
            assert "moai_adk" in str(template_root)

    def test_substitute_variables_simple(self):
        """Test simple variable substitution."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            context = {"PROJECT_NAME": "test-project"}
            processor.set_context(context)

            content = "Hello {{PROJECT_NAME}}!"
            result, warnings = processor._substitute_variables(content)

            assert result == "Hello test-project!"
            assert len(warnings) == 0

    def test_substitute_variables_multiple(self):
        """Test multiple variable substitution."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            context = {"PROJECT_NAME": "test-project", "AUTHOR": "Test Author", "VERSION": "1.0.0"}
            processor.set_context(context)

            content = "{{PROJECT_NAME}} by {{AUTHOR}}, version {{VERSION}}"
            result, warnings = processor._substitute_variables(content)

            assert result == "test-project by Test Author, version 1.0.0"
            assert len(warnings) == 0

    def test_substitute_variables_missing(self):
        """Test substitution with missing variables (should warn)."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            context = {"PROJECT_NAME": "test-project"}
            processor.set_context(context)

            content = "Hello {{PROJECT_NAME}} and {{UNKNOWN_VAR}}!"
            result, warnings = processor._substitute_variables(content)

            assert result == "Hello test-project and {{UNKNOWN_VAR}}!"
            assert len(warnings) > 0
            assert "Template variables not substituted" in warnings[0]

    def test_substitute_variables_deprecated_mapping(self):
        """Test deprecated variable mapping."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            context = {"PROJECT_DIR": "/path/to/project"}
            processor.set_context(context)

            # HOOK_PROJECT_DIR should be automatically mapped
            content = "{{HOOK_PROJECT_DIR}}"
            result, warnings = processor._substitute_variables(content)

            assert result == "/path/to/project"
            assert len(warnings) == 0

    def test_sanitize_value(self):
        """Test value sanitization."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            # Test control characters removal
            value = "test\n\r\tvalue{{placeholder}}"
            sanitized = processor._sanitize_value(value)

            assert "{{placeholder}}" not in sanitized  # Recursive substitution prevented
            assert "\n" in sanitized and "\r" in sanitized and "\t" in sanitized  # Whitespace preserved

    def test_is_text_file(self):
        """Test text file detection."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            # Test text files
            assert processor._is_text_file(Path("test.py")) is True
            assert processor._is_text_file(Path("test.md")) is True
            assert processor._is_text_file(Path("test.json")) is True
            assert processor._is_text_file(Path("test.txt")) is True
            assert processor._is_text_file(Path("test.yaml")) is True

            # Test non-text files
            assert processor._is_text_file(Path("test.jpg")) is False
            assert processor._is_text_file(Path("test.bin")) is False

    def test_localize_yaml_description(self):
        """Test YAML description localization."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            # Test content with multilingual description
            content = """---
title: Test
description:
  en: English description
  ko: Korean description
---
Content here
"""

            localized = processor._localize_yaml_description(content, "ko")

            assert "Korean description" in localized
            assert "English description" not in localized

    def test_localize_yaml_description_fallback(self):
        """Test YAML description localization with English fallback."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            # Test content with multilingual description, request unknown language
            content = """---
title: Test
description:
  en: English description
  ko: Korean description
---
Content here
"""

            localized = processor._localize_yaml_description(content, "ja")

            assert "English description" in localized  # Fallback to English

    def test_localize_yaml_description_no_frontmatter(self):
        """Test YAML localization without frontmatter."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            content = "No frontmatter here"
            result = processor._localize_yaml_description(content)

            assert result == content  # Should return unchanged

    def test_copy_file_with_substitution_text(self):
        """Test file copying with substitution for text files."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)
            processor.set_context({"PROJECT_NAME": "test-project"})

            # Create source file
            src_file = target_path / "src.txt"
            src_file.write_text("Hello {{PROJECT_NAME}}!")

            # Create destination file
            dst_file = target_path / "dst.txt"

            warnings = processor._copy_file_with_substitution(src_file, dst_file)

            assert dst_file.exists()
            assert dst_file.read_text() == "Hello test-project!"
            assert len(warnings) == 0

    def test_copy_file_with_substitution_binary(self):
        """Test file copying for binary files (no substitution)."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)
            processor.set_context({"PROJECT_NAME": "test-project"})

            # Create binary-like file
            src_file = target_path / "src.bin"
            src_file.write_bytes(b"binary content {{PROJECT_NAME}}")

            # Create destination file
            dst_file = target_path / "dst.bin"

            warnings = processor._copy_file_with_substitution(src_file, dst_file)

            assert dst_file.exists()
            # Should preserve original binary content (no substitution)
            assert dst_file.read_bytes() == b"binary content {{PROJECT_NAME}}"
            assert len(warnings) == 0

    def test_copy_file_with_substitution_no_context(self):
        """Test file copying without context."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)  # No context

            # Create source file with variables
            src_file = target_path / "src.txt"
            src_file.write_text("Hello {{PROJECT_NAME}}!")

            # Create destination file
            dst_file = target_path / "dst.txt"

            warnings = processor._copy_file_with_substitution(src_file, dst_file)

            assert dst_file.exists()
            assert dst_file.read_text() == "Hello {{PROJECT_NAME}}!"  # No substitution
            assert len(warnings) == 0

    def test_copy_dir_with_substitution(self):
        """Test directory copying with substitution."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)
            processor.set_context({"PROJECT_NAME": "test-project"})

            # Create source directory structure
            src_dir = target_path / "src"
            src_dir.mkdir()

            (src_dir / "file1.txt").write_text("Hello {{PROJECT_NAME}}!")
            (src_dir / "file2.txt").write_text("Version 1.0")

            sub_dir = src_dir / "subdir"
            sub_dir.mkdir()
            (sub_dir / "file3.txt").write_text("Subdirectory content")

            # Create destination directory
            dst_dir = target_path / "dst"

            processor._copy_dir_with_substitution(src_dir, dst_dir)

            assert dst_dir.exists()
            assert (dst_dir / "file1.txt").exists()
            assert (dst_dir / "file2.txt").exists()
            assert (dst_dir / "subdir" / "file3.txt").exists()

            # Check substitution happened
            assert (dst_dir / "file1.txt").read_text() == "Hello test-project!"
            # Check non-variable content preserved
            assert (dst_dir / "file2.txt").read_text() == "Version 1.0"

    def test_has_existing_files_false(self):
        """Test _has_existing_files when no files exist."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            result = processor._has_existing_files()
            assert result is False

    def test_has_existing_files_true(self):
        """Test _has_existing_files when files exist."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            # Create some files
            (target_path / ".moai").mkdir()
            (target_path / ".moai" / "config.json").write_text("{}")

            result = processor._has_existing_files()
            assert result is True

    def test_create_backup(self):
        """Test backup creation."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            # Create some files to backup
            (target_path / "existing.txt").write_text("content")

            backup_path = processor.create_backup()

            assert backup_path.exists()
            assert "backup" in backup_path.name.lower()

    def test_is_text_file_unsupported_extension(self):
        """Test text file detection with unsupported extension."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            # Test unsupported extensions
            assert processor._is_text_file(Path("test.unknown")) is False
            assert processor._is_text_file(Path("test.cpp")) is False  # Not in default list
            assert processor._is_text_file(Path("test.java")) is False  # Not in default list

    def test_substitute_variables_empty_context(self):
        """Test variable substitution with empty context."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)
            processor.set_context({})  # Empty context

            content = "Hello {{PROJECT_NAME}}!"
            result, warnings = processor._substitute_variables(content)

            assert result == "Hello {{PROJECT_NAME}}!"  # No changes
            assert len(warnings) > 0  # Should warn about unsubstituted variables

    def test_substitute_variables_sanitization(self):
        """Test variable substitution with value sanitization."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            # Test with value that contains problematic characters
            context = {"PROJECT_NAME": "test\nproject\r{{BAD}}", "SAFE_VALUE": "normal_value"}
            processor.set_context(context)

            content = "{{PROJECT_NAME}} and {{SAFE_VALUE}}"
            result, warnings = processor._substitute_variables(content)

            # Control characters should be preserved, but {{ }} should be removed
            assert "{{BAD}}" not in result
            assert "\n" in result and "\r" in result  # Control chars preserved
            assert "normal_value" in result

    def test_localize_yaml_description_invalid_yaml(self):
        """Test YAML localization with invalid YAML content."""
        # This test should fail initially
        from moai_adk.core.template.processor import TemplateProcessor

        with tempfile.TemporaryDirectory() as temp_dir:
            target_path = Path(temp_dir)
            processor = TemplateProcessor(target_path)

            # Test with invalid YAML
            content = """---
title: Test
description:
  en: English description
  ko: Korean description
invalid yaml here
---
Content here
"""

            result = processor._localize_yaml_description(content, "ko")

            # Should return original content if YAML parsing fails
            assert result == content
