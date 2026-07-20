---
id: SPEC-HANDOFF-THRESHOLD-001
title: "핸드오프 임계 완성 — 2단계 statusline suffix + HandoffConfig 소비 + context-usage.json 영속화 + Detection 독트린"
version: "0.1.0"
status: completed
created: 2026-07-06
updated: 2026-07-06
author: MoAI
priority: P2
phase: "v3.0.0"
module: "internal/statusline, internal/config"
lifecycle: spec-anchored
tags: "statusline, handoff-guide, context-window, threshold, two-stage, state-file, detection-heuristics, tier-m, epic-handoff-v2"
tier: M
era: V3R6
related_specs: [SPEC-HANDOFF-CTXGUIDE-001, SPEC-HANDOFF-MSGMODE-001, SPEC-HANDOFF-AUTORESUME-001]
---

# SPEC-HANDOFF-THRESHOLD-001 — 핸드오프 임계 완성 (Epic Handoff-v2 M4/4)

> Epic "Handoff-v2" **M4/4 (마지막 마일스톤)**. 선행: M1 `SPEC-HANDOFF-CTXGUIDE-001`(256K 밴드 로직, `completed`, origin `60db8e721`) · M2 `SPEC-HANDOFF-MSGMODE-001`(message-v2 mode-seed 독트린, `completed`, origin `97723664c`) · M3 `SPEC-HANDOFF-AUTORESUME-001`(HandoffConfig struct + `moai handoff save` CLI + SessionStart 주입/소비, `completed`, origin `b1ea0b9f9`). 본 SPEC은 M4로, **M1이 M4로 명시적으로 이연(defer)한 4개 작업**을 완성하고 M3이 landing한 `HandoffConfig`를 **소비**한다.

## HISTORY

- 2026-07-06: 초안 (draft). Tier M 5-artifact(spec/plan/acceptance/design/research) + progress.md §E skeleton. autoCompactThreshold 위치 실측(`internal/statusline/memory.go:16` `defaultAutoCompactPct=85` + `getAutoCompactThreshold()` 동일 패키지 → statusline 읽기 가능, open question 아님). 호출부 실측(`builder.Build`가 `input.SessionID` + `data.Memory` 동시 스코프). Detection 독트린 template mirror 존재 실측(Template-First 대상). 256K 행 M1 기존재 실측(중복 금지). LOCKED SCOPE 2건 준수(HandoffConfig 기존 필드만 소비 / 밴드 경계 하드코딩 defaults.go 상수).

---

## §A — Context (배경과 기존 상태)

### A.1 문제 — M1이 M4로 이연한 미완성 표면 4개

Epic Handoff-v2의 목표는 세션 경계를 넘는 컨텍스트 연속성이다. M1(`SPEC-HANDOFF-CTXGUIDE-001`)은 실사용 결함 1건(256K 윈도우에서 핸드오프 안내 영구 미표시)의 최소 수정에 한정했고, 나머지 4개 표면을 **명시적으로 M4로 이연**했다(M1 spec.md §1.3 Out of Scope):

1. **2단계 안내(D1)**: M1은 단일 단계 `(⚠️/clear)` suffix만 렌더. 2단계(`🛑/clear!` 하드 상한)는 미구현.
2. **HandoffConfig 소비(D2)**: M3이 `HandoffConfig{Mode, Guide}`를 landing만 하고, statusline은 config를 소비하지 않음.
3. **`.moai/state/context-usage.json` 영속화(D3)**: 상태 파일 미작성. Detection이 항상 휴리스틱.
4. **Detection Heuristics "state-file-first" 재작성(D4)**: 독트린이 바이트/reminder 휴리스틱만 서술.

### A.2 확정 사용자 결정 (LOCKED — 재론 금지)

1. **기존 `HandoffConfig{Mode, Guide}` 필드만 소비.** 신규 config 필드 도입 금지. 밴드 경계(50% / 90% / 95% / 500K 컷오프)는 config-override 불가 — `internal/config/defaults.go`에 **명명 상수**로 하드코딩(§14 inline magic number 금지). "M3 lands / M4 consumes" 분리 계약 보존.
2. **단일 M4 SPEC로 전체 확정 범위**(2단계 + config 소비 + 영속화 + 독트린). Tier M. Epic을 4-SPEC로 종료.

### A.3 실측 근거 (research.md §A~E)

- **autoCompactThreshold 위치**: `internal/statusline/memory.go:16` `const defaultAutoCompactPct = 85` + `getAutoCompactThreshold()`(env `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` 우선). renderer.go와 동일 패키지 → 하드 상한 공식이 직접 호출 가능. **open question 아님.**
- **호출부**: `internal/statusline/builder.go:138` `Build`가 `input`(→ `input.SessionID`) + `data`(→ `data.Memory.{ContextWindowSize,TokensUsed}`) 동시 스코프. `StatusData`에 SessionID 필드 없음 → 쓰기는 collector가 아닌 `Build`(collectAll 이후) 소관.
- **atomic write 선례**: `internal/statusline/model_cache.go` `WriteModelCache`(MkdirAll + temp+rename + silent-fail). 재사용, 신규 메커니즘 금지.
- **Detection 독트린 template mirror 존재**: `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`. D4는 Template-First(§2) 대상 — template 우선 편집 후 `make build`+sync. **task 전제("template 밖")는 오기이며 본 SPEC이 정정.**
- **256K 행 template drift (D1 정정)**: LIVE 독트린 § Context Window Targets 28행에 `Opus/Fable (256K) | 256,000 tokens | 90% | ~230,000 tokens` 존재(M1 추가, `grep -c '256,000'`=1). 그러나 **template mirror에는 부재**(`grep -c '256,000'`=0 — M1이 mirror를 동기화하지 않은 drift, 실측 확인). → D4는 (a) LIVE의 Detection Heuristics 절만 section-level 재작성(Targets 표 무접촉, 256K 중복 금지), (b) template mirror에 누락된 256K Targets 행을 **추가**(parity 회복). 두 사본 모두 `grep -c '256,000'`==1로 수렴. **full-file template→live overwrite / `moai update` full sync 금지**(LIVE 256K 행 삭제 회귀 경로).

---

## §B — Scope

### B.1 In-scope

**D1 — 2단계 statusline suffix**
1. NEW `handoffGuideStage(data) → {none, soft, hard}` 열거형 게이트 (`internal/statusline/renderer.go`)
2. `shouldShowHandoffGuide` = `stage != none` thin wrapper 유지 (M1 테스트/불변식 보존)
3. `renderBarsInline`: soft → `(⚠️/clear)`(M1 불변), hard → 2단계 마커 `(🛑/clear!)`
4. 하드 상한 = `min(HandoffHardCeilingCapPct, getAutoCompactThreshold() + HandoffHardCeilingMarginPct)`, `hard < soft-threshold` 시 soft로 clamp

**D1 상수 (config/defaults.go, §14)**
5. NEW 명명 상수: `HandoffSoftLargePct(50)` · `HandoffSoftStandardPct(90)` · `HandoffLargeWindowCutoff(500_000)` · `HandoffHardCeilingCapPct(95)` · `HandoffHardCeilingMarginPct` — inline literal 제거

**D2 — HandoffConfig 소비 reconciliation (M1 무회귀 불변식)**
6. statusline suffix(양 단계) + state-file write는 HandoffConfig 무관(무조건) — M1 불변식
7. `Guide`/`Mode` 소비 경계를 design.md에서 화해(reconcile): statusline은 config 미소비, `Mode`는 M3 auto-resume 유지, `Guide`는 advisory 경로 한정

**D3 — context-usage.json 영속화**
8. NEW `writeContextUsage` (model_cache.go 패턴 재사용, atomic temp+rename, best-effort)
9. `builder.Build`에서 호출(collectAll 이후) — session_id + Memory 동시 스코프
10. write-if-changed throttle (payload 불변 시 skip, render-rate 디스크 churn 방지)
11. session_id 스탬프 + guard semantics (last-writer-wins snapshot; reader는 session_id 불일치 시 stale)
12. fallback-UUID dead-path 처리 (session_id 부재 시 freshness 기반 유효성 — single-session 공통 경로 유지)
13. `writer_pid` discriminator (concurrent same-checkout empty-`session_id` 세션이 서로의 스냅샷을 오재개하지 못하도록; freshness 헬퍼가 writer_pid 일치 요구 — D2 hole 정정)

**D4 — Detection Heuristics 독트린 재작성**
13. `context-window-management.md` § Detection Heuristics를 "state-file-first"로 재작성 (state file 우선, 부재/stale/파싱실패 시 기존 휴리스틱 폴백) — **Detection Heuristics 절만 section-level 편집**(full-file overwrite / `moai update` full sync 금지)
14. Template-First: template mirror + live에 **동일 Detection 절** section-level 편집; template mirror에 누락된 256K Targets 행 **추가**(D1 drift parity 회복). LIVE는 256K 중복 금지, template 중립성 §25

### B.2 Out of Scope (Exclusions)

WHAT NOT to build — 아래는 본 SPEC 범위 밖이다.

### Out of Scope — 신규 config 필드 / 밴드 override
- `HandoffConfig`에 신규 필드 추가(예: `context_thresholds`, `stage2_pct`) — LOCKED SCOPE #1 위반. 기존 `{Mode, Guide}`만 소비.
- 밴드 경계(50/90/95/500K)를 config로 사용자 override 가능하게 노출 — 하드코딩(defaults.go 상수) 고정. "M3 lands / M4 consumes" 계약 보존.

### Out of Scope — M1/M3 표면 재구현
- M1 `shouldShowHandoffGuide` 밴드 로직(≥500K→50% / <500K→90%) 재설계 — soft 단계로 그대로 승계, 무변경.
- M3 `handoffInjectHandler` auto-resume 소비 경로(`source==clear ∧ mode==auto`) 변경 — 무접촉.
- M3 `moai handoff save`/`clear` CLI, `handoff/pending.json` 흐름 — 무접촉(context-usage.json은 별도 파일).

### Out of Scope — statusline suffix의 config 게이팅
- statusline `(⚠️/clear)`/`(🛑/clear!)` suffix를 `guide==true`에 게이팅 — **금지**(guide default false → suffix 소멸 = M1 회귀). suffix는 무조건 렌더.
- state-file write를 Mode/Guide로 게이팅 — 금지(무조건 write).

### Out of Scope — 신규 advisory 훅 코드
- M3 SessionStart handler에 state-file 읽기 advisory 훅 신규 추가 — Guide advisory는 **독트린(D4) 서술**로 처리(orchestrator 행동 지침), 신규 Go 훅 코드 도입 금지(M1 불변식 보호 + scope 최소).

### Out of Scope — cross-session 조정 / 세션 레지스트리 통합
- `.moai/state/active-sessions.json`(multi-session coordination) 통합, `moai session` 레지스트리 연동 — 후속 소관. context-usage.json은 독립 스냅샷.
- state file 장기 rotation/archive — best-effort overwrite만.

### Out of Scope — autoCompact 런타임 제어
- Claude Code auto-compact 임계(85%) 자체를 변경/제어 — 런타임 소관. 본 SPEC은 값을 **읽어** 하드 상한을 계산할 뿐.

---

## §C — GEARS Requirements

> GEARS 표기(current). `<subject>`는 일반화된 명사(gate / renderer / writer / doctrine). 모든 REQ는 1:1 이상 AC 추적(acceptance.md §D).

### D1 — 2단계 statusline suffix

#### REQ-THRESHOLD-001 (Ubiquitous) — 단계 열거형 게이트
The handoff-guide gate **shall** expose a stage classifier `handoffGuideStage(data) → {none, soft, hard}` replacing the bare-bool decision, and **shall** retain `shouldShowHandoffGuide(data)` as `stage != none` so that the M1 threshold behavior is byte-preserved.

#### REQ-THRESHOLD-002 (State-driven) — 2단계 렌더
**While** the gate returns `soft`, the renderer **shall** append `(⚠️/clear)` (the M1 suffix, unchanged); **while** the gate returns `hard`, the renderer **shall** append the distinct stage-2 marker `(🛑/clear!)`; **while** the gate returns `none`, the renderer **shall** append no suffix.

#### REQ-THRESHOLD-003 (Where) — auto-compact-aware 하드 상한
**Where** the gate computes the hard ceiling, it **shall** use `min(HandoffHardCeilingCapPct, getAutoCompactThreshold() + HandoffHardCeilingMarginPct)`, and **shall** clamp the ceiling up to the band's soft threshold when the computed ceiling would fall below it (degenerate override configs → stage-2 collapses onto the soft threshold; only `hard` is then shown for that band).

#### REQ-THRESHOLD-004 (Where) — 밴드 경계 명명 상수
**Where** the gate reads band boundaries, it **shall** reference named constants in `internal/config/defaults.go` (`HandoffSoftLargePct`, `HandoffSoftStandardPct`, `HandoffLargeWindowCutoff`, `HandoffHardCeilingCapPct`, `HandoffHardCeilingMarginPct`) and **shall not** carry inline magic-number literals for these boundaries in `renderer.go` (§14 hardcoding-prevention).

#### REQ-THRESHOLD-005 (Unwanted / documentation) — reachability 한계 명시
The doctrine and CHANGELOG **shall** state the stage-2 reachability limitation: because runtime auto-compact fires near `getAutoCompactThreshold()`% of the raw window, a hard ceiling at or above that point is frequently pre-empted (rarely fires) — an intentional, documented tradeoff of the auto-compact-aware formula. The doctrine **shall not** claim stage-2 always fires.

### D2 — HandoffConfig 소비 reconciliation

#### REQ-THRESHOLD-006 (Ubiquitous — INVARIANT, no M1 regression) — suffix는 config 무관
The statusline handoff suffix (both `soft` and `hard`) **shall** render as a pure function of context-window usage, independent of `HandoffConfig`; with the shipped defaults (`Mode == "manual"`, `Guide == false`) the `soft` suffix **shall** still render whenever `handoffGuideStage != none`. The renderer **shall not** gate the suffix on `Guide` (gating on `Guide == true` would make the suffix vanish by default — a regression from M1's shipped unconditional behavior).

#### REQ-THRESHOLD-007 (Ubiquitous) — state-file write는 config 무관
The `context-usage.json` write **shall** be unconditional (never gated by `HandoffConfig.Mode` or `HandoffConfig.Guide`), so that Detection has an authoritative snapshot regardless of auto-resume opt-in.

#### REQ-THRESHOLD-008 (Where) — Guide/Mode 소비 경계 화해
**Where** `HandoffConfig` is consumed by the Handoff-v2 stack, `Mode` **shall** retain its M3 auto-resume semantics (SessionStart handler) and `Guide` **shall** gate only the advisory / notice path (the D4 doctrine's state-file-derived advisory), and neither **shall** gate the statusline suffix or the state-file write; the statusline package **shall not** read `HandoffConfig`. design.md **shall** document this reconciliation.

### D3 — context-usage.json 영속화

#### REQ-THRESHOLD-009 (Event-driven) — atomic write (선례 재사용)
**When** the statusline builds a line and `data.Memory.Available` is true, the writer **shall** persist `<projectDir>/.moai/state/context-usage.json` using the atomic temp-file-plus-rename pattern of `model_cache.go` `WriteModelCache` (MkdirAll + write-temp + rename), best-effort with silent failure (no error surfaced, no statusline disruption).

#### REQ-THRESHOLD-010 (Ubiquitous) — state-file 스키마
The `context-usage.json` record **shall** carry at least `schema_version`, `session_id`, `writer_pid`, `captured_at`, `context_window_size`, `tokens_used`, `raw_pct`, `stage`, and `band`. (`writer_pid` = the writing process identity, the discriminator that distinguishes concurrent empty-`session_id` writers per REQ-THRESHOLD-018.)

#### REQ-THRESHOLD-011 (Where) — 호출부 배치 (session_id + Memory 동시 스코프)
**Where** the write is invoked, it **shall** occur where both the session id (from stdin `session_id`) and `Memory` (`context_window_size` / `tokens_used`) are in scope — i.e., in `builder.Build` after `collectAll`, **not** at a call-site lacking `session_id` (`collectAll` does not capture `session_id` into `StatusData`).

#### REQ-THRESHOLD-012 (Event-driven) — write-if-changed throttle
**When** the semantic payload (`session_id`, `stage`, integer-rounded `raw_pct`, `context_window_size`) is byte-equal to the on-disk record, the writer **shall** skip the write (write-if-changed), so that render-rate invocations do not cause disk churn.

#### REQ-THRESHOLD-013 (Where) — session_id guard semantics
**Where** the record is written, the writer **shall** stamp the current `session_id` as a last-writer-wins snapshot, and a reader **shall** treat a record whose `session_id` differs from the current session as stale (fall back to heuristics) — preventing a side-channel last-writer-wins false-match from resuming the wrong session's context.

#### REQ-THRESHOLD-014 (Where) — fallback-UUID dead-path 처리
**Where** `session_id` is empty or the environment-fallback sentinel (no real UUID), the writer **shall** still write the record, and a reader **shall** validate the record by `captured_at` freshness instead of `session_id` equality — so the primary state-file path stays live for the common single-session case (it **shall not** degrade to heuristics-always whenever the runtime does not expose a UUID).

### D4 — Detection Heuristics 독트린 재작성

#### REQ-THRESHOLD-015 (Ubiquitous) — state-file-first 재작성
The `context-window-management.md` § Detection Heuristics **shall** be rewritten state-file-first: read `.moai/state/context-usage.json` first (authoritative raw context usage + stage), and fall back to the existing byte / system-reminder / tool-result / Agent()-count heuristics **only when** the file is absent, stale (`session_id` mismatch or `captured_at` freshness-expired), or unparseable.

#### REQ-THRESHOLD-016 (Where) — Template-First + section-level edit
**Where** the doctrine is edited, the change **shall** apply a **section-level edit of the Detection Heuristics block only** to BOTH the template source (`internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`, kept language-neutral and internal-trace-free per §25) AND the live copy (`.claude/rules/moai/workflow/context-window-management.md`), **shall** run `make build`, and **shall not** perform a full-file template→live overwrite / `moai update` full sync (which would delete the live 256,000-window row — an M1 regression). The live § Context Window Targets table **shall not** gain a duplicate 256,000 row.

#### REQ-THRESHOLD-017 (Where) — template mirror 256K parity (D1 drift 정정)
**Where** D4 completes, the template mirror **shall** carry the `256,000`-window row in § Context Window Targets at parity with the live copy — M1 added it to live only, leaving the mirror missing it (a confirmed drift). Post-D4, `grep -c '256,000'` **shall** return 1 for BOTH the template mirror AND the live copy (drift eliminated).

#### REQ-THRESHOLD-018 (Where) — concurrent empty-`session_id` discriminator
**Where** two concurrent same-checkout sessions both lack a real `session_id` (UUID) and share one `context-usage.json`, the writer **shall** stamp `writer_pid` and the freshness helper **shall**, for an empty-`session_id` record, return own-session-valid only when `captured_at` is fresh AND the record's `writer_pid` matches the reader-supplied expected writer identity — so `captured_at` freshness alone **shall not** let session B read session A's snapshot as its own. This closes the B4 cross-contamination hole re-opened by REQ-014's empty-id fallback; the mechanical guard is exercised by the Go freshness helper, and the doctrine-only reader's residual limitation is documented (acceptance §D.4 / design §D.4).

---

## §D — Traceability

REQ-THRESHOLD-001..018 ↔ AC-THRESHOLD-001..018 (acceptance.md §D), 1:1. 각 AC는 binary pass/fail + test/grep 참조. **AC-THRESHOLD-006이 "no M1 statusline regression" 불변식 AC.** REQ-017(template mirror 256K parity — D1 drift 정정) + REQ-018(concurrent empty-`session_id` discriminator — D2 hole 정정)은 plan-auditor iter-1 0.84 SHOULD-FIX 반영으로 신설(16→18).

## §E — Cross-References

- research.md §A~E (실측 근거), design.md §A~F (stage enum / 하드 상한 공식 / state-file 스키마 / guide-semantics 화해 / 호출부)
- `internal/statusline/renderer.go` `shouldShowHandoffGuide`(578행), `renderBarsInline`(316행 `(⚠️/clear)`)
- `internal/statusline/memory.go` `getAutoCompactThreshold`(39행) / `defaultAutoCompactPct=85`(16행)
- `internal/statusline/builder.go` `Build`(138행) / `collectAll`(187행)
- `internal/statusline/model_cache.go` `WriteModelCache`(atomic 선례)
- `internal/config/types.go:601` `HandoffConfig{Mode, Guide}` / `internal/config/defaults.go` `DefaultHandoffMode`
- `internal/hook/handoff_inject.go` `handoffConfig`(146행, Guide 소비)
- `.claude/rules/moai/workflow/context-window-management.md` § Detection Heuristics(69행) + template mirror
- CLAUDE.local.md §2(Template-First) / §14(하드코딩 방지) / §25(template 중립성)
- M1 `SPEC-HANDOFF-CTXGUIDE-001` §1.3(M4 이연 4항목) — 본 SPEC이 완성
