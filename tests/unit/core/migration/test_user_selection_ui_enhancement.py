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

    def test_init_with_string_path(self):
        """Test initialization with string path."""
        with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
            mock_scanner.return_value = MagicMock()

            ui = UserSelectionUI("/project")

            assert ui.project_path == Path("/project").resolve()
            assert ui.scanner is not None

    def test_init_with_path_object(self):
        """Test initialization with Path object."""
        with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
            mock_scanner.return_value = MagicMock()

            ui = UserSelectionUI(Path("/project"))

            assert ui.project_path == Path("/project").resolve()

    def test_init_without_interactive_ui(self):
        """Test initialization when interactive UI is not available."""
        # Patch at the source module level before import
        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui",
            side_effect=ImportError("curses not available"),
        ):
            with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
                mock_scanner.return_value = MagicMock()

                # Need to reload the module to pick up the patched import
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                assert ui.interactive_ui is None
                assert ui.use_interactive is False

    def test_init_with_interactive_ui_available(self):
        """Test initialization when interactive UI is available."""
        mock_interactive = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui",
            return_value=mock_interactive,
        ):
            with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
                mock_scanner.return_value = MagicMock()

                # Need to reload the module to pick up the patched import
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                assert ui.interactive_ui == mock_interactive
                assert ui.use_interactive is True


class TestPromptUserSelection:
    """Test prompt_user_selection method."""

    def test_prompt_with_interactive_ui_success(self):
        """Test prompting with interactive UI when available."""
        mock_interactive = MagicMock()
        mock_interactive.prompt_user_selection.return_value = [".claude/agents/agent.md"]

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui",
            return_value=mock_interactive,
        ):
            with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
                mock_scanner.return_value = MagicMock()

                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))
                result = ui.prompt_user_selection(backup_available=True)

                assert result == [".claude/agents/agent.md"]
                mock_interactive.prompt_user_selection.assert_called_once_with(True)

    def test_prompt_with_interactive_ui_falls_back_on_error(self):
        """Test fallback to basic mode when interactive UI raises exception."""
        mock_interactive = MagicMock()
        mock_interactive.prompt_user_selection.side_effect = RuntimeError("Curses error")

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui",
            return_value=mock_interactive,
        ):
            with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
                mock_scanner.return_value = MagicMock()

                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch.object(ui, "_basic_selection_mode") as mock_basic:
                    mock_basic.return_value = [".claude/agents/agent.md"]
                    result = ui.prompt_user_selection(backup_available=True)

                    assert result == [".claude/agents/agent.md"]
                    mock_basic.assert_called_once_with(True)

    def test_prompt_without_interactive_ui(self):
        """Test prompting when interactive UI is not available."""
        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
                mock_scanner.return_value = MagicMock()

                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch.object(ui, "_basic_selection_mode") as mock_basic:
                    mock_basic.return_value = [".claude/agents/agent.md"]
                    result = ui.prompt_user_selection(backup_available=True)

                    assert result == [".claude/agents/agent.md"]
                    mock_basic.assert_called_once_with(True)

    def test_prompt_returns_none_when_no_elements(self):
        """Test prompting returns None when no elements available."""
        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
                mock_scanner.return_value = MagicMock()

                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch.object(ui, "_basic_selection_mode", return_value=None):
                    result = ui.prompt_user_selection(backup_available=True)

                    assert result is None


class TestBasicSelectionMode:
    """Test _basic_selection_mode method."""

    def test_basic_mode_no_custom_elements(self):
        """Test basic mode with no custom elements."""
        mock_scanner_instance = MagicMock()
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    result = ui._basic_selection_mode(backup_available=True)

                    assert result is None

    def test_basic_mode_no_backup_available(self):
        """Test basic mode when no backup is available."""
        mock_scanner_instance = MagicMock()
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [MagicMock(name="test-skill", path="skill-path", type="skill")],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    result = ui._basic_selection_mode(backup_available=False)

                    assert result is None

    def test_basic_mode_user_cancels(self):
        """Test basic mode when user cancels with empty input."""
        mock_scanner_instance = MagicMock()
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    with patch("builtins.input", return_value=""):
                        result = ui._basic_selection_mode(backup_available=True)

                        assert result is None

    def test_basic_mode_user_selects_all(self):
        """Test basic mode when user selects all elements."""
        mock_scanner_instance = MagicMock()
        test_skill = TemplateSkill(name="test-skill", path=Path(".claude/skills/test"), has_skill_md=True)
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [test_skill],
            "agents": [],
            "commands": [],
            "hooks": [],
        }
        mock_scanner_instance.get_custom_elements_display_list.return_value = [
            {"index": 1, "name": "test-skill", "path": ".claude/skills/test", "type": "skill"}
        ]

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                # Patch _get_elements_by_category to return proper structure
                with patch.object(
                    ui,
                    "_get_elements_by_category",
                    return_value={
                        "Agents": [],
                        "Commands": [],
                        "Skills": [{"name": "test", "path": ".claude/skills/test", "type": "skill"}],
                        "Hooks": [],
                    },
                ):
                    with patch("builtins.print"):
                        with patch("builtins.input", return_value="all"):
                            with patch.object(ui, "confirm_selection", return_value=True):
                                result = ui._basic_selection_mode(backup_available=True)

                                assert result is not None

    def test_basic_mode_user_selects_all_confirms_no(self):
        """Test basic mode when user selects all but declines confirmation."""
        mock_scanner_instance = MagicMock()
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    with patch("builtins.input", return_value="all"):
                        with patch.object(ui, "confirm_selection", return_value=False):
                            result = ui._basic_selection_mode(backup_available=True)

                            assert result is None

    def test_basic_mode_numbered_selection(self):
        """Test basic mode with numbered selection."""
        mock_scanner_instance = MagicMock()
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    # First input returns the display data, second returns the selection
                    ui._get_elements_by_category = lambda: {
                        "Agents": [{"name": "agent1.md", "path": ".claude/agents/agent1.md", "type": "agent"}],
                        "Commands": [],
                        "Skills": [],
                        "Hooks": [],
                    }

                    with patch("builtins.input", return_value="1"):
                        with patch.object(ui, "confirm_selection", return_value=True):
                            result = ui._basic_selection_mode(backup_available=True)

                            assert result == [".claude/agents/agent1.md"]

    def test_basic_mode_invalid_number(self):
        """Test basic mode with invalid number input."""
        mock_scanner_instance = MagicMock()
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    ui._get_elements_by_category = lambda: {
                        "Agents": [{"name": "agent1.md", "path": ".claude/agents/agent1.md", "type": "agent"}],
                        "Commands": [],
                        "Skills": [],
                        "Hooks": [],
                    }

                    with patch("builtins.input", return_value="99"):
                        result = ui._basic_selection_mode(backup_available=True)

                        assert result is None

    def test_basic_mode_invalid_input(self):
        """Test basic mode with invalid input."""
        mock_scanner_instance = MagicMock()
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    ui._get_elements_by_category = lambda: {
                        "Agents": [{"name": "agent1.md", "path": ".claude/agents/agent1.md", "type": "agent"}],
                        "Commands": [],
                        "Skills": [],
                        "Hooks": [],
                    }

                    with patch("builtins.input", return_value="invalid"):
                        result = ui._basic_selection_mode(backup_available=True)

                        assert result is None

    def test_basic_mode_keyboard_interrupt(self):
        """Test basic mode with keyboard interrupt."""
        mock_scanner_instance = MagicMock()
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    with patch("builtins.input", side_effect=KeyboardInterrupt):
                        result = ui._basic_selection_mode(backup_available=True)

                        assert result is None


class TestGetElementsByCategory:
    """Test _get_elements_by_category method."""

    def test_organizes_skills_by_category(self):
        """Test organizing skills by category."""
        with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
            test_skill = TemplateSkill(name="test-skill", path=Path("skills/test"), has_skill_md=True)
            mock_scanner.return_value.scan_custom_elements.return_value = {
                "skills": [test_skill],
                "agents": [],
                "commands": [],
                "hooks": [],
            }

            ui = UserSelectionUI(Path("/project"))
            result = ui._get_elements_by_category()

            assert "Skills" in result
            assert len(result["Skills"]) == 1
            assert result["Skills"][0]["name"] == "test-skill"

    def test_organizes_agents_by_category(self):
        """Test organizing agents by category."""
        with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
            mock_scanner.return_value.scan_custom_elements.return_value = {
                "skills": [],
                "agents": [Path(".claude/agents/custom/agent.md")],
                "commands": [],
                "hooks": [],
            }

            ui = UserSelectionUI(Path("/project"))
            result = ui._get_elements_by_category()

            assert "Agents" in result
            assert len(result["Agents"]) == 1

    def test_organizes_commands_by_category(self):
        """Test organizing commands by category."""
        with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
            mock_scanner.return_value.scan_custom_elements.return_value = {
                "skills": [],
                "agents": [],
                "commands": [Path(".claude/commands/custom/cmd.md")],
                "hooks": [],
            }

            ui = UserSelectionUI(Path("/project"))
            result = ui._get_elements_by_category()

            assert "Commands" in result
            assert len(result["Commands"]) == 1

    def test_organizes_hooks_by_category(self):
        """Test organizing hooks by category."""
        with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
            mock_scanner.return_value.scan_custom_elements.return_value = {
                "skills": [],
                "agents": [],
                "commands": [],
                "hooks": [Path(".claude/hooks/custom/hook.py")],
            }

            ui = UserSelectionUI(Path("/project"))
            result = ui._get_elements_by_category()

            assert "Hooks" in result
            assert len(result["Hooks"]) == 1

    def test_organizes_mixed_elements(self):
        """Test organizing mixed element types."""
        with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
            test_skill = TemplateSkill(name="skill1", path=Path("skills/skill1"), has_skill_md=True)
            mock_scanner.return_value.scan_custom_elements.return_value = {
                "skills": [test_skill],
                "agents": [Path(".claude/agents/agent.md")],
                "commands": [Path(".claude/commands/cmd.md")],
                "hooks": [Path(".claude/hooks/hook.py")],
            }

            ui = UserSelectionUI(Path("/project"))
            result = ui._get_elements_by_category()

            assert len(result["Skills"]) == 1
            assert len(result["Agents"]) == 1
            assert len(result["Commands"]) == 1
            assert len(result["Hooks"]) == 1

    def test_handles_empty_elements(self):
        """Test handling empty elements."""
        mock_scanner_instance = MagicMock()
        mock_scanner_instance.scan_custom_elements.return_value = {
            "skills": [],
            "agents": [],
            "commands": [],
            "hooks": [],
        }

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))
                result = ui._get_elements_by_category()

                assert all(len(v) == 0 for v in result.values())


class TestParseSelection:
    """Test _parse_selection method."""

    def test_parse_numeric_selection(self):
        """Test parsing numeric selection."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1"},
                    {"index": 2, "name": "agent2", "path": ".claude/agents/agent2.md", "display_name": "agent2"},
                ]

                result = ui._parse_selection("1", elements)

                assert result == [".claude/agents/agent1.md"]

    def test_parse_comma_separated_selection(self):
        """Test parsing comma-separated selection."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1"},
                    {"index": 2, "name": "agent2", "path": ".claude/agents/agent2.md", "display_name": "agent2"},
                ]

                result = ui._parse_selection("1,2", elements)

                assert len(result) == 2

    def test_parse_semicolon_separated_selection(self):
        """Test parsing semicolon-separated selection."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1"},
                    {"index": 2, "name": "agent2", "path": ".claude/agents/agent2.md", "display_name": "agent2"},
                ]

                result = ui._parse_selection("1;2", elements)

                assert len(result) == 2

    def test_parse_space_separated_selection(self):
        """Test parsing space-separated selection."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1"},
                    {"index": 2, "name": "agent2", "path": ".claude/agents/agent2.md", "display_name": "agent2"},
                ]

                result = ui._parse_selection("1 2", elements)

                assert len(result) == 2

    def test_parse_name_selection(self):
        """Test parsing by element name."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "my-agent", "path": ".claude/agents/my-agent.md", "display_name": "my-agent"},
                ]

                result = ui._parse_selection("my-agent", elements)

                assert result == [".claude/agents/my-agent.md"]

    def test_parse_partial_name_selection(self):
        """Test parsing by partial element name."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {
                        "index": 1,
                        "name": "my-custom-agent",
                        "path": ".claude/agents/my-custom-agent.md",
                        "display_name": "my-custom-agent",
                    },
                ]

                result = ui._parse_selection("custom", elements)

                assert result == [".claude/agents/my-custom-agent.md"]

    def test_parse_case_insensitive_selection(self):
        """Test case-insensitive name matching."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "MyAgent", "path": ".claude/agents/MyAgent.md", "display_name": "MyAgent"},
                ]

                result = ui._parse_selection("myagent", elements)

                assert result == [".claude/agents/MyAgent.md"]

    def test_parse_invalid_selection(self):
        """Test parsing invalid selection."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1"},
                ]

                with patch("builtins.print"):
                    result = ui._parse_selection("invalid", elements)

                    assert result == []

    def test_parse_mixed_separators(self):
        """Test parsing with mixed separators."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1"},
                    {"index": 2, "name": "agent2", "path": ".claude/agents/agent2.md", "display_name": "agent2"},
                    {"index": 3, "name": "agent3", "path": ".claude/agents/agent3.md", "display_name": "agent3"},
                ]

                result = ui._parse_selection("1, 2;3", elements)

                assert len(result) == 3


class TestConfirmSelection:
    """Test confirm_selection method."""

    def test_confirm_with_user_accept(self):
        """Test confirmation when user accepts."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    with patch("builtins.input", return_value="y"):
                        result = ui.confirm_selection([".claude/agents/agent.md"])

                        assert result is True

    def test_confirm_with_user_decline(self):
        """Test confirmation when user declines."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    with patch("builtins.input", return_value="n"):
                        result = ui.confirm_selection([".claude/agents/agent.md"])

                        assert result is False

    def test_confirm_with_keyboard_interrupt(self):
        """Test confirmation with keyboard interrupt."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    with patch("builtins.input", side_effect=KeyboardInterrupt):
                        result = ui.confirm_selection([".claude/agents/agent.md"])

                        assert result is False

    def test_confirm_displays_summary(self):
        """Test confirmation displays selection summary."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print") as mock_print:
                    with patch("builtins.input", return_value="y"):
                        ui.confirm_selection([".claude/agents/agent.md", ".claude/skills/skill"])

                        # Verify summary was printed
                        assert mock_print.call_count > 0


class TestGetElementType:
    """Test _get_element_type method."""

    def test_get_agent_type(self):
        """Test getting agent type."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                result = ui._get_element_type(".claude/agents/custom/agent.md")

                assert result == "agent"

    def test_get_command_type(self):
        """Test getting command type."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                result = ui._get_element_type(".claude/commands/custom/cmd.md")

                assert result == "command"

    def test_get_skill_type(self):
        """Test getting skill type."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                result = ui._get_element_type(".claude/skills/custom-skill")

                assert result == "skill"

    def test_get_hook_type(self):
        """Test getting hook type."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                result = ui._get_element_type(".claude/hooks/custom/hook.py")

                assert result == "hook"

    def test_get_unknown_type(self):
        """Test getting unknown type."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                result = ui._get_element_type(".claude/unknown/type")

                assert result == "unknown"


class TestDisplaySelectionInstructions:
    """Test display_selection_instructions function."""

    def test_display_instructions(self):
        """Test displaying selection instructions."""
        with patch("builtins.print") as mock_print:
            display_selection_instructions()

            assert mock_print.call_count > 0


class TestFactoryFunction:
    """Test create_user_selection_ui factory function."""

    def test_factory_with_string_path(self):
        """Test factory with string path."""
        result = create_user_selection_ui("/project")

        assert type(result).__name__ == "UserSelectionUI"
        assert result.project_path == Path("/project").resolve()

    def test_factory_with_path_object(self):
        """Test factory with Path object."""
        result = create_user_selection_ui(Path("/project"))

        assert type(result).__name__ == "UserSelectionUI"
        assert result.project_path == Path("/project").resolve()


class TestUserSelectionUIEdgeCases:
    """Test edge cases for UserSelectionUI."""

    def test_parse_selection_with_empty_input(self):
        """Test parsing with empty input."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1"}
                ]

                result = ui._parse_selection("", elements)

                assert result == []

    def test_parse_selection_with_whitespace_only(self):
        """Test parsing with whitespace only."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                elements = [
                    {"index": 1, "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1"}
                ]

                result = ui._parse_selection("   ", elements)

                assert result == []

    def test_confirm_selection_with_empty_list(self):
        """Test confirmation with empty list."""
        mock_scanner_instance = MagicMock()

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch("builtins.print"):
                    with patch("builtins.input", return_value="y"):
                        result = ui.confirm_selection([])

                        assert result is True

    def test_get_elements_by_category_with_missing_key(self):
        """Test _get_elements_by_category handles missing keys."""
        mock_scanner_instance = MagicMock()
        # Return empty dict (missing all keys)
        mock_scanner_instance.scan_custom_elements.return_value = {}

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui", side_effect=ImportError()
        ):
            with patch(
                "moai_adk.core.migration.user_selection_ui.create_custom_element_scanner",
                return_value=mock_scanner_instance,
            ):
                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                # Should not raise exception
                result = ui._get_elements_by_category()

                assert isinstance(result, dict)

    def test_interactive_ui_exception_with_fallback(self):
        """Test interactive UI exception triggers fallback."""
        mock_interactive = MagicMock()
        mock_interactive.prompt_user_selection.side_effect = RuntimeError("Curses failed")

        with patch(
            "moai_adk.core.migration.interactive_checkbox_ui.create_interactive_checkbox_ui",
            return_value=mock_interactive,
        ):
            with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
                mock_scanner.return_value = MagicMock()

                import importlib

                from moai_adk.core.migration import user_selection_ui

                importlib.reload(user_selection_ui)

                ui = user_selection_ui.UserSelectionUI(Path("/project"))

                with patch.object(ui, "_basic_selection_mode", return_value=[".claude/agents/agent.md"]):
                    with patch("builtins.print"):
                        result = ui.prompt_user_selection(backup_available=True)

                        assert result == [".claude/agents/agent.md"]
