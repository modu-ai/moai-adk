---
name: "moai-lang-shell"
version: "4.0.0"
tier: Language
description: Enterprise Shell scripting with Bash 5.2, ShellCheck, bats-core testing, POSIX compliance, and Context7 MCP integration for defensive scripting patterns.
primary-agent: "alfred"
secondary-agents: ["tdd-implementer", "test-engineer", "code-reviewer"]
keywords: ["shell", "bash", "shellcheck", "bats", "posix", "scripting", "testing", "linting"]
status: stable
allowed-tools:
  - Read
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# moai-lang-shell - Enterprise v4.0.0

## Skill Overview

**Enterprise Shell scripting** with Bash 5.2+, static analysis (ShellCheck), bats-core testing framework, POSIX compliance, and defensive scripting patterns. This skill provides production-grade shell scripting guidance for system administration, CI/CD, and infrastructure automation.

### Core Capabilities

- ✅ Bash 5.2+ modern features and best practices
- ✅ ShellCheck static analysis integration (0.10.0)
- ✅ bats-core testing framework (1.11.0)
- ✅ POSIX sh compliance validation
- ✅ Defensive scripting patterns (set -euo pipefail)
- ✅ Error handling and signal trapping
- ✅ Context7 MCP for real-time documentation
- ✅ Performance optimization techniques

---

## When to Use This Skill

**Automatic activation**:
- Shell script development (.sh, .bash files)
- POSIX compliance requirements
- Shell script testing setup (bats-core)
- CI/CD shell automation
- System administration scripts

**Manual invocation**:
- Design defensive shell scripts
- Implement error handling patterns
- Optimize shell script performance
- Review shell script quality

---

## Technology Stack (2025-11-13)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| **Bash** | 5.2.37 | Core shell | Current |
| **ShellCheck** | 0.10.0 | Static analysis | Current |
| **bats-core** | 1.11.0 | Testing framework | Current |
| **bats-assert** | 2.1.0 | Test assertions | Current |
| **jq** | 1.7.1 | JSON processing | Optional |
| **sed/awk** | GNU | Text processing | Built-in |

---

## Core Language Features

### 1. Variables and Parameter Expansion

Bash provides powerful parameter expansion for safe string handling:

```bash
# Safe defaults
config="${CONFIG_FILE:-/etc/app/config.txt}"

# Parameter expansion
filename="document.txt"
name="${filename%.*}"           # Remove extension
extension="${filename##*.}"     # Get extension only

# Default values and error handling
value="${undefined:?Variable must be set}"

# Substring extraction
text="Hello World"
echo "${text:0:5}"              # "Hello"
echo "${text:6}"                # "World"
```

**Key Benefits**:
- Avoid repeated variable checks
- Safe defaults prevent errors
- No external tools needed

### 2. Structured Concurrency with Signals

Proper signal handling prevents zombies and resource leaks:

```bash
#!/bin/bash
set -euo pipefail               # Defensive mode

cleanup() {
    [[ -n "${temp_file:-}" ]] && rm -f "$temp_file"
    [[ -n "${bg_pid:-}" ]] && kill $bg_pid 2>/dev/null || true
}

trap cleanup EXIT               # Cleanup on exit
trap 'exit 130' INT             # Handle Ctrl+C
trap 'exit 143' TERM            # Handle termination
```

**Key Benefits**:
- Automatic resource cleanup
- Zombie process prevention
- Graceful signal handling

### 3. Arrays and Associative Arrays

```bash
# Indexed arrays
files=("file1.txt" "file2.txt")
for file in "${files[@]}"; do
    process "$file"
done

# Associative arrays (dictionaries)
declare -A config
config[url]="https://api.example.com"
config[key]="secret123"
echo "${config[url]}"
```

---

## Defensive Scripting Patterns

### Error Handling

```bash
# Set error options
set -euo pipefail

# Error trapping
error_handler() {
    echo "Error on line $1: $(tail -1 <(sed -n "$1p" "$0"))" >&2
    exit 1
}
trap 'error_handler $LINENO' ERR

# Safe command execution
if ! command_that_might_fail; then
    echo "Failed with exit code $?" >&2
    exit 1
fi
```

### Input Validation

```bash
validate_not_empty() {
    local var=$1
    local name=$2
    [[ -z "$var" ]] && { echo "Error: $name is empty" >&2; return 1; }
}

validate_file() {
    local file=$1
    [[ ! -f "$file" ]] && { echo "Error: File not found: $file" >&2; return 1; }
}

# Usage
validate_not_empty "$username" "Username" || exit 1
validate_file "$config_file" || exit 1
```

---

## ShellCheck Integration

### Common Checks

```bash
# Good: Quoted variables
echo "$var"                     # SC2086 prevention

# Good: Command existence check
command -v python3 &>/dev/null || {
    echo "Python 3 not found" >&2
    exit 1
}

# Good: Proper array iteration
for file in "${files[@]}"; do   # SC2068 prevention
    process "$file"
done

# Disable warnings when necessary
# shellcheck disable=SC1091
source /etc/profile.d/custom.sh
```

---

## Testing with bats-core

### Writing Tests

```bash
#!/usr/bin/env bats

source ./functions.sh

@test "add function returns correct sum" {
    result=$(add 5 3)
    [[ "$result" == "8" ]]
}

@test "error handling returns error code" {
    run safe_divide 10 0
    [[ $status -eq 1 ]]
    [[ "$output" =~ "Error" ]]
}

@test "file operations handle missing files" {
    run process_file "nonexistent.txt"
    [[ $status -eq 1 ]]
}
```

### Running Tests

```bash
# Run all tests
bats tests/

# Run specific test
bats tests/test_functions.bats -f "add function"

# Run with TAP output
bats --tap tests/
```

---

## Performance Optimization

### Built-in Commands vs External

```bash
# GOOD: Use built-in commands
while read -r line; do
    echo "$line"
done < file.txt

# AVOID: Unnecessary external commands
cat file.txt | while read -r line
```

### Sequence vs Stream Processing

```bash
# Lazy evaluation with process substitution
diff <(sort file1.txt) <(sort file2.txt)

# Avoid:
sort file1.txt > temp1
sort file2.txt > temp2
diff temp1 temp2
rm temp1 temp2
```

### Local Variables in Functions

```bash
# GOOD: Local scope prevents polluting global namespace
fast_function() {
    local count=0
    local result=""
    for i in {1..1000}; do
        ((count++))
    done
    echo "$count"
}

# AVOID: Global variables cause issues
function slow_function() {
    count=0
    result=""
    # ...
}
```

---

## POSIX Compliance

For maximum portability across different shells (bash, dash, ksh):

```bash
#!/bin/sh
# POSIX-compliant script

# Use [ ] instead of [[ ]]
if [ -f "file.txt" ]; then
    echo "File exists"
fi

# Use $() instead of backticks
date=$(date +%Y-%m-%d)

# Use expr for arithmetic (POSIX)
result=$(expr 5 + 3)

# POSIX arrays (positional parameters)
set -- "apple" "banana" "cherry"
echo "First: $1"
echo "All: $*"
```

---

## Best Practices Summary

### 1. Always Start with Defensive Mode

```bash
#!/bin/bash
set -euo pipefail               # Exit on error, undefined, pipe failure
IFS=$'\n\t'                     # Safer default separator
```

### 2. Quote All Variables

```bash
# GOOD
echo "$var"
for file in "${files[@]}"

# AVOID
echo $var
for file in $files
```

### 3. Use Local Variables in Functions

```bash
function my_func() {
    local var="value"
    local result=$(compute)
    echo "$result"
}
```

### 4. Trap for Cleanup

```bash
trap cleanup EXIT
trap 'error_handler $LINENO' ERR
```

### 5. Validate Input Early

```bash
[[ -z "$username" ]] && { echo "Username required" >&2; exit 1; }
[[ ! -f "$config" ]] && { echo "Config not found" >&2; exit 1; }
```

---

## Common Patterns

### Logging

```bash
log() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $*" | tee -a app.log
}

log_error() {
    log "ERROR: $*" >&2
}
```

### Configuration Loading

```bash
load_config() {
    local file="${1:-.config}"
    [[ ! -f "$file" ]] && return 1

    while IFS='=' read -r key value; do
        [[ "$key" =~ ^#.*$ ]] && continue
        [[ -z "$key" ]] && continue
        export "$key=$value"
    done < "$file"
}
```

### Safe Temporary Files

```bash
temp_file=$(mktemp) || exit 1
trap 'rm -f "$temp_file"' EXIT

# Use temp_file
echo "data" > "$temp_file"
```

---

## Context7 MCP Integration

This skill integrates with Context7 for real-time documentation access:

- **Bash Manual**: `/gnu/bash`
- **GNU Coreutils**: `/gnu/coreutils`
- **POSIX Shell**: `/posix/shell`

---

## Testing Strategy

| Category | Target | Tools |
|----------|--------|-------|
| **Unit Tests** | 80% | bats-core |
| **Integration Tests** | 15% | bats + external tools |
| **Linting** | 100% | ShellCheck |

---

## Works Well With

- `moai-foundation-security` (Security patterns)
- `moai-foundation-testing` (Testing strategies)
- `moai-essentials-debug` (Debugging)
- `moai-cc-mcp-integration` (MCP integration)

---

## For Complete Information

- **Working Examples**: See `examples.md` (23 examples covering all patterns)
- **CLI Reference**: See `reference.md` (commands, syntax, troubleshooting)
- **GNU Bash Manual**: https://www.gnu.org/software/bash/manual/
- **ShellCheck**: https://www.shellcheck.net

---

## Changelog

- **v4.0.0** (2025-11-13): Refactored to Progressive Disclosure with comprehensive examples.md and reference.md
- **v3.0.0** (2025-03-15): Added POSIX compliance guide
- **v2.0.0** (2025-01-10): Testing and linting guide
- **v1.0.0** (2024-12-01): Initial release

---

_Last updated: 2025-11-13_
