---
name: moai-security-ssrf
version: 4.0.0
updated: 2025-11-20
status: stable
description: Enterprise SSRF protection with URL validation and network segmentation
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# SSRF Protection Expert

**Server-Side Request Forgery (SSRF) Prevention**

> **Focus**: URL Validation, Allowlisting, Network Segmentation  
> **Stack**: Python (urllib/ipaddress), TypeScript (URL API)

---

## Overview

Comprehensive defense strategies against SSRF attacks where an attacker induces the server to make requests to unintended locations.

### Core Defense Layers

1.  **Input Validation**: Strict allowlisting of domains and protocols.
2.  **Network Layer**: Blocking access to internal/private IP ranges (10.0.0.0/8, 127.0.0.1, etc.).
3.  **Application Layer**: Disabling redirects, validating response types.
4.  **Infrastructure**: Running services in isolated network environments.

---

## Implementation Patterns

### 1. Robust URL Validator (Python)

Validates URLs against allowlists and blocks private IP ranges (including DNS rebinding protection).

```python
import socket
import ipaddress
from urllib.parse import urlparse

class SSRFValidator:
    def __init__(self, allowed_domains=None):
        self.allowed_domains = allowed_domains or []
        self.blocked_networks = [
            ipaddress.ip_network('10.0.0.0/8'),
            ipaddress.ip_network('172.16.0.0/12'),
            ipaddress.ip_network('192.168.0.0/16'),
            ipaddress.ip_network('127.0.0.0/8'),
            ipaddress.ip_network('169.254.0.0/16'), # Cloud metadata
        ]

    def validate(self, url: str) -> bool:
        try:
            parsed = urlparse(url)

            # 1. Protocol Check
            if parsed.scheme not in ('http', 'https'):
                return False

            # 2. Domain Allowlist Check
            if self.allowed_domains and parsed.hostname not in self.allowed_domains:
                return False

            # 3. DNS Resolution & IP Check (Anti-DNS Rebinding)
            # Note: In production, use the resolved IP for the actual request
            ip_str = socket.gethostbyname(parsed.hostname)
            ip_addr = ipaddress.ip_address(ip_str)

            for network in self.blocked_networks:
                if ip_addr in network:
                    return False

            return True
        except Exception:
            return False

# Usage
validator = SSRFValidator(allowed_domains=['api.example.com', 'google.com'])
is_safe = validator.validate("https://169.254.169.254/latest/meta-data/") # False
```

### 2. Secure HTTP Client (Python)

Making requests safely using the validation logic.

```python
import requests
from requests.adapters import HTTPAdapter

class SecureClient:
    def __init__(self, validator):
        self.validator = validator
        self.session = requests.Session()
        self.session.mount('https://', HTTPAdapter(max_retries=3))

    def get(self, url: str, **kwargs):
        # 1. Validate URL first
        if not self.validator.validate(url):
            raise ValueError(f"Blocked potentially unsafe URL: {url}")

        # 2. Disable redirects to prevent open redirect bypasses
        kwargs['allow_redirects'] = False

        # 3. Set reasonable timeout
        kwargs.setdefault('timeout', 5.0)

        return self.session.get(url, **kwargs)
```

### 3. Network Segmentation (Infrastructure)

Ensure the application server cannot route to sensitive internal services.

**AWS Security Group (Outbound Rules)**:

- **Allow**: 0.0.0.0/0 (Internet) on port 80/443
- **Deny**: 10.0.0.0/8 (VPC Internal)
- **Deny**: 169.254.169.254/32 (Instance Metadata)

**Docker Network**:

```yaml
services:
  app:
    networks:
      - public_net
    # Isolate from internal services

networks:
  public_net:
    internal: false
```

---

## Security Best Practices

1.  **Allowlist over Blocklist**: Always prefer explicitly allowing known good domains over trying to block bad ones.
2.  **Disable Redirects**: Attackers often use open redirects on trusted domains to bypass filters.
3.  **Validate Response**: Check `Content-Type` and response size to prevent DoS or unexpected parsing.
4.  **Use Dedicated Parsers**: Don't use regex for URL parsing; use standard libraries (`urllib`, `URL`).

---

## Validation Checklist

- [ ] **Protocols**: Only `http` and `https` allowed? (No `file://`, `gopher://`)
- [ ] **Internal IPs**: 127.0.0.1, 10.x.x.x, 192.168.x.x blocked?
- [ ] **Cloud Metadata**: 169.254.169.254 blocked?
- [ ] **Redirects**: Disabled or strictly validated?
- [ ] **DNS**: Resolved IP checked against blocklist?

---

## Related Skills

- `moai-security-devsecops`: Security testing
- `moai-cloud-aws-advanced`: AWS network security
- `moai-security-encryption`: Secure communication

---

**Last Updated**: 2025-11-20
