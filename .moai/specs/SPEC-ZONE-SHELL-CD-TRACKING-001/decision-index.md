# SPEC-ZONE-SHELL-CD-TRACKING-001 — decision index

Authored at plan phase (interview.decision_gate: on, `.moai/config/sections/interview.yaml:6`). Stateless on the status axis; the SPEC's lifecycle lives in spec.md alone. Rows route by the four labels DECIDED / POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER; an unverifiable citation never relabels a row.

### Q1: Should a post-`--` hyphen-leading cd directory operand (`cd -- -zone`) be tracked like any other operand — denying a protected deletion — instead of keeping the documented untracked reset (allow)?

Label: FOUNDER
Class: product-level
Authority anchor: none — the settlement exists only in untracked card evidence
Why unresolved: Both the plan-audit D7 required fix (deny — the shape deletes `-zone/a.log`, a protected path; the drafted allow control was ruled a self-contradiction with SPEC-ZONE-SHELL-PARSING-001 REQ-ZSP-001) and the card t1584 dispatch (operator-approved expansion axis) direct deny, but both records live under `.moai/reports/` (untracked); no completed SPEC's HISTORY/Amendments row or `.moai/config/sections/*.yaml` setting decides the identical question, and the authority register cannot cite untracked material. This SPEC's REQ-ZSCD-001/REQ-ZSCD-005 freeze the deny direction pending the operator pin.
Operator verdict:

### Q2: Does an outside-root cd destination keep the documented untracked reset (allow), recorded outside the regression matrix?

Label: FOUNDER
Class: product-level
Authority anchor: none — the settlement exists only in untracked card evidence
Why unresolved: The card directs the retention of a landed allow behavior that is an HONEST accepted under-match (v0.2.0 rationale, per audit round 1 D2): a `..`-reaching relative spelling from an outside-root cwd reaches back into the zone while the guard answers allow — the tracked set was reset to the root, so the `..`-leading candidate is classified root-escaping while real bash resolves it inside the project (codex isolated repro at the audited tree: `cd <parent>/002/deep && rm ../../001/zone_dir/a.log` → allow, vs direct `rm zone_dir/a.log` → deny; code path `zoneNextCwd` :451 reset + `zoneRelativeToSet` :366 tracked-cwd-only join). The RETENTION itself stands — a card may keep a landed under-match, it may not mislabel it as sound; no committed register anchor decides the retention, so the operator pin formalizes keeping a known-reaching under-match outside the matrix (the alternative — hardening the class — grows scope beyond this card's two files). REQ-ZSCD-003 and its Out of Scope record now carry this true rationale.
Operator verdict:

### Q3: Where do the cd-class regression cells live?

Label: FOUNDER
Class: implementation-level
Default: extend the landed `internal/hook/protected_zone_shell_matrix_test.go` cell catalogue with a cd group (rule: the option that preserves current behavior — no new test file)
Alternate: a dedicated cd matrix test file
Why unresolved: A file-placement judgment surfaced during assembly; no policy or prior SPEC row decides it, and it needs no operator data — the published Default rule ranks it.
Operator verdict: DEFAULT-APPLIED 2026-10-09T03:15:40Z manager-spec (t1584-spec, plan-phase)
