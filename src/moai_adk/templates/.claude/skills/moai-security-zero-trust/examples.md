# Zero-Trust Architecture - Production Code Examples

## Example 1: Kubernetes Zero-Trust Policies

```yaml
# k8s/zero-trust-policies.yaml
---
# Default deny all traffic
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: production
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
---
# Allow frontend to backend only
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: frontend-to-backend
  namespace: production
spec:
  description: "Frontend to Backend communication"
  
  endpointSelector:
    matchLabels:
      app: backend
  
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: frontend
    toPorts:
    - ports:
      - port: "8080"
        protocol: TCP
      rules:
        http:
        - method: "GET"
          path: "/api/v1/.*"
        - method: "POST"
          path: "/api/v1/data"
```

## Example 2: Device Trust Assessment

```javascript
// security/device-trust-assessment.js
class DeviceTrustAssessment {
  async assessDevice(device) {
    let trustScore = 0;
    const checks = {};
    
    // Check 1: OS Patching
    const osPatched = this.isOSPatched(device);
    checks.os_patched = osPatched;
    trustScore += osPatched ? 25 : 0;
    
    // Check 2: Antivirus
    const avEnabled = device.antivirus_enabled;
    checks.antivirus = avEnabled;
    trustScore += avEnabled ? 25 : 0;
    
    // Check 3: Firewall
    const fwEnabled = device.firewall_enabled;
    checks.firewall = fwEnabled;
    trustScore += fwEnabled ? 25 : 0;
    
    // Check 4: Disk Encryption
    const encrypted = device.disk_encrypted;
    checks.encryption = encrypted;
    trustScore += encrypted ? 25 : 0;
    
    return {
      deviceId: device.id,
      trustScore,
      trustLevel: this.calculateTrustLevel(trustScore),
      checks,
    };
  }
  
  calculateTrustLevel(score) {
    if (score >= 90) return 'TRUSTED';
    if (score >= 70) return 'CONDITIONAL';
    return 'UNTRUSTED';
  }
  
  isOSPatched(device) {
    const daysSincePatch = Math.floor(
      (Date.now() - new Date(device.last_patch_date)) / (86400000)
    );
    return daysSincePatch <= 30;  // Patched within 30 days
  }
}

module.exports = DeviceTrustAssessment;
```

## Example 3: mTLS Certificate Management

```javascript
// security/mtls-manager.js
const tls = require('tls');
const fs = require('fs');

class mTLSManager {
  constructor() {
    this.certificates = new Map();
  }
  
  // Create mTLS server
  createSecureServer(port, options) {
    const tlsOptions = {
      key: fs.readFileSync(options.keyPath),
      cert: fs.readFileSync(options.certPath),
      ca: fs.readFileSync(options.caPath),
      requestCert: true,  // Require client certificate
      rejectUnauthorized: true,  // Reject unsigned certificates
    };
    
    const server = tls.createSecureContext(tlsOptions);
    
    return server;
  }
  
  // Verify client certificate
  async verifyClientCert(cert) {
    // Check certificate validity
    if (new Date() > new Date(cert.valid_to)) {
      throw new Error('Certificate expired');
    }
    
    // Check certificate in trusted store
    if (!this.isTrusted(cert)) {
      throw new Error('Certificate not trusted');
    }
    
    return {
      clientName: cert.subject.CN,
      trusted: true,
    };
  }
  
  isTrusted(cert) {
    // Check if certificate is signed by trusted CA
    return true;  // Implementation detail
  }
}

module.exports = mTLSManager;
```

## Example 4: Context7 MCP Policy Validation

```javascript
// compliance/context7-policy-validator.js
const { Context7Client } = require('context7-mcp');

class PolicyValidator {
  constructor(apiKey) {
    this.context7 = new Context7Client({ apiKey });
  }
  
  async validateNetworkPolicy(policy) {
    const validation = await this.context7.query({
      type: 'network_policy_validation',
      policy,
      tags: ['zero_trust', 'kubernetes'],
    });
    
    return {
      valid: validation.isValid,
      issues: validation.issues,
      recommendations: validation.recommendations,
    };
  }
  
  async checkPolicyConflicts(policies) {
    return await this.context7.query({
      type: 'policy_conflict_detection',
      policies,
    });
  }
}

module.exports = PolicyValidator;
```

## Example 5: Testing Zero-Trust Policies

```javascript
// tests/zero-trust-policies.test.js
const k8s = require('@kubernetes/client-node');

describe('Zero-Trust Network Policies', () => {
  test('default deny all policy is in place', async () => {
    const policies = await k8s.listNamespacedNetworkPolicy('production');
    
    const defaultDeny = policies.items.find(
      p => p.metadata.name === 'default-deny-all'
    );
    
    expect(defaultDeny).toBeDefined();
    expect(defaultDeny.spec.ingress).toEqual([]);  // No ingress allowed by default
    expect(defaultDeny.spec.egress).toEqual([]);  // No egress allowed by default
  });
  
  test('frontend to backend policy allows specific traffic', async () => {
    const policies = await k8s.listNamespacedNetworkPolicy('production');
    
    const feToBe = policies.items.find(
      p => p.metadata.name === 'frontend-to-backend'
    );
    
    expect(feToBe).toBeDefined();
    expect(feToBe.spec.ingress[0].fromEndpoints[0].matchLabels.app).toBe('frontend');
  });
});
```

