# MoAI-ADK (Agentic Development Kit)

**SPEC-First TDD with AI SuperAgent & Complete Skills**

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.13+-blue)](https://www.python.org/)
[![Tests](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml/badge.svg)](https://github.com/modu-ai/moai-adk/actions/workflows/moai-gitflow.yml)
[![codecov](https://codecov.io/gh/modu-ai/moai-adk/branch/develop/graph/badge.svg)](https://codecov.io/gh/modu-ai/moai-adk)
[![Coverage](https://img.shields.io/badge/coverage-87.84%25-brightgreen)](https://github.com/modu-ai/moai-adk)

> **A complete AI-powered development ecosystem that delivers SPEC → TEST → CODE → DOCUMENTATION in perfect sync.**

---

## 🌐 Quick Links

- **📚 Comprehensive Documentation**: [https://adk.mo.ai.kr](https://adk.mo.ai.kr)
- **🚀 Quick Start**: [Installation Guide](https://adk.mo.ai.kr/getting-started)
- **📖 API Reference**: [Commands & Skills](https://adk.mo.ai.kr/api)
- **💡 Examples & Tutorials**: [Learn More](https://adk.mo.ai.kr/examples)
- **🐛 Troubleshooting**: [Help Guide](https://adk.mo.ai.kr/troubleshooting)

---

## 🎯 Core Philosophy

**"No code without SPEC, no tests without code, no documentation without implementation"**

MoAI-ADK is an open-source framework that revolutionizes AI-powered development with a **SPEC-First TDD** approach. Led by the Alfred SuperAgent and a team of specialized AI agents, MoAI-ADK ensures every piece of code is traceable, tested, and documented.

---

## Key Features

### 🎯 SPEC-First Development
- Clear requirements before implementation
- EARS format for structured specifications
- Complete traceability from requirements to code

### 🧪 Auto TDD Workflow
- RED → GREEN → REFACTOR automatically
- 87.84%+ test coverage guaranteed
- Quality gates with TRUST 5 principles

### 🏷️ @TAG System
- Complete traceability from requirements to code
- Automatic impact analysis
- Orphan detection and chain validation

### 🤖 Alfred SuperAgent
- AI team that remembers your project context
- 19 specialized agents with 73+ production-ready skills
- Progressive skill loading for optimal performance

### 📚 Living Documentation
- Auto-synced docs that never drift from code
- Real-time documentation generation
- Multi-language support (Python, TypeScript, Go, Rust, etc.)

---

## 🚀 Getting Started

### Installation

```bash
# Install using uv (recommended)
uv install moai-adk

# Or using pip
pip install moai-adk
```

### 5-Minute Quick Start

```bash
# 1. Initialize your project
moai-adk init my-project

# 2. Create a specification
/alfred:1-plan "user authentication system"

# 3. Implement with TDD
/alfred:2-run AUTH-001

# 4. Sync documentation
/alfred:3-sync
```

---

## 🎯 The 4-Step Development Workflow

MoAI-ADK follows a simple 4-step workflow:

1. **`/alfred:0-project`** - Project setup and configuration
2. **`/alfred:1-plan`** - Create specifications using EARS format
3. **`/alfred:2-run`** - TDD implementation (RED → GREEN → REFACTOR)
4. **`/alfred:3-sync`** - Synchronize documentation and validate

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

## 🏗️ Core Architecture

### Alfred SuperAgent
The orchestrator that manages 19 specialized agents and 73+ skills, ensuring consistent, high-quality development.

### Specialized Agents
- **spec-builder**: Creates detailed specifications
- **code-builder**: Implements TDD workflows
- **test-engineer**: Ensures comprehensive testing
- **git-manager**: Handles version control workflows
- **doc-syncer**: Manages documentation synchronization
- And 15+ domain-specific experts

### Claude Skills
73+ specialized capabilities that provide:
- Domain expertise (UI/UX, backend, security)
- Technical skills (testing, documentation, deployment)
- Quality assurance (linting, validation, compliance)
- Language-specific support (Python, TypeScript, Go, Rust, etc.)

---

## 🚀 Why MoAI-ADK?

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

## 🌐 Community & Support

| Resource | Description |
|----------|-------------|
| **📚 Online Documentation** | [https://adk.mo.ai.kr](https://adk.mo.ai.kr) - Comprehensive guides and tutorials |
| **🏠 GitHub Repository** | [https://github.com/modu-ai/moai-adk](https://github.com/modu-ai/moai-adk) - Source code and issues |
| **🐛 Issues & Discussions** | [https://github.com/modu-ai/moai-adk/issues](https://github.com/modu-ai/moai-adk/issues) - Bug reports and feature requests |
| **📦 PyPI Package** | [https://pypi.org/project/moai-adk/](https://pypi.org/project/moai-adk/) - Installation |

---

## 📋 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**MoAI-ADK** — SPEC-First TDD with AI SuperAgent

> Build trustworthy AI-powered software with complete traceability, guaranteed testing, and living documentation.

> **🚀 Get started now: `uv install moai-adk` and explore [https://adk.mo.ai.kr](https://adk.mo.ai.kr)**