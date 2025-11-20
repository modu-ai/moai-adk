# Configuration Management Examples

Practical implementations for enterprise configuration management with validation, secrets handling, and multi-environment support.

---

## TypeScript Configuration with Zod

### Complete Configuration System

```typescript
// config/schema.ts
import { z } from "zod";

/**
 * Application configuration schema with strict validation.
 * Validates environment variables at startup and provides type-safe config object.
 */

// Database configuration schema
const dbConfigSchema = z.object({
  host: z.string().min(1, "DB_HOST is required"),
  port: z.coerce.number().int().positive().default(5432),
  database: z.string().min(1, "DB_NAME is required"),
  user: z.string().min(1, "DB_USER is required"),
  password: z.string().min(1, "DB_PASSWORD is required"),
  ssl: z.coerce.boolean().default(false),
  poolMin: z.coerce.number().int().nonnegative().default(2),
  poolMax: z.coerce.number().int().positive().default(10),
});

// Redis configuration schema
const redisConfigSchema = z.object({
  url: z.string().url("REDIS_URL must be a valid URL"),
  maxRetries: z.coerce.number().int().positive().default(3),
  enableOfflineQueue: z.coerce.boolean().default(true),
});

// Feature flags schema
const featureFlagsSchema = z.object({
  newDashboard: z.coerce.boolean().default(false),
  aiAssistant: z.coerce.boolean().default(false),
  experimentalSearch: z.coerce.boolean().default(false),
});

// Main configuration schema
export const configSchema = z.object({
  env: z.enum(["development", "staging", "production", "test"]),
  port: z.coerce.number().int().positive().default(3000),
  logLevel: z.enum(["debug", "info", "warn", "error"]).default("info"),

  db: dbConfigSchema,
  redis: redisConfigSchema,
  features: featureFlagsSchema,

  // API keys (secrets)
  apiKeys: z.object({
    stripe: z.string().startsWith("sk_", "Invalid Stripe key format"),
    sendgrid: z
      .string()
      .min(32, "SendGrid API key must be at least 32 characters"),
    openai: z
      .string()
      .startsWith("sk-", "Invalid OpenAI key format")
      .optional(),
  }),

  // Rate limiting
  rateLimits: z.object({
    windowMs: z.coerce
      .number()
      .positive()
      .default(15 * 60 * 1000), // 15 minutes
    maxRequests: z.coerce.number().positive().default(100),
  }),
});

export type Config = z.infer<typeof configSchema>;
```

```typescript
// config/index.ts
import dotenv from "dotenv";
import { configSchema, Config } from "./schema";

// Load environment variables from .env file
dotenv.config();

/**
 * Validates and exports application configuration.
 * Application will fail to start if configuration is invalid.
 */
function loadConfig(): Config {
  const rawConfig = {
    env: process.env.NODE_ENV,
    port: process.env.PORT,
    logLevel: process.env.LOG_LEVEL,

    db: {
      host: process.env.DB_HOST,
      port: process.env.DB_PORT,
      database: process.env.DB_NAME,
      user: process.env.DB_USER,
      password: process.env.DB_PASSWORD,
      ssl: process.env.DB_SSL,
      poolMin: process.env.DB_POOL_MIN,
      poolMax: process.env.DB_POOL_MAX,
    },

    redis: {
      url: process.env.REDIS_URL,
      maxRetries: process.env.REDIS_MAX_RETRIES,
      enableOfflineQueue: process.env.REDIS_OFFLINE_QUEUE,
    },

    features: {
      newDashboard: process.env.FEATURE_NEW_DASHBOARD,
      aiAssistant: process.env.FEATURE_AI_ASSISTANT,
      experimentalSearch: process.env.FEATURE_EXPERIMENTAL_SEARCH,
    },

    apiKeys: {
      stripe: process.env.STRIPE_API_KEY,
      sendgrid: process.env.SENDGRID_API_KEY,
      openai: process.env.OPENAI_API_KEY,
    },

    rateLimits: {
      windowMs: process.env.RATE_LIMIT_WINDOW_MS,
      maxRequests: process.env.RATE_LIMIT_MAX_REQUESTS,
    },
  };

  // Validate configuration
  const parsed = configSchema.safeParse(rawConfig);

  if (!parsed.success) {
    console.error("❌ Configuration validation failed:");
    console.error(JSON.stringify(parsed.error.format(), null, 2));
    process.exit(1);
  }

  return parsed.data;
}

export const config = loadConfig();

// Log non-sensitive configuration on startup
console.log("✓ Configuration loaded successfully");
console.log(`  Environment: ${config.env}`);
console.log(`  Port: ${config.port}`);
console.log(`  Log Level: ${config.logLevel}`);
console.log(
  `  Database: ${config.db.host}:${config.db.port}/${config.db.database}`
);
console.log(`  Redis: ${config.redis.url.replace(/\/\/.*@/, "//***@")}`); // Redact credentials
```

### Environment-Specific .env Files

```bash
# .env.development
NODE_ENV=development
PORT=3000
LOG_LEVEL=debug

DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp_dev
DB_USER=devuser
DB_PASSWORD=devpassword
DB_SSL=false

REDIS_URL=redis://localhost:6379

FEATURE_NEW_DASHBOARD=true
FEATURE_AI_ASSISTANT=true
FEATURE_EXPERIMENTAL_SEARCH=false

STRIPE_API_KEY=sk_test_xxxxxxxxxxxxx
SENDGRID_API_KEY=SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

RATE_LIMIT_WINDOW_MS=900000
RATE_LIMIT_MAX_REQUESTS=1000
```

```bash
# .env.production (Example - DO NOT commit to git!)
NODE_ENV=production
PORT=8080
LOG_LEVEL=warn

DB_HOST=prod-db.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_NAME=myapp_prod
DB_USER=produser
DB_PASSWORD=${PROD_DB_PASSWORD}  # Injected by secrets manager
DB_SSL=true
DB_POOL_MIN=5
DB_POOL_MAX=20

REDIS_URL=rediss://prod-redis.cache.amazonaws.com:6380

FEATURE_NEW_DASHBOARD=true
FEATURE_AI_ASSISTANT=false
FEATURE_EXPERIMENTAL_SEARCH=false

STRIPE_API_KEY=${STRIPE_LIVE_KEY}  # Injected by secrets manager
SENDGRID_API_KEY=${SENDGRID_LIVE_KEY}

RATE_LIMIT_WINDOW_MS=900000
RATE_LIMIT_MAX_REQUESTS=100
```

---

## Python Configuration with Pydantic

### Complete Settings Management

```python
# app/config.py
from pydantic import Field, PostgresDsn, RedisDsn, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict
from typing import Literal


class DatabaseSettings(BaseSettings):
    """Database connection settings."""

    host: str = Field(..., description="Database host")
    port: int = Field(5432, description="Database port")
    database: str = Field(..., description="Database name")
    user: str = Field(..., description="Database user")
    password: str = Field(..., description="Database password")
    ssl: bool = Field(False, description="Enable SSL connection")
    pool_min: int = Field(2, ge=1, description="Minimum pool size")
    pool_max: int = Field(10, ge=1, description="Maximum pool size")

    @field_validator('pool_max')
    @classmethod
    def validate_pool_max(cls, v, values):
        """Ensure pool_max >= pool_min."""
        if 'pool_min' in values.data and v < values.data['pool_min']:
            raise ValueError('pool_max must be >= pool_min')
        return v

    @property
    def dsn(self) -> str:
        """Generate PostgreSQL connection string."""
        protocol = "postgresql+asyncpg" if self.ssl else "postgresql"
        return f"{protocol}://{self.user}:{self.password}@{self.host}:{self.port}/{self.database}"


class RedisSettings(BaseSettings):
    """Redis connection settings."""

    url: RedisDsn = Field(..., description="Redis connection URL")
    max_retries: int = Field(3, ge=0, description="Max connection retries")
    timeout: int = Field(5, ge=1, description="Connection timeout in seconds")


class FeatureFlags(BaseSettings):
    """Application feature flags."""

    new_dashboard: bool = Field(False, description="Enable new dashboard UI")
    ai_assistant: bool = Field(False, description="Enable AI assistant")
    experimental_search: bool = Field(False, description="Enable experimental search")


class APIKeys(BaseSettings):
    """External API keys (secrets)."""

    stripe: str = Field(..., min_length=32, description="Stripe API key")
    sendgrid: str = Field(..., min_length=32, description="SendGrid API key")
    openai: str | None = Field(None, min_length=40, description="OpenAI API key")

    @field_validator('stripe')
    @classmethod
    def validate_stripe_key(cls, v):
        """Validate Stripe key format."""
        if not v.startswith('sk_'):
            raise ValueError('Stripe key must start with sk_')
        return v


class Settings(BaseSettings):
    """Main application settings."""

    model_config = SettingsConfigDict(
        env_file='.env',
        env_file_encoding='utf-8',
        case_sensitive=False,
        env_nested_delimiter='__',  # Support DB__HOST format
    )

    # Application settings
    env: Literal['development', 'staging', 'production', 'test'] = Field(
        'development',
        description="Application environment"
    )
    port: int = Field(3000, ge=1, le=65535, description="Application port")
    log_level: Literal['DEBUG', 'INFO', 'WARNING', 'ERROR'] = Field(
        'INFO',
        description="Logging level"
    )
    debug: bool = Field(False, description="Enable debug mode")

    # Nested configurations
    db: DatabaseSettings = Field(default_factory=DatabaseSettings)
    redis: RedisSettings = Field(default_factory=RedisSettings)
    features: FeatureFlags = Field(default_factory=FeatureFlags)
    api_keys: APIKeys = Field(default_factory=APIKeys)

    # Rate limiting
    rate_limit_window_ms: int = Field(900000, gt=0)  # 15 minutes
    rate_limit_max_requests: int = Field(100, gt=0)

    @field_validator('debug')
    @classmethod
    def validate_debug_mode(cls, v, values):
        """Ensure debug is disabled in production."""
        if values.data.get('env') == 'production' and v:
            raise ValueError('Debug mode must be disabled in production')
        return v

    def get_safe_dict(self) -> dict:
        """Return config dict with secrets redacted."""
        config_dict = self.model_dump()

        # Redact secrets
        if 'api_keys' in config_dict:
            for key in config_dict['api_keys']:
                if config_dict['api_keys'][key]:
                    config_dict['api_keys'][key] = '***REDACTED***'

        if 'db' in config_dict and 'password' in config_dict['db']:
            config_dict['db']['password'] = '***REDACTED***'

        return config_dict


# Load and validate settings
try:
    settings = Settings()
    print("✓ Configuration loaded successfully")
    print(f"  Environment: {settings.env}")
    print(f"  Port: {settings.port}")
    print(f"  Database: {settings.db.host}:{settings.db.port}/{settings.db.database}")
except Exception as e:
    print(f"❌ Configuration error: {e}")
    exit(1)
```

### Usage in Application

```python
# app/main.py
from app.config import settings
from sqlalchemy.ext.asyncio import create_async_engine
import redis.asyncio as aioredis
import logging

# Setup logging with configured level
logging.basicConfig(level=getattr(logging, settings.log_level))
logger = logging.getLogger(__name__)


async def init_database():
    """Initialize database connection using validated settings."""
    engine = create_async_engine(
        settings.db.dsn,
        pool_size=settings.db.pool_max,
        max_overflow=0,
        echo=settings.debug,
    )
    return engine


async def init_redis():
    """Initialize Redis connection using validated settings."""
    redis_client = await aioredis.from_url(
        str(settings.redis.url),
        max_connections=10,
        decode_responses=True,
    )
    return redis_client


async def main():
    """Application entry point."""
    logger.info(f"Starting application in {settings.env} mode")

    # Initialize connections
    db_engine = await init_database()
    redis_client = await init_redis()

    # Check feature flags
    if settings.features.ai_assistant:
        logger.info("AI Assistant feature is ENABLED")

    # Use API keys (already validated)
    stripe_key = settings.api_keys.stripe
    logger.info(f"Stripe key configured: {stripe_key[:10]}...")


if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
```

---

## Docker Multi-Environment Configuration

### Base Docker Compose

```yaml
# docker-compose.yml (base configuration)
version: "3.8"

services:
  app:
    image: myapp:latest
    environment:
      NODE_ENV: ${NODE_ENV:-production}
      PORT: ${PORT:-8080}
      LOG_LEVEL: ${LOG_LEVEL:-info}
    depends_on:
      - db
      - redis
    networks:
      - app_network

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ${DB_NAME:-myapp}
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:?DB_PASSWORD is required}
    volumes:
      - db_data:/var/lib/postgresql/data
    networks:
      - app_network

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD:?REDIS_PASSWORD is required}
    volumes:
      - redis_data:/data
    networks:
      - app_network

networks:
  app_network:
    driver: bridge

volumes:
  db_data:
  redis_data:
```

### Development Override

```yaml
# docker-compose.dev.yml
version: "3.8"

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app
      - /app/node_modules # Persist node_modules
    environment:
      NODE_ENV: development
      LOG_LEVEL: debug
      DEBUG: "true"

      # Development database
      DB_HOST: db
      DB_PORT: 5432
      DB_NAME: myapp_dev
      DB_USER: devuser
      DB_PASSWORD: devpassword

      # Development Redis
      REDIS_URL: redis://:devpassword@redis:6379

      # Feature flags (all enabled in dev)
      FEATURE_NEW_DASHBOARD: "true"
      FEATURE_AI_ASSISTANT: "true"

      # Test API keys
      STRIPE_API_KEY: sk_test_xxxxxxxxxxxxx
      SENDGRID_API_KEY: SG.test_xxxxxxxxxxxxx
    ports:
      - "3000:3000"
      - "9229:9229" # Node.js debugger

  db:
    ports:
      - "5432:5432" # Expose for local access
    environment:
      POSTGRES_DB: myapp_dev
      POSTGRES_USER: devuser
      POSTGRES_PASSWORD: devpassword

  redis:
    ports:
      - "6379:6379"
    command: redis-server --requirepass devpassword
```

### Production Override

```yaml
# docker-compose.prod.yml
version: "3.8"

services:
  app:
    image: myapp:${VERSION:-latest}
    restart: always
    environment:
      NODE_ENV: production
      LOG_LEVEL: warn

      # Production DB (from secrets)
      DB_HOST: ${PROD_DB_HOST}
      DB_PORT: ${PROD_DB_PORT}
      DB_NAME: ${PROD_DB_NAME}
      DB_USER: ${PROD_DB_USER}
      DB_PASSWORD: ${PROD_DB_PASSWORD}
      DB_SSL: "true"

      # Production Redis (from secrets)
      REDIS_URL: ${PROD_REDIS_URL}

      # Feature flags (controlled rollout)
      FEATURE_NEW_DASHBOARD: "true"
      FEATURE_AI_ASSISTANT: "false"

      # Production API keys (from secrets manager)
      STRIPE_API_KEY: ${STRIPE_LIVE_KEY}
      SENDGRID_API_KEY: ${SENDGRID_LIVE_KEY}
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: "2"
          memory: 2G
        reservations:
          cpus: "1"
          memory: 1G
```

### Running Different Environments

```bash
# Development
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Production
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Testing
docker-compose -f docker-compose.yml -f docker-compose.test.yml up --abort-on-container-exit
```

---

## Kubernetes ConfigMaps and Secrets

### ConfigMap for Non-Sensitive Configuration

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
data:
  NODE_ENV: "production"
  LOG_LEVEL: "info"
  PORT: "8080"

  # Feature flags
  FEATURE_NEW_DASHBOARD: "true"
  FEATURE_AI_ASSISTANT: "false"

  # Non-sensitive connection details
  DB_HOST: "prod-db.us-east-1.rds.amazonaws.com"
  DB_PORT: "5432"
  DB_NAME: "myapp_prod"
  DB_SSL: "true"

  REDIS_HOST: "prod-redis.cache.amazonaws.com"
  REDIS_PORT: "6380"
```

### Secret for Sensitive Data

```yaml
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: production
type: Opaque
stringData:
  DB_USER: "produser"
  DB_PASSWORD: "super-secret-password"
  REDIS_PASSWORD: "redis-secret-password"
  STRIPE_API_KEY: "sk_live_xxxxxxxxxxxxx"
  SENDGRID_API_KEY: "SG.live_xxxxxxxxxxxxx"
```

### Deployment with ConfigMap and Secret

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
        - name: app
          image: myapp:v1.0.0
          ports:
            - containerPort: 8080

          # Load environment variables from ConfigMap
          envFrom:
            - configMapRef:
                name: app-config
            - secretRef:
                name: app-secrets

          resources:
            limits:
              cpu: "1000m"
              memory: "1Gi"
            requests:
              cpu: "500m"
              memory: "512Mi"

          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
```

---

## Configuration Testing

### TypeScript Configuration Tests

```typescript
// config/__tests__/config.test.ts
import { describe, it, expect, beforeEach } from "vitest";
import { configSchema } from "../schema";

describe("Configuration Validation", () => {
  let validConfig: any;

  beforeEach(() => {
    validConfig = {
      env: "development",
      port: 3000,
      logLevel: "info",
      db: {
        host: "localhost",
        port: 5432,
        database: "testdb",
        user: "testuser",
        password: "testpass",
        ssl: false,
      },
      redis: {
        url: "redis://localhost:6379",
      },
      features: {
        newDashboard: false,
      },
      apiKeys: {
        stripe: "sk_test_xxxxxxxxxxxxxxxx",
        sendgrid: "SG.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      },
      rateLimits: {
        windowMs: 900000,
        maxRequests: 100,
      },
    };
  });

  it("should validate correct configuration", () => {
    const result = configSchema.safeParse(validConfig);
    expect(result.success).toBe(true);
  });

  it("should reject invalid environment", () => {
    validConfig.env = "invalid";
    const result = configSchema.safeParse(validConfig);
    expect(result.success).toBe(false);
  });

  it("should reject invalid Stripe key format", () => {
    validConfig.apiKeys.stripe = "invalid_key";
    const result = configSchema.safeParse(validConfig);
    expect(result.success).toBe(false);
  });

  it("should apply default values", () => {
    delete validConfig.port;
    const result = configSchema.safeParse(validConfig);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.port).toBe(3000);
    }
  });

  it("should coerce string numbers to numbers", () => {
    validConfig.port = "8080";
    const result = configSchema.safeParse(validConfig);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.port).toBe(8080);
      expect(typeof result.data.port).toBe("number");
    }
  });
});
```

---

**See also**: [SKILL.md](./SKILL.md) for configuration principles and secret management best practices
