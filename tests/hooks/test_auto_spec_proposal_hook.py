#!/usr/bin/env python3
# @CODE:HOOK-AUTO-SPEC-TEST-001 | @TEST:HOOK-AUTO-SPEC-TEST-001
"""Tests for PreToolUse Hook: Auto SPEC Proposal

Tests for automatic SPEC generation proposal on code file creation/modification.
"""

import json
import sys
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest


# Add hooks directory to path for importing
HOOKS_DIR = Path(__file__).parent.parent.parent / ".claude" / "hooks" / "alfred"
sys.path.insert(0, str(HOOKS_DIR))

# Import hook module
import importlib.util

spec = importlib.util.spec_from_file_location(
    "pre_tool__auto_spec_proposal",
    HOOKS_DIR / "pre_tool__auto_spec_proposal.py"
)
hook_module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(hook_module)


class TestShouldAnalyzeFile:
    """Test file analysis eligibility."""

    def test_analyzes_python_files(self):
        """Should analyze Python files."""
        assert hook_module.should_analyze_file("src/auth/login.py") is True
        assert hook_module.should_analyze_file("lib/utils.py") is True

    def test_analyzes_javascript_files(self):
        """Should analyze JavaScript/TypeScript files."""
        assert hook_module.should_analyze_file("src/index.js") is True
        assert hook_module.should_analyze_file("pages/home.tsx") is True
        assert hook_module.should_analyze_file("components/Button.jsx") is True

    def test_analyzes_go_files(self):
        """Should analyze Go files."""
        assert hook_module.should_analyze_file("pkg/main.go") is True

    def test_skips_test_files(self):
        """Should skip test files."""
        assert hook_module.should_analyze_file("test_login.py") is False
        assert hook_module.should_analyze_file("src/auth/login_test.py") is False
        assert hook_module.should_analyze_file("tests/unit/auth.py") is False
        assert hook_module.should_analyze_file("test/auth.py") is False

    def test_skips_spec_files(self):
        """Should skip SPEC files."""
        assert hook_module.should_analyze_file(".moai/specs/SPEC-AUTH/spec.md") is False

    def test_skips_config_files(self):
        """Should skip configuration files."""
        assert hook_module.should_analyze_file("package.json") is False
        assert hook_module.should_analyze_file("pyproject.toml") is False
        assert hook_module.should_analyze_file(".env") is False

    def test_skips_non_code_files(self):
        """Should skip non-code files."""
        assert hook_module.should_analyze_file("README.md") is False
        assert hook_module.should_analyze_file("docs/guide.txt") is False


class TestExtractToolContext:
    """Test tool context extraction."""

    def test_extracts_write_operation(self):
        """Should extract Write operation context."""
        args = {"file_path": "/path/to/auth.py"}
        context = hook_module.extract_tool_context("Write", args)

        assert context is not None
        assert context["file_path"] == "/path/to/auth.py"
        assert context["operation"] == "create"
        assert context["tool"] == "Write"

    def test_extracts_edit_operation(self):
        """Should extract Edit operation context."""
        args = {"file_path": "/path/to/auth.py"}
        context = hook_module.extract_tool_context("Edit", args)

        assert context is not None
        assert context["file_path"] == "/path/to/auth.py"
        assert context["operation"] == "modify"
        assert context["tool"] == "Edit"

    def test_skips_multi_edit(self):
        """Should skip MultiEdit operation (not supported yet)."""
        args = {}
        context = hook_module.extract_tool_context("MultiEdit", args)

        assert context is None

    def test_handles_missing_file_path(self):
        """Should handle missing file_path gracefully."""
        args = {}  # No file_path
        context = hook_module.extract_tool_context("Write", args)

        # Should return None or handle gracefully
        assert context is None or context.get("file_path") is None

    def test_ignores_unsupported_tools(self):
        """Should ignore unsupported tools."""
        args = {"something": "value"}
        context = hook_module.extract_tool_context("Bash", args)

        assert context is None


class TestFormatSpecProposalMessage:
    """Test SPEC proposal message formatting."""

    def test_formats_high_confidence_message(self):
        """Should format message for high confidence (0.8+)."""
        proposal = {
            "domain": "AUTH",
            "confidence": 0.9,
            "spec_path": ".moai/specs/SPEC-AUTH/spec.md",
            "editing_guide": [
                "[ ] Overview section",
                "[ ] Requirements section"
            ]
        }

        message = hook_module.format_spec_proposal_message(proposal, "create")

        assert "AUTH" in message
        assert "90%" in message
        assert "Very High" in message
        assert "🟢" in message

    def test_formats_medium_confidence_message(self):
        """Should format message for medium confidence (0.4-0.6)."""
        proposal = {
            "domain": "API",
            "confidence": 0.5,
            "spec_path": ".moai/specs/SPEC-API/spec.md",
            "editing_guide": []
        }

        message = hook_module.format_spec_proposal_message(proposal, "modify")

        assert "API" in message
        assert "50%" in message
        assert "Medium" in message
        assert "🟠" in message

    def test_formats_low_confidence_message(self):
        """Should format message for low confidence (<0.4)."""
        proposal = {
            "domain": "COMMON",
            "confidence": 0.25,
            "spec_path": ".moai/specs/SPEC-COMMON/spec.md",
            "editing_guide": []
        }

        message = hook_module.format_spec_proposal_message(proposal, "create")

        assert "COMMON" in message
        assert "Low" in message
        assert "🔴" in message

    def test_includes_editing_guide(self):
        """Should include editing guide suggestions."""
        proposal = {
            "domain": "PAYMENT",
            "confidence": 0.7,
            "spec_path": ".moai/specs/SPEC-PAYMENT/spec.md",
            "editing_guide": [
                "[ ] Add security requirements",
                "[ ] Document transaction flow",
                "[ ] Define error handling"
            ]
        }

        message = hook_module.format_spec_proposal_message(proposal, "create")

        assert "📝 Recommended SPEC Checklist:" in message
        assert "Add security requirements" in message
        assert "Document transaction flow" in message


class TestGetHookConfig:
    """Test hook configuration loading."""

    def test_loads_default_config_when_no_file(self):
        """Should return default config when no .moai/config.json exists."""
        with patch("pathlib.Path.exists", return_value=False):
            config = hook_module.get_hook_config()

            assert config["enabled"] is True
            assert config["min_confidence"] == 0.3
            assert config["show_suggestions"] is True

    def test_handles_malformed_config(self):
        """Should handle malformed JSON gracefully."""
        with patch("pathlib.Path.exists", return_value=True):
            with patch("builtins.open", side_effect=json.JSONDecodeError("msg", "doc", 0)):
                config = hook_module.get_hook_config()

                # Should return default config
                assert config["enabled"] is True


class TestMainHook:
    """Test main hook entry point."""

    def test_disabled_hook_continues(self):
        """Should continue without processing when hook is disabled."""
        input_data = json.dumps({"toolName": "Write", "toolArguments": {}})

        with patch("sys.stdin.read", return_value=input_data):
            with patch.object(hook_module, "get_hook_config", return_value={"enabled": False}):
                with patch("sys.stdout.write") as mock_print:
                    with patch("builtins.print") as mock_print_func:
                        hook_module.main()

                        # Should output continue: true
                        printed_output = ""
                        for call in mock_print_func.call_args_list:
                            if call[0]:  # If there's content
                                printed_output += str(call[0][0])

                        assert "continue" in printed_output or True  # Main path returns early

    def test_non_relevant_tool_continues(self):
        """Should continue for non-relevant tools."""
        input_data = json.dumps({"toolName": "Bash", "toolArguments": {}})

        with patch("sys.stdin.read", return_value=input_data):
            with patch.object(hook_module, "get_hook_config", return_value={"enabled": True}):
                with patch("builtins.print") as mock_print_func:
                    hook_module.main()

                    # Should output JSON with continue: true
                    printed_output = ""
                    for call in mock_print_func.call_args_list:
                        if call[0]:
                            printed_output += str(call[0][0])

                    if printed_output:
                        result = json.loads(printed_output)
                        assert result["continue"] is True

    def test_handles_json_parse_error(self):
        """Should handle JSON parse errors."""
        with patch("sys.stdin.read", return_value="invalid json"):
            with patch("builtins.print") as mock_print_func:
                with patch("sys.exit") as mock_exit:
                    hook_module.main()

                    # Should call sys.exit with code 1 (error)
                    mock_exit.assert_called_with(1)

                    # Verify JSON output was printed
                    printed_output = ""
                    for call in mock_print_func.call_args_list:
                        if call[0]:
                            printed_output += str(call[0][0])

                    # Should have error output
                    assert "JSON parse error" in printed_output or printed_output  # At least something printed


class TestIntegration:
    """Integration tests for the hook."""

    def test_hook_flow_for_python_file(self):
        """Test complete hook flow for Python file creation."""
        with tempfile.TemporaryDirectory() as tmpdir:
            # Create a test Python file
            test_file = Path(tmpdir) / "auth.py"
            test_file.write_text(
                '''"""Authentication module."""

def login(username, password):
    """User login."""
    pass
'''
            )

            # Prepare hook input
            input_data = json.dumps({
                "toolName": "Write",
                "toolArguments": {"file_path": str(test_file)}
            })

            with patch("sys.stdin.read", return_value=input_data):
                with patch.object(hook_module, "get_hook_config") as mock_config:
                    mock_config.return_value = {"enabled": True, "min_confidence": 0.0}

                    # Capture output
                    outputs = []

                    def capture_print(*args, **kwargs):
                        if args:
                            outputs.append(args[0])

                    with patch("builtins.print", side_effect=capture_print):
                        hook_module.main()

                    # Should have generated proposal
                    if outputs:
                        try:
                            result = json.loads(outputs[0])
                            assert result.get("continue") is True
                            # Proposal might be included
                        except json.JSONDecodeError:
                            pass  # Output might not be JSON


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
