---
name: "moai-lang-shell"
version: "4.0.0"
tier: Language
description: "Enterprise Shell scripting mastery with ShellCheck 0.10, bats-core 1.11, POSIX compliance, and defensive scripting patterns based on 1,335+ production code examples."
primary-agent: "alfred"
secondary-agents: ["tdd-implementer", "test-engineer", "code-reviewer"]
keywords: ["shell", "bash", "shellcheck", "bats", "posix", "scripting", "testing", "linting"]
status: stable
---

# moai-lang-shell

**Enterprise Shell Scripting with ShellCheck, Bats, and POSIX Compliance**

## Overview

Enterprise-grade Shell scripting guidance covering Bash 5.2+, ShellCheck 0.10 static analysis, bats-core 1.11 testing framework, POSIX sh compliance, and defensive scripting patterns validated against 1,335+ production code examples.

**Core Capabilities**:
- ✅ ShellCheck static analysis integration (1,125 examples)
- ✅ Bats-core testing framework (177 examples)
- ✅ POSIX sh compliance validation
- ✅ Defensive scripting patterns
- ✅ Error handling and robustness
- ✅ TDD workflow for shell scripts

**Production Data Sources**:
- ShellCheck (Trust Score 8.2): 1,125 linting examples
- bats-core (Trust Score 7.0): 177 testing examples
- bats-assert (Trust Score 7.0): 33 assertion examples

---

## Quick Reference

### When to Use This Skill

**Automatic Activation**:
- Shell script development (`.sh`, `.bash` files)
- POSIX compliance requirements
- Shell script testing setup
- CI/CD shell automation
- System administration scripts

**Manual Invocation**:
```
Skill("moai-lang-shell")
```

### Key Patterns Covered

1. **ShellCheck Integration** - Static analysis and linting
2. **Bats Testing Framework** - Unit tests for shell scripts
3. **POSIX Compliance** - Portable shell scripting
4. **Defensive Scripting** - Error handling and robustness
5. **Quote Safety** - Variable expansion protection
6. **Command Substitution** - Modern `$()` syntax
7. **Conditional Logic** - `[[ ]]` vs `[ ]` vs `test`
8. **File Operations** - Safe file handling patterns

---

## Pattern 1: ShellCheck Static Analysis Integration

### Overview

ShellCheck is the industry-standard static analysis tool for shell scripts, detecting syntax errors, semantic issues, and anti-patterns.

### Basic ShellCheck Integration

```bash
#!/bin/bash
# Good: ShellCheck-compliant script with proper quoting

set -euo pipefail  # Exit on error, undefined vars, pipe failures

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly LOG_FILE="${SCRIPT_DIR}/app.log"

log_message() {
    local message="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ${message}" >> "${LOG_FILE}"
}

process_files() {
    local pattern="$1"
    
    # Good: Proper glob usage, quoted variables
    for file in "${SCRIPT_DIR}"/${pattern}; do
        if [[ -f "${file}" ]]; then
            log_message "Processing: ${file}"
            # Process file
        fi
    done
}

main() {
    if [[ $# -lt 1 ]]; then
        echo "Usage: $0 <pattern>" >&2
        exit 1
    fi
    
    process_files "$1"
}

main "$@"
```

### Common ShellCheck Fixes

**SC2086: Unquoted Variable Expansion**

```bash
# BAD: Word splitting and globbing occur
rm -rf $STEAMROOT/*

# GOOD: Proper quoting prevents disasters
rm -rf "${STEAMROOT:?}"/*
```

**SC2046: Quote Command Substitution**

```bash
# BAD: Word splitting on command output
for f in $(ls *.txt); do
    echo "$f"
done

# GOOD: Use glob directly
for f in *.txt; do
    echo "$f"
done

# GOOD: If you must use a command
while IFS= read -r f; do
    echo "$f"
done < <(find . -name "*.txt" -print0 | xargs -0 ls)
```

**SC2155: Declare and Assign Separately**

```bash
# BAD: Masks return value
local result="$(complex_command)"

# GOOD: Separate declaration and assignment
local result
result="$(complex_command)"
```

### ShellCheck in Makefile

```makefile
# Integrate ShellCheck into build process
.PHONY: check-scripts
check-scripts:
	@echo "Running ShellCheck..."
	@shellcheck --severity=warning --shell=bash scripts/*.sh
	@echo "✓ All scripts passed ShellCheck"

.PHONY: check-posix
check-posix:
	@echo "Checking POSIX compliance..."
	@shellcheck --severity=info --shell=sh --exclude=SC2039 scripts/*.sh
	@echo "✓ POSIX compliance verified"
```

### ShellCheck CI Integration

```yaml
# .gitlab-ci.yml
shellcheck:
  image: koalaman/shellcheck-alpine:latest
  stage: test
  before_script:
    - apk update && apk add git
  script:
    # Check all shell scripts (including those without .sh extension)
    - git ls-files -c -z | xargs -0 awk -vORS='\0' 'FNR==1 && /^#!.*sh/ { print FILENAME }' | xargs -0r shellcheck
```

---

## Pattern 2: Bats Testing Framework

### Overview

Bats (Bash Automated Testing System) is a TAP-compliant testing framework providing structured unit testing for shell scripts.

### Basic Bats Test Structure

```bash
#!/usr/bin/env bats
# test/calculator.bats

setup() {
    # Runs before each test
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
    
    # Add script directory to PATH
    DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")" && pwd)"
    PATH="$DIR/../src:$PATH"
}

teardown() {
    # Runs after each test
    rm -f /tmp/test_output_*
}

@test "addition using bc" {
    result="$(echo 2+2 | bc)"
    assert_equal "$result" "4"
}

@test "calculator script handles valid input" {
    run calculator.sh add 5 3
    assert_success
    assert_output "8"
}

@test "calculator script rejects invalid input" {
    run calculator.sh divide 10 0
    assert_failure
    assert_output --partial "ERROR: Division by zero"
}
```

### Setup and Teardown Hooks

```bash
#!/usr/bin/env bats

# File-level setup (once per file)
setup_file() {
    export TEST_DB="/tmp/test_db_$$"
    sqlite3 "$TEST_DB" "CREATE TABLE users (id INT, name TEXT);"
}

# File-level teardown (once per file)
teardown_file() {
    rm -f "$TEST_DB"
}

# Test-level setup (before each test)
setup() {
    sqlite3 "$TEST_DB" "DELETE FROM users;"
    sqlite3 "$TEST_DB" "INSERT INTO users VALUES (1, 'Alice');"
}

# Test-level teardown (after each test)
teardown() {
    # Cleanup temporary files
    rm -f /tmp/query_result_*
}

@test "can query existing user" {
    result="$(sqlite3 "$TEST_DB" "SELECT name FROM users WHERE id=1;")"
    assert_equal "$result" "Alice"
}
```

### Bats Assertions

```bash
#!/usr/bin/env bats

@test "assert_success: command exits with 0" {
    run echo "Success"
    assert_success
}

@test "assert_failure: command exits with non-zero" {
    run bash -c "exit 1"
    assert_failure
}

@test "assert_output: exact match" {
    run echo "hello world"
    assert_output "hello world"
}

@test "assert_output: partial match" {
    run echo "Error: file not found"
    assert_output --partial "Error"
}

@test "assert_output: regex match" {
    run echo "Version 1.2.3"
    assert_output --regexp '^Version [0-9]+\.[0-9]+\.[0-9]+$'
}

@test "assert_line: check specific line" {
    run printf "line 1\nline 2\nline 3"
    assert_line --index 1 "line 2"
}

@test "assert_equal: value comparison" {
    local expected="42"
    local actual="$((6 * 7))"
    assert_equal "$actual" "$expected"
}
```

### Advanced Bats: Conditional Skipping

```bash
#!/usr/bin/env bats

@test "runs only on Linux" {
    if [[ "$(uname)" != "Linux" ]]; then
        skip "This test requires Linux"
    fi
    
    run df -h
    assert_success
}

@test "requires docker" {
    if ! command -v docker &> /dev/null; then
        skip "Docker not installed"
    fi
    
    run docker ps
    assert_success
}
```

---

## Pattern 3: POSIX Compliance

### Overview

POSIX-compliant shell scripts ensure maximum portability across Unix-like systems, avoiding Bash-specific features.

### POSIX vs Bash Comparison

```bash
#!/bin/sh
# POSIX-compliant script

# GOOD: POSIX-compliant features
var=42
var=$((var + 1))
[ "$var" -eq 43 ]
case "$1" in
    start) echo "Starting..." ;;
    stop)  echo "Stopping..." ;;
    *)     echo "Unknown command" ;;
esac

# BAD: Bash-specific (not POSIX)
# [[ "$var" == 43 ]]           # Use [ ] instead
# array=(1 2 3)                # Arrays not in POSIX
# var+=5                       # Use var=$((var + 5))
# ${var:0:1}                   # Substring expansion not POSIX
# function foo() { ... }       # Use foo() { ... } instead
# let x=2+2                    # Use x=$((2 + 2))
```

### Testing POSIX Compliance

```bash
#!/bin/sh
# Use dash for strict POSIX testing

set -eu

check_posix_compliance() {
    # Test with dash (strict POSIX shell)
    dash -n "$1" && echo "✓ POSIX syntax valid"
    
    # Test with shellcheck
    shellcheck --shell=sh "$1"
}

# Arithmetic in POSIX sh
count=0
count=$((count + 1))

# String comparison in POSIX sh
if [ "$USER" = "root" ]; then
    echo "Running as root"
fi

# Command substitution in POSIX sh
current_dir="$(pwd)"
timestamp="$(date +%Y%m%d)"
```

### POSIX String Operations

```bash
#!/bin/sh
# POSIX-compliant string operations

text="hello_world"

# Length
length="${#text}"

# Substring removal from beginning
prefix="${text#hello_}"      # Returns: world
prefix="${text##*_}"         # Returns: world (greedy)

# Substring removal from end
suffix="${text%_*}"          # Returns: hello
suffix="${text%%_*}"         # Returns: hello (greedy)

# Default values
echo "${VAR:-default}"       # Use default if VAR unset
echo "${VAR:=default}"       # Assign default if VAR unset
echo "${VAR:?error message}" # Exit with error if VAR unset
```

---

## Pattern 4: Defensive Scripting

### Overview

Defensive scripting practices prevent common errors and make scripts robust in production environments.

### Safe Script Template

```bash
#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Exit on error (-e)
# Exit on undefined variable (-u)
# Exit on pipe failure (-o pipefail)
# Set IFS to safe values

readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Trap errors
trap 'echo "Error: ${SCRIPT_NAME} failed at line $LINENO" >&2' ERR

cleanup() {
    # Cleanup function called on exit
    rm -f /tmp/"${SCRIPT_NAME}"_*
}
trap cleanup EXIT

main() {
    # Main logic here
    echo "Script started"
}

main "$@"
```

### Error Handling Patterns

```bash
#!/bin/bash

# Check prerequisites
check_prerequisites() {
    local missing_tools=()
    
    for tool in jq curl git; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        echo "ERROR: Missing required tools: ${missing_tools[*]}" >&2
        return 1
    fi
}

# Safe file operations
safe_copy() {
    local src="$1"
    local dst="$2"
    
    if [[ ! -f "$src" ]]; then
        echo "ERROR: Source file does not exist: $src" >&2
        return 1
    fi
    
    if [[ -e "$dst" ]]; then
        echo "WARNING: Destination exists, creating backup" >&2
        cp "$dst" "${dst}.bak"
    fi
    
    cp "$src" "$dst"
}

# Validate user input
validate_input() {
    local input="$1"
    
    # Check for empty input
    if [[ -z "$input" ]]; then
        echo "ERROR: Input cannot be empty" >&2
        return 1
    fi
    
    # Check for dangerous characters
    if [[ "$input" =~ [^a-zA-Z0-9_-] ]]; then
        echo "ERROR: Input contains invalid characters" >&2
        return 1
    fi
    
    return 0
}
```

---

## Pattern 5: Quote Safety and Variable Expansion

### Overview

Proper quoting prevents word splitting, globbing, and injection vulnerabilities.

### Quote Safety Rules

```bash
#!/bin/bash

# GOOD: Always quote variables
file="my document.txt"
cat "${file}"                    # Correct
rm -f "${file}"                  # Correct

# BAD: Unquoted variables
cat $file                        # FAILS: cat tries to open "my" and "document.txt"
rm -f $file                      # DANGEROUS: Could delete wrong files

# GOOD: Quote command substitutions
current_user="$(whoami)"
echo "User: ${current_user}"

# BAD: Unquoted command substitution
current_user=$(whoami)           # Unquoted but works (no spaces in output)
echo "User: $current_user"       # Should still quote

# GOOD: Array expansion
files=("file1.txt" "file2.txt" "file 3.txt")
for file in "${files[@]}"; do    # Correct: preserves spaces
    echo "$file"
done

# BAD: Unquoted array
for file in ${files[@]}; do      # FAILS: splits "file 3.txt" into two items
    echo "$file"
done

# GOOD: Quote in conditionals
if [[ -f "${config_file}" ]]; then
    echo "Config found"
fi

# GOOD: Parameter expansion safety
echo "${var:-default}"           # Safe default value
echo "${var:?Variable not set}"  # Error if unset
```

### Special Variables Quoting

```bash
#!/bin/bash

# $@ - All positional parameters (preserve word boundaries)
process_args() {
    echo "Number of arguments: $#"
    
    # GOOD: Quote "$@" to preserve arguments
    for arg in "$@"; do
        echo "Argument: $arg"
    done
}

# $* - All positional parameters (single string)
log_all_args() {
    # GOOD: Use "$*" when you want single string
    echo "All args: $*" >> logfile.txt
}

# Example usage
process_args "file 1.txt" "file 2.txt"  # Correctly processes 2 args
```

---

## Pattern 6: Modern Command Substitution

### Overview

Modern `$()` syntax is more readable and nestable than legacy backticks.

### Command Substitution Best Practices

```bash
#!/bin/bash

# GOOD: Modern $() syntax (readable, nestable)
current_date="$(date +%Y-%m-%d)"
file_count="$(find . -type f | wc -l)"

# Nested command substitution
project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# BAD: Legacy backticks (deprecated)
# current_date=`date +%Y-%m-%d`    # Hard to read
# file_count=`find . -type f | wc -l`  # Doesn't nest well

# GOOD: Multi-line command substitution
users="$(
    awk -F: '$3 >= 1000 {print $1}' /etc/passwd |
    sort |
    uniq
)"

# GOOD: Error handling with command substitution
if output="$(risky_command 2>&1)"; then
    echo "Success: $output"
else
    echo "Failed: $output" >&2
    exit 1
fi
```

---

## Pattern 7: Conditional Logic Patterns

### Overview

Understanding `[[ ]]`, `[ ]`, and `test` for robust conditional expressions.

### Conditional Comparison Matrix

```bash
#!/bin/bash

# [[ ]] - Bash/Ksh extended test (recommended in Bash)
if [[ "$string" == pattern* ]]; then
    echo "Pattern matching works"
fi

if [[ "$var" =~ ^[0-9]+$ ]]; then
    echo "Variable is a number (regex)"
fi

if [[ -f "$file" && -r "$file" ]]; then
    echo "File exists and is readable"
fi

# [ ] - POSIX test (portable)
if [ "$a" = "$b" ]; then
    echo "Strings are equal (portable)"
fi

if [ "$num" -gt 10 ]; then
    echo "Number is greater than 10"
fi

# Arithmetic comparison
count=5
if (( count > 0 )); then
    echo "Count is positive"
fi

# String operations
name="Alice"
if [[ -n "$name" ]]; then        # String is not empty
    echo "Name is set"
fi

if [[ -z "$name" ]]; then        # String is empty
    echo "Name is empty"
fi
```

### File Test Operators

```bash
#!/bin/bash

filepath="/path/to/file.txt"

# File existence and type
[[ -e "$filepath" ]]    # Exists (any type)
[[ -f "$filepath" ]]    # Regular file
[[ -d "$filepath" ]]    # Directory
[[ -L "$filepath" ]]    # Symbolic link
[[ -p "$filepath" ]]    # Named pipe
[[ -S "$filepath" ]]    # Socket

# File permissions
[[ -r "$filepath" ]]    # Readable
[[ -w "$filepath" ]]    # Writable
[[ -x "$filepath" ]]    # Executable

# File comparisons
[[ "$file1" -nt "$file2" ]]  # file1 is newer than file2
[[ "$file1" -ot "$file2" ]]  # file1 is older than file2
```

---

## Pattern 8: Safe File Operations

### Overview

Defensive patterns for file processing preventing common errors and race conditions.

### Safe File Processing Loop

```bash
#!/bin/bash

# GOOD: Process files safely with globs
process_files_with_glob() {
    local pattern="$1"
    
    shopt -s nullglob  # Glob expands to nothing if no matches
    for file in "$pattern"; do
        if [[ -f "$file" ]]; then
            echo "Processing: $file"
            # Process file
        fi
    done
    shopt -u nullglob
}

# GOOD: Process files from find with null delimiter
process_files_with_find() {
    while IFS= read -r -d '' file; do
        echo "Processing: $file"
        # Process file
    done < <(find . -name "*.txt" -type f -print0)
}

# GOOD: Safe directory change
safe_cd() {
    local target="$1"
    
    if ! cd "$target"; then
        echo "ERROR: Cannot change to directory: $target" >&2
        return 1
    fi
}

# GOOD: Create temporary directory safely
temp_dir="$(mktemp -d)"
trap 'rm -rf "$temp_dir"' EXIT

# GOOD: Atomic file write
atomic_write() {
    local target="$1"
    local content="$2"
    local temp
    
    temp="$(mktemp)"
    echo "$content" > "$temp"
    mv "$temp" "$target"  # Atomic on same filesystem
}
```

---

## TDD Workflow for Shell Scripts

### RED Phase: Write Failing Test

```bash
#!/usr/bin/env bats
# test/greeter.bats

@test "greeter says hello with name" {
    run greeter.sh "Alice"
    assert_success
    assert_output "Hello, Alice!"
}
```

### GREEN Phase: Minimal Implementation

```bash
#!/bin/bash
# src/greeter.sh

set -euo pipefail

main() {
    local name="$1"
    echo "Hello, ${name}!"
}

main "$@"
```

### REFACTOR Phase: Improve Implementation

```bash
#!/bin/bash
# src/greeter.sh (refactored)

set -euo pipefail

readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

usage() {
    cat <<EOF
Usage: ${SCRIPT_NAME} <name>

Greet a person by name.

Arguments:
    name    The name of the person to greet

Example:
    ${SCRIPT_NAME} Alice
EOF
}

greet() {
    local name="$1"
    
    if [[ -z "$name" ]]; then
        echo "ERROR: Name cannot be empty" >&2
        return 1
    fi
    
    echo "Hello, ${name}!"
}

main() {
    if [[ $# -ne 1 ]]; then
        usage >&2
        exit 1
    fi
    
    greet "$1"
}

main "$@"
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/shell-tests.yml
name: Shell Script Tests

on: [push, pull_request]

jobs:
  shellcheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run ShellCheck
        uses: ludeeus/action-shellcheck@master
        with:
          severity: warning

  bats-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Bats
        uses: mig4/setup-bats@v1
        with:
          bats-version: 1.11.0
      - name: Run Bats tests
        run: bats test/*.bats
```

---

## Best Practices Checklist

**Script Header**:
- [ ] Shebang line specified (`#!/bin/bash` or `#!/bin/sh`)
- [ ] `set -euo pipefail` enabled (Bash)
- [ ] `set -eu` enabled (POSIX sh)
- [ ] Script description in comments

**Error Handling**:
- [ ] All commands checked for exit status
- [ ] Trap handlers defined for cleanup
- [ ] Error messages sent to stderr (`>&2`)
- [ ] Meaningful error messages with context

**Variable Safety**:
- [ ] All variables quoted (`"$var"`)
- [ ] Arrays quoted correctly (`"${array[@]}"`)
- [ ] Command substitutions quoted (`"$(cmd)"`)
- [ ] Parameter expansion with defaults (`"${var:-default}"`)

**Functions**:
- [ ] Functions have clear names
- [ ] Local variables declared
- [ ] Return values checked
- [ ] Functions documented

**Testing**:
- [ ] Bats test files created (`test/*.bats`)
- [ ] Setup/teardown hooks defined
- [ ] All code paths tested
- [ ] Edge cases covered

**Static Analysis**:
- [ ] ShellCheck passes with no warnings
- [ ] POSIX compliance checked (if required)
- [ ] No SC2086, SC2046, SC2155 violations
- [ ] Style consistent throughout

**Portability**:
- [ ] Bash version specified or POSIX used
- [ ] No Bashisms in POSIX scripts
- [ ] Dependencies documented
- [ ] Tested on target platforms

---

## Common Pitfalls to Avoid

### Pitfall 1: Unquoted Variables

```bash
# BAD
file="my document.txt"
rm $file  # Tries to delete "my" and "document.txt"

# GOOD
rm "$file"  # Deletes "my document.txt"
```

### Pitfall 2: Using ls for File Lists

```bash
# BAD
for f in $(ls *.txt); do
    echo "$f"
done

# GOOD
for f in *.txt; do
    [[ -e "$f" ]] || continue  # Handle no matches
    echo "$f"
done
```

### Pitfall 3: Unsafe cd

```bash
# BAD
cd "$target"
rm -rf *  # Could delete wrong directory if cd fails!

# GOOD
cd "$target" || exit 1
rm -rf *
```

### Pitfall 4: Word Splitting in Arrays

```bash
# BAD
files=$(find . -name "*.txt")
for f in $files; do  # Word splitting breaks on spaces
    echo "$f"
done

# GOOD
while IFS= read -r -d '' f; do
    echo "$f"
done < <(find . -name "*.txt" -print0)
```

---

## Tool Versions (2025)

| Tool | Version | Purpose |
|------|---------|---------|
| Bash | 5.2.37+ | Primary shell |
| ShellCheck | 0.10.0+ | Static analysis |
| bats-core | 1.11.0+ | Testing framework |
| bats-support | 0.3.0+ | Bats helpers |
| bats-assert | 2.1.0+ | Assertions |
| dash | 0.5.12+ | POSIX testing |

---

## References

- [ShellCheck Wiki](https://github.com/koalaman/shellcheck/wiki) - Comprehensive rule explanations
- [Bats Documentation](https://bats-core.readthedocs.io/) - Testing framework guide
- [POSIX Shell Spec](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html) - Official specification
- [Google Shell Style Guide](https://google.github.io/styleguide/shellguide.html) - Industry best practices
- [Defensive Bash Programming](http://www.kfirlavi.com/blog/2012/11/14/defensive-bash-programming/) - Robustness patterns

---

## Changelog

- **v4.0.0** (2025-01-12): Enterprise upgrade with 1,335+ production examples, comprehensive ShellCheck integration, Bats testing framework, POSIX compliance patterns
- **v2.0.0** (2025-10-22): Major update with latest tool versions
- **v1.0.0** (2025-03-29): Initial release

---

## Works Well With

- `moai-foundation-trust` - Quality gates and TRUST 5 validation
- `moai-alfred-code-reviewer` - Code review integration
- `moai-essentials-debug` - Debugging support
- `moai-alfred-dev-guide` - TDD workflow guidance
