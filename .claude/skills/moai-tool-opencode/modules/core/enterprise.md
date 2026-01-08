---
source: https://opencode.ai/docs/enterprise/
fetched: 2026-01-08
title: Enterprise
---

# OpenCode Enterprise

## Overview

OpenCode Enterprise enables organizations to keep code and data within their own infrastructure through centralized configuration, SSO integration, and internal AI gateway support.

## Trial & Data Security

**Key Statement:** "OpenCode does not store any of your code or context data."

All processing occurs locally or via direct API calls to the provider. The only exception is the optional `/share` feature, which sends conversations to opencode.ai's CDN.

### Recommended Trial Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "share": "disabled"
}
```

### Code Ownership

Users retain all rights to code produced by OpenCode with no licensing restrictions.

## Enterprise Features

### Central Config

Single organizational configuration managing SSO and AI gateway access.

### SSO Integration

Authentication through existing identity management systems to obtain gateway credentials.

### Internal AI Gateway

Route all requests exclusively through approved organizational infrastructure.

### Self-Hosting

Currently on roadmap; option to host share pages on internal infrastructure.

## Pricing Model

- Per-seat licensing
- No token charges if organizations maintain their own LLM gateway
- Custom quotes available upon contact

## Private NPM Registry Support

Organizations using private registries (JFrog Artifactory, Nexus, etc.) must authenticate developers before running OpenCode:

```bash
npm login --registry=https://your-company.jfrog.io/api/npm/npm-virtual/
```

**Important Requirement:** "You must be logged into the private registry before running OpenCode."

## Contact

Organizations interested in implementation should email contact@anoma.ly for pricing and deployment discussions.
