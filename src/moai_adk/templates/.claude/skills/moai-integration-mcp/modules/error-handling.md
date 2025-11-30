# Error Handling and Resilience

## Overview
Comprehensive error handling framework for MCP integrations featuring retry logic, circuit breakers, fault tolerance, and graceful degradation strategies.

## Quick Implementation

### Robust Error Management

```python
class ErrorHandler:
 def __init__(self):
 self.retry_config = {
 'max_retries': 3,
 'backoff_factor': 2,
 'retryable_errors': ['timeout', 'rate_limit', 'temporary_failure']
 }

 async def handle_api_error(self, service: str, error: Exception, attempt: int = 1) -> dict:
 """Handle API errors with retry logic."""
 error_type = self.classify_error(error)

 if error_type in self.retry_config['retryable_errors'] and attempt <= self.retry_config['max_retries']:
 wait_time = self.retry_config['backoff_factor'] attempt
 await asyncio.sleep(wait_time)

 return {
 'should_retry': True,
 'wait_time': wait_time,
 'attempt': attempt + 1
 }

 return {
 'should_retry': False,
 'error_type': error_type,
 'message': str(error),
 'attempt': attempt
 }

 def classify_error(self, error: Exception) -> str:
 """Classify error type for retry determination."""
 if 'rate limit' in str(error).lower():
 return 'rate_limit'
 elif 'timeout' in str(error).lower():
 return 'timeout'
 elif 'temporary' in str(error).lower():
 return 'temporary_failure'
 else:
 return 'permanent_error'
```

### Circuit Breaker Pattern

```python
class CircuitBreaker:
 def __init__(self, failure_threshold=5, timeout=60, expected_exception=Exception):
 self.failure_threshold = failure_threshold
 self.timeout = timeout
 self.expected_exception = expected_exception
 self.failure_count = 0
 self.last_failure_time = None
 self.state = 'CLOSED' # CLOSED, OPEN, HALF_OPEN

 async def call(self, func, *args, kwargs):
 """Execute function with circuit breaker protection."""
 if self.state == 'OPEN':
 if self._should_attempt_reset():
 self.state = 'HALF_OPEN'
 else:
 raise Exception("Circuit breaker is OPEN")

 try:
 result = await func(*args, kwargs)
 self._on_success()
 return result

 except self.expected_exception as e:
 self._on_failure()
 raise e

 def _should_attempt_reset(self):
 """Check if circuit breaker should attempt to reset."""
 return time.time() - self.last_failure_time >= self.timeout

 def _on_success(self):
 """Handle successful operation."""
 self.failure_count = 0
 self.state = 'CLOSED'

 def _on_failure(self):
 """Handle failed operation."""
 self.failure_count += 1
 self.last_failure_time = time.time()

 if self.failure_count >= self.failure_threshold:
 self.state = 'OPEN'

# Usage example
class ResilientAPIClient:
 def __init__(self, api_client):
 self.api_client = api_client
 self.circuit_breaker = CircuitBreaker(
 failure_threshold=5,
 timeout=60
 )

 async def make_request(self, endpoint, kwargs):
 """Make API request with circuit breaker protection."""
 return await self.circuit_breaker.call(
 self.api_client.request,
 endpoint,
 kwargs
 )
```

### Retry with Exponential Backoff

```python
class RetryManager:
 def __init__(self, max_retries=3, base_delay=1, max_delay=60):
 self.max_retries = max_retries
 self.base_delay = base_delay
 self.max_delay = max_delay

 async def retry_with_backoff(self, func, *args, kwargs):
 """Execute function with exponential backoff retry."""
 last_exception = None

 for attempt in range(self.max_retries + 1):
 try:
 return await func(*args, kwargs)

 except Exception as e:
 last_exception = e

 if attempt < self.max_retries and self._should_retry(e):
 delay = self._calculate_delay(attempt)
 await asyncio.sleep(delay)
 else:
 break

 raise last_exception

 def _should_retry(self, exception: Exception) -> bool:
 """Determine if exception is retryable."""
 retryable_errors = [
 'timeout',
 'rate limit',
 'temporary',
 'connection',
 'network'
 ]

 error_str = str(exception).lower()
 return any(error in error_str for error in retryable_errors)

 def _calculate_delay(self, attempt: int) -> float:
 """Calculate exponential backoff delay."""
 delay = self.base_delay * (2 attempt)
 jitter = random.uniform(0, 0.1) * delay # Add jitter
 return min(delay + jitter, self.max_delay)

# Usage example
class RetryableServiceClient:
 def __init__(self):
 self.retry_manager = RetryManager(
 max_retries=3,
 base_delay=1,
 max_delay=30
 )

 async def get_resource(self, resource_id: str):
 """Get resource with retry logic."""
 return await self.retry_manager.retry_with_backoff(
 self._fetch_resource,
 resource_id
 )

 async def _fetch_resource(self, resource_id: str):
 """Actual resource fetching logic."""
 # Implementation that might fail
 pass
```

## Key Features

### 1. Error Classification
- Automatic error type detection
- Retryable vs non-retryable errors
- Service-specific error handling
- Custom error categories

### 2. Retry Strategies
- Exponential backoff with jitter
- Linear backoff
- Custom retry policies
- Maximum retry limits

### 3. Circuit Breaker Pattern
- Failure threshold management
- Automatic state transitions
- Half-open state testing
- Service isolation

### 4. Fault Tolerance
- Graceful degradation
- Fallback mechanisms
- Bulkhead isolation
- Timeout management

## Resilience Patterns

### 1. Bulkhead Pattern
```python
class Bulkhead:
 def __init__(self, max_concurrent_calls=10):
 self.semaphore = asyncio.Semaphore(max_concurrent_calls)
 self.active_calls = 0

 async def execute(self, func, *args, kwargs):
 """Execute function with bulkhead protection."""
 async with self.semaphore:
 self.active_calls += 1
 try:
 return await func(*args, kwargs)
 finally:
 self.active_calls -= 1
```

### 2. Timeout Management
```python
class TimeoutManager:
 def __init__(self, default_timeout=30):
 self.default_timeout = default_timeout

 async def execute_with_timeout(self, func, timeout=None, *args, kwargs):
 """Execute function with timeout protection."""
 timeout = timeout or self.default_timeout
 try:
 return await asyncio.wait_for(
 func(*args, kwargs),
 timeout=timeout
 )
 except asyncio.TimeoutError:
 raise Exception(f"Operation timed out after {timeout} seconds")
```

### 3. Fallback Mechanisms
```python
class FallbackManager:
 def __init__(self):
 self.fallback_strategies = {}

 def register_fallback(self, operation: str, fallback_func):
 """Register fallback strategy for operation."""
 self.fallback_strategies[operation] = fallback_func

 async def execute_with_fallback(self, operation: str, primary_func, *args, kwargs):
 """Execute operation with fallback support."""
 try:
 return await primary_func(*args, kwargs)
 except Exception as e:
 if operation in self.fallback_strategies:
 logging.warning(f"Primary operation failed, using fallback: {e}")
 return await self.fallback_strategies[operation](*args, kwargs)
 else:
 raise e
```

## Monitoring and Observability

### Error Metrics
```python
class ErrorMetrics:
 def __init__(self):
 self.error_counts = {}
 self.retry_counts = {}
 self.circuit_breaker_events = {}

 def record_error(self, service: str, error_type: str):
 """Record error occurrence."""
 key = f"{service}:{error_type}"
 self.error_counts[key] = self.error_counts.get(key, 0) + 1

 def record_retry(self, service: str):
 """Record retry attempt."""
 self.retry_counts[service] = self.retry_counts.get(service, 0) + 1

 def record_circuit_breaker_event(self, service: str, event: str):
 """Record circuit breaker state change."""
 key = f"{service}:{event}"
 self.circuit_breaker_events[key] = self.circuit_breaker_events.get(key, 0) + 1

 def get_metrics_summary(self) -> dict:
 """Get summary of all metrics."""
 return {
 'error_counts': self.error_counts,
 'retry_counts': self.retry_counts,
 'circuit_breaker_events': self.circuit_breaker_events
 }
```

## Best Practices

### 1. Error Handling Strategy
- Implement comprehensive error classification
- Use appropriate retry policies
- Set realistic timeout values
- Monitor and adjust thresholds

### 2. Circuit Breaker Configuration
- Set appropriate failure thresholds
- Configure reasonable timeout periods
- Implement proper state transitions
- Monitor circuit breaker states

### 3. Retry Policy Design
- Use exponential backoff with jitter
- Implement retry limits
- Handle different error types appropriately
- Consider service-specific requirements

### 4. Monitoring and Alerting
- Track error rates and patterns
- Monitor circuit breaker states
- Set up appropriate alerts
- Regular review and adjustment

## Integration Points
- External API clients
- Database connections
- Message queue systems
- File system operations
- Network communications
