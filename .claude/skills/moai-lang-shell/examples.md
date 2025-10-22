# Shell Scripting Examples

## Example 1: Deployment Script with TDD

### RED: Write Failing Test

```bash
# tests/test_deploy.bats
#!/usr/bin/env bats

@test "deployment validates environment" {
  run ./deploy.sh
  [ "$status" -ne 0 ]
  [[ "$output" =~ "DEPLOY_ENV required" ]]
}

@test "deployment creates backup before deploy" {
  export DEPLOY_ENV="staging"
  export BACKUP_DIR="/tmp/backups"

  run ./deploy.sh
  [ "$status" -eq 0 ]
  [ -d "$BACKUP_DIR" ]
}

@test "deployment rolls back on failure" {
  export DEPLOY_ENV="staging"
  export FORCE_FAILURE="true"

  run ./deploy.sh
  [ "$status" -ne 0 ]
  [[ "$output" =~ "Rolling back" ]]
}
```

### GREEN: Implement Feature

```bash
#!/usr/bin/env bash
# @CODE:DEPLOY-001 | SPEC: SPEC-DEPLOY-001.md | TEST: tests/test_deploy.bats
# Deployment script with validation and rollback
# @HISTORY: v0.1.0 (2025-01-15): Initial deployment implementation

set -euo pipefail
IFS=$'\n\t'

# Configuration
DEPLOY_ENV="${DEPLOY_ENV:-}"
BACKUP_DIR="${BACKUP_DIR:-/tmp/backups}"
APP_DIR="/var/www/app"

# Logging
log_info() {
  echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') - $*" >&2
}

log_error() {
  echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') - $*" >&2
}

# Validation
validate_environment() {
  if [[ -z "$DEPLOY_ENV" ]]; then
    log_error "DEPLOY_ENV required"
    return 1
  fi

  if [[ ! "$DEPLOY_ENV" =~ ^(staging|production)$ ]]; then
    log_error "Invalid environment: $DEPLOY_ENV"
    return 1
  fi
}

# Backup
create_backup() {
  log_info "Creating backup"

  mkdir -p "$BACKUP_DIR"
  local backup_file="$BACKUP_DIR/backup-$(date +%Y%m%d-%H%M%S).tar.gz"

  tar -czf "$backup_file" -C "$APP_DIR" . || {
    log_error "Backup failed"
    return 1
  }

  log_info "Backup created: $backup_file"
  echo "$backup_file"
}

# Deploy
deploy() {
  local backup_file="$1"

  log_info "Deploying to $DEPLOY_ENV"

  # Simulate deployment (replace with actual logic)
  if [[ "${FORCE_FAILURE:-}" == "true" ]]; then
    log_error "Deployment failed (forced)"
    rollback "$backup_file"
    return 1
  fi

  log_info "Deployment successful"
}

# Rollback
rollback() {
  local backup_file="$1"

  log_info "Rolling back from: $backup_file"

  tar -xzf "$backup_file" -C "$APP_DIR" || {
    log_error "Rollback failed"
    return 1
  }

  log_info "Rollback complete"
}

# Main
main() {
  validate_environment || exit 1

  local backup_file
  backup_file="$(create_backup)" || exit 1

  deploy "$backup_file" || exit 1
}

main "$@"
```

### REFACTOR: Improve Code

```bash
# Run ShellCheck
shellcheck deploy.sh

# Findings:
# - All variables quoted ✓
# - Error handling present ✓
# - Functions properly scoped ✓
# - POSIX compliance (with bash extensions) ✓

# Run tests
bats tests/test_deploy.bats

# ✓ deployment validates environment
# ✓ deployment creates backup before deploy
# ✓ deployment rolls back on failure
#
# 3 tests, 0 failures
```

---

## Example 2: Data Processing Script with Validation

### Script Implementation

```bash
#!/usr/bin/env bash
# @CODE:DATA-001 | SPEC: SPEC-DATA-001.md | TEST: tests/test_process.bats

set -euo pipefail

# Validate input file
validate_file() {
  local file="$1"

  [[ -f "$file" ]] || {
    echo "Error: File not found: $file" >&2
    return 1
  }

  [[ -r "$file" ]] || {
    echo "Error: File not readable: $file" >&2
    return 1
  }
}

# Process CSV data
process_csv() {
  local input="$1"
  local output="${2:-output.csv}"

  validate_file "$input" || return 1

  # Read and transform data
  while IFS=',' read -r col1 col2 col3; do
    # Skip header
    [[ "$col1" == "Header" ]] && continue

    # Validate data
    [[ "$col1" =~ ^[0-9]+$ ]] || {
      echo "Warning: Invalid ID: $col1" >&2
      continue
    }

    # Transform and output
    echo "${col1},${col2^^},${col3}"  # Uppercase col2
  done < "$input" > "$output"

  echo "Processed: $output"
}

# Main
if [[ $# -lt 1 ]]; then
  echo "Usage: $0 INPUT [OUTPUT]" >&2
  exit 1
fi

process_csv "$@"
```

### Test Suite

```bash
# tests/test_process.bats
#!/usr/bin/env bats

setup() {
  TEST_DIR="$(mktemp -d)"
  TEST_INPUT="$TEST_DIR/input.csv"
  TEST_OUTPUT="$TEST_DIR/output.csv"
}

teardown() {
  rm -rf "$TEST_DIR"
}

@test "process valid CSV file" {
  cat > "$TEST_INPUT" <<EOF
Header,Name,Value
1,test,100
2,data,200
EOF

  run bash process.sh "$TEST_INPUT" "$TEST_OUTPUT"

  [ "$status" -eq 0 ]
  [ -f "$TEST_OUTPUT" ]

  # Check first data line (skip header from input)
  result="$(head -n 1 "$TEST_OUTPUT")"
  [ "$result" = "1,TEST,100" ]
}

@test "handle missing file" {
  run bash process.sh "nonexistent.csv"

  [ "$status" -ne 0 ]
  [[ "$output" =~ "File not found" ]]
}

@test "validate numeric IDs" {
  cat > "$TEST_INPUT" <<EOF
Header,Name,Value
invalid,test,100
2,data,200
EOF

  run bash process.sh "$TEST_INPUT" "$TEST_OUTPUT"

  [ "$status" -eq 0 ]
  [[ "$output" =~ "Invalid ID" ]]

  # Only valid row processed
  lines="$(wc -l < "$TEST_OUTPUT")"
  [ "$lines" -eq 1 ]
}
```

---

## Example 3: CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/shell-quality.yml
name: Shell Script Quality

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Install ShellCheck
        run: |
          sudo apt-get update
          sudo apt-get install -y shellcheck

      - name: Install bats-core
        run: |
          git clone https://github.com/bats-core/bats-core.git
          cd bats-core
          sudo ./install.sh /usr/local

      - name: Run ShellCheck
        run: |
          find . -name "*.sh" -type f -exec shellcheck {} +

      - name: Run Tests
        run: |
          bats --tap tests/

      - name: Check Coverage
        run: |
          # Use kcov or similar for coverage reporting
          ./scripts/check-coverage.sh
```

### Pre-commit Hook

```bash
#!/usr/bin/env bash
# .git/hooks/pre-commit

set -e

echo "Running ShellCheck..."
find . -name "*.sh" -type f -exec shellcheck {} +

echo "Running bats tests..."
if [ -d "tests" ]; then
  bats tests/
fi

echo "✓ All checks passed"
```

---

## Example 4: Argument Parsing with Help

```bash
#!/usr/bin/env bash
# @CODE:CLI-001 | SPEC: SPEC-CLI-001.md

set -euo pipefail

# Default values
VERBOSE=false
DRY_RUN=false
CONFIG_FILE=""

# Show help
show_help() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Options:
  -h          Show this help message
  -v          Enable verbose output
  -n          Dry run mode
  -c FILE     Configuration file

Examples:
  $(basename "$0") -c config.ini
  $(basename "$0") -vn -c config.ini
EOF
}

# Parse options
while getopts "hvnc:" opt; do
  case "$opt" in
    h) show_help; exit 0 ;;
    v) VERBOSE=true ;;
    n) DRY_RUN=true ;;
    c) CONFIG_FILE="$OPTARG" ;;
    *) show_help; exit 1 ;;
  esac
done
shift $((OPTIND - 1))

# Validate required arguments
if [[ -z "$CONFIG_FILE" ]]; then
  echo "Error: -c CONFIG_FILE required" >&2
  show_help
  exit 1
fi

if [[ ! -f "$CONFIG_FILE" ]]; then
  echo "Error: Config file not found: $CONFIG_FILE" >&2
  exit 1
fi

# Main logic
$VERBOSE && echo "Verbose mode enabled"
$DRY_RUN && echo "Dry run mode enabled"

echo "Using config: $CONFIG_FILE"
# Process...
```

---

## Example 5: POSIX-Compliant Script

```bash
#!/bin/sh
# @CODE:POSIX-001 | SPEC: SPEC-POSIX-001.md
# Strict POSIX compliance for maximum portability

set -eu

# POSIX-compliant functions
is_numeric() {
  case "$1" in
    ''|*[!0-9]*) return 1 ;;
    *) return 0 ;;
  esac
}

file_exists() {
  [ -f "$1" ]
}

# Main
main() {
  value="${1:-}"

  if [ -z "$value" ]; then
    echo "Error: Value required" >&2
    return 1
  fi

  if is_numeric "$value"; then
    echo "Valid number: $value"
  else
    echo "Error: Not a number: $value" >&2
    return 1
  fi
}

main "$@"
```

### Test POSIX Compliance

```bash
# Verify POSIX compliance
shellcheck --shell=sh --posix posix-script.sh

# Test with different shells
sh posix-script.sh 42       # POSIX sh
dash posix-script.sh 42     # Debian Almquist Shell
bash posix-script.sh 42     # Bash (POSIX mode)
```

---

## Resources

- **bats-core Tutorial**: https://bats-core.readthedocs.io/en/stable/tutorial.html
- **ShellCheck Gallery**: https://github.com/koalaman/shellcheck#gallery-of-bad-code
- **Google Shell Style Guide**: https://google.github.io/styleguide/shellguide.html

---

**Version**: 2.0.0
**Last Updated**: 2025-10-22
