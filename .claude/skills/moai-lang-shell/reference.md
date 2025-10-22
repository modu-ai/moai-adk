# Shell Scripting Reference

## Official Documentation

### Core Tools
- **Bash 5.2+**: https://www.gnu.org/software/bash/manual/
- **ShellCheck 0.10+**: https://www.shellcheck.net/ | https://github.com/koalaman/shellcheck
- **bats-core 1.11+**: https://bats-core.readthedocs.io/ | https://github.com/bats-core/bats-core

### Standards
- **POSIX Shell**: https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html
- **Bash Reference Manual**: https://www.gnu.org/software/bash/manual/bash.html

---

## TRUST 5 Principles for Shell

### T - Test First

**Primary Framework**: bats-core 1.11+

```bash
# tests/test_script.bats
#!/usr/bin/env bats

@test "function returns expected output" {
  source ./script.sh
  result="$(my_function "input")"
  [ "$result" = "expected" ]
}

@test "script handles missing arguments" {
  run ./script.sh
  [ "$status" -ne 0 ]
  [[ "$output" =~ "Usage:" ]]
}
```

**Coverage Target**: ≥85%

### R - Readable

**Primary Linter**: ShellCheck 0.10.0

```bash
# Run ShellCheck
shellcheck script.sh

# Check POSIX compliance
shellcheck --shell=sh script.sh

# Exclude specific warnings
shellcheck --exclude=SC2086,SC2034 script.sh

# Format: all warnings
shellcheck --format=gcc script.sh
```

**Style Guidelines**:
- Use `set -euo pipefail` for error handling
- Quote all variables: `"$var"` not `$var`
- Use `[[ ]]` for Bash, `[ ]` for POSIX
- Functions before main logic
- 2-space indentation

### U - Unified (Type Safety)

Shell doesn't have static typing, enforce via:

```bash
# Input validation
validate_integer() {
  [[ "$1" =~ ^[0-9]+$ ]] || {
    echo "Error: Expected integer, got: $1" >&2
    return 1
  }
}

# Parameter constraints
process_file() {
  local file="$1"
  [[ -f "$file" ]] || {
    echo "Error: File not found: $file" >&2
    return 1
  }
  # process file...
}
```

### S - Secured

**Security Checklist**:
- [ ] Never use `eval` on user input
- [ ] Validate all external inputs
- [ ] Quote all variable expansions
- [ ] Use `--` to separate options from arguments
- [ ] Check command injection vectors with ShellCheck

```bash
# ❌ BAD: Command injection risk
file="$1"
rm $file  # Vulnerable to word splitting

# ✅ GOOD: Properly quoted
file="$1"
rm -- "$file"
```

### T - Trackable

**TAG Integration**:

```bash
#!/usr/bin/env bash
# @CODE:DEPLOY-001 | SPEC: SPEC-DEPLOY-001.md | TEST: tests/test_deploy.bats
# Deploy script with validation and rollback
# @HISTORY: v0.1.0 (2025-01-15): Initial deployment implementation

set -euo pipefail

# Function implementation...
```

---

## ShellCheck Integration

### Configuration File

Create `.shellcheckrc`:

```bash
# Exclude specific warnings
disable=SC2086  # Double quote to prevent word splitting
disable=SC2155  # Declare and assign separately

# Enable optional checks
enable=avoid-nullary-conditions
enable=quote-safe-variables

# Set shell dialect
shell=bash
```

### Common Warnings

| Code | Issue | Fix |
|------|-------|-----|
| SC2086 | Unquoted variable | `"$var"` |
| SC2046 | Quote command substitution | `"$(cmd)"` |
| SC2068 | Quote array expansion | `"${array[@]}"` |
| SC2006 | Use `$()` instead of backticks | `$(cmd)` |
| SC2155 | Declare and assign separately | Split declaration |

### POSIX Compliance Testing

```bash
# Test for POSIX compliance
shellcheck --shell=sh --posix script.sh

# Verify shebang
#!/bin/sh  # POSIX-compliant

# Avoid Bash-specific features
[[ ]] → [ ]           # Use POSIX test
(( )) → expr or $(())  # Use POSIX arithmetic
[[ =~ ]] → grep        # Use external tool
```

---

## bats-core Testing

### Installation

```bash
# macOS
brew install bats-core

# Ubuntu/Debian
sudo apt-get install bats

# From source
git clone https://github.com/bats-core/bats-core.git
cd bats-core
sudo ./install.sh /usr/local
```

### Test Structure

```bash
#!/usr/bin/env bats

# Setup runs before each test
setup() {
  export TEST_DIR="$(mktemp -d)"
  export PATH="$BATS_TEST_DIRNAME/..:$PATH"
}

# Teardown runs after each test
teardown() {
  rm -rf "$TEST_DIR"
}

@test "description of test" {
  # Test code here
}
```

### Assertions

```bash
# Exit status
run command
[ "$status" -eq 0 ]    # Success
[ "$status" -ne 0 ]    # Failure

# Output matching
run command
[ "$output" = "exact match" ]
[[ "$output" =~ pattern ]]

# Line-by-line
run command
[ "${lines[0]}" = "first line" ]
[ "${#lines[@]}" -eq 3 ]  # 3 lines total

# File existence
[ -f "file.txt" ]
[ -d "directory" ]
```

### Running Tests

```bash
# Run all tests
bats tests/

# Run specific file
bats tests/test_main.bats

# TAP output
bats --tap tests/

# Show timing
bats --timing tests/

# Filter tests
bats --filter "pattern" tests/
```

---

## Best Practices

### Error Handling

```bash
#!/usr/bin/env bash
set -euo pipefail  # Exit on error, undefined var, pipe failure

# IFS for security
IFS=$'\n\t'

# Trap errors
trap 'echo "Error on line $LINENO" >&2' ERR

# Function error handling
process_data() {
  local input="$1"

  if [[ ! -f "$input" ]]; then
    echo "Error: File not found: $input" >&2
    return 1
  fi

  # Process...
}
```

### Argument Parsing

```bash
# POSIX-compliant option parsing
while getopts "hf:o:" opt; do
  case "$opt" in
    h) show_help; exit 0 ;;
    f) input_file="$OPTARG" ;;
    o) output_file="$OPTARG" ;;
    *) echo "Invalid option" >&2; exit 1 ;;
  esac
done
shift $((OPTIND - 1))

# Validate required arguments
if [[ -z "${input_file:-}" ]]; then
  echo "Error: -f FILE required" >&2
  exit 1
fi
```

### Performance Tips

```bash
# Use built-in commands over external tools
# ✅ GOOD: Built-in
if [[ "$var" =~ pattern ]]; then

# ❌ BAD: External process
if echo "$var" | grep -q pattern; then

# Read files efficiently
while IFS= read -r line; do
  process "$line"
done < file.txt

# Avoid subshells in loops
# ✅ GOOD: Process substitution
while read -r line; do
  ((count++))
done < <(command)

# ❌ BAD: Pipe creates subshell
command | while read -r line; do
  ((count++))  # Lost after loop
done
```

---

## Tool Version Matrix

| Tool | Version | Release | Status |
|------|---------|---------|--------|
| Bash | 5.2.37 | 2024-12 | ✅ Current |
| ShellCheck | 0.10.0 | 2024-03 | ✅ Current |
| bats-core | 1.11.0 | 2024-02 | ✅ Current |

---

## Common Pitfalls

### Word Splitting

```bash
# ❌ BAD
files="file1.txt file2.txt"
rm $files  # Breaks on spaces in filenames

# ✅ GOOD
files=("file1.txt" "file2.txt")
rm -- "${files[@]}"
```

### Command Substitution

```bash
# ❌ BAD
output=`command`  # Deprecated

# ✅ GOOD
output="$(command)"  # Modern syntax
```

### Exit Codes

```bash
# Preserve exit codes
command || status=$?

# Check specific codes
if command; then
  # Success
else
  case $? in
    1) echo "Not found" ;;
    2) echo "Invalid" ;;
    *) echo "Unknown error" ;;
  esac
fi
```

---

## Resources

- **Google Shell Style Guide**: https://google.github.io/styleguide/shellguide.html
- **Bash Pitfalls**: https://mywiki.wooledge.org/BashPitfalls
- **ShellCheck Wiki**: https://github.com/koalaman/shellcheck/wiki
- **Advanced Bash Scripting Guide**: https://tldp.org/LDP/abs/html/

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
**Official Sources**: ✅ Verified
