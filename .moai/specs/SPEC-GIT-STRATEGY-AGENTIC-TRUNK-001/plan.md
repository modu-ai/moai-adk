# Plan — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001

Tier L. Priority order, no time estimates. Milestones lead with the decisions most likely to change (the value's data model and reader semantics, then user-visible surfaces) and end with the mechanical prose sweep. The live-value flip is deliberately NOT adjacent to the reader change: it waits for an installed binary that understands the value (design.md section 4).

## 1. Constraints that bind every milestone

- [HARD] Verification uses anchored `-run` selectors only: `go test ./internal/<pkg> -run '^(TestA|TestB)$' -count=1 -v`. No whole-package run of `internal/cli` (it exceeds the time limit) and no whole-package run of `internal/config`: main carries a red test there (`TestNoHardcodedContexts`, repair card t1858), so a package-wide verdict would mix this SPEC's result with an unrelated red.
- [HARD] A selector that matches nothing exits 0 and prints `[no tests to run]` (observed, research R6; acceptance.md ledger E01 to E17). Every criterion names the `--- PASS:` lines it expects; an output without them is red. A flip judge selects only new test names; existing tests go in a separate guard judge, because a mixed selector prints a passing line today.
- [HARD] No test skip, no `t.Skip` added, no timeout widening: the standing `-timeout 30m` of AGENTS.local.md section 4 is used as written. Existing tests whose assertion is the value-set size (4 entries) or an exact notice string are UPDATED to the new truth, never skipped.
- [HARD] Template-First: edit `internal/template/templates/**`, run `make build`, then bring the deployed copies in line; commit both in the same landing. Nothing deployed to user projects changes a default (REQ-AGT-015).
- [HARD] Tool provenance: any criterion judged by the `moai` binary invokes a binary built from the tree under test by its path (`./bin/moai` after `make build`), and the citation names that binary's commit next to the tree HEAD (verification-claim-integrity section 2.2).
- Stage by explicit pathspec; re-read `git rev-parse --short HEAD` and the branch before every commit (AGENTS.md section 2).

## 2. Milestones

### M0 — Baseline (no behavior change)

The RED-now cells are authored at plan time, on tree `f4175587119b5cfeb8d8348fc74cada8764e89ed`, in acceptance.md (ledger E01 to E21 and the inline observations); the Go selectors of every package, the guidance counts and the per-pair parity state (seven `cmp -s` pairs exit 0, `factory-dispatch.md` exits 1 by design) are already observed there. M0 therefore re-checks rather than first-measures: it diffs the pinned paths against the run-phase start tree (`git diff --stat f4175587119b5cfeb8d8348fc74cada8764e89ed..HEAD -- internal .claude AGENTS.local.md docs-site .moai/config .moai/docs`), re-observes only a selector whose package changed, and records the result in the progress record. The baseline lands in its own commit before M1 so the commit graph can witness the order (verification-claim-integrity section 2.3).

### M1 — Value set, disposition, reader (highest change likelihood: data model)

REQ-AGT-001 to REQ-AGT-005. Config package only. Add the value and its target constant, extend the classifier and the allowed set, add the reader field and the local-integration predicate, replace the second `git-flow` literal with the constant. Update `TestWorkflowDisposition` (allowed-set size 4 to 5) and add `TestAgenticTrunk*` tests, including the loader projection test of AC-025 Judge 1 (`DevelopBranch` is `main` under manual agentic-trunk with and without a `develop_branch` key). Gate: AC-001 to AC-004 and AC-025 Judge 1. This milestone stands alone: nothing else depends on guidance.

### M2 — Consumers

REQ-AGT-006 to REQ-AGT-011 and REQ-AGT-025. Session-exit auto-merge predicate and notice, the two guidance fix strings, a pinning test for the factory delivery path, an `integration acquire` test, the ignored-keys test, the landing-check grep guard, and the three consumer tests of the legacy develop-branch seam (closure push guard, contract push check, card base; design section 2 rows). Update `TestAutoMergeUnconfiguredTarget` if it asserts the old notice. Gate: AC-005 to AC-010 and AC-025 Judges 2 to 4.

### M3 — Display

REQ-AGT-012, REQ-AGT-013. Doctor case and statusline label with the provider file. Gate: AC-011, AC-012 (AC-012 carries the scoping against SPEC-PRIMARY-LANDING-ONLY-001 REQ-PLO-009; section 5).

### M4 — Pins and guards

REQ-AGT-014 to REQ-AGT-016. `moai update` value-survival test, shipped-defaults test, and the dedicated sweep-guard test `TestAgenticTrunkSweepGuardNotArmed` (three synthetic roots through `sweepGuardArmed`, `MOAI_SWEEP_GUARD_ARMED` cleared in the test; the existing `arming-switch` subtest prints a pass today and is only the guard judge). Pre-measure the armed finding count (`MOAI_SWEEP_GUARD_ARMED=1`, one anchored run of `TestGitHubFlowSweepGuard`) and record it for the OQ-2 follow-up card. Gate: AC-013 to AC-015.

### M5 — Template guidance (Template-First)

REQ-AGT-020, REQ-AGT-021. The six rule files (`factory-dispatch-mechanics`, `factory-dispatch`, `factory-dispatch-core`, `worktree-integration`, `delivery-policy`, `spec-workflow`), `delivery.md` (value domain, table row, the stop-on-unmatched-value sentence on line 29 extended to three values, a `Strategy: agentic-trunk` block after the git-flow block and without a push step), the `moai-ref-git-workflow/SKILL.md` section, and the template `workflow.yaml` comment; `make build`; deployed projections of the eight pairs. `manager-git.md`, `manager-git.toml`, `auto-semantics.md`, `spec-workflow-detail.md`, `tab_schema.json` and `git-strategy.yaml.tmpl` are excluded by name (design section 5). The `factory-dispatch.md` pair keeps its recorded divergence: the deployed copy keeps the live-only tail sentence of line 222 (removed from the template on purpose by commit `7ff326d99`, card t1815) and both copies take the same `names git-flow or agentic-trunk` edit. Gate: AC-019 to AC-021.

### M6 — Cutover (leader-executed; the commit is mechanical)

REQ-AGT-017, REQ-AGT-018. Executed by the leader, with the operator's confirmation, in a window with no lane mid-landing (replacing the installed binary affects every session; leader ruling, OQ-3). One precondition list, the same P1 to P5 as REQ-AGT-017 and design.md section 4: (P1) the M1 to M5 commits are ancestors of the flip commit; (P2) the installed binary is byte-identical to the build of the flip commit's parent (`scripts/verify-local-install.sh`, protocol section 9 build and reinstall); (P3) the installed binary's doctor reports the `agentic-trunk` row `ok` on the fixture project; (P4) every running `moai mcp-server` is reconnected and `moai doctor --check "MCP Server Version"` reads `ok` against the new commit; (P5) the operator's confirmation is recorded and `moai integration status` shows the window free. The leader records the commands and verbatim outputs of P1 to P5 in the progress record. Then one commit changes `.moai/config/sections/git-strategy.yaml` only: `workflow: agentic-trunk`, removal of the three git-flow-only keys from the manual profile, the stale comment replaced by a two-line statement of the model, `lead_push_threshold: 20` kept. Gate: AC-016, AC-017.

### M7 — Repo-local guidance

REQ-AGT-019, REQ-AGT-023. `gitflow-lane-protocol.md`, `AGENTS.local.md` section 4, `repo-local-pr-policy.md`, three `.moai/docs` notices. Lands after M6 so no sentence says the value is `agentic-trunk` while the file says `git-flow`. Before the first edit, record `N0` (the per-file count of `Integration-Worktree Rule`) in the progress record; reconcile with t1830 per section 5. Gate: AC-018, AC-023.

### M8 — Docs-site (sync phase, manager-docs)

REQ-AGT-022. Four `moai-sync.md` pages, Korean canonical first, then en, ja, zh in the same landing; the repo's docs verification recipe before return. Gate: AC-022.

### M9 — Close

AC-024 (hygiene) over the whole branch; progress record sections filled by their owners; sync audit.

## 3. Risk register

| Risk | Where | Mitigation |
|---|---|---|
| Live value flipped before the binary understands it: every base-branch consumer refuses an empty target | M6 | design.md section 4 and AC-016 P1 to P5: ancestry of M1 to M5, installed binary byte-identical to the parent's build, installed doctor `ok` on a fixture; separate commit; rollback is one revert |
| A running `moai mcp-server` keeps its start-time build and refuses the new value | M6 | P4: the leader reconnects every server and `moai doctor --check "MCP Server Version"` must read `ok` against the new commit |
| Installed-binary replacement affects every session on the machine | M6 | leader-executed with the operator's confirmation and a free integration window (P5); no lane performs it |
| `/moai sync` Step 3.2 stops delivery on an unknown value | M5, M6 | `delivery.md` value domain AND its line-29 stop sentence extended to three values in M5, before M6 (AC-020 Judges 4 and 5) |
| Existing tests pin the 4-value set or the old notice string | M1, M2 | update those assertions; do not skip (REQ-AGT-024) |
| Statusline label changes bytes that t1830 pins | M3 | label only for non-github-flow valid values; coordination note; OQ-4 |
| t1830 and this card edit the same two files | M7 | section 5; whichever lands second reconciles, keeps both |
| Sweep guard armed by mistake for the new value | M4 | dedicated test `TestAgenticTrunkSweepGuardNotArmed` (AC-015); OQ-2 follow-up |
| Key removal at the flip leaves the closure push guard, the contract push check and the card base without an integration branch | M1, M2 | the loader projects `main` for manual agentic-trunk; consumer tests of AC-025 pin the outcome with the key absent |
| Template edit lands without the deployed copy, or deletes the live-only sentence of `factory-dispatch.md` line 222 | M5 | AC-021 Judge B: sentinel in both copies, tail sentence only in the deployed copy, exactly one differing line |
| Heavy test packages | all | anchored `-run` only (section 1) |

## 4. Verification layout

Each AC in acceptance.md carries a single-invocation judge (no pipes, no chaining), the tokens its stdout must contain, and a RED-now cell authored at plan time (command, verbatim stdout, exit code, tree SHA; Go no-match runs sit in the evidence ledger E01 to E21). Go-test judges are `go test ./<pkg> -run '^(TestA|TestB)$' -count=1 -v`; the expected stdout lists the `--- PASS:` lines by name; `no tests to run` is red. A flip judge selects only new test names and an existing test gets its own guard judge. Grep judges are `grep -c` or `git grep -c` with an expected count; multi-file output is read by label, not by order.

## 5. Coordination with t1830 (SPEC-PRIMARY-LANDING-ONLY-001)

- Shared files: `AGENTS.local.md`, `gitflow-lane-protocol.md`, and the template rule sources t1830 names in its plan section B.1. This SPEC changes flow NAMING and the value's description; t1830 changes landing-surface RULES.
- Rule for the second lander: rebase onto the other's text, keep both. Concretely both `gitflow-lane-protocol.md` and `AGENTS.local.md` carry the landing-surface rule (today the landed t1846 phrase `performs landing and resync only`, after t1830 its `Integration-Worktree Rule` label) AND this SPEC's four-step sentinel; the protocol also carries the `agentic-trunk` title (REQ-AGT-023, AC-023). AC-023 has no not-applicable case: its sentinels are judged in both files whichever card lands first, and M7 records the per-file count of `Integration-Worktree Rule` before editing (N0) so nothing t1830 introduced is removed.
- Statusline: t1830's REQ-PLO-009 (byte-identical render when the checkout has a branch) is about its own change. This SPEC's label is a second, independent change to the same segment. The scoping is an artifact of this SPEC, not a message: AC-AGT-012 states that `TestAgenticTrunkStatuslineAbsent` pins the github-flow and unset cases REQ-PLO-009 measures, and that only a valid non-github-flow value shows the label (this repository after M6). The leader may also relay the one-line note to t1830's lane; the relay is not a precondition.
- Landing order is not forced by this SPEC. M7 is the only milestone that touches the shared files, and it is the last code-adjacent milestone, so this SPEC naturally lands second unless t1830 stays held.
