# Code Templates

Production-ready boilerplates for rapid project scaffolding.

## Template Categories

```
Code Templates Library:
 Backend
 FastAPI (REST API, async, Pydantic validation)
 Django (ORM, admin, authentication)
 Express.js (Node.js, middleware, routing)
 Frontend
 React (hooks, context, TypeScript)
 Next.js 15 (App Router, RSC, Suspense)
 Vue 3 (Composition API, Pinia, TypeScript)
 Infrastructure
 Docker (multi-stage, optimization)
 CI/CD (GitHub Actions, pytest, coverage)
 Kubernetes (deployment, service, configmap)
 Testing
 Pytest (fixtures, mocks, parametrize)
 Vitest (React components, hooks)
 Playwright (E2E, page objects)
```

## FastAPI REST API Template

```python
# Scaffolded FastAPI project structure
my-api/
 app/
 __init__.py
 main.py # FastAPI app initialization
 api/
 v1/
 endpoints/
 router.py
 core/
 config.py # Settings (Pydantic)
 security.py # Auth (JWT)
 db/
 session.py # DB session
 base.py # Base model
 models/
 schemas/ # Pydantic schemas
 services/
 tests/
 conftest.py # pytest fixtures
 test_api/
 alembic/ # DB migrations
 .env.example
 Dockerfile
 docker-compose.yml
 pyproject.toml
 README.md
```

## React Component Template

```typescript
// Scaffolded React component (TypeScript)
import React, { useState, useEffect } from 'react';

interface ComponentProps {
 title: string;
 onAction: () => void;
}

export const Component: React.FC<ComponentProps> = ({
 title,
 onAction
}) => {
 const [state, setState] = useState<string>('');

 useEffect(() => {
 // Initialization logic
 }, []);

 return (
 <div className="component">
 <h1>{title}</h1>
 <button onClick={onAction}>Action</button>
 </div>
 );
};

export default Component;
```

## Usage Examples

```python
# Generate FastAPI project structure
template = load_template("backend/fastapi")
project = template.scaffold(
 name="my-api",
 features=["auth", "database", "celery"],
 customizations={"db": "postgresql"}
)
```

## Template Variables

```json
{
 "variables": {
 "PROJECT_NAME": "my-project",
 "AUTHOR": "John Doe",
 "LICENSE": "MIT",
 "PYTHON_VERSION": "3.13"
 },
 "files": {
 "pyproject.toml": "substitute",
 "README.md": "substitute",
 "src//*.py": "copy"
 }
}
```

## Success Metrics

- Scaffold Time: 2 minutes for new projects (vs 30 minutes manual)
- Template Adoption: 95% of projects use templates
