---
id: SPEC-MCP-NEUTRALITY-001
title: "Design — target contracts for refusal structuring and harness neutrality"
version: "0.1.0"
created: 2026-10-14
updated: 2026-10-14
author: manager-spec
---

# SPEC-MCP-NEUTRALITY-001 — Design

Design stays at the contract level (data shapes, invariants, ownership); code-level structure is the implementer's judgment inside these contracts.

## §D1 Refusal result contract (REQ-MN-001/002)

A policy refusal is a *successful* MCP result whose payload describes the refusal:

```json
{
  "refused": true,
  "code": "serial-slot | owned | raced | parent-checkout | stale-run",
  "reason": "<human-readable sentence>",
  "hint": "<what would make it succeed, when one exists>"
}
```

Invariants:
- The set of codes is closed (plan §D3). The MCP error channel (`isError: true`) is never used for these five classes.
- The refusal JSON appears in BOTH `structuredContent` and the text content block (same serializer as REQ-MN-008).
- The failure observer classifies by the presence of `refused: true` in the structured payload: refused results never become lessons-inbox mcp rows. Classification happens at the observer seam (one place), not at each handler.

## §D2 Defect-to-requirement map

The full 36-row defect map with per-defect anchors lives in `plan.md` §B. Requirements cluster defects by shared behavior; no requirement maps one-to-one to a defect.

## §D3 Root resolution contract (REQ-MN-004/005)

One resolver serves every tool:

1. If the caller supplied `project_root` (string): validate — reject non-strings as invalid arguments; reject the home directory; confine to a registered worktree of the server's own repository (the `codexAuditValidateRoot` pattern generalized).
2. Else if the tool is mutating: refuse (missing-required), with `hint` naming the parameter.
3. Else: fall back to the resolved server root (existing precedence), and set `_root.warning`.

Every root-dependent result carries `_root {dir, source, warning}` where `source` ∈ {`param`, `server-cwd`, `fallback`}; `warning` is set for anything other than `param`.

## §D4 Identity contract (REQ-MN-003)

- Actor fields (`actor`/`session_id`/`run_id` as appropriate per tool) are first-class arguments; spawn env (`MOAI_FACTORY_ROLE`, `MOAI_FACTORY_WORKER`, `CLAUDE_CODE_SESSION_ID`, `MOAI_LAUNCH_PROVIDER`) supplies the default only.
- Positive evidence only: a role exists because a marker says so, never because a lane marker is absent. The operator mark extends the existing launcher-stamped role vocabulary (decision-index Q2 default).
- Run-id validation reads only the active run row; any other run id refuses `stale-run`.

## §D5 Job lifecycle model (REQ-MN-014)

Every background job record carries: `owner_server {pid, commit}`, `created_at`, `ttl`. On server exit, a shutdown pass cancels all live job contexts (codex, GLM, role-audit). `*_job_status` checks the owner's liveness and reports `stale` when dead. Foreground `codex_task` records its thread id at start so a timeout returns partial output plus a resumable thread. Registries prune by TTL.

## §D6 Broker claim model (REQ-MN-016/017/018)

- Claim identity: `(endpoint, generation)`. Re-claim by the same endpoint while its lease is live extends `claim_expires_at` and keeps `claim_token`. A NEW token is minted only when the previous lease genuinely expired.
- Generation advance: pending messages for the old generation migrate to the new one when the receiver re-registers; messages whose sender requested TTL expiry dead-letter with a sender notification. Zero permanently-unclaimable rows is the invariant.
- Rendering: the hook renders N visible lines within its byte cap, cutting on a line boundary, and claims exactly the rendered set.
- Stop: block only for kinds in the decision-required set (blocker, status_request); all other deliveries are counts. Degraded broker state logs on the Stop side.
- Peer attribution resolves the live session identity at call time; a stale spawn-time id refuses with guidance rather than silently addressing a dead endpoint.

## §D7 Audit evidence contract (REQ-MN-019..024)

- One scope object travels through every backend leg, result, receipt, and gate record: `{tree, head, base_ref, base_sha, digest, truncated, excluded_untracked}`. A receipt without a matching digest never corroborates.
- Verdict vocabulary: `pass | fail | inconclusive` for content; "could not run" is `inconclusive` + `gate_unmet` + preserved stderr, never `fail`. Exactly one retry on transport failure.
- Convergence requires participant evidence: a required-inconclusive or below-threshold result is a distinct not-converged outcome, never pass.
- Gate state is keyed by tree digest, written atomically, with validated session ids; could-not-run receipts are never reused as passes regardless of TTL.

## §D8 State-root seam contract (REQ-MN-025)

For each artifact class (receipts, review/gate state, broker store), exactly one resolution rule: the caller-named project root (REQ-MN-004 contract), falling back to the repository toplevel. Producers and consumers MUST resolve through the same rule — no consumer reads a path computed from a different root (server cwd, git common dir, or spawn env). This is the seam only; layout unification stays with card 3.2-2-1.
