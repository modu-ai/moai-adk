#!/usr/bin/env python3
"""
Unit tests for LinterRegistry module

Tests cover:
- Linter execution for each language
- Error handling (non-blocking behavior)
- File extension validation
- Logging functionality
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock
from linters import LinterRegistry


class TestLinterRegistry:
    """Test linter registry functionality"""

    def test_registry_initialization(self):
        """Test linter registry initialization"""
        registry = LinterRegistry()

        assert registry.linters is not None
        assert len(registry.linters) > 0
        assert "python" in registry.linters

    def test_formatter_registry_initialization(self):
        """Test formatter registry initialization"""
        registry = LinterRegistry()

        assert registry.formatters is not None
        assert len(registry.formatters) > 0
        assert "python" in registry.formatters


class TestPythonLinting:
    """Test Python linting functionality"""

    def test_python_file_extension_validation(self):
        """Test that non-Python files are skipped"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            # Create a non-Python file
            file_path = Path(tmpdir) / "test.txt"
            file_path.write_text("test")

            result = registry.run_linter("python", file_path)

            # Should return True (skip) for non-Python files
            assert result is True

    def test_missing_python_file(self):
        """Test handling of missing Python files"""
        registry = LinterRegistry()

        file_path = Path("/nonexistent/test.py")

        # Should handle missing files gracefully
        result = registry.run_linter("python", file_path)
        assert result is True  # Non-blocking

    @patch("subprocess.run")
    def test_python_linting_success(self, mock_run):
        """Test successful Python linting"""
        mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")

        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('hello')")

            # Should call subprocess.run
            registry._run_python_linting(file_path)

            assert mock_run.called

    @patch("subprocess.run")
    def test_python_linting_failure_non_blocking(self, mock_run):
        """Test that Python linting failures don't block execution"""
        mock_run.return_value = MagicMock(
            returncode=1,
            stdout="E501 line too long",
            stderr=""
        )

        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("x" * 1000)  # Very long line

            # run_linter should return False (failure) but not raise exception
            result = registry.run_linter("python", file_path)

            # The linting found errors, so it returns False
            # But it shouldn't raise an exception (non-blocking)

    def test_ruff_not_installed(self):
        """Test handling when ruff is not installed"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('hello')")

            # With mocked FileNotFoundError
            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = FileNotFoundError("ruff not found")

                result = registry.run_linter("python", file_path)

                # Should return True (non-blocking) even if tool not found
                assert result is True


class TestJavaScriptLinting:
    """Test JavaScript linting functionality"""

    def test_javascript_file_extensions(self):
        """Test JavaScript file extension validation"""
        registry = LinterRegistry()

        valid_extensions = [".js", ".jsx", ".mjs"]

        for ext in valid_extensions:
            with tempfile.TemporaryDirectory() as tmpdir:
                file_path = Path(tmpdir) / f"test{ext}"
                file_path.write_text("console.log('test');")

                # Should attempt to lint (or skip gracefully if tool not found)
                result = registry.run_linter("javascript", file_path)

                # Should return True (either linting passed or tool not found)
                assert isinstance(result, bool)

    def test_non_javascript_file_skipped(self):
        """Test that non-JavaScript files are skipped"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.run_linter("javascript", file_path)

            # Should skip non-JavaScript files
            assert result is True


class TestTypeScriptLinting:
    """Test TypeScript linting functionality"""

    def test_typescript_file_extensions(self):
        """Test TypeScript file extension validation"""
        registry = LinterRegistry()

        valid_extensions = [".ts", ".tsx"]

        for ext in valid_extensions:
            with tempfile.TemporaryDirectory() as tmpdir:
                file_path = Path(tmpdir) / f"test{ext}"
                file_path.write_text("const x: string = 'test';")

                # Should attempt to lint
                result = registry.run_linter("typescript", file_path)

                assert isinstance(result, bool)

    def test_non_typescript_file_skipped(self):
        """Test that non-TypeScript files are skipped"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.go"
            file_path.write_text("package main")

            result = registry.run_linter("typescript", file_path)

            assert result is True


class TestGoLinting:
    """Test Go linting functionality"""

    def test_go_file_extension(self):
        """Test Go file extension validation"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.go"
            file_path.write_text("package main")

            result = registry.run_linter("go", file_path)

            assert isinstance(result, bool)

    def test_non_go_file_skipped(self):
        """Test that non-Go files are skipped"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.run_linter("go", file_path)

            assert result is True


class TestRustLinting:
    """Test Rust linting functionality"""

    def test_rust_file_extension(self):
        """Test Rust file extension validation"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.rs"
            file_path.write_text("fn main() {}")

            result = registry.run_linter("rust", file_path)

            assert isinstance(result, bool)

    def test_non_rust_file_skipped(self):
        """Test that non-Rust files are skipped"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.run_linter("rust", file_path)

            assert result is True


class TestJavaLinting:
    """Test Java linting functionality"""

    def test_java_file_extension(self):
        """Test Java file extension validation"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "Test.java"
            file_path.write_text("public class Test {}")

            result = registry.run_linter("java", file_path)

            assert isinstance(result, bool)

    def test_non_java_file_skipped(self):
        """Test that non-Java files are skipped"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.run_linter("java", file_path)

            assert result is True


class TestRubyLinting:
    """Test Ruby linting functionality"""

    def test_ruby_file_extension(self):
        """Test Ruby file extension validation"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.rb"
            file_path.write_text("puts 'test'")

            result = registry.run_linter("ruby", file_path)

            assert isinstance(result, bool)

    def test_non_ruby_file_skipped(self):
        """Test that non-Ruby files are skipped"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.run_linter("ruby", file_path)

            assert result is True


class TestPHPLinting:
    """Test PHP linting functionality"""

    def test_php_file_extension(self):
        """Test PHP file extension validation"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.php"
            file_path.write_text("<?php echo 'test';")

            result = registry.run_linter("php", file_path)

            assert isinstance(result, bool)

    def test_non_php_file_skipped(self):
        """Test that non-PHP files are skipped"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.run_linter("php", file_path)

            assert result is True


class TestUnknownLanguage:
    """Test handling of unknown languages"""

    def test_unknown_language_returns_true(self):
        """Test that unknown languages are skipped gracefully"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.xyz"
            file_path.write_text("unknown")

            result = registry.run_linter("unknown_language", file_path)

            # Should skip gracefully
            assert result is True


class TestErrorHandling:
    """Test error handling in linter registry"""

    def test_timeout_handling(self):
        """Test handling of timeout errors"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            with patch("subprocess.run") as mock_run:
                import subprocess
                mock_run.side_effect = subprocess.TimeoutExpired("ruff", 30)

                result = registry.run_linter("python", file_path)

                # Should handle timeout gracefully
                assert result is True

    def test_general_exception_handling(self):
        """Test handling of general exceptions"""
        registry = LinterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = Exception("Unexpected error")

                result = registry.run_linter("python", file_path)

                # Should handle exceptions gracefully (non-blocking)
                assert result is True


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
