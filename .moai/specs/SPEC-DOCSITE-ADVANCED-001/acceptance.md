# acceptance.md — SPEC-DOCSITE-ADVANCED-001

> 33 Acceptance Criteria organized in 8 groups (A-H). Each AC is observable via
> a runnable command (CMD-*). All AC are must_pass (no nice-to-have AC in this
> SPEC). Run-phase manager-develop reports each AC in the §E.1 PASS/FAIL matrix
> with verbatim command output (per verification-claim-integrity §3 5-section
> format).

## §A Given-When-Then scenarios (representative)

### AC-DA-001 — 4-locale file-existence parity (24 files)

**Given** the docs-site repository at run-phase completion
**When** the verifier lists all advanced markdown files for the 6 new slugs across 4 locales
**Then** exactly 24 files exist at the paths `docs-site/content/{ko,en,ja,zh}/advanced/{tokenomics-overview,token-budget,no-haiku-3tier,plan-type-profiles,self-evolving,autonomous-loops}.md`

### AC-DA-040 — _meta.yaml parity debt closed (existing 14 entries registered everywhere)

**Given** the docs-site content tree
**When** the verifier counts entries in each locale's `advanced/_meta.yaml`
**Then** all 4 locales carry entries for ALL 14 existing advanced content files (catalog-system, decision-memory, harness-profiles, harness-v4-builder, hooks-reference included) — closing the pre-existing parity debt

### AC-DA-070 — Hugo build warning-free

**Given** the docs-site with all 29 file changes applied (24 new + 4 _meta + 1 main.yaml)
**When** the verifier runs `cd docs-site && hugo --minify --gc`
**Then** the command exits 0 with ZERO warnings (a malformed `_meta.yaml` or `main.yaml` entry surfaces here)

### AC-DA-060 — GLM overlay honesty caveat present

**Given** the `plan-type-profiles.md` page in any of the 4 locales
**When** the verifier greps the page body
**Then** the page contains the "wire validity pending live verification" (or locale-equivalent) caveat — does NOT claim the GLM overlay "works guaranteed"

### AC-DA-062 — Native /goal vs /moai goal distinction

**Given** the `autonomous-loops.md` page in any of the 4 locales
**When** the verifier greps the page body
**Then** both `/goal` (native, HUMAN-ONLY) and `/moai goal` (programmatic, MoAI-owned) are mentioned with the distinction that one is a user-TUI command and the other is an orchestrator-owned subcommand

---

## §B AC Matrix (33 AC, all must_pass)

### Group A — 4-Locale Content Parity (4 AC, Critical)

| AC ID | Requirement | CMD | Severity |
|-------|-------------|-----|----------|
| AC-DA-001 | 24 new files exist (6 slugs × 4 locales) | CMD-A | Critical |
| AC-DA-002 | Each new file has Hugo frontmatter (`title:` field) | CMD-A1 | Critical |
| AC-DA-003 | Cross-locale semantic parity (same sections, same diagrams, locale-translated prose) | CMD-A2 | Critical |
| AC-DA-004 | Canonical-KO is the source for each of the 6 slugs (ko file exists for every slug) | CMD-A | Critical |

### Group B — 6 Pages Canonical KO Authored (6 AC, Critical)

| AC ID | Requirement | CMD | Severity |
|-------|-------------|-----|----------|
| AC-DA-010 | `ko/advanced/tokenomics-overview.md` documents the 3-pillar narrative with Tokenomics as v3.0 pillar | CMD-B1 | Critical |
| AC-DA-011 | `ko/advanced/token-budget.md` documents the 4-layer token economy + `/clear` thresholds + paste-ready resume | CMD-B2 | Critical |
| AC-DA-012 | `ko/advanced/no-haiku-3tier.md` documents the 3-tier (Sonnet/Opus/Fable, no Haiku) + DeepSWE rationale + ApplyTierProfile | CMD-B3 | Critical |
| AC-DA-013 | `ko/advanced/plan-type-profiles.md` documents plan_type axis + 60-cell profile (CLOSED source) | CMD-B4 | Critical |
| AC-DA-014 | `ko/advanced/self-evolving.md` documents ACE 3-Loop (Loop 0/1/2) + EVOLVE-001/002/003 closed | CMD-B5 | Critical |
| AC-DA-015 | `ko/advanced/autonomous-loops.md` documents /goal + /moai goal + /moai loop distinct semantics | CMD-B6 | Critical |

### Group C — Derivation Chain (3 AC, Critical)

| AC ID | Requirement | CMD | Severity |
|-------|-------------|-----|----------|
| AC-DA-020 | 4-locale derivation chain followed (ko→en→ja/zh) per i18n canonical-locale rule | CMD-C | Critical |
| AC-DA-021 | Derived-locale pages translate prose, preserve code/diagrams verbatim | CMD-C1 | Critical |
| AC-DA-022 | Derived-locale pages use the locale-specific name field from main.yaml when referencing nav | CMD-C2 | Critical |

### Group D — Navigation Registration (4 AC, Critical)

| AC ID | Requirement | CMD | Severity |
|-------|-------------|-----|----------|
| AC-DA-030 | `main.yaml` Advanced section has 6 new sub-entries with 4-locale name fields | CMD-D | Critical |
| AC-DA-031 | Each per-locale `_meta.yaml` has the 6 new entries | CMD-D1 | Critical |
| AC-DA-032 | Slug registered in `main.yaml` implies slug registered in ALL 4 `_meta.yaml` (no partial registration) | CMD-D2 | Critical |
| AC-DA-033 | New entries appear in pillar order (Tokenomics → Harness → Loop) per design.md §C | CMD-D3 | Critical |

### Group E — Parity Debt Pre-Fix (3 AC, Critical)

| AC ID | Requirement | CMD | Severity |
|-------|-------------|-----|----------|
| AC-DA-040 | All 4 locales' `_meta.yaml` register the existing 14 advanced content files | CMD-E | Critical |
| AC-DA-041 | `main.yaml` Advanced section unchanged for the 14 existing entries (debt is in `_meta.yaml` only) | CMD-E1 | Critical |
| AC-DA-042 | No new content files were created for existing slugs (debt was `_meta.yaml` registration only — verified at plan-phase, re-verified at run-phase M0) | CMD-E2 | Critical |

### Group F — Design Regime (6 AC, High)

| AC ID | Requirement | CMD | Severity |
|-------|-------------|-----|----------|
| AC-DA-050 | Zero body-emoji in the 24 new files outside fenced code blocks | CMD-F | High |
| AC-DA-051 | Typography arrows `→ ← ↓ ✓ ✗` preserved verbatim (NOT stripped) | CMD-F1 | High |
| AC-DA-052 | Zero `flowchart LR|RL` or `graph LR|RL` in the 24 new files | CMD-G | High |
| AC-DA-053 | Zero blacklisted URLs (`docs.moai-ai.dev|adk.moai.com|adk.moai.kr`) in the 24 new files | CMD-D-url | High |
| AC-DA-054 | Emphasis-marker spacing rule honored (`**한글** (English)` form, not `**한글(English)**`) | CMD-F2 | High |
| AC-DA-055 | No `[data-theme="dark"]` branching in the 24 new files (light single theme) | CMD-F3 | High |

### Group G — Source Truthfulness (4 AC, High)

| AC ID | Requirement | CMD | Severity |
|-------|-------------|-----|----------|
| AC-DA-060 | `plan-type-profiles.md` (4 locales) carries the GLM wire-validity-pending caveat | CMD-H | High |
| AC-DA-061 | `no-haiku-3tier.md` (4 locales) distinguishes design report vs ApplyTierProfile live behavior | CMD-H1 | High |
| AC-DA-062 | `autonomous-loops.md` (4 locales) distinguishes native /goal (HUMAN-ONLY) vs /moai goal (programmatic) | CMD-I | High |
| AC-DA-063 | `self-evolving.md` (4 locales) discloses EVOLVE-004/005 as roadmap (not yet implemented) | CMD-H2 | High |

### Group H — Hugo Build & Site Integrity (3 AC, Critical)

| AC ID | Requirement | CMD | Severity |
|-------|-------------|-----|----------|
| AC-DA-070 | `hugo --minify --gc` exits 0 with 0 warnings | CMD-J | Critical |
| AC-DA-071 | 24 new pages appear in `sitemap.xml` at `/locale/advanced/slug/` paths | CMD-J1 | Critical |
| AC-DA-072 | Zero "page not found" warnings in hugo build for the 6 new slugs (4-locale _meta/main consistency) | CMD-J | Critical |

---

## §C Edge cases

- **Edge-1 — main.yaml comment header**: the file's comment header says `# Auto-generated by scripts/gen_menu.py` — but `gen_menu.py` DOES NOT exist (per `hns-oss-docs-structure-map` skill). This is a stale comment. Do NOT attempt to regenerate the file via the non-existent script; edit manually. The stale comment is preserved (out of scope to clean up).
- **Edge-2 — _meta.yaml counting rule**: geekdoc's `_meta.yaml` uses keyless top-level entries (`"slug":\n  title: "..."`). The count is the number of such entries (excluding the `index:` block). CMD-A/E pin the exact grep rule.
- **Edge-3 — Typography arrows vs body emoji**: `→ ← ↓ ✓ ✗` are NOT body emoji (they are typography) and MUST be preserved. CMD-F excludes them via the fenced-code-block boundary (emoji outside ```` ``` ```` is the violation).
- **Edge-4 — MoAI banner reproduction**: the orchestrator output-style banner contains emoji (`🤖 🗰 📋 🎯`) inside fenced code blocks when reproducing orchestrator output. These are preserved verbatim per i18n rule §4.
- **Edge-5 — vercel.json untouched**: no redirects needed for NEW pages. Do NOT add redirect entries.
- **Edge-6 — icon variant syntax**: `{{</* icon check ok */>}}` (name + variant). Variant is one of `ok|warn|danger|primary|muted`. Variant is optional but recommended for semantic color.
- **Edge-7 — menu placement collision**: the 6 new entries prepend the existing 14. If the prepended order causes a geekdoc warning ("duplicate weight" or similar), the entries may need explicit `weight:` fields. The verify recipe (CMD-J) catches this.
- **Edge-8 — concurrent session on `_meta.yaml`**: B.15 hazard. If `moai session list` returns an entry for this SPEC ID from a different session, return a blocker report.

---

## §D Quality gate criteria (run-phase exit gate)

The exit gate is the `hns-oss-docs-verify` recipe (loaded at run-phase start). All checks below MUST pass; failure on any = blocker report (do NOT commit M6 until fixed):

### CMD-A — 4-locale file-existence parity (24 new files)

```bash
count=0
for loc in ko en ja zh; do
  for slug in tokenomics-overview token-budget no-haiku-3tier plan-type-profiles self-evolving autonomous-loops; do
    f="docs-site/content/$loc/advanced/$slug.md"
    [ -f "$f" ] && count=$((count+1)) || echo "MISSING: $f"
  done
done
echo "count=$count"
# Expected: count=24
```

### CMD-A1 — Hugo frontmatter present

```bash
for loc in ko en ja zh; do
  for slug in tokenomics-overview token-budget no-haiku-3tier plan-type-profiles self-evolving autonomous-loops; do
    f="docs-site/content/$loc/advanced/$slug.md"
    head -5 "$f" | grep -qE '^title:' || echo "MISSING title: in $f"
  done
done
# Expected: 0 "MISSING title:" lines
```

### CMD-A2 — Cross-locale semantic parity (section heading count)

```bash
for slug in tokenomics-overview token-budget no-haiku-3tier plan-type-profiles self-evolving autonomous-loops; do
  ko_h2=$(grep -cE '^## ' "docs-site/content/ko/advanced/$slug.md")
  for loc in en ja zh; do
    loc_h2=$(grep -cE '^## ' "docs-site/content/$loc/advanced/$slug.md")
    [ "$ko_h2" = "$loc_h2" ] || echo "PARITY GAP: $slug ko=$ko_h2 $loc=$loc_h2"
  done
done
# Expected: 0 "PARITY GAP" lines (H2 section count matches across locales)
```

### CMD-B1 through CMD-B6 — canonical-KO content markers

```bash
# CMD-B1 tokenomics-overview: 3-pillar narrative + Tokenomics pillar
grep -qE '(Tokenomics|토크노믹스).*(Agentic Loop|에이전틱 루프).*(Agentic Harness|에이전틱 하네스)' \
  docs-site/content/ko/advanced/tokenomics-overview.md \
  || grep -cE '## ' docs-site/content/ko/advanced/tokenomics-overview.md
# Either the regex matches OR the H2 count is ≥3 (substantive page)

# CMD-B2 token-budget: 4-layer + /clear threshold + paste-ready
f=docs-site/content/ko/advanced/token-budget.md
grep -qE '(계층|층|layer)' "$f" && grep -qE '(/clear|클리어)' "$f" && grep -qE '(paste-ready|붙여넣기|이어서)' "$f" \
  && echo "PASS" || echo "FAIL"

# CMD-B3 no-haiku-3tier: Sonnet/Opus/Fable + DeepSWE + ApplyTierProfile
f=docs-site/content/ko/advanced/no-haiku-3tier.md
grep -qE '(Sonnet|소넷)' "$f" && grep -qE '(Opus|오퍼스)' "$f" && grep -qE '(Fable|페이블)' "$f" \
  && grep -qE '(DeepSWE|리더보드)' "$f" && grep -qE 'ApplyTierProfile' "$f" \
  && echo "PASS" || echo "FAIL"

# CMD-B4 plan-type-profiles: plan_type api|subscription + 60-cell
f=docs-site/content/ko/advanced/plan-type-profiles.md
grep -qE 'plan_type' "$f" && grep -qE '(api|subscription|구독|종량제)' "$f" \
  && grep -qE '(60|프로필|profile)' "$f" && echo "PASS" || echo "FAIL"

# CMD-B5 self-evolving: ACE + Loop 0/1/2
f=docs-site/content/ko/advanced/self-evolving.md
grep -qE '(ACE|Loop 0|Loop 1|Loop 2|관찰|반추|승격)' "$f" && echo "PASS" || echo "FAIL"

# CMD-B6 autonomous-loops: /goal + /moai goal + /moai loop
f=docs-site/content/ko/advanced/autonomous-loops.md
grep -qE '^/goal | `/goal' "$f" && grep -qE '/moai goal' "$f" && grep -qE '/moai loop' "$f" \
  && echo "PASS" || echo "FAIL"
```

### CMD-C, C1, C2 — derivation chain (structural)

(See CMD-A2 for parity; CMD-C verifies the chain conceptually by confirming en/ja/zh derive from ko via the locale-translator specialist's output — no machine-translator artifact exists.)

### CMD-D — main.yaml Advanced section 6 new entries

```bash
# Extract the Advanced sub-block and count new slug refs
awk '/ref: \/advanced$/,/ref: \/contributing$/' docs-site/data/menu/main.yaml \
  | grep -cE 'ref: /advanced/(tokenomics-overview|token-budget|no-haiku-3tier|plan-type-profiles|self-evolving|autonomous-loops)'
# Expected: 6
```

### CMD-D1 — per-locale _meta.yaml 6 new entries

```bash
for loc in ko en ja zh; do
  f="docs-site/content/$loc/advanced/_meta.yaml"
  echo "=== $loc ==="
  for slug in tokenomics-overview token-budget no-haiku-3tier plan-type-profiles self-evolving autonomous-loops; do
    grep -qE "^\"?$slug\"?:" "$f" && echo "  ✓ $slug" || echo "  ✗ MISSING $slug"
  done
done
# Expected: 4 locales × 6 = 24 ✓, 0 ✗
```

### CMD-D2 — main.yaml ⟷ _meta.yaml registration consistency

```bash
# For each new slug registered in main.yaml, ALL 4 _meta.yaml must have it
for slug in tokenomics-overview token-budget no-haiku-3tier plan-type-profiles self-evolving autonomous-loops; do
  for loc in ko en ja zh; do
    f="docs-site/content/$loc/advanced/_meta.yaml"
    grep -qE "^\"?$slug\"?:" "$f" || echo "INCONSISTENT: $slug missing from $loc _meta.yaml"
  done
done
# Expected: 0 INCONSISTENT lines
```

### CMD-E — _meta.yaml parity debt closed (existing 14 entries registered everywhere)

```bash
existing_slugs="skill-guide agent-guide builder-agents hooks-guide hooks-reference settings-json security-notes statusline claude-md-guide harness-profiles catalog-system decision-memory harness-v4-builder ultracode-workflows"
for loc in ko en ja zh; do
  f="docs-site/content/$loc/advanced/_meta.yaml"
  for slug in $existing_slugs; do
    grep -qE "^\"?$slug\"?:" "$f" || echo "PARITY DEBT REMAINING: $loc $slug"
  done
done
# Expected: 0 "PARITY DEBT REMAINING" lines
```

### CMD-F — body-emoji absence (outside fenced code blocks)

```bash
# Strategy: use awk to skip fenced code blocks, then grep for emoji ranges
for loc in ko en ja zh; do
  for slug in tokenomics-overview token-budget no-haiku-3tier plan-type-profiles self-evolving autonomous-loops; do
    f="docs-site/content/$loc/advanced/$slug.md"
    awk '/^```/{flag=!flag; next} !flag' "$f" \
      | grep -nP '[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}\x{2190}-\x{21FF}\x{2B00}-\x{2BFF}]' \
      | grep -vP '[→←↓↑✓✗]' \
      && echo "BODY EMOJI: $f" || true
  done
done
# Expected: 0 "BODY EMOJI" lines (typography arrows explicitly excluded)
```

### CMD-G — Mermaid TD-only

```bash
grep -rnE 'flowchart (LR|RL)|graph (LR|RL)' docs-site/content/{ko,en,ja,zh}/advanced/{tokenomics-overview,token-budget,no-haiku-3tier,plan-type-profiles,self-evolving,autonomous-loops}.md
# Expected: 0 matches
```

### CMD-H — GLM honesty caveat (plan-type-profiles.md, 4 locales)

```bash
# English anchor + locale-equivalents
for loc in ko en ja zh; do
  f="docs-site/content/$loc/advanced/plan-type-profiles.md"
  # Look for any of: "wire validity pending", "wire effectiveness pending", "실증 예정", "実証予定", "待验证"
  grep -qE '(wire (validity|effectiveness).*(pending|unverified)|실증.*예정|実証.*予定|待.*验证)' "$f" \
    && echo "✓ $loc caveat present" || echo "✗ $loc caveat MISSING"
done
# Expected: 4 ✓, 0 ✗
```

### CMD-I — Native /goal vs /moai goal distinction (autonomous-loops.md, 4 locales)

```bash
for loc in ko en ja zh; do
  f="docs-site/content/$loc/advanced/autonomous-loops.md"
  has_native=$(grep -cE '(/goal|HUMAN-ONLY|사용자 전용|ユーザー専用|用户专用)' "$f")
  has_programmatic=$(grep -cE '(/moai goal|PROGRAMMATIC|프로그래밍|プログラム的|编程)' "$f")
  echo "$loc: native=$has_native programmatic=$has_programmatic"
  [ "$has_native" -ge 1 ] && [ "$has_programmatic" -ge 1 ] || echo "✗ $loc distinction MISSING"
done
# Expected: 0 ✗ lines
```

### CMD-J — Hugo build warning-free

```bash
mkdir -p .moai/state/verify/SPEC-DOCSITE-ADVANCED-001/
cd docs-site && hugo --minify --gc 2>&1 | tee ../.moai/state/verify/SPEC-DOCSITE-ADVANCED-001/M6-hugo-build.log
exit=$?
cd ..
echo "hugo exit=$exit"
tail -20 .moai/state/verify/SPEC-DOCSITE-ADVANCED-001/M6-hugo-build.log
# Expected: exit=0, zero "WARN" or "ERROR" lines in tail
```

### CMD-J1 — sitemap contains 24 new paths

```bash
for loc in ko en ja zh; do
  for slug in tokenomics-overview token-budget no-haiku-3tier plan-type-profiles self-evolving autonomous-loops; do
    grep -q "/$loc/advanced/$slug/" docs-site/public/sitemap.xml || echo "SITEMAP MISSING: /$loc/advanced/$slug/"
  done
done
# Expected: 0 "SITEMAP MISSING" lines
```

---

## §E Definition of Done

The SPEC is "done" when ALL 33 AC PASS AND:

1. **Implementation Kickoff Approval was obtained** before M1 (per §19.1 — mandatory and score-independent).
2. **All 6 milestones (M1-M6) committed** with Conventional Commit subjects + `🗿 MoAI` trailer.
3. **Commits pushed to origin/main** (Route A — Hybrid Trunk main-direct) OR a PR opened (Route B — only if user explicitly requested `--pr`).
4. **progress.md §E.2 / §E.3** populated by manager-develop with run-phase evidence (commit SHAs, AC PASS matrix, hugo build log path).
5. **Zero unobserved claims**: every AC PASS row cites a verbatim command output (per `verification-claim-integrity.md` §1.1 surface 2 — manager-agent §E self-verification).
6. **No blocker reports outstanding**: any blocker surfaced during run-phase was either resolved or escalated via the orchestrator's AskUserQuestion channel.

### Sync-phase entry (after this SPEC's run-phase completes)

Sync-phase is owned by manager-docs (per the Status Transition Ownership Matrix). The single sync commit carries:
- `implemented → completed` frontmatter transition in spec.md
- `sync_commit_sha` population in progress.md §E.4
- CHANGELOG entry (per B.12 discipline)
- The 3-phase close (NO separate Mx commit; MX Tag validation is a sync sub-step).

---

Version: 0.1.0 | AC count: 33 (all must_pass) | Status: draft (plan-phase)
