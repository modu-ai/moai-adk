# Acceptance — SPEC-DOCTOR-STAT-SEAM-001

Card t563 · Tier S. Each AC is binary-testable and cites its GEARS requirement from `spec.md` §2.
Verification evidence: verbatim command output persisted under `.moai/reports/t563/`.

## §D AC Matrix

### AC-SEAM-001 — Both stat sites routed; Lstat survives (REQ: REQ-001)

**Given** the worktree after the M2 seam commit, **When** the grep assertions run:
`grep -n 'os\.Stat(' internal/cli/doctor_codex.go` returns NO matches (direct `os.Stat` count in
the file drops 2 → 0), AND `grep -n 'os\.Lstat(' internal/cli/doctor_codex.go` still reports the
`:441` site, **Then** both sites (`:459`, `:857`) are routed through `osStatFn` while the Lstat
classification call is untouched — the two assertions TOGETHER prove the grep distinguishes Stat
from Lstat.

### AC-SEAM-002 — Characterization precedes the swap (REQ: REQ-005)

**Given** the branch history after M2, **When** `git log --oneline` is read for this SPEC's
commits, **Then** the characterization-test commit (M1) precedes the seam-swap commit (M2), both
messages carry `t563`, AND the M1 tests pass when run against the M1 commit's tree
(`git stash`-free check: scoped run on the checkout of that commit, or the recorded run output
from M1's own execution — the commit graph is the only sequencing witness).

### AC-SEAM-003 — Site A observation: recorded path + injected-result buckets (REQ: REQ-002, REQ-006)

**Given** a test overriding `osStatFn` with a recording function (serial, save → replace →
`t.Cleanup` restore) and mirror fixtures in `t.TempDir()`, **When** the mirror check runs with (a)
an injected `fs.ErrNotExist`, (b) an injected `nil`, (c) an injected non-NotExist error, **Then**
(a) the entry name is appended to `st.dangling`, (b) no bucket changes, (c) `st.indeterminate`
increments — AND in every case the recorded argument equals
`filepath.Join(mirrorDir, entryName)` (the mirror-relative joined path).

### AC-SEAM-004 — Site B observation: recorded path + non-stat classifications (REQ: REQ-003, REQ-006)

**Given** the same recording override and a skill-config fixture exercising every classification,
**When** `codexStaleSkillFinding` runs, **Then** the recorded argument for an absolute entry equals
the declared path (and for a home-relative entry, the expanded path); an injected `fs.ErrNotExist`
drives the per-`Enabled`-state missing arms; an injected `nil` resolves — including a DIRECTORY
fixture (directories resolve by design); AND a relative entry and an oddly-formed entry produce
ZERO recorded stat calls (they are skipped before the stat, current behavior preserved).

### AC-SEAM-005 — Output identity across the swap (REQ: REQ-004)

**Given** the shared fixture set from M1 (real files/dirs for resolve, missing, dangling,
indeterminate — identical inputs on both sides), **When** the doctor runs on the pre-swap tree
(M1 baseline, recorded) and on the post-swap tree (M2), **Then** the problem-finding set, each
finding's text, and the detail strings are IDENTICAL — the M1 characterization suite and the
pre-flight scoped baseline selection both pass unmodified on the swapped tree. Any difference
fails this AC and the change is wrong by definition.

### AC-SEAM-006 — Swap-commit diff is minimal (REQ: REQ-001, REQ-004)

**Given** the M2 seam commit, **When** `git show <seam-sha> -- internal/cli/doctor_codex.go` is
read, **Then** the diff touches ONLY the two stat call-site hunks (no string literal, no
classification logic, no import-block change, no formatting churn elsewhere in the file).

### AC-SEAM-007 — Scoped verification only, 1800s timeout (REQ: REQ-007)

**Given** the local verification record for this SPEC, **When** the evidence in `progress.md`
§E.2 and `.moai/reports/t563/` is inspected, **Then** every `go test` invocation is a scoped
`./internal/cli/` run with `-run '<relevant>' -count=1 -timeout 1800s`, plus `go vet` and
`golangci-lint` on the package — and NO local full-suite (`go test ./...`) invocation appears
anywhere in the evidence. CI owns the full verdict.

### AC-SEAM-008 — Seam overrides stay serial (REQ: REQ-007)

**Given** every new or modified test that reassigns `osStatFn`, **When** the test files are read,
**Then** none declares `t.Parallel()`, and each follows the save → replace → `t.Cleanup` restore
pattern of `codex_skills_prune_test.go:55-57` (package-level var; concurrent reassignment is a
data race per the `@MX:WARN` at `update_preserve_inventory.go:53`).

## §D.1 Edge Cases Covered

- Directory at a classified skill path resolves (not missing) — site B by-design behavior, pinned.
- Relative and oddly-formed entries never reach the stat (zero recorded calls) — proves the skip
  survives the swap.
- Unresolvable home → indeterminate, never missing (fail-open preserved).
- Non-NotExist stat errors → indeterminate (never folded into absent), both sites.

## §D.2 Quality Gates

- `go vet ./internal/cli/...` — zero findings.
- `golangci-lint run internal/cli/...` — zero new findings vs the M1 baseline.
- Scoped tests green with `-timeout 1800s` (AC-SEAM-007 command shape).
- TRUST 5: Tested (characterization + observation suites), Readable/Unified (repo conventions,
  C5), Secured (no behavior change; no new input surface), Trackable (`t563` in both commit
  messages).

## §D.3 Definition of Done

1. AC-SEAM-001 through AC-SEAM-008 all PASS with verbatim recorded evidence.
2. Both commits landed on `WT-doctor-stat-shim`, messages carrying `t563`, ordering per
   AC-SEAM-002.
3. No out-of-scope file modified beyond `internal/cli/doctor_codex.go` (2 lines) + new test files.
