# Circuit Breaker Pattern

Fault tolerance and graceful degradation using circuit breaker pattern.

---

## Circuit Breaker States

```python
from enum import Enum
from datetime import datetime, timedelta

class CircuitState(str, Enum):
    """Circuit breaker states."""
    CLOSED = "closed"          # Normal operation
    OPEN = "open"              # Failing, reject requests
    HALF_OPEN = "half_open"    # Testing recovery

class CircuitBreaker:
    def __init__(
        self,
        failure_threshold: int = 5,
        reset_timeout_seconds: int = 60,
        expected_exception: type = Exception
    ):
        self.state = CircuitState.CLOSED
        self.failure_count = 0
        self.failure_threshold = failure_threshold
        self.reset_timeout = reset_timeout_seconds
        self.last_failure_time = None
        self.expected_exception = expected_exception

    def call(self, func, *args, **kwargs):
        """Execute function with circuit breaker protection."""
        if self.state == CircuitState.OPEN:
            if self._should_attempt_reset():
                self.state = CircuitState.HALF_OPEN
                logger.info("Circuit breaker: attempting reset")
            else:
                raise CircuitBreakerOpenError(
                    "Circuit breaker is OPEN. Service unavailable."
                )

        try:
            result = func(*args, **kwargs)
            self._on_success()
            return result

        except self.expected_exception as e:
            self._on_failure()
            raise

    def _on_success(self):
        """Handle successful call."""
        self.failure_count = 0
        if self.state == CircuitState.HALF_OPEN:
            self.state = CircuitState.CLOSED
            logger.info("Circuit breaker: recovered, closing")

    def _on_failure(self):
        """Handle failed call."""
        self.failure_count += 1
        self.last_failure_time = datetime.now()

        if self.failure_count >= self.failure_threshold:
            self.state = CircuitState.OPEN
            logger.error(
                f"Circuit breaker: opened after {self.failure_count} failures"
            )

    def _should_attempt_reset(self) -> bool:
        """Check if enough time has passed to attempt reset."""
        if not self.last_failure_time:
            return False

        reset_time = self.last_failure_time + timedelta(
            seconds=self.reset_timeout
        )
        return datetime.now() >= reset_time

    def get_state(self) -> str:
        """Get current state."""
        return self.state

class CircuitBreakerOpenError(Exception):
    """Exception raised when circuit breaker is open."""
    pass
```

---

## Using Circuit Breaker

### Protecting External Service Calls

```python
# Create circuit breaker for external API
external_api_breaker = CircuitBreaker(
    failure_threshold=5,  # Open after 5 failures
    reset_timeout_seconds=60  # Try reset after 60 seconds
)

@server.tool()
def call_external_api(endpoint: str, params: dict) -> dict:
    """Call external API with circuit breaker."""
    try:
        result = external_api_breaker.call(
            make_api_request,
            endpoint,
            params
        )
        return {"status": "success", "data": result}

    except CircuitBreakerOpenError:
        logger.error("External API unavailable (circuit open)")
        raise ValueError(
            "External service currently unavailable. "
            "Please try again later."
        )

    except Exception as e:
        logger.error(f"API call failed: {str(e)}")
        raise ValueError(f"API call failed: {str(e)}")
```

### Database Connection Pool

```python
# Circuit breaker for database
db_breaker = CircuitBreaker(
    failure_threshold=10,
    reset_timeout_seconds=30
)

@server.tool()
def query_database(sql: str) -> list[dict]:
    """Query database with circuit breaker."""
    try:
        results = db_breaker.call(execute_query, sql)
        return results

    except CircuitBreakerOpenError:
        logger.error("Database unavailable (circuit open)")

        # Fallback: return cached data if available
        cached = get_cached_result(sql)
        if cached:
            logger.info("Using cached results")
            return cached

        raise ValueError(
            "Database temporarily unavailable. "
            "Try again in a moment."
        )
```

---

## Fallback Strategies

### Graceful Degradation

```python
@server.tool()
def get_user_with_fallback(user_id: str) -> dict:
    """Get user with fallback to cached/default data."""
    try:
        user = external_api_breaker.call(fetch_user_from_api, user_id)
        return user

    except CircuitBreakerOpenError:
        logger.warning("API circuit open, using fallback")

        # Try cache first
        cached = get_cached_user(user_id)
        if cached:
            return {**cached, "from_cache": True}

        # Return minimal default
        return {
            "id": user_id,
            "status": "unavailable",
            "message": "User data temporarily unavailable"
        }

    except Exception as e:
        logger.error(f"Fallback failed: {str(e)}")
        raise
```

### Multiple Fallbacks

```python
@server.tool()
def get_data_with_multiple_fallbacks(data_id: str) -> dict:
    """Data retrieval with multiple fallback options."""
    # Option 1: Try primary source
    try:
        return primary_breaker.call(fetch_from_primary, data_id)
    except CircuitBreakerOpenError:
        logger.warning("Primary source unavailable")

    # Option 2: Try secondary source
    try:
        return secondary_breaker.call(fetch_from_secondary, data_id)
    except CircuitBreakerOpenError:
        logger.warning("Secondary source unavailable")

    # Option 3: Try cache
    cached = get_cached_data(data_id)
    if cached:
        logger.info("Using cached data")
        return cached

    # Option 4: Return error
    raise ValueError(
        "All data sources unavailable. "
        "Please try again later."
    )
```

---

## Circuit Breaker Dashboard

```python
@server.resource("health://circuit-breakers")
def get_circuit_breaker_status() -> dict:
    """Get status of all circuit breakers."""
    return {
        "timestamp": datetime.now().isoformat(),
        "breakers": {
            "external_api": {
                "state": external_api_breaker.get_state(),
                "failure_count": external_api_breaker.failure_count,
                "threshold": external_api_breaker.failure_threshold
            },
            "database": {
                "state": db_breaker.get_state(),
                "failure_count": db_breaker.failure_count,
                "threshold": db_breaker.failure_threshold
            }
        }
    }
```

---

## Best Practices

✅ **Circuit Breaker**:
- Use for external service calls
- Set reasonable failure thresholds
- Monitor circuit state
- Implement fallback strategies
- Log state changes
- Test failure scenarios

✅ **Tuning**:
- Failure threshold: 5-10 failures
- Reset timeout: 30-60 seconds
- Match expected failure patterns
- Document thresholds

❌ **Avoid**:
- Too aggressive opening (low threshold)
- Too long reset timeout
- Missing fallback strategies
- Not monitoring circuit state
- Ignoring cascading failures

---

**Last Updated**: 2025-11-27
