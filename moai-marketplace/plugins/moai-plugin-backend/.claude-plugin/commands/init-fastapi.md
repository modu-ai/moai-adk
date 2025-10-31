# init-fastapi

Initialize a new FastAPI project with professional architecture.

## Description

Scaffolds a production-ready FastAPI project with SQLAlchemy, Alembic, Pydantic, and proper project structure.

## Usage

```
/init-fastapi [project-name]
```

## Parameters

- `project-name`: Name of your FastAPI project (e.g., "user-api")

## What It Does

1. Creates project directory structure
2. Sets up virtual environment
3. Creates requirements.txt with dependencies
4. Initializes SQLAlchemy ORM setup
5. Configures Alembic for migrations
6. Creates main FastAPI application file
7. Sets up example endpoint

## Options

- `--async`: Use async/await patterns (default: true)
- `--database`: PostgreSQL or SQLite (default: PostgreSQL)
- `--with-tests`: Include pytest setup (default: true)

## Example

```bash
/init-fastapi my-api
```

Creates:
```
my-api/
├── app/
│   ├── main.py
│   ├── models.py
│   ├── schemas.py
│   └── routes/
├── migrations/
├── tests/
├── .env.example
├── requirements.txt
└── README.md
```

## Next Steps

1. Activate virtual environment: `source venv/bin/activate`
2. Install dependencies: `pip install -r requirements.txt`
3. Setup database: `alembic upgrade head`
4. Run server: `uvicorn app.main:app --reload`
