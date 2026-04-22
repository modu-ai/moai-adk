---
id: SPEC-DOCS-SITE-001
document: acceptance
version: 0.2.0
created: 2026-04-17
updated: 2026-04-20
---

# SPEC-DOCS-SITE-001 — 인수 기준 (Acceptance Criteria)

본 문서는 SPEC-DOCS-SITE-001 의 각 Phase 및 요구사항이 완료되었음을 판정하기 위한 Given-When-Then 시나리오, 자동 검증 방법, 수동 검증 체크리스트를 정의한다.

---

## Gate G1 — SPEC Plan Auditor Gate (Phase 1 완료 판정)

### AC-G1-01 — 3-file set 존재 및 최소 길이

**Given** `.moai/specs/SPEC-DOCS-SITE-001/` 디렉토리에
**When** 3개 파일 (spec.md, plan.md, acceptance.md) 이 존재하고
**Then** 각 파일이 최소 LOC 기준 (spec.md ≥ 250, plan.md ≥ 300, acceptance.md ≥ 250) 을 만족해야 한다.

**자동 검증**:
```bash
test -f .moai/specs/SPEC-DOCS-SITE-001/spec.md
test -f .moai/specs/SPEC-DOCS-SITE-001/plan.md
test -f .moai/specs/SPEC-DOCS-SITE-001/acceptance.md
test "$(wc -l < .moai/specs/SPEC-DOCS-SITE-001/spec.md)" -ge 250
test "$(wc -l < .moai/specs/SPEC-DOCS-SITE-001/plan.md)" -ge 300
test "$(wc -l < .moai/specs/SPEC-DOCS-SITE-001/acceptance.md)" -ge 250
```

### AC-G1-02 — EARS 패턴 및 Exclusions 준수

**Given** spec.md 의 요구사항 섹션에서
**When** `REQ-DS-` 프리픽스를 가진 요구사항을 grep 으로 추출하면
**Then** 15 ≤ 요구사항 개수 ≤ 36 이고 (D-001 상한 상향, plan.md §7 G1과 일치. REQ-DS-32의 32a/32b 분할과 신규 REQ-DS-35 추가로 총 36건), 각 요구사항은 EARS 패턴 (Ubiquitous / Event-driven / State-driven / Optional / Unwanted) 중 하나로 분류되어야 한다.

**자동 검증**:
```bash
grep -c "^\*\*REQ-DS-" .moai/specs/SPEC-DOCS-SITE-001/spec.md | awk '{if($1 >= 15 && $1 <= 36) exit 0; else exit 1}'
grep -c "^### " .moai/specs/SPEC-DOCS-SITE-001/spec.md   # 섹션 존재
grep -q "## 8. 제외사항" .moai/specs/SPEC-DOCS-SITE-001/spec.md
```

Exclusions 섹션 항목 수 ≥ 10.

### AC-G1-03 — D1~D4 결정사항 이중 명시

**Given** spec.md 의 § 5 Reference Documents 섹션과 plan.md 의 § 2 전제 결정사항 섹션에서
**When** D1, D2, D3, D4 각 결정사항을 확인하면
**Then** 네 결정이 모두 양쪽 문서에서 정확히 일관되게 명시되어야 한다.

**수동 검증**: spec.md 와 plan.md 를 나란히 읽어 D1~D4 문구 일치 확인.

### AC-G1-04 — Plan Auditor 독립 감사 통과

**Given** plan-auditor subagent 가
**When** 3-file set 을 읽고 편향 검사, EARS 준수, 완결성 감사를 수행하면
**Then** 수정 요청 개수가 3건 이하여야 한다.

**수동 검증**: plan-auditor 실행 결과 보고서에서 수정 요청 개수 확인.

---

## Gate G1.5 — Scaffold Readiness Gate (Phase 2 완료 판정, D-003 신설)

### AC-G1.5-01 — docs-site 디렉토리 + squash 커밋 단일성

**Given** Phase 2 `git subtree add --squash` 완료 직후
**When** `docs-site/` 디렉토리 존재 여부와 git log 을 확인하면
**Then** 디렉토리가 존재하고, 해당 경로에 대한 squash 커밋이 **단일 커밋**으로 기록되어야 한다.

**자동 검증**:
```bash
test -d docs-site
test "$(git log --oneline -- docs-site/ | wc -l)" -eq 1
```

### AC-G1.5-02 — hugo.yaml 필수 필드 존재

**Given** Phase 2 Hugo 설정 생성 완료 상태에서
**When** `docs-site/hugo.yaml` 을 파싱하면
**Then** `baseURL` 키 존재 + `languages.ko`, `languages.en`, `languages.ja`, `languages.zh` 4개 언어 엔트리가 정의되어야 한다.

**자동 검증**:
```bash
test -f docs-site/hugo.yaml
grep -q "^baseURL:" docs-site/hugo.yaml
grep -qE "^  ko:" docs-site/hugo.yaml
grep -qE "^  en:" docs-site/hugo.yaml
grep -qE "^  ja:" docs-site/hugo.yaml
grep -qE "^  zh:" docs-site/hugo.yaml
```

### AC-G1.5-03 — Nextra 잔재 0건 (Phase 2 완료 시점)

**Given** Phase 2 cleanup 완료 상태에서
**When** Nextra/Next.js/Bun 런타임 아티팩트를 검색하면
**Then** `docs-site/` 하위에 해당 파일/디렉토리가 **모두 부재**해야 한다.

**자동 검증**:
```bash
! test -f docs-site/package.json
! test -f docs-site/bun.lock
! test -f docs-site/biome.json
! test -f docs-site/next.config.mjs
! test -f docs-site/theme.config.tsx
! test -f docs-site/middleware.ts
! test -d docs-site/app
! test -d docs-site/components
! test -d docs-site/lib
```

### AC-G1.5-04 — hugo server 빈 콘텐츠 기동

**Given** Phase 2 scaffold 완료 + `docs-site/content/` 가 비어있거나 placeholder 만 있는 상태에서
**When** `cd docs-site && hugo server --port 1313 --watch=false` 를 백그라운드로 기동하고 5초 대기 후 stdout 을 확인하면
**Then** `Web Server is available at http://localhost:1313/` 또는 동등한 기동 메시지가 출력되고 에러 0건이어야 한다.

**수동 검증**: 백그라운드 기동 후 `curl -sI http://localhost:1313/` 가 HTTP 응답을 반환하는지 확인하고, stderr 에 `ERROR` 없는지 확인.

### AC-G1.5-05 — 서브트리 크기 + 빌드 베이스라인 문서

**Given** Phase 2 완료 시점에
**When** `.moai/plans/DOCS-SITE/phase-2-subtree-size.md` 와 `.moai/plans/DOCS-SITE/phase-2-build-baseline.md` 를 확인하면
**Then** 각 파일이 존재하고 실측치 (du 결과, hugo 빌드 시간) 가 기록되어야 한다.

**자동 검증**:
```bash
test -f .moai/plans/DOCS-SITE/phase-2-subtree-size.md
test -f .moai/plans/DOCS-SITE/phase-2-build-baseline.md
test "$(wc -l < .moai/plans/DOCS-SITE/phase-2-subtree-size.md)" -ge 5
test "$(wc -l < .moai/plans/DOCS-SITE/phase-2-build-baseline.md)" -ge 5
```

---

## Gate G2 — Hugo Build Gate (Phase 3/4 완료 판정)

### AC-G2-01 — Hugo 빌드 성공 (zero warnings)

**Given** `docs-site/` 디렉토리에 Hugo + Hextra 구성과 219 페이지 콘텐츠가 존재하는 상태에서
**When** `cd docs-site && hugo --minify --gc` 명령을 실행하면
**Then** exit code 0 반환, stdout/stderr 에 `WARN` 또는 `ERROR` 레벨 로그 0건, `docs-site/public/` 디렉토리에 정적 HTML 생성 완료되어야 한다.

**자동 검증**:
```bash
cd docs-site && hugo --minify --gc 2>&1 | tee /tmp/hugo-build.log
grep -iE "^(WARN|ERROR)" /tmp/hugo-build.log && exit 1 || exit 0
test -f docs-site/public/sitemap.xml
test -d docs-site/public/ko
test -d docs-site/public/en
test -d docs-site/public/ja
test -d docs-site/public/zh
```

### AC-G2-02 — 페이지 수 일치 (219 = ko 63 + en 52 + ja 52 + zh 52)

**Given** Phase 3 콘텐츠 마이그레이션 완료 상태에서
**When** `find docs-site/content -name '*.md' -not -name '_index.md' | wc -l` 을 실행하면
**Then** 결과가 219 (± `_index.md` 섹션 파일 허용) 이고, locale 별로 ko=63, en=52, ja=52, zh=52 를 만족해야 한다.

**자동 검증**:
```bash
test "$(find docs-site/content/ko -name '*.md' -not -name '_index.md' | wc -l)" -eq 63
test "$(find docs-site/content/en -name '*.md' -not -name '_index.md' | wc -l)" -eq 52
test "$(find docs-site/content/ja -name '*.md' -not -name '_index.md' | wc -l)" -eq 52
test "$(find docs-site/content/zh -name '*.md' -not -name '_index.md' | wc -l)" -eq 52
```

### AC-G2-03 — Nextra 잔재 제거 (REQ-DS-03)

**Given** Phase 2 cleanup 완료 상태에서
**When** `docs-site/` 하위에서 Nextra/Next.js 런타임 아티팩트를 검색하면
**Then** 0건이어야 한다.

**자동 검증**:
```bash
! test -f docs-site/package.json
! test -f docs-site/bun.lock
! test -f docs-site/biome.json
! test -f docs-site/next.config.mjs
! test -f docs-site/theme.config.tsx
! test -f docs-site/middleware.ts   # Edge Function 은 api/ 하위로 이동
! grep -r "from 'nextra" docs-site/content/
! grep -r 'from "nextra' docs-site/content/
! grep -r "import { Callout }" docs-site/content/
```

### AC-G2-04 — Callout shortcode 전환 완료 (735건, 절대 보존, D-006 강화)

**Given** Phase 3 의 `convert-callouts.go` 실행 완료 상태에서
**When** Hextra callout shortcode 를 grep 으로 세면
**Then** `{{< callout type=` 패턴이 **정확히 735건** 검출되고, 기존 `<Callout type=` JSX 패턴은 0건이어야 한다.

**자동 검증**:
```bash
TOTAL=$(grep -rh "{{< callout type=" docs-site/content/ | wc -l)
test "$TOTAL" -eq 735
! grep -r "<Callout type=" docs-site/content/
```

### AC-G2-05 — frontmatter 주입 완료 (219 페이지)

**Given** `scripts/inject-frontmatter.go` 실행 완료 상태에서
**When** 각 페이지의 첫 3줄을 검사하면
**Then** 모든 페이지가 `---` 로 시작하는 YAML frontmatter 를 가지고, 최소 `title`, `weight`, `draft` 키가 존재해야 한다.

**자동 검증**:
```bash
for f in $(find docs-site/content -name '*.md' -not -name '_index.md'); do
  head -1 "$f" | grep -q '^---$' || { echo "Missing frontmatter: $f"; exit 1; }
  head -10 "$f" | grep -q '^title:' || { echo "Missing title: $f"; exit 1; }
  head -10 "$f" | grep -q '^weight:' || { echo "Missing weight: $f"; exit 1; }
  head -10 "$f" | grep -q '^draft:' || { echo "Missing draft: $f"; exit 1; }
done
```

### AC-G2-06 — Mermaid 569블록 절대 보존 (D-005 강화)

**Given** Phase 3 완료 상태에서
**When** Mermaid 코드블록 개수를 세면
**Then** `\`\`\`mermaid` 패턴이 **정확히 569건** 이어야 한다 (tolerance 제거, 형식 정리 시에도 블록 카운트는 보존).

**자동 검증**:
```bash
COUNT=$(grep -r '^```mermaid' docs-site/content/ | wc -l)
test "$COUNT" -eq 569
```

### AC-G2-07 — Callout / Mermaid / 코드블록 시각 검수

**Given** Hugo 로컬 서버가 `hugo server` 로 기동 중이고
**When** 각 locale 에서 5페이지씩 총 20페이지를 브라우저로 열어 확인하면
**Then** Callout 색상 (tip/info/warning/error/success 5종), Mermaid 다이어그램, 코드블록 syntax highlighting 이 시각적으로 정상 렌더링되어야 한다.

**수동 검증 체크리스트**:
- [ ] ko/getting-started/installation 렌더링 정상
- [ ] en/core-concepts/ddd 렌더링 정상
- [ ] ja/worktree/guide Mermaid 다이어그램 렌더링 정상
- [ ] zh/advanced/agent-guide 코드블록 syntax highlighting 정상
- [ ] ko/quality-commands/review Callout 색상 정상
- [ ] (추가 15페이지 locale별 샘플 검수)

### AC-G2-08 — REQ-DS-19 버전 배너 Partial 존재 (D-009 누락 보완)

**Given** Phase 5 `layouts/partials/version-banner.html` 및 `baseof.html` 조건부 호출이 구현된 상태에서
**When** 배너 partial 파일을 검사하면
**Then** 파일이 존재하고, `v[0-9]` prefix 조건부 렌더 로직과 "Go to latest" 링크 산출 로직이 포함되어야 한다.

**자동 검증**:
```bash
test -f docs-site/layouts/partials/version-banner.html
grep -qE 'v[0-9]' docs-site/layouts/partials/version-banner.html
grep -q 'latest' docs-site/layouts/partials/version-banner.html
grep -q 'version-banner' docs-site/layouts/_default/baseof.html
```

### AC-G2-09 — REQ-DS-29 금지 패턴 CI 스크립트 존재 (D-009 누락 보완)

**Given** Phase 5 `scripts/docs-forbidden-check.sh` 작성 완료 상태에서
**When** 스크립트 파일을 확인하고 실행하면
**Then** 스크립트 파일 존재 + 3개 금지 카테고리 (forbidden URL / Mermaid LR / emoji) 검사 패턴 포함 + 현 docs-site 콘텐츠 대상 exit 0.

**자동 검증**:
```bash
test -x scripts/docs-forbidden-check.sh
grep -q 'docs.moai-ai.dev' scripts/docs-forbidden-check.sh
grep -q 'adk.moai.com' scripts/docs-forbidden-check.sh
grep -q 'adk.moai.kr' scripts/docs-forbidden-check.sh
grep -qE 'flowchart LR|graph LR' scripts/docs-forbidden-check.sh
bash scripts/docs-forbidden-check.sh
```

### AC-G2-10 — 서브트리 크기 실측 기록 (Gap 2)

**Given** Phase 2 subtree 이관 완료 상태에서
**When** `.moai/plans/DOCS-SITE/phase-2-subtree-size.md` 를 확인하면
**Then** Phase 2 이전/이후 `.git/` 크기 또는 squash 커밋 크기 실측치가 수치로 기록되어야 한다.

**자동 검증**:
```bash
test -f .moai/plans/DOCS-SITE/phase-2-subtree-size.md
grep -qE '(MB|KB)' .moai/plans/DOCS-SITE/phase-2-subtree-size.md
```

### AC-G2-11 — Hugo 빌드 시간 베이스라인 (Gap 3)

**Given** Phase 2 및 Phase 3 완료 시점에
**When** `.moai/plans/DOCS-SITE/phase-2-build-baseline.md` 를 확인하면
**Then** 빈 콘텐츠 상태 + 219페이지 상태 두 시점의 `hugo --minify --gc` 실행 시간이 기록되고, Vercel 기본 timeout 10분 대비 여유치가 계산되어야 한다.

**자동 검증**:
```bash
test -f .moai/plans/DOCS-SITE/phase-2-build-baseline.md
grep -qE '(real|seconds|s\s)' .moai/plans/DOCS-SITE/phase-2-build-baseline.md
```

---

## Gate G3 — Automation Readiness Gate (Phase 5/6 완료 판정)

### AC-G3-01 — Edge Function i18n 3시나리오 + runtime 검증 (D-007 강화)

**Given** `docs-site/api/i18n-detect.ts` 가 `vercel dev` 환경에서 실행 중이고
**When** 아래 3가지 요청을 보내면

| 시나리오 | 요청 | 예상 동작 |
|----------|------|-----------|
| S1 cookie 우선 | `GET /` 요청 + `Cookie: locale=ja` | 302 redirect to `/ja/` |
| S2 Accept-Language | `GET /` 요청 + `Accept-Language: zh-CN,zh;q=0.9` (cookie 없음) | 302 redirect to `/zh/` |
| S3 default | `GET /` 요청 + `Accept-Language: fr;q=1` (지원 안함, cookie 없음) | 302 redirect to `/ko/` |

**Then** 각 시나리오 결과가 정확히 일치하고, 추가로 **Edge runtime 임을 응답 헤더로 확인**해야 한다.

**자동 검증**:
```bash
# vercel dev 기동 후
curl -sI -H "Cookie: locale=ja" http://localhost:3000/ | grep -q "location: /ja"
curl -sI -H "Accept-Language: zh-CN,zh;q=0.9" http://localhost:3000/ | grep -q "location: /zh"
curl -sI -H "Accept-Language: fr;q=1" http://localhost:3000/ | grep -q "location: /ko"

# 파일 내부 runtime 선언 확인
grep -qE "runtime:\s*['\"]edge['\"]|export const runtime\s*=\s*['\"]edge['\"]" docs-site/api/i18n-detect.ts

# Deploy 후 Preview URL 대상 (AC-G4-03과 연계)
# curl -I $PREVIEW_URL/api/i18n-detect 의 x-vercel-cache / server 헤더로 Edge runtime 확인
```

### AC-G3-02 — Edge Function bypass 경로 검증

**Given** Edge Function 실행 중
**When** `/api/`, `/_next/`, `/static/`, `/robots.txt` (확장자 포함 경로) 에 요청하면
**Then** 302 redirect 없이 실제 리소스가 반환되어야 한다 (REQ-DS-14).

**자동 검증**:
```bash
curl -sI http://localhost:3000/robots.txt | grep -qE "^HTTP/.* 200"
curl -sI http://localhost:3000/static/og.png | grep -qE "^HTTP/.* 200"
```

### AC-G3-03 — docs-i18n-check 스크립트 통과

**Given** Phase 3 콘텐츠 마이그레이션 완료 상태에서
**When** `bash scripts/docs-i18n-check.sh` 를 실행하면
**Then** exit 0 반환하고, 아래 검증이 모두 통과해야 한다:
- 4개 locale 파일 개수 일치 (ko 전용 섹션 예외 허용)
- 경로 동형성 (`find content/ko -name '*.md'` 와 `find content/en -name '*.md'` 의 상대 경로 일치, contributing/multi-llm 제외)
- 각 파일 frontmatter `title` 비어있지 않음
- "MoAI-ADK" 문자열이 모든 locale 에서 번역되지 않고 유지됨

**자동 검증**: 위 스크립트 실행 exit code 0.

### AC-G3-04 — Version Snapshot Go 스크립트 단위 테스트 (D-011 교정)

**Given** `scripts/docs-version-snapshot.go` 가 구현되어 있고
**When** `go test ./scripts/docs-version-snapshot/...` 실행하면
**Then** 최소 아래 테스트 케이스가 통과해야 한다 (이름 교정 완료):
- **TestMinorRelease**: `v2.12.0` → `v2.13.0` 입력 시 `content/{locale}/v2.12/` 폴더 생성 (minor 릴리스)
- **TestPatchRelease**: `v2.12.1` → `v2.12.2` 입력 시 스냅샷 미생성 (patch 감지, REQ-DS-18 대응)
- **TestMajorRelease**: `v2.12.0` → `v3.0.0` 입력 시 `content/{locale}/v2.12/` 폴더 생성 (major 릴리스, 신규 추가)
- **TestSkipExistingVersions**: 이미 `v2.11/` 존재 시 중복 생성 안함

**자동 검증**:
```bash
go test ./scripts/docs-version-snapshot/... -v
```

### AC-G3-05 — 프로젝트 문서 갱신 (CLAUDE.local.md §17, D-024 정규식 강화)

**Given** Phase 6 문서 갱신 완료 상태에서
**When** `CLAUDE.local.md` 에서 `## 17.` 섹션과 §17.1~§17.6 헤더를 정규식으로 찾으면
**Then** §17 전체와 6개 소섹션이 모두 존재하고, 최소 125 LOC 이상이어야 한다.

**자동 검증**:
```bash
grep -q "^## 17\. docs-site" CLAUDE.local.md
# Heading 공백 변형 허용 정규식 (D-024)
grep -Eq "^### *(§?17\.1)" CLAUDE.local.md
grep -Eq "^### *(§?17\.2)" CLAUDE.local.md
grep -Eq "^### *(§?17\.3)" CLAUDE.local.md
grep -Eq "^### *(§?17\.4)" CLAUDE.local.md
grep -Eq "^### *(§?17\.5)" CLAUDE.local.md
grep -Eq "^### *(§?17\.6)" CLAUDE.local.md
```

### AC-G3-06 — .moai/project 3종 docs-site 반영

**Given** Phase 6 완료 상태에서
**When** `.moai/project/{product,structure,tech}.md` 3개 문서를 grep 하면
**Then** 각 문서에서 아래 키워드가 검출되어야 한다:
- product.md: "Documentation Site" 또는 "adk.mo.ai.kr"
- structure.md: "docs-site/" 디렉토리 구조 언급
- tech.md: "Hugo" AND "Hextra" 동시 언급

**자동 검증**:
```bash
grep -qE "Documentation Site|adk\.mo\.ai\.kr" .moai/project/product.md
grep -q "docs-site/" .moai/project/structure.md
grep -q "Hugo" .moai/project/tech.md
grep -q "Hextra" .moai/project/tech.md
```

### AC-G3-07 — SPEC-I18N-001 archive 이관

**Given** Phase 6 완료 상태에서
**When** `.moai/specs/SPEC-I18N-001/` 디렉토리를 확인하면
**Then** 3개 파일 존재, frontmatter 에 `status: archived` 및 `supersede_by: SPEC-DOCS-SITE-001` 필드가 있어야 한다.

**자동 검증**:
```bash
test -f .moai/specs/SPEC-I18N-001/spec.md
grep -q "status: archived" .moai/specs/SPEC-I18N-001/spec.md
grep -q "supersede_by: SPEC-DOCS-SITE-001" .moai/specs/SPEC-I18N-001/spec.md
```

### AC-G3-08 — REQ-DS-20 3종 custom partial 존재 (D-009 누락 보완)

**Given** Phase 4 partial 구현 완료 상태에서
**When** `docs-site/layouts/partials/custom/` 하위 3개 파일을 확인하면
**Then** `language-switch.html`, `navbar-end.html`, `structured-data.html` 이 모두 존재하고, `language-switch.html` 에는 "KO", "EN", "JA", "ZH" 4개 텍스트 라벨이 포함되어야 한다 (플래그 이모지 금지, D-012).

**자동 검증**:
```bash
test -f docs-site/layouts/partials/custom/language-switch.html
test -f docs-site/layouts/partials/custom/navbar-end.html
test -f docs-site/layouts/partials/custom/structured-data.html
grep -q "KO" docs-site/layouts/partials/custom/language-switch.html
grep -q "EN" docs-site/layouts/partials/custom/language-switch.html
grep -q "JA" docs-site/layouts/partials/custom/language-switch.html
grep -q "ZH" docs-site/layouts/partials/custom/language-switch.html
# 국기 이모지 미포함 확인 (플래그 유니코드 블록 U+1F1E6 ~ U+1F1FF)
! grep -P "[\x{1F1E6}-\x{1F1FF}]" docs-site/layouts/partials/custom/language-switch.html
```

### AC-G3-09 — `_meta.ts` JSON Schema 검증 CI (Gap 7)

**Given** Phase 5 `scripts/validate-meta-schema.go` 작성 완료 상태에서
**When** 변환된 `_index.md` frontmatter 집합을 입력으로 스크립트를 실행하면
**Then** exit 0 반환, 스키마 위반 0건 보고.

**자동 검증**:
```bash
go test ./scripts/validate-meta-schema/... -v
go run ./scripts/validate-meta-schema docs-site/content/
```

---

## Gate G4 — Production Cutover Gate [사용자 수동 승인 필수]

### Phase 7 사전 조사 (Preflight ACs, D-004 / D-013 / D-019 / Gap 1)

#### AC-PRE-01 — DNS 변경 필요성 조사 (D-004)

**Given** Phase 7 cutover preparation 시작 전
**When** Vercel 공식 문서 및 지원팀 응답을 근거로 "기존 프로젝트의 Git source 변경이 DNS 레코드 변경을 동반하는지" 를 조사하면
**Then** 결과가 `.moai/plans/DOCS-SITE/phase-7-dns-assessment.md` 에 기록되어야 한다. 불필요가 확정되면 DNS 조작은 수행하지 않으며, 필요로 확정되면 별도 변경 계획을 문서화 후 수행한다.

**자동 검증**:
```bash
test -f .moai/plans/DOCS-SITE/phase-7-dns-assessment.md
grep -qE '필요성|필요함|불필요' .moai/plans/DOCS-SITE/phase-7-dns-assessment.md
grep -qE 'Vercel.*(공식|문서|지원|support|docs)' .moai/plans/DOCS-SITE/phase-7-dns-assessment.md
```

#### AC-PRE-02 — SEO aggregateRating 영향 조사 (D-019)

**Given** Phase 4 구현 완료 후 Phase 7 사전 단계에서
**When** Google Search Console 에서 Rich Results 중 `aggregateRating` 의존 항목을 조회하고 제거 시 영향도를 평가하면
**Then** 결과가 `.moai/plans/DOCS-SITE/phase-4-seo-impact.md` 에 기록되어야 한다 (노출 건수, 트래픽 영향, Phase 7 모니터링 추가 지표 여부).

**자동 검증**:
```bash
test -f .moai/plans/DOCS-SITE/phase-4-seo-impact.md
grep -qE 'aggregateRating|Rich Results' .moai/plans/DOCS-SITE/phase-4-seo-impact.md
```

#### AC-PRE-03 — Vercel 무중단 전환 근거 확인 (Gap 1)

**Given** Phase 7 cutover preparation 시작 전
**When** Vercel 공식 문서 또는 지원팀 답변 등으로 "기존 Vercel project의 Git source 변경 시 도메인 바인딩이 유지되어 무중단 전환 가능하다" 는 근거를 수집하면
**Then** 결과가 `.moai/plans/DOCS-SITE/phase-7-zero-downtime-evidence.md` 에 기록되어야 한다 (출처 URL 또는 지원팀 티켓 번호 포함).

**자동 검증**:
```bash
test -f .moai/plans/DOCS-SITE/phase-7-zero-downtime-evidence.md
grep -qE 'vercel\.com|docs\.vercel' .moai/plans/DOCS-SITE/phase-7-zero-downtime-evidence.md
```

### AC-G4-01 — Vercel Preview URL 4 locale 홈 검증

**Given** Phase 7 Vercel Project 재바인딩 완료 후 Preview URL 이 발급된 상태에서
**When** 4개 locale 홈 URL (Preview URL + `/ko/`, `/en/`, `/ja/`, `/zh/`) 을 브라우저로 방문하면
**Then** 모든 홈 페이지가 정상 로드되고, 언어 스위처, 사이드바, 네비게이션 바, 푸터가 표시되어야 한다.

**수동 체크리스트**:
- [ ] `/ko/` 홈 로드 정상
- [ ] `/en/` 홈 로드 정상
- [ ] `/ja/` 홈 로드 정상
- [ ] `/zh/` 홈 로드 정상
- [ ] 언어 스위처 드롭다운에 4개 locale 노출
- [ ] 각 locale 에서 `KO / EN / JA / ZH` 라벨 정상 표시 (이모지 0건)

### AC-G4-02 — Preview URL 주요 페이지 10건 시각 검수

**Given** Preview 환경에서
**When** 아래 10개 주요 페이지를 각 locale 로 방문하면
**Then** 모든 페이지가 정상 렌더링되고, 링크 이동이 정상 작동해야 한다.

**대상 페이지 (locale 당 10건 = 총 40페이지)**:
1. `getting-started/installation`
2. `getting-started/quickstart`
3. `core-concepts/ddd`
4. `core-concepts/trust-5`
5. `workflow-commands/plan`
6. `workflow-commands/run`
7. `worktree/guide`
8. `worktree/faq`
9. `advanced/agent-guide`
10. `quality-commands/review`

**수동 체크리스트**:
- [ ] ko 10페이지 모두 정상
- [ ] en 10페이지 모두 정상
- [ ] ja 10페이지 모두 정상
- [ ] zh 10페이지 모두 정상
- [ ] Mermaid 다이어그램 라이트/다크 모드 각각 렌더링 정상
- [ ] 코드블록 syntax highlighting 정상
- [ ] Callout 5종 색상 정상

### AC-G4-03 — Edge Function 실환경 i18n 검증 (D-025 사용 지침 추가)

**Given** Preview URL 환경에서
**When** AC-G3-01 의 3시나리오 (cookie / Accept-Language / default) 를 실제 Preview URL 대상으로 curl 로 실행하면
**Then** 각 시나리오가 예상대로 302 redirect 되고, Edge runtime 응답 헤더 (`x-vercel-cache` 또는 `server`) 로 Edge Function 임이 확인되어야 한다.

**사용 지침**: Preview URL 은 Vercel Dashboard → Deployments → 해당 Preview Deployment 의 URL 을 복사하여 환경변수로 export 한 후 스크립트를 실행한다.

```bash
# 실행 전 초기화 필수 (placeholder 직접 실행 금지)
export PREVIEW_URL="https://<실제-preview-deployment-도메인>.vercel.app"
```

**자동 검증**:
```bash
test -n "$PREVIEW_URL" || { echo "PREVIEW_URL not set"; exit 1; }
curl -sI -H "Cookie: locale=ja" "$PREVIEW_URL/" | grep -q "location: /ja"
curl -sI -H "Accept-Language: zh-CN,zh;q=0.9" "$PREVIEW_URL/" | grep -q "location: /zh"
curl -sI -H "Accept-Language: fr;q=1" "$PREVIEW_URL/" | grep -q "location: /ko"

# Edge runtime 헤더 확인 (D-007 추가 검증)
curl -sI "$PREVIEW_URL/api/i18n-detect" | grep -qiE "x-vercel-cache|server"
```

### AC-G4-04 — redirect 규칙 (moai-rank) 보존

**Given** Preview URL 환경에서
**When** `/moai-rank/anything`, `/ko/moai-rank/anything` 에 요청하면
**Then** 각각 `/`, `/ko` 로 301 redirect 되어야 한다 (REQ-DS-23).

**자동 검증**:
```bash
test -n "$PREVIEW_URL" || { echo "PREVIEW_URL not set"; exit 1; }
curl -sI "$PREVIEW_URL/moai-rank/test" | grep -q "HTTP/.* 301"
curl -sI "$PREVIEW_URL/moai-rank/test" | grep -qE "location: /"
curl -sI "$PREVIEW_URL/ko/moai-rank/test" | grep -qE "location: /ko"
```

### AC-G4-05 — SEO 아티팩트 검증 + og.png 크기 (Gap 6)

**Given** Preview URL 환경에서
**When** `/sitemap.xml`, `/robots.txt`, `/og.png` 에 요청하면
**Then** 200 OK 응답, sitemap 에 219 페이지 + 4 locale 대응 hreflang 정보 포함, robots.txt 에 sitemap URL 정확히 기록, og.png Content-Length ≤ 500 KB (= 512000 bytes) 이어야 한다 (REQ-DS-22 + R9).

**자동 검증**:
```bash
test -n "$PREVIEW_URL" || { echo "PREVIEW_URL not set"; exit 1; }
curl -s "$PREVIEW_URL/sitemap.xml" | grep -c "<url>" | awk '{if($1 >= 200) exit 0; else exit 1}'
curl -s "$PREVIEW_URL/sitemap.xml" | grep -q 'hreflang="ko"'
curl -s "$PREVIEW_URL/sitemap.xml" | grep -q 'hreflang="en"'
curl -s "$PREVIEW_URL/sitemap.xml" | grep -q 'hreflang="ja"'
curl -s "$PREVIEW_URL/sitemap.xml" | grep -q 'hreflang="zh"'
curl -s "$PREVIEW_URL/robots.txt" | grep -q "Sitemap:"

# og.png 크기 (Gap 6 신규)
OG_SIZE=$(curl -sI "$PREVIEW_URL/og.png" | awk -F': ' 'tolower($1)=="content-length"{gsub(/\r/,""); print $2}')
test -n "$OG_SIZE" && test "$OG_SIZE" -le 512000
```

### AC-G4-06 — JSON-LD structured data 검증 (aggregateRating 제거 확인)

**Given** Preview URL 홈 페이지에서
**When** HTML 소스의 `<script type="application/ld+json">` 태그를 확인하면
**Then** Organization / SoftwareApplication / WebSite / TechArticle 4종 JSON-LD 가 존재하고, `aggregateRating` 필드는 어디에도 없어야 한다 (REQ-DS-21).

**자동 검증**:
```bash
test -n "$PREVIEW_URL" || { echo "PREVIEW_URL not set"; exit 1; }
curl -s "$PREVIEW_URL/ko/" | grep -c '"@type": "Organization"'
curl -s "$PREVIEW_URL/ko/" | grep -c '"@type": "SoftwareApplication"'
curl -s "$PREVIEW_URL/ko/" | grep -c '"@type": "WebSite"'
curl -s "$PREVIEW_URL/ko/" | grep -c '"@type": "TechArticle"'
! curl -s "$PREVIEW_URL/ko/" | grep -q "aggregateRating"
```

### AC-G4-07 — Lighthouse score baseline + Nextra 상대 비교 (D-013)

**Given** Preview URL 의 `/ko/getting-started/quickstart` 페이지에서
**When** Chrome Lighthouse Desktop/Mobile 감사를 실행하고, 사전 측정된 Nextra baseline 과 비교하면
**Then** 아래 절대 기준 **AND** 상대 기준을 모두 만족해야 한다:

| 지표 | 최저 기준 (Desktop) | 최저 기준 (Mobile) | Nextra baseline 대비 (Desktop) |
|------|---------------------|--------------------|-------------------------------|
| Performance | ≥ 85 | ≥ 75 | **baseline - 5 이내** |
| Accessibility | ≥ 90 | ≥ 90 | — |
| Best Practices | ≥ 90 | ≥ 90 | — |
| SEO | ≥ 95 | ≥ 95 | — |

Nextra baseline 은 Phase 7 사전 준비 단계에서 현재 프로덕션 `adk.mo.ai.kr` 에 대해 측정하여 `.moai/plans/DOCS-SITE/phase-7-preview-verification.md` 의 baseline 필드에 기록한다.

**수동 검증**: Lighthouse 리포트 HTML 저장 후 `.moai/plans/DOCS-SITE/phase-7-lighthouse-report.html` 로 보관, baseline 필드와 교차 검증.

### AC-G4-08 — 링크 체커 (404 0건)

**Given** Preview URL 에 대해 link checker 를 실행하면
**When** 전체 사이트 crawl 후 깨진 링크를 집계하면
**Then** 내부 링크 404 는 0건이어야 한다 (외부 링크는 참고용).

**자동 검증** (옵션 예시):
```bash
# muffet, lychee, 또는 유사 도구 사용
lychee --exclude-all-private --base "$PREVIEW_URL" "$PREVIEW_URL"
# exit 0 필수
```

### AC-G4-09 — 사용자 최종 승인 (AskUserQuestion)

**Given** AC-G4-01 ~ AC-G4-08 + AC-PRE-01 ~ AC-PRE-03 이 모두 통과된 상태에서
**When** 오케스트레이터가 AskUserQuestion 으로 "Preview 검증 완료. Production 전환을 진행하시겠습니까?" 질문을 GOOS 에게 제시하면
**Then** GOOS 가 "승인" 선택한 후에만 Production 프로모션이 진행되어야 한다 (REQ-DS-32b 트리거).

**수동 검증**: AskUserQuestion 응답 로그 확인.

### AC-G4-10 — Production 전환 후 초기 스모크 테스트

**Given** Vercel Dashboard 에서 Preview → Production 프로모션 완료 직후
**When** `adk.mo.ai.kr` 에 대해 4 locale 홈, 주요 10 페이지, Edge Function 3시나리오 를 재실행하면
**Then** 모든 검증이 Preview 시점과 동일하게 통과해야 한다.

**자동 검증**:
```bash
PROD_URL="https://adk.mo.ai.kr"
curl -sI "$PROD_URL/ko/" | grep -qE "^HTTP/.* 200"
curl -sI -H "Cookie: locale=ja" "$PROD_URL/" | grep -q "location: /ja"
# (AC-G4-01~05 의 Production 버전)
```

### AC-G4-11 — llms.txt 이관 및 URL 검증 (Gap 5)

**Given** Preview URL 환경에서
**When** `/llms.txt` 에 요청하면
**Then** 200 OK 응답이 반환되고, 파일 내부 URL 들이 새 도메인 (`adk.mo.ai.kr`) 또는 locale-relative 경로로 유효해야 한다 (구 도메인 `moai-adk-docs.vercel.app` 잔재 0건).

**자동 검증**:
```bash
test -n "$PREVIEW_URL" || { echo "PREVIEW_URL not set"; exit 1; }
curl -sI "$PREVIEW_URL/llms.txt" | grep -qE "^HTTP/.* 200"
curl -s "$PREVIEW_URL/llms.txt" > /tmp/llms.txt
! grep -q "moai-adk-docs.vercel.app" /tmp/llms.txt
```

---

## 48시간 모니터링 게이트 (Phase 7 → Phase 8)

### AC-MON-01 — 48h 무장애 운영

**Given** Production 전환 직후
**When** 48시간 경과 후 Vercel Analytics / Monitoring 대시보드를 확인하면
**Then** 아래 지표가 모두 만족되어야 한다:

| 지표 | 기준 |
|------|------|
| P1/P2 인시던트 | 0건 |
| 4xx 에러율 | 기존 Nextra 베이스라인 대비 ±10% 이내 |
| 5xx 에러율 | 0.1% 미만 |
| Largest Contentful Paint (p75) | 2.5s 이하 |
| TTFB (p75) | 800ms 이하 |
| DNS 해석 실패 보고 | 0건 |

**수동 검증**: Vercel Analytics 대시보드 캡처, 48h 통계 저장 → `.moai/plans/DOCS-SITE/phase-7-monitoring-48h.md`

### AC-MON-02 — 사용자 제보 채널 (Discord) 모니터링

**Given** 48h 모니터링 기간 중
**When** Discord `#moai-adk` 채널의 문서 사이트 관련 제보를 집계하면
**Then** 전환 관련 **치명 제보** 는 0건이어야 한다. 경미 제보는 허용.

**치명 / 경미 분류 기준**:

| 분류 | 정의 | 예시 |
|------|------|------|
| 치명 (blocker) | 사이트 완전 접근 불가, 번역 언어 오류로 타 locale 노출, 링크 대량 깨짐 (5건 이상) | `/ko/*` 전역 404, 중국어 페이지가 일본어로 표시, 메인 네비게이션 모든 링크 오류 |
| 경미 (non-blocker) | 개별 페이지 렌더링 오류, 단일 링크 오타, Mermaid 한 개 미렌더 등 | `/ko/foo` 페이지의 Callout 색상 이상 |

**수동 검증**: Discord 로그 검토 후 `.moai/plans/DOCS-SITE/phase-7-monitoring-48h.md` 에 요약.

### AC-MON-03 — 48h 위반 시 즉시 롤백 (D-014, REQ-DS-35 대응)

**Given** Production 전환 후 48시간 모니터링 중
**When** AC-MON-01 기준 중 어느 하나라도 위반되거나 AC-MON-02 치명 제보가 1건 이상 발생하면
**Then** 운영자는 즉시 다음 절차를 수행해야 한다 (REQ-DS-35):

1. Vercel Dashboard 에서 프로젝트 `prj_EZaVdfE3gJeXVbizafBEECpniINP` 의 Git source 를 `modu-ai/moai-adk-go` 에서 `modu-ai/moai-adk-docs` main 으로 되돌린다.
2. Root Directory 를 원래 값(빈 값 또는 moai-adk-docs 기존 설정)으로 복원한다.
3. Framework Preset 을 Next.js (또는 moai-adk-docs 원래 설정)로 복원한다.
4. 24시간 내 `.moai/plans/DOCS-SITE/phase-7-rollback.md` 에 위반 트리거, 롤백 타임스탬프, 사용자 영향 요약, 재시도 결정을 기록한다.

**자동 검증** (롤백 발생 시):
```bash
test -f .moai/plans/DOCS-SITE/phase-7-rollback.md
grep -qE 'trigger|위반|violation' .moai/plans/DOCS-SITE/phase-7-rollback.md
grep -qE 'revert|롤백|rollback' .moai/plans/DOCS-SITE/phase-7-rollback.md
```

**수동 검증**: Vercel Dashboard 에서 Git source 가 moai-adk-docs 로 되돌려진 이력 스크린샷 첨부.

---

## Phase 8 — Archive 완료 판정

### AC-P8-01 — moai-adk-docs archive 태그

**Given** 48h 모니터링 통과 후 (REQ-DS-35 롤백 미발생 확인 포함)
**When** `moai-adk-docs` 레포에 `archive/pre-migration-2026-04` 태그를 발행하면
**Then** 태그가 GitHub 에서 조회 가능해야 한다 (REQ-DS-34).

**자동 검증**:
```bash
git -C /Users/goos/moai/moai-docs tag -l "archive/pre-migration-2026-04" | grep -q .
gh api repos/modu-ai/moai-adk-docs/tags | jq -r '.[].name' | grep -q "archive/pre-migration-2026-04"
```

### AC-P8-02 — GitHub Archive 플래그

**Given** GitHub 레포 설정에서
**When** `moai-adk-docs` 의 Archive 상태를 조회하면
**Then** `archived: true` 가 설정되어야 한다.

**자동 검증**:
```bash
gh api repos/modu-ai/moai-adk-docs | jq -r '.archived' | grep -q "true"
```

### AC-P8-03 — archived 레포 README 전환 안내

**Given** archive 완료된 `moai-adk-docs` 레포에서
**When** `README.md` 상단을 확인하면
**Then** "This repository has been archived and migrated to `moai-adk-go/docs-site`" 문구와 이관 대상 링크가 첫 섹션에 포함되어야 한다.

**수동 검증**: README 직접 확인.

### AC-P8-04 — postmortem 기록

**Given** Phase 8 완료 시
**When** `.moai/plans/DOCS-SITE/phase-8-postmortem.md` 를 확인하면
**Then** Phase 0~7 에서 발생한 이슈, 해결 방법, 후속 개선점이 최소 5건 이상 기록되어야 한다.

**자동 검증**:
```bash
test -f .moai/plans/DOCS-SITE/phase-8-postmortem.md
test "$(wc -l < .moai/plans/DOCS-SITE/phase-8-postmortem.md)" -ge 80
```

---

## Definition of Done (최종 완료 기준)

본 SPEC 전체 완료 판정은 아래 모든 항목이 체크되어야 한다:

- [ ] Phase 1~8 의 모든 Gate (G1 / G1.5 / G2 / G3 / G4 + 48h Mon + Archive) 통과
- [ ] spec.md 의 REQ-DS-01 ~ REQ-DS-35 모두 대응 acceptance criteria 충족 (REQ-DS-32는 32a/32b 분할)
- [ ] SPEC-I18N-001 이 `.moai/specs/SPEC-I18N-001/` 에 archived 상태로 이관
- [ ] `CLAUDE.local.md` §17 존재 및 6개 소섹션 완성
- [ ] `.moai/project/{product,structure,tech}.md` 3종 갱신
- [ ] Vercel 프로젝트 `prj_EZaVdfE3gJeXVbizafBEECpniINP` 가 `moai-adk-go/docs-site` 를 소스로 운영 중
- [ ] `adk.mo.ai.kr` 에서 Hugo 빌드 결과가 정상 서빙
- [ ] 48h 프로덕션 안정성 확인, REQ-DS-35 롤백 미발생
- [ ] `moai-adk-docs` 레포 archive 완료
- [ ] Phase 8 postmortem 작성 완료

---

## 부록 A: 자동 검증 스크립트 일괄 실행

모든 자동 검증을 한 번에 실행하는 마스터 스크립트 경로 (Phase 5 산출물):

```bash
bash scripts/docs-site-verify-all.sh <gate: g1|g1.5|g2|g3|g4|mon|p8>
```

CI 통합 경로: `.github/workflows/docs-site.yml` — PR 트리거 시 G2 + G3 자동 실행.

## 부록 B: 수동 검증 체크리스트 통합

본 문서의 모든 "수동 검증" 체크박스를 Phase 7 운영 문서로 발췌하여 `.moai/plans/DOCS-SITE/phase-7-preview-verification.md` 에 정리 후 사용.

## 부록 C: REQ ↔ AC 매핑 매트릭스 (D-009 대응)

| REQ | 대응 AC | 비고 |
|-----|---------|------|
| REQ-DS-01 | AC-G1.5-01 | squash 커밋 단일성 |
| REQ-DS-02 | AC-G1.5-02, AC-G2-01 | hugo.yaml + 빌드 성공 |
| REQ-DS-03 | AC-G1.5-03, AC-G2-03 | Nextra 잔재 0건 |
| REQ-DS-04 | AC-G2-02 | 219 페이지 |
| REQ-DS-05 | AC-G2-04 | Callout 735건 절대 보존 |
| REQ-DS-06 | AC-G2-05 | frontmatter 주입 |
| REQ-DS-07 | AC-G3-09 | `_meta.ts` schema |
| REQ-DS-08 | AC-G3-04 | Go 스크립트 존재 |
| REQ-DS-09 | AC-G2-06 | Mermaid 569건 절대 보존 |
| REQ-DS-10 | plan.md § 5 Mermaid Strategy 문서화 | 문서 참조 |
| REQ-DS-11 | (Optional) | AC 면제 |
| REQ-DS-12 | AC-G4-01, AC-G4-02 | URL 구조 보존 |
| REQ-DS-13 | AC-G3-01, AC-G4-03 | Edge Function 3시나리오 + runtime |
| REQ-DS-14 | AC-G3-02 | bypass 경로 |
| REQ-DS-15 | AC-G2-02 (ko 전용 예외), plan.md Phase 3 step 2 | 스켈레톤 aliases |
| REQ-DS-16 | AC-G2-08 (버전 배너 partial) | |
| REQ-DS-17 | AC-G3-04 (TestMinorRelease, TestMajorRelease) | 이전 릴리스 태그 기준 |
| REQ-DS-18 | AC-G3-04 (TestPatchRelease) | |
| REQ-DS-19 | AC-G2-08 | 버전 배너 partial |
| REQ-DS-20 | AC-G3-08 | 3종 partial + 이모지 0건 |
| REQ-DS-21 | AC-G4-06 | aggregateRating 제거 |
| REQ-DS-22 | AC-G4-05 | SEO 아티팩트 + og.png |
| REQ-DS-23 | AC-G4-04 | moai-rank redirect |
| REQ-DS-24 | AC-G3-05 | CLAUDE.local.md §17 |
| REQ-DS-25 | AC-G3-06 (product.md) | |
| REQ-DS-26 | AC-G3-06 (structure.md) | |
| REQ-DS-27 | AC-G3-06 (tech.md) | |
| REQ-DS-28 | AC-G3-07 | SPEC-I18N-001 archive |
| REQ-DS-29 | AC-G2-09 | 금지 패턴 CI |
| REQ-DS-30 | AC-G2-01 | Hugo 빌드 성공 |
| REQ-DS-31 | AC-G3-03 | docs-i18n-check |
| REQ-DS-32a | Phase 7 step 2 (AC-PRE-01~03 선행) | 재바인딩 = G4 이전 |
| REQ-DS-32b | AC-G4-09 | Production promotion = G4 이후 |
| REQ-DS-33 | AC-G4-01 ~ AC-G4-08 | Preview 선행 검증 |
| REQ-DS-34 | AC-P8-01 ~ AC-P8-03 | Phase 8 automation |
| REQ-DS-35 | AC-MON-03 | 48h 위반 롤백 |
