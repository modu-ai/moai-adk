# Python 3.13 CLI Reference — Tool Command Matrix

**Framework**: Python 3.13.1 + pytest 8.4.2 + ruff 0.13.1 + mypy 1.8.0 + uv 0.9.3

---

## Python Runtime Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `python --version` | Check Python version | `python --version` → `Python 3.13.1` |
| `python -m venv .venv` | Create virtual environment | `python -m venv .venv` |
| `python -m pip list` | List installed packages | `python -m pip list` |
| `python -c "code"` | Execute Python code inline | `python -c "import sys; print(sys.version)"` |
| `python script.py` | Run Python script | `python script.py` |
| `python -m module` | Run module as script | `python -m pytest` |
| `python -i script.py` | Run script with interactive shell | `python -i script.py` |
| `python --help` | Show Python help | `python --help` |

---

## Virtual Environment Management (uv)

| Command | Purpose | Example |
|---------|---------|---------|
| `uv venv` | Create virtual environment | `uv venv` → creates `.venv/` |
| `uv venv --python 3.13` | Create with specific Python | `uv venv --python 3.13` |
| `source .venv/bin/activate` | Activate (macOS/Linux) | `source .venv/bin/activate` |
| `.venv\Scripts\activate` | Activate (Windows) | `.venv\Scripts\activate` |
| `deactivate` | Deactivate virtual environment | `deactivate` |
| `which python` | Check active Python | `which python` → `.venv/bin/python` |

---

## Package Management (uv 0.9.3)

| Command | Purpose | Example |
|---------|---------|---------|
| `uv add package` | Add package | `uv add pytest ruff mypy` |
| `uv add --dev package` | Add dev dependency | `uv add --dev pytest pytest-cov` |
| `uv sync` | Install from lock file | `uv sync` |
| `uv pip list` | List packages | `uv pip list` |
| `uv pip show package` | Show package details | `uv pip show pytest` |
| `uv remove package` | Remove package | `uv remove pytest` |
| `uv update` | Update all packages | `uv update` |
| `uv publish` | Publish to PyPI | `uv publish` |
| `uv build` | Build distribution | `uv build` |

---

## Testing Framework (pytest 8.4.2)

| Command | Purpose | Example |
|---------|---------|---------|
| `pytest` | Run all tests | `pytest` |
| `pytest tests/` | Run tests in directory | `pytest tests/` |
| `pytest file.py` | Run tests in file | `pytest tests/test_calculator.py` |
| `pytest -v` | Verbose output | `pytest -v` |
| `pytest -vv` | Very verbose output | `pytest -vv` |
| `pytest -s` | Show print statements | `pytest -s` |
| `pytest -x` | Stop on first failure | `pytest -x` |
| `pytest -k "pattern"` | Run matching tests | `pytest -k "test_add"` |
| `pytest -m asyncio` | Run async tests only | `pytest -m asyncio` |
| `pytest --cov=src` | Coverage report | `pytest --cov=src --cov-report=term` |
| `pytest --cov-report=html` | HTML coverage report | `pytest --cov-report=html` |
| `pytest --tb=short` | Short traceback format | `pytest --tb=short` |
| `pytest --tb=long` | Long traceback format | `pytest --tb=long` |
| `pytest --lf` | Run last failed tests | `pytest --lf` |
| `pytest --ff` | Fail-first (last failures first) | `pytest --ff` |
| `pytest --maxfail=3` | Stop after 3 failures | `pytest --maxfail=3` |
| `pytest -n 4` | Run in parallel (4 workers) | `pytest -n 4` |
| `pytest --junitxml=report.xml` | JUnit XML report | `pytest --junitxml=report.xml` |
| `pytest --co` | Collect tests only (don't run) | `pytest --co` |
| `pytest --setup-show` | Show fixture setup/teardown | `pytest --setup-show` |

**Coverage Quality Gate**:
```bash
# Ensure ≥85% coverage
pytest --cov=src --cov-report=term --cov-fail-under=85
```

---

## Code Quality Linter (ruff 0.13.1)

### Check Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `ruff check .` | Check all files | `ruff check .` |
| `ruff check src/` | Check directory | `ruff check src/` |
| `ruff check file.py` | Check file | `ruff check src/calculator.py` |
| `ruff check . --fix` | Check and auto-fix | `ruff check . --fix` |
| `ruff check --show-fixes` | Show what would be fixed | `ruff check --show-fixes` |
| `ruff check --select E501` | Check specific rule | `ruff check --select E501` |
| `ruff check --ignore E501` | Ignore specific rule | `ruff check --ignore E501` |
| `ruff check --statistics` | Show violation statistics | `ruff check --statistics` |

### Format Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `ruff format .` | Format all files | `ruff format .` |
| `ruff format src/` | Format directory | `ruff format src/` |
| `ruff format file.py` | Format file | `ruff format src/calculator.py` |
| `ruff format --check` | Check if formatting needed | `ruff format --check` |

### Rule Categories

| Category | Rules | Example |
|----------|-------|---------|
| **E** | Errors | `E501` (line too long) |
| **F** | PyFlakes (logic errors) | `F401` (unused import) |
| **W** | Warnings | `W291` (trailing whitespace) |
| **I** | Import sorting | `I001` (import sort order) |
| **C** | Complexity | `C901` (function too complex) |

**pyproject.toml Configuration**:
```toml
[tool.ruff]
line-length = 100
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "W", "I", "C901"]

[tool.ruff.format]
indent-width = 4
```

---

## Type Checking (mypy 1.8.0)

| Command | Purpose | Example |
|---------|---------|---------|
| `mypy .` | Type check all files | `mypy .` |
| `mypy src/` | Type check directory | `mypy src/` |
| `mypy file.py` | Type check file | `mypy src/models.py` |
| `mypy --strict .` | Strict type checking | `mypy --strict .` |
| `mypy --strict --show-error-codes` | Strict with error codes | `mypy --strict --show-error-codes` |
| `mypy --show-column-numbers .` | Precise error locations | `mypy --show-column-numbers .` |
| `mypy --html .` | Generate HTML report | `mypy --html .` |
| `mypy --help` | Show mypy help | `mypy --help` |

**Common Error Codes**:

| Code | Meaning | Example |
|------|---------|---------|
| `error: Untyped definition` | Function missing type hints | `def func(x): ...` |
| `error: Call to untyped function` | Calling function without types | Depends on context |
| `error: Incompatible types` | Type mismatch | `x: int = "string"` |
| `error: Missing return statement` | Function might not return | Missing `return` in all paths |

**pyproject.toml Configuration**:
```toml
[tool.mypy]
python_version = "3.13"
strict = true
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
```

---

## Python 3.13 Features

### New Syntax Examples

| Feature | Syntax | Example |
|---------|--------|---------|
| **Type Parameters (PEP 695)** | `class Foo[T]:` | `class Stack[T]: def push(self, item: T): ...` |
| **F-String Nesting (PEP 701)** | `f"{f'{x}'}"` | `f"Result: {f'{value:>10}'}"` |
| **Override Decorator (PEP 698)** | `@override` | `@override def method(self): ...` |
| **Match Statement** | `match value:` | Pattern matching expressions |

### Introspection Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `python -c "import sys; print(sys.version_info)"` | Check version info | `sys.version_info(major=3, minor=13, ...)` |
| `python -m inspect module` | Inspect module | `python -m inspect src.models` |
| `python -c "import typing; print(dir(typing))"` | List typing utilities | Shows all available type hints |

---

## Performance Profiling

| Command | Purpose | Example |
|---------|---------|---------|
| `python -m cProfile script.py` | Profile script | `python -m cProfile src/main.py` |
| `python -m timeit "code"` | Time code execution | `python -m timeit "sum(range(100))"` |
| `python -m pdb script.py` | Debug with pdb | `python -m pdb src/main.py` |
| `python -m trace --trace script.py` | Trace execution | `python -m trace --trace src/main.py` |

---

## Combined Workflow (Quality Gate)

**Before Commit** (all must pass):

```bash
# 1. Run tests with coverage (≥85%)
pytest --cov=src --cov-report=term --cov-fail-under=85

# 2. Check linting
ruff check .

# 3. Format code
ruff format .

# 4. Type check (strict mode)
mypy --strict .

# 5. All pass → commit
echo "✅ Quality gates passed!"
git add .
git commit -m "feat: implement feature with full coverage"
```

---

## Environment Setup Script

**`setup-dev.sh`** (complete dev environment):

```bash
#!/bin/bash
set -e

echo "🔧 Setting up Python 3.13 development environment..."

# Create and activate venv
uv venv --python 3.13
source .venv/bin/activate

# Install dependencies
echo "📦 Installing dependencies..."
uv add pytest ruff mypy fastapi uvicorn pydantic sqlalchemy alembic

# Install dev dependencies
echo "🛠️ Installing dev dependencies..."
uv add --dev pytest-asyncio pytest-cov pytest-mock coverage

# Verify setup
echo "✅ Verification:"
python --version
pytest --version
ruff --version
mypy --version

echo "🎉 Dev environment ready!"
```

**Run**:
```bash
chmod +x setup-dev.sh
./setup-dev.sh
```

---

## CI/CD Integration (GitHub Actions)

**`.github/workflows/test.yml`**:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: [3.12, 3.13]

    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: ${{ matrix.python-version }}

      - name: Install uv
        run: pip install uv

      - name: Create venv
        run: uv venv

      - name: Install dependencies
        run: uv sync

      - name: Run tests
        run: pytest --cov=src --cov-report=xml --cov-fail-under=85

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.xml

      - name: Lint check
        run: ruff check .

      - name: Type check
        run: mypy --strict .
```

---

**Version**: 0.1.0
**Created**: 2025-10-22
**Framework**: Python 3.13.1 CLI Tools Reference
