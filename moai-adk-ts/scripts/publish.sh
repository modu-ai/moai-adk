#!/usr/bin/env bash
# @CODE:PUBLISH-001 | SPEC: npm 배포 자동화 스크립트
# MoAI-ADK v0.2.10 배포 스크립트

set -e  # 오류 발생 시 즉시 중단

echo "🚀 MoAI-ADK v0.2.10 배포 시작..."

# 1. 현재 디렉토리 확인
if [ ! -f "package.json" ]; then
  echo "❌ package.json 파일을 찾을 수 없습니다."
  echo "   moai-adk-ts 디렉토리에서 실행하세요."
  exit 1
fi

# 2. 버전 확인
CURRENT_VERSION=$(node -p "require('./package.json').version")
echo "📦 현재 버전: v${CURRENT_VERSION}"

if [ "$CURRENT_VERSION" != "0.2.10" ]; then
  echo "❌ package.json 버전이 0.2.10이 아닙니다: ${CURRENT_VERSION}"
  exit 1
fi

# 3. Git 상태 확인
if [ -n "$(git status --porcelain)" ]; then
  echo "⚠️  Uncommitted changes detected:"
  git status --short
  read -p "계속 진행하시겠습니까? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "배포 취소됨."
    exit 1
  fi
fi

# 4. 의존성 설치
echo ""
echo "📥 의존성 설치 중..."
bun install

# 5. TypeScript 타입 체크
echo ""
echo "🔍 TypeScript 타입 체크 중..."
bun run type-check

# 6. 린트 체크
echo ""
echo "🧹 Lint 검사 중..."
bun run check:biome

# 7. 테스트 실행
echo ""
echo "🧪 테스트 실행 중..."
bun run test

# 8. 빌드
echo ""
echo "🔨 빌드 중..."
bun run build

# 9. 빌드 결과 검증
if [ ! -f "dist/index.js" ] || [ ! -f "dist/index.cjs" ]; then
  echo "❌ 빌드 실패: dist 파일이 생성되지 않았습니다."
  exit 1
fi

if [ ! -f "templates/.claude/hooks/alfred/session-notice.cjs" ]; then
  echo "❌ Hook 빌드 실패: session-notice.cjs가 생성되지 않았습니다."
  exit 1
fi

echo "✅ 빌드 성공"

# 10. 배포 확인
echo ""
echo "📦 배포 준비 완료"
echo "   버전: v${CURRENT_VERSION}"
echo "   파일: dist/index.js, dist/index.cjs, templates/"
echo ""
read -p "NPM에 배포하시겠습니까? (y/N): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "배포 취소됨."
  exit 0
fi

# 11. NPM 배포
echo ""
echo "📤 NPM 배포 중..."
npm publish --access public

if [ $? -eq 0 ]; then
  echo ""
  echo "✅ 배포 성공!"
  echo "   패키지: https://www.npmjs.com/package/moai-adk"
  echo "   버전: v${CURRENT_VERSION}"
  echo ""
  echo "🏷️  다음 단계: Git 태그 생성 및 GitHub Release"
  echo "   git tag -a v${CURRENT_VERSION} -m \"Release v${CURRENT_VERSION}\""
  echo "   git push origin v${CURRENT_VERSION}"
else
  echo "❌ 배포 실패"
  exit 1
fi
