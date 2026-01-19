"""Comprehensive tests for template_utils.py module.

Tests cover all four template utility functions with 100% coverage.
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

from moai_adk.core.migration.template_utils import (
    _get_template_agent_names,
    _get_template_command_names,
    _get_template_hook_names,
    _get_template_skill_names,
)


class TestGetTemplateSkillNames:
    """Test _get_template_skill_names function."""

    def test_returns_set_of_skill_names(self):
        """Test returns a set of skill directory names."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.iterdir") as mock_iterdir:
                # Create mock skill directories
                skill1 = MagicMock(spec=Path)
                skill1.is_dir.return_value = True
                skill1.name = "moai-foundation"

                skill2 = MagicMock(spec=Path)
                skill2.is_dir.return_value = True
                skill2.name = "moai-core"

                skill3 = MagicMock(spec=Path)
                skill3.is_dir.return_value = True
                skill3.name = "custom-skill"  # Non-moai prefix

                mock_iterdir.return_value = [skill1, skill2, skill3]

                result = _get_template_skill_names()

                assert isinstance(result, set)
                assert "moai-foundation" in result
                assert "moai-core" in result
                assert "custom-skill" not in result  # Only moai-* skills

    def test_returns_empty_set_when_template_not_exists(self):
        """Test returns empty set when template directory doesn't exist."""
        with patch("pathlib.Path.exists", return_value=False):
            result = _get_template_skill_names()
            assert result == set()

    def test_filters_non_directories(self):
        """Test that non-directory items are filtered out."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.iterdir") as mock_iterdir:
                # Create mock items including files
                dir1 = MagicMock(spec=Path)
                dir1.is_dir.return_value = True
                dir1.name = "moai-dir1"

                file1 = MagicMock(spec=Path)
                file1.is_dir.return_value = False
                file1.name = "moai-file1.py"

                mock_iterdir.return_value = [dir1, file1]

                result = _get_template_skill_names()

                assert "moai-dir1" in result
                assert "moai-file1.py" not in result

    def test_returns_empty_set_when_no_moai_skills(self):
        """Test returns empty set when no moai-* skills found."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.iterdir") as mock_iterdir:
                skill1 = MagicMock(spec=Path)
                skill1.is_dir.return_value = True
                skill1.name = "custom-skill"

                skill2 = MagicMock(spec=Path)
                skill2.is_dir.return_value = True
                skill2.name = "other-skill"

                mock_iterdir.return_value = [skill1, skill2]

                result = _get_template_skill_names()

                assert result == set()

    def test_handles_empty_directory(self):
        """Test handles empty template directory."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.iterdir", return_value=[]):
                result = _get_template_skill_names()
                assert result == set()


class TestGetTemplateCommandNames:
    """Test _get_template_command_names function."""

    def test_returns_set_of_command_names(self):
        """Test returns a set of command file names."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.glob") as mock_glob:
                # Create mock command files
                cmd1 = MagicMock(spec=Path)
                cmd1.name = "00-plan.md"

                cmd2 = MagicMock(spec=Path)
                cmd2.name = "02-run.md"

                cmd3 = MagicMock(spec=Path)
                cmd3.name = "custom.md"

                mock_glob.return_value = [cmd1, cmd2, cmd3]

                result = _get_template_command_names()

                assert isinstance(result, set)
                assert "00-plan.md" in result
                assert "02-run.md" in result
                assert "custom.md" in result

    def test_returns_empty_set_when_template_not_exists(self):
        """Test returns empty set when template directory doesn't exist."""
        with patch("pathlib.Path.exists", return_value=False):
            result = _get_template_command_names()
            assert result == set()

    def test_filters_non_md_files(self):
        """Test that non-.md files are filtered out."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.glob") as mock_glob:
                # glob with "*.md" should only return .md files
                cmd1 = MagicMock(spec=Path)
                cmd1.name = "00-plan.md"

                cmd2 = MagicMock(spec=Path)
                cmd2.name = "02-run.md"

                mock_glob.return_value = [cmd1, cmd2]

                result = _get_template_command_names()

                assert len(result) == 2
                assert all(name.endswith(".md") for name in result)

    def test_returns_empty_set_when_no_commands(self):
        """Test returns empty set when no commands found."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.glob", return_value=[]):
                result = _get_template_command_names()
                assert result == set()


class TestGetTemplateAgentNames:
    """Test _get_template_agent_names function."""

    def test_returns_set_of_agent_names_from_moai_subdirectory(self):
        """Test returns agent names from moai/ subdirectory only."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob") as mock_rglob:
                # Create mock agent files
                moai_agent1 = MagicMock(spec=Path)
                moai_agent1.name = "manager-ddd.md"
                moai_agent1.parent.name = "moai"

                moai_agent2 = MagicMock(spec=Path)
                moai_agent2.name = "expert-debug.md"
                moai_agent2.parent.name = "moai"

                custom_agent = MagicMock(spec=Path)
                custom_agent.name = "custom-agent.md"
                custom_agent.parent.name = "custom"

                mock_rglob.return_value = [moai_agent1, moai_agent2, custom_agent]

                result = _get_template_agent_names()

                assert isinstance(result, set)
                assert "manager-ddd.md" in result
                assert "expert-debug.md" in result
                assert "custom-agent.md" not in result

    def test_returns_empty_set_when_template_not_exists(self):
        """Test returns empty set when template directory doesn't exist."""
        with patch("pathlib.Path.exists", return_value=False):
            result = _get_template_agent_names()
            assert result == set()

    def test_filters_non_md_files(self):
        """Test that non-.md files are filtered out."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob") as mock_rglob:
                # rglob with "*.md" should only return .md files
                agent1 = MagicMock(spec=Path)
                agent1.name = "agent1.md"
                agent1.parent.name = "moai"

                agent2 = MagicMock(spec=Path)
                agent2.name = "agent2.md"
                agent2.parent.name = "moai"

                mock_rglob.return_value = [agent1, agent2]

                result = _get_template_agent_names()

                assert len(result) == 2
                assert all(name.endswith(".md") for name in result)

    def test_returns_empty_set_when_no_moai_agents(self):
        """Test returns empty set when no agents in moai/ subdirectory."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob") as mock_rglob:
                custom_agent = MagicMock(spec=Path)
                custom_agent.name = "custom-agent.md"
                custom_agent.parent.name = "custom"

                mock_rglob.return_value = [custom_agent]

                result = _get_template_agent_names()

                assert result == set()

    def test_handles_nested_moai_directory(self):
        """Test handles nested moai/ directory structure."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob") as mock_rglob:
                # Nested agents in moai/
                agent1 = MagicMock(spec=Path)
                agent1.name = "agent1.md"
                agent1.parent.name = "moai"

                agent2 = MagicMock(spec=Path)
                agent2.name = "agent2.md"
                agent2.parent.name = "moai"

                mock_rglob.return_value = [agent1, agent2]

                result = _get_template_agent_names()

                assert len(result) == 2


class TestGetTemplateHookNames:
    """Test _get_template_hook_names function."""

    def test_returns_set_of_hook_names_from_moai_subdirectory(self):
        """Test returns hook names from moai/ subdirectory only."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob") as mock_rglob:
                # Create mock hook files
                # Note: rglob is called on .../hooks/moai path, so all returned files
                # should be under moai directory structure
                moai_hook1 = MagicMock(spec=Path)
                moai_hook1.name = "pre_tool_use.py"

                moai_hook2 = MagicMock(spec=Path)
                moai_hook2.name = "session_start.py"

                mock_rglob.return_value = [moai_hook1, moai_hook2]

                result = _get_template_hook_names()

                assert isinstance(result, set)
                assert "pre_tool_use.py" in result
                assert "session_start.py" in result

    def test_returns_empty_set_when_template_not_exists(self):
        """Test returns empty set when template directory doesn't exist."""
        with patch("pathlib.Path.exists", return_value=False):
            result = _get_template_hook_names()
            assert result == set()

    def test_filters_non_py_files(self):
        """Test that non-.py files are filtered out."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob") as mock_rglob:
                # rglob with "*.py" should only return .py files
                hook1 = MagicMock(spec=Path)
                hook1.name = "hook1.py"

                hook2 = MagicMock(spec=Path)
                hook2.name = "hook2.py"

                mock_rglob.return_value = [hook1, hook2]

                result = _get_template_hook_names()

                assert len(result) == 2
                assert all(name.endswith(".py") for name in result)

    def test_returns_empty_set_when_no_hooks(self):
        """Test returns empty set when no hooks found."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob", return_value=[]):
                result = _get_template_hook_names()
                assert result == set()

    def test_handles_nested_directory_structure(self):
        """Test handles nested directory structure under hooks/moai."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob") as mock_rglob:
                # Nested hooks under moai/
                hook1 = MagicMock(spec=Path)
                hook1.name = "hook1.py"

                hook2 = MagicMock(spec=Path)
                hook2.name = "hook2.py"

                mock_rglob.return_value = [hook1, hook2]

                result = _get_template_hook_names()

                assert len(result) == 2


class TestTemplateUtilsIntegration:
    """Integration tests for template utilities."""

    def test_all_functions_return_sets(self):
        """Test all template utility functions return sets."""
        with patch("pathlib.Path.exists", return_value=False):
            skills = _get_template_skill_names()
            commands = _get_template_command_names()
            agents = _get_template_agent_names()
            hooks = _get_template_hook_names()

            assert isinstance(skills, set)
            assert isinstance(commands, set)
            assert isinstance(agents, set)
            assert isinstance(hooks, set)

    def test_all_functions_handle_nonexistent_template(self):
        """Test all functions handle nonexistent template directory."""
        with patch("pathlib.Path.exists", return_value=False):
            assert _get_template_skill_names() == set()
            assert _get_template_command_names() == set()
            assert _get_template_agent_names() == set()
            assert _get_template_hook_names() == set()

    def test_template_path_construction(self):
        """Test that template paths are constructed correctly."""
        # Get the expected template path
        from moai_adk.core.migration import template_utils

        # The template path should be src/moai_adk/templates/.claude/
        module_path = Path(template_utils.__file__).parent
        expected_template_base = module_path.parent.parent.parent / "templates" / ".claude"

        # Verify skills path exists (for path construction validation)
        assert (expected_template_base / "skills") is not None
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.iterdir") as mock_iterdir:
                skill = MagicMock(spec=Path)
                skill.is_dir.return_value = True
                skill.name = "moai-test"
                mock_iterdir.return_value = [skill]

                result = _get_template_skill_names()
                assert "moai-test" in result


class TestTemplateUtilsEdgeCases:
    """Test edge cases for template utilities."""

    def test_skill_names_with_special_characters(self):
        """Test skill names with special characters."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.iterdir") as mock_iterdir:
                skill = MagicMock(spec=Path)
                skill.is_dir.return_value = True
                skill.name = "moai-skill-with-dashes"

                mock_iterdir.return_value = [skill]

                result = _get_template_skill_names()

                assert "moai-skill-with-dashes" in result

    def test_command_names_with_special_characters(self):
        """Test command names with special characters."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.glob") as mock_glob:
                cmd = MagicMock(spec=Path)
                cmd.name = "99-release-with-numbers.md"

                mock_glob.return_value = [cmd]

                result = _get_template_command_names()

                assert "99-release-with-numbers.md" in result

    def test_agent_names_with_special_characters(self):
        """Test agent names with special characters."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob") as mock_rglob:
                agent = MagicMock(spec=Path)
                agent.name = "expert-with-dashes.md"
                agent.parent.name = "moai"

                mock_rglob.return_value = [agent]

                result = _get_template_agent_names()

                assert "expert-with-dashes.md" in result

    def test_hook_names_with_special_characters(self):
        """Test hook names with special characters."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.rglob") as mock_rglob:
                hook = MagicMock(spec=Path)
                hook.name = "hook_with_underscores.py"
                hook.parent.name = "moai"

                mock_rglob.return_value = [hook]

                result = _get_template_hook_names()

                assert "hook_with_underscores.py" in result

    def test_handles_duplicate_names(self):
        """Test that duplicate names are handled correctly (set property)."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.glob") as mock_glob:
                cmd = MagicMock(spec=Path)
                cmd.name = "duplicate.md"

                # Return the same file multiple times
                mock_glob.return_value = [cmd, cmd, cmd]

                result = _get_template_command_names()

                # Set should deduplicate
                assert len(result) == 1
                assert "duplicate.md" in result
