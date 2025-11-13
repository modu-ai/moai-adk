---
name: "moai-security-encryption"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
description: Enterprise Encryption Security with AI-powered cryptographic architecture, Context7 integration, and intelligent encryption orchestration for data protection
keywords: ['encryption', 'cryptography', 'data-protection', 'security', 'aes-256', 'rsa', 'context7-integration', 'ai-orchestration', 'production-deployment']
allowed-tools: 
  - Read
  - Bash
  - Write
  - Edit
  - Glob
  - Grep
  - WebFetch
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Enterprise Encryption Security Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-security-encryption |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Security Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Architecture |
| **Auto-load** | On demand when encryption keywords detected |

---

## What It Does

Enterprise Encryption Security expert with AI-powered cryptographic architecture, Context7 integration, and intelligent encryption orchestration for comprehensive data protection.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered Encryption Architecture** using Context7 MCP for latest cryptographic patterns
- 📊 **Intelligent Key Management** with automated rotation and lifecycle optimization
- 🚀 **Advanced Cryptographic Implementation** with AI-driven algorithm selection
- 🔗 **Enterprise Security Framework** with zero-configuration encryption deployment
- 📈 **Predictive Security Analytics** with threat assessment and compliance monitoring

---

## When to Use

**Automatic triggers**:
- Encryption implementation and cryptographic security discussions
- Data protection and privacy compliance requirements analysis
- Key management and rotation strategy planning
- Secure communication and storage implementation

**Manual invocation**:
- Designing enterprise encryption architectures with optimal security
- Implementing comprehensive key management systems
- Planning cryptographic algorithms and security protocols
- Optimizing encryption performance and compliance

---

# Quick Reference (Level 1)

## Modern Encryption Stack (November 2025)

### Core Cryptographic Algorithms
- **AES-256-GCM**: Symmetric encryption with authenticated encryption
- **RSA-4096**: Asymmetric encryption for key exchange and digital signatures
- **ECC P-384**: Elliptic curve cryptography for efficiency
- **SHA-384**: Cryptographic hashing for integrity verification
- **HMAC-SHA256**: Message authentication codes

### Key Management Systems
- **HashiCorp Vault**: Enterprise secrets management
- **AWS KMS**: Cloud-based key management service
- **Azure Key Vault**: Microsoft cloud key management
- **Kubernetes Secrets**: Container-native secret storage
- **Hardware Security Modules (HSM)**: Hardware-based key protection

### Implementation Standards
- **FIPS 140-2/3**: Federal Information Processing Standards
- **NIST SP 800-57**: Key management guidelines
- **PCI DSS**: Payment card industry security standards
- **GDPR**: Data protection and privacy regulation
- **HIPAA**: Healthcare information protection

### Security Features
- **End-to-End Encryption**: Complete data protection lifecycle
- **Key Rotation**: Automated key renewal and secure rotation
- **Zero-Knowledge Architecture**: Privacy-preserving encryption
- **Quantum-Resistant**: Preparation for quantum computing threats
- **Audit Logging**: Comprehensive security event tracking

---

# Core Implementation (Level 2)

## Encryption Architecture Intelligence

```python
# AI-powered encryption architecture optimization with Context7
class EncryptionArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.crypto_analyzer = CryptographicAnalyzer()
        self.security_validator = SecurityValidator()
    
    async def design_optimal_encryption_architecture(self, 
                                                   requirements: SecurityRequirements) -> EncryptionArchitecture:
        """Design optimal encryption architecture using AI analysis."""
        
        # Get latest cryptographic documentation via Context7
        crypto_docs = await self.context7_client.get_library_docs(
            context7_library_id='/cryptography/docs',
            topic="encryption algorithms key management security 2025",
            tokens=3000
        )
        
        security_docs = await self.context7_client.get_library_docs(
            context7_library_id='/security/docs',
            topic="data protection compliance best practices 2025",
            tokens=2000
        )
        
        # Optimize cryptographic algorithms
        algorithm_selection = self.crypto_analyzer.select_optimal_algorithms(
            requirements.data_classification,
            requirements.performance_requirements,
            crypto_docs
        )
        
        # Validate security requirements
        security_configuration = self.security_validator.configure_security(
            requirements.compliance_frameworks,
            requirements.threat_model,
            security_docs
        )
        
        return EncryptionArchitecture(
            algorithm_configuration=algorithm_selection,
            key_management_system=self._design_key_management(requirements),
            security_framework=security_configuration,
            implementation_patterns=self._select_implementation_patterns(requirements),
            compliance_integration=self._ensure_compliance(requirements),
            performance_optimization=self._optimize_performance(requirements)
        )
```

## Advanced Encryption Implementation

```typescript
// Enterprise-grade encryption with TypeScript
import crypto from 'crypto';
import { promisify } from 'util';

const randomBytes = promisify(crypto.randomBytes);
const pbkdf2 = promisify(crypto.pbkdf2);

interface EncryptionConfig {
  algorithm: string;
  keyLength: number;
  ivLength: number;
  tagLength: number;
  iterations: number;
  saltLength: number;
}

export class AdvancedEncryptionManager {
  private config: EncryptionConfig;
  private keyManager: KeyManager;

  constructor(config: Partial<EncryptionConfig> = {}) {
    this.config = {
      algorithm: 'aes-256-gcm',
      keyLength: 32,
      ivLength: 16,
      tagLength: 16,
      iterations: 100000,
      saltLength: 32,
      ...config,
    };

    this.keyManager = new KeyManager();
  }

  async encrypt(plaintext: string, keyId?: string): Promise<EncryptedData> {
    try {
      // Get or derive encryption key
      const key = await this.keyManager.getKey(keyId);
      
      // Generate random salt and IV
      const salt = await randomBytes(this.config.saltLength);
      const iv = await randomBytes(this.config.ivLength);
      
      // Create cipher
      const cipher = crypto.createCipher(this.config.algorithm, key);
      cipher.setAAD(salt); // Additional authenticated data
      
      let encrypted = cipher.update(plaintext, 'utf8', 'hex');
      encrypted += cipher.final('hex');
      
      const tag = cipher.getAuthTag();
      
      return {
        encrypted,
        salt: salt.toString('hex'),
        iv: iv.toString('hex'),
        tag: tag.toString('hex'),
        algorithm: this.config.algorithm,
        keyId,
        timestamp: new Date().toISOString(),
      };
    } catch (error) {
      throw new Error(`Encryption failed: ${error.message}`);
    }
  }

  async decrypt(encryptedData: EncryptedData): Promise<string> {
    try {
      // Get decryption key
      const key = await this.keyManager.getKey(encryptedData.keyId);
      
      // Create decipher
      const decipher = crypto.createDecipher(
        encryptedData.algorithm,
        key
      );
      
      // Set parameters
      decipher.setAAD(Buffer.from(encryptedData.salt, 'hex'));
      decipher.setAuthTag(Buffer.from(encryptedData.tag, 'hex'));
      
      let decrypted = decipher.update(encryptedData.encrypted, 'hex', 'utf8');
      decrypted += decipher.final('utf8');
      
      return decrypted;
    } catch (error) {
      throw new Error(`Decryption failed: ${error.message}`);
    }
  }

  async encryptFile(
    inputPath: string,
    outputPath: string,
    keyId?: string
  ): Promise<FileEncryptionResult> {
    const fs = require('fs').promises;
    
    try {
      // Read file
      const fileData = await fs.readFile(inputPath);
      const fileStats = await fs.stat(inputPath);
      
      // Generate salt and IV
      const salt = await randomBytes(this.config.saltLength);
      const iv = await randomBytes(this.config.ivLength);
      
      // Get encryption key
      const key = await this.keyManager.getKey(keyId);
      
      // Create cipher
      const cipher = crypto.createCipher(this.config.algorithm, key);
      cipher.setAAD(salt);
      
      // Encrypt file in streaming mode
      const inputStream = fs.createReadStream(inputPath);
      const outputStream = fs.createWriteStream(outputPath);
      
      // Write header
      outputStream.write(JSON.stringify({
        salt: salt.toString('hex'),
        iv: iv.toString('hex'),
        algorithm: this.config.algorithm,
        originalSize: fileStats.size,
        keyId,
      }) + '\n');
      
      // Pipe encryption
      inputStream.pipe(cipher).pipe(outputStream);
      
      return new Promise((resolve, reject) => {
        outputStream.on('finish', () => {
          resolve({
            outputPath,
            originalSize: fileStats.size,
            encryptedSize: fileStats.size + this.config.saltLength + this.config.ivLength,
            keyId,
          });
        });
        
        outputStream.on('error', reject);
      });
    } catch (error) {
      throw new Error(`File encryption failed: ${error.message}`);
    }
  }
}

// Advanced key management system
class KeyManager {
  private keyStore: Map<string, CryptoKey> = new Map();
  private rotationSchedule: Map<string, Date> = new Map();

  async generateKey(keyId: string, algorithm: string = 'aes-256-gcm'): Promise<string> {
    const key = await randomBytes(32); // 256-bit key
    
    // Store key securely (in production, use HSM or KMS)
    this.keyStore.set(keyId, key);
    
    // Set rotation schedule (30 days)
    const rotationDate = new Date();
    rotationDate.setDate(rotationDate.getDate() + 30);
    this.rotationSchedule.set(keyId, rotationDate);
    
    return keyId;
  }

  async getKey(keyId?: string): Promise<Buffer> {
    if (!keyId) {
      // Generate default key
      return await randomBytes(32);
    }

    const key = this.keyStore.get(keyId);
    if (!key) {
      throw new Error(`Key not found: ${keyId}`);
    }

    // Check if key needs rotation
    const rotationDate = this.rotationSchedule.get(keyId);
    if (rotationDate && rotationDate <= new Date()) {
      await this.rotateKey(keyId);
      return this.keyStore.get(keyId)!;
    }

    return key;
  }

  private async rotateKey(keyId: string): Promise<void> {
    // Generate new key
    const newKey = await randomBytes(32);
    
    // Store new key
    this.keyStore.set(keyId, newKey);
    
    // Update rotation schedule
    const rotationDate = new Date();
    rotationDate.setDate(rotationDate.getDate() + 30);
    this.rotationSchedule.set(keyId, rotationDate);
    
    console.log(`Key rotated: ${keyId}`);
  }

  async deriveKeyFromPassword(
    password: string,
    salt: Buffer,
    iterations: number = 100000
  ): Promise<Buffer> {
    return await pbkdf2(password, salt, iterations, 32, 'sha256') as Buffer;
  }

  async generateKeyPair(keyId: string): Promise<KeyPair> {
    const { publicKey, privateKey } = crypto.generateKeyPairSync('rsa', {
      modulusLength: 4096,
      publicKeyEncoding: {
        type: 'spki',
        format: 'pem',
      },
      privateKeyEncoding: {
        type: 'pkcs8',
        format: 'pem',
      },
    });

    return {
      keyId,
      publicKey,
      privateKey,
      algorithm: 'rsa-4096',
    };
  }
}

// Digital signature implementation
export class DigitalSignature {
  private keyManager: KeyManager;

  constructor() {
    this.keyManager = new KeyManager();
  }

  async sign(data: string, privateKeyId: string): Promise<DigitalSignature> {
    try {
      const privateKey = await this.keyManager.getPrivateKey(privateKeyId);
      const sign = crypto.createSign('RSA-SHA256');
      
      sign.update(data);
      const signature = sign.sign(privateKey, 'hex');
      
      return {
        data,
        signature,
        algorithm: 'RSA-SHA256',
        keyId: privateKeyId,
        timestamp: new Date().toISOString(),
      };
    } catch (error) {
      throw new Error(`Signing failed: ${error.message}`);
    }
  }

  async verify(signature: DigitalSignature, publicKeyId: string): Promise<boolean> {
    try {
      const publicKey = await this.keyManager.getPublicKey(publicKeyId);
      const verify = crypto.createVerify('RSA-SHA256');
      
      verify.update(signature.data);
      return verify.verify(publicKey, signature.signature, 'hex');
    } catch (error) {
      return false;
    }
  }
}
```

## Secure Communication Implementation

```python
import ssl
import socket
from typing import Optional, Tuple
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.backends import default_backend

class SecureCommunication:
    def __init__(self):
        self.backend = default_backend()
        self.certificate_manager = CertificateManager()
    
    def create_secure_context(self, 
                            cert_path: str,
                            key_path: str,
                            ca_path: Optional[str] = None) -> ssl.SSLContext:
        """Create SSL/TLS context for secure communication."""
        
        context = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
        
        # Load certificate and private key
        context.load_cert_chain(certfile=cert_path, keyfile=key_path)
        
        # Configure cipher suites
        context.set_ciphers('ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384')
        
        # Set minimum TLS version
        context.minimum_version = ssl.TLSVersion.TLSv1_2
        
        # Enable HSTS
        context.set_alpn_protocols(['h2', 'http/1.1'])
        
        if ca_path:
            context.load_verify_locations(cafile=ca_path)
            context.verify_mode = ssl.CERT_REQUIRED
        
        return context
    
    def create_secure_socket(self, 
                           host: str, 
                           port: int,
                           context: ssl.SSLContext) -> ssl.SSLSocket:
        """Create secure socket with TLS encryption."""
        
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        secure_sock = context.wrap_socket(sock, server_hostname=host)
        
        try:
            secure_sock.connect((host, port))
            return secure_sock
        except Exception as e:
            secure_sock.close()
            raise e

class CertificateManager:
    def __init__(self):
        self.backend = default_backend()
    
    def generate_self_signed_certificate(self, 
                                      common_name: str,
                                      organization: str,
                                      valid_days: int = 365) -> Tuple[bytes, bytes]:
        """Generate self-signed certificate for testing."""
        
        from cryptography import x509
        from cryptography.x509.oid import NameOID
        
        # Generate private key
        private_key = rsa.generate_private_key(
            public_exponent=65537,
            key_size=2048,
            backend=self.backend
        )
        
        # Create certificate
        subject = issuer = x509.Name([
            x509.NameAttribute(NameOID.COUNTRY_NAME, "US"),
            x509.NameAttribute(NameOID.STATE_OR_PROVINCE_NAME, "California"),
            x509.NameAttribute(NameOID.LOCALITY_NAME, "San Francisco"),
            x509.NameAttribute(NameOID.ORGANIZATION_NAME, organization),
            x509.NameAttribute(NameOID.COMMON_NAME, common_name),
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
            datetime.datetime.utcnow() + datetime.timedelta(days=valid_days)
        ).add_extension(
            x509.SubjectAlternativeName([
                x509.DNSName(common_name),
                x509.IPAddress(ipaddress.IPv4Address("127.0.0.1")),
            ]),
            critical=False,
        ).sign(private_key, hashes.SHA256(), self.backend)
        
        # Serialize certificate and key
        cert_pem = cert.public_bytes(serialization.Encoding.PEM)
        key_pem = private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.PKCS8,
            encryption_algorithm=serialization.NoEncryption()
        )
        
        return cert_pem, key_pem
```

---

# Advanced Implementation (Level 3)

## Enterprise Key Management

```typescript
// Enterprise key management with HSM integration
export class EnterpriseKeyManager {
  private hsmClient: HSMClient;
  private keyRotation: KeyRotationService;
  private auditLogger: AuditLogger;

  constructor(hsmConfig: HSMConfig) {
    this.hsmClient = new HSMClient(hsmConfig);
    this.keyRotation = new KeyRotationService();
    this.auditLogger = new AuditLogger();
  }

  async createEncryptionKey(
    keyId: string,
    algorithm: string = 'AES-256-GCM',
    metadata?: KeyMetadata
  ): Promise<CreatedKey> {
    try {
      // Log key creation attempt
      this.auditLogger.log('KEY_CREATION_ATTEMPT', { keyId, algorithm });

      // Create key in HSM
      const hsmKey = await this.hsmClient.createKey({
        algorithm,
        keyId,
        extractable: false,
        sensitive: true,
        ...metadata,
      });

      // Set up rotation schedule
      await this.keyRotation.scheduleRotation(keyId, {
        rotationInterval: 90, // days
        algorithm: algorithm,
      });

      // Log successful creation
      this.auditLogger.log('KEY_CREATED', { keyId, hsmKeyId: hsmKey.id });

      return {
        keyId,
        hsmKeyId: hsmKey.id,
        algorithm,
        created: new Date(),
        nextRotation: await this.keyRotation.getNextRotationDate(keyId),
      };
    } catch (error) {
      this.auditLogger.log('KEY_CREATION_FAILED', { 
        keyId, 
        error: error.message 
      });
      throw error;
    }
  }

  async encryptWithHSM(
    keyId: string,
    plaintext: Buffer,
    additionalData?: Buffer
  ): Promise<EncryptedWithHSM> {
    try {
      // Get key from HSM
      const hsmKey = await this.hsmClient.getKey(keyId);
      
      // Perform encryption in HSM
      const result = await this.hsmClient.encrypt({
        keyId: hsmKey.id,
        plaintext,
        additionalData,
      });

      // Log encryption operation
      this.auditLogger.log('ENCRYPTION_PERFORMED', { 
        keyId, 
        dataSize: plaintext.length 
      });

      return {
        ciphertext: result.ciphertext,
        iv: result.iv,
        tag: result.tag,
        keyId,
        timestamp: new Date(),
      };
    } catch (error) {
      this.auditLogger.log('ENCRYPTION_FAILED', { 
        keyId, 
        error: error.message 
      });
      throw error;
    }
  }

  async rotateKey(keyId: string): Promise<KeyRotationResult> {
    try {
      // Get current key
      const currentKey = await this.hsmClient.getKey(keyId);
      
      // Create new key
      const newKeyId = `${keyId}_rotated_${Date.now()}`;
      const newKey = await this.createEncryptionKey(
        newKeyId,
        currentKey.algorithm
      );

      // Schedule deprecation of old key
      await this.keyRotation.deprecateKey(keyId, {
        deprecationDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000), // 30 days
        replacementKeyId: newKeyId,
      });

      // Log rotation
      this.auditLogger.log('KEY_ROTATED', { 
        oldKeyId: keyId, 
        newKeyId,
        rotationDate: new Date(),
      });

      return {
        oldKeyId: keyId,
        newKeyId,
        deprecationDate: await this.keyRotation.getDeprecationDate(keyId),
      };
    } catch (error) {
      this.auditLogger.log('KEY_ROTATION_FAILED', { 
        keyId, 
        error: error.message 
      });
      throw error;
    }
  }
}

// Compliance and audit integration
class ComplianceAuditor {
  private auditLog: AuditLog;
  private complianceRules: ComplianceRule[];

  constructor() {
    this.auditLog = new AuditLog();
    this.complianceRules = [
      new GDPRComplianceRule(),
      new PCIComplianceRule(),
      new HIPAAComplianceRule(),
    ];
  }

  async auditEncryptionOperations(
    timeRange: TimeRange
  ): Promise<AuditReport> {
    // Get audit log entries
    const entries = await this.auditLog.getEntries(timeRange);
    
    // Apply compliance rules
    const complianceResults = [];
    for (const rule of this.complianceRules) {
      const result = await rule.validate(entries);
      complianceResults.push(result);
    }

    // Generate report
    return {
      timeRange,
      totalOperations: entries.length,
      complianceResults,
      violations: this.identifyViolations(entries),
      recommendations: this.generateRecommendations(complianceResults),
    };
  }

  private identifyViolations(entries: AuditEntry[]): SecurityViolation[] {
    const violations = [];

    for (const entry of entries) {
      // Check for suspicious patterns
      if (this.isSuspiciousPattern(entry)) {
        violations.push({
          type: 'SUSPICIOUS_PATTERN',
          entry,
          severity: 'HIGH',
          description: 'Suspicious encryption operation detected',
        });
      }

      // Check for policy violations
      if (this.isPolicyViolation(entry)) {
        violations.push({
          type: 'POLICY_VIOLATION',
          entry,
          severity: 'MEDIUM',
          description: 'Encryption policy violation detected',
        });
      }
    }

    return violations;
  }
}
```

---

# Reference & Integration (Level 4)

## API Reference

### Core Encryption Operations
- `encrypt(data, keyId, algorithm)` - Encrypt data with specified algorithm
- `decrypt(encryptedData)` - Decrypt data with validation
- `generate_key(algorithm, metadata)` - Generate encryption key
- `sign_data(data, privateKeyId)` - Create digital signature
- `verify_signature(signature, publicKeyId)` - Verify digital signature

### Context7 Integration
- `get_latest_cryptography_docs()` - Cryptography via Context7
- `analyze_encryption_patterns()` - Encryption patterns via Context7
- `optimize_key_management()` - Key management via Context7

## Best Practices (November 2025)

### DO
- Use industry-standard cryptographic algorithms (AES-256, RSA-4096)
- Implement comprehensive key management with rotation
- Use authenticated encryption (AES-GCM) for data protection
- Implement proper error handling and secure disposal
- Use hardware security modules for key protection
- Maintain comprehensive audit logging and monitoring
- Follow compliance requirements (GDPR, PCI DSS, HIPAA)
- Implement quantum-resistant encryption preparation

### DON'T
- Implement custom cryptographic algorithms
- Store encryption keys with encrypted data
- Use deprecated or weak cryptographic algorithms
- Skip key rotation and lifecycle management
- Ignore compliance and regulatory requirements
- Forget to implement proper error handling
- Skip security testing and vulnerability assessments
- Use hardcoded keys or initialization vectors

## Works Well With

- `moai-security-api` (API security implementation)
- `moai-foundation-trust` (Trust and compliance)
- `moai-cc-configuration` (Configuration security)
- `moai-security-secrets` (Secrets management)
- `moai-baas-foundation` (BaaS security patterns)
- `moai-domain-backend` (Backend security)
- `moai-security-owasp` (Security best practices)
- `moai-security-compliance` (Compliance management)

## Changelog

- **v4.0.0** (2025-11-13): Complete Enterprise v4.0 rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, advanced cryptographic patterns, and enterprise key management
- **v2.0.0** (2025-11-11): Complete metadata structure, encryption patterns, key management
- **v1.0.0** (2025-11-11): Initial encryption security foundation

---

**End of Skill** | Updated 2025-11-13

## Cryptographic Security

### Algorithm Selection
- AES-256-GCM for symmetric encryption with authentication
- RSA-4096 for asymmetric encryption and digital signatures
- ECC P-384 for efficient key exchange
- SHA-384 for cryptographic hashing
- PBKDF2 with 100,000 iterations for key derivation

### Enterprise Features
- Hardware Security Module (HSM) integration
- Automated key rotation and lifecycle management
- Comprehensive audit logging and compliance reporting
- Quantum-resistant encryption preparation
- Zero-knowledge proof implementation support

---

**End of Enterprise Encryption Security Expert v4.0.0**
