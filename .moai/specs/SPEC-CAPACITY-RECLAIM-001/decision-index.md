# Decision Index — SPEC-CAPACITY-RECLAIM-001

Authored because `.moai/config/sections/interview.yaml` sets `decision_gate: on` (line 6). Stateless on the status axis (no status field). Operator verdicts are left empty at authoring. No row is implementation-level, so no `DEFAULT-APPLIED` line is written at plan close.

### Q1: Should the shipped default of `workflow.hygiene.mode` change from `report` to `apply`?

Label: FOUNDER
Class: product-level
Authority anchor: none. The card's operator approval is recorded in the untracked queue and is not a committed authority.
Why unresolved: the flip changes a template default. The SessionStart auto path would mutate by default instead of reporting (`internal/template/templates/.moai/config/sections/workflow.yaml:318` reads `report`). The CLI form, `moai clean --audit-logs`, is unaffected: it mutates only with `--apply` on each invocation (`internal/cli/clean.go:113`). The card scope lists the flip, but that listing cannot be cited as authority.
Default: report (rule: the option that preserves current behavior)
Alternate: apply
Operator verdict:

### Q2: May the operator approve a bulk worktree-sweep disposal run against the candidates in its report?

Label: FOUNDER
Class: product-level (fail-closed: irreversible disposal of user work)
Authority anchor: none
Why unresolved: disposal is irreversible. `moai worktree sweep --yes` already disposes candidates whose predicates pass, and its path has no owner-confirmation check (`internal/cli/worktree/sweep.go:230`). An unpushed branch holds the only copy of its work. REQ-CAP-012 requires owner confirmation before each disposal; the operator decides whether a bulk run is approved once that gate exists.
Operator verdict:

### Q3: Which clones hold local data in `usage-log.jsonl` and `lessons-inbox.jsonl` that the untracking commit would delete on pull?

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: the number of clones and each clone's local data are not measured. Untracking removes the working copy on pull in every clone.
Operator verdict:

### Q4: Should `.moai/harness/usage-log.jsonl` be rotated by the hygiene rotator instead of its existing owner?

Label: FOUNDER
Class: product-level (fail-closed: it reverses a completed SPEC's ownership decision and changes which mechanism bounds a user-facing file)
Authority anchor: none in the authority register. SPEC-MOAI-HYGIENE-001 §D (`.moai/specs/SPEC-MOAI-HYGIENE-001/spec.md`, lines 107–110) assigns usage-log retention to `internal/harness/retention.go`. That is a body entry, not a HISTORY or `## Amendments` row, so it cannot anchor DECIDED or POLICY-COVERED.
Why unresolved: REQ-CAP-004 requires rotation by the hygiene rotator, while the completed SPEC records the opposite owner. The owner's note in `internal/hygiene/sinks.go:91` reads "30-day prune to monthly gz", and `internal/harness/retention.go` implements the prune into monthly archives. The tracked copy measured 88,629,933 bytes on 10-10, so the effect of that prune on the active file is not measured here. Registering the file also requires removing its owner exception.
Default: keep `internal/harness/retention.go` as the owner (rule: the option that preserves current behavior)
Alternate: register usage-log in the hygiene rotator and remove the owner exception
Operator verdict:
