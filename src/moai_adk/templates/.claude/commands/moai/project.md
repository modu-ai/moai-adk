---
name: moai:project
description: "Configure GLM settings"
argument-hint: "[--glm-on|--glm-off]"
allowed-tools:
  - Read
  - Write
  - Edit
  - AskUserQuestion
---

# 🔧 MoAI Project Configuration

Quick configuration for GLM (OpenAI-compatible API) integration settings.

---

## 🎯 Usage

### GLM Configuration

```bash
/moai:project --glm-on      # Enable GLM in settings.local.json
/moai:project --glm-off     # Disable GLM in settings.local.json
```

---

## 💡 GLM Configuration Guide

### Enable GLM

Run after GLM API token setup:

```bash
/moai:project --glm-on
```

**What happens**:
1. ✅ Reads `.claude/settings.local.json`
2. ✅ Sets `glm.enabled = true`
3. ✅ Shows environment variable setup instructions
4. ✅ Provides next steps

### Disable GLM

To turn off GLM integration:

```bash
/moai:project --glm-off
```

**What happens**:
1. ✅ Reads `.claude/settings.local.json`
2. ✅ Sets `glm.enabled = false`
3. ✅ Confirms disabling message

---

## 🌍 Environment Variables for GLM

Once GLM is enabled, set these environment variables:

```bash
# .env.local or shell profile
ANTHROPIC_AUTH_TOKEN=your-glm-api-token
ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic
ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-4.5-air
ANTHROPIC_DEFAULT_SONNET_MODEL=glm-4.6
ANTHROPIC_DEFAULT_OPUS_MODEL=glm-4.6
```

---

## 📋 Implementation Notes

**Phase Implementation**:
1. Parse command arguments (--glm-on, --glm-off)
2. Load `.claude/settings.local.json`
3. Modify GLM settings as requested
4. Provide user feedback with next steps
5. Handle errors gracefully

**Files Modified**:
- `.claude/settings.local.json` - GLM enabled/disabled settings

---

## 🔗 Related Commands

- `/alfred:0-project` - Full project setup and configuration
- `/mcp` - Manage MCP servers and connections
- `moai-adk update` - Update project templates
- `.claude/settings.local.json` - User-specific settings (gitignored)

