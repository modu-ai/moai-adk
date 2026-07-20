---
id: SPEC-INTERNAL-TEST-004
title: "Regenerate stale doctor/status golden testdata for version bump rc7→rc10 (whole-repo green)"
version: "0.1.0"
status: completed
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0-rc10"
module: "internal/cli/testdata"
lifecycle: spec-anchored
tags: "golden-test, test-fix, debt-cleanup, version-bump"
tier: S
depends_on: []
related_specs: [SPEC-INTERNAL-TEST-003, SPEC-INTERNAL-ARCH-001, SPEC-WEB-CONSOLE-011]
---

# acceptance.md — SPEC-INTERNAL-TEST-004

## §D. AC Matrix

| AC | REQ | Severity | Description |
|----|-----|----------|-------------|
| AC-001 | REQ-GOLD-001 | MUST | `TestDoctor_Current_Light` PASS |
| AC-002 | REQ-GOLD-001 | MUST | `TestDoctor_Current_Dark` PASS |
| AC-003 | REQ-GOLD-001 | MUST | `TestDoctor_NoColor` PASS |
| AC-004 | REQ-GOLD-001 | MUST | `TestStatus_Current_Light` PASS |
| AC-005 | REQ-GOLD-001 | MUST | `TestStatus_Current_Dark` PASS |
| AC-006 | REQ-GOLD-001 | MUST | `TestStatus_NoColor` PASS |
| AC-007 | REQ-GREEN-001 | MUST | Whole-repo `go test ./...` exit 0 |
| AC-008 | REQ-VER-001 | MUST | `git diff internal/cli/testdata/` shows only version-string line changes |
| AC-009 | REQ-PRESERVE-001, REQ-SCOPE-001 | MUST | PRESERVE verified — no path outside `internal/cli/testdata/` modified; statusline emoji edits remain uncommitted |

### Severity allocation rationale
- AC-001..006 (MUST): the 6 golden tests are the direct failure surface — each must individually PASS.
- AC-007 (MUST): whole-repo exit 0 is the SPEC's headline goal and the ARCH-001 unblock condition.
- AC-008 (MUST): the git-diff check guards against unintended golden mutation (e.g., a section-rendering regression masquerading as a version bump).
- AC-009 (MUST): the PRESERVE check guards against scope creep into WEB-CONSOLE-011 territory or unrelated working-tree paths.

---

## §D.1 AC Detail (Given-When-Then)

### AC-001 — `TestDoctor_Current_Light` PASS

**Given** the golden testdata at `internal/cli/testdata/doctor-light.golden` is regenerated to reflect `pkg/version/version.go` = `v3.0.0-rc9`,
**When** `go test ./internal/cli/ -run TestDoctor_Current_Light -count=1` is executed,
**Then** the test exits 0 (PASS — golden output matches actual doctor rendering at rc9).

**Evidence path**: `.moai/state/verify/<session>/test-004-ac001-doctor-light.log`

### AC-002 — `TestDoctor_Current_Dark` PASS

**Given** the golden testdata at `internal/cli/testdata/doctor-dark.golden` is regenerated to rc9,
**When** `go test ./internal/cli/ -run TestDoctor_Current_Dark -count=1` is executed,
**Then** the test exits 0.

**Evidence path**: `.moai/state/verify/<session>/test-004-ac002-doctor-dark.log`

### AC-003 — `TestDoctor_NoColor` PASS

**Given** the golden testdata at `internal/cli/testdata/doctor-nocolor.golden` is regenerated to rc9,
**When** `go test ./internal/cli/ -run TestDoctor_NoColor -count=1` is executed,
**Then** the test exits 0.

**Evidence path**: `.moai/state/verify/<session>/test-004-ac003-doctor-nocolor.log`

### AC-004 — `TestStatus_Current_Light` PASS

**Given** the golden testdata at `internal/cli/testdata/status-light.golden` is regenerated to rc9,
**When** `go test ./internal/cli/ -run TestStatus_Current_Light -count=1` is executed,
**Then** the test exits 0.

**Evidence path**: `.moai/state/verify/<session>/test-004-ac004-status-light.log`

### AC-005 — `TestStatus_Current_Dark` PASS

**Given** the golden testdata at `internal/cli/testdata/status-dark.golden` is regenerated to rc9,
**When** `go test ./internal/cli/ -run TestStatus_Current_Dark -count=1` is executed,
**Then** the test exits 0.

**Evidence path**: `.moai/state/verify/<session>/test-004-ac005-status-dark.log`

### AC-006 — `TestStatus_NoColor` PASS

**Given** the golden testdata at `internal/cli/testdata/status-nocolor.golden` is regenerated to rc9,
**When** `go test ./internal/cli/ -run TestStatus_NoColor -count=1` is executed,
**Then** the test exits 0.

**Evidence path**: `.moai/state/verify/<session>/test-004-ac006-status-nocolor.log`

### AC-007 — Whole-repo `go test ./...` exit 0

**Given** the 6 golden testdata files are regenerated to rc9 and no other package has a failing test,
**When** `go test ./...` is executed against the whole repository,
**Then** the command exits 0 with 0 FAIL packages (93+ ok packages expected), unblocking `SPEC-INTERNAL-ARCH-001` plan-audit M0 whole-repo-green precondition.

**Evidence path**: `.moai/state/verify/<session>/test-004-ac007-wholerepo.log` (verbatim output, exit code observed)

### AC-008 — Golden diff is version-string-only

**Given** the 6 golden testdata files have been regenerated via `UPDATE_GOLDEN=1`,
**When** `git diff internal/cli/testdata/` is executed,
**Then** the diff shows exactly 6 files changed, 6 insertions, 6 deletions, and the only changed content is the version-string line (`v3.0.0-rc7` → `v3.0.0-rc9`) — verifiable via `git diff internal/cli/testdata/ | grep -E '^[-+].*rc[0-9]' | sort -u` returning exactly 4 lines (2 `-` rc7 patterns + 2 `+` rc9 patterns).

**Evidence path**: `.moai/state/verify/<session>/test-004-ac008-golden-diff.log` (the `--stat` + the filtered `grep` output)

### AC-009 — PRESERVE verified

**Given** the run-phase commit(s) have been made,
**When** `git status --short` and `git diff HEAD~1 --name-only` are executed,
**Then** the only paths modified by TEST-004 are under `internal/cli/testdata/` (6 golden files), AND the pre-existing working-tree paths listed in `plan.md` §D.2 are unchanged in the commit (specifically: `internal/statusline/renderer.go` + `cache_hit_test.go` remain uncommitted with the ♻️ working-tree edit; `pkg/version/version.go` is untouched at rc9; no PRESERVE path is swept into the commit).

**Evidence path**: `.moai/state/verify/<session>/test-004-ac009-preserve.log` (`git show --stat HEAD` for the commit + `git status --short` showing PRESERVE paths still in their pre-existing state)

---

## §D.2 Indirect Verification (SHOULD — supporting evidence)

| Check | Command | Purpose |
|-------|---------|---------|
| Doctor full-suite green | `go test ./internal/cli/ -run TestDoctor -count=1` | Broader than AC-001..003 (catches non-`Current_*` doctor variants if any) |
| Status full-suite green | `go test ./internal/cli/ -run TestStatus -count=1` | Broader than AC-004..006 |
| `go vet ./internal/cli/...` | `go vet ./internal/cli/...` | Static-analysis sanity (no source change, but confirms package health) |
| version.go unchanged | `git diff HEAD~1 -- pkg/version/version.go` | Empty diff confirms version.go frozen at rc9 |

---

## §D.3 Edge Cases

| Edge case | Expected behavior |
|-----------|-------------------|
| `UPDATE_GOLDEN=1` run with too-broad `-run` filter | MUST use `-run 'TestDoctor_|TestStatus_'` (or the 6 explicit test names); a bare `UPDATE_GOLDEN=1 go test ./internal/cli/` could regenerate unrelated goldens if any are added in the future — the targeted filter prevents this |
| Golden diff contains MORE than the version string | FAIL AC-008 — indicates a section-rendering regression (not a version bump); halt and return blocker report (do not commit) |
| `internal/statusline/` tests fail after TEST-004 commit | Indicates the emoji working-tree edit was accidentally swept into the commit — FAIL AC-009; revert and re-commit with specific path |
| version.go at a different rc value at run time | Re-baseline: the golden must match whatever `pkg/version/version.go` says at run time (single source of truth); re-run research §2 if the value differs from rc9 |
| Parallel session commits version.go bump during TEST-004 run | Pre-spawn sync check (`agent-common-protocol.md` § Pre-Spawn Sync Check) should catch this; if it slips through, the golden regen produces the wrong version string — re-baseline against the new HEAD version.go |

---

## §D.4 Closure Gates (Definition of Done)

> **Closure status (2026-07-09, via external absorption `ce2a509dc`).** See spec.md §G for the full resolution record. The golden regeneration TEST-004 was authored to specify was applied by commit `ce2a509dc` (the rc10 version-bump commit), NOT by a TEST-004 own run-phase commit.

- [x] AC-001 through AC-006 **SATISFIED via `ce2a509dc`** (6 golden tests PASS — `internal/cli` goldens regenerated to rc10, matching `pkg/version/version.go`)
- [ ] AC-007 **DEBT-TRANSFERRED to SPEC-AGENT-ARCH-V2-001** — whole-repo NOT exit 0; the 3 remaining `internal/template` FAILs (`TestAllAgentsInCatalog`, `TestTemplateNoInternalContentLeak`, `TestRuleProvenanceAudit`) belong to the super-advisor agent integration, NOT to the golden drift
- [x] AC-008 **SATISFIED via `ce2a509dc`** (golden diff is version-string-only: rc7→rc10, confirming REQ-VER-001)
- [x] AC-009 **HOLDS** — no TEST-004 run-phase commit touched any PRESERVE path; the 6 goldens are committed inside `ce2a509dc` and PRESERVE paths in plan.md §D.2 remain in their pre-existing state
- [ ] Evidence persisted under `.moai/state/verify/<session>/` — N/A (no own run-phase; `ce2a509dc` is the evidence)
- [x] Commit + pushed to main — `ce2a509dc` is committed + tagged `v3.0.0-rc10` (NOT a TEST-004 commit, but carries the golden regen)
- [ ] SPEC-INTERNAL-ARCH-001 M0 whole-repo-green precondition is NOT unblockable by TEST-004 — debt-transferred to SPEC-AGENT-ARCH-V2-001 (super-advisor integration surface)

---

## §D.5 Forward-Looking Checks (post-close hygiene)

| Check | Purpose | Status |
|-------|---------|--------|
| Next version bump (rc11+) must regenerate goldens in the SAME commit | Prevents this drift from recurring; consider a CI guard that checks golden version string == version.go Version in a follow-up SPEC | **RETROACTIVELY VALIDATED by `ce2a509dc`** — the rc10 bump did exactly this (regenerated the 6 goldens in the same commit as the version bump), confirming the check's value as a drift-prevention invariant |
| TEST-003 AC-006 debt entry can be marked fully resolved | TEST-003 progress.md already marks it as external-debt transferred; TEST-004 close completes the resolution chain | **RESOLVED** — `ce2a509dc` applied the golden regen TEST-003 AC-006 was waiting on |
