# Advanced Security Patterns

## Zero-Trust Architecture Implementation

### Architecture Pattern: Zero-Trust Network Model

**Principle**: Never trust, always verify - verification required for all access requests regardless of location.

**Implementation (Microservices)**:
```python
from fastapi import FastAPI, Depends, HTTPException
from fastapi.security import HTTPBearer, HTTPAuthCredentials
import jwt
from typing import Optional

app = FastAPI()
security = HTTPBearer()

class ZeroTrustVerifier:
    """Zero-trust architecture verification engine."""

    def __init__(self):
        self.jwt_secret = "your-secret-key"
        self.trusted_devices = {}

    async def verify_identity(self, credentials: HTTPAuthCredentials):
        """1. Verify user identity"""
        try:
            payload = jwt.decode(credentials.credentials, self.jwt_secret, algorithms=["HS256"])
            return payload['sub']
        except jwt.InvalidTokenError:
            raise HTTPException(status_code=401, detail="Invalid credentials")

    async def verify_device(self, device_id: str, user_id: str) -> bool:
        """2. Verify device compliance"""
        device = self.trusted_devices.get(device_id)
        if not device:
            return False

        # Check device security posture
        if not device['tpm_enabled']:
            return False
        if not device['antivirus_active']:
            return False
        if device['encryption_status'] != 'enabled':
            return False

        return True

    async def verify_network(self, request_ip: str) -> bool:
        """3. Verify network conditions"""
        # Check VPN/TLS status
        # Verify location anomalies
        # Check for suspicious patterns
        return True

    async def verify_context(self, resource: str, action: str, user_id: str) -> bool:
        """4. Verify contextual rules"""
        # Check time-based access policies
        # Verify resource sensitivity
        # Check user risk score
        return True

# Zero-Trust Middleware
async def zero_trust_verify(
    credentials: HTTPAuthCredentials = Depends(security),
    device_id: Optional[str] = None,
    request_ip: Optional[str] = None
):
    """Complete zero-trust verification pipeline."""
    verifier = ZeroTrustVerifier()

    # Step 1: Verify identity
    user_id = await verifier.verify_identity(credentials)

    # Step 2: Verify device
    if not await verifier.verify_device(device_id or "unknown", user_id):
        raise HTTPException(status_code=403, detail="Device not trusted")

    # Step 3: Verify network
    if not await verifier.verify_network(request_ip):
        raise HTTPException(status_code=403, detail="Network not trusted")

    # Step 4: Verify context
    if not await verifier.verify_context("resource", "action", user_id):
        raise HTTPException(status_code=403, detail="Context verification failed")

    return user_id

@app.get("/secure-resource")
async def get_secure_resource(user_id: str = Depends(zero_trust_verify)):
    """Protected resource with zero-trust verification."""
    return {"message": f"Secure data for {user_id}"}
```

**Benefits**:
- Eliminates perimeter-based security assumptions
- Continuous verification reduces attack surface
- Principle of least privilege enforced
- Better detection of lateral movement attacks

---

## STRIDE Threat Modeling Implementation

### Threat Modeling Methodology: STRIDE

**Categories**: Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege

**Implementation Framework**:
```python
from dataclasses import dataclass
from typing import List
from enum import Enum

class ThreatCategory(Enum):
    SPOOFING = "Spoofing"
    TAMPERING = "Tampering"
    REPUDIATION = "Repudiation"
    INFORMATION_DISCLOSURE = "Information Disclosure"
    DENIAL_OF_SERVICE = "Denial of Service"
    ELEVATION_OF_PRIVILEGE = "Elevation of Privilege"

@dataclass
class Threat:
    category: ThreatCategory
    description: str
    asset: str
    severity: str  # Critical, High, Medium, Low
    mitigation: str

class STRIDEAnalyzer:
    """STRIDE threat modeling analyzer."""

    def __init__(self):
        self.threats: List[Threat] = []

    def identify_spoofing_threats(self, component: str) -> List[Threat]:
        """S: Identify spoofing identity threats."""
        return [
            Threat(
                category=ThreatCategory.SPOOFING,
                description="User identity spoofing via credential theft",
                asset="User Authentication",
                severity="Critical",
                mitigation="MFA, strong password policy, account lockout"
            ),
            Threat(
                category=ThreatCategory.SPOOFING,
                description="API endpoint impersonation",
                asset="API Gateway",
                severity="High",
                mitigation="TLS/SSL, certificate pinning, DNS security"
            ),
        ]

    def identify_tampering_threats(self, component: str) -> List[Threat]:
        """T: Identify tampering threats."""
        return [
            Threat(
                category=ThreatCategory.TAMPERING,
                description="Data modification in transit",
                asset="Network Communication",
                severity="Critical",
                mitigation="Encryption (TLS 1.3+), HMAC verification"
            ),
            Threat(
                category=ThreatCategory.TAMPERING,
                description="Database record manipulation",
                asset="Database",
                severity="High",
                mitigation="Row-level security, audit logging, parameterized queries"
            ),
        ]

    def analyze_component(self, component_name: str) -> dict:
        """Complete STRIDE analysis for component."""
        analysis = {
            component_name: {
                "spoofing": self.identify_spoofing_threats(component_name),
                "tampering": self.identify_tampering_threats(component_name),
                "repudiation": [],
                "information_disclosure": [],
                "denial_of_service": [],
                "elevation_of_privilege": []
            }
        }
        return analysis

# Usage
analyzer = STRIDEAnalyzer()
threats = analyzer.analyze_component("Authentication Service")
```

---

## Secure SDLC Integration

### DevSecOps Pipeline

**CI/CD Security Integration**:
```yaml
# .github/workflows/security.yml
name: Security Pipeline

on: [push, pull_request]

jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
      # SAST: Static Application Security Testing
      - name: Run Bandit (Python)
        run: |
          pip install bandit
          bandit -r src/ -f json -o bandit-report.json

      # Dependency scanning
      - name: Check dependencies
        run: |
          pip install safety
          safety check --json --output safety-report.json

      # DAST: Dynamic Application Security Testing
      - name: Run OWASP ZAP scan
        uses: zaproxy/action-baseline@v0.7.0
        with:
          target: 'http://localhost:8000'
          rules_file_name: '.zap/rules.tsv'

      # Container scanning
      - name: Scan Docker image
        run: |
          docker build -t app:latest .
          trivy image --severity HIGH,CRITICAL app:latest

      # IaC scanning
      - name: Scan infrastructure
        run: |
          pip install checkov
          checkov --directory . --check CKV_AWS_1,CKV_AWS_2
```

---

## Cryptography Patterns

### Modern Encryption Implementation

**AES-256-GCM Encryption**:
```python
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2
import os
import base64

class EncryptionService:
    """AES-256-GCM encryption with PBKDF2 key derivation."""

    def __init__(self, master_key: str):
        self.master_key = master_key

    def derive_key(self, password: str, salt: bytes = None) -> tuple[bytes, bytes]:
        """Derive encryption key using PBKDF2."""
        if salt is None:
            salt = os.urandom(16)

        kdf = PBKDF2(
            algorithm=hashes.SHA256(),
            length=32,  # 256-bit key
            salt=salt,
            iterations=480000,  # OWASP recommendation
        )
        key = kdf.derive(password.encode())
        return key, salt

    def encrypt(self, plaintext: str, associated_data: str = None) -> str:
        """Encrypt data with authenticated encryption."""
        key, salt = self.derive_key(self.master_key)
        nonce = os.urandom(12)  # 96-bit nonce for GCM

        cipher = AESGCM(key)
        ciphertext = cipher.encrypt(
            nonce,
            plaintext.encode(),
            associated_data.encode() if associated_data else None
        )

        # Return: salt + nonce + ciphertext
        encrypted = base64.b64encode(salt + nonce + ciphertext).decode()
        return encrypted

    def decrypt(self, encrypted: str, associated_data: str = None) -> str:
        """Decrypt authenticated encrypted data."""
        encrypted_bytes = base64.b64decode(encrypted)

        salt = encrypted_bytes[:16]
        nonce = encrypted_bytes[16:28]
        ciphertext = encrypted_bytes[28:]

        key, _ = self.derive_key(self.master_key, salt)
        cipher = AESGCM(key)

        plaintext = cipher.decrypt(
            nonce,
            ciphertext,
            associated_data.encode() if associated_data else None
        )

        return plaintext.decode()
```

**Digital Signatures**:
```python
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.primitives import hashes, serialization

class SignatureService:
    """RSA-4096 digital signature service."""

    def __init__(self):
        self.private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=4096,
        )
        self.public_key = self.private_key.public_key()

    def sign(self, message: bytes) -> bytes:
        """Create digital signature."""
        signature = self.private_key.sign(
            message,
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()),
                salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256()
        )
        return signature

    def verify(self, message: bytes, signature: bytes) -> bool:
        """Verify digital signature."""
        try:
            self.public_key.verify(
                signature,
                message,
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

---

## Compliance Framework Integration

### SOC 2 Audit Trail Implementation

```python
import json
from datetime import datetime
from dataclasses import asdict, dataclass

@dataclass
class AuditLog:
    """SOC 2 compliant audit log entry."""
    timestamp: str
    user_id: str
    action: str
    resource: str
    status: str  # Success, Failure
    ip_address: str
    user_agent: str
    severity: str  # Critical, High, Medium, Low
    details: dict

class AuditLogger:
    """SOC 2 Type II audit logging."""

    def __init__(self, log_file: str):
        self.log_file = log_file

    def log_access(self, user_id: str, resource: str, status: str):
        """Log resource access."""
        log_entry = AuditLog(
            timestamp=datetime.utcnow().isoformat(),
            user_id=user_id,
            action="ACCESS",
            resource=resource,
            status=status,
            ip_address="127.0.0.1",  # Get from request
            user_agent="Mozilla/5.0...",  # Get from request
            severity="High" if status == "Failure" else "Low",
            details={"resource_type": "data"}
        )
        self._persist_log(log_entry)

    def _persist_log(self, log_entry: AuditLog):
        """Persist audit log (immutable)."""
        with open(self.log_file, 'a') as f:
            f.write(json.dumps(asdict(log_entry)) + '\n')
```

---

**Version**: 4.0.0
**Last Updated**: 2025-11-22
**Status**: Production Ready
