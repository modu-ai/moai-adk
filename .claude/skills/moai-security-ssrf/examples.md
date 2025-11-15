# SSRF Prevention - Production Code Examples

## Example 1: URL Validation Middleware (Express)

```javascript
// middleware/ssrf-protection.js
const { URL } = require('url');
const dns = require('dns').promises;

class SSRFProtectionMiddleware {
  constructor(allowlist = []) {
    this.allowlist = allowlist || ['api.github.com', 'api.stripe.com'];
    this.dnsCache = new Map();
  }
  
  middleware() {
    return async (req, res, next) => {
      const { proxy_url } = req.body;
      
      if (!proxy_url) return next();
      
      try {
        const validation = await this.validateUrl(proxy_url);
        
        if (!validation.valid) {
          return res.status(403).json({
            error: `SSRF protection: ${validation.reason}`,
          });
        }
        
        req.proxiedUrl = proxy_url;
        next();
      } catch (error) {
        res.status(400).json({ error: error.message });
      }
    };
  }
  
  async validateUrl(urlString) {
    // 1. Parse URL
    let url;
    try {
      url = new URL(urlString);
    } catch {
      return { valid: false, reason: 'Invalid URL format' };
    }
    
    // 2. Check protocol
    if (!['http:', 'https:'].includes(url.protocol)) {
      return { valid: false, reason: 'Only HTTP/HTTPS allowed' };
    }
    
    // 3. Check allowlist
    if (!this.isAllowed(url.hostname)) {
      return { valid: false, reason: 'Domain not in allowlist' };
    }
    
    // 4. Resolve DNS and check IP
    const ip = await this.resolveDns(url.hostname);
    
    if (this.isPrivateIP(ip)) {
      return { valid: false, reason: 'Private IP detected' };
    }
    
    return { valid: true };
  }
  
  isAllowed(hostname) {
    return this.allowlist.some(pattern => {
      if (pattern instanceof RegExp) return pattern.test(hostname);
      return hostname === pattern || hostname.endsWith('.' + pattern);
    });
  }
  
  async resolveDns(hostname) {
    if (this.dnsCache.has(hostname)) {
      return this.dnsCache.get(hostname);
    }
    
    const ips = await dns.resolve4(hostname);
    const ip = ips[0];
    
    this.dnsCache.set(hostname, ip);
    setTimeout(() => this.dnsCache.delete(hostname), 60000);  // 1 minute TTL
    
    return ip;
  }
  
  isPrivateIP(ip) {
    const parts = ip.split('.').map(Number);
    
    return (
      parts[0] === 127 ||  // Loopback
      parts[0] === 10 ||  // Private 10.0.0.0/8
      (parts[0] === 172 && parts[1] >= 16 && parts[1] <= 31) ||  // 172.16.0.0/12
      (parts[0] === 192 && parts[1] === 168) ||  // 192.168.0.0/16
      (parts[0] === 169 && parts[1] === 254) ||  // 169.254.0.0/16 (AWS metadata!)
      parts[0] === 0  // 0.0.0.0/8
    );
  }
}

module.exports = SSRFProtectionMiddleware;
```

## Example 2: Secure HTTP Client

```javascript
// clients/ssrf-safe-http-client.js
const fetch = require('node-fetch');
const SSRFProtectionMiddleware = require('../middleware/ssrf-protection');

class SSRFSafeHttpClient {
  constructor(config) {
    this.config = config;
    this.protection = new SSRFProtectionMiddleware(config.allowlist);
    this.requestTimeout = 5000;  // 5 second timeout
  }
  
  async fetch(urlString, options = {}) {
    // Validate URL first
    const validation = await this.protection.validateUrl(urlString);
    
    if (!validation.valid) {
      throw new Error(`SSRF Protection: ${validation.reason}`);
    }
    
    // Make request with security measures
    try {
      const response = await fetch(urlString, {
        ...options,
        timeout: this.requestTimeout,
        redirect: 'manual',  // Don't follow redirects automatically
      });
      
      // Validate response status
      if ([301, 302, 303, 307, 308].includes(response.status)) {
        const location = response.headers.get('location');
        
        if (location) {
          // Validate redirect target
          const redirectValidation = await this.protection.validateUrl(location);
          if (!redirectValidation.valid) {
            throw new Error(`Redirect blocked: ${redirectValidation.reason}`);
          }
        }
      }
      
      return response;
    } catch (error) {
      console.error(`SSRF fetch error: ${error.message}`);
      throw error;
    }
  }
}

module.exports = SSRFSafeHttpClient;
```

## Example 3: WAF Integration

```javascript
// waf/ssrf-waf-rules.js
// ModSecurity/WAF rules exported as JSON

const ssrfWafRules = {
  rules: [
    {
      id: 10001,
      phase: 'REQUEST_HEADERS',
      message: 'AWS Metadata Service SSRF',
      pattern: '169\\.254\\.169\\.254',
      action: 'DENY',
      status: 403,
    },
    {
      id: 10002,
      phase: 'REQUEST_BODY',
      message: 'Kubernetes Service SSRF',
      pattern: 'kubernetes\\.default',
      action: 'DENY',
      status: 403,
    },
    {
      id: 10003,
      phase: 'REQUEST_BODY',
      message: 'Loopback IP SSRF',
      pattern: '^http.*:\\/\\/127\\.',
      action: 'DENY',
      status: 403,
    },
    {
      id: 10004,
      phase: 'REQUEST_BODY',
      message: 'Private Network SSRF',
      pattern: '^http.*:\\/\\/(10\\.|172\\.1[6-9]\\.|192\\.168\\.)',
      action: 'DENY',
      status: 403,
    },
  ],
};

module.exports = ssrfWafRules;
```

## Example 4: Context7 MCP Integration

```javascript
// clients/context7-ssrf-detector.js
const { Context7Client } = require('context7-mcp');

class Context7SSRFDetector {
  constructor(apiKey) {
    this.context7 = new Context7Client({
      apiKey,
      baseUrl: 'https://api.context7.ai',
    });
    this.threatCache = new Map();
  }
  
  async validateUrlWithThreatIntel(urlString) {
    const url = new URL(urlString);
    const cacheKey = url.hostname;
    
    // Check cache
    if (this.threatCache.has(cacheKey)) {
      return this.threatCache.get(cacheKey);
    }
    
    // Query Context7 threat intelligence
    try {
      const threat = await this.context7.query({
        type: 'url_reputation',
        hostname: url.hostname,
        tags: ['ssrf', 'internal', 'metadata_service'],
      });
      
      const result = {
        safe: threat.severity === 0,
        severity: threat.severity,
        indicators: threat.indicators,
        lastChecked: new Date(),
      };
      
      // Cache for 1 hour
      this.threatCache.set(cacheKey, result);
      setTimeout(() => this.threatCache.delete(cacheKey), 3600000);
      
      return result;
    } catch (error) {
      console.error('Context7 query failed:', error);
      return { safe: false, reason: 'Unable to verify' };
    }
  }
}

module.exports = Context7SSRFDetector;
```

## Example 5: Testing

```javascript
// tests/ssrf-protection.test.js
const SSRFProtectionMiddleware = require('../middleware/ssrf-protection');

describe('SSRF Protection', () => {
  let protection;
  
  beforeEach(() => {
    protection = new SSRFProtectionMiddleware([
      'api.github.com',
      /^.*\.example\.com$/,
    ]);
  });
  
  describe('URL Validation', () => {
    test('allows whitelisted domains', async () => {
      const result = await protection.validateUrl('https://api.github.com/users');
      expect(result.valid).toBe(true);
    });
    
    test('blocks private IP ranges', async () => {
      const result = await protection.validateUrl('http://127.0.0.1:8080');
      expect(result.valid).toBe(false);
      expect(result.reason).toContain('Private IP');
    });
    
    test('blocks AWS metadata service', async () => {
      const result = await protection.validateUrl('http://169.254.169.254/');
      expect(result.valid).toBe(false);
    });
    
    test('blocks non-HTTP protocols', async () => {
      const result = await protection.validateUrl('file:///etc/passwd');
      expect(result.valid).toBe(false);
    });
    
    test('blocks DNS rebinding attempts', async () => {
      // This would require DNS mocking
      // Example: hostname that resolves to private IP
    });
  });
});
```

