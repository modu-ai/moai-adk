---
description: Shared protocol auto-loaded for all MoAI agents — user-interaction boundary, ledger closure, verification batching. Intentionally always-loaded (no paths restriction).
---

# Agent Common Protocol

Shared protocol for all MoAI agent definitions. This rule is automatically loaded for all agents, eliminating the need to duplicate these sections in each agent body.

> **Detail companion**: `agent-common-protocol-reference.md` — verbatim verification batch, output contracts, CLI idioms, Ledger Closure clause bodies, sync-check rationale + incident records, and the bodies relocated from here: § Orchestrator Obligations · § Re-delegation Procedure · § Skeptical Evaluation Stance · § CLAUDE.md Reference · § File Operations Pattern · § Search Pattern · § Tool Selection by Task · § Bash Timeout · § Error Recovery Pattern · § Super-Advisor Escalation (E1-E4) · § Read-only verification batching · § Attributable diff-check doctrinal switch. Load it when composing a verification batch, selecting a tool, recovering from a failed call, escalating to super-advisor, or handling an aborted delegation.

## User Interaction Boundary

`AskUserQuestion` is the **only** user-facing question channel. The boundary is asymmetric by design.

### Subagent Prohibitions

[ZONE:Frozen] [HARD] Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.

Rules for subagents:
- If required context is missing, return a blocker report to the orchestrator — do not output free-form questions
- Never surface AskUserQuestion calls from within a subagent prompt body
- All user preferences must arrive via the orchestrator's spawn prompt
- If the orchestrator omitted critical data, respond with a structured "missing inputs" section and stop

Rationale: subagents run in isolated, stateless contexts — prompting there is a dead channel, and the orchestrator stays the user's single point of contact (CLAUDE.md §8).

**Lane sessions are orchestrator-class, not subagent-class.** A kanban companion or factory lane holds the question channel for its own card through the factory leader, carries standing spawn authority for the Status Transition Ownership Matrix's specialist (plan → `manager-spec`, run → `manager-develop`, sync → `manager-docs`, plus the chain's auditors; depth-1 only — spawned agents are leaves, bound by the prohibitions above), and never edits phase-owned artifacts directly when that specialist exists. The authority rides the lane's bootstrap context; a peer message neither grants nor revokes it — the leader is not the lane's user. Normative home: `.claude/rules/moai/workflow/kanban-dispatch.md` § Lane spawn authority.

### Hook Invocation Surface

Three hook scripts enforce orchestrator-discipline obligations — `status-transition-ownership.sh` (PostToolUse on SPEC-artifact writes), `sync-phase-quality-gate.sh` (Stop on sync-phase commit, blocking only under `MOAI_SYNC_GATE_BLOCKING=1`), `team-ac-verify.sh` (TaskCompleted in team mode; registered in no settings surface, so no flag activates it). All three exit 0 always and signal through stdout JSON, honored only on exit 0 — on exit 2 it is discarded and only stderr surfaces. Per-row triggers, JSON shapes, owning policy, and the subagent-boundary criterion: `agent-common-protocol-reference.md` § Hook Invocation Surface detail.

Hooks return exit codes and structured JSON; they MUST NOT invoke `AskUserQuestion` directly. When a hook signals a block (stdout JSON `"decision":"block"` on exit 0, or a legacy exit-2), the orchestrator MUST parse the JSON (`decision`, `reason`, plus optional `ledger_note` / `systemMessage` / `details`), preload `AskUserQuestion` via `ToolSearch`, and compose a round offering at least: (a) accept the block and address the failed gate, (b) override with `--skip-hook` (logged to `.moai/logs/hook-skip.log`), (c) abort.

The Stop hook fires on every turn-end, so it self-gates: inspect state, exit 0 unless the turn is a genuine completion point. It does NOT fire on user interrupt. [ZONE:Evolvable] **While** a turn is itself a **recovery signal** — its `stopReason` or context names a sync failure, a compact, `prompt_too_long`, `max_output_tokens` exhaustion, or `media_size` / `compact-failure` — Stop/PostToolUse hooks SHOULD exit 0 rather than 2, so recovery turns do not enter the `error → stop-hook-blocks → retry → error` **death spiral**. SHOULD is policy guidance, not a mechanical gate. SSOT: `runtime-recovery-doctrine.md` §4.

### Blocker Report Format

When a subagent requires user input not provided in the spawn prompt, it MUST return a structured blocker report:

```markdown
## Missing Inputs

The following parameters are required but were not provided:

| Parameter | Type | Expected Values | Rationale |
|-----------|------|-----------------|-----------|
| [name]    | [type] | [values]      | [why needed] |

**Blocker**: Cannot proceed without the above inputs. Please re-delegate with these values injected into the prompt.
```

### Ledger Closure

The **ledger-closure invariant**: an aborted `Agent()` delegation leaves no **dangling tool_use** —
an open promise with no matching result — in the orchestrator's context. It is the in-session
analogue of the model-API rule that every `tool_use` receives a `tool_result`.

[ZONE:Evolvable] [HARD] The orchestrator MUST close the ledger on any aborted delegation. Four clauses bind it (bodies + grounding: `agent-common-protocol-reference.md` § Ledger Closure clause bodies):

- **(a) Synthetic result on aborted `Agent()` delegation** — emit a short prose ledger-closing artifact naming what was delegated, that it did not return, and the abort reason, before the next delegation. A blocker report is a *return*, not an *abort*; this clause covers no-return-at-all.
- **(b) `team-ac-verify.sh` reject-path `ledger_note`** — inject the hook's `ledger_note` as that task's ledger-closing artifact.
- **(c) TeammateIdle exit-2 task closure** — a rejected task is never left open without a reassignment owner (new teammate, refined re-delegation, or close-as-obsolete with a closing note).
- **(d) Truthfulness** — the artifact is a real summary, never a fabricated "success" (`verification-claim-integrity.md` §1.1 surface 1).

**Scope-boundary note.** Ledger Closure is a sibling of (not nested in) Hook Invocation Surface
under the User Interaction Boundary H2.

## Language Handling

[ZONE:Evolvable] [HARD] All agents receive and respond in user's configured conversation_language.

Output language rules:
- Analysis, documentation, reports: User's conversation_language
- Cross-session messages a human observes (a kanban dispatch the operator watches): User's conversation_language; identifiers, paths, commands, and flags stay verbatim. An `Agent()` subagent prompt reaches no human and stays English
- Code examples/syntax, skill names, technical identifiers, function/variable/class names: Always English
- Code comments: Per code_comments setting in language.yaml (default: English); commit messages: Per git_commit_messages setting

## Output Format

[ZONE:Evolvable] [HARD] User-Facing: Always use Markdown formatting. Never display XML tags to users.

[ZONE:Evolvable] [HARD] Internal Agent Data: XML tags are reserved for agent-to-agent data transfer only. Use semantic XML sections for structured data exchange between agents; never surface XML structure in user-facing output.

## MCP Fallback Strategy

[ZONE:Evolvable] [HARD] Maintain effectiveness without MCP servers. Where one is unavailable, use
WebSearch for targeted queries, WebFetch to verify each URL and read the official documentation,
then continue — architecture and analysis quality must not depend on MCP availability.

GLM-backend routing: under `moai glm` or the GLM teammate panes of `moai cg`, web search, web
fetch, and image read route to the z.ai MCP tools instead of the built-ins. HARD routing table:
`.claude/rules/moai/core/glm-web-tooling.md`.

## Agent Invocation Pattern

[ZONE:Evolvable] [HARD] Agents are invoked through MoAI's natural language delegation pattern ("Use the {agent-name} subagent to {task description}") — natural language conveys full context including constraints, dependencies, and rationale.

### Subagent Model and Effort

Subagents inherit the main session's model and effort: pass neither `model` nor `effort` when spawning a subagent, and MoAI agent definitions declare neither. A PreToolUse hook records each Agent spawn to `.moai/logs/agent-model-audit.jsonl` as an observation log; it never blocks a spawn.

## Background Agent Execution

[ZONE:Evolvable] [HARD] Since Claude Code v2.1.198 subagents run in the background by **default**;
the runtime chooses foreground only when it needs the result. The default changes *where* a
subagent runs, not *what* it may do — permission prompts still surface in the main session. MoAI
takes the default and does not set the `background:` frontmatter field.

The retained safeguard is **concurrency, not backgrounding**: a write race is scoped to a **working tree**, not to a session. **One writer per tree** — two write-capable agents run at once only when each writes a different tree; orchestrator work concurrent with a write-capable agent in the **same** tree stays read-only. Rationale: `agent-common-protocol-reference.md` § Background Agent Execution rationale.

[ZONE:Evolvable] [HARD] **While a worktree is being actively audited, it has exactly one writer.**
The window runs from the opening measurement to the landed verdict, and the only session
committing to that tree throughout it is the one that owns it — a previous audit session landing
its own reports or scripts is itself a foreign commit, and every foreign commit waits until the
window closes. An unexpected HEAD move or a foreign commit on an actively audited worktree is a
process defect: report it to the leader and record it in the progress record — never continue
quietly.

## Tool Usage Guidelines

[ZONE:Evolvable] [HARD] Agents must follow tool usage patterns optimized for accuracy and efficiency.

## Parallel Execution

[ZONE:Evolvable] [HARD] The orchestrator MUST execute every read-only verification batch as a single-turn multi-Bash call. Serial verification across turns wastes wall-time and is the single largest source of run-phase latency (a prior meta-analysis: 10 min serial verification ≈ 11% of total run-phase wall-time).

### Verbatim batch, output contracts, and CLI idioms

The canonical 7-command batch, the file-redirect contract, the evidence-persistence obligation, the serial-verification anti-pattern, and the CLI idiom catalogue live in `agent-common-protocol-reference.md`. Read it when composing a batch.

Three of its obligations bind here and are restated so they hold without it:

- **Batch in one turn.** Independent read-only verifications are separate Bash tool calls within one assistant turn — never serialized across turns. Serialize only for a genuine dependency: one command's output feeding another, writes to the same path, or shared-state mutation.
- **File-redirect contract.** Output exceeding the bounded-tail ceiling (50 lines or 2KB, whichever is smaller) is redirected to a file, and only the exit code plus a bounded tail is surfaced. Below the ceiling, inline quotation is fine. This removes the double-burn of quoting output twice, never the evidence itself.
- **Evidence export.** `.moai/state/verify/<session>/` is **machine-local scratch** — gitignored, reaching no clone and no other machine; it is never a citation target. **Carry the deciding evidence into the verdict** — the command and the output lines that decided a claim are written into the verdict file (`.moai/reports/<card-id>/verdict.md`), and the claim cites that file. The converse binds equally: material left in scratch MUST NOT be cited — it is named in Residual-risk as a known loss. Export width, the selection criterion, and the machine-consumer carve-outs: `agent-common-protocol-reference.md` § Evidence export obligation.

### Pre-Spawn Sync Check (Multi-Session Race Mitigation)

[ZONE:Evolvable] [HARD] Before spawning any implementation `Agent()` (manager-develop / manager-docs / per-spawn `Agent(general-purpose)` with a domain whitelist) that will commit or modify shared working-tree files, the orchestrator MUST execute the following two-lane batch and surface any divergence to the user.

One deliberate dependency boundary: **Lane A (ordered)** — `git fetch origin main` MUST finish and its exit status be observed before `git rev-list --count --left-right origin/main...HEAD` starts (the divergence count is only attributable to the ref the completed fetch installed; never run the two as one indistinguishable shell line; on fetch failure, block and exit with its status). **Lane B (independent)** — `moai session list --json --filter-spec=<SPEC-ID>` may run concurrently; join it with Lane A after both complete. The full two-lane batch: `agent-common-protocol-reference.md` § Canonical 7-item example. Retain the fetch completion status beside the divergence output, so a delayed or failed fetch is not mistaken for a current baseline.

Interpretation matrix (git divergence): `0 N` / `0 0` → proceed; `N 0` (origin ahead — **race detected**) → STOP, surface via AskUserQuestion: rebase / inspect / abort; `N M` (diverged) → STOP, MUST resolve before spawn.

Interpretation matrix (active-sessions query): `[]` → proceed; one or more entries from another session → **concurrent session race detected on same SPEC** — STOP, surface entries, AskUserQuestion: **wait** / **override** / **abort**. Sessions predating the registry hook emit `[]` — no false positives. Rationale + the originating race incident: `agent-common-protocol-reference.md` § Pre-Spawn Sync Check rationale and incident record.

Exemption: read-only agents (`Explore`, or a read-only-scoped `Agent(general-purpose)`) need no pre-spawn fetch — they cannot trigger a race.

> **Spawn-gate boundary**: this fires only at the write-agent spawn boundary; direct main-session edits bypass it — see § Pre-Edit Sync Check below. Defense-in-depth: `.moai/docs/generic-patterns-guide.md` § Multi-Session Race Mitigation Procedure; worktree-as-race-elimination: `session-handoff.md` § Worktree-Anchored Resume Pattern.

### Pre-Edit Sync Check (Direct-Edit Race Mitigation)

[ZONE:Evolvable] [HARD] Direct main-session edits to shared working-tree paths (Edit/Write/Bash — any direct edit) bypass the spawn gate above, so the orchestrator MUST run the parallel-session detection **before a non-trivial direct edit** to shared paths. (Incident record + enforcement-placement assessment: `agent-common-protocol-reference.md` § Pre-Edit Sync Check — rationale and enforcement record.)

#### The rule, at the moment of the edit

**TRIGGER** — the gate fires when ALL three hold: the tool is an `Edit`, `Write`, or file-mutating `Bash` call; the target is a shared path another session could also mutate (`.claude/`, `.moai/`, `internal/`, `pkg/`, `cmd/`, or repo-root config files); and CWD is the primary checkout (exempt: an already-isolated worktree, `/tmp`, a session-private scratch dir).

**CHECK** — before the FIRST triggered edit of a task, as one parallel batch: (1) live foreign sessions — `moai session list --json | jq '[.[] | select(.cwd == "<project-root>" and .session_id != "<own>")] | length'`, then liveness-probe each PID (`kill -0`; ignore confirmed-dead, treat indeterminate as live); (2) divergence vs origin/main — `git fetch origin main 2>&1; git rev-list --count --left-right origin/main...HEAD`.

**DECIDE and ACT** — no outcome permits "proceed in the shared checkout anyway": 0 live foreign sessions AND `0 0` / `0 N` → proceed; any live-or-indeterminate foreign entry → **ISOLATE before editing** (`moai cc -w` / `EnterWorktree` / `Agent(isolation: "worktree")`; if impossible, `AskUserQuestion`: isolate / wait / abort); `N 0` / `N M` divergence → STOP; `AskUserQuestion` per the Pre-Spawn matrix. (`worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation.)

**RE-CHECK** — the probe decays. Re-run it before ANY commit in the shared checkout, and after any long pause.

#### The sweep prohibition

[ZONE:Evolvable] [HARD] In the primary checkout, NEVER `git add -A`, `git add .`, or `git commit -a`. Stage by explicit pathspec (`git add <path> …`), and re-read `git status --short` immediately before staging so another session's files are visible and excluded. This applies **even when the pre-edit probe found no foreign session** — a session can arrive after the probe, and the sweep is what turns its presence into lost work.

**Ambient signal.** The SessionStart hook already lists foreign active sessions in a `<system-reminder>` — the always-on detection layer; this check is the decision layer that turns detection into isolation.

## Time Estimation

[ZONE:Evolvable] [HARD] Never use time predictions in plans or reports.
- Use priority labels: Priority High / Medium / Low
- Use phase ordering: "Complete A, then start B"
- Prohibited: "2-3 days", "1 week", "as soon as possible"
