---
id: SPEC-INTEGRATION-STALE-BASE-001
title: "Base-freshness gate before the integration window: refuse entry when the candidate is more than K landings behind local main or carries no post-absorb re-measure"
version: "0.1.0"
status: draft
created: 2026-10-17
updated: 2026-10-17
author: manager-spec (card t1834)
priority: P1
phase: "v3.2.0 target"
module: "internal/factory, internal/cli, internal/config, internal/template/templates"
lifecycle: spec-anchored
tags: "integration-window, base-freshness, remeasure, absorb, gate, config, factory"
tier: M
related_specs:
  - SPEC-MERGE-WINDOW-QUEUE-001
  - SPEC-CANDIDATE-CI-001
  - SPEC-LOCAL-MAIN-FLOW-001
---

# SPEC-INTEGRATION-STALE-BASE-001 (card t1834)

Card: **t1834** (Class C, plan then run then sync, priority P1). Rationale in the card: Zuul-style gating (test a change on top of the changes that precede it) and trunk-based short-lived branches (reflect main first, then merge). Every figure and citation below was read in the card tree (`.moai/worktrees/t1834`, HEAD `fb749d372`, branch `WT-10-17-class-8`); anything unread is in plan.md §G (Gaps), never asserted here.

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-10-17 | Initial plan-phase authoring for card t1834. |

---

## §A — Context and problem

### A.1 What exists today (read in this tree)

1. **The re-measure record already carries the absorbed base.** `RemeasureRecord` is keyed by the candidate tree SHA and names `Base`, "the absorbed integration-branch commit" (`internal/factory/integration_remeasure.go:35-57`); `RunRemeasure` computes it as `git merge-base <integration branch> HEAD` (`integration_remeasure.go:612-620`, `:634-720`). A record is valid only through `ValidateRemeasureRecord` (`:135-165`), the single verifier every merge path consults (the `@MX:ANCHOR` at `:128-134`).
2. **The merge step is exact, and it fires inside the window.** `RunMergeStep` reads the record for the pinned tree and refuses cause 1 when it is missing or invalid (`internal/factory/integration_merge_step.go:321-327`), then refuses cause 2 (`MergeExitBaseMoved`, exit 2, `:32`) when `record.Base != tip` (`:332-335`). Every pre-merge cause releases the window (`releaseWindow`). So a lane that is stale, or whose tests were not re-run after the absorb, is today refused only after it has taken (or waited for and been promoted into) the window.
3. **Window entry has two doors, and both already host one pre-window precondition.** `moai integration acquire` runs the settings-drift precondition and the per-card red hold before any window-record mutation (`internal/cli/integration.go:441-458`), then records the window (`:476`) or, under `--wait`, enqueues a ticket (`:517`). `moai factory complete` acquires the window itself when the lane does not hold it: the candidate-CI hold sits at `internal/cli/factory_card.go:2193-2199` and the acquisition at `:2201`. A third path, queue promotion, grants the window to the head ticket (`internal/factory/integration_window_ops.go:472-486`) through the `GrantGate` seam, whose type carries only a card id (`:369`).
4. **The re-measure verb keys the candidate by the caller's working directory.** `remeasureWorktreeDir()` returns `os.Getwd()` (`internal/cli/integration_remeasure.go:118-120`), because a card session's `CLAUDE_PROJECT_DIR` names the primary checkout (comment at `:113-117`).
5. **Operator doctrine already orders the steps.** The re-measure runs before the lane joins the queue and the window schedules itself (`.claude/rules/moai/workflow/factory-dispatch-mechanics.md:114`); a lane at the tail "re-measures against the new tip, then re-acquires" (the cause-2 message, `integration_merge_step.go:333`).

### A.2 The gap

No check at window entry asks "is this candidate close enough to local main, and was it tested on top of what it absorbed?" A stale lane still queues, still waits behind a live holder, is still promoted, and is refused only at cause 1 or 2 after consuming a window turn (the window is released, but the turn and the queue position are spent). The card asks for that question to be answered before entry, with a tolerance K that the operator sets in configuration.

### A.3 Vocabulary

- **Landing**: one commit on the first-parent chain of the local integration branch (an integration merge, `Merge WT-... into main (card ..., integration merge)`, or a direct commit). Measured in this tree: `git rev-list --count --first-parent main~40..main` prints 40, while `git rev-list --count main~40..main` prints 664, so counting all commits would overstate "landings" by a factor of 16.6 (plan.md §G records the single-tree limit).
- **Landings behind (L)**: the number of landings on the local integration branch that the candidate's re-measure did not cover: `git rev-list --count --first-parent <record.Base>..<integration branch>`, where `record.Base` is the absorbed base stored in the valid re-measure record keyed to the candidate's HEAD tree (the very field merge cause 2 compares against the tip). When `record.Base` is not an ancestor of the tip, the candidate is stale whatever K is. At absorb-then-remeasure time `L = 0`. (The live `git merge-base` is not used: after a content-free absorb it moves while the stored base does not, and the two would disagree with cause 2.)
- **K**: the operator-set tolerance, `workflow.base_freshness_gate.max_behind_landings`.
- **Candidate tree**: the tree whose HEAD tree SHA keys the re-measure record, resolved the way the re-measure verb resolves it (the caller's working directory at `acquire`; the card's worktree at `complete`).

### A.4 What the gate asks, and why two conditions

(b) **Post-absorb re-measure**: a valid re-measure record keyed to the candidate tree's current HEAD tree. An absorb that changes content changes the HEAD tree, so a record keyed to the old tree cannot match, and a record keyed to the current tree was produced on content that includes everything absorbed. (a) **Absorb**: with that record in hand, `L <= K`, where L is measured from the record's stored `Base`. Measuring from the stored base makes the gate and merge cause 2 (`record.Base != tip`) agree: at K = 0 the gate refuses exactly when `record.Base != tip` (an ancestor `Base` makes L at least 1; a non-ancestor `Base` is refused outright). A content-free absorb leaves the tree and the record valid but moves the tip past the stored base, so the gate refuses until the lane re-runs `moai integration remeasure`, which rewrites the record with the new base; that is the same outcome cause 2 would force inside the window. The gate does not (and cannot) judge whether the recorded command is the right scoped suite; that remains the lane's declaration, exactly as today.

---

## §B — Requirements (GEARS)

16 requirements (Tier M ceiling 16). Each REQ line keeps `shall` on its first physical line. Acceptance criteria live in acceptance.md and map by `maps REQ-ISB-NNN`.

### B.1 Where the gate runs

- REQ-ISB-001: When `moai integration acquire` is invoked with a card id from a working directory whose HEAD is a card branch (neither the integration branch nor detached) and the gate is enabled, the verb shall evaluate base freshness before any window-record mutation and before any queue-ticket enqueue, at the position of the existing pre-window preconditions; when the invocation has no card id or the working directory is not a card branch, the gate is not applicable and the verb shall admit silently.
- REQ-ISB-002: When `moai factory complete` is about to acquire the window itself (the calling lane does not already hold it), the gate is enabled, and the case is not an adoption (REQ-ISB-016), the verb shall evaluate base freshness for the card's worktree before the acquisition, so that a refusal leaves the window record, the integration branch, and the card state unchanged.

### B.2 What it measures and decides

- REQ-ISB-003: The gate shall measure landings behind as the count of first-parent commits of the local integration branch that are not reachable from the `Base` stored in the valid re-measure record keyed to the candidate tree's HEAD tree, shall treat a stored `Base` that is not an ancestor of the integration tip as stale, and shall not count commits reachable only through a merged side branch.
- REQ-ISB-004: When landings behind exceeds K, the gate shall refuse window entry with a message that names the landings behind, K, the configuration key, the integration branch tip, and the remedy (absorb local `main` into the card branch, re-run `moai integration remeasure`, then retry).
- REQ-ISB-005: When the candidate tree's HEAD tree has no valid re-measure record, the gate shall refuse window entry whatever the value of landings behind, with a message naming the tree and the remedy (re-run `moai integration remeasure` after the last absorb).
- REQ-ISB-006: The gate shall judge the record through `ReadRemeasureRecord` and `ValidateRemeasureRecord` and shall not introduce a second record type, a second store, a second validator, or any verb that performs the absorb.

### B.3 Configuration

- REQ-ISB-007: The tolerance K shall be read from `workflow.base_freshness_gate.max_behind_landings` of the loaded configuration, shall default to 0 in the shipped template configuration, and shall appear as no integer literal in the gate's decision code.
- REQ-ISB-008: Where `workflow.base_freshness_gate.enabled` is false (the shipped default), the gate shall measure and shall print one advisory line when it would have refused, and shall refuse nothing and write nothing.

### B.4 Failure posture, authority, and refusal shape

- REQ-ISB-009: When the gate is applicable but cannot determine its verdict (configuration unreadable or K negative, integration branch unresolvable, or a git read fails), the gate shall admit the entry and print exactly one line naming the cause, and shall refuse only on positive evidence of staleness or of a missing or invalid record.
- REQ-ISB-010: The gate shall leave the merge step's cause 1 and cause 2 checks, their exit codes 1 and 2, and every exit code in 1-15 and 20 unchanged, and a K above 0 shall never relax them.
- REQ-ISB-011: Every gate refusal shall begin with the sentinel `BASE_FRESHNESS_VIOLATION:` and shall exit with code 16, which is distinct from 1-15 and 20.

### B.5 Registration and guidance

- REQ-ISB-012: The shipped template configuration and the repository's live configuration shall carry the new `workflow.base_freshness_gate` block, and the template block shall carry no card id, SPEC id, or date and no `enabled: true`.
- REQ-ISB-013: The new keys shall be registered wherever the configuration surface is registered (the Go struct and defaults, the configuration cache schema version, and the shipped-key inventory) so that the shipped-key reader guard passes.
- REQ-ISB-014: The gate shall read no audit verdict and no card evidence SHA, and shall write no factory-record field, no re-measure record, no window record, and no queue ticket.
- REQ-ISB-015: While the SPEC is in its run phase, the set of files changed between the plan pin and HEAD shall be a subset of the allow-list in plan.md §C plus the SPEC directory and `.moai/reports/t1834/`, so that no file owned by the cards in §D changes.
- REQ-ISB-016: When `moai factory complete` finds that the integration tip is a merge commit whose second parent is the card branch's current tip and that a valid re-measure record keys the tip's tree (the adoption predicate of REQ-MWQ-019 step 2), the verb shall not evaluate the freshness gate and the adoption path shall behave exactly as without the gate, whether or not the lane holds the window.

---

## §C — Out of Scope

### Out of Scope — evidence carry across an absorb (the card behind the t1715 split)

- Carrying an audit verdict and the card's evidence SHA across a conflict-free absorb, the card-owned path set, the carry predicate, and the `repair evidence` verbs.
- Any read or write of the audit verdict, the evidence SHA, or a `card.repair` event (REQ-ISB-014).

### Out of Scope — landing-surface guidance and mechanics (cards t1846 and t1847)

- The integration-worktree rule, the primary-checkout wording, the base-ref clause of card worktrees, and every edit to the 17 files of that card's allow-list.
- The primary-follow verb, the statusline landing-surface display, the configuration key that designates the integration worktree, the lease-time base warning, `worktree.baseRef` and `worktree_base_branch` handling, and any change to `workflow.local_main_integration.enabled`.

### Out of Scope — changing the merge step or the window queue

- The exactness of cause 2, the holder and lease logic, queue ordering, and promotion. The gate is an earlier, cheaper refusal and never an authority.
- Gating queue promotion: a lane that is fresh at entry and stale at promotion stays governed by cause 2.

### Out of Scope — performing the absorb or the re-measure

- The gate tells the lane to absorb and to re-measure; it runs neither. A lane merges `main` into its own branch by hand, as the doctrine already says.

### Out of Scope — judging the recorded command

- Whether the recorded re-measure command is the right scoped suite for the card. The record's command stays the lane's declaration.

---

## §D — Ownership boundary with neighbouring cards

| Neighbour | Owns | This SPEC owns | Overlap |
|-----------|------|----------------|---------|
| t1715 split, evidence carry (card D, t1738) | The audit verdict and the evidence SHA E surviving a conflict-free absorb with unchanged card-owned paths (`REQ-LSR-010..012` as held in `.moai/reports/t1715/moved/card-D.md`); it still requires a re-measure record for the tip tree (that file's REQ-LSR-010) | The refusal at window entry when landings behind exceeds K or no valid record keys the current tree | None. No requirement here reads or writes the verdict, E, A, or the repair verbs (REQ-ISB-014); the carried card still must pass this gate, which is the intended composition |
| t1715 itself (`SPEC-LANE-SELF-RECOVERY-001`, card A) | The `repair spec` and `repair evidence` verbs, `repair lease` and `repair tree` stubs | Nothing in that SPEC | None |
| t1846 (`SPEC-LANDING-SURFACE-RULES-001`) | Guidance only: no Go, no configuration; a 17-file allow-list that excludes every file this SPEC edits | Go gate, configuration, and one gates-doc section | None by file (REQ-ISB-015) |
| t1847 (`SPEC-LANDING-SURFACE-MECH-001`, forward reference) | The mechanical landing surface: refusals and warnings concerning which tree holds the integration branch, the lease-time base warning about a card tree created from the wrong base ref | The refusal about the candidate's distance from the integration tip and the evidence that it was re-measured after absorbing | The same two verbs (`acquire`, `factory complete`) are edited by both cards; the seam is semantic (distance and evidence versus surface and creation base) and is kept by placement (§E.4) |

## §E — Constraints and assumptions

1. **Authority stays in the merge step.** The gate is an early refusal whose fail-open posture is safe because `RunMergeStep` re-enforces the strict form (REQ-ISB-009, REQ-ISB-010).
2. **Default-off shipping posture.** The repository's guard family ships its refusal layer off and neutral in the template (`internal/config/defaults.go:1541-1550` for the two nearest siblings); this gate follows it. Whether this repository's live configuration opts in is decision-index Q2.
3. **K=0 default.** Because L is measured from the record's stored `Base` (REQ-ISB-003), K = 0 makes the gate test the same predicate as cause 2 (`record.Base != tip`), so at K = 0 enabling the gate refuses at entry what the merge step would refuse inside the window for that reason; causes 3 to 13 stay the merge step's alone (decision-index Q1).
4. **Placement keeps the seam with t1847.** The gate is one function in a new file under `internal/factory` plus one new CLI file; `integration.go` and `factory_card.go` each gain exactly one added call placed beside the existing candidate-hold precondition (in `acquire`, immediately after the existing target resolution), with no reordering of any existing line; the only edit to existing lines is one behaviour-preserving extraction in `factory_card.go` (the adoption predicate into a shared helper, REQ-ISB-016), pinned by AC-ISB-016.
5. **Template-first.** The template configuration and the gates guidance are edited in `internal/template/templates/` and the repository copies follow; the guidance pair is byte-identical today (`diff -q` printed nothing for `factory-dispatch-gates.md` and `factory-dispatch-mechanics.md`).
6. **No time estimates** appear in any artifact.
