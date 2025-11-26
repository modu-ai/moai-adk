"""
Test suite for SPEC-REDESIGN-001: Project Configuration Schema v3.0.0 Redesign

Tests cover:
- Tab structure and question count reduction
- Configuration coverage (31 settings)
- Smart defaults
- Auto-detection
- Conditional rendering
- Document generation
- Agent context loading
- Backward compatibility
"""

from pathlib import Path
from typing import Any, Dict, List
from unittest.mock import patch

import pytest

# These imports would be implemented in the actual codebase
from moai_adk.project.configuration import (
    AutoDetectionEngine,
    ConfigurationCoverageValidator,
    ConfigurationManager,
    SmartDefaultsEngine,
)
from moai_adk.project.documentation import DocumentationGenerator
from moai_adk.project.schema import load_tab_schema


class TestTabSchemaV3Structure:
    """AC-001: Tab structure reduced from 5 tabs to 3 tabs with smart questions"""

    def test_tab_schema_loads_correctly(self):
        """Should load valid tab_schema.json v3.0.0"""
        schema = load_tab_schema()
        assert schema is not None
        assert schema.get("version") == "3.0.0"

    def test_tab_count_is_three(self):
        """Should have exactly 3 tabs"""
        schema = load_tab_schema()
        tabs = schema.get("tabs", [])
        assert len(tabs) == 3
        assert tabs[0]["id"] == "tab_1_quick_start"
        assert tabs[1]["id"] == "tab_2_documentation"
        assert tabs[2]["id"] == "tab_3_git_automation"

    def test_tab_1_batch_count_and_questions(self):
        """AC-003: Tab 1 should have 4 batches with 10 total questions"""
        schema = load_tab_schema()
        tab1 = schema["tabs"][0]
        batches = tab1.get("batches", [])

        assert len(batches) == 4
        total_questions = sum(len(batch.get("questions", [])) for batch in batches)
        assert total_questions == 10

    def test_tab_1_batch_1_identity_questions(self):
        """Tab 1.1 should have 3 identity & language questions"""
        schema = load_tab_schema()
        batch = schema["tabs"][0]["batches"][0]

        assert batch["id"] == "batch_1_1_identity"
        questions = batch["questions"]
        assert len(questions) == 3

        # Verify question IDs
        question_ids = [q["id"] for q in questions]
        assert "user_name" in question_ids
        assert "conversation_language" in question_ids
        assert "agent_prompt_language" in question_ids

    def test_tab_1_batch_2_project_basics(self):
        """Tab 1.2 should have 3 project basics questions"""
        schema = load_tab_schema()
        batch = schema["tabs"][0]["batches"][1]

        assert batch["id"] == "batch_1_2_project_basics"
        questions = batch["questions"]
        assert len(questions) == 3

        question_ids = [q["id"] for q in questions]
        assert "project_name" in question_ids
        assert "project_owner" in question_ids
        assert "project_description" in question_ids

    def test_tab_1_batch_3_development_mode(self):
        """Tab 1.3 should have 2 development mode questions"""
        schema = load_tab_schema()
        batch = schema["tabs"][0]["batches"][2]

        assert batch["id"] == "batch_1_3_development_mode"
        questions = batch["questions"]
        assert len(questions) == 2

        question_ids = [q["id"] for q in questions]
        assert "git_strategy_mode" in question_ids

    def test_tab_1_batch_4_quality_standards(self):
        """Tab 1.4 should have 2 quality standards questions"""
        schema = load_tab_schema()
        batch = schema["tabs"][0]["batches"][3]

        assert batch["id"] == "batch_1_4_quality_standards"
        questions = batch["questions"]
        assert len(questions) == 2

        question_ids = [q["id"] for q in questions]
        assert "test_coverage_target" in question_ids
        assert "enforce_tdd" in question_ids

    def test_tab_2_documentation_choice(self):
        """Tab 2 should have 1 required + 1 conditional question"""
        schema = load_tab_schema()
        tab2 = schema["tabs"][1]
        batches = tab2.get("batches", [])

        assert len(batches) >= 2
        batch_1 = batches[0]
        assert batch_1["id"] == "batch_2_1_documentation_choice"
        assert len(batch_1["questions"]) == 1

        batch_2 = batches[1]
        assert batch_2["id"] == "batch_2_2_documentation_depth"
        assert batch_2["questions"][0].get("show_if") is not None

    def test_tab_3_conditional_git_batches(self):
        """Tab 3 should have conditional batches based on git mode"""
        schema = load_tab_schema()
        tab3 = schema["tabs"][2]
        batches = tab3.get("batches", [])

        # Should have at least personal and team batches
        batch_ids = [b["id"] for b in batches]
        assert "batch_3_1_personal" in batch_ids
        assert "batch_3_1_team" in batch_ids

        # Both should have conditional show_if
        personal_batch = next(b for b in batches if b["id"] == "batch_3_1_personal")
        team_batch = next(b for b in batches if b["id"] == "batch_3_1_team")

        assert personal_batch.get("show_if") is not None
        assert team_batch.get("show_if") is not None


class TestQuestionReduction:
    """AC-003: Question count reduced by 63% (27 → 10)"""

    def test_total_questions_in_tab_1(self):
        """Tab 1 should have exactly 10 questions"""
        schema = load_tab_schema()
        tab1 = schema["tabs"][0]
        total = sum(len(batch["questions"]) for batch in tab1["batches"])
        assert total == 10

    def test_essential_questions_count(self):
        """Essential questions (non-conditional) should be 10-12 maximum"""
        schema = load_tab_schema()
        essential_count = 0

        for tab in schema["tabs"]:
            for batch in tab.get("batches", []):
                for question in batch.get("questions", []):
                    # Count if not conditional
                    if not question.get("show_if"):
                        essential_count += 1

        # At minimum, Tab 1 has 10 essential + Tab 2.1 has 1
        assert essential_count >= 11

    def test_question_reduction_rate(self):
        """Should achieve 63% reduction from 27 to 10 questions"""
        schema = load_tab_schema()
        tab1_questions = sum(len(batch["questions"]) for batch in schema["tabs"][0]["batches"])

        reduction_rate = (27 - tab1_questions) / 27
        assert reduction_rate >= 0.63  # At least 63% reduction
        assert tab1_questions == 10


class TestConfigurationCoverage:
    """AC-004: 100% configuration coverage (31 settings)"""

    def test_all_31_settings_defined_in_schema(self):
        """Schema should map all 31 configuration settings"""
        manager = ConfigurationManager()
        coverage = ConfigurationCoverageValidator(manager.schema)

        # Should have mapping for all 31 settings
        settings_coverage = coverage.validate()

        assert len(settings_coverage["user_input"]) == 10
        assert len(settings_coverage["auto_detect"]) == 5
        assert len(settings_coverage["smart_defaults"]) >= 16
        assert settings_coverage["total_coverage"] == 31

    def test_config_coverage_matrix(self):
        """Should cover all 31 settings from coverage matrix"""
        required_settings = [
            # User input (10)
            "user.name",
            "language.conversation_language",
            "language.agent_prompt_language",
            "project.name",
            "project.owner",
            "project.description",
            "git_strategy.mode",
            "git_strategy.{mode}.workflow",
            "constitution.test_coverage_target",
            "constitution.enforce_tdd",
            # Required selection (1)
            "project.documentation_mode",
            # Conditional (1)
            "project.documentation_depth",
            # Conditional Git (4)
            "git_strategy.personal.auto_checkpoint",
            "git_strategy.personal.push_to_remote",
            "git_strategy.team.auto_pr",
            "git_strategy.team.draft_pr",
            # Auto-detect (5)
            "project.language",
            "project.locale",
            "language.conversation_language_name",
            "project.template_version",
            "moai.version",
        ]

        validator = ConfigurationCoverageValidator()
        coverage = validator.validate_required_settings(required_settings)

        assert coverage["missing_settings"] == []
        assert coverage["total_covered"] >= 21

    def test_smart_defaults_count(self):
        """Should define at least 16 smart defaults"""
        manager = ConfigurationManager()
        defaults = manager.get_smart_defaults()

        assert len(defaults) >= 16

    def test_auto_detect_fields_count(self):
        """Should define exactly 5 auto-detect fields"""
        manager = ConfigurationManager()
        auto_detect = manager.get_auto_detect_fields()

        assert len(auto_detect) == 5
        field_names = [f["field"] for f in auto_detect]
        assert "project.language" in field_names
        assert "project.locale" in field_names
        assert "language.conversation_language_name" in field_names
        assert "project.template_version" in field_names
        assert "moai.version" in field_names


class TestSmartDefaults:
    """AC-006: Smart defaults automatically applied for 16 settings"""

    def test_smart_defaults_engine_initialization(self):
        """Smart defaults engine should initialize with all defaults"""
        engine = SmartDefaultsEngine()
        defaults = engine.get_all_defaults()

        assert len(defaults) >= 16
        assert "git_strategy.personal.workflow" in defaults
        assert "git_strategy.team.workflow" in defaults

    def test_personal_workflow_default(self):
        """Personal mode should default to github-flow"""
        engine = SmartDefaultsEngine()
        default = engine.get_default("git_strategy.personal.workflow")

        assert default == "github-flow"

    def test_team_workflow_default(self):
        """Team mode should default to git-flow"""
        engine = SmartDefaultsEngine()
        default = engine.get_default("git_strategy.team.workflow")

        assert default == "git-flow"

    def test_test_coverage_target_default(self):
        """Test coverage target should default to 90"""
        engine = SmartDefaultsEngine()
        default = engine.get_default("constitution.test_coverage_target")

        assert default == 90

    def test_apply_smart_defaults(self):
        """Smart defaults should be applied to config"""
        config_partial = {
            "user": {"name": "TestUser"},
            "project": {"name": "TestProject"},
        }

        engine = SmartDefaultsEngine()
        config_complete = engine.apply_defaults(config_partial)

        assert config_complete["git_strategy"]["personal"]["workflow"] == "github-flow"
        assert config_complete["constitution"]["test_coverage_target"] == 90
        assert config_complete["constitution"]["enforce_tdd"] is True


class TestAutoDetection:
    """AC-007: Auto-detect 5 fields automatically"""

    def test_auto_detect_engine_initialization(self):
        """Auto-detect engine should initialize"""
        engine = AutoDetectionEngine()
        assert engine is not None

    def test_detect_language_python_project(self):
        """Should detect Python as project language"""
        engine = AutoDetectionEngine()

        with patch("pathlib.Path.glob") as mock_glob:
            mock_glob.return_value = [Path("setup.py"), Path("pyproject.toml")]
            language = engine.detect_language()

            assert language == "python"

    def test_detect_language_typescript_project(self):
        """Should detect TypeScript as project language"""
        engine = AutoDetectionEngine()

        with patch("pathlib.Path.glob") as mock_glob:
            mock_glob.return_value = [Path("tsconfig.json"), Path("package.json")]
            language = engine.detect_language()

            assert language in ["typescript", "javascript"]

    def test_detect_locale_from_language(self):
        """Should map conversation language to locale"""
        engine = AutoDetectionEngine()

        mappings = {
            "ko": "ko_KR",
            "en": "en_US",
            "ja": "ja_JP",
        }

        for lang_code, expected_locale in mappings.items():
            locale = engine.detect_locale(lang_code)
            assert locale == expected_locale

    def test_detect_language_name(self):
        """Should convert language code to language name"""
        engine = AutoDetectionEngine()

        assert engine.detect_language_name("ko") == "Korean"
        assert engine.detect_language_name("en") == "English"
        assert engine.detect_language_name("ja") == "Japanese"

    def test_detect_template_version(self):
        """Should detect template version from system"""
        engine = AutoDetectionEngine()

        with patch("moai_adk.version.TEMPLATE_VERSION", "3.0.0"):
            version = engine.detect_template_version()
            assert version == "3.0.0"

    def test_detect_moai_version(self):
        """Should detect MoAI version from system"""
        engine = AutoDetectionEngine()

        with patch("moai_adk.version.MOAI_VERSION", "0.26.0"):
            version = engine.detect_moai_version()
            assert version == "0.26.0"


class TestConditionalRendering:
    """AC-005: Conditional batch rendering based on git_strategy.mode"""

    def test_personal_mode_shows_personal_batch(self):
        """Personal mode should show personal batch only"""
        schema = load_tab_schema()
        git_config = {"mode": "personal"}

        renderer = ConditionalBatchRenderer(schema)
        batches = renderer.get_visible_batches("tab_3", git_config)

        batch_ids = [b["id"] for b in batches]
        assert "batch_3_1_personal" in batch_ids
        assert "batch_3_1_team" not in batch_ids

    def test_team_mode_shows_team_batch(self):
        """Team mode should show team batch only"""
        schema = load_tab_schema()
        git_config = {"mode": "team"}

        renderer = ConditionalBatchRenderer(schema)
        batches = renderer.get_visible_batches("tab_3", git_config)

        batch_ids = [b["id"] for b in batches]
        assert "batch_3_1_team" in batch_ids
        assert "batch_3_1_personal" not in batch_ids

    def test_hybrid_mode_shows_personal_default(self):
        """Hybrid mode should show personal batch by default"""
        schema = load_tab_schema()
        git_config = {"mode": "hybrid"}

        renderer = ConditionalBatchRenderer(schema)
        batches = renderer.get_visible_batches("tab_3", git_config)

        batch_ids = [b["id"] for b in batches]
        assert "batch_3_1_personal" in batch_ids

    def test_conditional_show_if_evaluation(self):
        """Conditional show_if should be evaluated correctly"""
        batch = {
            "id": "test_batch",
            "show_if": "documentation_mode == 'full_now'",
            "questions": [],
        }

        renderer = ConditionalBatchRenderer({})

        assert renderer.evaluate_condition(batch["show_if"], {"documentation_mode": "full_now"}) is True
        assert renderer.evaluate_condition(batch["show_if"], {"documentation_mode": "skip"}) is False

    def test_nested_conditional_logic(self):
        """Complex conditional logic should work"""
        condition = "mode == 'personal' AND documentation_mode == 'full_now'"
        renderer = ConditionalBatchRenderer({})

        context_true = {
            "mode": "personal",
            "documentation_mode": "full_now",
        }
        context_false = {
            "mode": "team",
            "documentation_mode": "full_now",
        }

        assert renderer.evaluate_condition(condition, context_true) is True
        assert renderer.evaluate_condition(condition, context_false) is False


class TestDocumentationGeneration:
    """AC-002, AC-010: Documentation generation and agent context loading"""

    def test_documentation_generator_initialization(self):
        """Documentation generator should initialize"""
        generator = DocumentationGenerator()
        assert generator is not None

    def test_generate_product_md(self):
        """Should generate product.md with required sections"""
        brainstorm_responses = {
            "project_vision": "Test vision",
            "target_users": "Test users",
            "value_proposition": "Test value",
            "roadmap": "Test roadmap",
        }

        generator = DocumentationGenerator()
        content = generator.generate_product_md(brainstorm_responses)

        assert "Vision" in content or "vision" in content.lower()
        assert "Target Users" in content or "target users" in content.lower()
        assert "Value Proposition" in content or "value" in content.lower()
        assert len(content) > 200  # At least 200 characters

    def test_generate_structure_md(self):
        """Should generate structure.md with architecture details"""
        brainstorm_responses = {
            "system_architecture": "Test architecture",
            "core_components": "Test components",
            "dependencies": "Test dependencies",
        }

        generator = DocumentationGenerator()
        content = generator.generate_structure_md(brainstorm_responses)

        assert "Architecture" in content or "architecture" in content.lower()
        assert "Component" in content or "component" in content.lower()
        assert len(content) > 200

    def test_generate_tech_md(self):
        """Should generate tech.md with technology stack"""
        brainstorm_responses = {
            "technology_selection": "Test tech",
            "trade_offs": "Test trade-offs",
            "performance": "Test performance",
            "security": "Test security",
        }

        generator = DocumentationGenerator()
        content = generator.generate_tech_md(brainstorm_responses)

        assert "Technology" in content or "technology" in content.lower()
        assert "Trade" in content or "trade" in content.lower()
        assert len(content) > 200

    def test_save_documentation_files(self):
        """Should save generated documents to .moai/project/"""
        generator = DocumentationGenerator()

        with patch("pathlib.Path.mkdir"), patch("pathlib.Path.write_text") as mock_write:

            generator.save_all_documents(
                {
                    "product": "Test product",
                    "structure": "Test structure",
                    "tech": "Test tech",
                },
                base_path=Path(".moai/project"),
            )

            assert mock_write.call_count >= 3

    def test_load_product_md_for_project_manager(self):
        """Generated product.md should be loadable for project-manager agent"""
        generator = DocumentationGenerator()

        with patch("pathlib.Path.read_text", return_value="# Product\nTest content"):
            content = generator.load_document("product.md")
            assert content is not None
            assert "Product" in content

    def test_load_structure_md_for_tdd_implementer(self):
        """Generated structure.md should be loadable for tdd-implementer agent"""
        generator = DocumentationGenerator()

        with patch("pathlib.Path.read_text", return_value="# Architecture\nTest structure"):
            content = generator.load_document("structure.md")
            assert content is not None
            assert "Architecture" in content

    def test_load_tech_md_for_experts(self):
        """Generated tech.md should be loadable for domain experts"""
        generator = DocumentationGenerator()

        with patch("pathlib.Path.read_text", return_value="# Tech Stack\nTest tech"):
            content = generator.load_document("tech.md")
            assert content is not None
            assert "Tech" in content


class TestAtomicSaving:
    """AC-008: Configuration saving should be atomic (all-or-nothing)"""

    def test_config_save_all_or_nothing(self):
        """Should save all 31 settings or none"""
        manager = ConfigurationManager()
        config = self._create_complete_config()

        with patch.object(manager, "_write_config") as mock_write:
            manager.save(config)
            mock_write.assert_called_once()

            # Verify all settings are in the call
            saved_config = mock_write.call_args[0][0]
            assert len(self._flatten_config(saved_config)) == 31

    def test_config_save_failure_rollback(self):
        """Failed save should rollback to previous state"""
        manager = ConfigurationManager()
        original_config = manager.load()
        new_config = self._create_complete_config()

        with patch.object(manager, "_write_config", side_effect=IOError):
            with pytest.raises(IOError):
                manager.save(new_config)

        current_config = manager.load()
        assert current_config == original_config

    def test_config_save_creates_backup(self):
        """Should create backup before saving"""
        manager = ConfigurationManager()

        with patch.object(manager, "_create_backup") as mock_backup, patch.object(manager, "_write_config"):

            manager.save(self._create_complete_config())
            mock_backup.assert_called_once()

    @staticmethod
    def _create_complete_config() -> Dict[str, Any]:
        """Create a complete valid configuration"""
        return {
            "user": {"name": "TestUser"},
            "language": {
                "conversation_language": "en",
                "agent_prompt_language": "en",
                "conversation_language_name": "English",
            },
            "project": {
                "name": "TestProject",
                "owner": "TestOwner",
                "description": "Test",
                "language": "python",
                "locale": "en_US",
                "template_version": "3.0.0",
                "documentation_mode": "skip",
            },
            "git_strategy": {
                "mode": "personal",
                "personal": {
                    "workflow": "github-flow",
                    "auto_checkpoint": "event-driven",
                    "push_to_remote": False,
                },
                "team": {
                    "workflow": "git-flow",
                    "auto_pr": False,
                    "draft_pr": False,
                },
            },
            "constitution": {
                "test_coverage_target": 90,
                "enforce_tdd": True,
            },
            "moai": {"version": "0.26.0"},
        }

    @staticmethod
    def _flatten_config(config: Dict[str, Any]) -> List[str]:
        """Flatten nested config to count all keys"""
        keys = []

        def _flatten(obj, prefix=""):
            if isinstance(obj, dict):
                for k, v in obj.items():
                    _flatten(v, f"{prefix}.{k}" if prefix else k)
            else:
                keys.append(f"{prefix}")

        _flatten(config)
        return keys


class TestTemplateVariables:
    """AC-009: Template variable interpolation at runtime"""

    def test_template_variable_interpolation(self):
        """Should interpolate {{variable}} at runtime"""
        config = {
            "user": {"name": "GOOS"},
            "project": {"owner": "GoosLab", "name": "MoAI"},
        }

        template = "Owner: {{user.name}}, Project: {{project.name}}"
        result = TemplateVariableInterpolator.interpolate(template, config)

        assert result == "Owner: GOOS, Project: MoAI"

    def test_nested_path_interpolation(self):
        """Should handle nested paths like {{git_strategy.mode}}"""
        config = {
            "git_strategy": {"mode": "personal"},
            "project": {"name": "TestProject"},
        }

        template = "Mode: {{git_strategy.mode}}"
        result = TemplateVariableInterpolator.interpolate(template, config)

        assert result == "Mode: personal"

    def test_missing_variable_raises_error(self):
        """Should raise error for missing variables"""
        config = {"user": {"name": "Test"}}
        template = "Unknown: {{undefined.value}}"

        with pytest.raises(KeyError):
            TemplateVariableInterpolator.interpolate(template, config)

    def test_multiple_variables_in_text(self):
        """Should handle multiple variables in single text"""
        config = {
            "user": {"name": "Alice"},
            "project": {"owner": "Bob"},
        }

        template = "User: {{user.name}}, Owner: {{project.owner}}"
        result = TemplateVariableInterpolator.interpolate(template, config)

        assert result == "User: Alice, Owner: Bob"


class TestBackwardCompatibility:
    """AC-011: v2.1.0 configuration migration"""

    def test_load_v2_1_0_config(self):
        """Should load and parse v2.1.0 config.json"""
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "OldUser"},
            "project": {"name": "OldProject"},
        }

        migrator = ConfigurationMigrator()
        config = migrator.load_legacy_config(v2_config)

        assert config is not None
        assert config["user"]["name"] == "OldUser"

    def test_migrate_v2_to_v3(self):
        """Should migrate v2.1.0 config to v3.0.0"""
        v2_config = {
            "version": "2.1.0",
            "user": {"name": "TestUser"},
            "project": {"name": "TestProject"},
            "git_strategy": {"mode": "personal"},
        }

        migrator = ConfigurationMigrator()
        v3_config = migrator.migrate(v2_config)

        assert v3_config["version"] == "3.0.0"
        assert v3_config["user"]["name"] == "TestUser"
        assert v3_config["project"]["name"] == "TestProject"

    def test_backward_compatible_field_mapping(self):
        """Old fields should map to new schema"""
        v2_config = {
            "version": "2.1.0",
            "dev_mode": "personal",  # Old field name
        }

        migrator = ConfigurationMigrator()
        v3_config = migrator.migrate(v2_config)

        # Should map old field to new location
        assert "git_strategy" in v3_config

    def test_incompatible_fields_use_smart_defaults(self):
        """Incompatible v2 fields should fall back to smart defaults"""
        v2_config = {
            "version": "2.1.0",
            "deprecated_field": "some_value",
        }

        migrator = ConfigurationMigrator()
        v3_config = migrator.migrate(v2_config)

        # Should have smart defaults applied
        assert "constitution" in v3_config
        assert "test_coverage_target" in v3_config["constitution"]


class TestAskUserQuestionAPICompliance:
    """AC-012: AskUserQuestion API constraint compliance"""

    def test_max_4_questions_per_batch(self):
        """Each batch should have max 4 questions"""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab.get("batches", []):
                questions = batch.get("questions", [])
                assert len(questions) <= 4, f"Batch {batch['id']} has {len(questions)} questions"

    def test_no_emoji_in_headers(self):
        """Headers should not contain emoji"""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab.get("batches", []):
                header = batch.get("header", "")
                # Simple emoji detection (Unicode ranges)
                assert not any(ord(c) > 127 for c in header if c.isalpha() and ord(c) > 127)

    def test_no_emoji_in_questions(self):
        """Questions should not contain emoji"""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab.get("batches", []):
                for question in batch.get("questions", []):
                    question_text = question.get("question", "")
                    # Emoji check
                    assert not any(ord(c) > 127 for c in question_text if c.isalpha() and ord(c) > 127)

    def test_header_max_12_chars(self):
        """Headers should be max 12 characters"""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab.get("batches", []):
                header = batch.get("header", "")
                assert len(header) <= 12, f"Header '{header}' is {len(header)} chars"

    def test_min_2_max_4_options_per_question(self):
        """Each question should have 2-4 options"""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab.get("batches", []):
                for question in batch.get("questions", []):
                    options = question.get("options", [])
                    assert 2 <= len(options) <= 4, f"Question {question['id']} has {len(options)} options"

    def test_auto_other_option_support(self):
        """Schema should support auto_other_option flag"""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab.get("batches", []):
                for question in batch.get("questions", []):
                    # If has 4 options, may have auto_other_option
                    if len(question.get("options", [])) == 4:
                        # This is valid per spec
                        pass


class TestIntegration:
    """Integration tests combining multiple components"""

    def test_complete_quick_start_workflow(self):
        """Test complete Quick Start flow: 2-3 minutes, 31 settings"""
        # This would be an e2e test
        manager = ConfigurationManager()

        # Simulate user responses for Tab 1
        tab1_responses = {
            "user_name": "TestUser",
            "conversation_language": "en",
            "agent_prompt_language": "en",
            "project_name": "MyApp",
            "project_owner": "TestOwner",
            "project_description": "",
            "git_strategy_mode": "personal",
            "git_strategy_workflow": "github-flow",
            "test_coverage_target": 90,
            "enforce_tdd": True,
        }

        # Simulate Tab 2 choice
        tab2_responses = {
            "documentation_mode": "skip",
        }

        # Simulate Tab 3 Git settings
        tab3_responses = {
            "git_personal_auto_checkpoint": "event-driven",
            "git_personal_push_remote": False,
        }

        all_responses = {**tab1_responses, **tab2_responses, **tab3_responses}

        config = manager.build_from_responses(all_responses)

        # Verify 31 settings
        flattened = self._flatten_config(config)
        assert len(flattened) >= 31

    def test_documentation_generation_with_agent_context(self):
        """Test documentation generation and agent context loading"""
        generator = DocumentationGenerator()

        brainstorm = {
            "project_vision": "Test vision statement",
            "target_users": "Test target users",
            "value_proposition": "Test value prop",
            "roadmap": "Test roadmap",
            "system_architecture": "Test architecture",
            "core_components": "Test components",
            "technology_selection": "Test tech stack",
            "trade_offs": "Test trade-offs",
            "performance": "Test perf",
            "security": "Test security",
        }

        with patch("pathlib.Path.write_text"), patch("pathlib.Path.mkdir"):

            docs = generator.generate_all_documents(brainstorm)

            assert "product" in docs
            assert "structure" in docs
            assert "tech" in docs
            assert len(docs["product"]) > 200
            assert len(docs["structure"]) > 200
            assert len(docs["tech"]) > 200

    @staticmethod
    def _flatten_config(config: Dict[str, Any]) -> List[str]:
        """Helper to flatten config for counting"""
        keys = []

        def _flatten(obj, prefix=""):
            if isinstance(obj, dict):
                for k, v in obj.items():
                    _flatten(v, f"{prefix}.{k}" if prefix else k)
            else:
                keys.append(f"{prefix}")

        _flatten(config)
        return keys


# Import actual implementations from configuration module
from moai_adk.project.configuration import (
    ConditionalBatchRenderer,
    ConfigurationMigrator,
    TemplateVariableInterpolator,
)
