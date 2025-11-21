---
name: encryption-architecture-algorithms
parent: moai-security-encryption
description: Cryptographic algorithms and implementation
---

# Module 1: Encryption Architecture & Algorithms

## Modern Encryption Stack (2025)

### Core Algorithms

- **AES-256-GCM**: Symmetric encryption with authentication
- **RSA-4096**: Asymmetric encryption and signatures
- **ECC P-384**: Elliptic curve for efficiency
- **SHA-384**: Cryptographic hashing
- **HMAC-SHA256**: Message authentication

## Advanced Encryption Implementation

```typescript
import crypto from 'crypto';

interface EncryptionConfig {
  algorithm: string;
  keyLength: number;
  ivLength: number;
  tagLength: number;
}

export class AdvancedEncryptionManager {
  private config: EncryptionConfig;

  constructor() {
    this.config = {
      algorithm: 'aes-256-gcm',
      keyLength: 32,
      ivLength: 16,
      tagLength: 16
    };
  }

  async encrypt(plaintext: string, key: Buffer): Promise<EncryptedData> {
    const iv = crypto.randomBytes(this.config.ivLength);
    const cipher = crypto.createCipheriv(
      this.config.algorithm,
      key,
      iv
    );

    let encrypted = cipher.update(plaintext, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    
    const tag = cipher.getAuthTag();

    return {
      encrypted,
      iv: iv.toString('hex'),
      tag: tag.toString('hex'),
      algorithm: this.config.algorithm
    };
  }

  async decrypt(data: EncryptedData, key: Buffer): Promise<string> {
    const decipher = crypto.createDecipheriv(
      data.algorithm,
      key,
      Buffer.from(data.iv, 'hex')
    );
    
    decipher.setAuthTag(Buffer.from(data.tag, 'hex'));

    let decrypted = decipher.update(data.encrypted, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    
    return decrypted;
  }
}
```

## Digital Signature

```typescript
export class DigitalSignature {
  async sign(data: string, privateKey: string): Promise<string> {
    const sign = crypto.createSign('RSA-SHA256');
    sign.update(data);
    return sign.sign(privateKey, 'hex');
  }

  async verify(
    data: string, 
    signature: string, 
    publicKey: string
  ): Promise<boolean> {
    const verify = crypto.createVerify('RSA-SHA256');
    verify.update(data);
    return verify.verify(publicKey, signature, 'hex');
  }
}
```

---

**Reference**: [NIST Cryptographic Standards](https://csrc.nist.gov/)
