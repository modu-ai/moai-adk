# SPEC-WORKTREE-STATE-ROOT-001 — Plan

Card t1213 (class C, Tier M). Base: develop `c630de892`. Worktree `.claude/worktrees/t1213`,
branch `WT-worktree-state-roots`.

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
the receipt guard and the two Stop review gates read depends on the stdin `cwd` value
and is **unmeasured**. The plan therefore picks a design whose correctness does not
depend on that value.

## §B Known issues and risks

- **R1 — unmeasured hook input root.** The stdin `cwd` / `CLAUDE_PROJECT_DIR` pair a
  hook receives in a worktree session was not separated (remeasure "Gaps"). Mitigation:
  REQ-WSR-001/006/007 make state agreement independent of it, and AC-WSR-006 runs both
  input roots. Residual: a guard input naming `P` for an audit on `W` records identity
  `P` and refuses a `W` receipt as "another tree" (loud, not silent). The optional M0
  probe below measures it without making any AC depend on the result.
- **R2 — correction to the evidence base.** The remeasure file attributes the two review
  gates to the env-first `HookInput.ProjectDir` fill. By code they use their own parser
  (`internal/cli/codex_review_gate.go:207-213`) and resolve stdin `project_dir` → stdin
  `cwd` → env (`:236-246`). spec.md §1.1 records the code reading; no requirement relies
  on the remeasure's attribution for these two readers.
- **R3 — shared verify store.** Verify keys are content hashes of a tree's state
  (`internal/verify/key.go:39-70`). With one store for `P` and its worktrees, two trees
  in identical state share a snapshot. Accepted: the key already asserts identical
  content; recorded as an edge case (acceptance.md §D.1).
- **R4 — predecessor tests touched.** SPEC-MCP-WORKTREE-UNTRACKED-001 AC-MWU-016 asserts
  the old `worktree_warning` text; AC-MWU-014 repeats calls "after the first `codex_audit`
  call has created `W/.moai/state/`", which no longer happens. The first is updated
  (AC-WSR-013); the second still passes (the repeat becomes a plain repeat) and is left
  as is.
- **R5 — hook package cannot import `internal/cli`.** The primary identification and the
  config-orphaned predicate live in `internal/cli/mcp_worktree_root.go`. The single
  mapping (REQ-WSR-001) needs a home both sides can import — `internal/auditreceipt` or
  a new small internal package. Decided at M2, not a user-facing decision.
- **R6 — hook latency.** A config-orphaned guard tree now runs a scrubbed primary
  identification (two git subprocesses) on SubagentStart/Stop. Only config-orphaned roots
  pay it (REQ-WSR-005); existing hook timeouts apply.

## §C Decisions for Implementation Kickoff

The requirements are written under each recommended default. The orchestrator presents
these at Implementation Kickoff; choosing an alternative changes the listed REQs/ACs.

| # | Decision | Recommended default | Alternatives | Fact each alternative changes |
|---|---|---|---|---|
| 1 | Where state of a config-orphaned worktree is stored (receipts, `audit_multi` convergence, verify snapshots) | **Primary checkout's `.moai/state`, reached through one mapping that writers and readers both apply to whatever root they hold (`W → P`, `P → P`); receipts keep tree identity `W`.** Correct whichever root a hook reader starts from, so it does not depend on the unmeasured hook input (D18); keeps `W` free of a `.moai` the repository does not track; receipts already carry and check tree identity (`internal/auditreceipt/store.go` `CheckCitedReceipts`, "another tree" cause). | (b) **Worktree's `.moai/state`** (t1202's old decision-2 default). (c) **Worktree store, readers search both** `P` and every registered worktree's store. | (b) A reader that resolves `P` — the measured behaviour of the env-first hook reader in this session — would not find `W`'s state: the `multi-review-gate` has only a session id to go on, so with `cwd = P` it misses the convergence result (AC-WSR-006 second cell stays red); `W` keeps getting a `.moai` created by writes. (c) Readers gain a `git worktree list` scan on every Stop hook and a rule for which store wins when two hold the same session id; REQ-WSR-001/006 and AC-WSR-001/002/005/014 change. |
| 2 | Receipt guard when the primary of a config-orphaned tree cannot be identified | **Fail-closed: treat the codex gate as `required` and name the assumed cause** (matches the MCP side, SPEC-MCP-WORKTREE-UNTRACKED-001 REQ-MWU-012). | (b) **Fail-open**: guard stays inactive. (c) **Keep today**: guard reads only the tree's own config (inactive on every config-orphaned tree). | (b) The MCP side reports the gate `required` and exposes a receipt id, but the hook never checks it, so an uncorroborated PASS is accepted on exactly the roots where the MCP side assumed `required`; REQ-WSR-010 and AC-WSR-009 change. (c) Same as (b) plus the identifiable-primary case: t1202 audit D18(b) (guard is a no-op on an untracked-`.moai` worktree) stays open; REQ-WSR-009/010 and AC-WSR-008/009/016 drop. |
| 3 | Catalogue source for `spec_progress` / `spec_drift` / `spec_audit` on a config-orphaned root | **Union of the worktree's and the primary's `.moai/specs`, each record tagged with its source; the worktree copy wins on an ID present in both, and the shadowed primary copy is named.** No SPEC can vanish (D25), and the primary catalogue becomes visible. | (b) **Primary only, plus a signal** when `W/.moai/specs` is non-empty. (c) **Worktree only** (today's interim, with the t1202 warning). | (b) A SPEC written inside `W` — the D25 case — is absent from the answer; only the signal says so, and a caller that reads records but not `_root` loses it silently. The durable config-orphaned predicate cannot tell "hidden" from "empty" without the extra signal (remeasure §D25). REQ-WSR-011/012/013 and AC-WSR-010/011 change. (c) The primary catalogue stays invisible from a worktree session — every existing SPEC reads as absent — and the warning is the only signal; REQ-WSR-011–015 and AC-WSR-010–013/016 change. |
| 4 | State already written under a worktree's `.moai/state` by the interim behaviour | **Leave it in place, never read it** (spec.md §6). Interim receipts are per-audit and verify snapshots are keyed to a tree state that has since moved, so neither is expected to be needed. | (b) **Read-through**: readers fall back to `W/.moai/state` when the primary store has nothing. (c) **Migrate** once into the primary store. | (b) Reintroduces a second store and the reader-disagreement class this SPEC removes; a stale `W` receipt could corroborate a new audit. (c) Adds a one-shot migration path with its own failure modes and a write into `P` outside any audit call. Either adds a REQ and an AC. |

## §D Pre-flight (run phase)

- Re-read branch and HEAD in the worktree; confirm base `c630de892` or its successor on
  develop; merge local develop first if it moved.
- Optional **M0 probe** (does not gate any AC): a temporary, uncommitted debug line in the
  guard's SubagentStart path logging stdin `cwd` and `CLAUDE_PROJECT_DIR` to the tree's log,
  exercised once from a live worktree session, then removed. Result recorded in
  progress.md §E.2 as evidence for R1. Skip if the operator declines.

## §E Milestones (ordered by decision reversibility)

Priority High first; each milestone ends with the targeted package tests, never a local
full-suite run.

- **M1 — Mapping contract and RED tests (High).** Decide the home of the single mapping
  (R5). Commit the mapping signature with a stub body returning its input, plus the RED
  tests for AC-WSR-001, -006, -008, -010, -014, -016 (RED-first witnesses per acceptance.md
  §D.2). Record each RED run's failing predicate.
- **M2 — Catalogue union (High; Decision 3).** `spec_progress` / `spec_drift` / `spec_audit`
  answer over the union with source tags, shadowing, and the unresolved-primary statement
  (AC-WSR-010, -011, -012).
- **M3 — State writers (High; Decision 1).** Route receipt, convergence, and verify writes
  through the mapping; unresolved-primary behaviour (AC-WSR-001, -002, -003, -004).
- **M4 — Paired readers (High).** Verify tools and CLI, `multi-review-gate` state read and
  both review gates' opt-in flags (AC-WSR-005, -006, -007).
- **M5 — Receipt guard (High; Decision 2).** Gate read from the primary, fail-closed with a
  named cause, store root for markers/receipts/rejections (AC-WSR-008, -009).
- **M6 — `_root` provenance (Medium).** Sources list, new `worktree_warning` text, update
  the predecessor AC-MWU-016 test (AC-WSR-013).
- **M7 — Documentation and templates (Medium).** Catalogue rule section, tool descriptions,
  template mirror, `make build` (AC-WSR-015).
- **M8 — End to end and self-verification (Medium).** AC-WSR-016; targeted `go test`,
  `go vet`, `golangci-lint run` on `internal/cli`, `internal/hook`, `internal/auditreceipt`,
  `internal/verify`; evidence into progress.md §E.2.

## §F Technical approach (non-binding)

- One exported function returns `(storeRoot, err)` for any root, built from the existing
  config-orphaned predicate and scrubbed primary identification, moved or re-exported so
  `internal/hook` can call it without importing `internal/cli`.
- Writers call it where they compute `<root>/.moai/state`; the receipt keeps
  `TreeRoot = canonical root` (identity), and only the directory passed to the store changes.
- The union catalogue reads two `ListDocs`-style inventories and merges by SPEC ID; audit and
  drift run per source and tag findings.

## §G Anti-patterns to avoid

- Branching on "does `W/.moai` exist" instead of the config-orphaned predicate (the D3/D17
  regression class of t1202).
- A reader that falls back to another root when its store root is unresolved.
- Asserting the hook input root in a test or a doc: it is unmeasured.

## §H Cross-references

- spec.md §1.1 (evidence table), §5 (measurement gap)
- `.moai/reports/t1213/remeasure-d18-d25.md`
- `.moai/specs/SPEC-MCP-WORKTREE-UNTRACKED-001/spec.md` §3.1, §4.4, §6
- primary checkout `.moai/reports/t1202/plan-audit-iter2.md` D18, D25
- `.claude/rules/moai/core/moai-mcp-tools.md` § The `project_root` input;
  `moai-mcp-tools-catalogue.md` § Linked worktrees
