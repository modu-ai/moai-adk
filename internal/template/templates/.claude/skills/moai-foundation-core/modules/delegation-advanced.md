# Delegation Advanced Patterns

Purpose: Error handling, recovery strategies, and hybrid delegation patterns for complex workflows.

Version: 1.0.0
Last Updated: 2026-01-06
Parent: [delegation-patterns.md](delegation-patterns.md)

---

## Error Handling and Recovery

### Resilient Delegation Pattern

```python
from typing import Optional
import asyncio

class ResilientDelegation:
    """Handle delegation failures with retry and fallback."""

    def __init__(self):
        self.max_retries = 3
        self.retry_delay = 5  # seconds

    async def delegate_with_retry(
        self,
        agent_type: str,
        prompt: str,
        context: dict,
        fallback_agent: Optional[str] = None
    ):
        """Delegate with automatic retry and fallback."""

        for attempt in range(self.max_retries):
            try:
                result = await Task(
                    subagent_type=agent_type,
                    prompt=prompt,
                    context=context
                )
                return result

            except AgentExecutionError as e:
                if attempt < self.max_retries - 1:
                    delay = self.retry_delay * (2 ** attempt)
                    await asyncio.sleep(delay)
                    continue

                if fallback_agent:
                    return await self._fallback_delegation(
                        fallback_agent, prompt, context, original_error=e
                    )

                raise

            except ContextTooLargeError as e:
                compressed_context = self._compress_context(context)
                return await Task(
                    subagent_type=agent_type,
                    prompt=prompt,
                    context=compressed_context
                )

    async def _fallback_delegation(
        self,
        fallback_agent: str,
        prompt: str,
        context: dict,
        original_error: Exception
    ):
        """Execute fallback delegation."""
        fallback_context = {
            **context,
            "original_error": str(original_error),
            "fallback_mode": True
        }

        return await Task(
            subagent_type=fallback_agent,
            prompt=f"[FALLBACK] {prompt}",
            context=fallback_context
        )
```

Usage:
```python
delegator = ResilientDelegation()

result = await delegator.delegate_with_retry(
    agent_type="code-backend",
    prompt="Implement complex feature",
    context=large_context,
    fallback_agent="support-debug"
)
```

---

## Hybrid Delegation Patterns

### Sequential + Parallel Combination

```python
async def hybrid_workflow(spec_id: str):
    """Combine sequential and parallel patterns."""

    # Phase 1: Sequential (SPEC → Design)
    spec = await Task(
        subagent_type="workflow-spec",
        prompt=f"Generate SPEC {spec_id}"
    )

    design = await Task(
        subagent_type="api-designer",
        prompt="Design API",
        context={"spec_id": spec.id}
    )

    execute_clear()

    # Phase 2: Parallel (Implementation)
    impl_results = await Promise.all([
        Task(
            subagent_type="code-backend",
            prompt="Backend",
            context={"spec_id": spec.id, "api": design}
        ),
        Task(
            subagent_type="code-frontend",
            prompt="Frontend",
            context={"spec_id": spec.id, "api": design}
        ),
        Task(
            subagent_type="data-database",
            prompt="Database",
            context={"spec_id": spec.id, "api": design}
        )
    ])

    backend, frontend, database = impl_results

    # Phase 3: Sequential (Testing → QA)
    tests = await Task(
        subagent_type="core-quality",
        prompt="Integration tests",
        context={
            "spec_id": spec.id,
            "backend": backend.summary,
            "frontend": frontend.summary,
            "database": database.summary
        }
    )

    qa = await Task(
        subagent_type="core-quality",
        prompt="Quality validation",
        context={
            "spec_id": spec.id,
            "tests": tests.results,
            "coverage": tests.coverage
        }
    )

    return {
        "spec": spec,
        "design": design,
        "backend": backend,
        "frontend": frontend,
        "database": database,
        "tests": tests,
        "qa": qa
    }
```

---

### Conditional + Parallel Combination

```python
async def conditional_parallel_workflow(requests: list):
    """Route multiple requests in parallel based on analysis."""

    # Phase 1: Parallel analysis
    analyses = await Promise.all([
        Task(
            subagent_type="plan",
            prompt=f"Analyze request",
            context={"request": req}
        )
        for req in requests
    ])

    # Phase 2: Conditional routing (grouped by type)
    security_tasks = []
    feature_tasks = []
    bug_tasks = []

    for analysis in analyses:
        if analysis.category == "security":
            security_tasks.append(
                Task(
                    subagent_type="security-expert",
                    prompt="Handle security issue",
                    context={"analysis": analysis}
                )
            )
        elif analysis.category == "feature":
            feature_tasks.append(
                Task(
                    subagent_type="code-backend",
                    prompt="Implement feature",
                    context={"analysis": analysis}
                )
            )
        elif analysis.category == "bug":
            bug_tasks.append(
                Task(
                    subagent_type="support-debug",
                    prompt="Debug issue",
                    context={"analysis": analysis}
                )
            )

    # Phase 3: Parallel execution by category
    results = await Promise.all([
        *security_tasks,
        *feature_tasks,
        *bug_tasks
    ])

    return results
```

---

## Context Compression Strategies

```python
class ContextCompressor:
    """Compress context when exceeding limits."""

    def compress_for_agent(self, context: dict, agent_type: str) -> dict:
        """Compress context based on agent requirements."""

        # Remove non-essential fields
        essential_fields = self._get_essential_fields(agent_type)
        compressed = {k: v for k, v in context.items() if k in essential_fields}

        # Compress large text fields
        for key, value in compressed.items():
            if isinstance(value, str) and len(value) > 5000:
                compressed[key] = self._summarize_text(value)

            elif isinstance(value, list) and len(value) > 100:
                compressed[key] = value[:50] + ["... (truncated)"] + value[-10:]

        return compressed

    def _summarize_text(self, text: str, max_length: int = 1000) -> str:
        """Summarize long text to key points."""
        paragraphs = text.split("\n\n")
        if len(paragraphs) > 2:
            summary = paragraphs[0] + "\n\n...\n\n" + paragraphs[-1]
            return summary[:max_length] if len(summary) > max_length else summary
        return text[:max_length]
```

---

## Works Well With

- [delegation-patterns.md](delegation-patterns.md) - Overview
- [delegation-implementation.md](delegation-implementation.md) - Basic patterns

---

Version: 1.0.0
Last Updated: 2026-01-06
