#!/usr/bin/env python3
"""Optimized Git Operations Manager

Provides efficient, thread-safe Git operations with connection pooling,
caching, and error handling to prevent performance bottlenecks and race conditions.

Features:
- Connection pooling for Git commands
- Intelligent caching with TTL
- Parallel execution with semaphore control
- Retry mechanisms for transient failures
- Comprehensive error handling and logging
- Cross-platform compatibility
"""

import hashlib
import logging
import subprocess
import threading
import time
from concurrent.futures import Future, ThreadPoolExecutor
from contextlib import contextmanager
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from enum import Enum
from pathlib import Path
from queue import Empty, Queue
from typing import Any, Callable, Dict, List, Optional, Union


class GitOperationType(Enum):
    """Types of Git operations for caching and optimization"""

    BRANCH = "branch"
    COMMIT = "commit"
    STATUS = "status"
    LOG = "log"
    DIFF = "diff"
    REMOTE = "remote"
    CONFIG = "config"


@dataclass
class GitCommand:
    """Git command specification"""

    operation_type: GitOperationType
    args: List[str]
    cache_ttl_seconds: int = 60
    retry_count: int = 2
    timeout_seconds: int = 10


@dataclass
class GitResult:
    """Result of Git operation with metadata"""

    success: bool
    stdout: str = ""
    stderr: str = ""
    return_code: int = -1
    execution_time: float = 0.0
    cached: bool = False
    cache_hit: bool = False
    operation_type: Optional[GitOperationType] = None
    command: List[str] = field(default_factory=list)


@dataclass
class CacheEntry:
    """Cache entry for Git results"""

    result: GitResult
    timestamp: datetime
    ttl: timedelta
    hit_count: int = 0


class GitOperationError(Exception):
    """Enhanced Git operation error with context"""

    def __init__(
        self,
        message: str,
        command: List[str] = None,
        return_code: int = -1,
        stderr: str = "",
        execution_time: float = 0.0,
    ):
        super().__init__(message)
        self.command = command or []
        self.return_code = return_code
        self.stderr = stderr
        self.execution_time = execution_time


class GitOperationsManager:
    """Optimized Git operations manager with connection pooling and caching

    Features:
    - Thread-safe execution with semaphore control
    - Intelligent caching based on operation type
    - Connection pooling to prevent resource contention
    - Retry mechanisms for transient failures
    - Performance monitoring and logging
    """

    def __init__(self, max_workers: int = 4, cache_size_limit: int = 100):
        self._logger = logging.getLogger(__name__)

        # Thread pool for parallel Git operations
        self._executor = ThreadPoolExecutor(max_workers=max_workers, thread_name_prefix="git_ops")
        self._semaphore = threading.Semaphore(max_workers)  # Limit concurrent Git operations

        # Cache management
        self._cache: Dict[str, CacheEntry] = {}
        self._cache_lock = threading.RLock()
        self._cache_size_limit = cache_size_limit

        # Performance tracking
        self._operation_stats = {
            "total_operations": 0,
            "cache_hits": 0,
            "cache_misses": 0,
            "errors": 0,
            "total_time": 0.0,
        }
        self._stats_lock = threading.Lock()

        # Git command queue for sequential operations when needed
        self._command_queue: "Queue[Any]" = Queue()
        self._queue_processor_thread: Optional[threading.Thread] = None
        self._queue_active = True

        # Start queue processor thread
        self._start_queue_processor()

        self._logger.info(f"GitOperationsManager initialized with max_workers={max_workers}")

    def _start_queue_processor(self) -> None:
        """Start background thread to process queued commands"""

        def process_queue():
            while self._queue_active:
                try:
                    # Wait for command with timeout
                    future: Future = self._command_queue.get(timeout=1.0)
                    try:
                        # Execute the command
                        command, callback = future
                        result = self._execute_git_command_unsafe(command)
                        if callback:
                            callback(result)
                    except Exception as e:
                        self._logger.error(f"Error processing queued command: {e}")
                    finally:
                        self._command_queue.task_done()
                except Empty:
                    continue
                except Exception as e:
                    self._logger.error(f"Queue processor error: {e}")

        self._queue_processor_thread = threading.Thread(target=process_queue, daemon=True)
        self._queue_processor_thread.start()

    def _generate_cache_key(self, operation_type: GitOperationType, args: List[str]) -> str:
        """Generate cache key for Git operation"""
        # Include current working directory and branch for context-aware caching
        try:
            cwd = str(Path.cwd())
            # Simple branch detection for cache key
            branch_info = ""
            if operation_type in [GitOperationType.STATUS, GitOperationType.DIFF]:
                try:
                    result = subprocess.run(
                        ["git", "rev-parse", "--abbrev-ref", "HEAD"],
                        capture_output=True,
                        text=True,
                        timeout=2,
                    )
                    if result.returncode == 0:
                        branch_info = result.stdout.strip()
                except Exception:
                    pass

            key_data = f"{operation_type.value}:{args}:{cwd}:{branch_info}"
        except Exception:
            # Fallback to simple key
            key_data = f"{operation_type.value}:{args}"

        return hashlib.md5(key_data.encode()).hexdigest()

    def _is_cache_valid(self, entry: CacheEntry) -> bool:
        """Check if cache entry is still valid"""
        return datetime.now() - entry.timestamp < entry.ttl

    def _cleanup_cache(self) -> int:
        """Clean up expired cache entries and enforce size limit"""
        with self._cache_lock:
            # Remove expired entries
            expired_keys = [key for key, entry in self._cache.items() if not self._is_cache_valid(entry)]
            for key in expired_keys:
                del self._cache[key]

            # Enforce size limit (remove least recently used)
            if len(self._cache) > self._cache_size_limit:
                # Sort by last access time (hit count as proxy)
                sorted_items = sorted(self._cache.items(), key=lambda x: (x[1].hit_count, x[1].timestamp))
                items_to_remove = len(self._cache) - self._cache_size_limit
                for key, _ in sorted_items[:items_to_remove]:
                    del self._cache[key]

            return len(expired_keys)

    def _get_from_cache(self, cache_key: str) -> Optional[GitResult]:
        """Get result from cache if valid"""
        with self._cache_lock:
            if cache_key in self._cache:
                entry = self._cache[cache_key]
                if self._is_cache_valid(entry):
                    entry.hit_count += 1
                    # Return a copy of the result
                    result = GitResult(
                        success=entry.result.success,
                        stdout=entry.result.stdout,
                        stderr=entry.result.stderr,
                        return_code=entry.result.return_code,
                        execution_time=entry.result.execution_time,
                        cached=True,
                        cache_hit=True,
                        operation_type=entry.result.operation_type,
                        command=entry.result.command.copy(),
                    )
                    return result
                else:
                    # Remove expired entry
                    del self._cache[cache_key]

        return None

    def _store_in_cache(self, cache_key: str, result: GitResult, ttl: int) -> None:
        """Store result in cache with TTL"""
        with self._cache_lock:
            self._cache[cache_key] = CacheEntry(result=result, timestamp=datetime.now(), ttl=timedelta(seconds=ttl))

        # Cleanup if cache is getting large
        if len(self._cache) > self._cache_size_limit * 0.8:
            self._cleanup_cache()

    def _execute_git_command_unsafe(self, command: GitCommand) -> GitResult:
        """Execute Git command without semaphore control (internal use)"""
        start_time = time.time()
        full_command = ["git"] + command.args

        try:
            # Execute Git command
            result = subprocess.run(
                full_command,
                capture_output=True,
                text=True,
                timeout=command.timeout_seconds,
                cwd=Path.cwd(),
            )

            execution_time = time.time() - start_time

            git_result = GitResult(
                success=result.returncode == 0,
                stdout=result.stdout.strip(),
                stderr=result.stderr.strip(),
                return_code=result.returncode,
                execution_time=execution_time,
                cached=False,
                cache_hit=False,
                operation_type=command.operation_type,
                command=full_command.copy(),
            )

            # Update statistics
            with self._stats_lock:
                self._operation_stats["total_operations"] += 1
                self._operation_stats["total_time"] += execution_time

                if result.returncode != 0:
                    self._operation_stats["errors"] += 1

            return git_result

        except subprocess.TimeoutExpired:
            execution_time = time.time() - start_time
            error_msg = f"Git command timed out after {command.timeout_seconds}s: {' '.join(full_command)}"
            self._logger.error(error_msg)

            with self._stats_lock:
                self._operation_stats["total_operations"] += 1
                self._operation_stats["errors"] += 1

            return GitResult(
                success=False,
                stderr=error_msg,
                execution_time=execution_time,
                operation_type=command.operation_type,
                command=full_command.copy(),
            )

        except Exception as e:
            execution_time = time.time() - start_time
            error_msg = f"Git command failed: {e}"
            self._logger.error(error_msg)

            with self._stats_lock:
                self._operation_stats["total_operations"] += 1
                self._operation_stats["errors"] += 1

            return GitResult(
                success=False,
                stderr=error_msg,
                execution_time=execution_time,
                operation_type=command.operation_type,
                command=full_command.copy(),
            )

    def execute_git_command(self, command: Union[GitCommand, str], *args) -> GitResult:
        """Execute Git command with caching and retry logic"""
        # Convert string command to GitCommand
        if isinstance(command, str):
            command = GitCommand(
                operation_type=GitOperationType.CONFIG,  # Default
                args=[command] + list(args),
            )

        # Check cache first
        cache_key = self._generate_cache_key(command.operation_type, command.args)
        cached_result = self._get_from_cache(cache_key)
        if cached_result:
            with self._stats_lock:
                self._operation_stats["cache_hits"] += 1
            return cached_result

        with self._stats_lock:
            self._operation_stats["cache_misses"] += 1

        # Execute with retry logic
        last_result = None
        for attempt in range(command.retry_count + 1):
            try:
                with self._semaphore:  # Limit concurrent Git operations
                    result = self._execute_git_command_unsafe(command)

                    if result.success:
                        # Cache successful result
                        self._store_in_cache(cache_key, result, command.cache_ttl_seconds)
                        return result
                    else:
                        last_result = result
                        if attempt < command.retry_count:
                            self._logger.warning(
                                f"Git command failed, retrying ({attempt + 1}/{command.retry_count}): {result.stderr}"
                            )
                            time.sleep(0.1 * (attempt + 1))  # Exponential backoff
                        else:
                            self._logger.error(f"Git command failed after retries: {result.stderr}")

            except Exception as e:
                self._logger.error(f"Git command exception (attempt {attempt + 1}): {e}")
                if attempt == command.retry_count:
                    return GitResult(
                        success=False,
                        stderr=str(e),
                        operation_type=command.operation_type,
                        command=["git"] + command.args,
                    )
                time.sleep(0.1 * (attempt + 1))

        return last_result or GitResult(success=False, stderr="Unknown error")

    def execute_parallel(self, commands: List[GitCommand]) -> List[GitResult]:
        """Execute multiple Git commands in parallel with controlled concurrency"""
        futures = []
        results = []

        # Submit all commands
        for command in commands:
            future = self._executor.submit(self.execute_git_command, command)
            futures.append((future, command))

        # Collect results as they complete
        for future, command in futures:
            try:
                result = future.result(timeout=command.timeout_seconds + 5)  # Extra buffer
                results.append(result)
            except Exception as e:
                self._logger.error(f"Parallel Git command failed: {e}")
                results.append(
                    GitResult(
                        success=False,
                        stderr=str(e),
                        operation_type=command.operation_type,
                        command=["git"] + command.args,
                    )
                )

        return results

    def get_project_info(self, use_cache: bool = True) -> Dict[str, Any]:
        """Get comprehensive project information efficiently"""
        commands = [
            GitCommand(
                operation_type=GitOperationType.BRANCH,
                args=["branch", "--show-current"],
                cache_ttl_seconds=30,
                timeout_seconds=5,
            ),
            GitCommand(
                operation_type=GitOperationType.LOG,
                args=["log", "--pretty=format:%h %s", "-1"],
                cache_ttl_seconds=10,
                timeout_seconds=5,
            ),
            GitCommand(
                operation_type=GitOperationType.LOG,
                args=["log", "--pretty=format:%ar", "-1"],
                cache_ttl_seconds=10,
                timeout_seconds=5,
            ),
            GitCommand(
                operation_type=GitOperationType.STATUS,
                args=["status", "--porcelain"],
                cache_ttl_seconds=5,  # Short TTL for status
                timeout_seconds=5,
            ),
        ]

        # If cache disabled, clear relevant entries
        if not use_cache:
            with self._cache_lock:
                for command in commands:
                    cache_key = self._generate_cache_key(command.operation_type, command.args)
                    self._cache.pop(cache_key, None)

        # Execute commands in parallel
        results = self.execute_parallel(commands)

        # Process results
        project_info = {
            "branch": "unknown",
            "last_commit": "unknown",
            "commit_time": "unknown",
            "changes": 0,
            "fetch_time": datetime.now().isoformat(),
        }

        if len(results) >= 4:
            if results[0].success:
                project_info["branch"] = results[0].stdout or "unknown"
            if results[1].success:
                project_info["last_commit"] = results[1].stdout or "unknown"
            if results[2].success:
                project_info["commit_time"] = results[2].stdout or "unknown"
            if results[3].success:
                changes_text = results[3].stdout.strip()
                project_info["changes"] = len(changes_text.splitlines()) if changes_text else 0

        return project_info

    def queue_command(
        self,
        command: GitCommand,
        callback: Optional[Callable[[GitResult], None]] = None,
    ) -> None:
        """Queue a Git command for background execution"""
        try:
            future = (command, callback)
            self._command_queue.put(future, timeout=1.0)
        except Exception as e:
            self._logger.error(f"Failed to queue Git command: {e}")

    def get_statistics(self) -> Dict[str, Any]:
        """Get performance and cache statistics"""
        with self._stats_lock, self._cache_lock:
            return {
                "operations": {
                    "total": self._operation_stats["total_operations"],
                    "cache_hits": self._operation_stats["cache_hits"],
                    "cache_misses": self._operation_stats["cache_misses"],
                    "cache_hit_rate": (
                        self._operation_stats["cache_hits"]
                        / (self._operation_stats["cache_hits"] + self._operation_stats["cache_misses"])
                        if (self._operation_stats["cache_hits"] + self._operation_stats["cache_misses"]) > 0
                        else 0
                    ),
                    "errors": self._operation_stats["errors"],
                    "average_execution_time": (
                        self._operation_stats["total_time"] / self._operation_stats["total_operations"]
                        if self._operation_stats["total_operations"] > 0
                        else 0
                    ),
                },
                "cache": {
                    "size": len(self._cache),
                    "size_limit": self._cache_size_limit,
                    "utilization": len(self._cache) / self._cache_size_limit,
                },
                "queue": {"pending": self._command_queue.qsize()},
            }

    def clear_cache(self, operation_type: Optional[GitOperationType] = None) -> int:
        """Clear cache entries, optionally filtered by operation type"""
        with self._cache_lock:
            if operation_type is None:
                # Clear all cache
                count = len(self._cache)
                self._cache.clear()
            else:
                # Clear specific operation type
                keys_to_remove = [
                    key for key, entry in self._cache.items() if entry.result.operation_type == operation_type
                ]
                for key in keys_to_remove:
                    del self._cache[key]
                count = len(keys_to_remove)

            return count

    def shutdown(self) -> None:
        """Shutdown the Git operations manager"""
        self._logger.info("Shutting down GitOperationsManager")

        # Stop queue processor
        self._queue_active = False
        if self._queue_processor_thread and self._queue_processor_thread.is_alive():
            self._queue_processor_thread.join(timeout=2.0)

        # Shutdown thread pool
        self._executor.shutdown(wait=True)

        # Clear cache
        with self._cache_lock:
            self._cache.clear()

        self._logger.info("GitOperationsManager shutdown complete")


# Global instance
_git_manager = None
_git_manager_lock = threading.Lock()


def get_git_manager() -> GitOperationsManager:
    """Get the global Git operations manager instance"""
    global _git_manager
    if _git_manager is None:
        with _git_manager_lock:
            if _git_manager is None:
                _git_manager = GitOperationsManager()
    return _git_manager


@contextmanager
def git_operation_context(max_workers: int = 4):
    """Context manager for Git operations with temporary manager"""
    manager = GitOperationsManager(max_workers=max_workers)
    try:
        yield manager
    finally:
        manager.shutdown()


# Convenience functions for common operations
def get_git_info(use_cache: bool = True) -> Dict[str, Any]:
    """Convenience function to get Git project information"""
    manager = get_git_manager()
    return manager.get_project_info(use_cache=use_cache)


def run_git_command(
    operation_type: GitOperationType,
    args: List[str],
    cache_ttl: int = 60,
    timeout: int = 10,
) -> GitResult:
    """Convenience function to run a single Git command"""
    command = GitCommand(
        operation_type=operation_type,
        args=args,
        cache_ttl_seconds=cache_ttl,
        timeout_seconds=timeout,
    )
    manager = get_git_manager()
    return manager.execute_git_command(command)
