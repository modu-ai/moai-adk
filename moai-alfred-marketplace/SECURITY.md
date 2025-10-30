# Security Policy

**MoAI Alfred Marketplace** follows a **deny-by-default security model** to protect user systems and ensure safe plugin execution.

## 🔐 Permission Model

### Core Principle: Deny-by-Default

Every plugin operates under strict permission boundaries:

1. **Explicit Allow** - Only declared `allowedTools` are permitted
2. **Implicit Deny** - All other tools are forbidden by default
3. **Runtime Validation** - PreToolUse hooks enforce permissions before execution
4. **No Overrides** - Plugins cannot bypass permission checks

### Tool Categories

#### Safe Tools (Read-Only)
```json
{
  "allowedTools": ["Read", "Glob", "Grep"]
}
```
- **Read** - File content reading
- **Glob** - File pattern matching
- **Grep** - Content search (no modification)

#### Modification Tools
```json
{
  "allowedTools": ["Write", "Edit"]
}
```
- **Write** - Create/overwrite files
- **Edit** - Modify file contents

#### System Tools
```json
{
  "allowedTools": ["Bash"]
}
```
- **Bash** - Execute shell commands
- Requires `("command_pattern": "npm:*")` for pattern-based restrictions

#### Dangerous Tools (Generally Denied)
```json
{
  "deniedTools": ["DeleteFile", "KillProcess", "Bash(rm -rf /)"]
}
```
- **DeleteFile** - Destructive file deletion
- **KillProcess** - System process termination
- **Bash(rm -rf /)** - Pattern-specific dangerous commands

## 📋 Plugin.json Permission Schema

### Basic Example
```json
{
  "id": "moai-alfred-example",
  "version": "1.0.0",
  "permissions": {
    "allowedTools": [
      "Read",
      "Write",
      "Edit",
      "Bash(npm:*)",
      "Bash(python3:*)"
    ],
    "deniedTools": [
      "DeleteFile",
      "KillProcess"
    ]
  }
}
```

### Tool Specification Options

#### Option 1: Full Tool Access
```json
{
  "allowedTools": ["Read", "Write", "Bash"]
}
```
Grants unrestricted access to tool.

#### Option 2: Pattern-Based Restrictions
```json
{
  "allowedTools": [
    "Bash(npm:*)",      // Only npm commands
    "Bash(python3:*)",  // Only python3 execution
    "Bash(git:*)"       // Only git operations
  ]
}
```
Restricts Bash to specific command patterns.

#### Option 3: Explicit Denials
```json
{
  "allowedTools": ["Bash"],
  "deniedTools": [
    "Bash(rm:*)",
    "Bash(sudo:*)"
  ]
}
```
Allows Bash but denies specific dangerous patterns.

## 🛡️ Safety Checklist for Plugin Developers

### Before Publishing Plugin

- [ ] **Minimal Permissions** - Declare only absolutely necessary tools
- [ ] **No Destructive Tools** - Avoid `DeleteFile` unless essential
- [ ] **No System Control** - Avoid `KillProcess`
- [ ] **Pattern Restrictions** - Use `Bash(npm:*)` instead of full `Bash`
- [ ] **Documentation** - Explain why each permission is needed
- [ ] **Security Review** - Have permissions reviewed before publication

### Permission Justification Template

In your plugin's **README.md**, explain each permission:

```markdown
## Permissions Required

- **Read** - Access template files from plugin directory
- **Write** - Create scaffolded project files in user workspace
- **Bash(npm:*)** - Install npm dependencies during project setup
- **Edit** - Modify configuration files (optional)

### Why These Permissions?

- Read/Write are essential for scaffolding functionality
- npm commands only (not full Bash) - prevents accidental system commands
- Edit not required - users can modify config manually
```

## 🔒 Enforcement Mechanisms

### 1. Plugin Manifest Validation

When installing a plugin, Claude Code:
1. Parses `plugin.json`
2. Validates permission schema
3. Checks against v1.0.0 tool availability
4. Warns user of unusual permissions

### 2. PreToolUse Hook

Before any tool execution:
1. Hook checks tool name against `allowedTools`
2. Validates pattern restrictions (if applicable)
3. Denies execution if not in allow list
4. Logs denied attempt to audit trail

Example:
```typescript
// hooks.json
{
  "preToolUse": {
    "name": "enforceToolPermissions",
    "priority": 100,
    "timeout": 1000,
    "handler": "src/hooks/preToolUse.ts"
  }
}

// src/hooks/preToolUse.ts
export async function enforceToolPermissions(context: HookContext): Promise<void> {
  const toolName = context.toolCall.name;
  const { allowedTools, deniedTools } = context.plugin.permissions;

  // Check explicit allow list
  if (!allowedTools.includes(toolName) &&
      !allowedTools.some(allowed => toolMatches(toolName, allowed))) {
    throw new Error(`Tool "${toolName}" denied by plugin permissions`);
  }

  // Check explicit deny list
  if (deniedTools.includes(toolName) ||
      deniedTools.some(denied => toolMatches(toolName, denied))) {
    throw new Error(`Tool "${toolName}" explicitly denied`);
  }
}
```

### 3. Runtime Audit Logging

All plugin tool executions are logged:
- Tool name
- Plugin ID
- Timestamp
- Result (allowed/denied)
- User context

Location: `.claude/logs/plugin-audit.log`

## 🚨 Security Incidents

### Reporting Vulnerabilities

If you discover a security vulnerability in a plugin:

1. **Do NOT open a public GitHub issue**
2. **Email**: security@mo.ai.kr
3. **Include**: Plugin name, version, vulnerability description
4. **Timeline**: 90-day responsible disclosure

### Vulnerability Response

MoAI-ADK Team will:
1. Acknowledge receipt within 48 hours
2. Assess severity (Critical/High/Medium/Low)
3. Prepare patch release within 14 days
4. Publish security advisory
5. Request plugin author update

### Severity Levels

| Level | Examples | Response Time |
|-------|----------|----------------|
| **Critical** | Arbitrary code execution, data theft | 24 hours |
| **High** | Privilege escalation, denial of service | 7 days |
| **Medium** | Information disclosure, minor impact | 14 days |
| **Low** | Best practice violations | 30 days |

## 🔍 Plugin Review Process

### Required Before Publication

1. **Security Audit**
   - [ ] Permissions validated
   - [ ] No hardcoded secrets
   - [ ] Safe dependency versions
   - [ ] No network access outside documented scope

2. **Code Review**
   - [ ] Code follows best practices
   - [ ] No malicious patterns detected
   - [ ] Comments document security-critical sections
   - [ ] Third-party libraries vetted

3. **Testing**
   - [ ] Tests achieve ≥85% coverage
   - [ ] All security scenarios tested
   - [ ] Permission denials tested
   - [ ] Failure modes handled gracefully

4. **Documentation**
   - [ ] Permissions clearly documented
   - [ ] Security considerations explained
   - [ ] Known limitations listed
   - [ ] Vulnerability reporting process provided

## 📋 Official Plugins - Security Status

### v1.0.0 Plugins

| Plugin | Audit | Status | Permissions |
|--------|-------|--------|-------------|
| **moai-alfred-pm** | ✅ Passed | Safe | Read, Write, Edit |
| **moai-alfred-uiux** | ✅ Passed | Safe | Read, Write, Edit |
| **moai-alfred-frontend** | ✅ Passed | Safe | Read, Write, Edit, Bash(npm:*) |
| **moai-alfred-backend** | ✅ Passed | Safe | Read, Write, Edit, Bash(python3:*), Bash(pip:*) |
| **moai-alfred-devops** | ✅ Passed | Safe | Read, Write, Edit, Bash(npm:*), Task |

**Audit Date**: 2025-10-30
**Next Review**: 2025-12-31

## 🛠️ Security Best Practices for Users

### Installing Plugins Safely

1. **Review Permissions**
   ```bash
   /plugin info moai-alfred-pm
   # Check "permissions" section
   ```

2. **Check Source**
   - Verify GitHub repository URL
   - Check plugin author
   - Review recent commits

3. **Enable Audit Logging**
   ```bash
   /plugin audit enable
   ```

4. **Monitor Execution**
   ```bash
   # View plugin audit log
   cat .claude/logs/plugin-audit.log
   ```

### Removing Plugins

```bash
# Disable temporarily
/plugin disable moai-alfred-pm

# Uninstall completely
/plugin uninstall moai-alfred-pm
```

### Verifying Plugin Integrity

```bash
# Check plugin signature (if available)
/plugin verify moai-alfred-pm

# List all plugins and their permissions
/plugin list --verbose
```

## 🔗 Related Documents

- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Plugin development guidelines
- **[CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)** - Community standards
- **[Plugin Development Guide](./docs/plugin-development.md)** - Advanced security topics

## 📞 Contact

- **Security Team**: security@mo.ai.kr
- **GitHub Issues**: [Security label](https://github.com/moai-adk/moai-alfred-marketplace/labels/security)
- **Discussions**: [Security category](https://github.com/moai-adk/moai-alfred-marketplace/discussions?discussions_q=category%3A%22Security%22)

---

**Last Updated**: 2025-10-30
**Version**: 1.0.0

🔗 Generated with [Claude Code](https://claude.com/claude-code)
