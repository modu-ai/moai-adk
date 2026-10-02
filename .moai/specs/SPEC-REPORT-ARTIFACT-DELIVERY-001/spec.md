---
id: SPEC-REPORT-ARTIFACT-DELIVERY-001
title: "html-report 스킬에 Claude Artifact 전달 형식 추가 — report.format=artifact·아티팩트 페이지 계약 템플릿 정합·전달 단계 개편 (card t1427)"
version: "0.1.0"
status: in-progress
created: 2026-10-02
updated: 2026-10-02
author: "MoAI lane (card t1427)"
priority: P1
phase: "v3.2.0"
module: ".claude/skills/moai-domain-html-report"
lifecycle: spec-anchored
tags: "html-report, artifact, report-format, delivery, font-policy, template-mirror"
tier: M
related_specs: [SPEC-REPORTS-LIFECYCLE-001]
---

# SPEC-REPORT-ARTIFACT-DELIVERY-001 — html-report 스킬에 Claude Artifact 전달 형식 추가

## HISTORY

| 날짜 | 변경 | 근거 |
|------|------|------|
| 2026-10-02 | 최초 작성 (Tier M, plan-phase) | 리더 배차 — 카드 t1427 (운영자 지시 2026-10-02, Class C) |
| 2026-10-02 | plan-audit 1차(FAIL 0.875 · MP-8) 수리 — D1 RED-now 4요소 증거 장부(EV-01~18) plan 단계 완성·AC-005/008 재고정·AC-014 regression-guard 재분류·D4 추적성 정정·D5 마일스톤 정렬·D6 .agents 17종 정정·D7 Jev 해결 기재 | `.moai/reports/t1427/plan-audit.md` (런타임 기록 — 미커밋, 경로 참조 전용) |

---

## §A 문제 정의 및 배경

### §A.1 배경

moai-domain-html-report 스킬은 마크다운 보고서를 단일 HTML 파일과 마크다운 쌍둥이로 렌더링해 `.moai/reports/`에 기록하고 브라우저로 자동 열어 전달한다. Claude 세션에는 같은 HTML을 claude.ai 아티팩트로 게시해 링크로 전달하는 Artifact 도구가 있지만, 스킬에는 그 경로가 없다. 아티팩트 뷰어는 자체 페이지 계약(단독 실행 문서·스타일시트 호스트 제한·mermaid 사전 렌더 등)을 강제하므로, 기존 템플릿을 그대로 게시하면 계약 위반이 된다.

카드 t1427(운영자 지시 2026-10-02)은 `report.format` 설정에 `artifact` 값을 추가해 아티팩트 게시를 제1 전달 경로로 만들고, 6개 모드 템플릿을 아티팩트 페이지 계약에 맞추며, 폰트·mermaid·전달 단계·라우팅 문구를 개정한다. **스킬 대체가 아니다** — 콘텐츠·모드·청중 등급·md 쌍둥이의 소유권은 html-report에 유지되고, 게시 형식만 Artifact 페이지 계약(artifact-design)을 따른다.

### §A.2 측정 확립 사실 (코드 실측, 2026-10-02, 트리 2e700a15e)

1. **report.format 폐쇄 집합**: `internal/settings/schema_sections.go:531` — `var reportFormatValues = []string{"html+md", "md"}`. `artifact` 값 부재. 라디오 FieldDef(`reportFields()`, :542-559)가 각 옵션에 i18n desc 키를 부여한다.
2. **설정 파일**: `.moai/config/sections/report.yaml` — `format: html+md` (라이브·템플릿 미러 byte 동일, 26 bytes). report 섹션은 승인된 설정 고아(acknowledged orphan)로 로더 변경이 불요하다.
3. **i18n**: `internal/web/assets/i18n.js` — en(:188-194)·ko(:1117-1123)·ja(:1918-1924)·zh(:2719-2725) 4 로캘 블록이 각각 `title`/`desc`/`opt.<값>`/`option.<값>.desc` 키를 보유한다.
4. **위저드**: `internal/cli/wizard/questions.go:102-113` — `report_format` 질문이 `reportFormatValues`를 미러링한다("Keep these two Values in sync with that SSOT" 주석 명시).
5. **라우팅 결합**: `.claude/rules/moai/workflow/skill-routing.md:38` — "values `html+md` \| `md`" 결합 문구. :42 — 리포트 렌더 요청의 artifact-design 로딩을 단일 라우팅 미스로 규정하는 안티패턴 문구.
6. **템플릿**: `references/templates/*.html.mustache` 6종(pr 18,473B · status 17,927B · incident 17,042B · plan 15,514B · financial 14,161B · explainer 14,037B) — 자체 `<!doctype html>`·charset·viewport·`:root` 토큰·body 배경을 이미 보유. 다크모드 블록 0건(`prefers-color-scheme`/`data-theme` grep 0히트 실측), Pretendard jsdelivr 링크 보유.
7. **폰트**: `references/fonts.md` — status/financial/pr/incident/plan = Pretendard(jsdelivr, pinned v1.3.9) + JetBrains Mono(Google), explainer = Noto Sans KR·Noto Serif KR·JetBrains Mono(전량 Google). **Pretendard는 Google Fonts 카탈로그에 없다.**
8. **미러 패리티**: 라이브 스킬 디렉터리와 skill-routing.md 모두 `internal/template/templates/` 미러와 byte 동일(`diff -rq` 실측 2026-10-02 — SKILL-MIRROR-PARITY-OK·ROUTING-MIRROR-PARITY-OK).
9. **Codex 배포 경로**: `internal/template/templates/.agents/skills/`와 라이브 `.agents/skills/` 모두 moai-* 커맨드 스킬 17종만 보유(`find -type d -name 'moai-*'` 재계측 — 1차 감사 D6 정정; 초판 "19종"은 ls 포맷 오독) — moai-domain-* 도메인 스킬의 `.agents` 미러는 존재하지 않는다(카드 ⑦의 동기 범위를 Claude 표면 + 템플릿 SSOT로 한정하는 근거).
10. **전달 단계**: SKILL.md "## After rendering — report back to the user" — 요약 출력 + 브라우저 auto-open(platform opener)이 현재 유일한 전달 행위.
11. **지침 예산**: `internal/hook/instructions_loaded.go:103` `charBudget = 40000`(rules 측정 기준). SKILL.md 현재 22,516B · 406행 — 증분 여유는 유한하다.

### §A.3 증거 기반 노트 (named gap)

리더의 결정 보고서 아티팩트(`https://claude.ai/artifact/YGGYoic1FDSpp3bjv4FUVN`)는 본 세션에서 읽을 수 없었다 — Artifact 도구 거부(ANTHROPIC_AUTH_TOKEN 세션, claude.ai 로그인 없음), 웹 리더는 claude.ai 로그인 장벽만 반환했다. 본 SPEC은 카드 본문(리더가 그 보고서에서 옮겨 적은 운영 지시)을 검증된 근거로 삼으며, 보고서 원문의 내용을 어디에도 재구성하지 않는다.

## §B 범위 — 카드 항목과 요구사항 매핑

| 카드 항목 | 요구사항 |
|---|---|
| ① `report.format`에 `artifact` 값 추가 + Artifact 도구 부재 시 html+md 자동 폴백 | REQ-001, REQ-002 |
| ② 6개 모드 템플릿의 아티팩트 계약 정합 | REQ-005, REQ-006, REQ-007, REQ-008 |
| ③ 폰트(Pretendard 차단 → Google Fonts 대체) | REQ-009 |
| ④ mermaid CDN 스크립트 제거(pre.mermaid 사전 렌더) | REQ-010 |
| ⑤ auto-open 대신 아티팩트 링크 제시 + 동일 경로 재게시로 URL 유지 | REQ-003, REQ-004 |
| ⑥ skill-routing.md 안티패턴 문구 개정 | REQ-011 |
| ⑦ 템플릿 미러 동기 + Codex 측 | REQ-012 |

## §C 요구사항 (GEARS)

- REQ-001 — 설정 폐쇄 집합 확장: **Where** the report format resolves through the settings chain (`.moai/config/sections/report.yaml`, persisted via `internal/settings`), the `report.format` closed set shall carry exactly three values — `html+md` (default), `md`, `artifact` — 설정 UI 라디오·초기화 위저드 `report_format` 질문·i18n 카탈로그(en/ko/ja/zh)는 세 값을 모두 제시하며 기본값 `html+md`는 불변이다.
- REQ-002 — Artifact 도구 부재 폴백: **When** `report.format`이 `artifact`인데 세션에 Artifact 도구가 없으면(Codex·GLM·API-key 세션), moai-domain-html-report shall deliver the report through the existing html+md dual-file path without treating it as an error — 파일 2종을 기록하고 브라우저 auto-open으로 전달하며, 보고 요약에 폴백 사실과 사유를 밝힌다.
- REQ-003 — 아티팩트 게시: **When** `report.format`이 `artifact`이고 Artifact 도구가 세션에 있으면, the skill shall publish the rendered HTML as a Claude Artifact and present the artifact link in the report-back summary — 브라우저 오프너를 실행하지 않는다.
- REQ-004 — 재게시·URL 유지: **When** 아티팩트로 전달된 보고서가 같은 출력 경로에서 다시 생성되면, the skill shall republish the same file path's content so the artifact URL is preserved — 그리고 `.moai/reports/<slug>-<YYYYMMDD>.{html,md}` 쌍 파일 기록은 계속 유지된다.
- REQ-005 — 단독 실행 문서: **When** the delivery format is artifact, the rendered document shall be a self-contained single file that starts with its own `<!doctype html>` and carries charset·viewport metas and inline base styles — 바깥 문서 골격이나 래퍼에 의존하지 않는다.
- REQ-006 — 타이틀 규칙: **When** the delivery format is artifact, the document `<title>` shall be a two-to-four-word name of the report ("이름: 설명" 형식 금지) and a one-sentence `description` shall accompany publication.
- REQ-007 — 다크모드 2블록: **When** the delivery format is artifact, the document shall define color tokens on `:root` and provide dark mode through BOTH blocks — `:root:not([data-theme="light"])` 가드가 붙은 `@media (prefers-color-scheme: dark)` 블록과 `:root[data-theme="dark"]` 블록 — 그리고 두 스킴 모두에서 `body`에 명시적 배경을 둔다.
- REQ-008 — 모바일 폭: **When** the delivery format is artifact, the layout shall render at phone width with a 16px side gutter and no horizontal page scroll.
- REQ-009 — 폰트: **While** the delivery format is artifact, the document shall load Korean fonts exclusively from Google Fonts (본문 Noto Sans KR, 세리프 모드 Noto Serif KR, 코드 JetBrains Mono) and shall not reference any non-Google-Fonts stylesheet host(jsdelivr Pretendard 포함). html 파일 형식의 Pretendard jsdelivr 적재는 변경되지 않는다.
- REQ-010 — mermaid: **When** the delivery format is artifact, the document shall contain no external mermaid `<script>` — 다이어그램은 뷰어의 기본 사전 렌더를 위한 `pre.mermaid` 코드 블록으로 내보낸다. html 파일 형식의 등급 게이트 mermaid CDN과 `<noscript>` 폴백은 변경되지 않는다.
- REQ-011 — 라우팅 문구 개정: The skill-routing doctrine shall state the ownership split — 콘텐츠·렌더링(6모드·청중 등급·md 쌍둥이)의 소유는 moai-domain-html-report에 있고, 아티팩트 게시 계약은 artifact-design을 참조한다 — 그리고 현행 "리포트 렌더 요청의 artifact-design 로딩 = 라우팅 미스" 단독 문구를 이 분리 서술로 대체한다. report.format 결합 라인은 세 값을 모두 나열한다.
- REQ-012 — 미러·Codex 동기: **Where** a live-surface file changes (`.claude/skills/moai-domain-html-report/**`, `.claude/rules/moai/workflow/skill-routing.md`), the same change set shall mirror it byte-identically into `internal/template/templates/` — 그리고 도메인 스킬의 Codex 측(`.agents/skills/`) 미러는 새로 만들지 않는다(배포 경로 부재 — §A.2 실측 9).

## §D 제약조건

- **D1 — 아티팩트 페이지 계약(구속력)**: 단독 실행 단일 파일(자체 `<!doctype html>`·charset·viewport·기본 스타일 인라인), 2-4 단어 `<title>` + 한 문장 description, `:root` 색 토큰, 다크모드 2블록(위 REQ-007 형태), 명시적 body 배경, 외부 스크립트는 cdnjs.cloudflare.com 또는 cdn.jsdelivr.net/npm/ 전용(아티팩트 형식에서는 스크립트 자체를 쓰지 않는 것이 기본), 스타일시트는 Google Fonts 전용, 16MB 상한에 `data:` URI 포함, 폰 폭 레이아웃(16px 거터·수평 스크롤 없음), mermaid는 뷰어 사전 렌더(`pre.mermaid`) — CDN 스크립트 금지.
- **D2 — 템플릿 중립성**: 템플릿 미러에 카드 내력·내부 날짜·기계 고유 값을 넣지 않는다(템플릿 중립성 CI 가드 엄격).
- **D3 — 지침 예산**: SKILL.md 증분은 최소화하고 대형 추가는 `references/`로 보낸다(§A.2 실측 11 — 22,516B 기준).
- **D4 — 소유권 분리**: 콘텐츠·모드·청중 등급·md 쌍둥이 = moai-domain-html-report 소유 불변. 게시 계약 준수만 artifact-design에서 가져온다.
- **D5 — 출력 관례**: `.moai/reports/<slug>-<YYYYMMDD>.{html,md}` 관례와 create-before-write 문구는 SPEC-REPORTS-LIFECYCLE-001 REQ-RLC-003/004가 소유 — 변경 금지(`internal/template/html_report_output_path_parity_test.go` 가드 존재).
- **D6 — 개발 방법론**: `quality.yaml` `constitution.development_mode: tdd` — M1의 Go 변경에는 RED-GREEN-REFACTOR를 적용한다(RED 출력 선(先) 확보).

## §E 가정 및 미해결 질의

- **A1 — "문서 골격 태그 제거"의 해석**: 카드 ②의 문구와 아티팩트 페이지 계약 원문("starting with its own `<!doctype html>` … no document-skeleton inheritance")은 표면상 상충한다. 본 SPEC은 계약 원문을 구속력 있는 해석으로 채택한다 — 아티팩트 산출물은 자체 골격을 갖춘 완전한 단독 문서이며, '제거'는 외부 골격 상속·래퍼 의존의 제거를 뜻한다. 카드 문구를 골격 태그의 물리적 제거로 읽으면 REQ-005 위반이 된다. → decision-index.md Q2에서 운영자 확정을 요청한다.
- **A2 — 재게시 URL 유지 의미론**: "같은 파일 경로 재게시로 URL 유지"는 카드가 명세한 스킬 독트린으로 규정한다. Artifact 도구의 라이브 게시 동작은 본 plan 단계에서 검증 불가(도구 부재 — §A.3)이므로, AC는 독트린의 존재와 폴백 트리거를 검증하고 라이브 게시는 검증하지 않는다.
- **A3 — 폰트 최종 결정**: 카드가 두 경로(Google Fonts 대체 / @font-face 인라인)를 병기하고 최종 결정을 저작 단계에 위임했다. plan.md §F M3에 결정과 근거를 기록했다(Noto 계열 Google Fonts — 기각 사유 포함). → decision-index.md Q1에서 운영자 확정을 요청한다.

## Out of Scope

아래 항목들은 본 SPEC이 구축하지 않는 것들이다.

### Out of Scope — 스킬 대체 및 구조 변경
- artifact는 추가 전달 형식이지 스킬 교체가 아니다 — 6모드(status/incident/plan/explainer/financial/pr) × 3청중 등급(expert/basic/learn) 구조와 md 쌍둥이 비대칭 원칙은 변경하지 않는다.
- 스킬의 html+md 기본 파이프라인의 재설계 또는 교체.

### Out of Scope — 아티팩트 디자인 저작 규칙 흡수
- artifact-design의 시각 아이덴티티·랜딩페이지 캘리브레이션 내용을 리포트 스킬로 가져오는 것 — 게시 계약 준수(§D1)만 취한다.

### Out of Scope — 신규 모드·등급
- 문서화만 된 `editorial`/`legal` 모드의 구현, 새 청중 등급의 추가.

### Out of Scope — 라이브 게시 자동화 검증
- CI·테스트 환경에서의 Artifact 도구 라이브 게시 실행 — AC는 메커니즘(계약 형상·폴백 트리거·미러 패리티)을 검증한다.

### Out of Scope — Codex 도메인 스킬 배포 경로 신설
- `.agents/skills/`에 moai-domain-html-report 미러를 새로 만드는 것 — §A.2 실측 9의 검증된 부재를 유지한다.

## §G 교차 참조

- SPEC-REPORTS-LIFECYCLE-001 — `.moai/reports/` 출력 관례·수명주기(REQ-RLC-003/004)
- `.claude/rules/moai/workflow/skill-routing.md` §1.1 — 오케스트레이터 직접 라우팅·report.format 결합
- `.claude/skills/moai-domain-html-report/references/fonts.md` — 폰트 매핑 SSOT
- `decision-index.md` Q1-Q3 — 미해결 결정 운반체
