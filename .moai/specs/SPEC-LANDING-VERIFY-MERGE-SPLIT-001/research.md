# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — research (plan phase, card t1872)

Measurement tree: `ec164d45d2e3c5a7902aa950ab212e8f385f5d08` (current main at authoring). Tool build: `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/c5bd52fe-ffbe-40e1-94ce-4491a79c0784/scratchpad/moai-t1872`, built with `go build -o <that path> ./cmd/moai` from the same clean tree (build exit 0), so the judging build and the tree coincide (verification-claim-integrity §2.2). Go toolchain `go1.26.8 darwin/arm64`.

This file is the plan-phase baseline. It is committed in its own commit, ahead of the spec/plan/acceptance commit (§2.3 ordering), and carries no implementation.

## 1. Which landing path this repository runs

- `.moai/config/sections/git-strategy.yaml:9` — `workflow: git-flow`, `develop_branch: main`; integration merges land locally on `main` ("Merge WT-… into main (card tNNNN, integration merge)" in `git log --first-parent`). Therefore the window path of `moai factory complete` applies, not the github-flow pull-request path.
- `internal/cli/factory_card.go:2071-2073` — `if factoryGitHubFlow(root) { return factoryCompleteGitHubFlow(...) }` bypasses the window and the re-measure record entirely. The distributed default is github-flow; **everything in this SPEC concerns the git-flow window path only** (the github-flow edge is untouched).

## 2. Where "base differs from main tip" is enforced (file:line)

| Place | Evidence |
|---|---|
| Merge step refuses when the record's base is not the tip | `internal/factory/integration_merge_step.go:321-333` — reads the record keyed by the pinned tree (`ReadRemeasureRecord`, :321), validates it (:325), then `if record.Base != tip { … MergeExitBaseMoved … "the integration branch moved past the record's base %s to %s — re-measure against the new tip, then re-acquire --wait" }` (:332-333). Exit code constant `MergeExitBaseMoved = 2` at :42. A second tip re-read inside the lock section at :464 refuses "moved mid-step". |
| `complete` step 3 | `internal/cli/factory_card.go:2281-2283` — no record for the CURRENT candidate tree → exit code `MergeExitBaseMoved`, message "no valid re-measure record for the candidate tree … run moai integration remeasure, then re-acquire --wait". Step 2 (adoption, :2229-2237) accepts an already-merged tip whose tree has a valid record. |
| Tree identity ties the landed tree to the measured tree | `internal/factorylane/merge.go:307` `checkTreeIdentityPreMerge`: `git merge-tree --write-tree <develop> <branch>` must equal the branch tree, i.e. the card branch must already CONTAIN the current tip. This is why a re-absorb is structural: a moved tip changes the merge result. |
| Record store is keyed by tree, base = merge-base | `internal/factory/integration_remeasure.go:45-62` (`RemeasureRecord`), :68-90 (store `.moai/state/remeasure/<tree>.json`), :699 `Base` is `gitMergeBase(worktree, baseBranch)` taken at record time. |
| Single verifier | `internal/factory/integration_remeasure.go:135` `ValidateRemeasureRecord` — consumed by the merge step (:325), `complete` (`factory_card.go:2234`, :2373-2377), `factory_merge.go:113`, and merge-check condition (d) (`internal/factorylane/merge.go` EvaluateMergeTriple :174, (d) block). |

## 3. Premise check of the re-measure record's evidential weight (new finding, leader 14:4x)

`ValidateRemeasureRecord` (`integration_remeasure.go:135-166`) accepts a record when `exit_code == 0` and, for non-`go test` commands, nothing else: line `StructuredRequired: isGoTestCommand(command)` (:698) is false for `sh script.sh`, and the validity rule is then command present + exit 0 (doc comment :185-193: "A tool with no recognized report … is valid on exit code zero and records command + exit only"). A script-wrapped call records no packages, no per-package terminal event, no attempt identity.

Observed real record (read-only): `/Users/goos/MoAI/moai-adk-go/.moai/state/remeasure/0f05ab3e3e2b3032bcbd650c6c4607d52beef109.json` — `"command": "sh …/t1834-gate-rest.sh"`, `"structured_required": false`, `"structured_count": false`, `exit_code 1`. The leader's note records that t1680's landed tree `1a3dc230c35ca0e2a997fc0fc3ac9444eb1f9198` was gated by a single `gate-b.sh` record with `structured_required=false` (`/tmp/adk-queue-repair-20261018/t1872-evidence-note.md`, machine-local). Consequence: nothing mechanical states which calls were required or that all of them passed on the same tree.

Standing scope rules (provenance only; they live on the machine-local decision board, not in the committed tree, so this SPEC restates the derivation it needs in `spec.md` §2): b2cd = "packages the card changed plus the packages importing them", whole package, no `-run`; the single recorded narrowing allowance 35e9 = `internal/cli` only (memory note `feedback_remeasure_scope_is_full_changed_package`). Board ids: standing rule `d-20261018T143825Z-5923`; t1709 ruling `d-20261018T133909Z-07c9`.

## 4. Slot lease (heavy slot) — current behaviour

- `internal/factory/slot_lease.go:127-142` record (one holder, `pid` = owning SESSION pid); :179 `Stale()` (pid ≤ 0 reads live); :189 `Expired`; :314 `AcquireSlotLease`; :353 takeover switch `case current.Stale()` → reason `stale`, then `expired`, then `force`; :404 `ReleaseSlotLease`. `internal/cli/slot.go:131-133` records `session.ResolveOwnerPID()` (the session, not the heavy process).
- Config: `internal/config/types.go:860-864` `SlotLeaseConfig{Enabled, DefaultMaxDuration, Resources}`; the repo's `workflow.yaml` carries no `slot_lease:` block (defaults).
- **Probe (item 4 premise).** Scratch git repo at `…/scratchpad/slotprobe`, a hand-written record with `pid 99999` (dead: `kill -0 99999` → "no such process"), far-future `expires_at`:
  - `moai slot status --resource probe-res` → `held by a session that is gone (reclaimable)`, exit 0; the record file STILL EXISTS afterwards (`ls` shows `probe-res.json`).
  - `moai slot acquire --resource probe-res --session other-session …` (no `--force`) → exit 0, `displaced: dead-session (pid 99999), reason stale`; audit row `{"event":"takeover","reason":"stale",…}`.
  So the card's premise "the record remains until its bound" is **only partly true**: a dead owner is already reclaimable without `--force` at the next acquire; what remains is (a) the record file persists until some verb rewrites it, (b) `moai slot status` and the guard report it but do not clear it, (c) liveness is bound to the SESSION pid — a live lane session whose heavy command died or was abandoned keeps the slot until its declared bound. Also: the integration window already auto-clears a dead holder (`internal/factory/integration_window_ops.go:44-52`, `holderOwnerGone`).
- Live slot now: `moai slot status` → `slot go-test-heavy: held`, lane-27, `t1680 landing remeasure (extend)`, bound 30m0s (read at authoring).
- Capacity: the lease core is strictly single-holder per resource (one JSON file per resource). Capacity 2 needs either a multi-holder record or N sibling records per resource; neither exists. `factory_slots.go` / `factory_slots_capacity_test.go` concern the LANE registry (different object).

## 5. Landing-throughput baseline (measured)

Command: `git log --first-parent --grep='integration merge' --format='%ct' -41 ec164d45d2` → consecutive intervals (minutes), n=40: `min=0.9 median=34.0 p90=158.5 max=460.3 mean=69.1`. `moai factory status` (tree-built binary) lists 16 cards in `state=merge-ready` at authoring (t1626 … t1871). Window state: `release-integration window: free`, policy open.

## 6. Overlap-eligibility baseline for item (1) (measured, approximate)

Method: `evidence/overlap_baseline.py` (committed beside this file), output `evidence/overlap_baseline.out`. Pair model: for each of the last 40 landings, the "waiting card" is landing *i+1* (its changed files = diff from the merge-base of its parents to its second parent), the "moved set" is landing *i* (first-parent diff). Verification scope ≈ directories of `go list -deps -test` closure of the Go package directories the card touched (computed ONCE at the analysis tree, an approximation) ∪ card files ∪ module files. Result (verbatim head of the output):

```
{ "pairs": 40, "decided": 37, "lenient_reusable": 10, "strict_reusable": 1, "undecidable": 3 }
```

- `lenient` ignores moved non-Go files outside the closure; `strict` treats any such file as overlap (the card's "non-Go read" hazard).
- Closure size is large: median 147 module-local package directories for the touched packages (`go list -deps -test -f '{{.Dir}}'` — `./internal/cli` 150, `./internal/factory` 24, `./internal/config` 26 directories; 1.1-2.3 s each).
- 3 of 40 landings touched no Go package and are undecidable by the closure rule (fail-closed → overlap).

**Reading:** under the card's own conservative rule (undecidable = overlap) only ≈3% of adjacent pairs keep a re-measure record; even the unsound lenient reading keeps ≈27%. Adjacent-pair is the most favourable case (one moved landing); with several landings between a record and its landing the cumulative moved set only grows. Item (1) therefore cannot be the main throughput lever on this history; this is carried to the operator as a decision (decision-index Q8) and is why instrumentation (item 5) records a **shadow eligibility counter** before any reuse relaxation opens.

## 7. Tests that exist for the touched surfaces

- Merge step: `internal/factory/integration_merge_step_test.go`, `integration_merge_step_overlap_test.go`, `mergestep_red_t1582_test.go`, `mergestep_atomic_t1582_test.go`; complete refusal: `internal/cli/factory_complete_mwq19_test.go:95` `TestMWQ19_Scenario3_MovedBaseRefusesWithReacquireCode`; remeasure: `internal/factory/integration_remeasure_test.go`, `integration_remeasure_run_test.go`, `remeasure_red_t1582_test.go`, `internal/cli/integration_remeasure_cmd_test.go`; slot: `internal/factory/slot_lease_test.go` (`TestSlotLease_Liveness`:343, `_DeclaredBound`:418, `_ForceRecordsDisplaced`:513), `slot_lease_cross_test.go`.
- Prior SPEC lineage: SPEC-MERGE-WINDOW-QUEUE-001/-002 (record, window, merge verb), SPEC-RESOURCE-SLOT-LEASE-001 and SPEC-HEAVY-TEST-SLOT-001 (slot), SPEC-VERIFY-RUN-REUSE-001 (`moai verify run`: same-tree same-command reuse — different object, cross-tree reuse here).

## 8. CLI-existence probes behave vacuously (RED-now design constraint)

`moai integration manifest|reuse|ready|metrics --help` each print the parent `integration` help and exit 0 (observed), so a help probe cannot be a RED-now cell. RED-now cells in `acceptance.md` use `go test -list` name-count assertions (the SPEC-VERIFY-RUN-REUSE-001 pattern) and `grep` exit codes.

## 9. Gaps

- Closure for §6 was computed at one tree for all pairs; per-pair trees were not checked out.
- The vouch-list (operator-declared never-read non-Go globs) variant of §6 was not measured; it needs the operator's list (decision-index Q8).
- No wall-clock measurement of a 15-25 minute re-measure was taken in this plan phase (figures are the card's, 10-18).
- The decision-board texts quoted in §3 were read through `moai decision read`; they are machine-local, not committed authority.
