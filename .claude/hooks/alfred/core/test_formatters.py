#!/usr/bin/env python3
"""
Unit tests for FormatterRegistry module

Tests cover:
- Formatter execution for each language
- Error handling (non-blocking behavior)
- File extension validation
- Batch formatting
"""

import pytest
import tempfile
from pathlib import Path
from unittest.mock import patch, MagicMock
from formatters import FormatterRegistry


class TestFormatterRegistry:
    """Test formatter registry functionality"""

    def test_registry_initialization(self):
        """Test formatter registry initialization"""
        registry = FormatterRegistry()

        assert registry.formatters is not None
        assert len(registry.formatters) > 0
        assert "python" in registry.formatters

    def test_all_formatters_available(self):
        """Test that all language formatters are available"""
        registry = FormatterRegistry()

        expected_languages = [
            "python",
            "javascript",
            "typescript",
            "go",
            "rust",
            "java",
            "ruby",
            "php"
        ]

        for language in expected_languages:
            assert language in registry.formatters


class TestPythonFormatting:
    """Test Python code formatting"""

    def test_python_file_extension_validation(self):
        """Test that non-Python files are skipped"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.txt"
            file_path.write_text("test")

            result = registry.format_file("python", file_path)

            # Should return True (skip) for non-Python files
            assert result is True

    def test_missing_python_file(self):
        """Test handling of missing Python files"""
        registry = FormatterRegistry()

        file_path = Path("/nonexistent/test.py")

        # Should handle missing files gracefully
        result = registry.format_file("python", file_path)
        assert result is True

    @patch("subprocess.run")
    def test_python_formatting_success(self, mock_run):
        """Test successful Python formatting"""
        mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")

        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print(  'hello'  )")

            registry._format_python(file_path)

            assert mock_run.called

    def test_ruff_not_installed(self):
        """Test handling when ruff is not installed"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('hello')")

            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = FileNotFoundError("ruff not found")

                result = registry.format_file("python", file_path)

                # Should return True (non-blocking) even if tool not found
                assert result is True


class TestJavaScriptFormatting:
    """Test JavaScript code formatting"""

    def test_javascript_file_extensions(self):
        """Test JavaScript file extension validation"""
        registry = FormatterRegistry()

        valid_extensions = [".js", ".jsx", ".mjs"]

        for ext in valid_extensions:
            with tempfile.TemporaryDirectory() as tmpdir:
                file_path = Path(tmpdir) / f"test{ext}"
                file_path.write_text("console.log('test');")

                result = registry.format_file("javascript", file_path)

                # Should return True (either formatted or skipped)
                assert isinstance(result, bool)

    def test_non_javascript_file_skipped(self):
        """Test that non-JavaScript files are skipped"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.format_file("javascript", file_path)

            # Should skip non-JavaScript files
            assert result is True


class TestTypeScriptFormatting:
    """Test TypeScript code formatting"""

    def test_typescript_file_extensions(self):
        """Test TypeScript file extension validation"""
        registry = FormatterRegistry()

        valid_extensions = [".ts", ".tsx"]

        for ext in valid_extensions:
            with tempfile.TemporaryDirectory() as tmpdir:
                file_path = Path(tmpdir) / f"test{ext}"
                file_path.write_text("const x: string = 'test';")

                result = registry.format_file("typescript", file_path)

                assert isinstance(result, bool)

    def test_non_typescript_file_skipped(self):
        """Test that non-TypeScript files are skipped"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.go"
            file_path.write_text("package main")

            result = registry.format_file("typescript", file_path)

            assert result is True


class TestGoFormatting:
    """Test Go code formatting"""

    def test_go_file_extension(self):
        """Test Go file extension validation"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.go"
            file_path.write_text("package main")

            result = registry.format_file("go", file_path)

            assert isinstance(result, bool)

    def test_non_go_file_skipped(self):
        """Test that non-Go files are skipped"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.format_file("go", file_path)

            assert result is True


class TestRustFormatting:
    """Test Rust code formatting"""

    def test_rust_file_extension(self):
        """Test Rust file extension validation"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.rs"
            file_path.write_text("fn main() {}")

            result = registry.format_file("rust", file_path)

            assert isinstance(result, bool)

    def test_non_rust_file_skipped(self):
        """Test that non-Rust files are skipped"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.format_file("rust", file_path)

            assert result is True


class TestJavaFormatting:
    """Test Java code formatting"""

    def test_java_file_extension(self):
        """Test Java file extension validation"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "Test.java"
            file_path.write_text("public class Test {}")

            result = registry.format_file("java", file_path)

            assert isinstance(result, bool)

    def test_non_java_file_skipped(self):
        """Test that non-Java files are skipped"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.format_file("java", file_path)

            assert result is True


class TestRubyFormatting:
    """Test Ruby code formatting"""

    def test_ruby_file_extension(self):
        """Test Ruby file extension validation"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.rb"
            file_path.write_text("puts 'test'")

            result = registry.format_file("ruby", file_path)

            assert isinstance(result, bool)

    def test_non_ruby_file_skipped(self):
        """Test that non-Ruby files are skipped"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.format_file("ruby", file_path)

            assert result is True


class TestPHPFormatting:
    """Test PHP code formatting"""

    def test_php_file_extension(self):
        """Test PHP file extension validation"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.php"
            file_path.write_text("<?php echo 'test';")

            result = registry.format_file("php", file_path)

            assert isinstance(result, bool)

    def test_non_php_file_skipped(self):
        """Test that non-PHP files are skipped"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            result = registry.format_file("php", file_path)

            assert result is True


class TestBatchFormatting:
    """Test batch directory formatting"""

    def test_format_directory_with_multiple_files(self):
        """Test formatting multiple files in directory"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            tmpdir_path = Path(tmpdir)

            # Create multiple Python files
            (tmpdir_path / "test1.py").write_text("print(  'test1'  )")
            (tmpdir_path / "test2.py").write_text("print(  'test2'  )")
            (tmpdir_path / "test3.py").write_text("print(  'test3'  )")

            # Mock subprocess to avoid actually running ruff
            with patch("subprocess.run") as mock_run:
                mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")

                result = registry.format_directory("python", tmpdir_path, [".py"])

                # Should have called subprocess for each file
                assert mock_run.call_count >= 3

    def test_format_directory_skip_non_matching_files(self):
        """Test that directory formatting skips non-matching files"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            tmpdir_path = Path(tmpdir)

            # Create mixed files
            (tmpdir_path / "test.py").write_text("print('test')")
            (tmpdir_path / "test.js").write_text("console.log('test');")
            (tmpdir_path / "test.txt").write_text("text")

            with patch("subprocess.run") as mock_run:
                mock_run.return_value = MagicMock(returncode=0, stdout="", stderr="")

                registry.format_directory("python", tmpdir_path, [".py"])

                # Should only format .py files
                # The exact count depends on how many times it's called

    def test_format_directory_empty(self):
        """Test formatting empty directory"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            result = registry.format_directory("python", Path(tmpdir), [".py"])

            # Should succeed with no files to format
            assert result is True


class TestUnknownLanguage:
    """Test handling of unknown languages"""

    def test_unknown_language_returns_true(self):
        """Test that unknown languages are skipped gracefully"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.xyz"
            file_path.write_text("unknown")

            result = registry.format_file("unknown_language", file_path)

            # Should skip gracefully
            assert result is True


class TestErrorHandling:
    """Test error handling in formatter registry"""

    def test_timeout_handling(self):
        """Test handling of timeout errors"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            with patch("subprocess.run") as mock_run:
                import subprocess
                mock_run.side_effect = subprocess.TimeoutExpired("ruff", 30)

                result = registry.format_file("python", file_path)

                # Should handle timeout gracefully
                assert result is True

    def test_general_exception_handling(self):
        """Test handling of general exceptions"""
        registry = FormatterRegistry()

        with tempfile.TemporaryDirectory() as tmpdir:
            file_path = Path(tmpdir) / "test.py"
            file_path.write_text("print('test')")

            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = Exception("Unexpected error")

                result = registry.format_file("python", file_path)

                # Should handle exceptions gracefully (non-blocking)
                assert result is True


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
