# t78 Review Verdict — PASS

- Card: t78 — audit_multi cross-model convergence wiring into `/moai review` verdict step
- Commit: `0ba4fa466` (branch `WT-t78`, merge-base `release/v3.1.1` @ `97daa3baf` — verified)
- Lens: default 4-perspective + self-reference extra scrutiny (dispatcher-requested)
- Reviewer session: release-v311 (2acd4be4), 2026-08-17

## Claim (주장)

The wiring is correct, fail-open, mirror-parity, neutral, and guarded; the guard's
content-assertion depth is proportionate (dispatcher asked for explicit judgment).

## Evidence (증거)

| Check | Command (own worktree, shared ODB) | Observed |
|---|---|---|
| File set | `git show --stat 0ba4fa466` | 3 files: review.md ×2 + `cross_model_audit_wiring_test.go` — matches claim |
| Byte-identity | `git show 0ba4fa466:<src> > a; git show 0ba4fa466:<tmpl> > b; cmp a b` | `BYTE-IDENTICAL: source == template` |
| Neutrality | `git show 0ba4fa466:<tmpl> \| grep -nE 'SPEC-\|REQ-\|/Users/\|CLAUDE.local\|date'` | only `updated: "2026-02-21"` — verified pre-existing in parent `97daa3baf` (same line 11); **0 new forbidden tokens** |
| Guard test re-run | `go -C <t78wt> test -run TestReviewWorkflowCrossModelConvergenceWiring ./internal/template/ -v` | PASS (source + template subtests, 0.00s) — reviewer-executed, not trusted from dispatch |
| SSOT field names | `grep` of `.claude/skills/moai-ref-cross-model-audit/SKILL.md` | `per_backend_verdicts` / `overall_verdict` / `disagreement_flag` / `residual_risk_note` / `fail_open_backends` / target enum `uncommittedChanges\|baseBranch` / session_id Stop-hook persistence — all present and semantically matching the compact table |
| Phase 2 mirror | `git show 0ba4fa466:...review.md \| grep 'verdict ownership'` | referenced HARD rule exists verbatim (line 71); Phase 3.5 owner clause (line 172) mirrors it faithfully |
| Contract semantics | full diff read (`commit-0ba4fa466.diff`, persisted here) | gating on `audit_model`; fail-open incl. MCP-server-absent case; `--lean` skip; `--deep` panel untouched; report-block omission semantics; step renumbering consistent |

## Baseline-attribution (baseline 귀속)

All rows above measured this session against commit `0ba4fa466` blobs (shared git
object database) and the live `t78` worktree for the test run. Diff artifact:
`.moai/reports/t78/commit-0ba4fa466.diff` (15,187 bytes).

## Findings (advisory — none blocking)

1. **Guard depth judgment (dispatcher's explicit question): adequate at this layer.**
   The content-assertion guard catches the realistic regressions for a doc-wiring
   change: silent removal, source↔template drift, missing skill-load instruction,
   missing fail-open doc. An E2E of the MCP call is NOT warranted here — the
   convergence tool's Go implementation owns behavioral tests, backends are
   fail-open by design (absent codex/glm ⇒ `inconclusive`), and a live-backend E2E
   would be flaky exactly where the contract says fall open. Cheap hardening
   available (same mechanism): also assert the carve-out/gating tokens
   (`--lean`, `--deep`, `audit_model`) so a future edit cannot silently remove the
   mode carve-outs the commit message promises.
2. **Cost observation (not a defect):** `audit_model` default is `codex+glm`
   (moai-mcp-tools.md), so Phase 3.5 is the de-facto default path wherever backends
   are installed — consistent with plan/sync auditors already declaring the tool;
   aligns `/moai review` with the same default rather than introducing a new cost class.
3. **Pre-existing cosmetic:** template review.md frontmatter `updated: "2026-02-21"`
   is stale — NOT introduced by this diff (present in parent). Candidate for a
   future doc-hygiene card; neutrality CI evidently tolerates it today.

## Gaps (미검증)

- `make build` EXIT=0 and the dispatcher's "wiring tests 3/3" figure are taken
  from dispatch evidence; my re-run covered `TestReviewWorkflowCrossModelConvergenceWiring`
  (2 subtests). The sibling tests (`TestCrossModelAuditConsumerLink`,
  `TestPlanAuditorHasNoStaticSkillPreload`) were not re-run — CI on the pushed ref
  is the closing authority.
- t78 working-tree cleanliness not directly observed (cross-tree `git -C` refused
  by the worktree-session guard); review is anchored on the commit object, which
  is what merges.

## Residual-risk (잔여 위험)

- Semantic drift between the compact call table and the SSOT skill is possible on
  future skill edits; the guard does not pin the table contents (see finding 1).
- The wiring's behavioral effect (reviewer actually invoking convergence) remains
  prompt-level — first live `/moai review` run under `audit_model: codex+glm`
  will be the true shakedown.

**Verdict: PASS** — integrate per the self-serve release-branch rule.
