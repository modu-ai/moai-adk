# Backend Plugin

**FastAPI 0.120.2 + uv scaffolding** — SQLAlchemy 2.0, Alembic migrations, Pydantic 2.12, async/await patterns, REST API design.

## 🎯 What It Does

Build production-ready FastAPI backends with professional architecture patterns:

```bash
/plugin install moai-plugin-backend
```

**Automatically provides**:
- 🏗️ FastAPI async endpoint patterns and OpenAPI documentation
- 🗄️ SQLAlchemy ORM with proper relationships and migrations
- 🔐 Request validation with Pydantic models
- 📊 Database indexing and query optimization
- ⚡ Async/await patterns for high performance
- 🔄 Alembic migration management

## 🏗️ Architecture

### 4 Specialist Agents

| Agent | Role | When to Use |
|-------|------|------------|
| **FastAPI Specialist** | Endpoint creation, async patterns, OpenAPI | Building REST APIs |
| **Backend Architect** | System design, scalability, performance | Architecture decisions |
| **Database Expert** | Schema design, indexes, migrations | Database setup |
| **API Designer** | Resource modeling, REST principles | API endpoint design |

### 4 Skills

1. **moai-lang-fastapi-patterns** — Async route handlers, dependency injection, Pydantic models
2. **moai-lang-python** — Python 3.13+, type hints, async/await, context managers
3. **moai-domain-backend** — Backend patterns, scalability, security, error handling
4. **moai-domain-database** — SQL, indexes, migrations, relationships, optimization

## ⚡ Quick Start

### Installation

```bash
/plugin install moai-plugin-backend
```

### Use with MoAI-ADK

The backend plugin provides agents for your development workflow:

1. **Design API endpoints** - FastAPI specialist handles endpoint creation
2. **Build database layer** - Database expert manages schema and migrations
3. **Optimize queries** - Backend architect optimizes performance
4. **Document APIs** - OpenAPI documentation automatically generated

### Example: Create FastAPI Project

Using the agents:
- **FastAPI Specialist**: Designs async endpoint patterns
- **Backend Architect**: Plans system architecture
- **Database Expert**: Designs schema and migrations
- **API Designer**: Creates resource models

## 📊 Typical Workflow

```
Project Setup
    ↓
[FastAPI Specialist]
├─ Initialize FastAPI app
├─ Setup uvicorn configuration
└─ Create project structure
    ↓
[Database Expert]
├─ Design database schema
├─ Create SQLAlchemy models
└─ Initialize Alembic
    ↓
[API Designer]
├─ Define resource models (Pydantic)
├─ Design endpoint routes
└─ Create OpenAPI documentation
    ↓
[Backend Architect]
├─ Optimize queries and indexes
├─ Add caching strategies
└─ Implement error handling
```

## 🎨 Features

### Async Framework
- Native async/await support
- Async SQLAlchemy with proper session management
- Non-blocking I/O for high concurrency
- Connection pooling and optimization

### Database Management
- SQLAlchemy ORM for type safety
- Alembic for version control of schema changes
- Migration scripts with rollback capability
- Proper relationships (one-to-many, many-to-many)

### API Development
- Automatic OpenAPI documentation
- Request validation with Pydantic
- Custom error responses
- Rate limiting and middleware patterns

### Development Workflow
- Local development with hot reload
- Testing patterns (pytest, fixtures)
- Environment-based configuration
- Logging and monitoring setup

## 📚 Skills Explained

### moai-lang-fastapi-patterns
Essential FastAPI patterns for modern API development:
- **Async Route Handlers** - Define non-blocking endpoints
- **Dependency Injection** - Reusable request dependencies
- **Pydantic Models** - Request/response validation
- **Middleware** - Cross-cutting concerns
- **Background Tasks** - Async job execution

### moai-lang-python
Python 3.13+ best practices:
- **Type Hints** - Static type checking with mypy
- **Async/Await** - Coroutines and async patterns
- **Context Managers** - Resource management
- **Generators** - Memory-efficient iteration
- **Decorators** - Function enhancement patterns

### moai-domain-backend
Backend architecture and patterns:
- **Request/Response Cycle** - HTTP lifecycle
- **Error Handling** - Graceful failure modes
- **Security** - Authentication, authorization
- **Performance** - Caching, optimization
- **Scalability** - Horizontal scaling patterns

### moai-domain-database
Database design and optimization:
- **Schema Design** - Normalization and relationships
- **Indexes** - Query optimization
- **Migrations** - Version control for schema
- **Transactions** - ACID compliance
- **Performance Tuning** - Query analysis

## 🚀 Common Use Cases

### Building REST APIs
Use the FastAPI Specialist to:
1. Design endpoint routes
2. Create Pydantic models
3. Generate OpenAPI docs
4. Handle errors gracefully

### Database-First Development
Use the Database Expert to:
1. Design schemas
2. Create relationships
3. Generate migrations
4. Optimize queries

### Scaling Applications
Use the Backend Architect to:
1. Identify bottlenecks
2. Implement caching
3. Optimize database queries
4. Add monitoring

### Async Patterns
Use FastAPI Specialist for:
1. Async route handlers
2. Background tasks
3. Streaming responses
4. WebSocket endpoints

## 🔗 Integration with Other Plugins

- **Frontend Plugin**: Consume APIs built with this plugin
- **DevOps Plugin**: Deploy FastAPI backends to Render, Vercel
- **Database Plugin** (future): Advanced PostgreSQL patterns

## 📖 Documentation

- FastAPI Official: https://fastapi.tiangolo.com
- SQLAlchemy: https://docs.sqlalchemy.org
- Alembic: https://alembic.sqlalchemy.org
- Pydantic: https://docs.pydantic.dev

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - See LICENSE file for details

---

**Created by**: GOOS
**Version**: 1.0.0-dev
**Status**: Development
**Updated**: 2025-10-31
