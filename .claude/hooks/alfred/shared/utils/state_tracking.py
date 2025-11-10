#!/usr/bin/env python3
"""State tracking utilities for hook execution and deduplication

Provides centralized state management for hook execution tracking,
command deduplication, and duplicate prevention.
"""

import json
import threading
import time
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, Any, Optional, List
import uuid

from core import HookPayload, HookResult


class HookStateManager:
    """Centralized state management for hook execution tracking and deduplication

    Handles:
    - Hook execution counting and tracking
    - Phase-based deduplication for SessionStart
    - Command deduplication within time windows
    - Thread-safe state operations
    - Persistent state storage
    """

    # Time windows for deduplication (in seconds)
    COMMAND_DEDUPE_WINDOW = 3.0
    HOOK_DEDUPE_WINDOW = 1.0

    def __init__(self, cwd: str):
        """Initialize state manager for given working directory"""
        self.cwd = cwd

        # Try to use .moai/memory directory, fallback to temp directory if not writable
        try:
            self.state_dir = Path(cwd) / ".moai" / "memory"
            self.state_dir.mkdir(parents=True, exist_ok=True)
            # Test if directory is writable
            test_file = self.state_dir / ".test_write"
            test_file.touch()
            test_file.unlink()
        except (OSError, PermissionError):
            # Fallback to temporary directory
            import tempfile
            temp_dir = Path(tempfile.gettempdir()) / "alfred_hooks" / cwd.replace("/", "_")
            self.state_dir = temp_dir
            self.state_dir.mkdir(parents=True, exist_ok=True)

        # Thread safety
        self._lock = threading.RLock()

        # State files
        self.hook_state_file = self.state_dir / "hook_execution_state.json"
        self.command_state_file = self.state_dir / "command_execution_state.json"

        # In-memory cache for performance
        self._hook_state_cache: Optional[Dict[str, Any]] = None
        self._command_state_cache: Optional[Dict[str, Any]] = None
        self._cache_timestamp = 0

    def _load_hook_state(self) -> Dict[str, Any]:
        """Load hook execution state with caching"""
        current_time = time.time()

        # Use cache if recent (within 5 seconds)
        if (self._hook_state_cache and
            current_time - self._cache_timestamp < 5.0):
            return self._hook_state_cache

        try:
            if self.hook_state_file.exists():
                with open(self.hook_state_file, "r", encoding="utf-8") as f:
                    state = json.load(f)
                self._hook_state_cache = state
                self._cache_timestamp = current_time
                return state
        except Exception:
            pass

        # Default state structure
        default_state = {}
        self._hook_state_cache = default_state
        self._cache_timestamp = current_time
        return default_state

    def _save_hook_state(self, state: Dict[str, Any]) -> None:
        """Save hook execution state"""
        try:
            with open(self.hook_state_file, "w", encoding="utf-8") as f:
                json.dump(state, f, indent=2)
            self._hook_state_cache = state
            self._cache_timestamp = time.time()
        except Exception:
            pass

    def _load_command_state(self) -> Dict[str, Any]:
        """Load command execution state with caching"""
        current_time = time.time()

        # Use cache if recent (within 5 seconds)
        if (self._command_state_cache and
            current_time - self._cache_timestamp < 5.0):
            return self._command_state_cache

        try:
            if self.command_state_file.exists():
                with open(self.command_state_file, "r", encoding="utf-8") as f:
                    state = json.load(f)
                self._command_state_cache = state
                self._cache_timestamp = current_time
                return state
        except Exception:
            pass

        # Default state structure
        default_state = {
            "last_command": None,
            "last_timestamp": None,
            "is_running": False,
            "execution_count": 0,
            "duplicate_count": 0
        }
        self._command_state_cache = default_state
        self._cache_timestamp = current_time
        return default_state

    def _save_command_state(self, state: Dict[str, Any]) -> None:
        """Save command execution state"""
        try:
            with open(self.command_state_file, "w", encoding="utf-8") as f:
                json.dump(state, f, indent=2)
            self._command_state_cache = state
            self._cache_timestamp = time.time()
        except Exception:
            pass

    def track_hook_execution(self, hook_name: str, phase: str = None) -> Dict[str, Any]:
        """Track hook execution and return execution information

        Args:
            hook_name: Name of the hook being executed
            phase: Optional phase for phase-based deduplication

        Returns:
            Dictionary with execution info including:
            - executed: Whether the hook was actually executed
            - execution_count: Total execution count
            - duplicate: Whether this was a duplicate execution
            - execution_id: Unique identifier for this execution
        """
        with self._lock:
            state = self._load_hook_state()

            # Initialize hook state if not exists
            if hook_name not in state:
                state[hook_name] = {
                    "count": 0,
                    "last_execution": 0,
                    "last_phase": None,
                    "executions": []
                }

            current_time = time.time()
            hook_state = state[hook_name]

            # Check for deduplication
            is_duplicate = False
            execution_id = str(uuid.uuid4())

            # Phase-based deduplication for SessionStart
            if hook_name == "SessionStart" and phase:
                # Phase transitions are allowed (clear->compact or compact->clear)
                if (phase == hook_state.get("last_phase") and
                    current_time - hook_state["last_execution"] < self.HOOK_DEDUPE_WINDOW):
                    # Same phase within time window - deduplicate
                    is_duplicate = True
                else:
                    # Different phase or time window expired - execute
                    pass
            else:
                # Regular deduplication based on time window
                if (current_time - hook_state["last_execution"] < self.HOOK_DEDUPE_WINDOW):
                    is_duplicate = True

            # Update state only if not duplicate
            if not is_duplicate:
                hook_state["count"] += 1
                hook_state["last_execution"] = current_time
                hook_state["last_phase"] = phase
                hook_state["executions"].append({
                    "timestamp": current_time,
                    "phase": phase,
                    "execution_id": execution_id
                })

                # Keep only recent executions (cleanup)
                recent_executions = [
                    e for e in hook_state["executions"]
                    if current_time - e["timestamp"] < 3600  # 1 hour
                ]
                if len(recent_executions) != len(hook_state["executions"]):
                    hook_state["executions"] = recent_executions

                self._save_hook_state(state)

            return {
                "hook_name": hook_name,
                "executed": not is_duplicate,
                "execution_count": hook_state["count"],
                "duplicate": is_duplicate,
                "execution_id": execution_id,
                "phase": phase,
                "timestamp": current_time
            }

    def deduplicate_command(self, command: str) -> Dict[str, Any]:
        """Check and deduplicate command execution within time window

        Args:
            command: Command string to check for deduplication

        Returns:
            Dictionary with deduplication info including:
            - executed: Whether the command should execute
            - duplicate: Whether this was a duplicate
            - reason: Reason for deduplication decision
            - execution_count: Total execution count
        """
        with self._lock:
            state = self._load_command_state()

            current_time = time.time()

            # Check if command is an Alfred command (only deduplicate these)
            if not command or not command.startswith("/alfred:"):
                return {
                    "command": command,
                    "executed": True,
                    "duplicate": False,
                    "reason": "non-alfred command",
                    "execution_count": state["execution_count"]
                }

            # Check for duplicate within time window
            last_cmd = state.get("last_command")
            last_timestamp = state.get("last_timestamp")

            if (last_cmd and last_timestamp and
                command == last_cmd and
                current_time - last_timestamp < self.COMMAND_DEDUPE_WINDOW):

                # Duplicate detected
                state["duplicate_count"] += 1
                state["is_running"] = True  # Mark as running to prevent further duplicates
                state["duplicate_timestamp"] = current_time.isoformat()
                self._save_command_state(state)

                return {
                    "command": command,
                    "executed": True,  # Allow execution but mark as duplicate
                    "duplicate": True,
                    "reason": f"duplicate within {self.COMMAND_DEDUPE_WINDOW}s window",
                    "execution_count": state["execution_count"],
                    "duplicate_count": state["duplicate_count"]
                }

            # Not a duplicate - update state and execute
            state["last_command"] = command
            state["last_timestamp"] = current_time
            state["is_running"] = True
            state["execution_count"] += 1
            self._save_command_state(state)

            return {
                "command": command,
                "executed": True,
                "duplicate": False,
                "reason": "normal execution",
                "execution_count": state["execution_count"]
            }

    def mark_command_complete(self, command: str = None) -> None:
        """Mark command execution as complete

        Args:
            command: Optional command that completed
        """
        with self._lock:
            state = self._load_command_state()
            state["is_running"] = False
            state["last_timestamp"] = time.time()
            if command:
                state["last_command"] = command
            self._save_command_state(state)

    def get_hook_execution_count(self, hook_name: str) -> int:
        """Get total execution count for a hook"""
        state = self._load_hook_state()
        return state.get(hook_name, {}).get("count", 0)

    def get_command_execution_count(self) -> int:
        """Get total command execution count"""
        state = self._load_command_state()
        return state.get("execution_count", 0)

    def cleanup_old_states(self, max_age_hours: int = 24) -> None:
        """Clean up old state entries to prevent state file bloat"""
        current_time = time.time()
        max_age_seconds = max_age_hours * 3600

        # Clean up hook state
        with self._lock:
            hook_state = self._load_hook_state()

            # Clean up old hook executions
            for hook_name in list(hook_state.keys()):
                hook_data = hook_state[hook_name]
                if "executions" in hook_data:
                    recent_executions = [
                        e for e in hook_data["executions"]
                        if current_time - e["timestamp"] < max_age_seconds
                    ]
                    if len(recent_executions) != len(hook_data["executions"]):
                        hook_data["executions"] = recent_executions

                # Remove hooks with no recent executions
                if (hook_data.get("last_execution", 0) < current_time - max_age_seconds):
                    del hook_state[hook_name]

            self._save_hook_state(hook_state)

            # Clean up command state
            command_state = self._load_command_state()
            if (command_state.get("last_timestamp", 0) < current_time - max_age_seconds):
                # Reset command state if too old
                command_state.update({
                    "last_command": None,
                    "last_timestamp": None,
                    "is_running": False,
                    "execution_count": 0,
                    "duplicate_count": 0
                })
                self._save_command_state(command_state)


# Global state manager instances (per-CWD)
_state_managers: Dict[str, HookStateManager] = {}
_state_manager_lock = threading.RLock()


def get_state_manager(cwd: str) -> HookStateManager:
    """Get or create state manager for given working directory"""
    with _state_manager_lock:
        if cwd not in _state_managers:
            _state_managers[cwd] = HookStateManager(cwd)
        return _state_managers[cwd]


def track_hook_execution(hook_name: str, cwd: str, phase: str = None) -> Dict[str, Any]:
    """Convenience function to track hook execution"""
    manager = get_state_manager(cwd)
    return manager.track_hook_execution(hook_name, phase)


def deduplicate_command(command: str, cwd: str) -> Dict[str, Any]:
    """Convenience function to deduplicate command"""
    manager = get_state_manager(cwd)
    return manager.deduplicate_command(command)


def mark_command_complete(command: str = None, cwd: str = None) -> None:
    """Convenience function to mark command complete"""
    if not cwd:
        cwd = "."
    manager = get_state_manager(cwd)
    manager.mark_command_complete(command)