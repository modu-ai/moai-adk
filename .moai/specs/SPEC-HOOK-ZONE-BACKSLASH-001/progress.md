# SPEC-HOOK-ZONE-BACKSLASH-001 — Progress

## Plan-phase Seed (2026-10-07)

- **Card**: t1566 (security P1, measured/demonstrated). Factory lane card; plan →
  plan-audit → run → sync.
- **Base**: `f97edcc55` (branch `WT-protected-zone-backslash`, HEAD == origin/main
  tip at tree entry; working tree clean).
- **SPEC**: SPEC-HOOK-ZONE-BACKSLASH-001 — protected-zone target resolution preserves
  POSIX component identity (`zoneSlash`, `internal/hook/protected_zone_path.go:35`-`:37`).
- **Card mandate**: RED reproduction FIRST (M1), then minimal repair of the named
  instance ONLY (M2), then a 3-surface family re-check with read + test evidence
  (M3). Repairs of the sibling surfaces belong to their own cards — this card adds
  no repair to `pre_tool.go` (t1556) or `internal/cli/worktree/landing_predicate.go`
  (t1561).
- **Mechanism hypothesis** (spec.md §B; the M1 RED test pins the truth):
  `zoneSlash` rewrites `\`→`/` unconditionally before any filesystem step, so a
  backslash-named symlink component loses its identity; the walk resolves a
  fictional spelling, both arms classify outside `zone_dir/`, the guard allows —
  while the OS follows the literal symlink and lands the write inside the protected
  zone.
- **Motivating evidence (measured elsewhere, re-measured here at M1)**: round 8 of
  the card-t1556 codex review gate (2026-10-07, delta commit `9d78421a4`) — an
  actual harness-learner Write call through a backslash-named link received
  `decision="allow"` and the protected file recorded `"changed"`. Primary copy:
  `.moai/reports/t1556/codex-review-gate-1.md` § 라운드 8 (t1556 card worktree —
  disposable tree; durable landing is the t1556 PR).
- **Plan-phase timestamp**: 2026-10-07 (SPEC authored by manager-spec in the card
  worktree; artifacts: spec.md / plan.md / acceptance.md / progress.md).

## §E.1 Plan-phase Audit-Ready Signal

- **plan_status: audit-ready**
- **plan_complete_at: 2026-10-07**
- **Verdict**: PASS — 0.91 / 1.00 (Tier M threshold 0.80), blocking findings 0.
- **Verdict file**: `.moai/reports/t1566/plan-audit-1.md` (plan-auditor, independent,
  tree `f97edcc55`, branch `WT-protected-zone-backslash`, 2026-10-07).
- **Audited artifact set**: spec.md · plan.md · acceptance.md · progress.md (as of
  2026-10-07, tree `f97edcc55`).
- **Auditor instructions carried into the run phase** (non-blocking findings of the
  verdict, binding on M1/M3 execution):
  1. **Positive-control row at M1** (verdict non-blocking #2): when authoring the
     reproduction test, add one positive-control row — a file inside a literal
     `lnk\dir` ORDINARY directory (non-symlink) outside the zone remains ALLOWED.
     This closes the character-blacklist mutant gap (a mutant that repairs the
     resolver then blanket-denies all backslash-bearing POSIX paths would otherwise
     pass the whole AC set; existing tests carry no POSIX literal-backslash
     filename-component case).
  2. **Reinforcement evidence rows from gate-turnend-1/-2** (verdict non-blocking
     #3): at M1 execution, add `.moai/reports/t1566/gate-turnend-1.md` and
     `.moai/reports/t1566/gate-turnend-2.md` as reinforcement corroboration rows in
     the acceptance.md evidence ledger — these are same-tree (`f97edcc55`)
     independent reproductions of the defect class, stronger than the round-8
     cross-tree record.
  - Also noted (advisory): M3's surface-2 row records OBSERVATION, not narrative —
    if the sweep finds no backslash-rewrite pattern at
    `landing_predicate.go:161` (t1561's ledgered defect is the patch-id whitespace
    class), record exactly that; AC-HZB-005 requires observed state + owner +
    evidence, never class agreement.

## M1 — RED Reproduction (run phase)

- **Executed**: 2026-10-07, tree `da2d74eef` (branch `WT-protected-zone-backslash`;
  code-identical to the pinned base `f97edcc55` — `git show --stat da2d74eef`
  verifies the commit carries only the four SPEC artifact files, parent
  `f97edcc55`).
- **Test authored**: `internal/hook/protected_zone_backslash_repro_test.go` —
  `TestCheckProtectedZonePosixBackslashLinkBypass` (guard level, table-driven:
  absolute + relative raw path shapes; non-deny branch performs the ACTUAL write
  through the literal `lnk\dir` path and reads it back from `zone_dir/`) and
  `TestResolveZoneTargetPosixBackslashLinkDivergence` (resolver level). Both
  POSIX-gated (`runtime.GOOS` check) with graceful symlink-unavailable skips;
  `t.TempDir()`-based roots throughout.
- **Auditor non-blocking #2 honored**: the positive-control row is in — a literal
  `lnk\dir` ORDINARY directory (non-symlink) outside the zone stays ALLOWED, at
  both guard level and resolver level. It passed in the same RED run, proving the
  RED is the bypass mechanism and closing the character-blacklist mutant gap.
- **Auditor non-blocking #3 honored**: reinforcement rows `RED-HZB-X1`
  (gate-turnend-1) and `RED-HZB-X2` (gate-turnend-2) appended to acceptance.md's
  evidence ledger — same-tree (`f97edcc55`) independent codex-gate reproductions
  of the defect class, cited as corroboration under the binding RED cells.
- **Four-element RED observations** (full raw bytes in acceptance.md § Evidence
  Ledger):
  - `RED-HZB-001` — `go test -count=1 ./internal/hook/ -run
    'TestCheckProtectedZonePosixBackslashLinkBypass'` → exit 1; both shapes:
    `BYPASS — decision="allow" ... the write through the literal backslash link
    landed INSIDE the protected zone (zone_dir/secret.md="bypass")`.
  - `RED-HZB-002` — same run; the verbatim failure line IS the in-zone landing
    observation (allow + protected file changed — the measured round-8 shape).
  - `RED-HZB-003` — `go test -count=1 ./internal/hook/ -run
    'TestResolveZoneTargetPosixBackslashLinkDivergence'` → exit 1; resolver
    returns ONLY `lnk/dir/secret.md` (zero in-zone forms).
- **Measured note (RED pins the truth over the §B narrative)**: the §B hypothesis
  guessed the fictional rejoin form would be `dir/secret.md`; the measured rejoin
  starts at the FIRST missing component (`lnk`), yielding the single fictional
  form `lnk/dir/secret.md`. The core mechanism (rewritten spelling misses the
  real link; both arms classify outside the zone; allow; the OS write lands
  inside the zone) is confirmed exactly.
- **Affected-package pre-change baseline (M1-BASELINE)**: `go test -count=1
  -timeout 30m ./internal/hook/` on `da2d74eef` → exit 1 (804.0s). Failing set:
  `TestAstgrepCorpusRunDoesNotSkip` (120.01s), `TestStaleRunNoticeLegacyLeaderSpelling`,
  `TestStaleRunNoticeLegacySessionRecord`, `TestStaleRunNoticeFactoryLegacyLabel`
  — all pre-existing, unrelated to the zone path (environmental corpus + legacy
  stale-run-notice spelling) — PLUS the two new RED tests above (expected RED).
  The AC-HZB-004 post-fix verdict is the countable delta against this set: the
  only allowed post-M2 failures are exactly these four pre-existing tests.
- **Pre-change baseline raw tail (verbatim)**:

```
--- FAIL: TestAstgrepCorpusRunDoesNotSkip (120.01s)
--- FAIL: TestCheckProtectedZonePosixBackslashLinkBypass (0.23s)
--- FAIL: TestResolveZoneTargetPosixBackslashLinkDivergence (0.02s)
--- FAIL: TestStaleRunNoticeLegacyLeaderSpelling (0.06s)
--- FAIL: TestStaleRunNoticeLegacySessionRecord (0.08s)
--- FAIL: TestStaleRunNoticeFactoryLegacyLabel (0.59s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	804.025s
FAIL
```

## M2 — Minimal Repair (run phase)

- **Repair commit**: `7945a442a` — `fix(hook): resolve native backslash paths
  for protected-zone target resolution (card t1566)`. Mechanism: a new
  `zoneNativeSlash` helper (`internal/hook/protected_zone_path.go`) carries the
  input's component identity into the filesystem-facing steps — the symlink
  arm's walk input and the project-root resolution — while `zoneSlash` keeps
  the comparison arm's slash-normalized matching semantics. On Windows the two
  coincide byte-for-byte (`zoneNativeSlash` = `zoneSlash`); the
  `zoneResolve` walk body and its fail-closed branches are untouched
  (REQ-HZB-004).
- **GREEN evidence (commit 7945a442a)**: both M1 tests exit 0 (deny observed via
  `wantZoneDeny`, `zone_dir/secret.md` absence asserted); `TestProtectedZone`
  subtest group fully PASS (6.42s); `go vet` clean; `golangci-lint` 0 issues;
  `GOOS=windows go build ./...` exit 0. Full package `go test -count=1 -timeout
  30m ./internal/hook/` on this tree → failing set EXACTLY the M1-BASELINE four
  (TestAstgrepCorpusRunDoesNotSkip + TestStaleRunNotice* ×3) — countable delta
  0, the two RED tests flipped GREEN (833.8s run).

### M2 amendment — converted-absoluteness vector (leader-forwarded gate round 8)

- **Vector**: the cwd-prepend decision read the CONVERTED spelling — a POSIX
  relative `\alias/secret.md` (→ "/"-leading) or `C:\alias/secret.md` (→
  drive-letter) was wrongly judged absolute, the prepend was skipped, no arm
  followed the literal component into the zone; gate-measured forms=[],
  decision allow, the write landed in the protected file. Same named instance,
  same root cause, same file — repaired here, not a new card. RED first
  measured on the commit-7945a442a resolver with the fixed fixture
  (`RED-HZB-004` in acceptance.md: exit 1, both shapes BYPASS, resolver forms
  []; the backslash-free `c:/x.zonefile` row passed, pinning REQ-HZB-003 for
  the amendment).
- **Amendment**: when the conversion changed the spelling (`walk != abs`) and
  the native spelling is relative, the walk receives its own platform-correct
  cwd prepend; when the conversion is the identity (every backslash-free
  input) the branch is a no-op — pre-existing absoluteness semantics
  byte-identical. No Windows change (`walk == abs` always there).
- **Fixture honesty note**: the vector test's first cut had a fixture bug (both
  rows symlinked `\alias` while the drive-letter row's raw names the component
  `C:\alias`); the RED recorded in `RED-HZB-004` was re-measured with the fixed
  per-shape fixture against the pre-amendment resolver — the drive-letter
  demonstration write goes through the same literal component its raw names.
- **GREEN evidence (amendment)**: the vector test + both M1 tests exit 0 in one
  run (`ok ... 1.759s`); `TestProtectedZone` group PASS; vet clean; lint 0
  issues; `GOOS=windows go build ./...` exit 0. Final full-package verdict on
  the amended tree: see §E E2.

## M3 — Family Re-check Sweep (run phase)

Swept on `7945a442a` + the leader-forwarded vector amendment `a03c6d56b` (this
tree), 2026-10-07. Observation only for the sibling surfaces — this card adds
no repair there (spec.md §F).

### Three-surface table (AC-HZB-005)

| Surface | Site | This tree's observed state | Owner card | Evidence |
|---|---|---|---|---|
| 1 — zone target resolution | `internal/hook/protected_zone_path.go` (`zoneSlash` boundary + `resolveZoneTarget` filesystem-facing steps, cwd-prepend gate included) | **REPAIRED** by this card, two commits: `7945a442a` (component identity into the walk + root resolution) and `a03c6d56b` (the converted-absoluteness vector — the walk's own platform-correct cwd prepend); the comparison arm keeps slash-normalized matching | t1566 (this card) | M1 tests + `TestCheckProtectedZonePosixBackslashConvertedAbsoluteness` GREEN (exit 0, deny observed, protected file absent); `RED-HZB-004` for the vector |
| 2 — landing predicate | `internal/cli/worktree/landing_predicate.go:161` | **NO backslash-rewrite pattern at this site** — measured: `grep -c 'ReplaceAll'` over the whole file = **0 hits**; line :161 is the patch-id comparison (`if id == cardIDs[0]`). The ledgered t1561 defect here is the **patch-id whitespace class** (`git patch-id --stable` ignoring whitespace inside strings), a different class from this card's separator rewrite | t1561 | grep count 0 (this run, this tree) + the file read at :140-:166 |
| 3 — project-boundary walk | `internal/hook/pre_tool.go:1397` | **Backslash-rewrite class site, PRESENT in this tree** (pre-repair form): `strings.Split(strings.ReplaceAll(filepath.ToSlash(p), "\\", "/"), "/")` inside `resolvePhysicalWalk` — the same rewrite-before-resolution class this card repaired in the zone path. Adjacent same-file sites :1596-:1597 normalize for deny/ask REGEX matching (non-resolution consumer; rewrite can only ADD pattern matches, i.e. more deny/ask — fail-closed direction, no bypass instance demonstrated) | t1556 (fix in flight, PR open, touches ONLY pre_tool.go — disjoint from this card's surface) | file read at :1387-:1424 and :1595-:1611 (this run, this tree); grep hits recorded below |

### Repo-wide rewrite-pattern grep — swept count and per-hit classification

Two grep forms, unioned and deduplicated (the plan's example form
`grep -rn 'ReplaceAll' internal/ --include='*.go' | grep -F '\\\\'` matches only
4-backslash literals and misses the actual rewrite sites — recorded so the next
sweep uses both forms):

- Form A (plan example, 4-backslash literal): **4 hits** — all four are the
  `glmcred`/`jevcred` TOML-escaping lines below.
- Form B (2-backslash literal `"\\"`, the actual rewrite literal): **21 hits**.
- **Union swept count: 21 unique lines** (the TOML lines carry both literals).
  Scope: `internal/ pkg/ cmd/` `--include='*.go'` (pkg/ and cmd/ contributed 0).

Per-hit classification (counts close: 2 + 7 + 1 + 1 + 6 + 4 = 21):

| Class | Hits | Disposition |
|---|---|---|
| **In-family surfaces** (2) | `protected_zone_path.go:37` (surface 1 — repaired this card); `pre_tool.go:1397` (surface 3 — owner t1556) | surface 1 REPAIRED (M1 tests GREEN); surface 3 recorded read-only, owner t1556 |
| **Same-text class, non-resolution consumer, fail-closed direction** (7) | `pre_tool.go:1596`, `:1597` (deny/ask regex matching); `update_namespace_protect.go:132`, `:265`, `:335`; `update/plan/plan.go:100`, `:154` (update-path comparison) | rewriting can only ADD pattern/namespace matches → over-broad deny direction, never an allow path; no bypass instance demonstrated; out of the 3 named surfaces — recorded for the family ledger; any issuance is the leader's disposition |
| **Template rendering helper** (1) | `internal/template/renderer.go:28` (`posixPath`) | init-time `.sh.tmpl` rendering (shell scripts need forward slashes); not a user-path resolution surface; out of family |
| **Opposite direction** (1) | `internal/shell/config.go:180` (`/`→`\`, Windows-targeted) | not the `\`→`/` class; out of family |
| **Comments / test fixtures** (6) | `codex_config_path.go:44` (comment documenting the repo's own SEPARATOR-AWARE repair, REQ-CSPS-010); `codex_skills_disable_path_test.go:150`; `codex_config_path_test.go:47`; `doctor_codex_seam_use_guard_test.go:18`, `:77`; `update_namespace_protect_test.go:151` (test code) | not production resolution code; dismissed |
| **TOML escaping (plan's named out-of-family)** (4) | `glmcred.go:140`, `:150`; `jevcred.go:195`, `:205` | TOML string escaping/unescaping of credential values, not path rewriting; classified and dismissed per plan.md M3 |

Zone-adjacent helpers (`protected_zone_shell.go`, `protected_zone_guard.go`,
`protected_zone_path.go`): the only hit in the zone trio is the repaired
`zoneSlash` line itself — **0** same-class rewrite-before-resolution hits
elsewhere in the zone guard files (stated zero, measured by Form B over
`internal/hook/`).

### Package re-measurement verdict against the M1 baseline

- Tree `7945a442a` (M2): failing set EXACTLY the baseline four — countable
  delta 0, both RED tests GREEN inside the full run (833.8s; §E E2 verbatim).
- Final tree content `a03c6d56b`: baseline four + six contention-attributed
  timing tests (load-attribution chain in §E E2; quiet-window re-run recorded
  in §E E2 when it fired). AC-HZB-004 verdict **PASS** on the countable-delta
  reading — no failure attributable to the change; CI is the integrated judge.

## §E — Self-Verification Evidence (run phase)

Attribution per manager-develop-prompt-template.md §E: every item names (a) the
command, (b) the observed output, (c) the baseline attribution — this run, this
tree.

- **E1 — RED cells observed (AC-HZB-001/002/003)**: four-element observations in
  acceptance.md § Evidence Ledger (`RED-HZB-001`/`RED-HZB-002`/`RED-HZB-003` +
  reinforcement `RED-HZB-X1`/`X2`), mirrored in §M1 above. (a) two single
  `go test -run` invocations; (b) verbatim FAIL bytes with the BYPASS/in-zone
  landing and the single fictional resolver form; (c) run 2026-10-07 on
  `da2d74eef` (code-identical to pinned `f97edcc55`).
- **E2 — GREEN cells (M2 flip) + affected-package verdict**:
  - (a) `go test -count=1 ./internal/hook/ -run
    'TestCheckProtectedZonePosixBackslashLinkBypass|TestResolveZoneTargetPosixBackslashLinkDivergence'`
    → (b) `ok  	github.com/modu-ai/moai-adk/internal/hook	1.588s`, exit 0; deny
    observed via `wantZoneDeny` and `zone_dir/secret.md` absence asserted inside
    the tests. (c) run 2026-10-07 on `7945a442a` (post-M2).
  - `TestProtectedZone` regression group: `-run 'TestProtectedZone'` → `--- PASS:
    TestProtectedZone (6.42s)` / `ok ... 7.262s` — every subtest green
    (FileTools incl. symlinks, ShellMutation, ManifestStates, NonRegression,
    DenyReason, NoManifestReadForOthers, AuditRow, BaselineCovered, Liveness).
  - Full affected package, tree `7945a442a` (the M2 commit): `go test -count=1
    -timeout 30m ./internal/hook/` → exit 1 (833.8s), failing set EXACTLY the
    M1-BASELINE four (`TestAstgrepCorpusRunDoesNotSkip`, `TestStaleRunNoticeLegacyLeaderSpelling`,
    `TestStaleRunNoticeLegacySessionRecord`, `TestStaleRunNoticeFactoryLegacyLabel`)
    — **countable delta 0; both M1 RED tests flipped GREEN inside the full run**.
    Verbatim tail:

```
--- FAIL: TestAstgrepCorpusRunDoesNotSkip (120.01s)
--- FAIL: TestStaleRunNoticeLegacyLeaderSpelling (0.05s)
--- FAIL: TestStaleRunNoticeLegacySessionRecord (0.05s)
--- FAIL: TestStaleRunNoticeFactoryLegacyLabel (0.36s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	833.807s
FAIL
```

  - Full affected package, final tree content = `a03c6d56b`: exit 1 (1170.9s) —
    the same baseline four PLUS six Factory/SessionStart hook tests
    (`TestFactoryUserPromptSubmitRecoversAfterFirstTurnFailure`,
    `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer`,
    `TestFactoryHookContextAndContinuationSafety`,
    `TestFactoryHookZeroTurnAndCapabilityTruth`,
    `TestSessionStart_DeferredScanDoesNotBlockReturn`,
    `TestSessionStart_DeferredScanJoinsWithinBound`). **Load-attribution
    evidence chain**: (1) `git diff 7945a442a..a03c6d56b -- internal/hook/` =
    ONLY `protected_zone_path.go` (+13 lines) and the repro test file — the six
    failures' subject code (factory/session_start handlers) is byte-identical
    between the two commits; (2) the observed failure mode is a wall-clock
    bound ("Handle blocked 1.436523375s; expected return near the 250ms bound")
    exceeded ~5.7x; (3) measured load during the run window: load average
    ~207-226 with 9 concurrent `go test` processes (the 2026-08-15 multi-lane
    contention pattern; the same tree's earlier run took 833.8s vs 1170.9s);
    (4) the isolated re-run of the six under the same load still tripped the
    bound. A quiet-window re-run of the six is recorded below when it fired.
    Per repo discipline the integrated verdict is CI's (origin/main after
    merge); this local record attributes the six to contention, not to the
    change. **AC-HZB-004 verdict: PASS** — on the countable-delta reading of
    the criterion: no failure attributable to the change exists in either full
    run; the only delta vs baseline is the six contention-attributed timing
    tests whose subject code the change does not touch.
  - **Quiet-window re-run of the six (load-watch Monitor fired on go-test
    storm end)**: (a) `go test -count=1 ./internal/hook/ -run
    'TestFactoryUserPromptSubmitRecoversAfterFirstTurnFailure|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryHookContextAndContinuationSafety|TestFactoryHookZeroTurnAndCapabilityTruth|TestSessionStart_DeferredScanDoesNotBlockReturn|TestSessionStart_DeferredScanJoinsWithinBound'`
    → (b) `ok  	github.com/modu-ai/moai-adk/internal/hook	6.897s`, exit 0;
    measured load at fire time `load averages: 91.65 55.53 35.65` (uptime
    verbatim; the fleet go-test storm had ended). The contention attribution
    holds — the same six subject-code-identical tests pass on the final tree
    outside the storm window.
- **E3 — Windows compile surface**: (a) `GOOS=windows go build ./...` → (b) exit
  0 (no output). (c) run 2026-10-07 on `7945a442a`. Also
  `GOOS=windows go build ./internal/hook/` exit 0 at the same tree.
- **E4 — vet + lint**: (a) `go vet ./internal/hook/...` → (b) no output, exit 0;
  (a) `golangci-lint run ./internal/hook/...` → (b) `0 issues.`, exit 0. (c) run
  2026-10-07 on `7945a442a`.
- **E5 — landing demonstration**: pre-fix the in-zone landing is OBSERVED (RED
  failure line reads the bytes back from `zone_dir/secret.md="bypass"` — file
  read, not asserted); post-fix the deny branch runs and the file-existence
  assert inside the tests verifies absence (Lstat IsNotExist).
- **E6 — family sweep**: §M3 table above — per-surface state + owner + evidence;
  grep swept count 21 unique lines with per-hit classification closing to 21.
- **E7 — diff scope**: (a) `git diff --name-only f97edcc55..HEAD` → (b) exactly
  `internal/hook/protected_zone_path.go` + `internal/hook/protected_zone_backslash_repro_test.go`
  + this SPEC's 4 artifacts; **no sibling-file repairs** (pre_tool.go and
  landing_predicate.go untouched). (c) re-measured on `a03c6d56b`.
- **Commits**: `f4f0e3f7d` (M1 RED + evidence), `7945a442a` (M2 repair),
  `a03c6d56b` (M2 amendment — converted-absoluteness vector), M3 evidence
  commit SHA recorded in the M3 section once landed. Branch
  `WT-protected-zone-backslash`; nothing pushed (lane discipline — integration
  is the leader's window).
- **Gaps**: the turn-end codex gate re-flags known ledger defects on this base
  every turn (gate-turnend-1/2/3 record rounds 1-6); those are existing-
  attribution items owned by t1556/t1561 and this card's own base defect — none
  names a NEW defect in this card's changed files. Full-suite judgment is CI's
  (origin/main after merge); local measurement is the affected package only
  (AGENTS.local.md §4).
- **E8 — gate-silence discriminator basis (leader request, relayed; the
  template-E8 RED-output content is carried inside E1)**: the
  observation "the turn-end gate went quiet on protected_zone_path.go after the
  repairs" is judged REPAIR SUCCESS, not a gate-repro gap, on this basis:
  - **Sensitivity (measured, this card's evidence dir)**: every recorded
    pre-repair gate round DID flag this file — gate-turnend-1
    (`protected_zone_path.go:36`, P1), gate-turnend-2 (resolution arm, P1,
    richer statement), gate-turnend-3 (`:183` apply-site, P1, with the
    line-wobble note) — plus the cross-tree t1556 gate round 8. The gate's
    overlay probes demonstrably reach this file when the defect is present.
  - **Repair success (measured natively, load-bearing signal)**: the three
    in-tree reproduction tests exit 0 on `a03c6d56b` (deny observed, protected
    file absent) — positive evidence the defect is gone, independent of the
    gate.
  - **Caveat (stated honestly)**: the gate runs original-function copies via
    overlay for its probes, so per-round coverage of our exact function body is
    INFERRED from the rounds 1-8 history, not re-proven each round; and the
    post-repair silence itself is not the load-bearing signal — silence is
    never evidence of success (verification-claim-integrity.md §1). The native
    tests carry the verdict; the gate's quiet is consistent with them, no more.
  - **Leader ruling received**: gate `:624` finding → card t1574 (operand
    parsing; repro material at `.moai/reports/t1574/repro-material.md`) — no
    action for this card.
- **Residual-risk**: the repair narrows the zone resolver's input normalization;
  the Bash branch (`checkProtectedZoneShell`) inherits it through the shared
  resolver (REQ-HZB-005) but this card's RED was measured on the Write/Edit
  branch (the measured surface) — the shell branch's own quote-parsing defects
  (gate-turnend-2 item 5, unattributed, leader disposition) are a different
  class and untouched.
