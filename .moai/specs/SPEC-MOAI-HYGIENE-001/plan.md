# plan.md — SPEC-MOAI-HYGIENE-001

## §A Context

Card t1518 (v3.2 stage 2, hygiene). The 2026-10-04 hygiene audit measured seven append-only audit sinks with no size bound (largest 14.1 MB, ~1 MB/day peak growth) and nine classes of finished-session state residue (largest: 1,017 dead `context-usage` files, 911 aged `state/verify` scratch files). Existing age-based pruning (`PruneObservationLogs`, REQ-OBH-002) structurally cannot bound an append-only sink, and nothing covers session-keyed state outside `logs/`. The session registry's pid is not a safe liveness source (all 61 entries read "live" at measurement; pid reuse observed). Development mode: TDD (quality.yaml `constitution.development_mode: tdd`, coverage target 85).

## §B Known Issues

- Registry pids go stale and get reused (pid 40177 shared by sessions registered 2026-09-26) — liveness must never trust the registry pid alone; the fingerprint seam defeats reuse.
- mtime can be bulk-touched (hundreds of card dirs shared one mtime on 10-02) — mtime is a fallback age signal only; undatable candidates are spared.
- Sink appenders take no lock; the rotator's remove-then-rename carries a one-rename-wide residual loss window on the deleted `<name>.1` inode (same documented class as the harness retention pruner). Accepted for audit tails; must be documented in code, not silently assumed.
- Windows: rename-over-open-file fails (sharing violation) → rotation skips and retries next run; the lockfile is an in-process mutex on Windows only.
- `moai clean` CLI and SessionStart hook paths must stay best-effort — cleanup never gates a launch (factory record_prune contract).
- The installed `moai` binary (v3.2.0-rc.27) lags this tree; its CLI output is not evidence about this tree's behavior. All verification in this SPEC runs from this tree's own build/tests.

## §C Pre-flight

- [ ] Worktree `WT-audit-log-gc` HEAD re-read immediately before each commit (staleness rule).
- [ ] No dependency on SPEC-AGENTS-CONTRACT-001 (held branch, not in develop).
- [ ] `internal/lockfile`, `internal/session` probe seam (`probeLiveness`), `internal/homestate` fingerprint machinery, and `internal/factory` record_prune contract confirmed present at implementation time (all read in plan phase — present at `6643c7bba`).
- [ ] Config key namespace `workflow.hygiene.*` free of collisions (verified: no `hygiene` key exists in `workflow.yaml` or `internal/config/defaults.go`).
- [ ] SPEC ID `SPEC-MOAI-HYGIENE-001` regex PASS + catalogue dedup confirmed (plan phase, 2026-10-05).

## §D Constraints

- plan phase is read-only on `internal/` — no implementation before run phase.
- All tests: `t.TempDir()` exclusively; never touch the repository's own `.moai` (REQ-HYG-018 [HARD]). No `t.Setenv("OTEL_EXPORTER_*", ...)` in parallel tests (package convention).
- Two independent units: `internal/hygiene` package, `rotator` and `gc` types with no shared mutable state; injected failure in one never blocks the other (AC-HYG-010).
- Closed registries: sink registry and target registry are named, exhaustive constants; nothing outside them is read for mutation decisions.
- Named thresholds only (`internal/config/defaults.go`): `HygieneAuditLogMaxBytes = 10 * 1024 * 1024`, `HygieneAuditLogKeptRotations = 1`, `HygieneTranscriptActivityWindow = 48 * time.Hour`, `HygieneMinAgeDays = 7`, mode default `report` — each with a `workflow.yaml` override key.
- External-process probes (lsof) live behind package-var seams (pattern: `platformProcessCWDs` in `internal/cli/worktree/sweep_cwd_posix.go`) so tests inject results; no test spawns lsof.
- Hook path: stat-only scans; at most one lsof invocation per run; every error logged and swallowed; never propagate to the hook return path.

## §E Self-Verification

Run-phase obligations (recorded here, discharged in run phase):

- `go test ./internal/hygiene/... -count=1 -race` green — the new units' suites (RED-first per milestone below).
- `go test ./internal/hook/... ./internal/cli/... -run '^(Hygiene|CleanAuditLogs|CleanSessionState)$' -count=1` green — wiring suites.
- `go vet ./internal/hygiene/...` clean; `golangci-lint run internal/hygiene/...` at the CI version.
- Coverage on `internal/hygiene` ≥ 85% (TRUST 5 Tested).
- AC commands in `acceptance.md` each executed once with verbatim output recorded in `progress.md` §E.2.
- Negative control for AC-HYG-014 observed: the grep command demonstrated non-zero on a deliberately seeded fixture before being trusted as zero on the real tree files (verification-completeness §1.1 — a green with an empty swept set asserts nothing).

## §F Milestones (TDD — RED-first on every new unit; priority-ordered, decision-reversibility-first)

- **M1 (High) — Rotator core + sink registry + completeness guard.** RED first: rotation-on-threshold, keep-1 cap, under-threshold untouched, absent-sink no-op, lockfile mutual exclusion (concurrent rotators, final state = exactly one `<name>.1` with all pre-rotation bytes preserved across sink+chunk), audit rows, registry-completeness guard test (fixture source tree in `t.TempDir` with an unregistered append writer → red naming it). Files: `internal/hygiene/rotate.go`, `internal/hygiene/sinks.go`, `internal/hygiene/rotate_test.go`, `internal/hygiene/sinks_test.go`.
- **M2 (High) — Liveness evaluator.** RED first: the full verdict matrix — pid-alive+fingerprint-match / pid-alive+fingerprint-mismatch (reuse → not affirmative) / pid-dead / probe-undetermined / transcript-fresh / transcript-stale / transcript-root-unresolvable / heartbeat-fresh / heartbeat-stale; compositions → LIVE / DEAD / INDETERMINATE; INDETERMINATE carries the unmeasured signal's reason. Files: `internal/hygiene/liveness.go`, `internal/hygiene/liveness_test.go` (seams for pid probe, fingerprint, transcript scan, heartbeat).
- **M3 (High) — GC sweep + modes + audit log + lock class.** RED first: closed target enumeration (each target class seeded in `t.TempDir`), report mode byte-identical (tree hash before/after, audit sink excluded from the hash), apply-mode deletion set (dead+datable+aged only; young-dead, undatable, unresolvable-key, LIVE, INDETERMINATE all kept), idempotent `already-gone`, audit-row completeness (path/reason/signals/mode), independence (injected rotator failure → GC still runs and vice versa), lock-class aged+fd-free removable / aged+fd-held kept / probe-fails kept. Files: `internal/hygiene/gc.go`, `internal/hygiene/targets.go`, `internal/hygiene/auditlog.go`, `internal/hygiene/gc_test.go`.
- **M4 (Medium) — Config + CLI.** Named constants + `workflow.yaml` `hygiene:` block + parser wiring; `moai clean --audit-logs` / `--session-state` with dry-run default and `--apply` gate (cobra tests in `t.TempDir`-rooted projects). Files: `internal/config/defaults.go` (constants only — smallest possible diff), `internal/config/types.go` (if the section type lives there), `internal/cli/clean.go` (+ `clean_hygiene_test.go`).
- **M5 (Medium) — SessionStart wiring + docs.** Best-effort invocation after existing SessionStart steps (after the registry protocol, before output assembly), sub-budget, errors swallowed; `hygiene-audit.jsonl` registered as a rotation sink (self-application, asserted by a test that the registry contains it); docs touch-up (`moai clean --help` text). Files: `internal/hook/session_start.go` (+ `session_start_hygiene_test.go`), `internal/hygiene/sinks.go` (registry entry), docs.
- **M6 (Low) — Closure.** Full §E self-verification batch, AC command run-through with verbatim outputs into `progress.md` §E.2, `@MX` annotations on the new seam vars (`@MX:WARN` class for the documented residual loss window), CHANGELOG entry deferred to sync phase.

Order rationale: M1–M3 carry the decisions most likely to change under review (verdict semantics, registry membership, deletion criteria) and land them first; M4–M6 are mechanical wiring that follows once the semantics hold.

## §G Anti-Patterns (shall not)

- Shall not glob or pattern-sweep `.moai/` — every mutation decision names an entry from a closed registry.
- Shall not use file mtime as the sole age signal, or delete an undatable candidate.
- Shall not let the GC trust the registry pid alone, or treat an undetermined probe as dead.
- Shall not block or fail a SessionStart launch on any hygiene error.
- Shall not add a config constant at a call site, or a threshold literal in `internal/hygiene/`.
- Shall not point any test at the repository's real `.moai` tree, or mock deletion by writing into the primary checkout.
- Shall not merge the rotator and GC into one failure domain (one try-block around both).

## §H Cross-References

- spec.md §B (REQ-HYG-001..019), §F (design sketch), acceptance.md (AC-HYG-001..015), decision-index.md (Q1–Q6).
- Code owners mapped in plan phase: `internal/hook/prune_logs.go` + `session_end.go:116`, `internal/hook/trace/writer.go`, `internal/harness/retention.go`, `internal/factory/record_prune.go`, `internal/session/session_pid.go` + `registry.go`, `internal/homestate/process_fingerprint_*.go`, `internal/cli/worktree/sweep_cwd_posix.go`, `internal/hook/inbox_lifecycle.go`, `internal/statusline/context_usage.go`, `internal/goal/state.go`, `internal/cli/codex_stop_chain.go`, `internal/codexwiring/stop_budget.go`, `internal/hook/agent_stop_guard.go`, `internal/harness/routing/types.go`, `internal/verify/store.go`, `internal/spec/lock.go`, `internal/cli/hook_sink.go`, `internal/hook/session_guard.go`, `internal/hook/instructions_loaded.go`, `internal/hook/agent_model_guard.go`, `internal/hook/subagent_write_guard.go`, `internal/codexadapter/diagnostics.go`, `internal/spec/audit_transition.go`.
- Related SPECs: SPEC-OBSERVE-HYGIENE-001, SPEC-WORKTREE-SWEEP-001.
