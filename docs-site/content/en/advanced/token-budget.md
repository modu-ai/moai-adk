---
title: Token Budget Management and Graceful Stop
weight: 2
draft: false
---

This page deep-dives into Layer D — Budget defense — of the 4-layer tokenomics structure. It covers the graceful-abort mechanism that ensures the session stops without data loss when an agent reaches the context-window limit, preserving progress so the next session can continue.

## Why Budget Defense

Anthropic SSE streams intermittently stall (`stream_idle_partial`) near the context-window ceiling. This is probabilistic but predictable above the model-specific threshold. When a stall occurs, an agent call may fail mid-stream, potentially losing progress.

Budget defense addresses this proactively. Before context usage reaches the threshold, the system performs a graceful abort, ensuring the session transitions to the next step without loss.

## Per-Model Context Thresholds

The operational threshold is model-specific. Larger windows tolerate higher percentage utilization; smaller windows hit the operational ceiling later in percentage terms but with less absolute headroom.

| Model class | Window | Handoff threshold | Absolute ceiling |
|-------------|--------|-------------------|------------------|
| Opus 4.8 (1M) | 1,000,000 tokens | 50% | ~500,000 tokens |
| GLM-5.2 (1M) | 1,000,000 tokens | 50% | ~500,000 tokens |
| Opus / Fable (256K) | 256,000 tokens | 90% | ~230,000 tokens |
| Sonnet / Opus standard (200K) | 200,000 tokens | 90% | ~180,000 tokens |
| Haiku (200K) | 200,000 tokens | 90% | ~180,000 tokens |

GLM-5.2 (via `moai glm` / `moai cg` GLM panels) is a 1M-context model and is operated at the 50% threshold. Claude Code reports `context_window_size` based on the Claude slot (Opus=1M, Sonnet/Haiku=200K), so raw telemetry may show ~180K under GLM; MoAI corrects this to 1M. Trust the statusline CW% gauge.

## Two-Stage Handoff Marker

The statusline appends a `/clear` hint to the context bar in two stages.

- {{< icon warning warn >}} **Soft marker** `(⚠️/clear)` — displayed at the band's soft threshold. An advisory signal allowing the user to decide whether to run `/clear`.
- {{< icon warning danger >}} **Hard marker** `(🛑/clear!)` — displayed at the auto-compact-aware ceiling. The next action MUST be `/clear`.

The hard ceiling is set near the auto-compact threshold, so runtime auto-compact frequently preempts it and the hard marker rarely fires in practice. This is an intentional tradeoff of the auto-compact-aware formula.

## Graceful Abort Procedure

The graceful-abort mechanism implemented by SPEC-TOKEN-BUDGET-STOP-001 works as follows.

1. **Detect** — `Tracker.IsAtHardLimit(agentName)` returns true (cumulative usage ≥ hard_clear_threshold, default 0.90)
2. **Persist state** — in-flight work state is persisted to `progress.md`
3. **Emit handoff** — a paste-ready resume message is generated (6-block structure)
4. **Recommend turn end** — the user is advised to `/clear` (HARD: auto-`/clear` is NEVER performed)
5. **Persist evidence** — verification evidence is persisted under `.moai/state/verify/`

`/clear` is never executed automatically. The system only recommends that the user run `/clear`; the user decides.

## Paste-Ready Resume 6-Block Structure

The session-handoff message follows the 6-block structure below. Each block is designed so the next session can continue the work with minimal information.

```text
✂──── Copy from here ────✂

ultrathink. <SPEC-ID> <phase> entering.
applied lessons: <memory-file-1>, <memory-file-2>

Preconditions:
1) <verifiable precondition 1>
2) <verifiable precondition 2>

Run: <command or action>

After merge: <next action or SPEC>

✂──── Copy to here ────✂
```

Role of each block:

- **Block 1** — `ultrathink.` opener sets effort:xhigh, declares the entering phase and SPEC-ID
- **Block 2** — `applied lessons:` references memory files learned from prior sessions (max 4)
- **Block 3** — `Preconditions:` verifiable preconditions the next session must check before starting (max 4, each ≤200 chars)
- **Block 4** — individual precondition items
- **Block 5** — `Run:` single primary action (typically `/moai <subcommand>`)
- **Block 6** — `After merge:` next action or SPEC ID

## Verify-Diet

The file-redirect contract that redirects long verification-command output to disk and leaves only a summary in context.

Rule: when a verification command's verbatim output exceeds the **bounded-tail ceiling** (default 50 lines or 2KB, whichever is smaller), the output is redirected to a file and only the exit code + bounded tail are surfaced in context.

```bash
go test ./... > /tmp/moai-verify/1-go-test.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/1-go-test.log
```

This contract keeps verbatim evidence on disk while the context carries only exit code + bounded tail. It removes the double-burn (inline output + banner re-quote), not the evidence itself.

## Evidence Persistence Obligation

Evidence written by the file-redirect contract to `/tmp` is periodically cleared by the OS (macOS reboot, Linux tmpfs remount, systemd-tmpfiles). When the cited path no longer resolves to a file, the evidence is unreachable at audit time.

The persistence obligation solves this. Verification evidence MUST be persisted under `.moai/state/verify/<session>/`. This directory is a gitignored runtime-state area, the same family as `context-usage.json` and `active-sessions.json`.

The exact persistence mechanism (direct write or `/tmp` write followed by a copy) is an implementation detail. The contract states the obligation: evidence MUST survive `/tmp` clearance and remain at a citable, audit-time-reachable path.

## Next Steps

- [Tokenomics Overview](/en/advanced/tokenomics-overview/) — full 4-layer structure overview
- [3-Tier Agent Architecture](/en/advanced/no-haiku-3tier/) — the model-policy foundation of Layer B routing
