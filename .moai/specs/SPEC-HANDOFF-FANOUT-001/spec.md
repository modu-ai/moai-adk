---
id: SPEC-HANDOFF-FANOUT-001
title: "Mode 4 (parallel-subagents) paste-time fan-out steering phrase — resume message directive coupling"
version: "0.1.1"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
author: GOOS행님
priority: P2
phase: "vNext"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tier: S
related_specs: [SPEC-HANDOFF-MSGMODE-001]
tags: "handoff, orchestration, mode-seed, session-resume, doctrine"
---

# SPEC-HANDOFF-FANOUT-001 — Mode 4 fan-out steering phrase (paste-time directive coupling)

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.1 | 2026-07-08 | Plan-audit 반영: D1(AC-HFO-004b vacuous → baseline-delta 규범 토큰, 004a/c 동일 클래스 경화) + D2(AC-HFO-011 locale-verbatim 기계 AC 추가) + D3(§B/§C 인용부호 의역 정정) + D4(AC-HFO-003 `-Flc` → `-Fl`). |
| 0.1.0 | 2026-07-08 | Plan-phase 저작 (draft). Tier S, 2 body artifacts + progress.md. |

## §A. Context

### A.1 계보 (lineage)

본 SPEC은 mode-seed doctrine (Epic Handoff-v2의 SPEC-HANDOFF-MSGMODE-001)의 후속이다. session-handoff.md (SSOT)의 Block 1 `mode:` orchestration-seed 라인은 Phase 0.95 mode catalog (orchestration-mode-selection.md §A)에 1:1 매핑되며, directive-coupling 표는 각 seed 값에 paste-time 트리거 장치를 결합한다:

- `agent-team` (Mode 3) → Block 5 run command에 `--team` append
- `dynamic-workflow` (Mode 6) → opener 라인에 bare `ultracode` append
- `parallel-subagents` (Mode 4) → **coupling 없음 (`—`) — 본 SPEC이 메우는 갭**
- `solo-sequential` (Mode 5) → 생략 (v1 byte-identical)

### A.2 갭 (why this matters)

Mode 4는 non-default seed 중 유일하게 paste-time 트리거 장치가 없어, 다음 세션 orchestrator의 `mode:` 메타데이터 파싱에 전적으로 의존한다. `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7+ Prompt Philosophy **Principle 4** (실측 L50): Opus 4.7+/4.8은 기본적으로 subagent를 덜 spawn하며, fan-out은 명시적으로 지시되어야 한다("This behavior is steerable: when fan-out helps, instruct explicitly").

핵심 메커니즘: paste-ready resume message는 **사용자가 직접 붙여넣는** 메시지이므로, 그 안의 문구는 Claude Code 런타임 계층에서 사용자 발화 지시어로 계수된다. 즉 `fan out subagents` 문구가 pasted message에 등장하면 user-authored explicit multi-agent opt-in으로 작동한다 — `ultrathink`/`ultracode` 키워드와 동일한 paste-time 클래스.

## §B. Requirements (GEARS)

### REQ-HFO-001 — Mode 4 directive coupling (fan-out steering phrase)

**Where** the seeded orchestration mode is `parallel-subagents`, **When** the orchestrator emits a paste-ready resume message, the resume message **shall** append a natural-language fan-out steering phrase to the Block 1 opener line — canonical form:

```
fan out subagents (<read-only investigation scope>)
```

appended after the `ultrathink. <SPEC-ID> <phase> 진입.` opener text.

The doctrine **shall** define `fan out subagents` as a **locale-verbatim protocol phrase** — preserved in English across all locales (the `mode:` value들과 동일 클래스; localization table에 행을 추가하지 않는다). The parenthesized scope qualifier **shall** translate per `conversation_language`.

### REQ-HFO-002 — invariant preservation (3 clauses)

- **(a) SEED-not-permission**: The doctrine **shall** state that the fan-out steering phrase does NOT authorize autonomous run-phase entry — Implementation Kickoff Approval remains mandatory. 기존 `mode:` / bare `ultracode` / `/goal` 조항과 동일한 binding 문구 강도를 유지한다.
- **(b) 3-5 concurrent ceiling**: The doctrine **shall** state that fan-out spawns respect the Anthropic 3-5 concurrent `Agent()` ceiling (`orchestration-mode-selection.md` §C.2; ceiling이 Mode 3과 Mode 4에 동등 적용된다는 조항 — 의역이며 verbatim 인용 아님, 실측 L133. run-phase는 이 문자열을 grep 앵커로 사용하지 말 것).
- **(c) Read-only scoping**: The fan-out phrase **shall** carry a read-only investigation scope qualifier. The doctrine **shall not** seed parallel WRITE fan-out via this phrase (consistent with `agent-common-protocol.md` § Background Agent Execution write restrictions).

### REQ-HFO-003 — disambiguation note (Mode 4, NOT Mode 3)

The doctrine **shall** state that the Claude Code UI tip — "Say 'fan out subagents' and Claude sends a team" — maps to **Mode 4** (parallel subagents: 단일 턴 multi-`Agent()` spawn), **NOT Mode 3** (agent-team: env prerequisite `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` + `workflow.team.enabled` + `--team` coupling 보유). 이 진술이 없으면 UI tip 문구가 Mode 3 팀 형성으로 overclaim될 수 있다.

### REQ-HFO-004 — surface parity (4 files + make build)

**When** the run-phase implements this SPEC, the implementer **shall** update all 4 surfaces consistently:

| # | Surface | Touch points |
|---|---------|-------------|
| i | `.claude/rules/moai/workflow/session-handoff.md` (SSOT) | directive-coupling 표 Mode 4 행 (coupling cell `—` → fan-out coupling), Block 1 skeleton `mode:` 주석 라인, Field-by-Field Block 1 directive-binding prose, § Anti-Patterns 엔트리 (REQ-HFO-006), Pre-emit self-check (paste-ready budget) 9 → **10 items** |
| ii | `.claude/output-styles/moai/moai.md` §8 (render surface) | Block 1 skeleton `mode:` 주석 라인, compact mapping prose (Mode 4 coupling 동일 canonical 문구), Pre-emit self-check (11 items) → **(12 items)** — concern-name qualifier parity (drift-mitigation sentinel, 양 파일 공통) 유지 |
| iii | `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` | (i)과 동일 편집 — 단 REQ-HFO-005 중립성 준수 |
| iv | `internal/template/templates/.claude/output-styles/moai/moai.md` | (ii)와 동일 편집 — 단 REQ-HFO-005 중립성 준수 |

이후 `make build`로 template embed를 재생성한다 (exit 0).

Parity 계약: SSOT 표의 Mode 4 coupling 텍스트와 render surface compact mapping의 Mode 4 coupling 텍스트는 동일한 canonical 문구 `fan out subagents (<read-only investigation scope>)`를 담아야 한다.

### REQ-HFO-005 — template neutrality

The template mirror edits **shall not** introduce internal SPEC IDs, REQ tokens, audit citations, internal dates, or commit SHAs (CI guard: `template-neutrality-check.yaml` + `internal_content_leak_test.go`). Template에 추가되는 doctrine 텍스트는 generic해야 한다. 본 SPEC ID (`SPEC-HANDOFF-FANOUT`) 및 `REQ-HFO` 토큰의 template 유입은 0건이어야 한다.

### REQ-HFO-006 — anti-pattern entry

The SSOT § Anti-Patterns **shall** gain an entry: `mode: parallel-subagents`가 seed되었는데 fan-out steering phrase를 생략하는 경우 — resumed Opus 4.8 세션이 Principle 4에 따라 조용히 under-spawn한다 (fan-out은 `ultrathink.` opener로 복원되지 않으며 명시적 지시가 필요).

## §C. Measured Baseline Anchors (2026-07-08 실측)

Line 번호는 drift 가능 — run-phase에서는 content-token 앵커를 우선 사용할 것.

| Anchor | Measured value |
|---|---|
| SSOT directive-coupling 표 Mode 4 행 | L85: `\| parallel-subagents \| Mode 4 (parallel, 3-5 concurrent Agent()) \| emitted \| — \|` — coupling cell `—` (갭 확인) |
| SSOT Block 1 skeleton `mode:` 주석 | L33 (dynamic-workflow→ultracode, agent-team→--team 결합만 존재) |
| SSOT Pre-emit self-check | L282: `### Pre-emit self-check (paste-ready budget) — 9 items` |
| render moai.md §8 skeleton / compact mapping / pre-emit | L682 / L701 / L722: `Pre-emit self-check (11 items)` |
| `grep -rc 'fan out subagents'` (4개 표면) | **0 matches** (baseline) |
| Template mirrors | 양 파일 모두 live와 **byte-identical** (`diff -q` exit 0, 2026-07-08) — full mirror, subset 아님 |
| 3-5 ceiling | orchestration-mode-selection.md §C.2 (L127) + L133 — ceiling의 Mode 3/Mode 4 동등 적용 조항 (의역 — grep 앵커로 사용 금지) |
| Principle 4 | moai-constitution.md L50 "Fewer subagents spawned by default … instruct explicitly" |

## §D. Exclusions — Out of Scope

본 SPEC이 다루지 않는 것 (out of scope):

### Out of Scope — Runtime / Go 코드 변경

- Go 소스 변경 없음 — doctrine-only SPEC. `make build`는 template embed 재생성 목적일 뿐 Go 코드 수정이 아니다.
- 신규 hook, 신규 lint rule, JSON twin (`schema_version: 2`) 구현 없음 — JSON twin forward-compat note는 기존 doctrine 그대로 유지.

### Out of Scope — 타 모드 coupling 변경

- Mode 3 `--team` coupling, Mode 6 bare `ultracode` coupling, Mode 5 omission default는 무접촉.
- `/goal` re-set 라인의 emit 조건 및 Block 1 line-order invariant 자체의 재설계 없음 (fan-out 문구는 opener 라인에 ride하므로 line-order 표 변경 불요).

### Out of Scope — Write fan-out seeding

- 병렬 WRITE fan-out을 seed하는 어떤 문구/장치도 도입하지 않음 (REQ-HFO-002(c)의 반대면). Write 작업은 기존 foreground-sequential 정책 유지.

### Out of Scope — Phase 0.95 선택 로직 변경

- auto-select threshold (domains ≥ 3 / files ≥ 10 / score ≥ 7), mode catalog 자체, seed 파생 로직 변경 없음 — 본 SPEC은 이미 seed된 `parallel-subagents` 값의 paste-time 표현만 추가한다.

## §E. Cross-references

- `.claude/rules/moai/workflow/session-handoff.md` — SSOT (편집 대상 i)
- `.claude/output-styles/moai/moai.md` §8 — render surface (편집 대상 ii)
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §A / §C.2 — mode catalog + 3-5 ceiling
- `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7+ Prompt Philosophy Principle 4
- `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution — read-only scoping 근거
- SPEC-HANDOFF-MSGMODE-001 — mode-seed doctrine 전신 (Epic Handoff-v2 M2)
