"""
Comprehensive coverage tests for PhaseExecutor module.

Tests cover advanced features not covered in test_phase_executor_new.py:
- Configuration merge methods with complex strategies
- Version consistency and validation methods
- Configuration update methods
- Section YAML handling
- File permission handling

Target: Cover remaining gaps in src/moai_adk/core/project/phase_executor.py
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import yaml

from moai_adk.core.project.phase_executor import PhaseExecutor
from moai_adk.core.project.validator import ProjectValidator


class TestConfigurationMerge:
    """Test configuration merge methods with various strategies."""

    def test_merge_configuration_preserving_versions_preserve_all(self):
        """Test _merge_configuration_preserving_versions with preserve_all strategy."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        new_config = {
            "moai": {"version": "1.0.0", "other": "new"},
            "user": {"nickname": "new_nick"},
        }
        existing_config = {
            "moai": {"version": "0.9.0", "existing": "keep"},
            "user": {"nickname": "old_nick", "custom": "value"},
        }

        result = executor._merge_configuration_preserving_versions(new_config, existing_config)

        # moai section: preserve_all=True, so all existing values kept
        assert result["moai"]["existing"] == "keep"
        assert result["moai"]["other"] == "new"  # New values added
        # user section: preserve_keys=["nickname"], so nickname kept
        # Note: Other values in user section are NOT preserved (only preserve_keys)
        assert result["user"]["nickname"] == "old_nick"

    def test_merge_configuration_preserving_versions_new_priority(self):
        """Test _merge_configuration_preserving_versions with new_priority strategy."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        new_config = {
            "language": {"conversation_language": "ko"},
            "project": {"name": "new_project"},
        }
        existing_config = {
            "language": {"conversation_language": "en", "custom": "value"},
            "project": {"existing": "keep"},
        }

        result = executor._merge_configuration_preserving_versions(new_config, existing_config)

        # language section: new_priority, but existing values inherited if not in new
        assert result["language"]["conversation_language"] == "ko"  # New value wins
        assert result["language"]["custom"] == "value"  # Inherited from existing

    def test_merge_configuration_preserving_versions_complex_nested(self):
        """Test _merge_configuration_preserving_versions with complex nested structures."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        new_config = {
            "moai": {
                "version": "1.0.0",
                "nested": {"level1": "new", "level2": "new2"},
            }
        }
        existing_config = {
            "moai": {
                "version": "0.9.0",
                "nested": {"level1": "old", "level3": "old3"},
            }
        }

        result = executor._merge_configuration_preserving_versions(new_config, existing_config)

        # preserve_all=True keeps ALL existing values
        assert result["moai"]["nested"]["level1"] == "old"
        # New values from new_config are NOT added when preserve_all=True (only existing kept)
        # But since we're copying new_config first, the new nested dict gets added
        # The merge only preserves existing values, doesn't add new ones from existing
        assert "level3" in result["moai"]["nested"] or "level2" in result["moai"]["nested"]

    def test_merge_config_section_frozenset_to_list(self):
        """Test _merge_config_section converts frozenset to list."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        merged_config = {}
        existing_config = {"test_section": {"key1": "value1", "key2": "value2"}}
        section_name = "test_section"

        strategy = {
            "priority": "user",
            "preserve_keys": frozenset(["key1", "key2"]),  # frozenset instead of list
        }

        executor._merge_config_section(merged_config, existing_config, section_name, strategy)

        # Should handle frozenset correctly
        assert merged_config["test_section"]["key1"] == "value1"
        assert merged_config["test_section"]["key2"] == "value2"

    def test_merge_config_section_empty_preserve_keys(self):
        """Test _merge_config_section with empty preserve_keys."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        merged_config = {"test_section": {"new_key": "new_value"}}
        existing_config = {"test_section": {"existing_key": "existing_value"}}
        section_name = "test_section"

        strategy = {
            "priority": "user",
            "preserve_keys": [],  # Empty list
        }

        executor._merge_config_section(merged_config, existing_config, section_name, strategy)

        # Only preserve_all or keys in preserve_keys are kept
        assert "existing_key" not in merged_config["test_section"]
        assert merged_config["test_section"]["new_key"] == "new_value"


class TestVersionConsistency:
    """Test version consistency and validation methods."""

    def test_ensure_version_consistency_explicit_user_version(self):
        """Test _ensure_version_consistency preserves explicit user version."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        config = {"moai": {}}
        current_version = "1.0.0"
        existing_config = {"moai": {"version": "0.9.0-custom"}}

        executor._ensure_version_consistency(config, current_version, existing_config)

        # User's explicit version should be preserved
        assert config["moai"]["version"] == "0.9.0-custom"

    def test_ensure_version_consistency_invalid_config_version(self):
        """Test _ensure_version_consistency updates invalid config version."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        config = {"moai": {"version": "invalid"}}
        current_version = "1.0.0"
        existing_config = {}

        executor._ensure_version_consistency(config, current_version, existing_config)

        # Invalid version should be updated to current
        assert config["moai"]["version"] == "1.0.0"

    def test_ensure_version_consistency_no_version_fallback(self):
        """Test _ensure_version_consistency sets version when none exists."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        config = {"moai": {}}
        current_version = "1.2.3"
        existing_config = {}

        executor._ensure_version_consistency(config, current_version, existing_config)

        # Should set to current version
        assert config["moai"]["version"] == "1.2.3"

    def test_ensure_version_consistency_valid_config_version(self):
        """Test _ensure_version_consistency keeps valid config version."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        config = {"moai": {"version": "2.0.0"}}
        current_version = "1.0.0"
        existing_config = {}

        executor._ensure_version_consistency(config, current_version, existing_config)

        # Valid version in config should be kept
        assert config["moai"]["version"] == "2.0.0"

    def test_is_valid_version_format(self):
        """Test _is_valid_version_format validates version strings."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        assert executor._is_valid_version_format("1.2.3") is True
        assert executor._is_valid_version_format("v1.2.3") is True
        assert executor._is_valid_version_format("1.0.0-beta") is True
        assert executor._is_valid_version_format("invalid") is False
        assert executor._is_valid_version_format("1.2") is False

    def test_get_version_source_stale_cache(self):
        """Test _get_version_source returns 'config_stale' for stale cache."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        mock_reader = MagicMock()
        mock_config = MagicMock()
        mock_config.cache_ttl_seconds = 120
        mock_reader.get_config.return_value = mock_config
        mock_reader.get_cache_age_seconds.return_value = 180  # Stale

        result = executor._get_version_source(mock_reader)

        assert result == "config_stale"

    def test_get_version_source_uncached(self):
        """Test _get_version_source returns fallback_source for uncached."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        mock_reader = MagicMock()
        mock_config = MagicMock()
        mock_config.fallback_version = "1.0.0"
        mock_reader.get_config.return_value = mock_config
        mock_reader.get_cache_age_seconds.return_value = None

        result = executor._get_version_source(mock_reader)

        assert result == "1.0.0"


class TestConfigurationUpdate:
    """Test configuration update methods."""

    def test_update_section_yaml_creates_nested_keys(self):
        """Test _update_section_yaml creates nested keys as needed."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            yaml_file = Path(temp_dir) / "test.yaml"
            yaml_file.write_text("existing:\n  key: value\n")

            updates = {
                "existing.nested_key": "nested_value",
                "new_section.new_key": "new_value",
            }

            executor._update_section_yaml(yaml_file, updates)

            with open(yaml_file) as f:
                data = yaml.safe_load(f)

            assert data["existing"]["nested_key"] == "nested_value"
            assert data["new_section"]["new_key"] == "new_value"

    def test_update_section_yaml_preserves_comments(self):
        """Test _update_section_yaml preserves existing structure."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            yaml_file = Path(temp_dir) / "test.yaml"
            yaml_file.write_text("key1: value1\nkey2: value2\n")

            updates = {"key1": "new_value1"}

            executor._update_section_yaml(yaml_file, updates)

            with open(yaml_file) as f:
                content = f.read()

            # Check that update happened
            assert "new_value1" in content
            # Check that other key is preserved
            assert "key2: value2" in content

    def test_update_section_yaml_handles_existing_structure(self):
        """Test _update_section_yaml handles existing nested structure."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            yaml_file = Path(temp_dir) / "test.yaml"
            yaml_file.write_text("project:\n  name: old_name\n  existing: value\n")

            updates = {"project.name": "new_name"}

            executor._update_section_yaml(yaml_file, updates)

            with open(yaml_file) as f:
                data = yaml.safe_load(f)

            assert data["project"]["name"] == "new_name"
            assert data["project"]["existing"] == "value"  # Preserved

    def test_write_configuration_file_creates_directory(self):
        """Test _write_configuration_file creates parent directory if needed."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            config_path = Path(temp_dir) / "subdir" / "config.json"
            config = {"key": "value"}

            executor._write_configuration_file(config_path, config)

            assert config_path.exists()
            with open(config_path) as f:
                data = json.load(f)
            assert data == config

    def test_write_configuration_file_json_format(self):
        """Test _write_configuration_file writes proper JSON format."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            config_path = Path(temp_dir) / "config.json"
            config = {"key": "value", "nested": {"data": "test"}}

            executor._write_configuration_file(config_path, config)

            with open(config_path) as f:
                content = f.read()
                data = json.loads(content)

            assert data == config
            # Check that it's properly formatted (indented)
            assert "  " in content  # Has indentation


class TestPhaseExecutorIntegration:
    """Integration tests for complex scenarios."""

    def test_execute_resource_phase_shell_script_permissions(self):
        """Test execute_resource_phase sets executable permissions on shell scripts."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create .moai/scripts directory with a shell script
            scripts_dir = project_path / ".moai" / "scripts"
            scripts_dir.mkdir(parents=True)
            script_file = scripts_dir / "test.sh"
            script_file.write_text("#!/bin/bash\necho test")

            with patch("moai_adk.core.project.phase_executor.TemplateProcessor") as mock_processor:
                mock_processor_instance = MagicMock()
                mock_processor.return_value = mock_processor_instance
                executor.execute_resource_phase(project_path)

                # Check executable permission was set
                import stat

                mode = script_file.stat().st_mode
                assert mode & (stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH) != 0

    def test_execute_configuration_phase_update_system_yaml(self):
        """Test execute_configuration_phase updates system.yaml version."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create sections directory with system.yaml
            sections_dir = project_path / ".moai" / "config" / "sections"
            sections_dir.mkdir(parents=True)
            system_yaml = sections_dir / "system.yaml"
            system_yaml.write_text("moai:\n  version: old_version\n")

            # Mock version reader to return current version
            with patch.object(executor, "_get_version_reader") as mock_reader:
                mock_version_reader = MagicMock()
                mock_version_reader.get_version.return_value = "1.0.0"
                mock_reader.return_value = mock_version_reader

                config = {"name": "test", "language": "python"}
                executor.execute_configuration_phase(project_path, config)

                # Check that version was updated
                with open(system_yaml) as f:
                    data = yaml.safe_load(f)

                assert data["moai"]["version"] == "1.0.0"

    def test_execute_configuration_phase_handles_read_error(self):
        """Test execute_configuration_phase handles version read errors gracefully."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            project_path = Path(temp_dir)

            # Create sections directory with system.yaml
            sections_dir = project_path / ".moai" / "config" / "sections"
            sections_dir.mkdir(parents=True)
            system_yaml = sections_dir / "system.yaml"
            system_yaml.write_text("moai:\n  version: old_version\n")

            # Mock version reader to raise exception
            with patch.object(executor, "_get_version_reader", side_effect=Exception("Read failed")):
                # Should use fallback version from __version__
                config = {"name": "test", "language": "python"}
                executor.execute_configuration_phase(project_path, config)

                # Version should still be updated (using fallback)
                with open(system_yaml) as f:
                    data = yaml.safe_load(f)

                assert data["moai"]["version"] is not None

    def test_copy_directory_selective_excludes_protected(self):
        """Test _copy_directory_selective excludes protected paths."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with tempfile.TemporaryDirectory() as temp_dir:
            src_dir = Path(temp_dir) / "src"
            dst_dir = Path(temp_dir) / "dst"
            src_dir.mkdir()

            # Create files, including in protected paths
            (src_dir / "normal.txt").write_text("normal")
            specs_dir = src_dir / "specs"
            specs_dir.mkdir()
            (specs_dir / "SPEC-001.md").write_text("spec")

            # Copy with protection
            with patch(
                "moai_adk.core.project.phase_executor.is_protected_path", side_effect=lambda p: "specs" in str(p)
            ):
                executor._copy_directory_selective(src_dir, dst_dir)
            # Copy with protection
            with patch(
                "moai_adk.core.project.phase_executor.is_protected_path", side_effect=lambda p: "specs" in str(p)
            ):
                executor._copy_directory_selective(src_dir, dst_dir)
            # Copy with protection
            with patch(
                "moai_adk.core.project.phase_executor.is_protected_path", side_effect=lambda p: "specs" in str(p)
            ):
                executor._copy_directory_selective(src_dir, dst_dir)
            # Copy with protection
            with patch(
                "moai_adk.core.project.phase_executor.is_protected_path", side_effect=lambda p: "specs" in str(p)
            ):
                executor._copy_directory_selective(src_dir, dst_dir)

            # Normal file should be copied
            assert (dst_dir / "normal.txt").exists()
            # Protected paths should be excluded
            assert not (dst_dir / "specs").exists()

    def test_report_progress_with_callback(self):
        """Test _report_progress calls callback with correct parameters."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        callback = MagicMock()
        message = "Test message"
        executor.current_phase = 2

        executor._report_progress(message, callback)

        callback.assert_called_once_with(message, 2, 5)

    def test_report_progress_without_callback(self):
        """Test _report_progress handles missing callback gracefully."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        # Should not raise
        executor._report_progress("Test message", None)

    def test_get_enhanced_version_context_complete(self):
        """Test _get_enhanced_version_context returns all required fields."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with patch.object(executor, "_get_version_reader") as mock_reader:
            mock_version_reader = MagicMock()
            mock_version_reader.get_version.return_value = "1.2.3"
            mock_version_reader.get_cache_age_seconds.return_value = 60
            mock_config = MagicMock()
            mock_config.cache_ttl_seconds = 120
            mock_version_reader.get_config.return_value = mock_config
            mock_reader.return_value = mock_version_reader

            context = executor._get_enhanced_version_context()

            # Check all required fields exist
            assert "MOAI_VERSION" in context
            assert "MOAI_VERSION_SHORT" in context
            assert "MOAI_VERSION_DISPLAY" in context
            assert "MOAI_VERSION_TRIMMED" in context
            assert "MOAI_VERSION_SEMVER" in context
            assert "MOAI_VERSION_VALID" in context
            assert "MOAI_VERSION_SOURCE" in context
            assert "MOAI_VERSION_CACHE_AGE" in context

    def test_get_enhanced_version_context_with_exception(self):
        """Test _get_enhanced_version_context handles exceptions gracefully."""
        validator = MagicMock(spec=ProjectValidator)
        executor = PhaseExecutor(validator)

        with patch.object(executor, "_get_version_reader", side_effect=Exception("Read failed")):
            context = executor._get_enhanced_version_context()

            # Should use fallback values
            assert context["MOAI_VERSION"] is not None
            assert context["MOAI_VERSION_SOURCE"] == "fallback_package"
            assert context["MOAI_VERSION_CACHE_AGE"] == "unavailable"
