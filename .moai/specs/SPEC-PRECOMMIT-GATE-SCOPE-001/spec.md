---
id: SPEC-PRECOMMIT-GATE-SCOPE-001
title: "Pre-commit heavy gate scope defect — project-wide quality gate must not block commits on unrelated pre-existing failures"
version: "1.0.0"
status: draft
created: 2026-09-03
updated: 2026-09-03
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/template/templates, internal/config"
lifecycle: spec-anchored
tags: "pre-commit, quality-gate, gate.yaml, template, hook"
related_specs: [SPEC-PRECOMMIT-001, SPEC-PRETOOL-GATE-MOVE-001, SPEC-PRECOMMIT-PRESERVE-001]
---

# SPEC-PRECOMMIT-GATE-SCOPE-001

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-09-03 | manager-spec | plan-phase 최초 작성 (card t461, Class C / Tier M) |
| 2026-09-03 | manager-spec | 수리 라운드: 운영자 결정 3건 반영 (축 (b) + 메커니즘 1 + moai web 편집 표면). REQ-003 키 확장, REQ-009 신설, plan.md NEEDS CLARIFICATION 해소, AC-009/AC-010 신설, AC-004 web 작성값 흡수 |

## A. 배경 (Background)

### 범위 요약 (WHAT)

pre-commit heavy gate(`moai gate` 호출)의 기본 동작을 프로젝트 전역 강제에서 **pre-commit 맥락 한정
기본 OFF + `gate.pre_commit.enabled` opt-in**으로 전환하고, 그 키를 `moai web` 설정 화면에서 편집
가능하게 하며, 실패 메시지가 실제 remedy 키를 안내하게 한다. 무관한 기존 실패가 커밋을 막는 결함
(SPEC-PRETOOL-GATE-MOVE-001 진입)을 제거한다. HOW(메커니즘 상세)는 plan.md, 판정은 acceptance.md.

`moai update`가 자동 설치하는 git pre-commit 훅(`internal/template/templates/.git_hooks/pre-commit`)은
후반부에서 `command -v moai` 성공 시 **무조건** `moai gate`를 실행한다. `moai gate`는 16개 프로그래밍
언어를 감지해 프로젝트 전역(project-wide)의 vet/lint/test/typecheck를 수행한다(internal/cli/gate.go
`runGate` → internal/hook/quality/gate.go `QualityGate.Run`). 즉 훅이 커밋 단위가 아닌 **프로젝트
전체 상태**를 커밋 허용 조건으로 삼는다.

결과로 관측된 사용자 사고: 한 건의 기존 실패(예: pytest 실패 목록)만 있어도 무관한 변경의 모든 커밋이
`Quality gate failed`(`runGate`의 `fmt.Errorf("quality gate failed")`)로 막히고, 유일한 탈출구는
매 커밋마다 `SKIP_MOAI_PRECOMMIT=1`을 붙이는 것이다. 사용자에게 귀책이 전혀 없다 — 훅은 `moai update`
가 자동 설치하며, 대부분의 사용자는 `--no-hooks` 플래그의 존재조차 모른다.

### 결함 기원 (verified)

| 커밋 | 날짜 | 내용 |
|------|------|------|
| `52b5e4bf5` | 2026-07-05 | SPEC-PRECOMMIT-001 — staged 파일 대상 gofmt+vet **fast subset** 훅 도입 (무해) |
| `883d53852` | 2026-07-28 | SPEC-PRETOOL-GATE-MOVE-001 (issue #1189) — PreToolUse 5초 예산의 heavy gate를 커밋 훅으로 이동하면서 **범위가 커밋 단위에서 프로젝트 전역으로 확대**됨. 결함 진입점 |

### 1. 원칙적 자기모순 (rationale)

이 저장소 자신의 규율(CLAUDE.local.md §4/§6)은 **로컬 전체 스위트 실행을 금지**하고, 변경이 영향을 줄
수 있는 패키지만 검증하며, 전체 스위트 판정은 CI의 몫으로 규정한다. 그런데 이 도구는 배포 기본값으로
정확히 그 금지 관행을 사용자 프로젝트의 모든 커밋에 탑재한다. 도구가 자기 저장소의 교리를 위반하는
형태로 배포되는 것은 설계 결함이다.

### 2. 검증된 코드 좌표 (심볼 기준 — 줄번호는 drift한다)

| 주장 | 검증 근거 (2026-09-03, 본 워크트리 a239cf050 기준) |
|------|------|
| 설치 호출점 | `internal/cli/update_template_sync.go`의 `installPreCommitHookOptional(projectRoot, getBoolFlag(cmd, "no-hooks"), out, errOut)` — 카드 텍스트는 :574였으나 실측 :575. 심볼로 인용 |
| twin 본문 | `internal/cli/hook_install_precommit.go`의 `preCommitHookContent` 상수와 `internal/template/templates/.git_hooks/pre-commit` (3,245 bytes) — byte identity 의무 (같은 파일 주석) |
| twin 강제 테스트 | `TestPreCommitTemplateMatchesConstant` (`internal/cli/hook_install_precommit_test.go`), `internal/cli/precommit_relocation_test.go`에서 AC-PGM-010으로 교차참조 |
| heavy gate 블록 | 템플릿 후반부 `if command -v moai ... then if ! moai gate ... fi` — staged 범위 검사 없이 무조건 실행 |
| 게임 규칙 기본값 | `internal/config/defaults.go` `NewDefaultGateConfig()`: `Enabled: true`, `SkipTests: false` — heavy gate 기본 ON, 전체 test 포함 |
| gate.yaml 경로·로더 | `.moai/config/sections/gate.yaml` — `internal/config/loader_gate.go` `loadGateSection`이 `Loader.Load`(`internal/config/loader.go`)의 `sectionsDir`에서 로드 |
| remedy 키 3종 | `internal/config/types.go` `GateConfig`: `enabled`, `skip_tests`, `disabled_steps` — 카드가 지목한 세 키 모두 존재 |
| 실패 메시지 | 훅이 내는 유일 안내는 `Override: SKIP_MOAI_PRECOMMIT=1 git commit` — gate.yaml 키는 어디에도 안내되지 않음 |
| gate.yaml 소멸 위험 | `internal/cli/update/deploy/deploy.go` `CleanMoaiManagedPaths`가 `.moai/config/`를 **통째 삭제 후 템플릿 재배포** — 사용자가 gate.yaml을 고쳐도 다음 `moai update`에서 기본값으로 되돌아감 (§D.3 제약) |
| Enabled 소비처 | `internal/hook/quality/gate.go` `QualityGate.Run` — `config.Enabled == false`이면 즉시 `(true, "")` 반환. 단독 `moai gate` CLI도 같은 스위치를 통과하므로, 템플릿 `gate.enabled`를 false로 뒤집으면 단독 실행도 꺼진다 |

## B. 요구사항 (GEARS Requirements)

### REQ-001 — 커밋 단위 품질 계약 (Ubiquitous)

The pre-commit hook shall enforce a commit-unit quality contract: only quality failures attributable to the staged change may block the commit.

### REQ-002 — 무관한 기존 실패가 커밋을 막지 않는다 (Event-driven, axis-conditional)

When a pre-existing project-wide quality failure unrelated to the staged change exists, the pre-commit hook shall allow the commit of an unrelated staged change.

> 이 REQ의 충족 형태는 설계 축 선택에 따라 달라진다. 축 (a): heavy gate가 staged 범위로 좁혀져 무관한
> 기존 실패가 판정에서 제외된다. 축 (b): heavy gate가 기본 OFF여서 기본 동작에서 실행되지 않는다.
> 어떤 축이 선택되든 REQ-002 자체는 불변이며, AC-002의 통과 조건이 선택된 축으로 확정된다.

### REQ-003 — 실패 메시지가 실제 remedy를 안내한다 (Event-detected, 축 (c) — 무조건)

When the heavy gate fails in the pre-commit context, the hook shall print a failure message that names, in addition to `SKIP_MOAI_PRECOMMIT=1`, the config path `.moai/config/sections/gate.yaml` and the four keys `gate.pre_commit.enabled`, `gate.enabled`, `gate.skip_tests`, `gate.disabled_steps`.

### REQ-004 — opt-in 시에만 heavy gate가 실행된다 (Capability gate, 축 (b) 메커니즘)

Where the user has opted in through gate.yaml configuration, the pre-commit hook shall run the heavy gate and shall block the commit on its failure.

### REQ-005 — twin byte identity 유지 (Ubiquitous)

The pair `preCommitHookContent` (`internal/cli/hook_install_precommit.go`) and `internal/template/templates/.git_hooks/pre-commit` shall remain byte-identical, as enforced by `TestPreCommitTemplateMatchesConstant`.

### REQ-006 — 사용자 gate.yaml 커스터마이징의 `moai update` 생존 (Event-driven)

When `moai update` runs against an existing install whose `.moai/config/sections/gate.yaml` carries user-customized values, the update shall preserve those user values (shall not reset them to the shipped template defaults).

### REQ-007 — non-moai 프로젝트 무음 통과 유지 (State-driven)

While `moai` is absent from PATH, the pre-commit hook shall skip the heavy gate and exit 0 (기존 non-moai downstream 프로젝트 동작 유지).

### REQ-008 — t237 소관 단계 무변경 (Ubiquitous)

The hook's fast-subset staged-scope go vet step (its staged-file collection and `go vet $BT_TAGS $PKGS` invocation semantics) shall remain semantically unchanged by this SPEC.

### REQ-009 — `moai web` 설정 표면에서 편집 가능 (Capability gate, 운영자 지시)

Where the `moai web` settings UI is running, the settings surface shall expose `gate.pre_commit.enabled` as an editable field and shall persist changes to exactly `.moai/config/sections/gate.yaml`.

> **명명 중복 주의 (naming-overlap hazard)** — 기존 `pre_commit` 키는 전부
> `git_strategy.<mode>.hooks.pre_commit`(문자열 필드, `internal/config/validation.go`
> `checkStringField` 3곳, `internal/cli/hook_install_precommit.go`의 config-agnostic 주석이
> 참조)이다. 이는 `git_strategy.*` 하위 트리로, 본 SPEC의 `gate.pre_commit.enabled`와는
> **다른 subtree**다. run-phase는 두 키를 혼동하지 않는다 — web 표면·게이트 러너·훅은 오직
> `gate.pre_commit.enabled`만 다룬다.

## C. 성공 기준 (Success Criteria)

- 기존 실패 1건이 있는 프로젝트에서 무관한 변경의 커밋이 기본 동작으로 성공한다 (REQ-002).
- pre-commit 마커 하에서 `gate.pre_commit.enabled`가 거짓이면 heavy gate가 실행되지 않는다 (REQ-001).
- heavy gate 실패 메시지에 `.moai/config/sections/gate.yaml`과 `gate.pre_commit.enabled`, `gate.enabled`, `gate.skip_tests`, `gate.disabled_steps`가 모두 등장한다 (REQ-003, grep 기계판정).
- `TestPreCommitTemplateMatchesConstant`가 통과한다 (REQ-005).
- 사용자가 고치거나(손편집·`moai web` 경로 모두) gate.yaml이 `moai update` 후에도 유지된다 (REQ-006).
- `moai web` 설정 화면에서 `gate.pre_commit.enabled`를 편집하면 gate.yaml에 반영된다 (REQ-009).
- fast-subset go vet 단계의 의미가 불변이다 (REQ-008).

## Out of Scope — t237 (go vet module-resolution defect)

- card t237 / issue #1641이 소관인 staged go vet 모듈해석 결함의 수리. 이 SPEC은 같은 twin 파일
  (`preCommitHookContent` 상수 + `.git_hooks/pre-commit` 템플릿)을 건드리므로 충돌 가능성이 있으며,
  병합 순서와 rebase 책임은 plan.md §C 제약과 progress.md에 기록한다. 이 SPEC은 t237의 패치 내용을
  선취·재현·수정하지 않는다.

## Out of Scope — 기존 탈출구와 단독 gate의 재설계

- `SKIP_MOAI_PRECOMMIT=1` 탈출구의 제거 또는 이름 변경 (REQ-003은 이를 유지하면서 안내를 확장한다).
- 단독 `moai gate` CLI 실행 자체의 제거. REQ-004의 메커니즘은 pre-commit 호출 맥락을 겨냥해야 하며,
  템플릿 `gate.enabled`를 단순히 false로 바꿔 단독 실행까지 꺼지게 하는 것은 허용되지 않는다.
- pre-push 훅, PreToolUse 훅 등 다른 훅 계층의 변경.

## Out of Scope — 프로젝트 전역 검증 능력의 제거

- `moai gate`가 프로젝트 전역 검증을 수행하는 능력 자체. 이 SPEC은 그것이 **커밋 게이트 기본값으로
  강제되는 것**을 문제 삼을 뿐, opt-in한 사용자의 명시적 실행(REQ-004)은 보존한다.
