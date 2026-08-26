# Sync-Audit Verdict — SPEC-V3R6-WORKFLOW-DOCS-001 (card t273)

- Auditor: sync-auditor (independent, fresh context)
- Tree: worktree t273, branch `WT-workflow-docs`, HEAD `4f39b8b82` (unpushed; audited against the tree)
- Content surface identical to evidence tree `d6927f855` (verified: `git diff --stat d6927f855..HEAD -- docs-site README*` = empty)
- Date: 2026-08-26 (KST)

## Verdict

**Overall: FAIL (narrow — one blocking finding, one-word fix)**

Must-pass firewall trips: Functionality dimension FAIL (AC-001 en heading-token leg). All other gates, regression checks, canon cross-reads, and measured CHANGELOG claims reproduce exactly. The defect delta for the confirming re-audit is AC-001's four locale greps only.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 88/100 | FAIL | 10/11 ACs fully reproduce; AC-001 en leg: `grep -c "Card Classes" content/en/advanced/kanban-mode.md` → **0** (actual heading `Card classes`, line 251). ko/ja/zh heading tokens 1 each; Class A/B/C 1(±)/1/1 per locale. |
| Security (25%) | 100/100 | PASS | Canonical blacklist grep over content+README×4 → exit 1 (0 matches). Secrets scan (ghp_/sk-/AKIA/xoxb/key=) → exit 1. External-URL scan on 8 new pages + 4 kanban pages → 0. Internal paths: only sanctioned `workers.json` (REQ-WFD-003) + product-public `.moai/specs/` layout; zero `.claude/rules|agents|worktrees` / `internal/*` leaks. |
| Craft (20%) | 96/100 | PASS | H2 parity on 3 touched pages exact ×4 (factory-mode 8, kanban-mode 13, spec-lifecycle 6). Parenthesized tokens natural: ko "세 클래스(Class A · Class B · Class C) 중 하나로" (:253), ja "3つのクラス(Class A · Class B · Class C)のいずれかに" (:253), zh "三个类别(Class A · Class B · Class C)之一" (:253). Mermaid TD-only; icon shortcodes + new-badge v3.2; register consistent per locale; body-emoji 0. |
| Consistency (15%) | 98/100 | PASS | 3-phase table = spec-workflow.md Phase Overview verbatim (commands/agents/30K-180K-40K); thresholds 0.75/0.80/0.85, artifacts 2/3/5, REQ/AC ceilings 8/16/25 (independent, not sum) all exact; Class A/B/C semantics faithful incl. Class B "skips plan, not the review" + evidence-to-progress-record + "speed is effect, not reason"; factory cap 10 code-grounded (`internal/config/defaults.go:439` DefaultLaneMaxConcurrentSubagents=10; :431 DefaultFactoryLeadWorkers=1); workers.json, socket dirs (`bootstrap.go:307-308`), sentinel (`factory.go:44`), `todo next` all verified in code. CHANGELOG "29 files, +1051/−128" reproduces exactly. −2 for the "11 PASS" count claim (see F2). |

**Harmonic mean: 95.3** — overridden by the must-pass firewall (Functionality FAIL → overall FAIL).

## AC Re-Run Matrix (all commands re-executed on this tree)

| AC | Re-run result | Status |
|----|---------------|--------|
| AC-001 | Class A 1/1/1/1 · Class B 1/1/1/1 · Class C 1/1/2/1; heading tokens: ko `카드 클래스` 1 · **en `Card Classes` 0** · ja `カードクラス` 1 · zh `卡片类别` 1 | **FAIL (en leg)** |
| AC-002 | README heading tokens 1/1/1/1 | PASS |
| AC-003 | 4 factory-mode.md exist (ko 10,751B); workers.json 2/2/2/2 | PASS |
| AC-004 | CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS 1/1/1/1 | PASS |
| AC-005 | factory-mode in kanban-mode 2/2/2/2 | PASS |
| AC-006 | 4 spec-lifecycle.md exist (ko 8,900B) | PASS |
| AC-007 | Implementation Kickoff 2/3/2/2; 0.75=2×4; 0.80=2×4; 0.85=2×4 | PASS |
| AC-008 | spec-based-dev→spec-lifecycle 1×4; spec-lifecycle→spec-based-dev 3×4 | PASS |
| AC-009 | per-directory anchors (D11 style): `"factory-mode":` in advanced/_meta ×4 = 1 each; `"spec-lifecycle":` in core-concepts/_meta ×4 = 1 each; cross-directory leakage negative control = 0×8; main.yaml `ref: /advanced/factory-mode` :728, `ref: /core-concepts/spec-lifecycle` :144; both name maps carry all 4 locale keys | PASS |
| AC-010 | 152/152/152/152 | PASS |
| AC-011 | sync-auditor 3×4; Functionality/Security/Craft/Consistency 1×4 each | PASS |

## Regression gates (re-executed)

- hugo `-s docs-site --minify --gc` exit 0, WARN/ERROR grep = 0, all 4 locale sites built, sitemap (multilingual index) present; 8 new-page HTML artifacts confirmed in public/
- README H2 parity 12/12/12/12
- Page counts 152×4
- Ratchet (canonical `^#\{2,\} ` recipe from hns-oss-docs-verify): divergence set = 54 = baseline 54; `comm -23` new-divergence = **empty** — §E.2 claim reproduced exactly. (Auditor note: an initial H2-only measurement produced 4 false "new divergences"; all 4 pre-exist at base `db1362739` with identical counts and converge under the canonical H2-and-deeper regex. Not a defect.)
- URL blacklist 0; secrets 0; Mermaid LR/RL 0; body-emoji 0

## Findings

- **F1** [SHOULD-FIX] [blocking] `docs-site/content/en/advanced/kanban-mode.md:251` — AC-001's normative en heading token `Card Classes` greps 0; actual heading is `Card classes` (sentence case). Original to M2 `aaad6fcb6` (verified via `git show`), not a regression from the late `d6927f855` commit. Required fix: retitle en H2 to `## Card Classes — not every card needs every column` (one word; also aligns with README.md en "Card Classes"), OR amend the acceptance token via manager-spec if sentence-case was the intended docs-site en convention. Either path clears the gate.
- **F2** [MINOR] [optional] §E.2 AC-001 row and the CHANGELOG entry claim "11/11 PASS"; the matrix's representative command showed only the Class A/B/C grep and never the per-locale heading-token grep — the en leg was unobserved, not passing (verification-claim surface 1 gap). Correct the two evidence lines when F1 lands.
- **F3** [MINOR] [optional] `docs-site/content/.moai/state/{config-cache,context-usage}.json` — runtime pollution inside content/ from a moai invocation with cwd drift; gitignored (`docs-site/.gitignore:7`), no commit risk; cosmetic.
- **F4** [MINOR] [optional] `public/*/sitemap.xml` carries only the 14 section URLs per locale — all 608 leaf pages absent. Pre-existing hugo.toml output policy untouched by this card; future-card material (SEO).

## Gaps (explicitly NOT observed)

- No cross-model backend (codex/GLM) consulted — the delegation made it optional and the deciding evidence is mechanical grep output, not model judgment.
- No CI verdict exists (branch unpushed); all evidence is local-tree.
- Craft spot-reads sampled ko fully (2 new pages + card-class section), en (card-class + factory summary + related-docs), ja/zh (one section each + intro paragraphs) — not a line-by-line read of every locale page.

## Residual risk

- F1's fix (one-word retitle) could be applied to en only, leaving a latent convention question (README title-case vs docs-site sentence-case) for future pages; harmless if resolved by the F1 decision itself.
- The ratchet baseline (54) still under-captures nothing new after the canonical recipe, but H2-only tooling would keep re-reporting the 4 pre-existing ko-vs-derived H2-only divergences; any hand-rolled future check should use the canonical regex.

## Re-audit scope (per finding-consumption discipline)

Re-run AC-001's four locale greps (heading tokens + Class A/B/C) after F1 lands; confirm the two F2 evidence-line corrections. No other dimension requires re-measurement — every other gate reproduced on this tree.
