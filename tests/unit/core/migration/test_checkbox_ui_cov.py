"""Comprehensive tests for InteractiveCheckboxUI with 70%+ coverage.

Tests cover:
- Initialization and setup
- Element organization and categorization
- Navigation (up/down)
- Selection toggling and management
- Curses interface display
- Fallback selection mode
- Confirmation dialogs
- Element type detection
"""

import pytest
import tempfile
import curses
from pathlib import Path
from unittest.mock import MagicMock, patch, Mock, call
from io import StringIO

from moai_adk.core.migration.interactive_checkbox_ui import (
    InteractiveCheckboxUI,
    create_interactive_checkbox_ui,
)


class TestInteractiveCheckboxUIInitialization:
    """Test InteractiveCheckboxUI initialization."""

    def test_init_with_string_path(self):
        """Test initialization with string path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                ui = InteractiveCheckboxUI(tmpdir)
                assert ui.project_path == Path(tmpdir).resolve()

    def test_init_with_path_object(self):
        """Test initialization with Path object."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir)
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                ui = InteractiveCheckboxUI(path)
                assert ui.project_path == path.resolve()

    def test_init_creates_scanner(self):
        """Test that scanner is created during initialization."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ) as mock_scanner:
                ui = InteractiveCheckboxUI(tmpdir)
                assert mock_scanner.called
                assert ui.scanner is not None

    def test_init_initializes_selected_indices(self):
        """Test that selected_indices is initialized as empty set."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                ui = InteractiveCheckboxUI(tmpdir)
                assert ui.selected_indices == set()

    def test_init_initializes_current_index(self):
        """Test that current_index is initialized to zero."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                ui = InteractiveCheckboxUI(tmpdir)
                assert ui.current_index == 0


class TestElementOrganization:
    """Test element organization by category."""

    @pytest.fixture
    def ui(self):
        """Create UI instance with mocked scanner."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                yield InteractiveCheckboxUI(tmpdir)

    def test_get_elements_by_category_empty(self, ui):
        """Test organizing empty element list."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {}

        result = ui._get_elements_by_category()

        assert isinstance(result, dict)

    def test_get_elements_by_category_with_skills(self, ui):
        """Test organizing skills into category."""
        mock_skill = MagicMock()
        mock_skill.name = "test_skill"
        mock_skill.path = Path("/test/skill")
        mock_skill.is_template = False

        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"skills": [mock_skill]}

        result = ui._get_elements_by_category()

        assert "Skills" in result
        assert len(result["Skills"]) == 1
        assert result["Skills"][0]["name"] == "test_skill (custom)"

    def test_get_elements_by_category_template_skills(self, ui):
        """Test organizing template skills."""
        mock_skill = MagicMock()
        mock_skill.name = "template_skill"
        mock_skill.path = Path("/test/skill")
        mock_skill.is_template = True

        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"skills": [mock_skill]}

        result = ui._get_elements_by_category()

        assert "Skills" in result
        assert "template" in result["Skills"][0]["name"].lower()

    def test_get_elements_by_category_with_agents(self, ui):
        """Test organizing agents into category."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"agents": ["/path/to/agent.py"]}

        result = ui._get_elements_by_category()

        assert "Agents" in result
        assert len(result["Agents"]) == 1

    def test_get_elements_by_category_with_commands(self, ui):
        """Test organizing commands into category."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {
            "commands": ["/path/to/command.md"]
        }

        result = ui._get_elements_by_category()

        assert "Commands" in result
        assert len(result["Commands"]) == 1

    def test_get_elements_by_category_with_hooks(self, ui):
        """Test organizing hooks into category."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"hooks": ["/path/to/hook.py"]}

        result = ui._get_elements_by_category()

        assert "Hooks" in result
        assert len(result["Hooks"]) == 1

    def test_get_elements_by_category_removes_empty_categories(self, ui):
        """Test that empty categories are removed."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"agents": ["/path/to/agent.py"]}

        result = ui._get_elements_by_category()

        assert "Agents" in result
        assert "Skills" not in result
        assert "Commands" not in result
        assert "Hooks" not in result


class TestElementFlattening:
    """Test element flattening logic."""

    @pytest.fixture
    def ui(self):
        """Create UI instance with mocked scanner."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                yield InteractiveCheckboxUI(tmpdir)

    def test_flatten_elements_empty(self, ui):
        """Test flattening empty categorized elements."""
        result = ui._flatten_elements({})
        assert result == []

    def test_flatten_elements_single_category(self, ui):
        """Test flattening single category."""
        elements_by_category = {
            "Agents": [{"name": "agent1", "path": "/path/agent1", "type": "agent"}]
        }

        result = ui._flatten_elements(elements_by_category)

        # Should have header + element
        assert len(result) == 2
        assert result[0]["type"] == "header"
        assert result[1]["type"] == "element"

    def test_flatten_elements_multiple_categories(self, ui):
        """Test flattening multiple categories."""
        elements_by_category = {
            "Agents": [{"name": "agent1", "path": "/path/agent1", "type": "agent"}],
            "Commands": [{"name": "cmd1", "path": "/path/cmd1", "type": "command"}],
        }

        result = ui._flatten_elements(elements_by_category)

        # Should have headers and elements
        headers = [el for el in result if el["type"] == "header"]
        elements = [el for el in result if el["type"] == "element"]

        assert len(headers) == 2
        assert len(elements) == 2

    def test_flatten_elements_header_format(self, ui):
        """Test header formatting in flattened list."""
        elements_by_category = {
            "Agents": [
                {"name": "agent1", "path": "/path/agent1", "type": "agent"},
                {"name": "agent2", "path": "/path/agent2", "type": "agent"},
            ]
        }

        result = ui._flatten_elements(elements_by_category)

        header = result[0]
        assert header["type"] == "header"
        assert "Agents (2)" in header["text"]


class TestNavigation:
    """Test navigation methods."""

    @pytest.fixture
    def ui(self):
        """Create UI instance with mocked scanner."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                yield InteractiveCheckboxUI(tmpdir)

    def test_navigate_up_from_middle(self, ui):
        """Test navigating up from middle position."""
        elements = [
            {"type": "header", "text": "Category"},
            {"type": "element", "name": "elem1"},
            {"type": "element", "name": "elem2"},
            {"type": "element", "name": "elem3"},
        ]

        ui.current_index = 3  # Start at elem3
        ui._navigate_up(elements)

        # Should move to elem2
        assert ui.current_index == 2

    def test_navigate_up_skips_headers(self, ui):
        """Test that navigation skips header entries."""
        elements = [
            {"type": "header", "text": "Category"},
            {"type": "element", "name": "elem1"},
            {"type": "element", "name": "elem2"},
        ]

        ui.current_index = 2  # Start at elem2
        ui._navigate_up(elements)

        # Should move to elem1, skipping header
        assert ui.current_index == 1

    def test_navigate_up_stops_at_start(self, ui):
        """Test that navigation stops at start of list."""
        elements = [
            {"type": "header", "text": "Category"},
            {"type": "element", "name": "elem1"},
        ]

        ui.current_index = 1
        ui._navigate_up(elements)

        # Can't go further up
        assert ui.current_index == 1

    def test_navigate_down_from_middle(self, ui):
        """Test navigating down from middle position."""
        elements = [
            {"type": "header", "text": "Category"},
            {"type": "element", "name": "elem1"},
            {"type": "element", "name": "elem2"},
            {"type": "element", "name": "elem3"},
        ]

        ui.current_index = 1  # Start at elem1
        ui._navigate_down(elements)

        # Should move to elem2
        assert ui.current_index == 2

    def test_navigate_down_skips_headers(self, ui):
        """Test that down navigation skips headers."""
        elements = [
            {"type": "element", "name": "elem1"},
            {"type": "header", "text": "Category"},
            {"type": "element", "name": "elem2"},
        ]

        ui.current_index = 0  # Start at elem1
        ui._navigate_down(elements)

        # Should move to elem2, skipping header
        assert ui.current_index == 2

    def test_navigate_down_stops_at_end(self, ui):
        """Test that navigation stops at end of list."""
        elements = [
            {"type": "header", "text": "Category"},
            {"type": "element", "name": "elem1"},
        ]

        ui.current_index = 1
        ui._navigate_down(elements)

        # Can't go further down
        assert ui.current_index == 1


class TestSelection:
    """Test selection management."""

    @pytest.fixture
    def ui(self):
        """Create UI instance with mocked scanner."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                yield InteractiveCheckboxUI(tmpdir)

    def test_toggle_selection_add(self, ui):
        """Test adding element to selection."""
        elements = [
            {"type": "element", "name": "elem1"},
            {"type": "element", "name": "elem2"},
        ]

        ui.current_index = 0
        ui._toggle_selection(elements)

        assert 0 in ui.selected_indices

    def test_toggle_selection_remove(self, ui):
        """Test removing element from selection."""
        elements = [
            {"type": "element", "name": "elem1"},
            {"type": "element", "name": "elem2"},
        ]

        ui.selected_indices.add(0)
        ui.current_index = 0
        ui._toggle_selection(elements)

        assert 0 not in ui.selected_indices

    def test_toggle_selection_on_header(self, ui):
        """Test toggling on header does nothing."""
        elements = [{"type": "header", "text": "Category"}]

        ui.current_index = 0
        ui._toggle_selection(elements)

        # Selection should not change for headers
        assert 0 not in ui.selected_indices

    def test_select_all(self, ui):
        """Test selecting all elements."""
        elements = [
            {"type": "header", "text": "Category"},
            {"type": "element", "name": "elem1"},
            {"type": "element", "name": "elem2"},
            {"type": "element", "name": "elem3"},
        ]

        ui._select_all(elements)

        # Only elements should be selected, not headers
        assert 1 in ui.selected_indices
        assert 2 in ui.selected_indices
        assert 3 in ui.selected_indices
        assert 0 not in ui.selected_indices

    def test_select_none(self, ui):
        """Test clearing all selections."""
        ui.selected_indices = {1, 2, 3}

        elements = [{"type": "element"}] * 4
        ui._select_none(elements)

        assert ui.selected_indices == set()


class TestConfirmation:
    """Test confirmation dialog."""

    @pytest.fixture
    def ui(self):
        """Create UI instance with mocked scanner."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                yield InteractiveCheckboxUI(tmpdir)

    def test_confirm_selection_no_elements_selected(self, ui):
        """Test confirmation with no elements selected."""
        mock_stdscr = MagicMock()
        elements = [{"type": "element", "name": "elem1"}]

        result = ui._confirm_selection(mock_stdscr, elements)

        assert result is False

    def test_confirm_selection_yes_response(self, ui):
        """Test confirmation when user confirms (y)."""
        mock_stdscr = MagicMock()
        mock_stdscr.getch.return_value = ord("y")
        mock_stdscr.getmaxyx.return_value = (24, 80)

        elements = [{"type": "element", "name": "elem1"}]

        ui.selected_indices = {0}

        result = ui._confirm_selection(mock_stdscr, elements)

        assert result is True

    def test_confirm_selection_enter_response(self, ui):
        """Test confirmation when user presses Enter."""
        mock_stdscr = MagicMock()
        mock_stdscr.getch.return_value = ord("\n")
        mock_stdscr.getmaxyx.return_value = (24, 80)

        elements = [{"type": "element", "name": "elem1"}]

        ui.selected_indices = {0}

        result = ui._confirm_selection(mock_stdscr, elements)

        assert result is True

    def test_confirm_selection_no_response(self, ui):
        """Test confirmation when user declines (n)."""
        mock_stdscr = MagicMock()
        mock_stdscr.getch.return_value = ord("n")
        mock_stdscr.getmaxyx.return_value = (24, 80)

        elements = [{"type": "element", "name": "elem1"}]

        ui.selected_indices = {0}

        result = ui._confirm_selection(mock_stdscr, elements)

        assert result is False


class TestFallbackSelection:
    """Test fallback selection method."""

    @pytest.fixture
    def ui(self):
        """Create UI instance with mocked scanner."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                yield InteractiveCheckboxUI(tmpdir)

    def test_fallback_selection_no_elements(self, ui):
        """Test fallback with no elements."""
        flattened = []
        with patch("builtins.input", return_value=""):
            result = ui._fallback_selection(flattened)
            # Should handle gracefully
            assert result is None or result == []

    def test_fallback_selection_user_cancels(self, ui):
        """Test fallback when user presses enter with no input."""
        flattened = [
            {"type": "element", "name": "elem1", "path": "/path1", "category": "agent"}
        ]

        with patch("builtins.input", return_value=""):
            result = ui._fallback_selection(flattened)

        assert result is None

    def test_fallback_selection_select_all(self, ui):
        """Test fallback selecting all elements."""
        flattened = [
            {"type": "element", "name": "elem1", "path": "/path1", "category": "agent"},
            {
                "type": "element",
                "name": "elem2",
                "path": "/path2",
                "category": "command",
            },
        ]

        with patch("builtins.input", return_value="all"):
            result = ui._fallback_selection(flattened)

        assert len(result) == 2
        assert "/path1" in result
        assert "/path2" in result

    def test_fallback_selection_select_by_number(self, ui):
        """Test fallback selecting by number."""
        flattened = [
            {"type": "element", "name": "elem1", "path": "/path1", "category": "agent"},
            {
                "type": "element",
                "name": "elem2",
                "path": "/path2",
                "category": "command",
            },
        ]

        with patch("builtins.input", return_value="1,2"):
            result = ui._fallback_selection(flattened)

        assert len(result) == 2

    def test_fallback_selection_invalid_input(self, ui):
        """Test fallback with invalid input."""
        flattened = [
            {"type": "element", "name": "elem1", "path": "/path1", "category": "agent"}
        ]

        with patch("builtins.input", return_value="invalid"):
            result = ui._fallback_selection(flattened)

        assert result is None

    def test_fallback_selection_out_of_range(self, ui):
        """Test fallback with out of range number."""
        flattened = [
            {"type": "element", "name": "elem1", "path": "/path1", "category": "agent"}
        ]

        with patch("builtins.input", return_value="999"):
            result = ui._fallback_selection(flattened)

        assert result is None


class TestElementTypeDetection:
    """Test element type detection."""

    @pytest.fixture
    def ui(self):
        """Create UI instance with mocked scanner."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                yield InteractiveCheckboxUI(tmpdir)

    def test_get_element_type_agent(self, ui):
        """Test detecting agent type."""
        elem_type = ui._get_element_type(".claude/agents/my_agent.py")
        assert elem_type == "agent"

    def test_get_element_type_command(self, ui):
        """Test detecting command type."""
        elem_type = ui._get_element_type(".claude/commands/my_command.md")
        assert elem_type == "command"

    def test_get_element_type_skill(self, ui):
        """Test detecting skill type."""
        elem_type = ui._get_element_type(".claude/skills/my_skill/")
        assert elem_type == "skill"

    def test_get_element_type_hook(self, ui):
        """Test detecting hook type."""
        elem_type = ui._get_element_type(".claude/hooks/my_hook.py")
        assert elem_type == "hook"

    def test_get_element_type_unknown(self, ui):
        """Test detecting unknown type."""
        elem_type = ui._get_element_type("some/random/path")
        assert elem_type == "unknown"


class TestPromptUserSelection:
    """Test main prompt_user_selection method."""

    @pytest.fixture
    def ui(self):
        """Create UI instance with mocked scanner."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                yield InteractiveCheckboxUI(tmpdir)

    def test_prompt_user_selection_no_elements(self, ui):
        """Test prompt when no custom elements found."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {}

        result = ui.prompt_user_selection()

        assert result is None

    def test_prompt_user_selection_no_backup(self, ui):
        """Test prompt when backup is not available."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"agents": ["/path/to/agent.py"]}

        result = ui.prompt_user_selection(backup_available=False)

        assert result is None

    def test_prompt_user_selection_curses_import_error(self, ui):
        """Test prompt with curses import error fallback."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"agents": ["/path/to/agent.py"]}

        with patch.object(
            ui, "_run_curses_interface", side_effect=ImportError("No curses")
        ):
            with patch.object(
                ui, "_fallback_selection", return_value=["/path/to/agent.py"]
            ):
                result = ui.prompt_user_selection()

                assert result == ["/path/to/agent.py"]

    def test_prompt_user_selection_curses_error(self, ui):
        """Test prompt with general curses error fallback."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"agents": ["/path/to/agent.py"]}

        with patch.object(
            ui, "_run_curses_interface", side_effect=Exception("Curses error")
        ):
            with patch.object(
                ui, "_fallback_selection", return_value=["/path/to/agent.py"]
            ):
                result = ui.prompt_user_selection()

                assert result == ["/path/to/agent.py"]

    def test_prompt_user_selection_cancelled(self, ui):
        """Test prompt when user cancels."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"agents": ["/path/to/agent.py"]}

        with patch.object(ui, "_run_curses_interface", return_value=None):
            result = ui.prompt_user_selection()

            assert result is None

    def test_prompt_user_selection_no_selection_made(self, ui):
        """Test prompt when user doesn't select anything."""
        ui.scanner = MagicMock()
        ui.scanner.scan_custom_elements.return_value = {"agents": ["/path/to/agent.py"]}

        with patch.object(ui, "_run_curses_interface", return_value=set()):
            result = ui.prompt_user_selection()

            assert result is None


class TestConfirmSelection:
    """Test confirm_selection method."""

    @pytest.fixture
    def ui(self):
        """Create UI instance with mocked scanner."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                yield InteractiveCheckboxUI(tmpdir)

    def test_confirm_selection_user_confirms(self, ui):
        """Test user confirming selection."""
        selected = ["/path/to/agent.py", "/path/to/command.md"]

        with patch("builtins.input", return_value="y"):
            result = ui.confirm_selection(selected)

        assert result is True

    def test_confirm_selection_user_declines(self, ui):
        """Test user declining selection."""
        selected = ["/path/to/agent.py"]

        with patch("builtins.input", return_value="n"):
            result = ui.confirm_selection(selected)

        assert result is False

    def test_confirm_selection_keyboard_interrupt(self, ui):
        """Test handling keyboard interrupt."""
        selected = ["/path/to/agent.py"]

        with patch("builtins.input", side_effect=KeyboardInterrupt):
            result = ui.confirm_selection(selected)

        assert result is False

    def test_confirm_selection_eof(self, ui):
        """Test handling EOF."""
        selected = ["/path/to/agent.py"]

        with patch("builtins.input", side_effect=EOFError):
            result = ui.confirm_selection(selected)

        assert result is False


class TestFactoryFunction:
    """Test factory function."""

    def test_create_interactive_checkbox_ui_with_string(self):
        """Test creating UI with string path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                ui = create_interactive_checkbox_ui(tmpdir)

                assert isinstance(ui, InteractiveCheckboxUI)
                assert ui.project_path == Path(tmpdir).resolve()

    def test_create_interactive_checkbox_ui_with_path(self):
        """Test creating UI with Path object."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir)
            with patch(
                "moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"
            ):
                ui = create_interactive_checkbox_ui(path)

                assert isinstance(ui, InteractiveCheckboxUI)
