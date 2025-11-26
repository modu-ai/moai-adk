"""Tests for custom files detection and restoration in update.py

This test module covers detection and restoration of custom user files:
- Custom commands (.md files in .claude/commands/moai/)
- Custom agents (files in .claude/agents/)
- Custom hooks (.py files in .claude/hooks/moai/)

Test Coverage Strategy:
1. Detection functions for each file type
2. Helper functions for template file comparison
3. Restoration logic with error handling
4. Integration into template sync workflow
"""

from unittest.mock import patch


class TestDetectCustomCommands:
    """Test detection of custom commands in .claude/commands/moai/"""

    def test_detect_custom_commands_finds_user_files(self, tmp_path):
        """Given: Project with custom .md files in .claude/commands/moai/
        When: _detect_custom_commands() is called
        Then: Returns list of custom command files excluding template files
        """
        # Setup project structure
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        commands_dir = project_path / ".claude" / "commands" / "moai"
        commands_dir.mkdir(parents=True)

        # Create template command files
        (commands_dir / "template-command-1.md").write_text("Template content")
        (commands_dir / "template-command-2.md").write_text("Template content")

        # Create custom user command files
        (commands_dir / "custom-command-1.md").write_text("User custom content")
        (commands_dir / "custom-command-2.md").write_text("User custom content")

        # Mock template commands
        template_commands = {"template-command-1.md", "template-command-2.md"}

        from moai_adk.cli.commands.update import _detect_custom_commands

        result = _detect_custom_commands(project_path, template_commands)

        # Should return only custom commands
        assert set(result) == {"custom-command-1.md", "custom-command-2.md"}

    def test_detect_custom_commands_missing_directory(self, tmp_path):
        """Given: Project with no .claude/commands/moai/ directory
        When: _detect_custom_commands() is called
        Then: Returns empty list
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        from moai_adk.cli.commands.update import _detect_custom_commands

        result = _detect_custom_commands(project_path, set())

        assert result == []

    def test_detect_custom_commands_empty_directory(self, tmp_path):
        """Given: Project with empty .claude/commands/moai/ directory
        When: _detect_custom_commands() is called
        Then: Returns empty list
        """
        project_path = tmp_path / "test_project"
        commands_dir = project_path / ".claude" / "commands" / "moai"
        commands_dir.mkdir(parents=True)

        from moai_adk.cli.commands.update import _detect_custom_commands

        result = _detect_custom_commands(project_path, set())

        assert result == []

    def test_detect_custom_commands_all_template(self, tmp_path):
        """Given: Project with only template command files
        When: _detect_custom_commands() is called
        Then: Returns empty list
        """
        project_path = tmp_path / "test_project"
        commands_dir = project_path / ".claude" / "commands" / "moai"
        commands_dir.mkdir(parents=True)

        # Create template commands
        (commands_dir / "template-cmd.md").write_text("Template")

        template_commands = {"template-cmd.md"}

        from moai_adk.cli.commands.update import _detect_custom_commands

        result = _detect_custom_commands(project_path, template_commands)

        assert result == []

    def test_detect_custom_commands_sorted_output(self, tmp_path):
        """Given: Project with multiple custom command files
        When: _detect_custom_commands() is called
        Then: Returns sorted list of custom commands
        """
        project_path = tmp_path / "test_project"
        commands_dir = project_path / ".claude" / "commands" / "moai"
        commands_dir.mkdir(parents=True)

        # Create files in non-sorted order
        (commands_dir / "zebra-cmd.md").write_text("Z")
        (commands_dir / "apple-cmd.md").write_text("A")
        (commands_dir / "banana-cmd.md").write_text("B")

        from moai_adk.cli.commands.update import _detect_custom_commands

        result = _detect_custom_commands(project_path, set())

        # Should be sorted
        assert result == ["apple-cmd.md", "banana-cmd.md", "zebra-cmd.md"]


class TestDetectCustomAgents:
    """Test detection of custom agents in .claude/agents/"""

    def test_detect_custom_agents_finds_user_files(self, tmp_path):
        """Given: Project with custom agent files in .claude/agents/
        When: _detect_custom_agents() is called
        Then: Returns list of custom agent files excluding template files
        """
        project_path = tmp_path / "test_project"
        agents_dir = project_path / ".claude" / "agents"
        agents_dir.mkdir(parents=True)

        # Create template agent files
        (agents_dir / "template-agent-1.md").write_text("Template")
        (agents_dir / "template-agent-2.yaml").write_text("Template")

        # Create custom user agent files
        (agents_dir / "custom-agent-1.md").write_text("User custom")
        (agents_dir / "custom-agent-2.yaml").write_text("User custom")

        template_agents = {"template-agent-1.md", "template-agent-2.yaml"}

        from moai_adk.cli.commands.update import _detect_custom_agents

        result = _detect_custom_agents(project_path, template_agents)

        assert set(result) == {"custom-agent-1.md", "custom-agent-2.yaml"}

    def test_detect_custom_agents_missing_directory(self, tmp_path):
        """Given: Project with no .claude/agents/ directory
        When: _detect_custom_agents() is called
        Then: Returns empty list
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        from moai_adk.cli.commands.update import _detect_custom_agents

        result = _detect_custom_agents(project_path, set())

        assert result == []

    def test_detect_custom_agents_empty_directory(self, tmp_path):
        """Given: Project with empty .claude/agents/ directory
        When: _detect_custom_agents() is called
        Then: Returns empty list
        """
        project_path = tmp_path / "test_project"
        agents_dir = project_path / ".claude" / "agents"
        agents_dir.mkdir(parents=True)

        from moai_adk.cli.commands.update import _detect_custom_agents

        result = _detect_custom_agents(project_path, set())

        assert result == []

    def test_detect_custom_agents_multiple_formats(self, tmp_path):
        """Given: Project with custom agents in various formats (.md, .yaml, .json)
        When: _detect_custom_agents() is called
        Then: Returns all custom agent files regardless of format
        """
        project_path = tmp_path / "test_project"
        agents_dir = project_path / ".claude" / "agents"
        agents_dir.mkdir(parents=True)

        # Create agents in different formats
        (agents_dir / "agent1.md").write_text("Markdown agent")
        (agents_dir / "agent2.yaml").write_text("YAML agent")
        (agents_dir / "agent3.json").write_text('{"agent": "json"}')

        from moai_adk.cli.commands.update import _detect_custom_agents

        result = _detect_custom_agents(project_path, set())

        assert set(result) == {"agent1.md", "agent2.yaml", "agent3.json"}


class TestDetectCustomHooks:
    """Test detection of custom hooks in .claude/hooks/moai/"""

    def test_detect_custom_hooks_finds_user_files(self, tmp_path):
        """Given: Project with custom .py files in .claude/hooks/moai/
        When: _detect_custom_hooks() is called
        Then: Returns list of custom hook files excluding template files
        """
        project_path = tmp_path / "test_project"
        hooks_dir = project_path / ".claude" / "hooks" / "moai"
        hooks_dir.mkdir(parents=True)

        # Create template hook files
        (hooks_dir / "template_hook_1.py").write_text("# template hook 1")
        (hooks_dir / "template_hook_2.py").write_text("# template hook 2")

        # Create custom user hook files
        (hooks_dir / "custom_hook_1.py").write_text("# custom hook 1")
        (hooks_dir / "custom_hook_2.py").write_text("# custom hook 2")

        template_hooks = {"template_hook_1.py", "template_hook_2.py"}

        from moai_adk.cli.commands.update import _detect_custom_hooks

        result = _detect_custom_hooks(project_path, template_hooks)

        assert set(result) == {"custom_hook_1.py", "custom_hook_2.py"}

    def test_detect_custom_hooks_missing_directory(self, tmp_path):
        """Given: Project with no .claude/hooks/moai/ directory
        When: _detect_custom_hooks() is called
        Then: Returns empty list
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        from moai_adk.cli.commands.update import _detect_custom_hooks

        result = _detect_custom_hooks(project_path, set())

        assert result == []

    def test_detect_custom_hooks_ignores_non_py_files(self, tmp_path):
        """Given: Project with mixed file types in .claude/hooks/moai/
        When: _detect_custom_hooks() is called
        Then: Returns only .py files
        """
        project_path = tmp_path / "test_project"
        hooks_dir = project_path / ".claude" / "hooks" / "moai"
        hooks_dir.mkdir(parents=True)

        # Create various file types
        (hooks_dir / "hook1.py").write_text("# python hook")
        (hooks_dir / "hook2.md").write_text("# markdown")
        (hooks_dir / "hook3.txt").write_text("text file")

        from moai_adk.cli.commands.update import _detect_custom_hooks

        result = _detect_custom_hooks(project_path, set())

        # Should only include .py files
        assert result == ["hook1.py"]

    def test_detect_custom_hooks_sorted_output(self, tmp_path):
        """Given: Project with multiple custom hook files
        When: _detect_custom_hooks() is called
        Then: Returns sorted list of custom hooks
        """
        project_path = tmp_path / "test_project"
        hooks_dir = project_path / ".claude" / "hooks" / "moai"
        hooks_dir.mkdir(parents=True)

        # Create files in non-sorted order
        (hooks_dir / "z_hook.py").write_text("z")
        (hooks_dir / "a_hook.py").write_text("a")
        (hooks_dir / "m_hook.py").write_text("m")

        from moai_adk.cli.commands.update import _detect_custom_hooks

        result = _detect_custom_hooks(project_path, set())

        assert result == ["a_hook.py", "m_hook.py", "z_hook.py"]


class TestGetTemplateFileNames:
    """Test helper functions for template file detection"""

    def test_get_template_command_names(self, tmp_path):
        """Given: Package template with command files
        When: _get_template_command_names() is called
        Then: Returns set of template command file names
        """
        # This will test actual template detection from installed package
        from moai_adk.cli.commands.update import _get_template_command_names

        result = _get_template_command_names()

        # Should return a set of command file names
        assert isinstance(result, set)
        # Template should have at least the standard commands
        # (This may vary by installation)

    def test_get_template_agent_names(self, tmp_path):
        """Given: Package template with agent files
        When: _get_template_agent_names() is called
        Then: Returns set of template agent file names
        """
        from moai_adk.cli.commands.update import _get_template_agent_names

        result = _get_template_agent_names()

        # Should return a set of agent file names
        assert isinstance(result, set)

    def test_get_template_hook_names(self, tmp_path):
        """Given: Package template with hook files
        When: _get_template_hook_names() is called
        Then: Returns set of template hook file names
        """
        from moai_adk.cli.commands.update import _get_template_hook_names

        result = _get_template_hook_names()

        # Should return a set of hook file names
        assert isinstance(result, set)


class TestGroupCustomFiles:
    """Test grouping custom files by type"""

    def test_group_custom_files_by_type(self, tmp_path):
        """Given: Custom files of different types
        When: _group_custom_files_by_type() is called
        Then: Returns grouped structure with commands, agents, hooks
        """
        from moai_adk.cli.commands.update import _group_custom_files_by_type

        grouped = _group_custom_files_by_type(
            custom_commands=["cmd1.md", "cmd2.md"], custom_agents=["agent1.md"], custom_hooks=["hook1.py"]
        )

        assert grouped["commands"] == ["cmd1.md", "cmd2.md"]
        assert grouped["agents"] == ["agent1.md"]
        assert grouped["hooks"] == ["hook1.py"]

    def test_group_custom_files_empty_all(self, tmp_path):
        """Given: No custom files
        When: _group_custom_files_by_type() is called
        Then: Returns empty grouped structure
        """
        from moai_adk.cli.commands.update import _group_custom_files_by_type

        grouped = _group_custom_files_by_type([], [], [])

        assert grouped["commands"] == []
        assert grouped["agents"] == []
        assert grouped["hooks"] == []


class TestPromptCustomFilesRestore:
    """Test UI prompts for custom files restoration"""

    def test_prompt_custom_files_restore_no_files(self):
        """Given: No custom files to restore
        When: _prompt_custom_files_restore() is called
        Then: Returns empty selections and skips UI
        """
        from moai_adk.cli.commands.update import _prompt_custom_files_restore

        result = _prompt_custom_files_restore(custom_commands=[], custom_agents=[], custom_hooks=[], yes=False)

        # Should return empty selections
        assert result == {"commands": [], "agents": [], "hooks": []}

    def test_prompt_custom_files_restore_yes_mode(self):
        """Given: --yes flag enabled
        When: _prompt_custom_files_restore() is called
        Then: Skips UI and returns empty selections (no restoration)
        """
        from moai_adk.cli.commands.update import _prompt_custom_files_restore

        result = _prompt_custom_files_restore(
            custom_commands=["cmd1.md", "cmd2.md"], custom_agents=["agent1.md"], custom_hooks=["hook1.py"], yes=True
        )

        # In --yes mode, skip restoration (return empty to be safe)
        assert result == {"commands": [], "agents": [], "hooks": []}

    @patch("questionary.checkbox")
    def test_prompt_custom_files_restore_user_selects_all(self, mock_checkbox):
        """Given: User selects all custom files via questionary
        When: _prompt_custom_files_restore() is called
        Then: Returns all selected files grouped by type
        """
        from moai_adk.cli.commands.update import _prompt_custom_files_restore

        # Mock questionary to return selections
        mock_checkbox.return_value.ask.return_value = {
            "commands": ["cmd1.md", "cmd2.md"],
            "agents": ["agent1.md"],
            "hooks": ["hook1.py"],
        }

        result = _prompt_custom_files_restore(
            custom_commands=["cmd1.md", "cmd2.md"], custom_agents=["agent1.md"], custom_hooks=["hook1.py"], yes=False
        )

        # Verify result structure
        assert isinstance(result, dict)
        assert "commands" in result
        assert "agents" in result
        assert "hooks" in result

    @patch("questionary.checkbox")
    def test_prompt_custom_files_restore_user_selects_none(self, mock_checkbox):
        """Given: User deselects all files in questionary
        When: _prompt_custom_files_restore() is called
        Then: Returns empty selections
        """
        from moai_adk.cli.commands.update import _prompt_custom_files_restore

        # Mock questionary to return no selections
        mock_checkbox.return_value.ask.return_value = None

        result = _prompt_custom_files_restore(
            custom_commands=["cmd1.md"], custom_agents=["agent1.md"], custom_hooks=["hook1.py"], yes=False
        )

        # When user cancels or selects nothing
        assert result == {"commands": [], "agents": [], "hooks": []}


class TestRestoreCustomFiles:
    """Test restoration of custom files from backup"""

    def test_restore_custom_files_success(self, tmp_path):
        """Given: Custom files in backup
        When: _restore_custom_files() is called with file selections
        Then: Files are copied to project successfully
        """
        # Setup backup with custom files
        backup_path = tmp_path / "backup"
        project_path = tmp_path / "project"
        project_path.mkdir()

        # Create backup structure
        backup_commands = backup_path / ".claude" / "commands" / "moai"
        backup_commands.mkdir(parents=True)
        (backup_commands / "custom.md").write_text("custom command content")

        from moai_adk.cli.commands.update import _restore_custom_files

        result = _restore_custom_files(
            project_path=project_path,
            backup_path=backup_path,
            selected_commands=["custom.md"],
            selected_agents=[],
            selected_hooks=[],
        )

        # Should succeed
        assert result is True

        # File should be in project
        restored = project_path / ".claude" / "commands" / "moai" / "custom.md"
        assert restored.exists()
        assert restored.read_text() == "custom command content"

    def test_restore_custom_files_missing_in_backup(self, tmp_path):
        """Given: Selected files missing in backup
        When: _restore_custom_files() is called
        Then: Returns False and reports missing files
        """
        backup_path = tmp_path / "backup"
        project_path = tmp_path / "project"
        project_path.mkdir()

        from moai_adk.cli.commands.update import _restore_custom_files

        result = _restore_custom_files(
            project_path=project_path,
            backup_path=backup_path,
            selected_commands=["nonexistent.md"],
            selected_agents=[],
            selected_hooks=[],
        )

        # Should handle gracefully
        assert result is False

    def test_restore_custom_files_all_types(self, tmp_path):
        """Given: Custom files of all types in backup
        When: _restore_custom_files() is called with selections
        Then: All file types are restored correctly
        """
        backup_path = tmp_path / "backup"
        project_path = tmp_path / "project"
        project_path.mkdir()

        # Create backup structure with all types
        backup_commands = backup_path / ".claude" / "commands" / "moai"
        backup_commands.mkdir(parents=True)
        (backup_commands / "cmd.md").write_text("command")

        backup_agents = backup_path / ".claude" / "agents"
        backup_agents.mkdir(parents=True)
        (backup_agents / "agent.md").write_text("agent")

        backup_hooks = backup_path / ".claude" / "hooks" / "moai"
        backup_hooks.mkdir(parents=True)
        (backup_hooks / "hook.py").write_text("hook")

        from moai_adk.cli.commands.update import _restore_custom_files

        result = _restore_custom_files(
            project_path=project_path,
            backup_path=backup_path,
            selected_commands=["cmd.md"],
            selected_agents=["agent.md"],
            selected_hooks=["hook.py"],
        )

        assert result is True
        assert (project_path / ".claude" / "commands" / "moai" / "cmd.md").exists()
        assert (project_path / ".claude" / "agents" / "agent.md").exists()
        assert (project_path / ".claude" / "hooks" / "moai" / "hook.py").exists()

    def test_restore_custom_files_empty_selections(self, tmp_path):
        """Given: Empty file selections
        When: _restore_custom_files() is called with empty lists
        Then: Returns True without error
        """
        backup_path = tmp_path / "backup"
        project_path = tmp_path / "project"
        project_path.mkdir()

        from moai_adk.cli.commands.update import _restore_custom_files

        result = _restore_custom_files(
            project_path=project_path,
            backup_path=backup_path,
            selected_commands=[],
            selected_agents=[],
            selected_hooks=[],
        )

        # Should succeed (no-op)
        assert result is True
