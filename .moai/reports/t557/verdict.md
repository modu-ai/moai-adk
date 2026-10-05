# t557 — GH #1661 graph-freshness red: settlement + staleness-axis verdict

Worktree: `WT-codemaps-restamp` @ `04de513e4` (base, "chore(develop): merge main dependency updates").
Check binary: built from this tree with the CI-identical command (`go build -o ./bin/moai ./cmd/moai`), bootstrap `mx scan --quiet` + `graph build` run first, mirroring `.github/workflows/graph-freshness.yml` step for step.

## Claim

1. The card's original question (orphaned-stamp: shallow-clone vs commit-design) is SETTLED as given: fixed on main by PR #1665 (`da791eb0a`) plus the pre-merge stamp-reachability guard in `graph-freshness.yml`.
2. The CURRENT red (`layer codemaps verdict=stale metric=described-source-diff value=45 threshold=40`, main @ `7374b183e`, CI run 34426828262) is REPRODUCED EXACTLY and is a main-side state only. It does NOT reproduce on this worktree's tree.
3. This worktree's check is GREEN at base and stays green after this card's commit. No codemap regeneration or re-stamp is required on this branch; the main-side red is closed by the t592 restamp + the new anchor-based check code already on the develop lineage, delivered to main by the release lane.

## Evidence

### E1 — BEFORE verdict, this worktree @ `04de513e4` (verbatim, exit 0)

```
codemaps  metric=described-source-diff value=14 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=fresh
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=fresh
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
```

`./bin/moai graph check --json` for the codemaps layer: `content_anchor=f7b491954`, `content_anchor_source=last-body-change`, `contribution=0`, `contribution_base=4ec93ef3`.

### E2 — Exact reproduction of the main-side 45 (baseline-attribution: git objects in this checkout)

Main @ `7374b183e` ran the OLD check code — `internal/graph/check.go` at that commit has NO `resolveContentAnchor` (grep count 0) and its `gitDiffNameCount` applies NO file filter: every changed path under `internal/ cmd/ pkg/` counts, tests and non-Go included. Main's tracked `provenance.json` at that commit names stamp `a995e58fa` (generated 2026-08-26). Therefore:

```
$ git diff --name-only a995e58fa 7374b183e -- internal cmd pkg | wc -l
45
```

45 = the observed `value=45`. The same measurement against today's origin/main tip (`2213871af`, which still carries the OLD check code and the OLD stamp) gives **47 ≥ 40 — main is still red now**; the red persists until the release lane lands develop's state.

Under THIS tree's check code (anchor-based, described-worthy-only: non-test `.go`, no `testdata`), the same main-side window measures far smaller: `da791eb0a → 7374b183e` = 23 described-worthy files. The red is thus jointly produced by the stale main-side stamp AND the old unfiltered counter.

### E3 — Which described sources drifted on THIS tree (the staleness axis, 14 files, all from one commit)

Anchor `f7b491954` (t592 codemaps regeneration, 2026-09-09) → this working tree, described-worthy (`mx.IsDescribedWorthy`):

| # | File |
|---|------|
| 1 | internal/cli/codex_contract.go |
| 2 | internal/cli/codex_init.go |
| 3 | internal/cli/codex_launcher.go |
| 4 | internal/cli/codex_readiness.go |
| 5 | internal/cli/doctor_codex.go |
| 6 | internal/cli/init.go |
| 7 | internal/cli/init_agent_wizard.go |
| 8 | internal/cli/update.go |
| 9 | internal/cli/update_codex_wiring.go |
| 10 | internal/cli/update_mirror_heal.go |
| 11 | internal/cli/update_version.go |
| 12 | internal/cli/wizard/questions.go |
| 13 | internal/cli/wizard/translations.go |
| 14 | internal/cli/wizard/types.go |

Single driving commit: `6c647bbe2 feat(cli)!: align LLM harness instruction contract` (the codex/init/wizard/update surface). 14 < 40 → fresh by design; the threshold exists to tolerate exactly this inter-restamp drift.

### E4 — Why the red closes without further work on this branch

- The t592 restamp (`f7b491954` body + `45adc855e` provenance) and the NEW anchor-based check (`2649fe296`, t478) are both ancestors of this HEAD and of origin/develop tip (`1cfc6f544`); both are ABSENT from origin/main (`2213871af`).
- Computed drift at origin/develop tip: `git diff --name-only f7b491954 1cfc6f544 -- internal cmd pkg` filtered described-worthy = 25 < 40.
- A release PR (develop → `release/vX` → main) carries BOTH the fresh stamp and the new counter to main in one merge: the PR-head-built binary runs the new code, the merged tree's provenance names `f7b491954`, anchor resolves `last-body-change`, drift ≈ 25-30 < 40 → green. The release-lane reachability guard passes: `f7b491954` is reachable from any develop-cut release head.

### E5 — AFTER verdict, this worktree at committed HEAD (verbatim, exit 0)

Sequence mirroring the CI job exactly — build from the committed HEAD, then the two bootstrap steps, then the check:

```
$ go build -o ./bin/moai ./cmd/moai          # at committed HEAD
$ ./bin/moai mx scan --quiet && ./bin/moai graph build
$ ./bin/moai graph check
codemaps  metric=described-source-diff value=14 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=fresh
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=fresh
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
exit: 0
```

Observation recorded between the BEFORE and AFTER runs: running `graph check` at the committed HEAD WITHOUT first re-running the bootstrap reported `edges metric=source-fingerprint-mismatch value=1 threshold=0 verdict=stale (source set(s) moved: reports)` — the commit of this report file moved the `reports` source set's fingerprint relative to the edges.jsonl built at base. This is the exact state the workflow's bootstrap step exists to resolve ("mx-index and edges.jsonl are untracked runtime artifacts … the bootstrap … scoping the CI signal to the tracked, curated codemaps layer", graph-freshness.yml): with the bootstrap re-run, edges is fresh. No tracked artifact moved.

### E6 — Re-stamped codemaps

**None.** No codemap file or `provenance.json` was modified by this card. Rationale: the check is green on this tree at base (E1); the only failing surface is main-side and is delivered-to by the existing t592 restamp via the release lane (E4); t592 regenerated the codemaps one day before this card; a second regeneration would rewrite curated prose as a side effect of a green gate and churn the release batch without changing any verdict. This card also does NOT touch the threshold (40 unchanged).

## Baseline-attribution

- Every command above was run in THIS worktree (`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a1b419183ab5e39af`, branch `WT-codemaps-restamp`) in this session, against HEAD `04de513e4` (before) and the committed HEAD (after), with the binary built from the same tree by `go build -o ./bin/moai ./cmd/moai`.
- The 45-reproduction (E2) is a tree-to-tree git measurement over objects present in this checkout's history; `a995e58fa` resolves as an object here (as it did in the fetch-depth:0 CI checkout).
- Main-side values (47 at `2213871af`, old-code grep counts) were measured by `git show`/`git diff` against the named commits, not carried from the CI run log; the CI run 34426828262 output itself was not fetched (no network reads).

## Gaps

- CI run 34426828262's own log was not read (network fetch of the run log out of scope); the 45 was reproduced mechanically instead (E2, exact match).
- The post-release main drift (E4, "≈25-30") is a projection from the develop-tip measurement (25) plus commits that may land before the release is cut; it was not measured on a future tree, which does not exist.
- This card did not and cannot change main directly (push/PR forbidden); main stays red (47 ≥ 40 under old code) until the release lane lands. The green on this branch is necessary, not sufficient, for #1661's closure — the release cut is the delivering act.
- origin/develop tip `1cfc6f544` ("fix(ci): resolve review and release gate regressions") landed after this card's base was cut; its 11 additional described-worthy files are counted in the develop-tip figure (25) but are not part of this branch's tree.

## Residual-risk

- If main accumulates ≥ 15 more all-path changes before the release merge, the release-PR merge-preview could momentarily re-measure under the OLD... no — the release PR builds its binary from the PR head (new code), so the old counter never runs again on main after the release lands. The residual risk is instead: main-side drift described-worthy count growing past 40 BEFORE the release lands would keep the (old-code, old-stamp) red alive longer, but cannot block the release PR itself.
- If the release lane is not cut promptly, main's red persists independently of this card's work (already the live state).
- The untracked-file union in the metric means an untracked `.go` file under internal/cmd/pkg on a future tree would add to the count; none exist on this tree.
