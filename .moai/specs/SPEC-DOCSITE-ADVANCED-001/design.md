# design.md — SPEC-DOCSITE-ADVANCED-001

> Page architecture, shared narrative spine, and per-page content outlines
> for the 6 new Advanced pages. This is the **HOW** layer (spec.md is WHAT/WHY).
> Run-phase manager-develop + the oss-docs harness specialists (content-author
> + locale-translator) consume this file as the structural contract.

## §A Page architecture

### A.1 The 3-pillar narrative spine (shared)

All 6 pages participate in the v3.0 product narrative established by `.moai/reports/readme-docs-redesign-20260713.md` and `project_readme_v3_rc11_redesign_draft`:

```
v3.0 MoAI-ADK product differentiation
├── Pillar 1 — 🪙 Tokenomics (측정하고, 라우팅하고, 다이어트하고, 방어한다)
│   ├── tokenomics-overview   (pillar entry — the 4-layer doctrine)
│   └── token-budget           (layer D — budget defense deep-dive)
├── Pillar 2 — 🔁 Agentic Loop (언제 멈추고 언제 계속할 것인가)
│   └── autonomous-loops       (the /goal · /moai goal · /moai loop primitives)
└── Pillar 3 — 🤖 Agentic Harness (어떤 에이전트가, 어떤 티어로, 어떻게 진화하는가)
    ├── no-haiku-3tier         (architectural foundation — Sonnet/Opus/Fable)
    ├── plan-type-profiles     (the 60-cell profile — plan_type × tier × agent)
    └── self-evolving          (the ACE 3-Loop — self-improvement)
```

**Cross-page navigation contract**: every page carries a top-of-page "3-pillar map" callout — a small Mermaid TD diagram or icon-shortcode row showing the pillar structure with the current page highlighted. This visual rhyme aids reader orientation across the 6-page cluster.

### A.2 Inter-page references

| Page | References OUT to | Cited BY |
|------|-------------------|----------|
| tokenomics-overview | token-budget, plan-type-profiles | (all 5 others — it is the pillar entry) |
| token-budget | tokenomics-overview | tokenomics-overview |
| no-haiku-3tier | plan-type-profiles | plan-type-profiles, self-evolving |
| plan-type-profiles | no-haiku-3tier | tokenomics-overview, no-haiku-3tier |
| self-evolving | (none of the other 5 directly; cites EVOLVE SPECs) | tokenomics-overview (as pillar member) |
| autonomous-loops | token-budget (the `/clear` threshold hooks into budget defense) | tokenomics-overview |

### A.3 Design system binding (Claude Warm Editorial)

Per CLAUDE.local.md §17.1:

- **Light single theme** — no `[data-theme="dark"]` branching in any new content.
- **Icon shortcodes** (`{{</* icon name [variant] */>}}`) for semantic markers; the icon palette is fixed at: check, check-circle, x, x-circle, warning, info, bulb, rocket, star, flash, sparkles, target, package, book, search, wrench, database, rotate, clock, arrow-right.
- **Code blocks** render via the existing `render-codeblock.html` hook (macOS dark card) — no new render hook.
- **Mermaid** TD-only via existing `foot.html` CDN UMD — `theme: 'base'` with coral themeVariables.
- **Typography arrows** (`→ ← ↓ ✓ ✗`) preserved verbatim — they are NOT body emoji.
- **MoAI banner emoji** (`🤖 🗿 📋 🎯`) preserved when reproducing orchestrator output inside fenced code blocks.

---

## §B 3-pillar narrative (the shared "why")

The narrative is sourced from `.moai/reports/readme-docs-redesign-20260713.md` + `project_v3_tokenomics_docs_plan` memory + `reference_lilian_weng_harness_selfimprove` (Weng's "Harness Engineering for Self-Improvement", 2026-07-04) + `reference_tokenomics_press_2026_07` (market context). The spine (quoted from `project_v3_tokenomics_docs_plan`):

> "최소 토큰 × 최고 품질 × 최대 가용성" — minimum tokens × maximum quality × maximum availability. Token 정액제→종량제 전환 + 구독 주간 쿼터 시대에 하네스 간 경쟁 축은 토큰 × 품질. MoAI-ADK는 계측→라우팅→검증경제→예산방어 4층 + 모델 티어 라우팅으로 이를 구조화한 유일한 하네스라는 포지셔닝.

The three pillars decompose this spine:

| Pillar | Korean tagline | English tagline | Pages |
|--------|----------------|-----------------|-------|
| 1 Tokenomics | 측정하고, 라우팅하고, 다이어트하고, 방어한다 | Meter, route, diet, defend | tokenomics-overview, token-budget |
| 2 Agentic Loop | 언제 멈추고 언제 계속할 것인가 | When to stop, when to continue | autonomous-loops |
| 3 Agentic Harness | 어떤 에이전트가, 어떤 티어로, 어떻게 진화하는가 | Which agent, at which tier, how it evolves | no-haiku-3tier, plan-type-profiles, self-evolving |

Each page opens with a 1-paragraph hook tying it to this spine.

---

## §C Page-by-page outline (canonical KO structure)

### §C.1 tokenomics-overview

**Slug**: `tokenomics-overview`
**KO title**: `토크노믹스 개요`
**Pillar**: 1 (entry page)

**Outline** (H2 sections):
1. **도입: 왜 토크노믹스인가** — market context hook (`reference_tokenomics_press_2026_07` — 토큰맥싱 후퇴, 98% 토큰가 하락 vs 320% 비용 상승, Fable5 단가)
2. **3-기둥 서사** — the v3.0 product narrative (cross-page map; icon-shortcode row)
3. **4-층 토크노믹스 구조** — Mermaid TD diagram of 계층 A/B/C/D
   - A 계층 (Metering) — per-SPEC 토큰 회계 (SPEC-TOKEN-ACCOUNTING-001, `moai spec audit` 컬럼)
   - B 계층 (Routing) — Tier×Phase 선언적 model/effort (SPEC-TOKEN-ROUTING-001)
   - C 계층 (Verify-diet economy) — verbatim 증거는 파일 리다이렉트, 컨텍스트엔 exit code+tail
   - D 계층 (Budget defense) — 90% hard-limit graceful stop (link to token-budget)
4. **모델 티어 라우팅** — link to plan-type-profiles (the 60-cell profile that operationalizes the B layer)
5. **CG 모드 (비용 최적화)** — brief mention; CG = Claude + GLM 60-70% 절감, cross-reference multi-llm/cg-mode
6. **검증된 사실과 로드맵** — honesty split: what ships today vs what is design-stage

**Source**: `.moai/reports/readme-docs-redesign-20260713.md` + `project_v3_tokenomics_docs_plan` memory.

### §C.2 token-budget

**Slug**: `token-budget`
**KO title**: `토큰 예산 관리와 우아한 중단`
**Pillar**: 1 (layer D deep-dive)

**Outline**:
1. **도입: 예산 방어의 필요성** — SSE stall (`stream_idle_partial`) near context ceiling
2. **모델별 컨텍스트 임계치** — table (Opus 4.8 1M = 50%, GLM-5.2 1M = 50%, Opus/Fable 256K = 90%, Sonnet/Opus 200K = 90%, Haiku 200K = 90%)
3. **2-단계 핸드오프 마커** — soft `(⚠️/clear)` vs hard `(🛑/clear!)` (cite statusline system)
4. **우아한 중단 절차 (SPEC-TOKEN-BUDGET-STOP-001)** — 90% hard-limit + paste-ready handoff emission + `.moai/state/verify/` 영속화
5. **paste-ready resume 6-블록 구조** — brief structural overview (full detail is in the session-handoff rule)
6. **검증-다이어트 (verify-diet)** — file-redirect contract (verbatim evidence on disk, context carries exit code + bounded tail ≤50 lines)
7. **검증 증거 영속화 의무** — `.moai/state/verify/<session>/` (survives `/tmp` clearance)

**Source**: `.moai/specs/SPEC-TOKEN-BUDGET-STOP-001/` + `.claude/rules/moai/workflow/{context-window-management,session-handoff}.md` + `agent-common-protocol.md` § File-redirect contract.

### §C.3 no-haiku-3tier

**Slug**: `no-haiku-3tier`
**KO title**: `3-티어 에이전트 아키텍처 (No-Haiku)`
**Pillar**: 3 (architectural foundation)

**Outline**:
1. **도입: 왜 Haiku를 배제했는가** — DeepSWE 리더보드 핵심 발견: "약한 모델 + 높은 effort = 가용성의 적" (sonnet-5 max $26.40/과제, opus-4.8 max $13.22/과제 — 단가 역전)
2. **3-티어 정의** — Mermaid TD showing the tier stack
   - Tier 1 (Mechanical / 기계) — Sonnet low (docs, git, mechanical refactors)
   - Tier 2 (Execution / 실행) — Opus high/medium (develop, harness implementation)
   - Tier 3 (Reasoning / 추론) — Fable high (spec, audit, design, advisor)
3. **DeepSWE 리더보드 근거** — full table ($/해결과제: opus $22.4 < fable $30.9 < sonnet-5 $48.9; 토큰/해결과제: fable 170k < opus 229k < sonnet 396k)
4. **설계 보고서 vs 구현** — REQ-DA-061 honesty caveat
   - 설계 단계 (design report, `.moai/reports/agent-architecture-redesign-v2-20260709.html`) — the v2 architecture intent
   - 구현된 동작 (SPEC-MODEL-TIER-PLANTYPE-001 CLOSED) — `ApplyTierProfile` 60-cell profile (link to plan-type-profiles)
5. **하네스 자가 진화와의 연결** — brief link to self-evolving (the architecture is the substrate the harness evolves on)

**Source**: `.moai/reports/agent-architecture-redesign-v2-20260709.html` + `project_model_tier_plantype_001_completed` memory.

### §C.4 plan-type-profiles

**Slug**: `plan-type-profiles`
**KO title**: `plan_type 티어 프로필`
**Pillar**: 3 (the 60-cell profile)

**Outline**:
1. **도입: plan_type 축** — `api` (종량제) vs `subscription` (구독) 과금 방식 분기
2. **plan_type 설정** — `moai init --plan-type` + `llm.plan_type` (llm.yaml) + `moai update --plan-type` 사후 전환
3. **60-셀 프로파일 매트릭스** — table or Mermaid TD showing 10 agents × 3 tiers × 2 plan_types
   - Plan A (api, rev2) — 추론=fable high, 실행=opus, 기계=sonnet low. 권장 A-medium.
   - Plan B (subscription) — 추론=opus high, 실행=sonnet high, 기계=sonnet low. 권장 B-max.
4. **ApplyTierProfile 메커니즘** — agent frontmatter의 model·effort 둘 다 교체 (전 agent에 effort: 있어 "보존"이면 무효라 replace-both)
5. **GLM 백엔드 effort 오버레이** — REQ-DA-060 **honesty caveat required**
   - 구현 + 배선 완료 (`IsGLMBackend` 감지 + 5→3단 collapse + coding-max override)
   - **wire 유효성 실증 예정** — z.ai가 Anthropic-compat shim으로 `ANTHROPIC_REASONING_EFFORT` 값을 실제 소비하는지는 라이브 GLM 세션 outbound 관측이 필요한 run-phase 실증 과제
   - 페이지에 "동작 보장"으로 서술 금지, "구현+배선 완료, wire 유효성 실증 예정"으로 기재
6. **모델 정책 보드 (moai web)** — `/model-policy` 보드에서 plan_type 쓰기 가능 (SPEC-WEB-CONSOLE-013의 승인된 예외)
7. **로드맵** — spawn-time 36-cell 라우팅 (SPEC-MODEL-TIER-ROUTING-PROFILES-001, descoped)

**Source**: `.moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/spec.md` (CLOSED, authoritative) + `.moai/reports/model-tier-redesign-20260712.md`.

### §C.5 self-evolving

**Slug**: `self-evolving`
**KO title**: `하네스 자가 진화`
**Pillar**: 3 (the 3-Loop)

**Outline**:
1. **도입: 왜 자가 진화인가** — Lilian Weng "Harness Engineering for Self-Improvement" (2026-07-04) — 하네스 경쟁력 = 자기개선 설계
2. **ACE 역할 모델** — Generator → Reflector → Curator (Weng's framework)
3. **3-Loop 구체화** — Mermaid TD of the 3-Loop
   - Loop 0 (관찰, 매 턴) — routing observation ledger (EVOLVE-001 closed)
   - Loop 1 (반추, 세션 경계) — auto-memory integration
   - Loop 2 (승격, tier 임계) — Curator editable surfaces (EVOLVE-002 closed) + production wiring (EVOLVE-003 closed)
4. **티어 ↔ 표면 매핑** — Tier1-2 → memory 임시 / Tier3 → CLAUDE.local.md 영구 / Tier4 → CLAUDE.md 관리 블록
5. **7 pillars (EVOLVE-003 PRODUCTION 배선)** — A1 frozen 확장, A6 tier↔surface, A7 negative-evidence, L2 canary, L3 contradiction, GLM observe-only, anti-fabrication
6. **로드맵 (REQ-DA-063 honesty)** — EVOLVE-004 (console verbs) + EVOLVE-005 (Recall wiring + typed parser) 진행 중; v5.1 MCE / v6 진화적 탐색 지평
7. **보상 해킹 방지 (A1)** — evaluator / 권한 (settings/hook/frozen-guard) 은 Frozen

**Source**: `.moai/reports/harness-self-evolving-redesign-final-20260712.html` (v5.1 FINAL SSOT) + SPEC-HARNESS-EVOLVE-{001,002,003} + `reference_lilian_weng_harness_selfimprove`.

### §C.6 autonomous-loops

**Slug**: `autonomous-loops`
**KO title**: `자율 연속 루프`
**Pillar**: 2

**Outline**:
1. **도입: 언제 멈추고 언제 계속할 것인가** — the agentic-loop question
2. **3가지 연속 루프 원시** — comparison table (Mermaid TD diagram of the 3 primitives)
   - `/goal` (native Claude Code) — 모델이 평가하는 종료 조건; 사용자 TUI 전용 (HUMAN-ONLY)
   - `/moai goal` (MoAI 프로그래밍) — same semantics, orchestrator-owned (Axis B)
   - `/moai loop` (Ralph Engine) — 진단 기반 결정론적 루프 (preset, not alias)
3. **native /goal 상세** — copy-able `/goal <condition>` template; evaluator cost; auto-compact + `/clear` interaction
4. **/moai goal 상세** — 4 verbs (arm/status/clear); session-start PruneOrphans; SPEC-GOAL-ENGINE-001 CLOSED
5. **/moai loop 상세** — Ralph Engine; bounded issue queue → goal engine; `/moai fix`와의 관계
6. **구현 vs 로드맵** — REQ-DA-062 distinction; AGENTIC-CORE epic 진행 중 (SPEC-1 ANALYZE-FIRST closed; SPEC-2 autonomous/semi-autonomous kickoff REQ 대기)
7. **안전 가드레일** — Implementation Kickoff Approval (plan→run HUMAN GATE)은 어떤 루프로도 bypass 불가; safety boundary unchanged

**Source**: `.claude/rules/moai/workflow/goal-directive.md` + `.moai/specs/SPEC-GOAL-ENGINE-001/` + `project_agentic_core_epic_progress` + `project_goal_engine_cli_gap_handoff`.

---

## §D Canonical-KO title table (4-locale main.yaml + _meta.yaml)

The 4-locale titles for the 6 new Advanced sub-entries. These strings populate both `main.yaml` (under the Advanced section's `sub:` array) and each per-locale `_meta.yaml` (as the `title:` field for each slug).

| Slug | ko | en | ja | zh |
|------|----|----|-----|-----|
| `tokenomics-overview` | 토크노믹스 개요 | Tokenomics Overview | トークノミクス概論 | 代币经济学概述 |
| `token-budget` | 토큰 예산 관리와 우아한 중단 | Token Budget Management and Graceful Stop | トークン予算管理と graceful stop | 代币预算管理与优雅停止 |
| `no-haiku-3tier` | 3-티어 에이전트 아키텍처 (No-Haiku) | 3-Tier Agent Architecture (No-Haiku) | 3層エージェントアーキテクチャ (No-Haiku) | 三层代理架构（无 Haiku） |
| `plan-type-profiles` | plan_type 티어 프로필 | plan_type Tier Profiles | plan_type ティアプロファイル | plan_type 层级配置 |
| `self-evolving` | 하네스 자가 진화 | Harness Self-Evolution | ハーネス自己進化 | 工具自我进化 |
| `autonomous-loops` | 자율 연속 루프 | Autonomous Continuation Loops | 自律連続ループ | 自主连续循环 |

**Rationale**: titles are short, descriptive, and use the English identifier (`plan_type`) where it is a code-level term (per agent-common-protocol.md § Language Handling — technical identifiers stay English). The "(No-Haiku)" parenthetical in `no-haiku-3tier` is preserved across all locales because it is a branding-significant phrase.

---

## §E Icon usage mapping (per page)

Recommended icon-shortcode usage to anchor semantic markers (per CLAUDE.local.md §17.1 — body-emoji prohibition). These are authoring suggestions, not requirements:

| Page | Section | Icon shortcode |
|------|---------|----------------|
| tokenomics-overview | 3-기둥 서사 개요 | `{{</* icon target */>}}` (3-pillar target) |
| tokenomics-overview | 4-층 구조 각 층 | `{{</* icon database */>}}` (A), `{{</* icon shuffle */>}}` (B), `{{</* icon wrench */>}}` (C), `{{</* icon warning */>}}` (D) |
| token-budget | 핸드오프 마커 | `{{</* icon warning warn */>}}` (soft), `{{</* icon warning danger */>}}` (hard) |
| token-budget | verify-diet | `{{</* icon check ok */>}}` |
| no-haiku-3tier | 3-티어 스택 | `{{</* icon database */>}}` (Tier 1), `{{</* icon flash */>}}` (Tier 2), `{{</* icon sparkles */>}}` (Tier 3) |
| plan-type-profiles | GLM caveat | `{{</* icon warning warn */>}}` (the wire-validity-pending caveat) |
| plan-type-profiles | 60-셀 매트릭스 | `{{</* icon package */>}}` |
| self-evolving | 3-Loop | `{{</* icon rotate */>}}` |
| self-evolving | roadmap | `{{</* icon clock */>}}` |
| autonomous-loops | 3 primitives | `{{</* icon arrow-right */>}}` |
| autonomous-loops | safety | `{{</* icon warning danger */>}}` |

---

## §F Open design decisions (definitively resolved at plan-phase)

### F.1 Menu placement: prepend vs interleave

**Decision**: PREPEND the 6 new entries in pillar order (Tokenomics → Harness → Loop) BEFORE the existing 14 entries, forming a "v3.0 pillar cluster" at the top of Advanced.

**Rationale**: the 6 new pages are conceptually distinct from the 14 component-reference pages (hooks/settings/statusline/claude-md). The 14 existing pages document individual components; the 6 new pages document product-level architecture. Grouping them at the top makes the v3.0 narrative visible at first glance.

**Alternative considered**: interleave by topic (e.g., autonomous-loops next to /moai goal under Workflow Commands). REJECTED — cross-section interleaving breaks the Advanced section's conceptual cohesion and the 6 pages belong together as the v3.0 pillar cluster.

### F.2 _meta.yaml counting rule (for CMD-A / CMD-E)

**Decision**: count entries by the `^[a-zA-Z_-]+:` pattern (key at column 0), EXCLUDING the `index:` block.

**Rationale**: geekdoc's `_meta.yaml` format uses either `"slug":\n  title: ...` or `slug:\n  title: ...` (both forms appear in the existing files). The most stable counting rule is the column-0 key pattern, excluding the `index:` special block.

**Expected post-M1 counts**: KO 14, EN 14, JA 14, ZH 14 (after parity debt fix).
**Expected post-M5 counts**: all 4 locales 20 (14 existing + 6 new).

### F.3 Per-page commit granularity

**Decision**: per-page commit for content (M2/M3/M4), combined commit for registration (M5), combined commit for verify (M6 — verification only, no content changes).

**Rationale**: per-page commits give clean revert boundaries if a single page's content needs revision. The registration step (M5) touches all 4 locales' _meta.yaml + main.yaml atomically — splitting it would create intermediate broken-registration states.

---

## §G Cross-references

- **spec.md** §B Group B — the GEARS requirements these pages satisfy (REQ-DA-010 through REQ-DA-015)
- **acceptance.md** §B Group B — the AC matrix rows for canonical-KO content markers (CMD-B1 through CMD-B6)
- **plan.md** §F — the milestone sequence (M2/M3/M4 per-page; M5 registration; M6 verify)
- **research.md** §B — per-page source-readiness evidence

---

Version: 0.1.0 | Tier: L | Status: draft (plan-phase)
