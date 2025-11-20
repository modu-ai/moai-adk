# Encryption Examples

Practical encryption implementations for Python and Node.js.

---

## AES-GCM Encryption

### Python Implementation

```python
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2
import os
import base64

class AESEncryption:
    """AES-256-GCM encryption wrapper."""

    def __init__(self, key: bytes):
        """Initialize with 256-bit key."""
        if len(key) != 32:  # 256 bits = 32 bytes
            raise ValueError("Key must be 32 bytes for AES-256")
        self.cipher = AESGCM(key)

    @classmethod
    def from_password(cls, password: str, salt: bytes = None):
        """Derive key from password using PBKDF2."""
        if salt is None:
            salt = os.urandom(16)

        kdf = PBKDF2(
            algorithm=hashes.SHA256(),
            length=32,
            salt=salt,
            iterations=100000,  # NIST recommendation
        )
        key = kdf.derive(password.encode())
        return cls(key), salt

    def encrypt(self, plaintext: bytes) -> tuple[bytes, bytes]:
        """Encrypt data and return (ciphertext, nonce)."""
        nonce = os.urandom(12)  # 96 bits for GCM
        ciphertext = self.cipher.encrypt(nonce, plaintext, None)
        return ciphertext, nonce

    def decrypt(self, ciphertext: bytes, nonce: bytes) -> bytes:
        """Decrypt data."""
        plaintext = self.cipher.decrypt(nonce, ciphertext, None)
        return plaintext

# Usage
key = AESGCM.generate_key(bit_length=256)
aes = AESEncryption(key)

# Encrypt
message = b"Sensitive data"
ciphertext, nonce = aes.encrypt(message)

# Decrypt
plaintext = aes.decrypt(ciphertext, nonce)
assert plaintext == message

# Store encrypted data
stored_data = {
    "ciphertext": base64.b64encode(ciphertext).decode(),
    "nonce": base64.b64encode(nonce).decode(),
}
```

### Node.js Implementation

```javascript
const crypto = require("crypto");

class AESEncryption {
  constructor(key) {
    if (key.length !== 32) {
      throw new Error("Key must be 32 bytes for AES-256");
    }
    this.key = key;
  }

  static fromPassword(password, salt = null) {
    if (!salt) {
      salt = crypto.randomBytes(16);
    }

    const key = crypto.pbkdf2Sync(
      password,
      salt,
      100000, // iterations
      32, // key length
      "sha256"
    );

    return { cipher: new AESEncryption(key), salt };
  }

  encrypt(plaintext) {
    const iv = crypto.randomBytes(12); // 96 bits for GCM
    const cipher = crypto.createCipheriv("aes-256-gcm", this.key, iv);

    let ciphertext = cipher.update(plaintext, "utf8", "hex");
    ciphertext += cipher.final("hex");

    const authTag = cipher.getAuthTag();

    return {
      ciphertext,
      iv: iv.toString("hex"),
      authTag: authTag.toString("hex"),
    };
  }

  decrypt(ciphertext, iv, authTag) {
    const decipher = crypto.createDecipheriv(
      "aes-256-gcm",
      this.key,
      Buffer.from(iv, "hex")
    );

    decipher.setAuthTag(Buffer.from(authTag, "hex"));

    let plaintext = decipher.update(ciphertext, "hex", "utf8");
    plaintext += decipher.final("utf8");

    return plaintext;
  }
}

// Usage
const key = crypto.randomBytes(32);
const aes = new AESEncryption(key);

const message = "Sensitive data";
const encrypted = aes.encrypt(message);
const decrypted = aes.decrypt(
  encrypted.ciphertext,
  encrypted.iv,
  encrypted.authTag
);

console.log(decrypted === message); // true
```

---

## RSA Encryption

### Python Implementation

```python
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.primitives import hashes, serialization

class RSAEncryption:
    """RSA-4096 encryption wrapper."""

    @staticmethod
    def generate_keypair():
        """Generate RSA-4096 key pair."""
        private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=4096,  # 2025 standard
        )
        public_key = private_key.public_key()
        return private_key, public_key

    @staticmethod
    def encrypt(public_key, plaintext: bytes) -> bytes:
        """Encrypt with public key."""
        ciphertext = public_key.encrypt(
            plaintext,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )
        return ciphertext

    @staticmethod
    def decrypt(private_key, ciphertext: bytes) -> bytes:
        """Decrypt with private key."""
        plaintext = private_key.decrypt(
            ciphertext,
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )
        return plaintext

    @staticmethod
    def save_private_key(private_key, filename: str, password: bytes = None):
        """Save private key to file (encrypted if password provided)."""
        encryption = serialization.BestAvailableEncryption(password) if password else serialization.NoEncryption()

        pem = private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=encryption
        )

        with open(filename, 'wb') as f:
            f.write(pem)

    @staticmethod
    def load_private_key(filename: str, password: bytes = None):
        """Load private key from file."""
        with open(filename, 'rb') as f:
            private_key = serialization.load_pem_private_key(
                f.read(),
                password=password
            )
        return private_key

# Usage
private_key, public_key = RSAEncryption.generate_keypair()

# Encrypt
message = b"Secret message"
ciphertext = RSAEncryption.encrypt(public_key, message)

# Decrypt
plaintext = RSAEncryption.decrypt(private_key, ciphertext)
assert plaintext == message

# Save keys
RSAEncryption.save_private_key(private_key, 'private.pem', password=b'strongpassword')
```

---

## Password Hashing

### bcrypt (Python)

```python
import bcrypt

class PasswordHasher:
    """Secure password hashing with bcrypt."""

    @staticmethod
    def hash_password(password: str, rounds: int = 12) -> str:
        """Hash password with bcrypt."""
        salt = bcrypt.gensalt(rounds=rounds)
        hashed = bcrypt.hashpw(password.encode('utf-8'), salt)
        return hashed.decode('utf-8')

    @staticmethod
    def verify_password(password: str, hashed: str) -> bool:
        """Verify password against hash."""
        return bcrypt.checkpw(
            password.encode('utf-8'),
            hashed.encode('utf-8')
        )

# Usage
hasher = PasswordHasher()

# Hash password
password = "MySecurePassword123!"
hashed = hasher.hash_password(password)

# Verify
is_valid = hasher.verify_password(password, hashed)  # True
is_invalid = hasher.verify_password("wrong", hashed)  # False
```

### Argon2 (Python)

```python
from argon2 import PasswordHasher
from argon2.exceptions import VerifyMismatchError

class Argon2Hasher:
    """Modern password hashing with Argon2id."""

    def __init__(self):
        self.ph = PasswordHasher(
            time_cost=3,        # iterations
            memory_cost=65536,  # 64 MB
            parallelism=4,      # threads
            hash_len=32,        # output length
            salt_len=16,        # salt length
        )

    def hash_password(self, password: str) -> str:
        """Hash password with Argon2id."""
        return self.ph.hash(password)

    def verify_password(self, password: str, hashed: str) -> bool:
        """Verify password against hash."""
        try:
            self.ph.verify(hashed, password)

            # Check if rehash needed (parameters changed)
            if self.ph.check_needs_rehash(hashed):
                # Rehash with new parameters
                return True, self.hash_password(password)

            return True, None
        except VerifyMismatchError:
            return False, None

# Usage
hasher = Argon2Hasher()

password = "MySecurePassword123!"
hashed = hasher.hash_password(password)

is_valid, new_hash = hasher.verify_password(password, hashed)
if new_hash:
    # Update database with new_hash
    pass
```

### bcrypt (Node.js)

```javascript
const bcrypt = require("bcrypt");

class PasswordHasher {
  static async hashPassword(password, rounds = 12) {
    const salt = await bcrypt.genSalt(rounds);
    const hashed = await bcrypt.hash(password, salt);
    return hashed;
  }

  static async verifyPassword(password, hashed) {
    return await bcrypt.compare(password, hashed);
  }
}

// Usage
(async () => {
  const password = "MySecurePassword123!";
  const hashed = await PasswordHasher.hashPassword(password);

  const isValid = await PasswordHasher.verifyPassword(password, hashed); // true
  const isInvalid = await PasswordHasher.verifyPassword("wrong", hashed); // false
})();
```

---

## Envelope Encryption

### Python Implementation

```python
class EnvelopeEncryption:
    """Envelope encryption pattern for large data."""

    def __init__(self, kek: bytes):
        """Initialize with Key Encryption Key (KEK)."""
        self.kek_cipher = AESGCM(kek)

    def encrypt_data(self, plaintext: bytes) -> dict:
        """
        Encrypt data using envelope encryption.

        Returns {
            'encrypted_data': encrypted with DEK,
            'encrypted_dek': DEK encrypted with KEK,
            'data_nonce': nonce for data,
            'dek_nonce': nonce for DEK
        }
        """
        # Generate Data Encryption Key (DEK)
        dek = AESGCM.generate_key(bit_length=256)
        dek_cipher = AESGCM(dek)

        # Encrypt data with DEK
        data_nonce = os.urandom(12)
        encrypted_data = dek_cipher.encrypt(data_nonce, plaintext, None)

        # Encrypt DEK with KEK
        dek_nonce = os.urandom(12)
        encrypted_dek = self.kek_cipher.encrypt(dek_nonce, dek, None)

        return {
            'encrypted_data': encrypted_data,
            'encrypted_dek': encrypted_dek,
            'data_nonce': data_nonce,
            'dek_nonce': dek_nonce,
        }

    def decrypt_data(self, envelope: dict) -> bytes:
        """Decrypt envelope-encrypted data."""
        # Decrypt DEK with KEK
        dek = self.kek_cipher.decrypt(
            envelope['dek_nonce'],
            envelope['encrypted_dek'],
            None
        )

        # Decrypt data with DEK
        dek_cipher = AESGCM(dek)
        plaintext = dek_cipher.decrypt(
            envelope['data_nonce'],
            envelope['encrypted_data'],
            None
        )

        return plaintext

# Usage
kek = AESGCM.generate_key(bit_length=256)  # Master key
envelope_enc = EnvelopeEncryption(kek)

# Encrypt large file
large_data = b"Large dataset..." * 1000
envelope = envelope_enc.encrypt_data(large_data)

# Store envelope (can rotate KEK without re-encrypting data)
import json
stored_envelope = {
    k: base64.b64encode(v).decode() if isinstance(v, bytes) else v
    for k, v in envelope.items()
}

# Decrypt
plaintext = envelope_enc.decrypt_data(envelope)
assert plaintext == large_data
```

### AWS KMS Integration (Python)

```python
import boto3
import base64

class AWSEnvelopeEncryption:
    """Envelope encryption using AWS KMS."""

    def __init__(self, kms_key_id: str):
        self.kms = boto3.client('kms')
        self.kms_key_id = kms_key_id

    def encrypt_data(self, plaintext: bytes) -> dict:
        """Encrypt using AWS KMS for KEK."""
        # Generate DEK using KMS
        response = self.kms.generate_data_key(
            KeyId=self.kms_key_id,
            KeySpec='AES_256'
        )

        dek_plaintext = response['Plaintext']
        dek_encrypted = response['CiphertextBlob']

        # Encrypt data with DEK
        cipher = AESGCM(dek_plaintext)
        nonce = os.urandom(12)
        encrypted_data = cipher.encrypt(nonce, plaintext, None)

        return {
            'encrypted_data': encrypted_data,
            'encrypted_dek': dek_encrypted,
            'nonce': nonce,
        }

    def decrypt_data(self, envelope: dict) -> bytes:
        """Decrypt using AWS KMS."""
        # Decrypt DEK using KMS
        response = self.kms.decrypt(
            CiphertextBlob=envelope['encrypted_dek']
        )
        dek_plaintext = response['Plaintext']

        # Decrypt data with DEK
        cipher = AESGCM(dek_plaintext)
        plaintext = cipher.decrypt(
            envelope['nonce'],
            envelope['encrypted_data'],
            None
        )

        return plaintext

# Usage (requires AWS credentials)
kms_enc = AWSEnvelopeEncryption('your-kms-key-id')

data = b"Sensitive data"
envelope = kms_enc.encrypt_data(data)
plaintext = kms_enc.decrypt_data(envelope)
```

---

## File Encryption

### Encrypt File (Python)

```python
def encrypt_file(input_path: str, output_path: str, password: str):
    """Encrypt file with AES-256-GCM."""
    # Derive key from password
    aes, salt = AESEncryption.from_password(password)

    # Read file
    with open(input_path, 'rb') as f:
        plaintext = f.read()

    # Encrypt
    ciphertext, nonce = aes.encrypt(plaintext)

    # Write encrypted file
    with open(output_path, 'wb') as f:
        f.write(salt)  # 16 bytes
        f.write(nonce)  # 12 bytes
        f.write(ciphertext)

def decrypt_file(input_path: str, output_path: str, password: str):
    """Decrypt file."""
    # Read encrypted file
    with open(input_path, 'rb') as f:
        salt = f.read(16)
        nonce = f.read(12)
        ciphertext = f.read()

    # Derive key
    aes, _ = AESEncryption.from_password(password, salt)

    # Decrypt
    plaintext = aes.decrypt(ciphertext, nonce)

    # Write decrypted file
    with open(output_path, 'wb') as f:
        f.write(plaintext)

# Usage
encrypt_file('secret.txt', 'secret.enc', 'MyPassword123!')
decrypt_file('secret.enc', 'secret_decrypted.txt', 'MyPassword123!')
```

---

**See also**: [reference.md](./reference.md) for NIST standards and algorithm details.
