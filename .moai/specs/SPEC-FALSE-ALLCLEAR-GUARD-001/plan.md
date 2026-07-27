# SPEC-FALSE-ALLCLEAR-GUARD-001 — Implementation Plan

> Ordering note: sections are ordered by decision-reversibility. §A-§C carry the decisions most likely to change under review (a new exported type, a user-facing default, and a shared function signature). §F-§G carry the mechanical steps. Review attention belongs at the top.

## §A Context

Two false all-clear defects from the census at `.moai/reports/census-2026-07-27-handoff.md`. Full problem statement in `spec.md` §A. Baseline for this plan: branch `feat/false-allclear-guard-001`, worktree `.claude/worktrees/faclear`, based on `origin/main` at `6763aff3b`.

**Tier: M.** Justification, stated both ways so the classification is auditable:

- LOC axis → M. The change is surgical: one new exported sentinel, one new logging decision function, three `errors.Is` branches, one propagated return value, one new doctor check, one documentation sentence in four locales. Estimated 400-700 LOC including tests; well under the >1000 threshold for L.
- File-count axis → borderline. Production Go files touched: **9** (count corrected per plan-audit iter-1 D13 — the prior "8" omitted the NEW logging file from M1 step 2): `deps.go`, `root.go`, the new logging file, `scanner.go`, `astgrep.go`, `astgrep_gate.go`, `gate.go`, `pre_tool.go`, `doctor.go`. Adding the 4 mechanical docs-site locale files and ~5 test files reaches **~18**, which crosses L's ">15 files" guidance.
- Deciding rationale: the file count crosses only because 4 of the files are same-sentence translations of one edit and ~5 are test files. The production surface is 9 files, inside M's 5-15 band, and the change is not constitutional. **Tier M**, artifact set spec.md + plan.md + acceptance.md, plan-auditor PASS threshold 0.80.
- **Recorded tier trade-off** (audit-noted, not a demand): Tier M sets the threshold at 0.80 rather than 0.85 and drops `design.md` + `research.md` — the artifacts most likely to have caught the G-6 error above (a single loader-file read). The tier remains the user's call; the trade-off is recorded so the choice is auditable.

## §B Verified baseline (measured during plan-phase, in this tree)

Each row was observed, not assumed. Anything not observed is listed in §E.

| Fact | Location | Observation |
|------|----------|-------------|
| Unconditional discard, full path | `internal/cli/deps.go:90-91` | `slog.New(slog.NewTextHandler(io.Discard, nil))` + `slog.SetDefault` |
| Unconditional discard, light path | `internal/cli/root.go:105-106` | same two lines |
| Only two `SetDefault` sites in non-test code | `internal/cli`, `cmd/` | grep returned exactly these two |
| Subcommand-knowable seam | `internal/cli/root.go:62-73` | `Execute()` branches on `isTrivialCommand(os.Args[1:])` |
| `hook` subcommand name | `internal/cli/hook.go:31` | `Use: "hook"` |
| Callsite counts (non-test, `internal/`+`cmd/`) | — | Warn 198, Error 20, Info 78, Debug 84 |
| `MOAI_LOG_LEVEL` constant | `internal/config/envkeys.go:22` | `EnvLogLevel = "MOAI_LOG_LEVEL"` |
| Env read into config | `internal/config/manager.go:397-398` | writes `cfg.System.LogLevel` |
| No handler consumer | repo-wide | `System.LogLevel` appears only at manager.go (write) and validation.go (string check) |
| Scanner skip site | `internal/astgrep/scanner.go:234-240` | `slog.Warn` then `return []Finding{}, nil` |
| Availability probe | `internal/astgrep/scanner.go:215-222` | `exec.LookPath` on `cfg.SGBinary`, default `"sg"` |
| `Scan` production callers | — | exactly two: `internal/cli/astgrep.go:91`, `internal/hook/quality/astgrep_gate.go:52` |
| CLI error path | `internal/cli/astgrep.go:91-94` | wraps and returns; findings-with-errors returns `&exitCodeError{code: 1}` |
| Gate error path | `internal/hook/quality/astgrep_gate.go:52-56` | `if err != nil { return true, "" }` |
| **Gate caller drops pass-path output** | `internal/hook/quality/gate.go:284-287` | `if ok, out := ...; !ok { return false, out }` — `out` unused when `ok` |
| **Hook caller drops pass-path output** | `internal/hook/pre_tool.go:390-397` | `passed, output := gate.Run(ctx)`; `output` unused when `passed` |
| `SystemMessage` channel exists | `internal/hook/types.go`, used in `post_tool.go:263`, `auto_update.go:100`, `post_compact.go:62` | `HookOutput.SystemMessage` |
| Correct sibling pattern | `internal/cli/astedit.go:87-91` | `IsSGAvailable` false → stdout notice → `return nil` |
| doctor check assembly | `internal/cli/doctor.go:164-169` | `systemChecks` slice of `{name, func(bool) DiagnosticCheck}` |
| Optional-binary check shape | `internal/cli/doctor.go:317-344` | `checkGitHubCLI`: `exec.LookPath("gh")` → `uikit.CheckWarn` + message |
| **doctor always exits 0** | `internal/cli/doctor.go:99-124` | counts failures for the `--fix` hint, then `return nil` unconditionally |
| Docs claim (all 4 locales) | `docs-site/content/{en,ko,ja,zh}/cli-reference/ast-grep.md:9` | "each command prints a notice and exits without an error" |

### B.1 Corrections to the census brief

Recorded so the run phase does not re-derive them:

1. **Gate call chain.** The census names `internal/hook/pre_tool.go:566 → internal/hook/quality/astgrep_gate.go:51`. The actual sole non-test caller of `RunAstGrepGateV2` is `internal/hook/quality/gate.go:284`, reached from `pre_tool.go:390` via `gate.Run(ctx)`. The chain has **three** frames, not two.
2. **Callsite count.** The census states "251 `slog.Warn`/`slog.Error` callsites". Measured here: 198 + 20 = **218** (non-test, `internal/` + `cmd/`). Use 218.
3. **`lsp doctor` pattern.** The census says to clone the `lsp doctor` check pattern. `moai lsp doctor` (`internal/cli/lsp_doctor.go`) is a *separate subcommand* producing its own report with an `InstallHint` field — not a `moai doctor` check. The structural template for a `moai doctor` check of an optional external binary is `checkGitHubCLI` (`doctor.go:317`). Take the *shape* from `checkGitHubCLI` and the *install-hint idea* from `lsp_doctor.go`.
4. **doctor exit code.** `moai doctor` returns nil regardless of failure count. No acceptance criterion may assert a non-zero exit from `moai doctor`.

## §C Decisions

Highest-reversibility first.

- **D-1 (new exported surface) — sentinel error.** Export `ErrScannerUnavailable` from `internal/astgrep`. `Scan` returns it wrapped (`fmt.Errorf("...: %w", ErrScannerUnavailable)`) so the message can name the binary and the install URL while `errors.Is` still matches. This is the only new exported symbol in the SPEC; changing its name or package later is a breaking change for the three call sites.
- **D-2 (user-facing default) — non-hook default level is `warn`, stderr.** Chosen so that operators see the 218 warn-and-above records without being flooded by the 162 info/debug records. `MOAI_LOG_LEVEL` raises or lowers it. Unrecognized values fall back to `warn` rather than erroring, so a typo cannot break a CLI run.
- **D-3 (scope of the hook carve-out) — unconditional.** `moai hook` discards records regardless of `MOAI_LOG_LEVEL`. This is the literal reading of the recorded decision. The alternative (let `MOAI_LOG_LEVEL` re-enable hook logging as a debugging escape hatch) was considered and not taken; hook-path observability is named Out of Scope in `spec.md` §C.
- **D-4 (shared signature) — pass-path reason propagation.** `QualityGate.Run` must carry a pass-path reason. Two shapes are available: keep `(bool, string)` and return `(true, reason)`, or widen the return. Prefer keeping `(bool, string)` — every existing caller already destructures two values, and the `!ok` branch semantics are unchanged. The change is that the `ok == true` return stops being hard-coded to `""`. Run-phase must check every `Run` caller and every test asserting `Run` returns `""` on success.
- **D-5 (emission channel) — `HookOutput.SystemMessage`, not `slog`.** Forced by D-3: a `slog.Warn` added in `pre_tool.go` runs on the `hook` path and is therefore discarded by this SPEC's own M1 decision. `SystemMessage` is the established non-blocking notice channel in this package.
- **D-6 (exit code) — `moai ast-grep` exits 1 when `sg` is absent.** Per `internal/cli/CLAUDE.md`, 1 = user error, 2 = system error. A missing prerequisite the user can install is a user error. Reuse the existing `exitCodeError` type (`internal/cli/constitution.go:318`) already used by this command for the findings-detected path.
- **D-7 (asymmetry with `ast-edit`) — deliberate and documented.** `ast-edit` keeps exit 0. `ast-grep` is a detector used as a CI gate, so a silent pass is a false all-clear; `ast-edit` is a mutator for which "nothing to apply" is a true no-op. The asymmetry is the point, and REQ-FAG-028 requires the docs to say so.
- **D-8 (doctor status) — `Warn`, not `Fail`.** `sg` is optional. `Fail` would misreport an intact installation and would inflate the `--fix` suggestion list.

## §D Constraints

- `internal/cli/CLAUDE.md`: stdout is machine-readable, stderr is human progress/warnings/errors; never mixed. Exit codes 0 / 1 / 2 as above. No `AskUserQuestion` anywhere in `internal/cli` or `internal/hook`.
- CLAUDE.local.md §14: env var names come from `internal/config/envkeys.go` constants; no inline literals.
- CLAUDE.local.md §17 + §2: any `docs-site` change is a 4-locale same-change-set obligation. `docs-site` is not template-managed, so no `make build` is implied by the docs edit alone.
- Go conventions (`.claude/rules/moai/languages/go.md`): table-driven tests, `t.TempDir()` for temp dirs, `fmt.Errorf(... %w ...)` for wrapping, `go vet` before commit.
- **Formatting scope (plan-audit iter-2 N1)**: `gofmt`-clean is required only for the files this SPEC touches (the explicit list in `acceptance.md` AC-FAG-018). `internal/cli/root.go` is currently unclean (6 whitespace lines in the `trivialCommands` map literal) and MUST be brought to clean by M1, which edits it anyway. The other **106** repo-wide unclean files are pre-existing debt and MUST NOT be reformatted here — see `spec.md` §C. Note the repo's real formatter is `gofumpt` (`Makefile:61`), which is **not installed** on the authoring host, and `.github/workflows/` has no format guard at all; do not add a criterion that depends on `gofumpt` being present.
- Test isolation (CLAUDE.local.md §6): never `t.Setenv("HOME", ...)`; use `t.TempDir()`. `t.Setenv` is incompatible with `t.Parallel()` in the same test — the `MOAI_LOG_LEVEL` tests must be non-parallel or use a level-resolution function that takes the value as a parameter.
- No time estimates anywhere. Priority labels only.

## §E Known gaps (not observed during plan-phase)

Listed explicitly so no acceptance criterion silently assumes them.

- **G-1.** Whether `sg` is installed on the run-phase machine was not checked. Every criterion in `acceptance.md` that needs `sg` absent constructs that condition by stripping `PATH`, so the outcome does not depend on the host.
- **G-2.** The existing tests that call `InitDependencies()` directly (`deps_test.go`, `integration_test.go`, `hook_e2e_test.go`, per the `@MX:ANCHOR` fan_in note at `deps.go:78-79`) were **not** read. Moving `slog.SetDefault` out of `InitDependencies` may make those tests emit warn records to the test log. This is the single most likely source of unexpected churn in M1 — read them first (see §F M1 step 1).
- **G-3.** Tests asserting `QualityGate.Run` returns an empty string on success were not enumerated. D-4 changes that value for one branch. `gate_test.go` is 26 KB and was not read.
- **G-4.** Whether any `--format=json` consumer parses `moai ast-grep` stdout in CI was not established. REQ-FAG-016 keeps the new message off stdout precisely so this does not matter, but the guarantee is by construction, not by observation of a consumer.
- **G-5.** The claim "reverting the slog change returns stderr to 0 bytes" (AC-FAG-004b) is a stated mechanism that was **not executed** during plan-phase. It is derived from reading `scanner.go:236` + `deps.go:90`. Run-phase must actually perform the revert-and-observe round trip rather than assert the mechanism.
- **G-6 — RETRACTED (plan-audit iter-1 D2).** A prior revision recorded this gap as "whether the ast-grep sub-gate is enabled on the config-loaded path is unresolved, and the two signals conflict". **That premise was false**, and it was asserted from partial reading inside a section whose header promises "Each row was observed, not assumed" — an unobserved-premise claim (`verification-claim-integrity.md` §1.1 surface 3). The repository resolves it definitively in a 31-line file the first pass did not read:
  - `internal/config/loader.go:36` — `cfg := NewDefaultConfig()` seeds the load from defaults.
  - `internal/config/defaults.go:297` — `AstGrepGate: {Enabled: true, BlockOnError: false, WarnOnlyMode: true}`.
  - `internal/config/loader.go:89` — calls `l.loadGateSection(sectionsDir, cfg)`.
  - `internal/config/loader_gate.go:20-21` — `wrapper := &gateFileWrapper{Gate: cfg.Gate}` seeds the wrapper from the already-populated default; the file's own doc comment states verbatim that *"an absent gate.yaml, or one that omits ast_grep_gate keys, yields Enabled=true + WarnOnlyMode=true"*.

  **Resolved fact: both paths yield `Enabled: true`, advisory (`WarnOnlyMode: true`). There is no conflict.** Consequence for the ACs: AC-FAG-011/012 still construct the config explicitly, but for **test determinism** (a test must not depend on ambient project config), NOT to dodge a phantom default ambiguity.

  **Stale-source warning**: `CLAUDE.local.md` §2.2 asserts "gate.yaml/gate 로더 부재 → `AstGrepGate.Enabled` 항상 컴파일 기본값 false". That local doc is stale relative to `loader_gate.go` and MUST NOT be used as a source for this question.
- **G-7 — SPLIT AND PARTLY CLOSED (plan-audit iter-2).** The prior blanket deferral of the AC-FAG-017/018 baselines was wrong in one half. **Rule adopted: a deferral is justified by cost, not by not having run it.** Now closed: `go vet ./...` → **exit 0**; `gofmt -l internal cmd` → **107**, and the scoped per-file measurement → **1** (`internal/cli/root.go`). Measuring `gofmt` is precisely what exposed N1 — the previous AC-FAG-018 expected 0, which no correct implementation of this SPEC could ever produce. Still deferred, and accepted as cost-justified: the `go test ./...` skip-count baseline (expensive, runs against a shared tree) and the Windows cross-build. Both remain Definition-of-Done gates measured before M1 begins.
- **G-8.** Shape 2 (`PATH`-stripping) is POSIX-only; the shell criteria carry no Windows coverage. Windows is covered only by AC-FAG-018's cross-build.
- **G-9 (new, plan-audit iter-2 N2).** Every warn-or-above record reachable on the full-init path is **conditional on a failure state** — `deps.go:112` (gopls config load error), `deps.go:120` (bridge init error), `deps.go:143` (config load error), `deps.go:174` (bridge nil), `internal/config/loader.go:41` (sections dir absent). None fires on a healthy host: `gopls` is present at `/Users/goos/go/bin/gopls`, `.moai/config/sections/lsp.yaml` exists and is enabled, and the sections dir exists. AC-FAG-006's discriminator therefore does NOT rely on log volume; it asserts the `deps` global (`deps.go:76`, assigned only by `InitDependencies` at `deps.go:132` / `deps.go:271`), an exact host-independent witness of whether full initialization ran. **If run-phase changes how or where `deps` is assigned, AC-FAG-006 must be re-derived — not silently kept.**

## §F Milestones

Sequential. M1 before M2 — M2's verification depends on M1's observable.

### M1 — Scope the slog suppression to the `moai hook` path (Priority: High)

Covers REQ-FAG-001 .. 009.

1. Read `internal/cli/deps_test.go`, `internal/cli/root_test.go`, and any other direct caller of `InitDependencies` (G-2) before editing. Decide whether the tests need an explicit handler reset.
2. Add a single logging-configuration function in `internal/cli` that takes the argument vector and returns the handler decision. It is the only place that calls `slog.SetDefault`.
   - `hook` detection mirrors the existing `isTrivialCommand` walk in `root.go:87-99` (skip flags, first non-flag arg is the subcommand).
   - Level resolution reads `os.Getenv(config.EnvLogLevel)`; parse to `slog.Level`; on empty or unparseable, use `slog.LevelWarn`.
   - Non-hook handler writes to `os.Stderr`. Hook handler writes to `io.Discard`.
3. Remove the `slog.SetDefault` + discard-handler lines from `InitDependencies` (`deps.go:90-91`) and `initLightDeps` (`root.go:105-106`); call the new function once from `Execute()` before the trivial/full branch.
4. Tests: level resolution table (unset → warn; `debug` → debug; `error` → error; `nonsense` → warn), hook-vs-non-hook handler selection, and a stdout-cleanliness assertion.
5. `go vet ./... && go test ./internal/cli/... ./internal/config/...`.

**Falsification for M1**: with `sg` absent, `moai ast-grep <path>` emits a non-zero number of bytes on stderr — produced by the *pre-existing* `slog.Warn` at `scanner.go:236`, with no M2 change in the tree. A comparison binary built from `6763aff3b` returns stderr to 0 bytes (AC-FAG-004).

**Why M1 precedes M2 (rationale corrected, plan-audit iter-1 D12).** The ordering decision (D-1 in the brief) is unchanged; only its stated justification is corrected. `spec.md` §A.3 previously claimed "any fix to Defect 2 can only be verified indirectly while Defect 1 is unrepaired". That is not supported by this SPEC's own criteria: AC-FAG-007/011/012 are Go tests independent of the logging path, and AC-FAG-008's stderr guidance comes from the CLI's *new* `cmd.ErrOrStderr()` write, not from `slog`. M2 is directly verifiable without M1. The **real** dependency runs the other way: AC-FAG-004 uses the *pre-existing* `scanner.go:236` warn as its observable, and that observable is superseded once M2 changes the CLI's stderr output — so AC-FAG-004 is only cleanly evaluable in the M1-done / M2-not-started window. M1 first is what preserves that window.

### M2 — ast-grep scanner sentinel, gate reason, doctor check, docs (Priority: High)

Covers REQ-FAG-010 .. 029. Sub-steps in dependency order.

1. **Sentinel** (`internal/astgrep/scanner.go`): export `ErrScannerUnavailable`; change the `!isSGAvailable()` branch to return the wrapped sentinel. Keep the existing `slog.Warn` — M1 has made it visible. Tests: `errors.Is` true for the unavailable path; `errors.Is` false for a clean scan.
2. **CLI** (`internal/cli/astgrep.go`): branch on `errors.Is(err, astgrep.ErrScannerUnavailable)` before the generic wrap at line 93; write guidance to `cmd.ErrOrStderr()`; return `&exitCodeError{code: 1, ...}`. The existing generic error path is unchanged for other errors. Assert `internal/cli/astedit.go` is untouched.
3. **Gate step** (`internal/hook/quality/astgrep_gate.go`): replace `if err != nil { return true, "" }` with two branches — sentinel → `(true, <skip reason>)`, other error → `(true, <distinct degraded reason>)`.
4. **Gate propagation** (`internal/hook/quality/gate.go:281-287`): capture the pass-path reason instead of discarding it, and return it from `Run` on the success path. Reconcile with G-3 (tests asserting `""`).
5. **Hook surface** (`internal/hook/pre_tool.go:388-397`): on `passed == true` with a non-empty output, return a `HookOutput` carrying `SystemMessage`. Preserve the existing deny path byte-for-byte.
6. **Doctor** (`internal/cli/doctor.go`): add `checkAstGrep` modelled on `checkGitHubCLI` (lines 317-344); register it in `systemChecks` (lines 164-169). `uikit.CheckWarn` + install guidance when `LookPath` fails.
7. **Docs** (mechanical, last): rewrite the shared sentence at line 9 of all four `docs-site/content/{en,ko,ja,zh}/cli-reference/ast-grep.md` files to state the differing behavior and carry install guidance. `ko` is the canonical translation source per the docs-site rules; `en` is the primary authored surface for this page — keep all four semantically identical.
8. `go vet ./... && go test ./...`.

## §G Risk register

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| R-1 | Moving `SetDefault` out of `InitDependencies` makes direct-calling tests noisy or non-deterministic (G-2) | High | Low | Read those tests first (M1 step 1); add an explicit test-side handler if needed |
| R-2 | `QualityGate.Run` returning a non-empty string on success breaks an existing assertion (G-3) | Medium | Medium | Enumerate `Run` assertions in `gate_test.go` before editing; the `!ok` branch is unchanged, so only success-path assertions are at risk |
| R-3 | The gate reason is implemented but never surfaces, because a frame still drops it (§B.1 correction 1) | Medium | **High** — reproduces the exact defect class this SPEC exists to remove | AC-FAG-011 is an end-to-end assertion through the real `pre_tool` handler, not a unit test of `astgrep_gate.go` |
| R-4 | Newly-visible warn records surprise users as apparent new errors on unrelated commands | Medium | Low | 218 warn/error callsites become visible at once; this is the intended repair, and `MOAI_LOG_LEVEL=error` is the documented narrowing lever |
| R-5 | `moai ast-grep` exiting 1 breaks a CI pipeline that currently passes green | Low | Medium | That green is exactly the false all-clear being removed; the change is intentional and is documented in all four locales (REQ-FAG-028) |
| R-6 | AC-FAG-004b's revert-and-observe mechanism was not executed at plan-phase (G-5) | Medium | Low | The criterion is written as an executed round trip, not as an assertion about a mechanism |
| R-7 | A `PATH`-stripped test invocation fails for an unrelated reason (binary not found) rather than for the reason under test | Medium | Medium | Every stripped-`PATH` criterion invokes the binary by absolute path, and pairs the assertion with a positive control |
| R-8 | Locale drift — one of the four docs pages is edited and the others are not | Low | Low | **Mitigation corrected (plan-audit iter-1 D5).** The prior mitigation claim was false: its two greps did not discriminate — `grep -lc 'ast-grep.github.io' … \| wc -l` already returned 4 on the unmodified tree, and the stale-claim grep's baseline was 1 (en only), so an en-only edit passed both. Replaced by AC-FAG-014(a) per-locale **anchored** diff vs `6763aff3b` (measured baseline 0/0/0/0 — all four must flip) plus (b) a 4/4 count of a locale-neutral anchor that appears in **zero** files today (`guide/quick-start`) |

## §H Cross-references

- `.moai/reports/census-2026-07-27-handoff.md` §3-A, §3-B C-1, §3-D SLOG-01, §7 Priority 1 (P1-A, P1-C)
- `internal/cli/CLAUDE.md` — stream discipline, exit codes, subagent boundary
- `internal/hook/CLAUDE.md` — hook JSON I/O contract, timeout discipline
- CLAUDE.local.md §14 (env constants), §17 (docs-site 4-locale), §6 (test isolation)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
- `SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001` (completed) — prior repair of two other `moai doctor` false signals; this SPEC adds a check to the same command without altering the checks that SPEC repaired
