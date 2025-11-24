---
name: moai-baas-cloudflare-ext
description: Moai Baas Cloudflare Ext - Professional implementation guide
version: 1.0.0
modularized: false
tags:
  - backend-as-a-service
  - cloudflare
  - platform
  - enterprise
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: ext, moai, baas, cloudflare  


$(head -n 30 /tmp/parent.md | head -n 20)

---

## Security Patterns

### Web Application Firewall (WAF)
- Layer 7 DDoS protection
- OWASP rule sets
- Custom firewall rules
- Rate limiting per IP

### Threat Mitigation

```typescript
// Worker with security headers
export default {
  async fetch(request) {
    const response = await handleRequest(request);
    response.headers.set('X-Content-Type-Options', 'nosniff');
    response.headers.set('X-Frame-Options', 'DENY');
    response.headers.set('X-XSS-Protection', '1; mode=block');
    return response;
  }
};
```

### Bot Management
- Challenge pages for suspicious traffic
- JavaScript challenges
- CAPTCHA integration

---

## Implementation Modules

For detailed patterns:
- **Core Implementation**: modules/core.md

---

**End of Skill** | Updated 2025-11-21
