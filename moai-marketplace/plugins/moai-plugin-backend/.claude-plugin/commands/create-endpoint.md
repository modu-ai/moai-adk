# create-endpoint

Generate a new FastAPI endpoint with validation and documentation.

## Usage

```
/create-endpoint [method] [path] [--resource-name]
```

## Parameters

- `method`: HTTP method (GET, POST, PUT, DELETE)
- `path`: API path (e.g., /users/{id})
- `--resource-name`: Resource type for endpoint

## What It Does

1. Creates endpoint function with async support
2. Adds Pydantic model validation
3. Generates OpenAPI documentation
4. Includes error handling
5. Sets up request/response schemas

## Example

```bash
/create-endpoint POST /users --resource-name User
```

Creates endpoint with user creation logic, validation, and docs.
