# Claude Code IAM & Permissions - Official Documentation Reference

Source: https://code.claude.com/docs/en/iam

## Key Concepts

### What is Claude Code IAM?

Identity and Access Management (IAM) in Claude Code provides a comprehensive permission system that controls access to tools, files, and external services. IAM implements tiered approval levels, role-based access control, and security boundaries to ensure safe and compliant operations.

### IAM Architecture

Tiered Permission System:
```
Level 1: Read-only Access (No Approval)
 Read, Grep, Glob
 Information gathering tools

Level 2: Bash Commands (User Approval Required)
 Bash, WebFetch, WebSearch
 System operations and external access

Level 3: File Modification (User Approval Required)
 Write, Edit, MultiEdit
 File system modifications

Level 4: Administrative (Enterprise Approval)
 Settings management
 User administration
 System configuration
```

## Tool-Specific Permission Rules

### Permission Rule Format

Basic Permission Structure:
```json
{
 "allowedTools": [
 "Read", // Read-only access (no approval)
 "Bash", // Commands with approval
 "Write", // File modification with approval
 "WebFetch(domain:*.example.com)" // Domain-specific web access
 ]
}
```

### Permission Levels and Tools

Level 1: Read-Only Tools (No Approval Required)
```json
{
 "readLevel": {
 "tools": ["Read", "Grep", "Glob"],
 "approval": "none",
 "description": "Information gathering and file exploration",
 "useCases": [
 "Code analysis and review",
 "File system exploration",
 "Pattern searching and analysis",
 "Documentation reading"
 ]
 }
}
```

Level 2: System Operations (User Approval Required)
```json
{
 "systemLevel": {
 "tools": ["Bash", "WebFetch", "WebSearch"],
 "approval": "user",
 "description": "System operations and external resource access",
 "useCases": [
 "Build and deployment operations",
 "External API integration",
 "System configuration changes",
 "Network operations"
 ]
 }
}
```

Level 3: File Modifications (User Approval Required)
```json
{
 "modificationLevel": {
 "tools": ["Write", "Edit", "MultiEdit", "NotebookEdit"],
 "approval": "user",
 "description": "File system modifications and content creation",
 "useCases": [
 "Code implementation and changes",
 "Documentation updates",
 "Configuration file modifications",
 "Content generation"
 ]
 }
}
```

Level 4: Administrative (Enterprise Approval Required)
```json
{
 "adminLevel": {
 "tools": ["Settings", "UserManagement", "SystemConfig"],
 "approval": "enterprise",
 "description": "System administration and user management",
 "useCases": [
 "System configuration changes",
 "User permission management",
 "Enterprise policy updates",
 "Security configuration"
 ]
 }
}
```

## Role-Based Access Control (RBAC)

### Predefined Roles

Developer Role:
```json
{
 "developer": {
 "allowedTools": [
 "Read", "Grep", "Glob",
 "Bash", "Write", "Edit",
 "WebFetch", "WebSearch",
 "AskUserQuestion", "Task", "Skill"
 ],
 "toolRestrictions": {
 "Bash": {
 "allowedCommands": ["git", "npm", "python", "make", "docker"],
 "blockedCommands": ["sudo", "chmod 777", "rm -rf /"],
 "requireConfirmation": true
 },
 "WebFetch": {
 "allowedDomains": ["*.github.com", "*.npmjs.com", "docs.python.org"],
 "blockedDomains": ["*.malicious-site.com"],
 "maxRequestsPerMinute": 60
 },
 "Write": {
 "allowedPaths": ["./src/", "./tests/", "./docs/"],
 "blockedPaths": ["./.env*", "./config/secrets"],
 "maxFileSize": 10000000
 }
 },
 "permissions": {
 "canCreateFiles": true,
 "canModifyFiles": true,
 "canExecuteCommands": true,
 "canAccessExternal": true
 }
 }
}
```

Security Reviewer Role:
```json
{
 "securityReviewer": {
 "allowedTools": [
 "Read", "Grep", "Glob",
 "Bash", "WebFetch",
 "AskUserQuestion", "Task"
 ],
 "toolRestrictions": {
 "Read": {
 "allowedPaths": ["./"],
 "blockedPatterns": ["*.key", "*.pem", ".env*"]
 },
 "Bash": {
 "allowedCommands": ["git", "grep", "find", "openssl"],
 "requireConfirmation": true
 }
 },
 "specialPermissions": {
 "canAccessSecurityLogs": true,
 "canRunSecurityScans": true,
 "canReviewPermissions": true,
 "cannotModifyProduction": true
 }
 }
}
```

DevOps Engineer Role:
```json
{
 "devopsEngineer": {
 "allowedTools": [
 "Read", "Grep", "Glob",
 "Bash", "Write", "Edit",
 "WebFetch", "WebSearch",
 "Task", "Skill"
 ],
 "toolRestrictions": {
 "Bash": {
 "allowedCommands": [
 "git", "docker", "kubectl", "helm", "terraform",
 "npm", "pip", "make", "curl", "wget"
 ],
 "blockedCommands": ["sudo", "chmod 777"],
 "requireConfirmation": false
 },
 "WebFetch": {
 "allowedDomains": ["*"],
 "requireConfirmation": false
 }
 },
 "permissions": {
 "canDeployToStaging": true,
 "canManageInfrastructure": true,
 "canAccessProduction": false,
 "canManageCI/CD": true
 }
 }
}
```

### Custom Role Definition

Role Template:
```json
{
 "customRole": {
 "name": "CustomRoleName",
 "description": "Role description and purpose",
 "allowedTools": ["Read", "Bash", "Write"],
 "toolRestrictions": {
 "Read": {
 "allowedPaths": ["./"],
 "blockedPaths": [".env*", "secrets/"]
 },
 "Bash": {
 "allowedCommands": ["git", "npm"],
 "blockedCommands": ["rm", "sudo"],
 "requireConfirmation": true
 }
 },
 "permissions": {
 "customPermission": "value"
 },
 "inherits": ["developer"]
 }
}
```

## Enterprise Policy Overrides

### Enterprise IAM Structure

Enterprise Policy Framework:
```json
{
 "enterprise": {
 "policies": {
 "tools": {
 "Bash": "never",
 "WebFetch": ["domain:*.company.com", "domain:*.partner.com"],
 "Write": ["path:./workspace/", "path:./temp/"]
 },
 "mcpServers": {
 "allowed": ["context7", "figma", "company-internal-mcp"],
 "blocked": ["custom-unverified-mcp", "external-scanner"]
 },
 "roles": {
 "default": "readonly-developer",
 "overrides": {
 "senior-developer": "developer",
 "devops": "devops-engineer"
 }
 },
 "compliance": {
 "auditRequired": true,
 "dataRetention": "7y",
 "encryptionRequired": true,
 "mfaRequired": true
 }
 }
 }
}
```

Policy Enforcement Mechanisms:
```json
{
 "policyEnforcement": {
 "validation": {
 "strict": true,
 "failOnViolation": true,
 "auditFrequency": "daily"
 },
 "overrides": {
 "allowUserOverrides": false,
 "requireManagerApproval": true,
 "emergencyOverrides": {
 "enabled": true,
 "duration": "24h",
 "approvalRequired": ["cto", "security-team"]
 }
 },
 "monitoring": {
 "realTimeAlerts": true,
 "anomalyDetection": true,
 "complianceReporting": true
 }
 }
}
```

## MCP Server Permissions

### MCP Access Control

MCP Server Configuration:
```json
{
 "allowedMcpServers": [
 "context7",
 "figma-dev-mode-mcp-server",
 "playwright",
 "company-internal-mcp"
 ],
 "blockedMcpServers": [
 "custom-unverified-mcp",
 "experimental-ai-mcp",
 "external-scanner-mcp"
 ],
 "mcpServerPermissions": {
 "context7": {
 "allowed": ["resolve-library-id", "get-library-docs"],
 "rateLimit": {
 "requestsPerMinute": 60,
 "burstSize": 10
 },
 "dataUsage": {
 "allowedDataTypes": ["documentation", "api-reference"],
 "blockedDataTypes": ["credentials", "private-keys"]
 }
 },
 "figma-dev-mode-mcp-server": {
 "allowed": ["get-design-context", "get-variable-defs", "get-screenshot"],
 "accessControl": {
 "allowedProjects": ["company-design-system"],
 "blockedProjects": ["competitor-designs"]
 }
 }
 }
}
```

MCP Security Validation:
```json
{
 "mcpSecurity": {
 "validationRules": {
 "requireSignature": true,
 "requireVersionCheck": true,
 "requirePermissionsReview": true
 },
 "sandbox": {
 "enabled": true,
 "isolatedNetwork": true,
 "fileSystemAccess": "restricted"
 },
 "monitoring": {
 "logAllCalls": true,
 "auditSensitiveOperations": true,
 "rateLimitViolations": "block"
 }
 }
}
```

## Domain-Specific Permissions

### Web Access Control

Domain-Based Web Permissions:
```json
{
 "webPermissions": {
 "allowedDomains": [
 "*.github.com",
 "*.npmjs.com",
 "docs.python.org",
 "*.company.com",
 "*.partner-site.com"
 ],
 "blockedDomains": [
 "*.malicious-site.com",
 "*.competitor.com",
 "*.social-media.com"
 ],
 "domainRestrictions": {
 "github.com": {
 "allowedPaths": ["/api/v3/", "/raw/"],
 "blockedPaths": ["/settings/", "/admin/"]
 },
 "npmjs.com": {
 "allowedPaths": ["/package/"],
 "blockedPaths": ["/settings/", "/account/"]
 }
 }
 }
}
```

### File System Access Control

Path-Based Permissions:
```json
{
 "fileSystemPermissions": {
 "allowedPaths": [
 "./src/",
 "./tests/",
 "./docs/",
 "./.claude/",
 "./.moai/"
 ],
 "blockedPaths": [
 "./.env*",
 "./secrets/",
 "./.ssh/",
 "./config/private/",
 "./node_modules/.cache/"
 ],
 "pathRestrictions": {
 "./src/": {
 "allowedExtensions": [".py", ".js", ".ts", ".md", ".json"],
 "blockedExtensions": [".exe", ".key", ".pem"]
 },
 "./config/": {
 "readOnly": true,
 "requireApproval": true
 }
 }
 }
}
```

## Permission Validation and Enforcement

### Pre-Execution Validation

Permission Check Workflow:
```python
def validate_tool_usage(tool_name, parameters, user_role):
 """
 Validate tool usage against IAM policies
 """
 # 1. Check if tool is allowed for user role
 if tool_name not in get_allowed_tools(user_role):
 return {"allowed": False, "reason": "Tool not permitted for role"}

 # 2. Check tool-specific restrictions
 restrictions = get_tool_restrictions(tool_name, user_role)
 if not validate_tool_restrictions(tool_name, parameters, restrictions):
 return {"allowed": False, "reason": "Tool restriction violation"}

 # 3. Check enterprise policy overrides
 if violates_enterprise_policy(tool_name, parameters):
 return {"allowed": False, "reason": "Enterprise policy violation"}

 # 4. Determine approval requirement
 approval_level = get_approval_level(tool_name, user_role)

 return {
 "allowed": True,
 "approvalRequired": approval_level != "none",
 "approvalLevel": approval_level
 }
```

### Real-Time Permission Monitoring

Permission Monitoring System:
```json
{
 "monitoring": {
 "realTimeValidation": {
 "enabled": true,
 "checkFrequency": "per-execution",
 "blockOnViolation": true
 },
 "auditLogging": {
 "enabled": true,
 "logLevel": "detailed",
 "retention": "90d",
 "format": "structured-json"
 },
 "alerts": {
 "permissionViolations": {
 "enabled": true,
 "channels": ["email", "slack"],
 "escalation": ["security-team", "management"]
 },
 "suspiciousActivity": {
 "enabled": true,
 "threshold": "5 violations in 1h",
 "action": "temporary-ban"
 }
 }
 }
}
```

## Security Compliance

### Compliance Framework Integration

SOC 2 Compliance:
```json
{
 "compliance": {
 "SOC2": {
 "security": {
 "accessControl": true,
 "encryptionRequired": true,
 "auditLogging": true,
 "incidentResponse": true
 },
 "availability": {
 "backupRequired": true,
 "disasterRecovery": true,
 "uptimeMonitoring": true
 },
 "processing": {
 "dataIntegrity": true,
 "accuracyValidation": true,
 "errorHandling": true
 },
 "confidentiality": {
 "dataEncryption": true,
 "accessControls": true,
 "dataMinimization": true
 }
 }
 }
}
```

ISO 27001 Compliance:
```json
{
 "compliance": {
 "ISO27001": {
 "accessControl": {
 "policyDocumented": true,
 "accessReview": "quarterly",
 "leastPrivilege": true,
 "segregationOfDuties": true
 },
 "informationSecurity": {
 "riskAssessment": "annual",
 "securityTraining": "mandatory",
 "incidentManagement": true,
 "businessContinuity": true
 }
 }
 }
}
```

## Best Practices

### Permission Management

Principle of Least Privilege:
```json
{
 "leastPrivilege": {
 "grantOnlyNecessary": true,
 "regularReview": "quarterly",
 "automaticRevocation": {
 "enabled": true,
 "inactivityPeriod": "90d"
 },
 "roleBasedAssignment": true
 }
}
```

Security Best Practices:
- Implement multi-factor authentication for administrative access
- Regular security audits and permission reviews
- Encrypted storage of sensitive configuration data
- Real-time monitoring and alerting for security events
- Incident response procedures for security violations

Compliance Best Practices:
- Document all permission policies and procedures
- Maintain comprehensive audit logs
- Regular compliance assessments and reporting
- Employee security training and awareness programs
- Automated compliance checking and validation

### Implementation Guidelines

Development Environment:
```json
{
 "development": {
 "permissionMode": "default",
 "allowedTools": ["Read", "Write", "Edit", "Bash"],
 "toolRestrictions": {
 "Bash": {"allowedCommands": ["git", "npm", "python"]},
 "Write": {"allowedPaths": ["./src/", "./tests/"]}
 }
 }
}
```

Production Environment:
```json
{
 "production": {
 "permissionMode": "restricted",
 "allowedTools": ["Read", "Grep"],
 "toolRestrictions": {
 "Read": {"allowedPaths": ["./logs/", "./config/readonly/"]}
 },
 "monitoring": {
 "realTimeAlerts": true,
 "auditAllAccess": true
 }
 }
}
```

This comprehensive IAM reference provides all the information needed to implement secure, compliant, and effective access control for Claude Code deployments at any scale.
