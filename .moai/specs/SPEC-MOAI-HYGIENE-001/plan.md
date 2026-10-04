# plan.md — SPEC-MOAI-HYGIENE-001

## §A Context

Card t1518 (v3.2 stage 2, hygiene). The 2026-10-04 hygiene audit measured seven append-only audit sinks with no size bound (largest 14.1 MB, ~1 MB/day peak growth) and nine classes of finished-session state residue (largest: 1,017 dead `context-usage` files, 911 aged `state/verify` scratch files). Existing age-based pruning (`PruneObservationLogs`, REQ-OBH-002) structurally cannot bound an append-only sink, and nothing covers session-keyed state outside `logs/`. The session registry's pid is not a safe liveness source (all 61 entries read "live" at measurement; pid reuse observed; 61 entries vs 1,116 files makes registry-absent/entry-missing the majority case). Development mode: TDD (quality.yaml `constitution.development_mode: tdd`, coverage target 85). Plan-audit iteration 1 returned FAIL 0.73 (D1–D14); this revision folds the full defect list — the design-level fixes are in spec.md v0.2.0 and the MP-8 evidence ledger is acceptance.md §A.1.

## §B Known Issues

- Registry pids go stale and get reused (pid 40177 shared by sessions registered 2026-09-26); 61 entries vs 1,116 files — liveness must never trust the registry pid alone, and registry-absent/entry-missing candidates are explicitly LIVE-or-INDETERMINATE-only (spec REQ-HYG-007).
- mtime can be bulk-touched (hundreds of card dirs shared one mtime on 10-02) and preserved by `cp -p`/rsync — mtime is never a deletion datum; undatable candidates are spared (spec REQ-HYG-009; the v0.1.0 mtime-fallback deletion wording is retired).
- Sink appenders take no lock; the rotator's residual loss window is bounded to the **displaced previous chunk's** inode and opens only after the fresh chunk is in place (same documented class as the harness retention pruner). Must be documented in code, not silently assumed.
- The v0.1.0 remove-then-rename ordering and the pre-lock size decision were rejected by the audit (D4, codex-reproduced): the statted decision is never acted on — size and existence are re-checked inside the lock, and the chunk replacement runs through the staging sequence (`<name>.1.staging`, promoted to `<name>.1` only while the replacement is staged).
- Iteration 2 added the crash-recovery half (D16): a crashed pass strands an orphan `<name>.1.staging` — the next pass completes the interrupted placement under the lock; a staged chunk is never deleted or overwritten, and both chunk artifact names are rotator-owned registry entries.
- Windows (D17): the in-process mutex is not cross-process — the codex experiment destroyed a fresh chunk through a stale cross-process re-stat. Rotation on Windows proceeds only under a taken-and-verified LockFileEx sidecar lock; otherwise `skipped-locked`/`skipped-platform` + retry. The O_EXCL lock protocol gives no acquirable handle for an existing spec-close lock.
- Symlink traversal would route a registered-path deletion to an external file (D5/D19, codex sentinel + the darwin `/var` → `/private/var` vacuity): refusal scopes strictly below the fully-resolved `.moai` root — the root's own symlinked ancestors never count — with the swap-after-check residual named and caught on the next pass.
- Report mode must not become a new unbounded sink (D6) and must not collide with the four rotation requirements (D15): the resolution is one decision — **all rotation is apply-mode-only**; report mode writes summary rows (row-rate growth, disclosed) and records `skipped-report-mode` for an over-threshold hygiene sink.
- Lock class (D18): empty-file dating is impossible without a writer change, and acquire-then-unlink re-admits a two-live-locks window (codex: two exclusive holders concurrently) — the class is excluded from GC scope; reclamation is upstream (decision-index Q7).
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
- All tests: `t.TempDir()` exclusively; never touch the repository's own `.moai` (REQ-HYG-015 [HARD]). Runtime guard (D22): when `testing.Testing()` is true, the units' entry points refuse any root outside the test's temporary directory. No `t.Setenv("OTEL_EXPORTER_*", ...)` in parallel tests (package convention).
- Two independent units: `internal/hygiene` package, `rotator` and `gc` types with no shared mutable state; injected failure in one never blocks the other (AC-HYG-010).
- Closed registries: sink registry and target registry are named, exhaustive constants; nothing outside them is read for mutation decisions. Session key = 36-character hyphenated UUID (the writers' shape); a non-matching name is out of scope.
- **Rotator ordering (D4+D16, binding):** decision (size ≥ threshold) is made **inside** the lock from a re-stat; the chunk replacement runs primary → `<name>.1.staging` → remove existing `<name>.1` (only while a staged replacement exists) → promote staging. Pass start recovers an orphan staging under the lock by completing the placement — a staged chunk is never deleted or overwritten, and `<name>.1`/`<name>.1.staging` are rotator-owned registry artifacts (REQ-HYG-003). TDD: the stale-decision skip, fresh-chunk-preservation, crash-recovery, and unverified-exclusion arms are authored RED before the mechanism exists.
- **mtime (D3, binding):** file mtime is never a deletion datum. Deletion dating reads a timestamp recorded in the file's content; a body without one is spared regardless of mtime. `TestApplyModeDeletionSet` seeds the mtime-only class RED before the rule exists.
- **Symlinks (D5+D19, binding):** both units resolve the project root fully (symlinked ancestors of the root never count — the darwin `/var` shape) and refuse any component strictly below the resolved root (`symlink-refused`); the check is re-evaluated immediately before each action and the final-window residual is named; `TestSymlinkRefusalParentSwap` authors the parent-swap and swap-after-check arms RED first.
- **Lock class (D7→D18, binding):** excluded from GC scope entirely — no evaluate, probe, acquire, or delete path exists; lock-named scan hits are reported `lock-class-excluded` (`TestLockClassExcluded`). Rationale lives in REQ-HYG-011; upstream reclamation is parked at decision-index Q7.
- **External-process probes** (lsof, if any remain after D7) live behind package-var seams (pattern: `platformProcessCWDs` in `internal/cli/worktree/sweep_cwd_posix.go`); no test spawns an external probe.
- Named thresholds only (`internal/config/defaults.go`): `HygieneAuditLogMaxBytes = 10 * 1024 * 1024`, `HygieneAuditLogKeptRotations = 1`, `HygieneTranscriptActivityWindow = 48 * time.Hour`, `HygieneHeartbeatStaleWindow = 24 * time.Hour`, `HygieneMinAgeDays = 7`, mode default `report` — each with a `workflow.yaml` override key.
- **Mode precedence (D8, binding):** auto path governed solely by `workflow.hygiene.mode`; CLI mutates only with `--apply` on that invocation (`--apply` overrides the config mode for the invocation; config alone never mutates the CLI). AC-HYG-001/002/004 pin `mode: apply` explicitly.
- **Rotation apply-mode-only (D15, binding — one decision, recorded in REQ-HYG-004):** every rotation, the hygiene sink's own included, happens only in apply mode; report mode records `skipped-report-mode` for an over-threshold hygiene sink and never creates a lockfile. No exemption clauses anywhere.
- **Report-mode granularity (D6, binding):** one summary row per unit per run; per-candidate rows only in apply mode or the CLI; lock-class acquire attempts are apply-mode-only (report mode records `would-probe`).
- Hook path: bounded scans (stat-based enumeration; cheap per-candidate probes; one non-blocking lock attempt per lock candidate); every error logged and swallowed; never propagated to the hook return path.

## §E Self-Verification

Run-phase obligations (recorded here, discharged in run phase):

- `go test ./internal/hygiene/... -count=1 -race` green — the new units' suites, each RB test flipped from its acceptance.md §A.1 ledger RED (L-001..L-012).
- `go test ./internal/hook/... ./internal/cli/... -run '^(Hygiene|CleanAuditLogs|CleanSessionState)$' -count=1` green — wiring suites; L-013/L-014 flip **only when the run reports actual tests** — `[no tests to run]` with exit 0 is the empty-sweep red, never a pass (verification-completeness.md §1.1).
- L-015/L-016 (RG) re-executed at their pinned green form — exit 1, empty output, over the now-existing files; seeded controls L-C1/L-C2 stand as the non-vacuous-pattern proof.
- `go vet ./internal/hygiene/...` clean; `golangci-lint run internal/hygiene/...` at the CI version.
- Coverage on `internal/hygiene` ≥ 85% (TRUST 5 Tested).
- Every AC command re-executed once green with verbatim output recorded in `progress.md` §E.2, citing the ledger cell it flips.
- Isolation verification (D22): a before/after content hash of `.moai/logs` and `.moai/state` taken around the verification batch — a git-status check alone detects nothing on gitignored runtime dirs. L-C3's escape-form control is observed before L-016's zero is trusted.

## §F Milestones (TDD — RED-first on every new unit; priority-ordered, decision-reversibility-first)

- **M1 (High) — Rotator core + sink registry + completeness guard.** RED first, from the ledger cells: rotation-on-threshold (L-001), keep-1 cap with the staged, crash-recoverable displacement order (L-002 — RED arms: a simulated crash after staging strands an orphan `<name>.1.staging` that the next pass completes without overwriting or deleting the staged chunk; a mid-sequence rename failure never leaves zero chunks where one existed), under-threshold/absent untouched (L-003), concurrency with the D4+D17 sub-cases — stale-decision skip (`skipped-stale`), no-fresh-chunk-destruction, and the unverified-exclusion skip (sidecar seam reports held/unverifiable ⇒ `skipped-locked`/`skipped-platform`, nothing destroyed) (L-004), registry completeness guard with seeded unregistered writer (L-005), audit rows, `hygiene-audit.jsonl` self-registration. Files: `internal/hygiene/rotate.go`, `internal/hygiene/sinks.go`, `+ tests`. L-015/L-016 flip to their pinned green form as soon as the package and its test files exist (same milestone); L-C3's escape-form control is observed in this milestone.
- **M2 (High) — Liveness evaluator.** RED first: full verdict matrix (L-006) including the D2 rows — registry-absent, entry-missing, missing-fingerprint, probe-undetermined — and the D20 rows — transcript-absent-under-resolvable-root (unmeasured, never DEAD-feeding) and both heartbeat-window sides against `HygieneHeartbeatStaleWindow` — each asserting the verdict can never be DEAD from unmeasured signals; seams for pid probe, fingerprint, transcript scan, heartbeat. Files: `internal/hygiene/liveness.go` + tests.
- **M3 (High) — GC sweep + modes + audit log + symlink defense.** RED first: closed target enumeration with session-key UUID resolution and the REQ-HYG-005 per-class dating table (each class's timestamp field pinned by a RED-first assertion; verify dated per-entry entry-by-entry; goal triple grouped), report mode byte-identical + no lockfile + ≤1 summary row per unit + the over-threshold-self-sink `skipped-report-mode` variant + `would-probe` (L-007), apply-mode deletion set incl. the mtime-only undatable seed (L-008), audit-row completeness incl. the report/apply granularity contract (L-009), unit independence (L-010), symlink refusal scoped below the resolved root + parent-swap and swap-after-check arms (L-011), lock-class exclusion — no evaluate/probe/acquire/delete path and a registry structurally free of lock entries (L-012). Files: `internal/hygiene/gc.go`, `internal/hygiene/targets.go`, `internal/hygiene/auditlog.go` + tests.
- **M4 (Medium) — Config + CLI.** Named constants + `workflow.yaml` `hygiene:` block + parser wiring; `moai clean --audit-logs` / `--session-state` with dry-run default, `--apply` per-invocation override, and the config-apply-without-`--apply` no-mutation arm (L-014 flip — the first run of this test with an actual test count). Files: `internal/config/defaults.go` (constants only — smallest possible diff), `internal/config/types.go` (if the section type lives there), `internal/cli/clean.go` (+ `clean_hygiene_test.go`).
- **M5 (Medium) — SessionStart wiring + docs.** Best-effort invocation after existing SessionStart steps, bounded scans, errors swallowed; docs touch-up (`moai clean --help` text). L-013 flips here. Files: `internal/hook/session_start.go` (+ `session_start_hygiene_test.go`), docs.
- **M6 (Low) — Closure.** Full §E self-verification batch, AC command run-through with verbatim outputs into `progress.md` §E.2 (each citing the ledger cell it flips), `@MX` annotations on the new seam vars (`@MX:WARN` class for the documented residual loss window and the Windows lock limits), CHANGELOG entry deferred to sync phase.

Order rationale: M1–M3 carry the decisions most likely to change under review (rotation safety order, verdict semantics, deletion criteria, refusal rules) and land them first; M4–M6 are mechanical wiring that follows once the semantics hold.

## §G Anti-Patterns (shall not)

- Shall not glob or pattern-sweep `.moai/` — every mutation decision names an entry from a closed registry.
- Shall not use file mtime as a deletion datum in any form — dating reads content-recorded timestamps; undatable ⇒ spared.
- Shall not decide a rotation from a pre-lock stat — the size/existence check happens inside the lock, and a stale decision is a no-op skip.
- Shall not remove a rotated chunk before its replacement is staged, shall not overwrite or delete an orphan `<name>.1.staging` (recovery completes it), and no code path removes a chunk the pass did not itself displace.
- Shall not rotate on a path whose cross-process exclusion is unverified — Windows rotates only under a taken-and-verified LockFileEx sidecar lock, else skips and retries.
- Shall not evaluate, probe, acquire, or delete a spec-close lock — the class is excluded (`lock-class-excluded`).
- Shall not refuse a path for symlinked ancestors above the fully-resolved root, and shall not act below the root without the immediately-before-action re-check.
- Shall not let the GC trust the registry pid alone, treat an undetermined probe or a missing fingerprint as measured, or classify registry-absent/entry-missing candidates as DEAD.
- Shall not let report mode create any lockfile, emit per-candidate audit rows, or rotate anything — the hygiene sink included.
- Shall not block or fail a SessionStart launch on any hygiene error.
- Shall not add a config constant at a call site, or a threshold literal in `internal/hygiene/`.
- Shall not point any test at the repository's real `.moai` tree, or mock deletion by writing into the primary checkout.
- Shall not merge the rotator and GC into one failure domain (one try-block around both).

## §H Cross-References

- spec.md §B (REQ-HYG-001..016), §F (design sketch), acceptance.md (AC-HYG-001..016 + §A.1 evidence ledger), decision-index.md (Q1–Q6).
- Plan-audit iteration 1 (FAIL 0.73, `.moai/reports/t1518/plan-audit.md`, audited `8039ea714`) — D1–D14 folded into this revision.
- Code owners mapped in plan phase: `internal/hook/prune_logs.go` + `session_end.go:116`, `internal/hook/trace/writer.go`, `internal/harness/retention.go`, `internal/factory/record_prune.go`, `internal/session/session_pid.go` + `registry.go`, `internal/homestate/process_fingerprint_*.go`, `internal/cli/worktree/sweep_cwd_posix.go`, `internal/hook/inbox_lifecycle.go`, `internal/statusline/context_usage.go`, `internal/goal/state.go`, `internal/cli/codex_stop_chain.go`, `internal/codexwiring/stop_budget.go`, `internal/hook/agent_stop_guard.go`, `internal/harness/routing/types.go`, `internal/verify/store.go`, `internal/spec/lock.go`, `internal/cli/hook_sink.go`, `internal/hook/session_guard.go`, `internal/hook/instructions_loaded.go`, `internal/hook/agent_model_guard.go`, `internal/hook/subagent_write_guard.go`, `internal/codexadapter/diagnostics.go`, `internal/spec/audit_transition.go`.
- Related SPECs: SPEC-OBSERVE-HYGIENE-001, SPEC-WORKTREE-SWEEP-001.
