# Dogfooding Validation Report — SPEC-V3R6-HARNESS-V4-001 M6 Part B

> **5-Section Evidence-Bearing Report Format** per
> `.claude/rules/moai/core/verification-claim-integrity.md` § The 5-Section
> Evidence-Bearing Report Format. This report satisfies AC-HV4-012a (harness
> built + runs task) and AC-HV4-012b (A/B disclosed as author-measured).
>
> **Disclosure header (AC-HV4-012b)**:
> - **Sample size (n)**: 1 (single component-level end-to-end dry-run of the
>   sample "moai-adk-dev" manifest). No with/without A/B live comparison was
>   executed (see §Gaps).
> - **Measurement posture**: **author-measured, 3rd-party replication pending**.
>   The author (manager-develop, this delegation) ran the verification commands
>   and observed the outputs. An independent third party has NOT yet replicated
>   the validation.
> - **revfactory +60% provenance**: the +60% productivity claim originates from
>   `github.com/revfactory/harness` (Apache-2.0) upstream analysis, NOT from
>   this v4 validation. It is cited as **upstream provenance** (the design
>   motivation), NOT co-opted as a verified v4 claim. This report makes NO
>   +60% claim for v4.

---

## §1. Claim (주장)

The v4 harness pipeline — the composition of the M3 (manifest schema + Validate
+ RunnerTemplate + DecideEvaluator), M4 (GenerateCommand + lifecycle), and M5
(DecideIsolation + EmitCleanupDirective) components — is validated end-to-end
at the **component level** on a sample "moai-adk-dev" manifest.

Specifically, the sample harness (name `moai-adk-dev`, 2 specialists, 2
patterns, 3-dimensional Sprint Contract) is built end-to-end via the real
components and the dispatch path runs: the manifest passes `Validate()`, each
specialist's `isolation` is cross-validated by `DecideIsolation()`, the
thin-wrapper command is emitted by `GenerateCommand()` referencing
`harness-moai-adk-dev-run.js`, the evaluator decision is conditional via
`DecideEvaluator()`, and the Runner template dispatch switch covers each
declared primitive verbatim.

This satisfies **AC-HV4-012a** at the component level (the sample harness is
built + the dispatch path runs). The full live-task execution is a Gap (§4).

## §2. Evidence (증거)

The evidence is the verbatim command + output of the component-level
end-to-end integration test
(`internal/harness/v4manifest/dogfood_test.go`, added at M6 Part B).

**Command**:

```
go test -run 'TestDogfood_MoaiAdkDevHarnessEndToEnd|TestDogfood_EmitCleanupDirectiveFiresForWorktreeSpecialist' -v ./internal/harness/v4manifest/...
```

**Verbatim output**:

```
=== RUN   TestDogfood_MoaiAdkDevHarnessEndToEnd
--- PASS: TestDogfood_MoaiAdkDevHarnessEndToEnd (0.00s)
=== RUN   TestDogfood_EmitCleanupDirectiveFiresForWorktreeSpecialist
--- PASS: TestDogfood_EmitCleanupDirectiveFiresForWorktreeSpecialist (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness/v4manifest	(cached)
```

The test exercises 6 steps of the v4 path on the sample manifest:

1. **Validate** — `Validate(manifest)` returns nil (the sample satisfies the
   canonical 8-field + 5-sub-field-per-specialist schema).
2. **DecideIsolation per specialist** — `DecideIsolation()` is called for each
   of the 2 specialists; `template-neutrality-auditor` (read-only) →
   `isolation=none` (AC-HV4-007a); `cli-codemaps-extractor` (risky) →
   `isolation=worktree` (REQ-HV4-007). Both decisions carry non-empty rationale.
3. **GenerateCommand** — `GenerateCommand(manifest)` emits the thin-wrapper
   markdown referencing `harness-moai-adk-dev-run.js` and `manifest.json`.
4. **DecideEvaluator** — `DecideEvaluator()` returns `Invoked=false` for a
   within-solo-range task (skip + rationale) and `Invoked=true` for a
   exceeds-solo-range task (dimensions echoed).
5. **RunnerTemplate dispatch switch** — the stamped Runner template body
   contains a `case "<primitive>":` for each primitive the sample declares
   (`sub-agent`, `worktree`); `manifest.json` is the single config-read path.
6. **Round-trip** — manifest.json + command + runner are written to
   `t.TempDir()`, re-read, and re-`Validate()`d; the round-tripped manifest is
   field-identical to the original on Name / RunnerWorkflow / specialist count.

**Complementary evidence — revfactory residual absence (AC-HV4-009a)**:

```
$ grep -rnE '7-Phase|Phase 7 LEARNING|Skeleton|Customization' \
    internal/harness/v4manifest/*.go internal/cli/harness/*.go | grep -v "_test.go"
exit=1
```

Exit code 1 = zero matches (the grep found nothing). The NEW v4 Go artifacts
carry ZERO revfactory 7-Phase residuals.

## §3. Baseline-attribution (baseline 귀속)

The evidence above was measured against:

- **Baseline commit**: `13d442ce9` (`origin/main`, synced 0 0). This is the M6
  Part A HEAD; Part B builds on it.
- **Components under test**: M3 (`internal/harness/v4manifest/{types,schema,
  validate,runner_template,sprint_contract}.go`), M4 (`command_template.go`
  + `internal/cli/harness/v4lifecycle*.go`), M5 (`isolation.go`).
- **Test surface**: `internal/harness/v4manifest/dogfood_test.go` (NEW at Part
  B) + the full v4manifest suite (19 M3-M5 tests + 1 M6 Part A migrated-
  specialists test + 2 Part B dogfood tests = 22 tests).
- **Coverage observed** (this run, against this tree):
  ```
  $ go test -cover ./internal/harness/v4manifest/...
  ok  github.com/modu-ai/moai-adk/internal/harness/v4manifest  coverage: 100.0% of statements
  ```
- **Lint observed** (this run):
  ```
  $ golangci-lint run --timeout=2m ./internal/harness/v4manifest/...
  0 issues.
  ```
- **Cross-platform build observed** (this run):
  ```
  $ go build ./...                        → exit 0
  $ GOOS=windows GOARCH=amd64 go build ./... → exit 0
  ```

These numbers are from THIS run against THIS tree — not carried over from a
prior unrelated measurement (per verification-claim integrity §2).

## §4. Gaps (미검증)

The following were explicitly **NOT** observed in this validation. They are
disclosed as Gaps, NOT claimed as passes (per AC-HV4-012b's author-measured
disclosure clause):

1. **Full live orchestrator-driven `/moai:harness` build NOT executed.** The
   v4 Builder is orchestrator-direct (M2 pivot: the Builder is orchestrator-
   side logic, NOT a dynamic-workflow script and NOT a separate agent). A
   subagent (this delegation) **cannot drive the orchestrator** — the Builder
   requires the orchestrator session to run Context-First Discovery →
   `Agent(Explore)` ANALYZE fan-out → `Agent(opus,xhigh)` PLAN →
   `AskUserQuestion` PLAN→GENERATE approval gate → `Agent()` GENERATE fan-out
   → `/goal` ACTIVATE. The component integration test in §2 exercises the
   Go components the Builder would invoke, but it does NOT prove a live
   orchestrator-driven build produces a coherent harness on a real codebase.

2. **Real-task with/without A/B comparison NOT executed.** AC-HV4-012 /
   AC-HV4-012b specify a with/without A/B comparison (harness vs no-harness
   baseline). This validation does NOT run a real moai-adk development task
   (e.g. "add a new hook event handler" or "extend the template-neutrality CI
   guard") both with and without the v4 harness, and does NOT measure a
   productivity delta. The sample size (n) for the A/B is therefore **0**, not
   the n=1 of the component test.

3. **3rd-party replication NOT obtained.** The evidence in §2 was observed by
   the author (manager-develop) only. An independent reviewer has not yet
   re-run the commands against an independent checkout.

4. **revfactory +60% NOT re-measured for v4.** The +60% figure is upstream
   provenance from `github.com/revfactory/harness`; it is NOT a verified v4
   claim and this report does not assert it for v4. Measuring a v4-specific
   productivity delta requires the live A/B (Gap #2) and is deferred.

5. **Pre-existing baseline failure (orthogonal, recorded for honesty).** The
   full repo suite (`go test ./...`) has one pre-existing failure unrelated to
   this SPEC: `internal/skills` `TestEntryRouterLOCCeiling` fails because
   `run.md` is 246 LOC (ceiling 200). The `run.md` file was last modified by
   `2ad42d830` (SPEC-V3R6-ORCH-IGGDA-001 M3 — a DIFFERENT SPEC's parallel
   workstream) and is 246 LOC on `origin/main` HEAD `13d442ce9`. This failure
   is NOT caused by M6 Part B (Part B added only `dogfood_test.go`; the
   `run.md` LOC violation predates it). It is recorded here so a reader of
   this report is not surprised by the `go test ./...` exit code; fixing it
   is out of scope for SPEC-V3R6-HARNESS-V4-001 and belongs to the
   ORCH-IGGDA-001 workstream (or a follow-up LOC-reduction commit).

## §5. Residual-risk (잔여 위험)

The risk that survives even after the §2 evidence:

- **Component composition ≠ system coherence.** The integration test verifies
  the dispatch/validate/generate/isolation/evaluator PLUMBING composes. It
  does NOT prove that a live orchestrator-driven build produces a harness
  whose specialists actually do useful moai-adk development work. The plumbing
  can be correct while the specialist prompts / pattern selection / Sprint
  Contract thresholds are miscalibrated for real tasks. Closing this residual
  risk requires the live A/B (Gap #2).

- **Determinism of the Runner script body.** The Runner template body is
  verified deterministic (no `Date.now()` / `Math.random()` — M3
  `TestRunnerTemplate_DeterministicScriptBody`), but the component test does
  not execute the Runner under the dynamic-workflow runtime. A resume-cache
  break from a non-deterministic runtime injection is a residual risk until a
  live Runner invocation is observed.

- **Sprint Contract threshold calibration.** The sample manifest's thresholds
  (`correctness: 0.95`, `template-neutrality: 1.0`, `coverage: 0.85`) are
  plausible defaults, NOT empirically calibrated. Whether these thresholds
  produce useful evaluator verdicts on real tasks is a residual risk that
  only the live A/B can address.

- **Conditional-evaluator solo-range boundary.** `DecideEvaluator` skips the
  evaluator when `WithinSoloRange=true`. The boundary between "within solo
  range" and "exceeds solo range" is a PLAN-phase judgment, not a mechanical
  threshold. A task mis-classified as within-solo-range would skip needed
  adversarial verification. This is a known design trade-off (C-HV4-001
  simplest-solution-first), not a defect, but it is a residual risk for
  tasks near the boundary.

---

## §A. AC Mapping

| AC ID | Status | Evidence |
|-------|--------|----------|
| AC-HV4-012a (harness built + runs task) | **PASS (component level)** | §2 — the sample "moai-adk-dev" harness is built end-to-end via the real M3/M4/M5 components and the dispatch path runs. Full live task = Gap §4.1. |
| AC-HV4-012b (A/B author-measured) | **PASS (disclosure)** | This report IS the 5-Section disclosure. Sample size n=1 (component test), n=0 (live A/B). Author-measured, 3rd-party pending. revfactory +60% cited as provenance, NOT verified v4 claim. |

## §B. Artifacts

- **Test**: `internal/harness/v4manifest/dogfood_test.go` (NEW at M6 Part B).
- **Report**: this file (`.moai/specs/SPEC-V3R6-HARNESS-V4-001/dogfooding-report.md`).
- **Pre-task evidence backfill**: `progress.md §E.2` M6 Part A subsection
  (commit `403369564`).

## §C. Verification-Claim Integrity Self-Check

- [x] Every PASS row in §A maps to a verbatim command + output in §2 (§1.1
      surface 2: manager-agent completion report).
- [x] Baseline-attribution names the commit (`13d442ce9`) + the tree (§3).
- [x] Gaps are explicit — the live build + real-task A/B + 3rd-party
      replication + revfactory re-measurement are all named as NOT observed
      (§4).
- [x] Residual-risk distinguishes "what was not observed" (§4 Gaps) from "what
      could still be wrong despite what WAS observed" (§5).
- [x] The pre-existing `internal/skills` failure is recorded honestly (§4.5)
      rather than hidden or falsely attributed to this SPEC.

---

Version: 1.0.0 (M6 Part B dogfooding validation — component-level end-to-end)
Classification: SPEC research artifact (NOT distributed — may carry SPEC IDs)
Origin: SPEC-V3R6-HARNESS-V4-001 M6 Part B (REQ-HV4-012 / AC-HV4-012a/012b)
