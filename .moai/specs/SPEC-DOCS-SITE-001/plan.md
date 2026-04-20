---
id: SPEC-DOCS-SITE-001
document: plan
version: 0.2.0
created: 2026-04-17
updated: 2026-04-20
---

# SPEC-DOCS-SITE-001 — 구현 계획 (Plan)

## 1. 개요

본 계획은 moai-docs 레포를 moai-adk-go 의 `docs-site/` 서브디렉토리로 흡수하고, Nextra → Hextra 전면 전환을 8개 Phase (Phase 0 완료 + Phase 1~8) 로 진행한다. 각 Phase는 명확한 산출물과 Gate 기준을 가지며, Gate G1 / G1.5 / G2 / G3 / G4 는 단계 이동을 위한 검증 체크포인트다. G4 (Vercel 프로덕션 전환) 는 **사용자 수동 승인 필수**.

## 2. 전제 결정사항 (User-Approved)

| ID | 결정 | 근거 문서 |
|----|------|-----------|
| D1 | `git subtree add --squash --prefix=docs-site` 로 이관. 39 MB `.git` → 단일 커밋 압축 | spec-project-diff.md § 7 D1 |
| D2 | 최신은 unversioned, 과거만 `v2.X/` 폴더. v2.12.0은 첫 스냅샷으로 생성하지 않음 (v2.13 시 스냅샷 생성) | spec-project-diff.md § 7 D2 |
| D3 | `middleware.ts` i18n 로직을 Vercel Edge Function으로 이식. cookie → Accept-Language → default ko 보존 | spec-project-diff.md § 7 D3 |
| D4 | G4 통과 + 48h 모니터링 후 `moai-adk-docs` archive. `archive/pre-migration-2026-04` 태그 + GitHub Archive 플래그 | spec-project-diff.md § 7 D4 |

## 3. Phase 개요

| Phase | 제목 | 주요 산출물 | Gate |
|-------|------|-------------|------|
| 0 | Discovery (완료) | migration-inventory.md, spec-project-diff.md | — |
| 1 | SPEC 문서화 (현재) | spec.md, plan.md, acceptance.md (3-file set) | G1 — Plan Auditor 승인 |
| 2 | Scaffold + Git Subtree | `docs-site/` 디렉토리 생성, squash import, Hugo/Hextra 뼈대 | **G1.5 — Scaffold Readiness** |
| 3 | 콘텐츠 마이그레이션 | 219 페이지 변환 완료, frontmatter 주입, Callout shortcode 치환 | G2 — `hugo build` 성공 + 페이지 수 검증 |
| 4 | 커스텀 컴포넌트 재현 | LanguageSelector/ClientNavbar/structured-data partial, SEO 메타, 이미지 최적화 | G2 (계속) |
| 5 | i18n Edge Function + 버전 스냅샷 자동화 | `api/i18n-detect.ts`, `scripts/docs-version-snapshot.go`, `scripts/docs-i18n-check.sh` | G3 — 로컬 hugo + Edge Function unit test 통과 |
| 6 | 프로젝트 문서 갱신 | CLAUDE.local.md §17 추가, product/structure/tech.md 업데이트, SPEC-I18N-001 archive 이관 | G3 (계속) |
| 7 | Vercel Preview → Production 전환 | Vercel 프로젝트 재바인딩, Preview 검증, Prod 전환 | **G4 — 사용자 수동 승인** |
| 8 | moai-adk-docs archive + 포스트모템 | archive 태그, Archive 플래그, README 전환 안내, Phase 0~7 lessons 기록 | — |

## 4. Phase별 상세 계획

### Phase 1 — SPEC 문서화 (현재 진행 중)

**목표**: SPEC-DOCS-SITE-001 3-file set 작성 및 plan-auditor 독립 감사 통과.

**Team role_profiles 매핑**:

- `manager-spec` (primary, foreground) — spec.md / plan.md / acceptance.md 작성
- `plan-auditor` (evaluator, foreground) — EARS 준수, 편향 검사, 완결성 감사

**산출물**:
- `.moai/specs/SPEC-DOCS-SITE-001/spec.md`
- `.moai/specs/SPEC-DOCS-SITE-001/plan.md` (본 문서)
- `.moai/specs/SPEC-DOCS-SITE-001/acceptance.md`

**Gate G1 기준**:
- 3-file set 존재 및 최소 길이 (spec.md 250+ LOC, plan.md 300+ LOC, acceptance.md 250+ LOC) — D-028 Iteration 4 통일
- EARS 패턴 적합성 15~36개 요구사항 (acceptance.md AC-G1-02와 일치)
- Exclusions 섹션 최소 10개 항목
- plan-auditor 독립 리뷰에서 수정 요청 3건 이하
- 결정 D1~D4가 spec.md 와 plan.md 양측에 명시

### Phase 2 — Scaffold + Git Subtree 이관

**목표**: `docs-site/` 디렉토리를 git subtree로 생성하고, Hugo/Hextra 뼈대를 놓는다.

**Team role_profiles 매핑**:
- `manager-git` (implementer, foreground) — `git subtree add --squash` 실행
- `expert-frontend` (architect, foreground) — Hugo `hugo.yaml`, `layouts/_default/baseof.html`, Hextra Hugo module 통합
- `expert-devops` (implementer, foreground) — `vercel.json` 최초 설정 (Framework Preset Hugo, build command)

**작업 단계**:
1. `git subtree add --squash --prefix=docs-site https://github.com/modu-ai/moai-adk-docs main` 실행 (D1)
2. 기존 Nextra/Next.js/Bun 잔재 파일 일괄 삭제 — REQ-DS-03 준수:
   - `package.json`, `bun.lock`, `biome.json`, `next.config.mjs`, `theme.config.tsx`, `app/`, `components/`, `lib/`, `middleware.ts`, `tsconfig.json`, `tailwind.config.ts`, `postcss.config.mjs`, `components.json`, `playwright.config.ts`, `mdx-components.tsx`, `.env.local`, `.gitignore` (Hugo 용으로 재작성)
3. Hugo 설정 파일 생성:
   - `docs-site/hugo.yaml` — baseURL, languages (ko/en/ja/zh), module.imports (Hextra)
   - `docs-site/go.mod` (**Hugo module system으로 잠금, D-015**) — Hextra를 `github.com/imfing/hextra` module로 import. git submodule 방식은 사용하지 않는다.
4. 기본 partial 디렉토리 스켈레톤: `layouts/partials/custom/`, `layouts/partials/head.html`
5. `static/` 디렉토리 유지 (favicon, og.png, manifest.json, robots.txt, llms.txt)
6. **서브트리 크기 실측 (Gap 2)**: `git subtree add` 직후 `du -sh .git/` 를 Phase 2 시작 전과 비교하여 `.moai/plans/DOCS-SITE/phase-2-subtree-size.md` 에 기록. 압축 목표 50 KB 대비 실측치를 Gate G1.5 증빙으로 첨부.
7. **빌드 시간 베이스라인 측정 (Gap 3, D-017)**: 콘텐츠가 없는 상태에서 `time hugo --minify --gc` 실행 결과를 기록. 5분 이상이면 Vercel 기본 빌드 timeout (10분)과의 여유 계산 및 대안 검토 지침을 `.moai/plans/DOCS-SITE/phase-2-build-baseline.md` 에 명시.

**산출물**:
- `docs-site/hugo.yaml` (baseURL=`https://adk.mo.ai.kr`, 4개 language 정의)
- `docs-site/go.mod` (Hextra Hugo module import)
- `docs-site/.gitignore` (public/, resources/ 등)
- 이관 완료 단일 squash 커밋
- `.moai/plans/DOCS-SITE/phase-2-subtree-size.md`, `.moai/plans/DOCS-SITE/phase-2-build-baseline.md`

**Gate G1.5 — Scaffold Readiness (Phase 2 → Phase 3, D-003 신설)**:
- `docs-site/` 디렉토리 존재
- `git log --oneline -1 -- docs-site/` 가 squash 커밋 단일 기록을 보여줌 (다수 커밋 누락 방지)
- `docs-site/hugo.yaml` 존재 + `baseURL` 및 `languages.{ko,en,ja,zh}` 정의 확인
- Nextra 잔재 0건 — `docs-site/package.json`, `docs-site/bun.lock`, `docs-site/next.config.mjs`, `docs-site/theme.config.tsx`, `docs-site/app/`, `docs-site/components/`, `docs-site/lib/`, `docs-site/middleware.ts` 모두 **부재** 확인
- `cd docs-site && hugo server` 가 빈 콘텐츠 상태로 에러 없이 기동 (프로세스 시작 후 `listening on` 로그 확인, 5초 안정 상태 유지 후 종료)
- subtree 크기 실측 + 빌드 시간 베이스라인 기록 문서 2종 존재

### Phase 3 — 콘텐츠 마이그레이션 (219 페이지)

**목표**: 219 MDX 페이지를 Hugo Markdown으로 변환하고 Hextra 호환 shortcode/frontmatter 를 주입한다.

**Team role_profiles 매핑**:
- `manager-ddd` (implementer, worktree isolation) — Go 스크립트 작성 및 실행
- `expert-backend` (implementer, worktree isolation) — `scripts/inject-frontmatter.go`, `scripts/convert-meta-to-frontmatter.go` TDD 개발
- `manager-docs` (reviewer, foreground) — 변환 결과 샘플 검수 (locale별 10페이지)

**작업 단계**:
1. **Go 스크립트 TDD 개발** (REQ-DS-08):
   - `scripts/convert-meta-to-frontmatter.go` — 38개 `_meta.ts` 파싱 → Hugo `_index.md` frontmatter `weight`, `title`, `_build.list/render` 생성
   - `scripts/inject-frontmatter.go` — 219 페이지에 YAML frontmatter 일괄 주입 (title from H1, weight from `_meta.ts`, draft: false)
   - `scripts/convert-callouts.go` — 735건 `<Callout type="...">` → `{{< callout type="..." >}}` 치환 (절대 보존, REQ-DS-05)
2. **단계적 적용**:
   - ko locale 먼저 변환 (63 페이지)
   - en/ja/zh 동시 변환 (각 52 페이지)
   - ko 전용 `contributing/`, `multi-llm/` 스켈레톤을 en/ja/zh 에 생성. 각 스켈레톤은 `_index.md` 한 개만 포함하며, frontmatter `draft: true` + `aliases: ["/ko/<section>/"]` 로 ko 원본으로 리다이렉트한다 (REQ-DS-15).
3. **Mermaid 569건 확인** — 코드 펜스 포맷 그대로 유지, Hextra 내장 지원 활성화 (`hugo.yaml` 의 `params.mermaid` 플래그)
4. **Dead code 제거 최종 확인** — REQ-DS-03 기준 검증

**산출물**:
- `scripts/convert-meta-to-frontmatter.go` (Go + unit tests)
- `scripts/inject-frontmatter.go` (Go + unit tests)
- `scripts/convert-callouts.go` (Go + unit tests)
- `docs-site/content/{ko,en,ja,zh}/` 전량 변환 완료
- 변환 로그 `.moai/plans/DOCS-SITE/phase-3-migration-log.md`

**Gate G2 기준**:
- `cd docs-site && hugo --minify` 성공 (exit 0), zero warnings
- 빌드 결과 HTML 페이지 수 = 219 (locale별 카운트 일치: ko 63, en/ja/zh 각 52)
- `grep -r "import { Callout }" docs-site/content/` → 0건
- `grep -r "from 'nextra" docs-site/content/` → 0건
- Callout shortcode 정확히 735건 (REQ-DS-05 절대 보존, AC-G2-04 대응)
- Mermaid 블록 정확히 569건 (REQ-DS-09 절대 보존, AC-G2-06 대응)
- 4개 locale 각 5페이지 수동 샘플링 시각 검수 — Callout, Mermaid, 코드블록 렌더링 정상

### Phase 4 — 커스텀 컴포넌트 재현

**목표**: LanguageSelector, ClientNavbar, structured-data 등 Hugo partial로 재현 + SEO 메타 구성 + OG 이미지 최적화.

**Team role_profiles 매핑**:
- `expert-frontend` (implementer, worktree isolation) — Hugo partial 작성
- `expert-backend` (implementer, foreground) — JSON-LD data file + partial 연동
- `expert-devops` (implementer, foreground) — og.png 7.5 MB → 500 KB 이하 압축

**작업 단계**:
1. `layouts/partials/custom/language-switch.html` — 4개 locale 드롭다운. **이모지 금지 정책 (coding-standards.md)에 따라 국기 이모지 대신 "KO / EN / JA / ZH" 텍스트 라벨 사용** (Non-goals 예외 조항 적용, spec.md §2 참조).
2. `layouts/partials/custom/navbar-end.html` — GitHub 아이콘 + shields.io stars 뱃지 (빌드타임 정적 링크)
3. `layouts/partials/head.html` — 24개 meta 태그 이관 (favicon, viewport, OG, Twitter, hreflang, canonical)
4. `layouts/partials/structured-data.html` + `data/structured_data.yaml` — Organization/SoftwareApplication/WebSite/TechArticle 4종 JSON-LD (REQ-DS-21 준수: `aggregateRating` 제거)
5. `static/og.png` 7.5 MB → WebP 또는 최적화 PNG 500 KB 이하 (AC-G4-05에서 크기 자동 검증)
6. `vercel.json` redirect 2건 보존 (`/moai-rank/*` → `/`)
7. `i18n/{ko,en,ja,zh}.toml` — per-locale 텍스트 (toc.title, editLink, feedback, footer)
8. **SEO 영향 사전 조사 (D-019, AC-PRE-02)**: Google Search Console에서 `aggregateRating` 기반 Rich Results 노출 이력을 조회하고, 제거 시 예상 트래픽 영향도를 `.moai/plans/DOCS-SITE/phase-4-seo-impact.md` 에 기록. Rich Results 의존 항목 0건이면 "영향 없음" 확정, 존재 시 Phase 7 모니터링 지표에 Rich Results 노출량을 추가 등록.

**산출물**:
- `docs-site/layouts/partials/custom/*.html` (language-switch, navbar-end)
- `docs-site/layouts/partials/head.html`
- `docs-site/layouts/partials/structured-data.html`
- `docs-site/data/structured_data.yaml`
- `docs-site/static/og.png` (최적화본 ≤ 500 KB)
- `docs-site/vercel.json` (redirect 2건 + Framework Preset Hugo)
- `docs-site/i18n/{ko,en,ja,zh}.toml`
- `.moai/plans/DOCS-SITE/phase-4-seo-impact.md`

**Gate**: G2 계속. Phase 4 완료 시 Hugo 빌드 결과가 기존 Nextra 대비 시각적 동등성 (color/layout/logo) 유지. (픽셀 diff는 도입하지 않고, locale당 5페이지 수동 스크린샷 비교 수준을 기준으로 한다. 정량 기준 부재는 AC-G2-07 체크리스트로 대체.)

### Phase 5 — i18n Edge Function + 버전 스냅샷 자동화

**목표**: `middleware.ts` 의 164 LOC 로직을 Vercel Edge Function으로 이식하고, 릴리스 스냅샷 자동화 Go 스크립트를 작성한다.

**Team role_profiles 매핑**:
- `expert-devops` (implementer, worktree isolation) — Edge Function TypeScript 작성
- `expert-backend` (implementer, worktree isolation) — `scripts/docs-version-snapshot.go` Go 스크립트
- `expert-backend` (implementer, foreground) — `scripts/docs-i18n-check.sh` CI 스크립트 + `scripts/docs-forbidden-check.sh`

**작업 단계**:
1. `docs-site/api/i18n-detect.ts` — Vercel Edge Function:
   - **Edge runtime 명시 (D-007)**: 파일 상단에 `export const config = { runtime: 'edge' };` (또는 `export const runtime = 'edge';`) 선언. 없으면 Vercel은 기본적으로 Serverless Function (Node.js) 으로 배포하므로 반드시 명시.
   - **Platform constraint 사전 검토 (D-007)**: Edge 실행 환경 제약 확인 — 실행 시간 ≤ 50ms (cold start 포함), 메모리 ≤ 4MB, Edge 호환 npm 패키지만 사용 (Node.js `fs`, `path` 등 비호환). 현재 middleware.ts 164 LOC는 path/header/cookie 단순 로직이므로 제약 내 수용 가능하나, 변환 과정에서 Node.js 전용 API 의존이 없는지 사전 확인하고 `.moai/plans/DOCS-SITE/phase-5-edge-constraint-check.md` 에 기록.
   - Path prefix 추출 → `/ko`, `/en`, `/ja`, `/zh` 이면 통과 + cookie set (REQ-DS-13)
   - Prefix 없으면 cookie > Accept-Language > `ko` 우선순위 판정 → 302 redirect
   - `/api/`, `/_next/`, `/static/`, 확장자 경로는 bypass (REQ-DS-14)
   - `vercel.json` 에 rewrite 규칙 등록
2. `scripts/docs-version-snapshot.go` (REQ-DS-17):
   - 입력: `<previous-version>` (예: `v2.12`)
   - 동작: **이전 릴리스 태그 커밋 기준**으로 (HEAD of main 아님, D-010) `content/{ko,en,ja,zh}/` 를 `content/{locale}/<previous-version>/` 로 복사, 기존 `v*/` 서브디렉토리는 스킵
   - Major/Minor 만 대응, Patch 는 스크립트가 감지하여 스킵 (REQ-DS-18)
   - **배너 구현 메커니즘 (D-023)**: `v2.X/` 접두사 경로에 대해 상단 배너 partial 자동 렌더. 구현 방식:
     - `layouts/partials/version-banner.html` 생성 — `.Page.File.Path` 에서 `v[0-9]+\.[0-9]+/` 접두사 존재 여부 판정 (Hugo `hasPrefix` 또는 `findRE` 사용).
     - 배너 텍스트: "Viewing v2.X. Go to latest" + latest URL 링크.
     - 대체 latest URL 산출: 현재 `.File.Path` 에서 `v<version>/` prefix 를 제거한 unversioned 경로.
     - `baseof.html` 에서 조건부로 `version-banner` partial 호출.
3. `scripts/docs-i18n-check.sh` (REQ-DS-31):
   - 4개 locale 파일 개수 일치 검사 (ko 전용 예외 허용)
   - 경로 동형성 검사
   - frontmatter title 비어있지 않음
   - MoAI 용어집 준수 (MoAI-ADK 불번역 확인)
4. `scripts/docs-forbidden-check.sh` (REQ-DS-29, 신규):
   - 금지 URL 패턴 (docs.moai-ai.dev, adk.moai.com, adk.moai.kr) 0건 검증
   - Mermaid `flowchart LR` / `graph LR` 방향 0건 검증
   - 문서 본문 이모지 0건 검증
5. `scripts/validate-meta-schema.go` (Gap 7): `_meta.ts` 변환 산출물 `_index.md` frontmatter 가 JSON Schema (title: string 필수, weight: integer 필수, `_build.list`/`_build.render` enum) 를 만족하는지 CI에서 자동 검증.

**산출물**:
- `docs-site/api/i18n-detect.ts` (Vercel Edge Function, `runtime: 'edge'` 명시)
- `scripts/docs-version-snapshot.go` + unit tests
- `scripts/docs-i18n-check.sh`
- `scripts/docs-forbidden-check.sh`
- `scripts/validate-meta-schema.go`
- `docs-site/layouts/partials/version-banner.html`
- `.github/workflows/docs-site.yml` (CI 통합)
- `.moai/plans/DOCS-SITE/phase-5-edge-constraint-check.md`

**Gate G3 기준**:
- `cd docs-site && hugo --minify` 성공 (zero warnings)
- Edge Function 로컬 Vercel dev 환경에서 cookie/Accept-Language/default 3시나리오 통과
- Edge Function 파일 상단 `runtime: 'edge'` 선언 검증 (grep)
- `bash scripts/docs-i18n-check.sh` 성공 (locale 불일치 0건, ko 전용 섹션 예외 정상 인식)
- `bash scripts/docs-forbidden-check.sh` 성공 (금지 URL / 방향 / 이모지 0건)
- `go test ./scripts/validate-meta-schema/...` 통과

### Phase 6 — 프로젝트 문서 갱신

**목표**: `CLAUDE.local.md` §17 신설, `.moai/project/*.md` 3종 갱신, SPEC-I18N-001 archive.

**Team role_profiles 매핑**:
- `manager-docs` (implementer, foreground) — CLAUDE.local.md, project docs 편집
- `manager-spec` (implementer, foreground) — SPEC-I18N-001 archive 메타 추가

**작업 단계**:
1. `CLAUDE.local.md` §17 추가 (REQ-DS-24):
   - §17.1 URL 표준 & 블랙리스트
   - §17.2 Markdown/Hextra 작성 규칙 (Mermaid TD only 유지)
   - §17.3 4개국어 동기화 규칙 (ko canonical, 48h/72h 번역 deadline, translation_status: pending frontmatter)
   - §17.4 버전 스냅샷 정책 (Major/Minor YES, Patch NO)
   - §17.5 실행 주체 (manager-docs primary, plan-auditor verification, manager-git snapshot)
   - §17.6 빌드/배포 체크리스트 + Vercel 프로젝트 바인딩 정보
2. `.moai/project/product.md` 갱신 (REQ-DS-25) — "Public Documentation Site" Core Feature 추가, Target Audience 에 Documentation Readers 추가
3. `.moai/project/structure.md` 갱신 (REQ-DS-26) — `docs-site/` 최상위 디렉토리 구조, Hugo 빌드 파이프라인, Vercel 배포 토폴로지
4. `.moai/project/tech.md` 갱신 (REQ-DS-27) — Documentation Site Stack 섹션 추가, 제거된 스택 (Nextra/Next.js/MDX/Bun/...) 명시
5. SPEC-I18N-001 archive 이관 (REQ-DS-28):
   - `.moai/specs/SPEC-I18N-001/` 디렉토리 생성
   - 원본 3개 파일 복사 + frontmatter 에 `status: archived`, `supersede_by: SPEC-DOCS-SITE-001` 추가

**산출물**:
- `CLAUDE.local.md` (+ §17 약 125 LOC 추가)
- `.moai/project/product.md`, `structure.md`, `tech.md` 갱신본
- `.moai/specs/SPEC-I18N-001/{spec,plan,acceptance}.md` (archived 메타)

**Gate**: G3 계속. 모든 편집물은 plan-auditor 에게 위임된 consistency check 통과.

### Phase 7 — Vercel Preview → Production 전환 [사용자 수동 승인 게이트 G4]

**목표**: Vercel 프로젝트 재바인딩을 통해 `adk.mo.ai.kr` 을 Hugo 빌드로 무중단 전환.

**Team role_profiles 매핑**:
- `expert-devops` (lead, foreground) — Vercel Dashboard 조작, DNS 모니터링
- `manager-git` (support, foreground) — 긴급 롤백 준비
- **GOOS** — G4 수동 승인

**작업 단계** (순차 진행, **자동 실행 금지**):
1. **사전 조사 및 준비** (AC-PRE-01 / AC-PRE-02 / AC-PRE-03 선행 필수):
   - **DNS 필요성 조사 (AC-PRE-01, D-004)**: Vercel 공식 문서 및 지원팀 응답을 근거로 "기존 프로젝트의 Git source 변경이 DNS 레코드 변경을 동반하는지" 확인. 불필요 확정 시 DNS 조작 0건으로 진행. 필요 확정 시 Vercel 지원팀 응답에 근거한 별도 변경 계획을 `.moai/plans/DOCS-SITE/phase-7-dns-assessment.md` 에 기록 후 진행.
   - **SEO 영향 확정 (AC-PRE-02, D-019)**: Phase 4에서 작성한 `.moai/plans/DOCS-SITE/phase-4-seo-impact.md` 리뷰.
   - **Vercel 무중단 근거 확인 (AC-PRE-03, Gap 1)**: "기존 Vercel project의 Git source 변경 시 도메인 바인딩이 유지되어 무중단 전환이 가능하다"는 Vercel 공식 문서 또는 지원팀 답변을 캡처하여 `.moai/plans/DOCS-SITE/phase-7-zero-downtime-evidence.md` 에 기록.
   - **Nextra baseline 측정 (D-013)**: 현재 프로덕션 `adk.mo.ai.kr` 주요 페이지의 Lighthouse Performance/SEO/Accessibility 점수를 측정하여 `.moai/plans/DOCS-SITE/phase-7-preview-verification.md` baseline 필드에 기록. AC-G4-07 비교 기준.
   - `moai-adk-go` main 브랜치에 Phase 1~6 결과 반영된 커밋 확정.
2. **Vercel Project 재바인딩** (REQ-DS-32a, G4 **이전** 실행):
   - Git 연결: `moai-adk-docs` → `modu-ai/moai-adk-go` 변경
   - Root Directory: `docs-site/` 설정
   - Framework Preset: Hugo 선택
   - Ignored Build Step: `git diff --quiet HEAD^ HEAD ./docs-site/ || exit 1`
   - Node.js version: (Hugo 는 Node 불요 — 기본값 유지)
   - DNS 변경 없음 (AC-PRE-01 결과 참조).
3. **Preview URL 검증** (REQ-DS-33):
   - acceptance.md § G4 체크리스트 전수 수행
   - 4개 locale 홈 + 각 locale 주요 10페이지 수동 시각 검수
   - Mermaid 렌더링 light/dark 모드 확인
   - Edge Function i18n 리다이렉트 3시나리오 확인
   - Edge runtime 응답 헤더 검증 (`curl -I` 의 `x-vercel-cache` / `server` 헤더)
   - Lighthouse score 측정 + Nextra baseline 대비 Performance -5 이내 확인
4. **G4 사용자 수동 승인** — GOOS 가 Preview 품질 만족 확인 후 AskUserQuestion 응답으로 진행 결정
5. **Production 프로모션** (REQ-DS-32b, G4 **이후** 실행):
   - Vercel Dashboard 에서 Preview → Production promote
   - DNS 변경 없음 (도메인 바인딩 유지, 프로젝트 내부 전환만)
6. **사후 모니터링 (48h)**:
   - Vercel Analytics 기준 4xx/5xx 에러율, LCP, TTFB 추적
   - 사용자 제보 채널 (Discord) 모니터링
   - REQ-DS-35 롤백 기준 위반 시 즉시 Vercel Git source 를 `modu-ai/moai-adk-docs` 로 되돌리고 `.moai/plans/DOCS-SITE/phase-7-rollback.md` 기록

**산출물**:
- Vercel Project 재설정 완료 (Dashboard 변경 기록)
- Preview URL 검증 리포트 `.moai/plans/DOCS-SITE/phase-7-preview-verification.md` (Nextra baseline 필드 포함)
- `.moai/plans/DOCS-SITE/phase-7-dns-assessment.md`
- `.moai/plans/DOCS-SITE/phase-7-zero-downtime-evidence.md`
- 48h 모니터링 로그 `.moai/plans/DOCS-SITE/phase-7-monitoring-48h.md`
- (롤백 발생 시) `.moai/plans/DOCS-SITE/phase-7-rollback.md`

**Gate G4 기준 (사용자 수동 승인)**:
- AC-PRE-01 / AC-PRE-02 / AC-PRE-03 사전 조사 완료
- Preview 에서 acceptance.md § G4 체크리스트 100% 통과
- 4개 locale 홈 + 주요 페이지 시각 검수 이상 없음
- Edge Function i18n 3시나리오 (cookie / Accept-Language / default) 정상 작동 + Edge runtime 헤더 확인
- Lighthouse Performance ≥ 85 AND Nextra baseline 대비 -5 이내, SEO ≥ 95, Accessibility ≥ 90
- GOOS 최종 승인

### Phase 8 — moai-adk-docs archive + 포스트모템

**목표**: Phase 7 G4 통과 + 48h 모니터링 완료 조건 만족 시 원 레포 archive 및 lessons 기록.

**Team role_profiles 매핑**:
- `manager-git` (implementer, foreground) — archive 태그, Archive 플래그
- `manager-docs` (implementer, foreground) — archived 레포 README 업데이트
- `manager-spec` (reviewer, foreground) — Phase 0~7 lessons 기록

**작업 단계** (REQ-DS-34, 주체: Phase 8 automation = manager-git + expert-devops):
1. 48h 모니터링 게이트 통과 확인 (P1/P2 incidents 0건, REQ-DS-35 롤백 미발생)
2. `moai-adk-docs` 레포에 `archive/pre-migration-2026-04` 태그 발행
3. GitHub 레포 설정 → Archive this repository 플래그 활성화
4. archived 레포의 `README.md` 첫 줄에 "This repository has been archived and migrated to [moai-adk-go/docs-site](https://github.com/modu-ai/moai-adk-go/tree/main/docs-site). Historical content preserved for reference." 추가
5. `.moai/plans/DOCS-SITE/phase-8-postmortem.md` 작성 — Phase 0~7 에서 발생한 예상/돌발 이슈, 학습 사항, 후속 개선점 기록
6. `.moai/state/lessons.md` 에 본 마이그레이션 lessons 3~5건 추가 (번역 체인, Hugo 모듈 통합, Edge Function 경계 등)

**산출물**:
- `moai-adk-docs` archive/pre-migration-2026-04 태그 + Archive 플래그
- archived README.md (이관 안내)
- `.moai/plans/DOCS-SITE/phase-8-postmortem.md`
- `.moai/state/lessons.md` 업데이트

## 5. Mermaid 전략 (Mermaid Strategy)

**결정**: Hextra 는 Mermaid 를 **클라이언트 사이드 (브라우저 내 JavaScript)** 로 렌더링한다 (공식 `imfing.github.io/hextra/docs/guide/diagrams/` 확인).

**근거**:
- Hextra 공식 문서: "a JavaScript based diagramming and charting tool that takes Markdown-inspired text definitions and creates diagrams dynamically in the browser"
- 서버 사이드 SVG 프리렌더링은 Hextra 기본 제공하지 않음

**본 SPEC 의 결정**:
- 초기 이관은 클라이언트 사이드 렌더링 수용 (REQ-DS-10).
- 569개 × 4 locale = 2,276개 다이어그램의 초기 로드 영향은 Hextra lazy-load 설정으로 완화 (REQ-DS-11).
- 성능 저하가 Lighthouse Performance < 85 로 확인되면 Phase 8 후속 개선 과제로 등록, 빌드타임 SVG 프리렌더 파이프라인 신설 검토.

## 6. Risk Register (R1~R10 from migration-inventory.md)

| ID | 리스크 | 심각도 | 완화책 | 담당 Phase |
|----|--------|--------|--------|------------|
| R1 | middleware.ts Edge 로직 (164 LOC) 재현 실패 | High | Vercel Edge Function 으로 이식 (D3). `runtime: 'edge'` 명시 + platform constraint 사전 확인 + 3시나리오 단위 테스트 | Phase 5 |
| R2 | 219 MDX 페이지 frontmatter 수동 주입 휴먼 에러 | High | Go 스크립트 TDD (`inject-frontmatter.go`), locale별 단위 검증 | Phase 3 |
| R3 | Mermaid 569건 × 4 locale 클라이언트 렌더 성능 저하 | Medium | Hextra lazy-load, Lighthouse 측정 후 SVG 프리렌더 여부 결정 | Phase 3, 7 |
| R4 | ko 전용 섹션 (contributing, multi-llm) 링크 체커 실패 | Medium | en/ja/zh 스켈레톤 (`_index.md` + `draft: true` + `aliases`) 생성, Hextra fallback 활용 | Phase 3 |
| R5 | structured-data `aggregateRating` 하드코딩 (Google 정책 위반) | Low | 마이그레이션 시 필드 제거 (REQ-DS-21), Phase 4 SEO 영향 조사 | Phase 4 |
| R6 | Dead code 이관으로 작업량 부풀리기 | Medium | REQ-DS-03 + Phase 2 삭제 리스트 준수, CI 검증 | Phase 2 |
| R7 | Vercel `regions: ["hkg1"]` 고정이 전세계 CDN 최적화 저해 | Low | Hugo 정적 사이트 특성 상 region 고정 제거, 기본 CDN 엣지 활용 | Phase 4 |
| R8 | `_meta.ts` TypeScript 타입 안전성 상실 | Medium | `convert-meta-to-frontmatter.go` TDD + JSON Schema 검증 CI (`validate-meta-schema.go`) | Phase 3, 5 |
| R9 | og.png 7.5 MB 용량 (소셜 공유 지연) | Medium | WebP 변환 또는 PNG 압축으로 500 KB 이하 달성 + AC-G4-05 자동 검증 | Phase 4 |
| R10 | `sitemap.ts` 하드코딩 경로 36개 중 `/claude-code/*` 7개 불일치 | Low | Hugo 자동 sitemap 으로 전환, 외부 백링크 301 redirect 추가 | Phase 4 |
| R11 (신규) | 48h 모니터링 기간 중 장애 발생 시 의사결정 지연 | High | REQ-DS-35 자동 롤백 경로 정의, Vercel Dashboard 원복 수순 사전 드릴 | Phase 7 |
| R12 (신규) | Hextra 빌드 시간이 Vercel timeout 초과 | Medium | Phase 2에서 빈 콘텐츠 baseline 측정, Phase 3 완료 후 재측정 | Phase 2, 3 |

## 7. 체크포인트 Gate 세부 기준

### G1 — SPEC Plan Auditor Gate (Phase 1 → Phase 2)

- 3-file set 존재 및 최소 길이 준수 (spec.md 250+ / plan.md 300+ / acceptance.md 250+ LOC) — D-026 Iteration 3 통일 조정
- **EARS 요구사항 15~36개** (D-001/D-008 통일, acceptance.md AC-G1-02와 일치. REQ-DS-32 → 32a/32b 분할 + REQ-DS-35 신규로 총 36건 수용), Exclusions 최소 10개
- D1~D4 가 spec.md 와 plan.md 양측에 명시
- plan-auditor 독립 리뷰 수정 요청 ≤ 3건

### G1.5 — Scaffold Readiness Gate (Phase 2 → Phase 3, D-003 신설)

- `docs-site/` 디렉토리 존재 및 squash 커밋 단일 기록
- `docs-site/hugo.yaml` baseURL + 4개 language 정의
- Nextra 잔재 0건 (package.json / bun.lock / next.config.mjs / theme.config.tsx / app/ / components/ / lib/ / middleware.ts)
- `cd docs-site && hugo server` 빈 콘텐츠 상태에서 에러 없이 기동
- subtree 크기 실측 + 빌드 시간 베이스라인 문서 2종 존재

### G2 — Hugo Build Gate (Phase 3/4 → Phase 5, D-020 표기 수정)

- `cd docs-site && hugo --minify --gc` exit 0, zero warnings
- 빌드 페이지 수 = 219 (± ko 전용 섹션 보정)
- Callout 정확히 735건 (REQ-DS-05 절대 보존, AC-G2-04)
- Mermaid 정확히 569건 (REQ-DS-09 절대 보존, AC-G2-06)
- 샘플 검수 통과 (locale별 5페이지 × 4)
- Dead code 및 Nextra 잔재 CI 검증 통과 (REQ-DS-03)

### G3 — Automation Readiness Gate (Phase 5/6 → Phase 7)

- Edge Function i18n 3시나리오 로컬 테스트 통과 + `runtime: 'edge'` 선언 검증
- `scripts/docs-i18n-check.sh` 성공 (locale 불일치 0건)
- `scripts/docs-forbidden-check.sh` 성공 (금지 URL / 방향 / 이모지 0건)
- `scripts/docs-version-snapshot.go` unit test 통과 (TestMinorRelease / TestPatchRelease / TestMajorRelease / TestSkipExistingVersions)
- `scripts/validate-meta-schema.go` 통과
- CLAUDE.local.md §17 추가 완료, project docs 3종 갱신 완료
- SPEC-I18N-001 archive 이관 완료

### G4 — Production Cutover Gate [사용자 수동 승인]

- AC-PRE-01 / AC-PRE-02 / AC-PRE-03 사전 조사 완료
- acceptance.md § G4 체크리스트 100% 통과
- 4개 locale 홈 + 주요 페이지 시각 검수 이상 없음
- Edge Function 실 환경 3시나리오 통과 + Edge runtime 헤더 확인
- Lighthouse Performance ≥ 85 (Desktop) / ≥ 75 (Mobile) **AND** Nextra baseline 대비 -5 이내, SEO ≥ 95, Accessibility ≥ 90
- GOOS 최종 승인 (AskUserQuestion)

## 8. 기술적 접근 (Technical Approach) 요약

- **Git 통합**: `git subtree --squash` 로 이력 압축 후 단일 커밋 import (D1).
- **Hugo 구성**: multilingual mode, `defaultContentLanguage: ko`, language codes ko/en/ja/zh.
- **Hextra 통합**: **Hugo module system (`docs-site/go.mod` 에 Hextra module import) — git submodule 사용 금지 (D-015 잠금)**. 버전 잠금 명시성 + CI 환경 단순성.
- **콘텐츠 변환**: Go 스크립트 3개 (convert-meta / inject-frontmatter / convert-callouts) + 1개 schema validator (validate-meta-schema) 로 219 파일 일괄 처리, TDD.
- **i18n 런타임**: Vercel Edge Function (TypeScript, `runtime: 'edge'` 명시) + `vercel.json` rewrites. Hugo 정적 빌드 결과는 순수 HTML/CSS/JS.
- **버전 스냅샷**: 릴리스 태그 시점에 `scripts/docs-version-snapshot.go` 호출 (Major/Minor만, 이전 릴리스 태그 커밋 기준).
- **검색**: FlexSearch는 본 SPEC에서 **기본 비활성** (`search: false` 유지). 활성화 여부는 Phase 8 이후 후속 SPEC 판단.
- **CI/CD**: GitHub Actions (`.github/workflows/docs-site.yml`) + Vercel 자동 배포. `Ignored Build Step` 으로 docs-site 외 변경 시 빌드 스킵. CI 단계에서 `docs-i18n-check.sh`, `docs-forbidden-check.sh`, `validate-meta-schema.go` 실행.

## 9. 후속 작업 (Not in this SPEC)

- en/ja/zh 의 `contributing/`, `multi-llm/` 섹션 번역 (별도 SPEC)
- Mermaid 서버 사이드 프리렌더 파이프라인 (Lighthouse 결과에 따라 Phase 8 이후 검토)
- 문서 검색 FlexSearch 최적화 — 본 SPEC에서 `search: false` 로 고정. 활성화 판단은 후속 SPEC.
- AI Agency 문서 신규 섹션 추가
- moai-docs의 `gate.yaml` / `memo.yaml` / `observability.yaml` 3개 설정 이관 여부 재검토

## 10. Out-of-Plan 조정

본 계획의 Phase 순서는 기술적 의존 관계가 강하므로 순차 진행을 원칙으로 한다. 다만 다음 경우 조정 가능:

- Phase 3 콘텐츠 변환과 Phase 4 partial 구현은 파일 충돌 분리 가능하면 병렬 실행 (worktree isolation 활용).
- Phase 5 Edge Function 개발과 Phase 6 프로젝트 문서 갱신은 완전 독립이므로 병렬 가능.
- Phase 7 G4 승인 후 Phase 8 은 48h 모니터링 완료를 대기 (게이트 조건 필수). REQ-DS-35 롤백 발생 시 Phase 8 은 중단되고 원인 분석 단계로 전환.
