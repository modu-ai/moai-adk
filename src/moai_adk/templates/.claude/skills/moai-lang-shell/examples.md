# moai-lang-shell - Working Examples

_Last updated: 2025-11-13_

## Quick Start (Bash 5.2 Minimal Setup)

```bash
#!/bin/bash
# Modern Bash with strict error handling

set -euo pipefail  # Exit on error, undefined vars, pipe failures

echo "Hello, World!"
```

Run it:
```bash
chmod +x script.sh
./script.sh
# Output: Hello, World!
```

## Basic Usage Examples

### Example 1: Variables and String Manipulation

```bash
#!/bin/bash

# Variable assignment
name="Alice"
age=30
echo "Name: $name, Age: $age"

# Concatenation
greeting="Hello, ${name}!"
echo "$greeting"

# Parameter expansion
file="document.txt"
echo "Filename: ${file%%.*}"     # Remove extension
echo "Extension: ${file##*.}"    # Get extension only

# Substring expansion
text="Hello World"
echo "${text:0:5}"              # "Hello"
echo "${text:6}"                # "World"

# Default values
value=${undefined:-"default"}   # Use default if not set
echo "$value"                   # Output: "default"
```

### Example 2: Conditional Logic

```bash
#!/bin/bash

# If-else statement
count=5
if [[ $count -gt 10 ]]; then
    echo "Greater than 10"
elif [[ $count -eq 5 ]]; then
    echo "Equal to 5"
else
    echo "Less than 10"
fi

# String comparison
name="Alice"
if [[ "$name" == "Alice" ]]; then
    echo "Welcome, Alice!"
fi

# File tests
if [[ -f "file.txt" ]]; then
    echo "File exists"
fi

if [[ -d "directory" ]]; then
    echo "Directory exists"
fi

# Logical operators
if [[ -f "file.txt" && -r "file.txt" ]]; then
    echo "File exists and is readable"
fi

if [[ -f "file1.txt" || -f "file2.txt" ]]; then
    echo "At least one file exists"
fi
```

### Example 3: Loops

```bash
#!/bin/bash

# For loop with range
for i in {1..5}; do
    echo "Number: $i"
done

# For loop with array
fruits=("apple" "banana" "cherry")
for fruit in "${fruits[@]}"; do
    echo "Fruit: $fruit"
done

# For loop with glob pattern
for file in *.txt; do
    [[ -f "$file" ]] && echo "File: $file"
done

# While loop
count=0
while [[ $count -lt 3 ]]; do
    echo "Count: $count"
    ((count++))
done

# Until loop (opposite of while)
count=0
until [[ $count -eq 3 ]]; do
    echo "Count: $count"
    ((count++))
done

# Continue and break
for i in {1..10}; do
    if [[ $i -eq 5 ]]; then
        continue  # Skip 5
    fi
    if [[ $i -eq 8 ]]; then
        break     # Exit at 8
    fi
    echo "Number: $i"
done
```

### Example 4: Functions

```bash
#!/bin/bash

# Simple function
greet() {
    echo "Hello, $1!"
}

greet "World"

# Function with return value
add() {
    local a=$1
    local b=$2
    echo $((a + b))
}

result=$(add 5 3)
echo "Sum: $result"  # Output: Sum: 8

# Function with multiple return values
get_user_info() {
    local id=$1
    local name="User_$id"
    local email="user${id}@example.com"
    echo "$name|$email"
}

IFS='|' read -r name email < <(get_user_info 123)
echo "Name: $name, Email: $email"

# Function with error handling
safe_divide() {
    local dividend=$1
    local divisor=$2

    if [[ $divisor -eq 0 ]]; then
        echo "Error: Division by zero" >&2
        return 1
    fi

    echo $((dividend / divisor))
}

safe_divide 10 2
echo "Result: $?"  # Exit status
```

## Intermediate Patterns

### Example 5: Arrays and Associative Arrays

```bash
#!/bin/bash

# Indexed array
fruits=("apple" "banana" "cherry")
echo "First fruit: ${fruits[0]}"
echo "All fruits: ${fruits[@]}"
echo "Fruit count: ${#fruits[@]}"

# Add to array
fruits+=("date")
echo "New fruits: ${fruits[@]}"

# Iterate with index
for i in "${!fruits[@]}"; do
    echo "[$i] = ${fruits[$i]}"
done

# Associative array (like a dictionary)
declare -A user
user[name]="Alice"
user[age]="30"
user[city]="Portland"

echo "Name: ${user[name]}"
echo "All keys: ${!user[@]}"
echo "All values: ${user[@]}"

# Iterate associative array
for key in "${!user[@]}"; do
    echo "$key: ${user[$key]}"
done
```

### Example 6: String Processing and Regular Expressions

```bash
#!/bin/bash

# Check if string matches pattern
email="user@example.com"
if [[ $email =~ ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$ ]]; then
    echo "Valid email"
fi

# Extract matched groups
text="Date: 2025-11-13"
if [[ $text =~ Date:\ ([0-9]{4})-([0-9]{2})-([0-9]{2}) ]]; then
    echo "Year: ${BASH_REMATCH[1]}"
    echo "Month: ${BASH_REMATCH[2]}"
    echo "Day: ${BASH_REMATCH[3]}"
fi

# String replacement
text="Hello World"
echo "${text/World/Bash}"  # First occurrence
echo "${text//l/L}"         # All occurrences

# Trimming
text="  Hello  "
trimmed="${text// }"       # Remove all spaces
echo "[$trimmed]"

# Character conversion
text="hello"
echo "${text^^}"            # Uppercase (Bash 4+)
echo "${text,,}"            # Lowercase (Bash 4+)
```

### Example 7: File Operations

```bash
#!/bin/bash

# Read file line by line
while IFS= read -r line; do
    echo "Line: $line"
done < "input.txt"

# Read file into array
mapfile -t lines < "input.txt"
for line in "${lines[@]}"; do
    echo "Line: $line"
done

# Write to file
cat > output.txt <<EOF
Line 1
Line 2
Line 3
EOF

# Append to file
echo "Line 4" >> output.txt

# File operations
if [[ -e "file.txt" ]]; then
    echo "File exists"
    echo "File size: $(stat -f%z file.txt)"  # macOS
    echo "Readable: $([[ -r "file.txt" ]] && echo yes || echo no)"
    echo "Writable: $([[ -w "file.txt" ]] && echo yes || echo no)"
fi

# Directory operations
mkdir -p "path/to/directory"
cd "path/to/directory" || exit 1
pwd
cd - || exit 1
```

### Example 8: Command Substitution and Pipelines

```bash
#!/bin/bash

# Command substitution (modern style)
current_date=$(date +%Y-%m-%d)
echo "Today: $current_date"

# Process substitution
diff <(sort file1.txt) <(sort file2.txt)

# Pipelines
echo -e "cherry\napple\nbanana" | sort | uniq

# Multiple pipes
find . -name "*.txt" | head -10 | wc -l

# Redirect and capture
output=$(ls -la 2>&1)
echo "Files and errors: $output"

# Pipe to variable
cat "file.txt" | while read -r line; do
    echo "Processing: $line"
done
```

### Example 9: Text Processing with sed and awk

```bash
#!/bin/bash

# Using sed for substitution
echo "Hello World" | sed 's/World/Bash/g'

# Using sed for line filtering
sed -n '2,4p' file.txt                    # Print lines 2-4
sed '/pattern/d' file.txt                 # Delete lines matching pattern
sed '/pattern/!d' file.txt                # Keep only lines matching pattern

# Using awk for column processing
echo -e "John 30\nAlice 25\nBob 35" | \
    awk '{print $1 " is " $2 " years old"}'

# AWK with field separator
echo "name,age,city" | \
    awk -F',' '{print $1 ": " $2}'

# AWK with conditionals
awk '$2 > 28 {print $1 " is over 28"}' < data.txt

# AWK with built-in variables
awk '{print NR ": " $0}' file.txt         # Line numbers
```

### Example 10: Error Handling and Debugging

```bash
#!/bin/bash

# Error handling with set
set -euo pipefail

# trap for cleanup
cleanup() {
    echo "Cleaning up..."
    rm -f tempfile.txt
}
trap cleanup EXIT

# Error trap
error_handler() {
    echo "Error on line $1" >&2
    exit 1
}
trap 'error_handler $LINENO' ERR

# Conditional execution
command1 && command2    # Run command2 only if command1 succeeds
command1 || command2    # Run command2 only if command1 fails

# Check command exists
if command -v python3 &> /dev/null; then
    echo "Python 3 is installed"
fi

# Check exit status
if false; then
    echo "Success"
else
    echo "Failed (exit code: $?)"
fi
```

## Advanced Patterns

### Example 11: Inline Classes and Case Statements

```bash
#!/bin/bash

case "$1" in
    start)
        echo "Starting service..."
        ;;
    stop)
        echo "Stopping service..."
        ;;
    restart)
        echo "Restarting service..."
        ;;
    *)
        echo "Usage: $0 {start|stop|restart}"
        exit 1
        ;;
esac

# Pattern matching in case
case "$filename" in
    *.txt)
        echo "Text file"
        ;;
    *.py)
        echo "Python file"
        ;;
    *)
        echo "Unknown file type"
        ;;
esac
```

### Example 12: Named Pipes and Process Substitution

```bash
#!/bin/bash

# Create named pipe
mkfifo mypipe

# Write to pipe in background
echo "Data" > mypipe &

# Read from pipe
cat < mypipe

# Remove pipe
rm mypipe

# Process substitution with diff
diff <(echo "line1"; echo "line2") \
     <(echo "line1"; echo "line3")
```

### Example 13: Parallel Processing

```bash
#!/bin/bash

# Background jobs
long_running_task() {
    echo "Task $1 starting"
    sleep $((RANDOM % 5))
    echo "Task $1 done"
}

# Run tasks in parallel
for i in {1..3}; do
    long_running_task $i &
done

# Wait for all background jobs
wait
echo "All tasks completed"

# Wait for specific job
background_task() {
    sleep 2
    echo "Background task done"
} &

job_pid=$!
echo "Job PID: $job_pid"

wait $job_pid
echo "Job completed"
```

### Example 14: Advanced Variable Expansion

```bash
#!/bin/bash

# Parameter expansion
var="Hello World"
echo "${var:0:5}"           # Substring: "Hello"
echo "${var: -5}"           # Last 5 chars: "World"
echo "${var/World/Bash}"    # Replace first

# Array expansion
arr=(a b c d e)
echo "${arr[@]:1:3}"        # Elements 1-3: "b c d"

# Default and assignment
value=${undefined:="default"}
echo "$value"               # "default"

# Indirect expansion
name="greeting"
greeting="Hello"
echo "${!name}"             # "Hello" (indirect reference)

# Length expansion
text="Hello"
echo "${#text}"             # 5

# Array length
arr=(1 2 3 4 5)
echo "${#arr[@]}"           # 5
```

### Example 15: Getopts for Command-Line Arguments

```bash
#!/bin/bash

usage() {
    echo "Usage: $0 [-v] [-o output] [-h]"
    echo "  -v        Enable verbose mode"
    echo "  -o FILE   Output file"
    echo "  -h        Show this help"
    exit 1
}

verbose=false
output_file=""

while getopts "vo:h" opt; do
    case $opt in
        v)
            verbose=true
            ;;
        o)
            output_file="$OPTARG"
            ;;
        h)
            usage
            ;;
        *)
            usage
            ;;
    esac
done

shift $((OPTIND - 1))

if [[ $verbose == true ]]; then
    echo "Verbose mode enabled"
fi

if [[ -n "$output_file" ]]; then
    echo "Output file: $output_file"
fi

echo "Remaining arguments: $*"
```

### Example 16: Configuration File Parsing

```bash
#!/bin/bash

# Load configuration from file
load_config() {
    local config_file="$1"

    if [[ ! -f "$config_file" ]]; then
        echo "Config file not found: $config_file" >&2
        return 1
    fi

    while IFS='=' read -r key value; do
        # Skip comments and empty lines
        [[ "$key" =~ ^#.*$ ]] && continue
        [[ -z "$key" ]] && continue

        # Trim whitespace
        key="${key// /}"
        value="${value// /}"

        export "$key=$value"
    done < "$config_file"
}

# Example config.txt:
# API_URL=https://api.example.com
# API_KEY=secret123
# DEBUG=true

load_config "config.txt"
echo "API URL: $API_URL"
echo "Debug: $DEBUG"
```

### Example 17: Logging Functions

```bash
#!/bin/bash

# Logging with timestamps
log() {
    local level=$1
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a app.log
}

log_info() {
    log "INFO" "$@"
}

log_error() {
    log "ERROR" "$@" >&2
}

log_debug() {
    [[ "${DEBUG:-false}" == "true" ]] && log "DEBUG" "$@"
}

# Usage
log_info "Application started"
log_error "An error occurred"
log_debug "Debug information"
```

### Example 18: JSON Processing with jq

```bash
#!/bin/bash

# Requires: jq (install with: brew install jq)

# Parse JSON
json='{"name":"Alice","age":30,"city":"Portland"}'

echo "$json" | jq '.name'           # "Alice"
echo "$json" | jq '.age'            # 30
echo "$json" | jq 'keys'            # ["age", "city", "name"]

# Array processing
json='[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]'

echo "$json" | jq '.[0].name'       # "Alice"
echo "$json" | jq '.[] | .name'     # "Alice" and "Bob"
echo "$json" | jq 'map(.name)'      # ["Alice", "Bob"]
echo "$json" | jq 'sort_by(.id)'    # Sorted array

# Create JSON
jq -n '{name:"Alice",age:30}'       # Create new object
jq -n '[1,2,3]'                     # Create array
```

### Example 19: Testing with bats-core

```bash
#!/usr/bin/env bats

# File: tests/test_script.bats

# Source the script to test
source ./script.sh

@test "add function returns correct sum" {
    result=$(add 5 3)
    [ "$result" = "8" ]
}

@test "add function with negative numbers" {
    result=$(add -5 3)
    [ "$result" = "-2" ]
}

@test "greeting function outputs correct message" {
    run greet "World"
    [ "$status" -eq 0 ]
    [[ "$output" == "Hello, World!" ]]
}

@test "file_exists returns true for existing file" {
    touch testfile.txt
    run file_exists testfile.txt
    [ "$status" -eq 0 ]
    rm testfile.txt
}
```

Run tests:
```bash
bats tests/test_script.bats
```

### Example 20: Comprehensive Script Template

```bash
#!/bin/bash

################################################################################
# Script: Deploy Application
# Purpose: Deploy application to production
# Usage: ./deploy.sh [-v] [-n] [-e ENV]
################################################################################

set -euo pipefail

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly LOG_FILE="${SCRIPT_DIR}/deploy.log"

# Global variables
verbose=false
dry_run=false
environment="production"

# Logging functions
log() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $*" | tee -a "$LOG_FILE"
}

log_error() {
    log "ERROR: $*" >&2
}

# Cleanup on exit
cleanup() {
    log "Cleanup: Removing temporary files"
    rm -f "$SCRIPT_DIR"/*.tmp
}
trap cleanup EXIT

# Parse arguments
parse_args() {
    while getopts "vne:h" opt; do
        case $opt in
            v) verbose=true ;;
            n) dry_run=true ;;
            e) environment="$OPTARG" ;;
            h)
                echo "Usage: $0 [-v] [-n] [-e ENV]"
                exit 0
                ;;
            *)
                log_error "Invalid option"
                exit 1
                ;;
        esac
    done
}

# Main deployment logic
deploy() {
    log "Starting deployment to $environment"

    if [[ "$dry_run" == "true" ]]; then
        log "DRY RUN: No changes will be made"
    fi

    # Deployment steps
    log "Building application..."
    log "Running tests..."
    log "Deploying to $environment..."
    log "Deployment completed successfully"
}

# Main entry point
main() {
    parse_args "$@"
    deploy
}

main "$@"
```

---

## Testing Examples

### Example 21: ShellCheck Integration

```bash
#!/bin/bash

# Install ShellCheck: brew install shellcheck

# Run ShellCheck on script
shellcheck script.sh

# ShellCheck directives to disable warnings
# shellcheck disable=SC2086  # Quote expansion
# shellcheck disable=SC1091  # Source not found

# Example with directive
function bad_function() {
    # shellcheck disable=SC2086
    echo $UNQUOTED_VAR
}
```

---

## POSIX Compliance

### Example 22: POSIX Shell Script

```sh
#!/bin/sh
# POSIX-compliant script (works with sh, bash, dash)

# Use [ ] instead of [[ ]]
if [ -f "file.txt" ]; then
    echo "File exists"
fi

# Use $() instead of backticks
current_date=$(date +%Y-%m-%d)

# Use expr for arithmetic
result=$(expr 5 + 3)

# Portable array (POSIX doesn't support arrays)
set -- "apple" "banana" "cherry"
echo "First: $1"
echo "All: $*"
```

---

## Performance Tips

### Example 23: Optimization Techniques

```bash
#!/bin/bash

# Use local variables in functions
fast_function() {
    local count=0
    for i in {1..1000}; do
        ((count++))
    done
    echo "$count"
}

# Avoid unnecessary subshells
slow_way() {
    cat file.txt | while read -r line; do
        echo "$line"
    done
}

fast_way() {
    while read -r line; do
        echo "$line"
    done < file.txt
}

# Use built-in commands instead of external tools
slow_approach() {
    total=$(cat numbers.txt | wc -l)
}

fast_approach() {
    total=0
    while read -r line; do
        ((total++))
    done < numbers.txt
}
```

---

_For more detailed reference and troubleshooting, see reference.md_
