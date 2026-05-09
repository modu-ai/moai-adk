---
name: moai-lang-python
description: >
  Python 3.13+ development specialist covering FastAPI, Django, async patterns, data science, testing with pytest, and modern Python features. Use when developing Python APIs, web applications, data pipelines, or writing tests.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Bash(python:*) Bash(python3:*) Bash(pytest:*) Bash(ruff:*) Bash(pip:*) Bash(uv:*) Bash(mypy:*) Bash(pyright:*) Bash(black:*) Bash(poetry:*) mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "language, python, fastapi, django, pytest, async, data-science"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Python", "Django", "FastAPI", "Flask", "asyncio", "pytest", "pyproject.toml", "requirements.txt", ".py"]
  languages: ["python"]
---

## Quick Reference (30 seconds)

Python 3.13+ Development Specialist - FastAPI, Django, async patterns, pytest, and modern Python features.

Auto-Triggers: Python files with .py extension, pyproject.toml, requirements.txt, pytest.ini, FastAPI or Django discussions

Core Capabilities:

- Python 3.13 Features: JIT compiler via PEP 744, GIL-free mode via PEP 703, pattern matching with match and case statements
- Web Frameworks: FastAPI 0.115 and later, Django 5.2 LTS
- Data Validation: Pydantic v2.9 with model_validate patterns
- ORM: SQLAlchemy 2.0 async patterns
- Testing: pytest with fixtures, async testing, parametrize decorators
- Package Management: poetry, uv, pip with pyproject.toml
- Type Hints: Protocol, TypeVar, ParamSpec, and modern typing patterns
- Async: asyncio, async generators, and task groups
- Data Science: numpy, pandas, and polars basics

### Quick Patterns

FastAPI Endpoint Pattern:

Import FastAPI and Depends from fastapi, and BaseModel from pydantic. Create a FastAPI application instance. Define a UserCreate model class inheriting from BaseModel with name and email string fields. Create an async post endpoint at the users path that accepts a UserCreate parameter and returns a User by calling UserService.create with await.

Pydantic v2.9 Validation Pattern:

Import BaseModel and ConfigDict from pydantic. Define a User class inheriting from BaseModel. Set model_config using ConfigDict with from_attributes set to True and str_strip_whitespace set to True. Add id as integer, name as string, and email as string fields. Use model_validate to create from ORM objects and model_validate_json to create from JSON data.

pytest Async Test Pattern:

Import pytest and mark the test function with pytest.mark.asyncio decorator. Create an async test function that takes async_client as a fixture parameter. Send a post request to the users endpoint with a JSON body containing a name field. Assert that the response status_code equals 201.

---

## Implementation Guide (5 minutes)

### Python 3.13 New Features

JIT Compiler via PEP 744:

- Experimental feature disabled by default
- Enable using the PYTHON_JIT environment variable set to 1
- Build option available as enable-experimental-jit flag
- Provides performance improvements for CPU-bound code
- Uses copy-and-patch JIT that translates specialized bytecode to machine code

GIL-Free Mode via PEP 703:

- Experimental free-threaded build available as python3.13t
- Allows true parallel thread execution
- Available in official Windows and macOS installers
- Best suited for CPU-intensive multi-threaded applications
- Not recommended for production use yet

Pattern Matching with match and case:

Create a process_response function that takes a response dictionary and returns a string. Use match statement on response. For case with status ok and data field, return success message with the data. For case with status error and message field, return error message. For case with status matching pending or processing using a guard condition, return in progress message. For default case using underscore, return unknown response.

### FastAPI 0.115+ Patterns

Async Dependency Injection:

Import FastAPI, Depends from fastapi, AsyncSession from sqlalchemy.ext.asyncio, and asynccontextmanager from contextlib. Create a lifespan async context manager decorated with asynccontextmanager that takes the FastAPI app. In the lifespan, call await init_db for startup, yield, then call await cleanup for shutdown. Create the FastAPI app with the lifespan parameter. Define an async get_db function returning AsyncGenerator of AsyncSession that uses async with on async_session and yields the session. Create a get endpoint for users with user_id path parameter, using Depends with get_db to inject the database session. Call await get_user_by_id and return UserResponse.model_validate with the user.

Class-Based Dependencies:

Create a Paginator class with an init method accepting page defaulting to 1 and size defaulting to 20. Set self.page to max of 1 and page, self.size to min of 100 and max of 1 and size, and self.offset to page minus 1 multiplied by size. Create a list_items endpoint using Depends on Paginator to inject pagination and return items using get_page with offset and size.

### Django 5.2 LTS Features

Composite Primary Keys:

Create an OrderItem model with ForeignKey to Order with CASCADE deletion, ForeignKey to Product with CASCADE deletion, and an IntegerField for quantity. In the Meta class, set pk to models.CompositePrimaryKey with order and product fields.

URL Reverse with Query Parameters:

Import reverse from django.urls. Call reverse with the search view name, query dictionary containing q set to django and page set to 1, and fragment set to results. The result is the search path with query string and fragment.

Automatic Model Imports in Shell:

Run python manage.py shell and models from all installed apps are automatically imported without explicit import statements.

### Pydantic v2.9 Deep Patterns

Reusable Validators with Annotated:

Import Annotated from typing and AfterValidator and BaseModel from pydantic. Define a validate_positive function that takes an integer v and returns an integer. If v is less than or equal to 0, raise ValueError with must be positive message. Otherwise return v. Create PositiveInt as Annotated with int and AfterValidator using validate_positive. Use PositiveInt in model fields for price and quantity.

Model Validator for Cross-Field Validation:

Import BaseModel and model_validator from pydantic, and Self from typing. Create a DateRange model with start_date and end_date as date fields. Add a model_validator decorator with mode set to after. In the validate_dates method returning Self, check if end_date is before start_date and raise ValueError if so, otherwise return self.

ConfigDict Best Practices:

Create a BaseSchema model with model_config set to ConfigDict. Set from_attributes to True for ORM object support, populate_by_name to True to allow aliases, extra to forbid to fail on unknown fields, and str_strip_whitespace to True to clean strings.

### SQLAlchemy 2.0 Async Patterns

Engine and Session Setup:

Import create_async_engine, async_sessionmaker, and AsyncSession from sqlalchemy.ext.asyncio. Create engine using create_async_engine with the postgresql+asyncpg connection string, pool_pre_ping set to True, and echo set to True. Create async_session using async_sessionmaker with the engine, class_ set to AsyncSession, and expire_on_commit set to False to prevent detached instance errors.

Repository Pattern:

Create a UserRepository class with an init method taking an AsyncSession. Define an async get_by_id method that executes a select query with a where clause for user_id, returning scalar_one_or_none result. Define an async create method that creates a User from UserCreate model_dump, adds to session, commits, refreshes, and returns the user.

Streaming Large Results:

Create an async stream_users function that takes an AsyncSession. Call await db.stream with the select User query. Use async for to iterate over result.scalars and yield each user.

### pytest Advanced Patterns

Async Fixtures with pytest-asyncio:

Import pytest, pytest_asyncio, and AsyncClient from httpx. Decorate fixtures with pytest_asyncio.fixture. Create an async_client fixture that uses async with on AsyncClient with app and base_url, yielding the client. Create a db_session fixture that uses async with on async_session and session.begin, yielding session and calling await session.rollback.

Parametrized Tests:

Use pytest.mark.parametrize decorator with input_data and expected_status parameter names. Provide test cases as tuples with dictionaries and expected status codes. Add ids for valid, empty_name, and missing_name cases. The test function takes async_client, input_data, and expected_status, posts to users endpoint, and asserts status_code matches expected.

Fixture Factories:

Create a user_factory fixture that returns an async function. The inner function takes db as AsyncSession and keyword arguments. Set defaults dictionary with name and email. Create User with defaults merged with kwargs using the pipe operator, add to db, commit, and return user.

### Type Hints Modern Patterns

Protocol for Structural Typing:

Import Protocol and runtime_checkable from typing. Apply runtime_checkable decorator. Define a Repository Protocol with generic type T. Add abstract async get method taking int id returning T or None, async create method taking dict data returning T, and async delete method taking int id returning bool.

ParamSpec for Decorators:

Import ParamSpec, TypeVar, and Callable from typing, and wraps from functools. Define P as ParamSpec and R as TypeVar. Create a retry decorator function taking times defaulting to 3 that returns a callable wrapper. The inner decorator wraps the function and the wrapper iterates for the specified times, trying to await the function and re-raising on the last attempt.

### Package Management

pyproject.toml with Poetry:

In the tool.poetry section, set name, version, and python version constraint. Under dependencies, add fastapi, pydantic, and sqlalchemy with asyncio extra. Under dev dependencies, add pytest, pytest-asyncio, and ruff. Configure ruff with line-length and target-version. Set pytest asyncio_mode to auto in ini_options.

uv Fast Package Manager:

Install uv using curl with the install script from astral.sh. Create virtual environment with uv venv. Install dependencies with uv pip install from requirements.txt. Add dependencies with uv add command.

---

## Advanced Implementation (10+ minutes)

For comprehensive coverage including:

- Production deployment patterns for Docker and Kubernetes
- Advanced async patterns including task groups and semaphores
- Data science integration with numpy, pandas, and polars
- Performance optimization techniques
- Security best practices following OWASP patterns
- CI/CD integration patterns

See:

- reference.md for complete reference documentation
- examples.md for production-ready code examples

---

## Context7 Library Mappings

- tiangolo/fastapi for FastAPI async web framework
- django/django for Django web framework
- pydantic/pydantic for data validation with type annotations
- sqlalchemy/sqlalchemy for SQL toolkit and ORM
- pytest-dev/pytest for testing framework
- numpy/numpy for numerical computing
- pandas-dev/pandas for data analysis library
- pola-rs/polars for fast DataFrame library

---

## Works Well With

- moai-domain-backend for REST API and microservices architecture
- moai-domain-database for SQL patterns and ORM optimization
- moai-workflow-testing for DDD and testing strategies
- moai-essentials-debug for AI-powered debugging
- moai-foundation-quality for TRUST 5 quality principles

---

## Troubleshooting

Common Issues:

Python Version Check:

Run python with version flag to verify 3.13 or later. Use python with -c flag to print sys.version_info for detailed version information.

Async Session Detached Error:

Set expire_on_commit to False in session configuration. Alternatively, use await session.refresh with the object after commit.

pytest asyncio Mode Warning:

In pyproject.toml under tool.pytest.ini_options, set asyncio_mode to auto and asyncio_default_fixture_loop_scope to function.

Pydantic v2 Migration:

The parse_obj method is now model_validate. The parse_raw method is now model_validate_json. The from_orm functionality requires from_attributes set to True in ConfigDict.

---

Last Updated: 2026-01-11
Status: Active (v1.1.0)
