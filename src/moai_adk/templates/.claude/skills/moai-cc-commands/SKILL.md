---
name: "moai-cc-commands"
version: "2.0.0"
created: 2025-10-22
updated: 2025-11-11
status: stable
description: Claude Code Commands system, workflow orchestration, and command-line interface patterns. Use when creating custom commands, managing workflows, or implementing CLI interfaces.
keywords: ['commands', 'cli', 'workflow', 'orchestration', 'interface']
allowed-tools: 
  - Read
  - Bash
  - Glob
---

# Claude Code Commands System

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-cc-commands |
| **Version** | 2.0.0 (2025-11-11) |
| **Allowed tools** | Read, Bash, Glob |
| **Auto-load** | On demand when command patterns detected |
| **Tier** | Claude Code (Core) |

---

## What It Does

Claude Code Commands system, workflow orchestration, and command-line interface patterns.

**Key capabilities**:
- ✅ Command creation and management
- ✅ Workflow orchestration
- ✅ CLI interface design
- ✅ Parameter handling
- ✅ Command validation

---

## When to Use

- ✅ Creating custom commands
- ✅ Managing development workflows
- ✅ Implementing CLI interfaces
- ✅ Orchestrating complex operations

---

## Core Command Patterns

### Command Architecture
1. **Command Registration**: Command discovery and loading
2. **Parameter Handling**: Input validation and processing
3. **Workflow Orchestration**: Multi-step command execution
4. **Error Handling**: Graceful failure recovery
5. **Help System**: Command documentation and usage

### Command Types
- **Utility Commands**: Helper and convenience functions
- **Workflow Commands**: Multi-step process automation
- **Integration Commands**: Third-party service interactions
- **Management Commands**: System administration tasks
- **Development Commands**: Development workflow support

---

## Dependencies

- Claude Code commands system
- CLI framework
- Parameter validation
- Workflow orchestration tools

---

## Works Well With

- `moai-cc-agents` (Command execution delegation)
- `moai-cc-hooks` (Command event handling)
- `moai-project-config-manager` (Project-specific commands)

---

## Changelog

- **v2.0.0** (2025-11-11): Added complete metadata, command architecture patterns
- **v1.0.0** (2025-10-22): Initial commands system

---

**End of Skill** | Updated 2025-11-11
