# Configuration Validation & Schemas

Complete guide to validating configurations with schema checking and business rules.

## Schema Definition with Zod

```typescript
import { z } from 'zod';

// Database configuration schema
const databaseSchema = z.object({
  host: z.string().min(1, 'Host required'),
  port: z.number().int().min(1).max(65535),
  database: z.string().min(1),
  username: z.string().min(1),
  password: z.string().min(8, 'Password must be 8+ chars'),
  ssl: z.boolean().default(false),
  connectionPool: z.object({
    min: z.number().int().min(1),
    max: z.number().int().min(2),
    idleTimeoutMillis: z.number().int().min(1000)
  })
});

// Authentication configuration
const authSchema = z.object({
  jwtSecret: z.string().min(32),
  jwtExpiration: z.string().regex(/^\d+[smhd]$/, 'Invalid duration format'),
  bcryptRounds: z.number().int().min(10).max(15)
});

// Complete application configuration
export const configSchema = z.object({
  nodeEnv: z.enum(['development', 'staging', 'production']),
  database: databaseSchema,
  auth: authSchema,
  features: z.record(z.boolean()),
  logging: z.object({
    level: z.enum(['debug', 'info', 'warn', 'error']),
    format: z.enum(['json', 'text'])
  })
});

export type AppConfig = z.infer<typeof configSchema>;
```

## JSON Schema Alternative

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "database": {
      "type": "object",
      "properties": {
        "host": { "type": "string", "minLength": 1 },
        "port": { "type": "integer", "minimum": 1, "maximum": 65535 },
        "ssl": { "type": "boolean", "default": false }
      },
      "required": ["host", "port"]
    },
    "auth": {
      "type": "object",
      "properties": {
        "jwtSecret": { "type": "string", "minLength": 32 },
        "jwtExpiration": { "type": "string", "pattern": "^\\d+[smhd]$" }
      },
      "required": ["jwtSecret"]
    }
  },
  "required": ["database", "auth"]
}
```

## Custom Validation Logic

```python
class ConfigurationValidator:
    def validate_configuration(self, config):
        # Schema validation
        try:
            schema_validator.validate(config)
        except ValidationError as e:
            return {'valid': False, 'schema_errors': e.errors()}

        # Business rule validation
        business_errors = self._validate_business_rules(config)

        # Security validation
        security_errors = self._validate_security(config)

        all_errors = business_errors + security_errors

        return {
            'valid': len(all_errors) == 0,
            'errors': all_errors
        }

    def _validate_business_rules(self, config):
        errors = []

        # Pool size validation
        if config['database']['pool']['min'] > config['database']['pool']['max']:
            errors.append({
                'field': 'database.pool.min',
                'message': 'Min pool size cannot exceed max',
                'severity': 'error'
            })

        # JWT expiration validation
        jwt_exp_hours = self._parse_duration_to_hours(config['auth']['jwtExpiration'])
        if jwt_exp_hours > 24:
            errors.append({
                'field': 'auth.jwtExpiration',
                'message': 'JWT expiration should not exceed 24 hours',
                'severity': 'warning'
            })

        return errors

    def _validate_security(self, config):
        errors = []

        # JWT secret strength
        if len(config['auth']['jwtSecret']) < 32:
            errors.append({
                'field': 'auth.jwtSecret',
                'message': 'JWT secret must be at least 32 characters',
                'severity': 'error'
            })

        # Production database SSL
        if config['nodeEnv'] == 'production' and not config['database']['ssl']:
            errors.append({
                'field': 'database.ssl',
                'message': 'SSL required for production database',
                'severity': 'error'
            })

        return errors
```

## Configuration Validation at Startup

```typescript
async function loadAndValidateConfig() {
  try {
    // Load configuration
    const rawConfig = loadConfigurationFromSources();

    // Validate against schema
    const validatedConfig = configSchema.parse(rawConfig);

    // Additional validation
    validateBusinessRules(validatedConfig);
    validateSecurityRequirements(validatedConfig);

    console.log('✓ Configuration valid');
    return validatedConfig;

  } catch (error) {
    console.error('✗ Configuration validation failed:', error.message);
    process.exit(1);
  }
}
```

## Best Practices

### ✅ DO
- Validate at application startup
- Use strict schema definitions
- Check security requirements
- Validate business rules
- Log validation errors (not values)
- Provide helpful error messages

### ❌ DON'T
- Skip validation in production
- Log sensitive values
- Use loose type checking
- Ignore security warnings
- Fail silently on errors
- Use default values for secrets

---

**Schema Validation Libraries**: Zod, Joi, Ajv, JSON Schema
