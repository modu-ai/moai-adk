# SPEC-LOCAL-MAIN-FLOW-001 — Acceptance Criteria (v0.6)

Gate status. **Independent**: the criterion does not depend on any decision or on the order of the run. **Ordered after ABS-0**: the criterion's edit commits satisfy `git merge-base --is-ancestor 09a42899c HEAD` (V8), checked per commit. **GATED-ON-DECISION**: the criterion is not evaluated until the named operator decision in `decision-index.md` is recorded.

Claim labels. A claim about the local repository (HEAD, local main, the primary checkout) is marked `[local]`. A claim about a remote-tracking ref, or about what a remote holds, is marked `[origin]`. BASELINE_SHA, the origin/main value recorded at pre-flight, is used only for `[origin]` claims.

Tree attribution. Revision 0.4 covers cells E-01 to E-60 at HEAD `e32f69c46`. E-58 to E-60 were measured in revision 0.4; E-01 to E-57 were re-measured in revision 0.3. E-32 is retired because its command is identical to E-21. Revision 0.5 adds the lint and audit cells E-61 to E-65, measured at the same HEAD, each with its judging build in the note. Revision 0.6 re-runs E-22, E-27, and E-61 to E-65 at HEAD `366b45155`, with the revision 0.6 edits to the SPEC directory uncommitted, and adds E-66 to E-77. E-66 shows that the only files changed between `e32f69c46` and `366b45155` are the five SPEC files, so the code named by the other cells' tree pins is unchanged. E-22 and E-27 are historical: their starting states (the absent SPEC directory and the absent decision index) no longer exist, so they cannot be re-executed to the same result and carry no release-blocking weight. Both test existence only; neither tests the gap specification that AC-LMF-015 names. E-23, E-34, and E-45 read the moving ref `main`. They are subject claims about what local main carries when the criterion is evaluated, and they are re-run at that time; they are not pins.

## D. AC Matrix

| AC | Requirement | Gate | Severity | Green-path proof | RED-now cell |
|----|-------------|------|----------|------------------|--------------|
| AC-LMF-001 | REQ-LMF-001 | Independent | Must | V2, V3, V6, and F-01 | E-01 |
| AC-LMF-002 | REQ-LMF-002 | Independent | Must | V4 (TestIntegrationSurface and TestIntegrationMergeWorktree families) | E-02, E-29 |
| AC-LMF-003 | REQ-LMF-003 and REQ-LMF-014 (re-sync) | Independent | Must | V4 (TestLocalMainMerge, TestFactoryCompletePrimaryTreeGate, and TestLocalMainResync families), V10 | E-03, E-28 |
| AC-LMF-004 | REQ-LMF-004 | Independent | Must | V4 (TestLocalMainMergeRefusesDirtyPrimary) | E-28 |
| AC-LMF-005 | REQ-LMF-005 | Independent | Must | V5 (TestMergeStepDirty family) | E-04, E-05, E-07 |
| AC-LMF-006 | REQ-LMF-006 | Independent | Must | V4 (TestLocalMainMergeStatusSetUnchanged) and V5 (TestMergeStepStatusSetChanged family) | E-06 |
| AC-LMF-007 | REQ-LMF-007 | Independent (design record) | Must | F-07 | E-25, E-26, E-52, E-54, E-55, E-56, E-57, E-58, E-59, E-60 |
| AC-LMF-008 | REQ-LMF-008 | Ordered after ABS-0 | Must | V8, F-08, V2, V4 | E-08, E-09, E-30 |
| AC-LMF-009 | REQ-LMF-009 | Independent (enabled check); GATED-ON-DECISION (Q1, value check) | Must | F-09 | E-10 (value), E-77 (enabled) |
| AC-LMF-010 | REQ-LMF-010 | Ordered after ABS-0 | Must | F-10, V11 | E-11, E-12, E-13 |
| AC-LMF-011 | REQ-LMF-011 | Ordered after ABS-0 | Must | F-11 | E-14, E-15, E-36 to E-44 |
| AC-LMF-012 | REQ-LMF-012 | Ordered after ABS-0 | Must | F-12 | E-17, E-18, E-19 |
| AC-LMF-013 | REQ-LMF-013 | Ordered after ABS-0 | Should (regression guard) | F-13 | E-20, E-20c, E-21, E-21c |
| AC-LMF-014 | REQ-LMF-011 (release procedure text) and REQ-LMF-014 (re-sync sentence) | Ordered after ABS-0 | Must | F-14 | E-16 |
| AC-LMF-015 | REQ-LMF-015 | Independent | Should (regression guard; see the note in D.1) | F-15 | E-22 (historical) |
| AC-LMF-016 | REQ-LMF-003 and REQ-LMF-014 (the first merge) | GATED-ON-DECISION (Q1 and Q2) | Must, once gated open | F-16 and plan §J | E-34, E-45 |

Counts: 16 criteria. Independent: 8 (AC-LMF-001 to 007, and 015); AC-LMF-009 also carries an independent check of the enabled key. Ordered after ABS-0: 6 (AC-LMF-008, 010 to 014). GATED-ON-DECISION: 2 (AC-LMF-009 for its value check on Q1; AC-LMF-016 on Q1 and Q2). Severity: 14 Must, 2 Should (AC-LMF-013 and AC-LMF-015). Definition: Must = release-blocking (the RED cell is re-executable on the current tree and the criterion gates release). Should = regression guard, not release-blocking. Tier M ceiling: 16.

Requirement coverage. Each of REQ-LMF-001 to REQ-LMF-015 maps to at least one criterion (see the Requirement column). REQ-LMF-003 and REQ-LMF-014 also appear in AC-LMF-016, because the first merge exercises both.

## D.1 Given-When-Then Scenarios

### AC-LMF-001 — The template key defaults to false, and the template comes first (Independent)

**Given** the template workflow configuration carries no `local_main_integration` key (E-01).
**When** the key is added with `enabled: false`, the struct field, the default, and the loader are added, and `make build` exits 0 (V6).
**Then** `TestLocalMainIntegrationDefaultsFalse` passes, including the absent-key case, and `TestLocalMainIntegrationReadsEnabled` passes (V2). `TestWorkflowWorktreeKeyHonesty` passes (V3). At M1 the repository copy still has no key (F-01a). At M3 the template commit is an ancestor of the repository-copy commit (F-01b).

### AC-LMF-002 — The surface follows the gate and the holder (Independent)

**Given** the existing refusal test (E-02) and the absent surface tests (E-29).
**When** the gate is false and the primary checkout is the only holder of the configured branch.
**Then** `TestIntegrationMergeWorktreeRefusesPrimaryHoldingBranch` passes, so the refusal is unchanged.
**And when** the gate is true and the primary checkout holds the configured branch, **then** `TestIntegrationSurfaceSelectsPrimaryWhenEnabled` and `TestIntegrationMergeWorktreeAcceptsPrimaryWhenEnabled` pass.
**And when** the gate is true and the primary checkout is on another branch, **then** `TestIntegrationSurfaceRefusesPrimaryOffBranch` passes, with the case-5 guidance.
**And when** nobody holds the configured branch, **then** `TestIntegrationSurfaceRefusesNoHolder` and `TestIntegrationMergeWorktreeRefusesUnheldBranch` pass.
**And** a separate worktree that holds the configured branch behaves the same whatever the gate (`TestIntegrationSurfaceSeparateWorktreeUnchanged`).

### AC-LMF-003 — The landing verb merges on the primary, the re-sync fast-forwards, and the guard treatment holds (Independent)

**Given** no primary-surface merge test exists (E-28), and the adoption test exists (E-03).
**When** the fixture merge runs on the primary with the gate on: `TestLocalMainMergeMergesIntoPrimary` passes, and the first parent of the new HEAD equals the pre-merge HEAD, so the branch did not change.
**And** `TestFactoryCompletePrimaryTreeGateEnabled` passes with the gate on, `TestFactoryCompletePrimaryTreeGateDisabled` passes with the gate off, and `TestFactoryCompleteNoIntegrationTreeRefused` passes in both states.
**And** `TestLocalMainResyncFastForwards` passes when local main is behind the origin value recorded as BASELINE_SHA by fast-forward (`[origin]` the target; `[local]` HEAD moves): the verb moves local main to BASELINE_SHA and reports the old and new SHAs. The test sets `refs/remotes/origin/main` with `git update-ref` in a scratch repository and contacts no remote.
**And** `TestLocalMainResyncRefusesDiverged` passes when local main holds a commit that BASELINE_SHA (`[origin]`) lacks: the verb refuses with guidance, and HEAD is unchanged (`[local]`).
**And** `TestLocalMainResyncAheadIsNoop` passes when local main already contains BASELINE_SHA (`[origin]`): the verb reports that no fast-forward is needed, and HEAD is unchanged (`[local]`).
**And** `TestLocalMainResyncPreservesIgnoredFile` passes: an ignored file sits at a path that the fast-forward would write. The verb refuses with guidance, HEAD is unchanged, and the file keeps its content. Without `--no-overwrite-ignore`, git replaces such a file, as measured in cell E-72.
**And** in the run phase, V10 observes in a scratch primary that a PreToolUse payload whose command is `moai integration merge --card t1616` is allowed, with no `BRANCH_GUARD_VIOLATION:` line. The observation is recorded in the run progress record.

### AC-LMF-004 — A dirty primary is refused with guidance (Independent)

**Given** no dirty-primary refusal exists (E-28).
**When** the primary has one uncommitted change, or only an untracked file, and the gate is on.
**Then** `TestLocalMainMergeRefusesDirtyPrimary` passes. It returns `MergeExitWorktreeDirty`, the guidance text of plan §B4 is printed (it names the remedies and forbids stashing), and the integration window is released.

### AC-LMF-005 — Overlap rule for a separate integration worktree (Independent)

**Given** the clean precondition at `integration_merge_step.go:243-249` (E-05), no overlap logic (E-04), and no disjoint test (E-07).
**When** a dirty path is disjoint from every target path, and no ignored file sits at a target path. **Then** `TestMergeStepDirtyDisjointPathsProceeds` passes and the merge commit is created.
**When** a dirty path equals a target path, lies inside a target directory, or is the source of a rename into one. **Then** `TestMergeStepDirtyPathOverlapRefuses` and `TestMergeStepDirtyDirectoryPrefixOverlapRefuses` pass, each returning `MergeExitWorktreeDirty`.
**When** an ignored file sits at a target path. **Then** `TestMergeStepDirtyIgnoredTargetRefuses` passes.
**And** the existing `12 dirty before merge` case (`integration_merge_step_test.go:410`) still returns `MergeExitWorktreeDirty` for its overlapping fixture.

### AC-LMF-006 — The status set is compared after the merge and after an abort (Independent)

**Given** the three call sites of the clean helper at 243, 472, and 533 (E-06).
**When** the merge and any abort leave the status set unchanged, on either surface. **Then** `TestMergeStepHappyPathCreatesNoFFMergeAndReleases`, `TestMergeStepMergeFailureCleanAbortsCause6`, and `TestLocalMainMergeStatusSetUnchanged` pass.
**When** the status set changes after the merge commit. **Then** `TestMergeStepStatusSetChangedAfterMergeHolds` passes and returns `MergeExitPostMerge`.
**When** it changes after the abort. **Then** `TestMergeStepStatusSetChangedAfterAbortHolds` passes and returns `MergeExitMergeDirty`.

### AC-LMF-007 — The closure rule and its boundary are recorded (Independent, design record)

**Given** the no-remote refusal is live at `card_transition.go:670` (E-25), the reserved-edge refusal is consulted at line 333 (E-26), and the tool-owned closure's anchors are present: the tree-identity check at `card_evidence_readers.go:314` (E-52), the adoption record read at `factory_card.go:1936` (E-54), the candidate-tree record read at `factory_card.go:1983` (E-55), the merge-tree record read at `factory_card.go:2056` (E-56), and the lane-file read at `card_evidence_readers.go:335` (E-57).
**When** plan.md §B7 and §G.1 record the rule, and AGENTS.local.md §4.0 records the boundary sentence (M4).
**Then** `grep -c '^### B7' plan.md` returns 1. `grep -c 't1621' plan.md` returns at least 2. `grep -c 'tool-owned' plan.md` returns at least 1. `grep -c 't1621' AGENTS.local.md` returns at least 1 after M4. `[local]` `git diff --quiet e32f69c46 HEAD -- internal/homestate` exits 0 at the close of the card, so this card changes no closure code.

### AC-LMF-008 — The repository's integration branch is main and auto-merge is off (Ordered after ABS-0)

**Given** `develop_branch: develop` at git-strategy.yaml line 16 (E-08), `auto_merge: true` at workflow.yaml line 164 with its comment at line 153 (E-09), and the absorb present (E-30).
**When** the values change to `main` and `false`, with the comment of plan §B8, in a commit for which `git merge-base --is-ancestor 09a42899c HEAD` exits 0 (V8; `[local]`).
**Then** `grep -n 'develop_branch' .moai/config/sections/git-strategy.yaml` shows `main`, and the `manual:` key stays at line 8. The only `auto_merge:` line is `auto_merge: false`, and the phrase `auto_merge 유지` is absent. `TestEmptyTargetGuidance`, `TestEmptyTargetGuidanceResolvedTargetIsSilent`, `TestTargetProvenance`, and the fourteen `TestAutoMerge*` tests pass (V2, V4).

### AC-LMF-009 — deny_commits_on keeps its value, and the key is enabled (Independent for the enabled check; GATED-ON-DECISION Q1 for the value check)

**Given** `deny_commits_on: [main]` at line 178 (E-10), the repository copy without the key (E-77), and operator decision Q1 open in `decision-index.md`.
**When** the repository copy gains the `local_main_integration` block with `enabled: true`.
**Then** (independent of Q1) the line after `local_main_integration:` reads `enabled: true`, and this check is evaluated now. (Q1-gated) the `deny_commits_on` line still reads `[main]`. If Q1 is decided before the value check is evaluated, that check is replaced by a check of the decided value, and the replacement is recorded in the decision record.

### AC-LMF-010 — AGENTS.md carries the template's generic wording (Ordered after ABS-0)

**Given** the hardcoded sentences at AGENTS.md lines 91, 94, and 166 (E-11), two occurrences of "from local `main`" (E-12), and the template sentence present once in the template (E-13).
**When** AGENTS.md lines 90–96 and 161–167 are replaced by the template text (plan §B9), in a commit for which `git merge-base --is-ancestor 09a42899c HEAD` exits 0.
**Then** `grep -n -e 'commit-dead' -e 'card PRs go to base' -e 'card PR goes to base' AGENTS.md` prints nothing. `grep -c 'from local `main`' AGENTS.md` returns 0. `grep -c 'remote default branch' AGENTS.md` returns 1, matching the template. `git diff --quiet e32f69c46 HEAD -- internal/template/templates/AGENTS.md.tmpl` exits 0, so the distributed template is unchanged.

### AC-LMF-011 — The section 4.0 flow exists, and each superseded clause is marked in place (Ordered after ABS-0)

**Given** no section 4.0 and no markers (E-14, E-15), and the seven target lines and the two excluded lines as they stand (E-36 to E-44).
**When** section 4.0 is inserted before the heading at AGENTS.local.md line 175, and the seven markers of plan §B10 are added in place.
**Then** `grep -c '^### §4.0' AGENTS.local.md` returns 1. `grep -c 'SUPERSEDED by §4.0' AGENTS.local.md` returns 7. The check is by clause text, not by line number, because section 4.0 shifts every line below it: each marked line carries the clause it marks, namely the §4.1 chain heading (`표준 체인`), rule 1 (`카드 브랜치는`), rule 2 (`원격 기본 브랜치다`), rule 5 (`commit-dead다`), the integration-status duty (`승인이 아니다`), the lane's integration-window duty (`창을 받으면:`), and the card-PR-only public path (`유일한 공개 경로다`). The two unmarked clauses, the Factory leader request (`Factory(리더`) and the self-dispatch exception (`self-dispatch lane 예외 — 병합 창`), carry no marker. The original text of rule 1 is still present, so no clause was deleted. The seven marked clauses are the ledger cells E-38 to E-44, and the two unmarked clauses are cells E-36 and E-37; those cells record the card base's line numbers, which are pre-state values.

### AC-LMF-012 — The four documents no longer state develop as live (Ordered after ABS-0)

**Given** eight `CLAUDE.local.md` matches in the integration chain doc (E-17), six in the lane protocol (E-18), and the develop branch point in the PR policy (E-19).
**When** each live statement is corrected or marked, and each retired-file pointer is rewritten (plan §M4 items 4 to 7).
**Then** `grep -c 'Card worktrees branch FROM' .claude/rules/local/repo-local-pr-policy.md` returns 0. Each remaining `CLAUDE.local.md` line in the two docs carries the word "retired" or "former" on that line (inspection of `grep -n 'CLAUDE.local.md'`). Each remaining develop statement in the four documents carries a superseded marker, or is a dated historical record (inspection).

### AC-LMF-013 — The byte budget holds, and each growth is stated (Ordered after ABS-0; Should, regression guard)

**Given** AGENTS.md at 24,228 bytes and 24,122 characters (E-20, E-20c), and AGENTS.local.md at 40,850 bytes and 28,364 characters (E-21, E-21c).
**When** the edits land.
**Then** `wc -c AGENTS.md` returns fewer bytes than 24,228. `wc -m AGENTS.md` returns fewer than 40,000. Each run-phase commit that grows an always-loaded file by more than 1,000 bytes states the measured before and after byte counts and the cost statement in its body (plan §E.2).
**Severity.** Should (regression guard). The green path has a clause that is checked by reading each run-phase commit body (F-13), not by a re-executable command, and no run-phase commit exists yet. The command clauses are a growth guard: E-20 (24,228 bytes) is red for "fewer than 24,228", E-20c (24,122 characters) already satisfies "fewer than 40,000", and E-21 and E-21c (AGENTS.local.md) are context figures that no clause of this criterion tests.

### AC-LMF-014 — The batch release procedure is written down (Ordered after ABS-0)

**Given** no release procedure in AGENTS.local.md (E-16).
**When** section 4.0 is written.
**Then** `grep -c 'release/main-batch-YYYYMMDD' AGENTS.local.md` returns at least 1. `grep -c 'git push origin main:refs/heads/release/main-batch-' AGENTS.local.md` returns 1. `grep -c -e 'no squash' -e 'merge commit' AGENTS.local.md` returns at least 2. `grep -c -e 'fast-forward' -e 'sync branch' AGENTS.local.md` returns at least 1. The re-sync sentence of REQ-LMF-014 (plan §B3a) is present, and its behavior is proved by the AC-LMF-003 tests.

### AC-LMF-015 — Gaps (b) and (c) are specified with procedures and open questions (Independent; regression guard)

**Given** the pre-plan state had no gap specification. Its only RED observation, E-22, was taken at the card base and is historical.
**When** plan.md §G.2 and §G.3 are written.
**Then** `grep -c '^### G\.[23] Gap' plan.md` returns 2. `grep -c 'Run-phase procedure' plan.md` returns 2. `grep -c '^Open:' plan.md` returns at least 2, each naming an open question or a missing input.
**Note.** Because the RED cell cannot be re-executed on the current tree, the criterion loses release-blocking eligibility and is a regression guard (verification-completeness §2.1, undecidable disposition).
- (a) Where the probed artifacts now exist: E-22 probed `.moai/specs/SPEC-LOCAL-MAIN-FLOW-001`, and E-27 probed `.moai/specs/SPEC-LOCAL-MAIN-FLOW-001/decision-index.md`. Both are carried by commit `0cca5364f` and exist at HEAD `366b45155`, where revision 0.6 re-runs them (E-22, E-27). Neither cell tests the gap specification, so this criterion has no RED cell that tests its own claim. After the revision 0.6 commit lands, E-22 and E-27 are re-run and re-pinned to that commit.
- (b) The RED cell is kept as a regression guard, not a release gate: E-22 is a regression-guard record, and this criterion's pass is not a release criterion.
- (c) Severity: Must = release-blocking (the RED cell is re-executable on the current tree and the criterion gates release). Should = regression guard, not release-blocking. Split: 14 Must, 2 Should (AC-LMF-013 and AC-LMF-015).

### AC-LMF-016 — The first merge into local main, executed only after the decisions (GATED-ON-DECISION Q1 and Q2)

**Given** `[local]` the card tip is not an ancestor of local main (E-34), `[local]` local main is at `2aab5f797` (E-45), and both decisions are open in `decision-index.md`.
**When** the operator records Q2 and the chosen method is executed (plan §J).
**Then** the merge is verified by the rows of plan §J for the chosen option; no row of another option applies.
- Option (a), the operator's one-off merge: `[local]` the card tip is an ancestor of local main at evaluation time (`git merge-base --is-ancestor <card tip> main` exits 0), and `[local]` a valid remeasure record exists for the merge tree.
- Option (b), the designed landing verb: `[local]` the merge SHA is recorded; `[local]` its first parent equals the pre-merge HEAD; `[local]` the status set is equal before and after (REQ-LMF-006); `[local]` a valid remeasure record exists for the merge tree; and the closure follows plan §B7. `[local]` Local main contains the card tip after the merge.
- Option (c), the separate integration branch: `[local]` `git rev-parse main` is unchanged, and the card tip is on the integration branch. Local main does not contain the card tip by design, so no check requires it.

No first-merge check names origin/main; BASELINE_SHA applies only to the re-sync (plan §B3a). If Q1 has been recorded, its decided value is applied to `deny_commits_on`, and AC-LMF-009 is re-evaluated against that value.
**Until** both decisions are recorded, this criterion is not evaluated, and no step of this card merges into main.

## E. Evidence ledger

Each entry records one command, its verbatim standard output, and its exit code, and names the tree it was measured on. Entries are fenced because a table cell mangles shell metacharacters (verification-completeness §2.1). Every cell was re-run in revision 0.3 at tree `e32f69c46`. E-22 and E-27 are historical (see the header above). E-35 is a pipeline and is not release-blocking: it is a byte measurement of a clause that plan §B9 replaces. E-32 is retired because it repeats E-21. Revision 0.4 adds E-58 to E-60 at the same tree. HEAD moved from `e32f69c46` to `0cca5364f` when the five SPEC files were committed, and to `366b45155` in the revision 0.5 commit. Revision 0.6 re-runs E-22, E-27, and E-61 to E-65 at `366b45155`, corrects the E-63 note, and adds E-66 to E-77. E-66 shows that the only files changed between `e32f69c46` and `366b45155` are the five SPEC files, so the code named in the other cells' tree fields is unchanged. The supersede set of AC-LMF-011 is cells E-38 to E-44 (lines 216, 218, 219, 179, 195, 196, and 199); the two unmarked lines are E-36 (line 212) and E-37 (line 214).

```text
E-01
command: grep -c 'local_main_integration' internal/template/templates/.moai/config/sections/workflow.yaml
stdout: 0
exit: 1
tree: e32f69c46
```

```text
E-02
command: grep -c 'func TestIntegrationMergeWorktreeRefusesPrimaryHoldingBranch' internal/cli/integration_merge_worktree_test.go
stdout: 1
exit: 0
tree: e32f69c46
```

```text
E-03
command: grep -n '^func Test.*[Aa]dopt' internal/cli/factory_complete*_test.go
stdout: internal/cli/factory_complete_mwq19_test.go:201:func TestMWQ19_Scenario5_AdoptionRefusedAfterNewCommit(t *testing.T) {
exit: 0
tree: e32f69c46
```

```text
E-04
command: grep -c 'overlap' internal/factory/integration_merge_step.go
stdout: 0
exit: 1
tree: e32f69c46
```

```text
E-05
command: grep -n 'if !clean' internal/factory/integration_merge_step.go
stdout: 247:	if !clean {
exit: 0
tree: e32f69c46
```

```text
E-06
command: grep -n 'gitIntegrationWorktreeClean' internal/factory/integration_merge_step.go
stdout: 243:	clean, status, err := gitIntegrationWorktreeClean(in.IntegrationWorktree)
stdout: 472:			clean, _, cleanErr := gitIntegrationWorktreeClean(in.IntegrationWorktree)
stdout: 533:	if clean, _, err := gitIntegrationWorktreeClean(in.IntegrationWorktree); err != nil || !clean {
exit: 0
tree: e32f69c46
```

```text
E-07
command: grep -c 'func TestMergeStepDirty' internal/factory/integration_merge_step_test.go
stdout: 0
exit: 1
tree: e32f69c46
```

```text
E-08
command: grep -n 'develop_branch' .moai/config/sections/git-strategy.yaml
stdout: 16:        develop_branch: develop
exit: 0
tree: e32f69c46
```

```text
E-09
command: grep -n 'auto_merge' .moai/config/sections/workflow.yaml
stdout: 153:        # auto_merge 유지 (사용자 수동 worktree 생성 시에만 자동 머지).
stdout: 164:        auto_merge: true
exit: 0
tree: e32f69c46
```

```text
E-10
command: grep -n 'deny_commits_on' .moai/config/sections/workflow.yaml
stdout: 178:        deny_commits_on: [main]
exit: 0
tree: e32f69c46
```

```text
E-11
command: grep -n -e 'commit-dead' -e 'card PRs go to base' -e 'card PR goes to base' AGENTS.md
stdout: 91:EXCEPT on `main` — in this repository `main` is commit-dead (no session commits there; the
stdout: 94:cut from local `main`; card PRs go to base `main` — `develop` is legacy since the 2026-10-05
stdout: 166:the card PR goes to base `main` and the factory leader lands `main`; `develop` is legacy since
exit: 0
tree: e32f69c46
```

```text
E-12
command: grep -c 'from local `main`' AGENTS.md
stdout: 2
exit: 0
tree: e32f69c46
```

```text
E-13
command: grep -c 'remote default branch' internal/template/templates/AGENTS.md.tmpl
stdout: 1
exit: 0
tree: e32f69c46
```

```text
E-14
command: grep -c 'SUPERSEDED by §4.0' AGENTS.local.md
stdout: 0
exit: 1
tree: e32f69c46
```

```text
E-15
command: grep -c '§4.0' AGENTS.local.md
stdout: 0
exit: 1
tree: e32f69c46
```

```text
E-16
command: grep -c 'main-batch' AGENTS.local.md
stdout: 0
exit: 1
tree: e32f69c46
```

```text
E-17
command: grep -c 'CLAUDE.local.md' .moai/docs/gitflow-integration-chain.md
stdout: 8
exit: 0
tree: e32f69c46
```

```text
E-18
command: grep -c 'CLAUDE.local.md' .claude/rules/local/gitflow-lane-protocol.md
stdout: 6
exit: 0
tree: e32f69c46
```

```text
E-19
command: grep -c 'Card worktrees branch FROM' .claude/rules/local/repo-local-pr-policy.md
stdout: 1
exit: 0
tree: e32f69c46
```

```text
E-20
command: wc -c AGENTS.md
stdout:    24228 AGENTS.md
exit: 0
tree: e32f69c46
```

```text
E-20c
command: wc -m AGENTS.md
stdout:    24122 AGENTS.md
exit: 0
tree: e32f69c46
```

```text
E-21
command: wc -c AGENTS.local.md
stdout:    40850 AGENTS.local.md
exit: 0
tree: e32f69c46
```

```text
E-21c
command: wc -m AGENTS.local.md
stdout:    28364 AGENTS.local.md
exit: 0
tree: e32f69c46
```

```text
E-22 (historical)
command: test -e .moai/specs/SPEC-LOCAL-MAIN-FLOW-001
stdout: (none)
exit: 0 at 366b45155 (re-run in revision 0.6; the directory exists; exit 0 also at e32f69c46)
tree: 366b45155 (re-pinned in revision 0.6 from e32f69c46; the revision 0.6 SPEC edits are uncommitted, so this pin is re-run after the revision commit lands)
note: at the card base 2aab5f797 the directory was absent and the same command exited 1. Not re-executable to that result; regression guard only. The probe tests existence, not the gap specification named by AC-LMF-015.
```

```text
E-23
command: git merge-base --is-ancestor 09a42899c main
stdout: (none)
exit: 1
tree: e32f69c46
note: [local] moving ref (local main). Subject claim about local main at evaluation time; local main is 2aab5f797 (E-45).
```

```text
E-24
command: git status --short
stdout: ?? .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/
exit: 0
tree: e32f69c46
note: at the card base the output was empty; the difference is the untracked SPEC directory.
```

```text
E-25
command: grep -n 'refused while a remote is configured' internal/homestate/card_transition.go
stdout: 670:				return plan, fmt.Errorf("%w: merged-local → done is refused while a remote is configured", ErrIllegalTransition)
exit: 0
tree: e32f69c46
```

```text
E-26
command: grep -n 'isReservedEdge' internal/homestate/card_transition.go
stdout: 203:// isReservedEdge — pushed → ci-green stays reserved: the CI verdict reader
stdout: 207:func isReservedEdge(from, to string) bool {
stdout: 333:	if isReservedEdge(cur.State, req.To) {
exit: 0
tree: e32f69c46
```

```text
E-27 (historical)
command: test -e .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/decision-index.md
stdout: (none)
exit: 0 at 366b45155 (re-run in revision 0.6; the file exists; exit 0 also at e32f69c46)
tree: 366b45155 (re-pinned in revision 0.6 from e32f69c46; the revision 0.6 SPEC edits are uncommitted, so this pin is re-run after the revision commit lands)
note: at the card base the file was absent and the same command exited 1. Not re-executable to that result; regression guard only. The probe tests existence, not the content of the decision record.
```

```text
E-28
command: grep -rl 'func TestLocalMainMerge' internal --include='*_test.go'
stdout: (none)
exit: 1
tree: e32f69c46
```

```text
E-29
command: grep -rl 'func TestIntegrationSurface' internal --include='*_test.go'
stdout: (none)
exit: 1
tree: e32f69c46
```

```text
E-30
command: git merge-base --is-ancestor 09a42899c HEAD
stdout: (none)
exit: 0
tree: e32f69c46
note: [local] the absorb gate (REQ-LMF-008, §E of spec.md).
```

```text
E-31
command: git merge-base --is-ancestor 09a42899c 2aab5f797
stdout: (none)
exit: 1
tree: e32f69c46
note: control; the absorb is not an ancestor of the pre-absorb base.
```

```text
E-33
command: git rev-parse --short HEAD
stdout: e32f69c46
exit: 0
tree: e32f69c46
```

```text
E-34
command: git merge-base --is-ancestor e32f69c46 main
stdout: (none)
exit: 1
tree: e32f69c46
note: [local] moving ref (local main). The card tip is not on local main; the first merge has not happened.
```

```text
E-35
command: sed -n '90,96p' AGENTS.md | wc -c
stdout: 564
exit: 0
tree: e32f69c46
note: pipeline, outside the single-invocation form; not release-blocking. The byte figure of the section 2 clause that plan §B9 replaces.
```

```text
E-36
command: sed -n '212p' AGENTS.local.md
stdout: Factory(리더 `moai cc -f` · 레인 `moai cc -l`) 모드에서 레인은 카드 작업이 끝나면 **반드시 리더에게 `main` 반영을 요청한다.** 레인이 스스로 병합 창을 잡지 않는다.
exit: 0
tree: e32f69c46
note: not superseded (MI-5).
```

```text
E-37
command: sed -n '214p' AGENTS.local.md
stdout: - **self-dispatch lane 예외 — 병합 창.** Claude self-dispatch 팩토리 run의 레인은 위 요청을 하지 않는다 — `moai integration merge --card`를 부르는 `moai factory complete`의 통합 절차로 스스로 통합 창을 잡고 자기 카드를 `main`에 반영한다(OD-2). Codex 레인은 예외가 아니다 — merge-ready에서 정지한다(REQ-SD-025). 이 예외도 위의 다른 큐 변경(`add`, `drop`, `done`, `edit` 등) 금지와 `moai contract sign` 금지는 바꾸지 않는다(카드 임대만 `moai factory next`로 허용 — OD-1; `.claude/rules/local/gitflow-lane-protocol.md` §6).
exit: 0
tree: e32f69c46
note: not superseded (MI-5).
```

```text
E-38
command: sed -n '216p' AGENTS.local.md
stdout: - `moai integration status`가 `free`인 것은 **승인이 아니다.** open 정책에서 대기열 맨 앞으로 승격되어 창을 쥔 레인은 리더 지명 없이 병합한다.
exit: 0
tree: e32f69c46
note: superseded marker target (B10).
```

```text
E-39
command: sed -n '218p' AGENTS.local.md
stdout: - 창을 받으면: `moai integration acquire --name <lane> --card <card-id>` → 본인 워크트리에서 `git merge main` 흡수(대상은 **로컬** `main` — 원격이 아니다. 흡수 **전에** 그 로컬 main 이 최신인지부터 본다 — 판정식과 갱신 경로는 `.claude/rules/local/gitflow-lane-protocol.md` §11, 아직 develop 서술 — drift) → **병합 트리에서 재측정** → 통합 워크트리 진입 → `main` 반영(카드 PR base `main` 병합 또는 `--no-ff` 병합 — 세부 형태는 이관 문서가 정한다) → `moai integration release` → `ExitWorktree keep` → 완료 보고(반영 SHA를 리더에게 보고 — push·병합은 리더가 일괄로 한다)
exit: 0
tree: e32f69c46
note: superseded marker target (B10).
```

```text
E-40
command: sed -n '219p' AGENTS.local.md
stdout: - **[HARD] 카드 PR(base `main`)이 유일한 공개 경로다 (2026-10-05 전환 개정).** 카드가 마감되면 레인은 카드 브랜치를 push해 PR(base `main`)을 열되 **직접 병합하지 않는다** — 병합·push는 리더가 일괄로 수행하고 레인은 그 주체가 아니다. 레인은 `gh run rerun`/`workflow dispatch` 등 CI를 직접 요청·재요청하지도 않는다 — CI 판정은 `main` 반영이 일으키는 실행에 맡기고, 판독은 리더 몫이다. (종전 금지 — 운영자 지시 2026-09-01, WT 브랜치 push 금지·develop 일괄 push 체제 — 는 develop 체인과 함께 역사가 됐다; 당일 lane-2가 `WT-version-stamp-predicate`를 origin에 push한 전례로 추가됐던 조항이다)
exit: 0
tree: e32f69c46
note: superseded marker target (B10).
```

```text
E-41
command: sed -n '179p' AGENTS.local.md
stdout: **[HARD] 표준 체인 — GitHub Flow 전환(운영자 지시, 2026-10-05 M2 착지)**
exit: 0
tree: e32f69c46
note: superseded marker target (B10).
```

```text
E-42
command: sed -n '195p' AGENTS.local.md
stdout: 1. **카드 브랜치는 `main`에서 판다.** 카드가 끝나면 카드 PR(base `main`)로 `main`에 반영한다. `develop`은 legacy다 — 새 카드의 기저로 쓰지 않는다.
exit: 0
tree: e32f69c46
note: superseded marker target (B10).
```

```text
E-43
command: sed -n '196p' AGENTS.local.md
stdout: 2. **`main`은 원격 기본 브랜치다.** 그 head의 CI가 통합 판정을 만든다. 병합·push는 **리더가 일괄**로 수행한다(2026-09-02 — 레인은 push하지 않는다).
exit: 0
tree: e32f69c46
note: superseded marker target (B10).
```

```text
E-44
command: sed -n '199p' AGENTS.local.md
stdout: 5. **로컬 `main`은 commit-dead다 (SPEC-MAIN-COMMIT-BAN-001, 카드 t1337).** 어느 세션도 primary 체크아웃의 `main` 안에서 커밋하지 않는다 — `git commit` / `git revert` / `git cherry-pick`은 BranchGuard(`workflow.branch_guard.deny_commits_on: [main]`)가 거부하고, 커밋은 `main`에서 분기한 카드 워크트리에서만 만든다. 리더의 `main` push는 배치 트리거로 닫는다 — `git rev-list --count origin/main..main`이 `git_strategy.manual.lead_push_threshold`에 닿으면 배치를 닫는다(설정 파일의 계수 주석은 아직 `origin/develop..develop`을 가리킨다 — drift). 현재값의 원천은 설정 파일이고 이 규율은 키와 계수 명령을 명명할 뿐이다. push는 초록 조건부를 따른다 — 카드 병합마다 통합 창의 병합 트리 재측정이 사전 게이트이고, 마지막 push의 `origin/main` CI가 red인 동안 다음 push는 보류된다. 상세는 `.claude/rules/local/gitflow-lane-protocol.md` §4(해당 룰은 아직 develop 서술 — drift).
exit: 0
tree: e32f69c46
note: superseded marker target (B10).
```

```text
E-45
command: git rev-parse --short main
stdout: 2aab5f797
exit: 0
tree: e32f69c46
note: [local] local main's tip at measurement; the subject of E-23 and E-34.
```

```text
E-46
command: git rev-parse 'e32f69c46^{tree}'
stdout: ddc07c9a3800ce62be9a87d54c06ed145283eec7
exit: 0
tree: e32f69c46
```

```text
E-47
command: git rev-parse '09a42899c^{tree}'
stdout: ddc07c9a3800ce62be9a87d54c06ed145283eec7
exit: 0
tree: e32f69c46
note: equal to E-46; the absorb adds no tree content beyond 09a42899c.
```

```text
E-48
command: git config --get core.hooksPath
stdout: /dev/null
exit: 0
tree: e32f69c46
note: the installed pre-push hook does not run in this checkout (residual risk R-3).
```

```text
E-49
command: git rev-parse --short 09a42899c
stdout: 09a42899c
exit: 0
tree: e32f69c46
```

```text
E-50
command: sed -n '217p' AGENTS.local.md
stdout: - **레인은 `main`에 손으로 git merge 하지 않고 moai integration merge --card 또는 그것을 부르는 moai factory complete 로만 병합한다.** 창은 대기열이 스케줄하고, 리더는 `moai integration policy open|hold`로 정책만 잡는다(창을 받으면 acquire --wait로 대기열에 선다 — 재측정은 대기 전, 창 안은 병합뿐이다).
exit: 0
tree: e32f69c46
note: the lane's merge rule; supports the designed-path reading of B3.
```

```text
E-52
command: grep -n 'differs from its second parent' internal/homestate/card_evidence_readers.go
stdout: 314:		return mergeEvidence{}, evidenceErr("merge %s tree %s differs from its second parent's tree %s", full, tree, second)
exit: 0
tree: e32f69c46
note: the tree-identity check is the condition at line 313 (tree != second).
```

```text
E-53
command: grep -n -E 'func RunRemeasure|WriteRemeasureRecord\(projectRoot, tree' internal/factory/integration_remeasure.go
stdout: 68:func WriteRemeasureRecord(projectRoot, treeSHA string, rec RemeasureRecord) error {
stdout: 550:func RunRemeasure(projectRoot, worktree, baseBranch, command string) (*RemeasureRecord, error) {
stdout: 610:	if err := WriteRemeasureRecord(projectRoot, tree, *rec); err != nil {
exit: 0
tree: e32f69c46
```

```text
E-54
command: grep -n 'ReadRemeasureRecord(lockRoot, tipTree)' internal/cli/factory_card.go
stdout: 1936:			if record, recErr := factory.ReadRemeasureRecord(lockRoot, tipTree); recErr == nil && factory.ValidateRemeasureRecord(record) == nil {
exit: 0
tree: e32f69c46
note: the adoption path (lines 1924–1960).
```

```text
E-55
command: grep -n 'ReadRemeasureRecord(lockRoot, candidateTree)' internal/cli/factory_card.go
stdout: 1983:	if _, err := factory.ReadRemeasureRecord(lockRoot, candidateTree); err != nil {
exit: 0
tree: e32f69c46
note: the non-adoption path's pre-merge check (lines 1962–2010).
```

```text
E-56
command: grep -n 'ReadRemeasureRecord(lockRoot, treeSHA)' internal/cli/factory_card.go
stdout: 2056:			rec, err := factory.ReadRemeasureRecord(lockRoot, treeSHA)
exit: 0
tree: e32f69c46
note: the T16 verifier keyed to the merge tree; it runs on both paths through completeTransitions.
```

```text
E-57
command: grep -n -F 'readBoundedFile(path)' internal/homestate/card_evidence_readers.go
stdout: 177:	raw, err := readBoundedFile(path)
stdout: 335:		if _, err := readBoundedFile(path); err != nil {
stdout: 347:	raw, err := readBoundedFile(path)
exit: 0
tree: e32f69c46
note: line 335 is inside verifyMerge (lines 290–338): a lane-supplied remeasure path is read there when one is given. Lines 177 and 347 are in other functions.
```

```text
E-58
command: sed -n '157p' internal/homestate/card_transition.go
stdout: 		{"T18", CardMergedLocal, CardDone, guardNoRemote},
exit: 0
tree: e32f69c46
note: the T18 edge; cited by plan §A and §B7.
```

```text
E-59
command: sed -n '686,688p' internal/homestate/card_transition.go
stdout: 			if !remote {
stdout: 				return plan, fmt.Errorf("%w: merged-local → pushed requires a configured remote", ErrIllegalTransition)
stdout: 			}
exit: 0
tree: e32f69c46
note: the pushed-requires-remote check (T17); cited by plan §A and §B7.
```

```text
E-60
command: sed -n '668,670p' internal/homestate/card_transition.go
stdout: 		if edge.guard == guardNoRemote {
stdout: 			if remote {
stdout: 				return plan, fmt.Errorf("%w: merged-local → done is refused while a remote is configured", ErrIllegalTransition)
exit: 0
tree: e32f69c46
note: the no-remote guard and its refusal; cited by plan §A and §B7. Cell E-25 is line 670 alone.
```

```text
E-61
command: moai spec lint SPEC-LOCAL-MAIN-FLOW-001
stdout: ✓ No findings — all SPEC documents are valid
exit: 0
tree: 366b45155 (revision 0.6 SPEC edits uncommitted)
note: judging build = installed `moai version` v3.2.0-rc.29, commit 4f8aba061. Ancestry check: `git merge-base --is-ancestor 4f8aba061 HEAD` exits 1 (E-65), so the installed build is not an ancestor of the tree HEAD 366b45155. This row is a lag-check result; the tree rows E-63 and E-64 judge the tree. Re-run in revision 0.6 after REQ-LMF-013 was rewritten in the GEARS shall-not form: an earlier run in the same revision reported LegacyEARSKeyword for an If/then clause, and that clause was removed.
```

```text
E-62
command: moai spec audit --filter-spec SPEC-LOCAL-MAIN-FLOW-001 --json
stdout: {
stdout:   "audited_at": "2026-10-09T19:58:07.258307Z",
stdout:   "total_specs": 1,
stdout:   "grandfathered": 0,
stdout:   "modern_era_clean": 1,
stdout:   "drift_findings": [
stdout:     {
stdout:       "spec_id": "SPEC-LOCAL-MAIN-FLOW-001",
stdout:       "era": "V3R6",
stdout:       "finding_type": "EraAutoDetected",
stdout:       "severity": "INFO",
stdout:       "details": {
stdout:         "heuristic_matched": "H-5 (modern phase or created date)"
stdout:       }
stdout:     }
stdout:   ]
stdout: }
exit: 0
tree: 366b45155 (revision 0.6 SPEC edits uncommitted)
note: judging build = installed `moai version` v3.2.0-rc.29, commit 4f8aba061. Ancestry check: E-65 (exit 1; not an ancestor of 366b45155). Re-run in revision 0.6. The MCP tool `mcp__moai__spec_audit` on the same installed build reported V3R6 with no drift at revision 0.5; it was not re-run in revision 0.6, and it is not a shell command, so it has no cell of its own.
```

```text
E-63
command: moai-tree-novcs spec lint SPEC-LOCAL-MAIN-FLOW-001
stdout: ✓ No findings — all SPEC documents are valid
exit: 0
tree: 366b45155 (revision 0.6 SPEC edits uncommitted; no Go file differs from e32f69c46, E-66)
note: judging build = a build of this tree: `go build -buildvcs=false -o <scratchpad>/moai-tree-novcs ./cmd/moai`, run from the worktree root. sha256 c3dab13107c42c7267219db763a06e8a5e94d165148c8c8284be9713aeebd9be; a second build with the same flags, writing to another path, reproduced that hash in revision 0.6. Its version identity is the source default (v3.1.3, commit none), not the release build. The binary carries no VCS stamp: it has no vcs.revision, vcs.time, or vcs.modified key and no buildvcs build setting (E-73). The binary is outside the repository, at the session scratchpad `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/2890cd8c-1262-4d8d-9c24-5b650c5ab263/scratchpad/moai-tree-novcs`; the build command above reproduces it.
```

```text
E-64
command: moai-tree-novcs spec audit --filter-spec SPEC-LOCAL-MAIN-FLOW-001 --json
stdout: {
stdout:   "audited_at": "2026-10-09T19:58:08.330304Z",
stdout:   "total_specs": 1,
stdout:   "grandfathered": 0,
stdout:   "modern_era_clean": 1,
stdout:   "drift_findings": [
stdout:     {
stdout:       "spec_id": "SPEC-LOCAL-MAIN-FLOW-001",
stdout:       "era": "V3R6",
stdout:       "finding_type": "EraAutoDetected",
stdout:       "severity": "INFO",
stdout:       "details": {
stdout:         "heuristic_matched": "H-5 (modern phase or created date)"
stdout:       }
stdout:     }
stdout:   ]
stdout: }
exit: 0
tree: 366b45155 (revision 0.6 SPEC edits uncommitted)
note: judging build = the tree build described in E-63 (sha256 c3dab13107c42c7267219db763a06e8a5e94d165148c8c8284be9713aeebd9be; no VCS stamp, E-73; version identity the source default v3.1.3). Same findings as the installed build (E-62): V3R6, no MUST-FIX. Re-run in revision 0.6.
```

```text
E-65
command: git merge-base --is-ancestor 4f8aba061 HEAD
stdout: (none)
exit: 1
tree: 366b45155 (revision 0.6 SPEC edits uncommitted)
note: re-run in revision 0.6 at HEAD 366b45155: exit 1. Earlier runs at 0cca5364f and at e32f69c46 also exited 1 (the code tree is identical, E-66). The installed build's commit 4f8aba061 is not an ancestor of the tree. Exit 1 (not 128) means both commits resolve and the first is not an ancestor of the second.
```

```text
E-66
command: git diff --name-only e32f69c46 366b45155
stdout: .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/acceptance.md
stdout: .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/decision-index.md
stdout: .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/plan.md
stdout: .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/progress.md
stdout: .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/spec.md
exit: 0
tree: 366b45155
note: all five changed paths are SPEC files; none is Go code, template, or configuration. The code named by the cells pinned to e32f69c46 is therefore unchanged at 366b45155.
```

```text
E-67
command: go test ./internal/config/ -list <the V2 selector of plan §F, block V2>
stdout: TestStructYAMLSymmetry_Constitution
stdout: TestStructYAMLSymmetry_Context
stdout: TestStructYAMLSymmetry_Interview
stdout: TestStructYAMLSymmetry_Design
stdout: TestStructYAMLSymmetry_Statusline
stdout: TestStructYAMLSymmetry_GitConvention
stdout: TestStructYAMLSymmetry_Gate
stdout: TestStructYAMLSymmetry
stdout: TestEmptyTargetGuidance
stdout: TestEmptyTargetGuidanceResolvedTargetIsSilent
stdout: TestTargetProvenance
stdout: ok  	github.com/modu-ai/moai-adk/internal/config	0.342s
exit: 0
tree: 366b45155
note: RED-now selection is 11 of the 13 names in block V2. TestLocalMainIntegrationDefaultsFalse and TestLocalMainIntegrationReadsEnabled are absent until M1 adds them.
```

```text
E-68
command: go test ./internal/template/ -list <the V3 selector of plan §F, block V3>
stdout: TestAgentsDisclosureCompleteness
stdout: TestWorkflowWorktreeKeyHonesty
stdout: TestAutoMergeRequiredChecks
stdout: ok  	github.com/modu-ai/moai-adk/internal/template	0.365s
exit: 0
tree: 366b45155
note: 3 of the 3 names in block V3 are selected now.
```

```text
E-69
command: go test ./internal/cli/ -list <the V4 selector of plan §F, block V4> | grep -c '^Test'
stdout: 22
exit: 0
tree: 366b45155
note: pipeline, outside the single-invocation form; not release-blocking (as E-35). 22 of the 39 names in block V4 are selected before M1; the other 17 are absent until M1 adds them.
```

```text
E-70
command: go test ./internal/factory/ -list <the V5 selector of plan §F, block V5> | grep -c '^Test'
stdout: 3
exit: 0
tree: 366b45155
note: pipeline, not release-blocking. 3 of the 9 names in block V5 are selected before M2; the other six are absent until M2 adds them.
```

```text
E-71
command: sed -n '162p;176,183p' internal/cli/session_worktree_automerge.go
stdout: 	if cfg == nil || !cfg.Workflow.Worktree.AutoMerge {
stdout: 	if !gitFlow.IsGitFlow() || gitFlow.DevelopBranch == "" {
stdout: 		// REQ-WKW-003: the integration target is inert when the project is
stdout: 		// not manual git-flow, develop_branch is empty, or the file is
stdout: 		// unreadable (the loader yields the zero value on every failure).
stdout: 		autoMergeNoticef(out, "skipped (no integration target): the project is not manual git-flow or git_strategy develop_branch is unset; set git_strategy.<mode>.develop_branch to enable session-exit auto-merge")
stdout: 		return
stdout: 	}
stdout: 	develop := gitFlow.DevelopBranch
exit: 0
tree: 366b45155
note: line 162 reads the auto_merge switch; lines 176–183 take the integration target from develop_branch (gitFlow.DevelopBranch). Cited by plan §B8 and M3.
```

```text
E-72
command: sh <scratchpad>/ff-ignore-probe.sh <empty scratch directory>  (scratch script outside the SPEC; its two merge invocations are `git merge --ff-only --no-overwrite-ignore <target>` and `git merge --ff-only <target>` in a clone)
stdout: git: git version 2.54.0 (Apple Git-157)
stdout: before: HEAD=32719f8 ignored=[b.txt] b.txt=[local ignored content]
stdout: --- run A: git merge --ff-only --no-overwrite-ignore TARGET
stdout: error: The following untracked working tree files would be overwritten by merge:
stdout: 	b.txt
stdout: Please move or remove them before you merge.
stdout: Aborting
stdout: Updating 32719f8..90aa951
stdout: exit=1
stdout: after A: HEAD=32719f8 b.txt=[local ignored content]
stdout: --- run B: git merge --ff-only TARGET (default overwrite-ignore)
stdout: Updating 32719f8..90aa951
stdout: Fast-forward
stdout:  b.txt | 1 +
stdout:  1 file changed, 1 insertion(+)
stdout:  create mode 100644 b.txt
stdout: exit=0
stdout: after B: HEAD=90aa951 b.txt=[committed]
exit: 0 (the probe script's exit status)
tree: 366b45155 (the probe runs in a scratch repository and does not read the tree)
note: observed in revision 0.6. Run A refuses with exit 1 and leaves HEAD and the ignored file unchanged; run B overwrites the ignored file. Commit SHAs differ between probe runs (commit timestamps), so the outcome lines are the observation. This is the basis of plan §B3a step 5 and AC-LMF-003.
```

```text
E-73
command: go version -m <scratchpad>/moai-tree-novcs | grep -c -e vcs.revision -e vcs.time -e vcs.modified -e buildvcs
stdout: 0
exit: 1
tree: 366b45155
note: pipeline, not release-blocking. Zero matches: the tree build has no VCS stamp and no buildvcs build setting (its build settings are lines 79 to 88 of the go version -m output; toolchain go1.26.8). E-63 cites this cell.
```

```text
E-74
command: grep -n -o -e '표준 체인' -e '카드 브랜치는' -e '원격 기본 브랜치다' -e 'commit-dead다' -e '승인이 아니다' -e '창을 받으면:' -e '유일한 공개 경로다' -e 'Factory(리더' -e 'self-dispatch lane 예외 — 병합 창' AGENTS.local.md
stdout: 179:표준 체인
stdout: 195:카드 브랜치는
stdout: 196:원격 기본 브랜치다
stdout: 199:commit-dead다
stdout: 212:Factory(리더
stdout: 214:self-dispatch lane 예외 — 병합 창
stdout: 216:승인이 아니다
stdout: 218:창을 받으면:
stdout: 219:유일한 공개 경로다
exit: 0
tree: 366b45155 (AGENTS.local.md is unchanged since e32f69c46, E-66)
note: one match per fragment, on the clause lines named in AC-LMF-011. Each fragment occurs on one line only, so each clause check is unambiguous.
```

```text
E-75
command: grep -c '^Open:' .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/plan.md
stdout: 2
exit: 0
tree: 366b45155 (revision 0.6 plan edits uncommitted)
note: the F-15 count after revision 0.6. The two open-item lines (the OQ-10 line in §G.2 and the MI-2 line in §G.3) both begin with "Open:". Before the revision the count was 1 (revision 0.5 audit B5).
```

```text
E-76
command: grep -c '^Operator verdict:$' .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/decision-index.md
stdout: 2
exit: 0
tree: 366b45155 (revision 0.6 decision-index edits uncommitted)
note: both verdict lines are empty; Q1 and Q2 stay open (F-16).
```

```text
E-77
command: grep -c 'local_main_integration' .moai/config/sections/workflow.yaml
stdout: 0
exit: 1
tree: 366b45155
note: RED-now for the enabled check of AC-LMF-009: the repository copy has no local_main_integration key until M3 adds it. The template copy is E-01.
```

## F. File checks and proving commands for documents and configuration

Each F-entry is numbered by the criterion it serves. Each is a single shell command or a short fixed set. Every git check in this section is `[local]` unless it is marked `[origin]`. The SPEC lint and audit results, with their judging builds, are cells E-61 to E-65 in §E. Entries marked inspection require reading the matching lines and recording them in the run progress record. Criteria 002 to 006 are proved by the V-commands in plan §F and have no file check.

- **F-01 (AC-LMF-001).** `grep -n 'local_main_integration' internal/template/templates/.moai/config/sections/workflow.yaml` shows the key and `enabled: false`. F-01a (M1): `grep -c 'local_main_integration' .moai/config/sections/workflow.yaml` returns 0. F-01b (M3): `git merge-base --is-ancestor <template commit> <repository copy commit>` exits 0, with the SHAs recorded in the run progress record.
- **F-07 (AC-LMF-007).** `grep -c '^### B7' .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/plan.md` returns 1. `grep -c 't1621' .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/plan.md` returns at least 2. `git diff --quiet e32f69c46 HEAD -- internal/homestate` exits 0 at the close of the card.
- **F-08 (AC-LMF-008).** `grep -n 'develop_branch' .moai/config/sections/git-strategy.yaml` shows `main`. `grep -n 'manual:' .moai/config/sections/git-strategy.yaml` shows line 8. `grep -n 'auto_merge:' .moai/config/sections/workflow.yaml` shows `auto_merge: false`. `grep -c 'auto_merge 유지' .moai/config/sections/workflow.yaml` returns 0. `git merge-base --is-ancestor 09a42899c HEAD` exits 0 (V8).
- **F-09 (AC-LMF-009).** `grep -n 'deny_commits_on' .moai/config/sections/workflow.yaml` shows `[main]` while Q1 is open. `grep -A1 'local_main_integration:' .moai/config/sections/workflow.yaml` shows `enabled: true`.
- **F-10 (AC-LMF-010).** `grep -n -e 'commit-dead' -e 'card PRs go to base' -e 'card PR goes to base' AGENTS.md` prints nothing. `grep -c 'remote default branch' AGENTS.md` returns 1. `grep -c 'from local `main`' AGENTS.md` returns 0. `git diff --quiet e32f69c46 HEAD -- internal/template/templates/AGENTS.md.tmpl` exits 0.
- **F-11 (AC-LMF-011).** `grep -c '^### §4.0' AGENTS.local.md` returns 1. `grep -c 'SUPERSEDED by §4.0' AGENTS.local.md` returns 7, and the seven marked lines carry the clauses named in AC-LMF-011, checked by clause text (no line number is part of the check). `grep -c '카드 브랜치는 `main`에서 판다' AGENTS.local.md` returns 1.
- **F-12 (AC-LMF-012).** `grep -c 'Card worktrees branch FROM' .claude/rules/local/repo-local-pr-policy.md` returns 0. Inspection of `grep -n 'CLAUDE.local.md' .moai/docs/gitflow-integration-chain.md` and `grep -n 'CLAUDE.local.md' .claude/rules/local/gitflow-lane-protocol.md`: each line carries "retired" or "former".
- **F-13 (AC-LMF-013).** `wc -c AGENTS.md` returns fewer bytes than 24228. `wc -m AGENTS.md` returns fewer than 40000. For each run-phase commit that grows an always-loaded file by more than 1,000 bytes, `git log -1 --format=%B <commit>` contains the measured byte counts and the cost statement.
- **F-14 (AC-LMF-014).** `grep -c 'release/main-batch-YYYYMMDD' AGENTS.local.md` returns at least 1. `grep -c 'git push origin main:refs/heads/release/main-batch-' AGENTS.local.md` returns 1. `grep -c -e 'fast-forward' -e 'sync branch' AGENTS.local.md` returns at least 1.
- **F-15 (AC-LMF-015).** `grep -c '^### G\.[23] Gap' .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/plan.md` returns 2. `grep -c 'Run-phase procedure' .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/plan.md` returns 2. `grep -c '^Open:' .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/plan.md` returns at least 2.
- **F-16 (AC-LMF-016).** `grep -c '^Operator verdict:$' .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/decision-index.md` returns 2 while both decisions are open. `grep -c 'GATED-ON-DECISION' .moai/specs/SPEC-LOCAL-MAIN-FLOW-001/acceptance.md` returns at least 2. `git merge-base --is-ancestor <card tip> main` is evaluated only after the decision is recorded.
- **F-17 (AC-LMF-008, tree identity).** `git rev-parse 'e32f69c46^{tree}'` and `git rev-parse '09a42899c^{tree}'` print the same tree (E-46 and E-47).

## G. Edge cases, quality gates, and Definition of Done

Edge cases covered by the scenarios above:

- A key absent from the file reads as `false` (AC-LMF-001).
- The gate is true and the primary is on another branch: refused with guidance and no branch switch (AC-LMF-002).
- A card whose own branch is the integration branch is refused whatever the gate (AC-LMF-003, `factory_card.go:1883-1885`).
- A dirty primary holding only untracked files is refused (AC-LMF-004).
- A rename whose source is a dirty path counts as an overlap on a separate worktree (AC-LMF-005).
- A dirty path that is a directory prefix of a target path, or the reverse, counts as an overlap (AC-LMF-005).
- The status set is compared byte for byte, including untracked-file rows, on both surfaces (AC-LMF-006).
- `[origin]` BASELINE_SHA and `[local]` local main diverged: the re-sync refuses and leaves HEAD unchanged (AC-LMF-003).
- An ignored file at a path that the re-sync would change is kept, and the re-sync refuses (AC-LMF-003; plan §B3a step 5).

Quality gates for the run phase:

- Tests are written RED first. Each new test is observed failing on the unchanged code before the change lands, and the observation is recorded in the run progress record.
- Every V-command runs with `-v`, and each named test must appear as `--- PASS`. An empty sweep fails the gate.
- `gofmt -l` over the four touched packages prints nothing.
- The `internal/factory` step family runs under its slot lease (V5), and the lease is released afterwards.
- No command in the run phase contacts a remote, and no pull request is opened. The re-sync tests set `refs/remotes/origin/main` with `git update-ref` in a scratch repository (`[origin]`, scratch only).
- No commit lands on local main from this card unless the decision for Q2 has been recorded (AC-LMF-016).

Definition of Done:

- Each criterion has green-path evidence in the run progress record (the verbatim output, the command, and the tree SHA), except the parts that depend on an open decision: the value check of AC-LMF-009 (Q1), and all of AC-LMF-016 (Q1 and Q2). Those parts are reported as not evaluated until the decisions are recorded.
- Every ordered criterion was evaluated after V8 passed on its own commit.
- Every byte statement required by AC-LMF-013 is present in its commit body.
- Open decisions Q1 and Q2 remain in `decision-index.md` with an empty `Operator verdict:` until the operator decides.
- Open item MI-2 is carried to the leader with its owner. MI-4 is confirmed by the lead. OQ-7, OQ-9, and OQ-10 are open items under residual risk, owned by the leader (plan §K, R-4 to R-6). OQ-3 is closed as residual risk R-2, OQ-6 is closed (git-flow stays), and OQ-8 is redirected to the tool-owned re-sync (plan §B3a). MI-5 is resolved with seven markers.
