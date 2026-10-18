# Plan — SPEC-GIT-STRATEGY-AGENTIC-TRUNK-001

Tier L. Priority order, no time estimates. Milestones lead with the decisions most likely to change (the value's data model and reader semantics, then user-visible surfaces) and end with the mechanical prose sweep. The live-value flip is deliberately NOT adjacent to the reader change: it waits for an installed binary that understands the value (design.md section 4).

## 1. Constraints that bind every milestone

- [HARD] Verification uses anchored `-run` selectors only: `go test ./internal/<pkg> -run '^(TestA|TestB)$' -count=1 -v`. No whole-package run of `internal/cli` (it exceeds the time limit) and no whole-package run of `internal/config`: main carries a red test there (`TestNoHardcodedContexts`, repair card t1858), so a package-wide verdict would mix this SPEC's result with an unrelated red.
- [HARD] A selector that matches nothing exits 0 and prints `[no tests to run]` (observed, research R6). Every criterion names the `--- PASS:` lines it expects; an output without them is red.
- [HARD] No test skip, no `t.Skip` added, no timeout widening: the standing `-timeout 30m` of AGENTS.local.md section 4 is used as written. Existing tests whose assertion is the value-set size (4 entries) or an exact notice string are UPDATED to the new truth, never skipped.
- [HARD] Template-First: edit `internal/template/templates/**`, run `make build`, then bring the deployed copies in line; commit both in the same landing. Nothing deployed to user projects changes a default (REQ-AGT-015).
- [HARD] Tool provenance: any criterion judged by the `moai` binary invokes a binary built from the tree under test by its path (`./bin/moai` after `make build`), and the citation names that binary's commit next to the tree HEAD (verification-claim-integrity section 2.2).
- Stage by explicit pathspec; re-read `git rev-parse --short HEAD` and the branch before every commit (AGENTS.md section 2).

## 2. Milestones

### M0 — Baseline (no behavior change)

Record, in the progress record, the RED-now observation of each Go-test selector in acceptance.md on the pre-change tree (the `[no tests to run]` line per package: `internal/config` is already observed, research R6; take `internal/cli`, `internal/statusline`, `internal/cli/update`, `internal/template`), and the per-file `agentic-trunk` grep counts of the guidance files. The baseline lands in its own commit before M1 so the commit graph can witness the order (verification-claim-integrity section 2.3).

### M1 — Value set, disposition, reader (highest change likelihood: data model)

REQ-AGT-001 to REQ-AGT-005. Config package only. Add the value and its target constant, extend the classifier and the allowed set, add the reader field and the local-integration predicate, replace the second `git-flow` literal with the constant. Update `TestWorkflowDisposition` (allowed-set size 4 to 5) and add `TestAgenticTrunk*` tests. Gate: AC-001 to AC-004. This milestone stands alone: nothing else depends on guidance.

### M2 — Consumers

REQ-AGT-006 to REQ-AGT-011. Session-exit auto-merge predicate and notice, the two guidance fix strings, a pinning test for the factory delivery path, an `integration acquire` test, the ignored-keys test, and the landing-check grep guard. Update `TestAutoMergeUnconfiguredTarget` if it asserts the old notice. Gate: AC-005 to AC-010.

### M3 — Display

REQ-AGT-012, REQ-AGT-013. Doctor case and statusline label with the provider file. Gate: AC-011, AC-012. Send t1830's lane the note in section 5 before this milestone lands.

### M4 — Pins and guards

REQ-AGT-014 to REQ-AGT-016. `moai update` value-survival test, shipped-defaults test, sweep-guard arming assertion. Pre-measure the armed finding count (`MOAI_SWEEP_GUARD_ARMED=1`, one anchored run of `TestGitHubFlowSweepGuard`) and record it for the OQ-2 follow-up card. Gate: AC-013 to AC-015.

### M5 — Template guidance (Template-First)

REQ-AGT-020, REQ-AGT-021. The seven rule sources, `delivery.md` (value domain, table row, `Strategy: agentic-trunk` block without a push step), `moai-ref-git-workflow/SKILL.md` section, `workflow.yaml` comment; `make build`; deployed projections. Read `manager-git.md` and `manager-git.toml` and edit them only if they enumerate the value domain. Gate: AC-019 to AC-021.

### M6 — Cutover (operator-held step inside; commit is mechanical)

REQ-AGT-017, REQ-AGT-018. Precondition: M1 to M3 are on local main and an installed binary built from that tree passes the doctor check (design.md section 4). The leader runs the rc build and reinstall (protocol section 9) and reconnects every running `moai mcp-server`. Then one commit changes `.moai/config/sections/git-strategy.yaml` only: `workflow: agentic-trunk`, removal of the three git-flow-only keys from the manual profile, the stale comment replaced by a two-line statement of the model, `lead_push_threshold: 20` kept. Gate: AC-016, AC-017.

### M7 — Repo-local guidance

REQ-AGT-019, REQ-AGT-023. `gitflow-lane-protocol.md`, `AGENTS.local.md` section 4, `repo-local-pr-policy.md`, three `.moai/docs` notices. Lands after M6 so no sentence says the value is `agentic-trunk` while the file says `git-flow`. Reconcile with t1830 per section 5. Gate: AC-018, AC-023.

### M8 — Docs-site (sync phase, manager-docs)

REQ-AGT-022. Four `moai-sync.md` pages, Korean canonical first, then en, ja, zh in the same landing; the repo's docs verification recipe before return. Gate: AC-022.

### M9 — Close

AC-024 (hygiene) over the whole branch; progress record sections filled by their owners; sync audit.

## 3. Risk register

| Risk | Where | Mitigation |
|---|---|---|
| Live value flipped before the binary understands it: every base-branch consumer refuses an empty target | M6 | design.md section 4: separate commit, doctor of the tree-built binary first, rollback is one revert |
| A running `moai mcp-server` keeps its start-time build and refuses the new value | M6 | the cutover step reconnects every lane's server; AC-016 judges by a path-invoked tree-built binary, not the shell `moai` |
| `/moai sync` Step 3.2 stops delivery on an unknown value | M5, M6 | `delivery.md` value domain extended in M5, before M6 |
| Existing tests pin the 4-value set or the old notice string | M1, M2 | update those assertions; do not skip (REQ-AGT-024) |
| Statusline label changes bytes that t1830 pins | M3 | label only for non-github-flow valid values; coordination note; OQ-4 |
| t1830 and this card edit the same two files | M7 | section 5; whichever lands second reconciles, keeps both |
| Sweep guard armed by mistake for the new value | M4 | explicit assertion (AC-015); OQ-2 follow-up |
| Heavy test packages | all | anchored `-run` only (section 1) |

## 4. Verification layout

Each AC in acceptance.md carries a single-invocation judge (no pipes, no chaining), the tokens its stdout must contain, and a RED-now cell. Go-test judges are `go test ./<pkg> -run '^(TestA|TestB)$' -count=1 -v`; the expected stdout lists the `--- PASS:` lines by name; `no tests to run` is red. Grep judges are `grep -c` or `git grep -c` with an expected count.

## 5. Coordination with t1830 (SPEC-PRIMARY-LANDING-ONLY-001)

- Shared files: `AGENTS.local.md`, `gitflow-lane-protocol.md`, and the template rule sources t1830 names in its plan section B.1. This SPEC changes flow NAMING and the value's description; t1830 changes landing-surface RULES.
- Rule for the second lander: rebase onto the other's text, keep both. Concretely the final `gitflow-lane-protocol.md` carries t1830's Integration-Worktree Rule sentences AND this SPEC's `agentic-trunk` title, section titles and four-step statement (REQ-AGT-023, AC-023).
- Statusline: t1830's REQ-PLO-009 (byte-identical render when the checkout has a branch) is about its own change. This SPEC's label is a second, independent change to the same segment. Send t1830's lane one line: the label appears only for valid non-github-flow workflows, so a byte-identity measurement taken with a github-flow or unset fixture is unaffected; one taken in this repository after M6 will show the label.
- Landing order is not forced by this SPEC. M7 is the only milestone that touches the shared files, and it is the last code-adjacent milestone, so this SPEC naturally lands second unless t1830 stays held.
