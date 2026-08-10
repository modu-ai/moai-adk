---
id: SPEC-KANBAN-BOOTSTRAP-001
title: "Kanban session topology, bootstrap, and the pointer-only dispatch protocol"
version: "0.4.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, bootstrap, topology, dispatch, cross-session-messaging, quorum, backend, entry-switch, sole-writer, role-resolution"
tier: L
dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-WORKTREE-001]
related_specs: [SPEC-KANBAN-MULTISESSION-001, SPEC-FACTORY-MODE-001]
---

## HISTORY

- **v0.4.0** (2026-08-11) — Plan-audit delta repair, four defects, **no requirement added and no criterion added**: the SPEC is at the Tier L requirement ceiling of 25 and five criteria over it, so all four closed by amendment in place. Three of the four are the same failure in different clothing — a rule this SPEC restated instead of citing, a mechanism it named instead of stating the guarantee, and a disclaimer it wrote instead of reading its own criterion.

  **(1) §A.11 carried a copy of a sibling's rule, and the sibling changed underneath it.** The section named `.moai/state/kanban/` as the board store and "the parent of `git rev-parse --git-common-dir`" as the resolution — both `REQ-KB-005`'s to state. `SPEC-KANBAN-BOARD-001` has since moved the store to `.moai/state/kanban-board/` (its §A.3(e), clearing a collision with `SPEC-KANBAN-RENAME-001` `REQ-KR-009`, which keeps the old name for **session records** resolved per-tree) and forbidden the bare probe form standing alone. The copy did not move: `grep -rc 'kanban-board' .` returned **0** in all six files of this directory. The stale probe had propagated into the execution gates, where it was wrong in the direction that hides — measured, the bare form prints `/Users/goos/MoAI/moai-adk-go/.git` from a worktree and `.git` from the primary checkout, so `AC-KS-001` as written passed from a worktree and failed from the primary. Restatement deleted in favour of citation at §A.11, `AC-KS-001`, `plan.md` §C check 5 and constraint C8, and `research.md` §D.

  **(2) Normative text named a POSIX-only system call as the mechanism (the MP-6 must-pass failure).** §A.8 rested the backend decision on `launchClaudeDefault` replacing the process "via `syscall.Exec`". Measured, the call at `launcher.go:791` is `execOrSpawnClaude`, which has two build-tagged definitions: `launch_exec_posix.go` (`//go:build !windows`) calls `syscall.Exec`, while `launch_exec_windows.go` (`//go:build windows`) spawns a child with `exec.Command` and assigns `child.Env = env`, its own comment recording that `syscall.Exec` returns `syscall.EWINDOWS` there at runtime. The **conclusion** — the constructed environment reaches the launched backend, so interleaved launches give each worker its own — holds on both platforms; only the named mechanism was wrong, and naming it prescribed on Windows the exact call that file exists to avoid. §A.8, `REQ-KS-003` and `AC-KS-003` are rephrased on the environment guarantee; the POSIX call now appears only as one platform's implementation, alongside the Windows one, at §E, `plan.md` §C check 8 and `research.md` §J.1. `launchClaudeDefault`'s own doc comment still describes only the POSIX path and predates the split, which is where the reading came from.

  **(3) §D.11 disclaimed a check its own criterion performs, and pointed at the wrong one.** It said quorum accounting "is REQ-KS-012's criterion and is not re-checked here" — but `AC-KS-030`'s fourth conjunct performs exactly that check three lines above, and `AC-KS-012` carries no declaration-related clause at all, correctly: the bound is REQ-KS-012's subject and the accounting **key** is REQ-KS-006's. §D.11 rewritten to record the split, with a pointer added under `AC-KS-012` so the three surfaces agree.

  **(4) The seventh unowned area: nobody recovers a role vacated after bootstrap.** Quorum here is bootstrap-scoped — REQ-KS-007 waits for it, REQ-KS-012 bounds that wait, neither is evaluated again — while both siblings' §C hand "the quorum bound" to this SPEC by name. A worker dying later leaves its role unoccupied; the worktree sibling recovers the card (`REQ-KW-011` releases the holder, `REQ-KW-012` makes a clean orphan immediately re-dispatchable) and REQ-KS-019 then aims it at the vacancy. Neither sibling requirement is wrong; the missing step is observing the vacancy at all. With 25 of 25 requirements used and this being a new runtime-lifecycle obligation rather than a widening of any existing one, it is **released explicitly** rather than owned: a §C entry naming what is unowned, why the ceiling forbids owning it here, and what a reader hitting the stall should do, plus `plan.md` §B.6 and a §E decision row — the treatment `plan.md` §B.4 already gives the unowned `backlog → plan` admission.

  **Budget position after the repair, re-measured in this worktree:** requirements 25 of 25, criteria 30 against 25 — unchanged in both directions, the five-criterion overflow being v0.3.0's and not widened here.

- **v0.3.0** (2026-08-11) — Plan-audit repair (0.857 against a Tier L threshold of 0.85 — a FAIL on three blocking findings, with all four v0.2.0 repairs verified closed and no regression). Three blocking defects are closed, one optional finding is taken, and no requirement is added: the SPEC sits at the Tier L ceiling of 25, so every fix is an amendment in place.

  **(1) Runtime role resolution was owned by no SPEC in the family — the F1 failure shape, again.** Two independent auditors reached it from opposite directions, here and in `SPEC-KANBAN-BOARD-001`. Five consumers depend on knowing which session occupies which role and each assumed another owned it: `REQ-KS-019` routes to "the session whose **declared role** owns the card's column" without defining declaration; `REQ-KS-004` defines the role *set* and elects `lead` but not runtime occupancy; `REQ-KS-006` defined only "a stable label set at launch", and did so for *addressability*; `SPEC-KANBAN-BOARD-001` `REQ-KB-004` records a *session identifier*, not a role, while its `AC-KB-017` requires refusing a runtime board write from a non-`lead` session — which presupposes the role is readable; and `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007` and `REQ-KW-011` both resolve "the session occupying the `lead` role" and defer the definition to `REQ-KS-004`, which does not carry it.

  The one derivable candidate is ruled out by this SPEC's own text. `REQ-KS-014` requires every emitted command's label to be distinct from every other's, while `AC-KS-013` has each unconfigured worker role emit **two** commands (one per supported backend) and `REQ-KS-005` permits **two** `run` sessions — so a role maps to two-or-more possible labels chosen by the operator at launch, and no label determines a role. `REQ-KS-006` is therefore widened in place from addressability to addressability **plus role declaration** (§A.13), with the label/declaration distinction stated rather than left to inference, and the declaration required resolvable from a session that is not the `lead` because the worktree sibling's gates run there. `AC-KS-030` observes it.

  **(2) The central backend decision rested on an unverified premise, and the premise is half wrong.** §A.8 and `REQ-KS-013` assumed `moai cc` and `moai glm` are per-session backend selectors. Measured in `internal/cli/launcher.go`, each is **both**: a per-session selector *and* a project-global mutation. The selector half holds — `setGLMEnv` (`internal/cli/glm.go:229`) writes the backend into the **process** environment and `launchClaudeDefault` replaces the process via `syscall.Exec`, which inherits it, so an interleaved launch does give each session its own backend. The half nobody had measured is that each invocation also writes project-global state that the other undoes: `applyGLMMode` persists `team_mode: glm` to `.moai/config/sections/llm.yaml` (`launcher.go:264`) and, in tmux, mutates the shared session environment; `applyCCMode` clears the tmux session environment (`:213`), strips GLM env from `.claude/settings.local.json` (`:218`), resets team mode (`:222`), and removes `worker-`-prefixed worktrees (`:227`). The consequence is now stated at §A.8 rather than assumed, the measurement is recorded at `research.md` §J, and `REQ-KS-003`'s re-measurement list is extended to cover the launcher surface — previously it re-measured five messaging-substrate properties while the launcher surface underneath the central decision was measured not at all.

  **(3) The dispatch cycle contradicted this SPEC's own role table.** §A.4 read "plan reports completion → lead instructs run → completion → lead instructs sync", omitting `review` — defined five rows above as owning the `review` column and running the verify exit gate — in the one place the protocol sequence is written, and contradicting `REQ-KB-003`'s fixed order `backlog → plan → run → review → sync → done`. Two further transitions were unstated: what admits a card from `backlog` (which has no owning session and so emits no completion report) and what moves it from `sync` to `done`. The cycle is corrected, `sync → done` is stated as the lead's terminal evidence-read write, entry into `plan` is stated as an operator act, and the fact that the *mechanism* of that operator admission is owned by no SPEC in the family is recorded in `plan.md` §B.4 rather than left silent.

  **(4) `REQ-KS-018` asserted a sibling's rule in normative voice** ("the resulting board write **shall** be performed by the `lead` alone"), in a SPEC that states five times that it duplicates none of `REQ-KB-017`'s content. Recast so the normative force sits on the dispatch this SPEC owns and the write stays `REQ-KB-017`'s.

  **Not amended, deliberately.** `REQ-KS-024`'s Template-First ordering clause and `AC-KS-019`'s missing positive control were both raised as optional and both declined: the first is genuinely unobservable after the fact, and the second's exhaustiveness is delegated to `AC-KB-017`, which does carry one.

- **v0.2.0** (2026-08-10) — Plan-audit repair (0.84, a narrow-delta FAIL rather than a rewrite). Four defects are closed and one consistency change rides along; no requirement is renumbered and the citation set of v0.1.0 is preserved intact.

  **(1) `AC-KS-009` scanned the wrong direction of an env-var access.** It looked for `os.Getenv("`, while the surface REQ-KS-009 governs is a **write**: measured in this worktree, `internal/cli/factory.go` performs `os.Setenv` at lines 93, 95 and 110, `os.LookupEnv` at 107, `os.Unsetenv` at 113 — and `grep -c 'os.Getenv('` over that same file returns **0**. A read-only scan was therefore not merely weak but vacuous on its own governed surface: it could not fire even against a deliberate violation, because the code never reads that way. The criterion now enumerates every access direction (§D.10).

  **(2) Three criteria demanded equivalence to a baseline nothing obliged anyone to record.** `AC-KS-022` and `AC-KS-023` required the coder chain and the human gates to be "unchanged", and `AC-KS-024` required a mirror pair's `diff` before the change to equal the `diff` after — while no requirement obliged the recording of that "before". At verification time the original is gone and the comparison can only be satisfied by assertion. `REQ-KS-011` already carries the correct shape for the entry-switch baseline; that shape is generalized in place to REQ-KS-022, REQ-KS-023 and REQ-KS-024, each of which now requires a **durable pre-change artifact** a verifier can re-read (§D.9). `AC-KS-020` was named alongside these by the audit and is **not** amended: it is three absence scans with positive controls and asserts no equivalence to a prior state.

  **(3) `REQ-KS-013`'s impossibility claim had no falsifying observation.** "The lead's backend is not operator-selectable" was satisfied equally by an implementation that never mentions it and by one that silently honors an override. The requirement now names the refusal per channel, §A.12 enumerates the three channels a backend can arrive through, and it records the fourth channel that **cannot** be refused — a hand-launched lead — as a finding rather than smoothing it over.

  **(4) `REQ-KS-024`'s compression was undisclosed and `AC-KS-024` bundled seven observations in one line.** It absorbs the predecessor's `REQ-KM-037` (template-first, `make build`, catalog), `REQ-KM-038` (mirror delta preservation) and `REQ-KM-039` (template content neutrality) — three separable obligations reported by a single pass/fail, under which six could be broken while it still passed. The absorption is now disclosed at §B.8 and §E, and the criterion is split (§D.7).

  **(5) Consistency with `SPEC-KANBAN-BOARD-001` v0.2.0.** That sibling restored `REQ-KB-017` — the `lead` session is the **sole writer** of board state — after the three-way split deleted it along with the rejected `column:` mechanism it had been bundled with. Part of what let the loss hide is that this SPEC's §C disclaimed the question ("this SPEC decides who is *told* about a card"), so between the two disclaimers nobody owned the write. Every place here that touches who may write the board, dispatch a card, or move it between columns now **defers explicitly** to `REQ-KB-017` (§A.4, §A.11, §C, REQ-KS-004, REQ-KS-018, REQ-KS-019). This SPEC elects the `lead` role; it defines no write path and duplicates none of REQ-KB-017's content.

- **v0.1.0** (2026-08-10) — Initial plan-phase authoring, split out of the superseded 59-requirement `SPEC-KANBAN-MULTISESSION-001` (plan-audit FAIL 0.87) together with `SPEC-KANBAN-BOARD-001` and `SPEC-KANBAN-WORKTREE-001`. Three corrections land with the split, and each is argued at the section named here rather than absorbed. **(1)** The predecessor's §A.7 and §D.3 cited `TestNew_NoAskUserQuestion` in `internal/cli/worktree/new_test.go` as a canonical static guard; measured in this worktree, that function has **zero definitions** and that file **does not exist** (§A.6). The criterion is re-anchored onto a guard that actually runs. **(2)** The predecessor's §D.1 required a backward-compatibility baseline but never said *when* it is recorded; because the rename prerequisite retires every identifier the baseline would encode, a pre-rename baseline makes every later comparison fail for the rename rather than for a regression (§A.7). **(3)** The predecessor's `column:` frontmatter field is rejected and removed; board state has a single origin under the primary checkout, owned by `SPEC-KANBAN-BOARD-001`, and this SPEC's obligation is to not contradict it (§A.11).

---

## §A. Context

### A.1 What already exists, and what the siblings own

`SPEC-FACTORY-MODE-001` (completed) shipped an entry switch on the `moai cc` / `moai glm` session launchers that opens **one** session pre-armed to drive a `plan → run → verify → sync` chain. `SPEC-KANBAN-RENAME-001` (prerequisite) renames that surface to Kanban Mode: the flag, the environment variables, the package, the state directory, the unsupported-backend sentinel, the goal preset, and the workflow document. Every identifier below is written in its **post-rename** form.

This SPEC adds the sessions: how many there are, what each is called, how they are launched, how the launch is bounded, and what one session is permitted to say to another. It does not add the board and it does not add the trees.

| Concern | Owner | Consumed here as |
|---|---|---|
| the column enumeration, the card record, the board state store and its single-origin resolution, WIP admission, the column↔status check, the unheld state | `SPEC-KANBAN-BOARD-001` | `REQ-KB-003`, `REQ-KB-004`, `REQ-KB-005`, `REQ-KB-008`, `REQ-KB-009`, `REQ-KB-011`, `REQ-KB-012` |
| worktree creation and disposal, orphan classification, stall detection, holder release, the holder-mutation lock | `SPEC-KANBAN-WORKTREE-001` | `REQ-KW-003`, `REQ-KW-005`, `REQ-KW-007`, `REQ-KW-011`, `REQ-KW-012`, `REQ-KW-013` |

Those requirement ids are cited, never restated. Where this SPEC needs one of them it names it; where a reader expects one of them here and does not find it, §C says why.

Three assets from the single-session chain are **repurposed, not discarded**, and all three live inside one coder session rather than across the board:

| Asset | Today | Under this SPEC |
|---|---|---|
| the verify exit gate (its severity partition, its rung attribute, its ceiling of at most two re-entries) | the exit of the one session's run-phase | the exit of a **coder session's** per-card run-phase |
| the revision dedup predicate | sync-phase suppression for the one chain | unchanged predicate, evaluated per card |
| the renamed goal preset (a documented goal condition; no Go symbol, no preset registry, no CLI verb) | drives the whole one-session chain | drives the **internal** chain of one coder session over one card |

### A.2 Preflight: a doctrine that lands ahead of this SPEC

The cross-session messaging doctrine — `.claude/rules/moai/workflow/cross-session-messaging.md` plus its template mirror — is the contract this SPEC's dispatch protocol implements. It lands **first, as its own pull request**, ahead of and independent of this SPEC. Sequencing it separately keeps a rule change that binds every session out of a feature SPEC's diff, where it would be reviewed as an implementation detail of the board rather than as doctrine.

The consequence is a simplification: at preflight the file is *expected present* on the base branch, so the check verifies presence rather than tracking an unpushed branch. Absence is a **halt**, not a tracking problem — proceeding against an unlanded contract would mean implementing a specialization of a clause nobody has agreed to yet.

It stays **outside** `dependencies:`, because `DependencyExistsRule` resolves that field against SPEC directories and a rule file is not a SPEC. The three entries that *are* in `dependencies:` are the rename and the two siblings.

**When the doctrine landed changed** — that is, when what merged differs from the version this SPEC was authored against — the reconciliation obligation binds. This SPEC's dispatch protocol is a *specialization* of the doctrine's role-boundary-dispatch clause (three conditions: declared role, pointer-only work item, isolated tree), so a change to those three conditions is a change to REQ-KS-016 and REQ-KS-019, and it is surfaced as a blocker before the dispatch milestone rather than absorbed silently. Landing first removes the uncertainty about *whether* the file will exist; it does not remove the obligation to read what actually landed.

The five substrate properties of §A.3 are likewise **re-measured** against the run-phase runtime. A property measured in a prior session is evidence about that session's runtime, not about this one.

### A.3 The messaging substrate — measured properties, and what each forbids

Each row names a design a naive reading would produce, and the measurement that kills it.

| Measured property | Design it forbids |
|---|---|
| Sending by **name alone is refused**; the send must carry the peer's short reference | An address book keyed on role name only. The lead retains the reference from the discovery tool's output, or from the refusal itself. |
| **Reply routing is not guaranteed** to reach the original sender | A lead that advances its board on received replies. Completion must be observable in the shared source of truth. |
| The sender's **bypass status is disclosed** to the receiver, and the receiver's inbound default keys on it | Assuming prompt delivery. A message from a bypassing sender is *more* likely to be held for approval, not less. |
| `--name <label>` passes through the launcher and appears in the peer list; `claude --help` binds `-n, --name <name>` | Deriving peer identity from the working directory. The label is a first-class runtime flag, so the bootstrap guidance can emit an executable command that sets it. |
| A full round trip was demonstrated end to end (send → receive → process → reply) | Treating the channel as theoretical. It works; it is just not *reliable*, which is a different claim. |

Two constraints follow from the substrate rather than from any single measurement:

- **Sockets cannot launch sessions.** There is no mechanism by which one session spawns an independent peer. Bootstrap is therefore manual **by necessity**, not by preference (§A.5).
- **A message is not consent.** The receiving runtime is told the text came from another session, not from the user. A peer reply is never approval for a gated action.

### A.4 The five roles, and the two knobs that are not one knob

| Role | Column it owns | Notes |
|---|---|---|
| `lead` | none — it watches all six | Monitors the board and dispatches. Does no card work. It is also the **sole writer** of board state — a rule owned by `REQ-KB-017` and not restated here; this SPEC elects the role, the board sibling says what the role may write. |
| `plan` | `plan` | Authors plan-phase artifacts for the card. |
| `run` | `run` | The **coder** session. Runs the internal chain of §A.10. |
| `review` | `review` | Runs the verify exit gate for the card. |
| `sync` | `sync` | Runs sync-phase and the close. |

`backlog` is a queue with no session; `done` is terminal with no session. Both are reported not-dispatchable by the board (`REQ-KB-012`), so this SPEC does not re-derive that — it simply never addresses a session for either.

The dispatch cycle runs the full column order `REQ-KB-003` fixes, `review` included: **plan reports completion → lead instructs run → completion → lead instructs review → completion → lead instructs sync**. Each arrow is one dispatch, and each "reports completion" is the lead reading evidence from the card's `progress.md` rather than trusting a reply (REQ-KS-018). Omitting `review` — as v0.1.0 and v0.2.0 did, in the one place the sequence is written — contradicted the role table five rows above, which gives `review` a column and the verify exit gate; and a protocol whose written form skips a role is how a role quietly stops being dispatched to.

The two ends of the cycle are not dispatches, and saying what they are is cheaper than leaving a reader to infer it:

- **`sync` → `done`** is the same act as every other arrow with the dispatch removed. The `sync` session reports completion in its `progress.md`, the lead reads it, and the lead writes the terminal transition; there is no session in `done` to address (`REQ-KB-012`), so the write is where the cycle ends.
- **`backlog` → `plan`** cannot be an arrow at all. `backlog` has no owning session (`REQ-KB-012`), so nothing reports a completion the lead could act on, and a lead that admitted cards on its own initiative would be generating work rather than scheduling it. Entry into `plan` is therefore an **operator act** — the same operator who launches the sessions decides what enters the board's working columns — and the lead's loop begins at the first dispatchable column. The *mechanism* by which an operator performs that admission is owned by no SPEC in this family; it is recorded as an out-of-band note in `plan.md` §B.4 rather than assumed here, because assuming it is what produced the omission this repair closes.

**The WIP limit and the deployed session count are two different knobs and are not conflated.** The `run` column admits at most two cards concurrently — a property of the board, owned by `REQ-KB-009` and held independent by `REQ-KB-010`. The number of `run` sessions deployed defaults to **1** and is raisable to **2** by configuration — a deployment knob, owned here. A configured value outside that range is rejected at load with a named error rather than clamped, because a silently clamped topology is a topology the operator did not write and will not recognize when it misbehaves.

With WIP 2 and one coder session, the second card enters `run` and waits there **unheld**. That state is the board's (`REQ-KB-011`) and is a legal steady state, not an error; the lead dispatches it the moment a coder session frees up. Gating admission on a free session instead would make the effective WIP equal the session count — which is the conflation `REQ-KB-010` exists to forbid, arriving by the back door.

### A.5 Bootstrap is manual, and the quorum bound aborts

The bootstrap is: **print guidance → the operator launches the sessions by hand → the lead polls the discovery tool until quorum → the lead dispatches.** No mechanism to spawn a peer is designed here, and none is wished for (§A.3).

The guidance must be *executable* — a copyable command per role, carrying the `--name` label the lead will address.

**The quorum wait is bounded by a configuration value under `.moai/config/`, defaulting to 300 seconds.** The bound is a configuration key, not an argument on the entry switch (§A.6). On expiry the bootstrap names which roles answered and which did not, and **exits non-zero**; it does not proceed with a partial team and does not re-print and keep waiting.

Proceeding with a partial team is the tempting option and the wrong one. A missing role means its column has no owner, so the first card to reach that column sits there while the lead — correctly, per REQ-KS-019 — refuses to dispatch it to anyone else. Nothing errors; the board simply stops moving, which presents as a **hang rather than as a fault**, and a hang is the failure shape that costs the most to diagnose. Aborting at bootstrap converts that silent late failure into a loud immediate one, and the recovery is cheap: the operator launches the missing session and re-runs the entry switch.

### A.6 The CLI cannot prompt — and the guard the predecessor cited does not exist

`internal/cli/CLAUDE.md` § Conventions states the boundary: CLI code MUST NOT call `AskUserQuestion` or any `mcp__askuser__*` tool; the CLI runs in subagent context, the orchestrator owns user interaction, and an interactive prompt is replaced by positional arguments plus `--flag` defaults plus structured stderr.

This binds here precisely because bootstrap *wants* to prompt. "Launch these four sessions, press enter when ready" is the natural shape of the thing and is prohibited. The bootstrap emits guidance to stderr and then polls; it never blocks on an answer it is not allowed to ask for. The same prohibition removes the natural implementation of backend selection (§A.8), which is why that choice is resolved by what is *printed* rather than by what is asked.

**The canonical guard the predecessor named is dead.** `SPEC-KANBAN-MULTISESSION-001` §A.7 and §D.3 cited `TestNew_NoAskUserQuestion` in `internal/cli/worktree/new_test.go`. Measured in this worktree:

- `grep -rn 'TestNew_NoAskUserQuestion' --include='*.go' .` returns **six hits, all comments** in other test files referring to the name — and **zero `func` definitions**.
- `internal/cli/worktree/new_test.go` **does not exist**. That directory holds `clean.go`, `done.go`, `guard.go`, `recover.go`, `remove.go`, `render.go`, `root.go`, `shared.go`, `sync.go` and their tests; there is no `new*.go` file at all.

So the pattern is real and widely deployed — 29 `func Test…NoAskUserQuestion` definitions exist across the tree, among them `TestMCP_NoAskUserQuestion` (`internal/cli/mcp_boundary_test.go`), `TestWeb_NoAskUserQuestion`, `TestPropose_NoAskUserQuestion`, `TestDoctor_NoAskUserQuestion`, and `TestGoal_NoAskUserQuestion` — but the *particular instance* the predecessor pointed at is a ghost that survived only in the comments of the tests that copied it. A criterion citing it would be unsatisfiable, and an implementer told to "extend the canonical guard" would find nothing to extend. This SPEC therefore anchors on a guard verified to have a live definition, and the run-phase re-verifies that anchor before citing it, because the same rot can happen again.

The stale citation in `internal/cli/CLAUDE.md` itself is a documentation defect this SPEC did not create and does not own; it is recorded in `plan.md` §B so the next reader does not re-propagate the dead name.

### A.7 The backward-compatibility baseline, and when it is recorded

Where no topology configuration is present, the entry switch falls through to the existing single-session chain with behavior **identical to the pre-change launcher**. This is the one property a reader will assume rather than check, so it is made mechanical: the switch's parse result, its environment mutation, and its exec argument vector must equal a recorded baseline.

The predecessor required the comparison but never said **when the baseline is captured**, and that omission is not cosmetic. `SPEC-KANBAN-RENAME-001` renames the flag, the environment variables, the package, the state directory, the sentinel, and the goal preset. A baseline captured before that rename lands encodes the **retired** identifiers — the old flag token in the argument vector, the old variable names in the environment mutation. Every subsequent comparison then fails, and it fails *for the rename*, not for a regression. The guard would fire on the one change everybody already agreed to, and it would fire on every run until somebody re-recorded it — at which point the honest response and the lazy response look identical, and the guard has taught its readers to re-record rather than to investigate. A guard that cries wolf on a sanctioned change is worse than no guard, because it consumes the attention the real regression would need.

So the baseline is captured **against the post-rename tree**: after the rename prerequisite has landed on the base branch, and before this SPEC's first edit to the entry switch. That window is narrow by construction and it is the only window in which the recording means what the comparison will read.

### A.8 Backends: the lead is fixed, the workers are the operator's choice

The `lead` session is always the Claude-backed entry-switch invocation, emitted exactly once, and its backend is **not operator-selectable**. Every worker role — `plan`, `run`, `review`, `sync` — launches under either the Claude launcher or the GLM launcher, and the operator chooses which at the moment they launch it.

So the guidance is not a fixed script. It has to present each worker launch as a choice while the lead's own line stays fixed, and it has to do that **without asking**, because §A.6 forbids the CLI from prompting — which rules out the natural implementation of "let the operator choose", an interactive selection.

The resolution takes the choice out of the prompt and puts it in two places that are already permitted:

- **Where the topology configuration names a backend for a role**, the guidance prints exactly one command for that role, in that backend.
- **Where it does not**, the guidance prints **both** forms for that role and the operator runs whichever they want. Printing both is a choice offered without a question asked.

This stays mechanically checkable rather than merely readable. Every printed command must parse and carry a distinct label; the lead must contribute exactly one line of the fixed form; each worker role must contribute one line (configured) or two lines whose backends are exactly the two supported ones (unconfigured); and no line may name the rejected mixed backend.

**What the two launchers actually do, measured rather than assumed.** Everything above rests on a premise nobody had checked: that `moai cc` and `moai glm` are per-session backend selectors, so an operator can launch `plan` under one and `run` under the other in the same project. The premise is half right, and the half that is wrong was invisible because this SPEC re-measures five messaging-substrate properties (REQ-KS-003) while measuring nothing about the launcher surface its central decision stands on.

The selector half holds, and the property it holds on is a **guarantee about the environment, not a named system call**: `setGLMEnv` writes the backend into the launching process's environment, and the launcher hands that constructed environment to the `claude` it launches. The backend therefore travels with the launched session and not through any file — the GLM path's own comment records that `settings.local.json` injection was deliberately removed so the backend would *not* leak into later `claude` invocations. Interleaved launches do give each worker its own backend, which is what §A.8 needs.

**The mechanism is platform-split and the guarantee is not, which is why the requirement is phrased on the guarantee.** Measured: `internal/cli/launcher.go:791` calls `execOrSpawnClaude`, which has two build-tagged definitions. `internal/cli/launch_exec_posix.go` (`//go:build !windows`) calls `syscall.Exec(claudeBin, args, env)` — the current process *becomes* `claude` and inherits the environment. `internal/cli/launch_exec_windows.go` (`//go:build windows`) spawns a child with `exec.Command`, sets `child.Env = env` explicitly, and propagates the exit code; its own comment records that Windows has no `execve(2)` and that `syscall.Exec` returns `syscall.EWINDOWS` there at runtime. Both paths deliver the constructed environment to the launched backend, by different means. Naming `syscall.Exec` as *the* mechanism — which this section did through v0.3.0 — asserts on Windows something the code explicitly does not do, while the conclusion the section draws survives on both platforms untouched. Where the POSIX call is mentioned below or in `plan.md`, it is named as one platform's implementation and never as the contract.

The half that was assumed away is that each invocation *also* mutates project-global state, and the two invocations undo each other's mutations: `moai glm` persists `team_mode: glm` into `.moai/config/sections/llm.yaml` and, inside tmux, writes the shared tmux session environment; `moai cc` clears that tmux session environment, strips the GLM variables from `.claude/settings.local.json`, resets `team_mode` while printing what it cleared, and removes `worker-`-prefixed worktrees under the two known worktree bases. The bootstrap prescribes running these commands up to five times in one project, and §A.5 prescribes re-running the entry switch after a quorum abort, so the interleaving is the normal case rather than an edge one (`research.md` §J).

Three consequences follow, and they are stated rather than smoothed over:

- **The interleaved sequence is tolerable**, because none of the mutated project-global state determines the backend of a session that has **already been launched** — the environment it received at launch is fixed on both platforms, whether the launcher replaced itself or spawned a child. A worker launched under GLM stays on GLM when a later `moai cc` resets `team_mode`.
- **The residual `team_mode` is not a record of the team's composition** and must not be read as one. After a mixed launch it holds whatever the last launcher wrote, which is a fact about launch *order*, not about which sessions are running under which backend. Anything wanting the team's composition reads the role declarations of §A.13, not the persisted mode.
- **`worker-` is a reserved worktree name prefix.** `moai cc` removes worktrees carrying it under the two known bases, so a per-card worktree named with that prefix would be destroyed by the ordinary act of launching a Claude-backed worker. The naming is `SPEC-KANBAN-WORKTREE-001` `REQ-KW-003`'s and not this SPEC's to set; the collision is recorded in `plan.md` §B.5 so the sibling does not walk into it.

**The mixed backend stays rejected**, sentinel unchanged. A mixed leader/teammate backend contradicts the one-session/one-backend premise the state record encodes, and sockets cannot launch sessions, so the messaging layer does not replace tmux. This is deliberate avoidance of a destructive change, not an unfilled gap, and it is recorded here so a later audit does not read it as one. Its exclusion is not a hole in the operator's choice; it is a preserved boundary.

### A.9 Instructions are pointers, and that is structural

A dispatch carries a **SPEC identifier, a file path, and a contract section reference** — never requirement text, never acceptance-criterion text, never a restatement of a contract that already exists on disk. The SPEC file is the single source of truth. This is the doctrine's pointer-only condition, and it is also the only design that survives the reply-unreliability of §A.3: a pointer is re-sendable and idempotent; a body is not.

Checking it by grepping the payload for requirement tokens is a **weak proxy** — a grep passes on a payload that paraphrases the requirement in its own words, which is exactly the failure mode worth catching. The property is therefore made structural rather than textual: the payload is a typed record whose fields are an identifier, a path, and a section reference, with **no free-text field**, so carrying a body is unrepresentable rather than merely discouraged. An implementer cannot smuggle prose past a type that has nowhere to put it.

Idempotency is load-bearing for the same reason. The substrate holds messages for approval more often than the naive reading expects, so a re-send is a normal event rather than an exceptional one, and a re-send for a card already in progress must advance nothing and corrupt nothing.

### A.10 The coder session's internal chain

Each `run` session drives its card through the chain the single-session mode already defines. This SPEC introduces **no second chaining mechanism** — the verify exit gate fires at the exit of a card's run-phase with its severity partition, its rung attribute, and its ceiling of at most two re-entries carried over unchanged; the revision dedup predicate is evaluated per card with no change to its inputs or its matching rule; and the renamed goal preset drives one coder session over one card, its arming rules unchanged (armed only after the plan-to-run approval, armed alongside the work rather than in place of it, bounded by flags rather than by prose).

Most of what could be written here is a **non-goal**, and it lives in §C accordingly. The affirmative content is small: reuse, unchanged. This SPEC changes neither the number, the ordering, nor the semantics of the human gates the chain already carries, and adds no gate of its own.

### A.11 Board state has one origin, and this SPEC does not contradict it

Board state lives in a single-origin store beneath the **primary checkout**, and both that store's location and the rule that resolves it are `SPEC-KANBAN-BOARD-001` `REQ-KB-005`'s. They are **cited here and not restated**, and the citation is load-bearing rather than stylistic. Through v0.3.0 this section carried its own copy of the rule — naming `.moai/state/kanban/` and "the parent of the directory reported by `git rev-parse --git-common-dir`" — and the sibling has since changed both halves: it moved the store to `.moai/state/kanban-board/` (its §A.3(e), clearing a collision with `SPEC-KANBAN-RENAME-001` `REQ-KR-009`'s session-record directory, which keeps the former name and its per-tree resolution) and it now forbids the bare `--git-common-dir` form from being used alone. The copy did not move with it. Measured in this directory before the repair, `grep -rc 'kanban-board' .` returned **0** in every one of the six files — this SPEC was naming the *session-record* path as the board's and prescribing a probe the sibling refuses. A restated rule is a rule with two owners and one maintainer, and this is what that costs.

The probe half was wrong in the direction that hides. Measured: from a worktree, `git rev-parse --git-common-dir` prints `/Users/goos/MoAI/moai-adk-go/.git`; from the primary checkout the same command prints `.git`, a repository-relative path whose "parent" is not a path at all. A rule phrased on the bare form therefore resolves correctly from every worktree and incorrectly from the one checkout the single-origin design points at. `REQ-KB-005` carries the correct resolution — a single `--path-format=absolute` probe with an older-git fallback — and this SPEC carries none of it.

The predecessor's `column:` SPEC-frontmatter field is **rejected and removed**; the board records the column in its own store, and `REQ-KB-007` forbids the board from writing any SPEC frontmatter field at all.

This SPEC's obligation is narrow and negative: **do not contradict it.** The guidance it prints and the dispatch it sends name a card by its SPEC identifier and nothing else, and no instruction this SPEC emits tells a session to read or write its own tree's copy of board state. A worker session that resolved board state relative to its own working tree would find a different file — or no file — and the board would silently fork per worktree, which is precisely the failure `REQ-KB-005` exists to prevent.

**And it has one origin *and* one writer.** `SPEC-KANBAN-BOARD-001` `REQ-KB-017` makes the `lead` session the sole writer of board state; every other session reads it. That rule is not restated here and is not this SPEC's to restate — but it is this SPEC's to not contradict, and the contradiction is easy to author by accident. A dispatch telling a worker to "mark the card done", or bootstrap guidance handing a worker a board-mutating command, would each install a second writer without anybody deciding to. So: the sessions this SPEC launches, other than the `lead`, are launched as **readers** of the board; the only board write this SPEC's protocol produces is the lead's own, performed after it reads the evidence in a card's `progress.md` (REQ-KS-018).

The division is worth stating plainly because it was lost once. This SPEC **elects** the `lead` role (REQ-KS-004); `REQ-KB-017` says what that role may write. The board sibling's v0.2.0 records how the rule went missing: its predecessor bundled the ownership rule with the rejected `column:` mechanism, the split deleted both, and each SPEC's exclusions disclaimed the write toward the other — this one on the grounds that it decides only who is *told*. Being told about a card is not writing it, which was true and was also the hole.

### A.12 How the lead's backend is refused, and the one channel that cannot be

§A.8 fixes the `lead` session's backend: it is the Claude-backed entry-switch invocation, and it is not operator-selectable. Stated only that way, the property is unfalsifiable — an implementation that never mentions the lead's backend satisfies it exactly as well as one that silently honors an override, and no observation distinguishes them. What makes it checkable is naming what an override **attempt** produces, for every channel a backend can arrive through.

Three channels exist in this SPEC's own surface, and each carries a different refusal:

| Channel | What an override attempt looks like | Required response |
|---|---|---|
| Topology configuration (§A.8, REQ-KS-013) | a backend named for the `lead` role in the configuration section | rejected at load with a named error, alongside the out-of-range session count of REQ-KS-005 — the load fails rather than dropping the key |
| Entry-switch argument | a backend flag passed on the launcher command line | no such parameter exists; REQ-KS-008 forbids expressing any topology parameter as an entry-switch argument, so the token is unrecognized and the parse fails |
| Environment variable (REQ-KS-009) | one of this SPEC's own variables set to a backend value | ignored for the lead's backend, and the guidance still emits the fixed Claude-backed line |

Only the first is a **silent** hazard, because a configuration key that is simply not read looks identical from outside to one that is honored, and both look identical to one that is refused. That is why it is the channel required to fail loudly rather than to be ignored.

**The fourth channel is not refusable, and saying so is better than implying otherwise.** Nothing here stops an operator from ignoring the printed guidance and launching the lead by hand under the GLM launcher. The entry switch prints; it does not supervise what the operator then types. So the property this SPEC can actually hold is *the system offers the operator no path to a non-Claude lead* — not *no non-Claude lead can exist*. The gap is narrow (it requires deliberately departing from emitted guidance) and it is real, and a criterion claiming to close it would be claiming an observation nobody can make.

### A.13 A session declares the role it occupies, and the declaration is not the label

The role table of §A.4 says which roles exist. It does not say how, at runtime, anybody finds out which session is occupying one — and until this revision neither did anything else in the family. Five consumers depend on that answer and each was written as though another supplied it:

| Consumer | What it needs | What it said |
|---|---|---|
| `REQ-KS-019` (this SPEC) | the session whose **declared role** owns a card's column | uses the phrase; defines no declaration |
| `REQ-KS-004` (this SPEC) | — | defines the role *set* and elects `lead`; says nothing about runtime occupancy |
| `REQ-KS-006` (this SPEC, before this revision) | — | "a stable label set at launch", and for *addressability* |
| `SPEC-KANBAN-BOARD-001` `AC-KB-017` | refuse a runtime board write from a non-`lead` session | presupposes the role is readable; `REQ-KB-004` records a **session identifier**, not a role |
| `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007`, `REQ-KW-011` | resolve the session occupying `lead`, or refuse | defers the definition to `REQ-KS-004` |

This is the same shape as the sole-writer loss of §A.11: several consumers, each deferring to a neighbour, and a contract nobody wrote. It is worth naming as a shape rather than as an incident, because the split that produced these three SPECs manufactures exactly this failure at every seam it creates.

**The one derivable candidate does not work, and this SPEC's own text is what rules it out.** The obvious answer is to key on the launch label: it exists, it is stable, and the lead already addresses it. But `REQ-KS-014` requires every emitted command's label to be distinct from every other's, while `AC-KS-013` has each unconfigured worker role emit **two** commands — one per supported backend — and `REQ-KS-005` permits **two** `run` sessions. A role therefore maps to two-or-more possible labels, and which one exists is decided by the operator at the moment they choose a printed command. No label determines a role, and no role determines a label. Reading a role out of a label would mean either collapsing the distinct-label rule or forbidding the operator's backend choice, and both are properties this SPEC holds for independent reasons.

So the declaration is a **second datum**, carried alongside the label rather than derived from it, and REQ-KS-006 now requires both. Three properties make it usable by everything in the table above:

- **It is the routing key.** The lead selects a dispatch target by declared role (REQ-KS-019), not by label, not by idleness.
- **It is the quorum key.** "Which roles answered and which did not" (REQ-KS-012) is an accounting over declarations; counting labels would count an operator's backend choice as a role.
- **It is resolvable from a session that is not the `lead`.** This is the clause the worktree sibling needs and the one an implementation would most plausibly omit: `REQ-KW-007` refuses a disposal when no `lead` is resolvable, and `REQ-KW-011` performs a release only as the `lead` — both evaluated from a session that is not it. A declaration readable only by the lead would satisfy this SPEC's own dispatch and silently break both of the sibling's gates.

What this SPEC does **not** decide is the mechanism — whether the declaration rides the launch command, the session registry, or the peer-discovery output. That is a run-phase choice, and it is left open on purpose: the contract is that a declaration exists, is distinct from the label, and is resolvable by any session. Naming a carrier here would fix a decision this SPEC has no measurement to fix.

---

## §B. Requirements (GEARS)

> Requirement count: 25 (`REQ-KS-001` … `REQ-KS-025`) — at the Tier L ceiling of 25, unchanged at v0.2.0 and again at v0.3.0: every repair in both revisions is an amendment in place and none adds a requirement. The v0.3.0 role-resolution repair in particular was authored as a widening of `REQ-KS-006` for this reason, the ceiling leaving no other shape available; where that constraint had bitten, the finding would have been reported rather than a twenty-sixth requirement written. Acceptance criteria: 30 (`AC-KS-001` … `AC-KS-030`), **exceeding the Tier L ceiling of 25 by five**. The excess is the cost of the three audit findings that could only be closed by adding observations: `AC-KS-026` … `AC-KS-028` are three of the seven observations `AC-KS-024` previously bundled into one line (§D.7), `AC-KS-029` is REQ-KS-013's missing refusal observation (§A.12, §D.8), and `AC-KS-030` is the role-declaration observation REQ-KS-006 gained at v0.3.0 (§A.13, §D.11) — kept separate from `AC-KS-006`'s label check precisely because the repair's whole content is that the two are different data, and one verdict over both would let the declaration half break while the label half passed. Tier L is the top tier, so there is no promotion available and the excess is reported rather than absorbed; re-bundling to fit the ceiling would restore precisely the defects being repaired. Whether to carry the excess, split this SPEC further, or accept a bundled criterion is the orchestrator's decision, not this document's.

### B.1 Prerequisites and preflight

**REQ-KS-001** — The implementation shall write every renamed identifier in its post-`SPEC-KANBAN-RENAME-001` form, and shall introduce no occurrence of `factory` in any identifier, path, environment variable, sentinel, preset name, or prose it authors.

**REQ-KS-002** — The implementer shall verify at preflight that the rename prerequisite, the board sibling's card record, and the cross-session messaging doctrine are all present on the base branch; **when** any of the three is found absent, the implementer shall halt and surface the absence rather than proceeding against an unlanded contract or supplying the missing part itself; and **when** the landed doctrine differs from the version this SPEC was authored against, the implementer shall surface the delta as a blocker before beginning the dispatch-protocol milestone rather than absorbing the change silently.

**REQ-KS-003** — The implementer shall re-measure the five messaging-substrate properties of §A.3 against the run-phase runtime and shall record each result, because a substrate property measured in a prior session is evidence about that session's runtime rather than about this one; and shall likewise re-measure the **launcher surface** the backend decision of §A.8 rests on — that each launcher carries the selected backend into the launched session through the environment it constructs for that session, by whichever launch mechanism the host platform provides, and that each additionally mutates project-global state the other undoes — recording both halves, since a decision resting on an unmeasured premise is a decision resting on nothing (§A.8, `research.md` §J).

### B.2 Session roles and topology

**REQ-KS-004** — The topology shall define exactly five roles — `lead`, `plan`, `run`, `review`, `sync` — with the column ownership of §A.4; the `lead` role shall own no column and perform no card work; the topology shall define no session for `backlog` or `done`, whose not-dispatchable status is the board's (`REQ-KB-012`) and shall not be re-derived here; and this requirement shall be the sole place the `lead` role is elected and assigned, while what that role may write to board state is `REQ-KB-017`'s and shall not be restated, weakened, or duplicated here.

**REQ-KS-005** — The deployed `run`-session count shall default to 1 and shall be raisable to 2 by configuration; **when** a configured value falls outside that range the topology shall reject it at load with a named error rather than clamping it; and the session count shall be held independent of the board's `run` WIP limit, which is owned by `REQ-KB-009` and `REQ-KB-010` and shall not be derived from the session count nor the session count from it.

**REQ-KS-006** — Each session shall be addressable by a stable label set at launch, and the bootstrap guidance shall emit that label as part of an executable launch command for each role; and each session shall additionally **declare the role it occupies**, that declaration being the key on which the lead selects a dispatch target (REQ-KS-019) and over which quorum is accounted (REQ-KS-012), and shall be **resolvable by a session that is not the `lead`**, so that a gate evaluated outside the lead — the disposal refusal of `SPEC-KANBAN-WORKTREE-001` `REQ-KW-007` and the release of `REQ-KW-011`, both of which resolve the occupant of the `lead` role — can reach it. The role declaration shall be a datum **distinct from the launch label** and shall not be derived from it, because a role does not determine a label and a label does not determine a role: an unconfigured worker role emits one command per supported backend (REQ-KS-013) and the `run` role may deploy two sessions (REQ-KS-005), while REQ-KS-014 requires every emitted command's label to differ from every other's, so one role corresponds to two-or-more possible labels chosen by the operator at launch (§A.13). This requirement shall be the sole place the declaration contract is defined, and shall fix no carrier for it — whether the declaration rides the launch command, the session registry, or the peer-discovery output is a run-phase decision.

### B.3 Entry switch, configuration, and the prompt prohibition

**REQ-KS-007** — **Where** a topology configuration exists, the Kanban Mode entry switch shall print bootstrap guidance and wait for quorum; and **where** it is absent, the switch shall fall through to the existing single-session chain with behavior identical to the pre-change launcher.

**REQ-KS-008** — The topology configuration shall live under `.moai/config/` and shall be surfaced through the `moai init` question flow, and no topology parameter shall be expressible as an additional argument on the entry switch.

**REQ-KS-009** — The environment-variable names introduced by this SPEC shall be carried as constants in `internal/config/envkeys.go`, and no call site shall inline either the name or its value as a string literal **in any access direction** — a write (`os.Setenv`, `os.Unsetenv`, a test's `Setenv`) being bound exactly as a read (`os.Getenv`, `os.LookupEnv`) is, because the entry-switch surface this requirement governs accesses the environment predominantly by writing (§D.10).

**REQ-KS-010** — The bootstrap surface shall not call `AskUserQuestion` or any `mcp__askuser__*` tool, and shall express every operator interaction as printed guidance on stderr plus flags and configuration, per the CLI subagent boundary of §A.6.

### B.4 The backward-compatibility baseline

**REQ-KS-011** — The backward-compatibility baseline — the entry switch's parse result, its environment mutation, and its exec argument vector — shall be recorded against the **post-rename** tree, after the rename prerequisite has landed on the base branch and before this SPEC's first edit to the entry switch; a baseline recorded before the rename shall be rejected rather than reconciled, because it encodes retired identifiers and would make every later comparison fail for the rename rather than for a regression; and the no-configuration equivalence of REQ-KS-007 shall be established by mechanical comparison against that recorded baseline rather than by inspection.

### B.5 Bootstrap, quorum, and the emitted command set

**REQ-KS-012** — The quorum wait shall be bounded by a value read from configuration under `.moai/config/`, shall default to 300 seconds, and shall not be expressible as an argument on the entry switch; and **when** the bound elapses without every expected role having answered, the bootstrap shall exit non-zero after naming both the roles that answered and those that did not, and shall neither proceed with a partial team nor continue waiting.

**REQ-KS-013** — The bootstrap guidance shall emit exactly one launch command for the `lead` role, whose backend is the Claude-backed entry-switch invocation and is not operator-selectable; and for each worker role it shall emit exactly one launch command **where** that role's backend is named in the topology configuration, and otherwise exactly two — one per supported backend — so that the operator selects a backend by choosing a printed command rather than by answering a prompt; and **when** a backend for the `lead` role is supplied through the topology configuration, the load shall fail with a named error rather than dropping the key, **when** it is supplied as an entry-switch argument the parse shall fail because no such parameter exists (REQ-KS-008), and **when** it is supplied through an environment variable it shall have no effect on the emitted `lead` line — each channel's response being observable, since a property whose violation produces no observation is a property nothing can check (§A.12).

**REQ-KS-014** — Every command the bootstrap guidance emits shall parse as an executable invocation and shall carry a label distinct from every other emitted command's, and this well-formedness shall be established by parsing the emitted guidance rather than by reading it.

**REQ-KS-015** — The mixed-backend rejection shall be preserved unchanged, including its `KANBAN_MODE_UNSUPPORTED_BACKEND` sentinel; no command the guidance emits shall name the rejected mixed backend; and this SPEC shall add no path by which a mixed-backend session enters the board.

### B.6 Dispatch protocol

**REQ-KS-016** — A dispatch payload shall be a typed record whose fields are a SPEC identifier, a file path, and a contract section reference, and shall carry no free-text field, so that requirement text, acceptance-criterion text, and any restatement of a contract that exists on disk are unrepresentable rather than merely discouraged.

**REQ-KS-017** — A dispatch shall be addressed using the peer's label together with the short reference the discovery tool reports, and **when** a send is refused for want of a reference, the lead shall re-send with the reference the refusal supplies rather than reporting the peer unreachable.

**REQ-KS-018** — The lead shall treat a socket reply as a nudge signal and the card's `progress.md` as the truth, shall advance a card only on evidence read from the shared source of truth, shall not stall on an unreceived reply, and shall reach the same board state whether a reply arrives, arrives late, or never arrives; and no dispatch this SPEC defines shall instruct a worker session to record a card's progression in board state rather than in its own `progress.md`, the write itself being `REQ-KB-017`'s and neither restated nor re-imposed here.

**REQ-KS-019** — The lead shall dispatch a card only to the session whose declared role owns the card's column, shall not dispatch to a session chosen because it is idle, and shall address no session for a card the board reports not dispatchable; and a dispatch shall confer no board-write authority on its recipient, the decision that a card *may* move being the board's (`REQ-KB-008`, `REQ-KB-009`) and the act of writing that move being the `lead`'s alone (`REQ-KB-017`).

**REQ-KS-020** — The lead shall not route an operator decision through a peer session, shall not treat a peer's reply as approval for any gated action, and shall not ask a peer to perform an action the lead's own session is not permitted to perform.

**REQ-KS-021** — A dispatch shall be idempotent with respect to board state, so that a re-sent dispatch for a card already in progress advances nothing and corrupts nothing.

### B.7 The coder session's internal chain

**REQ-KS-022** — Each `run` session shall drive its card through the internal chain the single-session mode already defines, and this SPEC shall introduce no second chaining mechanism: the verify exit gate shall fire at the exit of a card's run-phase with its severity partition, its rung attribute, and its ceiling of at most two re-entries carried over unchanged; the revision dedup predicate shall be evaluated per card without modification to its inputs or its matching rule; and the renamed goal preset shall drive one coder session over one card with its arming rules carried over unchanged. The three "unchanged" claims shall be established against a **durable pre-change artifact** — the gate's severity partition, rung attribute and re-entry ceiling, the predicate's inputs and matching rule, and the preset's arming rules, each recorded before this SPEC's first edit to the chain and readable at verification time — because an equivalence claim whose baseline was never written down is unfalsifiable once the original is gone, and can then be satisfied only by assertion.

**REQ-KS-023** — This SPEC shall change neither the number, the ordering, nor the semantics of the human gates the chain already carries, and shall add no gate of its own; and the gate inventory shall be **recorded before the first edit** — each gate's identity, its position in the ordering, and what it gates — as a durable artifact the post-change enumeration is compared against, this being the same recording obligation REQ-KS-011 carries for the entry switch, applied to the one claim whose silent violation ships unreviewed work.

### B.8 Mirror, neutrality, and verification

> **Disclosed compression.** `REQ-KS-024` absorbs three separable requirements of the superseded predecessor: `REQ-KM-037` (template source before its local counterpart, `make build`, commit the regenerated catalog), `REQ-KM-038` (a mirrored pair preserves its measured relationship) and `REQ-KM-039` (no SPEC identifier, REQ or AC token, internal date, or commit SHA under `internal/template/templates/`). They are kept in one requirement because all three bind the same milestone and the same file set; they are disclosed here because an auditor reconciling 25 requirements against the predecessor's 59 otherwise cannot see that three went in at this slot — and because a bundled requirement is exactly what the board sibling's v0.2.0 records losing half of. The compression is confined to the requirement: `AC-KS-024` and `AC-KS-026` … `AC-KS-028` decide the three obligations separately (§D.7).

**REQ-KS-024** — The implementer shall edit template source under `internal/template/templates/` before its local counterpart, shall run `make build`, and shall commit the regenerated `internal/template/catalog.yaml`; **while** applying a change to a mirrored pair, shall preserve that pair's measured relationship, so that a pair measured byte-identical remains byte-identical and a sanitized pair retains exactly the content its template side strips — a sanitized pair becoming byte-identical being a failure rather than a convergence; shall record each touched pair's classification and its pre-change `diff` as a durable artifact **before** editing that pair, since a comparison against a relationship nobody wrote down cannot fail; and no file authored or modified under `internal/template/templates/` shall contain a SPEC identifier, a REQ or AC token, an internal date, or a commit SHA.

**REQ-KS-025** — The verification shall run the full test suite rather than an affected-packages subset, because a prior run-phase in this repository missed a cross-cutting template guard by testing narrowly.

---

## §C. Exclusions

### Out of Scope — the board sibling

- The column enumeration, the card record's shape, the board state store and its single-origin resolution rule, WIP admission into `run`, the column↔status compatibility check, and the unheld state. All belong to `SPEC-KANBAN-BOARD-001` and are consumed here by requirement id (`REQ-KB-003`, `REQ-KB-004`, `REQ-KB-005`, `REQ-KB-008`, `REQ-KB-009`, `REQ-KB-011`, `REQ-KB-012`), never redefined.
- Deciding whether a card *may* move between columns, and **writing the move once it is decided**. The admission decision is the board's (`REQ-KB-008`, `REQ-KB-009`); the write is the `lead`'s and only the `lead`'s, under `REQ-KB-017`, whose atomicity (`REQ-KB-018`) and board-wide lock (`REQ-KB-019`) are likewise the board sibling's. This SPEC elects the role that holds that authority (REQ-KS-004) and defines the message that prompts a move; it defines no write path and restates none of REQ-KB-017.

  This replaces the v0.1.0 wording, which said only that "this SPEC decides who is told about a card and by what message". That sentence was true and was also the disclaimer half of a hole: the board sibling's own v0.1.0 exclusions said it named no actor, so with one SPEC claiming only notification and the other claiming no actor, the write was owned by neither and the ownership rule was deleted in the split. Being *told* about a card is not writing it — which is why the deferral now has to be explicit rather than implied by silence.
- **The mechanism by which an operator admits a card from `backlog` into `plan`.** §A.4 states that the admission is an operator act rather than a dispatch — `backlog` has no owning session, so no completion report exists for the lead to act on — and that is the whole of this SPEC's claim. How the operator performs it (a CLI verb, a hand edit, an import from an issue tracker) is owned by no SPEC in this family today; the gap is recorded as an out-of-band note in `plan.md` §B.4 so it is visible rather than assumed. This SPEC's dispatch loop begins at the first dispatchable column and is complete without it.
- Any reinstatement of the predecessor's `column:` SPEC-frontmatter field. It is rejected, and `REQ-KB-007` forbids the board from writing SPEC frontmatter at all. This SPEC's only obligation toward board state is the negative one of §A.11.

### Out of Scope — the worktree sibling

- Worktree creation, naming, per-card scope, disposal gating, and refused-removal handling. All belong to `SPEC-KANBAN-WORKTREE-001` (`REQ-KW-003`, `REQ-KW-005`, `REQ-KW-007`, `REQ-KW-008`).
- Stall detection, holder release, orphan classification, and the holder-mutation lock (`REQ-KW-009` through `REQ-KW-014`). This SPEC dispatches to a session; whether a session is judged dead, and whether its card may be re-dispatched, is decided there.
- The isolated tree each `run` session works in. REQ-KS-004 names the role; `REQ-KW-003` and `REQ-KW-005` provide the tree.

### Out of Scope — recovering a role vacated after bootstrap

- **Restoring a role whose session dies once bootstrap has completed.** Quorum in this SPEC is scoped to bootstrap and only to bootstrap: REQ-KS-007 waits for it at the entry switch and REQ-KS-012 bounds that wait, and neither is evaluated again afterwards. Nothing in this SPEC — and, measured, nothing in either sibling — re-establishes a role that falls vacant later.

  **What is unowned.** A worker session that dies mid-run leaves its role occupied by nobody. The board is unaffected and correct; the card is recoverable by the worktree sibling, whose `REQ-KW-011` releases the holder and whose `REQ-KW-012` makes a clean orphan "immediately re-dispatchable". But re-dispatch runs through REQ-KS-019, which addresses **the session whose declared role owns the card's column** — and that role now has no occupant, so the card is dispatched at a vacancy. What is missing is not a repair of those two requirements; it is the observation that a role has become vacant, and a decision about what follows from it.

  **Why it is not owned here.** This SPEC is at the Tier L requirement ceiling of 25, measured at the time of writing — `grep -cE '^\*\*REQ-KS-[0-9]{3}\*\*' spec.md` returns 25 — and every repair in this revision and the two before it is an amendment in place for that reason. Owning this would need a new obligation, not a widening of an existing one: it is a runtime lifecycle property (detect a vacancy, then either re-quorum, refuse dispatch, or surface it) and no requirement here is the natural host for it. The ceiling is reported rather than absorbed, exactly as the criteria overflow is (§B).

  **What a reader who hits this should do.** The symptom is precisely the one §A.5 argues about, arriving after bootstrap instead of at it: a card sits in a column, nothing errors, and the board presents as a **hang rather than as a fault**. Do not read it as a board defect or a dispatch bug — check first whether the column's owning role still has a live session, by resolving the role declaration of REQ-KS-006 across the launched set. The cheap operator recovery available today is the one §A.5 already prescribes for a failed bootstrap: launch the missing session and re-run the entry switch, which re-establishes quorum over the full role set. That is a workaround and not a design; the family's next revision should decide the owner, and `plan.md` §B.6 records the decision it faces.

  This entry exists because both siblings hand quorum to this SPEC by name — `SPEC-KANBAN-BOARD-001` §C and `SPEC-KANBAN-WORKTREE-001` §C each list "the quorum bound" among what belongs here — while this SPEC has only ever scoped it to bootstrap. Silence on this side would complete exactly the pattern §A.13 names: several consumers, each deferring to a neighbour, and a contract nobody wrote.

### Out of Scope — spawning and transport

- Any mechanism by which one session launches another. Sockets cannot launch sessions; bootstrap is manual by necessity (§A.5), and no automation of it is designed, wished for, or deferred with intent.
- Replacing tmux with the messaging layer. The two solve different problems.
- Any change to the messaging channel itself — its socket layout, its discovery tool, its inbound controls, or its configuration keys. This SPEC is a consumer of the channel, never its author.

### Out of Scope — the mixed backend

- Admitting the mixed backend to the board, or softening its rejection into a warning, a fallback, or an adaptation. The rejection and its sentinel are preserved verbatim (§A.8, REQ-KS-015). Its exclusion is a preserved boundary, not an unfilled gap.
- Admitting a third backend to the printed guidance.
- Any interactive backend prompt. The CLI may not ask (§A.6); the choice is expressed as printed alternatives or as configuration (REQ-KS-013).

### Out of Scope — the rename itself

- Every mapping in `SPEC-KANBAN-RENAME-001`. This SPEC consumes the renamed surface and re-litigates none of it. Should the rename not have landed, this SPEC halts (REQ-KS-002) rather than performing the rename itself.

### Out of Scope — the coder chain's internals

- Redefining the verify exit gate's severity partition, its rung attribute, or its re-entry ceiling. All three are carried over unchanged (REQ-KS-022); this SPEC relocates where the gate fires, not what it does.
- Redefining the revision dedup predicate's inputs or its matching rule.
- Re-litigating the goal preset's arming rules, adding a second preset, or promoting the preset to a Go symbol or a CLI verb. It has none today and gains none here.
- Adding, removing, reordering, or reinterpreting any human gate in the chain (REQ-KS-023). A board that quietly dropped a gate would be a board that ships unreviewed work faster.

### Out of Scope — the repaired citation

- Repairing the stale `TestNew_NoAskUserQuestion` reference in `internal/cli/CLAUDE.md`. It is a documentation defect this SPEC surfaced but did not create; it is recorded in `plan.md` §B as an out-of-band note so the dead name is not re-propagated, and it is not made a requirement of this SPEC.

### Out of Scope — the board as a product surface

- A web view, a TUI, a live dashboard, or any rendering of the board beyond what the lead reads and what the CLI prints. The read-only web board is a separate line of work.

---

## §D. Verification surfaces

### D.1 The CLI boundary is anchored on a guard that exists (REQ-KS-010)

The static-guard pattern is real and widely deployed — 29 `func Test…NoAskUserQuestion` definitions were measured across the tree — but the *instance* the predecessor cited is a ghost (§A.6). The criterion therefore names a guard whose definition was verified present, and the run-phase **re-verifies the anchor before citing it**: a `grep` for the chosen function's `func` definition must return exactly one hit before the new guard is written in its image. Citing a guard by name without confirming its definition is what produced the dead reference in the first place, and the same rot can recur.

The new guard itself is an absence claim over the bootstrap source, so it is paired with a **positive control** — a deliberately introduced reference that the same scan reports — establishing that the scan can fire at all.

### D.2 The backward-compatibility baseline is judged on its provenance, not only its content (REQ-KS-011)

Two observations, and the second is the one the predecessor omitted. First, the comparison: parse result, environment mutation, and exec argument vector all equal the recorded baseline with no topology configuration present, checked table-driven rather than asserted in prose. Second, the **provenance**: the baseline artifact must have been recorded against a tree in which the rename has already landed, evidenced by the baseline itself containing the post-rename identifiers and no retired one. A baseline that passes the first check while failing the second is a baseline recorded too early, and its passing comparison is an accident of the entry switch not yet having been touched.

### D.3 The pointer-only dispatch is checked structurally (REQ-KS-016)

A grep of the payload for requirement tokens passes on a payload that paraphrases, which is the failure worth catching. The criterion is the **shape of the type**: the payload record's field set is exactly the identifier, the path, and the section reference, and it carries no field of free-text type. The absence of a free-text field is checked as an absence, with a positive control — a deliberately added free-text field that the same check reports.

### D.4 The emitted command set is parsed, never read (REQ-KS-013, REQ-KS-014)

Reading the guidance and agreeing that it looks right is not a check. The emitted text is parsed into a command set and the set is asserted against four properties: the lead contributes exactly one line of the fixed form; each worker role contributes one line where its backend is configured and exactly two — one per supported backend — where it is not; every line carries a label distinct from every other's; and no line names the rejected mixed backend. The last is an absence claim and carries a positive control.

### D.5 Quorum expiry needs its failing row (REQ-KS-012)

Asserting only the happy path — every role answers, bootstrap proceeds — leaves the requirement's whole point untested, because an implementation that never times out passes it perfectly. The load-bearing row is the expiry: with one role deliberately absent and the bound set short, the bootstrap must exit non-zero, must name the absent role and the present ones separately, and must not have dispatched anything. The pairing with the happy-path row is what establishes that the bound reads configuration rather than simply always firing.

### D.6 Dispatch idempotency and reply-independence are checked as invariance (REQ-KS-018, REQ-KS-021)

Both are claims that a board state does *not* change, so both are checked by comparing board state before and after: a re-sent dispatch for a card already in progress leaves it byte-unchanged, and the same card reaches the same state across three runs in which the reply arrives, arrives after a delay, and never arrives. A criterion asserting only that "no error occurred" would pass against an implementation that advanced the card twice.

### D.7 Mirror delta preservation and neutrality are four checks, not one (REQ-KS-024)

REQ-KS-024 carries three absorbed obligations (§B.8), and at v0.1.0 a single criterion reported one verdict for **seven** distinct observations: the pre-versus-post `diff` equality, the byte-identical pair's invariance, the sanitized pair's retained content, the neutrality guard's exit code, the CI workflow's exit code, the catalog's regeneration and commit, and the template-content scan. Six of the seven could be broken while the criterion still passed, because a reader reporting the aggregate has no obligation to have looked at each part. They are separated into four criteria along the lines that actually fail independently.

**Delta preservation** (`AC-KS-024`) is one property evaluated per pair, so its two class-specific outcomes — a byte-identical pair staying byte-identical, a sanitized pair retaining exactly what its template side strips — stay in one criterion as explicit table rows rather than becoming two criteria of a property that is not two properties. What it needs instead is a **recorded** pre-change `diff`: comparing against a relationship that exists only in a reviewer's memory is a comparison that cannot fail. Pair classification is **time-varying** and is re-measured at run-phase rather than trusted from this document.

**Neutrality** (`AC-KS-026`) rests on two independent mechanical authorities — `internal/template/internal_content_leak_test.go` and `.github/workflows/template-neutrality-check.yaml` — whose exit codes are the verdict, and both are recorded separately because one passing says nothing about the other. This SPEC adds one directed check but does not reimplement the guard's regex; a hand-rolled reimplementation without the guard's exemption list is a false-failure machine.

**The catalog** (`AC-KS-027`) and **the template-content scan** (`AC-KS-028`) are each separable from everything above and from each other: a committed-but-stale catalog and a leaked SPEC identifier are different defects, found by different commands, and neither is implied by a green neutrality run.

### D.8 The lead's fixed backend needs an override attempt, not a silence (REQ-KS-013)

An implementation that never mentions the lead's backend satisfies "not operator-selectable" as thoroughly as one that honors an override — so the criterion is an **attempt** per channel, with the response each must produce (§A.12): a configuration naming a `lead` backend fails the load with a named error, an entry-switch backend argument fails to parse because no such parameter exists, and an environment variable carrying a backend leaves the emitted `lead` line unchanged. The configured channel is the one that must fail *loudly*, because a key silently ignored and a key silently honored are indistinguishable from outside.

The criterion stops where the system's authority stops. An operator who departs from the emitted guidance and launches the lead by hand under the other launcher is not refused by anything here, and no row of the criterion claims otherwise; what is checked is that the system offers no such path, not that no such session can exist.

### D.9 An equivalence claim is judged against a recorded artifact (REQ-KS-022, REQ-KS-023, REQ-KS-024)

Three criteria previously asked for something to be "identical to before" while nothing obliged anyone to capture what before *was*. That is not a weak check, it is an unfalsifiable one: at verification time the original is gone, so the only available evidence is the verifier's own recollection, and a criterion satisfiable by assertion is satisfied by every implementation including the broken one.

REQ-KS-011 already had the right shape — record the baseline, then compare — and it was applied to exactly one of the four equivalence claims this SPEC makes. The other three now carry it: the human-gate inventory before the first edit (REQ-KS-023), the coder chain's three carried-over behaviors (REQ-KS-022), and each touched mirror pair's classification and pre-change `diff` (REQ-KS-024). Each artifact must be **durable and re-readable at verification time** — a file a verifier can open — rather than a value held in working memory during the change, since the whole failure being closed is the disappearance of the comparand.

Both halves are then judged, as REQ-KS-011's already are (§D.2): the comparison itself, and the artifact's provenance — that it was recorded before the edit rather than after it. A baseline taken afterward agrees with the change it was meant to detect, and its passing comparison is a tautology.

### D.10 The env-constant rule is checked in the direction the code actually accesses (REQ-KS-009)

A scan for `os.Getenv("` over this SPEC's surface is not a weak check of REQ-KS-009 — on the measured precedent it is a **vacuous** one. `internal/cli/factory.go`, the existing entry-switch mutation this SPEC extends, accesses the environment at five call sites: `os.Setenv` at lines 93, 95 and 110, `os.LookupEnv` at 107, and `os.Unsetenv` at 113. `grep -c 'os.Getenv('` over that file returns **0**. A read-only scan therefore cannot fire against that shape of code even when the violation is deliberate, so it would report a clean result on a surface it never examined — and a violation introduced as `os.Setenv("MOAI_KANBAN", …)` would pass it untouched.

The criterion enumerates every access direction — the two reads (`os.Getenv`, `os.LookupEnv`) and the three writes (`os.Setenv`, `os.Unsetenv`, a test's `Setenv`) — and asserts of each call site that its key argument is a constant identifier from `internal/config/envkeys.go` rather than a string literal. Reads are in scope alongside writes because REQ-KS-009 binds *every* call site, not only the mutating ones; the reason the read-only form was insufficient is that it was **incomplete**, not that reads are exempt. A later reader tempted to "fix" this by narrowing it back to `os.Getenv` should read this paragraph first.

The check is paired with a positive control in each direction, because a scan that has never been shown to fire is the failure this section exists to describe.

### D.11 The role declaration is verified from outside the lead (REQ-KS-006)

The declaration has three properties (§A.13) and only one of them is checked by the obvious test. That a launched session declares a role, and that the lead reads it, is what an implementer would naturally write a test for — and an implementation passing exactly that test can still hold the declaration in a structure only the lead can reach, which satisfies this SPEC's dispatch and silently breaks `REQ-KW-007` and `REQ-KW-011`, both of which resolve the `lead` occupant from a session that is not the lead. So the criterion resolves the declaration **twice**: once from the lead, and once from a non-`lead` session, and both must succeed.

The second property is the label/declaration distinction, and it is checked as a **non-derivation** rather than as a statement. The load-bearing observation is the case that makes derivation impossible in the first place: an unconfigured worker role emitting two commands, only one of which the operator runs. Whichever of the two labels the launched session carries, its declared role is the same — so the check exercises both choices and asserts the declaration is invariant across them while the label differs. A criterion that only asserted "a declaration field exists" would pass against an implementation that populated it by parsing the label, which is the design §A.13 rules out.

The third is quorum accounting, and the division of labour is between the two criteria rather than a deferral by this one. `AC-KS-012` decides REQ-KS-012's subject — that the bound reads configuration, that expiry exits non-zero naming both groups, that nothing was dispatched — and it carries no declaration-related clause at all, because the accounting **key** is REQ-KS-006's and not REQ-KS-012's. So the key is checked here, as the fourth conjunct of `AC-KS-030`: the lead's accounting is over declared roles, and a role is reported answered only where a session declares it. Through v0.3.0 this paragraph said the property was "not re-checked here" and pointed at REQ-KS-012's criterion for it — a disclaimer that was wrong twice over, since the conjunct performing the check was already three lines above it in `AC-KS-030` and the criterion it pointed at was never going to grow one. What §D.5's expiry row rests on is exactly that conjunct: an implementation counting labels would report a role as answered on the strength of an operator's backend choice, and it is `AC-KS-030` that would catch it.

---

## §E. Cross-references

- `SPEC-KANBAN-RENAME-001` — the prerequisite rename. A `dependencies:` entry and a blocking preflight gate (REQ-KS-002); also the reason the backward-compatibility baseline has a recording window (§A.7).
- `SPEC-KANBAN-BOARD-001` — the prerequisite board model. `REQ-KB-003` (columns), `REQ-KB-004` (card record), `REQ-KB-005` (single-origin state), `REQ-KB-007` (writes no SPEC frontmatter), `REQ-KB-008` (column↔status), `REQ-KB-009`/`REQ-KB-010` (WIP limit, held independent), `REQ-KB-011` (unheld state), `REQ-KB-012` (`backlog`/`done` not dispatchable) are consumed here and defined there.
- `SPEC-KANBAN-BOARD-001` `REQ-KB-017` (the `lead` is the sole writer of board state), `REQ-KB-018` (atomic write) and `REQ-KB-019` (board-wide lock) — restored in that sibling at its v0.2.0 after the three-way split deleted the ownership rule along with the rejected `column:` mechanism it had been bundled with. This SPEC **elects** the `lead` role (REQ-KS-004) and defers write authority to `REQ-KB-017` at §A.4, §A.11, §C, REQ-KS-018 and REQ-KS-019; it defines no write path and duplicates none of those three requirements' content.
- `SPEC-KANBAN-WORKTREE-001` — the prerequisite worktree lifecycle. It provides the isolated tree each `run` session works in and decides whether a released card may be re-dispatched. Its `REQ-KW-007` (disposal refused where no `lead` is resolvable) and `REQ-KW-011` (release performed as the `lead`) are the two consumers that require the role declaration of REQ-KS-006 to be resolvable from a session that is **not** the `lead` (§A.13); both defer the definition to `REQ-KS-004`, which elects the role without defining runtime occupancy, so REQ-KS-006 is where they are actually served. Its `REQ-KW-003` owns per-card worktree naming, and the `worker-` prefix is unavailable to it (§A.8, `plan.md` §B.5).
- `SPEC-KANBAN-MULTISESSION-001` — the superseded 59-requirement predecessor. Its §A.2, §A.3, §A.5–A.9, §A.12, §B.1, §B.3–B.6 and §D.1–D.3 are this SPEC's primary material; three of its positions are corrected and each correction is argued at the section named in HISTORY. One compression is disclosed: `REQ-KS-024` absorbs its `REQ-KM-037`, `REQ-KM-038` and `REQ-KM-039` (§B.8). Its `REQ-KM-040` is carried one-to-one as `REQ-KS-025` and its `REQ-KM-001` as `REQ-KS-001`, neither being a compression.
- `SPEC-FACTORY-MODE-001` — the closed SPEC that delivered the single-session chain being extended. Preserved as a historical record; not amended.
- `.claude/rules/moai/workflow/cross-session-messaging.md` — the dispatch contract, and specifically its role-boundary-dispatch clause that REQ-KS-016 and REQ-KS-019 specialize. Lands ahead of this SPEC as its own pull request; verified present at preflight (§A.2). Deliberately **not** in `dependencies:` — a rule file is not a SPEC.
- `internal/cli/CLAUDE.md` § Conventions — the CLI subagent boundary of §A.6, and the location of the stale guard citation recorded in `plan.md` §B.
- `internal/cli/mcp_boundary_test.go` — `TestMCP_NoAskUserQuestion`, one of the live guard definitions verified present and the pattern REQ-KS-010's new guard extends.
- `internal/cli/cc.go` and `internal/cli/glm.go` — the two entry-switch call sites whose parse result, environment mutation, and exec argument vector the REQ-KS-011 baseline records. `internal/cli/glm.go` additionally carries `setGLMEnv`, which writes the selected backend into the **process** environment, and `persistTeamMode`, which writes `team_mode` into `.moai/config/sections/llm.yaml` — the two halves of §A.8's measured premise, one per-session and one project-global.
- `internal/cli/launcher.go` — `applyCCMode` and `applyGLMMode`, the mode application the printed guidance's commands run, and `launchClaudeDefault`, which at line 791 calls `execOrSpawnClaude` with the constructed environment: the point at which a per-session backend survives into the launched session. Their project-global mutations — tmux session environment, `.claude/settings.local.json`, `team_mode`, and `worker-`-prefixed worktree removal — are the measured reason §A.8 states a consequence rather than assuming per-session isolation, and the reason REQ-KS-003 re-measures the launcher surface (`research.md` §J).
- `internal/cli/launch_exec_posix.go` and `internal/cli/launch_exec_windows.go` — the two build-tagged definitions of `execOrSpawnClaude`, and the measured reason §A.8 states its premise as an environment guarantee rather than as a system call. The POSIX file (`//go:build !windows`) calls `syscall.Exec(claudeBin, args, env)`; the Windows file (`//go:build windows`) spawns a child with `exec.Command`, assigns `child.Env = env`, and propagates the exit code, its comment recording that `syscall.Exec` returns `syscall.EWINDOWS` on that platform. The conclusion §A.8 needs holds on both; the mechanism does not.
- `internal/config/envkeys.go` — the constants file REQ-KS-009 requires the new environment-variable names to live in; `EnvMoaiFactory` (line 142) and `EnvMoaiFactorySpec` (line 147) are the existing pair the rename retires and the shape the new constants follow.
- `internal/cli/factory.go` — the existing entry-switch environment mutation. Its five env-access call sites — `os.Setenv` at 93, 95 and 110, `os.LookupEnv` at 107, `os.Unsetenv` at 113, with **zero** occurrences of `os.Getenv(` — are the measured reason §D.10 checks REQ-KS-009 in every access direction rather than in the read direction alone.
- `.claude/rules/moai/workflow/worktree-integration.md` — the L2 worktree scheme the sibling automates; named here only because the bootstrap guidance references the tree each session enters.
- `CLAUDE.local.md` §2 (Template-First), §14 (env constants), §25 (Template Internal-Content Isolation).
