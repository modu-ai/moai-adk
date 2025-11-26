# # REMOVED_ORPHAN_TEST:INITIALIZER-001 | SPEC: SPEC-TEST-COVERAGE-001.md
"""Integration tests for ProjectInitializer

Tests the complete project initialization workflow:
- 5-phase initialization process
- Language detection
- Reinitialization handling
- Error handling
"""

from pathlib import Path
from unittest.mock import Mock, patch

from moai_adk.core.project.initializer import (
    InstallationResult,
    ProjectInitializer,
    initialize_project,
)


class TestInstallationResult:
    """Test InstallationResult class"""

    def test_installation_result_success(self) -> None:
        """Should create successful result"""
        result = InstallationResult(
            success=True,
            project_path="/path/to/project",
            language="python",
            mode="personal",
            locale="ko",
            duration=1500,
            created_files=[".moai/", ".claude/"],
        )

        assert result.success is True, f"Initialization failed with errors: {result.errors}"
        assert result.project_path == "/path/to/project"
        assert result.language == "python"
        assert result.mode == "personal"
        assert result.locale == "ko"
        assert result.duration == 1500
        assert len(result.created_files) == 2
        assert result.errors == []

    def test_installation_result_failure(self) -> None:
        """Should create failure result with errors"""
        result = InstallationResult(
            success=False,
            project_path="/path/to/project",
            language="unknown",
            mode="personal",
            locale="ko",
            duration=500,
            created_files=[],
            errors=["Validation failed", "Missing dependency"],
        )

        assert result.success is False
        assert len(result.errors) == 2
        assert "Validation failed" in result.errors


class TestProjectInitializerInit:
    """Test ProjectInitializer initialization"""

    def test_init_resolves_path(self, tmp_path: Path) -> None:
        """Should resolve project path to absolute"""
        initializer = ProjectInitializer(tmp_path)
        assert initializer.path.is_absolute()
        assert initializer.path == tmp_path.resolve()

    def test_init_creates_components(self, tmp_path: Path) -> None:
        """Should create validator and executor"""
        initializer = ProjectInitializer(tmp_path)
        assert initializer.validator is not None
        assert initializer.executor is not None


class TestIsInitialized:
    """Test initialization status check"""

    def test_is_initialized_returns_false_for_new_project(self, tmp_path: Path) -> None:
        """Should return False when .moai doesn't exist"""
        initializer = ProjectInitializer(tmp_path)
        assert initializer.is_initialized() is False

    def test_is_initialized_returns_true_when_moai_exists(self, tmp_path: Path) -> None:
        """Should return True when .moai exists"""
        (tmp_path / ".moai").mkdir()
        initializer = ProjectInitializer(tmp_path)
        assert initializer.is_initialized() is True


class TestInitialize:
    """Test project initialization"""

    def test_initialize_prevents_duplicate_init(self, tmp_path: Path) -> None:
        """Should return failure when already initialized"""
        (tmp_path / ".moai").mkdir()

        initializer = ProjectInitializer(tmp_path)

        # Should return failure when reinit=False and project exists
        result = initializer.initialize(reinit=False, backup_enabled=False)

        assert result.success is False
        assert len(result.errors) > 0
        assert "already initialized" in result.errors[0].lower()

    def test_initialize_allows_reinit_when_flagged(self, tmp_path: Path) -> None:
        """Should allow reinitialization when reinit=True"""
        (tmp_path / ".moai").mkdir()
        (tmp_path / ".moai" / "config.json").write_text("{}")

        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(reinit=True, backup_enabled=False)

        # Should succeed
        assert result.success is True, f"Initialization failed with errors: {result.errors}"

    def test_initialize_detects_language(self, tmp_path: Path) -> None:
        """Should accept explicit language parameter"""
        # Language detection moved to project-manager in /moai:0-project
        # ProjectInitializer accepts explicit language parameter
        (tmp_path / "requirements.txt").write_text("pytest==8.0.0")

        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(language="python")

        assert result.success is True, f"Initialization failed with errors: {result.errors}"
        assert result.language == "python"

    def test_initialize_uses_specified_language(self, tmp_path: Path) -> None:
        """Should use specified language instead of detection"""
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(language="typescript")

        assert result.success is True, f"Initialization failed with errors: {result.errors}"
        assert result.language == "typescript"

    def test_initialize_executes_all_phases(self, tmp_path: Path) -> None:
        """Should execute all 5 phases"""
        initializer = ProjectInitializer(tmp_path)

        with (
            patch.object(initializer.executor, "execute_preparation_phase") as mock_prep,
            patch.object(initializer.executor, "execute_directory_phase") as mock_dir,
            patch.object(initializer.executor, "execute_resource_phase") as mock_res,
            patch.object(initializer.executor, "execute_configuration_phase") as mock_conf,
            patch.object(initializer.executor, "execute_validation_phase") as mock_val,
        ):
            # Setup mocks
            mock_res.return_value = [".moai/", ".claude/"]
            mock_conf.return_value = [".moai/config.json"]

            result = initializer.initialize()

            # All phases should be called
            mock_prep.assert_called_once()
            mock_dir.assert_called_once()
            mock_res.assert_called_once()
            mock_conf.assert_called_once()
            mock_val.assert_called_once()

            assert result.success is True, f"Initialization failed with errors: {result.errors}"

    def test_initialize_passes_correct_config(self, tmp_path: Path) -> None:
        """Should pass correct configuration to phase 4"""
        initializer = ProjectInitializer(tmp_path)

        with patch.object(initializer.executor, "execute_configuration_phase") as mock_conf:
            mock_conf.return_value = [".moai/config.json"]

            initializer.initialize(mode="team", locale="en", language="go", backup_enabled=False)

            # Check config passed to phase 4
            call_args = mock_conf.call_args
            config = call_args[0][1]

            # Config should have project section
            assert "project" in config
            assert config["project"]["name"] == tmp_path.name
            assert config["project"]["mode"] == "team"
            assert config["project"]["locale"] == "en"
            assert config["project"]["language"] == "go"

    def test_initialize_calls_progress_callback(self, tmp_path: Path) -> None:
        """Should call progress callback during initialization"""
        callback = Mock()
        initializer = ProjectInitializer(tmp_path)

        initializer.initialize(progress_callback=callback, backup_enabled=False)

        # Callback should be called at least once (for each phase)
        assert callback.call_count >= 5

    def test_initialize_creates_complete_structure(self, tmp_path: Path) -> None:
        """Should create complete project structure"""
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize()

        assert result.success is True, f"Initialization failed with errors: {result.errors}"

        # Required directories should exist
        assert (tmp_path / ".moai").exists()
        assert (tmp_path / ".claude").exists()
        assert (tmp_path / ".github").exists()
        assert (tmp_path / "CLAUDE.md").exists()
        assert (tmp_path / ".moai" / "config" / "config.json").exists()

    def test_initialize_creates_english_claude_template(self, tmp_path: Path) -> None:
        """Should copy English CLAUDE.md template by default with variable substitution"""
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize()

        assert result.success is True
        claude_content = (tmp_path / "CLAUDE.md").read_text(encoding="utf-8")
        # Check that template variables are substituted (not literal {{VARIABLE}} placeholders)
        assert "You are the SuperAgent" in claude_content  # Updated to match current template
        assert "SuperAgent" in claude_content
        assert "{{PROJECT_NAME}}" not in claude_content  # Ensure variables are substituted
        assert "페르소나" not in claude_content

    def test_initialize_measures_duration(self, tmp_path: Path) -> None:
        """Should measure initialization duration"""
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize()

        assert result.success is True, f"Initialization failed with errors: {result.errors}"
        assert result.duration > 0
        assert isinstance(result.duration, int)

    def test_initialize_returns_created_files(self, tmp_path: Path) -> None:
        """Should return list of created files"""
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize()

        assert result.success is True, f"Initialization failed with errors: {result.errors}"
        assert len(result.created_files) > 0
        assert any(".moai" in f for f in result.created_files)
        assert any(".github" in f for f in result.created_files)

    def test_initialize_handles_errors_gracefully(self, tmp_path: Path) -> None:
        """Should handle errors and return failure result"""
        initializer = ProjectInitializer(tmp_path)

        with patch.object(initializer.executor, "execute_preparation_phase") as mock_prep:
            # Simulate error
            mock_prep.side_effect = Exception("Test error")

            result = initializer.initialize()

            assert result.success is False
            assert len(result.errors) > 0
            assert "Test error" in result.errors[0]

    def test_initialize_with_team_mode(self, tmp_path: Path) -> None:
        """Should initialize with team mode"""
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(mode="team")

        assert result.success is True, f"Initialization failed with errors: {result.errors}"
        assert result.mode == "team"

    def test_initialize_with_different_locale(self, tmp_path: Path) -> None:
        """Should support different locales"""
        initializer = ProjectInitializer(tmp_path)
        result = initializer.initialize(locale="en")

        assert result.success is True, f"Initialization failed with errors: {result.errors}"
        assert result.locale == "en"


class TestInitializeProjectFunction:
    """Test initialize_project helper function"""

    def test_initialize_project_creates_initializer(self, tmp_path: Path) -> None:
        """Should create ProjectInitializer and call initialize"""
        result = initialize_project(tmp_path)

        assert isinstance(result, InstallationResult)
        assert result.success is True, f"Initialization failed with errors: {result.errors}"

    def test_initialize_project_passes_callback(self, tmp_path: Path) -> None:
        """Should pass progress callback to initializer"""
        callback = Mock()
        result = initialize_project(tmp_path, progress_callback=callback)

        assert result.success is True, f"Initialization failed with errors: {result.errors}"
        # Callback should be called
        assert callback.call_count >= 5
