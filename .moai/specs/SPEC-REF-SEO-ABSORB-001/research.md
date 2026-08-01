# Research — SPEC-REF-SEO-ABSORB-001

`moai-ref-seo` 신규 레퍼런스 스킬을 클린룸 재작성(clean-room rewrite)으로 흡수하기 위한 plan-phase 조사 기록.
3개 read-only 렌즈(A: 원문 개념 인벤토리 / B: `moai-ref-*` 저작 표준 / C: 템플릿 배포 제약)와 오케스트레이터 직접 실측을 합친다.

이 문서는 Epic 1/3의 조사 기록이자, 형제 SPEC 2건(design-taste-frontend → ui-polish 상위 계층 / security + review-rubric 흡수)이 재사용할 **클린룸 프로토콜의 근거 기록**이다.

---

## §A. Lens A — 원문 개념 인벤토리

### A.1 원문 식별

| 항목 | 값 |
|---|---|
| 경로 | `~/.agents/skills/higgsfield-websites/references/seo.md` |
| 분량 | 861 lines (실측: `wc -l`) |
| 최상위 섹션 | 6개 |
| 구별되는 개념 | 43개 |

섹션 분량 배분: Meta tags & OG ~17%, Technical SEO ~25%, Schema markup ~20%, Entity SEO ~14%, GEO/content ~12%, Audit ~12%.

### A.2 내구성 3단계 분류 (durability triage)

**Tier 1 — 내구성 높음 (~55-60%, 1차 흡수 대상)**

- canonical URL 규율: 리소스당 하나의 URL, trailing-slash 중복쌍 금지, canonical 값이 실제 서빙 URL과 일치
- 페이지별 고유 title/description + 브랜드 포함 title 공식
- robots.txt / sitemap.xml을 호스트에서 파생되는 1급 산출물로 취급하고 실제 라우트와 동기화
- JSON-LD 구조화 데이터: 사이트 아키타입별 타입 선택, 타입별 필수 필드, 절대 URL, identifier 상호참조를 통한 graph 번들링
- 구조화 데이터는 눈에 보이는 콘텐츠를 반영해야 하며 보이지 않는 사실을 주장하지 않는다는 원칙
- 엔티티 정체성 위생: 하나의 정규 명칭, 검증된 소유 프로필 링크, 구조화 데이터/페이지/제3자 리스팅 간 name·address·phone 일관성
- 접근성-as-SEO 점검: 단일 H1 + 레벨 건너뛰기 없음, 이미지 alt, 서술적 anchor text, 라벨링된 폼 컨트롤, 가시적 포커스를 동반한 키보드 조작 가능성, 살아있는 in-page fragment 타깃
- 오류 경로와 리다이렉트 경로를 모두 덮는 단일 response-header 초크포인트
- 명시적 PASS/WARN/FAIL 어휘와 "처음부터 다시 실행" 수정 루프를 갖춘 pre-ship 게이트

**Tier 2 — 중간 내구성**

- 절대 URL 1200x630 이미지를 동반한 OG/Twitter 카드 세트
- 페이지 클래스별 robots 지시자 계층화
- 제3자 origin에 대한 preconnect/dns-prefetch를 스타일시트보다 앞에 배치
- 크롤러를 위한 server-rendered HTML 자세
- 모바일 타입 크기·행간 하한 + viewport 태그

**Tier 3 — 휘발성 높음 (기재된 수치 그대로 이관 금지)**

- title 60자 / description 150-160자 예산, "첫 100자 안에 키워드" — SERP 절단은 픽셀 기준이며 변동한다
- sitemap `priority` / `changefreq` — 주요 엔진이 공개적으로 후순위화
- generative-engine-optimization 원칙 전체 — 가장 최신이며 가장 덜 정착
- 200-word / 50-word content-to-code 임계값

### A.3 재현 금지 대상

**출처 없는 수치 주장 3건** — 원문에 인용이 없다. 재작성물에 이런 종류의 수치를 넣으려면 검증 가능한 출처를 달거나 아예 생략해야 한다.

- "80% less engagement"
- sub-100ms TTFB 주장
- preconnect 100-300ms 절감

**플랫폼 종속 (일반화된 대응물 없음, 전면 제외)**

- 빌드타임 메타데이터 사이드카 JSON을 1st-party 마켓플레이스 리스팅 카드에 동기화하는 흐름 — 유료 커버 비디오 생성 및 크레딧 비용 상호작용 규칙 포함

### A.4 원문의 가장 특징적인 기여 — 정보가 아니라 구조

- **10항목 pre-deploy 감사를 BLOCKING 게이트로 패키징**: source-read-only, 항목별 명시적 FAIL/WARN 임계값, FAIL 0이 될 때까지 반복, 규정된 리포트 렌더링. 개별 점검 항목 자체는 업계 표준이지만, **게이트 패키징과 구체적 수치 임계값은 저작 판단**이다.
- **접근성 점검을 SEO 감사에 의도적으로 융합** — 대부분의 SEO 레퍼런스가 하지 않는 선택.

### A.5 구조 골격 (구조는 사실이며 보호 대상 표현이 아니므로 학습 가능)

- 생산자 워크플로 순서를 따르는 topic-major 배열 (메타데이터를 먼저 작성, 검증을 마지막에)
- 각 섹션: 근거 산문 → 하위 섹션 → 레퍼런스 표 → 말미 5항목 pitfalls 목록
- 결정 트리가 아닌 평면 조회 매트릭스
- before/after 산문 예시 1건

---

## §B. Lens B — `moai-ref-*` 저작 표준

### B.1 frontmatter 요구 사항

`name`, `description`(folded scalar), `when_to_use`(folded scalar), `user-invocable: false`,
`metadata` 맵(따옴표 처리된 키·값: `version`, `category: "domain"`, `status: "active"`, `updated`, `tags` — tags는 `, reference`로 종료),
`progressive_disclosure` 블록(`enabled: true`, `level1_tokens: 100`, `level2_tokens: 3000`).

`description` + `when_to_use` 합산 1,536자 미만.

**CSV-vs-array 함정**: `allowed-tools`와 `paths`는 쉼표 구분 문자열(공백 구분 금지 — 조용히 오파싱). `skills:`만 YAML 배열이며, 스킬 자신의 frontmatter가 아니라 스킬을 **참조하는 쪽**에 등장한다.

### B.2 `description` 3악장 (기존 ref 스킬 9개 전수 관찰)

1. `reference`라는 단어로 끝나는 범위 명사구
2. 불변 역할 문장: "Agent-extending skill that amplifies X with production-grade Y patterns."
3. 마지막 `NOT for:` 부정 범위절 — 다른 스킬이 소유한 주제는 `(see moai-ref-x)` 형태로 인라인 지목

`when_to_use`는 동사 우선 사용자 의도 표현으로 전환한다.

### B.3 본문 구조

- 단일 H1 `# <Domain> Reference`
- 선택적 `## Target Agents`
- 4-10개 도메인 H2 — 산문이나 코드가 아니라 **압도적으로 마크다운 표**
- severity 표
- 말미 **필수 불변 3종 세트**(이 순서 고정):
  1. `## Common Rationalizations` — 2열(Rationalization/Reality) 표, 5-6행
  2. `## Red Flags` — 5-10개 구체적 관측 가능 결함
  3. `## Verification` — 6-9개 체크박스, 상당수가 증거를 요구하는 표현
- 3종 각각은 evolvable-zone HTML 주석으로 감싸며 id는 `rationalizations`, `red-flags`, `verification` — 기존 ref 스킬 전체에서 동일

### B.4 분량과 파일 구성

- 기존 ref 스킬: 7.2-20.8 KB / 159-345 lines
- 9개 중 8개가 SKILL.md 단일 파일
- `references/`를 쓰는 ref 스킬은 **없음**. 유일한 하위 디렉터리 선례는 `modules/`(3개 하위 도메인으로 분할한 스킬 1건)
- 신규 스킬 실무 목표: **~160-200 lines / 7-10 KB, SKILL.md 단일 파일**

참고 실측: `moai-ref-ui-polish/SKILL.md` = 184 lines, 명명 H2 ~~11개~~ **10개**(Target Agents + 도메인 8 + Review Modes) + 불변 3종 = **총 13개**.

> **정정 (실측)**: `grep -c '^## ' internal/template/templates/.claude/skills/moai-ref-ui-polish/SKILL.md` → `13`. 괄호 안 열거(Target Agents 1 + 도메인 8 + Review Modes 1 = 10)는 처음부터 정확했고, 앞의 숫자 `11`만 1건 과다 계상이었다. 이 오기가 `acceptance.md` AC-SEO-002의 "H2 11개 + 불변 3 = 14" 근거 문장으로 그대로 전파되었다가 v0.3.0에서 정정되었다. **H2 상한 14 자체는 무관하다** — 상한은 이 선례가 아니라 `spec.md` REQ-SEO-002의 도메인 H2 4-10 + 불변 3 + 선택 1에서 독립적으로 도출된다. 참고로 ref 스킬 9건 H2 실측 분포: secops 9 / api-patterns 11 / owasp-checklist 11 / react-patterns 11 / testing-pyramid 13 / ui-polish 13 / llm-security 14 / supply-chain 15 / git-workflow 16. 형제 SPEC 2건은 이 분포를 "상한 근거"로 오독하지 말 것 — 각자의 REQ가 도메인 H2 범위를 정한다.

### B.5 등록 표면 (아무도 참조하지 않는 스킬 파일은 inert)

템플릿 미러 / `catalog.yaml` 엔트리 / Go 카운트 상수 3개 / `delegation.yaml`(로컬·템플릿 양쪽) / 워크플로 본문 `Skill()` 주입 문자열 / ~~skill-routing 규칙~~ / docs-site skill-guide 표 4개 로케일 / skills-manifest 스팟체크 목록.

> **정정 (실측)**: 위 열거의 `skill-routing 규칙`은 **등록 표면이 아니다.** `grep -c 'moai-ref-secops' .claude/rules/moai/workflow/skill-routing.md` → `0` — 완전 등록 선례인 `moai-ref-secops`가 그 파일에 등장하지 않는다. 거기 있는 3개 스킬 예시는 고정 예시일 뿐이다. 실측 표면 집합은 §D.3 비교표와 `plan.md` §B.5 / `acceptance.md` AC-SEO-022를 따른다. 형제 SPEC 2건은 이 열거가 아니라 §D.3을 근거로 삼는다.

---

## §C. Lens C — 템플릿 배포 제약

### C.1 템플릿 트리 금지 콘텐츠 클래스

내부 SPEC ID, 내부 REQ/AC 토큰, 감사·사후분석 인용, 내부 세션 날짜, 내부 아카이브 경로, 내부 commit SHA, 내부 메모리 파일 참조, `/Users/` 하위 macOS 편향 절대 경로, 메인테이너 로컬 지시 파일 참조, PR 번호 참조, `| Word (canonical)` 표 헤더 형태.

스킬 본문 한정 추가분: 특정 내부 Go 구현 경로 참조, 특정 형태의 constraint 토큰.

**[CRITICAL] 본 SPEC의 acceptance criteria 저작에 직결**: `REQ-*` / `AC-*` 토큰은 **접두사와 무관하게** `.claude/skills/` 내부에서 플래그된다. 따라서 배포되는 SKILL.md는 그 토큰 형태를 쓰는 GEARS 예시 요구사항을 포함해서는 안 된다. 날짜 역시 strict 모드에서 frontmatter `updated:` 필드(unfenced) 외에는 금지된다.

### C.2 16개 언어 중립성

지원 집합(전부 동등, PRIMARY 없음, enabled/planned 분리 없음):
`go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift`

기계 강제: 스킬 본문(frontmatter·펜스 코드 제거 후)에 대한 대소문자 무시 정규식이 "<lang> is primary", "<lang> is the default", "only <lang> is supported" 류 우위 표현을 매칭한다. **교대 목록에 맨 토큰 `go`와 `r`이 포함**되므로 "Go is the default" 같은 우발적 문장이 hard-fail한다.

SEO 스킬에서 이것은 살아있는 위험이다. 완화책: 프레임워크 워크스루가 아니라 **프로토콜·출력 레이어**(robots.txt, sitemap.xml, canonical link element, JSON-LD, Core Web Vitals, hreflang, meta/OpenGraph 태그, 렌더링 모드를 "질문"으로 서술)에 서술을 고정한다. 프레임워크를 지목해야 하면 동일 깊이로 여러 개를 나열하거나 아예 나열하지 않는다. `when_to_use`에 명시적 중립성 문구를 두는 것은 이미 확립된 선례다(`moai-ref-ui-polish`가 "Implementation examples are Web/CSS; the design principles are platform-neutral" 문장을 사용).

### C.3 기타 가드

- 트리 전역 가드가 `GOOS=` 토큰을 포함한 템플릿 파일이 **정확히 2개**임을 단언한다 — 신규 스킬은 이 토큰을 추가해서는 안 된다.

### C.4 빌드 사이클

1. 템플릿 트리에서 먼저 저작
2. `catalog.yaml` 엔트리를 **수기로 추가** (생성기는 신규 스킬을 발견하지 않는다)
3. `make build` — 템플릿 상태에 대한 유일한 생성 효과는 catalog 해시 재생성
4. 재생성된 `catalog.yaml` 커밋 (전체 파일 재마샬링 — 주석 보존 안 됨)
5. 카운트 상수 증가
6. 로컬 미러 동기화

생성된 embedded-file 목록 파일은 **없다**. 임베딩은 컴파일 타임 디렉터리 워크로 이루어진다.

catalog 엔트리 형태 5키: `name`, `tier`, `path`(`templates/` 접두사 + 스킬 디렉터리는 후행 슬래시), `hash`(정규화된 SKILL.md만 대상 sha256, 수기 작성 금지 — 생성), `version`.

### C.5 tier 선택은 기능적 결정

catalog-tier 인지 파일시스템 래퍼가 slim 배포 경로에서 non-core 엔트리를 전부 숨긴다. `core`는 slim FS에 포함, `optional-pack:*`는 미포함.

### C.6 CI path-filter 위험

마크다운 전용 템플릿 변경은 docs-only로 분류되어 race 활성화 Go 테스트 잡이 **SKIP**된다. Go 카운트 상수 파일을 편집하면 분류기가 뒤집혀 전체 스위트가 강제된다. **순서가 중요하다** — 마크다운만 추가하는 변경은 가드를 한 번도 실행시키지 않는다.

---

## §D. 교차 렌즈 모순과 오케스트레이터 직접 실측

### D.1 [RESOLVED] Go 카운트 상수 개수 — Lens B 3개 vs Lens C 1개

Lens B가 옳다. **3개**다. 형제 Epic SPEC 2건이 동일 표면을 만나므로 검증 명령과 함께 기록한다.

```
internal/template/catalog_tier_audit_test.go:166   const expectedSkillCount = 31
internal/template/catalog_loader_test.go:63        const expectedTotal      = 41
internal/template/embed_catalog_test.go:54         const wantTotal          = 41
```

교차 검증:

```bash
git ls-files 'internal/template/templates/.claude/skills/*/SKILL.md' | wc -l   # → 31
grep -c '^\s*- name: ' internal/template/catalog.yaml                          # → 41
```

두 값이 상수와 정확히 일치한다. `expectedSkillCount`는 디스크상의 스킬 디렉터리 수, 나머지 둘은 catalog 전체 엔트리 수(스킬 + 에이전트 + 기타)를 센다.

가드 테스트 함수명(실측, `-run` 선택자 공허 통과 방지용):

- `TestAllSkillsInCatalog`, `TestCatalogNoDuplicateEntries`, `TestManifestHashFormat`, `TestCatalogTierValid` — `catalog_tier_audit_test.go`
- `TestLoadCatalog` — `catalog_loader_test.go`
- `TestLoadEmbeddedCatalog_Success` — `embed_catalog_test.go`
- `TestTemplateNoInternalContentLeak`, `TestSkillBodyLeakClassRecurrenceBackstop` — `internal_content_leak_test.go`
- `TestTemplateNeutralityAudit`, `TestTemplateNeutralityAuditC8Preserve` — `template_neutrality_audit_test.go`
- `TestEmbeddedMoaiSkillNames` — `skills_manifest_test.go`

### D.2 [RESOLVED, 정정] 라이선스 부재 근거의 정밀도

브리프는 `find ~/.agents/skills -iname 'LICENSE*'` → 결과 0건이라고 기술했다. 실측 결과는 **1건**이다.

```
/Users/goos/.agents/skills/notion-cli/LICENSE.md   # MIT License, Copyright (c) 2026 Notion Labs, Inc.
```

그러나 그 1건은 **무관한 다른 스킬**(`notion-cli`)의 것이고, 문제의 원문이 속한 배포에는 라이선스가 없다:

```bash
find ~/.agents/skills/higgsfield-websites -iname 'LICENSE*' -o -iname 'NOTICE*' -o -iname 'COPYING*' | wc -l   # → 0
```

**결론**: 클린룸 재작성 근거는 그대로 유효하다(해당 원문에 적용되는 라이선스 부여가 없다). 다만 근거 문장은 "스킬 루트 전역 0건"이 아니라 "**해당 배포 디렉터리에 라이선스·NOTICE·COPYING이 0건**"으로 정정되어야 한다. 형제 SPEC 2건도 원문별로 이 좁힌 형태의 확인을 다시 수행해야 한다.

### D.3 [신규 발견] docs-site skill-guide의 기존 4-로케일 드리프트

`moai-ref-ui-polish`(가장 최근 추가된 ref 스킬)의 등록 표면이 **불완전**하다. 실측:

| 표면 | `moai-ref-secops` | `moai-ref-ui-polish` |
|---|---|---|
| `catalog.yaml` | 있음 | 있음 |
| Go 카운트 상수 3개 | 반영 | 반영 |
| `delegation.yaml`(로컬·템플릿) | 있음 | **없음** |
| 워크플로 본문(`review.md`, `sync/quality-gates-quality.md`) | 있음 | **없음** |
| `skills_manifest_test.go` 스팟체크 | 있음 | **없음** |
| docs-site skill-guide | en·ja·zh·ko 4개 전부 | **ko 1개만** |

즉 Lens B가 열거한 등록 표면 전체를 채운 선례는 `moai-ref-secops`이고, `moai-ref-ui-polish`는 표면 절반을 비운 채 배포되었다.

추가로 skill-guide 산문의 **스킬 총계 주장 자체가 로케일 간 불일치**한다(실측):

| 로케일 | 카테고리 총계 주장 | 근거 |
|---|---|---|
| en | 27 (Foundation 4 + Workflow 8 + Domain 5 + Reference 8 + Meta/Harness 2) | `en/advanced/skill-guide.md:55` |
| ja | 27 (동일 분해) | `ja/.../skill-guide.md:57` |
| zh | 27 (동일 분해) | `zh/.../skill-guide.md:56` |
| ko | 30 (Foundation 4 + Workflow 9 + Domain 6 + Reference 9 + Meta/Harness 2) | `ko/.../skill-guide.md:57` |

디스크 실측(31)과도, 서로와도 맞지 않는다. 별개로 4개 로케일 모두 "naive approach: 31 skills" 문장은 일치한다.

이 드리프트는 **본 SPEC이 만든 것이 아니라 선행 부채**다. 그러나 신규 ref 스킬 행을 추가하면 각 로케일의 표와 산문 총계가 다시 어긋나므로, 어디까지 정정할지 결정이 필요하다 → plan.md §B.5 결정 기록(확정: 표 행만 추가하고 산문 카운트는 손대지 않는다).

### D.4 [RESOLVED] tier 배정 선례 직접 실측

```
moai-ref-ui-polish       → optional-pack:frontend
moai-ref-react-patterns  → optional-pack:frontend
moai-ref-git-workflow    → core
```

### D.5 [부분 해소] 접근성 중복 — `moai-ref-ui-polish` 직독 결과

open question 3은 "추정하지 말고 실제로 읽어서 중복을 확인하라"를 요구했다. `moai-ref-ui-polish/SKILL.md` 184줄 전체 H2 구조와 접근성 인접 토큰을 실측했다.

`moai-ref-ui-polish`의 H2: Target Agents / Core Philosophy / Geometry and Alignment / Elevation and Structure / Motion / Typography / Imagery / Interaction / Icons / Review Modes + 불변 3종.

접근성 인접 토큰 실측 (`focus|aria|alt |keyboard|contrast|accessib|heading` 대소문자 무시, 8행 매칭):

| 매칭 위치 | 내용 | Lens A Tier 1 접근성 항목과의 관계 |
|---|---|---|
| L59 | 테두리가 "selected/focus state"를 담당 | 시각적 포커스 **스타일링**만 — 키보드 조작 가능성 아님 |
| L72 | 모션은 유일한 피드백 채널이 되어서는 안 됨(접근성 언급) | 인접하나 SEO 감사 항목 아님 |
| L87-88 | tabular-nums, text-wrap: balance(headings) | 타이포그래피 — heading **레벨 구조**와 무관 |
| L129 | severity HIGH 정의에 "inaccessible" 어휘 | 등급 정의어일 뿐 |
| L179 | Verification 체크 1건(타이포그래피) | 무관 |

**판정**: Lens A가 지목한 Tier 1 접근성-as-SEO 항목 6종(단일 H1 + 레벨 건너뛰기 없음 / 이미지 alt / 서술적 anchor text / 라벨링된 폼 컨트롤 / 가시적 포커스를 동반한 **키보드 조작 가능성** / 살아있는 in-page fragment 타깃) 중 `moai-ref-ui-polish`가 다루는 것은 **0건**이다. 가장 가까운 L59도 포커스 링의 시각 처리이지 조작 가능성 보장이 아니다.

즉 **실질 중복은 없다**. 남은 것은 중복이 아니라 **소유권 경계 결정**이다(semantic HTML/a11y를 SEO 스킬이 갖는가, 별도 표면이 갖는가) → plan.md §B.3 결정 기록(확정: SEO 인과 4종은 포함, 조작성 3종은 `NOT for:` 절로 형제 표면에 위임). 이 조사는 "중복이 있는지 모른다"가 아니라 "중복은 없음을 확인했고 경계만 정하면 된다"는 좁혀진 상태로 결정 라운드에 넘겨졌다.

---

## §E. 조사에서 도출된 저작 제약 요약

1. Tier 3 수치는 값이 아니라 **결정 규칙**으로 재서술한다("픽셀 절단을 실측해 확인" 등).
2. 출처 없는 수치 3건은 어떤 형태로도 재현하지 않는다.
3. 배포 SKILL.md 본문에 `REQ-`/`AC-` 형태 토큰과 날짜를 넣지 않는다(frontmatter `updated:` 제외).
4. 프레임워크 이름은 동일 깊이 복수 나열 또는 무나열. 우위 표현 금지("<lang> is primary/the default", "only <lang> is supported").
5. `GOOS=` 토큰 추가 금지.
6. `references/` 하위 디렉터리를 만들지 않는다(ref 스킬 선례 0건).
7. Go 카운트 상수 변경을 마크다운과 **동일 변경**에 포함시켜 CI docs-only 스킵을 방지한다.
8. 등록 표면은 `moai-ref-secops`를 선례로 삼는다(`moai-ref-ui-polish`는 불완전 선례).
