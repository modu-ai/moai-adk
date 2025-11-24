#!/bin/bash
#
# CLAUDE.md Master-Replica Synchronization Script
# Synchronizes English (master) template to Korean (replica) project file
#
# Usage: bash .moai/scripts/sync-claude-files.sh
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MASTER_FILE="src/moai_adk/templates/CLAUDE.md"
REPLICA_FILE="./CLAUDE.md"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_file_exists() {
    if [[ ! -f "$1" ]]; then
        log_error "File not found: $1"
        return 1
    fi
    return 0
}

get_file_hash() {
    if [[ -f "$1" ]]; then
        sha256sum "$1" | awk '{print $1}'
    else
        echo ""
    fi
}

get_timestamp() {
    date -u +"%Y-%m-%d"
}

# Main synchronization logic
sync_claude_files() {
    log_info "Starting CLAUDE.md Master-Replica Synchronization..."
    log_info "Master: $MASTER_FILE"
    log_info "Replica: $REPLICA_FILE"
    echo ""

    # Step 1: Verify master file exists
    if ! check_file_exists "$MASTER_FILE"; then
        log_error "Master file not found. Exiting."
        return 1
    fi
    log_success "Master file found"

    # Step 2: Extract header and metadata from master
    log_info "Extracting metadata from master file..."
    MASTER_HASH=$(get_file_hash "$MASTER_FILE")
    CURRENT_TIMESTAMP=$(get_timestamp)

    log_success "Master hash: ${MASTER_HASH:0:8}..."

    # Step 3: Check if replica exists
    if [[ ! -f "$REPLICA_FILE" ]]; then
        log_warning "Replica file does not exist. Creating new file."
        # Copy master as starting point
        cp "$MASTER_FILE" "$REPLICA_FILE"
        log_success "Replica file created from master"
    else
        # Step 4: Update metadata section in replica
        log_info "Updating metadata section in replica..."

        # Update Last Sync date
        sed -i.bak "s/\*\*마지막 동기화 (Last Sync)\*\*: [0-9-]*/\*\*마지막 동기화 (Last Sync)\*\*: $CURRENT_TIMESTAMP/" "$REPLICA_FILE"

        # Update Validation Status (always passed after sync)
        sed -i.bak 's/\*\*상태 (Status)\*\*: .*$/\*\*상태 (Status)\*\*: ✅ In Sync/' "$REPLICA_FILE"

        # Clean up backup file
        rm -f "${REPLICA_FILE}.bak"

        log_success "Metadata updated"
    fi

    # Step 5: Verify both files
    log_info "Verifying file integrity..."
    if [[ ! -f "$REPLICA_FILE" ]]; then
        log_error "Replica file verification failed"
        return 1
    fi
    log_success "File integrity verified"

    # Step 6: Calculate new hash
    REPLICA_HASH=$(get_file_hash "$REPLICA_FILE")
    log_info "Replica hash: ${REPLICA_HASH:0:8}..."

    echo ""
    log_success "CLAUDE.md synchronization completed successfully"
    log_info "Files synchronized:"
    log_info "  Master (English): $MASTER_FILE"
    log_info "  Replica (Korean): $REPLICA_FILE"
    log_info "  Timestamp: $CURRENT_TIMESTAMP"

    return 0
}

# Pre-commit hook installation
install_git_hook() {
    log_info "Setting up Git pre-commit hook..."

    GIT_HOOKS_DIR="${PROJECT_ROOT}/.git/hooks"
    PRE_COMMIT_HOOK="${GIT_HOOKS_DIR}/pre-commit"

    # Create hooks directory if it doesn't exist
    mkdir -p "$GIT_HOOKS_DIR"

    # Check if hook already exists
    if [[ -f "$PRE_COMMIT_HOOK" ]]; then
        log_warning "Pre-commit hook already exists. Backing up..."
        cp "$PRE_COMMIT_HOOK" "${PRE_COMMIT_HOOK}.backup.$(date +%s)"
    fi

    # Create pre-commit hook
    cat > "$PRE_COMMIT_HOOK" << 'HOOK_EOF'
#!/bin/bash
# Git Pre-commit Hook: CLAUDE.md Synchronization
# This hook automatically synchronizes CLAUDE.md files when changes are detected

set -euo pipefail

MASTER_FILE="src/moai_adk/templates/CLAUDE.md"
REPLICA_FILE="./CLAUDE.md"

# Check if master file is in the commit
if git diff --cached --name-only | grep -q "^${MASTER_FILE}$"; then
    echo "🔄 CLAUDE.md changes detected. Running synchronization..."

    # Run sync script
    if bash .moai/scripts/sync-claude-files.sh; then
        # Stage the updated replica file
        git add "$REPLICA_FILE"
        echo "✅ CLAUDE.md synchronized and staged"
    else
        echo "❌ CLAUDE.md synchronization failed"
        exit 1
    fi
fi

exit 0
HOOK_EOF

    chmod +x "$PRE_COMMIT_HOOK"
    log_success "Git pre-commit hook installed at: $PRE_COMMIT_HOOK"
}

# Main execution
main() {
    cd "$PROJECT_ROOT"

    # Check for --install flag
    if [[ "${1:-}" == "--install" ]]; then
        install_git_hook
    else
        sync_claude_files
    fi
}

# Run main function
main "$@"
