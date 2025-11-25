# Secret Management & Encryption

Complete guide to managing secrets securely with encryption and rotation.

## Secret Storage Patterns

### Pattern 1: Encrypted File Storage

```python
from cryptography.fernet import Fernet
import json
import os

class EncryptedSecretStorage:
    def __init__(self):
        # Get encryption key from environment
        key = os.environ.get('SECRET_ENCRYPTION_KEY')
        self.cipher = Fernet(key)

    def store_secret(self, name: str, value: str) -> None:
        """Store encrypted secret in file."""
        encrypted = self.cipher.encrypt(value.encode())

        secrets = self._load_secrets()
        secrets[name] = encrypted.decode()

        with open('.secrets.json', 'w') as f:
            json.dump(secrets, f)

    def retrieve_secret(self, name: str) -> str:
        """Retrieve and decrypt secret."""
        secrets = self._load_secrets()
        encrypted = secrets[name].encode()
        decrypted = self.cipher.decrypt(encrypted)
        return decrypted.decode()

    def _load_secrets(self) -> dict:
        try:
            with open('.secrets.json', 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            return {}
```

### Pattern 2: Secret Caching

```typescript
class SecretCache {
  private cache = new Map();
  private ttl = 3600000; // 1 hour

  async getSecret(name: string): Promise<string> {
    // Check cache first
    const cached = this.cache.get(name);
    if (cached && Date.now() - cached.timestamp < this.ttl) {
      return cached.value;
    }

    // Fetch from vault
    const value = await this.vaultClient.getSecret(name);

    // Cache result
    this.cache.set(name, {
      value,
      timestamp: Date.now()
    });

    return value;
  }

  invalidate(name: string): void {
    this.cache.delete(name);
  }

  invalidateAll(): void {
    this.cache.clear();
  }
}
```

### Pattern 3: Secret Rotation

```python
class SecretRotator:
    def rotate_secret(self, name: str, old_secret: str, new_secret: str):
        """Rotate secret with versioning."""

        # Store old secret as previous version
        old_metadata = {
            'value': old_secret,
            'rotated_at': datetime.now().isoformat(),
            'version': 1
        }

        # Create new secret version
        new_metadata = {
            'value': new_secret,
            'created_at': datetime.now().isoformat(),
            'version': 2
        }

        # Store in vault with history
        self.vault.store_secret(name, new_metadata)
        self.vault.store_secret(f'{name}/history', old_metadata)

        # Notify services of rotation
        self.notify_consumers(name, new_secret)

    def notify_consumers(self, secret_name: str, new_secret: str):
        """Notify services that secret has been rotated."""
        services = self.registry.get_services_using_secret(secret_name)
        for service in services:
            self._send_notification(service, secret_name, new_secret)
```

## Password Generation & Strength

```python
import secrets
import string

class PasswordGenerator:
    @staticmethod
    def generate_strong_password(length: int = 32) -> str:
        """Generate cryptographically secure password."""
        charset = string.ascii_letters + string.digits + string.punctuation
        password = ''.join(secrets.choice(charset) for _ in range(length))
        return password

    @staticmethod
    def validate_password_strength(password: str) -> tuple:
        """Validate password strength."""
        score = 0
        requirements = []

        if len(password) >= 12:
            score += 1
            requirements.append('✓ Length >= 12')
        else:
            requirements.append('✗ Length < 12')

        if any(c.isupper() for c in password):
            score += 1
            requirements.append('✓ Has uppercase')
        else:
            requirements.append('✗ No uppercase')

        if any(c.islower() for c in password):
            score += 1
            requirements.append('✓ Has lowercase')
        else:
            requirements.append('✗ No lowercase')

        if any(c.isdigit() for c in password):
            score += 1
            requirements.append('✓ Has digits')
        else:
            requirements.append('✗ No digits')

        if any(c in string.punctuation for c in password):
            score += 1
            requirements.append('✓ Has special chars')
        else:
            requirements.append('✗ No special chars')

        is_strong = score >= 4
        return is_strong, requirements
```

## AWS Secrets Manager Integration

```python
import boto3

class AWSSecretsManager:
    def __init__(self, region: str = 'us-east-1'):
        self.client = boto3.client('secretsmanager', region_name=region)

    def create_secret(self, name: str, secret_value: str) -> dict:
        """Create a new secret in AWS Secrets Manager."""
        try:
            response = self.client.create_secret(
                Name=name,
                Description=f'Secret: {name}',
                SecretString=secret_value,
                Tags=[
                    {'Key': 'Environment', 'Value': 'production'},
                    {'Key': 'ManagedBy', 'Value': 'application'}
                ]
            )
            return response
        except Exception as e:
            print(f'Error creating secret: {e}')
            raise

    def get_secret(self, name: str) -> str:
        """Retrieve secret value."""
        try:
            response = self.client.get_secret_value(SecretId=name)
            return response['SecretString']
        except Exception as e:
            print(f'Error retrieving secret: {e}')
            raise

    def rotate_secret(self, name: str, new_secret: str) -> dict:
        """Rotate secret value."""
        return self.client.update_secret(
            SecretId=name,
            SecretString=new_secret
        )
```

## Encryption Best Practices

### ✅ DO
- Use AES-256 for encryption
- Rotate secrets regularly (monthly)
- Use different secrets per environment
- Store encryption keys separately from secrets
- Implement secret audit logging
- Use cryptographically secure random generation
- Encrypt secrets at rest AND in transit

### ❌ DON'T
- Hardcode secrets in code
- Use weak encryption (DES, MD5)
- Share secrets via email/chat
- Log secret values
- Use same secret across environments
- Store encryption key with secrets
- Commit secrets to version control

---

**Tools**: HashiCorp Vault, AWS Secrets Manager, Azure Key Vault, cryptography library
