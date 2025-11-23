# Token Budget Management

Allocate and manage token budgets across session components for optimal resource utilization.

## Core Budget Management Strategies

### Strategy 1: Budget Allocation Framework

Systematically allocate token budget across session components based on priority and usage patterns.

**Standard Allocation**:

```python
# Default token budget allocation (200K total)
standard_allocations = {
    "system_context": 0.10,      # 10% (20K) - System prompts and configuration
    "working_memory": 0.30,      # 30% (60K) - Active context and files
    "knowledge_base": 0.20,      # 20% (40K) - Skills and documentation
    "agent_context": 0.15,       # 15% (30K) - Agent states and execution
    "interaction_buffer": 0.25   # 25% (50K) - Conversation history
}
```

**Dynamic Budget Adjustment**:

```python
def adjust_allocations(usage_history: dict, current_allocations: dict) -> dict:
    """Adjust allocations based on historical usage patterns.

    Args:
        usage_history: Historical usage statistics per component
        current_allocations: Current allocation percentages

    Returns:
        Adjusted allocation percentages

    Adjustment Logic:
        - If component exhausted >3 times: +5% allocation
        - If component underutilized (<50%): -5% allocation
        - Rebalance across all components
        - Maintain total = 100%
    """
    adjusted = current_allocations.copy()

    # Phase 1: Identify exhausted components
    for component, history in usage_history.items():
        if history["exhaustion_count"] > 3:
            # Increase allocation by 5%
            adjusted[component] = min(adjusted[component] + 0.05, 0.50)

    # Phase 2: Identify underutilized components
    for component, history in usage_history.items():
        if history["avg_utilization"] < 0.50:
            # Decrease allocation by 5%
            adjusted[component] = max(adjusted[component] - 0.05, 0.05)

    # Phase 3: Normalize to 100%
    total = sum(adjusted.values())
    adjusted = {k: v / total for k, v in adjusted.items()}

    return adjusted

# Example adjustment based on usage
usage_history = {
    "system_context": {"exhaustion_count": 0, "avg_utilization": 0.98},
    "working_memory": {"exhaustion_count": 0, "avg_utilization": 0.75},
    "knowledge_base": {"exhaustion_count": 2, "avg_utilization": 0.95},
    "agent_context": {"exhaustion_count": 5, "avg_utilization": 0.98},
    "interaction_buffer": {"exhaustion_count": 0, "avg_utilization": 0.45}
}

# Adjusted allocations
adjusted_allocations = adjust_allocations(usage_history, standard_allocations)
# Result:
# agent_context: 0.20 (+5% due to frequent exhaustion)
# interaction_buffer: 0.20 (-5% due to underutilization)
```

---

### Strategy 2: Component Cleanup Strategies

Implement targeted cleanup strategies for each budget component.

**Working Memory Cleanup**:

```python
def cleanup_working_memory(session: SessionState) -> CleanupResult:
    """Free up working memory tokens.

    Cleanup Strategy:
        1. Compress large inactive contexts
        2. Remove old debugging information
        3. Archive completed task contexts
        4. Evict least recently used files

    Returns:
        Cleanup result with tokens freed
    """
    tokens_freed = 0
    start_time = datetime.now()

    # Step 1: Compress large inactive contexts (>10K tokens, not accessed in 1hr)
    for context in session.working_memory.contexts:
        if context.token_count > 10000 and context.last_accessed < datetime.now() - timedelta(hours=1):
            original_size = context.token_count
            context.compress(target_ratio=0.4)
            tokens_freed += (original_size - context.token_count)

    # Step 2: Remove old debugging information (>2 hours old)
    debug_info = session.working_memory.get_debug_info()
    for item in debug_info:
        if item.created_at < datetime.now() - timedelta(hours=2):
            tokens_freed += item.token_count
            session.working_memory.remove_debug_info(item)

    # Step 3: Archive completed task contexts
    completed_tasks = session.working_memory.get_completed_tasks()
    for task_context in completed_tasks:
        tokens_freed += task_context.token_count
        session.archive_task_context(task_context)
        session.working_memory.remove_task_context(task_context)

    # Step 4: Evict LRU files (if still above threshold)
    if session.working_memory.utilization() > 0.75:
        lru_files = session.working_memory.get_lru_files(count=5)
        for file_context in lru_files:
            tokens_freed += file_context.token_count
            session.working_memory.evict_file(file_context)

    duration = (datetime.now() - start_time).total_seconds() * 1000

    return CleanupResult(
        component="working_memory",
        tokens_freed=tokens_freed,
        duration_ms=duration,
        actions_taken=["compress_inactive", "remove_debug", "archive_tasks", "evict_lru"]
    )
```

**Knowledge Base Cleanup**:

```python
def cleanup_knowledge_base(session: SessionState) -> CleanupResult:
    """Free up knowledge base tokens.

    Cleanup Strategy:
        1. Unload unused skill modules
        2. Cache documentation excerpts instead of full docs
        3. Remove duplicate knowledge
        4. Compress skill documentation

    Returns:
        Cleanup result with tokens freed
    """
    tokens_freed = 0
    start_time = datetime.now()

    # Step 1: Unload unused skill modules (not accessed in 3 hours)
    for skill in session.knowledge_base.loaded_skills:
        if skill.last_accessed < datetime.now() - timedelta(hours=3):
            tokens_freed += skill.token_count
            session.knowledge_base.unload_skill(skill)

    # Step 2: Cache documentation excerpts (keep summaries only)
    for doc in session.knowledge_base.loaded_docs:
        if doc.is_full_loaded():
            original_size = doc.token_count
            doc.keep_summary_only()  # Keep 20% summary
            tokens_freed += (original_size - doc.token_count)

    # Step 3: Remove duplicate knowledge
    duplicates = session.knowledge_base.find_duplicates()
    for duplicate in duplicates:
        tokens_freed += duplicate.token_count
        session.knowledge_base.remove_duplicate(duplicate)

    # Step 4: Compress skill documentation
    for skill in session.knowledge_base.loaded_skills:
        if skill.has_large_docs():
            original_size = skill.doc_token_count
            skill.compress_docs(target_ratio=0.3)
            tokens_freed += (original_size - skill.doc_token_count)

    duration = (datetime.now() - start_time).total_seconds() * 1000

    return CleanupResult(
        component="knowledge_base",
        tokens_freed=tokens_freed,
        duration_ms=duration,
        actions_taken=["unload_skills", "cache_docs", "remove_duplicates", "compress_docs"]
    )
```

**Interaction Buffer Cleanup**:

```python
def cleanup_interaction_buffer(session: SessionState) -> CleanupResult:
    """Free up conversation history tokens.

    Cleanup Strategy:
        1. Summarize old interactions (>2 hours)
        2. Keep full text for recent interactions (<30 min)
        3. Remove redundant exchanges
        4. Archive to long-term storage

    Returns:
        Cleanup result with tokens freed
    """
    tokens_freed = 0
    start_time = datetime.now()

    # Step 1: Summarize old interactions (>2 hours old)
    old_threshold = datetime.now() - timedelta(hours=2)
    for interaction in session.interaction_buffer.history:
        if interaction.timestamp < old_threshold:
            original_size = interaction.token_count
            interaction.summarize(max_length=200)
            tokens_freed += (original_size - interaction.token_count)

    # Step 2: Keep full text for recent interactions (<30 min) - no action needed

    # Step 3: Remove redundant exchanges (similar Q&A within 10 min)
    redundant = session.interaction_buffer.find_redundant_exchanges()
    for redundant_interaction in redundant:
        tokens_freed += redundant_interaction.token_count
        session.interaction_buffer.remove(redundant_interaction)

    # Step 4: Archive to long-term storage (>4 hours old)
    archive_threshold = datetime.now() - timedelta(hours=4)
    archived_interactions = [
        i for i in session.interaction_buffer.history
        if i.timestamp < archive_threshold
    ]
    for interaction in archived_interactions:
        tokens_freed += interaction.token_count
        session.archive_interaction(interaction)
        session.interaction_buffer.remove(interaction)

    duration = (datetime.now() - start_time).total_seconds() * 1000

    return CleanupResult(
        component="interaction_buffer",
        tokens_freed=tokens_freed,
        duration_ms=duration,
        actions_taken=["summarize_old", "remove_redundant", "archive"]
    )
```

---

### Strategy 3: Budget Monitoring and Enforcement

Continuously monitor token usage and enforce budget constraints.

**Token Usage Tracking**:

```python
class TokenBudgetManager:
    """Manage token allocation with cleanup fallback."""

    def __init__(self, total_budget: int = 200000):
        self.total_budget = total_budget
        self.allocations = standard_allocations.copy()
        self.current_usage = {k: 0 for k in self.allocations}
        self.usage_history = []

    def allocate_tokens(self, component: str, requested: int) -> int:
        """Allocate tokens with automatic cleanup fallback.

        Args:
            component: Component requesting tokens
            requested: Number of tokens requested

        Returns:
            Number of tokens actually allocated

        Allocation Logic:
            1. Check if allocation available
            2. If yes: Grant and update usage
            3. If no: Trigger cleanup and retry
            4. If still no: Return partial allocation
        """
        max_allowed = int(self.total_budget * self.allocations[component])
        available = max_allowed - self.current_usage[component]

        # Full allocation available
        if requested <= available:
            self.current_usage[component] += requested
            self._log_allocation(component, requested, "full")
            return requested

        # Try cleanup to free tokens
        cleanup_result = self.cleanup_component(component)
        available_after_cleanup = max_allowed - self.current_usage[component]

        # Retry after cleanup
        if requested <= available_after_cleanup:
            self.current_usage[component] += requested
            self._log_allocation(component, requested, "after_cleanup")
            return requested

        # Partial allocation (best effort)
        allocated = available_after_cleanup
        if allocated > 0:
            self.current_usage[component] += allocated
            self._log_allocation(component, allocated, "partial")
        else:
            self._log_allocation(component, 0, "denied")

        return allocated

    def cleanup_component(self, component: str) -> CleanupResult:
        """Execute component-specific cleanup.

        Returns:
            Cleanup result with tokens freed
        """
        cleanup_functions = {
            "working_memory": cleanup_working_memory,
            "knowledge_base": cleanup_knowledge_base,
            "interaction_buffer": cleanup_interaction_buffer
        }

        cleanup_func = cleanup_functions.get(component)
        if cleanup_func:
            result = cleanup_func(self.session)
            self.current_usage[component] -= result.tokens_freed
            return result

        return CleanupResult(component=component, tokens_freed=0)

    def get_utilization_report(self) -> UtilizationReport:
        """Get current token utilization across all components.

        Returns:
            Detailed utilization report
        """
        report = {}

        for component in self.allocations:
            allocated = int(self.total_budget * self.allocations[component])
            used = self.current_usage[component]
            utilization_pct = (used / allocated * 100) if allocated > 0 else 0

            report[component] = {
                "allocated_tokens": allocated,
                "used_tokens": used,
                "available_tokens": allocated - used,
                "utilization_%": utilization_pct,
                "status": self._get_utilization_status(utilization_pct)
            }

        # Overall utilization
        total_used = sum(self.current_usage.values())
        report["overall"] = {
            "total_budget": self.total_budget,
            "total_used": total_used,
            "total_available": self.total_budget - total_used,
            "utilization_%": (total_used / self.total_budget * 100)
        }

        return UtilizationReport(report, timestamp=datetime.now())

    def _get_utilization_status(self, utilization_pct: float) -> str:
        """Determine utilization status based on percentage."""
        if utilization_pct >= 90:
            return "critical"
        elif utilization_pct >= 75:
            return "high"
        elif utilization_pct >= 50:
            return "moderate"
        else:
            return "low"

    def _log_allocation(self, component: str, tokens: int, status: str):
        """Log allocation event."""
        self.usage_history.append({
            "timestamp": datetime.now(),
            "component": component,
            "tokens": tokens,
            "status": status
        })
```

---

### Strategy 4: Predictive Budget Management

Anticipate token needs and proactively adjust allocations.

**Predictive Allocation**:

```python
class PredictiveBudgetManager(TokenBudgetManager):
    """Token budget manager with predictive allocation."""

    def predict_token_needs(self, upcoming_tasks: List[Task]) -> dict:
        """Predict token needs for upcoming tasks.

        Args:
            upcoming_tasks: List of tasks to be executed

        Returns:
            Predicted token requirements per component
        """
        predictions = {component: 0 for component in self.allocations}

        for task in upcoming_tasks:
            # Analyze task complexity
            complexity = self._analyze_task_complexity(task)

            # Estimate token needs based on complexity
            if complexity == "simple":
                predictions["working_memory"] += 10000
                predictions["agent_context"] += 5000
            elif complexity == "moderate":
                predictions["working_memory"] += 30000
                predictions["agent_context"] += 15000
                predictions["knowledge_base"] += 10000
            elif complexity == "complex":
                predictions["working_memory"] += 60000
                predictions["agent_context"] += 30000
                predictions["knowledge_base"] += 20000

        return predictions

    def preemptive_cleanup(self, predictions: dict):
        """Perform preemptive cleanup based on predictions."""
        for component, predicted_need in predictions.items():
            current_available = (
                int(self.total_budget * self.allocations[component]) -
                self.current_usage[component]
            )

            if predicted_need > current_available:
                # Cleanup needed
                tokens_needed = predicted_need - current_available
                self.cleanup_component(component)

    def _analyze_task_complexity(self, task: Task) -> str:
        """Analyze task complexity based on characteristics."""
        score = 0

        # Factor 1: Number of files involved
        score += len(task.files) * 10

        # Factor 2: Agent count
        score += len(task.agents) * 20

        # Factor 3: Estimated duration
        score += task.estimated_minutes * 2

        # Classify complexity
        if score < 50:
            return "simple"
        elif score < 150:
            return "moderate"
        else:
            return "complex"
```

---

## Performance Optimization

### Budget Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Allocation time | < 10ms | Per allocation request |
| Cleanup trigger time | < 500ms | Per component cleanup |
| Utilization accuracy | ± 5% | Predicted vs actual |
| Budget exhaustion events | < 3/session | Emergency cleanups |

### Optimization Targets

| Component | Target Utilization | Action if Exceeded |
|-----------|-------------------|-------------------|
| System context | 95-99% | Fixed (no action) |
| Working memory | 70-80% | Cleanup triggered |
| Knowledge base | 60-75% | Unload unused |
| Agent context | 70-85% | Archive completed |
| Interaction buffer | 60-75% | Summarize old |

**Overall Target**: 75-85% total utilization

---

## Best Practices

**DO**:
- ✅ Monitor component utilization continuously
- ✅ Adjust allocations based on usage patterns
- ✅ Trigger cleanup proactively (before exhaustion)
- ✅ Use predictive allocation for known tasks
- ✅ Archive aggressively to free tokens
- ✅ Log all allocation events for analysis

**DON'T**:
- ❌ Ignore utilization warnings (>90%)
- ❌ Use fixed allocations for all sessions
- ❌ Skip cleanup until budget exhausted
- ❌ Allocate tokens without tracking
- ❌ Delete data without archival
- ❌ Allow critical components to starve

---

**Version**: 3.0.0
**Last Updated**: 2025-11-23
**Status**: Production Ready
