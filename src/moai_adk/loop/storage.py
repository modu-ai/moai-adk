# Loop Storage
"""Persistent storage for loop state management.

Provides file-based storage for LoopState objects, enabling persistence
across sessions and recovery of loop execution state.
"""

from __future__ import annotations

import json
from datetime import datetime
from pathlib import Path
from typing import Any

from moai_adk.loop.state import (
    ASTIssueSnapshot,
    DiagnosticSnapshot,
    LoopState,
    LoopStatus,
)


class LoopStorage:
    """File-based storage for loop states.

    Manages persistence of LoopState objects to JSON files, enabling
    state recovery and history tracking across sessions.

    Attributes:
        storage_dir: Path to the storage directory.
    """

    def __init__(self, storage_dir: str = ".moai/loop"):
        """Initialize the loop storage.

        Creates the storage directory if it doesn't exist.

        Args:
            storage_dir: Path to the storage directory.
        """
        self.storage_dir = Path(storage_dir)
        self.storage_dir.mkdir(parents=True, exist_ok=True)

    def save_state(self, state: LoopState) -> None:
        """Save a loop state to storage.

        Serializes the LoopState to JSON and writes it to a file
        named after the loop_id.

        Args:
            state: The LoopState to save.
        """
        state_file = self.storage_dir / f"{state.loop_id}.json"
        data = self._serialize_state(state)
        with open(state_file, "w", encoding="utf-8") as f:
            json.dump(data, f, indent=2, default=str)

    def load_state(self, loop_id: str) -> LoopState | None:
        """Load a loop state from storage.

        Reads and deserializes a LoopState from the JSON file
        corresponding to the loop_id.

        Args:
            loop_id: The ID of the loop to load.

        Returns:
            The loaded LoopState, or None if not found.
        """
        state_file = self.storage_dir / f"{loop_id}.json"
        if not state_file.exists():
            return None

        with open(state_file, "r", encoding="utf-8") as f:
            data = json.load(f)

        return self._deserialize_state(data)

    def get_active_loop_id(self) -> str | None:
        """Get the ID of the currently active loop.

        Scans all stored loops and returns the ID of the most recently
        updated loop that is in an active state (RUNNING or PAUSED).

        Returns:
            The loop_id of the active loop, or None if no active loop.
        """
        active_loops: list[tuple[str, datetime]] = []

        for loop_id in self.list_loops():
            state = self.load_state(loop_id)
            if state is not None and state.is_active():
                active_loops.append((loop_id, state.updated_at))

        if not active_loops:
            return None

        # Return the most recently updated active loop
        active_loops.sort(key=lambda x: x[1], reverse=True)
        return active_loops[0][0]

    def list_loops(self) -> list[str]:
        """List all stored loop IDs.

        Scans the storage directory for loop state files and
        returns their IDs.

        Returns:
            List of loop IDs.
        """
        loop_ids = []
        for file_path in self.storage_dir.glob("*.json"):
            loop_ids.append(file_path.stem)
        return loop_ids

    def delete_state(self, loop_id: str) -> bool:
        """Delete a loop state from storage.

        Removes the JSON file corresponding to the loop_id.

        Args:
            loop_id: The ID of the loop to delete.

        Returns:
            True if the state was deleted, False if it didn't exist.
        """
        state_file = self.storage_dir / f"{loop_id}.json"
        if not state_file.exists():
            return False

        state_file.unlink()
        return True

    def _serialize_state(self, state: LoopState) -> dict[str, Any]:
        """Serialize a LoopState to a dictionary.

        Converts the LoopState and its nested objects to a JSON-serializable
        dictionary format.

        Args:
            state: The LoopState to serialize.

        Returns:
            Dictionary representation of the state.
        """
        return {
            "loop_id": state.loop_id,
            "promise": state.promise,
            "current_iteration": state.current_iteration,
            "max_iterations": state.max_iterations,
            "status": state.status.value,
            "created_at": state.created_at.isoformat(),
            "updated_at": state.updated_at.isoformat(),
            "diagnostics_history": [self._serialize_diagnostic_snapshot(s) for s in state.diagnostics_history],
            "ast_issues_history": [self._serialize_ast_snapshot(s) for s in state.ast_issues_history],
        }

    def _serialize_diagnostic_snapshot(self, snapshot: DiagnosticSnapshot) -> dict[str, Any]:
        """Serialize a DiagnosticSnapshot to a dictionary."""
        return {
            "timestamp": snapshot.timestamp.isoformat(),
            "error_count": snapshot.error_count,
            "warning_count": snapshot.warning_count,
            "info_count": snapshot.info_count,
            "files_affected": snapshot.files_affected,
        }

    def _serialize_ast_snapshot(self, snapshot: ASTIssueSnapshot) -> dict[str, Any]:
        """Serialize an ASTIssueSnapshot to a dictionary."""
        return {
            "timestamp": snapshot.timestamp.isoformat(),
            "total_issues": snapshot.total_issues,
            "by_severity": snapshot.by_severity,
            "by_rule": snapshot.by_rule,
            "files_affected": snapshot.files_affected,
        }

    def _deserialize_state(self, data: dict[str, Any]) -> LoopState:
        """Deserialize a dictionary to a LoopState.

        Converts a JSON-deserialized dictionary back to a LoopState
        with proper typing.

        Args:
            data: Dictionary representation of the state.

        Returns:
            The deserialized LoopState.
        """
        return LoopState(
            loop_id=data["loop_id"],
            promise=data["promise"],
            current_iteration=data["current_iteration"],
            max_iterations=data["max_iterations"],
            status=LoopStatus(data["status"]),
            created_at=datetime.fromisoformat(data["created_at"]),
            updated_at=datetime.fromisoformat(data["updated_at"]),
            diagnostics_history=[self._deserialize_diagnostic_snapshot(s) for s in data.get("diagnostics_history", [])],
            ast_issues_history=[self._deserialize_ast_snapshot(s) for s in data.get("ast_issues_history", [])],
        )

    def _deserialize_diagnostic_snapshot(self, data: dict[str, Any]) -> DiagnosticSnapshot:
        """Deserialize a dictionary to a DiagnosticSnapshot."""
        return DiagnosticSnapshot(
            timestamp=datetime.fromisoformat(data["timestamp"]),
            error_count=data["error_count"],
            warning_count=data["warning_count"],
            info_count=data["info_count"],
            files_affected=data["files_affected"],
        )

    def _deserialize_ast_snapshot(self, data: dict[str, Any]) -> ASTIssueSnapshot:
        """Deserialize a dictionary to an ASTIssueSnapshot."""
        return ASTIssueSnapshot(
            timestamp=datetime.fromisoformat(data["timestamp"]),
            total_issues=data["total_issues"],
            by_severity=data["by_severity"],
            by_rule=data["by_rule"],
            files_affected=data["files_affected"],
        )
