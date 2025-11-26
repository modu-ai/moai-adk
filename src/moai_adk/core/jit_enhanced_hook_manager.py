"""
JIT-Enhanced Hook Manager

Integrates Phase 2 JIT Context Loading System with Claude Code hook infrastructure
to provide intelligent, phase-aware hook execution with optimal performance.

Key Features:
- Phase-based hook optimization
- JIT context loading for hooks
- Intelligent skill filtering for hook operations
- Dynamic token budget management
- Real-time performance monitoring
- Smart caching and invalidation
"""

import asyncio
import json
import threading
import time
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, Set, Tuple

# Import JIT Context Loading System from Phase 2
try:
    from .jit_context_loader import (
        ContextCache as _ImportedContextCache,
    )
    from .jit_context_loader import (
        JITContextLoader as _ImportedJITContextLoader,
    )
    from .jit_context_loader import (
        Phase as _ImportedPhase,
    )
    from .jit_context_loader import (
        TokenBudgetManager as _ImportedTokenBudgetManager,
    )

    JITContextLoader = _ImportedJITContextLoader
    ContextCache = _ImportedContextCache
    TokenBudgetManager = _ImportedTokenBudgetManager
    Phase = _ImportedPhase
    _JIT_AVAILABLE = True
except ImportError:
    _JIT_AVAILABLE = False

    # Fallback for environments where JIT system might not be available
    class JITContextLoader:  # type: ignore[no-redef]
        def __init__(self, *args: Any, **kwargs: Any) -> None:
            pass

    class ContextCache:  # type: ignore[no-redef]
        def __init__(self, max_size: int = 100, max_memory_mb: int = 50) -> None:
            self.max_size = max_size
            self.max_memory_mb = max_memory_mb
            self.hits = 0
            self.misses = 0
            self.cache: dict[Any, Any] = {}

        def get(self, key: Any) -> Any:
            self.misses += 1
            return None

        def put(self, key: Any, value: Any, token_count: int = 0) -> None:
            pass

        def clear(self) -> None:
            pass

        def get_stats(self) -> dict[str, Any]:
            return {"hits": self.hits, "misses": self.misses}

    class TokenBudgetManager:  # type: ignore[no-redef]
        def __init__(self, *args: Any, **kwargs: Any) -> None:
            pass

    # Create Phase enum for hook system (fallback)
    class Phase(Enum):  # type: ignore[no-redef]
        SPEC = "SPEC"
        RED = "RED"
        GREEN = "GREEN"
        REFACTOR = "REFACTOR"
        SYNC = "SYNC"
        DEBUG = "DEBUG"
        PLANNING = "PLANNING"


class HookEvent(Enum):
    """Hook event types from Claude Code"""

    SESSION_START = "SessionStart"
    SESSION_END = "SessionEnd"
    USER_PROMPT_SUBMIT = "UserPromptSubmit"
    PRE_TOOL_USE = "PreToolUse"
    POST_TOOL_USE = "PostToolUse"
    SUBAGENT_START = "SubagentStart"
    SUBAGENT_STOP = "SubagentStop"


class HookPriority(Enum):
    """Hook execution priority levels"""

    CRITICAL = 1  # System-critical hooks (security, validation)
    HIGH = 2  # High-impact hooks (performance optimization)
    NORMAL = 3  # Standard hooks (logging, cleanup)
    LOW = 4  # Optional hooks (analytics, metrics)


@dataclass
class HookMetadata:
    """Metadata for a hook execution"""

    hook_path: str
    event_type: HookEvent
    priority: HookPriority
    estimated_execution_time_ms: float = 0.0
    last_execution_time: Optional[datetime] = None
    success_rate: float = 1.0
    phase_relevance: Dict[Phase, float] = field(default_factory=dict)
    token_cost_estimate: int = 0
    dependencies: Set[str] = field(default_factory=set)
    parallel_safe: bool = True


@dataclass
class HookExecutionResult:
    """Result of hook execution"""

    hook_path: str
    success: bool
    execution_time_ms: float
    token_usage: int
    output: Any
    error_message: Optional[str] = None
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class HookPerformanceMetrics:
    """Performance metrics for hook system"""

    total_executions: int = 0
    successful_executions: int = 0
    average_execution_time_ms: float = 0.0
    total_token_usage: int = 0
    cache_hits: int = 0
    cache_misses: int = 0
    phase_distribution: Dict[Phase, int] = field(default_factory=dict)
    event_type_distribution: Dict[HookEvent, int] = field(default_factory=dict)


class JITEnhancedHookManager:
    """
    Enhanced Hook Manager with JIT Context Loading System integration

    Provides intelligent hook execution with phase-aware optimization,
    token budget management, and performance monitoring.
    """

    def __init__(
        self,
        hooks_directory: Optional[Path] = None,
        cache_directory: Optional[Path] = None,
        max_concurrent_hooks: int = 5,
        enable_performance_monitoring: bool = True,
    ):
        """Initialize JIT-Enhanced Hook Manager

        Args:
            hooks_directory: Directory containing hook files
            cache_directory: Directory for hook cache and performance data
            max_concurrent_hooks: Maximum number of hooks to execute concurrently
            enable_performance_monitoring: Enable detailed performance tracking
        """
        self.hooks_directory = hooks_directory or Path.cwd() / ".claude" / "hooks"
        self.cache_directory = cache_directory or Path.cwd() / ".moai" / "cache" / "hooks"
        self.max_concurrent_hooks = max_concurrent_hooks
        self.enable_performance_monitoring = enable_performance_monitoring

        # Initialize JIT Context Loading System
        self.jit_loader = JITContextLoader()

        # Initialize caches and metadata storage
        self._initialize_caches()

        # Performance tracking
        self.metrics = HookPerformanceMetrics()
        self._performance_lock = threading.Lock()

        # Hook registry with metadata
        self._hook_registry: Dict[str, HookMetadata] = {}
        self._hooks_by_event: Dict[HookEvent, List[str]] = {}

        # Initialize hook registry
        self._discover_hooks()

    def _initialize_caches(self) -> None:
        """Initialize cache directories and data structures"""
        self.cache_directory.mkdir(parents=True, exist_ok=True)

        # Initialize hook result cache
        self._result_cache = ContextCache(max_size=100, max_memory_mb=50)

        # Initialize metadata cache
        self._metadata_cache: Dict[str, Dict[str, Any]] = {}

        # Performance log file
        self._performance_log_path = self.cache_directory / "performance.jsonl"

    def _discover_hooks(self) -> None:
        """Discover and register all available hooks"""
        if not self.hooks_directory.exists():
            return

        for hook_file in self.hooks_directory.rglob("*.py"):
            if hook_file.name.startswith("__") or hook_file.name.startswith("lib/"):
                continue

            hook_path_str = str(hook_file.relative_to(self.hooks_directory))

            # Extract event type from filename
            event_type = self._extract_event_type_from_filename(hook_file.name)
            if event_type:
                self._register_hook(hook_path_str, event_type)

    def _extract_event_type_from_filename(self, filename: str) -> Optional[HookEvent]:
        """Extract hook event type from filename pattern"""
        filename_lower = filename.lower()

        if "session_start" in filename_lower:
            return HookEvent.SESSION_START
        elif "session_end" in filename_lower:
            return HookEvent.SESSION_END
        elif "pre_tool" in filename_lower or "pretool" in filename_lower:
            return HookEvent.PRE_TOOL_USE
        elif "post_tool" in filename_lower or "posttool" in filename_lower:
            return HookEvent.POST_TOOL_USE
        elif "subagent_start" in filename_lower:
            return HookEvent.SUBAGENT_START
        elif "subagent_stop" in filename_lower:
            return HookEvent.SUBAGENT_STOP
        else:
            return None

    def _register_hook(self, hook_path: str, event_type: HookEvent) -> None:
        """Register a hook with metadata"""
        # Generate metadata based on hook characteristics
        metadata = HookMetadata(
            hook_path=hook_path,
            event_type=event_type,
            priority=self._determine_hook_priority(hook_path, event_type),
            estimated_execution_time_ms=self._estimate_execution_time(hook_path),
            phase_relevance=self._determine_phase_relevance(hook_path, event_type),
            token_cost_estimate=self._estimate_token_cost(hook_path),
            parallel_safe=self._is_parallel_safe(hook_path),
        )

        self._hook_registry[hook_path] = metadata

        if event_type not in self._hooks_by_event:
            self._hooks_by_event[event_type] = []
        self._hooks_by_event[event_type].append(hook_path)

    def _determine_hook_priority(self, hook_path: str, event_type: HookEvent) -> HookPriority:
        """Determine hook priority based on its characteristics"""
        filename = hook_path.lower()

        # Security and validation hooks are critical
        if any(keyword in filename for keyword in ["security", "validation", "health_check"]):
            return HookPriority.CRITICAL

        # Performance optimization hooks are high priority
        if any(keyword in filename for keyword in ["performance", "optimizer", "jit"]):
            return HookPriority.HIGH

        # Cleanup and logging hooks are normal priority
        if any(keyword in filename for keyword in ["cleanup", "log", "tracker"]):
            return HookPriority.NORMAL

        # Analytics and metrics are low priority
        if any(keyword in filename for keyword in ["analytics", "metrics", "stats"]):
            return HookPriority.LOW

        # Default priority based on event type
        if event_type == HookEvent.PRE_TOOL_USE:
            return HookPriority.HIGH  # Pre-execution validation is important
        elif event_type == HookEvent.SESSION_START:
            return HookPriority.NORMAL
        else:
            return HookPriority.NORMAL

    def _estimate_execution_time(self, hook_path: str) -> float:
        """Estimate hook execution time based on historical data and characteristics"""
        # Check cache for historical execution time
        cache_key = f"exec_time:{hook_path}"
        if cache_key in self._metadata_cache:
            cached_time = self._metadata_cache[cache_key].get("avg_time_ms")
            if cached_time:
                return cached_time

        # Estimate based on hook characteristics
        filename = hook_path.lower()

        # Hooks with git operations tend to be slower
        if "git" in filename:
            return 200.0  # 200ms estimate for git operations

        # Hooks with network operations are slower
        if any(keyword in filename for keyword in ["fetch", "api", "network"]):
            return 500.0  # 500ms estimate for network operations

        # Hooks with file I/O are moderate
        if any(keyword in filename for keyword in ["read", "write", "parse"]):
            return 50.0  # 50ms estimate for file I/O

        # Simple hooks are fast
        return 10.0  # 10ms estimate for simple operations

    def _determine_phase_relevance(self, hook_path: str, event_type: HookEvent) -> Dict[Phase, float]:
        """Determine hook relevance to different development phases"""
        filename = hook_path.lower()
        relevance = {}

        # Default relevance for all phases
        default_relevance = 0.5

        # SPEC phase relevance
        if any(keyword in filename for keyword in ["spec", "plan", "design", "requirement"]):
            relevance[Phase.SPEC] = 1.0
        else:
            relevance[Phase.SPEC] = default_relevance

        # RED phase relevance (testing)
        if any(keyword in filename for keyword in ["test", "red", "tdd", "assert"]):
            relevance[Phase.RED] = 1.0
        else:
            relevance[Phase.RED] = default_relevance

        # GREEN phase relevance (implementation)
        if any(keyword in filename for keyword in ["implement", "code", "green", "build"]):
            relevance[Phase.GREEN] = 1.0
        else:
            relevance[Phase.GREEN] = default_relevance

        # REFACTOR phase relevance
        if any(keyword in filename for keyword in ["refactor", "optimize", "improve", "clean"]):
            relevance[Phase.REFACTOR] = 1.0
        else:
            relevance[Phase.REFACTOR] = default_relevance

        # SYNC phase relevance (documentation)
        if any(keyword in filename for keyword in ["sync", "doc", "document", "deploy"]):
            relevance[Phase.SYNC] = 1.0
        else:
            relevance[Phase.SYNC] = default_relevance

        # DEBUG phase relevance
        if any(keyword in filename for keyword in ["debug", "error", "troubleshoot", "log"]):
            relevance[Phase.DEBUG] = 1.0
        else:
            relevance[Phase.DEBUG] = default_relevance

        # PLANNING phase relevance
        if any(keyword in filename for keyword in ["plan", "analysis", "strategy"]):
            relevance[Phase.PLANNING] = 1.0
        else:
            relevance[Phase.PLANNING] = default_relevance

        return relevance

    def _estimate_token_cost(self, hook_path: str) -> int:
        """Estimate token cost for hook execution"""
        # Base token cost for any hook
        base_cost = 100

        # Additional cost based on hook characteristics
        filename = hook_path.lower()

        if any(keyword in filename for keyword in ["analysis", "report", "generate"]):
            base_cost += 500  # Higher cost for analysis/generation
        elif any(keyword in filename for keyword in ["log", "simple", "basic"]):
            base_cost += 50  # Lower cost for simple operations

        return base_cost

    def _is_parallel_safe(self, hook_path: str) -> bool:
        """Determine if hook can be executed in parallel"""
        filename = hook_path.lower()

        # Hooks that modify shared state are not parallel safe
        if any(keyword in filename for keyword in ["write", "modify", "update", "delete"]):
            return False

        # Hooks with external dependencies might not be parallel safe
        if any(keyword in filename for keyword in ["database", "network", "api"]):
            return False

        # Most hooks are parallel safe by default
        return True

    async def execute_hooks(
        self,
        event_type: HookEvent,
        context: Dict[str, Any],
        user_input: Optional[str] = None,
        phase: Optional[Phase] = None,
        max_total_execution_time_ms: float = 1000.0,
    ) -> List[HookExecutionResult]:
        """Execute hooks for a specific event with JIT optimization

        Args:
            event_type: Type of hook event
            context: Execution context data
            user_input: User input for phase detection
            phase: Current development phase (if known)
            max_total_execution_time_ms: Maximum total execution time for all hooks

        Returns:
            List of hook execution results
        """
        start_time = time.time()

        # Detect phase if not provided
        if phase is None and user_input:
            try:
                phase = self.jit_loader.phase_detector.detect_phase(user_input)
            except AttributeError:
                # Fallback if JIT loader doesn't have phase detector
                phase = Phase.SPEC

        # Get relevant hooks for this event
        hook_paths = self._hooks_by_event.get(event_type, [])

        # Filter and prioritize hooks based on phase and performance
        prioritized_hooks = self._prioritize_hooks(hook_paths, phase)

        # Load optimized context using JIT system
        optimized_context = await self._load_optimized_context(event_type, context, phase, prioritized_hooks)

        # Execute hooks with optimization
        results = await self._execute_hooks_optimized(prioritized_hooks, optimized_context, max_total_execution_time_ms)

        # Update performance metrics
        if self.enable_performance_monitoring:
            self._update_performance_metrics(event_type, phase, results, start_time)

        return results

    def _prioritize_hooks(self, hook_paths: List[str], phase: Optional[Phase]) -> List[Tuple[str, float]]:
        """Prioritize hooks based on phase relevance and performance characteristics

        Args:
            hook_paths: List of hook file paths
            phase: Current development phase

        Returns:
            List of (hook_path, priority_score) tuples sorted by priority
        """
        hook_priorities = []

        for hook_path in hook_paths:
            metadata = self._hook_registry.get(hook_path)
            if not metadata:
                continue

            # Calculate priority score
            priority_score = 0.0

            # Base priority (lower number = higher priority)
            priority_score += metadata.priority.value * 10

            # Phase relevance bonus
            if phase and phase in metadata.phase_relevance:
                relevance = metadata.phase_relevance[phase]
                priority_score -= relevance * 5  # Higher relevance = lower score (higher priority)

            # Performance penalty (slower hooks get lower priority)
            priority_score += metadata.estimated_execution_time_ms / 100

            # Success rate bonus (more reliable hooks get higher priority)
            if metadata.success_rate < 0.9:
                priority_score += 5  # Penalize unreliable hooks

            hook_priorities.append((hook_path, priority_score))

        # Sort by priority score (lower is better)
        hook_priorities.sort(key=lambda x: x[1])

        return hook_priorities

    async def _load_optimized_context(
        self,
        event_type: HookEvent,
        context: Dict[str, Any],
        phase: Optional[Phase],
        prioritized_hooks: List[Tuple[str, float]],
    ) -> Dict[str, Any]:
        """Load optimized context using JIT system for hook execution

        Args:
            event_type: Hook event type
            context: Original context
            phase: Current development phase
            prioritized_hooks: List of prioritized hooks

        Returns:
            Optimized context with relevant information
        """
        # Create synthetic user input for context loading
        synthetic_input = f"Hook execution for {event_type.value}"
        if phase:
            synthetic_input += f" during {phase.value} phase"

        # Load context using JIT system
        try:
            jit_context, context_metrics = await self.jit_loader.load_context(
                user_input=synthetic_input, context=context
            )
        except (TypeError, AttributeError):
            # Fallback to basic context if JIT loader interface is different
            jit_context = context.copy()

        # Add hook-specific context
        optimized_context = jit_context.copy()
        optimized_context.update(
            {
                "hook_event_type": event_type.value,
                "hook_phase": phase.value if phase else None,
                "hook_execution_mode": "optimized",
                "prioritized_hooks": [hook_path for hook_path, _ in prioritized_hooks[:5]],  # Top 5 hooks
            }
        )

        return optimized_context

    async def _execute_hooks_optimized(
        self, prioritized_hooks: List[Tuple[str, float]], context: Dict[str, Any], max_total_execution_time_ms: float
    ) -> List[HookExecutionResult]:
        """Execute hooks with optimization and time management

        Args:
            prioritized_hooks: List of (hook_path, priority_score) tuples
            context: Optimized execution context
            max_total_execution_time_ms: Maximum total execution time

        Returns:
            List of hook execution results
        """
        results = []
        remaining_time = max_total_execution_time_ms

        # Separate hooks into parallel-safe and sequential
        parallel_hooks = []
        sequential_hooks = []

        for hook_path, _ in prioritized_hooks:
            metadata = self._hook_registry.get(hook_path)
            if metadata and metadata.parallel_safe:
                parallel_hooks.append(hook_path)
            else:
                sequential_hooks.append(hook_path)

        # Execute parallel hooks first (faster)
        if parallel_hooks and remaining_time > 0:
            parallel_results = await self._execute_hooks_parallel(parallel_hooks, context, remaining_time)
            results.extend(parallel_results)

            # Update remaining time
            total_parallel_time = sum(r.execution_time_ms for r in parallel_results)
            remaining_time -= total_parallel_time

        # Execute sequential hooks with remaining time
        if sequential_hooks and remaining_time > 0:
            sequential_results = await self._execute_hooks_sequential(sequential_hooks, context, remaining_time)
            results.extend(sequential_results)

        return results

    async def _execute_hooks_parallel(
        self, hook_paths: List[str], context: Dict[str, Any], max_total_time_ms: float
    ) -> List[HookExecutionResult]:
        """Execute hooks in parallel with time management"""
        results = []

        # Create semaphore to limit concurrent executions
        semaphore = asyncio.Semaphore(self.max_concurrent_hooks)

        async def execute_single_hook(hook_path: str) -> Optional[HookExecutionResult]:
            async with semaphore:
                try:
                    return await self._execute_single_hook(hook_path, context)
                except Exception as e:
                    return HookExecutionResult(
                        hook_path=hook_path,
                        success=False,
                        execution_time_ms=0.0,
                        token_usage=0,
                        output=None,
                        error_message=str(e),
                    )

        # Execute hooks with timeout
        tasks = [execute_single_hook(hook_path) for hook_path in hook_paths]

        try:
            # Wait for all hooks with total timeout
            completed_results = await asyncio.wait_for(
                asyncio.gather(*tasks, return_exceptions=True), timeout=max_total_time_ms / 1000.0
            )

            for result in completed_results:
                if isinstance(result, HookExecutionResult):
                    results.append(result)
                elif isinstance(result, Exception):
                    # Handle exceptions
                    error_result = HookExecutionResult(
                        hook_path="unknown",
                        success=False,
                        execution_time_ms=0.0,
                        token_usage=0,
                        output=None,
                        error_message=str(result),
                    )
                    results.append(error_result)

        except asyncio.TimeoutError:
            # Some hooks didn't complete in time
            pass

        return results

    async def _execute_hooks_sequential(
        self, hook_paths: List[str], context: Dict[str, Any], max_total_time_ms: float
    ) -> List[HookExecutionResult]:
        """Execute hooks sequentially with time management"""
        results = []
        remaining_time = max_total_time_ms

        for hook_path in hook_paths:
            if remaining_time <= 0:
                break

            try:
                result = await self._execute_single_hook(hook_path, context)
                results.append(result)

                # Update remaining time
                execution_time = result.execution_time_ms
                remaining_time -= execution_time

            except Exception as e:
                error_result = HookExecutionResult(
                    hook_path=hook_path,
                    success=False,
                    execution_time_ms=0.0,
                    token_usage=0,
                    output=None,
                    error_message=str(e),
                )
                results.append(error_result)

        return results

    async def _execute_single_hook(self, hook_path: str, context: Dict[str, Any]) -> HookExecutionResult:
        """Execute a single hook and return result

        Args:
            hook_path: Path to hook file
            context: Execution context

        Returns:
            Hook execution result
        """
        start_time = time.time()
        full_hook_path = self.hooks_directory / hook_path

        try:
            # Check cache for recent results
            cache_key = f"hook_result:{hook_path}:{hash(str(context))}"
            cached_entry = self._result_cache.get(cache_key)
            if cached_entry:
                # Extract HookExecutionResult from ContextEntry
                cached_result: HookExecutionResult = cached_entry.content if hasattr(cached_entry, "content") else cached_entry  # type: ignore[assignment]
                return cached_result

            # Prepare hook execution
            metadata = self._hook_registry.get(hook_path)
            if not metadata:
                raise ValueError(f"Hook metadata not found for {hook_path}")

            # Execute hook in subprocess for isolation
            result = await self._execute_hook_subprocess(full_hook_path, context, metadata)

            # Cache successful results
            if result.success:
                # Calculate token count for the result
                token_count = result.token_usage if hasattr(result, "token_usage") else 0
                self._result_cache.put(cache_key, result, token_count=token_count)

            # Update metadata
            self._update_hook_metadata(hook_path, result)

            return result

        except Exception as e:
            execution_time = (time.time() - start_time) * 1000

            return HookExecutionResult(
                hook_path=hook_path,
                success=False,
                execution_time_ms=execution_time,
                token_usage=0,
                output=None,
                error_message=str(e),
            )

    async def _execute_hook_subprocess(
        self, hook_path: Path, context: Dict[str, Any], metadata: HookMetadata
    ) -> HookExecutionResult:
        """Execute hook in isolated subprocess

        Args:
            hook_path: Full path to hook file
            context: Execution context
            metadata: Hook metadata

        Returns:
            Hook execution result
        """
        start_time = time.time()

        try:
            # Prepare input for hook
            hook_input = json.dumps(context)

            # Execute hook with timeout
            timeout_seconds = max(1.0, metadata.estimated_execution_time_ms / 1000.0)

            process = await asyncio.create_subprocess_exec(
                "uv",
                "run",
                str(hook_path),
                stdin=asyncio.subprocess.PIPE,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
                cwd=Path.cwd(),
            )

            try:
                stdout, stderr = await asyncio.wait_for(
                    process.communicate(input=hook_input.encode()), timeout=timeout_seconds
                )
            except asyncio.TimeoutError:
                process.kill()
                await process.wait()
                raise TimeoutError(f"Hook execution timed out after {timeout_seconds}s")

            execution_time_ms = (time.time() - start_time) * 1000
            success = process.returncode == 0

            # Parse output
            output = None
            if stdout:
                try:
                    output = json.loads(stdout.decode())
                except json.JSONDecodeError:
                    output = stdout.decode()

            error_message = None
            if stderr:
                error_message = stderr.decode().strip()
            elif process.returncode != 0:
                error_message = f"Hook exited with code {process.returncode}"

            return HookExecutionResult(
                hook_path=str(hook_path.relative_to(self.hooks_directory)),
                success=success,
                execution_time_ms=execution_time_ms,
                token_usage=metadata.token_cost_estimate,
                output=output,
                error_message=error_message,
            )

        except Exception as e:
            execution_time_ms = (time.time() - start_time) * 1000

            return HookExecutionResult(
                hook_path=str(hook_path.relative_to(self.hooks_directory)),
                success=False,
                execution_time_ms=execution_time_ms,
                token_usage=metadata.token_cost_estimate,
                output=None,
                error_message=str(e),
            )

    def _update_hook_metadata(self, hook_path: str, result: HookExecutionResult) -> None:
        """Update hook metadata based on execution result"""
        metadata = self._hook_registry.get(hook_path)
        if not metadata:
            return

        # Update execution time estimate
        cache_key = f"exec_time:{hook_path}"
        if cache_key not in self._metadata_cache:
            self._metadata_cache[cache_key] = {"count": 0, "total_time": 0.0}

        cache_entry = self._metadata_cache[cache_key]
        cache_entry["count"] += 1
        cache_entry["total_time"] += result.execution_time_ms
        cache_entry["avg_time_ms"] = cache_entry["total_time"] / cache_entry["count"]

        # Update success rate
        metadata.success_rate = (metadata.success_rate * 0.8) + (1.0 if result.success else 0.0) * 0.2
        metadata.last_execution_time = datetime.now()

    def _update_performance_metrics(
        self, event_type: HookEvent, phase: Optional[Phase], results: List[HookExecutionResult], start_time: float
    ) -> None:
        """Update performance metrics"""
        with self._performance_lock:
            self.metrics.total_executions += len(results)
            self.metrics.successful_executions += sum(1 for r in results if r.success)

            total_execution_time = sum(r.execution_time_ms for r in results)
            self.metrics.average_execution_time_ms = (self.metrics.average_execution_time_ms * 0.9) + (
                total_execution_time / len(results) * 0.1
            )

            self.metrics.total_token_usage += sum(r.token_usage for r in results)

            if phase:
                self.metrics.phase_distribution[phase] = self.metrics.phase_distribution.get(phase, 0) + 1

            self.metrics.event_type_distribution[event_type] = (
                self.metrics.event_type_distribution.get(event_type, 0) + 1
            )

            # Log performance data
            self._log_performance_data(event_type, phase, results, start_time)

    def _log_performance_data(
        self, event_type: HookEvent, phase: Optional[Phase], results: List[HookExecutionResult], start_time: float
    ) -> None:
        """Log performance data to file"""
        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "event_type": event_type.value,
            "phase": phase.value if phase else None,
            "total_hooks": len(results),
            "successful_hooks": sum(1 for r in results if r.success),
            "total_execution_time_ms": sum(r.execution_time_ms for r in results),
            "total_token_usage": sum(r.token_usage for r in results),
            "system_time_ms": (time.time() - start_time) * 1000,
            "results": [
                {
                    "hook_path": r.hook_path,
                    "success": r.success,
                    "execution_time_ms": r.execution_time_ms,
                    "token_usage": r.token_usage,
                    "error_message": r.error_message,
                }
                for r in results
            ],
        }

        try:
            with open(self._performance_log_path, "a") as f:
                f.write(json.dumps(log_entry) + "\n")
        except Exception:
            pass  # Silently fail on logging

    def get_performance_metrics(self) -> HookPerformanceMetrics:
        """Get current performance metrics"""
        with self._performance_lock:
            return HookPerformanceMetrics(
                total_executions=self.metrics.total_executions,
                successful_executions=self.metrics.successful_executions,
                average_execution_time_ms=self.metrics.average_execution_time_ms,
                total_token_usage=self.metrics.total_token_usage,
                cache_hits=self._result_cache.get_stats().get("hits", 0),
                cache_misses=self._result_cache.get_stats().get("misses", 0),
                phase_distribution=self.metrics.phase_distribution.copy(),
                event_type_distribution=self.metrics.event_type_distribution.copy(),
            )

    def get_hook_recommendations(
        self, event_type: Optional[HookEvent] = None, phase: Optional[Phase] = None
    ) -> Dict[str, Any]:
        """Get recommendations for hook optimization

        Args:
            event_type: Specific event type to analyze
            phase: Specific phase to analyze

        Returns:
            Dictionary with optimization recommendations
        """
        recommendations: Dict[str, List[Any]] = {
            "slow_hooks": [],
            "unreliable_hooks": [],
            "phase_mismatched_hooks": [],
            "optimization_suggestions": [],
        }

        # Analyze hook performance
        for hook_path, metadata in self._hook_registry.items():
            if event_type and metadata.event_type != event_type:
                continue

            # Check for slow hooks
            if metadata.estimated_execution_time_ms > 200:
                recommendations["slow_hooks"].append(
                    {
                        "hook_path": hook_path,
                        "estimated_time_ms": metadata.estimated_execution_time_ms,
                        "suggestion": "Consider optimizing or making this hook parallel-safe",
                    }
                )

            # Check for unreliable hooks
            if metadata.success_rate < 0.8:
                recommendations["unreliable_hooks"].append(
                    {
                        "hook_path": hook_path,
                        "success_rate": metadata.success_rate,
                        "suggestion": "Review error handling and improve reliability",
                    }
                )

            # Check for phase mismatch
            if phase:
                relevance = metadata.phase_relevance.get(phase, 0.0)
                if relevance < 0.3:
                    recommendations["phase_mismatched_hooks"].append(
                        {
                            "hook_path": hook_path,
                            "phase": phase.value,
                            "relevance": relevance,
                            "suggestion": "This hook may not be relevant for the current phase",
                        }
                    )

        # Generate optimization suggestions
        if recommendations["slow_hooks"]:
            recommendations["optimization_suggestions"].append(
                "Consider implementing caching for frequently executed slow hooks"
            )

        if recommendations["unreliable_hooks"]:
            recommendations["optimization_suggestions"].append(
                "Add retry logic and better error handling for unreliable hooks"
            )

        if recommendations["phase_mismatched_hooks"]:
            recommendations["optimization_suggestions"].append(
                "Use phase-based hook filtering to skip irrelevant hooks"
            )

        return recommendations

    async def cleanup(self) -> None:
        """Cleanup resources and save state"""
        # Save performance metrics
        metrics_file = self.cache_directory / "metrics.json"
        try:
            metrics_data = {
                "timestamp": datetime.now().isoformat(),
                "metrics": self.get_performance_metrics().__dict__,
                "hook_metadata": {
                    hook_path: {
                        "estimated_execution_time_ms": metadata.estimated_execution_time_ms,
                        "success_rate": metadata.success_rate,
                        "last_execution_time": (
                            metadata.last_execution_time.isoformat() if metadata.last_execution_time else None
                        ),
                    }
                    for hook_path, metadata in self._hook_registry.items()
                },
            }

            with open(metrics_file, "w") as f:
                json.dump(metrics_data, f, indent=2)
        except Exception:
            pass

        # Clear caches
        self._result_cache.clear()
        self._metadata_cache.clear()


# Global instance for easy access
_jit_hook_manager: Optional[JITEnhancedHookManager] = None


def get_jit_hook_manager() -> JITEnhancedHookManager:
    """Get or create global JIT hook manager instance"""
    global _jit_hook_manager
    if _jit_hook_manager is None:
        _jit_hook_manager = JITEnhancedHookManager()
    return _jit_hook_manager


# Convenience functions for common hook operations
async def execute_session_start_hooks(
    context: Dict[str, Any], user_input: Optional[str] = None
) -> List[HookExecutionResult]:
    """Execute SessionStart hooks with JIT optimization"""
    manager = get_jit_hook_manager()
    return await manager.execute_hooks(HookEvent.SESSION_START, context, user_input=user_input)


async def execute_pre_tool_hooks(
    context: Dict[str, Any], user_input: Optional[str] = None
) -> List[HookExecutionResult]:
    """Execute PreToolUse hooks with JIT optimization"""
    manager = get_jit_hook_manager()
    return await manager.execute_hooks(HookEvent.PRE_TOOL_USE, context, user_input=user_input)


async def execute_session_end_hooks(
    context: Dict[str, Any], user_input: Optional[str] = None
) -> List[HookExecutionResult]:
    """Execute SessionEnd hooks with JIT optimization"""
    manager = get_jit_hook_manager()
    return await manager.execute_hooks(HookEvent.SESSION_END, context, user_input=user_input)


def get_hook_performance_metrics() -> HookPerformanceMetrics:
    """Get current hook performance metrics"""
    manager = get_jit_hook_manager()
    return manager.get_performance_metrics()


def get_hook_optimization_recommendations(
    event_type: Optional[HookEvent] = None, phase: Optional[Phase] = None
) -> Dict[str, Any]:
    """Get hook optimization recommendations"""
    manager = get_jit_hook_manager()
    return manager.get_hook_recommendations(event_type, phase)


if __name__ == "__main__":
    # Example usage and testing
    async def test_jit_hook_manager():
        manager = JITEnhancedHookManager()

        # Test hook execution
        context = {"test": True, "user": "test_user"}
        results = await manager.execute_hooks(
            HookEvent.SESSION_START, context, user_input="Testing JIT enhanced hook system"
        )

        print(f"Executed {len(results)} hooks")
        for result in results:
            print(f"  {result.hook_path}: {'✓' if result.success else '✗'} ({result.execution_time_ms:.1f}ms)")

        # Show metrics
        metrics = manager.get_performance_metrics()
        print("\nPerformance Metrics:")
        print(f"  Total executions: {metrics.total_executions}")
        print(f"  Success rate: {metrics.successful_executions}/{metrics.total_executions}")
        print(f"  Avg execution time: {metrics.average_execution_time_ms:.1f}ms")

        # Cleanup
        await manager.cleanup()

    # Run test
    asyncio.run(test_jit_hook_manager())
