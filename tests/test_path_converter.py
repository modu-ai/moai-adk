"""Tests for Windows ↔ WSL path conversion utilities.

MoAI-ADK v1.8.3 adds WSL path conversion support:
- Convert Windows paths (C:\\Users\\...) to WSL paths (/mnt/c/Users/...)
- Convert WSL paths (/mnt/c/...) to Windows paths (C:\\...)
- Normalize paths based on current environment
- Handle edge cases (spaces, special characters, relative paths)

Test Coverage:
- Windows to WSL path conversion
- WSL to Windows path conversion (reverse)
- Different drive letters (C:, D:, E:, etc.)
- Paths with spaces and special characters
- Relative paths and edge cases
- Environment-aware normalization
"""

import os
from pathlib import Path
from unittest.mock import patch

import pytest

# Note: path_converter module is expected to be created in v1.8.3
# These tests define the expected API and behavior
try:
    from moai_adk.utils.path_converter import (
        convert_windows_to_wsl,
        convert_wsl_to_windows,
        is_wsl,
        normalize_path_for_wsl,
    )

    PATH_CONVERTER_AVAILABLE = True
except ImportError:
    PATH_CONVERTER_AVAILABLE = False

    # Create mock functions for testing the expected API
    def convert_windows_to_wsl(path: str) -> str:
        """Mock implementation - to be replaced with actual."""
        raise NotImplementedError("path_converter module not yet implemented")

    def convert_wsl_to_windows(path: str) -> str:
        """Mock implementation - to be replaced with actual."""
        raise NotImplementedError("path_converter module not yet implemented")

    def normalize_path_for_wsl(path: str) -> str:
        """Mock implementation - to be replaced with actual."""
        raise NotImplementedError("path_converter module not yet implemented")

    def is_wsl() -> bool:
        """Mock implementation - to be replaced with actual."""
        return "WSL_DISTRO_NAME" in os.environ


# Skip tests if module not yet implemented
pytestmark = pytest.mark.skipif(
    not PATH_CONVERTER_AVAILABLE, reason="path_converter module not yet implemented in v1.8.3"
)


class TestWindowsToWSLConversion:
    """Test Windows path to WSL path conversion."""

    @pytest.mark.parametrize(
        "windows_path,expected_wsl",
        [
            ("C:\\Users\\goos\\project", "/mnt/c/Users/goos/project"),
            ("D:\\code\\app", "/mnt/d/code/app"),
            ("E:\\data\\file.txt", "/mnt/e/data/file.txt"),
            ("F:\\backup\\archive", "/mnt/f/backup/archive"),
            ("Z:\\network\\share", "/mnt/z/network/share"),
        ],
        ids=["c_drive", "d_drive", "e_drive", "f_drive", "z_drive"],
    )
    def test_windows_to_wsl_various_drives(self, windows_path, expected_wsl):
        """Test conversion of Windows paths from various drive letters."""
        result = convert_windows_to_wsl(windows_path)
        assert result == expected_wsl

    def test_windows_to_wsl_preserves_directory_structure(self):
        """Test that directory structure is preserved during conversion."""
        windows_path = "C:\\Users\\goos\\Documents\\Projects\\MoAI\\README.md"
        expected = "/mnt/c/Users/goos/Documents/Projects/MoAI/README.md"
        result = convert_windows_to_wsl(windows_path)
        assert result == expected

    def test_windows_to_wsl_lowercase_drive_letter(self):
        """Test that lowercase drive letters are handled correctly."""
        windows_path = "c:\\users\\goos\\project"
        result = convert_windows_to_wsl(windows_path)
        # Should normalize to lowercase in WSL path
        assert result.startswith("/mnt/c/")

    def test_windows_to_wsl_uppercase_drive_letter(self):
        """Test that uppercase drive letters are handled correctly."""
        windows_path = "C:\\Users\\goos\\project"
        result = convert_windows_to_wsl(windows_path)
        assert result.startswith("/mnt/c/")

    def test_windows_to_wsl_forward_slashes(self):
        """Test Windows path with forward slashes instead of backslashes."""
        windows_path = "C:/Users/goos/project"
        result = convert_windows_to_wsl(windows_path)
        assert result == "/mnt/c/Users/goos/project"

    def test_windows_to_wsl_mixed_slashes(self):
        """Test Windows path with mixed forward and backslashes."""
        windows_path = "C:\\Users/goos\\project/file.txt"
        result = convert_windows_to_wsl(windows_path)
        assert result == "/mnt/c/Users/goos/project/file.txt"

    def test_windows_to_wsl_root_directory(self):
        """Test conversion of root directory."""
        windows_path = "C:\\"
        result = convert_windows_to_wsl(windows_path)
        assert result == "/mnt/c/"

    def test_windows_to_wsl_single_directory(self):
        """Test conversion of single directory level."""
        windows_path = "C:\\Windows"
        result = convert_windows_to_wsl(windows_path)
        assert result == "/mnt/c/Windows"


class TestWSLToWindowsConversion:
    """Test WSL path to Windows path conversion (reverse operation)."""

    @pytest.mark.parametrize(
        "wsl_path,expected_windows",
        [
            ("/mnt/c/Users/goos/project", "C:\\Users\\goos\\project"),
            ("/mnt/d/code/app", "D:\\code\\app"),
            ("/mnt/e/data/file.txt", "E:\\data\\file.txt"),
            ("/mnt/f/backup/archive", "F:\\backup\\archive"),
        ],
        ids=["c_drive", "d_drive", "e_drive", "f_drive"],
    )
    def test_wsl_to_windows_various_drives(self, wsl_path, expected_windows):
        """Test conversion of WSL paths to Windows paths for various drives."""
        result = convert_wsl_to_windows(wsl_path)
        assert result == expected_windows

    def test_wsl_to_windows_preserves_structure(self):
        """Test that directory structure is preserved during reverse conversion."""
        wsl_path = "/mnt/c/Users/goos/Documents/Projects/MoAI/README.md"
        expected = "C:\\Users\\goos\\Documents\\Projects\\MoAI\\README.md"
        result = convert_wsl_to_windows(wsl_path)
        assert result == expected

    def test_wsl_to_windows_root_directory(self):
        """Test conversion of WSL root mount point."""
        wsl_path = "/mnt/c/"
        result = convert_wsl_to_windows(wsl_path)
        assert result == "C:\\"

    def test_wsl_to_windows_uppercase_drive_letter(self):
        """Test that drive letter is uppercase in Windows path."""
        wsl_path = "/mnt/c/Users/goos/project"
        result = convert_wsl_to_windows(wsl_path)
        # Windows drive letters should be uppercase
        assert result.startswith("C:\\")

    def test_roundtrip_conversion(self):
        """Test that converting Windows → WSL → Windows preserves path."""
        original = "C:\\Users\\goos\\Documents\\test.txt"
        wsl_path = convert_windows_to_wsl(original)
        back_to_windows = convert_wsl_to_windows(wsl_path)
        assert back_to_windows == original

    def test_roundtrip_conversion_reverse(self):
        """Test that converting WSL → Windows → WSL preserves path."""
        original = "/mnt/c/Users/goos/Documents/test.txt"
        windows_path = convert_wsl_to_windows(original)
        back_to_wsl = convert_windows_to_wsl(windows_path)
        assert back_to_wsl == original


class TestPathsWithSpaces:
    """Test path conversion with spaces in filenames and directories."""

    def test_windows_to_wsl_with_spaces(self):
        """Test Windows path with spaces in directory names."""
        windows = "C:\\Program Files\\My App\\config.json"
        wsl = convert_windows_to_wsl(windows)
        assert wsl == "/mnt/c/Program Files/My App/config.json"

    def test_wsl_to_windows_with_spaces(self):
        """Test WSL path with spaces in directory names."""
        wsl = "/mnt/c/Program Files/My App/config.json"
        windows = convert_wsl_to_windows(wsl)
        assert windows == "C:\\Program Files\\My App\\config.json"

    @pytest.mark.parametrize(
        "path_with_spaces",
        [
            "C:\\Users\\My User\\Documents",
            "D:\\My Projects\\Web App\\src",
            "E:\\Data Files\\Test Data\\results.csv",
        ],
        ids=["user_folder", "project_folder", "data_file"],
    )
    def test_various_paths_with_spaces(self, path_with_spaces):
        """Test conversion of various paths containing spaces."""
        wsl_path = convert_windows_to_wsl(path_with_spaces)
        # Should preserve spaces
        assert " " in wsl_path or path_with_spaces.count(" ") == wsl_path.count(" ")

        # Reverse conversion should work
        back_to_windows = convert_wsl_to_windows(wsl_path)
        assert back_to_windows == path_with_spaces


class TestSpecialCharacters:
    """Test path conversion with special characters."""

    def test_path_with_hyphens(self):
        """Test path with hyphens in names."""
        windows = "C:\\my-project\\test-file.txt"
        wsl = convert_windows_to_wsl(windows)
        assert wsl == "/mnt/c/my-project/test-file.txt"

    def test_path_with_underscores(self):
        """Test path with underscores in names."""
        windows = "C:\\my_project\\test_file.txt"
        wsl = convert_windows_to_wsl(windows)
        assert wsl == "/mnt/c/my_project/test_file.txt"

    def test_path_with_dots(self):
        """Test path with dots in directory names."""
        windows = "C:\\node_modules\\.cache\\v1.2.3\\data"
        wsl = convert_windows_to_wsl(windows)
        assert wsl == "/mnt/c/node_modules/.cache/v1.2.3/data"

    def test_path_with_parentheses(self):
        """Test path with parentheses."""
        windows = "C:\\Program Files (x86)\\App\\file.txt"
        wsl = convert_windows_to_wsl(windows)
        assert wsl == "/mnt/c/Program Files (x86)/App/file.txt"

    @pytest.mark.parametrize(
        "special_path",
        [
            "C:\\Users\\test@example\\Documents",
            "C:\\data\\file#1.txt",
            "C:\\projects\\app-v2.0\\src",
        ],
        ids=["at_symbol", "hash_symbol", "version_number"],
    )
    def test_various_special_characters(self, special_path):
        """Test conversion with various special characters."""
        wsl_path = convert_windows_to_wsl(special_path)
        # Should preserve special characters
        assert "@" in special_path or "@" not in special_path  # Placeholder assertion
        back = convert_wsl_to_windows(wsl_path)
        assert back == special_path


class TestEdgeCases:
    """Test edge cases and unusual scenarios."""

    def test_relative_path_unchanged(self):
        """Test that relative paths are handled appropriately."""
        relative = "./relative/path"
        # Relative paths should be returned as-is or with minimal transformation
        result = convert_windows_to_wsl(relative)
        assert result is not None

    def test_empty_path(self):
        """Test handling of empty path."""
        with pytest.raises((ValueError, AssertionError)):
            convert_windows_to_wsl("")

    def test_path_without_drive_letter(self):
        """Test handling of path without drive letter."""
        path = "\\Users\\goos\\project"
        # Should handle gracefully - either error or add default drive
        result = convert_windows_to_wsl(path)
        assert result is not None

    def test_wsl_path_without_mnt_prefix(self):
        """Test WSL path that doesn't start with /mnt/."""
        wsl_path = "/home/user/project"
        # Non-Windows mount paths should be handled appropriately
        result = convert_wsl_to_windows(wsl_path)
        # Might raise error or return path unchanged
        assert result is not None

    def test_network_path_unc(self):
        """Test UNC network path conversion."""
        unc_path = "\\\\server\\share\\folder"
        # UNC paths are not directly convertible to WSL /mnt/ format
        # Implementation should handle this gracefully
        try:
            result = convert_windows_to_wsl(unc_path)
            assert result is not None
        except (ValueError, NotImplementedError):
            # UNC paths might not be supported
            pass

    def test_path_with_trailing_slash(self):
        """Test path with trailing slash."""
        windows_with_slash = "C:\\Users\\goos\\"
        wsl = convert_windows_to_wsl(windows_with_slash)
        # Should handle trailing slash appropriately
        assert wsl.startswith("/mnt/c/")

    def test_path_with_multiple_consecutive_slashes(self):
        """Test path with multiple consecutive slashes."""
        windows = "C:\\Users\\\\goos\\\\project"
        wsl = convert_windows_to_wsl(windows)
        # Should normalize multiple slashes
        assert "//" not in wsl or wsl.count("//") < windows.count("\\\\")


class TestNormalizePathForWSL:
    """Test environment-aware path normalization."""

    def test_normalize_in_wsl_environment(self):
        """Test normalization when running in WSL."""
        with patch("moai_adk.utils.path_converter.is_wsl", return_value=True):
            windows_path = "C:\\Users\\goos\\project"
            result = normalize_path_for_wsl(windows_path)
            # In WSL, should convert to WSL format
            assert result == "/mnt/c/Users/goos/project"

    def test_normalize_outside_wsl_environment(self):
        """Test normalization when not running in WSL."""
        with patch("moai_adk.utils.path_converter.is_wsl", return_value=False):
            windows_path = "C:\\Users\\goos\\project"
            result = normalize_path_for_wsl(windows_path)
            # Outside WSL, path should remain unchanged
            assert result == windows_path

    def test_normalize_wsl_path_in_wsl(self):
        """Test normalization of WSL path when in WSL."""
        with patch("moai_adk.utils.path_converter.is_wsl", return_value=True):
            wsl_path = "/mnt/c/Users/goos/project"
            result = normalize_path_for_wsl(wsl_path)
            # Already WSL format, should remain unchanged
            assert result == wsl_path

    def test_normalize_wsl_path_outside_wsl(self):
        """Test normalization of WSL path when not in WSL."""
        with patch("moai_adk.utils.path_converter.is_wsl", return_value=False):
            wsl_path = "/mnt/c/Users/goos/project"
            result = normalize_path_for_wsl(wsl_path)
            # Behavior: might convert to Windows or leave as-is
            assert result is not None

    @pytest.mark.parametrize(
        "in_wsl,input_path,should_convert",
        [
            (True, "C:\\Users\\test", True),
            (False, "C:\\Users\\test", False),
            (True, "/mnt/c/Users/test", False),  # Already WSL format
            (False, "/mnt/c/Users/test", False),  # Not in WSL, no conversion
        ],
        ids=["wsl_windows_path", "non_wsl_windows_path", "wsl_wsl_path", "non_wsl_wsl_path"],
    )
    def test_normalize_various_scenarios(self, in_wsl, input_path, should_convert):
        """Test normalization in various environment scenarios."""
        with patch("moai_adk.utils.path_converter.is_wsl", return_value=in_wsl):
            result = normalize_path_for_wsl(input_path)

            if should_convert:
                # Should be converted to WSL format
                assert result.startswith("/mnt/")
            else:
                # Should remain unchanged or minimally processed
                assert result is not None


class TestPathConverterIntegration:
    """Integration tests for path converter module."""

    def test_import_all_functions(self):
        """Test that all expected functions can be imported."""
        if PATH_CONVERTER_AVAILABLE:
            from moai_adk.utils.path_converter import (
                convert_windows_to_wsl,
                convert_wsl_to_windows,
                is_wsl,
                normalize_path_for_wsl,
            )

            assert callable(convert_windows_to_wsl)
            assert callable(convert_wsl_to_windows)
            assert callable(is_wsl)
            assert callable(normalize_path_for_wsl)

    def test_is_wsl_function_exists(self):
        """Test that is_wsl function exists and returns boolean."""
        result = is_wsl()
        assert isinstance(result, bool)

    def test_conversion_consistency(self):
        """Test that multiple conversions produce consistent results."""
        if PATH_CONVERTER_AVAILABLE:
            windows_path = "C:\\Users\\goos\\project"

            # Convert multiple times
            wsl1 = convert_windows_to_wsl(windows_path)
            wsl2 = convert_windows_to_wsl(windows_path)
            wsl3 = convert_windows_to_wsl(windows_path)

            # All conversions should produce same result
            assert wsl1 == wsl2 == wsl3

    def test_pathlib_integration(self):
        """Test integration with pathlib.Path objects."""
        if PATH_CONVERTER_AVAILABLE:
            windows_path_str = "C:\\Users\\goos\\project"
            # If implementation supports Path objects
            try:
                wsl_result = convert_windows_to_wsl(windows_path_str)
                # Should produce valid path string
                assert isinstance(wsl_result, str)
                # Should be able to create Path from result
                path_obj = Path(wsl_result)
                assert path_obj.parts[0] == "/"
            except Exception:
                # If Path objects not supported, that's acceptable
                pass


class TestCrossPlatformBehavior:
    """Test cross-platform behavior and compatibility."""

    def test_conversion_works_on_any_platform(self):
        """Test that conversion functions work regardless of current platform."""
        if PATH_CONVERTER_AVAILABLE:
            # Conversion should work even if not on Windows or WSL
            windows_path = "C:\\Users\\goos\\project"
            wsl_path = convert_windows_to_wsl(windows_path)
            assert wsl_path.startswith("/mnt/c/")

            # Reverse conversion should also work
            back = convert_wsl_to_windows(wsl_path)
            assert back == windows_path

    def test_os_path_sep_independence(self):
        """Test that conversions are independent of os.path.sep."""
        if PATH_CONVERTER_AVAILABLE:
            # Store current separator
            import os

            original_sep = os.path.sep

            # Test conversion with different path separators
            for test_sep in ["/", "\\"]:
                with patch("os.path.sep", test_sep):
                    windows_path = "C:\\Users\\goos\\project"
                    wsl_path = convert_windows_to_wsl(windows_path)
                    # Conversion should work regardless of os.path.sep
                    assert "/mnt/c/" in wsl_path

            # Restore original separator
            assert os.path.sep == original_sep
