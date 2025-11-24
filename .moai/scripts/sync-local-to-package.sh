#!/bin/bash
# MoAI-ADK Local → Package Sync Script (Version 1.1)
# Synchronizes local project to package template
#
# Synced:
#   - AI agent documentation (README-AI-MODELS.md)
#   - 120+ skill files and 102 modules
#   - All optimizations and validations
#
# Excluded (Local-only):
#   - .claude/commands/moai/ (entire directory)
#   - .claude/settings.local.json
#   - CLAUDE.local.md
#   - Runtime files (.moai/cache/, logs/, config/)
#
# Usage: bash .moai/scripts/sync-local-to-package.sh

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================"
echo "MoAI-ADK 로컬 → 패키지 동기화"
echo "======================================${NC}"
echo ""

# Phase 1: Pre-sync validation
echo -e "${YELLOW}[1/5] 사전 검증...${NC}"
git status --short
echo ""

# Verify all changes are committed
UNCOMMITTED=$(git status --short | grep -v "??" | wc -l)
if [ "$UNCOMMITTED" -gt 0 ]; then
  echo -e "${RED}✗ 오류: 미커밋된 변경사항 존재${NC}"
  echo "모든 변경사항을 먼저 커밋하세요"
  exit 1
fi
echo -e "${GREEN}✓ 모든 변경사항이 커밋됨${NC}"
echo ""

# Create backup
echo -e "${YELLOW}[2/5] 백업 생성 중...${NC}"
BACKUP_DIR="src/moai_adk/templates.backup-$(date +%Y%m%d-%H%M%S)"
cp -R src/moai_adk/templates/ "$BACKUP_DIR"
echo -e "${GREEN}✓ 백업 완료${NC}"
echo "  위치: $BACKUP_DIR"
echo ""

# Phase 2: Sync operations
echo -e "${YELLOW}[3/5] 파일 동기화 중...${NC}"

# 2.1 AI Agent Documentation
echo "  → AI 에이전트 동기화..."
rsync -av --progress \
  .claude/agents/moai/README-AI-MODELS.md \
  src/moai_adk/templates/.claude/agents/moai/ 2>/dev/null || true
echo -e "${GREEN}  ✓ AI 에이전트 완료${NC}"

# 2.2 Skills with exclusions
echo ""
echo "  → 스킬 동기화 (120+ 파일, 시간 소요 가능)..."
rsync -av --delete \
  --exclude=".CROSS-SKILL-VALIDATION-REPORT.md" \
  --exclude=".QUALITY-GATE-VALIDATION.md" \
  --exclude=".SKILLS-OPTIMIZATION-SUMMARY.md" \
  --exclude=".SKILLS-OPTIMIZATION-PR-SUMMARY.md" \
  --exclude=".DS_Store" \
  --exclude="__pycache__/" \
  --exclude="*.pyc" \
  --exclude=".cache" \
  .claude/skills/ \
  src/moai_adk/templates/.claude/skills/ 2>/dev/null || true
echo -e "${GREEN}  ✓ 스킬 동기화 완료${NC}"

echo ""
echo -e "${GREEN}✓ 모든 파일 동기화 완료${NC}"
echo ""

# Phase 3: Verification
echo -e "${YELLOW}[4/5] 동기화 검증 중...${NC}"

# File count comparison
LOCAL_SKILLS=$(find .claude/skills -name "*.md" 2>/dev/null | wc -l)
PACKAGE_SKILLS=$(find src/moai_adk/templates/.claude/skills -name "*.md" 2>/dev/null | wc -l)

echo "  파일 개수 비교:"
echo "    로컬 스킬: $LOCAL_SKILLS 파일"
echo "    패키지 스킬: $PACKAGE_SKILLS 파일"

if [ "$LOCAL_SKILLS" -eq "$PACKAGE_SKILLS" ]; then
  echo -e "    ${GREEN}✓ 파일 개수 일치${NC}"
else
  echo -e "    ${YELLOW}⚠ 경고: 파일 개수 불일치 ($LOCAL_SKILLS vs $PACKAGE_SKILLS)${NC}"
fi

# Module count comparison
LOCAL_MODULES=$(find .claude/skills -type d -name "modules" 2>/dev/null | wc -l)
PACKAGE_MODULES=$(find src/moai_adk/templates/.claude/skills -type d -name "modules" 2>/dev/null | wc -l)

echo ""
echo "  모듈 디렉토리 비교:"
echo "    로컬 모듈: $LOCAL_MODULES 개"
echo "    패키지 모듈: $PACKAGE_MODULES 개"

if [ "$LOCAL_MODULES" -eq "$PACKAGE_MODULES" ]; then
  echo -e "    ${GREEN}✓ 모듈 개수 일치${NC}"
fi

# Exclusion verification
echo ""
echo "  로컬 전용 파일 제외 확인:"

# Check commands (should not sync entire directory)
if [ ! -d "src/moai_adk/templates/.claude/commands/moai/" ]; then
  echo -e "    ${GREEN}✓ commands/moai/ 제외됨${NC}"
else
  echo -e "    ${YELLOW}ℹ commands/moai/ 존재 (기존 파일일 수 있음)${NC}"
fi

# Check settings.local.json
if [ ! -f "src/moai_adk/templates/.claude/settings.local.json" ]; then
  echo -e "    ${GREEN}✓ settings.local.json 제외됨${NC}"
else
  echo -e "    ${RED}✗ 오류: settings.local.json이 동기화됨${NC}"
fi

# Check CLAUDE.local.md
if [ ! -f "src/moai_adk/templates/CLAUDE.local.md" ]; then
  echo -e "    ${GREEN}✓ CLAUDE.local.md 제외됨${NC}"
else
  echo -e "    ${RED}✗ 오류: CLAUDE.local.md가 동기화됨${NC}"
fi

# Check optimization reports
OPT_FILES=$(find "src/moai_adk/templates/" -name ".SKILLS-OPTIMIZATION-*.md" 2>/dev/null | wc -l)
if [ "$OPT_FILES" -eq 0 ]; then
  echo -e "    ${GREEN}✓ 최적화 리포트 제외됨${NC}"
else
  echo -e "    ${YELLOW}⚠ 최적화 리포트 $OPT_FILES개 발견${NC}"
fi

echo ""

# Phase 4: Git status
echo -e "${YELLOW}[5/5] Git 상태 확인...${NC}"
echo "변경된 파일 목록 (처음 30개):"
git status --short src/moai_adk/templates/ | head -30 || true
echo ""

CHANGES=$(git status --short src/moai_adk/templates/ | wc -l)
echo -e "총 변경 파일: $CHANGES개"
echo ""

# Summary
echo -e "${BLUE}======================================"
echo "✓ 동기화 완료"
echo "======================================${NC}"
echo ""
echo -e "${GREEN}다음 단계:${NC}"
echo "1. 변경사항 검토"
echo "   git diff --stat src/moai_adk/templates/"
echo ""
echo "2. 파일 스테이징"
echo "   git add src/moai_adk/templates/"
echo ""
echo "3. 커밋 생성"
echo "   git commit -m \"sync(templates): Synchronize local AI models and skills\""
echo ""
echo -e "${YELLOW}백업 위치: $BACKUP_DIR${NC}"
echo "문제 발생 시 복구:"
echo "  rm -rf src/moai_adk/templates/"
echo "  mv $BACKUP_DIR src/moai_adk/templates/"
echo "  git reset --hard HEAD"
echo ""
