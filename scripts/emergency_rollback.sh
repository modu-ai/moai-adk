#!/bin/bash
# SCRIPT-EMERGENCY-ROLLBACK-001: Emergency rollback script
"""Emergency rollback script for MoAI-ADK system

This script provides emergency recovery functionality for the MoAI-ADK system
in case of critical failures or issues. It can restore the system to a previous
stable state.

Usage:
    ./scripts/emergency_rollback.sh [--checkpoint <id>] [--force] [--dry-run]

Options:
    --checkpoint <id>    Rollback to specific checkpoint ID
    --force             Force rollback without confirmation
    --dry-run           Show what would be done without executing
    --list              List available checkpoints
    --emergency         Emergency rollback to oldest stable checkpoint
    --help              Show this help message
"""

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CHECKPOINTS_DIR="$PROJECT_ROOT/.moai/checkpoints"
LOG_FILE="$PROJECT_ROOT/.moai/emergency_rollback.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show help
show_help() {
    cat << EOF
TAG Policy System Emergency Rollback Script

This script provides emergency recovery functionality for the TAG policy system.

USAGE:
    $0 [OPTIONS]

OPTIONS:
    --checkpoint <id>    Rollback to specific checkpoint ID
    --force             Force rollback without confirmation
    --dry-run           Show what would be done without executing
    --list              List available checkpoints
    --emergency         Emergency rollback to oldest stable checkpoint
    --help              Show this help message

EXAMPLES:
    $0 --list                                    # List available checkpoints
    $0 --emergency                               # Emergency rollback to stable state
    $0 --checkpoint ckpt_20251105_143022_123456 # Rollback to specific checkpoint
    $0 --dry-run --emergency                     # Preview emergency rollback

For more information, see: SPEC-EMERGENCY-ROLLBACK-001
EOF
}

# Check if we're in a MoAI-ADK project
check_project() {
    if [[ ! -f "$PROJECT_ROOT/.moai/config.json" ]]; then
        print_error "Not a MoAI-ADK project: .moai/config.json not found"
        exit 1
    fi

    if [[ ! -d "$PROJECT_ROOT/src/moai_adk" ]]; then
        print_error "MoAI-ADK source directory not found"
        exit 1
    fi

    print_success "MoAI-ADK project confirmed"
}

# List available checkpoints
list_checkpoints() {
    print_status "Available checkpoints:"

    if [[ ! -d "$CHECKPOINTS_DIR" ]]; then
        print_warning "No checkpoints directory found"
        return 1
    fi

    local found=false
    for checkpoint_file in "$CHECKPOINTS_DIR"/checkpoint_*.json; do
        if [[ -f "$checkpoint_file" ]]; then
            found=true
            local checkpoint_id=$(basename "$checkpoint_file" .json | sed 's/^checkpoint_//')
            local timestamp=$(python3 -c "
import json
try:
    with open('$checkpoint_file', 'r') as f:
        data = json.load(f)
        print(data.get('timestamp', 'Unknown'))
except:
    print('Unknown')
" 2>/dev/null || echo "Unknown")

            echo "  - $checkpoint_id ($timestamp)"
        fi
    done

    if [[ "$found" == false ]]; then
        print_warning "No checkpoints found"
        return 1
    fi

    return 0
}

# Validate checkpoint exists
validate_checkpoint() {
    local checkpoint_id="$1"
    local checkpoint_file="$CHECKPOINTS_DIR/checkpoint_$checkpoint_id.json"
    local backup_dir="$CHECKPOINTS_DIR/backup_$checkpoint_id"

    if [[ ! -f "$checkpoint_file" ]]; then
        print_error "Checkpoint file not found: $checkpoint_file"
        return 1
    fi

    if [[ ! -d "$backup_dir" ]]; then
        print_error "Backup directory not found: $backup_dir"
        return 1
    fi

    return 0
}

# Create emergency backup before rollback
create_emergency_backup() {
    local backup_name="emergency_before_rollback_$(date +%Y%m%d_%H%M%S)"
    local backup_dir="$PROJECT_ROOT/.moai/emergency_backups/$backup_name"

    print_status "Creating emergency backup: $backup_name"

    mkdir -p "$backup_dir"

    # Backup critical files
    cp -r "$PROJECT_ROOT/.claude" "$backup_dir/" 2>/dev/null || true
    cp -r "$PROJECT_ROOT/.moai" "$backup_dir/" 2>/dev/null || true
    cp -r "$PROJECT_ROOT/src" "$backup_dir/" 2>/dev/null || true

    # Create backup metadata
    cat > "$backup_dir/backup_info.json" << EOF
{
    "backup_name": "$backup_name",
    "created_at": "$(date -Iseconds)",
    "reason": "Emergency backup before TAG policy rollback",
    "git_status": "$(git status --porcelain 2>/dev/null || echo 'Not a git repository')"
}
EOF

    print_success "Emergency backup created: $backup_dir"
}

# Rollback to specific checkpoint
rollback_to_checkpoint() {
    local checkpoint_id="$1"
    local checkpoint_file="$CHECKPOINTS_DIR/checkpoint_$checkpoint_id.json"
    local backup_dir="$CHECKPOINTS_DIR/backup_$checkpoint_id"

    print_status "Rolling back to checkpoint: $checkpoint_id"

    # Validate checkpoint
    if ! validate_checkpoint "$checkpoint_id"; then
        return 1
    fi

    # Show what will be restored
    print_status "Files to be restored from checkpoint $checkpoint_id:"
    find "$backup_dir" -type f -not -path "*/__pycache__/*" | head -10
    local total_files=$(find "$backup_dir" -type f -not -path "*/__pycache__/*" | wc -l)
    echo "  ... and $((total_files - 10)) more files"

    if [[ "${FORCE:-}" != "true" ]]; then
        echo
        read -p "Do you want to proceed with rollback? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Rollback cancelled"
            return 0
        fi
    fi

    # Create emergency backup
    create_emergency_backup

    # Restore files
    print_status "Restoring files from checkpoint..."
    local restored_files=0
    local failed_files=0

    while IFS= read -r -d '' backup_file; do
        # Calculate relative path
        local rel_path="${backup_file#$backup_dir/}"
        local target_file="$PROJECT_ROOT/$rel_path"

        # Create target directory if needed
        mkdir -p "$(dirname "$target_file")"

        # Restore file
        if cp "$backup_file" "$target_file"; then
            ((restored_files++))
        else
            ((failed_files++))
            print_error "Failed to restore: $rel_path"
        fi
    done < <(find "$backup_dir" -type f -not -path "*/__pycache__/*" -print0)

    print_success "Rollback completed: $restored_files files restored, $failed_files failed"

    # Log rollback
    log "ROLLBACK: checkpoint=$checkpoint_id, restored=$restored_files, failed=$failed_files"
}

# Emergency rollback to oldest stable checkpoint
emergency_rollback() {
    print_status "Performing emergency rollback to oldest stable checkpoint..."

    # Find oldest stable checkpoint
    local oldest_checkpoint=""
    local oldest_time=""

    for checkpoint_file in "$CHECKPOINTS_DIR"/checkpoint_*.json; do
        if [[ -f "$checkpoint_file" ]]; then
            local checkpoint_id=$(basename "$checkpoint_file" .json | sed 's/^checkpoint_//')
            local checkpoint_time=$(python3 -c "
import json
try:
    with open('$checkpoint_file', 'r') as f:
        data = json.load(f)
        print(data.get('timestamp', ''))
except:
    pass
" 2>/dev/null || echo "")

            if [[ -n "$checkpoint_time" ]]; then
                if [[ -z "$oldest_time" ]] || [[ "$checkpoint_time" < "$oldest_time" ]]; then
                    oldest_time="$checkpoint_time"
                    oldest_checkpoint="$checkpoint_id"
                fi
            fi
        fi
    done

    if [[ -z "$oldest_checkpoint" ]]; then
        print_error "No suitable checkpoint found for emergency rollback"
        return 1
    fi

    print_warning "Emergency rollback will restore to: $oldest_checkpoint ($oldest_time)"
    print_warning "This will undo all recent changes to the TAG policy system"

    if [[ "${FORCE:-}" != "true" ]]; then
        echo
        read -p "This is a destructive operation. Continue? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Emergency rollback cancelled"
            return 0
        fi
    fi

    rollback_to_checkpoint "$oldest_checkpoint"
}

# Validate TAG system after rollback
validate_system() {
    print_status "Validating TAG system after rollback..."

    # Run basic validation
    if python3 -c "
import sys
sys.path.insert(0, 'src')
try:
    from moai_adk.core.tags.validator import CentralValidator
    validator = CentralValidator()
    result = validator.validate_directory('.')

    if result.is_valid:
        print('✅ TAG system validation passed')
        print(f'  Files scanned: {result.statistics.total_files_scanned}')
        print(f'  Tags found: {result.statistics.total_tags_found}')
        print(f'  Coverage: {result.statistics.coverage_percentage:.1f}%')
    else:
        print('❌ TAG system validation failed')
        print(f'  Errors: {result.statistics.error_count}')
        print(f'  Warnings: {result.statistics.warning_count}')
        sys.exit(1)
except Exception as e:
    print(f'❌ Validation error: {e}')
    sys.exit(1)
" 2>/dev/null; then
        print_success "TAG system validation passed"
    else
        print_warning "TAG system validation failed - manual intervention may be required"
    fi
}

# Main execution
main() {
    local action=""
    local checkpoint_id=""
    local dry_run="false"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --checkpoint)
                checkpoint_id="$2"
                action="rollback_to_checkpoint"
                shift 2
                ;;
            --force)
                export FORCE="true"
                shift
                ;;
            --dry-run)
                dry_run="true"
                shift
                ;;
            --list)
                action="list"
                shift
                ;;
            --emergency)
                action="emergency"
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Start logging
    log "Emergency rollback script started: action=$action, checkpoint=$checkpoint_id"

    # Check project
    check_project

    # Execute action
    case "$action" in
        "list")
            list_checkpoints
            ;;
        "rollback_to_checkpoint")
            if [[ -z "$checkpoint_id" ]]; then
                print_error "Checkpoint ID required for rollback"
                exit 1
            fi
            if [[ "$dry_run" == "true" ]]; then
                print_status "DRY RUN: Would rollback to checkpoint $checkpoint_id"
                validate_checkpoint "$checkpoint_id"
            else
                rollback_to_checkpoint "$checkpoint_id"
                validate_system
            fi
            ;;
        "emergency")
            if [[ "$dry_run" == "true" ]]; then
                print_status "DRY RUN: Would perform emergency rollback"
                list_checkpoints
            else
                emergency_rollback
                validate_system
            fi
            ;;
        "")
            print_error "No action specified. Use --help for usage information."
            exit 1
            ;;
        *)
            print_error "Unknown action: $action"
            exit 1
            ;;
    esac

    log "Emergency rollback script completed"
}

# Execute main function with all arguments
main "$@"