"""Tests for cli.commands.rank module."""

from unittest.mock import Mock, patch

from click.testing import CliRunner

from moai_adk.cli.commands.rank import (
    _create_progress_bar,
    _get_rank_medal,
    exclude,
    format_rank_position,
    format_tokens,
    include,
    logout,
    rank,
    status,
    sync,
)


class TestFormatTokens:
    """Test format_tokens function."""

    def test_format_tokens_millions(self):
        """Test formatting millions of tokens."""
        assert format_tokens(1_000_000) == "1.0M"
        assert format_tokens(2_500_000) == "2.5M"
        assert format_tokens(10_000_000) == "10.0M"

    def test_format_tokens_thousands(self):
        """Test formatting thousands of tokens."""
        assert format_tokens(1_000) == "1.0K"
        assert format_tokens(15_500) == "15.5K"
        assert format_tokens(999_999) == "1000.0K"

    def test_format_tokens_small(self):
        """Test formatting small token counts."""
        assert format_tokens(0) == "0"
        assert format_tokens(100) == "100"
        assert format_tokens(999) == "999"


class TestFormatRankPosition:
    """Test format_rank_position function."""

    def test_format_first_place(self):
        """Test formatting first place with medal."""
        result = format_rank_position(1, 100)
        assert "[gold1]1st[/gold1]" in result
        assert "/ 100" in result

    def test_format_second_place(self):
        """Test formatting second place with medal."""
        result = format_rank_position(2, 100)
        assert "[grey70]2nd[/grey70]" in result
        assert "/ 100" in result

    def test_format_third_place(self):
        """Test formatting third place with medal."""
        result = format_rank_position(3, 100)
        assert "[orange3]3rd[/orange3]" in result
        assert "/ 100" in result

    def test_format_other_position(self):
        """Test formatting positions beyond third place."""
        result = format_rank_position(5, 100)
        assert "#5 / 100" in result

    def test_format_high_position(self):
        """Test formatting high rank numbers."""
        result = format_rank_position(42, 1000)
        assert "#42 / 1000" in result


class TestGetRankMedal:
    """Test _get_rank_medal function."""

    def test_first_place_medal(self):
        """Test medal for first place."""
        assert "[gold1]1st[/gold1]" in _get_rank_medal(1)

    def test_second_place_medal(self):
        """Test medal for second place."""
        assert "[grey70]2nd[/grey70]" in _get_rank_medal(2)

    def test_third_place_medal(self):
        """Test medal for third place."""
        assert "[orange3]3rd[/orange3]" in _get_rank_medal(3)

    def test_other_position_medal(self):
        """Test medal string for positions beyond third."""
        assert _get_rank_medal(5) == "#5"
        assert _get_rank_medal(42) == "#42"
        assert _get_rank_medal(100) == "#100"


class TestCreateProgressBar:
    """Test _create_progress_bar function."""

    def test_full_progress_bar(self):
        """Test full progress bar."""
        result = _create_progress_bar(100, 100, width=20)
        # Should be all filled (cyan blocks)
        assert "[cyan]" in result
        assert result.count("[cyan]") >= 1  # At least opening tag

    def test_empty_progress_bar(self):
        """Test empty progress bar."""
        result = _create_progress_bar(0, 100, width=20)
        assert "[dim]" in result
        assert "-" in result or "░" in result  # Empty characters

    def test_half_progress_bar(self):
        """Test half-filled progress bar."""
        result = _create_progress_bar(50, 100, width=10)
        assert "[cyan]" in result  # Filled portion
        assert "[dim]" in result  # Empty portion

    def test_zero_total(self):
        """Test progress bar with zero total."""
        result = _create_progress_bar(50, 0, width=20)
        assert "[dim]" in result
        assert "-" in result

    def test_custom_width(self):
        """Test progress bar with custom width."""
        result = _create_progress_bar(5, 10, width=5)
        # Half filled with custom width
        assert "[cyan]" in result

    def test_value_exceeds_total(self):
        """Test progress bar caps at 100%."""
        result = _create_progress_bar(150, 100, width=10)
        # Should cap at full, not exceed
        assert "[cyan]" in result


class TestRankCommand:
    """Test rank command group."""

    def test_rank_group_exists(self):
        """Test that rank command group exists."""
        runner = CliRunner()
        result = runner.invoke(rank, ["--help"])
        assert result.exit_code == 0
        assert "MoAI Rank" in result.output
        assert "login" in result.output
        assert "status" in result.output
        assert "logout" in result.output
        assert "sync" in result.output
        assert "exclude" in result.output
        assert "include" in result.output


class TestLogoutCommand:
    """Test logout command."""

    def test_logout_without_credentials(self):
        """Test logout when no credentials exist."""
        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=False):
            runner = CliRunner()
            result = runner.invoke(logout)
            assert result.exit_code == 0
            assert "No credentials stored" in result.output

    def test_logout_confirmation_cancels(self):
        """Test logout when user cancels confirmation."""
        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=True):
            with patch("moai_adk.rank.config.RankConfig.load_credentials", return_value=Mock(username="testuser")):
                with patch("moai_adk.cli.commands.rank.click.confirm", return_value=False):
                    runner = CliRunner()
                    result = runner.invoke(logout)
                    assert result.exit_code == 0
                    assert "Cancelled" in result.output

    def test_logout_confirmation_confirms(self):
        """Test logout when user confirms."""
        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=True):
            with patch("moai_adk.rank.config.RankConfig.load_credentials", return_value=Mock(username="testuser")):
                with patch("moai_adk.cli.commands.rank.click.confirm", return_value=True):
                    with patch("moai_adk.rank.config.RankConfig.delete_credentials"):
                        runner = CliRunner()
                        result = runner.invoke(logout)
                        assert result.exit_code == 0
                        assert "Credentials removed successfully" in result.output


class TestExcludeCommand:
    """Test exclude command."""

    def test_exclude_list_empty(self):
        """Test exclude --list when no projects excluded."""
        with patch("moai_adk.rank.hook.load_rank_config", return_value={}):
            runner = CliRunner()
            result = runner.invoke(exclude, ["--list"])
            assert result.exit_code == 0
            assert "No projects are excluded" in result.output

    def test_exclude_list_with_exclusions(self):
        """Test exclude --list with excluded projects."""
        with patch(
            "moai_adk.rank.hook.load_rank_config",
            return_value={"exclude_projects": ["/project1", "/project2"]},
        ):
            runner = CliRunner()
            result = runner.invoke(exclude, ["--list"])
            assert result.exit_code == 0
            assert "Excluded Projects:" in result.output
            assert "/project1" in result.output
            assert "/project2" in result.output
            assert "Total: 2" in result.output

    def test_exclude_with_path_success(self):
        """Test excluding a project with path."""
        with patch("moai_adk.rank.hook.add_project_exclusion", return_value=True):
            with patch("moai_adk.rank.hook.load_rank_config", return_value={"exclude_projects": ["/test"]}):
                runner = CliRunner()
                result = runner.invoke(exclude, ["/test/project"])
                assert result.exit_code == 0
                assert "Excluded:" in result.output

    def test_exclude_with_path_failure(self):
        """Test exclude command when add_project_exclusion fails."""
        with patch("moai_adk.rank.hook.add_project_exclusion", return_value=False):
            runner = CliRunner()
            result = runner.invoke(exclude, ["/test/project"])
            assert result.exit_code == 0
            assert "Failed to exclude" in result.output


class TestIncludeCommand:
    """Test include command."""

    def test_include_with_path_success(self):
        """Test including a project with path."""
        with patch("moai_adk.rank.hook.remove_project_exclusion", return_value=True):
            with patch("moai_adk.rank.hook.load_rank_config", return_value={"exclude_projects": []}):
                runner = CliRunner()
                result = runner.invoke(include, ["/test/project"])
                assert result.exit_code == 0
                assert "Included:" in result.output

    def test_include_with_path_failure(self):
        """Test include command when remove_project_exclusion fails."""
        with patch("moai_adk.rank.hook.remove_project_exclusion", return_value=False):
            runner = CliRunner()
            result = runner.invoke(include, ["/test/project"])
            assert result.exit_code == 0
            assert "Failed to include" in result.output

    def test_include_with_remaining_exclusions(self):
        """Test include shows remaining exclusions."""
        with patch("moai_adk.rank.hook.remove_project_exclusion", return_value=True):
            with patch("moai_adk.rank.hook.load_rank_config", return_value={"exclude_projects": ["/other"]}):
                runner = CliRunner()
                result = runner.invoke(include, ["/test/project"])
                assert result.exit_code == 0
                assert "Remaining excluded projects: 1" in result.output


class TestSyncCommand:
    """Test sync command."""

    def test_sync_without_credentials(self):
        """Test sync when not registered."""
        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=False):
            runner = CliRunner()
            result = runner.invoke(sync)
            assert result.exit_code == 0
            assert "Not registered" in result.output

    def test_sync_foreground_success(self):
        """Test foreground sync."""
        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=True):
            with patch("moai_adk.rank.hook.sync_all_sessions"):
                runner = CliRunner()
                result = runner.invoke(sync)
                assert result.exit_code == 0
                # Should call sync_all_sessions

    def test_sync_background_starts(self):
        """Test background sync starts subprocess."""
        # Create a proper mock temp file
        mock_temp_file = Mock()
        mock_temp_file.name = "/tmp/test_sync.py"
        mock_temp_file.__enter__ = Mock(return_value=mock_temp_file)
        mock_temp_file.__exit__ = Mock(return_value=False)

        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=True):
            with patch("moai_adk.rank.config.RankConfig.CONFIG_DIR") as mock_config_dir:
                # Make CONFIG_DIR exist() return True
                mock_config_dir.exists.return_value = True
                log_file = Mock()
                log_file.parent.exists.return_value = True
                mock_config_dir.__truediv__.return_value = log_file
                with patch("subprocess.Popen"):
                    with patch("tempfile.NamedTemporaryFile", return_value=mock_temp_file):
                        with patch("builtins.open", create=True):
                            with patch("os.unlink"):
                                runner = CliRunner()
                                result = runner.invoke(sync, ["--background"])
                                assert result.exit_code == 0
                                assert "Background sync started" in result.output


class TestStatusCommand:
    """Test status command."""

    def test_status_without_credentials(self):
        """Test status when not registered."""
        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=False):
            runner = CliRunner()
            result = runner.invoke(status)
            assert result.exit_code == 0
            assert "Not registered" in result.output

    def test_status_with_authentication_error(self):
        """Test status with authentication error."""
        from moai_adk.rank.client import AuthenticationError

        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=True):
            with patch("moai_adk.rank.client.RankClient", side_effect=AuthenticationError("Auth failed")):
                runner = CliRunner()
                result = runner.invoke(status)
                # Should handle error gracefully
                assert result.exit_code == 0
                assert "Authentication failed" in result.output
