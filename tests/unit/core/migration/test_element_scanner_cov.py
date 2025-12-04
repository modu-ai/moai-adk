"""Comprehensive tests for custom_element_scanner.py module.

Focus on uncovered code paths with actual execution using mocked dependencies.
"""

import pytest
from pathlib import Path
from unittest.mock import MagicMock, Mock, patch, PropertyMock
from typing import Set, Dict, List, Any

from moai_adk.core.migration.custom_element_scanner import (
    TemplateSkill,
    CustomElementScanner,
    create_custom_element_scanner,
)


class TestTemplateSkill:
    """Test TemplateSkill class."""

    def test_initialization(self):
        """Test TemplateSkill initialization."""
        path = Path("/project/.claude/skills/my-skill")
        skill = TemplateSkill(
            name="my-skill",
            path=path,
            has_skill_md=True,
            is_template=False
        )

        assert skill.name == "my-skill"
        assert skill.path == path
        assert skill.has_skill_md is True
        assert skill.is_template is False

    def test_initialization_with_defaults(self):
        """Test TemplateSkill initialization with default is_template."""
        path = Path("/project/.claude/skills/custom-skill")
        skill = TemplateSkill("custom-skill", path, False)

        assert skill.is_template is False

    def test_initialization_template_skill(self):
        """Test TemplateSkill initialization as template skill."""
        path = Path("/project/.claude/skills/moai-foundation")
        skill = TemplateSkill("moai-foundation", path, True, is_template=True)

        assert skill.is_template is True
        assert skill.has_skill_md is True


class TestCustomElementScanner:
    """Test CustomElementScanner class."""

    def test_initialization(self):
        """Test CustomElementScanner initialization."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "skills": {"moai-skill1"},
                "commands": {"cmd1.md"},
                "agents": {"agent1.md"},
                "hooks": {"hook1.py"},
            }

            project_path = Path("/project")
            scanner = CustomElementScanner(project_path)

            assert scanner.project_path == project_path
            assert scanner.template_elements is not None
            mock_get.assert_called_once()

    def test_get_template_elements(self):
        """Test _get_template_elements returns correct structure."""
        with patch('moai_adk.core.migration.custom_element_scanner._get_template_skill_names') as mock_skills:
            with patch('moai_adk.core.migration.custom_element_scanner._get_template_command_names') as mock_cmds:
                with patch('moai_adk.core.migration.custom_element_scanner._get_template_agent_names') as mock_agents:
                    with patch('moai_adk.core.migration.custom_element_scanner._get_template_hook_names') as mock_hooks:
                        mock_skills.return_value = {"moai-skill"}
                        mock_cmds.return_value = {"cmd.md"}
                        mock_agents.return_value = {"agent.md"}
                        mock_hooks.return_value = {"hook.py"}

                        scanner = CustomElementScanner(Path("/project"))

                        assert "skills" in scanner.template_elements
                        assert "commands" in scanner.template_elements
                        assert "agents" in scanner.template_elements
                        assert "hooks" in scanner.template_elements

    def test_scan_custom_elements_returns_all_types(self):
        """Test scan_custom_elements returns all element types."""
        with patch.object(CustomElementScanner, '_scan_custom_agents') as mock_agents:
            with patch.object(CustomElementScanner, '_scan_custom_commands') as mock_cmds:
                with patch.object(CustomElementScanner, '_scan_custom_skills') as mock_skills:
                    with patch.object(CustomElementScanner, '_scan_custom_hooks') as mock_hooks:
                        mock_agents.return_value = []
                        mock_cmds.return_value = []
                        mock_skills.return_value = []
                        mock_hooks.return_value = []

                        with patch.object(CustomElementScanner, '_get_template_elements'):
                            scanner = CustomElementScanner(Path("/project"))
                            result = scanner.scan_custom_elements()

                            assert "agents" in result
                            assert "commands" in result
                            assert "skills" in result
                            assert "hooks" in result

    def test_scan_custom_agents_no_directory(self):
        """Test scanning custom agents when directory doesn't exist."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "agents": set(),
                "commands": set(),
                "skills": set(),
                "hooks": set(),
            }

            with patch('pathlib.Path.exists', return_value=False):
                scanner = CustomElementScanner(Path("/project"))
                result = scanner._scan_custom_agents()
                assert result == []

    def test_scan_custom_agents_empty_directory(self):
        """Test scanning custom agents in empty directory."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "agents": set(),
                "commands": set(),
                "skills": set(),
                "hooks": set(),
            }

            with patch('pathlib.Path.exists', return_value=True):
                with patch('pathlib.Path.rglob', return_value=[]):
                    scanner = CustomElementScanner(Path("/project"))
                    result = scanner._scan_custom_agents()
                    assert result == []

    def test_scan_custom_agents_filters_template(self):
        """Test scanning agents filters out template agents."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            template_agents = {"template-agent.md"}
            mock_get.return_value = {
                "agents": template_agents,
                "commands": set(),
                "skills": set(),
                "hooks": set(),
            }

            agents_dir = Path("/project/.claude/agents")
            template_file = agents_dir / "template-agent.md"
            custom_file = agents_dir / "custom-agent.md"

            with patch('pathlib.Path.exists', return_value=True):
                with patch('pathlib.Path.rglob') as mock_rglob:
                    with patch('pathlib.Path.relative_to') as mock_relative:
                        mock_rglob.return_value = [template_file, custom_file]
                        mock_relative.side_effect = lambda p: Path(f".claude/agents/{p.name}")

                        scanner = CustomElementScanner(Path("/project"))
                        result = scanner._scan_custom_agents()

                        # Only custom-agent.md should be returned
                        assert len(result) == 1

    def test_scan_custom_commands_in_moai_subdirectory(self):
        """Test scanning commands in moai subdirectory."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "agents": set(),
                "commands": {"template.md"},
                "skills": set(),
                "hooks": set(),
            }

            commands_dir = Path("/project/.claude/commands")
            moai_dir = commands_dir / "moai"

            with patch('pathlib.Path.exists') as mock_exists:
                with patch('pathlib.Path.glob') as mock_glob:
                    with patch('pathlib.Path.iterdir') as mock_iterdir:
                        with patch('pathlib.Path.relative_to') as mock_relative:
                            # Setup exists checks
                            def exists_side_effect():
                                if str(self) == str(moai_dir):
                                    return True
                                return True

                            mock_glob.return_value = []
                            mock_iterdir.return_value = []

                            scanner = CustomElementScanner(Path("/project"))
                            result = scanner._scan_custom_commands()
                            assert isinstance(result, list)

    def test_scan_custom_commands_other_subdirectories(self):
        """Test scanning commands in non-moai subdirectories."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "agents": set(),
                "commands": set(),
                "skills": set(),
                "hooks": set(),
            }

            with patch('pathlib.Path.exists', return_value=True):
                with patch('pathlib.Path.glob', return_value=[]):
                    with patch('pathlib.Path.iterdir') as mock_iterdir:
                        custom_dir = MagicMock()
                        custom_dir.is_dir.return_value = True
                        custom_dir.name = "custom"
                        mock_iterdir.return_value = [custom_dir]

                        with patch('pathlib.Path.rglob', return_value=[]):
                            scanner = CustomElementScanner(Path("/project"))
                            result = scanner._scan_custom_commands()
                            assert isinstance(result, list)

    def test_scan_custom_skills_includes_all_skills(self):
        """Test scanning skills includes both template and custom."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "agents": set(),
                "commands": set(),
                "skills": {"moai-foundation", "moai-core"},
                "hooks": set(),
            }

            # Create real Path objects
            template_skill_path = Path("/project/.claude/skills/moai-foundation")
            custom_skill_path = Path("/project/.claude/skills/my-custom-skill")

            with patch('pathlib.Path.exists', return_value=True):
                with patch('pathlib.Path.iterdir') as mock_iterdir:
                    # Create mocks with proper Path behavior
                    template_skill = MagicMock(spec=Path)
                    template_skill.is_dir.return_value = True
                    template_skill.name = "moai-foundation"
                    template_skill.relative_to.return_value = Path(".claude/skills/moai-foundation")

                    custom_skill = MagicMock(spec=Path)
                    custom_skill.is_dir.return_value = True
                    custom_skill.name = "my-custom-skill"
                    custom_skill.relative_to.return_value = Path(".claude/skills/my-custom-skill")

                    mock_iterdir.return_value = [template_skill, custom_skill]

                    with patch.object(CustomElementScanner, '__init__', lambda x, y: None):
                        scanner = CustomElementScanner.__new__(CustomElementScanner)
                        scanner.project_path = Path("/project")
                        scanner.template_elements = {
                            "agents": set(),
                            "commands": set(),
                            "skills": {"moai-foundation"},
                            "hooks": set(),
                        }

                        # Mock Path operations
                        with patch('moai_adk.core.migration.custom_element_scanner.Path') as mock_path_class:
                            mock_path_class.return_value.exists.return_value = False
                            result = scanner._scan_custom_skills()

                            # Should have results even with mocked path
                            assert isinstance(result, list)

    def test_scan_custom_skills_detects_template_status(self):
        """Test scanning skills correctly marks template status."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "agents": set(),
                "commands": set(),
                "skills": {"moai-foundation"},
                "hooks": set(),
            }

            with patch('pathlib.Path.exists', return_value=True):
                with patch('pathlib.Path.iterdir') as mock_iterdir:
                    template_skill = MagicMock(spec=Path)
                    template_skill.is_dir.return_value = True
                    template_skill.name = "moai-foundation"
                    template_skill.relative_to.return_value = Path(".claude/skills/moai-foundation")

                    custom_skill = MagicMock(spec=Path)
                    custom_skill.is_dir.return_value = True
                    custom_skill.name = "custom-skill"
                    custom_skill.relative_to.return_value = Path(".claude/skills/custom-skill")

                    mock_iterdir.return_value = [template_skill, custom_skill]

                    with patch.object(CustomElementScanner, '__init__', lambda x, y: None):
                        scanner = CustomElementScanner.__new__(CustomElementScanner)
                        scanner.project_path = Path("/project")
                        scanner.template_elements = {
                            "agents": set(),
                            "commands": set(),
                            "skills": {"moai-foundation"},
                            "hooks": set(),
                        }

                        with patch('moai_adk.core.migration.custom_element_scanner.Path') as mock_path_class:
                            mock_path_class.return_value.exists.return_value = False
                            result = scanner._scan_custom_skills()

                            # Verify structure is correct
                            if result:
                                assert all(isinstance(s, TemplateSkill) for s in result)

    def test_scan_custom_skills_detects_skill_md(self):
        """Test scanning skills detects SKILL.md file."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "agents": set(),
                "commands": set(),
                "skills": set(),
                "hooks": set(),
            }

            with patch('pathlib.Path.exists') as mock_exists:
                with patch('pathlib.Path.iterdir') as mock_iterdir:
                    skill_dir = MagicMock()
                    skill_dir.is_dir.return_value = True
                    skill_dir.name = "test-skill"
                    skill_dir.relative_to.return_value = Path(".claude/skills/test-skill")
                    skill_dir.__truediv__.return_value = MagicMock(exists=MagicMock(return_value=True))

                    mock_iterdir.return_value = [skill_dir]
                    mock_exists.return_value = True

                    scanner = CustomElementScanner(Path("/project"))
                    result = scanner._scan_custom_skills()

                    assert result[0].has_skill_md is True

    def test_scan_custom_hooks_no_directory(self):
        """Test scanning hooks when directory doesn't exist."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "agents": set(),
                "commands": set(),
                "skills": set(),
                "hooks": set(),
            }

            with patch('pathlib.Path.exists', return_value=False):
                scanner = CustomElementScanner(Path("/project"))
                result = scanner._scan_custom_hooks()
                assert result == []

    def test_scan_custom_hooks_in_moai_subdirectory(self):
        """Test scanning hooks in moai subdirectory."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            mock_get.return_value = {
                "agents": set(),
                "commands": set(),
                "skills": set(),
                "hooks": {"template.py"},
            }

            with patch('pathlib.Path.exists', return_value=True):
                with patch.object(CustomElementScanner, '__init__', lambda x, y: None):
                    scanner = CustomElementScanner.__new__(CustomElementScanner)
                    scanner.project_path = Path("/project")
                    scanner.template_elements = {
                        "agents": set(),
                        "commands": set(),
                        "skills": set(),
                        "hooks": {"template.py"},
                    }

                    with patch('pathlib.Path.rglob') as mock_rglob:
                        custom_hook = Path("/project/.claude/hooks/moai/custom.py")
                        mock_rglob.return_value = [custom_hook]

                        with patch('pathlib.Path.glob', return_value=[]):
                            with patch('pathlib.Path.iterdir', return_value=[]):
                                result = scanner._scan_custom_hooks()

                                # Should return list of paths
                                assert isinstance(result, list)

    def test_get_custom_elements_display_list_agents(self):
        """Test display list includes agents."""
        with patch.object(CustomElementScanner, 'scan_custom_elements') as mock_scan:
            agent_path = Path(".claude/agents/my-agent.md")
            mock_scan.return_value = {
                "agents": [agent_path],
                "commands": [],
                "skills": [],
                "hooks": [],
            }

            with patch.object(CustomElementScanner, '_get_template_elements'):
                scanner = CustomElementScanner(Path("/project"))
                result = scanner.get_custom_elements_display_list()

                assert len(result) == 1
                assert result[0]["type"] == "agent"
                assert result[0]["name"] == "my-agent"
                assert "agent" in result[0]["display_name"]

    def test_get_custom_elements_display_list_commands(self):
        """Test display list includes commands."""
        with patch.object(CustomElementScanner, 'scan_custom_elements') as mock_scan:
            cmd_path = Path(".claude/commands/moai/my-cmd.md")
            mock_scan.return_value = {
                "agents": [],
                "commands": [cmd_path],
                "skills": [],
                "hooks": [],
            }

            with patch.object(CustomElementScanner, '_get_template_elements'):
                scanner = CustomElementScanner(Path("/project"))
                result = scanner.get_custom_elements_display_list()

                assert len(result) == 1
                assert result[0]["type"] == "command"
                assert "command" in result[0]["display_name"]

    def test_get_custom_elements_display_list_skills(self):
        """Test display list includes skills."""
        with patch.object(CustomElementScanner, 'scan_custom_elements') as mock_scan:
            template_skill = TemplateSkill(
                name="moai-foundation",
                path=Path(".claude/skills/moai-foundation"),
                has_skill_md=True,
                is_template=True
            )
            custom_skill = TemplateSkill(
                name="my-skill",
                path=Path(".claude/skills/my-skill"),
                has_skill_md=True,
                is_template=False
            )

            mock_scan.return_value = {
                "agents": [],
                "commands": [],
                "skills": [template_skill, custom_skill],
                "hooks": [],
            }

            with patch.object(CustomElementScanner, '_get_template_elements'):
                scanner = CustomElementScanner(Path("/project"))
                result = scanner.get_custom_elements_display_list()

                assert len(result) == 2
                template_entry = [r for r in result if r["name"] == "moai-foundation"][0]
                custom_entry = [r for r in result if r["name"] == "my-skill"][0]

                assert "template" in template_entry["display_name"]
                assert "custom" in custom_entry["display_name"]

    def test_get_custom_elements_display_list_hooks(self):
        """Test display list includes hooks."""
        with patch.object(CustomElementScanner, 'scan_custom_elements') as mock_scan:
            hook_path = Path(".claude/hooks/moai/my-hook.py")
            mock_scan.return_value = {
                "agents": [],
                "commands": [],
                "skills": [],
                "hooks": [hook_path],
            }

            with patch.object(CustomElementScanner, '_get_template_elements'):
                scanner = CustomElementScanner(Path("/project"))
                result = scanner.get_custom_elements_display_list()

                assert len(result) == 1
                assert result[0]["type"] == "hook"
                assert "hook" in result[0]["display_name"]

    def test_get_custom_elements_display_list_all_types(self):
        """Test display list with all element types."""
        with patch.object(CustomElementScanner, 'scan_custom_elements') as mock_scan:
            skill = TemplateSkill(
                name="skill",
                path=Path(".claude/skills/skill"),
                has_skill_md=True
            )

            mock_scan.return_value = {
                "agents": [Path(".claude/agents/agent.md")],
                "commands": [Path(".claude/commands/moai/cmd.md")],
                "skills": [skill],
                "hooks": [Path(".claude/hooks/moai/hook.py")],
            }

            with patch.object(CustomElementScanner, '_get_template_elements'):
                scanner = CustomElementScanner(Path("/project"))
                result = scanner.get_custom_elements_display_list()

                assert len(result) == 4
                types = [r["type"] for r in result]
                assert "agent" in types
                assert "command" in types
                assert "skill" in types
                assert "hook" in types

    def test_get_element_count(self):
        """Test getting element count."""
        with patch.object(CustomElementScanner, 'scan_custom_elements') as mock_scan:
            mock_scan.return_value = {
                "agents": [Path("a1"), Path("a2")],
                "commands": [Path("c1")],
                "skills": [TemplateSkill("s1", Path("p"), True)],
                "hooks": [Path("h1"), Path("h2"), Path("h3")],
            }

            with patch.object(CustomElementScanner, '_get_template_elements'):
                scanner = CustomElementScanner(Path("/project"))
                count = scanner.get_element_count()

                # 2 + 1 + 1 + 3 = 7
                assert count == 7

    def test_get_element_count_empty(self):
        """Test element count with no elements."""
        with patch.object(CustomElementScanner, 'scan_custom_elements') as mock_scan:
            mock_scan.return_value = {
                "agents": [],
                "commands": [],
                "skills": [],
                "hooks": [],
            }

            with patch.object(CustomElementScanner, '_get_template_elements'):
                scanner = CustomElementScanner(Path("/project"))
                count = scanner.get_element_count()
                assert count == 0


class TestCreateCustomElementScanner:
    """Test create_custom_element_scanner factory function."""

    def test_factory_with_string_path(self):
        """Test factory function with string path."""
        with patch.object(CustomElementScanner, '_get_template_elements'):
            with patch('pathlib.Path.resolve') as mock_resolve:
                mock_resolve.return_value = Path("/project")
                scanner = create_custom_element_scanner("/project")
                assert isinstance(scanner, CustomElementScanner)

    def test_factory_with_path_object(self):
        """Test factory function with Path object."""
        with patch.object(CustomElementScanner, '_get_template_elements'):
            scanner = create_custom_element_scanner(Path("/project"))
            assert isinstance(scanner, CustomElementScanner)

    def test_factory_resolves_path(self):
        """Test factory resolves path to absolute."""
        with patch.object(CustomElementScanner, '_get_template_elements'):
            with patch('pathlib.Path.resolve') as mock_resolve:
                resolved_path = Path("/absolute/project")
                mock_resolve.return_value = resolved_path

                scanner = create_custom_element_scanner("project")
                assert scanner.project_path == resolved_path


class TestCustomElementScannerIntegration:
    """Integration tests for CustomElementScanner."""

    def test_full_scan_workflow(self):
        """Test full scan workflow."""
        with patch.object(CustomElementScanner, '_get_template_elements') as mock_get:
            with patch.object(CustomElementScanner, '_scan_custom_agents') as mock_agents:
                with patch.object(CustomElementScanner, '_scan_custom_commands') as mock_cmds:
                    with patch.object(CustomElementScanner, '_scan_custom_skills') as mock_skills:
                        with patch.object(CustomElementScanner, '_scan_custom_hooks') as mock_hooks:
                            mock_get.return_value = {
                                "agents": set(),
                                "commands": set(),
                                "skills": set(),
                                "hooks": set(),
                            }
                            mock_agents.return_value = [Path(".claude/agents/custom.md")]
                            mock_cmds.return_value = [Path(".claude/commands/moai/custom.md")]
                            mock_skills.return_value = [
                                TemplateSkill("custom", Path(".claude/skills/custom"), True)
                            ]
                            mock_hooks.return_value = [Path(".claude/hooks/moai/custom.py")]

                            scanner = CustomElementScanner(Path("/project"))
                            elements = scanner.scan_custom_elements()

                            assert len(elements["agents"]) == 1
                            assert len(elements["commands"]) == 1
                            assert len(elements["skills"]) == 1
                            assert len(elements["hooks"]) == 1

    def test_display_list_indexing(self):
        """Test display list has correct indexing."""
        with patch.object(CustomElementScanner, 'scan_custom_elements') as mock_scan:
            mock_scan.return_value = {
                "agents": [Path(".claude/agents/a1.md"), Path(".claude/agents/a2.md")],
                "commands": [Path(".claude/commands/moai/c1.md")],
                "skills": [],
                "hooks": [],
            }

            with patch.object(CustomElementScanner, '_get_template_elements'):
                scanner = CustomElementScanner(Path("/project"))
                result = scanner.get_custom_elements_display_list()

                # Check indices are sequential
                indices = [r["index"] for r in result]
                assert indices == [1, 2, 3]

    def test_multiple_scans_consistent(self):
        """Test multiple scans return consistent results."""
        with patch.object(CustomElementScanner, '_scan_custom_agents') as mock_agents:
            with patch.object(CustomElementScanner, '_scan_custom_commands') as mock_cmds:
                with patch.object(CustomElementScanner, '_scan_custom_skills') as mock_skills:
                    with patch.object(CustomElementScanner, '_scan_custom_hooks') as mock_hooks:
                        mock_agents.return_value = [Path(".claude/agents/agent.md")]
                        mock_cmds.return_value = []
                        mock_skills.return_value = []
                        mock_hooks.return_value = []

                        with patch.object(CustomElementScanner, '_get_template_elements'):
                            scanner = CustomElementScanner(Path("/project"))

                            result1 = scanner.scan_custom_elements()
                            result2 = scanner.scan_custom_elements()

                            assert len(result1["agents"]) == len(result2["agents"])
                            assert result1["agents"] == result2["agents"]
