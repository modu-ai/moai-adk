# Skill: moai-plugin-scaffolding

**Purpose**: Provides plugin directory structure conventions and file naming standards for Alfred Framework plugins.

**Status**: Stable
**Version**: 1.0.0
**Applies to**: All Alfred plugins

---

## Quick Reference

### Standard Plugin Directory Structure

```
moai-alfred-{plugin-name}/
├── .claude-plugin/
│   ├── plugin.json              # Plugin manifest with metadata
│   ├── hooks.json              # Lifecycle hooks (optional)
│   └── permissions.json        # Permission declarations (optional)
├── {plugin_name}/              # Main plugin package
│   ├── __init__.py            # Package initialization
│   ├── commands.py            # Command implementations
│   ├── models.py              # Data models (optional)
│   └── utils.py               # Helper functions (optional)
├── commands/                   # Command documentation
│   └── {command-name}.md      # Command syntax & usage
├── agents/                     # Agent specifications
│   └── {agent-name}.md        # Agent behavior & flow
├── skills/                     # Skill definitions
│   ├── {skill-name}.md        # Skill content
│   └── examples.md            # Skill examples
├── tests/                      # Test files
│   ├── test_commands.py       # Command tests
│   ├── test_models.py         # Model tests (optional)
│   └── conftest.py            # Pytest configuration
├── README.md                   # Plugin overview
├── USAGE.md                    # Practical guide
├── CHANGELOG.md               # Version history
├── LICENSE                    # MIT License
└── pyproject.toml            # Python project config
```

### Naming Conventions

#### Plugin Names

```
moai-alfred-{domain}
moai-alfred-{domain}-{function}

Examples:
✅ moai-alfred-pm              # Project Management
✅ moai-alfred-uiux            # UI/UX
✅ moai-alfred-backend         # Backend
✅ moai-alfred-frontend        # Frontend
✅ moai-alfred-devops          # DevOps
```

#### Command Names

```
/{action}-{noun}
/{action}-{noun}-{modifier}

Examples:
✅ /init-pm                    # Initialize project management
✅ /setup-shadcn-ui           # Setup shadcn UI component library
✅ /init-fastapi              # Initialize FastAPI backend
✅ /db-setup                  # Setup database
✅ /resource-crud             # Generate CRUD resource
```

#### Agent Names

```
{plugin-name}-{role}-agent
{action}-agent

Examples:
✅ pm-agent                   # Project Management Agent
✅ ui-agent                   # UI Component Agent
✅ backend-api-agent          # Backend API Agent
✅ scaffolding-agent          # Code generation/scaffolding agent
```

#### Skill Names

```
moai-{domain}-{topic}
moai-foundation-{concept}
moai-essentials-{practice}

Examples:
✅ moai-plugin-scaffolding     # Plugin structure standards
✅ moai-pm-patterns            # Project management patterns
✅ moai-foundation-ears        # EARS requirement syntax
✅ moai-essentials-testing     # Testing best practices
```

#### File Naming

```
# Python modules: snake_case
✅ init_pm.py
✅ command_result.py
✅ validate_input.py

# Test files: test_{module}.py
✅ test_commands.py
✅ test_models.py

# Documentation: kebab-case.md
✅ init-pm.md
✅ project-charter.md
✅ best-practices.md

# Directories: kebab-case
✅ commands/
✅ agents/
✅ skills/
```

---

## Standard File Templates

### 1. Plugin Manifest (plugin.json)

```json
{
  "id": "moai-alfred-{name}",
  "name": "Plugin Display Name",
  "version": "1.0.0-dev",
  "description": "Plugin purpose and capabilities",
  "author": "Author Name",
  "license": "MIT",
  "homepage": "https://github.com/anthropics/claude-code",

  "commands": [
    {
      "name": "init-{name}",
      "title": "Initialize {Name}",
      "description": "Initialize {name} with templates",
      "enabled": true
    }
  ],

  "agents": [
    {
      "name": "{name}-agent",
      "title": "{Name} Agent",
      "description": "Specialist agent for {name} automation",
      "path": "agents/{name}-agent.md",
      "type": "specialist"
    }
  ],

  "skills": [
    {
      "name": "moai-plugin-scaffolding",
      "required": true
    },
    {
      "name": "moai-{name}-patterns",
      "required": false
    }
  ],

  "permissions": {
    "allowedTools": ["Read", "Write", "Edit", "Bash"],
    "fileAccess": ".moai/**",
    "fsWrite": true,
    "maxProjectSize": 100
  },

  "dependencies": {
    "alfred-framework": ">=1.0.0"
  }
}
```

### 2. Command Documentation Template (commands/{command}.md)

```markdown
# /{command-name}

{One-line description}

## Syntax

\`\`\`bash
/{command-name} <required-arg> [optional-arg] [options]
\`\`\`

## Arguments

- **required-arg** (required): Description and constraints

## Options

- `--option` (optional): Option description

## Examples

### Basic Usage
\`\`\`bash
/{command-name} value
\`\`\`

## What it does

1. Validates input
2. Creates resources
3. Reports completion

## Output

Creates directory structure:
\`\`\`
.moai/
├── ...
\`\`\`

## Error Handling

### Error 1
\`\`\`
❌ Error message
\`\`\`
Solution: How to fix

---

Generated with [Claude Code](https://claude.com/claude-code)
```

### 3. Agent Specification Template (agents/{agent}.md)

```markdown
# {Agent Name} Agent

Specialist agent for {purpose}.

## Responsibilities

1. **Primary Responsibility**
   - Step 1
   - Step 2

2. **Secondary Responsibility**
   - Step 1

## Interaction Flow

\`\`\`
User Input → Validation → Processing → Output
\`\`\`

## Tools

- **Read**: Access files
- **Write**: Create files
- **Edit**: Modify files
- **Bash**: Execute commands

## Skills

- skill-1: Purpose
- skill-2: Purpose

## Examples

### Example 1: Basic Usage
\`\`\`
User: /command value
Agent: [Processing description]
\`\`\`

## Error Handling

### Error Case 1
- Detection: How error identified
- Recovery: How to fix

---

Generated with [Claude Code](https://claude.com/claude-code)
```

### 4. Skill Definition Template (skills/{skill}.md)

```markdown
# Skill: {skill-name}

**Purpose**: {One-line purpose}

**Status**: Stable/Beta/Alpha
**Version**: 1.0.0
**Applies to**: {Applicable contexts}

---

## Quick Reference

[Key concepts and quick lookup table]

## Detailed Patterns

### Pattern 1: {Name}

Description of pattern with context and use cases.

**When to use**: Specific conditions

**Example**:
\`\`\`
[Concrete example]
\`\`\`

## Best Practices

1. Practice 1: Description
2. Practice 2: Description

## Common Mistakes

❌ Mistake 1: What goes wrong
✅ Fix: How to correct

## Integration Points

[How this skill integrates with framework]

---

Generated with [Claude Code](https://claude.com/claude-code)
```

### 5. Test Structure Template (tests/test_commands.py)

```python
\"\"\"
Test suite for plugin commands

@TEST:COMMAND-001 - Plugin command tests
\"\"\"

import pytest
from pathlib import Path
from typing import Dict, Any

# @CODE:COMMAND-INIT-001:TEST
class TestCommandName:
    \"\"\"Test cases for /command-name\"\"\"

    @pytest.fixture
    def temp_dir(self) -> Path:
        \"\"\"Create temporary directory for testing\"\"\"
        with tempfile.TemporaryDirectory() as tmpdir:
            yield Path(tmpdir)

    # ========== NORMAL CASES ==========

    def test_basic_functionality(self, temp_dir):
        \"\"\"
        GIVEN: Input parameters
        WHEN: Command executed
        THEN: Expected output created
        \"\"\"
        # Test implementation

    # ========== ERROR CASES ==========

    def test_invalid_input(self, temp_dir):
        \"\"\"
        GIVEN: Invalid input
        WHEN: Command executed
        THEN: Proper error raised
        \"\"\"
        with pytest.raises(ValueError):
            # Test implementation
            pass
```

### 6. Python Module Template ({plugin_name}/commands.py)

```python
\"\"\"
{Plugin Name} Commands

@CODE:{PLUGIN}-INIT-001:COMMANDS
\"\"\"

from dataclasses import dataclass
from pathlib import Path
from typing import Optional, List

# @CODE:{PLUGIN}-RESULT-001:RESULT
@dataclass
class CommandResult:
    \"\"\"Result object for command execution\"\"\"
    success: bool
    message: str
    error: Optional[str] = None

class InitCommand:
    \"\"\"
    /{command-name} command implementation

    Generates {resource type} with {key feature}
    \"\"\"

    def validate_input(self, arg: str) -> bool:
        \"\"\"
        Validate input format

        @CODE:{PLUGIN}-VALIDATE-001:VALIDATION
        \"\"\"
        if not arg:
            raise ValueError("Input cannot be empty")
        return True

    def execute(self, arg: str) -> CommandResult:
        \"\"\"
        Execute /{command-name} command

        @CODE:{PLUGIN}-EXECUTE-001:MAIN
        \"\"\"
        self.validate_input(arg)
        try:
            # Implementation
            return CommandResult(
                success=True,
                message="✅ Execution successful"
            )
        except Exception as e:
            return CommandResult(
                success=False,
                message="❌ Execution failed",
                error=str(e)
            )

# Create module-level instance
init_command = InitCommand()
```

### 7. README Template

```markdown
# {Plugin Name} Plugin

**Version**: 1.0.0-dev
**Author**: {Author}
**License**: MIT

## Overview

{Plugin description and capabilities}

## Installation

\`\`\`bash
uv pip install moai-alfred-{plugin-name}
\`\`\`

## Quick Start

\`\`\`bash
/{command} value
\`\`\`

## Features

- Feature 1
- Feature 2
- Feature 3

## Command Reference

### /{command-name}

{Description}

## Testing

\`\`\`bash
pytest tests/ -v --cov=plugin_name
\`\`\`

## Contributing

See [CONTRIBUTING.md](../../../CONTRIBUTING.md)

---

Generated with [Claude Code](https://claude.com/claude-code)
```

---

## Development Workflow

### 1. Create Plugin Directory

```bash
mkdir -p moai-alfred-{name}/{plugin_name}
mkdir -p moai-alfred-{name}/{plugin_name}/{tests,commands,agents,skills}
```

### 2. Initialize Files

- Create `.claude-plugin/plugin.json` with manifest
- Create `{plugin_name}/__init__.py` with exports
- Create `{plugin_name}/commands.py` with command classes
- Create `tests/test_commands.py` with test suite
- Create `commands/{command}.md` documentation
- Create `agents/{agent}.md` specifications
- Create `skills/{skill}.md` definitions

### 3. Follow TDD Methodology

```
RED (Tests):     Write comprehensive test suite
    ↓
GREEN (Code):    Implement minimal working solution
    ↓
REFACTOR (Docs): Polish code, enhance tests, document thoroughly
```

### 4. Maintain Consistent Structure

- All commands in single file or module
- All agents documented in agents/ directory
- All skills documented in skills/ directory
- All tests in tests/ directory with conftest.py

### 5. Documentation Standards

- **README.md**: Overview, installation, quick start (1-2KB)
- **USAGE.md**: Practical examples and workflows (3-5KB)
- **CHANGELOG.md**: Version history and features (2-3KB)
- **commands/*.md**: Complete command reference (500-1000 words each)
- **agents/*.md**: Agent specifications and flows (500-1000 words each)
- **skills/*.md**: Reusable knowledge capsules (300-500 words each)

### 6. Quality Standards

- Test coverage: ≥85%
- Type safety: 100% (mypy strict mode)
- Linting: 0 errors (ruff)
- Documentation: Comprehensive with examples
- Code style: PEP 8 with consistent formatting

---

## Common Plugin Patterns

### Pattern 1: Single Command Plugin

```
moai-alfred-{name}/
├── {plugin_name}/
│   └── commands.py        # Single command class
├── commands/
│   └── {command}.md       # Command documentation
├── tests/
│   └── test_commands.py   # Comprehensive tests
```

### Pattern 2: Multi-Command Plugin

```
moai-alfred-{name}/
├── {plugin_name}/
│   ├── commands.py        # Multiple command classes
│   ├── command1.py        # Separate modules for organization
│   └── command2.py
├── commands/
│   ├── command1.md
│   └── command2.md
├── tests/
│   ├── test_command1.py
│   └── test_command2.py
```

### Pattern 3: Plugin with Agents and Skills

```
moai-alfred-{name}/
├── {plugin_name}/
│   └── commands.py
├── commands/
│   └── {command}.md
├── agents/
│   └── {agent}.md         # Agent for complex logic
├── skills/
│   ├── {skill}.md         # Custom skills
│   └── examples.md
├── tests/
│   └── test_commands.py
```

---

## Integration Checklist

- [ ] Directory structure matches standard pattern
- [ ] File naming follows conventions (kebab-case, snake_case)
- [ ] plugin.json manifest is complete and valid
- [ ] Commands documented in commands/ directory
- [ ] Agents specified in agents/ directory
- [ ] Skills defined in skills/ directory
- [ ] Test coverage ≥85% (pytest --cov)
- [ ] Type checking passes (mypy --strict)
- [ ] Linting passes (ruff check)
- [ ] README.md provides overview and installation
- [ ] USAGE.md provides practical examples
- [ ] CHANGELOG.md documents version history
- [ ] LICENSE file included (MIT)
- [ ] All @CODE markers linked to implementation
- [ ] All @TEST markers linked to tests

---

**Skill Version**: 1.0.0
**Last Updated**: 2025-10-30
**Generated with [Claude Code](https://claude.com/claude-code)**
**Co-Authored-By**: 🎩 Alfred <alfred@mo.ai.kr>
