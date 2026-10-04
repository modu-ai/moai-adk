# acceptance.md — SPEC-MOAI-HYGIENE-001

All scenarios run inside `t.TempDir()` fixture trees unless a scenario reads this repository's own source text (AC-HYG-005, AC-HYG-015, AC-HYG-016 — read-only greps over `internal/`). No scenario touches the repository's real `.moai` tree (REQ-HYG-015).

**Classification.** Release-blocking (RB): an AC whose failure blocks run-phase exit — every AC that guards mutation safety or a primary behavior of the two units. Regression-guard (RG): an AC whose starting observation is structurally undecidable on the pre-implementation tree (its target files do not exist yet); it pins the green form now and re-executes at its flipping milestone, and is never recorded as a pass before that flip (verification-completeness.md §2.1 undecidable disposition).

**Evidence ledger.** §A table cells cite ledger entries L-001..L-016. Every RB AC carries a four-element RED-now cell (command, verbatim stdout, exit code, tree SHA) observed by direct execution on the pre-implementation tree.

## §A AC Matrix

| AC | Scenario (maps REQ-HYG-…) | Class | Ledger | GREEN then (milestone) |
|---|---|---|---|---|
| AC-HYG-001 | Rotation on over-threshold sink (maps REQ-HYG-001, REQ-HYG-002) | RB | L-001 | test passes, exit 0 (M1) |
| AC-HYG-002 | Keep-1 count cap (maps REQ-HYG-001) | RB | L-002 | test passes, exit 0 (M1) |
| AC-HYG-003 | Under-threshold and absent sinks untouched (maps REQ-HYG-001) | RB | L-003 | test passes, exit 0 (M1) |
| AC-HYG-004 | Concurrent rotators serialize; stale-decision skip; no fresh-chunk destruction (maps REQ-HYG-002) | RB | L-004 | test passes, exit 0 (M1) |
| AC-HYG-005 | Registry completeness guard (maps REQ-HYG-003) | RB | L-005 | guard red on seeded unregistered writer, green on real tree (M1) |
| AC-HYG-006 | Liveness verdict matrix incl. registry-absent / entry-missing / missing-fingerprint rows (maps REQ-HYG-007) | RB | L-006 | test passes, exit 0 (M2) |
| AC-HYG-007 | Report mode non-mutating: byte-identical tree, no lockfile, ≤1 summary row per unit (maps REQ-HYG-004, REQ-HYG-010, REQ-HYG-013) | RB | L-007 | test passes, exit 0 (M3) |
| AC-HYG-008 | Apply mode deletes exactly the eligible set; mtime-only candidates kept (maps REQ-HYG-005, REQ-HYG-008, REQ-HYG-009) | RB | L-008 | test passes, exit 0 (M3) |
| AC-HYG-009 | Audit rows complete, report rows summarized, apply rows per-action (maps REQ-HYG-004, REQ-HYG-009) | RB | L-009 | test passes, exit 0 (M3) |
| AC-HYG-010 | Unit independence (maps REQ-HYG-012) | RB | L-010 | test passes, exit 0 (M3) |
| AC-HYG-011 | Symlink refusal — parent-swap / symlinked-component refusal in both units (maps REQ-HYG-006) | RB | L-011 | test passes, exit 0 (M3) |
| AC-HYG-012 | Lock class — GC-side non-blocking acquire; acquired/blocked/Windows-kept branches (maps REQ-HYG-011) | RB | L-012 | test passes, exit 0 (M3) |
| AC-HYG-013 | SessionStart wiring best-effort, never blocks launch (maps REQ-HYG-014) | RB | L-013 | test exists and passes, exit 0 (M5) |
| AC-HYG-014 | CLI dry-run default; `--apply` overrides config for that invocation; config alone never mutates the CLI (maps REQ-HYG-013) | RB | L-014 | test exists and passes, exit 0 (M4) |
| AC-HYG-015 | Named thresholds only — single-invocation grep, expected green exit 1 (maps REQ-HYG-016) | RG | L-015, L-C1 | exit 1 (zero hits over existing package) (M1) |
| AC-HYG-016 | Tests never touch the real .moai — real-path grep over the test files, expected green exit 1 (maps REQ-HYG-015) | RG | L-016, L-C2 | exit 1 (zero hits over existing files) (M1/M4/M5) |

### Scenario detail

- **AC-HYG-001** — Given a registered sink of 11 MiB in a temp tree, When the rotator runs, Then the sink becomes an empty primary, the 11 MiB of rows live at `<name>.1`, and one apply-mode audit row records the rotation.
- **AC-HYG-002** — Given a registered sink of 11 MiB with a pre-existing `<name>.1`, When the rotator runs, Then the previous chunk is displaced through the staging name and removed only after the fresh chunk is in place, so the tree never holds more than one rotated chunk per sink and a mid-sequence failure never leaves zero chunks where one existed.
- **AC-HYG-003** — Given registered sinks under 10 MiB and one registered name with no file, When the rotator runs, Then byte hashes of the under-threshold sinks are unchanged, the absent name produces no file and no error, and skipped rows carry their reason.
- **AC-HYG-004** — Given one over-threshold sink and N concurrent rotator invocations, plus a second scenario where a rotator enters its critical section carrying a stale over-threshold observation (modeled through the stat seam) while the sink is actually under threshold, When all complete, Then exactly one rotation occurred (one `<name>.1`), the stale-decision rotator performed no action and recorded a `skipped-stale` outcome, the union of primary + chunk bytes contains every pre-run row, and no invocation ever removed a chunk it did not itself displace.
- **AC-HYG-005** — Given the guard scanning `internal/` for append-only sink writers under `.moai/logs/`, When a fixture source tree (in `t.TempDir`) contains an unregistered writer, Then the guard fails naming it; against the real tree every current writer resolves to a registry entry and the guard passes.
- **AC-HYG-006** — Given the signal-combination table, When the evaluator runs each row, Then the verdicts hold as specified. Rows: pid-alive+fingerprint-match (affirmative); pid-alive+fingerprint-mismatch — reused pid (not affirmative); pid-dead (negative); **pid-alive+no-recorded-fingerprint (unmeasured)**; **registry file absent (pid+heartbeat unmeasured)**; **registry present, no entry for the key (pid+heartbeat unmeasured)**; probe-undetermined (unmeasured); transcript-fresh (affirmative); transcript-stale (negative); transcript-root-unresolvable (unmeasured); heartbeat-fresh (affirmative); heartbeat-stale (negative). Any affirmative → LIVE; all-measurable-all-negative → DEAD; any unmeasured with none affirmative → INDETERMINATE carrying the unmeasured signal's name; the registry-absent and entry-missing rows can never produce DEAD.
- **AC-HYG-007** — Given a temp tree seeded with dead-datable, live, indeterminate, lock, and over-threshold-sink candidates, mode `report`, When the GC and rotator run, Then a whole-tree hash taken before and after is identical (no lockfile created, no candidate or sink touched), `.moai/logs/hygiene-audit.jsonl` gained exactly one rotator summary row and one GC summary row, and the report states a keeping reason for every kept candidate.
- **AC-HYG-008** — Given the seed classes: dead+content-datable+aged (deletable), dead+young-datable (kept), **dead+body-undatable with a fresh-looking mtime (kept — `mtime-only`)**, unresolvable-key (kept), LIVE (kept), INDETERMINATE (kept), When mode `apply` runs, Then only the deletable class is gone, every kept class survives, and a re-run reports the already-vanished path as `already-gone` without error.
- **AC-HYG-009** — Given an apply-mode run that deleted at least one candidate and rotated at least one sink, and a report-mode run over the same seed, When the audit sink is read back (fixture path `.moai/logs/hygiene-audit.jsonl` inside the temp tree), Then every apply-mode mutation carries path, unit, mode, decision, reason, and signal evidence (pid verdict, transcript mtime age, heartbeat age, content-date age), every skipped candidate carries its keeping reason, and the report-mode run carries exactly the two summary rows.
- **AC-HYG-010** — Given an injected rotator failure (unwritable lock path) and, in a second scenario, an injected GC failure (read-only target dir), When the combined pass runs, Then in scenario one the GC still completes its report, in scenario two the rotator still completes, and both failures surface as logged per-unit outcomes — never as a blocked pass.
- **AC-HYG-011** — Given a target whose path traverses a symlinked component, and a second scenario where the parent directory is swapped to a symlink between enumeration and action (modeled through the resolution seam), When either unit evaluates the path, Then the outcome is `symlink-refused`, the path is untouched, and the resolved-real-path-inside-`.moai`-root check is what refused it.
- **AC-HYG-012** — Given an aged `spec-close-<SPEC>.lock` with the acquire seam stubbed three ways — acquires (free), `EWOULDBLOCK` (held), and platform-unsupported (Windows) — When apply mode runs, Then acquired removes the file while holding the handle then releases, held keeps it (reason `lock-held`), and Windows keeps it (reason `platform-unsupported`); no fd-probe-then-unlink sequence exists on the path.
- **AC-HYG-013** — Given a SessionStart invocation with the hygiene engine stubbed to fail, When the hook processes the event, Then the hook exits 0, the failure appears only as a logged warning, and the remaining SessionStart steps complete.
- **AC-HYG-014** — Given a temp-rooted project with eligible residue, When `moai clean --audit-logs` and `moai clean --session-state` run without `--apply`, Then the report prints decisions and reasons and the residue survives; with `config mode: apply` but no `--apply` the CLI still does not mutate; with `--apply` the eligible set is gone.
- **AC-HYG-015** — Given the `internal/hygiene/` package source, When the single-invocation threshold grep of L-015 runs over the existing package, Then it exits 1 with empty output (no size or duration literal outside `internal/config/defaults.go`); the seeded control L-C1 proves the pattern matches real content (5 hits) so the zero is measured, not vacuous.
- **AC-HYG-016** — Given the test files of the new units and wiring, When the single-invocation real-path grep of L-016 runs over the existing files, Then it exits 1 with empty output (no repository `.moai` path or absolute user path in any test file); the seeded control L-C2 proves the pattern matches real content (1 hit).

## §A.1 Evidence Ledger (RED-now cells, observed on the pre-implementation tree)

Tree SHA pin for every entry: `8039ea714c988f4264cd8785405dbb1c410370b6` (branch `WT-audit-log-gc`; the implementation does not exist at this SHA — `internal/hygiene` absent — so every RB go-test cell is red for the stated reason "package not yet implemented", the correct pre-implementation RED, and every RG grep cell is red because its target files do not exist yet).

**L-001 — AC-HYG-001.**
Command: `go test ./internal/hygiene/ -run '^TestRotator_RotatesOverThreshold$' -count=1`
```
# ./internal/hygiene
stat /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1518/internal/hygiene: directory not found
FAIL	./internal/hygiene [setup failed]
FAIL
```
Exit code: 1. RED because the package does not exist yet (correct RED for a new-unit criterion). Green path: M1 implements rotator core; the then-clause becomes a passing run, exit 0.

**L-002 — AC-HYG-002.**
Command: `go test ./internal/hygiene/ -run '^TestRotator_KeepOneCap$' -count=1`
Verbatim stdout: identical block to L-001 (`# ./internal/hygiene` / `stat …: directory not found` / `FAIL	./internal/hygiene [setup failed]` / `FAIL`).
Exit code: 1. Green path: M1.

**L-003 — AC-HYG-003.**
Command: `go test ./internal/hygiene/ -run '^TestRotator_UnderThresholdAndAbsent$' -count=1`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M1.

**L-004 — AC-HYG-004.**
Command: `go test ./internal/hygiene/ -run '^TestRotator_ConcurrentSerialize$' -count=1 -race`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M1 (stale-stat sub-case and no-fresh-chunk-destruction assertions are authored RED-first inside this test).

**L-005 — AC-HYG-005.**
Command: `go test ./internal/hygiene/ -run '^TestSinkRegistryCompleteness$' -count=1`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M1 (registry + guard born together; the guard's red-on-seeded-fixture and green-on-real-tree arms are inside this test).

**L-006 — AC-HYG-006.**
Command: `go test ./internal/hygiene/ -run '^TestLivenessVerdictMatrix$' -count=1`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M2. (Matrix rows include registry-absent, entry-missing, and missing-fingerprint — D2.)

**L-007 — AC-HYG-007.**
Command: `go test ./internal/hygiene/ -run '^TestReportModeByteIdentical$' -count=1`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M3. (Adds the no-lockfile and growth-bound then-clauses — D6/D13.)

**L-008 — AC-HYG-008.**
Command: `go test ./internal/hygiene/ -run '^TestApplyModeDeletionSet$' -count=1`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M3. (Seed includes the mtime-only undatable class — D3.)

**L-009 — AC-HYG-009.**
Command: `go test ./internal/hygiene/ -run '^TestAuditRowsComplete$' -count=1`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M3.

**L-010 — AC-HYG-010.**
Command: `go test ./internal/hygiene/ -run '^TestUnitIndependence$' -count=1`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M3.

**L-011 — AC-HYG-011.**
Command: `go test ./internal/hygiene/ -run '^TestSymlinkRefusalParentSwap$' -count=1`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M3 (D5).

**L-012 — AC-HYG-012.**
Command: `go test ./internal/hygiene/ -run '^TestLockClassNonBlockingAcquire$' -count=1`
Verbatim stdout: identical block to L-001.
Exit code: 1. Green path: M3 (D7).

**L-013 — AC-HYG-013.**
Command: `go test ./internal/hook/ -run '^TestSessionStartHygieneBestEffort$' -count=1`
```
ok  	github.com/modu-ai/moai-adk/internal/hook	0.740s [no tests to run]
```
Exit code: 0 — **empty-sweep RED** (verification-completeness.md §1.1: `no tests to run` is the evidence; the swept set is zero, so this exit 0 is not a pass). Red because the wiring test does not exist yet. Green path: M5 authors the test; the criterion flips only when the same command reports a run test count ≥ 1 and passes.

**L-014 — AC-HYG-014.**
Command: `go test ./internal/cli/ -run '^TestCleanHygieneFlags$' -count=1`
```
ok  	github.com/modu-ai/moai-adk/internal/cli	0.868s [no tests to run]
```
Exit code: 0 — empty-sweep RED, same reading as L-013. Green path: M4.

**L-015 — AC-HYG-015.**
Command: `grep -rnE '(10 ?\* ?1024 ?\* ?1024|10485760|<< ?20|[0-9]+ ?\* ?time\.(Hour|Second)|Days = [0-9]+)' internal/hygiene`
```
ugrep: warning: internal/hygiene: No such file or directory
```
Exit code: 2. Red because the target directory does not exist yet — the instrument cannot run, so the criterion is RG and is NOT recorded as a pass today. Green state (pinned now): the same single invocation over the existing package exits **1** with empty output (grep's no-match code — recorded per-cell so the zero is never read off a vanished pipeline exit). Green path: M1. Seeded positive control:

**L-C1 — control for L-015** (observed, same run).
Command: `printf 'maxBytes = 10 * 1024 * 1024\nthreshold = 10485760\nshift = 10 << 20\nwindow = 48 * time.Hour\nminAgeDays = 7\n' > /tmp/t1518-seed-threshold.go` then `grep -cE '(10 ?\* ?1024 ?\* ?1024|10485760|<< ?20|[0-9]+ ?\* ?time\.(Hour|Second)|Days = [0-9]+)' /tmp/t1518-seed-threshold.go`
```
5
```
Exit code: 0 — the pattern matches real content (5 hits); the L-015 zero is measured against a non-vacuous pattern.

**L-016 — AC-HYG-016.**
Command: `grep -rnE 'moai-adk-go/\.moai|/Users/' internal/hygiene internal/hook/session_start_hygiene_test.go internal/cli/clean_hygiene_test.go`
```
ugrep: warning: internal/hygiene: No such file or directory
ugrep: warning: internal/hook/session_start_hygiene_test.go: No such file or directory
ugrep: warning: internal/cli/clean_hygiene_test.go: No such file or directory
```
Exit code: 2 — RG, not a pass today (explicit real paths, no unexpanded glob — D11). Green state (pinned now): the same single invocation over the existing files exits **1** with empty output. Green path: M1/M4/M5 (as each file lands). Seeded positive control:

**L-C2 — control for L-016** (observed, same run).
Command: `printf 'root := "/Users/goos/MoAI/moai-adk-go/.moai"\n' > /tmp/t1518-seed-temppath.go` then `grep -cE 'moai-adk-go/\.moai|/Users/' /tmp/t1518-seed-temppath.go`
```
1
```
Exit code: 0 — the pattern matches real content; the L-016 zero is measured against a non-vacuous pattern.

## §B Edge Cases

- **Pid reuse** — a registry pid that is alive but belongs to a different process (fingerprint mismatch) is not an affirmative live signal; the session may still be DEAD via the other signals.
- **Live pid without a recorded fingerprint** — the pid signal is unmeasured (the reuse defense cannot run); the verdict falls to the remaining signals and can never be DEAD on that pid alone (D13).
- **Registry absent / entry missing** — the majority case (61 entries vs 1,116 files): pid+heartbeat unmeasured ⇒ LIVE-or-INDETERMINATE only, never DEAD (D2).
- **Bulk mtime touch** — a bulk-touched candidate reads LIVE (over-keeping) or content-undatable (spared); neither path deletes.
- **Preserved mtimes (`cp -p`, rsync)** — an old preserved mtime on a young file's body is ignored: dating reads the content timestamp, and mtime is never a deletion datum (D3).
- **Zero-byte sink** — under threshold; rotation no-ops.
- **Absent directories** — an absent `logs/` or state subdir is a state, not an error (record_prune contract).
- **Concurrent hook processes** — two sessions starting at once: the lockfile serializes rotation; two GC runs may race on the same dead file — the loser records `already-gone` (idempotent). A stale pre-lock rotation decision is re-checked under the lock and skipped (REQ-HYG-002).
- **Clock skew** — the minimum-age floor (7 days) exceeds any plausible skew; a fresh-content dead candidate stays inside the young-kept class.
- **Session id reuse** — a reused session key's files belong to the new session; the transcript-fresh signal keeps them (LIVE).
- **Windows** — rotation skips on sharing violation and retries next run; lockfile exclusion is in-process only on Windows (cross-process rotation serialization is best-effort; skip-and-retry is the backstop); the lock class is always kept on Windows (`platform-unsupported`) — `internal/spec/lock.go` semantics recorded in REQ-HYG-011.
- **Content-undatable bodies** — a target file whose body carries no recorded timestamp is spared in every mode regardless of mtime (D3 convergence; the prior mtime-fallback deletion path is removed).
- **Symlinked components** — refused in both units, `symlink-refused` outcome recorded; parent-swap between enumeration and action is covered by the resolution seam (D5).
- **Session-key shape** — 36-character hyphenated UUID as stamped by the writers; a non-matching name has no key and is out of scope (spared) — this covers the legacy per-session residue under `state/todo/` whose names predate the UUID convention only where they do not parse as UUIDs; the shared `backlog.json`/`backlog.db` are on the never-touch negative list regardless (D13).

## §C Quality Gate Criteria

- TRUST 5 Tested: `internal/hygiene` coverage ≥ 85%; every REQ-HYG-001..016 traces to ≥ 1 AC; every RB AC's command re-executed green with `-count=1` (`-race` where concurrency is asserted) before close, with verbatim output appended to `progress.md` §E.2.
- RED discipline: the L-001..L-014 cells are the pre-implementation REDs; each milestone's implementation flips its cells and the flip is evidenced by re-running the same command (exit 0 or run-test-count ≥ 1). The two `[no tests to run]` cells (L-013/L-014) flip only when the run reports actual tests — an exit 0 with an empty swept set never counts.
- TRUST 5 Readable/Unified: package godoc states the fail-closed contract; `go vet` + `golangci-lint` (CI version) clean.
- TRUST 5 Secured: closed registries (no path patterns), no deletion outside registry entries, symlink refusal, fail-closed on every unmeasurable signal, mutation gated behind per-surface opt-in, mtime never a deletion datum.
- TRUST 5 Trackable: conventional commits carrying card t1518.
- Mutant probes recorded for the safety ACs before adoption: AC-HYG-006 (a mutant classifying registry-absent as DEAD must fail the matrix), AC-HYG-008 (a mutant deleting the mtime-only seed must fail), AC-HYG-012 (a mutant probe-then-removing must fail the held branch), AC-HYG-004 (a mutant deciding from a pre-lock stat must fail the stale-decision arm).

## §D Definition of Done

1. All 16 ACs green with verbatim command output recorded in `progress.md` §E.2 — RB cells flipped from their L-001..L-014 REDs; RG cells at their pinned green form (exit 1, zero hits) with the seeded controls L-C1/L-C2 standing.
2. REQ→AC traceability complete — the canonical table is spec.md §E; every REQ-HYG-001..016 covered there.
3. The real repository's `.moai` untouched by the entire verification run (spot-checkable: `git status` clean of `.moai/` mutations on this tree after the run).
4. `workflow.yaml` carries the `hygiene:` block; `internal/config/defaults.go` carries the named constants.
5. Self-application asserted: `hygiene-audit.jsonl` present in the sink registry (test-asserted), and its rotation exempt from the mode gate (test-asserted).
6.plan.md milestone flip-map honored: M1 flips L-001..L-005/L-015/L-016(part), M2 flips L-006, M3 flips L-007..L-012, M4 flips L-014, M5 flips L-013.
