# Reference

## Core Protocol Steps

### 1. Analysis Phase

**Required Actions:**
- Examine git history: git log --oneline -10
- Check file differences: git diff HEAD~1..HEAD
- Identify root cause through commit analysis
- Map affected files across local and template paths

**Output Format:**
```
Detected Issue: [Merge Conflict/Overwritten Changes/Deprecated Code]

Root Cause:
- [Specific commits involved]
- [What caused the issue]

Impact Analysis:
- Files affected: [list]
- Breaking changes: [yes/no]
- User impact: [low/medium/high]

Proposed Solution:
- [Step-by-step fix plan]
- [Files to modify]
- [Verification steps]
```

### 2. User Confirmation

**AskUserQuestion Pattern:**
Use AskUserQuestion 도구 to get explicit approval:
- question: "Analysis complete. Should I proceed with the fix?"
- options: ["YES - Apply the proposed fix", "NO - Let me review manually"]

### 3. Execution Phase

**File Modification Rules:**
- Always modify both local and template paths
- Preserve user customizations when possible
- Create backup of original files
- Use atomic operations to prevent corruption

**Template Synchronization:**
```bash
# Sync local changes to package templates
LOCAL_PATH=".claude/hooks/alfred/shared/handlers/session.py"
TEMPLATE_PATH="src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/session.py"

cp "$LOCAL_PATH" "$TEMPLATE_PATH"
git add "$LOCAL_PATH" "$TEMPLATE_PATH"
```

### 4. Commit Documentation

**Required Commit Elements:**
- Clear problem description
- Root cause explanation  
- Solution approach
- Reference to original issue/commit
- Co-authorship attribution

## Detection Patterns

### Merge Conflict Indicators

```bash
# Git conflict markers in files
grep -r "<<<<<<< \|======= \|>>>>>>>" .claude/ src/

# Failed merge commits
git log --merge --oneline

# Unmerged paths
git status --porcelain | grep "^UU"
```

### Overwritten Changes Detection

```bash
# Recent commits that might overwrite user changes
git log --since="1 week ago" --author="Alfred" --oneline

# Files with recent modifications conflicting with user changes
git diff HEAD~5..HEAD --name-only
```

## Safety Mechanisms

### Pre-Fix Validation

Run all safety checks before applying fixes:
- Check git status is clean
- Verify backup available
- Confirm user approval
- Validate fix plan
- Test rollback procedure

### Rollback Procedures

```bash
# Create backup before changes
function create_backup() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    tar -czf ".moai-backups/autofix-backup-${timestamp}.tar.gz" \
        .claude/ src/moai_adk/templates/.claude/
}

# Rollback to backup
function rollback_to_backup() {
    local backup_file=$(ls -t .moai-backups/autofix-backup-*.tar.gz | head -1)
    tar -xzf "$backup_file"
    echo "Rolled back to: $backup_file"
}
```

## Integration Points

### With Alfred Commands

- `/alfred:0-project`: Template optimization workflows
- `/alfred:1-plan`: SPEC conflict detection and resolution
- `/alfred:2-run`: Code conflict detection during implementation
- `/alfred:3-sync`: Template synchronization validation

### With Git Workflow

- **Feature Branch**: Apply fixes to current branch
- **Main/Develop**: Require explicit approval before direct fixes
- **PR Conflicts**: Auto-detect and suggest fixes in pull requests

### With Template System

- **Local Changes**: Preserve user customizations
- **Package Updates**: Handle template-user conflicts
- **Version Upgrades**: Manage breaking changes safely
