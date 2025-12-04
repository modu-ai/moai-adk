"""
Comprehensive unit tests for InteractiveCheckboxUI with 85%+ coverage.

Tests cover:
- Initialization
- prompt_user_selection() with various conditions
- _get_elements_by_category() organization
- _flatten_elements() flattening logic
- Navigation methods (_navigate_up, _navigate_down)
- Selection methods (_toggle_selection, _select_all, _select_none)
- Display and confirmation dialogs
- Fallback selection method
- Element type detection
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch, Mock
from io import StringIO

from moai_adk.core.migration.interactive_checkbox_ui import (
    InteractiveCheckboxUI,
    create_interactive_checkbox_ui,
)


class TestInteractiveCheckboxUIInitialization:
    """Test InteractiveCheckboxUI initialization."""

    def test_init_with_path_string(self):
        """Test initialization with string path."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            assert ui.project_path == Path(tmpdir).resolve()

    def test_init_with_path_object(self):
        """Test initialization with Path object."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir)
            ui = InteractiveCheckboxUI(path)
            assert ui.project_path == path.resolve()

    def test_init_scanner_created(self):
        """Test scanner is created during initialization."""
        with tempfile.TemporaryDirectory() as tmpdir:
            with patch("moai_adk.core.migration.interactive_checkbox_ui.create_custom_element_scanner"):
                ui = InteractiveCheckboxUI(tmpdir)
                assert hasattr(ui, "scanner")

    def test_init_selected_indices_empty(self):
        """Test selected_indices starts empty."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            assert ui.selected_indices == set()

    def test_init_current_index_zero(self):
        """Test current_index starts at 0."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            assert ui.current_index == 0


class TestInteractiveCheckboxUIGetElementsByCategory:
    """Test _get_elements_by_category() method."""

    def test_get_elements_by_category_empty(self):
        """Test with no custom elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            with patch.object(ui.scanner, "scan_custom_elements", return_value={}):
                result = ui._get_elements_by_category()
                assert isinstance(result, dict)

    def test_get_elements_by_category_skills(self):
        """Test organizing skill elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            mock_skill = MagicMock()
            mock_skill.name = "test-skill"
            mock_skill.path = Path(tmpdir) / "skills" / "test-skill"
            mock_skill.is_template = False

            with patch.object(ui.scanner, "scan_custom_elements", return_value={"skills": [mock_skill]}):
                result = ui._get_elements_by_category()
                assert "Skills" in result

    def test_get_elements_by_category_agents(self):
        """Test organizing agent elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            with patch.object(
                ui.scanner,
                "scan_custom_elements",
                return_value={"agents": [str(Path(tmpdir) / "agents" / "test-agent.md")]}
            ):
                result = ui._get_elements_by_category()
                assert "Agents" in result

    def test_get_elements_by_category_commands(self):
        """Test organizing command elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            with patch.object(
                ui.scanner,
                "scan_custom_elements",
                return_value={"commands": [str(Path(tmpdir) / "commands" / "test-cmd.md")]}
            ):
                result = ui._get_elements_by_category()
                assert "Commands" in result

    def test_get_elements_by_category_hooks(self):
        """Test organizing hook elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            with patch.object(
                ui.scanner,
                "scan_custom_elements",
                return_value={"hooks": [str(Path(tmpdir) / "hooks" / "test-hook.py")]}
            ):
                result = ui._get_elements_by_category()
                assert "Hooks" in result

    def test_get_elements_by_category_removes_empty(self):
        """Test empty categories are removed."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            with patch.object(ui.scanner, "scan_custom_elements", return_value={}):
                result = ui._get_elements_by_category()
                # All empty categories should be filtered out
                assert len(result) == 0 or all(v for v in result.values())


class TestInteractiveCheckboxUIFlattenElements:
    """Test _flatten_elements() method."""

    def test_flatten_elements_empty(self):
        """Test flattening empty elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            result = ui._flatten_elements({})
            assert result == []

    def test_flatten_elements_single_category(self):
        """Test flattening single category."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            elements_by_category = {
                "Agents": [
                    {"name": "test-agent", "path": "/path/to/agent", "type": "agent"}
                ]
            }

            result = ui._flatten_elements(elements_by_category)

            # Should have header and element
            assert len(result) == 2
            assert result[0]["type"] == "header"
            assert result[1]["type"] == "element"

    def test_flatten_elements_multiple_categories(self):
        """Test flattening multiple categories."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            elements_by_category = {
                "Agents": [{"name": "agent1", "path": "/p1", "type": "agent"}],
                "Commands": [{"name": "cmd1", "path": "/p2", "type": "command"}],
            }

            result = ui._flatten_elements(elements_by_category)

            # Should have headers and elements
            headers = [e for e in result if e["type"] == "header"]
            elements = [e for e in result if e["type"] == "element"]
            assert len(headers) == 2
            assert len(elements) == 2

    def test_flatten_elements_preserves_paths(self):
        """Test flattened elements preserve paths."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            elements_by_category = {
                "Skills": [{"name": "test-skill", "path": "/test/path/skill", "type": "skill"}]
            }

            result = ui._flatten_elements(elements_by_category)
            element = result[1]
            assert element["path"] == "/test/path/skill"


class TestInteractiveCheckboxUINavigation:
    """Test navigation methods."""

    def test_navigate_up_from_middle(self):
        """Test navigating up from middle element."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.current_index = 2

            elements = [
                {"type": "header"},
                {"type": "element"},
                {"type": "element"},
                {"type": "element"},
            ]

            ui._navigate_up(elements)
            assert ui.current_index == 1

    def test_navigate_up_skips_headers(self):
        """Test navigate up skips header elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.current_index = 2

            elements = [
                {"type": "header"},
                {"type": "header"},
                {"type": "element"},
            ]

            ui._navigate_up(elements)
            assert ui.current_index == 2  # Can't go up past headers

    def test_navigate_up_at_boundary(self):
        """Test navigate up at list boundary."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.current_index = 0

            elements = [{"type": "element"}]

            ui._navigate_up(elements)
            # Index doesn't change if already at boundary
            assert ui.current_index == 0

    def test_navigate_down_from_middle(self):
        """Test navigating down from middle element."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.current_index = 0

            elements = [
                {"type": "element"},
                {"type": "element"},
                {"type": "element"},
            ]

            ui._navigate_down(elements)
            assert ui.current_index == 1

    def test_navigate_down_skips_headers(self):
        """Test navigate down skips header elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.current_index = 0

            elements = [
                {"type": "element"},
                {"type": "header"},
                {"type": "header"},
                {"type": "element"},
            ]

            ui._navigate_down(elements)
            assert ui.current_index == 3

    def test_navigate_down_at_boundary(self):
        """Test navigate down at list boundary."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.current_index = 0

            elements = [{"type": "element"}]

            ui._navigate_down(elements)
            # Index doesn't change if already at end
            assert ui.current_index == 0


class TestInteractiveCheckboxUISelection:
    """Test selection methods."""

    def test_toggle_selection_adds(self):
        """Test toggle selection adds to set."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.current_index = 0

            elements = [{"type": "element"}]

            ui._toggle_selection(elements)
            assert 0 in ui.selected_indices

    def test_toggle_selection_removes(self):
        """Test toggle selection removes from set."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.current_index = 0
            ui.selected_indices.add(0)

            elements = [{"type": "element"}]

            ui._toggle_selection(elements)
            assert 0 not in ui.selected_indices

    def test_toggle_selection_ignores_headers(self):
        """Test toggle selection ignores headers."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.current_index = 0

            elements = [{"type": "header"}]

            ui._toggle_selection(elements)
            assert 0 not in ui.selected_indices

    def test_select_all(self):
        """Test select all elements."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            elements = [
                {"type": "element"},
                {"type": "header"},
                {"type": "element"},
                {"type": "element"},
            ]

            ui._select_all(elements)

            # Should select all non-header elements
            assert 0 in ui.selected_indices
            assert 2 in ui.selected_indices
            assert 3 in ui.selected_indices
            assert 1 not in ui.selected_indices  # Header not selected

    def test_select_none(self):
        """Test select none clears selection."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)
            ui.selected_indices.add(0)
            ui.selected_indices.add(1)

            elements = [{"type": "element"}, {"type": "element"}]

            ui._select_none(elements)

            assert len(ui.selected_indices) == 0




class TestInteractiveCheckboxUIGetElementType:
    """Test _get_element_type() method."""

    def test_get_element_type_agent(self):
        """Test detecting agent element type."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            result = ui._get_element_type("/.claude/agents/test-agent.md")
            assert result == "agent"

    def test_get_element_type_command(self):
        """Test detecting command element type."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            result = ui._get_element_type("/.claude/commands/test-cmd.md")
            assert result == "command"

    def test_get_element_type_skill(self):
        """Test detecting skill element type."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            result = ui._get_element_type("/.claude/skills/test-skill")
            assert result == "skill"

    def test_get_element_type_hook(self):
        """Test detecting hook element type."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            result = ui._get_element_type("/.claude/hooks/test-hook.py")
            assert result == "hook"

    def test_get_element_type_unknown(self):
        """Test unknown element type."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            result = ui._get_element_type("/unknown/path")
            assert result == "unknown"


class TestInteractiveCheckboxUIConfirmSelection:
    """Test confirm_selection() method."""

    def test_confirm_selection_user_confirms(self):
        """Test user confirms selection."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            with patch("builtins.input", return_value="y"):
                result = ui.confirm_selection(["/path1", "/path2"])
                assert result is True

    def test_confirm_selection_user_denies(self):
        """Test user denies selection."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            with patch("builtins.input", return_value="n"):
                result = ui.confirm_selection(["/path1"])
                assert result is False

    def test_confirm_selection_keyboard_interrupt(self):
        """Test confirmation handles keyboard interrupt."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            with patch("builtins.input", side_effect=KeyboardInterrupt):
                result = ui.confirm_selection(["/path1"])
                assert result is False


class TestInteractiveCheckboxUIPromptUserSelection:
    """Test prompt_user_selection() main method."""

    def test_prompt_user_selection_no_elements(self):
        """Test prompt when no custom elements found."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            with patch.object(ui, "_get_elements_by_category", return_value={}):
                result = ui.prompt_user_selection()
                assert result is None


    def test_prompt_user_selection_curses_unavailable_uses_fallback(self):
        """Test fallback when curses not available."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = InteractiveCheckboxUI(tmpdir)

            elements = {"Agents": [{"name": "test", "path": "/p1", "type": "agent"}]}

            with patch.object(ui, "_get_elements_by_category", return_value=elements):
                with patch.object(ui, "_run_curses_interface", side_effect=ImportError):
                    with patch.object(ui, "_fallback_selection", return_value=["/p1"]):
                        result = ui.prompt_user_selection(backup_available=True)
                        assert result is not None


class TestCreateInteractiveCheckboxUIFactory:
    """Test factory function."""

    def test_factory_creates_instance(self):
        """Test factory function creates UI instance."""
        with tempfile.TemporaryDirectory() as tmpdir:
            ui = create_interactive_checkbox_ui(tmpdir)
            assert isinstance(ui, InteractiveCheckboxUI)

    def test_factory_with_path_object(self):
        """Test factory with Path object."""
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir)
            ui = create_interactive_checkbox_ui(path)
            assert isinstance(ui, InteractiveCheckboxUI)
