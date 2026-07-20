# plan.md — SPEC-DOCSITE-ADVANCED-001

> Implementation plan for the docs-site v3.0 Advanced Guides content expansion.
> Milestones ordered by **decision-reversibility** (Rule 1 Approach-First): the
> decisions most likely to change lead; the mechanical / refactoring steps defer
> to the bottom.

## §A Context

### A.1 Work location & branch

- **Project root**: `/Users/goos/MoAI/moai-adk-go/`
- **Work subtree**: `docs-site/` (the `adk.mo.ai.kr` Hugo site)
- **Route**: Route A — Hybrid Trunk main-direct (docs-only SPEC, Tier L but zero Go code; per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline, Route A is the default; Route B is reserved for Tier L OR explicit `--pr`). **Recommendation**: this SPEC stays on Route A unless the user explicitly requests a PR. The 24 file writes are mechanically uniform and a feature branch adds friction without value for a 1-person OSS docs edit.
- **Current HEAD** (at plan-phase authoring): consult `git rev-parse HEAD` at run-phase M0 pre-flight.

### A.2 Files touched (29 total)

| Category | Path | Action | Count |
|----------|------|--------|-------|
| New content | `docs-site/content/{ko,en,ja,zh}/advanced/{tokenomics-overview,token-budget,no-haiku-3tier,plan-type-profiles,self-evolving,autonomous-loops}.md` | new file | 24 |
| Parity fix + new entries | `docs-site/content/{ko,en,ja,zh}/advanced/_meta.yaml` | edit (add entries) | 4 |
| Menu registration | `docs-site/data/menu/main.yaml` | edit (add 6 sub-entries under Advanced) | 1 |
| **Total** | | | **29** |

**NOT touched** (PRESERVE — see §D): `docs-site/hugo.toml`, `docs-site/vercel.json`, `docs-site/layouts/**`, `docs-site/static/**`, `docs-site/data/menu/main.yaml` outside the Advanced sub-block, all other locales' non-advanced `_meta.yaml`.

### A.3 Plan-auditor verdict (to be filled at Phase 1)

```
plan-auditor iter-1 score: <pending>
verdict: <pending>
skip-eligible (≥0.90 + hash-unchanged + <24h): <pending>
```

### A.4 Implementation Kickoff Approval

Per CLAUDE.local.md §19.1 + orchestration-mode-selection.md header: this plan→run HUMAN GATE is **mandatory and score-independent**. Even if plan-auditor returns PASS ≥ 0.90, the orchestrator MUST obtain explicit user approval via `AskUserQuestion` before M1 commit. The skip-eligible policy governs ONLY Phase 1 verdict re-execution, NOT this gate.

---

## §B Known Issues (auto-injection per manager-develop-prompt-template §B1-B12)

Filtered to categories relevant to this docs-only SPEC:

| ID | Category | Applicability | Note |
|----|----------|---------------|------|
| B1 | Cross-platform build tags | N/A | zero Go code |
| B2 | Cross-SPEC policy conflict | LOW | confirm no retired-spec tokens; the v3.0 narrative references EVOLVE-* and MODEL-TIER-PLANTYPE-* which are CLOSED, not retired |
| B3 | C-HRA-008 subagent boundary | N/A | no `internal/harness/` or `internal/hook/` writes |
| B4 | Frontmatter canonical schema | N/A | this SPEC's own frontmatter is validated at plan-phase; the docs-site content files use Hugo frontmatter (a `title:` field only), not SPEC frontmatter |
| B5 | CI 3-tier awareness | MEDIUM | spec-lint runs on this SPEC's own artifacts; hugo build is run locally (no CI gate); the CI guard `internal/template/internal_content_leak_test.go` does NOT bind docs-site (it binds `internal/template/templates/`) |
| B6 | spec-lint heading convention | HIGH | this SPEC's own `### Out of Scope — <topic>` H3 sub-headings must satisfy `OutOfScopeRule` (verified: spec.md §E has 7 such sub-headings) |
| B7 | observer.go capture path | N/A | zero Go code |
| B8 | Working tree hygiene | HIGH | the git status snapshot at task start shows many `M .moai/harness/proposals/*` and `M .moai/config/sections/*` modified files — these are PRE-EXISTING uncommitted changes from prior sessions. **MUST use specific-path `git add` (never `git add -A`)** to avoid absorbing unrelated untracked state into this SPEC's commits. |
| B9 | Git commit + push direct (Hybrid Trunk) | APPLICABLE | direct-to-main per CLAUDE.local.md §23. Per-M commits with Conventional Commits format. Never `--no-verify` (a warn-only pre-commit hook is normal). |
| B10 | Untouched paths PRESERVE | HIGH | see §D PRESERVE list — many sibling SPEC dirs exist under `.moai/specs/`; do NOT touch unrelated SPEC dirs |
| B11 | AskUserQuestion prohibited (subagent) | APPLICABLE | manager-develop returns blocker reports, never asks the user. The orchestrator handles any user-decision routing. |
| B12 | CHANGELOG discipline | N/A | manager-docs owns sync-phase CHANGELOG, not this plan-phase |

### B.13 Memory carry-over discipline (lesson applied)

Per `feedback_defect_claim_verification` (cited in the spawn prompt): any claimed "source ready" MUST be verified, not accepted. Plan-phase recon verified all 6 sources (research.md §B). Run-phase manager-develop MUST re-verify by directly reading the cited source files BEFORE drafting each page's content. If a cited source is missing or insufficient at run-phase, return a blocker report — do NOT fabricate.

### B.14 Parallel-session race mitigation

Per `feedback_shared_checkout_concurrent_commit_race` + `agent-common-protocol.md` § Pre-Spawn Sync Check: before the first M1 commit, manager-develop runs `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` AND `moai session list --json --filter-spec=SPEC-DOCSITE-ADVANCED-001`. On divergence or active-session entry, return a blocker report.

### B.15 _meta.yaml merge hazard

`docs-site/content/<locale>/advanced/_meta.yaml` is a per-locale file shared across all advanced-page SPECs. If a parallel session edits the same file (e.g., adds a different page), `git add` and `git commit` will race. Mitigation: per-locale `_meta.yaml` edits land in dedicated commits (one commit per locale for the parity fix, one commit per locale for the new entries), and the pre-spawn sync check (B.14) MUST be re-run before each `_meta.yaml` write.

---

## §C Pre-flight (manager-develop executes before M1)

```bash
# 1. Branch + HEAD baseline
git branch --show-current
git rev-parse HEAD

# 2. Confirm Route A (no feature branch creation unless explicit --pr)
git status --short | head -20    # inspect pre-existing modified files

# 3. Pre-spawn sync check (B.14)
git fetch origin main
git rev-list --count --left-right origin/main...HEAD
moai session list --json --filter-spec=SPEC-DOCSITE-ADVANCED-001 2>/dev/null || echo "no registry hook"

# 4. Verify the parity debt baseline (re-measure; do NOT trust plan-phase numbers)
for loc in ko en ja zh; do
  echo "=== $loc/advanced/_meta.yaml entry count ==="
  grep -cE '^[a-zA-Z_-]+:' docs-site/content/$loc/advanced/_meta.yaml
done
# Expected at M0 start: ko=11 (the index block + 10 entry "name:" keys, OR count "name:" instead), en=8, ja=8, zh=8
# (use the exact counting rule from acceptance.md CMD-A)

# 5. Verify all 14 existing content files are present in every locale
for loc in ko en ja zh; do
  echo "=== $loc advanced content files ==="
  ls docs-site/content/$loc/advanced/*.md | wc -l
done
# Expected: 15 each (14 pages + _index.md)

# 6. Confirm docs-site builds clean before any change
cd docs-site && hugo --minify --gc 2>&1 | tail -5; cd ..
# Capture exit code; if non-zero, report blocker (pre-existing build break)

# 7. Confirm the canonical-ko authoring sources exist (B.13 lesson)
for f in \
  .moai/reports/readme-docs-redesign-20260713.md \
  .moai/reports/model-tier-redesign-20260712.md \
  .moai/reports/harness-self-evolving-redesign-final-20260712.html \
  .moai/reports/agent-architecture-redesign-v2-20260709.html \
  .moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/spec.md \
  .moai/specs/SPEC-TOKEN-BUDGET-STOP-001/spec.md \
  .moai/specs/SPEC-GOAL-ENGINE-001/spec.md \
  .claude/rules/moai/workflow/goal-directive.md \
  .claude/rules/moai/workflow/context-window-management.md \
  .claude/rules/moai/workflow/session-handoff.md; do
  [ -f "$f" ] && echo "  ✓ $f" || echo "  ✗ MISSING $f — BLOCKER"
done
```

---

## §D Constraints (DO NOT VIOLATE)

### D.1 PRESERVE list (scope discipline — Rule 5)

- `docs-site/hugo.toml` (version SSOT — never edited by content work)
- `docs-site/vercel.json` (Vercel binding — immutable per i18n rule §9)
- `docs-site/layouts/**` (render hooks, shortcodes, partials — design system, frozen)
- `docs-site/static/moai-brand.css` (FROZEN) + `static/moai-design.css`
- `docs-site/data/menu/main.yaml` — ONLY the Advanced sub-block (`ref: /advanced` parent's `sub:` array) is edited; all other top-level sections (Getting Started, Core Concepts, Claude Code, Workflow Commands, Utility Commands, Multi-LLM, Guides, Worktree, Contributing) are PRESERVED
- `docs-site/content/<locale>/_meta.yaml` for non-advanced sections
- All sibling SPEC directories under `.moai/specs/` (especially SPEC-DESKTOP-NATIVE-E2E-001 which the SubagentStart hook referenced — leave alone)
- All pre-existing modified files in the working tree (`.moai/harness/proposals/*`, `.moai/config/sections/*` — they predate this SPEC)
- `internal/template/templates/**` (template-tree, not bound by this SPEC)
- `.claude/rules/**`, `.claude/agents/**`, `.claude/skills/**` (doctrine-tree, not bound)

### D.2 Forbidden commands

- `git add -A` (B.8 hazard — would absorb the pre-existing modified files)
- `git add docs-site/` (too broad — would absorb unrelated docs-site drift if any)
- `git commit --no-verify` (a warn-only pre-commit hook is normal and expected)
- `--amend` on pushed commits (Hybrid Trunk §23 prohibition)
- `git push --force` to main

### D.3 Required commands

- Conventional Commits subject: `feat(SPEC-DOCSITE-ADVANCED-001): M<N> <subject>`
- Commit message trailer: `🗿 MoAI`
- Specific-path `git add docs-site/content/<locale>/advanced/<slug>.md` (per-file)
- Specific-path `git add docs-site/content/<locale>/advanced/_meta.yaml`
- Specific-path `git add docs-site/data/menu/main.yaml`

### D.4 Binary constraints (grep must return 0 / specific count)

- `grep -rE 'docs\.moai-ai\.dev|adk\.moai\.com|adk\.moai\.kr' docs-site/content/{ko,en,ja,zh}/advanced/{tokenomics-overview,token-budget,no-haiku-3tier,plan-type-profiles,self-evolving,autonomous-loops}.md` → 0 matches (URL blacklist)
- `grep -nE 'flowchart (LR|RL)|graph (LR|RL)' docs-site/content/{ko,en,ja,zh}/advanced/{6 slugs}.md` → 0 matches (Mermaid TD-only)
- `grep -P '[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}]' docs-site/content/{ko,en,ja,zh}/advanced/{6 slugs}.md` → 0 matches outside fenced code blocks (no body emoji — typography arrows `→ ← ↓ ✓ ✗` and banner-reproduction emoji inside ```` ``` ```` blocks are permitted; see acceptance.md CMD-G)

---

## §E Self-Verification Deliverables (E1-E7 per template)

Each reported per the verification-claim-integrity 5-section format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). Run-phase manager-develop reports these on completion:

- **E1**: AC binary PASS/FAIL matrix (33 AC — see acceptance.md §B)
- **E2**: Cross-platform build N/A (zero Go code) — but `cd docs-site && hugo --minify --gc` MUST exit 0 warning-free
- **E3**: Coverage N/A (docs-only SPEC — no Go coverage target)
- **E4**: Subagent boundary N/A (no `internal/harness/` or `internal/hook/` writes; no AskUserQuestion in docs-site content)
- **E5**: spec-lint on this SPEC's artifacts (`moai spec lint .moai/specs/SPEC-DOCSITE-ADVANCED-001/`) → 0 errors
- **E6**: Branch HEAD + push state (Route A: per-M commits pushed to main; Conventional Commit subjects + `🗿 MoAI` trailer)
- **E7**: Blocker reports if any source is missing or insufficient at run-phase

---

## §F Milestones (decision-reversibility ordered)

### §F.0 Tier decision (already resolved at plan-phase — recorded for audit)

**Decision: Tier L (single SPEC), NOT an Epic of 2-3 SPECs.**

Reversibility ranking (highest-change-likelihood first):

1. **Pillar narrative spine** (3-pillar structure, page-to-pillar mapping) — HIGHEST change-likelihood; an Epic would split the spine across SPEC boundaries and risk incoherence. A single SPEC keeps the spine in one design.md.
2. **Per-page content outlines** (canonical-KO structure per page) — medium-high; changes here ripple to all 4 locales. Centralized in design.md §C.
3. **_meta.yaml parity fix ordering** (M1 pre-fix vs M5 alongside) — medium; chose M1 pre-fix to land 6 new entries on a clean baseline. Reversible until M1 commit lands.
4. **Menu placement of 6 new entries** (prepend as v3.0 pillar cluster vs interleave by topic) — medium; chose prepend in pillar order (design.md §C). Reversible until main.yaml commit lands.
5. **Canonical-KO titles for the 6 pages** (the ko strings in main.yaml + _meta.yaml) — low-medium; locked at design.md §C table.
6. **24 file writes** (mechanical) — LOWEST change-likelihood; per-file Edit/Write once design is locked.

### §F.1 M1 — Parity debt pre-fix (4 commits, one per locale)

**Goal**: Bring all 4 locales' `advanced/_meta.yaml` to a clean 14-entry baseline before adding any new entries.

**Tasks per locale**:
- KO: add `decision-memory`, `harness-v4-builder`, `ultracode-workflows` (3 entries) with the locale-specific titles already used in `main.yaml` lines 367-383 (decision-memory), 373-377 (harness-v4-builder), 367-371 (ultracode-workflows).
- EN/JA/ZH: add `catalog-system`, `decision-memory`, `harness-profiles`, `harness-v4-builder`, `hooks-reference`, `ultracode-workflows` (6 entries each) with locale-specific titles matching `main.yaml`.

**Commit**: `feat(SPEC-DOCSITE-ADVANCED-001): M1 advanced/_meta.yaml parity debt fix (<locale>)` × 4 (or one combined commit `feat(SPEC-DOCSITE-ADVANCED-001): M1 advanced/_meta.yaml parity debt fix (4-locale)`).

**Verify**: `for loc in ko en ja zh; do grep -cE '^[a-zA-Z_-]+:' docs-site/content/$loc/advanced/_meta.yaml; done` → all 4 return the same count (15 = index block + 14 entries, or whatever counting rule acceptance.md CMD-A pins).

### §F.2 M2 — Pillar 1 (Tokenomics) authoring + derivation (8 files)

**Pages**: `tokenomics-overview`, `token-budget`.

**Order**: canonical KO first (`ko/advanced/tokenomics-overview.md`, `ko/advanced/token-budget.md`), then en derivation, then ja/zh derivation.

**Sources**:
- `tokenomics-overview`: `.moai/reports/readme-docs-redesign-20260713.md` (3-pillar narrative) + `project_v3_tokenomics_docs_plan` memory (4-axis material) + Token-Economy Epic A-D + `reference_tokenomics_press_2026_07` (market context hook).
- `token-budget`: `.moai/specs/SPEC-TOKEN-BUDGET-STOP-001/` + `.claude/rules/moai/workflow/{context-window-management,session-handoff}.md` + `agent-common-protocol.md` § File-redirect contract.

**Content structure**: see design.md §C.1 (tokenomics-overview) and §C.2 (token-budget).

**Commit**: per-page `feat(SPEC-DOCSITE-ADVANCED-001): M2 <slug> 4-locale derivation` × 2, or one combined.

### §F.3 M3 — Pillar 3 (Harness Architecture) authoring + derivation (12 files)

**Pages**: `no-haiku-3tier`, `plan-type-profiles`, `self-evolving`.

**Order**: canonical KO first per page, then en/ja/zh derivation.

**Sources**:
- `no-haiku-3tier`: `.moai/reports/agent-architecture-redesign-v2-20260709.html` + `project_model_tier_plantype_001_completed` (ApplyTierProfile live behavior).
- `plan-type-profiles`: `.moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/spec.md` (CLOSED, authoritative) + `.moai/reports/model-tier-redesign-20260712.md`.
- `self-evolving`: `.moai/reports/harness-self-evolving-redesign-final-20260712.html` (v5.1 FINAL SSOT) + SPEC-HARNESS-EVOLVE-{001,002,003} + `reference_lilian_weng_harness_selfimprove`.

**Honesty caveats** (REQ-DA-060 / 061 / 063 — content-bearing):
- `plan-type-profiles.md`: GLM overlay wire-effectiveness UNVERIFIED — page MUST say "구현 + 배선 완료, wire 유효성 실증 예정" / "implemented + wired, wire validity pending live verification".
- `no-haiku-3tier.md`: distinguish design report (intent) from ApplyTierProfile (live).
- `self-evolving.md`: disclose EVOLVE-004/005 as roadmap (Loop 2 production wiring is partial).

**Commit**: per-page `feat(SPEC-DOCSITE-ADVANCED-001): M3 <slug> 4-locale derivation` × 3.

### §F.4 M4 — Pillar 2 (Agentic Loop) authoring + derivation (4 files)

**Page**: `autonomous-loops`.

**Sources**: `.claude/rules/moai/workflow/goal-directive.md` (canonical /goal reference) + `.moai/specs/SPEC-GOAL-ENGINE-001/` + `project_agentic_core_epic_progress` (roadmap) + `project_goal_engine_cli_gap_handoff`.

**Honesty caveat** (REQ-DA-062): distinguish native `/goal` (HUMAN-ONLY TUI command) from `/moai goal` (programmatic, MoAI-owned, Axis B) from `/moai loop` (Ralph Engine diagnostic preset).

**Commit**: `feat(SPEC-DOCSITE-ADVANCED-001): M4 autonomous-loops 4-locale derivation`.

### §F.5 M5 — Navigation registration (main.yaml + per-locale _meta.yaml new entries)

**main.yaml edit**: under the existing Advanced section (`ref: /advanced`, `icon: school`, `defaultOpen: true`), PREPEND 6 new sub-entries in pillar order BEFORE the existing 14 entries (design.md §C.6 specifies the rationale: the 6 new pages form a v3.0 pillar cluster, conceptually distinct from the 14 component-reference pages).

**Per-locale _meta.yaml edit**: add 6 new entries per locale (KO/en/ja/zh) with locale-specific titles from design.md §C.5 title table.

**Commit**: `feat(SPEC-DOCSITE-ADVANCED-001): M5 menu + _meta registration (6 new Advanced entries, 4-locale)`.

### §F.6 M6 — Verify (hugo build + parity grep + design-regime grep)

**Verify recipe**: invoke `Skill("hns-oss-docs-verify")` at run-phase start — it carries the canonical exit-gate recipe. The manager-develop MUST run all checks below and report each in the §E.1 AC matrix:

1. **Hugo build warning-free**: `cd docs-site && hugo --minify --gc` exits 0 with 0 warnings.
2. **Sitemap generation**: `docs-site/public/sitemap.xml` contains the 24 new paths.
3. **4-locale file-existence parity**: 24 new files exist (CMD-A in acceptance.md).
4. **_meta.yaml 4-locale parity**: each locale `_meta.yaml` has 20 entries (14 existing + 6 new) — CMD-B.
5. **main.yaml registration**: Advanced section has 20 sub-entries (14 existing + 6 new) with 4-locale name fields — CMD-C.
6. **URL blacklist grep**: 0 matches for `docs.moai-ai.dev|adk.moai.com|adk.moai.kr` in the 24 new files — CMD-D.
7. **Mermaid TD-only grep**: 0 matches for `flowchart (LR|RL)|graph (LR|RL)` — CMD-E.
8. **Body-emoji grep**: 0 matches outside fenced code blocks — CMD-F.
9. **Icon-shortcode usage**: at least N icon-shortcode invocations across the 24 files (positive signal, not just absence of emoji) — CMD-G.
10. **GLM honesty caveat grep**: `plan-type-profiles.md` 4 locales contain the wire-validity-pending caveat — CMD-H.
11. **Native vs programmatic /goal distinction**: `autonomous-loops.md` 4 locales contain both `/goal` and `/moai goal` with the HUMAN-ONLY distinction — CMD-I.

**Commit**: `feat(SPEC-DOCSITE-ADVANCED-001): M6 verify (hugo build clean + 4-locale parity + design regime grep)` — note this commit is verification-only; if any check fails, return blocker report and do NOT commit until fixed.

---

## §G Anti-Patterns

- **AP-DA-001 — Fabricating content from memory carry-over**: drafting a page from recalled narrative without reading the cited source file at run-phase. Mitigation: B.13 lesson — re-verify each source by direct Read before drafting.
- **AP-DA-002 — `git add -A` kitchen-sink**: absorbing pre-existing `.moai/harness/proposals/*` and `.moai/config/sections/*` modifications into this SPEC's commits. Mitigation: specific-path `git add` only (B.8).
- **AP-DA-003 — Translating code blocks**: locale-translator inadvertently translates commands, file paths, or identifiers. Mitigation: REQ-DA-021 + the oss-docs-i18n canonical chain — code blocks are locale-invariant.
- **AP-DA-004 — Adding 6 new entries on top of broken _meta parity**: skipping M1 and going straight to M5 would stack 6 new entries on a 14-vs-11-vs-8 inconsistent baseline. Mitigation: M1 is a hard precondition; the milestones are NOT freely reorderable.
- **AP-DA-005 — Putting body emoji in decorative prose**: writing `📖 핵심 개념` instead of `{{</* icon book */>}} 핵심 개념`. Mitigation: REQ-DA-050 + CMD-F grep.
- **AP-DA-006 — Describing GLM overlay as "works"**: claiming wire-effectiveness for the GLM effort overlay without live outbound observation. Mitigation: REQ-DA-060 + CMD-H grep — the page MUST carry the "wire validity pending" caveat.
- **AP-DA-007 — Conflating native `/goal` with `/moai goal`**: treating them as aliases when they are HUMAN-ONLY vs PROGRAMMATIC respectively. Mitigation: REQ-DA-062 + CMD-I grep.
- **AP-DA-008 — Editing main.yaml outside the Advanced sub-block**: drifting menu structure for unrelated sections. Mitigation: §D.1 PRESERVE list pins main.yaml edits to the Advanced sub-block only.
- **AP-DA-009 — Modifying the FROZEN moai-brand.css**: attempting to add a new design token. Mitigation: §D.1 + §C.1 constraints.
- **AP-DA-010 — Skipping the pre-spawn sync check**: a parallel session edits the same per-locale `_meta.yaml`. Mitigation: B.14 + B.15 — re-run before each `_meta.yaml` write.

---

## §H Cross-References

- **spec.md**: `.moai/specs/SPEC-DOCSITE-ADVANCED-001/spec.md` (this SPEC's requirements — WHAT/WHY)
- **acceptance.md**: `.moai/specs/SPEC-DOCSITE-ADVANCED-001/acceptance.md` (33 AC with Given-When-Then + CMD-A through CMD-I verbatim)
- **design.md**: `.moai/specs/SPEC-DOCSITE-ADVANCED-001/design.md` (page architecture + 3-pillar spine + canonical-KO title table + per-page outline)
- **research.md**: `.moai/specs/SPEC-DOCSITE-ADVANCED-001/research.md` (parity-debt findings + per-page source readiness evidence + memory carry-over risk)
- **progress.md**: `.moai/specs/SPEC-DOCSITE-ADVANCED-001/progress.md` (§E skeleton — run/sync evidence goes here)
- **Prior art**: `.moai/specs/SPEC-DOCSITE-E2E-001/` (the previous 4-locale docs-site SPEC — same canonical-ko→en→ja/zh discipline, same parity-debt pattern, closed 2026-07-13)
- **Run-phase delegation**: manager-develop with the full Section A-E template (Tier L applicability per `manager-develop-prompt-template.md`); spawn prompt includes "At start, invoke Skill('hns-oss-docs-i18n-rules'), Skill('hns-oss-docs-structure-map'), and Skill('hns-oss-docs-verify') for the canonical i18n rules, structure schema, and exit-gate recipe" (per `skill-routing.md` §1).
- **Implementation Kickoff Approval**: CLAUDE.local.md §19.1 + orchestration-mode-selection.md header (mandatory, score-independent).

---

Version: 0.1.0 | Tier: L | Status: draft (plan-phase)
