# db-setup

Setup and configure database with migrations and initial data.

## Usage

```
/db-setup [--reset] [--seed]
```

## Parameters

- `--reset`: Drop and recreate all tables
- `--seed`: Populate with sample data

## What It Does

1. Creates database connection
2. Runs Alembic migrations
3. Validates schema
4. Seeds initial data (optional)

## Example

```bash
/db-setup --seed
```
