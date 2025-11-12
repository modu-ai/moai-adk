# moai-security-encryption: Advanced Cryptography & Data Protection

**Expert-Level Encryption Implementation for Enterprise Security**  
Trust Score: 10/10 | Version: 4.0.0 | Last Updated: 2025-11-11

## 🔐 Symmetric Encryption (AES)

### AES-256-GCM Implementation

```python
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.primitives import hashes, padding
from cryptography.hazmat.backends import default_backend
import os
import base64
from typing import Tuple, Dict, Any

class AESEncryptionService:
    def __init__(self):
        self.backend = default_backend()
    
    def generate_key(self) -> bytes:
        """Generate 256-bit AES key"""
        return os.urandom(32)
    
    def encrypt(self, plaintext: str, key: bytes) -> Dict[str, str]:
        """Encrypt using AES-256-GCM"""
        # Generate random nonce
        nonce = os.urandom(12)
        
        # Create cipher
        cipher = Cipher(
            algorithms.AES(key),
            modes.GCM(nonce),
            backend=self.backend
        )
        encryptor = cipher.encryptor()
        
        # Encrypt plaintext
        ciphertext = encryptor.update(plaintext.encode()) + encryptor.finalize()
        
        return {
            'ciphertext': base64.b64encode(ciphertext).decode(),
            'nonce': base64.b64encode(nonce).decode(),
            'tag': base64.b64encode(encryptor.tag).decode()
        }
    
    def decrypt(self, encrypted_data: Dict[str, str], key: bytes) -> str:
        """Decrypt using AES-256-GCM"""
        ciphertext = base64.b64decode(encrypted_data['ciphertext'])
        nonce = base64.b64decode(encrypted_data['nonce'])
        tag = base64.b64decode(encrypted_data['tag'])
        
        cipher = Cipher(
            algorithms.AES(key),
            modes.GCM(nonce, tag),
            backend=self.backend
        )
        decryptor = cipher.decryptor()
        
        plaintext = decryptor.update(ciphertext) + decryptor.finalize()
        return plaintext.decode()

# Asymmetric Encryption (RSA)
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.primitives import serialization

class RSAEncryptionService:
    def generate_key_pair(self, key_size: int = 2048) -> Tuple[bytes, bytes]:
        """Generate RSA key pair"""
        private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=key_size,
            backend=default_backend()
        )
        
        private_pem = private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption()
        )
        
        public_key = private_key.public_key()
        public_pem = public_key.public_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PublicFormat.SubjectPublicKeyInfo
        )
        
        return private_pem, public_pem
    
    def encrypt(self, plaintext: str, public_key_pem: bytes) -> bytes:
        """Encrypt with RSA public key"""
        public_key = serialization.load_pem_public_key(public_key_pem)
        ciphertext = public_key.encrypt(
            plaintext.encode(),
            padding.OAEP(
                mgf=padding.MGF1(algorithm=hashes.SHA256()),
                algorithm=hashes.SHA256(),
                label=None
            )
        )
        return ciphertext

# Key Derivation
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC
import secrets

class KeyDerivationService:
    def derive_key(self, password: str, salt: bytes, iterations: int = 100000) -> bytes:
        """Derive encryption key from password"""
        kdf = PBKDF2HMAC(
            algorithm=hashes.SHA256(),
            length=32,
            salt=salt,
            iterations=iterations,
            backend=default_backend()
        )
        return kdf.derive(password.encode())
    
    def generate_salt(self) -> bytes:
        """Generate cryptographic salt"""
        return os.urandom(16)

# Password Hashing
import hashlib
import hmac

class PasswordHashingService:
    def hash_password(self, password: str) -> Tuple[str, str]:
        """Hash password with salt"""
        salt = os.urandom(32)
        
        # Use Argon2 if available, otherwise PBKDF2
        try:
            import argon2
            hasher = argon2.PasswordHasher()
            hashed = hasher.hash(password + salt.hex())
            return hashed, salt.hex()
        except ImportError:
            # Fallback to PBKDF2
            kdf = PBKDF2HMAC(
                algorithm=hashes.SHA256(),
                length=32,
                salt=salt,
                iterations=100000,
                backend=default_backend()
            )
            hashed = base64.b64encode(kdf.derive(password.encode())).decode()
            return f"pbkdf2_sha256${hashed}", salt.hex()
    
    def verify_password(self, password: str, hashed: str, salt: str) -> bool:
        """Verify password against hash"""
        try:
            import argon2
            hasher = argon2.PasswordHasher()
            return hasher.verify(hashed, password + salt)
        except ImportError:
            # PBKDF2 verification
            kdf = PBKDF2HMAC(
                algorithm=hashes.SHA256(),
                length=32,
                salt=bytes.fromhex(salt),
                iterations=100000,
                backend=default_backend()
            )
            test_hash = base64.b64encode(kdf.derive(password.encode())).decode()
            return hmac.compare_digest(hashed.split('$', 1)[1], test_hash)

# Digital Signatures
from cryptography.hazmat.primitives import hashes

class DigitalSignatureService:
    def sign_message(self, message: str, private_key_pem: bytes) -> bytes:
        """Sign message with RSA private key"""
        private_key = serialization.load_pem_private_key(
            private_key_pem,
            password=None
        )
        
        signature = private_key.sign(
            message.encode(),
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()),
                salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256()
        )
        return signature
    
    def verify_signature(self, message: str, signature: bytes, public_key_pem: bytes) -> bool:
        """Verify signature with RSA public key"""
        public_key = serialization.load_pem_public_key(public_key_pem)
        
        try:
            public_key.verify(
                signature,
                message.encode(),
                padding.PSS(
                    mgf=padding.MGF1(hashes.SHA256()),
                    salt_length=padding.PSS.MAX_LENGTH
                ),
                hashes.SHA256()
            )
            return True
        except:
            return False

# Secure Key Management
class KeyManagementService:
    def __init__(self):
        self.keys = {}
        self.key_rotation_interval = 30  # days
    
    def store_key(self, key_id: str, key_data: bytes, metadata: Dict[str, Any]) -> None:
        """Securely store encryption key"""
        self.keys[key_id] = {
            'data': key_data,
            'created_at': datetime.now(),
            'metadata': metadata,
            'version': 1
        }
    
    def rotate_key(self, key_id: str) -> None:
        """Rotate encryption key"""
        if key_id in self.keys:
            old_key = self.keys[key_id]
            new_key_data = os.urandom(32)  # Generate new AES key
            
            self.keys[key_id] = {
                'data': new_key_data,
                'created_at': datetime.now(),
                'metadata': old_key['metadata'],
                'version': old_key['version'] + 1
            }
    
    def retire_key(self, key_id: str) -> None:
        """Retire encryption key"""
        if key_id in self.keys:
            del self.keys[key_id]
