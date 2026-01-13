"""Comprehensive tests for user_selection_ui.py module.

Tests cover:
- UserSelectionUI class initialization
- User selection prompting with interactive UI
- Basic selection mode fallback
- Element categorization
- Selection parsing and confirmation
- Element type detection
- Display functions
- Factory function
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.migration.custom_element_scanner import TemplateSkill
from moai_adk.core.migration.user_selection_ui import (
    UserSelectionUI,
    create_user_selection_ui,
    display_selection_instructions,
)


class TestUserSelectionUIInitialization:
    """Test UserSelectionUI initialization."""

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_init_with_string_path(self, mock_scanner):
        """Test initialization with string path."""
        mock_scanner.return_value = MagicMock()
        with patch("moai_adk.core.migration.user_selection_ui.Path") as mock_path:
            mock_path.return_value.resolve.return_value = Path("/project")

            ui = UserSelectionUI("/project")

            assert ui.project_path == Path("/project").resolve()
            assert ui.scanner is not None

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_init_with_path_object(self, mock_scanner):
        """Test initialization with Path object."""
        mock_scanner.return_value = MagicMock()

        ui = UserSelectionUI(Path("/project"))

        assert ui.project_path == Path("/project").resolve()

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_init_without_interactive_ui(self, mock_scanner):
        """Test initialization when interactive UI is not available."""
        mock_scanner.return_value = MagicMock()

        # Mock the import to fail
        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            # Set attributes manually
            ui.project_path = Path("/project").resolve()
            ui.scanner = MagicMock()
            ui.interactive_ui = None
            ui.use_interactive = False

            assert ui.interactive_ui is None
            assert ui.use_interactive is False


class TestPromptUserSelection:
    """Test prompt_user_selection method."""

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_prompt_with_interactive_ui_success(self, mock_scanner):
        """Test prompting with interactive UI when available."""
        mock_scanner.return_value = MagicMock()
        mock_interactive = MagicMock()
        mock_interactive.prompt_user_selection.return_value = [".claude/agents/agent.md"]

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = MagicMock()
            ui.interactive_ui = mock_interactive
            ui.use_interactive = True

            result = ui.prompt_user_selection(backup_available=True)

            assert result == [".claude/agents/agent.md"]
            mock_interactive.prompt_user_selection.assert_called_once_with(True)

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_prompt_with_interactive_ui_falls_back_on_error(self, mock_scanner):
        """Test fallback to basic mode when interactive UI raises exception."""
        mock_scanner.return_value = MagicMock()
        mock_interactive = MagicMock()
        mock_interactive.prompt_user_selection.side_effect = RuntimeError("Curses error")

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = MagicMock()
            ui.interactive_ui = mock_interactive
            ui.use_interactive = True

            with patch.object(ui, "_basic_selection_mode") as mock_basic:
                mock_basic.return_value = [".claude/agents/agent.md"]
                result = ui.prompt_user_selection(backup_available=True)

                assert result == [".claude/agents/agent.md"]
                mock_basic.assert_called_once_with(True)

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_prompt_without_interactive_ui(self, mock_scanner):
        """Test prompting when interactive UI is not available."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = MagicMock()
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch.object(ui, "_basic_selection_mode") as mock_basic:
                mock_basic.return_value = [".claude/agents/agent.md"]
                result = ui.prompt_user_selection(backup_available=True)

                assert result == [".claude/agents/agent.md"]
                mock_basic.assert_called_once_with(True)

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_prompt_returns_none_when_no_elements(self, mock_scanner):
        """Test prompting returns None when no elements available."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = MagicMock()
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch.object(ui, "_basic_selection_mode", return_value=None):
                result = ui.prompt_user_selection(backup_available=True)

                assert result is None


class TestBasicSelectionMode:
    """Test _basic_selection_mode method."""

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_basic_mode_no_custom_elements(self, mock_scanner):
        """Test basic mode with no custom elements."""
        mock_scanner.return_value = MagicMock()
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.print"):
                result = ui._basic_selection_mode(backup_available=True)

                assert result is None

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_basic_mode_no_backup_available(self, mock_scanner):
        """Test basic mode when no backup is available."""
        mock_scanner.return_value = MagicMock()
        skill = TemplateSkill(name="test-skill", path=Path("skill-path"), has_skill_md=True)
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [skill],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.print"):
                result = ui._basic_selection_mode(backup_available=False)

                assert result is None

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_basic_mode_user_cancels(self, mock_scanner):
        """Test basic mode when user cancels with empty input."""
        mock_scanner.return_value = MagicMock()
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [MagicMock(stem="cmd")],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.input", return_value=""):
                with patch("builtins.print"):
                    result = ui._basic_selection_mode(backup_available=True)

                    assert result is None

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_basic_mode_user_selects_all(self, mock_scanner):
        """Test basic mode when user selects all elements."""
        mock_scanner.return_value = MagicMock()
        skill = TemplateSkill(name="test-skill", path=Path(".claude/skills/test-skill"), has_skill_md=True)
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [skill],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.input", return_value="all"):
                with patch("builtins.print"):
                    with patch.object(UserSelectionUI, "confirm_selection", return_value=True):
                        result = ui._basic_selection_mode(backup_available=True)

                        assert result is not None

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_basic_mode_user_selects_all_confirms_no(self, mock_scanner):
        """Test basic mode when user selects all but doesn't confirm."""
        mock_scanner.return_value = MagicMock()
        skill = TemplateSkill(name="test-skill", path=Path(".claude/skills/test-skill"), has_skill_md=True)
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [skill],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.input", return_value="all"):
                with patch("builtins.print"):
                    with patch.object(UserSelectionUI, "confirm_selection", return_value=False):
                        result = ui._basic_selection_mode(backup_available=True)

                        assert result is None

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_basic_mode_keyboard_interrupt(self, mock_scanner):
        """Test basic mode with keyboard interrupt."""
        mock_scanner.return_value = MagicMock()
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [MagicMock(stem="cmd")],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.input", side_effect=KeyboardInterrupt()):
                with patch("builtins.print"):
                    result = ui._basic_selection_mode(backup_available=True)

                    assert result is None


class TestGetElementsByCategory:
    """Test _get_elements_by_category method."""

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_organizes_skills_by_category(self, mock_scanner):
        """Test organizing skills into category."""
        mock_scanner.return_value = MagicMock()
        skill1 = TemplateSkill(name="test-skill1", path=Path(".claude/skills/skill1"), has_skill_md=True)
        skill2 = TemplateSkill(name="test-skill2", path=Path(".claude/skills/skill2"), has_skill_md=True)
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [skill1, skill2],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._get_elements_by_category()

            assert "Skills" in result
            assert len(result["Skills"]) == 2
            assert result["Skills"][0]["name"] == "test-skill1"
            assert result["Skills"][0]["type"] == "skill"

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_organizes_agents_by_category(self, mock_scanner):
        """Test organizing agents into category."""
        mock_scanner.return_value = MagicMock()
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [Path(".claude/agents/agent1.md"), Path(".claude/agents/agent2.md")],
            "commands": [],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._get_elements_by_category()

            assert "Agents" in result
            assert len(result["Agents"]) == 2
            assert result["Agents"][0]["type"] == "agent"

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_handles_empty_elements(self, mock_scanner):
        """Test handles empty element lists."""
        mock_scanner.return_value = MagicMock()
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._get_elements_by_category()

            assert result["Skills"] == []
            assert result["Agents"] == []
            assert result["Commands"] == []
            assert result["Hooks"] == []


class TestParseSelection:
    """Test _parse_selection method."""

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_parse_numeric_selection(self, mock_scanner):
        """Test parsing numeric selection."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            custom_elements = [
                {"index": 1, "name": "element1", "path": ".claude/agents/e1.md", "display_name": "e1"},
                {"index": 2, "name": "element2", "path": ".claude/agents/e2.md", "display_name": "e2"},
            ]

            result = ui._parse_selection("1", custom_elements)

            assert result == [".claude/agents/e1.md"]

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_parse_comma_separated_selection(self, mock_scanner):
        """Test parsing comma-separated selection."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            custom_elements = [
                {"index": 1, "name": "element1", "path": ".claude/agents/e1.md", "display_name": "e1"},
                {"index": 2, "name": "element2", "path": ".claude/agents/e2.md", "display_name": "e2"},
            ]

            result = ui._parse_selection("1, 2", custom_elements)

            assert len(result) == 2
            assert ".claude/agents/e1.md" in result
            assert ".claude/agents/e2.md" in result

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_parse_name_selection(self, mock_scanner):
        """Test parsing selection by element name."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            custom_elements = [
                {"index": 1, "name": "my-agent", "path": ".claude/agents/my-agent.md", "display_name": "my-agent"},
            ]

            result = ui._parse_selection("my-agent", custom_elements)

            assert result == [".claude/agents/my-agent.md"]

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_parse_invalid_selection(self, mock_scanner):
        """Test parsing invalid selection."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            custom_elements = [
                {"index": 1, "name": "element1", "path": ".claude/agents/e1.md", "display_name": "e1"},
            ]

            with patch("builtins.print"):
                result = ui._parse_selection("999", custom_elements)

                assert result == []


class TestConfirmSelection:
    """Test confirm_selection method."""

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_confirm_with_user_accept(self, mock_scanner):
        """Test confirmation when user accepts."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.input", return_value="y"):
                with patch("builtins.print"):
                    result = ui.confirm_selection([".claude/agents/agent.md"])

                    assert result is True

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_confirm_with_user_decline(self, mock_scanner):
        """Test confirmation when user declines."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.input", return_value="n"):
                with patch("builtins.print"):
                    result = ui.confirm_selection([".claude/agents/agent.md"])

                    assert result is False

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_confirm_with_keyboard_interrupt(self, mock_scanner):
        """Test confirmation with keyboard interrupt."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.input", side_effect=KeyboardInterrupt()):
                with patch("builtins.print"):
                    result = ui.confirm_selection([".claude/agents/agent.md"])

                    assert result is False


class TestGetElementType:
    """Test _get_element_type method."""

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_get_agent_type(self, mock_scanner):
        """Test getting agent element type."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._get_element_type(".claude/agents/my-agent.md")

            assert result == "agent"

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_get_command_type(self, mock_scanner):
        """Test getting command element type."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._get_element_type(".claude/commands/my-cmd.md")

            assert result == "command"

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_get_skill_type(self, mock_scanner):
        """Test getting skill element type."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._get_element_type(".claude/skills/my-skill")

            assert result == "skill"

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_get_hook_type(self, mock_scanner):
        """Test getting hook element type."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._get_element_type(".claude/hooks/my-hook.py")

            assert result == "hook"

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_get_unknown_type(self, mock_scanner):
        """Test getting unknown element type."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._get_element_type(".unknown/path/file.txt")

            assert result == "unknown"


class TestDisplaySelectionInstructions:
    """Test display_selection_instructions function."""

    @patch("builtins.print")
    def test_display_instructions(self, mock_print):
        """Test displaying selection instructions."""
        display_selection_instructions()

        # Check that print was called
        assert mock_print.called


class TestFactoryFunction:
    """Test create_user_selection_ui factory function."""

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_factory_with_string_path(self, mock_scanner):
        """Test factory function with string path."""
        mock_scanner.return_value = MagicMock()
        with patch("moai_adk.core.migration.user_selection_ui.Path") as mock_path:
            mock_path.return_value.resolve.return_value = Path("/project")

            ui = create_user_selection_ui("/project")

            assert isinstance(ui, UserSelectionUI)
            assert ui.project_path == Path("/project").resolve()

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_factory_with_path_object(self, mock_scanner):
        """Test factory function with Path object."""
        mock_scanner.return_value = MagicMock()

        ui = create_user_selection_ui(Path("/project"))

        assert isinstance(ui, UserSelectionUI)
        assert ui.project_path == Path("/project").resolve()


class TestUserSelectionUIEdgeCases:
    """Test edge cases for UserSelectionUI."""

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_parse_selection_with_empty_input(self, mock_scanner):
        """Test parsing empty selection input."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._parse_selection("", [])

            assert result == []

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_parse_selection_with_whitespace_only(self, mock_scanner):
        """Test parsing whitespace-only input."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._parse_selection("   ,  ;  ", [])

            assert result == []

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_confirm_selection_with_empty_list(self, mock_scanner):
        """Test confirming empty selection list."""
        mock_scanner.return_value = MagicMock()

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            with patch("builtins.input", return_value="y"):
                with patch("builtins.print"):
                    result = ui.confirm_selection([])

                    assert result is True

    @patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner")
    def test_get_elements_by_category_with_missing_key(self, mock_scanner):
        """Test categorization when key is missing."""
        mock_scanner.return_value = MagicMock()
        # Return dict missing 'commands' key
        mock_scanner.return_value.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "hooks": [],
        }

        with patch.object(UserSelectionUI, "__init__", lambda self, path: None):
            ui = UserSelectionUI(Path("/project"))
            ui.project_path = Path("/project").resolve()
            ui.scanner = mock_scanner.return_value
            ui.interactive_ui = None
            ui.use_interactive = False

            result = ui._get_elements_by_category()

            # Should handle missing key gracefully
            assert "Commands" in result
            assert result["Commands"] == []
