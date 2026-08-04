---
description: Shared protocol auto-loaded for all MoAI agents — user-interaction boundary, ledger closure, verification batching. Intentionally always-loaded (no paths restriction).
---

# Agent Common Protocol

Shared protocol for all MoAI agent definitions. This rule is automatically loaded for all agents, eliminating the need to duplicate these sections in each agent body.

## User Interaction Boundary

`AskUserQuestion` is the **only** user-facing question channel. The boundary is asymmetric by design.

### Subagent Prohibitions

[ZONE:Frozen] [HARD] Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.

Rules for subagents:
- If required context is missing, return a blocker report to the orchestrator — do not output free-form questions
- Never surface AskUserQuestion calls from within a subagent prompt body
- All user preferences must arrive via the orchestrator's spawn prompt
- If the orchestrator omitted critical data, respond with a structured "missing inputs" section and stop

Rationale:
- Subagents run in isolated, stateless contexts and CANNOT interact with users directly
- Attempting to prompt inside a subagent produces a dead channel — no response arrives
- This rule preserves the orchestrator's single-point-of-contact with the user (see CLAUDE.md Section 8)

### Orchestrator Obligations

> Canonical: see `.claude/rules/moai/core/askuser-protocol.md` § Orchestrator Obligations for the full preload sequence (`ToolSearch(query: "select:AskUserQuestion")` before each call), the AskUserQuestion channel monopoly, the Socratic interview structure, and the option-description standards. This file owns only the subagent-side boundary (above) and the blocker-report → re-delegation flow (below).

The MoAI orchestrator collects all user preferences before delegating to subagents via `Agent()`. On receiving a blocker report from a subagent, it runs an `AskUserQuestion` round, injects the user's responses into a fresh subagent prompt, and re-delegates (procedure below).

### Hook Invocation Surface

Per the canonical hook invocation surface policy, the orchestrator interacts with three NEW hook scripts that mechanically enforce orchestrator-discipline obligations previously delegated to phantom `manager-quality` / `expert-security` spawn calls:

| Hook script | Trigger | Owning REQ | Exit-code semantics |
|-------------|---------|------------|---------------------|
| `.claude/hooks/moai/status-transition-ownership.sh` | PostToolUse on Write/Edit of `.moai/specs/SPEC-*/{spec,plan,acceptance}.md` body content | Status Transition Ownership Matrix per `.claude/rules/moai/development/spec-frontmatter-schema.md` | exit 0 always (advisory — the transition site is audit-logged to `.moai/logs/status-transition-audit.log`; exit-2 blocking is reserved for future ownership-mismatch enforcement) |
| `.claude/hooks/moai/sync-phase-quality-gate.sh` | Stop hook on sync-phase commit completion | sync-phase quality gate policy (lint + test + coverage delta) + dependency manifest audit on `go.mod` / `package-lock.json` / etc. changes | exit 0 always; a failing check emits an advisory `systemMessage`, and blocking mode (opt-in via `MOAI_SYNC_GATE_BLOCKING=1`) emits stdout JSON `{"decision":"block"}` — per Claude Code hook semantics, stdout JSON is honored only on exit 0 (on exit 2 it is discarded and only stderr is surfaced) |
| `.claude/hooks/moai/team-ac-verify.sh` | TaskCompleted in team mode (dormant by default — activates only under harness `thorough` + team mode prerequisites per the canonical team activation policy) | per-AC PASS evidence file verification | exit 0 always; a completion rejection is signaled via stdout JSON `{"continue":false,"stopReason":"AC verification failed: ...","ledger_note":"..."}` (the reject decision MUST ride the exit-0 stdout channel — exit 2 would discard it; `decision` is not valid for TaskCompleted) |

#### Orchestrator translation responsibility

Hooks return exit codes and structured JSON; they MUST NOT invoke `AskUserQuestion` directly per the orchestrator-subagent boundary above. When a hook signals a block (stdout JSON `"decision":"block"` on exit 0, or a legacy exit-2), the orchestrator MUST:

1. Parse the hook's structured JSON output (`decision`, `reason`, plus optional `ledger_note` / `systemMessage` / `details`)
2. Preload `AskUserQuestion` via `ToolSearch(query: "select:AskUserQuestion")`
3. Compose an `AskUserQuestion` round presenting the user with at least: (a) accept the block and address the failed gate, (b) override with `--skip-hook` opt-out (logged to `.moai/logs/hook-skip.log` per audit trail), (c) abort the workflow

This translation pattern preserves the orchestrator's single-point-of-contact with the user per CLAUDE.md §8 + this rule's User Interaction Boundary section above. Hook subagent boundary verification is covered by the canonical hook subagent boundary acceptance criterion:

```bash
grep -rn 'AskUserQuestion\|mcp__askuser' .claude/hooks/moai/ \
  | grep -v "^[^:]*:[0-9]*:[ \t]*#"
# Expected: no matches (hook scripts do not invoke AskUserQuestion)
```

#### Stop self-gate caveat

The `sync-phase-quality-gate.sh` row above describes the Stop hook in the sync-phase context, but the Stop hook is not exclusive to sync-commit completion. The Stop hook fires on every turn-end — not only when a task is complete — so a Stop hook must self-gate: it inspects the conversation/working-tree state and decides whether the turn is a genuine completion point before acting, otherwise exiting 0 to allow the turn to end without intervention. The Stop hook does NOT fire when the user interrupts the turn, so it cannot be relied on as a guaranteed end-of-work signal.

#### Recovery-Signal Carve-Out

**Recovery-Signal Carve-Out** — anti-death-spiral policy guidance for Stop/PostToolUse hooks. The canonical doctrine lives at `.claude/rules/moai/workflow/runtime-recovery-doctrine.md` §4 (SSOT); this subsection is the render surface.

[ZONE:Evolvable] **While** a turn's `stopReason` or surrounding context indicates the turn is itself a **recovery signal** — i.e., the turn is recovering from a sync failure, a compact, a `prompt_too_long` (PTL), a `max_output_tokens` exhaustion, or a `media_size` / `compact-failure` — Stop/PostToolUse hooks SHOULD exit 0 (allow the turn to end / the tool call to proceed) rather than exit 2 (block), so that recovery turns are NOT placed into the `error → stop-hook-blocks → retry → error` loop that book1 ch06 names the **death-spiral**.

This carve-out is **policy guidance** (a SHOULD recommendation), NOT a mechanically-enforced gate:

- The current `sync-phase-quality-gate.sh` (Stop) and `status-transition-ownership.sh` (PostToolUse) hooks receive PostToolUse/Stop JSON but do not parse a recovery signal from `stopReason` or turn context; they therefore cannot mechanically distinguish a recovery turn from a normal turn.
- Mechanical enforcement of this carve-out is deferred to a future runtime-layer SPEC that would add `stopReason` parsing.
- The carve-out does NOT weaken the hooks' gate function on non-recovery turns — the gates still exit 2 (block) on genuine gate failures during normal turns. The carve-out only says recovery turns SHOULD defer to the recovery.

Determining "is this a recovery turn?" is the mechanical step the current hooks cannot take. See the SSOT doctrine (`runtime-recovery-doctrine.md` §4) for the full scope binding, the named-hook list (`sync-phase-quality-gate.sh`, `status-transition-ownership.sh`), and the reason this is documentation-only at this layer.

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

### Re-delegation Procedure

On receiving a blocker report, the orchestrator:
1. Invokes `ToolSearch(query: "select:AskUserQuestion")`
2. Runs an AskUserQuestion round to collect the missing inputs from the user
3. Constructs a fresh subagent prompt with the user's answers injected
4. Re-delegates to the subagent

### Ledger Closure

The **ledger-closure invariant** (externally grounded in `github.com/wquguru/harness-books` book1 ch04 "账本闭环" — "只要系统向外承诺了一段执行，就要在中断时把账补平": whenever the system has promised an execution externally, it must close the ledger on interrupt) states that an aborted `Agent()` delegation MUST NOT leave a **dangling tool_use** — an open promise with no matching result — in the orchestrator's own context. The persistence-layer analogue is `session-handoff.md` Block 3-4 preconditions (a `/clear` boundary re-establishes verifiable preconditions before continuing); this subsection codifies the in-session interrupt case (no `/clear`). This is the orchestration-layer analogue of the model-API rule that every `tool_use` receives a `tool_result`.

[ZONE:Evolvable] [HARD] The orchestrator MUST close the ledger on any aborted delegation. Four clauses bind this obligation:

- **(a) Synthetic result on aborted Agent() delegation.** When an `Agent()` delegation is aborted — user interrupt (Ctrl+C), parent-abort propagation (the orchestrator's own turn was aborted and the sub-agent was killed), or timeout (no return before a wall-clock or token-budget ceiling) — the orchestrator SHALL emit a **synthetic ledger-closing artifact** into its own context before issuing the next delegation. The artifact is a short prose summary (NOT a structured data record; no JSON schema, no `.moai/state/ledger.json`), naming what was delegated, that it did not return, and the abort reason if known. Its purpose is to close the open promise so the next turn does not proceed as if the delegation returned cleanly. This clause does NOT change the "Missing Inputs" blocker-report pattern above: a blocker report is a *return*, not an *abort*; this clause covers only the case where no return is produced at all.
- **(b) team-ac-verify.sh reject-path `ledger_note` field.** When `.claude/hooks/moai/team-ac-verify.sh` rejects a `TaskCompleted`, it signals the rejection via stdout JSON `{"continue":false,"stopReason":"AC verification failed: ...","ledger_note":"..."}` and exits 0 — per Claude Code hook semantics, stdout JSON is honored only on exit 0 (on exit 2 stdout is discarded and only stderr is surfaced), so the reject decision and its `ledger_note` MUST ride the exit-0 stdout channel. The `decision` field is NOT used here because it is documented only for PostToolUse/Stop/SubagentStop/UserPromptSubmit/ConfigChange/PreCompact/PostToolBatch — NOT TaskCompleted (the official TaskCompleted reject contract is `continue:false` + `stopReason`). The orchestrator injects this `ledger_note` as the ledger-closing artifact for that task. (The reject-path trigger itself is a minimal stub; full AC-verification logic is out of scope and deferred to a follow-up SPEC.)
- **(c) TeammateIdle exit-2 task closure.** When the TeammateIdle hook rejects a task's completion via exit-2 ("keep working"), the rejected task's TaskList entry MUST NOT be left in an open state without a reassignment owner. The orchestrator re-assigns the task (spawn a new teammate, re-delegate to the same teammate with a refined prompt, or close it as obsolete with a synthetic closing note). This binds the orchestrator's TaskList hygiene, not the hook's exit-2 emission. The parent-abort propagation that book1 ch07 names — cleanup handlers registered to avoid orphan tasks — is the source for this clause.
- **(d) Cross-references.** This subsection cross-references three sources:
  - **book1 ch04** (账本闭环 — the ledger-closure invariant named in the opening paragraph above).
  - **book1 ch07** (parent-abort propagates to forked children; agents are observable lifecycle objects via SubagentStart/SubagentStop hooks, exit-code-2 stderr feedback).
  - `.claude/rules/moai/workflow/session-handoff.md` Block 3-4 preconditions (the persistence-layer analogue of ledger closure across `/clear`).
  - The ledger-closing artifact's truthfulness is bound by `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report) — the artifact MUST be a real summary, not a fabricated "success".

**Scope-boundary note.** This Ledger Closure subsection is distinct from the Hook Invocation Surface subsection above (owned by the sibling Recovery-Signal Carve-Out). The two are siblings under the User Interaction Boundary H2; Ledger Closure is NOT nested inside Hook Invocation Surface. See the orchestrator-interrupt-ledger placement contract for the collision-free placement.

## Language Handling

[ZONE:Evolvable] [HARD] All agents receive and respond in user's configured conversation_language.

Output language rules:
- Analysis, documentation, reports: User's conversation_language
- Code examples and syntax: Always English
- Code comments: Per code_comments setting in language.yaml (default: English)
- Commit messages: Per git_commit_messages setting in language.yaml
- Skill names and technical identifiers: Always English
- Function/variable/class names: Always English

## Output Format

[ZONE:Evolvable] [HARD] User-Facing: Always use Markdown formatting. Never display XML tags to users.

- Reports, architecture docs, analysis results: Markdown with code blocks
- Progress updates and status: Markdown

[ZONE:Evolvable] [HARD] Internal Agent Data: XML tags are reserved for agent-to-agent data transfer only.

- Use semantic XML sections for structured data exchange between agents
- Never surface XML structure in user-facing output

## Skeptical Evaluation Stance

<!-- @MX:WARN: Duplication prohibited — LR-07 lint rule detects copies of this section in agent files and flags as error. Canonical copy lives only in this file. -->

The reviewer mode operates as a fresh-judgment auditor:

- Treat every claim as suspect until evidence is shown
- Demand reproducible verification, not assertions
- Consider the null hypothesis: did this change actually fix anything?
- Score quality as the harmonic mean of dimensions, not the average
- Reject when must-pass criteria fail, regardless of nice-to-have scores
- Surface contradictions; never silently override a prior rule
- Resist agreement: the RLHF training gradient biases toward flattery, so treat any urge to PASS without cited evidence as a sycophancy signal, not a verdict

## MCP Fallback Strategy

[ZONE:Evolvable] [HARD] Maintain effectiveness without MCP servers.

MoAI does not provision MCP servers; use WebSearch and WebFetch to look up library documentation and established best-practice patterns. When external lookups are needed:
1. Use WebSearch with targeted queries to find candidate sources
2. Use WebFetch to verify each URL and read the official documentation
3. Deliver established best-practice patterns based on industry experience
4. Continue work — architecture/analysis quality must not depend on MCP availability

GLM-backend routing: when the session runs on the GLM backend (`moai glm` or the GLM teammate panes of `moai cg`), web search / web fetch / image read route to the z.ai MCP tools instead of the built-in `WebSearch` / `WebFetch` / `Read`. See `.claude/rules/moai/core/glm-web-tooling.md` for the HARD routing table.

## CLAUDE.md Reference

Agents follow MoAI's core execution directives defined in CLAUDE.md. Since CLAUDE.md is automatically loaded as project instructions, agents do not need to restate its rules. Key applicable principles:

- SPEC-based workflow (Plan-Run-Sync)
- TRUST 5 quality framework
- Agent delegation hierarchy
- Parallel execution safeguards

## Agent Invocation Pattern

[ZONE:Evolvable] [HARD] Agents are invoked through MoAI's natural language delegation pattern:
- "Use the {agent-name} subagent to {task description}"
- Natural language conveys full context including constraints, dependencies, and rationale

Architecture:
- Commands orchestrate through natural language delegation
- Agents own domain-specific expertise
- Skills auto-load based on YAML frontmatter configuration

## Background Agent Execution

[ZONE:Evolvable] [HARD] As of Claude Code v2.1.198, subagents run in the background by **default**; Claude runs one in the foreground only when it needs the result before continuing. The default changes *where* a subagent runs, not *what it may do* — a background subagent still surfaces every permission prompt in the main session, and (since v2.1.186) that prompt names the asking subagent (Esc denies just that one call). MoAI **aligns with this runtime default** rather than forcing foreground for write-capable agents, and does not set the `background:` frontmatter field — the runtime's per-call heuristic chooses.

The retained safeguard is **concurrency, not backgrounding**: MoAI does not run two write-capable agents concurrently, and orchestrator work performed concurrently with a write-capable agent is **read-only**. This targets the actual hazard — a file-write race between agents — which forbidding background writes never addressed. The superseded restriction — a blanket ban on background Write/Edit — had its stated basis (background writes auto-denied) removed by v2.1.186 and no longer describes the runtime.

Rules for agent spawning:
- **Read-only tasks** (research, analysis, review): safe in the background; while one is in flight the orchestrator continues independent read-only work.
- **Write tasks** (implementation, refactoring, file creation): the runtime chooses foreground or background, and the permission prompt surfaces in the main session either way — do not force the mode via `background:`.
- **Concurrency**: never run two write-capable agents at once; orchestrator work concurrent with a write-capable agent stays read-only.
- **Pre-approved writes**: add path patterns to settings.json `permissions.allow` to reduce prompts.

## Tool Usage Guidelines

[ZONE:Evolvable] [HARD] Agents must follow tool usage patterns optimized for accuracy and efficiency.

### File Operations Pattern

Read-before-write rule:
- ALWAYS Read a file before using Edit on it
- Use Grep to locate specific line numbers before targeted Read with offset/limit
- Use Glob to discover files before reading — never guess file paths
- Prefer Edit over Write for existing files (sends only the diff, preserves context)

Path handling:
- Use absolute paths for all file operations
- Never construct paths from assumptions — verify with Glob or Bash `ls` first
- When working in worktrees, use project-root-relative paths for write targets

### Search Pattern

Progressive narrowing:
1. Glob to find candidate files by pattern
2. Grep with `files_with_matches` to narrow by content
3. Grep with `content` mode + context lines for detailed inspection
4. Read with offset/limit for full section understanding

Avoid:
- Reading entire large files when only a specific section is needed
- Using Bash grep/find when Grep/Glob tools are available
- Searching without file type filters when the target language is known

### Tool Selection by Task

| Task | Preferred Tool | Avoid |
|------|---------------|-------|
| Find files by name | Glob | Bash find, Bash ls |
| Search file contents | Grep | Bash grep, Bash rg |
| Read file contents | Read | Bash cat, Bash head |
| Modify existing file | Edit | Bash sed, Write (overwrites) |
| Create new file | Write | Bash echo/cat heredoc |
| Run system commands | Bash | — |
| Explore codebase | Agent(Explore) | Multiple sequential Grep calls |

### Bash Timeout

The Bash tool supports an optional `timeout` parameter (milliseconds):

- Default: 120,000ms (2 minutes)
- Maximum: 600,000ms (10 minutes)
- Use for long-running commands: builds, test suites, installs

Specify via the `timeout` field when the command is expected to run longer than 2 minutes.

### Error Recovery Pattern

When a tool call fails:
1. Read the error message carefully — diagnose root cause
2. Verify assumptions: does the file/path exist? (Glob check)
3. Try an alternative approach — do not retry the identical call
4. After 3 failures on the same operation, report the blocker

**Retry safety is asymmetric with respect to a call's side effects.** Before applying the 3-retry ceiling above, classify the failed call by its side-effect profile:

- **Idempotent / read-only calls** (re-reading a file, re-running a search or query, re-running an initializer, fetching a URL) may be retried up to the ceiling — repeating them produces the same observable result, so a transient failure (a file lock, a network blip) is legitimately recovered by a retry.
- **Side-effecting calls** (writing/editing a file, committing, pushing, opening a pull request, deploying, mutating external-API state) carry a duplicate-effect hazard. When a side-effecting call fails *ambiguously* — the failure signal is present but whether the effect already landed is uncertain — first **observe the current state** to determine whether the effect already occurred, and retry only when the effect is confirmed absent. Retrying a side-effecting call without first observing state is the duplicate-effect hazard: a blind retry after an uncertain failure risks a duplicate commit, a duplicate pull request, or a double deploy. The absence of a success signal is not evidence the effect did not land.

This refines step 3 above ("do not retry the identical call") along the side-effect axis: for a side-effecting call, "try an alternative approach" begins with observing whether the effect already occurred.

### Super-Advisor Escalation (E1-E4)

When recovery via the 3-retry ceiling is insufficient OR a higher-reasoning consultation is warranted, the orchestrator escalates to the **super-advisor** agent (on-demand high-reasoning consultation). super-advisor returns **non-binding prescriptions** (diagnoses, options, recommendations); the orchestrator remains the decision owner. This is DISTINCT from auditor verdicts — `plan-auditor` / `sync-auditor` own binding PASS/FAIL judgment; super-advisor owns advisory consultation only. When the question is "should this PASS?", route to an auditor; when the question is "what should I do here?", route to super-advisor.

Entry conditions (exhaustive per the super-advisor entry-conditions contract; expansion is M4 doctrine territory):

| Trigger | Condition | Example |
|---------|-----------|---------|
| **E1 — bug-deadlock** | 3+ consecutive same-diagnostic failures | `manager-develop` retries the same failing test 3 times with the same root-cause hypothesis |
| **E2 — architecture/design decision point** | A spec-body or plan-body decision with ≥2 viable options, neither obviously correct | "Should this cache layer be write-through or write-behind?" at L-plan boundary |
| **E3 — second-opinion request** | Orchestrator uncertainty: < 80% confidence in the next delegation step | Ambiguous blocker-report from a worker; orchestrator deciding between re-spawn vs user-escalation |
| **E4 — loop-deadlock** | `/moai loop` or `/moai fix` ceiling-exit per the loop-verdict contract | Auto-fix iteration count exhausted without green CI |

On trigger: the orchestrator spawns `Agent(general-purpose)` with the super-advisor role profile (Opus + xhigh at max/medium tier; Sonnet + xhigh at low tier — GLM-backed sessions fall back to the session model per the GLM carve-out), receives a non-binding prescription, then either re-seeds the executor with the prescription or escalates to the user via `AskUserQuestion`.

Design source: `.moai/reports/agent-architecture-redesign-v2-20260709.html` §01 change ② + §05; agent file: `.claude/agents/moai/super-advisor.md`.

## Parallel Execution

[ZONE:Evolvable] [HARD] The orchestrator MUST execute every read-only verification
batch as a single-turn multi-Bash call. Serial verification across turns wastes
wall-time and is the single largest source of run-phase latency (a prior
workflow-optimization meta-analysis: 10 min serial verification ≈ 11% of total
run-phase wall-time). This rule was added by a prior workflow-optimization rule
Layer D in response to that finding.

### Read-only verification batching

When the orchestrator needs to verify implementation completion, it SHOULD issue
multiple Bash tool calls within a single response turn. Independent verifications
that do not share state are safe to parallelize.

### Verbatim batch, output contracts, and CLI idioms

The canonical 7-command batch, the file-redirect contract, the evidence-persistence
obligation, the serial-verification anti-pattern, and the single-command CLI idiom
catalogue (`gh pr checks --json … | jq`, `gh pr checks --watch --fail-fast` run in
background mode, `git log --format=…`, the per-turn `ToolSearch` preload) all live in
`agent-common-protocol-reference.md`. Read it when composing a batch.

Three obligations from that file bind here and are restated so they hold without it:

- **Batch in one turn.** Independent read-only verifications are issued as separate
  Bash tool calls within a single assistant turn — never serialized across turns.
  Serialize only for genuine dependencies: one command's output feeding another,
  writes to the same path, or shared-state mutation.
- **File-redirect contract.** When a command's verbatim output exceeds the bounded-tail
  ceiling (default: 50 lines or 2KB, whichever is smaller), redirect it to a file and
  surface only the exit code plus a bounded tail. Below the ceiling, inline quotation
  is fine. The contract removes the double-burn of quoting output twice, never the
  evidence itself.
- **Evidence persistence.** The cited path must still resolve at audit time, so
  evidence is persisted under `.moai/state/verify/<session>/` rather than left in
  `/tmp`, which the OS clears. A claim whose cited evidence path no longer resolves
  is an unattributed claim (`verification-claim-integrity.md` §2).


### Pre-Spawn Sync Check (Multi-Session Race Mitigation)

[ZONE:Evolvable] [HARD] Before spawning any implementation `Agent()`
(manager-develop / manager-docs / per-spawn `Agent(general-purpose)` with a domain whitelist) that will commit or modify
shared working-tree files, the orchestrator MUST execute the following
2-command parallel batch and surface any divergence to the user.

```bash
# 1. Fetch latest origin/main without merging
git fetch origin main 2>&1

# 2. Count divergence between local HEAD and origin/main
git rev-list --count --left-right origin/main...HEAD

# 3. Query active sessions on this host for the same SPEC scope (L1 of
#    the canonical 4-layer multi-session race mitigation policy).
#    Replace <SPEC-ID> with the SPEC about to be operated on.
moai session list --json --filter-spec=<SPEC-ID>
```

Interpretation matrix (git divergence):

| Output | Meaning | Action |
|--------|---------|--------|
| `0 N` | Local ahead by N (clean — your commits not yet pushed) | Proceed normally |
| `0 0` | Synced (local == origin/main) | Proceed normally |
| `N 0` | Origin ahead by N — **parallel session race detected** | STOP, surface via AskUserQuestion: rebase / inspect / abort |
| `N M` | Diverged (both ahead) | STOP, MUST resolve before spawn |

Interpretation matrix (active-sessions query — 3rd command):

| Output | Meaning | Action |
|--------|---------|--------|
| `[]` | No other session on this SPEC (per the multi-session coordination policy) | Proceed normally |
| `[{...}]` (≥1 entry from another session) | **Concurrent session race detected on same SPEC** | STOP, surface entries via prose summary, AskUserQuestion: **wait** / **override** / **abort** |

The 3rd command is **additive only** — the original 2-command batch
(git fetch + git rev-list) is preserved verbatim above.
Backward compatibility: sessions running before the multi-session
coordination policy was deployed (no registry hook) emit no entries,
so the 3rd command returns `[]` and the orchestrator proceeds normally
without false positives.

Rationale: When 2+ Claude Code sessions operate on the same project root
+ same memory hash (`~/.claude/projects/{hash}/memory/`), they may both
consume the same paste-ready resume and attempt the same `/moai <subcommand>`
work. The git working tree is shared; the memory file is shared. Without
a pre-spawn fetch, the second session works on a stale baseline and may
produce duplicate commits, conflicting frontmatter edits, or CHANGELOG
entry races.

Origin: an earlier sync-phase race incident — a parallel session
committed a spec.md frontmatter status update between manager-develop's
final run-phase commit and manager-docs' sync commit. Detection occurred
retrospectively when `git push` succeeded with an unexpected intermediate
commit in the push range. Lesson (parallel-session race during long agent runs) reinforced; a new lesson (pre-spawn fetch discipline) added.

Exemption: read-only agents (`Explore`, or a per-spawn `Agent(general-purpose)` scoped to read-only investigation) do not require pre-spawn fetch — they cannot trigger race conflicts.

> **Spawn-gate boundary**: this check fires only at the write-agent spawn boundary. Direct main-session edits (Edit/Write/Bash) bypass this gate; see § Pre-Edit Sync Check (Direct-Edit Race Mitigation) below for the direct-edit counterpart.

Cross-reference: `.moai/docs/generic-patterns-guide.md` § Multi-Session
Race Mitigation Procedure (defense-in-depth policy at user-facing
layer); `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored
Resume Pattern (L2/L3 worktree as race-elimination alternative).

### Pre-Edit Sync Check (Direct-Edit Race Mitigation)

[ZONE:Evolvable] [HARD] The Pre-Spawn Sync Check above binds only the spawn boundary. **Direct main-session edits to shared working-tree paths (Edit/Write/Bash in the orchestrator session — the MoAI-Easy hands-on style and any direct edit) bypass the spawn gate**, so a foreign active session goes undetected and a concurrent `git add -A && commit` can sweep the orchestrator's uncommitted work into another session's commit (observed in a production incident: staged files absorbed into a parallel session's auto-merged commit, landing on `main` entangled). To close that gap, the orchestrator MUST run the same parallel-session detection **before a non-trivial direct edit** to shared paths, not only before a write-agent spawn.

**"Non-trivial direct edit" trigger** — an Edit/Write/Bash mutation touching shared paths another session could also mutate: `.claude/`, `.moai/`, `internal/`, `pkg/`, `cmd/`, or repo-root config. Exempt: edits under an already-isolated worktree, `/tmp`, or a session-private scratch dir.

**Procedure** — before the first such edit of a task (and re-run if the task spans a long wall-clock window where a session may have started):
```bash
# 1. active foreign sessions (own session filtered out)
moai session list --json | jq '[.[] | select(.cwd == "<project-root>" and .session_id != "<own>")] | length'
# 2. divergence vs origin/main
git fetch origin main 2>&1; git rev-list --count --left-right origin/main...HEAD
```
Interpretation reuses the Pre-Spawn Sync Check matrix: `0 N` / `0 0` → proceed; `N 0` / `N M` → STOP (origin ahead or diverged — parallel session race); ≥1 foreign active session → the orchestrator SHALL isolate to a worktree (`moai cc -w` / `EnterWorktree` / `Agent(isolation: "worktree")`) OR surface via `AskUserQuestion` (isolate / wait / abort) before editing the shared tree. The conservative predicate (ANY foreign entry ⇒ isolate) is reused from the auto-isolation procedure in `.claude/rules/moai/workflow/worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation.

> **Stale-registry caveat**: the active-sessions registry can hold dead-PID entries (the session registry is not a reliable emptiness signal). A foreign entry's liveness MAY be probed with `kill -0 <pid>`; a confirmed-dead entry is ignored. When liveness is indeterminate, the conservative predicate isolates anyway (false positive is cheap; false negative corrupts the tree).

This is **procedural enforcement** (same trust model as the Pre-Spawn Sync Check). A mechanical PreToolUse-on-Edit advisory hook is the companion nudge (evaluated separately); the procedure is the load-bearing rule.

**Ambient signal.** The SessionStart hook already provides ambient foreign-session awareness: `internal/hook/session_start.go` Step 3 reads the registry (`session.QueryActiveWork`) and emits a `<system-reminder>` listing foreign active sessions via `session.FormatStderrReminder`. The session-start ambient signal is therefore satisfied without additional code; the PreToolUse-on-Edit advisory above adds per-edit awareness on top of that session-start baseline.

**Gate for orchestrator-direct Tier S/M push+PR.** The Pre-Spawn Sync Check (§ above) and the Pre-Edit Sync Check (§ above) are the mandatory gate for orchestrator-direct Tier S/M push+PR. When the orchestrator handles a Tier S/M push+PR directly via Bash (`git switch -c` / `git push -u` / `gh pr create` with `MOAI_BRANCH_GUARD_EXEMPT=1`) instead of spawning `manager-git`, the env-path branch-guard exemption is predicated on these sync checks passing first: the orchestrator MUST run the `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` + `moai session list --json` batch and halt (or auto-isolate per `worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation) on any divergence or foreign active session BEFORE issuing the branch-state-mutating Bash calls. The exemption sentinel admits the branch mutation mechanically; the sync check is what makes that admission safe.

## Time Estimation

[ZONE:Evolvable] [HARD] Never use time predictions in plans or reports.
- Use priority labels: Priority High / Medium / Low
- Use phase ordering: "Complete A, then start B"
- Prohibited: "2-3 days", "1 week", "as soon as possible"
