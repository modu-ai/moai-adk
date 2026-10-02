---
name: moai-ref-cross-model-audit
description: >
  Cross-model audit convergence reference for the plan-auditor and sync-auditor
  agents. Documents how to invoke the `audit_multi` MCP tool to fan a code
  review out across the Claude subscription, codex, and GLM (z.ai) backends,
  converge their independent verdicts, and fold the resulting
  per-backend verdicts + disagreement flag into the audit output. The single
  skill both audit entry points load — no duplication.

when_to_use: >
  Use when the audit plan (`moai verify audit-plan`) reports a cross-model
  backend AND the auditor needs a cross-backend second opinion before reaching
  a verdict. Also use when the
  auditor runs under a GPT or GLM main session and needs an independent Claude
  subscription verdict through `claude_audit`, or when it must explain why the
  convergence result is a pass, fail, or
  advisory-only disagreement.

user-invocable: false
metadata:
  version: "1.1.0"
  category: "domain"
  status: "active"
---

# Cross-Model Audit Convergence

This skill is the single load-point both plan-auditor and sync-auditor use when
the audit plan reports a cross-model backend. It documents the one MCP tool the
auditor calls, the independence rule that tool enforces, how to check the result
against the plan, and how to fold the returned convergence result into the
auditor's verdict.

## When to use convergence vs single-model

The auditor does not choose a backend by interpreting the `audit_model` value. It
runs `moai verify audit-plan --project-root <own toplevel>` first (the full flow
is in each agent's MCP Audit Tools section; pass the toplevel yourself, because
`CLAUDE_PROJECT_DIR` names the primary checkout in a worktree session) and
follows the plan:

| Outcome of the verb | Path | This skill |
|---|---|---|
| `config_status` `ok` or `absent`, `cross_model_active: false` (the distributed default, an explicit `claude` token) | Claude main: in-session review; GPT/GLM main: `claude_audit` | for external-main sessions |
| `cross_model_active: true` | `audit_multi` without a `gates` argument: the plan's backends, converged | **this skill** |
| The output is the `verify` group's help text (it contains `Shared diagnostic snapshot contract` and neither `config_status` nor an `audit-plan:` line) | The legacy path below, with "plan surface unreachable, legacy path used" named as a Gap | only as the legacy path says |
| `config_status: unreadable`, an `audit-plan:` error, or any other failure to run the verb (refused, crashed, timed out, malformed output) | Not the legacy path: a PASS-blocking Gap; a configured required backend that does not answer stays fail-closed | no PASS on this audit |

**Legacy path (a binary that predates the verb).** Behave as before the plan
verb existed: read the project's `audit_model` from `workflow.yaml` — `multi`
converges Claude, codex and GLM via `audit_multi`, `claude` keeps the
single-model path, `glm` and `codex` call that one backend directly. Wherever
`audit_multi` is called, pass it without a `gates` argument, and call it
whenever the tree's `workflow.yaml` sets an audit model other than `claude` or
any `audit.gates` key, so a plan-aware server applies the configured plan.

Single-model paths do NOT load this skill. Convergence is only the cross-model
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
| `claude_verdict` | object | conditional | Claude main sessions pass their in-session verdict. GPT/GLM/unknown-origin sessions may omit it; `audit_multi` ignores any supplied value and performs a fresh Claude subscription audit. |
| `target` | string | no | What the secondary backends review (`uncommittedChanges`, `baseBranch`). The string reaches both backends unchanged; for codex, `baseBranch`'s branch name is then resolved server-side from the reviewed tree (remote default head, then `main`) — it cannot be supplied here. |
| `focus` | string | no | Optional focus area forwarded to the secondary backends (e.g. `concurrency`, `auth`). |
| `gates` | object | no | Per-auditor gate map (`claude`/`codex`/`glm` ∈ `off`/`advisory`/`required`). Auditors do not pass it: omitted, the tree's own plan applies (the configured audit model and gates); with no configuration the distributed defaults apply — claude required, codex required, glm advisory — and an unconfigured gate stays fail-open. |
| `session_id` | string | no | When set, the result is persisted to `.moai/state/audit-multi/<session>.json` so the multi-review-gate Stop hook reads the most recent result rather than re-invoking convergence. |
| `project_root` | string | no *(REQUIRED in a worktree)* | The tree the backends should read — this session's own `git rev-parse --show-toplevel`. Omitted from a worktree, the fan-out reads the PRIMARY checkout instead, so the backends review a diff that is not the one under audit and nothing in the result says so. Omit it only in the primary checkout. An unusable path is rejected with an error naming it, never silently replaced. |

<!-- moai:closure-second-review:start -->
**Contract-mode second review (card-bound).** When the reviewed card runs under a
contract-based autonomy workflow, invoke the tool with the card argument so this
fan-out is recorded as the card's second review:

- pass `card_id` set to the card identifier from the reviewed card's contract;
- keep `target: "baseBranch"` — the review must cover the reviewed scope, never
  uncommitted changes;
- run the review AFTER the last commit that changes the card's governed paths
  (the contract's ownership `write` globs, excluding the SPEC's own directory),
  so the recorded scope is current for the commit that will be judged; a review
  recorded before that commit is stale for the closure push.

The tool appends one second-review record into the card evidence directory.
Without `card_id` no record is written and the tool behaves byte-identically to
the pre-argument surface.
<!-- moai:closure-second-review:end -->

### Output shape

The tool returns a `ConvergenceResult`:

```json
{
  "per_backend_verdicts": [
    {"backend": "claude", "source": "mcp_claude_audit", "gate": "required", "verdict": "pass", "summary": "...", "findings": [], "next_steps": [], "provenance": {"transport": "claude-code-cli", "auth_mode": "subscription", "requested_model": "sonnet", "resolved_model": "claude-sonnet-...", "requested_effort": "high", "session_persisted": false}},
    {"backend": "codex",  "gate": "required", "verdict": "fail", "summary": "...", "findings": [...], "next_steps": []},
    {"backend": "glm",    "gate": "advisory", "verdict": "pass", "summary": "...", "findings": [], "next_steps": []}
  ],
  "overall_verdict": "fail",
  "disagreement_flag": true,
  "participant_count": 3,
  "residual_risk_note": "required-backend FAIL: codex; cross-model disagreement: pass=[claude(required), glm(advisory)] fail=[codex(required)]",
  "fail_open_backends": []
}
```

- `overall_verdict` ∈ `{pass, fail}` — the existing review-output values. No
  new enum (disagreement is a flag, not a verdict value).
- `next_steps` is populated ONLY where the backend itself produced it. A
  backend that returns a structured review carries its model's own list;
  a backend that answers in prose carries none, so its entry shows `[]` even
  on a `fail` with findings. An empty `next_steps` beside a non-empty
  `findings` is therefore the expected shape, not a truncated result — do not
  read the finding text as steps.
- `participant_count` is how many backends contributed a comparable verdict:
  every entry whose gate is not `off` and whose verdict is `pass` or `fail`.
  `inconclusive` entries (missing, unauthenticated, erroring) are
  evidence-of-absence, not participants, and do not count. The field is always
  present, 0 included — it reports the count; no minimum-participant policy
  acts on it.
- `disagreement_flag` is three-valued. `true` = a divergence was observed: a
  required split, an advisory-only conflict, or a single participant's
  intra-backend synthesis divergence (directly observed divergences are never
  discarded, even below 2 participants). `false` = 2+ participants were
  compared and none diverged. `null` = undetermined: fewer than 2 participants
  were compared and no divergence was observed, so neither "they agreed" nor
  "they disagreed" is a grounded claim. The null is explicit — the member is
  always present in the JSON, never an absent key.
- `residual_risk_note` describes the convergence outcome in prose (which
  backend(s) failed, or the shape of the split). Surface this in the audit
  report's residual-risk section.
- `fail_open_backends` lists the backends that returned `inconclusive` (missing,
  unauthenticated, or erroring) — surfaced so the report can name them.
- `plan_source` is `config` when any backend's gate came from the tree's
  configuration, and absent otherwise. A result without it from a tree whose plan
  has configured gates comes from a server that predates the plan: the check
  below reports an unmet gate.
- `source` distinguishes an in-session Claude anchor from a real
  `mcp_claude_audit` call. `provenance` identifies transport, subscription auth,
  requested/resolved model, effort, tool surface, persistence, usage source,
  and sanitized error code. Token counts are `null` when the CLI does not
  report them; they are never guessed.

### How to invoke

In a Claude main session, call the tool with the in-session analysis folded into
the `claude_verdict` object. Do NOT pass the full analysis text as prompt
context for the other backends — see the Independence rule below.

```
result = mcp__moai__audit_multi({
  claude_verdict: { verdict: <your verdict>, summary: <one-line>, findings: [...], next_steps: [...] },
  target: "uncommittedChanges",
  focus: "concurrency",
  project_root: <git rev-parse --show-toplevel>,
  session_id: <current session id>
})
```

In a GPT or GLM main session, omit `claude_verdict` (or treat it as ignored):

```
result = mcp__moai__audit_multi({
  target: "uncommittedChanges",
  focus: "concurrency",
  project_root: <git rev-parse --show-toplevel>,
  session_id: <current session id>
})
```

The launch provider decides the path; prompt text cannot impersonate a Claude
main session. Direct single-backend use is `mcp__moai__claude_audit` with the
same target/focus/project-root scope and optional model/effort override.

The orchestrator-side question channel is preserved: the tool returns a
structured result, never prompts the user. When a required backend is
inconclusive, surface the structured `overall_verdict: fail` plus
`residual_risk_note` in the audit report and let the orchestrator translate.

## Independence rule (load-bearing)

> **Pass only the synthesized `claude_verdict` object to the MCP tool — NEVER
> the full Claude analysis text as prompt context for the secondary backends.**

The external backends (Claude subscription, codex, GLM) are SUPER-REVIEWS:
uncorrelated second opinions. Their value collapses to a re-sample of another
model's reasoning the moment
they see Claude's analysis. The convergence engine enforces this structurally —
the `claude_verdict` is consumed ONLY as a Claude-main anchor. For GPT/GLM
origins it is ignored, and the backends receive `(target, focus, project_root)`
— a scope, an area name, and a directory, carrying no analysis between them.
The auditor must not undermine the invariant by pasting another backend's
reasoning into the `focus` field either.

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
- **An explicitly configured `required` gate left unmet fails overall.** A gate
  the project sets to `required` in `workflow.audit.gates` must actually hold:
  when that backend returns `inconclusive` (missing binary, auth failure, error),
  the engine fails `overall_verdict` and names the unmet backend in
  `residual_risk_note`. A gate the project never configured keeps the fail-open
  behavior — the distributed default is NOT an opt-in. In both cases the
  backend's own `per_backend_verdicts` entry stays `inconclusive` and its
  `fail_open_backends` listing stays, so the audit trail keeps saying the
  backend never ran.
- **Below 2 participants the flag is `null`, not `false` — unless a divergence
  was observed.** The case table presumes a comparable field of 2+; when
  `participant_count` is 0 or 1 (for example the only required backend to
  produce a verdict is claude), "no disagreement detected" is not a grounded
  claim and the flag reports `null`. The carve-out: an intra-backend synthesis
  divergence observed by even a single participant keeps the flag `true` —
  observed information is never discarded.

## Fail-open identity

Claude subscription, Codex, and GLM audit transports are fail-open. A missing,
unauthenticated, erroring, or malformed
backend yields `verdict: inconclusive` in its `per_backend_verdicts` slot and
convergence continues over the remaining active backends. The autonomous flow is
NEVER hard-blocked on a missing optional dependency — `evidence-of-absence ≠
evidence-of-failure`.

In a Claude main session, when all external backends are inconclusive, the
overall verdict can fall back to the in-session Claude anchor — EXCEPT for a gate
the project explicitly configured `required` in `workflow.audit.gates`: that
gate left unmet fails `overall_verdict` instead (see the convergence policy
above).

## Checking the result against the plan

After an `audit_multi` call on a tree whose plan lists `enforced_required`
backends, the result is checked before a verdict is reached:

1. Write the digest of the result just obtained to
   `<toplevel>/.moai/state/audit-plan-result.json` with a `Write` tool — fresh,
   immediately before the check: overwrite any earlier file at that name and never
   reuse a file from an earlier audit. The digest holds `overall_verdict`,
   `gate_unmet`, `plan_source` and, per `per_backend_verdicts` entry, `backend`,
   `gate` and `verdict` — digest members only, never summary or finding text.
2. Run `moai verify audit-plan --project-root <toplevel> --result-file <that path>`
   and pass only the path — never the JSON on the command line (the worktree guard
   refuses braces and quotes).
3. Read `convergence_check`: `ok: false` is an unmet gate named by backend, no PASS,
   reported as an unmet gate and not as a reviewed defect.

`plan-auditor` carries `Write` and does steps 1-3 itself. In the sync phase the
orchestrator does them: a cold `sync-auditor` is read-only, so it returns the digest
members in its report, says its verdict is not final until the orchestrator's check
passes, and the orchestrator writes the file and runs the check.

## Folding the result into the audit verdict

The auditor's verdict and the convergence result relate as follows:

| `overall_verdict` | `disagreement_flag` | Auditor action |
|---|---|---|
| `pass` | `false` | Standard PASS. No residual-risk row needed. |
| `pass` | `null` | Standard PASS. The flag is undetermined (`participant_count` below 2, no observed divergence) — not an agreement claim. No residual-risk row needed. |
| `pass` | `true` | PASS with a residual-risk row naming the advisory disagreement. |
| `fail` | (any) | FAIL. Name the failing required backend(s) from `per_backend_verdicts`. The block is conservative (cases 2/3). |

In all cases, surface `residual_risk_note` verbatim in the audit report's
residual-risk section so a human reader sees which backend disagreed with which.

## Cross-references

- `mcp__moai__claude_audit`, `mcp__moai__codex_audit`, `mcp__moai__glm_audit` — the single-backend
  tools. The convergence engine calls the same backends through its own fan-out
  and does NOT route through these handlers, so a behavior read from one surface
  must be confirmed on the other rather than assumed shared.
- `workflow.audit.gates.*` — the per-auditor gate map (`off`/`advisory`/`required`).
  An explicit `required` is enforced on BOTH surfaces: on the convergence result
  an unmet required gate (its backend `inconclusive`) fails `overall_verdict`,
  and the single-backend `codex_audit` tool likewise returns `verdict: fail`
  with a non-empty `gate_unmet` and `isError: false` when an explicitly required
  codex gate is left without a verdict. Absent keys fall back to the distributed
  defaults WITHOUT that enforcement — write the key to opt in.
- **Audit receipts.** Where the codex gate is explicitly `required`, the server
  records a receipt for every codex audit it performs and returns its id on the
  result as `audit_receipt`. Cite the ids you received in the verdict line that
  ends your report:

  ```
  AUDIT-VERDICT: <PASS|PASS-WITH-DEBT|FAIL> spec=<SPEC-ID> receipts=<receipt-id>[,<receipt-id>...]
  ```

  The line is the LAST non-empty line of the final message; `receipts=none` says
  no receipt was issued. A PASS the receipt store cannot corroborate is refused
  at subagent stop, and the phase-entry spawns stay denied until a PASS citing a
  valid receipt is recorded. The check reads the store, never the report text —
  quoting an id the store does not carry proves nothing.
- `workflow.multi.review_gate.enabled` — opt-in toggle for the multi-review-gate
  Stop hook (the Path C fully-autonomous gate). Default OFF; opt in via local
  config.
