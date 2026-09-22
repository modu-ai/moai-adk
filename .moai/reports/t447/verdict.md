# t447 verdict — pipeline continuation carry now surfaces instead of degrading silently

Class B defect repair. Audit provenance: F1 [Medium] raised by lane-4 in the
t191 sync-audit (`af6d2dcf1`), routed to lane-9 by lead-1 on 2026-09-03.

## Premise re-measurement (the lead's condition for starting)

- The axis does NOT exist on the develop tip this card was cut from
  (`5a8449859`): no `project_continuation.go`, no `ProjectContinuation*`
  constants, no continuation rows in `doc-generation.md`, no
  `workflow.yaml` key. The whole
  SPEC-PROJECT-CONTINUATION-KEY-001 feature lives only on the unmerged
  branch `WT-project-continuation` (t191 tip `55885aae3`, NOT an ancestor
  of develop — verified with `git merge-base --is-ancestor`).
- Repair therefore merged the dependency first, per the kanban isolation
  doctrine ("a dependency is a reason to merge, never to reuse the
  tree"): `b8e84f862` merges t191's tip into this card's worktree.
- On the merged tree, lane-4's observation HOLDS: `none` skips issuance
  and is reported (`doc-generation.md:334`); `card` is the default;
  unmatched values resolve to card WITH a report line (`:333`, `:357`,
  rationale at `project_continuation.go:50-55`); `pipeline` — the only
  value asking for MORE than default — carried no signal anywhere if the
  orchestrator did not honor the carry. Zero tests read
  `doc-generation.md` (F2 re-confirmed on this tree).

## Repair-direction judgment (the decision the lead asked for)

Report line, matched to the unmatched-value form, plus a stop-short
naming clause — implemented. Rejected alternatives, with grounds:

- **Error / fail-loud**: disproportionate. The prose records why the
  unmatched path does not stop the run ("stopping a whole /moai project
  run over a mistyped presentation preference is disproportionate" —
  `project_continuation.go:52-54`); `pipeline` is a valid configured
  value of the same presentation class, without the irreversibility that
  buys a run-stopping failure.
- **Mechanical enforcement** (make the carry non-silent by construction):
  this is the `progression_mode` reading that
  SPEC-PROJECT-CONTINUATION-KEY-001 §3 D1.3 evaluated and rejected WITH
  measurement — reviving it is a Class C design change, out of this
  Class B card's scope.
- The timing asymmetry (the degradation happens after Step 4.2) is
  handled by the second sentence: the pipeline branch itself now names a
  stop short of the gate at the moment it happens, which covers
  legitimate stops (open `[NEEDS CLARIFICATION]` markers, context
  handoff, operator interrupt) rather than only the Step 4.2 promise.

## Claim

1. The Step 4.2 summary carries the carry-commitment line under
   `pipeline`, mirroring the offending-value line's placement and form.
2. The pipeline branch names a stop short of the gate, with reason.
3. Both `doc-generation.md` mirrors are byte-identical after the edit.
4. The content guard is non-vacuous: red observed before, green after.

## Evidence

| # | Claim | Command | Observed |
|---|---|---|---|
| 1 | RED before fix | `go test ./internal/template/ -run TestPipelineCarrySignal -count=1` (pre-edit) | exit 1 — carry-commitment + stop-short subtests FAIL for the missing fragments; unmatched-line subtest passes. `.moai/reports/t447/red-baseline.txt` |
| 2 | GREEN after fix | same compound invocation, post-edit | `ok github.com/modu-ai/moai-adk/internal/template 0.411s` — re-run by the lane, not only by the implementer. `.moai/reports/t447/green.log` |
| 3 | Mirror parity | `cmp .claude/.../doc-generation.md internal/template/templates/.../doc-generation.md` | exit 0 |
| 4 | Edit placement | grep on template copy | carry-commitment paragraph at `:359`; `**Stop-short naming**:` inside the pipeline bullet at `:390`, between the Ordering sentence and the Kanban Mode sentence |
| 5 | Affected packages | `go test ./internal/template/ ./internal/config/ -count=1` | template green; config red = inherited only (see Gaps). `.moai/reports/t447/affected-packages.log` |
| 6 | vet / gofmt | `go vet ./internal/template/`; `gofmt -l internal/template/` | vet exit 0; gofmt lists 10 pre-existing files, the new test NOT among them |

## Baseline-attribution

All measurements in this run, on branch `WT-pipeline-fallthrough`:
merge base `b8e84f862` (develop tip `5a8449859` + t191 tip `55885aae3`),
fix commit `106191798`. Implementer: manager-develop (one spawn); RED
observed before the prose edit existed, green judged by go-test compile
of the working tree (go:embed is compiled at test time), independently
re-run by the lane after the commit.

## Gaps

1. **`bin/moai` was not rebuilt.** `make build` fails at its first
   prerequisite (`agents-emit-check`): commit `4244c4a06` (t386/t387)
   edited `templates/.claude/agents/moai/sync-auditor.md` without `make
   agents-emit`, so the committed TOML predates the `.md`. Regenerating
   is outside this card's file mandate and is owned by the operator
   queue as the known catalog-hash drift from `4244c4a06` — named here,
   not repaired. Consequence: the local binary lags; the source-level
   and CI-level verification is unaffected (CI builds from source).
2. **Inherited reds, named and excluded (pre-measured, not caused by
   this card):** `TestAlwaysLoadedTokenBudget` (internal/config; 76,939
   vs 76,400 budget; its measured surface is `.claude/rules/moai/**` +
   root instruction files — excludes every file this card touched);
   `TestCatalogHashCoversSkillSubfiles` + `TestManifestHashFormat`
   (internal/template; born at `4244c4a06`, clear when the operator
   queue's `make agents-emit` + `gen-catalog-hashes --all` runs). The
   implementer measured the catalog tests failing identically BEFORE any
   edit (`3fff7dba…` vs tree `f005e873…`), so this card's edit moved the
   computed hash value but not any verdict.
3. **No full-suite run** (lane-local verification discipline; CI owns
   the full verdict at develop push).
4. **No runtime detector exists** for the degradation itself — it
   happens at orchestrator-behavior time, and the class-P prose carrier
   was the SPEC's deliberate design. The guard asserts the mechanically
   observable layer: the shipped text makes the carry and its failure
   legible.

## Residual-risk

- The signal is prose-carried: a context compaction mid-carry drops the
  rule AND both new lines from the working context — the compaction case
  is unfixable within the prose carrier and was already recorded in
  t191's residual risks.
- The stop-short clause relies on the stopping orchestrator reading the
  branch prose; a non-compliant orchestrator that ignores the carry can
  also ignore the naming clause. The Step 4.2 commitment line still
  gives the operator the promise-vs-observed pairing in that case.
- Window-order coupling: this branch CONTAINS t191's commits (dependency
  merge). If t447 merged before t191, lane-4's work would land under
  this card's branch — t191 MUST merge first at the window.

🗿 MoAI — card t447, lane-9
