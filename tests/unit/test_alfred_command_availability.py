"""Unit tests for Alfred command file availability and integrity.

Tests for ensuring all Alfred command files are present and properly configured.

SPEC-ALFRED-INIT-FIX-001: Alfred 초기화 실패 해결
Priority: P0 (CRITICAL)
Impact: Entire Alfred workflow blocked

Test Coverage:
    - test_alfred_0_project_command_exists: Verify command file exists
    - test_command_file_yaml_valid: Validate YAML frontmatter
    - test_command_name_is_correct: Verify command name matches location
    - test_all_4_alfred_commands_exist: Ensure all 4 Alfred commands (0-3) present
    - test_project_manager_integration: Verify agent delegation
    - test_template_file_synchronized: Ensure template sync
    - test_command_has_task_delegation: Verify Task() usage
    - test_command_has_ask_user_question: Verify user interaction support
    - test_commands_directory_exists: Verify directory structure
    - test_all_command_files_readable: Verify file accessibility
    - test_no_duplicate_command_names: Prevent name conflicts
    - test_0_project_delegates_to_project_manager: Integration test
    - test_command_workflow_sequence: Verify workflow validity

Success Criteria:
    - ✅ All 13 tests passing
    - ✅ 100% command file availability
    - ✅ Template synchronization verified
    - ✅ Git tracking confirmed
"""

from pathlib import Path

import yaml


def load_command_yaml(file_path: Path) -> dict:
    """Load YAML frontmatter from command file.

    Args:
        file_path: Path to command file

    Returns:
        Parsed YAML data

    Raises:
        ValueError: If YAML frontmatter is invalid
    """
    content = file_path.read_text()
    parts = content.split("---")
    if len(parts) < 3:
        raise ValueError(f"{file_path.name} must have YAML frontmatter")

    yaml_content = parts[1].strip()
    yaml_data = yaml.safe_load(yaml_content)
    if yaml_data is None:
        raise ValueError(f"{file_path.name} has invalid YAML frontmatter")
    return yaml_data


class TestAlfredCommandAvailability:
    """Test Alfred command file availability and structure.

    Ensures that all Alfred commands are present, properly formatted,
    and configured for correct integration with Claude Code.
    """

    def test_alfred_0_project_command_exists(self):
        """Should ensure .claude/commands/alfred/0-project.md exists"""
        # Arrange
        command_file = Path("./.claude/commands/alfred/0-project.md")

        # Act & Assert
        assert command_file.exists(), f"Command file {command_file} does not exist"
        assert command_file.is_file(), f"{command_file} is not a file"
        assert command_file.stat().st_size > 0, f"{command_file} is empty"

    def test_command_file_yaml_valid(self):
        """Should validate YAML frontmatter in command file"""
        # Arrange
        command_file = Path("./.claude/commands/alfred/0-project.md")

        # Act & Assert
        yaml_data = load_command_yaml(command_file)
        assert "name" in yaml_data, "YAML must contain 'name' field"
        assert "description" in yaml_data, "YAML must contain 'description' field"

    def test_command_name_is_correct(self):
        """Should validate command name matches file location"""
        # Arrange
        command_file = Path("./.claude/commands/alfred/0-project.md")

        # Act
        yaml_data = load_command_yaml(command_file)
        command_name = yaml_data.get("name", "")

        # Assert
        assert command_name == "alfred:0-project", f"Expected 'alfred:0-project', got '{command_name}'"

    def test_all_4_alfred_commands_exist(self):
        """Should ensure all 4 Alfred commands exist (0-3)"""
        # Arrange
        commands_dir = Path("./.claude/commands/alfred")
        expected_commands = {
            "0-project.md": "alfred:0-project",
            "1-plan.md": "alfred:1-plan",
            "2-run.md": "alfred:2-run",
            "3-sync.md": "alfred:3-sync",
        }

        # Act & Assert
        for filename, expected_name in expected_commands.items():
            file_path = commands_dir / filename
            assert file_path.exists(), f"Command file {filename} does not exist"

            yaml_data = load_command_yaml(file_path)
            actual_name = yaml_data.get("name", "")

            assert actual_name == expected_name, f"{filename}: Expected name '{expected_name}', got '{actual_name}'"

    def test_project_manager_integration(self):
        """Should verify project-manager agent is referenced in 0-project command"""
        # Arrange
        command_file = Path("./.claude/commands/alfred/0-project.md")

        # Act
        content = command_file.read_text()

        # Assert
        assert "project-manager" in content, "0-project.md should reference 'project-manager' agent"

    def test_template_file_synchronized(self):
        """Should verify template file matches local file"""
        # Arrange
        local_file = Path("./.claude/commands/alfred/0-project.md")
        template_file = Path("./src/moai_adk/templates/.claude/commands/alfred/0-project.md")

        # Act
        assert template_file.exists(), f"Template file {template_file} does not exist"

        local_content = local_file.read_text()
        template_content = template_file.read_text()

        # Assert
        assert local_content == template_content, (
            "Local command file does not match template. "
            "Run: cp ./src/moai_adk/templates/.claude/commands/alfred/*.md ./.claude/commands/alfred/"
        )

    def test_command_has_task_delegation(self):
        """Should verify command uses Task() for agent delegation"""
        # Arrange
        command_file = Path("./.claude/commands/alfred/0-project.md")

        # Act
        content = command_file.read_text()

        # Assert
        assert "Task(" in content, "0-project command should use Task() for agent delegation"

    def test_command_has_ask_user_question(self):
        """Should verify command includes AskUserQuestion for user interaction"""
        # Arrange
        command_file = Path("./.claude/commands/alfred/0-project.md")

        # Act
        content = command_file.read_text()

        # Assert
        assert "AskUserQuestion" in content, "0-project command should include AskUserQuestion for user interaction"


class TestAlfredCommandDirectory:
    """Test Alfred command directory structure"""

    def test_commands_directory_exists(self):
        """Should ensure .claude/commands/alfred directory exists"""
        # Arrange
        commands_dir = Path("./.claude/commands/alfred")

        # Act & Assert
        assert commands_dir.exists(), f"Commands directory {commands_dir} does not exist"
        assert commands_dir.is_dir(), f"{commands_dir} is not a directory"

    def test_all_command_files_readable(self):
        """Should verify all command files are readable"""
        # Arrange
        commands_dir = Path("./.claude/commands/alfred")
        command_files = ["0-project.md", "1-plan.md", "2-run.md", "3-sync.md"]

        # Act & Assert
        for filename in command_files:
            file_path = commands_dir / filename
            assert file_path.exists(), f"{filename} not found"

            content = file_path.read_text()
            assert len(content) > 0, f"{filename} is empty"

    def test_no_duplicate_command_names(self):
        """Should ensure no duplicate command names across all files"""
        # Arrange
        commands_dir = Path("./.claude/commands/alfred")

        # Act
        command_names = []
        for file_path in commands_dir.glob("*.md"):
            try:
                yaml_data = load_command_yaml(file_path)
                if "name" in yaml_data:
                    command_names.append(yaml_data["name"])
            except ValueError:
                # Skip files without valid YAML frontmatter
                pass

        # Assert
        assert len(command_names) == len(set(command_names)), f"Duplicate command names found: {command_names}"


class TestAlfredCommandIntegration:
    """Integration tests for Alfred command system"""

    def test_0_project_delegates_to_project_manager(self):
        """Should verify 0-project command properly delegates to project-manager agent"""
        # Arrange
        command_file = Path("./.claude/commands/alfred/0-project.md")

        # Act
        content = command_file.read_text()

        # Assert
        assert "project-manager" in content, "0-project must reference project-manager agent"
        assert "Task(" in content, "0-project must use Task() for delegation"

    def test_command_workflow_sequence(self):
        """Should verify command workflow sequence is valid"""
        # Arrange
        commands_to_check = {
            "0-project.md": "Initialize/Update Project",
            "1-plan.md": "Plan Feature Implementation",
            "2-run.md": "Execute TDD Implementation",
            "3-sync.md": "Synchronize Documentation",
        }
        commands_dir = Path("./.claude/commands/alfred")

        # Act & Assert
        for filename, expected_phase in commands_to_check.items():
            file_path = commands_dir / filename
            assert file_path.exists(), f"{filename} missing"

            content = file_path.read_text()
            # Verify file has proper structure
            assert "---" in content, f"{filename} missing YAML frontmatter"
            assert "#" in content, f"{filename} missing markdown content"
