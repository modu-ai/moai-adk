# SPEC-GLM-EFFORT-REBALANCE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M. Artifacts emitted: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.
- Scope was settled with the operator before authoring; no `[NEEDS CLARIFICATION]` markers remain.
- Coupled-file inventory verified against the tree during planning, not carried over from the delegation brief. Three corrections and four additions are recorded in `plan.md` §B.
- Requirements: REQ-GER-001 through REQ-GER-014 (GEARS). Criteria: AC-GER-001 through AC-GER-016, plus six items named in `acceptance.md` §F as not mechanically checkable.
- Amendment 1 (config shadow): surfaced by the coordinator from the primary checkout, which a worktree cannot see. `llm.profiles` is consulted before the Go default, so a Go-matrix-only edit would have been inert on every populated install. Added REQ-GER-012 / REQ-GER-013, milestone M0, and AC-GER-006 / AC-GER-013 / AC-GER-014. Corrected the `plan.md` §B-1 reason: the local `llm.yaml` is gitignored runtime state, not an absent file.
- Amendment 2 (plan-audit FAIL 0.64, Testability 0.50 — six criteria could not decide what they claimed). Resolved: two `internal/cli` test breakages invisible to the package set (D1); the false "sole consumer" claim, which hid a second main-session delivery path (D2); an unowned template `profiles:` block (D7); a false `llm.glm` premise (D8); an AC-GER-008 grep vacuous on the two `agent-authoring.md` surfaces it was meant to protect (D9); AC-GER-010 verifying nothing because `runModelProfile` never reads the embed (D4); AC-GER-013 written for the wrong YAML shape (D6); and the build step not refreshing the binary the criteria execute (D5). Added REQ-GER-014, AC-GER-015, AC-GER-016, and §B.0 Freshness gate.
- Status: `draft`. Awaiting re-audit and Implementation Kickoff Approval.

### Open gap for sync — MP-3 is PASS-by-inspection, not PASS-by-tool

`moai spec lint` exceeded the auditor's 120s window and never completed, so this SPEC's frontmatter validity (MP-3) was cleared by reading the 12 canonical fields, not by running the linter. Per `verification-claim-integrity.md` §1.1 surface 3, an inspection is a hypothesis where a dedicated tool exists. **The sync phase must re-run `moai spec lint` (with an extended timeout) against this SPEC and record the verbatim result** before treating frontmatter validity as verified.

### Audit-history note — D3 was real, then cleared by an unrelated session

The plan audit recorded `go vet ./...` and `GOOS=windows go vet ./...` both exiting 1 on `internal/cli/preference/home_isolation_test.go:75` (`undefined: userHomeDir`), which would have made AC-GER-009a/009b unsatisfiable. A re-run in the primary checkout after the audit exited 0 with zero output, and both `home_isolation_test.go` files no longer exist — a parallel session deleted them mid-audit. The auditor was not wrong; the tree moved underneath it.

The recommendation was kept anyway and AC-GER-009a/009b are now baseline-relative rather than absolute-exit-0. Three sessions share this checkout, so an absolute assertion is hostage to unrelated churn in both directions: it can fail correct work, and it can pass by luck.

## §E.2 Run-phase Evidence

_<pending run-phase>_

**Capture these BEFORE the first edit** — three criteria are undecidable without them:

| Artifact | Command | Consumed by |
|---|---|---|
| `baseline-vet-host.txt` | `go vet ./... > .moai/specs/SPEC-GLM-EFFORT-REBALANCE-001/baseline-vet-host.txt 2>&1` | AC-GER-009a |
| `baseline-vet-windows.txt` | `GOOS=windows go vet ./... > .moai/specs/SPEC-GLM-EFFORT-REBALANCE-001/baseline-vet-windows.txt 2>&1` | AC-GER-009b |
| `baseline-llm.yaml` | `cp .moai/config/sections/llm.yaml .moai/specs/SPEC-GLM-EFFORT-REBALANCE-001/baseline-llm.yaml` | AC-GER-014 |

Record alongside them: the pre-edit `moai model profile --json` output for the three agents (so the before/after delta is attributable), and `git rev-parse --short HEAD` at capture time.

These three baseline files are working artifacts, not deliverables — remove them at sync, or add them to the SPEC directory's ignore set if they are kept for audit.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
