---
name: "moai-security-threat"
version: "4.0.0"
status: stable
description: "Enterprise Skill for advanced development"
allowed-tools: "Read, Bash, WebSearch, WebFetch"
---

# moai-security-threat: Threat Modeling & IDS/IPS Rules

**Systematic Threat Modeling with STRIDE & Network Intrusion Detection**  
Trust Score: 9.8/10 | Version: 4.0.0 | Enterprise Mode | Last Updated: 2025-11-12

---

## Overview

Threat modeling methodology using STRIDE framework combined with network-based and application-layer intrusion detection. Covers Data Flow Diagrams (DFD), attack tree analysis, vulnerability mapping, and custom IDS/IPS rule writing with Snort 3.x, Suricata 7.x, and ModSecurity 3.x.

**When to use this Skill:**
- Threat modeling new systems and architectures
- Designing security defenses against known threats
- Writing custom IDS/IPS detection rules
- Protecting web applications with WAF
- Threat intelligence integration
- Security incident response planning
- STRIDE-AI for machine learning model security

---

## Level 1: Foundations

### STRIDE Threat Model

```
S - Spoofing: Pretending to be someone else
    Example: Attacker uses stolen credentials
    Mitigation: Multi-factor authentication, digital signatures

T - Tampering: Modifying data in transit or at rest
    Example: Man-in-the-middle modifies API response
    Mitigation: Encryption, integrity checks, TLS

R - Repudiation: Denying responsibility for actions
    Example: User claims they didn't perform action
    Mitigation: Audit logging, non-repudiation tokens

I - Information Disclosure: Leaking sensitive data
    Example: Attacker reads unencrypted database
    Mitigation: Encryption, access control, data masking

D - Denial of Service (DoS): Preventing access to service
    Example: Botnet floods API with requests
    Mitigation: Rate limiting, DDoS protection, auto-scaling

E - Elevation of Privilege: Gaining higher access than permitted
    Example: Regular user becomes admin
    Mitigation: Least privilege, RBAC, privilege escalation prevention
```

### DFD Example

```
User
  ↓ [Request] → API Gateway
              ↓ [Route]
              Web Server
              ↓ [Query]
              Database
              ↓ [Return]
              Web Server
              ↓ [Response]
User ← [HTTP Response]

Threats:
- Spoofing: User with stolen credentials → need MFA
- Tampering: Request/response interception → need TLS
- I-Disclosure: Database exposed → need encryption + access control
- DoS: Gateway overloaded → need rate limiting
- Elevation: Web server compromise → need container security
```

---

## Level 2: Core Patterns

### Pattern 1: STRIDE Threat Modeling Worksheet

```javascript
class STRIDEThreatModel {
  constructor(systemName) {
    this.systemName = systemName;
    this.threats = [];
    this.mitigations = [];
  }
  
  identifyThreats(asset, category) {
    const threatPatterns = {
      'Spoofing': {
        examples: ['fake credentials', 'DNS spoofing', 'IP spoofing'],
        controls: ['MFA', 'Digital signatures', 'DNS validation'],
      },
      'Tampering': {
        examples: ['data modification', 'MITM', 'code injection'],
        controls: ['TLS', 'Checksums', 'Code signing'],
      },
      'Repudiation': {
        examples: ['deny actions', 'unauthorized transactions'],
        controls: ['Audit logging', 'Digital signatures', 'Timestamps'],
      },
      'InformationDisclosure': {
        examples: ['data leaks', 'SQL injection', 'exposed APIs'],
        controls: ['Encryption', 'Access control', 'Input validation'],
      },
      'DoS': {
        examples: ['botnet attacks', 'resource exhaustion'],
        controls: ['Rate limiting', 'DDoS protection', 'Auto-scaling'],
      },
      'ElevationOfPrivilege': {
        examples: ['privilege escalation', 'insecure defaults'],
        controls: ['RBAC', 'Least privilege', 'Secure defaults'],
      },
    };
    
    const patterns = threatPatterns[category];
    
    return {
      asset,
      category,
      examples: patterns.examples,
      suggestedControls: patterns.controls,
    };
  }
  
  createAttackTree(rootThreat) {
    // Build attack tree showing attack paths
    return {
      root: rootThreat,
      branches: [
        {
          attack: 'Compromise web server',
          subattacks: [
            'Find unpatched vulnerability',
            'Exploit via public PoC',
            'Gain RCE (Remote Code Execution)',
          ],
          likelihood: 'MEDIUM',
          impact: 'HIGH',
        },
      ],
    };
  }
  
  riskScore(threat) {
    // Risk = Likelihood × Impact × Exposure
    const likelihood = { LOW: 1, MEDIUM: 2, HIGH: 3 };
    const impact = { LOW: 1, MEDIUM: 2, HIGH: 3 };
    const exposure = threat.isNetworkAccessible ? 3 : 1;
    
    return likelihood[threat.likelihood] * impact[threat.impact] * exposure;
  }
}

// Usage
const threatModel = new STRIDEThreatModel('Payment API');

const threats = [
  threatModel.identifyThreats('Customer Payment', 'Spoofing'),
  threatModel.identifyThreats('Payment Database', 'InformationDisclosure'),
  threatModel.identifyThreats('API Endpoint', 'DoS'),
];

threats.forEach(threat => {
  console.log(`Threat: ${threat.category} on ${threat.asset}`);
  console.log(`Controls: ${threat.suggestedControls.join(', ')}`);
});
```

### Pattern 2: Snort 3.x IDS Rules

```bash
# Snort rule syntax for network-based intrusion detection
# Format: action protocol source_ip source_port -> dest_ip dest_port (options)

# Rule 1: Detect SQL injection in HTTP requests
alert http any any -> any any (
  msg:"Possible SQL Injection Attempt";
  flow:to_server,established;
  content:"GET"; http_method;
  pcre:"/\bunion\b.*\bselect\b/i";
  sid:1000001;
  rev:1;
  priority:1;
)

# Rule 2: Detect XXE (XML External Entity) attacks
alert http any any -> any any (
  msg:"XXE Attack Detected";
  content:"POST"; http_method;
  content:"Content-Type|3a|"; http_header;
  content:"xml"; http_header;
  pcre:"<!ENTITY.*SYSTEM|<!DOCTYPE.*SYSTEM/i";
  sid:1000002;
  rev:1;
  priority:1;
)

# Rule 3: Detect SSRF to AWS metadata service
alert http any any -> any any (
  msg:"SSRF to AWS Metadata Service";
  content:"GET"; http_method;
  http_uri; content:"169.254.169.254";
  sid:1000003;
  rev:1;
  priority:1;
)

# Rule 4: Detect command injection in user input
alert http any any -> any any (
  msg:"Command Injection Detected";
  content:"POST"; http_method;
  pcre:"/(;|\||&|\$\(|`).*(cat|ls|whoami|id|bash)/i";
  sid:1000004;
  rev:1;
  priority:1;
)
```

### Pattern 3: Suricata 7.x Rules (Multi-threaded)

```javascript
// Suricata configuration (YAML format, faster than Snort)
const suricataConfig = `
rules-files:
  - rule-file: /etc/suricata/rules/web-app-threats.rules
  - rule-file: /etc/suricata/rules/ssrf-detection.rules
  - rule-file: /etc/suricata/rules/injection-attacks.rules

# Suricata handles multi-core natively
threading:
  set-cpu-affinity: 'yes'
  cpu-set: "0,1,2,3,4,5,6,7"

# Output to JSON for SIEM integration
outputs:
  - eve-log:
      enabled: yes
      filetype: regular
      filename: eve.json
      types:
        - alert
        - http
        - fileinfo
        - dns
`;

class SuricataRuleManager {
  constructor() {
    this.rules = [];
  }
  
  addRule(ruleStr) {
    // Parse and add Suricata/Snort rule
    const regex = /^(alert|drop|pass|reject) (\w+) (\S+) (\S+) (->|<>) (\S+) (\S+)/;
    const match = ruleStr.match(regex);
    
    if (match) {
      this.rules.push({
        action: match[1],
        protocol: match[2],
        src_ip: match[3],
        src_port: match[4],
        dest_ip: match[6],
        dest_port: match[7],
      });
    }
  }
  
  // Enable rules for fast-movers (zero-days)
  enableEmergencyRules() {
    const rules = this.rules.filter(r => r.action === 'alert');
    console.log(`Enabled ${rules.length} alert rules for emergency response`);
  }
}
```

### Pattern 4: ModSecurity 3.x WAF Rules

```javascript
// ModSecurity WAF configuration for application layer protection
const modsecurityConfig = `
# ModSecurity 3.x rules - Application layer (Layer 7)

# SecRule - Core rule syntax
# Pattern: SecRule [variables] [@operator] "[pattern]" "id:[id],phase:[2-4],action,msg,..."

# Rule: Detect basic SQL injection
SecRule ARGS "@rx (?i:union.*select|select.*from|insert.*into)" \
    "id:100001,phase:2,deny,status:403,msg:'SQL Injection'"

# Rule: Detect XSS payload
SecRule ARGS "@rx (?i:<script|javascript:|onerror=)" \
    "id:100002,phase:2,deny,status:403,msg:'XSS Detected'"

# Rule: Command injection detection
SecRule ARGS "@rx (?i:;\\s*(cat|ls|rm|whoami|id|bash))" \
    "id:100003,phase:2,deny,status:403,msg:'Command Injection'"

# Rule: Rate limiting per IP
SecAction "id:100004,phase:1,nolog,pass,\
    setvar:ip.request_count=+1,\
    expirevar:ip.request_count=60"

SecRule ip:request_count "@gt 100" \
    "id:100005,phase:2,deny,status:429,msg:'Rate limit exceeded'"

# Rule: Detect path traversal
SecRule URI "@rx \\.\\./|\\.\\.%2f" \
    "id:100006,phase:2,deny,status:403,msg:'Path Traversal'"

# Rule: Protect sensitive endpoints
SecRule REQUEST_URI "@rx /admin|/internal|/api/secret" \
    "id:100007,phase:2,require:secure,msg:'HTTPS Required for sensitive endpoint'"
`;

// ModSecurity implementation
class ModSecurityWAF {
  constructor() {
    this.rules = [];
  }
  
  applyRules(request) {
    const violations = [];
    
    // Check each rule
    for (const rule of this.rules) {
      const matches = this.matchRule(request, rule);
      if (matches) {
        violations.push({
          ruleId: rule.id,
          message: rule.msg,
          action: rule.action,  // deny, block, log, etc.
        });
      }
    }
    
    return violations;
  }
  
  matchRule(request, rule) {
    // Simplified rule matching
    if (rule.pattern === 'SQL_INJECTION') {
      return /union.*select|insert.*into/i.test(
        JSON.stringify(request.args)
      );
    }
    
    return false;
  }
}
```

---

## Level 3: Advanced

### Advanced: Context7 MCP Threat Intelligence

```javascript
const { Context7Client } = require('context7-mcp');

class ThreatIntelligenceIntegration {
  constructor(apiKey) {
    this.context7 = new Context7Client(apiKey);
    this.threatCache = new Map();
  }
  
  // Query threat intelligence for IOCs (Indicators of Compromise)
  async checkThreat(ioc) {
    const cacheKey = ioc.hash || ioc.ip || ioc.domain;
    
    if (this.threatCache.has(cacheKey)) {
      return this.threatCache.get(cacheKey);
    }
    
    const threat = await this.context7.query({
      type: 'ioc_reputation',
      ...ioc,
      tags: ['malware', 'c2_server', 'botnet'],
    });
    
    // Cache for 24 hours
    this.threatCache.set(cacheKey, threat);
    setTimeout(
      () => this.threatCache.delete(cacheKey),
      86400000
    );
    
    return threat;
  }
  
  // Enrich IDS alerts with threat context
  async enrichAlert(alert) {
    const threat = await this.checkThreat({
      ip: alert.sourceIp,
      domain: alert.domain,
    });
    
    return {
      ...alert,
      threatScore: threat.severity,
      threatIndicators: threat.indicators,
      recommendedAction: this.getAction(threat.severity),
    };
  }
  
  getAction(severity) {
    const actions = {
      'CRITICAL': 'BLOCK_IMMEDIATELY',
      'HIGH': 'BLOCK_SESSION',
      'MEDIUM': 'ALERT_SOC',
      'LOW': 'LOG_ONLY',
    };
    return actions[severity] || 'LOG_ONLY';
  }
}
```

---

## Checklist

- [ ] STRIDE threat model created
- [ ] Data Flow Diagram (DFD) documented
- [ ] Attack trees for critical assets created
- [ ] Snort rules compiled and tested
- [ ] Suricata rules deployed on network
- [ ] ModSecurity WAF rules applied
- [ ] IDS/IPS tuning completed (false positive reduction)
- [ ] Threat intelligence integrated (Context7)
- [ ] Alert correlation configured
- [ ] Incident response plan documented

---

## Quick Reference

| Tool | Layer | Use Case |
|------|-------|----------|
| Snort 3.x | Network (L3/4) | IDS/IPS |
| Suricata 7.x | Network (L3/4) | IDS/IPS (multi-core) |
| ModSecurity | Application (L7) | WAF |
| Context7 | Intelligence | Threat enrichment |

