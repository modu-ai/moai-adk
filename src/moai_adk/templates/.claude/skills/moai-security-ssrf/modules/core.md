        
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
