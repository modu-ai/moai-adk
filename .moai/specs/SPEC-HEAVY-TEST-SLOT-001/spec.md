---
id: SPEC-HEAVY-TEST-SLOT-001
title: "Heavy-test execution slot discipline: make the existing moai slot surface binding for lanes (visibility doc + lane-protocol pointer)"
version: "0.1.0"
status: draft
created: 2026-09-14
updated: 2026-09-14
author: lane-6
priority: P2
phase: "batch 9"
module: ".moai/docs (dev-only)"
lifecycle: spec-anchored
tags: "slot,heavy-test,serialization,factory,kanban,discipline,visibility"
tier: S
---

## §A — History

- **2026-09-14** — plan-phase v0.1.0 authored from card t774 ("무거운 테스트 패키지의 실행 순번을 레인이 볼 표면이 없다", lane-6 관측 2026-09-10). The card asked whether a separate execution-order surface was needed and, if so, its design. The plan-phase investigation `.moai/reports/t774/verdict.md` §2 measured that the surface **already exists and is already in use**: `moai slot` (internal/cli/slot.go) — acquire/status/release over cross-session leases rooted at the primary checkout's `.moai/state/slot-leases`, first audit entry 2026-09-12T12:28:42Z (`heavy-test`), 19 events to date (`heavy-test`×13, `t675-cli-tests`×4, `gotest-gateway`×2), refusal of a live holder's second acquirer measured live this session (exit 3, no displacement). The dispatch's own conditional therefore fires and the SPEC **shrinks to visibility + discipline**: one dev-only protocol document and one pointer from the lane protocol rule. Requirements 4, criteria 5.

## §B — Problem

### B.1 — The incident, and what was actually missing

On the 2026-09-10 batch, the `internal/cli` full suite ran just under the 600s threshold, and concurrent runs flipped verdicts (three independent lane observations, load 8–21). The lead declared serialization; two intrusions followed, with **different causes** — (1) a missed notification, (2) a notified lane with no way to see whose turn it was. The card's diagnosis stands as written: declaration, notification, and an enforceable procedure are different things; a discipline that lives only in the lead's memory is not a procedure.

### B.2 — The reframing: the surface question is already answered

The plan-phase investigation (evidence §2) measured all three of the card's judgment items:

- **(a) separate surface?** One exists — `moai slot`. It was built after the card was issued (first audit 09-12; card 09-10) and is in active use by lanes and the lead for exactly this purpose. What was missing on 09-10 and is still missing is the **binding procedure text** that tells a lane WHEN it must slot-acquire and that a holder exists to be checked.
- **(b) reuse `moai integration`?** Not applicable and correctly so — integration is the merge window (one holder per integration tree, released on merge completion); slot is execution order over a named resource. Distinct records, distinct lifetimes; the batch that produced this card needed both simultaneously.
- **(c) recorded fields?** `moai slot status` carries holder session + pid, name, command, since, and bound/ends — the card's full field list, measured.

### B.3 — Scope judgment: ships to deployment users?

`moai slot` itself ships as a product CLI verb. The deliverable of this SPEC is not the verb — it is the **factory/kanban lane discipline** that references it. That discipline is maintainer-side operational text: dev-only, no template mirror, no docs-site page.

## §C — Requirements (EARS)

The system shall meet the following requirements. M1 is the only milestone.

- **REQ-HTS-001** — When a lane or the lead is about to run a package-level test suite whose expected duration is minutes-scale or which is a known heavy package (e.g. `internal/cli`, `internal/kanban`, `internal/hook`), the system shall provide a dev-only protocol document `.moai/docs/heavy-test-slot-protocol.md` that states, in order: the WHEN rule above; the procedure (`moai slot status --resource …` to see the current holder, `moai slot acquire --resource <pkg>-tests --name <lane/card> --command "<suite>"` before running, `moai slot release --resource …` after); the enforcement fact that a second acquirer while a live holder is inside its bound is refused with exit 3 (measured); and the distinction between this lease and the `moai integration` merge window.
- **REQ-HTS-002** — The protocol document shall record the control group obligation: the 2026-09-10 incident observations (three lanes, load 8–21, concurrent `internal/cli` runs flipping verdicts) stand as the recorded "no surface → intrusion" evidence, and the exit-3 refusal demonstration stands as the enforcement evidence; the document shall not re-run concurrent heavy suites to re-demonstrate either (the repository's load discipline forbids exactly that class of background load).
- **REQ-HTS-003** — The lane protocol rule (`.claude/rules/local/gitflow-lane-protocol.md`) shall carry a pointer paragraph into the protocol document in its lane-duties area — a duty statement plus a forward reference, with no duplicated procedure body (two copies diverge).
- **REQ-HTS-004** — The document and pointer shall be dev-only: no file under `internal/template/templates/`, no docs-site page. `moai slot` itself is a product verb and is untouched by this SPEC.

## §D — Non-Goals

- No change to `moai slot` or any product code — the surface exists and is tested (internal/cli slot tests pin the refusal).
- No new enforcement hook: the opt-in slot PreToolUse guard already exists; making it mandatory is out of scope until the discipline is observed and measured insufficient.
- No docs-site content, no 4-locale obligation (§B.3).

## §E — Delivery

M1: author `.moai/docs/heavy-test-slot-protocol.md` (REQ-HTS-001/002/004), add the pointer paragraph (REQ-HTS-003), verify per acceptance.md, commit both plus the SPEC close on one sync commit.
