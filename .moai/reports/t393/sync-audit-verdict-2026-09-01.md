# Sync-Phase Audit Verdict — SPEC-INIT-HARNESS-PROMPT-001 (card t393)

Auditor: sync-auditor (independent, replacement attempt — two prior sessions lost to backend 429s)
Tree: worktree `.claude/worktrees/t393`, branch `WT-init-harness-prompt`, HEAD `e6ef685bc`, fork `2c18091d1`
Date: 2026-09-01

## Overall Verdict: WARN

Substance PASS; one required citation-cell repair (F1) before develop integration. No FAIL-driving
defect; must-pass firewall (Functionality + Security) holds.

## Dimension Scores

| Dimension | Weight | Score | Verdict | Evidence (verbatim basis) |
|-----------|--------|-------|---------|---------------------------|
| Functionality (40%) | 40% | 92 | PASS | This run: `go test ./internal/cli/ -run 'TestInitAgentFlag\|TestValidateInitFlags_Agent\|TestRunInit_Agent\|TestRunInit_Codex\|TestRunInit_CallsCodexWiring' -count=1 -timeout 240s` → `ok github.com/modu-ai/moai-adk/internal/cli 1.968s` exit 0 (non-empty sweep — no `[no tests to run]`), at HEAD `e6ef685bc`. Relayed+attributed: orchestrator unfiltered two-package run exit 0 (wizard ok 3.189s / cli ok 304.797s) + vet 0 + golangci-lint "0 issues." at `5b280204e` (evidence: `.moai/reports/t393/orchestrator-verify-20260901.log`); `git diff --stat 5b280204e..HEAD` = 4 files, docs/evidence only — **zero Go code changed after that measurement**, so the attribution carries. |
| Security (25%) | 25% | 95 | PASS | Diff read: no new trust boundary, no secrets, no injection surface; the wizard input is a closed-set Select normalized through the single `normalizeAgentWiring` mapping (init.go:159-166); fail-loud `validateInitFlags` rejection path is inside the test set this auditor ran green. No OWASP probe run (no new web/input surface introduced) — basis is the full-branch diff read. |
| Craft (20%) | 20% | 82 | PASS (debt noted) | Code read directly: single resolution point (`resolveAgentWiringWithWizard`), documented precedence + reversal condition, test guards updated with verbatim pins (`unchanged from HEAD 2c18091d1`). Deductions: F1 citation defect in the audit-ready §E.2.3 signal; F3 §E.4 mx_validation undercount. Coverage % not re-measured (Gap). |
| Consistency (15%) | 15% | 90 | PASS | Commit subjects follow convention with card id in every body; `🗿 MoAI` trailer present (bbc06cbe5 show); spec.md frontmatter transition matches the canonical Status Transition Ownership Matrix (merged 3-phase close is the canonical pattern per spec-frontmatter-schema.md); §E.4 structure matches the canonical section map; CHANGELOG house style matches sibling t379/t380 entries. |

Weighted: 90.5 / 100. Harmonic mean: **89.5 / 100**. Must-pass firewall: Functionality PASS, Security PASS — no dimension override.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. AC-IHP-011's substance (fail-loud `invalid --agent value` rejection + landed `--agent` behaviour tests) is green on this tree; its §E.2.3 citation as written is vacuous.
2. The `git add -f` evidence-log tracking is a precedented, twice-disclosed deviation, not a hygiene violation.
3. The sync commit touched only `status:` in spec.md; the merged close is canonical and honestly recorded.
4. All three spot-checked CHANGELOG claims match the code; one is loosely worded.
5. The 6 added `@MX` diff lines reconcile to 6 real tags (3 prod + 3 test), same type/ID, 0 deletions; pre-existing tags untouched.

### Evidence
1. Citation cell as cited — progress.md:267 contains `-run 'TestInitAgentFlag\|TestValidateInitFlags_Agent\|…'` (backslash-pipe, read this session). Corrected selector re-measured by this auditor: command + output under Functionality row above. Vacuity of the cited form: relayed from two independent attributions (orchestrator direct observation; §E.4 progress.md:450-453 quotes `ok … 0.949s [no tests to run]`).
2. `git ls-files .moai/reports/ | grep -c '\.log$'` → 5: four pre-existing t222 logs (`.moai/reports/t222/failing-attempt1-excerpt.log`, `repro-race20.log`, `repro-race8-after.log`, `repro-race8.log`) + `t393/orchestrator-verify-20260901.log`. `.gitignore:106` = `*.log`. Deviation disclosed in bbc06cbe5 commit body and §E.4 `sync_session_gaps` (progress.md:468-469).
3. `git show bbc06cbe5 -- .moai/specs/SPEC-INIT-HARNESS-PROMPT-001/spec.md` → exactly one hunk: `-status: in-progress` / `+status: completed`; diffstat `2 +-` (1+/1-). §E.4 `frontmatter_status_transitions.spec_md` (progress.md:432-433): "in-progress → completed (merged 3-phase close on this sync commit; run-phase never took the `implemented` intermediate step — recorded as it happened)". `sync_commit_sha: bbc06cbe5` backfilled in e6ef685bc per the D3 placeholder-backfill exemption.
4. (a) internal/cli/init_agent_wizard.go:44-52 — `if flagChanged && flagValue != ""` wins; else `res == nil → agentWiringClaude`; flag default `""` per init.go:132. (b) fork `2c18091d1:init.go:166` = `func wireCodexUnlessClaude(cmd *cobra.Command, projectRoot string)` → HEAD init.go:191 = `(cmd *cobra.Command, wiring agentWiring, projectRoot string)`; body (192-197) decides on `wiring` only, `cmd` used solely for `OutOrStdout()/ErrOrStderr()` — no flag read behind the resolution point. (c) wizard test diff verbatim: `-if got != 16` → `+if got != 17` ("Quality & Workflow (11)"), `-got != 17` → `+got != 18` ("6 + 12 page-3"), plus ID-sequence table update.
5. `git diff 2c18091d1..HEAD | grep '^-.*@MX' | wc -l` → 0. Added tag locations (grep this session): `internal/cli/init_agent_wizard.go:20`, `internal/cli/init.go:753`, `internal/cli/init.go:965` (the three §E.4 names) **plus** `internal/cli/init_agent_wizard_precedence_test.go:11`, `internal/cli/init_agent_wizard_test.go:16`, `internal/cli/wizard/agent_wiring_question_test.go:11` — all `@MX:SPEC: SPEC-INIT-HARNESS-PROMPT-001`. Pre-existing `CATALOG-002` NOTEs and `SPEC-INIT-WIZARD-REPAIR-001` tags verified present at HEAD init.go (grep) and fork (git show) — untouched. (`init_agent_wizard.go` absent at fork — new file; its tag line is part of the whole-file addition.)

### Baseline-attribution
Every measurement above is this run against worktree `.claude/worktrees/t393` at HEAD `e6ef685bc` (branch `WT-init-harness-prompt`), 2026-09-01. The two test commands this auditor executed: the corrected selector (`ok … 1.968s`, exit 0) and the git/grep read-only batch. Orchestrator figures (unfiltered suite, vet, lint) are relayed with attribution to HEAD `5b280204e` — code-identity to `e6ef685bc` proven via `git diff --stat 5b280204e..HEAD` (4 files: .log, progress.md, spec.md, CHANGELOG.md; 0 Go files). Coverage and 100%-function-coverage claims on `resolveAgentWiringWithWizard` are relayed-unverified (prior auditor attempt 1).

### Gaps
- No CI verdict — the branch is unpushed; no remote run observed for any of the five commits.
- Cross-platform (GOOS matrix) build not measured locally.
- Full-catalog `moai spec lint` not run.
- Coverage % not re-measured in this audit; the 100% function-coverage claim on `resolveAgentWiringWithWizard` remains relayed.
- Exact per-test count ("8 tests") not enumerated by this auditor — my run proves a non-empty sweep (`ok`, no `[no tests to run]`), the count itself is orchestrator/sync-agent attributed.
- Unfiltered suites not re-run by this auditor (dispatch constraint: no package-wide runs; rate-limit replacement context).

### Residual-risk
- The unfiltered `internal/cli` green is a single-run observation on a loaded machine (384s in run-phase; §E.2 itself flags single-run flake risk). CI on `origin/develop` remains the real verdict surface after integration.
- The `\|` mangling mechanism (markdown table-cell pipe escaping) is the exact defect class verification-completeness.md §2.1 names; until F1 is repaired, any auditor reading §E.2.3 without reading §E.4 reads a vacuous PASS as a real one.
- `normalizeAgentWiring` silent-claude fallback (progress.md:400-404 residual): a future wizard option without a matching constant resolves to claude silently — recorded, unchanged by this audit.

## Findings

- F1 [SHOULD-FIX] [blocking-for-integration] `.moai/specs/SPEC-INIT-HARNESS-PROMPT-001/progress.md:267` — AC-IHP-011 citation cell carries `\|` table-cell escapes → `-run` selector matches zero tests → vacuous green as cited (defect class: report-not-verdict, named in verification-completeness.md §1.1/§2.1). Substance proven (this auditor's corrected-selector run green). Required fix: manager-develop one-cell repair BEFORE develop integration — replace the escaped selector with a plain-`|` selector, preferably moved into a fenced evidence-ledger block per §2.1's carrier recommendation, citing observed output (`ok … 1.968s`, exit 0) and tree SHA `e6ef685bc`.
- F2 [MINOR] [optional] `internal/cli/init.go:187-191` + CHANGELOG entry — "takes the already-resolved wiring **instead of the command**" is loose: `cmd` remains a parameter (output streams). The load-bearing property (no second flag read behind the resolution point) is verified true. Optional wording fix ("instead of reading the flag from the command").
- F3 [MINOR] [optional] `progress.md` §E.4 `mx_validation` (~line 440) — "added three @MX:SPEC tags" undercounts: six landed (3 production + 3 new test-file tags). Reconciles cleanly once counted; optional correction (manager-docs-owned section).

## Per-Item Dispositions (dispatch items 1-5)

1. **AC-IHP-011 citation cell → disposition (b): require manager-develop citation-cell repair BEFORE develop integration.** A known-vacuous citation in the audit-ready §E.2.3 signal is precisely the report-not-verdict class the project doctrine names; the repair is one cell plus a fenced citation, and leaving it costs every future auditor (this defect has now consumed three audit sessions' attention). Not a Functionality FAIL — the AC's substance is proven.
2. **`git add -f` evidence log → acceptable recorded deviation.** Four precedented tracked `.log` files under `.moai/reports/` (t222); deviation disclosed twice (commit body + §E.4); the tracked evidence log serves the verification-claim-integrity evidence-persistence obligation. The blanket `*.log` gitignore vs. evidence-need tension is a repo-level question, not this card's defect.
3. **Close-transition integrity → PASS.** Sync touched only `status:` (1+/1- shown); merged 3-phase close is the canonical pattern per the Status Transition Ownership Matrix; skipped `implemented` intermediate honestly recorded in §E.4 as it happened; `sync_commit_sha` backfill follows the D3 exemption.
4. **CHANGELOG accuracy → PASS (one wording caveat).** (a) precedence + empty-string fallthrough verified by code read; (b) wiring-over-command verified — body reads only the resolved wiring, `cmd` kept for writers (F2 wording); (c) cardinality guards 16→17 / 17→18 verified verbatim in the test diff.
5. **MX reconciliation → RESOLVED.** 6 added tag lines = 6 real tags (3 prod + 3 test), all `@MX:SPEC` with the same SPEC-ID, 0 deletions; pre-existing CATALOG-002 / SPEC-INIT-WIZARD-REPAIR-001 tags untouched. §E.4's "three" is an undercount (F3, optional).

## Recommendations

- Land F1 as a small manager-develop commit on this branch before the develop integration window; cite the corrected selector's output and pin the tree SHA.
- Fold F2/F3 corrections into the same window if cheap; neither blocks.
- After integration, CI on `origin/develop` is the outstanding verdict surface (the only Gap that can still flip this verdict).
