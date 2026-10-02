# SPEC-REPORT-ARTIFACT-DELIVERY-001 — Decision Index

`interview.decision_gate: on`(`.moai/config/sections/interview.yaml:6`)에 따라 plan-phase 조립에서 드러난 미확정 결정의 운반체다. 각 행은 무엇이 왜 미해결인지만 서술하고, 어떤 행도 권고를 싣지 않는다. `Operator verdict:`는 저작 시점에 비어 있다. — **2026-10-02 갱신: Q1·Q2는 이후 Jev 판정(신뢰도 1.00×2, jev-1.13.0)으로 해결됐다 — 각 행의 verdict에 기재. 근거 파일 `.moai/reports/t1427/jev-decisions.md`는 런타임 기록으로 커밋되지 않는다(경로 참조 전용).**

### Q1: artifact 형식의 한국어 폰트를 어느 경로로 조달하는가 — Google Fonts CJK 패밀리(Noto 계열) 교체인가, @font-face data: URI 인라인인가?

Label: FOUNDER
Authority anchor: (none — 플랫폼 아티팩트 페이지 계약 원문은 커밋된 리포 아티팩트가 아니고, Pretendard의 Google Fonts 카탈로그 부재는 `references/fonts.md`의 매핑 실측일 뿐 정책이 아니다.)
Why unresolved: 카드 ③이 두 경로를 병기하고 최종 결정을 저작 단계에 위임했다. 본 SPEC은 Google Fonts Noto 계열(Noto Sans KR·Noto Serif KR + 기존 JetBrains Mono)로 결정했고 근거 3가지(호스트 제한·16MB 상한 vs 한국어 풀패밀리 용량·기존 패밀리 재사용으로 신규 CDN 관계 0)를 plan.md §F M3에 기록했다 — 운영자는 kickoff에서 확정하거나 다른 경로를 지시할 수 있다.
Operator verdict: RESOLVED — `google_fonts_noto`, confidence 1.00, jev-1.13.0 (2026-10-02). 근거: `.moai/reports/t1427/jev-decisions.md` (런타임 기록 — 미커밋, 경로 참조 전용). 본문 인코딩(REQ-009 · plan.md §F M3)과 일치.

### Q2: 카드 ②의 "문서 골격 태그 제거"는 무엇을 제거하라는 것인가 — 골격 상속·래퍼 의존의 제거인가, doctype 등 골격 태그 자체의 물리적 제거인가?

Label: FOUNDER
Authority anchor: (none — 카드 본문은 커밋된 정책 파일이 아니고, 아티팩트 페이지 계약 원문("self-contained single file starting with its own `<!doctype html>` … no document-skeleton inheritance")도 세션 외부 문구다.)
Why unresolved: 두 독해는 서로를 배타한다 — 물리적 태그 제거로 읽으면 플랫폼 계약의 "own `<!doctype html>`" 요구와 충돌한다. 본 SPEC은 계약 원문을 구속력 있는 해석으로 채택해 REQ-005를 "자체 골격을 갖춘 단독 문서, 외부 골격 의존 금지"로 규정했다(spec.md §E A1). 카드 문구의 원래 의도가 다르다면 REQ-005와 M3 템플릿 지시를 개정해야 한다.
Operator verdict: RESOLVED — `contract_binding_own_skeleton`, confidence 1.00, jev-1.13.0 (2026-10-02). 근거: `.moai/reports/t1427/jev-decisions.md` (런타임 기록 — 미커밋, 경로 참조 전용). REQ-005·spec.md §E A1 인코딩과 일치 — 카드 문언은 골격 상속·래퍼 가정의 제거로 확정.

### Q3: Tier M 분류가 적정한가?

Label: POLICY-COVERED
Authority anchor: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (분류표 — M = 300-1000 LOC 가이드, 5-15 파일, 3-file 아티팩트 세트, REQ/AC 16/16 상한).
Why unresolved: (none — 측정된 형상이 M 밴드에 들어온다: 영향 파일 12-15개(Go 3-4·웹 자산 1·mustache 6·마크다운 4-5·미러), 규모 300-1000 LOC, REQ 12/AC 14 사용. 카드가 분류 판단을 위임했다.)
Operator verdict:
