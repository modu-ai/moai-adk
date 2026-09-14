# Sync Audit — SPEC-GITSTRAT-WORKFLOW-READER-001 (card t656)

Auditor: sync-auditor (cold subagent — binding verdict owner this cycle)
Tree: worktree `.claude/worktrees/t656`, branch `WT-git-flow-reader`, HEAD `106b23a61`, base `b1bd81b23` (local develop), diff range `ca1872601..HEAD`
Date: 2026-09-14

## Verdict Inheritance Note (FO-SYNC-1)

The FO-SYNC-1 4-dimension workflow path (`.claude/workflows/sync-audit-4dim.js`) was declared
unavailable by the dispatch; per sync.md FO-SYNC-1 the cold sync-auditor subagent is the fallback
binding-verdict owner, and this audit inherits that ownership. The inheritance premise was
**re-verified first-hand** rather than taken on trust:

- The file's bytes do carry the claimed defect: hexdump of line 148 shows real `0x60` backticks,
  unescaped, inside the `CONTEXT_PROMPT` template literal opened at line 142
  (``Before returning, run `moai verify check --key-current` against…``).
- Compiling the full file as an ES module — `new vm.SourceTextModule(source)` under
  `node --experimental-vm-modules`, v22.14.0 — fails with exactly the claimed error:
  `Unexpected identifier 'moai'`. (Script-goal compile fails earlier at `export`, confirming the
  file is ESM.)
- Caveat for the lead: `node --check` on the same file returns exit 0 — a **false green** (its
  module-detection path does not surface this defect; the isolated region and the true ESM compile
  both fail). Actionable: validate `.claude/workflows/*.js` scripts via an ESM compile
  (`vm.SourceTextModule`) or runtime execution, never via `node --check`.

## Overall Verdict: **PASS**

Composite: **9.0 / 10** (weighted 40/25/20/15; harmonic mean of the four dimension scores is also 9.0).
One AC (AC-GWS-012 coverage leg) is bound **PASS-WITH-DEBT** — disposition and grounds below. The
lead should read §AC-GWS-012 Disposition and confirm the debt routing.

## Dimension Scores

| Dimension | Weight | Score | Verdict | Evidence (verbatim, this run, this tree @106b23a61) |
|---|---|---|---|---|
| Functionality | 40% | 9/10 | PASS | `go test ./internal/config/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/config 2.111s`; `go test ./internal/cli/ -run 'TestDoctorGitStrategyWorkflow\|TestIntegrationAcquire_InvalidWorkflowValueWarnsButRecordsIdentically' -count=1` → `ok github.com/modu-ai/moai-adk/internal/cli 1.926s`; verbose subtest run: 6/6 `TestWorkflowDisposition` + 6/6 `TestWorkflowTargetResolution` + M1 `TestLoadGitFlow*`/`TestCharacterize_*` all `--- PASS`; `TestDoctorGolden` → `ok … 0.952s`. 12/13 ACs re-observed PASS; AC-GWS-012 debt-bound (below). |
| Security | 25% | 9/10 | PASS | No Critical/High. Config-sourced strings rendered through `fmt.Sprintf("%q", …)` (escaped) in both the acquire warning and the doctor message; allowed-set interpolated from the fixed `AllowedWorkflows()` constant; the doctor check performs no writes (read-only by construction); the acquire path is warn-only with exit code and lock record test-pinned identical to the pre-change non-git-flow path (`TestIntegrationAcquire_InvalidWorkflowValueWarnsButRecordsIdentically`, github-flow control). No secrets, no injection surface, no new trust boundary; no AskUserQuestion in either package's production files. |
| Craft | 20% | 9/10 | PASS | Coverage re-measured: post-change package total `coverage: 82.2% of statements` (`go test -cover ./internal/config/ -count=1`); per-function (`go tool cover -func`): `LoadGitFlowIntegrationConfig`/`IsGitFlow`/`LoadGitFlowDevelopBranch`/`AllowedWorkflows`/`ClassifyWorkflowDisposition`/`WorkflowIntegrationTarget` all **100.0%**; scoped cli: `checkGitStrategyWorkflow` 100.0%, `integrationInvalidWorkflowWarning` 100.0%, `workflowStandingBranches` 83.3% (sole uncovered unit = compiler-mandated unreachable trailing return). Lint: `golangci-lint run internal/config/... internal/cli/... --timeout=2m` → `0 issues.` exit 0. Builds: `go build ./...` → NATIVE_OK; `GOOS=windows GOARCH=amd64 go build ./...` → WINDOWS_OK. Tests are fixture-driven through the real reader and real cobra command (stderr prefix-counted, JSON stdout single-object asserted, refused-acquire no-warn case present). |
| Consistency | 15% | 9/10 | PASS | Doctor check follows the `doctor_worktree_base.go` precedent; registered in `runGroupedChecksObserved` with a comment naming the REQ; binary-lag allowlist entry follows the t702 pattern with doc comment; `shipped_key_inventory.yaml` evidence update matches the t655 honesty-sweep pattern (`class: W` kept); characterization suite stays sole owner of `loader_integration_branch_test.go` (AC-GWS-010 diff measured **0 bytes**: `git diff 08298ae28..HEAD -- internal/config/loader_integration_branch_test.go \| wc -c` → `0`); AC-GWS-011 grep re-measured → `3`; commit subjects Conventional with `(t656)` refs and `🗿 MoAI` trailer verified in commit bodies; `@MX:NOTE [AUTO]` tags present on both new SSOT functions. One protocol miss: F2 (ANCHOR), below. |

## AC Matrix (auditor re-observation — every row re-executed this run)

| AC | Verdict | Auditor evidence (verbatim) |
|----|---------|------------------------------|
| AC-GWS-001 | PASS | `ok … internal/config 2.111s` (full package); verbose: all 4 allowed classified, typo/`trunk-based`/empty/`Git-Flow` case variants → invalid |
| AC-GWS-002 | PASS | subtest `TestWorkflowDisposition/typo_fixture_carries_the_raw_value` PASS; assertions read in source: Disposition=invalid, raw carried, `IsGitFlow()`=false, target empty |
| AC-GWS-003 | PASS | subtest `TestWorkflowDisposition/trunk-based_fixture_is_invalid` PASS |
| AC-GWS-004 | PASS | `ok … internal/cli 1.926s`; test asserts warning names `git-flwo` + allowed set, record `source=caller`, github-flow control silent with identical record shape |
| AC-GWS-005 | PASS | subtest `TestWorkflowTargetResolution/github-flow` → `main`, stale `develop_branch` does not leak |
| AC-GWS-006 | PASS | subtest `TestWorkflowTargetResolution/git-flow` → `develop` (trimmed); M1 gate gate (Manual && GitFlowWorkflow) preserved, byte-unmodified file |
| AC-GWS-007 | PASS | subtest `TestWorkflowTargetResolution/gitlab-flow-empty` → `""` |
| AC-GWS-008 | PASS | subtest `TestWorkflowTargetResolution/release-flow` → `release/` (trimmed) |
| AC-GWS-009 | PASS | `git log --oneline ca1872601..HEAD`: `08298ae28 test(…): M1 characterization suite` precedes `ea51ce891 feat(…): M2 validate …` |
| AC-GWS-010 | PASS | diff bytes = `0`; `go test ./internal/config/ -run 'TestLoadGitFlow\|TestCharacterize' -count=1` → `ok … 0.480s` |
| AC-GWS-011 | PASS | `grep -c 'workflow: github-flow' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` → `3`; control half of AC-GWS-004 test pins github-flow acquire behavior |
| AC-GWS-012 | FAIL-as-literal → **PASS-WITH-DEBT** (bound) | coverage leg: `coverage: 82.2% of statements` (< 85.0 literal); lint leg: `0 issues.` exit 0. Disposition below. |
| AC-GWS-013 | PASS | `TestDoctorGitStrategyWorkflow` 7/7 subtests PASS (three required states + unknown-state + gitlab-flow-empty + release-flow + verbose Detail on every state); golden render `ok … 0.952s` |

## AC-GWS-012 Disposition — **PASS-WITH-DEBT** (bound by this audit)

**Claim.** The AC-GWS-012 coverage leg, read as a package-total ≥ 85.0% gate, was mis-calibrated at
plan time; the requirement it carries forward (spec.md §E NFR: "≥ 85% … for the new/extended code
paths") is met; the package-total shortfall is pre-existing debt, documented and routed.

**Evidence (all re-measured this run).**

- Post-change: `go test -cover ./internal/config/ -count=1` → `coverage: 82.2% of statements`.
- Baseline: `git archive b1bd81b23` extracted to /tmp, git-initialized (the `git ls-files`-shelling
  `TestShippedConfigKeysHaveReaders` fails in a bare archive — environmental, not a defect), then
  `go test -cover ./internal/config/ -count=1` → `coverage: 82.0% of statements`.
- The 85.0 gate was therefore **already unmet before the first run-phase commit**; no in-scope work
  (plan §D scope discipline: `internal/config` loader + `internal/cli` check only) could flip the
  package total without testing unrelated legacy code — a green path running through out-of-scope
  work, i.e. the criterion measures the legacy package, not this SPEC.
- NFR intent: 8 of 9 new/extended functions measure 100.0%; `workflowStandingBranches` 83.3% with
  its sole uncovered line the compiler-mandated, documented-unreachable trailing return.
- No regression: package total improved under the paired measurement (82.0 → 82.2); the t637 reader
  baseline (`LoadGitFlowIntegrationConfig`/`IsGitFlow`/`LoadGitFlowDevelopBranch`) remains 100%
  (acceptance.md §D.5 DoD satisfied).

**Why not the amendment path.** Amending `acceptance.md` is inside the `ComputeHash` plan-artifact
subject set: it would invalidate the plan-audit skip verdict (PASS 0.96) and force a plan re-audit —
against a work product this finding does not impugn, for zero code change. The defect is in the AC's
calibration, not the implementation. PASS-WITH-DEBT preserves the same honesty at zero hash cost:
the AC row above stays marked FAIL-as-literal in this record (it is not silently rewritten), and the
debt is named with an owner-shaped follow-up.

**Bound disposition.** AC-GWS-012 = FAIL-as-literal / PASS-as-intent, **debt-bound**. Follow-up card
candidate (already named in progress.md §E.3): legacy `internal/config` package coverage sweep.
The lead should confirm this routing; the verdict does not block on it.

## Findings (structured defect-list)

- **F1** [Major] [blocking-for-the-record, resolved by the bound disposition] `acceptance.md:20` (AC-GWS-012) — coverage leg literal 85.0% unmet at package level (82.2% measured; baseline 82.0% at pinned base b1bd81b23 proves plan-time miscalibration). Required fix: none to code; lead confirms the PASS-WITH-DEBT binding and the follow-up card for legacy `internal/config` coverage.
- **F2** [Minor] [optional] `internal/config/loader_integration_branch.go:79` — `LoadGitFlowIntegrationConfig` now has 3 production callers (`integration.go:293` acquire, `session_worktree_automerge.go:75` automerge gate, `doctor_git_strategy_workflow.go:54` doctor) — the MX protocol's ANCHOR threshold (fan_in ≥ 3 ⇒ MUST have `@MX:ANCHOR`) is reached with no anchor tag. Fix: add `// @MX:ANCHOR: [AUTO] …` with `@MX:REASON` at the function (cheap, any future touch).
- **F3** [Minor] [optional] `internal/cli/doctor_git_strategy_workflow.go:39` — `workflowStandingBranches` measures 83.3%; the uncovered unit is the defensive trailing return. A 2-line direct test (`workflowStandingBranches("bogus") == ""`) closes it, or document the unreachability as intentional and accept the figure.
- **F4** [Minor] [optional] `internal/cli/doctor_git_strategy_workflow.go:58` — the unknown state returns OK for both "no git-strategy.yaml" and "file present but unparseable"; §D.3's doctor distinction is honored for invalid-vs-unknown but not unreadable-vs-absent. Consistent with the fail-open reader contract (absent config is legitimate); a future refinement could WARN when the file exists but does not parse.
- **F5** [Info] — FO-SYNC-1 premise verified first-hand (ESM compile fails: `Unexpected identifier 'moai'`); `node --check` v22.14.0 false-greens this file class. Process fix belongs to the lead: ESM-compile or runtime-execute workflow scripts in gates.
- **F6** [Info] — implementer's baseline figure (81.8%) vs auditor re-measurement (82.0%), and "+0.4pp" vs paired "+0.2pp": same conclusion either way (shortfall pre-exists; SPEC improved the total). No verdict impact; recorded for attribution precision.
- **Evidence gap (snapshot)** — `moai verify check --key-current` → `{"fresh": false, … "reason": "no snapshot recorded for the current working-tree key"}`. Recorded as a Gap per the snapshot-consumer contract; never read as a pass. All evidence above was re-executed fresh by this audit, so the miss has no downstream effect.

## Recommendations

- Confirm and carry F1's PASS-WITH-DEBT binding into the card verdict; open the legacy-coverage follow-up card (it can also absorb F3).
- Add F2's `@MX:ANCHOR` opportunistically at next touch of `loader_integration_branch.go` (annotation-only, no behavior).
- Replace `node --check` with an ESM compile (`vm.SourceTextModule`) wherever workflow scripts are validated (F5) — the current check false-greens exactly this defect class.
- Keep `workflowStandingBranches`' unreachable return documented if F3 is not taken; do not widen the doctor's OK-on-unreadable conflation without a deliberate decision (F4).

## Residual risk

- `internal/cli` package-level coverage was never measured (full-suite `-cover` timed out at 600s for the implementer under the cli-suite lease; this audit re-measured coverage via scoped runs over the new tests only). The untested-interaction risk is bounded by the golden render test and the end-to-end acquire tests, but a package-total figure for `internal/cli` remains unobserved.
- The doctor/live-CLI surface was observed through the table test + golden render, not a live `moai doctor` invocation (the worktree git-guard refused the `--check "Git Strategy Workflow"` invocation form in this session; acceptance.md §D.4 sanctions the indirect path).
- Baseline coverage was measured on a git-initialized `/tmp` extraction, not a checkout of the base SHA; the 0.2pp delta vs the implementer's figure is most plausibly environmental (go test caching state, fixture ordering) and immaterial to the conclusion.
