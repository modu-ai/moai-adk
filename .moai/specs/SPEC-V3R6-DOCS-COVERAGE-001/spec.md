---
id: SPEC-V3R6-DOCS-COVERAGE-001
title: "docs-site skill-count reconciliation + facts-bearing page coverage (Sprint 14 cohort 4/5)"
version: "0.2.0"
status: completed
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs-site, i18n, docs-truth, skill-count, reconciliation, 4-locale"
era: V3R6
depends_on: [SPEC-V3R6-DOCS-DOCSITE-001, SPEC-V3R6-DOCS-CODEMAPS-V3-001]
---

# SPEC-V3R6-DOCS-COVERAGE-001 — docs-site skill-count reconciliation

## §A. Context (배경)

본 SPEC은 Sprint 14 Docs-v3 코호트의 4/5 번째 SPEC으로, **SPEC-V3R6-DOCS-DOCSITE-001** (completed, origin a7c1b4d48)에서 명시적으로 연기된 skill-count 축을 단독으로 소유한다.

DOCSITE-001 spec.md §E (Exclusions)에 verbatim으로 기록됨:

> "'31 skills' 인접 drift — 별도 SPEC 대상. DOCSITE-001는 docs-truth 5 axes + language-count(REQ-008)에 한정. research.md에 기록만."

DOCSITE-001은 6개 docs-truth 축을 이미 정합했다: (a) agent catalog 8-retained, (b) archived-agent framing, (c) GLM tier-models glm-5.2[1m], (d) CLI 17 commands, (e) SPEC status enum + 12 frontmatter fields, (f) language count=16. **COVERAGE-001은 skill-count 축 + facts-bearing page coverage map만 소유하며, 상기 6축을 재건드리지 않는다.**

### 문제 정의

docs-site 4개 로케일(en/ko/ja/zh)에 "31 skills" / "31개 스킬" / "31 スキル" / "31个技能" stale fact가 산재한다. 그러나 primary source(`.claude/skills/`)의 실제 canonical count는 **32**이다(moai umbrella router 1 + 31 specialized). 또한 en/ko/zh 3 로케일은 Domain 카테고리에서 `moai-domain-humanize` 스킬이 누락되어 specialized 합계가 30으로 표기되고, **ja 및 ko 로케일은 v3 이전의 허구적 taxonomy(각각 9 카테고리, 37개씩 존재하지 않는 스킬명, 총 74개) 전체를 유지**하고 있어 단순 count patch가 아닌 구조적 재작성이 필요하다. 추가로 **en 및 zh 로케일은 category header 구조는 6 canonical로 정확하지만, Mermaid 다이어그램·코드 예제·ASCII tree·자동 로드 시나리오 등 개념 설명 영역에 허구적 스킬명(en: 10개, zh: 11개, 총 21개)이 잔존**하여 본문 정제(in-body cleanup)가 필요하다. 4-locale 허구적 스킬명 총합은 95개(en:10 + ko:37 + ja:37 + zh:11)이다.

상세 증거는 §C Background Research 및 research.md 참조.

### REQ-001 — Canonical skill count (Ubiquitous)

The docs-site **shall** state the canonical skill count as **32** (= `moai` umbrella router + 31 specialized skills) across all 4 locales (en/ko/ja/zh), traced to the primary source `find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d | wc -l`.

### REQ-002 — Category-specialist split (Ubiquitous)

The docs-site **shall** express the count as "1 umbrella router + 31 specialized skills classified into 6 categories" in every page that carries a skill-count claim, where the 6 categories are Foundation, Workflow, Domain, Reference, Meta/Harness, and Design.

### REQ-003 — Domain category completeness (Event-detected)

**When** the docs-site enumerates the Domain category skill list, the docs-site **shall** include all 9 actual `moai-domain-*` skills — specifically including `moai-domain-humanize` (Korean AI-tell taxonomy, added 2026-06-15) which is currently absent from all 4 locales.

### REQ-004 — Category sub-count integrity (State-driven)

**While** a docs-site page states per-category sub-counts (e.g., "Domain (8)"), the docs-site **shall** ensure each sub-count matches the primary-source `find` output per category prefix: Foundation=4, Workflow=10, Domain=9, Reference=5, Meta/Harness=2, Design=1, summed to 31 specialized + 1 umbrella = 32 total.

### REQ-005 — ja and ko locale structural reconciliation (Event-detected)

> **iter-2 scope expansion (2026-06-18):** iter-1 scoped this REQ to ja only. Independent re-derivation confirmed ko `advanced/skill-guide.md` carries the identical pre-v3 fictional taxonomy (9 categories, 37 nonexistent skill-name references — same magnitude as ja's 37). Both locales require the same structural treatment. This REQ is expressed as a unified requirement covering both locales; AC-006 verifies both via a single per-locale loop.

**When** the ja OR ko locale `advanced/skill-guide.md` carries a pre-v3 fictional taxonomy (9 categories, nonexistent skill names such as `moai-lang-*`, `moai-platform-*`, `moai-foundation-philosopher`, 37 references each), the docs-site **shall** replace that taxonomy with the canonical 6-category / 32-skill structure matching en/zh, eliminating all 37 references to nonexistent skill names per locale and introducing the 3 missing canonical categories (Reference, Meta/Harness, Design).

### REQ-006 — 4-locale parity invariant (Ubiquitous)

The docs-site **shall** apply every skill-count correction simultaneously across all 4 locales (en/ko/ja/zh). Single-locale correction constitutes a parity violation. The ja locale, despite needing deeper structural reconciliation, MUST land in the same commit boundary as en/ko/zh to preserve parity.

### REQ-007 — No invented locale-native phrasing (Unwanted)

The docs-site **shall not** introduce invented CJK locale idioms for the skill-count phrasing. Each locale MUST use its native idiom (en: "32 skills"; ko: "32개 스킬"; ja: "32スキル"; zh: "32个技能") verifiable against locale-native conventions, never machine-transliterated.

### REQ-008 — Facts-bearing page coverage scope (Capability gate)

**Where** a docs-site page carries a factual skill-count claim (number + skill-adjacent keyword), the docs-site **shall** be in the COVERAGE-001 edit surface. Pages that mention skills only in structural prose (navigation, breadcrumbs) without a numeric count are out of scope. The coverage map is enumerated in research.md § Coverage Map.

### REQ-009 — en/zh in-body fictional-name elimination (Event-detected)

> **iter-2 addition (2026-06-18):** independent grep found that en and zh `advanced/skill-guide.md` carry correct 6-canonical-category HEADERS but still reference nonexistent skill names (`moai-lang-*`, `moai-library-mermaid`, `moai-platform-supabase`) inside conceptual illustrations — Mermaid flowchart nodes, explicit-invocation code examples, ASCII directory trees, frontmatter examples, auto-load scenario comments, and closing callouts. These are not in the category-listing tables but still present fictional names as if they were shipped skills. This REQ closes that residual drift class. ko and ja are covered separately by REQ-005 (structural rewrite eliminates their 37 references each).

**When** the en OR zh locale `advanced/skill-guide.md` references a nonexistent skill name (matching the fictional-name regex `moai-lang-|moai-platform-|moai-library-|moai-framework-|moai-foundation-claude|moai-foundation-philosopher|moai-foundation-context`), the docs-site **shall** replace each such reference with a real-skill equivalent (e.g., `moai-lang-python` → `moai-domain-backend` with a note that Python patterns ship via `rules/moai/languages/`) or rephrase the illustration to avoid naming a nonexistent skill, so that the fictional-name regex returns 0 matches per locale post-correction (en: 10→0, zh: 11→0).

## §C. Background Research (요약)

상세 연구 결과는 `research.md` 참조. 핵심 발견:

1. **Canonical count = 32** (`find internal/template/templates/.claude/skills -maxdepth 1 -mindepth 1 -type d | wc -l` → 32). Template source와 local `.claude/skills/` 모두 동일(단 local은 user-owned `my-harness-*` 2개 추가 포함 = 34이나 이는 template scope 밖).
2. **en/zh**: "31 = umbrella + 30" → "32 = umbrella + 31" 수정 + Domain 카테고리에 `moai-domain-humanize` 추가(8→9). category header 구조는 6 canonical로 정확. 단, 개념 설명 영역(Mermaid·코드 예제·ASCII tree·자동 로드 시나리오)에 허구적 스킬명이 잔존(en: 10, zh: 11)하여 본문 정제 필요(REQ-009).
3. **ja 및 ko**: `advanced/skill-guide.md` 전체가 v3 이전 허구 taxonomy (각 9 categories, 각 37 nonexistent skills = 총 74). 구조적 재작성 필요 — en/zh와 동일한 6-category/32-skill 구조로 교체(REQ-005). ko 발견은 iter-2 재도출로 확인(iter-1은 ja-only로 scope 오류).
4. **Facts-bearing pages**: 5 page-families × 4 locales. `humanize`는 4 로케일 전체에서 0회 언급(drift의 원천 중 하나). 4-locale 허구적 스킬명 총합 = 95(en:10 + ko:37 + ja:37 + zh:11).

## §D. Constraints (제약사항)

1. **사실 정합만 수행** — docs-site IA 재설계, page merge/split/move, styling/build config 변경 금지. `nextra.config.*`, `vercel.json`, `_meta.yaml` 구조 동결. 내용 수정에 수반되는 최소한의 frontmatter 조정은 허용.
2. **4-locale parity는 load-bearing** — 모든 fact correction은 4 로케일에 동시 적용. 단일 로케일 correction = parity 위반. ja는 반드시 scan 대상 포함(DOCSITE-001 당시 orchestrator preliminary grep의 gap이었음).
3. **DOCSITE-001 scope 비중첩** — DOCSITE-001이 종료한 6축은 본 SPEC scope 밖. COVERAGE-001은 skill-count + coverage map만 소유.
4. **Primary-source traceability** — 모든 fact는 `.claude/skills/` 및 `internal/template/templates/.claude/skills/`에 추적. 허구 fact 생성 금지. count 숫자는 반드시 `find`/`ls` 명령의 verbatim 출력으로 입증.
5. **허구 CJK locale idiom 금지** — locale-native phrasing만 사용. 기계 번역 idiom 사용 시 zero-match false-pass 위험.

## §E. Exclusions (What NOT to Build)

### Out of Scope — Adjacent and excluded surfaces

- **DOCSITE-001이 종료한 6축** — agent catalog 8-retained / archived-agent framing / GLM tier-models / CLI 17 commands / SPEC status enum + 12 frontmatter fields / language count=16. 이 6축은 DOCSITE-001에서 정합 완료. 본 SPEC에서 재건드리지 않는다.
- **docs-site IA 재설계** — 페이지 신규 생성, 병합, 분할, 이동, URL 구조 변경, navigation 재구성은 본 SPEC scope 밖. 기존 페이지 내용의 사실 정합만 수행.
- **build/styling config** — `hugo.toml` / `config.toml` / `package.json` / theme 설정 / CSS / 레이아웃 템플릿 변경 금지. Hugo 빌드 구조 동결.
- **user-owned harness skills** — `my-harness-moaiadk-best-practices`, `my-harness-moaiadk-patterns` 2개는 user-owned(`moai update`가 보존)이며 template scope 밖. docs-site의 "built-in skills" count는 template-shipped 32에 한정하며 user-owned harness skills는 포함하지 않는다(per `.claude/CLAUDE.local.md` §24 Harness Namespace 분리 정책).
- **카테고리 내 스킬 description 본문** — 각 스킬의 description 문구(예: "moai-domain-backend: API design, microservices...")의 상세 표현 정합은 본 SPEC scope 밖. 본 SPEC은 count + category 구조 + `humanize` 누락 추가에 한정. description 정합은 별도 SPEC 대상.

## §F. Cross-References

- `.moai/specs/SPEC-V3R6-DOCS-DOCSITE-001/spec.md` §E — "31 skills" 본 SPEC으로 연기 명시
- `.moai/specs/SPEC-V3R6-DOCS-CODEMAPS-V3-001/` — 선행 코호트 (docs-truth.md refresh)
- `.moai/project/codemaps/docs-truth.md` — canonical facts checklist (본 SPEC이 skill-count 축을 추가 보강하는 대상; 단 본 SPEC 범위에서는 docs-truth.md 수정을 plan.md M5 후속으로 기록만)
- `.claude/skills/` — skill-count PRIMARY SOURCE (local, user-owned 포함 34)
- `internal/template/templates/.claude/skills/` — skill-count PRIMARY SOURCE (template-shipped 32, canonical)
- `scripts/docs-i18n-check.sh` + `.github/workflows/docs-i18n-check.yml` — i18n parity gate
- `.claude/CLAUDE.local.md` §24 — Harness Namespace 분리 정책 (user-owned vs template-shipped 구분 근거)
- `.claude/NOTICE.md` — im-not-ai import 기록 (`moai-domain-humanize` 추가 2026-06-15 provenance)

## §G. Risks (위험)

| Risk | Severity | Mitigation |
|------|----------|------------|
| ja 및 ko locale 구조적 재작성 중 번역 품질 저하 | Medium | en/zh의 6-category 구조를 template으로 사용; locale-native idiom만 사용; 기계 번역 금지(REQ-007) |
| ko structural divergence 누락 (iter-1이 ja-only로 scope 오류) | Medium | iter-2 재도출로 ko 37개 허구적 스킬명 확인; REQ-005가 ja+ko 통합, AC-006이 `for loc in ja ko` per-locale loop로 2회 검증 |
| en/zh in-body 허구적 스킬명 잔존 (Mermaid·코드·tree·callout) | Medium | REQ-009가 fictional-name regex 0-match를 요구; AC-011이 `for loc in en zh` per-locale loop로 검증 |
| `humanize` 추가 시 다른 카테고리 sub-count 연쇄 영향 | Low | Domain만 8→9 변경, 다른 5개 카테고리는 불변(REQ-004); 각 sub-count를 primary-source `find`로 재검증 |
| 4-locale 동시 적용 누락 (update.md statusline string은 en/zh에만 존재) | Low | research.md Coverage Map이 page-family × locale matrix를 명시; ko/ja `update.md`는 파일 존재하나 statusline string 무소유 — AC에서 per-locale grep로 독립 검증 |
| user-owned harness skills를 count에 포함하는 실수 | Low | REQ-002가 template-shipped 32에 한정 명시; §E.4 Exclusions로 방어 |

## §H. Success Criteria (성공 기준)

1. 4 로케일 모두에서 "31" skill-count claim이 "32"로 정정됨 (digit-boundary-anchored grep로 0残留 검증).
2. en/ko/zh의 Domain 카테고리에 `moai-domain-humanize`가 포함되고 sub-count가 9로 정정됨.
3. ja 및 ko `advanced/skill-guide.md`의 허구 taxonomy가 제거되고 canonical 6-category/32-skill 구조로 교체됨 (각각 37개씩, 총 74개 허구적 스킬명 제거).
4. en 및 zh `advanced/skill-guide.md`의 개념 설명 영역 허구적 스킬명이 제거됨 (en: 10, zh: 11, 총 21개; REQ-009).
5. 4-locale parity: 각 locale에서 동일한 count(32)와 동일한 category 구조(6 categories)가 관측됨.
6. `moai spec lint` 0 findings.
7. primary-source `find` 출력이 research.md에 verbatim으로 기록되어 모든 count가 추적 가능함.
