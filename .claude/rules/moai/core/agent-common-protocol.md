---
description: Shared protocol auto-loaded for all MoAI agents — user-interaction boundary, ledger closure, verification batching. Intentionally always-loaded (no paths restriction).
---

# Agent Common Protocol

Shared protocol for all MoAI agent definitions. This rule is automatically loaded for all agents, eliminating the need to duplicate these sections in each agent body.

> **Detail companion**: `agent-common-protocol-reference.md` (paths-scoped to this file) — verbatim verification batch, output contracts, CLI idioms, Ledger Closure clause bodies, attributable diff-check detail, sync-check rationale + incident records. Read it when composing a verification batch or handling an aborted delegation.

## User Interaction Boundary

`AskUserQuestion` is the **only** user-facing question channel. The boundary is asymmetric by design.

### Subagent Prohibitions

[ZONE:Frozen] [HARD] Subagents MUST NOT prompt the user. AskUserQuestion is reserved exclusively for the MoAI orchestrator.

Rules for subagents:
- If required context is missing, return a blocker report to the orchestrator — do not output free-form questions
- Never surface AskUserQuestion calls from within a subagent prompt body
- All user preferences must arrive via the orchestrator's spawn prompt
- If the orchestrator omitted critical data, respond with a structured "missing inputs" section and stop

Rationale (1 line): subagents run in isolated, stateless contexts — prompting there is a dead channel; this preserves the orchestrator's single-point-of-contact with the user (CLAUDE.md §8).

### Orchestrator Obligations

> Canonical: see `.claude/rules/moai/core/askuser-protocol.md` § Orchestrator Obligations for the full preload sequence (`ToolSearch(query: "select:AskUserQuestion")` before each call), the AskUserQuestion channel monopoly, the Socratic interview structure, and the option-description standards. This file owns only the subagent-side boundary (above) and the blocker-report → re-delegation flow (below).

The MoAI orchestrator collects all user preferences before delegating to subagents via `Agent()`. On receiving a blocker report from a subagent, it runs an `AskUserQuestion` round, injects the user's responses into a fresh subagent prompt, and re-delegates (procedure below).

### Hook Invocation Surface

The orchestrator interacts with three hook scripts that mechanically enforce orchestrator-discipline obligations (exit-code semantics: stdout JSON is honored only on exit 0 — on exit 2 it is discarded and only stderr is surfaced):

| Hook script | Trigger | Exit-code semantics |
|-------------|---------|---------------------|
| `.claude/hooks/moai/status-transition-ownership.sh` | PostToolUse on Write/Edit of `.moai/specs/SPEC-*/{spec,plan,acceptance}.md` body | exit 0 always (advisory; audit-logged to `.moai/logs/status-transition-audit.log`; exit-2 blocking reserved for future enforcement) |
| `.claude/hooks/moai/sync-phase-quality-gate.sh` | Stop hook on sync-phase commit completion | exit 0 always; failing check emits advisory `systemMessage`; blocking mode (opt-in `MOAI_SYNC_GATE_BLOCKING=1`) emits stdout JSON `{"decision":"block"}` |
| `.claude/hooks/moai/team-ac-verify.sh` | TaskCompleted in team mode (dormant — harness `thorough` + team prerequisites) | exit 0 always; rejection via stdout JSON `{"continue":false,"stopReason":...,"ledger_note":...}` (`decision` NOT valid for TaskCompleted) |

Full per-row owning-policy detail and the hook subagent-boundary acceptance criterion (grep verifying no hook invokes AskUserQuestion): `agent-common-protocol-reference.md` § Hook Invocation Surface detail.

#### Orchestrator translation responsibility

Hooks return exit codes and structured JSON; they MUST NOT invoke `AskUserQuestion` directly. When a hook signals a block (stdout JSON `"decision":"block"` on exit 0, or a legacy exit-2), the orchestrator MUST:

1. Parse the hook's structured JSON output (`decision`, `reason`, plus optional `ledger_note` / `systemMessage` / `details`)
2. Preload `AskUserQuestion` via `ToolSearch(query: "select:AskUserQuestion")`
3. Compose an `AskUserQuestion` round presenting the user with at least: (a) accept the block and address the failed gate, (b) override with `--skip-hook` opt-out (logged to `.moai/logs/hook-skip.log`), (c) abort the workflow

#### Stop self-gate caveat

The Stop hook fires on every turn-end, not only on task completion, so it must self-gate (inspect state, decide whether the turn is a genuine completion point, otherwise exit 0). It does NOT fire on user interrupt — it is not a guaranteed end-of-work signal.

#### Recovery-Signal Carve-Out

[ZONE:Evolvable] **While** a turn's `stopReason` or surrounding context indicates the turn is itself a **recovery signal** (recovering from a sync failure, a compact, a `prompt_too_long` (PTL), a `max_output_tokens` exhaustion, or a `media_size` / `compact-failure`), Stop/PostToolUse hooks SHOULD exit 0 rather than exit 2, so recovery turns are NOT placed into the `error → stop-hook-blocks → retry → error` **death-spiral** loop. SHOULD (policy guidance), not a mechanical gate. SSOT: `runtime-recovery-doctrine.md` §4.

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

The **ledger-closure invariant** (grounded in `github.com/wquguru/harness-books` book1 ch04 "账本闭环"): an aborted `Agent()` delegation MUST NOT leave a **dangling tool_use** — an open promise with no matching result — in the orchestrator's own context. This is the in-session interrupt analogue of the model-API rule that every `tool_use` receives a `tool_result` (persistence-layer analogue: `session-handoff.md` Block 3-4 preconditions).

[ZONE:Evolvable] [HARD] The orchestrator MUST close the ledger on any aborted delegation. Four clauses bind this obligation (full clause bodies + grounding: `agent-common-protocol-reference.md` § Ledger Closure clause bodies):

- **(a) Synthetic result on aborted Agent() delegation** — on abort (user interrupt, parent-abort propagation, or timeout), emit a short prose **synthetic ledger-closing artifact** (what was delegated, that it did not return, abort reason) into the orchestrator's own context before the next delegation. A blocker report is a *return*, not an *abort* — this clause covers only no-return-at-all.
- **(b) team-ac-verify.sh reject-path `ledger_note`** — on a TaskCompleted rejection, inject the hook's `ledger_note` as the ledger-closing artifact for that task.
- **(c) TeammateIdle exit-2 task closure** — a task rejected by TeammateIdle MUST NOT be left open without a reassignment owner (re-assign via new teammate, refined re-delegation, or close-as-obsolete with a synthetic closing note).
- **(d) Truthfulness** — the artifact MUST be a real summary, not a fabricated "success" (`verification-claim-integrity.md` §1.1 surface 1).

**Scope-boundary note.** Ledger Closure is a sibling of (not nested in) Hook Invocation Surface under the User Interaction Boundary H2.

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

MoAI does not provision MCP servers; use WebSearch and WebFetch to look up library documentation and established best-practice patterns: (1) WebSearch with targeted queries → (2) WebFetch to verify each URL and read the official documentation → (3) deliver established best-practice patterns → (4) continue work — architecture/analysis quality must not depend on MCP availability.

GLM-backend routing: when the session runs on the GLM backend (`moai glm` or the GLM teammate panes of `moai cg`), web search / web fetch / image read route to the z.ai MCP tools instead of the built-in `WebSearch` / `WebFetch` / `Read`. HARD routing table: `.claude/rules/moai/core/glm-web-tooling.md`.

## CLAUDE.md Reference

Agents follow MoAI's core execution directives defined in CLAUDE.md (auto-loaded, no restating needed).

## Agent Invocation Pattern

[ZONE:Evolvable] [HARD] Agents are invoked through MoAI's natural language delegation pattern ("Use the {agent-name} subagent to {task description}") — natural language conveys full context including constraints, dependencies, and rationale.

### Per-Spawn Model Injection

[ZONE:Evolvable] [HARD] When spawning a subagent, pass the model the active profile resolves for that agent as an explicit `model` argument on the spawn. (Why omitting is not neutral, and the full profile matrix: `agent-common-protocol-reference.md` § Per-Spawn Model Injection rationale; policy SSOT `.claude/rules/moai/development/model-policy.md`.)

- Resolve the value with `moai model profile --json` (reports the `{model, effort}` cell per retained agent under the active profile)
- Pass `model` per spawn. `effort` has no spawn-time parameter — it travels only through the agent file's frontmatter
- A spawn whose declared model differs from the resolved one is drift, not an override — change the profile instead
- Agents outside the retained catalog resolve to the inherit sentinel and take no injection

A PreToolUse hook observes every spawn and records the outcome to `.moai/logs/agent-model-audit.jsonl` (advisory; blocking is opt-in via `workflow.agent_model_guard.enabled`, refusing only a declared-vs-resolved conflict).

## Background Agent Execution

[ZONE:Evolvable] [HARD] As of Claude Code v2.1.198, subagents run in the background by **default**; the runtime chooses foreground only when it needs the result. The default changes *where* a subagent runs, not *what* it may do — a background subagent still surfaces every permission prompt in the main session (since v2.1.186 the prompt names the asking subagent). MoAI aligns with this default and does not set the `background:` frontmatter field. (Runtime-history rationale: `agent-common-protocol-reference.md` § Background Agent Execution rationale.)

The retained safeguard is **concurrency, not backgrounding** — it targets the actual hazard, a file-write race between agents:

- **Read-only tasks** (research, analysis, review): safe in the background; while one is in flight the orchestrator continues independent read-only work
- **Write tasks**: the runtime chooses foreground or background; the permission prompt surfaces either way — do not force the mode via `background:`
- **Concurrency**: never run two write-capable agents at once; orchestrator work concurrent with a write-capable agent stays read-only
- **Pre-approved writes**: add path patterns to settings.json `permissions.allow` to reduce prompts

## Tool Usage Guidelines

[ZONE:Evolvable] [HARD] Agents must follow tool usage patterns optimized for accuracy and efficiency.

### File Operations Pattern

- ALWAYS Read a file before using Edit on it
- Use Grep to locate specific line numbers before targeted Read with offset/limit; use Glob to discover files before reading — never guess file paths
- Prefer Edit over Write for existing files (sends only the diff, preserves context)
- Use absolute paths for all file operations; never construct paths from assumptions — verify with Glob or Bash `ls` first
- In worktrees, use project-root-relative paths for write targets

### Search Pattern

Progressive narrowing: (1) Glob by pattern → (2) Grep `files_with_matches` → (3) Grep `content` mode + context lines → (4) Read with offset/limit. Avoid reading entire large files when one section suffices; avoid Bash grep/find when Grep/Glob are available; filter by file type when the target language is known.

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

**MCP-over-CLI preference**: where an `mcp__moai__*` tool exists for a capability in the agent's `tools:` list, prefer it over the equivalent Bash CLI — same implementation, structured output, no shell-quoting hazards. Full catalogue: `.claude/rules/moai/core/moai-mcp-tools.md`.

### Bash Timeout

The Bash tool supports a `timeout` parameter (milliseconds): default 120,000ms, max 600,000ms. Set it for long-running commands (builds, test suites, installs).

### Error Recovery Pattern

When a tool call fails:
1. Read the error message carefully — diagnose root cause
2. Verify assumptions: does the file/path exist? (Glob check)
3. Try an alternative approach — do not retry the identical call
4. After 3 failures on the same operation, report the blocker

**Retry safety is asymmetric with respect to side effects.** Idempotent / read-only calls may be retried up to the ceiling. **Side-effecting calls** (write/edit, commit, push, PR, deploy, external-API mutation) that fail *ambiguously* require observing the current state first and retrying only when the effect is confirmed absent — a blind retry risks a duplicate commit / PR / deploy. The absence of a success signal is not evidence the effect did not land. (Full worked detail: `agent-common-protocol-reference.md` § Error Recovery retry-safety detail.)

### Super-Advisor Escalation (E1-E4)

When recovery via the 3-retry ceiling is insufficient OR a higher-reasoning consultation is warranted, the orchestrator escalates to the **super-advisor** agent. super-advisor returns **non-binding prescriptions**; the orchestrator remains the decision owner. DISTINCT from auditor verdicts — `plan-auditor` / `sync-auditor` own binding PASS/FAIL judgment; "should this PASS?" → an auditor, "what should I do here?" → super-advisor.

Entry conditions (exhaustive):

| Trigger | Condition | Example |
|---------|-----------|---------|
| **E1 — bug-deadlock** | 3+ consecutive same-diagnostic failures | same failing test retried 3 times with the same root-cause hypothesis |
| **E2 — architecture/design decision point** | A spec-body or plan-body decision with ≥2 viable options, neither obviously correct | "write-through or write-behind?" at L-plan boundary |
| **E3 — second-opinion request** | Orchestrator uncertainty: < 80% confidence in the next delegation step | ambiguous blocker-report; re-spawn vs user-escalation |
| **E4 — loop-deadlock** | `/moai loop` or `/moai fix` ceiling-exit per the loop-verdict contract | auto-fix iterations exhausted without green CI |

On trigger: spawn `Agent(general-purpose)` with the super-advisor role profile (Opus + xhigh at max/medium tier; Sonnet + xhigh at low tier — GLM-backed sessions fall back to the session model), receive the prescription, then re-seed the executor or escalate to the user via `AskUserQuestion`. Agent file: `.claude/agents/moai/super-advisor.md`.

## Parallel Execution

[ZONE:Evolvable] [HARD] The orchestrator MUST execute every read-only verification batch as a single-turn multi-Bash call. Serial verification across turns wastes wall-time and is the single largest source of run-phase latency (a prior meta-analysis: 10 min serial verification ≈ 11% of total run-phase wall-time).

### Read-only verification batching

When the orchestrator needs to verify implementation completion, it SHOULD issue multiple Bash tool calls within a single response turn. Independent verifications that do not share state are safe to parallelize.

### Verbatim batch, output contracts, and CLI idioms

The canonical 7-command batch, the file-redirect contract, the evidence-persistence obligation, the serial-verification anti-pattern, and the CLI idiom catalogue (`gh pr checks --json … | jq`, `--watch --fail-fast` in background mode, `git log --format=…`, the per-turn `ToolSearch` preload) all live in `agent-common-protocol-reference.md`. Read it when composing a batch.

Three obligations from that file bind here and are restated so they hold without it:

- **Batch in one turn.** Independent read-only verifications are issued as separate Bash tool calls within a single assistant turn — never serialized across turns. Serialize only for genuine dependencies: one command's output feeding another, writes to the same path, or shared-state mutation.
- **File-redirect contract.** When a command's verbatim output exceeds the bounded-tail ceiling (default: 50 lines or 2KB, whichever is smaller), redirect it to a file and surface only the exit code plus a bounded tail. Below the ceiling, inline quotation is fine. The contract removes the double-burn of quoting output twice, never the evidence itself.
- **Evidence persistence.** The cited path must still resolve at audit time, so evidence is persisted under `.moai/state/verify/<session>/` rather than left in `/tmp`, which the OS clears. A claim whose cited evidence path no longer resolves is an unattributed claim (`verification-claim-integrity.md` §2).

### Attributable diff-check doctrinal switch

Default-inversion switch in how the orchestrator COMPOSES the canonical batch: consult the shared diagnostic snapshot via `moai verify check --key-current` (keyed by HEAD SHA) BEFORE re-executing; on all-three attribution match, consume the attributable §E evidence for that dimension INSTEAD of re-executing the corresponding command. This is a composition-time doctrinal switch binding the orchestrator's batch-composition discipline, not a runtime hook. (Full mechanism + fallback contract: `agent-common-protocol-reference.md` § Attributable diff-check detail; pattern file `.claude/rules/moai/workflow/verification-batch-pattern.md`.)

**All-three attribution match → CONSUMES the attributable §E evidence (no re-execution) [DEFAULT].** When ALL THREE hold for a verification dimension:
1. **Snapshot key match** — the §E-cited HEAD SHA equals the current `moai verify check --key-current` key
2. **Command match** — the §E-cited command matches the snapshot's recorded command
3. **Output match** — the §E-cited observed output matches the snapshot's recorded output

the batch records the snapshot key + cited §E evidence path as its baseline-attribution per VCI §2 and DOES NOT re-execute that dimension (marked PASS-attributed, not PASS-reexecuted).

**Any mismatch → fallback to re-execution.** On ANY of `snapshot_key_drift` (HEAD SHA changed) / `command_drift` / `missing_section_e` (§E evidence missing or citing no observable output) / `output_drift`, the batch SHALL re-execute the affected dimension — any-mismatch → re-execute, never silent skip. The fallback is logged with the mismatch reason; the VCI §1.1 invariant holds on every path.

### Pre-Spawn Sync Check (Multi-Session Race Mitigation)

[ZONE:Evolvable] [HARD] Before spawning any implementation `Agent()` (manager-develop / manager-docs / per-spawn `Agent(general-purpose)` with a domain whitelist) that will commit or modify shared working-tree files, the orchestrator MUST execute the following parallel batch and surface any divergence to the user.

```bash
# 1. Fetch latest origin/main without merging
git fetch origin main 2>&1

# 2. Count divergence between local HEAD and origin/main
git rev-list --count --left-right origin/main...HEAD

# 3. Query active sessions on this host for the same SPEC scope (L1 of the
#    canonical 4-layer multi-session race mitigation policy).
moai session list --json --filter-spec=<SPEC-ID>
```

Interpretation matrix (git divergence):

| Output | Meaning | Action |
|--------|---------|--------|
| `0 N` | Local ahead by N (clean — your commits not yet pushed) | Proceed normally |
| `0 0` | Synced (local == origin/main) | Proceed normally |
| `N 0` | Origin ahead by N — **parallel session race detected** | STOP, surface via AskUserQuestion: rebase / inspect / abort |
| `N M` | Diverged (both ahead) | STOP, MUST resolve before spawn |

Interpretation matrix (active-sessions query):

| Output | Meaning | Action |
|--------|---------|--------|
| `[]` | No other session on this SPEC | Proceed normally |
| `[{...}]` (≥1 entry from another session) | **Concurrent session race detected on same SPEC** | STOP, surface entries, AskUserQuestion: **wait** / **override** / **abort** |

The 3rd command is additive only (the original 2-command batch is preserved verbatim). Sessions predating the registry hook emit no entries — `[]` — no false positives. Rationale + the originating sync-phase race incident record: `agent-common-protocol-reference.md` § Pre-Spawn Sync Check rationale and incident record.

Exemption: read-only agents (`Explore`, or a per-spawn `Agent(general-purpose)` scoped to read-only investigation) do not require pre-spawn fetch — they cannot trigger race conflicts.

> **Spawn-gate boundary**: this check fires only at the write-agent spawn boundary. Direct main-session edits bypass this gate — see § Pre-Edit Sync Check below. Defense-in-depth policy: `.moai/docs/generic-patterns-guide.md` § Multi-Session Race Mitigation Procedure; worktree-as-race-elimination: `session-handoff.md` § Worktree-Anchored Resume Pattern.

### Pre-Edit Sync Check (Direct-Edit Race Mitigation)

[ZONE:Evolvable] [HARD] Direct main-session edits to shared working-tree paths (Edit/Write/Bash — any direct edit) bypass the spawn gate above, so the orchestrator MUST run the parallel-session detection **before a non-trivial direct edit** to shared paths. (Incident record + enforcement-placement assessment: `agent-common-protocol-reference.md` § Pre-Edit Sync Check — rationale and enforcement record.)

#### The rule, at the moment of the edit

**TRIGGER** — the gate fires when ALL three hold:

| Condition | Test |
|---|---|
| Tool | an `Edit`, `Write`, or file-mutating `Bash` call |
| Target | a shared path another session could also mutate: `.claude/`, `.moai/`, `internal/`, `pkg/`, `cmd/`, or repo-root config files |
| Location | CWD is the primary checkout. Exempt: an already-isolated worktree, `/tmp`, or a session-private scratch dir |

**CHECK** — before the FIRST triggered edit of a task, as one parallel batch:
```bash
# 1. live foreign sessions (own session filtered out; then liveness-probe each PID)
moai session list --json | jq '[.[] | select(.cwd == "<project-root>" and .session_id != "<own>")] | length'
# 2. divergence vs origin/main
git fetch origin main 2>&1; git rev-list --count --left-right origin/main...HEAD
```

**DECIDE and ACT** — no outcome permits "proceed in the shared checkout anyway":

| Probe result | Required action |
|---|---|
| 0 live foreign sessions AND `0 0` / `0 N` | Proceed in the shared checkout |
| ≥1 live foreign session | **ISOLATE before editing**: `moai cc -w <name>` / `EnterWorktree(<path>)` / `Agent(isolation: "worktree")`. If isolation is impossible, surface via `AskUserQuestion` (isolate / wait / abort) |
| `N 0` / `N M` divergence | STOP; `AskUserQuestion` (rebase / inspect / abort) per the Pre-Spawn Sync Check matrix |

> **Stale-registry caveat**: registry entries can hold dead PIDs. Probe each foreign entry's liveness with `kill -0 <pid>`; ignore confirmed-dead, treat indeterminate as live and isolate anyway. ANY live-or-indeterminate foreign entry ⇒ isolate (`worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation).

**RE-CHECK** — the probe decays. Re-run it before ANY commit in the shared checkout, and after any long pause in the task.

#### The sweep prohibition

[ZONE:Evolvable] [HARD] In the primary checkout, NEVER `git add -A`, `git add .`, or `git commit -a`. Stage by explicit pathspec (`git add <path> …`), and re-read `git status --short` immediately before staging so another session's files are visible and excluded. This applies **even when the pre-edit probe found no foreign session** — a session can arrive after the probe, and the sweep is what turns its presence into lost work.

**Ambient signal.** The SessionStart hook already lists foreign active sessions via a `<system-reminder>` (`internal/hook/session_start.go` Step 3) — the always-on detection layer; this check is the decision layer that turns detection into isolation.

## Time Estimation

[ZONE:Evolvable] [HARD] Never use time predictions in plans or reports.
- Use priority labels: Priority High / Medium / Low
- Use phase ordering: "Complete A, then start B"
- Prohibited: "2-3 days", "1 week", "as soon as possible"
