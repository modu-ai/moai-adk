# Sync Audit — SPEC-TODO-LANDING-EVIDENCE-001 (card t359)

> **PROVISIONAL — audit in progress.** Written early by deliberate budget policy (a prior
> audit attempt died mid-run and returned nothing). Gate results and final dimension scores
> are pending; this file is revised in place before the audit closes.

- Auditor: sync-auditor (independent, single writer)
- Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`
- Branch: `WT-landing-evidence`
- HEAD measured by this auditor: `f5d57fca4` (tree clean at audit start)
- Tier L, PASS threshold 0.85, profile `default` (harness.yaml `default_profile: "default"`)

## Provisional verdict

PENDING — one confirmed Medium security finding (F1, path traversal, mechanically
demonstrated); gates still running.

## Findings so far

### F1 — `ReadPrimarySpecStatus` joins an unvalidated `spec_id`, reading a file outside the project root
- Severity: Medium · Confidence: HIGH (mechanically demonstrated, not inferred) · Blocking: TBD
- Location: `internal/kanban/status_read.go:281-289` (new in this card)
- Reached from: `internal/cli/todo_landed.go:300` → `internal/cli/todo.go:584` (`--spec` recorded verbatim)

Demonstrated end-to-end with a binary built from this tree; stored record read back from
the queue database contained `"spec_status":"PWNED-TRAVERSAL"`, sourced from a `spec.md`
outside the project root.

