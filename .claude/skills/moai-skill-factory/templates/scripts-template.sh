#!/bin/bash
#
# Script Template: [Script Purpose]
# Description: [What this script does in one sentence]
# Usage: ./script-name.sh [ARGS]
#
# This is a template for creating production-ready Bash scripts
# with proper error handling, validation, and logging.
#

set -euo pipefail

# ============================================================================
# CONFIGURATION
# ============================================================================

# Script metadata
SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Color codes for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Configuration variables
readonly LOG_LEVEL="${LOG_LEVEL:-INFO}"  # DEBUG, INFO, WARN, ERROR
readonly MAX_RETRIES="${MAX_RETRIES:-3}"
readonly RETRY_DELAY="${RETRY_DELAY:-2}"

# ============================================================================
# LOGGING FUNCTIONS
# ============================================================================

log_debug() {
    [[ "$LOG_LEVEL" == "DEBUG" ]] && echo -e "${BLUE}[DEBUG]${NC} $*" >&2
    return 0
}

log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_success() {
    echo -e "${GREEN}✓${NC} $*"
}

# ============================================================================
# ERROR HANDLING
# ============================================================================

# Trap errors and clean up
trap 'on_error $? $LINENO' ERR

on_error() {
    local exit_code=$1
    local line_number=$2
    log_error "Script failed with exit code $exit_code on line $line_number"
    cleanup
    exit "$exit_code"
}

# Trap signals for graceful shutdown
trap 'on_interrupt' INT TERM

on_interrupt() {
    log_warn "Script interrupted"
    cleanup
    exit 130
}

cleanup() {
    log_debug "Running cleanup..."
    # Add cleanup operations here (e.g., remove temp files, close connections)
    # Example:
    # [[ -n "${TEMP_FILE:-}" ]] && rm -f "$TEMP_FILE"
}

# ============================================================================
# VALIDATION FUNCTIONS
# ============================================================================

validate_arg() {
    local arg_name=$1
    local arg_value=$2

    if [[ -z "$arg_value" ]]; then
        log_error "Argument '$arg_name' is required but was not provided"
        print_usage
        exit 1
    fi
}

validate_file() {
    local file_path=$1

    if [[ ! -f "$file_path" ]]; then
        log_error "File not found: $file_path"
        return 1
    fi

    if [[ ! -r "$file_path" ]]; then
        log_error "File is not readable: $file_path"
        return 1
    fi

    return 0
}

validate_directory() {
    local dir_path=$1

    if [[ ! -d "$dir_path" ]]; then
        log_error "Directory not found: $dir_path"
        return 1
    fi

    if [[ ! -w "$dir_path" ]]; then
        log_error "Directory is not writable: $dir_path"
        return 1
    fi

    return 0
}

validate_command() {
    local cmd=$1

    if ! command -v "$cmd" &> /dev/null; then
        log_error "Required command not found: $cmd"
        return 1
    fi

    return 0
}

# ============================================================================
# RETRY LOGIC
# ============================================================================

retry() {
    local max_attempts=$1
    shift
    local cmd=("$@")
    local attempt=1

    while [[ $attempt -le $max_attempts ]]; do
        log_debug "Attempt $attempt of $max_attempts: ${cmd[*]}"

        if "${cmd[@]}"; then
            return 0
        fi

        if [[ $attempt -lt $max_attempts ]]; then
            log_warn "Command failed, retrying in ${RETRY_DELAY}s..."
            sleep "$RETRY_DELAY"
        fi

        ((attempt++))
    done

    log_error "Command failed after $max_attempts attempts: ${cmd[*]}"
    return 1
}

# ============================================================================
# MAIN FUNCTIONS
# ============================================================================

print_usage() {
    cat << EOF
Usage: $SCRIPT_NAME [OPTIONS] [ARGUMENTS]

Description:
  [Describe what this script does]

Options:
  -h, --help              Show this help message
  -v, --verbose           Enable verbose (debug) output
  -d, --dry-run           Show what would be done without doing it

Arguments:
  ARG1                    [Description of argument 1]
  ARG2                    [Description of argument 2]

Examples:
  # Basic usage
  $SCRIPT_NAME arg1 arg2

  # With verbose output
  $SCRIPT_NAME -v arg1 arg2

  # Dry run to preview changes
  $SCRIPT_NAME --dry-run arg1 arg2

Exit Codes:
  0   Success
  1   General error
  2   Invalid arguments
  130 Interrupted by user

Notes:
  [Any additional notes about the script]

See Also:
  [Reference to related documentation or scripts]
EOF
}

parse_args() {
    local dry_run=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h | --help)
                print_usage
                exit 0
                ;;
            -v | --verbose)
                export LOG_LEVEL="DEBUG"
                log_debug "Debug mode enabled"
                shift
                ;;
            -d | --dry-run)
                dry_run=true
                log_info "Dry run mode enabled"
                shift
                ;;
            -*)
                log_error "Unknown option: $1"
                print_usage
                exit 2
                ;;
            *)
                # Positional arguments
                break
                ;;
        esac
    done

    # Store remaining arguments
    readonly ARGS=("$@")
    readonly DRY_RUN=$dry_run

    # Validate required arguments
    if [[ ${#ARGS[@]} -lt 1 ]]; then
        log_error "Missing required argument"
        print_usage
        exit 2
    fi
}

main() {
    log_info "Starting $SCRIPT_NAME"

    # Validate prerequisites
    validate_command "grep" || exit 1
    validate_command "sed" || exit 1

    # Extract arguments
    local arg1="${ARGS[0]}"
    local arg2="${ARGS[1]:-}"

    log_debug "arg1: $arg1"
    log_debug "arg2: $arg2"

    # Validate arguments
    validate_arg "arg1" "$arg1"

    # Main logic here
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would process: $arg1"
        return 0
    fi

    # Example: Do something with retries
    retry "$MAX_RETRIES" \
        grep -r "pattern" "$arg1"

    # Success
    log_success "Operation completed successfully"
    return 0
}

# ============================================================================
# ENTRY POINT
# ============================================================================

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    parse_args "$@"
    main
fi
