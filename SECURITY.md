# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 2.x     | :white_check_mark: |
| 1.x     | :x:                |

## Reporting a Vulnerability

We take security seriously at MoAI-ADK. If you discover a security vulnerability, please report it responsibly.

### How to Report

1. **Do NOT** open a public GitHub issue for security vulnerabilities.
2. Email the security report to the maintainers via [GitHub Security Advisories](https://github.com/modu-ai/moai-adk/security/advisories/new).
3. Include a detailed description of the vulnerability, steps to reproduce, and any potential impact.

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your report within 48 hours.
- **Assessment**: We will assess the vulnerability and determine its severity within 7 days.
- **Resolution**: Critical vulnerabilities will be patched within 14 days. Non-critical issues will be addressed in the next scheduled release.
- **Disclosure**: We follow responsible disclosure practices. We will coordinate with you on the timeline for public disclosure.

### Scope

The following are in scope for security reports:

- MoAI-ADK Go binary (`moai`)
- Template deployment system
- Hook execution system
- Configuration file handling
- CLI command injection vectors

### Out of Scope

- Claude Code itself (report to Anthropic)
- Third-party MCP servers
- Issues in dependencies (report upstream)

## Security Best Practices

When using MoAI-ADK:

- Never commit secrets (API keys, tokens) to version control
- Use environment variables for sensitive configuration
- Review hook scripts before execution
- Keep MoAI-ADK updated to the latest version
