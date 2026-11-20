# Claude Code Sandboxing - Official Documentation Reference

Source: https://code.claude.com/docs/en/sandboxing
Updated: 2026-01-06

## Overview

Claude Code provides OS-level sandboxing to restrict file system and network access during code execution. This creates a security boundary that limits potential damage from malicious or buggy code.

## Sandbox Implementation

### Operating System Support

Linux: Uses bubblewrap (bwrap) for namespace-based isolation
macOS: Uses Seatbelt (sandbox-exec) for profile-based restrictions

### Default Behavior

When sandboxing is enabled:

- File writes are restricted to the current working directory
- Network access is limited to allowed domains
- System resources are protected from modification

## Filesystem Isolation

### Default Write Restrictions

By default, sandboxed commands can only write to:
- Current working directory
- Subdirectories of current working directory

Reads are generally unrestricted within user-accessible paths.

### Configuring Allowed Paths

Additional write paths can be configured in settings.json:

```json
{
  "sandbox": {
    "allowedPaths": [
      "/tmp/build-output",
      "~/project-cache"
    ],
    "deniedPaths": [
      "~/.ssh",
      "~/.aws"
    ]
  }
}
```

## Network Isolation

### Domain-Based Restrictions

Network access is filtered by domain. Configure allowed domains:

```json
{
  "sandbox": {
    "allowedDomains": [
      "api.example.com",
      "registry.npmjs.org"
    ],
    "deniedDomains": [
      "*.internal.corp"
    ]
  }
}
```

### Default Allowed Domains

Common development services are typically allowed:
- npm registry (registry.npmjs.org)
- GitHub (github.com, api.github.com)
- Claude API (api.anthropic.com)
- DNS services

### Port Configuration

Configure network port access:

```json
{
  "sandbox": {
    "allowedPorts": [80, 443, 8080],
    "deniedPorts": [22, 3306]
  }
}
```

## Auto-Allow Mode

When sandbox is enabled, bash commands that operate within sandbox restrictions can run without permission prompts.

### How Auto-Allow Works

If a command only:
- Reads from allowed paths
- Writes to allowed paths
- Accesses allowed network domains

Then it executes automatically without user confirmation.

### Commands Excluded from Sandbox

Some commands bypass sandbox restrictions:

```json
{
  "sandbox": {
    "excludedCommands": [
      "docker",
      "kubectl"
    ]
  }
}
```

Excluded commands require explicit user permission.

## Security Limitations

### Domain-Only Filtering

Network filtering operates at the domain level only:
- Cannot inspect traffic content
- Cannot filter by URL path
- Cannot decrypt HTTPS traffic

### Unix Socket Access

Unix sockets can grant system access:
- Docker socket provides host system access
- Some sockets bypass network restrictions
- Configure socket permissions carefully

### Permission Implications

Certain permissions grant broader access:
- Docker socket access equals root-equivalent access
- Some build tools require expanded permissions
- Evaluate security tradeoffs carefully

## Configuration Examples

### Restrictive Configuration

For maximum security:

```json
{
  "sandbox": {
    "enabled": true,
    "allowedPaths": [],
    "allowedDomains": [],
    "excludedCommands": []
  }
}
```

### Development Configuration

For typical development workflows:

```json
{
  "sandbox": {
    "enabled": true,
    "allowedPaths": [
      "/tmp",
      "~/.npm",
      "~/.cache"
    ],
    "allowedDomains": [
      "registry.npmjs.org",
      "github.com",
      "api.github.com"
    ]
  }
}
```

### CI/CD Configuration

For automated pipelines:

```json
{
  "sandbox": {
    "enabled": true,
    "allowedPaths": [
      "/workspace",
      "/build-cache"
    ],
    "allowedDomains": [
      "registry.npmjs.org",
      "docker.io"
    ],
    "excludedCommands": [
      "docker"
    ]
  }
}
```

## Monitoring Sandbox Violations

### Identifying Blocked Operations

When sandbox blocks an operation:
1. Permission dialog appears (if not in auto-allow mode)
2. Operation is logged
3. User can choose to allow or deny

### Reviewing Sandbox Logs

Check sandbox violation patterns:
- Repeated blocks may indicate configuration gaps
- Unexpected blocks may indicate security issues
- Review and adjust configuration as needed

## Best Practices

### Start Restrictive

Begin with minimal permissions:
1. Enable sandbox with default restrictions
2. Monitor for violations
3. Add specific allowances as needed

### Document Exceptions

When adding exclusions:
- Document why each exception is needed
- Review exceptions periodically
- Remove unnecessary exceptions

### Combine with IAM

Use sandbox as one layer of defense:
- Sandbox provides OS-level isolation
- IAM provides Claude-level permissions
- Together they create defense-in-depth

### Test Configuration

Before deploying:
- Test common workflows with sandbox enabled
- Verify necessary operations succeed
- Confirm sensitive operations are blocked

## Troubleshooting

### Command Fails in Sandbox

If a legitimate command is blocked:
1. Check if command needs excluded commands list
2. Verify path is in allowed paths
3. Check domain is in allowed domains

### Network Request Blocked

If network request fails:
1. Verify domain spelling
2. Check for subdomain requirements
3. Review port restrictions

### Performance Impact

Sandbox adds minimal overhead:
- Namespace creation is fast
- File checks are cached
- Network filtering is lightweight

If experiencing slowdowns, check:
- Large allowed paths lists
- Complex domain patterns
- Excessive sandbox violations
