"""Tests for moai_adk.project.configuration module."""

import json
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch, mock_open

import pytest

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


class TestConfigurationManager:
    """Test ConfigurationManager class."""

    def test_init_with_default_path(self):
        """Test ConfigurationManager initialization with default path."""
        manager = ConfigurationManager()
        assert manager.config_path == Path(".moai/config/config.yaml")
        assert manager.schema is None
        assert manager._config_cache is None

    def test_init_with_custom_path(self):
        """Test ConfigurationManager initialization with custom path."""
        custom_path = Path("/custom/config.json")
        manager = ConfigurationManager(custom_path)
        assert manager.config_path == custom_path

    def test_load_existing_file(self):
        """Test load method with existing config file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            config_path = Path(tmpdir) / "config.json"
            config_data = {"user": {"name": "TestUser"}}
            config_path.write_text(json.dumps(config_data))

            manager = ConfigurationManager(config_path)
            result = manager.load()
            assert result == config_data
            assert manager._config_cache == config_data

    def test_load_nonexistent_file(self):
        """Test load method with nonexistent config file."""
        manager = ConfigurationManager(Path("/nonexistent/config.json"))
        result = manager.load()
        assert result == {}

    def test_get_smart_defaults(self):
        """Test get_smart_defaults method."""
        manager = ConfigurationManager()
        defaults = manager.get_smart_defaults()
        assert isinstance(defaults, dict)
        assert "git_strategy.personal.workflow" in defaults
        assert "constitution.test_coverage_target" in defaults

    def test_get_auto_detect_fields(self):
        """Test get_auto_detect_fields method."""
        manager = ConfigurationManager()
        fields = manager.get_auto_detect_fields()
        assert isinstance(fields, list)
        assert len(fields) > 0
        assert any(f["field"] == "project.language" for f in fields)

    def test_parse_responses(self):
        """Test _parse_responses method."""
        manager = ConfigurationManager()
        responses = {
            "user_name": "GOOS",
            "conversation_language": "en",
            "project_name": "MoAI",
        }
        result = manager._parse_responses(responses)
        assert result["user"]["name"] == "GOOS"
        assert result["language"]["conversation_language"] == "en"
        assert result["project"]["name"] == "MoAI"

    def test_set_nested(self):
        """Test _set_nested method."""
        config = {}
        ConfigurationManager._set_nested(config, ("a", "b", "c"), "value")
        assert config["a"]["b"]["c"] == "value"

    def test_flatten_config(self):
        """Test _flatten_config method."""
        config = {
            "user": {"name": "Test"},
            "project": {"name": "Project1"},
        }
        flat = ConfigurationManager._flatten_config(config)
        assert flat["user.name"] == "Test"
        assert flat["project.name"] == "Project1"

    def test_validate_complete_with_required_fields(self):
        """Test _validate_complete with all required fields."""
        manager = ConfigurationManager()
        config = {
            "user": {"name": "Test"},
            "language": {"conversation_language": "en", "agent_prompt_language": "en"},
            "project": {"name": "P1", "owner": "Owner", "documentation_mode": "full"},
            "git_strategy": {"mode": "personal"},
            "constitution": {"test_coverage_target": 90, "enforce_tdd": True},
        }
        result = manager._validate_complete(config)
        assert result is True

    def test_validate_complete_missing_fields(self):
        """Test _validate_complete with missing fields."""
        manager = ConfigurationManager()
        config = {"user": {"name": "Test"}}
        result = manager._validate_complete(config)
        assert result is False

    def test_build_from_responses(self):
        """Test build_from_responses method."""
        manager = ConfigurationManager()
        responses = {
            "user_name": "GOOS",
            "conversation_language": "en",
            "project_name": "MoAI",
        }
        result = manager.build_from_responses(responses)
        assert result["user"]["name"] == "GOOS"


class TestSmartDefaultsEngine:
    """Test SmartDefaultsEngine class."""

    def test_init(self):
        """Test SmartDefaultsEngine initialization."""
        engine = SmartDefaultsEngine()
        assert engine.defaults is not None
        assert len(engine.defaults) >= 16

    def test_get_all_defaults(self):
        """Test get_all_defaults method."""
        engine = SmartDefaultsEngine()
        defaults = engine.get_all_defaults()
        assert isinstance(defaults, dict)
        assert "git_strategy.personal.workflow" in defaults
        assert defaults["git_strategy.personal.workflow"] == "github-flow"

    def test_get_default_existing_key(self):
        """Test get_default with existing key."""
        engine = SmartDefaultsEngine()
        default = engine.get_default("constitution.test_coverage_target")
        assert default == 85

    def test_get_default_nonexistent_key(self):
        """Test get_default with nonexistent key."""
        engine = SmartDefaultsEngine()
        default = engine.get_default("nonexistent.key")
        assert default is None

    def test_apply_defaults(self):
        """Test apply_defaults method."""
        engine = SmartDefaultsEngine()
        partial_config = {"user": {"name": "Test"}}
        complete_config = engine.apply_defaults(partial_config)
        assert complete_config["user"]["name"] == "Test"
        assert complete_config["git_strategy"]["personal"]["workflow"] == "github-flow"
        assert complete_config["constitution"]["test_coverage_target"] == 85


class TestAutoDetectionEngine:
    """Test AutoDetectionEngine class."""

    def test_detect_language_python(self):
        """Test detect_language with Python project."""
        with patch("pathlib.Path.cwd") as mock_cwd:
            mock_path = MagicMock()
            mock_cwd.return_value = mock_path

            mock_path.__truediv__ = lambda self, x: MagicMock(exists=lambda: x == "pyproject.toml")
            result = AutoDetectionEngine.detect_language()
            assert result == "python"

    def test_detect_language_typescript(self):
        """Test detect_language with TypeScript project."""
        with patch("pathlib.Path.cwd") as mock_cwd:
            mock_path = MagicMock()
            mock_cwd.return_value = mock_path

            mock_path.__truediv__ = lambda self, x: MagicMock(exists=lambda: x == "tsconfig.json")
            result = AutoDetectionEngine.detect_language()
            assert result == "typescript"

    def test_detect_language_default(self):
        """Test detect_language default to Python."""
        with patch("pathlib.Path.cwd") as mock_cwd:
            mock_path = MagicMock()
            mock_cwd.return_value = mock_path
            mock_path.__truediv__ = lambda self, x: MagicMock(exists=lambda: False)
            result = AutoDetectionEngine.detect_language()
            assert result == "python"

    def test_detect_locale_korean(self):
        """Test detect_locale with Korean language."""
        result = AutoDetectionEngine.detect_locale("ko")
        assert result == "ko_KR"

    def test_detect_locale_english(self):
        """Test detect_locale with English language."""
        result = AutoDetectionEngine.detect_locale("en")
        assert result == "en_US"

    def test_detect_language_name_korean(self):
        """Test detect_language_name with Korean code."""
        result = AutoDetectionEngine.detect_language_name("ko")
        assert result == "Korean"

    def test_detect_language_name_english(self):
        """Test detect_language_name with English code."""
        result = AutoDetectionEngine.detect_language_name("en")
        assert result == "English"

    def test_detect_template_version(self):
        """Test detect_template_version method."""
        result = AutoDetectionEngine.detect_template_version()
        assert result == "3.0.0"

    def test_detect_moai_version(self):
        """Test detect_moai_version method."""
        result = AutoDetectionEngine.detect_moai_version()
        assert isinstance(result, str)

    def test_detect_and_apply(self):
        """Test detect_and_apply method."""
        engine = AutoDetectionEngine()
        config = {"language": {"conversation_language": "en"}}
        result = engine.detect_and_apply(config)
        assert "project" in result
        assert "language" in result


class TestConfigurationCoverageValidator:
    """Test ConfigurationCoverageValidator class."""

    def test_init(self):
        """Test ConfigurationCoverageValidator initialization."""
        validator = ConfigurationCoverageValidator()
        assert validator.schema is None

    def test_validate(self):
        """Test validate method."""
        validator = ConfigurationCoverageValidator()
        result = validator.validate()
        assert isinstance(result, dict)
        assert result["total_coverage"] == 31
        assert len(result["user_input"]) == 10
        assert len(result["auto_detect"]) == 5

    def test_validate_required_settings(self):
        """Test validate_required_settings method."""
        validator = ConfigurationCoverageValidator()
        required = ["user.name", "language.conversation_language"]
        result = validator.validate_required_settings(required)
        assert "required" in result
        assert "covered" in result
        assert "missing_settings" in result


class TestConditionalBatchRenderer:
    """Test ConditionalBatchRenderer class."""

    def test_init(self):
        """Test ConditionalBatchRenderer initialization."""
        schema = {"tabs": []}
        renderer = ConditionalBatchRenderer(schema)
        assert renderer.schema == schema

    def test_get_visible_batches_with_true_condition(self):
        """Test get_visible_batches with true condition."""
        schema = {"tabs": [{"id": "tab_1", "batches": [{"id": "batch_1", "show_if": "true"}]}]}
        renderer = ConditionalBatchRenderer(schema)
        result = renderer.get_visible_batches("tab_1", {})
        assert len(result) == 1

    def test_evaluate_condition_true(self):
        """Test evaluate_condition with true."""
        renderer = ConditionalBatchRenderer({})
        result = renderer.evaluate_condition("true", {})
        assert result is True

    def test_evaluate_condition_equality(self):
        """Test evaluate_condition with equality."""
        renderer = ConditionalBatchRenderer({})
        result = renderer.evaluate_condition("mode == 'personal'", {"mode": "personal"})
        assert result is True

    def test_safe_evaluate_and_operator(self):
        """Test _safe_evaluate with AND operator."""
        renderer = ConditionalBatchRenderer({})
        result = renderer._safe_evaluate(
            "mode == 'personal' AND documentation_mode == 'full'",
            {"mode": "personal", "documentation_mode": "full"},
        )
        assert result is True

    def test_safe_evaluate_or_operator(self):
        """Test _safe_evaluate with OR operator."""
        renderer = ConditionalBatchRenderer({})
        result = renderer._safe_evaluate("mode == 'personal' OR mode == 'team'", {"mode": "personal"})
        assert result is True

    def test_resolve_operand_string_literal(self):
        """Test _resolve_operand with string literal."""
        renderer = ConditionalBatchRenderer({})
        result = renderer._resolve_operand("'value'", {})
        assert result == "value"

    def test_resolve_operand_number(self):
        """Test _resolve_operand with number."""
        renderer = ConditionalBatchRenderer({})
        result = renderer._resolve_operand("123", {})
        assert result == 123

    def test_resolve_operand_variable(self):
        """Test _resolve_operand with variable."""
        renderer = ConditionalBatchRenderer({})
        result = renderer._resolve_operand("mode", {"mode": "personal"})
        assert result == "personal"


class TestTemplateVariableInterpolator:
    """Test TemplateVariableInterpolator class."""

    def test_interpolate_simple_variable(self):
        """Test interpolate with simple variable."""
        template = "User: {{user.name}}"
        config = {"user": {"name": "GOOS"}}
        result = TemplateVariableInterpolator.interpolate(template, config)
        assert result == "User: GOOS"

    def test_interpolate_multiple_variables(self):
        """Test interpolate with multiple variables."""
        template = "User: {{user.name}}, Project: {{project.name}}"
        config = {"user": {"name": "GOOS"}, "project": {"name": "MoAI"}}
        result = TemplateVariableInterpolator.interpolate(template, config)
        assert result == "User: GOOS, Project: MoAI"

    def test_interpolate_missing_variable(self):
        """Test interpolate with missing variable."""
        template = "User: {{user.name}}"
        config = {}
        with pytest.raises(KeyError):
            TemplateVariableInterpolator.interpolate(template, config)

    def test_get_nested_value(self):
        """Test _get_nested_value method."""
        config = {"user": {"profile": {"name": "Test"}}}
        result = TemplateVariableInterpolator._get_nested_value(config, "user.profile.name")
        assert result == "Test"

    def test_get_nested_value_missing(self):
        """Test _get_nested_value with missing path."""
        config = {"user": {"name": "Test"}}
        result = TemplateVariableInterpolator._get_nested_value(config, "user.missing")
        assert result is None


class TestConfigurationMigrator:
    """Test ConfigurationMigrator class."""

    def test_load_legacy_config(self):
        """Test load_legacy_config method."""
        migrator = ConfigurationMigrator()
        v2_config = {"version": "2.1.0", "user": {"name": "Test"}}
        result = migrator.load_legacy_config(v2_config)
        assert result == v2_config
        # Verify it's a deep copy
        result["user"]["name"] = "Modified"
        assert v2_config["user"]["name"] == "Test"

    def test_migrate_v2_to_v3(self):
        """Test migrate method."""
        migrator = ConfigurationMigrator()
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "OldUser"},
            "git_strategy": {"mode": "personal"},
        }
        result = migrator.migrate(v2_config)
        assert result["version"] == "3.0.0"
        assert result["user"]["name"] == "OldUser"
        assert result["git_strategy"]["mode"] == "personal"


class TestTabSchemaValidator:
    """Test TabSchemaValidator class."""

    def test_validate_correct_schema(self):
        """Test validate with correct schema."""
        schema = {
            "version": "3.0.0",
            "tabs": [
                {"id": "tab1", "batches": []},
                {"id": "tab2", "batches": []},
                {"id": "tab3", "batches": []},
            ],
        }
        result = TabSchemaValidator.validate(schema)
        assert len(result) == 0

    def test_validate_wrong_version(self):
        """Test validate with wrong version."""
        schema = {"version": "2.0.0", "tabs": []}
        result = TabSchemaValidator.validate(schema)
        assert len(result) > 0

    def test_validate_wrong_tab_count(self):
        """Test validate with wrong tab count."""
        schema = {
            "version": "3.0.0",
            "tabs": [{"id": "tab1", "batches": []}, {"id": "tab2", "batches": []}],
        }
        result = TabSchemaValidator.validate(schema)
        assert len(result) > 0

    def test_has_emoji(self):
        """Test _has_emoji method."""
        assert TabSchemaValidator._has_emoji("No emoji") is False
        assert TabSchemaValidator._has_emoji("With emoji 😀") is True

    def test_validate_batch_with_too_many_questions(self):
        """Test _validate_batch with too many questions."""
        batch = {
            "questions": [
                {"header": "Q1", "question": "Q1", "options": ["1", "2"]},
                {"header": "Q2", "question": "Q2", "options": ["1", "2"]},
                {"header": "Q3", "question": "Q3", "options": ["1", "2"]},
                {"header": "Q4", "question": "Q4", "options": ["1", "2"]},
                {"header": "Q5", "question": "Q5", "options": ["1", "2"]},
            ]
        }
        errors = TabSchemaValidator._validate_batch(batch)
        assert len(errors) > 0
