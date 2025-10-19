#!/bin/bash

echo "=== Skills 표준화 검증 ==="

# 1. 파일명 검증
skill_md_count=$(find .claude/skills/ -name "skill.md" 2>/dev/null | wc -l | tr -d ' ')
SKILL_md_count=$(find .claude/skills/ -name "SKILL.md" 2>/dev/null | wc -l | tr -d ' ')

echo "1. 파일명 표준화:"
echo "   - skill.md (비표준): $skill_md_count (0이어야 함)"
echo "   - SKILL.md (표준): $SKILL_md_count (46이어야 함)"

# 2. 중복 템플릿 검증
duplicate_count=$(ls .claude/skills/ 2>/dev/null | grep -c "moai-cc-.*-template" || echo 0)

echo "2. 중복 템플릿:"
echo "   - moai-cc-*-template: $duplicate_count (0이어야 함)"

# 3. YAML 필드 검증
version_count=$(rg "^version:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')
model_count=$(rg "^model:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')
allowed_tools_count=$(rg "^allowed-tools:" .claude/skills/*/SKILL.md 2>/dev/null | wc -l | tr -d ' ')

echo "3. YAML 필드:"
echo "   - version 필드: $version_count (0이어야 함)"
echo "   - model 필드: $model_count (0이어야 함)"
echo "   - allowed-tools 필드: $allowed_tools_count (46이어야 함)"

# 종합 판정
if [ "$skill_md_count" -eq 0 ] && \
   [ "$SKILL_md_count" -eq 46 ] && \
   [ "$duplicate_count" -eq 0 ] && \
   [ "$version_count" -eq 0 ] && \
   [ "$model_count" -eq 0 ] && \
   [ "$allowed_tools_count" -eq 46 ]; then
    echo ""
    echo "✅ 모든 검증 통과!"
    exit 0
else
    echo ""
    echo "❌ 검증 실패. 위 항목을 확인하세요."
    exit 1
fi
