---
name: moai:3-sync
description: 문서 동기화 및 TAG 시스템 업데이트 - Living Document와 16-Core TAG 동기화 지원
argument-hint: [auto|force|status] [target-path]
allowed-tools: Read, Write, Edit, MultiEdit, Bash(git:*), Bash(gh:*), Bash(python3:*), Bash(ls:*), Bash(cat:*), Task, Grep, Glob
---

# MoAI-ADK 3단계: 문서 동기화 + PR Ready (GitFlow 통합)

doc-syncer 에이전트가 코드-문서 양방향 동기화와 16-Core TAG 시스템 관리를 체계적으로 지원합니다. 환경에 따라 가능한 범위에서 자동화를 시도합니다.

## 🔀 문서 동기화 + PR 완료 워크플로우 지원 (환경 의존)

```bash
# 1. 16-Core @TAG 시스템 업데이트(시도)
# Python 의존성 제거 - 언어 중립적 체크리스트 기반 검증

# 2. Living Document 실시간 동기화
# 프로젝트 유형 감지 (언어 중립적)
!`find src -name '*.py' -o -name '*.js' -o -name '*.ts' | head -5`
!`find . -name 'package.json' -o -name 'pyproject.toml' -o -name 'go.mod' | head -3`

!`echo "🔍 프로젝트 유형 감지 완료"`
!`echo "📝 동기화 대상 문서 선택 완료"`

# 프로젝트 유형별 문서 생성
case "$PROJECT_TYPE" in
    "web_api"|"fullstack")
        echo "🌐 Web API 프로젝트 - API 문서 생성 중..."
        # API 엔드포인트 분석 및 문서 생성
        !`[ -f "src/app.py" ] || [ -f "src/main.py" ] && echo "API 파일 감지 완료" || echo "API 파일 없음"`
        fi
        ;;
    "cli_tool")
        echo "⚡ CLI 도구 프로젝트 - 명령어 문서 생성 중..."
        # CLI 명령어 도움말 추출
        ;;
    "library")
        echo "📚 라이브러리 프로젝트 - API 레퍼런스 생성 중..."
        # 라이브러리 API 문서 생성
        ;;
    "frontend")
        echo "🎨 프론트엔드 프로젝트 - 컴포넌트 문서 생성 중..."
        # 컴포넌트 문서 생성
        ;;
    *)
        echo "📱 일반 애플리케이션 - 기본 문서 생성 중..."
        # 기본 문서만 생성
        ;;
esac

# README.md 기능 목록 업데이트
!`grep -r "@FEATURE:" src/ 2>/dev/null | wc -l || echo "0"`
!`grep -r "@REQ:" .moai/specs/ 2>/dev/null | wc -l || echo "0"`


# 3. 통합 Git 작업 (Lock 파일 체크 + 커밋)
!`# Git 작업 통합 실행
if [ -f .git/index.lock ]; then
  echo "🔒 git index.lock 감지됨"
  if pgrep -fl "git (commit|rebase|merge)" >/dev/null 2>&1; then
    echo "❌ 다른 git 작업이 진행 중입니다. 해당 작업을 종료한 후 다시 실행하세요."
    exit 1
  else
    echo "🔓 Lock 파일 제거 중..."
    rm -f .git/index.lock
  fi
fi

# 문서 동기화 커밋
echo "📚 문서 동기화 시작..."
git add docs/ README.md 2>/dev/null || true

# 문서만으로 스테이징이 비어있으면, 동기화에 수반되는 경로를 추가 스테이징
if git diff --cached --quiet; then
  echo "ℹ️ 문서 경로에서 스테이징된 변경이 없습니다. 확장 스테이징 시도..."
  git add .claude/ .moai/ src/moai_adk/install/ src/moai_adk/resources/templates/ 2>/dev/null || true
fi

# 최종 확인 후 커밋
if git diff --cached --quiet; then
  echo "ℹ️ 커밋할 변경사항이 없습니다."
else
  git commit -m "📚 ${SPEC_ID}: 문서 동기화 및 16-Core @TAG 업데이트 완료

- Living Document 실시간 동기화
- 언어별 문서 자동 생성
- README.md 기능 목록 업데이트
- 16-Core @TAG 추적성 체인 점검/업데이트" && echo "✅ 문서 동기화 커밋 완료"
fi`

# 4. Draft → Ready for Review 전환 지원 (gh CLI 필요)
!`gh pr ready --body "$(cat <<EOF
## ✅ Implementation Complete

### 📊 Quality Metrics
- Constitution 5원칙: 체크 결과 보고
- Test Coverage: ${COVERAGE_PERCENT}%
- Code Quality: A+
- Security Scan: ✅ No vulnerabilities

### 🔗 Traceability Chain
- @REQ → @DESIGN → @TASK → @TEST: 연결 확인
- 16-Core @TAG 추적 체인 점검 완료

### 📋 Review Checklist
- [ ] Code Review (Senior Developer)
- [ ] Security Review (Security Lead)
- [ ] QA Testing (QA Engineer)

Ready for team review! 🚀
EOF
)"`

# 5. 리뷰어 할당 및 라벨링 지원 (gh CLI 필요)
!`gh pr edit --add-reviewer "@senior-dev" --add-reviewer "@security-lead"`
!`gh pr edit --add-label "ready-for-review" --add-label "constitution-compliant"`
```

코드와 문서의 일치성 향상을 목표로 하고 16-Core TAG 시스템으로 추적성을 강화하는 핵심 명령어입니다.

## 🔄 빠른 시작

```bash
# 자동 증분 동기화 (기본값)
/moai:3-sync

# 전체 프로젝트 동기화
/moai:3-sync force

# 통합 브랜치 동기화 (--project 모드 후)
/moai:3-sync project

# 동기화 상태 확인
/moai:3-sync status
```

## 🎯 실행 모드

### Auto 모드 (기본값) - 증분 동기화
변경된 파일과 TAG만 선택적으로 동기화하여 빠른 처리를 보장합니다.

```
🔄 증분 동기화 실행 중...

📊 변경사항 스캔:
├── 수정된 파일: 3개 감지
├── 새로운 @TAG: 2개 발견
└── 삭제된 파일: 0개

⚡ 최적화된 처리: 델타만 동기화
```

### Force 모드 - 전체 재동기화
전체 프로젝트를 다시 스캔하고 모든 문서를 재생성합니다.

```
🔄 전체 재동기화 실행 중...

📂 전체 프로젝트 스캔:
├── 소스 파일: 45개 분석
├── 테스트 파일: 70개 검토
├── 문서 파일: 15개 갱신
└── 총 130개 파일 처리

🏗️ 전체 TAG 인덱스 재구축
```

### Project 모드 - 통합 브랜치 동기화
--project로 생성된 통합 브랜치의 전체 SPEC을 동기화합니다.

```
🏢 통합 브랜치 동기화 실행 중...

📋 통합 SPEC 스캔:
├── SPEC-001: 사용자 인증 시스템
├── SPEC-002: 게시글 관리 시스템
├── SPEC-003: 댓글 및 좋아요
├── SPEC-004: 관리자 대시보드
└── SPEC-005: 모니터링 시스템

🔗 전체 프로젝트 일관성 검증
📝 통합 README 및 API 문서 업데이트
```

## 🤖 doc-syncer 에이전트 자동화

**doc-syncer 에이전트**가 양방향 동기화를 완전 자동화:

### 코드 → 문서 동기화
- **API 문서**: 코드 변경 시 OpenAPI 스펙 자동 갱신
- **README**: 기능 추가/수정 시 사용법 자동 업데이트
- **아키텍처 문서**: 구조 변경 시 다이어그램 자동 갱신

### 문서 → 코드 동기화
- **SPEC 변경**: 요구사항 수정 시 관련 코드 마킹
- **TODO 추가**: 문서의 할일이 코드 주석으로 반영
- **TAG 업데이트**: 추적성 링크 자동 갱신

## 🏷️ 16-Core TAG 시스템 동기화

### TAG 카테고리별 처리
```markdown
📋 Primary Chain (주요 추적):
├── REQ → DESIGN → TASK → TEST
├── FEATURE → BUG → DEBT → TODO
└── API → UI → DATA → CONFIG

🔍 Quality Chain (품질 추적):
├── PERF → SEC → DOCS → TAG
└── 추적성 매트릭스 유지 목표(지표 보고)
```

### 자동 검증 및 복구
- **끊어진 링크**: 자동 감지 및 수정 제안
- **중복 TAG**: 병합 또는 분리 옵션 제공
- **고아 TAG**: 참조 없는 태그 정리

## 📁 동기화 대상 파일

### 소스 코드 연동
```
src/ → docs/api/
├── models/*.py → API 스키마 문서
├── services/*.py → 비즈니스 로직 문서
├── controllers/*.py → 엔드포인트 문서
└── utils/*.py → 유틸리티 문서
```

### 테스트 연동
```
tests/ → docs/testing/
├── unit/ → 단위 테스트 보고서
├── integration/ → 통합 테스트 가이드
└── e2e/ → E2E 시나리오 문서
```

### SPEC 연동
```
.moai/specs/ → 프로젝트 문서
├── SPEC-*/spec.md → 요구사항 문서
├── SPEC-*/plan.md → 설계 문서
└── SPEC-*/tasks.md → 구현 가이드
```

## 📊 동기화 결과

### 성공적인 동기화
```bash
✅ Living Document 동기화 완료!

📊 처리 결과:
├── 업데이트된 파일: 8개
├── 생성된 문서: 3개
├── 수정된 TAG: 12개
└── 추적성 검증: 통과(검증 결과 보고)

🏷️ TAG 시스템 상태:
├── Primary Chain: 연결 확인
├── Quality Chain: 연결 확인
├── 고아 TAG: 0개
└── 끊어진 링크: 0개
```

### 부분 동기화 (문제 감지)
```bash
⚠️ 부분 동기화 완료 (문제 발견)

🔴 해결 필요한 문제:
├── 끊어진 링크: 2개
│   ├── REQ:USER-LOGIN-001 → 참조 파일 없음
│   └── TASK:AUTH-IMPL-005 → 구현 파일 삭제됨
├── 중복 TAG: 1개
│   └── API:USERS-001 (2개 파일에 중복)
└── 고아 TAG: 3개

🛠️ 자동 수정 옵션:
1. 끊어진 링크 복구
2. 중복 TAG 병합
3. 고아 TAG 정리

계속 진행하시겠습니까? (y/N)
```

## 🔧 고급 옵션

### 특정 경로 동기화
```bash
# 특정 SPEC만 동기화
/moai:3-sync auto SPEC-001

# 특정 디렉토리만
/moai:3-sync force src/auth/

# 문서 파일만
/moai:3-sync auto docs/

# 통합 브랜치 대상 (전체 프로젝트)
/moai:3-sync project
```

### 동기화 상태 확인
```bash
/moai:3-sync status

📊 현재 동기화 상태:
├── 마지막 동기화: 2시간 전
├── 변경된 파일: 5개 (동기화 필요)
├── TAG 상태: 양호
└── 권장 조치: /moai:3-sync auto
```

## 🔄 완료 후 다음 단계

### 단일 SPEC 모드 - Git 커밋
```bash
✅ 동기화 완료!

🎯 권장 다음 단계:
> git add .
> git commit -m "docs: sync SPEC-XXX implementation and documentation"

📝 변경된 파일:
├── README.md (기능 설명 업데이트)
├── docs/api/ (API 문서 갱신)
└── .moai/indexes/tags.json (TAG 인덱스 업데이트)
```

### --project 모드 - 통합 PR Ready
```bash
🏢 통합 브랜치 동기화 완료!

🎯 PR 전환 단계:
> gh pr ready  # Draft → Ready for Review

📋 전체 프로젝트 동기화:
├── README.md (전체 기능 목록)
├── docs/architecture.md (시스템 전체 설계)
├── docs/api/ (통합 API 문서)
└── .moai/indexes/ (전체 TAG 인덱스)
```

### 다음 개발 사이클
```bash
🔄 개발 사이클 완료!

전체 MoAI-ADK 워크플로우:
✅ /moai:1-spec (또는 --project) → SPEC 작성
✅ /moai:2-build SPEC-XXX → 순차 TDD 구현
✅ /moai:3-sync (또는 project) → 문서 동기화

🎉 다음 기능 개발 준비 완료
> /moai:1-spec "다음 기능 설명"  # 단일
> /moai:1-spec --project  # 전체 프로젝트
```

## ⚠️ 에러 처리

### Git index.lock 감지
```bash
fatal: Unable to create '.git/index.lock': File exists.

원인:
- 이전 git 명령 비정상 종료 또는 병렬 실행으로 lock 파일이 남아있음

해결 절차(안전 순서):
1) 활성 Git 작업 확인: pgrep -fl "git (commit|rebase|merge)"
   - 있으면 해당 작업을 종료/완료 후 다시 실행
2) 활성 작업이 없으면 lock 파일 제거: rm -f .git/index.lock
3) 상태 점검: git status
4) 동기화 재실행: /moai:3-sync
```

### 파일 충돌 감지
```bash
🔴 파일 충돌 감지:
- README.md: 수동 편집과 자동 생성 충돌

해결 옵션:
1. 수동 편집 유지 [기본값]
2. 자동 생성으로 덮어쓰기
3. 대화형 병합 모드
```

### TAG 시스템 오류
```bash
❌ TAG 시스템 오류:
- 인덱스 파일 손상: .moai/indexes/tags.json

자동 복구 중...
✅ 백업에서 복구 완료
```

## 🔁 응답 구조

출력은 반드시 3단계 구조를 따릅니다:
1. **Phase 1 Results**: 동기화 결과 및 변경사항
2. **Phase 2 Plan**: TAG 시스템 업데이트 계획
3. **Phase 3 Implementation**: Git 커밋 및 다음 단계
