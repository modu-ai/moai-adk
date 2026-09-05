# SPEC Review Report: SPEC-LEAD-DEPUTY-001

- **SPEC-ID**: SPEC-LEAD-DEPUTY-001 (card t471)
- **Tier**: M (frontmatter `tier: M` — 판정 threshold 0.80)
- **측정 트리**: `.claude/worktrees/t471` @ `3ec569871900bbd6c0f3889aec0f29a660193ffb` (branch `WT-lead-bottleneck`, develop 흡수 병합 커밋)
- **날짜**: 2026-09-03
- **감사자**: plan-auditor (독립 감사 — M1 Context Isolation 적용, 작성자 추론 맥락 없이 산출물만 감사)
- **반복**: 1/2 (Tier M 상한)

## Verdict

**PASS-WITH-DEBT** — 점수 **0.925** (threshold 0.80 이상)

- skip-eligible **아님**: skip 정책은 정확히 `PASS`인 verdict만 인정한다(spec-workflow.md § Phase 1 Plan Audit Gate skip 조건 1). PASS-WITH-DEBT는 run-gate Phase 1 재실행 대상이며, 재감사는 Retry Loop Contract에 따라 아래 D1·D2 결함 델타로 범위 한정된다.
- 부채 내용: D1·D2(차단 클래스, 경미 심각도) — acceptance.md 계수 명령 2건의 계측 정밀도 결함. 요구사항·계획·권한 경계·측정가능성 자체는 모두 건전하다.

## Lead 지정 렌즈

### Lens 1 — deputy 권한 경계는 문서 앵커, 확대 금지 → **PASS**

이 트리의 정본 경계를 직접 읽고 SPEC 전체와 대조했다:

- 기준 문서 실재 확인: `.claude/agents/moai/manager-lead.md:198` § Deputy dispatch surface (Role B extension), 그 안의 위임 5종 표(:202-210), `DEPUTY-RETAINED-BY-LEAD` 6항목(:212-221), delivery-shape 검증 절(:223-225). `.claude/rules/moai/workflow/kanban-dispatch.md:96` § Deputy dispatch surface, `kanban-dispatch-detail.md:132` § Deputy mode.
- **REQ-LDP-004의 6항목 나열은 교리와 의미가 정확히 일치**한다: final merge approval / `FINAL VERDICT:` 토큰 금지 / operator gates / queue mutations(`moai todo`) / CodeRabbit 규율 판정 / cross-session dispute coordination — manager-lead.md:216-221의 6개 retained 항목과 1:1 대응. SPEC이 인용하는 두 토큰(`DEPUTY-RETAINED-BY-LEAD`, `FINAL VERDICT:`)은 **실제 교리에 존재**(grep 실측, 아래 Evidence) — SPEC이 경계 토큰을 창조하거나 왜곡하지 않았다.
- **확대 없음의 구조적 근거**: spec.md §4 "리드 보유 6종은 한 줄도 늘어나거나 줄어들지 않는다"(매트릭스 불변 선언) + §5 PRESERVE(manager-lead.md·kanban-dispatch.md 기존 [HARD] 절 전부 원문 보존) + §6 Out of Scope — 판정·게이트 권한 이관 명시적 배제 + §1.5 사라질 수 없는 것 3경계 못박기 + AC-LDP-010(편집 전후 retained 6항목 의미 변화 0건 검증).
- **위임 범위 검증**: REQ-LDP-002(raw 트리 판독→`RECOMMEND:` 요약)·REQ-LDP-005(측정 배치·표 초안)는 manager-lead.md:202-210의 위임 가능 5종 중 "First-pass evidence reading + recommendation"과 "Summary reporting"의 범위 안이다. 리드 귀속 의무는 REQ-LDP-005(리드 단언 수치=리드 귀속)와 AC-LDP-006(deputy 귀속 표기, 무귀속 0건)으로 `verification-claim-integrity.md` §2를 보존한다. 판정 소재는 REQ-LDP-004 + AC-LDP-004(`FINAL VERDICT:` 부재)로 리드에 고정.
- 경미 관찰 1건: D8(REQ-LDP-002 문언이 "리드의 증거 읽기 의무 소멸"로 오독될 여지 — §1.5.3 "읽는 횟수가 아니다"와 §5가 이미 상쇄하나, M1 추가 문단에 명시 1문장 권고). 권한 확대가 아닌 문안 리스크다.

**렌즈 1 판정: 경계 확대 0건, 문서 앵커 유지. PASS.**

### Lens 2 — 병목 감소 주장의 비공허성(측정가능성) → **PASS**

- **계수 기준 존재**: REQ-LDP-009 + AC-LDP-001 — 리드 턴 툴 배치 수(위임 3계열 {SendMessage 디스패치 발송, raw 트리 증거 판독, 보고 측정·초안 저작} 귀속분), 세션 로그 집계 레시피 명시.
- **비교 정의**: AC-LDP-001 Then — §D.0 baseline의 **50% 이하**. baseline은 acceptance.md §D.0 표로 존재(13건 인바운드 + 12건 아웃바운드 + 4회차 × 툴 배치 5-6회; 배치 수의 상한 근사임을 스스로 명기).
- **누가·언제·어디에**: WHERE = progress.md §E.2(측정 명령·원문 출력·창 구성 귀속) ✓, WHEN = 고정 관측 창(활성 레인 ≥8, 배치 1일, RED-now 창 구성 동시 기록; §D.2에 따라 첫 채택 배치 회차 보고에서 실측) ✓, WHO = **미명시**(implicit — 진단 없이는 판정 가능하나 D4로 기록, optional).
- **판정가능성**: 채택 후 감사자가 같은 레시피를 같은 창에 재실행해 PASS/FAIL에 도달할 수 있다. "빨라질 것이다"로 귀결되지 않는다. 교락 요인(턴 수=운영자 지시 빈도)을 명시적으로 배제(REQ-LDP-009, §6 Out of Scope — 리드 턴 수 감소)한 것도 계측 성숙도의 근거.
- **baseline 귀속 정직성**: §D.0 전 수치에 `[리드 자체 계수]` 귀속을 명기하고 "본 워크트리 HEAD가 아니므로 트리 SHA 대신 세션 좌표로 귀속"한다고 선언 — 귀속을 사칭하지 않았고, 재측정 레시피 소유자(AC-LDP-001)를 지정했다. 공허한 기준값이 아니다.

**렌즈 2 판정: 계수 기준·측정법·비교·기록면 모두 정의됨. PASS (D4는 optional 관찰).**

## Must-Pass 결과

| 항목 | 판정 | 근거 |
|---|---|---|
| MP-1 REQ 번호 일관성 | **PASS** | REQ-LDP-001~011 연속, 공백·중복 0건 (spec.md §3.1-§3.5 전수 확인). 11 ≤ Tier M 상한 16 |
| MP-2 GEARS 형식 (요구사항 계층) | **PASS** | 11건 전부 GEARS 5패턴 구조 적합: 001 Event-driven / 002 State-driven / 003 Event-driven 구조(라벨 표기만 비표준 — D9) / 004 Unwanted(shall not) / 005 State-driven(피동 주어 — D9) / 006·009·011 Ubiquitous / 007 Event-driven / 008 Unwanted / 010 Where(capability gate). 판정 계층: **요구사항 계층(spec.md §3)만** — AC의 Given-When-Then은 검증 계층 정규 형식으로 이 기준의 대상 아님(M3 § Scope) |
| MP-3 YAML frontmatter 유효성 | **PASS** | spec.md 12 필드 전부 존재·타입 적합(id/title/version "0.1.0"/status draft/created·updated 2026-09-03/author/priority P1/phase "v3.1.5 target"/module/lifecycle spec-anchored/tags CSV). snake_case 별칭 0건. plan.md·acceptance.md는 `status:` 무상태 규정 준수(해당 필드 부재) |
| MP-4 §22 언어 중립성 | **N/A** | 단일 언어(교리 문서 편집) 스코프 SPEC — 다중 언어 도구 다루지 않음. 자동 통과 |
| MP-5 D7 교차-SPEC 조정 | **PASS** | 참조 SPEC 2건 실재 + status 확인: SPEC-LEAD-DEBOTTLENECK-001 `completed`(=depends_on 이행), SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001 `completed`. retired/superseded/archived 참조 0건 → 조정 의무 미발생, BLOCKING 없음 |
| MP-6 D8 크로스플랫폼 규율 | **PASS (auto)** | 3개 산출물 모두 `syscall` 0매치 (grep 실측) → D8-4 자동 통과 |
| MP-7 clarification 게이트 | **PASS (주석)** | research.md 부재(Tier M 불요). plan.md의 `grep -n 'NEEDS CLARIFICATION'` = **1매치이나 negated 형태** `[NEEDS CLARIFICATION 없음]`(plan.md:75 — 콜론+토픽 없는 해결 선언문). 마커 관례 `[NEEDS CLARIFICATION: <topic>]`의 미해결 마커가 아니므로 PASS. 다만 나이브 게이트 grep이 걸릴 수 있어 D7(optional)로 기록 |

## 차원 점수

| 차원 | 점수 | 밴드 | 근거 |
|---|---|---|---|
| Clarity | 0.95 | 1.0 근접 | 요구사항 단일 해석 가능, 용어 정의(§2 상주 deputy), 귀속 규약 명시. D8·D9 수준의 문안 리스크만 존재 |
| Completeness | 0.95 | 1.0 근접 | HISTORY/§1 문제/§2 용어/§3 요구/§5 제약/§6 Out of Scope(H3 5개+불릿)/§7 교차참조 전부 존재. §D.0 baseline + AC 10건. D3(§D.1↔§D.2 정렬)·D5(deputy 생존 규율) 경미 공백 |
| Testability | 0.85 | 0.75-1.0 사이 | AC 8/10이 날카롭게 이진 판정형. 단, AC-LDP-002의 명시된 계수 명령이 약한 판별자+거짓 RED-now(D1), AC-LDP-007의 `git diff --stat`가 실제 대상 파일(비추적)에 공허 통과 위험(D2), AC-LDP-010의 "의미 변화 0건"은 판단 개입 소폭 필요 |
| Traceability | 0.95 | 1.0 근접 | REQ 11건 ↔ AC 10건 쌍방향 전수 대응(001→003, 002→004, 003→005, 004→010, 005→006, 006→007, 007/008→002, 009→001, 010→008/009, 011→009). 고아 REQ/AC 0건. plan §F 마일스톤 4개(M1-M4, Tier M 범위 3-6 내)가 REQ 인용 + AC에 [M1]-[M4] 태그 |
| **종합** | **0.925** | | 산술 평균 |

## 결함 목록

**D1** — `acceptance.md` AC-LDP-002 — 심각도: minor — 클래스: **blocking**
계수 명령 `grep -n "idle"`(kanban-dispatch.md + manager-lead.md)은 오늘 이미 kanban-dispatch.md:69에서 1매치(해당 줄이 idle-통지 경계 문언을 이미 담고 있음 — `notify_when_idle` + "the notice is not the completion signal")이므로, RED-now 서술 "적중 0"은 관측상 거짓이고 같은 명령은 구현 전후를 갈라놓지 못하는 약한 판별자다(공허 초록 위험).
**최소 수정**: 기계 검증을 오늘 실측 0인 정밀 판별자로 교체 — 예: `grep -c 'scheduling hint' .claude/rules/moai/workflow/kanban-dispatch.md .claude/agents/moai/manager-lead.md` (본 트리 3ec569871 실측 0·0) — 그리고 RED-now 줄을 그 값에 맞게 정정. M3가 cross-session-messaging.md § An idle notice is a scheduling hint(:91) 상호참조를 새로 넣을 때 뒤집힌다.

**D2** — `acceptance.md` AC-LDP-007 — 심각도: minor — 클래스: **blocking**
`git diff --stat`로 "회차당 변경 파일 ≤2"를 확인하는 명령은, 현재 관행상 회차 보고 파일이 비추적이면 공허 통과한다 — 실측: `git ls-files .moai/reports/lead/` = 0건(현 단일 대형 보고 `.moai/reports/lead/card-status-*.md` 미추적). 비추적 파일은 `git diff --stat`에 아무것도 나타나지 않아 규율 준수 여부와 무관하게 0 ≤ 2로 통과한다.
**최소 수정**: 회차 파일+인덱스를 추적 경로에 못박거나, 검증을 파일 목록 기반으로 재서술(예: 회차 디렉터리 `ls` 카운트 + 인덱스 1건 — git 의존 제거).

**D3** — `acceptance.md` §D.1/§D.2 — 심각도: minor — 클래스: optional
AC-LDP-003(release-blocking 분류)은 채택 후 관찰형(세션 로그)인데 §D.2의 2단 검증 목록(AC-LDP-001, 005/006/007)에 없어 run-phase 검증 형태가 미지정이다. 수정: AC-LDP-003을 §D.2에 추가하거나 protocol-readiness 형태를 명기.

**D4** — `acceptance.md` AC-LDP-001 — 심각도: minor — 클래스: optional
채택 후 측정의 수행 주체(WHO)가 미명시(WHERE=§E.2, WHEN=창 정의는 존재). 수정: Then 절에 측정자 1절 추가(리드 또는 지정 감사자).

**D5** — `spec.md` §2/REQ-LDP-001 — 심각도: minor — 클래스: optional
배치 중 deputy 사망 시 재기동 규율 미정의(배치 시작 시 정확히 1개만 못박음). 수정: M1 상주 절에 재기동/열화 허용 중 무엇인지 1문장.

**D6** — `plan.md` §A / `progress.md` §E.1 / `plan-evidence.md` — 심각도: minor — 클래스: optional
verdict 파일 예정 위치가 3개 표면에서 `.moai/reports/t471/verdict.md`로 예고됐으나 실제 감사 산출물은 `.moai/reports/t471/plan-audit.md`에 기록됨(리드 지시 경로). run-phase 진입 시 progress.md §E.1 정합화 권고.

**D7** — `plan.md:75` — 심각도: minor — 클래스: optional
negated 리터럴 `[NEEDS CLARIFICATION 없음]`이 나이브 게이트 grep(`\[NEEDS CLARIFICATION`)에 매치됨(실측 1매치). 관례상 미해결 마커가 아니어서 MP-7은 PASS지만, `NEEDS CLARIFICATION 항목: 없음` 식으로 재표기 권고.

**D8** — `spec.md` REQ-LDP-002 — 심각도: minor — 클래스: optional
"리드 턴은 요약을 받고 raw 판독 배치는 받지 않는다"는 문언이 `kanban-dispatch-detail.md` § Deputy mode의 "deputy의 1차 판독은 리드 자신의 증거 읽기 의무를 해소하지 않는다"와 충돌하는 것으로 오독될 여지. §1.5.3("줄어드는 것은 턴 점유지, 읽는 횟수가 아니다")과 §5가 이미 상쇄하므로 경계 확대는 아님. 수정: M1 추가 문단에 리드 읽기 의무 존속 1문장.

**D9** — `spec.md` REQ-LDP-003/005 — 심각도: minor — 클래스: optional
GEARS 표기 세부: REQ-LDP-003 라벨 "(Event-detected)"는 5패턴 카탈로그에 없는 비표준 라벨(문장 구조는 Event-driven 적합 — MP-2에는 영향 없음). REQ-LDP-005는 피동 주어("shall be delegated")로 시작. 문안 정리 권고.

## Evidence (측정 명령 원문 — 본 트리 3ec569871에서 실행)

| # | 확인 대상 | 명령 | 관측 결과 |
|---|---|---|---|
| E1 | 트리 SHA | `git rev-parse HEAD` | `3ec569871900bbd6c0f3889aec0f29a660193ffb` |
| E2 | 참조 SPEC 실재+상태 | `grep -H '^status:' .moai/specs/SPEC-LEAD-DEBOTTLENECK-001/spec.md` (+ TEAMMATE-REVIVAL) | 양쪽 `status: completed` |
| E3 | `DEPUTY-RETAINED-BY-LEAD` 토큰 | `grep -rn 'DEPUTY-RETAINED-BY-LEAD' .claude/` | manager-lead.md:212,214 (+detail:152, 템플릿 쌍) |
| E4 | `FINAL VERDICT:` 금지 토큰 | `grep -rn 'FINAL VERDICT' .claude/` | manager-lead.md:217 (deputy 출력 금지 조항) |
| E5 | 위임 5종/보유 6종 표 | Read manager-lead.md:196-232 | :202-210 위임 5종, :212-221 retained 6항목, :223-225 delivery-shape |
| E6 | D8 syscall | `grep -c 'syscall' spec.md plan.md acceptance.md` | 0 / 0 / 0 |
| E7 | MP-7 마커 | `grep -n 'NEEDS CLARIFICATION' plan.md` | :75 negated 1매치; research.md 부재 |
| E8 | AC-LDP-002 RED-now 검증 | `grep -c 'idle' kanban-dispatch.md manager-lead.md` | **1** / 0 — RED-now "적중 0" 거짓 확정 (kanban-dispatch.md:69에 경계 문언 기존) |
| E9 | AC-LDP-002 정밀 판별자 | `grep -c 'scheduling hint'` (두 파일) | 0 / 0 — 수정안의 판별자는 오늘 진짜 RED |
| E10 | AC-LDP-007 대상 추적 상태 | `git ls-files .moai/reports/lead/ \| wc -l` / `git check-ignore` | 0건 추적 / 무시 아님 — `git diff --stat` 공허 위험 확정 |
| E11 | AC-LDP-009 셀렉터 실재 | `grep -n 'func TestManagerLead...' internal/template/manager_lead_depth_test.go` | :40, :103 — 2매치, 빈 셀렉터 아님 |
| E12 | AC-LDP-008 중립성 baseline | `grep -rn "t471\|리드 자체 계수\|SPEC-LEAD-DEPUTY\|SPEC-LEAD-DEBOTTLENECK" internal/template/templates/ \| wc -l` | **0** — 보존형 기준선 성립 |
| E13 | 교차참조 앵커 | grep/Read | cross-session-messaging.md:91 § An idle notice is a scheduling hint / manager-lead.md:10 tools(SendMessage, ListAgents 포함) / template-internal-isolation-doctrine.md §25.1(:9) / kanban-dispatch-detail.md:132 § Deputy mode |
| E14 | SPEC 산출물 실재 | `ls .moai/specs/SPEC-LEAD-DEPUTY-001/` | 4파일 (spec/plan/acceptance/progress) |

## Gaps (관측하지 않은 것)

- `moai spec lint` 재실행 안 함 — plan-evidence.md의 자가 린트 통과 주장(exit 0)은 레인의 것으로 인용만 하고 독립 재실행하지 않았다(MP-3은 수동 필드별 대조로 별도 검증함).
- AC-LDP-001의 post-adoption 실측(50% 감소)은 채택 이후에만 가능 — 본 감사는 §D.2의 protocol-readiness 차원만 판정했다.
- `[리드 자체 계수]` baseline 수치 자체(13건/12건/4회차/5,550줄)는 리드 세션 좌표의 자체 계수로, 본 감사 트리에서 재유도 불가 — SPEC이 명시한 귀속대로 받아들였다(공허성은 판정하지 않음, 재측정 레시피 소유를 AC-LDP-001로 확인).
- 템플릿 `.codex/agents/moai/manager-lead.toml`(C3)의 내용 정합성은 `make agents-emit` 재생성 후에야 검증 가능 — run-phase 소관.

## Residual-risk

- D1·D2 미수정 시: AC-LDP-002는 구현 전에도 초록으로 읽힐 수 있고, AC-LDP-007은 비추적 대상에서 규율 위반을 관측 못 한다 — run-gate 재감사(델타 범위)에서 두 수정의 반영을 반드시 확인할 것.
- deputy 상주 모드의 실제 효과(AC-LDP-001의 50% 목표)는 첫 채택 배치에서 나오는 운영 데이터에 달려 있다 — protocol-readiness PASS는 효과 보증이 아니다(§D.2가 인정한 지연).
- D5(재기동 규율 부재)는 첫 배치에서 deputy 사망 시 리드가 즉석 판단해야 하는 빈칸으로 남는다.
