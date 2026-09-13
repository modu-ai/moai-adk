# SPEC-DOCS-TODO-TEMP-GUARD-001 — plan-audit 보고서 (카드 t575)

- Audit surface: plan-auditor (독립 감사 에이전트) — iteration 1 + D1-scoped confirming check
- Tree at final verdict: WT-docs-todo-locales @ e83b4ec20
- 이 파일은 감사 에이전트의 판정문을 영속화한 것 (원판정은 세션 메시지로만 존재해 리드가 트리에서 찾지 못함 — 리드 요청 2026-09-13)

## Iteration 1 (v1.0.0) — FAIL 0.94

- 판정: FAIL — MP-3(frontmatter) 단일 차단, 점수는 Tier S 기준 0.75 상회
- D1 [blocking]: spec.md:12 `lifecycle: spec-first` — 무효 열거값 (정론: spec-anchored|spec-lite|exploratory, SSOT spec-frontmatter-schema.md § Canonical 12 Required Fields). lint 엔진은 presence-only 검사라 못 잡음(감사가 잡음, corpus drift class 기록).
- D2 [optional]: acceptance.md §D.5 census 13-16행이 8개 클레임(×4 로케일 2벌)을 4 슬롯에 뭉뚱그림
- D3 [optional]: plan.md:6-7 "AC inline in spec.md §3" → §4 오타
- D4 [optional]: REQ-001 3클래스 나열이 5행 진리표를 다 반영하지 않는 듯 읽힘
- Must-pass: MP-1 PASS / MP-2 PASS / MP-3 FAIL / MP-4 N-A / MP-5 PASS / MP-6 PASS / MP-7 PASS
- 진리표 5행 전부 코드 대조로 검증됨 (todo_root.go:81-86·135-144·157-165, state_dir.go:28-40, temp_origin.go:60-107)
- Census 표본 검증: 17개 표본 전부 실재 확인 (moai-todo.md 59·248 ×4 / moai-web-console.md:108 TRUE / factory-mode.md:93·README 미검증 / 템플릿 todo-queue-storage.md 동종 결함)

## D1-scoped confirming check (v1.1.0, e83b4ec20) — PASS 1.00

- MP-3 → PASS: spec.md:12 `lifecycle: spec-anchored` 실측 확인, 12필드 모두 유효
- D2-D4 교란 없음: census 행 13-20·21 재번호 순차 무중복, §4 표기 2곳, REQ-001 5행 참조
- 최종 점수: 명확성 1.0 / 완전성 1.0 / 검증가능성 1.0 / 추적성 1.0 = **1.00**
- 판정: **run 진입 PASS** (skip-eligible 3조건 충족 — 아티팩트 해시 = e83b4ec20 시점)
- 잔여 미용 사항(채점 제외): census 행의 "2 per row-slot multiplier" 문구가 실제 산술(로케일 배수)과 어긋남 — 내용은 올바름

## Gaps (감사가 관측하지 않은 것)

- 문서 편집 자체·census 아티팩트(.moai/reports/t575/docs-census.md)·AC-001..007 실행 — run 단계 표면
- 임시 루트 하 factory 스토어 가드 존재 여부 — 의도적 범위 제외(후속 카드 t706)
- audit_multi 미호출 — 차단 결함이 기계적 열거값 확인이라 근거 없음 판단
