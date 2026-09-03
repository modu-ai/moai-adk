# lane-9 stack re-measurement (discipline ① — final merge-tree verification)

Cumulative stack of all 9 lane-9 cards over local develop `3bdd5a803`, built outside the
integration window (lead order 2026-09-03: code cards ≥2 per window → re-verify on the
final merge tree). 2026-09-03.

## Baseline attribution

- Stack base: local develop `3bdd5a803` (fast-forward from branch point; no reset used)
- Stack order (window merge order): t367 → t453 → t302 → t345 → t300 → t348 → t353 →
  t339 → t447 — zero conflicts; one AUTO-MERGE of `internal/template/catalog.yaml`
  (t348 entries joined; verified consistent by the catalog tests below)
- Final stack commit: `651079f89` · measured tree: `88c52a2f1df47c12448d25d58185f8b04a999c79`

## Commands + results (this tree)

- `go build ./...` → exit 0
- `go test -count=1 ./internal/template/... ./internal/spec/... ./internal/config/...`
  → exactly the canonical exclusion reds + the budget finding:
  - TestCatalogHashCoversSkillSubfiles / TestManifestHashFormat / TestCatalogHashParity
    (moai entry, t436; sync-auditor entry, t443 — computed moai whole-tree hash
    f01063f0… includes t447's doc-generation.md edit)
  - TestGoldenCommittedArtifactsMatchEmission (sync-auditor.toml only, t443)
  - TestAlwaysLoadedTokenBudget → **77,432 / 77,200, overflow 232** (see below)
  - plan-auditor entry ABSENT from all drift lists — t367's scoped repair holds on the stack
  - all other packages/tests green, including every card's own guard test
- `go test -count=1 -timeout 1800s ./internal/cli/...` — solo run → **ALL GREEN, exit 0**
  (internal/cli 377.027s + agentlint/harness/pr/preference/printer/specid/taskledger/
  uikit/update×5/wizard/worktree all ok). Absorbed-side cli composition verdict: PASS.

## Budget projection for the lead's decision (no unilateral raise)

| Anchor | Value | Source |
|---|---|---|
| develop `3bdd5a803` alone | 77,104 | lead measurement (lane-12) |
| lane-9 stack (9 cards) | **77,432** | this run (tree 88c52a2f1) |
| lane-9 card contribution | +328 | subtraction |
| + lane-5 t436 when it lands first | +81 | lead measurement |
| **projected surface at lane-9 landing** | **77,513+** (+ any lane-12 always-loaded growth) | projection |

t453's current raise (77,200) is insufficient for the landing-time surface. Options for the
lead: raise to cover the projection with margin, or trim always-loaded rules. Decision
rests with the lead per dispatch order; t453's branch stays at 77,200 until then.

## Window-time plan recorded (lane-9)

1. `moai integration acquire --name lane-9`
2. verify t191 landed: `git log --grep "card t191"` on local develop, then merge t447 last
3. re-absorb the newest develop tip per branch if it moved (cheap merges)
4. re-measure the budget test on the actual merged tree; report the number before merging t453
   if the lead has re-decided the raise
5. sequential `--no-ff` merges into develop in the order above
6. scoped `gen-catalog-hashes.go --entry moai` after the last merge (t447's skills-tree edit
   re-stales the moai entry that t436's landing refreshes) so the tree lane-9 leaves is green
7. `moai integration release` · no push (lead bulk-pushes)
