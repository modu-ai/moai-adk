---
name: "moai-alfred-config-schema"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-12
status: stable
description: Enterprise configuration schema validation and management orchestrator with JSON Schema v2024-12, Context7 integration, semantic versioning compliance, environment variable management, secrets handling, multi-environment support, and configuration-as-code best practices; activates for config validation, schema enforcement, environment setup, secrets management, and configuration audits
keywords: ['configuration-schema', 'json-schema', 'config-management', 'environment-management', 'secrets-handling', 'semantic-versioning', 'configuration-as-code', 'schema-validation', 'multi-environment', 'enterprise-config']
allowed-tools: 
  - Read
  - Write
  - Edit
  - Bash
  - AskUserQuestion
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
  - WebFetch
---

# Enterprise Configuration Schema Management v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-alfred-config-schema |
| **Version** | 4.0.0 Enterprise (2025-11-12) |
| **Standards** | JSON Schema v2024-12, RFC 8174 keywords, TOML/YAML best practices |
| **AI Integration** | ✅ Context7 MCP for official docs |
| **Auto-load** | When config validation or management needed |
| **Environments** | development, staging, production + custom |
| **Lines of Content** | 900+ with 12+ production examples |
| **Progressive Disclosure** | 3-level (quick-start, patterns, advanced) |

---

## What It Does

Provides comprehensive guidance for managing project configuration with JSON Schema validation, environment-specific overrides, secrets management, semantic versioning compliance, and configuration-as-code best practices.

---

## Configuration Hierarchy (3-Layer)

### Layer 1: Base Configuration
**File**: `.moai/config/config.json` (checked into repo)

```json
{
  "project": {
    "name": "moai-adk",
    "version": "0.22.5",
    "description": "SPEC-First TDD Development Kit"
  },
  "language": {
    "primary": "en",
    "conversation_language": "ko",
    "supported": ["en", "ko", "ja"]
  },
  "git_strategy": {
    "use_gitflow": true,
    "main_branch": "main",
    "develop_branch": "develop"
  },
  "document_management": {
    "enabled": true,
    "enforce_structure": true
  }
}
```

**Key Points**:
- Checked into version control
- No secrets, credentials, or API keys
- Environment-agnostic defaults
- Semantic versioning format

### Layer 2: Environment Overrides
**Files**: `.moai/config/.env.{environment}` (NOT checked in)

```env
# .moai/config/.env.production
DATABASE_URL=postgresql://prod-db.example.com/production
API_KEY=sk-prod-xxxxxxxxxxxx
REDIS_URL=redis://prod-cache.example.com:6379
DEBUG=false
LOG_LEVEL=error
```

**Key Points**:
- One file per environment
- Contains secrets and env-specific values
- MUST be in `.gitignore`
- Loaded at runtime based on NODE_ENV/ENVIRONMENT

### Layer 3: Local Development Overrides
**File**: `.moai/config/.env.local` (NOT checked in)

```env
# .moai/config/.env.local (developer-specific)
DATABASE_URL=postgresql://localhost/mydb
API_KEY=sk-dev-xxxxxxxxxxxx
DEBUG=true
LOG_LEVEL=debug
```

**Key Points**:
- Developer-specific settings
- Takes precedence over everything
- NEVER committed to repo
- Used for local testing with real services

---

## JSON Schema v2024-12 Validation

### Base Schema Structure

```json
{
  "$schema": "https://json-schema.org/draft/2024-12/schema",
  "$id": "https://moai-adk.dev/schemas/config-v4.0.0.json",
  "title": "MoAI-ADK Configuration Schema",
  "description": "Configuration schema for MoAI Agentic Development Kit v4.0.0",
  "type": "object",
  "required": ["project", "language", "git_strategy"],
  "properties": {
    "project": {
      "type": "object",
      "required": ["name", "version"],
      "properties": {
        "name": {
          "type": "string",
          "minLength": 1,
          "maxLength": 100,
          "pattern": "^[a-z0-9][a-z0-9-]*[a-z0-9]$",
          "description": "Project name (kebab-case)"
        },
        "version": {
          "type": "string",
          "pattern": "^\\d+\\.\\d+\\.\\d+(-[a-zA-Z0-9]+)*$",
          "description": "Semantic version (e.g., 0.22.5)"
        }
      }
    },
    "language": {
      "type": "object",
      "properties": {
        "conversation_language": {
          "type": "string",
          "enum": ["en", "ko", "ja", "es", "fr", "de", "zh", "pt", "ru"]
        }
      }
    }
  },
  "additionalProperties": false
}
```

---

## Configuration Validation Checklist

### Required Validations

| Check | Standard | Tool |
|-------|----------|------|
| **JSON Schema v2024-12 compliance** | JSON Schema | ajv, jsonschema |
| **No hardcoded secrets** | Security | grep, git-secrets |
| **Semantic versioning format** | RFC 8174 | regex, semver library |
| **Environment variables defined** | Best practice | dotenv-cli, envman |
| **Type correctness** | JSON typing | JSON validator |
| **Required fields present** | Schema | JSON Schema validator |

### Security Validation

```typescript
// ❌ Bad: Hardcoded secrets in config
{
  "database": {
    "password": "super_secret_123"
  },
  "api_key": "sk-1234567890"
}

// ✅ Good: Environment variables
{
  "database": {
    "password": "${DB_PASSWORD}"  // or env var
  },
  "api_key": "${API_KEY}"  // loaded at runtime
}

// ✅ Good: .gitignore protection
.env
.env.local
.env.*.local
.vercel/
config/.env*
```

---

## Environment Management

### Standard Environments

```
development  → Local development with mock services
staging      → Pre-production testing
production   → Live production environment
test         → Automated testing with fixtures
```

### Loading Strategy

```typescript
function loadConfig(environment: string) {
  // 1. Load base config
  const baseConfig = require('.moai/config/config.json');
  
  // 2. Load environment-specific config (.moai/config/.env.{environment})
  const envConfig = loadEnvFile(`.moai/config/.env.${environment}`);
  
  // 3. Load local overrides (.moai/config/.env.local)
  const localConfig = loadEnvFile('.moai/config/.env.local');
  
  // 4. Merge with precedence: local > env > base
  return {
    ...baseConfig,
    ...envConfig,
    ...localConfig
  };
}
```

---

## Secrets Management Best Practices

### DO
- ✅ Store secrets in `.env.{environment}` files
- ✅ Load at runtime via environment variables
- ✅ Use secret management tools (AWS Secrets Manager, Vault, etc.)
- ✅ Rotate secrets regularly
- ✅ Use strong, random values
- ✅ Document secret naming conventions
- ✅ Audit access to secrets
- ✅ Enable secret scanning in CI/CD

### DON'T
- ❌ Commit `.env` files to repo
- ❌ Hardcode API keys, passwords, tokens
- ❌ Share secrets in Slack, email, or tickets
- ❌ Use weak or predictable values
- ❌ Leave old secrets in git history
- ❌ Print secrets in logs
- ❌ Commit credentials to `.gitignore`
- ❌ Use same secret across environments

---

## Git Safety for Configuration

### .gitignore Configuration

```gitignore
# Configuration secrets
.env
.env.local
.env.*.local
.moai/config/.env*

# Deployment platform secrets
.vercel/
.netlify/
.firebase/
.aws/credentials

# IDE secrets
.vscode/settings.json
.idea/workspace.xml

# OS secrets
.DS_Store

# Never commit these patterns
**/secrets.*
**/credentials.*
**/api[_-]?key*
**/password*
**/token*
```

### Pre-Commit Verification

```bash
#!/bin/bash
# Verify no secrets in staged files

echo "Checking for committed secrets..."

# Check for common secret patterns
git diff --cached | grep -i \
  -e "api[_-]?key" \
  -e "secret" \
  -e "password" \
  -e "token" \
  -e ".env" \
  -e "credentials"

if [ $? -eq 0 ]; then
  echo "ERROR: Secrets detected in staged files!"
  exit 1
fi

exit 0
```

---

## Semantic Versioning in Config

**Format**: `MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]`

```
0.22.5              → Stable release
0.22.5-alpha        → Alpha pre-release
0.22.5-beta.1       → Beta release #1
0.22.5+build.123    → Build metadata
0.22.5-rc.1+build.2 → Release candidate with build info
```

**Upgrade Rules**:
- **MAJOR** (0→1): Breaking changes, major rewrites
- **MINOR** (22→23): New features, backward compatible
- **PATCH** (5→6): Bug fixes only
- **Pre-release**: Experimental, not production-ready
- **Build metadata**: Informational only, doesn't affect version precedence

---

## Configuration-as-Code Best Practices

### Pattern 1: Type-Safe Configuration

```typescript
// Type-safe config with validation
interface AppConfig {
  project: {
    name: string;
    version: string;
  };
  language: {
    conversation_language: 'en' | 'ko' | 'ja';
  };
  git_strategy: {
    use_gitflow: boolean;
    main_branch: string;
  };
}

function validateConfig(config: unknown): AppConfig {
  // Validate against schema
  const schema = require('./schemas/config-v4.0.0.json');
  const valid = ajv.validate(schema, config);
  
  if (!valid) {
    throw new Error(`Config validation failed: ${ajv.errorsText()}`);
  }
  
  return config as AppConfig;
}
```

### Pattern 2: Environment-Specific Defaults

```typescript
function getConfig(environment: string): AppConfig {
  const baseConfig = readConfigFile('config.json');
  const envConfig = readEnvFile(`.env.${environment}`);
  const localConfig = readEnvFile('.env.local');
  
  // Merge with environment-specific defaults
  const defaults = {
    development: { debug: true, logLevel: 'debug' },
    staging: { debug: false, logLevel: 'info' },
    production: { debug: false, logLevel: 'error' }
  };
  
  return {
    ...baseConfig,
    ...defaults[environment],
    ...envConfig,
    ...localConfig
  };
}
```

### Pattern 3: Secrets Validation

```typescript
function validateSecrets(config: AppConfig): void {
  const requiredSecrets = [
    'DATABASE_URL',
    'API_KEY',
    'JWT_SECRET'
  ];
  
  for (const secret of requiredSecrets) {
    if (!process.env[secret]) {
      throw new Error(`Missing required secret: ${secret}`);
    }
    
    // Validate secret format
    if (secret === 'API_KEY' && !process.env[secret].startsWith('sk-')) {
      throw new Error(`Invalid API_KEY format`);
    }
  }
}
```

---

## Related Skills

- `moai-alfred-practices` (Best practices patterns)
- `moai-foundation-specs` (Specification management)

---

**For detailed schema reference**: [reference.md](reference.md)  
**For real-world examples**: [examples.md](examples.md)  
**Last Updated**: 2025-11-12  
**Status**: Production Ready (Enterprise v4.0.0)
