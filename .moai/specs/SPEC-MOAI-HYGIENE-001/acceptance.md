# acceptance.md — SPEC-MOAI-HYGIENE-001

All scenarios run inside `t.TempDir()` fixture trees unless a scenario reads this repository's own source text (AC-HYG-005, AC-HYG-011, AC-HYG-014 — read-only greps over `internal/`). No scenario touches the repository's real `.moai` tree (REQ-HYG-018).

## §A AC Matrix

### AC-HYG-001 — Rotation on over-threshold sink (maps REQ-HYG-001, REQ-HYG-002)
- **Given** a registered sink of 11 MiB in a temp tree
- **When** the rotator runs
- **Then** the sink becomes an empty primary, the 11 MiB of rows live at `<name>.1`, and one audit row records the rotation
- Command: `go test ./internal/hygiene/ -run '^TestRotator_RotatesOverThreshold$' -count=1`

### AC-HYG-002 — Keep-1 count cap (maps REQ-HYG-001)
- **Given** a registered sink of 11 MiB with a pre-existing `<name>.1`
- **When** the rotator runs
- **Then** the previous `<name>.1` is removed before the rename, so the tree never holds more than one rotated chunk per sink
- Command: `go test ./internal/hygiene/ -run '^TestRotator_KeepOneCap$' -count=1`

### AC-HYG-003 — Under-threshold and absent sinks untouched (maps REQ-HYG-003)
- **Given** registered sinks under 10 MiB and one registered name with no file
- **When** the rotator runs
- **Then** byte hashes of the under-threshold sinks are unchanged, the absent name produces no file and no error, and skipped rows carry their reason
- Command: `go test ./internal/hygiene/ -run '^TestRotator_UnderThresholdAndAbsent$' -count=1`

### AC-HYG-004 — Concurrent rotators serialize; no data loss (maps REQ-HYG-002)
- **Given** one over-threshold sink and N concurrent rotator invocations
- **When** all complete
- **Then** exactly one rotation occurred (one `<name>.1`), and the union of primary + chunk bytes contains every pre-run row
- Command: `go test ./internal/hygiene/ -run '^TestRotator_ConcurrentSerialize$' -count=1 -race`

### AC-HYG-005 — Registry completeness guard (maps REQ-HYG-004)
- **Given** the guard test scanning `internal/` for append-only sink writers under `.moai/logs/`
- **When** a fixture source tree (in `t.TempDir`) contains an unregistered writer, the guard names it and fails; against the real tree, every current writer resolves to a registry entry and the guard passes
- **Then** adding a ninth sink without registering it turns the guard red on the next run
- Command: `go test ./internal/hygiene/ -run '^TestSinkRegistryCompleteness$' -count=1`

### AC-HYG-006 — Liveness verdict matrix (maps REQ-HYG-007, REQ-HYG-008)
- **Given** the table of signal combinations: pid-alive+fingerprint-match; pid-alive+fingerprint-mismatch (reused pid — not affirmative); pid-measured-dead; probe-undetermined; transcript-fresh; transcript-stale; transcript-root-unresolvable; heartbeat-fresh; heartbeat-stale
- **When** the evaluator runs each row
- **Then** any affirmative row-set yields LIVE; all-measurable-all-negative yields DEAD; any unmeasured with none affirmative yields INDETERMINATE carrying the unmeasured signal's name; no row reaches DEAD via the pid signal alone
- Command: `go test ./internal/hygiene/ -run '^TestLivenessVerdictMatrix$' -count=1`

### AC-HYG-007 — Report mode is non-mutating (maps REQ-HYG-006, REQ-HYG-012, REQ-HYG-015)
- **Given** a temp tree seeded with dead-datable, live, indeterminate, and lock candidates, mode `report`
- **When** the GC and rotator run
- **Then** a whole-tree hash taken before and after is identical except for appended rows in `.moai/logs/hygiene-audit.jsonl`, and the run report states a keeping reason for every kept candidate
- Command: `go test ./internal/hygiene/ -run '^TestReportModeByteIdentical$' -count=1`

### AC-HYG-008 — Apply mode deletes exactly the eligible set (maps REQ-HYG-006, REQ-HYG-009, REQ-HYG-010, REQ-HYG-011)
- **Given** the same seed classed: dead+datable+aged (deletable), dead+young (kept), dead+undatable (kept), unresolvable-key (kept), LIVE (kept), INDETERMINATE (kept)
- **When** mode `apply` runs
- **Then** only the deletable class is gone, every kept class survives, and a re-run reports the already-vanished path as `already-gone` without error
- Command: `go test ./internal/hygiene/ -run '^TestApplyModeDeletionSet$' -count=1`

### AC-HYG-009 — Audit rows are complete (maps REQ-HYG-005, REQ-HYG-010)
- **Given** an apply-mode run that deleted at least one candidate and rotated at least one sink
- **When** `.mologs/hygiene-audit.jsonl` is read back (fixture path)
- **Then** every mutated path has a row carrying path, unit, mode, decision, reason, and the signal evidence (pid verdict, transcript mtime age, heartbeat age); every skipped candidate has a row with its keeping reason
- Command: `go test ./internal/hygiene/ -run '^TestAuditRowsComplete$' -count=1`

### AC-HYG-010 — Unit independence (maps REQ-HYG-014)
- **Given** an injected rotator failure (unwritable lock path) and, in a second scenario, an injected GC failure (read-only target dir)
- **When** the combined pass runs
- **Then** in scenario one the GC still completes its report, in scenario two the rotator still completes, and both failures surface as logged per-unit outcomes — never as a blocked pass
- Command: `go test ./internal/hygiene/ -run '^TestUnitIndependence$' -count=1`

### AC-HYG-011 — Named thresholds only (maps REQ-HYG-019)
- **Given** the `internal/hygiene/` package source
- **When** grepped for raw size/age literals
- **Then** no byte-size or day-count literal appears outside `internal/config/defaults.go`, and the constants carry the documented values (10 MiB, keep 1, 48 h, 7 days, mode report)
- Command: `grep -rnE '[0-9]+ ?\* ?1024 ?\* ?1024|[0-9]+ ?\* ?time\.(Hour|Day)|Days = [0-9]+' internal/hygiene/ --include='*.go' | grep -v _test.go | wc -l` (expect `0`)

### AC-HYG-012 — SessionStart wiring is best-effort (maps REQ-HYG-016)
- **Given** a SessionStart invocation with the hygiene engine stubbed to fail
- **When** the hook processes the event
- **Then** the hook exits 0, the failure appears only as a logged warning, and the remaining SessionStart steps complete
- Command: `go test ./internal/hook/ -run '^TestSessionStartHygieneBestEffort$' -count=1`

### AC-HYG-013 — CLI dry-run default, --apply gate (maps REQ-HYG-015, REQ-HYG-017)
- **Given** a temp-rooted project with eligible residue
- **When** `moai clean --audit-logs` and `moai clean --session-state` run without `--apply`
- **Then** the report prints decisions and reasons, the residue survives; with `--apply` the eligible set is gone
- Command: `go test ./internal/cli/ -run '^TestCleanHygieneFlags$' -count=1`

### AC-HYG-014 — Tests never touch the real .moai (maps REQ-HYG-018) [HARD]
- **Given** the test files of the new units and wiring
- **When** grepped for the repository's own `.moai` path or any absolute user path
- **Then** zero matches — every fixture root derives from `t.TempDir()`
- Command: `grep -rnE 'moai-adk-go/\.moai|/Users/' internal/hygiene/*_test.go internal/hook/session_start_hygiene_test.go internal/cli/clean_hygiene_test.go | wc -l` (expect `0`)

### AC-HYG-015 — Lock class is fail-closed (maps REQ-HYG-013)
- **Given** an aged `spec-close-<SPEC>.lock` with the fd-probe seam stubbed three ways: reports no holder, reports a holder, returns an error
- **When** apply mode runs
- **Then** no-holder deletes it, holder keeps it (reason `fd-held`), error keeps it (reason `probe-unavailable`)
- Command: `go test ./internal/hygiene/ -run '^TestLockClassFailClosed$' -count=1`

## §B Edge Cases

- **Pid reuse** — a registry pid that is alive but belongs to a different process (fingerprint mismatch) is not an affirmative live signal; the session may still be DEAD via the other signals.
- **Bulk mtime touch** — a bulk-touched candidate reads LIVE (over-keeping) or undatable (spared); neither path deletes.
- **Zero-byte sink** — under threshold; rotation no-ops (matches `PruneObservationLogs`' zero-byte handling for trace files, unchanged).
- **Absent directories** — an absent `logs/` or state subdir is a state, not an error (record_prune contract).
- **Concurrent hook processes** — two sessions starting at once: the lockfile serializes rotation; two GC runs may race on the same dead file — the loser records `already-gone` (idempotent).
- **Clock skew** — the minimum-age floor (7 days) exceeds any plausible skew; a fresh-mtime dead candidate stays inside the young-kept class.
- **Session id reuse** — a reused session id's files belong to the new session; the transcript-fresh signal keeps them (LIVE).
- **Windows** — rotation skips on sharing violation and retries next run; lockfile exclusion is in-process only on Windows (documented residual, same as harness retention).
- **Undatable record bodies** — a target file whose body carries no timestamp and whose mtime is the only datum still deletes in apply mode (mtime fallback is legitimate for session-keyed files; the *bulk-touch* shape is what the datable rule guards, and a bulk-touched mtime reads recent → young-kept). Files whose mtime is unavailable entirely are spared.
- **Registry absent** — no registry file means no pid/heartbeat signals; candidates then decide on the transcript signal alone, and an unresolvable transcript root makes everything INDETERMINATE (kept).

## §C Quality Gate Criteria

- TRUST 5 Tested: `internal/hygiene` coverage ≥ 85%; every REQ-HYG-* traces to ≥ 1 AC; every AC command runs green with `-count=1` (and `-race` where concurrency is asserted).
- TRUST 5 Readable/Unified: package godoc states the fail-closed contract; `go vet` + `golangci-lint` (CI version) clean.
- TRUST 5 Secured: closed registries (no path patterns), no deletion outside registry entries, fail-closed on every unmeasurable signal, mutation gated behind explicit opt-in.
- TRUST 5 Trackable: conventional commits carrying card t1518.
- Negative controls observed before trusting zero-results: AC-HYG-005's guard demonstrated red on a seeded fixture; AC-HYG-011/014 grep patterns demonstrated non-zero on a seeded file before being trusted as zero (verification-completeness §1.1).

## §D Definition of Done

1. All 15 ACs green with verbatim command output recorded in `progress.md` §E.2.
2. REQ→AC traceability complete — the canonical table is spec.md §E; every REQ-HYG-001..019 covered there.
3. The real repository's `.moai` untouched by the entire verification run (spot-checkable: `git status` clean of `.moai/` mutations on this tree after the run).
4. `workflow.yaml` carries the `hygiene:` block; `internal/config/defaults.go` carries the five named constants.
5. Self-application asserted: `hygiene-audit.jsonl` present in the sink registry (test-asserted).
