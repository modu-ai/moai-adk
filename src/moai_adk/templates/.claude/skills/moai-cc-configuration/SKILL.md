---
name: moai-cc-configuration
version: 4.0.0
updated: 2025-11-20
status: stable
description: Enterprise configuration management with Zod/Pydantic validation
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# Configuration Management Expert

**Scalable Application Configuration & Secrets**

> **Focus**: Environment Variables, Validation (Zod/Pydantic), Secrets  
> **Stack**: TypeScript (Zod), Python (Pydantic), Docker, Kubernetes

---

## Overview

Patterns for managing application settings across environments (Dev, Staging, Prod) securely and reliably.

### Core Principles

1.  **12-Factor App**: Store config in the environment.
2.  **Strict Validation**: Fail fast if config is invalid at startup.
3.  **Secret Separation**: Never commit secrets; use Vault/KMS/Secrets Manager.
4.  **Type Safety**: Treat configuration as typed objects, not raw strings.

---

## Implementation Patterns

### 1. TypeScript Configuration (Zod)

Type-safe configuration with runtime validation.

```typescript
import { z } from "zod";
import dotenv from "dotenv";

dotenv.config();

const configSchema = z.object({
  env: z.enum(["development", "production", "test"]),
  port: z.coerce.number().default(3000),
  db: z.object({
    host: z.string(),
    port: z.coerce.number().default(5432),
    user: z.string(),
    password: z.string(), // Treat as secret
    ssl: z.coerce.boolean().default(false),
  }),
  redis: z.object({
    url: z.string().url(),
  }),
  features: z.object({
    newDashboard: z.coerce.boolean().default(false),
  }),
});

// Validate process.env
const parsed = configSchema.safeParse({
  env: process.env.NODE_ENV,
  port: process.env.PORT,
  db: {
    host: process.env.DB_HOST,
    port: process.env.DB_PORT,
    user: process.env.DB_USER,
    password: process.env.DB_PASSWORD,
    ssl: process.env.DB_SSL,
  },
  redis: {
    url: process.env.REDIS_URL,
  },
  features: {
    newDashboard: process.env.FEATURE_NEW_DASHBOARD,
  },
});

if (!parsed.success) {
  console.error("❌ Invalid configuration:", parsed.error.format());
  process.exit(1);
}

export const config = parsed.data;
```

### 2. Python Configuration (Pydantic)

Robust settings management using Pydantic BaseSettings.

```python
from pydantic_settings import BaseSettings
from pydantic import PostgresDsn, RedisDsn, Field

class Settings(BaseSettings):
    environment: str = Field("development", alias="NODE_ENV")
    database_url: PostgresDsn
    redis_url: RedisDsn
    api_key: str = Field(..., min_length=32)
    debug: bool = False

    class Config:
        env_file = ".env"
        case_sensitive = False

# Usage
try:
    settings = Settings()
    print(f"Running in {settings.environment} mode")
except Exception as e:
    print(f"❌ Configuration error: {e}")
    exit(1)
```

### 3. Multi-Environment Docker Compose

Managing overrides for different environments.

**docker-compose.yml (Base)**

```yaml
services:
  app:
    image: myapp
    environment:
      - NODE_ENV=${NODE_ENV:-production}
      - DB_HOST=db
    depends_on:
      - db
```

**docker-compose.dev.yml (Override)**

```yaml
services:
  app:
    build: .
    volumes:
      - .:/app
    environment:
      - DEBUG=true
```

**Command**: `docker-compose -f docker-compose.yml -f docker-compose.dev.yml up`

---

## Secret Management Best Practices

1.  **Development**: Use `.env` files (git-ignored) or local secret managers.
2.  **Production**: Inject secrets as environment variables via platform (AWS Secrets Manager, K8s Secrets, Vault).
3.  **Rotation**: Design apps to handle secret rotation (e.g., reconnect on DB auth failure).
4.  **Logging**: **NEVER** log configuration objects without redaction.

---

## Validation Checklist

- [ ] **Schema**: Is every config variable defined in a schema (Zod/Pydantic)?
- [ ] **Defaults**: Are sensible defaults provided for non-critical values?
- [ ] **Secrets**: Are secrets explicitly marked and handled securely?
- [ ] **Fail-Fast**: Does the app crash immediately on invalid config?
- [ ] **Types**: Is the configuration fully typed in the codebase?

---

## Related Skills

- `moai-devops-docker`: Container configuration
- `moai-devops-k8s`: Kubernetes ConfigMaps & Secrets
- `moai-security-encryption`: Encrypting sensitive config

---

**Last Updated**: 2025-11-20
