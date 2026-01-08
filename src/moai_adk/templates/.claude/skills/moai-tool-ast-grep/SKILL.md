---
name: "moai-tool-ast-grep"
description: "AST-based structural code search, security scanning, and refactoring using ast-grep (sg CLI). Supports 40+ languages with pattern matching and code transformation."
version: 1.1.0
category: "tool"
modularized: true
user-invocable: false
context: fork
agent: Explore
tags: ['ast', 'refactoring', 'code-search', 'lint', 'structural-search', 'security', 'codemod']
related-skills: ['moai-workflow-testing', 'moai-foundation-quality', 'moai-domain-backend', 'moai-domain-frontend']
updated: 2026-01-08
status: "active"
allowed-tools:
  - Read
  - Grep
  - Glob
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# AST-Grep Integration

Structural code search, lint, and transformation tool using Abstract Syntax Tree analysis.

## Quick Reference (30 seconds)

### What is AST-Grep?

AST-Grep (sg) is a fast, polyglot tool for structural code search and transformation. Unlike regex-based search, it understands code syntax and matches patterns based on AST structure.

### When to Use

- Searching for code patterns that regex cannot capture (e.g., nested function calls)
- Refactoring code across multiple files with semantic awareness
- Security scanning for vulnerability patterns (SQL injection, XSS, etc.)
- API migration and deprecation handling
- Enforcing code style rules at the syntax level

### Core Commands

```bash
# Pattern search
sg run --pattern 'console.log($MSG)' --lang javascript src/

# Security scan with rules
sg scan --config sgconfig.yml

# Code transformation
sg run --pattern 'foo($A)' --rewrite 'bar($A)' --lang python src/

# Test rules
sg test
```

### Pattern Syntax Basics

```
$VAR       - Matches any single AST node (meta-variable)
$$$ARGS    - Matches zero or more nodes (variadic)
$$_        - Matches any single node (anonymous)
```

### Supported Languages

Python, JavaScript, TypeScript, Go, Rust, Java, Kotlin, C, C++, Ruby, Swift, C#, PHP, Scala, Elixir, Lua, HTML, Vue, Svelte, and 30+ more.

---

## Implementation Guide (5 minutes)

### Installation

```bash
# macOS
brew install ast-grep

# npm (cross-platform)
npm install -g @ast-grep/cli

# Cargo (Rust)
cargo install ast-grep
```

### Basic Pattern Matching

#### Simple Pattern Search

```bash
# Find all console.log calls
sg run --pattern 'console.log($MSG)' --lang javascript

# Find all Python function definitions
sg run --pattern 'def $FUNC($$$ARGS): $$$BODY' --lang python

# Find React useState hooks
sg run --pattern 'useState($INIT)' --lang typescriptreact
```

#### Meta-variables

Meta-variables capture matching AST nodes:

```yaml
# $NAME - Single node capture
pattern: 'const $NAME = require($PATH)'

# $$$ARGS - Variadic capture (zero or more)
pattern: 'function $NAME($$$ARGS) { $$$BODY }'

# $$_ - Anonymous single capture (don't care)
pattern: 'if ($$_) { return $VALUE }'
```

### Code Transformation

#### Simple Rewrite

```bash
# Rename function
sg run --pattern 'oldFunc($ARGS)' --rewrite 'newFunc($ARGS)' --lang python

# Update API call
sg run --pattern 'axios.get($URL)' --rewrite 'fetch($URL)' --lang typescript
```

#### Complex Transformation with YAML Rules

```yaml
# rule.yml
id: convert-var-to-const
language: javascript
rule:
  pattern: 'var $NAME = $VALUE'
fix: 'const $NAME = $VALUE'
message: 'Prefer const over var'
severity: warning
```

```bash
sg scan --rule rule.yml src/
```

### Rule-Based Scanning

#### Configuration File (sgconfig.yml)

```yaml
ruleDirs:
  - ./rules/security
  - ./rules/quality

testConfigs:
  - ./rules/**/__tests__/*.yml

languageGlobs:
  python: ['**/*.py']
  typescript: ['**/*.ts', '**/*.tsx']
  javascript: ['**/*.js', '**/*.jsx']
```

#### Security Rule Example

```yaml
# rules/security/sql-injection.yml
id: sql-injection-risk
language: python
severity: error
message: 'Potential SQL injection vulnerability. Use parameterized queries.'
rule:
  any:
    - pattern: 'cursor.execute($QUERY % $ARGS)'
    - pattern: 'cursor.execute($QUERY.format($$$ARGS))'
    - pattern: 'cursor.execute(f"$$$SQL")'
fix: 'cursor.execute($QUERY, $ARGS)'
```

### Relational Rules

#### Inside Rule (Scoped Search)

```yaml
id: no-console-in-function
language: javascript
rule:
  pattern: 'console.log($$$ARGS)'
  inside:
    pattern: 'function $NAME($$$PARAMS) { $$$BODY }'
```

#### Has Rule (Contains Check)

```yaml
id: async-without-await
language: javascript
rule:
  pattern: 'async function $NAME($$$PARAMS) { $$$BODY }'
  not:
    has:
      pattern: 'await $EXPR'
message: 'Async function without await'
```

#### Follows/Precedes Rules

```yaml
id: missing-error-handling
language: go
rule:
  pattern: '$ERR := $CALL'
  not:
    follows:
      pattern: 'if $ERR != nil { $$$BODY }'
```

### Composite Rules

```yaml
id: complex-rule
language: typescript
rule:
  all:
    - pattern: 'useState($INIT)'
    - inside:
        pattern: 'function $COMPONENT($$$PROPS) { $$$BODY }'
    - not:
        precedes:
          pattern: 'useEffect($$$ARGS)'
```

---

## Advanced Patterns

For comprehensive documentation including:
- Complex multi-file transformations
- Custom language configuration
- CI/CD integration patterns
- Performance optimization tips

See the following module files:

- [modules/pattern-syntax.md](modules/pattern-syntax.md) - Complete pattern syntax reference
- [modules/security-rules.md](modules/security-rules.md) - Security scanning rule templates
- [modules/refactoring-patterns.md](modules/refactoring-patterns.md) - Common refactoring patterns
- [modules/language-specific.md](modules/language-specific.md) - Language-specific patterns

### Context7 Integration

For latest AST-Grep documentation:

```
Step 1: Resolve library ID
Use mcp__context7__resolve-library-id with query "ast-grep"

Step 2: Fetch documentation
Use mcp__context7__get-library-docs with the resolved library ID
```

### MoAI-ADK Integration

AST-Grep is integrated into MoAI-ADK through:

1. **Tool Registry**: Registered as AST_ANALYZER type in `tool_registry.py`
2. **PostToolUse Hook**: Automatic security scanning after Write/Edit operations
3. **Permissions**: `Bash(sg:*)` and `Bash(ast-grep:*)` auto-allowed

### Running Scans

```bash
# Scan with MoAI-ADK rules
sg scan --config .claude/skills/moai-tool-ast-grep/rules/sgconfig.yml

# Scan specific directory
sg scan --config sgconfig.yml src/

# JSON output for CI/CD
sg scan --config sgconfig.yml --json > results.json
```

---

## Works Well With

- **moai-workflow-testing** - TDD integration, test pattern detection
- **moai-foundation-quality** - TRUST 5 compliance, code quality gates
- **moai-domain-backend** - API pattern detection, security scanning
- **moai-domain-frontend** - React/Vue pattern optimization
- **moai-lang-python** - Python-specific security and style rules
- **moai-lang-typescript** - TypeScript type safety patterns

### Related Agents

- **expert-refactoring** - AST-based large-scale refactoring
- **expert-security** - Security vulnerability scanning
- **manager-quality** - Code complexity analysis
- **expert-debug** - Pattern-based debugging

---

## Reference

- [AST-Grep Official Documentation](https://ast-grep.github.io/)
- [AST-Grep GitHub Repository](https://github.com/ast-grep/ast-grep)
- [Pattern Playground](https://ast-grep.github.io/playground.html)
- [Rule Configuration Reference](https://ast-grep.github.io/reference/yaml.html)
