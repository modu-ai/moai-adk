# Threat Modeling & IDS/IPS - Production Code Examples

## Example 1: STRIDE Threat Analysis

```javascript
// threat-modeling/stride-analyzer.js
class STRIDEAnalyzer {
  constructor() {
    this.threats = [];
    this.mitigations = [];
  }
  
  analyzeThreat(asset, category, description) {
    const threatObject = {
      asset,
      category,
      description,
      likelihood: 'MEDIUM',
      impact: 'MEDIUM',
      riskScore: 0,
      mitigations: [],
    };
    
    // Calculate risk score
    threatObject.riskScore = this.calculateRiskScore(
      threatObject.likelihood,
      threatObject.impact
    );
    
    // Get suggested mitigations
    threatObject.mitigations = this.getSuggestedMitigations(category);
    
    this.threats.push(threatObject);
    return threatObject;
  }
  
  calculateRiskScore(likelihood, impact) {
    const likelihood_map = { LOW: 1, MEDIUM: 2, HIGH: 3 };
    const impact_map = { LOW: 1, MEDIUM: 2, HIGH: 3, CRITICAL: 4 };
    
    return likelihood_map[likelihood] * impact_map[impact];
  }
  
  getSuggestedMitigations(category) {
    const mitigations = {
      'Spoofing': ['MFA', 'Digital signatures', 'SAML/OIDC'],
      'Tampering': ['TLS encryption', 'Code signing', 'Integrity checks'],
      'Repudiation': ['Audit logging', 'Non-repudiation tokens'],
      'InformationDisclosure': ['Encryption', 'Access control', 'Data masking'],
      'DoS': ['Rate limiting', 'DDoS protection', 'Auto-scaling'],
      'ElevationOfPrivilege': ['RBAC', 'Least privilege', 'Privilege escalation detection'],
    };
    
    return mitigations[category] || [];
  }
  
  generateReport() {
    return {
      threatCount: this.threats.length,
      threats: this.threats.sort((a, b) => b.riskScore - a.riskScore),
      summary: this.generateSummary(),
    };
  }
  
  generateSummary() {
    const critical = this.threats.filter(t => t.riskScore >= 9).length;
    const high = this.threats.filter(t => t.riskScore >= 6 && t.riskScore < 9).length;
    
    return {
      critical,
      high,
      total: this.threats.length,
    };
  }
}

module.exports = STRIDEAnalyzer;
```

## Example 2: Snort Rules Configuration

```bash
# /etc/suricata/rules/custom-rules.rules
# Suricata rules for custom threats

# Rule 1: Detect SSRF to metadata service
alert http any any -> any any (
  msg:"SSRF AWS Metadata Service";
  content:"GET"; http_method;
  http_uri; content:"169.254.169.254";
  sid:2100001;
  rev:1;
  priority:1;
)

# Rule 2: Detect SQL injection patterns
alert http any any -> any any (
  msg:"Potential SQL Injection";
  content:"POST"; http_method;
  pcre:"/(union.*select|insert.*into|delete.*from)/i";
  sid:2100002;
  rev:1;
  priority:1;
)

# Rule 3: Detect XXE attacks
alert http any any -> any any (
  msg:"XXE Attack Detected";
  content:"Content-Type|3a|"; http_header;
  content:"xml"; http_header;
  pcre:"/<!ENTITY.*SYSTEM|<!DOCTYPE.*SYSTEM/i";
  sid:2100003;
  rev:1;
  priority:1;
)

# Rule 4: Detect command injection
alert http any any -> any any (
  msg:"Command Injection";
  content:"POST"; http_method;
  pcre:"/(;|&&|\|\||`|\$\().*?(cat|ls|whoami|bash|sh)/i";
  sid:2100004;
  rev:1;
  priority:1;
)
```

## Example 3: ModSecurity WAF Integration

```javascript
// waf/modsecurity-config.js
const ModSecurity = require('modsecurity');

class ModSecurityWAF {
  constructor() {
    this.waf = new ModSecurity();
    this.rules = [];
    this.loadCoreRules();
  }
  
  loadCoreRules() {
    // Load OWASP Core Rule Set (CRS)
    this.waf.addRules([
      // SQL Injection
      {
        id: 942100,
        msg: 'SQL Injection Attack',
        pattern: '/(union.*select|insert.*into)/i',
        action: 'block',
      },
      
      // XSS
      {
        id: 941100,
        msg: 'XSS Attack',
        pattern: /(script|onerror|onclick|javascript:)/i,
        action: 'block',
      },
      
      // RCE
      {
        id: 933100,
        msg: 'Remote Code Execution',
        pattern: /(exec|system|passthru|shell_exec|proc_open)/i,
        action: 'block',
      },
    ]);
  }
  
  processRequest(req) {
    const violations = this.waf.processRequest(req);
    
    if (violations.length > 0) {
      console.log(`WAF blocked request: ${violations[0].msg}`);
      return {
        blocked: true,
        violations,
      };
    }
    
    return { blocked: false };
  }
}

module.exports = ModSecurityWAF;
```

## Example 4: IDS Alert Correlation

```javascript
// ids-ips/alert-correlator.js
class AlertCorrelator {
  constructor() {
    this.alerts = [];
    this.correlationWindow = 300000;  // 5 minutes
  }
  
  addAlert(alert) {
    this.alerts.push({
      ...alert,
      timestamp: new Date(),
    });
    
    this.correlateAlerts();
  }
  
  correlateAlerts() {
    // Group alerts by source IP within correlation window
    const now = new Date();
    const recentAlerts = this.alerts.filter(
      a => now - a.timestamp < this.correlationWindow
    );
    
    const bySourceIp = {};
    
    for (const alert of recentAlerts) {
      if (!bySourceIp[alert.sourceIp]) {
        bySourceIp[alert.sourceIp] = [];
      }
      bySourceIp[alert.sourceIp].push(alert);
    }
    
    // Detect attack patterns
    for (const [sourceIp, ipAlerts] of Object.entries(bySourceIp)) {
      if (ipAlerts.length >= 3) {
        this.escalateAttack(sourceIp, ipAlerts);
      }
    }
  }
  
  escalateAttack(sourceIp, alerts) {
    console.warn(`ATTACK PATTERN DETECTED from ${sourceIp}:`);
    alerts.forEach(a => console.warn(`  - ${a.message}`));
    
    // Trigger automated response
    this.blockSourceIp(sourceIp);
  }
  
  blockSourceIp(sourceIp) {
    // Update firewall to block this IP
    console.log(`Blocking ${sourceIp} in firewall`);
  }
}

module.exports = AlertCorrelator;
```

## Example 5: Context7 MCP Threat Intelligence

```javascript
// threat-intelligence/context7-threats.js
const { Context7Client } = require('context7-mcp');

class ThreatIntelligence {
  constructor(apiKey) {
    this.context7 = new Context7Client({ apiKey });
  }
  
  async enrichAlert(alert) {
    const threat = await this.context7.query({
      type: 'ioc_reputation',
      ip: alert.sourceIp,
      domain: alert.domain,
      tags: ['malware', 'botnet', 'c2'],
    });
    
    return {
      ...alert,
      threatScore: threat.severity,
      indicators: threat.indicators,
      recommendation: this.getRecommendation(threat.severity),
    };
  }
  
  getRecommendation(severity) {
    const recs = {
      'CRITICAL': 'BLOCK_IMMEDIATELY',
      'HIGH': 'BLOCK_AND_ALERT',
      'MEDIUM': 'MONITOR_CLOSELY',
      'LOW': 'LOG_ONLY',
    };
    
    return recs[severity] || 'LOG_ONLY';
  }
}

module.exports = ThreatIntelligence;
```

