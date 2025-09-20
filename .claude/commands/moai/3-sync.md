---
name: moai:3-sync
description: MoAI-ADK SYNC 단계 - Living Document 동기화, 16-Core TAG 시스템 업데이트, PR Ready 전환. TDD 구현 완료 후 사용.
argument-hint: [auto|force|status|project] [target-path]
allowed-tools: Read, Write, Edit, MultiEdit, Bash(git:*), Bash(gh:*), Bash(python3:*), Bash(ls:*), Bash(find:*), Bash(grep:*), Bash(cat:*), Bash(pgrep:*), Bash(rm:*), Bash(sleep:*), Task, Grep, Glob, TodoWrite
model: sonnet
---

# MoAI-ADK SYNC 단계: 문서 동기화 + PR Ready

**doc-syncer** 서브에이전트를 활용하여 TDD 구현 완료 후 Living Document 동기화, 16-Core TAG 시스템 업데이트, Draft→Ready PR 전환을 수행합니다.

## 모드별 실행 방식

### 인수 처리
- **$1 (모드)**: `$1` → `auto`(기본값)|`force`|`status`|`project`
- **$2 (경로)**: `$2` → 동기화 대상 경로 (선택사항)

```bash
# 기본 자동 동기화
/moai:3-sync

# 전체 강제 동기화
/moai:3-sync force

# 동기화 상태 확인
/moai:3-sync status

# 통합 프로젝트 동기화
/moai:3-sync project

# 특정 경로 동기화
/moai:3-sync auto src/auth/
```

## 환경 정보 수집

현재 프로젝트 상태와 Git 환경을 확인합니다:

**Git 상태 확인**
!`git branch --show-current`
!`git status --porcelain`
!`git log --oneline -3`

**SPEC-ID 추출**
!`git branch --show-current | sed 's/feature/\(SPEC-[0-9]*\).*/\1/' || echo "SPEC-UNKNOWN"`

**프로젝트 유형 감지**
!`find . -maxdepth 2 -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.java" -o -name "*.kt" -o -name "*.cs" -o -name "*.swift" -o -name "*.dart" -o -name "*.rs" | head -8`
!`find . -maxdepth 1 -name "package.json" -o -name "pyproject.toml" -o -name "go.mod" -o -name "pom.xml" -o -name "Cargo.toml" -o -name "*.csproj" -o -name "*.sln" -o -name "Package.swift" -o -name "pubspec.yaml" | head -5`

## doc-syncer 서브에이전트 호출

doc-syncer 서브에이전트를 활용하여 체계적인 문서 동기화를 수행합니다:

### Phase 1: 프로젝트 분석 및 TAG 검증

doc-syncer 에이전트로 현재 프로젝트 상태를 분석하고 16-Core TAG 시스템의 무결성을 검증해주세요.

**분석 요청사항:**
- 모드: $1 (기본값: auto)
- 대상 경로: $2
- 현재 SPEC-ID 기반 문서 연결 확인
- 끊어진 TAG 링크 감지
- 프로젝트 유형별 문서 요구사항 분석

### Phase 2: Living Document 동기화

코드 변경사항을 기반으로 문서를 자동 동기화합니다:

**코드 → 문서 동기화:**
- API 문서 자동 갱신 (Web API 프로젝트)
- CLI 명령어 문서 업데이트 (CLI Tool 프로젝트)
- 컴포넌트 문서 갱신 (Frontend 프로젝트)
- 라이브러리 API 레퍼런스 업데이트 (Library 프로젝트)

**문서 → 코드 동기화:**
- SPEC 변경사항 코드 반영
- TODO 항목 코드 주석 동기화
- TAG 추적성 링크 업데이트

### Phase 3: 16-Core TAG 시스템 업데이트

TAG 시스템의 완전성을 보장합니다:

**Primary Chain 검증:**
- @REQ → @DESIGN → @TASK → @TEST 연결 확인
- @FEATURE → @API → @UI → @DATA 추적성 검증

**Quality Chain 관리:**
- @PERF → @SEC → @DOCS → @TAG 품질 체인 점검
- 고아 TAG 정리 및 중복 TAG 해결

## Git 작업 안전성 보장

Git 프로세스 충돌 방지 및 안전한 커밋을 수행합니다:

**Git 안전성 검사**
!`pgrep -fl "git (commit|rebase|merge)" >/dev/null 2>&1 && echo "CONFLICT" || echo "SAFE"`

**Lock 파일 정리**
!`[ -f .git/index.lock ] && rm -f .git/index.lock && echo "Lock removed" || echo "No lock"`

**변경사항 스테이징 전략**
1. 문서 파일 우선 스테이징: `docs/`, `README.md`, `*.md`
2. MoAI 시스템 파일: `.moai/`, `.claude/`
3. 템플릿 파일: `src/moai_adk/resources/templates/`

## 조건부 커밋 실행

변경사항이 있을 때만 커밋을 수행합니다:

!`
# 스테이징된 변경사항 확인
if git diff --cached --quiet; then
    echo "ℹ️ 커밋할 변경사항이 없습니다."
else
    SPEC_ID=$(git branch --show-current | sed 's/feature\/\(SPEC-[0-9]*\).*/\1/' || echo "SPEC-UNKNOWN")
    git commit -m "📚 $SPEC_ID: 문서 동기화 및 16-Core @TAG 업데이트 완료

- Living Document 실시간 동기화
- 프로젝트 유형별 문서 자동 생성/업데이트
- README.md 기능 목록 갱신
- 16-Core @TAG 추적성 체인 검증/수정
- 코드-문서 일치성 향상

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
    echo "✅ 문서 동기화 커밋 완료"
fi
`

## GitHub 통합 (환경 의존적)

gh CLI가 사용 가능한 경우 PR 상태 전환과 리뷰어 할당을 수행합니다:

**PR Ready 전환**
```bash
gh pr ready --body "$(cat <<'EOF'
## ✅ Implementation Complete

### 📊 Quality Metrics
- Constitution 5원칙: 체크 완료
- Test Coverage: 목표 달성 확인
- Code Quality: 품질 검증 완료
- Security Scan: 보안 검토 권장

### 🔗 Traceability Chain
- @REQ → @DESIGN → @TASK → @TEST: 연결 확인
- 16-Core @TAG 추적 체인 검증 완료

### 📋 Review Checklist
- [ ] Code Review (Senior Developer)
- [ ] Security Review (Security Lead)
- [ ] QA Testing (QA Engineer)

Ready for team review! 🚀

🤖 Generated with [Claude Code](https://claude.ai/code)
EOF
)"
```

**리뷰어 할당 (선택사항)**
```bash
# 환경에 따라 적절한 리뷰어 할당
gh pr edit --add-label "ready-for-review" --add-label "constitution-compliant"
```

## 결과 보고

동기화 결과를 구조화된 형식으로 보고합니다:

### 성공적인 동기화
```
✅ Living Document 동기화 완료!

📊 처리 결과:
├── 업데이트된 파일: X개 (실제 수치 보고)
├── 생성된 문서: X개
├── 수정된 TAG: X개
└── 추적성 검증: 통과 (검증 결과 기반)

🏷️ TAG 시스템 상태:
├── Primary Chain: 연결 확인
├── Quality Chain: 연결 확인
├── 고아 TAG: X개 (실제 수치)
└── 끊어진 링크: X개 (수정 완료)
```

### 부분 동기화 (문제 감지)
```
⚠️ 부분 동기화 완료 (문제 발견)

🔴 해결 필요한 문제:
├── 끊어진 링크: X개 (구체적 목록)
├── 중복 TAG: X개
└── 고아 TAG: X개

🛠️ 자동 수정 권장사항:
1. 끊어진 링크 복구
2. 중복 TAG 병합
3. 고아 TAG 정리
```

## 다음 단계 안내


### 개발 사이클 완료
```
🔄 MoAI-ADK 3단계 워크플로우 완성:
✅ /moai:1-spec → EARS 명세 작성
✅ /moai:2-build → TDD 구현
✅ /moai:3-sync → 문서 동기화

🎉 다음 기능 개발 준비 완료
> /moai:1-spec "다음 기능 설명"
```

### 통합 프로젝트 모드
```
🏢 통합 브랜치 동기화 완료!

📋 전체 프로젝트 동기화:
├── README.md (전체 기능 목록)
├── docs/architecture.md (시스템 설계)
├── docs/api/ (통합 API 문서)
└── .moai/indexes/ (전체 TAG 인덱스)

🎯 PR 전환 지원 완료
```
## 제약사항 및 가정

**환경 의존성:**
- Git 저장소 필수
- gh CLI (GitHub 통합 시 필요)
- Python3 (TAG 검증 스크립트)

**전제 조건:**
- MoAI-ADK 프로젝트 구조 (.moai/, .claude/)
- TDD 구현 완료 상태
- Constitution 5원칙 준수

**제한 사항:**
- TAG 검증은 파일 존재 기반 체크
- PR 자동 전환은 gh CLI 환경에서만 동작
- 커버리지 수치는 별도 측정 필요

---

**doc-syncer 서브에이전트와 연동하여 코드-문서 일치성 향상과 16-Core TAG 추적성 보장을 목표로 합니다.**
