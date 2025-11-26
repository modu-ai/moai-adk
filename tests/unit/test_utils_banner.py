"""Unit tests for banner.py module

Tests for ASCII banner and welcome message printing.
"""

from unittest.mock import patch

from moai_adk.utils.banner import MOAI_BANNER, print_banner, print_welcome_message


class TestBannerConstants:
    """Test banner constant definitions"""

    def test_moai_banner_exists(self):
        """MOAI_BANNER constant should be defined"""
        assert MOAI_BANNER is not None
        assert isinstance(MOAI_BANNER, str)
        assert len(MOAI_BANNER) > 0

    def test_moai_banner_contains_ascii_art(self):
        """MOAI_BANNER should contain ASCII art characters"""
        assert "█" in MOAI_BANNER or "M" in MOAI_BANNER


class TestPrintBanner:
    """Test print_banner function"""

    def test_print_banner_default_version(self):
        """Should print banner with default version"""
        with patch("moai_adk.utils.banner.console.print") as mock_print:
            print_banner()
            # Should be called 3 times (banner, subtitle, version)
            assert mock_print.call_count == 3

    def test_print_banner_custom_version(self):
        """Should print banner with custom version"""
        with patch("moai_adk.utils.banner.console.print") as mock_print:
            print_banner(version="1.2.3")
            # Check if custom version is in the calls
            calls = [str(call) for call in mock_print.call_args_list]
            assert any("1.2.3" in str(call) for call in calls)

    def test_print_banner_includes_alfred(self):
        """Should mention Alfred SuperAgent in output"""
        with patch("moai_adk.utils.banner.console.print") as mock_print:
            print_banner()
            calls = [str(call) for call in mock_print.call_args_list]
            assert any("Alfred" in str(call) for call in calls)


class TestPrintWelcomeMessage:
    """Test print_welcome_message function"""

    def test_print_welcome_message_outputs_text(self):
        """Should print welcome message"""
        with patch("moai_adk.utils.banner.console.print") as mock_print:
            print_welcome_message()
            # Should be called at least once
            assert mock_print.call_count >= 1

    def test_print_welcome_message_mentions_initialization(self):
        """Should mention project initialization"""
        with patch("moai_adk.utils.banner.console.print") as mock_print:
            print_welcome_message()
            calls = [str(call) for call in mock_print.call_args_list]
            # Check for keywords
            all_calls_str = " ".join(str(call) for call in calls)
            assert "Initialization" in all_calls_str or "initialization" in all_calls_str or "Welcome" in all_calls_str

    def test_print_welcome_message_mentions_cancel(self):
        """Should inform user about cancellation option"""
        with patch("moai_adk.utils.banner.console.print") as mock_print:
            print_welcome_message()
            calls = [str(call) for call in mock_print.call_args_list]
            all_calls_str = " ".join(str(call) for call in calls)
            assert "Ctrl+C" in all_calls_str or "cancel" in all_calls_str.lower()
