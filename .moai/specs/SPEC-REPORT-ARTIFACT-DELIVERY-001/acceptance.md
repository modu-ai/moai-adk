# SPEC-REPORT-ARTIFACT-DELIVERY-001 — 인수 기준

> 이 문서는 status 축에서 stateless다 — `status:` 필드를 갖지 않는다.

## §A 검증 원칙

- 모든 AC는 기계 검증 가능 — 명령과 관측 결과를 함께 명명한다.
- `format=artifact`의 라이브 게시는 CI·테스트 환경에서 실행 불가(Artifact 도구 부재)이므로, AC는 **메커니즘**(계약 형상·폴백 트리거·미러 패리티·독트린 존재)을 검증한다 — 라이브 게시는 검증하지 않는다.
- **verification-completeness §2 이원 셀 규율**: 각 release-blocking AC는 RED-now 관측(2026-10-02, 트리 2e700a15e — 본 plan 단계에서 실측한 값)과 전환 마일스톤을 함께 운반한다. run 단계는 §2.1의 4요소(단일 호출 명령·축어 stdout·exit 코드·트리 SHA)를 각 AC의 검증 기록에 남긴다.
- 회귀 방어형 AC(현시점 이미 그린)는 "그린 착지"로 표기하고 깨뜨리지 않는 것을 목표로 한다.

## §B 시나리오 (Given-When-Then)

### AC-RAD-001 — settings 폐쇄 집합 (REQ-001)
**Given** settings 스키마가 report.format의 폐쇄 집합을 소유하고, **When** `grep -n 'reportFormatValues' internal/settings/schema_sections.go`를 실행하면, **Then** 값 목록에 `html+md`·`md`·`artifact` 세 값이 순서대로 존재한다.
- RED-now(2e700a15e 실측): `var reportFormatValues = []string{"html+md", "md"}` — artifact 부재. → M1에서 전환.
- 보조: `go test ./internal/settings/` 그린(신규 폐쇄 집합 단언 테스트 포함).

### AC-RAD-002 — i18n 4 로캘 키 (REQ-001)
**Given** i18n 카탈로그가 en/ko/ja/zh 4 블록으로 구성되고, **When** `grep -c 'f.report.format.opt.artifact\|f.report.format.option.artifact.desc' internal/web/assets/i18n.js`를 실행하면, **Then** 카운트는 8이다(4 로캘 × 2키).
- RED-now: 0(artifact 키 부재 — :188-194 등 4블록 실측). → M1에서 전환.
- 보조: `go test ./internal/web/ -run '^(SchemaForm|I18n)$'` 그린 — 라디오 폼 유지(<select> 회귀 없음)와 번역 완비 확인.

### AC-RAD-003 — 위저드 옵션 (REQ-001)
**Given** 초기화 위저드의 `report_format` 질문이 reportFormatValues를 미러링하고, **When** `grep -c 'Value:.*"artifact"' internal/cli/wizard/questions.go`를 실행하면, **Then** 카운트는 1 이상이고 `go test ./internal/cli/...`가 그린이다.
- RED-now: 옵션 2종(html+md·md, :112-113 실측). → M1에서 전환.

### AC-RAD-004 — 라우팅 문구 개정 (REQ-011)
**Given** skill-routing.md가 리포트 라우팅의 결합 문구와 안티패턴 문구를 운반하고, **When** 결합 라인 grep(`grep -n 'html+md.*md.*artifact' .claude/rules/moai/workflow/skill-routing.md`)과 분리 문구 grep(콘텐츠=html-report / 게시 계약=artifact-design 마커)을 실행하면, **Then** 결합 라인은 3값을 나열하고 안티패턴 단락은 소유권 분리 서술과 "format=artifact 게시 단계의 artifact-design 참조는 정상 라우팅" 예외를 갖는다.
- RED-now: 결합 라인 "values `html+md` \| `md`"(:38), 단독 라우팅 미스 문구(:42) 실측. → M4에서 전환.

### AC-RAD-005 — SKILL.md 전달 단계 3분기 (REQ-002, 003, 004)
**Given** SKILL.md의 "After rendering" 절이 전달 행위를 규정하고, **When** `grep -n 'artifact\|fallback\|republish\|재게시\|폴백' .claude/skills/moai-domain-html-report/SKILL.md`를 실행하면, **Then** (a) Artifact 게시 시 링크 제시·auto-open 생략 분기, (b) Artifact 도구 부재(Codex/GLM/API-key) 시 html+md 자동 폴백 + 사유 명시 분기, (c) 동일 파일 경로 재게시로 URL 유지 문구가 모두 존재한다.
- RED-now: 전달 단계가 auto-open 단일(섹션 실측). → M2에서 전환.

### AC-RAD-006 — 템플릿 아티팩트 블록 계약 형상 (REQ-005, 007, 008)
**Given** 6개 모드 mustache가 `artifact:begin/end` 구획을 갖고, **When** 각 파일에 대해 `grep -c 'artifact:begin'`·`grep -c 'artifact:end'`·`grep -c 'prefers-color-scheme: dark'`·`grep -c 'data-theme="dark"'`·`grep -c 'data-theme="light"'`를 실행하면, **Then** 6파일 전부에서 begin=1·end=1·다크모드 마커 3종 각 1 이상이고, 블록 안에 명시적 body 배경과 16px 거터 규격이 존재하며, 파일 선두는 자체 `<!doctype html>`로 유지된다.
- RED-now: 템플릿 6종에서 `prefers-color-scheme`·`data-theme` grep 0히트(2026-10-02 실측), artifact 구획 부재. → M3에서 전환.

### AC-RAD-007 — 타이틀 규칙 독트린 (REQ-006)
**Given** references/artifact-contract.md가 게시 계약 요약을 운반하고, **When** `grep -n 'title' .claude/skills/moai-domain-html-report/references/artifact-contract.md`를 실행하면, **Then** 2-4 단어 이름 규칙과 "이름: 설명" 형식 금지, 한 문장 description 동반 규칙이 존재한다(파일 자체 존재도 이 AC의 일부다).
- RED-now: 파일 부재. → M2에서 전환.

### AC-RAD-008 — 폰트 분리 (REQ-009)
**Given** 각 mustache에 artifact 구획과 html 경로가 공존하고, **When** `sed -n '/artifact:begin/,/artifact:end/p' <템플릿>`의 출력에서 `grep -c 'fonts.googleapis.com'`과 `grep -c 'cdn.jsdelivr.net'`를 실행하면, **Then** 6파일 전부에서 Google Fonts 참조가 존재하고 jsdelivr 참조는 0이다. 블록 밖에서는 `grep -c 'pretendard' <파일 전체>`가 html 경로의 Pretendard jsdelivr 라인을 그대로 보존한다(파일당 1).
- RED-now: artifact 구획 부재 + Pretendard jsdelivr 링크 존재(status 템플릿 헤드 실측). → M3에서 전환.

### AC-RAD-009 — mermaid 분리 (REQ-010)
**Given** artifact 구획과 Diagram Policy가 갱신되고, **When** 아티팩트 구획 내 mermaid script grep과 `grep -n 'pre.mermaid' .claude/skills/moai-domain-html-report/SKILL.md`를 실행하면, **Then** 아티팩트 구획 안에는 외부 mermaid `<script>`가 0이고 SKILL.md의 artifact 분기는 `pre.mermaid` 사전 렌더 방침을 명시한다. 블록 밖(html 경로)에는 등급 게이트 CDN + `<noscript>` 폴백 문구가 유지된다.
- RED-now: SKILL.md artifact 분기 부재(템플릿 자체는 CDN 스크립트 원래 무보유 — grep 0 실측, 유지 확인 대상). → M2/M3에서 전환.

### AC-RAD-010 — 출력 관례 불변 (REQ-004, D5)
**Given** 출력 경로 관례는 SPEC-REPORTS-LIFECYCLE-001이 소유하고, **When** `go test ./internal/template/ -run '^TestHtmlReportOutputPathParity$'`를 실행하면, **Then** 그린이다 — `.moai/reports/<slug>-<YYYYMMDD>.{html,md}` 문구와 create-before-write 문구가 양 미러에 유지된다.
- 그린 착지(회귀 방어) — 본 SPEC은 이 관례를 깨지 않는다.

### AC-RAD-011 — 미러 패리티 (REQ-012)
**Given** 라이브 스킬·routing과 템플릿 미러가 존재하고, **When** `diff -rq .claude/skills/moai-domain-html-report internal/template/templates/.claude/skills/moai-domain-html-report`와 `diff -q .claude/rules/moai/workflow/skill-routing.md internal/template/templates/.claude/rules/moai/workflow/skill-routing.md`를 실행하면, **Then** 양쪽 모두 출력 0(동일)이고 `go test ./internal/template/`가 그린이다.
- 그린 착지(2026-10-02 실측 PARITY-OK) — M5 이후에도 유지.

### AC-RAD-012 — 템플릿 중립성 (REQ-012, D2)
**Given** 템플릿 미러는 카드 내력을 운반하지 않는다, **When** `grep -rn 't1427\|SPEC-REPORT-ARTIFACT' internal/template/templates/`를 실행하면, **Then** 히트 0이다.
- RED/GREEN 최초 관측은 run 단계 pre-flight에서 수행(본 plan 단계 미측정 — M5에서 전환 조건 아님, 항시 유지 조건).

### AC-RAD-013 — 기존 값 보존 (REQ-001, REQ-012 보존 축)
**Given** 기존 사용자의 report.yaml은 `html+md` 또는 `md`를 가지고, **When** `cat .moai/config/sections/report.yaml`과 `diff -q .moai/config/sections/report.yaml internal/template/templates/.moai/config/sections/report.yaml`를 실행하면, **Then** 기본값 `format: html+md`가 불변이고 live·미러가 동일하다 — 기존 두 값의 동작은 변경되지 않는다(마이그레이션 불요).
- 그린 착지(26 bytes, REPORT-YAML-MIRROR-OK 실측).

### AC-RAD-014 — 지침 예산 (D3)
**Given** SKILL.md는 지침 예산 여유가 유한하고, **When** `wc -c .claude/skills/moai-domain-html-report/SKILL.md`를 실행하면, **Then** 결과는 26,000 이하이다(현재 22,516 — 증분 상한 ~3.5KB, 대형 추가는 references/로).
- RED-now: 해당 없음(상한 준수 조건 — M2/M3 편집 후 관측).

## §D AC 행렬 요약

| AC | REQ | 전환 | 유형 | RED-now(2e700a15e) |
|----|-----|------|------|--------------------|
| AC-RAD-001 | REQ-001 | M1 | release-blocking | RED (2값) |
| AC-RAD-002 | REQ-001 | M1 | release-blocking | RED (키 0) |
| AC-RAD-003 | REQ-001 | M1 | release-blocking | RED (2옵션) |
| AC-RAD-004 | REQ-011 | M4 | release-blocking | RED (구문 부재) |
| AC-RAD-005 | REQ-002/003/004 | M2 | release-blocking | RED (auto-open 단일) |
| AC-RAD-006 | REQ-005/007/008 | M3 | release-blocking | RED (블록·다크모드 0) |
| AC-RAD-007 | REQ-006 | M2 | release-blocking | RED (파일 부재) |
| AC-RAD-008 | REQ-009 | M3 | release-blocking | RED (구획 부재) |
| AC-RAD-009 | REQ-010 | M2/M3 | release-blocking | RED (분기 부재) |
| AC-RAD-010 | REQ-004/D5 | M5 | regression-guard | GREEN 착지 |
| AC-RAD-011 | REQ-012 | M5 | regression-guard | GREEN 착지 |
| AC-RAD-012 | REQ-012/D2 | M5 | regression-guard | 항시 유지 |
| AC-RAD-013 | REQ-001 보존 | M1 | regression-guard | GREEN 착지 |
| AC-RAD-014 | D3 | M2/M3 | release-blocking | 관측 대상(상한) |

## §D.1 심각도 분류

- **release-blocking(11)** — AC-RAD-001~009, 014: 본 SPEC의 목적 자체. FAIL 시 run 단계 종결 불가.
- **regression-guard(3)** — AC-RAD-010~013: 기존 동작 보존 확인. 위반 시 즉시 수리 대상.

## §D.2 추적성 (REQ ↔ AC)

- REQ-001 → AC-001, 002, 003, 013 / REQ-002 → AC-005 / REQ-003 → AC-005 / REQ-004 → AC-005, 010 / REQ-005 → AC-006 / REQ-006 → AC-007 / REQ-007 → AC-006 / REQ-008 → AC-006 / REQ-009 → AC-008 / REQ-010 → AC-009 / REQ-011 → AC-004 / REQ-012 → AC-011, 012
- 모든 REQ가 최소 1개 AC로 덮인다(12/12), 모든 AC가 REQ에 귀속된다(14/14).

## §D.3 간접 검증 (라이브 게시 불가 항목)

- Artifact 도구 게시·링크 제시·재게시 URL 유지의 런타임 동작은 CI 밖 행위다 — AC-005는 독트린 존재로, 폴백 트리거는 문구 조건(Codex/GLM/API-key) 명시로 대체 검증한다. 라이브 검증은 운영 세션의 최초 사용 사례로 사후 확인한다(본 SPEC의 종결 조건 아님).

## §D.4 종결 게이트

- run 단계 종결: AC-RAD-001~014 전부 PASS + E1~E8 자가검증 산출 + `go build ./...` 그린.
- sync 단계 종결: CHANGELOG 항목 + 미러 패리티 재입증(diff 0) + 템플릿 중립성 재입증.

## §D.5 Forward-looking checks

- artifact 값 사용률 관측(운영 통계가 존재하게 되면) — 폴백 발생률이 높으면 폴백 사유 표면 개선 후속 SPEC 후보.
- 아티팩트 뷰어 계약 변화 감시 — CSP 호스트 목록·16MB 상한·mermaid 사전 렌더 사양이 바뀌면 references/artifact-contract.md를 갱신하는 후속 카드.

## §E 품질 게이트 (TRUST 5)

- **Tested** — M1 신규 테스트(폐쇄 집합 3값 단언) + 기존 settings/web/cli/template 스위트 그린.
- **Readable** — SKILL.md·mustache 주석의 명확성, artifact 구획의 자기 서술성.
- **Unified** — 6 템플릿의 블록 구조·마커 통일, i18n 4 로캘 대칭.
- **Secured** — 외부 리소스 호스트 화이트리스트 준수(Google Fonts 전용), 스크립트 주입 없음.
- **Trackable** — Conventional Commits + `(card t1427)` + `🗿 MoAI` 트레일러, AC 행렬 근거 링크.

## §F Definition of Done

1. AC-RAD-001~014 전부 PASS(§D 표의 명령으로 재실행 가능).
2. 라이브·템플릿 미러 byte 동일(diff 0) + 템플릿 중립성 0히트.
3. SKILL.md ≤ 26,000B, references/artifact-contract.md 존재.
4. `go test ./internal/settings/ ./internal/web/ ./internal/cli/... ./internal/template/` 그린.
5. 커밋 규율 준수(Conventional Commits · (card t1427) · 🗿 MoAI · pathspec 스테이징) — push·PR은 레인 규율상 리더 소관.
