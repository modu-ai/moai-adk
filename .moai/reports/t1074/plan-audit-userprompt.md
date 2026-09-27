# SPEC Review Report: SPEC-FACTORY-MIXED-HOOK-001

Audit scope: UserPromptSubmit lifecycle correction delta for card `t1074`

Iteration: lifecycle-correction delta after `plan-audit-revision.md`

Verdict: **FAIL**

Overall Score: **0.875**

Tier/threshold: Tier L by absent-`tier` backward-compatibility rule; PASS threshold `0.85`. The score exceeds the numerical threshold, but one unresolved blocking consistency/testability defect prevents PASS.

Reasoning context ignored per M1 Context Isolation. This audit used only the six named plan artifacts, the current tree/code/tests needed to verify their claims, the repository audit rules, and the official Codex hooks contract.

## Plausible failure modes checked first

- `SessionStart` remained specified as the required first-normal-turn signal.
- `UserPromptSubmit` binding was ordered after inbox claim rather than before it.
- empty/whitespace prompts could bind a provisional endpoint.
- repeated prompts could refresh `updated_at` or rewrite generation/peer identity.
- early-bind and required-fallback tests could be conflated.
- selector/JQ gates could pass on zero tests, skip, fail, or `NOT_RUN`.
- LIVE sentinel names could still encode the disproved SessionStart premise.
- 20 REQ / 21 AC traceability or the t1074/t1082/t1075 boundary could have drifted.
- the AC evidence allocation could require a behavior that its own LIVE procedure forbids observing.

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency**: `spec.md:L55-L101` contains sequential `REQ-FMH-001..012`; `spec.md:L109-L116` contains sequential `REQ-FMH-OPS-001..008`. No gaps or duplicates were observed in either declared namespace.
- [PASS] **MP-2 GEARS format compliance — requirement layer only**: the canonical requirement entries use `SHALL`/`When ... SHALL` forms, including the lifecycle obligation at `spec.md:L57` and the OPS production obligation at `spec.md:L115`. The Given/When/Then verification layer in the addendum was not graded as a requirement-layer defect.
- [PASS] **MP-3 YAML frontmatter validity**: `spec.md:L1-L15` contains all 12 canonical fields with valid types and names: `id`, `title`, quoted `version`, valid `status`, ISO `created`/`updated`, `author`, valid `priority`, release-target `phase`, path-like `module`, valid `lifecycle`, and comma-string `tags`. No rejected aliases were found.
- [N/A] **MP-4 language neutrality**: this is a Go-repository, Codex/Claude factory integration SPEC rather than universal 16-language template-bound tooling.
- [PASS] **MP-5 D7 cross-SPEC reconciliation**: the only SPEC-ID matched across the audited artifacts was the subject `SPEC-FACTORY-MIXED-HOOK-001`; its current `status` is `in-progress`, not retired/superseded/archived. No unresolved D7 BLOCKING finding was emitted.
- [PASS] **MP-6 D8 cross-platform discipline**: `rg` found no literal `syscall` in `spec.md`; D8 auto-passes.
- [PASS] **MP-7 clarification gate**: `rg -n '\[NEEDS CLARIFICATION' plan.md research.md` returned no matches.
- [PASS, delta-scoped] **MP-8 RED-now re-execution**: the two new release-blocking lifecycle selectors cited at `acceptance.md:L57-L58` were re-executed against the current working tree and both reproduced RED with empty stdout and exit 1. The original 15 ACs and previously introduced OPS roster criteria were not re-audited from scratch in this lifecycle-correction delta.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---|---|
| Clarity | 0.75 | one material ambiguity/contradiction | `plan.md:L64` and `acceptance.md:L37` assign empty/whitespace proof to the LIVE chain, but `operational-lane-status-addendum.md:L127` prohibits that input. |
| Completeness | 1.00 | all required Tier L artifacts/sections present | HISTORY/WHY/WHAT/HOW/requirements/out-of-scope are at `spec.md:L19-L126`; plan, acceptance, design, and research were all read. |
| Testability | 0.75 | one criterion lacks an executable evidence path for one claimed predicate | `operational-lane-status-addendum.md:L175-L180` has required rebind/no-rewrite sentinels but no empty/whitespace no-bind sentinel or procedure. |
| Traceability | 1.00 | complete | Mechanical count produced `requirements=20 referenced=20 acceptance=21`, with empty uncovered/unknown lists. |

## Defects Found

### D1. AC-FMH-OPS-006 assigns LIVE proof to a behavior the LIVE procedure forbids observing

- **Artifact/lines**: `acceptance.md:L13`, `acceptance.md:L37`, `plan.md:L64`, `operational-lane-status-addendum.md:L127`, `operational-lane-status-addendum.md:L158-L160`, `operational-lane-status-addendum.md:L175-L180`
- **Severity**: major
- **Class**: blocking
- **Confidence**: high
- **Merge-blocking**: yes
- **Description**: `acceptance.md:L37` says a newly installed binary and restarted lead MCP prove “no bind on empty/whitespace input.” `plan.md:L64` likewise says the LIVE chain must prove it. However, the normative LIVE procedure at addendum `L127` expressly prohibits empty/whitespace prompts, AC-FMH-OPS-006's scenario at `L158` sends only a non-empty first request and a normal subsequent request, and the exact LIVE predicate at `L177` requires `USERPROMPT_REBIND_OK` and `BOUND_PROMPT_NO_REWRITE_OK` but has no empty/whitespace no-bind sentinel. The unit selector may test empty input, but `acceptance.md:L13` states that a fixture proves only its named unit contract while LIVE rows require real separate CLI/model contexts. Therefore the current unit+LIVE gates cannot jointly establish the AC as written without an explicit evidence-allocation rule.
- **Impact**: the exact gates can return PASS while the installed-binary LIVE claim “no bind on empty/whitespace input” remains unobserved. Conversely, an implementer who follows the normative “do not send empty/whitespace” procedure cannot satisfy the literal AC. This makes completion non-binary and permits a false-positive operational verdict.
- **Required fix**: choose and state one evidence boundary consistently. Recommended minimum: make empty/whitespace no-bind a unit/hook contract owned by `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer` (or a separately named exact selector), and revise `acceptance.md:L37`, `plan.md:L64`, addendum AC-FMH-OPS-006, and §6 prose to say AC-FMH-OPS-006 is satisfied by the explicit **unit + LIVE combination**. Keep LIVE responsible for real non-empty rebind, inbox-before ordering, and bound-endpoint no-rewrite. If installed-binary LIVE proof is intentionally retained, add a legitimate non-synthetic procedure that can actually emit the event and an exact `EMPTY_PROMPT_NO_BIND_OK` sentinel; merely adding a sentinel without an observable procedure is insufficient.

## Lifecycle Contract Findings

### 1. SessionStart is no longer the required normal-turn signal — PASS

The official hooks page classifies `SessionStart` as a session/subagent-start event and limits its `source` values to `startup`, `resume`, `clear`, and `compact`; it classifies `UserPromptSubmit` as a during-turn event with the prompt “about to be sent” and common `session_id`. Source: <https://developers.openai.com/codex/hooks>.

The audited documents match that boundary:

- `research.md:L35` explicitly says SessionStart is not a per-normal-user-turn event.
- `design.md:L45-L46` limits SessionStart to best-effort early bind and makes non-empty UserPromptSubmit required.
- `spec.md:L57` says correctness does not depend on SessionStart ordering.
- `plan.md:L30-L32` and addendum `L112-L118` preserve the same split.

No requirement making startup SessionStart the mandatory first-normal-turn signal was found.

### 2. Required ordering and identity — PASS

`spec.md:L57`, `plan.md:L31`, `design.md:L32/L46`, research `L67`, and addendum `L74/L116/L154` consistently specify:

`launcher launch-pending` → first legitimate non-empty `UserPromptSubmit` → actual `session_id` + resolved owner PID/process-start atomic bind → same hook inbox processing.

The current code is still correctly RED for this amendment: `internal/hook/user_prompt_submit.go:L108-L115` calls `factoryHookBatch` without the required preceding bind, while `internal/hook/session_start.go:L458-L466` still owns the existing bind. This is implementation work to perform, not a plan contradiction.

### 3. Empty input and no-write steady state contract — PASS in requirements, FAIL in evidence allocation

- Normative behavior is consistent at `spec.md:L57`, `plan.md:L31-L32`, `design.md:L32/L46/L58`, research `L22/L67`, and addendum `L74/L116/L118`.
- The current `RegisterPeer` path unconditionally upserts `updated_at` at `internal/factorymsg/store.go:L367`, matching `research.md:L22` and justifying a pre-write identical-endpoint guard.
- The two required new selectors are RED as documented.
- D1 prevents the empty-input portion of the installed-LIVE AC from being testable as written.

### 4. Early bind versus fallback tests — PASS

The selector ledger explicitly distinguishes:

- `TestFactorySessionStartRebindsLaunchPendingPeer`: best-effort early bind, present but not the required fallback (`acceptance.md:L56`).
- `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer`: required fallback, current RED (`acceptance.md:L57`).
- `TestFactoryBoundUserPromptSubmitDoesNotRewritePeer`: bound steady state, current RED (`acceptance.md:L58`).

The exact unit gate includes all three distinct selectors plus launcher registration and four roster/status selectors.

### 5. Gate cardinality and mutant resistance — PASS, subject to D1

The unit gate requires exactly eight matching pass events and globally rejects skip/fail/`NOT_RUN` (`operational-lane-status-addendum.md:L167-L170`). Synthetic predicate mutation produced:

```text
[
  true,
  false,
  false,
  true,
  false,
  false
]
```

The six values respectively represent: complete 8-test unit control; one-test empty-pass mutant; unit skip mutant; complete two-sentinel LIVE control; missing `BOUND_PROMPT_NO_REWRITE_OK`; and `NOT_RUN`. Thus the exact predicates reject missing cardinality, skip, missing required sentinel, and `NOT_RUN`.

The required corrected LIVE sentinels `USERPROMPT_REBIND_OK` and `BOUND_PROMPT_NO_REWRITE_OK` are present at addendum `L177`. The current implementation still prints the old `SESSIONSTART_REBIND_OK` at `internal/cli/factory_operational_live_test.go:L166`, so the corrected gate remains RED rather than falsely passing.

### 6. Traceability and card boundaries — PASS

Observed traceability output:

```text
requirements=20 referenced=20 acceptance=21
uncovered_requirements:
unknown_requirement_refs:
```

The boundary is consistently preserved:

- t1074: initial launcher `launch-pending → bound` and hook-boundary delivery (`spec.md:L50`, `plan.md:L20/L79`).
- t1082: worktree creation, `/cd` or headless cwd handoff, and later endpoint transition/rebind (`spec.md:L50/L123`, `design.md:L34-L36`, `plan.md:L79`).
- t1075: idle wake only after the current endpoint is BOUND (`spec.md:L50/L122`, `design.md:L36`, `plan.md:L80`).

## Evidence

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

The six audited plan artifacts were modified in the working tree. Therefore this report attributes content to `HEAD 8c5d9be99 + the observed working-tree diff`, not to HEAD alone.

### Current selector inventory

Command: exact function-name counts under `internal`.

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

### RED-now re-execution

Commands and verbatim outputs:

```text
$ rg -n -F 'func TestFactoryUserPromptSubmitRebindsLaunchPendingPeer(' internal
<empty>
exit=1

$ rg -n -F 'func TestFactoryBoundUserPromptSubmitDoesNotRewritePeer(' internal
<empty>
exit=1
```

### Exact unit-gate execution on the current tree

The exact eight-selector Go/JQ gate from addendum §6.1 was executed with isolated `MOAI_HOME` and `GOCACHE`.

Verbatim final output and exit:

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

This is the expected amendment RED: six named tests pass, while the two required UserPromptSubmit selectors do not exist, so the exact cardinality gate refuses closure.

### Formatting check

Command:

```bash
git diff --check -- .moai/specs/SPEC-FACTORY-MIXED-HOOK-001
```

Verbatim output: `<empty>`; exit `0`.

## Baseline-attribution

- Repository/worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`
- Branch: `WT-factory-mixed-hook`
- HEAD: `8c5d9be99`
- Subject: the uncommitted six-artifact lifecycle correction diff observed in this audit run.
- Official contract read in this audit run: <https://developers.openai.com/codex/hooks> (redirected official documentation page).
- Installed `moai` CLI output was not used; the installed-binary lag warning therefore does not affect this verdict.

## Verified Non-findings

- No stale requirement that SessionStart must fire on the first normal prompt.
- No inbox-before-bind ordering in the plan artifacts.
- No requirement allowing empty/whitespace prompt binding.
- No requirement allowing repeated identical bound prompts to refresh `updated_at`.
- No conflation of the SessionStart early-bind selector with the UserPromptSubmit fallback selector.
- No zero-test, skip, fail, or `NOT_RUN` acceptance in the exact revised unit/LIVE predicates tested.
- No missing or orphaned requirement in the 20-REQ / 21-AC map.
- No t1074 expansion into t1082 worktree handoff or t1075 idle wake.
- No `[NEEDS CLARIFICATION]` marker, D7 retired-SPEC conflict, or D8 `syscall` issue.

## Gaps

- The production LIVE tests were not executed in this plan audit; current implementation inspection and selector execution are plan-phase RED evidence, not operational PASS evidence.
- The original 15 AC runtime gates and historical prior audit findings were not re-audited from scratch; this report is scoped to the UserPromptSubmit lifecycle correction delta plus full REQ/AC traceability.
- The official documentation establishes event fields and lifecycle classes, not MoAI's atomic database semantics; those semantics remain implementation/test obligations.

## Residual-risk

- After D1 is repaired, the implementation can still order bind incorrectly unless the test observes that inbox lookup succeeds in the same UserPromptSubmit event immediately after rebind.
- A no-write test that checks only generation can miss an `updated_at` rewrite; the selector must compare the full peer snapshot named by the SPEC.
- Codex hook trust and real process ownership are runtime conditions; unit green does not replace the required production launcher proof.

## Recommendation

1. Revise the empty/whitespace evidence allocation exactly as described in D1, preferably as an explicit unit + LIVE composition.
2. Preserve the two corrected LIVE sentinels and the eight-test exact-count predicate.
3. Re-run this delta audit after the four affected wording surfaces agree; then implement the two RED selectors and the UserPromptSubmit-before-inbox seam.
