---
id: SPEC-KANBAN-WORKTREE-001
title: "Progress — per-card worktree lifecycle with holder liveness and mutual exclusion"
version: "0.3.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, worktree, progress, evidence"
tier: L
---

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` (Tier L set). Revised at v0.2.0 to repair seven plan-audit defects and promoted from Tier M to Tier L in the same revision; revised again at v0.3.0 to repair a second audit's findings, which scored v0.2.0 at 0.79 against the Tier L threshold of 0.85 while verifying every v0.2.0 repair as genuinely closed with no regressions.

- Requirements: 23 (`REQ-KW-001` … `REQ-KW-023`) — **23 of the Tier L ceiling of 25.**
- Acceptance criteria: 23 (`AC-KW-001` … `AC-KW-023`) — **23 of the Tier L ceiling of 25.**
- The v0.3.0 additions are `REQ-KW-019` / `AC-KW-019` (the branch-match ambiguity refusal), `REQ-KW-020` and `REQ-KW-021` with their criteria (the two operator escapes), and `REQ-KW-022` / `REQ-KW-023` (the creation actor and its serialization). Four further requirements — `REQ-KW-003`, `REQ-KW-014`, `REQ-KW-017`, `REQ-KW-018` — are amended in place, and nine criteria gain rows. Folding was available for each addition and refused; the temptations and the reason each was declined are recorded in `spec.md` §B.
- The v0.2.0 additions were `REQ-KW-017` / `AC-KW-017` (the merge-observation decision) and `REQ-KW-018` / `AC-KW-018` (the reuse-without-a-cycle constraint), which took the SPEC two past the Tier M ceiling of 16 and forced the promotion. Tier L also raises the plan-auditor PASS threshold from 0.80 to 0.85.
- Milestones: 5 (M0 … M4), ordered by decision-reversibility — the lock substrate and the holder critical section first, the mirror and catalog work last. At v0.3.0 the branch match rule and the creation actor join M2, the lock-clearing race condition lands in M1 with the lock rather than as a follow-up, and the two operator escapes land last within M3.
- Dependencies: `SPEC-KANBAN-RENAME-001` and `SPEC-KANBAN-BOARD-001`, both unlanded on `origin/main`. M0 gates on both.
- Siblings: `SPEC-KANBAN-BOARD-001` (consumed by REQ id), `SPEC-KANBAN-BOOTSTRAP-001`. The scope boundary is stated in `spec.md` §C so an omission is not read as a gap.

Five predecessor decisions changed, each with the measurement that forced it:

| Change | Measurement |
|---|---|
| Clean-tree assignment fence rejected → advisory file lock over the `internal/spec` pattern | `internal/lockfile`'s Windows path is an in-process mutex (cross-process **not** protected, by its own documented decision); `internal/spec/lock_windows.go` uses atomic create and **is** cross-process |
| Stall criterion re-anchored from a constant comparison to a configuration check | `grep -rn '14400' --include='*.go' internal/ pkg/` → **exit 1, zero matches**; only occurrence is prose at `.claude/skills/moai/workflows/factory.md:96` |
| Branch name settled as `feature/<SPEC-ID>` via the existing helper | `internal/cli/worktree/shared.go` `resolveSpecBranch` already implements exactly this mapping |
| Creation idempotency settled: no-op on match, two distinct named refusals on the two mismatches | — (predecessor was silent; the four conditions are enumerated in `spec.md` §A.3) |
| Disposal executor named (lead) and merge state observed | `internal/cli/session_worktree_prmerge.go` `branchMergedForCleanup` reads `gh pr view --json state` with a `git branch --merged` fallback |

Seven plan-audit defects repaired at v0.2.0, each with the measurement that decided it:

| Defect | Measurement | Repair |
|---|---|---|
| §A.6 and §A.7 contradicted each other on tree cleanliness, and the contradiction carried the whole false-positive safety argument | the contradiction is internal to v0.1.0's own text: §A.6 records the post-commit clean instant that §A.7's argument requires never to occur | release now requires **positive evidence of death** (REQ-KW-009 age criterion is necessary-not-sufficient; REQ-KW-011 adds a process probe); the dirty-tree gate is demoted to defence in depth |
| the branch prefix was wrong, so the disposal gate could never fire | `git for-each-ref … \| uniq -c` → **63 `feat/`** against **3 `feature/`** at v0.2.0 authoring time (64 against 3 at promotion time — `research.md` §C); `resolveSpecBranch` (`shared.go:32`) returns `"feature/" + name` | REQ-KW-003 observes the reported branch and recognizes a card's by the SPEC identifier it carries, not by prefix |
| the named merge predicate could not express the two-PR condition | `branchMergedForCleanup` (`session_worktree_prmerge.go:179`) takes one branch name + one bool, returns one bool | REQ-KW-007 keys on discovered pull-request identities, in the `gh pr view <PR>` form of `spec-workflow.md:437` |
| the two merge-detection paths were described as equivalent | the helper's own comment records the `gh`-absent fallback as squash-merge blind; `origin/main`'s last 200 first-parent commits: **0** merge commits, **199/200** carrying `(#N)` | new REQ-KW-017 — no disposal, a once-per-invocation notice, and no merged-branch substitution |
| both reused symbols are unexported, one un-exportable | both measured unexported; `internal/cli` is imported by `cmd/moai/main.go` and will import kanban, so the reverse edge is a cycle. `internal/worktree` measured **dependency-free** | new REQ-KW-018 — extract branch derivation into the leaf; own the merge contract here |
| a lost Windows lock left a card permanently stuck | `internal/spec/lock_windows.go` header: stale detection is "post-MVP", cleanup is manual `del` — and this SPEC's recovery is itself a holder change needing that lock | REQ-KW-014 gains a bounded clearing act gated on the recorded process being observed absent |
| the `lead` was a runtime dependency declared nowhere | `SPEC-KANBAN-BOOTSTRAP-001` already lists this SPEC in its own `dependencies:`, so the reverse declaration would be a cycle | REQ-KW-002 records it as a **runtime** dependency; REQ-KW-007 and REQ-KW-011 refuse when no `lead` resolves |

Two of this document's own v0.1.0 claims are corrected. `grep -rniE 'clean tree|clean-tree|working tree is clean|tree is clean|concurrent assignment'` over `SPEC-KANBAN-MULTISESSION-001` reports **2 hits, both orphan classification** — the predecessor carried **no** assignment fence, so REQ-KW-013's lock is an addition rather than a replacement. And §A.5's "the peer registry is not a liveness signal in either direction" was over-broad: dead-PID entries defeat only its positive reading, and REQ-KW-011 consumes the negative one.

`SPEC-KANBAN-BOARD-001` reached Tier L in the same pass and gained `REQ-KB-017` / `REQ-KB-018` / `REQ-KB-019`. Every holder write described here is consequently a `lead` act, and the card-scoped lock of REQ-KW-013 stands alongside the board-wide lock rather than being superseded by it — which `REQ-KB-019` states explicitly.

Seven further defects repaired at v0.3.0, each with the measurement that decided it:

| Defect | Measurement | Repair |
|---|---|---|
| `REQ-KW-017`'s rejection premise was falsified by a predicate this SPEC already adopts | `IsBranchMerged` (`internal/core/git/worktree.go:233`, interface `types.go:194`) is documented as reporting merge "irrespective of the merge strategy", its **S4** signal being a dedicated squash probe; **zero** `gh` invocations in that package; live at `internal/cli/worktree/clean.go` | re-grounded on arity — a per-branch predicate cannot answer a per-pull-request-identity question. Outcome unchanged, argument replaced; `AC-KW-017` now scans for this predicate by name |
| "the SPEC identifier the branch carries" was multi-valued and superstring-vulnerable | 3 distinct branch names carry `SPEC-CODEX-PHASE2-001`; **20 of 35** SPEC-carrying branch segments are phase-suffixed; **no** identifier is a hyphen-delimited prefix of another among 31 | `REQ-KW-003` gains an exact-token rule bounded by end-of-segment or hyphen; new `REQ-KW-019` refuses on multiple matches, scoped to single-resolution uses only |
| two terminal states had no escape, no actor, and no end-state | none needed — the omission is internal, and the family already applies the escape shape twice (`REQ-KB-013`, `REQ-KW-014`) | new `REQ-KW-020` (force-release, gated on the holder being unprobeable) and `REQ-KW-021` (orphan-clear, which never touches the tree) |
| the lock-clearing act had a check-then-unlink race | `internal/spec/lock_unix.go` releases by closing the descriptor, so `.moai/state/`'s **14** `spec-close-*.lock` artifacts (oldest 2026-05-30) are inert — the defect is Windows-only | `REQ-KW-014` conditions removal on the artifact still being the one inspected; `AC-KW-015` gains a concurrent re-acquisition row, and every row records its platform |
| the extraction target was the package §C excludes | `internal/worktree/doc.go:1-5` declares **working tree state guard** primitives; `ls internal/kanban` → no such directory, so "both consumers already reach it" was false | `REQ-KW-018` targets `internal/core/git` — branch lifecycle by its own `doc.go`, `internal/foundation` its only internal dependency, already imported by `internal/cli/worktree` |
| worktree creation had no actor and no serialization | disposal and release are lead-only with refusals; creation was "the system shall create", and neither sibling claims it | new `REQ-KW-022` (the lead creates, refuses when unresolvable) and `REQ-KW-023` (creation serialized under the card lock) |
| `worker-` is unusable as a worktree name prefix | `cleanupMoaiWorktrees` (`launcher.go:481`) runs on every `moai cc` via `applyCCMode` (`:227`); the `worker-` filter gates **both** bases and removal is non-force, so a SPEC-identifier-named tree survives | `REQ-KW-003` carries the prohibition; `AC-KW-002` judges it as a behaviour with a positive control |

Two sibling consumptions landed with this revision. `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` now carries the role declaration and requires it resolvable from a session that is not the `lead` — the clause exists for this SPEC's three gates, which previously deferred to `REQ-KS-004`, a requirement that elects the role without defining how it is read. And `SPEC-KANBAN-BOARD-001` `REQ-KB-023` hardened its own clearing act against the same race repaired here.

**One finding is surfaced rather than repaired.** This SPEC and `SPEC-KANBAN-BOARD-001` now each name the other in `dependencies:` — a declared cycle, created when the board sibling promoted this SPEC out of `related_specs:`. The resolution belongs there: this SPEC's need is a **landing** dependency (the holder field must exist), the board's is a **contract** dependency (`REQ-KB-020` consumes `REQ-KW-003`'s rule, readable from the document). `AC-KW-001` carries the observation and records that it currently fails; `plan.md` §C.16 re-runs it; `plan.md` AP-23 forbids resolving it by deleting this SPEC's entry.

Branch-guard baseline recorded at plan time (re-read at run-phase, never carried forward): primary checkout `main` @ `b59a8ba7d`; `git rev-parse --git-common-dir` from this worktree → `/Users/goos/MoAI/moai-adk-go/.git`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
