---
name: moai-ref-cross-model-audit
description: >
  Cross-model audit convergence reference for the plan-auditor and sync-auditor
  agents. Documents how to invoke the `audit_multi` MCP tool to fan a code
  review out across the codex and GLM (z.ai) backends in parallel, converge
  their verdicts with the in-session Claude verdict, and fold the resulting
  per-backend verdicts + disagreement flag into the audit output. The single
  skill both audit entry points load — no duplication.

when_to_use: >
  Use when the project's `audit_model` is `multi` AND the auditor needs a
  cross-backend second opinion before reaching a verdict. Single-backend paths
  (claude-only, codex-only, or glm-only) do NOT load this skill — the
  `audit_multi` tool is the multi-model entry point only. Also use when the
  auditor must explain WHY the convergence result is a pass, fail, or
  advisory-only disagreement.

user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
---

# Cross-Model Audit Convergence

This skill is the single load-point both plan-auditor and sync-auditor use when
the project opts into multi-model audit (`audit_model: multi`). It documents the
one MCP tool the auditor calls, the independence rule that tool enforces, and
how to fold the returned convergence result into the auditor's verdict.

## When to use convergence vs single-model

| Project setting | Path | Skill |
|---|---|---|
| `audit_model: claude` (default) | Claude reviews alone | (none — no second opinion needed) |
| `audit_model: codex` | Codex reviews alone | `moai-ref-owasp-checklist` etc., no convergence |
| `audit_model: glm` | GLM reviews alone | (same) |
| `audit_model: multi` | Claude + codex + GLM, converged | **this skill** |

Single-model paths do NOT load this skill. Convergence is only the multi-model
concern.

## The `audit_multi` MCP tool

The single tool surface is:

```
mcp__moai__audit_multi
```

It is exposed by the `moai mcp-server` stdio server (the self-hosted MCP server
shipped with the binary). The tool is a thin wrapper over the convergence
engine: it does NOT re-implement the codex or GLM backends — it fans out by
calling the existing single-backend handlers in parallel and synthesizes their
results.

### Input parameters

| Parameter | Type | Required | Notes |
|---|---|---|---|
| `claude_verdict` | object | YES | The in-session Claude review verdict. Object shape: `{verdict, summary, findings, next_steps}` — the same `review-output.schema.json` the single-backend tools return. |
| `target` | string | no | What the secondary backends review (`uncommittedChanges`, `baseBranch`). Passed through unchanged. |
| `focus` | string | no | Optional focus area forwarded to the secondary backends (e.g. `concurrency`, `auth`). |
| `gates` | object | no | Per-auditor gate map (`claude`/`codex`/`glm` ∈ `off`/`advisory`/`required`). When omitted, distributed defaults apply: claude required, codex required, glm advisory. |
| `session_id` | string | no | When set, the result is persisted to `.moai/state/audit-multi/<session>.json` so the multi-review-gate Stop hook reads the most recent result rather than re-invoking convergence. |

### Output shape

The tool returns a `ConvergenceResult`:

```json
{
  "per_backend_verdicts": [
    {"backend": "claude", "gate": "required", "verdict": "pass", "summary": "...", "findings": [], "next_steps": []},
    {"backend": "codex",  "gate": "required", "verdict": "fail", "summary": "...", "findings": [...], "next_steps": [...]},
    {"backend": "glm",    "gate": "advisory", "verdict": "pass", "summary": "...", "findings": [], "next_steps": []}
  ],
  "overall_verdict": "fail",
  "disagreement_flag": true,
  "residual_risk_note": "cross-model disagreement (advisory, NOT a block): pass=[claude(required), glm(advisory)] fail=[codex(required)]",
  "fail_open_backends": []
}
```

- `overall_verdict` ∈ `{pass, fail}` — the existing review-output values. No
  new enum (disagreement is a flag, not a verdict value).
- `disagreement_flag` is `true` when either a required split OR an advisory-only
  conflict was detected.
- `residual_risk_note` describes the convergence outcome in prose (which
  backend(s) failed, or the shape of the split). Surface this in the audit
  report's residual-risk section.
- `fail_open_backends` lists the backends that returned `inconclusive` (missing,
  unauthenticated, or erroring) — surfaced so the report can name them.

### How to invoke

Call the tool with the in-session Claude analysis folded into the
`claude_verdict` object. Do NOT pass the full Claude analysis text as prompt
context for the secondary backends — see the Independence rule below.

```
result = mcp__moai__audit_multi({
  claude_verdict: { verdict: <your verdict>, summary: <one-line>, findings: [...], next_steps: [...] },
  target: "uncommittedChanges",
  focus: "concurrency",
  session_id: <current session id>
})
```

The orchestrator-side question channel is preserved: the tool returns a
structured result, never prompts the user. On a missing anchor or inconclusive
condition, surface the structured `overall_verdict: fail` + `residual_risk_note`
in the audit report and let the orchestrator translate.

## Independence rule (load-bearing)

> **Pass only the synthesized `claude_verdict` object to the MCP tool — NEVER
> the full Claude analysis text as prompt context for the secondary backends.**

The secondary backends (codex, GLM) are SUPER-REVIEWS: uncorrelated second
opinions. Their value collapses to a re-sample of Claude's reasoning the moment
they see Claude's analysis. The convergence engine enforces this structurally —
the `claude_verdict` is consumed ONLY by the synthesis step, and the secondary
backends receive `(target, focus)` alone — but the auditor must not undermine
the invariant by pasting Claude's reasoning into the `focus` field either.

Concretely:

- `focus` carries a short AREA name (`concurrency`, `auth`, `secret handling`),
  not a paragraph of analysis.
- `claude_verdict.summary` is a one-line verdict rationale, not the full review.
- The findings you surface in the audit report come from
  `per_backend_verdicts[].findings` (each backend's own findings), NOT from
  echoing Claude's findings back.

## Convergence policy (how overall_verdict is derived)

The engine derives `overall_verdict` per a 4-case table:

| Case | Condition | overall_verdict | disagreement_flag |
|---|---|---|---|
| 1 | All required backends PASS | `pass` | `false` |
| 2 | Any required FAIL (no required PASS to split against) | `fail` | `false` |
| 3 | Required split (≥1 required PASS + ≥1 required FAIL) | `fail` (conservative) | `true` |
| 4 | Advisory-only conflict (all required PASS, ≥1 advisory FAIL) | `pass` | `true` |

Two invariants follow:

- **Disagreement is advisory, NOT a block.** A `disagreement_flag: true` result
  is surfaced as residual-risk + advisory in the audit report; it never
  hard-blocks the flow on its own. The required-gate contract holds per backend,
  so the only block-shaped outcome is a required FAIL (cases 2/3 →
  `overall_verdict: fail`).
- **Advisory backends never flip overall to fail.** Case 4 records the advisory
  conflict but keeps `overall_verdict: pass`. This is the fixed user-policy term:
  an advisory FAIL is reported, not enforced.

## Fail-open identity

Codex and GLM are OPTIONAL. A missing, unauthenticated, erroring, or malformed
backend yields `verdict: inconclusive` in its `per_backend_verdicts` slot and
convergence continues over the remaining active backends. The autonomous flow is
NEVER hard-blocked on a missing optional dependency — `evidence-of-absence ≠
evidence-of-failure`.

When ALL non-Claude backends are inconclusive, the overall verdict fails open to
the in-session Claude verdict (the always-available anchor).

## Folding the result into the audit verdict

The auditor's verdict and the convergence result relate as follows:

| `overall_verdict` | `disagreement_flag` | Auditor action |
|---|---|---|
| `pass` | `false` | Standard PASS. No residual-risk row needed. |
| `pass` | `true` | PASS with a residual-risk row naming the advisory disagreement. |
| `fail` | (any) | FAIL. Name the failing required backend(s) from `per_backend_verdicts`. The block is conservative (cases 2/3). |

In all cases, surface `residual_risk_note` verbatim in the audit report's
residual-risk section so a human reader sees which backend disagreed with which.

## Cross-references

- `mcp__moai__codex_audit`, `mcp__moai__glm_audit` — the single-backend tools
  whose handlers the convergence engine reuses (the engine does NOT re-implement
  them).
- `workflow.audit.gates.*` — the per-auditor gate map (`off`/`advisory`/`required`).
- `workflow.multi.review_gate.enabled` — opt-in toggle for the multi-review-gate
  Stop hook (the Path C fully-autonomous gate). Default OFF; opt in via local
  config.
