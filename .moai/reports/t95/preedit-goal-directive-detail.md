---
description: "Detail companion for goal-directive.md — condition templates, comparing-approaches table, integration notes, Native /goal Prohibition rationale"
paths: "**/.moai/state/goal/**,**/.claude/skills/moai/workflows/goal.md,**/goal-directive.md,**/goal-directive-detail.md"
---

# Goal Directive — Detail Companion

> This is the detail companion of `goal-directive.md`. The always-loaded stub owns: the `/moai goal` semantics, the Goal-Presentation Timing arm-only invariant, the Hard Preconditions, and the T1-T4 trigger one-liners. This file owns the expanded detail: the Comparing Approaches table, the condition-authoring guide, the T1-T4 condition templates, the MoAI Integration Notes, and the Native `/goal` Prohibition rationale. Load this file when you are actively arming a goal, choosing between `/moai goal` and `/moai loop`, or need the native-prohibition grounding.

## Comparing Autonomous-Continuation Approaches

Three approaches keep the session running between prompts. Pick by **what should start the next turn**:

| Approach | Next turn starts when | Stops when |
|----------|----------------------|------------|
| `/moai goal` | The previous turn finishes and the `stop-goal` evaluator finds the condition unmet | The conditions hold, the turn ceiling is reached, the stagnation guard fires, or `/moai goal clear` is issued |
| `/loop` (Claude Code native) | A fixed time interval elapses (re-runs the prompt/command on a schedule) | The user cancels the loop |
| `/moai loop` (goal preset — project-wide sweep) | A diagnostic scan builds a finite issue queue; the goal engine (`stop-goal`) then evaluates "queue drained + diagnostics clean" each turn-end. It is a **goal preset** distinct from `/moai goal` — a preset that pre-fills the condition rather than asking the user to author it. | The queue drains + diagnostics clean, or the iteration ceiling is reached |
| Stop hook (`type: prompt` / `type: agent`) | The previous turn finishes | The hook's own script or model decides |

> Note: the Claude Code native `/loop` (time-interval scheduler) and MoAI's `/moai loop` (diagnostic-driven Ralph Engine) are distinct commands — native `/loop` re-runs a prompt on a wall-clock interval, while `/moai loop` iterates on tooling-detected work. They are not interchangeable.

> **Sanctioned composition surface**: `.claude/rules/moai/workflow/cadence-bridge.md` defines the read-only recipe catalog that composes native `/loop` with read-only `/moai` entry points on a schedule — it composes the two commands above without merging their semantics, under a HARD invariant that scheduled runs never commit, never push, and never enter run-phase.

`/moai goal` and `/moai loop` are complementary, not competitors:

- **`/moai loop`** is MoAI's deterministic, diagnostic-driven fix loop — it knows the project's quality tooling and the SPEC lifecycle. Use it for "fix everything the tooling flags".
- **`/moai goal`** is a condition-declared loop over both mechanical checks and transcript claims. Use it for "keep going until this stated end-state is demonstrably true".

## Writing an Effective Condition

A condition is judged by the `stop-goal` evaluator at turn-end: mechanical conditions by running their commands, model conditions against what the session has surfaced in the transcript. Write something that can actually be decided that way. A durable condition usually has:

- **One measurable end state**: a test result, a build exit code, a file count, an empty queue.
- **A stated check**: how the session should prove it (`go test ./... exits 0`, `git status is clean`).
- **Constraints that matter**: what must not change on the way (`no other test file is modified`).

To bound the run, include a turn clause (`or stop after 20 turns`) — otherwise the default turn ceiling applies. Prefer fast, targeted mechanical commands (`go test -run <pattern>` over the full suite): the evaluator runs them at every turn-end, so a slow command taxes each iteration. Check status with `/moai goal status`, which reports the condition text, the conditions array, turns used against the ceiling, the progress log, and the lifecycle status (`armed` / `satisfied` / `ceiling-exit` / `cleared`). End the loop early with `/moai goal clear`. Running `/clear` starts a fresh session, so a goal armed against the previous session is no longer in effect for it.

## Arming Under Multi-Session Concurrency

The default arm path resolves the session id from a per-project side-channel file (`.moai/state/current-session-id.txt`) that the SessionStart hook overwrites unconditionally on every session launch. That file is **single-session-safe only**: with two or more Claude Code sessions in the same project, the second SessionStart clobbers the first's id, so the file holds only the most-recently-started session's id — which may be a *foreign* session. Arming under a foreign id lands the goal under the wrong session's state file and silently breaks the arm↔eval keying (the authoritative id Claude Code passes to the `stop-goal` hook via stdin differs). When `moai goal arm` detects ≥2 concurrent sessions in the project directory it surfaces a non-empty warning rather than arming silently. The orchestrator SHOULD arm deterministically with the authoritative id — `moai goal "<condition>" --session <authoritative-id>` — where `<authoritative-id>` is the `source_session_id` the SessionStart hook injected into the session's additional context. This sidesteps the side-channel file entirely.

## Trigger condition templates

T1-T3 carry a copy-able condition template; T4 is a decision-route trigger with no template (it routes between `/moai goal` and `/moai loop`, then the condition is authored per § Writing an Effective Condition). Every condition follows the 3-part shape where applicable (one measurable end state + a stated check + constraints that matter — § Writing an Effective Condition frames these as "usually") and carries a turn bound.

- **T1 — Long run-phase / multi-milestone (Tier M/L)**. After Implementation Kickoff Approval and before the first implementation `Agent()` spawn, for a SPEC whose run-phase spans multiple milestones. The run-phase autonomy wiring in `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (`ac_converge`) owns this case; the condition IS the `ac_converge` block there — surface that block verbatim from `run.md`, do not compress or re-author it (compression yields a weaker condition and creates a drift surface).
- **T2 — Migration / refactoring across many call sites**. When the work is a sweeping migration or behavior-preserving refactor whose completion is "every call site compiles and tests pass". Arm once the call-site inventory is enumerated and transcript-visible.
  - Template shape: `/moai goal "every call site compiles && go test ./... exits 0 && git status is clean, or stop after 20 turns"`
- **T3 — TDD cycle / SPEC AC convergence**. During a RED-GREEN-REFACTOR loop or while driving toward all acceptance criteria holding. For SPEC-scoped work this overlaps T1 and the `run.md` `ac_converge` wiring is the SSOT; for non-SPEC TDD work arm a test-suite-shaped condition.
  - Template shape: `/moai goal "the target test suite is green && lint is clean, or stop after 15 turns"`
- **T4 — `/moai loop` alternative, made explicit**. When about to enter `/moai loop` for work better expressed as a verifiable end-state than as "fix what the tooling flags", surface the § Comparing Autonomous-Continuation Approaches table and the two options. The decision axis is what should start the next turn: a tooling diagnostic (`/moai loop`) vs a declared condition (`/moai goal`). Both are orchestrator-invocable, so when the choice materially changes scope it is a genuine user decision and routes through `AskUserQuestion` per the Channel Monopoly; when the scope is unambiguous the orchestrator selects and states its choice.

### Guardrails (dedup against sibling doctrine)

- **`run.md` owns the run-phase `ac_converge` wiring**; this section is the higher-level recommendation guide, not a re-statement of the inline condition. Do not duplicate the hard-coded `ac_converge` condition here.
- **`session-handoff.md` owns the resume-context goal** — the Block 5 rule in § Canonical Format (a `Run:` line MAY carry `/moai goal "<condition>"` where the next SPEC declares a machine-verifiable end-state) plus § Auto-Injected Resume Flow. This section governs the **in-session active-work** context. The two are distinct: handoff = crossing a `/clear` boundary; proactive = within an active session before any boundary.
- **Graceful degradation**: per `run.md` § Autonomy invariants — when the goal engine is unavailable (hooks disabled via `disableAllHooks` / `allowManagedHooksOnly`), the workflow degrades to the standard manual per-turn flow. The full enumeration lives in `run.md` and § MoAI Integration Notes; not restated here.

## MoAI Integration Notes

- **Persistence alignment**: `/moai goal` operationalizes MoAI's long-horizon persistence doctrine (`.claude/output-styles/moai/moai.md` § Persistence & Context Awareness) — the orchestrator does not stop early; the goal evaluator decides completion. When a goal is armed, treat the condition itself as the directive and keep working, saving progress to memory as the context window approaches its threshold.
- **`ultrathink.` resume pairing**: a goal condition pairs naturally with a paste-ready resume message (`.claude/rules/moai/workflow/session-handoff.md`). The resume message's `ultrathink.` opener restores reasoning effort; the goal itself rides Block 5, whose `Run:` line MAY carry `/moai goal "<condition>"` where the next SPEC declares a machine-verifiable end-state. Because the directive is orchestrator-invocable, the resumed session's orchestrator arms it directly after Implementation Kickoff Approval — there is no separate follow-up block to send and no reminder-to-type obligation. Arming still does not authorize autonomous run-phase entry; Implementation Kickoff Approval remains required.
- **Auto-injected resume path (`handoff.mode: auto`)**: where the project config `.moai/config/sections/handoff.yaml` sets mode=auto, the previous session's `moai handoff save` record is injected automatically as session-start context at the next `/clear`, so the resumed user sends ONE message. The SSOT for the flow (including the `/clear`-only injection boundary and the injected-precondition verification obligation) is `.claude/rules/moai/workflow/session-handoff.md` § Auto-Injected Resume Flow. Implementation Kickoff Approval remains required in both modes.
- **AskUserQuestion still governs questions**: an armed goal removes per-turn STOP prompts, not the orchestrator's obligation to route genuine user decisions through `AskUserQuestion`. A goal does not authorize bypassing Implementation Kickoff Approval (the plan-to-implement human gate) — if run-phase entry needs user approval, the orchestrator still asks before proceeding. The `stop-goal` hook itself never calls `AskUserQuestion`; in semi-autonomous mode it emits checkpoint JSON and the orchestrator runs the confirm round (`workflows/goal.md` § Semi-autonomous checkpoint flow).
- **Safety boundary**: an armed goal does not relax the "confirm before hard-to-reverse / shared-system actions" boundary. The goal evaluator only decides whether to continue; it does not pre-approve destructive operations.
- **Auto mode pairs with `/moai goal`**: Claude Code's auto mode (per-tool auto-approval) is complementary to the goal loop (per-turn continuation). Together they enable an unattended `ac_converge` run — auto mode removes the per-tool approval prompts while the armed goal removes the per-turn STOP prompts. The Implementation Kickoff Approval plan-to-implement human gate is still required before run-phase entry.
- **Evaluator cost**: the turn-end check runs the goal's mechanical condition commands, so its cost is the cost of those commands. Keep them targeted and deterministic (`workflows/goal.md` § Goal `cmd`s SHOULD be fast); the `stop-goal` Stop-hook timeout is 120s, but a faster command keeps the turn loop tight. Model conditions add no separate provider configuration — they are judged against the transcript the session already produced.
- **Disable scope (per-flag)**: the goal loop is unavailable when hooks are disabled, and the disabling flags differ in scope — `disableAllHooks` turns off hooks at any settings level, while `allowManagedHooksOnly` permits only managed (org-level) hooks. In both cases the goal engine cannot evaluate, and the workflow degrades to the manual per-turn flow.
- **Non-interactive use**: a headless invocation (`claude -p`) can carry an armed goal, so a condition-declared run is usable in CI or a scheduled check. Interrupt with Ctrl+C. Non-interactive surfaces also include the Claude desktop app and Remote Control, not only the headless `-p` CLI.

## Native `/goal` Prohibition

[ZONE:Evolvable] [HARD] The MoAI pipeline emits no native `/goal` line on any surface. This is the single retained home for native-`/goal` references in this rule: the prohibition's rationale, the classification it rests on, and the runtime-interoperation invariant that follows from it.

**Why the pipeline cannot emit it.** Native `/goal` is HUMAN-ONLY — a built-in TUI command the model cannot invoke or set on the user's behalf, and no tool call can trigger it (`.claude/rules/moai/workflow/native-invocation-model.md` § Classification Matrix). A doctrine surface that instructed the orchestrator to emit a native `/goal` line would be instructing it to emit inert text: pasted mid-body it is parsed as plain text, because official slash-command recognition is input-start-only, and the model cannot type a standalone message on the user's behalf either. The instruction therefore has no executable path, which is why the emission surface is unified on `/moai goal` instead.

**Axis B — why `/moai goal` exists.** `/moai goal` is the MoAI-owned, PROGRAMMATIC reimplementation of native `/goal` semantics: it arms a per-session condition-declared loop that the `stop-goal` Stop-hook evaluator checks each turn-end, so the orchestrator can register and arm a goal without a human typing a native command line. This is the Axis B worked illustration — where the nearest native equivalent is HUMAN-ONLY, a MoAI subcommand automating that capability inside the pipeline is the ONLY pipeline path (`native-invocation-model.md` § Axis B).

**What native `/goal` does when a human types it** (facts that remain true, and are not emission instructions). It sets a session-scoped completion condition and Claude keeps working toward it without a prompt each step; after every turn a small fast model (Haiku by default) checks the condition against what Claude surfaced in the conversation, and the goal clears automatically once met. It is a wrapper around a session-scoped prompt-based Stop hook, and requires Claude Code v2.1.139 or later, an accepted workspace trust dialog, and hooks enabled. Bare `/goal` reports the active condition with turns and tokens spent; `/goal clear` ends it early (aliases `stop`, `off`, `reset`, `none`, `cancel`); `/clear` also removes an active goal, and `ultrathink.` does not restore one. A goal active at session end is restored on `--resume` / `--continue`. Non-interactively, `claude -p "/goal <condition>"` runs the loop to completion in a single invocation.

**No transposition of native-only conditions.** The runtime-version floor above (v2.1.139) and the workspace-trust requirement are properties of the **native** command's Stop-hook wrapper, NOT of `moai hook stop-goal`. Do not carry them onto `/moai goal` when rewriting an availability or graceful-degradation clause: `/moai goal` requires hooks enabled and carries no version floor of its own. Transposing the floor would assert a requirement `/moai goal` does not have.

**Runtime interoperation (a safety invariant, not an emission).** When the runtime signals an active native `/goal`, the `stop-goal` evaluator **yields** — it does not block — so a user who typed the native command is not double-blocked by MoAI's evaluator as well. Where the runtime does not expose the signal, the hook degrades to always evaluating the MoAI goal (accepted debt: possible double evaluation, no correctness hazard). This invariant is implemented in the goal-engine evaluator and recorded at `.claude/skills/moai/workflows/goal.md` § Safety Invariants; it is interoperation *with* the native command and must not be removed as a stale reference.

Canonical native documentation: `https://code.claude.com/docs/en/goal`.

## Cross-references

- `.claude/rules/moai/workflow/native-invocation-model.md` § Classification Matrix / § Axis B — the HUMAN-ONLY classification and the automation justification the Native `/goal` Prohibition rests on
- `https://code.claude.com/docs/en/hooks-guide` — prompt-based / agent-based Stop hooks (the mechanism class the goal evaluator belongs to)
- `.claude/output-styles/moai/moai.md` § Persistence & Context Awareness — long-horizon non-stop doctrine
- `.claude/skills/moai/workflow-loop` — `/moai loop` Ralph Engine (deterministic diagnostic loop)
