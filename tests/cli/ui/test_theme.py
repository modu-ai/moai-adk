"""Tests for CLI theme configuration."""

from unittest.mock import patch

from moai_adk.cli.ui.theme import (
    MOAI_COLORS,
    MOAI_THEME,
    SPINNER_FRAMES,
    SYMBOLS,
    get_category_separator,
    get_styled_choice,
    print_styled,
)


class TestThemeConstants:
    """Test theme constants."""

    def test_moai_colors_has_required_keys(self):
        """Test MOAI_COLORS has all required color keys."""
        required_keys = [
            "primary",
            "secondary",
            "accent",
            "info",
            "warning",
            "error",
            "muted",
            "text",
            "claude_terra",
            "claude_black",
        ]
        for key in required_keys:
            assert key in MOAI_COLORS
            assert isinstance(MOAI_COLORS[key], str)

    def test_moai_theme_has_required_keys(self):
        """Test MOAI_THEME has all required style keys."""
        required_keys = [
            "questionmark",
            "question",
            "answered_question",
            "answer",
            "input",
            "instruction",
            "pointer",
            "checkbox",
            "selected",
            "separator",
            "fuzzy_prompt",
            "fuzzy_info",
            "fuzzy_border",
            "fuzzy_match",
            "validator",
            "marker",
            "long_instruction",
            "skipped",
        ]
        for key in required_keys:
            assert key in MOAI_THEME
            assert isinstance(MOAI_THEME[key], str)

    def test_symbols(self):
        """Test SYMBOLS dictionary."""
        assert "checkbox_selected" in SYMBOLS
        assert "checkbox_unselected" in SYMBOLS
        assert "pointer" in SYMBOLS
        assert "success" in SYMBOLS
        assert "error" in SYMBOLS
        assert "warning" in SYMBOLS
        assert "info" in SYMBOLS

    def test_spinner_frames(self):
        """Test SPINNER_FRAMES list."""
        assert isinstance(SPINNER_FRAMES, list)
        assert len(SPINNER_FRAMES) == 10
        for frame in SPINNER_FRAMES:
            assert isinstance(frame, str)


class TestGetStyledChoice:
    """Test get_styled_choice function."""

    def test_basic_choice(self):
        """Test creating a basic choice."""
        choice = get_styled_choice("Test Title", "test_value")
        assert choice["name"] == "Test Title"
        assert choice["value"] == "test_value"
        assert choice["enabled"] is False

    def test_enabled_choice(self):
        """Test creating an enabled choice."""
        choice = get_styled_choice("Test Title", "test_value", enabled=True)
        assert choice["enabled"] is True

    def test_choice_with_description(self):
        """Test creating a choice with description."""
        choice = get_styled_choice("Test Title", "test_value", description="This is a test description")
        assert "Test Title" in choice["name"]
        assert "This is a test description" in choice["name"]
        assert choice["value"] == "test_value"


class TestPrintStyled:
    """Test print_styled function."""

    @patch("moai_adk.cli.ui.theme._console")
    def test_print_styled_basic(self, mock_console):
        """Test basic styled print."""
        print_styled("Test message", style="primary")
        mock_console.print.assert_called_once()

    @patch("moai_adk.cli.ui.theme._console")
    def test_print_styled_bold(self, mock_console):
        """Test styled print with bold."""
        print_styled("Test message", style="secondary", bold=True)
        mock_console.print.assert_called_once()

    @patch("moai_adk.cli.ui.theme._console")
    def test_print_styled_no_newline(self, mock_console):
        """Test styled print without newline."""
        print_styled("Test message", style="info", newline=False)
        mock_console.print.assert_called_once()
        # Check that end parameter is "\n" when newline=True and "" when False
        call_args = mock_console.print.call_args
        assert call_args[1]["end"] == ""

    @patch("moai_adk.cli.ui.theme._console")
    def test_print_styled_unknown_style(self, mock_console):
        """Test styled print with unknown style (defaults to text)."""
        # Should not raise exception
        print_styled("Test message", style="unknown_style")
        mock_console.print.assert_called_once()

    @patch("moai_adk.cli.ui.theme._console")
    def test_print_styled_with_newline(self, mock_console):
        """Test styled print with newline (default)."""
        print_styled("Test message", style="warning", newline=True)
        mock_console.print.assert_called_once()
        call_args = mock_console.print.call_args
        assert call_args[1]["end"] == "\n"


class TestGetCategorySeparator:
    """Test get_category_separator function."""

    def test_category_separator(self):
        """Test creating a category separator."""
        separator = get_category_separator("Test Category")
        assert separator["value"] is None
        assert "Test Category" in separator["name"]
        assert "──" in separator["name"]
        assert separator["disabled"] is True
