# SPEC Review Report: SPEC-FACTORY-MIXED-HOOK-001

Audit scope: UserPromptSubmit lifecycle correction delta, D1 re-audit for card `t1074`

Iteration: 2/2 for this correction delta

Verdict: **PASS**

Overall Score: **1.00**

Tier/threshold: Tier L by absent-`tier` backward-compatibility rule; PASS threshold `0.85`.

Reasoning context ignored per M1 Context Isolation. This re-audit inspected only the corrected artifacts, the current selector/test surface needed to verify the prior defect, the prior exported defect statement, and repository audit rules.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency**: `spec.md:L55-L101` remains sequential `REQ-FMH-001..012`; `spec.md:L109-L116` remains sequential `REQ-FMH-OPS-001..008`, with no gap or duplicate in either namespace.
- [PASS] **MP-2 GEARS format compliance — requirement layer only**: the lifecycle behavior remains canonical at `spec.md:L57` and `spec.md:L115`; the acceptance Given/When/Then layer was not misgraded as a requirement.
- [PASS] **MP-3 YAML frontmatter validity**: `spec.md:L1-L15` still contains all 12 canonical fields with valid names and types.
- [N/A] **MP-4 language neutrality**: this is a single-repository Go/Codex/Claude integration, not universal 16-language template-bound tooling.
- [PASS] **MP-5 D7 cross-SPEC reconciliation**: no retired/superseded/archived cross-SPEC dependency was introduced by the D1 correction.
- [PASS] **MP-6 D8 cross-platform discipline**: no literal `syscall` occurs in `spec.md`.
- [PASS] **MP-7 clarification gate**: `plan.md` and `research.md` contain no `[NEEDS CLARIFICATION]` marker.
- [PASS, delta-scoped] **MP-8 RED-now re-execution**: both correction selectors at `acceptance.md:L57-L58` reproduced RED with empty stdout and exit 1 on the current working tree.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---|---|
| Clarity | 1.00 | single unambiguous evidence allocation | `acceptance.md:L37`, `plan.md:L64`, addendum `L160/L170/L180` explicitly divide unit versus LIVE responsibility. |
| Completeness | 1.00 | all required artifacts/sections present | The D1 correction covers the AC summary, plan, selector ledger, implementation plan, binary AC, and both gate descriptions. |
| Testability | 1.00 | every corrected predicate has a binary owner | Unit owns empty/whitespace no-bind; LIVE owns real non-empty bind-before-inbox and bound no-write. |
| Traceability | 1.00 | complete | Mechanical count remains `requirements=20 referenced=20 acceptance=21`, with no uncovered or unknown requirement. |

## Defects Found

No defects found.

## Regression Check

Defect from the previous iteration:

- **D1 — AC-FMH-OPS-006 assigned LIVE proof to a behavior the LIVE procedure forbade observing — RESOLVED.**
  - `acceptance.md:L37` now states that combined unit + LIVE evidence is required and names `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer` as the empty/whitespace no-bind owner.
  - `acceptance.md:L57` labels that selector as the required fallback and empty-input guard.
  - `plan.md:L64` explicitly states that LIVE does not send empty/whitespace input and confines that proof to the unit selector.
  - `operational-lane-status-addendum.md:L118` requires the named unit test to check both empty and whitespace-only inputs while preserving launch-pending generation/identity/`updated_at`.
  - `operational-lane-status-addendum.md:L160` makes §6.1 unit + §6.2 LIVE jointly necessary and assigns each behavior to one surface.
  - `operational-lane-status-addendum.md:L170` places empty/whitespace no-bind responsibility solely in the eight-selector unit gate.
  - `operational-lane-status-addendum.md:L180` explicitly excludes empty/whitespace proof from LIVE and retains LIVE responsibility for real non-empty rebind and bound no-write.

The corrected allocation is internally consistent: the normative M4 prohibition at addendum `L127` is no longer in conflict with the AC because LIVE no longer claims to exercise empty/whitespace input.

## Evidence

### Claim

The prior blocking D1 is resolved without weakening the behavior: AC-FMH-OPS-006 now requires a binary combined package in which unit evidence proves the input guard and LIVE evidence proves the production lifecycle.

### Baseline identity

Command:

```bash
git rev-parse --show-toplevel && git rev-parse --short HEAD && git branch --show-current
```

Verbatim output:

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074
8c5d9be99
WT-factory-mixed-hook
```

All six plan artifacts remain modified, so content attribution is `HEAD 8c5d9be99 + the working-tree diff observed in this audit`, not HEAD alone.

### Traceability

Command: extract unique requirement IDs from `spec.md`, unique requirement references and AC IDs from the acceptance table, then compare sets.

Verbatim output:

```text
requirements=20 referenced=20 acceptance=21
uncovered_requirements:
unknown_requirement_refs:
```

### Gate inventory

Command: extract unique selector names from addendum line 167 and required lifecycle sentinel names from line 177.

Verbatim output:

```text
unit_unique_selectors=8
TestFactoryBoundUserPromptSubmitDoesNotRewritePeer
TestFactoryLaneRosterProjectIsolation
TestFactoryLaneRosterStateTruth
TestFactoryLauncherRegistersLaunchPendingPeers
TestFactoryLeadNoticeUsesOperationalStatus
TestFactoryMsgStatusReadOnlyRoster
TestFactorySessionStartRebindsLaunchPendingPeer
TestFactoryUserPromptSubmitRebindsLaunchPendingPeer
required_lifecycle_sentinels=2
BOUND_PROMPT_NO_REWRITE_OK
USERPROMPT_REBIND_OK
```

### Current selector presence/RED state

Verbatim output:

```text
TestFactoryLaneRosterStateTruth=1
TestFactoryLaneRosterProjectIsolation=1
TestFactoryMsgStatusReadOnlyRoster=1
TestFactoryLeadNoticeUsesOperationalStatus=1
TestFactoryLauncherRegistersLaunchPendingPeers=1
TestFactorySessionStartRebindsLaunchPendingPeer=1
TestFactoryUserPromptSubmitRebindsLaunchPendingPeer=0
TestFactoryBoundUserPromptSubmitDoesNotRewritePeer=0
TestFactoryLiveOperationalRosterBeforePrompt=1
TestFactoryLiveOperationalLauncherChain=1
```

Re-executed RED commands:

```text
$ rg -n -F 'func TestFactoryUserPromptSubmitRebindsLaunchPendingPeer(' internal
<empty>
exit=1

$ rg -n -F 'func TestFactoryBoundUserPromptSubmitDoesNotRewritePeer(' internal
<empty>
exit=1
```

### Exact eight-selector unit gate

The exact §6.1 Go/JQ gate was executed using isolated `MOAI_HOME` and `GOCACHE`.

Verbatim final output:

```text
false
exit=1
```

Bounded event summary:

```json
{
  "named_passes": [
    "TestFactoryLaneRosterProjectIsolation",
    "TestFactoryLaneRosterStateTruth",
    "TestFactoryLauncherRegistersLaunchPendingPeers",
    "TestFactoryLeadNoticeUsesOperationalStatus",
    "TestFactoryMsgStatusReadOnlyRoster",
    "TestFactorySessionStartRebindsLaunchPendingPeer"
  ],
  "skips": 0,
  "fails": 0,
  "not_run": 0
}
```

This is a valid plan RED: six existing selectors passed, both new lifecycle selectors were absent, and the exact count gate refused closure.

### Empty-pass and LIVE-sentinel mutant check

The unit/LIVE predicates were evaluated against: complete unit control, empty event list, one-test input, skip mutant, complete two-sentinel LIVE control, missing-no-rewrite sentinel, and `NOT_RUN` mutant.

Verbatim output:

```text
[
  true,
  false,
  false,
  false,
  true,
  false,
  false
]
```

The predicates accept only the complete controls and reject empty/partial unit evidence, skip, missing one of the two required LIVE sentinels, and `NOT_RUN`.

### Format/clarification/D8 checks

Commands:

```bash
rg -n '\[NEEDS CLARIFICATION' .moai/specs/SPEC-FACTORY-MIXED-HOOK-001/{plan.md,research.md}
rg -n 'syscall' .moai/specs/SPEC-FACTORY-MIXED-HOOK-001/spec.md
git diff --check -- .moai/specs/SPEC-FACTORY-MIXED-HOOK-001
```

Verbatim output: all three commands produced `<empty>`; the first two have no matches and `git diff --check` exited 0.

## Baseline-attribution

- Worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`
- Branch: `WT-factory-mixed-hook`
- HEAD: `8c5d9be99`
- Subject: current uncommitted plan-artifact correction diff.
- Installed `moai` CLI was not used.

## Verified Non-findings

- No remaining claim that LIVE sends or proves empty/whitespace no-bind.
- No weakening of the empty/whitespace behavior itself; the unit owner must verify both empty and whitespace-only forms and unchanged peer state.
- No loss of real production proof: LIVE still requires `USERPROMPT_REBIND_OK` and `BOUND_PROMPT_NO_REWRITE_OK` for all three lanes.
- No conflation of SessionStart early bind with the required UserPromptSubmit fallback.
- No empty, partial, skip, fail, or `NOT_RUN` acceptance in the checked predicates.
- No selector-count drift: exactly eight unique unit selectors.
- No lifecycle-sentinel drift: exactly two required lifecycle sentinels.
- No traceability drift: all 20 requirements are referenced by the 21 acceptance criteria.
- No t1074/t1082/t1075 boundary regression was introduced by the D1 repair.

## Gaps

- LIVE gates were not executed in this plan re-audit; this PASS approves the corrected plan and evidence allocation, not runtime completion.
- The two new implementation selectors remain RED by design and must be implemented before the run gate can pass.
- Existing AC-FMH-001..015 runtime regressions were not rerun; this was a defect-delta re-audit under the iteration contract.

## Residual-risk

- A future unit test with the required name but without both empty and whitespace-only subcases could satisfy the name/cardinality gate while violating addendum `L118/L170`; sync audit must inspect test semantics, not only names.
- LIVE sentinel emission must remain after the underlying assertions; a log-only sentinel would not prove the behavior.
- Unit evidence cannot substitute for the production launcher chain, and LIVE evidence cannot substitute for the empty-input guard; the final verdict must require both gates.

## Recommendation

Proceed to implementation of the two RED selectors and the UserPromptSubmit bind-before-inbox/no-write seam. Preserve the explicit unit + LIVE evidence split during sync verification.
