# Advanced Patterns - Workflow Orchestration

**Multi-agent coordination and workflow design patterns**

---

## Workflow Architecture Patterns

### Sequential Workflow

```python
class SequentialWorkflow:
    """Execute agents in strict order, each depending on previous output."""
    
    async def execute(self):
        # Stage 1: Specification
        spec = await self.spec_agent.execute(self.input)
        
        # Stage 2: Implementation (depends on spec)
        impl = await self.impl_agent.execute(spec.output)
        
        # Stage 3: Testing (depends on impl)
        test = await self.test_agent.execute(impl.output)
        
        return test.output
```

### Parallel Workflow

```python
class ParallelWorkflow:
    """Execute independent agents in parallel."""
    
    async def execute(self):
        # All agents run simultaneously
        spec, impl, test = await asyncio.gather(
            self.spec_agent.execute(self.input),
            self.impl_agent.execute(self.input),
            self.test_agent.execute(self.input)
        )
        
        return self.merge_results([spec, impl, test])
```

### Conditional Workflow

```python
class ConditionalWorkflow:
    """Route execution based on conditions."""
    
    async def execute(self):
        spec = await self.spec_agent.execute(self.input)
        
        if spec.complexity > 5:
            # Complex: Use multiple agents
            return await self.complex_path(spec)
        else:
            # Simple: Use single agent
            return await self.simple_path(spec)
```

---

## Agent Coordination Patterns

### Request-Response Pattern
```python
async def request_response(agent, request):
    response = await agent.execute(request)
    return response.result
```

### Publish-Subscribe Pattern
```python
class PubSubCoordination:
    def __init__(self):
        self.subscribers = {}
    
    def subscribe(self, event_type, agent):
        self.subscribers.setdefault(event_type, []).append(agent)
    
    async def publish(self, event_type, data):
        for agent in self.subscribers[event_type]:
            await agent.on_event(event_type, data)
```

### State Machine Pattern
```python
class WorkflowStateMachine:
    states = ['pending', 'running', 'completed', 'failed']
    
    async def transition(self, new_state):
        if self._is_valid_transition(self.state, new_state):
            self.state = new_state
            await self.on_state_change(new_state)
```

---

## Error Handling & Resilience

### Retry Strategy

```python
async def execute_with_retry(agent, input_data, max_retries=3):
    for attempt in range(max_retries):
        try:
            return await agent.execute(input_data)
        except TransientError:
            if attempt < max_retries - 1:
                await asyncio.sleep(2 ** attempt)  # Exponential backoff
            else:
                raise
```

### Fallback Strategy

```python
async def execute_with_fallback(primary_agent, fallback_agent, input_data):
    try:
        return await primary_agent.execute(input_data)
    except PermanentError:
        return await fallback_agent.execute(input_data)
```

### Circuit Breaker Pattern

```python
class CircuitBreaker:
    def __init__(self, failure_threshold=5, timeout=60):
        self.failure_count = 0
        self.failure_threshold = failure_threshold
        self.timeout = timeout
        self.last_failure_time = None
    
    async def call(self, agent, input_data):
        if self.is_open():
            raise CircuitBreakerOpenError()
        
        try:
            result = await agent.execute(input_data)
            self.on_success()
            return result
        except Exception as e:
            self.on_failure()
            raise
```

---

## Context7 Integration

### Pattern Lookup

```python
async def research_workflow_pattern(workflow_type, complexity):
    """Use Context7 to find best practices for workflow type."""
    
    docs = await context7.get_library_docs(
        context7_library_id="/workflow-patterns/orchestration",
        topic=f"{workflow_type} workflow {complexity} complexity 2025",
        tokens=3000
    )
    
    return extract_best_practices(docs)
```

### Library Integration

```python
WORKFLOW_LIBRARIES = {
    'celery': '/celery/celery',
    'airflow': '/apache/airflow',
    'dask': '/dask/dask',
    'temporal': '/temporalio/temporal',
    'asyncio': '/python/asyncio'
}
```

---

## Monitoring & Observability

### Workflow Metrics

```python
class WorkflowMetrics:
    def __init__(self):
        self.execution_times = []
        self.agent_times = {}
        self.error_count = 0
        self.success_count = 0
    
    def record_execution(self, duration, agent_name, success):
        self.execution_times.append(duration)
        self.agent_times[agent_name] = duration
        
        if success:
            self.success_count += 1
        else:
            self.error_count += 1
    
    def get_stats(self):
        return {
            'avg_time': sum(self.execution_times) / len(self.execution_times),
            'success_rate': self.success_count / (self.success_count + self.error_count),
            'agent_breakdown': self.agent_times
        }
```

### Logging Strategy

```python
import logging

logger = logging.getLogger('workflow')

async def execute_with_logging(agent, input_data):
    logger.info(f"Starting {agent.name}")
    start_time = time.time()
    
    try:
        result = await agent.execute(input_data)
        duration = time.time() - start_time
        logger.info(f"Completed {agent.name} in {duration:.2f}s")
        return result
    except Exception as e:
        logger.error(f"Failed {agent.name}: {e}")
        raise
```

---

## Performance Optimization

### Caching

```python
class WorkflowCache:
    def __init__(self):
        self.agent_results = {}
    
    async def execute_cached(self, agent, input_hash):
        if input_hash in self.agent_results:
            return self.agent_results[input_hash]
        
        result = await agent.execute(...)
        self.agent_results[input_hash] = result
        return result
```

### Rate Limiting

```python
class RateLimiter:
    def __init__(self, max_requests=10, time_window=60):
        self.max_requests = max_requests
        self.time_window = time_window
        self.requests = deque()
    
    async def wait_if_needed(self):
        now = time.time()
        self.requests = deque(
            r for r in self.requests if r > now - self.time_window
        )
        
        if len(self.requests) >= self.max_requests:
            await asyncio.sleep(self.time_window)
        
        self.requests.append(now)
```

---

**Last Updated**: 2025-11-22
**Status**: Production Ready
