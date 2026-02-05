# Ralph Engine Complete Reference

## Configuration Reference

### Complete ralph.yaml Options

| Option                                              | Type    | Default                                                | Description                                               |
| --------------------------------------------------- | ------- | ------------------------------------------------------ | --------------------------------------------------------- |
| `ralph.enabled`                                     | boolean | `true`                                                 | Enable/disable Ralph Engine globally                      |
| `ralph.lsp.auto_start`                              | boolean | `true`                                                 | Auto-start LSP servers when needed                        |
| `ralph.lsp.timeout_seconds`                         | integer | `30`                                                   | Timeout for LSP operations (seconds)                      |
| `ralph.lsp.poll_interval_ms`                        | integer | `500`                                                  | Polling interval for diagnostics (milliseconds)           |
| `ralph.lsp.graceful_degradation`                    | boolean | `true`                                                 | Continue if LSP unavailable (fallback to linters)         |
| `ralph.ast_grep.enabled`                            | boolean | `true`                                                 | Enable AST-grep security scanning                         |
| `ralph.ast_grep.config_path`                        | string  | `.claude/skills/moai-tool-ast-grep/rules/sgconfig.yml` | Path to AST-grep configuration                            |
| `ralph.ast_grep.security_scan`                      | boolean | `true`                                                 | Enable security vulnerability scanning                    |
| `ralph.ast_grep.quality_scan`                       | boolean | `true`                                                 | Enable code quality pattern scanning                      |
| `ralph.ast_grep.auto_fix`                           | boolean | `false`                                                | Auto-fix without confirmation (dangerous)                 |
| `ralph.loop.max_iterations`                         | integer | `10`                                                   | Maximum feedback loop iterations                          |
| `ralph.loop.auto_fix`                               | boolean | `false`                                                | Require confirmation before auto-fixing                   |
| `ralph.loop.require_confirmation`                   | boolean | `true`                                                 | Ask user before applying fixes                            |
| `ralph.loop.cooldown_seconds`                       | integer | `2`                                                    | Minimum time between loop iterations                      |
| `ralph.loop.completion.zero_errors`                 | boolean | `true`                                                 | Require zero LSP errors to complete                       |
| `ralph.loop.completion.zero_warnings`               | boolean | `false`                                                | Require zero LSP warnings to complete                     |
| `ralph.loop.completion.tests_pass`                  | boolean | `true`                                                 | Require all tests to pass                                 |
| `ralph.loop.completion.coverage_threshold`          | integer | `85`                                                   | Minimum test coverage percentage (0 to disable)           |
| `ralph.git.auto_branch`                             | boolean | `false`                                                | Auto-create git branches (use git-strategy.yaml instead)  |
| `ralph.git.auto_pr`                                 | boolean | `false`                                                | Auto-create pull requests (use git-strategy.yaml instead) |
| `ralph.hooks.post_tool_lsp.enabled`                 | boolean | `true`                                                 | Enable LSP diagnostic hook                                |
| `ralph.hooks.post_tool_lsp.trigger_on`              | list    | `["Write", "Edit"]`                                    | Tools that trigger LSP diagnostics                        |
| `ralph.hooks.post_tool_lsp.severity_threshold`      | string  | `"error"`                                              | Minimum severity to report: error, warning, info          |
| `ralph.hooks.stop_loop_controller.enabled`          | boolean | `true`                                                 | Enable loop controller hook                               |
| `ralph.hooks.stop_loop_controller.check_completion` | boolean | `true`                                                 | Check completion conditions on each response              |

### Example Configuration

```yaml
ralph:
  enabled: true

  lsp:
    auto_start: true
    timeout_seconds: 30
    poll_interval_ms: 500
    graceful_degradation: true

  ast_grep:
    enabled: true
    security_scan: true
    quality_scan: true
    auto_fix: false

  loop:
    max_iterations: 10
    auto_fix: false
    require_confirmation: true
    cooldown_seconds: 2

    completion:
      zero_errors: true
      zero_warnings: false
      tests_pass: true
      coverage_threshold: 85

  hooks:
    post_tool_lsp:
      enabled: true
      trigger_on: ["Write", "Edit"]
      severity_threshold: "error"

    stop_loop_controller:
      enabled: true
      check_completion: true
```

---

## Environment Variables Reference

| Variable                       | Type    | Description                            | Example                          |
| ------------------------------ | ------- | -------------------------------------- | -------------------------------- |
| `MOAI_DISABLE_LSP_DIAGNOSTIC`  | boolean | Disable LSP diagnostic hook            | `MOAI_DISABLE_LSP_DIAGNOSTIC=1`  |
| `MOAI_DISABLE_LOOP_CONTROLLER` | boolean | Disable loop controller hook           | `MOAI_DISABLE_LOOP_CONTROLLER=1` |
| `MOAI_LOOP_ACTIVE`             | boolean | Loop active flag (set by commands)     | `MOAI_LOOP_ACTIVE=1`             |
| `MOAI_LOOP_ITERATION`          | integer | Current iteration number               | `MOAI_LOOP_ITERATION=3`          |
| `CLAUDE_PROJECT_DIR`           | string  | Project root path (set by Claude Code) | `/path/to/project`               |

### Environment Variable Usage

Enable/Disable Hooks:

```bash
# Disable LSP diagnostics temporarily
export MOAI_DISABLE_LSP_DIAGNOSTIC=1
claude -p "Make changes"

# Disable loop controller
export MOAI_DISABLE_LOOP_CONTROLLER=1
claude -p "Single run only"
```

Set Loop State (for CI/CD):

```bash
# Start loop with iteration count
export MOAI_LOOP_ACTIVE=1
export MOAI_LOOP_ITERATION=0
claude -p "/moai:loop --max-iterations 5"
```

---

## API Reference

> Note: The Go edition provides LSP capabilities through compiled hook subcommands in internal/lsp/. The examples below show the conceptual API; actual invocation is via moai hook post-tool-use.

### MoAILSPClient

High-level LSP client interface for getting diagnostics, finding references, renaming symbols, and other LSP operations.

#### Constructor

```python
MoAILSPClient(project_root: str | Path)
```

Initialize the LSP client.

**Parameters:**

- `project_root`: Path to the project root directory

**Raises:**

- `LSPClientError`: If initialization fails

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

client = MoAILSPClient(project_root="/path/to/project")
```

#### Methods

##### get_diagnostics

```python
async def get_diagnostics(file_path: str) -> list[Diagnostic]
```

Get diagnostics for a file.

**Parameters:**

- `file_path`: Path to the file (relative or absolute)

**Returns:**

- List of `Diagnostic` objects

**Example:**

```python
diagnostics = await client.get_diagnostics("src/auth.py")

for diag in diagnostics:
    if diag.is_error():
        print(f"Error at line {diag.range.start.line}: {diag.message}")
```

##### find_references

```python
async def find_references(file_path: str, position: Position) -> list[Location]
```

Find all references to the symbol at position.

**Parameters:**

- `file_path`: Path to the file
- `position`: Position of the symbol (`Position(line, character)`)

**Returns:**

- List of `Location` objects

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

position = Position(line=45, character=10)
references = await client.find_references("src/user.py", position)

for ref in references:
    print(f"Found in {ref.uri} at line {ref.range.start.line}")
```

##### rename_symbol

```python
async def rename_symbol(file_path: str, position: Position, new_name: str) -> WorkspaceEdit
```

Rename the symbol at position across the project.

**Parameters:**

- `file_path`: Path to the file
- `position`: Position of the symbol
- `new_name`: New name for the symbol

**Returns:**

- `WorkspaceEdit` with all changes

**Example:**

```python
position = Position(line=10, character=5)
edit = await client.rename_symbol("src/user.py", position, "new_name")

print(f"Will modify {edit.file_count()} file(s)")
for uri, text_edits in edit.changes.items():
    print(f"  {uri}: {len(text_edits)} edit(s)")
```

##### get_hover_info

```python
async def get_hover_info(file_path: str, position: Position) -> HoverInfo | None
```

Get hover information for position.

**Parameters:**

- `file_path`: Path to the file
- `position`: Position to get hover info for

**Returns:**

- `HoverInfo` or `None` if not available

**Example:**

```python
position = Position(line=20, character=15)
hover = await client.get_hover_info("src/utils.py", position)

if hover:
    print(f"Documentation: {hover.contents}")
```

##### get_language_for_file

```python
def get_language_for_file(file_path: str) -> str | None
```

Get the language identifier for a file.

**Parameters:**

- `file_path`: Path to the file

**Returns:**

- Language identifier (e.g., "python", "typescript") or `None`

**Example:**

```python
language = client.get_language_for_file("src/app.py")
# Returns: "python"
```

##### ensure_server_running

```python
async def ensure_server_running(language: str) -> None
```

Ensure an LSP server is running for a language.

**Parameters:**

- `language`: Language identifier (e.g., "python", "typescript")

**Example:**

```python
await client.ensure_server_running("python")
```

##### cleanup

```python
async def cleanup() -> None
```

Clean up by stopping all LSP servers.

**Example:**

```python
await client.cleanup()
```

---

### Diagnostic Models

#### DiagnosticSeverity

Enum for diagnostic severity levels (LSP 3.17 specification).

| Value         | Integer | Description            |
| ------------- | ------- | ---------------------- |
| `ERROR`       | 1       | Reports an error       |
| `WARNING`     | 2       | Reports a warning      |
| `INFORMATION` | 3       | Reports an information |
| `HINT`        | 4       | Reports a hint         |

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

if diagnostic.severity == DiagnosticSeverity.ERROR:
    print("Critical error found")
```

#### Position

Zero-based line and character position in a text document.

**Attributes:**

- `line`: Line position (zero-based)
- `character`: Character offset on a line (zero-based)

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

# Line 45, character 10 (like an editor cursor)
pos = Position(line=44, character=10)
```

#### Range

Range in a text document expressed as start and end positions.

**Attributes:**

- `start`: Range's start position (inclusive)
- `end`: Range's end position (exclusive)

**Methods:**

- `contains(position: Position) -> bool`: Check if position is within range
- `is_single_line() -> bool`: Check if range spans only one line

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

range = Range(
    start=Position(line=10, character=0),
    end=Position(line=10, character=20)
)

if range.contains(Position(line=10, character=5)):
    print("Position is in range")

if range.is_single_line():
    print("Single-line range")
```

#### Diagnostic

Represents a diagnostic issue (error, warning, etc.) in source code.

**Attributes:**

- `range`: Range where the message applies
- `severity`: Diagnostic severity (`DiagnosticSeverity`)
- `code`: Diagnostic code (string, int, or None)
- `source`: Diagnostic source (e.g., "pyright", "mypy")
- `message`: Diagnostic message

**Methods:**

- `is_error() -> bool`: Check if severity is ERROR

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

diagnostic = Diagnostic(
    range=Range(Position(45, 0), Position(45, 10)),
    severity=DiagnosticSeverity.ERROR,
    code="E0602",
    source="pyright",
    message="Undefined name 'x'"
)

if diagnostic.is_error():
    print(f"Error: {diagnostic.message}")
```

#### Location

Location inside a resource (file path + range).

**Attributes:**

- `uri`: Resource URI (e.g., "file:///path/to/file.py")
- `range`: Range within the resource

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

location = Location(
    uri="file:///home/user/project/src/main.py",
    range=Range(Position(10, 0), Position(10, 20))
)
```

#### TextEdit

Text edit applicable to a text document.

**Attributes:**

- `range`: Range to be manipulated
- `new_text`: String to be inserted (empty for delete)

**Methods:**

- `is_delete() -> bool`: Check if edit is a deletion
- `is_insert() -> bool`: Check if edit is an insertion

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

# Replace text
edit = TextEdit(
    range=Range(Position(5, 0), Position(5, 10)),
    new_text="new_value"
)

# Delete text
delete_edit = TextEdit(
    range=Range(Position(10, 0), Position(11, 0)),
    new_text=""
)

if delete_edit.is_delete():
    print("This is a deletion")
```

#### WorkspaceEdit

Workspace edit represents changes to many resources.

**Attributes:**

- `changes`: Dict mapping URI to list of `TextEdit` objects

**Methods:**

- `file_count() -> int`: Get number of files affected

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

edit = WorkspaceEdit(changes={
    "file:///path/to/file1.py": [
        TextEdit(Range(Position(10, 0), Position(10, 5)), "new_name")
    ],
    "file:///path/to/file2.py": [
        TextEdit(Range(Position(20, 0), Position(20, 5)), "new_name")
    ]
})

print(f"Editing {edit.file_count()} file(s)")
```

#### HoverInfo

Hover information for a symbol.

**Attributes:**

- `contents`: Hover content (can be markdown)
- `range`: Optional range for the symbol

**Example:**

```python
# Go edition: internal/lsp/ package provides these capabilities

hover = HoverInfo(
    contents="**my_function**\n\nCalculates the sum of two numbers.",
    range=Range(Position(10, 0), Position(10, 11))
)
```

---

## Hook Specifications

### PostToolUse Hook (post_tool\_\_lsp_diagnostic)

Triggered after Write/Edit operations to check for LSP diagnostics.

#### Input Format

```json
{
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/path/to/file.py",
    "content": "..."
  }
}
```

#### Output Format

```json
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "LSP: 2 error(s), 3 warning(s) in file.py\n  - [ERROR] Line 45: undefined name 'x'\n  - [ERROR] Line 52: type mismatch\n  - [WARNING] Line 30: unused variable"
  }
}
```

#### Exit Codes

| Code | Meaning                         | Effect                                 |
| ---- | ------------------------------- | -------------------------------------- |
| 0    | No action needed                | Hook completes normally                |
| 2    | Attention needed (errors found) | Claude Code displays diagnostic output |

#### Configuration

Controlled by `ralph.hooks.post_tool_lsp` in ralph.yaml:

```yaml
hooks:
  post_tool_lsp:
    enabled: true
    trigger_on: ["Write", "Edit"]
    severity_threshold: "error" # error, warning, info
```

#### Disable Hook

```bash
export MOAI_DISABLE_LSP_DIAGNOSTIC=1
```

---

### Stop Hook (stop\_\_loop_controller)

Triggered after each Claude response to control feedback loop.

#### Input Format

```json
{
  "conversation_context": {
    "messages": [...],
    "current_task": "..."
  }
}
```

Note: Input is consumed but not currently used. Reserved for future enhancements.

#### Output Format

```json
{
  "hookSpecificOutput": {
    "hookEventName": "Stop",
    "additionalContext": "Ralph Loop: CONTINUE | Iteration: 3/10 | Errors: 2 | Warnings: 5 | Tests: FAIL | Coverage: 78.5%\nNext actions: Fix 2 error(s), Fix failing tests, Increase coverage from 78.5% to 85%"
  }
}
```

#### Exit Codes

| Code | Meaning                          | Effect                                    |
| ---- | -------------------------------- | ----------------------------------------- |
| 0    | Loop complete or inactive        | Claude Code stops processing              |
| 1    | Continue loop (more work needed) | Claude Code continues with next iteration |

#### Configuration

Controlled by `ralph.hooks.stop_loop_controller` in ralph.yaml:

```yaml
hooks:
  stop_loop_controller:
    enabled: true
    check_completion: true
```

#### Disable Hook

```bash
export MOAI_DISABLE_LOOP_CONTROLLER=1
```

---

## State File Format

### Loop State File (.moai_loop_state.json)

Location: `.moai/cache/.moai_loop_state.json`

#### Schema

```json
{
  "active": true,
  "iteration": 3,
  "max_iterations": 10,
  "last_error_count": 2,
  "last_warning_count": 5,
  "files_modified": ["src/auth.py", "src/user.py"],
  "start_time": 1704380400.0,
  "completion_reason": null
}
```

#### Field Descriptions

| Field                | Type           | Description                                                                     |
| -------------------- | -------------- | ------------------------------------------------------------------------------- |
| `active`             | boolean        | Whether the loop is currently active                                            |
| `iteration`          | integer        | Current iteration number (1-based)                                              |
| `max_iterations`     | integer        | Maximum allowed iterations                                                      |
| `last_error_count`   | integer        | Number of errors from last check                                                |
| `last_warning_count` | integer        | Number of warnings from last check                                              |
| `files_modified`     | array          | List of files modified during loop                                              |
| `start_time`         | float          | Unix timestamp when loop started                                                |
| `completion_reason`  | string or null | Reason for completion ("All conditions met", "Max iterations reached", or null) |

#### State Transitions

```
Initial State:
{
  "active": false,
  "iteration": 0,
  ...
}

After /moai:loop:
{
  "active": true,
  "iteration": 1,
  "start_time": <current_timestamp>,
  ...
}

During Loop (errors found):
{
  "active": true,
  "iteration": 2,
  "last_error_count": 3,
  ...
}

Completion (success):
{
  "active": false,
  "completion_reason": "All conditions met"
}

Completion (max iterations):
{
  "active": false,
  "iteration": 10,
  "completion_reason": "Max iterations reached"
}
```

---

## LSP Configuration

### .lsp.json Format

Location: `.lsp.json` (project root)

#### Schema

```json
{
  "servers": {
    "python": {
      "command": "pyright-langserver",
      "args": ["--stdio"],
      "file_extensions": [".py"],
      "initialization_options": {}
    },
    "typescript": {
      "command": "typescript-language-server",
      "args": ["--stdio"],
      "file_extensions": [".ts", ".tsx", ".js", ".jsx"]
    },
    "go": {
      "command": "gopls",
      "args": ["serve"],
      "file_extensions": [".go"]
    }
  },
  "global_settings": {
    "timeout_seconds": 30,
    "retry_attempts": 3
  }
}
```

#### Field Descriptions

**Server Configuration:**

- `command`: LSP server command (must be in PATH)
- `args`: Command-line arguments
- `file_extensions`: File extensions this server handles
- `initialization_options`: Server-specific initialization options (optional)

**Global Settings:**

- `timeout_seconds`: Default timeout for LSP operations
- `retry_attempts`: Number of retry attempts on failure

#### Supported Language Servers

| Language              | Server        | Command                              | Installation                                 |
| --------------------- | ------------- | ------------------------------------ | -------------------------------------------- |
| Python                | Pyright       | `pyright-langserver --stdio`         | `npm install -g pyright`                     |
| Python                | pylsp         | `pylsp`                              | `pip install python-lsp-server`              |
| TypeScript/JavaScript | tsserver      | `typescript-language-server --stdio` | `npm install -g typescript-language-server`  |
| Go                    | gopls         | `gopls serve`                        | `go install golang.org/x/tools/gopls@latest` |
| Rust                  | rust-analyzer | `rust-analyzer`                      | Via rustup                                   |
| Java                  | jdtls         | `jdtls`                              | Via Eclipse JDT LS                           |
| C/C++                 | clangd        | `clangd`                             | Via LLVM                                     |

---

## AST-grep Configuration

### sgconfig.yml Format

Location: `.claude/skills/moai-tool-ast-grep/rules/sgconfig.yml`

#### Schema

```yaml
ruleDirs:
  - rules/security
  - rules/quality

rules:
  - id: sql-injection
    language: python
    message: Potential SQL injection vulnerability
    severity: error
    pattern: execute($SQL)
    constraints:
      SQL:
        kind: string
        not:
          has:
            kind: identifier

  - id: xss-vulnerability
    language: typescript
    message: Potential XSS vulnerability
    severity: error
    pattern: innerHTML = $VAR
    constraints:
      VAR:
        not:
          matches: sanitize.*

  - id: unused-import
    language: python
    message: Unused import statement
    severity: warning
    pattern: import $MODULE
    fix: ""
```

#### Field Descriptions

**Top-Level:**

- `ruleDirs`: Directories containing additional rule files
- `rules`: List of rule definitions

**Rule Definition:**

- `id`: Unique rule identifier
- `language`: Target language (python, typescript, go, rust, etc.)
- `message`: Diagnostic message
- `severity`: Severity level (error, warning, info)
- `pattern`: AST-grep search pattern
- `constraints`: Pattern constraints (optional)
- `fix`: Auto-fix template (optional)

#### Pattern Syntax

Metavariables:

- `$VAR`: Matches any expression
- `$STMT`: Matches any statement
- `$FUNC`: Matches function names

Constraints:

```yaml
constraints:
  VAR:
    kind: identifier # Match specific AST node type
    matches: ^[A-Z].* # Regex pattern
    has: # Contains pattern
      kind: string
    not: # Negation
      matches: sanitize.*
```

---

## Troubleshooting

### Common Issues and Solutions

#### Loop Not Starting

**Symptoms:**

- `/moai:loop` command does nothing
- No loop state file created

**Solutions:**

1. Check if Ralph is enabled:

   ```yaml
   # .moai/config/sections/ralph.yaml
   ralph:
     enabled: true
   ```

2. Verify loop controller hook is enabled:

   ```yaml
   ralph:
     hooks:
       stop_loop_controller:
         enabled: true
   ```

3. Check environment variable:

   ```bash
   unset MOAI_DISABLE_LOOP_CONTROLLER
   ```

4. Verify state file is writable:
   ```bash
   mkdir -p .moai/cache
   chmod 755 .moai/cache
   ```

---

#### LSP Diagnostics Missing

**Symptoms:**

- No diagnostics after Write/Edit
- Hook exits with code 0 immediately

**Solutions:**

1. Check if LSP hook is enabled:

   ```yaml
   ralph:
     hooks:
       post_tool_lsp:
         enabled: true
   ```

2. Verify language server is installed:

   ```bash
   # Python
   which pyright-langserver
   pip install pyright

   # TypeScript
   which typescript-language-server
   npm install -g typescript-language-server
   ```

3. Check .lsp.json configuration:

   ```json
   {
     "servers": {
       "python": {
         "command": "pyright-langserver",
         "args": ["--stdio"]
       }
     }
   }
   ```

4. Check environment variable:

   ```bash
   unset MOAI_DISABLE_LSP_DIAGNOSTIC
   ```

5. Enable graceful degradation to use fallback linters:
   ```yaml
   ralph:
     lsp:
       graceful_degradation: true
   ```

---

#### Loop Stuck/Infinite Loop

**Symptoms:**

- Loop continues past max_iterations
- Never reaches completion

**Solutions:**

1. Check max_iterations setting:

   ```yaml
   ralph:
     loop:
       max_iterations: 10 # Increase if needed
   ```

2. Review completion conditions:

   ```yaml
   ralph:
     loop:
       completion:
         zero_errors: true
         zero_warnings: false # Set to false to allow warnings
         tests_pass: true
         coverage_threshold: 85 # Set to 0 to disable
   ```

3. Delete state file:
   ```bash
   rm .moai/cache/.moai_loop_state.json
   ```

---

#### Tests Not Detected

**Symptoms:**

- "No test framework detected" message
- tests_pass always true

**Solutions:**

1. Ensure test framework is installed:

   ```bash
   # Python
   pip install pytest

   # JavaScript/TypeScript
   npm install --save-dev jest
   ```

2. Verify test configuration exists:

   ```bash
   # Python
   ls pyproject.toml pytest.ini

   # JavaScript
   ls package.json
   ```

3. Check if tests can run manually:

   ```bash
   # Python
   pytest

   # JavaScript
   npm test
   ```

---

#### Coverage Not Reported

**Symptoms:**

- Coverage shows -1.0 or missing
- coverage_met always true

**Solutions:**

1. Install coverage tool:

   ```bash
   # Python
   pip install pytest-cov

   # JavaScript
   npm install --save-dev @coverage/jest
   ```

2. Generate coverage report:

   ```bash
   # Python
   pytest --cov --cov-report=json

   # JavaScript
   npm test -- --coverage --coverageReporters=json
   ```

3. Verify coverage file exists:

   ```bash
   ls coverage.json coverage.xml
   ```

4. Disable coverage requirement:
   ```yaml
   ralph:
     loop:
       completion:
         coverage_threshold: 0
   ```

---

#### AST-grep Not Running

**Symptoms:**

- No security/quality warnings
- ast_grep diagnostics missing

**Solutions:**

1. Check if AST-grep is enabled:

   ```yaml
   ralph:
     ast_grep:
       enabled: true
   ```

2. Verify ast-grep is installed:

   ```bash
   which sg
   cargo install ast-grep
   ```

3. Check configuration file exists:

   ```bash
   ls .claude/skills/moai-tool-ast-grep/rules/sgconfig.yml
   ```

4. Test ast-grep manually:
   ```bash
   sg scan --config sgconfig.yml
   ```

---

#### Graceful Degradation Not Working

**Symptoms:**

- Hook fails when LSP unavailable
- No fallback to linters

**Solutions:**

1. Enable graceful degradation:

   ```yaml
   ralph:
     lsp:
       graceful_degradation: true
   ```

2. Ensure fallback linters are installed:

   ```bash
   # Python
   pip install ruff

   # JavaScript
   npm install -g eslint

   # Go
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

3. Check linter configuration:

   ```bash
   # Python
   ls ruff.toml pyproject.toml

   # JavaScript
   ls .eslintrc.js .eslintrc.json
   ```

---

### Performance Optimization

#### Reduce LSP Timeout

For faster feedback in CI/CD:

```yaml
ralph:
  lsp:
    timeout_seconds: 15
    poll_interval_ms: 250
```

#### Disable Expensive Checks

For rapid iteration:

```yaml
ralph:
  loop:
    completion:
      zero_errors: true
      zero_warnings: false
      tests_pass: false # Disable for faster loops
      coverage_threshold: 0 # Disable coverage check
```

#### Use Specific Severity Threshold

Only report errors, not warnings:

```yaml
ralph:
  hooks:
    post_tool_lsp:
      severity_threshold: "error"
```

---

## Advanced Configuration Examples

### CI/CD Integration

GitHub Actions workflow with Ralph:

```yaml
name: Ralph Auto-Fix

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  ralph-fix:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: "3.13"

      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          npm install -g pyright

      - name: Run Ralph Loop
        run: |
          export MOAI_LOOP_ACTIVE=1
          export MOAI_LOOP_ITERATION=0
          claude -p "/moai:loop --max-iterations 5" \
            --allowedTools "Read,Write,Edit,Bash,Grep,Glob"

      - name: Commit fixes
        if: success()
        run: |
          git config user.name "Ralph Bot"
          git config user.email "ralph@moai-adk.dev"
          git add .
          git commit -m "fix: Auto-fixes from Ralph Engine" || true
          git push
```

---

### Multi-Language Project

Configuration for projects with multiple languages:

```json
{
  "servers": {
    "python": {
      "command": "pyright-langserver",
      "args": ["--stdio"],
      "file_extensions": [".py"]
    },
    "typescript": {
      "command": "typescript-language-server",
      "args": ["--stdio"],
      "file_extensions": [".ts", ".tsx", ".js", ".jsx"]
    },
    "go": {
      "command": "gopls",
      "args": ["serve"],
      "file_extensions": [".go"]
    },
    "rust": {
      "command": "rust-analyzer",
      "args": [],
      "file_extensions": [".rs"]
    }
  }
}
```

---

### Custom Completion Conditions

Extend the loop controller for project-specific checks:

```python
# .claude/hooks/moai/custom_completion_check
def check_custom_conditions() -> bool:
    """Add project-specific completion checks."""
    # Example 1: Check for TODO comments
    todos = count_todo_comments()
    if todos > 0:
        return False

    # Example 2: Check for print statements in production code
    prints = find_debug_prints()
    if prints:
        return False

    # Example 3: Verify API schema validity
    if not validate_openapi_schema():
        return False

    return True
```

Register in ralph.yaml:

```yaml
ralph:
  loop:
    completion:
      custom_checks:
        - .claude/hooks/moai/custom_completion_check
```

---

### Language-Specific AST-grep Rules

Python security rules:

```yaml
# rules/python-security.yml
rules:
  - id: eval-usage
    language: python
    message: Use of eval() is dangerous
    severity: error
    pattern: eval($EXPR)
    fix: "ast.literal_eval($EXPR)"

  - id: pickle-load
    language: python
    message: Pickle deserialization vulnerability
    severity: error
    pattern: pickle.load($FILE)
    note: "Use safer serialization formats like JSON"

  - id: hardcoded-password
    language: python
    message: Hardcoded password detected
    severity: error
    pattern: password = $VALUE
    constraints:
      VALUE:
        kind: string
        matches: ".{8,}"
```

TypeScript security rules:

```yaml
# rules/typescript-security.yml
rules:
  - id: dangerous-html
    language: typescript
    message: Dangerous innerHTML assignment
    severity: error
    pattern: $EL.innerHTML = $VAR
    constraints:
      VAR:
        not:
          matches: sanitize.*

  - id: eval-usage
    language: typescript
    message: Use of eval() is dangerous
    severity: error
    pattern: eval($CODE)

  - id: weak-crypto
    language: typescript
    message: Weak cryptographic algorithm
    severity: warning
    pattern: crypto.createHash($ALG)
    constraints:
      ALG:
        kind: string
        matches: (md5|sha1)
```

---

## Performance Metrics

### Typical Operation Times

| Operation                     | Average Time | Notes                           |
| ----------------------------- | ------------ | ------------------------------- |
| LSP diagnostics (single file) | 100-500ms    | Depends on file size and server |
| AST-grep scan (project)       | 500ms-2s     | Depends on project size         |
| Test execution                | 2-30s        | Depends on test count           |
| Coverage generation           | 3-15s        | Depends on project size         |
| Loop iteration (complete)     | 5-60s        | Sum of all checks               |

### Resource Usage

| Component            | Memory     | CPU    | Disk               |
| -------------------- | ---------- | ------ | ------------------ |
| LSP Client           | ~50MB      | Low    | None               |
| LSP Server (pyright) | ~100-300MB | Medium | None               |
| AST-grep             | ~20-50MB   | Medium | None               |
| Loop Controller      | ~10MB      | Low    | < 1KB (state file) |

---

Last Updated: 2026-01-10
Version: 1.0.0
Specification: LSP 3.17, AST-grep 0.20+
