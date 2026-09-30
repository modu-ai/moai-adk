# SPEC-CODEX-LANE-SLOTS-001 — Implementation Plan

## §A — Context

Card t1378. `moai codex -f lane` cannot attach to an active run whose only free numbering lies beyond the launcher's join bound, because the codex twin claims within `1..DefaultFactoryLeaderLanes` (1) while the automatic scan never looks past `maxSlots` even though `maxCanonical` (one past the highest live claim) is already computed a few lines above. The same card carries the leader's scope addition: explain the rc.23 immediate-error → long-launch behavior change (①–③) and repair the `moai factory runs` owner-probe cost (④).

Working tree for this card: the t1378 worktree, branch base `d194083fb`.

## §B — Known Issues (plan-phase diagnosis record)

1. **① Claim path has NO wait/retry.** `claimFactoryLane` (`internal/kanban/factory_slots.go:246-256`) fails immediately on slot exhaustion; no sleep, retry, or deadline exists in the claim path. Therefore the rc.23 long-launch delay is NOT in the claim itself — it is in the pre-claim launch sequence: `enterCodexFactory` (`internal/cli/codex_factory.go:119-156`) → `enterFactoryLaneRun` (`internal/cli/factory.go:492`) → `enterSelectedFactoryRun` (`factory.go:417`, git-tree check + `factorymsg.ResolveActiveRun` SQLite read) → the codex pre-exec init (config wiring, skills seed, local instructions) before the launcher replaces itself with codex — POSIX-only via `syscall.Exec` (`internal/cli/codex_direct_posix.go:53`, `//go:build !windows`); Windows takes its own exec mechanism (`internal/cli/codex_direct_windows.go`), so the timing instrumentation of REQ-012 must sit on the shared pre-exec sequence, ahead of the platform exec split. Run-phase must instrument per-step timing (REQ-012) and attribute the delta between the last immediate-error build and `d194083fb`; the candidate set above is where the instrumentation lands first.
2. **② No worktree scan found in the launch path.** Greps for `filepath.Walk|os.ReadDir|filepath.Glob` over `internal/cli/codex_launcher.go`, `codex_init.go`, `session*.go` return nothing. The ghost-directory hypothesis for the launch delay is weakened for launcher code; ghost hygiene remains sweep-tooling territory (Out of Scope).
3. **③ Hook-trust message is a wiring-pass emit, not a launcher gate.** `ReTrustGuidance` (`internal/codexwiring/codexwiring.go:74`) is printed when `moai update` rewrites `.codex/hooks.json`; codex then holds changed hooks untrusted. The launcher execs codex and does not wait on hook approval — a post-TUI stall is codex-side and outside moai code (boundary note, Out of Scope).
4. **④ Owner probe is per-run-row.** `moai factory runs` = `OpenFactory` + `ClassifyRuns` (`internal/cli/factory_handoff_recover.go:79-101`); classification probes each owner pid through `platformProcessFingerprint(pid)` — a per-pid seam with three build-tagged variants: `process_fingerprint_unix.go` (`//go:build !windows && !darwin`, one `ps -o lstart= -p <pid>` subprocess per probe), `process_fingerprint_darwin.go` (`//go:build darwin`, one `unix.SysctlKinfoProc` sysctl per probe — no subprocess), and `process_fingerprint_windows.go` (`//go:build windows`). The unix-file `ps` citation alone is wrong on this darwin tree where the 3.0s user / 1.86s sys cost was measured — the darwin variant pays one sysctl per row, and whatever share of the measured cost is not probe work (e.g. open/schema work) is attributed by M3's measurement. `OwnerClassifier` (`internal/homestate/factory_run_retire.go:44`) is already the seam a batched probe plugs into; the batching surface is the classifier seam plus ALL platform variants, and AC-007's instrumented test pins the platform-appropriate variant(s) (darwin here; unix/windows via their build-tagged test files).

## §C — Pre-flight

- [ ] Branch base re-read immediately before work (`git rev-parse --short HEAD` = `d194083fb` expected).
- [ ] `go test -timeout 30m ./internal/kanban/...` green on the base tree before the first change.
- [ ] Existing tests to keep passing named: `internal/kanban/factory_slots_test.go` (`TestClaimFactoryLaneWithinBounds`, `TestClaimFactoryLaneWithinConcurrentOneSlot`), `internal/kanban/role_naming_m1_test.go` (legacy refusal trio, lines 76/109/126).

## §D — Constraints

Restated from `spec.md` §D: Go error-wrapping and no-hardcoding conventions; `t.TempDir()`; no OTEL `t.Setenv`; lane-env-reading tests pin all axes; re-measurement unit is the whole `internal/kanban` package plus targeted touched-package tests; capacity policy changes go through `internal/config/defaults.go`.

## §E — Design Decisions (highest change-likelihood first)

1. **Where capacity authority lives (data-model decision).** Option A: grow the bounded auto scan to `maxCanonical+1` unconditionally — rejected: it silently voids the t1294 contract for operator-declared capacity. Option B (chosen): record the run's declared capacity at leader start — explicit `-f N` records N; absence records a derived-capacity marker — and the join reads run state: explicit capacity keeps the hard bound (REQ-006), the derived marker enables growth (REQ-005). This adds one datum to run registration (`recordFactoryRunStart`, `internal/cli/factory.go:564`) and one read on the join path; it is the first milestone because the claim semantics depend on it.
2. **Shared pool statement (policy decision).** Lane numbers stay one run-scoped pool across backends (REQ-007). The codebase already treats lane joins as backend-crossing (`refuseCodexLeaderRun` doc, `internal/cli/factory.go:545-546`); a per-backend namespace would break label-addressed leader routing. Chosen: make the implicit policy explicit, zero migration.
3. **Default capacity marker semantics (config decision).** `config.DefaultFactoryLeaderLanes` remains the leader's default fan-out; the derived-capacity marker is a distinct, explicit value in the run record — not an overload of "1". The constant's role narrows to "leader fan-out default", documented in `internal/config/defaults.go`.
4. **④ fix shape.** Batch the process-identity probe at the existing `OwnerClassifier` seam (bounded probe work per listing invocation instead of one `platformProcessFingerprint` call per row), covering every build-tagged platform variant the fix touches. If the platform-correct batch read proves larger than a single-function change, the diagnosis (probe-count instrumented test + measured before/after) still lands in-card and the batching itself becomes a named follow-up card — record the split decision in progress evidence.

## §F — Milestones (priority-ordered, no time estimates)

- **M1 (High) — run capacity record.** Record declared capacity vs derived marker at leader start; join path reads it. AC-006. Files: `internal/cli/factory.go` (record side), run-state store (`internal/homestate`), join read (`internal/cli/codex_factory.go`).
- **M2 (High) — capacity-aware claim growth.** Bounded auto scan grows to one-past-highest-live-claim when the run is capacity-open; explicit requests and explicit-capacity refusals unchanged. AC-001, AC-002, AC-003. Files: `internal/kanban/factory_slots.go` + tests.
- **M3 (High) — ① latency instrumentation + ④ owner-probe batching.** Per-step pre-exec timing surfaced on the codex lane launch path, ahead of the POSIX/Windows exec split, with the slow-launch threshold default in `internal/config/defaults.go` (REQ-012; AC-010); batched owner probe at the `OwnerClassifier` seam across all `platformProcessFingerprint` build-tagged variants, with a probe-count test pinning the platform-appropriate variant(s) and before/after `moai factory runs` wall-time measurement recorded (REQ-013, REQ-014, AC-007). Files: `internal/cli/codex_launcher.go` or the adjacent shared pre-exec seam, `internal/homestate/process_fingerprint_{unix,darwin,windows}.go`.
- **M4 (Low, mechanical) — legacy-path parity verification.** Add the bounded-path legacy refusal test (AC-004); confirm the legacy trio and label-refusal tests pass unmodified (AC-005); whole-package re-measurement (AC-008).

## §G — Anti-Patterns (must not)

- Do not grow the bound for explicit `lane-<n>` requests or for explicitly-sized runs.
- Do not encode the capacity policy as an inline literal at a call site — config defaults live in `internal/config/defaults.go`.
- Do not make the claim path retry or wait — the observed immediate refusal is correct behavior for a genuinely full declared run; only the bound's derivation changes.
- Do not break or rewrite the legacy-run refusal tests to fit new behavior — new behavior gets new tests.
- Do not scan `.claude/worktrees/` in the launch path (finding ② — no scan belongs there).

## §H — Cross-References

- `spec.md` §B requirements; `acceptance.md` full Given-When-Then matrix.
- Premise verification with file:line evidence: `progress.md` § Premise Verification.
- Diagnosis boundary notes: `spec.md` §F (Out of Scope).
