# research.md — SPEC-V3R6-DOCS-COVERAGE-001

> Plan-phase research artifact. Primary-source evidence + 4-locale drift inventory + facts-bearing page coverage map. All counts verified via verbatim command output on 2026-06-18 against HEAD a7c1b4d48.

---

## §1. Primary-Source Skill Count (Canonical)

### §1.1 Template-shipped skills (canonical 32)

The canonical skill count is established from the **template source** (`internal/template/templates/.claude/skills/`), which is the single source of truth for what ships to end users.

**Command + verbatim output:**

```bash
$ find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d | wc -l
32

$ find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -exec basename {} \; | sort
moai                       ← umbrella router
moai-design-system
moai-domain-backend
moai-domain-brand-design
moai-domain-copywriting
moai-domain-database
moai-domain-design-handoff
moai-domain-frontend
moai-domain-humanize       ← the missing skill in docs
moai-domain-ideation
moai-domain-research
moai-foundation-cc
moai-foundation-core
moai-foundation-quality
moai-foundation-thinking
moai-harness-learner
moai-meta-harness
moai-ref-api-patterns
moai-ref-git-workflow
moai-ref-owasp-checklist
moai-ref-react-patterns
moai-ref-testing-pramid
moai-workflow-ci-loop
moai-workflow-ddd
moai-workflow-design
moai-workflow-gan-loop
moai-workflow-loop
moai-workflow-project
moai-workflow-spec
moai-workflow-tdd
moai-workflow-testing
moai-workflow-worktree
```

**Total: 32 entries** = 1 `moai` umbrella router + 31 specialized skills.

### §1.2 Per-category breakdown (template source)

```bash
$ echo "moai-foundation-*: $(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-foundation-*' | wc -l | tr -d ' ')"
moai-foundation-*: 4

$ echo "moai-workflow-*:   $(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-workflow-*' | wc -l | tr -d ' ')"
moai-workflow-*:   10

$ echo "moai-domain-*:     $(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-domain-*' | wc -l | tr -d ' ')"
moai-domain-*:     9

$ echo "moai-ref-*:        $(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-ref-*' | wc -l | tr -d ' ')"
moai-ref-*:        5

$ echo "moai-harness-*:    $(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-harness-*' | wc -l | tr -d ' ')"
moai-harness-*:    1

$ echo "moai-meta-*:       $(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-meta-*' | wc -l | tr -d ' ')"
moai-meta-*:       1

$ echo "moai-design-*:     $(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-design-*' | wc -l | tr -d ' ')"
moai-design-*:     1

$ echo "moai (umbrella):   $(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai' | wc -l | tr -d ' ')"
moai (umbrella):   1
```

**Category summary:**

| Category | Prefix(es) | Count | Docs claim | Drift |
|----------|-----------|-------|-----------|-------|
| Foundation | `moai-foundation-*` | 4 | 4 | OK |
| Workflow | `moai-workflow-*` | 10 | 10 | OK |
| Domain | `moai-domain-*` | **9** | **8** | **−1 (missing humanize)** |
| Reference | `moai-ref-*` | 5 | 5 | OK |
| Meta/Harness | `moai-meta-*` + `moai-harness-*` | 2 | 2 | OK |
| Design | `moai-design-*` | 1 | 1 | OK |
| Umbrella router | `moai` | 1 | 1 | OK |
| **Specialized total** | | **31** | **30** | **−1** |
| **Grand total** | | **32** | **31** | **−1** |

### §1.3 Local `.claude/skills/` cross-check (includes user-owned)

```bash
$ find .claude/skills -maxdepth 1 -mindepth 1 -type d | wc -l
34
```

The local count (34) = 32 template-shipped + 2 user-owned (`my-harness-moaiadk-best-practices`, `my-harness-moaiadk-patterns`). The user-owned harness skills are out of template scope per `.claude/CLAUDE.local.md` §24 Harness Namespace 분리 정책 and MUST NOT be included in the docs-site "built-in skills" count.

### §1.4 The missing skill: `moai-domain-humanize`

Provenance: per `.claude/NOTICE.md`, the `moai-domain-humanize` skill (Korean AI-tell taxonomy, im-not-ai import) was added **2026-06-15**. The docs-site was never updated to include it.

**Verification that `humanize` appears in ZERO docs pages:**

```bash
$ for loc in en ko ja zh; do
    echo "--- $loc humanize mentions ---"
    grep -rl 'humanize' "docs-site/content/$loc/" 2>/dev/null | head -5
  done
--- en humanize mentions ---
--- ko humanize mentions ---
--- ja humanize mentions ---
--- zh humanize mentions ---
```

All 4 locales: 0 matches. The skill exists in the codebase but is invisible in the docs.

### §1.5 en docs Domain listing — missing entry identified

```bash
$ comm -23 \
    <(find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-domain-*' -exec basename {} \; | sort) \
    <(sed -n '91,103p' docs-site/content/en/advanced/skill-guide.md | grep -oE 'moai-domain-[a-z-]+' | sort)
moai-domain-humanize
```

The en docs Domain section lists 8 of the 9 actual Domain skills. `moai-domain-humanize` is the sole missing entry.

---

## §2. 4-Locale Drift Inventory

### §2.1 Methodology

Per applied lesson `feedback_digit_boundary_locale_grep_4parity`, each locale was scanned independently with:

1. **Digit-boundary anchored grep**: `grep -rnE '(^|[^0-9])31([^0-9]|$)'` — the `(^|[^0-9])` prefix prevents "31" matching inside "231" or "314" (false-positive); the `([^0-9]|$)` suffix prevents matching inside "315".
2. **Per-locale loop** (`for loc in en ko ja zh`) — NEVER globbed the total (`grep -r docs-site/content/`), because globbing masks per-locale drift.
3. **Locale-native idiom filter** — each locale's skill-adjacent keyword was matched in its native script (en: "skill", ko: "스킬", ja: "スキル", zh: "技能"), never invented CJK.

### §2.2 en locale (10 count-claim occurrences + 10 in-body fictional-name residuals)

**Count-claim inventory (10 occurrences across 4 pages):**

| File | Line | Excerpt |
|------|------|---------|
| `advanced/builder-agents.md` | 15 | "the 31 built-in skills and 8 agents" |
| `advanced/skill-guide.md` | 65 | "a total of **31 skills** — the `moai` umbrella router plus 30 specialized skills" |
| `advanced/skill-guide.md` | 127 | "counted in the total of 31 but is not a categorized capability skill" |
| `advanced/skill-guide.md` | 167 | "Load all 31 skills = ~155,000 tokens" |
| `getting-started/introduction.md` | 133 | "**8** specialized AI agents + **31** skills" |
| `getting-started/introduction.md` | 156 | "**31 Skills**: An extensible skill library" |
| `getting-started/introduction.md` | 163 | "8 specialized AI agents and 31 skills" |
| `getting-started/update.md` | 396 | `"🗿 MoAI-ADK: 8 Specialized Agents + 31 Skills with SPEC-First DDD"` |
| `core-concepts/what-is-moai-adk.md` | 7 | "8 specialized AI agents and 31 skills collaborate" |
| `core-concepts/what-is-moai-adk.md` | 48 | "**8** specialized AI agents + **31** skills" |
| `core-concepts/what-is-moai-adk.md` | 267 | "### 31 Skills (Progressive Disclosure)" |
| `core-concepts/what-is-moai-adk.md` | 652 | "skills/moai-*/ # 31 skill modules" |

**In-body fictional-name residuals (10 matches — iter-2 additional finding):**

The en `advanced/skill-guide.md` category STRUCTURE is correct (6 canonical headers), but the page body carries 10 references to nonexistent skill names (`moai-lang-*`, `moai-library-mermaid`, `moai-platform-supabase`) inside conceptual illustrations: Mermaid flowchart nodes (L185, L333, L335), explicit-invocation code examples (L218, L220), ASCII directory-tree illustrations (L244), skill-file-structure frontmatter example (L266), auto-load scenario comments (L309, L314), and a closing callout (L369). These are NOT in the category-listing tables (the en Domain/Foundation/Workflow tables list only real skills), but they still present fictional names as if they were shipped skills — a drift class covered by REQ-009 / AC-011.

```bash
$ grep -cE 'moai-lang-|moai-platform-|moai-library-|moai-framework-|moai-foundation-claude|moai-foundation-philosopher|moai-foundation-context' \
    docs-site/content/en/advanced/skill-guide.md
10
```

### §2.3 ko locale — STRUCTURAL DIVERGENCE (identical magnitude to ja)

> **iter-2 correction (2026-06-18):** the iter-1 inventory below listed only 10 count-claim occurrences and treated ko as a count-patch target. Independent re-derivation via the fictional-name regex (the same regex applied to ja in §2.4) revealed ko `advanced/skill-guide.md` carries the **identical pre-v3 fictional taxonomy** as ja: 9 fictional categories, 37 nonexistent skill-name references, and 3 missing canonical categories (Reference, Meta/Harness, Design). The count-claim inventory is retained for completeness, but the load-bearing ko defect is structural, not numeric.

**Mechanical re-derivation evidence:**

```bash
$ grep -cE 'moai-lang-|moai-platform-|moai-library-|moai-framework-|moai-foundation-claude|moai-foundation-philosopher|moai-foundation-context' \
    docs-site/content/ko/advanced/skill-guide.md
37

$ grep -nE '^###' docs-site/content/ko/advanced/skill-guide.md | head -10
64:### Foundation (핵심 철학) - 5개        ← wrong sub-count (actual 4) + lists nonexistent foundation-claude/philosopher/context
74:### Workflow (자동화 워크플로우) - 11개  ← wrong sub-count (actual 10)
90:### Domain (도메인 전문성) - 4개         ← wrong sub-count (actual 9) + missing humanize
99:### Language (프로그래밍 언어) - 16개    ← FICTIONAL category (canonical delivery is rules/moai/languages/, not skills)
120:### Platform (클라우드/BaaS) - 4개      ← FICTIONAL category
128:### Library (특수 라이브러리) - 4개     ← FICTIONAL category
137:### Tool (개발 도구) - 2개              ← FICTIONAL category
144:### Framework (앱 프레임워크) - 1개     ← FICTIONAL category
150:### Design Tools (디자인 도구) - 1개    ← FICTIONAL label (canonical is "Design", not "Design Tools")
184:### 각 레벨의 역할                       (downstream — not a category header)

$ grep -nE '^### (Reference|Meta/Harness|Design )' docs-site/content/ko/advanced/skill-guide.md
150:### Design Tools (디자인 도구) - 1개
# → Reference ABSENT, Meta/Harness ABSENT, Design present only as mis-labeled "Design Tools"
```

**ko fictional taxonomy (mirrors ja §2.4):** 9 categories (Foundation/Workflow/Domain/Language/Platform/Library/Tool/Framework/Design Tools), 37 nonexistent skill-name references (`moai-lang-*` 16, `moai-platform-*` 4, `moai-library-*` 4, `moai-framework-*` 1, `moai-foundation-claude/philosopher/context` 3, and others), and 3 missing canonical categories (Reference, Meta/Harness, and the correctly-labeled Design). The Domain section lists `moai-domain-uiux` (also nonexistent — the actual UI/design skills are `moai-domain-brand-design`, `moai-domain-design-handoff`, `moai-domain-copywriting`, `moai-domain-humanize`) and omits 5 real Domain skills including `moai-domain-humanize`.

**ko count-claim inventory (retained for completeness):**

| File | Line | Excerpt |
|------|------|---------|
| `advanced/builder-agents.md` | 16 | "기본 제공되는 31개 스킬, 8개 에이전트" |
| `advanced/skill-guide.md` | 62 | "총 **31개 스킬**이 6개 카테고리로 분류" |
| `advanced/skill-guide.md` | 194 | "31개 스킬 전체 로드 = 약 150,000 토큰" |
| `getting-started/introduction.md` | 133 | "**8개** 전문 AI 에이전트 + **31개** 스킬" |
| `getting-started/introduction.md` | 156 | "**31개 스킬**: 다양한 개발 시나리오" |
| `getting-started/introduction.md` | 163 | "8개의 전문화된 AI 에이전트와 31개의 스킬" |
| `core-concepts/what-is-moai-adk.md` | 7 | "8개의 전문 AI 에이전트와 31개의 스킬이 협력" |
| `core-concepts/what-is-moai-adk.md` | 48 | "**8개** 전문 AI 에이전트 + **31개** 스킬" |
| `core-concepts/what-is-moai-adk.md` | 267 | "### 31개 스킬 (Progressive Disclosure)" |
| `core-concepts/what-is-moai-adk.md` | 652 | "skills/moai-*/ # 31개 스킬 모듈" |

Note: ko `getting-started/update.md` exists (see §3.1) but does NOT carry the `31 <skill-count>` statusline string — only en/zh `update.md` do. ko `update.md` requires no count correction in this page-family.

### §2.4 ja locale (7 count-claim occurrences + 37 in-body fictional-name residuals) — STRUCTURAL DIVERGENCE

| File | Line | Excerpt |
|------|------|---------|
| `advanced/builder-agents.md` | 17 | "基本提供される 31 スキル、8 エージェント" |
| `advanced/skill-guide.md` | 59 | "計**31スキル**が**9カテゴリ**に分類" ← category count also wrong |
| `advanced/skill-guide.md` | 189 | "31スキル全ロード = 約260,000トークン" |
| `getting-started/introduction.md` | 133 | "**8** の専門 AI エージェント + **31** のスキル" |
| `getting-started/introduction.md` | 156 | "**31 のスキル**: さまざまな開発シナリオ" |
| `getting-started/introduction.md` | 163 | "8 の専門 AI エージェントと 31 のスキル" |
| `core-concepts/what-is-moai-adk.md` | 267 | "### 31個スキル (Progressive Disclosure)" |

**ja locale `advanced/skill-guide.md` is an ENTIRELY DIFFERENT (pre-v3 fictional) taxonomy.** It claims "9 categories" and references 46+ skill names that do not exist in the current codebase:

Fictional categories in ja: Foundation(5), Workflow(11), Domain(4), Language(16), Platform(4), Library(4), Tool(2), Framework(1), Design Tools(1).

Nonexistent skill names referenced in ja (sample):
- `moai-foundation-claude`, `moai-foundation-philosopher`, `moai-foundation-context` (Foundation has only 4: core/cc/quality/thinking)
- `moai-lang-*` (16 entries — programming-language support is via `rules/moai/languages/`, NOT separate skills, per the en docs line 65)
- `moai-platform-*` (`moai-platform-auth`, `moai-platform-supabase`, etc. — do not exist)
- `moai-library-*` (`moai-library-nextra`, `moai-library-shadcn` — do not exist)
- `moai-framework-electron`, `moai-design-tools`, `moai-tool-ast-grep`, etc.

This is the old pre-v3 MoAI-ADK skill structure that was never reconciled when the catalog was consolidated to the 7-prefix / 32-skill / 6-category structure. The ja `advanced/skill-guide.md` requires a full structural rewrite, not a count patch.

Note: ja locale `what-is-moai-adk.md` lines 7, 48, 652 do NOT carry the "31" claim (verified — the grep returned only line 267 for that file). This asymmetry is itself a parity issue to resolve.

### §2.5 zh locale (10 count-claim occurrences + 11 in-body fictional-name residuals)

**Count-claim inventory (10 occurrences across 4 pages):**

| File | Line | Excerpt |
|------|------|---------|
| `advanced/builder-agents.md` | 15 | "31 个内置技能和 8 个代理" |
| `advanced/skill-guide.md` | 61 | "**31 个技能** — moai 伞形路由器加上 30 个专业技能,分为 6 大类: Foundation(4)、Workflow(10)、Domain(8)、Reference(5)、Meta-Harness(2)、Design(1)" |
| `advanced/skill-guide.md` | 160 | "31 个技能全部加载 = 约 260,000 tokens" |
| `getting-started/introduction.md` | 133 | "**8** 个专业 AI 代理 + **31** 项技能" |
| `getting-started/introduction.md` | 156 | "**31 项技能**: 支持各种开发场景" |
| `getting-started/introduction.md` | 163 | "8 个专业 AI 代理和 31 项技能" |
| `getting-started/update.md` | 396 | `"🗿 MoAI-ADK: 8个专业代理 + 31个技能的SPEC-First DDD"` |
| `core-concepts/what-is-moai-adk.md` | 7 | "8个专业AI Agent和31个技能协同" |
| `core-concepts/what-is-moai-adk.md` | 48 | "**8个** 专业AI Agent + **31个** 技能" |
| `core-concepts/what-is-moai-adk.md` | 267 | "### 31个技能 (Progressive Disclosure)" |
| `core-concepts/what-is-moai-adk.md` | 661 | "skills/moai-*/ # 31个技能模块" |

zh `skill-guide.md` line 61 carries the most detailed sub-count breakdown: "Foundation(4)、Workflow(10)、Domain(8)、Reference(5)、Meta-Harness(2)、Design(1)" — note Domain(8) is wrong (should be 9), and the sum 4+10+8+5+2+1 = 30 specialized (should be 31). This is the clearest single-line correction target and the mechanical anchor for AC-005.

**In-body fictional-name residuals (11 matches — iter-2 additional finding):**

The zh `advanced/skill-guide.md` category STRUCTURE is correct (6 canonical headers), but the page body carries 11 references to nonexistent skill names inside the same conceptual-illustration sites as en (Mermaid nodes L45/L178/L325/L327, explicit-invocation code L211/L213, ASCII tree L236, frontmatter example L258, auto-load comments L301/L306, closing callout L360). Same drift class as en — covered by REQ-009 / AC-011.

```bash
$ grep -cE 'moai-lang-|moai-platform-|moai-library-|moai-framework-|moai-foundation-claude|moai-foundation-philosopher|moai-foundation-context' \
    docs-site/content/zh/advanced/skill-guide.md
11
```

### §2.6 4-locale fictional-name inventory summary (iter-2 locale-complete)

The fictional-name regex `moai-lang-|moai-platform-|moai-library-|moai-framework-|moai-foundation-claude|moai-foundation-philosopher|moai-foundation-context` was applied per-locale to `advanced/skill-guide.md` (the page-family that carries category structure). The inventory is locale-complete:

| Locale | Fictional-name matches | Category structure | Drift class | Owning REQ/AC |
|--------|------------------------|--------------------|-------------|---------------|
| en | 10 | 6 canonical ✓ (correct headers) | in-body residual (Mermaid / code / tree / callout) | REQ-009 / AC-011 |
| ko | 37 | 9 fictional categories ✗ | structural rewrite (full taxonomy replacement) | REQ-005 (ko scope) / AC-006 |
| ja | 37 | 9 fictional categories ✗ | structural rewrite (full taxonomy replacement) | REQ-005 (ja scope) / AC-006 |
| zh | 11 | 6 canonical ✓ (correct headers) | in-body residual (Mermaid / code / tree / callout) | REQ-009 / AC-011 |

**Locale-complete total: 95 fictional-name matches across 4 locales.** The ko and ja structural rewrites (REQ-005, unified ja+ko scope) eliminate 37+37 = 74; the en/zh in-body cleanup (REQ-009) eliminates 10+11 = 21. AC-006 covers ko+ja (structural), AC-011 covers en+zh (in-body). No locale is left unaccounted for.

```bash
# Locale-complete inventory reproduction (per-locale, never glob)
for loc in en ko ja zh; do
  n=$(grep -cE 'moai-lang-|moai-platform-|moai-library-|moai-framework-|moai-foundation-claude|moai-foundation-philosopher|moai-foundation-context' "docs-site/content/$loc/advanced/skill-guide.md")
  echo "$loc: $n"
done
# iter-2 baseline output (pre-run-phase):
# en: 10
# ko: 37
# ja: 37
# zh: 11
```

---

The edit surface is bounded by which pages carry factual skill-count claims. A page is "facts-bearing" if it contains a digit + skill-adjacent keyword (skill/스킬/スキル/技能). Pages mentioning skills only in structural prose (nav, breadcrumbs) without a numeric count are out of scope.

### §3.1 Page-family × locale matrix

| Page-family | en | ko | ja | zh | Total |
|-------------|----|----|----|----|-------|
| `advanced/builder-agents.md` | ✓ (L15) | ✓ (L16) | ✓ (L17) | ✓ (L15) | 4 |
| `advanced/skill-guide.md` | ✓ (L65,127,167) **+ in-body cleanup (10 fictional names)** | ✓ (L62,194) **+ structural rewrite (37 fictional names, 9→6 categories)** | ✓ (L59,189) **+ structural rewrite (37 fictional names, 9→6 categories)** | ✓ (L61,160) **+ in-body cleanup (11 fictional names)** | 4 (+ko rewrite, +ja rewrite) |
| `getting-started/introduction.md` | ✓ (L133,156,163) | ✓ (L133,156,163) | ✓ (L133,156,163) | ✓ (L133,156,163) | 4 |
| `getting-started/update.md` | ✓ (L396, statusline string) | (file present, no statusline string — no count correction needed) | (file present, no statusline string — no count correction needed) | ✓ (L396, statusline string) | 2 |
| `core-concepts/what-is-moai-adk.md` | ✓ (L7,48,267,652) | ✓ (L7,48,267,652) | ✓ (L267 only — asymmetric) | ✓ (L7,48,267,661) | 4 |

**Total facts-bearing pages: 18** (en:5, ko:4, ja:4, zh:5). `update.md` exists in all 4 locales, but only en/zh `update.md` carry the `31 <skill-count>` statusline string; ko/ja `update.md` do not carry a skill-count claim and require no count correction in this page-family. The ko and ja `advanced/skill-guide.md` both require full structural rewrites (9 fictional categories → 6 canonical, 37 nonexistent skill names eliminated, Reference/Meta-Harness/Design introduced). The en and zh `advanced/skill-guide.md` have correct category structure but require in-body cleanup of residual fictional names (10 and 11 respectively). The ja `what-is-moai-adk.md` is asymmetric (only line 267, missing lines 7/48/652 equivalents — a parity gap to resolve during run-phase).

### §3.2 Page-family classification rationale

- **`advanced/skill-guide.md`**: the primary skill-count authority page. Carries the most detailed breakdown (category sub-counts). ko and ja versions need structural rewrite (§2.3, §2.4); en and zh versions need in-body cleanup of residual fictional names (§2.2, §2.5). All 4 locales need the Domain sub-count 8→9 and `humanize` addition.
- **`advanced/builder-agents.md`**: references skill count as context for builder agents ("in addition to the N built-in skills"). Count correction only.
- **`getting-started/introduction.md`**: high-level "8 agents + N skills" overview. Count correction only.
- **`getting-started/update.md`**: statusline string literal in a code block. `update.md` exists in all 4 locales; only en/zh carry the `31 <skill-count>` statusline string and need count correction. ko/ja `update.md` carry no skill-count claim and need no count correction.
- **`core-concepts/what-is-moai-adk.md`**: conceptual overview + directory-tree ASCII art. Count correction in prose (L7,48,267) + tree comment (L652/661).

---

## §4. Category Structure Verification

### §4.1 en/zh category structure (6 canonical categories — correct headers, wrong Domain sub-count, in-body fictional-name residuals)

> **iter-2 correction (2026-06-18):** the iter-1 prose of this section claimed "en/ko/zh correctly identify 6 categories". That claim was **independently refuted**: ko does NOT have 6 canonical categories — ko has the same 9-fictional-category pre-v3 taxonomy as ja (see §2.3). The claim is true for en and zh only.

The en and zh locales correctly identify **6 canonical category headers**: Foundation, Workflow, Domain, Reference, Meta/Harness, Design. The structure (category count + category names) is correct. Two defects remain for en/zh:

1. **Domain sub-count wrong (8 → should be 9)**: the Domain category lists 8 skills, missing `moai-domain-humanize`. Correcting this to 9 makes the specialized total 31 (was 30).
2. **Total count wrong (31 → should be 32)**: every page carries "31 skills"; should be "32 skills" (umbrella router + 31 specialized).
3. **In-body fictional-name residuals (en: 10, zh: 11)**: category LISTING TABLES are correct (only real skills listed), but conceptual illustrations elsewhere in `advanced/skill-guide.md` (Mermaid flowchart nodes, explicit-invocation code examples, ASCII directory trees, frontmatter example, auto-load scenario comments, closing callouts) reference `moai-lang-*` / `moai-library-*` / `moai-platform-*` as if they were shipped skills. These are covered by REQ-009 / AC-011.

### §4.2 ko and ja category structure (9 fictional categories — entirely pre-v3 taxonomy, identical divergence)

The ko and ja locales both claim **9 categories** including 5 that do not exist in the current catalog: Language, Platform, Library, Tool, Framework, plus a mis-labeled "Design Tools" (canonical is "Design"). The actual structure is 6 canonical categories (Foundation, Workflow, Domain, Reference, Meta/Harness, Design). Programming-language support is delivered via `rules/moai/languages/` (NOT separate skills), per the en docs line 65 verbatim: "Programming-language support is delivered through rules under `rules/moai/languages/`, not as separate skills."

Both ko and ja `advanced/skill-guide.md` carry 37 nonexistent skill-name references each (74 combined) — `moai-lang-*`, `moai-platform-*`, `moai-library-*`, `moai-framework-*`, `moai-foundation-claude/philosopher/context`. The ko and ja "Language (16)" categories claiming `moai-lang-go`, `moai-lang-python`, etc. as skills are the most severe single-page drifts in the docs-site, tied in magnitude.

Both ko and ja require full structural rewrites (REQ-005 unified ja+ko scope, AC-006): replace the 9-fictional-category taxonomy with the canonical 6-category / 32-skill structure, eliminate all 37 nonexistent references each, and introduce the 3 missing canonical categories (Reference, Meta/Harness, correctly-labeled Design).

---

## §5. Tier Justification

**Tier L** is justified:

1. **4-locale × multi-page-family**: 18 facts-bearing pages across 4 locales, with per-locale independent grep verification required (digit-boundary discipline).
2. **One locale (ja) needs structural rewrite**: `advanced/skill-guide.md` is not a count patch — it is a full taxonomy replacement (9 fictional categories → 6 canonical categories).
3. **Primary-source evidence depth**: research.md must carry verbatim `find`/`ls` output for count defensibility (defect-claim verification lesson).
4. **Cross-SPEC dependency**: explicit follow-up to DOCSITE-001 (depends_on frontmatter), cohort context requires full artifact set.

Tier M (3-file) would be insufficient because the ja structural divergence + 4-locale parity + primary-source evidence justify the research.md depth. design.md is omitted — factual reconciliation involves no architectural decisions.

---

## §6. Applied Lessons Incorporated

| Lesson | How applied |
|--------|-------------|
| `feedback_digit_boundary_locale_grep_4parity` | §2.1 methodology — `(^|[^0-9])` prefix anchor, per-locale loop, locale-native idioms. Each AC in acceptance.md reproduces grep evidence per locale independently. |
| `feedback_defect_claim_verification` | §1 — canonical count is a DEFECT CLAIM until `find ... | wc -l` confirms it. Verbatim output recorded. No REQ asserts "32" without the primary-source command evidence backing it. |
