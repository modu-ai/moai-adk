#!/usr/bin/env python3
"""Unified Hook Timeout Manager

Prevents timeout conflicts across hooks by providing a centralized
timeout management system with proper resource cleanup and graceful degradation.

Architecture:
- Global timeout registry to prevent multiple signal.alarm() conflicts
- Cross-platform compatibility (Windows/Unix)
- Resource cleanup and memory monitoring
- Configurable timeout policies per hook type
- Graceful degradation with retry mechanisms
"""

import contextlib
import logging
import platform
import signal
import threading
import time
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Callable, Dict, Optional, Set

import yaml


class TimeoutPolicy(Enum):
    """Timeout policy types for different hook categories"""

    FAST = "fast"  # 1-2 seconds (PreTool hooks)
    NORMAL = "normal"  # 3-5 seconds (SessionStart/End)
    SLOW = "slow"  # 10-15 seconds (Complex operations)
    CUSTOM = "custom"  # User-defined timeout


@dataclass
class HookTimeoutConfig:
    """Configuration for hook timeout behavior"""

    policy: TimeoutPolicy = TimeoutPolicy.NORMAL
    custom_timeout_ms: Optional[int] = None
    retry_count: int = 0
    retry_delay_ms: int = 100
    graceful_degradation: bool = True
    memory_limit_mb: Optional[int] = None
    on_timeout_callback: Optional[Callable] = None


@dataclass
class TimeoutSession:
    """Active timeout session tracking"""

    hook_id: str
    start_time: datetime
    timeout_seconds: float
    thread_id: int
    callback: Optional[Callable] = None
    completed: bool = False
    cleanup_actions: list = field(default_factory=list)


class HookTimeoutError(Exception):
    """Enhanced timeout error with context"""

    def __init__(
        self,
        message: str,
        hook_id: str = "",
        timeout_seconds: float = 0,
        execution_time: float = 0,
        will_retry: bool = False,
    ):
        super().__init__(message)
        self.hook_id = hook_id
        self.timeout_seconds = timeout_seconds
        self.execution_time = execution_time
        self.will_retry = will_retry


class UnifiedTimeoutManager:
    """Centralized timeout manager for all hooks

    Features:
    - Prevents signal conflicts by tracking active timeouts
    - Cross-platform compatibility (Windows threading, Unix signals)
    - Memory usage monitoring
    - Configurable retry mechanisms
    - Graceful degradation with proper cleanup
    """

    # Global singleton instance
    _instance: Optional["UnifiedTimeoutManager"] = None
    _lock = threading.Lock()

    def __new__(cls) -> "UnifiedTimeoutManager":
        """Singleton pattern to ensure single timeout manager"""
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self) -> None:
        if hasattr(self, "_initialized"):
            return

        self._initialized = True
        self._logger = logging.getLogger(__name__)

        # Active timeout tracking
        self._active_sessions: Dict[str, TimeoutSession] = {}
        self._session_lock = threading.RLock()

        # Platform detection
        self._is_windows = platform.system() == "Windows"

        # Signal management for Unix
        self._original_signal_handler: Optional[Callable] = None
        self._signal_lock = threading.Lock()

        # Resource monitoring
        self._memory_tracker: Dict[str, Any] = {}
        self._cleanup_registry: Set[str] = set()

        # Default timeout configurations
        self._default_configs = {
            TimeoutPolicy.FAST: HookTimeoutConfig(
                policy=TimeoutPolicy.FAST,
                custom_timeout_ms=2000,  # 2 seconds
                retry_count=1,
                retry_delay_ms=100,
            ),
            TimeoutPolicy.NORMAL: HookTimeoutConfig(
                policy=TimeoutPolicy.NORMAL,
                custom_timeout_ms=5000,  # 5 seconds
                retry_count=1,
                retry_delay_ms=200,
            ),
            TimeoutPolicy.SLOW: HookTimeoutConfig(
                policy=TimeoutPolicy.SLOW,
                custom_timeout_ms=15000,  # 15 seconds
                retry_count=2,
                retry_delay_ms=500,
            ),
        }

        self._logger.info("UnifiedTimeoutManager initialized")

    def load_config(self) -> Dict[str, Any]:
        """Load timeout configuration from project config"""
        try:
            config_file = Path(".moai/config/config.yaml")
            if config_file.exists():
                with open(config_file, "r", encoding="utf-8") as f:
                    config = yaml.safe_load(f) or {}
                    return config.get("hooks", {}).get("timeout_manager", {})
        except Exception as e:
            self._logger.warning(f"Failed to load timeout config: {e}")
        return {}

    def get_timeout_config(
        self, hook_name: str, custom_config: Optional[HookTimeoutConfig] = None
    ) -> HookTimeoutConfig:
        """Get timeout configuration for a specific hook"""
        if custom_config:
            return custom_config

        # Load from config file
        config = self.load_config()
        hook_configs = config.get("hook_configs", {})

        if hook_name in hook_configs:
            hook_config = hook_configs[hook_name]
            policy_name = hook_config.get("policy", "normal")
            policy = TimeoutPolicy(policy_name)

            return HookTimeoutConfig(
                policy=policy,
                custom_timeout_ms=hook_config.get("timeout_ms"),
                retry_count=hook_config.get("retry_count", 1),
                retry_delay_ms=hook_config.get("retry_delay_ms", 200),
                graceful_degradation=hook_config.get("graceful_degradation", True),
                memory_limit_mb=hook_config.get("memory_limit_mb"),
            )

        # Use default based on hook name patterns
        if "pre_tool" in hook_name.lower():
            return self._default_configs[TimeoutPolicy.FAST]
        elif any(x in hook_name.lower() for x in ["session_start", "session_end"]):
            return self._default_configs[TimeoutPolicy.NORMAL]
        else:
            return self._default_configs[TimeoutPolicy.NORMAL]

    def create_timeout_session(self, hook_name: str, config: Optional[HookTimeoutConfig] = None) -> TimeoutSession:
        """Create a new timeout session for a hook"""
        if not config:
            config = self.get_timeout_config(hook_name)

        # Calculate timeout in seconds
        timeout_ms = config.custom_timeout_ms
        if timeout_ms is None:
            # Default based on policy
            policy_timeouts = {
                TimeoutPolicy.FAST: 2000,
                TimeoutPolicy.NORMAL: 5000,
                TimeoutPolicy.SLOW: 15000,
            }
            timeout_ms = policy_timeouts.get(config.policy, 5000)

        timeout_seconds = timeout_ms / 1000.0

        # Generate unique hook ID
        hook_id = f"{hook_name}_{int(time.time() * 1000)}_{threading.get_ident()}"

        session = TimeoutSession(
            hook_id=hook_id,
            start_time=datetime.now(),
            timeout_seconds=timeout_seconds,
            thread_id=threading.get_ident(),
            callback=config.on_timeout_callback,
        )

        with self._session_lock:
            self._active_sessions[hook_id] = session

        return session

    def start_timeout(self, session: TimeoutSession) -> None:
        """Start timeout monitoring for a session"""
        if self._is_windows:
            self._start_windows_timeout(session)
        else:
            self._start_unix_timeout(session)

    def _start_windows_timeout(self, session: TimeoutSession) -> None:
        """Windows: Use threading.Timer for timeout"""

        def timeout_handler():
            if session.hook_id in self._active_sessions and not session.completed:
                execution_time = (datetime.now() - session.start_time).total_seconds()
                error_msg = (
                    f"Hook {session.hook_id} timed out after "
                    f"{session.timeout_seconds}s (execution: {execution_time:.2f}s)"
                )

                # Execute callback if provided
                if session.callback:
                    try:
                        session.callback(session)
                    except Exception as e:
                        self._logger.error(f"Timeout callback failed: {e}")

                # Don't raise exception in timer thread, just mark as completed
                session.completed = True
                self._logger.warning(error_msg)

        timer = threading.Timer(session.timeout_seconds, timeout_handler)
        timer.daemon = True
        timer.start()

        # Store timer for cleanup
        session.cleanup_actions.append(lambda: timer.cancel() if timer.is_alive() else None)

    def _start_unix_timeout(self, session: TimeoutSession) -> None:
        """Unix: Use signal.SIGALRM for timeout"""
        with self._signal_lock:
            # Register our signal handler if not already done
            if self._original_signal_handler is None:
                self._original_signal_handler = signal.signal(  # type: ignore[assignment]
                    signal.SIGALRM, self._unix_signal_handler
                )

            # Set alarm for this session
            signal.alarm(int(session.timeout_seconds))

            # Store cleanup action
            session.cleanup_actions.append(lambda: signal.alarm(0))

    def _unix_signal_handler(self, signum: int, frame) -> None:
        """Unix signal handler for timeout"""
        with self._session_lock:
            # Find the session that timed out
            current_thread = threading.get_ident()
            timed_out_session = None

            for hook_id, session in self._active_sessions.items():
                if session.thread_id == current_thread and not session.completed:
                    timed_out_session = session
                    break

            if timed_out_session:
                execution_time = (datetime.now() - timed_out_session.start_time).total_seconds()
                error_msg = (
                    f"Hook {timed_out_session.hook_id} timed out after "
                    f"{timed_out_session.timeout_seconds}s (execution: {execution_time:.2f}s)"
                )

                # Execute callback if provided
                if timed_out_session.callback:
                    try:
                        timed_out_session.callback(timed_out_session)
                    except Exception as e:
                        self._logger.error(f"Timeout callback failed: {e}")

                # Mark as completed
                timed_out_session.completed = True

                # Raise timeout exception
                raise HookTimeoutError(
                    error_msg,
                    hook_id=timed_out_session.hook_id,
                    timeout_seconds=timed_out_session.timeout_seconds,
                    execution_time=execution_time,
                )

    def cancel_timeout(self, session: TimeoutSession) -> None:
        """Cancel timeout for a session and perform cleanup"""
        session.completed = True

        # Execute cleanup actions
        for cleanup_action in session.cleanup_actions:
            try:
                if cleanup_action:
                    cleanup_action()
            except Exception as e:
                self._logger.warning(f"Cleanup action failed: {e}")

        # Remove from active sessions
        with self._session_lock:
            if session.hook_id in self._active_sessions:
                del self._active_sessions[session.hook_id]

    def execute_with_timeout(
        self,
        hook_name: str,
        func: Callable,
        *args,
        config: Optional[HookTimeoutConfig] = None,
        **kwargs,
    ) -> Any:
        """Execute a function with timeout management and retry logic"""
        if not config:
            config = self.get_timeout_config(hook_name)

        last_exception: Optional[Exception] = None

        for attempt in range(config.retry_count + 1):
            session = self.create_timeout_session(hook_name, config)

            try:
                # Check memory limit if specified
                if config.memory_limit_mb:
                    self._check_memory_usage(config.memory_limit_mb)

                # Start timeout
                self.start_timeout(session)

                # Execute function
                start_time = time.time()
                result = func(*args, **kwargs)
                execution_time = time.time() - start_time

                # Cancel timeout on success
                self.cancel_timeout(session)

                self._logger.debug(f"Hook {hook_name} completed in {execution_time:.2f}s")
                return result

            except HookTimeoutError as e:
                last_exception = e
                self.cancel_timeout(session)

                if attempt < config.retry_count:
                    self._logger.warning(f"Hook {hook_name} timeout, retrying ({attempt + 1}/{config.retry_count})")
                    time.sleep(config.retry_delay_ms / 1000.0)
                else:
                    self._logger.error(f"Hook {hook_name} failed after {config.retry_count} retries")

                    if config.graceful_degradation:
                        return self._get_graceful_degradation_result(hook_name)
                    else:
                        raise

            except Exception as e:
                last_exception = e
                self.cancel_timeout(session)

                if attempt < config.retry_count:
                    self._logger.warning(f"Hook {hook_name} error, retrying ({attempt + 1}/{config.retry_count}): {e}")
                    time.sleep(config.retry_delay_ms / 1000.0)
                else:
                    self._logger.error(f"Hook {hook_name} failed with exception: {e}")
                    raise

        # This should not be reached, but handle just in case
        if last_exception:
            raise last_exception

    def _check_memory_usage(self, limit_mb: int) -> None:
        """Check current memory usage against limit"""
        try:
            import psutil

            process = psutil.Process()
            memory_mb = process.memory_info().rss / 1024 / 1024

            if memory_mb > limit_mb:
                self._logger.warning(f"Memory usage ({memory_mb:.1f}MB) exceeds limit ({limit_mb}MB)")
                # Could implement more aggressive cleanup here
        except ImportError:
            # psutil not available, skip memory checking
            pass
        except Exception as e:
            self._logger.warning(f"Memory check failed: {e}")

    def _get_graceful_degradation_result(self, hook_name: str) -> Any:
        """Return a safe default result for graceful degradation"""
        # Return different defaults based on hook type
        if "session_start" in hook_name.lower():
            return {
                "continue": True,
                "systemMessage": "⚠️ Session start hook timeout - continuing with reduced functionality",
                "graceful_degradation": True,
            }
        elif "session_end" in hook_name.lower():
            return {
                "continue": True,
                "success": False,
                "error": "Hook timeout but continuing due to graceful degradation",
                "graceful_degradation": True,
            }
        elif "pre_tool" in hook_name.lower():
            return {
                "continue": True,
                "systemMessage": "⚠️ Validation timeout - operation proceeding",
                "graceful_degradation": True,
            }
        else:
            return {
                "continue": True,
                "error": "Hook timeout but continuing due to graceful degradation",
                "graceful_degradation": True,
            }

    def get_active_sessions(self) -> Dict[str, Dict[str, Any]]:
        """Get information about active timeout sessions"""
        with self._session_lock:
            return {
                hook_id: {
                    "hook_id": session.hook_id,
                    "start_time": session.start_time.isoformat(),
                    "timeout_seconds": session.timeout_seconds,
                    "thread_id": session.thread_id,
                    "completed": session.completed,
                    "elapsed": (datetime.now() - session.start_time).total_seconds(),
                }
                for hook_id, session in self._active_sessions.items()
            }

    def cleanup_completed_sessions(self) -> int:
        """Clean up completed sessions and return count cleaned"""
        with self._session_lock:
            completed_ids = [hook_id for hook_id, session in self._active_sessions.items() if session.completed]

            for hook_id in completed_ids:
                del self._active_sessions[hook_id]

            return len(completed_ids)

    def shutdown(self) -> None:
        """Shutdown the timeout manager and clean up all resources"""
        with self._session_lock:
            # Cancel all active sessions
            for session in list(self._active_sessions.values()):
                try:
                    self.cancel_timeout(session)
                except Exception as e:
                    self._logger.warning(f"Error cancelling session: {e}")

            self._active_sessions.clear()

        # Restore original signal handler on Unix
        if not self._is_windows and self._original_signal_handler:
            with self._signal_lock:
                try:
                    signal.signal(signal.SIGALRM, self._original_signal_handler)
                    self._original_signal_handler = None
                except Exception as e:
                    self._logger.warning(f"Error restoring signal handler: {e}")

        self._logger.info("UnifiedTimeoutManager shutdown complete")


# Global instance
_timeout_manager = None


def get_timeout_manager() -> UnifiedTimeoutManager:
    """Get the global timeout manager instance"""
    global _timeout_manager
    if _timeout_manager is None:
        _timeout_manager = UnifiedTimeoutManager()
    return _timeout_manager


@contextlib.contextmanager
def hook_timeout_context(hook_name: str, config: Optional[HookTimeoutConfig] = None):
    """Context manager for hook execution with timeout management

    Usage:
        with hook_timeout_context("session_start", fast_config):
            result = perform_hook_operations()
    """
    manager = get_timeout_manager()
    session = manager.create_timeout_session(hook_name, config)

    try:
        manager.start_timeout(session)
        yield session
    except Exception as e:
        # Re-raise with enhanced context
        if isinstance(e, HookTimeoutError):
            raise
        else:
            raise
    finally:
        manager.cancel_timeout(session)
