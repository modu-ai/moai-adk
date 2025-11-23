# Debugging Patterns

Enterprise debugging patterns for systematic troubleshooting across distributed systems and modern architectures.

## Five Core Debugging Patterns

### Pattern 1: Binary Search Debugging

**Objective**: Narrow down error location using divide-and-conquer.

**When to Use**:
- Large codebase with unclear error location
- Regression testing failures
- Performance degradation without obvious cause

**Implementation**:
```python
class BinarySearchDebugger:
    """Binary search debugging implementation."""

    def locate_error(self, code_path: str, test_function: callable) -> dict:
        """
        Locate error using binary search approach.

        Args:
            code_path: Path to code file
            test_function: Function that returns True if error occurs

        Returns:
            Dictionary with error location details
        """
        with open(code_path, 'r') as f:
            lines = f.readlines()

        low = 0
        high = len(lines) - 1
        error_line = None

        while low <= high:
            mid = (low + high) // 2

            # Test with lines up to mid
            test_code = ''.join(lines[:mid+1])
            error_occurs = test_function(test_code)

            if error_occurs:
                error_line = mid
                high = mid - 1
            else:
                low = mid + 1

        return {
            'error_line': error_line,
            'error_context': self._get_context(lines, error_line),
            'confidence': self._calculate_confidence(error_line)
        }
```

**Example Use Case**:
```python
def test_code_segment(code: str) -> bool:
    """Test if code segment triggers error."""
    try:
        exec(code)
        return False
    except ValueError:
        return True

debugger = BinarySearchDebugger()
result = debugger.locate_error('app.py', test_code_segment)
print(f"Error at line {result['error_line']}")
```

### Pattern 2: Rubber Duck Debugging

**Objective**: Explain code step-by-step to discover logical errors.

**When to Use**:
- Logic errors without exceptions
- Complex algorithmic issues
- Design flaws

**Implementation**:
```python
class RubberDuckDebugger:
    """Automated rubber duck debugging with AI analysis."""

    def explain_code(self, code: str, context: dict) -> dict:
        """
        Analyze code by explaining each step.

        Args:
            code: Code to analyze
            context: Execution context

        Returns:
            Analysis with potential issues
        """
        steps = self._parse_code_steps(code)
        explanations = []
        issues = []

        for i, step in enumerate(steps):
            explanation = self._explain_step(step, context)
            explanations.append(explanation)

            # Check for logical inconsistencies
            if self._detect_inconsistency(step, explanation, context):
                issues.append({
                    'step': i,
                    'issue': explanation['detected_issue'],
                    'suggestion': explanation['suggestion']
                })

            # Update context
            context = self._update_context(context, step)

        return {
            'explanations': explanations,
            'issues': issues,
            'recommendations': self._generate_recommendations(issues)
        }

    def _detect_inconsistency(self, step: dict, explanation: dict, context: dict) -> bool:
        """Detect logical inconsistencies in step execution."""
        # Check for type mismatches
        if explanation.get('expected_type') != explanation.get('actual_type'):
            return True

        # Check for boundary conditions
        if self._is_boundary_violation(step, context):
            return True

        # Check for unreachable code
        if self._is_unreachable(step, context):
            return True

        return False
```

### Pattern 3: Wolf Fence Debugging

**Objective**: Systematically eliminate possibilities to isolate the problem.

**When to Use**:
- Intermittent bugs
- Environment-specific issues
- Complex integration problems

**Implementation**:
```python
class WolfFenceDebugger:
    """Wolf fence debugging for systematic elimination."""

    def isolate_component(self, components: list, test_function: callable) -> dict:
        """
        Isolate problematic component using wolf fence approach.

        Args:
            components: List of system components
            test_function: Function to test component health

        Returns:
            Problematic component details
        """
        results = {}

        # Test all components
        for component in components:
            result = test_function(component)
            results[component.name] = {
                'healthy': result['success'],
                'metrics': result['metrics'],
                'errors': result.get('errors', [])
            }

        # Analyze results
        unhealthy = [c for c, r in results.items() if not r['healthy']]

        if len(unhealthy) == 0:
            return {'status': 'all_healthy', 'suspect': None}

        if len(unhealthy) == 1:
            return {
                'status': 'component_identified',
                'suspect': unhealthy[0],
                'details': results[unhealthy[0]]
            }

        # Multiple failures: check dependencies
        dependency_graph = self._build_dependency_graph(components)
        root_cause = self._find_root_cause(unhealthy, dependency_graph)

        return {
            'status': 'cascading_failure',
            'root_cause': root_cause,
            'affected_components': unhealthy
        }
```

### Pattern 4: Logging-Driven Debugging

**Objective**: Use structured logging to trace execution flow and state.

**When to Use**:
- Production debugging (no debugger access)
- Distributed systems
- Asynchronous code

**Implementation**:
```python
import structlog
from functools import wraps

class LoggingDebugger:
    """Structured logging for debugging."""

    def __init__(self):
        self.logger = structlog.get_logger()

    def trace_function(self, func):
        """Decorator to trace function execution."""
        @wraps(func)
        def wrapper(*args, **kwargs):
            # Log entry
            self.logger.info(
                "function_entry",
                function=func.__name__,
                args=args,
                kwargs=kwargs
            )

            try:
                result = func(*args, **kwargs)

                # Log success
                self.logger.info(
                    "function_success",
                    function=func.__name__,
                    result=result
                )

                return result

            except Exception as e:
                # Log error with context
                self.logger.error(
                    "function_error",
                    function=func.__name__,
                    error=str(e),
                    error_type=type(e).__name__,
                    args=args,
                    kwargs=kwargs
                )
                raise

        return wrapper

    def trace_state(self, obj: object, attributes: list):
        """Trace object state changes."""
        for attr in attributes:
            value = getattr(obj, attr)
            self.logger.debug(
                "state_trace",
                object=obj.__class__.__name__,
                attribute=attr,
                value=value
            )
```

**Example Use**:
```python
debugger = LoggingDebugger()

@debugger.trace_function
def process_payment(user_id: str, amount: float):
    # Function logic
    pass

# Automatic logging of entry, success, and errors
```

### Pattern 5: Time-Travel Debugging

**Objective**: Replay execution history to understand state evolution.

**When to Use**:
- Race conditions
- Non-deterministic bugs
- Complex state management

**Implementation**:
```python
class TimeTravelDebugger:
    """Record and replay execution history."""

    def __init__(self):
        self.history = []
        self.current_index = 0

    def record_state(self, state: dict, operation: str):
        """Record state snapshot."""
        self.history.append({
            'timestamp': datetime.utcnow().isoformat(),
            'index': len(self.history),
            'state': state.copy(),
            'operation': operation
        })

    def go_to_state(self, index: int) -> dict:
        """Travel to specific state in history."""
        if 0 <= index < len(self.history):
            self.current_index = index
            return self.history[index]['state']
        raise IndexError("State index out of range")

    def replay(self, start_index: int = 0, end_index: int = None) -> list:
        """Replay execution from start to end."""
        end = end_index or len(self.history)
        replay_results = []

        for i in range(start_index, end):
            state = self.history[i]
            replay_results.append({
                'index': i,
                'operation': state['operation'],
                'state_before': self.history[i-1]['state'] if i > 0 else None,
                'state_after': state['state'],
                'changes': self._compute_changes(i)
            })

        return replay_results

    def _compute_changes(self, index: int) -> dict:
        """Compute state changes between snapshots."""
        if index == 0:
            return {}

        current = self.history[index]['state']
        previous = self.history[index-1]['state']

        changes = {}
        for key in current:
            if key not in previous:
                changes[key] = {'type': 'added', 'value': current[key]}
            elif current[key] != previous[key]:
                changes[key] = {
                    'type': 'modified',
                    'old': previous[key],
                    'new': current[key]
                }

        return changes
```

## Advanced Debugging Techniques

### Conditional Breakpoints

```python
class ConditionalBreakpoint:
    """Advanced conditional breakpoint system."""

    def __init__(self):
        self.breakpoints = {}

    def add(self, location: str, condition: callable):
        """Add conditional breakpoint."""
        self.breakpoints[location] = {
            'condition': condition,
            'hit_count': 0,
            'active': True
        }

    def check(self, location: str, context: dict) -> bool:
        """Check if breakpoint should trigger."""
        if location not in self.breakpoints:
            return False

        bp = self.breakpoints[location]
        if not bp['active']:
            return False

        should_break = bp['condition'](context)
        if should_break:
            bp['hit_count'] += 1

        return should_break
```

**Example**:
```python
bp = ConditionalBreakpoint()

# Break only when user_id is 'admin' and request fails
bp.add(
    location='api.process_request:42',
    condition=lambda ctx: ctx['user_id'] == 'admin' and ctx['status'] == 'error'
)
```

### Performance Profiling Integration

```python
class ProfilingDebugger:
    """Integrate profiling with debugging."""

    def profile_function(self, func: callable) -> dict:
        """Profile function execution."""
        profiler = cProfile.Profile()

        profiler.enable()
        result = func()
        profiler.disable()

        stats = pstats.Stats(profiler)
        stats.sort_stats('cumulative')

        return {
            'result': result,
            'profile': self._format_stats(stats),
            'hotspots': self._identify_hotspots(stats)
        }

    def _identify_hotspots(self, stats: pstats.Stats) -> list:
        """Identify performance hotspots."""
        hotspots = []

        for func, (cc, nc, tt, ct, callers) in stats.stats.items():
            if tt > 0.1:  # Functions taking >100ms
                hotspots.append({
                    'function': func,
                    'time': tt,
                    'calls': nc,
                    'avg_time': tt / nc if nc > 0 else 0
                })

        return sorted(hotspots, key=lambda h: h['time'], reverse=True)
```

## Debugging Best Practices

### DO's:
- ✅ Use systematic patterns (binary search, wolf fence)
- ✅ Implement structured logging
- ✅ Record execution history for complex bugs
- ✅ Profile performance before optimization
- ✅ Use conditional breakpoints for rare conditions
- ✅ Document debugging process

### DON'Ts:
- ❌ Debug without reproducing the issue first
- ❌ Change multiple variables simultaneously
- ❌ Skip logging in production code
- ❌ Ignore performance profiling data
- ❌ Debug in production without safeguards

---

**End of Debugging Patterns** | Status: Production Ready
