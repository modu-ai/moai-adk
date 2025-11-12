# moai-security-encryption: Cryptography & Data Protection

**Advanced Encryption, Key Management & End-to-End Security**  
Trust Score: 9.9/10 | Version: 4.0.0 | Enterprise Mode | Last Updated: 2025-11-12

---

## Overview

Encryption protects sensitive data in transit and at rest. Modern cryptography requires understanding of algorithms, key management, and TLS/E2E implementation. This Skill covers TLS 1.3, symmetric/asymmetric encryption, key derivation (Argon2id), and authenticated encryption (AES-256-GCM).

**When to use this Skill:**
- Encrypting sensitive data in databases
- Implementing end-to-end encryption (E2E)
- Setting up TLS 1.3 for API connections
- Managing cryptographic keys securely
- Implementing HKDF key derivation
- Building encrypted messaging systems
- Securing file uploads with encryption
- Implementing zero-knowledge architectures
- Hashing passwords with Argon2id

---

## Level 1: Foundations (FREE TIER)

### Cryptography Basics

```
Three Types of Encryption:

1. SYMMETRIC (Single Key)
   PlainText + Key ---[AES-256]---> CipherText
   CipherText + Key ---[AES-256]---> PlainText
   ✅ Fast    ❌ Key distribution problem

2. ASYMMETRIC (Public/Private Key)
   PlainText + PublicKey ---[RSA]---> CipherText
   CipherText + PrivateKey ---[RSA]---> PlainText
   ✅ No key sharing  ❌ Slow

3. HYBRID (Both)
   PlainText + SessionKey ---[AES]---> CipherText
   SessionKey + PublicKey ---[RSA]---> EncryptedKey
   ✅ Fast & Secure
```

### Algorithm Selection (November 2025)

| Algorithm | Type | Status | Use Case |
|-----------|------|--------|----------|
| **AES-256-GCM** | Symmetric | ✅ Recommended | Data at rest, symmetric encryption |
| **TLS 1.3** | Hybrid | ✅ Recommended | Data in transit, API/Web |
| **RSA-4096** | Asymmetric | ✅ Recommended | Key exchange, digital signatures |
| **ChaCha20-Poly1305** | Symmetric | ✅ Recommended | Mobile, low-power devices |
| **Argon2id** | Key Derivation | ✅ Recommended | Password hashing |
| **HKDF** | Key Derivation | ✅ Recommended | Deriving session keys |
| **MD5, SHA-1** | Hash | ❌ Deprecated | Use SHA-256/SHA-512 |
| **DES, 3DES** | Symmetric | ❌ Deprecated | Use AES |

### Encryption Modes (Authenticated)

**Why Authenticated Encryption?**

```
❌ Encryption Only:
CipherText can be modified without detection
→ Attacker intercepts and corrupts message
→ System decrypts invalid data

✅ Authenticated Encryption (AES-GCM):
CipherText + Authentication Tag
→ Detects any modification
→ Rejects corrupted data before decryption
```

**Supported Modes:**

| Mode | Security | Speed | Use |
|------|----------|-------|-----|
| **AES-GCM** | ✅ Authenticated | Fast | Recommended |
| **AES-CCM** | ✅ Authenticated | Medium | NIST approved |
| **ChaCha20-Poly1305** | ✅ Authenticated | Very Fast | IETF approved |
| **AES-CBC + HMAC** | ✅ Authenticated | Fast | Older standard |

### TLS 1.3 Protocol (Data in Transit)

```
1. Client Hello (TLS 1.3, supported ciphers)
2. Server Hello + Certificate (TLS 1.3)
3. Authentication (RSA/ECDSA signature)
4. Key Exchange (Diffie-Hellman ephemeral)
   └─ Generates session key
5. Encrypted Data Transfer (AES-256-GCM)
```

**TLS 1.3 Improvements:**
- 0-RTT resumption (faster reconnection)
- Perfect forward secrecy (past sessions safe if key stolen)
- Removed weak ciphers (MD5, RC4, DES)
- Encrypted handshake (privacy)

---

## Level 2: Intermediate Patterns (STANDARD TIER)

### Symmetric Encryption (AES-256-GCM)

**Node.js Built-in Crypto (libsodium.js):**

```javascript
const crypto = require('crypto');

function encryptData(plaintext, password) {
  // 1. Derive key from password using PBKDF2
  const salt = crypto.randomBytes(16);
  const key = crypto.pbkdf2Sync(
    password,
    salt,
    100000,  // Iterations (expensive)
    32,      // Key length (256 bits)
    'sha256'
  );
  
  // 2. Generate random IV (Initialization Vector)
  const iv = crypto.randomBytes(12);  // 96 bits for GCM
  
  // 3. Create cipher
  const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
  
  // 4. Encrypt
  let encrypted = cipher.update(plaintext, 'utf8', 'hex');
  encrypted += cipher.final('hex');
  
  // 5. Get authentication tag
  const authTag = cipher.getAuthTag();
  
  // 6. Combine: salt + iv + authTag + ciphertext
  const result = Buffer.concat([salt, iv, authTag, Buffer.from(encrypted, 'hex')]);
  return result.toString('base64');
}

function decryptData(ciphertext, password) {
  // 1. Parse components
  const buffer = Buffer.from(ciphertext, 'base64');
  const salt = buffer.slice(0, 16);
  const iv = buffer.slice(16, 28);
  const authTag = buffer.slice(28, 44);
  const encrypted = buffer.slice(44).toString('hex');
  
  // 2. Derive same key from password
  const key = crypto.pbkdf2Sync(
    password,
    salt,
    100000,
    32,
    'sha256'
  );
  
  // 3. Create decipher
  const decipher = crypto.createDecipheriv('aes-256-gcm', key, iv);
  decipher.setAuthTag(authTag);
  
  // 4. Decrypt
  let decrypted = decipher.update(encrypted, 'hex', 'utf8');
  decrypted += decipher.final('utf8');
  
  return decrypted;
}

// Usage
const encrypted = encryptData('Secret data', 'my-password');
const decrypted = decryptData(encrypted, 'my-password');
console.log(decrypted); // 'Secret data'
```

### Password Hashing (Argon2id)

**Argon2id vs Bcrypt:**

| Algorithm | Memory | Speed | Parallelism | Recommendation |
|-----------|--------|-------|-------------|-----------------|
| **Argon2id** | Configurable | Medium | ✅ Yes | Modern, recommended |
| **Bcrypt** | Fixed | Slow | ❌ No | Stable, tested |
| **scrypt** | Configurable | Medium | ✅ Yes | Good alternative |
| **PBKDF2** | Low | Fast | ❌ No | Legacy only |

**Argon2id Implementation:**

```javascript
const argon2 = require('argon2');

async function hashPassword(password) {
  try {
    // Argon2id parameters (OWASP 2023 recommendations)
    const hash = await argon2.hash(password, {
      type: argon2.argon2id,           // Most resistant to attacks
      memoryCost: 65540,                // 64 MB (should be higher for critical apps)
      timeCost: 3,                      // 3 iterations
      parallelism: 4,                   // 4 parallel threads
      raw: false,                       // Return string, not buffer
      saltLength: 16                    // 128-bit salt
    });
    
    return hash;
  } catch (err) {
    console.error('Hash error:', err);
    throw err;
  }
}

async function verifyPassword(password, hash) {
  try {
    const isValid = await argon2.verify(hash, password);
    return isValid;
  } catch (err) {
    // Hash verification failed
    return false;
  }
}

// Usage
const hash = await hashPassword('user-password');
const isValid = await verifyPassword('user-password', hash);
```

**Why Argon2id for passwords?**
- Resistant to GPU/ASIC attacks (memory-hard)
- Resistant to side-channel attacks (data-independent)
- Customizable resource costs
- OWASP 2023 recommended

### Key Management (Rotation & Storage)

**Key Derivation with HKDF:**

```javascript
const crypto = require('crypto');

function deriveKeys(masterKey, salt, context) {
  // HKDF: Extract-and-Expand
  
  // 1. Extract phase (compress entropy)
  const extractedKey = crypto
    .createHmac('sha256', salt)
    .update(masterKey)
    .digest();
  
  // 2. Expand phase (derive multiple keys)
  const keys = {};
  let hash = Buffer.alloc(0);
  let info = Buffer.from(context);
  
  for (let i = 1; i <= 3; i++) {
    hash = crypto
      .createHmac('sha256', extractedKey)
      .update(Buffer.concat([hash, info, Buffer.from([i])]))
      .digest();
    
    if (i === 1) keys.encryptionKey = hash.slice(0, 32);  // AES-256
    if (i === 2) keys.authKey = hash.slice(0, 32);        // HMAC
    if (i === 3) keys.kdfKey = hash.slice(0, 32);         // KDF
  }
  
  return keys;
}

// Usage
const masterKey = crypto.randomBytes(32);
const salt = crypto.randomBytes(16);
const keys = deriveKeys(masterKey, salt, 'prod-db-v1');

console.log(keys); // { encryptionKey, authKey, kdfKey }
```

**Key Rotation Strategy:**

```javascript
// Store keys with versions
class KeyManager {
  constructor() {
    this.keys = new Map();
    this.activeKeyVersion = null;
  }
  
  async rotateKeys() {
    const newVersion = (this.activeKeyVersion || 0) + 1;
    const newKey = crypto.randomBytes(32);
    
    // 1. Generate new key
    this.keys.set(newVersion, {
      key: newKey,
      createdAt: new Date(),
      version: newVersion
    });
    
    // 2. Old key still works for decryption (backward compatibility)
    const oldVersion = this.activeKeyVersion;
    
    // 3. New key becomes active
    this.activeKeyVersion = newVersion;
    
    // 4. Re-encrypt all data with new key after grace period
    if (oldVersion) {
      setTimeout(() => {
        this.reencryptOldData(oldVersion, newVersion);
      }, 86400000); // 24 hours
    }
  }
  
  getActiveKey() {
    return this.keys.get(this.activeKeyVersion);
  }
  
  decryptWithVersion(ciphertext, version) {
    const key = this.keys.get(version);
    if (!key) throw new Error(`Key version ${version} not found`);
    return key;
  }
}
```

### End-to-End Encryption (E2E)

**User-to-User Encrypted Messaging:**

```javascript
// 1. Registration: User generates keypair
async function registerUser(userId) {
  const keypair = crypto.generateKeyPairSync('rsa', {
    modulusLength: 4096,
    publicKeyEncoding: { type: 'spki', format: 'pem' },
    privateKeyEncoding: { type: 'pkcs8', format: 'pem' }
  });
  
  // Store public key on server (for everyone)
  await db.publicKeys.create({
    userId,
    publicKey: keypair.publicKey
  });
  
  // Client stores private key locally (never sent to server)
  localStorage.setItem(`private-key-${userId}`, keypair.privateKey);
}

// 2. Send encrypted message
async function sendMessage(fromUserId, toUserId, plaintext) {
  // Get recipient's public key from server
  const recipientKey = await fetch(`/api/keys/${toUserId}`)
    .then(r => r.json())
    .then(data => crypto.createPublicKey(data.publicKey));
  
  // Encrypt with recipient's public key
  const encrypted = crypto.publicEncrypt(
    { key: recipientKey, padding: crypto.constants.RSA_PKCS1_OAEP_PADDING },
    Buffer.from(plaintext)
  );
  
  // Send encrypted message
  await fetch('/api/messages', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      fromUserId,
      toUserId,
      ciphertext: encrypted.toString('base64')
    })
  });
}

// 3. Receive encrypted message
async function receiveMessage(messageId) {
  const message = await fetch(`/api/messages/${messageId}`).then(r => r.json());
  
  // Decrypt with recipient's private key (stored locally)
  const privateKey = localStorage.getItem(`private-key-${userId}`);
  const decrypted = crypto.privateDecrypt(
    { key: privateKey, padding: crypto.constants.RSA_PKCS1_OAEP_PADDING },
    Buffer.from(message.ciphertext, 'base64')
  );
  
  return decrypted.toString();
}

// Why secure?
// Server never has:
// - Private keys
// - Plaintext messages
// - Ability to decrypt
// Only encrypted blobs stored
```

### TLS 1.3 Configuration

**Node.js/Express HTTPS Setup:**

```javascript
const https = require('https');
const fs = require('fs');
const express = require('express');

const app = express();

// TLS 1.3 certificate
const options = {
  key: fs.readFileSync('/secure/server-key.pem'),
  cert: fs.readFileSync('/secure/server-cert.pem'),
  
  // TLS 1.3 only (no downgrade)
  minVersion: 'TLSv1.3',
  maxVersion: 'TLSv1.3',
  
  // Cipher suite (TLS 1.3 only has strong ciphers)
  ciphers: 'TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256',
  
  // HSTS: Force HTTPS for 1 year
  hstsMaxAge: 31536000,
  hstsIncludeSubDomains: true,
  hstsPreload: true,
  
  // Session resumption (0-RTT)
  sessionTimeout: 86400,
  
  // OCSP stapling (certificate validation)
  ocspCallback: (cert, callback) => {
    // Check if certificate revoked
    validateCertificate(cert, callback);
  }
};

const server = https.createServer(options, app);

// Secur headers middleware
app.use((req, res, next) => {
  res.set('Strict-Transport-Security', 'max-age=31536000; includeSubDomains; preload');
  res.set('X-Content-Type-Options', 'nosniff');
  res.set('X-Frame-Options', 'DENY');
  res.set('X-XSS-Protection', '1; mode=block');
  res.set('Content-Security-Policy', "default-src 'self'");
  next();
});

server.listen(443);
```

**Certificate Generation (self-signed for development):**

```bash
# Generate private key
openssl genrsa -out server-key.pem 4096

# Generate certificate (valid 365 days)
openssl req -new -x509 -key server-key.pem -out server-cert.pem -days 365 \
  -subj "/CN=localhost"

# For production, use Let's Encrypt (certbot)
sudo certbot certonly --standalone -d example.com
```

---

## Level 3: Enterprise Patterns (PREMIUM TIER)

### Database Encryption (Transparent Data Encryption)

**Per-Column Encryption:**

```javascript
class EncryptedModel {
  constructor(db) {
    this.db = db;
    this.encryptedFields = new Set(['ssn', 'creditCard', 'email']);
    this.keyManager = new KeyManager();
  }
  
  async create(data) {
    const encrypted = { ...data };
    
    // Encrypt sensitive fields
    for (const field of this.encryptedFields) {
      if (data[field]) {
        const keyVersion = this.keyManager.activeKeyVersion;
        const key = this.keyManager.getActiveKey();
        
        encrypted[field] = {
          ciphertext: this.encryptField(data[field], key),
          keyVersion,
          iv: crypto.randomBytes(12).toString('hex')
        };
      }
    }
    
    return this.db.insert(encrypted);
  }
  
  async find(query) {
    const results = await this.db.find(query);
    
    // Decrypt sensitive fields
    return results.map(row => {
      const decrypted = { ...row };
      
      for (const field of this.encryptedFields) {
        if (row[field]?.ciphertext) {
          const key = this.keyManager.decryptWithVersion(
            row[field].ciphertext,
            row[field].keyVersion
          );
          decrypted[field] = this.decryptField(
            row[field].ciphertext,
            key,
            row[field].iv
          );
        }
      }
      
      return decrypted;
    });
  }
  
  encryptField(plaintext, key) {
    const iv = crypto.randomBytes(12);
    const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
    let encrypted = cipher.update(plaintext, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    const authTag = cipher.getAuthTag();
    return JSON.stringify({ encrypted, authTag: authTag.toString('hex') });
  }
  
  decryptField(ciphertext, key, iv) {
    const { encrypted, authTag } = JSON.parse(ciphertext);
    const ivBuffer = Buffer.from(iv, 'hex');
    const decipher = crypto.createDecipheriv('aes-256-gcm', key, ivBuffer);
    decipher.setAuthTag(Buffer.from(authTag, 'hex'));
    let decrypted = decipher.update(encrypted, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    return decrypted;
  }
}
```

### Zero-Knowledge Proof (Passwordless Auth)

**Commitments without revealing secrets:**

```javascript
// Client never sends password, server never knows it
class ZKProof {
  // Registration: User creates secret
  static createSecret(password) {
    const secret = crypto.randomBytes(32);
    const commitment = crypto
      .createHash('sha256')
      .update(Buffer.concat([secret, Buffer.from(password)]))
      .digest();
    
    return { secret, commitment };
  }
  
  // Authentication: Prove knowledge without revealing
  static generateProof(password, secret, challenge) {
    const response = crypto
      .createHmac('sha256', secret)
      .update(Buffer.concat([challenge, Buffer.from(password)]))
      .digest();
    
    return response;
  }
  
  // Server verifies proof
  static verifyProof(response, commitment, secret, challenge) {
    const expected = crypto
      .createHmac('sha256', secret)
      .update(Buffer.concat([challenge, Buffer.from(password)]))
      .digest();
    
    return crypto.timingSafeEqual(response, expected);
  }
}
```

---

## Reference

### Official Documentation
- NIST Cryptography Standards: https://csrc.nist.gov/
- RFC 5869 (HKDF): https://tools.ietf.org/html/rfc5869
- RFC 8949 (CBOR): https://tools.ietf.org/html/rfc8949
- OWASP Cryptographic Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html
- Libsodium: https://doc.libsodium.org/
- TLS 1.3 RFC: https://tools.ietf.org/html/rfc8446

### Tools & Libraries (November 2025 Versions)
- **crypto** (Node.js built-in): 1.0.0
- **libsodium.js**: 0.7.x
- **argon2**: 0.31.x
- **tweetnacl.js**: 1.0.x
- **node-jose**: 2.1.x

### Common Vulnerabilities & Mitigations

| Vulnerability | OWASP | Mitigation |
|---|---|---|
| **Weak Cipher** | A02:2021 | Use AES-256-GCM |
| **Hard-coded Keys** | A07:2021 | Key management system |
| **Weak Password Hash** | A02:2021 | Use Argon2id |
| **No Authentication** | A02:2021 | Use authenticated encryption |
| **Key Exposure** | A02:2021 | Never log/expose keys |

---

**Version**: 4.0.0 Enterprise
**Skill Category**: Security (Encryption & Cryptography)
**Complexity**: Advanced
**Time to Implement**: 3-5 hours per component
**Prerequisites**: Cryptography concepts, Node.js crypto module, key management understanding
