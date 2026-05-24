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

The MoAI orchestrator MUST follow these obligations when using AskUserQuestion:

- [ZONE:Frozen] [HARD] The orchestrator MUST preload AskUserQuestion via `ToolSearch(query: "select:AskUserQuestion")` before each call — AskUserQuestion is a deferred tool and its schema is not loaded at session start
- [ZONE:Frozen] [HARD] All user-facing questions MUST go through AskUserQuestion — free-form prose questions in response text are prohibited
- Collect all user preferences before delegating to subagents via Agent()
- On receiving a blocker report from a subagent: run an AskUserQuestion round, inject the user's responses into a fresh subagent prompt, and re-delegate

Canonical reference: see `.claude/rules/moai/core/askuser-protocol.md` for full preload sequence, Socratic interview structure, and anti-pattern catalog.

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

## MCP Fallback Strategy

[ZONE:Evolvable] [HARD] Maintain effectiveness without MCP servers.

When Context7 MCP is unavailable:
1. Detect unavailability immediately when MCP tools fail or return errors
2. Inform user that Context7 is unavailable and provide alternative approach
3. Use WebFetch to access official documentation as fallback
4. Deliver established best practice patterns based on industry experience
5. Continue work — architecture/analysis quality must not depend on MCP availability

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

Background agents auto-deny all non-pre-approved permission prompts because they cannot interact with the user. Even with `mode: "bypassPermissions"`, the background execution context does not fully inherit the parent session's permission allowlist.

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

## Parallel Execution

[ZONE:Evolvable] [HARD] The orchestrator MUST execute every read-only verification
batch as a single-turn multi-Bash call. Serial verification across turns wastes
wall-time and is the single largest source of run-phase latency (W3 meta-analysis:
10 min serial verification ≈ 11% of total run-phase wall-time). This rule was
added by SPEC-V3R5-WORKFLOW-OPT-001 Layer D in response to that finding.

### Read-only verification batching

When the orchestrator needs to verify implementation completion, it SHOULD issue
multiple Bash tool calls within a single response turn. Independent verifications
that do not share state are safe to parallelize.

### Canonical 7-item example (AC-WO-007)

The following 7 verification commands cover the standard read-only verification
batch for a typical run-phase completion. The orchestrator SHOULD invoke all 7
in parallel within a single response turn:

```bash
# 1. Full test suite (Go)
go test ./...

# 2. Coverage report (per-package)
go test -coverprofile=cover.out ./internal/<pkg>/...

# 3. Subagent-boundary grep (sentinel C-HRA-008)
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"

# 4. Sentinel-key audit (build-tag, retired SPEC, etc.)
grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN' internal/ | head -20

# 5. CLI smoke check (cmd/moai)
go run ./cmd/moai --version

# 6. Benchmark micro-suite (optional)
go test -bench=. -benchmem -run=^$ ./internal/<pkg>/...

# 7. Lint baseline (golangci-lint)
golangci-lint run --timeout=2m
```

In Claude's response, all 7 commands are invoked as separate Bash tool calls
within the same assistant turn. The orchestrator does NOT issue them serially
across multiple turns.

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

- AC-WO-007 (SPEC-V3R5-WORKFLOW-OPT-001) verifies this section contains the
  7 verification keywords (`go test`, `coverprofile`, `grep `, `sentinel`,
  `cmd/moai`, `bench`, `lint`).
- `.claude/rules/moai/workflow/verification-batch-pattern.md` documents the
  formal verification grouping pattern.

### Pre-Spawn Sync Check (Multi-Session Race Mitigation)

[ZONE:Evolvable] [HARD] Before spawning any implementation `Agent()`
(manager-develop / manager-docs / expert-*) that will commit or modify
shared working-tree files, the orchestrator MUST execute the following
2-command parallel batch and surface any divergence to the user.

```bash
# 1. Fetch latest origin/main without merging
git fetch origin main 2>&1

# 2. Count divergence between local HEAD and origin/main
git rev-list --count --left-right origin/main...HEAD

# 3. Query active sessions on this host for the same SPEC scope (L1 of
#    SPEC-V3R6-MULTI-SESSION-COORD-001 4-layer race mitigation).
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
| `[]` | No other session on this SPEC (REQ-COORD-018) | Proceed normally |
| `[{...}]` (≥1 entry from another session) | **Concurrent session race detected on same SPEC** | STOP, surface entries via prose summary, AskUserQuestion: **wait** / **override** / **abort** |

The 3rd command is **additive only** (REQ-COORD-020) — the original
2-command batch (git fetch + git rev-list) is preserved verbatim above.
Backward compatibility: sessions running before
SPEC-V3R6-MULTI-SESSION-COORD-001 was deployed (no registry hook) emit
no entries, so the 3rd command returns `[]` and the orchestrator
proceeds normally without false positives.

Rationale: When 2+ Claude Code sessions operate on the same project root
+ same memory hash (`~/.claude/projects/{hash}/memory/`), they may both
consume the same paste-ready resume and attempt the same `/moai <subcommand>`
work. The git working tree is shared; the memory file is shared. Without
a pre-spawn fetch, the second session works on a stale baseline and may
produce duplicate commits, conflicting frontmatter edits, or CHANGELOG
entry races.

Origin: SPEC-V3R6-LEGACY-CLEANUP-001 sync-phase race (2026-05-23) —
parallel session committed `aea0cf7b9` (spec.md frontmatter status update)
between manager-develop M4 (`ccd1fa9cf`) and manager-docs sync
(`19bc873ff`). Detection occurred retrospectively when `git push` succeeded
with an unexpected intermediate commit in the push range. L9 reinforced
(parallel session race during long agent runs) + L44 NEW (pre-spawn fetch
discipline).

Exemption: read-only agents (`Explore`, `manager-quality` in diagnostic
mode) do not require pre-spawn fetch — they cannot trigger race conflicts.

Cross-reference: `CLAUDE.local.md` §23.8 Multi-Session Race Mitigation
(defense-in-depth policy at user-facing layer); `.claude/rules/moai/
workflow/session-handoff.md` § Worktree-Anchored Resume Pattern (L2/L3
worktree as race-elimination alternative).

## Tool Optimization Patterns

[ZONE:Evolvable] [HARD] Agents MUST use single-command idioms over multi-step
shell pipelines when a CLI tool provides structured output (JSON). The
canonical patterns below replace the prose alternatives that previously
expanded into multiple sequential commands.

### CI Status Query

```bash
# Canonical pattern (AC-WO-013) — single command, structured JSON output.
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

- AC-WO-013 (SPEC-V3R5-WORKFLOW-OPT-001) verifies this section contains
  `gh pr checks --json` and `jq` literals in proximity.

## Time Estimation

[ZONE:Evolvable] [HARD] Never use time predictions in plans or reports.
- Use priority labels: Priority High / Medium / Low
- Use phase ordering: "Complete A, then start B"
- Prohibited: "2-3 days", "1 week", "as soon as possible"
