# Goal Directive (`/goal`) — Autonomous Continuation

Guidance for the Claude Code `/goal` command — a session-scoped completion condition that keeps Claude working across turns until a fast model confirms the condition holds.

> **Loading scope**: intentionally always-loaded (no `paths:` restriction) — the proactive triggers fire during active work (run-phase / migration / TDD), so this rule must be reachable in those contexts just like `askuser-protocol.md` / `session-handoff.md`. Read when a user sets a `/goal`, when deciding between `/goal`, `/moai loop`, and a Stop hook, or when the orchestrator recognizes a Proactive Recommendation Trigger (§ below).

## What It Is

`/goal <condition>` sets a completion condition and Claude keeps working toward it without a prompt each step. After every turn, a small fast model (Haiku by default) checks whether the condition holds against what Claude has surfaced in the conversation. If not, Claude starts another turn instead of returning control; the goal clears automatically once the condition is met.

`/goal` is a wrapper around a session-scoped prompt-based Stop hook. Requires Claude Code v2.1.139 or later, an accepted workspace trust dialog, and hooks enabled (unavailable when `disableAllHooks` or `allowManagedHooksOnly` is set). Reference: https://code.claude.com/docs/en/goal

## Comparing Autonomous-Continuation Approaches

Three approaches keep the session running between prompts. Pick by **what should start the next turn**:

| Approach | Next turn starts when | Stops when |
|----------|----------------------|------------|
| `/goal` | The previous turn finishes | A fresh model confirms the condition is met |
| `/loop` (Claude Code native) | A fixed time interval elapses (re-runs the prompt/command on a schedule) | The user cancels the loop |
| `/moai loop` (goal preset — project-wide sweep) | A diagnostic scan builds a finite issue queue; the goal engine (`stop-goal`) then evaluates "queue drained + diagnostics clean" each turn-end. It is a **goal preset** distinct from native `/goal` and `/moai goal` — a preset that pre-fills the condition rather than asking the user to author it. | The queue drains + diagnostics clean, or the iteration ceiling is reached |
| Stop hook (`type: prompt` / `type: agent`) | The previous turn finishes | The hook's own script or model decides |

> Note: the Claude Code native `/loop` (time-interval scheduler) and MoAI's `/moai loop` (diagnostic-driven Ralph Engine) are distinct commands — native `/loop` re-runs a prompt on a wall-clock interval, while `/moai loop` iterates on tooling-detected work. They are not interchangeable.

> **Sanctioned composition surface**: `.claude/rules/moai/workflow/cadence-bridge.md` defines the read-only recipe catalog that composes native `/loop` with read-only `/moai` entry points on a schedule — it composes the two commands above without merging their semantics, under a HARD invariant that scheduled runs never commit, never push, and never enter run-phase.

`/goal` and `/moai loop` are complementary, not competitors:

- **`/moai loop`** is MoAI's deterministic, diagnostic-driven fix loop — it knows the project's quality tooling and the SPEC lifecycle. Use it for "fix everything the tooling flags".
- **`/goal`** is a model-evaluated condition over the conversation transcript — it does not run commands or read files itself; it judges what Claude has already surfaced. Use it for "keep going until this stated end-state is demonstrably true in the transcript".

### `/moai goal` — the PROGRAMMATIC MoAI counterpart (Axis B)

Native `/goal` is HUMAN-ONLY (a built-in TUI command the model cannot set on the user's behalf). `/moai goal` is the MoAI-owned, PROGRAMMATIC reimplementation of `/goal` semantics: it arms a per-session condition-declared loop that the `stop-goal` Stop-hook evaluator checks each turn-end, so an orchestrator can register and arm a goal without a human typing the native `/goal` line. This is the Axis B worked illustration — where the nearest native equivalent is HUMAN-ONLY, a MoAI subcommand automating that capability inside the pipeline is the ONLY pipeline path (see `.claude/rules/moai/workflow/native-invocation-model.md` § Axis B). See `.claude/skills/moai/workflows/goal.md` for the four verbs and the autonomous/semi-autonomous progression-mode axis.

## Writing an Effective Condition

The evaluator judges the condition against Claude's own output, so write something Claude's output can demonstrate. A durable condition usually has:

- **One measurable end state**: a test result, a build exit code, a file count, an empty queue.
- **A stated check**: how Claude should prove it (`"go test ./... exits 0"`, `"git status is clean"`).
- **Constraints that matter**: what must not change on the way (`"no other test file is modified"`).

To bound the run, include a turn or time clause (`"or stop after 20 turns"`). The condition can be up to 4,000 characters. Check status with bare `/goal` — which reports the active condition along with the turns and tokens spent so far. While a goal is active a `◎ /goal active` indicator is shown, and after each turn the evaluator surfaces its reason for continuing or stopping. Clear early with `/goal clear` (aliases: `stop`, `off`, `reset`, `none`, `cancel`). Running `/clear` also removes an active goal. A goal active at session end is restored on `--resume`/`--continue` (turn/timer/token baselines reset).

## Proactive Recommendation Triggers

The preceding sections describe `/goal` semantics and condition authoring. This section adds the **proactive layer**: when the orchestrator, mid-workflow, recognizes a situation where `/goal` is the right continuation primitive, it surfaces a recommendation so the user can set one. This closes the gap between the two contexts where `/goal` is already wired (the user setting one unprompted, and the session-handoff resume flow) and the active-workflow context, where previously nothing prompted the orchestrator to recommend `/goal`.

### Hard preconditions for every recommendation

- **`/goal` is user-only (HUMAN-ONLY)**: `/goal` is a TUI command the model cannot invoke or set on the user's behalf (`.claude/rules/moai/workflow/native-invocation-model.md`). The orchestrator **recommends and supplies a copy-able condition template**; the user types the `/goal` line. The delivery model is this file's reminder-obligation pattern — natural-language status guidance, NOT `AskUserQuestion` and NOT a mid-paste slash line. (A `/goal`-turn agent's inability to surface `AskUserQuestion` mid-run is a property of MoAI's run-phase delegation model, NOT of `/goal` itself — `/goal` only removes per-turn STOP prompts; see § MoAI Integration Notes.)
- **Implementation Kickoff Approval comes first**: any run-phase `/goal` recommendation is downstream of the Implementation Kickoff Approval human gate (`AskUserQuestion`, plan→run) and never substitutes for or bypasses it. `run.md` § Run-phase Autonomy #1 owns the preferences-drained rationale (all user preferences are collected at that gate before any autonomy begins).
- **Safety boundary unchanged** (see § MoAI Integration Notes for the full statement).
- **`run.md` "set" shorthand reconciliation**: `run.md` § Run-phase Autonomy uses the shorthand "the orchestrator MAY set the `ac_converge` `/goal`". Per the user-only Hard precondition above, read this as "the orchestrator recommends the condition and the user types the `/goal` line" — the model cannot set `/goal` (HUMAN-ONLY). T1/T3 below defer to `run.md`'s condition under this reading; they do not authorize the model to set `/goal`.

### Triggers and condition templates

T1-T3 carry a copy-able `/goal <condition>` template the user may paste; T4 is a decision-route trigger with no template (it routes between `/goal` and `/moai loop`, then the user authors the condition per § Writing an Effective Condition). Every condition follows the 3-part shape where applicable (one measurable end state + a stated check + constraints that matter — § Writing an Effective Condition frames these as "usually") and carries a turn/time bound.

- **T1 — Long run-phase / multi-milestone (Tier M/L)**. After Implementation Kickoff Approval and before the first implementation `Agent()` spawn, for a SPEC whose run-phase spans multiple milestones. The run-phase autonomy wiring in `.claude/skills/moai/workflows/run.md` § Run-phase Autonomy (`/goal ac_converge`) owns this case; the user-pastable condition IS the `ac_converge` block there — surface that block verbatim from `run.md`, do not compress or re-author it (compression yields a weaker condition and creates a drift surface).
- **T2 — Migration / refactoring across many call sites**. When the work is a sweeping migration or behavior-preserving refactor whose completion is "every call site compiles and tests pass" — the canonical `/goal` use case from the official docs. Recommend `/goal` once the call-site inventory is enumerated and transcript-visible.
  - Template shape: `/goal every call site compiles && go test ./... exits 0 && git status is clean, or stop after 20 turns`
- **T3 — TDD cycle / SPEC AC convergence**. During a RED-GREEN-REFACTOR loop or while driving toward all acceptance criteria holding. For SPEC-scoped work this overlaps T1 and the `run.md` `/goal ac_converge` wiring is the SSOT; for non-SPEC TDD work recommend `/goal` with a test-suite-shaped condition.
  - Template shape: `/goal the target test suite is green && lint is clean, or stop after 15 turns`
- **T4 — `/moai loop` alternative, made explicit**. When about to enter `/moai loop` for work better expressed as a verifiable end-state than as "fix what the tooling flags", surface the § Comparing Autonomous-Continuation Approaches table and recommend `/goal`. The decision axis is what should start the next turn: a tooling diagnostic (`/moai loop`) vs a model-evaluated condition (`/goal`). This routes a decision but does NOT route it through `AskUserQuestion`: `/goal` is HUMAN-ONLY, so the model cannot execute either choice — it recommends in natural language and the user decides by typing or declining (status guidance, per the Channel Monopoly status-statement exception).

### Guardrails (dedup against sibling doctrine)

- **`run.md` owns the run-phase `ac_converge` wiring**; this section is the higher-level recommendation guide, not a re-statement of the inline condition. Do not duplicate the hard-coded `ac_converge` condition here.
- **`session-handoff.md` owns the resume-context `/goal`** (Post-Paste `/goal` Follow-up Block, Auto-Injected Resume Flow); this section governs the **in-session active-work** recommendation context. The two are distinct: handoff = crossing a `/clear` boundary; proactive = within an active session before any boundary.
- **Graceful degradation**: per `run.md` § Autonomy invariants — when `/goal` is unavailable (runtime < v2.1.139, hooks disabled via `disableAllHooks`/`allowManagedHooksOnly`, or workspace trust dialog not accepted), the workflow degrades to the standard manual per-turn flow. The full enumeration lives in `run.md` and § MoAI Integration Notes; not restated here.

## MoAI Integration Notes

- **Persistence alignment**: `/goal` operationalizes MoAI's long-horizon persistence doctrine (`.claude/output-styles/moai/moai.md` § Persistence & Context Awareness) — the orchestrator does not stop early; the goal evaluator decides completion. When a goal is active, treat the condition itself as the directive and keep working, saving progress to memory as the context window approaches its threshold.
- **`ultrathink.` resume pairing**: a `/goal` condition pairs naturally with a paste-ready resume message (`.claude/rules/moai/workflow/session-handoff.md`). The resume message's `ultrathink.` opener restores reasoning effort, but the `/goal` is delivered by a **two-step mechanism**, NOT a line inside the pasted body: the main resume block carries NO `/goal` line (a mid-paste slash line is inert — official slash-command parsing recognizes a `/` command only at input start), and the orchestrator emits a separate **post-paste `/goal` follow-up block** OUTSIDE and AFTER the main block — a standalone-message block containing exactly `/goal <completion-condition>` — emitted ONLY when the next SPEC is run-phase AND has a machine-verifiable completion condition. See `.claude/rules/moai/workflow/session-handoff.md` § Post-Paste /goal Follow-up Block. Because `/clear` removes an active goal and `ultrathink.` does NOT restore it, this follow-up block (sent as its own standalone message) is the mechanism that re-arms the autonomous-continuation loop in the resumed session. The follow-up block does NOT authorize autonomous run-phase entry — Implementation Kickoff Approval remains required.
- **Resumed-session `/goal` reminder obligation**: because the model cannot invoke `/goal` on the user's behalf (it is a user-only TUI command, not a model-invocable tool/skill), the resumed session's orchestrator carries a **reminder obligation** — it MUST remind the user, via natural-language status guidance (NOT `AskUserQuestion`), to send the `/goal` line as a standalone message at the recommended moment (after Implementation Kickoff Approval, since setting a goal starts a turn immediately). Detection is from the handoff auto-memory entry (the follow-up block is persisted verbatim there) or by re-deriving the emission condition (run-phase next SPEC with a machine-verifiable end-state); the pasted main body itself carries no `/goal` reference.
- **Auto-injected goal-first path (`handoff.mode: auto`)**: where the project config `.moai/config/sections/handoff.yaml` sets mode=auto, the previous session's `moai handoff save` record is injected automatically as session-start context at the next `/clear`, and the resumed user's ONE message can be the standalone `/goal <condition>` line itself — the goal-first single-message flow. The two-step post-paste mechanism above remains the mode=manual fallback path. The SSOT for the flow (including the /clear-only injection boundary and the injected-precondition verification obligation) is `.claude/rules/moai/workflow/session-handoff.md` § Auto-Injected Resume Flow. Implementation Kickoff Approval remains required in both modes.
- **AskUserQuestion still governs questions**: `/goal` removes per-turn STOP prompts, not the orchestrator's obligation to route genuine user decisions through `AskUserQuestion`. A goal does not authorize bypassing Implementation Kickoff Approval (the plan-to-implement human gate) — if run-phase entry needs user approval, the orchestrator still asks before proceeding.
- **Safety boundary**: an active goal does not relax the "confirm before hard-to-reverse / shared-system actions" boundary. The goal evaluator only decides whether to continue; it does not pre-approve destructive operations.
- **Auto mode pairs with `/goal`**: Claude Code's auto mode (per-tool auto-approval) is complementary to `/goal` (per-turn continuation). Together they enable an unattended `ac_converge` loop — auto mode removes the per-tool approval prompts while `/goal` removes the per-turn STOP prompts, so an acceptance-criteria-convergence run can proceed without interruption. The Implementation Kickoff Approval plan-to-implement human gate is still required before run-phase entry.
- **Evaluator cost**: the after-turn condition check runs on a small fast model (Haiku by default) and is negligible relative to the main turn. It runs on the session's own provider — including GLM when the session is GLM-backed — so no separate provider configuration is needed.
- **Disable scope (per-flag)**: `/goal` is unavailable when hooks are disabled, but the disabling flags differ in scope — `disableAllHooks` turns off hooks at any settings level, while `allowManagedHooksOnly` permits only managed (org-level) hooks; in both cases the `/goal` command explains why it is unavailable rather than failing silently.
- **Non-interactive use**: `claude -p "/goal <condition>"` runs the loop to completion in a single invocation (useful for CI/scheduled checks). Interrupt with Ctrl+C. Non-interactive surfaces also include the Claude desktop app and Remote Control, not only the headless `-p` CLI.

## Cross-references

- https://code.claude.com/docs/en/goal — canonical `/goal` documentation
- https://code.claude.com/docs/en/hooks-guide — prompt-based / agent-based Stop hooks (the mechanism `/goal` wraps)
- `.claude/output-styles/moai/moai.md` § Persistence & Context Awareness — long-horizon non-stop doctrine
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume + `ultrathink.` opener
- `.claude/skills/moai-workflow-loop` — `/moai loop` Ralph Engine (deterministic diagnostic loop)

---

Version: 1.0.0
Classification: Evolvable orchestration guidance — applies to autonomous multi-turn continuation
