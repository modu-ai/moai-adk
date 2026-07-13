# SPEC-DOCSITE-E2E-001 — Acceptance Criteria

> SSOT for the AC matrix. Every AC is verified by an executed command from § Executable Command Block (CMD-*) — commands live in fenced blocks, NEVER in table cells (escaped-pipe vacuity hazard). All ring baselines cited below are plan-phase measurements (2026-07-13); run-phase pre-flight re-measures them (AC-DSE-013) and the reconciled numbers govern.

## §D AC Matrix

| AC | REQ | Criterion (observable) | Verify via |
|----|-----|------------------------|------------|
| AC-DSE-001 | REQ-DSE-001 | `moai-e2e.md` exists in all 4 locale trees under `utility-commands/` | CMD-D |
| AC-DSE-002 | REQ-DSE-002 | The ko+en pages each contain the platform matrix (Playwright, Maestro, Appium, Electron, Tauri all named), auto-detection prose, token-minimization prose, and the desktop-native deferral notice | CMD-D2 |
| AC-DSE-003 | REQ-DSE-003 | All 4 locale page files + `_meta.yaml` edits + menu edit land in the same run-phase commit set (no locale left behind at push time) | CMD-E6 (git show --stat per commit) |
| AC-DSE-004 | REQ-DSE-004 | All 4 locale pages carry `title: /moai e2e`, `draft: false`, identical `weight`; H2 count identical across locales | CMD-D3 |
| AC-DSE-005 | REQ-DSE-005 | `"moai-e2e"` entry present in all 4 locale `_meta.yaml` files | CMD-D |
| AC-DSE-006 | REQ-DSE-006 | main.yaml utility-commands section carries the moai-e2e sub-entry with ko/en/ja/zh name keys + ref | CMD-E |
| AC-DSE-007 | REQ-DSE-007 | `layouts/partials/menu.html` is UNCHANGED (no new icon value introduced) | CMD-E2 |
| AC-DSE-008 | REQ-DSE-008 | New pages: 0 blacklisted URLs, 0 Mermaid LR/RL, 0 internal SPEC/REQ tokens, 0 decorative body emoji | CMD-H |
| AC-DSE-009 | REQ-DSE-009 | New pages follow emphasis-spacing rule; no `data-theme="dark"` additions anywhere in the diff | CMD-H2 |
| AC-DSE-010 | REQ-DSE-100 | Ring A invariance: English ERE family grep over docs-site/ → 0 matches | CMD-A |
| AC-DSE-011 | REQ-DSE-101 | Ring A′: every catalog-count match updated (incl. `22 → 17 → 8 → 10` → `… → 11`); residual matches are ONLY the documented false positives, enumerated verbatim in progress.md §E.2 | CMD-A2 |
| AC-DSE-012 | REQ-DSE-102 | Ring B/B′ invariance per locale: 10-family and 9-family digit-boundary greps → 0 catalog-count residuals (documented false positives enumerated) | CMD-B, CMD-B2 |
| AC-DSE-013 | REQ-DSE-104 | Run-phase pre-flight re-measurement executed BEFORE first edit; reconciled inventory (counts + per-match classification) recorded in progress.md §E.2 | CMD-PRE-1..4 outputs quoted in §E.2 |
| AC-DSE-014 | REQ-DSE-103 | `e2e-specialist` NAMED in every applicable enumeration surface (per reconciled inventory), each locale variant included | CMD-C |
| AC-DSE-015 | REQ-DSE-105 | `layouts/index.html` stat row renders 11 (agent count); comment updated | CMD-C2 |
| AC-DSE-016 | REQ-DSE-200 | `hugo --minify` exit 0, zero NEW warnings vs pre-flight baseline; `public/sitemap.xml` exists | CMD-F |
| AC-DSE-017 | REQ-DSE-201 | Built site nav reaches the new page in all 4 locales | CMD-G |
| AC-DSE-018 | REQ-DSE-202 + 203 | Entry precondition verified (clean docs-site tree at entry); every commit pathspec-scoped | CMD-E6 + progress.md §E.2 entry record |
| AC-DSE-019 | REQ-DSE-204 | `scripts/docs-i18n-check.sh` executed; PASS, or its failures classified pre-existing-baseline with evidence | CMD-F2 |
| AC-DSE-020 | REQ-DSE-205 | Diff touches ONLY docs-site/ + this SPEC directory | CMD-E6 (git diff --stat range check) |

## Given-When-Then Scenarios

### Scenario 1 — New user discovers /moai e2e documentation (happy path)
- **Given** the docs-site is built from the post-run tree,
- **When** a ko-locale user opens the sidebar utility-commands section,
- **Then** a `/moai e2e` entry renders (main.yaml sub-entry + ko name key + `_meta.yaml` entry all present), and the page at `/utility-commands/moai-e2e` documents the platform matrix, auto-detection, token-minimized execution, and the desktop-native deferral branch — and the same holds for en/ja/zh (AC-DSE-001/002/004/005/006/017).

### Scenario 2 — Agent-count consistency after normalization
- **Given** the run-phase sweep completed,
- **When** any docs-site page previously claiming the 10/9 catalog is read in any of the 4 locales,
- **Then** the page states the 11-catalog (11 retained = 10 MoAI-custom + 1 Explore) AND names `e2e-specialist` in its enumeration where the catalog members are listed; the ring invariance greps return 0 catalog-count residuals (AC-DSE-010/011/012/014).

### Scenario 3 — Surface drift between plan and run (edge)
- **Given** parallel sessions modified docs-site between plan authoring and run entry,
- **When** run-phase pre-flight re-measures the ring baselines (CMD-PRE-1..4) and they differ from the plan-phase numbers (12 / 14 / 19-19-21 / 9-9-9),
- **Then** the executor records the reconciled inventory in progress.md §E.2 and edits against the RE-MEASURED surface, not the plan-phase list; if the docs-site tree is not clean at entry, the executor STOPS and returns a blocker report (AC-DSE-013/018).

### Scenario 4 — False-positive protection (edge)
- **Given** a Ring A′/B match that is NOT a catalog-count claim (plan-confirmed example: `content/en/claude-code/agentic/agent-view.md` "Running 10 agents in parallel"),
- **When** the sweep classifies matches before editing,
- **Then** the non-catalog match remains byte-identical, and it is enumerated in the documented-false-positive list in progress.md §E.2 (AC-DSE-011/012).

## Edge Cases

1. **Menu name-map missing one locale key** → the sidebar silently falls back or renders empty for that locale; CMD-E asserts all 4 keys.
2. **`weight` divergence between locales** → per-locale ordering differs; CMD-D3 asserts identical weight values.
3. **ja/zh spacing conventions** — ja writes `11 個の` (spaces), zh `11 个`; digit-boundary greps must not assume ko spacing.
4. **Hugo pre-existing warnings** — gate is NEW-warning-free vs baseline, not absolute zero if baseline is dirty (pin baseline in §E.2).
5. **docs-i18n-check.sh failing on unrelated tree state** — classify as pre-existing baseline with evidence; do not scope-creep (AC-DSE-019).

## Executable Command Block

All commands run from the repo root (`/Users/goos/MoAI/moai-adk-go`) unless stated. Baselines in comments are plan-phase (2026-07-13).

```bash
# ---- Pre-flight re-measurement (REQ-DSE-104 / AC-DSE-013; run BEFORE first edit) ----

# CMD-PRE-1 — Ring A English ERE family (plan baseline: 12 matches / 10 files)
grep -rn -E "10 retained agents|9 MoAI-custom|10-agent" docs-site/ --include="*.md" --include="*.html" | wc -l
grep -rln -E "10 retained agents|9 MoAI-custom|10-agent" docs-site/ --include="*.md" --include="*.html" | sort

# CMD-PRE-2 — Ring A' English loose catalog forms (plan baseline: 14 matches incl. false positives)
grep -rn -iE "(^|[^0-9])10[^0-9]{0,12}agent" docs-site/content/en docs-site/layouts | grep -vE "10-agent|10 retained" | wc -l

# CMD-PRE-3 — Ring B locale-language 10-family, per locale (plan baseline: ko 19 / ja 19 / zh 21)
for L in ko ja zh; do echo -n "$L: "; grep -rn -E "(^|[^0-9])10(개|人|個|个)?[^0-9]{0,8}(에이전트|エージェント|智能体|代理)" "docs-site/content/$L" | wc -l; done

# CMD-PRE-4 — Ring B' locale-language 9-family, widened any-gap pattern (audit D1)
#             (plan baseline v0.1.1: ko 9 / ja 9 / zh 9 = 27 combined. The [^0-9]{0,8} gap is
#              REQUIRED — spaced ja/zh forms "9 個の MoAI" / "9 个 MoAI 自定义" escape a counter-only form)
for L in ko ja zh; do echo -n "$L: "; grep -rn -E "(^|[^0-9])9[^0-9]{0,8}(MoAI|커스텀|사용자 정의|カスタム|自定义)" "docs-site/content/$L" | wc -l; done

# ---- Post-change invariance (AC-DSE-010/011/012) ----

# CMD-A — Ring A invariance (expect: 0)
grep -rn -E "10 retained agents|9 MoAI-custom|10-agent" docs-site/ --include="*.md" --include="*.html" | wc -l

# CMD-A2 — Ring A' residuals (expect: ONLY documented false positives; list them verbatim)
grep -rn -iE "(^|[^0-9])10[^0-9]{0,12}agent" docs-site/content/en docs-site/layouts | grep -vE "11-agent|11 retained"

# CMD-B — Ring B invariance per locale (expect: 0 catalog-count residuals per locale)
for L in ko ja zh; do echo "--- $L ---"; grep -rn -E "(^|[^0-9])10(개|人|個|个)?[^0-9]{0,8}(에이전트|エージェント|智能体|代理)" "docs-site/content/$L"; done

# CMD-B2 — Ring B' invariance, widened pattern (expect: 0 catalog-count residuals; verified
#           non-self-tripping on post-change 10/11 forms — they contain no digit 9)
grep -rn -E "(^|[^0-9])9[^0-9]{0,8}(MoAI|커스텀|사용자 정의|カスタム|自定义)" docs-site/content/ko docs-site/content/ja docs-site/content/zh

# ---- Reachability (AC-DSE-014/015; NOT mere token presence) ----

# CMD-C — e2e-specialist named per enumeration surface (audit D5: iterate the EXPLICIT REQ-DSE-103
#          file list — never derive the iteration set from post-change 11-tokens, which is circular.
#          Expect: >=1 for every file the progress.md §E.2 reconciled inventory classifies as an
#          enumeration surface; 0 acceptable ONLY for files classified count-only in §E.2)
for base in advanced/agent-guide.md claude-code/agentic/sub-agents.md getting-started/introduction.md \
            getting-started/faq.md advanced/builder-agents.md advanced/claude-md-guide.md \
            core-concepts/harness-engineering.md core-concepts/what-is-moai-adk.md \
            multi-llm/model-policy.md workflow-commands/moai-harness.md; do
  for L in en ko ja zh; do
    f="docs-site/content/$L/$base"
    [ -f "$f" ] && { printf "%s: " "$f"; grep -c "e2e-specialist" "$f"; }
  done
done

# CMD-C2 — landing stat row (audit D6 expectations: the agent-count stat renders 11, AND
#           grep -c "e2e-specialist" = 0 is the CORRECT result — the stat row renders counts,
#           not member names; REQ-DSE-105 owns this surface, REQ-DSE-103 excludes it)
grep -n -E "(^|[^0-9])11([^0-9]|$)" docs-site/layouts/index.html | grep -iE "agent|에이전트"
grep -c "e2e-specialist" docs-site/layouts/index.html

# ---- New page + navigation (AC-DSE-001/004/005/006/007) ----

# CMD-D — 4-locale existence + meta parity (expect: 4 file paths + "1" x4)
for L in en ko ja zh; do
  ls "docs-site/content/$L/utility-commands/moai-e2e.md"
  grep -c '"moai-e2e"' "docs-site/content/$L/utility-commands/_meta.yaml"
done

# CMD-D2 — content coverage probe, ko+en (expect: each named toolchain >=1 per file)
for f in docs-site/content/ko/utility-commands/moai-e2e.md docs-site/content/en/utility-commands/moai-e2e.md; do
  echo "--- $f ---"
  for t in Playwright Maestro Appium Electron Tauri; do printf "%s: " "$t"; grep -ci "$t" "$f"; done
done

# CMD-D3 — frontmatter + H2 parity (expect: identical weight x4; identical H2 count x4)
for L in en ko ja zh; do
  printf "%s weight=" "$L"; grep -m1 "^weight:" "docs-site/content/$L/utility-commands/moai-e2e.md"
  printf "%s h2=" "$L"; grep -c "^## " "docs-site/content/$L/utility-commands/moai-e2e.md"
done

# CMD-E — menu registration (expect: sub-entry with ko/en/ja/zh keys + ref /utility-commands/moai-e2e)
grep -n -B 6 "ref: /utility-commands/moai-e2e" docs-site/data/menu/main.yaml

# CMD-E2 — menu.html untouched (expect: 0 lines in diff range for menu.html)
git diff --stat <run-base-sha>..HEAD -- docs-site/layouts/partials/menu.html | wc -l

# CMD-E6 — pathspec-commit + scope audit (expect: every commit touches only docs-site/ or .moai/specs/SPEC-DOCSITE-E2E-001/)
git log --format='%h' <run-base-sha>..HEAD | while read -r c; do
  echo "--- $c ---"; git show --stat --format='' "$c" | head -30
done

# ---- Hygiene (AC-DSE-008/009) ----

# CMD-H — blacklist / Mermaid direction / internal tokens on the new pages (expect: 0 / 0 / 0)
grep -rnE "docs\.moai-ai\.dev|adk\.moai\.com|adk\.moai\.kr" docs-site/content/*/utility-commands/moai-e2e.md | wc -l
grep -rnE "flowchart (LR|RL)|graph (LR|RL)" docs-site/content/*/utility-commands/moai-e2e.md | wc -l
grep -rnE "REQ-[A-Z]+-[0-9]{3}|SPEC-[A-Z][A-Z0-9-]+-[0-9]{3}" docs-site/content/*/utility-commands/moai-e2e.md | wc -l

# CMD-H2 — dark-theme guard on the whole diff (expect: 0)
git diff <run-base-sha>..HEAD -- docs-site/ | grep -c 'data-theme="dark"'

# ---- Build gates (AC-DSE-016/017/019) ----

# CMD-F — hugo build + sitemap (expect: exit 0; zero NEW warnings vs pre-flight baseline; sitemap exists)
#         Evidence persists under .moai/state/verify/ (survives /tmp clearance — audit D7)
mkdir -p .moai/state/verify/SPEC-DOCSITE-E2E-001
(cd docs-site && hugo --minify) > .moai/state/verify/SPEC-DOCSITE-E2E-001/hugo.log 2>&1; echo "exit=$?"; tail -20 .moai/state/verify/SPEC-DOCSITE-E2E-001/hugo.log
test -f docs-site/public/sitemap.xml && echo "sitemap OK"

# CMD-F2 — auxiliary i18n check (REQ-DSE-204; PASS or classified pre-existing baseline)
mkdir -p .moai/state/verify/SPEC-DOCSITE-E2E-001
bash scripts/docs-i18n-check.sh > .moai/state/verify/SPEC-DOCSITE-E2E-001/i18n.log 2>&1; echo "exit=$?"; tail -20 .moai/state/verify/SPEC-DOCSITE-E2E-001/i18n.log

# CMD-G — built-site nav reachability (expect: >=1 hit per locale tree)
for L in en ko ja zh; do printf "%s: " "$L"; grep -rl "utility-commands/moai-e2e" "docs-site/public/$L" 2>/dev/null | wc -l; done

# ---- SPEC lint gate ----

# CMD-I — spec lint (expect: 0 errors; StatusGitConsistency WARNING structural until sync close — no lint.skip)
moai spec lint .moai/specs/SPEC-DOCSITE-E2E-001/spec.md
```

> `<run-base-sha>` = the HEAD SHA captured at run-phase entry (P2 pre-flight); record it in progress.md §E.2.

## Quality Gates

| Gate | Criterion |
|------|-----------|
| Build | CMD-F exit 0, zero NEW warnings, sitemap exists |
| Parity | CMD-D / CMD-D3 / CMD-G all pass for 4 locales |
| Invariance | CMD-A = 0; CMD-A2/B/B2 residuals ⊆ documented false-positive list |
| Reachability | CMD-C ≥1 per applicable surface; CMD-E menu entry complete |
| Scope | CMD-E6 all commits pathspec-scoped, docs-site + SPEC dir only |
| Lint | CMD-I 0 errors |

## Definition of Done

- [ ] AC-DSE-001..020 all PASS with verbatim CMD outputs in the E1 matrix
- [ ] progress.md §E.2 carries: run-base SHA, reconciled ring inventory, false-positive list, hugo baseline
- [ ] All commits pushed; 4-locale set landed together (no partial-locale push)
- [ ] No writes outside docs-site/ + this SPEC directory
- [ ] Vercel production deploy of adk.mo.ai.kr picks up the change (post-push observation; Vercel binding untouched)
