---
name: key-management-systems
parent: moai-security-encryption
description: Enterprise key management and rotation
---

# Module 2: Key Management Systems

## Enterprise Key Management

```typescript
export class EnterpriseKeyManager {
  private hsmClient: HSMClient;
  private keyRotation: KeyRotationService;
  private auditLogger: AuditLogger;

  async createEncryptionKey(
    keyId: string,
    algorithm: string = 'AES-256-GCM'
  ): Promise<CreatedKey> {
    this.auditLogger.log('KEY_CREATION_ATTEMPT', { keyId });

    const hsmKey = await this.hsmClient.createKey({
      algorithm,
      keyId,
      extractable: false,
      sensitive: true
    });

    await this.keyRotation.scheduleRotation(keyId, {
      rotationInterval: 90 // days
    });

    this.auditLogger.log('KEY_CREATED', { keyId });

    return {
      keyId,
      hsmKeyId: hsmKey.id,
      algorithm,
      created: new Date(),
      nextRotation: await this.keyRotation.getNextDate(keyId)
    };
  }

  async rotateKey(keyId: string): Promise<KeyRotationResult> {
    const currentKey = await this.hsmClient.getKey(keyId);
    
    const newKeyId = `${keyId}_rotated_${Date.now()}`;
    const newKey = await this.createEncryptionKey(newKeyId);

    await this.keyRotation.deprecateKey(keyId, {
      deprecationDate: new Date(Date.now() + 30 * 86400000),
      replacementKeyId: newKeyId
    });

    this.auditLogger.log('KEY_ROTATED', {
      oldKeyId: keyId,
      newKeyId
    });

    return { oldKeyId: keyId, newKeyId };
  }
}
```

## Key Management with HashiCorp Vault

```python
import hvac

class VaultKeyManager:
    def __init__(self, vault_url: str, token: str):
        self.client = hvac.Client(url=vault_url, token=token)
    
    def create_encryption_key(self, key_name: str):
        self.client.secrets.transit.create_key(key_name)
    
    def encrypt_data(self, key_name: str, plaintext: str):
        return self.client.secrets.transit.encrypt_data(
            name=key_name,
            plaintext=plaintext
        )
    
    def rotate_key(self, key_name: str):
        self.client.secrets.transit.rotate_key(key_name)
```

## Compliance Auditing

```typescript
class ComplianceAuditor {
  async auditEncryptionOperations(
    timeRange: TimeRange
  ): Promise<AuditReport> {
    const entries = await this.auditLog.getEntries(timeRange);
    
    const complianceResults = [];
    for (const rule of this.complianceRules) {
      complianceResults.push(await rule.validate(entries));
    }

    return {
      timeRange,
      totalOperations: entries.length,
      complianceResults,
      violations: this.identifyViolations(entries)
    };
  }
}
```

---

**Reference**: [HashiCorp Vault](https://www.vaultproject.io/)
