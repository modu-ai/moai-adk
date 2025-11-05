# Examples

## Basic Usage

### Detecting and Reporting Merge Conflict

```bash
# Alfred detects merge conflict in session.py
# 1. Analysis Phase
echo "Root Cause: Commit A removed function, Commit B re-added it"
echo "Files: .claude/hooks/alfred/shared/handlers/session.py"
echo "Impact: Function duplication, import errors"

# 2. User Confirmation  
AskUserQuestion(
    question="Detected merge conflict. Should I proceed with fixing?",
    options=["YES - Fix the conflict", "NO - Manual review needed"]
)

# 3. Execute Fix (if approved)
git checkout --theirs .claude/hooks/alfred/shared/handlers/session.py
git add .claude/hooks/alfred/shared/handlers/session.py
git commit -m "Fix merge conflict in session.py

🤖 Generated with Claude Code

Co-Authored-By: 🎩 Alfred@MoAI"
```

### Template Synchronization

```bash
# After fixing local files, synchronize with package templates
cp .claude/hooks/alfred/shared/handlers/session.py \
   src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/session.py

git add src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/session.py
git commit -m "Synchronize session.py fix with package templates"
```

## Common Scenarios

### Overwritten Changes Detection

```python
# Alfred detects that recent commit overwrote user changes
def detect_overwritten_changes():
    # 1. Analyze git history
    git_log = get_git_log()
    user_changes = get_user_modifications()
    
    # 2. Report findings
    report = f"""
    Detected Overwritten Changes:
    
    User Changes: {user_changes}
    Overwritten by: {git_log[-1]}
    Files Affected: {get_affected_files()}
    
    Proposed Fix: Re-apply user changes on top of latest commit
    """
    
    # 3. Get approval
    return AskUserQuestion("Should I restore your changes?")
```

### Deprecated Code Removal

```python
# Alfred finds deprecated patterns that need updating
def fix_deprecated_code():
    deprecated_patterns = [
        "old_function_call()",
        "deprecated_import",
        "legacy_syntax"
    ]
    
    for pattern in deprecated_patterns:
        if pattern in codebase:
            # 1. Report what will be changed
            report_replacement(pattern, new_syntax)
            
            # 2. Get user approval
            if get_user_approval():
                apply_replacement(pattern, new_syntax)
```

## Error Prevention

### Before Making Changes

```python
def safety_check_before_fix():
    # Always run this before any auto-fix
    checklist = [
        "Have I analyzed the root cause?",
        "Did I create a clear report?", 
        "Did I get explicit user approval?",
        "Will I update both local and templates?",
        "Is there a rollback plan?"
    ]
    
    return all(checklist)
```

### Rollback Procedure

```bash
# If auto-fix causes issues, provide rollback
function rollback_autofix() {
    echo "Rolling back last auto-fix..."
    git revert HEAD --no-edit
    echo "Rollback complete. Please review the changes."
}
```
