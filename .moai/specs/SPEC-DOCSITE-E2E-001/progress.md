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

_<populated per milestone below>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Inputs: tier=M, scope ~45 files (docs-site 4-locale pages + count sweep), domains=1 (docs-site), language mix=markdown/yaml/html, concurrency benefit=LOW (sequential authoring chain ko→en→ja/zh with shared menu/nav files)
- Mode evaluation: trivial NO (multi-file) / background NO (write work) / agent-team RETIRED / parallel NO (single domain, write-conflict on shared nav files) / workflow NO (not uniform-mechanical: authoring + translation are semantic) / sub-agent YES
- Decision: sub-agent
- Justification: docs authoring is coding-heavy-equivalent sequential work (Anthropic coding-task parallelism caveat); shared files (main.yaml, _meta.yaml, CHANGELOG) forbid parallel writers. Single manager-develop spawn executes M1-M5 per plan.md with hns-oss-docs skills injected.
