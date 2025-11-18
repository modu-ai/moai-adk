#!/bin/bash

# MoAI-ADK Agent Sync Script
# Purpose: Synchronize local agents with template versions
# Version: 1.0.0
# Date: 2025-11-19

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
LOCAL_AGENTS_DIR=".claude/agents/moai"
TEMPLATE_AGENTS_DIR="src/moai_adk/templates/.claude/agents/moai"
BACKUP_DIR=".moai/backup/agents-sync-$(date +%Y-%m-%d-%H%M%S)"

# Initialize
PHASE_RESULTS=()
TOTAL_CHANGES=0

# Functions
print_header() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

backup_files() {
    print_header "Phase 0: Backing Up Original Files"

    if mkdir -p "$BACKUP_DIR"; then
        cp -r "$LOCAL_AGENTS_DIR"/* "$BACKUP_DIR/"
        print_success "Backup created: $BACKUP_DIR"
        echo "Restore with: cp -r $BACKUP_DIR/* $LOCAL_AGENTS_DIR/"
    else
        print_error "Failed to create backup directory"
        exit 1
    fi
}

phase_1_simple_changes() {
    print_header "Phase 1: Simple Skill Changes (13 files)"

    local files=(
        "accessibility-expert.md"
        "api-designer.md"
        "backend-expert.md"
        "component-designer.md"
        "devops-expert.md"
        "figma-expert.md"
        "frontend-expert.md"
        "migration-expert.md"
        "monitoring-expert.md"
        "performance-engineer.md"
        "ui-ux-expert.md"
    )

    local count=0
    for file in "${files[@]}"; do
        if [ -f "$LOCAL_AGENTS_DIR/$file" ]; then
            sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' "$LOCAL_AGENTS_DIR/$file"
            print_success "Updated: $file"
            ((count++))
        else
            print_warning "File not found: $file"
        fi
    done

    PHASE_RESULTS+=("Phase 1: Updated $count files")
    TOTAL_CHANGES=$((TOTAL_CHANGES + count * 2))
}

phase_2_complex_changes() {
    print_header "Phase 2: Complex Skill Changes (5 files)"

    # cc-manager.md
    if [ -f "$LOCAL_AGENTS_DIR/cc-manager.md" ]; then
        sed -i 's/moai-alfred-workflow/moai-core-workflow/g' "$LOCAL_AGENTS_DIR/cc-manager.md"
        sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' "$LOCAL_AGENTS_DIR/cc-manager.md"
        sed -i 's/moai-alfred-tag-scanning/moai-core-tag-scanning/g' "$LOCAL_AGENTS_DIR/cc-manager.md"
        print_success "Updated: cc-manager.md (3 patterns)"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 10))
    fi

    # debug-helper.md
    if [ -f "$LOCAL_AGENTS_DIR/debug-helper.md" ]; then
        sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' "$LOCAL_AGENTS_DIR/debug-helper.md"
        sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' "$LOCAL_AGENTS_DIR/debug-helper.md"
        sed -i 's/moai-alfred-tag-scanning/moai-core-tag-scanning/g' "$LOCAL_AGENTS_DIR/debug-helper.md"
        print_success "Updated: debug-helper.md (3 patterns)"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 8))
    fi

    # doc-syncer.md
    if [ -f "$LOCAL_AGENTS_DIR/doc-syncer.md" ]; then
        sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' "$LOCAL_AGENTS_DIR/doc-syncer.md"
        sed -i 's/moai-alfred-tag-scanning/moai-core-tag-scanning/g' "$LOCAL_AGENTS_DIR/doc-syncer.md"
        print_success "Updated: doc-syncer.md (2 patterns)"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 16))
    fi

    # git-manager.md
    if [ -f "$LOCAL_AGENTS_DIR/git-manager.md" ]; then
        sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' "$LOCAL_AGENTS_DIR/git-manager.md"
        sed -i 's/moai-alfred-git-workflow/moai-core-git-workflow/g' "$LOCAL_AGENTS_DIR/git-manager.md"
        sed -i 's/moai-alfred-trust-validation/moai-core-trust-validation/g' "$LOCAL_AGENTS_DIR/git-manager.md"
        print_success "Updated: git-manager.md (3 patterns)"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 12))
    fi

    # implementation-planner.md
    if [ -f "$LOCAL_AGENTS_DIR/implementation-planner.md" ]; then
        sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' "$LOCAL_AGENTS_DIR/implementation-planner.md"
        sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' "$LOCAL_AGENTS_DIR/implementation-planner.md"
        print_success "Updated: implementation-planner.md (2 patterns)"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 20))
    fi

    PHASE_RESULTS+=("Phase 2: Updated 5 complex files")
}

phase_3_large_scale_changes() {
    print_header "Phase 3: Large-Scale Namespace Updates (6 files)"

    # agent-factory.md
    if [ -f "$LOCAL_AGENTS_DIR/agent-factory.md" ]; then
        sed -i 's/moai-alfred-agent-factory/moai-core-agent-factory/g' "$LOCAL_AGENTS_DIR/agent-factory.md"
        print_success "Updated: agent-factory.md"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 12))
    fi

    # quality-gate.md
    if [ -f "$LOCAL_AGENTS_DIR/quality-gate.md" ]; then
        sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' "$LOCAL_AGENTS_DIR/quality-gate.md"
        sed -i 's/moai-alfred-trust-validation/moai-core-trust-validation/g' "$LOCAL_AGENTS_DIR/quality-gate.md"
        print_success "Updated: quality-gate.md"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 18))
    fi

    # skill-factory.md
    if [ -f "$LOCAL_AGENTS_DIR/skill-factory.md" ]; then
        sed -i 's/moai-alfred-skill-factory/moai-core-skill-factory/g' "$LOCAL_AGENTS_DIR/skill-factory.md"
        sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' "$LOCAL_AGENTS_DIR/skill-factory.md"
        print_success "Updated: skill-factory.md"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 30))
    fi

    # spec-builder.md
    if [ -f "$LOCAL_AGENTS_DIR/spec-builder.md" ]; then
        sed -i 's/moai-alfred-spec-authoring/moai-core-spec-authoring/g' "$LOCAL_AGENTS_DIR/spec-builder.md"
        sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' "$LOCAL_AGENTS_DIR/spec-builder.md"
        sed -i 's/moai-alfred-ears-authoring/moai-core-ears-authoring/g' "$LOCAL_AGENTS_DIR/spec-builder.md"
        print_success "Updated: spec-builder.md"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 18))
    fi

    # tdd-implementer.md
    if [ -f "$LOCAL_AGENTS_DIR/tdd-implementer.md" ]; then
        sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' "$LOCAL_AGENTS_DIR/tdd-implementer.md"
        sed -i 's/moai-alfred-language-detection/moai-core-language-detection/g' "$LOCAL_AGENTS_DIR/tdd-implementer.md"
        print_success "Updated: tdd-implementer.md"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 18))
    fi

    # trust-checker.md
    if [ -f "$LOCAL_AGENTS_DIR/trust-checker.md" ]; then
        sed -i 's/moai-alfred-ask-user-questions/moai-core-ask-user-questions/g' "$LOCAL_AGENTS_DIR/trust-checker.md"
        sed -i 's/moai-alfred-trust-validation/moai-core-trust-validation/g' "$LOCAL_AGENTS_DIR/trust-checker.md"
        print_success "Updated: trust-checker.md"
        TOTAL_CHANGES=$((TOTAL_CHANGES + 16))
    fi

    PHASE_RESULTS+=("Phase 3: Updated 6 large-scale files")
}

verify_synchronization() {
    print_header "Phase 4: Verification"

    # Check for remaining alfred references
    local alfred_count=$(grep -r "moai-alfred" "$LOCAL_AGENTS_DIR" 2>/dev/null | wc -l)

    if [ "$alfred_count" -eq 0 ]; then
        print_success "All moai-alfred references removed (0 remaining)"
    else
        print_warning "Found $alfred_count remaining moai-alfred references"
        echo "Details:"
        grep -r "moai-alfred" "$LOCAL_AGENTS_DIR" | head -10
    fi

    # Check file count
    local file_count=$(find "$LOCAL_AGENTS_DIR" -name "*.md" | wc -l)
    print_success "Total agent files: $file_count"

    # Check for core references
    local core_count=$(grep -r "moai-core" "$LOCAL_AGENTS_DIR" 2>/dev/null | wc -l)
    print_success "moai-core references found: $core_count"

    PHASE_RESULTS+=("Phase 4: Verification completed - $alfred_count remaining issues")
}

generate_summary() {
    print_header "Synchronization Summary"

    echo -e "${BLUE}Results:${NC}"
    for result in "${PHASE_RESULTS[@]}"; do
        echo "  - $result"
    done

    echo -e "\n${BLUE}Statistics:${NC}"
    echo "  Total changes applied: $TOTAL_CHANGES"
    echo "  Backup location: $BACKUP_DIR"
    echo "  Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"

    echo -e "\n${BLUE}Next steps:${NC}"
    echo "  1. Review changes: git diff .claude/agents/moai/"
    echo "  2. Verify functionality: /moai:0-project"
    echo "  3. Commit: git add .claude/agents/moai/ && git commit -m 'chore(agents): Sync with v0.26.0'"
    echo "  4. If issues: Restore backup with: cp -r $BACKUP_DIR/* $LOCAL_AGENTS_DIR/"
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║         MoAI-ADK Agent Synchronization Script              ║"
    echo "║                  Version 1.0.0                              ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    # Pre-flight checks
    if [ ! -d "$LOCAL_AGENTS_DIR" ]; then
        print_error "Local agents directory not found: $LOCAL_AGENTS_DIR"
        exit 1
    fi

    if [ ! -d "$TEMPLATE_AGENTS_DIR" ]; then
        print_warning "Template directory not found (optional): $TEMPLATE_AGENTS_DIR"
    fi

    print_info "Starting synchronization process..."

    # Execute phases
    backup_files
    phase_1_simple_changes
    phase_2_complex_changes
    phase_3_large_scale_changes
    verify_synchronization
    generate_summary

    echo -e "\n${GREEN}Synchronization complete!${NC}\n"
}

# Run main function
main
