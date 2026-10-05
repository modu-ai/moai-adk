---
id: SPEC-HEAVY-TEST-SLOT-001
title: "Heavy-test execution slot discipline: extend the existing lane slot paragraph with the enforcement code, named heavy packages, and the incident record"
version: "0.2.0"
status: completed
created: 2026-09-14
updated: 2026-09-14
author: lane-6
priority: P2
phase: "batch 9"
module: ".claude/rules/local (dev-only)"
lifecycle: spec-anchored
tags: "slot,heavy-test,serialization,factory,kanban,discipline,visibility"
tier: S
---

## §A — History

- **2026-09-14** — plan-phase v0.1.0 authored from card t774 ("무거운 테스트 패키지의 실행 순번을 레인이 볼 표면이 없다", lane-6 관측 2026-09-10). Plan-phase investigation (`.moai/reports/t774/investigation.md`) measured that the execution-order surface already exists — `moai slot`, first audit entry 2026-09-12T12:28:42Z (`heavy-test`), 19 events at plan time and growing — and is enforceable (second acquirer of a live holder refused, exit 3, demonstrated live on the card-scoped resource `t774-repro-demo`). The dispatch's shrink conditional fired: visibility + discipline, not surface construction.
- **2026-09-14** — plan-phase v0.2.0 after plan-audit iteration 1 returned **FAIL 0.70** (`.moai/reports/t774/plan-audit.md`; Tier S single-iteration ceiling). The audit's survey finding was decisive and is folded in whole: **the lane-facing procedure text already exists** — `.claude/rules/local/gitflow-lane-protocol.md` §8 has carried the slot duty (acquire → run → release, `status` lookup, the integration distinction, `--max-duration` expiry, the opt-in guard) since 2026-09-12 (e78fd0ee6, card t607 / SPEC-RESOURCE-SLOT-LEASE-001), and `.claude/rules/moai/workflow/resource-slot-lease.md` already documents the surface **and is template-mirrored to deployment users**. v0.1.0's problem statement ("binding procedure text is still missing") was therefore false on the base tree, and its proposed new standalone document would have created a third procedure copy — the divergence REQ-HTS-003 itself warned against. Consequences: **(1)** the standalone document is DROPPED; the deliverable becomes an extension of the existing §8 slot paragraph with exactly the three genuinely additive items (the refusal exit code 3, the named heavy-package WHEN list, the 2026-09-10 incident/control-group record); **(2)** the mirrored `resource-slot-lease.md` gains nothing — internal card/incident material must not enter a user-shipped document; **(3)** `status:` fields removed from plan.md/acceptance.md frontmatter (artifact statelessness, D2); **(4)** AC-HTS-005's product-untouched check re-based from a pinned SHA to a merge-base-relative diff so a develop absorb cannot false-FAIL it (D3); **(5)** §A's evidence citation now resolves (the investigation record exists, D5); **(6)** the lint-mandated `### Out of scope` heading added (D6). Requirements 4 → 3, criteria 5 → 4.
- **2026-09-14** — run-phase M1 landed (37b5ede0f): the §8 slot paragraph extended by three sentences (the exit-3 enforcement fact, the named heavy-package WHEN list, the 2026-09-10 incident/control-group record with the re-demonstration prohibition). All four acceptance criteria verified on the merge-base anchor `d416f8162` — AC-HTS-001 exit3=1·internal_cli=1; AC-HTS-002 incident=2·re-demo=2; AC-HTS-003 changed-file set carries none of the three forbidden prefixes; AC-HTS-004 slot.go/slot_test.go diff empty. Evidence: `.moai/reports/t774/verdict.md`. Sync-phase close: status → completed on the sync commit.

## §B — Problem

### B.1 — The incident, and what was actually missing

On the 2026-09-10 batch, the `internal/cli` full suite ran just under the 600s threshold, and concurrent runs flipped verdicts (three independent lane observations, load 8–21). The lead declared serialization; two intrusions followed, with **different causes** — (1) a missed notification, (2) a notified lane with no way to see whose turn it was. The card's diagnosis stands: declaration, notification, and an enforceable procedure are different things; a discipline that lives only in the lead's memory is not a procedure.

### B.2 — The gap that remains today (measured, post-t607)

The surface question is answered. `moai slot` exists, is enforceable, carries the card's full field list (holder session + pid, name, command, since, bound/ends — measured via `slot status`), and has been in active lane/lead use since 2026-09-12; the lane duty itself is already written (§8 of the lane protocol rule, e78fd0ee6). What the existing text does NOT give a lane is three concrete things, each of which the 09-10 episode shows is load-bearing:

1. **The enforcement fact as a number** — a second acquirer while a live holder is inside its bound is refused **with exit code 3**, silently, scriptably, without displacing anyone. A lane that knows the refusal is a number can probe-and-branch mechanically instead of asking the lead.
2. **A named WHEN list** — §8 says "무거운 실행이 겹칠 자리에서는", which on 09-10 did not map to "internal/cli 전체 스위트" for the lane that intruded. The known-heavy packages deserve to be named.
3. **The incident as control group** — the recorded observation that without the surface, intrusion happened (three lanes, load 8–21), and that the refusal demo is the enforcement half, so nobody re-runs concurrent 600s suites to "re-demonstrate" (the load discipline forbids exactly that).

### B.3 — Scope judgment: ships to deployment users?

`moai slot` ships as a product CLI verb and its user-facing document (`.claude/rules/moai/workflow/resource-slot-lease.md`) is template-mirrored. That document gains nothing from this SPEC: the deliverable is maintainer-side lane discipline in the dev-only lane rule, and internal card/incident material must not enter a mirrored file.

## §C — Requirements (EARS)

The system shall meet the following requirements. M1 is the only milestone.

- **REQ-HTS-001** — When the lane protocol's slot paragraph (`.claude/rules/local/gitflow-lane-protocol.md` §8) is next extended, it shall name the enforcement fact — a second acquirer while a live holder is inside its declared bound is refused with exit code 3, silently and without displacement — and shall name the known-heavy packages that make the WHEN rule concrete (at minimum `internal/cli`, the 09-10 incident subject; `internal/kanban` and `internal/hook` as further measured-minute-scale packages).
- **REQ-HTS-002** — The slot paragraph shall record the control group: the 2026-09-10 incident (three lanes, load 8–21, concurrent `internal/cli` runs flipping verdicts) as the no-surface evidence, and the 2026-09-14 exit-3 demonstration as the enforcement evidence, together with an explicit prohibition on re-running concurrent heavy suites to re-demonstrate either.
- **REQ-HTS-003** — The extension shall stay dev-only: no change to `.claude/rules/moai/workflow/resource-slot-lease.md` (template-mirrored), no new file under `internal/template/templates/`, no docs-site page, and no change to `moai slot` or any product code.

## §D — Non-Goals

### Out of scope

- Any change to `moai slot`, its tests, or its user-facing mirrored document — the surface exists and is pinned by tests.
- Enabling the opt-in slot PreToolUse guard (`workflow.slot_lease.enabled`) or building any new enforcement hook: if lanes later ignore the §8 duty, that decision has its own home and its own measurement first.
- Re-running concurrent heavy test suites to reproduce the 09-10 verdict flips — forbidden by the load discipline; the recorded incident plus the exit-3 demo are the control group.
- docs-site content, 4-locale obligation, template mirrors (§B.3).
