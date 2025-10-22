# Language Detection Examples

_Last updated: 2025-10-22_

## Example 1: Python Project Detection

### Project Structure
```
project/
├── pyproject.toml
├── setup.py
├── requirements.txt
└── src/
    └── main.py
```

### Detection Logic
```python
def detect_python_project(root_path):
    indicators = {
        'pyproject.toml': 10,    # High confidence
        'setup.py': 8,
        'requirements.txt': 6,
        'Pipfile': 7,
        'poetry.lock': 9
    }

    score = 0
    for file, weight in indicators.items():
        if (root_path / file).exists():
            score += weight

    return score >= 6  # Threshold
```

### Framework Detection (from pyproject.toml)
```toml
[tool.poetry.dependencies]
django = "^4.2"        # Django framework
fastapi = "^0.109"     # FastAPI framework
flask = "^3.0"         # Flask framework
```

---

## Example 2: TypeScript/JavaScript Detection

### Project Structure
```
project/
├── package.json
├── tsconfig.json
├── node_modules/
└── src/
    └── index.ts
```

### Detection Logic
```typescript
interface ProjectIndicators {
  file: string;
  weight: number;
}

function detectJavaScriptProject(rootPath: string): boolean {
  const indicators: ProjectIndicators[] = [
    { file: 'package.json', weight: 10 },
    { file: 'tsconfig.json', weight: 9 },
    { file: 'bun.lockb', weight: 8 },
    { file: 'package-lock.json', weight: 7 }
  ];

  let score = 0;
  indicators.forEach(({ file, weight }) => {
    if (fs.existsSync(path.join(rootPath, file))) {
      score += weight;
    }
  });

  return score >= 7;
}
```

### Framework Detection (from package.json)
```json
{
  "dependencies": {
    "react": "^18.2.0",           // React framework
    "next": "^14.1.0",            // Next.js framework
    "@nestjs/core": "^10.3.0"     // NestJS framework
  }
}
```

---

## Example 3: Multi-Language Project

### Project Structure
```
fullstack/
├── backend/
│   ├── pyproject.toml       # Python
│   └── src/
├── frontend/
│   ├── package.json         # TypeScript
│   └── src/
└── infrastructure/
    └── main.tf              # Terraform
```

### Detection Result
```json
{
  "primary": "python",
  "languages": [
    {
      "name": "python",
      "confidence": 0.9,
      "framework": "fastapi",
      "path": "backend/"
    },
    {
      "name": "typescript",
      "confidence": 0.85,
      "framework": "react",
      "path": "frontend/"
    },
    {
      "name": "terraform",
      "confidence": 0.7,
      "path": "infrastructure/"
    }
  ]
}
```

---

## Example 4: Go Project Detection

### Project Structure
```
project/
├── go.mod
├── go.sum
├── main.go
└── internal/
    └── handler/
```

### Detection Logic (from go.mod)
```go
module github.com/user/project

go 1.22

require (
    github.com/gin-gonic/gin v1.9.1        // Gin framework
    github.com/labstack/echo/v4 v4.11.4    // Echo framework
)
```

---

_For detection patterns, see reference.md_
