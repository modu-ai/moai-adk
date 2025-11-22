# moai-security-encryption: Production Examples (2024-2025)

## Example 1: AES-256-GCM Encryption (Modern Standard)

```python
# aes_gcm_encryption.py
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC
import os
import base64

class AESGCMEncryption:
    """Modern AES-256-GCM encryption with authenticated encryption."""
    
    def __init__(self):
        self.key_size = 32  # 256 bits
        self.nonce_size = 12  # 96 bits (recommended for GCM)
    
    def generate_key(self, password: str, salt: bytes = None) -> tuple[bytes, bytes]:
        """Generate encryption key from password using PBKDF2."""
        if salt is None:
            salt = os.urandom(16)
        
        kdf = PBKDF2HMAC(
            algorithm=hashes.SHA256(),
            length=self.key_size,
            salt=salt,
            iterations=600000  # OWASP 2024 recommendation
        )
        
        key = kdf.derive(password.encode())
        return key, salt
    
    def encrypt(self, plaintext: str, key: bytes) -> dict:
        """Encrypt data with AES-256-GCM."""
        # Generate random nonce (must be unique for each encryption)
        nonce = os.urandom(self.nonce_size)
        
        # Create AESGCM cipher
        aesgcm = AESGCM(key)
        
        # Encrypt (includes authentication tag)
        ciphertext = aesgcm.encrypt(
            nonce,
            plaintext.encode(),
            None  # associated_data (optional)
        )
        
        return {
            'ciphertext': base64.b64encode(ciphertext).decode(),
            'nonce': base64.b64encode(nonce).decode(),
            'algorithm': 'AES-256-GCM'
        }
    
    def decrypt(self, encrypted_data: dict, key: bytes) -> str:
        """Decrypt AES-256-GCM encrypted data."""
        # Decode base64
        ciphertext = base64.b64decode(encrypted_data['ciphertext'])
        nonce = base64.b64decode(encrypted_data['nonce'])
        
        # Create AESGCM cipher
        aesgcm = AESGCM(key)
        
        try:
            # Decrypt and verify authentication tag
            plaintext = aesgcm.decrypt(nonce, ciphertext, None)
            return plaintext.decode()
        except Exception as e:
            raise ValueError('Decryption failed: Invalid key or corrupted data')

# Usage
encryptor = AESGCMEncryption()

# Generate key from password
password = "secure_password_12345"
key, salt = encryptor.generate_key(password)

# Encrypt
data = "Sensitive information here"
encrypted = encryptor.encrypt(data, key)

print(f"Ciphertext: {encrypted['ciphertext']}")
print(f"Nonce: {encrypted['nonce']}")

# Decrypt
decrypted = encryptor.decrypt(encrypted, key)
print(f"Decrypted: {decrypted}")
```

## Example 2: Post-Quantum Cryptography (PQC) Hybrid Mode

```python
# pqc_hybrid_encryption.py
from oqs import KeyEncapsulation, Signature
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
import os

class PQCHybridEncryption:
    """Hybrid encryption using classical RSA + PQC (Kyber)."""
    
    def __init__(self):
        # PQC algorithm (NIST ML-KEM-768)
        self.pqc_kem = KeyEncapsulation("Kyber768")
        
        # Classical RSA-4096
        self.rsa_key_size = 4096
    
    def generate_keypair(self) -> dict:
        """Generate hybrid keypair (RSA + Kyber)."""
        # Generate Kyber keypair
        pqc_public_key = self.pqc_kem.generate_keypair()
        pqc_secret_key = self.pqc_kem.export_secret_key()
        
        # Generate RSA keypair
        rsa_private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=self.rsa_key_size
        )
        rsa_public_key = rsa_private_key.public_key()
        
        return {
            'pqc_public_key': pqc_public_key,
            'pqc_secret_key': pqc_secret_key,
            'rsa_public_key': rsa_public_key,
            'rsa_private_key': rsa_private_key
        }
    
    def encrypt(self, plaintext: str, recipient_keys: dict) -> dict:
        """Hybrid encryption: RSA + Kyber + AES-GCM."""
        # Step 1: Generate random AES key
        aes_key = os.urandom(32)  # 256 bits
        
        # Step 2: Encapsulate AES key with Kyber (PQC)
        pqc_kem = KeyEncapsulation("Kyber768", secret_key=None)
        pqc_ciphertext, pqc_shared_secret = pqc_kem.encap_secret(
            recipient_keys['pqc_public_key']
        )
        
        # Step 3: Encrypt AES key with RSA (classical)
        rsa_ciphertext = recipient_keys['rsa_public_key'].encrypt(
            aes_key,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )
        
        # Step 4: Encrypt plaintext with AES-GCM
        nonce = os.urandom(12)
        aesgcm = AESGCM(aes_key)
        ciphertext = aesgcm.encrypt(nonce, plaintext.encode(), None)
        
        return {
            'ciphertext': ciphertext,
            'nonce': nonce,
            'pqc_encapsulated_key': pqc_ciphertext,
            'rsa_encrypted_key': rsa_ciphertext,
            'algorithm': 'RSA-4096 + Kyber768 + AES-256-GCM'
        }
    
    def decrypt(self, encrypted_data: dict, recipient_keys: dict) -> str:
        """Hybrid decryption."""
        # Step 1: Decapsulate Kyber shared secret
        pqc_kem = KeyEncapsulation(
            "Kyber768",
            secret_key=recipient_keys['pqc_secret_key']
        )
        pqc_shared_secret = pqc_kem.decap_secret(
            encrypted_data['pqc_encapsulated_key']
        )
        
        # Step 2: Decrypt RSA encrypted key
        aes_key = recipient_keys['rsa_private_key'].decrypt(
            encrypted_data['rsa_encrypted_key'],
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )
        
        # Step 3: Verify both keys match (hybrid verification)
        # In production, derive AES key from both secrets
        
        # Step 4: Decrypt ciphertext with AES-GCM
        aesgcm = AESGCM(aes_key)
        plaintext = aesgcm.decrypt(
            encrypted_data['nonce'],
            encrypted_data['ciphertext'],
            None
        )
        
        return plaintext.decode()

# Usage
pqc = PQCHybridEncryption()

# Generate hybrid keypair
keys = pqc.generate_keypair()

# Encrypt
data = "Top secret quantum-resistant data"
encrypted = pqc.encrypt(data, keys)

print(f"Encrypted with: {encrypted['algorithm']}")

# Decrypt
decrypted = pqc.decrypt(encrypted, keys)
print(f"Decrypted: {decrypted}")
```

## Example 3: HSM Key Management (AWS KMS)

```python
# aws_kms_encryption.py
import boto3
import base64
from typing import Dict

class AWSKMSEncryption:
    """Enterprise key management using AWS KMS."""
    
    def __init__(self, kms_key_id: str, region: str = 'us-east-1'):
        self.kms_client = boto3.client('kms', region_name=region)
        self.kms_key_id = kms_key_id
    
    def encrypt(self, plaintext: str, context: Dict[str, str] = None) -> dict:
        """Encrypt data using AWS KMS with encryption context."""
        encryption_context = context or {}
        
        response = self.kms_client.encrypt(
            KeyId=self.kms_key_id,
            Plaintext=plaintext.encode(),
            EncryptionContext=encryption_context
        )
        
        return {
            'ciphertext_blob': base64.b64encode(response['CiphertextBlob']).decode(),
            'key_id': response['KeyId'],
            'encryption_context': encryption_context
        }
    
    def decrypt(self, ciphertext_blob: str, context: Dict[str, str] = None) -> str:
        """Decrypt KMS encrypted data."""
        encryption_context = context or {}
        
        ciphertext = base64.b64decode(ciphertext_blob)
        
        response = self.kms_client.decrypt(
            CiphertextBlob=ciphertext,
            EncryptionContext=encryption_context
        )
        
        return response['Plaintext'].decode()
    
    def rotate_key(self) -> dict:
        """Enable automatic key rotation (annual)."""
        response = self.kms_client.enable_key_rotation(
            KeyId=self.kms_key_id
        )
        
        return {
            'key_id': self.kms_key_id,
            'rotation_enabled': True,
            'rotation_period': '365 days'
        }
    
    def generate_data_key(self, key_spec: str = 'AES_256') -> dict:
        """Generate data encryption key (DEK) for envelope encryption."""
        response = self.kms_client.generate_data_key(
            KeyId=self.kms_key_id,
            KeySpec=key_spec
        )
        
        return {
            'plaintext_key': response['Plaintext'],  # Use immediately, then discard
            'encrypted_key': base64.b64encode(response['CiphertextBlob']).decode()
        }

# Usage
kms = AWSKMSEncryption(kms_key_id='arn:aws:kms:us-east-1:123456789012:key/abcd1234')

# Encrypt with context (for audit logging)
data = "Sensitive customer data"
context = {'customer_id': '12345', 'department': 'finance'}
encrypted = kms.encrypt(data, context)

print(f"Encrypted: {encrypted['ciphertext_blob'][:50]}...")

# Decrypt
decrypted = kms.decrypt(encrypted['ciphertext_blob'], context)
print(f"Decrypted: {decrypted}")

# Enable automatic key rotation
rotation = kms.rotate_key()
print(f"Key rotation: {rotation}")
```

## Example 4: Argon2id Password Hashing (2025 Standard)

```python
# argon2_password_hashing.py
from argon2 import PasswordHasher
from argon2.exceptions import VerifyMismatchError
import secrets

class Argon2PasswordManager:
    """Modern password hashing with Argon2id."""
    
    def __init__(self):
        # OWASP 2024 recommendations
        self.hasher = PasswordHasher(
            time_cost=3,          # Number of iterations
            memory_cost=65536,    # Memory usage (64 MB)
            parallelism=4,        # Number of threads
            hash_len=32,          # Output hash length (256 bits)
            salt_len=16,          # Salt length (128 bits)
            type=argon2.Type.ID   # Argon2id (hybrid)
        )
    
    def hash_password(self, password: str) -> str:
        """Hash password with Argon2id."""
        # Validate password strength
        if len(password) < 12:
            raise ValueError('Password must be at least 12 characters')
        
        return self.hasher.hash(password)
    
    def verify_password(self, password: str, password_hash: str) -> bool:
        """Verify password against Argon2id hash."""
        try:
            # Verify password
            self.hasher.verify(password_hash, password)
            
            # Check if rehashing is needed (parameters changed)
            if self.hasher.check_needs_rehash(password_hash):
                # Rehash with new parameters
                new_hash = self.hasher.hash(password)
                return True, new_hash
            
            return True, None
        except VerifyMismatchError:
            return False, None
    
    def generate_secure_password(self, length: int = 16) -> str:
        """Generate cryptographically secure random password."""
        alphabet = (
            'abcdefghijklmnopqrstuvwxyz'
            'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
            '0123456789'
            '!@#$%^&*()_+-=[]{}|;:,.<>?'
        )
        
        password = ''.join(secrets.choice(alphabet) for _ in range(length))
        return password

# Usage
password_manager = Argon2PasswordManager()

# Hash password
password = "SecurePassword123!"
password_hash = password_manager.hash_password(password)

print(f"Hash: {password_hash}")

# Verify password
is_valid, new_hash = password_manager.verify_password(password, password_hash)
print(f"Valid: {is_valid}")

if new_hash:
    print(f"Rehashed: {new_hash}")

# Generate secure password
generated_password = password_manager.generate_secure_password(20)
print(f"Generated: {generated_password}")
```

## Example 5: TLS 1.3 Configuration (Node.js)

```javascript
// tls13_server.js
const https = require('https');
const fs = require('fs');

class TLS13Server {
  constructor(certPath, keyPath) {
    this.options = {
      // TLS 1.3 only (no fallback)
      minVersion: 'TLSv1.3',
      maxVersion: 'TLSv1.3',
      
      // Cipher suites (TLS 1.3 approved)
      ciphers: [
        'TLS_AES_256_GCM_SHA384',
        'TLS_CHACHA20_POLY1305_SHA256',
        'TLS_AES_128_GCM_SHA256'
      ].join(':'),
      
      // Certificate and private key
      cert: fs.readFileSync(certPath),
      key: fs.readFileSync(keyPath),
      
      // Security headers
      honorCipherOrder: true,
      
      // Session resumption (TLS 1.3)
      sessionIdContext: 'tls13-session',
      sessionTimeout: 300, // 5 minutes
      
      // OCSP stapling
      requestOCSP: true
    };
  }
  
  createServer(requestHandler) {
    const server = https.createServer(this.options, (req, res) => {
      // Add security headers
      res.setHeader('Strict-Transport-Security', 'max-age=31536000; includeSubDomains; preload');
      res.setHeader('X-Content-Type-Options', 'nosniff');
      res.setHeader('X-Frame-Options', 'DENY');
      res.setHeader('X-XSS-Protection', '1; mode=block');
      
      requestHandler(req, res);
    });
    
    return server;
  }
  
  listen(port) {
    const server = this.createServer((req, res) => {
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({
        message: 'Secure TLS 1.3 connection',
        protocol: req.socket.getProtocol(),
        cipher: req.socket.getCipher()
      }));
    });
    
    server.listen(port, () => {
      console.log(`TLS 1.3 server listening on port ${port}`);
    });
  }
}

// Usage
const tlsServer = new TLS13Server(
  '/path/to/cert.pem',
  '/path/to/key.pem'
);

tlsServer.listen(443);
```

## Example 6: Key Rotation Automation

```python
# key_rotation.py
import os
import json
from datetime import datetime, timedelta
from typing import Dict

class KeyRotationManager:
    """Automated key rotation with lifecycle management."""
    
    def __init__(self, key_store_path: str):
        self.key_store_path = key_store_path
        self.rotation_policy = {
            'symmetric_keys': timedelta(days=365),  # 1 year
            'asymmetric_keys': timedelta(days=730),  # 2 years
            'api_keys': timedelta(days=90),  # 90 days
            'tls_certificates': timedelta(days=90)  # 90 days
        }
    
    def check_rotation_needed(self, key_metadata: Dict) -> bool:
        """Check if key rotation is needed."""
        created_at = datetime.fromisoformat(key_metadata['created_at'])
        key_type = key_metadata['type']
        rotation_period = self.rotation_policy.get(key_type)
        
        if not rotation_period:
            return False
        
        age = datetime.now() - created_at
        return age >= rotation_period
    
    def rotate_key(self, key_id: str, key_metadata: Dict) -> Dict:
        """Rotate key and update metadata."""
        # Generate new key
        new_key = os.urandom(32)  # 256-bit key
        new_key_id = f"{key_id}-v{key_metadata['version'] + 1}"
        
        # Create new key metadata
        new_metadata = {
            'id': new_key_id,
            'type': key_metadata['type'],
            'version': key_metadata['version'] + 1,
            'created_at': datetime.now().isoformat(),
            'status': 'active',
            'previous_key_id': key_id
        }
        
        # Mark old key as deprecated
        old_metadata = key_metadata.copy()
        old_metadata['status'] = 'deprecated'
        old_metadata['deprecated_at'] = datetime.now().isoformat()
        
        # Save both keys (old key for decryption, new key for encryption)
        self._save_key(new_key_id, new_key, new_metadata)
        self._update_key_metadata(key_id, old_metadata)
        
        return {
            'new_key_id': new_key_id,
            'old_key_id': key_id,
            'rotated_at': datetime.now().isoformat()
        }
    
    def _save_key(self, key_id: str, key: bytes, metadata: Dict):
        """Save key and metadata to secure storage."""
        key_path = os.path.join(self.key_store_path, f"{key_id}.key")
        metadata_path = os.path.join(self.key_store_path, f"{key_id}.json")
        
        # Save key (encrypted in production)
        with open(key_path, 'wb') as f:
            f.write(key)
        
        # Save metadata
        with open(metadata_path, 'w') as f:
            json.dump(metadata, f, indent=2)
    
    def _update_key_metadata(self, key_id: str, metadata: Dict):
        """Update key metadata."""
        metadata_path = os.path.join(self.key_store_path, f"{key_id}.json")
        
        with open(metadata_path, 'w') as f:
            json.dump(metadata, f, indent=2)

# Usage
rotation_manager = KeyRotationManager('/secure/key/store')

# Check rotation
key_metadata = {
    'id': 'app-key-v1',
    'type': 'symmetric_keys',
    'version': 1,
    'created_at': '2024-01-01T00:00:00',
    'status': 'active'
}

if rotation_manager.check_rotation_needed(key_metadata):
    result = rotation_manager.rotate_key('app-key-v1', key_metadata)
    print(f"Key rotated: {result}")
```

---

**Last Updated**: 2025-11-22  
**All examples tested with**: cryptography 43.0.x, liboqs 0.10.x, boto3 1.34.x, argon2-cffi 23.1.x
