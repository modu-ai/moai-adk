# Plan — SPEC-CODEX-E2E-GUARD-001

## §A Context

Card t500 (C7 · test card), tree `ace1c5440`, branch `WT-codex-e2e-guard`, worktree
`.claude/worktrees/t500`. Three axes: (1) the doctor "Codex Wiring" judgment has never run
behind the real init command path; (2) the card's statusline-gap premise is REFUTED — the
assertion exists — so the deliverable is a mutant proof plus a correction record; (3) the
launcher static guards scope to a 2-file set while 12 non-test codex CLI files exist. Side-note:
the "41 files" figure resolves to the repo-wide codex-named test-file population (spec.md §F.2).

All file:line citations in this plan were read against `ace1c5440`; run-phase re-pins its own
SHA before relying on them.

## §B Known Issues (plan-phase findings)

1. `wireProjectForDoctor` (`doctor_codex_test.go:26-34`) bypasses init — the doctor check is
   only ever judged against a `Wire`-only project (axis-1 gap).
2. The init-side tests assert artifacts/announcements only; no init→doctor chain exists
   (`init_agent_wizard_test.go:64-107`).
3. `codexSpecFiles` (`codex_launcher_guards_test.go:30`) covers 2 of 12 non-test codex files
   for AC-CL-014/016; AC-CL-013's cobra walk covers command strings only.
4. AC-CL-016 cannot blanket-extend: `codex_review_gate.go:129` legitimately starts `git`, and
   `mcp_codex.go` has three `binaryPath` sites — a single expected first-arg would either
   false-RED today or pass vacuously.
5. Card premise on axis 2 is false (spec.md §F.1) — recorded, not silently absorbed.

## §C Pre-flight (before any test execution)

1. Re-pin base: `git rev-parse --short HEAD` → record in progress.md §E.2.
2. Baseline GREEN (already observed at plan phase, re-run at run phase):
   - `go test ./internal/codexwiring/ -run TestStatusLine -count=1`
   - `go test ./internal/cli/ -run 'TestCodexSpecFiles|TestCodexCommand_NeutralityScan|TestCodexSpawn_TmuxDiagnosticSingleSource' -count=1 -timeout 600s`
3. Confirm `git status --porcelain internal/` is clean before the mutant step (M1) so the
   revert is attributable.

## §D Constraints (from spec.md §D — binding)

- Scoped verification only: `./internal/cli/...`, `./internal/codexwiring/...`. NEVER
  `go test ./...` locally (10 lanes share this machine). CI owns the full-suite verdict.
- `internal/cli` runs carry `-timeout 600s`.
- `t.TempDir()` isolation; no OTEL `t.Setenv`; `HOME` pinned per the init-test convention.
- Axis-1 test hermetic: no real codex binary, no network.
- Axis-2 mutant transient: never staged, never committed; revert verified by
  `git status --porcelain internal/codexwiring/configtoml.go` printing nothing.
- No guard weakened; the 2-file baseline assertions survive inside the 12-file form.

## §E Self-Verification (run-phase close)

E1 AC matrix GREEN/FAIL with verbatim `go test` outputs; E2 scoped-package verification batch
(both packages, `-count=1`); E3 `golangci-lint run ./internal/cli/... ./internal/codexwiring/...`;
E4 `git status --porcelain internal/codexwiring/configtoml.go` empty (mutant not committed);
E5 progress.md §E.2 carries the RED and GREEN mutant outputs verbatim; E6 `moai spec lint`
0 findings at sync.

## §F Milestones

### M1 — Axis-2 mutant proof (HIGH, evidence collection — transient)

1. Apply the 1-char mutant to `internal/codexwiring/configtoml.go:46`
   (`"git-branch"` → `"git-branchx"`).
2. Run `go test ./internal/codexwiring/ -run 'TestStatusLineDefaultSubsetOfAllowlist' -count=1`
   → observe RED (expected failure: `default token "git-branchx" is not in
   statusLineAllowlist`). Capture verbatim.
3. Revert (restore the exact original byte). Run the same command → GREEN. Capture verbatim.
4. `git status --porcelain internal/codexwiring/configtoml.go` → empty.
5. Record all three outputs in progress.md §E.2. Nothing from this milestone is committed
   except the progress record.

### M2 — Axis-1 init→doctor e2e test (HIGH)

1. RED-first discipline does not apply directly (this is a coverage gap, not a bug): write the
   test FIRST and expect GREEN. A RED outcome is a DISCOVERED DEFECT — stop and report to the
   orchestrator; do not repair `checkCodexWiring` or the init flow inside this card.
2. New file `internal/cli/doctor_codex_e2e_test.go` (same package — reuses the existing
   package-level helpers `runInitForAutonomyAtHomeCapturingOut`, `stubMoaiLookup`,
   `stubCodexHome`, `assertCodexArtifacts`):
   - `TestRunInit_ThenDoctorCodexWiringHealthy`: wizard answers codex (`AgentWiring: "codex"`,
     `MCPProvision: true`), `HOME` pinned to a temp dir, real init via
     `runInitForAutonomyAtHomeCapturingOut`; then `stubMoaiLookup(t, true)` +
     `stubCodexHome(t, t.TempDir())`; then `checkCodexWiring(projectDir, false)` →
     assert `CheckOK` and no codex-finding text. Sanity leg: `assertCodexArtifacts(t,
     projectDir, true)` so a silently-unwired init cannot vacuously pass the OK assertion.
   - Companion negative `TestRunInit_ClaudeOnlyThenDoctorStaysSilent` (REQ-CEG-002): wizard
     claude-only, `stubCodexLookup(t, true, false)` (no codex binary) → `CheckOK` informational
     outcome, no codex `Warn`. **Descope rule**: if this companion balloons (e.g. the
     claude-only init path needs seams the codex path does not), descope it at the kickoff
     gate with the reason recorded in progress.md — REQ-CEG-001 is the binding half.
3. Constraints: hermetic (no real codex, no network), `t.TempDir()`, `HOME` pinned, no OTEL
   env, `-timeout 600s` on the package run.

### M3 — Axis-3 guard expansion: build-tag + exec (HIGH)

1. Rename/extend the guard set (`codex_launcher_guards_test.go:30`) from the 2-file
   `codexSpecFiles` to the 12-file set (keep the old name or rename to `codexGuardFiles` with
   the comment updated — the AC-CL-007 reconciliation comment at `:27-29` is rewritten to
   state the sentinel-vs-process-start distinction recorded in spec.md §A axis 3). The set
   variable is length-pinned: `len(guardFiles) != 12` is a `t.Fatalf("guard swept %d files,
   want 12 — vacuous green", …)` at the top of every iterating guard test, so a green
   byte-identical to today's 2-file run cannot occur (RED-now observations in acceptance.md
   AC-CEG-005/006; cross-checked against the independent count
   `ls internal/cli/*codex*.go | grep -v _test | wc -l` → `12`).
2. AC-CL-014 form: unchanged assertions, wider iteration — expected GREEN (plan-phase measured
   zero build tags / zero syscall imports on the 10 new files; re-verify at run phase).
3. AC-CL-016 form — per-file expected-first-argument table (from spec.md §A, re-verified at
   run phase by reading each site):

   | File | Expected first argument |
   |---|---|
   | `codex_launcher.go` | `req.Program` |
   | `mcp_codex.go` | `binaryPath` (all three sites: `:350`, `:433`, `:1889`) |
   | `codex_review_gate.go` | `"git"` (the porcelain probe) |
   | `codex_readiness.go` | zero call sites (existing assertion kept) |
   | `codex_contract.go`, `codex_init.go`, `codex_job_control.go`, `codex_jobs.go`, `codex_task.go`, `doctor_codex.go`, `hook_harness_codex.go`, `update_codex_wiring.go` | zero call sites |

   Preserve: comment-only-line skip (`:156-160`), `t.Helper()` helpers, ordered-phrase style.
   The zero-call-site expectation becomes a positive assertion (any new process-start primitive
   in those 8 files must update the table — that friction is the guard working).
4. Run the guard tests GREEN on the 12-file form.

### M4 — Axis-3 neutrality extension (MEDIUM)

1. Baseline-measure first: extract string literals (comments excluded) from the 12-file set and
   run the existing forbidden-pattern classes (`codex_launcher_guards_test.go:87-95`) +
   non-ASCII count over them. Observe the baseline BEFORE writing assertions. The scan
   iterates the same length-pinned 12-file set as M3 (`len != 12` → `t.Fatalf`) and carries a
   POSITIVE-CONTROL canary: the scan must observe at least one known existing literal
   (`"git"` at `codex_review_gate.go:129`), so a scanning failure reads RED instead of a
   vacuous green (acceptance.md AC-CEG-007 RED-now/green-path pair).
2. Where the baseline is clean: land the literal-scan guard (same pattern classes, same
   error style). Where a baseline literal violates a class: narrow or allowlist that class
   with the justification recorded in this plan section and progress.md — never weaken the
   existing `TestCodexCommand_NeutralityScan` launcher scan.
3. Reconciliation: state in the guard comment that AC-CL-007 owns sentinel-VALUE classification
   and this guard owns string-literal neutrality — different surfaces, no double-claim
   (spec.md §A axis 3, `codex_launcher_guards_test.go:27-29`).
4. Run GREEN.

### M5 — Verification, evidence persistence, commit (LOW)

1. Scoped batch (one turn, parallel):
   - `go test ./internal/cli/... -count=1 -timeout 600s` (affected package full — this IS the
     scoped suite for internal/cli changes; still not `./...`)
   - `go test ./internal/codexwiring/... -count=1`
   - `golangci-lint run ./internal/cli/... ./internal/codexwiring/...`
2. Persist evidence under `.moai/state/verify/` (resolvable path, not `/tmp`), including the
   swept-count observations (the `-v` guard runs showing the `len == 12` assertion passing,
   plus the independent `ls … | wc -l` → `12` cross-check) that pair with the RED-now cells
   in acceptance.md.
3. Commit by explicit pathspec (test files + SPEC artifacts); Conventional Commit
   `test(cli): …` carrying the card id `t500`. Re-read `git status --short` and branch state
   immediately before staging.

Priority ordering: M1 (cheap, unlocks the §F.1 evidence) → M2 (the card's headline) →
M3 (mechanical, table already measured) → M4 (design-heaviest, baseline-gated) → M5.

## §G Anti-Patterns

- Designing the exec guard from the armchair — the table in M3 exists because the call sites
  were READ; re-read them at run phase before coding (memory: "a brief written without reading
  the fixtures reproduces the defect").
- Blanket `req.Program` expectation — false-RED on `codex_review_gate.go`'s `git` today.
- Committing the mutant, or "temporarily" leaving it — M1.4 is the gate.
- `go test ./...` locally; missing `-timeout 600s`; `t.Setenv` with OTEL vars.
- Weakening an existing guard assertion to make a wider scan pass.
- Treating the axis-1 test going RED as "fix the doctor" — it is a defect REPORT.
- Restating GEARS REQs as Given-When-Then inside spec.md — GWT lives in acceptance.md only.

## §H Cross-References

- `SPEC-CODEX-E2E-MEASURE-001` — the measurement card whose filename-axis defined the 41
  population (spec.md §F.2); this SPEC's predecessor in the same card family.
- `SPEC-CODEX-WIRING-001` — owns `checkCodexWiring`, the doctor "Codex Wiring" check, and the
  statusline allowlist/default constants (AC-CW-012/013; `codex_readiness_test.go`,
  `statusline_test.go` lineage).
- `SPEC-CODEX-LAUNCHER-001` — owns AC-CL-013/014/016 and the `codexSpecFiles` guard file.
- `SPEC-CODEX-INIT-001` / `SPEC-INIT-HARNESS-PROMPT-001` — own the init wizard seam the axis-1
  test drives.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the correction-travels-with-
  record rule §F implements.
