---
name: moai-cc-configuration
description: Enterprise Configuration Management with AI-powered settings architecture
version: 1.0.1
modularized: true
---

## 📊 Skill Metadata

**Name**: moai-cc-configuration
**Domain**: Configuration Management & Secret Management
**Freedom Level**: high
**Target Users**: DevOps engineers, backend developers, infrastructure architects
**Invocation**: Skill("moai-cc-configuration")
**Progressive Disclosure**: SKILL.md (core) → modules/ (detailed guides)
**Last Updated**: 2025-11-23
**Modularized**: true

---

## 🎯 Quick Reference (30 seconds)

**Purpose**: Enterprise configuration management with multi-environment support and secret management.

**Key Capabilities**:
- Environment-based configuration loading (dev/staging/prod)
- HashiCorp Vault integration (secret management)
- Configuration validation with schema checking
- Kubernetes ConfigMap/Secret integration
- Configuration change monitoring & alerts
- Encryption at rest (AES-256)

**Core Tools**:
- HashiCorp Vault (secret storage)
- Kubernetes ConfigMaps (configuration data)
- Zod/JSON Schema (validation)
- Docker Compose (multi-environment setup)

---

## 📚 Core Patterns (5-10 minutes)

### Pattern 1: Environment-Based Configuration Loading

**Key Concept**: Load different configs for development, staging, and production.

**Approach**:
```typescript
// Multi-layer configuration loading
class ConfigurationManager {
  private loadConfiguration() {
    // Priority order: Secrets > Environment > Base config
    const baseConfig = this.loadBaseConfiguration();      // config/{env}.json
    const envConfig = this.loadEnvironmentConfiguration(); // process.env
    const secretConfig = this.loadSecretConfiguration();   // Vault

    // Merge with precedence
    return { ...baseConfig, ...envConfig, ...secretConfig };
  }

  // Load from JSON file per environment
  private loadBaseConfiguration() {
    const configPath = `./config/${process.env.NODE_ENV}.json`;
    return require(configPath); // dev.json, staging.json, prod.json
  }

  // Load from environment variables
  private loadEnvironmentConfiguration() {
    return {
      database: {
        host: process.env.DB_HOST,
        port: parseInt(process.env.DB_PORT || '5432')
      }
    };
  }

  // Load from Vault (secrets)
  private loadSecretConfiguration() {
    // Retrieve JWT secrets, credentials, etc. from Vault
    return this.vaultManager.retrieveSecrets();
  }
}
```

**Use Case**: Separate configs for localhost, staging server, production cluster.

### Pattern 2: Vault Secret Management

**Key Concept**: Centralized secure storage for passwords, API keys, tokens.

**Pattern**:
```typescript
// Vault secret manager with caching
class VaultSecretManager {
  private secretCache = new Map();

  async retrieveSecret(path: string, cacheKey?: string) {
    // Check cache first
    if (this.secretCache.has(cacheKey)) {
      return this.secretCache.get(cacheKey);
    }

    // Fetch from Vault
    const response = await fetch(`${vaultUrl}/v1/secret/data/${path}`, {
      headers: { 'X-Vault-Token': vaultToken }
    });

    const secret = await response.json();

    // Cache result
    if (cacheKey) {
      this.secretCache.set(cacheKey, secret);
    }

    return secret;
  }

  async rotateSecret(path: string, newData: Record<string, any>) {
    // Create new version + invalidate cache
    await this.createSecret(path, { ...oldData, ...newData });
    this.secretCache.delete(path);
  }
}
```

**Benefits**: Secrets never hardcoded, automatic rotation, audit trail.

### Pattern 3: Configuration Validation & Monitoring

**Key Concept**: Validate config at startup and monitor runtime changes.

**Implementation**:
```python
class ConfigurationValidator:
    def validate_configuration(self, config, schema):
        # Schema validation
        schema_errors = self.schema_validator.validate(config, schema)

        # Business rule validation
        business_errors = self._validate_business_rules(config)

        # Security validation
        security_errors = self._validate_security_requirements(config)

        return {
            'is_valid': len(schema_errors) == 0,
            'errors': schema_errors + business_errors + security_errors
        }

    def _validate_business_rules(self, config):
        # Example: min pool size < max pool size
        if config.db.pool.min > config.db.pool.max:
            return [ValidationError(...)]

        # Example: JWT expiration <= 24 hours
        if self._parse_duration(config.auth.jwt_exp) > 24:
            return [ValidationError(...)]
```

**Triggers**: Application startup, configuration update events.

### Pattern 4: Kubernetes Integration

**Key Concept**: ConfigMaps for non-secret config, Secrets for sensitive data.

**Workflow**:
```yaml
# ConfigMap: Non-sensitive configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  log_level: debug
  port: "3000"

---
# Secret: Sensitive data (base64 encoded)
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
data:
  db_password: YWJjMTIz  # base64 encoded
  jwt_secret: c2VjcmV0  # base64 encoded
```

**Pod Usage**: Mount as volumes or env variables.

### Pattern 5: Configuration Change Management

**Key Concept**: Monitor config changes and alert on breaking updates.

**Pattern**:
```python
class ConfigurationMonitor:
    def monitor_changes(self, old_config, new_config):
        # Detect breaking changes
        breaking_changes = self._detect_breaking_changes(old_config, new_config)

        if breaking_changes:
            self.alerting.send_alert(
                severity="high",
                message="Breaking changes detected",
                details=breaking_changes
            )

        # Record metrics
        self.metrics.record("config_changes", count=1)
```

**Examples**: Database connection pool changes, API endpoint changes.

---

## 📖 Advanced Documentation

This Skill uses Progressive Disclosure. For detailed patterns:

- **[modules/vault-integration.md](modules/vault-integration.md)** - Vault setup and usage
- **[modules/configuration-validation.md](modules/configuration-validation.md)** - Validation schemas and rules
- **[modules/environment-management.md](modules/environment-management.md)** - Multi-environment strategies
- **[modules/secret-management.md](modules/secret-management.md)** - Secret rotation and encryption
- **[modules/advanced-patterns.md](modules/advanced-patterns.md)** - Feature flags, config caching
- **[modules/reference.md](modules/reference.md)** - API reference and checklist

---

## 🎯 Configuration Workflow

**Step 1**: Define base configuration (JSON files)
**Step 2**: Override with environment variables
**Step 3**: Load secrets from Vault
**Step 4**: Validate merged configuration
**Step 5**: Monitor configuration changes
**Step 6**: Alert on breaking changes

---

## 🔄 Git Strategy Configuration (3-Mode System)

### Branch Creation: Two-Level Control

The `.moai/config/config.json` includes **branch_creation** settings with two fields:

```json
{
  "git_strategy": {
    "mode": "personal",              // manual | personal | team
    "branch_creation": {
      "prompt_always": true,         // Ask on every SPEC
      "auto_enabled": false          // Enable auto-creation after approval
    }
  }
}
```

**Level 1**: `prompt_always` - Controls whether to ask user on every SPEC creation

**Level 2**: `auto_enabled` - (Personal/Team modes) Enable auto-branch creation after one-time user approval

**Three Execution Paths**:

1. **Maximum Control** (`prompt_always: true, auto_enabled: false`)
   - User chooses on every SPEC

2. **Approval Workflow** (`prompt_always: false, auto_enabled: false`)
   - First SPEC: Offer one-time approval for automation
   - After approval: Auto-create branches (config updated to `auto_enabled: true`)

3. **Full Automation** (`prompt_always: false, auto_enabled: true`)
   - All SPECs: Auto-create branches (no prompts)

---

## 🔗 Integration with Other Skills

**Complementary Skills**:
- Skill("moai-domain-backend") - Backend configuration patterns
- Skill("moai-domain-cloud") - Cloud-specific config (AWS, GCP, Azure)
- Skill("moai-domain-devops") - Deployment configuration
- Skill("moai-security-identity") - Secret management

---

## 📈 Version History

**1.0.1** (2025-11-23)
- 🔄 Refactored with Progressive Disclosure
- ✨ 5 Core Patterns highlighted
- ✨ Modularized advanced content

**1.0.0** (2025-11-22)
- ✨ Vault integration
- ✨ Multi-environment support
- ✨ Configuration validation

---

**Maintained by**: alfred
**Domain**: Configuration Management
**Generated with**: MoAI-ADK Skill Factory
