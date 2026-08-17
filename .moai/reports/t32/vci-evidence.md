# t32 VCI Evidence — docs-site v3.1 residuals (B5/B6/B7/B8)

Card: t32 (bundled with t98 — t98 runs after this card integrates)
Branch: WT-t32 @ base ba679791c (origin/release/v3.1.1)
Date: 2026-08-18
Session: db221a6c-e73f-4806-b60e-bc00af9ab6fa (run lane)

## 1. Claim (주장)

1. **B5** — the 8 priority pages where ko (v3.1 KO-SSOT rewrite, commit b50c4de71) outstripped
   en/ja/zh were re-derived in all 3 derived locales: en, ja, zh (24 files, full re-derivation
   from ko canonical, not a merge).
2. **B6** — premise OVERTURNED (already resolved before this card): all 7 advanced/ pages
   (kanban-mode, manager-kanban, bas-navigator, multi-model-audit, autonomy-tier,
   harness-learning, mx-scanner-internals) ARE present in `docs-site/data/menu/main.yaml`
   (lines 675–716, 4-locale names) and in `docs-site/content/en/advanced/_meta.yaml`,
   with `new_items` badge list present (5 entries). Menu entry added by 2427eaf2d (#1545)
   and bd06ea127. Badge machinery confirmed live (`layouts/partials/menu.html:40-43`,
   `layouts/_default/single.html:27-34` read `_meta.yaml new_items`). No edit needed.
3. **B7** — premise OVERTURNED (already exists): `moai epic status` IS documented —
   `docs-site/content/{ko,en,ja,zh}/cli-reference/epic.md` (4,850/4,522/5,268/4,350 B),
   menu `main.yaml:420`, `_meta.yaml:44`. No edit needed.
4. **B8** — premise OVERTURNED (already exists): branch_guard IS documented —
   `## workflow.yaml — branch_guard` section at line 154 of
   `docs-site/content/{ko,en,ja,zh}/advanced/config-sections.md` (key table, default-off
   rationale, scope, exemptions, fail-open). No edit needed.
5. **Ratchet tightened**: the 8 derived pages converged (4-locale H2+ parity exact), so
   8 lines were pruned from `docs-site/.locale-parity-baseline` (63 → 55 page entries).
6. Verify recipe (hns-oss-docs-verify) passes on every must_pass dimension.

## 2. Evidence (증거) — command + verbatim output

### B5 content gap was real (before)

`grep -c '^#\{2,\}'` (gate metric, H2+H3+) — before values ko/en/ja/zh:

| page | ko | en | ja | zh |
|---|---|---|---|---|
| multi-llm/model-policy | 24 | 13 | 13 | 13 |
| multi-llm/cg-mode | 23 | 18 | 18 | 18 |
| cost-optimization/prompt-caching | 16 | 8 | 8 | 8 |
| getting-started/introduction | 27 | 24 | 23 | 23 |
| advanced/harness-profiles | 12 | 10 | 10 | 10 |
| advanced/no-haiku-3tier | 12 | 9 | 9 | 9 |
| advanced/analyze-first-routing | 12 | 10 | 10 | 10 |
| claude-code/agentic/best-practices | 31 | 24 | 24 | 24 |

Section deficit total: 123 across the 3 derived locales.

### B5 after derivation — orchestrator re-measurement (not translator self-report)

`grep -c '^#\{2,\}' <24 files>` verbatim results:

```
docs-site/content/en/multi-llm/model-policy.md:24    docs-site/content/ja/multi-llm/model-policy.md:24    docs-site/content/zh/multi-llm/model-policy.md:24
docs-site/content/en/multi-llm/cg-mode.md:23         docs-site/content/ja/multi-llm/cg-mode.md:23         docs-site/content/zh/multi-llm/cg-mode.md:23
docs-site/content/en/cost-optimization/prompt-caching.md:16   ja:16   zh:16
docs-site/content/en/getting-started/introduction.md:27       ja:27   zh:27
docs-site/content/en/advanced/harness-profiles.md:12          ja:12   zh:12
docs-site/content/en/advanced/no-haiku-3tier.md:12            ja:12   zh:12
docs-site/content/en/advanced/analyze-first-routing.md:12     ja:12   zh:12
docs-site/content/en/claude-code/agentic/best-practices.md:31 ja:31   zh:31
```

8/8 pages × 3 locales = exact ko match. Byte sizes en: 17,929/12,439/15,798/15,932/14,415/12,433/12,247/17,841; ja: up from 9,610/9,524/7,594/16,658/5,378/9,569/6,287/14,308; zh: 16,574/11,375/14,246/14,043/12,631/11,157/10,472/15,879 (per-translator reports; heading counts independently re-measured above).

### Verify recipe — all checks

1. build-clean: `hugo --source docs-site --minify --gc` → `Total in 2789 ms`; WARN/ERROR grep → exit 1 (zero matches); `test -f docs-site/public/sitemap.xml` → `sitemap OK`.
2. URL blacklist: `grep -rn 'docs\.moai-ai\.dev\|adk\.moai\.com\|adk\.moai\.kr' docs-site/content README.md README.ko.md README.ja.md README.zh.md` → exit 1 (zero matches).
3. Mermaid TD-only: `grep -rn 'flowchart LR\|graph LR\|flowchart RL\|graph RL' docs-site/content` → exit 1 (zero matches). All diagrams in the 24 files are `flowchart TD` (grep verified).
4. locale-parity ratchet: full-tree divergence set (recipe awk over grep -rc '^#\{2,\} ') → `55` lines == pruned baseline entries (55). Grep of divergence set for the 8 page slugs → only `claude-code/context-memory/prompt-caching.md` (a DIFFERENT, pre-existing baselined page; matched by the "prompt-caching" substring). Zero NEW divergence; 8 pages converged.
5. README 4-file parity: `grep -c '^## ' README.md README.ko.md README.ja.md README.zh.md` → 12/12/12/12.
6. Body emoji: `rg -c "[\x{1F300}-\x{1FAFF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}]" <24 files>` → `rg-exit: 1` (zero matches, not even allowed-class symbols). Card-added diff lines scan (`git diff | rg '^\+.*[…]' | rg -v -F '{{<'`) → zero matches.

### Changed-file census

`git status --porcelain -- docs-site/` → exactly 24 ` M` entries (8 pages × en/ja/zh) + 1 ` M docs-site/.locale-parity-baseline`. No additions/deletions elsewhere.

## 3. Baseline-attribution (baseline 귀속)

All measurements above were run in THIS session against THIS tree: worktree
`.claude/worktrees/t32`, branch WT-t32, HEAD ba679791c (= origin/release/v3.1.1 at card
start), after the 3 translator agents returned. Before-values were measured in the same
tree BEFORE spawning the translators (same command form). Hugo 0.x at /opt/homebrew/bin/hugo.
No carry-over from other cards or other trees.

## 4. Gaps (미검증)

- Recipe §4 file-existence loop (`for f in $(cd ko && find . …)`) NOT run — blocked by the
  worktree-isolation command guard (rejects for-loops). Mitigation: git status shows only
  ` M` entries (no file added/removed), so existence parity is unchanged from the
  pre-existing state by construction; hugo page counts 179/177/177/177 identical before
  and after (KO 179 vs others 177 is pre-existing).
- Whole-tree body-emoji scan shows 470 pre-existing emoji-bearing lines (allowed-class
  banner-example code blocks per recipe §5's code-block-context carve-out). Not
  individually adjudicated — pre-existing state, zero lines introduced by this card.
- Per-page content fidelity beyond structural parity (facts/figures/tables) rests on the
  three translator specialists' same-PR reports + orchestrator H2+/H3/mermaid/frontmatter
  spot-checks (model-policy drift/enforcement section read in full; cg-mode CLI verb
  cross-checked against internal/cli/glm.go:120 `Use: "setup [api-key]"` → `moai glm setup` confirmed). Full sentence-by-sentence audit of 24 files not performed.

## 5. Residual-risk (잔여 위험)

- **ko-canonical defects carried verbatim (flagged, not fixed — canonical is out of card scope):**
  1. `ko/getting-started/introduction.md:52` "profile matrix — 12 에이전트 × 3 프로필" vs same page :107/:127 "11개" and `ko/multi-llm/model-policy.md:21,:103` "11 × 3 = 33 셀" (catalog-12 vs matrix-11 conflation in the canonical).
  2. `ko/getting-started/introduction.md:247,:250` CG savings "50~70%" vs `ko/multi-llm/cg-mode.md:5,:13` "60-70%" (CLAUDE.md also says 60-70%).
  All 4 locales now carry these consistently → fixing ko later requires a 4-locale sync edit (same-PR obligation).
- zh introduction grew least (14,043 B vs ko 16,797) — Chinese typography is denser than Korean; heading parity is exact, prose depth parity is translator-judged.
- Vercel production deploy fires on push of the release branch (batch PR stage) — not on this card branch; no user-facing exposure until the lead merges.

## Execution note

3 translator agents (hns-oss-docs-locale-translator-specialist) run sequentially (write-capable
agent concurrency rule): en → ja → zh. Each loaded Skill("hns-oss-docs-i18n-rules") first,
derived from ko canonical only, and ran its own exit gates. Orchestrator independently
re-measured every heading-count claim (VCI §1.1 surface 2 discipline).
