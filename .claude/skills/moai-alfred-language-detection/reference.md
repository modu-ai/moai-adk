# Language Detection Reference

_Last updated: 2025-10-22_

## Detection Patterns by Language

### Python
**Primary Indicators**:
- `pyproject.toml` (modern, PEP 518)
- `setup.py` (legacy)
- `requirements.txt`
- `Pipfile` / `Pipfile.lock` (pipenv)
- `poetry.lock` (poetry)
- `pdm.lock` (pdm)
- `uv.lock` (uv)

**Framework Detection** (from dependencies):
- Django: `django` package
- FastAPI: `fastapi` package
- Flask: `flask` package
- Pyramid: `pyramid` package

### JavaScript/TypeScript
**Primary Indicators**:
- `package.json`
- `tsconfig.json` (TypeScript)
- `bun.lockb` (Bun)
- `pnpm-lock.yaml` (pnpm)
- `yarn.lock` (Yarn)
- `package-lock.json` (npm)

**Framework Detection** (from dependencies):
- React: `react`, `react-dom`
- Vue: `vue`
- Angular: `@angular/core`
- Svelte: `svelte`
- Next.js: `next`
- Nuxt: `nuxt`

### Go
**Primary Indicators**:
- `go.mod`
- `go.sum`
- `*.go` files

**Framework Detection** (from imports):
- Gin: `github.com/gin-gonic/gin`
- Echo: `github.com/labstack/echo`
- Fiber: `github.com/gofiber/fiber`

### Rust
**Primary Indicators**:
- `Cargo.toml`
- `Cargo.lock`
- `*.rs` files

**Framework Detection** (from dependencies):
- Actix Web: `actix-web`
- Rocket: `rocket`
- Axum: `axum`

---

## Configuration File Parsing

### Python (pyproject.toml)
```python
import toml

def detect_python_framework(project_root):
    pyproject_path = project_root / "pyproject.toml"
    if not pyproject_path.exists():
        return None

    data = toml.load(pyproject_path)

    # Poetry format
    deps = data.get("tool", {}).get("poetry", {}).get("dependencies", {})

    # PEP 621 format
    if not deps:
        deps = data.get("project", {}).get("dependencies", [])

    frameworks = {
        "django": "Django",
        "fastapi": "FastAPI",
        "flask": "Flask"
    }

    for pkg, name in frameworks.items():
        if pkg in str(deps).lower():
            return name

    return None
```

### JavaScript (package.json)
```javascript
function detectFramework(packageJson) {
  const deps = {
    ...packageJson.dependencies,
    ...packageJson.devDependencies
  };

  const frameworks = {
    'react': 'React',
    'vue': 'Vue',
    '@angular/core': 'Angular',
    'svelte': 'Svelte',
    'next': 'Next.js'
  };

  for (const [pkg, name] of Object.entries(frameworks)) {
    if (pkg in deps) {
      return name;
    }
  }

  return null;
}
```

---

## Confidence Scoring

### Scoring Algorithm
```python
def calculate_confidence(indicators_found, total_indicators):
    """
    Calculate confidence score (0.0 - 1.0)

    indicators_found: Number of matching indicators
    total_indicators: Total possible indicators
    """
    if total_indicators == 0:
        return 0.0

    base_score = indicators_found / total_indicators

    # Boost for critical indicators
    critical_boost = 0.2 if has_critical_indicator() else 0.0

    # Penalty for conflicting indicators
    conflict_penalty = 0.1 if has_conflicts() else 0.0

    confidence = min(1.0, base_score + critical_boost - conflict_penalty)
    return confidence
```

### Confidence Levels

| Score | Level | Action |
|-------|-------|--------|
| 0.9 - 1.0 | Very High | Proceed with confidence |
| 0.7 - 0.89 | High | Safe to proceed |
| 0.5 - 0.69 | Medium | Request confirmation |
| 0.3 - 0.49 | Low | Manual selection needed |
| 0.0 - 0.29 | Very Low | Cannot detect |

---

## Multi-Language Projects

### Detection Strategy
1. Scan root directory for all config files
2. Identify language per subdirectory
3. Determine primary language (most files)
4. List secondary languages
5. Return hierarchical structure

### Example Output
```json
{
  "primary_language": "python",
  "languages": [
    {
      "language": "python",
      "framework": "django",
      "version": "3.11",
      "path": "backend/",
      "confidence": 0.95
    },
    {
      "language": "typescript",
      "framework": "react",
      "version": "5.3",
      "path": "frontend/",
      "confidence": 0.90
    }
  ],
  "tooling": {
    "python": ["pytest", "ruff", "mypy"],
    "typescript": ["vitest", "biome", "tsc"]
  }
}
```

---

## References

- [Python Packaging (PEP 518)](https://peps.python.org/pep-0518/)
- [Node.js package.json spec](https://docs.npmjs.com/cli/v10/configuring-npm/package-json)
- [Go Modules Reference](https://go.dev/ref/mod)
- [Cargo Book](https://doc.rust-lang.org/cargo/)

---

_For detection examples, see examples.md_
