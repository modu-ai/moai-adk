#!/usr/bin/env python3
"""Tests for platform detection utilities

Unit tests for core/platform.py module to ensure reliable cross-platform
detection across Windows, macOS, Linux, and unknown platforms.

Run with:
    python -m pytest test_platform.py -v

@TEST:WINDOWS-HOOK-COMPAT-001
"""

import sys
import unittest
from unittest.mock import patch

from core.platform import PlatformType, get_platform, is_windows


class TestPlatformDetection(unittest.TestCase):
    """Test suite for platform detection utilities

    Ensures reliable OS detection across all supported platforms using
    mock sys.platform values to simulate different environments.
    """

    def test_get_platform_windows_win32(self):
        """Test Windows platform detection with win32

        Standard Windows Python reports sys.platform as 'win32' even on 64-bit.
        This is the most common case for Windows detection.
        """
        with patch.object(sys, "platform", "win32"):
            result = get_platform()
            self.assertEqual(result, "windows")
            self.assertIsInstance(result, str)

    def test_get_platform_windows_variants(self):
        """Test Windows platform detection with cygwin and msys

        Git Bash and other Unix-like shells on Windows may report different
        sys.platform values. Ensure all Windows variants are detected.
        """
        # Test cygwin (Git Bash on Windows)
        with patch.object(sys, "platform", "cygwin"):
            self.assertEqual(get_platform(), "windows")

        # Test msys (MSYS2 on Windows)
        with patch.object(sys, "platform", "msys"):
            self.assertEqual(get_platform(), "windows")

        # Test win64 (rare but possible)
        with patch.object(sys, "platform", "win64"):
            self.assertEqual(get_platform(), "windows")

    def test_get_platform_macos(self):
        """Test macOS platform detection

        macOS always reports sys.platform as 'darwin' across all versions
        (Intel and Apple Silicon).
        """
        with patch.object(sys, "platform", "darwin"):
            result = get_platform()
            self.assertEqual(result, "darwin")
            self.assertIsInstance(result, str)

    def test_get_platform_linux(self):
        """Test Linux platform detection with variants

        Linux may report 'linux' (Python 3) or 'linux2' (Python 2 legacy).
        Ensure both are detected correctly.
        """
        # Modern Python 3.x
        with patch.object(sys, "platform", "linux"):
            self.assertEqual(get_platform(), "linux")

        # Legacy Python 2.x (still present in some systems)
        with patch.object(sys, "platform", "linux2"):
            self.assertEqual(get_platform(), "linux")

        # Case insensitivity test
        with patch.object(sys, "platform", "LINUX"):
            self.assertEqual(get_platform(), "linux")

    def test_get_platform_unknown(self):
        """Test unknown platform detection

        For unsupported platforms (FreeBSD, Solaris, etc.), function should
        gracefully return 'unknown' without raising exceptions.
        """
        # Test FreeBSD
        with patch.object(sys, "platform", "freebsd12"):
            self.assertEqual(get_platform(), "unknown")

        # Test Solaris
        with patch.object(sys, "platform", "sunos5"):
            self.assertEqual(get_platform(), "unknown")

        # Test completely unknown
        with patch.object(sys, "platform", "exotic-os-v1"):
            self.assertEqual(get_platform(), "unknown")

    def test_is_windows_true(self):
        """Test is_windows returns True on Windows platforms

        Convenience function should correctly identify all Windows variants
        as Windows platform.
        """
        # Standard Windows
        with patch.object(sys, "platform", "win32"):
            self.assertTrue(is_windows())
            self.assertIsInstance(is_windows(), bool)

        # Git Bash
        with patch.object(sys, "platform", "cygwin"):
            self.assertTrue(is_windows())

        # MSYS2
        with patch.object(sys, "platform", "msys"):
            self.assertTrue(is_windows())

    def test_is_windows_false_on_unix(self):
        """Test is_windows returns False on Unix-like platforms

        Should return False for macOS, Linux, and any non-Windows platform.
        """
        # macOS
        with patch.object(sys, "platform", "darwin"):
            self.assertFalse(is_windows())
            self.assertIsInstance(is_windows(), bool)

        # Linux
        with patch.object(sys, "platform", "linux"):
            self.assertFalse(is_windows())

        # Unknown platform
        with patch.object(sys, "platform", "freebsd12"):
            self.assertFalse(is_windows())

    def test_platform_type_annotation(self):
        """Test PlatformType literal type hint coverage

        Ensures all valid return values are covered by the PlatformType
        type annotation for better IDE support and type checking.
        """
        # Test that all possible return values match PlatformType
        valid_platforms = ["windows", "darwin", "linux", "unknown"]

        with patch.object(sys, "platform", "win32"):
            self.assertIn(get_platform(), valid_platforms)

        with patch.object(sys, "platform", "darwin"):
            self.assertIn(get_platform(), valid_platforms)

        with patch.object(sys, "platform", "linux"):
            self.assertIn(get_platform(), valid_platforms)

        with patch.object(sys, "platform", "exotic"):
            self.assertIn(get_platform(), valid_platforms)

    def test_case_insensitivity(self):
        """Test platform detection is case-insensitive

        Some environments may report sys.platform in different cases.
        Ensure detection works regardless of case.
        """
        # Windows uppercase
        with patch.object(sys, "platform", "WIN32"):
            self.assertEqual(get_platform(), "windows")

        # Darwin uppercase
        with patch.object(sys, "platform", "DARWIN"):
            self.assertEqual(get_platform(), "darwin")

        # Linux mixed case
        with patch.object(sys, "platform", "Linux"):
            self.assertEqual(get_platform(), "linux")


class TestPlatformDetectionIntegration(unittest.TestCase):
    """Integration tests for platform detection

    Tests actual system behavior without mocks to ensure real-world
    compatibility on the current platform.
    """

    def test_actual_platform_detection(self):
        """Test platform detection on actual system

        This test verifies that the current system's platform is detected
        and returns one of the expected values.
        """
        platform = get_platform()
        self.assertIn(platform, ["windows", "darwin", "linux", "unknown"])

    def test_is_windows_returns_bool(self):
        """Test is_windows returns boolean on actual system

        Ensures the convenience function always returns a proper boolean
        value in the real environment.
        """
        result = is_windows()
        self.assertIsInstance(result, bool)

    def test_platform_consistency(self):
        """Test platform detection is consistent across multiple calls

        Ensures get_platform() returns the same value when called multiple
        times in the same process.
        """
        first_call = get_platform()
        second_call = get_platform()
        self.assertEqual(first_call, second_call)


if __name__ == "__main__":
    # Run tests with verbose output
    unittest.main(verbosity=2)
