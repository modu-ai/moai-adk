# Tool Hooks Detailed Guide

Hooks that automatically execute before/after tool execution.

## Purpose

### PreToolUse Hook

**Before** tool execution:

- Block dangerous commands (git push --force, rm -rf)
- Permission validation
- Context delivery

### PostToolUse Hook

**After** tool execution:

- Result analysis
- Error detection
- Auto-fix suggestions

## PreToolUse Hook

### Blocked Commands

```bash
# Absolutely blocked
git push --force          # Force push
git reset --hard          # Hard reset
rm -rf /                  # Full deletion
chmod -R 777 /            # Full permission open

# Execute after confirmation
git rebase -i             # Interactive rebase
rm *.py                   # Multiple file deletion
```

### Permission Validation Logic

```bash
# Permission check
if command in dangerous_list:
    # Check settings.json
    if "deny" in permissions:
        → Block execution
    elif "ask" in permissions:
        → Request user confirmation
    else:
        → Allow execution
```

### Example: Git Push Validation

```bash
# When git push is executed
PreToolUse Hook execution:
1. Detect "push"
2. Check "push --force" → NO
3. Check target branch → develop (OK)
4. Check remote status → Updated
5. ✅ Execution allowed
```

## PostToolUse Hook

### Result Analysis

```bash
# After tool execution
PostToolUse Hook:
1. Check exit code
2. Analyze stdout/stderr
3. Detect side effects
4. Suggest auto-fixes
```

### Error Detection Examples

#### Bash Command Error

```bash
# User command
mkdir /Users/goos/test/nested/dir

# PreToolUse: Check parent directory → Not exists
# PostToolUse result:
:x: mkdir: cannot create directory: No such file or directory

🔧 Auto-fix suggestion:
   mkdir -p /Users/goos/test/nested/dir
```

#### Git Merge Conflict

```bash
# User command
git merge feature/auth

# PostToolUse result:
⚠️ Merge conflict detected in src/auth.py

🔧 Solution:
1. Fix conflict section
2. git add src/auth.py
3. git commit
```

### Auto-Fix Protocol

```
1️⃣ Error Analysis
   └─→ Identify cause

2️⃣ Fix Possibility Judgment
   ├─ YES → Step 3
   └─ NO → Provide guide only

3️⃣ User Confirmation
   └─→ AskUserQuestion

4️⃣ Auto-Fix Execution
   └─→ Re-execute

5️⃣ Result Validation
   └─→ Confirm success
```

## Hook Validation Rules

| Tool  | PreToolUse     | PostToolUse    |
| ----- | -------------- | -------------- |
| Bash  | Command validation | Exit code check |
| Git   | Branch check    | Merge status check |
| Read  | File path check | Encoding validation |
| Write | Path validation | Size limit    |
| Edit  | File existence check | Syntax validation |

## Hook Configuration

### .claude/settings.json

```json
{
  "hooks": {
    "pre_tool_use": {
      "enabled": true,
      "timeout": 5000,
      "dangerous_commands": [
        "git push --force",
        "git reset --hard",
        "rm -rf"
      ]
    },
    "post_tool_use": {
      "enabled": true,
      "timeout": 5000,
      "auto_fix": true,
      "error_detection": true
    }
  }
}
```

## Hook Chain Example

```
User: git push

↓ PreToolUse Hook
├─→ Detect "push"
├─→ Check branch: develop
├─→ Check force push: None
└─→ ✅ Execution allowed

↓ Git Push Execution
$ git push origin develop

↓ PostToolUse Hook
├─→ Exit code: 0 (success)
├─→ Analyze stdout
└─→ ✅ Success message

Complete!
```

## 🆘 Hook Error Handling

### Hook Itself Errors

```bash
:x: Hook execution failure
│
├─ Timeout (exceeds 5s)
│  └─→ Output warning only, execute tool
│
├─ Permission error
│  └─→ Adjust permissions and retry
│
└─ Script error
   └─→ Save log, continue
```

### Debugging

```bash
# Check Hook logs
cat ~/.claude/projects/*/hook-logs/*.log

# Disable Hooks
# .claude/settings.json:
# "hooks.enabled": false

# Disable specific Hook only
# "hooks.pre_tool_use.enabled": false
```

______________________________________________________________________

**Next**: [Hooks Overview](index.md) or [SessionStart Hook](session.md)



