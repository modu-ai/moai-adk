# AST-Grep Examples

Practical code examples for AST-Grep (sg) across multiple languages and use cases.

## Installation and Quick Start

### Install AST-Grep

```bash
# macOS
brew install ast-grep

# npm (cross-platform)
npm install -g @ast-grep/cli

# Cargo (Rust)
cargo install ast-grep

# Verify installation
sg --version
```

### Basic Usage

```bash
# Search for pattern
sg run --pattern 'console.log($MSG)' --lang javascript src/

# Transform code
sg run --pattern 'oldFunc($A)' --rewrite 'newFunc($A)' --lang python src/

# Scan with rules
sg scan --config sgconfig.yml
```

---

## 1. Basic Pattern Search

### JavaScript/TypeScript

```bash
# Find all console.log calls
sg run --pattern 'console.log($$$ARGS)' --lang javascript

# Find all variable declarations
sg run --pattern 'const $NAME = $VALUE' --lang typescript

# Find function definitions
sg run --pattern 'function $NAME($$$PARAMS) { $$$BODY }' --lang javascript

# Find arrow functions
sg run --pattern 'const $NAME = ($$$PARAMS) => $BODY' --lang typescript

# Find React useEffect hooks
sg run --pattern 'useEffect($$$ARGS)' --lang typescriptreact
```

### Python

```bash
# Find function definitions
sg run --pattern 'def $NAME($$$ARGS): $$$BODY' --lang python

# Find class definitions
sg run --pattern 'class $NAME($$$BASES): $$$BODY' --lang python

# Find print statements
sg run --pattern 'print($$$ARGS)' --lang python

# Find decorators
sg run --pattern '@$DECORATOR' --lang python
```

### Go

```bash
# Find function definitions
sg run --pattern 'func $NAME($$$PARAMS) $$$RET { $$$BODY }' --lang go

# Find struct definitions
sg run --pattern 'type $NAME struct { $$$FIELDS }' --lang go

# Find goroutines
sg run --pattern 'go $FUNC($$$ARGS)' --lang go
```

### Rust

```bash
# Find function definitions
sg run --pattern 'fn $NAME($$$PARAMS) -> $RET { $$$BODY }' --lang rust

# Find macro invocations
sg run --pattern '$MACRO!($$$ARGS)' --lang rust

# Find struct definitions
sg run --pattern 'struct $NAME { $$$FIELDS }' --lang rust
```

---

## 2. Relational Patterns (inside, has, follows, not)

### Inside Rule - Scoped Search

```yaml
# rules/no-console-in-production.js
id: no-console-in-production
language: javascript
severity: warning
message: 'Remove console.log from production code'
rule:
  pattern: 'console.log($$$ARGS)'
  inside:
    pattern: 'function $NAME($$$PARAMS) { $$$BODY }'
```

```bash
# Usage
sg scan --rule rules/no-console-in-production.js src/
```

### Has Rule - Contains Check

```yaml
# rules/async-without-await.yml
id: async-without-await
language: javascript
severity: warning
message: 'Async function declared but never uses await'
rule:
  pattern: 'async function $NAME($$$PARAMS) { $$$BODY }'
  not:
    has:
      pattern: 'await $EXPR'
```

### Follows Rule - Sequential Check

```yaml
# rules/missing-error-check.yml
id: missing-error-check
language: go
severity: error
message: 'Function call may return error but error is not checked'
rule:
  pattern: '$_, $ERR := $CALL()'
  not:
    follows:
      pattern: 'if $ERR != nil { $$$BODY }'
```

### Precedes Rule - Preceding Check

```yaml
# rules/missing-initialization.yml
id: missing-initialization
language: python
severity: error
message: 'Variable used before initialization'
rule:
  pattern: '$VAR = $VALUE'
  not:
    precedes:
      pattern: 'print($VAR)'
```

### Not Rule - Negation

```yaml
# rules/no-direct-state-mutation.yml
id: no-direct-state-mutation
language: javascript
severity: warning
message: 'Use setState instead of direct mutation'
rule:
  pattern: 'this.state.$PROP = $VALUE'
  not:
    has:
      pattern: 'setState'
```

### Combined Relational Rules

```yaml
# rules/react-hooks-order.yml
id: react-hooks-order
language: typescriptreact
severity: error
message: 'React hooks must be called in the same order'
rule:
  all:
    - pattern: 'useState($$$ARGS)'
    - inside:
        pattern: 'function $COMPONENT($$$PROPS) { $$$BODY }'
    - not:
        follows:
          pattern: 'useEffect($$$ARGS)'
```

---

## 3. String Pattern Search

### Exact String Match

```bash
# Find exact string literals
sg run --pattern '"TODO"' --lang python

# Find template literals
sg run --pattern '`$STRING`' --lang javascript
```

### String Pattern with Wildcards

```yaml
# rules/hardcoded-secrets.yml
id: hardcoded-secrets
language: python
severity: error
message: 'Hardcoded secret detected. Use environment variables.'
rule:
  any:
    - pattern: 'api_key = "$SECRET"'
    - pattern: 'password = "$SECRET"'
    - pattern: 'token = "$SECRET"'
  not:
    has:
      pattern: 'os.getenv'
```

### String in Specific Context

```yaml
# rules/sql-string-concat.yml
id: sql-string-concat
language: python
severity: error
message: 'SQL query constructed via string concatenation'
rule:
  pattern: 'cursor.execute("$SQL" + $MORE)'
```

### Multi-line String Patterns

```bash
# Find multi-line strings
sg run --pattern '"""$CONTENT"""' --lang python

# Find template literals with expressions
sg run --pattern '`$PREFIX ${$EXPR} $SUFFIX`' --lang javascript
```

---

## 4. Code Transformation (Codemod) Patterns

### Simple Rename

```bash
# Rename function globally
sg run --pattern 'oldFunc($ARGS)' --rewrite 'newFunc($ARGS)' --lang python src/

# Rename variable
sg run --pattern 'var $NAME = $VALUE' --rewrite 'let $NAME = $VALUE' --lang javascript src/
```

### API Migration

```yaml
# rules/migrate-axios-to-fetch.yml
id: migrate-axios-to-fetch
language: typescript
rule:
  pattern: 'axios.get($URL)'
fix: |
  fetch($URL)
    .then(res => res.json())
```

### Convert Callback to Promise

```yaml
# rules/callback-to-promise.yml
id: callback-to-promise
language: javascript
rule:
  pattern: |
    fs.readFile($PATH, function($ERR, $DATA) {
      $$$BODY
    })
fix: |
  fs.promises.readFile($PATH)
    .then($DATA => {
      $$$BODY
    })
```

### Add Error Handling

```yaml
# rules/add-try-catch.yml
id: add-try-catch
language: python
rule:
  pattern: |
    $CALL()
  inside:
    pattern: 'def $FUNC($$$ARGS): $$$BODY'
fix: |
    try:
        $CALL()
    except Exception as e:
        logger.error(f"Error in $FUNC: {e}")
        raise
```

### Extract Common Pattern

```yaml
# rules/extract-logger.yml
id: extract-logger
language: python
rule:
  pattern: |
    print("DEBUG: $MSG")
fix: |
    logger.debug($MSG)
```

### Convert ES5 to ES6

```yaml
# rules/var-to-const.yml
id: var-to-const
language: javascript
rule:
  pattern: 'var $NAME = $VALUE'
fix: 'const $NAME = $VALUE'
```

### Add Type Annotations

```yaml
# rules/add-type-hints.yml
id: add-type-hints
language: python
rule:
  pattern: 'def $NAME($$$ARGS):'
fix: 'def $NAME($$$ARGS) -> None:'
```

---

## 5. Refactoring Patterns

### Function Extraction

```yaml
# rules/extract-validation.yml
id: extract-validation
language: python
severity: suggestion
message: 'Consider extracting validation logic'
rule:
  pattern: |
    if $COND1 and $COND2:
        $$$BODY
  inside:
    pattern: 'def $FUNC($$$ARGS):'
```

### Variable Renaming

```bash
# Rename poorly named variables
sg run --pattern '$OBJ.doSomething()' --rewrite '$OBJ.performAction()' --lang javascript

# Rename to snake_case (Python)
sg run --pattern '$myVariable' --rewrite '$my_variable' --lang python
```

### Extract Magic Numbers

```yaml
# rules/magic-number.yml
id: magic-number
language: python
severity: warning
message: 'Magic number detected. Use named constant.'
rule:
  pattern: 'timeout = $NUM'
  not:
    has:
      pattern: 'CONST_'
where:
  NUM:
    regex: '^[0-9]+$'
```

### Simplify Conditional

```yaml
# rules/simplify-boolean.yml
id: simplify-boolean
language: javascript
severity: suggestion
message: 'Simplify boolean expression'
rule:
  pattern: 'if ($COND == true) { $BODY }'
fix: 'if ($COND) { $BODY }'
```

### Remove Dead Code

```yaml
# rules/dead-import.yml
id: dead-import
language: python
severity: warning
message: 'Imported module never used'
rule:
  pattern: 'import $MODULE'
  not:
    has:
      pattern: '$MODULE.'
```

### Convert to Modern Syntax

```yaml
# rules/convert-to-f-string.yml
id: convert-to-f-string
language: python
severity: suggestion
message: 'Use f-string instead of format()'
rule:
  pattern: '"{} ".format($$$ARGS)'
fix: 'f"{$$$ARGS}"'
```

---

## 6. AST-Based Code Exploration

### Find All Call Sites of a Function

```bash
# Find all calls to specific function
sg run --pattern 'myFunction($$$ARGS)' --lang python

# Find all method calls on specific object
sg run --pattern '$OBJ.method($$$ARGS)' --lang javascript
```

### Find Function Return Patterns

```yaml
# rules/early-return.yml
id: early-return
language: python
severity: info
message: 'Early return pattern detected'
rule:
  pattern: |
    if $COND:
        return $VALUE
    $$$REST
  inside:
    pattern: 'def $FUNC($$$ARGS):'
```

### Find Nested Functions

```bash
# Find nested function definitions
sg run --pattern 'def outer($$$ARGS): { def inner($$$ARGS): $$$BODY }' --lang python
```

### Find Class Hierarchy

```yaml
# rules/find-override.yml
id: find-override
language: python
severity: info
message: 'Method override detected'
rule:
  pattern: |
    class $CLASS($BASE):
        def $METHOD($$$ARGS):
            $$$BODY
```

### Find Async Patterns

```bash
# Find all async/await patterns
sg run --pattern 'async function $NAME($$$ARGS) { $$$BODY }' --lang javascript

# Find Promise chains
sg run --pattern '$PROMISE.then($$$ARGS).catch($$$CATCH)' --lang typescript
```

---

## 7. Security Scanning (OWASP Top 10)

### SQL Injection

```yaml
# rules/security/sql-injection.yml
id: sql-injection
language: python
severity: error
message: 'Potential SQL injection vulnerability. Use parameterized queries.'
rule:
  any:
    - pattern: 'cursor.execute($QUERY % $ARGS)'
    - pattern: 'cursor.execute($QUERY.format($$$ARGS))'
    - pattern: 'cursor.execute(f"$$$SQL")'
    - pattern: 'cursor.execute("$SQL" + $MORE)'
fix: 'cursor.execute($QUERY, $ARGS)'
```

### XSS (Cross-Site Scripting)

```yaml
# rules/security/xss-risk.yml
id: xss-risk
language: javascript
severity: error
message: 'Potential XSS vulnerability. Sanitize user input.'
rule:
  pattern: 'innerHTML = $USER_INPUT'
  inside:
    pattern: 'function $FUNC($$$ARGS):'
```

### Hardcoded Credentials

```yaml
# rules/security/hardcoded-credentials.yml
id: hardcoded-credentials
language: python
severity: error
message: 'Hardcoded credentials detected. Use environment variables.'
rule:
  any:
    - pattern: 'password = "$Creds"'
    - pattern: 'api_key = "$Creds"'
    - pattern: 'secret = "$Creds"'
    - pattern: 'token = "$Creds"'
```

### Command Injection

```yaml
# rules/security/command-injection.yml
id: command-injection
language: python
severity: error
message: 'Potential command injection. Use subprocess.run with list args.'
rule:
  pattern: 'os.system($CMD)'
  not:
    has:
      pattern: 'shlex.quote'
```

### Insecure Deserialization

```yaml
# rules/security/insecure-deserialize.yml
id: insecure-deserialize
language: python
severity: error
message: 'Insecure deserialization detected. Use safe alternatives.'
rule:
  pattern: 'pickle.loads($DATA)'
```

### Weak Cryptography

```yaml
# rules/security/weak-crypto.yml
id: weak-crypto
language: python
severity: error
message: 'Weak cryptography algorithm detected. Use AES-256.'
rule:
  any:
    - pattern: 'Cipher.algorithms_ecb($$$ARGS)'
    - pattern: 'DES.new($$$ARGS)'
    - pattern: 'MD5.new($$$ARGS)'
```

### Path Traversal

```yaml
# rules/security/path-traversal.yml
id: path-traversal
language: python
severity: error
message: 'Path traversal vulnerability. Validate user input.'
rule:
  pattern: 'open($USER_INPUT)'
  not:
    has:
      pattern: 'os.path.abspath'
```

### Sensitive Data Exposure

```yaml
# rules/security/log-sensitive-data.yml
id: log-sensitive-data
language: javascript
severity: warning
message: 'Logging sensitive data detected. Remove before production.'
rule:
  pattern: 'console.log($$$SENSITIVE)'
  has:
    pattern: 'password'
```

### Missing Authentication

```yaml
# rules/security/missing-auth.yml
id: missing-auth
language: javascript
severity: error
message: 'Public endpoint without authentication check'
rule:
  pattern: |
    app.get($ROUTE, function($REQ, $RES) {
      $$$BODY
    })
  not:
    has:
      pattern: 'authenticate'
```

### CSRF Protection

```yaml
# rules/security/csrf-protection.yml
id: csrf-protection
language: javascript
severity: error
message: 'State-changing operation without CSRF token'
rule:
  pattern: |
    app.post($ROUTE, function($REQ, $RES) {
      $$$BODY
    })
  not:
    has:
      pattern: 'csrfToken'
```

---

## 8. Language-Specific Examples

### Python

```yaml
# Find context managers
pattern: 'with $CTX as $VAR: $$$BODY'

# Find list comprehensions
pattern: '[$EXPR for $VAR in $ITER]'

# Find type hints
pattern: 'def $FUNC($ARGS) -> $RET: $$$BODY'

# Find dataclass definitions
pattern: '@dataclass class $NAME: $$$BODY'
```

### JavaScript/TypeScript

```yaml
# Find React components
pattern: 'function $NAME($$$PROPS) { return $$$JSX }'

# Find class components
pattern: 'class $NAME extends React.Component { render() { return $$$JSX } }'

# Find JSX elements
pattern: '<$TAG $$$PROPS>$CHILDREN</$TAG>'

# Find TypeScript interfaces
pattern: 'interface $NAME { $$$MEMBERS }'
```

### Go

```yaml
# Find interface implementations
pattern: 'func ($SELF *$TYPE) $METHOD($$$ARGS) $$$RET { $$$BODY }'

# Find channel operations
pattern: '$CHAN <- $VALUE'

# Find select statements
pattern: 'select { $$$CASES }'

# Find goroutines
pattern: 'go $FUNC($$$ARGS)'
```

### Rust

```yaml
# Find trait implementations
pattern: 'impl $TRAIT for $TYPE { $$$METHODS }'

# Find match expressions
pattern: 'match $EXPR { $$$ARMS }'

# Find closures
pattern: '|$CAPTURE| $BODY'

# Find async functions
pattern: 'async fn $NAME($$$ARGS) -> $RET { $$$BODY }'
```

### Java

```yaml
# Find class definitions
pattern: 'class $NAME $$$EXTENDS $$$IMPLEMENTS { $$$BODY }'

# Find method definitions
pattern: 'public $RET $NAME($$$ARGS) { $$$BODY }'

# Find annotations
pattern: '@$ANNOTATION'

# Find lambda expressions
pattern: '($$$PARAMS) -> $EXPR'
```

### C++

```yaml
# Find class definitions
pattern: 'class $NAME { $$$MEMBERS }'

# Find template functions
pattern: 'template<$$$PARAMS> $RET $FUNC($$$ARGS)'

# Find lambda expressions
pattern: '[$$$CAPTURE]($$$ARGS) { $$$BODY }'

# Find smart pointers
pattern: 'std::unique_ptr<$TYPE>'
```

---

## 9. CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/ast-grep-scan.yml
name: AST-Grep Security Scan

on: [push, pull_request]

jobs:
  ast-grep:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install AST-Grep
        run: npm install -g @ast-grep/cli

      - name: Run Security Scan
        run: sg scan --config .claude/skills/moai-tool-ast-grep/rules/sgconfig.yml --format github

      - name: Upload Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: ast-grep-results
          path: sg-report.json
```

### Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
echo "Running AST-Grep security scan..."

sg scan --config .claude/skills/moai-tool-ast-grep/rules/sgconfig.yml --format compact

if [ $? -ne 0 ]; then
    echo "AST-Grep found issues. Please fix before committing."
    exit 1
fi

echo "AST-Grep scan passed."
```

### GitLab CI

```yaml
# .gitlab-ci.yml
ast-grep-scan:
  stage: test
  image: node:latest
  script:
    - npm install -g @ast-grep/cli
    - sg scan --config sgconfig.yml --json > sg-report.json
  artifacts:
    reports:
      sast: sg-report.json
  only:
    - merge_requests
    - main
```

### Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any
    stages {
        stage('AST-Grep Scan') {
            steps {
                sh 'npm install -g @ast-grep/cli'
                sh 'sg scan --config sgconfig.yml --format json > sg-report.json'
            }
        }
    }
    post {
        always {
            archiveArtifacts artifacts: 'sg-report.json'
        }
    }
}
```

---

## 10. CLI Usage Examples

### Search Operations

```bash
# Basic search
sg run --pattern 'console.log($MSG)' --lang javascript

# Recursive search in directory
sg run --pattern 'def $FUNC($$$ARGS):' --lang python src/

# Search with JSON output
sg run --pattern 'useState($$$ARGS)' --lang typescriptreact --json

# Search with file filter
sg run --pattern 'TODO:' --lang python --glob '**/*.py'

# Interactive search
sg run --pattern 'fetch($$$ARGS)' --lang javascript --interactive
```

### Transform Operations

```bash
# Simple rename
sg run --pattern 'foo($A)' --rewrite 'bar($A)' --lang python

# Complex transformation with YAML
sg run --rule convert-to-arrow-function.yml src/

# Dry run (preview changes)
sg run --pattern 'var $X = $Y' --rewrite 'let $X = $Y' --lang javascript --dry-run

# Backup before transformation
sg run --pattern 'oldAPI($$$ARGS)' --rewrite 'newAPI($$$ARGS)' --lang javascript --backup
```

### Scan Operations

```bash
# Scan with configuration
sg scan --config sgconfig.yml

# Scan specific directory
sg scan --config sgconfig.yml src/

# Scan with severity filter
sg scan --config sgconfig.yml --severity error

# JSON output
sg scan --config sgconfig.yml --json > results.json

# SARIF format for CI/CD
sg scan --config sgconfig.yml --format sarif -o results.sarif
```

### Test Operations

```bash
# Test all rules
sg test

# Test specific rule
sg test rules/security/sql-injection.yml

# Verbose test output
sg test --verbose

# Watch mode (auto-run on file change)
sg test --watch
```

### Utility Commands

```bash
# Check version
sg --version

# Help for specific command
sg run --help

# Generate configuration template
sg init

# Validate configuration
sg validate sgconfig.yml

# List supported languages
sg lang list
```

---

## Common Workflows

### Find and Replace Function Name

```bash
# Step 1: Find all usages
sg run --pattern 'oldFunc($$$ARGS)' --lang javascript --json > usages.json

# Step 2: Preview changes
sg run --pattern 'oldFunc($$$ARGS)' --rewrite 'newFunc($$$ARGS)' --lang javascript --dry-run

# Step 3: Apply changes
sg run --pattern 'oldFunc($$$ARGS)' --rewrite 'newFunc($$$ARGS)' --lang javascript
```

### Security Audit

```bash
# Scan for security issues
sg scan --config .claude/skills/moai-tool-ast-grep/rules/sgconfig.yml --severity error

# Generate SARIF report for GitHub Security
sg scan --config sgconfig.yml --format sarif -o security-report.sarif

# View results in GitHub
# Upload security-report.sarif to GitHub Security tab
```

### Code Quality Check

```bash
# Run all quality rules
sg scan --config sgconfig.yml --severity warning

# Generate HTML report
sg scan --config sgconfig.yml --format html -o quality-report.html

# Filter by rule ID
sg scan --config sgconfig.yml --rule sql-injection
```

### Refactoring Session

```bash
# Step 1: Identify patterns
sg run --pattern 'var $NAME = $VALUE' --lang javascript

# Step 2: Create transformation rule
# Edit refactoring-rule.yml

# Step 3: Test rule
sg test refactoring-rule.yml

# Step 4: Apply transformation
sg run --rule refactoring-rule.yml src/

# Step 5: Verify with tests
pytest tests/
```

---

## Performance Tips

### Optimize Search Speed

```bash
# Use language filter
sg run --pattern 'function($$$ARGS)' --lang javascript

# Use glob pattern for file filtering
sg run --pattern 'import $MODULE' --lang python --glob '**/models/*.py'

# Exclude directories
sg run --pattern 'const $X = $Y' --lang javascript --exclude-dir node_modules

# Parallel execution (default)
sg run --pattern 'TODO:' --lang python -j 4
```

### Memory Management

```bash
# Limit file size
sg scan --config sgconfig.yml --max-file-size 1MB

# Process files in batches
sg run --pattern 'TODO:' --lang python --batch-size 100

# Use incremental scanning
sg scan --config sgconfig.yml --incremental
```

---

## Troubleshooting

### Pattern Not Matching

```bash
# Check AST structure
sg parse --lang python file.py

# Test pattern interactively
sg run --pattern 'def $FUNC($$$ARGS):' --lang python --interactive

# Verify language support
sg lang list | grep python
```

### Transformation Issues

```bash
# Dry run first
sg run --pattern 'foo($A)' --rewrite 'bar($A)' --lang python --dry-run

# Check rewrite syntax
sg test --rule rule.yml

# Backup before apply
sg run --pattern 'foo($A)' --rewrite 'bar($A)' --lang python --backup
```

### Configuration Problems

```bash
# Validate configuration
sg validate sgconfig.yml

# Check rule syntax
sg test --rule rule.yml --verbose

# Debug mode
sg scan --config sgconfig.yml --debug
```

---

## Reference

- [AST-Grep Official Documentation](https://ast-grep.github.io/)
- [Pattern Syntax Reference](https://ast-grep.github.io/reference/pattern.html)
- [Rule Configuration](https://ast-grep.github.io/reference/yaml.html)
- [Pattern Playground](https://ast-grep.github.io/playground.html)
- [GitHub Repository](https://github.com/ast-grep/ast-grep)

---

**Version**: 1.0.0
**Last Updated**: 2026-01-06
**Skill**: moai-tool-ast-grep
