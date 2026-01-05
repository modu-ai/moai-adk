"""
High-coverage unit tests for CLI UI prompt functions.

Tests focus on InquirerPy wrapper functions including fuzzy search,
select, and text input with proper error handling.
"""

from unittest.mock import MagicMock, patch

from moai_adk.cli.ui.prompts import (
    fuzzy_checkbox,
)


class TestFuzzyCheckbox:
    """Test fuzzy_checkbox prompt function."""

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_basic(self, mock_fuzzy):
        """Test basic fuzzy checkbox prompt."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1", "choice2"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2", "choice3"]

        # Act
        result = fuzzy_checkbox("Select items:", choices=choices)

        # Assert
        assert result == ["choice1", "choice2"]
        mock_fuzzy.assert_called_once()

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_with_dict_choices(self, mock_fuzzy):
        """Test fuzzy checkbox with dict choice objects."""
        # Arrange
        mock_execute = MagicMock(return_value=["val1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = [
            {"name": "Option 1", "value": "val1"},
            {"name": "Option 2", "value": "val2"},
        ]

        # Act
        result = fuzzy_checkbox("Select:", choices=choices)

        # Assert
        assert result == ["val1"]

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_with_default(self, mock_fuzzy):
        """Test fuzzy checkbox with default selections."""
        # Arrange
        mock_execute = MagicMock(return_value=["default1", "default2"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2", "choice3"]
        defaults = ["default1", "default2"]

        # Act
        result = fuzzy_checkbox("Select:", choices=choices, default=defaults)

        # Assert
        assert result == ["default1", "default2"]

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_with_custom_marker(self, mock_fuzzy):
        """Test fuzzy checkbox with custom marker."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]
        marker = "✓"

        # Act
        fuzzy_checkbox("Select:", choices=choices, marker=marker)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["marker"] == marker

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_no_border(self, mock_fuzzy):
        """Test fuzzy checkbox without border."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices, border=False)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["border"] is False

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_custom_height(self, mock_fuzzy):
        """Test fuzzy checkbox with custom height."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices, height=20)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["max_height"] == 20

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_returns_none_on_cancel(self, mock_fuzzy):
        """Test fuzzy checkbox returns None when cancelled."""
        # Arrange
        mock_fuzzy.return_value.execute = MagicMock(return_value=None)

        choices = ["choice1", "choice2"]

        # Act
        result = fuzzy_checkbox("Select:", choices=choices)

        # Assert
        assert result is None

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_with_validation(self, mock_fuzzy):
        """Test fuzzy checkbox with validation function."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2", "choice3"]
        def validate_fn(x):
            return len(x) > 0

        # Act
        fuzzy_checkbox("Select:", choices=choices, validate=validate_fn)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["validate"] == validate_fn

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_custom_invalid_message(self, mock_fuzzy):
        """Test fuzzy checkbox with custom invalid message."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]
        invalid_msg = "Please select at least 2 items"

        # Act
        fuzzy_checkbox("Select:", choices=choices, invalid_message=invalid_msg)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["invalid_message"] == invalid_msg

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_custom_instruction(self, mock_fuzzy):
        """Test fuzzy checkbox with custom instruction text."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]
        instruction = "[Space] Toggle  [Enter] Confirm"

        # Act
        fuzzy_checkbox("Select:", choices=choices, instruction=instruction)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["instruction"] == instruction

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_sets_multiselect(self, mock_fuzzy):
        """Test fuzzy checkbox sets multiselect parameter."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert "multiselect" in call_kwargs

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_processes_string_choices(self, mock_fuzzy):
        """Test fuzzy checkbox processes simple string choices."""
        # Arrange
        mock_execute = MagicMock(return_value=["item1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["item1", "item2", "item3"]

        # Act
        result = fuzzy_checkbox("Select:", choices=choices)

        # Assert
        assert result == ["item1"]
        # Verify choices were processed
        call_kwargs = mock_fuzzy.call_args[1]
        assert "choices" in call_kwargs

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_default_keybindings(self, mock_fuzzy):
        """Test fuzzy checkbox sets default keybindings."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert "keybindings" in call_kwargs
        keybindings = call_kwargs["keybindings"]
        assert "toggle" in keybindings
        assert "toggle-all" in keybindings

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_custom_keybindings(self, mock_fuzzy):
        """Test fuzzy checkbox with custom keybindings."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]
        custom_keybindings = {
            "toggle": [{"key": "space"}],
            "custom_action": [{"key": "ctrl_x"}],
        }

        # Act
        fuzzy_checkbox("Select:", choices=choices, keybindings=custom_keybindings)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["keybindings"] == custom_keybindings

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_with_marker_pl(self, mock_fuzzy):
        """Test fuzzy checkbox with marker placeholder."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices, marker_pl="○")

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["marker_pl"] == "○"

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_max_height_default(self, mock_fuzzy):
        """Test fuzzy checkbox uses max_height when no fixed height."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices, height=None, max_height=20)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert call_kwargs["max_height"] == 20

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_transformer(self, mock_fuzzy):
        """Test fuzzy checkbox includes transformer."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert "transformer" in call_kwargs
        transformer = call_kwargs["transformer"]
        # Test transformer function
        assert transformer(["a", "b"]) == "2 selected"

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_style_applied(self, mock_fuzzy):
        """Test fuzzy checkbox applies MOAI theme style."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert "style" in call_kwargs

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_qmark_set(self, mock_fuzzy):
        """Test fuzzy checkbox sets qmark symbol."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert "qmark" in call_kwargs

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_amark_set(self, mock_fuzzy):
        """Test fuzzy checkbox sets amark symbol."""
        # Arrange
        mock_execute = MagicMock(return_value=["choice1"])
        mock_fuzzy.return_value.execute = mock_execute

        choices = ["choice1", "choice2"]

        # Act
        fuzzy_checkbox("Select:", choices=choices)

        # Assert
        call_kwargs = mock_fuzzy.call_args[1]
        assert "amark" in call_kwargs

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_multiple_returns(self, mock_fuzzy):
        """Test fuzzy checkbox returns multiple selected items."""
        # Arrange
        selected = ["file1.py", "file2.py", "file3.py"]
        mock_fuzzy.return_value.execute = MagicMock(return_value=selected)

        choices = [
            {"name": "File 1", "value": "file1.py"},
            {"name": "File 2", "value": "file2.py"},
            {"name": "File 3", "value": "file3.py"},
            {"name": "File 4", "value": "file4.py"},
        ]

        # Act
        result = fuzzy_checkbox("Select files:", choices=choices)

        # Assert
        assert result == selected
        assert len(result) == 3

    @patch("moai_adk.cli.ui.prompts.inquirer.fuzzy")
    def test_fuzzy_checkbox_empty_selection(self, mock_fuzzy):
        """Test fuzzy checkbox with empty selection."""
        # Arrange
        mock_fuzzy.return_value.execute = MagicMock(return_value=[])

        choices = ["choice1", "choice2"]

        # Act
        result = fuzzy_checkbox("Select:", choices=choices)

        # Assert
        assert result == []
