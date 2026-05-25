---
id: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001
artifact: acceptance
version: "0.1.0"
created: 2026-05-25
updated: 2026-05-25
---

## HISTORY

### v0.1.0 (2026-05-25, manager-spec)
- Initial acceptance criteria authored
- 15 AC-LSG items binding 1:1 to REQ-LSG-001..015 (functional requirements)
- 3 NFR-binding AC items (AC-LSG-016, AC-LSG-017, AC-LSG-018a)
- 1 M6 dogfood verification AC (AC-LSG-018)

---

## A. Acceptance Criteria Matrix

### A.1 Functional AC Items (REQ-LSG-001..015 binding)

| AC ID | REQ binding | Severity | Verification Method |
|-------|-------------|----------|---------------------|
| AC-LSG-001 | REQ-LSG-001 | MUST-PASS | `moai spec close` integration test verifying single atomic commit |
| AC-LSG-002 | REQ-LSG-002 | MUST-PASS | `moai spec audit` unit test classifying 5 era buckets |
| AC-LSG-003 | REQ-LSG-003 | MUST-PASS | Pre-commit hook smoke test on mismatch scenario |
| AC-LSG-004 | REQ-LSG-004 | MUST-PASS | spec-lint extension unit test on manager-develop direct transition |
| AC-LSG-005 | REQ-LSG-005 | MUST-PASS | Rule file presence + content check |
| AC-LSG-006 | REQ-LSG-006 | MUST-PASS | `moai spec close` precondition validation test |
| AC-LSG-007 | REQ-LSG-007 | MUST-PASS | `moai spec audit --json` schema validation |
| AC-LSG-008 | REQ-LSG-008 | MUST-PASS | Pre-commit hook canonical subject enforcement test |
| AC-LSG-009 | REQ-LSG-009 | MUST-PASS | Y_Y_Y_Y drift finding emission test |
| AC-LSG-010 | REQ-LSG-010 | MUST-PASS | File lock contention test |
| AC-LSG-011 | REQ-LSG-011 | MUST-PASS | Hook source code grep for AskUserQuestion absence |
| AC-LSG-012 | REQ-LSG-012 | MUST-PASS | lint.skip opt-out behavior test |
| AC-LSG-013 | REQ-LSG-013 | MUST-PASS | EraAutoDetected informational finding test |
| AC-LSG-014 | REQ-LSG-014 | MUST-PASS | Precondition failure abort + exit 1 + no staging test |
| AC-LSG-015 | REQ-LSG-015 | MUST-PASS | Pre-commit hook mismatch exit 2 + JSON output test |

### A.2 Non-Functional AC Items (NFR-LSG-001..005 binding)

| AC ID | NFR binding | Severity | Verification Method |
|-------|-------------|----------|---------------------|
| AC-LSG-016 | NFR-LSG-001 | SHOULD-PASS | `time moai spec audit` on 200-SPEC fixture < 5s |
| AC-LSG-017 | NFR-LSG-002 | MUST-PASS | Legacy 5-commit close still functional (regression test) |
| AC-LSG-018a | NFR-LSG-005 | MUST-PASS | Concurrent close test (2 sessions, same SPEC, file lock holds) |

### A.3 M6 Dogfood AC

| AC ID | M6 binding | Severity | Verification Method |
|-------|-----------|----------|---------------------|
| AC-LSG-018 | M6 dogfood | MUST-PASS | All 5 known modern-era violations resolved via `moai spec close --backfill-only` |

---

## B. AC Detailed Specifications

### B.1 AC-LSG-001 — Atomic Single-Commit Close

**Given** a SPEC at status: implemented with progress.md §E.2 sync section + §E.5 Mx section both present AND all AC PASS,
**When** the user invokes `moai spec close SPEC-XXX`,
**Then** the CLI SHALL produce exactly one commit whose tree includes:
1. spec.md frontmatter `status: completed`
2. progress.md `§E.3 status: completed`
3. progress.md `§E.2 sync_commit_sha: <SHA>` (atomic backfill of sync commit)
4. progress.md `§E.5 mx_commit_sha: <SHA>` (atomic backfill of mx commit; self-reference SHA via post-staging amend OR `--allow-empty-message` placeholder; design.md §D specifies the mechanism)
5. spec.md `§A Lifecycle Sync` row updated with both SHAs

**And** the commit subject SHALL match `^chore\(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001\): 4-phase close — atomic$` (literal subject pattern, parameterized by SPEC-ID at runtime).

**Verification**: `internal/cli/spec_close_test.go::TestAtomicClose` — fixture SPEC with sync+mx done, run `moai spec close FIXTURE-001`, assert `git log -1 --name-only` shows exactly spec.md + progress.md, parse both files, verify all 5 state transitions present.

### B.2 AC-LSG-002 — Era Classification 5 Buckets

**Given** the audit tool is invoked against the project's `.moai/specs/` directory,
**When** `moai spec audit --json` runs,
**Then** every `.moai/specs/SPEC-*/` directory SHALL appear in the output classified into exactly one of: `V2.x`, `V3R2-R4`, `V3R5`, `V3R6`, `unclassified`.

**Verification**: `internal/spec/audit_test.go::TestEraClassification` — fixture directory with 5 SPECs (one per era), assert each emits the expected era field.

### B.3 AC-LSG-003 — Pre-Commit Hook Mismatch Detection

**Given** a staged commit modifying spec.md `status: completed` but progress.md `§E.3 status` is still `in-progress`,
**When** the pre-commit hook executes,
**Then** the hook SHALL exit with code 2 AND emit JSON output containing `"continue": false, "stopReason": "spec.md/progress.md status field mismatch", "details": {...}`.

**Verification**: Bash test harness invoking `handle-pre-commit-spec-status.sh` with fixture stdin JSON describing the mismatch, assert exit code = 2, assert stdout JSON shape.

### B.4 AC-LSG-004 — Spec-Lint Ownership Transition Detection

**Given** a commit with subject matching `^(feat|fix|chore)\(SPEC-[A-Z0-9-]+\):.*$` AND author trailer indicating `manager-develop` AND diff containing spec.md frontmatter `status: in-progress → implemented`,
**When** `moai spec lint --include-ownership` is invoked on the commit range,
**Then** the lint engine SHALL emit one `OwnershipTransitionInvalid` finding referencing the offending commit SHA + the matrix rule from spec-frontmatter-schema.md.

**Verification**: `internal/spec/lint_ownership_test.go::TestOwnershipTransitionDetected` — fixture commit with the described signature, assert finding emission.

### B.5 AC-LSG-005 — Rule File Authored

**Given** the M5 milestone is complete,
**When** `ls .claude/rules/moai/workflow/lifecycle-sync-gate.md` is run,
**Then** the file SHALL exist AND contain ≥ 250 lines AND contain the literal section headings: `## Era Classification Heuristic`, `## Grandfather Clause Policy`, `## Frontmatter Era Field Semantics`, `## Status Transition Ownership Matrix Cross-Reference`, `## Worked Example: Era Auto-Detection`.

**Verification**: `grep -c "^## " .claude/rules/moai/workflow/lifecycle-sync-gate.md` ≥ 5 AND `wc -l .claude/rules/moai/workflow/lifecycle-sync-gate.md` ≥ 250.

### B.6 AC-LSG-006 — Precondition Matrix Validation

**Given** a SPEC at status: implemented but progress.md §E.5 Mx section absent,
**When** the user invokes `moai spec close SPEC-XXX`,
**Then** the CLI SHALL exit with code 1 AND emit `Error: precondition not met — missing §E.5 Mx-phase audit-ready signal in progress.md` to stderr AND stage NO file changes.

**Verification**: `internal/cli/spec_close_test.go::TestPreconditionMissingMx` — fixture SPEC without Mx section, run close, assert exit code 1, assert `git status --porcelain` shows no staged changes.

### B.7 AC-LSG-007 — Audit JSON Output Schema

**Given** the audit tool is invoked,
**When** `moai spec audit --json` runs,
**Then** the stdout SHALL parse as valid JSON conforming to schema:
```json
{
  "audited_at": "RFC3339 timestamp",
  "total_specs": <integer>,
  "grandfathered": <integer>,
  "modern_era_clean": <integer>,
  "drift_findings": [
    {"spec_id": "SPEC-...", "era": "V3R6", "finding_type": "Y_N_N_Y|Y_Y_N_Y|Y_Y_Y_Y_StatusDrift", "severity": "MUST-FIX|INFO", "remediation": "<command>"}
  ]
}
```

**Verification**: `internal/cli/spec_audit_test.go::TestJSONSchema` — invoke command, parse stdout via `encoding/json`, assert all fields present and types correct.

### B.8 AC-LSG-008 — Canonical 4-Phase Close Subject Enforcement

**Given** a commit subject `chore(SPEC-V3R6-FOO-001): 4-phase close — atomic` AND the staged diff does NOT include spec.md frontmatter `status: completed`,
**When** the pre-commit hook executes,
**Then** the hook SHALL exit with code 2 AND emit JSON with `"stopReason": "canonical 4-phase close subject requires spec.md status: completed in diff"`.

**Verification**: Bash test harness with fixture diff lacking spec.md status, assert exit code 2 + JSON shape.

### B.9 AC-LSG-009 — Y_Y_Y_Y Drift Finding Emission

**Given** a SPEC with progress.md §E.2 sync section + §E.5 Mx section + sync_commit_sha + mx_commit_sha all present BUT spec.md `status: implemented` (not `completed`),
**When** `moai spec audit --json --filter-era=V3R6` runs,
**Then** the output SHALL contain a `drift_findings` entry with `finding_type: "Y_Y_Y_Y_StatusDrift"`, `severity: "MUST-FIX"`, `remediation: "moai spec close SPEC-XXX --backfill-only"`.

**Verification**: `internal/spec/audit_test.go::TestY4StatusDriftDetection` — fixture matching the scenario, assert finding emission.

### B.10 AC-LSG-010 — File Lock Contention

**Given** two concurrent invocations of `moai spec close SPEC-FIXTURE-001`,
**When** both processes attempt to acquire `.moai/state/spec-close-SPEC-FIXTURE-001.lock` simultaneously,
**Then** exactly one process SHALL acquire the lock AND complete the close AND the other process SHALL exit with code 1 + message `Error: another close operation in progress (lock held)`.

**Verification**: Go test using `t.Parallel()` + `sync.WaitGroup` to spawn 2 goroutines invoking the close function on the same SPEC fixture, assert exactly one success + one lock-error result.

### B.11 AC-LSG-011 — Hook AskUserQuestion Absence

**Given** the pre-commit hook script `handle-pre-commit-spec-status.sh` is authored,
**When** the script source is grepped for AskUserQuestion references,
**Then** the command `grep -E 'AskUserQuestion|mcp__askuser' .claude/hooks/moai/handle-pre-commit-spec-status.sh` SHALL return 0 matches (excluding lines starting with `#` comments that reference the rule cross-link).

**Verification**: Bash command in CI/test harness.

### B.12 AC-LSG-012 — Lint Skip Opt-Out

**Given** a SPEC frontmatter containing `lint.skip: [OwnershipTransitionInvalid]`,
**When** `moai spec lint --include-ownership` runs against that SPEC,
**Then** no `OwnershipTransitionInvalid` finding SHALL be emitted; an `OwnershipTransitionSkipped` informational finding SHALL be emitted instead.

**Verification**: `internal/spec/lint_ownership_test.go::TestSkipOptOut` — fixture SPEC with lint.skip entry + commit signature matching INVALID pattern, assert no INVALID finding + presence of Skipped finding.

### B.13 AC-LSG-013 — Era Auto-Detection

**Given** a SPEC frontmatter WITHOUT the optional `era:` field,
**When** `moai spec audit --json` runs,
**Then** the audit output SHALL include `era: "<detected>"` AND a `findings` entry with `finding_type: "EraAutoDetected"`, `severity: "INFO"`, `details: {"heuristic_matched": "<rule-name>"}`.

**Verification**: `internal/spec/audit_test.go::TestEraAutoDetection` — fixture without era field matching V3R6 heuristic, assert era=V3R6 + INFO finding.

### B.14 AC-LSG-014 — Precondition Abort Atomicity

**Given** any of the 4 close preconditions is missing (sync section / mx section / AC PASS / no PASS-WITH-DEBT),
**When** the close command runs,
**Then** exit code SHALL be 1 AND `git status --porcelain` SHALL show no staged changes AND stderr SHALL identify the specific missing precondition by name.

**Verification**: Parametric test in `spec_close_test.go` iterating 4 fixture variants (one missing precondition each), asserting exit code + porcelain output + stderr substring.

### B.15 AC-LSG-015 — Hook Mismatch Exit 2 JSON Shape

**Given** any spec.md/progress.md status mismatch scenario,
**When** the pre-commit hook detects the mismatch,
**Then** exit code SHALL be 2 AND stdout SHALL contain JSON parseable as:
```json
{"continue": false, "stopReason": "<descriptive>", "details": {"spec_id": "SPEC-...", "spec_md_status": "<value>", "progress_md_status": "<value>", "resolution_command": "moai spec close ..."}}
```

**Verification**: Bash test harness with fixture mismatch, parse stdout JSON via `jq`, assert all required fields present and types correct.

### B.16 AC-LSG-016 — Audit Performance Bound

**Given** a fixture project containing 200 dummy SPEC directories,
**When** `time moai spec audit --json` is executed,
**Then** real-time SHALL be ≤ 5 seconds on a baseline development machine (Darwin/Linux x86_64, 2+ cores).

**Verification**: Benchmark test `internal/spec/audit_bench_test.go` using `testing.B` with N=200 fixture SPECs, assert per-op latency converted to total ≤ 5s.

### B.17 AC-LSG-017 — Backward Compatibility (Legacy 5-Commit Close)

**Given** a SPEC closed via the legacy 5-commit cadence (no `moai spec close` involvement),
**When** the SPEC is queried via `moai spec audit --json --include-grandfathered`,
**Then** the SPEC SHALL classify as `era_final: true` (if pre-V3R6) OR `modern_era_clean: true` (if V3R6 + complete) — NO drift finding emitted.

**Verification**: Regression test on actual repo SPECs that were closed pre-this-SPEC (e.g., SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 closed via legacy cadence), assert clean audit result.

### B.18 AC-LSG-018 — M6 Dogfood: 5 Known Violations Resolved

**Given** the 5 known modern-era violations listed in spec.md §A.1,
**When** M6 milestone executes `moai spec close SPEC-XXX --backfill-only` for each,
**Then** after M6 completion:
1. `moai spec audit --filter-era=V3R6 --json` SHALL return 0 entries in `drift_findings`
2. Each of the 5 SPECs SHALL have spec.md `status: completed`
3. Each close SHALL have produced exactly one atomic chore commit attributable to SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 M6

**Verification**: Post-M6 manual + automated check: `moai spec audit --filter-era=V3R6 --json | jq '.drift_findings | length'` returns 0; `git log --oneline --grep="SPEC-V3R6-LIFECYCLE-SYNC-GATE-001.*M6"` returns ≥ 5 commits.

### B.19 AC-LSG-018a — Concurrent Close Safety (NFR-LSG-005)

**Given** two sessions concurrently attempt `moai spec close SPEC-FOO-001`,
**When** both processes execute simultaneously,
**Then** exactly one SHALL succeed (atomic commit produced) AND the other SHALL fail with lock-held error AND post-execution `git log` SHALL show exactly ONE close commit (not two, not zero).

**Verification**: Same harness as AC-LSG-010, additionally verify post-execution git state.

---

## C. Definition of Done

A milestone is "Done" when:
1. All MUST-PASS AC items binding to milestone deliverables PASS
2. `go test ./...` exits 0 (full suite)
3. `golangci-lint run --timeout=2m` exits 0
4. Coverage for new/modified packages meets D.2 SHOULD thresholds (warn-only)
5. CHANGELOG entry drafted (sync-phase will commit)
6. Commit attribution `feat|docs|fix(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001): Mn ...` clean

SPEC 4-phase close is "Done" when:
1. All 18 functional AC items + 3 NFR AC items PASS (21 total)
2. M6 dogfood verifies 0 drift findings post-execution
3. sync-phase chore committed + audit-ready signal §E.4 emitted
4. Mx-phase chore committed + audit-ready signal §E.5 emitted
5. spec.md status: completed via `moai spec close` (eating own dog food — meta-dogfood)

---

## D. Bidirectional Traceability Matrix

### D.1 REQ → AC Mapping

| REQ | Bound AC |
|-----|----------|
| REQ-LSG-001 | AC-LSG-001 |
| REQ-LSG-002 | AC-LSG-002 |
| REQ-LSG-003 | AC-LSG-003 |
| REQ-LSG-004 | AC-LSG-004 |
| REQ-LSG-005 | AC-LSG-005 |
| REQ-LSG-006 | AC-LSG-006, AC-LSG-014 |
| REQ-LSG-007 | AC-LSG-007 |
| REQ-LSG-008 | AC-LSG-008 |
| REQ-LSG-009 | AC-LSG-009 |
| REQ-LSG-010 | AC-LSG-010, AC-LSG-018a |
| REQ-LSG-011 | AC-LSG-011 |
| REQ-LSG-012 | AC-LSG-012 |
| REQ-LSG-013 | AC-LSG-013 |
| REQ-LSG-014 | AC-LSG-014 |
| REQ-LSG-015 | AC-LSG-015 |
| NFR-LSG-001 | AC-LSG-016 |
| NFR-LSG-002 | AC-LSG-017 |
| NFR-LSG-005 | AC-LSG-018a |

### D.2 AC → REQ Reverse Mapping

| AC | Traces to REQ/NFR |
|----|-------------------|
| AC-LSG-001 | REQ-LSG-001 |
| AC-LSG-002 | REQ-LSG-002 |
| AC-LSG-003 | REQ-LSG-003 |
| AC-LSG-004 | REQ-LSG-004 |
| AC-LSG-005 | REQ-LSG-005 |
| AC-LSG-006 | REQ-LSG-006 |
| AC-LSG-007 | REQ-LSG-007 |
| AC-LSG-008 | REQ-LSG-008 |
| AC-LSG-009 | REQ-LSG-009 |
| AC-LSG-010 | REQ-LSG-010 |
| AC-LSG-011 | REQ-LSG-011 |
| AC-LSG-012 | REQ-LSG-012 |
| AC-LSG-013 | REQ-LSG-013 |
| AC-LSG-014 | REQ-LSG-006, REQ-LSG-014 |
| AC-LSG-015 | REQ-LSG-015 |
| AC-LSG-016 | NFR-LSG-001 |
| AC-LSG-017 | NFR-LSG-002 |
| AC-LSG-018 | M6 dogfood (verifies all REQ in production usage) |
| AC-LSG-018a | NFR-LSG-005, REQ-LSG-010 |

### D.3 AC → Milestone Mapping

| AC | Bound Milestone(s) |
|----|---------------------|
| AC-LSG-001 | M1, M2 |
| AC-LSG-002 | M1, M2 |
| AC-LSG-003 | M3 |
| AC-LSG-004 | M4 |
| AC-LSG-005 | M5 |
| AC-LSG-006 | M1, M2 |
| AC-LSG-007 | M1, M2 |
| AC-LSG-008 | M3 |
| AC-LSG-009 | M1 |
| AC-LSG-010 | M1 |
| AC-LSG-011 | M3 |
| AC-LSG-012 | M4 |
| AC-LSG-013 | M1, M5 |
| AC-LSG-014 | M1, M2 |
| AC-LSG-015 | M3 |
| AC-LSG-016 | M1, M2 |
| AC-LSG-017 | M1 (regression check) |
| AC-LSG-018 | M6 |
| AC-LSG-018a | M1 |

**Coverage Summary**: 100% bidirectional traceability — every REQ binds to ≥1 AC and every AC traces back to ≥1 REQ/NFR. All 6 milestones (M1-M6) referenced by ≥1 AC.
