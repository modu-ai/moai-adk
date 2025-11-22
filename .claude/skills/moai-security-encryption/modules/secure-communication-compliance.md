---
name: secure-communication-compliance
parent: moai-security-encryption
description: TLS configuration and regulatory compliance
---

# Module 3: Secure Communication & Compliance

## Secure Communication

```python
import ssl
import socket

class SecureCommunication:
    def create_secure_context(
        self,
        cert_path: str,
        key_path: str
    ) -> ssl.SSLContext:
        context = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
        
        context.load_cert_chain(
            certfile=cert_path,
            keyfile=key_path
        )
        
        context.set_ciphers(
            'ECDHE-ECDSA-AES256-GCM-SHA384'
        )
        
        context.minimum_version = ssl.TLSVersion.TLSv1_3
        
        return context
    
    def create_secure_socket(
        self,
        host: str,
        port: int,
        context: ssl.SSLContext
    ) -> ssl.SSLSocket:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        secure_sock = context.wrap_socket(
            sock,
            server_hostname=host
        )
        secure_sock.connect((host, port))
        return secure_sock
```

## Certificate Management

```python
from cryptography import x509
from cryptography.hazmat.primitives.asymmetric import rsa

class CertificateManager:
    def generate_self_signed_certificate(
        self,
        common_name: str,
        valid_days: int = 365
    ):
        private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=2048
        )
        
        subject = issuer = x509.Name([
            x509.NameAttribute(NameOID.COMMON_NAME, common_name)
        ])
        
        cert = x509.CertificateBuilder().subject_name(
            subject
        ).issuer_name(
            issuer
        ).public_key(
            private_key.public_key()
        ).serial_number(
            x509.random_serial_number()
        ).not_valid_before(
            datetime.datetime.utcnow()
        ).not_valid_after(
            datetime.datetime.utcnow() + 
            datetime.timedelta(days=valid_days)
        ).sign(private_key, hashes.SHA256())
        
        return cert, private_key
```

## Compliance Standards

### FIPS 140-2/3
- Cryptographic module validation
- Hardware security requirements
- Physical security testing

### NIST SP 800-57
- Key management lifecycle
- Key generation requirements
- Rotation schedules

### PCI DSS
- Cardholder data encryption
- Key management procedures
- Access control requirements

### GDPR/HIPAA
- Data protection measures
- Privacy safeguards
- Audit trail requirements

---

**References**:
- [NIST Standards](https://csrc.nist.gov/)
- [PCI DSS](https://www.pcisecuritystandards.org/)
- [GDPR](https://gdpr.eu/)
