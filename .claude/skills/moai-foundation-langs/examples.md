# Language Detection Examples

_Last updated: 2025-10-22_

## Example 1: Next.js (TypeScript) Project

**Project Structure**:
```
my-nextjs-app/
├── package.json
├── tsconfig.json
├── next.config.js
├── src/
│   ├── app/
│   └── components/
└── tests/
```

**Detection Result**:
```json
{
  "primary": "typescript",
  "framework": "nextjs",
  "testing": "vitest",
  "recommended_skills": [
    "moai-lang-typescript",
    "moai-domain-frontend"
  ]
}
```

---

## Example 2: FastAPI (Python) Project

**Project Structure**:
```
my-api/
├── pyproject.toml
├── src/
│   └── api/
└── tests/
```

**Detection Result**:
```json
{
  "primary": "python",
  "framework": "fastapi",
  "testing": "pytest",
  "recommended_skills": [
    "moai-lang-python",
    "moai-domain-backend"
  ]
}
```

---

**For complete detection patterns, see [reference.md](reference.md)**
