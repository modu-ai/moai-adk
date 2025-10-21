#!/bin/bash
# Phase 4 구현 검증 스크립트
# Commands → Sub-agents → Skills 통합 워크플로우

set -e

PROJECT_ROOT="/Users/goos/MoAI/MoAI-ADK"
TEMPLATE_DIR="$PROJECT_ROOT/src/moai_adk/templates/.claude-ko"

echo "🔍 Phase 4 구현 검증 시작..."
echo ""

# 1. Commands 파일 존재 확인
echo "✅ Commands 파일 검증..."
if [ -f "$TEMPLATE_DIR/commands/alfred/3-sync.md" ]; then
    echo "  ✓ /alfred:3-sync 존재"
else
    echo "  ✗ /alfred:3-sync 없음"
    exit 1
fi

# 2. Skills 힌트 존재 확인
echo ""
echo "✅ Skills 힌트 검증..."

if grep -q "자동 활성화 Skills 정보" "$TEMPLATE_DIR/commands/alfred/3-sync.md"; then
    echo "  ✓ Skills 힌트 섹션 존재"
else
    echo "  ✗ Skills 힌트 섹션 없음"
    exit 1
fi

if grep -q "doc-syncer 에이전트" "$TEMPLATE_DIR/commands/alfred/3-sync.md"; then
    echo "  ✓ doc-syncer Skills 매핑 존재"
else
    echo "  ✗ doc-syncer Skills 매핑 없음"
    exit 1
fi

if grep -q "tag-agent 에이전트" "$TEMPLATE_DIR/commands/alfred/3-sync.md"; then
    echo "  ✓ tag-agent Skills 매핑 존재"
else
    echo "  ✗ tag-agent Skills 매핑 없음"
    exit 1
fi

# 3. 독립 컨텍스트 설명 확인
echo ""
echo "✅ 독립 컨텍스트 설명 검증..."

if grep -q "Sub-agents의 독립 컨텍스트" "$TEMPLATE_DIR/commands/alfred/3-sync.md"; then
    echo "  ✓ 독립 컨텍스트 설명 존재"
else
    echo "  ✗ 독립 컨텍스트 설명 없음"
    exit 1
fi

# 4. Skills description 검증 (선택)
echo ""
echo "✅ Skills description 검증 (선택적)..."

SKILLS_WITHOUT_USE_WHEN=$(find "$TEMPLATE_DIR/skills" -name "SKILL.md" -exec grep -L "Use when" {} \; | wc -l)
echo "  ℹ️  'Use when' 패턴 없는 Skills: $SKILLS_WITHOUT_USE_WHEN개"

if [ $SKILLS_WITHOUT_USE_WHEN -gt 0 ]; then
    echo "  ⚠️  일부 Skills에 'Use when' 패턴 누락 (비필수)"
fi

# 5. YAML frontmatter 구문 검증
echo ""
echo "✅ YAML frontmatter 검증..."

INVALID_YAML=0
for skill in $(find "$TEMPLATE_DIR/skills" -name "SKILL.md"); do
    # YAML frontmatter 추출 (첫 번째 ---부터 두 번째 ---까지)
    YAML_CONTENT=$(awk '/^---$/{p++;next} p==1' "$skill")

    # Python으로 YAML 파싱 테스트
    if ! echo "$YAML_CONTENT" | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)" 2>/dev/null; then
        echo "  ✗ YAML 오류: $skill"
        INVALID_YAML=$((INVALID_YAML + 1))
    fi
done

if [ $INVALID_YAML -eq 0 ]; then
    echo "  ✓ 모든 Skills YAML 유효"
else
    echo "  ✗ $INVALID_YAML개 Skills YAML 오류"
    exit 1
fi

# 6. 통계 출력
echo ""
echo "📊 통계:"
echo "  - Commands: $(find "$TEMPLATE_DIR/commands" -name "*.md" | wc -l)개"
echo "  - Agents: $(find "$TEMPLATE_DIR/agents" -name "*.md" | wc -l)개"
echo "  - Skills: $(find "$TEMPLATE_DIR/skills" -name "SKILL.md" | wc -l)개"

echo ""
echo "✅ Phase 4 검증 완료!"
echo ""
echo "📋 다음 단계:"
echo "  1. /alfred:3-sync 실행하여 Skills 자동 활성화 확인"
echo "  2. doc-syncer 에이전트가 Skills를 사용하는지 확인"
echo "  3. 필요 시 Skills description에 'Use when' 패턴 추가"
