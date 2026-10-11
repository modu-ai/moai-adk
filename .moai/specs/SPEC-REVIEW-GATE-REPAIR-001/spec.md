---
id: SPEC-REVIEW-GATE-REPAIR-001
title: "Codex 리뷰 게이트 수리 — 프롬프트 진입 교착을 풀고 리뷰 기준을 하나로 모은다"
version: "0.2.11"
status: completed
created: 2026-10-10
updated: 2026-10-11
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: L
tags: "codex, review-gate, turn-end, prompt-entry, receipt, merge-base, card-t1620"
amendment_of: SPEC-REVIEW-GATE-REPAIR-001
---

# SPEC-REVIEW-GATE-REPAIR-001 — Codex 리뷰 게이트의 프롬프트 진입 교착을 풀고 리뷰 기준을 하나로 모은다

카드: **t1620** (class C — 훅·영수증·리뷰 범위·설정 네 서브시스템을 가로지르는 설계 변경) · 기준 코드 트리 `2aab5f797` (작업 브랜치 `WT-review-gate-fix`, plan 커밋 `6218869fa` 위)

## HISTORY

- 2026-10-10 · v0.1.0 · manager-spec · 최초 작성(draft). Tier L 판정 근거는 plan.md §A. 측정 원천은 research.md(인용 전수 재검증, 기준선 셀 실행). 운영자 판정이 필요한 항목 Q1–Q6 은 decision-index.md 에 개방 상태로 두었다.
- 2026-10-10 · v0.1.1 · manager-spec · plan-audit 1회차 FAIL(score 0.625, MP-7 실패; 근거 `.moai/reports/t1620/plan-audit.md` §1.4 의 D1–D12) 을 수리했다. 운영자 판정 Q1–Q6 과 보드 판정 D3·D4·D11 을 decision-index.md 에 원문으로 기록했다(보드 기록 d-20261010T053034Z-c50f, d-20261010T053834Z-422b, d-20261010T054047Z-3fc2). 요구사항 16개, 인수 기준 20개(Tier L 상한은 각각 25). 의미를 바꾸지 않는 감사 선택 항목(O1–O11)을 적용했다. **Q6 처분**: SPEC-GATE-BOTTLENECK-001 의 REQ-GBN-002·REQ-GBN-004·AC-GBN-004 는 이 SPEC 의 HISTORY 와 F5 에만 기록하며, 그 SPEC 은 편집하지 않는다. 재감사를 통과하기 전에는 Implementation Kickoff 를 열지 않는다.
- 2026-10-10 · v0.1.2 · manager-spec · plan-audit 2회차 FAIL(codex, 영수증 rcpt-0efbcef22b780f388f658933)을 수리했다. **F-1**(감사자 주장, 레인은 임시 저장소로 재현하지 않음): 설계가 github-flow 에서 남아 있는 develop_branch 를 무시하지 않고 develop 을 항상 후보에 넣었다. 운영자 판정 Q1 은 (가) 합집합 유지 + 활성 흐름 가드로 개정되었다(영실이 판단, 운영자 위임, 보드 기록 d-20261010T063222Z-43e0). REQ-RG-006·007·008 을 `manual.workflow` 로 분기하는 네 절로 다시 썼다. 절 (c) 는 github-flow 에서 develop 을 후보에서 제외한다. 절 (d) 에는 리더가 확인한 판독에 따라 manual.workflow 부재, 알 수 없는 값, 두 키와의 모순을 inconclusive 조건으로 명시했다. 모순의 구체적 조합은 절 (a)–(c) 에서 확인되지 않았다(plan.md §F 위험 12). **F-2**: plan.md 의 마일스톤 중단 조건을 승인 파일 목록 기준으로 정정했다(workflow.yaml 금지는 §F 그대로). 같은 보드 기록이 보류 d-20261010T063038Z-f1a1 을 해소한다. 인수 기준 AC-RG-006·007·008·009 를 새 규칙에 맞추고 AC-RG-021~024 를 추가했다(총 24, 상한 25). 요구사항은 16 으로 유지했다. F6(문서 불일치의 t1635 이관, 기록만)을 등록했다. 리더 판정 d-20261010T065057Z-229e 를 반영했다: progress.md 를 M1–M7 공통 승인 파일로 plan.md §C 에 등록했고, 절 (d) 는 두 키를 읽을 수 없으면 항상 적용한다. manual.workflow 가 없거나 알 수 없는 값이거나 두 키와 모순되면, 두 키가 서로 다른 브랜치를 가리킬 때에만 적용한다. 절 (a) 가 맞으면 manual.workflow 를 읽지 않는다. 이 판독은 리더가 이의 없음으로 확인했다. [ID mapping v0.2.2: 적용되는 009–016 대응 없음. 이 행의 REQ-RG-006·007·008 은 v0.2.0 에서 폐기된 식별자다.]
- 2026-10-10 · v0.2.0 · manager-spec · 분할 판정 d-20261010T070728Z-6cdf(영실이 판단, 운영자 위임)에 따라 범위를 카드 항목 (1)·(3)·(5) 로 좁혔다 (판정 121e 옵션 (a): (6)~(10) 은 t1637, (2)·(4) 는 t1662, (6) 중 REQ-RG-016 은 이 SPEC). 계획 감사 3회차 FAIL(codex, 영수증 rcpt-3e341c03b359a26ff7d01963)의 발견 F-3(활성 모드를 무시하고 `manual.workflow` 를 읽는 P2 결함)은 부채가 아니라 결함이므로 이 SPEC 에서 갚지 않고, 기준 브랜치 선택 부분을 후속 카드 t1662 로 떼어 냈다(REQ-ACE-005). 제거된 요구사항: REQ-RG-005, REQ-RG-006, REQ-RG-007, REQ-RG-008. 제거된 인수 기준: AC-RG-006, AC-RG-007, AC-RG-008, AC-RG-009, AC-RG-016, AC-RG-020, AC-RG-021, AC-RG-022, AC-RG-023, AC-RG-024. 모두 t1662 로 이관되었다. REQ-RG-005 와 AC-RG-006 은 같은 공용 해석 함수의 설계와 소비자 일치를 다루므로 조정 요청에 따라 함께 이관했다. 남은 요구사항은 12개(REQ-RG-001–004, REQ-RG-009–016), 인수 기준은 14개(AC-RG-001–005, AC-RG-010–015, AC-RG-017–019)다. 범위 축소 판정 d-20261010T072237Z-121e(영실이 판단, 운영자 위임)에서 option (a) 를 선택했다. 이어 리더 판정 d-20261010T072546Z-6d67(option (b))로 REQ-RG-016 과 묶인 AC 는 이 SPEC 에 둔다. 이 SPEC 은 카드 항목 (1)·(3)·(5)와 (6) 중 REQ-RG-016 부분만 다루며, (2)·(4)는 t1662, (6)~(10)은 t1637(3.2-1-4)로 이관하되 (6) 중 REQ-RG-016 부분은 이 SPEC 에 남는다. ID 는 재번호 매기지 않았으며 번호 간격은 제거된 ID 다. 이전 기록 행은 편집하지 않았다. [ID mapping v0.2.2: REQ-RG-009→REQ-RG-005, REQ-RG-010→REQ-RG-006, REQ-RG-011→REQ-RG-007, REQ-RG-012→REQ-RG-008, REQ-RG-013→REQ-RG-009, REQ-RG-014→REQ-RG-010, REQ-RG-015→REQ-RG-011, REQ-RG-016→REQ-RG-012. 이 행의 REQ-RG-005–008 은 폐기된 식별자다.]
- 2026-10-10 · v0.2.1 · manager-spec · 계획 감사 1회차 (v0.2.0) FAIL, 발견 F-4 (plan.md:104)에 따라 §C 승인 목록에 명시된 테스트 파일은 다음 네 개다. M5 기존 테스트 파일 `internal/cli/codex_review_cache_test.go`, M2 기존 테스트 파일 `internal/cli/codex_review_delay_test.go`, M2 신규 RED 테스트 파일 `internal/cli/codex_review_entry_delivery_test.go`, M5 신규 RED 테스트 파일 `internal/cli/codex_review_inconclusive_test.go`. 리더 판정 d-20261010T073725Z-791c 와 승인에 따름.
- 2026-10-10 · v0.2.2 · manager-spec · plan-audit 5회차 FAIL(`.moai/reports/t1620/plan-audit-iter5.md`, MP-1 실패, 필수 결함 D1·D2·D3)의 수리 라운드다. 운영자 위임 판정 d-20261010T080954Z-4294(kind=ceiling-exception, 영실이 판단)에 따라 D1·D2·D3 만 고쳤고, 판정 d-20261010T081517Z-9c46 에 따라 완결성 항목 두 건(design.md §4 본문, plan.md 의 M4 자리표시 제목 제거)을 같은 라운드에 더했다. 선택 항목 O1–O9 과 R2 는 건드리지 않았으며, 이 행은 감사 결과를 주장하지 않는다. **D1(연속 번호)**: 요구사항을 REQ-RG-009→REQ-RG-005, REQ-RG-010→REQ-RG-006, REQ-RG-011→REQ-RG-007, REQ-RG-012→REQ-RG-008, REQ-RG-013→REQ-RG-009, REQ-RG-014→REQ-RG-010, REQ-RG-015→REQ-RG-011, REQ-RG-016→REQ-RG-012 로 재배열했다. REQ-RG-001–004 와 모든 AC-RG 번호는 그대로다. 번호 빈자리를 없애려고 REQ-RG-005–008 은 살아 있는 번호로 다시 쓰였다. 폐기된 식별자 표는 §B 에 두었고, 과거 행과 날짜가 붙은 기록의 원문은 고치지 않고 대응 주석만 달았다. **D2(추적성)**: plan.md §F 위험 1·12·13 의 폐기된 AC-RG 참조와 'M4 의 RED' 문구를 t1662 이관 항목으로 바꿨다(REQ-ACE-005, 이 SPEC 에서 구현하지 않음). design.md §7 의 메시지 상태 어휘에서 '기준 판정 inconclusive(두 이름 포함)'를 뺐다. **D3(설계)**: 재정의 파일의 형식(JSON 객체, 필드 `run_id`·`codex_review_gate`), 무시 조건, 단일 활성 run 해석 규칙을 design.md §5 에 고정했다(코드 근거: `internal/factorymsg/store.go:343–395`, `internal/cli/factory_card.go:2313–2322`). 연기 문장 네 곳은 design.md §5 참조로 바꿨다(spec.md REQ-RG-007 본문, design.md §5 재정의 파일 항목, decision-index.md Q3, plan.md 위험 7). AC-RG-012·013 의 Given 은 design.md §5 를 인용하도록 했고 Then 은 바꾸지 않았다. version 0.2.1→0.2.2.
- 2026-10-10 · v0.2.3 · manager-spec · 최종 문서 수리 라운드. 운영자 위임 판정 d-20261010T090503Z-8963(영실이 판단, 리더 전달)을 적용했다. 리더 정정에 따라 살아 있는 번호는 v0.2.2 그대로 REQ-RG-001–012 로 둔다(판정 (1)의 번호 복원은 적용하지 않았다). (1) 폐기된 식별자 표는 요구사항 본문에서 빼고 이 행으로 옮겼다. v0.2.2 번호 주석의 내용도 이 행에 남긴다. REQ-RG-005–008 은 v0.2.2 에서 살아 있는 요구사항의 번호로 다시 쓰였고, 폐기된 식별자는 아래 두 줄과 과거 기록 행에만 남는다. (2) 진행 중 표식은 검토가 어떤 판정(pass, fail, inconclusive)으로 끝나든 완료 시점에 풀리고, 소유자 pid 가 살아 있지 않은 고아 표식은 새 판정 전에 회수된다. 같은 커밋의 재검토는 결과 캐시(pass·fail)가 막으며, 표식은 동시 실행만 막는다(REQ-RG-006, REQ-RG-011, design.md §2·§3, AC-RG-011). (3) 재정의 파일은 그 run 의 member lane(카드 워크트리)에만 적용되며, 저장소 전체 스위치는 없다(REQ-RG-007, design.md §5, AC-RG-012 시나리오 (e)). (4)·(5) codex 필수 게이트의 대상과 plan-aware 영수증은 t1662 로 이관한다(Out of Scope, F1). (6) 레인 판단으로 보정했다(리더 확인 대상). plan.md 의 M5 RED·GREEN 에 '살아 있는' 한정, 완료 시 해제, 고아 회수를, M6 GREEN 에 member lane 검사를 적었고, plan.md:5 의 폐기 식별자에 판 번호 v0.1.2를 붙였다. design.md §3 (다)에 pid 를 읽을 수 없는 표식의 처리(고아로 회수, fail-open)를, §5 에 경로 정규화 실패의 처리(member 아님)를 더했다. acceptance.md AC-RG-011 에 (iii) 시나리오를 더했고, decision-index.md 에 4회차 이후 판정 기록 절을 더했다. 이 행은 감사 결과를 주장하지 않는다. [ID mapping v0.2.3: 적용 없음. 번호는 v0.2.2 와 같다.]
  - v0.1.2의 REQ-RG-005~008(퇴역 → t1662, v0.2.0에서 제거)
  - v0.1.2의 AC-RG-006~009·016·020~024(퇴역 → t1662, v0.2.0에서 제거)
- 2026-10-10 · v0.2.4 · manager-spec · plan-audit 7회차(v0.2.3, `.moai/reports/t1620/plan-audit-iter7.md`, FAIL 0.75, 차단 B1–B4, 필수 codex 레그 inconclusive)의 수리 라운드다. 운영자 결정 d-20261010T113614Z-2413 에 따라 이 카드의 plan 감사에서 codex 필수 레그를 advisory 로 바꾸고(도구 결함, t1662 범위, 판정 파일에 기록), B1–B4 를 한 번 고친다. B1: AC-RG-019 에 (d) reviewer 없음을 더했다. B3: AC-RG-019 의 각 경우를 상태 이름에 묶었다. B4: AC-RG-012 (a) 의 systemMessage 는 턴 종료 게이트만 낸다. B2: 마일스톤의 RED 선행은 커밋 그래프로 증명할 수 없으므로, 순서 조항을 기록 증거(progress.md §E.2)로만 확인한다는 문장으로 바꿨다(plan.md, spec.md §D.2, acceptance.md). 이 행은 감사 결과를 주장하지 않는다. [ID mapping v0.2.4: 적용 없음. 번호는 v0.2.2 와 같다.]
- 2026-10-10 · v0.2.5 · manager-spec · plan-audit iter8 FAIL(`.moai/reports/t1620/plan-audit-iter8.md`, 0.75, 차단 D1·D2)의 수리 라운드다. 운영자 결정 d-20261010T130929Z-a280 에 따라 D1 은 (나)안으로 고쳤다(RED 선행 순서를 커밋 그래프 검사로 재작성, 이 SPEC 은 승격을 하지 않는다). D2 는 plan §C 승인 파일 목록에 M3·M5·M6 의 테스트 파일과 원천 파일을 보강했고, 파일 목록은 §C 가 정한다고 적었다. 이 행은 감사 결과를 주장하지 않는다. [ID mapping v0.2.5: 적용 없음. 번호는 v0.2.2 와 같다.]
- 2026-10-10 · v0.2.6 · manager-spec · plan-audit iter9 FAIL(`.moai/reports/t1620/plan-audit-iter9.md`, 0.75, 차단 D1·D2·D3)의 수리 라운드다. 운영자 결정 d-20261010T135342Z-51a4 (가)에 따라 감사자가 제시한 문장을 그대로 적용했다. D1: 덮어쓰기 단위를 plan.md:61, spec.md §D.2, acceptance.md §E 세 곳에 같은 말로 정했다. D2: 완료 정의에 구현 뒤 GREEN 요구를 되살렸다. D3: M2·M5·M6 의 테스트 파일을 plan §C 에 묶고 기대값 변경 목록에 더했다. M6 셸 래퍼의 메시지 방식은 바이너리 호출로 정했다(REQ-RG-008, design.md §5). 이 행은 감사 결과를 주장하지 않는다. [ID mapping v0.2.6: 적용 없음. 번호는 v0.2.2 와 같다.]
- 2026-10-10 · v0.2.7 · manager-spec · plan-audit iter10 FAIL(`.moai/reports/t1620/plan-audit-iter10.md`, 0.75, 차단 B1·B2)의 뒤 분할 개정이다. 운영자 결정 51a4 의 (나) 조건과 리더 결정 d-20261010T143913Z-d5a3 에 따라 M5·M6 (요구 REQ-RG-005–008, AC-RG-010–013, 파일 묶음)을 후속 카드 t1674 로 옮겼다(depends t1620). B2 컴파일 전제 한 문장을 M2·M3 의 덮어쓰기 검사에 더했다. 이 행은 감사 결과를 주장하지 않는다. [ID mapping v0.2.7: 살아 있는 번호는 그대로다. 이관된 ID 는 이 행과 과거 기록 행에만 남는다.]
- 2026-10-10 · v0.2.8 · manager-spec · plan-audit split-iter1 FAIL(`.moai/reports/t1620/plan-audit-split-iter1.md`, 0.75, MP-1 실패, 차단 D1–D5)의 수리 라운드다. 운영자 위임 판정 d-20261010T151958Z-1adf 에 따라 REQ-RG-011 을 t1674 로 옮기고 살아 있는 REQ 를 001–007 로 다시 매겼다(구→신: 009→005, 010→006, 012→007). D3 의 inconclusive 메시지 AC(AC-RG-025)를 더했다. D4·D5 는 감사자 문장대로 고쳤다. 이 행은 감사 결과를 주장하지 않는다.
- 2026-10-11 · v0.2.9 · manager-spec · 메시지 상태 어휘를 아홉 상태에서 열두 상태로 늘렸다. 정책 건너뜀 두 경로에 상태 이름을 더했다: `stale 바이너리 건너뜀`, `tree_scope 건너뜀`. 현재 코드는 이 두 경로에서 메시지를 내지 않는다(progress.md §E.2, M3). `차단 횟수 기록 실패`를 `차단 횟수 소진`에서 분리했다. 이제 `차단 횟수 소진`은 차단 횟수가 C 에 이른 상한 경로만 가리키며, 차단 횟수를 읽거나 쓰지 못해 허용하는 경로는 `차단 횟수 기록 실패`를 낸다. 코드가 내는 문구를 이 어휘에 맞추는 일은 sync 감사 항목이다. 허용 결정은 바뀌지 않는다(REQ-RG-006). 이 행은 감사 결과를 주장하지 않는다.
- 2026-10-11 · v0.2.10 · manager-spec · 완료 SPEC 의 제자리 수정(amendment)이다. `completed → in-progress (amendment)` 전환을 적용했고 근거는 `## Amendments` 절에 있다. 요구의 문장과 번호, 인수 기준 번호는 바꾸지 않았다. REQ-RG-007 에 배포 래퍼의 조용한 종료 네 곳(35·36·60·79–80행)을 열거하고 범위 한계로 문서화했으며, 인수 기준 §F 와 설계 §7 에 같은 열거를 더했다. 행위 변경과 래퍼 수정은 없다(t1674 몫). 이 행은 감사 결과를 주장하지 않는다.
- 2026-10-11 · v0.2.11 · manager-spec · 델타 감사(iter2) 발견 F1 을 수리했다(본문의 Follow-up F1 과 다른 번호). REQ-RG-007 범위 한계 절은 래퍼 종료가 네 곳이라고 적었으나 사실과 달랐다. 래퍼 83–86행은 처리기 호출 뒤 종료 상태를 확인하지 않고 `exit 0` 으로 끝나 처리기 실패를 삼키며, Go 765–769행은 빈 입력과 형식 오류 입력을, Go 773–777행은 처리기 오류를 상태 이름 없이 허용한다. 세 경로를 표에 더하고 열거가 완결이라는 주장을 뺐다. 빈 입력 경로에는 상태 이름을 더하지 않았고 이 SPEC 의 잔여 위험으로 기록했다. 래퍼 명명은 t1674 의 몫이며 이 SPEC 은 래퍼를 바꾸지 않는다. 상태 어휘는 열두 상태 그대로이고 REQ-RG-007 규범 문장은 바꾸지 않았다(어휘 변경 없음). frontmatter `status` 는 `completed` 에서 `in-progress` 로 되돌렸다(`completed → in-progress (amendment)`, 기록은 `## Amendments` 절). 이 행은 감사 결과를 주장하지 않는다.

## Amendments

- **이전 완료 버전**: 0.2.10. `status: completed` 는 3-phase 재종결 커밋 `c861336ee6` 에서 설정되었다. v0.2.11 개정은 이 완료 상태에서 `completed → in-progress (amendment)` 로 시작한다. (앞선 v0.2.10 개정의 이전 완료 버전은 0.2.9 였고, 그 종결은 sync 단일 커밋 `7716c99ff` 다.)
- **prior_completed_sha**: c861336ee64bbb81f304eec02bea7a9f97dd5af2. `progress.md` §E.4 의 `sync_commit_sha` 와 같은 값이다(backfill 커밋 `6995a1696a` 가 자리표시를 이 값으로 바꿨다). 앞선 v0.2.10 개정의 prior_completed_sha 는 `7716c99ff2347dc004ee0e209791fc648772e1ef` 였다.
- **사유**:
  - sync 감사 FAIL. 발견 SF-1(REQ-RG-004, 키당 차단 횟수가 동시 턴 종료에서 상한 C 를 넘는 경합)은 `4bd5d433a`(RED, 배리어 테스트)와 `a13e02705`(GREEN, 차단 횟수 확인과 기록을 키별 잠금 한 단계로 묶음)로 수리했다. 수리 증거는 `progress.md` §E.2 의 'SF-1 수리' 절과 커밋 `e62a8bdd5` 에 기록했다.
  - 발견 SF-2(REQ-RG-007, 배포 래퍼의 조용한 종료 경로): 래퍼가 Go 처리기를 부르기 전에 네 곳에서 exit 0 으로 끝나는데, 이 SPEC 본문은 60행 한 곳만 적고 있었다. 이번 수정은 네 곳을 명명하고 범위 한계로 문서화한다.
  - (v0.2.11 개정) 델타 감사(iter2) 발견 F1(본문 Follow-up F1 과 다른 번호): 범위 한계 절은 래퍼 종료가 네 곳이라고 적었으나, 래퍼 83–86행(처리기 호출 뒤 `exit 0`)과 Go 765–769행·773–777행(상태 이름 없는 허용)이 빠져 있었다. 표를 사실에 맞춰 고쳤다. 상태 어휘와 REQ-RG-007 규범 문장은 바꾸지 않았다.
- **범위**:
  - v0.2.10: REQ-RG-004 수리 증거의 연결(SF-1), REQ-RG-007 의 래퍼 종료 경로 열거와 범위 한계(SF-2), 인수 기준 §F, 설계 §7. 요구 문장과 번호는 바꾸지 않았다.
  - v0.2.11: REQ-RG-004 키별 차단 잠금 경로의 심볼릭 링크 거부(F2 수리; 본문 Follow-up F2 와 무관한 번호). 커밋 `a93efb76d3`(테스트 RED, `codex_review_gate_symlink_test.go`), `40f56c3c70`(수리, `codex_review_gate.go`·`gate_lock_unix.go`·`gate_lock_windows.go`), `d480f2704d`(§E.2 증거). REQ-RG-007 범위 한계 절의 정정(F1), 인수 기준 §F, 설계 §7 의 같은 정정. 요구 문장과 번호는 바꾸지 않았다.
  - 이 SPEC 의 v0.2.11 문서 커밋은 래퍼, `progress.md`, `plan.md`, `research.md`, `decision-index.md` 를 바꾸지 않는다. 구현 파일 변경은 위 F2 커밋 셋뿐이다.

---

## §0 지배 원칙 [HARD]

> **턴 종료 게이트만 차단하고, 프롬프트 진입은 차단하지 않는다. 검토 없이 턴 종료가 허용될 때마다 화면에 드러난다.** (검토가 이뤄지지 않은 상태의 재사용 금지는 t1674 몫이다.)

교착의 구조는 이렇다. 진입 단계에서 막힌 프롬프트는 모델에게 턴을 주지 않으므로, 리뷰 키가 움직일 방법이 없다(research.md §4, 설계 분석). 모델이 조치할 수 있는 지점은 턴 종료뿐이다. 그래서 검토 결과는 진입 시 추가 맥락으로 한 번 전달되고, 차단은 턴 종료에서만 일어나며, 그 차단은 키당 총 C 회로 상한을 가진다(C = 1). 검토되지 않은 채 허용되는 턴 종료는 systemMessage 로 드러난다(REQ-RG-007). 이 교착의 관측 범위와 관측하지 못한 전제는 §A.1 에 적었다.

---

## §A 배경 (측정)

### A.1 관측 사실 — 기계 로컬 fail 스냅샷 한 건 (recorded_at 2026-10-10T01:37:46+09:00, key 2aab5f797…:7f1e2657f0831d9f)

- 관측: 레인 저장소의 `.moai/worktrees/t1543/.moai/state/verify/snapshots/2aab5f797b75-7f1e2657f0831d9f.json`(기계 로컬, 커밋 대상 아님)에서 `codex-review` 검사는 `verdict: fail`, `exit_code: 1`, `recorded_at: 2026-10-10T01:37:46.265725+09:00`, `tool_version: codex-cli 0.161.0` 이다. 키는 `2aab5f797…:7f1e2657f0831d9f` 이며, 앞부분은 전체 SHA `2aab5f797b75983e132af451da68f69e3426557b`(이 트리의 기준 커밋)의 접두사다.
- 상세 파일의 발견 사항 두 건은 `internal/homestate/card_transition.go:287` 과 `internal/cli/factory_card.go:2894` 를 가리키며, 두 경로는 이 트리에 있다. 발견 사항이 옳은지는 이 SPEC 이 판단하지 않는다(research.md §4).
- 미관측(Gap) 세 항목:
  1. 레인 자체가 카드 변경을 하지 않았다는 전제. 이 세션의 git 가드가 다른 워크트리 대상 호출을 거부했다(research.md §8 Gap 1).
  2. 24시간 잠금 기간. 기록 시각에 유효기간 24시간(`internal/cli/codex_review_cache.go:29`)을 더한 값은 산술 결과이며 관측이 아니다(research.md §4).
  3. 진입 훅이 실제로 프롬프트를 막았다는 사실. 재현하지 않았다(research.md §3 #22, §8 Gap 2).

### A.2 리뷰 기준은 두 해석기가 서로 다른 설정 키를 읽는다

- 턴 종료 게이트와 card 범위 리뷰는 `cardMergeBase` 로 기준을 정한다. 이 함수는 `IntegrationTarget` 을 읽고(`internal/cli/codex_review_scope.go:189-191`), 그 값은 git-flow 수동 프로필에서 `develop_branch`(`.moai/config/sections/git-strategy.yaml:16`, 값 `develop`; 키 경로 `git_strategy.manual.develop_branch`)다(`internal/config/loader_integration_branch.go:123`).
- 감사 도구는 `resolveReviewBase` 를 쓰며, 이 함수는 첫 단계로 `worktree_base_branch` 를 읽는다(`internal/cli/mcp_review_material.go:159`, 값 `main`, `git-strategy.yaml:5`). 이 단계는 커밋 `f82a078bb`(2026-10-03, 카드 t1426)가 추가했다.
- 토폴로지(git 측정): `merge-base(main, develop) = db0c514d3`, `db0c514d3..main` 139 커밋, `db0c514d3..develop` 149 커밋.
- 같은 저장소에서 두 해석기는 서로 다른 기준을 고른다. 두 해석을 하나의 함수로 모으는 이유다(t1662, REQ-ACE-005).

### A.3 완료된 SPEC 의 결정이 코드에서 뒤집혀 있다

- SPEC-CODEX-REVIEW-TARGET-001(status: completed) §A.7 은 codex 네이티브 경로가 `worktree_base_branch` 를 읽지 않는다고 결정했다(`.moai/specs/SPEC-CODEX-REVIEW-TARGET-001/spec.md` 102–111행).
- 카드 t1426 이 이 결정을 뒤집었다. 감사 기준 테스트의 주석이 이를 명시한다: "Card t1426 reverses the spec.md §A.7 non-read clause"(`internal/cli/codex_review_target_test.go`, AC-CRT-002 절).
- 해당 SPEC 의 본문에는 t1426 을 가리키는 항목이 없다(`grep -c t1426` = 0). 결정의 기록은 코드와 테스트뿐이다. 운영자 판정은 Q1 에서 받았다.

### A.4 진입 차단은 진행 중 SPEC 의 명시 요구다, 기록 처분은 Q6 이다

- SPEC-GATE-BOTTLENECK-001(status: in-progress)의 REQ-GBN-002(`spec.md:43–47`): "WHEN no fresh receipt exists at Stop, THE GATE SHALL allow the turn, start the review in the background, and leave the verdict to be enforced at the NEXT turn entry; a fresh FAIL at turn entry SHALL block. `async:true` SHALL NOT be used." 진입 훅의 헤더 주석(`internal/cli/codex_review_delay.go` 1–14행)도 같은 설계를 적었다. REQ-RG-001 은 이 진입 차단 절을 뒤집는다.
- 같은 SPEC 의 REQ-GBN-004(`spec.md:51–54`)는 모든 경로가 fail-open(missing/inconclusive reviewer 는 허용)을 지키도록 요구하고, AC-GBN-004(`spec.md:67–68`)의 inconclusive 절은 실패 시 허용을 단언한다. 이 SPEC 은 REQ-GBN-004 를 뒤집지 않는다. REQ-RG-001 은 inconclusive 와 모든 진입 상태에서 허용을 유지한다.
- 기록 처분 Q6 = (b)(보드 기록 d-20261010T054047Z-3fc2): 이 SPEC 의 HISTORY 와 F5 에만 기록하고, 대상 SPEC 은 편집하지 않는다. 대상 SPEC 이 in-progress 인 채로 코드와 어긋나는 잔여 위험은 F5 에 적었다.

### A.5 영수증과 재사용 규칙

- 영수증의 결합 필드는 다섯 개다: head, tree_digest, config_digest, command, tool_version(`internal/verify/receipt.go:70-76`). 유효기간은 24시간이다(`internal/cli/codex_review_cache.go:29`).
- 현재 턴 종료 게이트는 수용된 영수증이 pass 든 inconclusive 든 허용한다(`internal/cli/codex_review_gate.go:176-178`). 관측: `TestReviewGate_InconclusiveReceiptIsNotPassEvidence` 가 통과하며, stderr 에 "reusing the cached inconclusive verdict for the unchanged tree" 를 남긴다(acceptance.md §C, 셀 B-2).

### A.6 설정 상태

- 추적 파일 `.moai/config/sections/workflow.yaml` 302행의 `codex.review_gate.enabled` 는 HEAD 에서 `true` 다(`git show HEAD:` 로 확인). 이 워크트리의 작업 트리는 HEAD 와 같다(`git status --short` 출력 없음, plan 단계 측정). 앞선 측정 세션이 기록한 리더의 미커밋 변경(`false`)은 이 워크트리에서 재현되지 않았다(research.md §5 P-10 정정, Gap).
- 스위치가 없을 때의 기본값은 커밋된 값이다(보드 판정 D11). 현재 Go 층은 판독 트리의 작업 트리 파일을 읽으므로(`internal/cli/mcp_codex.go:2503-2510`), 기본값을 커밋된 값으로 바꾸는 것은 후속 카드 t1674 의 행위 변경이다.

### A.7 카드의 두 주장 정정

- 카드가 적은 `internal/hook/codex_review_*.go` 경로는 이 트리에 없다. 네 파일은 모두 `internal/cli/` 에 있다(research.md §2).
- "두 층을 모두 켜면 doctor 경고가 뜬다"는 주장은 코드에서 관측되지 않았다. `internal/cli/doctor*.go` 71개 파일 전체 검색에 해당 검사가 없고(research.md §5 P-6), 문서 한 문장(`.claude/rules/moai/workflow/factory-dispatch-detail.md:157`)만 있다(Q5).

---

## §B 요구사항 (GEARS)

### REQ-RG-001 — 진입 훅은 프롬프트를 막지 않는다 (Ubiquitous, M2)

The codex review entry hook shall not return a block decision for any receipt state.

어떤 영수증 상태에서도 진입 훅은 차단 결정을 내지 않는다(absent, stale, key-mismatch, fresh-pass, fresh-fail, fresh-inconclusive 모두 포함). 이 요구가 교착을 끊는다. 발견 사항의 전달은 REQ-RG-002 가 맡는다. 이 요구는 SPEC-GATE-BOTTLENECK-001 REQ-GBN-002 의 진입 차단 절을 뒤집는다(기록 처분 Q6 = (b), §A.4).

### REQ-RG-002 — 발견 사항은 키마다 한 번 추가 맥락으로 전달된다 (event-driven, M2)

When a prompt is submitted and a fresh fail receipt exists for the current review key, the entry hook shall add the fail detail to the model context as additional context exactly once for that key, and shall record that delivery against that key.

전달 수단은 UserPromptSubmit 출력의 `additionalContext` 다(`.claude/rules/moai/core/hooks-system.md:91`). 이미 전달된 키는 프롬프트마다 반복하지 않는다. 리뷰 키(영수증 결합 필드 다섯 개의 조합)가 바뀌면 새 전달 기록이 생긴다.

### REQ-RG-003 — 차단은 턴 종료에서만, 키당 C 회까지 한다 (event-driven, M3)

When a turn ends with a fresh fail receipt for the current review key, the block count for that key is below C, and the stop has not been released by the stop-hook-active flag, the turn-end gate shall block the stop and shall give the fail detail as the reason.

현행 턴 종료 차단을 유지한다. 모델이 조치할 수 있는 유일한 지점이기 때문이다(§0). 차단할 때마다 그 키의 차단 횟수를 1 올린다(design.md §3). C 의 값은 운영자 판정 Q2 에 따라 1 이다(REQ-RG-004).

### REQ-RG-004 — 키당 차단 횟수는 C 로 총량 제한된다 (event-driven, M3, C = 1)

When a turn ends with a fresh fail receipt for a review key whose block count has reached C, the turn-end gate shall allow the stop and shall show the unreviewed state in a visible system message.

차단 횟수는 리뷰 키 단위의 총량이다(보드 판정 D3 = 키당 총 C회로 제한 (허용이 카운터를 되돌리지 않음)). 허용은 차단 횟수를 되돌리지 않는다. 그 키에 fresh pass 영수증이 기록되면 그 키의 차단 횟수를 제거한다. 키가 바뀌면 새 키는 0 에서 시작한다. C = 1(운영자 판정 Q2)이므로, 같은 fail 키는 처음 한 번만 차단되고 그 뒤의 턴 종료는 모두 허용되며 메시지가 표시된다. 이 규칙은 턴 단위로 카운터를 되돌리는 방식이 만드는 차단·허용의 반복(sawtooth)을 막는다. 허용 뒤의 다음 턴 종료는 AC-RG-017 이 확인한다.

### REQ-RG-005 — 현재 pass 영수증은 재검토 없이 허용한다 (State-driven, 회귀)

While a fresh pass receipt exists for the current review key, the entry hook and the turn-end gate shall allow without starting a new review.

### REQ-RG-006 — 검토 건너뜀의 결정은 바뀌지 않는다 (Ubiquitous, 회귀)

The codex review gate shall keep its current allow decision for disabled gates, active stop-hook releases, trees without reviewable changes, and absent reviewer binaries.

결정(허용)은 그대로다. 이 경우 턴 종료에 붙는 메시지는 REQ-RG-007 이 더한다. 메시지는 결정을 바꾸지 않는다.

### REQ-RG-007 — 검토 없이 허용되는 턴 종료는 상태를 명명한 메시지를 낸다 (event-driven, M3, 보드 판정 D4)

When the turn-end gate allows a stop and no fresh passing receipt backs that allow, the turn-end gate shall emit a visible system message that names the state.

보드 판정 D4 = 항상 메시지 표시. 이 문장은 "allows a stop without a fresh passing review" 를 문언대로 옮긴 것이다. 따라서 absent, stale, key-mismatch, inconclusive, 차단 횟수 소진, 차단 횟수 기록 실패, 스위치 꺼짐, 검토할 변경 없음, stop-hook 해제, reviewer 없음, stale 바이너리 건너뜀, tree_scope 건너뜀이 모두 포함된다. 키가 없는 상태(스위치 꺼짐, 검토할 변경 없음)도 포함하는 것은 §0 의 "검토가 되지 않은 상태는 드러나야 한다"에 따른다. 메시지는 Stop 훅의 `systemMessage` 로 낸다(`.claude/rules/moai/core/hooks-system.md:103`). 허용 결정 자체는 바뀌지 않는다(REQ-RG-006). 상태 어휘는 design.md §7 이 정한다. 배포 경로의 한계는 아래 범위 한계 절에 적는다. 그 절은 턴 종료 게이트 래퍼가 Go 처리기를 부르기 전이나 뒤에 상태 없이 끝나는 경로와, 처리기가 입력을 읽지 못하거나 오류를 낼 때 상태 없이 허용하는 경로를 적으며, 열거가 완결이라고 주장하지 않는다.

**범위 한계 (documented scope limit, v0.2.10; 정정 v0.2.11).** 위 규범 문장은 Go 처리기가 내는 메시지를 요구한다. 아래 표는 이 SPEC 이 읽어서 확인한 상태 없는 허용 경로를 적는다. 표는 완결된 열거가 아니다. 처리기 본체(`HandleCodexReviewGate`)의 허용 분기는 이번 정정에서 전수 대조하지 않았다. 778–780행의 빈 출력 대체 분기는 읽었으나, 처리기가 nil 출력을 돌려주는 경로가 있는지는 확인하지 않았다(acceptance.md §F). 턴 종료 게이트 래퍼(`.claude/hooks/moai/handle-codex-review-gate.sh`)의 템플릿 미러 `internal/template/templates/.claude/hooks/moai/handle-codex-review-gate.sh` 는 내용이 같다(이번 정정에서 `diff` 로 확인). 앞의 네 행(35·36·60·79–80행)은 v0.2.10 의 처분을 그대로 둔다. 뒤의 세 행은 v0.2.11 에서 더했다. 앞의 네 행의 관측은 v0.2.10 기록이며 이번 정정에서 재실행하지 않았다. 각 경로의 선행 조건만 실패하도록 격리한 래퍼 실행에서 네 경우 모두 종료 코드 0, 표준 출력·표준 오류 0바이트였다. 뒤의 세 행에서는 래퍼가 처리기의 실패를 삼키거나(83–86행) 처리기가 상태 이름 없이 허용한다(765–777행). 따라서 이 요구는 표에 적힌 경로를 뺀 나머지 경로에서만 충족되며, 나머지 경로 가운데 전수 대조하지 않은 부분이 있음은 위 열거의 한계와 같다. 처분은 아래 표와 같다.

| 줄 | 종료 조건 | 처분 |
|---|---|---|
| 35행 | `workflow.yaml` 파일이 없다 | Go 처리기가 이 경로에서 실행되지 않으므로 design.md §7 의 상태 어휘에 이름이 없다. t1674 wrapper scope(후속 카드 t1674 의 래퍼 범위)로 배정한다. |
| 36행 | `awk` 명령이 없다 | Go 처리기가 이 경로에서 실행되지 않으므로 design.md §7 의 상태 어휘에 이름이 없다. t1674 wrapper scope 로 배정한다. |
| 60행 | 게이트가 꺼져 있거나 설정을 읽지 못한다 | 상태 이름 `스위치 꺼짐` 은 어휘에 있으나, 이 경로에서는 래퍼가 먼저 끝나 메시지가 나오지 않는다. 기존 처분을 유지한다: 후속 카드 t1674. |
| 79–80행 | moai 바이너리를 세 단계 모두에서 찾지 못한다 | Go 처리기가 이 경로에서 실행되지 않으므로 design.md §7 의 상태 어휘에 이름이 없다. t1674 wrapper scope 로 배정한다. |
| 83–86행 (래퍼) | 83행의 처리기 호출 `"$MOAI_BIN" hook codex-review-gate` 뒤의 `exit 0`(86행)은 처리기의 종료 상태를 확인하지 않는다. 처리기 프로세스가 0 이 아닌 상태로 실패하면 래퍼는 상태 이름 없이 0 으로 끝난다. 84–85행의 주석은 `exit 0` 의 이유를 차단 결정이 표준 출력 JSON 으로 실리기 때문이라고만 적고 처리기 실패는 다루지 않는다. | 상태 이름을 더하지 않는다(어휘 불변). 래퍼 명명은 t1674 의 몫이다(t1674 카드 본문의 래퍼 범위). 이 SPEC 은 래퍼를 바꾸지 않는다. 관측(v0.2.11 정정, 실행): 종료 3 으로 실패하는 스텁 `moai` 를 `CLAUDE_PROJECT_DIR` 아래에 둔 격리 실행에서 래퍼는 종료 0, 표준 출력 0바이트, 표준 오류는 스텁의 한 줄이었다. |
| Go 765–769행 (`internal/cli/codex_review_gate.go`, `runCodexReviewGate`) | 표준 입력을 JSON 으로 해석하지 못하면(빈 입력과 형식 오류 입력 모두. `readHookInput`(802–812행)의 `json.Unmarshal` 이 오류를 돌려준다) 765행의 분기가 표준 오류에 진단 한 줄을 쓰고 빈 `hook.HookOutput` 을 내보낸다(허용). 상태 이름과 `systemMessage` 는 없다. | 상태 이름을 더하지 않는다(어휘 불변). 이 경로는 이 SPEC 의 잔여 위험으로 기록한다(acceptance.md §F). 관측: 읽기(765–769행, 802–812행)와 디코더 프로브. `json.Unmarshal` 은 빈 입력에 `unexpected end of JSON input` 을, 형식 오류 입력에 `invalid character 'n' looking for beginning of object key string` 을 낸다. 처리기를 빈 입력으로 실행한 관측은 없다. |
| Go 773–777행 (같은 함수) | 처리기 `HandleCodexReviewGate` 가 오류를 돌려주면 773행의 분기가 표준 오류에 한 줄을 쓰고 빈 `hook.HookOutput` 을 내보낸다(허용). 상태 이름과 `systemMessage` 는 없다. | 상태 이름을 더하지 않는다(어휘 불변). 잔여 위험으로 기록한다(acceptance.md §F). 처분의 확정은 리더 판정 대상이다. 관측: 읽기만 했다(773–777행). 처리기 오류를 일으키는 입력은 만들지 않았다. |

앞의 네 행과 83–86행의 래퍼 수정, 곧 래퍼가 상태를 명명한 메시지를 내도록 하는 일은 t1674 의 몫이다. 두 Go 행의 처분은 위 표가 정한다. 이 SPEC 은 래퍼를 수정하지 않는다.

---

## §C AC 와 검증 구속

- 모든 인수 기준은 acceptance.md 에 있다(AC-RG-001–005, AC-RG-014–015, AC-RG-017–019, AC-RG-025, 요구와 1:1 이상 연결, 고아 기준 없음. 제거된 ID 는 HISTORY v0.2.0 행과 v0.2.3 행의 두 줄, 그리고 v0.2.7·v0.2.8 행에 있다). 이 SPEC 은 plan 시점에 릴리스 게이트를 하나도 갖지 않는다. 이유: 행위 요구를 단언하는 테스트가 현재 트리에 없고(research.md §6, 테스트 이름 확인), 기존 행위 테스트는 결함을 고정할 뿐이다(acceptance.md §A). 릴리스 게이트로 승격하려면 run 단계가 구현 전 트리에서 RED 를 원문으로 기록해야 한다.
- 검증 명령은 항상 작업 트리 루트에서, 단일 호출로 실행한다. 실행된 테스트 수를 함께 기록하며, 0개는 측정 불가로 보고한다.

## §D 실행 순서 구속

1. 운영자 판정 Q2–Q6 과 보드 판정 D3·D4·D11 은 decision-index.md 에 기록되어 있다. Q1 은 t1662 로 이관되었다(d-20261010T070728Z-6cdf). run 단계는 Q2–Q6 행의 `Operator verdict` 가 채워져 있는지 확인하고 시작한다.
2. 각 마일스톤의 RED 선행은 plan.md 의 마일스톤 순서 조항이 정한 커밋 그래프 검사(첫 부모 트리에 덮어쓰기 단위를 얹어 EXPECTED_RED 재현. 덮어쓰기 단위는 마일스톤 커밋이 추가하거나 바꾼 _test.go 파일의 그 커밋 시점 내용이다)로 확인하고, 그 원문은 progress.md §E.2 에 남긴다. 덮어쓰기 단위의 테스트는 첫 부모의 비테스트 코드만으로 컴파일되어야 한다. 컴파일 실패는 tdd-result-contract.md 의 TOOL_FAILURE 이며 RED 로 세지 않는다.
3. **변경 세트(change set)는 커밋 하나다.** 두 면을 가진 파일은 같은 커밋에서 함께 갱신한다. 훅 래퍼(`.claude/hooks/moai/handle-codex-review-{entry,gate}.sh` 와 `internal/template/templates/.claude/hooks/moai/handle-codex-review-{entry,gate}.sh`)의 변경은 후속 카드 t1674 에 속하며 이 SPEC 에서는 하지 않는다. M7 의 규칙 문장 정정(F4)도 규칙 파일(`.claude/rules/moai/workflow/factory-dispatch-detail.md`)과 그 템플릿 미러를 한 커밋에서 맞춘다. 두 면이 서로 다른 커밋에 갈라지는 상태는 허용하지 않는다.

---

## §E 범위 밖 (Out of Scope)

이 SPEC은 (1)·(3)·(5)와, 아래 (6) 중 REQ-RG-007 부분만 다룬다. (2)·(4)는 t1662로 이관한다. (6)~(10)은 t1637(3.2-1-4)이 후속으로 맡는다. (6) 중 REQ-RG-007의 상태 메시지(D4)는 이 SPEC에 남고, 나머지 가시성 확장은 t1637이 맡는다.

- (6) 미리뷰 systemMessage 가시화 → REQ-RG-007 의 상태 메시지(D4)는 이 SPEC 에 남고, 나머지 가시성 확장은 t1637
- (7) 차단 사유 중복·경로·길이 상한 → t1637
- (8) 킥 중복 제거·동시 리뷰 상한·마커 누수 → t1637
- (9) 빈 diff 무킥 → t1637
- (10) 진입 훅 timeout 정합 → t1637
- (4) 영수증을 병합 기준·내용 다이제스트에 결속 → t1662

구 REQ-RG-005~008·011은 t1674가 이어받음.

| 구 ID | 신 ID |
|---|---|
| REQ-RG-001 | REQ-RG-001 |
| REQ-RG-002 | REQ-RG-002 |
| REQ-RG-003 | REQ-RG-003 |
| REQ-RG-004 | REQ-RG-004 |
| REQ-RG-009 | REQ-RG-005 |
| REQ-RG-010 | REQ-RG-006 |
| REQ-RG-012 | REQ-RG-007 |

### Out of Scope — 영수증 결합 기준의 변경 (F1)

- 영수증을 HEAD 대신 (merge base, 내용 digest) 에 묶는 변경은 이 SPEC 에서 하지 않는다. t1662 로 추적한다(카드 항목 4, 영수증을 병합 기준·내용 다이제스트에 결속).
- 현재 결합 필드 다섯 개의 구성 변경도 하지 않는다.
- codex 필수 게이트의 대상 (main...HEAD 범위에는 SPEC 문서만 들어 있다)과 plan-aware 영수증은 t1662 범위다 (운영자 판정 d-20261010T090503Z-8963 (4)(5)). 이 SPEC 의 감사는 두 항목을 차단으로 세지 않는다.

### Out of Scope — 가시성, 중복 제거, kick 정리 (F2)

- 차단 사유의 중복 제거와 길이 상한, 미검토 상태의 systemMessage 표시 확장(REQ-RG-007 의 범위를 넘는 부분), kick 의 pid·liveness 기반 중복 제거와 레인 전체 동시성 상한, 마커 누수 정리, 추적 diff 가 비면 kick 하지 않는 규칙, 진입 훅 timeout 의 5초 정렬은 이 SPEC 밖이다. Follow-up F2 (t1637) 로 추적한다.

### Out of Scope — factory 저장소 경로 (F3)

- `internal/factory/todo_dispatch_current.go:53` 과 `internal/factory/todo_runtime.go:364` 의 두 주장은 이 SPEC 의 diff 에 없다. 카드 t1620 밖의 별도 후보로 추적한다.

### Out of Scope — 감사 도구의 나머지 결함

- `audit_multi` 의 codex 레그가 target 을 무시하는 문제, claude 자식 프로세스 env 의 denylist 방식, codex 샌드박스 정책 미전달, 작업 수명 정리 누락, factory 메시지의 null 반환은 카드 항목 밖이다. 연구 입력 lens-3 의 별도 후보로 남긴다.

### Out of Scope — doctor 진단 신설 (F4)

- 두 리뷰 계층이 동시에 켜졌을 때 doctor 경고를 추가하는 일은 이 SPEC 에서 하지 않는다(운영자 판정 Q5 = (b)). Follow-up F4 로 등록한다.

### Out of Scope — 두 리뷰 계층 중 하나의 폐기

- 턴 종료 게이트나 card-review 단계 가운데 하나를 없애는 결정은 이 SPEC 이 하지 않는다. 중복 검토는 후보 결함으로만 기록한다(규칙 문장 F4 와 연결).

### Out of Scope — GLM 턴 종료 게이트 신설

- 현재 GLM 턴 종료 게이트는 존재하지 않는다(research.md §3). 신설은 범위 밖이다.

### Out of Scope — SPEC-GATE-BOTTLENECK-001 의 편집 (F5)

- 대상 SPEC 의 REQ-GBN-002, REQ-GBN-004 와 AC-GBN-004 의 inconclusive 절은 편집하지 않는다(운영자 판정 Q6 = (b)). 처분 기록은 이 SPEC 의 HISTORY 와 F5 에만 있다.

### Out of Scope — 워크플로 설정 파일

- `.moai/config/sections/workflow.yaml` 은 편집하지 않는다.

### Out of Scope — 문서 불일치의 t1635 이관 (F6)

- AGENTS.md:94–95 · :166–167 과 `.claude/rules/local/gitflow-lane-protocol.md:16` 의 충돌(카드 기준 브랜치가 main 인지 develop 인지)은 이 SPEC 에서 고치지 않는다. 기록만 남기고 카드 t1635(3.2-1-2)로 이관한다.

### Out of Scope — 기준 브랜치 선택 (REQ-ACE-005, t1662)

- 리뷰 기준 브랜치 선택(후보 집합, 활성 흐름·모드 인식, 키 불일치, 후보 없음 처리)은 이 SPEC 에서 하지 않는다. 후속 카드 t1662 가 맡는다(리더 판정 d-20261010T070728Z-6cdf).

---

## Follow-up

- **F1 — 영수증 결합 기준 변경.** 영수증을 HEAD 대신 (merge base, 내용 digest) 에 묶는다(카드 항목 4). t1662 로 추적한다(카드 항목 4, 영수증을 병합 기준·내용 다이제스트에 결속). 요구사항을 붙이지 않는다.
- **F2 — 가시성과 운영 정리.** 카드 항목 6–10 은 t1637(3.2-1-4)이 맡는다. 다만 항목 6 중 REQ-RG-007 의 상태 메시지(D4)는 이 SPEC 에 남는다. 카드 항목 6–10: 미검토 상태를 systemMessage 로 드러내기(REQ-RG-007 밖의 범위), 차단 사유의 중복 제거와 길이 상한, kick 을 pid·liveness 로 중복 제거하고 레인 전체 상한을 두며 마커 약 64개 정리, 추적 diff 가 비면 kick 하지 않기, 진입 훅 timeout 을 5초에서 15초 예산과 맞추기. 요구사항을 붙이지 않는다.
- **F3 — factory 후보 두 건(카드 t1620 밖).** (a) `internal/factory/todo_dispatch_current.go:53` 의 백로그 SQLite 열기가 JSON 큐 이관 전에 빈 SQLite 를 만들 수 있다. (b) `internal/factory/todo_runtime.go:364` 의 할당 트랜잭션에서 외부 저장소 변경 뒤 내부 커밋(약 368행)이 실패하면 복구되지 않는다. 두 주장의 검증 결과는 research.md §7 에 있다. 요구사항을 붙이지 않는다.
- **F4 — 두 계층 동시 활성 doctor 점검 (운영자 판정 Q5 = (b)).** doctor 검사는 이 SPEC 에서 추가하지 않고 후속 항목으로 등록한다. 규칙 문장 `.claude/rules/moai/workflow/factory-dispatch-detail.md:157`(두 계층이 같은 카드 diff 를 두 번 검토한다는 문장)은 Q5 (b) 의 구성 요소로 run 단계(M7)에서 정정한다. 정정은 그 규칙 파일과 템플릿 미러를 같은 커밋에서 맞춘다. 이 plan 단계에서는 그 규칙 파일을 편집하지 않는다.
- **F5 — Q6 기록 처분 (운영자 판정 Q6 = (b)).** 이 SPEC 은 SPEC-GATE-BOTTLENECK-001 의 REQ-GBN-002(`spec.md:43–47`, 진입 차단 절), REQ-GBN-004(`spec.md:51–54`)와 AC-GBN-004(`spec.md:67–68`, inconclusive 절)를 편집하지 않는다. 처분은 이 SPEC 의 HISTORY 와 이 항목에만 기록한다. 잔여 위험: 대상 SPEC 은 in-progress 이며, 코드의 진입 차단 동작과 REQ-GBN-002 가 주석 없이 어긋난 상태로 남는다.
- **F6 — 문서 불일치의 t1635 이관 (운영자 판정, 보드 d-20261010T063222Z-43e0).** AGENTS.md:94–95 · :166–167 은 카드 브랜치를 로컬 main 에서 자르고 develop 을 legacy 라고 적으며, `.claude/rules/local/gitflow-lane-protocol.md:16` 은 카드 워크트리를 develop 에서 만든다고 적는다. 이 SPEC 은 두 문서를 고치지 않고 기록만 남겨 카드 t1635(3.2-1-2)로 넘긴다. 요구사항을 붙이지 않는다.
- **F-3 — 기준 브랜치 선택 (활성 모드 인식 포함, 계획 감사 3회차 발견).** `manual.workflow` 를 활성 모드와 무관하게 읽는 결함이며 부채가 아니다. 후속 카드 t1662 로 이관했다(REQ-ACE-005, 리더 판정 d-20261010T070728Z-6cdf). 요구사항을 붙이지 않는다.

---

## §F 제약

- 이 SPEC 의 산출물은 `.moai/specs/SPEC-REVIEW-GATE-REPAIR-001/` 안에만 둔다.
- 작업 트리의 `.moai/config/sections/workflow.yaml` 은 건드리지 않는다.
- 기준 코드 SHA 는 `2aab5f797b75983e132af451da68f69e3426557b`(짧은 형식 `2aab5f797`)다. 설치된 빌드(`4f8aba061`)는 트리 기준의 후손 커밋이므로 트리의 증거로 인용하지 않는다.

## §G 참조

- research.md — 인용 재검증 등록부, 재현 근거, 전제 확인(§9: 수리 1회차 재검증)
- design.md — 상태 모델, 스위치 해석, 판정 표(기준 해석 함수는 t1662 로 이관)
- plan.md — 마일스톤과 기술 접근
- acceptance.md — 인수 기준 매트릭스와 기준선 셀
- decision-index.md — 운영자 판정 Q2–Q6 과 보드 판정 D3·D4·D11, 2회차 기록(보드 d-20261010T063222Z-43e0), Q1 의 t1662 이관(d-20261010T070728Z-6cdf)
- `.moai/specs/SPEC-CODEX-REVIEW-TARGET-001/spec.md` §A.7 · `.moai/specs/SPEC-GATE-BOTTLENECK-001/spec.md` REQ-GBN-002 · REQ-GBN-004 · AC-GBN-004
- `.moai/reports/t1620/plan-audit.md` §1.4 — 수리의 입력(감사 결함 D1–D12, 선택 항목 O1–O11). 이 SPEC 은 그 파일을 권한으로 인용하지 않는다.
