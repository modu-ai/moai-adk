---
description: "Verbatim verification-batch example, output-representation contracts, and CLI idiom catalogue for the agent common protocol"
paths: "**/agent-common-protocol.md"
---

# Agent Common Protocol — Reference Detail

> Detail companion to `agent-common-protocol.md` (the SSOT). That file carries the
> binding obligations; this file carries the verbatim command batch, the worked
> contracts, and the CLI idiom catalogue. Loaded only when
> `agent-common-protocol.md` itself is being edited — read it directly when
> composing a verification batch.

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
# Linter set + default timeout governed by root .golangci.yml; the --timeout=2m flag here overrides it for the quick-check budget.
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


---

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

#### Waiting for checks to finish — `--watch`, run in the background

[ZONE:Evolvable] [HARD] The query above **samples** CI once. When the orchestrator instead needs to **wait** for checks to reach a terminal state, it MUST use `gh pr checks --watch`, and it MUST issue that command in the Bash tool's background mode. A hand-rolled `sleep`-and-poll loop is prohibited.

```bash
# Canonical wait pattern — issue with the Bash tool's background mode.
gh pr checks <PR> --watch --fail-fast
```

- `--watch` blocks until every check is terminal, so no polling interval has to be chosen or tuned.
- `--fail-fast` returns non-zero the moment any check fails, so the **exit code alone is the verdict** — no output parsing is needed to decide pass/fail.
- Background mode keeps the turn unblocked: the orchestrator continues independent read-only work and is re-invoked when the watch exits.

A manual polling loop burns one turn per iteration, hard-codes an interval that is simultaneously too slow for fast checks and too fast for slow ones, and re-implements — less reliably — a wait the CLI already provides. It also holds the turn open for the full CI duration, which the background `--watch` does not.

```bash
# Anti-pattern — manual polling re-implements --watch and burns a turn per iteration
for i in 1 2 3 4 5; do sleep 60; gh pr checks <PR>; done
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
- `.claude/rules/moai/workflow/cache-aware-execution.md` — prompt-cache-aware
  ordering (stagger-spawn for parallel same-type agents, gate placement,
  session-loaded file edit timing).
