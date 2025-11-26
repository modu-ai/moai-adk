# MCP Server Testing

Comprehensive testing strategies for MCP servers.

---

## Unit Testing with pytest

### Basic Tool Testing

**File: `test_tools.py`**:

```python
import pytest
from fastmcp import FastMCP

@pytest.fixture
def server():
    """Create test server."""
    s = FastMCP("test-server")

    @s.tool()
    def add(a: int, b: int) -> int:
        """Add two numbers."""
        return a + b

    @s.tool()
    def divide(a: float, b: float) -> float:
        """Divide two numbers."""
        if b == 0:
            raise ValueError("Cannot divide by zero")
        return a / b

    return s

class TestBasicTools:
    def test_add_positive_numbers(self, server):
        """Test adding positive numbers."""
        result = server.invoke_tool("add", {"a": 2, "b": 3})
        assert result == 5

    def test_add_negative_numbers(self, server):
        """Test adding negative numbers."""
        result = server.invoke_tool("add", {"a": -1, "b": -2})
        assert result == -3

    def test_divide_valid(self, server):
        """Test division with valid input."""
        result = server.invoke_tool("divide", {"a": 10, "b": 2})
        assert result == 5.0

    def test_divide_by_zero(self, server):
        """Test division by zero raises error."""
        with pytest.raises(ValueError, match="Cannot divide by zero"):
            server.invoke_tool("divide", {"a": 10, "b": 0})

    def test_invalid_parameter_type(self, server):
        """Test invalid parameter type."""
        with pytest.raises(TypeError):
            server.invoke_tool("add", {"a": "not-a-number", "b": 3})
```

---

### Tool with Database

**File: `test_database_tools.py`**:

```python
import pytest
from fastmcp import FastMCP
import sqlite3
from unittest.mock import patch, MagicMock

@pytest.fixture
def db():
    """Create test database."""
    conn = sqlite3.connect(":memory:")
    cursor = conn.cursor()
    cursor.execute("""
        CREATE TABLE users (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL
        )
    """)
    cursor.execute("INSERT INTO users VALUES (1, 'Alice', 'alice@example.com')")
    cursor.execute("INSERT INTO users VALUES (2, 'Bob', 'bob@example.com')")
    conn.commit()
    yield conn
    conn.close()

@pytest.fixture
def server(db):
    """Create server with mock database."""
    s = FastMCP("db-test-server")

    @s.tool()
    def get_user(user_id: int) -> dict:
        """Get user by ID."""
        cursor = db.cursor()
        cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))
        row = cursor.fetchone()
        if not row:
            raise ValueError(f"User not found: {user_id}")
        return {"id": row[0], "name": row[1], "email": row[2]}

    @s.tool()
    def list_users(limit: int = 10) -> list[dict]:
        """List all users."""
        cursor = db.cursor()
        cursor.execute("SELECT * FROM users LIMIT ?", (limit,))
        rows = cursor.fetchall()
        return [{"id": row[0], "name": row[1], "email": row[2]} for row in rows]

    return s

class TestDatabaseTools:
    def test_get_user_exists(self, server):
        """Test getting existing user."""
        result = server.invoke_tool("get_user", {"user_id": 1})
        assert result["name"] == "Alice"
        assert result["email"] == "alice@example.com"

    def test_get_user_not_found(self, server):
        """Test getting non-existent user."""
        with pytest.raises(ValueError, match="User not found"):
            server.invoke_tool("get_user", {"user_id": 999})

    def test_list_users(self, server):
        """Test listing users."""
        result = server.invoke_tool("list_users", {"limit": 10})
        assert len(result) == 2
        assert result[0]["name"] == "Alice"
        assert result[1]["name"] == "Bob"

    def test_list_users_with_limit(self, server):
        """Test listing with limit."""
        result = server.invoke_tool("list_users", {"limit": 1})
        assert len(result) == 1
```

---

### Integration Testing

**File: `test_integration.py`**:

```python
import pytest
from fastmcp import FastMCP
from unittest.mock import patch, MagicMock

@pytest.fixture
def server():
    """Create full integration test server."""
    s = FastMCP("integration-test-server")

    class UserService:
        def __init__(self):
            self.users = {}

        def create(self, email: str, name: str) -> str:
            """Create user and return ID."""
            user_id = str(len(self.users) + 1)
            self.users[user_id] = {"id": user_id, "email": email, "name": name}
            return user_id

        def get(self, user_id: str) -> dict:
            """Get user by ID."""
            if user_id not in self.users:
                raise ValueError(f"User not found: {user_id}")
            return self.users[user_id]

        def list(self) -> list[dict]:
            """List all users."""
            return list(self.users.values())

    service = UserService()

    @s.tool()
    def create_user(email: str, name: str) -> dict:
        """Create new user."""
        user_id = service.create(email, name)
        return {"status": "success", "user_id": user_id}

    @s.tool()
    def get_user(user_id: str) -> dict:
        """Get user."""
        return service.get(user_id)

    @s.tool()
    def list_users() -> list[dict]:
        """List users."""
        return service.list()

    return s

class TestIntegration:
    def test_create_and_get_user(self, server):
        """Test creating and retrieving user."""
        # Create user
        result = server.invoke_tool("create_user", {
            "email": "test@example.com",
            "name": "Test User"
        })
        assert result["status"] == "success"
        user_id = result["user_id"]

        # Get user
        user = server.invoke_tool("get_user", {"user_id": user_id})
        assert user["email"] == "test@example.com"
        assert user["name"] == "Test User"

    def test_workflow_multiple_users(self, server):
        """Test workflow with multiple users."""
        # Create first user
        result1 = server.invoke_tool("create_user", {
            "email": "alice@example.com",
            "name": "Alice"
        })

        # Create second user
        result2 = server.invoke_tool("create_user", {
            "email": "bob@example.com",
            "name": "Bob"
        })

        # List all users
        users = server.invoke_tool("list_users", {})
        assert len(users) == 2
        assert users[0]["name"] == "Alice"
        assert users[1]["name"] == "Bob"
```

---

## Performance Testing

### Load Testing

**File: `test_performance.py`**:

```python
import pytest
import time
from fastmcp import FastMCP

@pytest.fixture
def server():
    """Create server for performance testing."""
    s = FastMCP("perf-test-server")

    @s.tool()
    def slow_operation(duration_ms: int) -> dict:
        """Simulate operation with specified duration."""
        start = time.time()
        time.sleep(duration_ms / 1000)
        elapsed = (time.time() - start) * 1000

        return {
            "requested_duration_ms": duration_ms,
            "actual_duration_ms": elapsed
        }

    return s

class TestPerformance:
    def test_operation_latency(self, server):
        """Test operation completes within acceptable time."""
        start = time.time()
        result = server.invoke_tool("slow_operation", {"duration_ms": 100})
        elapsed = (time.time() - start) * 1000

        # Should complete in ~100ms
        assert 90 < elapsed < 150

    def test_throughput(self, server):
        """Test throughput with multiple calls."""
        start = time.time()
        for _ in range(100):
            server.invoke_tool("slow_operation", {"duration_ms": 10})
        elapsed = (time.time() - start)

        # 100 calls of 10ms each = ~1 second
        throughput = 100 / elapsed
        assert throughput > 50  # At least 50 calls/second
```

---

## Resource Testing

### Resource Access

**File: `test_resources.py`**:

```python
import pytest
from fastmcp import FastMCP
from unittest.mock import MagicMock

@pytest.fixture
def server():
    """Create server with resources."""
    s = FastMCP("resource-test-server")

    users_db = {
        "user-1": {"id": "user-1", "name": "Alice"},
        "user-2": {"id": "user-2", "name": "Bob"}
    }

    @s.resource("user://{user_id}")
    def get_user_resource(user_id: str) -> dict:
        """Get user resource."""
        if user_id not in users_db:
            raise ValueError(f"User not found: {user_id}")
        return users_db[user_id]

    return s

class TestResources:
    def test_get_resource_success(self, server):
        """Test getting existing resource."""
        result = server.get_resource("user://user-1")
        assert result["name"] == "Alice"

    def test_get_resource_not_found(self, server):
        """Test getting non-existent resource."""
        with pytest.raises(ValueError, match="User not found"):
            server.get_resource("user://user-999")
```

---

## Mocking External Services

**File: `test_mocks.py`**:

```python
import pytest
from fastmcp import FastMCP
from unittest.mock import patch, MagicMock

@pytest.fixture
def server():
    """Create server with external service."""
    s = FastMCP("mock-test-server")

    @s.tool()
    def call_external_api(endpoint: str) -> dict:
        """Call external API."""
        import requests
        response = requests.get(f"https://api.example.com/{endpoint}")
        return response.json()

    return s

class TestMockedServices:
    @patch("requests.get")
    def test_api_call_success(self, mock_get, server):
        """Test successful API call."""
        # Setup mock
        mock_response = MagicMock()
        mock_response.json.return_value = {"status": "success", "data": [1, 2, 3]}
        mock_get.return_value = mock_response

        # Test
        result = server.invoke_tool("call_external_api", {"endpoint": "data"})
        assert result["status"] == "success"
        assert result["data"] == [1, 2, 3]

        # Verify mock was called
        mock_get.assert_called_once_with("https://api.example.com/data")

    @patch("requests.get")
    def test_api_call_failure(self, mock_get, server):
        """Test API call failure."""
        # Setup mock to raise exception
        mock_get.side_effect = Exception("Connection error")

        # Test
        with pytest.raises(Exception, match="Connection error"):
            server.invoke_tool("call_external_api", {"endpoint": "data"})
```

---

## Running Tests

### pytest Configuration

**File: `pytest.ini`**:

```ini
[pytest]
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*
addopts = -v --tb=short --cov=. --cov-report=html
```

### Run Tests

```bash
# Run all tests
pytest

# Run specific test file
pytest tests/test_tools.py

# Run specific test class
pytest tests/test_tools.py::TestBasicTools

# Run with coverage
pytest --cov=. --cov-report=html

# Run with verbose output
pytest -v

# Run with specific markers
pytest -m "integration"
```

---

## Continuous Integration

### GitHub Actions

**File: `.github/workflows/test.yml`**:

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: [3.9, "3.10", "3.11"]

    steps:
    - uses: actions/checkout@v3

    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: ${{ matrix.python-version }}

    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt
        pip install pytest pytest-cov

    - name: Run tests
      run: |
        pytest --cov=. --cov-report=xml

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.xml
```

---

## Test Coverage Goals

- **Minimum**: 80% code coverage
- **Target**: 90%+ code coverage
- **Critical paths**: 100% coverage

**Check coverage**:

```bash
pytest --cov=. --cov-report=term-missing
```

---

## Best Practices

✅ **DO**:
- Test happy path and error cases
- Use fixtures for setup/teardown
- Mock external dependencies
- Test pagination and limits
- Verify error messages
- Test with realistic data

❌ **DON'T**:
- Skip testing error cases
- Make tests dependent on external services
- Use hardcoded test data
- Skip integration tests
- Test implementation details only
- Ignore edge cases

---

**Last Updated**: 2025-11-27 | Production-grade testing patterns
