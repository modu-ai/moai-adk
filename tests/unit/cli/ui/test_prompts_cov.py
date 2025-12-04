"""Comprehensive tests for prompts.py module.

Focus on uncovered code paths with actual execution using mocked dependencies.
"""

import pytest
from unittest.mock import MagicMock, Mock, patch, call
from typing import Any, Dict, List

from InquirerPy.base.control import Choice, Separator
from moai_adk.cli.ui.prompts import (
    fuzzy_checkbox,
    fuzzy_select,
    styled_checkbox,
    styled_select,
    styled_input,
    styled_confirm,
    _process_choices,
    create_grouped_choices,
)


class TestFuzzyCheckbox:
    """Test fuzzy_checkbox function."""

    def test_basic_fuzzy_checkbox(self):
        """Test basic fuzzy checkbox prompt."""
        choices = ["File 1", "File 2", "File 3"]

        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = ["File 1", "File 2"]
            mock_fuzzy.return_value = mock_result

            result = fuzzy_checkbox("Select files", choices)
            assert result == ["File 1", "File 2"]
            mock_fuzzy.assert_called_once()

    def test_fuzzy_checkbox_with_default(self):
        """Test fuzzy checkbox with default values."""
        choices = ["Option A", "Option B", "Option C"]

        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = ["Option B"]
            mock_fuzzy.return_value = mock_result

            result = fuzzy_checkbox("Select options", choices, default=["Option B"])
            assert result == ["Option B"]

    def test_fuzzy_checkbox_keyboard_interrupt(self):
        """Test fuzzy checkbox handles KeyboardInterrupt."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.side_effect = KeyboardInterrupt()
            mock_fuzzy.return_value = mock_result

            result = fuzzy_checkbox("Select", [])
            assert result is None

    def test_fuzzy_checkbox_custom_instruction(self):
        """Test fuzzy checkbox with custom instruction."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = []
            mock_fuzzy.return_value = mock_result

            fuzzy_checkbox("Choose", ["A", "B"], instruction="Custom help text")

            _, kwargs = mock_fuzzy.call_args
            assert kwargs["instruction"] == "Custom help text"

    def test_fuzzy_checkbox_custom_marker(self):
        """Test fuzzy checkbox with custom marker."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = []
            mock_fuzzy.return_value = mock_result

            fuzzy_checkbox("Choose", ["A", "B"], marker="★")

            _, kwargs = mock_fuzzy.call_args
            assert kwargs["marker"] == "★"

    def test_fuzzy_checkbox_custom_keybindings(self):
        """Test fuzzy checkbox with custom keybindings."""
        custom_keys = {"toggle": [{"key": "space"}]}

        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = []
            mock_fuzzy.return_value = mock_result

            fuzzy_checkbox("Choose", ["A", "B"], keybindings=custom_keys)

            _, kwargs = mock_fuzzy.call_args
            assert kwargs["keybindings"] == custom_keys

    def test_fuzzy_checkbox_default_keybindings(self):
        """Test fuzzy checkbox with default keybindings."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = []
            mock_fuzzy.return_value = mock_result

            fuzzy_checkbox("Choose", ["A", "B"])

            _, kwargs = mock_fuzzy.call_args
            assert "keybindings" in kwargs
            keybindings = kwargs["keybindings"]
            assert "toggle" in keybindings
            assert "toggle-all" in keybindings

    def test_fuzzy_checkbox_with_border(self):
        """Test fuzzy checkbox border configuration."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = []
            mock_fuzzy.return_value = mock_result

            fuzzy_checkbox("Choose", ["A"], border=False)
            _, kwargs = mock_fuzzy.call_args
            assert kwargs["border"] is False

    def test_fuzzy_checkbox_height_parameters(self):
        """Test fuzzy checkbox with height parameters."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = []
            mock_fuzzy.return_value = mock_result

            fuzzy_checkbox("Choose", ["A"], height=20)
            _, kwargs = mock_fuzzy.call_args
            assert kwargs["max_height"] == 20

    def test_fuzzy_checkbox_validation(self):
        """Test fuzzy checkbox with validation."""
        validator = lambda x: len(x) > 0

        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = ["A"]
            mock_fuzzy.return_value = mock_result

            fuzzy_checkbox("Choose", ["A"], validate=validator)
            _, kwargs = mock_fuzzy.call_args
            assert kwargs["validate"] == validator

    def test_fuzzy_checkbox_transformer(self):
        """Test that fuzzy checkbox has transformer function."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = ["A", "B"]
            mock_fuzzy.return_value = mock_result

            result = fuzzy_checkbox("Choose", ["A", "B"])
            _, kwargs = mock_fuzzy.call_args
            assert "transformer" in kwargs
            # Test transformer
            transformer = kwargs["transformer"]
            assert "2 selected" in transformer(["A", "B"])


class TestFuzzySelect:
    """Test fuzzy_select function."""

    def test_basic_fuzzy_select(self):
        """Test basic fuzzy select prompt."""
        choices = ["Option A", "Option B", "Option C"]

        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = "Option B"
            mock_fuzzy.return_value = mock_result

            result = fuzzy_select("Select one", choices)
            assert result == "Option B"

    def test_fuzzy_select_with_default(self):
        """Test fuzzy select with default value."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = "B"
            mock_fuzzy.return_value = mock_result

            result = fuzzy_select("Select", ["A", "B"], default="B")
            assert result == "B"

    def test_fuzzy_select_keyboard_interrupt(self):
        """Test fuzzy select handles KeyboardInterrupt."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.side_effect = KeyboardInterrupt()
            mock_fuzzy.return_value = mock_result

            result = fuzzy_select("Select", ["A"])
            assert result is None

    def test_fuzzy_select_multiselect_false(self):
        """Test that fuzzy_select has multiselect=False."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = "A"
            mock_fuzzy.return_value = mock_result

            fuzzy_select("Select", ["A", "B"])
            _, kwargs = mock_fuzzy.call_args
            assert kwargs["multiselect"] is False

    def test_fuzzy_select_instruction(self):
        """Test fuzzy_select custom instruction."""
        with patch("moai_adk.cli.ui.prompts.inquirer.fuzzy") as mock_fuzzy:
            mock_result = MagicMock()
            mock_result.execute.return_value = "A"
            mock_fuzzy.return_value = mock_result

            fuzzy_select("Select", ["A"], instruction="Custom help")
            _, kwargs = mock_fuzzy.call_args
            assert kwargs["instruction"] == "Custom help"


class TestStyledCheckbox:
    """Test styled_checkbox function."""

    def test_basic_styled_checkbox(self):
        """Test basic styled checkbox prompt."""
        choices = ["Check 1", "Check 2"]

        with patch("moai_adk.cli.ui.prompts.inquirer.checkbox") as mock_checkbox:
            mock_result = MagicMock()
            mock_result.execute.return_value = ["Check 1"]
            mock_checkbox.return_value = mock_result

            result = styled_checkbox("Select", choices)
            assert result == ["Check 1"]

    def test_styled_checkbox_with_default(self):
        """Test styled checkbox with default values."""
        with patch("moai_adk.cli.ui.prompts.inquirer.checkbox") as mock_checkbox:
            mock_result = MagicMock()
            mock_result.execute.return_value = ["Option A"]
            mock_checkbox.return_value = mock_result

            result = styled_checkbox(
                "Choose", ["Option A", "Option B"], default=["Option A"]
            )
            assert result == ["Option A"]

    def test_styled_checkbox_keyboard_interrupt(self):
        """Test styled checkbox handles KeyboardInterrupt."""
        with patch("moai_adk.cli.ui.prompts.inquirer.checkbox") as mock_checkbox:
            mock_result = MagicMock()
            mock_result.execute.side_effect = KeyboardInterrupt()
            mock_checkbox.return_value = mock_result

            result = styled_checkbox("Choose", [])
            assert result is None

    def test_styled_checkbox_cycle(self):
        """Test styled checkbox with cycle parameter."""
        with patch("moai_adk.cli.ui.prompts.inquirer.checkbox") as mock_checkbox:
            mock_result = MagicMock()
            mock_result.execute.return_value = []
            mock_checkbox.return_value = mock_result

            styled_checkbox("Choose", ["A"], cycle=False)
            _, kwargs = mock_checkbox.call_args
            assert kwargs["cycle"] is False

    def test_styled_checkbox_instruction(self):
        """Test styled checkbox with custom instruction."""
        with patch("moai_adk.cli.ui.prompts.inquirer.checkbox") as mock_checkbox:
            mock_result = MagicMock()
            mock_result.execute.return_value = []
            mock_checkbox.return_value = mock_result

            styled_checkbox("Choose", ["A"], instruction="Help text")
            _, kwargs = mock_checkbox.call_args
            assert kwargs["instruction"] == "Help text"


class TestStyledSelect:
    """Test styled_select function."""

    def test_basic_styled_select(self):
        """Test basic styled select prompt."""
        with patch("moai_adk.cli.ui.prompts.inquirer.select") as mock_select:
            mock_result = MagicMock()
            mock_result.execute.return_value = "Option A"
            mock_select.return_value = mock_result

            result = styled_select("Choose", ["Option A", "Option B"])
            assert result == "Option A"

    def test_styled_select_with_default(self):
        """Test styled select with default value."""
        with patch("moai_adk.cli.ui.prompts.inquirer.select") as mock_select:
            mock_result = MagicMock()
            mock_result.execute.return_value = "B"
            mock_select.return_value = mock_result

            result = styled_select("Choose", ["A", "B"], default="B")
            assert result == "B"

    def test_styled_select_keyboard_interrupt(self):
        """Test styled select handles KeyboardInterrupt."""
        with patch("moai_adk.cli.ui.prompts.inquirer.select") as mock_select:
            mock_result = MagicMock()
            mock_result.execute.side_effect = KeyboardInterrupt()
            mock_select.return_value = mock_result

            result = styled_select("Choose", ["A"])
            assert result is None

    def test_styled_select_cycle(self):
        """Test styled select with cycle parameter."""
        with patch("moai_adk.cli.ui.prompts.inquirer.select") as mock_select:
            mock_result = MagicMock()
            mock_result.execute.return_value = "A"
            mock_select.return_value = mock_result

            styled_select("Choose", ["A"], cycle=False)
            _, kwargs = mock_select.call_args
            assert kwargs["cycle"] is False

    def test_styled_select_instruction(self):
        """Test styled select with custom instruction."""
        with patch("moai_adk.cli.ui.prompts.inquirer.select") as mock_select:
            mock_result = MagicMock()
            mock_result.execute.return_value = "A"
            mock_select.return_value = mock_result

            styled_select("Choose", ["A"], instruction="Navigate")
            _, kwargs = mock_select.call_args
            assert kwargs["instruction"] == "Navigate"


class TestStyledInput:
    """Test styled_input function."""

    def test_basic_styled_input(self):
        """Test basic styled input prompt."""
        with patch("moai_adk.cli.ui.prompts.inquirer.text") as mock_text:
            mock_result = MagicMock()
            mock_result.execute.return_value = "user input"
            mock_text.return_value = mock_result

            result = styled_input("Enter name")
            assert result == "user input"

    def test_styled_input_with_default(self):
        """Test styled input with default value."""
        with patch("moai_adk.cli.ui.prompts.inquirer.text") as mock_text:
            mock_result = MagicMock()
            mock_result.execute.return_value = "default value"
            mock_text.return_value = mock_result

            result = styled_input("Enter", default="default value")
            assert result == "default value"

    def test_styled_input_keyboard_interrupt(self):
        """Test styled input handles KeyboardInterrupt."""
        with patch("moai_adk.cli.ui.prompts.inquirer.text") as mock_text:
            mock_result = MagicMock()
            mock_result.execute.side_effect = KeyboardInterrupt()
            mock_text.return_value = mock_result

            result = styled_input("Enter")
            assert result is None

    def test_styled_input_required(self):
        """Test styled input with required field."""
        with patch("moai_adk.cli.ui.prompts.inquirer.text") as mock_text:
            mock_result = MagicMock()
            mock_result.execute.return_value = "value"
            mock_text.return_value = mock_result

            styled_input("Enter", required=True)
            _, kwargs = mock_text.call_args
            # Validators are passed as a parameter

    def test_styled_input_optional(self):
        """Test styled input with optional field."""
        with patch("moai_adk.cli.ui.prompts.inquirer.text") as mock_text:
            mock_result = MagicMock()
            mock_result.execute.return_value = ""
            mock_text.return_value = mock_result

            styled_input("Enter", required=False)
            _, kwargs = mock_text.call_args
            # No empty validator for optional fields

    def test_styled_input_validation(self):
        """Test styled input with custom validation."""
        validator = lambda x: len(x) > 3

        with patch("moai_adk.cli.ui.prompts.inquirer.text") as mock_text:
            mock_result = MagicMock()
            mock_result.execute.return_value = "valid"
            mock_text.return_value = mock_result

            styled_input("Enter", validate=validator)
            _, kwargs = mock_text.call_args
            assert kwargs["validate"] == validator

    def test_styled_input_instruction(self):
        """Test styled input with instruction."""
        with patch("moai_adk.cli.ui.prompts.inquirer.text") as mock_text:
            mock_result = MagicMock()
            mock_result.execute.return_value = "text"
            mock_text.return_value = mock_result

            styled_input("Enter", instruction="Type something")
            _, kwargs = mock_text.call_args
            assert kwargs["instruction"] == "Type something"


class TestStyledConfirm:
    """Test styled_confirm function."""

    def test_basic_styled_confirm(self):
        """Test basic styled confirm prompt."""
        with patch("moai_adk.cli.ui.prompts.inquirer.confirm") as mock_confirm:
            mock_result = MagicMock()
            mock_result.execute.return_value = True
            mock_confirm.return_value = mock_result

            result = styled_confirm("Proceed?")
            assert result is True

    def test_styled_confirm_with_default_false(self):
        """Test styled confirm with default=False."""
        with patch("moai_adk.cli.ui.prompts.inquirer.confirm") as mock_confirm:
            mock_result = MagicMock()
            mock_result.execute.return_value = False
            mock_confirm.return_value = mock_result

            result = styled_confirm("Proceed?", default=False)
            assert result is False
            _, kwargs = mock_confirm.call_args
            assert kwargs["default"] is False

    def test_styled_confirm_keyboard_interrupt(self):
        """Test styled confirm handles KeyboardInterrupt."""
        with patch("moai_adk.cli.ui.prompts.inquirer.confirm") as mock_confirm:
            mock_result = MagicMock()
            mock_result.execute.side_effect = KeyboardInterrupt()
            mock_confirm.return_value = mock_result

            result = styled_confirm("Proceed?")
            assert result is None

    def test_styled_confirm_instruction(self):
        """Test styled confirm with instruction."""
        with patch("moai_adk.cli.ui.prompts.inquirer.confirm") as mock_confirm:
            mock_result = MagicMock()
            mock_result.execute.return_value = True
            mock_confirm.return_value = mock_result

            styled_confirm("Continue?", instruction="y/n")
            _, kwargs = mock_confirm.call_args
            assert kwargs["instruction"] == "y/n"


class TestProcessChoices:
    """Test _process_choices helper function."""

    def test_process_string_choices(self):
        """Test processing simple string choices."""
        choices = ["Option A", "Option B"]
        result = _process_choices(choices)

        assert len(result) == 2
        assert all(isinstance(c, Choice) for c in result)
        assert result[0].value == "Option A"
        assert result[1].value == "Option B"

    def test_process_choices_with_defaults(self):
        """Test processing choices with defaults."""
        choices = ["A", "B", "C"]
        defaults = ["A", "C"]
        result = _process_choices(choices, defaults)

        assert result[0].enabled is True  # A is in defaults
        assert result[1].enabled is False  # B is not in defaults
        assert result[2].enabled is True  # C is in defaults

    def test_process_dict_choices(self):
        """Test processing dictionary choices."""
        choices = [
            {"name": "First", "value": "first"},
            {"name": "Second", "value": "second"},
        ]
        result = _process_choices(choices)

        assert len(result) == 2
        assert result[0].value == "first"
        assert result[0].name == "First"

    def test_process_dict_choices_with_title_fallback(self):
        """Test dictionary choices fallback to title key."""
        choices = [{"title": "Option", "value": "opt"}]
        result = _process_choices(choices)

        assert result[0].name == "Option"

    def test_process_dict_disabled_choice(self):
        """Test processing disabled dictionary choice."""
        choices = [
            {"name": "Enabled", "value": "enabled"},
            {"name": "Disabled", "value": "disabled", "disabled": True},
        ]
        result = _process_choices(choices)

        assert isinstance(result[0], Choice)
        assert isinstance(result[1], Separator)

    def test_process_choice_objects(self):
        """Test processing Choice objects."""
        choice_obj = Choice(value="test", name="Test")
        result = _process_choices([choice_obj])

        assert len(result) == 1
        assert result[0] == choice_obj

    def test_process_separator_objects(self):
        """Test processing Separator objects."""
        sep = Separator("Category")
        result = _process_choices([sep])

        assert len(result) == 1
        assert result[0] == sep

    def test_process_mixed_choice_types(self):
        """Test processing mixed choice types."""
        choices = [
            "String choice",
            {"name": "Dict choice", "value": "dict"},
            Choice(value="obj", name="Object"),
        ]
        result = _process_choices(choices)

        assert len(result) == 3
        assert result[0].value == "String choice"
        assert result[1].value == "dict"
        assert result[2].value == "obj"

    def test_process_empty_choices(self):
        """Test processing empty choices list."""
        result = _process_choices([])
        assert result == []

    def test_process_dict_choice_value_default_to_name(self):
        """Test dict choice defaults value to name."""
        choices = [{"name": "Option A"}]
        result = _process_choices(choices)

        assert result[0].value == "Option A"
        assert result[0].name == "Option A"

    def test_process_dict_choice_enabled_from_defaults(self):
        """Test dict choice enabled status from defaults."""
        choices = [{"name": "Option", "value": "opt"}]
        result = _process_choices(choices, defaults=["opt"])

        assert result[0].enabled is True

    def test_process_dict_choice_explicit_enabled(self):
        """Test dict choice explicit enabled overrides defaults."""
        choices = [{"name": "Option", "value": "opt", "enabled": True}]
        result = _process_choices(choices)

        assert result[0].enabled is True


class TestCreateGroupedChoices:
    """Test create_grouped_choices function."""

    def test_basic_grouped_choices(self):
        """Test creating basic grouped choices."""
        groups = {
            "Group A": [
                {"name": "A1", "value": "a1"},
                {"name": "A2", "value": "a2"},
            ],
            "Group B": [
                {"name": "B1", "value": "b1"},
            ],
        }
        result = create_grouped_choices(groups)

        # Should have separators + choices
        assert len(result) > 2
        # First should be separator for Group A
        assert isinstance(result[0], Separator)

    def test_grouped_choices_with_defaults(self):
        """Test grouped choices with default values."""
        groups = {
            "Options": [
                {"name": "Option 1", "value": "opt1"},
                {"name": "Option 2", "value": "opt2"},
            ]
        }
        result = create_grouped_choices(groups, defaults=["opt1"])

        # Find the opt1 choice
        choices = [c for c in result if isinstance(c, Choice) and c.value == "opt1"]
        assert len(choices) == 1
        assert choices[0].enabled is True

    def test_grouped_choices_category_separator(self):
        """Test that category separators are included."""
        groups = {
            "Commands": [{"name": "run", "value": "run"}],
            "Settings": [{"name": "config", "value": "config"}],
        }
        result = create_grouped_choices(groups)

        separators = [r for r in result if isinstance(r, Separator)]
        assert len(separators) >= 2

    def test_grouped_choices_empty_group_skipped(self):
        """Test that empty groups are skipped."""
        groups = {
            "Empty": [],
            "Full": [{"name": "item", "value": "item"}],
        }
        result = create_grouped_choices(groups)

        # Should only have separator and choice for "Full"
        separators = [r for r in result if isinstance(r, Separator)]
        assert len(separators) == 1

    def test_grouped_choices_indented_names(self):
        """Test that choice names are indented in groups."""
        groups = {"Category": [{"name": "Item", "value": "item"}]}
        result = create_grouped_choices(groups)

        choices = [r for r in result if isinstance(r, Choice)]
        assert len(choices) == 1
        assert choices[0].name.startswith("  ")

    def test_grouped_choices_no_defaults(self):
        """Test grouped choices without defaults."""
        groups = {
            "Group": [
                {"name": "A", "value": "a"},
                {"name": "B", "value": "b"},
            ]
        }
        result = create_grouped_choices(groups)

        choices = [r for r in result if isinstance(r, Choice)]
        assert all(c.enabled is False for c in choices)


class TestPromptsIntegration:
    """Integration tests for prompts module."""

    def test_multiple_prompt_types(self):
        """Test using multiple prompt types in sequence."""
        with patch("moai_adk.cli.ui.prompts.inquirer") as mock_inquirer:
            mock_fuzzy = MagicMock()
            mock_fuzzy.execute.return_value = ["A"]
            mock_inquirer.fuzzy.return_value = mock_fuzzy

            result1 = fuzzy_checkbox("Select", ["A", "B"])

            with patch("moai_adk.cli.ui.prompts.inquirer.select") as mock_select:
                mock_select_result = MagicMock()
                mock_select_result.execute.return_value = "A"
                mock_select.return_value = mock_select_result

                result2 = styled_select("Choose", ["A", "B"])

            assert result1 is not None
            assert result2 is not None

    def test_choice_processing_with_mixed_types(self):
        """Test processing choices with mixed types."""
        mixed_choices = [
            "Simple string",
            {"name": "Dict", "value": "dict"},
            Choice(value="choice_obj", name="Choice Object"),
            Separator("Category"),
        ]

        result = _process_choices(mixed_choices)
        assert len(result) == 4

    def test_grouped_choices_large_groups(self):
        """Test grouped choices with many items."""
        groups = {
            "Category A": [
                {"name": f"Item {i}", "value": f"item_{i}"} for i in range(10)
            ],
            "Category B": [
                {"name": f"Option {i}", "value": f"option_{i}"} for i in range(5)
            ],
        }

        result = create_grouped_choices(groups)
        choices = [r for r in result if isinstance(r, Choice)]
        assert len(choices) == 15
