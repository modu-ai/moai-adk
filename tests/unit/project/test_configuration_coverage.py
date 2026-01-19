"""Comprehensive TDD tests for moai_adk.project.configuration module.

Targets 100% coverage for all classes and methods.
Tests cover edge cases, error conditions, and all code paths.
"""

import json
import tempfile
from pathlib import Path
from unittest.mock import patch

import pytest
import yaml

from moai_adk.project.configuration import (
    AutoDetectionEngine,
    ConditionalBatchRenderer,
    ConfigurationCoverageValidator,
    ConfigurationManager,
    ConfigurationMigrator,
    SmartDefaultsEngine,
    TabSchemaValidator,
    TemplateVariableInterpolator,
)


class TestConfigurationManagerCoverage:
    """Test ConfigurationManager for 100% coverage."""

    def test_load_yaml_file(self):
        """Test load method with YAML file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.yaml"
            config_data = {"user": {"name": "TestUser"}}
            config_path.write_text(yaml.safe_dump(config_data))

            manager = ConfigurationManager(config_path)
            result = manager.load()
            assert result == config_data
            assert manager._config_cache == config_data

    def test_load_yml_file(self):
        """Test load method with .yml extension."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.yml"
            config_data = {"project": {"name": "TestProject"}}
            config_path.write_text(yaml.safe_dump(config_data))

            manager = ConfigurationManager(config_path)
            result = manager.load()
            assert result == config_data

    def test_save_yaml_file(self):
        """Test save method with YAML file (line 70-71)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.yaml"
            config_data = {
                "user": {"name": "User"},
                "language": {"conversation_language": "en", "agent_prompt_language": "en"},
                "project": {"name": "Project", "documentation_mode": "skip"},
                "git_strategy": {"mode": "personal"},
                "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
            }

            manager = ConfigurationManager(config_path)
            result = manager.save(config_data)

            assert result is True
            assert config_path.exists()
            assert manager._config_cache == config_data

    def test_save_creates_backup(self):
        """Test save creates backup before overwriting (lines 196-202)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.yaml"
            backup_path = Path(tmpdir) / "config.backup"
            original_data = {
                "user": {"name": "Original"},
                "language": {"conversation_language": "en", "agent_prompt_language": "en"},
                "project": {"name": "Project", "documentation_mode": "skip"},
                "git_strategy": {"mode": "personal"},
                "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
            }
            new_data = {
                "user": {"name": "New"},
                "language": {"conversation_language": "en", "agent_prompt_language": "en"},
                "project": {"name": "Project", "documentation_mode": "skip"},
                "git_strategy": {"mode": "personal"},
                "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
            }

            # Create original config
            config_path.write_text(yaml.safe_dump(original_data))

            manager = ConfigurationManager(config_path)
            manager.save(new_data)

            # Verify backup was created
            assert backup_path.exists()
            assert yaml.safe_load(backup_path.read_text()) == original_data
            assert yaml.safe_load(config_path.read_text()) == new_data

    def test_save_missing_required_fields(self):
        """Test save raises ValueError for missing required fields."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.yaml"
            incomplete_config = {"user": {"name": "Test"}}  # Missing many required fields

            manager = ConfigurationManager(config_path)
            with pytest.raises(ValueError, match="missing required fields"):
                manager.save(incomplete_config)

    def test_write_config_internal_method(self):
        """Test _write_config internal method (line 86)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.yaml"
            config_data = {
                "user": {"name": "User"},
                "language": {"conversation_language": "en", "agent_prompt_language": "en"},
                "project": {"name": "Project", "documentation_mode": "skip"},
                "git_strategy": {"mode": "personal"},
                "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
            }

            manager = ConfigurationManager(config_path)
            manager._write_config(config_data)

            assert config_path.exists()
            assert manager._config_cache == config_data

    def test_save_creates_directory(self):
        """Test save creates parent directories if needed."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "deep" / "nested" / "config.yaml"
            config_data = {
                "user": {"name": "User"},
                "language": {"conversation_language": "en", "agent_prompt_language": "en"},
                "project": {"name": "Project", "documentation_mode": "skip"},
                "git_strategy": {"mode": "personal"},
                "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
            }

            manager = ConfigurationManager(config_path)
            result = manager.save(config_data)

            assert result is True
            assert config_path.exists()
            assert config_path.parent.exists()

    def test_save_json_file(self):
        """Test save method with JSON file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_data = {
                "user": {"name": "User"},
                "language": {"conversation_language": "en", "agent_prompt_language": "en"},
                "project": {"name": "Project", "documentation_mode": "skip"},
                "git_strategy": {"mode": "personal"},
                "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
            }

            manager = ConfigurationManager(config_path)
            result = manager.save(config_data)

            assert result is True
            loaded = json.loads(config_path.read_text())
            assert loaded == config_data

    def test_save_atomic_operation_failure_cleanup(self):
        """Test save cleans up temp file on failure."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.yaml"
            temp_path = config_path.with_suffix(".tmp")
            incomplete_config = {"user": {"name": "Test"}}

            manager = ConfigurationManager(config_path)
            try:
                manager.save(incomplete_config)
            except ValueError:
                pass

            # Temp file should be cleaned up
            assert not temp_path.exists()

    def test_parse_responses_unknown_keys(self):
        """Test _parse_responses handles unknown keys (lines 145-151)."""
        manager = ConfigurationManager()
        responses = {
            "user_name": "GOOS",
            "unknown_key": "unknown_value",  # Unknown key
            "another_unknown": 123,
        }
        result = manager._parse_responses(responses)
        assert result["unknown_key"] == "unknown_value"
        assert result["another_unknown"] == 123

    def test_parse_responses_nested_input(self):
        """Test _parse_responses handles nested dict input."""
        manager = ConfigurationManager()
        responses = {
            "user": {"name": "GOOS"},
            "project": {"name": "MoAI", "description": "Test"},
        }
        result = manager._parse_responses(responses)
        assert result["user"]["name"] == "GOOS"
        assert result["project"]["name"] == "MoAI"
        assert result["project"]["description"] == "Test"

    def test_build_from_responses_complete_workflow(self):
        """Test build_from_responses with all engines."""
        manager = ConfigurationManager()
        responses = {
            "user_name": "GOOS",
            "conversation_language": "en",
            "agent_prompt_language": "en",
            "project_name": "MoAI",
            "git_strategy_mode": "personal",
            "test_coverage_target": 90,
            "enforce_tdd": True,
            "documentation_mode": "skip",
        }

        result = manager.build_from_responses(responses)
        assert result["user"]["name"] == "GOOS"
        assert result["git_strategy"]["mode"] == "personal"
        # Should have smart defaults applied
        assert "workflow" in result.get("git_strategy", {}).get("personal", {})

    def test_validate_complete_all_required_fields(self):
        """Test _validate_complete with all required fields."""
        manager = ConfigurationManager()
        config = {
            "user": {"name": "User"},
            "language": {"conversation_language": "en", "agent_prompt_language": "en"},
            "project": {"name": "Project", "documentation_mode": "skip"},
            "git_strategy": {"mode": "personal"},
            "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
        }
        assert manager._validate_complete(config) is True

    def test_validate_complete_missing_required_field(self):
        """Test _validate_complete with missing required field."""
        manager = ConfigurationManager()
        config = {
            "user": {"name": "User"},
            "language": {"conversation_language": "en"},  # Missing agent_prompt_language
            "project": {"name": "Project"},
            "git_strategy": {"mode": "personal"},
            "constitution": {"test_coverage_target": 85, "enforce_tdd": True},
        }
        assert manager._validate_complete(config) is False


class TestSmartDefaultsEngineCoverage:
    """Test SmartDefaultsEngine for 100% coverage."""

    def test_apply_defaults_skips_empty_string_defaults(self):
        """Test apply_defaults skips auto-detect fields with empty defaults (line 311-312)."""
        engine = SmartDefaultsEngine()
        config = {}

        result = engine.apply_defaults(config)

        # Auto-detect fields should not be set (empty strings are skipped)
        # These fields won't exist in the result since empty string defaults are skipped
        # Only non-empty defaults are applied
        assert "git_strategy" in result
        assert "personal" in result["git_strategy"]
        assert "constitution" in result
        assert "language" in result
        assert "project" in result

        # Check that non-empty defaults were applied
        assert result["git_strategy"]["personal"]["workflow"] == "github-flow"
        assert result["constitution"]["test_coverage_target"] == 85

    def test_apply_defaults_preserves_existing_values(self):
        """Test apply_defaults preserves existing non-None values."""
        engine = SmartDefaultsEngine()
        config = {
            "git_strategy": {"personal": {"workflow": "custom-workflow"}},
            "constitution": {"test_coverage_target": 95},
            "language": {"agent_prompt_language": "ko"},
            "project": {"description": "Custom description"},
        }

        result = engine.apply_defaults(config)

        # Existing values should be preserved
        assert result["git_strategy"]["personal"]["workflow"] == "custom-workflow"
        assert result["constitution"]["test_coverage_target"] == 95
        assert result["language"]["agent_prompt_language"] == "ko"
        assert result["project"]["description"] == "Custom description"

    def test_apply_defaults_creates_nested_structure(self):
        """Test apply_defaults creates necessary nested structure (lines 296-307)."""
        engine = SmartDefaultsEngine()
        config = {}  # Empty config

        result = engine.apply_defaults(config)

        # All nested structures should exist
        assert "git_strategy" in result
        assert "personal" in result["git_strategy"]
        assert "team" in result["git_strategy"]
        assert "constitution" in result
        assert "language" in result
        assert "project" in result

    def test_apply_defaults_sets_none_values(self):
        """Test apply_defaults sets values when current is None."""
        engine = SmartDefaultsEngine()
        config = {
            "git_strategy": {"personal": {"workflow": None}},
            "constitution": {"test_coverage_target": None},
        }

        result = engine.apply_defaults(config)

        # None values should be replaced with defaults
        assert result["git_strategy"]["personal"]["workflow"] == "github-flow"
        assert result["constitution"]["test_coverage_target"] == 85

    def test_get_default_returns_none_for_unknown_field(self):
        """Test get_default returns None for unknown field."""
        engine = SmartDefaultsEngine()
        assert engine.get_default("unknown.field.path") is None

    def test_get_all_defaults_returns_copy(self):
        """Test get_all_defaults returns a deep copy."""
        engine = SmartDefaultsEngine()
        defaults1 = engine.get_all_defaults()
        defaults2 = engine.get_all_defaults()

        # Modify one copy
        defaults1["git_strategy.personal.workflow"] = "modified"

        # Other copy should be unchanged
        assert defaults2["git_strategy.personal.workflow"] == "github-flow"


class TestAutoDetectionEngineCoverage:
    """Test AutoDetectionEngine for 100% coverage."""

    def test_detect_and_apply_creates_sections(self):
        """Test detect_and_apply creates missing sections (lines 336-341)."""
        engine = AutoDetectionEngine()
        config = {
            "language": {"conversation_language": "ko"},
        }

        result = engine.detect_and_apply(config)

        assert "project" in result
        assert "language" in result
        assert "moai" in result

    def test_detect_language_javascript(self):
        """Test detect_language detects JavaScript (line 393)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create package.json without pyproject.toml or tsconfig.json
            pkg_json = Path(tmpdir) / "package.json"
            pkg_json.write_text('{"name": "test"}')

            with patch("pathlib.Path.cwd", return_value=Path(tmpdir)):
                lang = AutoDetectionEngine.detect_language()
                assert lang == "javascript"

    def test_detect_language_go(self):
        """Test detect_language detects Go (line 397)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create go.mod without other indicators
            go_mod = Path(tmpdir) / "go.mod"
            go_mod.write_text("module test")

            with patch("pathlib.Path.cwd", return_value=Path(tmpdir)):
                lang = AutoDetectionEngine.detect_language()
                assert lang == "go"

    def test_detect_language_default_python(self):
        """Test detect_language defaults to Python."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Empty directory
            with patch("pathlib.Path.cwd", return_value=Path(tmpdir)):
                lang = AutoDetectionEngine.detect_language()
                assert lang == "python"

    def test_detect_language_priority_order(self):
        """Test detect_language priority: TypeScript > Python > JavaScript > Go."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create all files - TypeScript should win
            (Path(tmpdir) / "tsconfig.json").write_text("{}")
            (Path(tmpdir) / "pyproject.toml").write_text("[project]")
            (Path(tmpdir) / "package.json").write_text("{}")
            (Path(tmpdir) / "go.mod").write_text("module test")

            with patch("pathlib.Path.cwd", return_value=Path(tmpdir)):
                lang = AutoDetectionEngine.detect_language()
                assert lang == "typescript"

    def test_detect_locale_mapping(self):
        """Test detect_locale maps language codes correctly."""
        assert AutoDetectionEngine.detect_locale("ko") == "ko_KR"
        assert AutoDetectionEngine.detect_locale("en") == "en_US"
        assert AutoDetectionEngine.detect_locale("ja") == "ja_JP"
        assert AutoDetectionEngine.detect_locale("zh") == "zh_CN"
        assert AutoDetectionEngine.detect_locale("unknown") == "en_US"  # Default

    def test_detect_language_name_mapping(self):
        """Test detect_language_name maps codes to names."""
        assert AutoDetectionEngine.detect_language_name("ko") == "Korean"
        assert AutoDetectionEngine.detect_language_name("en") == "English"
        assert AutoDetectionEngine.detect_language_name("ja") == "Japanese"
        assert AutoDetectionEngine.detect_language_name("zh") == "Chinese"
        assert AutoDetectionEngine.detect_language_name("unknown") == "English"  # Default

    def test_detect_template_version(self):
        """Test detect_template_version returns version."""
        version = AutoDetectionEngine.detect_template_version()
        assert isinstance(version, str)
        assert len(version) > 0

    def test_detect_moai_version(self):
        """Test detect_moai_version returns version."""
        version = AutoDetectionEngine.detect_moai_version()
        assert isinstance(version, str)
        assert len(version) > 0


class TestConditionalBatchRendererCoverage:
    """Test ConditionalBatchRenderer for 100% coverage."""

    def test_get_visible_batches_exact_match(self):
        """Test get_visible_batches with exact tab ID match (line 665)."""
        schema = {
            "tabs": [
                {
                    "id": "tab_3_git_automation",
                    "batches": [
                        {"id": "batch_1", "show_if": "mode == 'personal'"},
                    ],
                }
            ]
        }
        renderer = ConditionalBatchRenderer(schema)
        batches = renderer.get_visible_batches("tab_3_git_automation", {"mode": "personal"})

        assert len(batches) == 1
        assert batches[0]["id"] == "batch_1"

    def test_get_visible_batches_partial_match(self):
        """Test get_visible_batches with partial tab ID match."""
        schema = {
            "tabs": [
                {
                    "id": "tab_3_git_automation",
                    "batches": [
                        {"id": "batch_1", "show_if": "mode == 'personal'"},
                    ],
                }
            ]
        }
        renderer = ConditionalBatchRenderer(schema)
        batches = renderer.get_visible_batches("tab_3", {"mode": "personal"})

        assert len(batches) == 1

    def test_get_visible_batches_maps_mode_to_git_strategy_mode(self):
        """Test get_visible_batches maps 'mode' to 'git_strategy_mode' (lines 658-661)."""
        schema = {
            "tabs": [
                {
                    "id": "tab_3",
                    "batches": [
                        {"id": "batch_1", "show_if": "git_strategy_mode == 'personal'"},
                    ],
                }
            ]
        }
        renderer = ConditionalBatchRenderer(schema)
        batches = renderer.get_visible_batches("tab_3", {"mode": "personal"})

        assert len(batches) == 1

    def test_evaluate_condition_empty_returns_true(self):
        """Test evaluate_condition returns True for empty condition."""
        renderer = ConditionalBatchRenderer({})
        assert renderer.evaluate_condition("", {"mode": "personal"}) is True

    def test_evaluate_condition_true_string(self):
        """Test evaluate_condition returns True for 'true' string."""
        renderer = ConditionalBatchRenderer({})
        assert renderer.evaluate_condition("true", {"mode": "personal"}) is True

    def test_evaluate_condition_exception_returns_true(self):
        """Test evaluate_condition returns True on exception (lines 711-713)."""
        renderer = ConditionalBatchRenderer({})
        # Malformed expression will cause exception
        assert renderer.evaluate_condition("malformed [[[", {}) is True

    def test_safe_evaluate_or_operator(self):
        """Test _safe_evaluate handles OR operator."""
        result = ConditionalBatchRenderer._safe_evaluate("mode == 'personal' OR mode == 'team'", {"mode": "personal"})
        assert result is True

        result = ConditionalBatchRenderer._safe_evaluate("mode == 'personal' OR mode == 'team'", {"mode": "other"})
        assert result is False

    def test_safe_evaluate_and_operator(self):
        """Test _safe_evaluate handles AND operator."""
        result = ConditionalBatchRenderer._safe_evaluate(
            "mode == 'personal' AND documentation_mode == 'full_now'",
            {"mode": "personal", "documentation_mode": "full_now"},
        )
        assert result is True

        result = ConditionalBatchRenderer._safe_evaluate(
            "mode == 'personal' AND documentation_mode == 'full_now'",
            {"mode": "personal", "documentation_mode": "skip"},
        )
        assert result is False

    def test_evaluate_comparison_equals(self):
        """Test _evaluate_comparison with == operator."""
        result = ConditionalBatchRenderer._evaluate_comparison("mode == 'personal'", {"mode": "personal"})
        assert result is True

    def test_evaluate_comparison_not_equals(self):
        """Test _evaluate_comparison with != operator."""
        result = ConditionalBatchRenderer._evaluate_comparison("mode != 'personal'", {"mode": "team"})
        assert result is True

    def test_evaluate_comparison_less_than(self):
        """Test _evaluate_comparison with < operator."""
        result = ConditionalBatchRenderer._evaluate_comparison("count < 5", {"count": 3})
        assert result is True

    def test_evaluate_comparison_greater_than(self):
        """Test _evaluate_comparison with > operator."""
        result = ConditionalBatchRenderer._evaluate_comparison("count > 5", {"count": 10})
        assert result is True

    def test_evaluate_comparison_less_than_or_equal(self):
        """Test _evaluate_comparison with <= operator."""
        result = ConditionalBatchRenderer._evaluate_comparison("count <= 5", {"count": 5})
        assert result is True

    def test_evaluate_comparison_greater_than_or_equal(self):
        """Test _evaluate_comparison with >= operator."""
        result = ConditionalBatchRenderer._evaluate_comparison("count >= 5", {"count": 5})
        assert result is True

    def test_resolve_operand_string_literal(self):
        """Test _resolve_operand with string literal."""
        result = ConditionalBatchRenderer._resolve_operand("'personal'", {})
        assert result == "personal"

    def test_resolve_operand_double_quoted_string(self):
        """Test _resolve_operand with double-quoted string."""
        result = ConditionalBatchRenderer._resolve_operand('"personal"', {})
        assert result == "personal"

    def test_resolve_operand_float_number(self):
        """Test _resolve_operand with float number (lines 801-802)."""
        result = ConditionalBatchRenderer._resolve_operand("3.14", {})
        assert result == 3.14
        assert isinstance(result, float)

    def test_resolve_operand_int_number(self):
        """Test _resolve_operand with integer number."""
        result = ConditionalBatchRenderer._resolve_operand("42", {})
        assert result == 42
        assert isinstance(result, int)

    def test_resolve_operand_variable_from_context(self):
        """Test _resolve_operand looks up variable in context."""
        result = ConditionalBatchRenderer._resolve_operand("mode", {"mode": "personal"})
        assert result == "personal"

    def test_resolve_operand_unknown_raises_error(self):
        """Test _resolve_operand raises ValueError for unknown operand (line 809)."""
        with pytest.raises(ValueError, match="Unknown operand"):
            ConditionalBatchRenderer._resolve_operand("unknown_var", {})


class TestTemplateVariableInterpolatorCoverage:
    """Test TemplateVariableInterpolator for 100% coverage."""

    def test_interpolate_basic_variables(self):
        """Test interpolate replaces basic variables."""
        config = {"user": {"name": "GOOS"}, "project": {"name": "MoAI"}}
        template = "User: {{user.name}}, Project: {{project.name}}"
        result = TemplateVariableInterpolator.interpolate(template, config)

        assert result == "User: GOOS, Project: MoAI"

    def test_interpolate_missing_variable_raises_keyerror(self):
        """Test interpolate raises KeyError for missing variable."""
        config = {"user": {"name": "GOOS"}}
        template = "User: {{user.name}}, Project: {{project.name}}"

        with pytest.raises(KeyError, match="not found in config"):
            TemplateVariableInterpolator.interpolate(template, config)

    def test_get_nested_value_found(self):
        """Test _get_nested_value returns value for valid path."""
        config = {"user": {"name": "Test"}, "project": {"name": "P1"}}
        result = TemplateVariableInterpolator._get_nested_value(config, "user.name")
        assert result == "Test"

    def test_get_nested_value_not_found_returns_none(self):
        """Test _get_nested_value returns None for invalid path."""
        config = {"user": {"name": "Test"}}
        result = TemplateVariableInterpolator._get_nested_value(config, "missing.path")
        assert result is None

    def test_get_nested_value_non_dict_in_path(self):
        """Test _get_nested_value when intermediate value is not a dict."""
        config = {"user": "not_a_dict"}
        result = TemplateVariableInterpolator._get_nested_value(config, "user.name")
        assert result is None


class TestConfigurationMigratorCoverage:
    """Test ConfigurationMigrator for 100% coverage."""

    def test_load_legacy_config_returns_copy(self):
        """Test load_legacy_config returns deep copy."""
        migrator = ConfigurationMigrator()
        v2_config = {"version": "2.1.0", "user": {"name": "Test"}}
        loaded = migrator.load_legacy_config(v2_config)

        # Modify original
        v2_config["user"]["name"] = "Modified"

        # Loaded copy should be unchanged
        assert loaded["user"]["name"] == "Test"

    def test_migrate_creates_v3_structure(self):
        """Test migrate creates complete v3 structure."""
        migrator = ConfigurationMigrator()
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "OldUser"},
            "language": {"conversation_language": "en"},
            "project": {"name": "OldProject"},
            "git_strategy": {"mode": "personal"},
            "constitution": {"test_coverage_target": 80},
        }

        v3_config = migrator.migrate(v2_config)

        # Check v3 structure exists
        assert v3_config["version"] == "3.0.0"
        assert "personal" in v3_config["git_strategy"]
        assert "team" in v3_config["git_strategy"]
        assert "moai" in v3_config

    def test_migrate_preserves_v2_fields(self):
        """Test migrate preserves compatible v2 fields."""
        migrator = ConfigurationMigrator()
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "OldUser", "email": "old@example.com"},
            "language": {"conversation_language": "ko"},
        }

        v3_config = migrator.migrate(v2_config)

        assert v3_config["user"]["name"] == "OldUser"
        assert v3_config["user"]["email"] == "old@example.com"
        assert v3_config["language"]["conversation_language"] == "ko"

    def test_migrate_applies_smart_defaults(self):
        """Test migrate applies smart defaults for new v3 fields."""
        migrator = ConfigurationMigrator()
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "OldUser"},
        }

        v3_config = migrator.migrate(v2_config)

        # Should have smart defaults applied
        assert "workflow" in v3_config["git_strategy"]["personal"]
        assert "workflow" in v3_config["git_strategy"]["team"]

    def test_migrate_handles_partial_v2_config(self):
        """Test migrate handles v2 config with missing sections (lines 993-999)."""
        migrator = ConfigurationMigrator()
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "OldUser"},
            # Missing language, project, git_strategy, constitution
        }

        v3_config = migrator.migrate(v2_config)

        # Should create all sections
        assert "user" in v3_config
        assert "language" in v3_config
        assert "project" in v3_config
        assert "git_strategy" in v3_config
        assert "constitution" in v3_config


class TestTabSchemaValidatorCoverage:
    """Test TabSchemaValidator for 100% coverage."""

    def test_validate_valid_schema(self):
        """Test validate returns no errors for valid schema."""
        schema = {
            "version": "3.0.0",
            "tabs": [
                {
                    "id": "tab_1",
                    "batches": [
                        {
                            "questions": [
                                {
                                    "question": "Test?",
                                    "header": "Short",
                                    "options": [
                                        {"label": "A", "value": "a"},
                                        {"label": "B", "value": "b"},
                                    ],
                                }
                            ]
                        }
                    ],
                },
                {
                    "id": "tab_2",
                    "batches": [
                        {
                            "questions": [
                                {
                                    "question": "Test?",
                                    "header": "Short",
                                    "options": [
                                        {"label": "A", "value": "a"},
                                        {"label": "B", "value": "b"},
                                    ],
                                }
                            ]
                        }
                    ],
                },
                {
                    "id": "tab_3",
                    "batches": [
                        {
                            "questions": [
                                {
                                    "question": "Test?",
                                    "header": "Short",
                                    "options": [
                                        {"label": "A", "value": "a"},
                                        {"label": "B", "value": "b"},
                                    ],
                                }
                            ]
                        }
                    ],
                },
            ],
        }

        errors = TabSchemaValidator.validate(schema)
        assert len(errors) == 0

    def test_validate_wrong_version(self):
        """Test validate detects wrong version."""
        schema = {
            "version": "2.0.0",
            "tabs": [],
        }

        errors = TabSchemaValidator.validate(schema)
        assert "Schema version must be 3.0.0" in errors

    def test_validate_wrong_tab_count(self):
        """Test validate detects wrong tab count."""
        schema = {
            "version": "3.0.0",
            "tabs": [
                {"id": "tab_1", "batches": []},
                {"id": "tab_2", "batches": []},
            ],
        }

        errors = TabSchemaValidator.validate(schema)
        assert "Must have exactly 3 tabs" in errors[0]

    def test_validate_too_many_questions_in_batch(self):
        """Test _validate_batch detects too many questions."""
        # Create 5 questions
        questions = [
            {
                "question": f"Q{i}?",
                "header": "H",
                "options": [{"label": "A", "value": "a"}, {"label": "B", "value": "b"}],
            }
            for i in range(5)
        ]
        batch = {"questions": questions}

        errors = TabSchemaValidator._validate_batch(batch)
        assert len(errors) > 0
        assert "max is 4" in errors[0]

    def test_validate_question_header_too_long(self):
        """Test _validate_question detects header > 12 chars (line 1069)."""
        question = {
            "header": "VeryLongHeader",
            "question": "Test?",
            "options": [{"label": "A", "value": "a"}, {"label": "B", "value": "b"}],
        }

        errors = TabSchemaValidator._validate_question(question)
        assert "exceeds 12 chars" in errors[0]

    def test_validate_question_contains_emoji(self):
        """Test _validate_question detects emoji (lines 1074)."""
        question = {
            "header": "Test",
            "question": "What is your name? 🚀",
            "options": [{"label": "A", "value": "a"}, {"label": "B", "value": "b"}],
        }

        errors = TabSchemaValidator._validate_question(question)
        assert "contains emoji" in errors[0]

    def test_validate_question_wrong_options_count(self):
        """Test _validate_question detects wrong options count (line 1079)."""
        # Only 1 option (min is 2)
        question = {
            "header": "Test",
            "question": "Test?",
            "options": [{"label": "A", "value": "a"}],
        }

        errors = TabSchemaValidator._validate_question(question)
        assert "must be 2-4" in errors[0]

    def test_validate_tab_propagates_batch_errors(self):
        """Test _validate_tab propagates batch errors with context (lines 1039-1040)."""
        tab = {
            "batches": [
                {
                    "questions": [
                        {
                            "header": "Test",
                            "question": "Question with emoji 🎉",
                            "options": [{"label": "A", "value": "a"}],
                        }
                    ]
                }
            ]
        }

        errors = TabSchemaValidator._validate_tab(tab, 0)
        assert len(errors) == 2  # Emoji error + options error
        assert "Tab 0, Batch 0" in errors[0]

    def test_has_emoji_detects_emoji_ranges(self):
        """Test _has_emoji detects Unicode emoji ranges."""
        # Common emoji range
        assert TabSchemaValidator._has_emoji("Test 🚀") is True
        # Misc symbols range
        assert TabSchemaValidator._has_emoji("Test ☀️") is True
        # No emoji
        assert TabSchemaValidator._has_emoji("Test text") is False


class TestConfigurationCoverageValidatorCoverage:
    """Test ConfigurationCoverageValidator for 100% coverage."""

    def test_validate_returns_coverage_breakdown(self):
        """Test validate returns complete coverage breakdown."""
        validator = ConfigurationCoverageValidator()
        coverage = validator.validate()

        assert "user_input" in coverage
        assert "auto_detect" in coverage
        assert "smart_defaults" in coverage
        assert coverage["total_coverage"] == 31

    def test_validate_user_input_fields(self):
        """Test validate includes all 10 user input fields."""
        validator = ConfigurationCoverageValidator()
        coverage = validator.validate()

        assert len(coverage["user_input"]) == 10
        assert "user.name" in coverage["user_input"]
        assert "language.conversation_language" in coverage["user_input"]

    def test_validate_auto_detect_fields(self):
        """Test validate includes all 5 auto-detect fields."""
        validator = ConfigurationCoverageValidator()
        coverage = coverage = validator.validate()

        assert len(coverage["auto_detect"]) == 5
        assert "project.language" in coverage["auto_detect"]
        assert "moai.version" in coverage["auto_detect"]

    def test_validate_required_settings_all_covered(self):
        """Test validate_required_settings with all covered."""
        validator = ConfigurationCoverageValidator()
        required = ["user.name", "project.name", "language.conversation_language"]

        result = validator.validate_required_settings(required)

        assert result["total_covered"] == 3
        assert len(result["missing_settings"]) == 0

    def test_validate_required_settings_missing_some(self):
        """Test validate_required_settings with missing settings."""
        validator = ConfigurationCoverageValidator()
        required = ["user.name", "nonexistent.field"]

        result = validator.validate_required_settings(required)

        assert result["total_covered"] == 1
        assert "nonexistent.field" in result["missing_settings"]
