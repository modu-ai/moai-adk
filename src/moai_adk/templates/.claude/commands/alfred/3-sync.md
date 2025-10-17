---
name: alfred:3-sync
description: 문서 동기화 + PR Ready 전환
argument-hint: "모드 대상경로 - 모드: auto(기본)|force|status|project, 대상경로: 동기화 대상 경로"
allowed-tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash(git:*)
  - Bash(gh:*)
  - Bash(python3:*)
  - Task
  - Grep
  - Glob
  - TodoWrite
---

# 📚 MoAI-ADK 3단계: 문서 동기화(+선택적 PR Ready)

## 🎯 커맨드 목적

코드 변경사항을 Living Document에 동기화하고, @TAG 시스템을 검증하여 완벽한 추적성을 보장합니다.

**문서 동기화 대상**: $ARGUMENTS

> **표준 2단계 워크플로우** (자세한 내용: `CLAUDE.md` - "Alfred 커맨드 실행 패턴" 참조)

## 📋 실행 흐름

1. **프로젝트 상태 분석**: Git 변경사항 및 TAG 시스템 검증
2. **동기화 범위 결정**: 전체/부분/선택적 동기화 전략
3. **사용자 확인**: 동기화 계획 검토 및 승인
4. **문서 동기화**: Living Document 갱신 및 TAG 무결성 보장
5. **Git 작업**: git-manager를 통한 커밋 및 PR 상태 전환

## 🔗 연관 에이전트

- **Phase 1**: quality-gate (🛡️ 품질 보증 엔지니어) - 동기화 전 품질 검증 (조건부)
- **Primary**: doc-syncer (📖 테크니컬 라이터) - 문서 동기화 전담
- **Secondary**: git-manager (🚀 릴리스 엔지니어) - Git 커밋/PR 전담

## 💡 사용 예시

사용자가 다음과 같이 커맨드를 실행할 수 있습니다:
- `/alfred:3-sync` - 자동 동기화 (PR Ready만)
- `/alfred:3-sync --auto-merge` - PR 자동 머지 + 브랜치 정리
- `/alfred:3-sync force` - 강제 전체 동기화
- `/alfred:3-sync status` - 동기화 상태 확인
- `/alfred:3-sync project` - 통합 프로젝트 동기화

### 🚀 완전 자동화된 GitFlow (--auto-merge)

**Team 모드에서 사용 시 다음 작업을 자동으로 수행합니다**:
1. 문서 동기화 완료
2. PR Ready 전환
3. CI/CD 상태 확인
4. PR 자동 머지 (squash)
5. develop 체크아웃 및 동기화
6. 로컬 feature 브랜치 정리
7. **다음 작업 준비 완료** ✅

**권장 사용 시점**: TDD 구현 완료 후 한 번에 머지까지 완료하고 싶을 때

**Personal 모드**: 로컬 main/develop 머지 및 브랜치 정리 자동화

## 🔍 STEP 1: 동기화 범위 분석 및 계획 수립

프로젝트 상태를 분석하여 동기화 범위를 결정하고 체계적인 동기화 계획을 수립한 후 사용자 확인을 받습니다.

**doc-syncer 에이전트가 자동으로 TAG 체인 스캔 및 Git 변경사항을 확인하여 분석합니다.**

### 🔍 TAG 체인 탐색 (선택사항)

**TAG 체인이 복잡하거나 광범위한 경우** Explore 에이전트를 먼저 활용합니다:

```
Task tool 호출 (Explore 에이전트):
- subagent_type: "Explore"
- description: "TAG 시스템 전체 스캔"
- prompt: "프로젝트 전체에서 @TAG 시스템을 스캔해주세요:
          - @SPEC TAG 위치 (.moai/specs/)
          - @TEST TAG 위치 (tests/)
          - @CODE TAG 위치 (src/)
          - @DOC TAG 위치 (docs/)
          - 고아 TAG 및 끊어진 참조 탐지
          thoroughness 레벨: very thorough"
```

**Explore 에이전트 사용 시점**:
- ✅ 대규모 프로젝트 (100개 이상 파일)
- ✅ TAG 체인 무결성 검증이 필요한 경우
- ✅ 여러 SPEC에 걸친 변경사항
- ❌ 단일 SPEC의 간단한 변경

### ⚙️ 에이전트 호출 방법

**STEP 1에서는 Task tool을 사용하여 doc-syncer와 tag-agent를 호출합니다**:

```
1. tag-agent 호출 (TAG 검증):
   - subagent_type: "tag-agent"
   - description: "TAG 시스템 검증"
   - prompt: "전체 TAG 체인 무결성을 검증해주세요.
             @SPEC, @TEST, @CODE, @DOC TAG의 완전성과
             고아 TAG를 확인해주세요.
             (선택) Explore 결과: $EXPLORE_RESULTS"

2. doc-syncer 호출 (동기화 계획):
   - subagent_type: "doc-syncer"
   - description: "문서 동기화 계획 수립"
   - prompt: "Git 변경사항을 분석하여 문서 동기화 계획을 수립해주세요.
             $ARGUMENTS
             (선택) TAG 검증 결과: $TAG_VALIDATION_RESULTS"
```

### 동기화 분석 진행

1. **프로젝트 상태 확인**
   - Git 상태 및 변경된 파일 목록
   - 코드-문서 일치성 검사
   - @TAG 시스템 검증 (tag-agent 또는 Explore 활용)
   - (선택) Explore 결과 기반 광범위한 TAG 스캔

2. **동기화 범위 결정**
   - Living Document 업데이트 필요 영역
   - TAG 인덱스 갱신 필요성
   - PR 상태 전환 가능성 (팀 모드)

3. **동기화 전략 수립**
   - 모드별 동기화 접근 방식
   - 예상 작업 시간 및 우선순위
   - 잠재적 위험 요소 식별

### Phase 1 세부: 품질 사전 검증 (조건부 자동 실행)

동기화 전 코드 품질을 빠르게 확인합니다.

**Phase 3 (2-build)와의 차이점**:
- **Phase 3**: TDD 구현 완료 후 심층 검증 (테스트 커버리지, 코드 품질, 보안)
- **Phase 1**: 동기화 전 빠른 스캔 (파일 손상, Critical 이슈만)

**목적**: 품질 문제가 있는 코드의 문서화 방지

**실행 조건 (자동 판단)**:
- Git diff로 코드 변경 라인 수 확인
- 변경 라인 > 50줄: 자동 실행
- 변경 라인 ≤ 50줄: 건너뛰기
- 문서만 변경: 건너뛰기

**검증 항목**:
- **변경 파일만 검증**: Git diff로 확인된 파일 대상
- **TRUST 원칙 검증**: trust-checker 스크립트 실행
- **코드 스타일**: 린터 실행 (변경 파일만)
- **TAG 체인**: 변경된 TAG 무결성 확인

**실행 방식**:
Alfred가 코드 변경이 많을 때 자동으로 quality-gate 에이전트를 호출하여 문서 동기화 전 빠른 품질 검증을 수행합니다.

**검증 결과 처리**:

✅ **PASS (Critical 0개)**: 동기화 진행

⚠️ **WARNING (Critical 0개, Warning 있음)**: 경고 표시 후 동기화 진행

❌ **CRITICAL (Critical 1개 이상)**: 동기화 중단, 수정 권장
- Critical 이슈 발견: 동기화 중단, 수정 권장
- 사용자 선택: "수정 후 재시도" 또는 "강제 진행"

**검증 생략 옵션**:
사전 검증을 건너뛰려면 `/alfred:3-sync --skip-pre-check` 옵션을 사용합니다.

---

### 사용자 확인 단계

동기화 계획 검토 후 다음 중 선택하세요:
- **"진행"** 또는 **"시작"**: 계획대로 동기화 시작
- **"수정 [내용]"**: 동기화 계획 수정 요청
- **"중단"**: 동기화 작업 중단

---

## 🚀 STEP 2: 문서 동기화 실행 (사용자 승인 후)

사용자 승인 후 doc-syncer 에이전트가 **Living Document 동기화와 @TAG 업데이트**를 수행하고, 팀 모드에서만 PR Ready 전환을 선택적으로 실행합니다.

### Phase 2 세부: SPEC 완료 처리 (자동)

doc-syncer 에이전트가 TDD 구현 완료 여부를 자동으로 판단하여 SPEC 메타데이터를 업데이트합니다.

**자동 업데이트 조건**:
- status가 `draft`인 SPEC
- RED → GREEN → REFACTOR 커밋 존재
- @TEST 및 @CODE TAG 존재

**업데이트 내용**:
- `status: draft` → `status: completed`
- `version: 0.0.x` → `version: 0.1.0`
- HISTORY 섹션 자동 추가

**조건 미충족 시**: Phase 2 세부 작업 자동 건너뜀

## 기능

- **자동 문서 동기화**: doc-syncer 에이전트가 Living Document 동기화와 @TAG 업데이트를 수행합니다. 팀 모드에서만 PR Ready 전환을 선택적으로 실행합니다.

## 동기화 산출물

- `.moai/reports/sync-report.md` 생성/갱신
- TAG 체인 검증: 코드 직접 스캔 (`rg '@TAG' -n src/ tests/`)

## 모드별 실행 방식

## 📋 STEP 1 실행 가이드: 동기화 범위 분석 및 계획 수립

### 1. 프로젝트 상태 분석

Alfred는 doc-syncer 에이전트를 호출하여 동기화 대상과 범위를 분석합니다.

#### 분석 체크리스트

- [ ] **Git 상태**: 변경된 파일, 브랜치 상태, 커밋 히스토리
- [ ] **문서 일치성**: 코드-문서 간 동기화 필요성
- [ ] **TAG 시스템**: @TAG 체계 검증 및 끊어진 링크
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
- **TAG 인덱스**: `.moai/indexes/tags.db` 갱신
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

### ⚠️ 주의사항
- **잠재적 충돌**: [문서 충돌 가능성]
- **TAG 문제**: [끊어진 링크, 중복 TAG]
- **성능 영향**: [대용량 동기화 예상시간]

### ✅ 예상 산출물
- **sync-report.md**: [동기화 결과 요약]
- **tags.db**: [업데이트된 TAG 인덱스]
- **Living Documents**: [갱신된 문서 목록]
- **PR 상태**: [팀 모드에서 PR 전환]

---
**승인 요청**: 위 계획으로 동기화를 진행하시겠습니까?
("진행", "수정 [내용]", "중단" 중 선택)
```

---

## 🚀 STEP 2 실행 가이드: 문서 동기화 (승인 후)

사용자가 **"진행"** 또는 **"시작"**을 선택한 경우에만 Alfred는 doc-syncer 에이전트를 호출하여 Living Document 동기화와 TAG 업데이트를 수행합니다.

### 동기화 단계별 가이드

1. **Living Document 동기화**: 코드 → 문서 자동 반영
2. **TAG 시스템 검증**: @TAG 체계 무결성 확인
3. **인덱스 업데이트**: 트레이시빌리티 매트릭스 갱신
4. **보고서 생성**: 동기화 결과 요약 작성

### 에이전트 협업 구조

- **1단계**: `doc-syncer` 에이전트가 Living Document 동기화 및 @TAG 관리를 전담합니다.
- **2단계**: `git-manager` 에이전트가 모든 Git 커밋, PR 상태 전환, 동기화를 전담합니다.
- **단일 책임 원칙**: doc-syncer는 문서 작업만, git-manager는 Git 작업만 수행합니다.
- **순차 실행**: doc-syncer → git-manager 순서로 실행하여 명확한 의존성을 유지합니다.
- **에이전트 간 호출 금지**: 각 에이전트는 다른 에이전트를 직접 호출하지 않고, 커맨드 레벨에서만 순차 실행합니다.

## 🚀 최적화된 병렬/순차 하이브리드 워크플로우

### Phase 1: 빠른 상태 확인 (병렬 실행)

다음 작업들을 **동시에** 수행:

```
Task 1 (haiku): Git 상태 체크
├── 변경된 파일 목록 수집
├── 브랜치 상태 확인
└── 동기화 필요성 판단

Task 2 (sonnet): 문서 구조 분석
├── 프로젝트 유형 감지
├── TAG 목록 수집
└── 동기화 범위 결정
```

### Phase 2: 문서 동기화 (순차 실행)

`doc-syncer` 에이전트(sonnet)가 집중 처리:

- Living Document 동기화
- @TAG 시스템 검증 및 업데이트
- 문서-코드 일치성 체크
- TAG 추적성 매트릭스 갱신

### Phase 3: Git 작업 처리 (순차 실행)

`git-manager` 에이전트(haiku)가 최종 처리:

- 문서 변경사항 커밋
- 모드별 동기화 전략 적용
- Team 모드에서 PR Ready 전환
- 리뷰어 자동 할당 (gh CLI 사용)

### Phase 4: PR 머지 및 브랜치 정리 (선택적)

`--auto-merge` 플래그 사용 시 `git-manager`가 추가 처리:

**Team 모드 (GitFlow)**:
1. PR 상태 확인 (CI/CD 통과 체크)
2. PR 자동 머지 (develop 브랜치로)
3. 원격 feature 브랜치 삭제
4. 로컬 develop 체크아웃 및 동기화
5. 로컬 feature 브랜치 정리
6. 다음 작업 준비 완료 알림

**Personal 모드**:
1. 로컬 main/develop 머지
2. feature 브랜치 삭제
3. 베이스 브랜치 체크아웃
4. 다음 작업 준비 완료 알림

**성능 향상**: 초기 확인 단계를 병렬화하여 대기 시간 최소화

### 인수 처리

- **$1 (모드)**: `$1` → `auto`(기본값)|`force`|`status`|`project`
- **$2 (경로)**: `$2` → 동기화 대상 경로 (선택사항)
- **플래그**:
  - `--auto-merge`: PR 자동 머지 및 브랜치 정리 활성화 (Team 모드)
  - `--skip-pre-check`: 사전 품질 검증 건너뛰기
  - `--skip-quality-check`: 최종 품질 검증 건너뛰기

**커맨드 사용 예시**:
- `/alfred:3-sync` - 기본 자동 동기화 (모드별 최적화)
- `/alfred:3-sync --auto-merge` - PR 자동 머지 + 브랜치 정리 (Team 모드 권장)
- `/alfred:3-sync force` - 전체 강제 동기화
- `/alfred:3-sync status` - 동기화 상태 확인
- `/alfred:3-sync project` - 통합 프로젝트 동기화
- `/alfred:3-sync auto src/auth/` - 특정 경로 동기화
- `/alfred:3-sync --auto-merge --skip-pre-check` - 빠른 머지

### 에이전트 역할 분리

#### doc-syncer 전담 영역

- Living Document 동기화 (코드 ↔ 문서)
- @TAG 시스템 검증 및 업데이트
- API 문서 자동 생성/갱신
- README 및 아키텍처 문서 동기화
- 문서-코드 일치성 검증

#### git-manager 전담 영역

- 모든 Git 커밋 작업 (add, commit, push)
- 모드별 동기화 전략 적용
- PR 상태 전환 (Draft → Ready)
- **PR 자동 머지** (--auto-merge 플래그 시)
  - CI/CD 상태 확인
  - 충돌 검증
  - Squash 머지 실행
  - 원격 브랜치 삭제
- **브랜치 정리 및 전환**
  - 로컬 develop 체크아웃
  - 원격 동기화 (git pull)
  - 로컬 feature 브랜치 삭제
- 리뷰어 자동 할당 및 라벨링
- GitHub CLI 연동 및 원격 동기화

### 🧪 개인 모드 (Personal)

- git-manager 에이전트가 동기화 전/후 자동으로 체크포인트 생성
- README·심층 문서·PR 본문 정리는 체크리스트에 따라 수동 마무리

### 🏢 팀 모드 (Team)

- Living Document 완전 동기화 + @TAG 검증/보정
- gh CLI가 설정된 경우에 한해 PR Ready 전환과 라벨링을 선택적으로 실행
- **--auto-merge 플래그 사용 시 완전 자동화**:
  1. 문서 동기화 완료
  2. git push origin feature/SPEC-{ID}
  3. gh pr ready {PR_NUMBER}
  4. CI/CD 상태 확인 (gh pr checks)
  5. gh pr merge --squash --delete-branch
  6. git checkout develop && git pull origin develop
  7. 다음 작업 준비 완료 알림

**중요**: 모든 Git 작업(커밋, 동기화, PR 관리)은 git-manager 에이전트가 전담하므로, 이 커멘드에서는 Git 작업을 직접 실행하지 않습니다.

**브랜치 정책**:
- 베이스 브랜치: `develop` (GitFlow 표준)
- 머지 후: 자동으로 `develop` 체크아웃
- 다음 `/alfred:1-spec`은 자동으로 `develop`에서 시작

## 동기화 상세(요약)

1. 프로젝트 분석 및 TAG 검증 → 끊어진/중복/고아 TAG 점검
2. 코드 ↔ 문서 동기화 → API/README/아키텍처 문서 갱신, SPEC ↔ 코드 TODO 동기화
3. TAG 체인 검증 → `rg '@TAG' -n src/ tests/` (코드 직접 스캔)

## 다음 단계

**권장사항**: 다음 단계 진행 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하면 더 나은 성능과 컨텍스트 관리를 경험할 수 있습니다.

- 문서 동기화 완료 후 전체 MoAI-ADK 워크플로우 완성
- 모든 Git 작업은 git-manager 에이전트가 전담하여 일관성 보장
- 에이전트 간 직접 호출 없이 커멘드 레벨 오케스트레이션만 사용

## 결과 보고

동기화 결과를 구조화된 형식으로 보고합니다:

### 성공적인 동기화(요약 예시)

✅ 문서 동기화 완료 — 업데이트 N, 생성 M, TAG 수정 K, 검증 통과

### 부분 동기화 (문제 감지)

```
⚠️ 부분 동기화 완료 (문제 발견)

❌ 해결 필요한 문제:
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

**기본 모드 (PR Ready만)**:
```
🔄 MoAI-ADK 3단계 워크플로우 완성:
✅ /alfred:1-spec → EARS 명세 작성 (feature/SPEC-{ID} 브랜치)
✅ /alfred:2-build → TDD 구현
✅ /alfred:3-sync → 문서 동기화 + PR Ready

⏳ 다음 단계: PR 리뷰 및 수동 머지 필요
> gh pr view (PR 확인)
> gh pr merge --squash (리뷰 후 머지)
```

**자동 머지 모드 (권장)**:
```
🔄 완전 자동화된 GitFlow 워크플로우:
✅ /alfred:1-spec → EARS 명세 작성 (from develop)
✅ /alfred:2-build → TDD 구현
✅ /alfred:3-sync --auto-merge → 문서 동기화 + PR 머지 + 브랜치 정리

🎉 develop 브랜치로 자동 전환 완료!
📍 현재 위치: develop (다음 작업 준비됨)
> /alfred:1-spec "다음 기능 설명"  # develop에서 새 브랜치 생성
```

### 통합 프로젝트 모드

**사용 시점**:
- 여러 SPEC의 구현이 완료되어 프로젝트 전체 문서를 업데이트할 때
- Personal 모드에서 주기적인 전체 문서 동기화가 필요할 때

**Personal/Team 모드와의 차이**:
- **Personal/Team 모드**: 특정 SPEC 관련 문서만 동기화
- **Project 모드**: README, 아키텍처 문서, 전체 API 문서 동기화

**산출물**:
- README.md (전체 기능 목록 업데이트)
- docs/architecture.md (시스템 설계 갱신)
- docs/api/ (통합 API 문서)
- .moai/indexes/ (전체 TAG 인덱스 재구성)

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

## 🧠 Context Management (컨텍스트 관리)

> 자세한 내용: `.moai/memory/development-guide.md` - "Context Engineering" 섹션 참조

### 이 커맨드의 핵심 전략

**우선 로드**: `.moai/reports/sync-report-latest.md` (이전 동기화 상태)

**권장사항**: 문서 동기화가 완료되었습니다. 전체 MoAI-ADK 사이클(1-spec → 2-build → 3-sync)이 완료되었으니, 다음 기능 개발 전 `/clear` 또는 `/new` 명령으로 새로운 대화 세션을 시작하세요.

---

**doc-syncer 서브에이전트와 연동하여 코드-문서 일치성 향상과 @TAG 추적성 보장을 목표로 합니다.**
