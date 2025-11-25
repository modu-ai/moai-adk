# Multi-Environment Configuration Strategies

Complete guide to managing configurations across development, staging, and production.

## Environment Hierarchy

```
Base Configuration (config/base.json)
  ↓
Environment Override (config/{dev|staging|prod}.json)
  ↓
Environment Variables (process.env)
  ↓
Secrets (Vault/AWS Secrets Manager)
```

## Configuration File Structure

```
config/
├── base.json              # Common defaults
├── development.json       # Dev overrides
├── staging.json           # Staging overrides
├── production.json        # Prod overrides
└── .env.example          # Environment variable template
```

### Example base.json

```json
{
  "app": {
    "name": "my-app",
    "version": "1.0.0",
    "port": 3000
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "pool": {
      "min": 2,
      "max": 10
    }
  },
  "logging": {
    "level": "info",
    "format": "json"
  }
}
```

### Example development.json

```json
{
  "app": {
    "port": 3000
  },
  "database": {
    "host": "localhost",
    "database": "myapp_dev",
    "ssl": false
  },
  "logging": {
    "level": "debug"
  }
}
```

### Example production.json

```json
{
  "database": {
    "host": "prod-db.example.com",
    "database": "myapp_prod",
    "ssl": true,
    "pool": {
      "min": 5,
      "max": 50
    }
  },
  "logging": {
    "level": "warn"
  },
  "features": {
    "maintenanceMode": false,
    "enableAnalytics": true
  }
}
```

## Environment-Specific Loading

```typescript
class EnvironmentConfigLoader {
  loadConfiguration(): AppConfig {
    const env = process.env.NODE_ENV || 'development';

    // Load in order of precedence
    const baseConfig = this.loadFile('config/base.json');
    const envConfig = this.loadFile(`config/${env}.json`);
    const envVars = this.loadFromEnvironment();
    const secrets = this.loadFromVault();

    // Merge configurations
    const merged = {
      ...baseConfig,
      ...envConfig,
      ...envVars,
      ...secrets
    };

    // Validate and return
    return configSchema.parse(merged);
  }

  private loadFile(path: string): Record<string, any> {
    try {
      return require(path);
    } catch {
      return {};
    }
  }

  private loadFromEnvironment(): Record<string, any> {
    return {
      app: {
        port: parseInt(process.env.PORT || '3000')
      },
      database: {
        host: process.env.DB_HOST,
        port: parseInt(process.env.DB_PORT || '5432'),
        username: process.env.DB_USER,
        password: process.env.DB_PASSWORD
      }
    };
  }

  private loadFromVault(): Record<string, any> {
    // Load secrets from Vault/AWS Secrets Manager
    return this.vaultClient.getAllSecrets();
  }
}
```

## Docker Compose Multi-Environment

```yaml
version: '3.8'

services:
  app:
    build: .
    environment:
      NODE_ENV: ${NODE_ENV}
      DB_HOST: ${DB_HOST}
      DB_PASSWORD: ${DB_PASSWORD}
    volumes:
      - ./config:/app/config

  # Development with hot reload
  app-dev:
    extends: app
    environment:
      NODE_ENV: development
      LOG_LEVEL: debug
    volumes:
      - .:/app
      - /app/node_modules
    command: npm run dev

  # Production with replicas
  app-prod:
    extends: app
    environment:
      NODE_ENV: production
      LOG_LEVEL: warn
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1'
          memory: 1G
```

## Kubernetes Environment Management

```yaml
# ConfigMap for non-sensitive data
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config-dev
data:
  LOG_LEVEL: debug
  PORT: "3000"

---
# Secret for sensitive data
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets-dev
type: Opaque
stringData:
  DB_PASSWORD: dev-password
  JWT_SECRET: dev-jwt-secret

---
# Deployment using both
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-dev
spec:
  template:
    spec:
      containers:
      - name: app
        envFrom:
        - configMapRef:
            name: app-config-dev
        - secretRef:
            name: app-secrets-dev
```

## Environment-Specific Feature Flags

```python
class FeatureFlags:
    def __init__(self, config):
        self.flags = config.get('features', {})
        self.environment = config.get('nodeEnv')

    def is_enabled(self, feature: str) -> bool:
        if feature not in self.flags:
            return False

        flag_value = self.flags[feature]

        # Allow per-environment overrides
        if isinstance(flag_value, dict):
            return flag_value.get(self.environment, False)

        return flag_value

# Usage
features = FeatureFlags(config)

if features.is_enabled('new_dashboard'):
    # Load new dashboard
    pass
```

## Best Practices

### ✅ DO
- Use environment variable precedence
- Keep secrets separate from config files
- Version control config files (not secrets)
- Use descriptive config file names
- Document required environment variables
- Validate configuration at startup
- Use .env.example for templates

### ❌ DON'T
- Hardcode environment-specific values
- Commit secrets to version control
- Use different config formats per environment
- Override via code (use config only)
- Store sensitive data in JSON files
- Skip validation for any environment

---

**Tools**: Docker, Kubernetes, direnv, dotenv
