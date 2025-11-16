---
name: "moai-cc-configuration"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
description: Enterprise Configuration Management with AI-powered settings architecture, Context7 integration, and intelligent configuration orchestration for scalable applications
keywords: ['configuration', 'settings', 'environment-variables', 'secrets-management', 'context7-integration', 'ai-orchestration', 'production-deployment']
allowed-tools: 
  - Read
  - Bash
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise Configuration Management Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-cc-configuration |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Configuration Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when configuration keywords detected |

---

## What It Does

Enterprise Configuration Management expert with AI-powered settings architecture, Context7 integration, and intelligent configuration orchestration for scalable applications.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Configuration Architecture** using Context7 MCP for latest config patterns
- 📊 **Intelligent Settings Orchestration** with automated environment optimization
- 🚀 **Advanced Secret Management** with AI-driven security and encryption
- 🔗 **Enterprise Config Framework** with zero-configuration deployment integration
- 📈 **Predictive Configuration Analytics** with usage forecasting and optimization

---

## When to Use

**Automatic triggers**:
- Configuration management and environment variable discussions
- Secret management and security implementation planning
- Multi-environment deployment and configuration strategies
- Settings architecture and configuration optimization

**Manual invocation**:
- Designing enterprise configuration systems with optimal patterns
- Implementing comprehensive secret management strategies
- Planning multi-environment configuration deployments
- Optimizing configuration performance and security

---

# Quick Reference (Level 1)

## Configuration Management Stack (November 2025)

### Core Components
- **Environment Variables**: Docker, Kubernetes, CI/CD pipeline integration
- **Configuration Files**: JSON, YAML, TOML, .env formats
- **Secret Management**: HashiCorp Vault, AWS Secrets Manager, Kubernetes Secrets
- **Configuration Validation**: Schema validation, type checking, default values
- **Environment Management**: Development, staging, production configurations

### Popular Solutions
- **Docker Compose**: Multi-container application configuration
- **Kubernetes ConfigMaps**: Configuration data for pods
- **AWS AppConfig**: Managed configuration service
- **Azure App Configuration**: Feature flags and settings
- **HashiCorp Consul**: Service discovery and configuration

### Security Features
- **Encryption at Rest**: AES-256 encryption for sensitive data
- **Access Control**: Role-based access management (RBAC)
- **Audit Logging**: Configuration changes and access tracking
- **Secret Rotation**: Automated secret rotation and renewal
- **Compliance**: SOC2, HIPAA, GDPR compliance features

### Integration Benefits
- **Scalability**: Dynamic configuration without application restarts
- **Security**: Centralized secret management with encryption
- **Reliability**: Configuration validation and rollback capabilities
- **Observability**: Configuration change tracking and monitoring

---

# Core Implementation (Level 2)

## Configuration Architecture Intelligence

```python
# AI-powered configuration architecture optimization with Context7
class ConfigurationArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.config_analyzer = ConfigurationAnalyzer()
        self.security_validator = SecurityValidator()
    
    async def design_optimal_configuration_architecture(self, 
                                                     requirements: ConfigurationRequirements) -> ConfigurationArchitecture:
        """Design optimal configuration architecture using AI analysis."""
        
        # Get latest configuration management documentation via Context7
        config_docs = await self.context7_client.get_library_docs(
            context7_library_id='/configuration-management/docs',
            topic="environment variables secrets management security 2025",
            tokens=3000
        )
        
        security_docs = await self.context7_client.get_library_docs(
            context7_library_id='/security/docs',
            topic="secret management encryption compliance 2025",
            tokens=2000
        )
        
        # Optimize configuration structure
        config_structure = self.config_analyzer.optimize_structure(
            requirements.application_complexity,
            requirements.deployment_environments,
            config_docs
        )
        
        # Validate security requirements
        security_configuration = self.security_validator.configure_security(
            requirements.security_level,
            requirements.compliance_requirements,
            security_docs
        )
        
        return ConfigurationArchitecture(
            configuration_structure=config_structure,
            security_configuration=security_configuration,
            environment_management=self._design_environment_management(requirements),
            deployment_integration=self._integrate_deployment_pipeline(requirements),
            monitoring_setup=self._configure_monitoring(),
            validation_framework=self._setup_validation_framework()
        )
```

## Advanced Configuration Implementation

```typescript
// Enterprise configuration management with TypeScript
interface ConfigurationSchema {
  database: DatabaseConfig;
  redis: RedisConfig;
  auth: AuthConfig;
  features: FeatureFlags;
  logging: LoggingConfig;
  monitoring: MonitoringConfig;
}

interface DatabaseConfig {
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  ssl: boolean;
  connectionPool: {
    min: number;
    max: number;
    idleTimeoutMillis: number;
  };
}

interface AuthConfig {
  jwtSecret: string;
  jwtExpiration: string;
  refreshTokenSecret: string;
  refreshTokenExpiration: string;
  bcryptRounds: number;
}

interface FeatureFlags {
  newUserDashboard: boolean;
  advancedAnalytics: boolean;
  betaFeatures: boolean;
  maintenanceMode: boolean;
}

// Configuration validator with Zod
import { z } from 'zod';

const databaseConfigSchema = z.object({
  host: z.string().min(1),
  port: z.number().int().min(1).max(65535),
  database: z.string().min(1),
  username: z.string().min(1),
  password: z.string().min(8),
  ssl: z.boolean(),
  connectionPool: z.object({
    min: z.number().int().min(0),
    max: z.number().int().min(1),
    idleTimeoutMillis: z.number().int().min(1000),
  }),
});

const configurationSchema = z.object({
  database: databaseConfigSchema,
  redis: redisConfigSchema,
  auth: authConfigSchema,
  features: featureFlagsSchema,
  logging: loggingConfigSchema,
  monitoring: monitoringConfigSchema,
});

type ValidatedConfiguration = z.infer<typeof configurationSchema>;

// Configuration manager class
export class ConfigurationManager {
  private config: ValidatedConfiguration;
  private environment: string;
  private encryptionKey: string;

  constructor(environment: string, encryptionKey: string) {
    this.environment = environment;
    this.encryptionKey = encryptionKey;
    this.config = this.loadConfiguration();
  }

  private loadConfiguration(): ValidatedConfiguration {
    // Load configuration from multiple sources
    const baseConfig = this.loadBaseConfiguration();
    const envConfig = this.loadEnvironmentConfiguration();
    const secretConfig = this.loadSecretConfiguration();
    
    // Merge configurations with precedence
    const mergedConfig = {
      ...baseConfig,
      ...envConfig,
      ...secretConfig,
    };

    // Validate configuration
    const validatedConfig = configurationSchema.parse(mergedConfig);
    
    return validatedConfig;
  }

  private loadBaseConfiguration(): Partial<ConfigurationSchema> {
    try {
      const configPath = `./config/${this.environment}.json`;
      return require(configPath);
    } catch (error) {
      console.warn(`Base configuration not found for ${this.environment}`);
      return {};
    }
  }

  private loadEnvironmentConfiguration(): Partial<ConfigurationSchema> {
    const envConfig: Partial<ConfigurationSchema> = {};

    // Database configuration from environment
    if (process.env.DB_HOST) {
      envConfig.database = {
        host: process.env.DB_HOST,
        port: parseInt(process.env.DB_PORT || '5432'),
        database: process.env.DB_NAME || 'app',
        username: process.env.DB_USER || 'postgres',
        password: process.env.DB_PASSWORD || '',
        ssl: process.env.DB_SSL === 'true',
        connectionPool: {
          min: parseInt(process.env.DB_POOL_MIN || '2'),
          max: parseInt(process.env.DB_POOL_MAX || '10'),
          idleTimeoutMillis: parseInt(process.env.DB_POOL_IDLE || '30000'),
        },
      };
    }

    // Feature flags from environment
    if (process.env.FEATURE_NEW_DASHBOARD) {
      envConfig.features = {
        newUserDashboard: process.env.FEATURE_NEW_DASHBOARD === 'true',
        advancedAnalytics: process.env.FEATURE_ADVANCED_ANALYTICS === 'true',
        betaFeatures: process.env.FEATURE_BETA === 'true',
        maintenanceMode: process.env.MAINTENANCE_MODE === 'true',
      };
    }

    return envConfig;
  }

  private loadSecretConfiguration(): Partial<ConfigurationSchema> {
    // Load secrets from secure storage
    const secretConfig: Partial<ConfigurationSchema> = {};

    try {
      // JWT secrets from encrypted storage
      const jwtSecret = this.decryptSecret('JWT_SECRET');
      const refreshSecret = this.decryptSecret('REFRESH_TOKEN_SECRET');

      if (jwtSecret) {
        secretConfig.auth = {
          jwtSecret,
          jwtExpiration: process.env.JWT_EXPIRATION || '1h',
          refreshTokenSecret: refreshSecret || jwtSecret,
          refreshTokenExpiration: process.env.REFRESH_TOKEN_EXPIRATION || '7d',
          bcryptRounds: parseInt(process.env.BCRYPT_ROUNDS || '12'),
        };
      }
    } catch (error) {
      console.error('Failed to load secret configuration:', error);
    }

    return secretConfig;
  }

  private decryptSecret(secretName: string): string | null {
    // Implement secure decryption logic
    // This would integrate with your secrets management system
    return process.env[secretName] || null;
  }

  public getConfiguration(): ValidatedConfiguration {
    return this.config;
  }

  public getDatabaseConfig(): DatabaseConfig {
    return this.config.database;
  }

  public getAuthConfig(): AuthConfig {
    return this.config.auth;
  }

  public getFeatureFlags(): FeatureFlags {
    return this.config.features;
  }

  public isFeatureEnabled(feature: keyof FeatureFlags): boolean {
    return this.config.features[feature];
  }

  public updateConfiguration(updates: Partial<ConfigurationSchema>): void {
    // Validate updates
    const updatedConfig = { ...this.config, ...updates };
    configurationSchema.parse(updatedConfig);
    
    // Apply updates
    this.config = updatedConfig;
    
    // Notify configuration change
    this.notifyConfigurationChange(updates);
  }

  private notifyConfigurationChange(changes: Partial<ConfigurationSchema>): void {
    // Emit configuration change events
    console.log('Configuration updated:', changes);
  }
}
```

## Environment-Specific Configuration

```yaml
# docker-compose.yml for multi-environment configuration
version: '3.8'

services:
  app:
    build: .
    environment:
      - NODE_ENV=${NODE_ENV:-development}
      - PORT=${PORT:-3000}
      - DB_HOST=${DB_HOST:-db}
      - DB_PORT=${DB_PORT:-5432}
      - DB_NAME=${DB_NAME:-app}
      - DB_USER=${DB_USER:-postgres}
      - DB_PASSWORD=${DB_PASSWORD}
      - REDIS_HOST=${REDIS_HOST:-redis}
      - REDIS_PORT=${REDIS_PORT:-6379}
      - JWT_SECRET=${JWT_SECRET}
      - REFRESH_TOKEN_SECRET=${REFRESH_TOKEN_SECRET}
    ports:
      - "${PORT:-3000}:3000"
    depends_on:
      - db
      - redis
    volumes:
      - ./logs:/app/logs
      - ./config:/app/config

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=${DB_NAME:-app}
      - POSTGRES_USER=${DB_USER:-postgres}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "${DB_PORT:-5432}:5432"

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "${REDIS_PORT:-6379}:6379"

volumes:
  postgres_data:
  redis_data:

# Environment-specific configurations
configs:
  development:
    environment:
      - NODE_ENV=development
      - LOG_LEVEL=debug
      - HOT_RELOAD=true
    volumes:
      - .:/app
      - /app/node_modules

  staging:
    environment:
      - NODE_ENV=staging
      - LOG_LEVEL=info
      - HOT_RELOAD=false
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '1'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M

  production:
    environment:
      - NODE_ENV=production
      - LOG_LEVEL=warn
      - HOT_RELOAD=false
    deploy:
      replicas: 5
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
        window: 120s
```

---

# Advanced Implementation (Level 3)

## Secret Management Integration

```typescript
// Advanced secret management with HashiCorp Vault
export class VaultSecretManager {
  private vaultUrl: string;
  private vaultToken: string;
  private secretCache: Map<string, Secret> = new Map();

  constructor(vaultUrl: string, vaultToken: string) {
    this.vaultUrl = vaultUrl;
    this.vaultToken = vaultToken;
  }

  async retrieveSecret(path: string, cacheKey?: string): Promise<Secret> {
    // Check cache first
    if (cacheKey && this.secretCache.has(cacheKey)) {
      return this.secretCache.get(cacheKey)!;
    }

    try {
      const response = await fetch(`${this.vaultUrl}/v1/secret/data/${path}`, {
        headers: {
          'X-Vault-Token': this.vaultToken,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to retrieve secret: ${response.statusText}`);
      }

      const data = await response.json();
      const secret = {
        data: data.data.data,
        metadata: data.data.metadata,
        retrievedAt: new Date(),
      };

      // Cache the secret
      if (cacheKey) {
        this.secretCache.set(cacheKey, secret);
      }

      return secret;
    } catch (error) {
      console.error(`Error retrieving secret ${path}:`, error);
      throw error;
    }
  }

  async createSecret(path: string, data: Record<string, any>): Promise<void> {
    try {
      const response = await fetch(`${this.vaultUrl}/v1/secret/data/${path}`, {
        method: 'POST',
        headers: {
          'X-Vault-Token': this.vaultToken,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          data: data,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to create secret: ${response.statusText}`);
      }

      // Clear cache for this path
      this.clearCacheByPattern(path);
    } catch (error) {
      console.error(`Error creating secret ${path}:`, error);
      throw error;
    }
  }

  async rotateSecret(path: string, newData: Record<string, any>): Promise<void> {
    try {
      // Retrieve current secret
      const currentSecret = await this.retrieveSecret(path);
      
      // Create new version
      await this.createSecret(path, {
        ...currentSecret.data,
        ...newData,
        rotatedAt: new Date().toISOString(),
      });

      // Invalidate cache
      this.clearCacheByPattern(path);
    } catch (error) {
      console.error(`Error rotating secret ${path}:`, error);
      throw error;
    }
  }

  private clearCacheByPattern(pattern: string): void {
    for (const [key] of this.secretCache) {
      if (key.includes(pattern)) {
        this.secretCache.delete(key);
      }
    }
  }
}

// Kubernetes ConfigMaps and Secrets integration
export class KubernetesConfigManager {
  private namespace: string;

  constructor(namespace: string) {
    this.namespace = namespace;
  }

  async createConfigMap(name: string, data: Record<string, string>): Promise<void> {
    const configMap = {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name,
        namespace: this.namespace,
      },
      data,
    };

    try {
      const response = await fetch('/api/v1/namespaces/default/configmaps', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(configMap),
      });

      if (!response.ok) {
        throw new Error(`Failed to create ConfigMap: ${response.statusText}`);
      }
    } catch (error) {
      console.error(`Error creating ConfigMap ${name}:`, error);
      throw error;
    }
  }

  async createSecret(name: string, data: Record<string, string>): Promise<void> {
    const secret = {
      apiVersion: 'v1',
      kind: 'Secret',
      metadata: {
        name,
        namespace: this.namespace,
      },
      type: 'Opaque',
      data: Object.fromEntries(
        Object.entries(data).map(([key, value]) => [
          key,
          Buffer.from(value).toString('base64'),
        ])
      ),
    };

    try {
      const response = await fetch('/api/v1/namespaces/default/secrets', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(secret),
      });

      if (!response.ok) {
        throw new Error(`Failed to create Secret: ${response.statusText}`);
      }
    } catch (error) {
      console.error(`Error creating Secret ${name}:`, error);
      throw error;
    }
  }
}
```

### Configuration Validation and Monitoring

```python
class ConfigurationValidator:
    def __init__(self):
        self.schema_validator = SchemaValidator()
        self.monitor = ConfigurationMonitor()
    
    def validate_configuration(self, 
                            config: Configuration,
                            schema: ConfigurationSchema) -> ValidationResult:
        """Validate configuration against schema and business rules."""
        
        # Schema validation
        schema_errors = self.schema_validator.validate(config, schema)
        
        # Business rule validation
        business_errors = self._validate_business_rules(config)
        
        # Security validation
        security_errors = self._validate_security_requirements(config)
        
        return ValidationResult(
            is_valid=len(schema_errors) == 0 and len(business_errors) == 0 and len(security_errors) == 0,
            schema_errors=schema_errors,
            business_errors=business_errors,
            security_errors=security_errors,
            warnings=self._generate_warnings(config)
        )
    
    def _validate_business_rules(self, config: Configuration) -> List[ValidationError]:
        """Validate business-specific rules."""
        errors = []
        
        # Database connection pool validation
        if config.database.connectionPool.min > config.database.connectionPool.max:
            errors.append(ValidationError(
                field="database.connectionPool.min",
                message="Minimum pool size cannot be greater than maximum",
                value=config.database.connectionPool.min
            ))
        
        # JWT expiration validation
        jwt_hours = self._parse_duration_to_hours(config.auth.jwtExpiration)
        if jwt_hours > 24:
            errors.append(ValidationError(
                field="auth.jwtExpiration",
                message="JWT expiration should not exceed 24 hours for security",
                value=config.auth.jwtExpiration
            ))
        
        return errors
    
    def _validate_security_requirements(self, config: Configuration) -> List[ValidationError]:
        """Validate security requirements."""
        errors = []
        
        # Password strength validation
        if len(config.auth.jwtSecret) < 32:
            errors.append(ValidationError(
                field="auth.jwtSecret",
                message="JWT secret must be at least 32 characters long",
                value="***"  # Don't log actual secret
            ))
        
        # SSL validation
        if not config.database.ssl:
            errors.append(ValidationError(
                field="database.ssl",
                message="Database SSL should be enabled for production environments",
                value=config.database.ssl
            ))
        
        return errors

class ConfigurationMonitor:
    def __init__(self):
        self.metrics_collector = MetricsCollector()
        self.alerting = AlertingSystem()
    
    def monitor_configuration_changes(self, 
                                   old_config: Configuration,
                                   new_config: Configuration): void:
        """Monitor configuration changes and detect issues."""
        
        # Detect breaking changes
        breaking_changes = self._detect_breaking_changes(old_config, new_config)
        
        # Collect metrics
        metrics = self._collect_change_metrics(old_config, new_config)
        
        # Send alerts if necessary
        if breaking_changes:
            self.alerting.send_alert(
                severity="high",
                message="Breaking configuration changes detected",
                details=breaking_changes
            )
        
        # Record metrics
        self.metrics_collector.record("configuration_changes", metrics)
```

---

# Reference & Integration (Level 4)

## API Reference

### Core Configuration Operations
- `load_configuration(environment, schema)` - Load and validate configuration
- `manage_secrets(path, data)` - Secure secret management
- `validate_configuration(config, rules)` - Configuration validation
- `monitor_configuration_changes()` - Configuration change monitoring
- `deploy_configuration(config, environment)` - Deploy configuration to environment

### Context7 Integration
- `get_latest_config_docs()` - Configuration management via Context7
- `analyze_secret_patterns()` - Secret management patterns via Context7
- `optimize_deployment_config()` - Deployment optimization via Context7

## Best Practices (November 2025)

### DO
- Use environment variables for configuration with proper validation
- Implement comprehensive secret management with encryption
- Validate configuration at startup and runtime
- Use configuration schemas with strong typing
- Implement proper error handling and default values
- Monitor configuration changes and detect breaking changes
- Use different configurations for different environments
- Implement secure secret rotation and renewal

### DON'T
- Hardcode configuration values in application code
- Store secrets in configuration files or version control
- Skip configuration validation and error handling
- Use weak secrets or encryption algorithms
- Ignore security best practices for configuration management
- Forget to implement configuration change monitoring
- Use production configuration in development environments
- Skip backup and recovery planning for configuration

## Works Well With

- `moai-security-api` (Security integration)
- `moai-foundation-trust` (Trust and compliance)
- `moai-domain-devops` (DevOps and deployment)
- `moai-essentials-perf` (Performance optimization)
- `moai-baas-foundation` (BaaS configuration)
- `moai-domain-backend` (Backend configuration)
- `moai-domain-frontend` (Frontend configuration)
- `moai-security-encryption` (Encryption and security)

## Changelog

- **v4.0.0** (2025-11-13): Complete Enterprise v4.0 rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, advanced secret management patterns, and comprehensive validation framework
- **v2.0.0** (2025-11-11): Complete metadata structure, configuration patterns, security integration
- **v1.0.0** (2025-11-11): Initial configuration management foundation

---

**End of Skill** | Updated 2025-11-13

## Configuration Security

### Secret Management
- Enterprise-grade secret encryption with AES-256
- Automated secret rotation and renewal
- Role-based access control for sensitive configuration
- Comprehensive audit logging and compliance reporting

### Environment Security
- Isolated configuration environments
- Secure configuration transmission and storage
- Configuration validation and sanitization
- Compliance with SOC2, HIPAA, GDPR requirements

---

**End of Enterprise Configuration Management Expert v4.0.0**
