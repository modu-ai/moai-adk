# Expert Agents Detailed Guide

A complete reference for Alfred's 6 domain experts.

## Overview

| #   | Expert           | Domain                    | Activation Keywords                  | Skills |
| --- | ---------------- | ------------------------- | ----------------------------------- | ------ |
| 1   | backend-expert   | API, Server, DB           | server, api, database, microservice | 12     |
| 2   | frontend-expert  | UI, State Management, Perf| frontend, ui, component, state       | 10     |
| 3   | devops-expert    | Deploy, CI/CD, Infrastructure | deploy, docker, kubernetes, ci/cd | 14     |
| 4   | ui-ux-expert     | Design System, Accessibility | design, ux, accessibility, figma  | 8      |
| 5   | security-expert  | Security, Auth            | security, auth, encryption, owasp   | 11     |
| 6   | database-expert  | DB Design, Optimization   | database, schema, query, index      | 9      |

______________________________________________________________________

## 1. backend-expert

**Domain**: API, Server, Database Architecture

### Activation Conditions

Automatically activated when the following keywords are included in SPEC:

- `server`, `api`, `endpoint`, `microservice`
- `authentication`, `authorization`
- `database`, `ORM`

### Areas of Expertise

| Area              | Tech Stack              | Responsibilities                    |
| ----------------- | ----------------------- | ---------------------------------- |
| **API Design**    | REST, GraphQL           | OpenAPI 3.1 spec writing           |
| **Frameworks**    | FastAPI, Flask, Django  | Framework selection and structure  |
| **Authentication** | JWT, OAuth 2.0, Session | Secure authentication system impl  |
| **Microservices** | Celery, RabbitMQ        | Async task processing              |
| **Caching**       | Redis, Memcached        | Performance optimization            |

### Key Responsibilities

1. **API Design**

   - Follow RESTful principles
   - Design endpoint structure
   - Define request/response schemas
   - Error handling strategy

2. **Data Modeling**

   - Entity-Relationship diagrams
   - ORM model design
   - Relationship setup (1:1, 1:N, N:N)
   - Indexing strategy

3. **Performance Optimization**

   - Query optimization
   - Database indexing
   - Caching strategy
   - Load balancing

### Example: REST API Design

```python
# @CODE:SPEC-002:backend-design
from fastapi import FastAPI, HTTPException, Depends
from sqlalchemy.orm import Session

app = FastAPI(title="Todo API v1.0")

# Endpoint design
@app.post("/api/v1/todos", status_code=201)
async def create_todo(
    title: str,
    description: str,
    db: Session = Depends(get_db)
):
    """Create todo"""
    todo = Todo(title=title, description=description)
    db.add(todo)
    db.commit()
    return todo

@app.get("/api/v1/todos/{todo_id}")
async def get_todo(todo_id: int, db: Session = Depends(get_db)):
    """Get todo"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    return todo

@app.put("/api/v1/todos/{todo_id}")
async def update_todo(
    todo_id: int,
    title: str,
    description: str,
    db: Session = Depends(get_db)
):
    """Update todo"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    todo.title = title
    todo.description = description
    db.commit()
    return todo

@app.delete("/api/v1/todos/{todo_id}")
async def delete_todo(todo_id: int, db: Session = Depends(get_db)):
    """Delete todo"""
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404)
    db.delete(todo)
    db.commit()
    return {"status": "deleted"}
```

### Generated Artifacts

- OpenAPI 3.1 specification
- API endpoint list
- Request/response schemas
- Error code documentation
- Authentication flow

______________________________________________________________________

## 2. frontend-expert

**Domain**: UI Components, State Management, Performance Optimization

### Activation Conditions

Automatically activated when the following keywords are included in SPEC:

- `frontend`, `ui`, `component`, `page`
- `state`, `store`, `context`
- `performance`, `optimization`

### Areas of Expertise

| Area          | Tech Stack                  | Responsibilities           |
| ------------- | --------------------------- | -------------------------- |
| **Frameworks** | React 19, Vue 3.5, Angular 19 | Framework selection and structure |
| **State Mgmt** | Redux, Zustand, Pinia      | Global state design        |
| **Components** | Composition, Hooks         | Reusable components        |
| **Performance** | Bundle optimization, lazy loading | Rendering performance improvement |
| **Accessibility** | WCAG 2.2, ARIA             | Support for all users      |

### Key Responsibilities

1. **Component Design**

   - Reusable component structure
   - Props interface definition
   - Styling strategy (CSS-in-JS, Tailwind)

2. **State Management**

   - Global state structure
   - State update logic
   - Performance optimization (memoization)

3. **Performance Optimization**

   - Bundle size minimization
   - Rendering optimization
   - Image optimization
   - Caching strategy

### Example: React Component Design

```typescript
// @CODE:SPEC-003:frontend-component
import React, { useState, useCallback } from 'react';
import { useTodoStore } from './store';

// State management (Zustand)
const useTodoStore = create((set) => ({
  todos: [],
  addTodo: (todo) => set((state) => ({
    todos: [...state.todos, todo]
  })),
  removeTodo: (id) => set((state) => ({
    todos: state.todos.filter(t => t.id !== id)
  }))
}));

// Reusable component
const TodoItem = React.memo(({ todo, onRemove }) => (
  <div className="todo-item">
    <h3>{todo.title}</h3>
    <p>{todo.description}</p>
    <button onClick={() => onRemove(todo.id)}>
      Delete
    </button>
  </div>
));

// Main component
export const TodoList = () => {
  const [input, setInput] = useState('');
  const { todos, addTodo } = useTodoStore();

  const handleAdd = useCallback(() => {
    if (input.trim()) {
      addTodo({ id: Date.now(), title: input });
      setInput('');
    }
  }, [input]);

  return (
    <div>
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        placeholder="Enter new todo"
      />
      <button onClick={handleAdd}>Add</button>

      <div className="todo-list">
        {todos.map(todo => (
          <TodoItem
            key={todo.id}
            todo={todo}
            onRemove={() => /* remove */}
          />
        ))}
      </div>
    </div>
  );
};
```

### Generated Artifacts

- Component tree diagram
- Props interface definition
- State management diagram
- Performance optimization report
- Accessibility audit results

______________________________________________________________________

## 3. devops-expert

**Domain**: Deployment, CI/CD, Cloud Infrastructure

### Activation Conditions

Automatically activated when the following keywords are included in SPEC:

- `deploy`, `deployment`, `ci/cd`
- `docker`, `kubernetes`
- `infrastructure`, `cloud`

### Areas of Expertise

| Area              | Tech Stack                 | Responsibilities          |
| ----------------- | -------------------------- | ------------------------- |
| **Containers**    | Docker, Docker Compose     | Dockerfile and image mgmt |
| **Orchestration** | Kubernetes, Helm           | Deployment and scaling    |
| **CI/CD**         | GitHub Actions, GitLab CI  | Automation pipeline       |
| **Cloud**         | AWS, GCP, Azure            | Infrastructure code       |
| **Monitoring**    | Prometheus, Grafana        | Performance monitoring    |

### Key Responsibilities

1. **Deployment Pipeline Design**

   - Test → Build → Deploy automation
   - Canary deployment and rollback strategy
   - Zero-downtime deployment

2. **Infrastructure Configuration**

   - Production environment setup
   - Load balancing
   - Database backup/recovery

3. **Monitoring & Logging**

   - Application performance monitoring
   - Log collection and analysis
   - Alert configuration

### Example: GitHub Actions CI/CD

```yaml
# @CODE:SPEC-004:devops-pipeline
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: '3.13'

      # Tests
      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r requirements.txt

      - name: Run tests
        run: pytest --cov=src tests/

      - name: Check coverage
        run: |
          coverage report --fail-under=85

  deploy:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      # Deployment
      - name: Build Docker image
        run: docker build -t app:latest .

      - name: Deploy to production
        run: |
          docker tag app:latest app:${{ github.sha }}
          # Execute deployment script
          ./scripts/deploy.sh
```

### Generated Artifacts

- Dockerfile
- docker-compose.yml
- Kubernetes manifests
- CI/CD pipeline
- Deployment guide
- Monitoring configuration

______________________________________________________________________

## 4. ui-ux-expert

**Domain**: Design System, Accessibility, User Experience

### Activation Conditions

Automatically activated when the following keywords are included in SPEC:

- `design`, `ui`, `ux`
- `accessibility`, `a11y`
- `figma`, `design-system`

### Areas of Expertise

| Area              | Tech Stack         | Responsibilities         |
| ----------------- | ------------------ | ----------------------- |
| **Design System** | Figma, Storybook   | Component library       |
| **Accessibility** | WCAG 2.2, ARIA     | Inclusion for all users |
| **User Research** | User testing, analysis | UX improvement      |
| **Performance**   | Load time, responsiveness | User satisfaction |

### Key Responsibilities

1. **Design System Construction**

   - Color, typography, spacing definition
   - Component library
   - Design tokens

2. **Accessibility Assurance**

   - Screen reader support
   - Keyboard navigation
   - Color contrast

3. **User Experience Improvement**

   - User testing
   - Feedback collection
   - Continuous improvement

### Example: Accessibility Checklist

```markdown
# WCAG 2.2 Accessibility Audit

## Perceivable
- [ ] Provide alt text for images
- [ ] Don't convey information with color alone
- [ ] Contrast ratio 4.5:1 or higher

## Operable
- [ ] All functions operable via keyboard
- [ ] Focus order is logical
- [ ] No flashing content

## Understandable
- [ ] High text readability
- [ ] Predictable navigation
- [ ] Clear error messages

## Robust
- [ ] Valid HTML/CSS
- [ ] Correct ARIA usage
- [ ] Compatibility tests pass
```

### Generated Artifacts

- Design system guide
- Figma component library
- Storybook documentation
- Accessibility audit report
- User testing results

______________________________________________________________________

## 5. security-expert

**Domain**: Security, Authentication, Encryption

### Activation Conditions

Automatically activated when the following keywords are included in SPEC:

- `security`, `auth`, `encryption`
- `vulnerability`, `owasp`
- `compliance`, `privacy`

### Areas of Expertise

| Area          | Tech Stack            | Responsibilities         |
| ------------- | --------------------- | ----------------------- |
| **Auth**       | JWT, OAuth 2.0, SAML  | Secure auth system      |
| **Encryption** | AES-256, RSA, HTTPS   | Data protection         |
| **OWASP**      | Top 10, SAST/DAST      | Vulnerability prevention |
| **Access Ctrl** | RBAC, ABAC           | Permission management   |
| **Audit**      | Logging, monitoring    | Security event tracking |

### Key Responsibilities

1. **Security Design**

   - Threat modeling
   - Security architecture
   - Intrusion prevention strategy

2. **Vulnerability Prevention**

   - SQL injection prevention
   - XSS prevention
   - CSRF tokens
   - Input validation

3. **Security Monitoring**

   - Log analysis
   - Intrusion detection
   - Incident response

### Example: Security Implementation

```python
# @CODE:SPEC-005:security-implementation
from flask import Flask, request
from werkzeug.security import generate_password_hash, check_password_hash
from cryptography.fernet import Fernet
import jwt

app = Flask(__name__)
SECRET_KEY = "your-secret-key"

# Password hashing
def hash_password(password: str) -> str:
    """Safely hash password"""
    return generate_password_hash(password, method='pbkdf2:sha256')

def verify_password(hashed, password: str) -> bool:
    """Verify password"""
    return check_password_hash(hashed, password)

# JWT token
def create_token(user_id: int) -> str:
    """Create JWT token"""
    return jwt.encode(
        {'user_id': user_id},
        SECRET_KEY,
        algorithm='HS256'
    )

# Input validation
@app.before_request
def validate_input():
    """Validate all inputs"""
    if request.method == 'POST':
        # CSRF token validation
        token = request.headers.get('X-CSRF-Token')
        if not verify_csrf_token(token):
            return {'error': 'Invalid CSRF token'}, 403

        # SQL injection prevention (parameterized queries)
        # - Use SQLAlchemy ORM

        # XSS prevention (HTML escape)
        # - Jinja2 auto-escape

# Force HTTPS
@app.after_request
def secure_headers(response):
    """Set security headers"""
    response.headers['X-Content-Type-Options'] = 'nosniff'
    response.headers['X-Frame-Options'] = 'DENY'
    response.headers['X-XSS-Protection'] = '1; mode=block'
    response.headers['Strict-Transport-Security'] = 'max-age=31536000'
    return response
```

### Generated Artifacts

- Security policy document
- Threat modeling diagram
- Security audit checklist
- Penetration test report
- Compliance documents (GDPR, HIPAA)

______________________________________________________________________

## 6. database-expert

**Domain**: Database Design, Optimization, Migration

### Activation Conditions

Automatically activated when the following keywords are included in SPEC:

- `database`, `db`, `schema`
- `query`, `index`, `migration`
- `optimization`, `performance`

### Areas of Expertise

| Area             | Tech Stack                  | Responsibilities        |
| ---------------- | --------------------------- | ----------------------- |
| **Design**       | PostgreSQL, MySQL, MongoDB  | Schema design           |
| **Optimization** | Indexing, query tuning      | Performance improvement |
| **Migration**    | Alembic, Flyway             | Version control          |
| **Scalability**  | Partitioning, sharding      | Large-scale data processing |
| **Backup**       | PITR, replication           | Data safety             |

### Key Responsibilities

1. **Database Design**

   - Entity-Relationship diagrams
   - Normalization (1NF ~ 3NF)
   - Constraint setup

2. **Performance Optimization**

   - Appropriate index creation
   - Query optimization
   - Execution plan analysis

3. **Migration Management**

   - Version control
   - Rollback strategy
   - Zero-downtime migration

### Example: Database Design

```sql
-- @CODE:SPEC-006:database-schema
-- User table
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email)
);

-- Todo table
CREATE TABLE todos (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_completed (user_id, completed)
);

-- Relationship diagram
-- users (1) --< (N) todos
```

### Generated Artifacts

- ERD (Entity-Relationship Diagram)
- DDL (Data Definition Language) scripts
- Migration scripts
- Performance tuning report
- Backup/recovery plan

______________________________________________________________________

## Expert Activation Matrix

| SPEC Keyword | backend | frontend | devops | ui-ux | security | database |
| ------------- | ------- | -------- | ------ | ----- | -------- | -------- |
| API           | ✅      |          |        |       |          |          |
| Frontend      |         | ✅       |        | ✅    |          |          |
| Database      | ✅      |          |        |       |          | ✅       |
| Deploy        |         |          | ✅     |       |          |          |
| Security      |         |          |        |       | ✅       |          |
| Performance   | ✅      | ✅       |        | ✅    |          | ✅       |

______________________________________________________________________

**Next**: [Core Sub-agents](core.md) or [Agents Overview](index.md)



