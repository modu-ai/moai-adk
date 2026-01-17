# MCP Security and Authorization

Comprehensive security guide for MCP implementations.

---

## Security Model Overview

### Trust Boundaries

MCP defines clear trust boundaries between components:

Client Trust: The MCP client (Claude Code) is trusted to manage user interactions and enforce access policies.

Server Trust: MCP servers should be treated with appropriate trust levels based on their source and purpose.

Transport Trust: Stdio transport is inherently trusted (local process). Network transports require authentication.

### Principle of Least Privilege

Grant minimum necessary permissions to servers.

Limit resource access to required files and directories.

Restrict tool capabilities to essential operations.

Avoid exposing sensitive system information.

---

## Authentication Patterns

### No Authentication (Stdio)

Stdio transport inherently authenticates through process ownership.

Parent process controls child process access.

No additional authentication needed for local servers.

### API Key Authentication

Simple authentication for remote servers.

Pass API key in Authorization header or custom header.

Validate key server-side before processing requests.

Store keys securely, never in code or configuration.

### Bearer Token Authentication

Standard OAuth 2.0 bearer token pattern.

Include token in Authorization: Bearer header.

Validate token signature and expiration.

Support token refresh for long-running sessions.

### Mutual TLS

Strongest authentication for high-security environments.

Both client and server present certificates.

Certificates verified against trusted CA.

Provides authentication and encryption.

---

## Authorization Patterns

### Resource-Based Authorization

Control access based on resource URI.

Define allowed URI patterns for each client.

Validate resource access before returning content.

Implement path traversal protection.

### Tool-Based Authorization

Control which tools clients can invoke.

Define allowed tool lists per client or role.

Validate tool access before execution.

Log all tool invocations for audit.

### Role-Based Access Control

Define roles with specific permissions.

Assign roles to authenticated clients.

Check role permissions before operations.

Support role hierarchies for complex scenarios.

---

## Input Validation

### Parameter Validation

Validate all input parameters against schemas.

Reject invalid or unexpected parameter types.

Sanitize string inputs to prevent injection.

Limit parameter sizes to prevent resource exhaustion.

### Path Validation

Normalize file paths before use.

Reject paths with directory traversal patterns.

Validate paths against allowed directories.

Use allowlists for permitted file types.

### SQL Injection Prevention

Use parameterized queries for database access.

Never concatenate user input into SQL strings.

Validate input types match expected database types.

Limit query result sizes.

### Command Injection Prevention

Avoid shell command execution when possible.

Use subprocess with argument arrays, not shell strings.

Validate and sanitize any command arguments.

Restrict allowed commands to specific allowlist.

---

## Secure Communication

### Transport Security

Use HTTPS for all network communication.

Configure TLS with strong cipher suites.

Validate server certificates.

Implement certificate pinning for high-security.

### Message Integrity

JSON-RPC provides basic message structure validation.

Consider message signing for sensitive operations.

Validate message IDs to prevent replay attacks.

### Credential Protection

Never log credentials or secrets.

Use environment variables for sensitive configuration.

Encrypt credentials at rest.

Rotate credentials regularly.

---

## Error Handling Security

### Information Disclosure Prevention

Avoid exposing internal details in error messages.

Log detailed errors server-side only.

Return generic error messages to clients.

Never include stack traces in responses.

### Error Rate Limiting

Track error rates per client.

Implement progressive delays for repeated errors.

Block clients with excessive error rates.

Alert on unusual error patterns.

---

## Audit and Logging

### Operation Logging

Log all tool invocations with parameters.

Log resource access with URIs.

Log authentication events.

Include timestamps and client identifiers.

### Sensitive Data Handling

Mask sensitive data in logs.

Never log credentials or secrets.

Implement log retention policies.

Secure log storage access.

### Monitoring

Monitor for unusual access patterns.

Alert on authentication failures.

Track resource usage metrics.

Implement anomaly detection.

---

## Server Security Best Practices

### Process Isolation

Run servers with minimum required privileges.

Use separate user accounts for servers.

Consider containerization for isolation.

Limit file system access.

### Dependency Security

Keep dependencies updated.

Monitor for security vulnerabilities.

Use dependency scanning tools.

Pin versions for reproducibility.

### Configuration Security

Store configuration securely.

Validate configuration on startup.

Reject insecure configurations.

Document security-relevant settings.

---

## Client Security Best Practices

### Server Verification

Verify server identity before connecting.

Validate server capabilities.

Check for unexpected capability changes.

Monitor server behavior.

### Response Validation

Validate response structure.

Check for unexpected response content.

Handle malicious response patterns.

Limit response sizes.

### Sandboxing

Run servers in sandboxed environments.

Limit network access for local servers.

Restrict file system access.

Monitor resource usage.

---

## Incident Response

### Detection

Monitor for security anomalies.

Log security-relevant events.

Implement alerting thresholds.

Regular security audits.

### Response

Document incident response procedures.

Isolate compromised components.

Preserve evidence for analysis.

Notify affected parties.

### Recovery

Rotate compromised credentials.

Patch vulnerable components.

Restore from known-good state.

Update security measures.

---

## Compliance Considerations

### Data Protection

Classify data sensitivity levels.

Implement data access controls.

Document data handling procedures.

Support data subject requests.

### Audit Requirements

Maintain audit logs.

Implement log integrity protection.

Support audit report generation.

Regular compliance reviews.

### Regulatory Frameworks

Consider GDPR for personal data.

Consider SOC 2 for service security.

Consider industry-specific requirements.

Document compliance measures.
