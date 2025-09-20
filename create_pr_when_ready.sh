#!/bin/bash

echo "🚀 SPEC-003 PR 생성 스크립트"
echo "============================="

# 1. Git 상태 확인
echo "📊 Git 상태 확인..."
git status

# 2. 최종 커밋
echo "💾 최종 변경사항 커밋..."
git add .
git commit -m "🎯 SPEC-003: Package Optimization 최종 완료

✅ 패키지 크기 80% 감소 달성 (948KB → 192KB)
✅ 에이전트 파일 93% 감소 (60개 → 4개)
✅ 명령어 파일 77% 감소 (13개 → 3개)
✅ Constitution 5원칙 100% 준수
✅ TDD 완전 구현 (Red-Green-Refactor)

@REQ:OPT-CORE-001 @DESIGN:PKG-ARCH-001 @TASK:CLEANUP-IMPL-001 @TEST:UNIT-OPT-001"

# 3. 원격 푸시
echo "📤 원격 저장소에 푸시..."
git push origin feature/SPEC-003-package-optimization

# 4. PR 생성
echo "🔗 Pull Request 생성..."
gh pr create \
  --title "🚀 SPEC-003: Package Optimization System 구현 완료" \
  --body-file "/Users/goos/MoAI/MoAI-ADK/PR_TEMPLATE_SPEC-003.md" \
  --base main \
  --head feature/SPEC-003-package-optimization \
  --assignee @me \
  --label "enhancement,SPEC-003,optimization" \
  --milestone "v0.1.26"

echo "✅ PR 생성 완료!"
echo "🔗 GitHub에서 PR을 확인하고 리뷰를 요청하세요."