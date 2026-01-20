# Windows Setup Guide

MoAI-ADK Windows installation and troubleshooting guide.

## Prerequisites

- Windows 10/11
- Node.js 18+ (for npx)
- Python 3.11+
- Claude Code CLI installed

## Quick Setup

```powershell
# Install MoAI-ADK
pip install moai-adk

# Initialize project (auto-detects Windows)
moai init
```

## Common Issues

### MCP Server "cmd /c" Warning

**Symptom:**
```
[Warning] [context7] mcpServers.context7: Windows requires 'cmd /c' wrapper to execute npx
```

**Cause:** The `.mcp.json` file uses Unix-style npx execution.

**Solution 1: Reinitialize (Recommended)**
```powershell
moai init --force
```

This regenerates `.mcp.json` with Windows-compatible settings.

**Solution 2: Manual Fix**

Edit `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "context7": {
      "command": "cmd",
      "args": ["/c", "npx", "-y", "@upstash/context7-mcp@latest"]
    }
  },
  "staggeredStartup": {
    "enabled": true,
    "delayMs": 500,
    "connectionTimeout": 15000
  }
}
```

### CLAUDE.md Size Warning

**Symptom:**
```
Large CLAUDE.md file detected (58,103 chars > 40,000)
```

**Cause:** CLAUDE.md exceeds the recommended size limit.

**Solution:**
```powershell
# Regenerate from template
moai init --force

# Or manually review and trim CLAUDE.md
```

The template CLAUDE.md is optimized at ~15,000 characters. If your file is larger, it may contain duplicated content.

### Path Issues

**Symptom:** File not found errors with Unix-style paths.

**Solution:** MoAI-ADK uses `pathlib.Path` for cross-platform compatibility. If you encounter path issues:

1. Ensure you're using the latest version:
   ```powershell
   pip install --upgrade moai-adk
   ```

2. Report the issue:
   ```powershell
   moai feedback
   ```

### PowerShell Execution Policy

**Symptom:** Scripts cannot be executed.

**Solution:**
```powershell
# Allow local scripts (run as Administrator)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

## Platform Detection

MoAI-ADK automatically detects Windows and adapts:

- **MCP Configuration**: Wraps `npx` with `cmd /c`
- **File Paths**: Uses `pathlib.Path` for cross-platform compatibility
- **Settings**: Installs Windows-specific `settings.json` when available

### Verify Platform Detection

```python
from moai_adk.core.mcp.setup import MCPSetupManager
from pathlib import Path

manager = MCPSetupManager(Path("."))
print(f"Windows detected: {manager.is_windows}")
```

## Support

- GitHub Issues: https://github.com/your-org/moai-adk/issues
- Use `/moai:9-feedback` command to report issues directly

## Version Compatibility

| MoAI-ADK Version | Windows Support |
|------------------|-----------------|
| 1.3.9+           | Full support with auto-detection |
| 1.3.0-1.3.8      | Partial (manual MCP config needed) |
| < 1.3.0          | Limited |

---

Last Updated: 2026-01-18
