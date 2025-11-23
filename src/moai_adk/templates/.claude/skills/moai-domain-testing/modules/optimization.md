# Test Performance & Optimization - Enterprise Test Speed & Reliability

**Version**: 4.0.0 (2025-11-22)
**Status**: Production Ready

---

## Test Execution Optimization

### Parallel Test Execution

```python
# pytest.ini or pyproject.toml
[tool:pytest]
addopts = -n auto --dist=loadscope
# Uses pytest-xdist for parallelization

# Command: pytest tests/ -n 4 (4 workers)
```

### Smart Test Ordering

```python
import pytest

class TestOptimizer:
    """Order tests for maximum efficiency."""

    @pytest.fixture(scope='session')
    def setup_once(self):
        """Setup run once per session."""
        setup_database()
        yield
        teardown_database()

    def test_1_create_user(self, setup_once):
        """Fast test first (50ms)."""
        assert create_user('john') is not None

    def test_2_query_user(self, setup_once):
        """Depends on test_1, run after (30ms)."""
        user = query_user('john')
        assert user is not None

    # Slow tests last
    @pytest.mark.slow
    def test_integration_flow(self):
        """Slow integration test (2 seconds)."""
        pass

# Run fast tests first: pytest -m "not slow"
```

### Test Result Caching

```python
from functools import lru_cache
from datetime import datetime, timedelta

class TestCachingStrategy:
    """Cache expensive test setup/teardown."""

    def __init__(self):
        self.cache = {}
        self.cache_ttl = timedelta(hours=1)

    def cache_fixture(self, key: str, value, ttl=None):
        """Cache fixture with TTL."""
        self.cache[key] = {
            'value': value,
            'timestamp': datetime.now(),
            'ttl': ttl or self.cache_ttl
        }

    def get_cached(self, key: str):
        """Retrieve cached fixture if not expired."""
        if key not in self.cache:
            return None

        cached = self.cache[key]
        if datetime.now() - cached['timestamp'] > cached['ttl']:
            del self.cache[key]
            return None

        return cached['value']

    def clear_expired(self):
        """Remove all expired cache entries."""
        expired_keys = [
            k for k, v in self.cache.items()
            if datetime.now() - v['timestamp'] > v['ttl']
        ]
        for k in expired_keys:
            del self.cache[k]
```

---

## Fixture Optimization

### Shared Database Fixtures

```python
import pytest
from sqlalchemy import create_engine

class SharedDatabaseFixture:
    """Reuse database across tests for speed."""

    @pytest.fixture(scope='session')
    def db_engine():
        """Create once per session."""
        engine = create_engine('postgresql://localhost/test_db')
        yield engine
        engine.dispose()

    @pytest.fixture(scope='function')
    def db_session(db_engine):
        """Transaction per test (rollback after)."""
        connection = db_engine.connect()
        transaction = connection.begin()
        session = Session(bind=connection)

        yield session

        session.close()
        transaction.rollback()
        connection.close()

    @pytest.fixture(scope='module')
    def sample_data(db_session):
        """Shared sample data across module."""
        users = [
            User(name='john', email='john@example.com'),
            User(name='jane', email='jane@example.com'),
        ]
        db_session.add_all(users)
        db_session.commit()
        return users
```

### Lazy Fixture Loading

```python
import pytest

class LazyFixtureOptimizer:
    """Load fixtures only when needed."""

    expensive_resource = None

    @pytest.fixture
    def expensive_fixture(self):
        """Load expensive resource on demand."""
        if self.expensive_resource is None:
            self.expensive_resource = self._initialize_resource()
        return self.expensive_resource

    def _initialize_resource(self):
        """Initialize only when first needed."""
        # Expensive operation
        import time
        time.sleep(5)  # Simulate expensive setup
        return {'data': 'expensive_resource'}
```

---

## Test Isolation & Cleanup

### Resource Cleanup Patterns

```python
class ResourceManager:
    """Ensure proper cleanup for each test."""

    def __init__(self):
        self.resources = []

    def register_resource(self, resource, cleanup_func):
        """Register resource with cleanup."""
        self.resources.append({
            'resource': resource,
            'cleanup': cleanup_func
        })
        return resource

    def cleanup_all(self):
        """Cleanup all registered resources."""
        # Cleanup in reverse order (LIFO)
        for item in reversed(self.resources):
            try:
                item['cleanup'](item['resource'])
            except Exception as e:
                print(f"Cleanup error: {e}")

        self.resources = []
```

### Test Isolation with Docker

```yaml
# docker-compose.test.yml
version: '3'
services:
  postgres-test:
    image: postgres:15
    environment:
      POSTGRES_DB: test_db
      POSTGRES_PASSWORD: test
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis-test:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

---

## Test Flakiness Prevention

### Retry Strategy for Flaky Tests

```python
import pytest
import time

def retry_on_failure(max_attempts=3, delay=1):
    """Decorator to retry flaky tests."""
    def decorator(func):
        def wrapper(*args, **kwargs):
            for attempt in range(max_attempts):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    if attempt == max_attempts - 1:
                        raise
                    print(f"Attempt {attempt+1} failed, retrying...")
                    time.sleep(delay)
        return wrapper
    return decorator

class FlakyTestHandler:
    """Handle flaky tests gracefully."""

    @retry_on_failure(max_attempts=3)
    def test_network_request(self):
        """Test that might fail due to network."""
        response = requests.get('http://api.example.com/data')
        assert response.status_code == 200

    def test_with_proper_waits(self):
        """Proper wait handling for async operations."""
        from selenium.webdriver.support.ui import WebDriverWait
        from selenium.webdriver.support import expected_conditions as EC

        element = WebDriverWait(driver, 10).until(
            EC.presence_of_element_located(('id', 'myElement'))
        )
        assert element is not None
```

### Test Determinism

```python
import random

class DeterministicTest:
    """Ensure tests are deterministic."""

    def test_with_fixed_seed(self):
        """Use fixed seed for reproducible randomness."""
        random.seed(42)
        values = [random.randint(0, 100) for _ in range(10)]
        assert values[0] == 81  # Always same with seed 42

    def test_sorted_collections(self):
        """Never depend on dict/set iteration order."""
        data = {'a': 1, 'b': 2, 'c': 3}
        # WRONG: for key in data.keys()
        # RIGHT: for key in sorted(data.keys())
        keys = sorted(data.keys())
        assert keys == ['a', 'b', 'c']

    def test_datetime_mocking(self, mocker):
        """Mock datetime for consistent testing."""
        mock_time = mocker.patch('time.time')
        mock_time.return_value = 1234567890

        assert get_timestamp() == 1234567890
```

---

## Coverage Optimization

### Smart Coverage Reporting

```python
# .coveragerc
[run]
branch = True
source = src
omit =
    */tests/*
    */test_*.py
    */__main__.py

[report]
exclude_lines =
    pragma: no cover
    def __repr__
    raise AssertionError
    raise NotImplementedError
    if __name__ == .__main__.:
    if TYPE_CHECKING:
    @abstractmethod
    @abstractproperty

precision = 2
show_missing = True
skip_covered = True

[html]
directory = htmlcov
```

### Coverage Gap Analysis

```python
class CoverageAnalyzer:
    """Identify coverage gaps efficiently."""

    def analyze_coverage(self, coverage_report):
        """Find uncovered high-complexity code."""
        gaps = []
        for file, coverage in coverage_report.items():
            if coverage['coverage'] < 80:
                gaps.append({
                    'file': file,
                    'coverage': coverage['coverage'],
                    'uncovered_lines': coverage['missing_lines'],
                    'priority': self._calculate_priority(coverage)
                })

        return sorted(gaps, key=lambda x: x['priority'], reverse=True)

    def _calculate_priority(self, coverage):
        """Prioritize by complexity."""
        # Uncovered complex code = higher priority
        complexity = coverage.get('complexity', 1)
        missing = len(coverage.get('missing_lines', []))
        return complexity * missing
```

---

## CI/CD Optimization

### Test Sharding Across CI Jobs

```yaml
# GitHub Actions
strategy:
  matrix:
    test-shard: [1, 2, 3, 4]
steps:
  - name: Run tests (shard ${{ matrix.test-shard }})
    run: |
      pytest tests/ \
        --shard=${{ matrix.test-shard }}/4 \
        --durations=0 \
        -v
```

### Test Impact Analysis

```python
class TestImpactAnalyzer:
    """Run only affected tests based on code changes."""

    def get_changed_files(self, git_diff_output):
        """Parse git diff to find changed files."""
        changed = set()
        for line in git_diff_output.split('\n'):
            if line.startswith(('M', 'A', 'D')):
                changed.add(line.split()[1])
        return changed

    def get_affected_tests(self, changed_files):
        """Map changed files to test files."""
        affected = set()
        for file in changed_files:
            # src/module.py -> tests/test_module.py
            test_file = file.replace('src/', 'tests/').replace('.py', '_test.py')
            affected.add(test_file)
        return affected
```

---

## Best Practices

**DO**:
- Run fast unit tests first
- Use parallel execution
- Cache expensive fixtures
- Mock external services
- Isolate test databases
- Clean up resources properly
- Monitor test execution time trends

**DON'T**:
- Share state between tests
- Make tests dependent on execution order
- Use sleep() instead of proper waits
- Test implementation details
- Run slow E2E tests for every change
- Skip cleanup in finally blocks
- Ignore flaky test signals

---

**Related Skills**: moai-domain-backend, moai-domain-frontend, moai-essentials-perf
