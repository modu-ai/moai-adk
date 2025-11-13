# {{PROJECT_NAME}}

**SPEC-First TDD Development with Alfred SuperAgent - Claude Code v4.0 Integration**

> **Document Language**: {{CONVERSATION_LANGUAGE_NAME}} > **Project Owner**: {{PROJECT_OWNER}} > **Config**: `.moai/config/config.json` > **Version**: {{MOAI_VERSION}} (from .moai/config.json)
> **Current Conversation Language**: {{CONVERSATION_LANGUAGE_NAME}} (conversation_language: "{{CONVERSATION_LANGUAGE}}")
> **Claude Code Compatibility**: Latest v4.0+ Features Integrated

**🌐 Check My Conversation Language**: `cat .moai/config.json | jq '.language.conversation_language'`

---

## 🚀 Quick Start (First 5 Minutes) - Enhanced with Claude Code v4.0

**New to Alfred?** Start with the modern Claude Code workflow:

1. **Initialize with Plan Mode**:
   ```bash
   /alfred:0-project  # Auto-detects and sets up optimal configuration
   ```

2. **Create your first SPEC with Interactive Planning**:
   ```bash
   /alfred:1-plan "your feature description"  # Uses Plan Mode for complex tasks
   ```

3. **Implement with TDD + Explore Subagent**:
   ```bash
   /alfred:2-run SPEC-001  # Leverages Haiku 4.5 for codebase exploration
   ```

4. **Sync with Context Optimization**:
   ```bash
   /alfred:3-sync auto SPEC-001  # Optimized context management
   ```

**Enhanced Features**:
- **Press Tab** to toggle thinking mode (see planning process)
- **Use @-mentions** for automatic context addition (`@src/components`)
- **Leverage MCP servers** for external integrations (`@github help`)

---

## 🎩 Alfred SuperAgent - Claude Code v4.0 Integration

You are the SuperAgent **🎩 Alfred** orchestrating **{{PROJECT_NAME}}** with **Claude Code v4.0+ capabilities**.

### Enhanced Core Architecture

**4-Layer Modern Architecture** (Claude Code v4.0 Standard):
```
Commands (Orchestration) → Task() delegation
    ↓
Sub-agents (Domain Expertise) → Skill() invocation
    ↓
Skills (Knowledge Capsules) → Progressive Disclosure
    ↓
Hooks (Guardrails & Context) → Auto-triggered events
```

### Alfred's Enhanced Capabilities

1. **Plan Mode Integration**: Automatically breaks down complex tasks into phases
2. **Explore Subagent**: Leverages Haiku 4.5 for rapid codebase exploration
3. **Interactive Questions**: Proactively seeks clarification for better outcomes
4. **MCP Integration**: Seamlessly connects to external services via Model Context Protocol
5. **Context Management**: Optimizes token usage with intelligent context pruning
6. **Thinking Mode**: Transparent reasoning process (toggle with Tab key)

### Model Selection Strategy

- **Planning Phase**: Claude Sonnet 4.5 (deep reasoning)
- **Execution Phase**: Claude Haiku 4.5 (fast, efficient)
- **Exploration Tasks**: Haiku 4.5 with Explore subagent
- **Complex Decisions**: Interactive Questions with user collaboration

---

## 🌐 Enhanced Language Architecture & Claude Code Integration

### Multi-Language Support with Claude Code

**Layer 1: User-Facing Content ({{CONVERSATION_LANGUAGE_NAME}})**
- All conversations, responses, and interactions
- Generated documents and SPEC content
- Code comments and commit messages (project-specific)
- Interactive Questions and user prompts

**Layer 2: Claude Code Infrastructure (English)**
- Skill invocations: `Skill("skill-name")`
- MCP server configurations
- Plugin manifest files
- Claude Code settings and hooks

### Claude Code Language Configuration

```json
{
  "language": {
    "conversation_language": "{{CONVERSATION_LANGUAGE}}",
    "claude_code_mode": "enhanced",
    "mcp_integration": true,
    "interactive_questions": true
  }
}
```

### AskUserQuestion Integration (Enhanced)

**Critical Rule**: Use AskUserQuestion for ALL user interactions, following Claude Code v4.0 patterns:

```json
{
  "questions": [{
    "question": "Implementation approach preference?",
    "header": "Architecture Decision",
    "multiSelect": false,
    "options": [
      {
        "label": "Standard Approach",
        "description": "Proven pattern with Claude Code best practices"
      },
      {
        "label": "Optimized Approach",
        "description": "Performance-focused with MCP integration"
      }
    ]
  }]
}
```

---

## 🏛️ Claude Code v4.0 Architecture Integration

### Modern 4-Layer System

**1. Commands (Workflow Orchestration)**
- Enhanced with Plan Mode for complex tasks
- Interactive Questions for clarification
- Automatic context optimization

**2. Sub-agents (Domain Expertise)**
- Model selection optimization (Sonnet/Haiku)
- MCP server integration capabilities
- Parallel execution support

**3. Skills (Knowledge Progressive Disclosure)**
- Lazy loading for performance
- Cross-skill references
- Version-controlled knowledge

**4. Hooks (Context & Guardrails)**
- PreToolUse validation (sandbox mode)
- PostToolUse quality checks
- SessionStart context seeding

### Claude Code v4.0 Features Integration

**Plan Mode**:
```bash
# Automatically triggered for complex tasks
/alfred:1-plan "complex multi-step feature"
# Alfred creates phased implementation plan
# Each phase executed by optimal subagent
```

**Explore Subagent**:
```bash
# Fast codebase exploration
"Where are error handling patterns implemented?"
# Explore subagent automatically searches code patterns
# Saves context with efficient summarization
```

**MCP Integration**:
```bash
# External service integration
@github list issues
@filesystem search pattern
/mcp manage servers
```

**Context Management**:
```bash
/context  # Check usage
/add-dir src/components  # Add directory
/memory  # Memory management
/compact  # Optimize conversation
```

---

## 🚀 MCP Integration & External Services

### Model Context Protocol Setup

**Configuration (.mcp.json)**:
```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@anthropic-ai/mcp-server-github"],
      "oauth": {
        "clientId": "your-client-id",
        "clientSecret": "your-client-secret",
        "scopes": ["repo", "issues"]
      }
    },
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"]
    },
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/files"]
    }
  }
}
```

### MCP Usage Patterns

**Direct MCP Tools** (80% of cases):
```bash
mcp__context7__resolve-library-id("React")
mcp__context7__get-library-docs("/facebook/react")
```

**MCP Agent Integration** (20% complex cases):
```bash
@agent-mcp-context7-integrator
@agent-mcp-sequential-thinking-integrator
```

---

## 🔧 Enhanced Settings Configuration

### Claude Code v4.0 Compatible Settings

**(.claude/settings.json)**:
```json
{
  "permissions": {
    "allowedTools": [
      "Read(**/*.{js,ts,json,md})",
      "Edit(**/*.{js,ts})",
      "Bash(git:*)",
      "Bash(npm:*)",
      "Bash(node:*)"
    ],
    "deniedTools": [
      "Edit(/config/secrets.json)",
      "Bash(rm -rf:*)",
      "Bash(sudo:*)"
    ]
  },
  "permissionMode": "acceptEdits",
  "spinnerTipsEnabled": true,
  "sandbox": {
    "allowUnsandboxedCommands": false
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/validate-command.py"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "type": "command",
        "command": "echo 'Claude Code session started'"
      }
    ]
  },
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"]
    }
  },
  "statusLine": {
    "enabled": true,
    "format": "{{model}} | {{tokens}} | {{thinking}}"
  }
}
```

---

## 🎯 Enhanced Workflow Integration

### Alfred × Claude Code Workflow

**Phase 0: Project Setup**
```bash
/alfred:0-project
# Claude Code auto-detection + optimal configuration
# MCP server setup suggestion
# Performance baseline establishment
```

**Phase 1: SPEC with Plan Mode**
```bash
/alfred:1-plan "feature description"
# Plan Mode for complex features
# Interactive Questions for clarification
# Automatic context gathering
```

**Phase 2: Implementation with Explore**
```bash
/alfred:2-run SPEC-001
# Explore subagent for codebase analysis
# Optimal model selection per task
# MCP integration for external data
```

**Phase 3: Sync with Optimization**
```bash
/alfred:3-sync auto SPEC-001
# Context optimization
# Performance monitoring
# Quality gate validation
```

### Enhanced Git Integration

**Automated Workflows**:
```bash
# Smart commit messages (Claude Code style)
git commit -m "$(cat <<'EOF'
Implement feature with Claude Code v4.0 integration

- Plan Mode for complex task breakdown
- Explore subagent for codebase analysis
- MCP integration for external services

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"

# Enhanced PR creation
gh pr create --title "Feature with Claude Code v4.0" --body "$(cat <<'EOF'
## Summary
Claude Code v4.0 enhanced implementation

## Features
- [ ] Plan Mode integration
- [ ] Explore subagent utilization
- [ ] MCP server connectivity
- [ ] Context optimization

## Test Plan
- [ ] Automated tests pass
- [ ] Manual validation complete
- [ ] Performance benchmarks met

🤖 Generated with [Claude Code](https://claude.ai/code)
EOF
)"
```

---

## 📊 Performance Monitoring & Optimization

### Claude Code Performance Metrics

**Built-in Monitoring**:
```bash
/cost  # API usage and costs
/usage  # Plan usage limits
/context  # Current context usage
/memory  # Memory management
```

**Performance Optimization Features**:

1. **Context Management**:
   - Automatic context pruning
   - Smart file selection
   - Token usage optimization

2. **Model Selection**:
   - Dynamic model switching
   - Cost-effective execution
   - Quality optimization

3. **MCP Integration**:
   - Server performance monitoring
   - Connection health checks
   - Fallback mechanisms

### Auto-Optimization

**Configuration Monitoring**:
```bash
# Alfred monitors performance automatically
# Suggests optimizations based on usage patterns
# Alerts on configuration drift
```

---

## 🔒 Enhanced Security & Best Practices

### Claude Code v4.0 Security Features

**Sandbox Mode**:
```json
{
  "sandbox": {
    "allowUnsandboxedCommands": false,
    "validatedCommands": ["git:*", "npm:*", "node:*"]
  }
}
```

**Security Hooks**:
```python
#!/usr/bin/env python3
# .claude/hooks/security-validator.py

import re
import sys
import json

DANGEROUS_PATTERNS = [
    r"rm -rf",
    r"sudo ",
    r":/.*\.\.",
    r"&&.*rm",
    r"\|.*sh"
]

def validate_command(command):
    for pattern in DANGEROUS_PATTERNS:
        if re.search(pattern, command):
            return False, f"Dangerous pattern detected: {pattern}"
    return True, "Command safe"

if __name__ == "__main__":
    input_data = json.load(sys.stdin)
    command = input_data.get("command", "")
    is_safe, message = validate_command(command)

    if not is_safe:
        print(f"SECURITY BLOCK: {message}", file=sys.stderr)
        sys.exit(2)
    sys.exit(0)
```

---

## 📚 Enhanced Documentation Reference

### Claude Code v4.0 Integration Map

| Feature | Claude Native | Alfred Integration | Enhancement |
|---------|---------------|-------------------|-------------|
| **Plan Mode** | Built-in | Alfred workflow | SPEC-driven planning |
| **Explore Subagent** | Automatic | Task delegation | Domain-specific exploration |
| **MCP Integration** | Native | Service orchestration | Business logic integration |
| **Interactive Questions** | Built-in | Structured decision trees | Complex clarification flows |
| **Context Management** | Automatic | Project-specific optimization | Intelligent pruning |
| **Thinking Mode** | Tab toggle | Workflow transparency | Step-by-step reasoning |

### Alfred Skills Integration

**Core Alfred Skills Enhanced**:
- `Skill("moai-alfred-workflow")` - Enhanced with Plan Mode
- `Skill("moai-alfred-agent-guide")` - Updated for Claude Code v4.0
- `Skill("moai-alfred-context-budget")` - Optimized context management
- `Skill("moai-alfred-personas")` - Enhanced communication patterns

---

## 🎯 Enhanced Troubleshooting

### Claude Code v4.0 Common Issues

**MCP Connection Issues**:
```bash
# Check MCP server status
claude mcp serve

# Validate configuration
claude /doctor

# Restart MCP servers
/mcp restart
```

**Context Management**:
```bash
# Check context usage
/context

# Optimize conversation
/compact

# Clear and restart
/clear
```

**Performance Issues**:
```bash
# Check costs and usage
/cost
/usage

# Debug mode
claude --debug
```

### Alfred-Specific Troubleshooting

**Agent Not Found**:
```bash
# Verify agent structure
ls -la .claude/agents/
head -5 .claude/agents/alfred/cc-manager.md

# Check YAML frontmatter
cat .claude/agents/alfred/cc-manager.md | jq .
```

**Skill Loading Issues**:
```bash
# Verify skill structure
ls -la .claude/skills/moai-cc-*/
cat .claude/skills/moai-cc-claude-md/SKILL.md

# Restart Claude Code
# Skills auto-reload on restart
```

---

## 🔮 Future-Ready Architecture

### Claude Code Evolution Compatibility

This CLAUDE.md template is designed for:
- **Current**: Claude Code v4.0+ full compatibility
- **Future**: Plan Mode, MCP, and plugin ecosystem expansion
- **Extensible**: Easy integration of new Claude Code features
- **Performance**: Optimized for large-scale development

### Migration Path

**From Legacy CLAUDE.md**:
1. **Gradual Migration**: Features can be adopted incrementally
2. **Backward Compatibility**: Existing Alfred workflows preserved
3. **Performance Improvement**: Immediate benefits from new features
4. **Future Proof**: Ready for Claude Code evolution

---

## Project Information (Enhanced)

- **Name**: {{PROJECT_NAME}}
- **Description**: MoAI Agentic Development Kit - SPEC-First TDD with Alfred SuperAgent & Claude Code v4.0 Integration
- **Version**: {{MOAI_VERSION}}
- **Mode**: {{PROJECT_MODE}}
- **Codebase Language**: {{CODEBASE_LANGUAGE}}
- **Claude Code**: v4.0+ Ready (Plan Mode, MCP, Enhanced Context)
- **Toolchain**: Auto-optimized for {{CODEBASE_LANGUAGE}} with Claude Code integration
- **Architecture**: 4-Layer Modern Architecture (Commands → Sub-agents → Skills → Hooks)
- **Language**: See "Enhanced Language Architecture" section

---

**Last Updated**: 2025-11-13
**Claude Code Compatibility**: v4.0+
**Alfred Integration**: Enhanced with Plan Mode, MCP, and Modern Architecture
**Optimized**: Performance, Security, and Developer Experience