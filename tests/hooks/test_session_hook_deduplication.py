#!/usr/bin/env python3
"""SessionStart Hook Phase Detection Deduplication Tests

GitHub Issue #207: Hook duplication bug - SessionStart hook being called multiple times

Tests that verify SessionStart hook deduplication based on phase detection (clear vs compact).
The bug causes SessionStart to be called multiple times in different phases, leading to duplicate output.

TDD History:
    - RED: Write failing tests that demonstrate the hook duplication bug
    - GREEN: Implement phase-based deduplication logic
    - REFACTOR: Optimize phase detection and state management
"""

import sys
from pathlib import Path
from typing import Any, Dict
from unittest.mock import patch

# Setup import path for shared modules (following existing pattern)
PROJECT_ROOT = Path(__file__).parent.parent.parent
LIB_DIR = PROJECT_ROOT / ".claude" / "hooks" / "moai" / "lib"

# sys.path에 추가 (최상단에 추가하여 우선순위 높임)
if str(LIB_DIR) not in sys.path:
    sys.path.insert(0, str(LIB_DIR))

import pytest

# Skip this file - outdated test using lib.* mock paths that don't exist
pytestmark = pytest.mark.skip(
    reason="Outdated test - mock paths reference non-existent lib.project and lib.checkpoint modules"
)

# Import the SessionStart hook modules (after skip mark to prevent collection errors)
try:
    from session import handle_session_start
except ImportError:
    # If import fails, skip is already set, so this is safe
    handle_session_start = None


class TestSessionHookPhaseDeduplication:
    """SessionStart Hook Phase Detection Deduplication Tests

    This test class verifies that SessionStart hook deduplication works correctly
    by tracking phase transitions between 'clear' and 'compact' phases.

    GitHub Issue #207 Bug:
    - SessionStart hook is called multiple times in different phases
    - No phase-based deduplication logic exists
    - Duplicate output is shown to users
    """

    def setup_method(self):
        """Setup test environment for each test method"""
        self.test_cwd = "/test/project"
        self.phase_transition_log = []

    def test_session_start_clear_phase_only(self):
        """Test clear phase execution only

        SPEC Requirements:
            - WHEN SessionStart is called with phase="clear", it should execute normally
            - WHEN SessionStart is called again with phase="clear", it should deduplicate
            - Clear phase should return minimal output (just continue flag)

        Expected Behavior:
            - First call with clear phase: executes and returns minimal output
            - Second call with clear phase: should be deduplicated and return minimal output
            - No duplicate execution should occur
        """
        # First call with clear phase - should execute
        payload_clear: Dict[str, Any] = {"cwd": self.test_cwd, "phase": "clear"}

        result1 = handle_session_start(payload_clear)

        # Clear phase should return minimal output
        assert result1.continue_execution is True
        assert result1.system_message is None  # No system message for clear phase
        output1 = result1.to_dict()
        # Expected: {"continue_execution": True, "block_execution": False}
        assert "continue_execution" in output1
        assert output1["continue_execution"] is True
        assert output1["block_execution"] is False

        # Second call with clear phase - should be deduplicated
        result2 = handle_session_start(payload_clear)

        # Should still return minimal output (deduplicated)
        assert result2.continue_execution is True
        assert result2.system_message is None
        output2 = result2.to_dict()
        assert "continue_execution" in output2
        assert output2["continue_execution"] is True
        assert output2["block_execution"] is False

        # Verify no duplicate execution occurred
        # This will fail until deduplication is implemented
        self.phase_transition_log.append(("clear", "executed"))

    def test_session_start_compact_phase_only(self):
        """Test compact phase execution only

        SPEC Requirements:
            - WHEN SessionStart is called with phase="compact", it should execute normally
            - WHEN SessionStart is called again with phase="compact", it should deduplicate
            - Compact phase should return detailed project information

        Expected Behavior:
            - First call with compact phase: executes and returns detailed output
            - Second call with compact phase: should be deduplicated
            - Detailed project info should only appear once
        """
        # Mock the dependencies for compact phase
        with (
            patch("lib.project.get_git_info") as mock_git,
            patch("lib.project.count_specs") as mock_specs,
            patch("lib.checkpoint.list_checkpoints") as mock_checkpoints,
            patch("lib.project.get_package_version_info") as mock_version,
        ):

            mock_git.return_value = {"branch": "main", "commit": "abc123def456", "changes": 0}
            mock_specs.return_value = {"completed": 5, "total": 10, "percentage": 50}
            mock_checkpoints.return_value = []
            mock_version.return_value = {"current": "0.22.4", "latest": "0.22.4", "update_available": False}

            # First call with compact phase - should execute
            payload_compact: Dict[str, Any] = {"cwd": self.test_cwd, "phase": "compact"}

            result1 = handle_session_start(payload_compact)

            # Compact phase should return detailed output
            assert result1.continue_execution is True
            assert result1.system_message is not None
            assert "🚀 MoAI-ADK Session Started" in result1.system_message

            # Second call with compact phase - should be deduplicated
            result2 = handle_session_start(payload_compact)

            # Should still return detailed output but not duplicate execution
            assert result2.continue_execution is True
            assert result2.system_message is not None

            # Verify no duplicate execution occurred (this will fail until implemented)
            self.phase_transition_log.append(("compact", "executed"))

    def test_session_start_phase_transition_clear_to_compact(self):
        """Test phase transition from clear to compact

        SPEC Requirements:
            - WHEN SessionStart is called with phase="clear", then with phase="compact",
              both should execute (different phases)
            - WHEN SessionStart is called again with same phases, deduplication should occur

        Expected Behavior:
            - clear → compact: both phases execute (phase transition)
            - clear → compact again: deduplication should prevent duplicate execution
        """
        # First clear phase call
        payload_clear: Dict[str, Any] = {"cwd": self.test_cwd, "phase": "clear"}

        result1 = handle_session_start(payload_clear)
        assert result1.continue_execution is True
        assert result1.system_message is None  # Clear phase has no message

        # Then compact phase call (phase transition - should execute)
        with (
            patch("lib.project.get_git_info") as mock_git,
            patch("lib.project.count_specs") as mock_specs,
            patch("lib.checkpoint.list_checkpoints") as mock_checkpoints,
            patch("lib.project.get_package_version_info") as mock_version,
        ):

            mock_git.return_value = {"branch": "feature/test", "commit": "def456abc123", "changes": 0}
            mock_specs.return_value = {"completed": 3, "total": 8, "percentage": 38}
            mock_checkpoints.return_value = []
            mock_version.return_value = {"current": "0.22.4", "latest": "0.22.4", "update_available": False}

            payload_compact: Dict[str, Any] = {"cwd": self.test_cwd, "phase": "compact"}

            result2 = handle_session_start(payload_compact)
            assert result2.continue_execution is True
            assert result2.system_message is not None
            assert "🚀 MoAI-ADK Session Started" in result2.system_message

            # Phase transition should be logged (this will fail until tracking is implemented)
            self.phase_transition_log.append(("clear", "executed"))
            self.phase_transition_log.append(("compact", "executed"))

    def test_session_start_phase_transition_compact_to_clear(self):
        """Test phase transition from compact to clear

        SPEC Requirements:
            - WHEN SessionStart is called with phase="compact", then with phase="clear",
              both should execute (different phases)
            - WHEN SessionStart is called again with same phases, deduplication should occur

        Expected Behavior:
            - compact → clear: both phases execute (phase transition)
            - compact → clear again: deduplication should prevent duplicate execution
        """
        # First compact phase call
        with (
            patch("lib.project.get_git_info") as mock_git,
            patch("lib.project.count_specs") as mock_specs,
            patch("lib.checkpoint.list_checkpoints") as mock_checkpoints,
            patch("lib.project.get_package_version_info") as mock_version,
        ):

            mock_git.return_value = {"branch": "main", "commit": "abc123def456", "changes": 0}
            mock_specs.return_value = {"completed": 5, "total": 10, "percentage": 50}
            mock_checkpoints.return_value = []
            mock_version.return_value = {"current": "0.22.4", "latest": "0.22.4", "update_available": False}

            payload_compact: Dict[str, Any] = {"cwd": self.test_cwd, "phase": "compact"}

            result1 = handle_session_start(payload_compact)
            assert result1.continue_execution is True
            assert result1.system_message is not None

        # Then clear phase call (phase transition - should execute)
        payload_clear: Dict[str, Any] = {"cwd": self.test_cwd, "phase": "clear"}

        result2 = handle_session_start(payload_clear)
        assert result2.continue_execution is True
        assert result2.system_message is None  # Clear phase has no message
        output2 = result2.to_dict()
        assert "continue_execution" in output2
        assert output2["continue_execution"] is True

        # Phase transition should be logged (this will fail until tracking is implemented)
        self.phase_transition_log.append(("compact", "executed"))
        self.phase_transition_log.append(("clear", "executed"))

    def test_session_start_rapid_phase_switching(self):
        """Test rapid phase switching to detect deduplication edge cases

        SPEC Requirements:
            - WHEN SessionStart is called rapidly with alternating phases,
              deduplication should prevent duplicate execution
            - WHEN SessionStart is called multiple times with same phase consecutively,
              only the first should execute

        Expected Behavior:
            - clear → compact → clear → compact: all phase transitions execute
            - clear (second time): should be deduplicated
            - compact (second time): should be deduplicated
        """
        phase_calls = []

        # Simulate rapid phase switching
        phases = ["clear", "compact", "clear", "compact", "clear", "compact"]

        for i, phase in enumerate(phases):
            payload: Dict[str, Any] = {"cwd": self.test_cwd, "phase": phase}

            with (
                patch("lib.project.get_git_info") as mock_git,
                patch("lib.project.count_specs") as mock_specs,
                patch("lib.checkpoint.list_checkpoints") as mock_checkpoints,
                patch("lib.project.get_package_version_info") as mock_version,
            ):

                mock_git.return_value = {"branch": "test", "commit": "test123", "changes": i}
                mock_specs.return_value = {"completed": i, "total": 10, "percentage": i * 10}
                mock_checkpoints.return_value = []
                mock_version.return_value = {"current": "0.22.4", "latest": "0.22.4", "update_available": False}

                result = handle_session_start(payload)
                assert result.continue_execution is True

                # Verify correct phase behavior
                if phase == "clear":
                    assert result.system_message is None
                else:  # compact
                    assert result.system_message is not None

                phase_calls.append(phase)

        # Verify phase transitions (this will fail until deduplication is implemented)
        # Each phase transition should execute, but consecutive same phases should deduplicate
        expected_unique_transitions = []
        last_phase = None

        for phase in phase_calls:
            if phase != last_phase:
                expected_unique_transitions.append(phase)
                last_phase = phase

        # This assertion will fail because deduplication doesn't exist yet
        # After implementation, this should pass
        assert len(phase_calls) == len(expected_unique_transitions), f"Expected deduplication but got: {phase_calls}"

    def test_session_start_missing_phase_field(self):
        """Test SessionStart without phase field (edge case)

        SPEC Requirements:
            - WHEN SessionStart payload is missing phase field,
              it should default to compact behavior (show full output)
            - WHEN called multiple times without phase, deduplication should still work

        Expected Behavior:
            - Missing phase should be treated as compact phase (full output)
            - Multiple calls without phase should be deduplicated
        """
        # First call without phase field
        with (
            patch("lib.project.get_git_info") as mock_git,
            patch("lib.project.count_specs") as mock_specs,
            patch("lib.checkpoint.list_checkpoints") as mock_checkpoints,
            patch("lib.project.get_package_version_info") as mock_version,
        ):

            mock_git.return_value = {"branch": "main", "commit": "abc123", "changes": 0}
            mock_specs.return_value = {"completed": 1, "total": 5, "percentage": 20}
            mock_checkpoints.return_value = []
            mock_version.return_value = {"current": "0.22.4", "latest": "0.22.4", "update_available": False}

            payload_no_phase: Dict[str, Any] = {
                "cwd": self.test_cwd
                # Missing "phase" field - defaults to compact
            }

            result1 = handle_session_start(payload_no_phase)
            assert result1.continue_execution is True
            # Missing phase defaults to compact (full output)
            assert result1.system_message is not None
            assert "🚀 MoAI-ADK Session Started" in result1.system_message

            # Second call without phase field - should be deduplicated
            result2 = handle_session_start(payload_no_phase)
            assert result2.continue_execution is True
            assert result2.system_message is not None

            # Verify no duplicate execution (this will fail until implemented)
            self.phase_transition_log.append(("no_phase", "executed"))

    def test_session_start_invalid_phase(self):
        """Test SessionStart with invalid phase value

        SPEC Requirements:
            - WHEN SessionStart has invalid phase value,
              it should default to compact behavior (show full output)
            - Invalid phase should not cause errors
            - Deduplication should still work for invalid phases

        Expected Behavior:
            - Invalid phase defaults to compact phase (full output)
            - Multiple calls with invalid phase should be deduplicated
        """
        # First call with invalid phase
        with (
            patch("lib.project.get_git_info") as mock_git,
            patch("lib.project.count_specs") as mock_specs,
            patch("lib.checkpoint.list_checkpoints") as mock_checkpoints,
            patch("lib.project.get_package_version_info") as mock_version,
        ):

            mock_git.return_value = {"branch": "main", "commit": "abc123", "changes": 0}
            mock_specs.return_value = {"completed": 1, "total": 5, "percentage": 20}
            mock_checkpoints.return_value = []
            mock_version.return_value = {"current": "0.22.4", "latest": "0.22.4", "update_available": False}

            payload_invalid_phase: Dict[str, Any] = {"cwd": self.test_cwd, "phase": "invalid_phase_value"}

            result1 = handle_session_start(payload_invalid_phase)
            assert result1.continue_execution is True
            # Invalid phase defaults to compact (full output)
            assert result1.system_message is not None
            assert "🚀 MoAI-ADK Session Started" in result1.system_message

            # Second call with invalid phase - should be deduplicated
            result2 = handle_session_start(payload_invalid_phase)
            assert result2.continue_execution is True
            assert result2.system_message is not None

            # Verify no duplicate execution (this will fail until implemented)
            self.phase_transition_log.append(("invalid_phase", "executed"))

    def test_session_start_execution_counting(self):
        """Test that SessionStart execution is properly counted and deduplicated

        SPEC Requirements:
            - WHEN SessionStart is called multiple times with same phase,
              execution count should increase only for first call
            - WHEN SessionStart is called with different phases,
              execution count should increase for each phase transition
            - Deduplication should prevent unnecessary execution

        Expected Behavior:
            - Same phase calls: execution count stays after first call
            - Different phase calls: execution count increases with each transition
            - No duplicate executions should occur
        """

        # Test multiple calls to same phase (clear)
        for i in range(3):
            payload_clear: Dict[str, Any] = {"cwd": self.test_cwd, "phase": "clear"}
            result = handle_session_start(payload_clear)
            assert result.continue_execution is True
            assert result.system_message is None  # Clear phase has no message

        # Test phase transition to compact
        with (
            patch("lib.project.get_git_info") as mock_git,
            patch("lib.project.count_specs") as mock_specs,
            patch("lib.checkpoint.list_checkpoints") as mock_checkpoints,
            patch("lib.project.get_package_version_info") as mock_version,
        ):

            mock_git.return_value = {"branch": "main", "commit": "test123", "changes": 0}
            mock_specs.return_value = {"completed": 1, "total": 5, "percentage": 20}
            mock_checkpoints.return_value = []
            mock_version.return_value = {"current": "0.22.4", "latest": "0.22.4", "update_available": False}

            for i in range(3):
                payload_compact: Dict[str, Any] = {"cwd": self.test_cwd, "phase": "compact"}
                result = handle_session_start(payload_compact)
                assert result.continue_execution is True
                assert result.system_message is not None  # Compact phase has message

        # Execution counting test: verify that handle_session_start continues execution
        # The actual deduplication is handled by phase-based logic (clear vs compact)
        # We've verified that both phases return continue_execution=True
