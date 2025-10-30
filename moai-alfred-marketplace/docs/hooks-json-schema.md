# hooks.json Schema Reference

**Complete guide to hook lifecycle definitions for Alfred Framework plugins**

## Overview

Plugins can optionally define a `hooks.json` file in `.claude-plugin/` directory to register event handlers. Hooks allow plugins to react to Claude Code lifecycle events without user invocation.

## Hook Lifecycle

```
Session Start
    ↓
    [sessionStart hooks fire]
    ↓
User invokes command/tool
    ↓
    [preToolUse hooks fire]
    ↓
Tool executes
    ↓
    [postToolUse hooks fire]
    ↓
(repeat for each command/tool)
    ↓
Session ends
    ↓
    [sessionEnd hooks fire]
    ↓
Session terminated
```

## Schema Structure

### Root Object

```json
{
  "sessionStart": {
    "name": "string (required)",
    "description": "string (optional)",
    "priority": "number (optional, 0-100)",
    "timeout": "number (optional, milliseconds)",
    "conditions": {object (optional)},
    "handler": "string (optional, path to handler)"
  },
  "preToolUse": {
    "name": "string (required)",
    "description": "string (optional)",
    "priority": "number (optional, 0-100)",
    "timeout": "number (optional, milliseconds)",
    "conditions": {object (optional)},
    "handler": "string (optional, path to handler)"
  },
  "postToolUse": {
    "name": "string (required)",
    "description": "string (optional)",
    "priority": "number (optional, 0-100)",
    "timeout": "number (optional, milliseconds)",
    "conditions": {object (optional)},
    "handler": "string (optional, path to handler)"
  },
  "sessionEnd": {
    "name": "string (required)",
    "description": "string (optional)",
    "priority": "number (optional, 0-100)",
    "timeout": "number (optional, milliseconds)",
    "conditions": {object (optional)},
    "handler": "string (optional, path to handler)"
  }
}
```

## Hook Types

### sessionStart

Fires when Claude Code session begins.

**Timing**: Before user can issue first command

**Use Cases**:
- Initialize plugin state
- Load configuration
- Check prerequisites
- Display welcome message

**Context Available**:
- Plugin metadata
- User configuration
- Project information

**Example**: Load and display plugin status

### preToolUse

Fires before any tool execution.

**Timing**: User triggers tool (command, read, write, bash, etc.)

**Use Cases**:
- Validate tool call against permissions
- Log audit trail
- Block dangerous operations
- Require confirmation

**Context Available**:
- Tool name
- Tool arguments
- Plugin permissions
- User context

**Example**: Enforce permission boundaries

### postToolUse

Fires after tool execution completes.

**Timing**: Tool returns result (success or failure)

**Use Cases**:
- Cache results
- Log execution
- Cleanup resources
- Trigger follow-up actions

**Context Available**:
- Tool name
- Tool result
- Execution time
- Error information

**Example**: Log tool execution to audit trail

### sessionEnd

Fires when Claude Code session terminates.

**Timing**: End of session (user closes CLI)

**Use Cases**:
- Save plugin state
- Close connections
- Cleanup resources
- Generate summary

**Context Available**:
- Session duration
- Operations performed
- Final state

**Example**: Save session logs to disk

## Hook Object Fields

### `name` (required)

- Type: `string`
- Handler function name
- Referenced from `plugin.json` hooks section

```json
{
  "sessionStart": {
    "name": "onSessionStart"
  }
}
```

### `description` (optional)

- Type: `string`
- Human-readable description of hook
- Used for logging and debugging

```json
{
  "sessionStart": {
    "description": "Initialize plugin state on session start"
  }
}
```

### `priority` (optional)

- Type: `number`
- Range: 0-100 (higher = earlier execution)
- Default: 50
- Controls execution order when multiple plugins have same hook

```json
{
  "preToolUse": {
    "priority": 100  // Runs before plugins with priority 50
  }
}
```

**Priority Recommendations**:
- 90-100: Security checks (permissions, validation)
- 50-80: Logging, monitoring
- 10-40: Cleanup, finalization
- 0-10: Low-priority tasks

### `timeout` (optional)

- Type: `number` (milliseconds)
- Default: 5000 for sessionStart, 1000 for others
- Max: 30000 (30 seconds)
- Hook execution stops if timeout exceeded

```json
{
  "sessionStart": {
    "timeout": 3000  // 3 second timeout
  }
}
```

**Recommended Timeouts**:
- sessionStart: 3000-5000ms (initialization takes time)
- preToolUse: 500-1000ms (fast validation needed)
- postToolUse: 1000-2000ms (logging/caching)
- sessionEnd: 2000-5000ms (cleanup allowed)

### `conditions` (optional)

- Type: `object`
- Conditional execution based on environment
- Hook only fires if ALL conditions match

```json
{
  "sessionStart": {
    "conditions": {
      "minClaudeCodeVersion": "1.0.0",
      "maxClaudeCodeVersion": "2.0.0",
      "platform": "darwin",
      "environment": "production"
    }
  }
}
```

**Condition Fields**:
- `minClaudeCodeVersion`: Minimum version (semver)
- `maxClaudeCodeVersion`: Maximum version (semver)
- `platform`: OS filter (darwin, linux, win32)
- `environment`: Environment (development, production, test)

### `handler` (optional)

- Type: `string`
- Path to hook implementation file
- JavaScript/TypeScript file relative to plugin root

```json
{
  "preToolUse": {
    "handler": "src/hooks/preToolUse.ts"
  }
}
```

**Handler File Location**:
- Recommended: `src/hooks/{hookType}.ts` or `.js`
- Pattern: Named exports matching hook names

## Hook Execution Context

### sessionStart Context

```typescript
interface SessionStartContext {
  plugin: PluginMetadata
  config: PluginConfig
  project: ProjectMetadata
  environment: {
    claudeCodeVersion: string
    platform: string
    locale: string
  }
}
```

### preToolUse Context

```typescript
interface PreToolUseContext {
  plugin: PluginMetadata
  toolCall: {
    name: string
    arguments: Record<string, any>
  }
  permissions: {
    allowedTools: string[]
    deniedTools: string[]
  }
  user: {
    id: string
    email: string
  }
}
```

### postToolUse Context

```typescript
interface PostToolUseContext {
  plugin: PluginMetadata
  toolCall: {
    name: string
    arguments: Record<string, any>
  }
  result: {
    success: boolean
    data?: any
    error?: string
    executionTimeMs: number
  }
}
```

### sessionEnd Context

```typescript
interface SessionEndContext {
  plugin: PluginMetadata
  session: {
    durationMs: number
    commandsExecuted: number
    toolsInvoked: number
    state: Record<string, any>
  }
}
```

## Implementation Examples

### Example 1: Simple sessionStart Hook

**hooks.json**:
```json
{
  "sessionStart": {
    "name": "onSessionStart",
    "description": "Initialize plugin on session start",
    "priority": 50,
    "timeout": 3000
  }
}
```

**src/hooks/sessionStart.ts**:
```typescript
export async function onSessionStart(context: SessionStartContext): Promise<void> {
  console.log(`Plugin initialized for ${context.project.name}`);
  console.log(`Claude Code version: ${context.environment.claudeCodeVersion}`);

  // Check compatibility
  if (context.environment.claudeCodeVersion < "1.0.0") {
    throw new Error("Plugin requires Claude Code v1.0.0 or later");
  }

  // Load plugin state
  const pluginState = await loadPluginState();
  console.log(`Loaded state: ${JSON.stringify(pluginState)}`);
}
```

### Example 2: Permission Enforcement Hook

**hooks.json**:
```json
{
  "preToolUse": {
    "name": "enforcePermissions",
    "description": "Validate tool permissions before execution",
    "priority": 100,
    "timeout": 1000
  }
}
```

**src/hooks/preToolUse.ts**:
```typescript
export async function enforcePermissions(context: PreToolUseContext): Promise<void> {
  const { toolCall, permissions } = context;
  const toolName = toolCall.name;

  // Check if tool is in allowed list
  const isAllowed = permissions.allowedTools.some(allowed => {
    if (allowed === toolName) return true;

    // Handle pattern matching (e.g., "Bash(npm:*)")
    if (allowed.includes("*")) {
      const pattern = allowed.replace("*", ".*");
      return new RegExp(pattern).test(toolName);
    }

    return false;
  });

  if (!isAllowed) {
    throw new Error(`Tool '${toolName}' denied by plugin permissions`);
  }

  // Check explicit denials
  const isDenied = permissions.deniedTools.some(denied => {
    if (denied === toolName) return true;
    if (denied.includes("*")) {
      const pattern = denied.replace("*", ".*");
      return new RegExp(pattern).test(toolName);
    }
    return false;
  });

  if (isDenied) {
    throw new Error(`Tool '${toolName}' explicitly denied`);
  }
}
```

### Example 3: Audit Logging Hook

**hooks.json**:
```json
{
  "postToolUse": {
    "name": "logToolExecution",
    "description": "Log all tool executions to audit trail",
    "priority": 50,
    "timeout": 2000
  }
}
```

**src/hooks/postToolUse.ts**:
```typescript
import * as fs from "fs";
import * as path from "path";

const AUDIT_LOG = ".claude/logs/plugin-audit.log";

export async function logToolExecution(context: PostToolUseContext): Promise<void> {
  const logEntry = {
    timestamp: new Date().toISOString(),
    plugin: context.plugin.id,
    tool: context.toolCall.name,
    success: context.result.success,
    executionTimeMs: context.result.executionTimeMs,
    error: context.result.error || null
  };

  // Append to audit log
  const logLine = JSON.stringify(logEntry);
  fs.appendFileSync(AUDIT_LOG, logLine + "\n");
}
```

### Example 4: Conditional Hook with Timeout

**hooks.json**:
```json
{
  "sessionStart": {
    "name": "checkDatabaseConnection",
    "description": "Verify database connection on startup",
    "priority": 80,
    "timeout": 5000,
    "conditions": {
      "minClaudeCodeVersion": "1.0.0",
      "environment": "production"
    }
  }
}
```

**src/hooks/sessionStart.ts**:
```typescript
import { createPool } from "mysql2/promise";

export async function checkDatabaseConnection(
  context: SessionStartContext
): Promise<void> {
  if (!context.project.config?.databaseUrl) {
    console.warn("Database URL not configured, skipping connection check");
    return;
  }

  try {
    const pool = createPool({
      connectionLimit: 1,
      uri: context.project.config.databaseUrl,
      waitForConnections: true,
      connectionTimeout: 2000
    });

    const connection = await pool.getConnection();
    console.log("✅ Database connection verified");
    connection.release();
  } catch (error) {
    throw new Error(`Database connection failed: ${error.message}`);
  }
}
```

## Hook Best Practices

### Performance

1. **Keep hooks fast**
   - sessionStart: <3 seconds
   - preToolUse: <1 second
   - postToolUse: <2 seconds
   - sessionEnd: <5 seconds

2. **Avoid blocking operations**
   - Use async/await properly
   - Don't call slow external APIs in preToolUse
   - Defer heavy work to background tasks

3. **Set reasonable timeouts**
   - Default timeouts usually sufficient
   - Override only if necessary
   - Never exceed 30 seconds

### Error Handling

1. **Throw errors to block execution**
   - preToolUse: Throw to deny tool execution
   - sessionStart: Throw to prevent session continuation
   - postToolUse: Errors logged but don't block

2. **Provide clear error messages**
   ```typescript
   throw new Error("Database connection failed: connection timeout after 2000ms");
   ```

3. **Handle missing context gracefully**
   ```typescript
   const config = context.plugin.config || {};
   const timeout = config.timeout || 5000;
   ```

### Security

1. **Never log sensitive data**
   - Exclude passwords, tokens, API keys
   - Use masking if logging user input
   ```typescript
   const maskedKey = apiKey.substring(0, 4) + "****";
   ```

2. **Validate all inputs**
   - Check context fields exist
   - Validate tool arguments
   - Sanitize log output

3. **Use explicit denials**
   - Deny-by-default principle
   - Enumerate dangerous patterns
   - Review security regularly

## Complete Example: Full hooks.json

```json
{
  "sessionStart": {
    "name": "onSessionStart",
    "description": "Initialize plugin and verify prerequisites",
    "priority": 100,
    "timeout": 5000,
    "conditions": {
      "minClaudeCodeVersion": "1.0.0"
    },
    "handler": "src/hooks/sessionStart.ts"
  },

  "preToolUse": {
    "name": "enforcePermissions",
    "description": "Validate tool execution against plugin permissions",
    "priority": 100,
    "timeout": 1000,
    "handler": "src/hooks/preToolUse.ts"
  },

  "postToolUse": {
    "name": "logToolExecution",
    "description": "Log tool execution to audit trail",
    "priority": 50,
    "timeout": 2000,
    "handler": "src/hooks/postToolUse.ts"
  },

  "sessionEnd": {
    "name": "onSessionEnd",
    "description": "Cleanup and save plugin state",
    "priority": 50,
    "timeout": 3000,
    "handler": "src/hooks/sessionEnd.ts"
  }
}
```

## Validation Rules

### Required Fields

- ✅ At least one hook type (sessionStart, preToolUse, postToolUse, or sessionEnd)
- ✅ Each hook must have `name` field
- ✅ Handler file must exist if specified

### Optional Hooks

- ⚪ `sessionStart` - Initialize (recommended)
- ⚪ `preToolUse` - Validate/prevent (recommended for security)
- ⚪ `postToolUse` - Log/react (recommended for monitoring)
- ⚪ `sessionEnd` - Cleanup (optional)

### Constraints

| Field | Min | Max | Default |
|-------|-----|-----|---------|
| `priority` | 0 | 100 | 50 |
| `timeout` | 100ms | 30000ms | Type-dependent |
| `name` length | 1 | 100 chars | N/A |

## See Also

- [plugin.json Schema](./plugin-json-schema.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Security Policy](../SECURITY.md)
- [SPEC-CH08-001](../../.moai/specs/SPEC-CH08-001/spec.md)

---

**Version**: 1.0.0
**Last Updated**: 2025-10-30

🔗 Generated with [Claude Code](https://claude.com/claude-code)
