#!/bin/bash
# IGGDA Phase Driver — Stop hook for Intent-Gated Goal-Directed Autonomy.
#
# Fires at turn-end during an IGGDA pipeline run. Reads progress.md + invokes
# `moai spec audit --json --filter-spec=<SPEC-ID>` and emits a /goal-style
# auto-advance signal when the current phase's safe-transition predicate holds.
#
# This script is a STANDALONE shell hook (not a `moai hook` subcommand). It
# runs directly under the Stop hook. It MUST NOT invoke AskUserQuestion (the
# asymmetric orchestrator-subagent boundary — hooks return exit codes + JSON,
# the orchestrator translates blocks to AskUserQuestion).
#
# Recovery-Signal Carve-Out: when the Stop input carries stop_hook_active=true
# (a Stop hook already continued this turn — the official loop guard) OR a
# best-effort recovery-turn keyword, this hook exits 0 immediately (never
# blocks a recovery). See
# .claude/rules/moai/workflow/runtime-recovery-doctrine.md §4.
#
# Delivery channel: the hook always exits 0. A block is signaled via stdout
# JSON {"decision":"block","reason":"..."} — per Claude Code Stop-hook
# semantics, stdout JSON is honored only on exit 0 (on exit 2 stdout is
# discarded and only stderr is surfaced, which the generated wrappers redirect
# to a log file). On a pass, stdout is intentionally empty (silent continue).

set -u

# --- Read stdin JSON from Claude Code ---
INPUT=$(cat)

# --- Loop guard + Recovery-Signal Carve-Out ---
# stop_hook_active (official Stop input field) is true when a Stop hook has
# already continued the turn — blocking again risks the death-spiral loop, so
# exit 0 immediately. The stopReason keyword scan below is best-effort only
# (stopReason is not part of the official Stop stdin contract).
stop_reason=""
if command -v jq &> /dev/null; then
    stop_hook_active=$(echo "$INPUT" | jq -r '.stop_hook_active // false' 2>/dev/null || echo "false")
    if [ "$stop_hook_active" = "true" ]; then
        exit 0
    fi
    stop_reason=$(echo "$INPUT" | jq -r '.stopReason // empty' 2>/dev/null || echo "")
fi
# Recovery-signal keywords (runtime-recovery-doctrine §1 withheld-recoverable set).
case "$stop_reason" in
    *prompt_too_long*|*max_output_tokens*|*media_size*|*compact*)
        # Recovery turn — do not block. Exit 0.
        exit 0
        ;;
esac

# --- Self-gate: only act on genuine IGGDA completion turns ---
# The Stop hook fires on every turn-end. This driver self-gates: it inspects
# whether an IGGDA pipeline is active for a SPEC before evaluating the predicate.
# If no SPEC_ID is detectable from the environment or progress markers, exit 0.
SPEC_ID="${MOAI_IGGDA_SPEC_ID:-}"
if [ -z "$SPEC_ID" ]; then
    # No active IGGDA SPEC — nothing to evaluate. Exit 0 (non-blocking).
    exit 0
fi

# --- Read progress.md for the current SPEC ---
# CLAUDE_PROJECT_DIR is guarded (set -u): fall back to $PWD when unset.
PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-$PWD}"
PROGRESS_MD="$PROJECT_ROOT/.moai/specs/${SPEC_ID}/progress.md"
if [ ! -f "$PROGRESS_MD" ]; then
    # progress.md absent — cannot evaluate predicate. Graceful degradation: exit 0.
    exit 0
fi

# --- Invoke moai spec audit --json --filter-spec=<SPEC-ID> ---
# This is the dedicated verification tool per verification-claim-integrity.md §1.1
# surface 3. We do NOT infer phase completion from frontmatter text alone.
MUST_FIX_COUNT=0
if command -v moai &> /dev/null; then
    AUDIT_JSON=$(moai spec audit --json --filter-spec="${SPEC_ID}" 2>/dev/null || echo "")
    if [ -n "$AUDIT_JSON" ] && command -v jq &> /dev/null; then
        MUST_FIX_COUNT=$(echo "$AUDIT_JSON" | jq '[.drift_findings[]? | select(.severity == "MUST-FIX")] | length' 2>/dev/null || echo 0)
    fi
fi
# Graceful degradation: if moai or jq is unavailable, MUST_FIX_COUNT stays 0
# (non-blocking). The orchestrator's verification-batch re-checks deterministically.

# --- Evaluate the current phase's safe-transition predicate ---
# Phase determination is derived from progress.md §E markers (NOT frontmatter).
# This is a conservative evaluation — the orchestrator makes the final advance
# decision based on the full predicate (plan-auditor verdict, sync-auditor score,
# go test exit, git status). This hook only blocks when MUST-FIX drift exists.
if [ "${MUST_FIX_COUNT:-0}" -gt 0 ]; then
    # MUST-FIX drift blocks phase advance. Emit the Stop block decision on
    # stdout + exit 0 (the JSON decision field is the blocking channel; exit 2
    # would discard stdout). The orchestrator translates this to
    # AskUserQuestion (NEVER done here). The reason text doubles as the
    # ledger-closing note per agent-common-protocol.md § Ledger Closure.
    printf '{"decision":"block","reason":"IGGDA phase advance blocked: %s MUST-FIX drift finding(s) detected by moai spec audit --filter-spec=%s"}\n' \
        "$MUST_FIX_COUNT" "$SPEC_ID"
    exit 0
fi

# --- Predicate holds (or could not be evaluated) → exit 0 (continue) ---
exit 0
