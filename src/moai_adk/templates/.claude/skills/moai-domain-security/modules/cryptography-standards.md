# Cryptography Standards

Modern encryption, hashing, and digital signature patterns.

## Symmetric Encryption

### AES-256 Encryption

```python
from cryptography.fernet import Fernet
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
import os

class SymmetricEncryption:
    def encrypt_with_aes256(self, plaintext: str, key: bytes) -> str:
        # Generate random IV
        iv = os.urandom(16)

        cipher = Cipher(
            algorithms.AES(key),
            modes.CBC(iv)
        )

        encryptor = cipher.encryptor()

        # Padding for block cipher
        padded_data = self._pad(plaintext.encode())
        ciphertext = encryptor.update(padded_data) + encryptor.finalize()

        # Combine IV and ciphertext
        return (iv + ciphertext).hex()

    def decrypt_with_aes256(self, encrypted: str, key: bytes) -> str:
        encrypted_bytes = bytes.fromhex(encrypted)

        iv = encrypted_bytes[:16]
        ciphertext = encrypted_bytes[16:]

        cipher = Cipher(
            algorithms.AES(key),
            modes.CBC(iv)
        )

        decryptor = cipher.decryptor()
        padded_plaintext = decryptor.update(ciphertext) + decryptor.finalize()

        return self._unpad(padded_plaintext).decode()

    def _pad(self, data: bytes) -> bytes:
        # PKCS7 padding
        pad_len = 16 - (len(data) % 16)
        return data + (bytes([pad_len]) * pad_len)

    def _unpad(self, data: bytes) -> bytes:
        pad_len = data[-1]
        return data[:-pad_len]
```

## Asymmetric Encryption

### RSA Encryption

```python
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.primitives import serialization

class AsymmetricEncryption:
    def generate_rsa_keypair(self):
        private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=2048
        )

        return private_key, private_key.public_key()

    def encrypt_with_public_key(self, plaintext: str, public_key) -> str:
        ciphertext = public_key.encrypt(
            plaintext.encode(),
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )

        return ciphertext.hex()

    def decrypt_with_private_key(self, ciphertext: str, private_key) -> str:
        plaintext = private_key.decrypt(
            bytes.fromhex(ciphertext),
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )

        return plaintext.decode()
```

## Hashing

### Password Hashing with Bcrypt

```python
import bcrypt

class PasswordHashing:
    def hash_password(self, password: str) -> str:
        # Generate salt and hash
        salt = bcrypt.gensalt(rounds=12)
        hashed = bcrypt.hashpw(password.encode(), salt)

        return hashed.decode()

    def verify_password(self, password: str, hashed: str) -> bool:
        return bcrypt.checkpw(password.encode(), hashed.encode())

    def hash_with_argon2(self, password: str) -> str:
        from argon2 import PasswordHasher

        ph = PasswordHasher()
        return ph.hash(password)

    def verify_argon2(self, password: str, hashed: str) -> bool:
        from argon2 import PasswordHasher
        from argon2.exceptions import VerifyMismatchError

        ph = PasswordHasher()

        try:
            ph.verify(hashed, password)
            return True
        except VerifyMismatchError:
            return False
```

### Data Integrity Hashing

```python
from cryptography.hazmat.primitives import hashes

class DataIntegrityHashing:
    def hash_with_sha256(self, data: str) -> str:
        digest = hashes.Hash(hashes.SHA256())
        digest.update(data.encode())

        return digest.finalize().hex()

    def hash_with_blake2b(self, data: str) -> str:
        # BLAKE2b is faster and more secure than SHA-256
        digest = hashes.Hash(hashes.BLAKE2b(64))
        digest.update(data.encode())

        return digest.finalize().hex()
```

## Digital Signatures

### RSA Digital Signatures

```python
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives import hashes

class DigitalSignatures:
    def sign_data(self, data: str, private_key) -> str:
        signature = private_key.sign(
            data.encode(),
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()),
                salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256()
        )

        return signature.hex()

    def verify_signature(self, data: str, signature: str, public_key) -> bool:
        try:
            public_key.verify(
                bytes.fromhex(signature),
                data.encode(),
                padding.PSS(
                    mgf=padding.MGF1(hashes.SHA256()),
                    salt_length=padding.PSS.MAX_LENGTH
                ),
                hashes.SHA256()
            )
            return True
        except Exception:
            return False
```

## Key Derivation

### PBKDF2 Key Derivation

```python
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2
from cryptography.hazmat.primitives import hashes
import os

class KeyDerivation:
    def derive_key_from_password(self, password: str, salt: bytes = None) -> tuple:
        if salt is None:
            salt = os.urandom(32)

        kdf = PBKDF2(
            algorithm=hashes.SHA256(),
            length=32,
            salt=salt,
            iterations=100000,
        )

        key = kdf.derive(password.encode())

        return key, salt

    def verify_key(self, password: str, salt: bytes, derived_key: bytes) -> bool:
        kdf = PBKDF2(
            algorithm=hashes.SHA256(),
            length=32,
            salt=salt,
            iterations=100000,
        )

        try:
            kdf.verify(password.encode(), derived_key)
            return True
        except Exception:
            return False
```

## Best Practices

### ✅ DO
- Use AES-256 for symmetric encryption
- Use RSA-2048 or ECC for asymmetric encryption
- Use bcrypt or Argon2 for password hashing
- Use SHA-256 or BLAKE2b for hashing
- Use PBKDF2 for key derivation
- Generate cryptographically secure random keys
- Use authenticated encryption (AES-GCM)
- Rotate encryption keys regularly

### ❌ DON'T
- Use MD5 or SHA-1
- Use DES or RC4
- Hardcode encryption keys
- Use weak key sizes
- Reuse IV/nonce with same key
- Store plaintext passwords
- Use same key for multiple purposes
- Ignore algorithm deprecation

---

**Libraries**: cryptography, bcrypt, argon2, pynacl
