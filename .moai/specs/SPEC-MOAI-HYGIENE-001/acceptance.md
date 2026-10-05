# acceptance.md — SPEC-MOAI-HYGIENE-001

All scenarios run inside `t.TempDir()` fixture trees unless a scenario reads this repository's own source text (AC-HYG-005, AC-HYG-015, AC-HYG-016 — read-only greps over `internal/`). No scenario touches the repository's real `.moai` tree (REQ-HYG-015, enforced at runtime by the `testing.Testing()` root guard).

**Classification.** Release-blocking (RB): an AC whose failure blocks run-phase exit — every AC that guards mutation safety or a primary behavior of the two units. Regression-guard (RG): an AC whose starting observation is structurally undecidable on the pre-implementation tree (its target files do not exist yet); it pins the green form now and re-executes at its flipping milestone, and is never recorded as a pass before that flip (verification-completeness.md §2.1 undecidable disposition).

**Evidence ledger.** §A table cells cite ledger entries L-001..L-016. Every RB AC carries a four-element RED-now cell (command, verbatim stdout, exit code, tree SHA) observed by direct execution on the pre-implementation tree; iteration 2 of the plan audit re-executed L-001/L-013/L-014/L-015/L-016 on the repair commit and reproduced them byte-identically. L-012 was re-captured at v0.3.0 under the scenario's renamed test (same pre-implementation shape).

## §A AC Matrix

| AC | Scenario (maps REQ-HYG-…) | Class | Ledger | GREEN then (milestone) |
|---|---|---|---|---|
| AC-HYG-001 | Rotation on over-threshold sink (maps REQ-HYG-001, REQ-HYG-002) | RB | L-001 | test passes, exit 0 (M1) |
| AC-HYG-002 | Keep-1 cap + crash-recovery of an orphan staging chunk (maps REQ-HYG-001, REQ-HYG-002) | RB | L-002 | test passes, exit 0 (M1) |
| AC-HYG-003 | Under-threshold and absent sinks untouched (maps REQ-HYG-001) | RB | L-003 | test passes, exit 0 (M1) |
| AC-HYG-004 | Concurrency: stale-decision skip; no fresh-chunk destruction; unverified-exclusion skip (maps REQ-HYG-002) | RB | L-004 | test passes, exit 0 (M1) |
| AC-HYG-005 | Registry completeness guard (maps REQ-HYG-003) | RB | L-005 | guard red on seeded unregistered writer, green on real tree (M1) |
| AC-HYG-006 | Liveness verdict matrix incl. registry-absent / entry-missing / missing-fingerprint / transcript-absent rows (maps REQ-HYG-007) | RB | L-006 | test passes, exit 0 (M2) |
| AC-HYG-007 | Report mode non-mutating: byte-identical tree, no lockfile, ≤1 summary row per unit, over-threshold self-sink skipped (maps REQ-HYG-004, REQ-HYG-010, REQ-HYG-013) | RB | L-007 | test passes, exit 0 (M3) |
| AC-HYG-008 | Apply mode deletes exactly the eligible set; mtime-only candidates kept; per-class dating incl. verify directory + goal-triple grouping (maps REQ-HYG-005, REQ-HYG-008, REQ-HYG-009) | RB | L-008 | test passes, exit 0 (M3) |
| AC-HYG-009 | Audit rows complete, report rows summarized, apply rows per-action (maps REQ-HYG-004, REQ-HYG-009) | RB | L-009 | test passes, exit 0 (M3) |
| AC-HYG-010 | Unit independence (maps REQ-HYG-012) | RB | L-010 | test passes, exit 0 (M3) |
| AC-HYG-011 | Symlink refusal scoped below the resolved root; swap-after-check residual arm (maps REQ-HYG-006) | RB | L-011 | test passes, exit 0 (M3) |
| AC-HYG-012 | Lock class excluded — never evaluated, probed, acquired, or deleted (maps REQ-HYG-011) | RB | L-012 | test passes, exit 0 (M3) |
| AC-HYG-013 | SessionStart wiring best-effort, never blocks launch (maps REQ-HYG-014) | RB | L-013 | test exists and passes, exit 0 (M5) |
| AC-HYG-014 | CLI dry-run default; `--apply` overrides config for that invocation; config alone never mutates the CLI (maps REQ-HYG-013) | RB | L-014 | test exists and passes, exit 0 (M4) |
| AC-HYG-015 | Named thresholds only — single-invocation grep, expected green exit 1 (maps REQ-HYG-016) | RG | L-015, L-C1 | exit 1 (zero hits over existing package) (M1) |
| AC-HYG-016 | Tests never touch the real .moai — escape-pattern grep + runtime temp-root guard (maps REQ-HYG-015) | RG | L-016, L-C2, L-C3 | exit 1 (zero hits) + guard refuses non-temp root (M1/M4/M5) |

### Scenario detail

- **AC-HYG-001** — Given a registered sink of 11 MiB in a temp tree, When the rotator runs, Then the sink becomes an empty primary, the 11 MiB of rows live at `<name>.1`, and one apply-mode audit row records the rotation.
- **AC-HYG-002** — Given a registered sink of 11 MiB with a pre-existing `<name>.1`, When the rotator runs, Then the sequence stages the fresh chunk at `<name>.1.staging`, removes the existing `<name>.1` only while the replacement is staged, and promotes staging to `<name>.1` — the tree never holds more than one rotated chunk per sink. **Crash-recovery arms (RED-first):** given an orphan `<name>.1.staging` plus the old `<name>.1` (crash between staging and removal), the next pass completes the placement — the old chunk is removed, the staged chunk is promoted, and nothing is overwritten; given an orphan staging with no `.1`, the pass promotes it directly; in no arm is a staged chunk ever deleted.
- **AC-HYG-003** — Given registered sinks under 10 MiB and one registered name with no file, When the rotator runs, Then byte hashes of the under-threshold sinks are unchanged, the absent name produces no file and no error, and skipped rows carry their reason.
- **AC-HYG-004** — Given one over-threshold sink and N concurrent rotator invocations, plus (a) a second scenario where a rotator enters its critical section carrying a stale over-threshold observation (modeled through the stat seam) while the sink is actually under threshold, and (b) a third where cross-process exclusion cannot be established (the sidecar-lock seam reports held/unverifiable — the Windows shape), When all complete, Then exactly one rotation occurred under verified exclusion (one `<name>.1`), the stale-decision rotator performed no action and recorded a `skipped-stale` outcome, the unexcluded rotator recorded `skipped-locked`/`skipped-platform` and destroyed nothing, the union of primary + chunk bytes contains every pre-run row, and no invocation ever removed a chunk it did not itself displace.
- **AC-HYG-005** — Given the guard scanning `internal/` for append-only sink writers under `.moai/logs/`, When a fixture source tree (in `t.TempDir`) contains an unregistered writer, Then the guard fails naming it; against the real tree every current writer resolves to a registry entry and the guard passes.
- **AC-HYG-006** — Given the signal-combination table, When the evaluator runs each row, Then the verdicts hold as specified. Rows: pid-alive+fingerprint-match (affirmative); pid-alive+fingerprint-mismatch — reused pid (not affirmative); pid-dead (negative); pid-alive+no-recorded-fingerprint (unmeasured); registry file absent (pid+heartbeat unmeasured); registry present, no entry for the key (pid+heartbeat unmeasured); probe-undetermined (unmeasured); transcript-fresh (affirmative); transcript-stale (negative); **transcript absent under a resolvable root (unmeasured — never negative)**; transcript-root-unresolvable (unmeasured); heartbeat younger than `HygieneHeartbeatStaleWindow` (affirmative); heartbeat older than the window (negative). Any affirmative → LIVE; all-measurable-all-negative → DEAD; any unmeasured with none affirmative → INDETERMINATE carrying the unmeasured signal's name; the registry-absent, entry-missing, and transcript-absent rows can never produce DEAD.
- **AC-HYG-007** — Given a temp tree seeded with dead-datable, live, indeterminate, and over-threshold-sink candidates — **including the hygiene sink itself seeded over threshold** — mode `report`, When the GC and rotator run, Then a whole-tree hash taken before and after is identical (no lockfile created, no candidate or sink rotated, no lock-class interaction), with `.moai/logs/hygiene-audit.jsonl` itself excluded from the compared hash as the named exception, the audit sink gained exactly one rotator summary row and one GC summary row, the rotator's summary records the over-threshold hygiene sink as `skipped-report-mode`, and the report states a keeping reason for every kept candidate.
- **AC-HYG-008** — Given the seed classes: dead+content-datable+aged (deletable), dead+young-datable (kept), dead+body-undatable with a fresh-looking mtime (kept — `mtime-only`), unresolvable-key (kept), LIVE (kept), INDETERMINATE (kept), **a `state/verify/<key>/` directory whose entries individually straddle the age floor (entry-by-entry removal; the directory removed only when empty; undatable entries spared — the dating field per class pinned by this test's RED-first assertions), and a `state/goal` triple dated from its `.json` member (all three removed together, **the dating member last**; no partial deletion)**, When mode `apply` runs, Then only the qualifying classes are gone, every kept class survives, and a re-run reports the already-vanished path as `already-gone` without error. **Writer-race arm (D28, RED-first):** a candidate whose file is replaced between enumeration and action (seeded through the inode seam — the writer's temp-then-rename shape, fresh state) is NOT deleted; the action-time re-stat + content re-read detects the inode change and aborts as `rejudged-keep` with the refreshed state surviving. **Crash-mid-group arm (D29, RED-first):** an interruption after the goal triple's sibling removals leaves the `.json` dating member in place (still-datable group); the next pass completes the group.
- **AC-HYG-009** — Given an apply-mode run that deleted at least one candidate and rotated at least one sink, and a report-mode run over the same seed, When the audit sink is read back (fixture path `.moai/logs/hygiene-audit.jsonl` inside the temp tree), Then every apply-mode mutation carries path, unit, mode, decision, reason, and signal evidence (pid verdict, transcript mtime age, heartbeat age, content-date age), every skipped candidate carries its keeping reason, and the report-mode run carries exactly the two summary rows.
- **AC-HYG-010** — Given an injected rotator failure (unwritable lock path) and, in a second scenario, an injected GC failure (read-only target dir), When the combined pass runs, Then in scenario one the GC still completes its report, in scenario two the rotator still completes, and both failures surface as logged per-unit outcomes — never as a blocked pass.
- **AC-HYG-011** — Given a target whose path contains a symlinked component strictly below the resolved `.moai` root, and a second scenario where the path sits under symlinked ancestors **above** the root (the macOS `/var` → `/private/var` shape — must NOT be refused), and a third where the parent is swapped to a symlink after the initial check (modeled through the resolution seam), When either unit evaluates the paths, Then the first outcome is `symlink-refused` with the path untouched, the second proceeds (the root's own resolution never counts as a violation), and the third is closed by the anchored-handle execution (D27): the swap arm's action is issued through the directory-fd-anchored root handle, so the swapped-in symlink makes the anchored action fail instead of escaping — the swap arm asserts refusal/no-escape, and the residual that remains (a swap to a different real directory inside the root inside the final window) is recorded as the named residual and caught by the next pass's re-check.
- **AC-HYG-012** — Given a scan that encounters `spec-close-<SPEC-ID>.lock` files among otherwise eligible residue, When the GC runs in either mode, Then every lock-named file is reported `lock-class-excluded` and left byte-identical, and the target registry itself contains no lock entry (asserted structurally — no code path exists that probes, acquires, or deletes a lock).
- **AC-HYG-013** — Given a SessionStart invocation with the hygiene engine stubbed to fail, When the hook processes the event, Then the hook exits 0, the failure appears only as a logged warning, and the remaining SessionStart steps complete.
- **AC-HYG-014** — Given a temp-rooted project with eligible residue, When `moai clean --audit-logs` and `moai clean --session-state` run without `--apply`, Then the report prints decisions and reasons and the residue survives; with `config mode: apply` but no `--apply` the CLI still does not mutate; with `--apply` the eligible set is gone. **Config-validation arms (D30, RED-first):** `audit_log_kept_rotations` ≠ 1 in the config ⇒ the run refuses mutation with a `config-invalid` outcome (never clamped silently); any window/age/size override ≤ 0 ⇒ `config-invalid`; an unrecognizable mode string ⇒ the run proceeds as `report` (non-mutating).
- **AC-HYG-015** — Given the `internal/hygiene/` package source, When the single-invocation threshold grep of L-015 runs over the existing package, Then it exits 1 with empty output (no size or duration literal outside `internal/config/defaults.go`); the seeded control L-C1 proves the pattern matches real content so the zero is measured, not vacuous.
- **AC-HYG-016** — Given the test files of the new units and wiring, When the single-invocation escape-pattern grep of L-016 runs over the existing files, Then it exits 1 with empty output — the pattern covers repository `.moai` paths, absolute user paths, **relative `../` escapes toward `.moai`, `os.Getwd()`-derived roots, and `os.UserHomeDir()`-derived roots**; the seeded control L-C2 (absolute) and L-C3 (Getwd/UserHomeDir forms) prove the pattern matches real content; and the runtime guard is asserted separately: with `testing.Testing()` true, an entry point handed a root outside the test's temp directory refuses to run.

## §A.1 Evidence Ledger (RED-now cells, observed on the pre-implementation tree)

Tree SHA pin: L-001..L-011, L-013..L-016 + controls were observed at `8039ea714c988f4264cd8785405dbb1c410370b6` and re-verified to reproduce at `58c01b4e15f483f25c261ced08e92d7647ea7815` by plan-audit iteration 2; L-012 was re-captured at v0.3.0 (tree `58c01b4e1`, still pre-implementation). The implementation does not exist at any of these SHAs — `internal/hygiene` absent — so every RB go-test cell is red for the stated reason "package not yet implemented".

**L-001 — AC-HYG-001.**
Command: `go test ./internal/hygiene/ -run '^TestRotator_RotatesOverThreshold$' -count=1`
```
# ./internal/hygiene
stat /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1518/internal/hygiene: directory not found
FAIL	./internal/hygiene [setup failed]
FAIL
```
Exit code: 1. RED because the package does not exist yet (correct RED for a new-unit criterion). Green path: M1.

**L-002 — AC-HYG-002.** Command: `go test ./internal/hygiene/ -run '^TestRotator_KeepOneCap$' -count=1`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M1 (crash-recovery arms authored RED-first inside this test).

**L-003 — AC-HYG-003.** Command: `go test ./internal/hygiene/ -run '^TestRotator_UnderThresholdAndAbsent$' -count=1`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M1.

**L-004 — AC-HYG-004.** Command: `go test ./internal/hygiene/ -run '^TestRotator_ConcurrentSerialize$' -count=1 -race`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M1 (stale-decision, fresh-chunk-destruction, and unverified-exclusion arms authored RED-first inside this test).

**L-005 — AC-HYG-005.** Command: `go test ./internal/hygiene/ -run '^TestSinkRegistryCompleteness$' -count=1`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M1.

**L-006 — AC-HYG-006.** Command: `go test ./internal/hygiene/ -run '^TestLivenessVerdictMatrix$' -count=1`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M2.

**L-007 — AC-HYG-007.** Command: `go test ./internal/hygiene/ -run '^TestReportModeByteIdentical$' -count=1`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M3 (over-threshold-self-sink variant inside this test).

**L-008 — AC-HYG-008.** Command: `go test ./internal/hygiene/ -run '^TestApplyModeDeletionSet$' -count=1`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M3.

**L-009 — AC-HYG-009.** Command: `go test ./internal/hygiene/ -run '^TestAuditRowsComplete$' -count=1`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M3.

**L-010 — AC-HYG-010.** Command: `go test ./internal/hygiene/ -run '^TestUnitIndependence$' -count=1`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M3.

**L-011 — AC-HYG-011.** Command: `go test ./internal/hygiene/ -run '^TestSymlinkRefusalParentSwap$' -count=1`. Verbatim stdout: identical block to L-001. Exit code: 1. Green path: M3.

**L-012 — AC-HYG-012** (re-captured at v0.3.0, HEAD `58c01b4e15f483f25c261ced08e92d7647ea7815`, under the renamed test).
Command: `go test ./internal/hygiene/ -run '^TestLockClassExcluded$' -count=1`
```
# ./internal/hygiene
stat /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1518/internal/hygiene: directory not found
FAIL	./internal/hygiene [setup failed]
FAIL
```
Exit code: 1. Green path: M3. (v0.2.0's `TestLockClassNonBlockingAcquire` cell — captured at `8039ea714`, identical shape — is superseded by the D18 exclusion: the acquire mechanism is not built; the exclusion is.)

**L-013 — AC-HYG-013.**
Command: `go test ./internal/hook/ -run '^TestSessionStartHygieneBestEffort$' -count=1`
```
ok  	github.com/modu-ai/moai-adk/internal/hook	0.740s [no tests to run]
```
Exit code: 0 — **empty-sweep RED** (verification-completeness.md §1.1: `no tests to run` is the evidence; the swept set is zero, so this exit 0 is not a pass). Green path: M5 authors the test; the criterion flips only when the same command reports a run test count ≥ 1 and passes.

**L-014 — AC-HYG-014.** Command: `go test ./internal/cli/ -run '^TestCleanHygieneFlags$' -count=1`. Verbatim stdout: `ok  github.com/modu-ai/moai-adk/internal/cli	0.868s [no tests to run]`. Exit code: 0 — empty-sweep RED, same reading as L-013. Green path: M4.

**L-015 — AC-HYG-015.**
Command (pattern widened at v0.4.0 per D35 — adds the reversed multiplication, numeric day-product, and underscore-literal forms; the v0.3.0 RED observation below was taken with the narrower pattern, and the M1 green observation uses this widened form):
`grep -rnE '(10 ?\* ?1024 ?\* ?1024|10485760|10_000_000|<< ?20|[0-9]+ ?\* ?time\.(Hour|Second)|time\.(Hour|Second) ?\* ?[0-9]+|[0-9]+ ?\* ?24|Days = [0-9]+)' internal/hygiene`
```
ugrep: warning: internal/hygiene: No such file or directory
```
Exit code: 2. Red because the target directory does not exist yet — the instrument cannot run, so the criterion is RG and is NOT recorded as a pass today. Green state (pinned now): the same single invocation over the existing package exits **1** with empty output. Green path: M1. Seeded positive control:

**L-C1 — control for L-015** (observed at v0.3.0 with the narrower pattern, 5 hits; to be re-observed at M1 with the widened pattern — the re-seeded control gains a sixth line in a widened form, e.g. `time.Hour * 48`, and must report ≥ 6 hits, exit 0 — the pattern matches real content). Command: seed `/tmp/t1518-seed-threshold.go` with threshold-form lines, then `grep -cE '(10 ?\* ?1024 ?\* ?1024|10485760|10_000_000|<< ?20|[0-9]+ ?\* ?time\.(Hour|Second)|time\.(Hour|Second) ?\* ?[0-9]+|[0-9]+ ?\* ?24|Days = [0-9]+)' /tmp/t1518-seed-threshold.go`.

**L-016 — AC-HYG-016.**
Command: `grep -rnE 'moai-adk-go/\.moai|/Users/|\.\./.*\.moai|os\.Getwd\(\)|os\.UserHomeDir\(\)' internal/hygiene internal/hook/session_start_hygiene_test.go internal/cli/clean_hygiene_test.go`
```
ugrep: warning: internal/hygiene: No such file or directory
ugrep: warning: internal/hook/session_start_hygiene_test.go: No such file or directory
ugrep: warning: internal/cli/clean_hygiene_test.go: No such file or directory
```
Exit code: 2 — RG, not a pass today (explicit real paths, no unexpanded glob; the widened alternation covers the D22 escape forms). Green state (pinned now): the same single invocation over the existing files exits **1** with empty output. Green path: M1/M4/M5. Seeded positive controls:

**L-C2 — absolute control for L-016** (observed, same run): seed `/tmp/t1518-seed-temppath.go` with `root := "/Users/goos/MoAI/moai-adk-go/.moai"`, then `grep -cE 'moai-adk-go/\.moai|/Users/|…' /tmp/t1518-seed-temppath.go` → `1`, exit 0.
**L-C3 — escape-form control for L-016** (authored at v0.3.0, to be observed at M1 before the zero is trusted): seed a file containing `filepath.Join(os.Getwd(), "..", "..", ".moai")` and `os.UserHomeDir()` root derivation; the same grep must report ≥ 2 hits, exit 0 — recorded here as the obligation; the M1 ledger append carries its verbatim output.

## §B Edge Cases

- **Pid reuse** — a registry pid that is alive but belongs to a different process (fingerprint mismatch) is not an affirmative live signal; the session may still be DEAD via the other signals.
- **Live pid without a recorded fingerprint** — the pid signal is unmeasured; the verdict falls to the remaining signals and can never be DEAD on that pid alone.
- **Registry absent / entry missing** — the majority case (61 entries vs 1,116 files): pid+heartbeat unmeasured ⇒ LIVE-or-INDETERMINATE only, never DEAD.
- **Transcript absent under a resolvable root** — unmeasured, never negative: the session may live under a profile root this install does not scan (worktree and secondary-profile sessions are the measured shape).
- **Bulk mtime touch** — a bulk-touched candidate reads LIVE (over-keeping) or content-undatable (spared); neither path deletes.
- **Preserved mtimes (`cp -p`, rsync)** — an old preserved mtime is ignored: dating reads the content timestamp, and mtime is never a deletion datum.
- **Zero-byte sink** — under threshold; rotation no-ops.
- **Absent directories** — an absent `logs/` or state subdir is a state, not an error (record_prune contract).
- **Concurrent hook processes** — two sessions starting at once: the lockfile serializes rotation; two GC runs may race on the same dead file — the loser records `already-gone` (idempotent). A stale pre-lock rotation decision is re-checked under the lock and skipped.
- **Crash mid-rotation** — an orphan `<name>.1.staging` is completed by the next pass under the lock; a staged chunk is never deleted or overwritten (REQ-HYG-002).
- **Windows** — rotation requires the LockFileEx sidecar lock taken and verified; `skipped-locked`/`skipped-platform` otherwise, retried next run; rename-over-open-file sharing violations skip and retry likewise; the lock class is excluded everywhere.
- **Content-undatable bodies** — spared in every mode regardless of mtime (REQ-HYG-009 + the REQ-HYG-005 class table).
- **Symlinked ancestors above the root** — never a violation (macOS `/var` → `/private/var`); refusal scopes strictly below the resolved root; the swap-after-check residual is named and caught on the next pass.
- **Session-key shape** — 36-character hyphenated UUID; a non-matching name is out of scope (spared); the lock class is excluded on its own name form, not by the session-key rule; the shared `backlog.json`/`backlog.db` are on the never-touch negative list regardless.

## §C Quality Gate Criteria

- TRUST 5 Tested: `internal/hygiene` coverage ≥ 85%; every REQ-HYG-001..016 traces to ≥ 1 AC; every RB AC's command re-executed green with `-count=1` (`-race` where concurrency is asserted) before close, with verbatim output appended to `progress.md` §E.2.
- RED discipline: the L-001..L-014 cells are the pre-implementation REDs; each milestone's implementation flips its cells and the flip is evidenced by re-running the same command (exit 0 or run-test-count ≥ 1). The two `[no tests to run]` cells (L-013/L-014) flip only when the run reports actual tests — an exit 0 with an empty swept set never counts. L-C3's escape-form control is observed at M1 before L-016's zero is trusted.
- TRUST 5 Readable/Unified: package godoc states the fail-closed contract; `go vet` + `golangci-lint` (CI version) clean.
- TRUST 5 Secured: closed registries (no path patterns), no deletion outside registry entries, symlink refusal scoped below the resolved root, fail-closed on every unmeasurable signal, mutation gated behind per-surface opt-in, mtime never a deletion datum, lock class excluded.
- TRUST 5 Trackable: conventional commits carrying card t1518.
- Mutant probes recorded for the safety ACs before adoption: AC-HYG-006 (a mutant classifying registry-absent or transcript-absent as DEAD must fail the matrix), AC-HYG-008 (a mutant deleting the mtime-only seed, or partially deleting the goal triple, must fail), AC-HYG-004 (a mutant deciding from a pre-lock stat, or rotating without verified cross-process exclusion, must fail), AC-HYG-002 (a mutant recovering by overwriting the staged chunk must fail), AC-HYG-012 (a mutant evaluating any lock-named path must fail).

## §D Definition of Done

1. All 16 ACs green with verbatim command output recorded in `progress.md` §E.2 — RB cells flipped from their L-001..L-014 REDs; RG cells at their pinned green form (exit 1, zero hits) with the seeded controls L-C1/L-C2/L-C3 standing.
2. REQ→AC traceability complete — the canonical table is spec.md §E; every REQ-HYG-001..016 covered there.
3. The real repository's `.moai` untouched by the entire verification run — verified by a before/after content hash of `.moai/logs` and `.moai/state` taken around the verification batch (a git-status check alone detects nothing on gitignored runtime dirs).
4. `workflow.yaml` carries the `hygiene:` block; `internal/config/defaults.go` carries the six named constants (REQ-HYG-016 list).
5. Self-application asserted: `hygiene-audit.jsonl` present in the sink registry, its rotation apply-mode-only, and its report-mode skip observable (all test-asserted).
6. Milestone flip-map honored: M1 flips L-001..L-005/L-015/L-016(part), M2 flips L-006, M3 flips L-007..L-012, M4 flips L-014, M5 flips L-013.
