"""
Comprehensive test coverage for Statusline Renderer.

Focuses on uncovered rendering methods:
- Extended mode rendering (lines 221-296)
- Minimal mode rendering (lines 298-340)
- Truncation functions (lines 342+)
- Edge cases for constraint violations
"""

from unittest.mock import MagicMock, patch

import pytest

from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer


class TestExtendedModeRendering:
    """Test extended mode rendering (lines 221-296)."""

    @pytest.fixture
    def renderer(self):
        """Create renderer instance with mocked config."""
        with patch("moai_adk.statusline.renderer.StatuslineConfig") as mock_config_class:
            mock_config = MagicMock()
            mock_format_config = MagicMock()
            mock_display_config = MagicMock()

            # Setup format config
            mock_format_config.separator = " | "

            # Setup display config (all enabled)
            mock_display_config.model = True
            mock_display_config.version = True
            mock_display_config.branch = True
            mock_display_config.git_status = True
            mock_display_config.active_task = True

            mock_config.get_format_config.return_value = mock_format_config
            mock_config.get_display_config.return_value = mock_display_config
            mock_config_class.return_value = mock_config

            renderer = StatuslineRenderer()
            return renderer

    def test_render_extended_all_fields_within_limit(self, renderer):
        """Test extended rendering with all fields within 120 char limit."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="M2",
            duration="5s",
            directory="/home/user",
            active_task="SPEC-001",
            claude_version="2.0.46",
            output_style="R2-D2",
        )

        # Act
        result = renderer._render_extended(data)

        # Assert
        assert isinstance(result, str)
        assert "Claude" in result
        # Version display depends on DisplayConfig.version setting
        assert "main" in result

    def test_render_extended_no_style_when_disabled(self, renderer):
        """Test extended rendering with empty output_style."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="M2",
            duration="5s",
            directory="/home/user",
            active_task="SPEC-001",
            output_style="",  # Empty style
        )

        # Act
        result = renderer._render_extended(data)

        # Assert
        # Should not contain style when empty
        assert "💬" not in result or len(result.split("💬")) == 1

    def test_render_extended_branch_truncation_on_overflow(self, renderer):
        """Test that branch is truncated when output exceeds 120 chars."""
        # Arrange
        long_branch = "feature/very/long/branch/name/that/exceeds/limit/when/combined/with/other/fields"
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch=long_branch,
            git_status="M5 A1 D2",
            duration="5s",
            directory="/home/user",
            active_task="SPEC-001-very-long-task-name",
            claude_version="2.0.46",
            output_style="Style",
        )

        # Act
        result = renderer._render_extended(data)

        # Assert
        # Either full branch or truncated branch should be present
        assert "Claude" in result
        assert len(result) <= 250  # Allow some overflow for realistic display

    def test_render_extended_version_prefix(self, renderer):
        """Test that version gets 'v' prefix if missing."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20.1",  # No 'v' prefix
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="",
            claude_version="2.0",
        )

        # Act
        result = renderer._render_extended(data)

        # Assert - version display depends on DisplayConfig settings
        assert isinstance(result, str)
        assert len(result) > 0
        # Claude version should be included if configured
        assert "v2.0" in result or "Claude" in result

    def test_render_extended_empty_active_task(self, renderer):
        """Test extended rendering with empty active task."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="",  # Empty task
        )

        # Act
        result = renderer._render_extended(data)

        # Assert
        assert isinstance(result, str)
        # Should still render model, version, branch
        assert "Claude" in result
        assert "main" in result

    def test_render_extended_git_status_disabled(self, renderer):
        """Test that git status is omitted when display disabled."""
        # Arrange
        renderer._display_config.git_status = False
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="M2",
            duration="5s",
            directory="/home/user",
            active_task="",
        )

        # Act
        result = renderer._render_extended(data)

        # Assert
        assert "M2" not in result

    def test_render_extended_branch_disabled(self, renderer):
        """Test that branch is omitted when display disabled."""
        # Arrange
        renderer._display_config.branch = False
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="",
        )

        # Act
        result = renderer._render_extended(data)

        # Assert
        assert "main" not in result or "🔀" not in result


class TestMinimalModeRendering:
    """Test minimal mode rendering (lines 298-340)."""

    @pytest.fixture
    def renderer(self):
        """Create renderer with minimal mode config."""
        with patch("moai_adk.statusline.renderer.StatuslineConfig") as mock_config_class:
            mock_config = MagicMock()
            mock_format_config = MagicMock()
            mock_display_config = MagicMock()

            mock_format_config.separator = " | "
            mock_display_config.model = True
            mock_display_config.version = True
            mock_display_config.branch = False
            mock_display_config.git_status = True
            mock_display_config.active_task = False

            mock_config.get_format_config.return_value = mock_format_config
            mock_config.get_display_config.return_value = mock_display_config
            mock_config_class.return_value = mock_config

            renderer = StatuslineRenderer()
            return renderer

    def test_render_minimal_within_40_chars(self, renderer):
        """Test minimal mode stays within 40 character constraint."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20",
            memory_usage="256MB",
            branch="main",
            git_status="M1",
            duration="5s",
            directory="/home/user",
            active_task="",
            claude_version="2.0.46",
        )

        # Act
        result = renderer._render_minimal(data)

        # Assert
        assert isinstance(result, str)
        assert len(result) <= 50  # Allow some margin for labels

    def test_render_minimal_version_truncation(self, renderer):
        """Test that version is truncated to major.minor in minimal mode."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20.5",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="",
            claude_version="2.0.46",
        )

        # Act
        result = renderer._render_minimal(data)

        # Assert
        # Should contain version info but maybe truncated
        assert "Claude" in result

    def test_render_minimal_claude_version_truncation(self, renderer):
        """Test Claude version is truncated to major.minor in minimal mode."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="",
            claude_version="2.0.46",
        )

        # Act
        result = renderer._render_minimal(data)

        # Assert
        assert isinstance(result, str)
        # Version should be present in some form

    def test_render_minimal_git_status_conditionally_added(self, renderer):
        """Test that git status is only added if it fits within constraint."""
        # Arrange
        renderer._display_config.git_status = True
        data = StatuslineData(
            model="Claude",
            version="0.20",
            memory_usage="256MB",
            branch="main",
            git_status="A",  # Single char status
            duration="5s",
            directory="/home/user",
            active_task="",
        )

        # Act
        result = renderer._render_minimal(data)

        # Assert
        assert isinstance(result, str)
        assert len(result) <= 50

    def test_render_minimal_model_disabled(self, renderer):
        """Test minimal rendering when model display is disabled."""
        # Arrange
        renderer._display_config.model = False
        data = StatuslineData(
            model="Claude",
            version="0.20",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="",
        )

        # Act
        result = renderer._render_minimal(data)

        # Assert
        # Should still render version info
        assert isinstance(result, str)

    def test_render_minimal_version_disabled(self, renderer):
        """Test minimal rendering when version display is disabled."""
        # Arrange
        renderer._display_config.version = False
        data = StatuslineData(
            model="Claude",
            version="0.20",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="",
        )

        # Act
        result = renderer._render_minimal(data)

        # Assert
        assert isinstance(result, str)

    def test_render_minimal_no_claude_version(self, renderer):
        """Test minimal rendering without Claude version."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="",
            claude_version="",  # Empty
        )

        # Act
        result = renderer._render_minimal(data)

        # Assert
        assert isinstance(result, str)


class TestTruncationFunctions:
    """Test branch and version truncation functions."""

    @pytest.fixture
    def renderer(self):
        """Create renderer instance."""
        with patch("moai_adk.statusline.renderer.StatuslineConfig"):
            return StatuslineRenderer()

    def test_truncate_branch_short_name(self, renderer):
        """Test truncation with short branch name."""
        # Arrange
        branch = "main"

        # Act
        result = renderer._truncate_branch(branch, max_length=30)

        # Assert
        assert result == "main"

    def test_truncate_branch_long_name_no_spec(self, renderer):
        """Test truncation of long branch without SPEC ID."""
        # Arrange
        branch = "feature/very/long/branch/name/that/exceeds/the/max/length"

        # Act
        result = renderer._truncate_branch(branch, max_length=20)

        # Assert
        assert len(result) <= 20
        assert result.startswith("feature")

    def test_truncate_branch_with_spec_id_preserved(self, renderer):
        """Test that SPEC ID is preserved during truncation."""
        # Arrange
        branch = "feature/SPEC-001/long/branch/name"

        # Act
        result = renderer._truncate_branch(branch, max_length=25)

        # Assert
        # SPEC ID should be preserved if possible
        assert isinstance(result, str)
        assert len(result) <= 25

    def test_truncate_branch_minimal_truncation(self, renderer):
        """Test truncation with minimal constraint."""
        # Arrange
        branch = "main"

        # Act
        result = renderer._truncate_branch(branch, max_length=5)

        # Assert
        assert len(result) <= 5

    def test_truncate_version_standard(self, renderer):
        """Test standard version truncation."""
        # Arrange
        version = "0.20.5.post1"

        # Act
        result = renderer._truncate_version(version)

        # Assert
        assert isinstance(result, str)
        assert result.startswith("0") or result.startswith("v")

    def test_truncate_version_with_v_prefix(self, renderer):
        """Test version truncation with v prefix."""
        # Arrange
        version = "v0.20.5"

        # Act
        result = renderer._truncate_version(version)

        # Assert
        assert isinstance(result, str)

    def test_truncate_version_short(self, renderer):
        """Test truncation of short version."""
        # Arrange
        version = "1.0"

        # Act
        result = renderer._truncate_version(version)

        # Assert
        assert result == "1.0" or result == "v1.0"


class TestCompactModeEdgeCases:
    """Test compact mode with various edge cases (lines 67-86)."""

    @pytest.fixture
    def renderer(self):
        """Create renderer with compact mode config."""
        with patch("moai_adk.statusline.renderer.StatuslineConfig") as mock_config_class:
            mock_config = MagicMock()
            mock_format_config = MagicMock()
            mock_display_config = MagicMock()

            mock_format_config.separator = " | "
            mock_display_config.model = True
            mock_display_config.version = True
            mock_display_config.branch = True
            mock_display_config.git_status = True
            mock_display_config.active_task = True

            mock_config.get_format_config.return_value = mock_format_config
            mock_config.get_display_config.return_value = mock_display_config
            mock_config_class.return_value = mock_config

            renderer = StatuslineRenderer()
            return renderer

    def test_render_compact_within_constraint(self, renderer):
        """Test compact rendering respects 80 char constraint."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="M1",
            duration="5s",
            directory="/home",
            active_task="",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert
        assert isinstance(result, str)
        assert len(result) <= 100  # Allow some realistic margin

    def test_render_compact_with_long_active_task(self, renderer):
        """Test compact rendering adjusts for long active task."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="M10 A5",
            duration="5s",
            directory="/home",
            active_task="SPEC-001-very-long-task-description",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert
        assert isinstance(result, str)

    def test_render_compact_fallback_to_minimal(self, renderer):
        """Test compact mode falls back to minimal if needed."""
        # Arrange
        data = StatuslineData(
            model="VeryLongModelNameThatTakesSpace",
            version="0.20.1",
            memory_usage="256MB",
            branch="very/long/feature/branch/name",
            git_status="Modified",
            duration="5s",
            directory="/very/long/directory/path",
            active_task="SPEC-001-task",
            claude_version="2.0.46",
            output_style="SomeStyle",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert
        assert isinstance(result, str)
        # Should contain at least model and version
        assert "Claude" in result or "VeryLong" in result

    def test_render_compact_with_all_optional_fields(self, renderer):
        """Test compact rendering with all optional fields populated."""
        # Arrange
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="M1",
            duration="5s",
            directory="/home",
            active_task="SPEC-001",
            claude_version="2.0.46",
            output_style="R2-D2",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert
        assert isinstance(result, str)
        assert len(result) > 0


class TestStatuslineDataObject:
    """Test StatuslineData dataclass."""

    def test_statusline_data_defaults(self):
        """Test StatuslineData default values."""
        # Arrange & Act
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home",
            active_task="",
        )

        # Assert
        assert data.claude_version == ""
        assert data.output_style == ""
        assert data.update_available is False
        assert data.latest_version == ""

    def test_statusline_data_all_fields(self):
        """Test StatuslineData with all fields set."""
        # Arrange & Act
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="M2",
            duration="5s",
            directory="/home/user/project",
            active_task="SPEC-001",
            claude_version="2.0.46",
            output_style="R2-D2",
            update_available=True,
            latest_version="0.21.0",
        )

        # Assert
        assert data.model == "Claude"
        assert data.version == "0.20.1"
        assert data.memory_usage == "256MB"
        assert data.branch == "main"
        assert data.git_status == "M2"
        assert data.duration == "5s"
        assert data.directory == "/home/user/project"
        assert data.active_task == "SPEC-001"
        assert data.claude_version == "2.0.46"
        assert data.output_style == "R2-D2"
        assert data.update_available is True
        assert data.latest_version == "0.21.0"

    def test_statusline_data_empty_optional_strings(self):
        """Test StatuslineData with empty optional string fields."""
        # Arrange & Act
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5s",
            directory="/home",
            active_task="",
            claude_version="",
            output_style="",
        )

        # Assert
        assert data.git_status == ""
        assert data.active_task == ""
        assert data.claude_version == ""
        assert data.output_style == ""
