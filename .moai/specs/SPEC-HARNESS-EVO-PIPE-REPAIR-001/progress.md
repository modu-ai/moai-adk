# SPEC-HARNESS-EVO-PIPE-REPAIR-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-02
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md
author: manager-spec

### plan-auditor 감사 결과 (iteration 1/3, 2026-07-02)

- **Verdict: PASS-WITH-DEBT** · Overall Score: **0.85** (harmonic mean; Clarity 0.85 / Completeness 0.88 / Testability 0.78 / Traceability 0.90)
- 임계: Tier M 선언 0.80 충족. 단, frontmatter `tier:` 필드 부재 → 엄격 규칙상 Tier L 0.85로 감사(경계선 충족) — D4 참조.
- Must-Pass: MP-1 PASS (REQ-HEP-001~013 연속·중복 0, grep 실측) · MP-2 PASS (13 REQ 전건 GEARS When/While/Where/shall/shall-not) · MP-3 PASS (canonical 12필드 전건, rejected alias 0) · MP-4 N/A (다언어 툴체인 열거 없음)
- D7 cross-SPEC: 참조 4 SPEC 전건 `completed`(V4-001/PROPOSAL-GEN-001/CLASSIFIER-WIRING-001/LOOP-CLOSURE-001), retired/superseded 충돌 0; PROPOSAL-GEN-001 부분 supersede는 plan §H에 reconcile 명시 ✓. Epic 미작성 3 SPEC(RUN-REPORT/WRITE-SURFACE/REQ-ARTIFACT) NOT-FOUND는 의도된 forward-ref (SHOULD, 결함 아님).
- D8 syscall: 4 artifact 전건 0 match → auto-PASS.
- 앵커 검증: spec.md §B 인용 앵커 **전량 live 재검증 일치** (types.go:217-245, mapper.go:36/:41-44, learner.go:98-100, hook.go:148-163, frozen_guard.go:21-37, harness.go:486, v4lifecycle.go:32-36, doctor.go:194, Runner :1/:31/:56, harness.md:91/:135/:173/§2.2, learner SKILL:135-144, harness.yaml:119-121, settings grep). 실데이터 검증: usage-log 608 ✓ / promotions 16 ✓ / applied·proposals 부재 ✓ / pattern_key 3형 + confidence 1 + to_tier {observation,auto_update}만 존재 ✓. verification-claim-integrity §F 미검증 항목 명시 우수.
- **Defects (요지)**:
  - **D1 [major/Testability]** AC-HEP-013a의 8-표면 일괄 `diff live↔template` 검증법은 2/8 표면에서 false-fail: harness.yaml은 SPEC 이전부터 대규모 divergence(실측 diff 수십 hunk), settings.json↔settings.json.tmpl은 render-pair(템플릿 변수+bash 이중 항목, tmpl:116/118 vs local:82). "8개 전부 MIRRORED" 주장은 6/8만 byte-identical. → 표면 클래스별 검증(6 byte-diff / settings 토큰-grep / harness.yaml rate_limit 키 수준)으로 AC 재작성 필요.
  - **D2 [major/Clarity+Testability]** EC-4 전제 사실 오류: github/release는 "manifest 있음+Runner 부재"가 아니라 **manifest·Runner 모두 부재**(find 실측: manifest는 release-update/ 1개뿐; v4lifecycle.go:64-66 `ManifestMissing` partial state). REQ-HEP-005 축2(manifest 필수)대로 구현 시 doctor가 github/release 2건 finding → plan §E-7 "수리 후 0 findings" 전역 스모크 false-fail. → EC-4 전제 정정 + command-only thin 하네스에 대한 doctor 정책(skip/INFO vs finding) 정의 필요.
  - D3 [minor/Traceability] spec.md:142 "AC-HEP-001 ~ 015" — AC-HEP-015 비실존(최대 014); acceptance.md DoD-1 "001a ~ 013b"는 역으로 014 누락. 양단 범위 인용 드리프트.
  - D4 [minor/Completeness] frontmatter `tier: M` 필드 부재(부재=Tier L 0.85 감사 규칙); plan/progress 선언과 정합하도록 추가 권고.
  - D5 [minor/Clarity] REQ/AC-HEP-011 "frozen_guard.go와 문서 일치"가 agents/{harness,moai}만 명명 — learner SKILL L1의 `.moai/project/brand/**`(SKILL:142)는 frozenPrefixes에도 allowedPrefixes에도 없어 처분 미정의(naive 전체 목록 재작성 시 오정렬 위험).
  - D6 [minor/Testability] AC-HEP-004 Given "make build 후 렌더된 settings" — 로컬 .claude/settings.json은 make build 산출물 아님(직접 편집/moai init 렌더); AC-HEP-008 presence-grep은 파일 전역 산문과 충돌해 "(Phase 0 절)" 스코핑이 수동.
- 정보성: (i) 기존 hook_harness_observe_stop_test.go가 현행 Stop 핸들러 동작 고정 — M2 classify 연쇄는 fail-open으로 그린 유지 or 의식적 갱신 필요; (ii) template harness.md:91/:253에 REQ-HRN-FND-018 토큰 기존재(중립성 CI 허용 상태) — REQ-HEP-013 grep 가드가 REQ-HEP로 스코프된 것은 정확; (iii) `related_specs`는 비표준 optional 필드(표준은 depends_on, decoder 허용).
- 재감사 조건: D1·D2는 acceptance.md/plan.md 소폭 수정으로 해소 가능 — 수정 반영 시 iteration 2 불요(PASS-with-debt 승인, run-phase 진입 전 D1/D2 정정 의무).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
