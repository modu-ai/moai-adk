# AI Models Integration Agents - Quick Start

**Agents**: `ai-codex`, `ai-gemini`
**Version**: 1.0.0
**Created**: 2025-11-23

---

## Overview

MoAI-ADK now supports integration with external AI models:

- **`ai-codex`** - OpenAI Codex CLI integration (backend-focused)
- **`ai-gemini`** - Google Gemini CLI integration (frontend-focused, multimodal)

These agents extend Alfred's capabilities by leveraging specialized AI models for specific task types.

---

## Quick Setup (5 minutes)

### Step 1: Install CLI Tools

```bash
# OpenAI Codex
npm install -g @openai/codex-cli

# Google Gemini
npm install -g @google/gemini-cli
```

### Step 2: Configure API Keys

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Google
export GOOGLE_API_KEY="AIza..."
```

### Step 3: Authenticate

```bash
# Codex
codex login
codex whoami

# Gemini
gemini auth login
gemini whoami
```

### Step 4: Update MoAI Config

Edit `.moai/config/config.json`:

```json
{
  "ai_models": {
    "preferred_model": "auto",
    "codex": {
      "enabled": true,
      "installed": true,
      "logged_in": true
    },
    "gemini": {
      "enabled": true,
      "installed": true,
      "logged_in": true
    }
  }
}
```

---

## Usage Examples

### Example 1: Backend API with Codex

```bash
# Explicit invocation
Task(
  subagent_type="ai-codex",
  prompt="Create a FastAPI REST API with user authentication"
)

# Auto-detection (if preferred_model="codex" or "auto")
User: "Create a backend API with JWT authentication"
Alfred: → Delegates to ai-codex
```

### Example 2: Frontend UI with Gemini

```bash
# Explicit invocation
Task(
  subagent_type="ai-gemini",
  prompt="Create a Next.js 14 landing page with Tailwind CSS"
)

# Auto-detection (if preferred_model="gemini" or "auto")
User: "Create a React dashboard with charts"
Alfred: → Delegates to ai-gemini
```

### Example 3: Auto Mode

```json
{
  "ai_models": {
    "preferred_model": "auto",
    "auto_detection": {
      "backend_tasks": "codex",
      "frontend_tasks": "gemini",
      "database_tasks": "codex",
      "documentation_tasks": "gemini"
    }
  }
}
```

Alfred intelligently chooses:
- Backend request → `ai-codex`
- Frontend request → `ai-gemini`
- Database request → `ai-codex`
- Documentation request → `ai-gemini`

---

## Configuration Modes

### Mode 1: Native Only (Default)

**Use Case**: No external dependencies, Claude Code only

```json
{
  "ai_models": {
    "preferred_model": "native",
    "codex": { "enabled": false },
    "gemini": { "enabled": false }
  }
}
```

**Pros**: Free, no setup, stable
**Cons**: Limited to Claude Code capabilities

---

### Mode 2: Codex Only

**Use Case**: Backend-heavy projects (Python, Node.js, Go)

```json
{
  "ai_models": {
    "preferred_model": "codex",
    "codex": {
      "enabled": true,
      "max_tokens": 4096,
      "temperature": 0.2
    },
    "gemini": { "enabled": false }
  }
}
```

**Pros**: Excellent backend code generation, Python expertise
**Cons**: API costs, backend-focused only

---

### Mode 3: Gemini Only

**Use Case**: Frontend-heavy projects (React, Next.js, Vue)

```json
{
  "ai_models": {
    "preferred_model": "gemini",
    "codex": { "enabled": false },
    "gemini": {
      "enabled": true,
      "max_tokens": 8192,
      "temperature": 0.3,
      "model_version": "gemini-pro"
    }
  }
}
```

**Pros**: Multimodal (text/image), faster, lower cost
**Cons**: Less specialized for backend tasks

---

### Mode 4: Auto (Both)

**Use Case**: Full-stack projects requiring both backend and frontend

```json
{
  "ai_models": {
    "preferred_model": "auto",
    "codex": { "enabled": true },
    "gemini": { "enabled": true },
    "auto_detection": {
      "backend_tasks": "codex",
      "frontend_tasks": "gemini"
    }
  }
}
```

**Pros**: Best of both worlds, intelligent routing
**Cons**: Higher API costs, requires both CLIs

---

## Error Handling

### Automatic Fallback

If AI model fails, MoAI-ADK automatically falls back to native Claude Code:

```
User: "Create a FastAPI project"
Alfred: Trying ai-codex...
Codex: API quota exceeded
Alfred: ✅ Fallback to native Claude Code
Alfred: ✅ Creating FastAPI project...
```

### Manual Fallback

Disable AI models anytime:

```json
{
  "ai_models": {
    "preferred_model": "native"
  }
}
```

---

## Comparison Matrix

| Feature | Native | Codex | Gemini |
|---------|--------|-------|--------|
| **Backend API** | ✅ Good | ✅✅ Excellent | ✅ Good |
| **Frontend UI** | ✅ Good | ✅ Good | ✅✅ Excellent |
| **Database** | ✅ Good | ✅✅ Excellent | ✅ Fair |
| **Documentation** | ✅ Good | ✅ Good | ✅✅ Excellent |
| **Multimodal** | ❌ No | ❌ No | ✅✅ Yes |
| **Cost** | ✅✅ Free | ❌ High | ✅ Low |
| **Speed** | ✅✅ Fast | ✅ Standard | ✅✅ Fast |
| **Setup** | ✅✅ None | ❌ Complex | ❌ Complex |

---

## Troubleshooting

### "Command not found: codex"

```bash
npm install -g @openai/codex-cli
export PATH="$PATH:$(npm bin -g)"
```

### "Authentication failed"

```bash
export OPENAI_API_KEY="sk-..."
codex login
```

### "JSON parsing error"

Check CLI version:
```bash
codex --version
gemini --version
```

Update to latest:
```bash
npm update -g @openai/codex-cli
npm update -g @google/gemini-cli
```

---

## Cost Estimation

### OpenAI Codex
- **Free Tier**: None
- **Pricing**: $0.02 per 1K tokens
- **Typical Request**: 1500 tokens = $0.03
- **Daily (50 requests)**: $1.50
- **Monthly**: ~$45

### Google Gemini
- **Free Tier**: 60 requests/minute
- **Pricing**: $0.001 per 1K tokens
- **Typical Request**: 2500 tokens = $0.0025
- **Daily (50 requests)**: $0.125
- **Monthly**: ~$3.75

### Native Claude Code
- **Cost**: Free (included)

---

## Security Best Practices

**DO**:
- ✅ Store API keys in environment variables
- ✅ Use `.env` files (never commit)
- ✅ Rotate keys regularly
- ✅ Monitor API usage
- ✅ Enable quota management

**DON'T**:
- ❌ Hardcode API keys
- ❌ Commit keys to git
- ❌ Share keys across projects
- ❌ Disable security validation

---

## Integration with Alfred Workflow

### Phase 1: /moai:0-project

During project initialization:

```
Alfred: "Would you like to enable AI model integration?"
Options:
1. Native (Claude Code only)
2. OpenAI Codex
3. Google Gemini
4. Auto (both)

User selects: Auto
Alfred: Configuring both Codex and Gemini...
Alfred: ✅ Config updated
```

### Phase 2: /moai:1-plan

AI models can assist in SPEC generation:

```
Alfred: Using Codex for backend SPEC generation...
Codex: Generated SPEC-001 with FastAPI requirements
Alfred: ✅ SPEC-001 created
```

### Phase 3: /moai:2-run

AI models integrated into TDD cycle:

```
TDD Cycle:
  RED: Native Claude Code (more accurate for tests)
  GREEN: Codex/Gemini (faster implementation)
  REFACTOR: Native Claude Code (preserve behavior)
```

---

## Migration Guide

### From Native to AI Models

1. Install CLIs
2. Configure API keys
3. Update `.moai/config/config.json`
4. Test with simple task
5. Monitor costs

### From AI Models to Native

1. Set `preferred_model: "native"`
2. No other changes needed
3. All features work

---

## Support & Documentation

- **Full Guide**: `.moai/docs/ai-models-integration-guide.md`
- **Config Schema**: `.moai/config/schema-ai-models.json`
- **Examples**: `.moai/config/examples/ai-models-config.json`
- **Feedback**: `/moai:9-feedback "AI models: ..."`

---

**Maintained By**: MoAI-ADK Development Team
**Version**: 1.0.0
**Last Updated**: 2025-11-23
