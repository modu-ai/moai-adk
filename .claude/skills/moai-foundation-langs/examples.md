# Foundation Langs Examples

## Language Detection Examples

### Example 1: Node.js/JavaScript Project Detection

**Project Structure**:
```
my-web-app/
├─ package.json
├─ tsconfig.json
├─ src/
│  └─ index.ts
└─ tests/
   └─ index.test.ts
```

**Detection Output**:
```
Detected Language: JavaScript/TypeScript
Primary Config: package.json
Version Info:
  - Node.js: 22.11.0 (from engines.node)
  - TypeScript: 5.9.2 (from devDependencies)
  - Test Framework: Vitest 2.1.0

Recommended Stack:
  - Runtime: Node.js 22.x LTS (support until 2027-04-30)
  - Framework: Express.js 4.21.x OR Fastify 4.28.x
  - Testing: Vitest 2.x (native support for TypeScript)
  - Linting: ESLint 9.x + Prettier 3.3.x
```

---

### Example 2: Python Project Detection

**Project Structure**:
```
data-science-app/
├─ pyproject.toml
├─ src/
│  └─ main.py
└─ tests/
   └─ test_main.py
```

**pyproject.toml Content**:
```toml
[project]
name = "data-science-app"
version = "0.1.0"
requires-python = ">=3.11"
dependencies = [
    "pandas>=2.2.0",
    "scikit-learn>=1.5.0",
    "numpy>=2.1.0"
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"
```

**Detection Output**:
```
Detected Language: Python
Primary Config: pyproject.toml
Version Info:
  - Python: 3.13.x (latest stable)
  - Minimum Required: >=3.11

Current Dependencies:
  - pandas: 2.2.x (Data manipulation)
  - scikit-learn: 1.5.x (Machine learning)
  - numpy: 2.1.x (Numerical computing)

Recommended Tools:
  - Package Manager: uv (fastest, modern)
  - Testing: Pytest 8.3.x
  - Linting: Ruff 0.8.x (replaces black+isort+flake8)
  - Type Checking: Pyright 1.1.x (VS Code native)
```

---

### Example 3: Go Project Detection

**Project Structure**:
```
microservice-app/
├─ go.mod
├─ go.sum
├─ main.go
└─ handlers/
   └─ api.go
```

**go.mod Content**:
```
module github.com/example/microservice

go 1.25

require (
    github.com/gin-gonic/gin v1.10.0
    github.com/lib/pq v1.10.9
)
```

**Detection Output**:
```
Detected Language: Go
Primary Config: go.mod
Version Info:
  - Go: 1.25.x (released Aug 2025)
  - Framework: Gin 1.10.x
  - Database: PostgreSQL driver pq 1.10.x

Recommended Stack:
  - Web Framework: Gin 1.10.x (high performance)
  - Alternative: Fiber 3.0.x (Express.js-like)
  - Testing: testing (stdlib) + Testify 1.9.x
  - Code Quality: golangci-lint 1.61.x
```

---

## Version Management Examples

### Example 1: Python Version Pinning Strategy

**Development Dependencies**:
```toml
# pyproject.toml
[project.optional-dependencies]
dev = [
    "pytest==8.3.0",              # Exact pin for test framework
    "ruff==0.8.2",                # Exact pin for code quality
    "pyright==1.1.382",           # Exact pin for type checking
]

[project.dependencies]
requests = "^2.31.0"              # Allow patch updates (2.31.x)
fastapi = "^0.115.0"              # Allow minor updates (0.115.x)
pydantic = "^2.6.0"               # Allow minor updates (2.6.x)
```

**Version Update Strategy**:
```
Quarterly Review Schedule:
1. Week 1: Scan for security updates
   $ pip-audit

2. Week 2: Check compatibility
   $ poetry show --outdated

3. Week 3: Test with latest minor versions
   $ poetry update --dry-run

4. Week 4: Merge safe updates to develop
   $ poetry update <package>
   $ pytest              # Full test suite
   $ git push

Urgent: Security patches
   $ pip-audit --fix
   $ pytest
   $ git commit (fast-track to main)
```

---

### Example 2: Node.js Version Pinning

**package.json**:
```json
{
  "name": "web-app",
  "version": "1.0.0",
  "engines": {
    "node": "22.11.0",
    "npm": "11.x"
  },
  "devDependencies": {
    "typescript": "5.9.2",
    "vitest": "^2.1.0",
    "eslint": "^9.15.0",
    "prettier": "3.3.0"
  },
  "dependencies": {
    "express": "^4.21.0",
    "axios": "^1.7.0"
  }
}
```

**.nvmrc File**:
```
22.11.0
```

**Installation Workflow**:
```bash
# Load correct Node.js version
nvm use         # Reads .nvmrc

# Install exact dependencies
npm ci           # Uses package-lock.json (CI-safe)

# Verify versions
node --version   # 22.11.0
npm --version    # 11.x
```

---

### Example 3: Rust Dependency Management

**Cargo.toml**:
```toml
[package]
name = "web-service"
version = "0.1.0"
edition = "2021"

[dependencies]
tokio = { version = "1.41", features = ["full"] }
axum = "0.8"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
uuid = { version = "1.0", features = ["v4"] }

[dev-dependencies]
tokio-test = "0.4"
criterion = "0.5"

[profile.release]
opt-level = 3
lto = true
codegen-units = 1
```

**Update Strategy**:
```bash
# Check for updates
cargo outdated

# Test with newer versions (dry-run)
cargo update --aggressive

# Run full test suite
cargo test --all
cargo bench --bench main

# Commit if all tests pass
git add Cargo.lock
git commit -m "chore: update dependencies"
```

---

## Best Practices Examples

### Example 1: Python Testing Setup with Pytest

**tests/conftest.py**:
```python
"""
Shared pytest configuration and fixtures.

"""
import pytest
from pathlib import Path

@pytest.fixture
def project_root():
    """Return project root directory."""
    return Path(__file__).parent.parent

@pytest.fixture
def temp_dir(tmp_path):
    """Return temporary directory for test files."""
    return tmp_path

@pytest.fixture
def sample_data():
    """Return sample test data."""
    return {
        "id": 1,
        "name": "Test Item",
        "status": "active"
    }
```

**tests/test_core.py**:
```python
"""
Test core functionality.

"""
import pytest
from src.core import process_data

class TestProcessData:
    """Test data processing function."""
    
    def test_process_valid_data(self, sample_data):
        """Test processing valid input."""
        result = process_data(sample_data)
        assert result["status"] == "processed"
    
    def test_process_empty_input(self):
        """Test handling empty input."""
        with pytest.raises(ValueError):
            process_data({})
    
    @pytest.mark.parametrize("input,expected", [
        ({"value": 1}, 1),
        ({"value": 100}, 100),
    ])
    def test_parametrized(self, input, expected):
        """Test with multiple inputs."""
        result = process_data(input)
        assert result["value"] == expected
```

**Run Tests**:
```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=src --cov-report=html

# Run specific test
pytest tests/test_core.py::TestProcessData::test_process_valid_data

# Watch mode (auto-run on file change)
pytest-watch
```

---

### Example 2: TypeScript Testing with Vitest

**vitest.config.ts**:
```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      lines: 85,
      functions: 85,
      branches: 85,
      statements: 85,
    },
  },
});
```

**tests/core.test.ts**:
```typescript
import { describe, it, expect } from 'vitest';
import { processData } from '../src/core';

/**
 */
describe('processData', () => {
  it('should process valid data', () => {
    const input = { id: 1, name: 'Test' };
    const result = processData(input);
    expect(result).toHaveProperty('status');
    expect(result.status).toBe('processed');
  });

  it('should throw on invalid input', () => {
    expect(() => processData({})).toThrow(ValueError);
  });

  it.each([
    [{ value: 1 }, 1],
    [{ value: 100 }, 100],
  ])('should handle value %j', (input, expected) => {
    const result = processData(input);
    expect(result.value).toBe(expected);
  });
});
```

**Run Tests**:
```bash
# Run all tests
vitest

# Watch mode (default in dev)
vitest watch

# Coverage report
vitest run --coverage

# Run UI
vitest --ui
```

---

### Example 3: Go Testing with Table-Driven Tests

**handlers/api.go**:
```go
package handlers

import (
    "fmt"
)

func ProcessData(input map[string]interface{}) (map[string]interface{}, error) {
    if len(input) == 0 {
        return nil, fmt.Errorf("empty input")
    }
    return map[string]interface{}{"status": "processed"}, nil
}
```

**handlers/api_test.go**:
```go
package handlers

import (
    "testing"
)

func TestProcessData(t *testing.T) {
    tests := []struct {
        name    string
        input   map[string]interface{}
        wantErr bool
        want    string
    }{
        {
            name:    "valid input",
            input:   map[string]interface{}{"id": 1},
            wantErr: false,
            want:    "processed",
        },
        {
            name:    "empty input",
            input:   map[string]interface{}{},
            wantErr: true,
            want:    "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ProcessData(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessData() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && got["status"].(string) != tt.want {
                t.Errorf("ProcessData() = %v, want %v", got["status"], tt.want)
            }
        })
    }
}
```

**Run Tests**:
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

### Example 4: Rust Testing and Benchmarking

**src/lib.rs**:
```rust
/// Process data according to specification.
///
pub fn process_data(input: &str) -> Result<String, String> {
    if input.is_empty() {
        return Err("empty input".to_string());
    }
    Ok(format!("processed: {}", input))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_process_valid_data() {
        let result = process_data("test");
        assert_eq!(result.unwrap(), "processed: test");
    }

    #[test]
    fn test_process_empty_input() {
        let result = process_data("");
        assert!(result.is_err());
    }

    #[test]
    fn test_multiple_inputs() {
        let inputs = vec!["a", "ab", "abc"];
        for input in inputs {
            let result = process_data(input);
            assert!(result.is_ok());
        }
    }
}
```

**benches/benchmark.rs**:
```rust
use criterion::{black_box, criterion_group, criterion_main, Criterion};
use my_crate::process_data;

fn bench_process_data(c: &mut Criterion) {
    c.bench_function("process_short", |b| {
        b.iter(|| process_data(black_box("test")))
    });
    
    c.bench_function("process_long", |b| {
        b.iter(|| process_data(black_box("a".repeat(1000))))
    });
}

criterion_group!(benches, bench_process_data);
criterion_main!(benches);
```

**Run Tests**:
```bash
# Run unit tests
cargo test

# Run with output
cargo test -- --nocapture

# Run benchmarks
cargo bench

# Generate flamegraph
cargo install flamegraph
cargo flamegraph --bench benchmark
```

---

## Security Best Practices Examples

### Example 1: Dependency Vulnerability Scanning (Python)

**GitHub Actions Workflow**:
```yaml
name: Security Scan

on: [push, pull_request]

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-python@v5
        with:
          python-version: '3.13'
      
      - name: Install dependencies
        run: |
          pip install --upgrade pip
          pip install pip-audit
      
      - name: Audit packages
        run: pip-audit
      
      - name: Check lock file
        run: |
          pip install poetry
          poetry check --lock
```

---

### Example 2: Dependency Vulnerability Scanning (JavaScript)

**npm audit in CI**:
```bash
# Check for vulnerabilities
npm audit

# Auto-fix where possible
npm audit fix

# Fail CI if vulnerabilities found
npm audit --audit-level=high
```

**Dependabot Configuration** (.github/dependabot.yml):
```yaml
version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
    reviewers:
      - "security-team"
    
  - package-ecosystem: "pip"
    directory: "/"
    schedule:
      interval: "weekly"
```

---

### Example 3: Security Scanning in Go

**golangci-lint Configuration** (.golangci.yml):
```yaml
linters:
  enable:
    - gosec      # Security scanning
    - govet      # Vet
    - staticcheck

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec    # Tests can use weak crypto for testing
```

**GitHub Actions**:
```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v4
  with:
    version: latest
    args: --timeout 5m
```

---

## Migration Examples

### Example 1: Python 3.11 → 3.13 Migration

**Step 1: Update pyproject.toml**:
```toml
# Before
requires-python = ">=3.11"

# After
requires-python = ">=3.13"
```

**Step 2: Check for deprecations**:
```bash
python -W all -m py_compile src/
```

**Step 3: Update dependencies**:
```bash
uv pip compile pyproject.toml --python-version 3.13
```

**Step 4: Run tests**:
```bash
pytest --cov=src
```

---

### Example 2: Node.js 18 → 22 Migration

**Step 1: Update .nvmrc and package.json**:
```bash
echo "22.11.0" > .nvmrc
nvm use 22.11.0
```

**Step 2: Update lock file**:
```bash
npm install  # Updates package-lock.json
```

**Step 3: Check for breaking changes**:
```bash
npm audit
npm outdated
```

**Step 4: Run full test suite**:
```bash
npm test
npm run build
```

---

## Summary

These examples demonstrate:
- Language detection patterns for 10+ languages
- Version pinning strategies
- Testing frameworks per language
- Security scanning workflows
- Dependency management approaches
- Migration paths

**Use when**: You need concrete implementation examples for language selection, testing setup, or version management.
