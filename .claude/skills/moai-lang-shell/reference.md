# moai-lang-shell - CLI Reference

_Last updated: 2025-11-13_

## Quick Reference

### Installation

#### macOS
```bash
# Install Bash 5.2 (default is 3.2)
brew install bash

# Add to /etc/shells
echo /usr/local/bin/bash | sudo tee -a /etc/shells

# Change shell
chsh -s /usr/local/bin/bash
```

#### Linux
```bash
# Ubuntu/Debian
sudo apt-get install bash

# RHEL/CentOS
sudo yum install bash

# Verify version
bash --version
```

#### ShellCheck Installation
```bash
# macOS
brew install shellcheck

# Linux
sudo apt-get install shellcheck

# Verify
shellcheck --version
```

#### bats-core Installation
```bash
# macOS
brew install bats-core

# Linux
npm install -g bats

# Verify
bats --version
```

### Common Commands

#### Running Scripts

```bash
# Execute script
bash script.sh
./script.sh          # Requires chmod +x

# Run with debugging
bash -x script.sh    # Print commands before execution
bash -n script.sh    # Check syntax without execution

# Set options
bash -euo pipefail script.sh  # Exit on error, undefined vars, pipe failures
```

#### Testing

```bash
# Run bats tests
bats tests/test_script.bats

# Run all bats tests in directory
bats tests/

# Run with tap output
bats --tap tests/test_script.bats

# Run specific test
bats tests/test_script.bats -f "add function"
```

#### Linting and Formatting

```bash
# ShellCheck analysis
shellcheck script.sh
shellcheck -S style script.sh   # Show style issues
shellcheck -S error script.sh   # Show errors only

# ShellCheck exclusions
shellcheck -x script.sh         # Follow source files

# Format with shfmt (if installed)
shfmt -i 4 -w script.sh        # Format with 4-space indentation
shfmt -d script.sh             # Show diffs
```

#### Debugging

```bash
# Enable debugging
set -x
set -v

# Disable debugging
set +x
set +v

# Use PS4 for better output
PS4='+ ${BASH_SOURCE}:${LINENO}: ' bash -x script.sh

# Trace execution to file
exec 3>debug.log
bash -x script.sh 2>&3
```

---

## Tool Versions (2025-11-13)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Bash** | 5.2.37 | Core shell | Current |
| **ShellCheck** | 0.10.0 | Linting | Current |
| **bats-core** | 1.11.0 | Testing | Current |
| **bats-assert** | 2.1.0 | Assertions | Current |
| **shfmt** | 3.8.0 | Formatting | Optional |
| **jq** | 1.7.1 | JSON processing | Optional |

---

## Bash Version Compatibility (2025-11-13)

| Bash Version | Status | Key Features |
|---|---|---|
| **5.2+** | Current | Best features, performance |
| **5.1** | LTS | Mostly compatible |
| **5.0** | Maintenance | Most patterns work |
| **4.4** | Legacy | Avoid newer features |
| **3.2** | Deprecated | Never use (macOS default) |

---

## Syntax Reference

### Variables

```bash
# Assignment
var="value"
var=123
var=$(command)              # Command substitution
var=$((5 + 3))             # Arithmetic

# Quoting
"$var"                      # Expansion
'$var'                      # Literal
"${var}"                    # Explicit delimiter
```

### Parameter Expansion

```bash
${var}                      # Basic expansion
${var:-default}             # Default if unset
${var:=default}             # Assign default if unset
${var:+alternate}           # Use if set
${var:?error message}       # Error if unset
${var#pattern}              # Remove prefix (shortest)
${var##pattern}             # Remove prefix (longest)
${var%pattern}              # Remove suffix (shortest)
${var%%pattern}             # Remove suffix (longest)
${var/old/new}              # Replace first
${var//old/new}             # Replace all
${var:offset:length}        # Substring
${#var}                     # Length
${!var}                     # Indirect reference
${var@Q}                    # Quote for re-input
${var@E}                    # Expand escapes
```

### Arrays

```bash
# Indexed array
arr=(1 2 3)
arr[0]=10
${arr[0]}                   # Element
${arr[@]}                   # All elements
${arr[*]}                   # All as string
${#arr[@]}                  # Length
${!arr[@]}                  # Indices

# Associative array
declare -A dict
dict[key]=value
${dict[key]}
${!dict[@]}                 # Keys
${dict[@]}                  # Values
```

### Arithmetic

```bash
$((5 + 3))                  # Addition
$((10 - 2))                 # Subtraction
$((4 * 3))                  # Multiplication
$((20 / 4))                 # Division
$((17 % 5))                 # Modulo
$((2 ** 3))                 # Exponent
((count++))                 # Increment
((count--))                 # Decrement
((count += 5))              # Add assign
```

### Conditionals

```bash
# String comparisons
[[ $str = "value" ]]        # Equal
[[ $str != "value" ]]       # Not equal
[[ $str == pat* ]]          # Pattern match
[[ $str =~ regex ]]         # Regex match
[[ -z $str ]]               # Empty
[[ -n $str ]]               # Not empty
[[ $str < $str2 ]]          # Less than (lexical)

# Numeric comparisons
[[ $num -eq 5 ]]            # Equal
[[ $num -ne 5 ]]            # Not equal
[[ $num -lt 5 ]]            # Less than
[[ $num -le 5 ]]            # Less or equal
[[ $num -gt 5 ]]            # Greater than
[[ $num -ge 5 ]]            # Greater or equal

# File tests
[[ -e file ]]               # Exists
[[ -f file ]]               # Regular file
[[ -d dir ]]                # Directory
[[ -r file ]]               # Readable
[[ -w file ]]               # Writable
[[ -x file ]]               # Executable
[[ -s file ]]               # Non-empty
[[ -L file ]]               # Symlink

# Logical operators
[[ cond1 && cond2 ]]        # AND
[[ cond1 || cond2 ]]        # OR
[[ ! cond ]]                # NOT
```

### Control Flow

```bash
# If-else
if [[ condition ]]; then
    commands
elif [[ condition ]]; then
    commands
else
    commands
fi

# Case statement
case "$var" in
    pattern1)
        commands
        ;;
    pattern2|pattern3)
        commands
        ;;
    *)
        default
        ;;
esac

# For loop
for var in list; do
    commands
done

for ((i=0; i<10; i++)); do
    commands
done

# While loop
while [[ condition ]]; do
    commands
done

# Until loop
until [[ condition ]]; do
    commands
done

# Continue and break
continue [n]                # Skip to next iteration
break [n]                   # Exit loop
```

### Functions

```bash
# Define function
function_name() {
    commands
    return 0
}

# With explicit function keyword
function name() {
    commands
}

# Local variables
local var=value

# Return values
return 0                    # Success
return 1                    # Error

# Positional parameters
$0                          # Script name
$1, $2, ...                 # Arguments
$*                          # All arguments
$@                          # All arguments (quoted)
$#                          # Argument count
```

### Redirection

```bash
# Output redirection
command > file              # Write to file
command >> file             # Append to file
command 2> file             # Redirect stderr
command 2>&1                # Merge stderr to stdout
command &> file             # Both stdout and stderr

# Input redirection
command < file              # Read from file
command << EOF              # Here-document
command <<< "string"        # Here-string

# File descriptors
exec 3> file                # Open for writing
exec 3< file                # Open for reading
exec 3>&-                   # Close file descriptor
```

---

## Best Practices

### 1. Defensive Scripting

```bash
#!/bin/bash
set -euo pipefail           # Exit on error, undefined, pipe fail
IFS=$'\n\t'                 # Safer IFS

trap cleanup EXIT           # Cleanup on exit
trap 'error $LINENO' ERR    # Error handling
```

### 2. Quote Safety

```bash
# GOOD: Always quote variables
echo "$var"
for file in "${files[@]}"

# AVOID: Unquoted variables
echo $var                   # Vulnerable to word splitting
for file in $files          # May expand incorrectly
```

### 3. Function Guidelines

```bash
# GOOD: Use local variables
function my_function() {
    local var="value"
    local result=$(do_work)
    echo "$result"
}

# AVOID: Global variables
function bad_function() {
    global_var="value"
    global_result=$(do_work)
}
```

### 4. Error Handling

```bash
# GOOD: Check return status
if command; then
    echo "Success"
else
    echo "Failed: $?"
fi

# GOOD: Use || for alternatives
command || handle_error

# GOOD: Use && for dependent commands
cmd1 && cmd2 && cmd3
```

### 5. Performance

```bash
# GOOD: Use built-in commands
while read -r line; do
    echo "$line"
done < file.txt

# AVOID: External commands when not needed
cat file.txt | while read -r line
```

---

## Common Patterns

### Logging

```bash
log() {
    local level=$1
    shift
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [$level] $*" | tee -a app.log
}

log_info() { log "INFO" "$@"; }
log_error() { log "ERROR" "$@" >&2; }
```

### Safe Defaults

```bash
# Use readonly for constants
readonly CONFIG_FILE="/etc/app/config.txt"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Use set for safety
set -euo pipefail
```

### Cleanup

```bash
cleanup() {
    rm -f "$temp_file"
    kill $background_pid 2>/dev/null || true
}

trap cleanup EXIT
```

### Validation

```bash
validate_not_empty() {
    local var=$1
    if [[ -z "$var" ]]; then
        echo "Error: Variable cannot be empty" >&2
        return 1
    fi
}
```

---

## Troubleshooting

### Common Issues

**Issue**: `command not found` in script but works in terminal
- **Cause**: Script uses wrong shell or missing PATH
- **Solution**: Use `#!/bin/bash` and full paths: `${var:?error}`

**Issue**: `set -e` doesn't exit on failing command
- **Cause**: Command is part of pipe or conditional
- **Solution**: Use `set -o pipefail` and check conditions

**Issue**: Variables from pipe are lost
- **Cause**: Pipe creates subshell
- **Solution**: Use process substitution or `mapfile`

**Issue**: `[[ $var ]]` has unexpected behavior
- **Cause**: Mixing `[ ]` (POSIX) with `[[ ]]` (Bash)
- **Solution**: Use only `[[ ]]` in Bash, only `[ ]` in sh

**Issue**: ShellCheck warns about good code
- **Cause**: Legitimate use cases SC2086 warns against
- **Solution**: Use `# shellcheck disable=SC2086` (sparingly)

---

## Useful One-Liners

```bash
# Count lines in files
find . -name "*.sh" | xargs wc -l

# Replace text in multiple files
find . -name "*.sh" -exec sed -i 's/old/new/g' {} \;

# Create backup
for file in *.sh; do cp "$file" "$file.bak"; done

# Remove trailing whitespace
sed -i 's/[[:space:]]*$//' *.sh

# Extract IP addresses
grep -oE '\b([0-9]{1,3}\.){3}[0-9]{1,3}\b' file.txt

# Monitor file changes
watch -n 1 "cat log.txt | tail -20"

# Process each line
while IFS= read -r line; do process "$line"; done < input.txt

# Parallel execution
parallel --pipe --block 10M command < largefile.txt
```

---

## Resources

- **GNU Bash Manual**: https://www.gnu.org/software/bash/manual/
- **ShellCheck**: https://www.shellcheck.net
- **bats-core**: https://github.com/bats-core/bats-core
- **POSIX Shell**: https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html
- **Advanced Bash-Scripting Guide**: https://tldp.org/LDP/abs/html/

---

## Changelog

- ** .0** (2025-11-13): Reference documentation update
- **v3.0.0** (2025-03-15): POSIX compliance guide
- **v2.0.0** (2025-01-10): Testing and linting guide
- **v1.0.0** (2024-12-01): Initial release
