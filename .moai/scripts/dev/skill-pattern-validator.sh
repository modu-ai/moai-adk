#!/bin/bash

# MoAI-ADK 스킬 호출 패턴 검증 스크립트
# 사용법: ./skill-pattern-validator.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CLAUDE_DIR="$PROJECT_ROOT/../.claude"

echo "🔍 MoAI-ADK 스킬 호출 패턴 검증"
echo "프로젝트: $PROJECT_ROOT"
echo "검사 대상: $CLAUDE_DIR"
echo ""

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 카운터 초기화
ERRORS=0
WARNINGS=0

# 검증 함수
check_skill_invocation() {
    local pattern="$1"
    local description="$2"
    local severity="$3"  # "error" or "warning"

    echo -e "${BLUE}검사: $description${NC}"

    local results=$(grep -r "$pattern" "$CLAUDE_DIR" --include="*.md" 2>/dev/null || true)

    if [ -n "$results" ]; then
        local count=$(echo "$results" | wc -l)
        echo -e "  ${YELLOW}발견 ($count건):${NC}"
        echo "$results" | head -5 | while read -r line; do
            local file=$(echo "$line" | cut -d: -f1)
            local content=$(echo "$line" | cut -d: -f2-)
            echo "    - $file: $content"
        done

        if [ "$count" -gt 5 ]; then
            echo "    - ... 그 외 $((count - 5))건 더 발견"
        fi

        if [ "$severity" = "error" ]; then
            ((ERRORS++))
            echo -e "    ${RED}❌ 오류: 반드시 수정해야 합니다${NC}"
        else
            ((WARNINGS++))
            echo -e "    ${YELLOW}⚠️  경고: 개선 권장${NC}"
        fi
    else
        echo -e "  ${GREEN}✅ 통과${NC}"
    fi
    echo ""
}

# 검증 항목 실행
echo "=== 🚨 기본 검증 ==="
check_skill_invocation "skill-name" "플레이스홀더 실제 사용 (설명용 제외)" "warning"

echo "=== ⚠️  일관성 검증 ==="
check_skill_invocation "AskUserQuestion tool" "비표준 AskUserQuestion 참조" "warning"
check_skill_invocation "AskUserQuestion tool (documented in" "긴 AskUserQuestion 설명" "warning"

# 스킬 이름 정합성 검증
echo "=== 🔤 스킬 이름 정합성 검증 ==="
echo "검사: 등록된 스킬 이름 목록"
KNOWN_SKILLS=$(find "$CLAUDE_DIR/skills" -maxdepth 1 -type d -name "moai-*" -exec basename {} \; 2>/dev/null | sort)

echo "  발견된 스킬 ($(echo "$KNOWN_SKILLS" | wc -l | tr -d ' ')개):"
echo "$KNOWN_SKILLS" | head -10
if [ $(echo "$KNOWN_SKILLS" | wc -l | tr -d ' ') -gt 10 ]; then
    echo "    ... 그 외 $(($(echo "$KNOWN_SKILLS" | wc -l | tr -d ' ') - 10))개 더 발견"
fi
echo ""

# 스킬 호출 현황
echo "검사: 주요 스킬 호출 현황"
COMMON_SKILLS="moai-alfred-ask-user-questions moai-foundation-tags moai-foundation-trust moai-cc-agents moai-cc-commands"

for skill in $COMMON_SKILLS; do
    count=$(grep -r "Skill(\"$skill\")" "$CLAUDE_DIR" --include="*.md" 2>/dev/null | wc -l | tr -d ' ')
    if [ "$count" -gt 0 ]; then
        echo "  - $skill: $count회 호출"
    fi
done
echo ""

# 요약 보고
echo "=== 📊 검증 결과 요약 ==="
echo -e "오류: ${RED}$ERRORS${NC}건 (반드시 수정 필요)"
echo -e "경고: ${YELLOW}$WARNINGS${NC}건 (개선 권장)"

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}🎉 모든 스킬 호출 패턴이 표준을 준수합니다!${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠️  일부 패턴 개선이 권장됩니다.${NC}"
    exit 0
else
    echo -e "${RED}❌ 오류가 발견되었습니다. 즉시 수정하세요.${NC}"
    exit 1
fi