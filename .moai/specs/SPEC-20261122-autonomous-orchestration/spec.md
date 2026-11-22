---
id: SPEC-20261122-autonomous-orchestration
title: "Orchestration is one ledger-driven pass that plans, starts, wakes, audits and surfaces through native Claude Code features, with Orca as an optional adapter"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-flow-skeleton
---

# Orchestration is one ledger-driven pass that plans, starts, wakes, audits and surfaces through native Claude Code features, with Orca as an optional adapter

## Intent

User request of 2026-11-22: moai-adk shall have autonomous orchestration built on Claude Code's native features, with Orca only an optional host adapter. The decision in force is that orchestration is code plus the ledger and never a resident leader session. This SPEC settles the shape after reading whole, on 2026-11-22, the official pages `agent-view`, `worktrees`, `scheduled-tasks`, `goal`, `cross-session-messaging`, `remote-control`, `sub-agents` and `tools-reference` (code.claude.com/docs/en; installed Claude Code 2.1.330), and the SPECs that already carry its parts: the flow skeleton, the headless run contract, the session messages, the non-stop lane and the background-session measurement.

Facts from those pages that decide the shape (each with its page; version floors are all below 2.1.330):

- A program can start a lane without a terminal only through `claude --bg` (agent view, research preview): "Start the session in the background and return immediately"; a session started in a terminal needs a person's terminal. The state of every session is read with `claude agents --json`, "the supported way to read session state from outside Claude Code", with `state`, `status` and `waitingFor`; the files under `~/.claude/jobs/` are "not a stable interface".
- There is no command that sends a message to another session. Messaging is a tool the model calls (`SendMessage`, `ListAgents`) and an undocumented socket for scripts, whose line format the page does not give; `notify_when_idle` can be subscribed only by "the Claude in your main conversation" and only for sessions on this machine. So a program cannot wake a live idle session by itself; a session can, and a session that has exited can be continued with `claude --resume <id> --bg "<prompt>"`.
- A message "can't approve anything: a message from another session never counts as your consent, so it can't answer a pending permission prompt on your behalf", and a background session's permission prompt is answered by attaching to it. An orchestrator therefore has no way, and no right, to answer a prompt or a user question.
- A background session "commits without asking, and pushes the branch when the repository has a remote" unless the task, `CLAUDE.md` or memory says otherwise; a lane's worktree is made by moai from the integration branch with git, because `worktree.baseRef` cannot name a branch.
- A `claude -p` run is the audit runner's shape (the headless run contract pins its flags and records its row); a subagent cannot be messaged by name unless agent teams are on, and a named background subagent can become a teammate that runs in the main directory, so the audit runner is a separate `claude -p` process and not a subagent.
- Same-machine messaging needs every session under one Claude config root, and the loop of a session is the only native timer (`/loop` self-paced with a ledger watch, SPEC-20261122-nonstop-lane).

What exists in the repository today and is replaced or reused: the Orca-only `leader-round`, `lane-screen`, `lane-send` and `lane-registry` read terminal screens and type into them (replaced by the host adapter, as the session-messages SPEC decides); `card next`, `card timeline` and `card inbox` read the ledger; `usage_limit` evaluates the usage record; the audit-claim and audit-gate modules hold the author-never-audits rule.

This SPEC is the research-and-SPEC slice: no code. It depends on the measurement report of SPEC-20261122-bg-session-research for the choice of the native start route; until that report exists the native start route is marked trial and the Orca adapter and the stand-in are the tested routes.

## Non-goals

- A resident leader, a daemon or a timer of moai: one pass runs per call and leaves nothing behind.
- Answering, approving or choosing for the user: permission prompts, user questions and decisions are shown, never answered.
- Landing and closing: `moai card land` stays the only way the integration branch moves, and a lane closes its own card.
- The headless runner itself, the cross-session wake text and the non-stop loop prompt: those are the headless run contract, the session messages and the non-stop lane; this SPEC calls them.
- Remote Control, cloud sessions and other machines: Remote Control needs a claude.ai login and is refused for a gateway base URL, so it is out (the session-messages decision D-1 holds).
- Choosing which card a lane takes: `card next` does, and the lane takes the card itself.

## Requirements

- REQ-1: When an orchestration pass runs, the pass shall read the ledger and the host state through one adapter and print the planned actions before it performs any · failure: none — `plan` writes nothing and `run` performs only what the printed plan holds.
- REQ-2: The pass shall run once per call and leave no process, timer, loop or session of its own behind · failure: none — repetition belongs to the caller's loop and a pass started while another runs for the same project waits on the project's lock and then plans from the new state.
- REQ-3: When a ready card is unheld and `card next` names it for no lane and the lane cap is not reached, the pass shall start one lane for it through the adapter and never take the card for the lane · failure: none — a start the adapter cannot perform becomes a `needs_user` action with the command a person runs.
- REQ-4: When a card of a live lane has a verdict, an answer or a landing note newer than the lane's last cursor, the pass shall wake the lane with one message that names the ledger change and carries no state the ledger does not hold · failure: none — a lane whose session is not reachable gets the reason reported once and no terminal input.
- REQ-5: When a wake needs a tool that only a session holds, the pass shall list it as a session action with the exact text and the recipient name for the session that runs the pass to carry out and record as delivered, held or refused · failure: none — an unrecorded session action is planned again at the next pass under the same key.
- REQ-6: When a lane's session has exited and its card is still held, the pass shall continue it in the background with the saved session id and the wake text · failure: none — a session that is running or whose id is unknown is never continued, because continuing a running session starts a copy.
- REQ-7: When a card in review has no verdict for its head, the pass shall dispatch one audit run through the headless run contract with a runner that is not the card's author session · failure: none — a requester that is the author session refuses the dispatch and records the refusal.
- REQ-8: When a permission prompt, a user question or a decision request is waiting, the pass shall list it as an attention item with where it is answered and with no answer, no approval and no key press written for it · failure: none — an item with an unreadable source is listed as unread.
- REQ-9: When the running lanes reach the lane cap or the usage record reads near or reached, the pass shall start no lane and report the reason · failure: none — a missing usage field reads unknown and is never counted as zero.
- REQ-10: When the pass performs or plans an action, the pass shall write one ledger event with a stable key and act at most once on that key · failure: none — a pass run twice on an unchanged state plans nothing the second time.
- REQ-11: When the adapter is `orca`, the pass shall use the Orca commands only for the lanes that Orca hosts and read each lane's state without typing into its screen · failure: none — an Orca that does not answer is reported and the other adapter still runs.
- REQ-12: When any answer, help text or event of the pass is printed or recorded, the text shall hold no card number in an identifier, no key and no path of the real home or real config folder · failure: none — a lane is named by its SPEC id or its registered lane name.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-10): with a stand-in `claude` and a temporary ledger holding one unheld ready card, one `run` pass starts one lane through the stand-in (its recorded argv names the lane and the worktree), writes one `orchestrate` event whose key names the card, takes no reservation, and a second pass on the same state plans nothing · verify: `cargo test --locked --offline --test orchestrate_cli start_`
- AC-2 (REQ-4, REQ-5, REQ-6): a lane whose card got a verdict note after its cursor is planned as one wake whose text names the note and nothing else, a live lane's wake is a session action with the recipient name, and an exited lane with a known session id is continued with `--resume <id> --bg` while a running one is not · verify: `cargo test --locked --offline --test orchestrate_cli wake_`
- AC-3 (REQ-7): a card in review without a verdict gets one audit dispatch through the stand-in runner, an audit requested by the author session refuses, and the dispatch is recorded once per head · verify: `cargo test --locked --offline --test orchestrate_cli audit_`
- AC-4 (REQ-8): with a stand-in listing that shows a session waiting for a permission prompt and a ledger note asking for a decision, the pass lists both as attention items, and the stand-in's record shows no keystroke, no message and no settings write answering either · verify: `cargo test --locked --offline --test orchestrate_cli attention_`
- AC-5 (REQ-9): with the lane cap reached or the usage record near, no lane starts and the reason is printed; with the usage field absent the answer reads unknown · verify: `cargo test --locked --offline --test orchestrate_cli budget_`
- AC-6 (REQ-11): the Orca adapter plans a start and a state read for an Orca-hosted lane through a stand-in `orca` and never writes into a screen, and with no Orca installed the other adapter is unaffected · verify: `cargo test --locked --offline --test orchestrate_cli orca_`
- AC-7 (REQ-12): the help text, the events and the plan of every test carry no card number identifier, no key-shaped text and no real-home path · verify: `cargo test --locked --offline --test orchestrate_cli text_`

## Scope

- write (this slice): `.moai/specs/SPEC-20261122-autonomous-orchestration/spec.md`
- write (code slices, joined by card scope when their turn comes): the pass module and its adapter module under `src/`, `tests/orchestrate_cli.rs` with its stand-in `claude` and `orca`, the `orchestrate` event kind writer, the document row and the skill text of `/moai work`
- read-only: `src/card/**`, `src/usage_limit.rs`, `src/audit_claim.rs`, `src/audit_gate.rs`, the headless run contract and session messages SPECs
- last step, by pair request: `src/main.rs` and `src/help.rs` (the action word) and the skill source manifest
- forbidden: a real `claude` or `orca` process or a model call in tests, any input into a lane's screen, any write of a reservation, a decision note or an audit verdict by the pass itself

## Decisions

- D-1: The command is `moai card orchestrate` with the actions `plan` (read-only), `run` (performs the plan), `done` (records a session action's result) and `attention` (lists what waits for the user). The flow skeleton's D-1 allows no new top-level word, so the card's working name `moai orchestrate` is not used (open: the leader may overrule, reversible).
- D-2: A pass is plan-then-perform with one ledger event per action under a stable key; the key is `<kind>:<card>:<anchor>` where the anchor is the newest note id, the head or the session id the action answers, so a repeated pass finds its own event and does nothing.
- D-3: Whoever runs the pass is an ordinary session: the user's session at `/moai work`, or a lane inside the non-stop loop. No role persists, so there is no resident leader; the lock of the project makes two passes take turns.
- D-4: The adapter has three capabilities, `start`, `state` and `wake`. The native adapter starts with `claude --bg --name <lane> --agent <agent> "<first prompt>"` in a worktree moai created from the integration branch, reads state with `claude agents --json --all`, and wakes through a session action or `--resume <id> --bg`. The route is trial until the measurement report names keep, trial or drop; when it names drop, the native `start` becomes a `needs_user` action that prints `claude --worktree <name> -n <SPEC>` or the deep link for a person to open.
- D-5: The Orca adapter is selected only by the project profile; it starts and reads lanes through `orca terminal create` and `orca terminal read` and never types into a screen (the screen-send accident of 2026-11-19 stays the reason). It is optional: nothing here requires Orca.
- D-6: Waking is a message that points at the ledger and a cursor (session-messages D-4). A live idle lane needs a session action because no command sends a message; the pass prints the exact `SendMessage` text and recipient name, the carrying session sends it and calls `done` with delivered, held or refused. A held message waits for a person by default when sender and receiver differ in bypass, so a held result is reported, never re-sent.
- D-7: Audits go to a `claude -p` runner built by the headless run contract with the audit agent and no write tools; the runner records the requesting session and the pass passes the author session of the card so the runner can refuse it. The orchestrator never writes the verdict line: the runner's typed verdict does.
- D-8: User items are shown, not decided: `attention` lists permission prompts from `waitingFor` of the session list (the answer is `claude attach <id>`), unanswered `kind=decision` notes with their options and default, and a count of held messages. The session that runs the pass shows the decision items to the user as a selection with the recommended option first; the answer is the user's note, never the pass's.
- D-9: Budgets are two numbers: the lane cap from the profile (default 3, and the headless runs count against it) and the usage record through the existing evaluation; a missing `rate_limits` (third-party providers and API-key sessions do not have it) is unknown and does not block, because the cap still holds.
- D-10: Lanes run under the config root of the session the user talks to (session-messages REQ-5), so the pass reads one list and one messaging channel; a lane in another root is reported unreachable.
- D-11: The native adapter's version floors (`claude agents` needs 2.1.139 and names need 2.1.290) are checked once per pass from `claude --version`; a lower version makes the adapter report its reason and the pass uses the other adapter.
- D-12: Nothing here runs on Windows until measured; the SPEC names macOS and Linux and lists Windows as not measured.

## Plan

1. This SPEC, read by the leader and the audit; the start route question waits for the measurement report.
2. Slice 2, tests first (RED): `orchestrate_cli` `start_` and `text_` with the stand-in `claude`; the `plan` action, the event key and the idempotent second pass.
3. Slice 3: `wake_` and `attention_` (session actions, `done`, continuing an exited lane, the attention list).
4. Slice 4: `audit_` and `budget_` on the headless runner once its code landed; the lane cap and the usage evaluation.
5. Slice 5: the Orca adapter (`orca_`) and the removal of the screen-reading leader round, on its own landing and after the others.
6. Per slice: light checks, at most the mutant cap, one affected run, then the READY note with head and base only.

## Risks

- The native start route is a research preview whose interface and keys may change: the adapter is behind a trait, the measurement report dates each row, and a failed start is a `needs_user` action and never a silent loss.
- A wake that needs a session tool depends on the session carrying it out: an unrecorded session action is planned again under the same key, and the ledger, not the message, is the record, so a lost wake costs a delay only.
- A message from a bypassing sender to a prompting receiver is held for a person by default and a `-p` worker drops a held message after five minutes: headless workers start with inbound accept in their settings (session-messages REQ-10) and a held result is shown to the user.
- Continuing a session that is running starts a copy under the page's own rule: the pass continues only an exited session and reads the state again just before.
- The usage record is subscription-only: on a gateway the lane cap is the only budget, and the report says usage is unknown.
- Two passes of two sessions can race: the project lock and the stable keys make the second pass plan from the first one's events.
- A permission prompt that nobody answers stops a lane at `waitingFor`: the attention list is the whole remedy by design, because answering it from another agent is forbidden.
- Idle background sessions stop after about an hour and a stopped session does not auto-resume: the pass reads the state at each call and continues only what the ledger still holds.
