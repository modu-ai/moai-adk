---
id: SPEC-CODEX-E2E-GUARD-001
title: "Codex init-to-doctor end-to-end verdict, launcher guard coverage expansion, and statusline mutant proof"
version: "0.1.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/cli,internal/codexwiring"
lifecycle: spec-anchored
tags: "codex, testing, doctor, guards, statusline, t500"
tier: M
---

# SPEC-CODEX-E2E-GUARD-001 — Codex init→doctor e2e verdict + launcher guard expansion + statusline mutant proof

## §A Context and Base

Factory card t500 (C7 · test card): "codex init→doctor 종단 판정과 statusline 기본값 불변이 미검증".
Three axes plus a side-note. The card's own numbers came from an investigation agent; the lane
re-measured at the plan base, and this SPEC re-measured the load-bearing facts again on its own
tree (every figure below carries its own attribution).

- **Base SHA (plan-phase)**: `ace1c5440` (branch `WT-codex-e2e-guard`, worktree
  `.claude/worktrees/t500`). Run-phase measurements re-pin the SHA they were taken at.
- **Attribution shorthand**: "lane re-measured at ace1c5440" marks figures the lane measured and
  this plan phase verified by reading the cited source; "measured this tree" marks figures this
  plan phase produced by running the command itself (command + output in §F).

### Axis 1 — init→doctor end-to-end gap (CONFIRMED)

- `internal/cli/doctor_codex_test.go:26-34` — `wireProjectForDoctor()` calls
  `codexwiring.Wire(root, …)` DIRECTLY (`:30`); the doctor "Codex Wiring" check is never
  exercised through the real init command path.
- `internal/cli/init_agent_wizard_test.go` — DOES run the real init path via the existing helper
  `runInitForAutonomyAtHomeCapturingOut` (defined `internal/cli/init_autonomy_wiring_test.go:40`),
  but asserts only artifact existence (`assertCodexArtifacts`, `:42-53`) plus MCP-announcement
  strings. No test runs init and then a doctor judgment on the result.
- Doctor-side seams already exist for hermetic use: `stubMoaiLookup` (`doctor_codex_test.go:38`),
  `stubCodexLookup` (`:53`), `stubCodexHome` (`:76`), plus `checkCodexWiring(root, verbose)` and
  `runGroupedChecks(false, "Codex Wiring")` (registration proven by
  `TestDoctor_CodexWiringRegistered`, `:688`).

### Axis 2 — statusline default⊆allowlist invariant: CARD PREMISE REFUTED

The card claimed no test asserts `configtoml.go`'s default 5 tokens ⊆ the 29-token allowlist.
**The premise is false on this tree** and is recorded as a first-class correction (§F.1), not a
footnote:

- `internal/codexwiring/statusline_test.go` carries exactly the claimed-missing assertion —
  `TestStatusLineDefaultSubsetOfAllowlist` (`:47`, AC-CW-013b) — plus
  `TestStatusLineDefaultIsCanonicalFive` (`:31`), `TestStatusLineDefaultHasNoParseAliases`
  (`:64`), `TestStatusLineAllowlistMatchesCanonicalSet` (`:79`).
- Source under test: `internal/codexwiring/configtoml.go` — `StatusLineAllowlist` (29 tokens,
  `:32-40`), `defaultStatusLine` (5 tokens, `:45-47`), `DefaultStatusLine()` copy accessor
  (`:50-54`).
- Baseline: `go test ./internal/codexwiring/ -run TestStatusLine -count=1` → `ok … 0.617s`
  (measured this tree, at `ace1c5440`).
- Consequence: axis 2's deliverable is NOT a new test. It is (a) a run-phase **mutant proof**
  that the existing machine judgment bites (REQ-CEG-004), and (b) this correction record.

### Axis 3 — launcher guard blind spot (CONFIRMED)

- `internal/cli/codex_launcher_guards_test.go:30` — `codexSpecFiles =
  []string{"codex_launcher.go", "codex_readiness.go"}`. Two guards iterate this 2-file set:
  AC-CL-014 (`:119`, build tags / syscall / process-replacement / GOOS-suffix) and AC-CL-016
  (`:150`, every process-start primitive hands the codex path variable as first arg). AC-CL-013
  (`:82`) does not iterate the set — it walks the `codexCmd` cobra surface, so it covers command
  strings only.
- `internal/cli` carries exactly 12 non-test codex files: `codex_contract.go`, `codex_init.go`,
  `codex_job_control.go`, `codex_jobs.go`, `codex_launcher.go`, `codex_readiness.go`,
  `codex_review_gate.go`, `codex_task.go`, `doctor_codex.go`, `hook_harness_codex.go`,
  `mcp_codex.go`, `update_codex_wiring.go` (measured this tree) — the 10 beyond
  `codexSpecFiles` are uncovered by all three guards.
- Exec call sites, read per file (not designed from the armchair):

  | File | Site | First argument | Legitimate value |
  |---|---|---|---|
  | `codex_launcher.go` | `:457` `exec.Command` | `req.Program` | launch request's resolved codex binary |
  | `mcp_codex.go` | `:350` `exec.CommandContext` (`realCodexRunner.run`) | `binaryPath` | codex binary handed to the `codexCommandRunner` seam |
  | `mcp_codex.go` | `:433` `exec.CommandContext` (`realCodexSessionRunner.start`) | `binaryPath` | codex binary handed to the `codexSessionRunner` seam |
  | `mcp_codex.go` | `:1889` `exec.CommandContext` (`defaultLoginStatusRunner`) | `binaryPath` | codex binary param, args `"login", "status"` |
  | `codex_review_gate.go` | `:129` `exec.Command` | `"git"` | git literal — the porcelain probe is a legitimate non-codex primitive |
  | other 8 files | — | — | zero process-start call sites |

  → AC-CL-016 CANNOT blanket-extend with `req.Program`; it needs the per-file table above.
- Lane measured on the 10 uncovered files: zero `//go:build` lines, zero `"syscall"` imports —
  AC-CL-014 extends cleanly to all 12 and stays GREEN.
- Reconciliation with AC-CL-007: the comment at `codex_launcher_guards_test.go:27-29` says
  mcp_codex.go's diff parts are covered by the AC-CL-007 closed-set judgment elsewhere.
  AC-CL-007 covers SENTINEL VALUE classification (`codex_readiness_test.go:26` — sentinel
  values distinct from real values), not process-start first arguments. Extending AC-CL-016 to
  `mcp_codex.go` therefore does not double-claim that surface; the SPEC records the distinction.
- Guard conventions to preserve: comment-only lines skipped in the exec scan
  (`codex_launcher_guards_test.go:156-160` — doc comments legitimately mention primitives),
  `t.Helper()` fixture helpers, ordered-phrase assertion style.

### Side-note — the "41 files" figure (MEASURED + SWEPT, partly corrected)

- `codex_contract_fixture_unix_test.go` (62 lines) and `codex_contract_fixture_windows_test.go`
  (18 lines) contain ZERO `func Test` (measured this tree). They are build-tag fixture helpers
  (`//go:build !windows` FIFO/socket fixtures; the windows twin returns
  `errCodexFixtureUnsupported`). Across the codex test surface they are the ONLY files without
  any `func Test`.
- The figure **41 is not a phantom**: `find internal -name '*codex*_test.go' | wc -l` → `41`
  on this tree — the repo-wide codex-named test-file population (38 `internal/cli` + 3
  elsewhere), which is exactly the filename-axis definition `SPEC-CODEX-E2E-MEASURE-001`
  spec.md:40 measured at its own base. The lane's "NONE equals 41" finding held only for its own
  two population sets (cli-only 38/36; cli+codexwiring 45/43). §F.2 records the full
  reconciliation; no repo edit is required for the number.

## §B Definitions

- **Real init path** — `runInitForAutonomyAtHomeCapturingOut` (drives `runInit` through the
  cobra command with the wizard seam swapped), as opposed to calling `codexwiring.Wire` directly.
- **Doctor judgment** — `checkCodexWiring(root, verbose)` (or the same check reached through
  `runGroupedChecks(false, "Codex Wiring")`), returning a `DiagnosticCheck` with a
  `uikit.Check*` status.
- **Hermetic** — the test requires neither a real `codex` binary on PATH nor network access;
  all lookups route through the existing seams (`codexWiringLookPath`, `codexUserHomeDir`,
  `CODEX_HOME` env, `runWizardFn`).
- **Mutant proof** — a transient one-character source mutation demonstrating a guard's RED,
  immediately reverted; evidence is the pair of observed outputs, never a committed diff.
- **The 12-file set** — the 12 non-test codex CLI files listed in §A axis 3.

## §C Requirements (GEARS)

- **REQ-CEG-001** (axis 1) — **When** `moai init` completes on a fresh project with the wizard
  answering codex, the doctor "Codex Wiring" check shall report the healthy verdict (`CheckOK`)
  for that inited project, evaluated under the hermetic stubs (`codexWiringLookPath` resolving,
  fresh `CODEX_HOME`), exercising the REAL init command path (`runInitForAutonomyAtHomeCapturingOut`)
  rather than a direct `codexwiring.Wire` call.
- **REQ-CEG-002** (axis 1, negative companion) — **When** `moai init` completes with a
  claude-only wiring selection, the doctor "Codex Wiring" check shall stay silent about codex
  wiring findings for the inited project (a `CheckOK` informational outcome, never a codex
  `Warn`). Companion case; descope handling in plan.md M2.
- **REQ-CEG-003** (axis 2) — The SPEC record shall carry the axis-2 premise refutation (the
  default⊆allowlist assertion already exists as `TestStatusLineDefaultSubsetOfAllowlist`) as a
  first-class measurement-correction section including the evidence commands (§F.1).
- **REQ-CEG-004** (axis 2) — **While** collecting axis-2 evidence, the run phase shall
  demonstrate the existing machine judgment bites: a one-character mutation of one
  `defaultStatusLine` token (e.g. `"git-branch"` → `"git-branchx"`) turns
  `TestStatusLineDefaultSubsetOfAllowlist` RED, and the revert restores GREEN; the mutant shall
  never be committed.
- **REQ-CEG-005** (axis 3) — The AC-CL-014 build-tag/syscall guard shall scope to the 12-file
  set and shall stay GREEN on the current tree (zero OS build tags, zero `"syscall"` imports,
  zero process-replacement identifiers, zero GOOS-suffixed files).
- **REQ-CEG-006** (axis 3) — The AC-CL-016 process-start guard shall enforce, across the
  12-file set, the per-file expected-first-argument table measured in §A (req.Program /
  binaryPath / "git" / zero call sites); it shall preserve the comment-only-line skip and shall
  not weaken any existing guard assertion.
- **REQ-CEG-007** (axis 3) — The AC-CL-013 neutrality judgment shall extend beyond the
  cobra-command walk to the string-literal surface (comments excluded) of the 12-file set for
  the same forbidden-pattern classes, with a swept-count assertion (`len == 12`) and a
  positive-control canary literal so the green is self-describing.
- **REQ-CEG-008** (side-note) — The SPEC record shall carry the corrected codex test-file
  populations and the sweep result for the "41" figure (§F.2), including the finding that 41
  denotes the repo-wide codex-named test-file population at the plan-phase base.
- **REQ-CEG-009** (axis 3) — **Where** a baseline literal of the 12-file set already violates
  a neutrality pattern class, the neutrality guard shall narrow or allowlist that class with
  the justification recorded in this SPEC, and shall never weaken the existing
  `TestCodexCommand_NeutralityScan` launcher scan.
- **REQ-CEG-010** (axis 3) — The neutrality extension shall reconcile with the AC-CL-007
  closed-set judgment — AC-CL-007 owns sentinel-VALUE classification and this guard owns
  string-literal neutrality — and the two judgments shall not double-claim each other's
  surface.

## §D Constraints

1. **Verification scope**: touched packages only — `go test ./internal/cli/...` and
   `go test ./internal/codexwiring/...`. NEVER `go test ./...` locally (10 parallel lanes share
   this machine; 2026-08-15 load-413 precedent). Full-suite verdict is CI's.
2. `internal/cli` test invocations carry a `-timeout 600s` floor.
3. `t.TempDir()` isolation everywhere; no `t.Setenv("OTEL_*")`; `HOME` pinned via `t.Setenv`
   per the existing init-test convention (`init_agent_wizard_test.go:67-69`).
4. The axis-1 e2e test MUST NOT require a real codex binary or network — hermetic seams only.
5. The axis-2 mutant is transient evidence collection: apply → observe RED → revert → observe
   GREEN → confirm the file is clean; never staged, never committed.
6. No guard may be weakened to make an extension pass; the 2-file baseline assertions survive
   inside the 12-file form.

## §E Acceptance

AC→REQ coverage (Given-When-Then scenarios in `acceptance.md` §D):

| AC | Verifies | Subject matter | Severity |
|---|---|---|---|
| AC-CEG-001 | REQ-CEG-001 | init(codex)→doctor healthy verdict, real init path | must-pass |
| AC-CEG-002 | REQ-CEG-002 | init(claude-only)→doctor stays silent | should-pass |
| AC-CEG-003 | REQ-CEG-004 | statusline mutant RED→revert GREEN, never committed | must-pass |
| AC-CEG-004 | REQ-CEG-003, REQ-CEG-008 | §F correction record complete at sync | must-pass |
| AC-CEG-005 | REQ-CEG-005 | build-tag/syscall guard over the 12-file set | must-pass |
| AC-CEG-006 | REQ-CEG-006 | exec guard per-file expected-first-arg table | must-pass |
| AC-CEG-007 | REQ-CEG-007, REQ-CEG-009, REQ-CEG-010 | neutrality extension over string literals | must-pass |

Quality gate: scoped-package test GREEN + `golangci-lint run` on touched packages +
`moai spec lint` 0 findings.

## §F Measurement-Correction Record

### §F.1 Axis-2 premise refutation (card premise FALSE)

- **Refuted claim**: "no test asserts configtoml.go's default 5 tokens ⊆ the 29-token
  allowlist; statusline_test.go covers only the 3-branch merge rule."
- **Observed state** (lane re-measured at ace1c5440; source re-read this tree):
  `internal/codexwiring/statusline_test.go:47` — `TestStatusLineDefaultSubsetOfAllowlist` is
  exactly that assertion (AC-CW-013b), and the merge-rule coverage claim is also false —
  `statusline_test.go` carries the four default/allowlist tests (`:31`, `:47`, `:64`, `:79`)
  plus six merge-branch tests.
- **Evidence command** (run this tree at `ace1c5440`):
  `go test ./internal/codexwiring/ -run TestStatusLine -count=1` →
  `ok  github.com/modu-ai/moai-adk/internal/codexwiring  0.617s`
- **Consequence**: the axis-2 deliverable is the mutant proof (REQ-CEG-004) proving the existing
  judgment is not vacuous, plus this record. No new statusline test is authored.

### §F.2 The "41" figure — populations and sweep

Populations measured this tree at `ace1c5440` (commands in HISTORY):

| Population | Count | Files without `func Test` |
|---|---|---|
| Repo-wide codex-named `*_test.go` (`find internal -name '*codex*_test.go'`) | **41** | 2 |
| `internal/cli` codex-named test files | 38 | 2 |
| `internal/cli` codex-named with tests | 36 | — |
| `internal/cli` + `internal/codexwiring` test files | 45 | 2 |
| same, with tests | 43 | — |

- The only test files without any `func Test` are `codex_contract_fixture_unix_test.go` (62
  lines) and `codex_contract_fixture_windows_test.go` (18 lines) — build-tag fixture helpers by
  design, not coverage debt.
- **41 is a live population**: the repo-wide codex-named test-file count, identical to the
  filename-axis figure `SPEC-CODEX-E2E-MEASURE-001` spec.md:40 measured at `e9c6a8564`. The
  lane's "matches no current population" finding was true only within its own two population
  sets; this record supersedes it.
- **Sweep**: the phrasing "41본" is cited nowhere in the tracked repo (code, README*, docs-site,
  CHANGELOG) nor under `.moai/reports/` (lane sweep; spot-checked this tree). The tracked
  citation of the count 41 is the `SPEC-CODEX-E2E-MEASURE-001` inventory table itself, where it
  is correct for its own definition. No repo edit is needed for the number; this section is the
  record.

## §G Out of Scope

### Out of Scope — New statusline tests

- No new test is added to `internal/codexwiring`; the four existing `TestStatusLine*` guard
  tests already carry the invariant (§F.1). The axis-2 deliverable is the mutant proof + record.

### Out of Scope — Full-suite local verification

- `go test ./...` stays out of local scope; CI owns the full-suite verdict (§D.1).

### Out of Scope — Wiring repair and doctor behavior change

- This SPEC adds judgment coverage, not behavior: `checkCodexWiring`, `codexwiring.Wire`, and
  the init flow are observed, not modified. A RED on REQ-CEG-001 is a discovered defect reported
  to the orchestrator, never silently repaired here.

### Out of Scope — Adjacent-card work

- t462's measurement record (`SPEC-CODEX-E2E-MEASURE-001`), the codex skill-axis work, and any
  e2e journey-script authoring stay with their own cards.

### Out of Scope — Repo edits for the "41" figure

- The count is correct for its defining population (§F.2); no doc, README, or SPEC sweep edit
  is performed for it.

## HISTORY

| Date | Author | Change |
|---|---|---|
| 2026-09-07 | manager-spec | Plan-phase authoring at `ace1c5440` (card t500). Plan-phase measurements (this tree): `find internal -name '*codex*_test.go' \| wc -l` → `41`; `ls internal/cli/*codex*_test.go \| wc -l` → `38`; `ls internal/codexwiring/*_test.go \| wc -l` → `7`; `grep -L "^func Test"` over the 41 → exactly the two fixture files; `wc -l` fixtures → 62 / 18; `go test ./internal/codexwiring/ -run TestStatusLine -count=1` → `ok 0.617s`; `go test ./internal/cli/ -run 'TestCodexSpecFiles\|TestCodexCommand_NeutralityScan\|TestCodexSpawn_TmuxDiagnosticSingleSource' -count=1 -timeout 600s` → `ok 0.929s`; exec call sites read at `codex_launcher.go:457`, `mcp_codex.go:350/433/1889`, `codex_review_gate.go:129`. Axis-2 card premise recorded REFUTED (§F.1); "41" recorded as the repo-wide population (§F.2), superseding the lane's no-population finding. |
| 2026-09-07 | manager-spec | Plan-audit iter-1 repair (FAIL 0.94 → D1-D4; report `.moai/reports/t500/plan-audit-iter1.md`). D1: RED-now adoption cells added to all six must-pass ACs in acceptance.md (commands + verbatim stdout + exit code + SHA `ace1c5440`, measured this tree pre-implementation; AC-CEG-003 recorded self-RED, mutant NOT executed at plan phase); vacuous-green on AC-CEG-005/006/007 fixed via `len(guardFiles) == 12` swept-count assertion + M4 positive-control canary in plan.md M3/M4 and matching evidence cells in acceptance.md §D.3. D2: compound REQ-CEG-007 split into REQ-CEG-007/009/010 (one pattern per REQ). D3: undeclared `related_specs` frontmatter field removed (cross-refs live in plan.md §H). D4: "an CheckOK" → "a CheckOK" (acceptance.md AC-CEG-002 + spec.md REQ-CEG-002). Spec lint re-run → 0 findings. |
