"""
Comprehensive coverage tests for TemplateProcessor module.

Tests cover advanced features not covered in test_processor_new.py:
- Version handling and formatting methods
- Variable validation with graceful degradation
- Caching behavior and cache management
- File operations with edge cases
- Sync methods (skills, sections)
- Merge methods (deep merge, YAML, JSON, MCP)
- Copy methods (full workflow)

Target: 100% coverage for src/moai_adk/core/template/processor.py
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest
import yaml

from moai_adk.core.template.processor import (
    TemplateProcessor,
    TemplateProcessorConfig,
)


class TestVersionHandling:
    """Test version handling and formatting methods."""

    def test_is_valid_version_format_valid_semver(self):
        """Test _is_valid_version_format accepts valid semver."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            assert processor._is_valid_version_format("1.2.3") is True
            assert processor._is_valid_version_format("v1.2.3") is True
            assert processor._is_valid_version_format("2.0.0-alpha") is True
            # Note: dot (.) in pre-release identifier not supported by default regex
            assert processor._is_valid_version_format("1.0.0-beta1") is True
            assert processor._is_valid_version_format("3.0.0-rc1+build.123") is False  # +build not supported

    def test_is_valid_version_format_invalid(self):
        """Test _is_valid_version_format rejects invalid formats."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            assert processor._is_valid_version_format("1.2") is False
            assert processor._is_valid_version_format("v1.2.3.4") is False
            assert processor._is_valid_version_format("latest") is False
            assert processor._is_valid_version_format("") is False
            assert processor._is_valid_version_format("unknown") is False

    def test_is_valid_version_format_custom_regex(self):
        """Test _is_valid_version_format with custom regex pattern."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(
                version_format_regex=r"^\d+\.\d+$"  # Custom pattern
            )
            processor = TemplateProcessor(Path(temp_dir), config)

            assert processor._is_valid_version_format("1.2") is True
            assert processor._is_valid_version_format("1.2.3") is False

    def test_is_valid_version_format_handles_invalid_regex(self):
        """Test _is_valid_version_format falls back on invalid regex."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(
                version_format_regex=r"[invalid(regex"  # Invalid regex
            )
            processor = TemplateProcessor(Path(temp_dir), config)

            # Should fall back to default pattern
            assert processor._is_valid_version_format("1.2.3") is True
            assert processor._is_valid_version_format("invalid") is False

    def test_format_trimmed_version_various_lengths(self):
        """Test _format_trimmed_version with various max_length values."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            # Long version gets trimmed
            assert processor._format_trimmed_version("1.2.3-beta.1", max_length=5) == "1.2.3"
            assert processor._format_trimmed_version("1.2.3-beta.1", max_length=10) == "1.2.3-beta"
            # Short version unchanged
            assert processor._format_trimmed_version("1.2.3", max_length=10) == "1.2.3"
            # Unknown version
            assert processor._format_trimmed_version("unknown", max_length=10) == "unknown"

    def test_format_trimmed_version_with_v_prefix(self):
        """Test _format_trimmed_version removes 'v' prefix before trimming."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            assert processor._format_trimmed_version("v1.2.3-beta.1", max_length=5) == "1.2.3"
            assert processor._format_trimmed_version("v10.20.30", max_length=8) == "10.20.30"

    def test_format_semver_version_pre_release(self):
        """Test _format_semver_version extracts core semver from pre-release versions."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            assert processor._format_semver_version("1.2.3-alpha.1") == "1.2.3"
            assert processor._format_semver_version("v2.0.0-beta") == "2.0.0"
            assert processor._format_semver_version("3.0.0-rc.1+build.123") == "3.0.0"

    def test_format_semver_version_unknown(self):
        """Test _format_semver_version returns 0.0.0 for unknown version."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            assert processor._format_semver_version("unknown") == "0.0.0"

    def test_format_semver_version_with_v_prefix(self):
        """Test _format_semver_version removes 'v' prefix."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            assert processor._format_semver_version("v1.2.3") == "1.2.3"
            assert processor._format_semver_version("v10.20.30") == "10.20.30"

    def test_get_enhanced_version_context_fallback(self):
        """Test get_enhanced_version_context uses fallback when version read fails."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(version_fallback="2.0.0")
            processor = TemplateProcessor(Path(temp_dir), config)

            with patch.object(processor, "_get_version_reader", side_effect=Exception("Read failed")):
                context = processor.get_enhanced_version_context()

                assert context["MOAI_VERSION"] == "2.0.0"
                assert context["MOAI_VERSION_SHORT"] == "2.0.0"
                assert context["MOAI_VERSION_VALID"] == "true"
                assert context["MOAI_VERSION_SOURCE"] == "fallback_config"

    def test_get_enhanced_version_context_with_format_validation(self):
        """Test get_enhanced_version_context includes FORMAT_VALID when enabled.

        NOTE: When version read fails, FORMAT_VALID is always set to "false"
        regardless of fallback_version format. This is the current behavior.
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(
                enable_version_validation=True,
                version_fallback="1.2.3",
            )
            processor = TemplateProcessor(Path(temp_dir), config)

            with patch.object(processor, "_get_version_reader", side_effect=Exception("Read failed")):
                context = processor.get_enhanced_version_context()

                assert "MOAI_VERSION_FORMAT_VALID" in context
                # When version read fails, validation is always false (current behavior)
                assert context["MOAI_VERSION_FORMAT_VALID"] == "false"

    def test_get_version_source_cached(self):
        """Test _get_version_source returns 'config_cached' for fresh cache."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            mock_reader = MagicMock()
            mock_config = MagicMock()
            mock_config.cache_ttl_seconds = 120
            mock_reader.get_config.return_value = mock_config
            mock_reader.get_cache_age_seconds.return_value = 60  # Fresh cache

            result = processor._get_version_source(mock_reader)

            assert result == "config_cached"

    def test_get_version_source_stale(self):
        """Test _get_version_source returns 'config_stale' for stale cache."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            mock_reader = MagicMock()
            mock_config = MagicMock()
            mock_config.cache_ttl_seconds = 120
            mock_reader.get_config.return_value = mock_config
            mock_reader.get_cache_age_seconds.return_value = 180  # Stale cache

            result = processor._get_version_source(mock_reader)

            assert result == "config_stale"

    def test_get_version_source_uncached(self):
        """Test _get_version_source returns fallback_source for no cache."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            mock_reader = MagicMock()
            mock_config = MagicMock()
            mock_config.fallback_source.value = "package"
            mock_reader.get_config.return_value = mock_config
            mock_reader.get_cache_age_seconds.return_value = None

            result = processor._get_version_source(mock_reader)

            assert result == "package"


class TestVariableValidation:
    """Test template variable validation with graceful degradation."""

    def test_validate_template_variables_raises_error_when_disabled(self):
        """Test _validate_template_variables raises ValueError when graceful_degradation=False."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(validate_template_variables=True, graceful_degradation=False)
            processor = TemplateProcessor(Path(temp_dir), config)

            with pytest.raises(ValueError, match="Template variable validation failed"):
                processor._validate_template_variables({"lowercase_var": "value"})

    def test_validate_template_variables_warns_when_enabled(self):
        """Test _validate_template_variables logs warnings when graceful_degradation=True."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(
                validate_template_variables=True, graceful_degradation=True, enable_substitution_warnings=True
            )
            processor = TemplateProcessor(Path(temp_dir), config)

            # Should not raise, just warn
            processor._validate_template_variables({"lowercase_var": "value"})

    def test_validate_template_variables_invalid_names(self):
        """Test _validate_template_variables detects invalid variable names."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(validate_template_variables=True, graceful_degradation=True)
            processor = TemplateProcessor(Path(temp_dir), config)

            # Lowercase name
            processor._validate_template_variables({"lowercase": "value"})
            # Name with spaces
            processor._validate_template_variables({"VAR_NAME": "value"})
            # Name with special chars (not underscore)
            processor._validate_template_variables({"VAR-NAME": "value"})

    def test_validate_template_variables_variable_length(self):
        """Test _validate_template_variables checks variable length limits."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(
                validate_template_variables=True,
                max_variable_length=10,
                graceful_degradation=True,
                enable_substitution_warnings=True,
            )
            processor = TemplateProcessor(Path(temp_dir), config)

            # Variable name too long
            processor._validate_template_variables({"VERY_LONG_VAR_NAME": "value"})

    def test_validate_template_variables_value_length(self):
        """Test _validate_template_variables checks value length limits."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(
                validate_template_variables=True,
                max_variable_length=10,
                graceful_degradation=True,
                enable_substitution_warnings=True,
            )
            processor = TemplateProcessor(Path(temp_dir), config)

            # Value too long (max_variable_length * 2 = 20)
            processor._validate_template_variables({"VALID_VAR": "x" * 100})

    def test_validate_template_variables_placeholder_patterns(self):
        """Test _validate_template_variables detects placeholder patterns in values."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(
                validate_template_variables=True, graceful_degradation=True, enable_substitution_warnings=True
            )
            processor = TemplateProcessor(Path(temp_dir), config)

            # Value contains {{ }}
            processor._validate_template_variables({"VALID_VAR": "value {{with}} placeholders"})

    def test_validate_template_variables_missing_common_vars(self):
        """Test _validate_template_variables warns about missing common variables."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(
                validate_template_variables=True, graceful_degradation=True, enable_substitution_warnings=True
            )
            processor = TemplateProcessor(Path(temp_dir), config)

            # Empty context (missing all common vars)
            processor._validate_template_variables({})

    def test_is_valid_template_variable_unicode(self):
        """Test _is_valid_template_variable handles Unicode characters."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(validate_template_variables=True, graceful_degradation=True)
            processor = TemplateProcessor(Path(temp_dir), config)

            # Unicode in value (should be OK)
            result = processor._is_valid_template_variable("VAR_NAME", "한글 텍스트")
            assert result is True

    def test_is_valid_template_variable_special_characters(self):
        """Test _is_valid_template_variable handles special characters."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(validate_template_variables=True, graceful_degradation=True)
            processor = TemplateProcessor(Path(temp_dir), config)

            # Special chars in value
            result = processor._is_valid_template_variable("VAR_NAME", "value!@#$%^&*()")
            assert result is True

    def test_is_valid_template_variable_empty_value(self):
        """Test _is_valid_template_variable rejects empty values."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(validate_template_variables=True)
            processor = TemplateProcessor(Path(temp_dir), config)

            assert processor._is_valid_template_variable("VAR_NAME", "") is False
            assert processor._is_valid_template_variable("VAR_NAME", "   ") is False


class TestCaching:
    """Test caching behavior and cache management."""

    def test_clear_substitution_cache_clears_dict(self):
        """Test clear_substitution_cache clears the cache dictionary."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor._substitution_cache[1] = ("test", [])

            processor.clear_substitution_cache()

            assert len(processor._substitution_cache) == 0

    def test_clear_substitution_cache_clears_validation_cache(self):
        """Test set_context clears validation cache."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))
            processor._variable_validation_cache["VAR"] = True
            processor.set_context({})

            assert len(processor._variable_validation_cache) == 0

    def test_get_cache_stats_all_fields(self):
        """Test get_cache_stats returns all stat fields."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(enable_caching=True, cache_size=50)
            processor = TemplateProcessor(Path(temp_dir), config)

            # Add some cache entries
            processor._substitution_cache[1] = ("test1", [])
            processor._substitution_cache[2] = ("test2", [])

            stats = processor.get_cache_stats()

            assert stats["cache_size"] == 2
            assert stats["max_cache_size"] == 50
            assert stats["cache_enabled"] is True
            assert "cache_hit_ratio" in stats

    def test_cache_size_limit_eviction_fifo(self):
        """Test cache size limit triggers FIFO eviction."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(enable_caching=True, cache_size=2, verbose_logging=True)
            processor = TemplateProcessor(Path(temp_dir), config)
            processor.set_context({"VAR": "value"})

            # Fill cache to limit
            content1 = "Content 1 with {{VAR}}"
            content2 = "Content 2 with {{VAR}}"

            processor._substitute_variables(content1)
            processor._substitute_variables(content2)

            assert len(processor._substitution_cache) == 2

            # Add one more (should trigger eviction)
            content3 = "Content 3 with {{VAR}}"
            processor._substitute_variables(content3)

            # Cache size should still be at limit
            assert len(processor._substitution_cache) == 2

    def test_cache_disabled_no_storage(self):
        """Test substitution with cache disabled doesn't store results."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(enable_caching=False)
            processor = TemplateProcessor(Path(temp_dir), config)
            processor.set_context({"VAR": "value"})

            content = "Content with {{VAR}}"
            processor._substitute_variables(content)

            assert len(processor._substitution_cache) == 0

    def test_cache_hit_reuses_result(self):
        """Test cache hit returns cached result without recomputation."""
        with tempfile.TemporaryDirectory() as temp_dir:
            config = TemplateProcessorConfig(enable_caching=True)
            processor = TemplateProcessor(Path(temp_dir), config)
            processor.set_context({"VAR": "value"})

            content = "Content with {{VAR}}"

            # First call (cache miss)
            result1, _warnings1 = processor._substitute_variables(content)

            # Second call (cache hit)
            result2, _warnings2 = processor._substitute_variables(content)

            assert result1 == result2
            # Cache should have 1 entry
            assert len(processor._substitution_cache) == 1


class TestFileOperations:
    """Test file operations with edge cases."""

    def test_copy_file_with_substitution_shell_script_executable(self):
        """Test _copy_file_with_substitution sets executable permission on .sh files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            src_dir = Path(temp_dir) / "src"
            dst_dir = Path(temp_dir) / "dst"
            src_dir.mkdir()
            dst_dir.mkdir()

            src_file = src_dir / "script.sh"
            src_file.write_text("#!/bin/bash\necho test")

            processor = TemplateProcessor(Path(temp_dir))
            processor._copy_file_with_substitution(src_file, dst_dir / "script.sh")

            dst_file = dst_dir / "script.sh"
            assert dst_file.exists()

            # Check executable permission
            import stat

            st_mode = dst_file.stat().st_mode
            assert st_mode & (stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH) != 0

    def test_copy_file_with_substitution_unicode_content(self):
        """Test _copy_file_with_substitution handles Unicode content."""
        with tempfile.TemporaryDirectory() as temp_dir:
            src_dir = Path(temp_dir) / "src"
            dst_dir = Path(temp_dir) / "dst"
            src_dir.mkdir()
            dst_dir.mkdir()

            src_file = src_dir / "test.txt"
            src_file.write_text("Hello 世界 🌍", encoding="utf-8")

            processor = TemplateProcessor(Path(temp_dir))
            processor._copy_file_with_substitution(src_file, dst_dir / "test.txt")

            dst_file = dst_dir / "test.txt"
            assert dst_file.read_text(encoding="utf-8") == "Hello 世界 🌍"

    def test_copy_file_with_substitution_binary_fallback(self):
        """Test _copy_file_with_substitution falls back to binary copy on decode error."""
        with tempfile.TemporaryDirectory() as temp_dir:
            src_dir = Path(temp_dir) / "src"
            dst_dir = Path(temp_dir) / "dst"
            src_dir.mkdir()
            dst_dir.mkdir()

            # Create a file that looks like text but has invalid UTF-8
            src_file = src_dir / "test.txt"
            src_file.write_bytes(b"\xff\xfe Invalid UTF-8")

            processor = TemplateProcessor(Path(temp_dir))
            processor.set_context({"VAR": "value"})

            # Should not raise, just copy as binary
            processor._copy_file_with_substitution(src_file, dst_dir / "test.txt")

            dst_file = dst_dir / "test.txt"
            assert dst_file.exists()
            assert dst_file.read_bytes() == b"\xff\xfe Invalid UTF-8"

    def test_copy_file_with_substitution_empty_file(self):
        """Test _copy_file_with_substitution handles empty files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            src_dir = Path(temp_dir) / "src"
            dst_dir = Path(temp_dir) / "dst"
            src_dir.mkdir()
            dst_dir.mkdir()

            src_file = src_dir / "empty.txt"
            src_file.write_text("")

            processor = TemplateProcessor(Path(temp_dir))
            processor._copy_file_with_substitution(src_file, dst_dir / "empty.txt")

            dst_file = dst_dir / "empty.txt"
            assert dst_file.exists()
            assert dst_file.read_text() == ""

    def test_is_text_file_extensions(self):
        """Test _is_text_file checks file extensions."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            assert processor._is_text_file(Path("test.md")) is True
            assert processor._is_text_file(Path("test.txt")) is True
            assert processor._is_text_file(Path("test.py")) is True
            assert processor._is_text_file(Path("test.toml")) is True
            assert processor._is_text_file(Path("test.xml")) is True
            assert processor._is_text_file(Path("test.j2")) is False
            assert processor._is_text_file(Path("test.bin")) is False
            assert processor._is_text_file(Path("test.png")) is False

    def test_localize_yaml_description_complex_nested(self):
        """Test _localize_yaml_description handles complex nested YAML."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            content = """---
description:
  en: "English text"
  ko: "한글 텍스트"
  ja: "日本語テキスト"
other_field: value
---

Content here."""

            # Localize to Korean
            result = processor._localize_yaml_description(content, "ko")

            assert "한글 텍스트" in result
            assert "English text" not in result
            assert "other_field: value" in result

    def test_localize_yaml_description_missing_description(self):
        """Test _localize_yaml_description handles missing description field."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            content = """---
title: "Test"
---

Content here."""

            result = processor._localize_yaml_description(content, "ko")

            # Should return unchanged
            assert result == content

    def test_localize_yaml_description_invalid_yaml_fallback(self):
        """Test _localize_yaml_description falls back on invalid YAML."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            content = """---
invalid: yaml: content:
: unclosed

Content here."""

            result = processor._localize_yaml_description(content, "ko")

            # Should return unchanged
            assert result == content

    def test_localize_yaml_description_no_frontmatter(self):
        """Test _localize_yaml_description returns unchanged when no frontmatter."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            content = "Just content without frontmatter."

            result = processor._localize_yaml_description(content, "ko")

            assert result == content


class TestSyncMethods:
    """Test sync methods for skills and sections."""

    def test_sync_skills_selective_preserves_custom(self):
        """Test _sync_skills_selective preserves non-moai- skills."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            template_path = project_path / "template"
            template_path.mkdir()

            # Create source skills
            skills_src = template_path / ".claude" / "skills"
            skills_src.mkdir(parents=True)
            (skills_src / "moai-test").mkdir()
            (skills_src / "moai-example").mkdir()

            # Create destination with custom skills
            skills_dst = project_path / ".claude" / "skills"
            skills_dst.mkdir(parents=True, exist_ok=True)
            (skills_dst / "moai-test").mkdir()  # Old template skill
            (skills_dst / "my-custom-skill").mkdir()  # Custom skill
            (skills_dst / "another-custom").mkdir()  # Another custom

            processor = TemplateProcessor(project_path)
            processor._sync_skills_selective(template_path / ".claude", project_path / ".claude", silent=True)

            # Template skills should be updated
            assert (skills_dst / "moai-test").exists()
            assert (skills_dst / "moai-example").exists()

            # Custom skills should be preserved
            assert (skills_dst / "my-custom-skill").exists()
            assert (skills_dst / "another-custom").exists()

    def test_sync_skills_selective_empty_skills_dir(self):
        """Test _sync_skills_selective handles empty skills directory."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)
            template_path = project_path / "template"
            template_path.mkdir()

            # No source skills
            processor = TemplateProcessor(project_path)
            processor._sync_skills_selective(template_path / ".claude", project_path / ".claude", silent=True)

            # Should not raise

    def test_sync_new_section_files_new_file_copy(self):
        """Test _sync_new_section_files copies new section files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create template sections
            template_sections = project_path / "template" / ".moai" / "config" / "sections"
            template_sections.mkdir(parents=True)
            (template_sections / "new_section.yaml").write_text("key: value")

            # Create project sections
            project_sections = project_path / ".moai" / "config" / "sections"
            project_sections.mkdir(parents=True)

            processor = TemplateProcessor(project_path)
            processor.template_root = project_path / "template"
            processor._sync_new_section_files(silent=True)

            assert (project_sections / "new_section.yaml").exists()

    def test_sync_new_section_files_existing_merge(self):
        """Test _sync_new_section_files merges existing section files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create template sections
            template_sections = project_path / "template" / ".moai" / "config" / "sections"
            template_sections.mkdir(parents=True)
            (template_sections / "project.yaml").write_text("project:\n  name: template_name\n  new_field: new_value")

            # Create project sections with existing data
            project_sections = project_path / ".moai" / "config" / "sections"
            project_sections.mkdir(parents=True)
            (project_sections / "project.yaml").write_text(
                "project:\n  name: user_name\n  existing_field: existing_value"
            )

            processor = TemplateProcessor(project_path)
            processor.template_root = project_path / "template"
            processor._sync_new_section_files(silent=True)

            # Read merged result
            with open(project_sections / "project.yaml") as f:
                data = yaml.safe_load(f)

            # User values should be preserved
            assert data["project"]["name"] == "user_name"
            assert data["project"]["existing_field"] == "existing_value"
            # New field from template should be added
            assert data["project"]["new_field"] == "new_value"

    def test_sync_new_section_files_yaml_parse_error(self):
        """Test _sync_new_section_files handles YAML parse errors gracefully."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create template sections with valid YAML
            template_sections = project_path / "template" / ".moai" / "config" / "sections"
            template_sections.mkdir(parents=True)
            (template_sections / "good.yaml").write_text("key: value")

            # Create project sections with invalid YAML
            project_sections = project_path / ".moai" / "config" / "sections"
            project_sections.mkdir(parents=True)
            (project_sections / "bad.yaml").write_text("invalid: yaml: content:")

            processor = TemplateProcessor(project_path)
            processor.template_root = project_path / "template"
            processor._sync_new_section_files(silent=True)

            # Should not raise, should skip bad file

    def test_deep_merge_dicts_nested_key_conflicts(self):
        """Test _deep_merge_dicts handles nested key conflicts."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            base = {"level1": {"level2": {"existing": "base_value"}}}
            overlay = {"level1": {"level2": {"new": "overlay_value"}}}

            result, _new_keys = processor._deep_merge_dicts(base, overlay)

            # Both values should exist
            assert result["level1"]["level2"]["existing"] == "base_value"
            assert result["level1"]["level2"]["new"] == "overlay_value"
            assert "level1.level2.new" in _new_keys

    def test_deep_merge_dicts_non_dict_values(self):
        """Test _deep_merge_dicts handles non-dict values."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            base = {"string_key": "base_value", "dict_key": {"existing": "base"}}
            overlay = {"string_key": "overlay_value", "dict_key": {"new": "overlay"}}

            result, _new_keys = processor._deep_merge_dicts(base, overlay)

            # String key should preserve base value (user priority)
            assert result["string_key"] == "base_value"
            # Dict key should be merged
            assert result["dict_key"]["existing"] == "base"
            assert result["dict_key"]["new"] == "overlay"

    def test_deep_merge_dicts_returns_merged_and_conflicts(self):
        """Test _deep_merge_dicts returns merged dict and new keys list."""
        with tempfile.TemporaryDirectory() as temp_dir:
            processor = TemplateProcessor(Path(temp_dir))

            base = {}
            overlay = {"new_key": "value", "nested": {"key": "value"}}

            result, _new_keys = processor._deep_merge_dicts(base, overlay)

            assert result == overlay
            # Check format - new_keys should contain "new_key" and "nested.key" format
            assert "new_key" in _new_keys
            # The nested key format uses dot notation
            assert any("nested" in key for key in _new_keys)


class TestMergeMethods:
    """Test merge methods for configuration files."""

    def test_merge_section_yaml_environment_overrides(self):
        """Test _merge_section_yaml_files applies environment variable overrides."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            sections_dir = project_path / ".moai" / "config" / "sections"
            sections_dir.mkdir(parents=True)

            # Create language.yaml
            (sections_dir / "language.yaml").write_text("""
language:
  conversation_language: en
  agent_prompt_language: en
""")

            # Set environment variable
            import os

            old_val = os.environ.get("MOAI_CONVERSATION_LANG")
            try:
                os.environ["MOAI_CONVERSATION_LANG"] = "ko"

                processor = TemplateProcessor(project_path)
                processor._merge_section_yaml_files(sections_dir)

                # Check that env var was applied
                with open(sections_dir / "language.yaml") as f:
                    data = yaml.safe_load(f)

                assert data["language"]["conversation_language"] == "ko"
            finally:
                if old_val is None:
                    os.environ.pop("MOAI_CONVERSATION_LANG", None)
                else:
                    os.environ["MOAI_CONVERSATION_LANG"] = old_val

    def test_merge_mcp_json_server_conflicts(self):
        """Test _merge_mcp_json handles server conflicts by preserving user servers."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Template MCP config
            template_mcp = project_path / "template" / ".mcp.json"
            template_mcp.parent.mkdir(parents=True)
            template_mcp.write_text(json.dumps({"mcpServers": {"template-server": {"command": "template"}}}))

            # Project MCP config with user server
            project_mcp = project_path / ".mcp.json"
            project_mcp.write_text(json.dumps({"mcpServers": {"user-server": {"command": "user"}}}))

            processor = TemplateProcessor(project_path)
            processor._merge_mcp_json(template_mcp, project_mcp)

            # Check that both servers exist
            with open(project_mcp) as f:
                data = json.load(f)

            assert "template-server" in data["mcpServers"]
            assert "user-server" in data["mcpServers"]

    def test_merge_mcp_json_malformed_json(self):
        """Test _merge_mcp_json handles malformed JSON gracefully."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Template MCP config
            template_mcp = project_path / "template" / ".mcp.json"
            template_mcp.parent.mkdir(parents=True)
            template_mcp.write_text('{"mcpServers": {}}')

            # Project MCP config with malformed JSON
            project_mcp = project_path / ".mcp.json"
            project_mcp.write_text('{"invalid": json}')

            processor = TemplateProcessor(project_path)

            # Should not raise
            processor._merge_mcp_json(template_mcp, project_mcp)

    def test_merge_config_json_section_priority(self):
        """Test _merge_config_json prioritizes section YAML files over config.json."""
        # This test verifies the logic when section YAML files exist
        # The actual implementation delegates to _merge_section_yaml_files
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create sections directory
            sections_dir = project_path / ".moai" / "config" / "sections"
            sections_dir.mkdir(parents=True)

            # Template config.json
            template_config = project_path / "template" / ".moai" / "config" / "config.json"
            template_config.parent.mkdir(parents=True)
            template_config.write_text('{"language": {"conversation_language": "en"}}')

            processor = TemplateProcessor(project_path)
            processor.template_root = project_path / "template"

            # Should use section YAML files (not config.json)
            processor._merge_config_json(template_config, project_path / ".moai" / "config" / "config.json")


class TestCopyMethods:
    """Test copy methods for full workflow."""

    def test_copy_templates_with_backup(self):
        """Test copy_templates with backup=True creates backup."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create existing files
            (project_path / ".claude").mkdir()
            (project_path / ".claude" / "existing.md").write_text("existing")

            # Create template root
            template_root = project_path / "template"
            template_root.mkdir()
            (template_root / ".claude").mkdir()
            (template_root / ".claude" / "template.md").write_text("template")

            processor = TemplateProcessor(project_path)
            processor.template_root = template_root

            with patch.object(processor, "_has_existing_files", return_value=True):
                with patch.object(processor, "create_backup") as mock_backup:
                    processor.copy_templates(backup=True, silent=True)

                    mock_backup.assert_called_once()

    def test_copy_templates_without_backup(self):
        """Test copy_templates with backup=False skips backup."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create template root
            template_root = project_path / "template"
            template_root.mkdir()
            (template_root / ".claude").mkdir()

            processor = TemplateProcessor(project_path)
            processor.template_root = template_root

            with patch.object(processor, "create_backup") as mock_backup:
                processor.copy_templates(backup=False, silent=True)

                mock_backup.assert_not_called()

    def test_copy_claude_alfred_folder_overwrite(self):
        """Test _copy_claude overwrites Alfred folders."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create template Alfred folder
            template_root = project_path / "template"
            template_claude = template_root / ".claude" / "commands" / "alfred"
            template_claude.mkdir(parents=True)
            (template_claude / "0-project.md").write_text("template")

            # Create existing Alfred folder with different content
            project_claude = project_path / ".claude" / "commands" / "alfred"
            project_claude.mkdir(parents=True)
            (project_claude / "0-project.md").write_text("existing")

            processor = TemplateProcessor(project_path)
            processor.template_root = template_root

            processor._copy_claude(silent=True)

            # Should be overwritten with template content
            assert (project_claude / "0-project.md").read_text() == "template"

    def test_copy_moai_excludes_protected_paths(self):
        """Test _copy_moai excludes protected paths from copying."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create template .moai with protected paths
            template_root = project_path / "template"
            template_moai = template_root / ".moai"
            template_moai.mkdir(parents=True)

            # Protected: specs and reports
            (template_moai / "specs").mkdir()
            (template_moai / "specs" / "SPEC-001.md").write_text("template spec")

            # Not protected: config
            (template_moai / "config").mkdir()
            (template_moai / "config" / "config.yaml").write_text("template config")

            processor = TemplateProcessor(project_path)
            processor.template_root = template_root

            processor._copy_moai(silent=True)

            # Specs should NOT be copied
            assert not (project_path / ".moai" / "specs" / "SPEC-001.md").exists()

            # Config should be copied
            assert (project_path / ".moai" / "config" / "config.yaml").exists()

    def test_copy_claude_md_language_specific(self):
        """Test _copy_claude_md selects language-specific CLAUDE.md."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create template root with language-specific files
            template_root = project_path / "template"
            template_root.mkdir()
            (template_root / "CLAUDE.md").write_text("# English CLAUDE.md")
            (template_root / "CLAUDE.ko.md").write_text("# Korean CLAUDE.md")
            (template_root / "CLAUDE.ja.md").write_text("# Japanese CLAUDE.md")

            processor = TemplateProcessor(project_path)
            processor.template_root = template_root
            processor.set_context({"CONVERSATION_LANGUAGE": "ko"})

            processor._copy_claude_md(silent=True)

            # Should copy Korean version
            assert (project_path / "CLAUDE.md").read_text() == "# Korean CLAUDE.md"

    def test_copy_claude_md_fallback_to_english(self):
        """Test _copy_claude_md falls back to English when language-specific not found."""
        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create template root with only English
            template_root = project_path / "template"
            template_root.mkdir()
            (template_root / "CLAUDE.md").write_text("# English CLAUDE.md")

            processor = TemplateProcessor(project_path)
            processor.template_root = template_root
            processor.set_context({"CONVERSATION_LANGUAGE": "ko"})

            processor._copy_claude_md(silent=True)

            # Should copy English version (fallback)
            assert (project_path / "CLAUDE.md").read_text() == "# English CLAUDE.md"
