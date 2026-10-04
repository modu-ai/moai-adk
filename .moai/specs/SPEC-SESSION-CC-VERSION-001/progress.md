# SPEC-SESSION-CC-VERSION-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored (Tier M set): `spec.md`, `plan.md`, `acceptance.md`, plus this
  `progress.md`.
- Tier: M (justification in `plan.md` §B — ~13 files, 2 packages, no doctrine or docs-site
  sweep). Threshold 0.80.
- SPEC ID regex check executed as Bash, output `PASS`. ID uniqueness confirmed:
  `ls .moai/specs | grep -i "SESSION-CC-VERSION"` returns no match.
- Budget: **10 requirements, 10 acceptance criteria** (ceilings 16 / 16).
- Every absence-based criterion carries a pre-change baseline measured in this tree at
  `30ce3a02d` (branch `WT-session-cc-version`, cut from develop).
- Evidence base: `.moai/reports/t1348/verdict.md` items ② and ④, every code citation
  re-verified in this tree before use.
- `moai spec lint SPEC-SESSION-CC-VERSION-001` → `✓ No findings — all SPEC documents are
  valid`, exit 0 (after two fix rounds: AC→REQ mappings restated in the `maps REQ-…` house
  form, and `-run` selectors anchored `'^Test…$'`; the first measurement's 20 warnings
  included 2 hidden by a `tail`-truncated read — full output captured on re-run).
- **Plan-audit iter1: FAIL 0.81** (Tier M threshold 0.80 met numerically; FAIL carried by
  blocking defects D1-D4; report `.moai/reports/t1465/plan-audit-iter1.md`). All six deltas
  repaired in place (no requirement rewrites):
  - **D1 (critical)** — the documented emergency form `moai cc -l --name lane-<n>` is itself
    refused (`laneFlagNameError`, `factory.go:391-394`; `operatorSuppliedName`,
    `factory_launch_helpers.go:390-404`). Corrected to the bare lane join
    `moai cc -l -- --resume <session-id>` in spec §A.3, plan M4 + §F.1 + §F.3, AC-SCV-010;
    the refusal re-observed by running this tree's build from /tmp with the env stamps
    scrubbed → `ERROR: -L/--Lane already names the role; drop the --name/-n flag.`, exit 1.
  - **D2 (major)** — the iter-0 lsof baseline did not reproduce (real: 2 comment hits under
    `internal/session/`; 4 exec sites repo-wide, all cwd/port). Restated in spec §A.1 and
    AC-SCV-001 from re-measured greps.
  - **D3 (minor)** — the exit-0 assertion added to AC-SCV-005, where AC-SCV-003 delegates it.
  - **D4 (major)** — AC-SCV-004's selector widened to the three real test names as an anchored
    alternation, with a swept-count-3 + no-`[no tests to run]` requirement.
  - **D5 (adopted)** — both `--resume` spellings (`--resume <v>`, `--resume=<v>`) pinned in
    REQ-SCV-009/010, AC-SCV-009/010, plan M4.
  - **D6 (adopted)** — the lsof txt parse anchored to the claude-binary line in spec §A.2 and
    plan M1.
- Version 0.1.0 → **0.1.1** (HISTORY row added). The REQ layer survives intact per the
  auditor; `updated:` fields refreshed (same-day).
- Status: `draft`. Plan-audit iter1 repaired; **iter2 pending**.

### Card stage plan (lane-23, card t1465)

This plan phase is the card's first stage. The lane task list carries the remaining stages in
order: plan-audit (independent audit) → plan→run kickoff decision record → run (manager-develop,
M1-M5 of `plan.md` §D) → verification batch (env-scrubbed build + lint + targeted tests) →
sync (manager-docs) → card-review + merge-ready report to the leader. The lane orchestrator
commits these artifacts; nothing is committed or pushed by the plan phase.

### Assumptions made at plan phase

1. **Flag name `--cc-version`** on `moai session list` — a naming choice. The wire contract
   that matters (additive JSON fields under the flag, default path byte-identical) is pinned
   by REQ-SCV-005/006; renaming the flag before run is cheap.
2. **The doctor check is advisory (WARN-only, never `CheckFail`)** — in the `checkFlagSlot`
   pattern. An operator wanting staleness to gate `moai doctor`'s exit status would be a
   separate decision, not this SPEC's.
3. **`unknown` never warns** in the doctor check: staleness cannot be judged from unknown, and
   warning on it would nag npm-style installs with unversioned binary paths.
4. **The relaunch guard refuses the whole loop** rather than stripping the token or honoring it
   once: under `relaunch` a resume token cannot mean what it says (the loop leases new cards),
   so refusing with the safe form in the error text is the honest behavior.
5. **The pure assembler (REQ-SCV-008) changes no one-shot behavior** — the pass-through
   already survives the launcher mechanically (`spec.md` §A.3.1); the assembler names and
   tests what exists, and M4's only behavior change is the relaunch guard.
6. **Versions are never persisted** into the registry file — the live read at query time is
   the whole truth (`spec.md` §D, last exclusion).
7. **Installed-version read is path-parse only** — no `claude --version` exec (REQ-SCV-002);
   installs whose resolved path carries no version segment degrade to `unknown`.
8. **The exact spelling `moai cc -f lane-<n> -- --resume <id>` from the dispatch is refused
   today** (`factoryFlagUsageError`, `factory.go:92`); the SPEC treats the lane-join form as
   the emergency path and has the run-phase observe the refusal verbatim (`plan.md` §F.2).

## §E.2 Run-phase Evidence

### Pre-flight baselines (measured before M1, tree `3dc8c9760`)

| # | Command | Observed result |
|---|---|---|
| 1 | `git branch --show-current` + `git rev-parse --short HEAD` | `WT-session-cc-version` @ `3dc8c9760`, clean tree |
| 2 | `go build ./...` | exit 0 |
| 3 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| 4 | `golangci-lint run --timeout=2m ./internal/session/... ./internal/cli/...` | `0 issues.` (installed build: v2.1.6 — the CI-pinned version) |
| 5 | `grep -rn "Retired\|TestHarnessRetirement\|superseded" internal/session/ internal/cli/` | hits only in unrelated retirement records (`cg.go` retired launcher, `harness_route.go` supersedence note) — no policy conflict with this SPEC's surface |

### M1 — The version reads behind seams

**E6 RED evidence (TDD)** — the three fixture tests of AC-SCV-001..003, run against
signature-only stubs (no reader logic; bodies returned zero values), before any reader
existed. Command: `go test ./internal/session/ -run '^(TestRunningVersionFromInjectedMapping|TestInstalledVersionFromResolvedPath|TestVersionDegradationRendersUnknown)$' -v`
→ exit **1**; verbatim (abridged to the assertion lines — full log retained):

```
=== RUN   TestRunningVersionFromInjectedMapping
    ccversion_test.go:49: ResolveCCVersions(4242).Running = "", want 2.1.281 (a library mapping's version shape must not satisfy the read)
--- FAIL: TestRunningVersionFromInjectedMapping (0.00s)
=== RUN   TestInstalledVersionFromResolvedPath
=== RUN   TestInstalledVersionFromResolvedPath/versions_shape
    ccversion_test.go:66: installed version = "", want 2.1.288
=== RUN   TestInstalledVersionFromResolvedPath/claude-code_shape
    ccversion_test.go:72: installed version = "", want 2.1.284
=== NAME  TestInstalledVersionFromResolvedPath
    ccversion_test.go:87: versionSegmentFromPath("/Users/dev/.local/share/claude/versions/2.1.281/claude") = "", want "2.1.281"
    ccversion_test.go:87: versionSegmentFromPath("/opt/node/lib/node_modules/@anthropic-ai/claude-code/2.1.284/cli") = "", want "2.1.284"
--- FAIL: TestInstalledVersionFromResolvedPath (0.01s)
=== RUN   TestVersionDegradationRendersUnknown
    ccversion_test.go:133: dead pid running = "", want "unknown"
    ccversion_test.go:141: probe error running = "", want "unknown"
    ccversion_test.go:152: unversioned running = "", want "unknown"
    ccversion_test.go:168: unsupported platform running = "", want "unknown"
--- FAIL: TestVersionDegradationRendersUnknown (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/session	0.317s
```

**GREEN** — the same selector after the readers landed → exit 0; verbatim tail:

```
--- PASS: TestRunningVersionFromInjectedMapping (0.00s)
--- PASS: TestInstalledVersionFromResolvedPath (0.01s)
--- PASS: TestVersionDegradationRendersUnknown (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/session	0.355s
```

**M1 milestone gate** — `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...`
→ exit 0; `go test ./internal/session/` → `ok … 17.618s` exit 0 (whole package, no selector);
`golangci-lint run --timeout=2m ./internal/session/... ./internal/cli/...` → `0 issues.`.

**Gaps (M1)**: none — every AC-SCV-001..003 command was run in this phase on this tree.
**Residual-risk (M1)**: the `ccversion_other.go` runtime path cannot execute on this darwin
machine; its contract is carried by the `unsupported platform read` subtest standing in
through the shared seam, plus the windows cross-build proving the file compiles. The darwin
`lsof` exec site itself is seam-excluded from unit coverage per REQ-SCV-004 (no test spawns a
process); its real-world behavior is measured live in §F.4 below.

### M2 — `moai session list --cc-version`

**RED** (flag absent) — `go test ./internal/cli/ -run '^(TestSessionListCCVersion|TestSessionListDefaultNoProbe)$' -v`
→ exit 1; cobra `unknown flag: --cc-version` usage error failed every `TestSessionListCCVersion`
subtest, and the key-set comparison surfaced a test-side ordering bug (fixed: both sides
sorted — the criterion is set equality). **GREEN** → exit 0: `ok github.com/modu-ai/moai-adk/internal/cli`.

**M2 milestone gate** — builds native + windows exit 0; `golangci-lint … ./internal/cli/...` →
`0 issues.`; the full cli suite under the default 10-minute timeout exceeded it on this loaded
machine (see Gaps) — the session-scoped family (`-run 'TestSession'`) passed: `ok … 15.996s`.

### M3 — The doctor staleness check

**RED** (check absent) — `go test ./internal/cli/ -run '^TestDoctorCCVersionStaleness$' -v` →
exit 1, `[build failed]` (`undefined: checkSessionCCVersionStaleness`, `undefined: doctorCCVersionEntries`).
**GREEN** → exit 0: 7 subtests `--- PASS`. First GREEN attempt failed with
`expected 'package', found 'import'` (missing `package cli` declaration) — repaired, GREEN.
The doctor golden snapshots caught the new row (4 golden tests failed); regenerated with
`UPDATE_GOLDEN=1` and verified green — the diff is exactly one added `ok` row
(`Session CC Version — no active sessions registered — nothing to compare`) and the count
`13 ok → 14 ok`. Check name registered in the binary-lag allowlist (`binary_lag_test.go`).

### M4 — The resume emergency path

**RED** (pieces absent) — the three test functions → exit 1, `[build failed]`
(`undefined: laneJoinChildArgv`, `undefined: validateResumeArgs`). **GREEN** → exit 0 after
implementation; the first GREEN run's positive-control subtest (`no resume token enters the
loop`) failed on `factoryAssertParentCheckout` (the test binary runs in the worktree, not the
primary checkout) — repaired by reusing the rerun SPEC's `relaunchLoopDrive` harness, which
drives the loop against a fixture primary checkout. `TestRelaunchLoopReResolvesRun` (the
existing loop contract) unchanged and passing.

**M4 milestone gate** — builds exit 0; the launcher-scoped family
(`TestParseLauncherEntry|TestParseFactoryFlag|TestLauncher|TestLaneJoin|TestResume|TestRelaunch|TestCCWorktree|TestSpawn|TestSession|TestDoctorGolden|TestBinaryLag|TestProfileFlag`)
→ `ok … 36.530s`; lint `0 issues.`. A wider `-run 'TestFactory…'` scoped batch exceeded its own
8-minute timeout with zero failures — the running-at-alarm test
(`TestFactoryNextNominateRecordStateTokens`) passes standalone (`ok … 68.848s`; real git
fixtures per subtest — queue latency on a loaded machine, not a defect).

### M5 — §F live measurements (E7)

- **§F.1 — the one-shot emergency form, end to end.** Mechanism note (measured, not assumed):
  the launcher's debug dump prints step timings only (`moai-launcher-debug: lane claim took …
  (label=lane-1)`), NOT the child argv — so the argv observation was obtained by placing a
  stub `claude` script on PATH (prints its argv, exits 0 — no interactive child, no claude
  process left running; pgrep verified). Command: fixture git repo at
  `/tmp/t1465-probe/project` with a seeded active run row, then `env -u <factory stamps>
  CLAUDE_PROJECT_DIR=… PATH=<stub dir>:… /tmp/t1465-moai cc -l -d -- --resume probe-t1465`.
  The stub observed the child argv **verbatim**:

  ```
  1=[-d]
  2=[--name]
  3=[lane-1]
  4=[--resume]
  5=[probe-t1465]
  6=[--settings]
  7=[/var/folders/…/moai-factory-93761-1791100496168819000.json]
  ```

  All three carry: the desugared `--name lane-1` (the launcher claims the lane name), the
  injected settings pair, and the pass-through `--resume` token (exit 0; `-d` forwarded
  verbatim per its observe-only contract).
- **§F.2 — the refused spelling, observed.** From /tmp with the factory stamps scrubbed:
  `/tmp/t1465-moai cc -f lane-3 -- --resume probe-t1465` → exit 1, rendered
  `-F/--Factory takes no argument (bare -f starts the factory leader); a lane joins with -l or
  --lane, got "lane-3".` (`factoryFlagUsageError`).
- **§F.3 — the relaunch guard, observed.** Same fixture, seeded active run:
  `/tmp/t1465-moai cc -l --clear-policy relaunch -- --resume probe-t1465` → exit 1, rendered
  `Factory lane: --resume cannot run under --clear-policy relaunch — every card session the
  loop starts would resume the same conversation in a foreign worktree; the emergency form is
  the one-shot lane join: moai cc -l -- --resume <session-id>.` — the safe form named
  verbatim; zero card sessions started (the guard precedes the parent-checkout assertion and
  the loop's first iteration).
- **§F.4 — the live lsof positive control.** `lsof -a -d txt -p 3900` (the live lane-1 claude
  process) → the binary mapping line
  `2.1.287 3900 goos txt REG … /Users/goos/.local/share/claude/versions/2.1.287` — **which
  caught a real M1 defect**: the native installer ships the binary NAMED BY ITS VERSION, so
  the original `/claude`-suffix anchor rendered unknown for every real session. Repaired
  (commit `312ff5c47`): the anchor now satisfies the product-directory layout
  (`…/claude/versions/…`), the npm layout (`…/claude-code/…`), and the binary-named tail;
  the case-sensitive match still refuses macOS frameworks' capitalized `Versions/<n>`.
  Re-verified end to end on this machine: fixture registry pointing at pid 3900 →
  `/tmp/t1465-moai session list --cc-version` → `cc=2.1.287 installed=2.1.289` (readlink
  confirms `~/.local/bin/claude → …/versions/2.1.289`); `moai doctor --check "Session CC Version"`
  → `warn Session CC Version  a session runs Claude Code 2.1.287 but the installed version is
  2.1.289 — …` with **exit 0** — the staleness this SPEC exists to surface is live on this
  machine, and the advisory property (doctor's exit status unchanged) is observed, not assumed.

### Final verification batch (tree `312ff5c47`)

- **E1** — all ten AC commands, acceptance.md forms verbatim: AC-SCV-001..010 exit 0 each
  (logs `/tmp/t1465-ac1..10.log`; session package `ok`, cli package `ok`). AC-SCV-004's
  swept count: `--- PASS` ×11 (3 top-level + 8 subtests), `[no tests to run]` ×0,
  `grep -n "exec.Command" internal/session/ccversion*_test.go` → 0 hits.
- **E2** — `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E3** — coverage: see the coverage lines below (measured on this tree).
- **E4** — `grep -rn "AskUserQuestion" internal/session/ internal/cli/ | grep -v _test | grep -v "// "`
  → 3 hits, ALL pre-existing at the base SHA (`internal/cli/harness.go:224/226/291` — the
  harness-learner boundary's own documentation strings, verified via
  `git grep … 3dc8c9760`), and `git diff 3dc8c9760..HEAD -- internal/cli/harness.go` is
  empty — **0 new hits**.
- **E5** — `golangci-lint run --timeout=2m ./internal/session/... ./internal/cli/...` (v2.1.6):
  baseline `0 issues.` → final `0 issues.` — **0 new issues**.
- `go vet ./internal/session/ ./internal/cli/` → exit 0.
- **PRESERVE** — `git diff --name-only 3dc8c9760..HEAD` outside `internal/session/`,
  `internal/cli/`, `.moai/specs/` → empty.
- **E8** — commit SHAs (one per milestone): M1 `dec4fd34a`, M2 `7cf7ad0fb`, M3 `3ca2f3a5b`,
  M4 `2ee51ef7f`, M5 anchor repair `312ff5c47`.

### Gaps (run-phase, explicit)

1. **The Linux running read has no live measurement** — `ccversion_linux.go`
   (`/proc/<pid>/exe`) is fixture-tested only; this darwin machine cannot execute it (plan
   §F.4 anticipated exactly this gap).
2. **The `ccversion_other.go` runtime path** cannot execute here; its contract is carried by
   the seam stand-in subtest plus the windows cross-build.
3. **The darwin lsof exec site is seam-excluded from unit coverage** (REQ-SCV-004 forbids a
   test spawning a process); its real-world behavior is covered by the §F.4 live measurement.
4. **The full cli suite exceeds go test's default 10-minute timeout on this loaded machine**
   (observed twice: 601s at M2, an 8m scoped batch at M4) — the only failing member both
   times was the pre-existing `TestStopChainMemberCostWithinBudget` timing-budget test
   (codex stop-hook chain, domain-disjoint from this SPEC's diff; fails standalone under
   load: hook member 1.46s > its declared 1s budget). Every scoped family touching this
   SPEC's surfaces passed.

### Residual-risk (run-phase)

- The installed read resolves the FIRST `claude` on PATH — a PATH-shadowed install reads that
  one (that is the definition of "installed" this SPEC chose: what the next launch would run).
- The staleness compare assumes numeric dot segments; the parser only ever produces those, so
  a non-numeric segment reads as "not older" rather than guessing.
- The relaunch guard lives at `runFactoryLaneRelaunch`'s entry; a future second relaunch door
  must re-apply it (the M4 comment names the REQ).

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_complete_at: 2026-10-04
- AC matrix: **10/10 PASS** (AC-SCV-001..010, each command run verbatim from
  `acceptance.md` on this tree; §E.2 Final verification batch carries the attributions)
- Builds: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- `go vet ./internal/session/ ./internal/cli/`: exit 0
- Lint: golangci-lint v2.1.6 — baseline `0 issues.` → final `0 issues.` (0 new)
- E4 boundary grep: 0 new hits (3 pre-existing `internal/cli/harness.go` doc strings, present
  at the base SHA, file untouched by this run)
- Coverage (new files, `go tool cover -func` on this tree):
  - `internal/session/ccversion.go` — runningCCVersion 100%, mappingPathNamesClaudeBinary
    100%, versionSegmentFromPath 100%, runningCCVersionFromMapping 90%,
    installedCCVersion 77.8% (aggregate ≈ 94%)
  - `internal/cli/lane_resume.go` — 100% (all four functions)
  - `internal/cli/doctor_ccversion.go` — checkSessionCCVersionStaleness 86.8%,
    ccVersionOlder 90.0% (aggregate ≈ 88%)
  - `internal/session/ccversion_darwin.go` — 0% by design: the lsof exec site is
    REQ-SCV-004-excluded from unit tests (no test spawns a process); its real-world behavior
    is verified live by §F.4 (the measurement that caught the anchor defect)
  - `internal/session` package total: 85.9% of statements
- §F live measurements: F.1/F.2/F.3/F.4 all observed (none inferred); §F.4 caught and repaired
  a real M1 defect (commit `312ff5c47`), then verified the full chain live — running 2.1.287 vs
  installed 2.1.289 on this machine, doctor warn naming both with exit 0
- Commits (E8): `dec4fd34a` M1 · `7cf7ad0fb` M2 · `3ca2f3a5b` M3 · `2ee51ef7f` M4 ·
  `312ff5c47` M5 anchor repair — branch `WT-session-cc-version`, base `3dc8c9760`
- PRESERVE: `git diff --name-only 3dc8c9760..HEAD` outside `internal/session/`,
  `internal/cli/`, `.moai/specs/` → empty; `registry.go` untouched
- sync-phase handoff note for manager-docs: the doctor check's name is
  `Session CC Version` (`--check "Session CC Version"`); CHANGELOG should name the
  `--cc-version` flag, the doctor check, the assembler/validation/guard trio, and the
  REQ-SCV-010 refusal text's verbatim form `moai cc -l -- --resume <session-id>`

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-04
sync_commit_sha: pending-backfill-sync
status_transition: in-progress → implemented → completed, frontmatter `status` only, riding the single sync commit (spec.md `status: completed`; `updated: 2026-10-04` already carried today's date, so the frontmatter's only changed line is `status`)
sync_commits: the one sync commit (subject `chore(SPEC-SESSION-CC-VERSION-001): sync-phase artifacts — 3-phase close (card t1465)`); its own SHA is backfilled by a following commit per the D3 exemption
files_changed_by_sync: `CHANGELOG.md` (one `[Unreleased]` / Added entry), `.moai/specs/SPEC-SESSION-CC-VERSION-001/spec.md` (frontmatter `status` only), this file (§E.4)
sync_scope_note: Tier M internal CLI feature — no README or docs-site sweep per plan §B; codemap regeneration not run (owned by the periodic codemaps cards, t1443/t1456 pattern)
b12_self_test: pre-emission `grep -c SPEC-SESSION-CC-VERSION-001 CHANGELOG.md` was 0 before the entry and is 1 after; live AC count 10 of 10 (acceptance.md carries no `[RETIRED]`/`[REF]` markers), matching the entry's "10 acceptance criteria AC-SCV-001..010"; every file path named in the entry ls-verified
handoff_followed: §E.3's sync-phase note — the entry names the `--cc-version` flag, the doctor staleness check, the assembler/validation/guard trio, and the verbatim emergency form `moai cc -l -- --resume <session-id>`

## §F Phase 4 Mode Selection

Input parameters: tier M; scope ~13 files across 2 packages (`internal/session`, `internal/cli`); domain count 2 (Go source only); file language mix 100% Go; concurrency benefit LOW (coding-heavy implementation, coupled milestone ordering M1→M5); Agent Teams prereqs not requested (no operator `--team`).

| Mode | Selected | Rationale |
|---|---|---|
| `direct` | not selected | Semantic multi-file feature work, not a typo/single-line fix |
| `serial` | **selected** | Coding-heavy work in 2 coupled packages; single writer in the card worktree; per-Anthropic coding-task parallelism caveat |
| `fanout` | not selected | No independent multi-domain research split; coding tasks favor sequential |
| `sweep` | not selected | ~13 files, semantic new-code work — fails the ≥~30-file mechanical-uniform test |
| `agent-team` | not selected | Explicit-request-only experimental surface; no request |

Decision: `serial`

Justification: the milestones are ordered by decision reversibility (plan §D — M1 seam shape, M4 guard semantics) and each consumes the previous one's types, so parallel spawns would only create integration risk inside one worktree. `serial` keeps one writer per tree and matches the coding-heavy caveat.

Kickoff gate: met in autonomous form — plan-audit iter2 PASS 1.0 (threshold 0.80), artifact hash `49065ad2…` unchanged since the verdict (re-measured this run); decision record: `.moai/reports/t1465/kickoff-decision.md`.
