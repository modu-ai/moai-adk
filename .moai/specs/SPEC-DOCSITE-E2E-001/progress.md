# SPEC-DOCSITE-E2E-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-07-13
plan_status: audit-ready
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
req_count: 21
ac_count: 20
plan_revision: "0.1.1 (plan-audit iter-1 FAIL 0.84, D1-D7 defect-fix)"
baseline_note: "Ring baselines measured 2026-07-13 plan-phase (A: 12/10 files; A': 14; B: ko19/ja19/zh21; B' widened D1 any-gap pattern: ko9/ja9/zh9) — run-phase MUST re-measure per REQ-DSE-104"
```

## §E.2 Run-phase Evidence

### Pre-flight record (REQ-DSE-104 / AC-DSE-013 / AC-DSE-018 — executed BEFORE first edit, 2026-07-13)

- **run-base SHA**: `23d979cc6d1761116ee1a03a4bd0eb05427df9d8` (branch `main`, = plan-artifacts commit)
- **P1 working-tree cleanliness (REQ-DSE-203)**: `git status --porcelain docs-site/` → **0 lines (clean)**. The plan-time 2-file residual (`content/{ja,zh}/advanced/settings-json.md`) was resolved by commit `57f49c47f` before run entry. PASS.
- **P2 pre-spawn sync**: `git fetch origin main` OK; `git rev-list --count --left-right origin/main...HEAD` → `0 1` (local ahead by 1 = plan commit; proceed). `moai session list --json --filter-spec=SPEC-DOCSITE-E2E-001` → `[]` (no concurrent session).
- **P4 hugo baseline**: `(cd docs-site && hugo --minify)` → exit 0, **0 WARN/ERROR lines** (log: `.moai/state/verify/SPEC-DOCSITE-E2E-001/hugo-baseline.log`). Gate for CMD-F = zero NEW warnings vs this clean baseline.
- **P6 sibling weights (en == ko verified)**: moai 20 / moai-goal 25 / moai-loop 40 / moai-fix 50 / moai-codemaps 50 / moai-clean 60 / moai-feedback 80 → **new page weight = 90** (ordered last, identical across 4 locales).

### Re-measured ring baselines (CMD-PRE-1..4; plan-phase deltas: ALL ZERO)

| Ring | Re-measured (run) | Plan baseline | Delta |
|------|-------------------|---------------|-------|
| A — English ERE family | **12 matches / 10 files** | 12 / 10 | 0 |
| A′ — English loose forms | **14 matches** | 14 | 0 |
| B — locale 10-family | **ko 19 / ja 19 / zh 21** | ko 19 / ja 19 / zh 21 | 0 |
| B′ — locale 9-family (widened any-gap) | **ko 9 / ja 9 / zh 9** | ko 9 / ja 9 / zh 9 | 0 |

### Per-match classification (catalog-count vs documented false positive)

- **Ring A (12/12 catalog-count → update)**: ko/claude-code/agentic/sub-agents.md:17; en/advanced/builder-agents.md:177; en/advanced/claude-md-guide.md:103; en/claude-code/agentic/sub-agents.md:17,155; en/workflow-commands/moai-harness.md:301; en/multi-llm/model-policy.md:15; en/getting-started/faq.md:103; en/core-concepts/harness-engineering.md:27; en/core-concepts/what-is-moai-adk.md:43,257; layouts/index.html:22 (comment). 0 false positives.
- **Ring A′ (13 catalog-count → update; 1 documented FP)**: catalog = builder-agents.md:17; agent-guide.md:7,21 (evolution prose `22 → 17 → 8 → 10` → extend `→ 11`),36,38; model-policy.md:15; introduction.md:31,148,168,170; what-is-moai-adk.md:24,72,259. **FP (unchanged)**: `en/claude-code/agentic/agent-view.md:88` — "Running 10 agents in parallel" (generic parallelism prose, not a catalog claim).
- **Ring B ko (19/19 catalog-count → update)**: builder-agents.md:177; agent-guide.md:7,36,38; claude-md-guide.md:103; sub-agents.md:17,155; moai-harness.md:301; introduction.md:31,148,170; faq.md:103; harness-engineering.md:27; what-is-moai-adk.md:7,24,43,72,257,259. 0 FPs.
- **Ring B ja (19/19 catalog-count → update)**: same file/line structure as ko. 0 FPs.
- **Ring B zh (19 catalog-count + 1 evolution prose → update; 1 documented FP)**: same 19-file/line structure as ko + agent-guide.md:21 (`22 → 17 → 8 → **10**` evolution prose → extend `→ 11`). **FP (unchanged)**: `zh/claude-code/agentic/agent-view.md:88` — "并行跑 10 个智能体" (parallelism prose).
- **Ring B′ (9/9/9 all catalog-count → update, 0 FPs)**: per locale — agent-guide.md:38,111; claude-md-guide.md:103; sub-agents.md:17; introduction.md:148; faq.md:103; what-is-moai-adk.md:72,257,610 (directory-tree comment "9개 MoAI 커스텀" → "10개"; note zh:610 carries untranslated ko text — pre-existing untranslated backlog, count-only minimal edit per spec §E exclusion).

### Milestone evidence

| M | Commit | Scope | Evidence |
|---|--------|-------|----------|
| M1 | `8e550cc34` | ko canonical page (11 H2) + spec.md `draft → in-progress` + §E.2 pre-flight record | CMD-D2 ko: Playwright 6 / Maestro 4 / Appium 2 / Electron 3 / Tauri 3 |
| M2 | `9a995a5bb` | en/ja/zh derivation (3 files, +525) | CMD-D3: weight 90 ×4, H2 11 ×4; CMD-D2 en all ≥1; CMD-H 0/0/0 |
| M3 | `564551883` | main.yaml sub-entry (4-locale name map) + `_meta.yaml` ×4 | CMD-E: entry at main.yaml L260-265; meta count 1 ×4; menu.html untouched |
| M4 | `6496d69d4` | 41 files (40 md + index.html): count sweep + enumeration insertion | CMD-A 0; CMD-A2 = 1 documented FP only; CMD-B ko/ja 0, zh = 1 documented FP only; CMD-B2 0 |
| M5 | (this commit + backfill) | verification + §E.2/§E.3 close | full CMD block below |

### M4 surface-growth notes (beyond re-measured grep inventory)

Three grep-escaped catalog claims discovered during per-file structure reads were normalized in M4 (same logical surface; recorded per the surface-can-grow obligation):
1. `en/core-concepts/what-is-moai-adk.md:7` — "10 specialist AI agents" (15-char gap exceeds the A′ 12-char window)
2. `multi-llm/model-policy.md` ko/ja/zh catalog sentence — noun-precedes-number ordering escapes the B pattern
3. `advanced/builder-agents.md:17` ko/ja/zh — "카탈로그 (10개)가" noun-precedes-number ordering

Evolution prose extended in ALL 4 locales of agent-guide.md (ko `10**으로`/ja `10** へと` escaped the grep; updated for parity with the en/zh matched lines).

### Enumeration-surface classification (CMD-C applicability, per REQ-DSE-103)

| File family | Classification | e2e-specialist insertion |
|-------------|----------------|--------------------------|
| advanced/agent-guide.md ×4 | ENUMERATION (full catalog guide) | `### Specialist` section + table row + file-tree entry (grep-c = 2/locale) |
| claude-code/agentic/sub-agents.md ×4 | ENUMERATION (role prose L155) | named in role enumeration (1/locale) |
| getting-started/introduction.md ×4 | ENUMERATION (category table) | `Specialist | 1 | e2e-specialist` row (1/locale) |
| getting-started/faq.md ×4 | ENUMERATION (tier-assignment tables) | named in session-model prose grounded in `model: inherit` frontmatter (1/locale) |
| advanced/claude-md-guide.md ×4 | ENUMERATION (category table) | `Specialist (1)` row (1/locale) |
| core-concepts/what-is-moai-adk.md ×4 | ENUMERATION (table + mermaid) | table row + `Specialist` mermaid subgraph (2/locale) |
| advanced/builder-agents.md ×4 | COUNT-ONLY (catalog referenced in contrast to harness) | none — 0 is correct |
| core-concepts/harness-engineering.md ×4 | COUNT-ONLY (pillar-table cell) | none — 0 is correct |
| multi-llm/model-policy.md ×4 | COUNT-ONLY (catalog count sentence; assignment table is an explicit 7-agent policy subset, and e2e-specialist is `model: inherit`, not tier-assigned) | none — 0 is correct |
| workflow-commands/moai-harness.md ×4 | COUNT-ONLY (related-docs link text) | none — 0 is correct |
| layouts/index.html | COUNT-ONLY stat row (REQ-DSE-105) | none — grep-c = 0 is the CORRECT result per audit D6 |

### M5 verification results (evidence: `.moai/state/verify/SPEC-DOCSITE-E2E-001/`)

- **CMD-F**: exit 0, 0 WARN/ERROR lines (delta vs clean pre-flight baseline = 0); `public/sitemap.xml` exists. Log: `hugo.log`.
- **CMD-F2**: exit 1 — **16 errors, ALL "no H1 heading" class, ALL pre-existing baseline**: {advanced/decision-memory.md, advanced/harness-profiles.md, utility-commands/moai-goal.md, workflow-commands/moai-harness.md} × 4 locales. decision-memory/harness-profiles/moai-goal were never touched by this SPEC; moai-harness.md was touched ONLY at the L301 link text (H1 absence predates the edit). The 4 NEW moai-e2e.md pages produce ZERO script errors. Classified pre-existing baseline per plan.md B4 — not chased. Log: `i18n.log`.
- **CMD-G**: built-site nav reachability — 97 files per locale reference `utility-commands/moai-e2e` (menu renders site-wide). 4/4 locales ≥1.
- **CMD-E6**: 4 pathspec commits (`8e550cc34`, `9a995a5bb`, `564551883`, `6496d69d4`); per-commit name-only audit shows ZERO files outside `docs-site/` + this SPEC directory. Full run-range diff: 52 files (50 docs-site + spec.md + progress.md).
- **CMD-I**: `moai spec lint` → "No findings — all SPEC documents are valid" (0 errors).
- **CMD-H/H2**: new pages 0 blacklisted URLs / 0 Mermaid LR-RL / 0 internal SPEC-REQ tokens; whole-diff added-line Mermaid LR/RL guard 0; whole-diff `data-theme="dark"` guard 0.
- **Whole-tree 4-locale file-existence parity**: 0 MISSING.
- **Body-emoji scan (new pages)**: 0 hits.
- **Working-tree discipline**: residual dirt at run end = runtime-managed `.moai/{config,harness,state}` files only (untouched, uncommitted per B8/B10); `git status docs-site/` = clean.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-13
run_commit_sha: "pending-backfill-M5"
run_status: complete
ac_pass_count: 20
ac_fail_count: 0
ac_notes: "AC-DSE-019 PASS via the classified-pre-existing-baseline branch (16 no-H1 errors on untouched-structure files, evidence in i18n.log)"
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "done (P2: origin/main 0 behind, local +1 plan commit)"
l44_post_push_fetch: "n/a — push deferred to orchestrator per spawn contract"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  status: "n/a"
  reason: "docs-only SPEC — zero Go code; binding build gate is hugo --minify (exit 0, warning-free)"
total_run_phase_files: 52
m1_to_mN_commit_strategy: "per-milestone pathspec commits (M1 8e550cc34 / M2 9a995a5bb / M3 564551883 / M4 6496d69d4 / M5 evidence+backfill); git add -A never used; NO push — orchestrator pushes after independent verification"
```

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Inputs: tier=M, scope ~45 files (docs-site 4-locale pages + count sweep), domains=1 (docs-site), language mix=markdown/yaml/html, concurrency benefit=LOW (sequential authoring chain ko→en→ja/zh with shared menu/nav files)
- Mode evaluation: trivial NO (multi-file) / background NO (write work) / agent-team RETIRED / parallel NO (single domain, write-conflict on shared nav files) / workflow NO (not uniform-mechanical: authoring + translation are semantic) / sub-agent YES
- Decision: sub-agent
- Justification: docs authoring is coding-heavy-equivalent sequential work (Anthropic coding-task parallelism caveat); shared files (main.yaml, _meta.yaml, CHANGELOG) forbid parallel writers. Single manager-develop spawn executes M1-M5 per plan.md with hns-oss-docs skills injected.
