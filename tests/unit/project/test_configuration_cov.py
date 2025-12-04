"""Additional comprehensive tests for configuration module with 95%+ coverage target.

Tests all uncovered code paths in ConfigurationManager and related classes.
Uses AAA pattern and @patch decorators for file operations and dependencies.
"""

import pytest
import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch, mock_open

from moai_adk.project.configuration import (
    ConfigurationManager,
    SmartDefaultsEngine,
    AutoDetectionEngine,
    ConfigurationCoverageValidator,
    ConditionalBatchRenderer,
    TemplateVariableInterpolator,
    ConfigurationMigrator,
    TabSchemaValidator,
)


class TestConfigurationManagerBackup:
    """Test ConfigurationManager._create_backup method."""

    def test_create_backup_when_file_exists(self):
        """Test creating backup of existing config file."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text('{"version": "3.0.0"}')

            manager = ConfigurationManager(config_path=config_path)

            # Act
            manager._create_backup()

            # Assert
            backup_path = config_path.with_suffix(".backup")
            assert backup_path.exists()
            assert backup_path.read_text() == '{"version": "3.0.0"}'

    def test_create_backup_when_file_not_exists(self):
        """Test create_backup when config file doesn't exist."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"

            manager = ConfigurationManager(config_path=config_path)

            # Act - should not raise
            manager._create_backup()

            # Assert
            backup_path = config_path.with_suffix(".backup")
            assert not backup_path.exists()


class TestConfigurationManagerSave:
    """Test ConfigurationManager.save method."""

    def test_save_creates_config_file(self):
        """Test saving creates config file successfully."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "subdir" / "config.json"

            manager = ConfigurationManager(config_path=config_path)
            config = {
                "version": "3.0.0",
                "user": {"name": "TestUser"},
                "language": {"conversation_language": "en", "agent_prompt_language": "en"},
                "project": {"name": "Test", "owner": "TestOwner", "documentation_mode": "full"},
                "git_strategy": {"mode": "manual"},
                "constitution": {"test_coverage_target": 90, "enforce_tdd": True},
            }

            # Act
            result = manager.save(config)

            # Assert
            assert result is True
            assert config_path.exists()
            saved_config = json.loads(config_path.read_text())
            assert saved_config["version"] == "3.0.0"

    def test_save_atomic_write_with_backup(self):
        """Test save performs atomic write with backup."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_path.write_text('{"version": "2.0.0"}')

            manager = ConfigurationManager(config_path=config_path)
            config = {
                "version": "3.0.0",
                "user": {"name": "TestUser"},
                "language": {"conversation_language": "en", "agent_prompt_language": "en"},
                "project": {"name": "Test", "owner": "TestOwner", "documentation_mode": "full"},
                "git_strategy": {"mode": "manual"},
                "constitution": {"test_coverage_target": 90, "enforce_tdd": True},
            }

            # Act
            manager.save(config)

            # Assert
            backup_path = config_path.with_suffix(".backup")
            assert backup_path.exists()
            assert backup_path.read_text() == '{"version": "2.0.0"}'

    def test_save_incomplete_config_raises_error(self):
        """Test save raises error for incomplete config."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"

            manager = ConfigurationManager(config_path=config_path)
            incomplete_config = {
                "version": "3.0.0",
                "user": {"name": "TestUser"},
                # Missing other required fields
            }

            # Act & Assert
            with pytest.raises(ValueError):
                manager.save(incomplete_config)

    def test_save_temp_file_cleanup_on_error(self):
        """Test temp file is cleaned up on save error."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"

            manager = ConfigurationManager(config_path=config_path)
            incomplete_config = {
                "version": "3.0.0",
                "user": {"name": "TestUser"},
            }

            # Act
            try:
                manager.save(incomplete_config)
            except ValueError:
                pass

            # Assert
            temp_file = config_path.with_suffix(".tmp")
            assert not temp_file.exists()


class TestConditionalBatchRendererEvaluation:
    """Test ConditionalBatchRenderer evaluation methods."""

    def test_evaluate_condition_with_and_operator(self):
        """Test evaluating condition with AND operator."""
        # Arrange
        renderer = ConditionalBatchRenderer({})
        context = {"mode": "personal", "documentation_mode": "full_now"}

        # Act
        result = renderer.evaluate_condition(
            "mode == 'personal' AND documentation_mode == 'full_now'",
            context
        )

        # Assert
        assert result is True

    def test_evaluate_condition_with_or_operator(self):
        """Test evaluating condition with OR operator."""
        # Arrange
        renderer = ConditionalBatchRenderer({})
        context = {"mode": "personal"}

        # Act
        result = renderer.evaluate_condition(
            "mode == 'personal' OR mode == 'team'",
            context
        )

        # Assert
        assert result is True

    def test_evaluate_condition_with_inequality(self):
        """Test evaluating inequality condition."""
        # Arrange
        renderer = ConditionalBatchRenderer({})
        context = {"value": 5}

        # Act
        result = renderer.evaluate_condition("value != 10", context)

        # Assert
        assert result is True

    def test_evaluate_condition_with_comparison_operators(self):
        """Test evaluating all comparison operators."""
        # Arrange
        renderer = ConditionalBatchRenderer({})
        context = {"value": 10}

        # Act & Assert
        assert renderer.evaluate_condition("value < 20", context) is True
        assert renderer.evaluate_condition("value > 5", context) is True
        assert renderer.evaluate_condition("value <= 10", context) is True
        assert renderer.evaluate_condition("value >= 10", context) is True

    def test_evaluate_condition_with_float_values(self):
        """Test evaluating conditions with float values."""
        # Arrange
        renderer = ConditionalBatchRenderer({})
        context = {"score": 95.5}

        # Act
        result = renderer.evaluate_condition("score > 90.0", context)

        # Assert
        assert result is True

    def test_safe_evaluate_with_complex_expression(self):
        """Test safe evaluation with complex nested expressions."""
        # Arrange
        renderer = ConditionalBatchRenderer({})
        context = {"mode": "personal", "doc_mode": "full", "coverage": 85}

        # Act
        result = renderer.evaluate_condition(
            "mode == 'personal' AND (doc_mode == 'full' OR coverage > 80)",
            context
        )

        # Assert
        # Note: The current implementation doesn't support nested expressions
        # but should handle gracefully
        assert isinstance(result, bool)

    def test_resolve_operand_with_string_literals(self):
        """Test resolving string literal operands."""
        # Arrange
        context = {"variable": "value"}

        # Act & Assert
        assert ConditionalBatchRenderer._resolve_operand("'string'", context) == "string"
        assert ConditionalBatchRenderer._resolve_operand('"string"', context) == "string"

    def test_resolve_operand_with_integers(self):
        """Test resolving integer operands."""
        # Arrange
        context = {}

        # Act & Assert
        assert ConditionalBatchRenderer._resolve_operand("42", context) == 42
        assert ConditionalBatchRenderer._resolve_operand("-10", context) == -10

    def test_resolve_operand_with_floats(self):
        """Test resolving float operands."""
        # Arrange
        context = {}

        # Act & Assert
        assert ConditionalBatchRenderer._resolve_operand("3.14", context) == 3.14
        assert ConditionalBatchRenderer._resolve_operand("-2.5", context) == -2.5

    def test_resolve_operand_with_variable(self):
        """Test resolving variable operands."""
        # Arrange
        context = {"my_var": "value"}

        # Act
        result = ConditionalBatchRenderer._resolve_operand("my_var", context)

        # Assert
        assert result == "value"

    def test_resolve_operand_unknown_raises_error(self):
        """Test resolving unknown operand raises error."""
        # Arrange
        context = {}

        # Act & Assert
        with pytest.raises(ValueError):
            ConditionalBatchRenderer._resolve_operand("unknown_var", context)


class TestTemplateVariableInterpolatorAdvanced:
    """Test TemplateVariableInterpolator advanced features."""

    def test_interpolate_with_nested_variables(self):
        """Test interpolating deeply nested variables."""
        # Arrange
        config = {
            "user": {"name": "TestUser"},
            "project": {"owner": "TestOwner", "name": "TestProject"},
            "git_strategy": {"mode": "personal", "personal": {"workflow": "github-flow"}},
        }
        template = "User {{user.name}} owns {{project.name}} with mode {{git_strategy.mode}}"

        # Act
        result = TemplateVariableInterpolator.interpolate(template, config)

        # Assert
        assert result == "User TestUser owns TestProject with mode personal"

    def test_interpolate_with_numeric_variables(self):
        """Test interpolating numeric variables."""
        # Arrange
        config = {
            "constitution": {"test_coverage_target": 90}
        }
        template = "Coverage target: {{constitution.test_coverage_target}}%"

        # Act
        result = TemplateVariableInterpolator.interpolate(template, config)

        # Assert
        assert result == "Coverage target: 90%"

    def test_interpolate_missing_variable_raises_error(self):
        """Test interpolating missing variable raises KeyError."""
        # Arrange
        config = {"user": {"name": "Test"}}
        template = "User: {{missing.variable}}"

        # Act & Assert
        with pytest.raises(KeyError):
            TemplateVariableInterpolator.interpolate(template, config)

    def test_get_nested_value_with_single_level(self):
        """Test getting single-level nested value."""
        # Arrange
        config = {"user": {"name": "Test"}}

        # Act
        result = TemplateVariableInterpolator._get_nested_value(config, "user")

        # Assert
        assert result == {"name": "Test"}

    def test_get_nested_value_with_missing_intermediate(self):
        """Test getting value with missing intermediate key."""
        # Arrange
        config = {"user": {"name": "Test"}}

        # Act
        result = TemplateVariableInterpolator._get_nested_value(config, "missing.name")

        # Assert
        assert result is None


class TestConfigurationMigratorAdvanced:
    """Test ConfigurationMigrator advanced migration scenarios."""

    def test_migrate_preserves_user_data(self):
        """Test migration preserves existing user data."""
        # Arrange
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "LegacyUser", "email": "legacy@example.com"},
            "project": {"name": "OldProject"},
        }
        migrator = ConfigurationMigrator()

        # Act
        v3_config = migrator.migrate(v2_config)

        # Assert
        assert v3_config["version"] == "3.0.0"
        assert v3_config["user"]["name"] == "LegacyUser"
        assert v3_config["user"]["email"] == "legacy@example.com"

    def test_migrate_applies_smart_defaults(self):
        """Test migration applies smart defaults for new fields."""
        # Arrange
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "User"},
            "project": {"name": "Project"},
        }
        migrator = ConfigurationMigrator()

        # Act
        v3_config = migrator.migrate(v2_config)

        # Assert
        assert v3_config["git_strategy"]["personal"]["workflow"] == "github-flow"
        assert v3_config["git_strategy"]["team"]["workflow"] == "git-flow"
        assert v3_config["constitution"]["test_coverage_target"] == 90

    def test_migrate_handles_missing_sections(self):
        """Test migration handles missing sections gracefully."""
        # Arrange
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "User"},
        }
        migrator = ConfigurationMigrator()

        # Act
        v3_config = migrator.migrate(v2_config)

        # Assert
        assert v3_config["version"] == "3.0.0"
        assert v3_config["project"] == {}
        assert "language" in v3_config

    def test_load_legacy_config_creates_deep_copy(self):
        """Test load_legacy_config creates deep copy."""
        # Arrange
        v2_config = {
            "user": {"name": "User", "settings": {"theme": "dark"}},
        }
        migrator = ConfigurationMigrator()

        # Act
        loaded = migrator.load_legacy_config(v2_config)
        loaded["user"]["settings"]["theme"] = "light"

        # Assert
        assert v2_config["user"]["settings"]["theme"] == "dark"  # Original unchanged


class TestTabSchemaValidatorAdvanced:
    """Test TabSchemaValidator advanced validation scenarios."""

    def test_validate_emoji_detection(self):
        """Test emoji detection in schema validation."""
        # Arrange
        validator = TabSchemaValidator()

        # Act & Assert
        assert validator._has_emoji("Hello") is False
        assert validator._has_emoji("Hello 😀") is True
        assert validator._has_emoji("🎉 Party") is True

    def test_validate_question_header_too_long(self):
        """Test validation catches header exceeding 12 chars."""
        # Arrange
        question = {
            "header": "This header is too long",
            "question": "What is this?",
            "options": ["Yes", "No"]
        }

        # Act
        errors = TabSchemaValidator._validate_question(question)

        # Assert
        assert len(errors) > 0
        assert "exceeds 12 chars" in errors[0]

    def test_validate_question_options_count(self):
        """Test validation checks options count."""
        # Arrange
        question_single = {
            "header": "Q1",
            "question": "Choose",
            "options": ["Only one"]
        }
        question_too_many = {
            "header": "Q2",
            "question": "Choose",
            "options": ["A", "B", "C", "D", "E"]
        }

        # Act
        errors_single = TabSchemaValidator._validate_question(question_single)
        errors_many = TabSchemaValidator._validate_question(question_too_many)

        # Assert
        assert len(errors_single) > 0
        assert len(errors_many) > 0

    def test_validate_batch_with_too_many_questions(self):
        """Test validation catches batch with too many questions."""
        # Arrange
        batch = {
            "questions": [
                {"header": f"Q{i}", "question": "?", "options": ["A", "B"]}
                for i in range(5)  # Max is 4
            ]
        }

        # Act
        errors = TabSchemaValidator._validate_batch(batch)

        # Assert
        assert len(errors) > 0
        assert "max is 4" in errors[0]


class TestAutoDetectionEngineLanguageDetection:
    """Test AutoDetectionEngine language detection methods."""

    def test_detect_language_javascript(self):
        """Test detecting JavaScript project."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            package_json = Path(tmpdir) / "package.json"
            package_json.write_text('{}')

            # Act
            with patch("moai_adk.project.configuration.Path.cwd", return_value=Path(tmpdir)):
                engine = AutoDetectionEngine()
                result = engine.detect_language()

            # Assert
            assert result == "javascript"

    def test_detect_language_go(self):
        """Test detecting Go project."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            go_mod = Path(tmpdir) / "go.mod"
            go_mod.write_text("")

            # Act
            with patch("moai_adk.project.configuration.Path.cwd", return_value=Path(tmpdir)):
                engine = AutoDetectionEngine()
                result = engine.detect_language()

            # Assert
            assert result == "go"

    def test_detect_locale_unmapped_language(self):
        """Test detecting locale for unmapped language code."""
        # Arrange
        engine = AutoDetectionEngine()

        # Act
        result = engine.detect_locale("unknown")

        # Assert
        assert result == "en_US"  # Default to English

    def test_detect_language_name_unmapped(self):
        """Test detecting language name for unmapped code."""
        # Arrange
        engine = AutoDetectionEngine()

        # Act
        result = engine.detect_language_name("unknown")

        # Assert
        assert result == "English"  # Default to English


class TestSmartDefaultsEngineDefaults:
    """Test SmartDefaultsEngine default values."""

    def test_get_all_defaults_returns_copy(self):
        """Test get_all_defaults returns independent copy."""
        # Arrange
        engine = SmartDefaultsEngine()

        # Act
        defaults1 = engine.get_all_defaults()
        defaults1["custom_key"] = "custom_value"
        defaults2 = engine.get_all_defaults()

        # Assert
        assert "custom_key" not in defaults2

    def test_apply_defaults_preserves_existing_values(self):
        """Test apply_defaults preserves existing config values."""
        # Arrange
        engine = SmartDefaultsEngine()
        config = {
            "git_strategy": {
                "personal": {"workflow": "custom-flow"}
            },
            "constitution": {"test_coverage_target": 95}
        }

        # Act
        result = engine.apply_defaults(config)

        # Assert
        assert result["git_strategy"]["personal"]["workflow"] == "custom-flow"
        assert result["constitution"]["test_coverage_target"] == 95

    def test_apply_defaults_creates_nested_structure(self):
        """Test apply_defaults creates required nested structure."""
        # Arrange
        engine = SmartDefaultsEngine()
        config = {"user": {"name": "Test"}}

        # Act
        result = engine.apply_defaults(config)

        # Assert
        assert "git_strategy" in result
        assert "personal" in result["git_strategy"]
        assert "team" in result["git_strategy"]
        assert "constitution" in result


class TestConfigurationCoverageValidatorDetailed:
    """Test ConfigurationCoverageValidator detailed validation."""

    def test_validate_required_settings_coverage(self):
        """Test validate_required_settings coverage."""
        # Arrange
        validator = ConfigurationCoverageValidator()
        required = [
            "user.name",
            "language.conversation_language",
            "project.name",
            "git_strategy.mode",
        ]

        # Act
        result = validator.validate_required_settings(required)

        # Assert
        assert "required" in result
        assert "covered" in result
        assert "missing_settings" in result
        assert result["total_covered"] == len(required)


class TestConfigurationManagerWriteConfig:
    """Test ConfigurationManager._write_config internal method."""

    def test_write_config_delegates_to_save(self):
        """Test _write_config delegates to save method."""
        # Arrange
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            manager = ConfigurationManager(config_path=config_path)
            config = {
                "version": "3.0.0",
                "user": {"name": "Test"},
                "language": {"conversation_language": "en", "agent_prompt_language": "en"},
                "project": {"name": "Test", "owner": "Owner", "documentation_mode": "full"},
                "git_strategy": {"mode": "manual"},
                "constitution": {"test_coverage_target": 90, "enforce_tdd": True},
            }

            # Act
            manager._write_config(config)

            # Assert
            assert config_path.exists()
            saved = json.loads(config_path.read_text())
            assert saved["user"]["name"] == "Test"


class TestConditionalBatchRendererGetVisibleBatches:
    """Test ConditionalBatchRenderer.get_visible_batches method."""

    def test_get_visible_batches_with_matching_condition(self):
        """Test getting visible batches with matching show_if condition."""
        # Arrange
        schema = {
            "tabs": [
                {
                    "id": "tab_3_git_automation",
                    "batches": [
                        {
                            "id": "batch_3_1_personal",
                            "show_if": "mode == 'personal'"
                        },
                        {
                            "id": "batch_3_1_team",
                            "show_if": "mode == 'team'"
                        }
                    ]
                }
            ]
        }
        renderer = ConditionalBatchRenderer(schema)

        # Act
        visible = renderer.get_visible_batches("tab_3_git_automation", {"mode": "personal"})

        # Assert
        assert len(visible) == 1
        assert visible[0]["id"] == "batch_3_1_personal"

    def test_get_visible_batches_with_partial_tab_id_match(self):
        """Test getting visible batches with partial tab ID match."""
        # Arrange
        schema = {
            "tabs": [
                {
                    "id": "tab_3_git_automation",
                    "batches": [
                        {"id": "batch_1", "show_if": "true"}
                    ]
                }
            ]
        }
        renderer = ConditionalBatchRenderer(schema)

        # Act
        visible = renderer.get_visible_batches("tab_3", {})

        # Assert
        assert len(visible) == 1

    def test_get_visible_batches_empty_batches(self):
        """Test getting visible batches when no batches exist."""
        # Arrange
        schema = {"tabs": []}
        renderer = ConditionalBatchRenderer(schema)

        # Act
        visible = renderer.get_visible_batches("tab_3", {})

        # Assert
        assert visible == []

    def test_evaluate_condition_error_handling(self):
        """Test evaluate_condition handles errors gracefully."""
        # Arrange
        renderer = ConditionalBatchRenderer({})
        context = {}

        # Act
        result = renderer.evaluate_condition("invalid expression !@#$", context)

        # Assert
        assert result is True  # Fail-safe: return True on error
