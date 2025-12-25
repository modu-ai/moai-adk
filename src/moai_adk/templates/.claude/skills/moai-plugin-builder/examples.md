# Plugin Builder Examples

## Working Plugin Examples

### Example 1: Code Quality Plugin

Purpose: Provides code quality analysis commands and agents.

Directory Structure:
```
code-quality-plugin/
  .claude-plugin/
    plugin.json
  commands/
    lint.md
    format.md
    analyze.md
  agents/
    code-reviewer.md
  skills/
    quality-patterns/
      SKILL.md
  hooks/
    hooks.json
  README.md
```

plugin.json:
```json
{
  "name": "code-quality-plugin",
  "version": "1.0.0",
  "description": "Code quality analysis and formatting tools",
  "author": {
    "name": "Quality Team",
    "email": "quality@example.com"
  },
  "keywords": ["linting", "formatting", "code-quality"],
  "commands": ["./commands"],
  "agents": ["./agents"],
  "skills": ["./skills"],
  "hooks": ["./hooks/hooks.json"]
}
```

commands/lint.md:
```markdown
---
description: Run linting checks on codebase
---

# Lint Command

Run comprehensive linting on the specified path.

Target: $1 (default: current directory)
Options: $2 (optional: --fix, --strict)

Steps:
1. Identify project type (JavaScript, Python, etc.)
2. Run appropriate linters
3. Report findings with severity levels
4. Suggest fixes for common issues

If --fix provided, automatically apply safe fixes.
```

commands/analyze.md:
```markdown
---
description: Deep code analysis for patterns and issues
---

# Analyze Command

Perform deep analysis of codebase.

Target path: $1
Analysis type: $2 (complexity, security, performance, all)

Provide:
- Complexity metrics
- Security vulnerabilities
- Performance bottlenecks
- Improvement recommendations
```

agents/code-reviewer.md:
```markdown
---
name: code-reviewer
description: Thorough code review agent for quality assurance
tools: Read, Grep, Glob
model: sonnet
skills:
  - quality-patterns
---

# Code Reviewer Agent

You are an expert code reviewer. Analyze code for:

1. Code style and consistency
2. Potential bugs and issues
3. Performance concerns
4. Security vulnerabilities
5. Best practice violations

Provide actionable feedback with specific line references.
```

hooks/hooks.json:
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "prompt",
            "prompt": "Verify the written code follows project style guidelines."
          }
        ]
      }
    ]
  }
}
```

### Example 2: Documentation Plugin

Purpose: Automated documentation generation and management.

Directory Structure:
```
docs-plugin/
  .claude-plugin/
    plugin.json
  commands/
    generate-docs.md
    update-readme.md
  agents/
    docs-writer.md
  skills/
    documentation-patterns/
      SKILL.md
  output-styles/
    docs-style.md
  README.md
```

plugin.json:
```json
{
  "name": "docs-plugin",
  "version": "1.0.0",
  "description": "Documentation generation and management",
  "commands": ["./commands"],
  "agents": ["./agents"],
  "skills": ["./skills"],
  "outputStyles": ["./output-styles"]
}
```

commands/generate-docs.md:
```markdown
---
description: Generate documentation from source code
---

# Generate Documentation

Generate comprehensive documentation for the project.

Source directory: $1 (default: ./src)
Output directory: $2 (default: ./docs)
Format: $3 (markdown, html, both)

Steps:
1. Scan source files for documentation comments
2. Extract function signatures and types
3. Generate API reference documentation
4. Create module overview pages
5. Build table of contents

Include:
- Module descriptions
- Function documentation
- Type definitions
- Usage examples
```

agents/docs-writer.md:
```markdown
---
name: docs-writer
description: Technical documentation writing specialist
tools: Read, Write, Edit, Grep, Glob
model: sonnet
skills:
  - documentation-patterns
---

# Documentation Writer Agent

You are a technical documentation specialist.

Responsibilities:
1. Write clear, concise documentation
2. Follow documentation best practices
3. Maintain consistent style and tone
4. Include practical examples
5. Keep documentation up-to-date

Documentation Standards:
- Use active voice
- Include code examples
- Provide context and rationale
- Link related documentation
```

output-styles/docs-style.md:
```markdown
---
name: docs-style
description: Documentation output formatting
---

# Documentation Style

Format all documentation with:

## Structure
- Clear hierarchical headings
- Consistent section ordering
- Logical information flow

## Content
- Brief introduction
- Detailed explanation
- Code examples
- Related links

## Tone
- Professional and clear
- Helpful and instructive
- Concise but complete
```

### Example 3: Testing Plugin with MCP Server

Purpose: Test automation with custom MCP server integration.

Directory Structure:
```
testing-plugin/
  .claude-plugin/
    plugin.json
  commands/
    run-tests.md
    coverage.md
  agents/
    test-writer.md
  servers/
    test-server.js
  .mcp.json
  README.md
```

plugin.json:
```json
{
  "name": "testing-plugin",
  "version": "1.0.0",
  "description": "Test automation and coverage tools",
  "commands": ["./commands"],
  "agents": ["./agents"],
  "mcpServers": ["./.mcp.json"]
}
```

.mcp.json:
```json
{
  "mcpServers": {
    "test-runner": {
      "command": "node",
      "args": ["${CLAUDE_PLUGIN_ROOT}/servers/test-server.js"],
      "env": {
        "PROJECT_DIR": "${CLAUDE_PROJECT_DIR}",
        "TEST_TIMEOUT": "30000"
      }
    }
  }
}
```

commands/run-tests.md:
```markdown
---
description: Execute test suite with detailed reporting
---

# Run Tests

Execute the project test suite.

Test path: $1 (default: ./tests)
Options: $2 (--watch, --coverage, --verbose)

Use the test-runner MCP server for:
- Parallel test execution
- Real-time results
- Coverage collection
- Performance metrics

Report failures with:
- Test name and location
- Expected vs actual values
- Stack trace for debugging
```

commands/coverage.md:
```markdown
---
description: Generate test coverage report
---

# Coverage Report

Generate comprehensive test coverage analysis.

Target: $1 (default: ./src)
Threshold: $2 (default: 80%)

Report includes:
- Line coverage percentage
- Branch coverage percentage
- Function coverage percentage
- Uncovered lines by file
- Coverage trend comparison
```

agents/test-writer.md:
```markdown
---
name: test-writer
description: Test case generation specialist
tools: Read, Write, Edit, Grep, Glob
model: sonnet
---

# Test Writer Agent

You are a test automation specialist.

Capabilities:
1. Generate unit tests from source code
2. Create integration test scenarios
3. Write edge case coverage
4. Design test fixtures

Testing Principles:
- One assertion per test concept
- Descriptive test names
- Arrange-Act-Assert pattern
- Minimal test dependencies
```

### Example 4: Multi-Language LSP Plugin

Purpose: Language server support for multiple languages.

Directory Structure:
```
polyglot-lsp-plugin/
  .claude-plugin/
    plugin.json
  .lsp.json
  README.md
```

plugin.json:
```json
{
  "name": "polyglot-lsp-plugin",
  "version": "1.0.0",
  "description": "Language server support for multiple languages",
  "lspServers": ["./.lsp.json"]
}
```

.lsp.json:
```json
{
  "lspServers": {
    "typescript": {
      "command": "typescript-language-server",
      "args": ["--stdio"],
      "extensionToLanguage": {
        ".ts": "typescript",
        ".tsx": "typescriptreact",
        ".js": "javascript",
        ".jsx": "javascriptreact"
      }
    },
    "python": {
      "command": "pylsp",
      "args": [],
      "extensionToLanguage": {
        ".py": "python",
        ".pyi": "python"
      },
      "env": {
        "PYTHONPATH": "${CLAUDE_PROJECT_DIR}"
      }
    },
    "rust": {
      "command": "rust-analyzer",
      "args": [],
      "extensionToLanguage": {
        ".rs": "rust"
      }
    },
    "go": {
      "command": "gopls",
      "args": ["serve"],
      "extensionToLanguage": {
        ".go": "go"
      },
      "env": {
        "GOPATH": "${HOME}/go"
      }
    }
  }
}
```

### Example 5: Git Workflow Plugin with Hooks

Purpose: Enhanced Git workflow with validation hooks.

Directory Structure:
```
git-workflow-plugin/
  .claude-plugin/
    plugin.json
  commands/
    commit.md
    pr.md
  hooks/
    hooks.json
    validate-commit.sh
    check-branch.sh
  README.md
```

plugin.json:
```json
{
  "name": "git-workflow-plugin",
  "version": "1.0.0",
  "description": "Enhanced Git workflow with validation",
  "commands": ["./commands"],
  "hooks": ["./hooks/hooks.json"]
}
```

hooks/hooks.json:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash ${CLAUDE_PLUGIN_ROOT}/hooks/check-branch.sh"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash ${CLAUDE_PLUGIN_ROOT}/hooks/validate-commit.sh"
          }
        ]
      }
    ]
  }
}
```

hooks/validate-commit.sh:
```bash
#!/bin/bash

# Validate commit message format
COMMIT_MSG="$1"

# Check for conventional commit format
if [[ ! "$COMMIT_MSG" =~ ^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?: ]]; then
  echo "WARNING: Commit message should follow conventional commit format"
  echo "Example: feat(auth): add login functionality"
fi

exit 0
```

hooks/check-branch.sh:
```bash
#!/bin/bash

# Check if on protected branch
BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null)

if [[ "$BRANCH" == "main" || "$BRANCH" == "master" ]]; then
  echo "WARNING: You are on the $BRANCH branch"
  echo "Consider creating a feature branch for changes"
fi

exit 0
```

commands/commit.md:
```markdown
---
description: Create a well-formatted commit
---

# Commit Command

Create a commit with conventional commit format.

Type: $1 (feat, fix, docs, style, refactor, test, chore)
Scope: $2 (optional component scope)
Message: $3 (commit description)

Steps:
1. Stage relevant changes
2. Review staged files
3. Generate commit message
4. Execute commit

Follow conventional commits specification:
- feat: New feature
- fix: Bug fix
- docs: Documentation
- style: Formatting
- refactor: Code restructuring
- test: Test additions
- chore: Maintenance
```

commands/pr.md:
```markdown
---
description: Create a pull request with template
---

# Pull Request Command

Create a pull request with proper template.

Title: $1
Base branch: $2 (default: main)

PR template includes:
- Summary of changes
- Type of change
- Testing performed
- Checklist items

Ensure:
- All tests pass
- Code is reviewed
- Documentation updated
- No merge conflicts
```

## Usage Examples

Installing Example Plugins:

Local installation:
```bash
claude plugin install ./code-quality-plugin
```

Using plugin commands:
```bash
# Start claude session
claude

# Use plugin commands
/code-quality-plugin:lint ./src --fix
/code-quality-plugin:analyze ./src complexity
/docs-plugin:generate-docs ./src ./docs markdown
```

Invoking plugin agents:
```bash
# List available agents
/agents

# Use code-reviewer agent
Use the code-reviewer agent to review the authentication module
```

Testing hooks:
```bash
# Enable debug to see hook execution
claude --debug

# Hooks trigger automatically on matching events
```
