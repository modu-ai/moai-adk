# deploy-render
Deploy FastAPI backend to Render.

## Usage
```
/deploy-render [service-name] [--prod] [--env]
```

## Options
- `--prod`: Deploy to production
- `--env`: Configure environment variables
- `--scale`: Auto-scaling configuration

## What It Does
1. Builds Docker image
2. Deploys service to Render
3. Configures health checks
4. Sets up environment variables
5. Monitors deployment status

## Example
```bash
/deploy-render api --prod --env
```
