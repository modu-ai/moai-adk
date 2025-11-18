---
name: "moai-alfred-env-security"
version: "0.26.0"
created: 2025-11-18
updated: 2025-11-18
status: stable
tier: specialization
description: ".env Security Best Practices for MoAI-ADK. Complete guide for environment variable management, secrets protection, multi-environment setup, and incident response."
allowed-tools: "Read, Edit, Bash"
primary-agent: "security-expert"
secondary-agents: ["devops-expert", "backend-expert"]
keywords: [".env", "secrets", "security", "environment", "credentials", "gitignore"]
tags: ["security", "devops", "best-practices"]
orchestration:
  multi_agent: false
  supports_chaining: false
can_resume: false
typical_chain_position: "planning"
depends_on: []
---

# moai-alfred-env-security

**.env Security Best Practices for MoAI-ADK**

> **Primary Agent**: security-expert
> **Secondary Agents**: devops-expert, backend-expert
> **Version**: 0.26.0
> **Keywords**: env, secrets, security, credentials, environment

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (50 lines)

**Core Purpose**: Manage environment variables securely using local development flexibility and production-grade security defaults.

**Hybrid Security Approach**:

| Aspect | Package Template | Local Development | Production |
|---|---|---|---|
| **.env access** | Denied (secure-by-default) | Permitted (dev convenience) | Prohibited (secrets mgmt) |
| **Settings file** | settings.json | settings.local.json | Platform secrets |
| **Credentials** | Never stored | Local-only development | AWS/GitHub/Vercel secrets |
| **Scope** | New projects | Current project | Deployed environment |

**Quick Setup**:

```bash
# 1. Create local override file
mkdir -p .claude
cat > .claude/settings.local.json << 'EOF'
{
  "permissions": {
    "allow": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Write(./.env)",
      "Edit(./.env)"
    ]
  }
}
EOF

# 2. Add to .gitignore (CRITICAL!)
echo ".claude/settings.local.json" >> .gitignore
echo ".env*" >> .gitignore

# 3. Restart Claude Code

# 4. Never commit credentials!
```

**Essential .gitignore Patterns**:

```gitignore
# Environment files (ALL variations)
.env*

# Platform secrets
.vercel/
.netlify/
.firebase/
.aws/credentials
.env.local
.env.*.local
.env.production.local

# SSH and credentials
.ssh/
private-keys/
secrets/
credentials.json
```

---

### Level 2: Core Implementation (150 lines)

**Multi-Environment Configuration**

#### Development Environment (.env.local)

```bash
# .env.local (Git ignored - safe for local-only values)

# API Keys (development/testing only)
DATABASE_URL=postgresql://localhost:5432/dev_db
API_KEY_DEVELOPMENT=dev-key-12345
STRIPE_TEST_KEY=sk_test_...

# Feature Flags (local testing)
ENABLE_DEBUG=true
LOG_LEVEL=debug
ENABLE_ANALYTICS=false

# Local Services
REDIS_URL=redis://localhost:6379
SMTP_SERVER=localhost:1025

# Purpose: Local development only
# Never use these in production!
```

#### Test Environment (.env.test)

```bash
# .env.test (Git ignored - for test runs)

DATABASE_URL=postgresql://localhost:5432/test_db
API_KEY_TESTING=test-key-67890
STRIPE_TEST_KEY=sk_test_...

# Test-specific settings
NODE_ENV=test
SEED_DATABASE=true
MOCK_EXTERNAL_APIS=true

# Purpose: Running automated tests
# Isolated from development/production
```

#### Production Environment (Platform Secrets)

```bash
# ❌ NEVER use .env files in production!

# ✅ USE platform-specific secret management:

# GitHub Secrets
PRODUCTION_DATABASE_URL=*** (in GitHub Secrets)
PRODUCTION_API_KEY=*** (in GitHub Secrets)
STRIPE_LIVE_KEY=*** (in GitHub Secrets)

# Vercel Environment Variables
Settings → Environment Variables → Production
- DATABASE_URL
- API_KEY
- STRIPE_KEY

# AWS Systems Manager Parameter Store
/myapp/prod/database-url
/myapp/prod/api-key
/myapp/prod/stripe-key
```

---

**Security Policies**

#### Local Development Policy

```yaml
File: .env.local
Access: Developer machine only
Credentials: Development/test only
Secrets: Never store production keys
Git: Always .gitignore
Backup: Optional (local machine backup)
```

#### Template/Documentation Policy

```yaml
File: .env.example
Access: Public (committed to Git)
Content: Variable names only, no values
Purpose: Document required environment variables
Example:
  DATABASE_URL=postgresql://localhost/myapp
  API_KEY=your_api_key_here
  STRIPE_KEY=sk_test_...
```

#### Production Policy

```yaml
Storage: Platform-specific secrets vault
Examples:
  - GitHub: Organization Secrets + Repository Secrets
  - Vercel: Project → Settings → Environment Variables
  - AWS: Systems Manager Parameter Store
  - Azure: Key Vault
  - Heroku: Config Vars
Access: Deployment automation only
Rotation: Every 90 days (minimum)
Audit: All access logged
```

---

**Claude Code Integration**

#### settings.local.json Pattern

```json
{
  "permissions": {
    "allow": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Write(./.env)",
      "Edit(./.env)"
    ]
  }
}
```

**Why This Pattern**:
- Package template: Default deny (secure-by-default)
- Local project: Optional allow (developer convenience)
- User controls: Via settings.local.json (explicit)
- Version control: Never tracked (Git ignored)

#### settings.json (Package Template - Default Deny)

```json
{
  "permissions": {
    "deniedTools": [
      "Edit(.env*)",
      "Write(.env*)"
    ]
  }
}
```

---

### Level 3: Incident Response (200+ lines)

**What to Do if Credentials Are Compromised**

#### Scenario 1: Credentials Accidentally Committed

```bash
# 1. STOP - Don't continue working
# All committed credentials are now public!

# 2. Immediately regenerate credentials
# Example: Stripe API key
# → Go to Stripe Dashboard → API Keys → Rotate keys

# 3. Remove from Git history (force!)
git filter-branch --tree-filter 'rm -f .env .env.local' HEAD

# ⚠️ WARNING: This rewrites history!
# Only do this if credentials were committed

# 4. Force push to remote (dangerous!)
git push origin --force

# 5. Audit GitHub access logs
GitHub → Settings → Audit Log
  - Check for unauthorized access
  - Review who accessed secrets
  - Check deployment history

# 6. Check Stripe audit log
Stripe Dashboard → Developers → Webhooks → Event logs
  - Look for suspicious API calls
  - Check transaction history
  - Monitor for fraud

# 7. Update all .env files with new credentials
.env.local → Regenerated keys
.env.test → New test keys
production → New production keys via platform
```

**Real Example**:

```bash
# Oops! Committed .env with Stripe key

# Step 1: Check what was committed
git log -p --all | grep "sk_live_" | head -1
# Output: sk_live_Abc123Xyz789 (EXPOSED!)

# Step 2: Regenerate in Stripe
# Stripe API Keys → Rotate key
# New key: sk_live_Def456Uvw012

# Step 3: Remove from history
git filter-branch --tree-filter 'rm -f .env .env.local' HEAD

# Step 4: Force push
git push origin --force

# Step 5: Monitor for abuse
# Check Stripe for unauthorized charges

# Resolution: All clear ✅
```

---

#### Scenario 2: .vercel/project.json Exposed

**What's at Risk**:
```json
// .vercel/project.json (exposes project metadata)
{
  "projectId": "prj_Abc123xyz789",
  "orgId": "team_Def456uvw012",
  "name": "myapp-production"
}
```

**Attack Vector**:
```bash
# Attacker gets project ID
# → Access Vercel API with stolen token
# → Deploy malicious code
# → Access environment variables
# → Read all secrets!
```

**Recovery**:

```bash
# 1. Remove .vercel from repository
git rm -r .vercel
git commit -m "Remove Vercel config"
git push

# 2. Rotate Vercel tokens
Vercel → Settings → Tokens → Create new token
(Revoke old compromised token)

# 3. Verify deployments
Vercel Dashboard → Deployments
- Check for suspicious builds
- Review deployment logs

# 4. Update GitHub Secrets
GitHub → Settings → Secrets → Update VERCEL_TOKEN

# 5. Re-deploy with new token
git push origin main → Triggers CI/CD with new token
```

---

#### Scenario 3: AWS Credentials Leaked

**Immediate Actions**:

```bash
# 1. Identify which credentials exposed
grep -r "AKIA" .  # AWS Access Key pattern
grep -r "wJalrXUtnFEMI" .  # AWS Secret Key pattern

# 2. Deactivate immediately (AWS Console)
IAM → Users → Security Credentials → Deactivate

# 3. Check CloudTrail for abuse
CloudTrail → Event history
- Who accessed what
- When (timestamps)
- What was changed
- Estimated blast radius

# 4. Kill active sessions
AWS CLI: aws iam delete-access-key --access-key-id AKIA...

# 5. Create new credentials
IAM → Users → Create access key (new)

# 6. Update application
.env.production → New credentials (via platform)
CI/CD → Update AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY

# 7. Monitor billing
AWS → Billing → Check for unusual charges
Set billing alert: Alert if >$100/day
```

---

**Monitoring & Prevention**

#### Pre-Commit Hooks

```bash
# .git/hooks/pre-commit (prevent .env commits)

#!/bin/bash
# Prevent committing .env files

if git diff --cached --name-only | grep -E '\.env'; then
  echo "❌ Error: Attempting to commit .env file!"
  echo "❌ .env files must NEVER be committed to git"
  exit 1
fi

# Check for common credential patterns
if git diff --cached | grep -E "sk_live_|AKIA|BEGIN RSA PRIVATE KEY"; then
  echo "❌ Error: Attempting to commit credentials!"
  exit 1
fi

exit 0
```

**Installation**:
```bash
chmod +x .git/hooks/pre-commit
```

#### GitHub Secret Scanning

```yaml
# GitHub automatically detects committed credentials

# If credentials are found:
# - GitHub alerts repository owner
# - Provides notification with exact location
# - Shows commit hash and timestamp
# - Blocks public exposure

# Required: If credentials found:
# 1. Acknowledge the alert
# 2. Regenerate the exposed credentials
# 3. Remove from git history (git filter-branch)
```

#### Automated Scanning Tools

```bash
# Install git-secrets (prevents commits)
brew install git-secrets
git secrets --install
git secrets --register-aws

# Scan for patterns
git secrets --scan

# truffleHog (scan historical commits)
pip install truffleHog
truffleHog --entropy filesystem /path/to/repo
```

---

## 🎯 Real-World Workflows

### Workflow 1: New Developer Setup

```bash
# 1. Clone repository
git clone https://github.com/myorg/myapp

# 2. Copy example file
cp .env.example .env.local

# 3. Get actual credentials from team
# → Slack message with credentials
# → OR secure 1Password/LastPass link
# → OR GitHub secret access

# 4. Edit .env.local
nano .env.local
# DATABASE_URL=postgresql://...
# API_KEY=sk_test_...
# (Now local to their machine)

# 5. Add to .gitignore (automatic)
# .gitignore already has .env*

# 6. Verify setup
npm run dev
# Server starts ✅
```

### Workflow 2: GitHub Actions CI/CD

```yaml
# .github/workflows/deploy.yml

name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Deploy
        env:
          # GitHub Secrets (never exposed in logs!)
          DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
          API_KEY: ${{ secrets.PROD_API_KEY }}
          STRIPE_KEY: ${{ secrets.STRIPE_LIVE_KEY }}
        run: |
          npm run build
          npm run deploy

      # Logs NEVER show actual secrets!
```

### Workflow 3: Rotating Credentials Quarterly

```bash
# Every 3 months, rotate credentials

# 1. Generate new credentials
AWS IAM → Create new access key
Stripe → Rotate API key
GitHub → Create new personal token

# 2. Update production environment
Vercel → Environment Variables → Update values
GitHub → Settings → Secrets → Update values
AWS → .env.production → New keys

# 3. Test with new credentials
Deploy to staging → Test all integrations
Monitor logs for errors

# 4. Deploy to production
Merge to main → CI/CD deploys with new keys

# 5. Revoke old credentials
AWS → Delete old access key
Stripe → Revoke old API key
GitHub → Delete old token

# 6. Document rotation
CHANGELOG.md → "Rotated Q4 credentials"
```

---

## 🔧 Complete Setup Checklist

```bash
# 1. .gitignore Configuration
echo ".env*" >> .gitignore
echo ".vercel/" >> .gitignore
echo ".aws/credentials" >> .gitignore
echo ".firebase/" >> .gitignore
echo ".claude/settings.local.json" >> .gitignore
git add .gitignore && git commit -m "Security: Add .gitignore patterns"

# 2. .env.example Creation
cat > .env.example << 'EOF'
# Database
DATABASE_URL=postgresql://user:pass@localhost/dbname

# API Keys
API_KEY=your_api_key_here
STRIPE_KEY=sk_test_...

# Feature Flags
DEBUG=false
LOG_LEVEL=info
EOF
git add .env.example && git commit -m "docs: Add .env.example template"

# 3. Local Settings (not committed)
mkdir -p .claude
cat > .claude/settings.local.json << 'EOF'
{
  "permissions": {
    "allow": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Write(./.env)",
      "Edit(./.env)"
    ]
  }
}
EOF

# 4. Production Secrets Setup
# → GitHub: Settings → Secrets
# → Vercel: Project → Settings → Environment Variables
# → AWS: Systems Manager Parameter Store

# 5. Pre-commit Hook
mkdir -p .git/hooks
chmod +x .git/hooks/pre-commit

# 6. Documentation
echo "See SETUP.md for .env configuration" >> README.md

✅ Security setup complete!
```

---

## 📊 Best Practices

### ✅ Do's

- ✅ Use .env.example for documentation
- ✅ Add .env* to .gitignore
- ✅ Use platform secrets for production
- ✅ Rotate credentials every 90 days
- ✅ Use local .claude/settings.local.json for Claude Code
- ✅ Enable GitHub secret scanning
- ✅ Monitor audit logs regularly
- ✅ Document how to set up environments

### ❌ Don'ts

- ❌ Commit .env files to Git
- ❌ Use production keys in development
- ❌ Hardcode credentials in code
- ❌ Share credentials via Slack/email
- ❌ Forget to .gitignore secret files
- ❌ Use same credentials across environments
- ❌ Skip credential rotation
- ❌ Log credentials in debug output

---

## 🔗 Related Skills

- **moai-domain-security** - General security patterns
- **moai-cc-hooks** - Claude Code hooks
- **moai-baas-vercel-ext** - Vercel secrets management

---

**Last Updated**: 2025-11-18
**Version**: 0.26.0
**Status**: Production Ready
