# HashiCorp Vault Integration

Complete guide to integrating HashiCorp Vault for centralized secret management.

## Vault Setup

```bash
# Install Vault
brew install vault

# Start Vault development server
vault server -dev

# Set Vault address
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='s.xxxxxxxxxxxxxxxx'  # From server output
```

## Secret Management Patterns

### Pattern 1: Basic Secret Storage

```typescript
class VaultSecretManager {
  async storeSecret(path: string, data: Record<string, any>) {
    const response = await fetch(`${this.vaultUrl}/v1/secret/data/${path}`, {
      method: 'POST',
      headers: { 'X-Vault-Token': this.vaultToken },
      body: JSON.stringify({ data })
    });

    if (!response.ok) throw new Error('Failed to store secret');
    return response.json();
  }

  async retrieveSecret(path: string) {
    const response = await fetch(`${this.vaultUrl}/v1/secret/data/${path}`, {
      headers: { 'X-Vault-Token': this.vaultToken }
    });

    if (!response.ok) throw new Error('Failed to retrieve secret');
    return response.json();
  }

  async deleteSecret(path: string) {
    const response = await fetch(`${this.vaultUrl}/v1/secret/data/${path}`, {
      method: 'DELETE',
      headers: { 'X-Vault-Token': this.vaultToken }
    });

    if (!response.ok) throw new Error('Failed to delete secret');
    return response.json();
  }
}
```

### Pattern 2: Secret Rotation

```typescript
class SecretRotator {
  async rotateSecret(path: string, newSecret: string) {
    // Get current version
    const current = await this.vault.retrieveSecret(path);

    // Create new version with metadata
    await this.vault.storeSecret(path, {
      secret: newSecret,
      rotatedAt: new Date().toISOString(),
      previousVersion: current.data.data.secret
    });

    // Invalidate cache
    this.cache.delete(path);

    // Notify dependent services
    await this.notifyRotation(path);
  }

  private async notifyRotation(path: string) {
    // Send webhook to services using this secret
    await fetch('http://webhook-service/notify', {
      method: 'POST',
      body: JSON.stringify({ path, rotated: true })
    });
  }
}
```

### Pattern 3: Authentication Methods

```bash
# App Role authentication
vault write auth/approle/role/my-app \
  token_ttl=1h \
  token_max_ttl=4h \
  policies="app-policy"

# Get role ID
vault read auth/approle/role/my-app/role-id

# Generate secret ID
vault write -f auth/approle/role/my-app/secret-id
```

```typescript
// Login with App Role
class AppRoleAuth {
  async login(roleId: string, secretId: string) {
    const response = await fetch(`${this.vaultUrl}/v1/auth/approle/login`, {
      method: 'POST',
      body: JSON.stringify({ role_id: roleId, secret_id: secretId })
    });

    const data = await response.json();
    this.vaultToken = data.auth.client_token;
    this.tokenTTL = data.auth.lease_duration;
  }
}
```

## Best Practices

### ✅ DO
- Use App Role for applications (not tokens)
- Implement secret caching to reduce API calls
- Rotate secrets regularly (weekly/monthly)
- Monitor audit logs for access
- Use policies for least privilege
- Enable encryption at rest

### ❌ DON'T
- Hardcode Vault tokens
- Log secret values
- Use root token in production
- Store Vault token in version control
- Disable audit logging
- Use same secret across environments

---

**Vault Documentation**: [Official Docs](https://www.vaultproject.io/docs)
