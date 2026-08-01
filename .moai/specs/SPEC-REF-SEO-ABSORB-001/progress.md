# Progress — SPEC-REF-SEO-ABSORB-001

## §E.1 Plan-phase Audit-Ready Signal

- 산출 아티팩트: `research.md`, `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M + research.md)
- SPEC ID 정규식 사전 점검: 실행됨, 출력 `PASS`
- 요구사항: GEARS 표기 **23건** (REQ-SEO-001..009 콘텐츠 9 / 010..015 프로토콜 6 / 020..027 배포 8). 016-019는 **의도적 블록 결번**이며 spec.md §B 도입부에 관례로 명시되어 있다. 검증: `grep -o 'REQ-SEO-[0-9]\{3\}' spec.md | sort -u | wc -l` → `23`
- 수용 기준: **24건**, 전부 명령 + 기대 출력 + 실패 형태 명시. 검증: `grep -c '^### AC-SEO-' acceptance.md` → `24`
- 기계 판정 AC가 없는 REQ: **4건**(010 / 014 / 026 / 027) — 전부 §E DoD에 REQ 번호 명시 체크박스 보유. 전수 매핑은 `acceptance.md` §D.1
- 결정 게이트: `plan.md` §B 5건 **전부 확정**(tier=`optional-pack:frontend` / 감사 게이트 중간안 / 접근성 SEO 인과 4종만 / GEO 전면 제외 / docs-site 표 행만 추가). 검증: plan.md·research.md 양쪽 clarification 마커 grep → 각 `0`
- plan-phase 실측 baseline: skill dirs 31, catalog entries 41, Go 상수 31/41/41, 원문 861줄 · `sha256 c088f089…98ee7`, self-trip 8-gram 4600, 대조 skill 8-gram 0, LCS 17-23자, `GOOS=` 보유 템플릿 파일 2, ref 스킬 `## Verification` 체크박스 6-9개, catalog tier 분포 core 29 / devops 4 / backend 3 / frontend 3 / design 1 / harness-generated 1
- 상태: `draft`. Implementation Kickoff Approval 이전에 plan-audit 재감사 필요.

### plan-audit 이력

| iteration | verdict | score | 처리 |
|---|---|---|---|
| 1 | FAIL | 0.72 (Tier M 통과선 0.80) | MUST-FIX 6건 + SHOULD-FIX 일부를 v0.2.0에서 정정, 결정 5건 확정 |
| 2 | FAIL | 0.79 (0.01 미달) | iteration-1 MUST-FIX 6건 **전부 CLOSED 확정** — 재검증 대상 아님. FAIL은 v0.2.0 신규 콘텐츠의 1차 감사 결함 3건(N1/N2/N3)에 의한 것. v0.3.0에서 N1-N3 + SHOULD-FIX N4-N7 + SF-10 정정 |
| 3 | **PASS** | **0.87** (통과선 0.80) | must-pass 7/7, 회귀 0. N1-N3 전부 CLOSED — 감사가 24개 AC 본문에서 REQ→AC 매핑을 독립 재도출해 §D.1의 5건 오프셋·미판정 4건·팬텀 AC 0건을 재현 확인. 잔여 SHOULD-FIX 3건 중 **N10 즉시 정정**, **N8·N9는 사용자 결정으로 run-phase M6 이연** |

v0.3.0 정정 요약 — **N1** AC-SEO-015의 표 헤더 판독 명령 2건이 `grep -- '---' -B1` 옵션 파싱 오류로 실행되지 않던 문제(`-B1`이 파일명으로 해석 → exit 1 + 출력 0행). `-B1`을 `--` 앞으로 옮기고, 원문 측 실측 baseline(섹션 7행 / 표 헤더 10행)과 **네 `count`가 모두 1 이상일 때에만 판정**하는 전제 검사를 추가. **N2** §D.1 추적성 표를 내용 기준 전수 매핑으로 재작성 — §B.2 번호 오프셋(010→011 / 011→012 / 012→013) 명시, 기계 판정 AC 없는 REQ를 2건→**4건**(010/014/026/027)으로 정정, 재부여로 해소하지 않는 이유(다대다 관계라 1:1 불가) 기록. 같은 거짓에 기대던 spec.md §B 번호 관례 근거 문장도 실제 매핑에 맞게 재작성. **N3** AC-SEO-013 Then절 `60자`→`40자`(기대·실패 절과 통일). **N4** AC-SEO-002 H2 상한 근거 정정(ui-polish 실측 13, v0.2.0 기재 14는 research §B.4의 과다 계상 상속). **N5** AC-SEO-005 정규식 `reference[:,]`(9건 중 4건 통과)→`reference\b`(9/9), `NOT for:` 절 **내용** 판정 토큰 2종 신설. **N6** research.md §B.5 skill-routing 표면 열거에 실측 정정 각주. **N7** 접근성 6 vs 7 산술 분해 기록. **SF-10** REQ-SEO-027 라벨 `(Event-detected)`→`(Event-driven)`.

v0.2.0 정정 요약 — MF-1 AC-SEO-025가 언어 중립성 가드를 실행하지 않던 문제(강제 파일·`-run` 선택자 오지정), MF-2 AC-SEO-022의 secops 표면 집합 역전(skill-routing 제거 + 워크플로 본문 4건 추가), MF-3 접근성 결정과 REQ-SEO-006의 정면 충돌(개념 4종 편입 + AC 토큰 루프 확장), MF-4 요구사항 개수 오기재(21→23), MF-5 미해소 마커 5건 해소, MF-6 REQ 번호 결번 관례 명시. 추가로 클린룸 판정 대상 트리를 템플릿으로 통일(+AC-SEO-011b), 원문 digest 고정, 구조 발산 판정 신설(AC-SEO-015), tier 유효성 AC 신설(AC-SEO-020c), AC-SEO-006 공허 토큰 제거, AC-SEO-013 임계값 근거 부여, AC-SEO-002 H2 상한 정정, AC-SEO-021b 브랜치 전제 명시, 판정 AC가 없던 REQ-SEO-009에 AC-SEO-009 신설.

**v0.3.1 추가 정정 (N10)**: `research.md` §B.4의 `moai-ref-ui-polish` H2 수치를 `11개` → `명명 10개 + 불변 3 = 총 13개`로 정정하고 실측 각주를 달았다(N6와 동일 처리 — 조사 결과 본문 유지, 수치와 각주만 정정). 이 오기가 v0.2.0 AC-SEO-002의 근거 문장으로 전파되었던 것이 N4였다. **research.md는 형제 SPEC 2건이 재사용하도록 설계된 문서이므로 발원지 정정이 필수다** — acceptance.md만 고치면 아직 착수되지 않은 형제 작업에 같은 수치가 다시 흘러간다.

**run-phase M6 이연 (사용자 결정, 이번 개정에서 손대지 않음)**:

- **N8** — §D.1은 REQ-SEO-020을 완전 판정으로 제시하나, `catalog_tier_audit_test.go:278-279`가 `TrimPrefix`/`TrimSuffix`로 `templates/` 접두사와 후행 슬래시를 벗겨내므로 기계 판정되는 것은 `tier` 뿐이다.
- **N9** — AC-SEO-004는 REQ-SEO-004의 6개 frontmatter 필드 중 길이 절만 판정한다. `name` / `user-invocable` / `metadata` / `progressive_disclosure`가 전부 부재해도 스크립트는 정상 종료한다.

두 건 모두 AC 판별력 공백이며, **수정이 실제로 동작하는지는 `moai-ref-seo/SKILL.md`가 존재해야만 관측 가능**하다. 산출물 없이 AC를 고치면 v0.2.0에서 반복한 "실행해 보지 않은 판정 명령" 패턴을 재생산하므로, M6 검증 단계에서 산출물을 좌변에 놓고 함께 닫는다.

**forward-reference 정합 (해소)**: `research.md` §D.3·§D.5의 두 forward-reference가 v0.1.0 plan.md 마커를 가리키고 있었다. 마커 치환으로 무효가 되었으므로 각각 plan.md §B.5·§B.3 **결정 기록**을 가리키도록 포인터 절만 정정했다(조사 결과 본문은 무수정 — 감사에서 정확성이 확인된 부분이다). 5개 아티팩트 전부 잔여 clarification 마커 0건이며, 재감사의 MP-7 grep은 `plan.md`·`research.md` 양쪽에서 0을 반환한다.

## §E.2 Run-phase Evidence

### M3 — 스킬 본문 저작 (템플릿 트리)

산출물 1건: `internal/template/templates/.claude/skills/moai-ref-seo/SKILL.md` (207줄). 로컬 미러·catalog 엔트리·Go 상수·docs-site는 M4/M5 소관이므로 이번 마일스톤에서 손대지 않았다.

#### 판정 결과

| AC | 판정 | 명령 | 실측 출력 |
|---|---|---|---|
| AC-SEO-011 n-gram 중첩 | PASS | acceptance.md §B 8-gram 스크립트 (좌변 = 템플릿 트리) | `shared_8grams=0` (HIT 0행) |
| AC-SEO-012 self-trip + 원문 고정 | PASS | acceptance.md §B self-trip 스크립트 | `source_pin_match=True` / `selftrip_shared_8grams=4600` — 기대값 정확 일치 |
| AC-SEO-013 LCS 상한 | PASS | acceptance.md §B `SequenceMatcher` 스크립트 | `lcs_chars=27`, `lcs_text=' a non empty alt attribute '` (상한 40) |
| AC-SEO-014 무출처 수치 미재현 | PASS | `grep -nE '80%\|100ms\|100-300ms\|300ms'` | 산출물 `0` / 원문 `3` (양성 대조 성립) |
| AC-SEO-008 휘발성 수치 미탑재 | PASS | acceptance.md §C 휘발성 패턴 grep | 산출물 `0` / 원문 `changefreq` `7` (양성 대조 성립) |
| AC-SEO-009 플랫폼 종속 흐름 미포함 | PASS | `grep -inE 'marketplace\|listing card\|credit cost\|cover video\|sidecar'` | 산출물 `0` / 원문 `7` (양성 대조 성립) |
| AC-SEO-015 구조 발산 | PASS | acceptance.md §B 4-count 판독 명령 | 산출물 H2 `12` / 원문 섹션 `7` / 산출물 표 헤더 `8` / 원문 표 헤더 `10` — 네 count 전부 1 이상, 원문 측 baseline 7·10 일치 |
| AC-SEO-002 분량·구조 | **PARTIAL** | `wc -l` / `grep -c '^# '` / `grep -c '^## '` | 줄 수 `207`(150-220 ✔), H2 `12`(7-14 ✔), **H1 `2`** — 아래 선택자 결함 참조 |
| AC-SEO-003 불변 3종 + zone id | PASS | acceptance.md §C zone·헤딩 grep | zone id 3행(165/179/195), 헤딩 3행(166/180/196) — Rationalizations < Red Flags < Verification 순서 증가 |
| AC-SEO-004 frontmatter 길이 | PASS | acceptance.md §C 길이 스크립트 | `desc_plus_when_chars=1434` (상한 1536) |
| AC-SEO-005 description 3악장 | PASS | acceptance.md §C 5-토큰 스크립트 | 다섯 값 전부 `True` (`ends_reference_phrase` / `has_amplify` / `has_not_for` / `not_for_names_geo` / `not_for_names_a11y`) |
| AC-SEO-006 개념 커버리지 12종 | PASS | acceptance.md §C 토큰 루프 | canonical 9 / robots.txt 6 / sitemap.xml 6 / json-ld 6 / meta description 4 / title 4 / entity 7 / redirect 10 / heading 4 / alt 5 / anchor text 2 / fragment 5 — 12종 전부 1 이상 |
| AC-SEO-020 leak 가드 | PASS | `go test ./internal/template/ -run '...' -count=1 -v` | `--- PASS: TestTemplateNoInternalContentLeak` / `--- PASS: TestSkillBodyLeakClassRecurrenceBackstop` 두 행 출력에 등장 |
| AC-SEO-025 언어 중립성 | PASS | acceptance.md §D 4-테스트 선택자 + `GOOS=` 카운트 | `--- PASS:` 4행 전부 등장(`TestLanguageNeutrality` 포함), `ok` 종료, `GOOS=` 보유 템플릿 파일 `2` |
| AC-SEO-001 단일 파일 구성 | **PARTIAL** | `find internal/template/templates/.claude/skills/moai-ref-seo -type f` | 템플릿 트리 1행(`SKILL.md`)만 — `references/`·`modules/` 미생성. 로컬 미러는 M4 소관이라 2행 조건은 M4에서 성립 |

#### AC-SEO-002 H1 선택자 결함 (N8/N9 계열, 산출물 결함 아님)

`grep -c '^# '`는 frontmatter 안의 YAML 주석 `# MoAI Extension: Progressive Disclosure`를 함께 센다. 기존 ref 스킬 9건 전수 실측:

```
moai-ref-api-patterns h1_grep=2 h1_body_only=1   moai-ref-git-workflow   h1_grep=2 h1_body_only=1
moai-ref-llm-security h1_grep=2 h1_body_only=1   moai-ref-owasp-checklist h1_grep=2 h1_body_only=1
moai-ref-react-patterns h1_grep=2 h1_body_only=1 moai-ref-secops         h1_grep=2 h1_body_only=1
moai-ref-supply-chain h1_grep=2 h1_body_only=1   moai-ref-testing-pyramid h1_grep=2 h1_body_only=1
moai-ref-ui-polish    h1_grep=2 h1_body_only=1   moai-ref-seo(신규)      h1_grep=2 h1_body_only=1
```

10건 전부 `h1_grep=2` / `h1_body_only=1`이다. 즉 **현행 선택자로는 어떤 적합한 ref 스킬도 `1`을 낼 수 없다.** 산출물은 실제 요구사항(본문 단일 H1)을 만족하며 선례 9건과 동일 형태다. 본문에서 `# MoAI Extension:` 주석을 빼면 선택자는 통과하지만 저작 표준(research.md §B.1)과 선례 9건에서 이탈하므로 그렇게 하지 않았다. AC 본문 수정은 manager-spec 소관이므로 M6 판정 보강 대상으로 넘긴다(N8/N9와 동일 처리).

#### 구조 발산 판정 (AC-SEO-015, 사람 판독)

`structural_divergence: PASS`

**질문 1 — 섹션 순서가 1:1 대응하는가? 대응하지 않는다.** 원문은 생산자 워크플로 순서(메타데이터 작성 → 기술 인프라 → 스키마 → 엔티티 → 생성형 검색 → 말미 감사)를 따르는 6개 최상위 섹션이다. 산출물은 **문서를 기계가 파싱하는 순서**를 조직 원리로 삼아 Document Semantics를 맨 앞 도메인 섹션에 두었다 — 원문은 같은 항목들을 맨 뒤 감사 체크리스트 안에 흩어 놓았으므로 위치가 정반대다. robots/sitemap은 원문에서 2번째 섹션(Technical SEO) 안에 있으나 산출물에서는 뒤에서 두 번째 독립 섹션이다. 산출물에는 원문의 생성형 검색 섹션과 감사 섹션에 대응하는 섹션이 아예 없고(결정 §B.2·§B.4에 따른 제외), 원문에는 산출물의 Core Principle·Document Semantics·Delivery Chokepoints에 대응하는 최상위 섹션이 없다. 유일하게 상대 순서가 같은 인접쌍은 Structured Data → Entity 2건인데, 이는 엔티티 서술이 스키마 위에 얹히는 실제 의존 관계에서 나온 것이며 7개 중 2개의 국소 일치는 순서의 1:1 대응이 아니다.

**질문 2 — 표 열 구성이 1:1 대응하는가? 대응하지 않는다.** 원문 표 헤더 10개 중 6개가 동일한 `Field | Value` 2열 반복(스키마 타입별 필수 필드 나열)이고, 나머지는 `Page type | robots value`, `Intake field | Meta target`, `Site type | Schema types to apply`, `Field | Example | Required`다. 산출물 표 헤더 8개는 `Rule | How to check it | Failure it prevents`, `Decision | Rule`×2, `Field | Rule | Recurring defect`, `Surface | Requirement`, `Artifact | Rule`, `Concern | Rule`, `Rationalization | Reality`로, **원문의 `Field | Value` 반복 패턴을 한 번도 쓰지 않는다.** 원문이 타입별 필드 나열에 6개 표를 쓴 자리를 산출물은 Structured Data 단일 표의 `Required fields` 행 1줄로 접었다(결정 §B.2의 압축 방침과 정합). 열 축도 다르다 — 원문은 대체로 "항목 → 값" 조회 매트릭스이고, 산출물은 "결정 → 규칙" 또는 "규칙 → 검증 방법 → 예방되는 실패"의 판정 축이다.

#### 임계값 조정 금지 준수 (REQ-SEO-014)

8-gram HIT는 처음부터 0이었으므로 재작성을 강제한 HIT는 없다. 다만 **1차 저작본의 LCS가 39자(상한 40)로 상한에 1자 차이까지 붙었고, 일치 구절이 `' profiles the entity actually controls '`로 원문 문장과 사실상 동일했다.** 기계 임계값은 통과 상태였으나 어휘 연속성이 곧 파생성의 신호이므로 해당 행을 재작성했다(`Only profiles the entity actually controls, ...` → `Restricted to accounts this entity itself administers, ...`). 재실행 결과 `lcs_chars` 39 → **27**, 일치 구절은 도메인 필수 어휘(`a non empty alt attribute`)로 바뀌었고 8-gram은 0을 유지했다. 임계값은 조정하지 않았다.

#### M3 범위 밖 기대 실패 (M4 소관, 이번에 고치지 않음)

```
catalog_tier_audit_test.go:168: expected 31 skill directories on disk, found 32
--- FAIL: TestAllSkillsInCatalog
```

디스크 스킬 수가 31 → 32로 늘었으나 `expectedSkillCount` 상수는 31 그대로다. catalog 엔트리·`make build` 해시·Go 상수 3개는 M4 단일 변경 소관이므로 **의도된 M3 시점 실패**다. `catalog.yaml`과 상수 파일은 건드리지 않았다. `TestLoadCatalog` / `TestLoadEmbeddedCatalog_Success`는 이 시점에도 PASS다(카운트가 catalog 엔트리 수 기준이라 마크다운 추가만으로는 흔들리지 않는다).

### M4 — 등록 표면 채우기 + 카운트 상수 (단일 변경)

12개 파일 단일 커밋. `catalog.yaml` 엔트리는 placeholder 해시로 먼저 적고 `make build`(`gen-catalog-hashes --all`) 산출물만 커밋했다 — 수기 해시 없음. docs-site 4-로케일은 M5 소관이라 손대지 않았다.

#### 표면 6종 실측 (`grep -c 'moai-ref-seo'`)

| 표면 | 파일 | 카운트 |
|---|---|---|
| 1. catalog 엔트리 | `internal/template/catalog.yaml` | 2 (name + path) |
| 2. Go 상수 3개 | `catalog_tier_audit_test.go` / `catalog_loader_test.go` / `embed_catalog_test.go` | 1 / 1 / 1 |
| 3. delegation 양 트리 | 로컬 + 템플릿 `delegation.yaml` | 1 / 1 |
| 4. 워크플로 본문 4건 | `review.md` / `sync/quality-gates-quality.md` + 템플릿 미러 2건 | 1 / 1 / 1 / 1 |
| 5. 스킬 매니페스트 스팟체크 | `internal/template/skills_manifest_test.go` | 1 |
| 6. 로컬 미러 | `.claude/skills/moai-ref-seo/SKILL.md` | 1 |

`.claude/rules/moai/workflow/skill-routing.md`는 계획대로 건드리지 않았다(secops 부재 실측 근거).

#### 판정 결과

| AC | 판정 | 명령 | 실측 출력 |
|---|---|---|---|
| AC-SEO-001 단일 파일 구성 | **PASS** (M3 PARTIAL 해소) | `find … -type f \| sort` | 2행 — 양 트리 각각 `SKILL.md` 1개뿐 |
| AC-SEO-011b 로컬·템플릿 동일성 | PASS | `diff` + `shasum -a 256` | `IDENTICAL`, 양쪽 `39d3ad2b348dc0748c7baa5912dba34865d3321b7c5a4b990cc047cb5a131065` |
| AC-SEO-020b catalog 정합 | PASS | `go test -run '<7개 선택자>' -count=1 -v` | 요구된 7개 테스트 전부 `--- PASS:` 행으로 등장, `ok` 종료 |
| AC-SEO-020c tier 유효성·워크플로 커버리지 | PASS | tier grep + `TestCatalogTierValid\|TestWorkflowTriggerCoverage` + `required-skills` grep | `tier: optional-pack:frontend` 4건, 신규 엔트리 tier 확인, 두 테스트 `--- PASS:`, `required-skills` `0` (전제 유지) |
| AC-SEO-021 상수 증가·지상 진실 일치 | PASS | 상수 3 grep + `git ls-files … \| wc -l` + `grep -c '- name: '` | `32` / `42` / `42` / 디스크 `32` / catalog 엔트리 `42` — 기대값 전건 일치 |
| AC-SEO-024 템플릿 패키지 회귀 | PASS | `go test ./internal/template/...` | `ok … 15.431s` |
| AC-SEO-025 언어 중립성 | PASS | 4-테스트 선택자 | `TestLanguageNeutrality` 포함 4행 전부 `--- PASS:` |
| AC-SEO-020 leak 가드 | PASS | `TestTemplateNoInternalContentLeak` | `--- PASS: TestTemplateNoInternalContentLeak (5.83s)` |

#### 선택자 공허 통과 방지 (plan.md §G 안티패턴)

최초 선택자 `-run 'TestCatalog|TestEmbeddedCatalog|TestSkillsManifest'`는 5개 테스트만 매칭했고 **카운트 상수 3개를 보유한 테스트를 하나도 실행하지 않았다** — exit 0이지만 무효. 실제 보유 함수는 `TestAllSkillsInCatalog`(`expectedSkillCount`) / `TestLoadCatalog`(`expectedTotal`) / `TestLoadEmbeddedCatalog_Success`(`wantTotal`)이며, 선택자를 정정한 뒤 세 행이 모두 `--- PASS:`로 등장함을 확인했다.

추가로 **반증 왕복**을 수행했다: `expectedSkillCount`를 32 → 31로 되돌리자

```
--- FAIL: TestAllSkillsInCatalog (0.00s)
    catalog_tier_audit_test.go:170: expected 31 skill directories on disk, found 32
```

가드가 실제로 카운트를 검사함이 확인되었고 상수를 즉시 복원했다. 통과가 도달성 없는 공허 통과가 아님을 증명한 유일한 증거다.

#### 빌드·린트

| 항목 | 결과 |
|---|---|
| `go build ./...` | exit 0 |
| `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| `gofmt -l <수정한 4개 test 파일>` | 무출력 |
| `golangci-lint run` | `0 issues.` (변경 전 baseline도 `0 issues.` — 신규 지적 0건) |

### M5 — docs-site 4-로케일 (커밋 `e01029fc4`)

en·ja·zh·ko `advanced/skill-guide.md` 레퍼런스 표에 `moai-ref-seo` 행을 1개씩 추가. 4파일 4행 삽입, 그 외 무변경.

**산문 카테고리 총계는 손대지 않았다** — plan.md §B.5 결정(표 행만 추가) 준수. 결과적으로 en/ja/zh는 27을, ko는 30을 계속 주장하고 디스크 실측은 32다. 이 드리프트는 본 SPEC이 만든 것이 아니라 선행 부채이며(research.md §D.3), 카테고리 분해 재계산을 동반하므로 본 SPEC 범위 밖이다.

### M6 — 검증 (AC 24건 전수 실행)

판정은 오케스트레이터가 직접 실행해 관측했다. 아래는 그 결과 기록이다.

#### §B 클린룸 프로토콜 (7건)

| AC | 판정 | 실측 출력 |
|---|---|---|
| AC-SEO-010 출처 라이선스 부재 | PASS | 배포 디렉터리 `0` / 상위 스킬 루트 `1` — 양성 대조 성립(명령이 무력하지 않음) |
| AC-SEO-011 8-gram 중첩 | PASS | `shared_8grams=0` |
| AC-SEO-011b 로컬·템플릿 동일성 | PASS | `IDENTICAL`, 양쪽 sha256 `39d3ad2b348dc0748c7baa5912dba34865d3321b7c5a4b990cc047cb5a131065` |
| AC-SEO-012 self-trip + 원문 고정 | PASS | `source_pin_match=True`, `selftrip_shared_8grams=4600` — 기대값 정확 일치(검사기 자체가 살아 있음을 증명) |
| AC-SEO-013 LCS 상한 | PASS | `lcs_chars=27` (상한 40), `lcs_text=' a non empty alt attribute '` |
| AC-SEO-014 무출처 수치 미재현 | PASS | 산출물 `0` / 원문 `3` |
| AC-SEO-015 구조 발산 | PASS | 아래 별도 판정 기록 |

#### §C 콘텐츠 (8건)

| AC | 판정 | 실측 출력 |
|---|---|---|
| AC-SEO-001 단일 파일 구성 | PASS | `find` 2행 — 양 트리 각 `SKILL.md` 1개 |
| AC-SEO-002 분량·구조 | PASS (N12 정정 후) | `207` 줄 / H2 `12` / `body_h1=1` — 세 측정 전부 범위 내 |
| AC-SEO-003 불변 3종 + zone id | PASS | zone id 3행(165·179·195), 헤딩 3행(166 < 180 < 196) 순서 증가 |
| AC-SEO-004 frontmatter 길이 | PASS | `desc_plus_when_chars=1434` (상한 1536) |
| AC-SEO-005 description 3악장 | PASS | 5개 토큰 값 전부 `True` |
| AC-SEO-006 개념 커버리지 12종 | PASS | canonical 9 / robots.txt 6 / sitemap.xml 6 / json-ld 6 / meta-description 4 / title 4 / entity 7 / redirect 10 / heading-structure 4 / image-alt 5 / anchor-text 2 / fragment 5 — 12종 전부 ≥1 |
| AC-SEO-008 휘발성 수치 미탑재 | PASS | 산출물 `0` / 원문 `7` |
| AC-SEO-009 플랫폼 종속 흐름 미포함 | PASS | 산출물 `0` / 원문 `7` |

#### §D 배포 가드 (9건)

| AC | 판정 | 실측 출력 |
|---|---|---|
| AC-SEO-020 leak 가드 | PASS | `--- PASS: TestTemplateNoInternalContentLeak` |
| AC-SEO-020b catalog 정합 | PASS | 요구된 7개 테스트 전부 `--- PASS:` |
| AC-SEO-020c tier 유효성 | PASS | `tier: optional-pack:frontend`, `TestCatalogTierValid`·`TestWorkflowTriggerCoverage` `--- PASS:`, `required-skills` `0` |
| AC-SEO-021 카운트 상수 | PASS | `32` / `42` / `42` = 디스크 `32` · catalog `42` |
| AC-SEO-021b go+md 동일 변경 | PASS | `.go`·`.md` 양쪽 변경집합에 존재 |
| AC-SEO-022 등록 표면 6종 | **PASS (M5로 완성)** | `moai-ref-secops` 15건 = `moai-ref-seo` 15건 — **누락 0**. M4 시점 11/15(docs-site 4건 미충족)였고 M5가 나머지 4건을 채웠다 |
| AC-SEO-023 docs-site 4로케일 | PASS | en·ja·zh·ko 전부 `1` |
| AC-SEO-024 템플릿 패키지 회귀 | PASS | `ok github.com/modu-ai/moai-adk/internal/template 8.571s` |
| AC-SEO-025 언어 중립성 | PASS | `TestLanguageNeutrality`·`TestSkillBodyNoLangReference` 둘 다 `--- PASS:` — 본 SPEC 1순위 위험이 검증됐다는 유일한 증거 |

**집계: 24건 PASS / 0건 FAIL.**

#### AC-SEO-015 구조 발산 판정 기록 (사람 판독)

acceptance.md가 "근거 없는 PASS 한 줄은 판정으로 인정하지 않는다"고 명시하므로 두 질문에 대한 근거를 함께 남긴다.

전제 검사: 원문 `count=7`(H1/H2) · `count=10`(표 헤더) — acceptance.md에 고정된 baseline과 일치. 산출물 `count=12` · `count=8`. 네 count 모두 1 이상이므로 판정 성립.

`structural_divergence: PASS`

**질문 1 — 섹션 순서가 원문과 1:1 대응하는가? 대응하지 않는다.** 원문은 최상위 H2 6개(Meta tags & OG → Technical SEO → Schema markup → Entity SEO → GEO/content → Audit)인데 산출물은 도메인 H2 9개 + 불변 3개로 재분할됐다. 원문 1위 `Meta tags & OG`는 산출물에서 5위 `Per-Page Metadata`로 밀렸고, 원문 2위 `Technical SEO`는 단일 대응 섹션 없이 `Host-Derived Crawl Artifacts`(8위)와 `Delivery Chokepoints`(9위)로 분해됐다. 산출물 앞머리의 `Document Semantics`·`Identity and Canonical Address`는 원문에 대응하는 최상위 섹션이 없다(원문은 canonical을 Technical SEO 하위에 배치). 원문 5위 `GEO / content`는 의도적으로 제외됐다.

**질문 2 — 표 열 구성이 원문 대응 표와 1:1 대응하는가? 대응하지 않는다.** 원문 헤더 10건은 `Field | Value` 반복 6건이 중심인 *값 나열* 구조다. 산출물 8건은 `Decision | Rule`, `Surface | Requirement`, `Artifact | Rule`, `Concern | Rule`, `Rationalization | Reality` 등 *규칙 중심* 구조다. 어휘가 겹치는 `Field`조차 원문은 2열 `Field | Value`, 산출물은 3열 `Field | Rule | Recurring defect`로 열 수와 의미가 다르다.

#### N12 — AC-SEO-002 판정 명령 공허 (산출물 결함 아님)

M3 시점 AC-SEO-002가 H1=2로 PARTIAL이었던 원인은 **산출물이 아니라 판정 명령이었다.** `grep -c '^# '`를 파일 전체에 적용해 frontmatter의 `# MoAI Extension: Progressive Disclosure` 주석 행을 H1로 계수한 것이다.

발견 근거는 선례 전수 실측이다 — 템플릿 트리 `moai-ref-*` **10건 전수가 H1=2**였다. 즉 이 명령은 신규 산출물뿐 아니라 **기존 선례 9건도 9/9 실패시키는 명령**이었고, 임계값 1을 만족할 수 있는 파일이 애초에 존재하지 않았다. frontmatter를 제거하면 10건 전부 H1=1이다.

처리: manager-spec이 커밋 `e3be03011`로 acceptance.md를 정정했다 — frontmatter 제거 후 계수하도록 명령을 바꾸고 **임계값 1은 유지**했다(plan.md §G의 "임계값을 올려 통과시키기" 안티패턴 회피 — 고친 것은 측정 방법이지 기준이 아니다). 정정된 명령을 파일에 적힌 그대로 실행한 결과 `207 / 12 / body_h1=1`로 **AC-SEO-002 PASS**.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-01
run_commit_sha: pending-backfill   # 이 커밋 자신 — 자기참조 불가, 커밋 후 백필
run_status: audit-ready
milestones: M2-M6 (M1 결정 게이트는 plan-phase 소관)
ac_pass_count: 24
ac_fail_count: 0
preserve_list_post_run_count: 0    # skill-routing.md 무변경(secops 부재 실측), SPEC 본문 3종 무변경
l44_pre_commit_fetch: run          # origin/main 선행 발산 관측 (수치는 변동 중 — 아래 잔여 위험 참조)
l44_post_push_fetch: not-run       # push 미수행 — 브랜치 protection(enforce_admins) 및 orchestrator 소관
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues (baseline도 0 issues)
cross_platform_build:
  host: pass                       # go build ./... exit 0
  windows_amd64: pass              # GOOS=windows GOARCH=amd64 go build ./... exit 0
full_test_suite: pass              # go test ./internal/template/... ok
total_run_phase_files: 17          # M3 템플릿 SKILL.md 1 + M4 등록 표면 12 + M5 docs-site 4
m1_to_mN_commit_strategy: M3 a525a236e / M4 184a325bc(단일 커밋, 마크다운+Go 상수 동시) / M5 e01029fc4 / N12 acceptance 정정 e3be03011(manager-spec 소관) / M6 판정 기록 this commit
```

미해소 위험 2건(차단 요소 아님):

1. **origin/main 선행 발산 (진행 중, 수치 고정 불가)** — `git rev-list --count --left-right origin/main...HEAD`가 M4 시점 `3 5`, M6 기록 시점 `5 8`로 관측됐다. **origin 쪽 수치는 본 SPEC 작업 중에도 계속 증가했다** — 처음 관측한 `#1268`/`#1269`/`#1270`에 더해 `#1272`/`#1271`이 추가로 머지됐다. 따라서 이 값은 시점 스냅샷이며 감사 시점에 재측정해야 한다. 본 SPEC이 만든 발산은 아니다(M3 커밋 `a525a236e`가 이미 `9ced435e9` 기준이었다). 브랜치+PR 시 rebase/merge 판단이 필요하다.
2. **접근성 조작성 항목 공백** — plan.md §F에 열린 채로 기록된 위험. 형제 SPEC이 아직 존재하지 않아 `depends_on:`을 걸 대상이 없으며 본 SPEC에서 완화되지 않는다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-02
sync_commit_sha: pending-backfill   # 이 커밋 자신 — 자기참조 불가, 커밋 직후 백필
sync_status: audit-ready
run_merge_commit_sha: 1594cf60e     # PR #1274 squash merge — origin/main 에서 실측 확인
b12_self_test_a: PASS               # grep -c 'SEO-ABSORB' CHANGELOG.md → 0 (중복 없음, emission 전 관측)
b12_self_test_b: PASS               # acceptance.md AC 제목 24건 (grep -cE '^### AC-SEO') == CHANGELOG 기재 24/24
b12_self_test_c: PASS               # CHANGELOG 인용 경로 전건 ls 확인 (SKILL.md 2종, catalog.yaml, Go 상수 3파일)
changelog_entry_position: "[Unreleased] → ### Added 최상단"
frontmatter_status_transitions:
  spec_md: "in-progress → completed"   # 단일 sync 커밋이 implemented 를 경유해 completed 로 종결
  updated_field: "2026-08-01 → 2026-08-02"
  phase_field: "plan → sync"
  note: "plan.md / acceptance.md / progress.md 는 frontmatter 미보유(본 SPEC 은 spec.md 만 frontmatter 를 가진다) — 실측 확인"
docs_surface_delta: 0                # M5 가 이미 4-로케일 skill-guide 행을 랜딩 — sync-phase 신규 문서 추가 없음
readme_delta: 0                      # README 4종은 개별 ref 스킬을 나열하지 않는다(moai-ref-secops 참조 0건 실측) — 동기화 대상 아님
canary_compliance_check: n/a         # 본 SPEC 은 미래지향 정책을 정의하지 않는다
sync_phase_verification:
  go_build: pass                     # go build ./... exit 0
  template_suite: pass               # go test ./internal/template/... → ok 4.638s
  mirror_parity: pass                # diff 로컬↔템플릿 SKILL.md → 차이 없음
  registration_surface_files: 17     # moai-ref-seo 참조 파일(SPEC 산출물·CHANGELOG 제외)
```

미검증 잔여 (차단 요소 아님):

1. **전체 테스트 스위트 미실행** — `go test ./internal/template/...` 만 실행했다. 본 sync 커밋은 마크다운·frontmatter만 바꾸므로 다른 패키지에 영향이 없다고 판단했으나, `go test ./...` 전량은 관측하지 않았다.
2. **크로스 플랫폼 빌드 미재실행** — run-phase(§E.3)에서 `GOOS=windows` 빌드 pass 를 기록했다. sync 커밋은 Go 코드를 바꾸지 않아 재실행하지 않았다.
3. **docs-site 빌드 미실행** — 4-로케일 skill-guide 행은 M5 에서 이미 머지됐고 본 커밋이 손대지 않았으므로 hugo 빌드를 재실행하지 않았다.
