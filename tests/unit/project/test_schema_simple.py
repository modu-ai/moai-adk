"""
Simple unit tests for moai_adk.project.schema module.

Tests cover:
- load_tab_schema function
- Tab schema structure and validation
"""

import pytest

from moai_adk.project.schema import (
    _create_tab1_quick_start,
    _create_tab2_documentation,
    _create_tab3_git_automation,
    load_tab_schema,
)


class TestLoadTabSchema:
    """Test load_tab_schema function."""

    def test_load_tab_schema_returns_dict(self):
        """Test that load_tab_schema returns a dictionary."""
        schema = load_tab_schema()
        assert isinstance(schema, dict)

    def test_load_tab_schema_has_version(self):
        """Test that schema has version field."""
        schema = load_tab_schema()
        assert "version" in schema
        assert schema["version"] == "3.0.0"

    def test_load_tab_schema_has_three_tabs(self):
        """Test that schema contains exactly three tabs."""
        schema = load_tab_schema()
        assert "tabs" in schema
        assert len(schema["tabs"]) == 3

    def test_load_tab_schema_tab_ids(self):
        """Test that tabs have correct IDs."""
        schema = load_tab_schema()
        tabs = schema["tabs"]

        assert tabs[0]["id"] == "tab_1_quick_start"
        assert tabs[1]["id"] == "tab_2_documentation"
        assert tabs[2]["id"] == "tab_3_git_automation"

    def test_load_tab_schema_tab_labels(self):
        """Test that tabs have labels."""
        schema = load_tab_schema()
        tabs = schema["tabs"]

        assert tabs[0]["label"] == "Essential Setup"
        assert tabs[1]["label"] == "Documentation"
        assert tabs[2]["label"] == "Git"

    def test_load_tab_schema_tab_descriptions(self):
        """Test that tabs have descriptions."""
        schema = load_tab_schema()
        tabs = schema["tabs"]

        assert "description" in tabs[0]
        assert "description" in tabs[1]
        assert "description" in tabs[2]


class TestCreateTab1QuickStart:
    """Test _create_tab1_quick_start function."""

    def test_tab1_structure(self):
        """Test Tab 1 structure."""
        tab = _create_tab1_quick_start()

        assert tab["id"] == "tab_1_quick_start"
        assert tab["label"] == "Essential Setup"
        assert "batches" in tab
        assert len(tab["batches"]) == 4

    def test_tab1_batch_structure(self):
        """Test Tab 1 batch structure."""
        tab = _create_tab1_quick_start()
        batches = tab["batches"]

        # Check first batch
        assert batches[0]["id"] == "batch_1_1_identity"
        assert batches[0]["batch_number"] == 1
        assert batches[0]["total_batches"] == 4
        assert "questions" in batches[0]

    def test_tab1_batch1_questions(self):
        """Test Tab 1 Batch 1 has 3 questions."""
        tab = _create_tab1_quick_start()
        batch1 = tab["batches"][0]

        assert len(batch1["questions"]) == 3
        assert batch1["questions"][0]["id"] == "user_name"
        assert batch1["questions"][1]["id"] == "conversation_language"
        assert batch1["questions"][2]["id"] == "agent_prompt_language"

    def test_tab1_batch2_questions(self):
        """Test Tab 1 Batch 2 has 3 questions."""
        tab = _create_tab1_quick_start()
        batch2 = tab["batches"][1]

        assert len(batch2["questions"]) == 3
        assert batch2["questions"][0]["id"] == "project_name"
        assert batch2["questions"][1]["id"] == "github_profile_name"  # Changed from project_owner
        assert batch2["questions"][2]["id"] == "project_description"

    def test_tab1_batch3_questions(self):
        """Test Tab 1 Batch 3 has 2 questions."""
        tab = _create_tab1_quick_start()
        batch3 = tab["batches"][2]

        assert len(batch3["questions"]) == 2
        assert batch3["questions"][0]["id"] == "git_strategy_mode"
        assert batch3["questions"][1]["id"] == "git_strategy_workflow"

    def test_tab1_batch4_questions(self):
        """Test Tab 1 Batch 4 has 2 questions."""
        tab = _create_tab1_quick_start()
        batch4 = tab["batches"][3]

        assert len(batch4["questions"]) == 2
        assert batch4["questions"][0]["id"] == "test_coverage_target"
        assert batch4["questions"][1]["id"] == "enforce_tdd"

    def test_tab1_question_has_required_fields(self):
        """Test that questions have required fields."""
        tab = _create_tab1_quick_start()
        batch1 = tab["batches"][0]
        question = batch1["questions"][0]

        assert "id" in question
        assert "question" in question
        assert "type" in question
        assert "required" in question


class TestCreateTab2Documentation:
    """Test _create_tab2_documentation function."""

    def test_tab2_structure(self):
        """Test Tab 2 structure."""
        tab = _create_tab2_documentation()

        assert tab["id"] == "tab_2_documentation"
        assert tab["label"] == "Documentation"
        assert "batches" in tab
        assert len(tab["batches"]) == 2

    def test_tab2_batch1_questions(self):
        """Test Tab 2 Batch 1 has documentation mode question."""
        tab = _create_tab2_documentation()
        batch1 = tab["batches"][0]

        assert len(batch1["questions"]) == 1
        assert batch1["questions"][0]["id"] == "documentation_mode"
        assert batch1["questions"][0]["type"] == "select_single"

    def test_tab2_batch2_questions(self):
        """Test Tab 2 Batch 2 has conditional depth question."""
        tab = _create_tab2_documentation()
        batch2 = tab["batches"][1]

        assert len(batch2["questions"]) == 1
        assert batch2["questions"][0]["id"] == "documentation_depth"
        assert "show_if" in batch2

    def test_tab2_documentation_mode_options(self):
        """Test Tab 2 documentation mode options."""
        tab = _create_tab2_documentation()
        batch1 = tab["batches"][0]
        question = batch1["questions"][0]

        assert len(question["options"]) == 3
        option_values = [opt["value"] for opt in question["options"]]
        assert "skip" in option_values
        assert "full_now" in option_values
        assert "minimal" in option_values


class TestCreateTab3GitAutomation:
    """Test _create_tab3_git_automation function."""

    def test_tab3_structure(self):
        """Test Tab 3 structure."""
        tab = _create_tab3_git_automation()

        assert tab["id"] == "tab_3_git_automation"
        assert tab["label"] == "Git"
        assert "batches" in tab
        assert len(tab["batches"]) == 2

    def test_tab3_batch1_personal_settings(self):
        """Test Tab 3 Batch 1 contains personal settings."""
        tab = _create_tab3_git_automation()
        batch1 = tab["batches"][0]

        assert "batch_3_1_personal" in batch1["id"]
        assert len(batch1["questions"]) == 2
        assert batch1["questions"][0]["id"] == "git_personal_auto_checkpoint"
        assert batch1["questions"][1]["id"] == "git_personal_push_remote"

    def test_tab3_batch2_team_settings(self):
        """Test Tab 3 Batch 2 contains team settings."""
        tab = _create_tab3_git_automation()
        batch2 = tab["batches"][1]

        assert "batch_3_1_team" in batch2["id"]
        assert len(batch2["questions"]) == 2
        assert batch2["questions"][0]["id"] == "git_team_auto_pr"
        assert batch2["questions"][1]["id"] == "git_team_draft_pr"

    def test_tab3_batches_have_conditional_logic(self):
        """Test that Tab 3 batches have conditional show_if logic."""
        tab = _create_tab3_git_automation()
        batch1 = tab["batches"][0]
        batch2 = tab["batches"][1]

        assert "show_if" in batch1
        assert "show_if" in batch2
        assert "personal" in batch1["show_if"]
        assert "team" in batch2["show_if"]


class TestSchemaValidation:
    """Test schema validation and completeness."""

    def test_all_questions_have_id(self):
        """Test that all questions have id field."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    assert "id" in question
                    assert isinstance(question["id"], str)
                    assert len(question["id"]) > 0

    def test_all_questions_have_question_text(self):
        """Test that all questions have question text."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    assert "question" in question
                    assert isinstance(question["question"], str)

    def test_all_questions_have_type(self):
        """Test that all questions have type."""
        schema = load_tab_schema()

        valid_types = {"text_input", "select_single", "number_input", "select_multiple"}

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    assert "type" in question
                    assert question["type"] in valid_types

    def test_all_questions_have_required_field(self):
        """Test that all questions have required field."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    assert "required" in question
                    assert isinstance(question["required"], bool)

    def test_select_questions_have_options(self):
        """Test that select questions have options."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    if question["type"] in {"select_single", "select_multiple"}:
                        assert "options" in question
                        assert len(question["options"]) > 0

    def test_all_batches_have_required_fields(self):
        """Test that all batches have required fields."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                assert "id" in batch
                assert "header" in batch
                assert "batch_number" in batch
                assert "total_batches" in batch
                assert "questions" in batch

    def test_schema_is_consistent(self):
        """Test that schema structure is consistent across tabs."""
        schema = load_tab_schema()

        # All items in tabs should be dicts
        for tab in schema["tabs"]:
            assert isinstance(tab, dict)
            assert "id" in tab
            assert "label" in tab
            assert "description" in tab
            assert "batches" in tab
            assert isinstance(tab["batches"], list)

    def test_tab1_has_exactly_10_questions(self):
        """Test that Tab 1 has exactly 10 questions total."""
        tab1 = _create_tab1_quick_start()

        total_questions = sum(len(batch["questions"]) for batch in tab1["batches"])
        assert total_questions == 10

    def test_conditional_workflow_mapping_exists(self):
        """Test that conditional mappings exist for git_strategy_workflow."""
        tab1 = _create_tab1_quick_start()
        batch3 = tab1["batches"][2]
        workflow_question = batch3["questions"][1]

        assert "conditional_mapping" in workflow_question
        assert "personal" in workflow_question["conditional_mapping"]
        assert "team" in workflow_question["conditional_mapping"]
        assert "hybrid" in workflow_question["conditional_mapping"]
