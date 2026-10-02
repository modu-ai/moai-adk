# SPEC-CTX-TABLE1M-001 — Progress

> SPEC: SPEC-CTX-TABLE1M-001 — Context-window 1M-default correction (CC 2.1.285/2.1.287)
> Card: t1415 (leader direct dispatch; queue pick 2026-10-02T04:20:55Z) · Tree: .moai/worktrees/t1415 · Branch: WT-ctx-table-1m-fix
> This file is maintained by the lane under record authority; the §E.1 evidence is harvested from the manager-spec plan-phase return.

## Phase 1 SKIP Rationale

Operator card t1415 pre-scoped the targets (exact files and lines) and the lane measured the stale-claim inventory plus the upstream changelog facts before dispatch — no separate exploration round was needed.

## Mode Selection (Phase 4)

- Input parameters: tier S · scope 4 files (2 rule files + 2 template mirrors) · 1 domain (documentation) · markdown-only · concurrency benefit LOW
- Evaluation: direct (no — SPEC artifact set, multi-file), fanout (no — doc/coding-heavy serial work), sweep (no — 4 files, non-mechanical prose edits)
- Decision: serial
- Justification: doc correction with deterministic replacement text; a single sequential specialist per phase is the safe default (Anthropic coding-task parallelism caveat).

## Phase 6 (Deep Research) disposition

No separate research.md: the upstream facts (CC 2.1.285/2.1.287 verbatim bullets, fetched 2026-10-02 from https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md) are folded into spec.md §Background and plan.md, and the file-state measurements are the §E.1 ledger below. Skip recorded per plan.md (Phase 6 is recommended, not mandatory). Repair iteration 1 (2026-10-02) added research.md per coordinator instruction — it consolidates the upstream bullets, the source URL, and the file-state measurements; the no-separate-research statement above describes iteration 0 and stands as history.

## §E.1 Plan-Phase Evidence (harvested from manager-spec return; baseline tree c50da9c2f, worktree clean)

Scope found: 5 lines / 6 stale claims across the two rule files (each live + template mirror):

- `context-window-management.md:16` — 2 claims (Bedrock-200K clause + gateway `unless sonnet[1m]` clause)
- `model-policy.md:22` — gateway-sonnet 200K claim, alias-entry phrasing
- `model-policy.md:28` — provider-window paragraph (cross-reference target of line 66)
- `model-policy.md:66` — gateway + Bedrock 200K enumeration
- `model-policy.md:90` — dispatch-named `[1m]` surviving-paths bullet

Line 43 deliberately excluded (credit-entitlement claim, not a window-default claim; documented in spec.md §3.1). Line 68 stays consistent after the corrections.

RED-now ledger (28 entries E1–E28, single-invocation form, 4-element carriers), key rows:

| Entry | Command (abridged) | stdout | exit |
|---|---|---|---|
| E1/E2 | `grep -c "Opus 4.8+ running with a 200K window"` cwm live/mirror | 1 / 1 | 0 |
| E3/E4 | `grep -c "unless \`sonnet\[1m\]\` is selected"` cwm live/mirror | 1 / 1 | 0 |
| E5/E6 | `grep -c "CC 2.1.285"` cwm live/mirror | 0 / 0 | 1 |
| E7/E8 | `grep -c "or with \`CLAUDE_CODE_DISABLE_1M_CONTEXT=1\`, \`sonnet\` budgets 200K"` mp live/mirror | 2 / 2 | 0 |
| E9/E10 | `grep -c "run with a 200K window on some providers"` mp live/mirror | 1 / 1 | 0 |
| E11/E12 | `grep -c "Opus on a 200K provider such as Amazon Bedrock"` mp live/mirror | 1 / 1 | 0 |
| E13/E14 | `grep -c "CC 2.1.285"` mp live/mirror | 0 / 0 | 1 |
| E15/E16/E19 | anchors (cwm flags=1, mp flags=3, leading sentence=1) | recorded | 0 |
| E17/E18 | `cmp` live/mirror ×2 | (no output) | 0 |
| E20–E28 | `wc -c` (6999) + neutrality greps ×8 (card/SPEC-ID absence) | recorded | 0/1 |

Byte arithmetic (iteration 1 re-measure): cwm 382→481 bytes (+99 B byte-measured, always-loaded, ≤7,099 B budget AC, expected 7,098 B); model-policy +637 B total (`paths:`-scoped, no budget duty). Deterministic replacement drafts: `/tmp/t1415-drafts/` (recorded verbatim in plan.md §B). `make build` verification tagged VK-BUILD in acceptance.md.

Audit-ready signal (iteration 2 delta re-audit PASS — signal ON):

- plan_complete_at: 2026-10-02T05:05:09Z (iteration-0 authoring; repair iteration 1 same day)
- plan_status: audit-ready
- plan_audit_verdict: PASS 0.96/1.00 (iteration 2, supersedes iteration 1 FAIL 0.88; `VERDICT: PASS (iteration 2)` at .moai/reports/t1415/plan-audit.md:245, tree c50da9c2f)

## Decision Point 1 disposition

Autonomous proceed: the card was dispatched by the leader on the operator's pick with cmd `/moai plan (카드 본문 기준)`; the card names no operator gate. Per AGENTS.local.md §31 and auto-semantics.md §9.1, the plan→run Kickoff takes its default autonomous form after the independent audit cross (plan-auditor verdict + audit-ready signal + plan-artifact hash integrity + no open blocker), with the decision record written in this file.

## Plan-audit

- [x] Iteration 1 (2026-10-02): **FAIL 0.88/1.00** — single must-pass failure
  MP-3 (frontmatter `id` fails the enforced SPEC-ID shape; report Evidence
  E-B). All other must-pass criteria PASS; 14/14 RED cells re-executed and
  reproduced by the auditor. Report: `.moai/reports/t1415/plan-audit.md`
  (tree `c50da9c2f`).
- Repair iteration 1 applied (2026-10-02): D0 rename + HISTORY provenance
  (operator decisions 1-2); D1 AC-CTM-008 green cell de-ordering-sensitized;
  D7 REQ-CTM-008 added, cited by AC-CTM-007/008. Optionals applied: D2
  (`/autocompact 200k` caveat in the C1 row draft), D3 (provider-surface
  note, per-surface model lists kept changelog-faithful — the suggested
  merged wording would have overstated the provider case for Sonnet), D4
  (Claude apps gateway in P3), D6 (never-`moai update` in plan §E). D5
  recorded: acceptance.md Residual risks + decision-index.md Q1.
- Delta re-audit (2026-10-02): **PASS 0.96/1.00** (iteration 2, supersedes
  iteration 1). Blocking set D0/D1/D7 closed with fresh mechanical evidence;
  both repair deviations accepted (D7 REQ-CTM-008 at the exact 8-REQ Tier S
  ceiling; D3 per-surface faithful form — the auditor withdrew its own
  iteration-1 merged wording as changelog-inaccurate for Sonnet on providers).
  Byte arithmetic independently verified: cwm row 481 B, file projection
  7,098 ≤ 7,099 budget. No new defects. Report updated in place:
  `.moai/reports/t1415/plan-audit.md:245`.

## Plan→Run Kickoff — Decision Record (autonomous form)

Per AGENTS.local.md §31 and auto-semantics.md §9.1 (default autonomous
transition), the lane records the Kickoff decision:

- Independent plan-audit verdict: **PASS 0.96/1.00** ≥ Tier S threshold 0.75
  (iteration 2; report cited above).
- Audit-ready: recorded in §E.1 above (`plan_status: audit-ready`).
- Artifact-hash integrity: shasum snapshot of the 6-artifact set taken
  2026-10-02 immediately after the PASS verdict, tree `c50da9c2f` —
  acceptance `e612ddb9d325`, decision-index `c0c665cbdfee`, plan
  `99a541e3f781`, progress `70ce36df22ee`, research `c7b7964c4477`, spec
  `686fb63056a1`. No writer touched the SPEC directory between the verdict
  and this snapshot (the lane stayed read-only while the auditor worked);
  the run phase re-checks this baseline before its first commit.
- Open blockers: none (decision-index Q1/Q2 `EVIDENCE-NEEDED` items are
  recorded residual questions, explicitly non-blocking per the verdict).
- Decision: **ENTER RUN PHASE autonomously.** Mode: serial (§ Mode
  Selection). Delegation: per-spawn general-type worker executing the
  manager-develop role (worktree-pinned card work; manager-develop-type
  spawns auto-isolate to their own L1 tree and are not used for pinned work).
