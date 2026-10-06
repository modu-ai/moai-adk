# SPEC-HOOK-BACKSLASH-SYMLINK-001 — Acceptance Criteria

> Stateless artifact. Every AC is machine-verifiable; the command forms are the plain,
> single-invocation forms runnable in the card worktree. Symlink-fixture ACs carry a
> platform skip guard; the verdict on a skipped runner is recorded, never silently
> counted as a pass.

## D. AC Matrix

### AC-HBS-001 — Reproduction-first: backslash-named symlink escape is denied (release-blocking)

**Given** a POSIX host able to create directory symlinks, a temporary project root,
and a directory literally named `innocent\dir` inside it, symlinked to a directory
outside the project root,
**When** `checkFileAccess` is invoked with a Write `tool_input` of
`file_path: <project>/innocent\dir/escaped.txt`,
**Then** the decision is `deny`, and `<outside>/escaped.txt` does not exist.

Two-cell adoption (verification-completeness.md §2):

- **RED-now cell** — pinned tree `cad44a751` (branch `WT-backslash-symlink`, clean).
  Command: `go test -count=1 ./internal/hook/ -run 'TestCheckFileAccessPosixBackslashSymlinkEscape'`
  Expected pre-fix: exit 1 — the test observes an ALLOW and (in a fixture variant that
  exercises the real write path or asserts the resolution divergence) the escape.
  To be executed and its verbatim output recorded in `progress.md` §E.2 by the repair
  agent BEFORE the fix lands. RED reason (stated): the pre-fix walk converts
  `innocent\dir` to `innocent/dir`, validates the fictional in-project rejoined path,
  and returns no deny.
- **Green path cell** — M2 (platform-appropriate segmentation) flips this criterion;
  the passing output is exit 0 with the deny observed and the external file absent.

### AC-HBS-002 — No byte reaches the external destination (release-blocking)

**Given** the AC-HBS-001 fixture,
**When** the post-fix guard runs against the escape path,
**Then** a file-existence check (`os.Stat` in the test) on
`<outside>/escaped.txt` reports the file does not exist.

- **RED-now**: implied by AC-HBS-001's fixture variant (the reproduced finding wrote
  `"escaped"` externally pre-fix).
- **Green path**: M2.

### AC-HBS-003 — Outside resolution through the backslash component is visible to the boundary check

**Given** the AC-HBS-001 fixture,
**When** the walk resolves `<project>/innocent\dir/escaped.txt` on POSIX,
**Then** the resolved path it validates is the outside destination (the literal
component followed as the symlink it is), not any in-project rejoined spelling —
observable by asserting the deny reason / resolution in the unit against
`resolveThroughExistingParent` directly.

- **RED-now**: pre-fix `resolveThroughExistingParent` returns an in-project path
  (`ok=true`, inside the project) for the escape input.
- **Green path**: M2.

### AC-HBS-004 — No false positive on legitimate POSIX backslash names (release-blocking)

**Given** a POSIX host and a project containing a NEW file path whose component name
legitimately contains `\` (e.g. `<project>/weird\name.txt`, not yet existing),
**When** `checkFileAccess` is invoked with a Write `tool_input` for it,
**Then** the decision is NOT a boundary deny (the write is allowed or handled by other
checks exactly as an equivalent plain name would be).

- **RED-now**: expected GREEN already on the pre-fix tree for the plain-allow branch —
  this criterion is the regression guard against a character-blacklist "fix"; if the
  chosen M2 mechanism breaks it, it must fail during M3, never after.
- **Green path**: M3.

### AC-HBS-005 — Windows separator semantics preserved (release-blocking)

**Given** the segmentation logic with a Windows-platform selection,
**When** unit-tested at the string level (no Windows filesystem required) with
`C:\proj\linked\..\x`-shaped input,
**Then** the split treats `\` as a separator (segments `C:`, `proj`, `linked`, `..`, `x`)
and the t1530 physical-walk guarantees hold unchanged.

- **RED-now**: expected GREEN already (behavior preserved); this criterion pins the
  preservation — it guards the M2 mechanism against breaking Windows.
- **Green path**: M3. Compile surface verified by `GOOS=windows go build ./...` (exit 0).

### AC-HBS-006 — t1530 guarantee regression net stays green (release-blocking)

**Given** the existing internal/hook tests covering the t1530 repair (new-file tail
rejoin, `..` popped against resolved prefixes, symlink depth-bound fail-closed),
**When** `go test -timeout 30m -count=1 ./internal/hook/` runs after the fix,
**Then** the package passes with exit 0.

- **RED-now**: GREEN on the pre-fix tree (baseline); post-fix must remain GREEN.
- **Green path**: M4 (full package re-measurement).

### AC-HBS-007 — Fail-closed semantics not weakened (release-blocking)

**Given** a path whose walk cannot be vouched for (symlink chain past the depth bound,
or `..` above a resolved prefix),
**When** the post-fix guard resolves it on POSIX,
**Then** the same fail-closed/fallback behavior as the pre-fix walk is observed
(unit assert on `resolveThroughExistingParent` second return and on the resulting
decision).

- **RED-now**: GREEN pre-fix (preservation criterion; guards the repair from opening a
  fail-open path).
- **Green path**: M3.

## D.1 Severity

- Release-blocking: AC-HBS-001, AC-HBS-002, AC-HBS-003, AC-HBS-004, AC-HBS-005,
  AC-HBS-006, AC-HBS-007.
- Regression-guard (preservation, expected-green-at-arrival): AC-HBS-004, AC-HBS-005,
  AC-HBS-006, AC-HBS-007 — these carry no RED-now expectation by design; their RED
  would be observed only if the repair regresses them.

## D.2 Traceability

| AC | REQ | Milestone |
|----|-----|-----------|
| AC-HBS-001 | REQ-HBS-001, REQ-HBS-002, REQ-HBS-003 | M1 (RED) → M2 (GREEN) |
| AC-HBS-002 | REQ-HBS-003 | M2 |
| AC-HBS-003 | REQ-HBS-002 | M2 |
| AC-HBS-004 | REQ-HBS-005 | M3 |
| AC-HBS-005 | REQ-HBS-004 | M3 |
| AC-HBS-006 | REQ-HBS-004, REQ-HBS-006 | M4 |
| AC-HBS-007 | REQ-HBS-006 | M3 |

## D.3 Quality gate criteria

- `go test -timeout 30m -count=1 ./internal/hook/` exit 0 (post-fix).
- `GOOS=windows go build ./...` exit 0.
- `go vet ./internal/hook/` exit 0; `golangci-lint run` clean on touched packages.
- TRUST 5 Secured: the deny path is exercised by an observed test (not by reading).

## D.4 Definition of Done

1. All release-blocking ACs GREEN with verbatim outputs recorded in `progress.md` §E.2.
2. The RED cell of AC-HBS-001 recorded (pre-fix tree pinned `cad44a751`) — a green-only
   adoption is incomplete.
3. Scope check: `git diff --stat` against the base shows changes only under
   `internal/hook/` (+ this SPEC's artifacts) — no `zoneSlash` touch.
4. No new dependency, no public API change outside `internal/hook`.
