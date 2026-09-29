# M1 measurement caps — SPEC-SESSION-DOUBLELOAD-001 (card t1279)

Committed alone, before any probe output exists (REQ-SDL-004; plan §D). Raising either
value-bearing field below is a plan change, not a probe-time choice.

caps: probes<=6 turns<=2 model=haiku
timeout_s: 300
wall_cap_minutes: 45

- Measured environment declared before the first probe: Claude Code 2.1.284, tmux 3.6a, macOS (darwin). Headless probes declare `--model haiku`.
- Probes run in a disposable scratchpad repository structure under /tmp (verdict §① geometry: an outer git repo with an L1 worktree under `.claude/worktrees/w1`). Neither the primary checkout nor the t1279 worktree is a probe subject (REQ-SDL-006). Fixture repos are disposable test structures; creating their worktree with `git worktree add` inside /tmp is the fixture-construction act, not a doctrine-scope operation.
- Every probe is bounded by an external `timeout 300`; the tmux-hosted interactive probe wraps its in-session claude with the same bound (plan §E.1).
- One spare probe is reserved for a single retry of a timed-out or cap-blocked probe (plan §E).
- A reached cap makes the unobserved item a Gap, not a result (REQ-SDL-005). Declared gap route for probe 3 (`clear_restores_skill_set: gap`, `clear_method: none`) is an expected outcome, not a failure (plan §E.1).
- Probe commands are single compound invocations in the scrubbed `unset MOAI_KANBAN … && timeout 300 claude …` form; no heredocs, no `env -u` (plan §D).
- No process is killed by name; the tmux session ends with `tmux kill-session -t <probe-unique-name>` (plan §G).
