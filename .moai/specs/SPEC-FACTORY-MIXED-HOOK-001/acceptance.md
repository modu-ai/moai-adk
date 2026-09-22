---
id: SPEC-FACTORY-MIXED-HOOK-001
document: acceptance
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
card: t1074
module: "internal/factorymsg"
---

# Acceptance — SPEC-FACTORY-MIXED-HOOK-001

All criteria are MUST-PASS. A fixture proves only its named unit contract; the four live rows require real separate CLI/model contexts.

| AC | Requirements | Acceptance evidence |
|---|---|---|
| AC-FMH-001 | REQ-FMH-001, REQ-FMH-002, REQ-FMH-003 | Same repository from two linked worktrees resolves one canonical broker; another run/project cannot list, claim, read, or receipt it. |
| AC-FMH-002 | REQ-FMH-001, REQ-FMH-002, REQ-FMH-011 | Zero run returns `NO_ACTIVE_FACTORY`; many return `AMBIGUOUS_FACTORY`; explicit run works; concurrent worker claims have no duplicate slot; `--` tail is unchanged. |
| AC-FMH-003 | REQ-FMH-001, REQ-FMH-002 | Session UUID/generation/process-start mismatch, PID reuse, stale launcher env, and subagent inheritance cannot consume or acknowledge the registered mailbox. |
| AC-FMH-004 | REQ-FMH-004, REQ-FMH-005, REQ-FMH-006 | Closed kinds and all required envelope fields validate; retry deduplicates; stale generation/token ACK fails; lease expiry redelivers; expected revision prevents duplicate side effects. |
| AC-FMH-005 | REQ-FMH-005, REQ-FMH-006 | Crash before hook output, after output, before disposition, and before receipt preserves a recoverable pending/claim; only explicit receipt after persisted disposition acknowledges. |
| AC-FMH-006 | REQ-FMH-005, REQ-FMH-009 | TTL, poison, and overflow create bounded dead-letter/backpressure records with reason; legacy sessionmsg data and tests remain unchanged. |
| AC-FMH-007 | REQ-FMH-007, REQ-FMH-008 | Final UserPromptSubmit/Stop JSON contains at most 16 IDs/2 KiB of metadata, no raw malicious body; Stop continues at most once; Stop→UserPrompt does not duplicate a batch; other deny/stop wins. |
| AC-FMH-008 | REQ-FMH-008, REQ-FMH-010 | Empty inbox, receipt-only, permission wait, interrupt, disabled/untrusted hooks, and lock timeout create zero extra model turns and truthful pending/degraded capability state. |
| AC-FMH-009 | REQ-FMH-004, REQ-FMH-007, REQ-FMH-009 | Path traversal, fake run/session, unknown kind, prompt-injection payload, and unauthorized task revision are rejected; a normal authorized dispatch pointer succeeds without queue mutation. |
| AC-FMH-010 | REQ-FMH-012 | Real Codex lead↔Codex worker exchanges unique nonces both ways and observes generation-bound explicit receipts in separate CLI/model contexts. |
| AC-FMH-011 | REQ-FMH-012 | Real Codex lead↔Claude worker exchanges unique nonces both ways and observes generation-bound explicit receipts in separate CLI/model contexts. |
| AC-FMH-012 | REQ-FMH-012 | Real Claude lead↔Codex worker exchanges unique nonces both ways and observes generation-bound explicit receipts in separate CLI/model contexts. |
| AC-FMH-013 | REQ-FMH-009, REQ-FMH-012 | Real Claude↔Claude regression passes the same nonce/receipt contract; report-only traffic cannot mark a card complete without independent evidence. |
| AC-FMH-014 | REQ-FMH-010, REQ-FMH-012 | Message arriving after both sides become idle is reported pending-until-next-turn, not idle-wake; no private socket/tmux key injection or empty polling model turn occurs. |
| AC-FMH-015 | REQ-FMH-012 | Measured benchmark covers empty/16/1000, 1/10 sessions, warm/cold, contention; empty p95≤50 ms and inspection deadline≤200 ms or verdict is FAIL with unchanged target. |

## Required evidence shape

`.moai/reports/t1074/verdict.md` SHALL include Claim, command plus verbatim-output path, baseline commit/version attribution, explicit NOT_RUN gaps, residual risk, per-AC PASS/FAIL, process cleanup evidence, and nonce/receipt identifiers with payload bodies redacted where appropriate.

## Non-empty-pass execution protocol

Each command below writes Go's JSON event stream and then requires one exact `Action=pass` event for the named test. A missing test, a skipped live test, a package setup failure, or an ordinary `go test -run` empty match therefore fails the AC. Every invocation SHALL scrub provider credentials in the same compound shell, use a card-scoped `MOAI_HOME`/`GOCACHE`, and preserve the JSON log under `.moai/reports/t1074/`.

### Plan-phase RED-now ledger

The commands below were executed read-only at tree `758314007d8c696ff1af377dc8cdc46d76368314`. Each exact named test is absent, so `rg` returned no stdout and exit 1. This is the expected plan-phase RED reason: no AC can empty-pass because the green gate additionally requires an exact Go JSON `Action=pass` event. The run phase SHALL first add the named test, observe a behavioral RED, then implement and execute the corresponding green command below; mere test presence does not satisfy an AC.

| AC | Milestone | RED-now command | Verbatim stdout | Exit | Tree SHA | Green path |
|---|---|---|---|---:|---|---|
| AC-FMH-001 | M1 | `rg -n -F 'func TestFactoryCanonicalNamespaceAndIsolation(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact Go/JQ gate in AC-FMH-001 → `true` |
| AC-FMH-002 | M1 | `rg -n -F 'func TestFactoryRunSelectionAtomicSlotsAndArgv(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact Go/JQ gate in AC-FMH-002 → `true` |
| AC-FMH-003 | M1 | `rg -n -F 'func TestFactorySessionGenerationOwnership(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact Go/JQ gate in AC-FMH-003 → `true` |
| AC-FMH-004 | M2 | `rg -n -F 'func TestFactoryEnvelopeIdempotencyAndStaleAck(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact Go/JQ gate in AC-FMH-004 → `true` |
| AC-FMH-005 | M2 | `rg -n -F 'func TestFactoryCrashRecoveryExplicitReceipt(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact Go/JQ gate in AC-FMH-005 → `true` |
| AC-FMH-006 | M2 | `rg -n -F 'func TestFactoryDeadLetterAndLegacyIsolation(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact Go/JQ gate in AC-FMH-006 → `true` |
| AC-FMH-007 | M3 | `rg -n -F 'func TestFactoryHookContextAndContinuationSafety(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact Go/JQ gate in AC-FMH-007 → `true` |
| AC-FMH-008 | M3 | `rg -n -F 'func TestFactoryHookZeroTurnAndCapabilityTruth(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact Go/JQ gate in AC-FMH-008 → `true` |
| AC-FMH-009 | M2 | `rg -n -F 'func TestFactoryBrokerTrustBoundaries(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact Go/JQ gate in AC-FMH-009 → `true` |
| AC-FMH-010 | M4 | `rg -n -F 'func TestFactoryLiveCodexCodex(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact live Go/JQ gate in AC-FMH-010 → `true` |
| AC-FMH-011 | M4 | `rg -n -F 'func TestFactoryLiveCodexClaude(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact live Go/JQ gate in AC-FMH-011 → `true` |
| AC-FMH-012 | M4 | `rg -n -F 'func TestFactoryLiveClaudeCodex(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact live Go/JQ gate in AC-FMH-012 → `true` |
| AC-FMH-013 | M4 | `rg -n -F 'func TestFactoryLiveClaudeClaudeCompletionSeparation(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact live Go/JQ gate in AC-FMH-013 → `true` |
| AC-FMH-014 | M4 | `rg -n -F 'func TestFactoryLiveHookBoundaryIdleTruth(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact live Go/JQ gate in AC-FMH-014 → `true` |
| AC-FMH-015 | M4 | `rg -n -F 'func TestFactoryHookBenchmarkBudget(' internal` | `<empty>` | 1 | `758314007d8c696ff1af377dc8cdc46d76368314` | Exact benchmark Go/JQ gate in AC-FMH-015 → `true` |

### AC-FMH-001 — Canonical namespace and isolation

**Given** two linked worktrees for one repository plus distinct run/project fixtures, **when** both resolve and access the factory broker, **then** same project/run paths are equal and every distinct project/run access is denied.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && mkdir -p .moai/reports/t1074 && MOAI_HOME=/tmp/t1074-ac01-home GOCACHE=/tmp/t1074-ac01-cache go test -json ./internal/factorymsg -run '^TestFactoryCanonicalNamespaceAndIsolation$' -count=1 >.moai/reports/t1074/ac01.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryCanonicalNamespaceAndIsolation")' .moai/reports/t1074/ac01.jsonl
```

Expected final output: `true`.

### AC-FMH-002 — Run selection, atomic slots, and argv

**Given** zero, one, and multiple active-run fixtures and concurrent worker joins, **when** lead/worker CLI parsing executes, **then** exact diagnostics/select behavior, unique slots, and byte-identical post-`--` argv are observed.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-ac02-home GOCACHE=/tmp/t1074-ac02-cache go test -json ./internal/cli -run '^TestFactoryRunSelectionAtomicSlotsAndArgv$' -count=1 >.moai/reports/t1074/ac02.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryRunSelectionAtomicSlotsAndArgv")' .moai/reports/t1074/ac02.jsonl
```

Expected final output: `true`.

### AC-FMH-003 — Session and generation ownership

**Given** current, stale, PID-reused, and inherited-subagent identities, **when** each tries to claim/read/receipt, **then** only the exact session UUID/generation/process-start tuple succeeds.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-ac03-home GOCACHE=/tmp/t1074-ac03-cache go test -json ./internal/factorymsg ./internal/hook -run '^TestFactorySessionGenerationOwnership$' -count=1 >.moai/reports/t1074/ac03.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactorySessionGenerationOwnership")' .moai/reports/t1074/ac03.jsonl
```

Expected final output: `true`.

### AC-FMH-004 — Envelope, idempotency, and stale ACK

**Given** valid/invalid envelopes, retry keys, expired claims, stale tokens, and task revisions, **when** broker operations run, **then** the closed schema, one stored send, redelivery, stale-ACK rejection, and revision guard are observed.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-ac04-home GOCACHE=/tmp/t1074-ac04-cache go test -json ./internal/factorymsg -run '^TestFactoryEnvelopeIdempotencyAndStaleAck$' -count=1 >.moai/reports/t1074/ac04.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryEnvelopeIdempotencyAndStaleAck")' .moai/reports/t1074/ac04.jsonl
```

Expected final output: `true`.

### AC-FMH-005 — Crash recovery and explicit receipt

**Given** failures before hook output, after output, before disposition, and before receipt, **when** the lease expires and the receiver resumes, **then** each message is recoverable and acknowledgement occurs only after persisted disposition plus matching explicit receipt.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-ac05-home GOCACHE=/tmp/t1074-ac05-cache go test -json ./internal/factorymsg -run '^TestFactoryCrashRecoveryExplicitReceipt$' -count=1 >.moai/reports/t1074/ac05.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryCrashRecoveryExplicitReceipt")' .moai/reports/t1074/ac05.jsonl
```

Expected final output: `true`.

### AC-FMH-006 — Dead-letter and legacy isolation

**Given** TTL, poison, overflow, and pre-existing legacy sessionmsg fixtures, **when** the factory broker handles them, **then** bounded reason-bearing dead letters/backpressure appear and legacy bytes remain unchanged.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-ac06-home GOCACHE=/tmp/t1074-ac06-cache go test -json ./internal/factorymsg ./internal/sessionmsg -run '^TestFactoryDeadLetterAndLegacyIsolation$' -count=1 >.moai/reports/t1074/ac06.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryDeadLetterAndLegacyIsolation")' .moai/reports/t1074/ac06.jsonl
```

Expected final output: `true`.

### AC-FMH-007 — Hook context and continuation safety

**Given** 17 messages including a malicious raw body and an existing higher-priority stop decision, **when** UserPromptSubmit then Stop run, **then** output carries at most 16 IDs/2 KiB, contains no body, continues once, does not re-claim, and preserves the earlier stop.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-ac07-home GOCACHE=/tmp/t1074-ac07-cache go test -json ./internal/hook ./internal/codexadapter -run '^TestFactoryHookContextAndContinuationSafety$' -count=1 >.moai/reports/t1074/ac07.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryHookContextAndContinuationSafety")' .moai/reports/t1074/ac07.jsonl
```

Expected final output: `true`.

### AC-FMH-008 — Zero-turn and truthful capability cases

**Given** empty, receipt-only, permission-wait, interrupted, disabled, untrusted, incompatible, and lock-timeout cases, **when** hooks run, **then** no continuation/model-turn signal appears and status names pending/degraded/capability reason.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-ac08-home GOCACHE=/tmp/t1074-ac08-cache go test -json ./internal/hook -run '^TestFactoryHookZeroTurnAndCapabilityTruth$' -count=1 >.moai/reports/t1074/ac08.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryHookZeroTurnAndCapabilityTruth")' .moai/reports/t1074/ac08.jsonl
```

Expected final output: `true`.

### AC-FMH-009 — Trust-boundary attacks

**Given** traversal IDs, fake identity/run, unknown kind, prompt-injection payload, unauthorized revision, and one valid dispatch pointer, **when** the MCP/core surfaces process them, **then** every attack is rejected without queue mutation and the valid pointer succeeds.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-ac09-home GOCACHE=/tmp/t1074-ac09-cache go test -json ./internal/factorymsg ./internal/cli -run '^TestFactoryBrokerTrustBoundaries$' -count=1 >.moai/reports/t1074/ac09.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryBrokerTrustBoundaries")' .moai/reports/t1074/ac09.jsonl
```

Expected final output: `true`.

### AC-FMH-010 — Live Codex lead ↔ Codex worker

**Given** two real separately launched Codex CLI/model contexts in one run, **when** both directions exchange unique nonces and explicit receipts, **then** each target model reports the other nonce and the broker records matching recipient generation/claim token.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=codex-codex MOAI_HOME=/tmp/t1074-live-cc-home GOCACHE=/tmp/t1074-live-cc-cache go test -json ./internal/cli -run '^TestFactoryLiveCodexCodex$' -count=1 -timeout=180s >.moai/reports/t1074/ac10.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryLiveCodexCodex") and (any(.[]; .Action=="skip" and .Test=="TestFactoryLiveCodexCodex")|not) and (any(.[]; ((.Output // "") | contains("NOT_RUN")))|not)' .moai/reports/t1074/ac10.jsonl
```

Expected final output: `true`; `skip` and `NOT_RUN` fail.

### AC-FMH-011 — Live Codex lead ↔ Claude worker

**Given** real separately launched Codex lead and Claude worker model contexts, **when** both directions exchange nonces and receipts, **then** each target model reports the other nonce with matching receipt binding.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=codex-claude MOAI_HOME=/tmp/t1074-live-cxcl-home GOCACHE=/tmp/t1074-live-cxcl-cache go test -json ./internal/cli -run '^TestFactoryLiveCodexClaude$' -count=1 -timeout=180s >.moai/reports/t1074/ac11.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryLiveCodexClaude") and (any(.[]; .Action=="skip" and .Test=="TestFactoryLiveCodexClaude")|not) and (any(.[]; ((.Output // "") | contains("NOT_RUN")))|not)' .moai/reports/t1074/ac11.jsonl
```

Expected final output: `true`; `skip` and `NOT_RUN` fail.

### AC-FMH-012 — Live Claude lead ↔ Codex worker

**Given** real separately launched Claude lead and Codex worker model contexts, **when** both directions exchange nonces and receipts, **then** each target model reports the other nonce with matching receipt binding.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=claude-codex MOAI_HOME=/tmp/t1074-live-clcx-home GOCACHE=/tmp/t1074-live-clcx-cache go test -json ./internal/cli -run '^TestFactoryLiveClaudeCodex$' -count=1 -timeout=180s >.moai/reports/t1074/ac12.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryLiveClaudeCodex") and (any(.[]; .Action=="skip" and .Test=="TestFactoryLiveClaudeCodex")|not) and (any(.[]; ((.Output // "") | contains("NOT_RUN")))|not)' .moai/reports/t1074/ac12.jsonl
```

Expected final output: `true`; `skip` and `NOT_RUN` fail.

### AC-FMH-013 — Live Claude regression and completion separation

**Given** real Claude lead/worker contexts and a report without independent work evidence, **when** they exchange nonces/receipts and the report is evaluated, **then** transport passes but card completion remains refused.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=claude-claude MOAI_HOME=/tmp/t1074-live-clcl-home GOCACHE=/tmp/t1074-live-clcl-cache go test -json ./internal/cli -run '^TestFactoryLiveClaudeClaudeCompletionSeparation$' -count=1 -timeout=180s >.moai/reports/t1074/ac13.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryLiveClaudeClaudeCompletionSeparation") and (any(.[]; .Action=="skip" and .Test=="TestFactoryLiveClaudeClaudeCompletionSeparation")|not) and (any(.[]; ((.Output // "") | contains("NOT_RUN")))|not)' .moai/reports/t1074/ac13.jsonl
```

Expected final output: `true`; `skip` and `NOT_RUN` fail.

### AC-FMH-014 — Hook-boundary idle truth

**Given** both real sessions have become idle and a message arrives after final Stop, **when** no human input occurs, **then** it remains `pending-until-next-turn`, no model call/key injection occurs, and the next legitimate turn receives it.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=idle-boundary MOAI_HOME=/tmp/t1074-live-idle-home GOCACHE=/tmp/t1074-live-idle-cache go test -json ./internal/cli -run '^TestFactoryLiveHookBoundaryIdleTruth$' -count=1 -timeout=180s >.moai/reports/t1074/ac14.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryLiveHookBoundaryIdleTruth") and (any(.[]; .Action=="skip" and .Test=="TestFactoryLiveHookBoundaryIdleTruth")|not) and (any(.[]; ((.Output // "") | contains("NOT_RUN")))|not)' .moai/reports/t1074/ac14.jsonl
```

Expected final output: `true`; `skip` and `NOT_RUN` fail.

### AC-FMH-015 — Measured performance budget

**Given** empty/16/1000 queues, 1/10 sessions, warm/cold runs, and lock contention, **when** the full hook path is measured on the reported baseline, **then** the evidence contains p50/p95/read/write/process/injected-byte metrics and asserts empty added p95≤50 ms plus inspection≤200 ms.

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_BENCH=1 MOAI_HOME=/tmp/t1074-bench-home GOCACHE=/tmp/t1074-bench-cache go test -json ./internal/hook -run '^TestFactoryHookBenchmarkBudget$' -count=1 -timeout=180s >.moai/reports/t1074/ac15.jsonl && jq -se 'any(.[]; .Action=="pass" and .Test=="TestFactoryHookBenchmarkBudget") and (any(.[]; .Action=="skip" and .Test=="TestFactoryHookBenchmarkBudget")|not) and (any(.[]; ((.Output // "") | contains("NOT_RUN")))|not)' .moai/reports/t1074/ac15.jsonl
```

Expected final output: `true`; a silently raised threshold, missing matrix cell, `skip`, or `NOT_RUN` fails.
