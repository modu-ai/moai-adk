---
id: SPEC-HANDOFF-AUTORESUME-001
title: "핸드오프 auto-resume — 구현 계획 (Tier L)"
version: "0.1.0"
status: draft
created: 2026-07-05
updated: 2026-07-05
author: MoAI
priority: P2
phase: "v3.0.0"
module: "internal/config, internal/hook, internal/cli"
lifecycle: spec-anchored
tags: "session-handoff, auto-resume, handoff-config, sessionstart, hook, cli, tier-l, epic-handoff-v2"
tier: L
era: V3R6
related_specs: [SPEC-HANDOFF-CTXGUIDE-001, SPEC-HANDOFF-MSGMODE-001, SPEC-V3R6-SESSION-HANDOFF-AUTO-001]
---

# Plan — SPEC-HANDOFF-AUTORESUME-001 (Handoff-v2 M3/4, auto-resume)

> Tier L 구현 계획. 우선순위 라벨만 사용(시간 추정 금지). 확정 사용자 결정에 따라 milestone-split M1 → M2 → M3. 각 milestone은 독립 커밋 단위이며 AC 바인딩을 갖는다.

## §A — Context

역방향 핸드오프(save → SessionStart 주입/소비)를 3개 milestone으로 landing한다. 기존 SessionEnd → memory 절반(`persist.go`)은 무접촉. 실측 근거는 research.md, 설계 결정은 design.md.

개발 방식: `quality.yaml` `development_mode` 따름. 본 SPEC은 Go 신규 코드 다수(config struct + hook handler + CLI) → TDD 권장(RED → GREEN → REFACTOR). 각 milestone은 test-first.

## §B — Known Issues (본 SPEC 적용분)

- **B3/B11 서브에이전트 경계 (C-HRA-008)**: `handoffInjectHandler`는 AskUserQuestion/mcp__askuser 미호출. 정적 grep `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go | grep -v "^[^:]*:[0-9]*:[ \t]*//"` = 0. → AC-AUTORESUME-016.
- **B4 frontmatter 12-field**: spec.md 준수(canonical 이름, snake_case alias 금지). 검증: `moai spec lint spec.md`.
- **B6 Out-of-Scope h3**: spec.md §B.2에 `### Out of Scope — <topic>` h3 sub-section 6개 + `-` bullet. h2 단독 금지.
- **B7 projectDir 해석**: `handoffInjectHandler`는 `resolveProjectDir(input)`(compact.go:75, CWD-first, ProjectDir fallback) 재사용 — worktree cwd 대응.
- **B8 working-tree hygiene**: 훅은 자기 `handoff/{pending,consumed}` 영역만 mutate. `session-handoff/` 및 기타 runtime-managed 디렉터리 무접촉.
- **B10 PRESERVE**: worktree 내에서 `.moai/specs/SPEC-HANDOFF-AUTORESUME-001/`만 authoring. 구현(run-phase)은 별도.

## §C — Pre-flight (실측 완료, plan-phase)

- HEAD `97723664c` 확인 (clean base).
- `grep -rn 'HandoffConfig' internal/` = 0 (greenfield struct in existing package).
- registry.go accumulate-all 실측(§C research). SessionStart matcher 이미 clear 포함 실측(§D research).
- ResearchConfig 패턴 6-지점 확인(§E research).

## §D — Constraints (위반 금지)

- Tier L 5-artifact: spec/plan/acceptance/design/research (+ progress.md skeleton).
- GEARS REQ (REQ-AUTORESUME-NNN) + AC (AC-AUTORESUME-NNN).
- 3개 확정 결정 FIXED: mode default=manual / directive degrade-to-guidance / M1-M2-M3 split. 대안 제시 금지.
- verification-claim-integrity: 주입 콘텐츠는 xhigh/ultrathink 활성 미주장.
- authoring ONLY — git add/commit/push 금지 (orchestrator가 plan-auditor + Kickoff Approval 후 커밋).
- AskUserQuestion 미호출. blocker는 구조화 "Missing Inputs" 보고.

## §E — Self-Verification (plan-phase audit-ready signal)

- [ ] 5-artifact 작성 완료 + progress.md §E skeleton
- [ ] frontmatter 12-field + tier/era/related_specs
- [ ] REQ 19개 ↔ AC 19개 1:1 (REQ-007 split → REQ-019; Consume 필드 제거)
- [ ] Out-of-Scope h3 6개 + bullet
- [ ] `moai spec lint spec.md` + `plan.md` (file-path form) No findings
- [ ] 경로 분리 verdict(handoff/ vs session-handoff/) design 명시
- [ ] registry accumulate-all + matcher-already-clear 실측 반영

## §F — Milestones (우선순위 순)

### M1 — HandoffConfig landing (Priority: High)

**목표**: `{mode, guide}` config를 ResearchConfig 패턴으로 landing. auto-resume의 기반. (`consume` 필드는 YAGNI 제거 — design.md §E.1.)

**파일 세트**:
- `internal/config/types.go` (+ `HandoffConfig` struct, Config `Handoff` 필드, `handoffFileWrapper`)
- `internal/config/defaults.go` (+ `NewDefaultHandoffConfig`, `NewDefaultConfig` 등록)
- `internal/config/loader.go` (+ `loadHandoffSection`, `Load()` 순서 추가)
- `internal/config/audit_registry.go` (+ `"handoff": "HandoffConfig"`)
- `internal/config/*_test.go` (default/partial-override/parity 테스트)
- `internal/template/templates/.moai/config/sections/handoff.yaml` (신규, 중립)
- `.moai/config/sections/handoff.yaml` (live, `make build` 후 sync)

**AC 바인딩**: AC-AUTORESUME-001, 002, 003, 004.

**검증**: `go test ./internal/config/...`, `go test -run TestAuditParity ./internal/config/`, `make build`.

**주의**: Template-First — handoff.yaml을 template에 먼저 추가 후 `make build`. audit registry 미등록 시 CI orphan 실패.

### M2 — `moai handoff save` / `clear` CLI (Priority: High)

**목표**: `handoff/pending.json`을 별도 경로에 atomic 작성하는 CLI. `session-handoff/` 무접촉.

**파일 세트** (D5 — 등록 위치·writer 패키지 확정, "또는" 제거):
- `internal/cli/handoff.go` (신규 — `handoffCmd`(top-level) + `save`/`clear` 하위, cobra. `init()`에서 `rootCmd.AddCommand(handoffCmd)` 자체 등록 — glm.go/cc.go/doctor.go와 동일 패턴)
- `internal/cli/handoff_test.go` (신규)
- `internal/hook/handoff/` (기존 패키지 재사용 — pending.json writer 추가, `atomicWriteFile` 재사용. 신규 `internal/handoff/` 생성하지 않음)

**AC 바인딩**: AC-AUTORESUME-005, 006, 007 (clear CLI만; stale TTL cleanup은 M3 AC-019로 분리 — milestone 경계 정합).

**검증**: `go test ./internal/cli/... ./internal/hook/handoff/...`, `moai handoff save --body <x>` smoke, pending.json 스키마 assert. `session-handoff/pending.md` 미접촉 grep.

**주의**: save/clear는 `handoff/`만. `session-handoff/pending.md` write/read 금지(REQ-008/007 회귀). body는 flag/stdin. TTL 상수는 `config/defaults.go` 단일 정의(M3 handler가 소비).

### M3 — SessionStart 주입/소비 (Priority: High)

**목표**: `source==clear ∧ mode==auto`일 때만 주입+소비하는 `handoffInjectHandler`. claim-then-inject + degrade-to-guidance + nonce fallback.

**파일 세트**:
- `internal/hook/handoff_inject.go` (신규 — `handoffInjectHandler`, `NewHandoffInjectHandler(cfg ConfigProvider)`, EventSessionStart. 4-source×mode branch table + claim-then-inject + auto-only TTL cleanup + nonce fallback + i18n degrade)
- `internal/hook/handoff_inject_test.go` (신규 — branch table 8-cell(auto 4 + manual 4), guide∈{true,false}, race, rename-fail fail-open, nonce shape, i18n, degrade, TTL auto-only, 3-handler coexist)
- `internal/cli/deps.go` (+ `deps.HookRegistry.Register(hook.NewHandoffInjectHandler(deps.Config))` — sessionStartHandler/autoUpdateHandler 다음, 3번째 등록)
- (settings.json 무변경 — research.md §D)

**AC 바인딩**: AC-AUTORESUME-008, 009, 010, 011, 012, 013, 014, 015, 016, 017, 018, 019 (12개; 019=auto-only TTL cleanup).

**검증**: `go test ./internal/hook/...`, `go test -race ./internal/hook/...` (race AC-013), 경계 grep(AC-016), branch table 8-cell + guide 양분기 테스트.

**주의**:
- claim-then-inject 순서(rename 성공 후 주입) — 중복 주입 방지. rename 실패는 errno 무관 fail-open(AC-013, Windows `MoveFileEx` 호환).
- TTL cleanup은 `mode==auto`에서만(AC-019); manual은 stale이어도 pure no-op(AC-009 stale sub-case).
- 3개 SessionStart 핸들러 additionalContext 공존 e2e 테스트(registry accumulate-all 회귀, AC-018).
- matcher 실측 grep(assertion, 변경 아님) — `startup|resume|clear|compact` 존재 확인.

## §G — Anti-Patterns (회피)

- 경로 공유(`session-handoff/pending.md`에 auto-resume 쓰기) → SessionEnd race. **별도 handoff/ 트리 필수.**
- 주입 콘텐츠에 "ultrathink 활성" 주장 → verification-claim-integrity 위반. degrade-to-guidance만.
- settings.json matcher 수정 → 불필요 + template regression 위험. 이미 clear 포함, assertion만.
- delete-on-consume → audit trail 소실. rename-to-consumed 필수.
- inject-then-claim 순서 → race 중복 주입. claim(rename)-then-inject 필수.
- config default를 auto로 → 기본 UX 변경. manual 고정.

### Out of Scope — plan-level (구현 계획 범위 밖)
- SessionEnd → memory 흐름(`persist.go`) 재구현/통합 — 무접촉 (spec.md §B.2 참조).
- settings.json matcher 변경 — 이미 `startup|resume|clear|compact`, assertion만.
- M4 threshold-guidance 2-stage 로직 — 후속 SPEC 소관.
- 6-block paste-ready 파싱/생성 — orchestrator self-discipline.

## §H — Cross-References

- research.md §A~H, design.md §A~I, acceptance.md §D
- `internal/hook/handoff/persist.go` (SessionEnd 절반, 무접촉 참조)
- `internal/hook/registry.go`, `internal/hook/compact.go` resolveProjectDir
- `internal/config/loader.go` loadResearchSection (미러)
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1
- `.claude/rules/moai/workflow/session-handoff.md`
