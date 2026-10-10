# Decision Index — SPEC-CAPACITY-RECLAIM-001

Authored because `.moai/config/sections/interview.yaml` sets `decision_gate: on` (line 6). Stateless on the status axis (no status field). Operator verdicts are left empty at authoring. No row is implementation-level, so no `DEFAULT-APPLIED` line is written at plan close. All four rows remain OPEN in revision v0.1.2.

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
Why unresolved: disposal is irreversible. `moai worktree sweep --yes` disposes every candidate whose predicates pass; its evaluation has no owner predicate and no unpushed-commit predicate today (`internal/cli/worktree/sweep.go:409–525`), and its `--yes` loop has no owner-confirmation step. An unpushed branch holds the only copy of its work until a remote has it. Under this SPEC's REQ-CAP-011, the owner predicate preserves every tree that no factory card binds. A bulk run therefore reaches only card-bound trees whose card is `done`; manually created trees keep the `moai worktree done` path (`internal/cli/worktree/done.go`). Whether that scope is acceptable for a bulk run is part of this decision. The approval gates only the bulk run this SPEC executes (M4); it does not remove the `--yes` verb, which REQ-CAP-011 narrows.
Default: reject (no bulk run executed under this SPEC) (rule: the option that preserves current behavior: no disposal happened)
Alternate: approve a bulk run against the report, after M3 and M4 are in place
Operator verdict:

### Q3: Which clones hold local data in `usage-log.jsonl` and `lessons-inbox.jsonl` that the untracking commit would affect on pull?

Label: EVIDENCE-NEEDED
Authority anchor: none
Why unresolved: the number of clones and each clone's local state are not measured. A commit that removes a tracked file changes the working copy of each clone that pulls it. How git treats a locally modified copy on that pull (refusal, or removal of a copy identical to HEAD) is git's behavior and is not measured in this SPEC. Measurement recipe, to be run by the operator in each clone: `git status --porcelain -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl` (a listed path differs from HEAD) and `wc -c .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl`. Record the count of clones with a listed path, and the byte counts, in the row before the verdict.
Default: hold the M2 push until the evidence is recorded (rule: the option that preserves current behavior: the files stay tracked)
Alternate: push the M2 commit after the evidence is recorded
Operator verdict:

### Q4: Should `.moai/harness/usage-log.jsonl` be rotated by the hygiene rotator instead of its existing owner?

Label: FOUNDER
Class: product-level (fail-closed: it reverses a completed SPEC's ownership decision and changes which mechanism bounds a user-facing file)
Authority anchor: none in the authority register. SPEC-MOAI-HYGIENE-001 §D (`.moai/specs/SPEC-MOAI-HYGIENE-001/spec.md`, lines 107–110) assigns usage-log retention to `internal/harness/retention.go`. That is a body entry, not a HISTORY or `## Amendments` row, so it cannot anchor DECIDED or POLICY-COVERED.
Why unresolved: REQ-CAP-004 requires rotation by the hygiene rotator, while the completed SPEC records the opposite owner. The owner note in `internal/hygiene/sinks.go:91` reads "30-day prune to monthly gz", and `internal/harness/retention.go` implements that prune into monthly archives. The tracked copy measured 88,629,933 bytes on 10-10, so the effect of the prune on the active file is not measured here. The approval has a consequence the operator must accept: a rotation displaces the file into its one kept `.1` rotation, and the next rotation removes that `.1` chunk without archiving it (`internal/hygiene/rotate.go:241–245`). Events in that chunk that the 30-day prune has not yet archived are then lost. The alternate keeps the prune and serializes it with the rotation; it does not retire the prune, so events still in the primary file keep reaching the monthly archive.
Default: keep `internal/harness/retention.go` as the owner (rule: the option that preserves current behavior)
Alternate: register usage-log in the hygiene rotator, remove its owner exception, and serialize the rotation with the 30-day prune (the consequence above applies)
Operator verdict:
