#!/bin/bash
# ==============================================================================
# backup-before-delete.sh
# ==============================================================================
# Safety script to automatically backup critical local-only files before deletion
#
# Usage:
#   .moai/scripts/backup-before-delete.sh [--dry-run] [file1] [file2] ...
#
# Protected directories (never delete):
#   - .moai/yoda/              (Lecture material generation)
#   - .moai/release/           (/moai:release command tool)
#   - .moai/docs/              (Local documentation)
#   - .claude/commands/moai/   (Local tools)
#
# =============================================================================

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'  # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BACKUP_ROOT="$PROJECT_ROOT/.moai-backups"
PROTECTED_PATHS=(
    ".moai/yoda"
    ".moai/release"
    ".moai/docs"
    ".claude/commands/moai"
)

DRY_RUN=${1:-}
shift || true

# ==============================================================================
# Functions
# ==============================================================================

log_info() {
    echo -e "${BLUE}ℹ️  $*${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $*${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $*${NC}"
}

log_error() {
    echo -e "${RED}❌ $*${NC}"
}

is_protected() {
    local path="$1"
    for protected in "${PROTECTED_PATHS[@]}"; do
        if [[ "$path" == "$protected"* ]]; then
            return 0
        fi
    done
    return 1
}

create_backup() {
    local source_path="$1"
    local backup_timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_name="${source_path//\//_}"
    local backup_dir="$BACKUP_ROOT/delete-backup_$backup_timestamp/$source_path"

    if [[ ! -e "$PROJECT_ROOT/$source_path" ]]; then
        log_warning "Source path does not exist: $source_path"
        return 1
    fi

    mkdir -p "$(dirname "$backup_dir")"

    if [[ -d "$PROJECT_ROOT/$source_path" ]]; then
        cp -r "$PROJECT_ROOT/$source_path" "$backup_dir"
    else
        cp "$PROJECT_ROOT/$source_path" "$backup_dir"
    fi

    echo "$backup_dir"
}

# ==============================================================================
# Main Logic
# ==============================================================================

echo -e "\n${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}🔒 Safe Delete Backup System${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}\n"

if [[ "$DRY_RUN" == "--dry-run" ]]; then
    log_info "DRY RUN MODE: No actual backups will be created"
fi

# If no files specified, show protected paths
if [[ $# -eq 0 ]]; then
    log_info "Protected directories (never delete):"
    for path in "${PROTECTED_PATHS[@]}"; do
        if [[ -e "$PROJECT_ROOT/$path" ]]; then
            log_success "  $path"
        else
            log_warning "  $path (not found)"
        fi
    done
    echo ""
    log_info "Usage: $0 [--dry-run] [file1] [file2] ..."
    echo ""
    exit 0
fi

# Process each file/directory
backed_up_count=0
protected_count=0
error_count=0

for target_path in "$@"; do
    # Normalize path (remove ./ prefix)
    target_path="${target_path#./}"

    echo -e "\n${BLUE}Processing: $target_path${NC}"

    # Check if protected
    if is_protected "$target_path"; then
        log_error "PROTECTED PATH - Cannot delete or backup protected file!"
        log_error "This file is critical to MoAI-ADK and must be preserved:"
        log_error "  $target_path"
        ((protected_count++))
        continue
    fi

    # Check if exists
    if [[ ! -e "$PROJECT_ROOT/$target_path" ]]; then
        log_warning "Path does not exist: $target_path"
        continue
    fi

    # Create backup
    if [[ "$DRY_RUN" == "--dry-run" ]]; then
        log_info "Would backup to: $BACKUP_ROOT/delete-backup_*/..."
        ((backed_up_count++))
    else
        backup_result=$(create_backup "$target_path" || echo "ERROR")

        if [[ "$backup_result" == "ERROR" ]]; then
            log_error "Failed to backup: $target_path"
            ((error_count++))
        else
            log_success "Backed up to: $backup_result"
            ((backed_up_count++))
        fi
    fi
done

# Summary
echo -e "\n${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}📊 Summary${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════════${NC}"
echo "  Backed up:      $backed_up_count file(s)"
echo "  Protected:      $protected_count file(s) - NOT DELETED"
echo "  Errors:         $error_count file(s)"

if [[ $protected_count -gt 0 ]]; then
    echo -e "\n${RED}⚠️  PROTECTED FILES DETECTED - These files are critical and must never be deleted:${NC}"
    for path in "${PROTECTED_PATHS[@]}"; do
        if [[ -e "$PROJECT_ROOT/$path" ]]; then
            echo -e "${RED}    - $path${NC}"
        fi
    done
fi

if [[ $error_count -gt 0 ]]; then
    exit 1
fi

echo -e "\n${GREEN}✅ Safe delete operation completed${NC}\n"
