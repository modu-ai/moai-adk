# Decision Index — SPEC-CARD-METADATA-VERBS-001

Authored under `interview.decision_gate: on` (`.moai/config/sections/interview.yaml`). One row per decision surfaced during assembly that the operator delegation did not settle. Stateless on the status axis (no `status:` field); the SPEC's lifecycle lives in `spec.md` alone.

Iteration-2 note (2026-10-17): the plan-audit iteration-1 verdict (FAIL 0.81, `.moai/reports/t1840/plan-audit.md`) assessed Q1 and Q2 defaults as safe and recommended option (a) for Q3. The lane ruled on all three per the decision ladder (`auto-semantics.md` §6 — the lane is the decision authority for in-card judgments); the rulings are recorded below and are not reopened.

---

### Q1: Verb family — `moai factory set-spec` / `moai factory reclassify`, or `moai todo set-spec` / `moai todo reclassify`?

Label: verb-naming
Authority anchor: none (no prior SPEC or operator setting names a metadata-verb family)
Why unresolved: The delegation says "card-metadata editing verbs for the factory/todo card record" — both families are named, neither is chosen. `set-spec` writes the factory card record beside `factory assign`; `reclassify`'s primary audience (queued cards) has no factory record and lives where the `todo` verb family operates. `todo classify` already exists as an add-time judgment verb, so `todo reclassify` would read as its post-issuance sibling, while `factory set-spec` would read as `assign --spec`'s post-pick sibling — each option reads naturally on one verb and forcedly on the other.
Ruling provenance (2026-10-17): lane ruling per the decision ladder — the Default stands (`moai factory` family only); plan-audit iteration 1 assessed the default as safe ("smallest user-visible surface, additive to reverse"). Product-level row: the operator confirms at Kickoff; the lane ruling does not substitute for that confirmation.
Operator verdict:

Class: product-level (names a new user-facing command surface).
Default: `moai factory` family only, no todo alias (rule: smaller user-visible surface — one family, two verbs, no alias surface).
Alternate: `moai todo` family, or both families with one alias.

---

### Q2: Lane carve-out — may a lane session run `set-spec` on its OWN card to fill a BLANK spec_id (t1715's lane-recovery affordance), or are the verbs leader/operator-only entirely?

Label: lane-boundary-carveout
Authority anchor: SPEC-FACTORY-SELF-DISPATCH-001 (completed) — the lane-mutation boundary base rule (REQ-SD-015/-016; sentinel and refusal helper in `internal/cli/factory_card.go:120-131`). The base rule is settled; the carve-out is not.
Why unresolved: The base doctrine refuses lane sessions every queue mutation, and REQ-CMV-015 encodes that default. But the delegation's coordination section says t1715's item (4) — "link a SPEC id when the record has none" — is a LANE recovery affordance that REUSES this verb; a reuse the base doctrine refuses is a contradiction only the operator can resolve: either the carve-out is granted (lane may fill a blank spec_id on its own card, never overwrite a non-blank one), or t1715's item (4) is re-scoped to ask the leader.
Ruling provenance (2026-10-17): lane ruling per the decision ladder — no carve-out; plan-audit iteration 1 assessed the default as safe ("preserves today's behavior exactly; t1715's item (4) routes through the leader at the cost of one message"). Recorded as DEFAULT-APPLIED below.
Operator verdict: DEFAULT-APPLIED 2026-10-17T13:02:13Z lane-15 (factory lane, decision-ladder ruling, run tn1p4e)

Class: implementation-level
Class detail: permission detail of a new verb; does not change an existing shipped command's behavior.
Default: no carve-out — leader/operator-only; t1715's item (4) routes through the leader (rule: preserves current behavior — today no lane queue mutation is admitted anywhere).
Alternate: own-card blank-fill carve-out for `set-spec` only (never `reclassify`, never a non-blank overwrite).

---

### Q3: Audit-trail landing — a new card-keyed `card_metadata_events` table, or the existing run-keyed `events` table with an empty run for queue-only cards?

Label: audit-surface-landing
Authority anchor: none eligible — the design doc (`.moai/reports/card-policy-redesign-20261017/design.md`) is machine-local and untracked, and untracked material is never cited as authority. Its principle 5 ("durable card identity; runs are membership") motivates the card-keyed option but cannot anchor it. The disposition below is a lane ruling per the decision ladder (`auto-semantics.md` §6), recorded as the Default so downstream artifacts cite a real decision.
Why unresolved (iteration-2 corrected, plan-audit D8): The delegation says "study the existing events mechanism in internal/homestate and reuse its shape" and names both landings as open. The existing `events` table is run-keyed (`run_id`), and a queued card has no run. Schema legality is NOT the barrier: `events.run_id` is `TEXT NOT NULL DEFAULT ''` (`internal/homestate/factory.go:81`), so an empty `run_id` is schema-legal. The barrier is consumer/keying discipline: run-keyed rows for card mutations have no indexed card lookup, mix two key disciplines in one table, and otherwise force an invented run id for cards that have none. A third option — adding a nullable card column to the existing `events` table — is real but FORECLOSED by this SPEC's own row-preserving migration discipline (REQ-CMV-014: additive `CREATE TABLE IF NOT EXISTS` only); it is a foreclosed option, not a missed one.
Ruling provenance (2026-10-17): lane ruling per the decision ladder — option (a); plan-audit iteration 1 recommended the same (D1 fix route; the D8 wording correction above is applied). Rationale: design principle 5 (durable card identity; runs are membership) — the new table is keyed by the identity the data actually has; the run-keyed option forces invented runs or unindexed lookups; the nullable-column option is foreclosed by REQ-CMV-014.
Operator verdict: DEFAULT-APPLIED 2026-10-17T13:02:13Z lane-15 (factory lane, decision-ladder ruling, run tn1p4e)

Class: implementation-level
Default: option (a) — a NEW card-keyed `card_metadata_events` table in the homestate database, keyed by card id, reusing the `events` shape (autoincrement seq, kind, JSON payload, created-at).
Alternate: the existing run-keyed `events` table with `run_id = ''` for queue-only cards (schema-legal; rejected on consumer/keying discipline grounds, not schema legality).

---
