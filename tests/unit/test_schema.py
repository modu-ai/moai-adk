"""Tests for moai_adk.project.schema module."""

from moai_adk.project.schema import (
    _create_tab1_quick_start,
    _create_tab2_documentation,
    _create_tab3_git_automation,
    load_tab_schema,
)


class TestTabSchemaLoading:
    """Test schema loading functions."""

    def test_load_tab_schema_returns_dict(self):
        """Test that load_tab_schema returns a dictionary."""
        schema = load_tab_schema()
        assert isinstance(schema, dict)

    def test_load_tab_schema_version(self):
        """Test that loaded schema has correct version."""
        schema = load_tab_schema()
        assert schema["version"] == "3.0.0"

    def test_load_tab_schema_contains_tabs(self):
        """Test that schema contains tabs array."""
        schema = load_tab_schema()
        assert "tabs" in schema
        assert isinstance(schema["tabs"], list)
        assert len(schema["tabs"]) == 3

    def test_load_tab_schema_tab_ids(self):
        """Test that all tabs have correct IDs."""
        schema = load_tab_schema()
        tab_ids = [tab["id"] for tab in schema["tabs"]]
        assert "tab_1_quick_start" in tab_ids
        assert "tab_2_documentation" in tab_ids
        assert "tab_3_git_automation" in tab_ids

    def test_tab1_quick_start_structure(self):
        """Test Tab 1 Quick Start structure."""
        tab = _create_tab1_quick_start()
        assert tab["id"] == "tab_1_quick_start"
        assert tab["label"] == "Essential Setup"
        assert "batches" in tab
        assert len(tab["batches"]) == 4

    def test_tab1_batch_numbers(self):
        """Test that Tab 1 batches have correct numbering."""
        tab = _create_tab1_quick_start()
        for idx, batch in enumerate(tab["batches"], 1):
            assert batch["batch_number"] == idx
            assert batch["total_batches"] == 4

    def test_tab1_questions_count(self):
        """Test Tab 1 question distribution."""
        tab = _create_tab1_quick_start()
        total_questions = sum(len(batch["questions"]) for batch in tab["batches"])
        assert total_questions == 10

    def test_tab2_documentation_structure(self):
        """Test Tab 2 Documentation structure."""
        tab = _create_tab2_documentation()
        assert tab["id"] == "tab_2_documentation"
        assert tab["label"] == "Documentation"
        assert "batches" in tab
        assert len(tab["batches"]) == 2

    def test_tab2_batch_structure(self):
        """Test Tab 2 batch structure."""
        tab = _create_tab2_documentation()
        for batch in tab["batches"]:
            assert "id" in batch
            assert "header" in batch
            assert "batch_number" in batch
            assert "total_batches" in batch
            assert batch["total_batches"] == 2

    def test_tab2_conditional_question(self):
        """Test Tab 2 conditional question."""
        tab = _create_tab2_documentation()
        second_batch = tab["batches"][1]
        assert "show_if" in second_batch
        assert "documentation_mode == 'full_now'" in second_batch["show_if"]

    def test_tab3_git_automation_structure(self):
        """Test Tab 3 Git Automation structure."""
        tab = _create_tab3_git_automation()
        assert tab["id"] == "tab_3_git_automation"
        assert tab["label"] == "Git"
        assert "batches" in tab
        assert len(tab["batches"]) == 2

    def test_tab3_personal_batch_conditional(self):
        """Test Tab 3 personal batch conditional logic."""
        tab = _create_tab3_git_automation()
        personal_batch = tab["batches"][0]
        assert "show_if" in personal_batch
        assert "git_strategy_mode == 'personal' OR git_strategy_mode == 'hybrid'" in personal_batch["show_if"]

    def test_tab3_team_batch_conditional(self):
        """Test Tab 3 team batch conditional logic."""
        tab = _create_tab3_git_automation()
        team_batch = tab["batches"][1]
        assert "show_if" in team_batch
        assert "git_strategy_mode == 'team'" in team_batch["show_if"]

    def test_all_questions_have_ids(self):
        """Test that all questions have unique IDs."""
        schema = load_tab_schema()
        all_question_ids = []

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    assert "id" in question
                    all_question_ids.append(question["id"])

        # Check for uniqueness
        assert len(all_question_ids) == len(set(all_question_ids))

    def test_all_questions_have_type(self):
        """Test that all questions have a type field."""
        schema = load_tab_schema()

        valid_types = {"text_input", "select_single", "select_multiple", "number_input"}

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    assert "type" in question
                    assert question["type"] in valid_types

    def test_number_input_constraints(self):
        """Test number input questions have min/max constraints."""
        tab = _create_tab1_quick_start()
        number_questions = []

        for batch in tab["batches"]:
            for question in batch["questions"]:
                if question["type"] == "number_input":
                    number_questions.append(question)
                    assert "min" in question
                    assert "max" in question

        assert len(number_questions) > 0

    def test_tab1_git_strategy_conditional_mapping(self):
        """Test git strategy conditional mapping in Tab 1."""
        tab = _create_tab1_quick_start()
        batch_3 = tab["batches"][2]
        git_workflow_q = batch_3["questions"][1]

        assert "conditional_mapping" in git_workflow_q
        assert "personal" in git_workflow_q["conditional_mapping"]
        assert "team" in git_workflow_q["conditional_mapping"]
        assert "hybrid" in git_workflow_q["conditional_mapping"]

    def test_required_field_presence(self):
        """Test that all questions have required field."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    assert "required" in question
                    assert isinstance(question["required"], bool)

    def test_smart_default_values(self):
        """Test that questions have appropriate default values."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    if "smart_default" in question:
                        # Just verify it's present and not None
                        assert question["smart_default"] is not None

    def test_schema_is_immutable_on_reload(self):
        """Test that multiple loads return equivalent schemas."""
        schema1 = load_tab_schema()
        schema2 = load_tab_schema()

        assert schema1 == schema2

    def test_select_single_has_options(self):
        """Test that select_single questions have options."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    if question["type"] == "select_single":
                        assert "options" in question
                        assert len(question["options"]) > 0
                        for option in question["options"]:
                            assert "label" in option
                            assert "value" in option

    def test_batch_headers_exist(self):
        """Test that all batches have headers."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                assert "header" in batch
                assert len(batch["header"]) > 0

    def test_question_ids_follow_naming_convention(self):
        """Test that question IDs follow naming conventions."""
        schema = load_tab_schema()

        for tab in schema["tabs"]:
            for batch in tab["batches"]:
                for question in batch["questions"]:
                    q_id = question["id"]
                    # IDs should be lowercase with underscores
                    assert q_id.islower()
                    assert "_" in q_id or len(q_id) > 5
