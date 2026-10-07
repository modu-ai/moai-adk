# SPEC-HOOK-ZONE-BACKSLASH-001 — Implementation Plan

> Stateless artifact. Lifecycle status lives in `spec.md` frontmatter.

## A. Context

- Card t1566 (security P1, measured/demonstrated). Defect source: round 8 of the
  card-t1556 codex review gate (2026-10-07, delta commit `9d78421a4`) — primary copy
  `.moai/reports/t1556/codex-review-gate-1.md` § 라운드 8 in the t1556 card worktree
  (a disposable tree; the durable landing is the t1556 PR). Measured shape: an actual
  harness-learner Write call through a backslash-named link into the protected zone
  received `decision="allow"` and the protected file recorded `"changed"`.
- Defect site read and confirmed in this tree (`f97edcc55`):
  `internal/hook/protected_zone_path.go:35-37` — `zoneSlash` rewrites every `\` to
  `/` unconditionally on both platforms; callers `:58`/`:59` (`zoneLexicalRel`),
  `:183`/`:190` (raw and cwd-joined input in `resolveZoneTarget`),
  `:208`/`:210` (root resolution in the symlink arm) — all inside the zone target
  resolution path.
- Attack mechanics (the spec.md §B hypothesis, to be pinned by M1's RED): a project
  entry literally named `lnk\dir` symlinked to the protected `zone_dir`. The rewrite
  validates the fictional `<project>/lnk/dir/secret.md`; the walk's probe of
  `<project>/lnk` fails Lstat as genuinely missing, the unresolved-tail branch
  (`:147`-`:150`) rejoins `dir/secret.md` onto the resolved prefix, and both arms
  classify the target outside `zone_dir/` — zero forms, no match, allow. The real
  Write follows the literal symlink and lands inside the protected zone.
- Consumers sharing the resolver: `checkProtectedZone` (Write/Edit,
  `protected_zone_guard.go:157`) and `checkProtectedZoneShell` (Bash,
  `protected_zone_shell.go:1014`). One repair covers both (REQ-HZB-005).

## B. Known Issues

- The defect is an input-normalization error inside the t1510 walking pattern, not a
  failure of the pattern: `zoneResolve`'s probe-existing-then-rejoin structure is
  correct once it is fed correctly spelled segments. The repair narrows to
  platform-appropriate slash conversion at the `zoneSlash` boundary; the walk body
  (probe / rejoin / `..` pop / depth-bound fail-closed) is retained as-is unless M1's
  RED observation contradicts the hypothesis — the RED test pins the truth.
- The matching layer stays slash-separated: `zoneLexicalRel`'s returned relative forms
  feed `config.FoldZoneText` and manifest `entry.Match`, which compare
  `/`-separated spellings. A POSIX-correct repair keeps backslash-bearing components
  as SINGLE components through resolution and still yields slash-separated relative
  forms for matching (a literal `lnk\dir` component under the root yields the
  relative form `lnk\dir/secret.md` — a manifest entry listing the protected
  directory matches by the directory prefix, so the component's internal spelling
  rides through untouched).
- The round-8 measurement ran against the t1556 worktree at delta commit
  `9d78421a4`; this tree's `f97edcc55` carries the identical
  `protected_zone_path.go` (verified by read: `:35`-`:37` byte-shape matches the
  finding). The RED is re-measured HERE per the baseline-attribution rule — the
  round-8 record is motivation evidence, never a substitutable measurement.

## C. Pre-flight

1. `git rev-parse --short HEAD` in the card worktree — confirm the tree the RED
   measurement is pinned to (`f97edcc55`, or a code-identical docs-only delta).
2. Locate the existing fixture conventions before authoring: the guard-level tests
   build a `preToolHandler` and inject the manifest via the `zoneLoader` seam
   (`protected_zone_guard.go:138`-`:143`; usage e.g.
   `protected_zone_guard_test.go:456`); the cwd for relative paths is injectable via
   the `zoneGetwd` var (`protected_zone_path.go:25`, "Tests replace it"). The
   `TestProtectedZone` subtest-group convention (sweep-count per subtest) is the
   house style for pure-surface coverage.
3. Record the affected-package baseline verdict at M1 (which tests, if any, fail on
   the pre-change tree) so AC-HZB-004's post-fix claim is a countable delta, not an
   unreachable absolute.
4. The t1530/t1510-era zone tests (new-file tail rejoin, `..` physical pop, depth
   bound) are the regression net for REQ-HZB-004.

## D. Constraints

- Reproduction-first (Rule 4): the failing test lands BEFORE the fix, and its RED on
  the pre-repair tree is observed and recorded (two-cell adoption,
  `verification-completeness.md` §2). The plan-phase artifacts define the cells; M1
  execution populates them.
- No new dependencies; no API changes outside `internal/hook`.
- Windows semantics untouched (REQ-HZB-002): on Windows the `\`→`/` conversion stays
  correct. `GOOS=windows go build ./...` exit 0 is the compile surface.
- Symlink-fixture tests skip on platforms that cannot create directory symlinks
  unprivileged (`t.Skip`); the POSIX gate is a runtime `GOOS` check so the Windows
  runner skips the fixture while the separator-semantics assertions stay
  platform-independent where pure.
- Scope discipline (spec.md §F): no repair in `pre_tool.go` or
  `landing_predicate.go`; the M3 sweep reads them read-only.

## E. Self-Verification

- E1: RED cells of AC-HZB-001/002/003 observed on the pre-repair tree (verbatim
  `go test` output, exit code, pinned SHA) — recorded in progress.md §M1 and the
  acceptance.md evidence ledger.
- E2: GREEN cells observed after M2, same command form.
- E3: `GOOS=windows go build ./...` exit 0.
- E4: `go vet ./internal/hook/` and `golangci-lint run ./internal/hook/...` clean.
- E5: pre-fix the in-zone landing is observed (file read back from `zone_dir/`);
  post-fix the deny branch runs and the file is absent (file-existence check, not
  asserted).
- E6: the M3 sweep table carries a per-surface state + owner, and the grep sweep
  states its swept count (empty-sweep rule).
- E7: `git diff --stat` against the base shows ONLY `internal/hook/protected_zone_path.go`
  + the new test file + this SPEC's artifacts (no sibling-file repairs).

## F. Milestones

Ordered by decision-reversibility: the reproduction leads (proves the defect shape on
this tree); the repair follows; the family sweep and re-measurement trail.

- **M1 — RED reproduction (test authored, RED observed)** [Priority High]
  Author `internal/hook/protected_zone_backslash_repro_test.go` in
  `internal/hook`:
  - Guard-level `TestCheckProtectedZonePosixBackslashLinkBypass` (table-driven, two
    shapes: absolute raw path; relative raw path resolved via the `zoneGetwd` seam):
    t.TempDir()-based project root; protected zone directory `zone_dir` declared via
    the `zoneLoader` seam; a project entry literally named `lnk\dir` symlinked to
    `zone_dir`; invoke the guard with `file_path: <project>/lnk\dir/secret.md`.
    POSIX-gated (`runtime.GOOS != "windows"`, graceful `t.Skip` if symlink creation
    fails). The fixture's non-deny branch performs the ACTUAL write through the
    literal path and reads the file back from `zone_dir/secret.md` — so the RED
    demonstrates the in-zone landing, not just the decision string (the measured
    round-8 shape: decision "allow" + "changed" in the protected file).
  - Resolver-level `TestResolveZoneTargetPosixBackslashLinkDivergence`:
    `resolveZoneTarget(root, raw)` must yield a form whose folded value matches
    inside the zone for the raw backslash-link path (resolution sees the real
    target). Same fixture shape.
  - Run both on the pre-repair tree; observe RED (decision allow + in-zone write
    verified; resolver returns only outside forms). Record the verbatim output,
    exit code, and tree SHA into progress.md §M1 and the acceptance.md evidence
    ledger (`RED-HZB-001`/`RED-HZB-002`/`RED-HZB-003`). Also record the
    affected-package baseline (pre-change failing set, if any).
  - The test file is committed WITH the SPEC revision batch, before any fix exists —
    the ordering claim rests on the recorded observation, taken now.
- **M2 — Minimal repair of the named instance (GREEN)** [Priority High]
  Platform-appropriate slash conversion in the zone target resolution path of
  `internal/hook/protected_zone_path.go` ONLY (the `zoneSlash` boundary and its
  callers `:58`/`:59`/`:183`/`:190`/`:208`/`:210`): on POSIX, `\` stays an ordinary
  filename character through resolution; on Windows, the rewriting behavior is
  preserved. Keep the matching forms slash-separated (spec.md §D). The walk body and
  the fail-closed branches are unchanged (REQ-HZB-004). Flip M1's tests to GREEN;
  the passing output is exit 0 with deny observed and the in-zone file absent.
- **M3 — Three-surface family re-check sweep + package re-measurement** [Priority
  Medium]
  - Sweep the 3 named surfaces and record the state of each as evidence (progress.md
    §M3 table, with counts):
    1. `internal/hook/protected_zone_path.go` — REPAIRED by this card; verify by
       the M1 tests' GREEN flip.
    2. `internal/hook/pre_tool.go:1397` — read-only: record this tree's current
       class-instance state (pre-repair form present as of `f97edcc55`); owner
       card t1556 (fix in flight, PR open, disjoint file); NO repair here.
    3. `internal/cli/worktree/landing_predicate.go:161` — read-only: record this
       tree's current state; owner card t1561; NO repair here.
  - Back the table with a repo-wide grep for the unconditional `\`→`/` rewrite
    pattern feeding security-relevant resolution (e.g.
    `grep -rn 'ReplaceAll' internal/ --include='*.go' | grep -F '\\\\'`), enumerate
    hits, and classify each against the 3 surfaces + out-of-family (the
    `glmcred`/`jevcred` hits are TOML escaping, not path rewriting — classify and
    dismiss explicitly). State the swept count; an empty sweep asserts nothing.
  - Re-measure the affected package: `go test -timeout 30m -count=1
    ./internal/hook/` (verdict = no failure beyond the M1-recorded baseline set),
    `GOOS=windows go build ./...`, `go vet`, `golangci-lint run
    ./internal/hook/...`. Write the measured evidence into progress.md §E for the
    auditor.

## G. Anti-Patterns

- Do NOT repair `pre_tool.go` or `landing_predicate.go` in the same change — scope
  discipline (spec.md §F); their cards own the repairs.
- Do NOT deny every backslash-bearing POSIX path — that breaks legitimate names and
  REQ-HZB-003; the fix is resolution fidelity, not character blacklisting.
- Do NOT collapse backslash-bearing components before the filesystem walk
  (`filepath.Clean`/`Join` on the raw spelling) — that re-mints the fictional path
  the defect creates.
- Do NOT change the folded-matching semantics (`config.FoldZoneText`, `entry.Match`)
  to make the test pass — the matching layer stays slash-separated relative forms.
- Do NOT weaken the fail-closed branches (unresolved-tail rejoin for genuinely
  missing components, unreadable-entry return, depth bound) to flip the test.
- Do NOT cite the round-8 record as the RED measurement — it is motivation; the RED
  is re-measured on this tree at M1 (baseline-attribution rule).

## H. Cross-References

- Evidence: `.moai/reports/t1556/codex-review-gate-1.md` § 라운드 8 (card t1556
  review gate; primary copy lives in the t1556 card worktree — disposable; durable
  landing is the t1556 PR).
- Defect family: SPEC-HOOK-BACKSLASH-SYMLINK-001 (surface 1, card t1556 — its AC
  discipline is reused here), card t1561 (surface 2), the t1510 `zoneResolve`
  physical-walk repair, the t1530 physical-`..` walk.
- Zone founding SPEC: SPEC-SELF-IMPROVE-PROTECTED-ZONE-001 (the guard's identity
  gate, baseline floor, manifest state machine — all out of scope, unchanged).
- Doctrine: `.claude/rules/moai/development/verification-completeness.md` §2
  (two-cell adoption), §1.1 (empty-sweep), AGENTS.local.md §4 (scoped test
  execution).
