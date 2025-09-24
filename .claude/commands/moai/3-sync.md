---
name: moai:3-sync
description: 문서 동기화 + PR Ready 전환
argument-hint: "모드 대상경로 - 모드: auto(기본)|force|status|project, 대상경로: 동기화 대상 경로"
allowed-tools: Read, Write, Edit, MultiEdit, Bash(git status:*), Bash(git add:*), Bash(git diff:*), Bash(git commit:*), Bash(gh:*), Bash(python3:*), Task, Grep, Glob, TodoWrite
---

# MoAI-ADK 3단계: 문서 동기화(+선택적 PR Ready)

doc-syncer 에이전트가 Living Document 동기화와 16-Core @TAG 업데이트를 수행합니다. 팀 모드에서만 PR Ready 전환을 선택적으로 실행합니다.

## 에이전트 협업 구조

- **1단계**: `doc-syncer` 에이전트가 Living Document 동기화 및 16-Core TAG 관리를 전담합니다.
- **2단계**: `git-manager` 에이전트가 모든 Git 커밋, PR 상태 전환, 동기화를 전담합니다.
- **단일 책임 원칙**: doc-syncer는 문서 작업만, git-manager는 Git 작업만 수행합니다.
- **순차 실행**: doc-syncer → git-manager 순서로 실행하여 명확한 의존성을 유지합니다.
- **에이전트 간 호출 금지**: 각 에이전트는 다른 에이전트를 직접 호출하지 않고, 커멘드 레벨에서만 순차 실행합니다.

## 브레인스토밍 리포트 (선택) - 커맨드 레벨 오케스트레이션

- `.moai/config.json.brainstorming.enabled` 가 `true` 인 경우, `/moai:3-sync` 커맨드에서 다음과 같이 오케스트레이션합니다:

### 에이전트 호출 시퀀스

1. **doc-syncer 에이전트**: 기본 동기화 수행
2. **브레인스토밍 단계** (병렬 실행):
   ```
   Task: codex-bridge (문서 개선점 분석)
   Task: gemini-bridge (문서 일관성 검토)
   Task: codex-bridge (리스크 분석)
   (최대 10개 병렬 처리)
   ```
3. **doc-syncer 재호출**: 브레인스토밍 결과를 요약에 반영

- 설정이 비활성화되어 있으면 doc-syncer만 단일 실행합니다.

## 동기화 산출물 (0.2.2)

- `.moai/reports/sync-report.md` 생성/갱신
- TAG 인덱스 업데이트: `python3 .moai/scripts/check-traceability.py --update`

## 모드별 실행 방식

## 워크플로우 실행 순서

당신은 다음 순서로 에이전트들을 **순차 호출**해야 합니다:

### 1단계: 문서 동기화

먼저 `doc-syncer` 에이전트를 호출하여 다음 작업을 완료합니다:

- Living Document 동기화
- 16-Core TAG 시스템 검증 및 업데이트
- 문서-코드 일치성 체크
- TAG 추적성 매트릭스 갱신

### 2단계: Git 작업 처리

`doc-syncer` 완료 후, `git-manager` 에이전트를 호출하여 다음 작업을 수행합니다:

- 문서 변경사항 커밋
- 모드별 동기화 전략 적용
- Team 모드에서 PR Ready 전환
- 리뷰어 자동 할당 (gh CLI 사용)

**중요**: 각 에이전트는 독립적으로 실행되며, 에이전트 간 직접 호출은 금지됩니다.

### 인수 처리

- **$1 (모드)**: `$1` → `auto`(기본값)|`force`|`status`|`project`
- **$2 (경로)**: `$2` → 동기화 대상 경로 (선택사항)

```bash
# 기본 자동 동기화 (모드별 최적화)
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

### 에이전트 역할 분리

#### doc-syncer 전담 영역

- Living Document 동기화 (코드 ↔ 문서)
- 16-Core TAG 시스템 검증 및 업데이트
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

- git-manager 에이전트가 동기화 전/후 자동으로 체크포인트 생성
- README·심층 문서·PR 본문 정리는 체크리스트에 따라 수동 마무리

### 🏢 팀 모드 (Team)

- Living Document 완전 동기화 + 16-Core TAG 검증/보정
- gh CLI가 설정된 경우에 한해 PR Ready 전환과 라벨링을 선택적으로 실행

**중요**: 모든 Git 작업(커밋, 동기화, PR 관리)은 git-manager 에이전트가 전담하므로, 이 커멘드에서는 Git 작업을 직접 실행하지 않습니다.

## 동기화 상세(요약)

1. 프로젝트 분석 및 TAG 검증 → 끊어진/중복/고아 TAG 점검
2. 코드 ↔ 문서 동기화 → API/README/아키텍처 문서 갱신, SPEC ↔ 코드 TODO 동기화
3. TAG 인덱스 업데이트 → `python3 .moai/scripts/check-traceability.py --update`

## 다음 단계

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

**doc-syncer 서브에이전트와 연동하여 코드-문서 일치성 향상과 16-Core TAG 추적성 보장을 목표로 합니다.**
