---
id: SPEC-KANBAN-BOOTSTRAP-001
title: "Implementation plan — Kanban session topology, bootstrap, and dispatch"
version: "0.5.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, bootstrap, topology, dispatch, plan, recorded-baseline, sole-writer, role-resolution"
tier: L
dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-WORKTREE-001]
related_specs: [SPEC-KANBAN-MULTISESSION-001, SPEC-FACTORY-MODE-001]
---

## §A. Context

This plan implements `spec.md` REQ-KS-001 … REQ-KS-025. It is the largest of the three kanban SPECs and the last of them to become implementable: it consumes the board sibling's card record and state-store rule, and the worktree sibling's per-card tree, and it is the only one of the three that touches the launcher entry switch.

The milestones below are ordered by **decision reversibility**, not by dependency convenience. The two decisions most expensive to revisit — the dispatch payload's type shape (M2) and the printed guidance's command set (M3) — come before the mechanical work, so that a reader reviewing this plan spends attention where a wrong choice is costly. The baseline recording (M1) precedes both because its window closes: it can only be taken after the rename lands and before the entry switch is touched.

At v0.2.0 M1 carries **four** recordings rather than one. The entry-switch baseline was always there; the human-gate inventory, the coder chain's carried-over behaviors, and each mirror pair's classification and pre-change `diff` join it, because each is the comparand of an equivalence claim that was previously compared against nothing (`spec.md` §D.9). Two of the four have windows that close later than M1's — the chain recording before M5, the pair recording before M6 — but all four are placed here for one reason: a recording is only ever taken too late, never too early, and a milestone whose whole purpose is "write down what is true now" is the one place the discipline is visible.

At v0.3.0 nothing moves between milestones, and two gain content. **M0** gains the launcher measurement (§C checks 8 and 9): the backend decision M3 implements rests on a premise about what `moai cc` and `moai glm` actually do, and that premise had never been measured — so it is measured before the surface that depends on it is written, not after. **M3** gains the role declaration each session carries, which REQ-KS-006 now requires alongside the launch label; its contract is consumed by M2's routing and M4's quorum accounting, its carrier is chosen at M3, and the two are separable, which is why the ordering does not change.

## §B. Known issues and unlanded ground

**B.1 Three prerequisites, none of them optional.** The rename, the board sibling, and the worktree sibling are all in `dependencies:`. The cross-session messaging doctrine is a fourth prerequisite that is deliberately **not** in `dependencies:` because `DependencyExistsRule` resolves that field against SPEC directories and a rule file is not a SPEC; it is gated by REQ-KS-002 at preflight instead.

**B.2 Out-of-band note — a stale citation this SPEC does not own.** `internal/cli/CLAUDE.md` § Conventions cites `TestNew_NoAskUserQuestion` (`worktree/new_test.go`) as the canonical static guard, and repeats it in its file list alongside `TestPropose_NoAskUserQuestion` (`harness/route_test.go`). Measured in this worktree, neither the function definition nor the file `internal/cli/worktree/new_test.go` exists; the name survives only in the comments of six test files that copied the pattern from it. `TestPropose_NoAskUserQuestion` *does* exist, but in `internal/cli/harness/propose_boundary_test.go`, not in the cited `route_test.go`.

This is a documentation defect this SPEC surfaced but did not create, and repairing it is **out of scope** (`spec.md` §C). It is recorded here for one reason: every new guard in this repository is written by copying the citation, so an uncorrected citation manufactures a fresh dead reference on each copy. Whoever next edits that file should re-anchor both names onto definitions verified present.

**B.2a The sibling that moved under this plan.** `SPEC-KANBAN-BOARD-001` reached v0.2.0 after this plan's v0.1.0 was written, restoring `REQ-KB-017` (the `lead` is the sole writer of board state), `REQ-KB-018` (atomic write) and `REQ-KB-019` (board-wide lock). Nothing in this plan's milestones implements them — they are the sibling's M1 — but two things here are downstream of them. **M2** must not define a dispatch that instructs a worker to record a card's progression anywhere but its own `progress.md`, and **M3**'s guidance must not hand a worker a board-mutating command; either would install a second writer without a decision being taken. And the sibling's board-wide lock supersedes card scope for board mutations, so a lead that batches several card moves takes one lock for the batch rather than one per card. Consumed, not restated (`spec.md` §A.11, §C).

**B.3 The guard anchor is re-verified at run-phase, not trusted from here.** This plan names `TestMCP_NoAskUserQuestion` (`internal/cli/mcp_boundary_test.go`) as the pattern to extend, and that definition was verified present at authoring time. It is verified again at M0 before being cited in code, because B.2 is exactly the failure of trusting an authoring-time citation.

**B.4 Out-of-band note — nobody owns how a card enters the board.** `spec.md` §A.4 now states that entry into `plan` is an operator act rather than a dispatch: `backlog` has no owning session (`REQ-KB-012`), so no completion report exists for the lead to act on, and a lead that admitted cards on its own initiative would be generating work rather than scheduling it. What no SPEC in this family states is the **mechanism** — whether the operator admits a card through a CLI verb, a hand edit of board state, or an import from an issue tracker. `SPEC-KANBAN-BOARD-001` defines the card record (`REQ-KB-004`) and which moves are legal (`REQ-KB-008`) but never says who creates a card; its §A.2 explicitly contemplates a backlog item with no `spec.md` at all, so creation cannot be a side effect of plan-phase authoring either.

This is recorded rather than resolved because it is upstream of everything here: the lead's loop begins at the first dispatchable column and is complete without it, so making it a requirement of this SPEC would be claiming a contract this SPEC does not need in order to reach a decision it has no measurement to make. It is flagged for the family's next revision. The reason it is written down at all is that the same silence — a transition nobody states because each document assumes a neighbour states it — is exactly what produced the two ownership gaps this SPEC has now repaired twice.

**B.5 Out-of-band note — `worker-` is a reserved worktree name prefix.** Measured at `internal/cli/launcher.go`: `applyCCMode` calls `cleanupMoaiWorktrees`, which enumerates `git worktree list --porcelain` and removes every worktree whose base name begins with `worker-` under either of two bases — `<project>/.claude/worktrees/` and each directory under `~/.moai/worktrees/`. Launching a Claude-backed session is therefore a destructive act toward anything named that way.

Per-card worktree naming belongs to `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003`, not here, so this is a note and not a constraint on this plan. It is recorded because the collision is invisible from that sibling's side: the naming decision is made in one SPEC and the removal happens in a launcher neither SPEC's requirements mention, and a card worktree named `worker-<id>` would be deleted by the ordinary bootstrap of a Claude-backed worker with no error and no obvious cause.

**B.6 Out-of-band note — a role vacated after bootstrap is now reported, but nobody re-fills it.** Quorum is a bootstrap-time property here and nowhere else: REQ-KS-007 waits for it at the entry switch, REQ-KS-012 bounds that wait, and neither is evaluated again. When a worker session dies later, its role has no occupant. Measured, both siblings hand quorum to this SPEC — `SPEC-KANBAN-BOARD-001` §C and `SPEC-KANBAN-WORKTREE-001` §C each name "the quorum bound" among what belongs here — so the temporal gap sits on this side of the seam and is stated in `spec.md` §C rather than left implied.

The worktree sibling recovers the **card** and, read alongside this SPEC's dispatch rule, aims it back at the vacancy: `REQ-KW-011` releases the holder and leaves the column unchanged, `REQ-KW-012` makes a clean orphan "immediately re-dispatchable", and REQ-KS-019 dispatches to the session whose declared role owns that column — which is the role that just died. Neither sibling requirement is wrong; each does what it says.

**What changed at v0.5.0.** Through v0.4.0 this note ended by releasing the whole area, on the ground that observing the vacancy "is a new runtime-lifecycle obligation rather than a widening of an existing requirement, so there is no in-place shape available for it". That ground held for recovery and not for observation. The lead resolves the owning role's occupant on **every** dispatch; the vacancy is the null result of a selection REQ-KS-019 already governs, not a separate act of detection, so requiring the lead to report that null instead of swallowing it widens REQ-KS-019's own subject and costs no requirement — the same in-place move v0.3.0 made on REQ-KS-006. REQ-KS-019 now carries a **where**-clause: no dispatch into a vacancy, and the unoccupied role and waiting card both surfaced. `AC-KS-019` gains the conjunct that decides it, and judges the **report** rather than the refusal, because the refusal alone was already entailed and a do-nothing implementation satisfies it (`spec.md` §D.12).

**What the next revision still faces** is narrower than the three-way choice recorded here before, since one of its three options has been taken. Two items remain unowned: **recovery** — re-establishing an occupant, whether by re-quorum over the full role set, a targeted relaunch of the one missing role, or an operator act with no automation — and **monitoring**, observing a vacancy when the session dies rather than when the next card needs the role. The second is a genuine limit of the dispatch-time check and not a hole: where no dispatchable card is waiting for the vacated role, nothing is reported and nothing is stalled, so the reported set and the costly set coincide. Both belong with whichever SPEC that revision gives the role lifecycle to; REQ-KS-006 already owns the declaration either would read. Until then the operator recovery is `spec.md` §A.5's: launch the missing session and re-run the entry switch. Whoever implements this plan should now expect the symptom as a **named report** rather than a stalled column with no error — and should treat a stalled column that produces no such report as evidence the where-clause was implemented as a bare refusal, which is exactly what `AC-KS-019` fails.

## §C. Pre-flight (M0 — run these before any edit)

```bash
# 1. rename prerequisite present
test -d internal/kanban && echo RENAME_OK || echo RENAME_ABSENT

# 2. board sibling's card record present (holder + last-transition instant)
grep -rn 'Holder' internal/kanban/ --include='*.go' | head

# 3. messaging doctrine present on the base branch
test -f .claude/rules/moai/workflow/cross-session-messaging.md && echo DOCTRINE_OK || echo DOCTRINE_ABSENT
test -f internal/template/templates/.claude/rules/moai/workflow/cross-session-messaging.md && echo MIRROR_OK || echo MIRROR_ABSENT

# 4. guard anchor still has exactly one definition (B.3)
grep -rn '^func TestMCP_NoAskUserQuestion' --include='*_test.go' . | wc -l   # expect 1

# 5. single-origin discriminant behaves — run the form REQ-KB-005 prescribes,
#    from BOTH checkouts. The bare --git-common-dir form is not used alone:
#    it returns a repository-relative '.git' in the primary checkout, where
#    "the parent of it" is not a path (spec.md §A.11, REQ-KB-005).
git rev-parse --path-format=absolute --git-dir --git-common-dir
git -C <primary-checkout> rev-parse --path-format=absolute --git-dir --git-common-dir

# 6. the sibling's sole-writer rule is present at the version this SPEC defers to
grep -c 'REQ-KB-017' .moai/specs/SPEC-KANBAN-BOARD-001/spec.md   # expect >= 1

# 7. the env-access direction this SPEC's scans must cover, re-measured
grep -cE 'os\.(Setenv|Unsetenv|LookupEnv)\(' internal/cli/factory.go   # expect >= 1
grep -c 'os.Getenv(' internal/cli/factory.go                          # recorded as 0 at authoring

# 8. the launcher premise the backend decision rests on (REQ-KS-003, spec.md §A.8)
#    per-session half: the backend goes into the constructed env, and the launcher
#    hands that env to claude. The delivery is build-tagged, so BOTH definitions
#    are probed — a POSIX-only probe reports nothing on Windows (spec.md §A.8).
grep -n 'func setGLMEnv' internal/cli/glm.go
grep -n 'execOrSpawnClaude' internal/cli/launcher.go
grep -n 'syscall.Exec' internal/cli/launch_exec_posix.go
grep -n 'child.Env' internal/cli/launch_exec_windows.go
#    project-global half: each launcher writes state the other undoes
grep -n 'persistTeamMode(root, "glm")' internal/cli/launcher.go
grep -nE 'clearTmuxSessionEnv\(\)|removeGLMEnv\(settingsPath\)|resetTeamModeForCC\(root\)|cleanupMoaiWorktrees\(root\)' internal/cli/launcher.go

# 9. the reserved worktree prefix the sibling's naming must avoid (§B.5)
grep -n 'worker-' internal/cli/launcher.go
```

Any `*_ABSENT`, or a count other than 1 on check 4, halts per REQ-KS-002 / §B.3.

### C.1 Two shell conventions this plan inherits

- Every `grep` that asserts an **absence** is paired with a positive control before its result is believed. An absence-grep that cannot fire is not evidence.
- A table cell never carries a raw `|` inside a `grep -E` pattern. Commands that need alternation are written on their own line outside a table.
- A scan is written for the access direction the code uses, not the one that is easiest to type. Measured: `internal/cli/factory.go` accesses the environment five times — `os.Setenv` at 93, 95, 110; `os.LookupEnv` at 107; `os.Unsetenv` at 113 — and `grep -c 'os.Getenv('` over it returns **0**, so a read-only env scan over this SPEC's surface reports clean on code it never examined (`spec.md` §D.10).
- An equivalence claim is compared against a **file**, not against a memory. Every "unchanged from before" in this plan names the artifact holding "before" and the point it was recorded (`spec.md` §D.9).

## §D. Constraints

| # | Constraint | Where it binds | Surface in this SPEC |
|---|---|---|---|
| C1 | Template-First: template source before local counterpart, then `make build`, then commit the regenerated `internal/template/catalog.yaml` | REQ-KS-024 | the workflow document, the config template, the `moai init` question |
| C2 | Mirror pairs preserve their **measured** delta; a sanitized pair becoming byte-identical is a failure | REQ-KS-024 | the messaging-doctrine mirror, the workflow document mirror |
| C3 | `CLAUDE.local.md` §25: no SPEC ID, REQ/AC token, internal date, or commit SHA under `internal/template/templates/` | REQ-KS-024 | every template file this SPEC touches |
| C4 | Verification runs the **full** suite, not an affected-packages subset | REQ-KS-025 | a prior run-phase here missed a cross-cutting template guard by testing narrowly |
| C5 | Post-rename identifiers only; zero occurrences of `factory` in anything authored | REQ-KS-001 | all of it |
| C6 | Env-var names are constants in `internal/config/envkeys.go`; no inlined literals at call sites | REQ-KS-009 | the topology signal, the quorum-bound override if any |
| C7 | CLI may not call `AskUserQuestion` or `mcp__askuser__*` | REQ-KS-010 | the bootstrap surface specifically, which *wants* to prompt |
| C8 | Board state lives where `REQ-KB-005` puts it and is resolved by the probe `REQ-KB-005` prescribes, never from the session's own tree | `REQ-KB-005`; negative obligation here | no emitted guidance or dispatch tells a session to read its own tree's board copy, and none names a board path or probe form of its own |
| C8a | The `lead` is the **sole writer** of board state; every other session reads it | `REQ-KB-017`; negative obligation here | no dispatch instructs a worker to record progression in board state, and no emitted command gives a worker a board-mutating invocation |
| C9 | Primary-checkout branch guard: no branch-state mutation in the primary checkout | worktree sibling's `REQ-KW-006` | this SPEC performs none; named so nobody adds one |

C8, C8a and C9 have **no affirmative surface** in this SPEC — all three are obligations to not do something. They are recorded here rather than spending a requirement slot on a prohibition already owned by a sibling. C8a is new at v0.2.0 and is the constraint form of the deferral `spec.md` §C now makes explicit: the board sibling restored the sole-writer rule after the split deleted it, and part of what let the loss hide was that this SPEC disclaimed the question rather than deferring it by name.

## §E. Settled decisions and what each one costs

| Decision | Cost carried |
|---|---|
| Bootstrap is manual, and no spawn mechanism is designed | Every start of the board needs an operator at a keyboard. Accepted: sockets cannot launch sessions, so the alternative does not exist. |
| Quorum expiry **aborts** rather than proceeding with a partial team | An operator who mistimed a launch has to re-run the entry switch. Bought: a missing role becomes a loud immediate failure instead of a board that silently stops moving. |
| The choice of worker backend is expressed by **printing both commands**, not by asking | The guidance is longer and, for an unconfigured topology, roughly doubles in line count. Bought: the CLI stays inside its prompt prohibition without giving up operator choice. |
| The dispatch payload is a **typed record with no free-text field** | A future need to attach a note has nowhere to put it and will require a deliberate type change. That friction is the point. |
| The backward-compatibility baseline is recorded **post-rename** | The recording window is narrow and sequencing-sensitive; a run-phase that reorders M1 after M4 destroys it. Bought: the guard fires on regressions instead of on the rename. |
| `moai cg` stays rejected | Operators wanting a mixed-backend team are not served. Preserved boundary, not a gap. |
| Every equivalence claim is compared against a **recorded artifact**, not an inspection | M1 grows from one recording to four, and two of them are taken well before the window that would strictly require them. Bought: four claims that can actually fail. |
| The `lead` backend's fixity is enforced **per channel**, and the hand-launch channel is declared unclosable | The property the system can hold is narrower than the sentence "the lead's backend is not selectable" suggests. Bought: a checkable claim instead of an unfalsifiable one. |
| Write authority over board state is **deferred by name** to `REQ-KB-017`, not disclaimed | This SPEC's exclusions get longer and read as if they concede something. Bought: the gap that deleted the rule once cannot re-open from this side. |
| A session's role is a **declared datum**, not a value derived from its launch label | A second thing to carry at launch and a second thing to keep consistent. Bought: the only alternative — deriving the role from the label — requires either collapsing REQ-KS-014's distinct-label rule or removing the operator's backend choice, since one role corresponds to two-or-more labels. |
| The declaration's **carrier** is left to run-phase | The SPEC cannot be read to learn where the role lives, and two implementations could differ. Accepted: no measurement here favours the launch command over the registry over the discovery output, and fixing it would be fixing a decision on taste. |
| Entry into `plan` is an **operator act**, and the mechanism is left unowned | The board cannot be started without a human, and the family still has a hole at its upstream end. Bought: the hole is written down (§B.4) instead of being assumed closed by whichever SPEC the reader happens to be in. |
| Quorum is scoped to **bootstrap**; a vacancy is refused and reported at dispatch (REQ-KS-019 where-clause, v0.5.0), while **re-filling** the role stays released | A worker that dies mid-run leaves its column unowned. Through v0.4.0 the whole area was released and the board stalled silently — the §A.5 failure shape arriving after bootstrap instead of at it. Bought at v0.5.0: the stall becomes a named report, at the cost of no requirement, because the vacancy is the null result of a selection REQ-KS-019 already performs rather than a new act of detection. Still released: recovery, and detection independent of a dispatch attempt — the latter reporting nothing only in the case where nothing is stalled (§B.6, `spec.md` §C, §D.12). Ceiling unchanged and reported honestly at 25 of 25, measured. |

## §F. Milestones

### M0 — Preflight and prerequisite gates

Run §C. Halt on any absence. Re-read the landed messaging doctrine and diff its role-boundary-dispatch clause against `spec.md` §A.9; surface any delta as a blocker **before M2** (REQ-KS-002). Re-measure the five substrate properties of `spec.md` §A.3 against this runtime and record each result, and re-measure the launcher surface of `spec.md` §A.8 in both halves — the per-session backend reaching the launched session through the environment the launcher constructs for it — probed on both build-tagged launch paths, not on the POSIX one alone — and the project-global mutations each launcher performs against the other's (REQ-KS-003, checks 8 and 9 of §C).

The launcher measurement is placed at M0 rather than at M3, where the guidance is written, because it is the premise of a decision M3 only implements: where the per-session half did **not** hold, the printed-command mechanism of REQ-KS-013 would not deliver per-worker backends at all, and that is a finding to surface before writing the surface rather than after.

Satisfies: REQ-KS-002, REQ-KS-003.

### M1 — Record the four baselines (windows close after this)

Highest-reversibility-cost item, and the only milestone whose content is temporal rather than technical. Each recording below is a **durable artifact** — a file a verifier can re-read at verification time — because the failure being closed is the disappearance of the comparand, and a value held in working memory during the change disappears with it.

1. **The entry-switch baseline** (REQ-KS-011). With the rename landed and **before touching the entry switch**, record the parse result, environment mutation, and exec argument vector of `internal/cli/cc.go` and `internal/cli/glm.go` for the no-topology-configuration path. Assert the recorded artifact contains post-rename identifiers and no retired one; a baseline failing that assertion is discarded and re-recorded, never reconciled.
2. **The human-gate inventory** (REQ-KS-023). Each gate's identity, its position in the ordering, and what it gates. Its window closes at M5, but it is taken here — a gate quietly dropped is the most expensive silent failure in this SPEC, and it is invisible without a row-by-row comparand.
3. **The coder chain's carried-over behaviors** (REQ-KS-022). The verify exit gate's severity partition, rung attribute and re-entry ceiling; the revision dedup predicate's inputs and matching rule; the goal preset's arming rules. Window closes at M5.
4. **Each mirror pair's classification and pre-change `diff`** (REQ-KS-024). Window closes at M6, and the classification is the one recorded here that must be **re-measured** rather than carried from `spec.md` — pair classification is time-varying, so the artifact records what was measured at this point, and M6 compares against that measurement rather than against a document.

Satisfies: REQ-KS-011 (recording half), REQ-KS-022 / REQ-KS-023 / REQ-KS-024 (recording halves).

### M2 — The dispatch payload type and the protocol (highest design reversibility)

Define the payload as a typed record: SPEC identifier, file path, contract section reference. **No free-text field.** Then the protocol around it — addressing by label plus short reference with the refusal-supplied-reference re-send; reply as nudge and `progress.md` as truth; role-ownership routing — selecting by the **declared role** of REQ-KS-006 rather than by label, a distinction that matters because one role corresponds to two-or-more possible labels (`spec.md` §A.13) and routing by label would be routing by the operator's backend choice; the three prohibitions; idempotency against board state.

The dispatch cycle this protocol drives is `plan → run → review → sync`, `review` included; its two ends — the operator's admission into `plan` and the lead's terminal write into `done` — are not dispatches and are not implemented here (`spec.md` §A.4, §B.4 above).

This lands before the bootstrap because the payload's shape is the decision most expensive to revisit: every later surface that sends a dispatch is written against it.

Satisfies: REQ-KS-016 … REQ-KS-021.

### M3 — Topology, configuration, and the emitted command set

The five roles and the lead's no-column property; the run-session count default 1 / max 2 with a named-error rejection; stable labels, and — new at v0.3.0 — the **role declaration** each session carries alongside its label, resolvable from a session that is not the `lead` (REQ-KS-006, `spec.md` §A.13). The declaration's **carrier** is chosen here — the SPEC deliberately fixes none — and whatever is chosen must be readable outside the lead, because `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007` and `REQ-KW-011` resolve the `lead` occupant from sessions that are not it.

Its **contract** is consumed earlier than that, and the ordering is deliberate rather than an oversight: M2's routing selects by declared role and M4's quorum accounts over declared roles, both of which are written against REQ-KS-006's contract (a declaration exists, is distinct from the label, and resolves from any session) and neither of which needs to know the carrier. That separation is why the contract could be fixed in the SPEC while the carrier was left open — and it is the reason M2 keeps its position ahead of M3 despite consuming something M3 implements. The topology configuration under `.moai/config/`, its `moai init` question, and the prohibition on expressing any topology parameter as an entry-switch argument. Env-var constants in `envkeys.go`.

Then the guidance itself: one fixed lead line, one-or-two lines per worker role, distinct labels, no mixed-backend line — emitted to stderr, never asked. The `lead` line's fixity gains a **refusal per channel** at v0.2.0: a configured `lead` backend fails the load with a named error, an entry-switch backend argument fails to parse, an environment variable carrying a backend leaves the emitted line unchanged (`spec.md` §A.12). The configured channel is the one that must fail loudly — a key silently dropped looks exactly like a key silently honored.

Env-var constants land here rather than at their call sites, and every access to them is written with C6 in mind in **both** directions: this surface writes more than it reads (`spec.md` §D.10).

Satisfies: REQ-KS-004 … REQ-KS-006, REQ-KS-008, REQ-KS-009, REQ-KS-013, REQ-KS-014.

The `moai init` question this milestone adds writes a section into `.moai/config/sections/`, the same directory `moai glm` writes `team_mode` into. They do not collide — different section files — but the residual `team_mode` there must not be read as a record of the running team's composition; after a mixed launch it holds whatever the last launcher wrote (`spec.md` §A.8).

### M4 — The entry switch, the quorum loop, and the prompt boundary

Wire the switch: configuration present → print guidance and poll for quorum; absent → fall through. The quorum bound from configuration (default 300s), and its expiry path — name answered and unanswered roles, exit non-zero. The `AskUserQuestion` absence guard on the bootstrap source, written in the image of the anchor re-verified at M0, with its positive control. The mixed-backend rejection preserved with its sentinel.

Then compare against the M1 baseline (REQ-KS-011, comparison half).

Satisfies: REQ-KS-007, REQ-KS-010, REQ-KS-011 (comparison half), REQ-KS-012, REQ-KS-015.

### M5 — Coder-chain wiring (mostly a no-op, deliberately)

Point the existing chain at a card. The verify exit gate, the revision dedup predicate, and the goal preset are re-used unchanged; the human gates are untouched. If this milestone produces a large diff, something has been reimplemented that should have been reused.

Close both equivalence claims against the M1 recordings — the chain-behavior artifact and the gate inventory — comparing row by row rather than by inspection, and judging each artifact's provenance as well as its content.

Satisfies: REQ-KS-022, REQ-KS-023.

### M6 — Template mirror, neutrality, catalog, and full-suite verification

Mechanical, and last on purpose. Template source first, `make build`, commit the regenerated catalog. Compare each touched pair against the classification and pre-change `diff` recorded at M1 rather than against this document. Run the neutrality guard and the full test suite.

The four obligations REQ-KS-024 carries are verified **separately**, not as one green light: the pair deltas (AC-KS-024), the two neutrality authorities with their exit codes recorded individually (AC-KS-026), the catalog's regeneration and its presence in the commit (AC-KS-027), and the four-pattern content scan with its positive control (AC-KS-028). A single aggregate verdict here is the defect the v0.2.0 audit repair removed.

Satisfies: REQ-KS-001 (final sweep), REQ-KS-024, REQ-KS-025.

## §G. Anti-patterns

- **AP-1 — Citing a guard without confirming its definition.** The exact defect that produced the dead `TestNew_NoAskUserQuestion` reference (§B.2). Every guard citation in code is preceded by a `grep` for its `func` definition returning exactly one hit.
- **AP-2 — Recording the baseline late.** Taking the M1 recording after the entry switch has been edited produces a baseline that agrees with the change it was meant to detect. The recording is worthless and its passing comparison is a tautology.
- **AP-3 — Checking the pointer-only rule by grepping for requirement tokens.** Passes on a paraphrase. The check is the type's field set, not the payload's text.
- **AP-4 — Asserting only the happy path on quorum.** An implementation that never times out passes a quorum test that only exercises success. The expiry row is the load-bearing one.
- **AP-5 — Reading the emitted guidance instead of parsing it.** "It looks right" is not a check, and the mixed-backend absence in particular is an absence claim that needs a positive control.
- **AP-6 — Resolving board state from the session's own tree.** Silently forks the board per worktree. `REQ-KB-005` owns the rule; this SPEC's obligation is to emit nothing that contradicts it.
- **AP-7 — Reimplementing a sibling's requirement because it is more convenient than depending on it.** A second worktree path helper, a second WIP check, a second holder release. Each one is a fork of a contract that already exists.
- **AP-8 — Testing narrowly at M6.** Named because it already happened here once: a prior run-phase missed a cross-cutting template guard by running an affected-packages subset.
- **AP-9 — Scanning the environment in the read direction only.** A `grep` for `os.Getenv(` over this SPEC's surface returns zero against code that never reads that way — measured, that pattern occurs **0** times in `internal/cli/factory.go` while five write and lookup sites do. The scan then reports clean on a surface it did not examine, and a violation authored as `os.Setenv` is invisible. All five access forms, both directions, with a positive control in each (REQ-KS-009, `spec.md` §D.10).
- **AP-10 — Claiming "unchanged from before" with no recorded before.** The claim cannot fail: at verification time the original is gone, the comparand is a recollection, and every implementation passes including the one that widened the re-entry ceiling or dropped a gate. Each equivalence claim in this SPEC names the M1 artifact it is compared against (REQ-KS-011, REQ-KS-022, REQ-KS-023, REQ-KS-024).
- **AP-11 — Recording a baseline after the edit.** The generalization of AP-2 to the three recordings added at v0.2.0. A baseline taken afterward agrees with the change it was meant to detect; each criterion therefore judges provenance as well as content.
- **AP-12 — Telling a worker to update the board.** A dispatch that instructs a session to mark its own card done, or guidance that hands a worker a board-mutating command, installs a second writer without anyone deciding to — and `REQ-KB-017` was lost once already by nobody claiming it. Workers write `progress.md`; the `lead` reads it and writes the board.
- **AP-13 — Reporting one verdict for REQ-KS-024's four obligations.** "Neutrality passed" is an aggregate over two independent authorities, a catalog check, and a four-pattern scan; three of the four can be broken while the aggregate reads green. Each exit code and each scan result is recorded separately (AC-KS-024, AC-KS-026 … AC-KS-028).
- **AP-14 — Deriving a session's role from its launch label.** The shortest path to "which session is the `plan` session" and the one this SPEC's own text forbids: an unconfigured worker role emits two commands and the `run` role may deploy two sessions, while every emitted label must differ, so one role corresponds to two-or-more labels and the mapping does not invert. An implementation that parses a role out of a label passes a naive declaration test and misroutes the moment an operator picks the second printed command (REQ-KS-006, `spec.md` §A.13, AC-KS-030).
- **AP-15 — A role declaration only the lead can read.** Satisfies this SPEC's dispatch routing completely and silently breaks `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007` and `REQ-KW-011`, both of which resolve the `lead` occupant from a session that is not the lead. The failure surfaces in the sibling's gates, not here, which is what makes it worth naming here.
- **AP-16 — Reading `team_mode` as the team's composition.** After a mixed launch `.moai/config/sections/llm.yaml` holds whatever the last launcher wrote, which is a fact about launch order and not about which sessions run under which backend (`spec.md` §A.8). Anything wanting composition reads role declarations.

## §H. Out of Scope

### Out of Scope — carried from spec.md §C

- Everything the board and worktree siblings own, consumed here by requirement id only.
- Any session-spawning mechanism, any change to the messaging channel, any admission of the mixed backend, and any interactive prompt on the CLI.
- Repairing `internal/cli/CLAUDE.md`'s stale citation — recorded in §B.2 as an out-of-band note, deliberately not a requirement.

### Out of Scope — plan-phase deliverables

- Any code. This plan authors no Go, no template, and no configuration; it is the ordering and the constraint record for a run-phase that has not begun.
- Milestone time estimates. Ordering is by reversibility and dependency; there are no durations here to miss.

## §I. Cross-references

- `spec.md` — the requirements this plan implements, and §A.6 / §A.7 for the two corrections that shaped M0 and M1.
- `acceptance.md` — the criteria each milestone is judged against.
- `design.md` — the shape of the topology, the bootstrap, and the dispatch record.
- `research.md` — what was measured, and the alternatives that were rejected.
- `SPEC-KANBAN-BOARD-001` `plan.md`, `SPEC-KANBAN-WORKTREE-001` `plan.md` — the two sibling plans whose milestones must land before M2.
