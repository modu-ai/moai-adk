---
id: SPEC-V3R6-LIFECYCLE-SYNC-GATE-001
artifact: acceptance
version: "0.1.3"
created: 2026-05-25
updated: 2026-05-26
---

## HISTORY

### v0.1.3 (2026-05-26, manager-spec — iter-5 narrow-scope residual defect resolution)
- No AC body changes (v0.1.3 amendment is plan.md F.1/F.5 reverse-traceability symmetric restoration + spec.md F.4 Description path correction). acceptance.md §D.3 remains the canonical SSOT.
- AC count unchanged at 22 (15 functional + 5 NFR + 1 dogfood + 1 backfill-only).
- Cross-reference to spec.md v0.1.3 HISTORY for full D1/D2 residual defect catalogue.

### v0.1.2 (2026-05-26, manager-spec — iter-3 narrow-scope defect resolution)
- D1 BLOCKING resolved (acceptance.md side): AC-LSG-018 Given/When/Then triplet rewritten from active backfill dogfood (5 modern-era violations resolved) to no-op regression validation (5 already-discharged SPECs at `status: completed`, verify no-op success path). New AC-LSG-018 binds to AC-LSG-022's `fully-completed-noop` fixture state (parametric `TestBackfillOnlyVariants` last variant) — no new fixture or AC required; AC count stays at 22.
- No changes to AC-LSG-001..017, AC-LSG-019, AC-LSG-020, AC-LSG-021, AC-LSG-022 — narrow-scope amendment per Path A
- Cross-reference to spec.md v0.1.2 HISTORY for full D1/D2/D3 defect catalogue and trajectory analysis

### v0.1.1 (2026-05-25, manager-spec — iter-2 narrow-scope defect resolution)
- D1 BLOCKING resolved: added AC-LSG-019 (Cross-Platform binding NFR-LSG-003) + AC-LSG-020 (Observability binding NFR-LSG-004) — NFR coverage 3/5 → 5/5 (100%)
- D3 MUST-FIX resolved: DoD math reconciled to "21 total" = 15 functional + 1 M6 dogfood + 5 NFR
- D4 SHOULD-FIX resolved: AC-LSG-018a renumbered to AC-LSG-021 (5 occurrences updated)
- D7 SHOULD-FIX resolved: added AC-LSG-022 specifying `--backfill-only` semantics (input precondition + operation + exit codes + verification command)
- D5 MUST-FIX resolved: AC-LSG-004 Then clause amended to bind on `Authored-By-Agent: manager-develop` trailer convention (defined in spec.md §D.1.6 HARD)
- Total AC count: 19 → 22 (+3 net new; existing IDs preserved except AC-018a → AC-021 renumber)
- Cross-reference to spec.md v0.1.1 HISTORY for full D1-D10 defect catalogue

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

### A.2 Non-Functional AC Items (NFR-LSG-001..005 binding) — 5/5 coverage post-v0.1.1 D1 fix

| AC ID | NFR binding | Severity | Verification Method |
|-------|-------------|----------|---------------------|
| AC-LSG-016 | NFR-LSG-001 | SHOULD-PASS | `time moai spec audit` on 200-SPEC fixture < 5s |
| AC-LSG-017 | NFR-LSG-002 | MUST-PASS | Legacy 5-commit close still functional (regression test) |
| AC-LSG-019 | NFR-LSG-003 | MUST-PASS | Cross-Platform — GitHub Actions matrix runs M1 atomic commit verification on macOS AND Linux runners |
| AC-LSG-020 | NFR-LSG-004 | MUST-PASS | Observability — `.moai/logs/lifecycle-close.log` contains ≥ 5 entries after M6 dogfood execution |
| AC-LSG-021 | NFR-LSG-005 | MUST-PASS | Concurrent close test (2 sessions, same SPEC, file lock holds) — formerly AC-LSG-018a, renumbered per v0.1.1 D4 fix |

### A.3 M6 Dogfood AC

| AC ID | M6 binding | Severity | Verification Method |
|-------|-----------|----------|---------------------|
| AC-LSG-018 | M6 dogfood (v0.1.2 reframe: no-op regression) | MUST-PASS | All 5 already-discharged SPECs exit 0 as no-op via `moai spec close --backfill-only` (verifies `fully-completed-noop` fixture state at integration level — see §B.18 v0.1.2 reframe) |

### A.4 Backfill-Only Semantics AC

| AC ID | REQ binding | Severity | Verification Method |
|-------|-------------|----------|---------------------|
| AC-LSG-022 | REQ-LSG-001 (extended) | MUST-PASS | `--backfill-only` flag semantics: input precondition + operation + exit codes + verification command (see §B.20) |

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

**Given** a commit with subject matching `^(feat|fix|chore)\(SPEC-[A-Z0-9-]+\):.*$` AND commit body trailer containing the literal line `Authored-By-Agent: manager-develop` (canonical convention per spec.md §D.1.6 HARD) AND diff containing spec.md frontmatter `status: in-progress → implemented`,
**When** `moai spec lint --include-ownership` is invoked on the commit range,
**Then** the lint engine SHALL emit exactly one `OwnershipTransitionInvalid` finding referencing the offending commit SHA + the canonical owner per the Status Transition Ownership Matrix in `.claude/rules/moai/development/spec-frontmatter-schema.md`.

**Trailer convention (D5 resolution)**: The `Authored-By-Agent` trailer is the mechanical signal consumed by this AC. Trailer values: `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `orchestrator-direct` (lowercase, single-token). Commits without the trailer (legacy / non-MoAI / pre-v3.0.1) are NOT subject to OwnershipTransitionRule and produce no finding (silent SKIP).

**Verification**: `internal/spec/lint_ownership_test.go::TestOwnershipTransitionDetected` — fixture commit with subject + `Authored-By-Agent: manager-develop` trailer + `status: in-progress → implemented` diff. Assertion: exactly one `OwnershipTransitionInvalid` finding emitted referencing the commit SHA. Negative fixture: same diff without trailer → zero findings (silent SKIP).

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

### B.18 AC-LSG-018 — M6 No-Op Regression Validation (5 Already-Discharged SPECs)

> **v0.1.2 reframe per D1 BLOCKING resolution**: Originally specified active backfill dogfood of 5 modern-era violations. After iter-2 PASS verdict (2026-05-25 20:37) but before iter-3 audit (2026-05-26), orchestrator-direct retroactive Mx chores `a1fb04625` (ARR-001) + `8d0b1fdf9` (FCG-001) + `d167eb08b` (TMD-001) + `ac8ba9a99` (TMC-001) + `adc75a33c` (HCW-001 PROCEED-WITH-DEBT) discharged all 5 target SPECs to `status: completed`. M6 scope is reframed to no-op regression validation that exercises the `fully-completed-noop` fixture state of AC-LSG-022's parametric test in production usage. The reframe binds AC-LSG-018 to AC-LSG-022's existing `fully-completed-noop` fixture (acceptance.md §B.22 last variant in `TestBackfillOnlyVariants`); no new fixture is introduced.

**Given** the 5 SPECs originally identified in spec.md §A.1 are now all at `status: completed` (ARR-001 / FCG-001 / HCW-001 / TMD-001 / TMC-001),
**When** M6 milestone executes `moai spec close SPEC-XXX --backfill-only` for each of the 5 SPECs,
**Then** after M6 completion:
1. Each invocation SHALL exit code 0 (no-op success path)
2. Each invocation SHALL produce **0 commits** (no staging change, no commit object created — verifies the implementation handles already-completed precondition gracefully)
3. Each invocation SHALL log a single line entry to `.moai/logs/lifecycle-close.log` with `mode: "backfill-only"`, `result: "success"`, `transitions: {}` (empty object — no fields needed transition), and a structured noop signal (stderr or log body) matching the pattern `noop: already completed` (case-insensitive substring acceptable; implementation-level detail in M1 closer.go)
4. `moai spec audit --filter-era=V3R6 --json` SHALL return 0 entries in `drift_findings` for these 5 SPECs (precondition already satisfied — they are already-completed; verifies the audit tool does not erroneously surface drift for already-completed modern-era SPECs)
5. All 5 SPECs' spec.md frontmatter status SHALL remain `completed` (unchanged — no-op produces no transition)

**Verification**: Post-M6 automated checks:
```bash
# 1. No-op exit code 0 verification (per-SPEC)
moai spec close SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 --backfill-only ; echo $?   # expected: 0
moai spec close SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 --backfill-only ; echo $?    # expected: 0
moai spec close SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 --backfill-only ; echo $?      # expected: 0
moai spec close SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 --backfill-only ; echo $?          # expected: 0
moai spec close SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001 --backfill-only ; echo $?        # expected: 0

# 2. Zero commits produced by no-op invocations
git log --oneline --grep="SPEC-V3R6-LIFECYCLE-SYNC-GATE-001.*M6.*close"                # expected: empty (no commits)

# 3. Log audit trail emitted (per AC-LSG-020 cross-binding)
jq -c 'select(.mode == "backfill-only" and .result == "success" and .transitions == {})' .moai/logs/lifecycle-close.log | wc -l   # expected: ≥ 5

# 4. Audit drift findings empty for these 5 SPECs
moai spec audit --filter-era=V3R6 --json | jq '[.drift_findings[] | select(.spec_id | test("ARR-001|FCG-001|HCW-001|TMD-001|TMC-001"))] | length'   # expected: 0

# 5. SPEC status unchanged (still completed)
for s in ARR-001 FCG-001 HCW-001 TMD-001 TMC-001; do
  grep "^status:" ".moai/specs/SPEC-V3R6-${s}/spec.md"   # expected each: status: completed
done
```

**Cross-binding to AC-LSG-022**: AC-LSG-018 references the `fully-completed-noop` fixture state already enumerated in AC-LSG-022's parametric `TestBackfillOnlyVariants` test (acceptance.md §B.22 fixture list: `Y_N_N_Y`, `Y_Y_N_Y`, `Y_Y_Y_Y_StatusDrift`, **`fully-completed-noop`**). M6 production execution against the 5 already-completed SPECs exercises the same fixture state at the integration level; AC-LSG-022 covers the unit-test level. No new fixture is introduced.

### B.19 AC-LSG-021 — Concurrent Close Safety (NFR-LSG-005)

> **Renumbered per v0.1.1 D4 fix**: previously AC-LSG-018a; renumbered to AC-LSG-021 to follow standard sequential ID format after D1 introduced AC-LSG-019/020.

**Given** two sessions concurrently attempt `moai spec close SPEC-FOO-001`,
**When** both processes execute simultaneously,
**Then** exactly one SHALL succeed (atomic commit produced) AND the other SHALL fail with lock-held error AND post-execution `git log` SHALL show exactly ONE close commit (not two, not zero).

**Verification**: Same harness as AC-LSG-010, additionally verify post-execution git state.

### B.20 AC-LSG-019 — Cross-Platform GitHub Actions Matrix (NFR-LSG-003)

**Given** the M1 milestone commits the `moai spec close` atomic commit verification test (`internal/cli/spec_close_test.go::TestAtomicClose`),
**When** the CI workflow runs on `macos-latest` AND `ubuntu-latest` GitHub Actions runners,
**Then** the test SHALL pass on BOTH platforms AND the workflow matrix configuration SHALL include both `os: macos-latest` and `os: ubuntu-latest` entries. Windows support is best-effort per NFR-LSG-003 wording; failure on Windows runner SHALL NOT block the matrix (continue-on-error allowed for `windows-latest`).

**Verification**: Two parallel CI commands:
```bash
# Inspect workflow file
grep -E 'os:\s+(macos-latest|ubuntu-latest)' .github/workflows/*.yml | sort -u | wc -l  # expected: ≥ 2 (both OS entries present)

# Inspect latest CI run after M1 merge
gh run list --workflow=ci.yml --branch=main --limit=1 --json conclusion,name,jobs | jq '.[].jobs[] | select(.name | contains("macos") or contains("ubuntu")) | {name, conclusion}'
# expected: both jobs conclusion: "success"
```

### B.21 AC-LSG-020 — Observability Audit Trail (NFR-LSG-004)

**Given** the M6 dogfood milestone executes `moai spec close SPEC-XXX --backfill-only` on each of the 5 known modern-era violations,
**When** the M6 milestone completes,
**Then** the file `.moai/logs/lifecycle-close.log` SHALL exist AND contain ≥ 5 entries (one per closed SPEC). Each entry SHALL be a single JSON line containing `timestamp` (RFC3339), `spec_id`, `mode` (`full-close|backfill-only`), `transitions` (object listing changed fields), `commit_sha`, `result` (`success|failure`), `duration_ms`.

**Verification**: Post-M6 commands:
```bash
test -f .moai/logs/lifecycle-close.log                                              # expected: exit 0
wc -l .moai/logs/lifecycle-close.log | awk '{print $1}'                              # expected: ≥ 5
jq -c 'select(.result == "success") | .spec_id' .moai/logs/lifecycle-close.log | sort -u | wc -l  # expected: ≥ 5 (5 distinct successful closes)
```

### B.22 AC-LSG-022 — Backfill-Only Mode Semantics (REQ-LSG-001 extended, D7 fix)

**Given** a SPEC with the following partial state (any subset of the following missing fields):
- spec.md `status:` ≠ `completed` (typically `implemented`)
- progress.md `§E.2 sync_commit_sha:` missing or empty (Y_N_N_Y or Y_Y_N_Y drift cases)
- progress.md `§E.5 mx_commit_sha:` missing or empty
- progress.md `§E.3 status:` ≠ `completed`

AND progress.md `§E.2 sync section` is present (sync-phase has run) AND progress.md `§E.5 mx section` is present (Mx-phase has run),

**When** the user invokes `moai spec close SPEC-XXX --backfill-only`,

**Then** the CLI SHALL:
1. Atomically transition ONLY the missing fields among the 4 above (leaving existing fields untouched)
2. Produce exactly one commit with subject `chore(SPEC-XXX): 4-phase close — atomic (backfill-only)` and body listing transitioned fields
3. Exit code 0 on success
4. Exit code 1 on precondition failure (missing sync or mx section in progress.md) with stderr identifying the missing precondition
5. Exit code 2 on lock contention (same as full close mode)

**Verification**:
```bash
# Dry-run on a known Y_Y_Y_Y drift fixture
moai spec close SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 --backfill-only --dry-run
# expected stdout: dry-run summary listing 1 transition (spec.md status: implemented → completed) + exit 0
```

Plus parametric unit test `internal/cli/spec_close_test.go::TestBackfillOnlyVariants` iterating 4 fixture states (Y_N_N_Y, Y_Y_N_Y, Y_Y_Y_Y_StatusDrift, fully-completed-noop), asserting per-fixture exit code + commit subject suffix + transitioned field set.

---

## C. Definition of Done

A milestone is "Done" when:
1. All MUST-PASS AC items binding to milestone deliverables PASS
2. `go test ./...` exits 0 (full suite)
3. `golangci-lint run --timeout=2m` exits 0
4. Coverage for new/modified packages meets D.2 SHOULD thresholds (warn-only)
5. CHANGELOG entry drafted (sync-phase will commit)
6. Commit attribution `feat|docs|fix(SPEC-V3R6-LIFECYCLE-SYNC-GATE-001): Mn ...` clean

SPEC 4-phase close is "Done" when (v0.1.1 post-D3 fix):
1. All **15 functional AC items + 1 M6 dogfood AC + 5 NFR AC items + 1 backfill-only AC = 22 total** PASS
   - Functional (15): AC-LSG-001..015
   - M6 dogfood (1): AC-LSG-018
   - NFR (5): AC-LSG-016 (NFR-001) + AC-LSG-017 (NFR-002) + AC-LSG-019 (NFR-003 NEW) + AC-LSG-020 (NFR-004 NEW) + AC-LSG-021 (NFR-005, renumbered from AC-018a)
   - Backfill-only (1): AC-LSG-022 (REQ-LSG-001 extended, NEW per D7)
2. M6 dogfood verifies 0 drift findings post-execution
3. sync-phase chore committed + audit-ready signal §E.4 emitted
4. Mx-phase chore committed + audit-ready signal §E.5 emitted
5. spec.md status: completed via `moai spec close` (eating own dog food — meta-dogfood)

---

## D. Bidirectional Traceability Matrix

### D.1 REQ → AC Mapping (v0.1.1 — 5/5 NFR coverage + AC-022 backfill-only)

| REQ | Bound AC |
|-----|----------|
| REQ-LSG-001 | AC-LSG-001, AC-LSG-022 (backfill-only mode extension per D7) |
| REQ-LSG-002 | AC-LSG-002 |
| REQ-LSG-003 | AC-LSG-003 |
| REQ-LSG-004 | AC-LSG-004 |
| REQ-LSG-005 | AC-LSG-005 |
| REQ-LSG-006 | AC-LSG-006, AC-LSG-014 |
| REQ-LSG-007 | AC-LSG-007 |
| REQ-LSG-008 | AC-LSG-008 |
| REQ-LSG-009 | AC-LSG-009 |
| REQ-LSG-010 | AC-LSG-010, AC-LSG-021 (formerly AC-018a) |
| REQ-LSG-011 | AC-LSG-011 |
| REQ-LSG-012 | AC-LSG-012 |
| REQ-LSG-013 | AC-LSG-013 |
| REQ-LSG-014 | AC-LSG-014 |
| REQ-LSG-015 | AC-LSG-015 |
| NFR-LSG-001 | AC-LSG-016 |
| NFR-LSG-002 | AC-LSG-017 |
| NFR-LSG-003 | AC-LSG-019 (NEW per D1) |
| NFR-LSG-004 | AC-LSG-020 (NEW per D1) |
| NFR-LSG-005 | AC-LSG-021 (formerly AC-018a per D4 renumber) |

### D.2 AC → REQ Reverse Mapping (v0.1.1)

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
| AC-LSG-019 | NFR-LSG-003 (Cross-Platform — NEW per D1) |
| AC-LSG-020 | NFR-LSG-004 (Observability — NEW per D1) |
| AC-LSG-021 | NFR-LSG-005, REQ-LSG-010 (formerly AC-018a per D4 renumber) |
| AC-LSG-022 | REQ-LSG-001 (backfill-only mode extension — NEW per D7) |

### D.3 AC → Milestone Mapping (v0.1.1)

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
| AC-LSG-019 | M1 (CI workflow matrix configured at M1 commit + verified by M1 PR check) |
| AC-LSG-020 | M1 (log emission code path) + M6 (verifies ≥ 5 entries after dogfood) |
| AC-LSG-021 | M1 (formerly AC-018a per D4 renumber) |
| AC-LSG-022 | M1 (Close core), M2 (CLI flag wiring), M6 (--backfill-only used by all 5 dogfood closes) |

**Coverage Summary (v0.1.1)**: 100% bidirectional traceability — every REQ + every NFR binds to ≥ 1 AC, and every AC traces back to ≥ 1 REQ/NFR. All 6 milestones (M1-M6) referenced by ≥ 1 AC. NFR coverage 5/5 (100%, up from 3/5 in v0.1.0). Total AC count: 22 (15 functional + 1 M6 dogfood + 5 NFR + 1 backfill-only).
