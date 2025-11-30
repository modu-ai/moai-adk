"""Tests for UI enhancement in UserSelectionUI

This test module covers the improvement from manual number input to checkbox/space selection
for custom file restoration, making the UI more user-friendly and accessible.
"""

from pathlib import Path
from unittest.mock import patch, MagicMock
from io import StringIO

from moai_adk.core.migration.user_selection_ui import create_user_selection_ui


class TestUserSelectionUIEnhancement:
    """Test enhanced UI for custom element selection"""

    def test_ui_display_with_checkbox_instructions(self, tmp_path):
        """Given: Custom elements to select
        When: UserSelectionUI displays elements
        Then: Should show checkbox/space selection instructions instead of number input
        """
        # Setup project with custom elements
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        # Mock the scanner to return custom elements
        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            # Mock custom elements
            mock_elements = [
                {
                    "index": 1,
                    "type": "agent",
                    "name": "my-agent",
                    "path": ".claude/agents/my-agent.md",
                    "display_name": "my-agent (agent)"
                },
                {
                    "index": 2,
                    "type": "command",
                    "name": "my-command",
                    "path": ".claude/commands/moai/my-command.md",
                    "display_name": "my-command (command)"
                }
            ]
            mock_scanner_instance.get_custom_elements_display_list.return_value = mock_elements
            mock_scanner_instance.get_element_count.return_value = 2

            ui = create_user_selection_ui(project_path)

            # Capture output to verify instructions
            # First input is element selection, second input is confirmation
            with patch('builtins.input', side_effect=['1', '']):
                result = ui.prompt_user_selection(backup_available=True)

            # Should return selected element
            assert result is not None
            assert len(result) == 1
            assert ".claude/agents/my-agent.md" in result

    def test_ui_all_selection_enhanced(self, tmp_path):
        """Given: Multiple custom elements
        When: User types 'all'
        Then: Should select all elements and provide better confirmation
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            # Mock 4 custom elements
            mock_elements = [
                {"index": 1, "type": "agent", "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1 (agent)"},
                {"index": 2, "type": "command", "name": "cmd1", "path": ".claude/commands/moai/cmd1.md", "display_name": "cmd1 (command)"},
                {"index": 3, "type": "skill", "name": "skill1", "path": ".claude/skills/skill1", "display_name": "skill1 (skill)"},
                {"index": 4, "type": "hook", "name": "hook1", "path": ".claude/hooks/moai/hook1.py", "display_name": "hook1 (hook)"},
            ]
            mock_scanner_instance.get_custom_elements_display_list.return_value = mock_elements
            mock_scanner_instance.get_element_count.return_value = 4

            ui = create_user_selection_ui(project_path)

            # Test 'all' selection
            with patch('builtins.input', return_value='all'):
                with patch.object(ui, 'confirm_selection', return_value=True) as mock_confirm:
                    result = ui.prompt_user_selection(backup_available=True)

            # Should select all elements
            assert result is not None
            assert len(result) == 4

            # Should call confirmation with all elements
            mock_confirm.assert_called_once()
            all_paths = [".claude/agents/agent1.md", ".claude/commands/moai/cmd1.md",
                        ".claude/skills/skill1", ".claude/hooks/moai/hook1.py"]
            assert set(result) == set(all_paths)

    def test_ui_space_selection_enhanced(self, tmp_path):
        """Given: Custom elements for selection
        When: User uses space-separated input (1 3 4)
        Then: Should parse space-separated numbers and select elements
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            # Mock elements
            mock_elements = [
                {"index": 1, "type": "agent", "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1 (agent)"},
                {"index": 2, "type": "command", "name": "cmd1", "path": ".claude/commands/moai/cmd1.md", "display_name": "cmd1 (command)"},
                {"index": 3, "type": "skill", "name": "skill1", "path": ".claude/skills/skill1", "display_name": "skill1 (skill)"},
                {"index": 4, "type": "hook", "name": "hook1", "path": ".claude/hooks/moai/hook1.py", "display_name": "hook1 (hook)"},
            ]
            mock_scanner_instance.get_custom_elements_display_list.return_value = mock_elements
            mock_scanner_instance.get_element_count.return_value = 4

            ui = create_user_selection_ui(project_path)

            # Test space-separated selection (1 3 4 = agent1, skill1, hook1)
            # First input is element selection, second input is confirmation
            with patch('builtins.input', side_effect=['1 3 4', '']):
                result = ui.prompt_user_selection(backup_available=True)

            # Should select elements 1, 3, and 4
            expected = [".claude/agents/agent1.md", ".claude/skills/skill1", ".claude/hooks/moai/hook1.py"]
            assert set(result) == set(expected)

    def test_ui_enhanced_error_handling(self, tmp_path):
        """Given: Invalid user input
        When: User enters invalid selections
        Then: Should provide better error messages and continue gracefully
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            # Mock elements
            mock_elements = [
                {"index": 1, "type": "agent", "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1 (agent)"},
                {"index": 2, "type": "command", "name": "cmd1", "path": ".claude/commands/moai/cmd1.md", "display_name": "cmd1 (command)"},
            ]
            mock_scanner_instance.get_custom_elements_display_list.return_value = mock_elements
            mock_scanner_instance.get_element_count.return_value = 2

            ui = create_user_selection_ui(project_path)

            # Test with invalid selection (number 5 doesn't exist)
            with patch('builtins.input', return_value='5'):
                result = ui.prompt_user_selection(backup_available=True)

            # Should return None or empty list for invalid selection
            assert result is None

    def test_ui_enhanced_cancel_handling(self, tmp_path):
        """Given: User cancels selection
        When: User presses Ctrl+C or EOF
        Then: Should handle gracefully with enhanced cancellation message
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            mock_elements = [
                {"index": 1, "type": "agent", "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1 (agent)"},
            ]
            mock_scanner_instance.get_custom_elements_display_list.return_value = mock_elements
            mock_scanner_instance.get_element_count.return_value = 1

            ui = create_user_selection_ui(project_path)

            # Test cancellation with KeyboardInterrupt
            with patch('builtins.input', side_effect=KeyboardInterrupt):
                result = ui.prompt_user_selection(backup_available=True)

            # Should return None for cancellation
            assert result is None

            # Test cancellation with EOFError
            with patch('builtins.input', side_effect=EOFError):
                result = ui.prompt_user_selection(backup_available=True)

            assert result is None

    def test_ui_enhanced_no_elements(self, tmp_path):
        """Given: No custom elements found
        When: UserSelectionUI is initialized
        Then: Should show enhanced message and return None
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            # No custom elements
            mock_scanner_instance.get_custom_elements_display_list.return_value = []
            mock_scanner_instance.get_element_count.return_value = 0

            ui = create_user_selection_ui(project_path)

            result = ui.prompt_user_selection(backup_available=True)

            # Should return None for no elements
            assert result is None

    def test_ui_enhanced_backslash_compatibility(self, tmp_path):
        """Given: User input with backslash-separated numbers
        When: User enters "1\3\4" (Windows-style)
        Then: Should still work and show enhanced compatibility message
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            mock_elements = [
                {"index": 1, "type": "agent", "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1 (agent)"},
                {"index": 2, "type": "command", "name": "cmd1", "path": ".claude/commands/moai/cmd1.md", "display_name": "cmd1 (command)"},
                {"index": 3, "type": "skill", "name": "skill1", "path": ".claude/skills/skill1", "display_name": "skill1 (skill)"},
            ]
            mock_scanner_instance.get_custom_elements_display_list.return_value = mock_elements
            mock_scanner_instance.get_element_count.return_value = 3

            ui = create_user_selection_ui(project_path)

            # Test backslash-separated input (Windows compatibility)
            # First input is element selection, second input is confirmation
            with patch('builtins.input', side_effect=['1\\3', '']):
                with patch('builtins.print') as mock_print:
                    result = ui.prompt_user_selection(backup_available=True)

            # Should handle gracefully and select elements 1 and 3
            expected = [".claude/agents/agent1.md", ".claude/skills/skill1"]
            assert set(result) == set(expected)

    def test_ui_enhanced_mixed_separators(self, tmp_path):
        """Given: User input with mixed separators
        When: User enters "1, 3; 4" or similar mixed format
        Then: Should enhance parsing to handle mixed separators
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            mock_elements = [
                {"index": 1, "type": "agent", "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1 (agent)"},
                {"index": 2, "type": "command", "name": "cmd1", "path": ".claude/commands/moai/cmd1.md", "display_name": "cmd1 (command)"},
                {"index": 3, "type": "skill", "name": "skill1", "path": ".claude/skills/skill1", "display_name": "skill1 (skill)"},
                {"index": 4, "type": "hook", "name": "hook1", "path": ".claude/hooks/moai/hook1.py", "display_name": "hook1 (hook)"},
            ]
            mock_scanner_instance.get_custom_elements_display_list.return_value = mock_elements
            mock_scanner_instance.get_element_count.return_value = 4

            ui = create_user_selection_ui(project_path)

            # Test mixed separators
            # First input is element selection, second input is confirmation
            with patch('builtins.input', side_effect=['1, 3; 4', '']):
                result = ui.prompt_user_selection(backup_available=True)

            # Should select elements 1, 3, and 4
            expected = [".claude/agents/agent1.md", ".claude/skills/skill1", ".claude/hooks/moai/hook1.py"]
            assert set(result) == set(expected)

    def test_ui_enhanced_confirmation_enhancement(self, tmp_path):
        """Given: User makes selection
        When: Confirmation is requested
        Then: Should show enhanced confirmation with better element type display
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            mock_elements = [
                {"index": 1, "type": "agent", "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1 (agent)"},
                {"index": 2, "type": "command", "name": "cmd1", "path": ".claude/commands/moai/cmd1.md", "display_name": "cmd1 (command)"},
            ]
            mock_scanner_instance.get_custom_elements_display_list.return_value = mock_elements
            mock_scanner_instance.get_element_count.return_value = 2

            ui = create_user_selection_ui(project_path)

            selected_elements = [".claude/agents/agent1.md", ".claude/commands/moai/cmd1.md"]

            # Test enhanced confirmation display
            with patch('builtins.input', return_value='y'):
                result = ui.confirm_selection(selected_elements)

            # Should return True for confirmation
            assert result is True

    def test_ui_enhanced_empty_input_handling(self, tmp_path):
        """Given: User enters empty input
        When: Selection prompt is shown
        Then: Should handle gracefully and return None
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        with patch('moai_adk.core.migration.user_selection_ui.create_custom_element_scanner') as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance

            mock_elements = [
                {"index": 1, "type": "agent", "name": "agent1", "path": ".claude/agents/agent1.md", "display_name": "agent1 (agent)"},
            ]
            mock_scanner_instance.get_custom_elements_display_list.return_value = mock_elements
            mock_scanner_instance.get_element_count.return_value = 1

            ui = create_user_selection_ui(project_path)

            # Test empty input (just pressing Enter)
            with patch('builtins.input', return_value=''):
                result = ui.prompt_user_selection(backup_available=True)

            # Should return None for empty input
            assert result is None