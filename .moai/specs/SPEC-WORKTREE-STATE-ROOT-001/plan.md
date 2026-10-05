# SPEC-WORKTREE-STATE-ROOT-001 — Plan

Card t1213 (class C, Tier M). Base: develop `c630de892`. Worktree `.claude/worktrees/t1213`,
branch `WT-worktree-state-roots`. Revision 0.2.0 answers plan-audit iter-1
(`.moai/reports/t1213/plan-audit-iter1.md`, FAIL 0.71).

## §A Context

SPEC-MCP-WORKTREE-UNTRACKED-001 (card t1202, completed) deferred configuration,
catalogue, and state routing of a config-orphaned linked worktree to this card. Its
plan audit iter-2 raised D18 (the state-destination default rested on an unmeasured
reader root) and D25 (routing the catalogue to the primary would hide worktree-local
SPECs while reporting success). The orchestrator re-measured both on `c630de892`:
`.moai/reports/t1213/remeasure-d18-d25.md`.

What the re-measurement establishes, and what it does not, is summarised in spec.md
§1.1. In one line: readers inside one hook process were measured to disagree (an
env-first reader wrote to `P`, a working-directory reader wrote to `W`); the root that
the receipt guard and the two Stop review gates read depends on the stdin payload and
is **unmeasured**. The plan therefore picks a design whose store reachability does not
depend on that value, and adds a run-phase measurement of the review gates'
input-to-root mapping (acceptance AC-WSR-006).

## §B Known issues and risks

- **R1 — unmeasured hook input root.** The stdin `cwd` / `project_dir` /
  `CLAUDE_PROJECT_DIR` triple a hook receives in a worktree session was not separated
  (remeasure "Gaps"). Mitigation: REQ-WSR-001/006/007 make store reachability
  independent of it; AC-WSR-006 measures, for all 27 input combinations, which root
  each Stop review gate reads today and requires after the change. Residuals (spec.md
  §5): a guard input naming `P` for an audit on `W` refuses a `W` receipt as "another
  tree" (loud); a spawn input naming a different tree than the refused stop does not
  see the refusal (silent, and equally true today with per-tree stores). The optional
  M0 probe (§D) captures the real payload without gating any criterion.
- **R2 — correction to the evidence base.** The remeasure file attributes the two review
  gates to the env-first `HookInput.ProjectDir` fill. By code they use their own parser
  (`internal/cli/codex_review_gate.go:207-219`) and resolve stdin `project_dir` → stdin
  `cwd` → env (`:236-246`). spec.md §1.1 records the code reading; no requirement relies
  on the remeasure's attribution for these two readers.
- **R3 — shared verify store.** Verify keys are content hashes of a tree's state
  (`internal/verify/key.go:39-70`). With one store for `P` and its worktrees, two trees
  in identical state share a snapshot. Accepted: the key already asserts identical
  content; recorded as an edge case (acceptance.md §D.1).
- **R4 — predecessor tests that break (corrected at 0.2.0).** Two existing tests assert
  behaviour this SPEC changes and are edited in the same change (acceptance AC-WSR-004
  names them): `TestConfigOrphanedWorktree_PrimaryGateEnforced`
  (`internal/cli/mcp_project_root_worktree_gate_test.go:50`) fails its premise at L61-64
  ("the first codex_audit call should have created W/.moai/state") once REQ-WSR-002
  keeps `W` free of `.moai`; its premise becomes "`W/.moai` does not exist after the first
  call" and its second round creates `W/.moai/state` directly. `TestLinkedWorktree_DescriptionsStateTheGateSource`
  (`internal/cli/mcp_project_root_worktree_ac_test.go:296`) asserts the two phrases
  REQ-WSR-016 removes (L300, L315). The AC-MWU-016 test
  `TestConfigOrphanedWorktree_CatalogueWarning` checks only the presence of the
  `worktree_warning` key and the unchanged `_root.warning` text (L238, L256-L266), so it
  is **not** edited; the new warning text is asserted by AC-WSR-013's own test. The
  0.1.0 statement of R4 had both halves backwards.
- **R5 — hook package cannot import `internal/cli`.** The primary identification and the
  config-orphaned predicate live in `internal/cli/mcp_worktree_root.go`. REQ-WSR-001 needs
  every call site to reach the same answer, so the resolution needs a home both sides can
  import — `internal/auditreceipt` or a new small internal package. Decided at M1, not a
  user-facing decision.
- **R6 — hook latency.** A config-orphaned guard tree now runs a scrubbed primary
  identification (two git subprocesses) on SubagentStart, on SubagentStop, and on the
  PreToolUse phase-entry spawn check (`checkAuditReceiptSpawn`,
  `internal/hook/audit_receipt_guard.go:198`), and a config-orphaned review-gate root runs
  it on every Stop. Only config-orphaned roots pay it (REQ-WSR-005); existing hook
  timeouts apply.
- **R7 — review gates switch on in worktree sessions (user-visible).** Today both Stop
  review gates are disabled in a config-orphaned `W` because they read their opt-in flag
  from `W`, which has no `workflow.yaml`. REQ-WSR-008 makes them read the primary's flag,
  so in a repository whose primary enables `workflow.codex.review_gate.enabled`, every
  Stop in a worktree session with uncommitted changes starts running a codex review
  (bounded by the gate's 900 s timeout, fail-open), and the multi gate can start blocking.
  This is the intended behaviour, but it is a change users will notice; the sync phase
  names it in the release notes.
- **R8 — shared-store mixing of records (plan-audit D1, D6).** Rejections carry no tree
  identity and are listed and cleared store-wide (`internal/auditreceipt/store.go:270,
  300, 410-415`); convergence results are keyed by session id only
  (`internal/cli/mcp_convergence.go:903-917`) — the reason commit `720149668` (card t182)
  moved them under the audited tree. Routing worktree state into the primary's store
  would reopen both. Mitigation: REQ-WSR-003 gives every routed record a tree identity,
  forbids cross-tree replacement, and filters the guard's list/clear by tree; REQ-WSR-007
  makes the gate block on any `fail` of the session. Decision 5 below records the gate
  rule and its alternatives.
- **R9 — fail-closed spawn denial when the primary is unidentifiable (decision 2).** The
  recommended default denies phase-entry spawns from a config-orphaned worktree whose
  primary cannot be identified. (i) It does so even in repositories that never enabled
  the receipt gate, so manager-develop / manager-docs / manager-git spawns from that tree
  are all refused. (ii) Under a separate-git-dir layout the tree satisfies the git-free
  config-orphaned evidence (`internal/cli/mcp_worktree_root.go:220-259`) while primary
  identification fails structurally (predecessor test
  `TestConfigOrphanedWorktree_FailsClosedWhenPrimaryUnidentified`, W3 case), so the denial
  is permanent there. (iii) This differs from the MCP side (REQ-MWU-012), which changes only
  the verdict an audit call returns; the hook refuses spawns with no audit in play.
  Release path: put `workflow.yaml` in `W` (`W/.moai/config/sections/workflow.yaml`) —
  `W` is then not config-orphaned and the guard reads `W`'s own config. No REQ change;
  the denial text (REQ-WSR-010) names the assumed cause.

## §C Decisions for Implementation Kickoff

The requirements are written under each recommended default. The orchestrator presents
these at Implementation Kickoff; choosing an alternative changes the listed REQs/ACs.
Ordered by how likely each is to change.

| # | Decision | Recommended default | Alternatives | Fact each alternative changes |
|---|---|---|---|---|
| 1 | Where state of a config-orphaned worktree is stored (receipts, start markers, rejections, `audit_multi` convergence, verify snapshots) | **Primary checkout's `.moai/state`, reached through one store-root answer that writers and readers both derive from whatever root they hold (`W → P`, `P → P`); every routed record carries tree identity `W`.** Store reachability then holds whichever root a hook reader starts from, so it does not depend on the unmeasured hook input (D18). Tree-identity comparisons (receipt vs start marker, rejection filter) still depend on the input root; a `P` input for a `W` audit ends in a loud "another tree" refusal (spec.md §5). Keeps `W` free of a `.moai` the repository does not track. | (b) **Worktree's `.moai/state`** (t1202's old decision-2 default). (c) **Worktree store, readers search both** `P` and every registered worktree's store. | (b) A reader that resolves `P` — the measured behaviour of the env-first hook reader in this session — would not find `W`'s state: the `multi-review-gate` has only a session id to go on, so rows of AC-WSR-006 whose `R = P` would miss a `W` result; `W` keeps getting a `.moai` created by writes; REQ-WSR-003's shared-store clauses become unnecessary. (c) Readers gain a `git worktree list` scan on every Stop hook and a rule for which store wins when two hold the same session id; REQ-WSR-001/006 and AC-WSR-001/002/005/014 change. |
| 2 | Receipt guard when the primary of a config-orphaned tree cannot be identified | **Fail-closed: treat the codex gate as `required`, write nothing (no store exists), refuse a PASS at first stop, notice without blocking at a re-entrant stop, and deny phase-entry spawns from that worktree — each naming the assumed cause** (the same `required` assumption as the MCP side, SPEC-MCP-WORKTREE-UNTRACKED-001 REQ-MWU-012). Because no store exists, no receipt can be recorded there either (REQ-WSR-004), so phase-entry spawns from such a worktree stay denied until the primary becomes identifiable. Blast radius (risk R9): (i) the denial applies even in a repository that never enabled the receipt gate — every manager-develop / manager-docs / manager-git spawn from that worktree is refused; (ii) under a separate-git-dir layout primary identification fails structurally, so the denial is permanent, not transient; (iii) it differs from the MCP side, which only changes the verdict of an audit call, whereas the hook blocks spawns with no audit involved. Release path: place a `workflow.yaml` in `W` (`W/.moai/config/sections/workflow.yaml`), which makes `W` no longer config-orphaned. | (b) **Fail-open**: guard stays inactive. (c) **Keep today**: guard reads only the tree's own config (inactive on every config-orphaned tree). (d) **Fail-closed at stop only** (the 0.1.0 wording). | (b) The MCP side reports the gate `required` and exposes a receipt id, but the hook never checks it, so an uncorroborated PASS is accepted on exactly the roots where the MCP side assumed `required`; REQ-WSR-010 and AC-WSR-009 change. (c) Same as (b) plus the identifiable-primary case: t1202 audit D18(b) (guard is a no-op on an untracked-`.moai` worktree) stays open; REQ-WSR-009/010 and AC-WSR-007/008/009/016 drop. (d) The first stop blocks, but a re-entrant stop has no store to leave a rejection in, so the next phase-entry spawn is allowed — the fail-closed stance ends after one re-entry (plan-audit D5); AC-WSR-009's spawn cell drops. |
| 3 | Catalogue source for `spec_progress` / `spec_drift` / `spec_audit` on a config-orphaned root | **Union of the worktree's and the primary's `.moai/specs`, each record and finding tagged with its source; the worktree copy wins on an ID present in both, and the shadowed primary copy is named.** No SPEC can vanish (D25), and the primary catalogue becomes visible. | (b) **Primary only, plus a signal** when `W/.moai/specs` is non-empty. (c) **Worktree only** (today's interim, with the t1202 warning). | (b) A SPEC written inside `W` — the D25 case — is absent from the answer; only the signal says so, and a caller that reads records but not `_root` loses it silently. The durable config-orphaned predicate cannot tell "hidden" from "empty" without the extra signal (remeasure §D25). REQ-WSR-011/012/013 and AC-WSR-010/011 change. (c) The primary catalogue stays invisible from a worktree session — every existing SPEC reads as absent — and the warning is the only signal; REQ-WSR-011–015 and AC-WSR-010–013/016 change. |
| 4 | How the `multi-review-gate` judges a session whose results for several trees share one store | **Block when any result of the session in the store root is `fail`** (REQ-WSR-007); results never replace another tree's (REQ-WSR-003). Independent of which root the gate's input names. | (b) **Read only the result whose tree identity equals the gate's resolved root.** (c) **Assume one tree per session** and keep the session-only key (accept t182's overwrite risk). | (b) Root-dependent: with `R = P` for a `W` audit the gate finds no matching result and allows — the D18 premise returns for the rows of AC-WSR-006 whose `R` differs from the audited tree; REQ-WSR-007's second clause and AC-WSR-002 cell 2 change. (c) A later PASS for one tree overwrites an earlier FAIL for another and the gate allows silently (the t182 class, commit `720149668`); REQ-WSR-003's convergence clause and AC-WSR-002 cell 2 drop. |
| 5 | State already written under a worktree's `.moai/state` by the interim behaviour | **Leave it in place, never read it** (spec.md §6). Interim receipts are per-audit and verify snapshots are keyed to a tree state that has since moved, so neither is expected to be needed. AC-WSR-006's decoy in `W/.moai/state` pins that it is not read. | (b) **Read-through**: readers fall back to `W/.moai/state` when the primary store has nothing. (c) **Migrate** once into the primary store. | (b) Reintroduces a second store and the reader-disagreement class this SPEC removes; a stale `W` receipt could corroborate a new audit. (c) Adds a one-shot migration path with its own failure modes and a write into `P` outside any audit call. Either adds a REQ and an AC. |

## §D Pre-flight (run phase)

- Re-read branch and HEAD in the worktree; confirm base `c630de892` or its successor on
  develop; merge local develop first if it moved.
- Optional **M0 probe** (evidence for R1; gates no criterion; skip if the operator
  declines). In a live worktree session of this repository, capture once, with temporary
  and uncommitted debug output removed afterwards: (a) the SubagentStart hook's stdin
  `cwd` and the hook process's `CLAUDE_PROJECT_DIR` (receipt guard path); (b) the Stop
  hook's raw stdin payload — its `cwd` and any `project_dir` field — and the Stop hook
  process's `CLAUDE_PROJECT_DIR` and working directory (review-gate path). Record the
  observed values, the command, and the session id in progress.md §E.2, and name which
  row of AC-WSR-006's matrix the live payload falls on.

## §E Milestones (ordered by decision reversibility)

Priority High first; each milestone ends with the targeted package tests, never a local
full-suite run.

- **M1 — Store-root answer and RED tests (High).** Decide the home of the resolution (R5).
  Commit the RED tests for AC-WSR-001, -002, -006, -007, -008, -010, -014, -016 (RED-first
  witnesses per acceptance.md §D.2) and record each RED run's failing predicate, plus
  AC-WSR-006's observed per-row matrix.
- **M2 — Tree identity in the shared store (High; Decision 4).** Rejections and
  convergence results carry tree identity and coexist; the guard's list/clear filter by
  tree; the gate blocks on any `fail` of the session (AC-WSR-002, -007). Execute the
  named discriminating mutants.
- **M3 — Catalogue union (High; Decision 3).** `spec_progress` / `spec_drift` /
  `spec_audit` answer over the union with source tags, shadowing, and the
  unresolved-primary statement (AC-WSR-010, -011, -012).
- **M4 — State writers (High; Decision 1).** Route receipt, start marker, rejection,
  convergence, and verify writes through the store root; unresolved-primary behaviour
  (AC-WSR-001, -003, -004, -014).
- **M5 — Paired readers (High).** Verify tools and CLI, `multi-review-gate` state read and
  both review gates' opt-in flags (AC-WSR-005, -006).
- **M6 — Receipt guard (High; Decision 2).** Gate read from the primary, fail-closed without
  a store including the re-entrant stop and the spawn check (AC-WSR-008, -009).
- **M7 — `_root` provenance (Medium).** Sources list and new `worktree_warning` text
  (AC-WSR-013).
- **M8 — Documentation and templates (Medium).** Catalogue rule section, shared
  `project_root` description, tool descriptions, template mirror, `make build`; edit the
  two predecessor tests named in R4 (AC-WSR-004, -015).
- **M9 — End to end and self-verification (Medium).** AC-WSR-016; targeted `go test`,
  `go vet`, `golangci-lint run` on `internal/cli`, `internal/hook`, `internal/auditreceipt`,
  `internal/verify`; evidence into progress.md §E.2.

## §F Technical approach (non-binding)

- One exported function returns `(storeRoot, err)` for any root, built from the existing
  config-orphaned predicate and scrubbed primary identification, moved or re-exported so
  `internal/hook` can call it without importing `internal/cli`.
- Writers call it where they compute `<root>/.moai/state`; records keep
  `tree_root = canonical root` (identity), and only the directory passed to the store
  changes. Rejection and convergence file keys gain a tree component only for
  config-orphaned trees, so a non-orphaned root's paths stay byte-identical (REQ-WSR-005).
- The union catalogue reads two `ListDocs`-style inventories and merges by SPEC ID; audit and
  drift run per source and tag findings.

## §G Anti-patterns to avoid

- Branching on "does `W/.moai` exist" instead of the config-orphaned predicate (the D3/D17
  regression class of t1202).
- A reader that falls back to another root when its store root is unresolved.
- Asserting in a test or a doc which hook input root Claude Code sends: it is unmeasured.
  AC-WSR-006 measures the gates' mapping from each input to a root, not the input itself.
- A record routed into a shared store without a tree identity, or a list/clear that spans
  trees.

## §H Cross-references

- spec.md §1.1 (evidence table), §1.4 (shared-store consequences), §5 (measurement gap)
- `.moai/reports/t1213/remeasure-d18-d25.md`, `.moai/reports/t1213/plan-audit-iter1.md`
- `.moai/specs/SPEC-MCP-WORKTREE-UNTRACKED-001/spec.md` §3.1, §4.4, §6
- primary checkout `.moai/reports/t1202/plan-audit-iter2.md` D18, D25
- commit `720149668` (card t182, convergence result under the audited tree)
- `.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input;
  `moai-mcp-tools-catalogue.md` § Linked worktrees
