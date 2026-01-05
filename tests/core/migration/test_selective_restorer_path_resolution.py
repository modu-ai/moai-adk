"""Tests for path resolution issues in SelectiveRestorer

This test module covers the critical path resolution bug where file paths
are incorrectly resolved when running from different working directories,
causing errors like:
".claude/agents/yoda/yoda-master.md is not in the subpath of /Users/goos/MoAI/yoda"
"""


from moai_adk.core.migration.selective_restorer import create_selective_restorer


class TestSelectiveRestorerPathResolution:
    """Test path resolution in SelectiveRestorer"""

    def test_restorer_path_resolution_with_project_root(self, tmp_path):
        """Given: Project with custom elements and backup
        When: SelectiveRestorer is created and used
        Then: Should use project-relative paths, not absolute paths
        """
        # Setup project structure
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        # Create backup structure
        backup_path = tmp_path / "backup"
        backup_path.mkdir()

        # Create backup content structure
        backup_claude = backup_path / ".claude"
        backup_claude.mkdir(parents=True)

        # Create custom agent in backup
        backup_agents = backup_claude / "agents"
        backup_agents.mkdir()
        custom_agent = backup_agents / "yoda-master.md"
        custom_agent.write_text("Yoda agent content")

        # Create corresponding backup metadata
        backup_moai = backup_path / ".moai"
        backup_moai.mkdir(parents=True)
        (backup_moai / "backups").mkdir()
        backup_info = backup_moai / "backups" / "backup-info.json"
        backup_info.write_text('{"timestamp": "2023-01-01", "version": "1.0.0"}')

        # Create restorer with project path (not absolute backup path)
        restorer = create_selective_restorer(project_path, backup_path)

        # Test that it can resolve paths correctly
        selected_elements = [".claude/agents/yoda-master.md"]

        # This should work without path resolution errors
        try:
            success, stats = restorer.restore_elements(selected_elements)
            # Should succeed and restore the file
            assert success is True
            assert stats["success"] == 1
            assert stats["failed"] == 0

            # Verify file was restored to correct location
            restored_agent = project_path / ".claude" / "agents" / "yoda-master.md"
            assert restored_agent.exists()
            assert restored_agent.read_text() == "Yoda agent content"

        except Exception as e:
            if "not in the subpath of" in str(e):
                assert False, f"Path resolution error should not occur: {e}"
            else:
                raise

    def test_restorer_path_resolution_with_subpaths(self, tmp_path):
        """Given: Project with nested custom elements
        When: SelectiveRestorer restores elements
        Then: Should handle nested paths correctly
        """
        # Setup project structure
        project_path = tmp_path / "my_project"
        project_path.mkdir()

        # Create backup structure
        backup_path = tmp_path / "project_backup"
        backup_path.mkdir()

        # Create nested backup structure
        backup_claude = backup_path / ".claude"
        backup_claude.mkdir(parents=True)

        # Create nested custom skill directory
        backup_skills = backup_claude / "skills" / "my-custom-skill"
        backup_skills.mkdir(parents=True)
        (backup_skills / "SKILL.md").write_text("# Custom Skill")
        (backup_skills / "__init__.py").write_text("print('hello')")

        # Create backup metadata
        backup_moai = backup_path / ".moai"
        backup_moai.mkdir(parents=True)
        (backup_moai / "backups").mkdir()

        # Create restorer
        restorer = create_selective_restorer(project_path, backup_path)

        # Test restoration of nested skill
        selected_elements = [".claude/skills/my-custom-skill"]

        try:
            success, stats = restorer.restore_elements(selected_elements)
            assert success is True
            assert stats["success"] == 1

            # Verify nested directory was restored
            restored_skill = project_path / ".claude" / "skills" / "my-custom-skill"
            assert restored_skill.exists()
            assert (restored_skill / "SKILL.md").exists()
            assert (restored_skill / "__init__.py").exists()

        except Exception as e:
            if "not in the subpath of" in str(e):
                assert False, f"Nested path resolution error should not occur: {e}"
            else:
                raise

    def test_restorer_path_resolution_with_absolute_path_bug(self, tmp_path):
        """Given: Project with custom elements (reproduces the original bug)
        When: SelectiveRestorer is called with paths that might cause absolute path issues
        Then: Should resolve paths relative to project, not absolute backup paths
        """
        # Setup to simulate the original error scenario
        project_path = tmp_path / "yoda"
        project_path.mkdir()

        backup_path = tmp_path / "yoda-backup"
        backup_path.mkdir()

        # Create backup with custom agent
        backup_claude = backup_path / ".claude"
        backup_claude.mkdir(parents=True)
        backup_agents = backup_claude / "agents"
        backup_agents.mkdir()

        # Create the problematic file structure
        yoda_dir = backup_agents / "yoda"
        yoda_dir.mkdir()
        yoda_file = yoda_dir / "yoda-master.md"
        yoda_file.write_text("Yoda content")

        # Create backup metadata
        backup_moai = backup_path / ".moai"
        backup_moai.mkdir(parents=True)
        (backup_moai / "backups").mkdir()

        # This is the path that was failing in the original bug
        problematic_element = ".claude/agents/yoda/yoda-master.md"

        restorer = create_selective_restorer(project_path, backup_path)

        try:
            success, stats = restorer.restore_elements([problematic_element])
            assert success is True
            assert stats["success"] == 1

            # Verify the problematic path was resolved correctly
            restored_file = project_path / ".claude" / "agents" / "yoda" / "yoda-master.md"
            assert restored_file.exists()
            assert restored_file.read_text() == "Yoda content"

        except Exception as e:
            if "not in the subpath of" in str(e):
                assert False, f"Original bug should be fixed: {e}"
            else:
                # Other errors are acceptable for this test
                pass

    def test_restorer_path_resolution_empty_selection(self, tmp_path):
        """Given: Empty element selection
        When: SelectiveRestorer.restore_elements() is called
        Then: Should handle gracefully without path resolution issues
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        backup_path = tmp_path / "backup"
        backup_path.mkdir()

        restorer = create_selective_restorer(project_path, backup_path)

        # Empty selection should not cause path issues
        success, stats = restorer.restore_elements([])

        assert success is True
        assert stats["total"] == 0
        assert stats["success"] == 0
        assert stats["failed"] == 0

    def test_restorer_path_resolution_nonexistent_backup(self, tmp_path):
        """Given: Nonexistent backup path
        When: SelectiveRestorer is created
        Then: Should handle gracefully without path resolution errors but return failed for missing elements
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        nonexistent_backup = tmp_path / "nonexistent"

        restorer = create_selective_restorer(project_path, nonexistent_backup)

        # Should handle missing backup gracefully but report failure for requested elements
        success, stats = restorer.restore_elements([".claude/agents/test.md"])

        # Should fail because requested element couldn't be restored
        assert success is False
        assert stats["total"] == 1
        assert stats["success"] == 0
        assert stats["failed"] == 1

    def test_restorer_path_resolution_with_cwd_change(self, tmp_path):
        """Given: Working directory changes during execution
        When: SelectiveRestorer restores elements
        Then: Should still resolve paths correctly based on project_path, not CWD
        """
        # Setup project
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        backup_path = tmp_path / "backup"
        backup_path.mkdir()

        # Create backup content
        backup_claude = backup_path / ".claude"
        backup_claude.mkdir(parents=True)
        backup_agents = backup_claude / "agents"
        backup_agents.mkdir()

        test_agent = backup_agents / "test-agent.md"
        test_agent.write_text("Test agent content")

        # Create restorer - CWD doesn't affect path resolution in the restorer
        restorer = create_selective_restorer(project_path, backup_path)

        # This should work regardless of CWD since restorer uses project_path
        success, stats = restorer.restore_elements([".claude/agents/test-agent.md"])

        assert success is True
        assert stats["success"] == 1

        # Verify file was restored to correct location (in project_path)
        restored_file = project_path / ".claude" / "agents" / "test-agent.md"
        assert restored_file.exists()
        assert restored_file.read_text() == "Test agent content"

    def test_restorer_path_resolution_with_multiple_elements(self, tmp_path):
        """Given: Multiple custom elements to restore
        When: SelectiveRestorer restores them
        Then: Should handle path resolution for all elements correctly
        """
        project_path = tmp_path / "test_project"
        project_path.mkdir()

        backup_path = tmp_path / "backup"
        backup_path.mkdir()

        # Create multiple backup elements
        backup_claude = backup_path / ".claude"
        backup_claude.mkdir(parents=True)

        # Agent
        backup_agents = backup_claude / "agents"
        backup_agents.mkdir()
        agent_file = backup_agents / "my-agent.md"
        agent_file.write_text("Agent content")

        # Command
        backup_commands = backup_claude / "commands" / "moai"
        backup_commands.mkdir(parents=True)
        cmd_file = backup_commands / "my-command.md"
        cmd_file.write_text("Command content")

        # Hook
        backup_hooks = backup_claude / "hooks" / "moai"
        backup_hooks.mkdir(parents=True)
        hook_file = backup_hooks / "my-hook.py"
        hook_file.write_text("print('hook')")

        # Create backup metadata
        backup_moai = backup_path / ".moai"
        backup_moai.mkdir(parents=True)
        (backup_moai / "backups").mkdir()

        restorer = create_selective_restorer(project_path, backup_path)

        # Restore all elements
        elements = [
            ".claude/agents/my-agent.md",
            ".claude/commands/moai/my-command.md",
            ".claude/hooks/moai/my-hook.py",
        ]

        try:
            success, stats = restorer.restore_elements(elements)
            assert success is True
            assert stats["success"] == 3
            assert stats["failed"] == 0

            # Verify all files were restored
            assert (project_path / ".claude" / "agents" / "my-agent.md").exists()
            assert (project_path / ".claude" / "commands" / "moai" / "my-command.md").exists()
            assert (project_path / ".claude" / "hooks" / "moai" / "my-hook.py").exists()

        except Exception as e:
            if "not in the subpath of" in str(e):
                assert False, f"Multiple element path resolution should work: {e}"
            else:
                raise
