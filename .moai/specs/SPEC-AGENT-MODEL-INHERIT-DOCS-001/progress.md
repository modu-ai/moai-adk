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

M2 — B-cluster deprecated-surface docs (6 pages × 4 locales = 24 files), authored ko→en→ja/zh in one work session (resumed once after a transient 429; tree unchanged at HEAD `5f4f199ef` across the interruption). Evidence captured 2026-09-29.

- Measured CLI surface before writing (commands + verbatim outputs):
  - `grep -n 'model-policy' internal/cli/init.go` → lines 105-109: `--model-policy`, `--high`, `--medium-alias`, `--low`, `--profile` all registered with `deprecatedAgentModelFlagUsage`; line 355 `warnDeprecatedAgentModelFlags(cmd, "profile", ...)` — warning points at `moai profile setup`.
  - `sed -n '415,460p' internal/cli/wizard/translations.go` → `model_policy` / `effort_level` / `model` keys carry the "Session model policy / Session effort level / Default model override" wording in 4 locales — these serve the PROFILE SETUP wizard screen (per-profile fields, "launched with this profile"), which is what the docs now document.
  - `sharedInitRemovedIDs` (question_removal_test.go:33) = `["project_name", "report_format"]` — the init wizard's model question is retired (per SPEC §2, t1246 commit 41cf11c4d).
- AC-AMD-002: deprecated-stub framing present per locale — the 5 flags documented as stubs that print a warning naming `moai profile setup`; example fences no longer pass `--model-policy medium` / `--profile medium`; the retired `Select model policy:` wizard screen and tier tables removed from all 4 locales. Residue grep `Select model policy|모델 정책 선택:|モデルポリシーを選択|选择模型策略|Max - Opus 5.5` → single hit zh faq.md:80, which is the FAQ question heading (parallel to en "How do I choose a model policy?"), not the retired wizard screen.
- AC-AMD-003: cli-reference/profile.md retained in all 4 locales, rewritten in place — wizard config items now list "default model override / session model policy / session effort level" (matching the measured translations.go) plus a session-inheritance callout; `grep -rn 'moai model profile'` over the B-cluster → no output, exit=1 (zero dead-command documentation); no vercel.json diff.
- AC-AMD-008: 4-locale file existence 24/24; heading parity — cli 17/13 all four, introduction 15/12 all four, faq 8/6 all four, what-is-moai-adk 18/24/9 all four, cli-reference/profile 8/4 all four; init-wizard ko 10/8 vs en/ja/zh 11/8 — PRE-EXISTING at baseline (git show HEAD: ko 10/8, en 11/8; M2 changed no headings).
- AC-AMD-008 emoji scan: only pre-existing hits — 🗿/🔅 inside statusline example output blocks (i18n rule §4 preserved branding examples), 🔴🟠🔵⚪ cost-color glyphs in what-is tables (7 per locale, count unchanged); zero new emoji introduced.
- Locale chain honored: ko first, then en, then ja/zh, same work session; 24 files + progress.md in one commit.

M3 — C-cluster treatment + outside-cluster dispositions. Evidence captured on merged tree `4321c28fe`, 2026-09-29.

- AC-AMD-005 prompt-caching pair (before → after):
  - Before: `grep -cin 'per-spawn model injection'` — cost-optimization en=1 (~line 216-221 "two devices" pair: profile-matrix pinning + per-spawn injection) ; ko/ja/zh carry the same pair in native phrasing (measured: ko 203 "에이전트별 모델 주입", ja 147 "エージェント別モデル注入", zh 147 "逐智能体的模型注入"); plus the Model Policy link description teaching "per-agent model injection" in all 4 (en 308, ko 286, ja 204, zh 204).
  - After: pair rewritten ×4 to the inheritance narrative ("subagents inherit the session's model and effort, so all spawns share one cache; the retired devices are gone; the remaining rule is don't switch the session model mid-session") + link descriptions reworded to "session model policy and effort fallback". Post-edit residue grep over all 4 caching pages: only retired-framing mentions remain (zero live-teaching).
  - context-memory sibling (`claude-code/context-memory/prompt-caching.md`): 0 injection-related hits in ALL FOUR locales (measured with en + native patterns) → recorded NO-CHANGE.
- Outside-cluster dispositions (measured per line, ×4 locales where the sentence exists):
  - REWRITE: `advanced/config-sections.md` — the whole llm.yaml section taught retired config keys; measured against the shipped template (`internal/template/templates/.moai/config/sections/llm.yaml`: "the former per-agent profile matrix (profile / profiles / performance_tier / harness_agents / agent_overrides) is retired; `moai update` strips these keys") → section rewritten ×4 to the current keys (harness/team_mode/claude_bin/glm) with the retirement note and `moai profile setup` pointer.
  - REWRITE: `advanced/autonomy-tier.md` 117 — link description "the single matrix for choosing each agent's {model, effort}" taught the retired matrix → reworded to "what replaced the former model-assignment matrix: session inheritance" ×4.
  - REWRITE: `advanced/_index.md` 50 — table caption "13 agents × {model, effort} across 39 cells" → "what replaced the model-assignment matrix — session inheritance" ×4.
  - NO-CHANGE: `advanced/autonomy-tier.md` 16/116 — the "model tier" orthogonality explanation and its link description make no per-agent assignment claim (the "user picks a model tier" sentence holds at session level); verified by read.
  - NO-CHANGE: `advanced/self-evolving.md` 96 — link description "the model architecture substrate" is historical framing, no per-agent claim.
  - NO-CHANGE: `advanced/token-budget.md` 107 — link description "the model-policy foundation of Layer B routing" names the linked page without teaching per-agent assignment.
  - NO-CHANGE: `advanced/_index.md` 49 — "DeepSWE-leaderboard rationale and the 3-tier policy" describes the (now intent-framed) 3-tier page, no per-agent claim.
  - NO-CHANGE: `multi-llm/_index.md` en/ja/zh — zero hits for `moai model profile` / `profile matrix` / プロファイルマトリクス / 配置矩阵 (measured); only the ko page taught the retired narrative.
  - TREAT: `multi-llm/_index.md` ko — measured retired teaching (matrix column narrative, `moai model profile --json` inspection, spawn-time injection/drift story, 33-cell references) → matrix section replaced with the session-selection narrative (model/effort/session-model-policy table + revised mermaid TD), model-policy section rewritten to "세션이 정하고 에이전트가 따른다", description/links updated. en/ja/zh siblings have no such content, so the 4-locale same-change obligation is satisfied vacuously (recorded disposition, not an omission).
  - NO-CHANGE: statusline, moai-web-console, decision-memory, agent-teams, best-practices, features-overview — measured grep (`moai model profile|profile matrix|per-agent|agent별 모델|エージェント別のモデル|逐智能体`) returns zero hits on all of them.
- AC-AMD-008 on M3-touched pages: heading parity — prompt-caching 13/3 ×4, config-sections 12/0 ×4, autonomy-tier 7/0 ×4, advanced/_index 2/4 ×4; multi-llm/_index ko 7/4 vs en/ja/zh 4/1 is a pre-existing locale-content divergence (the ko page is a fuller section index; en/ja/zh carry different thinner content, verified at baseline). Emoji scan on all 17 M3 files: 0 hits each.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Input parameters: tier=M; scope≈52 files (A-cluster 20 + B-cluster 24 + C-cluster ~4 + Go verify-only 0-edit); domains=2 (docs-site, Go guard); file language mix=markdown-dominant; concurrency benefit=low (one file cluster, canonical-locale chain is ordered); agent-teams prereqs=not requested.

Mode evaluation: direct=not selected (semantic multi-file rewrite); fanout=not selected (same file cluster — one writer per tree; canonical locale chain ko→en→ja/zh is ordered, not parallel); sweep=not selected (semantic rewrite, not mechanical-uniform transform; Kickoff gate cleared but uniformity precondition fails); agent-team=not requested (explicit-request-only); serial=SELECTED.

Decision: serial

Justification: the run is semantic docs rewriting inside one file cluster with a fixed canonical-locale chain — parallel writers would collide on sibling locales of the same pages and the chain itself is ordered. M4 (verify-only Go regression guard) is one bounded check. Serial manager-docs spawn per milestone with lane verification between matches Anthropic's coding-task caution and the one-writer-per-tree rule (agent-common-protocol § Background Agent Execution). Kickoff approval: lead-approved per develop §31 (operator policy t1266) + operator blanket approval, relayed 2026-09-29. N1 (audit iter2): M4 verification MUST compare against `git show 8a969dfc0:internal/cli/wizard/translations.go` lines 417/429/441/453 payloads — NOT paste-grep of the quoted baseline in the SPEC text.
