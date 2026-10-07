# SPEC-HOOK-ZONE-BACKSLASH-001 — Acceptance Criteria

> Stateless artifact. Every AC is machine-verifiable. The reproduction tests are
> AUTHORED at run-phase M1 (the plan-phase scope for this card excludes code files),
> and the RED observations below are recorded INTO this document's evidence ledger at
> M1 execution, before any fix exists — per the two-cell adoption rule
> (verification-completeness.md §2). Until a release-blocking criterion's four-element
> RED observation (command, verbatim stdout, exit code, tree SHA) is recorded, that
> criterion is UNADOPTED — not a pass, not a gate. The tree pin for every RED cell is
> `f97edcc55` (or a tree code-identical to it carrying only this SPEC's docs — state
> the equivalence when citing).

## D. AC Matrix

### AC-HZB-001 — Reproduction-first: a Write through a backslash-named link into the protected zone is denied (release-blocking)

**Given** a POSIX host able to create directory symlinks, a t.TempDir()-based project
root with a protected zone directory `zone_dir` (declared through the guard's
`zoneLoader` seam), and a project entry literally named `lnk\dir` symlinked to
`zone_dir`,
**When** the Write/Edit guard (`checkProtectedZone`) is invoked with
`file_path: <project>/lnk\dir/secret.md`,
**Then** the decision is `deny` (the sentinel reason path), and no file exists at
`<project>/zone_dir/secret.md` (the fixture's non-deny branch performs the actual
write through the literal path and reads it back — AC-HZB-002).

Two-cell adoption (verification-completeness.md §2):

- **RED-now cell** — recorded at M1 execution, evidence ledger `RED-HZB-001`:
  - **Command**: `go test -count=1 ./internal/hook/ -run 'TestCheckProtectedZonePosixBackslashLinkBypass'`
  - **Observed stdout (verbatim)**: *(populated at M1 — the expected pre-repair shape
    is the guard returning allow and the fixture reporting the in-zone landing;
    the raw bytes decide)*
  - **Exit code**: *(populated at M1 — expected `1` pre-repair)*
  - **Tree SHA**: `f97edcc55` (or stated code-identical delta)
  - **RED reason (stated)**: pre-repair, `zoneSlash` rewrites `lnk\dir` to
    `lnk/dir` before resolution; both arms classify the fictional spelling outside
    `zone_dir/`; zero forms match; the guard allows — while the actual write lands
    inside the protected zone.
- **Green path cell** — M2 (the platform-appropriate slash conversion) flips this
  criterion; the passing output is exit 0 with deny observed and
  `<project>/zone_dir/secret.md` absent.

### AC-HZB-002 — The byte demonstrably lands inside the protected zone on the pre-repair tree (release-blocking)

**Given** the AC-HZB-001 fixture, whose execution model is "perform the actual Write
through the literal path whenever the guard decision is NOT deny, then read the file
back from inside `zone_dir`",
**When** the guard runs against the escape path on the pre-repair tree,
**Then** the in-zone write IS observed with verified content (matching the measured
round-8 shape: decision "allow" + the protected file changed); post-M2 the decision
is `deny`, the write branch is skipped, and the file is absent.

Two-cell adoption:

- **RED-now cell** — recorded at M1 execution, evidence ledger `RED-HZB-002` (same
  solo run as `RED-HZB-001` or its own): the verbatim failure line IS the in-zone
  landing observation — the read-back from `<project>/zone_dir/secret.md` proves the
  file the OS created through the literal `lnk\dir` symlink sits inside the zone.
  Command and exit code as in `RED-HZB-001`.
- **Green path cell** — M2: the deny branch runs, the non-deny write branch is
  skipped, and the file-existence assert inside the test verifies absence.

### AC-HZB-003 — Resolver-level: zone target resolution sees the real target through the backslash component (release-blocking)

**Given** the AC-HZB-001 fixture shape,
**When** `resolveZoneTarget(root, "<project>/lnk\dir/secret.md")` runs on POSIX
(and the relative-path variant through the `zoneGetwd` seam),
**Then** the returned forms include one whose folded value is inside the zone (the
literal component followed as the symlink it is) — never only the fictional
`lnk/dir/...` spellings.

Two-cell adoption:

- **RED-now cell** — recorded at M1 execution, evidence ledger `RED-HZB-003`:
  - **Command**: `go test -count=1 ./internal/hook/ -run 'TestResolveZoneTargetPosixBackslashLinkDivergence'`
  - **Observed stdout (verbatim)**: *(populated at M1)*
  - **Exit code**: *(populated at M1 — expected `1` pre-repair)*
  - **Tree SHA**: `f97edcc55` (or stated code-identical delta)
  - **RED reason (stated)**: pre-repair, both arms resolve the rewritten spelling;
    the resolver returns zero in-zone forms for a path that actually resolves inside
    the zone.
- **Green path cell** — M2: the same test passes because the resolver follows the
  literal symlinked component.

### AC-HZB-004 — Windows semantics and backslash-free paths unchanged (regression-guard)

**Given** the M2 repair,
**When** `go test -timeout 30m -count=1 ./internal/hook/` runs after the fix,
**Then** every existing zone test (`TestProtectedZone` subtest group and the rest of
the package) passes — the ONLY failing tests, if any, are exactly the set recorded as
the pre-change baseline at M1 (countable delta, no new failure); the M1 reproduction
tests are GREEN; and `GOOS=windows go build ./...` exits 0.

- **RED-now**: expected GREEN on the pre-repair tree except the M1 reproduction
  tests (observed at M1 alongside `RED-HZB-001`-`003`); this criterion is the
  regression guard against a character-blacklist "fix" and against breaking the
  Windows half.
- **Green path**: M3 (full package re-measurement + windows compile surface).

### AC-HZB-005 — Three-surface family sweep evidence recorded (regression-guard)

**Given** the card mandate's 3-surface family,
**When** M3 executes the sweep,
**Then** progress.md carries a table with one row per surface stating: the surface's
site, THIS tree's observed class-instance state, the owning card, and the evidence
(check or read), plus a repo-wide grep of the rewrite pattern with its swept count
and per-hit classification (in-family surface / out-of-family, e.g. the
`glmcred`/`jevcred` TOML-escaping hits are classified and dismissed). Surfaces 2 and
3 carry read-only evidence naming cards t1556 and t1561 as repair owners; this SPEC
adds no repair there.

- **RED-now**: not applicable (evidence-recording criterion); the sweep is judged by
  the swept count being stated and every named surface having a row — an empty sweep
  asserts nothing (verification-completeness.md §1.1).
- **Green path**: M3.

## Evidence Ledger

Entries `RED-HZB-001`, `RED-HZB-002`, `RED-HZB-003` are appended here at M1
execution with the four elements each (command, verbatim stdout as raw bytes, exit
code as its own field, tree SHA). The round-8 gate record
(`.moai/reports/t1556/codex-review-gate-1.md` § 라운드 8, delta commit `9d78421a4`,
in the t1556 card worktree) is the motivating measurement — cited for motivation,
never as a substitute for the M1 observation on this tree.
