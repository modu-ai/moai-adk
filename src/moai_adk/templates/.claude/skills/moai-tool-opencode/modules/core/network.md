---
source: https://opencode.ai/docs/network/
fetched: 2026-01-08
title: Network Configuration
---

# Network Configuration Guide

OpenCode provides enterprise-grade networking support through proxy and certificate management.

## Proxy Setup

OpenCode recognizes standard proxy environment variables for routing traffic through corporate proxies:

```bash
export HTTPS_PROXY=https://proxy.example.com:8080
export HTTP_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1
```

**Critical requirement:** The `NO_PROXY` variable must exclude localhost addresses. The TUI relies on local HTTP server communication, and proxying this traffic creates routing loops.

## Authentication

For credential-protected proxies, embed credentials directly in the proxy URL:

```bash
export HTTPS_PROXY=http://username:password@proxy.example.com:8080
```

### Security Best Practices

The documentation advises against hardcoding passwords, recommending:
- Environment variables
- Secure credential management systems

### Advanced Authentication

For advanced authentication schemes like NTLM or Kerberos, using an LLM Gateway with native support is suggested.

## Custom Certificates

Enterprise environments using custom certificate authorities can configure OpenCode to trust them:

```bash
export NODE_EXTRA_CA_CERTS=/path/to/ca-cert.pem
```

This setting applies to both proxy connections and direct API communications.

## Server Configuration

OpenCode supports configurable server ports and hostnames via CLI flags:

```bash
opencode --port 3000 --hostname localhost
```

## Complete Example

```bash
# Proxy configuration
export HTTPS_PROXY=https://proxy.corporate.com:8080
export HTTP_PROXY=http://proxy.corporate.com:8080
export NO_PROXY=localhost,127.0.0.1,.internal.corp

# Custom CA certificate
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/corporate-ca.pem

# Start OpenCode
opencode
```
