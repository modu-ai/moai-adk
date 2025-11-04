# Auto-Fix & Merge Conflict Protocol

Safety protocol for Alfred when handling automatic code fixes, merge conflicts, and user approval workflows.

## Purpose

Ensures that any automatic code modifications follow a strict safety protocol to prevent unintended changes and maintain user control.

## When to Use

- **Merge Conflicts**: When git merge conflicts are detected
- **Overwritten Changes**: When recent commits may have overwritten user modifications  
- **Deprecated Code**: When outdated patterns need updating
- **Template Conflicts**: When local and package templates diverge

## Key Principles

1. **Analysis First**: Always understand the root cause before making changes
2. **User Approval**: Never modify code without explicit user consent
3. **Dual Updates**: Always update both local and template paths together
4. **Rollback Ready**: Maintain ability to revert changes if needed
5. **Transparent Documentation**: Clearly document all changes in commit messages

## Quick Start

```bash
# When Alfred detects an issue:
Skill("moai-alfred-autofixes")

# Follow the 4-step protocol:
# 1. Analysis & Reporting
# 2. User Confirmation (AskUserQuestion)
# 3. Execute Only After Approval  
# 4. Commit with Full Context
```

## Files Structure

- `SKILL.md` - Complete protocol documentation
- `examples.md` - Practical usage examples
- `reference.md` - Technical reference and patterns
- `README.md` - This quick overview

## Integration

This skill integrates with:
- All `/alfred:*` commands for conflict detection
- Git workflow for safe modifications
- Template system for synchronization
