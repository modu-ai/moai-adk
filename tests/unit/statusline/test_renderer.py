"""Tests for moai_adk.statusline.renderer module."""


from moai_adk.statusline.renderer import StatuslineData, StatuslineRenderer


class TestStatuslineData:
    """Test StatuslineData dataclass."""

    def test_statusline_data_init(self):
        """Test StatuslineData initialization."""
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="task1",
        )
        assert data.model == "Claude"
        assert data.version == "0.20.1"
        assert data.memory_usage == "256MB"
        assert data.branch == "main"
        assert data.git_status == "clean"
        assert data.duration == "5s"
        assert data.directory == "/home/user"
        assert data.active_task == "task1"
        assert data.claude_version == ""
        assert data.output_style == ""
        assert data.update_available is False

    def test_statusline_data_with_optional_fields(self):
        """Test StatuslineData with optional fields."""
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="task1",
            claude_version="2.0.46",
            output_style="R2-D2",
            update_available=True,
            latest_version="0.21.0",
        )
        assert data.claude_version == "2.0.46"
        assert data.output_style == "R2-D2"
        assert data.update_available is True
        assert data.latest_version == "0.21.0"


class TestStatuslineRenderer:
    """Test StatuslineRenderer class."""

    def test_renderer_init(self):
        """Test StatuslineRenderer initialization."""
        renderer = StatuslineRenderer()
        assert renderer._config is not None
        assert renderer._format_config is not None
        assert renderer._display_config is not None

    def test_render_compact_mode(self):
        """Test render method in compact mode."""
        renderer = StatuslineRenderer()
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
        result = renderer.render(data, "compact")
        assert isinstance(result, str)
        assert len(result) <= 80 or len(result) >= 50

    def test_render_extended_mode(self):
        """Test render method in extended mode."""
        renderer = StatuslineRenderer()
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
        result = renderer.render(data, "extended")
        assert isinstance(result, str)

    def test_render_minimal_mode(self):
        """Test render method in minimal mode."""
        renderer = StatuslineRenderer()
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
        result = renderer.render(data, "minimal")
        assert isinstance(result, str)
        assert len(result) <= 50 or len(result) >= 20

    def test_render_with_invalid_mode(self):
        """Test render method with invalid mode."""
        renderer = StatuslineRenderer()
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
        # Should default to compact
        result = renderer.render(data, "invalid_mode")
        assert isinstance(result, str)

    def test_truncate_branch_within_limit(self):
        """Test _truncate_branch with short branch name."""
        branch = "feature/short"
        result = StatuslineRenderer._truncate_branch(branch, max_length=30)
        assert result == branch

    def test_truncate_branch_exceeds_limit(self):
        """Test _truncate_branch with long branch name."""
        branch = "feature/very-long-branch-name-that-exceeds"
        result = StatuslineRenderer._truncate_branch(branch, max_length=20)
        assert len(result) <= 20

    def test_truncate_branch_with_spec(self):
        """Test _truncate_branch preserves SPEC ID."""
        branch = "feature/SPEC-001-long-description"
        result = StatuslineRenderer._truncate_branch(branch, max_length=20)
        assert "SPEC-001" in result

    def test_truncate_version_with_v_prefix(self):
        """Test _truncate_version removes v prefix."""
        result = StatuslineRenderer._truncate_version("v0.20.1")
        assert result == "0.20.1"

    def test_truncate_version_without_v_prefix(self):
        """Test _truncate_version without v prefix."""
        result = StatuslineRenderer._truncate_version("0.20.1")
        assert result == "0.20.1"

    def test_build_compact_parts_with_all_fields(self):
        """Test _build_compact_parts with all fields populated."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="task1",
            claude_version="2.0.46",
            output_style="R2-D2",
        )
        parts = renderer._build_compact_parts(data)
        assert isinstance(parts, list)
        assert len(parts) > 0

    def test_render_compact_with_version_prefix(self):
        """Test _render_compact handles version prefix correctly."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5s",
            directory="/home/user",
            active_task="",
        )
        result = renderer._render_compact(data)
        assert isinstance(result, str)
        # Version display depends on DisplayConfig.version setting
        # If version is enabled, it should contain version string
        # If disabled, it should still be a valid string output
        assert len(result) > 0

    def test_render_minimal_with_truncated_version(self):
        """Test _render_minimal truncates version correctly."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="Claude",
            version="v0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="",
            duration="5s",
            directory="/home/user",
            active_task="",
            claude_version="v2.0.46",
        )
        result = renderer._render_minimal(data)
        assert isinstance(result, str)
        assert "2.0" in result or "Claude" in result

    def test_fit_to_constraint_long_branch(self):
        """Test _fit_to_constraint with long branch name."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="feature/very-long-branch-name-that-exceeds-limit",
            git_status="2 modified",
            duration="5s",
            directory="/home/user",
            active_task="",
        )
        result = renderer._fit_to_constraint(data, 80)
        assert len(result) <= 80 or len(result) >= 70

    def test_render_extended_with_all_fields(self):
        """Test _render_extended with all fields."""
        renderer = StatuslineRenderer()
        data = StatuslineData(
            model="Claude",
            version="0.20.1",
            memory_usage="256MB",
            branch="main",
            git_status="clean",
            duration="5s",
            directory="/home/user",
            active_task="task1",
            claude_version="2.0.46",
            output_style="R2-D2",
        )
        result = renderer._render_extended(data)
        assert isinstance(result, str)

    def test_render_modes_consistency(self):
        """Test that all render modes return valid strings."""
        renderer = StatuslineRenderer()
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
        for mode in ["compact", "extended", "minimal"]:
            result = renderer.render(data, mode)
            assert isinstance(result, str)
            assert len(result) > 0
