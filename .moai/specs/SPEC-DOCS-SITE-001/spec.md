---
id: SPEC-DOCS-SITE-001
version: 0.2.0
status: Planned
created: 2026-04-17
updated: 2026-04-20
author: manager-spec
priority: High
issue_number: null
---

# SPEC-DOCS-SITE-001 — moai-docs 흡수 및 Hextra 전면 전환

## HISTORY

- 2026-04-20 v0.2.0 (Iteration 2): Plan Auditor g1-audit-report.md의 25개 결함 반영.
  - D-001 REQ 개수 상한 36으로 상향 (AC-G1-02 + plan.md §7 G1 동시 통일). REQ-DS-32 분할(32a/32b) + REQ-DS-35 신규로 총 36건 수용.
  - D-002 REQ-DS-32를 REQ-DS-32a (재바인딩, Phase 7 cutover preparation 진입 시) / REQ-DS-32b (Production 프로모션, G4 승인 후)로 분할.
  - D-003 Phase 2 전용 게이트 G1.5 (Scaffold Readiness) 신설, plan.md §7 추가.
  - D-004 DNS TTL 조항을 REQ-DS-32 본문에서 제거하고 Phase 7 사전 조사 AC-PRE-01로 이관 (필요성 미검증 조작 방지).
  - D-005 AC-G2-06 tolerance 제거, `test "$COUNT" -eq 569`로 강화.
  - D-006 AC-G2-04 기준을 `test "$TOTAL" -eq 735`로 강화.
  - D-007 REQ-DS-13에 `runtime: 'edge'` 명시, Phase 5 platform constraint 사전 검토 단계 추가, AC-G3-01에 runtime 검증 추가.
  - D-008 plan.md §7 G1 범위를 acceptance.md와 일치시켜 "15~35"로 통일.
  - D-009 누락 REQ 대응 AC 추가 (REQ-DS-19/20/29 자동 검증 AC 신설: AC-G2-08 / AC-G3-08 / AC-G2-09).
  - D-010 REQ-DS-17에 스냅샷 대상 커밋 (previous release tag commit) 명시.
  - D-011 AC-G3-04 테스트 이름 교정 + TestMajorRelease (v2.12.0 → v3.0.0) 추가.
  - D-012 Non-goals에 "이모지 금지 정책에 따른 국기 이모지 → 텍스트 라벨 전환은 예외 허용" 조항 추가.
  - D-013 AC-G4-07에 "Nextra baseline 대비 Performance -5 이내" 조항 추가.
  - D-014 신규 REQ-DS-35 (Unwanted, 48h 롤백 의무) 추가 + AC-MON-03 대응.
  - D-015 Hextra 통합 방식을 Hugo module system으로 잠금 (§7 제약에 명시).
  - D-016 REQ-DS-15 "empty skeleton placeholder folders" 내용 구체화 (`_index.md` + `draft: true` + redirect).
  - D-017 Vercel 빌드 timeout 선제 조사 지침을 plan.md Phase 2 완료 단계에 추가.
  - D-018 §4 용어집 뒤에 "Hextra 한계" 소섹션 신설 (versioning 내장 없음, React/JSX 미지원).
  - D-019 Phase 4 pre-production SEO 영향 조사 AC (AC-PRE-02) 추가.
  - D-020 plan.md §7 G2 표기 "(Phase 3/4 → Phase 5)"로 수정.
  - D-021 REQ-DS-34 주어를 "the system (Phase 8 automation)"으로 명확화.
  - D-022 Exclusions에 moai-docs의 gate.yaml / memo.yaml / observability.yaml 3개 이관 제외 명시.
  - D-023 REQ-DS-19 배너 구현 메커니즘을 plan.md Phase 5에 구체화.
  - D-024 AC-G3-05 정규식을 `grep -Eq "^### *(§?17\.[1-6])"`로 강화.
  - D-025 AC-G4-03에 `$PREVIEW_URL` 초기화 지침 추가.
  - Gap 1~7: Vercel 무중단 근거 AC-PRE-03, subtree 크기 실측 AC-G2-10, Hugo 빌드 시간 베이스라인 AC-G2-11, FlexSearch 기본 비활성 결정 고정, llms.txt URL 검증 AC-G4-11, og.png 크기 AC-G4-05 보강, `_meta.ts` JSON Schema CI AC-G3-09 추가.
- 2026-04-17 v0.1.0: 초안 작성. Phase 0 Discovery 산출물 (migration-inventory.md 825 LOC + spec-project-diff.md 562 LOC) 기반. 사용자 승인 결정사항 D1~D4 반영. SPEC-I18N-001 흡수 결정 포함.

## 1. 배경 (Context)

현재 `adk.mo.ai.kr` 공식 문서 사이트는 독립 레포 `github.com/modu-ai/moai-adk-docs`에서 Nextra 4 + Next.js 16 + Bun 스택으로 운영되며 Vercel 프로젝트 `prj_EZaVdfE3gJeXVbizafBEECpniINP` 로 배포된다. 해당 사이트는 4개 locale (ko/en/ja/zh) × 52~63 페이지 총 **219 MDX 페이지**와 **735건의 `<Callout>` 사용**, **569개의 Mermaid 블록**, **38개의 `_meta.ts` 사이드바 설정**, **164 LOC의 Edge middleware**, 9개 커스텀 React 컴포넌트를 포함한다.

moai-adk-go 프로젝트가 Go 단일 바이너리 기반으로 정리되면서, 문서 사이트도 Node/Bun 생태계 의존을 제거하고 Hugo 단일 바이너리 빌드로 통합 관리할 필요가 대두되었다. 또한 SPEC-I18N-001 (moai-docs 레포의 유일한 SPEC)의 용어집/번역 체인은 유지하되 Nextra 의존 부분은 재작성해야 한다.

본 SPEC은 **moai-docs 전체 콘텐츠를 moai-adk-go 레포의 `docs-site/` 서브디렉토리로 이관**하고, 문서 생성기를 **Nextra → Hextra (Hugo 테마)** 로 전면 교체하며, `adk.mo.ai.kr` 서비스의 무중단 Vercel 전환을 달성하는 것을 목표로 한다.

## 2. 목표 (Goals) / 비목표 (Non-goals)

### Goals (In-scope)

- moai-docs 레포의 `content/` (219 MDX 페이지), `public/` (정적 자산), `theme.config.tsx` 메타, `middleware.ts` i18n 로직, `vercel.json` 리다이렉트 2건을 `docs-site/` 하위로 100% 이관한다.
- Nextra 4 / Next.js 16 / Bun / Biome / Tailwind v4 / Radix UI / Playwright(docs 전용) 의존을 **모두 제거**하고, Hugo + Hextra 단일 스택으로 운영한다.
- 4개 locale (ko/en/ja/zh) 콘텐츠 구조를 그대로 보존한다 (ko 63, en/ja/zh 각 52). ko 전용 `contributing/`, `multi-llm/` 2개 섹션은 ko 로케일에 한해 유지한다.
- 기존 URL 구조 `/{locale}/{section}/{slug}` 를 unversioned latest로 보존하며, 과거 버전만 `/{locale}/v2.X/{section}/{slug}` 로 추가 노출한다 (D2).
- `adk.mo.ai.kr` 도메인은 무중단으로 신규 빌드에 연결한다. Vercel 프로젝트 바인딩을 `moai-adk-docs` 에서 `moai-adk-go/docs-site` 로 전환한다 (D4).
- SPEC-I18N-001의 용어집·번역 체인·부작용 방지 규정을 본 SPEC으로 흡수한다.
- `CLAUDE.local.md` 에 `§17 docs-site 4개국어 문서 동기화 규칙` 섹션을 신설하여 모든 docs-site 변경을 다국어 동기화 관점에서 통제한다.
- `.moai/project/{product,structure,tech}.md` 3종 문서를 docs-site 신규 구성요소를 반영하여 갱신한다.

### Non-goals (Out-of-scope)

- 디자인 개편 — Hextra 기본 테마 + moai-docs 기존 색상/로고만 반영하며, 시각적 리디자인은 수행하지 않는다.
  - **예외 (coding-standards.md 필수 적용)**: 이모지 금지 정책에 따라, 기존 LanguageSelector의 국기 이모지(🇰🇷/🇺🇸/🇯🇵/🇨🇳)는 "KO / EN / JA / ZH" 텍스트 라벨로 전환한다. 이 전환은 디자인 리디자인이 아니라 코딩 표준 준수를 위한 필수 허용 변경이며, 색상·레이아웃·로고·타이포그래피 등 그 외 시각 요소는 기존과 동일하게 유지한다.
- 신규 문서 콘텐츠 추가 — 기존 콘텐츠 이식에 집중하며, 새로운 챕터/페이지/번역은 본 SPEC 범위 외이다.
- `mo.ai.kr` 등 다른 Vercel 프로젝트는 변경하지 않는다.
- LaTeX 지원 — 실제 사용 0건으로 확인되었으므로 제거한다.
- 한국어 외 로케일 (en/ja/zh)의 ko 전용 섹션 (`contributing`, `multi-llm`) 번역 — 본 SPEC에서는 빈 스켈레톤 폴더만 생성하고, 실제 번역은 후속 SPEC으로 분리한다.
- Dead code 이식 — `components/ui/*` 4개 (dialog/card/button/input), `CodeBlock.tsx`, `navbar.tsx` 의 `GitHubStarBadge`, `lib/page-map.ts`, `app/[...meta].json`, `app/_content.mdx`, `app/debug-pagemap/` 는 이식 대상에서 제외한다.

## 3. 이해관계자 (Stakeholders)

| 역할 | 이름/팀 | 책임 |
|------|---------|------|
| Product Owner | GOOS | 최종 승인, 도메인/Vercel 전환 게이트 (G4) 결정 |
| Tech Lead | manager-strategy | Phase 2~6 아키텍처 결정, Hextra 매핑 전략 |
| Implementation Lead | expert-frontend, manager-ddd | Hextra 테마 overlay, shortcode 이식 |
| Translation QA | manager-docs | 4개 locale 동기화 확인, SPEC-I18N-001 용어집 준수 |
| Release | manager-git | git subtree 병합, archive 태그 발행, 릴리스 스냅샷 자동화 |
| Infrastructure | expert-devops | Vercel Project 재바인딩, Edge Function 배포, DNS 모니터링 |

## 4. 용어집 (Glossary)

| 용어 | 정의 |
|------|------|
| docs-site | moai-adk-go 레포의 `docs-site/` 디렉토리. Hugo + Hextra 기반 정적 사이트 소스. |
| Hextra | Hugo용 문서 테마 (imfing/hextra). shadcn 스타일 UI, i18n, FlexSearch 내장. |
| 219 MDX 페이지 | moai-docs `content/` 전체 페이지 수. ko 63 + en 52 + ja 52 + zh 52. |
| Callout | Nextra의 경고/팁 컴포넌트. Hextra에서는 `{{< callout type="..." >}}` shortcode로 이식. |
| `_meta.ts` | Nextra의 사이드바 메뉴/순서 정의 파일 (38개). Hugo frontmatter `weight` + `title`로 변환. key 순서는 `_meta.ts`에 기록된 출현 순서대로 10단위(10, 20, 30, ...)로 `weight`에 매핑하며, 중간 삽입 여지를 확보한다. |
| frontmatter 주입 | 219 MDX 페이지 전량에 YAML frontmatter (title, weight, draft) 를 스크립트로 일괄 추가하는 작업. |
| v2.12.0 스냅샷 | 첫 번째 버전 스냅샷. D2 결정에 따라 **v2.13 릴리스 시점**에 `content/{locale}/` 트리를 그 시점의 **v2.12.X 최종 릴리스 태그 커밋 기준**으로 복사한다. `HEAD of main` 기준이 아니다. |
| Vercel Project 재바인딩 | 기존 Vercel 프로젝트 (`prj_EZaVdfE3gJeXVbizafBEECpniINP`) 의 Git 연결을 `moai-adk-docs` 에서 `moai-adk-go` 로 교체하고 Root Directory를 `docs-site/` 로 설정하는 작업. |
| Edge Function i18n | 기존 `middleware.ts` 의 Accept-Language + cookie 로직을 Vercel Edge Function (`api/i18n-detect.ts` 또는 동등 경로) 으로 이식한 런타임 컴포넌트. |
| Gate G1 / G1.5 / G2 / G3 / G4 | Phase별 완료 확인 게이트. G1.5는 Phase 2 Scaffold Readiness, G4는 사용자 수동 승인 (Vercel 전환). |

### 4.1 Hextra 한계 (Acknowledged Limitations)

본 SPEC이 Hextra를 선택하면서 수용하는 본질적 한계를 사전 명시한다. 후속 요구가 발생할 때 재검토 없이 거절할 수 있도록 한다.

- **내장 versioning 기능 없음** — Hextra는 Docusaurus와 달리 버전별 문서 스냅샷을 내장 지원하지 않는다. 본 SPEC은 `scripts/docs-version-snapshot.go` Go 스크립트로 릴리스 태그 시점에 `content/{locale}/v2.X/` 폴더를 생성하는 방식으로 우회한다 (REQ-DS-17, D2).
- **React/JSX 컴포넌트 실행 불가** — Hextra는 Hugo 기반 정적 사이트 생성기이므로 React/JSX 런타임을 제공하지 않는다. 향후 인터랙티브 컴포넌트 요구가 발생해도 본 스택 내에서는 Hugo partial + 바닐라 JavaScript로만 구현해야 한다. MDX import/export 구문은 지원되지 않으며, 모든 커스텀 시각 요소는 Hugo shortcode 또는 partial로 재작성한다.
- **서버 사이드 Mermaid 프리렌더링 미제공** — Mermaid 다이어그램은 클라이언트 측 JavaScript로만 렌더링된다 (REQ-DS-10). 초기 로드 성능 저하가 관측되면 Phase 8 이후 별도 SPEC으로 프리렌더 파이프라인을 검토한다.

## 5. 참조 문서 (Reference Documents)

본 SPEC의 근거가 된 Phase 0 Discovery 산출물:

- `/Users/goos/MoAI/moai-adk-go/.moai/plans/DOCS-SITE/migration-inventory.md` (825 LOC) — moai-docs 전체 인벤토리. 페이지 수, 컴포넌트 사용, Nextra 의존, 리스크 R1~R10.
- `/Users/goos/MoAI/moai-adk-go/.moai/plans/DOCS-SITE/spec-project-diff.md` (562 LOC) — `.moai/*` 차이 분석, SPEC-I18N-001 이관 매트릭스, `CLAUDE.local.md §17 DRAFT`.

사용자 승인 결정사항 (본 SPEC 전반에 적용, 변경 금지):

- **D1 Git 이관 방식**: `git subtree add --squash --prefix=docs-site https://github.com/modu-ai/moai-adk-docs main` — squash merge. 39 MB `.git` 을 약 50 KB 단일 커밋으로 압축. 원 이력은 `moai-adk-docs` archive 레포에서 참조.
- **D2 버전 URL 구조**: 최신은 unversioned (`/{locale}/{section}/{slug}`), 과거 버전만 `/{locale}/v2.X/{section}/{slug}`. **v2.12.0은 첫 스냅샷으로 생성하지 않음** — v2.13 릴리스 시 v2.12 폴더가 최초 스냅샷으로 동결된다.
- **D3 middleware i18n 대체**: Vercel Edge Function 으로 이식. cookie → Accept-Language → default ko 3단계 우선순위 로직 100% 보존.
- **D4 moai-docs 레포 archive 시점**: Phase 7 G4 통과 **+ 48시간 Vercel 모니터링 완료** 후 `archive/pre-migration-2026-04` 태그 발행 + GitHub 레포 Archive 플래그 적용.

## 6. 요구사항 (EARS Requirements)

### 6.1 Git 이관 및 디렉토리 구조

**REQ-DS-01 (Ubiquitous)** — The system shall house all docs-site source files under the `docs-site/` directory at the moai-adk-go repository root, imported via `git subtree add --squash --prefix=docs-site` from `https://github.com/modu-ai/moai-adk-docs` (D1).

**REQ-DS-02 (Ubiquitous)** — The docs-site directory shall contain a Hugo-compatible layout consisting of `hugo.yaml`, `content/{ko,en,ja,zh}/`, `layouts/`, `static/`, `assets/`, `i18n/`, `data/`, and `vercel.json`, with `public/` listed in `.gitignore`.

**REQ-DS-03 (Unwanted)** — If any Node, Bun, Next.js, Nextra, Biome, Tailwind v4, Radix UI, or React runtime artifact (including `package.json`, `bun.lock`, `biome.json`, `next.config.mjs`, `theme.config.tsx`) is detected inside `docs-site/`, then the system shall fail the build with a non-zero exit code.

### 6.2 콘텐츠 이식 및 Hextra 매핑

**REQ-DS-04 (Ubiquitous)** — The system shall migrate 219 MDX pages from moai-docs `content/` to `docs-site/content/` as Hugo-compatible Markdown files, preserving the 4-locale directory structure (ko 63 pages, en/ja/zh 52 pages each).

**REQ-DS-05 (Event-driven)** — When the migration script encounters an `import { Callout } from 'nextra/components'` statement, the system shall remove the import line and convert all `<Callout type="tip|info|warning|error|success">...</Callout>` instances (735 total, absolute preservation required) to the equivalent Hextra `{{< callout type="..." >}}...{{< /callout >}}` shortcode with 1:1 type mapping.

**REQ-DS-06 (Ubiquitous)** — The system shall inject YAML frontmatter into every one of the 219 migrated pages via a Go script `scripts/inject-frontmatter.go`, setting at minimum `title` (derived from the first H1), `weight` (derived from `_meta.ts` key order), and `draft: false`.

**REQ-DS-07 (Event-driven)** — When the migration script reads a `_meta.ts` file (38 total), the system shall generate equivalent Hugo `_index.md` files (one per section/locale) with `menu` frontmatter preserving key order as `weight` in increments of 10 (e.g., first key → 10, second → 20), and shall reproduce `display: "hidden"` semantics via `_build.list: never` + `_build.render: always`.

**REQ-DS-08 (Ubiquitous)** — The `_meta.ts` to Hugo frontmatter conversion shall be implemented as `scripts/convert-meta-to-frontmatter.go` in the moai-adk-go Go codebase, not as a JavaScript/TypeScript tool.

### 6.3 Mermaid 다이어그램

**REQ-DS-09 (Ubiquitous)** — The system shall preserve all 569 Mermaid code blocks (absolute count, no tolerance) across 4 locales in their original ` ```mermaid ... ``` ` fenced-code format, relying on Hextra's built-in Mermaid support.

**REQ-DS-10 (State-driven)** — While Hextra renders Mermaid diagrams client-side in the browser via a JavaScript runtime, the system shall document this rendering model in `plan.md` § Mermaid Strategy and shall not attempt server-side SVG pre-rendering in the initial migration.

**REQ-DS-11 (Optional)** — Where a page contains more than 5 Mermaid blocks, the system may apply Hextra's lazy-load configuration to defer rendering until viewport intersection, provided the configuration does not break existing diagrams.

### 6.4 다국어 (i18n) 및 라우팅

**REQ-DS-12 (Ubiquitous)** — The system shall preserve the URL structure `/{locale}/{section}/{slug}` for all 219 migrated pages with no breaking changes to existing links (D2).

**REQ-DS-13 (Event-driven)** — When a user requests the site root `/` without a locale prefix, a Vercel Edge Function located at `docs-site/api/i18n-detect.ts` (or equivalent path) shall determine the target locale using the priority chain cookie (`locale`, 1-year maxAge) > `Accept-Language` header > default `ko`, and shall redirect with HTTP 302 to `/{locale}{originalPath}` (D3). The function shall declare `export const config = { runtime: 'edge' };` (or equivalently `export const runtime = 'edge';`) to ensure Edge runtime deployment instead of the default Serverless Function (Node.js) deployment.

**REQ-DS-14 (Unwanted)** — If the Edge Function encounters a path matching `/api/`, `/_next/`, `/static/`, or a path ending in a file extension, then the system shall bypass locale detection and pass the request through unchanged.

**REQ-DS-15 (State-driven)** — While the site uses the 4-locale set (ko, en, ja, zh) with `defaultContentLanguage: ko`, the system shall serve the `ko`-only sections (`contributing/`, `multi-llm/`) exclusively under the `ko` locale and shall create placeholder folders under en/ja/zh containing a single `_index.md` with `draft: true` frontmatter plus a redirect alias to the equivalent ko locale path, to prevent link-checker failures while making the missing-translation state explicit.

### 6.5 버전별 문서 관리

**REQ-DS-16 (State-driven)** — While the latest documentation release is unversioned at `/{locale}/{section}/{slug}`, the system shall provide archived documentation at `/{locale}/v2.X/{section}/{slug}` for each prior major/minor release (D2).

**REQ-DS-17 (Event-driven)** — When a new major or minor release (e.g., v2.13.0) is tagged, the release automation shall invoke the snapshot script against the commit tagged as the previous release (e.g., the commit tagged v2.12.X), not HEAD of main, via `scripts/docs-version-snapshot.go <previous-version>`, which copies that tagged commit's `content/{locale}/` tree (minus existing `v*` subdirectories) into `content/{locale}/v<previous-version>/`, preserving the 4-locale structure.

**REQ-DS-18 (Unwanted)** — If a patch release (e.g., v2.12.1) is tagged, then the snapshot script shall not be invoked; only the unversioned latest content shall be updated.

**REQ-DS-19 (State-driven)** — While a user is viewing any `v2.X` archived documentation page, the Hextra layout shall display a banner at the page top reading "Viewing v2.X. Go to latest" with a link to the unversioned equivalent, implemented as a Hugo partial conditionally rendered when the request path matches `v[0-9]+\.[0-9]+/` (implementation details: see plan.md Phase 5 step 2).

### 6.6 커스텀 컴포넌트 / SEO

**REQ-DS-20 (Ubiquitous)** — The system shall reproduce three custom partials under `docs-site/layouts/partials/custom/`: (a) `language-switch.html` replacing the React LanguageSelector with a 4-locale dropdown using "KO / EN / JA / ZH" text labels (no flag emoji, per coding-standards.md), (b) `navbar-end.html` replicating ClientNavbar's GitHub icon and static stars badge, and (c) `structured-data.html` rendering JSON-LD output. All three partials shall preserve visual and functional equivalence except for the explicit emoji-to-text-label transition.

**REQ-DS-21 (Unwanted)** — If the `structured-data` partial includes an `aggregateRating` JSON-LD field, then the system shall omit it entirely; the field was hardcoded (`ratingValue: "4.8"`, `ratingCount: "42"`) without verifiable basis and violates Google Search Central policy.

**REQ-DS-22 (Ubiquitous)** — The system shall preserve SEO artifacts (`sitemap.xml`, `robots.txt`, Open Graph images, JSON-LD Organization/SoftwareApplication/WebSite/TechArticle schemas) with equivalent content, and shall preserve hreflang entries for 5 language codes (ko, en, zh, ja, x-default).

**REQ-DS-23 (Ubiquitous)** — The system shall preserve the two existing redirect rules in `vercel.json`: `/:locale(ko|en|ja|zh)/moai-rank/:path*` → `/:locale` and `/moai-rank/:path*` → `/`, both with `permanent: true`.

### 6.7 설정 문서 갱신 (SPEC 내부 작업 산출물)

**REQ-DS-24 (Event-driven)** — When this SPEC reaches Phase 6, the system shall append `§17 docs-site 4개국어 문서 동기화 규칙` to `/Users/goos/MoAI/moai-adk-go/CLAUDE.local.md` with subsections §17.1 (URL 표준 & 블랙리스트), §17.2 (Markdown/Hextra 작성 규칙), §17.3 (4개국어 동기화 규칙), §17.4 (버전 스냅샷), §17.5 (실행 주체), §17.6 (빌드/배포 체크리스트), based on the DRAFT in `spec-project-diff.md` § 17.

**REQ-DS-25 (Ubiquitous)** — The system shall update `.moai/project/product.md` to add "Public Documentation Site (adk.mo.ai.kr)" as a new Core Feature and "Documentation Readers" as a new Target Audience.

**REQ-DS-26 (Ubiquitous)** — The system shall update `.moai/project/structure.md` to document the `docs-site/` top-level directory layout, the Hugo build pipeline, and the Vercel deployment topology.

**REQ-DS-27 (Ubiquitous)** — The system shall update `.moai/project/tech.md` to add a "Documentation Site Stack" section listing Hugo v0.140+, Hextra theme, Mermaid v11+ (client-side), Hextra FlexSearch (disabled by default in this SPEC, enabling deferred to post-Phase 8 follow-up), and Vercel Hugo preset, and shall explicitly mark Nextra/Next.js/MDX/Bun/Biome/Tailwind-v4/Radix-UI/Playwright-docs as removed.

### 6.8 SPEC-I18N-001 흡수

**REQ-DS-28 (Ubiquitous)** — The system shall absorb SPEC-I18N-001 (moai-docs `/Users/goos/moai/moai-docs/.moai/specs/SPEC-I18N-001/`) into SPEC-DOCS-SITE-001 by: (a) preserving REQ-5 glossary, REQ-7 side-effect prevention, AC-5/AC-7 acceptance scenarios, and the ko→en→zh/ja translation source chain as binding requirements here; (b) rewriting REQ-4 (MDX/JSX) as Hextra shortcode requirements (REQ-DS-05, REQ-DS-07); (c) rewriting REQ-6 (`npm run build`) as Hugo build requirements (REQ-DS-30); (d) archiving the original SPEC-I18N-001 under `.moai/specs/SPEC-I18N-001/` in moai-adk-go with `status: archived` and `supersede_by: SPEC-DOCS-SITE-001` metadata added.

### 6.9 품질 및 빌드

**REQ-DS-29 (Ubiquitous)** — The system shall fail the CI build if any of the following anti-patterns are detected in `docs-site/`: forbidden URLs (`docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr`), Mermaid `flowchart LR` / `graph LR` direction, or emoji characters in documentation body (per `coding-standards.md`). The CI check shall be implemented as `scripts/docs-forbidden-check.sh` with explicit regex patterns for each anti-pattern category.

**REQ-DS-30 (Ubiquitous)** — The `docs-site/` production build shall succeed via `hugo --minify --gc` with zero warnings and shall emit a complete static site under `docs-site/public/` including a generated `sitemap.xml`.

**REQ-DS-31 (Event-driven)** — When a pull request modifies any file under `docs-site/content/ko/`, the CI pipeline shall execute `scripts/docs-i18n-check.sh` which verifies (a) equal file counts across 4 locales (allowing ko-only exceptions for `contributing/` and `multi-llm/`), (b) identical relative paths after locale prefix removal, (c) non-empty frontmatter `title`, and (d) MoAI glossary compliance (e.g., "MoAI-ADK" remains untranslated in all locales).

### 6.10 Vercel 전환 및 Archive

**REQ-DS-32a (Event-driven)** — When Phase 7 enters the cutover preparation step (prior to Preview validation and prior to G4 approval), the operator shall reconfigure the existing Vercel project `prj_EZaVdfE3gJeXVbizafBEECpniINP` with: (a) Git source → `modu-ai/moai-adk-go`, (b) Root Directory → `docs-site`, (c) Framework Preset → Hugo, (d) Ignored Build Step → `git diff --quiet HEAD^ HEAD ./docs-site/ || exit 1`. DNS is **not** modified during this step (see AC-PRE-01 for DNS necessity assessment).

**REQ-DS-32b (Event-driven)** — When acceptance.md § G4 checklist passes and GOOS grants G4 approval via AskUserQuestion, the operator shall promote the current Preview deployment to Production via the Vercel Dashboard, without any further Git source or framework preset modification.

**REQ-DS-33 (State-driven)** — While the production cutover is in progress, the operator shall first validate a Preview deployment against the acceptance scenarios in `acceptance.md` § G4 before promoting to production.

**REQ-DS-34 (Event-driven)** — When Phase 7 G4 passes and 48 hours of production stability are observed (no P1/P2 incidents, DNS resolution stable), the system (Phase 8 automation, executed by manager-git and expert-devops) shall archive the `modu-ai/moai-adk-docs` repository by: (a) tagging `archive/pre-migration-2026-04`, (b) enabling the GitHub Archive flag, (c) updating `README.md` of the archived repo with a pointer to `moai-adk-go/docs-site` (D4).

**REQ-DS-35 (Unwanted)** — If any AC-MON-01 threshold is violated during the 48-hour monitoring window (5xx error rate ≥ 0.1%, P1/P2 incident count ≥ 1, LCP p75 > 2.5s, DNS resolution failure reported, or critical Discord user report), then the operator shall immediately revert the Vercel project Git source from `modu-ai/moai-adk-go` back to `modu-ai/moai-adk-docs` main, and shall record the incident in `.moai/plans/DOCS-SITE/phase-7-rollback.md` including violation trigger, revert timestamp, and user-facing impact summary.

## 7. 제약 (Constraints)

- **성능**: 초기 페이지 로드 시 4G 네트워크 기준 Largest Contentful Paint 2.5s 이하 유지. 기존 Nextra 베이스라인보다 느려서는 안 된다. Phase 7 Preview 검증 시 Nextra baseline을 사전 측정하여 비교 근거로 사용한다 (AC-G4-07).
- **접근성**: WCAG 2.1 AA 준수. Hextra 기본 테마가 제공하는 수준 유지.
- **SEO**: 기존 URL 구조 100% 보존. sitemap.xml 엔트리 수가 기존 대비 감소하지 않아야 한다 (단, `/claude-code/*` 7개 미사용 경로는 의도적 제거 허용).
- **언어**: 4개 locale (ko/en/ja/zh) 이외 언어 추가 금지. 본 SPEC은 기존 4개 유지에 한정.
- **의존성**: Hugo 바이너리 단 1개로 빌드 가능해야 한다. Node/Bun/Python 런타임 의존 금지.
- **Hextra 통합 방식 (D-015 잠금)**: Hextra는 **Hugo module system (`docs-site/go.mod` import)** 으로만 통합한다. `git submodule` 방식은 사용하지 않는다. 선택 이유: 버전 잠금이 `go.mod`에 명시적이며, CI/Vercel 환경에서 추가 `git submodule init` 단계가 불필요하다.
- **이모지 금지**: `.claude/rules/moai/development/coding-standards.md` 에 따라 `docs-site/layouts/`, `docs-site/content/`, `docs-site/i18n/` 에 이모지 삽입 금지 (국기 이모지 포함). 이 제약이 LanguageSelector의 국기 이모지를 "KO / EN / JA / ZH" 텍스트 라벨로 전환하는 근거이며, Non-goals "디자인 리디자인 금지"의 예외 범주이다 (§2 Non-goals 예외 조항 참조).

## 8. 제외사항 (Exclusions — What NOT to Build)

- [HARD] **Nextra 런타임 컴포넌트 포팅 금지** — `Tabs`, `Cards`, `Steps`, `FileTree`, `Bleed` 등 사용 0건 컴포넌트는 Hextra 측에 대응 shortcode를 구현하지 않는다.
- [HARD] **Dead code 이식 금지** — `components/ui/{dialog,card,button,input}.tsx` 4개, `components/CodeBlock.tsx`, `components/navbar.tsx` 의 `GitHubStarBadge`, `lib/page-map.ts`, `app/[...meta].json`, `app/_content.mdx`, `app/debug-pagemap/` 는 이식 범위에서 제외한다.
- [HARD] **LaTeX 지원 제거** — `next.config.mjs` 의 `latex: true` 는 사용 0건이므로 Hextra 쪽에 KaTeX/MathJax 통합을 구현하지 않는다.
- [HARD] **`aggregateRating` JSON-LD 제거** — `ratingValue: "4.8"`, `ratingCount: "42"` 하드코딩 필드는 Google 정책 위반 소지로 이식 시 삭제한다 (REQ-DS-21).
- [HARD] **Playwright docs E2E 테스트 제거** — `playwright.config.ts` 는 moai-adk-go 테스트 전략과 맞지 않으므로 이관 대상에서 제외한다. Hugo 빌드 검증은 REQ-DS-30 + G1~G4 수동 체크로 대체한다.
- [HARD] **Vercel `regions: ["hkg1"]` 고정 제거** — Hugo 정적 사이트는 전세계 CDN 엣지 캐싱으로 운영되므로 region 고정은 제거한다.
- [HARD] **신규 문서 추가 금지** — 본 SPEC 범위 내에서 새 페이지/챕터/콘텐츠 생성 금지. 기존 219 페이지 이식에 한정.
- [HARD] **디자인 리디자인 금지** — Hextra 기본 테마 + 기존 moai-docs 브랜드 색/로고만 유지. 신규 디자인 시스템 도입 금지. (이모지 금지 정책에 따른 국기 이모지 → 텍스트 라벨 전환은 §2 Non-goals에서 명시한 예외로 허용.)
- [HARD] **en/ja/zh로의 ko 전용 섹션 번역 금지** — `contributing/`, `multi-llm/` 번역은 후속 SPEC으로 분리. 본 SPEC은 `_index.md` + `draft: true` + ko 리다이렉트 스켈레톤 생성에 한정 (REQ-DS-15).
- [HARD] **신규 Vercel 프로젝트 생성 금지** — 기존 프로젝트 ID `prj_EZaVdfE3gJeXVbizafBEECpniINP` 를 재바인딩하며, 다운타임 리스크 있는 프로젝트 신규 생성 접근은 배제한다 (REQ-DS-32a).
- [HARD] **moai-docs 전용 config 이관 제외 (D-022)** — moai-docs 레포의 `.moai/config/sections/gate.yaml`, `.moai/config/sections/memo.yaml`, `.moai/config/sections/observability.yaml` 3개 파일은 docs-site 런타임 운영과 무관하므로 본 SPEC의 이관 대상에서 완전히 제외한다. 필요 시 후속 SPEC에서 별도 판단한다.
- [HARD] **DNS 조작 사전 승인 금지** — REQ-DS-32a 범위에서 `adk.mo.ai.kr` DNS TTL 조정, NS/CNAME 변경은 **사전 불필요성 확인 없이 수행 금지**. AC-PRE-01 조사 결과 불필요로 확정되면 수행하지 않고, 필요로 확정되면 Vercel 지원팀 응답을 근거로 별도 변경 계획을 수립한다.

## 9. 가정 (Assumptions)

1. moai-adk-go 레포에 `docs-site/` 디렉토리가 사전 존재하지 않는다 (git subtree 최초 import 가능 상태).
2. `moai-adk-docs` Vercel 프로젝트의 관리 권한을 GOOS 계정에서 행사 가능하다.
3. Hugo v0.140+ 및 Hextra 최신 안정 버전이 마이그레이션 시점에 배포판으로 가용하다.
4. GitHub `moai-adk-docs` 레포의 `main` 브랜치가 현재 Vercel 프로덕션 배포와 동일한 커밋을 가리킨다.
5. DNS `adk.mo.ai.kr` 의 NS/CNAME 레코드 소유권이 GOOS에 있다 (AC-PRE-01 조사 후 필요 시 조정 가능).

## 10. 성공 기준 (Success Criteria — 요약)

상세 acceptance criteria는 `acceptance.md` 참조. 주요 성공 지표:

- 219 MDX 페이지 모두 Hugo 빌드에 포함되어 렌더링됨.
- 4개 locale의 섹션·페이지 수 일치 검증 통과 (ko 전용 예외 허용).
- 기존 URL 100% 보존 (404 0건).
- Vercel Preview 환경에서 4개 locale 홈페이지 및 상위 10개 페이지 수동 시각 검수 통과.
- Production 전환 후 48시간 무장애 운영. 위반 시 REQ-DS-35에 따라 즉시 롤백.
- `moai-adk-docs` 레포 archive 완료.

---

**Related SPECs**: SPEC-I18N-001 (absorbed, archived by this SPEC)
**Supersedes**: SPEC-I18N-001
