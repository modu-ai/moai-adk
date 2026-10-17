# Decision Index — SPEC-CARD-METADATA-VERBS-001

Authored under `interview.decision_gate: on` (`.moai/config/sections/interview.yaml`). One row per decision surfaced during assembly that the operator delegation did not settle. Stateless on the status axis (no `status:` field); the SPEC's lifecycle lives in `spec.md` alone.

---

### Q1: Verb family — `moai factory set-spec` / `moai factory reclassify`, or `moai todo set-spec` / `moai todo reclassify`?

Label: verb-naming
Authority anchor: none (no prior SPEC or operator setting names a metadata-verb family)
Why unresolved: The delegation says "card-metadata editing verbs for the factory/todo card record" — both families are named, neither is chosen. `set-spec` writes the factory card record beside `factory assign`; `reclassify`'s primary audience (queued cards) has no factory record and lives where the `todo` verb family operates. `todo classify` already exists as an add-time judgment verb, so `todo reclassify` would read as its post-issuance sibling, while `factory set-spec` would read as `assign --spec`'s post-pick sibling — each option reads naturally on one verb and forcedly on the other.
Operator verdict:

Class: product-level (names a new user-facing command surface).
Default: `moai factory` family only, no todo alias (rule: smaller user-visible surface — one family, two verbs, no alias surface).
Alternate: `moai todo` family, or both families with one alias.

---

### Q2: Lane carve-out — may a lane session run `set-spec` on its OWN card to fill a BLANK spec_id (t1715's lane-recovery affordance), or are the verbs leader/operator-only entirely?

Label: lane-boundary-carveout
Authority anchor: SPEC-FACTORY-SELF-DISPATCH-001 (completed) — the lane-mutation boundary base rule (REQ-SD-015/-016; sentinel and refusal helper in `internal/cli/factory_card.go:120-131`). The base rule is settled; the carve-out is not.
Why unresolved: The base doctrine refuses lane sessions every queue mutation, and REQ-CMV-019 encodes that default. But the delegation's coordination section says t1715's item (4) — "link a SPEC id when the record has none" — is a LANE recovery affordance that REUSES this verb; a reuse the base doctrine refuses is a contradiction only the operator can resolve: either the carve-out is granted (lane may fill a blank spec_id on its own card, never overwrite a non-blank one), or t1715's item (4) is re-scoped to ask the leader.
Operator verdict:

Class: implementation-level (permission detail of a new verb; does not change an existing shipped command's behavior).
Default: no carve-out — leader/operator-only; t1715's item (4) routes through the leader (rule: preserves current behavior — today no lane queue mutation is admitted anywhere).
Alternate: own-card blank-fill carve-out for `set-spec` only (never `reclassify`, never a non-blank overwrite).

---

### Q3: Audit-trail landing — a new card-keyed `card_metadata_events` table, or the existing run-keyed `events` table with an empty run for queue-only cards?

Label: audit-surface-landing
Authority anchor: none eligible — the design doc (`.moai/reports/card-policy-redesign-20261017/design.md`) is machine-local and untracked, and untracked material is never cited as authority. Its principle 5 ("durable card identity; runs are membership") motivates the card-keyed option but cannot anchor it.
Why unresolved: The delegation says "study the existing events mechanism in internal/homestate and reuse its shape" and names both landings as open ("cards table events vs a new metadata_events table"). The existing `events` table is run-keyed (`run_id`); a queued card has no run, so reclassify of a queued card would need either a synthetic run id (invents identity) or an empty run_id (rows outside the table's key discipline). A new card-keyed table reuses the shape (seq, kind, JSON payload, created-at) with the key the data actually has — but adds a table stage B's identity migration must carry.
Operator verdict:

Class: implementation-level.
Default: none — the published Default rule ties at every step (no current behavior to preserve for a surface that does not exist; both options undo by a single revert of this SPEC's commits; neither is user-visible). This row blocks autonomous Kickoff until the operator rules.

---
