# Plan: SPEC-TODO-LAND-AUTO-DONE-001

## §A Context

Card t684 (Tier M, Class C). The queue is the delegation channel; a card that has landed
but not been closed distorts dispatch (45-card backlog measured 2026-09-13). Every
building part this SPEC composes already exists in the tree at HEAD `74d872aaf`; the run
phase is mostly composition plus three guards, not new substrate.

### Observed building parts (all verified in-source this session)

| Part | Where | What it gives this SPEC |
|------|-------|------------------------|
| `newTodoDoneCmd` + `--require-landed` + `todoRequireLanded` | `internal/cli/todo.go` (~644, ~794) | the close path (`ArchiveCard` inside `Mutate`, byte-identity on refusal), the canonical `done <id> landing=<verdict>` stdout contract |
| `GitLandedQuerier.Landed` / `subjectAttribution` / `LandedSubjectArgs` | `internal/kanban/prlink_landed.go` | three-valued landed answer, subject-stream-only predicate (%s, never %B), six positional attribution shapes + merge-target non-attribution rule |
| `LandingEvidence` / `Validate` / `DecodeLandingEvidence` | `internal/kanban/landing_evidence.go` | the stored evidence record; `Landing.SHA` + `sha_source=operator` closed set; `ref_head` is an observation, never a delivery claim |
| `moai todo landed` (`runTodoLanded`, `validateSuppliedSHA`) | `internal/cli/todo_landed.go` | the only path a delivering SHA enters the store; existence + reachability validation |
| `moai todo undone` (`newTodoUndoneCmd`, `RestoreCard`) | `internal/cli/todo_undone.go` | the exact inverse of `done` (card + findings, refused on reissued id) — the reversibility substrate |
| `ReadPrimarySpecStatus` / `readCardSpecStatus` | `internal/cli/todo_landed.go` (re-exported via kanban) | SPEC frontmatter status read at record/scan time, three-valued (empty / unknown / value) |
| `BacklogItem.State` ∈ {queued, picked, dropped} | `internal/kanban/backlog_store.go:57-89` | the live-item filter; `Landing *LandingEvidence` column |
| Integration window (`AcquireIntegrationLock` / `ReleaseIntegrationLock`) | `internal/kanban/integration_lock.go`, `internal/cli/integration.go` | the merge-window mechanism — see §C for why release is NOT the trigger |
| `RuntimeStateDirForRoot` | `internal/kanban/state_dir.go` | home for the append-only execution log |
| `temp_origin.go` | `internal/kanban/temp_origin.go` | fixture git repos for guard tests |
| `recordFactoryCardState(id, specID, "completed", "card.completed")` | `internal/cli/todo.go` (done path) | factory-state propagation the scan must preserve on close |

## §B Known issues this SPEC answers

1. **Stale-queue defect (primary)**: landed cards stay `queued`/`picked` until a human
   runs `done`; 45 instances on 2026-09-13.
2. **Reissued-id collisions (t654/t656/t657)**: the queue reissues `t<n>` ids it cannot
   see (archived cards, other worktrees) — a name-match against the landed ref cannot
   distinguish a predecessor's landed commits from the current card's pending work.
3. **Run-vs-sync gap**: `--require-landed` states its own limit in
   `todoDoneLong`: it "cannot tell a run commit from a sync commit". A close during that
   gap is the misfire guard M2.
4. **Non-landing declarations (t581 precedent)**: commits exist whose message says the
   work did NOT land; today the subject predicate would still attribute such a subject's
   token when it matches a positional shape.

## §C Trigger design decision (the card's core question)

The trigger must be the REMOTE-LANDING-CONFIRMATION moment, never the local merge moment:
between merge and push, the worktree branch is the only copy of the work
(CLAUDE.local.md §4.1). The lead is the sole push actor; it confirms landing with
`git fetch origin develop && git rev-parse origin/develop` after the batch push.

| Option | Mechanism | Pros | Cons | Verdict |
|--------|-----------|------|------|---------|
| (a) Hook after the lead's develop push | PostToolUse Bash-matching hook fires the scan | fires near the push instant | tool-call pattern matching is measured-fragile in this repo (BranchGuard over/under-matching history); a missed match is a SILENT stall — the exact defect class being fixed; untestable end-to-end without a live Claude session; cannot re-run itself; fires even when the push is not a landing batch | rejected |
| (b) `moai integration release` | scan runs when the merge window closes | zero new plumbing; sits inside an existing verified verb | WRONG MOMENT: release precedes push by design (2026-09-02 discipline — "push is outside the window, the lead's batch act"), sometimes by hours (batch closing). It is a LOCAL-merge moment; closing there violates the HARD constraint. It also has `--force` displacement and bypass paths a close-on-release would inherit | rejected |
| (c) Landing-scan command the lead runs at confirmation | `moai todo auto-done`, invoked right after `git rev-parse origin/develop` confirms; also cron/loop-runnable | the confirmation instant IS the trigger; idempotent and re-runnable; deterministic and unit-testable against fixture repos; works regardless of HOW the push happened (script, manual, lead verb); fail-closed (skips, never guesses); no hook fragility; `--dry-run` gives the lead a preview before closing | latency until the scan runs (acceptable: the lead's own procedure runs it immediately after confirmation) | **CHOSEN** |

**Why (c) wins against the stated constraint**: push is the LEAD's single responsibility
and the lead already performs a post-push confirmation read. Making the scan the very
next step of that same procedure (REQ-AD-014) binds automation to the exact
remote-landing-confirmation moment with no new event plumbing, and keeps every close
evidence-gated and reversible. (a) approximates the moment with a fragile matcher;
(b) names the wrong moment outright.

## §D Constraints

- **C1**: the scan is invoked, never ambient (REQ-AD-002). No other command's code path
  may call it as a side effect.
- **C2**: the `Landing` column's provenance closed set is frozen (REQ-AD-009). The
  scan's machine evidence goes to its own JSONL log only.
- **C3**: closes inherit `Mutate`'s byte-identity-on-refusal contract — every guard runs
  INSIDE the mutation callback or before it, never after.
- **C4**: existing stdout readers key on the `done <id> ` prefix; the scan's close lines
  preserve it (`done <id> landing=landed source=auto-land ...`).
- **C5**: the `undone` verb is the ONLY reversal; no second undo verb is added.
- **C6**: template-first — doc changes land in
  `internal/template/templates/.claude/skills/moai/workflows/todo.md` and mirror locally.

## §E Self-verification plan

- `go test ./internal/kanban/...` (guard predicates, negation rule, collision gate
  against `temp_origin.go` fixture repos)
- `go test ./internal/cli/...` (scan verb: evidence forms, gates, log rows, dry-run
  byte-identity, undone reversal row, `TestNew_NoAskUserQuestion` boundary)
- `golangci-lint run ./internal/...`
- LSP baseline captured before M1; zero errors at close.

## §F Milestones (priority-ordered, no time estimates)

**M1 — Guard predicates in `internal/kanban` (Priority High — highest-change-likelihood
decisions live here: the skip-reason vocabulary, the negation rule, the collision
gate)**
1. Negation exclusion in the subject predicate (`not merged` / `not landed`,
   case-insensitive, same-subject-only) with fixture tests — the (t581) guard.
2. Collision gate helper: given id, live item, and archive, return whether the id has
   carried >1 distinct text — the reissue guard.
3. Scan-decision pure function: live items + landed ref answer + recorded SHAs + SPEC
   statuses → per-card decision {close(form) | skip(reason)} — so the whole policy is
   unit-testable before any CLI surface exists. The skip `reason` output vocabulary is
   the CLOSED four-token set canonically enumerated in spec.md REQ-AD-010:
   `ambiguous-id`, `spec-not-completed`, `not-landed`, `query-inconclusive`.

**M2 — CLI verb `moai todo auto-done` (Priority High)**
4. Verb + flags (`--fetch`, `--dry-run`, `--json`), wiring the M1 decision function;
   close via `ArchiveCard` inside `store.Mutate`; `recordFactoryCardState` preserved;
   canonical close-line contract (C4).
5. Append-only JSONL execution log under `RuntimeStateDirForRoot`; reversal rows
   appended by `todo undone` when it restores an auto-closed card (REQ-AD-012).
6. `--dry-run` byte-identity test (whole-queue record identical before/after).

**M3 — Procedure docs + template mirror (Priority Medium — mechanical)**
7. Lead post-push procedure step in
   `internal/template/templates/.claude/skills/moai/workflows/todo.md` + local mirror
   (C6); one line in `.claude/rules/moai/workflow/kanban-dispatch.md`'s lead push
   section if the rule text names the push procedure.

## §G Anti-patterns to refuse

- Closing on the landed grep alone (name matching) — the SPEC-TODO-LANDING-EVIDENCE-001
  ruling (a mention is not a delivery) is inherited, and M1's collision gate makes it
  structural rather than conventional.
- Closing a `picked` card whose SPEC is `implemented` but not `completed`.
- Treating `Landing.RefHead` (an observed ref position) as a delivering SHA — the
  confusion REQ-TLE-013 exists to prevent.
- Writing machine-derived evidence into the `Landing` column to "help future audits".

## §H Cross-references

- SPEC-TODO-LANDING-EVIDENCE-001 — the evidence record and its closed provenance.
- SPEC-TODO-LANDING-ATTRIBUTION-001 — the six positional attribution shapes.
- SPEC-TODO-LANDING-STATE-001 — the three-valued answer and the landed-ref chain.
- SPEC-TODO-DESTRUCTIVE-GUARD-001 — archive/undone symmetry this SPEC's reversibility
  reuses.
- CLAUDE.local.md §4.1 — the git-flow chain, the lead-only push discipline, and the
  post-push confirmation read the scan attaches to.
