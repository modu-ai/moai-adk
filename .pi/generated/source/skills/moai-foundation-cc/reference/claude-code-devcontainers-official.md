# Claude Code Dev Containers - Official Documentation Reference

Source: https://code.claude.com/docs/en/devcontainer
Updated: 2026-01-06

## Overview

Claude Code dev containers provide security-hardened development environments using container technology. They enable isolated, reproducible, and secure Claude Code sessions.

## Architecture

### Base Configuration

Dev containers are built on:
- Node.js 20 with essential development tools
- Custom security firewall
- VS Code Dev Containers integration

### Components

1. devcontainer.json: Container configuration and settings
2. Dockerfile: Image definition and tool installation
3. init-firewall.sh: Network security rule initialization

## Security Features

### Network Isolation

Default-deny policy with whitelisted outbound connections:

Allowed by default:
- npm registry (registry.npmjs.org)
- GitHub (github.com, api.github.com)
- Claude API (api.anthropic.com)
- DNS services
- SSH for git operations

All other external connections are blocked.

### Firewall Configuration

The init-firewall.sh script establishes:
- Outbound whitelist rules
- Default-deny for unlisted domains
- Startup verification of firewall status

### Customizing Network Access

Modify init-firewall.sh to add custom allowed domains:

```bash
# Add custom domain to whitelist
iptables -A OUTPUT -d custom.example.com -j ACCEPT
```

## VS Code Integration

### Required Extensions

The devcontainer.json can specify VS Code extensions:

```json
{
  "customizations": {
    "vscode": {
      "extensions": [
        "ms-python.python",
        "esbenp.prettier-vscode"
      ]
    }
  }
}
```

### Settings Override

Container-specific VS Code settings:

```json
{
  "customizations": {
    "vscode": {
      "settings": {
        "editor.formatOnSave": true
      }
    }
  }
}
```

## Volume Mounts

### Default Mounts

Typical dev container mounts:
- Workspace directory
- Git credentials
- SSH keys (optional)

### Custom Mounts

Add custom mounts in devcontainer.json:

```json
{
  "mounts": [
    "source=${localWorkspaceFolder},target=/workspace,type=bind",
    "source=${localEnv:HOME}/.npm,target=/home/node/.npm,type=bind"
  ]
}
```

## Unattended Operation

### Skip Permissions Flag

For fully automated environments:

```bash
claude --dangerously-skip-permissions
```

This bypasses all permission prompts.

### Security Warning

When using --dangerously-skip-permissions:

- Container has full access to mounted volumes
- Malicious code can access Claude Code credentials
- Only use with fully trusted repositories
- Never expose container to untrusted input

### Recommended Use Cases

Safe usage scenarios:
- Controlled CI/CD pipelines
- Isolated testing environments
- Trusted internal repositories

Unsafe scenarios:
- Public code execution
- Untrusted repository analysis
- User-facing automation

## Resource Configuration

### CPU and Memory

Configure resource limits in devcontainer.json:

```json
{
  "hostRequirements": {
    "cpus": 4,
    "memory": "8gb",
    "storage": "32gb"
  }
}
```

### GPU Access

For AI/ML workloads:

```json
{
  "hostRequirements": {
    "gpu": "optional"
  }
}
```

## Shell Configuration

### Default Shell

Set default shell in Dockerfile:

```dockerfile
RUN chsh -s /bin/zsh node
```

### Shell Customization

Add custom shell configuration:

```dockerfile
COPY .zshrc /home/node/.zshrc
```

## Tool Installation

### System Packages

In Dockerfile:

```dockerfile
RUN apt-get update && apt-get install -y \
    git \
    curl \
    jq \
    && rm -rf /var/lib/apt/lists/*
```

### Development Tools

```dockerfile
RUN npm install -g \
    typescript \
    eslint \
    prettier
```

### Language Runtimes

```dockerfile
# Python
RUN apt-get install -y python3 python3-pip

# Go
RUN wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz && \
    tar -xzf go1.21.0.linux-amd64.tar.gz -C /usr/local
```

## Use Cases

### Client Project Isolation

Isolate client work:
- Separate container per client
- Independent credentials
- No cross-contamination risk

### Team Onboarding

Standardized setup:
- Consistent tool versions
- Pre-configured environment
- Reduced setup time

### CI/CD Mirroring

Match production:
- Same dependencies
- Same security policies
- Reproducible builds

### Development Standardization

Team consistency:
- Shared configurations
- Common tooling
- Unified workflows

## Creating a Dev Container

### Step 1: Create Directory

```bash
mkdir -p .devcontainer
```

### Step 2: Create devcontainer.json

```json
{
  "name": "Claude Code Development",
  "build": {
    "dockerfile": "Dockerfile"
  },
  "customizations": {
    "vscode": {
      "extensions": ["anthropic.claude-code"]
    }
  },
  "postCreateCommand": "npm install"
}
```

### Step 3: Create Dockerfile

```dockerfile
FROM node:20-slim

# Install essential tools
RUN apt-get update && apt-get install -y \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code
RUN npm install -g @anthropic-ai/claude-code

# Set up non-root user
USER node
WORKDIR /workspace
```

### Step 4: Create Firewall Script

```bash
#!/bin/bash
# init-firewall.sh

# Default deny
iptables -P OUTPUT DROP

# Allow established connections
iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# Allow localhost
iptables -A OUTPUT -o lo -j ACCEPT

# Allow DNS
iptables -A OUTPUT -p udp --dport 53 -j ACCEPT

# Allow HTTPS
iptables -A OUTPUT -p tcp --dport 443 -j ACCEPT

# Allow specific domains (resolve IPs)
# Add your domain allowlist here
```

### Step 5: Open in Container

In VS Code:
1. Install Remote - Containers extension
2. Command Palette: "Dev Containers: Reopen in Container"

## Best Practices

### Security

- Review firewall rules regularly
- Minimize allowed domains
- Audit tool installations
- Use specific image versions

### Performance

- Use volume caching for dependencies
- Pre-build images for common configurations
- Optimize Dockerfile layers

### Maintenance

- Document customizations
- Version control devcontainer configs
- Test container builds regularly
- Update base images periodically

## Troubleshooting

### Container Build Fails

Check:
- Dockerfile syntax
- Network access during build
- Base image availability

### Network Issues

If connectivity problems occur:
- Verify firewall rules
- Check DNS resolution
- Test allowed domains manually

### Permission Issues

If permission denied errors:
- Check user configuration
- Verify volume mount permissions
- Review file ownership

### VS Code Connection Issues

If VS Code cannot connect:
- Verify Docker is running
- Check extension installation
- Review devcontainer.json syntax
