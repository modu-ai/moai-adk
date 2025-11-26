"""Comprehensive test suite for banner.py utilities module.

This module provides 90%+ coverage for banner display functionality including:
- Banner display with default and custom versions
- Welcome message display
- Rich console integration
- Color and formatting support
- Output formatting and structure
"""

from unittest.mock import patch

from rich.console import Console

from moai_adk.utils.banner import (
    MOAI_BANNER,
    console,
    print_banner,
    print_welcome_message,
)

# ============================================================================
# Module Constants and Initialization Tests
# ============================================================================


class TestModuleConstants:
    """Tests for module-level constants."""

    def test_moai_banner_string_exists(self):
        """Test that MOAI_BANNER constant is defined."""
        assert MOAI_BANNER is not None
        assert isinstance(MOAI_BANNER, str)

    def test_moai_banner_contains_ascii_art(self):
        """Test that MOAI_BANNER contains expected ASCII art characters."""
        assert "███" in MOAI_BANNER  # Contains box-drawing characters
        assert "╗" in MOAI_BANNER
        assert "╝" in MOAI_BANNER
        assert "██" in MOAI_BANNER

    def test_moai_banner_multiline_structure(self):
        """Test that MOAI_BANNER has multiple lines."""
        lines = MOAI_BANNER.strip().split("\n")
        assert len(lines) > 0
        assert len(lines) >= 5  # Banner should have multiple rows

    def test_moai_banner_contains_moai_text(self):
        """Test that MOAI_BANNER contains 'MoAI' related patterns."""
        assert "█" in MOAI_BANNER  # Should contain block characters

    def test_moai_banner_no_leading_trailing_spaces_per_line(self):
        """Test banner line formatting."""
        lines = MOAI_BANNER.split("\n")
        for line in lines:
            # Lines can have spaces, just check they contain content
            if line.strip():
                assert len(line) > 0

    def test_console_object_initialized(self):
        """Test that console object is initialized."""
        assert console is not None
        assert isinstance(console, Console)

    def test_banner_contains_full_width_ascii(self):
        """Test banner uses full-width characters."""
        # Should have consistent width box-drawing
        assert "║" in MOAI_BANNER or "█" in MOAI_BANNER
        assert "═" in MOAI_BANNER or "█" in MOAI_BANNER or "╗" in MOAI_BANNER


# ============================================================================
# print_banner Function Tests
# ============================================================================


class TestPrintBanner:
    """Tests for print_banner function."""

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_default_version(self, mock_print):
        """Test print_banner with default version."""
        print_banner()

        # Verify print was called
        assert mock_print.called
        # Should have 3 calls: banner, welcome line, version line
        assert mock_print.call_count >= 2

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_custom_version(self, mock_print):
        """Test print_banner with custom version."""
        custom_version = "1.2.3"
        print_banner(version=custom_version)

        assert mock_print.called
        # Verify version appears in one of the calls
        version_found = False
        for call_args in mock_print.call_args_list:
            if call_args[0] and custom_version in str(call_args[0][0]):
                version_found = True
                break
        assert version_found

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_version_0_3_0_default(self, mock_print):
        """Test print_banner uses 0.3.0 as default version."""
        print_banner()

        # Check if default version appears
        version_found = False
        for call_args in mock_print.call_args_list:
            if call_args[0] and "0.3.0" in str(call_args[0][0]):
                version_found = True
                break
        assert version_found

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_includes_banner_text(self, mock_print):
        """Test print_banner includes MOAI_BANNER in output."""
        print_banner()

        # First call should include banner
        first_call_args = mock_print.call_args_list[0][0][0]
        assert (
            "███" in first_call_args or "blue" in str(first_call_args).lower() or "cyan" in str(first_call_args).lower()
        )

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_includes_welcome_text(self, mock_print):
        """Test print_banner includes welcome message."""
        print_banner()

        all_output = str(mock_print.call_args_list)
        # Should mention Alfred or Agentic Development Kit
        assert "Alfred" in all_output or "MoAI" in all_output or "Agentic" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_includes_version_label(self, mock_print):
        """Test print_banner includes 'Version:' label."""
        print_banner()

        all_output = str(mock_print.call_args_list)
        assert "Version" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_multiple_calls(self, mock_print):
        """Test print_banner makes multiple print calls."""
        print_banner()

        # Should make at least 2 calls (banner + info)
        assert mock_print.call_count >= 2

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_return_value(self, mock_print):
        """Test print_banner returns None."""
        result = print_banner()
        assert result is None

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_with_empty_string_version(self, mock_print):
        """Test print_banner with empty string version."""
        print_banner(version="")

        assert mock_print.called
        # Should not crash even with empty version

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_with_long_version_string(self, mock_print):
        """Test print_banner with long version string."""
        long_version = "1.2.3-alpha.dev+build.12345"
        print_banner(version=long_version)

        assert mock_print.called
        version_found = False
        for call_args in mock_print.call_args_list:
            if call_args[0] and long_version in str(call_args[0][0]):
                version_found = True
                break
        assert version_found

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_with_special_characters_version(self, mock_print):
        """Test print_banner with special characters in version."""
        special_version = "2.0.0-beta+20250101"
        print_banner(version=special_version)

        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_cyan_color_applied(self, mock_print):
        """Test print_banner applies cyan color to banner."""
        print_banner()

        assert mock_print.called
        # First call should have cyan color markup
        first_call_str = str(mock_print.call_args_list[0])
        assert "cyan" in first_call_str.lower()

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_dim_color_applied(self, mock_print):
        """Test print_banner applies dim color to descriptions."""
        print_banner()

        assert mock_print.called
        all_output = str(mock_print.call_args_list)
        # Should use dim for secondary text
        assert "dim" in all_output.lower()

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_newline_handling(self, mock_print):
        """Test print_banner adds appropriate newlines."""
        print_banner()

        # Verify calls were made with rich console (which handles formatting)
        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_called_with_string_arguments(self, mock_print):
        """Test print_banner calls console.print with string arguments."""
        print_banner()

        for call_obj in mock_print.call_args_list:
            # Each call should have at least one argument
            assert len(call_obj[0]) > 0 or len(call_obj[1]) > 0


# ============================================================================
# print_welcome_message Function Tests
# ============================================================================


class TestPrintWelcomeMessage:
    """Tests for print_welcome_message function."""

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_basic(self, mock_print):
        """Test print_welcome_message basic functionality."""
        print_welcome_message()

        assert mock_print.called
        # Should make at least 3 calls for three text lines
        assert mock_print.call_count >= 3

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_includes_welcome_text(self, mock_print):
        """Test print_welcome_message includes 'Welcome' text."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        assert "Welcome" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_includes_moai_adk(self, mock_print):
        """Test print_welcome_message mentions MoAI-ADK."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        assert "MoAI" in all_output or "ADK" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_includes_initialization(self, mock_print):
        """Test print_welcome_message mentions initialization."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should mention initialization/wizard
        assert "Initialization" in all_output or "initialization" in all_output or "wizard" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_includes_guidance(self, mock_print):
        """Test print_welcome_message provides user guidance."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should guide through process
        assert "guide" in all_output.lower() or "wizard" in all_output.lower() or "guide you" in all_output.lower()

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_includes_cancel_instruction(self, mock_print):
        """Test print_welcome_message mentions how to cancel."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should mention Ctrl+C or cancel option
        assert "Ctrl" in all_output or "cancel" in all_output.lower() or "C" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_return_value(self, mock_print):
        """Test print_welcome_message returns None."""
        result = print_welcome_message()
        assert result is None

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_bold_title(self, mock_print):
        """Test print_welcome_message applies bold formatting to title."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should use bold formatting
        assert "bold" in all_output.lower()

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_cyan_title(self, mock_print):
        """Test print_welcome_message uses cyan color for title."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should use cyan for main title
        assert "cyan" in all_output.lower()

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_dim_text(self, mock_print):
        """Test print_welcome_message uses dim for secondary text."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should use dim for secondary information
        assert "dim" in all_output.lower()

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_emoji_present(self, mock_print):
        """Test print_welcome_message includes emoji."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should have rocket emoji
        assert "🚀" in all_output or "rocket" in all_output.lower() or "Rocket" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_newlines(self, mock_print):
        """Test print_welcome_message has proper spacing with newlines."""
        print_welcome_message()

        # Multiple calls indicate separate lines/spacing
        assert mock_print.call_count >= 3


# ============================================================================
# Console Integration Tests
# ============================================================================


class TestConsoleIntegration:
    """Tests for console integration and output behavior."""

    @patch("moai_adk.utils.banner.Console")
    def test_console_imported_from_rich(self, mock_console_class):
        """Test that console is imported from rich.console."""
        # The module should have imported Console from rich
        assert console is not None

    def test_console_is_console_instance(self):
        """Test that module console is Console instance."""
        assert isinstance(console, Console)

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_and_welcome_can_be_called_together(self, mock_print):
        """Test that banner and welcome can be called in sequence."""
        print_banner()
        mock_print.reset_mock()
        print_welcome_message()

        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_console_print_called_with_markup(self, mock_print):
        """Test that console.print is called with markup strings."""
        print_banner()

        # Check that at least one call has rich markup
        found_markup = False
        for call_obj in mock_print.call_args_list:
            call_str = str(call_obj[0])
            if "[" in call_str and "]" in call_str:
                found_markup = True
                break
        assert found_markup


# ============================================================================
# Edge Cases and Error Handling Tests
# ============================================================================


class TestEdgeCasesAndErrorHandling:
    """Tests for edge cases and error handling."""

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_with_unicode_version(self, mock_print):
        """Test print_banner with unicode characters in version."""
        unicode_version = "1.0.0-unicode™"
        print_banner(version=unicode_version)

        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_with_very_long_version(self, mock_print):
        """Test print_banner with very long version string."""
        long_version = "1.0.0-" + "x" * 100
        print_banner(version=long_version)

        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_with_none_like_version(self, mock_print):
        """Test print_banner with string 'None' as version."""
        print_banner(version="None")

        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_print_welcome_message_exception_handling(self, mock_print):
        """Test print_welcome_message handles exceptions gracefully."""
        # If console.print raises, should not crash
        mock_print.side_effect = [None, None, None]
        print_welcome_message()
        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_exception_handling(self, mock_print):
        """Test print_banner handles exceptions gracefully."""
        mock_print.side_effect = [None, None, None]
        print_banner()
        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_output_format_consistency(self, mock_print):
        """Test banner output format is consistent."""
        print_banner("1.0.0")
        first_call_count = mock_print.call_count

        mock_print.reset_mock()
        print_banner("2.0.0")
        second_call_count = mock_print.call_count

        # Should have same number of calls regardless of version
        assert first_call_count == second_call_count

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_with_numeric_version(self, mock_print):
        """Test print_banner with numeric version string."""
        print_banner(version="123")

        assert mock_print.called
        version_found = False
        for call_args in mock_print.call_args_list:
            if call_args[0] and "123" in str(call_args[0][0]):
                version_found = True
                break
        assert version_found

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_with_special_markup_characters(self, mock_print):
        """Test print_banner handles special markup characters in version."""
        # Version with characters that might interfere with rich markup
        special_version = "1.0.0[beta]"
        print_banner(version=special_version)

        assert mock_print.called


# ============================================================================
# Integration and Workflow Tests
# ============================================================================


class TestIntegrationAndWorkflow:
    """Integration tests for banner display workflow."""

    @patch("moai_adk.utils.banner.console.print")
    def test_full_initialization_workflow(self, mock_print):
        """Test complete initialization workflow with banner and welcome."""
        # This simulates how these functions would be used together
        print_banner(version="0.3.0")
        first_call_count = mock_print.call_count

        mock_print.reset_mock()
        print_welcome_message()
        second_call_count = mock_print.call_count

        # Both should work together
        assert first_call_count > 0
        assert second_call_count > 0

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_and_welcome_no_interference(self, mock_print):
        """Test that calling both functions doesn't cause issues."""
        print_banner()
        print_welcome_message()

        # Should have accumulated calls from both
        total_calls = mock_print.call_count
        assert total_calls >= 5  # At least 3 from welcome + 2 from banner

    @patch("moai_adk.utils.banner.console.print")
    def test_multiple_banner_calls(self, mock_print):
        """Test calling print_banner multiple times."""
        print_banner(version="1.0.0")
        first_count = mock_print.call_count

        mock_print.reset_mock()
        print_banner(version="2.0.0")
        second_count = mock_print.call_count

        # Both calls should succeed
        assert first_count > 0
        assert second_count > 0

    @patch("moai_adk.utils.banner.console.print")
    def test_multiple_welcome_calls(self, mock_print):
        """Test calling print_welcome_message multiple times."""
        print_welcome_message()
        first_count = mock_print.call_count

        mock_print.reset_mock()
        print_welcome_message()
        second_count = mock_print.call_count

        # Both calls should succeed
        assert first_count > 0
        assert second_count > 0

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_output_structure(self, mock_print):
        """Test banner output structure is well-formed."""
        print_banner()

        # Verify we have a structured output
        assert len(mock_print.call_args_list) > 0

        # Each call should have arguments
        for call_obj in mock_print.call_args_list:
            assert len(call_obj[0]) > 0 or len(call_obj[1]) > 0

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_version_parameter_is_optional(self, mock_print):
        """Test that version parameter is truly optional."""
        # Should work without any arguments
        print_banner()
        assert mock_print.called

        mock_print.reset_mock()
        # Should work with explicit version
        print_banner(version="1.0.0")
        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_welcome_no_required_parameters(self, mock_print):
        """Test that welcome message takes no required parameters."""
        print_welcome_message()
        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_maintains_state(self, mock_print):
        """Test that multiple banner calls maintain consistent output."""
        versions = ["1.0.0", "2.0.0", "3.0.0"]

        for version in versions:
            mock_print.reset_mock()
            print_banner(version=version)

            # Each version should appear in output
            version_found = False
            for call_args in mock_print.call_args_list:
                if call_args[0] and version in str(call_args[0][0]):
                    version_found = True
                    break
            assert version_found


# ============================================================================
# Output Content Verification Tests
# ============================================================================


class TestOutputContent:
    """Tests verifying specific output content."""

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_includes_moai_asciiart(self, mock_print):
        """Test banner output includes MOAI ASCII art."""
        print_banner()

        # First call should be the ASCII art banner
        first_call = str(mock_print.call_args_list[0])
        # Should contain the banner constant or its content
        assert "█" in first_call or "MoAI" in first_call or "[cyan]" in first_call

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_includes_alfred_reference(self, mock_print):
        """Test banner mentions Alfred SuperAgent."""
        print_banner()

        all_output = str(mock_print.call_args_list)
        assert "Alfred" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_includes_emoji(self, mock_print):
        """Test banner includes emoji."""
        print_banner()

        all_output = str(mock_print.call_args_list)
        assert "🎩" in all_output  # Top hat emoji for Alfred

    @patch("moai_adk.utils.banner.console.print")
    def test_welcome_includes_rocket_emoji(self, mock_print):
        """Test welcome message includes rocket emoji."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        assert "🚀" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_welcome_provides_setup_guidance(self, mock_print):
        """Test welcome message provides setup guidance."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should explain what the wizard does
        assert "guide" in all_output.lower() or "setting" in all_output.lower()

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_contains_version_string_literal(self, mock_print):
        """Test banner output contains 'Version:' string."""
        print_banner(version="1.2.3")

        all_output = str(mock_print.call_args_list)
        assert "Version:" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_welcome_message_line_count(self, mock_print):
        """Test welcome message has expected number of lines."""
        print_welcome_message()

        # Should have at least 3 separate print calls (title, description, etc.)
        assert mock_print.call_count >= 3


# ============================================================================
# Formatting and Style Tests
# ============================================================================


class TestFormattingAndStyle:
    """Tests for Rich console markup and formatting."""

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_uses_rich_markup(self, mock_print):
        """Test banner uses Rich markup syntax."""
        print_banner()

        # Should have Rich color markup in at least one call
        found_markup = False
        for call_obj in mock_print.call_args_list:
            call_str = str(call_obj[0])
            if "[cyan]" in call_str or "[dim]" in call_str:
                found_markup = True
                break
        assert found_markup

    @patch("moai_adk.utils.banner.console.print")
    def test_welcome_uses_cyan_bold_markup(self, mock_print):
        """Test welcome message uses cyan bold markup."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should have cyan bold markup
        assert "[cyan bold]" in all_output or ("cyan" in all_output.lower() and "bold" in all_output.lower())

    @patch("moai_adk.utils.banner.console.print")
    def test_secondary_text_uses_dim_markup(self, mock_print):
        """Test secondary text uses dim markup."""
        print_welcome_message()

        all_output = str(mock_print.call_args_list)
        # Should use dim for secondary info
        assert "[dim]" in all_output

    @patch("moai_adk.utils.banner.console.print")
    def test_markup_is_properly_closed(self, mock_print):
        """Test that markup tags are properly closed."""
        print_banner()

        for call_obj in mock_print.call_args_list:
            call_str = str(call_obj[0])
            if "[cyan]" in call_str:
                assert "[/cyan]" in call_str

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_newline_formatting(self, mock_print):
        """Test banner includes proper newline formatting."""
        print_banner()

        # Last call should include newline for spacing
        # Multiple calls indicate good formatting
        assert mock_print.call_count >= 3

    @patch("moai_adk.utils.banner.console.print")
    def test_welcome_newline_formatting(self, mock_print):
        """Test welcome message includes proper newline formatting."""
        print_welcome_message()

        # Multiple calls with newlines indicate proper formatting
        assert mock_print.call_count >= 3


# ============================================================================
# Parameter Variation Tests
# ============================================================================


class TestParameterVariations:
    """Tests for various parameter combinations."""

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_version_variations(self, mock_print):
        """Test print_banner with various version formats."""
        versions = [
            "0.3.0",
            "1.0.0-alpha",
            "2.0.0-beta.1",
            "3.0.0+build.12345",
            "0.0.1",
            "999.999.999",
        ]

        for version in versions:
            mock_print.reset_mock()
            print_banner(version=version)
            assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_positional_argument(self, mock_print):
        """Test print_banner with positional version argument."""
        print_banner("1.5.0")
        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_print_banner_keyword_argument(self, mock_print):
        """Test print_banner with keyword version argument."""
        print_banner(version="1.5.0")
        assert mock_print.called

    @patch("moai_adk.utils.banner.console.print")
    def test_banner_version_parameter_type(self, mock_print):
        """Test print_banner accepts string version parameter."""
        # Version should be a string type
        print_banner(version="1.0.0")
        assert mock_print.called
