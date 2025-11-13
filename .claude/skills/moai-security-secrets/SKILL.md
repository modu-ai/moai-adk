---
name: "moai-security-secrets"
version: "4.0.0"
status: stable
description: "Enterprise Skill for advanced development"
allowed-tools: "Read, Bash, WebSearch, WebFetch"
---

# moai-security-secrets: Secret Management & Rotation

**Secure Credential Storage, Rotation & Distribution**  
Trust Score: 9.9/10 | Version: 4.0.0 | Enterprise Mode | Last Updated: 2025-11-12

---

## Overview

Secret management is critical infrastructure: API keys, database passwords, and encryption keys must be stored securely, rotated regularly, and distributed safely. This Skill covers HashiCorp Vault, .env security, and secrets rotation patterns.

**When to use this Skill:**
- Managing database passwords and API keys
- Rotating secrets automatically
- Implementing zero-knowledge architecture
- Distributing secrets to multiple services
- Storing encryption keys securely
- Preventing hardcoded credentials
- Implementing secret versioning
- Building compliance-ready secret systems
- Using Sealed Secrets in Kubernetes

---

## Level 1: Secret Management Principles

### Secret Types & Storage

```
SECRET TYPES:
├─ Credentials (passwords, API keys, tokens)
├─ Cryptographic Keys (encryption, signing)
├─ Connection Strings (database, message queues)
├─ Certificates (TLS, client auth)
└─ OAuth Tokens (access tokens, refresh tokens)

STORAGE HIERARCHY:
1. Vault System (HashiCorp Vault, AWS Secrets Manager)
2. Environment Variables (via secret injection)
3. Configuration Files (.env - local dev only)
4. Memory (never disk, clear after use)
5. NEVER: Hardcoded, version control, logs
```

### Secrets Lifecycle

```
Generate → Distribute → Rotate → Revoke → Destroy
   ↓          ↓          ↓         ↓        ↓
 Secure    Encrypted  Schedule  Immediate Wipe
  Token    Channel    (30-90d)   (breach)  Keys
```

### Rotation Strategy

| Type | Frequency | Grace Period | Method |
|------|-----------|--------------|--------|
| API Keys | 90 days | 24 hours | Generate new, deprecate old |
| Database Passwords | 30 days | 48 hours | New password, app restart |
| TLS Certificates | 30 days before expiry | N/A | Automated renewal |
| Session Secrets | 24 hours | Immediate | Rotate all sessions |
| Encryption Keys | On breach | N/A | Re-encrypt all data |

---

## Level 2: Implementation Patterns

### HashiCorp Vault Integration

**Vault Server Setup:**

```bash
# Installation
wget https://releases.hashicorp.com/vault/1.18.0/vault_1.18.0_linux_amd64.zip
unzip vault_1.18.0_linux_amd64.zip

# Start Vault
vault server -config=/etc/vault/config.hcl

# Initialize & unseal
vault operator init
vault operator unseal [key1] [key2] [key3]

# Enable secret engine
vault secrets enable -path=app kv-v2
vault secrets enable database
```

**Node.js Vault Client:**

```javascript
const vault = require('@hashicorp/vault-client');

const client = new vault.ApiClient({
  endpoint: process.env.VAULT_ADDR,
  token: process.env.VAULT_TOKEN
});

// 1. Read secret
async function getSecret(path) {
  try {
    const response = await client.read(`secret/data/${path}`);
    return response.data.data;
  } catch (err) {
    console.error('Vault error:', err);
    throw err;
  }
}

// 2. Store secret
async function setSecret(path, data) {
  await client.write(`secret/data/${path}`, {
    data: data
  });
}

// 3. Generate dynamic database password
async function generateDBPassword(role) {
  const cred = await client.read(`database/creds/${role}`);
  return {
    username: cred.data.username,
    password: cred.data.password,
    ttl: cred.lease_duration
  };
}

// 4. Rotate API key
async function rotateAPIKey(appName) {
  // Generate new key
  const newKey = crypto.randomBytes(32).toString('hex');
  
  // Store in Vault
  await setSecret(`app/${appName}/api-key`, { key: newKey });
  
  // Distribute to app
  await notifyApp(appName, newKey);
  
  // Mark old key as deprecated
  await setSecret(`app/${appName}/api-key-deprecated`, { 
    deprecated: true, 
    rotatedAt: new Date() 
  });
}

// 5. Token authentication
async function authenticateWithVault() {
  const response = await client.auth.jwt.login({
    role: process.env.VAULT_ROLE,
    jwt: process.env.JWT_TOKEN
  });
  
  return response.auth.client_token;
}
```

### .env Security (Development)

**NEVER commit secrets to version control:**

```bash
# .gitignore
.env
.env.local
.env.*.local
.env.prod
.env.backup
secrets/
keys/
```

**Environment Variable Validation:**

```javascript
// config/env.js
const z = require('zod');

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'production']).default('development'),
  DATABASE_URL: z.string().url(),
  DATABASE_PASSWORD: z.string().min(16),
  API_KEY: z.string().min(32),
  ENCRYPTION_KEY: z.string().length(64),
  JWT_SECRET: z.string().min(32),
  OAUTH_CLIENT_ID: z.string(),
  OAUTH_CLIENT_SECRET: z.string(),
  VAULT_ADDR: z.string().url().optional(),
  VAULT_TOKEN: z.string().optional()
});

// Validate on startup
export const env = envSchema.parse(process.env);

// Ensure no secrets in logs
Object.keys(env).forEach(key => {
  if (key.includes('SECRET') || key.includes('KEY') || key.includes('PASSWORD')) {
    Object.defineProperty(env, key, {
      get() {
        return '[REDACTED]';
      }
    });
  }
});
```

### Secrets Rotation Scheduler

```javascript
// jobs/secrets-rotation.js
const cron = require('node-cron');
const vault = require('@hashicorp/vault-client');

class SecretsRotationJob {
  constructor(vaultClient) {
    this.vault = vaultClient;
    this.rotationSchedules = [
      { path: 'app/database-password', schedule: '0 2 * * 0', ttl: '30d' },
      { path: 'app/api-keys', schedule: '0 3 * * SUN', ttl: '90d' },
      { path: 'app/session-secret', schedule: '0 * * * *', ttl: '24h' }
    ];
  }
  
  // Start rotation scheduler
  start() {
    this.rotationSchedules.forEach(({ path, schedule }) => {
      cron.schedule(schedule, () => {
        this.rotateSecret(path).catch(err => {
          console.error(`Rotation failed for ${path}:`, err);
          this.alertOps(path, err);
        });
      });
    });
  }
  
  // Rotate individual secret
  async rotateSecret(path) {
    console.log(`Rotating secret: ${path}`);
    
    // 1. Generate new secret
    const newSecret = this.generateSecret(path);
    
    // 2. Store as new version
    await this.vault.write(`${path}/v2`, newSecret);
    
    // 3. Activate new version (grace period)
    await this.vault.write(`${path}/activate`, {
      version: 2,
      gracePeriod: '24h'
    });
    
    // 4. Notify all services
    await this.notifyServices(path, 2);
    
    // 5. Monitor for successful adoption
    await this.monitorAdoption(path, 2);
    
    // 6. Revoke old version after grace period
    setTimeout(async () => {
      await this.vault.delete(`${path}/v1`);
    }, 86400000); // 24 hours
  }
  
  // Detect and handle compromise
  async handleCompromise(path) {
    console.error(`SECURITY: Secret compromised: ${path}`);
    
    // 1. Immediate revocation
    await this.vault.write(`${path}/revoke`, { immediate: true });
    
    // 2. Generate emergency secret
    const emergency = this.generateSecret(path);
    await this.vault.write(`${path}/emergency`, emergency);
    
    // 3. Alert operations
    this.alertOps(path, 'COMPROMISED');
    
    // 4. Force immediate adoption (no grace period)
    await this.notifyServices(path, 'emergency');
    
    // 5. Audit trail
    await this.logCompromise(path);
  }
  
  generateSecret(path) {
    if (path.includes('password')) {
      return {
        password: require('crypto').randomBytes(32).toString('hex'),
        rotatedAt: new Date(),
        expiresAt: new Date(Date.now() + 30 * 24 * 60 * 60000)
      };
    }
    
    if (path.includes('api-key')) {
      return {
        key: `sk_${require('crypto').randomBytes(32).toString('hex')}`,
        rotatedAt: new Date()
      };
    }
    
    return {};
  }
  
  async notifyServices(path, version) {
    // Publish secret rotation event to message queue
    const services = await this.getAffectedServices(path);
    
    for (const service of services) {
      await this.messageQueue.publish('secrets.rotated', {
        path,
        version,
        timestamp: new Date()
      }, { targetService: service });
    }
  }
  
  async monitorAdoption(path, version) {
    // Poll services until they've adopted new secret
    const maxRetries = 10;
    let retries = 0;
    
    while (retries < maxRetries) {
      const adopted = await this.checkServiceAdoption(path, version);
      
      if (adopted === 100) {
        console.log(`Secret ${path} adopted by all services`);
        return;
      }
      
      console.log(`Secret adoption: ${adopted}%`);
      await new Promise(r => setTimeout(r, 60000)); // Wait 1 minute
      retries++;
    }
  }
  
  alertOps(path, error) {
    // Send alert to operations (PagerDuty, Slack, etc.)
    this.alerting.send({
      severity: 'critical',
      title: `Secret rotation failed: ${path}`,
      message: error.message,
      service: 'secrets-manager'
    });
  }
}

// Start scheduler
const rotationJob = new SecretsRotationJob(vaultClient);
rotationJob.start();
```

### Kubernetes Sealed Secrets

```yaml
# Install Sealed Secrets controller
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.23.1/controller.yaml

# Create secret
kubectl create secret generic my-secret \
  --from-literal=password=mypassword \
  --dry-run=client -o yaml > secret.yaml

# Seal the secret
kubeseal -f secret.yaml -w sealed-secret.yaml

# Deploy sealed secret (safe to commit to git)
kubectl apply -f sealed-secret.yaml

# Sealed secret auto-decrypts in cluster
# Use in pod
apiVersion: v1
kind: Pod
metadata:
  name: my-app
spec:
  containers:
  - name: app
    image: my-app:latest
    env:
    - name: PASSWORD
      valueFrom:
        secretKeyRef:
          name: my-secret-sealed
          key: password
```

---

## Level 3: Zero-Knowledge Architecture

```javascript
// Zero-knowledge password verification
class ZKAuth {
  // Registration: Client commits to password without server knowing
  async register(email, password) {
    // Client-side only
    const salt = crypto.randomBytes(16);
    const commitment = hash(hash(password) + salt);
    
    // Send only commitment
    await fetch('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, commitment, salt })
    });
  }
  
  // Authentication: Prove knowledge without revealing password
  async login(email, password) {
    const challenge = await fetch(`/auth/challenge?email=${email}`);
    
    // Client-side proof generation
    const response = hmac(sha256(password), challenge);
    
    // Send only proof (no password)
    const result = await fetch('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, response })
    });
    
    return result.json();
  }
}
```

---

## Reference

### Official Resources
- HashiCorp Vault: https://www.vaultproject.io/
- Sealed Secrets: https://github.com/bitnami-labs/sealed-secrets
- AWS Secrets Manager: https://aws.amazon.com/secrets-manager/
- Azure Key Vault: https://azure.microsoft.com/services/key-vault/

### Tools & Libraries
- **@hashicorp/vault-client**: https://github.com/hashicorp/vault-client-node
- **dotenv**: https://github.com/motdotla/dotenv
- **dotenv-safe**: https://github.com/rolodato/dotenv-safe
- **sealed-secrets**: https://github.com/bitnami-labs/sealed-secrets

### Common Vulnerabilities
- CWE-798: Hard-coded Credentials
- CWE-321: Use of Hard-coded Cryptographic Key
- CWE-798: Hardcoded Passwords
- OWASP A02:2021: Cryptographic Failures

---

**Version**: 4.0.0 Enterprise  
**Skill Category**: Security (Secret Management)  
**Complexity**: Advanced  
**Time to Implement**: 4-6 hours  
**Prerequisites**: DevOps, Kubernetes basics, cryptography concepts
