# Acceptance — SPEC-CODEX-E2E-GUARD-001

Adoption discipline (two-cell, per `.claude/rules/moai/development/verification-completeness.md`
§2): every release-blocking AC below carries a **RED-now** cell adopted on the
pre-implementation tree — a single-invocation read-only command, its verbatim stdout, the exit
code as its own field, and the tree SHA — plus the **green-path** cell naming the milestone
that flips it. A criterion with one cell is unadopted. All RED-now observations below were
measured on this tree at `ace1c5440` before any implementation work.

## §D AC Matrix (Given-When-Then)

### Axis 1 — init→doctor end-to-end

**AC-CEG-001** (maps REQ-CEG-001; binding)
- **RED-now** (adopted at `ace1c5440`, pre-implementation):
  - Command: `go test ./internal/cli/ -run TestRunInit_ThenDoctorCodexWiringHealthy -count=1 -timeout 600s`
  - Verbatim stdout: `ok  	github.com/modu-ai/moai-adk/internal/cli	0.929s [no tests to run]`
  - Exit code: `0`
  - Reading: the selector matches no test — no init→doctor judgment exists yet. The gap is
    real, so the future green certifies a change rather than a vacuous pass.
- **Green-path** (flipped by plan.md M2):
  - **Given** a fresh temp project, `HOME` pinned to a temp dir, the wizard seam answering
    `AgentWiring: "codex"` with `MCPProvision: true`, and the hermetic stubs
    (`stubMoaiLookup(t, true)`, `stubCodexHome(t, t.TempDir())`)
  - **When** the real init path runs (`runInitForAutonomyAtHomeCapturingOut`) and the doctor
    judgment runs on the inited project (`checkCodexWiring(projectDir, false)`)
  - **Then** the check status is `CheckOK` with no codex finding text; the wiring artifacts
    exist (`assertCodexArtifacts(…, true)` as the anti-vacuous sanity leg)
- **Evidence**: verbatim GREEN output in progress.md §E.2, paired with the RED-now cell above.

**AC-CEG-002** (maps REQ-CEG-002; companion — descope rule in plan.md M2.2)
- **Given** the same setup with a claude-only wizard answer and `stubCodexLookup(t, true,
  false)` (codex binary absent)
- **When** the real init path runs and the doctor judgment runs on the inited project
- **Then** the check is a `CheckOK` informational outcome and carries no codex `Warn` finding
- **Evidence**: verbatim test output in progress.md §E.2; if descoped at kickoff, the recorded
  reason replaces the evidence.

### Axis 2 — mutant proof + premise correction

**AC-CEG-003** (maps REQ-CEG-004; binding)
- **RED-now** (self-RED disposition, recorded at `ace1c5440`): this criterion's red-now cell
  IS the mutant observation itself — the RED/GREEN pair does not exist until M1 executes it.
  The auditor deliberately did NOT execute the mutant on the audited tree, and it is not
  executed at plan phase either; plan.md M1 remains the run-phase step. The observable current
  state the work flips is the HALF-PAIRED baseline, adopted now:
  - Command: `go test ./internal/codexwiring/ -run TestStatusLineDefaultSubsetOfAllowlist -count=1`
  - Verbatim stdout: `ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.591s`
  - Exit code: `0`
  - Reading: the judgment is live and GREEN today, and no RED/GREEN pair is recorded anywhere
    (grep for the mutant record in progress.md returns nothing) — the pair's absence is the
    unflipped state M1 flips. The procedure is deterministic and re-executable, so the pair
    can be (re)adopted at any time.
- **Green-path** (flipped by plan.md M1):
  - **Given** a clean tree at the run-phase base (`git status --porcelain
    internal/codexwiring/configtoml.go` empty)
  - **When** one character of one `defaultStatusLine` token is mutated
    (`"git-branch"` → `"git-branchx"`) and
    `go test ./internal/codexwiring/ -run TestStatusLineDefaultSubsetOfAllowlist -count=1` runs,
    then the mutation is reverted and the same command re-runs
  - **Then** the first run is RED with the allowlist-violation message; the second run is
    GREEN; `git status --porcelain internal/codexwiring/configtoml.go` is empty after revert;
    the mutant appears in no commit
- **Evidence**: both verbatim outputs + the empty porcelain output in progress.md §E.2
  (`git log --all -- internal/codexwiring/configtoml.go` shows no mutant commit).

**AC-CEG-004** (maps REQ-CEG-003, REQ-CEG-008; binding at sync)
- **RED-now** (starting observation, adopted at `ace1c5440`): the correction record §F is
  already PRESENT at plan phase — the sync verdict's input exists; what remains unflipped is
  the sync-phase judgment over it:
  - Command: `grep -n "^## §F\|^### §F" .moai/specs/SPEC-CODEX-E2E-GUARD-001/spec.md`
  - Verbatim stdout:
    ```
    196:## §F Measurement-Correction Record
    198:### §F.1 Axis-2 premise refutation (card premise FALSE)
    213:### §F.2 The "41" figure — populations and sweep
    ```
  - Exit code: `0`
  - Reading: §F.1/§F.2 exist now; the sync-phase verdict (progress.md §E.4, currently a
    placeholder) is the state this AC's completion flips.
- **Green-path** (flipped at sync): **When** the sync-phase verdict reads spec.md §F.1 and
  §F.2, **Then** §F.1 carries the axis-2 refutation WITH its evidence command and observed
  output, and §F.2 carries the corrected populations (41 repo-wide / 38 cli / 45
  cli+codexwiring; the two fixture files as the only `func Test`-less files) and the "41"
  resolution.
- **Evidence**: spec.md §F present and lint-clean at sync.

### Axis 3 — guard expansion

**AC-CEG-005** (maps REQ-CEG-005; binding)
- **RED-now** (adopted at `ace1c5440`, pre-implementation): the build-tag/syscall guard sweeps
  2 files today, not 12:
  - Command: `grep -n "codexSpecFiles = " internal/cli/codex_launcher_guards_test.go`
  - Verbatim stdout: `30:var codexSpecFiles = []string{"codex_launcher.go", "codex_readiness.go"}`
  - Exit code: `0`
  - Reading: 2 of the 12 non-test codex files are guarded; the extension has not happened.
- **Green-path** (flipped by plan.md M3, self-describing): the expanded guard test opens with
  a swept-count assertion — `len(guardFiles) != 12` is a `t.Fatalf` ("vacuous green") — so a
  green byte-identical to today's 2-file `ok` cannot occur:
  - **When** `go test ./internal/cli/ -run TestCodexSpecFiles_NoBuildTagsOrSyscall -count=1
    -timeout 600s` runs over the 12-file set
  - **Then** GREEN — zero OS build tags, zero `"syscall"` imports, zero process-replacement
    identifiers, zero GOOS-suffixed files across all 12
- **Evidence**: verbatim GREEN output PLUS the swept-count proof in progress.md §E.2 — the
  `-v` run showing the count assertion passing, cross-checked against the independent count
  `ls internal/cli/*codex*.go | grep -v _test | wc -l` → `12` (exit 0, observed at plan phase).

**AC-CEG-006** (maps REQ-CEG-006; binding)
- **RED-now** (adopted at `ace1c5440`, pre-implementation): the exec guard knows exactly one
  expected first argument today — `req.Program` — over the same 2-file set (no `binaryPath`
  rows, no `"git"` row, no zero-call rows for the other 8):
  - Command: `grep -n "req.Program" internal/cli/codex_launcher_guards_test.go`
  - Verbatim stdout:
    ```
    147:// argument (here: req.Program, the launch request's resolved binary). A
    170:			if first != "req.Program" {
    171:				t.Errorf("%s: process-start first argument %q is not the codex path variable (req.Program)", name, first)
    ```
  - Exit code: `0`
  - Reading: one expectation, two files — the per-file table of spec.md §A does not exist.
- **Green-path** (flipped by plan.md M3, self-describing): the per-file expected-first-
  argument table (`req.Program` / `binaryPath` / `"git"` / zero-call for the other 8) enforced
  over the 12-file set with the same `len(guardFiles) != 12` swept-count assertion:
  - **When** `go test ./internal/cli/ -run TestCodexSpecFiles_ExecPrimitivesCodexOnly -count=1
    -timeout 600s` runs
  - **Then** GREEN on the current tree; the comment-only-line skip is preserved; the previous
    2-file assertions hold inside the wider form (no guard weakened)
- **Evidence**: verbatim GREEN output + swept-count proof in progress.md §E.2 (same shape as
  AC-CEG-005).

**AC-CEG-007** (maps REQ-CEG-007, REQ-CEG-009, REQ-CEG-010; binding)
- **RED-now** (adopted at `ace1c5440`, pre-implementation): the neutrality judgment walks the
  cobra command surface ONLY — no string-literal scan of any file set exists:
  - Command: `grep -n "walk(codexCmd)" internal/cli/codex_launcher_guards_test.go`
  - Verbatim stdout: `74:	walk(codexCmd)`
  - Exit code: `0`
  - Reading: the single walk call is the whole neutrality surface; the 12-file literal scan of
    plan.md M4 does not exist.
- **Green-path** (flipped by plan.md M4, self-describing): the literal-scan guard iterates the
  same 12-file set with the `len != 12` swept-count assertion AND a positive-control canary —
  the scan must observe at least one known existing literal (`"git"` at
  `codex_review_gate.go:129`), so a scanning failure reads RED instead of a vacuous green:
  - **When** the neutrality guard tests run (`-run TestCodexCommand` family)
  - **Then** GREEN; any baseline-literal violation found at M4 is handled by a recorded
    narrowing/allowlist justification (in plan.md M4 + progress.md), and the existing
    `TestCodexCommand_NeutralityScan` launcher scan is unchanged; the guard comment states the
    AC-CL-007 (sentinel-value) vs literal-neutrality surface distinction without double-claiming
- **Evidence**: verbatim GREEN output + swept-count proof + the canary observation in
  progress.md §E.2.

## §D.1 Severity

| AC | Severity | Rationale |
|---|---|---|
| AC-CEG-001 | must-pass | The card's headline: the init→doctor chain judgment |
| AC-CEG-002 | should-pass | Companion negative; descope allowed with recorded reason |
| AC-CEG-003 | must-pass | Proves an existing guard is not vacuous; the axis-2 deliverable; self-RED disposition recorded above |
| AC-CEG-004 | must-pass | The correction record is the deliverable for a refuted premise |
| AC-CEG-005 | must-pass | Mechanical extension, measured GREEN at plan phase, self-describing green |
| AC-CEG-006 | must-pass | The per-file table is the design centerpiece of axis 3 |
| AC-CEG-007 | must-pass | Design-heaviest; baseline-gated so GREEN is achievable honestly |

## §D.2 Edge cases

1. The axis-1 test runs the FULL init flow — a legit doctor finding unrelated to wiring
   (e.g. statusline sub-check) could flip the verdict; if so, the finding is reported (spec.md
   §G), and the assertion narrows to "no codex-WIRING finding" only with the reason recorded.
2. `codex_review_gate.go`'s `"git"` first argument must not be pattern-matched by a
   codex-only rule; the per-file table exists precisely for this.
3. Comment lines mentioning `exec.Command` (`codex_launcher.go:136`, `mcp_codex.go:339`) must
   stay skipped — removing the skip false-REDs on documentation.
4. The two fixture files (`codex_contract_fixture_{unix,windows}_test.go`) have `func Test`-less
   bodies BY DESIGN; no coverage debt is filed against them (spec.md §F.2).
5. The mutant's revert must restore the exact original bytes — a "close enough" revert fails
   M1.4's porcelain check.
6. The swept-count assertion (`len == 12`) must track the true population: if a future codex
   CLI file lands, the assertion REDs and forces a deliberate table update — that friction is
   the guard working, never a reason to loosen the count.

## §D.3 Definition of Done

1. All must-pass ACs GREEN with verbatim evidence in progress.md §E.2; AC-CEG-002 either GREEN
   or descoped with recorded reason at kickoff.
2. Every must-pass AC's RED-now cell paired with its green evidence — a green whose RED-now
   cell is missing is an unadopted criterion, not a pass.
3. Guard greens are self-describing: the `len(guardFiles) == 12` swept-count assertion and the
   M4 canary pass, and progress.md §E.2 records the swept count actually observed.
4. Scoped verification batch GREEN: `go test ./internal/cli/... -count=1 -timeout 600s`,
   `go test ./internal/codexwiring/... -count=1`, `golangci-lint run` on both packages.
5. `git status --porcelain internal/codexwiring/configtoml.go` empty (mutant never committed).
6. `moai spec lint SPEC-CODEX-E2E-GUARD-001` → 0 findings at sync.
7. No existing guard test weakened (2-file baseline assertions hold inside the 12-file form).
