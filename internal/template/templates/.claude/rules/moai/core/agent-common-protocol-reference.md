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

---

## Pre-Edit Sync Check — rationale and enforcement record

> Relocated verbatim from `agent-common-protocol.md` § Pre-Edit Sync Check (Direct-Edit Race Mitigation) to keep the always-loaded file within its size budget. The binding gate (TRIGGER / CHECK / DECIDE / RE-CHECK) and the sweep prohibition remain inline there; this section carries the incident record, the failure analysis of the previous version, and the enforcement-placement assessment.

### Incident record (why the rule is binding)

In the recorded incident, five Claude sessions worked the same shared primary checkout simultaneously. 104 uncommitted files accumulated there; **73 of them existed on no branch at all** and had to be rescued into a pull request by hand. The destructive mechanism is ordinary: a concurrent `git add -A && commit` in one session sweeps another session's uncommitted work into a commit that was never meant to carry it; a branch switch or stash strands it outright. A rule without the incident attached gets skipped; this one has the receipts.

### Why the previous version of this check failed

Two distinct failure modes — they need different fixes:

1. **Nothing executed it.** The check was procedural prose in this rule file. No hook fires it, no tool result reminds the agent of it, and a "run three probes before your first edit" instruction competes with task momentum — and loses. This was the dominant failure in the recorded incident: the procedure was directionally correct and was simply never run.
2. **It sampled at the wrong moment.** Even perfectly executed, a once-per-task check cannot see a session that starts mid-task, and the destruction mechanism is not the edit — it is the OTHER session's `git add -A` at commit time. The previous text placed all discipline on the editor and none on the sweeper.

The rewrite addresses both: the decision procedure is compressed into a moment-of-edit gate an agent can actually run, and a sweep prohibition binds the commit-side primitive that does the damage.

### Enforcement assessment — why there is no PreToolUse-on-Edit/Write advisory hook

A per-edit advisory hook was considered and declined:

- **Measured cost**: the branch-guard PreToolUse hook — the closest comparable, it spawns git subprocesses — averages 135-256ms per invocation. `Edit`/`Write` is the most frequent mutation in a coding session; a per-edit advisory multiplies that tax across every edit of every concurrent session — the exact multi-session load pattern that pushed the box over during the incident day.
- **Wrong sampling rate**: the probe's answer changes on the scale of minutes (sessions start and stop), not edits. Paying per-edit for a per-task answer taxes the wrong frequency.
- **Wrong target**: the edit is not the destructive primitive; the sweep is. If mechanical enforcement is ever wanted, the cheap and mechanism-adjacent placement is the commit-time Bash surface — extend the branch-guard hook family to warn on `git add -A` / `git add .` / `git commit -a` in the primary checkout (rare, once per task) — not on every Edit/Write.
- **Detection is already ambient**: the SessionStart signal carries foreign-session awareness to every session at zero per-edit cost; the procedure is the decision layer on top of it.

If a per-edit nudge is ever re-proposed, the only defensible variant is stateful: fire the probe on the FIRST `Edit`/`Write` of a session (or serve it from a short-TTL cache), never per edit.
