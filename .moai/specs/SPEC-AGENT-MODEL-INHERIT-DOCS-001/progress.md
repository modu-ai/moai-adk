# progress.md — SPEC-AGENT-MODEL-INHERIT-DOCS-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
phase: plan
plan_status: audit-ready
plan_complete_at: 2026-09-29
spec: SPEC-AGENT-MODEL-INHERIT-DOCS-001
card: t1300
status: draft
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
baseline_tree: 8a969dfc0
baseline_branch: WT-model-docs-sweep
red_now_recorded: acceptance.md §D baseline table (R1-R6)
req_count: 10
ac_count: 9
open_questions: none blocking plan-phase entry
```

Plan artifacts authored 2026-09-29 by manager-spec in worktree `.moai/worktrees/t1300`.
Scope decisions recorded in spec.md §2 (cli-reference/profile.md measured as surviving-surface
page → retain-in-place) and REQ-AMD-005 (default zero removals → zero redirects).
Delta revision v0.2.0 applied 2026-09-29 for plan-audit iter-1 (FAIL 0.75): D1 premise-collapse —
REQ-AMD-007 reconceptualized as a pure regression guard (H24 rewrite landed pre-baseline in
t1246 commit `41cf11c4d`; M4 demoted to verify-only; `translations.go` is a forbidden-edit file,
map keys MUST NOT be renamed); D2 outside-cluster dispositions added to M3; D4 inbound-link
enumeration corrected to 4 (tokenomics en 27/51/109/131).

## §E.2 Run-phase Evidence

M1 — A-cluster inheritance rewrite (5 pages × 4 locales = 20 files), authored ko→en→ja/zh in one work session. Evidence below captured on tree HEAD `620328f00` + M1 edits, worktree `.moai/worktrees/t1300`, 2026-09-29.

- AC-AMD-001(a) retired-marker grep (20 A-cluster files, all locales):

  ```
  $ grep -rn 'The three profiles\|Per-agent assignment table\|Model / effort\|モデル / effort\|模型 / effort\|Tier × Phase' <20 A-cluster files>
  (no output) exit=1
  ```

- AC-AMD-001(a2) `moai model profile` residue — 8 remaining hits, ALL inside the explicit history/retired sections of the rewritten model-policy/profile-matrix pages ("물러났습니다 / retired / 退きました" framing). No page presents it as live:

  ```
  ko/advanced/profile-matrix.md:88, ko/multi-llm/model-policy.md:166,
  en/advanced/profile-matrix.md:95, en/multi-llm/model-policy.md:175,
  ja/advanced/profile-matrix.md:53, ja/multi-llm/model-policy.md:112,
  zh/advanced/profile-matrix.md:53, zh/multi-llm/model-policy.md:112
  ```

- AC-AMD-001(b) inheritance-narrative markers per locale — each locale's native wording of "subagents inherit the main session's model and effort" present in ≥4 of 5 pages per locale (ko: 4, en: 4, ja: 4, zh: 4; the 5th page, tokenomics-overview, carries the equivalent session-inheritance wording in a different phrasing, verified by read).

- AC-AMD-001(c)/AC-AMD-008 4-locale file-existence + heading parity (5 pages):

  ```
  multi-llm/model-policy:    ko 9/5  en 9/5  ja 9/5  zh 9/5   (h2/h3 — full parity)
  advanced/profile-matrix:   ko 7/0  en 7/0  ja 7/0  zh 7/0   (full parity)
  advanced/no-haiku-3tier:   ko 9/3  en 9/3  ja 9/3  zh 9/3   (full parity)
  advanced/agent-guide:      ko 15/7 en 14/16 ja 14/16 zh 14/16 (PRE-EXISTING divergence, baseline-identical: git show HEAD counts ko 15/7, en 14/16 — M1 changed no heading on this page)
  advanced/tokenomics-overview: ko 8/4 en 9/4 ja 8/4 zh 7/4      (PRE-EXISTING divergence, baseline-identical: ko 8/4 en 9/4 zh 7/4 at HEAD)
  ```

- AC-AMD-008 body-emoji scan (20 files): `grep -rnP '[\x{1F300}-\x{1FAFF}...]'` → no output, exit=1 (clean). Icon shortcodes used; no body emoji introduced.
- AC-AMD-004 partial: no page removed — `find docs-site/content -name '*.md' | wc -l` = 620 (unchanged); `git diff docs-site/vercel.json` EMPTY.
- AC-AMD-006/007 partial (guards): `git diff -- internal/cli/wizard/translations.go` and `... profile_setup_translations.go` EMPTY (never touched).
- Locale chain honored: ko authored first per page, en derived, ja/zh derived in the same work session; all 20 files in one commit.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Input parameters: tier=M; scope≈52 files (A-cluster 20 + B-cluster 24 + C-cluster ~4 + Go verify-only 0-edit); domains=2 (docs-site, Go guard); file language mix=markdown-dominant; concurrency benefit=low (one file cluster, canonical-locale chain is ordered); agent-teams prereqs=not requested.

Mode evaluation: direct=not selected (semantic multi-file rewrite); fanout=not selected (same file cluster — one writer per tree; canonical locale chain ko→en→ja/zh is ordered, not parallel); sweep=not selected (semantic rewrite, not mechanical-uniform transform; Kickoff gate cleared but uniformity precondition fails); agent-team=not requested (explicit-request-only); serial=SELECTED.

Decision: serial

Justification: the run is semantic docs rewriting inside one file cluster with a fixed canonical-locale chain — parallel writers would collide on sibling locales of the same pages and the chain itself is ordered. M4 (verify-only Go regression guard) is one bounded check. Serial manager-docs spawn per milestone with lane verification between matches Anthropic's coding-task caution and the one-writer-per-tree rule (agent-common-protocol § Background Agent Execution). Kickoff approval: lead-approved per develop §31 (operator policy t1266) + operator blanket approval, relayed 2026-09-29. N1 (audit iter2): M4 verification MUST compare against `git show 8a969dfc0:internal/cli/wizard/translations.go` lines 417/429/441/453 payloads — NOT paste-grep of the quoted baseline in the SPEC text.
