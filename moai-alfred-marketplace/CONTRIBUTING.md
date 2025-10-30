# Contributing to MoAI Alfred Marketplace

Thank you for your interest in contributing to **MoAI Alfred Marketplace**! This document provides guidelines for developing, testing, and submitting plugins.

## 🎯 Code of Conduct

All contributors must follow our [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md). We're committed to providing a welcoming, inclusive community.

## 📚 Table of Contents

- [Getting Started](#getting-started)
- [Plugin Development Guide](#plugin-development-guide)
- [Testing Requirements](#testing-requirements)
- [Documentation Standards](#documentation-standards)
- [Submission Process](#submission-process)
- [Review Criteria](#review-criteria)
- [Licensing](#licensing)

## 🚀 Getting Started

### Prerequisites

- Claude Code v1.0.0 or later
- Git 2.30+
- Node.js 18+ OR Python 3.11+ (depending on plugin type)
- Basic understanding of Alfred Framework (read [README.md](./README.md))

### Development Setup

1. **Fork the repository**
   ```bash
   git clone https://github.com/YOUR-USERNAME/moai-alfred-marketplace.git
   cd moai-alfred-marketplace
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/my-awesome-plugin
   ```

3. **Create plugin directory structure**
   ```bash
   mkdir -p plugins/moai-alfred-myPlugin/{.claude-plugin,commands,agents,skills,tests}
   ```

4. **Initialize plugin metadata**
   ```bash
   cat > plugins/moai-alfred-myPlugin/.claude-plugin/plugin.json << 'EOF'
   {
     "id": "moai-alfred-myPlugin",
     "name": "My Awesome Plugin",
     "version": "0.1.0-dev",
     "description": "Brief description",
     "author": "Your Name",
     "repository": "https://github.com/YOUR-USERNAME/moai-alfred-myPlugin",
     "minClaudeCodeVersion": "1.0.0",
     "commands": [],
     "permissions": {
       "allowedTools": ["Read"],
       "deniedTools": []
     }
   }
   EOF
   ```

## 🔧 Plugin Development Guide

### Plugin Structure

```
moai-alfred-myPlugin/
├── .claude-plugin/
│   ├── plugin.json              # Plugin metadata (REQUIRED)
│   └── hooks.json               # Hook definitions (optional)
├── commands/                    # User-facing commands
│   ├── cmd-1.md
│   └── cmd-2.md
├── agents/                      # AI agents for complex workflows
│   ├── agent-1.md
│   └── agent-2.md
├── skills/                      # Reusable knowledge modules
│   ├── SKILL-FEATURE-001.md
│   └── SKILL-FEATURE-002.md
├── README.md                    # Installation & overview
├── USAGE.md                     # Usage examples
├── CHANGELOG.md                 # Version history
├── LICENSE                      # MIT License (required)
└── tests/
    ├── test_commands.py        # Command tests
    └── test_integration.py      # Integration tests
```

### plugin.json Schema

```json
{
  "id": "moai-alfred-myPlugin",
  "name": "My Plugin",
  "version": "1.0.0-dev",
  "description": "Clear description of plugin purpose",
  "author": "Your Name <email@example.com>",
  "repository": "https://github.com/yourname/moai-alfred-myPlugin",
  "license": "MIT",
  "minClaudeCodeVersion": "1.0.0",
  "commands": [
    {
      "name": "my-command",
      "description": "What this command does"
    }
  ],
  "agents": [
    {
      "name": "my-agent",
      "description": "What this agent does",
      "type": "specialist"
    }
  ],
  "hooks": {
    "sessionStart": "hooks.json#onSessionStart"
  },
  "mcpServers": [],
  "permissions": {
    "allowedTools": ["Read", "Write"],
    "deniedTools": ["DeleteFile"]
  },
  "skills": ["skill-name-1", "skill-name-2"],
  "settings": {
    "apiKey": {
      "type": "secret",
      "description": "API key for external service (optional)"
    }
  }
}
```

### Command Template

Create `commands/my-command.md`:

```markdown
# /my-command

Brief description of what the command does.

## Syntax

\`\`\`bash
/my-command [required-arg] [--optional-flag value]
\`\`\`

## Arguments

- **required-arg** (required): Description of argument

## Options

- `--flag value` (optional): Description of option

## Examples

\`\`\`bash
/my-command arg-value --flag=test
\`\`\`

## What it does

1. Step 1
2. Step 2
3. Step 3

## Output

Description of what gets created/modified.

## Related

- `moai-skill-name` (skill documentation)
- [Usage Guide](../USAGE.md)
```

### Agent Template

Create `agents/my-agent.md`:

```markdown
# My Agent

Specialist agent for [specific responsibility].

## Responsibilities

1. Parse command arguments
2. Invoke skills as needed
3. Generate output

## Tools

- Read (access files)
- Write (create files)
- Edit (modify files)

## Interaction Flow

User executes: `/my-command args`
↓
Agent receives: command + args
↓
Agent invokes: Skill("skill-name")
↓
Agent generates: Output files/report
↓
User receives: Result summary
```

### Skill Template

Create `skills/SKILL-FEATURE-001.md` (max 500 words):

```markdown
# Feature Skill

## Overview

Brief description of what developers learn from this skill.

## Key Concepts

- Concept 1
- Concept 2
- Concept 3

## Patterns

### Pattern 1: Use Case Description

When to use this pattern and why.

[Example code]

### Pattern 2: Another Pattern

[Example code]

## Checklist

- [ ] Item 1
- [ ] Item 2
- [ ] Item 3

## See Also

- Related skills
- External documentation
```

## ✅ Testing Requirements

### Minimum Coverage

- **Python plugins**: 85% test coverage (pytest)
- **JavaScript plugins**: 80% test coverage (Vitest)
- **Type safety**: 100% (mypy strict, TypeScript strict)
- **Linting**: 0 errors (ruff, Biome)

### Test Structure

```python
# tests/test_commands.py
import pytest
from moai_alfred_myPlugin import my_command

def test_my_command_creates_output():
    """Test that /my-command creates expected output."""
    result = my_command.execute(args=['test-arg'])
    assert result.success
    assert result.output_dir.exists()

def test_my_command_invalid_args():
    """Test that /my-command handles invalid arguments."""
    with pytest.raises(ValueError):
        my_command.execute(args=[])

def test_permission_denied():
    """Test that denied tools trigger errors."""
    # Verify tool restrictions work
    pass
```

### Running Tests

```bash
# Python
pytest tests/ --cov=moai_alfred_myPlugin --cov-report=html

# JavaScript
npm test -- --coverage

# Type checking
mypy moai_alfred_myPlugin --strict  # Python
tsc --noEmit                        # TypeScript
```

### Test Checklist

- [ ] All commands can be invoked
- [ ] All arguments are validated
- [ ] Permission denials work correctly
- [ ] Error messages are clear
- [ ] Output files are created correctly
- [ ] Edge cases are handled
- [ ] Dependencies are mocked properly

## 📝 Documentation Standards

### README.md Requirements

Your README.md must include:

```markdown
# Plugin Name

[Brief description]

## Installation

\`\`\`bash
/plugin install moai-alfred-myPlugin
\`\`\`

## Quick Start

[Usage example]

## Commands

- `/my-command` - [description]

## Configuration

[Optional settings]

## Permissions

- **Read** - Why this is needed
- **Write** - Why this is needed

## Development

[How to contribute/develop further]

## License

MIT

## Author

[Your name and contact]
```

### USAGE.md Requirements

Provide practical examples:

```markdown
# Usage Guide

## Basic Usage

[Example 1]

## Advanced Usage

[Example 2]

## Troubleshooting

[Common issues and solutions]

## FAQs

[Common questions]
```

### CHANGELOG.md Requirements

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

## [1.0.0] - 2025-10-30

### Added
- `/my-command` for feature X
- Permission enforcement hooks

### Changed
- Improved error messages

### Fixed
- Bug in argument parsing

## [0.1.0] - 2025-10-20

### Added
- Initial plugin structure
```

## 🔐 Security Checklist

Before submitting, verify:

- [ ] No hardcoded secrets (API keys, tokens, passwords)
- [ ] Minimal permissions (only what's needed)
- [ ] No destructive operations (DeleteFile, KillProcess)
- [ ] All external dependencies are vetted
- [ ] Code follows security best practices
- [ ] Permission justifications documented

See [SECURITY.md](./SECURITY.md) for details.

## 📋 Submission Process

### Step 1: Create Feature Branch

```bash
git checkout -b feature/moai-alfred-myPlugin
git add .
git commit -m "feat: Add moai-alfred-myPlugin"
```

### Step 2: Run Pre-submission Checks

```bash
# Tests
pytest tests/ --cov=moai_alfred_myPlugin
npm test -- --coverage

# Type checking
mypy . --strict

# Linting
ruff check .
biome lint .

# Security audit
/plugin audit moai-alfred-myPlugin
```

### Step 3: Create Pull Request

1. Push your branch
   ```bash
   git push origin feature/moai-alfred-myPlugin
   ```

2. Create PR with template:
   ```markdown
   ## Description
   Briefly describe your plugin

   ## Type of Change
   - [ ] New plugin
   - [ ] Enhancement to existing plugin

   ## Testing
   - [ ] Tests added/updated
   - [ ] Test coverage: 85%+

   ## Documentation
   - [ ] README.md updated
   - [ ] USAGE.md provided
   - [ ] CHANGELOG.md updated

   ## Security
   - [ ] Permissions reviewed
   - [ ] No hardcoded secrets
   - [ ] Code reviewed for vulnerabilities

   ## Checklist
   - [ ] Code follows standards
   - [ ] All tests pass
   - [ ] Documentation is complete
   ```

### Step 4: Address Review Feedback

Reviewers will check:
- Code quality and standards
- Test coverage
- Documentation completeness
- Security audit results
- Functionality verification

## 🎯 Review Criteria

Your plugin will be evaluated on:

| Criterion | Requirement | Notes |
|-----------|-------------|-------|
| **Functionality** | Works as described | Tests validate all features |
| **Security** | Passes security audit | See SECURITY.md |
| **Code Quality** | Follows standards | No linting errors |
| **Testing** | 85%+ coverage | All critical paths tested |
| **Documentation** | Complete & clear | README, USAGE, CHANGELOG |
| **Performance** | No timeouts | Commands complete in <5s |
| **Compatibility** | Works with v1.0.0+ | No deprecated features |

## 📄 Licensing

All plugins must be licensed under **MIT License**:

```
MIT License

Copyright (c) 2025 [Your Name]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## 🤝 Community Support

- **Questions**: [GitHub Discussions](https://github.com/moai-adk/moai-alfred-marketplace/discussions)
- **Issues**: [GitHub Issues](https://github.com/moai-adk/moai-alfred-marketplace/issues)
- **Security**: [SECURITY.md](./SECURITY.md#reporting-vulnerabilities)

## 📚 Additional Resources

- [Alfred Framework Guide](./docs/alfred-framework-guide.md)
- [Plugin Development Tutorial](./docs/plugin-development.md)
- [Hook Lifecycle & Events](./docs/hooks.md)
- [Permission Model & Security](./docs/permissions.md)

---

**Thank you for contributing to MoAI Alfred Marketplace!** 🎉

Your plugin will help make Claude Code extensible, powerful, and accessible to everyone.

🔗 Generated with [Claude Code](https://claude.com/claude-code)
