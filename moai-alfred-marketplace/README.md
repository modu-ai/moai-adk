# MoAI Alfred Marketplace

**Official plugin registry for Claude Code Alfred Framework**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Status: Development](https://img.shields.io/badge/Status-Development-blue.svg)](#status)
[![Version](https://img.shields.io/badge/Version-1.0.0--dev-orange.svg)](./CHANGELOG.md)

## Overview

MoAI Alfred Marketplace is the official plugin registry for **Claude Code Alfred Framework** (v1.0+). It provides a centralized catalog of community-vetted plugins that extend Claude Code functionality with specialized commands, agents, hooks, and skills.

### What is Alfred Framework?

Alfred Framework is the next-generation extensibility system for Claude Code:
- **Plugins** - Containers for commands, agents, hooks, and skills
- **Commands** - User-facing entry points (e.g., `/init-pm`, `/deploy-config`)
- **Agents** - Specialized AI actors for complex workflows
- **Hooks** - Event-driven lifecycle management (SessionStart, PreToolUse, PostToolUse)
- **Skills** - Reusable knowledge modules for contextual guidance

## 📦 Available Plugins

### v1.0 Official Plugins

| Plugin | Category | Version | Purpose |
|--------|----------|---------|---------|
| **moai-alfred-pm** | Project Management | 1.0.0-dev | Generate SPEC templates, project charters, risk matrices |
| **moai-alfred-uiux** | Design System | 1.0.0-dev | Tailwind CSS + shadcn/ui component initialization |
| **moai-alfred-frontend** | Frontend Framework | 1.0.0-dev | Next.js 16 + React 19.2 + Biome scaffolding |
| **moai-alfred-backend** | Backend Framework | 1.0.0-dev | FastAPI + uv + SQLAlchemy + Alembic scaffolding |
| **moai-alfred-devops** | Deployment | 1.0.0-dev | Vercel, Supabase, Render multi-cloud configuration |

## 🚀 Quick Start

### Prerequisites

- Claude Code v1.0.0 or later
- Python 3.11+ (for certain plugins)
- npm/yarn/pnpm/bun (for frontend plugins)

### Installation

#### Method 1: Using `/plugin` Command (Recommended)

```bash
/plugin install moai-alfred-pm
/plugin install moai-alfred-uiux
/plugin install moai-alfred-frontend
/plugin install moai-alfred-backend
/plugin install moai-alfred-devops
```

#### Method 2: Manual Installation

1. Download plugin from GitHub releases
2. Extract to `.claude/plugins/{plugin-name}/`
3. Update `.claude/settings.json`:

```json
{
  "plugins": {
    "moai-alfred-pm": {
      "version": "1.0.0-dev",
      "enabled": true
    }
  }
}
```

### Verify Installation

```bash
# List installed plugins
/plugin list

# Check plugin status
/plugin status moai-alfred-pm

# View plugin commands
/plugin help moai-alfred-pm
```

## 📋 Plugin Categories

### Project Management (moai-alfred-pm)

Kickstart project planning with EARS-based requirements:

```bash
/init-pm my-awesome-project --template=moai-spec
```

**Output**:
- `spec.md` - EARS requirement specification
- `plan.md` - Implementation plan
- `acceptance.md` - Acceptance criteria

### Design System (moai-alfred-uiux)

Initialize Tailwind CSS + shadcn/ui with 20 pre-configured components:

```bash
/setup-shadcn-ui --components=button,card,dialog,form,input
```

**Output**:
- `tailwind.config.ts` - Tailwind configuration
- `globals.css` - Global styles
- `components/ui/` - 20 shadcn/ui components

### Frontend (moai-alfred-frontend)

Scaffold Next.js 16 + React 19.2 + Biome project:

```bash
/init-next ecommerce-app --pm=bun
/biome-setup
```

**Output**:
- Full Next.js 16 App Router structure
- Biome linter configuration
- Example pages and components

### Backend (moai-alfred-backend)

Scaffold FastAPI + SQLAlchemy project:

```bash
/init-fastapi inventory-api
/db-setup --database=postgresql
/resource-crud User --from-spec=SPEC-INV-001
```

**Output**:
- FastAPI application structure
- Alembic migrations
- SQLAlchemy models
- CRUD endpoints

### DevOps (moai-alfred-devops)

Configure multi-cloud deployment:

```bash
/deploy-config --platform=vercel
/connect-vercel
/connect-supabase
```

**Supported Platforms**:
- ✅ Vercel (Next.js frontend)
- ✅ Supabase (PostgreSQL backend)
- ✅ Render (FastAPI backend)

## 🔧 Plugin Architecture

### Directory Structure

```
plugin-name/
├── .claude-plugin/
│   ├── plugin.json              # Plugin metadata & schema
│   └── hooks.json               # Hook lifecycle definitions
├── commands/
│   ├── command-1.md
│   └── command-2.md
├── agents/
│   ├── agent-1.md
│   └── agent-2.md
├── skills/
│   ├── SKILL-FEATURE-001.md
│   └── SKILL-FEATURE-002.md
├── README.md
├── USAGE.md
├── CHANGELOG.md
└── tests/
    └── test_plugin.py
```

### plugin.json Schema

Every plugin requires a `plugin.json` manifest:

```json
{
  "id": "moai-alfred-pm",
  "name": "PM Plugin",
  "version": "1.0.0-dev",
  "description": "Project Management kickoff automation",
  "author": "GOOS🪿",
  "repository": "https://github.com/moai-adk/moai-alfred-marketplace",
  "minClaudeCodeVersion": "1.0.0",
  "commands": [
    {
      "name": "init-pm",
      "description": "Initialize project management templates"
    }
  ],
  "agents": [],
  "hooks": [],
  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Bash"],
    "deniedTools": ["DeleteFile"]
  }
}
```

### hooks.json Schema

Optional hook definitions for event-driven behavior:

```json
{
  "sessionStart": {
    "name": "onSessionStart",
    "description": "Initialize plugin on session start",
    "priority": 100,
    "timeout": 5000
  },
  "preToolUse": {
    "name": "onPreToolUse",
    "description": "Validate tool permissions before execution",
    "priority": 50,
    "timeout": 1000
  }
}
```

## 📚 Documentation

- **[SECURITY.md](./SECURITY.md)** - Permission model and governance policies
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Plugin development guidelines
- **[CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)** - Community standards

## 🔐 Security

### Plugin Permissions

All plugins operate under a **deny-by-default** security model:

- **allowedTools** - Explicitly permitted tools (Read, Write, Bash, etc.)
- **deniedTools** - Explicitly forbidden tools (DeleteFile, KillProcess, etc.)

Example:
```json
{
  "permissions": {
    "allowedTools": ["Read", "Write"],
    "deniedTools": ["DeleteFile", "Bash"]
  }
}
```

### Governance

- **Security Policy**: [SECURITY.md](./SECURITY.md)
- **Code of Conduct**: [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)
- **Contribution Guide**: [CONTRIBUTING.md](./CONTRIBUTING.md)

## 📊 Statistics

| Metric | Value |
|--------|-------|
| Total Plugins | 5 |
| Stable | 0 |
| Beta | 0 |
| Development | 5 |
| Total Downloads | 0 |
| Last Updated | 2025-10-30 |

## 🗺️ Roadmap

### v1.0.0 (December 2025)
- [ ] All 5 plugins reach stable status
- [ ] 100% documentation coverage
- [ ] Integration tests for all plugins

### v1.1.0 (Q1 2026)
- [ ] CLI plugin management tool
- [ ] Plugin dependency resolver
- [ ] Marketplace UI dashboard
- [ ] Plugin analytics

## 📖 Learning Resources

### Getting Started
1. [Claude Code Official Docs](https://docs.claude.com)
2. [Alfred Framework Guide](./docs/alfred-framework-guide.md)
3. [Plugin Development Tutorial](./docs/plugin-development.md)

### Advanced Topics
- [Hook Lifecycle & Events](./docs/hooks.md)
- [Permission Model & Security](./docs/permissions.md)
- [Skill Integration for Plugins](./docs/skills.md)
- [MCP Server Integration](./docs/mcp.md)

## 🤝 Contributing

We welcome plugin contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for:

- Plugin development standards
- Testing requirements
- Documentation guidelines
- Submission process

## 📝 License

All plugins in this marketplace are licensed under the **MIT License**. See [LICENSE](./LICENSE) for details.

## 🆘 Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/moai-adk/moai-alfred-marketplace/issues)
- **Discussions**: [Community Q&A](https://github.com/moai-adk/moai-alfred-marketplace/discussions)
- **Email**: support@mo.ai.kr

## 🎊 Acknowledgments

MoAI Alfred Marketplace is maintained by the **MoAI-ADK Team** in collaboration with the Claude Code community.

- **Project Lead**: GOOS🪿 (GoosLab)
- **Alfred Framework**: MoAI-ADK v1.0+
- **Powered by**: Claude (Anthropic)

---

**Status**: Development (v1.0.0-dev)
**Last Updated**: 2025-10-30
**Claude Code Version**: 1.0.0+

🔗 Generated with [Claude Code](https://claude.com/claude-code)
