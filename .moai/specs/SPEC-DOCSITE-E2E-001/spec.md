---
id: SPEC-DOCSITE-E2E-001
title: "docs-site 4-locale /moai e2e documentation + 11-agent catalog count-literal normalization (deferred follow-up of SPEC-E2E-REVIVAL-001)"
version: "0.1.1"
status: draft
created: 2026-07-13
updated: 2026-07-13
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "docs-site (content/{en,ko,ja,zh} + data/menu/main.yaml + layouts/index.html)"
lifecycle: spec-anchored
tags: "docs-site, i18n, e2e, documentation, count-literal, menu, hugo, 4-locale"
era: V3R6
tier: M
depends_on: [SPEC-E2E-REVIVAL-001]
---

# SPEC-DOCSITE-E2E-001 — docs-site 4-locale /moai e2e documentation + 11-agent count-literal normalization

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-13 | manager-spec | Plan-phase artifact set authored (Tier M, 3 artifacts + progress.md). Deferral provenance: SPEC-E2E-REVIVAL-001 spec.md §E Exclusions "Out of Scope — docs-site 4-locale documentation and count-literal corrections" (user decision 2026-07-13). Plan-phase re-measurement (this tree, 2026-07-13) confirmed the deferred English-literal inventory (10 files / 12 matches) AND discovered a substantially larger locale-language ring not in the deferred inventory: ko 19 / ja 19 / zh 21 ten-family matches + 25 nine-family matches (see §A baselines). |
| 0.1.1 | 2026-07-13 | manager-spec | Defect-fix pass (plan-audit iter-1 FAIL 0.84, D1-D7). D1 (BLOCKING): Ring B′ 9-family pattern widened to the digit-boundary any-gap form `(^\|[^0-9])9[^0-9]{0,8}(MoAI\|커스텀\|사용자 정의\|カスタム\|自定义)` — the prior counter-form (`9(개\|個\|个)? ?…`) missed spaced ja/zh forms ("9 個の MoAI", "9 个 MoAI 自定义"; verified escapees: ja/advanced/agent-guide.md:111, ja/core-concepts/what-is-moai-adk.md, zh/advanced/agent-guide.md:111). Re-measured: ko 9 / ja 9 / zh 9 (27 combined); non-self-tripping on post-change 10/11 forms (they contain no digit 9). D2: §A.1 B′ row re-attributed with exact command + scope — the v0.1.0 "25" figure came from a DIFFERENT probe regex (with a `个` noun alternation) than the pinned CMD-PRE-4 (which yields 15 on ko/ja/zh); both superseded by the widened per-locale measurement. D3: REQ count corrected 20 → 21 (Group A 9 + B 6 + C 6). D4: working-tree residual re-measured — 2 files (ja+zh advanced/settings-json.md). D5: acceptance.md CMD-C rewritten to iterate the explicit REQ-DSE-103 file list (never derived from post-change 11-tokens — circular). D6: CMD-C2 expectation stated (e2e-specialist grep = 0 on the count-only stat row; index.html excluded from the REQ-DSE-103 enumeration list, owned by REQ-DSE-105). D7: CMD-F/F2 evidence redirects moved to `.moai/state/verify/SPEC-DOCSITE-E2E-001/` with `mkdir -p`. |

---

## §A Context & Problem

SPEC-E2E-REVIVAL-001 (status: completed, sync push `4cef092b7`, 2026-07-13) revived the `/moai e2e` subcommand as a multi-platform, token-minimized E2E testing subsystem and added the 11th retained agent `e2e-specialist` (catalog: 11 retained = 10 MoAI-custom + 1 Anthropic built-in `Explore`). Its spec.md §E Exclusions EXPLICITLY deferred two docs-site items to this follow-up:

1. **adk.mo.ai.kr user documentation** for the revived `/moai e2e` workflow, 4-locale (en/ko/ja/zh), including menu registration.
2. **Count-literal corrections** across the docs-site inventory carrying the stale "10 retained agents | 9 MoAI-custom | 10-agent" family, to be normalized to the 11/10 catalog under the 4-locale parity obligation.

This SPEC is docs-only: zero Go code, zero template-tree (`internal/template/`) writes, zero doctrine-tree (`.claude/`) writes. The doctrine/template/README/Go-display count rings were already closed by SPEC-E2E-REVIVAL-001 rings 1-4; this SPEC owns exclusively the deferred docs-site ring.

### A.1 Measured baselines (verified 2026-07-13, this tree, plan-phase — run-phase MUST re-measure per REQ-DSE-104)

| Ring | Grep family | Baseline | Files |
|------|-------------|----------|-------|
| A — English ERE family | `10 retained agents\|9 MoAI-custom\|10-agent` over docs-site/ (md + html) | **12 matches** | **10 files** (identical to the SPEC-E2E-REVIVAL-001 §E deferred inventory): content/en/{advanced/builder-agents.md, advanced/claude-md-guide.md, claude-code/agentic/sub-agents.md, core-concepts/harness-engineering.md, core-concepts/what-is-moai-adk.md, getting-started/faq.md, multi-llm/model-policy.md, workflow-commands/moai-harness.md}; content/ko/claude-code/agentic/sub-agents.md; layouts/index.html |
| A′ — English loose catalog forms | case-insensitive digit-boundary `10 …agent` minus ring-A overlap | **14 matches** (incl. ≥1 documented false positive — see REQ-DSE-101) | e.g. advanced/agent-guide.md ("catalog of 10 core agents", "22 → 17 → 8 → **10**"), builder-agents.md, multi-llm/model-policy.md, getting-started/introduction.md, claude-code/agentic/agent-view.md (false positive) |
| B — locale-language 10-family | digit-boundary `10(개\|個\|个)? … (에이전트\|エージェント\|智能体\|代理)` per locale | **ko 19 / ja 19 / zh 21 matches** | ko 9 / ja 9 / zh 10 files (agent-guide.md, sub-agents.md, introduction.md, faq.md, builder-agents.md, claude-md-guide.md, moai-harness.md 등) |
| B′ — locale-language 9-family | digit-boundary any-gap `(^\|[^0-9])9[^0-9]{0,8}(MoAI\|커스텀\|사용자 정의\|カスタム\|自定义)`, run per locale over `docs-site/content/{ko,ja,zh}` (CMD-PRE-4) | **ko 9 / ja 9 / zh 9 matches (27 combined)** | ko/ja/zh trees — includes the spaced ja/zh forms ("9 個の MoAI", "9 个 MoAI 自定义") a counter-only regex misses (audit D1). The v0.1.0 "25" figure was pattern/scope-misattributed (a different probe regex; the previously pinned CMD-PRE-4 counter-form yields 15 on the same scope) — superseded by this row (audit D2). |

> **Discovery note**: rings B/B′ were NOT in the deferred 10-file inventory — the SPEC-E2E-REVIVAL-001 inventory grep used English literals only. The mission's "re-measure, the surface can grow" obligation surfaced them at plan-phase. They are the same logical surface (stale catalog-count claims) and are in scope.

### A.2 Structure decisions (measured against the live docs-site tree)

| Decision | Value | Evidence |
|----------|-------|----------|
| New page section | `utility-commands` (NOT workflow-commands) | workflow-commands = plan/run/sync/project/harness (lifecycle pipeline); utility-commands = moai/goal/loop/fix/clean/codemaps/feedback (tool-class subcommands). `/moai e2e` is a tool-class subcommand sibling of fix/loop. |
| New page path/URL | `content/{locale}/utility-commands/moai-e2e.md` → `/utility-commands/moai-e2e` | Sibling convention (`moai-fix.md` → `/utility-commands/moai-fix`) |
| Page frontmatter convention | `title: /moai e2e` + shared `weight` + `draft: false` | Measured from sibling `moai-fix.md` (en/ko identical fields) |
| Menu registration | sub-entry under the existing utility-commands section (`ref: /utility-commands`, `icon: build`, main.yaml L214-259) with a 4-locale name map | Sibling sub-entries carry name map (ko/en/ja/zh) + ref only — **no icon on sub-entries** |
| menu.html SVG case | **NOT needed** — no new section icon value is introduced (page rides the existing section) | REQ-DSE-007 conditional guard retained |
| `_meta.yaml` | per-locale `"moai-e2e": title: "/moai e2e"` entry, 4 files | All 4 locale `_meta.yaml` verified structurally identical |
| Canonical authoring locale | **ko** (chain ko → en → ja/zh) | `.moai/docs/docs-site-i18n-rules.md` §17.3 [HARD] "Canonical source는 ko" — the mission brief's "en/ko/ja/zh" is a locale SET, not a chain order |
| `gen_menu.py` | does NOT exist — main.yaml header comment "Auto-generated by scripts/gen_menu.py" is stale; **manual edit** is the actual path | `find`/`ls` 0 hits (2026-07-13) |
| `scripts/docs-i18n-check.sh` | EXISTS (8,066 bytes) but behavior unverified at plan | REQ-DSE-204 capability gate; binding gates remain hugo build + explicit greps |
| hugo toolchain | hugo v0.160.1+extended available at `/opt/homebrew/bin/hugo` | version probe 2026-07-13 |

### A.3 Working-tree state (plan-time snapshot)

The 26-file docs-site menu-restructure residue named in the delegation brief was committed by a parallel session BEFORE plan authoring (commits `e7e8115a2` restructure + `1fe9d5a2b` Context7 follow-up, 2026-07-13). Plan-time residual (re-measured at v0.1.1, audit D4): **2 uncommitted files** (`docs-site/content/ja/advanced/settings-json.md`, `docs-site/content/zh/advanced/settings-json.md`). The run-phase entry precondition (REQ-DSE-203) still binds: re-verify cleanliness at run-phase entry — the shared checkout hosts parallel sessions and the state moves.

---

## §B Requirements (GEARS)

### Group A — New `/moai e2e` documentation page (4-locale)

- **REQ-DSE-001** (Ubiquitous): The docs-site shall gain a `/moai e2e` user documentation page at `docs-site/content/{en,ko,ja,zh}/utility-commands/moai-e2e.md` (4 new files, URL `/utility-commands/moai-e2e`), placed as a sibling of `/moai fix` and `/moai loop`.
- **REQ-DSE-002** (Ubiquitous): The page shall document, in user-facing language: (a) the platform-toolchain matrix — web → Playwright CLI (default), mobile → Maestro (default) with Appium fallback, desktop → Playwright-Electron for Electron apps / WebdriverIO + `@wdio/tauri-service` for Tauri apps; (b) project-type auto-detection from project markers; (c) CLI-first token-minimized execution (CLI output parsing preferred over MCP round-trips, bounded output tails, artifact file redirects with citable paths); (d) the graceful branch — "no e2e target detected" reporting and the desktop-native (non-Electron/non-Tauri) deferral notice. The content source of truth is the LIVE shipped workflow (`.claude/skills/moai/workflows/e2e.md` + `.claude/agents/moai/e2e-specialist.md` + `.claude/commands/moai/e2e.md`), re-read at run-phase — not this SPEC's prose.
- **REQ-DSE-003** (Ubiquitous): The ko locale shall be the canonical authoring source (translation chain ko → en → ja/zh per `.moai/docs/docs-site-i18n-rules.md` §17.3), and all 4 locale files shall land within the same run-phase commit set (same-PR / same-push obligation, §17.3 [HARD]).
- **REQ-DSE-004** (Ubiquitous): Each locale file shall carry the sibling frontmatter convention (`title: /moai e2e`, `draft: false`, and a `weight` value identical across all 4 locales, ordered after the existing utility-commands siblings — the concrete weight is derived at run-phase from the measured sibling weights), and the H2 section structure shall be parity-equal across locales.
- **REQ-DSE-005** (Ubiquitous): The `utility-commands/_meta.yaml` file in each of the 4 locales shall gain a `"moai-e2e"` entry with `title: "/moai e2e"`.
- **REQ-DSE-006** (Ubiquitous): `docs-site/data/menu/main.yaml` shall gain a sub-entry under the existing utility-commands section (`ref: /utility-commands`, `icon: build`) carrying the 4-locale name map (ko/en/ja/zh each `/moai e2e`) and `ref: /utility-commands/moai-e2e`.
- **REQ-DSE-007** (Capability gate): **Where** a NEW section-level icon value were introduced into `data/menu/main.yaml` (not expected — the page rides the existing utility-commands section and sub-entries carry no icon), `layouts/partials/menu.html` shall gain the matching SVG case in the same commit; absent a new icon value, `menu.html` shall remain untouched.
- **REQ-DSE-008** (Unwanted): The new pages shall not contain decorative body emoji (typographic symbols `→ ← ↓ ✓ ✗` and orchestrator-banner emoji inside example code blocks are exempt per the §17.1 conventions), shall not use Mermaid `LR`/`RL` direction (TD/TB only), shall not introduce blacklisted URLs (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr`), and shall not reference internal SPEC IDs, REQ/AC tokens, or commit SHAs.
- **REQ-DSE-009** (Ubiquitous): The pages shall follow the emphasis-marker spacing rule (`**강조** (괄호)`, never `**강조(괄호)**`) and the light-theme-only rule (no `[data-theme="dark"]` CSS/markup additions).

### Group B — Count-literal normalization (10 → 11 retained / 9 → 10 MoAI-custom)

- **REQ-DSE-100** (Ubiquitous): Ring A — every match of the English ERE family `10 retained agents|9 MoAI-custom|10-agent` inside `docs-site/` (measured baseline: 10 files / 12 matches, §A.1) shall be updated to the 11-catalog form ("11 retained agents", "10 MoAI-custom", "11-agent"); the post-change invariance grep (CMD-A, acceptance.md § Executable Command Block) shall return 0 matches.
- **REQ-DSE-101** (Ubiquitous): Ring A′ — English loose catalog-count claims ("10 agents", "10 core agents", "catalog of 10", "**10**" as catalog count; measured baseline 14 matches) shall be updated where the match asserts the CATALOG count, including extending the catalog-evolution prose `22 → 17 → 8 → 10` (advanced/agent-guide.md) to `… → 11`. Matches that are NOT catalog-count claims are documented false positives and shall remain unchanged — the one confirmed at plan-phase is `content/en/claude-code/agentic/agent-view.md` ("Running 10 agents in parallel", generic parallelism prose); run-phase re-measurement (REQ-DSE-104) shall re-classify every Ring A′ match individually before editing.
- **REQ-DSE-102** (Ubiquitous): Ring B — locale-language catalog-count claims in ko/ja/zh (10-family digit-boundary baseline: ko 19 / ja 19 / zh 21 matches; 9-family widened any-gap baseline: ko 9 / ja 9 / zh 9 matches, §A.1) shall be updated to the locale-language 11/10 forms (ko `11개 … 에이전트 (10 MoAI 커스텀 + 1 …)`, ja `11 個のエージェント (10 個の MoAI カスタム + 1 …)`, zh `11 个智能体 (10 个 MoAI 自定义 + 1 …)`), verified per-locale with the digit-boundary-safe greps (CMD-B family). Non-catalog-count matches follow the same documented-false-positive discipline as REQ-DSE-101.
- **REQ-DSE-103** (Ubiquitous): Reachability over token presence — every agent-catalog ENUMERATION surface among the matched files (all 4 locales of advanced/agent-guide.md and claude-code/agentic/sub-agents.md; getting-started/{introduction,faq}.md; advanced/{builder-agents,claude-md-guide}.md; core-concepts/{harness-engineering,what-is-moai-adk}.md; multi-llm/model-policy.md; workflow-commands/moai-harness.md — per-file applicability re-derived at run-phase; `layouts/index.html` is EXCLUDED from this list: it is a count-only stat row owned by REQ-DSE-105 and the enumeration obligation does not apply to it, audit D6) shall NAME `e2e-specialist` wherever the catalog members are enumerated (table row, role enumeration prose, or list). A count bump without the 11th agent named does NOT satisfy this requirement.
- **REQ-DSE-104** (Event-driven): **When** run-phase pre-flight begins, the executor shall re-measure ALL ring baselines with the pinned CMD-PRE commands (acceptance.md § Executable Command Block) — the surface can grow or shrink between plan and run — and shall record the reconciled file inventory in `progress.md` §E.2 BEFORE the first edit.
- **REQ-DSE-105** (Ubiquitous): The `docs-site/layouts/index.html` landing stat row (the template whose comment mandates 실측 stat values — "placeholder 수치 금지") shall render the 11-agent count; the guidance comment itself is updated to the 11/10 wording, not removed.

### Group C — Verification & publication discipline

- **REQ-DSE-200** (Ubiquitous): `hugo --minify` (executed in `docs-site/`) shall complete with zero warnings, and `docs-site/public/sitemap.xml` shall exist after the build.
- **REQ-DSE-201** (Ubiquitous): 4-locale parity shall hold: the `moai-e2e.md` page exists in all 4 locale trees; the `_meta.yaml` entry exists in all 4; the menu sub-entry name map carries all 4 locale keys; per-locale H2 section counts of the new page match across locales.
- **REQ-DSE-202** (Ubiquitous): Every run-phase commit shall be a pathspec commit (`git commit -- <explicit paths>`); `git add -A` / `git add .` shall not be used (shared-checkout parallel-session race defense).
- **REQ-DSE-203** (Event-driven): **When** run-phase entry is attempted, the docs-site working tree shall be clean of unrelated uncommitted modifications (`git status --porcelain docs-site/` empty, or every residual line explicitly resolved by user decision). Plan-time residual (v0.1.1 re-measure): 2 files (`content/ja/advanced/settings-json.md`, `content/zh/advanced/settings-json.md` — parallel-session residue; the 26-file menu-restructure set was already committed at `e7e8115a2` / `1fe9d5a2b` before plan authoring).
- **REQ-DSE-204** (Capability gate): **Where** `scripts/docs-i18n-check.sh` executes successfully in this tree (existence verified at plan-phase; runtime behavior unverified), it shall pass after the changes; the BINDING gates remain REQ-DSE-200/201 (hugo build + explicit parity greps) regardless of the script's availability.
- **REQ-DSE-205** (Unwanted): The run shall not modify `README*`, `CLAUDE.md`, `.claude/**`, `internal/**`, or any tree outside `docs-site/` + `.moai/specs/SPEC-DOCSITE-E2E-001/` — the doctrine/template/README/Go-display count rings were closed by SPEC-E2E-REVIVAL-001 and re-opening them here is a scope violation.

---

## §C Acceptance Criteria

The canonical AC matrix (AC-DSE-001..020), Given-When-Then scenarios, edge cases, and the Executable Command Block (CMD-PRE-1..4, CMD-A..CMD-I) live in `acceptance.md` (SSOT). Summary counts: 21 REQs (Group A 9 + Group B 6 + Group C 6) / 20 ACs.

---

## §D Constraints

- Docs-only SPEC: write scope = `docs-site/**` + `.moai/specs/SPEC-DOCSITE-E2E-001/**`. No Go, no template tree, no doctrine tree (REQ-DSE-205).
- 4-locale same-commit-set obligation (REQ-DSE-003) — no ko-only or en-only partial landing.
- Pathspec commits only (REQ-DSE-202).
- All hygiene rules of `.moai/docs/docs-site-i18n-rules.md` §17.1-§17.3 and CLAUDE.local.md §17.1 (icon shortcodes, light-theme-only, Mermaid TD) bind the new content.

---

## §E Exclusions

The exclusions below are out of scope for this SPEC.

### Out of Scope — Non-docs-site count-literal surfaces

- `CLAUDE.md`, `.claude/rules/**`, `.claude/agents/**`, `internal/template/**`, `internal/web/**`, `README*` count literals — closed by SPEC-E2E-REVIVAL-001 rings 1-4. This SPEC touches docs-site only.

### Out of Scope — Desktop-native automation documentation beyond the deferral notice

- The new page documents the graceful deferral branch ONLY. Authoring docs for a native-desktop (non-Electron/non-Tauri) automation capability is deferred with the capability itself (SPEC-E2E-REVIVAL-001 §E).

### Out of Scope — Dogfood e2e suite and CI integration docs

- No e2e suite is authored for moai-adk-go or the docs-site itself, and no CI-integration guide is written (mirrors the SPEC-E2E-REVIVAL-001 exclusions).

### Out of Scope — docs-site version snapshot

- No `content/{locale}/v{X}/` frozen snapshot is created — this is not a Major/Minor release event (§17.4 of the i18n rules).

### Out of Scope — Menu icon system changes

- No new icon value, no `layouts/partials/menu.html` SVG case, no `moai-brand.css` (FROZEN) or `moai-design.css` change. REQ-DSE-007 exists solely as a conditional guard.

### Out of Scope — Unrelated untranslated-content backlog

- The pre-existing docs-site untranslated backlog (ko-only prose inside en/ja/zh files elsewhere in the tree) is a separate known debt and is not swept here. Only the surfaces this SPEC touches are held to parity.

### Out of Scope — moai.md unified-command overview restructuring

- `content/{locale}/utility-commands/moai.md` (the `/moai` overview page) gains at most a subcommand-list mention of `e2e` IF its subcommand enumeration is one of the re-measured count/enumeration surfaces; a full restructure of that page is excluded.

---

## §F Traceability

| Group | REQs | ACs (acceptance.md) |
|-------|------|---------------------|
| A — new page 4-locale | REQ-DSE-001..009 | AC-DSE-001..009 |
| B — count normalization | REQ-DSE-100..105 | AC-DSE-010..015 |
| C — verification/publication | REQ-DSE-200..205 | AC-DSE-016..020 |

Provenance: SPEC-E2E-REVIVAL-001 spec.md §E Exclusions ("Out of Scope — docs-site 4-locale documentation and count-literal corrections", user decision 2026-07-13). Menu-structure SSOT for run-phase execution: `hns-oss-docs-structure-map` skill + measured evidence in §A.2.
