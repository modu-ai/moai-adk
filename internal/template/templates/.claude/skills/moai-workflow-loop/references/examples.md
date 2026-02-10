# Ralph Engine Practical Examples

## Basic Loop Usage

### Starting a Feedback Loop

```bash
# Start Ralph loop for automated error fixing
/moai:loop

# Start with custom max iterations
/moai:loop --max-iterations 5

# Start with specific completion criteria
/moai:loop --zero-warnings true
```

**Loop State Management:**

```json
// .moai/cache/.moai_loop_state.json
{
  "active": true,
  "iteration": 1,
  "max_iterations": 10,
  "last_error_count": 5,
  "start_time": "2026-01-10T10:30:00Z",
  "completion_reason": null
}
```

**Expected Behavior:**

1. Ralph activates and begins monitoring code changes
2. PostToolUse hook runs after each Write/Edit operation
3. LSP diagnostics are collected and reported to Claude
4. Stop hook checks completion conditions after each response
5. Loop continues until zero errors or max iterations reached

**Example Conversation Flow:**

```
User: /moai:loop
Claude: Ralph loop activated. Starting automated error correction...

[Claude makes code changes]

Hook Output: LSP: 3 error(s), 2 warning(s) in auth.py
  - [ERROR] Line 45: undefined name 'user_id'
  - [ERROR] Line 67: too many arguments for function
  - [ERROR] Line 89: incompatible return type

Claude: Fixing 3 errors found by LSP diagnostics...

[Claude fixes errors]

Hook Output: LSP: 1 error(s), 0 warning(s) in auth.py
  - [ERROR] Line 45: type mismatch

Claude: Fixing remaining error...

[Claude fixes final error]

Hook Output: LSP: 0 error(s), 0 warning(s) in auth.py
Stop Hook: Ralph Loop: COMPLETE | All errors resolved!

Claude: ✓ Ralph loop completed successfully. All diagnostics clear.
```

### Monitoring Progress

**Stop Hook Progress Report:**

```
Ralph Loop: CONTINUE | Iteration: 3/10 | Errors: 2
Next actions: Fix 2 error(s) in src/api/users.py
```

**Completion Messages:**

```
Ralph Loop: COMPLETE | Zero errors achieved
Ralph Loop: COMPLETE | Max iterations (10) reached
Ralph Loop: COMPLETE | Manual cancellation
```

### Handling Completion

**Successful Completion:**

```json
{
  "active": false,
  "iteration": 5,
  "max_iterations": 10,
  "last_error_count": 0,
  "completion_reason": "zero_errors",
  "end_time": "2026-01-10T10:45:00Z"
}
```

**Max Iterations Reached:**

```json
{
  "active": false,
  "iteration": 10,
  "max_iterations": 10,
  "last_error_count": 2,
  "completion_reason": "max_iterations",
  "remaining_issues": [
    "src/auth.py:45: type mismatch",
    "src/auth.py:67: undefined variable"
  ]
}
```

---

## Fix Command Examples

### One-Time Error Fixes

**Basic Usage:**

```bash
# Fix all current LSP errors
/moai:fix

# Fix errors in specific file
/moai:fix src/auth.py

# Fix with specific severity threshold
/moai:fix --severity error
```

**Example Session:**

```
User: /moai:fix

Claude: Running LSP diagnostics...

LSP Results:
  src/auth.py:
    - [ERROR] Line 45: undefined name 'user_id'
    - [ERROR] Line 67: missing return statement
  src/api.py:
    - [WARNING] Line 23: unused import 'os'

Fixing 2 errors...

[Claude makes fixes]

Claude: ✓ Fixed 2 errors
  - Defined user_id in auth.py line 45
  - Added return statement in auth.py line 67

Remaining: 1 warning (use --severity warning to include)
```

### Fixing Different Error Types

**Python Type Errors:**

```python
# Before
def get_user(id: int) -> User:
    user = db.query(User).filter(User.id == id).first()
    return user  # Error: might return None

# After (Ralph fixes)
def get_user(id: int) -> User | None:
    user = db.query(User).filter(User.id == id).first()
    return user
```

**TypeScript Type Errors:**

```typescript
// Before
function processUser(user: User): string {
  return user.name; // Error: Property 'name' does not exist
}

// After (Ralph fixes)
interface User {
  name: string;
  email: string;
}

function processUser(user: User): string {
  return user.name;
}
```

**Go Type Errors:**

```go
// Before
func GetUser(id int) User {
    user := db.Find(id)
    return user  // Error: cannot use *User as User
}

// After (Ralph fixes)
func GetUser(id int) *User {
    user := db.Find(id)
    return user
}
```

---

## Configuration Examples

### Basic Configuration (ralph.yaml)

**Minimal Setup:**

```yaml
ralph:
  enabled: true

  lsp:
    auto_start: true

  loop:
    max_iterations: 10
```

**Production Configuration:**

```yaml
ralph:
  enabled: true

  lsp:
    auto_start: true
    timeout_seconds: 30
    graceful_degradation: true
    servers:
      python: "pyright"
      typescript: "tsserver"
      go: "gopls"

  ast_grep:
    enabled: true
    security_scan: true
    quality_scan: true
    custom_rules:
      - .moai/ast-grep/security/**/*.yml
      - .moai/ast-grep/quality/**/*.yml

  loop:
    max_iterations: 15
    auto_fix: false
    require_confirmation: true
    completion:
      zero_errors: true
      zero_warnings: false
      tests_pass: true
      coverage_threshold: 85

  hooks:
    post_tool_lsp:
      enabled: true
      severity_threshold: "error"
      ignore_patterns:
        - "*.test.py"
        - "tests/**/*"
    stop_loop_controller:
      enabled: true
      verbose: true
```

### Development vs Production Configurations

**Development (relaxed):**

```yaml
ralph:
  enabled: true

  loop:
    max_iterations: 20
    auto_fix: true
    require_confirmation: false
    completion:
      zero_errors: true
      zero_warnings: false
      tests_pass: false
```

**Production (strict):**

```yaml
ralph:
  enabled: true

  loop:
    max_iterations: 5
    auto_fix: false
    require_confirmation: true
    completion:
      zero_errors: true
      zero_warnings: true
      tests_pass: true
      coverage_threshold: 90
      security_scan: true
```

### Language-Specific Configurations

**Python Project:**

```yaml
ralph:
  enabled: true

  lsp:
    auto_start: true
    servers:
      python: "pyright"
    settings:
      pyright:
        typeCheckingMode: "strict"
        useLibraryCodeForTypes: true

  ast_grep:
    enabled: true
    rules:
      - .moai/ast-grep/python/**/*.yml

  loop:
    completion:
      zero_errors: true
      coverage_threshold: 90
```

**TypeScript/React Project:**

```yaml
ralph:
  enabled: true

  lsp:
    auto_start: true
    servers:
      typescript: "tsserver"
      javascript: "tsserver"
    settings:
      typescript:
        strict: true
        noImplicitAny: true

  ast_grep:
    enabled: true
    rules:
      - .moai/ast-grep/react/**/*.yml
      - .moai/ast-grep/typescript/**/*.yml
```

**Multi-Language Project:**

```yaml
ralph:
  enabled: true

  lsp:
    auto_start: true
    servers:
      python: "pyright"
      typescript: "tsserver"
      go: "gopls"
      rust: "rust-analyzer"

  ast_grep:
    enabled: true
    rules:
      - .moai/ast-grep/**/*.yml
```

---

## Hook Integration Examples

### PostToolUse Hook Communication

**Hook Input Schema:**

```json
{
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/Users/project/src/auth.py",
    "content": "def authenticate(user: str, password: str):\n    return True"
  },
  "tool_output": "File written successfully"
}
```

**Hook Output Schema:**

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "LSP: 2 error(s), 1 warning(s) in auth.py\n  - [ERROR] Line 2: missing return type annotation\n  - [ERROR] Line 3: security issue - hardcoded credentials\n  - [WARNING] Line 1: function too complex"
  }
}
```

**Exit Codes:**

```python
# post_tool__lsp_diagnostic (Go compiled hook)

# Exit 0: No diagnostics or all clear
sys.exit(0)

# Exit 2: Errors found, attention needed
sys.exit(2)

# Exit 1: Hook execution error (rare)
sys.exit(1)
```

### Stop Hook Loop Control

**State File Management:**

```python
# stop__loop_controller (Go compiled hook)

def load_loop_state() -> dict:
    state_file = Path(".moai/cache/.moai_loop_state.json")
    if state_file.exists():
        return json.loads(state_file.read_text())
    return {"active": False}

def save_loop_state(state: dict):
    state_file = Path(".moai/cache/.moai_loop_state.json")
    state_file.parent.mkdir(parents=True, exist_ok=True)
    state_file.write_text(json.dumps(state, indent=2))

def update_iteration(state: dict) -> dict:
    state["iteration"] += 1
    state["last_updated"] = datetime.now(timezone.utc).isoformat()
    return state
```

**Completion Check Logic:**

```python
def check_completion(state: dict, config: dict) -> tuple[bool, str | None]:
    # Check max iterations
    if state["iteration"] >= config["loop"]["max_iterations"]:
        return True, "max_iterations"

    # Check zero errors
    if config["loop"]["completion"]["zero_errors"]:
        if state["last_error_count"] > 0:
            return False, None

    # Check zero warnings
    if config["loop"]["completion"]["zero_warnings"]:
        if state["last_warning_count"] > 0:
            return False, None

    # Check tests pass
    if config["loop"]["completion"]["tests_pass"]:
        if not state["tests_passing"]:
            return False, None

    # All conditions met
    return True, "zero_errors"
```

**Hook Output Generation:**

```python
def generate_hook_output(state: dict, complete: bool, reason: str | None) -> dict:
    if complete:
        message = f"Ralph Loop: COMPLETE | {reason}"
    else:
        iteration_info = f"{state['iteration']}/{state['max_iterations']}"
        error_info = f"Errors: {state['last_error_count']}"
        message = f"Ralph Loop: CONTINUE | Iteration: {iteration_info} | {error_info}"

    return {
        "hookSpecificOutput": {
            "hookEventName": "Stop",
            "additionalContext": message
        }
    }
```

---

## LSP Diagnostics Examples

### Python Diagnostics (Pyright)

**Type Errors:**

```python
# Code with errors
def process_user(user_id: int) -> str:
    user = get_user(user_id)  # Returns User | None
    return user.name  # Error: 'None' has no attribute 'name'

# LSP Diagnostic
{
    "range": {
        "start": {"line": 2, "character": 11},
        "end": {"line": 2, "character": 20}
    },
    "severity": 1,  # Error
    "code": "reportOptionalMemberAccess",
    "source": "pyright",
    "message": "'name' is not a known member of 'None'"
}

# Ralph's fix
def process_user(user_id: int) -> str:
    user = get_user(user_id)
    if user is None:
        raise ValueError(f"User {user_id} not found")
    return user.name
```

**Import Errors:**

```python
# Code with error
from typing import Dict  # Error: Use dict instead (PEP 585)

# LSP Diagnostic
{
    "severity": 2,  # Warning
    "message": "\"Dict\" is deprecated, use \"dict\" instead",
    "source": "pyright"
}

# Ralph's fix
from typing import TYPE_CHECKING
```

### TypeScript Diagnostics (tsserver)

**Type Mismatches:**

```typescript
// Code with error
interface User {
    id: number;
    name: string;
}

function getUser(id: string): User {  // Error: id should be number
    return { id, name: "Test" };  // Error: Type 'string' is not assignable to type 'number'
}

// LSP Diagnostic
{
    "range": {
        "start": {"line": 5, "character": 13},
        "end": {"line": 5, "character": 15}
    },
    "severity": 1,
    "code": 2322,
    "source": "ts",
    "message": "Type 'string' is not assignable to type 'number'"
}

// Ralph's fix
function getUser(id: number): User {
    return { id, name: "Test" };
}
```

### Go Diagnostics (gopls)

**Unused Variables:**

```go
// Code with error
func ProcessUser(id int) error {
    user, err := GetUser(id)  // Error: user declared but not used
    if err != nil {
        return err
    }
    return nil
}

// LSP Diagnostic
{
    "severity": 1,
    "message": "user declared and not used",
    "source": "compiler"
}

// Ralph's fix
func ProcessUser(id int) error {
    _, err := GetUser(id)
    if err != nil {
        return err
    }
    return nil
}
```

---

## AST-grep Security Scan Examples

### SQL Injection Detection

**AST-grep Rule:**

```yaml
# .moai/ast-grep/security/sql-injection.yml
id: sql-injection-python
language: python
rule:
  pattern: cursor.execute($SQL)
  where:
    SQL:
      kind: binary_expression
      has:
        kind: string
message: "Potential SQL injection vulnerability"
severity: error
```

**Code Detected:**

```python
# Vulnerable code
def get_user(user_id):
    cursor.execute("SELECT * FROM users WHERE id = " + user_id)
```

**Ralph's Fix:**

```python
# Fixed code
def get_user(user_id):
    cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))
```

### XSS Prevention

**AST-grep Rule:**

```yaml
# .moai/ast-grep/security/xss-prevention.yml
id: xss-react
language: typescript
rule:
  pattern: dangerouslySetInnerHTML={{ __html: $HTML }}
  where:
    HTML:
      not:
        has:
          kind: call_expression
          has:
            field: function
            regex: "sanitize|escape"
message: "Unsanitized HTML may cause XSS"
severity: error
```

**Code Detected:**

```typescript
// Vulnerable code
function UserProfile({ bio }: { bio: string }) {
    return <div dangerouslySetInnerHTML={{ __html: bio }} />;
}
```

**Ralph's Fix:**

```typescript
// Fixed code
import DOMPurify from 'dompurify';

function UserProfile({ bio }: { bio: string }) {
    const sanitizedBio = DOMPurify.sanitize(bio);
    return <div dangerouslySetInnerHTML={{ __html: sanitizedBio }} />;
}
```

### Hardcoded Secrets

**AST-grep Rule:**

```yaml
# .moai/ast-grep/security/hardcoded-secrets.yml
id: hardcoded-api-key
language: python
rule:
  any:
    - pattern: API_KEY = "$KEY"
    - pattern: SECRET_KEY = "$KEY"
    - pattern: PASSWORD = "$PWD"
  where:
    KEY:
      regex: "^[A-Za-z0-9]{20,}$"
message: "Hardcoded secret detected"
severity: error
```

**Code Detected:**

```python
# Vulnerable code
API_KEY = "sk_live_1234567890abcdefghij"
```

**Ralph's Fix:**

```python
# Fixed code
import os
API_KEY = os.environ.get("API_KEY")
if not API_KEY:
    raise ValueError("API_KEY environment variable not set")
```

---

## CI/CD Integration

### GitHub Actions Workflow

**Basic Integration:**

```yaml
# .github/workflows/ralph-quality.yml
name: Ralph Quality Check

on:
  pull_request:
    branches: [main]

jobs:
  ralph-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.13"

      - name: Install MoAI-ADK
        run: |
          go install github.com/modu-ai/moai-adk/cmd/moai@latest
          moai init

      - name: Run Ralph Loop
        run: |
          claude -p "/moai:loop --max-iterations 5" \
            --allowedTools "Read,Write,Edit,Bash" \
            --output-format json
        env:
          MOAI_LOOP_ACTIVE: "true"
          CLAUDE_API_KEY: ${{ secrets.CLAUDE_API_KEY }}

      - name: Check Results
        run: |
          python .moai/scripts/check_ralph_results.py
```

**Advanced CI/CD Integration:**

```yaml
# .github/workflows/ralph-advanced.yml
name: Ralph Advanced Quality

on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  ralph-full-check:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        check: [lsp, ast-grep, tests, coverage]

    steps:
      - uses: actions/checkout@v4

      - name: Install Dependencies
        run: |
          go install github.com/modu-ai/moai-adk/cmd/moai@latest
          moai init

      - name: Run ${{ matrix.check }} Check
        run: |
          case "${{ matrix.check }}" in
            lsp)
              claude -p "/moai:fix --severity error" \
                --allowedTools "Read,Write,Edit"
              ;;
            ast-grep)
              moai ast-grep scan --security
              ;;
            tests)
              pytest tests/ -v
              ;;
            coverage)
              pytest tests/ --cov --cov-report=json
              ;;
          esac
        env:
          CLAUDE_API_KEY: ${{ secrets.CLAUDE_API_KEY }}

      - name: Upload Results
        uses: actions/upload-artifact@v4
        with:
          name: ${{ matrix.check }}-results
          path: .moai/reports/${{ matrix.check }}
```

### Pre-commit Hook Integration

**Setup Script:**

```bash
# .git/hooks/pre-commit
#!/bin/bash

echo "Running Ralph pre-commit checks..."

# Run LSP diagnostics
moai lsp diagnose --changed-files

if [ $? -ne 0 ]; then
    echo "❌ LSP errors found. Run '/moai:fix' to resolve."
    exit 1
fi

# Run AST-grep security scan
moai ast-grep scan --security --changed-files

if [ $? -ne 0 ]; then
    echo "❌ Security issues found. Review and fix."
    exit 1
fi

echo "✓ All Ralph checks passed"
exit 0
```

### Docker Integration

**Dockerfile with Ralph:**

```dockerfile
FROM python:3.13-slim

WORKDIR /app

# Install MoAI-ADK
RUN go install github.com/modu-ai/moai-adk/cmd/moai@latest

# Copy project
COPY . .

# Initialize MoAI-ADK
RUN moai init

# Run Ralph quality check
RUN claude -p "/moai:loop --max-iterations 3" \
    --allowedTools "Read,Write,Edit" \
    --output-format json || exit 1

# Continue with application build
CMD ["python", "app.py"]
```

---

## Complete Integration Example

### Full Project Setup with Ralph

**Project Structure:**

```
my-project/
├── .moai/
│   ├── config/
│   │   └── sections/
│   │       └── ralph.yaml
│   ├── cache/
│   │   └── .moai_loop_state.json
│   └── ast-grep/
│       ├── security/
│       │   ├── sql-injection.yml
│       │   ├── xss-prevention.yml
│       │   └── secrets.yml
│       └── quality/
│           ├── complexity.yml
│           └── best-practices.yml
├── .claude/
│   ├── hooks/
│   │   └── moai/
│   │       ├── post_tool__lsp_diagnostic
│   │       └── stop__loop_controller
│   └── settings.json
├── .lsp.json
└── src/
```

**Complete ralph.yaml:**

```yaml
ralph:
  enabled: true

  lsp:
    auto_start: true
    timeout_seconds: 30
    graceful_degradation: true
    servers:
      python: "pyright"
      typescript: "tsserver"
    settings:
      pyright:
        typeCheckingMode: "strict"

  ast_grep:
    enabled: true
    security_scan: true
    quality_scan: true
    custom_rules:
      - .moai/ast-grep/security/**/*.yml
      - .moai/ast-grep/quality/**/*.yml

  loop:
    max_iterations: 10
    auto_fix: false
    require_confirmation: true
    completion:
      zero_errors: true
      zero_warnings: false
      tests_pass: true
      coverage_threshold: 85

  hooks:
    post_tool_lsp:
      enabled: true
      severity_threshold: "error"
    stop_loop_controller:
      enabled: true
      verbose: true
```

**Typical Workflow:**

```bash
# 1. Start development
/moai:1-plan "User authentication system"

# 2. Implement with Ralph loop
/moai:loop

# (Ralph automatically fixes LSP errors during implementation)

# 3. Verify quality
/moai:fix --severity warning

# 4. Run security scan
moai ast-grep scan --security

# 5. Sync documentation
/moai:3-sync
```

---

## Troubleshooting Examples

### Debug LSP Issues

**Check LSP Server Status:**

```bash
# View LSP logs
cat .moai/logs/lsp_diagnostic.log

# Test LSP connection
moai lsp test-connection python

# View diagnostics directly
moai lsp diagnose src/auth.py
```

**Common LSP Errors:**

```
Error: LSP server not found
Solution: Install language server (e.g., pip install pyright)

Error: LSP timeout
Solution: Increase timeout in ralph.yaml:
  lsp:
    timeout_seconds: 60

Error: Invalid LSP configuration
Solution: Verify .lsp.json configuration format
```

### Debug Loop Issues

**Check Loop State:**

```bash
# View current state
cat .moai/cache/.moai_loop_state.json

# Reset loop state
rm .moai/cache/.moai_loop_state.json

# View loop logs
cat .moai/logs/loop_controller.log
```

**Common Loop Issues:**

```
Issue: Loop not starting
Check:
  - ralph.enabled: true in config
  - MOAI_DISABLE_LOOP_CONTROLLER not set
  - State file writable

Issue: Loop stuck
Check:
  - Max iterations setting
  - Completion conditions
  - Error count not decreasing
Solution: Send any message to stop loop, then review errors manually

Issue: Loop completes too early
Check:
  - Completion conditions too relaxed
  - zero_warnings: false (warnings ignored)
Solution: Tighten completion criteria in ralph.yaml
```

---

Last Updated: 2026-01-10
Version: 1.0.0
