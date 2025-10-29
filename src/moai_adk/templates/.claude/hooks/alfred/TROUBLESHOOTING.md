# Alfred Hooks Troubleshooting Guide

This guide helps you diagnose and resolve common issues with MoAI-ADK's Alfred hooks system.

## Quick Diagnosis

Run this command to verify hook integrity:

```bash
# Test hook execution
echo '{"cwd": "."}' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart
```

**Expected output**: JSON with `"continue": true` and a system message
**If failed**: See error-specific sections below

---

## Error 1: "Hook not found"

### Symptom
```
error: Failed to spawn: `.claude/hooks/alfred/alfred_hooks.py`
  Caused by: No such file or directory (os error 2)
```

### Root Causes
1. **Project initialized with older MoAI-ADK version** (before hooks system)
2. **Hooks directory deleted accidentally**
3. **Template copy failed during `/alfred:0-project` initialization**
4. **Working directory mismatch** (Claude Code started from wrong directory)

### Solutions

#### Solution 1: Update Project (Recommended)
```bash
# Re-run project initialization to restore hooks
/alfred:0-project
```

#### Solution 2: Manual Template Copy
```bash
# Copy hooks from MoAI-ADK installation
python3 -c "
import moai_adk
from pathlib import Path
import shutil

template_dir = Path(moai_adk.__file__).parent / 'templates' / '.claude' / 'hooks' / 'alfred'
target_dir = Path('.claude/hooks/alfred')
target_dir.parent.mkdir(parents=True, exist_ok=True)
shutil.copytree(template_dir, target_dir, dirs_exist_ok=True)
print(f'Copied hooks to {target_dir}')
"
```

#### Solution 3: Verify File Existence
```bash
# Check all required files exist
ls -la .claude/hooks/alfred/
ls -la .claude/hooks/alfred/handlers/
ls -la .claude/hooks/alfred/core/

# Expected files:
# .claude/hooks/alfred/alfred_hooks.py
# .claude/hooks/alfred/handlers/__init__.py
# .claude/hooks/alfred/handlers/session.py
# .claude/hooks/alfred/handlers/user.py
# .claude/hooks/alfred/handlers/tool.py
# .claude/hooks/alfred/handlers/notification.py
# .claude/hooks/alfred/core/__init__.py
# .claude/hooks/alfred/core/project.py
# .claude/hooks/alfred/core/context.py
# .claude/hooks/alfred/core/checkpoint.py
```

---

## Error 2: "Import error: No module named 'handlers'"

### Symptom
```python
ImportError: No module named 'handlers'
ModuleNotFoundError: No module named 'core'
```

### Root Causes
1. **Corrupted hooks directory** (partial copy, missing `__init__.py`)
2. **Python path resolution issue** (sys.path manipulation failed)
3. **Missing handler/core modules**

### Solutions

#### Solution 1: Verify Module Structure
```bash
# Check all __init__.py files exist
find .claude/hooks/alfred -name "__init__.py"

# Expected output:
# .claude/hooks/alfred/handlers/__init__.py
# .claude/hooks/alfred/core/__init__.py
```

#### Solution 2: Test Imports Manually
```bash
cd .claude/hooks/alfred && python3 -c "
import sys
from pathlib import Path
sys.path.insert(0, str(Path.cwd()))

from handlers import handle_session_start
from core import HookResult
print('✅ Imports successful')
"
```

#### Solution 3: Re-initialize Hooks
```bash
# Force re-copy from template
/alfred:0-project update --force
```

---

## Error 3: "Hook execution timeout"

### Symptom
```
Hook timeout after 5 seconds
⚠️ Hook execution timeout - continuing without session info
```

### Root Causes
1. **Slow Git operations** (large repository, many branches)
2. **Slow file I/O** (network drive, slow disk)
3. **Heavy SPEC counting** (many `.moai/specs/` directories)
4. **Subprocess hang** (rare, usually indicates system issue)

### Solutions

#### Solution 1: Identify Slow Operations
```bash
# Time hook execution
time echo '{"cwd": "."}' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart

# If >5 seconds, check:
# - Git repository size: du -sh .git
# - SPEC count: find .moai/specs -type d -name "SPEC-*" | wc -l
# - Disk I/O: iotop (Linux) or sudo fs_usage (macOS)
```

#### Solution 2: Increase Timeout (Temporary Workaround)
```bash
# Edit alfred_hooks.py line 129
# Change: signal.alarm(5)  # 5 seconds
# To:     signal.alarm(10)  # 10 seconds

# Location: .claude/hooks/alfred/alfred_hooks.py
```

**Note**: This is a workaround. File an issue if hooks consistently timeout.

#### Solution 3: Disable Slow Features (Not Currently Supported)
**Future enhancement**: Add `.moai/config.json` option to disable expensive checks:
```json
{
  "hooks": {
    "session_start": {
      "skip_git_info": false,
      "skip_spec_count": false
    }
  }
}
```

---

## Error 4: "JSON parse error"

### Symptom
```
JSON parse error: Expecting value: line 1 column 1 (char 0)
```

### Root Causes
1. **Empty stdin** (Claude Code passed no payload)
2. **Invalid JSON format** (malformed payload)
3. **Encoding issues** (non-UTF-8 characters)

### Solutions

#### Solution 1: Test with Minimal Payload
```bash
# Empty payload (should work - returns empty dict)
echo '' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart

# Minimal valid payload
echo '{"cwd": "."}' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart
```

#### Solution 2: Check Claude Code Version
```bash
# Ensure Claude Code is up-to-date
claude-code --version

# Update if needed
pip install --upgrade claude-code
```

---

## Error 5: "Permission denied"

### Symptom
```
PermissionError: [Errno 13] Permission denied: '.claude/hooks/alfred/alfred_hooks.py'
```

### Root Causes
1. **File not executable** (missing execute permission)
2. **Directory permissions** (parent directory not readable)
3. **Ownership issues** (file owned by different user)

### Solutions

#### Solution 1: Fix Permissions
```bash
# Make executable
chmod +x .claude/hooks/alfred/alfred_hooks.py

# Fix directory permissions
chmod -R u+rX .claude/hooks/alfred/
```

#### Solution 2: Check Ownership
```bash
# Check file owner
ls -l .claude/hooks/alfred/alfred_hooks.py

# Change owner if needed (replace USER with your username)
sudo chown -R USER:USER .claude/hooks/alfred/
```

---

## Error 6: "UV environment issues"

### Symptom
```
uv: command not found
error: failed to create virtualenv
```

### Root Causes
1. **UV not installed**
2. **UV not in PATH**
3. **Virtual environment corruption**

### Solutions

#### Solution 1: Install/Update UV
```bash
# Install UV
pip install uv

# Or update
pip install --upgrade uv

# Verify
uv --version
```

#### Solution 2: Check PATH
```bash
# Find UV installation
which uv

# Add to PATH if needed
export PATH="$HOME/.local/bin:$PATH"
```

#### Solution 3: Rebuild Virtual Environment
```bash
# Remove existing environment
rm -rf .venv

# Recreate
uv venv
uv sync
```

---

## Platform-Specific Issues

### macOS: SIGALRM Not Working
**Note**: SIGALRM is fully supported on macOS. If timeout protection isn't working:

```bash
# Verify Python version (3.8+ required)
python3 --version

# Test signal module
python3 -c "import signal; print(hasattr(signal, 'SIGALRM'))"
# Expected: True
```

### Windows: SIGALRM Not Available
**Known Limitation**: SIGALRM is not available on Windows.

**Workaround**: Hooks must complete in <2 seconds (no timeout protection).

**Alternative**: Use threading-based timeout (future enhancement):
```python
import threading

def run_with_timeout(func, timeout=5):
    result = []
    def wrapper():
        result.append(func())

    thread = threading.Thread(target=wrapper)
    thread.start()
    thread.join(timeout)

    if thread.is_alive():
        raise TimeoutError("Function timeout")
    return result[0]
```

### Linux: Signal Conflicts
If other tools use SIGALRM (rare), hooks may conflict.

**Diagnosis**:
```bash
# Check for signal handlers
python3 -c "
import signal
print('Current SIGALRM handler:', signal.getsignal(signal.SIGALRM))
"
```

---

## Advanced Debugging

### Enable Hook Logging
```bash
# Create custom hook wrapper with logging
cat > .claude/hooks/alfred/debug_hooks.sh <<'EOF'
#!/bin/bash
EVENT=$1
INPUT=$(cat)

echo "[$(date)] Event: $EVENT" >> /tmp/alfred_hooks.log
echo "[$(date)] Input: $INPUT" >> /tmp/alfred_hooks.log

OUTPUT=$(echo "$INPUT" | uv run .claude/hooks/alfred/alfred_hooks.py "$EVENT" 2>&1)
EXIT_CODE=$?

echo "[$(date)] Output: $OUTPUT" >> /tmp/alfred_hooks.log
echo "[$(date)] Exit: $EXIT_CODE" >> /tmp/alfred_hooks.log

echo "$OUTPUT"
exit $EXIT_CODE
EOF

chmod +x .claude/hooks/alfred/debug_hooks.sh

# Update settings.json to use debug wrapper
# "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/debug_hooks.sh SessionStart"
```

### Monitor Hook Performance
```bash
# Benchmark all hook events
for event in SessionStart UserPromptSubmit PreToolUse PostToolUse SessionEnd; do
  echo "Testing $event..."
  time echo '{"cwd": "."}' | uv run .claude/hooks/alfred/alfred_hooks.py $event > /dev/null
done
```

### Validate Hook Output Schema
```bash
# Test output format
OUTPUT=$(echo '{"cwd": "."}' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart)

# Check required fields
echo "$OUTPUT" | python3 -c "
import json, sys
data = json.load(sys.stdin)
assert 'continue' in data, 'Missing continue field'
assert isinstance(data['continue'], bool), 'continue must be boolean'
print('✅ Valid hook output schema')
"
```

---

## Getting Help

### Collect Diagnostic Information
```bash
# Create diagnostic report
cat > /tmp/hooks_diagnostic.txt <<EOF
=== MoAI-ADK Hooks Diagnostic Report ===
Date: $(date)

=== Environment ===
OS: $(uname -s)
Python: $(python3 --version)
UV: $(uv --version 2>&1)
Claude Code: $(claude-code --version 2>&1)

=== File Structure ===
$(find .claude/hooks/alfred -type f 2>&1)

=== Permissions ===
$(ls -la .claude/hooks/alfred/ 2>&1)

=== Hook Test ===
$(echo '{"cwd": "."}' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart 2>&1)

=== Settings ===
$(cat .claude/settings.json 2>&1 | grep -A 10 "SessionStart")
EOF

cat /tmp/hooks_diagnostic.txt
```

### Report Issues
1. **GitHub Issues**: https://github.com/modu-ai/moai-adk/issues
2. **Include**: Diagnostic report (above) + error message + steps to reproduce
3. **Security**: Do NOT include secrets, API keys, or sensitive paths

---

## Preventive Maintenance

### Regular Health Checks
```bash
# Add to .git/hooks/post-merge
cat > .git/hooks/post-merge <<'EOF'
#!/bin/bash
# Verify hooks after pulling updates
echo '{"cwd": "."}' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "⚠️ Alfred hooks failed - run /alfred:0-project update"
fi
EOF

chmod +x .git/hooks/post-merge
```

### Keep MoAI-ADK Updated
```bash
# Check version
python3 -c "import moai_adk; print(moai_adk.__version__)"

# Update
pip install --upgrade moai-adk

# Re-initialize project
/alfred:0-project update
```

---

**Last Updated**: 2025-10-29
**Applies To**: MoAI-ADK v0.7.0+
**Hooks Architecture Version**: Hybrid Modular (9 modules)
