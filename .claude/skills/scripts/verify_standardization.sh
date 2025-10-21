#!/bin/bash

echo "=== Skills standardization validation ==="

# 1. File name validation
skill_md_count=$(find .claude/skills/ -name "skill.md" 2>/dev/null | wc -l | tr -d ' ')
SKILL_md_count=$(find .claude/skills/ -name "SKILL.md" 2>/dev/null | wc -l | tr -d ' ')

echo "1. File name standardization:"
echo " - skill.md (non-standard): $skill_md_count (must be 0)"
echo " - SKILL.md (standard): $SKILL_md_count (should be 46)"

# 2. Duplicate template validation
duplicate_count=$(ls .claude/skills/ 2>/dev/null | grep -c "moai-cc-.*-template" || echo 0)

echo "2. Duplicate template:"
echo " - moai-cc-*-template: $duplicate_count (must be 0)"

# 3. YAML field validation
version_count=$(rg "^version:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')
model_count=$(rg "^model:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')
allowed_tools_count=$(rg "^allowed-tools:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')

echo "3. YAML field:"
echo " - version field: $version_count (must be 0)"
echo " - model field: $model_count (must be 0)"
echo " - allowed-tools field: $allowed_tools_count (should be 46)"

# Comprehensive judgment
if [ "$skill_md_count" -eq 0 ] && \
   [ "$SKILL_md_count" -eq 46 ] && \
   [ "$duplicate_count" -eq 0 ] && \
   [ "$version_count" -eq 0 ] && \
   [ "$model_count" -eq 0 ] && \
   [ "$allowed_tools_count" -eq 46 ]; then
    echo ""
    echo "✅ All validation passed!"
    exit 0
else
    echo ""
    echo "❌ validation failed. Check the items above."
    exit 1
fi
