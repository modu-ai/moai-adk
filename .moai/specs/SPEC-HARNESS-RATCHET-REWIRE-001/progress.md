# SPEC-HARNESS-RATCHET-REWIRE-001 — Progress

> Plan-phase artifact. Run-phase evidence (§E.2, §E.3) and sync-phase evidence (§E.4) are populated by manager-develop and manager-docs respectively. This file is parser-load-bearing: `internal/spec/era.go` `hasAnyProgressMarker` greps for the literal `§E.2` / `§E.3` / `§E.4` substrings — the headings below MUST be preserved verbatim.

---

## §E.1 Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: M
epic: Workflow-Reflex (1 of 6)
artifacts: 4 (spec.md + plan.md + acceptance.md + progress.md)
gears_requirements: 9
acceptance_criteria: 13
out_of_scope_topics: 6
audit_findings_traced: H1, H2, L2
spec_id_self_check: PASS (SPEC-HARNESS-RATCHET-REWIRE-001 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$)
```

**Plan-phase self-verification**:
- [x] All GEARS requirements use valid patterns (Ubiquitous / When / Where / shall-not; no IF/THEN).
- [x] All gap claims cite measured source + observed pattern (vci §2 attribution; measured 2026-07-09).
- [x] §Out of Scope has ≥1 `### Out of Scope — <topic>` H3 sub-heading with `-` bullets (6 topics).
- [x] 12 canonical frontmatter fields + era + tier; no snake_case aliases.
- [x] Human apply gate (`auto_apply: false`) preservation stated as HARD constraint.
- [x] Fail-open + 5s hook-budget constraint stated for the Stop-path propose chain.

---

## §E.2 Run-phase Evidence

### Pre-flight baseline (measured 2026-07-09 by manager-develop)

```
git branch --show-current → worktree-agent-a949030202916a3bf (fast-forwarded from main)
git rev-parse HEAD        → ab4f38def (== origin/main, divergence 0 0 after ff)
go build ./...            → exit 0
GOOS=windows GOARCH=amd64 go build ./... → exit 0
golangci-lint run         → 0 issues (clean baseline)
go test ./internal/harness/... ./internal/hook/... ./internal/cli/harness/... → all ok (exit 0)
grep "classifyHarnessPatterns" internal/cli/hook.go → line 772 (Stop chain anchor)
grep "Retired\|superseded" internal/harness internal/cli/harness → no output (B2 clean)
```

### Open decision resolutions (plan.md §D)

- **D1 (failure signature)**: RESOLVED → low-cardinality error-class token. Reuse the existing `ErrorCategory` (TimeoutError/ExitError/etc. ~7 tokens) as the `<signature>` ContextHash. Avoids cardinality explosion.
- **D2 (propose invocation seam)**: RESOLVED → extract a callable `generateProposals(root)` in cli/hook.go reusing `proposalgen.ReadPromotions`/`MapPromotions`/`WriteProposals`. No subprocess (within 5s budget).
- **D3 (inbox stub schema)**: RESOLVED → append-only JSONL (ts / event_key / summary / source). Drain marking deferred to the orchestrator's Lessons Protocol (named by M3 doctrine cross-ref).
- **D4 (eligibility predicate placement)**: RESOLVED → filter at classification time (in `classifyHarnessPatterns` loop, before `WritePromotion`).

### D2-audit note (AC-HRR-012 grep refinement)

acceptance.md §D.4 AC-HRR-012 grep matches 6 string-literal lines in the clean baseline (propose.go:48/50/59, execute.go:320, install.go:117, scaffolder.go:111). The E4 self-verification anchors on the call pattern `AskUserQuestion(` to avoid the false-positive FAIL.

_<M1/M2/M3 evidence appended below as milestones complete>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_
