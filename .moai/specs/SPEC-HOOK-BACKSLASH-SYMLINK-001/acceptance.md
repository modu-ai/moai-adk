# SPEC-HOOK-BACKSLASH-SYMLINK-001 — Acceptance Criteria

> Stateless artifact. Every AC is machine-verifiable. The reproduction tests were
> AUTHORED AND EXECUTED at plan phase (audit iteration-1 D1): the test file
> `internal/hook/pre_tool_backslash_repro_test.go` lands with this SPEC revision
> commit, BEFORE any fix exists — the RED observations below were taken on that
> pre-fix tree and are recorded verbatim here per the four-element rule
> (verification-completeness.md §2.1). The plan-phase evidence carrier is this
> document (a table cell / evidence-ledger entry), not progress.md.

## D. AC Matrix

### AC-HBS-001 — Reproduction-first: backslash-named symlink escape is denied (release-blocking)

**Given** a POSIX host able to create directory symlinks, a temporary project root,
and a directory literally named `innocent\dir` inside it, symlinked to a directory
outside the project root,
**When** `checkFileAccess` is invoked with a Write `tool_input` of
`file_path: <project>/innocent\dir/escaped.txt`,
**Then** the decision is `deny`, and `<outside>/escaped.txt` does not exist
(the fixture's non-deny branch performs the actual Write through the literal path
before failing, so a vulnerable tree demonstrably writes outside — AC-HBS-002).

Two-cell adoption (verification-completeness.md §2):

- **RED-now cell** — OBSERVED at plan phase. Evidence ledger `RED-HBS-001`:
  - **Command**: `go test -count=1 ./internal/hook/ -run 'TestCheckFileAccessPosixBackslashSymlinkEscape'`
  - **Observed stdout (verbatim, raw bytes)**:
    ```text
    --- FAIL: TestCheckFileAccessPosixBackslashSymlinkEscape (0.01s)
        pre_tool_backslash_repro_test.go:59: guard allowed the backslash-symlink escape: decision="" reason="" (external write landed at /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCheckFileAccessPosixBackslashSymlinkEscape2624617744/002/escaped.txt)
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/hook	0.880s
    FAIL
    ```
  - **Exit code**: `1` (observed, own field — `echo EXIT_CODE=$?` on the unredirected-target run printed `EXIT_CODE=1`)
  - **Tree SHA**: `785cfaaff` (branch `WT-backslash-symlink`, working tree carrying only this SPEC's artifacts + the two repro test files; code identical to the pinned `cad44a751` — the auditor's own read confirmed only SPEC docs differ).
  - **RED reason (stated)**: the pre-fix walk converts `innocent\dir` into
    `innocent/dir`, validates the fictional in-project rejoined path, and returns
    decision `""` (allow); the fixture's simulated tool write then landed at the
    EXTERNAL destination recorded verbatim above.
- **Green path cell** — M2 (platform-appropriate segmentation) flips this criterion;
  the passing output is exit 0 with the deny observed and the external file absent.

### AC-HBS-002 — No byte reaches the external destination (release-blocking)

**Given** the AC-HBS-001 fixture, whose execution model is "perform the actual Write
through the literal path whenever the guard decision is NOT `deny`" (implemented in
`TestCheckFileAccessPosixBackslashSymlinkEscape`'s non-deny branch),
**When** the guard runs against the escape path,
**Then** pre-fix the external write IS observed (matching the reproduced t1533
finding — `"escaped"` lands outside); post-fix the decision is `deny`, no write is
performed, and a file-existence check on `<outside>/escaped.txt` reports the file
does not exist.

Two-cell adoption:

- **RED-now cell** — OBSERVED at plan phase, same run as `RED-HBS-001`: the verbatim
  failure line above IS the external-write observation (`external write landed at
  /var/.../002/escaped.txt` — the `002` directory is the OUTSIDE temp root, distinct
  from the project root `001`). Command and exit code as in `RED-HBS-001`.
- **Green path cell** — M2: the deny branch runs, the non-deny write branch is
  skipped, and the `os.Stat` assert inside the test verifies external absence.

### AC-HBS-003 — Outside resolution through the backslash component is visible to the boundary check (release-blocking)

**Given** the AC-HBS-001 fixture shape (project-local directory `innocent\dir`
symlinked to an outside directory),
**When** `resolveThroughExistingParent` resolves
`<project>/innocent\dir/escaped.txt` on POSIX,
**Then** the resolved path is OUTSIDE the project (the literal component followed as
the symlink it is) — never a fictional in-project rejoined spelling such as
`<project>/innocent/dir/escaped.txt`.

Two-cell adoption:

- **RED-now cell** — OBSERVED at plan phase. Evidence ledger `RED-HBS-003`:
  - **Command**: `go test -count=1 ./internal/hook/ -run 'TestResolveThroughExistingParentPosixBackslashSymlinkDivergence'`
  - **Observed stdout (verbatim excerpt)**:
    ```text
    --- FAIL: TestResolveThroughExistingParentPosixBackslashSymlinkDivergence (0.00s)
        pre_tool_backslash_repro_test.go:106: walk validated a fictional IN-PROJECT path "/private/var/.../001/innocent/dir/escaped.txt" while the literal component `innocent\dir` is a symlink to "/var/.../002" — validated path diverges from the path the OS walks
    ```
  - **Exit code**: `1` (observed; joint run with AC-HBS-001 printed `EXIT_CODE=1`)
  - **Tree SHA**: `785cfaaff` (code identical to `cad44a751`).
  - **RED reason (stated)**: the pre-fix split invents the `innocent` component, the
    Lstat probe fails, and the walk rejoins the in-project tail with `ok=true`.
- **Green path cell** — M2: the same test passes because the walk follows the literal
  symlinked component and the resolution lands outside.

### AC-HBS-004 — No false positive on legitimate POSIX backslash names (regression-guard)

**Given** a POSIX host and a project containing a NEW file path whose component name
legitimately contains `\` (e.g. `<project>/weird\name.txt`, not yet existing),
**When** `checkFileAccess` is invoked with a Write `tool_input` for it,
**Then** the decision is NOT a boundary deny (the write is allowed or handled by other
checks exactly as an equivalent plain name would be).

- **RED-now**: expected GREEN already on the pre-fix tree — this criterion is the
  regression guard against a character-blacklist "fix"; a violation must surface
  during M3, never after.
- **Green path**: M3.

### AC-HBS-005 — Windows separator semantics preserved (regression-guard)

**Given** the segmentation logic with a Windows-platform selection,
**When** unit-tested at the string level (no Windows filesystem required) with
`C:\proj\linked\..\x`-shaped input,
**Then** the split treats `\` as a separator (segments `C:`, `proj`, `linked`, `..`, `x`)
and the t1530 physical-walk guarantees hold unchanged.

- **RED-now**: expected GREEN already (behavior preserved); this criterion pins the
  preservation — it guards the M2 mechanism against breaking Windows.
- **Green path**: M3. Compile surface verified by `GOOS=windows go build ./...` (exit 0).

### AC-HBS-006 — Existing t1530 guarantee tests stay green (regression-guard)

**Given** the existing internal/hook tests covering the t1530 repair's new-file tail
rejoin and `..`-popped-against-resolved-prefixes behavior,
**When** `go test -timeout 30m -count=1 ./internal/hook/` runs after the fix,
**Then** the ONLY failing tests are the three recorded pre-existing baseline
failures (`TestStaleRunNoticeLegacyLeaderSpelling`, `TestStaleRunNoticeLegacySessionRecord`,
`TestStaleRunNoticeFactoryLegacyLabel` — observed failing in isolation on this tree
before any fix, unrelated to this diff: this SPEC touches only the path-interpretation
layer, and the three tests fail identically with the SPEC's changes absent);
the two plan-phase repro tests must be GREEN, and NO test beyond the recorded trio
may fail. A bare exit-0 claim is not measurable on this tree today — the baseline
trio is recorded here so the post-fix verdict is a countable delta, not an
unreachable absolute.

- **Coverage claim scope (corrected per audit iteration-1 D3)**: this criterion
  claims ONLY that the existing tests remain green — it does NOT claim coverage of
  the symlink depth-bound branch (`pre_tool.go:1444`), which the audit measured as
  executing 0 times across the existing t1530 tests. Depth-bound preservation is
  asserted by AC-HBS-007's explicit fixtures, not by this pass.
- **RED-now**: GREEN pre-fix except the two plan-phase repro tests (observed,
  `RED-HBS-001`/`RED-HBS-003`); post-fix must be fully GREEN.
- **Green path**: M4 (full package re-measurement).

### AC-HBS-007 — Fail-closed semantics preserved, with explicit fixtures (regression-guard)

**Given** two fail-closed fixtures, authored in M3 alongside the M2 mechanism:

1. **Depth-bound-exceeded**: a chain of directory symlinks longer than
   `zoneSymlinkDepthBound` (each link pointing at the next, terminal target an
   existing directory), resolved through a path that traverses the whole chain;
2. **Unresolvable `..`**: a path of the form `<project>/<missing-prefix>/../leaf`
   where `<missing-prefix>` does not exist, so `..` pops over a prefix the walk
   never resolved;

**When** `resolveThroughExistingParent` resolves each fixture path,
**Then** for fixture 1 it returns `ok=false` (chain past the depth bound), and for
fixture 2 it returns `ok=false` (`..` over an unresolved prefix) — the same
fail-closed second returns the pre-fix walk produces; a companion decision-level
assert applies the existing `checkFileAccess` fallback semantics unchanged.

- **Verification command form**: `go test -count=1 ./internal/hook/ -run 'TestResolvePhysicalWalkFailClosed'` (test authored in M3; command recorded here as its specified form).
- **RED-now**: expected GREEN pre-fix (preservation criterion — these fixtures
  characterize current behavior); guards the repair from opening a fail-open path.
- **Green path**: M3.

## D.1 Severity

- **Release-blocking** (carrying observed four-element RED-now cells): AC-HBS-001,
  AC-HBS-002, AC-HBS-003.
- **Regression-guard** (expected-green-at-arrival preservation criteria; no RED-now
  expectation by design — their RED would be observed only if the repair regresses
  them): AC-HBS-004, AC-HBS-005, AC-HBS-006, AC-HBS-007.

## D.2 Traceability

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-HBS-001 | REQ-HBS-001, REQ-HBS-002, REQ-HBS-003 | M1 (RED observed at plan phase) → M2 (GREEN) |
| AC-HBS-002 | REQ-HBS-003 | M2 |
| AC-HBS-003 | REQ-HBS-002 | M2 |
| AC-HBS-004 | REQ-HBS-005 | M3 |
| AC-HBS-005 | REQ-HBS-004 | M3 |
| AC-HBS-006 | REQ-HBS-004, REQ-HBS-006 | M4 |
| AC-HBS-007 | REQ-HBS-006 | M3 (fixtures) |

## D.3 Quality gate criteria

- `go test -timeout 30m -count=1 ./internal/hook/` fails on the recorded pre-existing
  baseline trio ONLY (AC-HBS-006; no NEW failure post-fix).
- `GOOS=windows go build ./...` exit 0.
- `go vet ./internal/hook/` exit 0; `golangci-lint run` clean on touched packages.
- TRUST 5 Secured: the deny path is exercised by an observed test (not by reading).

## D.4 Definition of Done

1. All release-blocking ACs GREEN with verbatim outputs recorded in `progress.md` §E.2.
2. The RED cells of AC-HBS-001 and AC-HBS-003 recorded in this document (pre-fix tree
   pinned `785cfaaff`, code identical to `cad44a751`) — a green-only adoption is
   incomplete.
3. Scope check: `git diff --stat` against the base shows changes only under
   `internal/hook/` (+ this SPEC's artifacts + the repro test file) — no `zoneSlash` touch.
4. No new dependency, no public API change outside `internal/hook`.
