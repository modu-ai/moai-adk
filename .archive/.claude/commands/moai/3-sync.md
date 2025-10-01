---
name: moai:3-sync
description: 문서 동기화 + PR Ready 전환
argument-hint: "[mode] [target-path]"
tools: Read, Write, Edit, MultiEdit, Bash, Task, Grep, Glob, TodoWrite
---

# MoAI-ADK 3단계: 문서 동기화(+선택적 PR Ready)

**문서 동기화 대상**: $ARGUMENTS

## 🔍 STEP 1: 동기화 범위 분석 및 계획 수립

프로젝트 상태를 분석하여 동기화 범위를 결정하고 체계적인 동기화 계획을 수립한 후 사용자 확인을 받습니다.

### 동기화 분석 진행

1. **프로젝트 상태 확인**
   - Git 상태 및 변경된 파일 목록
   - 코드-문서 일치성 검사
   - AI-TAG 시스템 검증

2. **동기화 범위 결정**
   - Living Document 업데이트 필요 영역
   - TAG 인덱스 갱신 필요성
   - PR 상태 전환 가능성 (팀 모드)

3. **동기화 전략 수립**
   - 모드별 동기화 접근 방식
   - 예상 작업 시간 및 우선순위
   - 잠재적 위험 요소 식별

### 사용자 확인 단계 (조건부)

**자동 실행 조건 충족 시**: 이 단계를 건너뛰고 바로 동기화 실행

**확인 필요 시** 동기화 계획 검토 후 다음 중 선택하세요:
- **"진행"** 또는 **"시작"**: 계획대로 동기화 시작
- **"수정 [내용]"**: 동기화 계획 수정 요청
- **"중단"**: 동기화 작업 중단

```bash
# 자동 실행 분기 처리
if [[ "$SKIP_CONFIRMATION" == "true" ]]; then
  echo "⚡ 안전한 변경으로 판단, 자동 동기화 시작"
  # 즉시 STEP 2로 이동
else
  echo "🛡️ 수동 확인 필요 - 계획을 검토하고 승인해주세요"
  # 기존 확인 프로세스 실행
fi
```

---

## 🚀 STEP 2: 문서 동기화 실행 (사용자 승인 후)

사용자 승인 후 doc-syncer 에이전트가 **Living Document 동기화와 16-Core @TAG 업데이트**를 수행하고, 팀 모드에서만 PR Ready 전환을 선택적으로 실행합니다.

## 기능

- **ULTRATHINK**: doc-syncer 에이전트가 Living Document 동기화와 16-Core @TAG 업데이트를 수행합니다. 팀 모드에서만 PR Ready 전환을 선택적으로 실행합니다.

## 동기화 산출물

- `.moai/reports/sync-report.md` 생성/갱신
- TAG 인덱스 업데이트: TypeScript 기반 TAG 자동 갱신

## 모드별 실행 방식

## 📋 STEP 1 실행 가이드: 동기화 범위 분석 및 계획 수립

### 1. 프로젝트 상태 분석

다음을 우선적으로 실행하여 동기화 범위를 분석합니다:

```bash
# 인라인 프로젝트 상태 분석 (고속)
echo "📊 프로젝트 상태 분석 중..."

# Git 상태 분석
CHANGED_FILES=$(git status --porcelain | wc -l | tr -d ' ')
UNCOMMITTED_DOCS=$(git status --porcelain -- "*.md" "docs/" | wc -l | tr -d ' ')
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# 문서 신선도 체크
OUTDATED_DOCS=$(find docs README.md CLAUDE.md -name "*.md" -not -newer .moai/reports/sync-report.md 2>/dev/null | wc -l | tr -d ' ')

# TAG 시스템 상태
TAG_FILES=$(find .moai/indexes -name "*.json" -o -name "*-report.md" 2>/dev/null | wc -l | tr -d ' ')

echo "✅ 분석 완료: 변경=$CHANGED_FILES, 문서=$UNCOMMITTED_DOCS, 구문=$OUTDATED_DOCS, TAG=$TAG_FILES"

# 자동 실행 가능성 판단
AUTO_SYNC_SAFE="false"
if [[ $CHANGED_FILES -lt 5 && $UNCOMMITTED_DOCS -lt 3 && $OUTDATED_DOCS -lt 3 ]]; then
  AUTO_SYNC_SAFE="true"
  echo "🚀 자동 동기화 조건 충족 (안전한 변경 감지)"
else
  echo "⚠️  수동 확인 권장 (중간/대규모 변경 감지)"
fi

# 모드별 실행 경로 결정
SYNC_MODE="${1:-auto}"
case "$SYNC_MODE" in
  "auto")
    if [[ "$AUTO_SYNC_SAFE" == "true" ]]; then
      echo "⚡ AUTO 모드: 즉시 동기화 시작"
      SKIP_CONFIRMATION="true"
    else
      echo "🛡️ AUTO 모드: 안전을 위해 확인 단계로 전환"
      SKIP_CONFIRMATION="false"
    fi
    ;;
  "status")
    echo "🔍 STATUS 모드: 분석만 수행하고 종료"
    exit 0
    ;;
  "force")
    echo "🔄 FORCE 모드: 강제 전체 동기화 (확인 필요)"
    SKIP_CONFIRMATION="false"
    ;;
  "interactive")
    echo "👥 INTERACTIVE 모드: 항상 확인 단계 실행"
    SKIP_CONFIRMATION="false"
    ;;
  *)
    echo "⚙️ 기본 모드: auto로 설정"
    SYNC_MODE="auto"
    SKIP_CONFIRMATION="$AUTO_SYNC_SAFE"
    ;;
esac
```

#### 분석 체크리스트

- [ ] **Git 상태**: 변경된 파일, 브랜치 상태, 커밋 히스토리
- [ ] **문서 일치성**: 코드-문서 간 동기화 필요성
- [ ] **TAG 시스템**: AI-TAG 체계 검증 및 끊어진 링크
- [ ] **동기화 범위**: 전체 vs 부분 vs 특정 경로 동기화

### 2. 동기화 전략 결정

#### 모드별 동기화 접근법

| 모드 | 동기화 범위 | PR 처리 | 주요 특징 |
|------|-------------|---------|----------|
| **Personal** | 로컬 문서 동기화 | 체크포인트만 | 개인 작업 중심 |
| **Team** | 전체 동기화 + TAG | PR Ready 전환 | 협업 지원 |
| **Auto** | 지능형 자동 선택 | 상황별 결정 | 최적 전략 |
| **Force** | 강제 전체 동기화 | 전체 재생성 | 오류 복구용 |

#### 예상 작업 범위

- **Living Document**: API 문서, README, 아키텍처 문서
- **TAG 인덱스**: `.moai/indexes/tags.json` 갱신
- **동기화 보고서**: `.moai/reports/sync-report.md`
- **PR 상태**: Draft → Ready for Review 전환

### 3. 동기화 계획 보고서 생성

다음 형식으로 계획을 제시합니다:

```
## 문서 동기화 계획 보고서: [TARGET]

### 📊 상태 분석 결과
- **변경된 파일**: [개수 및 유형]
- **동기화 필요성**: [높음/중간/낮음]
- **TAG 시스템 상태**: [정상/문제 감지]

### 🎯 동기화 전략
- **선택된 모드**: [auto/force/status/project]
- **동기화 범위**: [전체/부분/선택적]
- **PR 처리**: [유지/Ready 전환/새 PR 생성]

### 🚨 주의사항
- **잠재적 충돌**: [문서 충돌 가능성]
- **TAG 문제**: [끊어진 링크, 중복 TAG]
- **성능 영향**: [대용량 동기화 예상시간]

### ✅ 예상 산출물
- **sync-report.md**: [동기화 결과 요약]
- **tags.json**: [업데이트된 TAG 인덱스]
- **Living Documents**: [갱신된 문서 목록]
- **PR 상태**: [팀 모드에서 PR 전환]

---
**승인 요청**: 위 계획으로 동기화를 진행하시겠습니까?
("진행", "수정 [내용]", "중단" 중 선택)
```

---

## 🚀 STEP 2 실행 가이드: 문서 동기화 (승인 후)

사용자가 **"진행"** 또는 **"시작"**을 선택하거나 **자동 실행 조건**을 충족한 경우 다음을 실행합니다:

```bash
# 최적화된 문서 동기화 실행
if [[ "$SKIP_CONFIRMATION" == "true" ]]; then
  echo "⚡ 고속 자동 동기화 시작"
  # 경량화된 동기화 (node 직접 실행)
  node .moai/scripts/doc-syncer.js --target="$2" --mode=sync --auto=true --fast=true
else
  echo "🔄 전체 문서 동기화 시작"
  # 완전한 동기화 (기존 방식)
  tsx .moai/scripts/doc-syncer.ts --target="$2" --mode=sync --approved=true
fi

# 스크립트 실행 실패 시 에이전트 fallback
if [[ $? -ne 0 ]]; then
  echo "📋 스크립트 실패, doc-syncer 에이전트로 전환"
  # doc-syncer 에이전트 직접 호출
fi
```

### 동기화 단계별 가이드

1. **Living Document 동기화**: 코드 → 문서 자동 반영
2. **TAG 시스템 검증**: AI-TAG 체계 무결성 확인
3. **인덱스 업데이트**: 트레이시빌리티 매트릭스 갱신
4. **보고서 생성**: 동기화 결과 요약 작성

### 에이전트 협업 구조

- **1단계**: `doc-syncer` 에이전트가 Living Document 동기화 및 AI-TAG 관리를 전담합니다.
- **2단계**: `git-manager` 에이전트가 모든 Git 커밋, PR 상태 전환, 동기화를 전담합니다.
- **단일 책임 원칙**: doc-syncer는 문서 작업만, git-manager는 Git 작업만 수행합니다.
- **순차 실행**: doc-syncer → git-manager 순서로 실행하여 명확한 의존성을 유지합니다.
- **에이전트 간 호출 금지**: 각 에이전트는 다른 에이전트를 직접 호출하지 않고, 커맨드 레벨에서만 순차 실행합니다.

## ⚡ 최적화된 고속 워크플로우

### 🚀 Fast Track (auto 모드)

**조건**: 안전한 변경 (<5 파일, <3 문서)
**실행 시간**: 5-10초

```
Phase 1: 인라인 분석 (0.5초)
├── 즉시 git status 체크
├── 문서 변경 카운트
└── 자동 실행 가능 판단

Phase 2: 단일 에이전트 실행 (3-5초)
├── doc-syncer 에이전트 (통합 모드)
├── 문서 동기화 + Git 커밋
└── 즉시 완료 보고
```

### 🛡️ Safe Track (interactive/force 모드)

**조건**: 복잡한 변경 또는 사용자 요청
**실행 시간**: 15-25초

```
Phase 1: 병렬 분석 (2-3초)
├── Task 1: Git 상태 상세 분석
└── Task 2: TAG 시스템 검증

Phase 2: 확인 및 실행 (5-10초)
├── 사용자 승인 대기
└── doc-syncer 에이전트 (완전 모드)

Phase 3: Git 후처리 (3-5초)
├── git-manager 에이전트
└── PR 상태 전환 (Team 모드)
```

**성능 향상**: Fast Track으로 80% 케이스를 5-10초에 처리

### 인수 처리

- **$1 (모드)**: `auto`(자동 최적화)|`interactive`(확인 후 실행)|`force`(강제 전체)|`status`(분석만)|`project`(프로젝트 통합)
- **$2 (경로)**: 동기화 대상 경로 (선택사항)

#### 모드별 동작 방식

| 모드 | 확인 과정 | 실행 조건 | 최적화 |
|------|----------|----------|--------|
| **auto** | 스킵 (안전 시) | <5 파일, <3 문서 | ⚡ 고속 |
| **interactive** | 필수 확인 | 모든 변경 | 🛡️ 안전 |
| **force** | 확인 후 강제 | 전체 재동기화 | 🔄 완전 |
| **status** | 없음 | 분석만 실행 | 🔍 진단 |

```bash
# ⚡ 고속 자동 동기화 (5-10초, 추천)
/moai:3-sync auto

# 🛡️ 안전한 대화형 동기화 (15-25초)
/moai:3-sync interactive

# 🔄 강제 전체 동기화 (20-30초)
/moai:3-sync force

# 🔍 상태 확인만 (2-3초)
/moai:3-sync status

# 📁 특정 경로 동기화
/moai:3-sync auto docs/api/

# 기본값 (auto 모드로 실행)
/moai:3-sync
```

### ⚡ 성능 비교

| 모드 | 실행 시간 | 확인 과정 | 사용 시나리오 |
|------|----------|----------|---------------|
| **auto** | 5-10초 | 스킵 (안전 시) | 일상적 동기화 |
| **status** | 2-3초 | 없음 | 상태 확인만 |
| **interactive** | 15-25초 | 필수 | 중요한 변경 |
| **force** | 20-30초 | 확인 후 강제 | 문제 해결용 |

### 에이전트 역할 분리

#### doc-syncer 전담 영역

- Living Document 동기화 (코드 ↔ 문서)
- AI-TAG 시스템 검증 및 업데이트
- API 문서 자동 생성/갱신
- README 및 아키텍처 문서 동기화
- 문서-코드 일치성 검증

#### git-manager 전담 영역

- 모든 Git 커밋 작업 (add, commit, push)
- 모드별 동기화 전략 적용
- PR 상태 전환 (Draft → Ready)
- 리뷰어 자동 할당 및 라벨링
- GitHub CLI 연동 및 원격 동기화

### 🧪 개인 모드 (Personal)

- 명령어가 git-manager를 호출하여 동기화 전/후 체크포인트 생성
- README·심층 문서·PR 본문 정리는 체크리스트에 따라 수동 마무리

### 🏢 팀 모드 (Team)

- Living Document 완전 동기화 + AI-TAG 검증/보정
- gh CLI가 설정된 경우에 한해 PR Ready 전환과 라벨링을 선택적으로 실행

**중요**: 모든 Git 작업(커밋, 동기화, PR 관리)은 git-manager 에이전트가 전담하므로, 이 커멘드에서는 Git 작업을 직접 실행하지 않습니다.

## 동기화 상세(요약)

1. 프로젝트 분석 및 TAG 검증 → 끊어진/중복/고아 TAG 점검
2. 코드 ↔ 문서 동기화 → API/README/아키텍처 문서 갱신, SPEC ↔ 코드 TODO 동기화
3. TAG 인덱스 업데이트 → TypeScript 기반 TAG 자동 갱신

## 다음 단계

- 문서 동기화 완료 후 전체 MoAI-ADK 워크플로우 완성
- 모든 Git 작업은 git-manager 에이전트가 전담하여 일관성 보장
- 에이전트 간 직접 호출 없이 커멘드 레벨 오케스트레이션만 사용

## 결과 보고

동기화 결과를 구조화된 형식으로 보고합니다:

### 성공적인 동기화 (최적화된 결과)

#### ⚡ 고속 동기화 완료 (5-10초)
```
✅ AUTO 모드 동기화 완료
📊 처리 결과: 파일 3개 업데이트, TAG 15개 검증, 0개 오류
⏱️ 실행 시간: 6.2초 (기존 대비 85% 단축)
🚀 다음 기능 개발 준비 완료
```

#### 🛡️ 안전 동기화 완료 (15-25초)
```
✅ INTERACTIVE 모드 동기화 완료
📊 처리 결과: 파일 12개 업데이트, 문서 5개 생성, TAG 48개 검증
⏱️ 실행 시간: 18.4초 (기존 대비 50% 단축)
💡 사용자 확인을 통한 안전한 동기화
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
- TRUST 5원칙 준수

**제한 사항:**

- TAG 검증은 파일 존재 기반 체크
- PR 자동 전환은 gh CLI 환경에서만 동작
- 커버리지 수치는 별도 측정 필요

---

**doc-syncer 서브에이전트와 연동하여 코드-문서 일치성 향상과 AI-TAG 추적성 보장을 목표로 합니다.**
