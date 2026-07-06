---
id: SPEC-HANDOFF-AUTORESUME-001
title: "핸드오프 auto-resume — handoff.yaml landing + save CLI + SessionStart 주입/소비"
version: "0.1.0"
status: completed
created: 2026-07-05
updated: 2026-07-06
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

# SPEC-HANDOFF-AUTORESUME-001 — 핸드오프 auto-resume

> Epic "Handoff-v2" M3/4. 선행: M1 SPEC-HANDOFF-CTXGUIDE-001 (256K 밴드, `completed`) · M2 SPEC-HANDOFF-MSGMODE-001 (message-v2 mode-seed, `completed`, origin `97723664c`). 후속: M4 threshold-guidance (본 SPEC이 landing하는 HandoffConfig를 소비). 본 SPEC은 M3으로, **역방향 핸드오프(save → SessionStart 주입/소비) + HandoffConfig landing**을 담당한다.

## HISTORY

- 2026-07-05: 초안 (draft). Plan-phase 3-artifact + design/research (Tier L 5-artifact). registry merge accumulate-all 실측 반증, SessionStart matcher 이미 clear 포함 실측 반증(settings 무변경), 경로 분리(`handoff/` vs `session-handoff/`) verdict 반영. 3개 확정 사용자 결정(mode default=manual, directive degrade-to-guidance, M1/M2/M3 split) 준수.

---

## §A — Context (배경과 기존 상태)

### A.1 문제

Epic Handoff-v2의 목표는 세션 경계를 넘는 컨텍스트 연속성이다. M1(256K 밴드 안내)·M2(message-v2 mode-seed)는 **사람이 붙여넣는(manual)** paste-ready 흐름을 개선했다. M3은 opt-in **자동 재개**를 추가한다: 이전 세션이 저장한 핸드오프를 다음 세션 시작 시 `additionalContext`로 주입한다.

기존 `SPEC-V3R6-SESSION-HANDOFF-AUTO-001`(status `completed`)은 **SessionEnd → memory** 절반을 구현했다(`internal/hook/handoff/persist.go`). 본 SPEC은 **역방향 절반**(`moai handoff save` → SessionStart 주입/소비)을 추가한다. 두 절반은 별도 경로·포맷·소비자로 완전 분리된다(research.md §B, design.md §B).

### A.2 확정 사용자 결정 (FIXED — 재론 금지)

1. **Scope**: 단일 Tier L SPEC, milestone-split M1(HandoffConfig landing) → M2(`moai handoff save` CLI) → M3(SessionStart clear-branch 주입/소비). 3개 SPEC 아님.
2. **Directive degrade-to-guidance**: SessionStart 자동 주입 시 핸드오프 컨텍스트+body만 additionalContext에 주입. 모드-변경 지시자(`ultrathink.`/`ultracode`/`/goal`)는 "이 줄을 입력하여 복원" 안내로 격하. 주입 콘텐츠는 ultrathink/xhigh 활성을 **주장하지 않는다**(verification-claim-integrity §1.1). 신규 SSOT carve-out 도입 없음.
3. **Default mode**: handoff.yaml `mode` default = `manual` (auto-resume opt-in). 기본 UX 불변.

### A.3 실측으로 반증된 stale 가설 (research.md §C/§D)

- **registry merge**: "later handler의 additionalContext DROP" 주장은 stale. HEAD `97723664c`의 `mergeHandlerOutput`(registry.go:208-215)는 EVERY hook의 additionalContext를 `\n`-join 누적한다 → 신규 SessionStart 핸들러 공존 가능.
- **SessionStart matcher**: "live matcher = `startup|resume`(clear 미발화)" 주장은 stale. live + template 모두 이미 `startup|resume|clear|compact` → **settings.json 무변경**, `clear` source 이미 전달됨.

---

## §B — Scope

### B.1 In-scope

**M1 (config landing)**
1. NEW `HandoffConfig{Mode, Guide}` struct (`internal/config/types.go`) + Config 필드 + `handoffFileWrapper` (no `Consume` field — YAGNI, design.md §E.1)
2. NEW `NewDefaultHandoffConfig()` (`internal/config/defaults.go`, mode default `manual`) + `NewDefaultConfig` 등록
3. NEW `loadHandoffSection` (`internal/config/loader.go`, partial-override seed)
4. NEW audit registry 바인딩 `"handoff": "HandoffConfig"` (`internal/config/audit_registry.go`)
5. NEW 중립 `handoff.yaml` (template + live)

**M2 (save CLI)**
6. NEW `moai handoff save` — `handoff/pending.json` atomic 작성 (별도 경로)
7. NEW `moai handoff clear` — pending.json 수동 제거 (CLI, 즉시 제거)

**M3 (SessionStart 주입/소비)**
8. NEW `handoffInjectHandler` (EventSessionStart 3번째 핸들러) — `source==clear ∧ mode==auto`일 때만 주입+소비
9. NEW degrade-to-guidance 렌더 (i18n header, directive 안내화)
10. NEW claim-then-inject: atomic rename `pending.json → consumed/<ts>-<nonce>.json` 후 주입
11. NEW NULL session_id nonce fallback
12. NEW auto-mode stale TTL cleanup (`mode==auto` ∧ `saved_at` 초과 시 조용히 제거; manual mode는 stale라도 무접촉)

### B.2 Out of Scope (Exclusions)

WHAT NOT to build — 아래는 본 SPEC 범위 밖이다. 각 항목은 `handoff/` 트리 격리 또는 후속 SPEC 소관이다.

### Out of Scope — 기존 SessionEnd memory 흐름 변경
- `internal/hook/handoff/persist.go` `PersistIfPending` 및 `session-handoff/pending.md` 경로 — 무접촉. 본 SPEC은 별도 `handoff/` 트리만 다룬다.
- SessionEnd → memory 영속화 로직 재구현 또는 통합 — decoupled 유지.

### Out of Scope — settings.json / matcher 변경
- `.claude/settings.json` 및 `settings.json.tmpl`의 SessionStart matcher 수정 — HEAD에서 이미 `startup|resume|clear|compact` (research.md §D). 본 SPEC은 matcher를 assertion(grep 회귀 잠금)만 하고 변경하지 않는다.

### Out of Scope — 자율 run-phase 진입 / effort 조작
- 훅이 effort/model/thinking-mode를 실제로 변경 — 불가능. 주입은 안내 텍스트만(verification-claim-integrity §1.1).
- `/goal` line이 있어도 자율 run-phase 진입 승인 — Implementation Kickoff Approval human gate 불변.

### Out of Scope — M4 threshold-guidance
- 2-stage threshold 안내 + config 소비 완성 — M4 SPEC 소관. 본 SPEC은 HandoffConfig를 landing만 하고 threshold 로직은 구현하지 않는다.

### Out of Scope — 6-block 파싱/생성
- paste-ready 6-block 메시지의 구조적 파싱·생성 — orchestrator self-discipline (session-handoff.md). `moai handoff save`는 body를 opaque하게 저장.

### Out of Scope — consumed/ archive 관리
- consumed/ audit trail의 장기 archive/rotation — TTL cleanup(pending)만 in-scope. consumed/ 청소는 저우선 후속.

---

## §C — GEARS Requirements

> GEARS 표기(current). `<subject>`는 일반화된 명사(component/handler/CLI/config loader). 모든 REQ는 1:1 AC 추적(acceptance.md).

### M1 — HandoffConfig landing

#### REQ-AUTORESUME-001 (Ubiquitous) — HandoffConfig 구조 + manual default
The config loader **shall** expose a `HandoffConfig` with fields `Mode` (`"manual"` default | `"auto"`) and `Guide` (`false` default), where `NewDefaultHandoffConfig()` returns `Mode == "manual"`. (No `Consume` field — the sole consume source `clear` is a fixed semantic boundary, not a configurable value; see design.md §E.1 YAGNI rationale.)

#### REQ-AUTORESUME-002 (Where) — audit parity 바인딩
**Where** the yaml↔struct audit registry is consulted, the registry **shall** bind `"handoff"` to `"HandoffConfig"` so that `handoff.yaml` is not reported as an orphan section.

#### REQ-AUTORESUME-003 (Event-driven) — partial-override 로드
**When** `handoff.yaml` specifies a subset of keys (e.g. `mode` only), the config loader **shall** retain the construction-time default for the omitted key (`guide`) rather than collapsing it to its zero-value.

#### REQ-AUTORESUME-004 (Ubiquitous) — 중립 template
The `handoff.yaml` template artifact **shall** contain only language-neutral, internal-trace-free content (no SPEC IDs, no REQ tokens) per the template internal-content isolation doctrine.

### M2 — save CLI

#### REQ-AUTORESUME-005 (Event-driven) — save는 별도 경로에 atomic 작성
**When** `moai handoff save` is invoked with a resume body, the CLI **shall** write the pending record to `<projectDir>/.moai/state/handoff/pending.json` using an atomic temp-file-plus-rename write, and **shall not** read or write `<projectDir>/.moai/state/session-handoff/pending.md`.

#### REQ-AUTORESUME-006 (Ubiquitous) — pending.json 스키마
The `moai handoff save` command **shall** persist a JSON record carrying at least `schema_version`, `body` (verbatim resume), `directives` (ultrathink/ultracode/goal metadata), `conversation_language`, and `saved_at`.

#### REQ-AUTORESUME-007 (Event-driven) — clear CLI (M2)
**When** `moai handoff clear` is invoked, the CLI **shall** remove `handoff/pending.json` and **shall not** read or write `session-handoff/pending.md`. (The SessionStart auto-mode stale-TTL cleanup is a separate M3 concern — see REQ-AUTORESUME-019.)

### M3 — SessionStart 주입/소비

#### REQ-AUTORESUME-008 (Where While When) — 유일 소비 셀
**Where** `handoff/pending.json` exists **While** `cfg.Handoff.Mode == "auto"` **When** the SessionStart source is `clear`, the `handoffInjectHandler` **shall** inject the handoff content into `hookSpecificOutput.additionalContext` and consume the pending record; for every other (source, mode) combination the handler **shall not** consume the pending record.

#### REQ-AUTORESUME-009 (While) — manual mode pure no-op (stale 포함)
**While** `cfg.Handoff.Mode == "manual"`, the `handoffInjectHandler` **shall** perform no injection and **shall** preserve `handoff/pending.json` byte-unchanged — including when the pending record is stale (past its TTL); manual mode **shall never** remove or rename the pending record (pure no-op), preserving the unchanged default UX. This resolves the contradiction with the auto-only TTL cleanup of REQ-AUTORESUME-019.

#### REQ-AUTORESUME-010 (While) — non-clear source는 notice-only
**While** `cfg.Handoff.Mode == "auto"` and the SessionStart source is one of `startup`, `resume`, or `compact`, the handler **shall not** consume the pending record; **where** `cfg.Handoff.Guide == true` the handler **shall** emit a best-effort stderr hint only.

#### REQ-AUTORESUME-011 (Ubiquitous) — degrade-to-guidance (no unobserved claim)
The injected `additionalContext` **shall not** assert that `ultrathink`, `xhigh`, `ultracode`, or `/goal` is active; mode-change directives **shall** be rendered as manual-paste restoration guidance, honoring verification-claim-integrity §1.1.

#### REQ-AUTORESUME-012 (Event-driven) — claim-then-inject (atomic rename)
**When** the handler consumes the pending record, it **shall** first atomically rename `handoff/pending.json` to `handoff/consumed/<ts>-<nonce>.json`, and **shall** inject the content only after the rename succeeds; the handler **shall not** delete the memory `.md` audit copy nor the `consumed/` audit trail.

#### REQ-AUTORESUME-013 (Event-driven) — race 패자 fail-open (rename 실패 = 주입 생략)
**When** the atomic rename that claims `pending.json` fails for any reason (a concurrent session already claimed it, a permission error, or a platform-specific `MoveFileEx` failure), the handler **shall** skip injection and proceed normally without error. The rename-as-claim guarantee means at most one session injects; the losing session is detected by rename failure regardless of the specific errno (not `os.IsNotExist`-only, for cross-platform Windows `MoveFileEx` compatibility).

#### REQ-AUTORESUME-014 (Where) — NULL session_id nonce fallback
**Where** `saved_by_session` is empty (environment-fallback), the consumed filename nonce **shall** be derived from a `crypto/rand` value (with a deterministic `UnixNano` low-bits fallback if `crypto/rand` fails), producing a collision-safe filename without relying on the session id.

#### REQ-AUTORESUME-015 (Ubiquitous) — i18n 주입 header
The injection header and restoration-guidance text **shall** be rendered in the user's `conversation_language`, falling back to English for locales outside the {ko, en, ja, zh} set.

#### REQ-AUTORESUME-016 (Unwanted / event-detected) — 서브에이전트 경계
The `handoffInjectHandler` **shall not** invoke `AskUserQuestion` or `mcp__askuser` and **shall not** emit free-form user-directed questions; a static grep of `internal/hook/` for these tokens (excluding test files and comments) **shall** return zero matches.

#### REQ-AUTORESUME-017 (Ubiquitous) — fail-open
Every new hook and CLI path **shall** be best-effort: on any failure the SessionStart hook **shall** log via `slog.Warn("session_start: handoff: ...")` and return allow without blocking the session, matching the `persist.go` best-effort contract.

#### REQ-AUTORESUME-018 (Where) — additionalContext 공존
**Where** the existing `sessionStartHandler` and `autoUpdateHandler` also register for `EventSessionStart`, the registry's accumulate-all merge **shall** keep the `handoffInjectHandler`'s `additionalContext` (`\n`-joined), so that no handler's context is dropped.

#### REQ-AUTORESUME-019 (Where When) — auto-mode stale TTL cleanup (M3)
**Where** `cfg.Handoff.Mode == "auto"`, **when** a SessionStart consume-eligibility check observes a `handoff/pending.json` whose `saved_at` exceeds the stale TTL, the `handoffInjectHandler` **shall** silently remove the stale pending record (best-effort, no injection). This cleanup is auto-mode ONLY; in manual mode the handler is a pure no-op per REQ-AUTORESUME-009 (no removal even when stale). **Precedence (N1)**: auto-mode stale-cleanup (REQ-AUTORESUME-019) takes precedence over the notice-only hint (REQ-AUTORESUME-010) — when a pending record is stale, the handler performs the cleanup and suppresses the hint (there is nothing live to resume), regardless of `source` or `Guide`.

---

## §D — Traceability

REQ-AUTORESUME-001..019 ↔ AC-AUTORESUME-001..019 (acceptance.md §D). 각 AC는 binary pass/fail, test 참조 포함. (REQ-007 split → REQ-019 신설로 18→19; D3 Consume 필드 제거는 REQ 수에 영향 없음.)

## §E — Cross-References

- research.md (실측 근거 A~H), design.md (branch table §C.2, nonce §C.4, config §E, i18n §D)
- `internal/hook/handoff/persist.go` (SessionEnd 절반, PRESERVE)
- `internal/hook/registry.go` `mergeHandlerOutput` (accumulate-all)
- `internal/config/loader.go` `loadResearchSection` (미러 패턴)
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1
- `.claude/rules/moai/workflow/session-handoff.md` (6-block, Diet, Localization)
