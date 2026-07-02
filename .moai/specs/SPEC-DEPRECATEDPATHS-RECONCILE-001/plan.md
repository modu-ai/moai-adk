---
id: SPEC-DEPRECATEDPATHS-RECONCILE-001
title: "Reconcile DeprecatedPaths — implementation plan (Tier S)"
version: "0.1.0"
status: draft
tier: S
era: V3R6
---

# Implementation Plan — SPEC-DEPRECATEDPATHS-RECONCILE-001 (Tier S)

## §A. Context

- Work location: `/Users/goos/MoAI/moai-adk-go` (project root, absolute).
- Branch: created path-limited by the orchestrator (this agent does NOT commit).
- SPEC artifacts: `.moai/specs/SPEC-DEPRECATEDPATHS-RECONCILE-001/{spec.md, plan.md, progress.md}` (Tier S — AC inline in spec.md §3; no acceptance.md).
- Change surface: pure Go manifest + pinned tests + doc comments. **No template file touched → no `make build`.**
- Recommended direction: **Direction A (un-deprecate)** — remove `design.yaml` + `db.yaml` from `DeprecatedPaths` (see spec.md §1.3).

## §B. Baseline (green, this tree — plan-phase observation)

- `go test ./internal/defs/ -run TestDeprecatedPaths` → expected `ok` at 43 entries (current pinned state).
- `grep -c '_TBD_' internal/template/templates/.moai/config/sections/db.yaml` → `0` (REQ-VVCR-012 premise stale).
- `internal/cli/hook.go:448 loadMigrationPatterns` reads `db.yaml`; `internal/config/loader.go:92` reads `design.yaml` — both live.

## §C. Approach

Pure lockstep edit. Two live config yaml file entries removed from the `DeprecatedPaths` slice; the pinned
count tests and doc-comment counts updated in the SAME commit so no intermediate commit is red.

### Exact lockstep-update sites (verified)

| File | Site | Change |
|------|------|--------|
| `internal/defs/dirs.go` | lines ~226-237 (2 struct entries) | REMOVE `.moai/config/sections/design.yaml` + `.moai/config/sections/db.yaml` entries |
| `internal/defs/dirs.go` | line 225 `// deprecated config yaml files` | comment now covers 3 files (gate/github-actions/memo) — update wording (cosmetic) |
| `internal/defs/dirs.go` | lines 43-44, 47 (header + `@MX:REASON`) | `43 entries` / `43-entry total` → `41` |
| `internal/defs/dirs_test.go` | line 21 `const want = 43` | → `41` |
| `internal/defs/dirs_test.go` | line 24 error msg `9 Category A + 31 Category B + 3 Category C` | → `9 + 29 + 3` |
| `internal/defs/dirs_test.go` | `TestDeprecatedPathsCategorySplit` `wantCategoryB = 31` | → `29` |
| `internal/defs/dirs_test.go` | `TestDeprecatedPathsCategoryBExpectedEntries` `wantCategoryB` slice (lines 110-111) | REMOVE the 2 config-yaml paths |
| `internal/defs/dirs_test.go` | lines 3-7, 19 (`@MX:ANCHOR`/`@MX:REASON` header) | `43-entry total + 9/31/3` → `41-entry total + 9/29/3`; cite this reconcile SPEC (REQ-DPR-004) |
| `internal/cli/v2_detection.go` | line 22 doc comment `43 entries: Category A 9 +` | → `41 entries` (cosmetic) |

### Self-adjusting sites (verify only — no edit)

- `internal/cli/update_e2e_test.go:100-101` and `internal/cli/update_cleanup_test.go:145-146` assert against
  `len(defs.DeprecatedPaths)` dynamically (no hardcoded 43). They auto-adjust to 41; **re-run to confirm** they
  create/remove one fewer fixture per removed entry and stay green.

### §A.4 count-derivation reconciliation (REQ-DPR-004 — run-phase decision point)

The `dirs_test.go` `@MX:ANCHOR` names origin SPEC §A.4 as the canonical 43-entry derivation. Two options:

- **Option A.2 (recommended)** — redirect the `@MX:ANCHOR`/`@MX:REASON` in `dirs.go` + `dirs_test.go` to state
  the 41-entry count is now governed by SPEC-DEPRECATEDPATHS-RECONCILE-001 (this SPEC), citing origin §A.4 as
  the historical 43-entry derivation. Keeps the completed origin SPEC body immutable. Zero completed-SPEC-body edit.
- **Option A.1 (alternative)** — add a bounded correction note + HISTORY line to origin SPEC §A.4 recording the
  43→41 reconciliation. Honors the literal "update spec.md §A.4 atomically" mandate but touches a completed SPEC
  body (a `manager-spec`-owned edit per the Status Transition Ownership Matrix — would require orchestrator
  re-delegation if performed mid-run).

Recommendation: Option A.2. Surface to the user at Implementation Kickoff Approval if a decision is wanted.

## §F. Milestones

### M1 — Manifest + pinned-test lockstep edit (RED → GREEN)

1. RED: flip `dirs_test.go` assertions (`want` 43→41, `wantCategoryB` 31→29, drop the 2 paths from the
   enumeration slice, update the `@MX` header). Run `go test ./internal/defs/` → RED (slice still has 43).
2. GREEN: remove the 2 struct entries from `dirs.go`; reconcile the `dirs.go` header + `@MX:REASON` counts and
   the `// deprecated config yaml files` comment. Run `go test ./internal/defs/` → GREEN.

### M2 — Reconciliation + full verification

1. Apply the §A.4 count-derivation reconciliation (Option A.2 recommended — REQ-DPR-004).
2. Update the `v2_detection.go:22` doc comment count (cosmetic).
3. Full verification batch: `go test ./...`, `go build ./...`, `GOOS=windows GOARCH=amd64 go build ./...`,
   and the `internal/cli` update tests (self-adjusting, re-run to confirm).

## §D. Constraints (DO NOT VIOLATE)

- PRESERVE: all loaders (`loadDesignSection`, `loadMigrationPatterns`, `LoadDesignConfig`), all template files
  (`internal/template/templates/**`), and all other `DeprecatedPaths` entries (Category A 9, Category C 3,
  `.moai/db`, `.moai/project/brand`, gate/github-actions/memo yamls).
- NO `make build` (no template changed).
- NO change to production Go behavior — manifest + tests + doc comments only.
- Keep the `dirs_test.go` slice and the `dirs.go` slice in lockstep in a SINGLE commit (never a red intermediate).
- Do NOT rewrite origin SPEC REQ-VVCR-011/012/014 (completed SPEC; §A.4 reconciliation only, per Option A.2).
- Conventional Commits; `Authored-By-Agent:` trailer; no `--no-verify`, no `--amend`, no force-push.

## §E. Self-Verification plan (run-phase deliverables)

- E1: AC-DPR-001..009 PASS/FAIL matrix with verbatim command output.
- E2: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- E3: `go test ./internal/defs/... ./internal/cli/...` → ok (incl. the self-adjusting update tests).
- E5: `golangci-lint run --timeout=2m` → no NEW findings vs baseline.
- E6: new commit SHA(s) + push state.

## §G. Anti-Patterns to avoid

- Editing the slice without the test (or vice versa) → red intermediate commit (violates the `@MX:ANCHOR` lockstep contract).
- Removing `.moai/db` or `.moai/project/brand` (they are valid v2 removal targets — NOT in scope).
- Touching `gate.yaml` / `github-actions.yaml` / `memo.yaml` entries (out of scope).
- Removing the template `design.yaml` / `db.yaml` or their loaders (Direction B — rejected).
- Rewriting completed origin SPEC REQs (only §A.4 count-derivation reconciliation, Option A.2 recommended).
- Running `make build` (no template touched).

## §H. Cross-references

- spec.md §1.1 (verified findings), §1.2 (ground-truth-shift framing), §4 (Out of Scope + Direction B rejection).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (§A.4 edit ownership nuance).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` (era V3R6 classification; progress.md §E markers).
