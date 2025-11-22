                },
              },
            ],
          },
        ],
      },
    };
    
    // Apply policies
    await this.k8sApi.createNamespacedCustomObject(
      'networking.k8s.io',
      'v1',
      namespace,
      'networkpolicies',
      denyPolicy
    );
    
    await this.k8sApi.createNamespacedCustomObject(
      'cilium.io',
      'v2',
      namespace,
      'ciliumnetworkpolicies',
      l7Policy
    );
    
    console.log(`Zero-trust policy applied to ${serviceName}`);
  }
}
```

### Pattern 2: mTLS Enforcement (Service Mesh)

```javascript
// Service mesh (Cilium, Istio) enforces mTLS between services
const { Issuer } = require('openid-client');

class mTLSEnforcement {
  constructor() {
    this.certs = new Map();
    this.trustStore = [];
  }
  
  // Issue certificate to service
  async issueCertificate(serviceName, namespace) {
    const cert = {
      subject: `/CN=${serviceName}.${namespace}.svc.cluster.local`,
      validity: {
        notBefore: new Date(),
        notAfter: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000),
      },
      keySize: 4096,
      algorithm: 'RSA',
    };
    
    // Store in secret
    await this.storeInK8sSecret(serviceName, namespace, cert);
    
    this.certs.set(`${serviceName}.${namespace}`, cert);
    
    return cert;
  }
  
  // Verify mutual TLS handshake
  async verifyMutualTLS(clientCert, serverCert) {
    // 1. Verify client certificate signature
    if (!this.verifyCertificateChain(clientCert)) {
      throw new Error('Invalid client certificate chain');
    }
    
    // 2. Verify server certificate signature
    if (!this.verifyCertificateChain(serverCert)) {
      throw new Error('Invalid server certificate chain');
    }
    
    // 3. Verify certificates are in trust store
    if (!this.isInTrustStore(clientCert)) {
      throw new Error('Client certificate not trusted');
    }
    
    if (!this.isInTrustStore(serverCert)) {
      throw new Error('Server certificate not trusted');
    }
    
    // 4. Check certificate validity
    const now = new Date();
    if (now < new Date(clientCert.notBefore) || now > new Date(clientCert.notAfter)) {
      throw new Error('Client certificate expired');
    }
    
    if (now < new Date(serverCert.notBefore) || now > new Date(serverCert.notAfter)) {
      throw new Error('Server certificate expired');
    }
    
    return {
      valid: true,
      clientName: clientCert.subject,
      serverName: serverCert.subject,
    };
  }
  
  verifyCertificateChain(cert) {
    // Verify signature using CA public key
    return true;  // Implementation detail
  }
  
  isInTrustStore(cert) {
    // Check if certificate is in trusted CA store
    return this.trustStore.some(ca => ca.subject === cert.issuer);
  }
}
```

### Pattern 3: Device Trust Verification (BeyondCorp)

```javascript
class DeviceTrustAssessment {
  async assessDeviceTrust(device) {
    const assessment = {
      deviceId: device.id,
      timestamp: new Date(),
      score: 0,
      checks: {},
    };
    
    // Check 1: Operating System
    const osCheck = await this.checkOS(device);
    assessment.checks.os = osCheck;
    assessment.score += osCheck.trusted ? 25 : 0;
    
    // Check 2: Antivirus/Anti-malware
    const avCheck = await this.checkAntivirus(device);
    assessment.checks.antivirus = avCheck;
    assessment.score += avCheck.enabled ? 25 : 0;
    
    // Check 3: Firewall
    const fwCheck = await this.checkFirewall(device);
    assessment.checks.firewall = fwCheck;
    assessment.score += fwCheck.enabled ? 25 : 0;
    
    // Check 4: Disk Encryption
    const encCheck = await this.checkDiskEncryption(device);
    assessment.checks.encryption = encCheck;
    assessment.score += encCheck.enabled ? 25 : 0;
    
    // Overall trust level
    assessment.trustLevel = this.calculateTrustLevel(assessment.score);
    
    return assessment;
  }
  
  calculateTrustLevel(score) {
    if (score >= 90) return 'TRUSTED';
    if (score >= 70) return 'CONDITIONAL';
    return 'UNTRUSTED';
  }
  
  async checkOS(device) {
    // Verify OS is up-to-date with security patches
    return {
      os: device.osType,
      version: device.osVersion,
      lastPatchDate: device.lastPatchDate,
      trusted: this.isOSPatched(device),
    };
  }
  
  async checkAntivirus(device) {
    // Verify antivirus is installed and current
    return {
      enabled: device.antivirusEnabled,
      product: device.antivirusProduct,
      lastUpdate: device.antivirusLastUpdate,
      signatureAge: this.calculateSignatureAge(device.antivirusLastUpdate),
    };
  }
  
  isOSPatched(device) {
    const daysSincePatch = Math.floor(
      (Date.now() - new Date(device.lastPatchDate)) / (24 * 60 * 60 * 1000)
    );
    
    // Consider patched if updated within 30 days
    return daysSincePatch <= 30;
  }
  
  calculateSignatureAge(lastUpdate) {
    return Math.floor(
      (Date.now() - new Date(lastUpdate)) / (24 * 60 * 60 * 1000)
    );
  }
}

// Usage
const deviceTrust = new DeviceTrustAssessment();

app.use(async (req, res, next) => {
  const device = req.deviceInfo;  // From device management agent
  
  const assessment = await deviceTrust.assessDeviceTrust(device);
  
  if (assessment.trustLevel === 'UNTRUSTED') {
    return res.status(403).json({
      error: 'Device does not meet security requirements',
      assessment,
    });
  }
  
  if (assessment.trustLevel === 'CONDITIONAL') {
    // Allow with MFA requirement
    req.requiresMFA = true;
  }
  
  next();
});
```


## Checklist

- [ ] Default deny-all NetworkPolicy implemented
- [ ] Explicit allow rules for service-to-service communication
- [ ] Cilium L7 policies for HTTP methods/paths
- [ ] mTLS enforced between all services
- [ ] Service certificates issued and managed
- [ ] Device trust assessment process implemented
- [ ] BeyondCorp device verification working
- [ ] Continuous monitoring of network traffic
- [ ] Hubble observability enabled
- [ ] Zero-trust validated against threat intelligence





## Context7 Integration

### Related Libraries & Tools
- [Cloudflare Zero Trust](/cloudflare/cloudflare-docs): Zero trust platform

### Official Documentation
- [Documentation](https://www.cloudflare.com/learning/security/glossary/what-is-zero-trust/)
- [API Reference](https://developers.cloudflare.com/cloudflare-one/)

### Version-Specific Guides
Latest stable version: Latest
- [Release Notes](https://developers.cloudflare.com/cloudflare-one/changelog/)
- [Migration Guide](https://developers.cloudflare.com/cloudflare-one/setup/)
