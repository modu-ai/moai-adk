# SPEC-UPDATE-HOOK-DELIVERY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-03
tier: M
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md
baseline_sha: d592b0551
open_decision: RESOLVED at run-phase M1 — Option B (detect + guide), operator decision 2026-09-03; verdict recorded in design.md §G
```

### M1 — Decision landing (2026-09-03, tree b77ae5d5e)

- **Decision: Option B (detect + guide only).** Operator decision fixed 2026-09-03 at Implementation Kickoff Approval. `moai update` merge behavior unchanged; `moai doctor` gains a read-only check that flags template hook entries missing from the project's `.claude/settings.json` within carried event keys, with per-entry guidance. Verdict + rejection rationale in design.md §G.
- **N/A markings (option-gated, per acceptance.md header):** REQ-UHD-007 (Option A) N/A; REQ-UHD-010 (Option C) N/A; AC-UHD-004, AC-UHD-005 (Option A gate) N/A; AC-UHD-007 (Option C gate) N/A. Binding option-gated set: REQ-UHD-008, REQ-UHD-009, AC-UHD-006, AC-UHD-013.
- **Detection identity rule:** entry identity = `.claude/hooks/moai/*.sh` handler path (canonical-JSON extraction); no-identity fallback = canonical JSON. Template side = embedded `settings.json.tmpl` rendered with the project's `hook.opt_in.enabled` (no false positives on opt-out projects). No template content change required.

### M1 pre-flight baseline (tree b77ae5d5e, branch WT-update-hook-delivery, worktree clean)

| Check | Command | Observed |
|---|---|---|
| Build | `go build ./...` | exit 0 |
| Cross-platform | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Affected packages | `go test ./internal/cli/update/... ./internal/merge/...` | all `ok`, exit 0 |
| Lint | `golangci-lint run --timeout=2m ./internal/cli/... ./internal/merge/...` | `0 issues.` |

### Defect-chain anchor re-verification (absorbed tree b77ae5d5e vs plan-phase d592b0551)

| Anchor | Plan-phase | This tree | Moved? |
|---|---|---|---|
| `pruneToShared` map-only recursion / wholesale copy / shared-key exclusion | base.go:112-131 (:124 recurse, :128 copy, :116-121 exclusion) | identical content at base.go:112-131 | no |
| "Only user changed" keeps user array | strategies.go:427-429 | identical content at strategies.go:427-429 | no |
| `checkHooksConfig` single os.Stat, settings.json never opened | doctor.go:768-783 | identical content at doctor.go:767-783 | 1 line |

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_
