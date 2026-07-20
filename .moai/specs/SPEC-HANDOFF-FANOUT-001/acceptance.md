---
id: SPEC-HANDOFF-FANOUT-001
title: "Mode 4 (parallel-subagents) paste-time fan-out steering phrase — resume message directive coupling"
version: "0.1.1"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: GOOS행님
priority: P2
phase: "vNext"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tier: S
tags: "handoff, orchestration, mode-seed, session-resume, doctrine"
---

# SPEC-HANDOFF-FANOUT-001 — Acceptance Criteria

모든 AC는 기계적 명령(grep / diff / make)으로 관찰 가능하다. 주관 판단 AC 없음. Baseline: `fan out subagents` 4개 표면 전부 0 matches, 양 mirror byte-identical (2026-07-08 실측 — spec.md §C).

## §A. Traceability Matrix

| REQ | AC(s) | Observation |
|---|---|---|
| REQ-HFO-001 (coupling + canonical form + locale-verbatim) | AC-HFO-001a/b, AC-HFO-002a/b, AC-HFO-003, AC-HFO-011 | grep count ≥ 2 per surface; canonical literal 4-표면 존재; locale-verbatim 선언 baseline-delta grep |
| REQ-HFO-002 (invariant 3 clauses) | AC-HFO-004a/b/c | whole-file baseline-delta grep — canonical literal에 포함되지 않는 규범 토큰만 사용 (vacuous 방지) |
| REQ-HFO-003 (disambiguation) | AC-HFO-005 | UI tip 인용 + Mode 4 매핑 grep |
| REQ-HFO-004 (surface parity + make build) | AC-HFO-003, AC-HFO-006a/b, AC-HFO-009, AC-HFO-010 | canonical literal parity, pre-emit counts, mirror diff, build exit |
| REQ-HFO-005 (template neutrality) | AC-HFO-008 | 0 matches |
| REQ-HFO-006 (anti-pattern entry) | AC-HFO-007 | 섹션 내 grep ≥ 1 |

## §B. Given-When-Then Scenarios

### Scenario 1 — Mode 4 seed → fan-out phrase 방출 (happy path)

- **Given** 다음 SPEC이 read-only 다중 도메인 조사를 요구하고 Phase 0.95 seed가 `parallel-subagents`인 multi-SPEC Epic 세션 종료 시점,
- **When** orchestrator가 paste-ready resume message를 방출하면,
- **Then** Block 1 opener 라인은 `ultrathink. <SPEC-ID> <phase> 진입. fan out subagents (<read-only 조사 범위>)` 형태를 갖고, `mode: parallel-subagents` 라인이 그 아래 위치하며,
- **And** 다음 세션(Opus 4.8)에서 pasted phrase가 user-authored multi-agent opt-in으로 계수되어 3-5개 read-only `Agent()` fan-out이 명시적으로 steer된다,
- **And** run-phase 진입은 여전히 Implementation Kickoff Approval을 요구한다 (문구는 permission grant가 아님).

### Scenario 2 — 타 모드 seed → 문구 미방출 (no cross-contamination)

- **Given** seed가 `solo-sequential`(생략) / `agent-team` / `dynamic-workflow` 중 하나일 때,
- **When** resume message가 방출되면,
- **Then** `fan out subagents` 문구는 등장하지 않고, 기존 coupling(`--team` append / bare `ultracode` append / v1 byte-identical omission)만 각각 유지된다.

### Scenario 3 — locale-verbatim (Korean locale)

- **Given** `conversation_language: ko` 세션에서 `parallel-subagents` seed,
- **When** resume message가 방출되면,
- **Then** `fan out subagents`는 영어 verbatim으로 보존되고, 괄호 안 scope qualifier만 한국어로 렌더된다 (예: `fan out subagents (read-only 코드베이스 조사)`).

## §C. Acceptance Criteria (grep-verifiable)

작업 디렉터리: `/Users/goos/MoAI/moai-adk-go`.

| AC | Command | Expected |
|---|---|---|
| AC-HFO-001a | `grep -c 'fan out subagents' .claude/rules/moai/workflow/session-handoff.md` | ≥ 2 (표 행 + Field-by-Field prose; baseline 0) |
| AC-HFO-001b | `grep -c 'fan out subagents' internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` | ≥ 2 |
| AC-HFO-002a | `grep -c 'fan out subagents' .claude/output-styles/moai/moai.md` | ≥ 2 (skeleton 주석 + compact mapping) |
| AC-HFO-002b | `grep -c 'fan out subagents' internal/template/templates/.claude/output-styles/moai/moai.md` | ≥ 2 |
| AC-HFO-003 | `grep -Fl 'fan out subagents (<read-only investigation scope>)' <위 4개 파일>` | 4개 파일명 전부 출력 (canonical form parity — SSOT 표와 render compact mapping이 동일 문구. 주의: darwin에서 `-l`은 `-c`를 무시하므로 `-Flc` 조합 금지 — 파일명 출력으로 판정) |
| AC-HFO-004a | `grep -c 'Implementation Kickoff Approval' .claude/rules/moai/workflow/session-handoff.md` | ≥ 4 (baseline 3, 2026-07-08 실측 — fan-out 조항의 SEED-not-permission binding이 신규 추가되어야만 증가) |
| AC-HFO-004b | `grep -ci 'write fan-out' .claude/rules/moai/workflow/session-handoff.md` | ≥ 1 (baseline 0 — WRITE fan-out 금지 조항의 규범 토큰. canonical literal의 `<read-only investigation scope>`에 포함되지 않는 토큰이므로 비-vacuous; read-only scope qualifier 자체는 AC-HFO-003 literal이 검증) |
| AC-HFO-004c | `grep -ci 'ceiling' .claude/rules/moai/workflow/session-handoff.md` | ≥ 2 (baseline 1, 2026-07-08 실측 — 3-5 ceiling 조항이 fan-out 조항에 신규 추가되어야만 증가. `3-5` 토큰은 기존 directive-coupling 표 행에 이미 존재해 vacuous이므로 grep 토큰으로 사용 금지) |
| AC-HFO-005 | `grep -A3 "sends a team" .claude/rules/moai/workflow/session-handoff.md \| grep -c 'Mode 4'` | ≥ 1 (UI tip → Mode 4 매핑, NOT Mode 3 진술 동반) |
| AC-HFO-006a | `grep -c 'Pre-emit self-check (paste-ready budget) — 10 items' .claude/rules/moai/workflow/session-handoff.md` | = 1 (9→10; template mirror에도 동일 = 1) |
| AC-HFO-006b | `grep -c 'Pre-emit self-check (12 items)' .claude/output-styles/moai/moai.md` | = 1 (11→12; template mirror에도 동일 = 1) |
| AC-HFO-007 | `sed -n '/^## Anti-Patterns/,/^## /p' .claude/rules/moai/workflow/session-handoff.md \| grep -c 'fan out subagents\|fan-out steering'` | ≥ 1 (생략 anti-pattern 엔트리) |
| AC-HFO-008 | `grep -rn 'SPEC-HANDOFF-FANOUT\|REQ-HFO' internal/template/templates/` | 0 matches (template neutrality) |
| AC-HFO-009 | `diff -q .claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md && diff -q .claude/output-styles/moai/moai.md internal/template/templates/.claude/output-styles/moai/moai.md` | exit 0 (양 mirror byte-identical 유지 — baseline identical 실측) |
| AC-HFO-010 | `make build` | exit 0 (template embed 재생성) |
| AC-HFO-011 | `grep -c 'locale-verbatim' .claude/rules/moai/workflow/session-handoff.md` | ≥ 2 (baseline 1, 2026-07-08 실측 — 기존 `mode:` locale-verbatim 선언 1건에 더해 fan-out phrase의 locale-verbatim 선언이 신규 추가되어야만 증가; REQ-HFO-001 locale-verbatim 하위 조항의 기계 검증. Scenario 3는 서사적 보조일 뿐 판정 기준 아님) |

> AC sub-ID 표기(001a/001b 등)는 acceptance.md 내부의 paired sub-criteria 관례이며 SPEC ID 규칙과 무관하다.

## §D. Edge Cases

1. **AC-HFO-009 재베이스라인 조건**: run-phase 착수 전 병렬 세션이 mirror를 diverge시킨 경우(baseline diff ≠ 0), byte-parity AC는 재실측 후 progress.md §E.2에 재베이스라인을 기록하고 "본 SPEC이 추가한 텍스트에 한정된 parity"로 축소 해석한다.
2. **`/goal` 동시 존재**: fan-out 문구는 opener 라인에 ride하므로 Block 1 line-order invariant(opener → `mode:` → `# /goal`)와 충돌하지 않는다 — line-order 표 변경 없음이 정상.
3. **baseline-delta 재실측 조건**: AC-HFO-004a/c 및 AC-HFO-011은 2026-07-08 실측 baseline(각 3/1/1) 대비 증가(delta ≥ +1)를 판정한다. run-phase 착수 전 병렬 세션이 SSOT를 편집해 baseline이 이동한 경우, run-phase는 착수 시점에 동일 grep을 재실측해 threshold를 재계산하고 progress.md §E.2에 재베이스라인을 기록한다 (delta ≥ +1 판정 자체는 불변).
4. **neutrality vs parity 양립**: AC-HFO-008(내부 토큰 0)과 AC-HFO-009(byte-identical)는 양립해야 한다 — live 표면에 추가하는 doctrine 텍스트 자체를 generic하게 작성하면 동일 텍스트가 template에 그대로 들어가도 중립성이 성립한다. live에만 SPEC ID를 넣고 template에서 빼는 방식은 AC-HFO-009를 깨므로 금지.

## §E. Quality Gates + Definition of Done

- **Quality gate**: doctrine-only SPEC — Go 테스트 신규 작성 없음. 게이트 = §C grep 전 항목 + `make build` exit 0 + template-neutrality CI green.
- **DoD**:
  - [ ] AC-HFO-001a ~ AC-HFO-011 전 항목 PASS (검증 명령 출력 verbatim 기록 — verification-claim-integrity §3.2)
  - [ ] 4개 표면 편집 + `make build` 완료 (embed 재생성)
  - [ ] run-phase 커밋에 4개 표면 + 재생성 산출물만 specific-path로 포함 (git add -A 금지)
  - [ ] progress.md §E.2/§E.3 채움 (manager-develop 소관)
