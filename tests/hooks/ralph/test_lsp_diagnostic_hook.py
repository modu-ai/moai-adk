"""Unit tests for post_tool__lsp_diagnostic.py hook.

Tests the LSP diagnostic hook that runs after Write/Edit operations.
"""

from __future__ import annotations

import json
import sys
from io import StringIO
from pathlib import Path
from typing import Any
from unittest.mock import patch

import pytest

# Add hooks directory to path
HOOKS_DIR = Path(__file__).parent.parent.parent.parent / ".claude" / "hooks" / "moai"
sys.path.insert(0, str(HOOKS_DIR))


class TestIsSupportedFile:
    """Test is_supported_file function."""

    def test_python_file_supported(self):
        """Python files should be supported."""
        from post_tool__lsp_diagnostic import is_supported_file

        assert is_supported_file("/path/to/file.py") is True
        assert is_supported_file("/path/to/file.pyi") is True

    def test_typescript_file_supported(self):
        """TypeScript files should be supported."""
        from post_tool__lsp_diagnostic import is_supported_file

        assert is_supported_file("/path/to/file.ts") is True
        assert is_supported_file("/path/to/file.tsx") is True

    def test_javascript_file_supported(self):
        """JavaScript files should be supported."""
        from post_tool__lsp_diagnostic import is_supported_file

        assert is_supported_file("/path/to/file.js") is True
        assert is_supported_file("/path/to/file.jsx") is True
        assert is_supported_file("/path/to/file.mjs") is True

    def test_go_file_supported(self):
        """Go files should be supported."""
        from post_tool__lsp_diagnostic import is_supported_file

        assert is_supported_file("/path/to/file.go") is True

    def test_rust_file_supported(self):
        """Rust files should be supported."""
        from post_tool__lsp_diagnostic import is_supported_file

        assert is_supported_file("/path/to/file.rs") is True

    def test_unsupported_file(self):
        """Unsupported files should return False."""
        from post_tool__lsp_diagnostic import is_supported_file

        assert is_supported_file("/path/to/file.txt") is False
        assert is_supported_file("/path/to/file.md") is False
        assert is_supported_file("/path/to/file.json") is False
        assert is_supported_file("/path/to/file.yaml") is False

    def test_empty_path(self):
        """Empty path should return False."""
        from post_tool__lsp_diagnostic import is_supported_file

        assert is_supported_file("") is False

    def test_case_insensitive(self):
        """Extension check should be case insensitive."""
        from post_tool__lsp_diagnostic import is_supported_file

        assert is_supported_file("/path/to/file.PY") is True
        assert is_supported_file("/path/to/file.Ts") is True


class TestGetLanguageForFile:
    """Test get_language_for_file function."""

    def test_python_language(self):
        """Python files should return python."""
        from post_tool__lsp_diagnostic import get_language_for_file

        assert get_language_for_file("/path/to/file.py") == "python"

    def test_typescript_language(self):
        """TypeScript files should return correct language."""
        from post_tool__lsp_diagnostic import get_language_for_file

        assert get_language_for_file("/path/to/file.ts") == "typescript"
        assert get_language_for_file("/path/to/file.tsx") == "typescriptreact"

    def test_javascript_language(self):
        """JavaScript files should return correct language."""
        from post_tool__lsp_diagnostic import get_language_for_file

        assert get_language_for_file("/path/to/file.js") == "javascript"
        assert get_language_for_file("/path/to/file.jsx") == "javascriptreact"

    def test_unknown_language(self):
        """Unknown files should return None."""
        from post_tool__lsp_diagnostic import get_language_for_file

        assert get_language_for_file("/path/to/file.txt") is None


class TestFormatDiagnosticOutput:
    """Test format_diagnostic_output function."""

    def test_format_with_error(self):
        """Format output when LSP error occurred."""
        from post_tool__lsp_diagnostic import format_diagnostic_output

        result: dict[str, Any] = {
            "available": False,
            "error": "LSP server not available",
            "error_count": 0,
            "warning_count": 0,
            "info_count": 0,
            "hint_count": 0,
            "diagnostics": [],
        }

        output = format_diagnostic_output(result, "/path/to/file.py")
        assert "unavailable" in output.lower() or "not available" in output.lower()

    def test_format_no_issues(self):
        """Format output when no issues found."""
        from post_tool__lsp_diagnostic import format_diagnostic_output

        result: dict[str, Any] = {
            "available": True,
            "error": None,
            "error_count": 0,
            "warning_count": 0,
            "info_count": 0,
            "hint_count": 0,
            "diagnostics": [],
        }

        output = format_diagnostic_output(result, "/path/to/file.py")
        assert "No issues" in output or "no issues" in output.lower()

    def test_format_with_errors(self):
        """Format output when errors found."""
        from post_tool__lsp_diagnostic import format_diagnostic_output

        result: dict[str, Any] = {
            "available": True,
            "error": None,
            "error_count": 2,
            "warning_count": 1,
            "info_count": 0,
            "hint_count": 0,
            "diagnostics": [
                {
                    "severity": "error",
                    "message": "undefined name 'x'",
                    "line": 10,
                    "source": "pyright",
                    "code": "reportUndefinedVariable",
                },
                {
                    "severity": "error",
                    "message": "missing return",
                    "line": 20,
                    "source": "pyright",
                    "code": None,
                },
                {
                    "severity": "warning",
                    "message": "unused variable",
                    "line": 15,
                    "source": "pyright",
                    "code": None,
                },
            ],
        }

        output = format_diagnostic_output(result, "/path/to/file.py")
        assert "2 error" in output
        assert "1 warning" in output
        assert "undefined name" in output or "ERROR" in output

    def test_format_truncates_long_diagnostics(self):
        """Long diagnostic lists should be truncated."""
        from post_tool__lsp_diagnostic import format_diagnostic_output

        result: dict[str, Any] = {
            "available": True,
            "error": None,
            "error_count": 10,
            "warning_count": 0,
            "info_count": 0,
            "hint_count": 0,
            "diagnostics": [
                {
                    "severity": "error",
                    "message": f"error {i}",
                    "line": i,
                    "source": "test",
                    "code": None,
                }
                for i in range(10)
            ],
        }

        output = format_diagnostic_output(result, "/path/to/file.py")
        # Should show top 5 and indicate more
        assert "more" in output.lower()


class TestLoadRalphConfig:
    """Test load_ralph_config function."""

    def test_default_config(self):
        """Default config should be returned when file not found."""
        from post_tool__lsp_diagnostic import load_ralph_config

        with patch("post_tool__lsp_diagnostic.get_project_dir") as mock_dir:
            mock_dir.return_value = Path("/nonexistent/path")
            config = load_ralph_config()

            assert config["enabled"] is True
            assert config["hooks"]["post_tool_lsp"]["enabled"] is True

    def test_config_from_file(self, tmp_path):
        """Config should be loaded from file when available."""
        from post_tool__lsp_diagnostic import load_ralph_config

        # Create config file
        config_dir = tmp_path / ".moai" / "config" / "sections"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "ralph.yaml"
        config_file.write_text("""
ralph:
  enabled: false
  hooks:
    post_tool_lsp:
      enabled: true
      severity_threshold: warning
""")

        with patch("post_tool__lsp_diagnostic.get_project_dir") as mock_dir:
            mock_dir.return_value = tmp_path
            config = load_ralph_config()

            assert config["enabled"] is False
            assert config["hooks"]["post_tool_lsp"]["severity_threshold"] == "warning"


class TestRunFallbackDiagnostics:
    """Test run_fallback_diagnostics function."""

    def test_fallback_for_python(self):
        """Fallback should work for Python files."""
        from post_tool__lsp_diagnostic import run_fallback_diagnostics

        with patch("shutil.which") as mock_which:
            mock_which.return_value = None  # No tools available
            result = run_fallback_diagnostics("/path/to/file.py")

            assert result.get("fallback") is True

    def test_fallback_unsupported_language(self):
        """Fallback should return empty for unsupported languages."""
        from post_tool__lsp_diagnostic import run_fallback_diagnostics

        result = run_fallback_diagnostics("/path/to/file.txt")
        assert result["available"] is False


class TestMainFunction:
    """Test main hook entry point."""

    def test_disabled_via_env(self):
        """Hook should exit when disabled via environment."""
        import post_tool__lsp_diagnostic

        with patch.dict("os.environ", {"MOAI_DISABLE_LSP_DIAGNOSTIC": "true"}):
            with pytest.raises(SystemExit) as exc_info:
                post_tool__lsp_diagnostic.main()
            assert exc_info.value.code == 0

    def test_non_write_tool_ignored(self):
        """Non-Write/Edit tools should be ignored."""
        import post_tool__lsp_diagnostic

        input_data = {"tool_name": "Read", "tool_input": {"file_path": "/path/to/file.py"}}

        with patch("sys.stdin", StringIO(json.dumps(input_data))):
            with patch.dict("os.environ", {}, clear=False):
                # Remove disable flag if present
                import os

                os.environ.pop("MOAI_DISABLE_LSP_DIAGNOSTIC", None)

                with pytest.raises(SystemExit) as exc_info:
                    post_tool__lsp_diagnostic.main()
                assert exc_info.value.code == 0

    def test_unsupported_file_ignored(self):
        """Unsupported files should be ignored."""
        import post_tool__lsp_diagnostic

        input_data = {"tool_name": "Write", "tool_input": {"file_path": "/path/to/file.txt"}}

        with patch("sys.stdin", StringIO(json.dumps(input_data))):
            with patch.dict("os.environ", {}, clear=False):
                import os

                os.environ.pop("MOAI_DISABLE_LSP_DIAGNOSTIC", None)

                with pytest.raises(SystemExit) as exc_info:
                    post_tool__lsp_diagnostic.main()
                assert exc_info.value.code == 0
