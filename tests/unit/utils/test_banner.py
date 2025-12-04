"""Unit tests for moai_adk.utils.banner module.

Tests for banner printing functionality.
"""

from unittest.mock import patch

import pytest

from moai_adk.utils.banner import (
    CLAUDE_TERRA_COTTA,
    MOAI_BANNER,
    print_banner,
    print_welcome_message,
)


class TestBannerConstants:
    """Test banner constants."""

    def test_claude_terra_cotta_color(self):
        """Test Claude terra cotta color is defined."""
        assert CLAUDE_TERRA_COTTA == "#DA7756"

    def test_moai_banner_exists(self):
        """Test MOAI banner is defined."""
        assert MOAI_BANNER is not None
        assert len(MOAI_BANNER) > 0
        assert "MoAI" in MOAI_BANNER or "███" in MOAI_BANNER


class TestPrintBanner:
    """Test print_banner function."""

    @patch("moai_adk.utils.banner.console")
    def test_print_banner_default_version(self, mock_console):
        """Test print_banner with default version."""
        print_banner()
        assert mock_console.print.called

    @patch("moai_adk.utils.banner.console")
    def test_print_banner_custom_version(self, mock_console):
        """Test print_banner with custom version."""
        print_banner(version="1.0.0")
        assert mock_console.print.called

    @patch("moai_adk.utils.banner.console")
    def test_print_banner_calls_multiple_times(self, mock_console):
        """Test print_banner calls console.print multiple times."""
        print_banner(version="0.1.0")
        # Should call print multiple times (banner + version + welcome)
        assert mock_console.print.call_count >= 2


class TestPrintWelcomeMessage:
    """Test print_welcome_message function."""

    @patch("moai_adk.utils.banner.console")
    def test_print_welcome_message(self, mock_console):
        """Test print_welcome_message."""
        print_welcome_message()
        assert mock_console.print.called

    @patch("moai_adk.utils.banner.console")
    def test_print_welcome_calls_multiple_times(self, mock_console):
        """Test print_welcome_message calls console.print multiple times."""
        print_welcome_message()
        # Should call print multiple times for different messages
        assert mock_console.print.call_count >= 2
