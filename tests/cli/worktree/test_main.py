"""Comprehensive tests for Git Worktree CLI __main__ module.

Test Coverage Strategy:
- main() function: Success path (returns 0), error path (returns 1)
- Exception handling: Verify exceptions are caught and handled
- Error output: Verify errors are printed to stderr
- CLI invocation: Verify worktree CLI is called correctly
- Edge cases: Various exception types, edge conditions
"""

from unittest.mock import MagicMock, patch

import pytest
from moai_adk.cli.worktree import __main__


class TestMainFunction:
    """Test main() function behavior."""

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_returns_zero_on_success(self, mock_worktree: MagicMock) -> None:
        """Test that main() returns 0 when worktree CLI executes successfully."""
        # Mock worktree to succeed without error
        mock_worktree.return_value = None

        result = __main__.main()

        assert result == 0
        mock_worktree.assert_called_once_with(standalone_mode=False)

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_calls_worktree_with_standalone_mode_false(self, mock_worktree: MagicMock) -> None:
        """Test that main() calls worktree with standalone_mode=False."""
        mock_worktree.return_value = None

        __main__.main()

        mock_worktree.assert_called_once_with(standalone_mode=False)

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_returns_one_on_exception(self, mock_worktree: MagicMock) -> None:
        """Test that main() returns 1 when worktree CLI raises an exception."""
        # Mock worktree to raise an exception
        mock_worktree.side_effect = Exception("CLI error")

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    @patch("sys.stderr", new_callable=MagicMock)
    def test_main_prints_error_to_stderr(self, mock_stderr: MagicMock, mock_worktree: MagicMock) -> None:
        """Test that main() prints error message to stderr on exception."""
        # Mock worktree to raise an exception
        error_message = "Worktree not found"
        mock_worktree.side_effect = Exception(error_message)

        __main__.main()

        # Verify error was printed to stderr
        mock_stderr.write.assert_called()
        error_calls = " ".join(str(call) for call in mock_stderr.write.call_args_list)
        assert "Error:" in error_calls
        assert error_message in error_calls

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_handles_runtime_error(self, mock_worktree: MagicMock) -> None:
        """Test that main() handles RuntimeError exceptions."""
        mock_worktree.side_effect = RuntimeError("Runtime error occurred")

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_handles_value_error(self, mock_worktree: MagicMock) -> None:
        """Test that main() handles ValueError exceptions."""
        mock_worktree.side_effect = ValueError("Invalid value")

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_handles_generic_exception(self, mock_worktree: MagicMock) -> None:
        """Test that main() handles generic exceptions."""
        mock_worktree.side_effect = Exception("Unexpected error")

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_with_click_exception(self, mock_worktree: MagicMock) -> None:
        """Test that main() handles Click exceptions."""
        # Import Click to test Click-specific exceptions
        try:
            from click import ClickException

            mock_worktree.side_effect = ClickException("Click error")

            result = __main__.main()

            assert result == 1
        except ImportError:
            pytest.skip("Click not available")

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_preserves_exception_message(
        self, mock_worktree: MagicMock, capsys: pytest.CaptureFixture[str]
    ) -> None:
        """Test that main() preserves the original exception message."""
        original_message = "Worktree for SPEC-001 not found"
        mock_worktree.side_effect = Exception(original_message)

        __main__.main()

        captured = capsys.readouterr()
        assert original_message in captured.err


class TestMainEdgeCases:
    """Test edge cases and boundary conditions for main()."""

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_with_empty_exception_message(self, mock_worktree: MagicMock) -> None:
        """Test main() with exception that has empty message."""
        mock_worktree.side_effect = Exception("")

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_with_long_exception_message(self, mock_worktree: MagicMock) -> None:
        """Test main() with exception that has very long message."""
        long_message = "Error: " + "A" * 1000
        mock_worktree.side_effect = Exception(long_message)

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_with_unicode_exception_message(self, mock_worktree: MagicMock) -> None:
        """Test main() with exception that has Unicode characters."""
        unicode_message = "错误：工作树未找到"
        mock_worktree.side_effect = Exception(unicode_message)

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_with_special_characters_in_exception(self, mock_worktree: MagicMock) -> None:
        """Test main() with exception containing special characters."""
        special_message = "Error: file<script>.py has issues: <>&\"'"
        mock_worktree.side_effect = Exception(special_message)

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_multiple_calls(self, mock_worktree: MagicMock) -> None:
        """Test that main() can be called multiple times."""
        mock_worktree.return_value = None

        # Call main() multiple times
        result1 = __main__.main()
        result2 = __main__.main()
        result3 = __main__.main()

        assert result1 == 0
        assert result2 == 0
        assert result3 == 0
        assert mock_worktree.call_count == 3


class TestMainIntegration:
    """Integration tests for main() with worktree CLI."""

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_worktree_invocation_args(self, mock_worktree: MagicMock) -> None:
        """Test that main() invokes worktree with correct arguments."""
        mock_worktree.return_value = None

        __main__.main()

        # Verify worktree was called with standalone_mode=False
        mock_worktree.assert_called_once()
        call_kwargs = mock_worktree.call_args.kwargs
        assert "standalone_mode" in call_kwargs
        assert call_kwargs["standalone_mode"] is False

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_main_no_positional_args(self, mock_worktree: MagicMock) -> None:
        """Test that main() doesn't pass positional args to worktree."""
        mock_worktree.return_value = None

        __main__.main()

        # Verify no positional arguments were passed
        call_args = mock_worktree.call_args.args
        assert len(call_args) == 0


class TestMainExitCodes:
    """Test exit codes for various scenarios."""

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_exit_code_zero_success(self, mock_worktree: MagicMock) -> None:
        """Test exit code 0 for successful execution."""
        mock_worktree.return_value = None

        result = __main__.main()

        assert result == 0

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_exit_code_one_generic_error(self, mock_worktree: MagicMock) -> None:
        """Test exit code 1 for generic errors."""
        mock_worktree.side_effect = Exception("Generic error")

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_exit_code_one_keyboard_interrupt(self, mock_worktree: MagicMock) -> None:
        """Test exit code 1 for KeyboardInterrupt (Ctrl+C)."""
        mock_worktree.side_effect = KeyboardInterrupt()

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_exit_code_one_system_exit(self, mock_worktree: MagicMock) -> None:
        """Test exit code 1 for SystemExit exceptions."""
        mock_worktree.side_effect = SystemExit(1)

        result = __main__.main()

        assert result == 1


class TestMainErrorOutputFormat:
    """Test error output formatting."""

    @patch("moai_adk.cli.worktree.__main__.worktree")
    @patch("sys.stderr", new_callable=MagicMock)
    def test_error_output_starts_with_error_prefix(self, mock_stderr: MagicMock, mock_worktree: MagicMock) -> None:
        """Test that error output starts with 'Error: ' prefix."""
        mock_worktree.side_effect = Exception("Something went wrong")

        __main__.main()

        # Get the written content
        written_content = str(mock_stderr.write.call_args)
        assert "Error:" in written_content

    @patch("moai_adk.cli.worktree.__main__.worktree")
    @patch("sys.stderr", new_callable=MagicMock)
    def test_error_output_includes_exception_message(self, mock_stderr: MagicMock, mock_worktree: MagicMock) -> None:
        """Test that error output includes the exception message."""
        error_msg = "Worktree for SPEC-001 not found"
        mock_worktree.side_effect = Exception(error_msg)

        __main__.main()

        # Get the written content
        written_content = str(mock_stderr.write.call_args)
        assert error_msg in written_content

    @patch("moai_adk.cli.worktree.__main__.worktree")
    @patch("sys.stderr", new_callable=MagicMock)
    def test_error_output_goes_to_stderr_not_stdout(
        self, mock_worktree: MagicMock, capsys: pytest.CaptureFixture[str]
    ) -> None:
        """Test that error output goes to stderr, not stdout."""
        mock_worktree.side_effect = Exception("Error")

        __main__.main()

        captured = capsys.readouterr()
        # stdout should be empty
        assert captured.out == ""
        # stderr should have content
        assert len(captured.err) > 0


class TestMainStandaloneMode:
    """Test standalone_mode behavior."""

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_standalone_mode_false_allows_exception_handling(self, mock_worktree: MagicMock) -> None:
        """Test that standalone_mode=False allows exceptions to be caught."""
        mock_worktree.side_effect = Exception("Handled error")

        result = __main__.main()

        # Exception should be caught, not propagated
        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_standalone_mode_false_prevents_system_exit(self, mock_worktree: MagicMock) -> None:
        """Test that standalone_mode=False prevents Click from calling sys.exit."""
        # When standalone_mode=False, Click should not call sys.exit
        # Instead, it should raise ClickException or return normally
        mock_worktree.return_value = None

        result = __main__.main()

        # Should return normally, not exit
        assert result == 0


class TestMainWithWorktreeExceptions:
    """Test main() with worktree-specific exceptions."""

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_handles_worktree_exists_error(self, mock_worktree: MagicMock) -> None:
        """Test handling of WorktreeExistsError."""
        from moai_adk.cli.worktree.exceptions import WorktreeExistsError

        from pathlib import Path

        mock_worktree.side_effect = WorktreeExistsError("SPEC-001", Path("/path"))

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_handles_worktree_not_found_error(self, mock_worktree: MagicMock) -> None:
        """Test handling of WorktreeNotFoundError."""
        from moai_adk.cli.worktree.exceptions import WorktreeNotFoundError

        mock_worktree.side_effect = WorktreeNotFoundError("SPEC-001")

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_handles_uncommitted_changes_error(self, mock_worktree: MagicMock) -> None:
        """Test handling of UncommittedChangesError."""
        from moai_adk.cli.worktree.exceptions import UncommittedChangesError

        mock_worktree.side_effect = UncommittedChangesError("SPEC-001")

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_handles_git_operation_error(self, mock_worktree: MagicMock) -> None:
        """Test handling of GitOperationError."""
        from moai_adk.cli.worktree.exceptions import GitOperationError

        mock_worktree.side_effect = GitOperationError("Git command failed")

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_handles_merge_conflict_error(self, mock_worktree: MagicMock) -> None:
        """Test handling of MergeConflictError."""
        from moai_adk.cli.worktree.exceptions import MergeConflictError

        mock_worktree.side_effect = MergeConflictError("SPEC-001", ["file1.py", "file2.py"])

        result = __main__.main()

        assert result == 1

    @patch("moai_adk.cli.worktree.__main__.worktree")
    def test_handles_registry_inconsistency_error(self, mock_worktree: MagicMock) -> None:
        """Test handling of RegistryInconsistencyError."""
        from moai_adk.cli.worktree.exceptions import RegistryInconsistencyError

        mock_worktree.side_effect = RegistryInconsistencyError("Registry corrupted")

        result = __main__.main()

        assert result == 1


class TestMainModuleImport:
    """Test module import and structure."""

    def test_main_function_is_callable(self) -> None:
        """Test that main function is callable."""
        assert callable(__main__.main)

    def test_main_function_signature(self) -> None:
        """Test that main function has no parameters."""
        import inspect

        sig = inspect.signature(__main__.main)
        # Should have no required parameters
        assert len(sig.parameters) == 0

    def test_module_has_main_guard(self) -> None:
        """Test that module has __name__ == '__main__' guard."""
        # Read the module source
        module_path = __main__.__file__
        with open(module_path) as f:
            source = f.read()

        # Check for __name__ == "__main__" guard
        assert '__name__ == "__main__"' in source
        assert "sys.exit(main())" in source
