# SPEC-MANAGER-TODO-001 — design.md

Design decisions for the run phase. Each decision names its alternatives and why it was chosen. Cross-references: research.md §C (conflict report), spec.md §B (REQs).

## D-1 — Rename in place; repurpose by body rewrite, not by file split

The agent is renamed `mission-governor` → `manager-todo` in the same file path pattern (`.claude/agents/moai/manager-todo.md`). Alternatives rejected: (a) keep mission-governor and add a new agent — net catalog addition, breaks the 13/12 invariant (C-1) and the "adding a new MoAI-custom agent without a SPEC justifying recurrence" anti-pattern; (b) two agents (judgment + queue) — same violation, plus two-copy maintenance. The repurpose is carried by the body: the sealed-snapshot judgment contract (read-only scope boundary, one bounded decision object, blocker-on-out-of-scope output shape) is preserved verbatim as a clearly-labeled sub-role section inside the new body, and the queue-management / dispatch-guidance mission becomes the primary mission.

Frontmatter disposition: `tools` widens beyond the judgment-only set to support queue inspection and dispatch guidance (run-phase finalizes the exact CSV, minimum-privilege; candidates: Read, Grep, Glob, Grep, Skill + the minimum mutation surface the queue role needs, decided at run against the flat-hierarchy tool rules); `permissionMode` re-evaluated at run (the judgment sub-role's read-only discipline moves from frontmatter mode to prompt contract where the queue role needs more than `plan`); `memory: project` retained; no `model:`/`effort:` fields (model-policy inherit-by-default).

## D-2 — `--auto` lives on the todo CLI surface, reached through both entry points

Measured: `/moai:todo` is a thin alias routing to `gtd` (`.claude/commands/moai/todo.md` → `Skill("moai") with arguments: gtd $ARGUMENTS`), and the CLI verbs live under `moai todo` / the gtd surface (`internal/cli/todo.go`). The `--auto` flag is implemented once on the todo CLI command tree and works through both entry points (`/moai:todo --auto` and the gtd skill routing). No second implementation at the skill layer: the slash command stays a thin router (thin-command pattern), the skill/workflow text documents the serial contract, the CLI owns the mechanics.

## D-3 — Codex read-only role roster: RESOLVED — rename-registered, never removed (lead decision 2026-09-29)

The measured conflict (all four surfaces put `mission-governor` in the read-only role roster: `codex_audit_mcp.go:199`, `agents-codex.yaml:83/489`, the catalogue rule row, the fingerprint prose) is resolved by lead decision 2026-09-29, option (a): **the roster keeps the role — mission-governor is renamed-registered as `manager-todo`, never removed.** The sealed-snapshot + read-only contract continues as manager-todo's sub-role, including on the Codex path. All four measured surfaces update in the rename direction with zero removals. The roster-renewal coupling cost of keeping a dispatch-owning agent registered in a read-only roster is explicitly accepted by the lead (the sub-role stays read-only by prompt contract; roster entries are re-affirmed at the next roster review rather than re-derived). The earlier removal disposition recorded in the iter-1 draft is superseded; research.md §C carries the same resolution. No run-entry blocker remains from this decision.

## D-4 — Pickup ownership: two-channel liveness measurement, conservative default

The dead/ended-owner predicate (REQ-MT-008) requires BOTH channels to agree the owner is gone:

1. **Session registry channel** — the owning session has no live registry entry (`moai session list` / `.moai/state/active-sessions.json` equivalent), or its recorded PID fails a process probe (`kill -0`).
2. **lsof cwd channel** — `lsof` shows no live process whose working directory sits inside the owning session's tree (the pattern `internal/cli/update_worktree_processes.go` already uses; the project lesson "judge a card owner with lsof cwd, not the registry alone" makes the registry insufficient by itself).

Decision rule: registry-dead AND lsof-clean → owner dead, card is a pickup target; EITHER channel shows life → owner alive, card untouchable (REQ-MT-009, conservative default — a false "alive" costs a skipped card; a false "dead" steals a living session's work). Measurement is re-taken at each pickup decision, never cached across cards.

## D-5 — Pickup predicate in positive state vocabulary (hold-forward-compatible)

The cycle's selection is written as `state == "queued"` plus the dead-owner `picked` carve-out — never `state != "done"`. When card t1308's future `hold` state arrives, `hold` cards simply do not match the positive predicate and are skipped with zero SPEC revision. AC-MT-010 verifies this by asserting the predicate's source form and by a table-driven test enumerating all known states with their expected pickup outcomes.

## D-6 — Codemaps disposition: note in-milestone, regenerate on next codemaps run

The 8 codemap hits (`.moai/project/codemaps/{data-flow,docs-truth}.md`) record the catalog as generated truth. Editing generated files by hand re-introduces drift. Chosen: M2 records the pending rename in the baseline; the codemaps are regenerated by the next `/moai codemaps` run after the rename lands (the run-phase completion checklist names this). If run-phase measures a CI guard failing on the stale codemaps, the disposition flips to in-milestone regeneration — recorded as a conditional, not an open question.

## D-7 — /clear guidance emission is a CLI output contract, not a hook

The per-card /clear guidance (REQ-MT-011) is emitted by the `--auto` command itself as part of its per-card completion output (naming the completed card and the next step), before the next pickup. Alternatives rejected: a Stop-hook emission (wrong surface — the operator may not be in the emitting session; also violates the hook self-gate discipline) and a skill-layer-only convention (unmechanical). The AC greps the command's output-construction path and runs the command against a fixture queue to observe the guidance between two completed cards.

## D-8 — Batch approval reconciliation (operator-act doctrine)

The existing doctrine: "promotion is the operator's act, always." `--auto` reconciles by making the **invocation** the operator's act: the operator who types `--auto` has, by that act, approved serial consumption of the queue in queue order — nothing else is promoted. The command derives its authority exclusively from the invocation event, never from queue emptiness, card readiness, or a peer's request (the anti-patterns the doctrine names). The skill text states this explicitly, carrying this canonical sentence (the string AC-MT-014 pins): "`/moai:todo --auto` is the operator's batch approval: it authorizes serial consumption of the queue in queue order and nothing else." The CLI implements no reordering or admission logic beyond queue order.

## D-11 — Card processing means: the foreman pattern (in-session worker dispatch, evidence-read completion)

REQ-MT-007's "process" stage is defined by reusing the existing `moai-kanban-foreman` precedent (reuse-ladder step 2 — the pattern already exists at `.claude/skills/moai-kanban-foreman/`): **pick → dispatch one isolated in-session `Agent()` worker for the card → judge completion by reading the worker's disk evidence (progress record / verdict path), never on the worker's own claims → record the `done` transition on that evidence.** The completion-observation mechanism is read-on-read: the cycle transitions a card to `done` only after reading evidence the worker wrote to disk, per the completion-is-read-never-trusted doctrine.

Failure path: a worker that dies mid-card, or whose evidence file is absent or unreadable at collection, leaves the card **unpicked** (returned to `queued` via the existing `unpick` verb) with a labelled non-finding in the cycle output — never silently `done`, never left picked by the `--auto` session.

Boundary (restates REQ-MT-013): this dispatch is **in-session `Agent()` fan-out owned by the `--auto` CLI process — it creates no factory run, no factory lease, and no factory slot**, and it is not the factory messaging path; the factory integration remains a seam for a follow-up card only.

Session reference for guidance emission (closes the iter-1 residual-risk note on D-7): the per-card /clear guidance names the session that invoked `--auto` — the operator's session hosting the cycle and the next card's dispatch — not the ephemeral worker session.

## D-9 — GTD absorption: sub-role, receipt schema frozen

The goal workflow's judgment clause names manager-todo; the judgment sub-role section in the agent body carries the read-only scope boundary, the one-decision-object output shape, and the blocker-on-excess-scope behavior of the former mission-governor, so the sealed-snapshot contract is textually preserved. `goal run --governor-receipt` keeps its flag name (`--governor-receipt`) — the flag name is a CLI surface with consumers; only its help text changes ("manager-todo decision receipt"). Receipt path `.mo/state` convention unchanged.

## D-10 — Milestone order (decision-reversibility note)

The card's suggested order M1→M5 is kept with one justification: M1 (rename) is mechanically the lowest-uncertainty milestone but is a **naming dependency of every other milestone** — M2's sweep checklist, M3's CLI prose, M4's roster and goal-workflow text, and M5's docs all reference the post-rename name, so landing the rename first prevents every later milestone from re-touching files. The genuinely high-change-likelihood decisions (pickup semantics D-4/D-5, processing means D-11) are front-loaded in this design document and in plan.md §F's milestone risk notes, so human review lands there first.
