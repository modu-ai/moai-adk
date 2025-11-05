# MoAI-ADK (Agentic Development Kit)

[English](README.md) | [한국어](README.ko.md) | [日本語](README.ja.md) | [中文](README.zh.md) | **[Online Documentation](https://adk.mo.ai.kr)**

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.13+-blue)](https://www.python.org/)
[![Tests](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml/badge.svg)](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml)
[![codecov](https://codecov.io/gh/modu-ai/moai-adk/branch/develop/graph/badge.svg)](https://codecov.io/gh/modu-ai/moai-adk)
[![Coverage](https://img.shields.io/badge/coverage-87.84%25-brightgreen)](https://github.com/modu-ai/moai-adk)

> **MoAI-ADK delivers a seamless development workflow that naturally connects SPEC → TEST (TDD) → CODE → DOCUMENTATION with AI.**

> **📚 For comprehensive documentation, visit [our online documentation](https://adk.mo.ai.kr).**

---

## Quick Overview

**MoAI-ADK** is an open-source framework that revolutionizes AI-powered development with a **SPEC-First TDD** approach. Led by the Alfred SuperAgent and a team of 19 specialized AI agents, MoAI-ADK ensures every piece of code is traceable, tested, and documented.

### Core Philosophy

> **"No code without SPEC, no tests without code, no documentation without implementation"**

### Key Features

- 🎯 **SPEC-First Development**: Clear requirements before implementation
- 🧪 **Auto TDD Workflow**: RED → GREEN → REFACTOR automatically
- 🏷️ **@TAG System**: Complete traceability from requirements to code
- 🤖 **Alfred SuperAgent**: AI team that remembers your project context
- 📚 **Living Documentation**: Auto-synced docs that never drift from code
- 🔧 **56 Specialized Skills**: Domain-specific AI capabilities

---

## Installation

```bash
# Install using uv (recommended)
uv install moai-adk

# Or using pip
pip install moai-adk
```

### Quick Start

```bash
# 1. Initialize your project
moai-adk init

# 2. Create a specification
/alfred:1-plan "user authentication system"

# 3. Implement with TDD
/alfred:2-run AUTH-001

# 4. Sync documentation
/alfred:3-sync
```

---

## Core Workflow

MoAI-ADK follows a simple 4-step workflow:

1. **`/alfred:0-project`** - Project setup and configuration
2. **`/alfred:1-plan`** - Create specifications using EARS format
3. **`/alfred:2-run`** - TDD implementation (RED → GREEN → REFACTOR)
4. **`/alfred:3-sync`** - Synchronize documentation and validate

Each step builds upon the previous one, ensuring complete traceability and quality.

---

## The @TAG System

Every artifact in your project gets a unique `@TAG` identifier:

```
@SPEC:AUTH-001 (Requirements)
    ↓
@TEST:AUTH-001 (Tests)
    ↓
@CODE:AUTH-001:SERVICE (Implementation)
    ↓
@DOC:AUTH-001 (Documentation)
```

This creates complete traceability, allowing you to:
- Find all code affected by requirement changes
- Ensure no orphaned code or missing tests
- Navigate between related artifacts instantly

---

## Key Components

### Alfred SuperAgent
The orchestrator that manages 19 specialized agents and 56 skills, ensuring consistent, high-quality development.

### Specialized Agents
- **spec-builder**: Creates detailed specifications
- **code-builder**: Implements TDD workflows
- **test-engineer**: Ensures comprehensive testing
- **git-manager**: Handles version control workflows
- And 15 more domain-specific experts

### Claude Skills
56 specialized capabilities that provide:
- Domain expertise (UI/UX, backend, security)
- Technical skills (testing, documentation, deployment)
- Quality assurance (linting, validation, compliance)

---

## Why MoAI-ADK?

### Traditional AI Development Problems
- ❌ Unclear requirements leading to wrong implementations
- ❌ Missing tests causing production bugs
- ❌ Documentation drifting from code
- ❌ Lost context and repeated explanations
- ❌ Impossible impact analysis

### MoAI-ADK Solutions
- ✅ **Clear SPECs** before any code is written
- ✅ **85%+ test coverage** guaranteed through TDD
- ✅ **Auto-synced documentation** that never drifts
- ✅ **Persistent context** remembered by Alfred
- ✅ **Instant impact analysis** with @TAG system

---

## Resources & Documentation

### 🌐 Online Documentation
Visit **[https://adk.mo.ai.kr](https://adk.mo.ai.kr)** for comprehensive guides including:

- **Getting Started**: Installation and basic usage
- **Guides**: Detailed tutorials for all features
- **Reference**: Complete API documentation
- **Examples**: Real-world implementation patterns

### Core Topics in Documentation
- **SPEC System**: EARS format and specification writing
- **TDD Workflow**: Test-Driven Development with AI
- **@TAG System**: Complete traceability guide
- **Alfred Commands**: Detailed command reference
- **Skills & Agents**: Understanding AI capabilities
- **Best Practices**: Team workflows and patterns

### Community & Support
- 🏠 **GitHub**: https://github.com/modu-ai/moai-adk
- 🐛 **Issues**: https://github.com/modu-ai/moai-adk/issues
- 📦 **PyPI**: https://pypi.org/project/moai-adk/
- 📚 **Documentation**: https://adk.mo.ai.kr

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

**MoAI-ADK** — SPEC-First TDD with AI SuperAgent

> Build trustworthy AI-powered software with complete traceability, guaranteed testing, and living documentation.

> **🚀 Get started now: `uv install moai-adk` and visit [https://adk.mo.ai.kr](https://adk.mo.ai.kr)**