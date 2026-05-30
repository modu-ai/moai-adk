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
| `/moai loop` (Ralph Engine) | A diagnostic cycle (LSP / AST-grep / test / coverage) finds remaining work | All issues resolved or max iterations reached |
| Stop hook (`type: prompt` / `type: agent`) | The previous turn finishes | The hook's own script or model decides |

`/goal` and `/moai loop` are complementary, not competitors:

- **`/moai loop`** is MoAI's deterministic, diagnostic-driven fix loop — it knows the project's quality tooling and the SPEC lifecycle. Use it for "fix everything the tooling flags".
- **`/goal`** is a model-evaluated condition over the conversation transcript — it does not run commands or read files itself; it judges what Claude has already surfaced. Use it for "keep going until this stated end-state is demonstrably true in the transcript".

## Writing an Effective Condition

The evaluator judges the condition against Claude's own output, so write something Claude's output can demonstrate. A durable condition usually has:

- **One measurable end state**: a test result, a build exit code, a file count, an empty queue.
- **A stated check**: how Claude should prove it (`"go test ./... exits 0"`, `"git status is clean"`).
- **Constraints that matter**: what must not change on the way (`"no other test file is modified"`).

To bound the run, include a turn or time clause (`"or stop after 20 turns"`). The condition can be up to 4,000 characters. Check status with bare `/goal`; clear early with `/goal clear` (aliases: `stop`, `off`, `reset`, `none`, `cancel`). Running `/clear` also removes an active goal. A goal active at session end is restored on `--resume`/`--continue` (turn/timer/token baselines reset).

## MoAI Integration Notes

- **Persistence alignment**: `/goal` operationalizes MoAI's long-horizon persistence doctrine (`.claude/output-styles/moai/moai.md` § Persistence & Context Awareness) — the orchestrator does not stop early; the goal evaluator decides completion. When a goal is active, treat the condition itself as the directive and keep working, saving progress to memory as the context window approaches its threshold.
- **`ultrathink.` resume pairing**: a `/goal` condition pairs naturally with a paste-ready resume message (`.claude/rules/moai/workflow/session-handoff.md`). The resume message's `ultrathink.` opener restores reasoning effort; a re-set `/goal` restores the autonomous-continuation loop after `/clear`.
- **AskUserQuestion still governs questions**: `/goal` removes per-turn STOP prompts, not the orchestrator's obligation to route genuine user decisions through `AskUserQuestion`. A goal does not authorize bypassing GATE-2 (the plan-to-implement human gate) — if run-phase entry needs user approval, the orchestrator still asks before proceeding.
- **Safety boundary**: an active goal does not relax the "confirm before hard-to-reverse / shared-system actions" boundary. The goal evaluator only decides whether to continue; it does not pre-approve destructive operations.
- **Non-interactive use**: `claude -p "/goal <condition>"` runs the loop to completion in a single invocation (useful for CI/scheduled checks). Interrupt with Ctrl+C.

## Cross-references

- https://code.claude.com/docs/en/goal — canonical `/goal` documentation
- https://code.claude.com/docs/en/hooks-guide — prompt-based / agent-based Stop hooks (the mechanism `/goal` wraps)
- `.claude/output-styles/moai/moai.md` § Persistence & Context Awareness — long-horizon non-stop doctrine
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume + `ultrathink.` opener
- `.claude/skills/moai-workflow-loop` — `/moai loop` Ralph Engine (deterministic diagnostic loop)

---

Version: 1.0.0
Classification: Evolvable orchestration guidance — applies to autonomous multi-turn continuation
