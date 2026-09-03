# acceptance.md — SPEC-PRECOMMIT-GATE-SCOPE-001

id: SPEC-PRECOMMIT-GATE-SCOPE-001
created: 2026-09-03

## D. AC Matrix

> 축 (a)/(b) 최종 선택은 Implementation Kickoff Approval에서 운영자가 확정한다(plan.md §C).
> `axis-conditional` 표기 AC는 확정된 축이 통과 조건을 결정하며, 미확정 상태에서는 어느 축으로도
> 판정 가능한 형태로 기술한다.

### AC-001 — 실패 메시지가 remedy를 안내한다 (축 (c), 무조건) — REQ-003

**Given** `moai gate`가 실패하는 프로젝트(예: 기존 실패 테스트 1건)
**When** pre-commit 훅이 실행되어 heavy gate가 실패한다
**Then** 훅의 stderr 출력에 다음 4개 문자열이 모두 존재한다 (grep 기계판정):
`.moai/config/sections/gate.yaml`, `gate.enabled`, `gate.skip_tests`, `gate.disabled_steps`.
또한 기존 안내 `SKIP_MOAI_PRECOMMIT=1`이 유지된다.

### AC-002 — 무관한 기존 실패가 커밋을 막지 않는다 (axis-conditional) — REQ-002

**Given** 기존(project-wide) 실패 1건이 있고, staged 변경은 그 실패와 무관하다
**When** 기본 설정(사용자가 아무 것도 opt-in하지 않은 상태)으로 `git commit`한다
**Then** 커밋이 성공한다.
- 축 (a) 확정 시: heavy gate가 staged 범위로 실행되어 무관한 기존 실패가 판정에서 제외됨으로써 성공.
- 축 (b) 확정 시: heavy gate가 기본 OFF로 실행되지 않음으로써 성공.

### AC-003 — twin byte identity (무조건) — REQ-005

**Given** `preCommitHookContent`(`internal/cli/hook_install_precommit.go`)와
`internal/template/templates/.git_hooks/pre-commit`이 편집되어 있다
**When** `go test ./internal/cli/ -run TestPreCommitTemplateMatchesConstant -count=1`를 실행한다
**Then** 테스트가 통과한다. (같은 목적의 전 package 실행 `go test ./internal/cli/...`에서도 통과)

### AC-004 — 사용자 gate.yaml이 `moai update`를 생존한다 (무조건) — REQ-006

**Given** 기존 설치 프로젝트의 `.moai/config/sections/gate.yaml`에 사용자 커스텀 값이 있다
(예: `skip_tests: true` 또는 `disabled_steps` 항목)
**When** `moai update`를 실행한다
**Then** update 후에도 해당 커스텀 값이 유지된다 (shipped 템플릿 기본값으로 되돌아가지 않는다).
테스트는 fixture 프로젝트에서 update 흐름을 실행해 파일 내용을 직접 비교한다.

### AC-005 — t237 소관 단계 무변경 (무조건) — REQ-008

**Given** 이 SPEC의 구현이 완료되어 있다
**When** 훅 본문의 fast-subset 구간을 검사한다
**Then** staged 파일 수집(`git diff --cached --name-only --diff-filter=ACM | grep '\.go$'`)과
`go vet $BT_TAGS $PKGS` 호출 및 `.moai/config/build-tags` 읽기 로직이 구현 전과 의미적으로 동일하다.
기계판정: 해당 구간에 대한 diff가 주석 수준이거나 0이며, `go vet` 관련 줄이 변경되지 않았음을
`git diff <base> -- internal/template/templates/.git_hooks/pre-commit`으로 확인한다.

### AC-006 — opt-in 시 heavy gate가 실행·차단한다 (axis-conditional, 축 (b) 확정 시 필수) — REQ-004

**Given** 사용자가 gate.yaml로 heavy gate를 opt-in했다 (확정된 메커니즘의 키)
**When** `moai gate`가 실패하는 상태에서 pre-commit이 실행된다
**Then** 커밋이 차단되고(종료코드 1) AC-001의 실패 메시지가 출력된다.
반대 케이스: opt-in 상태에서 gate가 통과하면 커밋이 성공한다.

### AC-007 — non-moai 프로젝트 무음 통과 (무조건) — REQ-007

**Given** PATH에 `moai`가 없는 환경이다
**When** pre-commit 훅이 실행된다
**Then** heavy gate 없이 종료코드 0으로 통과한다 (기존 동작 회귀 없음).

### AC-008 — 기존 설치에 수리가 반영된다 (무조건) — REQ-002 전제

**Given** 이전 버전의 MoAI 훅(marker 일치)이 설치된 프로젝트다
**When** 새 바이너리로 `moai update`를 실행한다
**Then** `.git/hooks/pre-commit`이 새 내용으로 교체된다 (기존 marker-based overwrite 경로,
`InstallPreCommitHook`의 marker 분기 — 별도 마이그레이션 없이 자동 반영됨을 테스트로 확인).

## D.1 심각도

| AC | 심각도 | 근거 |
|----|--------|------|
| AC-001 | Must | 사용자가 remedy에 도달하지 못하면 축 (a)/(b) 어느 쪽으로도 잔여 결함 |
| AC-002 | Must | 사고의 직접 재현 형태 — 이 SPEC의 존재 이유 |
| AC-003 | Must | twin 계약 위반은 배포판과 개발본의 갈라짐을 만든다 |
| AC-004 | Must | remedy 표면이 update마다 소멸하면 REQ-003 안내가 허위가 된다 |
| AC-005 | Must | t237 경합 범위 침범 방지 |
| AC-006 | Must (축 (b) 시) / N/A (축 (a) 시) | opt-in 계약의 양방향 검증 |
| AC-007 | Should | 기존 동작 회귀 방지 |
| AC-008 | Should | 기존 설치 반영은 `moai update` overwrite로 자동이므로 확인용 |

## D.2 간접 검증 (Indirect Verification)

- AC-002는 end-to-end(실제 git 커밋 + fixture 프로젝트)로 검증하며, 단위 테스트만으로 대체하지 않는다.
- AC-004는 `CleanMoaiManagedPaths` 경로를 실제로 통과하는 update 흐름 테스트로 검증한다
  (fixture는 `t.TempDir()` — CLAUDE.local.md §6 격리 규칙).

## D.3 품질 게이트 (Quality Gate Criteria)

- `go test ./internal/cli/... -count=1` 통과
- `go test ./internal/config/... -count=1` 통과
- `go test ./internal/hook/... -count=1` 통과
- `make build` 성공 (템플릿 재생성 + agents-emit-check + catalog hash 체인)
- 템플릿 중립성: 훅 템플릿에 SPEC ID / 내부 날짜 / 커밋 SHA 없음 (§25.1 forbidden classes)
- go vet 영향 패키지 경고 0

## D.4 Definition of Done

1. AC-001 ~ AC-005, AC-007, AC-008 PASS (축 확정 시 AC-006 포함). 각 PASS는 실행한 명령과
   관측 출력을 인용한다.
2. 축 (a)/(b) 선택이 progress.md에 기록되고, axis-conditional AC의 통과 조건이 확정 축으로 확정돼 있다.
3. plan.md §C의 [NEEDS CLARIFICATION] 마커가 해소돼 있다.
4. t237과의 병합 순서/충돌 메모가 progress.md에 기록돼 있다.
