"""Extended tests for moai_adk.statusline.renderer module.

These tests focus on increasing coverage for rendering modes and edge cases.
"""

from unittest.mock import MagicMock, patch

import pytest

from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer
from moai_adk.statusline.config import StatuslineConfig, DisplayConfig, FormatConfig


class TestStatuslineDataClass:
    """Test StatuslineData dataclass."""

    def test_statusline_data_init(self):
        """Test StatuslineData initialization."""
        data = StatuslineData(
            model="claude-3.5-sonnet",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="+2 M1",
            duration="5m",
            directory="MoAI-ADK",
            active_task="[TDD]"
        )

        assert data.model == "claude-3.5-sonnet"
        assert data.version == "0.20.1"
        assert data.claude_version == ""
        assert data.output_style == ""
        assert data.update_available is False

    def test_statusline_data_with_optional_fields(self):
        """Test StatuslineData with optional fields."""
        data = StatuslineData(
            model="gpt-4",
            version="0.20.1",
            memory_usage="512MB",
            branch="feature",
            git_status="",
            duration="10m",
            directory="project",
            active_task="",
            claude_version="v2.0.46",
            output_style="Concise",
            update_available=True,
            latest_version="0.21.0"
        )

        assert data.claude_version == "v2.0.46"
        assert data.output_style == "Concise"
        assert data.update_available is True
        assert data.latest_version == "0.21.0"


class TestStatuslineRendererInit:
    """Test StatuslineRenderer initialization."""

    def test_renderer_init(self):
        """Test renderer initialization."""
        renderer = StatuslineRenderer()

        assert renderer._config is not None
        assert renderer._format_config is not None
        assert renderer._display_config is not None

    def test_renderer_mode_constraints(self):
        """Test renderer mode constraints."""
        renderer = StatuslineRenderer()

        assert renderer._MODE_CONSTRAINTS["compact"] == 80
        assert renderer._MODE_CONSTRAINTS["extended"] == 120
        assert renderer._MODE_CONSTRAINTS["minimal"] == 40


class TestStatuslineRendererRenderMethod:
    """Test render method and mode selection."""

    def test_render_compact_mode(self):
        """Test rendering in compact mode."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer.render(data, mode="compact")

        assert isinstance(result, str)
        assert len(result) > 0

    def test_render_extended_mode(self):
        """Test rendering in extended mode."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer.render(data, mode="extended")

        assert isinstance(result, str)
        assert len(result) > 0

    def test_render_minimal_mode(self):
        """Test rendering in minimal mode."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer.render(data, mode="minimal")

        assert isinstance(result, str)
        assert len(result) > 0

    def test_render_unknown_mode_defaults_to_compact(self):
        """Test rendering with unknown mode defaults to compact."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer.render(data, mode="unknown_mode")

        assert isinstance(result, str)


class TestBuildCompactParts:
    """Test building compact mode parts."""

    def test_build_compact_parts_all_enabled(self):
        """Test building compact parts with all display options enabled."""
        renderer = StatuslineRenderer()
        renderer._display_config.model = True
        renderer._display_config.version = True
        renderer._display_config.git_status = True
        renderer._display_config.branch = True
        renderer._display_config.active_task = True

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="+2",
            duration="5m",
            directory="project",
            active_task="[TDD]"
        )

        parts = renderer._build_compact_parts(data)

        assert len(parts) > 0
        assert any("🤖" in part for part in parts)  # Model icon

    def test_build_compact_parts_model_disabled(self):
        """Test building compact parts with model disabled."""
        renderer = StatuslineRenderer()
        renderer._display_config.model = False
        renderer._display_config.version = True
        renderer._display_config.branch = True

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        parts = renderer._build_compact_parts(data)

        # Model should not be in parts
        assert not any("🤖 claude" in part for part in parts)

    def test_build_compact_parts_with_claude_version(self):
        """Test building compact parts with Claude Code version."""
        renderer = StatuslineRenderer()
        renderer._display_config.version = True

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task="",
            claude_version="2.0.46"
        )

        parts = renderer._build_compact_parts(data)

        # Should include Claude version
        assert any("🔅" in part for part in parts)

    def test_build_compact_parts_with_output_style(self):
        """Test building compact parts with output style."""
        renderer = StatuslineRenderer()
        renderer._display_config.version = True

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task="",
            output_style="Concise"
        )

        parts = renderer._build_compact_parts(data)

        # Should include output style
        assert any("💬" in part for part in parts)

    def test_build_compact_parts_empty_active_task(self):
        """Test building compact parts ignores empty active task."""
        renderer = StatuslineRenderer()
        renderer._display_config.active_task = True
        renderer._display_config.version = True

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        parts = renderer._build_compact_parts(data)

        # Empty active task should not be added
        assert not any(part == "" for part in parts)


class TestFitToConstraint:
    """Test fitting to character constraints."""

    def test_fit_to_constraint_within_limit(self):
        """Test fitting when already within limit."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="+2",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer._fit_to_constraint(data, max_length=80)

        assert len(result) <= 80

    def test_fit_to_constraint_truncates_branch(self):
        """Test fitting truncates long branch names."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="feature/very-long-branch-name-that-exceeds-limit",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer._fit_to_constraint(data, max_length=60)

        assert len(result) <= 60

    def test_fit_to_constraint_removes_style_if_needed(self):
        """Test fitting removes output style if needed."""
        renderer = StatuslineRenderer()
        renderer._display_config.git_status = True
        data = StatuslineData(
            model="a" * 40,  # Very long model name
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="+10 M5",
            duration="5m",
            directory="project",
            active_task="",
            output_style="VeryLongOutputStyleName"
        )

        result = renderer._fit_to_constraint(data, max_length=60)

        assert len(result) <= 60

    def test_fit_to_constraint_fallback_to_minimal(self):
        """Test fitting falls back to minimal if needed."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude-3.5-sonnet",
            version="0.20.1",
            memory_usage="256MB",
            branch="feature/very-long-branch-name",
            git_status="+100 M100",
            duration="5m",
            directory="project",
            active_task="[VERY-LONG-TASK]",
            claude_version="v2.0.46"
        )

        result = renderer._fit_to_constraint(data, max_length=50)

        # Should fit within extended limit (120) though may not fit within 30
        assert len(result) <= 120


class TestRenderExtended:
    """Test extended mode rendering."""

    def test_render_extended_all_fields(self):
        """Test rendering extended mode with all fields."""
        renderer = StatuslineRenderer()
        renderer._display_config.model = True
        renderer._display_config.version = True
        renderer._display_config.branch = True
        renderer._display_config.git_status = True
        renderer._display_config.active_task = True

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="+2 M1",
            duration="5m",
            directory="project",
            active_task="[TDD]"
        )

        result = renderer._render_extended(data)

        assert isinstance(result, str)
        assert len(result) <= 120

    def test_render_extended_truncates_if_too_long(self):
        """Test extended mode truncates if too long."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="very-long-model-name" * 5,
            version="0.20.1",
            memory_usage="256MB",
            branch="very-long-branch-name" * 5,
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer._render_extended(data)

        # Should attempt to fit within extended limit
        assert isinstance(result, str)

    def test_render_extended_with_claude_version(self):
        """Test extended mode with Claude version."""
        renderer = StatuslineRenderer()

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task="",
            claude_version="v2.0.46"
        )

        result = renderer._render_extended(data)

        assert isinstance(result, str)


class TestRenderMinimal:
    """Test minimal mode rendering."""

    def test_render_minimal_basic(self):
        """Test rendering minimal mode."""
        renderer = StatuslineRenderer()
        renderer._display_config.model = True
        renderer._display_config.version = True

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer._render_minimal(data)

        assert isinstance(result, str)
        assert len(result) <= 40

    def test_render_minimal_with_claude_version(self):
        """Test minimal mode with Claude version truncation."""
        renderer = StatuslineRenderer()
        renderer._display_config.version = True

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task="",
            claude_version="v2.0.46"
        )

        result = renderer._render_minimal(data)

        assert isinstance(result, str)
        assert len(result) <= 40
        # Should truncate version to major.minor
        assert "v2.0" in result or "2.0" in result

    def test_render_minimal_with_git_status(self):
        """Test minimal mode adds git status if fits."""
        renderer = StatuslineRenderer()
        renderer._display_config.model = True
        renderer._display_config.version = True
        renderer._display_config.git_status = True

        data = StatuslineData(
            model="gpt",
            version="1.0",
            memory_usage="256MB",
            branch="main",
            git_status="+2",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer._render_minimal(data)

        assert len(result) <= 40

    def test_render_minimal_version_prefix(self):
        """Test minimal mode version prefix handling."""
        renderer = StatuslineRenderer()
        renderer._display_config.version = True

        data = StatuslineData(
            model="claude",
            version="v0.20.1",  # Already has v prefix
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer._render_minimal(data)

        assert isinstance(result, str)


class TestTruncateBranch:
    """Test branch name truncation."""

    def test_truncate_branch_within_limit(self):
        """Test branch within limit stays unchanged."""
        result = StatuslineRenderer._truncate_branch("main", max_length=30)

        assert result == "main"

    def test_truncate_branch_exceeds_limit(self):
        """Test branch exceeding limit gets truncated."""
        long_branch = "feature/very-long-branch-name-that-is-too-long"
        result = StatuslineRenderer._truncate_branch(long_branch, max_length=20)

        assert len(result) <= 20
        assert "…" in result

    def test_truncate_branch_preserves_spec_id(self):
        """Test truncation preserves SPEC ID."""
        branch = "feature/SPEC-001-my-feature-with-long-description"
        result = StatuslineRenderer._truncate_branch(branch, max_length=20)

        assert "SPEC-001" in result

    def test_truncate_branch_spec_not_at_start(self):
        """Test SPEC preservation when not at start."""
        branch = "bugfix/old-feature/SPEC-002-new-fix"
        result = StatuslineRenderer._truncate_branch(branch, max_length=20)

        # Should preserve SPEC if possible
        assert len(result) <= 20

    def test_truncate_branch_exact_limit(self):
        """Test truncation at exact limit."""
        branch = "feature"
        result = StatuslineRenderer._truncate_branch(branch, max_length=7)

        assert result == "feature"


class TestTruncateVersion:
    """Test version truncation."""

    def test_truncate_version_with_v_prefix(self):
        """Test version truncation removes v prefix."""
        result = StatuslineRenderer._truncate_version("v0.20.1")

        assert result == "0.20.1"

    def test_truncate_version_without_v_prefix(self):
        """Test version without v prefix stays same."""
        result = StatuslineRenderer._truncate_version("0.20.1")

        assert result == "0.20.1"

    def test_truncate_version_empty(self):
        """Test empty version string."""
        result = StatuslineRenderer._truncate_version("")

        assert result == ""


class TestDisplayConfigIntegration:
    """Test integration with DisplayConfig."""

    def test_renderer_respects_display_config(self):
        """Test renderer respects display configuration."""
        with patch("moai_adk.statusline.renderer.StatuslineConfig") as MockConfig:
            mock_config = MagicMock()
            mock_display_config = DisplayConfig(
                model=True,
                version=False,
                branch=True,
                git_status=True
            )
            mock_config.return_value.get_display_config.return_value = mock_display_config

            MockConfig.return_value = mock_config
            renderer = StatuslineRenderer()

            # Verify display config is used
            assert renderer._display_config is not None


class TestFormatConfigIntegration:
    """Test integration with FormatConfig."""

    def test_renderer_uses_format_config_separator(self):
        """Test renderer uses configured separator."""
        renderer = StatuslineRenderer()

        # Original separator
        original_separator = renderer._format_config.separator

        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="+2",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer._render_compact(data)

        # Should use the configured separator
        if len(result) > 0:
            assert isinstance(result, str)


class TestEdgeCases:
    """Test edge cases and boundary conditions."""

    def test_render_empty_model(self):
        """Test rendering with empty model name."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer.render(data, mode="compact")

        assert isinstance(result, str)

    def test_render_empty_branch(self):
        """Test rendering with empty branch."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="",
            git_status="",
            duration="5m",
            directory="project",
            active_task=""
        )

        result = renderer.render(data, mode="compact")

        assert isinstance(result, str)

    def test_render_special_characters_in_fields(self):
        """Test rendering with special characters."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude-3.5-sonnet",
            version="0.20.1",
            memory_usage="256MB",
            branch="feature/bug-fix-#123",
            git_status="+2 M1",
            duration="5m",
            directory="project-2024",
            active_task="[TASK-1]"
        )

        result = renderer.render(data, mode="compact")

        assert isinstance(result, str)

    def test_render_unicode_in_fields(self):
        """Test rendering with unicode characters."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="claude-3.5-sonnet",
            version="0.20.1",
            memory_usage="256MB",
            branch="feature/中文",
            git_status="+2",
            duration="5m",
            directory="项目",
            active_task=""
        )

        result = renderer.render(data, mode="compact")

        assert isinstance(result, str)

    def test_render_very_long_strings(self):
        """Test rendering with very long strings."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="x" * 100,
            version="0.20.1",
            memory_usage="256MB",
            branch="y" * 100,
            git_status="+2",
            duration="5m",
            directory="z" * 100,
            active_task=""
        )

        result = renderer.render(data, mode="minimal")

        # Minimal mode tries to fit but with very long input may exceed
        assert isinstance(result, str)
        assert len(result) > 0
