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
  - **Observed stdout (verbatim)**: *(populated at M1 — measured; raw bytes in
    ledger entry `RED-HZB-001` below)*
  - **Exit code**: `1` (measured at M1)
  - **Tree SHA**: `da2d74eef` — code-identical to the pinned `f97edcc55`
    (verified: `git show --stat da2d74eef` = the four SPEC artifact files only,
    parent `f97edcc55`; zero Go-code delta)
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
  - **Observed stdout (verbatim)**: *(populated at M1 — measured; raw bytes in
    ledger entry `RED-HZB-003` below)*
  - **Exit code**: `1` (measured at M1)
  - **Tree SHA**: `da2d74eef` — code-identical to the pinned `f97edcc55`
    (SPEC-docs-only delta, verified as above)
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

### RED-HZB-001 — guard-level backslash-link bypass (appended at M1, 2026-10-07)

- **Command**: `go test -count=1 ./internal/hook/ -run 'TestCheckProtectedZonePosixBackslashLinkBypass'`
- **Observed stdout (verbatim)**:

```
--- FAIL: TestCheckProtectedZonePosixBackslashLinkBypass (0.71s)
    protected_zone_backslash_repro_test.go:87: absolute raw path: BYPASS — decision="allow" reason="", want deny; the write through the literal backslash link landed INSIDE the protected zone (zone_dir/secret.md="bypass")
    protected_zone_backslash_repro_test.go:87: relative raw path: BYPASS — decision="allow" reason="", want deny; the write through the literal backslash link landed INSIDE the protected zone (zone_dir/secret.md="bypass")
    protected_zone_backslash_repro_test.go:128: swept=3
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.584s
FAIL
```

- **Exit code**: 1
- **Tree SHA**: `da2d74eef` (code-identical to pinned `f97edcc55` — SPEC-docs-only
  delta, verified `git show --stat da2d74eef`)
- **RED reason (observed, matches the stated mechanism)**: both shapes returned
  `decision="allow"` and the demonstration write through the literal
  `<project>/lnk\dir/secret.md` landed inside the protected zone
  (`zone_dir/secret.md` contains `bypass`) — the round-8 shape (allow + protected
  file changed) reproduced on this tree. The positive-control row (ordinary
  `lnk\dir` directory outside the zone) PASSED in the same run — the RED is the
  bypass mechanism, not a fixture error.

### RED-HZB-002 — the byte demonstrably lands inside the protected zone (appended at M1, 2026-10-07)

- **Command / exit code / tree SHA**: as `RED-HZB-001` (same solo run).
- **Observed evidence (verbatim line)**:

```
    protected_zone_backslash_repro_test.go:87: absolute raw path: BYPASS — decision="allow" reason="", want deny; the write through the literal backslash link landed INSIDE the protected zone (zone_dir/secret.md="bypass")
```

- **RED reason (observed)**: the failure line itself is the in-zone landing
  observation — the fixture wrote through the literal backslash path after the
  allow decision and read the bytes back from `<project>/zone_dir/secret.md`,
  proving the OS followed the `lnk\dir` symlink into the zone while the guard
  allowed.

### RED-HZB-003 — resolver-level divergence (appended at M1, 2026-10-07)

- **Command**: `go test -count=1 ./internal/hook/ -run 'TestResolveZoneTargetPosixBackslashLinkDivergence'`
- **Observed stdout (verbatim)**:

```
--- FAIL: TestResolveZoneTargetPosixBackslashLinkDivergence (0.07s)
    protected_zone_backslash_repro_test.go:175: absolute raw path: no returned form resolves inside the protected zone for a path that goes through the backslash-named symlink (want a folded form under zone_dir/): [{Display:lnk/dir/secret.md Folded:lnk/dir/secret.md}]
    protected_zone_backslash_repro_test.go:175: relative raw path: no returned form resolves inside the protected zone for a path that goes through the backslash-named symlink (want a folded form under zone_dir/): [{Display:lnk/dir/secret.md Folded:lnk/dir/secret.md}]
    protected_zone_backslash_repro_test.go:198: swept=3
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	0.829s
FAIL
```

- **Exit code**: 1
- **Tree SHA**: `da2d74eef` (code-identical to pinned `f97edcc55`)
- **RED reason (observed, pins the truth over the §B narrative)**: the resolver
  returned ONLY the fictional `lnk/dir/secret.md` form — zero in-zone forms for a
  path that physically resolves inside the zone. Measured note: the §B hypothesis
  guessed the unresolved-tail rejoin would mint `dir/secret.md`; the observed
  rejoin starts at the FIRST missing component (`lnk`), yielding the single
  fictional form `lnk/dir/secret.md`. The core mechanism (the rewritten spelling
  misses the real link; both arms classify outside the zone; allow) is confirmed
  exactly; the narrative's intermediate spelling was a hypothesis and the RED
  pins the measured one.

### RED-HZB-X1 — reinforcement: gate-turnend-1 independent reproduction (appended at M1, 2026-10-07)

- **Evidence**: `.moai/reports/t1566/gate-turnend-1.md` (turn-end codex review
  gate verdict 1, measured on the same tree base `f97edcc55`/"f97ec555", no lane
  work committed). The gate independently reproduced this card's defect class via
  a temporary Go overlay: "The same rewrite in `protected_zone_path.go:36`
  allowed Write/Edit/Bash through a backslash alias into the protected zone."
  Stronger than the round-8 cross-tree record because it is same-tree. Gate
  overlay reproduction; no repository files changed by the gate.

### RED-HZB-X2 — reinforcement: gate-turnend-2 richer statement (appended at M1, 2026-10-07)

- **Evidence**: `.moai/reports/t1566/gate-turnend-2.md` (turn-end codex review
  gate verdict 2, measured at the same base). Richer statement of this card's
  defect: backslash alias `alias\dir → zone_dir` symlink — "Write and Bash both
  allowed into the protected zone; the backslash→slash rewrite before filesystem
  lookup misses the real link; 'separate comparison normalization from real-path
  resolution'" — pinning the resolution arm this SPEC's M2 repairs. Cited as
  independent corroboration of the M1 RED; the binding RED observations remain
  `RED-HZB-001`/`RED-HZB-002`/`RED-HZB-003` above.

### RED-HZB-004 — converted-absoluteness vector (appended at run phase, 2026-10-07; leader-forwarded gate round 8)

The turn-end gate round 8 (relayed by the leader mid-run; the lane's
gate-turnend files record rounds 1-6, so this relayed note is carried here as
its record surface) measured ONE MORE vector on this card's named surface: the
cwd-prepend decision in `resolveZoneTarget` reads the slash-CONVERTED spelling,
so a POSIX relative raw `\alias/secret.md` converts to a "/"-leading form and
`C:\alias/secret.md` to a drive-letter form — both wrongly judged absolute, the
prepend is skipped, and no arm follows the literal component into the zone
(gate-measured: forms=[], decision allow, the write reached the protected
file). Same named instance, same root cause (rewrite-before-decision), same
file — repaired in this card's M2 amendment, not a new card.

Two-cell adoption (measured here after the relay, per baseline attribution):

- **RED-now cell** — test
  `TestCheckProtectedZonePosixBackslashConvertedAbsoluteness`
  (`internal/hook/protected_zone_backslash_repro_test.go`), run against the
  PRE-AMENDMENT resolver (Go code exactly `7945a442a`, checked out via
  `git show 7945a442a:` into the working tree; the working tree differed from
  `7945a442a` only by this uncommitted test file):
  - **Command**: `go test -count=1 ./internal/hook/ -run 'TestCheckProtectedZonePosixBackslashConvertedAbsoluteness'`
  - **Observed stdout (verbatim)**:

```
--- FAIL: TestCheckProtectedZonePosixBackslashConvertedAbsoluteness (0.83s)
    protected_zone_backslash_repro_test.go:185: leading backslash relative: BYPASS — decision="allow" reason="", want deny; the write through the literal component landed INSIDE the protected zone (zone_dir/secret.md="bypass")
    protected_zone_backslash_repro_test.go:185: drive-letter prefixed relative: BYPASS — decision="allow" reason="", want deny; the write through the literal component landed INSIDE the protected zone (zone_dir/secret.md="bypass")
    protected_zone_backslash_repro_test.go:209: resolver: no returned form resolves inside the protected zone for the leading-backslash relative raw (want a folded form under zone_dir/): []
    protected_zone_backslash_repro_test.go:228: swept=4
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.681s
FAIL
```

  - **Exit code**: 1
  - **Tree SHA**: Go code under test = `7945a442a` (the M2 commit, pre-amendment
    resolver; stated equivalence per the two-cell rule — the working tree
    carried only the uncommitted vector test).
  - **RED reason (observed)**: both shapes returned `decision="allow"` and the
    demonstration write through the SAME literal component the raw names
    (`\alias`, `C:\alias` — each its own symlink to `zone_dir`) landed inside
    the protected zone; the resolver returned zero forms (the gate's measured
    `forms=[]` shape). The fourth row (backslash-free `c:/x.zonefile` → no
    forms) PASSED in the same run, pinning that the amendment must not change
    backslash-free drive-letter-form behavior.
- **Green path cell**: the M2 amendment (the walk receives its own
  platform-correct cwd prepend when the conversion changed the spelling) flips
  it — measured: exit 0, both shapes deny via `wantZoneDeny`, resolver yields a
  `zone_dir/`-prefixed form, protected file absent, and the `c:/x.zonefile`
  regression row still holds.
