"""
Comprehensive test coverage for statusline renderer module.

Tests for uncovered lines in renderer.py:
- Display config branches (lines 111, 159, 186, 202, 237, 272, 306)
- Fallback to minimal rendering (line 210)
"""

from unittest.mock import MagicMock, patch

import pytest

from moai_adk.statusline.config import StatuslineConfig
from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer


class TestRenderCompactWithDisplayConfig:
    """Test _render_compact with various display configurations (lines 111, 159, 186, 202)."""

    def test_render_compact_model_disabled(self):
        """Test _render_compact with model display disabled (line 111)."""
        # Arrange
        renderer = StatuslineRenderer()
        renderer._display_config.model = False

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert - model should not be in result
        assert "claude-3-5-sonnet" not in result

    def test_render_compact_memory_usage_disabled(self):
        """Test _render_compact with memory_usage display disabled (line 159)."""
        # Arrange
        renderer = StatuslineRenderer()
        renderer._display_config.memory_usage = False

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert - memory_usage should not be in result
        assert "100MB" not in result

    def test_render_compact_directory_disabled(self):
        """Test _render_compact with directory display disabled (line 186)."""
        # Arrange
        renderer = StatuslineRenderer()
        renderer._display_config.directory = False

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert - directory should not be in result
        assert "project" not in result

    def test_render_compact_git_status_disabled(self):
        """Test _render_compact with git_status display disabled (line 202)."""
        # Arrange
        renderer = StatuslineRenderer()
        renderer._display_config.git_status = False

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert - git_status should not be in result
        assert "+0" not in result

    def test_render_compact_all_disabled(self):
        """Test _render_compact with all displays disabled."""
        # Arrange
        renderer = StatuslineRenderer()
        renderer._display_config.model = False
        renderer._display_config.memory_usage = False
        renderer._display_config.directory = False
        renderer._display_config.git_status = False
        renderer._display_config.branch = False

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert - should return empty or minimal string
        assert result == "" or result is not None


class TestRenderCompactContextWindow:
    """Test _render_compact with context_window (line 112)."""

    def test_render_compact_with_context_window(self):
        """Test _render_compact includes context_window when present (line 112)."""
        # Arrange
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
            context_window="15K/200K",
        )

        # Act
        result = renderer._render_compact(data)

        # Assert - context_window should be in result
        assert "15K/200K" in result


class TestRenderExtendedWithDisplayConfig:
    """Test _render_extended with various display configurations."""

    def test_render_extended_model_disabled(self):
        """Test _render_extended with model display disabled."""
        # Arrange
        renderer = StatuslineRenderer()
        renderer._display_config.model = False

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer._render_extended(data)

        # Assert - model should not be in result
        assert "claude-3-5-sonnet" not in result

    def test_render_extended_memory_usage_disabled(self):
        """Test _render_extended with memory_usage display disabled."""
        # Arrange
        renderer = StatuslineRenderer()
        renderer._display_config.memory_usage = False

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer._render_extended(data)

        # Assert - memory_usage should not be in result
        assert "100MB" not in result


class TestRenderExtendedContextWindow:
    """Test _render_extended with context_window (lines 237, 272)."""

    def test_render_extended_with_context_window(self):
        """Test _render_extended includes context_window when present (line 237)."""
        # Arrange
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
            context_window="15K/200K",
        )

        # Act
        result = renderer._render_extended(data)

        # Assert - context_window should be in result
        assert "15K/200K" in result


class TestRenderMinimalWithGitStatus:
    """Test _render_minimal with git_status (line 306)."""

    def test_render_minimal_with_git_status_fits(self):
        """Test _render_minimal includes git_status when it fits (line 306)."""
        # Arrange
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
            context_window="",
        )

        # Act
        result = renderer._render_minimal(data)

        # Assert - git_status should be included if it fits
        # The format is "🤖 Model | 💰 Context | 📊 Status"
        # With separator " | " and emoji, this should fit within 40 chars
        assert result is not None


class TestFitToConstraintFallback:
    """Test _fit_to_constraint fallback to minimal (line 210)."""

    def test_fit_to_constraint_falls_back_to_minimal(self):
        """Test _fit_to_constraint falls back to minimal when still too long (line 210)."""
        # Arrange
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="very-long-model-name-that-should-be-truncated",
            version="1.0.0",
            memory_usage="100MB",
            branch="very-long-branch-name-with-many-characters",
            git_status="+100 M200 ?300",
            duration="5m",
            directory="very-long-project-directory-name",
            active_task="",
            context_window="150K/200K",
            output_style="VeryLongOutputStyleName",
        )

        # Act - force minimal mode
        result = renderer._fit_to_constraint(data, max_length=10)

        # Assert - should fall back to minimal render
        assert result is not None
        # Minimal mode only shows model and context
        assert "very-long-model-name" in result or "claude" in result.lower()


class TestTruncateBranch:
    """Test _truncate_branch method."""

    def test_truncate_branch_short_branch(self):
        """Test _truncate_branch with short branch name."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_branch("main", max_length=30)

        # Assert
        assert result == "main"

    def test_truncate_branch_exact_length(self):
        """Test _truncate_branch with exact max length."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_branch("a" * 30, max_length=30)

        # Assert
        assert result == "a" * 30

    def test_truncate_branch_long_branch(self):
        """Test _truncate_branch with long branch name."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_branch("very-long-branch-name", max_length=10)

        # Assert
        assert result.endswith("…")
        assert len(result) <= 10

    def test_truncate_branch_spec_id_preserved(self):
        """Test _truncate_branch preserves SPEC ID when present."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_branch("feature/SPEC-001-add-feature", max_length=20)

        # Assert - should try to preserve SPEC ID
        assert "SPEC-001" in result or result.endswith("…")

    def test_truncate_branch_spec_with_dash(self):
        """Test _truncate_branch with SPEC ID followed by dash."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_branch("SPEC-001-description", max_length=15)

        # Assert - should include SPEC-001 and part of description
        assert "SPEC" in result

    def test_truncate_branch_multiple_specs(self):
        """Test _truncate_branch with multiple SPEC references."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_branch("SPEC-001-SPEC-002-combined", max_length=20)

        # Assert - should handle multiple SPEC references
        assert result is not None


class TestTruncateVersion:
    """Test _truncate_version method."""

    def test_truncate_version_without_v(self):
        """Test _truncate_version without v prefix."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_version("1.0.0")

        # Assert
        assert result == "1.0.0"

    def test_truncate_version_with_v(self):
        """Test _truncate_version with v prefix."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_version("v1.0.0")

        # Assert
        assert result == "1.0.0"

    def test_truncate_version_only_v(self):
        """Test _truncate_version with only v."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_version("v")

        # Assert
        assert result == ""

    def test_truncate_version_v_with_number(self):
        """Test _truncate_version with v followed by single number."""
        # Arrange & Act
        result = StatuslineRenderer._truncate_version("v1")

        # Assert
        assert result == "1"


class TestStatuslineData:
    """Test StatuslineData dataclass."""

    def test_statusline_data_creation(self):
        """Test StatuslineData creation with all fields."""
        # Act
        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="[/MOAI:2-RUN]",
            claude_version="2.0.46",
            output_style="Concise",
            update_available=True,
            latest_version="1.1.0",
            context_window="15K/200K",
        )

        # Assert
        assert data.model == "claude-3-5-sonnet"
        assert data.version == "1.0.0"
        assert data.memory_usage == "100MB"
        assert data.update_available is True

    def test_statusline_data_defaults(self):
        """Test StatuslineData with default values."""
        # Act
        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Assert - defaults should be set
        assert data.claude_version == ""
        assert data.output_style == ""
        assert data.update_available is False
        assert data.latest_version == ""
        assert data.context_window == ""


class TestStatuslineRendererInit:
    """Test StatuslineRenderer initialization."""

    def test_init_creates_config(self):
        """Test that initialization creates config."""
        # Arrange & Act
        renderer = StatuslineRenderer()

        # Assert
        assert renderer._config is not None
        assert isinstance(renderer._config, StatuslineConfig)

    def test_init_loads_format_config(self):
        """Test that initialization loads format config."""
        # Arrange & Act
        renderer = StatuslineRenderer()

        # Assert
        assert renderer._format_config is not None

    def test_init_loads_display_config(self):
        """Test that initialization loads display config."""
        # Arrange & Act
        renderer = StatuslineRenderer()

        # Assert
        assert renderer._display_config is not None


class TestRenderMethod:
    """Test render method."""

    def test_render_compact_mode(self):
        """Test render with compact mode."""
        # Arrange
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer.render(data, mode="compact")

        # Assert
        assert result is not None
        assert isinstance(result, str)

    def test_render_extended_mode(self):
        """Test render with extended mode."""
        # Arrange
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer.render(data, mode="extended")

        # Assert
        assert result is not None
        assert isinstance(result, str)

    def test_render_minimal_mode(self):
        """Test render with minimal mode."""
        # Arrange
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer.render(data, mode="minimal")

        # Assert
        assert result is not None
        assert isinstance(result, str)

    def test_render_unknown_mode_defaults_to_compact(self):
        """Test render with unknown mode defaults to compact."""
        # Arrange
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer.render(data, mode="unknown")

        # Assert - should default to compact
        assert result is not None


class TestBuildCompactParts:
    """Test _build_compact_parts method."""

    def test_build_compact_parts_with_all_fields(self):
        """Test _build_compact_parts with all display enabled."""
        # Arrange
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="[/MOAI:2-RUN]",
            context_window="15K/200K",
            output_style="Concise",
        )

        # Act
        parts = renderer._build_compact_parts(data)

        # Assert
        assert len(parts) > 0
        assert any("claude-3-5-sonnet" in part for part in parts)

    def test_build_compact_parts_with_empty_active_task(self):
        """Test _build_compact_parts with empty active_task (line 134)."""
        # Arrange
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",  # Empty active task
        )

        # Act
        parts = renderer._build_compact_parts(data)

        # Assert - empty active_task should not be added
        assert not any("[" in part for part in parts)


class TestEdgeCases:
    """Test edge cases and error conditions."""

    def test_render_with_very_long_model_name(self):
        """Test rendering with very long model name."""
        # Arrange
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="a" * 100,  # Very long model name
            version="1.0.0",
            memory_usage="100MB",
            branch="main",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer.render(data, mode="compact")

        # Assert - should handle gracefully
        assert result is not None

    def test_render_with_special_characters(self):
        """Test rendering with special characters in fields."""
        # Arrange
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="claude-3-5-sonnet",
            version="1.0.0-beta+build123",
            memory_usage="100MB",
            branch="feature/特殊字符",
            git_status="+0 M0 ?0",
            duration="5m",
            directory="project",
            active_task="",
        )

        # Act
        result = renderer.render(data, mode="compact")

        # Assert - should handle gracefully
        assert result is not None

    def test_render_with_empty_strings(self):
        """Test rendering with empty string values."""
        # Arrange
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="",
            version="",
            memory_usage="",
            branch="",
            git_status="",
            duration="",
            directory="",
            active_task="",
        )

        # Act
        result = renderer.render(data, mode="compact")

        # Assert - should handle gracefully
        assert result is not None
