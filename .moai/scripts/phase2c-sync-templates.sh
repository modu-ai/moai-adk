#!/bin/bash
# Phase 2C: Synchronize all changes to templates directory

echo "PHASE 2C: SYNCHRONIZATION & VALIDATION"
echo "======================================"

MAIN_DIR=".claude/skills"
TEMPLATE_DIR="src/moai_adk/templates/.claude/skills"

echo "Syncing from main to templates..."

# Count files before sync
BEFORE=$(find "$TEMPLATE_DIR" -name "*.md" | wc -l)

# Sync all skill directories
rsync -av --delete "$MAIN_DIR/" "$TEMPLATE_DIR/"

# Count files after sync
AFTER=$(find "$TEMPLATE_DIR" -name "*.md" | wc -l)

echo ""
echo "Sync Statistics:"
echo "  Files before: $BEFORE"
echo "  Files after: $AFTER"
echo "  Change: $(($AFTER - $BEFORE))"

# Verify key skills synchronized
echo ""
echo "Verification (sample checks):"

check_skill() {
    skill_name=$1
    if [ -f "$TEMPLATE_DIR/$skill_name/SKILL.md" ]; then
        main_lines=$(wc -l < "$MAIN_DIR/$skill_name/SKILL.md")
        template_lines=$(wc -l < "$TEMPLATE_DIR/$skill_name/SKILL.md")
        if [ "$main_lines" -eq "$template_lines" ]; then
            echo "  ✓ $skill_name: $template_lines lines (matched)"
        else
            echo "  ✗ $skill_name: MISMATCH (main=$main_lines, template=$template_lines)"
        fi
    else
        echo "  ✗ $skill_name: NOT FOUND"
    fi
}

# Check representative skills from each tier
check_skill "moai-essentials-debug"
check_skill "moai-cc-configuration"
check_skill "moai-domain-toon"
check_skill "moai-context7-integration"
check_skill "moai-foundation-specs"

echo ""
echo "======================================"
echo "✓ PHASE 2C COMPLETE: Synchronization finished"
