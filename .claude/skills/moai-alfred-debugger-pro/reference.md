# Debugging Reference Documentation

_Last updated: 2025-10-22_

## Debugging Tools by Language (2025)

### Python
- **Debugger**: pdb (built-in), ipdb (enhanced)
- **Profiler**: cProfile, line_profiler, memory_profiler
- **Error Tracking**: Sentry, Rollbar

```bash
# Interactive debugging
python -m pdb script.py

# Post-mortem debugging
python -c "import pdb; pdb.pm()"

# Profiling
python -m cProfile -s cumtime script.py
```

### JavaScript/TypeScript
- **Debugger**: Node inspect, Chrome DevTools
- **Error Tracking**: Sentry, Bugsnag
- **Profiler**: Chrome DevTools, clinic.js

```bash
# Node.js debugging
node inspect script.js
node --inspect-brk script.js

# Chrome DevTools
# Navigate to chrome://inspect
```

### Go
- **Debugger**: Delve (dlv)
- **Profiler**: pprof
- **Error Tracking**: Sentry

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug
dlv test
```

---

## Stack Trace Analysis Patterns

### Reading Stack Traces

**Key elements**:
1. Exception type (what failed)
2. Error message (why it failed)
3. Call stack (where it failed)
4. Line numbers (exact location)

**Example interpretation**:
```
IndexError: list index out of range
  at process_items (main.py:42)    ← Where
  at handle_request (api.py:15)    ← Caller
  at app.run (app.py:8)            ← Entry point
```

### Common Error Patterns

| Pattern | Cause | Solution |
|---------|-------|----------|
| NullPointerException | Accessing null object | Add null checks, use Optional |
| IndexError | Array out of bounds | Validate index range |
| ConnectionTimeout | Network issues | Add retry logic, increase timeout |
| MemoryError | Insufficient memory | Optimize data structures, stream data |

---

## Error Tracking Tools (2025)

### Sentry
```python
import sentry_sdk

sentry_sdk.init(
    dsn="YOUR_DSN",
    traces_sample_rate=1.0,
    environment="production"
)

# Automatic error capture
try:
    risky_operation()
except Exception as e:
    sentry_sdk.capture_exception(e)
```

### Features
- Stack trace grouping
- Release tracking
- Performance monitoring
- AI-powered fix suggestions

---

## Debugging Workflow

### 1. Reproduce
- Isolate minimal reproduction case
- Document steps to reproduce
- Capture environment details

### 2. Analyze
- Read stack trace top-to-bottom
- Identify error type and message
- Trace execution flow

### 3. Hypothesize
- Form theories about root cause
- Prioritize most likely causes
- Design tests for each hypothesis

### 4. Test
- Add logging/breakpoints
- Run with debugger
- Verify hypothesis

### 5. Fix
- Implement solution
- Add regression test
- Document in commit message

---

## Performance Profiling

### CPU Profiling
```python
import cProfile
import pstats

# Profile code
cProfile.run('my_function()', 'output.prof')

# Analyze results
stats = pstats.Stats('output.prof')
stats.sort_stats('cumulative')
stats.print_stats(10)
```

### Memory Profiling
```python
from memory_profiler import profile

@profile
def my_function():
    large_list = [0] * 10000000
    return sum(large_list)
```

---

## References

- [Python Debugging](https://docs.python.org/3/library/pdb.html)
- [Delve Debugger](https://github.com/go-delve/delve)
- [Sentry Error Tracking](https://sentry.io/welcome/)
- [Chrome DevTools](https://developer.chrome.com/docs/devtools/)

---

_For practical examples, see examples.md_
