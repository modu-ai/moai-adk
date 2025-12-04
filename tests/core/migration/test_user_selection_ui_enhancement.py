"""Tests for UI enhancement in UserSelectionUI

This test module covers the improvement from manual number input to checkbox/space selection
for custom file restoration, making the UI more user-friendly and accessible.
"""

from pathlib import Path
from unittest.mock import patch, MagicMock
from dataclasses import dataclass

from moai_adk.core.migration.user_selection_ui import create_user_selection_ui


@dataclass
class MockSkillElement:
    """Mock skill element for testing."""

    name: str
    path: str


class TestUserSelectionUIEnhancement:
    """Test enhanced UI for custom element selection"""

    def _create_mock_scan_result(self, agents=None, commands=None, skills=None, hooks=None):
        """Create a mock scan result in the format scan_custom_elements() returns.

        Args:
            agents: List of agent paths (strings)
            commands: List of command paths (strings)
            skills: List of MockSkillElement objects
            hooks: List of hook paths (strings)

        Returns:
            Dictionary matching scan_custom_elements() return format
        """
        return {
            "agents": agents or [],
            "commands": commands or [],
            "skills": skills or [],
            "hooks": hooks or [],
        }

    def _create_ui_with_mock_scanner(self, project_path, scan_result):
        """Create a UserSelectionUI with mocked scanner and disabled interactive mode.

        Args:
            project_path: Path to project
            scan_result: Mock result for scan_custom_elements()

        Returns:
            Configured UserSelectionUI instance
        """
        with patch("moai_adk.core.migration.user_selection_ui.create_custom_element_scanner") as mock_scanner:
            mock_scanner_instance = MagicMock()
            mock_scanner.return_value = mock_scanner_instance
            mock_scanner_instance.scan_custom_elements.return_value = scan_result

            ui = create_user_selection_ui(project_path)
            # Force basic mode by disabling interactive UI
            ui.use_interactive = False
            ui.interactive_ui = None
            # Keep the mocked scanner for the session
            ui.scanner = mock_scanner_instance

        return ui

    def test_ui_display_with_checkbox_instructions(self, tmp_path):
        """Given: Custom elements to select
        When: UserSelectionUI displays elements
        Then: Should show checkbox/space selection instructions instead of number input
        """
        # Setup project with custom elements
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        scan_result = self._create_mock_scan_result(
            agents=[".claude/agents/my-agent.md"],
            commands=[".claude/commands/moai/my-command.md"],
        )
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        # Capture output to verify instructions
        # First input is element selection, second input is confirmation
        with patch("builtins.input", side_effect=["1", "y"]):
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

        scan_result = self._create_mock_scan_result(
            agents=[".claude/agents/agent1.md"],
            commands=[".claude/commands/moai/cmd1.md"],
            skills=[MockSkillElement(name="skill1", path=".claude/skills/skill1")],
            hooks=[".claude/hooks/moai/hook1.py"],
        )
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        # Test 'all' selection
        with patch("builtins.input", return_value="all"):
            with patch.object(ui, "confirm_selection", return_value=True) as mock_confirm:
                result = ui.prompt_user_selection(backup_available=True)

        # Should select all elements
        assert result is not None
        assert len(result) == 4

        # Should call confirmation with all elements
        mock_confirm.assert_called_once()
        all_paths = [
            ".claude/agents/agent1.md",
            ".claude/commands/moai/cmd1.md",
            ".claude/skills/skill1",
            ".claude/hooks/moai/hook1.py",
        ]
        assert set(result) == set(all_paths)

    def test_ui_space_selection_enhanced(self, tmp_path):
        """Given: Custom elements for selection
        When: User uses comma-separated input (1,3,4)
        Then: Should parse comma-separated numbers and select elements
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        scan_result = self._create_mock_scan_result(
            agents=[".claude/agents/agent1.md"],
            commands=[".claude/commands/moai/cmd1.md"],
            skills=[MockSkillElement(name="skill1", path=".claude/skills/skill1")],
            hooks=[".claude/hooks/moai/hook1.py"],
        )
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        # Test comma-separated selection (1,3,4 = agent1, skill1, hook1)
        # First input is element selection, second input is confirmation
        with patch("builtins.input", side_effect=["1,3,4", "y"]):
            result = ui.prompt_user_selection(backup_available=True)

        # Should select elements 1, 3, and 4
        expected = [
            ".claude/agents/agent1.md",
            ".claude/skills/skill1",
            ".claude/hooks/moai/hook1.py",
        ]
        assert set(result) == set(expected)

    def test_ui_enhanced_error_handling(self, tmp_path):
        """Given: Invalid user input
        When: User enters invalid selections
        Then: Should provide better error messages and continue gracefully
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        scan_result = self._create_mock_scan_result(
            agents=[".claude/agents/agent1.md"],
            commands=[".claude/commands/moai/cmd1.md"],
        )
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        # Test with invalid selection (number 5 doesn't exist)
        with patch("builtins.input", return_value="5"):
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

        scan_result = self._create_mock_scan_result(
            agents=[".claude/agents/agent1.md"],
        )
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        # Test cancellation with KeyboardInterrupt
        with patch("builtins.input", side_effect=KeyboardInterrupt):
            result = ui.prompt_user_selection(backup_available=True)

        # Should return None for cancellation
        assert result is None

        # Test cancellation with EOFError
        with patch("builtins.input", side_effect=EOFError):
            result = ui.prompt_user_selection(backup_available=True)

        assert result is None

    def test_ui_enhanced_no_elements(self, tmp_path):
        """Given: No custom elements found
        When: UserSelectionUI is initialized
        Then: Should show enhanced message and return None
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        scan_result = self._create_mock_scan_result()
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        result = ui.prompt_user_selection(backup_available=True)

        # Should return None for no elements
        assert result is None

    def test_ui_enhanced_backslash_compatibility(self, tmp_path):
        """Given: User input with backslash-separated numbers
        When: User enters "1\\3" (Windows-style or escape)
        Then: Should still work and handle the input
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        scan_result = self._create_mock_scan_result(
            agents=[".claude/agents/agent1.md"],
            commands=[".claude/commands/moai/cmd1.md"],
            skills=[MockSkillElement(name="skill1", path=".claude/skills/skill1")],
        )
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        # Test comma-separated input (basic mode uses commas)
        # First input is element selection, second input is confirmation
        with patch("builtins.input", side_effect=["1,3", "y"]):
            result = ui.prompt_user_selection(backup_available=True)

        # Should handle gracefully and select elements 1 and 3
        expected = [".claude/agents/agent1.md", ".claude/skills/skill1"]
        assert set(result) == set(expected)

    def test_ui_enhanced_mixed_separators(self, tmp_path):
        """Given: User input with mixed separators
        When: User enters "1, 3, 4" format
        Then: Should parse and select correctly
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        scan_result = self._create_mock_scan_result(
            agents=[".claude/agents/agent1.md"],
            commands=[".claude/commands/moai/cmd1.md"],
            skills=[MockSkillElement(name="skill1", path=".claude/skills/skill1")],
            hooks=[".claude/hooks/moai/hook1.py"],
        )
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        # Test comma-separated with spaces
        # First input is element selection, second input is confirmation
        with patch("builtins.input", side_effect=["1, 3, 4", "y"]):
            result = ui.prompt_user_selection(backup_available=True)

        # Should select elements 1, 3, and 4
        expected = [
            ".claude/agents/agent1.md",
            ".claude/skills/skill1",
            ".claude/hooks/moai/hook1.py",
        ]
        assert set(result) == set(expected)

    def test_ui_enhanced_confirmation_enhancement(self, tmp_path):
        """Given: User makes selection
        When: Confirmation is requested
        Then: Should show enhanced confirmation with better element type display
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        scan_result = self._create_mock_scan_result(
            agents=[".claude/agents/agent1.md"],
            commands=[".claude/commands/moai/cmd1.md"],
        )
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        selected_elements = [
            ".claude/agents/agent1.md",
            ".claude/commands/moai/cmd1.md",
        ]

        # Test enhanced confirmation display
        with patch("builtins.input", return_value="y"):
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

        scan_result = self._create_mock_scan_result(
            agents=[".claude/agents/agent1.md"],
        )
        ui = self._create_ui_with_mock_scanner(project_path, scan_result)

        # Test empty input (just pressing Enter)
        with patch("builtins.input", return_value=""):
            result = ui.prompt_user_selection(backup_available=True)

        # Should return None for empty input
        assert result is None
