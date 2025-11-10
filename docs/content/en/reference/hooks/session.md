# SessionStart Hook Detailed Guide

Hook that automatically executes when Claude Code session starts.

## Purpose

Automatically performs the following when session starts:

- ✅ Project status check
- ✅ Session log analysis
- ✅ Configuration validation
- ✅ Dependency status check
- ✅ Git repository status check

## Execution Content

```bash
#!/bin/bash
# SessionStart Hook

# 1. Load project metadata
config=$(cat .moai/config.json)
project_name=$(echo $config | jq -r '.project.name')

# 2. Check dependencies
python3 --version
uv --version
git --version

# 3. Check Git status
current_branch=$(git branch --show-current)
commits_ahead=$(git rev-list --count HEAD..origin/main)

# 4. Analyze session logs
last_session=$(ls -t ~/.claude/projects/*/session-*.json | head -1)
# Output log analysis results

# 5. Warn on problem detection
if [ "$commits_ahead" -gt 10 ]; then
    echo "⚠️ More than 10 commits ahead of main"
fi
```

## Session Log Analysis

SessionStart Hook analyzes logs from previous sessions:

### Analysis Items

| Item          | Analysis Content             |
| ------------- | ---------------------------- |
| **Tool Usage** | Most frequently used tools   |
| **Error Patterns** | Recurring errors             |
| **Performance** | Average execution time       |
| **Efficiency** | Success rate                 |

### Analysis Results

```
📊 MoAI-ADK Session Meta-Analysis

Tool Usage TOP 5:
1. Bash (git status) - 45 times
2. Read (file reading) - 38 times
3. Edit (file editing) - 22 times
4. Grep (search) - 18 times
5. Write (file writing) - 15 times

⚠️ Error Patterns:
- "File not found": 3 times
- "Permission denied": 1 time

:bullseye: Improvement Suggestions:
- Optimize file search with glob
- Check paths before operations
```

## Configuration Validation

```bash
# .moai/config.json validation
- Verify project metadata
- Check language settings
- Verify TRUST 5 principle settings

# .claude/settings.json validation
- Hook activation status
- Permission setting consistency
- MCP server settings
```

## Problem Detection

Problems detected by SessionStart Hook:

```
:x: Critical Issues (Work Stopped)
- .moai/ directory corruption
- config.json parsing error
- Git repository corruption

⚠️ Warnings (Proceed with Caution)
- More than 10 commits ahead
- Files with failing tests
- Coverage target not met

💡 Informational Messages
- New version released
- Recommended setting changes
- Performance optimization suggestions
```

## Hook Chain

```
SessionStart
    ├─→ Load config.json
    ├─→ Check dependencies
    ├─→ Check Git status
    ├─→ Analyze session logs
    └─→ Detect and report problems
         └─→ Critical → Stop
         └─→ Warning → Proceed + Message
         └─→ Info → Proceed + Tips
```

______________________________________________________________________

**Next**: [Tool Hooks](tool.md) or [Hooks Overview](index.md)



