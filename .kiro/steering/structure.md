# MoAI-ADK Project Structure

## Repository Organization

### Core Package Structure
```
src/moai_adk/                 # Main Python package
├── __init__.py              # Package initialization and exports
├── _version.py              # Version management and metadata
├── config.py                # Configuration classes and validation
├── cli/                     # Command-line interface modules
├── core/                    # Core functionality (security, config, templates)
├── install/                 # Installation and setup utilities
├── resources/               # Templates, agents, and static resources
└── utils/                   # Shared utilities and helpers
```

### Configuration Structure
```
.claude/                     # Claude Code integration
├── settings.json           # Permissions and hooks configuration
├── agents/moai/            # Agent definitions and workflows
├── commands/moai/          # Custom command implementations
├── hooks/moai/             # Automation hooks and guards
└── memory/                 # Shared agent memory and context

.moai/                      # MoAI-ADK project configuration
├── config.json            # Project settings and strategies
├── project/               # Project documentation
│   ├── product.md         # Product requirements and vision
│   ├── structure.md       # Architecture and organization
│   └── tech.md           # Technical stack and standards
├── specs/                 # SPEC documents (SPEC-001.md, etc.)
├── indexes/               # Traceability and TAG management
├── reports/               # Sync reports and status
├── memory/                # Agent memory and development guides
└── scripts/               # Utility scripts and automation
```

### Development Structure
```
tests/                      # Test suite organization
├── unit/                  # Unit tests for individual components
├── integration/           # Integration tests for agent workflows
├── fixtures/              # Test data and mock objects
└── conftest.py           # Pytest configuration and shared fixtures

scripts/                   # Build and maintenance scripts
├── build.sh              # Build automation
├── run-tests.sh          # Test execution with coverage
├── bump_version.py       # Version management
└── validate_*.py         # Quality assurance scripts

docs/                      # Documentation and guides
├── sections/             # Structured documentation sections
└── MOAI-ADK-*-GUIDE.md  # Comprehensive user guides
```

## Architectural Patterns

### Module Responsibilities

#### Core Modules
- **cli/**: Command-line interface and user interaction
- **core/**: Essential functionality (security, configuration, templates)
- **install/**: Project initialization and setup
- **utils/**: Shared utilities and helper functions

#### Agent System
- **project-manager**: Project setup and configuration management
- **spec-builder**: SPEC document creation and @TAG management
- **code-builder**: TDD implementation and quality assurance
- **doc-syncer**: Documentation synchronization and reporting
- **git-manager**: Version control and branch operations

### Data Flow Architecture

#### Configuration Flow
1. **User Input** → CLI commands and options
2. **Config Validation** → Pydantic models ensure data integrity
3. **Agent Dispatch** → Route commands to appropriate agents
4. **State Management** → Update .moai/ and .claude/ configurations

#### Development Workflow
1. **Specification** → EARS pattern documents in .moai/specs/
2. **Implementation** → TDD cycle with quality gates
3. **Documentation** → Living docs synchronized with code
4. **Traceability** → @TAG system maintains full audit trail

## File Naming Conventions

### Python Code
- **Modules**: `snake_case.py` (e.g., `config_manager.py`)
- **Classes**: `PascalCase` (e.g., `ConfigManager`)
- **Functions**: `snake_case` (e.g., `validate_config`)
- **Constants**: `UPPER_SNAKE_CASE` (e.g., `DEFAULT_TIMEOUT`)

### Documentation
- **SPEC Files**: `SPEC-{number}.md` (e.g., `SPEC-001.md`)
- **Project Docs**: `lowercase.md` (e.g., `product.md`, `tech.md`)
- **Guides**: `UPPERCASE-GUIDE.md` (e.g., `MOAI-ADK-0.2.2-GUIDE.md`)

### Configuration
- **JSON Config**: `config.json`, `settings.json`
- **Python Config**: `config.py`, `settings.py`
- **Environment**: `.env`, `.env.local`

## Directory Conventions

### Source Code Organization
- Keep modules focused and single-responsibility
- Maximum 300 lines per file
- Group related functionality in subdirectories
- Use `__init__.py` for clean package interfaces

### Test Organization
- Mirror source structure in tests/
- One test file per source module
- Use descriptive test class and method names
- Group fixtures in conftest.py

### Documentation Structure
- Keep README.md concise and actionable
- Detailed docs in docs/ directory
- API documentation generated from docstrings
- Examples and tutorials in separate files

## Import Conventions

### Internal Imports
```python
# Absolute imports for clarity
from moai_adk.core.config_manager import ConfigManager
from moai_adk.utils.logger import get_logger

# Relative imports only within same package
from .security import SecurityManager
from ..utils.helpers import validate_path
```

### External Dependencies
```python
# Standard library first
import asyncio
import json
from pathlib import Path
from typing import Dict, List, Optional

# Third-party libraries
import click
from pydantic import BaseModel, Field

# Local imports last
from moai_adk.config import Config
```

## Error Handling Patterns

### Exception Hierarchy
```python
class MoAIError(Exception):
    """Base exception for MoAI-ADK"""

class ConfigurationError(MoAIError):
    """Configuration validation errors"""

class AgentExecutionError(MoAIError):
    """Agent execution failures"""

class GitOperationError(MoAIError):
    """Git operation failures"""
```

### Logging Standards
- Use structured logging with contextual information
- Include agent name, command, and execution context
- Mask sensitive information (tokens, passwords)
- Provide actionable error messages for users

## Quality Gates

### Pre-commit Requirements
- All tests must pass
- Code coverage ≥ 85%
- No linting errors (ruff)
- Type checking passes (mypy)
- Security scan clean (bandit)

### Code Review Standards
- TRUST 5 principles compliance
- @TAG traceability maintained
- Documentation updated
- Test coverage for new features
- Performance impact considered