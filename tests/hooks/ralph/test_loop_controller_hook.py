"""Unit tests for stop__loop_controller.py hook.

Tests the loop controller hook that manages the Ralph feedback loop.
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


class TestLoopState:
    """Test LoopState dataclass."""

    def test_default_state(self):
        """Default state should be inactive."""
        from stop__loop_controller import LoopState

        state = LoopState()
        assert state.active is False
        assert state.iteration == 0
        assert state.max_iterations == 10

    def test_to_dict(self):
        """State should serialize to dictionary."""
        from stop__loop_controller import LoopState

        state = LoopState(active=True, iteration=3, max_iterations=5)
        data = state.to_dict()

        assert data["active"] is True
        assert data["iteration"] == 3
        assert data["max_iterations"] == 5

    def test_from_dict(self):
        """State should deserialize from dictionary."""
        from stop__loop_controller import LoopState

        data = {
            "active": True,
            "iteration": 5,
            "max_iterations": 10,
            "last_error_count": 2,
            "last_warning_count": 3,
            "files_modified": ["file1.py", "file2.py"],
            "start_time": 1234567890.0,
            "completion_reason": None,
        }

        state = LoopState.from_dict(data)
        assert state.active is True
        assert state.iteration == 5
        assert state.last_error_count == 2
        assert state.files_modified == ["file1.py", "file2.py"]

    def test_from_dict_with_defaults(self):
        """Missing fields should use defaults."""
        from stop__loop_controller import LoopState

        data = {"active": True}
        state = LoopState.from_dict(data)

        assert state.active is True
        assert state.iteration == 0
        assert state.max_iterations == 10


class TestCompletionStatus:
    """Test CompletionStatus dataclass."""

    def test_default_status(self):
        """Default status should have conditions unmet."""
        from stop__loop_controller import CompletionStatus

        status = CompletionStatus()
        assert status.zero_errors is False
        assert status.tests_pass is False
        assert status.all_conditions_met is False

    def test_all_conditions_met(self):
        """Status with all conditions should be complete."""
        from stop__loop_controller import CompletionStatus

        status = CompletionStatus(
            zero_errors=True,
            zero_warnings=True,
            tests_pass=True,
            coverage_met=True,
            all_conditions_met=True,
        )

        assert status.all_conditions_met is True


class TestLoadLoopState:
    """Test load_loop_state function."""

    def test_load_from_env(self):
        """State should load from environment variables."""
        from stop__loop_controller import load_loop_state

        with patch.dict(
            "os.environ",
            {"MOAI_LOOP_ACTIVE": "true", "MOAI_LOOP_ITERATION": "3"},
        ):
            state = load_loop_state()
            assert state.active is True
            assert state.iteration == 3

    def test_load_from_file(self, tmp_path):
        """State should load from file when env not set."""
        from stop__loop_controller import load_loop_state

        # Create state file
        cache_dir = tmp_path / ".moai" / "cache"
        cache_dir.mkdir(parents=True)
        state_file = cache_dir / ".moai_loop_state.json"
        state_file.write_text(
            json.dumps(
                {
                    "active": True,
                    "iteration": 5,
                    "max_iterations": 10,
                }
            )
        )

        with patch("stop__loop_controller.get_state_file_path") as mock_path:
            mock_path.return_value = state_file
            with patch.dict("os.environ", {}, clear=False):
                # Clear loop env vars
                import os

                os.environ.pop("MOAI_LOOP_ACTIVE", None)
                os.environ.pop("MOAI_LOOP_ITERATION", None)

                state = load_loop_state()
                assert state.active is True
                assert state.iteration == 5

    def test_load_default_when_no_state(self):
        """Default state when no env or file."""
        from stop__loop_controller import load_loop_state

        with patch("stop__loop_controller.get_state_file_path") as mock_path:
            mock_path.return_value = Path("/nonexistent/path/.moai_loop_state.json")
            with patch.dict("os.environ", {}, clear=False):
                import os

                os.environ.pop("MOAI_LOOP_ACTIVE", None)
                os.environ.pop("MOAI_LOOP_ITERATION", None)

                state = load_loop_state()
                assert state.active is False


class TestSaveAndClearLoopState:
    """Test save_loop_state and clear_loop_state functions."""

    def test_save_state(self, tmp_path):
        """State should be saved to file."""
        from stop__loop_controller import LoopState, save_loop_state

        state_file = tmp_path / ".moai" / "cache" / ".moai_loop_state.json"

        with patch("stop__loop_controller.get_state_file_path") as mock_path:
            mock_path.return_value = state_file

            state = LoopState(active=True, iteration=3)
            save_loop_state(state)

            assert state_file.exists()
            data = json.loads(state_file.read_text())
            assert data["active"] is True
            assert data["iteration"] == 3

    def test_clear_state(self, tmp_path):
        """State file should be removed."""
        from stop__loop_controller import clear_loop_state

        state_file = tmp_path / ".moai_loop_state.json"
        state_file.write_text("{}")

        with patch("stop__loop_controller.get_state_file_path") as mock_path:
            mock_path.return_value = state_file

            clear_loop_state()
            assert not state_file.exists()


class TestLoadRalphConfig:
    """Test load_ralph_config function."""

    def test_default_config(self):
        """Default config should be returned when file not found."""
        from stop__loop_controller import load_ralph_config

        with patch("stop__loop_controller.get_project_dir") as mock_dir:
            mock_dir.return_value = Path("/nonexistent/path")
            config = load_ralph_config()

            assert config["enabled"] is True
            assert config["loop"]["max_iterations"] == 10
            assert config["loop"]["completion"]["zero_errors"] is True

    def test_config_from_file(self, tmp_path):
        """Config should be loaded from file when available."""
        from stop__loop_controller import load_ralph_config

        # Create config file
        config_dir = tmp_path / ".moai" / "config" / "sections"
        config_dir.mkdir(parents=True)
        config_file = config_dir / "ralph.yaml"
        config_file.write_text("""
ralph:
  enabled: true
  loop:
    max_iterations: 5
    completion:
      zero_errors: true
      zero_warnings: true
""")

        with patch("stop__loop_controller.get_project_dir") as mock_dir:
            mock_dir.return_value = tmp_path
            config = load_ralph_config()

            assert config["loop"]["max_iterations"] == 5


class TestCheckCompletionConditions:
    """Test check_completion_conditions function."""

    def test_all_conditions_met(self):
        """All conditions met should return complete status."""
        from stop__loop_controller import check_completion_conditions

        config: dict[str, Any] = {
            "loop": {
                "completion": {
                    "zero_errors": True,
                    "zero_warnings": False,
                    "tests_pass": False,
                    "coverage_threshold": 0,
                }
            }
        }

        with patch("stop__loop_controller.check_lsp_errors") as mock_errors:
            mock_errors.return_value = (0, 0)

            status = check_completion_conditions(config)
            assert status.zero_errors is True
            assert status.all_conditions_met is True

    def test_errors_present(self):
        """Errors present should fail zero_errors condition."""
        from stop__loop_controller import check_completion_conditions

        config: dict[str, Any] = {
            "loop": {
                "completion": {
                    "zero_errors": True,
                    "zero_warnings": False,
                    "tests_pass": False,
                    "coverage_threshold": 0,
                }
            }
        }

        with patch("stop__loop_controller.check_lsp_errors") as mock_errors:
            mock_errors.return_value = (2, 3)

            status = check_completion_conditions(config)
            assert status.zero_errors is False
            assert status.all_conditions_met is False


class TestFormatLoopOutput:
    """Test format_loop_output function."""

    def test_format_complete(self):
        """Format output for completed loop."""
        from stop__loop_controller import CompletionStatus, LoopState, format_loop_output

        state = LoopState(active=False, iteration=5)
        status = CompletionStatus(
            zero_errors=True,
            tests_pass=True,
            all_conditions_met=True,
            details={"errors": 0, "warnings": 0},
        )

        output = format_loop_output(state, status, "COMPLETE")
        assert "COMPLETE" in output
        assert "Errors: 0" in output

    def test_format_continue(self):
        """Format output for continuing loop."""
        from stop__loop_controller import CompletionStatus, LoopState, format_loop_output

        state = LoopState(active=True, iteration=3, max_iterations=10)
        status = CompletionStatus(
            zero_errors=False,
            all_conditions_met=False,
            details={"errors": 2, "warnings": 1},
        )

        output = format_loop_output(state, status, "CONTINUE")
        assert "CONTINUE" in output
        assert "3/10" in output
        assert "Errors: 2" in output


class TestMainFunction:
    """Test main hook entry point."""

    def test_disabled_via_env(self):
        """Hook should exit when disabled via environment."""
        import stop__loop_controller

        with patch.dict("os.environ", {"MOAI_DISABLE_LOOP_CONTROLLER": "true"}):
            with pytest.raises(SystemExit) as exc_info:
                stop__loop_controller.main()
            assert exc_info.value.code == 0

    def test_inactive_loop_exits(self):
        """Inactive loop should exit immediately."""
        import stop__loop_controller

        with patch.dict("os.environ", {}, clear=False):
            import os

            os.environ.pop("MOAI_DISABLE_LOOP_CONTROLLER", None)
            os.environ.pop("MOAI_LOOP_ACTIVE", None)

            with patch("stop__loop_controller.load_loop_state") as mock_load:
                from stop__loop_controller import LoopState

                mock_load.return_value = LoopState(active=False)

                with patch("sys.stdin", StringIO("{}")):
                    with pytest.raises(SystemExit) as exc_info:
                        stop__loop_controller.main()
                    assert exc_info.value.code == 0

    def test_complete_loop_exits_zero(self):
        """Completed loop should exit with code 0."""
        import stop__loop_controller

        with patch.dict("os.environ", {}, clear=False):
            import os

            os.environ.pop("MOAI_DISABLE_LOOP_CONTROLLER", None)

            with patch("stop__loop_controller.load_loop_state") as mock_load:
                from stop__loop_controller import LoopState

                mock_load.return_value = LoopState(active=True, iteration=1)

                with patch("stop__loop_controller.load_ralph_config") as mock_config:
                    mock_config.return_value = {
                        "enabled": True,
                        "loop": {
                            "completion": {
                                "zero_errors": True,
                                "zero_warnings": False,
                                "tests_pass": False,
                                "coverage_threshold": 0,
                            }
                        },
                        "hooks": {"stop_loop_controller": {"enabled": True, "check_completion": True}},
                    }

                    with patch("stop__loop_controller.check_completion_conditions") as mock_check:
                        from stop__loop_controller import CompletionStatus

                        mock_check.return_value = CompletionStatus(
                            zero_errors=True,
                            all_conditions_met=True,
                            details={"errors": 0},
                        )

                        with patch("stop__loop_controller.clear_loop_state"):
                            with patch("sys.stdin", StringIO("{}")):
                                with pytest.raises(SystemExit) as exc_info:
                                    stop__loop_controller.main()
                                assert exc_info.value.code == 0

    def test_continue_loop_exits_one(self):
        """Continuing loop should exit with code 1."""
        import stop__loop_controller

        with patch.dict("os.environ", {}, clear=False):
            import os

            os.environ.pop("MOAI_DISABLE_LOOP_CONTROLLER", None)

            with patch("stop__loop_controller.load_loop_state") as mock_load:
                from stop__loop_controller import LoopState

                mock_load.return_value = LoopState(active=True, iteration=1, max_iterations=10)

                with patch("stop__loop_controller.load_ralph_config") as mock_config:
                    mock_config.return_value = {
                        "enabled": True,
                        "loop": {
                            "completion": {
                                "zero_errors": True,
                                "zero_warnings": False,
                                "tests_pass": False,
                                "coverage_threshold": 0,
                            }
                        },
                        "hooks": {"stop_loop_controller": {"enabled": True, "check_completion": True}},
                    }

                    with patch("stop__loop_controller.check_completion_conditions") as mock_check:
                        from stop__loop_controller import CompletionStatus

                        mock_check.return_value = CompletionStatus(
                            zero_errors=False,
                            all_conditions_met=False,
                            details={"errors": 2},
                        )

                        with patch("stop__loop_controller.save_loop_state"):
                            with patch("sys.stdin", StringIO("{}")):
                                with pytest.raises(SystemExit) as exc_info:
                                    stop__loop_controller.main()
                                assert exc_info.value.code == 1
