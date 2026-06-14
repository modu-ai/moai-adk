# Goal Directive (`/goal`) — Autonomous Continuation

Guidance for the Claude Code `/goal` command — a session-scoped completion condition that keeps Claude working across turns until a fast model confirms the condition holds.

> **Loading scope**: orchestration-level guidance. Read when a user sets a `/goal`, or when deciding between `/goal`, `/moai loop`, and a Stop hook for autonomous work.

## What It Is

`/goal <condition>` sets a completion condition and Claude keeps working toward it without a prompt each step. After every turn, a small fast model (Haiku by default) checks whether the condition holds against what Claude has surfaced in the conversation. If not, Claude starts another turn instead of returning control; the goal clears automatically once the condition is met.

`/goal` is a wrapper around a session-scoped prompt-based Stop hook. Requires Claude Code v2.1.139 or later, an accepted workspace trust dialog, and hooks enabled (unavailable when `disableAllHooks` or `allowManagedHooksOnly` is set). Reference: https://code.claude.com/docs/en/goal

## Comparing Autonomous-Continuation Approaches

Three approaches keep the session running between prompts. Pick by **what should start the next turn**:

| Approach | Next turn starts when | Stops when |
|----------|----------------------|------------|
| `/goal` | The previous turn finishes | A fresh model confirms the condition is met |
| `/loop` (Claude Code native) | A fixed time interval elapses (re-runs the prompt/command on a schedule) | The user cancels the loop |
| `/moai loop` (Ralph Engine) | A diagnostic cycle (LSP / AST-grep / test / coverage) finds remaining work | All issues resolved or max iterations reached |
| Stop hook (`type: prompt` / `type: agent`) | The previous turn finishes | The hook's own script or model decides |

> Note: the Claude Code native `/loop` (time-interval scheduler) and MoAI's `/moai loop` (diagnostic-driven Ralph Engine) are distinct commands — native `/loop` re-runs a prompt on a wall-clock interval, while `/moai loop` iterates on tooling-detected work. They are not interchangeable.

`/goal` and `/moai loop` are complementary, not competitors:

- **`/moai loop`** is MoAI's deterministic, diagnostic-driven fix loop — it knows the project's quality tooling and the SPEC lifecycle. Use it for "fix everything the tooling flags".
- **`/goal`** is a model-evaluated condition over the conversation transcript — it does not run commands or read files itself; it judges what Claude has already surfaced. Use it for "keep going until this stated end-state is demonstrably true in the transcript".

## Writing an Effective Condition

The evaluator judges the condition against Claude's own output, so write something Claude's output can demonstrate. A durable condition usually has:

- **One measurable end state**: a test result, a build exit code, a file count, an empty queue.
- **A stated check**: how Claude should prove it (`"go test ./... exits 0"`, `"git status is clean"`).
- **Constraints that matter**: what must not change on the way (`"no other test file is modified"`).

To bound the run, include a turn or time clause (`"or stop after 20 turns"`). The condition can be up to 4,000 characters. Check status with bare `/goal` — which reports the active condition along with the turns and tokens spent so far. While a goal is active a `◎ /goal active` indicator is shown, and after each turn the evaluator surfaces its reason for continuing or stopping. Clear early with `/goal clear` (aliases: `stop`, `off`, `reset`, `none`, `cancel`). Running `/clear` also removes an active goal. A goal active at session end is restored on `--resume`/`--continue` (turn/timer/token baselines reset).

## MoAI Integration Notes

- **Persistence alignment**: `/goal` operationalizes MoAI's long-horizon persistence doctrine (`.claude/output-styles/moai/moai.md` § Persistence & Context Awareness) — the orchestrator does not stop early; the goal evaluator decides completion. When a goal is active, treat the condition itself as the directive and keep working, saving progress to memory as the context window approaches its threshold.
- **`ultrathink.` resume pairing**: a `/goal` condition pairs naturally with a paste-ready resume message (`.claude/rules/moai/workflow/session-handoff.md`). The resume message's `ultrathink.` opener restores reasoning effort; a re-set `/goal` restores the autonomous-continuation loop after `/clear`.
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
