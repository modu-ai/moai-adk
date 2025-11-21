# moai-security-encryption: Production Examples

## Example 1: Symmetric Encryption with AES-256-GCM

```javascript
// encryption-service.js
const crypto = require('crypto');

class EncryptionService {
  // Encrypt data with password derivation
  static encrypt(plaintext, password) {
    // 1. Derive key from password
    const salt = crypto.randomBytes(16);
    const key = crypto.pbkdf2Sync(
      password,
      salt,
      100000,  // iterations
      32,      // key length (256 bits)
      'sha256'
    );
    
    // 2. Generate IV
    const iv = crypto.randomBytes(12);
    
    // 3. Encrypt
    const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
    let encrypted = cipher.update(plaintext, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    
    // 4. Get auth tag
    const authTag = cipher.getAuthTag();
    
    // 5. Combine components
    const result = Buffer.concat([salt, iv, authTag, Buffer.from(encrypted, 'hex')]);
    return result.toString('base64');
  }
  
  // Decrypt data
  static decrypt(ciphertext, password) {
    // 1. Parse components
    const buffer = Buffer.from(ciphertext, 'base64');
    const salt = buffer.slice(0, 16);
    const iv = buffer.slice(16, 28);
    const authTag = buffer.slice(28, 44);
    const encrypted = buffer.slice(44).toString('hex');
    
    // 2. Derive same key
    const key = crypto.pbkdf2Sync(
      password,
      salt,
      100000,
      32,
      'sha256'
    );
    
    // 3. Decrypt
    const decipher = crypto.createDecipheriv('aes-256-gcm', key, iv);
    decipher.setAuthTag(authTag);
    
    let decrypted = decipher.update(encrypted, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    
    return decrypted;
  }
}

// Usage
const encrypted = EncryptionService.encrypt(
  'Secret data',
  'my-secure-password'
);

const decrypted = EncryptionService.decrypt(
  encrypted,
  'my-secure-password'
);

console.log(decrypted); // 'Secret data'
```

## Example 2: Argon2id Password Hashing

```javascript
// password-service.js
const argon2 = require('argon2');

class PasswordService {
  // Hash password
  static async hash(password) {
    try {
      const hash = await argon2.hash(password, {
        type: argon2.argon2id,
        memoryCost: 65540,    // 64 MB
        timeCost: 3,          // 3 iterations
        parallelism: 4,       // 4 threads
        saltLength: 16,       // 128-bit salt
        raw: false
      });
      
      return hash;
    } catch (err) {
      console.error('Hashing error:', err);
      throw err;
    }
  }
  
  // Verify password
  static async verify(password, hash) {
    try {
      return await argon2.verify(hash, password);
    } catch (err) {
      return false;
    }
  }
  
  // Check if rehash needed (for migration from bcrypt)
  static needsRehash(hash) {
    // Check if using old algorithm
    return !hash.startsWith('$argon2id$');
  }
}

// Usage
const hash = await PasswordService.hash('user-password');
const isValid = await PasswordService.verify('user-password', hash);
```

## Example 3: HKDF Key Derivation

```javascript
// key-derivation.js
const crypto = require('crypto');

class KeyDerivation {
  // HKDF for deriving multiple keys from master key
  static deriveKeys(masterKey, salt, context, keyCount = 3) {
    // 1. Extract phase
    const extractedKey = crypto
      .createHmac('sha256', salt)
      .update(masterKey)
      .digest();
    
    // 2. Expand phase
    const keys = [];
    let hash = Buffer.alloc(0);
    let info = Buffer.from(context);
    
    for (let i = 1; i <= keyCount; i++) {
      hash = crypto
        .createHmac('sha256', extractedKey)
        .update(Buffer.concat([hash, info, Buffer.from([i])]))
        .digest();
      
      keys.push(hash.slice(0, 32));
    }
    
    return {
      encryptionKey: keys[0],
      authenticationKey: keys[1],
      kdfKey: keys[2]
    };
  }
  
  // Single key derivation
  static deriveKey(password, salt, length = 32) {
    return crypto.pbkdf2Sync(password, salt, 100000, length, 'sha256');
  }
}

// Usage
const salt = crypto.randomBytes(16);
const keys = KeyDerivation.deriveKeys(
  crypto.randomBytes(32),
  salt,
  'app-v1-keys'
);

console.log(keys);
// { encryptionKey, authenticationKey, kdfKey }
```

## Example 4: End-to-End Encryption (User-to-User)

```javascript
// e2e-messaging.js
const crypto = require('crypto');

class E2EMessaging {
  // Generate keypair for user
  static generateKeypair() {
    return crypto.generateKeyPairSync('rsa', {
      modulusLength: 4096,
      publicKeyEncoding: { type: 'spki', format: 'pem' },
      privateKeyEncoding: { type: 'pkcs8', format: 'pem' }
    });
  }
  
  // Encrypt message for recipient
  static encryptMessage(plaintext, recipientPublicKey) {
    const publicKey = crypto.createPublicKey(recipientPublicKey);
    
    const encrypted = crypto.publicEncrypt(
      {
        key: publicKey,
        padding: crypto.constants.RSA_PKCS1_OAEP_PADDING
      },
      Buffer.from(plaintext)
    );
    
    return encrypted.toString('base64');
  }
  
  // Decrypt message with private key
  static decryptMessage(ciphertext, privateKeyPem) {
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    
    const decrypted = crypto.privateDecrypt(
      {
        key: privateKey,
        padding: crypto.constants.RSA_PKCS1_OAEP_PADDING
      },
      Buffer.from(ciphertext, 'base64')
    );
    
    return decrypted.toString();
  }
  
  // Sign message (for authentication)
  static signMessage(message, privateKeyPem) {
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    
    const signature = crypto.sign(
      'sha256',
      Buffer.from(message),
      privateKey
    );
    
    return signature.toString('base64');
  }
  
  // Verify signature
  static verifySignature(message, signature, publicKeyPem) {
    const publicKey = crypto.createPublicKey(publicKeyPem);
    
    return crypto.verify(
      'sha256',
      Buffer.from(message),
      publicKey,
      Buffer.from(signature, 'base64')
    );
  }
}

// Usage: User registration
const user1Keypair = E2EMessaging.generateKeypair();
const user2Keypair = E2EMessaging.generateKeypair();

// Store public keys on server
await db.publicKeys.create({
  userId: 'user1',
  publicKey: user1Keypair.publicKey
});

// Send encrypted message (user1 → user2)
const encrypted = E2EMessaging.encryptMessage(
  'Secret message',
  user2Keypair.publicKey
);

// Receive & decrypt (only user2 can decrypt)
const decrypted = E2EMessaging.decryptMessage(
  encrypted,
  user2Keypair.privateKey  // Never sent to server
);
```

## Example 5: Database Column Encryption

```javascript
// encrypted-model.js
const crypto = require('crypto');

class EncryptedModel {
  constructor(db, tableName, encryptedFields) {
    this.db = db;
    this.tableName = tableName;
    this.encryptedFields = new Set(encryptedFields);
    this.keyVersion = 1;
  }
  
  // Get encryption key
  getKey() {
    // Load from secure key management system
    return Buffer.from(process.env.ENCRYPTION_KEY, 'hex');
  }
  
  // Encrypt field
  encryptField(plaintext) {
    const key = this.getKey();
    const iv = crypto.randomBytes(12);
    
    const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
    let encrypted = cipher.update(plaintext, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    
    const authTag = cipher.getAuthTag();
    
    return JSON.stringify({
      encrypted,
      iv: iv.toString('hex'),
      authTag: authTag.toString('hex'),
      version: this.keyVersion
    });
  }
  
  // Decrypt field
  decryptField(encryptedData) {
    const data = JSON.parse(encryptedData);
    const key = this.getKey();
    
    const decipher = crypto.createDecipheriv(
      'aes-256-gcm',
      key,
      Buffer.from(data.iv, 'hex')
    );
    
    decipher.setAuthTag(Buffer.from(data.authTag, 'hex'));
    
    let decrypted = decipher.update(data.encrypted, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    
    return decrypted;
  }
  
  // Create with encryption
  async create(data) {
    const encrypted = { ...data };
    
    for (const field of this.encryptedFields) {
      if (data[field]) {
        encrypted[field] = this.encryptField(data[field]);
      }
    }
    
    return this.db.insert(this.tableName, encrypted);
  }
  
  // Find and decrypt
  async find(query) {
    const rows = await this.db.select(this.tableName, query);
    
    return rows.map(row => {
      const decrypted = { ...row };
      
      for (const field of this.encryptedFields) {
        if (row[field]) {
          try {
            decrypted[field] = this.decryptField(row[field]);
          } catch (err) {
            console.error(`Failed to decrypt ${field}:`, err);
          }
        }
      }
      
      return decrypted;
    });
  }
}

// Usage
const usersModel = new EncryptedModel(
  db,
  'users',
  ['ssn', 'creditCard', 'bankAccount']
);

await usersModel.create({
  email: 'user@example.com',
  ssn: '123-45-6789',        // Auto-encrypted
  creditCard: '4532xxxxxxxxxx' // Auto-encrypted
});

const users = await usersModel.find({ email: 'user@example.com' });
console.log(users[0].ssn); // '123-45-6789' (auto-decrypted)
```

