"""
Phase-Optimized Hook Scheduler

Intelligent scheduling system for hooks based on development phases,
token budget constraints, and performance requirements.

Key Features:
- Phase-aware hook scheduling
- Token budget optimization
- Dynamic priority adjustment
- Performance-based scheduling
- Dependency resolution
- Resource management
"""

import asyncio
import threading
import time
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional, Set

from .jit_enhanced_hook_manager import (
    HookEvent,
    HookMetadata,
    HookPriority,
    JITEnhancedHookManager,
    Phase,
)


class SchedulingStrategy(Enum):
    """Hook scheduling strategies"""

    PRIORITY_FIRST = "priority_first"  # Execute highest priority hooks first
    PERFORMANCE_FIRST = "performance_first"  # Execute fastest hooks first
    PHASE_OPTIMIZED = "phase_optimized"  # Optimize for current phase
    TOKEN_EFFICIENT = "token_efficient"  # Minimize token usage
    BALANCED = "balanced"  # Balance all factors


class SchedulingDecision(Enum):
    """Scheduling decisions for hooks"""

    EXECUTE = "execute"  # Execute hook now
    DEFER = "defer"  # Defer to later
    SKIP = "skip"  # Skip execution
    PARALLEL = "parallel"  # Execute in parallel group
    SEQUENTIAL = "sequential"  # Execute sequentially


@dataclass
class HookSchedulingContext:
    """Context for hook scheduling decisions"""

    event_type: HookEvent
    current_phase: Phase
    user_input: str
    available_token_budget: int
    max_execution_time_ms: float
    current_time: datetime = field(default_factory=datetime.now)
    system_load: float = 0.5  # 0.0 (idle) to 1.0 (busy)
    recent_performance: Dict[str, float] = field(default_factory=dict)  # hook_path -> avg_time_ms
    active_dependencies: Set[str] = field(default_factory=set)


@dataclass
class ScheduledHook:
    """Hook scheduled for execution"""

    hook_path: str
    metadata: HookMetadata
    priority_score: float
    estimated_cost: int  # Token cost
    estimated_time_ms: float
    scheduling_decision: SchedulingDecision
    execution_group: Optional[int] = None  # Group ID for parallel execution
    dependencies: Set[str] = field(default_factory=set)
    dependents: Set[str] = field(default_factory=set)
    retry_count: int = 0
    max_retries: int = 2


@dataclass
class SchedulingResult:
    """Result of hook scheduling"""

    scheduled_hooks: List[ScheduledHook]
    execution_plan: List[List[ScheduledHook]]  # Groups of hooks to execute
    estimated_total_time_ms: float
    estimated_total_tokens: int
    skipped_hooks: List[ScheduledHook]
    deferred_hooks: List[ScheduledHook]
    scheduling_strategy: SchedulingStrategy


@dataclass
class ExecutionGroup:
    """Group of hooks to execute together"""

    group_id: int
    execution_type: SchedulingDecision  # PARALLEL or SEQUENTIAL
    hooks: List[ScheduledHook]
    estimated_time_ms: float
    estimated_tokens: int
    max_wait_time_ms: float
    dependencies: Set[str] = field(default_factory=set)


class PhaseOptimizedHookScheduler:
    """
    Intelligent scheduler for hooks with phase awareness and resource optimization

    Provides optimal hook execution scheduling based on:
    - Current development phase
    - Token budget constraints
    - Performance requirements
    - Hook dependencies
    - System load
    """

    def __init__(
        self,
        hook_manager: Optional[JITEnhancedHookManager] = None,
        default_strategy: SchedulingStrategy = SchedulingStrategy.PHASE_OPTIMIZED,
        max_parallel_groups: int = 3,
        enable_adaptive_scheduling: bool = True,
    ):
        """Initialize Phase-Optimized Hook Scheduler

        Args:
            hook_manager: JIT-Enhanced Hook Manager instance
            default_strategy: Default scheduling strategy
            max_parallel_groups: Maximum parallel execution groups
            enable_adaptive_scheduling: Enable adaptive strategy selection
        """
        self.hook_manager = hook_manager or JITEnhancedHookManager()
        self.default_strategy = default_strategy
        self.max_parallel_groups = max_parallel_groups
        self.enable_adaptive_scheduling = enable_adaptive_scheduling

        # Performance tracking
        self._scheduling_history: List[Dict[str, Any]] = []
        self._performance_cache: Dict[str, Dict[str, float]] = {}
        self._scheduling_lock = threading.Lock()

        # Phase-specific optimization parameters
        self._phase_parameters = self._initialize_phase_parameters()

        # Adaptive strategy selection
        self._strategy_performance: Dict[SchedulingStrategy, Dict[str, float]] = {
            strategy: {"success_rate": 1.0, "avg_efficiency": 0.8, "usage_count": 0} for strategy in SchedulingStrategy
        }

    def _initialize_phase_parameters(self) -> Dict[Phase, Dict[str, Any]]:
        """Initialize phase-specific optimization parameters"""
        return {
            Phase.SPEC: {
                "max_total_time_ms": 1000.0,
                "token_budget_ratio": 0.3,  # 30% of total budget
                "priority_weights": {
                    HookPriority.CRITICAL: 1.0,
                    HookPriority.HIGH: 0.9,
                    HookPriority.NORMAL: 0.6,
                    HookPriority.LOW: 0.2,
                },
                "prefer_parallel": False,  # Sequential for consistency
            },
            Phase.RED: {
                "max_total_time_ms": 800.0,
                "token_budget_ratio": 0.2,  # 20% of total budget
                "priority_weights": {
                    HookPriority.CRITICAL: 1.0,
                    HookPriority.HIGH: 1.0,  # Testing is high priority
                    HookPriority.NORMAL: 0.8,
                    HookPriority.LOW: 0.3,
                },
                "prefer_parallel": True,  # Parallel for faster test feedback
            },
            Phase.GREEN: {
                "max_total_time_ms": 600.0,
                "token_budget_ratio": 0.15,  # 15% of total budget
                "priority_weights": {
                    HookPriority.CRITICAL: 1.0,
                    HookPriority.HIGH: 0.8,
                    HookPriority.NORMAL: 0.7,
                    HookPriority.LOW: 0.4,
                },
                "prefer_parallel": True,  # Parallel for faster implementation
            },
            Phase.REFACTOR: {
                "max_total_time_ms": 1200.0,
                "token_budget_ratio": 0.2,  # 20% of total budget
                "priority_weights": {
                    HookPriority.CRITICAL: 1.0,
                    HookPriority.HIGH: 0.9,  # Code quality is important
                    HookPriority.NORMAL: 0.8,
                    HookPriority.LOW: 0.5,
                },
                "prefer_parallel": False,  # Sequential for safety
            },
            Phase.SYNC: {
                "max_total_time_ms": 1500.0,
                "token_budget_ratio": 0.1,  # 10% of total budget
                "priority_weights": {
                    HookPriority.CRITICAL: 1.0,
                    HookPriority.HIGH: 0.7,
                    HookPriority.NORMAL: 0.9,  # Documentation is important
                    HookPriority.LOW: 0.6,
                },
                "prefer_parallel": False,  # Sequential for consistency
            },
            Phase.DEBUG: {
                "max_total_time_ms": 500.0,
                "token_budget_ratio": 0.05,  # 5% of total budget
                "priority_weights": {
                    HookPriority.CRITICAL: 1.0,
                    HookPriority.HIGH: 0.9,  # Debug info is critical
                    HookPriority.NORMAL: 0.8,
                    HookPriority.LOW: 0.3,
                },
                "prefer_parallel": True,  # Parallel for faster debugging
            },
            Phase.PLANNING: {
                "max_total_time_ms": 800.0,
                "token_budget_ratio": 0.25,  # 25% of total budget
                "priority_weights": {
                    HookPriority.CRITICAL: 1.0,
                    HookPriority.HIGH: 0.8,
                    HookPriority.NORMAL: 0.7,
                    HookPriority.LOW: 0.4,
                },
                "prefer_parallel": False,  # Sequential for careful planning
            },
        }

    async def schedule_hooks(
        self,
        event_type: HookEvent,
        context: HookSchedulingContext,
        strategy: Optional[SchedulingStrategy] = None,
    ) -> SchedulingResult:
        """Schedule hooks for execution with optimization

        Args:
            event_type: Hook event type
            context: Scheduling context with phase and constraints
            strategy: Scheduling strategy (uses default if None)

        Returns:
            Scheduling result with execution plan
        """
        start_time = time.time()

        # Select optimal strategy
        selected_strategy = strategy or self._select_optimal_strategy(event_type, context)

        # Get available hooks for event type
        available_hooks = self.hook_manager._hooks_by_event.get(event_type, [])

        # Create scheduled hooks with initial analysis
        scheduled_hooks = await self._create_scheduled_hooks(available_hooks, context, selected_strategy)

        # Filter and prioritize hooks
        filtered_hooks = self._filter_hooks_by_constraints(scheduled_hooks, context)
        prioritized_hooks = self._prioritize_hooks(filtered_hooks, context, selected_strategy)

        # Resolve dependencies
        dependency_resolved_hooks = self._resolve_dependencies(prioritized_hooks)

        # Create execution groups
        execution_groups = self._create_execution_groups(dependency_resolved_hooks, context, selected_strategy)

        # Optimize execution order
        optimized_groups = self._optimize_execution_order(execution_groups, context)

        # Calculate estimates
        total_time_ms = sum(group.estimated_time_ms for group in optimized_groups)
        total_tokens = sum(group.estimated_tokens for group in optimized_groups)

        # Create execution plan
        execution_plan = [
            [hook for hook in group.hooks if hook.scheduling_decision == SchedulingDecision.EXECUTE]
            for group in optimized_groups
        ]

        # Separate skipped and deferred hooks
        executed_hooks = [hook for group in optimized_groups for hook in group.hooks]
        skipped_hooks = [
            hook
            for hook in scheduled_hooks
            if hook not in executed_hooks and hook.scheduling_decision == SchedulingDecision.SKIP
        ]
        deferred_hooks = [
            hook
            for hook in scheduled_hooks
            if hook not in executed_hooks and hook.scheduling_decision == SchedulingDecision.DEFER
        ]

        # Create scheduling result
        result = SchedulingResult(
            scheduled_hooks=scheduled_hooks,
            execution_plan=execution_plan,
            estimated_total_time_ms=total_time_ms,
            estimated_total_tokens=total_tokens,
            skipped_hooks=skipped_hooks,
            deferred_hooks=deferred_hooks,
            scheduling_strategy=selected_strategy,
        )

        # Update scheduling history and performance
        self._update_scheduling_history(result, time.time() - start_time)
        self._update_strategy_performance(selected_strategy, result, context)

        return result

    def _select_optimal_strategy(self, event_type: HookEvent, context: HookSchedulingContext) -> SchedulingStrategy:
        """Select optimal scheduling strategy based on context"""
        if not self.enable_adaptive_scheduling:
            return self.default_strategy

        # Strategy selection based on phase and constraints
        if context.available_token_budget < 5000:
            # Low token budget - prioritize token efficiency
            return SchedulingStrategy.TOKEN_EFFICIENT

        elif context.max_execution_time_ms < 500:
            # Tight time constraint - prioritize performance
            return SchedulingStrategy.PERFORMANCE_FIRST

        elif context.system_load > 0.8:
            # High system load - prioritize priority
            return SchedulingStrategy.PRIORITY_FIRST

        elif context.current_phase == Phase.SYNC:
            # Documentation phase - use phase-optimized
            return SchedulingStrategy.PHASE_OPTIMIZED

        else:
            # Use best performing strategy from history
            best_strategy = self.default_strategy
            best_efficiency = 0.0

            for strategy, performance in self._strategy_performance.items():
                if performance["usage_count"] > 0:
                    efficiency = performance["success_rate"] * performance["avg_efficiency"]
                    if efficiency > best_efficiency:
                        best_efficiency = efficiency
                        best_strategy = strategy

            return best_strategy

    async def _create_scheduled_hooks(
        self,
        hook_paths: List[str],
        context: HookSchedulingContext,
        strategy: SchedulingStrategy,
    ) -> List[ScheduledHook]:
        """Create scheduled hooks from hook paths with analysis"""
        scheduled_hooks = []

        for hook_path in hook_paths:
            metadata = self.hook_manager._hook_registry.get(hook_path)
            if not metadata:
                continue

            # Calculate priority score based on strategy
            priority_score = self._calculate_priority_score(metadata, context, strategy)

            # Estimate execution cost
            estimated_cost = self._estimate_hook_cost(metadata, context)
            estimated_time = self._estimate_hook_time(metadata, context)

            # Make initial scheduling decision
            scheduling_decision = self._make_initial_scheduling_decision(
                metadata, context, estimated_cost, estimated_time
            )

            scheduled_hook = ScheduledHook(
                hook_path=hook_path,
                metadata=metadata,
                priority_score=priority_score,
                estimated_cost=estimated_cost,
                estimated_time_ms=estimated_time,
                scheduling_decision=scheduling_decision,
                dependencies=metadata.dependencies.copy(),
                dependents=set(),
            )

            scheduled_hooks.append(scheduled_hook)

        return scheduled_hooks

    def _calculate_priority_score(
        self,
        metadata: HookMetadata,
        context: HookSchedulingContext,
        strategy: SchedulingStrategy,
    ) -> float:
        """Calculate priority score for hook based on strategy"""
        base_score = 0.0

        if strategy == SchedulingStrategy.PRIORITY_FIRST:
            # Priority-based scoring
            phase_params = self._phase_parameters.get(context.current_phase, {})
            priority_weights = phase_params.get("priority_weights", {})
            base_score = priority_weights.get(metadata.priority, 0.5) * 100

        elif strategy == SchedulingStrategy.PERFORMANCE_FIRST:
            # Performance-based scoring (faster hooks get higher priority)
            base_score = max(0, 100 - metadata.estimated_execution_time_ms / 10)

        elif strategy == SchedulingStrategy.PHASE_OPTIMIZED:
            # Phase relevance scoring
            phase_relevance = metadata.phase_relevance.get(context.current_phase, 0.5)
            base_score = phase_relevance * 100

            # Add priority bonus
            priority_bonus = (5 - metadata.priority.value) * 10
            base_score += priority_bonus

        elif strategy == SchedulingStrategy.TOKEN_EFFICIENT:
            # Token efficiency scoring (lower token cost gets higher priority)
            base_score = max(0, 100 - metadata.token_cost_estimate / 10)

        else:  # BALANCED
            # Combine all factors
            priority_score = (5 - metadata.priority.value) * 20
            phase_score = metadata.phase_relevance.get(context.current_phase, 0.5) * 30
            performance_score = max(0, 30 - metadata.estimated_execution_time_ms / 20)
            token_score = max(0, 20 - metadata.token_cost_estimate / 50)

            base_score = priority_score + phase_score + performance_score + token_score

        # Apply success rate modifier
        base_score *= metadata.success_rate

        return base_score

    def _estimate_hook_cost(self, metadata: HookMetadata, context: HookSchedulingContext) -> int:
        """Estimate token cost for hook execution"""
        base_cost = metadata.token_cost_estimate

        # Phase-based cost adjustment
        phase_relevance = metadata.phase_relevance.get(context.current_phase, 0.5)
        if phase_relevance < 0.3:
            # Low relevance - increase cost to discourage execution
            base_cost = int(base_cost * 2.0)
        elif phase_relevance > 0.8:
            # High relevance - reduce cost to encourage execution
            base_cost = int(base_cost * 0.7)

        # System load adjustment
        if context.system_load > 0.8:
            base_cost = int(base_cost * 1.3)  # Increase cost under high load

        return base_cost

    def _estimate_hook_time(self, metadata: HookMetadata, context: HookSchedulingContext) -> float:
        """Estimate execution time for hook"""
        base_time = metadata.estimated_execution_time_ms

        # System load adjustment
        load_factor = 1.0 + (context.system_load * 0.5)  # Up to 1.5x slower under load

        # Success rate adjustment (unreliable hooks might take longer due to retries)
        reliability_factor = 2.0 - metadata.success_rate  # 1.0 to 2.0

        return base_time * load_factor * reliability_factor

    def _make_initial_scheduling_decision(
        self,
        metadata: HookMetadata,
        context: HookSchedulingContext,
        estimated_cost: int,
        estimated_time: float,
    ) -> SchedulingDecision:
        """Make initial scheduling decision for hook"""
        # Critical hooks always execute
        if metadata.priority == HookPriority.CRITICAL:
            return SchedulingDecision.EXECUTE

        # Check constraints
        if estimated_cost > context.available_token_budget:
            return SchedulingDecision.SKIP

        if estimated_time > context.max_execution_time_ms:
            return SchedulingDecision.DEFER

        # Phase relevance check
        phase_relevance = metadata.phase_relevance.get(context.current_phase, 0.5)
        if phase_relevance < 0.2:
            return SchedulingDecision.SKIP

        # Success rate check
        if metadata.success_rate < 0.3:
            return SchedulingDecision.DEFER

        # System load check
        if context.system_load > 0.9 and metadata.priority != HookPriority.HIGH:
            return SchedulingDecision.DEFER

        # Default to execution
        return SchedulingDecision.EXECUTE

    def _filter_hooks_by_constraints(
        self, scheduled_hooks: List[ScheduledHook], context: HookSchedulingContext
    ) -> List[ScheduledHook]:
        """Filter hooks based on constraints"""
        filtered_hooks = []
        remaining_token_budget = context.available_token_budget
        remaining_time_budget = context.max_execution_time_ms

        # Sort by priority score (highest first)
        sorted_hooks = sorted(scheduled_hooks, key=lambda h: h.priority_score, reverse=True)

        for hook in sorted_hooks:
            if hook.scheduling_decision != SchedulingDecision.EXECUTE:
                filtered_hooks.append(hook)
                continue

            # Check if hook fits within constraints
            if hook.estimated_cost <= remaining_token_budget and hook.estimated_time_ms <= remaining_time_budget:
                filtered_hooks.append(hook)
                remaining_token_budget -= hook.estimated_cost
                remaining_time_budget -= hook.estimated_time_ms
            else:
                # Can't fit - mark as deferred
                hook.scheduling_decision = SchedulingDecision.DEFER
                filtered_hooks.append(hook)

        return filtered_hooks

    def _prioritize_hooks(
        self,
        scheduled_hooks: List[ScheduledHook],
        context: HookSchedulingContext,
        strategy: SchedulingStrategy,
    ) -> List[ScheduledHook]:
        """Prioritize hooks for execution"""
        # Separate by scheduling decision
        execute_hooks = [h for h in scheduled_hooks if h.scheduling_decision == SchedulingDecision.EXECUTE]
        other_hooks = [h for h in scheduled_hooks if h.scheduling_decision != SchedulingDecision.EXECUTE]

        # Sort execute hooks by priority (descending)
        execute_hooks.sort(key=lambda h: h.priority_score, reverse=True)

        # Sort other hooks by priority (ascending - least important first for skipping)
        other_hooks.sort(key=lambda h: h.priority_score)

        return execute_hooks + other_hooks

    def _resolve_dependencies(self, hooks: List[ScheduledHook]) -> List[ScheduledHook]:
        """Resolve hook dependencies and adjust execution order"""
        # Build dependency graph
        for hook in hooks:
            for dep_path in hook.dependencies:
                # Find dependent hook
                for other_hook in hooks:
                    if other_hook.hook_path == dep_path:
                        other_hook.dependents.add(hook.hook_path)
                        break

        # Topological sort considering dependencies
        resolved_hooks = []
        remaining_hooks = hooks.copy()

        while remaining_hooks:
            # Find hooks with no unresolved dependencies
            ready_hooks = [
                h
                for h in remaining_hooks
                if not any(dep in [rh.hook_path for rh in remaining_hooks] for dep in h.dependencies)
            ]

            if not ready_hooks:
                # Circular dependency - break by priority
                ready_hooks = sorted(remaining_hooks, key=lambda h: h.priority_score, reverse=True)
                ready_hooks = [ready_hooks[0]]

            # Add highest priority ready hook
            next_hook = max(ready_hooks, key=lambda h: h.priority_score)
            resolved_hooks.append(next_hook)
            remaining_hooks.remove(next_hook)

        return resolved_hooks

    def _create_execution_groups(
        self,
        hooks: List[ScheduledHook],
        context: HookSchedulingContext,
        strategy: SchedulingStrategy,
    ) -> List[ExecutionGroup]:
        """Create execution groups for optimal performance"""
        groups = []
        current_group = None
        group_id = 0

        # Get phase preferences
        phase_params = self._phase_parameters.get(context.current_phase, {})
        prefer_parallel = phase_params.get("prefer_parallel", True)

        for hook in hooks:
            if hook.scheduling_decision != SchedulingDecision.EXECUTE:
                continue

            # Determine execution type
            execution_type = (
                SchedulingDecision.PARALLEL
                if (prefer_parallel and hook.metadata.parallel_safe)
                else SchedulingDecision.SEQUENTIAL
            )

            # Create new group if needed
            if (
                current_group is None
                or current_group.execution_type != execution_type
                or len(current_group.hooks) >= self.max_parallel_groups
            ):
                current_group = ExecutionGroup(
                    group_id=group_id,
                    execution_type=execution_type,
                    hooks=[],
                    estimated_time_ms=0.0,
                    estimated_tokens=0,
                    max_wait_time_ms=0.0,
                    dependencies=set(),
                )
                groups.append(current_group)
                group_id += 1

            # Add hook to current group
            current_group.hooks.append(hook)
            current_group.estimated_time_ms += hook.estimated_time_ms
            current_group.estimated_tokens += hook.estimated_cost
            current_group.dependencies.update(hook.dependencies)

        # Set max wait time for parallel groups
        for group in groups:
            if group.execution_type == SchedulingDecision.PARALLEL:
                # For parallel groups, wait time is determined by slowest hook
                group.max_wait_time_ms = max(h.estimated_time_ms for h in group.hooks)
            else:
                # For sequential groups, sum all hook times
                group.max_wait_time_ms = group.estimated_time_ms

        return groups

    def _optimize_execution_order(
        self, groups: List[ExecutionGroup], context: HookSchedulingContext
    ) -> List[ExecutionGroup]:
        """Optimize execution order of groups"""
        if len(groups) <= 1:
            return groups

        # Sort groups by priority and execution type
        def group_score(group: ExecutionGroup) -> float:
            # Calculate average priority of hooks in group
            avg_priority = sum(h.priority_score for h in group.hooks) / len(group.hooks)

            # Parallel groups get bonus for speed
            parallel_bonus = 10 if group.execution_type == SchedulingDecision.PARALLEL else 0

            # Smaller groups get bonus for better time management
            size_bonus = max(0, 5 - len(group.hooks))

            return avg_priority + parallel_bonus + size_bonus

        # Sort by score (descending)
        optimized_groups = sorted(groups, key=group_score, reverse=True)

        return optimized_groups

    def _update_scheduling_history(self, result: SchedulingResult, planning_time_ms: float) -> None:
        """Update scheduling history for learning"""
        with self._scheduling_lock:
            history_entry: Dict[str, Any] = {
                "timestamp": datetime.now().isoformat(),
                "strategy": result.scheduling_strategy.value,
                "planning_time_ms": planning_time_ms,
                "total_hooks": len(result.scheduled_hooks),
                "executed_hooks": sum(len(group) for group in result.execution_plan),
                "skipped_hooks": len(result.skipped_hooks),
                "deferred_hooks": len(result.deferred_hooks),
                "estimated_time_ms": result.estimated_total_time_ms,
                "estimated_tokens": result.estimated_total_tokens,
            }

            self._scheduling_history.append(history_entry)

            # Keep history manageable
            if len(self._scheduling_history) > 1000:
                self._scheduling_history = self._scheduling_history[-500:]

    def _update_strategy_performance(
        self,
        strategy: SchedulingStrategy,
        result: SchedulingResult,
        context: HookSchedulingContext,
    ) -> None:
        """Update strategy performance metrics"""
        performance = self._strategy_performance[strategy]

        # Update usage count
        performance["usage_count"] += 1

        # Calculate efficiency (higher is better)
        total_hooks = len(result.scheduled_hooks)
        executed_hooks = sum(len(group) for group in result.execution_plan)

        if total_hooks > 0:
            execution_ratio = executed_hooks / total_hooks

            # Consider resource utilization
            time_efficiency = min(1.0, result.estimated_total_time_ms / context.max_execution_time_ms)
            token_efficiency = min(
                1.0,
                result.estimated_total_tokens / max(1, context.available_token_budget),
            )

            # Overall efficiency
            overall_efficiency = (execution_ratio * 0.5) + (time_efficiency * 0.3) + (token_efficiency * 0.2)

            # Update rolling average
            current_avg = performance["avg_efficiency"]
            performance["avg_efficiency"] = (current_avg * 0.8) + (overall_efficiency * 0.2)

        # Update success rate (assume successful if no critical errors)
        # This would be updated with actual execution results
        performance["success_rate"] = min(1.0, performance["success_rate"] * 0.95 + 0.05)

    def get_scheduling_statistics(self) -> Dict[str, Any]:
        """Get scheduling statistics and performance data"""
        with self._scheduling_lock:
            total_schedules = len(self._scheduling_history)

            if total_schedules == 0:
                return {
                    "total_schedules": 0,
                    "strategy_performance": {},
                    "recent_performance": [],
                }

            # Calculate strategy statistics
            strategy_stats = {}
            for strategy, performance in self._strategy_performance.items():
                if performance["usage_count"] > 0:
                    strategy_stats[strategy.value] = {
                        "usage_count": performance["usage_count"],
                        "success_rate": performance["success_rate"],
                        "avg_efficiency": performance["avg_efficiency"],
                        "recommendation_score": performance["success_rate"] * performance["avg_efficiency"],
                    }

            # Get recent performance
            recent_performance = self._scheduling_history[-10:] if total_schedules >= 10 else self._scheduling_history

            return {
                "total_schedules": total_schedules,
                "strategy_performance": strategy_stats,
                "recent_performance": recent_performance,
                "recommended_strategy": (
                    max(
                        strategy_stats.keys(),
                        key=lambda k: strategy_stats[k]["recommendation_score"],
                    )
                    if strategy_stats
                    else None
                ),
            }

    def get_phase_optimization_insights(self, phase: Phase) -> Dict[str, Any]:
        """Get insights for phase-specific optimization"""
        phase_params = self._phase_parameters.get(phase, {})

        # Analyze historical performance for this phase
        phase_schedules = [
            s
            for s in self._scheduling_history
            if any(phase.value.lower() in s.get("context", "").lower() for phase in [phase])
        ]

        insights: Dict[str, Any] = {
            "phase": phase.value,
            "parameters": phase_params,
            "historical_schedules": len(phase_schedules),
            "optimization_recommendations": [],
        }

        # Generate recommendations based on phase parameters
        if phase_params.get("prefer_parallel", False):
            recs_list = insights["optimization_recommendations"]
            if isinstance(recs_list, list):
                recs_list.append("This phase benefits from parallel hook execution")
        else:
            recs_list = insights["optimization_recommendations"]
            if isinstance(recs_list, list):
                recs_list.append("This phase prefers sequential hook execution for consistency")

        if phase_params.get("token_budget_ratio", 0) > 0.2:
            recs_list = insights["optimization_recommendations"]
            if isinstance(recs_list, list):
                recs_list.append("This phase requires significant token budget - consider token-efficient scheduling")

        # Add strategy recommendations
        strategy_stats = self.get_scheduling_statistics()["strategy_performance"]
        best_strategy = (
            max(strategy_stats.items(), key=lambda x: x[1]["recommendation_score"]) if strategy_stats else None
        )

        if best_strategy:
            insights["recommended_strategy"] = best_strategy[0]
            score = best_strategy[1]["recommendation_score"]
            insights["strategy_rationale"] = f"Best performance with {best_strategy[0]} strategy (score: {score:.2f})"

        return insights


# Convenience functions for common scheduling operations
async def schedule_session_start_hooks(
    context: HookSchedulingContext, strategy: Optional[SchedulingStrategy] = None
) -> SchedulingResult:
    """Schedule SessionStart hooks with phase optimization"""
    scheduler = PhaseOptimizedHookScheduler()
    return await scheduler.schedule_hooks(HookEvent.SESSION_START, context, strategy)


async def schedule_pre_tool_hooks(
    context: HookSchedulingContext, strategy: Optional[SchedulingStrategy] = None
) -> SchedulingResult:
    """Schedule PreToolUse hooks with phase optimization"""
    scheduler = PhaseOptimizedHookScheduler()
    return await scheduler.schedule_hooks(HookEvent.PRE_TOOL_USE, context, strategy)


def get_hook_scheduling_insights(phase: Phase) -> Dict[str, Any]:
    """Get phase-specific hook scheduling insights"""
    scheduler = PhaseOptimizedHookScheduler()
    return scheduler.get_phase_optimization_insights(phase)


if __name__ == "__main__":
    # Example usage and testing
    async def test_phase_optimized_scheduler():
        scheduler = PhaseOptimizedHookScheduler()

        # Create test context
        context = HookSchedulingContext(
            event_type=HookEvent.SESSION_START,
            current_phase=Phase.SPEC,
            user_input="Creating new specification for user authentication",
            available_token_budget=10000,
            max_execution_time_ms=1000.0,
        )

        # Schedule hooks
        result = await scheduler.schedule_hooks(HookEvent.SESSION_START, context)

        print(f"Scheduled {len(result.execution_plan)} execution groups")
        print(f"Estimated time: {result.estimated_total_time_ms:.1f}ms")
        print(f"Estimated tokens: {result.estimated_total_tokens}")
        print(f"Strategy: {result.scheduling_strategy.value}")

        # Show insights
        insights = scheduler.get_phase_optimization_insights(Phase.SPEC)
        print("\nPhase insights for SPEC:")
        for rec in insights.get("optimization_recommendations", []):
            print(f"  - {rec}")

        # Show statistics
        stats = scheduler.get_scheduling_statistics()
        print("\nScheduling statistics:")
        print(f"  Total schedules: {stats['total_schedules']}")
        if stats.get("recommended_strategy"):
            print(f"  Recommended strategy: {stats['recommended_strategy']}")

    # Run test
    asyncio.run(test_phase_optimized_scheduler())
