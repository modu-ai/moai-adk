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
| `.claude/hooks/moai/team-ac-verify.sh` | TaskCompleted in team mode (dormant by default — activates only under harness `thorough` + team mode prerequisites per the canonical team activation policy) | per-AC PASS evidence file verification | exit 0 always; a completion rejection is signaled via stdout JSON `{"decision":"block","reason":...,"ledger_note":...}` (the block decision MUST ride the exit-0 stdout channel — exit 2 would discard it) |

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

- **(a) Synthetic result on aborted Agent() delegation.** When an `Agent()` delegation is aborted — user interrupt (Ctrl+C), parent-abort propagation (the orchestrator's own turn was aborted and the sub-agent was killed), or timeout (no return before a wall-clock or token-budget ceiling) — the orchestrator SHALL emit a **synthetic ledger-closing artifact** into its own context before issuing the next delegation. The artifact is a short prose summary (NOT a structured data record; no JSON schema, no `.moai/state/ledger.json`), naming what was delegated, that it did not return, and the abort reason if known. Its purpose is to close the open promise so the next turn does not proceed as if the delegation returned cleanly. This clause does NOT change the "Missing Inputs" blocker-report pattern above: a blocker report is a *return*, not an *abort*; clause (a) covers only the case where no return is produced at all.
- **(b) team-ac-verify.sh reject-path `ledger_note` field.** When `.claude/hooks/moai/team-ac-verify.sh` rejects a `TaskCompleted`, it signals the rejection via stdout JSON `{"decision":"block","reason":...,"ledger_note":...}` and exits 0 — per Claude Code hook semantics, stdout JSON is honored only on exit 0 (on exit 2 stdout is discarded and only stderr is surfaced), so the block decision and its `ledger_note` MUST ride the exit-0 stdout channel. The orchestrator injects this `ledger_note` as the ledger-closing artifact for that task. (The reject-path trigger itself is a minimal stub; full AC-verification logic is out of scope and deferred to a follow-up SPEC.)
- **(c) TeammateIdle exit-2 task closure.** When the TeammateIdle hook rejects a task's completion via exit-2 ("keep working"), the rejected task's TaskList entry MUST NOT be left in an open state without a reassignment owner. The orchestrator re-assigns the task (spawn a new teammate, re-delegate to the same teammate with a refined prompt, or close it as obsolete with a synthetic closing note). This binds the orchestrator's TaskList hygiene, not the hook's exit-2 emission. The parent-abort propagation that book1 ch07 names — cleanup handlers registered to avoid orphan tasks — is the source for this clause.
- **(d) Cross-references.** This subsection cross-references three sources:
  - **book1 ch04** (账本闭环 — the ledger-closure invariant named in the opening paragraph above).
  - **book1 ch07** (parent-abort propagates to forked children; agents are observable lifecycle objects via SubagentStart/SubagentStop hooks, exit-code-2 stderr feedback).
  - `.claude/rules/moai/workflow/session-handoff.md` Block 3-4 preconditions (the persistence-layer analogue of ledger closure across `/clear`).
  - The ledger-closing artifact's truthfulness is bound by `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report) — the artifact MUST be a real summary, not a fabricated "success".

**Scope-boundary note.** This Ledger Closure subsection is distinct from the Hook Invocation Surface subsection above (owned by the sibling runtime-recovery SPEC's Recovery-Signal Carve-Out). The two are siblings under the User Interaction Boundary H2; Ledger Closure is NOT nested inside Hook Invocation Surface. See the sibling interrupt-ledger SPEC for the collision-free placement contract.

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

When Context7 MCP is unavailable:
1. Detect unavailability immediately when MCP tools fail or return errors
2. Inform user that Context7 is unavailable and provide alternative approach
3. Use WebFetch to access official documentation as fallback
4. Deliver established best practice patterns based on industry experience
5. Continue work — architecture/analysis quality must not depend on MCP availability

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

[ZONE:Frozen] [HARD] Background subagents (`run_in_background: true`) MUST NOT perform Write/Edit operations.

As of Claude Code v2.1.186, when a background subagent reaches a tool call that needs permission, the prompt surfaces in the main session and names the asking subagent (Esc denies just that one call). Before v2.1.186, background subagents auto-denied any prompting tool call — the prior basis for this rule. MoAI nonetheless keeps `run_in_background: false` for write tasks as a conservative default: in standard permission mode each background write raises a main-session permission prompt that interrupts the leader's flow and undercuts the parallelism benefit of backgrounding, whereas foreground execution keeps write-permission flow deterministic. Read-only tasks (research, analysis, review) remain safe and efficient in the background.

Rules for agent spawning:
- **Read-only tasks** (research, analysis, review): `run_in_background: true` is safe
- **Write tasks** (implementation, refactoring, file creation): `run_in_background: false` required
- **Parallel writes needed**: Process directly from the main session, or use sequential foreground agents
- **Pre-approved writes**: Add path patterns to settings.json `permissions.allow` for background write support

Decision matrix:
- Agent reads files only → `run_in_background: true` (parallel, fast)
- Agent writes files → `run_in_background: false` (sequential, reliable)
- Multiple agents need to write different files → Use main session directly or foreground agents in sequence

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

## Parallel Execution

[ZONE:Evolvable] [HARD] The orchestrator MUST execute every read-only verification
batch as a single-turn multi-Bash call. Serial verification across turns wastes
wall-time and is the single largest source of run-phase latency (empirical
analysis showed 10 min serial verification ≈ 11% of total run-phase wall-time). This rule was
added in response to that finding.

### Read-only verification batching

When the orchestrator needs to verify implementation completion, it SHOULD issue
multiple Bash tool calls within a single response turn. Independent verifications
that do not share state are safe to parallelize.

### Canonical 7-item example

The following 7 verification commands cover the standard read-only verification
batch for a typical run-phase completion. The orchestrator SHOULD invoke all 7
in parallel within a single response turn:

```bash
# 1. Full test suite (Go)
go test ./... > /tmp/moai-verify/1-go-test.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/1-go-test.log

# 2. Coverage report (per-package)
go test -coverprofile=cover.out ./internal/<pkg>/... > /tmp/moai-verify/2-cover.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/2-cover.log

# 3. Subagent-boundary grep (sentinel C-HRA-008)
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//" > /tmp/moai-verify/3-boundary.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/3-boundary.log

# 4. Sentinel key audit (build-tag, retired SPEC, etc.)
grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN' internal/ | head -20 > /tmp/moai-verify/4-sentinel.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/4-sentinel.log

# 5. CLI smoke check (cmd/moai)
go run ./cmd/moai --version > /tmp/moai-verify/5-cli.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/5-cli.log

# 6. Benchmark micro-suite (optional)
go test -bench=. -benchmem -run=^$ ./internal/<pkg>/... > /tmp/moai-verify/6-bench.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/6-bench.log

# 7. Lint baseline (golangci-lint)
golangci-lint run --timeout=2m > /tmp/moai-verify/7-lint.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/7-lint.log
```

In Claude's response, all 7 commands are invoked as separate Bash tool calls
within the same assistant turn. The orchestrator does NOT issue them serially
across multiple turns.

### File-redirect contract

The canonical batch above also demonstrates the **file-redirect contract**: when a verification command's verbatim output exceeds the **bounded-tail ceiling** (concrete default: **≤50 lines OR ≤2KB, whichever is smaller**), the orchestrator redirects the verbatim output to a file on disk and surfaces only **exit code + bounded-tail summary** in conversation context. Each command above shows the redirected form (`> /tmp/moai-verify/<N>-<slug>.log 2>&1; echo "exit=$?"; tail -50 …`).

This contract governs *how* verification output is represented in context, NOT *whether* the commands run in parallel — the single-turn multi-Bash HARD obligation above is unchanged. The cited file path MUST appear in the Verification Matrix / Completion Report banner (`.claude/output-styles/moai/moai.md` §8) or in the manager-agent `§E` self-verification block, so the verbatim evidence remains reachable at audit time. This preserves `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 **surface 1** (orchestrator self-report) and **surface 2** (manager-agent `§E` self-verification): every claim row remains attributable to a directly-observed command whose verbatim output is reachable at the cited file path.

The contract is **"verbatim evidence lives on disk with a citable path; context carries exit code + bounded tail"** — NOT **"drop the evidence"**. Inline quotation is PERMITTED when verbatim output is below the ceiling (the redirect obligation triggers only on exceedance); the diet removes the *double-burn* (Bash inline output + banner re-quote), not the evidence itself. The exact ceiling value and directory scheme are tunable per-domain; the contract holds regardless of the specific numbers.

### Evidence persistence obligation

The cited evidence path MUST remain reachable at audit time, including after `/tmp` directory clearance. `/tmp` is OS-cleared periodically (macOS reboot, Linux tmpfs re-mount, systemd-tmpfiles); a cited path that no longer resolves to a file violates `verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report) and surface 2 (manager-agent §E self-verification) — every claim row MUST remain attributable to a directly-observed command whose verbatim output is reachable at the cited file path.

To satisfy this reachability obligation, evidence SHALL be persisted under `.moai/state/verify/<session>/` (gitignored runtime state, same directory family as `context-usage.json` and `active-sessions.json`). The exact persist mechanism — direct write to `.moai/state/verify/<session>/`, or `/tmp` write followed by a copy step — is a run-phase implementation detail; the contract states the OBLIGATION (evidence survives `/tmp` clearance), not the mechanism. **"Persist evidence" ≠ "drop evidence"**: the diet removes the *double-burn* (inline output + banner re-quote), NOT the evidence itself. The verbatim output MUST remain on disk at a citable, audit-time-reachable path.

### Anti-pattern: serial verification across turns

```
Turn 1: go test ./...     → wait for completion → Turn 1 ends
Turn 2: golangci-lint ... → wait for completion → Turn 2 ends
Turn 3: grep -rn ...      → wait for completion → Turn 3 ends
```

This pattern locks the orchestrator into N sequential turns where 1 turn would
suffice. Each turn adds round-trip latency. For 7 verifications averaging 2 s
each, serial execution adds ~14 s of dead-time per run-phase completion.

### When to use serial execution

- Commands that depend on each other (e.g., `make build` before `go test ./...`)
- Commands that write to the same file or directory
- Commands that mutate shared state (filesystem, env vars)

### Cross-reference

- The canonical verification-batch acceptance criterion (recorded in the
  predecessor workflow optimization rule) verifies this section contains the
  7 verification keywords (`go test`, `coverprofile`, `grep `, `sentinel`,
  `cmd/moai`, `bench`, `lint`).
- `.claude/rules/moai/workflow/verification-batch-pattern.md` documents the
  formal verification grouping pattern.

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

Cross-reference: `.moai/docs/generic-patterns-guide.md` § Multi-Session
Race Mitigation Procedure (defense-in-depth policy at user-facing
layer); `.claude/rules/moai/workflow/session-handoff.md` § Worktree-Anchored
Resume Pattern (L2/L3 worktree as race-elimination alternative).

## Tool Optimization Patterns

[ZONE:Evolvable] [HARD] Agents MUST use single-command idioms over multi-step
shell pipelines when a CLI tool provides structured output (JSON). The
canonical patterns below replace the prose alternatives that previously
expanded into multiple sequential commands.

### CI Status Query

```bash
# Canonical pattern — single command, structured JSON output.
gh pr checks <PR> --json name,state,conclusion | jq '.[] | select(.conclusion != "SUCCESS")'

# Why: single round-trip, parseable, easier to integrate with subsequent steps.
# Avoid: gh pr checks <PR> | grep -E 'FAIL|PENDING'  (string parsing, brittle)
```

### Recent Commit Inspection

```bash
# Canonical pattern — single command, structured.
git log --format='%h %s %ci' -10 | head -10

# Why: built-in format string avoids multi-step git log | awk pipelines.
# Avoid: git log --pretty=oneline | awk '{print $1}' | xargs git show
```

### ToolSearch Per-Turn Preload

```
ToolSearch(query: "select:AskUserQuestion,TaskCreate,TaskUpdate,TaskList,TaskGet", max_results: 5)
```

This canonical preload SHOULD be invoked at the start of every orchestrator
turn where deferred tools may be needed. See
`.claude/rules/moai/core/askuser-protocol.md` for the full preload contract.

### Cross-reference

- The canonical CI-status-query acceptance criterion (recorded in the
  predecessor workflow optimization rule) verifies this section contains
  `gh pr checks --json` and `jq` literals in proximity.

## Time Estimation

[ZONE:Evolvable] [HARD] Never use time predictions in plans or reports.
- Use priority labels: Priority High / Medium / Low
- Use phase ordering: "Complete A, then start B"
- Prohibited: "2-3 days", "1 week", "as soon as possible"
