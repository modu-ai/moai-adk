# GitHub Actions for Documentation Deployment

## Overview

This directory contains GitHub Actions workflows for automatically deploying documentation to Vercel.

## Workflows

### `deploy-docs.yml`

Automatically deploys documentation to Vercel when changes are made to the `docs/` directory.

**Triggers:**
- **Pull Request**: Creates a preview deployment when PR includes changes to `docs/`
- **Push to develop/main**: Deploys to production when changes are pushed

## Setup Instructions

### 1. Get Vercel Credentials

```bash
# Install Vercel CLI globally (if not already installed)
bun add -g vercel

# Login to Vercel
vercel login

# Link your project (run in project root)
cd /Users/goos/MoAI/MoAI-ADK
vercel link

# Get your Vercel Token
# Visit: https://vercel.com/account/tokens
# Create a new token with name: "GitHub Actions - MoAI-ADK"
```

### 2. Get Project IDs

```bash
# Get Organization ID and Project ID
cat .vercel/project.json
```

Output example:
```json
{
  "projectId": "prj_370IeY7AeOXToCkUrRLQlDQUBKjj",
  "orgId": "team_Zv0jP5JyxzA17P1RpojlM0VO"
}
```

### 3. Add GitHub Secrets

Go to your GitHub repository:
`https://github.com/modu-ai/moai-adk/settings/secrets/actions`

Add the following secrets:

| Secret Name | Value | Where to find |
|-------------|-------|---------------|
| `VERCEL_TOKEN` | Your Vercel token | https://vercel.com/account/tokens |
| `VERCEL_ORG_ID` | `team_Zv0jP5JyxzA17P1RpojlM0VO` | From `.vercel/project.json` |
| `VERCEL_PROJECT_ID` | `prj_370IeY7AeOXToCkUrRLQlDQUBKjj` | From `.vercel/project.json` |

### 4. Test the Workflow

```bash
# Make a change to docs
echo "test" >> docs/index.md

# Commit and push
git add docs/index.md
git commit -m "test: trigger docs deployment"
git push origin develop

# Check GitHub Actions tab to see the deployment
```

## Workflow Features

### Preview Deployments (Pull Requests)
- 🔍 Automatic preview URL for every PR
- 💬 Bot comments with preview link
- 🚀 Deploy before merge to test changes

### Production Deployments (Push)
- ✅ Automatic deployment to https://moai-adk.vercel.app
- 📊 Deployment summary in GitHub Actions
- 🔐 Only triggers for `develop` and `main` branches

## Troubleshooting

### Deployment fails with "Missing VERCEL_TOKEN"
- Make sure you've added all three secrets to GitHub repository settings

### Preview URL not commented on PR
- Check if the GitHub token has permission to comment on PRs
- Verify the workflow has `write` permission for pull requests

### Build fails
- Check the build logs in GitHub Actions
- Verify `vercel.json` and `package.json` are correct
- Test build locally: `bun run docs:build`

## Monitoring

- **GitHub Actions**: https://github.com/modu-ai/moai-adk/actions
- **Vercel Dashboard**: https://vercel.com/goos/moai-adk
- **Production URL**: https://moai-adk.vercel.app