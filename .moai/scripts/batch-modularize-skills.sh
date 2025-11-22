#!/bin/bash

# Phase 4 Week 1 스킬 배치 모듈화 스크립트
# 10개 HIGH PRIORITY 스킬을 자동으로 모듈화

set -e

SKILLS_DIR=".claude/skills"
TEMPLATES_DIR="src/moai_adk/templates/.claude/skills"

# Week 1 HIGH PRIORITY 스킬 목록
SKILLS=(
  "moai-lang-ruby"
  "moai-lang-php"
  "moai-lang-scala"
  "moai-lang-cpp"
  "moai-lang-kotlin"
  "moai-lang-html-css"
  "moai-lang-rust"
  "moai-domain-frontend"
  "moai-domain-figma"
  "moai-domain-monitoring"
)

echo "🚀 Phase 4 Week 1 스킬 배치 모듈화 시작"
echo "========================================"
echo ""

for skill in "${SKILLS[@]}"; do
  echo "📍 처리 중: $skill"

  skill_path="$SKILLS_DIR/$skill"

  # modules 디렉토리 생성
  mkdir -p "$skill_path/modules"

  # 메인 디렉토리에서 템플릿 디렉토리로 동기화
  mkdir -p "$TEMPLATES_DIR/$skill"

  echo "  ✓ 디렉토리 생성 완료"
  echo ""
done

echo "✅ 배치 모듈화 기초 설정 완료"
echo ""
echo "📝 다음 단계:"
echo "  1. 각 스킬의 SKILL.md 축약 (≤400줄)"
echo "  2. examples.md 확충 (10+ 예제)"
echo "  3. reference.md 작성"
echo "  4. modules/advanced-patterns.md 생성"
echo "  5. modules/optimization.md 생성"
echo ""
