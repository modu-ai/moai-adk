# t444 — 10 inherited reds intake: 1 repaired, 9 environmental (verdict)

Card t444 (Class B, run → sync), lane-7, branch `WT-doctor-freshness-reds`,
base `304bc8158` (develop tip at dispatch), fix commit `c26cd7cc0`.
Harness classification: **minimal** (one test-fixture file).

## Claim

Of the 10 inherited reds lane-7 surfaced during the 2026-09-02 integration
window, exactly **1** reproduces as a code defect at `304bc8158` and is
repaired here (`TestGateGraphFreshness_AllLayersFreshNotice`). The other
**9** (the Doctor family) are **environmental transients, not code defects**
— proven by a same-SHA flip experiment — and no repair was applied to them.

## Evidence

### E1 — RED-now census at the dispatch tree (304bc8158)

| # | Test | Result at 304bc8158 | Command (verbatim) |
|---|------|--------------------|--------------------|
| 1 | `TestGateGraphFreshness_AllLayersFreshNotice` | **FAIL** — `fresh posture must carry the literal all-fresh notice, got: "graph-freshness: citations=absent(value=0 threshold=0) (advisory)"` (gate_graph_freshness_test.go:208) | `go test ./internal/hook/quality/ -run 'TestGateGraphFreshness_AllLayersFreshNotice' -v` |
| 2-10 | `TestRunDoctor_{WithExport,WithFix,Verbose,AllFlags,VerboseAndDetail,ExportMode}`, `TestDoctorCmd_{Execution,ExportFlag,VerboseExecution}` | **all PASS** — 8 swept in one `-v` run, `TestRunDoctor_ExportMode` re-swept solo (`--- PASS (7.24s)`) after a selector typo in the 8-test run was caught by the run-count check; first full-selector run `ok internal/cli 75.220s` with zero FAIL lines | `go test ./internal/cli/ -run '^(TestRunDoctor_…\|TestDoctorCmd_…)$' -v` |

Measurement-integrity note: the first Doctor census run had a selector typo
(`TestDoctorCmd_ExportMode` typed for `TestRunDoctor_ExportMode`), caught by
counting `=== RUN` lines (8, expected 9) — the missing test was re-swept solo
before any claim was recorded. The lead's prediction that "some may already
be green" was correct and the green reads are non-vacuous (9 named runs
observed).

### E2 — Root-cause judgment (the lead's asked-for determination)

**The 10 were two different causes, not one:**

- **Doctor 9 — environmental transient (code 0).** Discriminating
  experiment: `TestDoctorCmd_Execution` failed at commit `2524c3603` during
  the 2026-09-02 integration window (transcript verbatim: `--- FAIL:
  TestDoctorCmd_Execution (5.44s)`, twice within ~30 min across two runs)
  and **passes at the same commit today**, as do all 9 at three trees
  (`2524c3603`, `48c35a4d4`, `304bc8158`). Stable-red era → stable-green
  era with **zero doctor-path commits between**
  (`git log 826a63ebf..304bc8158 -- internal/cli/doctor*.go` → empty; the
  only in-between commits are t365/t368, neither touching doctor paths).
  A same-SHA result flip across time is environment, not code. The
  installed binary is unchanged across the transition
  (`v3.1.2-1308-g65196a5a7`, built 2026-09-02T05:53:08Z), so the moving
  factor is environment state the doctor checks read (candidate: live
  MCP-server/doctor-check state during the heavy multi-lane window), not a
  tree or binary change. **No repair applied — there is no code defect to
  repair.**
- **Freshness 1 — real, fixture-side.** Root cause (manager-develop,
  verified): the fourth freshness layer `citations`
  (SPEC-CODEMAPS-ACCURACY-001, REQ-CMA-002) judges the codemaps docs
  themselves; a doc-less codemaps directory is `VerdictAbsent` —
  unjudgeable-absent, **not fresh** (`internal/graph/check_citations.go:86-90`)
  — a deliberate layer contract, so the emitter's offending branch is
  correct. The gate test's all-fresh fixture `gfStampAllLayers` was authored
  in the three-layer era and never grew the citation-free codemaps doc; the
  sibling CLI fixture `writeCodemapsProvenance`
  (`internal/graph/check_test.go:88-95`) was already four-layer-aware,
  proving the lag side. Fix: write `modules.md` into the fixture's codemaps
  dir + comment corrections (`internal/hook/quality/gate_graph_freshness_test.go`,
  +14/−5). No emitter change, no expectation reversal.

### E3 — Repair verification (fix commit c26cd7cc0)

```
$ go test ./internal/hook/quality/ -run 'TestGateGraphFreshness_AllLayersFreshNotice' -v
--- PASS: TestGateGraphFreshness_AllLayersFreshNotice (0.60s)

$ go test ./internal/hook/quality/ -count=1
ok  github.com/modu-ai/moai-adk/internal/hook/quality  12.290s

$ go vet ./internal/hook/quality/ && echo VET_CLEAN
VET_CLEAN

$ gofmt -l internal/hook/quality/   → (empty)
```

Subagent (manager-develop, model opus) verification ran uncached pre-commit;
the orchestrator re-ran the diff review and the `-count=1` package sweep +
vet + gofmt after its own comment edit. Implementation delegated to
manager-develop per the Status Transition Ownership Matrix (lane standing
spawn authority); the orchestrator held commit authority throughout.

## Baseline-attribution

All measurements this card: commands as listed in E1/E3, run in
`.claude/worktrees/t444` on branch `WT-doctor-freshness-reds`, pinned per
row — census at `304bc8158`, the discriminating experiment at `2524c3603`
and `48c35a4d4` (detached, returned), post-fix at `c26cd7cc0`'s tree. Window-era
red citations are attributed to their transcript-verbatim runs of
2026-09-02 and are historical observations, not re-executable here.

## Gaps

- The specific environmental factor that reddened the Doctor 9 during the
  window is **not identified** — the era's failure output was not captured
  (names only; a lane measurement gap, recorded as such). The same-SHA flip
  proves "environment, not code" without naming which part of the
  environment.
- `internal/graph` package tests were not run (package untouched by the
  fix). `golangci-lint` was not run (`go vet` is the mission's static
  check). No CI verdict exists (develop push is the lead's; the merge
  window is requested after this verdict).
- The catalog family (`TestCatalogHashParity` 35-entry drift +
  `TestManifestHashFormat`) was deliberately **not touched** — t443
  (lane-14) scope per dispatch; both remain red on develop.

## Residual-risk

- The new fixture write relies on `.moai/project/codemaps/*.md` paths never
  being counted as described-worthy by the endpoint-diff union; the CLI
  sibling fixture shares this exposure, so a future change to
  `mx.IsDescribedWorthy` would stale both at once (subagent-identified,
  accepted).
- The Doctor environmental reds can **recur** under similar environment
  states (e.g., heavy multi-lane windows) and would look like code
  regressions; the discriminator recorded here is a same-SHA re-run before
  diagnosing.
- Review lens for this card: default 4-perspective, run orchestrator-direct
  (diff is 1 test file, +14/−5); harness minimal. The final PASS/FAIL
  verdict remains the lead's on the evidence above.
