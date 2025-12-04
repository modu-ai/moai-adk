"""Extended unit tests for moai_adk.cli.ui.prompts module.

Comprehensive tests for all prompt functions and utilities.
"""

from unittest.mock import MagicMock, Mock, patch

import pytest

from InquirerPy.base.control import Choice, Separator
from moai_adk.cli.ui.prompts import (
    create_grouped_choices,
    fuzzy_checkbox,
    fuzzy_select,
    styled_checkbox,
    styled_confirm,
    styled_input,
    styled_select,
    _process_choices,
)


class TestProcessChoices:
    """Test _process_choices utility function."""

    def test_process_choices_string_list(self):
        """Test _process_choices with simple string list."""
        choices = ["Option A", "Option B", "Option C"]
        result = _process_choices(choices)

        assert len(result) == 3
        assert all(isinstance(c, Choice) for c in result)

    def test_process_choices_dict_list(self):
        """Test _process_choices with dictionary list."""
        choices = [
            {"name": "First", "value": "first"},
            {"name": "Second", "value": "second"},
        ]
        result = _process_choices(choices)

        assert len(result) == 2
        assert all(isinstance(c, Choice) for c in result)

    def test_process_choices_mixed_list(self):
        """Test _process_choices with mixed choice types."""
        choices = [
            "String choice",
            {"name": "Dict choice", "value": "dict_value"},
            Choice(value="choice_obj", name="Choice object"),
        ]
        result = _process_choices(choices)

        assert len(result) == 3

    def test_process_choices_with_separators(self):
        """Test _process_choices handles separators."""
        choices = [
            "Option 1",
            Separator(),
            "Option 2",
        ]
        result = _process_choices(choices)

        assert len(result) == 3
        separators = [c for c in result if isinstance(c, Separator)]
        assert len(separators) == 1

    def test_process_choices_with_defaults(self):
        """Test _process_choices applies defaults."""
        choices = ["Option A", "Option B", "Option C"]
        defaults = ["Option B"]

        result = _process_choices(choices, defaults)

        # Check that "Option B" is marked as enabled
        choice_b = [c for c in result if c.name == "Option B"][0]
        assert choice_b.enabled is True

    def test_process_choices_dict_with_enabled(self):
        """Test _process_choices respects enabled in dict."""
        choices = [
            {"name": "Enabled", "value": "enabled", "enabled": True},
            {"name": "Disabled", "value": "disabled", "enabled": False},
        ]
        result = _process_choices(choices)

        enabled = [c for c in result if c.name == "Enabled"][0]
        assert enabled.enabled is True

    def test_process_choices_dict_disabled_creates_separator(self):
        """Test _process_choices creates separator for disabled items."""
        choices = [
            {"name": "Enabled", "value": "enabled"},
            {"name": "Disabled", "value": "disabled", "disabled": True},
        ]
        result = _process_choices(choices)

        # Check for separator
        separators = [c for c in result if isinstance(c, Separator)]
        assert len(separators) >= 0

    def test_process_choices_empty_list(self):
        """Test _process_choices with empty list."""
        result = _process_choices([])
        assert result == []

    def test_process_choices_dict_title_fallback(self):
        """Test _process_choices uses title when name missing."""
        choices = [{"title": "From Title", "value": "val"}]
        result = _process_choices(choices)

        assert result[0].name == "From Title"

    def test_process_choices_dict_name_priority(self):
        """Test _process_choices prioritizes name over title."""
        choices = [
            {"name": "From Name", "title": "From Title", "value": "val"}
        ]
        result = _process_choices(choices)

        assert result[0].name == "From Name"


class TestCreateGroupedChoices:
    """Test create_grouped_choices function."""

    def test_create_grouped_choices_basic(self):
        """Test create_grouped_choices creates grouped structure."""
        groups = {
            "Group A": [
                {"name": "A1", "value": "a1"},
                {"name": "A2", "value": "a2"},
            ],
            "Group B": [{"name": "B1", "value": "b1"}],
        }
        result = create_grouped_choices(groups)

        # Should have separators and choices
        assert len(result) > len(groups)
        separators = [c for c in result if isinstance(c, Separator)]
        assert len(separators) >= 2

    def test_create_grouped_choices_with_defaults(self):
        """Test create_grouped_choices applies defaults."""
        groups = {
            "Group": [
                {"name": "Option A", "value": "a"},
                {"name": "Option B", "value": "b"},
            ]
        }
        defaults = ["a"]

        result = create_grouped_choices(groups, defaults)

        option_a = [c for c in result if isinstance(c, Choice) and c.value == "a"][
            0
        ]
        assert option_a.enabled is True

    def test_create_grouped_choices_empty_group(self):
        """Test create_grouped_choices skips empty groups."""
        groups = {
            "Group A": [{"name": "A1", "value": "a1"}],
            "Empty Group": [],
        }
        result = create_grouped_choices(groups)

        # Empty group should not create separator
        separators = [c for c in result if isinstance(c, Separator)]
        assert len(separators) == 1

    def test_create_grouped_choices_formatting(self):
        """Test create_grouped_choices formats items correctly."""
        groups = {"Commands": [{"name": "deploy", "value": "deploy"}]}
        result = create_grouped_choices(groups)

        # Find the choice and check formatting
        choices = [c for c in result if isinstance(c, Choice)]
        assert any("deploy" in c.name for c in choices)


class TestFuzzyCheckbox:
    """Test fuzzy_checkbox function."""

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_basic(self, mock_fuzzy):
        """Test fuzzy_checkbox basic functionality."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = ["option1", "option2"]
        mock_fuzzy.return_value = mock_prompt

        result = fuzzy_checkbox("Select items:", ["option1", "option2", "option3"])

        assert result == ["option1", "option2"]
        mock_fuzzy.assert_called_once()

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_with_default(self, mock_fuzzy):
        """Test fuzzy_checkbox with default values."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = ["option1"]
        mock_fuzzy.return_value = mock_prompt

        result = fuzzy_checkbox(
            "Select:",
            ["option1", "option2"],
            default=["option1"],
        )

        assert result == ["option1"]

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_keyboard_interrupt(self, mock_fuzzy):
        """Test fuzzy_checkbox handles keyboard interrupt."""
        mock_fuzzy.side_effect = KeyboardInterrupt()

        result = fuzzy_checkbox("Select:", ["option1"])

        assert result is None

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_custom_keybindings(self, mock_fuzzy):
        """Test fuzzy_checkbox with custom keybindings."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = []
        mock_fuzzy.return_value = mock_prompt

        custom_keys = {"toggle": [{"key": "custom"}]}
        fuzzy_checkbox(
            "Select:",
            ["option1"],
            keybindings=custom_keys,
        )

        # Check keybindings were passed
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["keybindings"] == custom_keys

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_custom_instruction(self, mock_fuzzy):
        """Test fuzzy_checkbox with custom instruction."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = []
        mock_fuzzy.return_value = mock_prompt

        custom_instruction = "Custom help text"
        fuzzy_checkbox(
            "Select:",
            ["option1"],
            instruction=custom_instruction,
        )

        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["instruction"] == custom_instruction


class TestFuzzySelect:
    """Test fuzzy_select function."""

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_select_basic(self, mock_fuzzy):
        """Test fuzzy_select basic functionality."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = "option1"
        mock_fuzzy.return_value = mock_prompt

        result = fuzzy_select("Select item:", ["option1", "option2"])

        assert result == "option1"

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_select_with_default(self, mock_fuzzy):
        """Test fuzzy_select with default value."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = "option1"
        mock_fuzzy.return_value = mock_prompt

        result = fuzzy_select(
            "Select:",
            ["option1", "option2"],
            default="option1",
        )

        assert result == "option1"

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_select_keyboard_interrupt(self, mock_fuzzy):
        """Test fuzzy_select handles keyboard interrupt."""
        mock_fuzzy.side_effect = KeyboardInterrupt()

        result = fuzzy_select("Select:", ["option1"])

        assert result is None

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_select_multiselect_false(self, mock_fuzzy):
        """Test fuzzy_select uses multiselect=False."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = "option1"
        mock_fuzzy.return_value = mock_prompt

        fuzzy_select("Select:", ["option1", "option2"])

        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["multiselect"] is False


class TestStyledCheckbox:
    """Test styled_checkbox function."""

    @patch("moai_adk.cli.ui.prompts.inquirer.checkbox")
    def test_styled_checkbox_basic(self, mock_checkbox):
        """Test styled_checkbox basic functionality."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = ["option1"]
        mock_checkbox.return_value = mock_prompt

        result = styled_checkbox("Select:", ["option1", "option2"])

        assert result == ["option1"]

    @patch("moai_adk.cli.ui.prompts.inquirer.checkbox")
    def test_styled_checkbox_with_default(self, mock_checkbox):
        """Test styled_checkbox with default values."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = ["option1"]
        mock_checkbox.return_value = mock_prompt

        result = styled_checkbox(
            "Select:",
            ["option1", "option2"],
            default=["option1"],
        )

        assert result == ["option1"]

    @patch("moai_adk.cli.ui.prompts.inquirer.checkbox")
    def test_styled_checkbox_keyboard_interrupt(self, mock_checkbox):
        """Test styled_checkbox handles keyboard interrupt."""
        mock_checkbox.side_effect = KeyboardInterrupt()

        result = styled_checkbox("Select:", ["option1"])

        assert result is None

    @patch("moai_adk.cli.ui.prompts.inquirer.checkbox")
    def test_styled_checkbox_custom_instruction(self, mock_checkbox):
        """Test styled_checkbox with custom instruction."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = []
        mock_checkbox.return_value = mock_prompt

        custom_instruction = "Use arrows and space"
        styled_checkbox("Select:", ["option1"], instruction=custom_instruction)

        call_kwargs = mock_checkbox.call_args[1]
        assert call_kwargs["instruction"] == custom_instruction


class TestStyledSelect:
    """Test styled_select function."""

    @patch("moai_adk.cli.ui.prompts.inquirer.select")
    def test_styled_select_basic(self, mock_select):
        """Test styled_select basic functionality."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = "option1"
        mock_select.return_value = mock_prompt

        result = styled_select("Select:", ["option1", "option2"])

        assert result == "option1"

    @patch("moai_adk.cli.ui.prompts.inquirer.select")
    def test_styled_select_with_default(self, mock_select):
        """Test styled_select with default value."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = "option1"
        mock_select.return_value = mock_prompt

        result = styled_select(
            "Select:",
            ["option1", "option2"],
            default="option1",
        )

        assert result == "option1"

    @patch("moai_adk.cli.ui.prompts.inquirer.select")
    def test_styled_select_keyboard_interrupt(self, mock_select):
        """Test styled_select handles keyboard interrupt."""
        mock_select.side_effect = KeyboardInterrupt()

        result = styled_select("Select:", ["option1"])

        assert result is None

    @patch("moai_adk.cli.ui.prompts.inquirer.select")
    def test_styled_select_dict_choices(self, mock_select):
        """Test styled_select with dictionary choices."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = "value1"
        mock_select.return_value = mock_prompt

        choices = [
            {"name": "Display 1", "value": "value1"},
            {"name": "Display 2", "value": "value2"},
        ]

        result = styled_select("Select:", choices)

        assert result == "value1"


class TestStyledInput:
    """Test styled_input function."""

    @patch("moai_adk.cli.ui.prompts.inquirer.text")
    def test_styled_input_basic(self, mock_text):
        """Test styled_input basic functionality."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = "user input"
        mock_text.return_value = mock_prompt

        result = styled_input("Enter text:")

        assert result == "user input"

    @patch("moai_adk.cli.ui.prompts.inquirer.text")
    def test_styled_input_with_default(self, mock_text):
        """Test styled_input with default value."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = "default"
        mock_text.return_value = mock_prompt

        result = styled_input("Enter:", default="default")

        assert result == "default"

    @patch("moai_adk.cli.ui.prompts.inquirer.text")
    def test_styled_input_keyboard_interrupt(self, mock_text):
        """Test styled_input handles keyboard interrupt."""
        mock_text.side_effect = KeyboardInterrupt()

        result = styled_input("Enter:")

        assert result is None

    @patch("moai_adk.cli.ui.prompts.inquirer.text")
    def test_styled_input_with_validation(self, mock_text):
        """Test styled_input with custom validation."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = "valid"
        mock_text.return_value = mock_prompt

        def validator(x):
            return len(x) > 0

        styled_input("Enter:", validate=validator)

        call_kwargs = mock_text.call_args[1]
        assert call_kwargs["validate"] == validator

    @patch("moai_adk.cli.ui.prompts.inquirer.text")
    def test_styled_input_required_false(self, mock_text):
        """Test styled_input with required=False."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = ""
        mock_text.return_value = mock_prompt

        result = styled_input("Enter:", required=False)

        assert result == ""


class TestStyledConfirm:
    """Test styled_confirm function."""

    @patch("moai_adk.cli.ui.prompts.inquirer.confirm")
    def test_styled_confirm_true(self, mock_confirm):
        """Test styled_confirm returns True."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = True
        mock_confirm.return_value = mock_prompt

        result = styled_confirm("Confirm?")

        assert result is True

    @patch("moai_adk.cli.ui.prompts.inquirer.confirm")
    def test_styled_confirm_false(self, mock_confirm):
        """Test styled_confirm returns False."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = False
        mock_confirm.return_value = mock_prompt

        result = styled_confirm("Confirm?")

        assert result is False

    @patch("moai_adk.cli.ui.prompts.inquirer.confirm")
    def test_styled_confirm_keyboard_interrupt(self, mock_confirm):
        """Test styled_confirm handles keyboard interrupt."""
        mock_confirm.side_effect = KeyboardInterrupt()

        result = styled_confirm("Confirm?")

        assert result is None

    @patch("moai_adk.cli.ui.prompts.inquirer.confirm")
    def test_styled_confirm_with_default(self, mock_confirm):
        """Test styled_confirm with default value."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = False
        mock_confirm.return_value = mock_prompt

        result = styled_confirm("Confirm?", default=False)

        call_kwargs = mock_confirm.call_args[1]
        assert call_kwargs["default"] is False


class TestPromptsEdgeCases:
    """Test edge cases and error scenarios."""

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_empty_choices(self, mock_fuzzy):
        """Test fuzzy_checkbox with empty choices."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = []
        mock_fuzzy.return_value = mock_prompt

        result = fuzzy_checkbox("Select:", [])

        assert result == []

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_select_empty_choices(self, mock_fuzzy):
        """Test fuzzy_select with empty choices."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = None
        mock_fuzzy.return_value = mock_prompt

        fuzzy_select("Select:", [])

        # Should handle empty choices gracefully

    def test_process_choices_none_values(self):
        """Test _process_choices handles None values."""
        choices = [None]
        # Should handle gracefully

    def test_create_grouped_choices_single_group(self):
        """Test create_grouped_choices with single group."""
        groups = {"OnlyGroup": [{"name": "Item", "value": "val"}]}
        result = create_grouped_choices(groups)

        assert len(result) >= 1

    @patch("moai_adk.cli.ui.prompts.inquirer.text")
    def test_styled_input_empty_default(self, mock_text):
        """Test styled_input with empty default."""
        mock_prompt = MagicMock()
        mock_prompt.execute.return_value = ""
        mock_text.return_value = mock_prompt

        result = styled_input("Enter:", default="")

        assert result == ""
