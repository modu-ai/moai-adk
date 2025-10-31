# deploy-vercel
Deploy Next.js frontend to Vercel with configuration.

## Usage
```
/deploy-vercel [--prod] [--preview] [--env-sync]
```

## Options
- `--prod`: Deploy to production
- `--preview`: Create preview deployment
- `--env-sync`: Sync environment variables

## What It Does
1. Builds Next.js project
2. Deploys to Vercel
3. Configures environment variables
4. Sets up custom domain (if configured)
5. Runs performance checks

## Example
```bash
/deploy-vercel --prod
```
