# setup-supabase
Setup Supabase PostgreSQL database and authentication.

## Usage
```
/setup-supabase [project-name] [--auth] [--rls]
```

## Options
- `--auth`: Configure authentication
- `--rls`: Setup row-level security
- `--realtime`: Enable real-time subscriptions

## What It Does
1. Creates Supabase project
2. Configures PostgreSQL database
3. Sets up authentication providers
4. Enables RLS policies
5. Creates API keys

## Example
```bash
/setup-supabase my-db --auth --rls
```
