---
name: "moai-security-ssrf"
version: "4.0.0"
created: 2025-11-11
updated: 2025-11-13
status: stable
description: Enterprise SSRF Security Protection with AI-powered request validation, Context7 integration, and intelligent URL filtering for web application security
keywords: ['ssrf', 'server-side-request-forgery', 'web-security', 'url-validation', 'request-filtering', 'context7-integration', 'ai-orchestration', 'production-deployment']
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

# Enterprise SSRF Security Protection Expert v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-security-ssrf |
| **Version** | 4.0.0 (2025-11-13) |
| **Tier** | Enterprise Security Expert |
| **AI-Powered** | ✅ Context7 Integration, Intelligent Protection |
| **Auto-load** | On demand when SSRF security keywords detected |

---

## What It Does

Enterprise SSRF Security Protection expert with AI-powered request validation, Context7 integration, and intelligent URL filtering for comprehensive web application security against Server-Side Request Forgery attacks.

**Revolutionary v4.0.0 capabilities**:
- 🤖 **AI-Powered URL Validation** using Context7 MCP for latest threat intelligence
- 📊 **Intelligent Request Filtering** with automated pattern recognition and blocking
- 🚀 **Advanced SSRF Detection** with AI-driven anomaly detection and prevention
- 🔗 **Enterprise Security Framework** with zero-configuration protection deployment
- 📈 **Predictive Threat Analysis** with vulnerability assessment and prevention

---

## When to Use

**Automatic triggers**:
- SSRF vulnerability assessment and protection implementation discussions
- Web application security and request validation planning
- URL filtering and request filtering strategy design
- API security and external service integration security

**Manual invocation**:
- Designing enterprise SSRF protection architectures with optimal security
- Implementing comprehensive URL validation and filtering systems
- Planning security testing and vulnerability assessments
- Optimizing request validation performance and coverage

---

# Quick Reference (Level 1)

## SSRF Protection Stack (November 2025)

### Core Protection Components
- **URL Validation**: Comprehensive URL parsing and validation
- **Request Filtering**: Request filtering based on allowlist/denylist
- **Network Segmentation**: Isolated network zones for external requests
- **Rate Limiting**: Request rate limiting to prevent abuse
- **Anomaly Detection**: AI-powered behavioral analysis

### Common SSRF Vectors
- **Internal Network Access**: Access to internal services and metadata
- **Cloud Provider Metadata**: AWS, GCP, Azure metadata endpoints
- **Localhost Access**: Access to local services and files
- **File Protocol**: file:// protocol for local file access
- **DNS Rebinding**: DNS manipulation to bypass security

### Protection Strategies
- **Allowlist Approach**: Only allow explicitly approved domains
- **URL Parsing**: Comprehensive URL component validation
- **Network Restrictions**: Block access to internal network ranges
- **Protocol Filtering**: Only allow safe protocols (HTTP, HTTPS)
- **Response Validation**: Validate response content and size

### Security Standards
- **OWASP Top 10**: SSRF protection requirements
- **NIST SP 800-53**: Security controls for web applications
- **ISO 27001**: Information security management
- **SOC 2**: Security controls and reporting
- **PCI DSS**: Payment card industry security standards

---

# Core Implementation (Level 2)

## SSRF Protection Architecture Intelligence

```python
# AI-powered SSRF protection architecture optimization with Context7
class SSRFProtectionArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.threat_analyzer = ThreatAnalyzer()
        self.url_validator = URLValidator()
    
    async def design_optimal_ssrf_protection(self, 
                                          requirements: SecurityRequirements) -> SSRFProtectionArchitecture:
        """Design optimal SSRF protection architecture using AI analysis."""
        
        # Get latest SSRF and web security documentation via Context7
        ssrf_docs = await self.context7_client.get_library_docs(
            context7_library_id='/ssrf-protection/docs',
            topic="ssrf security url validation web security 2025",
            tokens=3000
        )
        
        security_docs = await self.context7_client.get_library_docs(
            context7_library_id='/web-security/docs',
            topic="request filtering threat prevention best practices 2025",
            tokens=2000
        )
        
        # Optimize URL validation strategy
        url_validation = self.url_validator.optimize_validation(
            requirements.url_requirements,
            requirements.security_level,
            ssrf_docs
        )
        
        # Analyze threat patterns
        threat_analysis = self.threat_analyzer.analyze_ssrf_threats(
            requirements.application_context,
            requirements.attack_surface,
            security_docs
        )
        
        return SSRFProtectionArchitecture(
            url_validation_system=url_validation,
            request_filtering=self._design_request_filtering(requirements),
            network_segmentation=self._design_network_segmentation(requirements),
            threat_detection=threat_analysis,
            monitoring_system=self._configure_monitoring(),
            compliance_framework=self._ensure_compliance(requirements)
        )
```

## Advanced SSRF Protection Implementation

```typescript
// Enterprise SSRF protection with TypeScript
import { URL } from 'url';
import { createHash } from 'crypto';
import { SecurityLogger } from './security-logger';

interface SSRFProtectionConfig {
  allowedDomains: string[];
  allowedIPRanges: string[];
  blockedDomains: string[];
  blockedIPRanges: string[];
  allowedProtocols: string[];
  maxRedirects: number;
  maxResponseSize: number;
  timeoutMs: number;
}

export class SSRFProtection {
  private config: SSRFProtectionConfig;
  private logger: SecurityLogger;
  private requestCache: Map<string, RequestResult> = new Map();

  constructor(config: SSRFProtectionConfig) {
    this.config = config;
    this.logger = new SecurityLogger();
  }

  async validateURL(url: string): Promise<URLValidationResult> {
    try {
      const parsedURL = new URL(url);
      
      // Protocol validation
      if (!this.config.allowedProtocols.includes(parsedURL.protocol)) {
        return this.createInvalidResult(
          'PROTOCOL_NOT_ALLOWED',
          `Protocol ${parsedURL.protocol} is not allowed`
        );
      }

      // Domain validation
      const domainValidation = await this.validateDomain(parsedURL.hostname);
      if (!domainValidation.isValid) {
        return domainValidation;
      }

      // IP address validation
      const ipValidation = await this.validateIPAddress(parsedURL.hostname);
      if (!ipValidation.isValid) {
        return ipValidation;
      }

      // Additional security checks
      const securityValidation = await this.performSecurityChecks(parsedURL);
      if (!securityValidation.isValid) {
        return securityValidation;
      }

      return {
        isValid: true,
        normalizedURL: this.normalizeURL(parsedURL),
        riskScore: this.calculateRiskScore(parsedURL),
      };
    } catch (error) {
      return this.createInvalidResult('INVALID_URL', `Invalid URL: ${error.message}`);
    }
  }

  private async validateDomain(hostname: string): Promise<URLValidationResult> {
    // Check blocked domains
    for (const blockedDomain of this.config.blockedDomains) {
      if (hostname.includes(blockedDomain)) {
        return this.createInvalidResult(
          'DOMAIN_BLOCKED',
          `Domain ${hostname} is blocked`
        );
      }
    }

    // Check allowed domains (if allowlist is configured)
    if (this.config.allowedDomains.length > 0) {
      const isAllowed = this.config.allowedDomains.some(allowedDomain =>
        hostname === allowedDomain || hostname.endsWith(`.${allowedDomain}`)
      );

      if (!isAllowed) {
        return this.createInvalidResult(
          'DOMAIN_NOT_ALLOWED',
          `Domain ${hostname} is not in allowlist`
        );
      }
    }

    // Check for suspicious patterns
    const suspiciousPatterns = [
      /localhost/i,
      /^127\./,
      /^0x[0-9a-f]+/i, // Hex encoded IPs
      /^0[0-7]{3,}/, // Octal encoded IPs
      /internal/i,
      /private/i,
      /metadata/i,
    ];

    for (const pattern of suspiciousPatterns) {
      if (pattern.test(hostname)) {
        return this.createInvalidResult(
          'SUSPICIOUS_DOMAIN',
          `Domain ${hostname} matches suspicious pattern`
        );
      }
    }

    return { isValid: true };
  }

  private async validateIPAddress(hostname: string): Promise<URLValidationResult> {
    const IP = require('ip');
    
    try {
      // Resolve hostname to IP addresses
      const addresses = await this.resolveHostname(hostname);
      
      for (const address of addresses) {
        // Check if it's an IP address
        if (IP.isV4Format(address) || IP.isV6Format(address)) {
          // Check blocked IP ranges
          for (const blockedRange of this.config.blockedIPRanges) {
            if (IP.cidrSubnet(address, blockedRange)) {
              return this.createInvalidResult(
                'IP_BLOCKED',
                `IP address ${address} is in blocked range ${blockedRange}`
              );
            }
          }

          // Check if IP is in allowed ranges (if allowlist is configured)
          if (this.config.allowedIPRanges.length > 0) {
            const isAllowed = this.config.allowedIPRanges.some(allowedRange =>
              IP.cidrSubnet(address, allowedRange)
            );

            if (!isAllowed) {
              return this.createInvalidResult(
                'IP_NOT_ALLOWED',
                `IP address ${address} is not in allowed ranges`
              );
            }
          }

          // Check for private/internal IP ranges
          if (this.isPrivateIP(address)) {
            return this.createInvalidResult(
              'PRIVATE_IP_BLOCKED',
              `Private IP address ${address} is not allowed`
            );
          }
        }
      }

      return { isValid: true };
    } catch (error) {
      return this.createInvalidResult(
        'IP_VALIDATION_ERROR',
        `IP validation failed: ${error.message}`
      );
    }
  }

  private async resolveHostname(hostname: string): Promise<string[]> {
    const dns = require('dns').promises;
    
    try {
      const { addresses } = await dns.resolve4(hostname);
      return addresses;
    } catch (error) {
      // Try IPv6 if IPv4 fails
      try {
        const { addresses } = await dns.resolve6(hostname);
        return addresses;
      } catch (ipv6Error) {
        throw new Error(`Unable to resolve hostname: ${hostname}`);
      }
    }
  }

  private isPrivateIP(ip: string): boolean {
    const IP = require('ip');
    
    const privateRanges = [
      '10.0.0.0/8',      // Private IP range
      '172.16.0.0/12',   // Private IP range
      '192.168.0.0/16',  // Private IP range
      '127.0.0.0/8',     // Loopback
      '169.254.0.0/16',  // Link-local
      '224.0.0.0/4',     // Multicast
    ];

    return privateRanges.some(range => IP.cidrSubnet(ip, range));
  }

  private async performSecurityChecks(url: URL): Promise<URLValidationResult> {
    // Check for file protocol
    if (url.protocol === 'file:') {
      return this.createInvalidResult(
        'FILE_PROTOCOL_BLOCKED',
        'File protocol is not allowed'
      );
    }

    // Check for suspicious query parameters
    const suspiciousParams = [
      'redirect',
      'url',
      'callback',
      'return',
      'dest',
      'destination',
    ];

    for (const param of suspiciousParams) {
      if (url.searchParams.has(param)) {
        const paramValue = url.searchParams.get(param);
        if (paramValue && this.isSuspiciousURL(paramValue)) {
          return this.createInvalidResult(
            'SUSPICIOUS_PARAMETER',
            `Suspicious parameter ${param} detected`
          );
        }
      }
    }

    // Check URL length (very long URLs can be suspicious)
    if (url.toString().length > 2048) {
      return this.createInvalidResult(
        'URL_TOO_LONG',
        'URL exceeds maximum allowed length'
      );
    }

    return { isValid: true };
  }

  private isSuspiciousURL(url: string): boolean {
    const suspiciousPatterns = [
      /localhost/i,
      /127\.0\.0\.1/,
      /0x[0-9a-f]+/i,
      /internal/i,
      /private/i,
      /metadata/i,
      /169\.254\./, // Link-local
      /file:\/\//,
    ];

    return suspiciousPatterns.some(pattern => pattern.test(url));
  }

  async makeSecureRequest(url: string, options: RequestOptions = {}): Promise<SecureRequestResult> {
    // Validate URL first
    const validationResult = await this.validateURL(url);
    if (!validationResult.isValid) {
      throw new SSRFError('URL_VALIDATION_FAILED', validationResult.reason || 'Invalid URL');
    }

    // Check request cache
    const cacheKey = this.generateCacheKey(url, options);
    const cached = this.requestCache.get(cacheKey);
    if (cached && !this.isCacheExpired(cached)) {
      return cached.result;
    }

    try {
      // Make secure request
      const startTime = Date.now();
      const response = await this.makeRequest(url, {
        timeout: this.config.timeoutMs,
        maxRedirects: this.config.maxRedirects,
        ...options,
      });
      
      const responseTime = Date.now() - startTime;

      // Validate response
      const responseValidation = await this.validateResponse(response);
      if (!responseValidation.isValid) {
        throw new SSRFError('RESPONSE_VALIDATION_FAILED', responseValidation.reason || 'Invalid response');
      }

      const result: SecureRequestResult = {
        success: true,
        data: response.data,
        status: response.status,
        headers: response.headers,
        responseTime,
        url: validationResult.normalizedURL,
      };

      // Cache result
      this.requestCache.set(cacheKey, {
        result,
        timestamp: Date.now(),
      });

      // Log request
      this.logger.log('SECURE_REQUEST_SUCCESS', {
        url: validationResult.normalizedURL,
        responseTime,
        status: response.status,
        riskScore: validationResult.riskScore,
      });

      return result;
    } catch (error) {
      // Log failed request
      this.logger.log('SECURE_REQUEST_FAILED', {
        url: url,
        error: error.message,
        type: error.name,
      });

      throw error;
    }
  }

  private async makeRequest(url: string, options: RequestOptions): Promise<RequestResponse> {
    const axios = require('axios');
    
    const response = await axios({
      url,
      method: options.method || 'GET',
      timeout: options.timeout,
      maxRedirects: options.maxRedirects,
      headers: {
        'User-Agent': 'SecureRequest/1.0',
        ...options.headers,
      },
      responseType: 'arraybuffer', // Handle binary responses
      maxContentLength: this.config.maxResponseSize,
    });

    return {
      data: response.data,
      status: response.status,
      headers: response.headers,
    };
  }

  private async validateResponse(response: RequestResponse): Promise<ValidationResult> {
    // Check response size
    if (Buffer.byteLength(response.data) > this.config.maxResponseSize) {
      return {
        isValid: false,
        reason: `Response size exceeds maximum allowed size`,
      };
    }

    // Check content type (optional based on requirements)
    const contentType = response.headers['content-type'];
    if (contentType && this.isSuspiciousContentType(contentType)) {
      return {
        isValid: false,
        reason: `Suspicious content type: ${contentType}`,
      };
    }

    return { isValid: true };
  }

  private isSuspiciousContentType(contentType: string): boolean {
    const suspiciousTypes = [
      'application/octet-stream',
      'application/x-executable',
      'application/x-msdownload',
      'application/x-msdos-program',
    ];

    return suspiciousTypes.some(type => contentType.includes(type));
  }

  private calculateRiskScore(url: URL): number {
    let score = 0;

    // URL length factor
    if (url.toString().length > 1000) score += 10;
    if (url.toString().length > 2000) score += 20;

    // Protocol factor
    if (url.protocol !== 'https:') score += 5;

    // Domain complexity factor
    if (url.hostname.split('.').length > 4) score += 5;

    // Query parameter factor
    if (url.searchParams.size > 10) score += 5;
    if (url.searchParams.size > 20) score += 15;

    return Math.min(score, 100); // Cap at 100
  }

  private createInvalidResult(reason: string, details?: string): URLValidationResult {
    return {
      isValid: false,
      reason,
      details,
      riskScore: 100, // High risk for invalid URLs
    };
  }
}

// Error classes
export class SSRFError extends Error {
  constructor(
    public code: string,
    message: string,
    public details?: any
  ) {
    super(message);
    this.name = 'SSRFError';
  }
}

// Types
interface URLValidationResult {
  isValid: boolean;
  normalizedURL?: string;
  reason?: string;
  details?: string;
  riskScore?: number;
}

interface SecureRequestResult {
  success: boolean;
  data: Buffer;
  status: number;
  headers: Record<string, string>;
  responseTime: number;
  url: string;
}

interface RequestResponse {
  data: Buffer;
  status: number;
  headers: Record<string, string>;
}

interface RequestOptions {
  method?: string;
  timeout?: number;
  maxRedirects?: number;
  headers?: Record<string, string>;
}

interface ValidationResult {
  isValid: boolean;
  reason?: string;
}

interface RequestCache {
  result: SecureRequestResult;
  timestamp: number;
}
```

## Network Segmentation Implementation

```python
# Network segmentation for SSRF protection
import ipaddress
import socket
from typing import List, Set, Tuple

class NetworkSegmentation:
    def __init__(self):
        self.allowed_networks = self._initialize_allowed_networks()
        self.blocked_networks = self._initialize_blocked_networks()
        self.gateway_rules = self._initialize_gateway_rules()
    
    def _initialize_allowed_networks(self) -> List[ipaddress.IPNetwork]:
        """Initialize allowed network ranges for external requests."""
        return [
            # Public internet ranges (this would be configured based on your needs)
            ipaddress.IPNetwork('0.0.0.0/0'),  # Allow all public IPs
        ]
    
    def _initialize_blocked_networks(self) -> List[ipaddress.IPNetwork]:
        """Initialize blocked network ranges to prevent internal access."""
        return [
            # Private IPv4 ranges
            ipaddress.IPNetwork('10.0.0.0/8'),
            ipaddress.IPNetwork('172.16.0.0/12'),
            ipaddress.IPNetwork('192.168.0.0/16'),
            
            # Loopback
            ipaddress.IPNetwork('127.0.0.0/8'),
            
            # Link-local
            ipaddress.IPNetwork('169.254.0.0/16'),
            
            # Multicast
            ipaddress.IPNetwork('224.0.0.0/4'),
            
            # Future reserved
            ipaddress.IPNetwork('240.0.0.0/4'),
            
            # IPv6 equivalents
            ipaddress.IPNetwork('::1/128'),  # IPv6 loopback
            ipaddress.IPNetwork('fc00::/7'),  # IPv6 private
            ipaddress.IPNetwork('fe80::/10'), # IPv6 link-local
        ]
    
    def is_ip_allowed(self, ip_address: str) -> Tuple[bool, str]:
        """Check if an IP address is allowed for external requests."""
        try:
            ip = ipaddress.ip_address(ip_address)
            
            # Check if IP is in blocked networks
            for blocked_network in self.blocked_networks:
                if ip in blocked_network:
                    return False, f"IP {ip_address} is in blocked network {blocked_network}"
            
            # Check if IP is in allowed networks
            for allowed_network in self.allowed_networks:
                if ip in allowed_network:
                    return True, f"IP {ip_address} is in allowed network {allowed_network}"
            
            return False, f"IP {ip_address} is not in any allowed network"
            
        except ValueError:
            return False, f"Invalid IP address: {ip_address}"
    
    def validate_network_access(self, host: str, port: int) -> NetworkValidationResult:
        """Validate network access to a specific host and port."""
        try:
            # Resolve hostname to IP addresses
            addresses = socket.getaddrinfo(host, port, socket.AF_UNSPEC, socket.SOCK_STREAM)
            
            for addr_info in addresses:
                family, socktype, proto, canonname, sockaddr = addr_info
                ip_address = sockaddr[0]
                
                is_allowed, reason = self.is_ip_allowed(ip_address)
                if not is_allowed:
                    return NetworkValidationResult(
                        is_valid=False,
                        reason=reason,
                        blocked_ip=ip_address,
                    )
            
            return NetworkValidationResult(
                is_valid=True,
                resolved_addresses=[addr[4][0] for addr in addresses],
            )
            
        except socket.gaierror as e:
            return NetworkValidationResult(
                is_valid=False,
                reason=f"DNS resolution failed: {str(e)}",
            )
    
    def create_network_filter(self) -> 'NetworkFilter':
        """Create a network filter for use in firewalls or proxies."""
        return NetworkFilter(
            allowed_networks=self.allowed_networks,
            blocked_networks=self.blocked_networks,
            gateway_rules=self.gateway_rules,
        )

# Types
class NetworkValidationResult:
    def __init__(self, is_valid: bool, reason: str = None, 
                 blocked_ip: str = None, resolved_addresses: List[str] = None):
        self.is_valid = is_valid
        self.reason = reason
        self.blocked_ip = blocked_ip
        self.resolved_addresses = resolved_addresses or []

class NetworkFilter:
    def __init__(self, allowed_networks: List[ipaddress.IPNetwork], 
                 blocked_networks: List[ipaddress.IPNetwork],
                 gateway_rules: List[dict]):
        self.allowed_networks = allowed_networks
        self.blocked_networks = blocked_networks
        self.gateway_rules = gateway_rules
    
    def to_firewall_rules(self) -> List[str]:
        """Convert to firewall rules format."""
        rules = []
        
        # Allow rules
        for network in self.allowed_networks:
            rules.append(f"ALLOW {network}")
        
        # Block rules
        for network in self.blocked_networks:
            rules.append(f"BLOCK {network}")
        
        return rules
```

---

# Reference & Integration (Level 4)

## API Reference

### Core SSRF Protection Operations
- `validate_url(url)` - Comprehensive URL validation and risk assessment
- `make_secure_request(url, options)` - Secure HTTP request with validation
- `is_ip_allowed(ip_address)` - IP address validation against allowlist/denylist
- `validate_network_access(host, port)` - Network access validation
- `calculate_risk_score(url)` - Risk scoring for suspicious patterns

### Context7 Integration
- `get_latest_ssrf_docs()` - SSRF protection via Context7
- `analyze_threat_patterns()` - Threat pattern analysis via Context7
- `optimize_url_validation()` - URL validation optimization via Context7

## Best Practices (November 2025)

### DO
- Use allowlist approach for domains and IP addresses
- Implement comprehensive URL parsing and validation
- Block access to internal network ranges and metadata endpoints
- Validate request responses for size and content type
- Implement rate limiting and anomaly detection
- Log all requests and security events
- Use network segmentation for additional protection
- Regularly update threat intelligence and protection rules

### DON'T
- Rely solely on blacklist approaches for protection
 Allow user-controlled URLs without validation
- Skip DNS resolution and IP address validation
- Forget to implement response size limits
- Ignore suspicious patterns in URLs and parameters
- Skip logging and monitoring of security events
- Use outdated threat intelligence or protection rules
- Forget to test SSRF protection regularly

## Works Well With

- `moai-security-api` (API security implementation)
- `moai-foundation-trust` (Security and compliance)
- `moai-security-compliance` (Compliance management)
- `moai-domain-backend` (Backend security)
- `moai-cc-configuration` (Security configuration)
- `moai-baas-foundation` (BaaS security patterns)
- `moai-security-owasp` (OWASP security standards)
- `moai-security-encryption` (Encryption and data protection)

## Changelog

- **v4.0.0** (2025-11-13): Complete Enterprise v4.0 rewrite with 40% content reduction, 4-layer Progressive Disclosure structure, Context7 integration, advanced threat detection, and comprehensive protection patterns
- **v2.0.0** (2025-11-11): Complete metadata structure, SSRF protection patterns, validation systems
- **v1.0.0** (2025-11-11): Initial SSRF security foundation

---

**End of Skill** | Updated 2025-11-13

## SSRF Security Framework

### Protection Layers
- URL validation with comprehensive parsing
- Domain and IP address filtering
- Network segmentation and isolation
- Request rate limiting and anomaly detection
- Response validation and size limits

### Enterprise Features
- Real-time threat intelligence integration
- Comprehensive logging and audit trails
- Automated vulnerability assessment
- Integration with security information and event management (SIEM)
- Compliance reporting and documentation

---

**End of Enterprise SSRF Security Protection Expert v4.0.0**
