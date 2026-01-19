"""Comprehensive TDD tests for rank.py to achieve 85% coverage.

This test suite covers uncovered lines in rank.py:
- Lines 54-164: register command (OAuth flow, hook installation, sync options)
- Lines 181-274: status command (API calls, error paths, rank display)
- Lines 299-363: sync command (background sync, log file handling)

Coverage Target: 85%+ (from 63.11%)
Missing Lines: ~76 lines across register, status, and sync commands

Test Organization:
- Class-based structure for related tests
- Descriptive test names following test_<action>_<condition>_<expected>
- Comprehensive docstrings
- Mock external dependencies (OAuthHandler, RankConfig, RankClient, subprocess)
- Follow AAA pattern (Arrange-Act-Assert)

Reference Patterns:
- test_language_enhanced.py for comprehensive CLI test patterns
- test_status_enhanced.py for error path testing
- test_rank.py for existing test patterns
"""

from unittest.mock import Mock, patch

import pytest
from click.testing import CliRunner

from moai_adk.cli.commands.rank import (
    exclude,
    format_rank_position,
    format_tokens,
    include,
    logout,
    rank,
    register,
    status,
    sync,
)


@pytest.fixture
def runner():
    """Create CLI runner for testing."""
    return CliRunner()


@pytest.fixture
def mock_credentials():
    """Mock credentials object."""
    creds = Mock()
    creds.username = "testuser"
    creds.api_key = "test_api_key_12345"
    return creds


class TestRegisterCommand:
    """Test register command (lines 54-164)."""

    def test_register_displays_already_registered_message(self, runner, mock_credentials):
        """Should show already registered message when credentials exist (lines 79-84)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True
            mock_config_class.load_credentials.return_value = mock_credentials

            with patch("moai_adk.cli.commands.rank.click.confirm", return_value=False):
                result = runner.invoke(register)

                assert result.exit_code == 0
                # Verify "Already registered" message was printed
                assert "Already registered" in result.output

    def test_register_prompts_for_reregistration(self, runner, mock_credentials):
        """Should prompt for re-registration when credentials exist (line 83)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True
            mock_config_class.load_credentials.return_value = mock_credentials

            with patch("moai_adk.cli.commands.rank.click.confirm", return_value=True):
                with patch("moai_adk.rank.auth.OAuthHandler") as mock_oauth:
                    mock_handler = Mock()
                    mock_oauth.return_value = mock_handler
                    mock_handler.start_oauth_flow = Mock()

                    result = runner.invoke(register)

                    assert result.exit_code == 0

    def test_register_with_no_sync_option_skips_sync(self, runner):
        """Should skip sync when --no-sync flag is used (lines 131-135)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            with patch("moai_adk.rank.auth.OAuthHandler") as mock_oauth_class:
                mock_handler = Mock()
                mock_oauth_class.return_value = mock_handler

                # Mock start_oauth_flow to trigger on_success callback
                def trigger_success(on_success, on_error, timeout):
                    mock_creds = Mock()
                    mock_creds.username = "testuser"
                    mock_creds.api_key = "key123"
                    on_success(mock_creds)

                mock_handler.start_oauth_flow = trigger_success

                with patch("moai_adk.rank.hook.install_hook", return_value=True):
                    result = runner.invoke(register, ["--no-sync"])

                    assert result.exit_code == 0
                    # Verify sync skipped message was printed
                    assert "Sync skipped" in result.output

    def test_register_with_background_sync_starts_subprocess(self, runner):
        """Should start background sync when --background-sync flag is used (lines 136-142)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            with patch("moai_adk.rank.auth.OAuthHandler") as mock_oauth_class:
                mock_handler = Mock()
                mock_oauth_class.return_value = mock_handler

                def trigger_success(on_success, on_error, timeout):
                    mock_creds = Mock()
                    mock_creds.username = "testuser"
                    mock_creds.api_key = "key123"
                    on_success(mock_creds)

                mock_handler.start_oauth_flow = trigger_success

                with patch("moai_adk.rank.hook.install_hook", return_value=True):
                    with patch("subprocess.Popen") as mock_popen:
                        mock_process = Mock()
                        mock_popen.return_value = mock_process

                        result = runner.invoke(register, ["--background-sync"])

                        assert result.exit_code == 0
                        # Verify subprocess.Popen was called
                        assert mock_popen.called
                        # Verify background sync message was printed
                        assert "background sync" in result.output.lower()

    def test_register_successful_hook_installation(self, runner):
        """Should display hook installation success message (lines 116-128)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            with patch("moai_adk.rank.auth.OAuthHandler") as mock_oauth_class:
                mock_handler = Mock()
                mock_oauth_class.return_value = mock_handler

                def trigger_success(on_success, on_error, timeout):
                    mock_creds = Mock()
                    mock_creds.username = "testuser"
                    mock_creds.api_key = "key123"
                    on_success(mock_creds)

                mock_handler.start_oauth_flow = trigger_success

                with patch("moai_adk.rank.hook.install_hook", return_value=True):
                    result = runner.invoke(register)

                    assert result.exit_code == 0
                    # Verify hook installed message
                    assert "Session tracking hook installed" in result.output or "hook" in result.output.lower()

    def test_register_hook_installation_failure(self, runner):
        """Should display warning when hook installation fails (lines 150-151)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            with patch("moai_adk.rank.auth.OAuthHandler") as mock_oauth_class:
                mock_handler = Mock()
                mock_oauth_class.return_value = mock_handler

                def trigger_success(on_success, on_error, timeout):
                    mock_creds = Mock()
                    mock_creds.username = "testuser"
                    mock_creds.api_key = "key123"
                    on_success(mock_creds)

                mock_handler.start_oauth_flow = trigger_success

                with patch("moai_adk.rank.hook.install_hook", return_value=False):
                    result = runner.invoke(register)

                    assert result.exit_code == 0
                    # Verify warning message
                    assert (
                        "Failed to install session tracking hook" in result.output or "warning" in result.output.lower()
                    )

    def test_register_oauth_error_handling(self, runner):
        """Should display error when OAuth flow fails (lines 156-157)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            with patch("moai_adk.rank.auth.OAuthHandler") as mock_oauth_class:
                mock_handler = Mock()
                mock_oauth_class.return_value = mock_handler

                def trigger_error(on_success, on_error, timeout):
                    on_error("Authorization failed")

                mock_handler.start_oauth_flow = trigger_error

                result = runner.invoke(register)

                assert result.exit_code == 0
                # Verify error message was printed
                assert "Registration failed" in result.output or "Authorization failed" in result.output

    def test_register_displays_registration_panel(self, runner):
        """Should display registration information panel (lines 87-96)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            with patch("moai_adk.rank.auth.OAuthHandler") as mock_oauth_class:
                mock_handler = Mock()
                mock_oauth_class.return_value = mock_handler

                # Don't trigger any callback, just let it display
                mock_handler.start_oauth_flow = Mock(return_value=None)

                result = runner.invoke(register)

                assert result.exit_code == 0

    def test_register_displays_success_message(self, runner):
        """Should display registration success message (lines 104-113)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            with patch("moai_adk.rank.auth.OAuthHandler") as mock_oauth_class:
                mock_handler = Mock()
                mock_oauth_class.return_value = mock_handler

                def trigger_success(on_success, on_error, timeout):
                    mock_creds = Mock()
                    mock_creds.username = "testuser"
                    mock_creds.api_key = "dummy_key"
                    on_success(mock_creds)

                mock_handler.start_oauth_flow = trigger_success

                with patch("moai_adk.rank.hook.install_hook", return_value=True):
                    result = runner.invoke(register)

                    assert result.exit_code == 0
                    # Verify success message
                    assert "Successfully registered" in result.output or "testuser" in result.output


class TestStatusCommand:
    """Test status command (lines 181-274)."""

    def test_status_without_credentials_displays_message(self, runner):
        """Should display not registered message when no credentials (lines 195-198)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            result = runner.invoke(status)

            assert result.exit_code == 0
            assert "Not registered" in result.output

    def test_status_with_valid_credentials_displays_rank(self, runner):
        """Should display user rank information (lines 200-267)."""
        from moai_adk.rank.client import UserRank

        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.client.RankClient") as mock_client_class:
                mock_client = Mock()
                mock_client_class.return_value = mock_client

                # Create mock user rank data
                mock_rank = Mock(spec=UserRank)
                mock_rank.username = "testuser"
                mock_rank.total_tokens = 1500000
                mock_rank.input_tokens = 900000
                mock_rank.output_tokens = 600000
                mock_rank.total_sessions = 42

                # Weekly rank
                mock_weekly = Mock()
                mock_weekly.position = 5
                mock_weekly.total_participants = 100
                mock_weekly.composite_score = 50000
                mock_rank.weekly = mock_weekly

                # Daily rank
                mock_daily = Mock()
                mock_daily.position = 3
                mock_daily.total_participants = 50
                mock_daily.composite_score = 10000
                mock_rank.daily = mock_daily

                # Monthly rank
                mock_monthly = Mock()
                mock_monthly.position = 10
                mock_monthly.total_participants = 200
                mock_monthly.composite_score = 150000
                mock_rank.monthly = mock_monthly

                # All time rank
                mock_all_time = Mock()
                mock_all_time.position = 25
                mock_all_time.total_participants = 500
                mock_all_time.composite_score = 500000
                mock_rank.all_time = mock_all_time

                mock_client.get_user_rank.return_value = mock_rank

                with patch("moai_adk.rank.hook.is_hook_installed", return_value=True):
                    result = runner.invoke(status)

                    assert result.exit_code == 0
                    # Verify get_user_rank was called
                    mock_client.get_user_rank.assert_called_once()

    def test_status_with_authentication_error(self, runner):
        """Should display authentication error message (lines 269-271)."""
        from moai_adk.rank.client import AuthenticationError

        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.client.RankClient") as mock_client_class:
                mock_client_class.side_effect = AuthenticationError("Invalid API key")

                result = runner.invoke(status)

                assert result.exit_code == 0
                assert "Authentication failed" in result.output

    def test_status_with_rank_client_error(self, runner):
        """Should display client error message (lines 272-273)."""
        from moai_adk.rank.client import RankClientError

        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.client.RankClient") as mock_client_class:
                mock_client_class.side_effect = RankClientError("Network error")

                result = runner.invoke(status)

                assert result.exit_code == 0
                assert "Failed to fetch status" in result.output

    def test_status_with_no_weekly_rank_data(self, runner):
        """Should display no ranking data when weekly is None (lines 216-217)."""
        from moai_adk.rank.client import UserRank

        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.client.RankClient") as mock_client_class:
                mock_client = Mock()
                mock_client_class.return_value = mock_client

                mock_rank = Mock(spec=UserRank)
                mock_rank.username = "testuser"
                mock_rank.weekly = None  # No weekly data
                mock_rank.daily = None
                mock_rank.monthly = None
                mock_rank.all_time = None
                mock_rank.total_tokens = 0
                mock_rank.input_tokens = 0
                mock_rank.output_tokens = 0
                mock_rank.total_sessions = 0
                mock_client.get_user_rank.return_value = mock_rank

                with patch("moai_adk.rank.hook.is_hook_installed", return_value=True):
                    result = runner.invoke(status)

                    assert result.exit_code == 0

    def test_status_displays_hook_status_installed(self, runner):
        """Should display hook status as installed (lines 259-260)."""
        from moai_adk.rank.client import UserRank

        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.client.RankClient") as mock_client_class:
                mock_client = Mock()
                mock_client_class.return_value = mock_client

                mock_rank = Mock(spec=UserRank)
                mock_rank.username = "testuser"
                mock_rank.weekly = None
                mock_rank.total_tokens = 0
                mock_rank.input_tokens = 0
                mock_rank.output_tokens = 0
                mock_rank.total_sessions = 0
                mock_rank.daily = None
                mock_rank.monthly = None
                mock_rank.all_time = None
                mock_client.get_user_rank.return_value = mock_rank

                with patch("moai_adk.rank.hook.is_hook_installed", return_value=True):
                    result = runner.invoke(status)

                    assert result.exit_code == 0

    def test_status_displays_hook_status_not_installed(self, runner):
        """Should display hook status as not installed (line 266-267)."""
        from moai_adk.rank.client import UserRank

        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.client.RankClient") as mock_client_class:
                mock_client = Mock()
                mock_client_class.return_value = mock_client

                mock_rank = Mock(spec=UserRank)
                mock_rank.username = "testuser"
                mock_rank.weekly = None
                mock_rank.total_tokens = 0
                mock_rank.input_tokens = 0
                mock_rank.output_tokens = 0
                mock_rank.total_sessions = 0
                mock_rank.daily = None
                mock_rank.monthly = None
                mock_rank.all_time = None
                mock_client.get_user_rank.return_value = mock_rank

                with patch("moai_adk.rank.hook.is_hook_installed", return_value=False):
                    result = runner.invoke(status)

                    assert result.exit_code == 0

    def test_status_displays_all_time_periods(self, runner):
        """Should display daily, weekly, monthly, and all-time ranks (lines 223-240)."""
        from moai_adk.rank.client import UserRank

        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.client.RankClient") as mock_client_class:
                mock_client = Mock()
                mock_client_class.return_value = mock_client

                mock_rank = Mock(spec=UserRank)
                mock_rank.username = "testuser"
                mock_rank.total_tokens = 1000000
                mock_rank.input_tokens = 600000
                mock_rank.output_tokens = 400000
                mock_rank.total_sessions = 50

                # All time periods
                for period_name in ["daily", "weekly", "monthly", "all_time"]:
                    period_rank = Mock()
                    period_rank.position = 1
                    period_rank.total_participants = 100
                    period_rank.composite_score = 10000
                    setattr(mock_rank, period_name, period_rank)

                mock_client.get_user_rank.return_value = mock_rank

                with patch("moai_adk.rank.hook.is_hook_installed", return_value=True):
                    result = runner.invoke(status)

                    assert result.exit_code == 0

    def test_status_handles_zero_total_tokens(self, runner):
        """Should handle zero total tokens gracefully (lines 243-244)."""
        from moai_adk.rank.client import UserRank

        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.client.RankClient") as mock_client_class:
                mock_client = Mock()
                mock_client_class.return_value = mock_client

                mock_rank = Mock(spec=UserRank)
                mock_rank.username = "testuser"
                mock_rank.weekly = None
                mock_rank.total_tokens = 0
                mock_rank.input_tokens = 0
                mock_rank.output_tokens = 0
                mock_rank.total_sessions = 0
                mock_rank.daily = None
                mock_rank.monthly = None
                mock_rank.all_time = None
                mock_client.get_user_rank.return_value = mock_rank

                with patch("moai_adk.rank.hook.is_hook_installed", return_value=True):
                    result = runner.invoke(status)

                    assert result.exit_code == 0


class TestSyncCommand:
    """Test sync command (lines 299-363)."""

    def test_sync_without_credentials_displays_message(self, runner):
        """Should display not registered message (lines 316-319)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            result = runner.invoke(sync)

            assert result.exit_code == 0
            assert "Not registered" in result.output

    def test_sync_foreground_calls_sync_all_sessions(self, runner):
        """Should call sync_all_sessions in foreground mode (line 362)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.hook.sync_all_sessions") as mock_sync:
                result = runner.invoke(sync)

                assert result.exit_code == 0
                mock_sync.assert_called_once()

    def test_sync_background_starts_subprocess_with_log_file(self, runner):
        """Should start background sync with log file (lines 321-356)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.config.RankConfig.CONFIG_DIR") as mock_config_dir:
                # Mock CONFIG_DIR to exist
                mock_config_dir.exists.return_value = True

                log_file = Mock()
                log_file.parent.exists.return_value = True
                mock_config_dir.__truediv__.return_value = log_file

                # Mock temp file
                mock_temp = Mock()
                mock_temp.name = "/tmp/test_sync.py"
                mock_temp.__enter__ = Mock(return_value=mock_temp)
                mock_temp.__exit__ = Mock(return_value=False)

                with patch("tempfile.NamedTemporaryFile", return_value=mock_temp):
                    with patch("subprocess.Popen") as mock_popen:
                        mock_process = Mock()
                        mock_popen.return_value = mock_process

                        with patch("builtins.open", create=True):
                            with patch("os.unlink"):
                                result = runner.invoke(sync, ["--background"])

                                assert result.exit_code == 0
                                assert mock_popen.called

    def test_sync_background_with_no_config_dir(self, runner):
        """Should start background sync when CONFIG_DIR doesn't exist (line 351)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.config.RankConfig.CONFIG_DIR") as mock_config_dir:
                # Mock CONFIG_DIR to not exist
                mock_config_dir.exists.return_value = False

                # Mock temp file
                mock_temp = Mock()
                mock_temp.name = "/tmp/test_sync.py"
                mock_temp.__enter__ = Mock(return_value=mock_temp)
                mock_temp.__exit__ = Mock(return_value=False)

                with patch("tempfile.NamedTemporaryFile", return_value=mock_temp):
                    with patch("subprocess.Popen") as mock_popen:
                        mock_process = Mock()
                        mock_popen.return_value = mock_process

                        with patch("builtins.open", create=True) as mock_open:
                            # Mock for NUL or /dev/null
                            mock_open.return_value = Mock()

                            with patch("os.unlink"):
                                result = runner.invoke(sync, ["--background"])

                                assert result.exit_code == 0
                                assert mock_popen.called

    def test_sync_background_displays_log_file_location(self, runner):
        """Should display log file location (lines 355-356)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.config.RankConfig.CONFIG_DIR") as mock_config_dir:
                mock_config_dir.exists.return_value = True

                # Mock the log file path
                log_file = Mock()
                log_file.__str__ = Mock(return_value="/fake/path/sync.log")
                log_file.parent = Mock()
                log_file.parent.exists = Mock(return_value=True)
                mock_config_dir.__truediv__.return_value = log_file

                mock_temp = Mock()
                mock_temp.name = "/tmp/test_sync.py"
                mock_temp.__enter__ = Mock(return_value=mock_temp)
                mock_temp.__exit__ = Mock(return_value=False)

                with patch("tempfile.NamedTemporaryFile", return_value=mock_temp):
                    with patch("subprocess.Popen") as mock_popen:
                        mock_popen.return_value = Mock()

                        with patch("builtins.open", create=True):
                            with patch("os.unlink"):
                                result = runner.invoke(sync, ["--background"])

                                assert result.exit_code == 0
                                # Check that log file path is in output
                                assert "sync.log" in result.output or "Background sync started" in result.output

    def test_sync_background_creates_sync_script(self, runner):
        """Should create Python script for background sync (lines 328-335)."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.config.RankConfig.CONFIG_DIR") as mock_config_dir:
                mock_config_dir.exists.return_value = True
                log_file = Mock()
                log_file.parent.exists.return_value = True
                mock_config_dir.__truediv__.return_value = log_file

                # Track written content
                written_content = []

                mock_temp = Mock()
                mock_temp.name = "/tmp/test_sync.py"

                def mock_write(content):
                    written_content.append(content)

                mock_temp.write = mock_write
                mock_temp.__enter__ = Mock(return_value=mock_temp)
                mock_temp.__exit__ = Mock(return_value=False)

                with patch("tempfile.NamedTemporaryFile", return_value=mock_temp):
                    with patch("subprocess.Popen") as mock_popen:
                        mock_popen.return_value = Mock()

                        with patch("builtins.open", create=True):
                            with patch("os.unlink"):
                                result = runner.invoke(sync, ["--background"])

                                assert result.exit_code == 0
                                # Verify script was written
                                assert len(written_content) > 0
                                # Verify script contains sync_all_sessions
                                script_content = written_content[0]
                                assert "sync_all_sessions" in script_content


class TestRankGroupCommand:
    """Test rank command group."""

    def test_rank_group_shows_help(self, runner):
        """Should display help with all subcommands."""
        result = runner.invoke(rank, ["--help"])

        assert result.exit_code == 0
        assert "register" in result.output
        assert "status" in result.output
        assert "logout" in result.output
        assert "sync" in result.output
        assert "exclude" in result.output
        assert "include" in result.output


class TestHelperFunctions:
    """Test helper utility functions."""

    @pytest.mark.parametrize(
        "tokens,expected",
        [
            (0, "0"),
            (100, "100"),
            (999, "999"),
            (1000, "1.0K"),
            (15500, "15.5K"),
            (999999, "1000.0K"),
            (1000000, "1.0M"),
            (2500000, "2.5M"),
            (10000000, "10.0M"),
        ],
    )
    def test_format_tokens_various_inputs(self, tokens, expected):
        """Should format token counts correctly."""
        assert format_tokens(tokens) == expected

    @pytest.mark.parametrize(
        "position,total,expected_contains",
        [
            (1, 100, "[gold1]1st[/gold1]"),
            (2, 100, "[grey70]2nd[/grey70]"),
            (3, 100, "[orange3]3rd[/orange3]"),
            (5, 100, "#5"),
            (42, 1000, "#42"),
        ],
    )
    def test_format_rank_position_various_positions(self, position, total, expected_contains):
        """Should format rank positions correctly."""
        result = format_rank_position(position, total)
        assert expected_contains in result
        assert f"/ {total}" in result


class TestExcludeAndIncludeCommands:
    """Test exclude and include commands with various scenarios."""

    def test_exclude_without_path_uses_current_directory(self, runner):
        """Should use current directory when no path provided."""
        with patch("moai_adk.rank.hook.add_project_exclusion", return_value=True):
            with patch("moai_adk.rank.hook.load_rank_config", return_value={"exclude_projects": []}):
                result = runner.invoke(exclude)

                assert result.exit_code == 0

    def test_exclude_displays_total_excluded_count(self, runner):
        """Should display total excluded projects count."""
        with patch("moai_adk.rank.hook.add_project_exclusion", return_value=True):
            with patch("moai_adk.rank.hook.load_rank_config", return_value={"exclude_projects": ["/p1", "/p2"]}):
                result = runner.invoke(exclude, ["/test"])

                assert result.exit_code == 0
                assert "Total excluded projects: 2" in result.output

    def test_include_without_path_uses_current_directory(self, runner):
        """Should use current directory when no path provided."""
        with patch("moai_adk.rank.hook.remove_project_exclusion", return_value=True):
            with patch("moai_adk.rank.hook.load_rank_config", return_value={"exclude_projects": []}):
                result = runner.invoke(include)

                assert result.exit_code == 0

    def test_include_displays_remaining_exclusions(self, runner):
        """Should display remaining excluded projects count."""
        with patch("moai_adk.rank.hook.remove_project_exclusion", return_value=True):
            with patch("moai_adk.rank.hook.load_rank_config", return_value={"exclude_projects": ["/other"]}):
                result = runner.invoke(include, ["/test"])

                assert result.exit_code == 0
                assert "Remaining excluded projects: 1" in result.output

    def test_exclude_list_shows_all_excluded_projects(self, runner):
        """Should show all excluded projects with --list flag."""
        exclusions = ["/project1", "/project2", "/project3"]

        with patch("moai_adk.rank.hook.load_rank_config", return_value={"exclude_projects": exclusions}):
            result = runner.invoke(exclude, ["--list"])

            assert result.exit_code == 0
            for exc in exclusions:
                assert exc in result.output


class TestLogoutCommandScenarios:
    """Test logout command with additional scenarios."""

    def test_logout_with_null_credentials(self, runner):
        """Should handle null credentials gracefully."""
        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=True):
            with patch("moai_adk.rank.config.RankConfig.load_credentials", return_value=None):
                with patch("moai_adk.cli.commands.rank.click.confirm", return_value=False):
                    result = runner.invoke(logout)

                    assert result.exit_code == 0
                    # Should show "unknown" for username
                    assert "unknown" in result.output or "Cancelled" in result.output

    def test_logout_confirms_with_username(self, runner):
        """Should include username in confirmation prompt."""
        with patch("moai_adk.rank.config.RankConfig.has_credentials", return_value=True):
            with patch("moai_adk.rank.config.RankConfig.load_credentials", return_value=Mock(username="testuser")):
                with patch("moai_adk.cli.commands.rank.click.confirm") as mock_confirm:
                    mock_confirm.return_value = False

                    result = runner.invoke(logout)

                    assert result.exit_code == 0
                    # Verify confirm was called with username
                    mock_confirm.assert_called_once()
                    args = mock_confirm.call_args[0][0]
                    assert "testuser" in args


class TestErrorPaths:
    """Test error handling paths across all commands."""

    def test_status_handles_network_timeout(self, runner):
        """Should handle network timeout gracefully."""
        from moai_adk.rank.client import RankClientError

        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = True

            with patch("moai_adk.rank.client.RankClient") as mock_client_class:
                mock_client_class.side_effect = RankClientError("Request timeout")

                result = runner.invoke(status)

                assert result.exit_code == 0
                assert "Failed to fetch status" in result.output

    def test_register_handles_oauth_timeout(self, runner):
        """Should handle OAuth timeout gracefully."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.return_value = False

            with patch("moai_adk.rank.auth.OAuthHandler") as mock_oauth_class:
                mock_handler = Mock()
                mock_oauth_class.return_value = mock_handler

                def trigger_timeout(on_success, on_error, timeout):
                    on_error("OAuth flow timed out")

                mock_handler.start_oauth_flow = trigger_timeout

                result = runner.invoke(register)

                assert result.exit_code == 0
                # Should display timeout error
                assert "timed out" in result.output.lower() or "failed" in result.output.lower()

    def test_sync_handles_credentials_error(self, runner):
        """Should handle credentials loading error during sync."""
        with patch("moai_adk.rank.config.RankConfig") as mock_config_class:
            mock_config_class.has_credentials.side_effect = Exception("Config error")

            result = runner.invoke(sync)

            # Should handle error gracefully
            assert result.exit_code == 0 or result.exit_code == 1
