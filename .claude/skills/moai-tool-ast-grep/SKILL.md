---
name: moai-tool-ast-grep
description: >
  AST-based structural code search, security scanning, and refactoring using ast-grep
  (sg CLI) with pattern matching and code transformation across 40+ languages.
  Use when performing structural code search, AST-based refactoring, codemod operations,
  security pattern scanning, or syntax-aware code transformations across files.
  Do NOT use for simple text search (use Grep tool instead)
  or full codebase exploration (use Explore agent instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Bash(sg:*) Bash(ast-grep:*) mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.2.0"
  category: "tool"
  modularized: "true"
  status: "active"
  updated: "2026-01-11"
  tags: "ast, refactoring, code-search, lint, structural-search, security, codemod"
  related-skills: "moai-workflow-testing, moai-foundation-quality, moai-domain-backend, moai-domain-frontend"
  context: "fork"
  agent: "Explore"

# MoAI Extension: Triggers
triggers:
  keywords: ["ast", "refactoring", "code search", "lint", "structural search", "security", "codemod", "ast-grep"]
---

# AST-Grep Integration

Structural code search, lint, and transformation tool using Abstract Syntax Tree analysis.

## Quick Reference

### What is AST-Grep

AST-Grep (sg) is a fast, polyglot tool for structural code search and transformation. Unlike regex-based search, it understands code syntax and matches patterns based on AST structure.

### When to Use

- Searching for code patterns that regex cannot capture such as nested function calls
- Refactoring code across multiple files with semantic awareness
- Security scanning for vulnerability patterns including SQL injection and XSS
- API migration and deprecation handling
- Enforcing code style rules at the syntax level

### Core Commands

Pattern search: Execute sg run with pattern option specifying the code pattern to find, lang option for the programming language, and the source directory path.

Security scan with rules: Execute sg scan with config option pointing to your sgconfig.yml file.

Code transformation: Execute sg run with pattern option for the code to find, rewrite option for the replacement, lang option for the language, and source directory path.

Test rules: Execute sg test to validate your rule definitions.

### Pattern Syntax Basics

The dollar sign followed by a variable name such as VAR matches any single AST node and acts as a meta-variable for capturing.

The dollar sign followed by three dollar signs and a variable name such as ARGS matches zero or more nodes using variadic capture.

The double dollar sign followed by underscore matches any single node as an anonymous capture when the value is not needed.

### Supported Languages

Python, JavaScript, TypeScript, Go, Rust, Java, Kotlin, C, C++, Ruby, Swift, C#, PHP, Scala, Elixir, Lua, HTML, Vue, Svelte, and 30+ more.

---

## Implementation Guide

### Installation

For macOS, use brew install ast-grep.

For cross-platform via npm, use npm install -g @ast-grep/cli.

For Rust via Cargo, use cargo install ast-grep.

### Basic Pattern Matching

#### Simple Pattern Search

To find all console.log calls, run sg with pattern console.log($MSG) and lang javascript.

To find all Python function definitions, run sg with pattern def $FUNC($$$ARGS): $$$BODY and lang python.

To find React useState hooks, run sg with pattern useState($INIT) and lang tsx.

#### Explore/Search Performance Optimization

AST-Grep provides significant performance benefits for codebase exploration compared to text-based search:

**Why AST-Grep is Faster for Exploration**
- Structural understanding eliminates false positives (50-80% reduction in irrelevant results)
- Syntax-aware matching reduces full file scans
- Single pass through AST vs multiple regex passes

**Common Exploration Patterns**

Find all function calls matching a pattern:
```bash
sg -p 'authenticate($$$)' --lang python -r src/
```

Find all classes inheriting from a base class:
```bash
sg -p 'class $A extends BaseService' --lang python -r src/
```

Find specific import patterns:
```bash
sg -p 'import fastapi' --lang python -r src/
```

Find React hooks usage:
```bash
sg -p 'useState($$)' --lang tsx -r src/
```

Find async function declarations:
```bash
sg -p 'async def $NAME($$$ARGS):' --lang python -r src/
```

**Performance Comparison**
- `grep -r "class.*Service" src/` - scans all files textually (~10s for large codebase)
- `sg -p 'class $X extends Service' --lang python -r src/` - structural match (~2s)

**Integration with Explore Agent**
When using the Explore agent, AST-Grep is automatically prioritized for:
- Class hierarchy analysis
- Function signature matching
- Import dependency mapping
- API usage pattern detection

#### Meta-variables

Meta-variables capture matching AST nodes in patterns.

Single node capture uses $NAME syntax. For example, pattern const $NAME = require($PATH) captures the variable name and path.

Variadic capture uses $$$ARGS syntax. For example, pattern function $NAME($$$ARGS) captures function name and all arguments.

Anonymous single capture uses $$_ syntax when you need to match but not reference the value.

### Code Transformation

#### Simple Rewrite

To rename a function, run sg with pattern oldFunc($ARGS), rewrite newFunc($ARGS), and lang python.

To update an API call, run sg with pattern axios.get($URL), rewrite fetch($URL), and lang typescript.

#### Complex Transformation with YAML Rules

Create a YAML rule file with the following structure. Set the id field to a unique rule identifier such as convert-var-to-const. Set language to the target language such as javascript. Under the rule section, specify the pattern to match such as var $NAME = $VALUE. Set the fix field to the replacement pattern such as const $NAME = $VALUE. Add a message describing the issue and set severity to warning or error.

Run sg scan with the rule option pointing to your rule file and the source directory.

### Rule-Based Scanning

#### Configuration File

Create an sgconfig.yml file with the following sections. The ruleDirs section lists directories containing rule files such as ./rules/security and ./rules/quality. The testConfigs section specifies test file patterns. The languageGlobs section maps languages to file patterns, mapping python to .py files, typescript to .ts and .tsx files, and javascript to .js and .jsx files.

#### Security Rule Example

Create a security rule file for SQL injection detection. Set the id to sql-injection-risk. Set language to python and severity to error. Write a descriptive message about the vulnerability. Under the rule section, use the any operator to match multiple patterns including cursor.execute with percent formatting, cursor.execute with format method, and cursor.execute with f-string interpolation. Set the fix to show the parameterized query alternative.

### Relational Rules

#### Inside Rule for Scoped Search

Create a rule that searches for console.log calls only inside function declarations. Set the pattern to console.log($$$ARGS) and add an inside constraint with pattern function $NAME($$$PARAMS).

#### Has Rule for Contains Check

Create a rule to find async functions without await. Set the pattern to async function $NAME($$$PARAMS) with a not constraint containing a has rule with pattern await $EXPR. Add message indicating async function without await.

#### Follows and Precedes Rules

Create a rule to detect missing error handling. Set the pattern to match error assignment $ERR := $CALL and add a not constraint with follows rule checking for if $ERR != nil error handling block.

### Composite Rules

Create complex rules using the all operator to combine multiple conditions. For example, combine pattern useState($INIT) with inside constraint for function component and not precedes constraint for useEffect call.

---

## Advanced Patterns

For comprehensive documentation including complex multi-file transformations, custom language configuration, CI/CD integration patterns, and performance optimization tips, see the following module files.

Pattern syntax reference is available in modules/pattern-syntax.md.

Security scanning rule templates are documented in modules/security-rules.md.

Common refactoring patterns are covered in modules/refactoring-patterns.md.

Language-specific patterns are detailed in modules/language-specific.md.

### Context7 Integration

For latest AST-Grep documentation, follow this two-step process.

Step 1: Use mcp__context7__resolve-library-id with query ast-grep to resolve the library identifier.

Step 2: Use mcp__context7__get-library-docs with the resolved library ID to fetch current documentation.

### MoAI-ADK Integration

AST-Grep is integrated into MoAI-ADK through the Tool Registry as AST_ANALYZER type in internal/hook/registry.go, PostToolUse Hook for automatic security scanning after Write/Edit operations, and Permissions with Bash(sg:*) and Bash(ast-grep:*) auto-allowed.

### Running Scans

To scan with MoAI-ADK rules, execute sg scan with config pointing to .claude/skills/moai-tool-ast-grep/rules/sgconfig.yml.

To scan a specific directory, execute sg scan with config sgconfig.yml and the src/ directory.

For JSON output suitable for CI/CD, execute sg scan with config and json flag, redirecting to results.json.

---

## Works Well With

- moai-workflow-testing: DDD integration and test pattern detection
- moai-foundation-quality: TRUST 5 compliance and code quality gates
- moai-domain-backend: API pattern detection and security scanning
- moai-domain-frontend: React/Vue pattern optimization
- moai-lang-python: Python-specific security and style rules
- moai-lang-typescript: TypeScript type safety patterns

### Related Agents

- expert-refactoring: AST-based large-scale refactoring
- expert-security: Security vulnerability scanning
- manager-quality: Code complexity analysis
- expert-debug: Pattern-based debugging

---

## Reference

For additional information, consult the AST-Grep Official Documentation at ast-grep.github.io, the AST-Grep GitHub Repository at github.com/ast-grep/ast-grep, the Pattern Playground at ast-grep.github.io/playground.html, and the Rule Configuration Reference at ast-grep.github.io/reference/yaml.html.
