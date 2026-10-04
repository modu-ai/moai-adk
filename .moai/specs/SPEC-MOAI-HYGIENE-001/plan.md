# plan.md — SPEC-MOAI-HYGIENE-001

## §A Context

Card t1518 (v3.2 stage 2, hygiene). The 2026-10-04 hygiene audit measured seven append-only audit sinks with no size bound (largest 14.1 MB, ~1 MB/day peak growth) and nine classes of finished-session state residue (largest: 1,017 dead `context-usage` files, 911 aged `state/verify` scratch files). Existing age-based pruning (`PruneObservationLogs`, REQ-OBH-002) structurally cannot bound an append-only sink, and nothing covers session-keyed state outside `logs/`. The session registry's pid is not a safe liveness source (all 61 entries read "live" at measurement; pid reuse observed; 61 entries vs 1,116 files makes registry-absent/entry-missing the majority case). Development mode: TDD (quality.yaml `constitution.development_mode: tdd`, coverage target 85). Plan-audit iteration 1 returned FAIL 0.73 (D1–D14); this revision folds the full defect list — the design-level fixes are in spec.md v0.2.0 and the MP-8 evidence ledger is acceptance.md §A.1.

## §B Known Issues

- Registry pids go stale and get reused (pid 40177 shared by sessions registered 2026-09-26); 61 entries vs 1,116 files — liveness must never trust the registry pid alone, and registry-absent/entry-missing candidates are explicitly LIVE-or-INDETERMINATE-only (spec REQ-HYG-007).
- mtime can be bulk-touched (hundreds of card dirs shared one mtime on 10-02) and preserved by `cp -p`/rsync — mtime is never a deletion datum; undatable candidates are spared (spec REQ-HYG-009; the v0.1.0 mtime-fallback deletion wording is retired).
- Sink appenders take no lock; the rotator's residual loss window is bounded to the **displaced previous chunk's** inode and opens only after the fresh chunk is in place (same documented class as the harness retention pruner). Must be documented in code, not silently assumed.
- The v0.1.0 remove-then-rename ordering and the pre-lock size decision were rejected by the audit (D4, codex-reproduced): the statted decision is never acted on — size and existence are re-checked inside the lock, and the fresh chunk is staged before the displaced chunk is removed.
- Windows: rename-over-open-file fails (sharing violation) → rotation skips and retries next run; the lockfile is an in-process mutex on Windows only (cross-process rotation serialization best-effort, skip-and-retry backstops it); the O_EXCL lock protocol gives the GC no acquirable handle for an existing spec-close lock, so the lock class is always kept on Windows (`internal/spec/lock.go` semantics).
- Symlink traversal would route a registered-path deletion to an external file (D5, codex sentinel): both units refuse symlinked components and anchor on the resolved real path inside the project `.moai` root.
- Report mode must not become a new unbounded sink (D6): one summary row per unit per run; the hygiene sink's own rotation is mode-gate-exempt.
- `moai clean` CLI and SessionStart hook paths stay best-effort — cleanup never gates a launch (factory record_prune contract).
- The installed `moai` binary (v3.2.0-rc.27) lags this tree; its CLI output is not evidence about this tree's behavior. All verification runs from this tree's own build/tests.

## §C Pre-flight

- [ ] Worktree `WT-audit-log-gc` HEAD re-read immediately before each commit (staleness rule).
- [ ] No dependency on SPEC-AGENTS-CONTRACT-001 (held branch, not in develop).
- [ ] `internal/lockfile`, `internal/session` probe seam (`probeLiveness`), `internal/homestate` fingerprint machinery, `internal/factory` record_prune contract, `internal/spec/lock.go` platform semantics — all confirmed present at `8039ea714` (plan + audit verification).
- [ ] Config key namespace `workflow.hygiene.*` free of collisions (verified: no `hygiene` key in `workflow.yaml` or `internal/config/defaults.go`).
- [ ] MP-8 ledger (acceptance.md §A.1) re-checked at run-phase start: the L-001..L-014 RED cells must still be red (or already flipped by an earlier milestone of this same run) before implementation begins — a flipped cell before its milestone means someone implemented ahead of the plan.

## §D Constraints

- plan phase is read-only on `internal/` — no implementation before run phase.
- All tests: `t.TempDir()` exclusively; never touch the repository's own `.moai` (REQ-HYG-015 [HARD]). No `t.Setenv("OTEL_EXPORTER_*", ...)` in parallel tests (package convention).
- Two independent units: `internal/hygiene` package, `rotator` and `gc` types with no shared mutable state; injected failure in one never blocks the other (AC-HYG-010).
- Closed registries: sink registry and target registry are named, exhaustive constants; nothing outside them is read for mutation decisions. Session key = 36-character hyphenated UUID (the writers' shape); a non-matching name is out of scope.
- **Rotator ordering (D4, binding):** decision (size ≥ threshold) is made **inside** the lock from a re-stat; the primary is renamed to a staging name first; the displaced previous chunk is removed only after the fresh chunk sits at its final `<name>.1` position. No code path removes a chunk the pass did not itself displace. TDD: the stale-decision skip and the fresh-chunk-preservation assertions are authored RED inside `TestRotator_ConcurrentSerialize` before the mechanism exists.
- **mtime (D3, binding):** file mtime is never a deletion datum. Deletion dating reads a timestamp recorded in the file's content; a body without one is spared regardless of mtime. `TestApplyModeDeletionSet` seeds the mtime-only class RED before the rule exists.
- **Symlinks (D5, binding):** both units resolve the fully-resolved real path and refuse symlinked components (`symlink-refused`); the resolution check sits behind a seam; `TestSymlinkRefusalParentSwap` authors the parent-swap RED first.
- **Lock class (D7, binding):** GC-side non-blocking acquire (Unix `flock(2)` LOCK_NB: acquired ⇒ free ⇒ remove while holding ⇒ release; `EWOULDBLOCK` ⇒ kept `lock-held`); Windows: lock class always kept `platform-unsupported`. No fd-probe-then-unlink sequence anywhere.
- **External-process probes** (lsof, if any remain after D7) live behind package-var seams (pattern: `platformProcessCWDs` in `internal/cli/worktree/sweep_cwd_posix.go`); no test spawns an external probe.
- Named thresholds only (`internal/config/defaults.go`): `HygieneAuditLogMaxBytes = 10 * 1024 * 1024`, `HygieneAuditLogKeptRotations = 1`, `HygieneTranscriptActivityWindow = 48 * time.Hour`, `HygieneMinAgeDays = 7`, mode default `report` — each with a `workflow.yaml` override key.
- **Mode precedence (D8, binding):** auto path governed solely by `workflow.hygiene.mode`; CLI mutates only with `--apply` on that invocation (`--apply` overrides the config mode for the invocation; config alone never mutates the CLI). AC-HYG-001/002/004 pin `mode: apply` explicitly.
- **Report-mode granularity (D6, binding):** one summary row per unit per run; per-candidate rows only in apply mode or the CLI; the hygiene sink's own rotation is exempt from the mode gate. Report mode creates no lockfile.
- Hook path: bounded scans (stat-based enumeration; cheap per-candidate probes; one non-blocking lock attempt per lock candidate); every error logged and swallowed; never propagated to the hook return path.

## §E Self-Verification

Run-phase obligations (recorded here, discharged in run phase):

- `go test ./internal/hygiene/... -count=1 -race` green — the new units' suites, each RB test flipped from its acceptance.md §A.1 ledger RED (L-001..L-012).
- `go test ./internal/hook/... ./internal/cli/... -run '^(Hygiene|CleanAuditLogs|CleanSessionState)$' -count=1` green — wiring suites; L-013/L-014 flip **only when the run reports actual tests** — `[no tests to run]` with exit 0 is the empty-sweep red, never a pass (verification-completeness.md §1.1).
- L-015/L-016 (RG) re-executed at their pinned green form — exit 1, empty output, over the now-existing files; seeded controls L-C1/L-C2 stand as the non-vacuous-pattern proof.
- `go vet ./internal/hygiene/...` clean; `golangci-lint run internal/hygiene/...` at the CI version.
- Coverage on `internal/hygiene` ≥ 85% (TRUST 5 Tested).
- Every AC command re-executed once green with verbatim output recorded in `progress.md` §E.2, citing the ledger cell it flips.

## §F Milestones (TDD — RED-first on every new unit; priority-ordered, decision-reversibility-first)

- **M1 (High) — Rotator core + sink registry + completeness guard.** RED first, from the ledger cells: rotation-on-threshold (L-001), keep-1 cap with staging-name displacement order (L-002 — the RED asserts the previous chunk survives a simulated mid-sequence rename failure), under-threshold/absent untouched (L-003), concurrency with the two D4 sub-cases — stale-decision skip (pre-lock stat says over, re-stat under lock says under ⇒ no action, `skipped-stale` row) and no-fresh-chunk-destruction (a second rotator entering after a completed rotation removes nothing) (L-004), registry completeness guard with seeded unregistered writer (L-005), audit rows, `hygiene-audit.jsonl` self-registration. Files: `internal/hygiene/rotate.go`, `internal/hygiene/sinks.go`, `+ tests`. L-015/L-016 flip to their pinned green form as soon as the package and its test files exist (same milestone).
- **M2 (High) — Liveness evaluator.** RED first: full verdict matrix (L-006) including the D2 rows — registry-absent, entry-missing, missing-fingerprint, probe-undetermined — each asserting the verdict can never be DEAD from unmeasured signals; seams for pid probe, fingerprint, transcript scan, heartbeat. Files: `internal/hygiene/liveness.go` + tests.
- **M3 (High) — GC sweep + modes + audit log + lock class + symlink defense.** RED first: closed target enumeration with session-key UUID resolution (L-008 seed classes incl. the mtime-only undatable seed), report mode byte-identical + no lockfile + ≤1 summary row per unit (L-007), apply-mode deletion set (L-008), audit-row completeness incl. the report/apply granularity contract (L-009), unit independence (L-010), symlink refusal + parent-swap via the resolution seam (L-011), lock-class non-blocking acquire with the three stubbed branches (L-012). Files: `internal/hygiene/gc.go`, `internal/hygiene/targets.go`, `internal/hygiene/auditlog.go` + tests.
- **M4 (Medium) — Config + CLI.** Named constants + `workflow.yaml` `hygiene:` block + parser wiring; `moai clean --audit-logs` / `--session-state` with dry-run default, `--apply` per-invocation override, and the config-apply-without-`--apply` no-mutation arm (L-014 flip — the first run of this test with an actual test count). Files: `internal/config/defaults.go` (constants only — smallest possible diff), `internal/config/types.go` (if the section type lives there), `internal/cli/clean.go` (+ `clean_hygiene_test.go`).
- **M5 (Medium) — SessionStart wiring + docs.** Best-effort invocation after existing SessionStart steps, bounded scans, errors swallowed; docs touch-up (`moai clean --help` text). L-013 flips here. Files: `internal/hook/session_start.go` (+ `session_start_hygiene_test.go`), docs.
- **M6 (Low) — Closure.** Full §E self-verification batch, AC command run-through with verbatim outputs into `progress.md` §E.2 (each citing the ledger cell it flips), `@MX` annotations on the new seam vars (`@MX:WARN` class for the documented residual loss window and the Windows lock limits), CHANGELOG entry deferred to sync phase.

Order rationale: M1–M3 carry the decisions most likely to change under review (rotation safety order, verdict semantics, deletion criteria, refusal rules) and land them first; M4–M6 are mechanical wiring that follows once the semantics hold.

## §G Anti-Patterns (shall not)

- Shall not glob or pattern-sweep `.moai/` — every mutation decision names an entry from a closed registry.
- Shall not use file mtime as a deletion datum in any form — dating reads content-recorded timestamps; undatable ⇒ spared.
- Shall not decide a rotation from a pre-lock stat — the size/existence check happens inside the lock, and a stale decision is a no-op skip.
- Shall not remove a rotated chunk before its replacement is staged (no remove-then-rename), and no code path removes a chunk the pass did not itself displace.
- Shall not probe-then-remove a spec-close lock — the GC acquires it non-blockingly or keeps it; on Windows the class is always kept.
- Shall not traverse a symlinked path component in either unit — resolve the real path, require it inside the project `.moai` root, refuse otherwise.
- Shall not let the GC trust the registry pid alone, treat an undetermined probe or a missing fingerprint as measured, or classify registry-absent/entry-missing candidates as DEAD.
- Shall not let report mode create the rotator lockfile, emit per-candidate audit rows, or leave the hygiene sink unbounded in any mode.
- Shall not block or fail a SessionStart launch on any hygiene error.
- Shall not add a config constant at a call site, or a threshold literal in `internal/hygiene/`.
- Shall not point any test at the repository's real `.moai` tree, or mock deletion by writing into the primary checkout.
- Shall not merge the rotator and GC into one failure domain (one try-block around both).

## §H Cross-References

- spec.md §B (REQ-HYG-001..016), §F (design sketch), acceptance.md (AC-HYG-001..016 + §A.1 evidence ledger), decision-index.md (Q1–Q6).
- Plan-audit iteration 1 (FAIL 0.73, `.moai/reports/t1518/plan-audit.md`, audited `8039ea714`) — D1–D14 folded into this revision.
- Code owners mapped in plan phase: `internal/hook/prune_logs.go` + `session_end.go:116`, `internal/hook/trace/writer.go`, `internal/harness/retention.go`, `internal/factory/record_prune.go`, `internal/session/session_pid.go` + `registry.go`, `internal/homestate/process_fingerprint_*.go`, `internal/cli/worktree/sweep_cwd_posix.go`, `internal/hook/inbox_lifecycle.go`, `internal/statusline/context_usage.go`, `internal/goal/state.go`, `internal/cli/codex_stop_chain.go`, `internal/codexwiring/stop_budget.go`, `internal/hook/agent_stop_guard.go`, `internal/harness/routing/types.go`, `internal/verify/store.go`, `internal/spec/lock.go`, `internal/cli/hook_sink.go`, `internal/hook/session_guard.go`, `internal/hook/instructions_loaded.go`, `internal/hook/agent_model_guard.go`, `internal/hook/subagent_write_guard.go`, `internal/codexadapter/diagnostics.go`, `internal/spec/audit_transition.go`.
- Related SPECs: SPEC-OBSERVE-HYGIENE-001, SPEC-WORKTREE-SWEEP-001.
