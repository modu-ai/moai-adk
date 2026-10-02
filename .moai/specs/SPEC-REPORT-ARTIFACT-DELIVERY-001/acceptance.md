# SPEC-REPORT-ARTIFACT-DELIVERY-001 — 인수 기준

> 이 문서는 status 축에서 stateless다 — `status:` 필드를 갖지 않는다.

## §A 검증 원칙

- 모든 AC는 기계 검증 가능 — 명령과 관측 결과를 함께 명명한다.
- **§2.1 4요소(단일 호출 명령 · 축어 stdout · exit 코드 · 트리 핀)는 plan 단계에서 완성됐다** — §E 증거 장부(EV-01~EV-17)가 그 운반체다(fenced ledger — 표 셀의 셸 메타문자 훼손 회피, verification-completeness §2.1 권장 carrier). run 단계는 같은 명령의 GREEN 측 재실행 기록을 같은 양식으로 남긴다.
- RED 관측 트리 핀: **5457f5832**(plan 커밋 — `.moai/specs/`만 추가하므로 코드 표면은 2e700a15e와 동일, `git diff --stat 2e700a15e 5457f5832` = 5 files/404 insertions로 등가성 입증 — plan-audit.md 실측과 부합). 관측 시각 2026-10-02, 본 워크트리.
- 관측 방법 공개: 각 계측은 `<명령>; echo "EXIT:$?"` 래퍼로 실행했다 — 장부의 명령 필드는 내부 명령 축어이고 EXIT 필드는 래퍼가 관측한 값이다.
- `format=artifact`의 라이브 게시는 CI·테스트 환경에서 실행 불가(Artifact 도구 부재)이므로, AC는 **메커니즘**(계약 형상·폴백 트리거·미러 패리티·독트린 존재)을 검증한다 — 라이브 게시는 검증하지 않는다.

## §B 시나리오 (Given-When-Then)

### AC-RAD-001 — settings 폐쇄 집합 (REQ-001) — release-blocking
**Given** settings 스키마가 report.format의 폐쇄 집합을 소유하고, **When** EV-01 명령을 실행하면, **Then** 현재 `0`(exit 1) — RED. GREEN 전환(M1): 같은 명령이 ≥1을 내고 `go test ./internal/settings/`가 그린이다(폐쇄 집합 3값 단언 테스트 포함).

### AC-RAD-002 — i18n 4 로캘 키 (REQ-001) — release-blocking
**Given** i18n 카탈로그가 en/ko/ja/zh 4 블록으로 구성되고, **When** EV-02 명령을 실행하면, **Then** 현재 `0`(exit 1) — RED. GREEN 전환(M1): 결합 패턴 `grep -c 'f.report.format.opt.artifact\|f.report.format.option.artifact.desc'`이 **8**(4 로캘 × 2키)을 내고 `go test ./internal/web/ -run '^(SchemaForm|I18n)$'`가 그린이다(라디오 폼 유지 — `<select>` 회귀 없음 — 및 번역 완비). 키명 `opt.artifact`/`option.artifact.desc`는 기존 카탈로그 관례(`opt.html+md`/`option.html_md.desc`)와 부합함이 감사에서 확인됐다.

### AC-RAD-003 — 위저드 옵션 (REQ-001) — release-blocking
**Given** 초기화 위저드의 `report_format` 질문이 reportFormatValues를 미러링하고, **When** EV-03 명령을 실행하면, **Then** 현재 `0`(exit 1 — 옵션 2종) — RED. GREEN 전환(M1): 같은 명령이 ≥1이고 `go test ./internal/cli/...`가 그린이다.

### AC-RAD-004 — 라우팅 문구 개정 (REQ-011) — release-blocking
**Given** skill-routing.md가 결합 문구와 안티패턴 단락을 운반하고, **When** EV-04(분리 서술 마커 `publication contract`)와 EV-05(3값 결합 라인)를 실행하면, **Then** 둘 다 현재 `0`(exit 1) — RED. GREEN 전환(M4): EV-04·05가 각 ≥1이고 안티패턴 단락은 "콘텐츠·렌더링 = moai-domain-html-report / 아티팩트 게시 계약 = artifact-design" 분리 서술과 format=artifact 게시 단계의 artifact-design 참조 예외를 갖는다.

### AC-RAD-005 — SKILL.md 전달 단계 3분기 (REQ-002, 003, 004) — release-blocking
**Given** SKILL.md의 "After rendering" 절이 전달 행위를 규정하고, **When** 변별 축 세 가지 — EV-06(`republish`), EV-07(`artifact link`), EV-08(`Artifact tool`) — 을 실행하면, **Then** 셋 모두 현재 `0`(exit 1) — RED.
- **비변별 기저 기록**(EV-17): 일반 서술 축은 baseline에서 이미 논그린이다 — `artifact`×7, `fallback`×8(본 세션 실측, exit 0). 본 AC는 이 축들을 앵커로 쓰지 않는다(1차 감사 D2 시정).
- GREEN 전환(M2): EV-06·07·08 각 ≥1 — (a) artifact 게시 시 링크 제시·auto-open 생략 분기, (b) Artifact 도구 부재(Codex/GLM/API-key) 시 html+md 자동 폴백 + 사유 명시 분기, (c) 동일 파일 경로 재게시로 URL 유지 문구.

### AC-RAD-006 — 템플릿 아티팩트 블록 계약 형상 (REQ-005, 007, 008) — release-blocking
**Given** 6개 모드 mustache가 존재하고, **When** EV-10(다크모드 마커)과 EV-11(`artifact:begin`)을 실행하면, **Then** 두 명령 모두 6파일 전부 `:0`(exit 1) — RED. GREEN 전환(M3): 파일당 `artifact:begin`/`artifact:end` 각 1, `prefers-color-scheme: dark`·`:root:not([data-theme="light"])`·`data-theme="dark"` 각 ≥1, 블록 안 명시적 body 배경(양 스킴)과 16px 측면 거터 규격 존재, 파일 선두는 자체 `<!doctype html>` 유지(REQ-005 — Jev Q2 확정 해석).

### AC-RAD-007 — 타이틀 규칙 독트린 (REQ-006) — release-blocking
**Given** references/artifact-contract.md가 게시 계약 요약을 운반하고, **When** EV-14(`wc -c`)를 실행하면, **Then** 현재 파일 부재(stdout 없음, stderr `open: No such file or directory`, exit 1) — RED. GREEN 전환(M2): 파일 존재 + 2-4 단어 이름 규칙·"이름: 설명" 형식 금지·한 문장 description 동반 마커 존재.

### AC-RAD-008 — 폰트 분리 (REQ-009) — release-blocking
**Given** 각 mustache에 artifact 구획과 html 경로가 공존하고, **When** EV-11(블록 부재)을 RED로, EV-12(pretendard 출현 수)를 보존 베이스라인으로 삼아 실행하면, **Then** 현재 블록 부재(6파일 `:0`) — RED이며 베이스라인은 **explainer 0, plan 3, financial/incident/pr/status 각 4**(exit 0)로 핀된다.
- GREEN 전환(M3): 블록 범위 스캔(`sed -n '/artifact:begin/,/artifact:end/p' <파일>`)에서 `fonts.googleapis.com` ≥1 · `cdn.jsdelivr.net` 0(6파일 전부) — 그리고 **블록 밖 pretendard 출현 수가 EV-12 핀치와 파일별로 동일**(explainer 0, plan 3, 나머지 4 — "파일당 1" 문언은 1차 감사 D3에서 실측 모순으로 철회됨).

### AC-RAD-009 — mermaid 분리 (REQ-010) — release-blocking
**Given** SKILL.md의 Diagram Policy가 갱신 대상이고, **When** EV-13(`pre.mermaid`)을 실행하면, **Then** 현재 `0`(exit 1) — RED. GREEN 전환(M2/M3): EV-13 ≥1(artifact 분기 — `pre.mermaid` 사전 렌더 방침) + html 경로의 등급 게이트 CDN·`<noscript>` 폴백 문구 유지. 템플릿의 외부 mermaid 스크립트 무보유는 그린 착지 회귀 가드다(explainer의 기존 인라인 `<script>`는 CDN이 아님 — 블록 흡수 금지, plan §F M3 주의 문항).

### AC-RAD-010 — 출력 관례 불변 (REQ-004, D5) — regression-guard
**Given** 출력 경로 관례는 SPEC-REPORTS-LIFECYCLE-001이 소유하고, **When** `go test ./internal/template/ -run '^TestHtmlReportOutputPathParity$'`를 실행하면, **Then** 그린이다 — `.moai/reports/<slug>-<YYYYMMDD>.{html,md}` 문구와 create-before-write 문구가 양 미러에 유지된다.

### AC-RAD-011 — 미러 패리티 (REQ-012) — regression-guard
**Given** 라이브 스킬·routing과 템플릿 미러가 존재하고, **When** `diff -rq .claude/skills/moai-domain-html-report internal/template/templates/.claude/skills/moai-domain-html-report`와 `diff -q .claude/rules/moai/workflow/skill-routing.md internal/template/templates/.claude/rules/moai/workflow/skill-routing.md`를 실행하면, **Then** 양쪽 모두 출력 0·exit 0이고 `go test ./internal/template/`가 그린이다. 착수 시점 실측 exit 0(PARITY-OK, 2026-10-02) — M5 이후에도 유지.

### AC-RAD-012 — 템플릿 중립성 (REQ-012, D2) — regression-guard
**Given** 템플릿 미러는 카드 내력을 운반하지 않는다, **When** EV-16을 실행하면, **Then** 현재 stdout 없음·exit 1(0히트 — 본 세션 핀 관측)이며 항시 유지된다.

### AC-RAD-013 — 기존 값 보존 (REQ-001 보존 축) — regression-guard
**Given** 기존 사용자의 report.yaml은 `html+md` 또는 `md`를 가지고, **When** `cat .moai/config/sections/report.yaml`과 `diff -q .moai/config/sections/report.yaml internal/template/templates/.moai/config/sections/report.yaml`를 실행하면, **Then** 기본값 `format: html+md` 불변·live 미러 동일(26 bytes — 착수 실측)이며 기존 두 값의 동작은 변경되지 않는다(마이그레이션 불요).

### AC-RAD-014 — 지침 예산 (D3) — regression-guard (1차 감사 D1 required fix로 재분류)
**Given** SKILL.md는 지침 예산 여유가 유한하고, **When** EV-15(`wc -c`)를 실행하면, **Then** 현재 `22516`(exit 0) ≤ 26,000 — 가드 베이스라인. 오늘 이미 그린 상한 준수 조건에는 RED가 존재하지 않으므로 release-blocking에서 재분류했다(전환 없음 — 항시 유지, M5 재입증). 대형 독트린의 references/ 배치(신설 파일)의 RED는 AC-RAD-007이 운반한다.

## §D AC 행렬 요약

| AC | REQ | 전환 | 유형 | RED-now (트리 5457f5832) | 근거 |
|----|-----|------|------|--------------------------|------|
| AC-RAD-001 | REQ-001 | M1 | release-blocking | RED — `0`, exit 1 | EV-01 |
| AC-RAD-002 | REQ-001 | M1 | release-blocking | RED — `0`, exit 1 | EV-02 |
| AC-RAD-003 | REQ-001 | M1 | release-blocking | RED — `0`, exit 1 | EV-03 |
| AC-RAD-004 | REQ-011 | M4 | release-blocking | RED — `0`×2, exit 1×2 | EV-04, EV-05 |
| AC-RAD-005 | REQ-002/003/004 | M2 | release-blocking | RED — `0`×3, exit 1×3 | EV-06, EV-07, EV-08 (+EV-09, EV-17) |
| AC-RAD-006 | REQ-005/007/008 | M3 | release-blocking | RED — 6파일 `:0`×2, exit 1×2 | EV-10, EV-11 |
| AC-RAD-007 | REQ-006 | M2 | release-blocking | RED — 파일 부재, exit 1 | EV-14 |
| AC-RAD-008 | REQ-009 | M3 | release-blocking | RED — 블록 6파일 `:0` + 베이스라인 핀 | EV-11, EV-12 |
| AC-RAD-009 | REQ-010 | M2/M3 | release-blocking | RED — `0`, exit 1 | EV-13 |
| AC-RAD-010 | REQ-004/D5 | M5 재검증 | regression-guard | GREEN 착지 | — |
| AC-RAD-011 | REQ-012 | M5 재검증 | regression-guard | GREEN 착지 (exit 0 실측) | — |
| AC-RAD-012 | REQ-012/D2 | 항시 유지 | regression-guard | 핀 관측 — 0행, exit 1 | EV-16 |
| AC-RAD-013 | REQ-001 보존 | 항시 유지 | regression-guard | GREEN 착지 (26B) | — |
| AC-RAD-014 | D3 파생 | 항시 유지 | regression-guard | 가드 베이스라인 — 22516, exit 0 | EV-15 |

## §D.1 심각도 분류

- **release-blocking(9)** — AC-RAD-001~009: 본 SPEC의 목적 자체. FAIL 시 run 단계 종결 불가. 전원이 §E 장부의 4요소 RED 기록을 운반한다.
- **regression-guard(5)** — AC-RAD-010~014: 기존 동작 보존·상한 확인. AC-RAD-014는 1차 감사 D1 required fix에 따라 release-blocking에서 재분류됐다(오늘 이미 그린 상한 조건에는 RED가 존재하지 않음). 위반 시 즉시 수리 대상.

## §D.2 추적성 (REQ ↔ AC)

- REQ-001 → AC-001, 002, 003, 013 / REQ-002 → AC-005 / REQ-003 → AC-005 / REQ-004 → AC-005, 010 / REQ-005 → AC-006 / REQ-006 → AC-007 / REQ-007 → AC-006 / REQ-008 → AC-006 / REQ-009 → AC-008 / REQ-010 → AC-009 / REQ-011 → AC-004 / REQ-012 → AC-011, 012
- 모든 REQ가 최소 1개 AC로 덮인다(**12/12**). AC 귀속은 **13/14**가 REQ에 직결되고 AC-RAD-014는 제약 D3(지침 예산) 파생이다(1차 감사 D4 정정 — "14/14" 주장 철회).

## §D.3 간접 검증 (라이브 게시 불가 항목)

- Artifact 도구 게시·링크 제시·재게시 URL 유지의 런타임 동작은 CI 밖 행위다 — AC-005는 변별 마커의 존재로, 폴백 트리거는 조건 명시(Codex/GLM/API-key)로 대체 검증한다. 라이브 검증은 운영 세션의 최초 사용 사례로 사후 확인한다(본 SPEC의 종결 조건 아님).

## §D.4 종결 게이트

- run 단계 종결: AC-RAD-001~014 전부 PASS + E1~E8 자가검증 산출 + `go build ./...` 그린.
- sync 단계 종결: CHANGELOG 항목 + 미러 패리티 재입증(diff 0) + 템플릿 중립성 재입증.

## §D.5 Forward-looking checks

- artifact 값 사용률 관측(운영 통계가 존재하게 되면) — 폴백 발생률이 높으면 폴백 사유 표면 개선 후속 SPEC 후보.
- 아티팩트 뷰어 계약 변화 감시 — CSP 호스트 목록·16MB 상한·mermaid 사전 렌더 사양이 바뀌면 references/artifact-contract.md를 갱신하는 후속 카드.

## §E 증거 장부 (evidence ledger — §2.1 4요소 운반체)

관측 조건: 2026-10-02, 트리 핀 **5457f5832**(plan 커밋 — 코드 표면은 2e700a15e와 동일). 관측 방법: `<명령>; echo "EXIT:$?"` 래퍼 — 명령 필드는 내부 명령 축어, EXIT는 래퍼 관측값.

```text
EV-01  cmd:   grep -c '"artifact"' internal/settings/schema_sections.go
       stdout: 0
       exit:  1
EV-02  cmd:   grep -c 'f.report.format.opt.artifact' internal/web/assets/i18n.js
       stdout: 0
       exit:  1
EV-03  cmd:   grep -c 'Value: "artifact"' internal/cli/wizard/questions.go
       stdout: 0
       exit:  1
EV-04  cmd:   grep -c 'publication contract' .claude/rules/moai/workflow/skill-routing.md
       stdout: 0
       exit:  1
EV-05  cmd:   grep -cE '`html\+md`.{0,4}`md`.{0,4}`artifact`' .claude/rules/moai/workflow/skill-routing.md
       stdout: 0
       exit:  1
EV-06  cmd:   grep -ci 'republish' .claude/skills/moai-domain-html-report/SKILL.md
       stdout: 0
       exit:  1
EV-07  cmd:   grep -c 'artifact link' .claude/skills/moai-domain-html-report/SKILL.md
       stdout: 0
       exit:  1
EV-08  cmd:   grep -c 'Artifact tool' .claude/skills/moai-domain-html-report/SKILL.md
       stdout: 0
       exit:  1
EV-09  cmd:   grep -ciE 'republish|재게시' .claude/skills/moai-domain-html-report/SKILL.md
       stdout: 0
       exit:  1
EV-10  cmd:   grep -c 'prefers-color-scheme: dark' .claude/skills/moai-domain-html-report/references/templates/*.mustache
       stdout: financial:0 / explainer:0 / plan:0 / incident:0 / status:0 / pr:0 (6행 전부 :0)
       exit:  1
EV-11  cmd:   grep -c 'artifact:begin' .claude/skills/moai-domain-html-report/references/templates/*.mustache
       stdout: financial:0 / explainer:0 / plan:0 / incident:0 / status:0 / pr:0 (6행 전부 :0)
       exit:  1
EV-12  cmd:   grep -ci 'pretendard' .claude/skills/moai-domain-html-report/references/templates/*.mustache
       stdout: explainer:0 / plan:3 / financial:4 / incident:4 / pr:4 / status:4
       exit:  0   (보존 베이스라인 — AC-008 GREEN 셀의 파일별 핀치)
EV-13  cmd:   grep -c 'pre.mermaid' .claude/skills/moai-domain-html-report/SKILL.md
       stdout: 0
       exit:  1
EV-14  cmd:   wc -c .claude/skills/moai-domain-html-report/references/artifact-contract.md
       stdout: (없음)  stderr: wc: ...: open: No such file or directory
       exit:  1
EV-15  cmd:   wc -c .claude/skills/moai-domain-html-report/SKILL.md
       stdout: 22516 .claude/skills/moai-domain-html-report/SKILL.md
       exit:  0   (AC-014 가드 베이스라인 — ≤ 26000)
EV-16  cmd:   grep -rn 't1427\|SPEC-REPORT-ARTIFACT' internal/template/templates/
       stdout: (없음)
       exit:  1   (AC-012 핀 관측 — 0히트)
EV-17  cmd:   grep -c 'artifact' .claude/skills/moai-domain-html-report/SKILL.md
       stdout: 7   exit: 0
EV-17b cmd:   grep -c 'fallback' .claude/skills/moai-domain-html-report/SKILL.md
       stdout: 8   exit: 0   (AC-005 비변별 기저 기록 — 이 축은 앵커로 쓰지 않음)
```

보강 계측(D6 재측정 — ls long-format 왜곡 교정):

```text
EV-18  cmd:   find internal/template/templates/.agents/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-*' | wc -l
       stdout: 17   exit: 0
EV-18b cmd:   find .agents/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-*' | wc -l
       stdout: 17   exit: 0
EV-18c cmd:   find internal/template/templates/.agents/skills -maxdepth 1 -mindepth 1 -type d -name 'moai-domain-*' | wc -l
       stdout: 0    exit: 0   (도메인 스킬 미러 부재 — REQ-012 근거)
```

## §E.1 품질 게이트 (TRUST 5)

- **Tested** — M1 신규 테스트(폐쇄 집합 3값 단언) + 기존 settings/web/cli/template 스위트 그린.
- **Readable** — SKILL.md·mustache 주석의 명확성, artifact 구획의 자기 서술성.
- **Unified** — 6 템플릿의 블록 구조·마커 통일, i18n 4 로캘 대칭.
- **Secured** — 외부 리소스 호스트 화이트리스트 준수(Google Fonts 전용), 스크립트 주입 없음.
- **Trackable** — Conventional Commits + `(card t1427)` + `🗿 MoAI` 트레일러, AC 행렬 근거 링크.

## §F Definition of Done

1. AC-RAD-001~014 전부 PASS(§D 표·§E 장부의 명령으로 재실행 가능 — RED 셀은 GREEN 측 재실행 기록으로 대체).
2. 라이브·템플릿 미러 byte 동일(diff 0) + 템플릿 중립성 0히트 유지.
3. SKILL.md ≤ 26,000B, references/artifact-contract.md 존재.
4. `go test ./internal/settings/ ./internal/web/ ./internal/cli/... ./internal/template/` 그린.
5. 커밋 규율 준수(Conventional Commits · (card t1427) · 🗿 MoAI · pathspec 스테이징) — push·PR은 레인 규율상 리더 소관.
