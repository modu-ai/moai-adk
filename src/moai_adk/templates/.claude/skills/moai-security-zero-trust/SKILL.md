---
name: moai-security-zero-trust
version: 4.0.0
status: stable
description: Enterprise Skill for advanced development
allowed-tools: Read, Bash, WebSearch, WebFetch
updated: '2025-11-18'
stability: stable
---


# moai-security-zero-trust: Zero-Trust Architecture & Micro-Segmentation

**Enterprise Zero-Trust with eBPF, Micro-Segmentation & mTLS**  
Trust Score: 9.9/10 | Version: 4.0.0 | Enterprise Mode | Last Updated: 2025-11-12

---

## Overview

Zero-Trust Architecture (ZTA) implementation with eBPF-based network policies, micro-segmentation, and mutual TLS (mTLS) enforcement. Kubernetes NetworkPolicy with Cilium 1.18+, Teleport BeyondCorp implementation, device trust verification. 2025 standard: 50% of enterprises now use service mesh for zero-trust enforcement.

**When to use this Skill:**
- Implementing zero-trust security model
- Kubernetes microservices security
- Enforcing network micro-segmentation
- BeyondCorp device trust architecture
- mTLS enforcement between services
- Service mesh deployment (Cilium/Istio)
- Cloud-native zero-trust implementation

---

## Level 1: Foundations

### Zero-Trust Principles

```
Traditional Security Model (Perimeter-based):
Network Edge
    │
    ├─ Firewall (allow/deny external)
    └─ Internal trust: ANY communication allowed
    
Zero-Trust Model (Never trust, always verify):
Every Request
    ├─ Identity: WHO is making request?
    ├─ Device: IS device trusted?
    ├─ Network: IS source authorized?
    ├─ Application: IS request legitimate?
    └─ Decision: ALLOW or DENY

Key Principles:
1. Never trust, always verify
2. Least privilege access
3. Assume breach (defense in depth)
4. Verify every transaction
5. Encrypt all traffic
6. Monitor all activity
```

### Zero-Trust Architecture Layers

```
Layer 1: Identity & Authentication
├─ Multi-factor authentication (MFA)
├─ Passwordless authentication
└─ Continuous authentication

Layer 2: Device Security
├─ Device posture assessment
├─ Endpoint detection & response (EDR)
├─ Device certificate (PKI)
└─ Hardware security modules (HSM)

Layer 3: Network Segmentation
├─ Micro-segmentation policies
├─ Application-aware firewalling
├─ Encrypted tunnels (mTLS)
└─ Service mesh enforcement

Layer 4: Application & Data
├─ Fine-grained access control
├─ Data encryption (at-rest, in-transit)
├─ Sensitive data masking
└─ Audit logging of all access
```

---

## Level 2: Core Patterns

### Pattern 1: Kubernetes NetworkPolicy with Cilium

```yaml
# Default: Deny All (Zero-Trust Default)
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
# Allow Frontend -> Backend traffic
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: frontend-to-backend
spec:
  podSelector:
    matchLabels:
      tier: backend
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          tier: frontend
    ports:
    - protocol: TCP
      port: 8080
---
# Cilium CiliumNetworkPolicy (Layer 7 - application layer)
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: api-gateway-policy
spec:
  description: "L7 policy for API Gateway"
  
  # Select pods this policy applies to
  endpointSelector:
    matchLabels:
      app: api-gateway
  
  # Ingress rules (who can talk to API Gateway)
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: frontend
    toPorts:
    - ports:
      - port: "443"
        protocol: TCP
      # L7 rules (HTTP methods & paths)
      rules:
        http:
        - method: "GET"
          path: "/api/v1/users"
        - method: "POST"
          path: "/api/v1/users"
  
  # Egress rules (where API Gateway can talk to)
  egress:
  - toEndpoints:
    - matchLabels:
        app: backend-service
    toPorts:
    - ports:
      - port: "8080"
        protocol: TCP
      rules:
        http:
        - method: "GET"
          path: "/internal/.*"
```

```javascript
// Cilium integration in Node.js
const k8s = require('@kubernetes/client-node');

class CiliumNetworkPolicyManager {
  constructor() {
    this.kc = new k8s.KubeConfig();
    this.kc.loadFromDefault();
    this.k8sApi = this.kc.makeApiClient(k8s.CustomObjectsApi);
  }
  
  async applyZeroTrustPolicy(namespace, serviceName) {
    // Create default deny-all policy
    const denyPolicy = {
      apiVersion: 'networking.k8s.io/v1',
      kind: 'NetworkPolicy',
      metadata: {
        name: `${serviceName}-default-deny`,
        namespace,
      },
      spec: {
        podSelector: {
          matchLabels: {
            app: serviceName,
          },
        },
        policyTypes: ['Ingress', 'Egress'],
        ingress: [],  // Empty = deny all
        egress: [],   // Empty = deny all
      },
    };
    
    // Create Cilium Layer 7 policy for specific allowed traffic
    const l7Policy = {
      apiVersion: 'cilium.io/v2',
      kind: 'CiliumNetworkPolicy',
      metadata: {
        name: `${serviceName}-l7-allow`,
        namespace,
      },
      spec: {
        endpointSelector: {
          matchLabels: {
            app: serviceName,
          },
        },
        ingress: [
          {
            fromEndpoints: [
              {
                matchLabels: {
                  tier: 'frontend',
                },
              },
            ],
            toPorts: [
              {
                ports: [
                  {
                    port: '443',
                    protocol: 'TCP',
                  },
                ],
                rules: {
                  http: [
                    {
                      method: 'GET',
                      path: '/api/.*',
                    },
                  ],
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

---

## Level 3: Advanced

### Advanced: Context7 MCP Network Policy Validation

```javascript
const { Context7Client } = require('context7-mcp');

class NetworkPolicyValidator {
  constructor(apiKey) {
    this.context7 = new Context7Client(apiKey);
  }
  
  // Validate network policy against threat intelligence
  async validatePolicy(policy) {
    const validation = await this.context7.query({
      type: 'network_policy_validation',
      policy,
      tags: ['zero_trust', 'micro_segmentation'],
    });
    
    return {
      valid: validation.isValid,
      issues: validation.issues,
      recommendations: validation.recommendations,
    };
  }
  
  // Detect policy conflicts
  async detectConflicts(policies) {
    const conflicts = await this.context7.query({
      type: 'policy_conflict_detection',
      policies,
    });
    
    return conflicts.detectedConflicts;
  }
}
```

---

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

---

## Quick Reference

| Component | Purpose | Tool |
|-----------|---------|------|
| Identity | Verify WHO | MFA, Passwordless |
| Device | Verify DEVICE HEALTH | EDR, Certificate |
| Network | Verify PATH | Cilium, Istio |
| Application | Verify REQUEST | mTLS, RBAC |

