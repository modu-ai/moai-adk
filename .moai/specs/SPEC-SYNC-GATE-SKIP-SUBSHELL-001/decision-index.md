# decision-index.md — SPEC-SYNC-GATE-SKIP-SUBSHELL-001

Authored because `interview.decision_gate: on` (`.moai/config/sections/interview.yaml:6`). Stateless on the status axis; the SPEC's lifecycle lives in `spec.md` alone.

### Q1: Is "delete the external pipeline on the four check lines" the settled repair shape, or should the skip record be reconstructed from a subshell-immune surface (e.g. derived from `steps.journal` at the audit line) instead of carried in the `SKIPPED_TOOLS` variable?

Label: FOUNDER

Authority anchor: none found. Checked: no prior SPEC's HISTORY or `## Amendments` row decides the pipeline-wrapping-`run_step` question (catalog grep over `.moai/specs/` for `run_step`/gate-script owners — the owning SPECs cover the t602 exit-status shape, the cpp zero-target case, and the cache contract, not the subshell-variable shape); no `.moai/config/sections/*.yaml` key or constitution clause governs gate-script repair shape.

Why unresolved: the card dispatch directs "remove the external pipeline — `run_step` already owns its own log redirection", and the plan-phase verification confirms that direction loses nothing observable (spec.md §B P2). But card text is not in the authority register (committed artifacts only), and one alternative shape — journal-derived skip reconstruction — is strictly more robust against ANY future subshell regression of this family (the card notes this is the 16th reproduction), at the cost of more moving parts than a 4-line deletion. The trade between minimal-diff (default, plan.md IN-2) and regression-immune-reconstruction is a judgment call the operator owns.

Operator verdict:
