# Memory Profiling and Reference

Monitor memory usage and profile session performance for optimal resource utilization.

## Core Profiling Strategies

### Strategy 1: Memory Usage Profiling

Comprehensive memory profiling across all session layers and components.

**Profile Session Memory Usage**:

```python
def profile_memory_usage(session: SessionState) -> MemoryProfile:
    """Profile memory consumption across all layers.

    Args:
        session: Current session state

    Returns:
        Detailed memory profile with token counts

    Profiling Layers:
        - Working memory (active contexts)
        - Long-term memory (archived contexts)
        - Cache layer (cached items)
        - Agent states (execution contexts)
        - Conversation history (interaction buffer)
    """
    start_time = datetime.now()

    # Layer 1: Working Memory
    working_memory_tokens = measure_tokens(session.active_contexts)
    working_memory_breakdown = {
        "files": measure_tokens(session.working_memory.files),
        "modules": measure_tokens(session.working_memory.modules),
        "tasks": measure_tokens(session.working_memory.tasks),
        "debug_info": measure_tokens(session.working_memory.debug_info)
    }

    # Layer 2: Long-Term Memory
    long_term_memory_tokens = measure_tokens(session.archived_contexts)
    long_term_breakdown = {
        "archived_tasks": measure_tokens(session.long_term.archived_tasks),
        "consolidated": measure_tokens(session.long_term.consolidated),
        "compressed": measure_tokens(session.long_term.compressed)
    }

    # Layer 3: Cache
    cache_tokens = measure_tokens(session.cache.items)
    cache_breakdown = {
        "hot_cache": measure_tokens(session.cache.hot_items),
        "warm_cache": measure_tokens(session.cache.warm_items),
        "cold_cache": measure_tokens(session.cache.cold_items)
    }

    # Layer 4: Agent States
    agent_tokens = sum(measure_tokens(a.state) for a in session.agents)
    agent_breakdown = {
        agent.name: measure_tokens(agent.state) for agent in session.agents
    }

    # Layer 5: Conversation History
    conversation_tokens = measure_tokens(session.history)
    conversation_breakdown = {
        "recent": measure_tokens(session.history[-10:]),
        "medium": measure_tokens(session.history[-50:-10]),
        "old": measure_tokens(session.history[:-50])
    }

    # Calculate totals
    total_tokens = (
        working_memory_tokens +
        long_term_memory_tokens +
        cache_tokens +
        agent_tokens +
        conversation_tokens
    )

    utilization_rate = total_tokens / session.token_budget if session.token_budget > 0 else 0

    profile = MemoryProfile(
        working_memory=working_memory_tokens,
        working_memory_breakdown=working_memory_breakdown,
        long_term_memory=long_term_memory_tokens,
        long_term_breakdown=long_term_breakdown,
        cache=cache_tokens,
        cache_breakdown=cache_breakdown,
        agent_states=agent_tokens,
        agent_breakdown=agent_breakdown,
        conversation_history=conversation_tokens,
        conversation_breakdown=conversation_breakdown,
        total_tokens=total_tokens,
        utilization_rate=utilization_rate,
        token_budget=session.token_budget,
        profiling_duration_ms=(datetime.now() - start_time).total_seconds() * 1000,
        timestamp=datetime.now()
    )

    return profile

class MemoryProfile:
    """Container for memory profiling results."""

    def __init__(self, working_memory: int, working_memory_breakdown: dict,
                 long_term_memory: int, long_term_breakdown: dict,
                 cache: int, cache_breakdown: dict,
                 agent_states: int, agent_breakdown: dict,
                 conversation_history: int, conversation_breakdown: dict,
                 total_tokens: int, utilization_rate: float,
                 token_budget: int, profiling_duration_ms: float,
                 timestamp: datetime):
        self.working_memory = working_memory
        self.working_memory_breakdown = working_memory_breakdown
        self.long_term_memory = long_term_memory
        self.long_term_breakdown = long_term_breakdown
        self.cache = cache
        self.cache_breakdown = cache_breakdown
        self.agent_states = agent_states
        self.agent_breakdown = agent_breakdown
        self.conversation_history = conversation_history
        self.conversation_breakdown = conversation_breakdown
        self.total_tokens = total_tokens
        self.utilization_rate = utilization_rate
        self.token_budget = token_budget
        self.profiling_duration_ms = profiling_duration_ms
        self.timestamp = timestamp

    def get_summary(self) -> dict:
        """Get summary statistics."""
        return {
            "total_tokens": self.total_tokens,
            "utilization_%": self.utilization_rate * 100,
            "available_tokens": self.token_budget - self.total_tokens,
            "largest_component": self._get_largest_component(),
            "status": self._get_status()
        }

    def _get_largest_component(self) -> str:
        """Identify largest memory component."""
        components = {
            "working_memory": self.working_memory,
            "long_term_memory": self.long_term_memory,
            "cache": self.cache,
            "agent_states": self.agent_states,
            "conversation_history": self.conversation_history
        }
        return max(components, key=components.get)

    def _get_status(self) -> str:
        """Determine memory status."""
        if self.utilization_rate >= 0.95:
            return "critical"
        elif self.utilization_rate >= 0.85:
            return "high"
        elif self.utilization_rate >= 0.70:
            return "moderate"
        else:
            return "healthy"
```

---

### Strategy 2: Monitoring and Alerting

Continuous monitoring with automated alerting for critical conditions.

**Monitoring Points**:

```python
class MemoryMonitor:
    """Continuous memory monitoring with alerting."""

    def __init__(self, session: SessionState, alert_thresholds: dict):
        self.session = session
        self.thresholds = alert_thresholds
        self.metrics_history = []
        self.alert_history = []

    def monitor_continuously(self):
        """Monitor memory metrics continuously.

        Monitored Metrics:
            - Working Memory Utilization (0-100%, alert >80%)
            - Cache Hit Rate (0-100%, target 70-80%)
            - Token Budget Usage (0-100%, alert >90%)
            - Compression Effectiveness (ratio)
            - Cleanup Frequency (times per session)
        """
        # Collect current metrics
        metrics = {
            "working_memory_utilization": self.session.working_memory.utilization(),
            "cache_hit_rate": self.session.cache.hit_rate(),
            "token_budget_usage": self.session.token_usage / self.session.token_budget,
            "compression_ratio": self._calculate_compression_ratio(),
            "cleanup_count": self.session.cleanup_count,
            "timestamp": datetime.now()
        }

        # Check thresholds and generate alerts
        alerts = self._check_thresholds(metrics)

        # Log metrics
        self.metrics_history.append(metrics)
        if alerts:
            self.alert_history.extend(alerts)

        # Trigger actions based on alerts
        self._handle_alerts(alerts)

        return metrics, alerts

    def _check_thresholds(self, metrics: dict) -> List[Alert]:
        """Check metrics against thresholds."""
        alerts = []

        # Working Memory Utilization
        if metrics["working_memory_utilization"] > self.thresholds["working_memory_critical"]:
            alerts.append(Alert(
                severity="critical",
                metric="working_memory_utilization",
                value=metrics["working_memory_utilization"],
                threshold=self.thresholds["working_memory_critical"],
                message="Working memory critically high. Immediate cleanup required."
            ))
        elif metrics["working_memory_utilization"] > self.thresholds["working_memory_warning"]:
            alerts.append(Alert(
                severity="warning",
                metric="working_memory_utilization",
                value=metrics["working_memory_utilization"],
                threshold=self.thresholds["working_memory_warning"],
                message="Working memory approaching limit. Cleanup recommended."
            ))

        # Cache Hit Rate
        if metrics["cache_hit_rate"] < self.thresholds["cache_hit_rate_min"]:
            alerts.append(Alert(
                severity="warning",
                metric="cache_hit_rate",
                value=metrics["cache_hit_rate"],
                threshold=self.thresholds["cache_hit_rate_min"],
                message="Cache hit rate below target. Review caching strategy."
            ))

        # Token Budget Usage
        if metrics["token_budget_usage"] > self.thresholds["token_budget_critical"]:
            alerts.append(Alert(
                severity="critical",
                metric="token_budget_usage",
                value=metrics["token_budget_usage"],
                threshold=self.thresholds["token_budget_critical"],
                message="Token budget critically low. Aggressive cleanup required."
            ))
        elif metrics["token_budget_usage"] > self.thresholds["token_budget_warning"]:
            alerts.append(Alert(
                severity="warning",
                metric="token_budget_usage",
                value=metrics["token_budget_usage"],
                threshold=self.thresholds["token_budget_warning"],
                message="Token budget approaching limit. Cleanup recommended."
            ))

        return alerts

    def _handle_alerts(self, alerts: List[Alert]):
        """Handle alerts by triggering appropriate actions."""
        for alert in alerts:
            if alert.severity == "critical":
                # Trigger immediate cleanup
                if alert.metric == "working_memory_utilization":
                    cleanup_working_memory(self.session)
                elif alert.metric == "token_budget_usage":
                    aggressive_cleanup(self.session)

            elif alert.severity == "warning":
                # Schedule cleanup
                self.session.schedule_cleanup(alert.metric)

            # Log alert
            log_alert(alert)

    def _calculate_compression_ratio(self) -> float:
        """Calculate overall compression effectiveness."""
        if not self.session.compression_history:
            return 1.0

        total_original = sum(h["original_size"] for h in self.session.compression_history)
        total_compressed = sum(h["compressed_size"] for h in self.session.compression_history)

        return total_compressed / total_original if total_original > 0 else 1.0

class Alert:
    """Alert container."""

    def __init__(self, severity: str, metric: str, value: float,
                 threshold: float, message: str):
        self.severity = severity
        self.metric = metric
        self.value = value
        self.threshold = threshold
        self.message = message
        self.timestamp = datetime.now()

# Default alert thresholds
default_thresholds = {
    "working_memory_warning": 0.75,
    "working_memory_critical": 0.85,
    "cache_hit_rate_min": 0.65,
    "token_budget_warning": 0.85,
    "token_budget_critical": 0.95
}
```

---

### Strategy 3: Performance Diagnosis

Comprehensive diagnostic tools for identifying and resolving memory issues.

**Troubleshooting Checklist**:

```python
class MemoryDiagnostics:
    """Diagnostic tools for memory issues."""

    def __init__(self, session: SessionState):
        self.session = session

    def diagnose_high_working_memory(self) -> DiagnosticReport:
        """Diagnose high working memory usage.

        Symptoms: Working memory > 80%

        Diagnostic Steps:
            1. Identify large contexts (>10K tokens)
            2. Find inactive contexts (not accessed in 1hr)
            3. Check for memory leaks (growing contexts)
            4. Analyze context access patterns
        """
        issues = []

        # Step 1: Large contexts
        large_contexts = [
            ctx for ctx in self.session.working_memory.contexts
            if ctx.token_count > 10000
        ]
        if large_contexts:
            issues.append(Issue(
                type="large_contexts",
                severity="high",
                count=len(large_contexts),
                total_tokens=sum(ctx.token_count for ctx in large_contexts),
                recommendation="Compress or evict large inactive contexts"
            ))

        # Step 2: Inactive contexts
        one_hour_ago = datetime.now() - timedelta(hours=1)
        inactive_contexts = [
            ctx for ctx in self.session.working_memory.contexts
            if ctx.last_accessed < one_hour_ago
        ]
        if inactive_contexts:
            issues.append(Issue(
                type="inactive_contexts",
                severity="medium",
                count=len(inactive_contexts),
                total_tokens=sum(ctx.token_count for ctx in inactive_contexts),
                recommendation="Archive completed tasks and remove inactive contexts"
            ))

        # Step 3: Memory leaks (growing contexts)
        growing_contexts = self._identify_growing_contexts()
        if growing_contexts:
            issues.append(Issue(
                type="memory_leak",
                severity="critical",
                count=len(growing_contexts),
                contexts=growing_contexts,
                recommendation="Investigate growing contexts for memory leaks"
            ))

        # Step 4: Access patterns
        access_analysis = self._analyze_access_patterns()
        if access_analysis["low_hit_rate"]:
            issues.append(Issue(
                type="poor_access_patterns",
                severity="low",
                details=access_analysis,
                recommendation="Increase consolidation frequency"
            ))

        return DiagnosticReport(
            symptom="high_working_memory",
            issues=issues,
            overall_severity=self._determine_overall_severity(issues),
            timestamp=datetime.now()
        )

    def diagnose_low_cache_hit_rate(self) -> DiagnosticReport:
        """Diagnose low cache hit rate.

        Symptoms: Cache hit rate < 60%

        Diagnostic Steps:
            1. Analyze cache TTL settings
            2. Check cache eviction patterns
            3. Profile access patterns
            4. Review preload strategy
        """
        issues = []

        # Step 1: TTL analysis
        avg_ttl = self._calculate_avg_cache_ttl()
        if avg_ttl < timedelta(hours=12):
            issues.append(Issue(
                type="short_ttl",
                severity="medium",
                avg_ttl=avg_ttl,
                recommendation="Increase cache TTL for frequent items"
            ))

        # Step 2: Eviction patterns
        eviction_rate = self.session.cache.eviction_rate
        if eviction_rate > 0.15:  # >15% evictions
            issues.append(Issue(
                type="high_eviction_rate",
                severity="high",
                eviction_rate=eviction_rate,
                recommendation="Preload expected items on demand"
            ))

        # Step 3: Access patterns
        access_patterns = self._profile_cache_access_patterns()
        if access_patterns["random_access_ratio"] > 0.7:
            issues.append(Issue(
                type="random_access",
                severity="medium",
                ratio=access_patterns["random_access_ratio"],
                recommendation="Profile access patterns and adjust cache eviction policy"
            ))

        # Step 4: Preload strategy
        preload_effectiveness = self._calculate_preload_effectiveness()
        if preload_effectiveness < 0.5:
            issues.append(Issue(
                type="ineffective_preload",
                severity="low",
                effectiveness=preload_effectiveness,
                recommendation="Review and improve preload strategy"
            ))

        return DiagnosticReport(
            symptom="low_cache_hit_rate",
            issues=issues,
            overall_severity=self._determine_overall_severity(issues),
            timestamp=datetime.now()
        )

    def diagnose_token_budget_exhaustion(self) -> DiagnosticReport:
        """Diagnose token budget exhaustion.

        Symptoms: Total usage > 90% of budget

        Diagnostic Steps:
            1. Identify budget allocation imbalances
            2. Check for component over-utilization
            3. Review compression effectiveness
            4. Analyze cleanup frequency
        """
        issues = []

        # Step 1: Budget allocation
        budget_analysis = self._analyze_budget_allocation()
        for component, analysis in budget_analysis.items():
            if analysis["over_allocated"]:
                issues.append(Issue(
                    type="budget_imbalance",
                    severity="high",
                    component=component,
                    over_allocation=analysis["excess_pct"],
                    recommendation=f"Reduce {component} allocation or increase cleanup"
                ))

        # Step 2: Component over-utilization
        for component, utilization in self.session.component_utilization.items():
            if utilization > 0.95:
                issues.append(Issue(
                    type="component_exhaustion",
                    severity="critical",
                    component=component,
                    utilization=utilization,
                    recommendation=f"Trigger component cleanup for {component}"
                ))

        # Step 3: Compression effectiveness
        compression_ratio = self._calculate_compression_ratio()
        if compression_ratio > 0.7:
            issues.append(Issue(
                type="poor_compression",
                severity="medium",
                compression_ratio=compression_ratio,
                recommendation="Increase compression target ratio"
            ))

        # Step 4: Cleanup frequency
        cleanup_freq = self.session.cleanup_count / self.session.uptime_hours
        if cleanup_freq > 3:
            issues.append(Issue(
                type="frequent_cleanup",
                severity="high",
                frequency=cleanup_freq,
                recommendation="Reduce history retention to prevent frequent cleanups"
            ))

        return DiagnosticReport(
            symptom="token_budget_exhaustion",
            issues=issues,
            overall_severity=self._determine_overall_severity(issues),
            timestamp=datetime.now()
        )

    def _identify_growing_contexts(self) -> List[dict]:
        """Identify contexts with growing token counts."""
        growing = []

        for ctx in self.session.working_memory.contexts:
            growth_history = ctx.token_count_history
            if len(growth_history) >= 5:
                # Check for consistent growth
                recent_growth = growth_history[-5:]
                if all(recent_growth[i] < recent_growth[i+1] for i in range(len(recent_growth)-1)):
                    growing.append({
                        "context": ctx.name,
                        "growth_rate": recent_growth[-1] / recent_growth[0],
                        "current_size": ctx.token_count
                    })

        return growing

    def _determine_overall_severity(self, issues: List[Issue]) -> str:
        """Determine overall severity from issues."""
        if any(i.severity == "critical" for i in issues):
            return "critical"
        elif any(i.severity == "high" for i in issues):
            return "high"
        elif any(i.severity == "medium" for i in issues):
            return "medium"
        else:
            return "low"
```

---

## Best Practices

### Session Management Best Practices

**Lifecycle Approach**:

```python
# Best Practice Example: Complete Session Management

# 1. Initialize: Seed with essentials only
session = SessionState(session_id)
session.seed_context(essential_files)

# 2. Monitor: Track utilization continuously
monitor = MemoryMonitor(session, default_thresholds)
monitor.monitor_continuously()

# 3. Optimize: Compress and cache proactively
if session.working_memory.utilization() > 0.75:
    trigger_cleanup(session)

# 4. Cleanup: Trigger before hitting limits
if session.token_budget_utilization() > 0.85:
    aggressive_cleanup(session)

# 5. Archive: Save long-term after session
if session.is_inactive(hours=24):
    archive_session(session)
```

**Memory Strategy Checklist**:

- ✅ Keep working memory under 70% capacity
- ✅ Maintain 70-80% cache hit rate
- ✅ Compress automatically at 80% utilization
- ✅ Archive sessions after 24 hours inactive
- ✅ Review retention policies quarterly
- ✅ Monitor metrics continuously
- ✅ Trigger cleanup proactively
- ✅ Use predictive allocation for known tasks

**Token Budget Best Practices**:

- ✅ Monitor component allocations continuously
- ✅ Adjust based on usage patterns
- ✅ Trigger cleanup before exhaustion
- ✅ Document budget changes
- ✅ Archive metrics for analysis
- ✅ Use predictive cleanup
- ✅ Profile regularly

---

## Integration Examples

### With Agent Workflows

```python
# Memory-aware agent execution

# Before agent execution
memory_profile = profile_memory_usage(session)
if memory_profile.utilization_rate > 0.85:
    trigger_cleanup(session)

# Execute agent with memory monitoring
result = execute_agent_with_monitoring(session, agent)

# After completion
consolidate_memory(session)
update_cache(session)
```

### With Long-Running Tasks

```python
# Checkpoint-based memory management for long tasks

for checkpoint in task.checkpoints:
    # Save state
    snapshot = create_memory_snapshot(session)

    # Cleanup
    cleanup_expired_items(session)
    compress_memory(session)

    # Continue execution
    continue_task()

    # Profile after checkpoint
    profile = profile_memory_usage(session)
    log_checkpoint_profile(checkpoint, profile)
```

---

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Profiling overhead | < 100ms | Time to generate profile |
| Monitoring frequency | 1/minute | Continuous monitoring |
| Alert latency | < 1 second | Time from threshold to alert |
| Diagnostic time | < 5 seconds | Full diagnostic report |

---

**Version**: 3.0.0
**Last Updated**: 2025-11-23
**Status**: Production Ready

## References

- Claude Code Context Management
- Token Optimization Strategy
- Memory Architecture Patterns
- Cache Invalidation Best Practices
