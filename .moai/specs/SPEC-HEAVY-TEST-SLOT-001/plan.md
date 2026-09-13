---
id: SPEC-HEAVY-TEST-SLOT-001
title: "Plan — heavy-test slot discipline doc + lane-protocol pointer"
version: "0.1.0"
status: draft
created: 2026-09-14
updated: 2026-09-14
author: lane-6
tier: S
---

## §A — Approach

The card's design question was answered by measurement, not by design: the execution-order surface already exists (`moai slot`), is enforceable (live-holder acquire refused, exit 3 — demonstrated this session on the card-scoped resource `t774-repro-demo`), carries the card's full field list, and is in active lane/lead use since 2026-09-12 (19 audit events). The remaining defect is that lanes have no binding text telling them WHEN the slot is required. The deliverable is therefore a single dev-only protocol document plus one pointer — no product change, no surface construction.

## §B — Milestones

### M1 — protocol document + pointer (one run commit + one sync close)

1. Author `.moai/docs/heavy-test-slot-protocol.md` (dev-only, Korean, clean written register):
   - WHEN: package-level suites that are minutes-scale or known-heavy (`internal/cli`, `internal/kanban`, `internal/hook`, or any suite a lane expects past the ~5-minute fence).
   - HOW: `moai slot status --resource <pkg>-tests` (see the current holder) → `moai slot acquire --resource <pkg>-tests --name <lane/card> --command "<suite>"` → run → `moai slot release --resource <pkg>-tests`.
   - Enforcement fact: a second acquirer while a live holder is inside its bound is refused, exit 3, silent — scriptable; measured 2026-09-14 on `t774-repro-demo`.
   - Distinction: this is execution order; `moai integration` is the merge window. Neither substitutes for the other (the 09-10 batch needed both).
   - Control-group record (REQ-HTS-002): the 09-10 incident observations are the no-surface evidence; the exit-3 demo is the enforcement evidence; concurrent heavy-suite re-runs to "re-demonstrate" are forbidden by the load discipline.
   - Dev-only header (no template mirror, no docs-site — REQ-HTS-004).
2. Add a pointer paragraph to `.claude/rules/local/gitflow-lane-protocol.md` lane-duties area: the duty (slot-acquire before minutes-scale suites) + forward reference, zero procedure duplication (REQ-HTS-003).
3. Verify per acceptance.md; commit (run phase) and close the SPEC on the sync commit.

## §C — Risks / Notes

- `.claude/rules/local/*` and `.moai/docs/heavy-test-slot-protocol.md` are dev-only files — template-neutrality classes (§25) do not apply because nothing mirrors, but the doc must still avoid-for-shipping markers anyway (no template authoring occurs here).
- The protocol names `internal/cli` etc. as known-heavy from measured history (1582s package-wide observation, t586; ~600s threshold incident, t774 card). The WHEN rule is a judgement threshold, deliberately soft — hard-coding a duration gate is a product change this SPEC excludes.
- If lanes later ignore the doc, the next escalation is enabling the existing opt-in slot PreToolUse guard — recorded as a non-goal for this SPEC so the future decision has a named home.
