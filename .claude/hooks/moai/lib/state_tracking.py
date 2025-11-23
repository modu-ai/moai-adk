#!/usr/bin/env python3
"""State tracking utilities for hook execution and deduplication

Provides centralized state management for hook execution tracking,
command deduplication, and duplicate prevention.
"""

import atexit
import json
import threading
import time
import uuid
import weakref
from datetime import datetime
from typing import Any, Dict, Optional

from lib import (
    ExecutionResult,
    HookConfiguration,
    configure_logging,
    get_logger,
    get_performance_metrics,
    record_cache_hit,
    record_cache_miss,
    record_execution_metrics,
)


class SingletonMeta(type):
    """Thread-safe Singleton metaclass with cleanup support"""

    _instances: Dict[type, Any] = {}
    _lock = threading.Lock()

    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            with cls._lock:
                if cls not in cls._instances:
                    instance = super().__call__(*args, **kwargs)
                    cls._instances[cls] = instance
        return cls._instances[cls]

    def cleanup_all_instances(cls):
        """Cleanup all singleton instances"""
        with cls._lock:
            for instance in cls._instances.values():
                if hasattr(instance, 'cleanup'):
                    try:
                        instance.cleanup()
                    except Exception:
                        pass  # Silently ignore cleanup errors
            cls._instances.clear()


class HookStateManager(metaclass=SingletonMeta):
    """Centralized state management for hook execution tracking and deduplication

    Handles:
    - Hook execution counting and tracking
    - Phase-based deduplication for SessionStart
    - Command deduplication within time windows
    - Thread-safe state operations
    - Persistent state storage
    - Performance monitoring and metrics collection
    - Configurable deduplication parameters
    """

    def __init__(self, cwd: str, config: Optional[HookConfiguration] = None):
        """Initialize state manager for given working directory

        Args:
            cwd: Current working directory path
            config: Optional configuration object, defaults to environment-based config
        """
        self.cwd = cwd
        self.config = config or HookConfiguration.from_env()
        self.logger = get_logger()

        # Configure logging based on config
        configure_logging(
            debug_mode=self.config.debug_mode,
            verbose=self.config.enable_verbose_logging
        )

        # Initialize state directory with fallback logic
        self.state_dir = self.config.get_state_dir(cwd)
        self.logger.debug(f"Using state directory: {self.state_dir}")

        # Thread safety with configurable timeout
        self._lock = threading.RLock()
        self._lock_timeout = self.config.lock_timeout_seconds

        # Thread lifecycle management
        self._cleanup_event = threading.Event()
        self._cleanup_thread = None
        self._threads: list[threading.Thread] = []  # Track all created threads for proper cleanup

        # State files
        self.hook_state_file = self.state_dir / "hook_execution_state.json"
        self.command_state_file = self.state_dir / "command_execution_state.json"
        self.performance_metrics_file = self.state_dir / "performance_metrics.json"

        # In-memory cache for performance
        self._hook_state_cache: Optional[Dict[str, Any]] = None
        self._command_state_cache: Optional[Dict[str, Any]] = None
        self._cache_timestamp: float = 0.0

        # Performance tracking
        self._performance_metrics = get_performance_metrics()

        # Start cache cleanup thread if enabled
        if self.config.enable_caching and self.config.cache_cleanup_interval > 0:
            self._start_cache_cleanup_thread()

        # Register cleanup on garbage collection
        weakref.finalize(self, self._cleanup_on_finalize)

    def _cleanup_on_finalize(self):
        """Cleanup method called during garbage collection"""
        try:
            self.cleanup()
        except Exception:
            pass  # Silently ignore cleanup errors during finalization

    def _start_cache_cleanup_thread(self):
        """Start background thread for cache cleanup"""
        def cleanup_task():
            while not self._cleanup_event.is_set():
                # Use event.wait() instead of time.sleep() for immediate response to stop signal
                if self._cleanup_event.wait(timeout=self.config.cache_cleanup_interval):
                    break  # Stop signal received
                try:
                    self._cleanup_expired_cache_entries()
                except Exception as e:
                    self.logger.warning(f"Cache cleanup task failed: {e}")

        cleanup_thread = threading.Thread(target=cleanup_task, daemon=True)
        cleanup_thread.start()
        self._threads.append(cleanup_thread)
        self.logger.debug("Started cache cleanup background thread")

    def _cleanup_expired_cache_entries(self):
        """Clean up expired cache entries"""
        current_time = time.time()

        # Clean hook state cache
        if (self._hook_state_cache and
            current_time - self._cache_timestamp > self.config.state_cache_ttl):
            self._hook_state_cache = None
            self._cache_timestamp = 0
            self.logger.debug("Cleaned expired hook state cache")

        # Clean command state cache
        if (self._command_state_cache and
            current_time - self._cache_timestamp > self.config.state_cache_ttl):
            self._command_state_cache = None
            self._cache_timestamp = 0
            self.logger.debug("Cleaned expired command state cache")

    def _load_hook_state(self) -> Dict[str, Any]:
        """Load hook execution state with caching and error handling

        Returns:
            Dictionary containing hook execution state
        """
        current_time = time.time()

        # Use cache if recent (within configured TTL)
        if (self._hook_state_cache and
            current_time - self._cache_timestamp < self.config.state_cache_ttl):
            if self.config.log_state_changes:
                self.logger.debug("Hook state cache hit")
            record_cache_hit()
            return self._hook_state_cache

        try:
            if self.hook_state_file.exists():
                with open(self.hook_state_file, "r", encoding=self.config.state_file_encoding) as f:
                    state = json.load(f)

                self._hook_state_cache = state
                self._cache_timestamp = current_time
                self._performance_metrics.record_state_read()
                if self.config.log_state_changes:
                    self.logger.debug(f"Loaded hook state from {self.hook_state_file}")
                record_cache_hit()
                return state
        except (IOError, json.JSONDecodeError, Exception) as e:
            self.logger.warning(f"Failed to load hook state: {e}")
            self._performance_metrics.record_io_error()
            record_cache_miss()

        # Default state structure
        default_state: Dict[str, Any] = {}
        self._hook_state_cache = default_state
        self._cache_timestamp = current_time
        record_cache_miss()
        return default_state

    def _save_hook_state(self, state: Dict[str, Any]) -> bool:
        """Save hook execution state with error handling

        Args:
            state: Hook state to save

        Returns:
            bool: True if save was successful, False otherwise
        """
        if not self.config.enable_state_persistence:
            return False

        try:
            # Create backup if enabled
            if self.config.backup_on_write and self.hook_state_file.exists():
                backup_file = self.hook_state_file.with_suffix('.json.backup')
                import shutil
                shutil.copy2(self.hook_state_file, backup_file)

            with open(self.hook_state_file, "w", encoding=self.config.state_file_encoding) as f:
                json.dump(state, f, indent=self.config.state_file_indent)

            self._hook_state_cache = state
            self._cache_timestamp = time.time()
            self._performance_metrics.record_state_write()

            if self.config.log_state_changes:
                self.logger.debug(f"Saved hook state to {self.hook_state_file}")
            return True

        except (IOError, OSError, Exception) as e:
            self.logger.error(f"Failed to save hook state: {e}")
            self._performance_metrics.record_io_error()
            return False

    def _load_command_state(self) -> Dict[str, Any]:
        """Load command execution state with caching and error handling

        Returns:
            Dictionary containing command execution state
        """
        current_time = time.time()

        # Use cache if recent (within configured TTL)
        if (self._command_state_cache and
            current_time - self._cache_timestamp < self.config.state_cache_ttl):
            if self.config.log_state_changes:
                self.logger.debug("Command state cache hit")
            record_cache_hit()
            return self._command_state_cache

        try:
            if self.command_state_file.exists():
                with open(self.command_state_file, "r", encoding=self.config.state_file_encoding) as f:
                    state = json.load(f)

                self._command_state_cache = state
                self._cache_timestamp = current_time
                self._performance_metrics.record_state_read()
                if self.config.log_state_changes:
                    self.logger.debug(f"Loaded command state from {self.command_state_file}")
                record_cache_hit()
                return state
        except (IOError, json.JSONDecodeError, Exception) as e:
            self.logger.warning(f"Failed to load command state: {e}")
            self._performance_metrics.record_io_error()
            record_cache_miss()

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
        record_cache_miss()
        return default_state

    def _save_command_state(self, state: Dict[str, Any]) -> bool:
        """Save command execution state with error handling

        Args:
            state: Command state to save

        Returns:
            bool: True if save was successful, False otherwise
        """
        if not self.config.enable_state_persistence:
            return False

        try:
            # Create backup if enabled
            if self.config.backup_on_write and self.command_state_file.exists():
                backup_file = self.command_state_file.with_suffix('.json.backup')
                import shutil
                shutil.copy2(self.command_state_file, backup_file)

            with open(self.command_state_file, "w", encoding=self.config.state_file_encoding) as f:
                json.dump(state, f, indent=self.config.state_file_indent)

            self._command_state_cache = state
            self._cache_timestamp = time.time()
            self._performance_metrics.record_state_write()

            if self.config.log_state_changes:
                self.logger.debug(f"Saved command state to {self.command_state_file}")
            return True

        except (IOError, OSError, Exception) as e:
            self.logger.error(f"Failed to save command state: {e}")
            self._performance_metrics.record_io_error()
            return False

    def track_hook_execution(self, hook_name: str, phase: str = None) -> ExecutionResult:
        """Track hook execution and return execution information

        Args:
            hook_name: Name of the hook being executed
            phase: Optional phase for phase-based deduplication

        Returns:
            ExecutionResult with detailed execution information including:
            - executed: Whether the hook was actually executed
            - duplicate: Whether this was a duplicate execution
            - execution_id: Unique identifier for this execution
            - execution_count: Total execution count
            - performance metrics and error information
        """
        start_time = time.time()

        try:
            with self._lock:
                # Acquire lock with timeout
                if not self._lock.acquire(timeout=self._lock_timeout):
                    self.logger.warning("Failed to acquire lock for hook tracking")
                    self._performance_metrics.record_concurrent_access_error()
                    return ExecutionResult(
                        executed=True,  # Allow execution to continue despite lock issue
                        duplicate=False,
                        execution_id=str(uuid.uuid4()),
                        timestamp=start_time,
                        error="Failed to acquire lock for hook tracking"
                    )

                state = self._load_hook_state()
                current_time = time.time()
                execution_id = str(uuid.uuid4())

                # Initialize hook state if not exists
                if hook_name not in state:
                    state[hook_name] = {
                        "count": 0,
                        "last_execution": 0,
                        "last_phase": None,
                        "executions": []
                    }

                hook_state = state[hook_name]

                # Check for deduplication
                is_duplicate = False
                deduplication_reason = None

                # Phase-based deduplication for SessionStart
                if hook_name == "SessionStart" and phase:
                    # Phase transitions are allowed (clear->compact or compact->clear)
                    if (phase == hook_state.get("last_phase") and
                        current_time - hook_state["last_execution"] < self.config.hook_dedupe_window):
                        # Same phase within time window - deduplicate
                        is_duplicate = True
                        deduplication_reason = f"same phase within {self.config.hook_dedupe_window}s window"
                    else:
                        # Different phase or time window expired - execute
                        pass
                else:
                    # Regular deduplication based on time window
                    if (current_time - hook_state["last_execution"] < self.config.hook_dedupe_window):
                        is_duplicate = True
                        deduplication_reason = f"within {self.config.hook_dedupe_window}s deduplication window"

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
                        if current_time - e["timestamp"] < self.config.max_state_file_age_hours * 3600
                    ]
                    if len(recent_executions) != len(hook_state["executions"]):
                        hook_state["executions"] = recent_executions

                    # Save state
                    save_success = self._save_hook_state(state)
                    if not save_success:
                        self.logger.warning("Failed to save hook state")

                # Create execution result
                execution_time_ms = (time.time() - start_time) * 1000
                execution_count = hook_state["count"]
                duplicate_count = state.get("duplicate_count", 0)

                result = ExecutionResult(
                    executed=not is_duplicate,
                    duplicate=is_duplicate,
                    execution_id=execution_id,
                    timestamp=current_time,
                    hook_name=hook_name,
                    phase=phase,
                    reason=deduplication_reason,
                    execution_time_ms=execution_time_ms,
                    execution_count=execution_count,
                    duplicate_count=duplicate_count,
                    state_operations_count=2,  # Load + Save
                    cache_hit=bool(self._hook_state_cache),
                    warning=None if save_success else "Failed to save state but continuing execution"
                )

                # Record performance metrics
                record_execution_metrics(
                    execution_time_ms,
                    success=True,
                    is_duplicate=is_duplicate
                )

                if self.config.log_state_changes:
                    msg = (
                        f"Hook execution tracked: {hook_name} "
                        f"(executed: {not is_duplicate}, duplicate: {is_duplicate})"
                    )
                    self.logger.info(msg)

                return result

        except Exception as e:
            execution_time_ms = (time.time() - start_time) * 1000

            # Record failure metrics
            record_execution_metrics(
                execution_time_ms,
                success=False,
                is_duplicate=False
            )

            self.logger.error(f"Error in hook execution tracking: {e}")
            self._performance_metrics.record_other_error()

            return ExecutionResult(
                executed=True,  # Allow execution to continue despite error
                duplicate=False,
                execution_id=str(uuid.uuid4()),
                timestamp=start_time,
                hook_name=hook_name,
                phase=phase,
                error=str(e),
                execution_time_ms=execution_time_ms,
                warning="Execution tracking failed but continuing execution"
            )

        finally:
            # Release lock (safe to release even if not acquired in this method)
            try:
                self._lock.release()
            except RuntimeError:
                pass  # Lock was not acquired by this thread

    def deduplicate_command(self, command: str) -> ExecutionResult:
        """Check and deduplicate command execution within time window

        Args:
            command: Command string to check for deduplication

        Returns:
            ExecutionResult with detailed deduplication information including:
            - executed: Whether the command should execute
            - duplicate: Whether this was a duplicate
            - reason: Reason for deduplication decision
            - execution_count: Total execution count
            - performance metrics and error information
        """
        start_time = time.time()

        try:
            with self._lock:
                # Acquire lock with timeout
                if not self._lock.acquire(timeout=self._lock_timeout):
                    self.logger.warning("Failed to acquire lock for command deduplication")
                    self._performance_metrics.record_concurrent_access_error()
                    return ExecutionResult(
                        executed=True,  # Allow execution to continue despite lock issue
                        duplicate=False,
                        execution_id=str(uuid.uuid4()),
                        timestamp=start_time,
                        command=command,
                        error="Failed to acquire lock for command deduplication"
                    )

                state = self._load_command_state()
                current_time = time.time()
                execution_id = str(uuid.uuid4())

                # Check if command is an Alfred command (only deduplicate these)
                if not command or not command.startswith("/alfred:"):
                    result = ExecutionResult(
                        executed=True,
                        duplicate=False,
                        execution_id=execution_id,
                        timestamp=current_time,
                        command=command,
                        reason="non-alfred command",
                        execution_count=state["execution_count"],
                        execution_time_ms=(time.time() - start_time) * 1000,
                        state_operations_count=1  # Load
                    )

                    # Record performance metrics
                    record_execution_metrics(
                        result.execution_time_ms,
                        success=True,
                        is_duplicate=False
                    )

                    if self.config.log_state_changes:
                        self.logger.info(f"Non-Alfred command: {command}")

                    return result

                # Check for duplicate within time window
                last_cmd = state.get("last_command")
                last_timestamp = state.get("last_timestamp")

                is_duplicate = False
                deduplication_reason = None

                if (last_cmd and last_timestamp and
                    command == last_cmd and
                    current_time - last_timestamp < self.config.command_dedupe_window):

                    # Duplicate detected
                    is_duplicate = True
                    deduplication_reason = f"within {self.config.command_dedupe_window}s deduplication window"
                    state["duplicate_count"] += 1
                    state["is_running"] = True  # Mark as running to prevent further duplicates
                    state["duplicate_timestamp"] = datetime.fromtimestamp(current_time).isoformat()

                    # Save state
                    save_success = self._save_command_state(state)
                    if not save_success:
                        self.logger.warning("Failed to save command state for duplicate detection")

                    result = ExecutionResult(
                        executed=True,  # Allow execution but mark as duplicate
                        duplicate=True,
                        execution_id=execution_id,
                        timestamp=current_time,
                        command=command,
                        phase=None,
                        reason=deduplication_reason,
                        execution_count=state["execution_count"],
                        duplicate_count=state["duplicate_count"],
                        execution_time_ms=(time.time() - start_time) * 1000,
                        state_operations_count=2,  # Load + Save
                        cache_hit=bool(self._command_state_cache),
                        warning=None if save_success else "Failed to save duplicate state"
                    )

                else:
                    # Not a duplicate - update state and execute
                    state["last_command"] = command
                    state["last_timestamp"] = current_time
                    state["is_running"] = True
                    state["execution_count"] += 1

                    # Save state
                    save_success = self._save_command_state(state)
                    if not save_success:
                        self.logger.warning("Failed to save command state for normal execution")

                    result = ExecutionResult(
                        executed=True,
                        duplicate=False,
                        execution_id=execution_id,
                        timestamp=current_time,
                        command=command,
                        reason="normal execution",
                        execution_count=state["execution_count"],
                        execution_time_ms=(time.time() - start_time) * 1000,
                        state_operations_count=2,  # Load + Save
                        cache_hit=bool(self._command_state_cache),
                        warning=None if save_success else "Failed to save command state"
                    )

                # Record performance metrics
                record_execution_metrics(
                    result.execution_time_ms,
                    success=True,
                    is_duplicate=is_duplicate
                )

                if self.config.log_state_changes:
                    msg = (
                        f"Command deduplication: {command} "
                        f"(executed: {not is_duplicate}, duplicate: {is_duplicate})"
                    )
                    self.logger.info(msg)

                return result

        except Exception as e:
            execution_time_ms = (time.time() - start_time) * 1000

            # Record failure metrics
            record_execution_metrics(
                execution_time_ms,
                success=False,
                is_duplicate=False
            )

            self.logger.error(f"Error in command deduplication: {e}")
            self._performance_metrics.record_other_error()

            return ExecutionResult(
                executed=True,  # Allow execution to continue despite error
                duplicate=False,
                execution_id=str(uuid.uuid4()),
                timestamp=start_time,
                command=command,
                error=str(e),
                execution_time_ms=execution_time_ms,
                warning="Command deduplication failed but continuing execution"
            )

        finally:
            # Release lock (safe to release even if not acquired in this method)
            try:
                self._lock.release()
            except RuntimeError:
                pass  # Lock was not acquired by this thread

    def mark_command_complete(self, command: Optional[str] = None) -> None:
        """Mark command execution as complete

        Args:
            command: Optional command that completed
        """
        try:
            with self._lock:
                state = self._load_command_state()
                state["is_running"] = False
                state["last_timestamp"] = time.time()
                if command:
                    state["last_command"] = command

                self._save_command_state(state)

                if self.config.log_state_changes:
                    self.logger.info(f"Command marked as complete: {command or 'unknown'}")

        except Exception as e:
            self.logger.error(f"Failed to mark command as complete: {e}")
            self._performance_metrics.record_other_error()

    def get_hook_execution_count(self, hook_name: str) -> int:
        """Get total execution count for a hook"""
        state = self._load_hook_state()
        return state.get(hook_name, {}).get("count", 0)

    def get_command_execution_count(self) -> int:
        """Get total command execution count"""
        state = self._load_command_state()
        return state.get("execution_count", 0)

    def cleanup_old_states(self, max_age_hours: int = None) -> None:
        """Clean up old state entries to prevent state file bloat

        Args:
            max_age_hours: Maximum age for state entries in hours, defaults to config setting
        """
        if not self.config.enable_state_persistence:
            return

        max_age = max_age_hours or self.config.max_state_file_age_hours
        current_time = time.time()
        max_age_seconds = max_age * 3600

        try:
            with self._lock:
                # Clean up hook state
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
                            if self.config.log_state_changes:
                                self.logger.debug(f"Cleaned up {len(recent_executions)} executions for {hook_name}")

                    # Remove hooks with no recent executions
                    if (hook_data.get("last_execution", 0) < current_time - max_age_seconds):
                        del hook_state[hook_name]
                        if self.config.log_state_changes:
                            self.logger.debug(f"Removed old hook state: {hook_name}")

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
                    if self.config.log_state_changes:
                        self.logger.debug("Reset old command state")

        except Exception as e:
            self.logger.error(f"Failed to clean up old states: {e}")
            self._performance_metrics.record_other_error()

    def get_performance_summary(self) -> Dict[str, Any]:
        """Get performance summary for the state manager

        Returns:
            Dictionary with performance metrics summary
        """
        return self._performance_metrics.get_summary()

    def reset_performance_metrics(self) -> None:
        """Reset performance metrics for this state manager"""
        # Note: This functionality requires _global_metrics_lock and PerformanceMetrics
        # which are not currently defined in the module
        self._performance_metrics = get_performance_metrics()
        if self.config.log_state_changes:
            self.logger.info("Performance metrics reset")

    def cleanup(self, timeout: float = 5.0) -> None:
        """Cleanup resources and stop all threads

        Args:
            timeout: Maximum time to wait for threads to finish (default: 5.0 seconds)
        """
        if self._cleanup_event.is_set():
            return  # Already cleaned up

        # Signal all threads to stop
        self._cleanup_event.set()

        # Wait for all threads to finish
        for thread in self._threads:
            if thread.is_alive():
                thread.join(timeout=timeout)
                if thread.is_alive():
                    self.logger.warning(f"Thread {thread.name} did not finish within timeout")

        # Clear thread list
        self._threads.clear()

        # Clear caches
        self._hook_state_cache = None
        self._command_state_cache = None
        self._cache_timestamp = 0

        if self.config.log_state_changes:
            self.logger.debug("HookStateManager cleanup completed")

    def stop(self) -> None:
        """Alias for cleanup method - stop the state manager"""
        self.cleanup()

    def __enter__(self):
        """Context manager entry"""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit - ensures cleanup"""
        self.cleanup()
        return False  # Don't suppress exceptions

    def __del__(self):
        """Destructor - ensures cleanup on garbage collection"""
        try:
            if not self._cleanup_event.is_set():
                self.cleanup(timeout=1.0)  # Quick cleanup during garbage collection
        except Exception:
            pass  # Silently ignore cleanup errors during destruction


# Global state manager instances (per-CWD)
_state_managers: Dict[str, HookStateManager] = {}
_state_manager_lock = threading.RLock()


def get_state_manager(cwd: str, config: Optional[HookConfiguration] = None) -> HookStateManager:
    """Get or create state manager for given working directory

    Note: With Singleton pattern, each cwd gets its own unique instance,
    but duplicate instances for the same cwd are prevented.

    Args:
        cwd: Current working directory path
        config: Optional configuration object

    Returns:
        HookStateManager instance for the given directory
    """
    with _state_manager_lock:
        # Check if we already have an instance for this cwd
        existing_instance_key = None
        for key, instance in _state_managers.items():
            if hasattr(instance, 'cwd') and instance.cwd == cwd:
                existing_instance_key = key
                break

        if existing_instance_key is not None:
            return _state_managers[existing_instance_key]

        # Create new instance with cwd as part of key for uniqueness
        instance_key = f"{cwd}_{len(_state_managers)}"
        try:
            instance = HookStateManager(cwd, config)
            _state_managers[instance_key] = instance
            return instance
        except Exception as e:
            # Clean up on creation failure
            if instance_key in _state_managers:
                del _state_managers[instance_key]
            raise e


def track_hook_execution(
    hook_name: str,
    cwd: str,
    phase: str = None,
    config: Optional[HookConfiguration] = None,
) -> ExecutionResult:
    """Convenience function to track hook execution

    Args:
        hook_name: Name of the hook being executed
        cwd: Current working directory
        phase: Optional phase for phase-based deduplication
        config: Optional configuration object

    Returns:
        ExecutionResult with execution information
    """
    manager = get_state_manager(cwd, config)
    return manager.track_hook_execution(hook_name, phase)


def deduplicate_command(command: str, cwd: str, config: Optional[HookConfiguration] = None) -> ExecutionResult:
    """Convenience function to deduplicate command

    Args:
        command: Command string to check for deduplication
        cwd: Current working directory
        config: Optional configuration object

    Returns:
        ExecutionResult with deduplication information
    """
    manager = get_state_manager(cwd, config)
    return manager.deduplicate_command(command)


def mark_command_complete(
    command: Optional[str] = None,
    cwd: Optional[str] = None,
    config: Optional[HookConfiguration] = None,
) -> None:
    """Convenience function to mark command complete

    Args:
        command: Optional command that completed
        cwd: Current working directory
        config: Optional configuration object
    """
    if not cwd:
        cwd = "."
    manager = get_state_manager(cwd, config)
    manager.mark_command_complete(command)


def cleanup_old_states(
    max_age_hours: Optional[int] = None,
    cwd: Optional[str] = None,
    config: Optional[HookConfiguration] = None,
) -> None:
    """Convenience function to clean up old states

    Args:
        max_age_hours: Maximum age for state entries in hours
        cwd: Current working directory
        config: Optional configuration object
    """
    if not cwd:
        cwd = "."
    manager = get_state_manager(cwd, config)
    manager.cleanup_old_states(max_age_hours)


def get_performance_summary(cwd: Optional[str] = None, config: Optional[HookConfiguration] = None) -> Dict[str, Any]:
    """Convenience function to get performance summary

    Args:
        cwd: Current working directory
        config: Optional configuration object

    Returns:
        Dictionary with performance metrics summary
    """
    if not cwd:
        cwd = "."
    manager = get_state_manager(cwd, config)
    return manager.get_performance_summary()


def cleanup_all_state_managers(timeout: float = 5.0) -> None:
    """Cleanup all state manager instances and stop all threads

    Args:
        timeout: Maximum time to wait for threads to finish (default: 5.0 seconds)
    """
    with _state_manager_lock:
        for instance in list(_state_managers.values()):
            try:
                instance.cleanup(timeout)
            except Exception as e:
                # Log error but continue cleanup of other instances
                try:
                    logger = get_logger()
                    logger.error(f"Error during state manager cleanup: {e}")
                except Exception:
                    pass  # Silently ignore logging errors

        # Clear all instances
        _state_managers.clear()


def force_cleanup_all_singletons() -> None:
    """Force cleanup of all singleton instances using the metaclass cleanup method"""
    try:
        HookStateManager.cleanup_all_instances()
    except Exception:
        pass  # Silently ignore cleanup errors


# Module-level cleanup on import/unload
atexit.register(cleanup_all_state_managers)
atexit.register(force_cleanup_all_singletons)
