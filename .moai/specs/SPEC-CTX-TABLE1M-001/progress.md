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

## §E.2 Run-phase Evidence

All commands below are single-invocation, run 2026-10-02 on worktree
t1415 against the tree they name. RED-now cells are the §E.1 ledger
(tree `c50da9c2f`); each row below is that pair's GREEN observation.
Final verification tree: HEAD `6b961a333` (working tree clean).

Commits (branch `WT-ctx-table-1m-fix`, base `c50da9c2f`):

| Commit | Subject | Files |
|---|---|---|
| `2153b2c39` | feat(SPEC-CTX-TABLE1M-001): plan-phase artifacts (Tier S scope, 3 artifacts per dispatch) | 6 SPEC artifacts |
| `d3c5d5829` | docs(SPEC-CTX-TABLE1M-001): correct 200K-sessions row for CC 2.1.285/2.1.287 gateway and provider defaults | spec.md status flip + cwm live/mirror |
| `6b961a333` | docs(SPEC-CTX-TABLE1M-001): correct gateway/provider 200K claims in model-policy | model-policy live/mirror |

No separate M3 commit: `make build` produced zero tracked-file delta
(`git status --short` → empty immediately after), so per plan §D the
embed refresh folds into M2.

AC matrix (GREEN cells; every value verbatim from this run):

| AC | Claim | Command (one row per distinct form; live/mirror pairs grouped) | Verbatim stdout | exit |
|---|---|---|---|---|
| AC-CTM-001 | PASS — cwm pair corrected | `grep -c "Opus 4.8+ running with a 200K window"` on cwm live and mirror | `0` / `0` | 1 / 1 |
| AC-CTM-001 | PASS | `grep -c "unless \`sonnet\[1m\]\` is selected"` on cwm live and mirror | `0` / `0` | 1 / 1 |
| AC-CTM-001 | PASS — new fact present | `grep -c "CC 2.1.285"` on cwm live and mirror | `1` / `1` | 0 / 0 |
| AC-CTM-001/005 | PASS — anchors intact | `grep -c "CLAUDE_CODE_DISABLE_1M_CONTEXT=1"` and `grep -c "takes the 200K row"` on cwm live | `1` / `1` | 0 / 0 |
| AC-CTM-002 | PASS — model-policy pair corrected | `grep -c "or with \`CLAUDE_CODE_DISABLE_1M_CONTEXT=1\`, \`sonnet\` budgets 200K"`, `grep -c "run with a 200K window on some providers"`, `grep -c "Opus on a 200K provider such as Amazon Bedrock"` on live and mirror (6 invocations) | `0` ×6 | 1 ×6 |
| AC-CTM-002 | PASS — expected 3 citations (lines 22/66/90) | `grep -c "CC 2.1.285"` on model-policy live and mirror | `3` / `3` | 0 / 0 |
| AC-CTM-002 | PASS — flag anchor intact | `grep -c "CLAUDE_CODE_DISABLE_1M_CONTEXT"` on model-policy live | `3` | 0 |
| AC-CTM-003 | PASS — byte parity ×2 | `cmp` live vs mirror, cwm pair and model-policy pair | (no output) ×2 | 0 / 0 |
| AC-CTM-004 | PASS — 7,098 ≤ 7,099 (baseline 6,999) | `LC_ALL=C wc -c .claude/rules/moai/workflow/context-window-management.md` | `    7098 .claude/rules/moai/workflow/context-window-management.md` | 0 |
| AC-CTM-004 | recorded, no duty (`paths:`-scoped) | `LC_ALL=C wc -c .claude/rules/moai/development/model-policy.md` | `   27688 .claude/rules/moai/development/model-policy.md` | 0 |
| AC-CTM-006 | PASS — embed refresh | `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && make build` (single compound invocation; output redirected, bounded tail) | `exit=0`; build stamp `Commit=6b961a333` (built FROM this tree); `catalog.yaml updated successfully (13900 bytes)` | 0 |
| AC-CTM-007 | PASS — neutrality ×8 | `grep -c "t1415"` and `grep -c "SPEC-CTX-TABLE"` on all 4 files | `0` ×8 | 1 ×8 |
| AC-CTM-008 | PASS — docs-only scope | `git status --short` (post-build); `git diff --name-only c50da9c2f..HEAD` | (empty); exactly 10 files — the 4 targets + 6 SPEC artifacts, zero `.go` paths | 0 / 0 |

Procedure note (honest deviation from the dispatch ORDER summary): the
dispatch ORDER sketched M3 as `cp` mirror→live, but plan §A/§F R1
("apply the identical Edit so the diff stays reviewable — do not fix by
copying the mirror over the live file") is the authoritative edit
mechanics the mission mandates, so both copies received the identical
Edit-tool replacements and `cmp` proves the same byte-identical end
state the cp path would have produced. No `moai update` was run.

Baseline-attribution: every GREEN value above was observed in this run
against this worktree's tree (`6b961a333` for the final batch; the M1
batch at `2153b2c39`+working, the M2 batch pre-commit at `d3c5d5829`
+working — same content that then landed in the cited commits; the
final batch re-ran every form on the committed tree and matched).

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-10-02T06:24:31Z
- run_status: audit-ready
- run_evidence: this §E.2 (all 8 ACs PASS; 2 release-blocking ACs
  flipped RED→GREEN with the §E.1 pairs complete; 6 guards held)
- run_commits: 2153b2c39 (plan artifacts) · d3c5d5829 (M1 + status
  transition, Authored-By-Agent: manager-develop) · 6b961a333 (M2;
  M3 embed refresh folded — no tracked delta)
- open_blockers: none
- sync handoff: manager-docs owns CHANGELOG/README/docs-site sync and
  the `implemented → completed` transition on the single sync commit
  (§E.4 `sync_commit_sha` remains manager-docs' field — untouched here).

## §E.4 Sync-phase Audit-Ready Signal

- sync_complete_at: 2026-10-02T06:33:43Z
- sync_status: audit-ready
- sync_commit_sha: 5a8893e510630fcbcf9e6f50ecc3b8ab2c638c3d  # D3-exempt backfill: the pending-backfill-sync placeholder in the sync commit replaced with the real SHA by the phase-owning agent (spec-frontmatter-schema § SHA placeholder backfill exemption)
- sync scope (single sync commit): CHANGELOG.md `### Fixed` entry (card t1415,
  B12 duplicate pre-check measured 0) + spec.md frontmatter
  `in-progress → implemented → completed` merged transition (status + updated
  only — SPEC body untouched) + this §E.4 signal
- No-op dispositions (Step 4, recorded): MX tag validation is a no-op —
  doc-only change, zero code symbols touched (4 markdown files: 2 rule files +
  2 byte-identical template mirrors), so no @MX annotation surface exists.
  Codemaps rotation is a no-op for the same reason — no Go source, no package
  structure change. README/docs-site sync is out of scope: the corrected files
  are internal dev-facing rule files under `.claude/rules/` (+ their
  `go:embed` template mirrors); user-facing product docs are unaffected.
- open_blockers: none
