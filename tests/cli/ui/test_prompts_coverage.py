"""Additional coverage tests for CLI UI prompts.

Tests for lines not covered by existing tests.
"""

from unittest.mock import MagicMock, patch

from moai_adk.cli.ui.prompts import styled_password, styled_select


class TestStyledSelectFallbackToQuestionary:
    """Test styled_select fallback to questionary on errors."""

    def test_styled_select_fallback_on_keyboard_interrupt(self):
        """Should fallback to questionary when InquirerPy raises KeyboardInterrupt."""
        choices = ["Option 1", "Option 2", "Option 3"]

        # Make inquirer.select().execute() raise KeyboardInterrupt to trigger fallback
        with patch("moai_adk.cli.ui.prompts.inquirer") as mock_inquirer:
            mock_inquirer.select.return_value.execute.side_effect = KeyboardInterrupt()

            # Patch builtins.__import__ to intercept questionary import
            import builtins

            original_import = builtins.__import__

            def mock_import(name, *args, **kwargs):
                if name == "questionary":
                    mock_module = MagicMock()
                    mock_module.select.return_value.ask.return_value = "Option 1"
                    return mock_module
                return original_import(name, *args, **kwargs)

            with patch("builtins.__import__", side_effect=mock_import):
                result = styled_select("Select an option", choices=choices)

                # Should have called questionary as fallback
                assert result == "Option 1"

    def test_styled_select_fallback_with_dict_choices(self):
        """Should fallback to questionary with dict choices and map values correctly."""
        choices = [
            {"name": "Option 1", "value": "opt1"},
            {"name": "Option 2", "value": "opt2"},
        ]

        with patch("moai_adk.cli.ui.prompts.inquirer") as mock_inquirer:
            mock_inquirer.select.return_value.execute.side_effect = OSError("Terminal error")

            import builtins

            original_import = builtins.__import__

            def mock_import(name, *args, **kwargs):
                if name == "questionary":
                    mock_module = MagicMock()
                    mock_module.select.return_value.ask.return_value = "Option 1"
                    return mock_module
                return original_import(name, *args, **kwargs)

            with patch("builtins.__import__", side_effect=mock_import):
                result = styled_select("Select an option", choices=choices)

                # Should map back to value
                assert result == "opt1"

    def test_styled_select_fallback_with_default_value(self):
        """Should fallback to questionary with default value."""
        choices = [
            {"name": "Option 1", "value": "opt1"},
            {"name": "Option 2", "value": "opt2"},
        ]

        with patch("moai_adk.cli.ui.prompts.inquirer") as mock_inquirer:
            mock_inquirer.select.return_value.execute.side_effect = Exception("Error")

            import builtins

            original_import = builtins.__import__

            def mock_import(name, *args, **kwargs):
                if name == "questionary":
                    mock_module = MagicMock()
                    mock_module.select.return_value.ask.return_value = "Option 2"
                    return mock_module
                return original_import(name, *args, **kwargs)

            with patch("builtins.__import__", side_effect=mock_import):
                result = styled_select("Select an option", choices=choices, default="opt2")

                # Should return the mapped value
                assert result == "opt2"

    def test_styled_select_fallback_when_user_cancels(self):
        """Should return None when user cancels in questionary fallback."""
        choices = ["Option 1", "Option 2"]

        with patch("moai_adk.cli.ui.prompts.inquirer") as mock_inquirer:
            mock_inquirer.select.return_value.execute.side_effect = OSError()

            import builtins

            original_import = builtins.__import__

            def mock_import(name, *args, **kwargs):
                if name == "questionary":
                    mock_module = MagicMock()
                    mock_module.select.return_value.ask.return_value = None
                    return mock_module
                return original_import(name, *args, **kwargs)

            with patch("builtins.__import__", side_effect=mock_import):
                result = styled_select("Select an option", choices=choices)

                assert result is None


class TestStyledPassword:
    """Test styled_password function."""

    def test_styled_password_success(self):
        """Should return password when entered successfully."""
        mock_inquirer = MagicMock()
        mock_inquirer.secret.return_value.execute.return_value = "secure_password"

        with patch("moai_adk.cli.ui.prompts.inquirer", mock_inquirer):
            result = styled_password("Enter password")

            assert result == "secure_password"

    def test_styled_password_keyboard_interrupt(self):
        """Should return None when user presses Ctrl+C."""
        mock_inquirer = MagicMock()
        mock_inquirer.secret.return_value.execute.side_effect = KeyboardInterrupt()

        with patch("moai_adk.cli.ui.prompts.inquirer", mock_inquirer):
            result = styled_password("Enter password")

            assert result is None
