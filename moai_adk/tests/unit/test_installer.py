"""
Unit tests for the installer module.

Tests the ProjectInstaller class and its installation orchestration methods
to ensure proper MoAI-ADK project setup and integration.
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock, call

from moai_adk.install.installer import ProjectInstaller
from moai_adk.config import Config, RuntimeConfig

from moai_adk.install.installation_result import InstallationResult
class TestProjectInstaller:
    """Test cases for ProjectInstaller class."""

    @pytest.fixture
    def temp_dir(self):
        """Create a temporary directory for testing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    @pytest.fixture
    def sample_config(self, temp_dir):
        """Create a sample Config instance for testing."""
        return Config(
            name="test-project",
            
            template="standard",
            runtime=RuntimeConfig("python"),
            path=temp_dir / "test_project"
        )

    @pytest.fixture
    def installer(self, sample_config):
        """Create a ProjectInstaller instance for testing."""
        with patch.object(ProjectInstaller, '_setup_template_directory'):
            with patch.object(ProjectInstaller, '_initialize_managers'):
                installer = ProjectInstaller(sample_config)
                # Manually set up mocked managers for testing
                installer.security_manager = MagicMock()
                installer.file_manager = MagicMock()
                installer.directory_manager = MagicMock()
                installer.config_manager = MagicMock()
                installer.git_manager = MagicMock()
                installer.system_manager = MagicMock()
                installer.progress = MagicMock()
                installer.template_dir = sample_config.project_path.parent / "templates"
                return installer

    def test_init_with_config(self, sample_config):
        """Test ProjectInstaller initialization with config."""
        with patch.object(ProjectInstaller, '_setup_template_directory'):
            with patch.object(ProjectInstaller, '_initialize_managers'):
                installer = ProjectInstaller(sample_config)
                assert installer.config == sample_config

    def test_setup_template_directory_with_importlib(self):
        """Test template directory setup using importlib.resources."""
        with patch('moai_adk.installer.files') as mock_files:
            mock_files.return_value.__truediv__.return_value = Path("/mocked/templates")

            config = Config("test", template="standard", runtime=RuntimeConfig("python"))
            with patch.object(ProjectInstaller, '_initialize_managers'):
                installer = ProjectInstaller(config)

            assert installer.template_dir == Path("/mocked/templates")

    def test_setup_template_directory_fallback(self, sample_config):
        """Test template directory setup with fallback method."""
        with patch('moai_adk.installer.files', side_effect=Exception("Import failed")):
            with patch.object(ProjectInstaller, '_initialize_managers'):
                installer = ProjectInstaller(sample_config)

            # Should fallback to relative path
            expected_path = Path(__file__).parent.parent.parent / "moai_adk" / "templates"
            assert installer.template_dir.name == "templates"

    def test_initialize_managers(self, sample_config):
        """Test initialization of all manager components."""
        with patch.object(ProjectInstaller, '_setup_template_directory'):
            installer = ProjectInstaller(sample_config)

        # Check that all managers are initialized
        assert hasattr(installer, 'security_manager')
        assert hasattr(installer, 'file_manager')
        assert hasattr(installer, 'directory_manager')
        assert hasattr(installer, 'config_manager')
        assert hasattr(installer, 'git_manager')
        assert hasattr(installer, 'system_manager')
        assert hasattr(installer, 'progress')

    def test_install_success_complete_workflow(self, installer):
        """Test successful complete installation workflow."""
        # Mock all manager methods to return successful results
        installer.directory_manager.create_project_directory.return_value = None
        installer.directory_manager.create_directory_structure.return_value = [Path("/test/dir")]

        # Mock all installation methods
        installer._install_agents = MagicMock(return_value=[Path("/agent1.md")])
        installer._install_commands = MagicMock(return_value=[Path("/cmd1.md")])
        installer._install_hook_scripts = MagicMock(return_value=[Path("/hook1.py")])
        installer._install_templates = MagicMock(return_value=[Path("/template1.md")])
        installer._install_memory_files = MagicMock(return_value=[Path("/memory1.md")])
        installer._install_github_files = MagicMock(return_value=[Path("/workflow1.yml")])
        installer._install_verification_scripts = MagicMock(return_value=[Path("/script1.py")])
        installer._create_steering_templates = MagicMock(return_value=[Path("/steering1.md")])
        installer._create_claude_md = MagicMock(return_value=Path("/CLAUDE.md"))
        installer._install_output_styles = MagicMock(return_value=[Path("/style1.md")])
        installer._generate_next_steps = MagicMock(return_value=["step1", "step2"])

        # Mock config manager methods
        installer.config_manager.create_initial_indexes.return_value = [Path("/index1.json")]
        installer.config_manager.create_claude_settings_file.return_value = Path("/settings.json")
        installer.config_manager.create_moai_config_file.return_value = Path("/moai_config.json")
        installer.config_manager.create_package_json.return_value = Path("/package.json")

        # Mock git manager
        installer.git_manager.initialize_git_repository.return_value = (True, True)

        # Mock system manager
        installer.system_manager.should_create_package_json.return_value = True
        installer.system_manager.check_nodejs_and_npm.return_value = True

        # Mock progress tracker
        installer.progress.update_progress = MagicMock()

        result = installer.install()

        # Check result
        assert isinstance(result, InstallationResult)
        assert result.success is True
        assert len(result.files_created) > 0

        # Verify all installation methods were called
        installer._install_agents.assert_called_once()
        installer._install_commands.assert_called_once()
        installer._install_hook_scripts.assert_called_once()
        installer._install_templates.assert_called_once()
        installer._install_memory_files.assert_called_once()
        installer._install_github_files.assert_called_once()
        installer._install_verification_scripts.assert_called_once()

    def test_install_with_progress_callback(self, installer):
        """Test installation with progress callback."""
        # Setup minimal mocks for successful installation
        installer.directory_manager.create_project_directory.return_value = None
        installer.directory_manager.create_directory_structure.return_value = []

        # Mock all installation methods to return empty lists
        for method_name in [
            '_install_agents', '_install_commands', '_install_hook_scripts',
            '_install_templates', '_install_memory_files', '_install_github_files',
            '_install_verification_scripts', '_create_steering_templates',
            '_install_output_styles'
        ]:
            setattr(installer, method_name, MagicMock(return_value=[]))

        installer._create_claude_md = MagicMock(return_value=Path("/CLAUDE.md"))
        installer._generate_next_steps = MagicMock(return_value=[])

        installer.config_manager.create_initial_indexes.return_value = []
        installer.config_manager.create_claude_settings_file.return_value = Path("/settings.json")
        installer.config_manager.create_moai_config_file.return_value = Path("/config.json")

        installer.git_manager.initialize_git_repository.return_value = (True, False)
        installer.system_manager.should_create_package_json.return_value = False
        installer.system_manager.check_nodejs_and_npm.return_value = True

        # Mock progress callback
        progress_callback = MagicMock()

        result = installer.install(progress_callback)

        # Verify progress callback was used
        installer.progress.update_progress.assert_called()
        # Check that progress callback was passed to update_progress calls
        for call_args in installer.progress.update_progress.call_args_list:
            assert len(call_args[0]) >= 1  # At least step description
            assert call_args[0][1] == progress_callback or call_args[1].get('callback') == progress_callback

    def test_install_exception_handling(self, installer):
        """Test installation exception handling."""
        # Make directory creation fail
        installer.directory_manager.create_project_directory.side_effect = Exception("Directory creation failed")

        result = installer.install()

        # Check that failure result is returned
        assert isinstance(result, InstallationResult)
        assert result.success is False
        assert "Directory creation failed" in result.errors

    def test_install_git_scenarios(self, installer):
        """Test different Git initialization scenarios."""
        # Setup basic mocks
        installer.directory_manager.create_project_directory.return_value = None
        installer.directory_manager.create_directory_structure.return_value = []

        # Mock all installation methods
        for method_name in [
            '_install_agents', '_install_commands', '_install_hook_scripts',
            '_install_templates', '_install_memory_files', '_install_github_files',
            '_install_verification_scripts', '_create_steering_templates',
            '_install_output_styles'
        ]:
            setattr(installer, method_name, MagicMock(return_value=[]))

        installer._create_claude_md = MagicMock(return_value=Path("/CLAUDE.md"))
        installer._generate_next_steps = MagicMock(return_value=[])

        installer.config_manager.create_initial_indexes.return_value = []
        installer.config_manager.create_claude_settings_file.return_value = Path("/settings.json")
        installer.config_manager.create_moai_config_file.return_value = Path("/config.json")

        installer.system_manager.should_create_package_json.return_value = False
        installer.system_manager.check_nodejs_and_npm.return_value = True

        # Test scenario 1: Git initialized successfully
        installer.git_manager.initialize_git_repository.return_value = (True, True)
        installer.config.project_path = Path("/test")

        with patch.object(Path, 'exists', return_value=True):
            result = installer.install()
            assert result.success is True

        # Test scenario 2: Git already exists
        installer.git_manager.initialize_git_repository.return_value = (True, False)
        result = installer.install()
        assert result.success is True

        # Test scenario 3: Git initialization failed
        installer.git_manager.initialize_git_repository.return_value = (False, False)
        result = installer.install()
        assert result.success is True  # Should still succeed

    def test_install_package_json_creation(self, installer):
        """Test package.json creation scenarios."""
        # Setup basic mocks
        installer.directory_manager.create_project_directory.return_value = None
        installer.directory_manager.create_directory_structure.return_value = []

        # Mock all installation methods
        for method_name in [
            '_install_agents', '_install_commands', '_install_hook_scripts',
            '_install_templates', '_install_memory_files', '_install_github_files',
            '_install_verification_scripts', '_create_steering_templates',
            '_install_output_styles'
        ]:
            setattr(installer, method_name, MagicMock(return_value=[]))

        installer._create_claude_md = MagicMock(return_value=Path("/CLAUDE.md"))
        installer._generate_next_steps = MagicMock(return_value=[])

        installer.config_manager.create_initial_indexes.return_value = []
        installer.config_manager.create_claude_settings_file.return_value = Path("/settings.json")
        installer.config_manager.create_moai_config_file.return_value = Path("/config.json")
        installer.config_manager.create_package_json.return_value = Path("/package.json")

        installer.git_manager.initialize_git_repository.return_value = (True, False)
        installer.system_manager.check_nodejs_and_npm.return_value = True

        # Test scenario 1: Should create package.json
        installer.system_manager.should_create_package_json.return_value = True
        result = installer.install()
        installer.config_manager.create_package_json.assert_called_once()

        # Reset mock and test scenario 2: Should not create package.json
        installer.config_manager.create_package_json.reset_mock()
        installer.system_manager.should_create_package_json.return_value = False
        result = installer.install()
        installer.config_manager.create_package_json.assert_not_called()

    def test_install_agents(self, installer, temp_dir):
        """Test agent installation method."""
        # Setup template directory structure
        agents_source = temp_dir / "templates" / ".claude" / "agents" / "moai"
        agents_source.mkdir(parents=True)
        (agents_source / "agent1.md").write_text("agent content")

        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        # Mock file manager
        installer.file_manager.copy_template_files.return_value = [Path("/agent1.md")]

        result = installer._install_agents()

        assert len(result) == 1
        installer.file_manager.copy_template_files.assert_called_once()

    def test_install_agents_missing_source(self, installer, temp_dir):
        """Test agent installation when source directory is missing."""
        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        result = installer._install_agents()

        # Should return empty list when source doesn't exist
        assert result == []

    def test_install_commands(self, installer, temp_dir):
        """Test command installation method."""
        # Setup template directory structure
        commands_source = temp_dir / "templates" / ".claude" / "commands" / "moai"
        commands_source.mkdir(parents=True)
        (commands_source / "command1.md").write_text("command content")

        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        # Mock file manager
        installer.file_manager.copy_template_files.return_value = [Path("/command1.md")]

        result = installer._install_commands()

        assert len(result) == 1
        installer.file_manager.copy_template_files.assert_called_once()

    def test_install_hook_scripts(self, installer):
        """Test hook scripts installation method."""
        installer.config.project_path = Path("/test/project")
        installer.file_manager.copy_hook_scripts.return_value = [Path("/hook1.py")]

        result = installer._install_hook_scripts()

        assert len(result) == 1
        expected_target = Path("/test/project") / ".claude" / "hooks" / "moai"
        installer.file_manager.copy_hook_scripts.assert_called_once_with(expected_target)

    def test_install_templates(self, installer):
        """Test template installation method."""
        installer.template_dir = Path("/templates")
        installer.config.project_path = Path("/project")
        installer.file_manager.copy_template_files.return_value = [Path("/template1.md")]

        result = installer._install_templates()

        assert len(result) == 1
        expected_source = Path("/templates") / ".moai" / "templates"
        expected_target = Path("/project") / ".moai" / "templates"
        installer.file_manager.copy_template_files.assert_called_once_with(
            expected_source, expected_target, "*.md"
        )

    def test_install_memory_files(self, installer, temp_dir):
        """Test memory files installation method."""
        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        # Create template structure
        claude_memory = installer.template_dir / ".claude" / "memory"
        claude_memory.mkdir(parents=True)

        moai_memory = installer.template_dir / ".moai" / "memory"
        moai_memory.mkdir(parents=True)

        decisions_dir = moai_memory / "decisions"
        decisions_dir.mkdir()

        # Mock file manager calls
        installer.file_manager.copy_template_files.side_effect = [
            [Path("/claude_memory1.md")],  # claude memory
            [Path("/moai_memory1.md")],    # moai memory
            [Path("/decision1.md")]        # decisions
        ]

        result = installer._install_memory_files()

        assert len(result) == 3
        assert installer.file_manager.copy_template_files.call_count == 3

    def test_install_memory_files_missing_decisions(self, installer, temp_dir):
        """Test memory files installation when decisions directory is missing."""
        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        # Create template structure without decisions
        claude_memory = installer.template_dir / ".claude" / "memory"
        claude_memory.mkdir(parents=True)

        moai_memory = installer.template_dir / ".moai" / "memory"
        moai_memory.mkdir(parents=True)

        # Mock file manager calls
        installer.file_manager.copy_template_files.side_effect = [
            [Path("/claude_memory1.md")],  # claude memory
            [Path("/moai_memory1.md")]     # moai memory
        ]

        result = installer._install_memory_files()

        assert len(result) == 2
        assert installer.file_manager.copy_template_files.call_count == 2

    def test_install_github_files(self, installer, temp_dir):
        """Test GitHub files installation method."""
        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        # Create template structure
        workflows_dir = installer.template_dir / ".github" / "workflows"
        workflows_dir.mkdir(parents=True)

        pr_template = installer.template_dir / ".github" / "PULL_REQUEST_TEMPLATE.md"
        pr_template.parent.mkdir(exist_ok=True)
        pr_template.write_text("PR template")

        # Mock file manager and config
        installer.file_manager.copy_template_files.return_value = [Path("/workflow1.yml")]
        installer.file_manager.copy_and_render_template.return_value = True
        installer.config.get_template_context = MagicMock(return_value={})

        result = installer._install_github_files()

        assert len(result) == 2  # workflow + PR template
        installer.file_manager.copy_template_files.assert_called_once()
        installer.file_manager.copy_and_render_template.assert_called_once()

    def test_install_verification_scripts(self, installer):
        """Test verification scripts installation method."""
        installer.config.project_path = Path("/project")
        installer.file_manager.copy_verification_scripts.return_value = [Path("/script1.py")]

        result = installer._install_verification_scripts()

        assert len(result) == 1
        expected_target = Path("/project") / ".moai" / "scripts"
        installer.file_manager.copy_verification_scripts.assert_called_once_with(expected_target)

    def test_create_steering_templates(self, installer, temp_dir):
        """Test steering templates creation method."""
        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        # Create template files
        steering_dir = installer.template_dir / "steering"
        steering_dir.mkdir(parents=True)

        for template_name in ["product", "structure", "tech"]:
            (steering_dir / f"{template_name}.template.md").write_text("template content")

        # Mock file manager and config
        installer.file_manager.copy_and_render_template.return_value = True
        installer.config.get_template_context = MagicMock(return_value={})

        result = installer._create_steering_templates()

        assert len(result) == 3
        assert installer.file_manager.copy_and_render_template.call_count == 3

    def test_create_steering_templates_missing_templates(self, installer, temp_dir):
        """Test steering templates creation when templates are missing."""
        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        # Don't create template files

        result = installer._create_steering_templates()

        assert result == []

    def test_create_claude_md(self, installer, temp_dir):
        """Test CLAUDE.md creation method."""
        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        # Create template file
        claude_md_template = installer.template_dir / "CLAUDE.md"
        claude_md_template.write_text("CLAUDE.md content")

        # Create project directory
        installer.config.project_path.mkdir(parents=True)

        result = installer._create_claude_md()

        expected_path = installer.config.project_path / "CLAUDE.md"
        assert result == expected_path
        assert expected_path.exists()

    def test_create_claude_md_missing_template(self, installer, temp_dir):
        """Test CLAUDE.md creation when template is missing."""
        installer.template_dir = temp_dir / "templates"
        installer.config.project_path = temp_dir / "project"

        # Create project directory but not template
        installer.config.project_path.mkdir(parents=True)

        result = installer._create_claude_md()

        expected_path = installer.config.project_path / "CLAUDE.md"
        assert result == expected_path

    def test_install_output_styles(self, installer):
        """Test output styles installation method."""
        installer.config.project_path = Path("/project")
        installer.file_manager.install_output_styles.return_value = [Path("/style1.md")]
        installer.config.get_template_context = MagicMock(return_value={})

        result = installer._install_output_styles()

        assert len(result) == 1
        expected_target = Path("/project") / ".claude" / "output-styles"
        installer.file_manager.install_output_styles.assert_called_once_with(
            expected_target, installer.config.get_template_context()
        )

    def test_generate_next_steps(self, installer):
        """Test next steps generation method."""
        installer.config.name = "test-project"

        result = installer._generate_next_steps()

        assert isinstance(result, list)
        assert len(result) > 0
        assert any("test-project" in step for step in result)
        assert any("claude" in step for step in result)

    def test_integration_all_methods_called(self, installer):
        """Test that all installation methods are called during install."""
        # Mock all dependencies for successful installation
        installer.directory_manager.create_project_directory.return_value = None
        installer.directory_manager.create_directory_structure.return_value = []

        # Mock all private installation methods
        installation_methods = [
            '_install_agents', '_install_commands', '_install_hook_scripts',
            '_install_templates', '_install_memory_files', '_install_github_files',
            '_install_verification_scripts', '_create_steering_templates',
            '_create_claude_md', '_install_output_styles', '_generate_next_steps'
        ]

        for method_name in installation_methods:
            if method_name == '_create_claude_md':
                setattr(installer, method_name, MagicMock(return_value=Path("/test.md")))
            else:
                setattr(installer, method_name, MagicMock(return_value=[]))

        # Mock config manager
        installer.config_manager.create_initial_indexes.return_value = []
        installer.config_manager.create_claude_settings_file.return_value = Path("/settings.json")
        installer.config_manager.create_moai_config_file.return_value = Path("/config.json")

        # Mock other managers
        installer.git_manager.initialize_git_repository.return_value = (True, False)
        installer.system_manager.should_create_package_json.return_value = False
        installer.system_manager.check_nodejs_and_npm.return_value = True

        result = installer.install()

        # Verify all installation methods were called
        for method_name in installation_methods:
            method = getattr(installer, method_name)
            method.assert_called_once()

        assert result.success is True