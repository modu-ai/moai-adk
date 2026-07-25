---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization"
version: "0.10.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: ".claude/agents/moai, .claude/skills/moai/workflows, .claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md, .claude/rules/moai/workflow, .claude/workflows, internal/template/templates, internal/template/internal_content_leak_test.go"
lifecycle: spec-anchored
tags: "agent-diet, parallelization, fan-out, write-concurrency, workflow-wiring, template-first"
tier: L
related_specs: [SPEC-DWF-CODEMAPS-PILOT-001, SPEC-WORKFLOW-CACHE-OPT-001, SPEC-V3R6-WORKFLOW-EFFORT-MAP-001, SPEC-SUBAGENT-NESTING-DOCTRINE-001]
partially_supersedes: [SPEC-DWF-CODEMAPS-PILOT-001]
---

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-25 | manager-spec | 최초 plan-phase draft. §F Ground Truth 전량 본 세션 실측(agent line count 10파일, orphan grep, template mirror 8경로, write-concurrency 3표면, run/sync/plan phase 번호). 브리프 2건 실측 정정(§F.7). Tier L(§E). |
| 0.10.0 | 2026-07-25 | manager-spec | **§D.2 라인 상한 M3 실측 재보정 — 수치를 현실에 맞춤**(사용자 승인). `plan-auditor.md` 상한 **340 → 430**. M3가 505→454(−51) 후 잔여 114행을 "에이전트 고유 소유"로 보고했고, **본 세션 독립 판독 결과 그 보고는 대체로 옳으나 한 부류를 놓쳤다** — M3의 렌즈는 *파일 간* SSOT 중복이었고 잔여는 *파일 내* 동일 규칙 반복(§A 축 1의 (c))이다. Group 7/8이 각 차원을 5중 진술(요약 불릿·유래 산문·번호 체크·bash 검증 동사·Severity rubric)하며 이 중 재진술 3종은 조작적 내용 없이 제거 가능하다. 직접 판독으로 제거 가능 **28행**을 확인해 하한 426을 도출하고 여유 4행을 두어 430 확정(§D.2.1에 구간별 산출표 기록). **`manager-spec.md` 230 유지** — 실측 236의 초과 6행은 § Status Responsibility Matrix의 1행 표가 § SPEC Artifact Ownership "Status transitions owned"를 재진술하는 구간이라 상쇄 가능하므로, 실측이 상한을 넘었다는 사실만으로 상한을 올리지 않는다(사소하게 충족되는 상한은 아무것도 측정하지 않는다). `manager-develop.md` 240 유지(실측 238 충족). **합계 ≤1,907/1,927 → ≤1,997/2,017**, 감축률 주장 **20% → 16%**(분기 A 17.3% / B 16.5%) — 감축률은 목표가 아니라 **파생값**이므로 합계에 맞춰 재계산했다. **공백 전용 압축 금지** [HARD] 신설 — 빈 줄을 접어 상한을 맞추면(약 19행 확보 가능) 지표가 아무것도 측정하지 않는다. §D.2.1에 향후 재보정 4단계 절차 코드화: 나머지 7파일 상한은 여전히 **미검증 추정치**이며 `plan-auditor`와 동일한 낙관 편향 위험을 안는다. **§F.1 baseline 2,417은 불변** — `builder-harness.md` 195가 plan-phase 측정 오류라는 M3 보고는 **실행으로 반증됨**: `origin/main`과 plan 커밋 `e3a3eb4e4` 양쪽 모두 195이고, live 222는 본 SPEC이 편집하지 않는 파일에 대한 병렬 세션의 미커밋 편집이다(10파일 합계 `origin/main` 실측 2,417 재확인). M3 3파일의 병렬 세션 오염도 `effort:` 1행 치환(+1/−1, 라인 중립)뿐임을 diff로 확인. `plan.md` M3에 작업 1b(잔여 재진술 제거) 추가, `acceptance.md` AC-APO-055 합계·공백압축 판정 갱신. |
| 0.9.0 | 2026-07-25 | manager-spec | **가드 재구현 결함 해소 — AC-APO-061/071을 "가드가 권위" 형태로 전환**(코디네이터 blocker 해소 지시). 코디네이터가 `TestTemplateNoInternalContentLeak` PASS를 실증하며 v0.8.0이 blocker로 올린 2 토큰이 leak이 아니라 **`pedagogicalAllowlist` 등재 항목**임을 확정했다. 근본 결함은 파이프 버그가 아니라 **설계**였다 — AC-APO-061/071이 가드의 C1/C2/S2 정규식을 **면제 목록 없이 재구현**하면서 스코프는 더 좁았고(변경 파일 ⊂ 전체 트리), 이 조합은 거짓 실패만 생산할 수 있고 가드가 놓친 것은 잡지 못한다. M3 대상이 `manager-spec.md`이므로 거짓 실패는 곧 발생할 예정이었다. **AC-APO-061**: (a) 가드 실행을 권위 판정으로, (b) 보조 스캔은 본 SPEC 고유 토큰 계열로만 좁히고 **변경 라인**(`git diff -U0`) 스코프로 전환(실측: 파일 스코프 2 → 라인 스코프 0). **AC-APO-071**: 손수 쓴 정규식 전면 폐기 — M1이 `leakTextExtensions`에 `.js`를 등재해 가드에 완전 포섭되며, 게다가 S2의 `requireHexLetter` 정련 누락으로 십진 상수를 오탐하는 형태였다(`10485760` 실증). **AC-APO-071b는 존치** — 클래스 단위 실측 대조 결과 (α) 날짜는 `S1`이 있으나 strict 티어가 `MOAI_TEMPLATE_LEAK_STRICT=1` opt-in이고 `.github/`·`Makefile` 어디에도 미설정이라 CI 미강제, (β) `/Users/`는 대응 클래스 부재, (γ) sha `{9,40}`은 S2의 `{7,8}`과 비중첩 → 가드가 CI에서 강제하지 않는 잔여 구간의 유일한 커버리지로 확인. 보조 스캔의 additive 근거도 실측 확정 — `SPEC-AGENT-PARALLEL-OPT-001`은 **어떤 가드 클래스에도 미매치**, `REQ-APO-`를 잡는 `S3`는 `skillBodyScoped`라 본 SPEC 대상인 agents/workflows에서 미발화. **AC-APO-050은 중립성 가드와 무관**함을 확인해 셀에 명시(실측 8건 = 라이브 `manager-docs.md`의 `Nextra`×5 + `WCAG`×3, M3 본문 다이어트 대상이며 가드는 `internal/template/templates/`만 스캔) — allowlist 인지 처리 불요, M3 실행 시 해소. §A.1에 규칙 4 신설(가드 정규식을 면제 없이 재구현 금지, 보조 스캔은 조기 경보로 표기 + 라인 스코프)하고 기존 이식성 규칙을 5번으로 이동. 요구사항 의미 불변. `acceptance.md`도 0.9.0 동반 상향. |
| 0.8.0 | 2026-07-25 | manager-spec | **M2 run-phase 발견 AC 결함 3건 수정 + 동일 버그 클래스 스윕**(코디네이터 재위임, `acceptance.md` 판정 명령 한정). **D1 — AC-APO-024b conjunct 2 이중 공허**: (i) 구 패턴이 **한국어**를 찾았으나 대상 `sync/doc-execution.md`는 CLAUDE.md §9에 따라 **영어 전용**이라 매치 자체가 불가능했다(금지 영어 문장 7종 심은 fixture에서 구 명령 0 반환 실측). (ii) `grep -E`의 `\|`는 교대가 아니라 **리터럴 파이프**라 구 패턴은 두 절이 연접해야 하는 단일 패턴이었다(한국어 금지 문장에도 0, 정상 교대형은 1 — 독립 확인). 영어 기반 4항 판정(`CMD-024b`)으로 교체하고 잔여 취약점(마커 어휘 밖 표현)을 AC 셀에 공개. **D2 — 존재형 AC 2건 파일 앵커화**: AC-APO-027을 3개 표면(run Phase 9/18, sync Phase 9) **파일별** 판정으로, AC-APO-022를 `run/task-decomposition.md` 앵커로 교체(재귀형 `==N`도 한 파일 몰림 시 통과하므로 파일별만 유효). 등급은 SHOULD 유지. **D3 — 구조 AC 6건 판정 명령 신설**(020/021/024/025/028/030): 소유자명을 포함한 verdict 앵커, drafter 표 행 **구조 카운트**(==5), `disjoint` ==0(사용자 결정 D2의 유일한 기계 가드), 부정어 지배를 고려한 nesting 판정 등. 028(b)/030(b)의 "모든 사이트" 전칭은 **파일별 카운트까지만 기계화**했음을 AC 셀에 명시(과신 방지). **동일 버그 클래스 스윕(3건 외 추가 발견)**: `-E` + `\|` 조합이 AC-APO-023(부당 FAIL — 정상 구현을 0으로 판정), 050(**공허 GREEN**, 정상 교대형 실측 8), 061(**공허 GREEN**, 실측 2 — 단 2건 모두 `origin/main` 선행 부채), 071/071b(잠복, 실측 0)에도 존재. 전량 `-e` 반복 또는 코드블록으로 교체. 교대 명령용 코드블록 §D.2.1 / §D.5.1 신설, 재발 방지 규칙 §A.1(4개) 추가. 모든 신규·수정 명령은 RED/GREEN 왕복 + BSD grep/ugrep 이식성 실측 완료(`\b` 다중 사용 시 ugrep이 조용히 0을 반환하는 사례 발견 → POSIX 문자클래스로 회피). 요구사항 의미 불변(판정 정확화). `acceptance.md`도 0.8.0 동반 상향. |
| 0.7.0 | 2026-07-25 | manager-spec | **약한-AC 스윕**(코디네이터 승인). `acceptance.md`의 `≥ 1` 임계값 7곳을 전수 점검해 부분 커버리지로 통과하는 3건을 파일별 판정으로 강화. **AC-APO-014가 최악 사례** — `harness-builder.md:81`이 REQ-APO-014과 무관한 선례 인용으로 `codemaps-extract`를 포함하므로, 재귀 `≥1`은 요구 표면 `codemaps.md` 배선을 삭제해도 그 무관 인용으로 GREEN 유지(공허 실측 확인). **AC-APO-012**는 2개 요구 표면(`sync.md` + `run/task-decomposition.md`)에 재귀 `≥1`이라 한쪽만으로 통과 — 재귀형 `== 2`도 두 매치가 한 파일에 몰리면 통과하므로 파일별 판정으로 교체. **AC-APO-010**은 현재 트리에서 유일 매치라 오늘은 정상 FAIL하지만 재귀형이 표면 앵커를 갖지 않아 향후 이설 시 공허해지는 해저드 — 예방적으로 `plan.md` 앵커. 3건 모두 RED/GREEN 왕복 실측(배선 삭제 시 강화형 0=FAIL, 동일 트리에서 구 재귀형은 1=PASS 잔존). **AC-APO-015는 의도적 느슨함으로 존치** — 목적이 고아 스크립트 탐지이지 표면 커버리지가 아니며 스크립트당 `≥1`이 이미 per-item 판정, 그 사유를 AC 셀에 명문화해 향후 반사적 강화를 차단. 존치 판정 근거: AC-041/045/072는 단일 명명 파일 대상이라 부분 커버리지 불가, AC-043의 `≥1`은 분기 선택자이지 통과 임계값이 아님. 요구사항 의미 불변(판정 강화, 스코프 변경 아님). `acceptance.md`도 0.7.0으로 동반 상향(HISTORY 표는 `spec.md` 단독 보유). |
| 0.6.0 | 2026-07-25 | manager-spec | M1 후속 정밀화 3건(코디네이터 재위임). (1) REQ-APO-070 AC-APO-070(b) 충족 — supersede 대상을 **ID로 명명**(`AC-DCP-010`, 앵커 `acceptance.md:79` / `progress.md:86`, 소유 요구사항 `REQ-DCP-009` / `REQ-DCP-010`). 종전 문구는 "비배포 acceptance criterion"이라는 산문 서술만 있어 ID 인용 0건이었다. (2) AC-APO-070(c) 충족 — 인용 grep을 축약형 `grep -r "codemaps-extract" …`에서 정식 문구 `grep -r "codemaps-extract\|codemaps-pilot" …`로 복원(`\|codemaps-pilot` 교대 누락 시 엄격 판정 FAIL). §H 교차참조(선행 SPEC 항목)에도 동일 ID·문구 반영. (3) REQ-APO-012 정밀화 — Phase 13/16/17의 소유자를 진입 라우터 `run.md`가 아닌 하위 스킬 `run/task-decomposition.md`로 정정. `run.md`는 `TestEntryRouterLOCCeiling`(`internal/skills`)이 강제하는 200 LOC 상한 아래의 얇은 라우터(실측 197)로 배선 수용 불가이며, 실제 Phase 정의(L108/178/217)와 4dim 배선(L104)은 하위 스킬에 있다. M1이 라우터에 선배치했다가 203 LOC로 가드를 깨고 하위 스킬로 이설한 사실과 정합. `sync.md` Phase 7 배선은 진입 라우터 본문(`sync.md:56`)에 실재하므로 무변경. 요구사항 의미 불변(정밀화, 스코프 변경 아님). AC-APO-012 판정 명령은 `.claude/skills/moai/workflows/` 재귀 grep이라 경로 정정 불요(하위 스킬 포함). |
| 0.5.0 | 2026-07-25 | manager-spec | plan-audit iter-3 0.847(임계 0.85 대비 -0.003) → 마감 편집 F1-F4. F1: AC-049(c)의 `M` 마커를 정의 heading(`^## `)으로 앵커링 — 앵커 없는 형태는 :49 산문 cross-reference까지 세어 "계층형 유지" 분기를 부당 FAIL시켰고 기재 baseline(합 2)도 실제(합 3)와 불일치했다. F2: AC-024b가 v0.2.0 상태로 잔존(구-넘버링 4번째 사례) → REQ-024b 문구로 재작성. F3: REQ-071의 5개 금지 클래스 중 3개가 SHOULD로 새던 구조적 누수 → AC-071b에 `== 0` 판정 추가. F4: 미한정 "Group 1" 2곳에 `구` 한정어 적용. 사전 제출 검사에 **REQ-touched ∖ AC-changed 집합차** 절차 추가. |
| 0.4.0 | 2026-07-25 | manager-spec | plan-audit iter-2 FAIL 0.77 → **결정의 binding surface 미전파** 결함 해소. iter-1에서 서사(narrative)만 고치고 MUST AC 셀·마일스톤 실행 지시를 v0.2.0 상태로 남긴 3건이 근본 원인 — N1(AC-043 + M3 작업 0 게이트 명령), N2(AC-071 정규식), N3(AC-055 합계 상한). 부가: N4 구-넘버링 잔재 4곳(§E Tier 근거 포함), N5 AC-049 이진 판정화, N6 `module:` 전체 경로, N7 AC-076 다중파일 `grep -c` 오용, N8 AC-074 과잉 `== 0`, N9 AC-023 Phase 귀속 정정. AC 총수 55 → 56(AC-071b 신설: manual-only 클래스 표기 의무 이행). |
| 0.3.0 | 2026-07-25 | manager-spec | plan-audit FAIL 0.69 → 결함 해소. **구 Group 1(write-concurrency, REQ-001..005) 전면 철회** — REQ-024b가 이미 Phase 12를 Group 1에서 분리했고 §C가 disjoint-writer를 제외하므로 in-scope 수혜자가 0인 채 `[HARD]` 가드만 넓히는 구조였다(사용자 승인). 후속 SPEC으로 이관. Group 6 신설(REQ-074..078): shipped rule 모순 해소(D2)·SSOT 방향(D12)·e2e 목적지 명명(D11). D4 게이트 grep 판별형 교체(12,133→0), D5 공허 AC 실문구 교정, D6 `design.md` 신규 저작, D7 C1 정규식 사실오류 정정, D8/D9/D10 AC 실행가능화. |
| 0.2.0 | 2026-07-25 | manager-spec | 사용자 결정 D1/D2/D3 반영. D1=3개 fan-out 스크립트 **템플릿 배포**(REQ-069~073 신설, DWF-CODEMAPS-PILOT-001 비배포 AC 부분 supersede). D2=sync P12를 drafter+단일 적용자로 **확정**하고 M1 의존 해제. D3=SPEC-ID 마커는 **선행 grep 게이트** 후 결정. 배포 전제 3건 실측 정정(§F.8) — split_namespace 가드는 이미 prefix-scoped라 개정 불요, leak 가드는 `.js` 미스캔이라 확장 필수, `moai update` 보존목록도 prefix-scoped. |

---

## §A Context — 문제 정의

### A.1 배경

MoAI 오케스트레이션 표면은 두 축에서 동시에 비용을 지불하고 있다.

**축 1 — 에이전트 본문 과복잡도.** `.claude/agents/moai/` 10개 에이전트 본문 합계는 **2,417 라인**(§F.1 실측)이다. 이 중 상당량이 (a) 스스로 SSOT라고 선언한 규칙 파일의 내용을 본문에 다시 복제한 것, (b) `agent-authoring.md` § Prompt Craft가 명시적으로 금지한 Opus 4.6-era 방어적 스캐폴딩, (c) 동일 에이전트 안에서 2~3회 반복되는 동일 제약이다. 이 라인들은 **매 spawn마다 컨텍스트에 적재**되므로 비용은 호출 횟수에 비례해 누적된다.

**축 2 — 워크플로우 직렬 병목.** plan/run/sync 파이프라인의 다수 단계가 본질적으로 독립적인 작업을 단일 에이전트에 직렬로 통과시킨다. 동시에, 병렬 fan-out을 위해 이미 **작성·배포된 3개의 dynamic workflow 스크립트가 어느 워크플로우 문서에서도 호출되지 않는다**(§F.2 실측). 즉 병렬화 수단이 존재하는데 배선만 없다.

두 축은 하나의 근원을 공유한다. **병렬 배칭 지시가 카탈로그 전체에 단 1개 파일에만 존재한다** — `verification-batch-pattern.md` 또는 `agent-common-protocol.md § Parallel Execution`을 참조하는 에이전트 본문은 `plan-auditor.md` **단 1개**이며 나머지 9개는 참조가 전무하다(§F.3 실측). 병렬화 규범이 에이전트 계층에 도달하지 못한 채 규칙 파일 안에만 갇혀 있다.

### A.2 병렬화를 위해 write-concurrency 완화가 필요하지 **않다**

초기 초안(v0.1.0~v0.2.0)은 `agent-common-protocol.md`의 절대 금지형 동시성 규칙을 스코프 한정형으로 완화하는 것을 세 번째 축으로 두었다. 그러나 본 SPEC이 채택한 병렬화 형태는 **전부 read-only fan-out + single-writer**다 — drafter/judge는 쓰기를 하지 않고 적용자는 항상 1개다. 따라서 규칙 완화의 in-scope 수혜자가 0이며, `[HARD]` 파일 쓰기 레이스 가드를 넓힐 근거가 본 SPEC 안에 존재하지 않는다.

이 축은 v0.3.0에서 **전면 철회**되어 후속 SPEC으로 이관되었다(§C `Out of Scope — write-concurrency rule relaxation`). 본 SPEC의 병렬화는 현행 `[HARD]` 규칙을 **한 글자도 바꾸지 않은 채** 성립한다.

### A.3 목표

에이전트 본문에서 SSOT 중복·금지된 스캐폴딩·반복 제약을 제거하고(축 1), 기존 fan-out 자산을 배선·배포하며 read-only fan-out + single-writer 패턴으로 plan/run/sync를 재구조화한다(축 2). 배포에 수반되는 shipped-rule 정합성과 SSOT 방향도 함께 정리한다(축 3, Group 5~6).

---

## §B Requirements (GEARS)

### B.1 Group 1 — 고아 fan-out 자산 배선

#### REQ-APO-010 (Ubiquitous)
`plan.md` Phase Routing Table은 `plan-research-fanout.js`를 Phase 2(Project Exploration)와 Phase 6(Deep Research)를 통합한 리서치 수행 수단으로 명시 **shall**.

#### REQ-APO-011 (Where)
**Where** `.claude/workflows/<script>.js`가 실행 환경에 존재하고 런타임이 dynamic workflow를 지원하는 경우 오케스트레이터는 해당 스크립트를 launch **shall**; **Where** 어느 한 조건이라도 부재한 경우 기존 단일 에이전트 경로로 fallback **shall** 하며 이때 기능 손실이 발생하지 **shall not**.

capability gate는 배포(REQ-APO-069) 이후에도 유지 **shall** — 배포는 파일 존재를 보장할 뿐 런타임 지원(Claude Code dynamic workflow 최소 버전)을 보장하지 않으므로 gate는 여전히 필요하다.

#### REQ-APO-012 (Ubiquitous)
`run/task-decomposition.md` 품질 단계(Phase 13/16/17)와 `sync.md` 품질 단계(Phase 7)는 `sync-audit-4dim.js`를 4차원 증거 수집 수단으로 명시 **shall**.

Phase 13/16/17을 소유하는 것은 하위 스킬 `.claude/skills/moai/workflows/run/task-decomposition.md`이며 진입 라우터 `run.md`가 **아니다** — `run.md`는 해당 Phase들을 하위 스킬로 위임하는 라우팅 표만 보유하고(`run.md:51`), `internal/skills`의 `TestEntryRouterLOCCeiling`이 강제하는 200 LOC 상한 아래(실측 197)의 얇은 라우터이므로 배선을 수용할 여지가 없다. 반면 `sync.md`의 Phase 7 배선은 진입 라우터 본문에 위치하므로(`sync.md:56`) 위 문장의 `sync.md` 지정은 그대로 유효하다.

#### REQ-APO-013 (Ubiquitous)
3개 workflow 스크립트는 **증거 수집 수단(evidence vehicle)**으로만 기능 **shall** — 구속력 있는 PASS/FAIL verdict 소유권은 `plan-auditor` / `sync-auditor`에 유지되며 스크립트 산출물이 verdict를 대체하지 **shall not**.

#### REQ-APO-014 (Where)
**Where** 대상 소스 패키지 수가 높은(high-count) 경우에 한해 codemaps 단계는 `codemaps-extract.js`를 아키텍처 인사이트 증강 수단으로 호출 **shall** — SPEC-DWF-CODEMAPS-PILOT-001이 확정한 "추출 대체가 아닌 증강" 스코핑을 유지 **shall**.

#### REQ-APO-015 (When)
**When** 배선이 완료되면, 3개 스크립트 각각이 최소 1개 워크플로우 문서(`.claude/skills/moai/workflows/**`)에서 참조되는 것이 grep으로 확인 **shall**.

#### REQ-APO-016 (When)
**When** 배선(REQ-APO-010/012/015)과 배포(REQ-APO-069)가 완료되면, docs-site 4-locale이 주장하는 "두 스크립트가 실제 파이프라인에 투입되어 있다"는 서술은 **참이 되며**, 그 참임이 검증 **shall** — 즉 본 요구사항은 "미검증 주장을 정정"이 아니라 **"주장을 참으로 만든 뒤 검증"**이다. 검증은 zero-orphan grep(AC-APO-015) + 배포 존재 grep(AC-APO-069)의 동시 통과로 성립 **shall**.

**When** 배선 또는 배포 중 하나라도 미완인 채 SPEC이 종료되면, 해당 서술은 실제 상태에 맞게 정정 **shall** — 미검증 주장을 잔존시키지 **shall not**.

### B.2 Group 2 — read-only fan-out + single-writer 재구조화

#### REQ-APO-020 (Ubiquitous)
plan Phase 11(Independent SPEC Review)은 복수의 read-only 심사 렌즈를 병렬 수집한 뒤 `plan-auditor` 단일 에이전트가 구속력 있는 verdict를 산출하는 구조 **shall**.

#### REQ-APO-021 (Ubiquitous)
plan Phase 10 산출물 생성은 **단일 writer(manager-spec)** 유지 **shall** 하되, 산출물 전량을 단일 턴 병렬 `Write` 호출로 생성 **shall**, 그리고 `manager-spec.md` 본문의 산출물 개수 서술은 실제 산출물 집합과 일치 **shall**.

#### REQ-APO-022 (Where)
**Where** development_mode가 `tdd`인 경우, RED 단계 테스트 초안은 복수 read-only drafter가 병렬 작성하고 단일 `manager-develop`이 적용 **shall** — 테스트 파일에 대한 동시 쓰기는 발생하지 **shall not**.

#### REQ-APO-023 (Ubiquitous)
run Phase 13 / 16 / 17의 3회 직렬 감사 패스는 1회 병렬 증거 수집 + 1회 구속력 있는 `sync-auditor` verdict로 축약 **shall** 하며, 기존 최대 3회 반복 상한은 보존 **shall**.

#### REQ-APO-024 (Ubiquitous)
sync Phase 12(Execute Document Synchronization)는 5개 산출물군(CHANGELOG / README+docs-site / project-docs / SPEC-artifacts / codemaps)에 대해 **병렬 read-only drafter**를 운용하고 **단일 `manager-docs`가 순차 적용** **shall**. 이 형태가 확정 설계이며 disjoint-writer 변형은 채택하지 **shall not**.

#### REQ-APO-024b (Ubiquitous)
REQ-APO-024의 구조는 write-concurrency 규칙 개정과 **독립** **shall** — drafter는 전원 read-only이고 적용자는 단일이므로 동시 write가 발생하지 않는다. 따라서 현행 `[HARD]` 절대 금지형 규칙을 **그대로 둔 채** Phase 12 재구조화가 성립 **shall** 하며, §C가 후속 SPEC으로 이관한 규칙 완화의 진행 여부는 본 요구사항의 전제가 되지 **shall not**.

#### REQ-APO-025 (Where)
**Where** sync Phase 10 커버리지 갭이 복수 패키지에 걸쳐 있는 경우, 패키지별 테스트 생성은 병렬 fan-out으로 수행 **shall**.

#### REQ-APO-026 (When)
**When** sync Phase 1(`gate-sync-1`)과 Phase 7이 실행될 때, run Phase 15가 기록한 `moai verify` 스냅샷이 신선(키 일치 + TTL 이내)하면 해당 검사 범주를 재실행하지 **shall not** 하고 스냅샷을 증거로 인용 **shall**; 신선하지 않으면 재실행 **shall**.

#### REQ-APO-027 (Where)
**Where** MX 스캔 대상이 복수 패키지·디렉터리로 분할 가능한 경우, 스캔은 샤딩된 read-only fan-out으로 수행 **shall**.

#### REQ-APO-028 (Ubiquitous)
본 SPEC이 도입하는 모든 fan-out은 **오케스트레이터가 launch** **shall** — subagent nesting에 의존하지 **shall not** (평면 계층 유지).

#### REQ-APO-029 (Ubiquitous)
기존 HUMAN GATE는 전량 보존 **shall**: plan Decision Point 1, Implementation Kickoff Approval, `gate-sync-1`, `gate-sync-2`. 어떤 fan-out도 게이트를 우회하거나 무인 통과시키지 **shall not**.

#### REQ-APO-030 (Ubiquitous)
`AskUserQuestion` 오케스트레이터 전용 경계는 보존 **shall** — 모든 drafter/judge 서브에이전트는 사용자에게 질문하지 **shall not** 하고 구조화된 blocker report를 반환 **shall**.

### B.3 Group 3 — 에이전트 본문 다이어트

#### REQ-APO-040 (Ubiquitous)
`plan-auditor.md`는 12-field frontmatter 스키마를 본문에서 **최대 1회만** 서술 **shall** — 현재 MP-3와 FC-1..FC-12 두 곳에 존재하는 중복 열거 중 하나는 SSOT(`spec-frontmatter-schema.md`) 교차참조로 대체 **shall**.

#### REQ-APO-041 (Ubiquitous)
`manager-spec.md`의 12-field frontmatter 스키마 블록은 SSOT 교차참조로 대체 **shall**.

#### REQ-APO-042 (Ubiquitous)
`plan-auditor.md`의 M6 Chain-of-Verification 절과 그 보고 템플릿 섹션은 제거 **shall** — `agent-authoring.md § Prompt Craft`가 금지한 Opus 4.6-era 방어적 스캐폴딩에 해당한다.

#### REQ-APO-043 (Ubiquitous)
`manager-spec.md`의 SPEC-ID 사전 자가검사 프로토콜은 **실행 가능한 Bash 정규식 검사를 유지한 채** 의례적 서술(단계별 decomposition 출력 마커 강제, 예시 표, AC sub-ID 혼동 표)을 축약 **shall**.

#### REQ-APO-044 (Ubiquitous)
`manager-spec.md` Step 5 검증 체크리스트는 Step 4가 이미 서술한 제약을 재진술하지 **shall not**.

#### REQ-APO-045 (Ubiquitous)
`manager-spec.md`의 GEARS/EARS 문법 패턴 표는 `moai-workflow-spec` 스킬 교차참조로 대체 **shall**.

#### REQ-APO-046 (When)
**When** `manager-spec.md` Step 4가 산출물 병렬 생성을 지시할 때, 서술된 파일 개수는 실제 열거된 파일 개수와 일치 **shall**.

#### REQ-APO-047 (Ubiquitous)
`manager-develop.md`의 DDD / TDD 두 워크플로우는 공통 골격 1개 + 모드별 차이 서술로 통합 **shall** — 동형 워크플로우를 전문 2회 기술하지 **shall not**.

#### REQ-APO-048 (Where)
**Where** 대상 변경이 서로 독립적인 패키지에 걸쳐 있는 경우, `manager-develop.md`의 "one atomic change at a time" 제약은 패키지 내부 범위로 한정 **shall** — 독립 패키지 간 동시 진행을 금지하지 **shall not**.

#### REQ-APO-049 (Ubiquitous)
`sync-auditor.md`는 단일 scoring model과 단일 report template만 서술 **shall**.

#### REQ-APO-050 (Ubiquitous)
`manager-docs.md`에서 실제 소유 범위(CHANGELOG / README / docs-site / frontmatter 전이)와 무관한 레거시 서술(Nextra 프레임워크 설정, WCAG 접근성 점수, page-speed 지표)은 본문에서 제거 **shall**.

#### REQ-APO-051 (Where)
**Where** e2e 레시피가 현재 호스트 OS에서 실행 불가능한 경우, 해당 레시피는 에이전트 본문이 아니라 on-demand 스킬 레퍼런스에 위치 **shall**.

#### REQ-APO-052 (Ubiquitous)
`manager-git.md`의 `merge_method` 해석 규칙은 본문에서 1회만 서술 **shall**.

#### REQ-APO-053 (Ubiquitous)
`builder-harness.md`에서 `model-policy.md`를 재진술하는 블록과 이미 종료된 마이그레이션 안내는 제거 **shall**.

#### REQ-APO-054 (Ubiquitous)
다중 검증을 수행하는 모든 retained 에이전트 본문은 병렬 배칭 규범(`agent-common-protocol.md § Parallel Execution` / `verification-batch-pattern.md`)에 대한 교차참조를 1줄 보유 **shall**.

#### REQ-APO-055 (Ubiquitous)
`.claude/agents/moai/*.md` 10개 파일의 합계 라인 수는 지정된 상한 이하 **shall** 하며, 파일별 상한도 각각 충족 **shall**(§D.2 표).

### B.4 Group 4 — 불변식 및 Template-First

#### REQ-APO-060 (When)
**When** 템플릿 미러가 존재하는 파일을 수정할 때, 편집은 `internal/template/templates/` 원본에 먼저 수행하고 `make build` 후 로컬로 동기화 **shall**.

#### REQ-APO-061 (Ubiquitous)
템플릿 미러 산출물은 내부 개발 흔적(SPEC ID, REQ 토큰, 내부 날짜, commit SHA)을 포함하지 **shall not**.

#### REQ-APO-062 (Ubiquitous)
`.claude/workflows/` 하위의 **dev-only 및 user-owned harness Runner** 비배포 불변식은 유지 **shall** — `hns-*` / `harness-*` 접두 파일이 `internal/template/templates/.claude/workflows/`에 존재하지 **shall not**. 이 불변식은 3개 generic fan-out 스크립트의 배포(REQ-APO-069)와 **양립** **shall** — 두 집합은 접두사로 분리된다.

#### REQ-APO-063 (Where)
**Where** 배포 사용자 환경에 `.claude/workflows/`가 존재하지 않는 경우, 본 SPEC이 추가한 모든 참조는 graceful degradation **shall** — 오류·경고·워크플로우 중단을 유발하지 **shall not**.

#### REQ-APO-064 (When)
**When** 구현이 완료되면 `go test ./...`가 green **shall** 하고 template-neutrality CI guard가 green **shall**.

#### REQ-APO-065 (Ubiquitous)
에이전트 frontmatter의 동작 결정 필드(`name`, `description`, `tools`, `model`, `effort`, `skills`)는 변경되지 **shall not** — 본 SPEC은 본문(body) 범위 작업이다.

#### REQ-APO-066 (Ubiquitous)
archived 12개 에이전트 이름은 어떤 편집 산출물에도 신규 도입되지 **shall not**.

#### REQ-APO-067 (When)
**When** 템플릿 편집이 수행되면 `internal/template/split_namespace_test.go`와 `internal/template/internal_content_leak_test.go`가 green을 유지 **shall**.

#### REQ-APO-068 (Ubiquitous)
제거된 모든 본문 중복은 선언된 SSOT를 가리키는 교차참조로 대체 **shall** — 정보의 무성 소실(silent information loss)이 발생하지 **shall not**.

### B.5 Group 5 — dynamic workflow 스크립트 배포 (사용자 결정 D1)

#### REQ-APO-069 (Ubiquitous)
`plan-research-fanout.js`, `sync-audit-4dim.js`, `codemaps-extract.js` 3개 스크립트는 `internal/template/templates/.claude/workflows/`에 미러되어 배포 사용자에게 전달 **shall**.

#### REQ-APO-070 (Ubiquitous)
본 SPEC은 `SPEC-DWF-CODEMAPS-PILOT-001`의 **비배포 acceptance criterion을 명시적으로 supersede** **shall** — 대상 criterion은 `AC-DCP-010`(정의 위치 `acceptance.md:79`, 판정 기록 `progress.md:86`)이며, 이를 소유하는 요구사항은 `REQ-DCP-009` / `REQ-DCP-010`이다. 해당 AC의 정식 판정 문구인 `grep -r "codemaps-extract\|codemaps-pilot" internal/template/templates/` → nothing 은 본 SPEC 이후 무효이며, 그 사실이 본 SPEC 산출물과 선행 SPEC 아티팩트 양쪽에 기록 **shall**. 선행 판정을 침묵 속에 위반하지 **shall not**.

#### REQ-APO-071 (Ubiquitous)
배포되는 3개 스크립트는 §25 템플릿 중립성을 충족 **shall** — 내부 SPEC ID, REQ/AC 토큰, 내부 날짜, commit SHA, 메인테이너 절대 경로가 스크립트 헤더·주석·문자열 어디에도 잔존하지 **shall not**.

#### REQ-APO-072 (When)
**When** `.js` 자산이 템플릿 트리에 최초 배포되면, 중립성 leak 스캐너의 대상 확장자 집합에 `.js`가 추가 **shall** — 스캐너가 `.js`를 읽지 않는 상태에서 통과하는 중립성 판정은 공허(vacuous)하며 증거로 인정되지 **shall not**.

#### REQ-APO-073 (Ubiquitous)
배포된 3개 스크립트는 `moai update`의 **template-managed**(덮어쓰기 가능) 집합에 속 **shall** — user-owned 보존 집합(`hns-*` / `harness-*`)에 편입되지 **shall not**. 사용자 개인 Runner Workflow의 보존 계약은 변경되지 **shall not**.

### B.6 Group 6 — 배포 정합성 (shipped rule / SSOT / 목적지)

#### REQ-APO-074 (When)
**When** 3개 스크립트가 템플릿에 배포되면, `dynamic-workflows.md`의 두 서술 — "MoAI does not ship any saved workflows by default; the user-owned `.claude/workflows/` directory is not template-managed"(L80)와 "the validated script lives in the local, user-owned `.claude/workflows/` directory (not template-managed, per the statement above)"(L131) — 은 개정 **shall**. 개정 후 서술은 **MoAI-shipped generic fan-out 스크립트**와 **user-owned `hns-*` / `harness-*` Runner**를 구분 **shall** 하며, 배포된 스크립트가 존재하는 상태에서 "어떤 saved workflow도 배포하지 않는다"는 서술이 잔존하지 **shall not**.

#### REQ-APO-075 (When)
**When** REQ-APO-074 개정이 수행되면, 로컬 `.claude/rules/moai/workflow/dynamic-workflows.md`와 `internal/template/templates/.claude/rules/moai/workflow/dynamic-workflows.md`는 개정 후에도 byte-parity를 유지 **shall** — 두 파일은 개정 전 0-diff 상태이므로 한쪽만 고치면 미러가 깨진다.

#### REQ-APO-076 (Ubiquitous)
배포되는 `.js` 스크립트 헤더 주석은 자기 자신을 "user-owned workflows"로 기술하지 **shall not** — `plan-research-fanout.js` L36과 `sync-audit-4dim.js` L38의 해당 문구는 배포 후 사실과 모순되므로 정정 **shall**.

#### REQ-APO-077 (Ubiquitous)
배포되는 3개 generic fan-out 스크립트의 SSOT는 `internal/template/templates/.claude/workflows/` **shall** — 편집은 템플릿 원본에서 수행하고 로컬 `.claude/workflows/`는 파생본으로 취급 **shall**. 로컬을 먼저 고쳐 템플릿에 복사하는 방향은 금지 **shall not**.

이 SSOT 방향의 실무적 귀결은 문서화 **shall**: 3개 스크립트는 user-owned 보존 집합 밖이므로(REQ-APO-073) `moai update` 실행 시 **로컬 사본이 템플릿 내용으로 덮어써진다**. 사용자가 로컬에서 스크립트를 수정하면 다음 `moai update`에서 소실되며, 이 사실이 `dynamic-workflows.md` 개정 서술에 명시 **shall**.

#### REQ-APO-078 (Ubiquitous)
REQ-APO-051이 이관하는 e2e 비-호스트 OS 레시피의 목적지는 `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` **shall** — 목적지 경로는 spec.md / plan.md / §E.2 파일 인벤토리 / frontmatter `module:`에 모두 명시 **shall**. 해당 경로는 템플릿 미러가 존재하므로 Template-First 및 byte-parity 의무가 적용 **shall**.

---

## §C Exclusions — 범위 제외

### Out of Scope — Go 프로덕션 코드 변경

- `internal/` 및 `pkg/` 하위 Go **프로덕션** 구현 변경은 본 SPEC 범위 밖이다. 본 SPEC은 지시문(instruction) 계층 — 에이전트 본문, 워크플로우 스킬 문서, 규칙 파일, 템플릿 자산 — 이 주 대상이다.
- `internal/template/templates/` 하위 **템플릿 자산** 편집은 Template-First 원칙상 필수이며 범위 내다.
- **유일한 예외**: `internal/template/internal_content_leak_test.go`의 `leakTextExtensions`에 `.js`를 추가하는 변경은 범위 내다(REQ-APO-072). 이것이 없으면 중립성 판정이 공허해진다(§F.8.3). 그 외 신규 CI 가드 테스트 작성은 범위 밖이다.
- `split_namespace_test.go` 가드 개정은 범위 밖이며 **불필요**하다 — 이미 prefix-scoped이다(§F.8.2). `internal/cli/update/plan/plan.go` 보존 분류 변경도 범위 밖이며 불필요하다(§F.8.4).

### Out of Scope — 에이전트 카탈로그 변경

- 11개 retained 에이전트의 추가·삭제·이름 변경은 범위 밖이다.
- 에이전트 frontmatter의 `model` / `effort` / `tools` 튜닝은 범위 밖 — 해당 축은 `moai model profile` 3-tier 프로파일 소관이다.
- `sync-auditor`의 read-only nesting 파일럿(`Agent` tool 보유) 설정 변경은 범위 밖이며, 본 SPEC의 모든 fan-out은 오케스트레이터 launch로 설계된다(REQ-APO-028).

### Out of Scope — 게이트 및 사용자 상호작용 의미 변경

- HUMAN GATE의 발화 조건·순서·필수성 변경은 범위 밖이다. 병렬화는 게이트 사이 구간의 실행 형태만 바꾼다.
- `AskUserQuestion` 채널 독점 규칙 및 orchestrator-subagent 비대칭 경계의 변경은 범위 밖이다.
- plan-auditor / sync-auditor의 verdict 소유권 이전은 범위 밖이다(REQ-APO-013).

### Out of Scope — 성능 실측 및 벤치마크

- 병렬화로 인한 wall-time 단축량의 정량 측정(A/B 벤치마크)은 범위 밖이다. 본 SPEC의 AC는 구조적 검증(grep / line count / 참조 도달성)만 요구한다.
- 토큰 절감량의 정량 측정 역시 범위 밖이다. 라인 수 감소는 대리 지표(proxy)로만 사용한다.

### Out of Scope — 신규 workflow 스크립트 작성

- 새 `.js` dynamic workflow 스크립트를 작성하지 않는다. 본 SPEC은 **이미 존재하는 3개** 스크립트를 배선하고 배포한다.
- 3개 스크립트 내부 **로직**의 기능 변경은 범위 밖이다. 배포를 위한 §25 중립화(헤더·주석의 내부 토큰 제거, REQ-APO-071)는 범위 내이며 실행 로직을 바꾸지 않는다.
- `hns-oss-docs-run.js` / `hns-release-update-run.js` 등 harness Runner의 배포는 범위 밖이다(비배포 유지, REQ-APO-062).

### Out of Scope — write-concurrency rule relaxation

v0.1.0~v0.2.0의 Group 1(REQ-APO-001..005 — `agent-common-protocol.md` / `CLAUDE.md`의 write-concurrency 규칙을 절대 금지형에서 스코프 한정형으로 완화)은 **전면 철회**되어 본 SPEC 범위 밖이다.

- **철회 사유**: REQ-APO-024b가 이미 sync Phase 12를 Group 1에서 분리했고, 바로 아래 `Out of Scope — disjoint-writer 병렬 쓰기 변형`이 유일한 잠재 소비자였던 disjoint-writer 변형을 제외한다. 따라서 본 SPEC 안에 규칙 완화의 수혜자가 **0**이며, `[HARD]` 파일 쓰기 레이스 가드를 수혜자 없이 넓히는 결과가 된다.
- **본 SPEC의 모든 병렬화는 현행 규칙을 바꾸지 않고 성립한다** — drafter/judge는 전원 read-only이고 적용자는 항상 단일이므로 동시 write 자체가 발생하지 않는다.
- **이관 대상 작업**: manifest 포맷 정의, manifest와 spawn 프롬프트의 바인딩 방식, 교집합 판정 checker, 파일 이외 공유 상태(git index, 원격 브랜치, 외부 API)의 취급, 그리고 최소 1개의 실제 강제 지점(enforcement point).
- **후속 SPEC 소관**: 위 5개 항목은 `design.md`를 갖춘 별도 SPEC에서 다룬다(§G 후속 SPEC 항목). 본 SPEC은 후속 SPEC을 저작하지 않는다.

### Out of Scope — disjoint-writer 병렬 쓰기 변형

- sync Phase 12의 경로 소유 writer 병렬 실행 변형은 **채택하지 않는다**(사용자 결정 D2). 확정 설계는 read-only drafter + 단일 적용자다(REQ-APO-024).
- 해당 변형은 **문서화된 향후 선택지**로만 보존한다 — 후속 SPEC이 구 Group 1(write-concurrency) 규칙 완화를 정착시킨 이후 재검토할 수 있으나 본 SPEC의 산출물·AC·마일스톤에 포함되지 않는다.
- write-concurrency 규칙 완화는 그 자체로 독립 가치를 가질 수 있으나 본 SPEC에서는 철회되었고(위 `Out of Scope — write-concurrency rule relaxation`), Phase 12는 그 진행 여부에 의존하지 않는다(REQ-APO-024b).

### Out of Scope — docs-site 본문 재작성

- docs-site 4-locale 콘텐츠의 전면 개편은 범위 밖이다. REQ-APO-016은 오직 "파이프라인 투입" 주장 1건의 진위 정합화만 요구한다.

---

## §D AC Matrix

### D.1 REQ → AC 매핑

| REQ 그룹 | REQ 범위 | AC 범위 | 검증 성격 |
|---|---|---|---|
| Group 1 — fan-out 배선 | REQ-APO-010..016 (7) | AC-APO-010..016 (7) | zero-orphan grep + capability-gate 문구 grep |
| Group 2 — 재구조화 | REQ-APO-020..030 + 024b (12) | AC-APO-020..030 + 024b (12) | 단계 서술 grep + 게이트 보존 grep + 규칙 독립성 서술 |
| Group 3 — 본문 다이어트 | REQ-APO-040..055 (16) | AC-APO-040..055 (16) | 중복 카운트 grep + line-count 분기 상한 |
| Group 4 — 불변식 | REQ-APO-060..068 (9) | AC-APO-060..068 (9) | 미러 diff + CI green + 부재 grep |
| Group 5 — 배포 (D1) | REQ-APO-069..073 (5) | AC-APO-069..073 + 071b + 072b (7) | 배포 존재 grep + supersession(AC ID 인용) + 비공허 중립성(CI 정렬) + manual-only 표기 + hns 차단 유지 + update 분류 |
| Group 6 — 배포 정합성 | REQ-APO-074..078 (5) | AC-APO-074..078 (5) | shipped rule 개정 grep + byte-parity + SSOT 방향 + 목적지 실재 |
| **합계** | **54 REQ** | **56 AC** | — |

전체 AC 열거·판정 명령·MUST/SHOULD 등급은 `acceptance.md` §D에 있다.

### D.2 라인 수 상한 (REQ-APO-055 판정 기준)

| 파일 | 현재(실측) | 상한 | 근거 |
|---|---:|---:|---|
| `plan-auditor.md` | 505 | **430** | **M3 실측 기반 재보정(§D.2.1)** — Groups 7/8의 5중 진술 중 재진술 3종 제거 + `verification-batch-pattern.md` 중복 AP 1행 + frontmatter 재진술. 구 상한 340은 루브릭 밴드·D7/D8 검증 동사·Output Format 템플릿을 삭제해야만 도달 가능해 REQ-APO-068에 저촉 |
| `manager-spec.md` | 317 | **230** (분기 A) / **250** (분기 B) | frontmatter 스키마 블록 + GEARS 표 + Step5 중복 제거. **분기 조건부** — D3 게이트(§`plan.md` §B.3)가 마커 소비자 0건이면 분기 A(230), ≥1건이면 출력 계약 보존분만큼 완화된 분기 B(250). **M3 실측 236 — 상한 유지**(잔여 6행이 정보 손실 없이 상쇄 가능, §D.2.1) |
| `manager-develop.md` | 311 | **240** | DDD/TDD 동형 워크플로우 통합. **M3 실측 238로 충족** |
| `sync-auditor.md` | 221 | **150** | scoring model 2→1, report template 2→1 |
| `manager-git.md` | 211 | **190** | merge_method 3회→1회 |
| `manager-design.md` | 201 | **205** | 다이어트 대상 아님(REQ-APO-054 교차참조 1줄 허용) |
| `builder-harness.md` | 195 | **170** | model-policy 재진술 + stale 마이그레이션 안내 제거 |
| `e2e-tester.md` | 182 | **150** | 비-호스트 OS 레시피 스킬 레퍼런스 이관 |
| `manager-docs.md` | 167 | **120** | Nextra/WCAG/page-speed 레거시 제거 |
| `super-advisor.md` | 107 | **112** | 다이어트 대상 아님(교차참조 1줄 허용) |
| **합계** | **2,417** | **≤ 1,997** (분기 A) / **≤ 2,017** (분기 B) | 최소 16% 감축 (분기 A 17.3% / 분기 B 16.5%) |

상한은 각각 **이하(≤)** 판정이며, 합계 상한은 개별 상한의 합이다. 개별 파일이 상한보다 더 줄어드는 것은 허용된다(단 REQ-APO-068 정보 무성 소실 금지가 우선한다).

`manager-spec.md` 행은 **분기 조건부 상한**이다. D3 게이트 결과에 따라 적용 상한이 결정되며, 어느 분기든 해당 상한은 MUST로 판정된다 — 상한을 맞추기 위해 마커 출력 계약을 깨는 것은 금지된다(그 경우 분기 B 상한이 적용되어야 하며, "상한 미달성 + 사유 기록"으로 회피하지 않는다).

### D.2.1 상한 재보정 근거 (M3 실측)

§D.2 초기 상한은 **plan-phase 추정치**였다. M3 실행으로 3개 파일의 실측이 확보되어 아래와 같이 재보정한다. 재보정 원칙은 **수치에 내용을 맞추지 않고 내용에 수치를 맞춘다** — REQ-APO-068(정보 무성 소실 금지)이 상한보다 우선한다. 단, 그 역도 성립한다: **실측이 상한을 넘었다는 사실만으로 상한을 올리지 않는다.** 사소하게 충족되는 상한은 아무것도 측정하지 않기 때문이다.

| 파일 | plan 추정 | M3 실측 | 재보정 | 판정 |
|---|---:|---:|---:|---|
| `plan-auditor.md` | 340 | 454 | **430** | 상향 — 추정이 90행 낙관 |
| `manager-spec.md` | 230 (분기 A) | 236 | **230** 유지 | 잔여 6행 상쇄 가능 |
| `manager-develop.md` | 240 | 238 | **240** 유지 | 충족 |

**`plan-auditor.md` — 340이 도달 불가인 이유.** M3는 파일 간 SSOT 중복 제거로 505→454(−51)를 달성한 뒤 정지했고, 잔여 114행은 이 에이전트가 고유 소유한 내용(루브릭 밴드 정의, D7/D8 검증 동사, Audit Checklist Group 3-6, Output Format 템플릿)이라고 보고했다. **본 세션의 독립 판독 결과 그 보고는 대체로 옳으나 한 부류를 놓쳤다.** M3가 적용한 렌즈는 *파일 간* SSOT 중복이었고, 잔여 구간에 남은 것은 *파일 내* 동일 규칙 반복이다 — §A 축 1의 (c) "동일 에이전트 안에서 2~3회 반복되는 동일 제약". Group 7/8은 각 차원의 규칙을 **5중으로 진술**한다: ① H3 제목을 되받는 요약 불릿, ② 유래 산문, ③ 번호 체크(D7-1..5 / D8-1..4), ④ bash 검증 동사, ⑤ ③을 되받는 Severity rubric. ③④만 조작적이고 ①②⑤는 재진술이므로, 그 제거는 정보 손실이 아니다.

상한 430의 산출 근거 — 정보 손실 없이 제거 가능한 것으로 **직접 판독하여 확인한** 28행:

| 구간 | 행 | 근거 |
|---|---:|---|
| G7 요약 불릿 | 2 | H3 제목 `Group 7: Cross-SPEC Reconciliation (D7)`의 재진술 |
| G7 유래 산문 | 6 | 조작적 내용이 D7-2/D7-3/D7-4와 동일 |
| G7 Severity rubric | 3 | D7-4(BLOCKING)/D7-5(SHOULD) 재진술 |
| G8 요약 불릿 | 2 | H3 제목의 재진술 |
| G8 유래 산문 | 6 | 조작적 내용이 D8-2와 동일. lesson #21 동기는 D8-3에 존치 |
| G8 Severity rubric | 3 | D8-3/D8-4 재진술 |
| AP-VEM-003 | 1 | `verification-batch-pattern.md` AP-VBP-002와 동일(이미 교차참조된 파일) |
| Delegation Note 말미 | 2 | frontmatter `description`의 `NOT for:` 재진술 — M3가 `manager-spec.md`에 이미 적용한 선례 |
| Invocation Examples 2/3 | 2 | 보고서 경로만 다른 준동일 예시 |
| Output Format 2-stream 산문 | 1 | 인용한 SSOT(`spec-workflow.md` § Report Persistence)의 재진술 |
| **합계** | **28** | 454 − 28 = **426**(하한 추정) |

상한 430은 이 중 **24행**을 요구하고 4행의 여유를 남긴다 — 일부 항목이 run-phase 판독에서 load-bearing으로 재판정될 여지를 흡수한다.

[HARD] **공백 전용 압축 금지.** 430은 *실질 내용*의 재진술 제거로만 달성한다. 루브릭 밴드 사이 빈 줄을 접는 방식(약 19행 확보 가능)은 금지한다 — 라인 수는 컨텍스트 비용의 대리 지표이고 빈 줄 제거는 그 비용을 거의 줄이지 않으므로, 그렇게 충족한 상한은 아무것도 측정하지 않는다.

**`manager-spec.md` — 230 유지 근거.** 실측 236으로 6행 초과이나, § Status Responsibility Matrix의 1행 표가 § SPEC Artifact Ownership "Status transitions owned"와 동일 사실(`(none) → draft`, 4개 plan-phase 산출물)을 진술한다. 두 구간 병합으로 6~7행이 정보 손실 없이 확보되므로 상한을 완화하지 않는다.

**재보정 절차(향후 동일 상황).** 나머지 7개 파일의 상한은 여전히 **미검증 추정치**이며, 90행 낙관 편향을 보인 `plan-auditor.md`와 같은 방법으로 산출되었다. run-phase가 어떤 파일에서 "상한 도달이 REQ-APO-068 위반 없이는 불가"를 발견하면:

1. 삭제로 상한을 맞추지 **않는다**(§G 안티패턴 "라인 수를 맞추려 압축").
2. 잔여 구간이 고유 소유인지 재진술인지 **직접 판독**하여 제거 가능 행을 위 표와 동일 형식으로 열거한다.
3. 그 열거를 근거로 해당 파일 상한을 재보정하고 **합계 = 개별 상한의 합**으로 재계산한다.
4. 감축률 주장을 재계산된 합계에 맞춰 갱신한다 — 감축률은 목표가 아니라 **파생값**이다.

---

## §E Tier 판정 — L

Tier L로 판정한다. 근거:

- **표면 수**: 에이전트 본문 10 + 워크플로우 스킬 문서 3(plan/run/sync) + 하위 스킬 문서 최소 4(`run/phase-execution.md`, `run/task-decomposition.md`, `sync/doc-execution.md`, `sync/quality-gates-quality.md`) + 정규 규칙 2(`agent-common-protocol.md`, `CLAUDE.md`) + 각각의 템플릿 미러. 편집 파일 수는 **30개 이상**이다.
- **도메인 수**: 에이전트 정의 / 워크플로우 스킬 / 정규 규칙 / 템플릿 배포 / docs-site 콘텐츠 — 5개 도메인.
- **배포 위험**: Group 5~6은 **배포 사용자에게 출하되는 자산**(3개 `.js` + shipped rule 문구)을 변경한다. 중립성 누락 시 내부 토큰이 출하되고, shipped rule 미개정 시 자기모순 상태로 출하된다 — 되돌리기 비용이 로컬 편집보다 크므로 최고 심사 밀도가 필요하다. (참고: 본 SPEC은 `[HARD]` write-concurrency 규칙을 편집하지 않는다 — §A.2·§C 참조.)
- **불변식 밀도**: mirror byte-parity, 템플릿 중립성, 비배포 불변식, 게이트 보존, verdict 소유권 보존 — 5종 불변식이 동시에 걸린다.

Tier M은 30+ 파일 다중 미러 표면을 과소 커버한다. Tier L의 5-artifact 의무(spec + plan + acceptance + design + research)는 본 SPEC이 충족한다 — `progress.md`는 모든 Tier에 존재하므로 5번째 산출물로 계수되지 **않는다**.

### §E.2 파일 인벤토리 (편집 대상)

| 경로 | 성격 | 템플릿 미러 | REQ |
|---|---|---|---|
| `.claude/agents/moai/*.md` (10) | 에이전트 본문 | 있음 | 040-055 |
| `.claude/skills/moai/workflows/{plan,run,sync,codemaps}.md` | 워크플로우 진입 | 있음 | 010-016 |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | plan 하위 | 있음 | 020-021 |
| `.claude/skills/moai/workflows/run/{phase-execution,task-decomposition}.md` | run 하위 | 있음 | 022-023, 026-027 |
| `.claude/skills/moai/workflows/sync/{doc-execution,quality-gates-quality,quality-gates-context}.md` | sync 하위 | 있음 | 024-026 |
| `.claude/skills/moai-workflow-testing/references/e2e-desktop-native-recipes.md` | **신규** — e2e 비-호스트 레시피 목적지 | 있음 | 051, 078 |
| `.claude/rules/moai/workflow/dynamic-workflows.md` | shipped rule (L80/L131 개정) | 있음 | 074-075, 077 |
| `.claude/workflows/{plan-research-fanout,sync-audit-4dim,codemaps-extract}.js` | fan-out 스크립트 (중립화 + 헤더 정정) | **신규 생성** | 069, 071, 076 |
| `internal/template/templates/.claude/workflows/*.js` (3) | **신규** — 배포 원본(SSOT) | (원본) | 069, 077 |
| `internal/template/internal_content_leak_test.go` | `.js` 확장자 추가 | n/a | 072 |
| `docs-site/content/{en,ko,ja,zh}/claude-code/agentic/workflows.md` | 주장 검증 | n/a | 016 |
| `.moai/specs/SPEC-DWF-CODEMAPS-PILOT-001/*` | supersession 주석 | n/a | 070 |

편집 대상은 **로컬 + 템플릿 미러 쌍**을 각각 1건으로 세면 약 **40 파일**이다.

미편집 확인 대상(변경 없음을 검증): `internal/template/split_namespace_test.go`(§F.8.2), `internal/cli/update/plan/plan.go`(§F.8.4), `.claude/rules/moai/core/agent-common-protocol.md`(구 Group 1 철회로 무편집), `CLAUDE.md`(동일).

---

## §F Ground Truth — 본 세션 실측

모든 수치는 본 plan-phase 세션에서 직접 명령을 실행해 관측한 것이다. 전임 보고서·기억에서 이월한 값은 없다.

### F.1 에이전트 본문 라인 수

`wc -l .claude/agents/moai/*.md` 관측: plan-auditor 505 / manager-spec 317 / manager-develop 311 / sync-auditor 221 / manager-git 211 / manager-design 201 / builder-harness 195 / e2e-tester 182 / manager-docs 167 / super-advisor 107, 합계 **2,417**.

### F.2 고아 fan-out 자산

`.claude/workflows/` 실재 파일: `plan-research-fanout.js`, `sync-audit-4dim.js`, `codemaps-extract.js`, `hns-oss-docs-run.js`, `hns-release-update-run.js`.

`.claude/skills/moai/` 하위에서 앞의 3개 스크립트명을 검색한 결과 **0건**. 즉 plan/run/sync/codemaps 워크플로우 문서 어디에서도 호출되지 않는다.

### F.3 병렬 배칭 교차참조 분포

`grep -ln "verification-batch-pattern\|Parallel Execution" .claude/agents/moai/*.md` 관측: **`plan-auditor.md` 1개 파일**만 매치. 나머지 9개 에이전트는 병렬 배칭 규범 참조 전무.

### F.4 write-concurrency 3표면 (후속 SPEC 입력 — 본 SPEC은 무편집)

구 Group 1 철회에 따라 아래 실측은 **후속 SPEC의 입력 자료**로만 보존된다. 본 SPEC은 이 3표면을 편집하지 않는다.

- `agent-common-protocol.md` L191 / L193 / L198 — 절대 금지 형태.
- `CLAUDE.md` L250 및 `internal/template/templates/CLAUDE.md` L250 — 동일 절대 금지 문장(byte-parity 상태).
- `.claude/skills/moai/workflows/e2e.md` L251 — **이미 스코프 한정 형태** ("on overlapping scope") — 두 표면의 불일치는 실재하나 본 SPEC의 소관이 아니다.
- 추가 관찰(후속 SPEC 주의): read-only 안전장치 문장이 `agent-common-protocol.md`에 **두 번** 나타난다 — L198의 "orchestrator work concurrent with a write-capable agent stays read-only"와 L193의 변형 "orchestrator work **performed concurrently with** a write-capable agent is read-only". 단일 리터럴 grep으로 보존을 검증하면 L193 변형을 놓친다.

### F.5 템플릿 미러 인벤토리

미러 존재: `.claude/agents/moai`, `.claude/skills/moai/workflows`(+ `run`/`sync`/`plan` 하위), `.claude/rules/moai/core`, `.claude/rules/moai/workflow`, `CLAUDE.md`.
미러 **부재**: `.claude/workflows/` (3개 `.js` 포함).

### F.6 단계 번호 실측

- plan: Phase 2 Project Exploration / Phase 6 Deep Research / Phase 10 SPEC Document Creation / Phase 11 Independent SPEC Review.
- run: Phase 13 Quality Validation / Phase 15 Pre-Review Quality Gate(`moai verify` 스냅샷 record 지점) / Phase 16 Active Quality Evaluation / Phase 17 TRUST 5 Static Verification / Phase 9 Pre-Implementation MX Context Scan / Phase 18 MX Tag Update.
- sync: Phase 1 Pre-Sync Quality Gate(`gate-sync-1`) / Phase 7 Quality Check / Phase 10 Coverage Analysis / Phase 11 Analysis + `gate-sync-2` / Phase 12 Execute Document Synchronization(Step 2.2 CHANGELOG·README, Step 2.2.5 project docs + codemaps, Step 2.3 post-sync quality, Step 2.4 SPEC status, Step 2.4.1 issue sync).

### F.7 브리프 대비 실측 정정 2건

**정정 1 — "orphaned = zero references"는 부분적으로 거짓.** 3개 스크립트는 워크플로우 문서에서는 참조 0건이 맞으나, `docs-site/content/{en,ko,ja,zh}/claude-code/agentic/workflows.md:120`이 4개 로케일 전부에서 두 스크립트를 "실제 파이프라인에 투입"했다고 서술하고 있으며, `dynamic-workflows.md:105`는 `codemaps-extract.js`를 canonical worked example로 문서화하고 있다. 따라서 정확한 진술은 **"워크플로우 호출 경로에서 고아이며, 동시에 공개 문서가 배선되었다고 주장하는 미검증 클레임이 존재한다"**이다. 이 사실이 REQ-APO-016을 신설하게 했다.

**정정 2 — `.claude/workflows/`는 템플릿 미러가 없다.** 브리프는 "어떤 파일이 미러를 갖는지 확인하고 스코프하라"고만 지시했으나, 실측 결과 3개 `.js`는 의도적 비배포이며 SPEC-DWF-CODEMAPS-PILOT-001이 이를 명시 AC로 고정해 두었다. 반면 배선 대상인 `plan.md`/`run.md`/`sync.md`는 **미러 대상**이다. 따라서 배선은 반드시 capability-gated 형태(REQ-APO-011/063)여야 하며, 무조건적 참조는 배포 사용자에게 존재하지 않는 파일을 가리키게 된다. 이 비대칭은 사용자 결정 D1(배포)로 해소되었으나 **capability gate 자체는 유지**된다 — 파일 배포가 런타임 지원까지 보장하지는 않기 때문이다(§F.8.4).

---

## §F.8 사용자 결정 D1/D2/D3 반영 및 배포 전제 실측 정정

### F.8.1 결정 요약

| 결정 | 내용 | 반영 |
|---|---|---|
| D1 | 3개 fan-out 스크립트를 **템플릿 배포** | Group 6 신설(REQ-APO-069..073), REQ-APO-062 재정의 |
| D2 | sync Phase 12 = **read-only drafter + 단일 적용자** 확정, disjoint-writer 불채택 | REQ-APO-024 확정형, REQ-APO-024b(M1 독립성) 신설 |
| D3 | SPEC-ID 마커 축약은 **선행 grep 게이트 후** 결정 | `plan.md` M4 첫 작업으로 게이트화, REQ-APO-043 조건부 |

### F.8.2 정정 1 — split_namespace 가드는 이미 prefix-scoped이며 개정 불요

`internal/template/split_namespace_test.go` L93-104는 `.claude/workflows/*.js` 전체를 차단하지 **않는다**. 실제 차단 조건은 `splitHarnessAgentPrefixes`(`harness-release-update`, `harness-github`, `harness-release`, `hns-release-update`, `hns-github`, `hns-release`) 중 하나로 시작하는 basename이다. 5개 스크립트명에 대해 접두사 매칭을 실행한 결과:

```
plan-research-fanout.js          NOT blocked
sync-audit-4dim.js               NOT blocked
codemaps-extract.js              NOT blocked
hns-oss-docs-run.js              NOT blocked (user-owned, 의도적)
hns-release-update-run.js        BLOCKED (prefix=hns-release)
```

즉 가드는 이미 정확히 원하는 분리를 수행하고 있다 — **가드 개정·축소 작업은 필요하지 않으며**, 요구되는 것은 배포 이후에도 `hns-*` 차단이 유지됨을 확인하는 **불변식 단언**(AC-APO-072b)뿐이다. 존재하지 않는 차단을 전제로 가드를 수정하면 오히려 dev-only 보호를 약화시킨다.

### F.8.3 정정 2 — 중립성 leak 가드는 `.js`를 스캔하지 않는다 (공허 통과 위험)

`internal/template/internal_content_leak_test.go`의 `leakTextExtensions`는 `.md` / `.tmpl` / `.yaml` / `.yml` / `.sh` / `.json` 6종이며 **`.js`가 없다**. 따라서 3개 스크립트를 템플릿에 추가한 뒤 "중립성 가드 green"을 근거로 삼으면 **가드가 파일을 읽지도 않은 채 통과한 공허 판정**이 된다.

실제 중립성 위반은 존재한다:

| 파일 | 라인 | 위반 |
|---|---|---|
| `codemaps-extract.js` (62줄) | — | **0건 — 이미 중립** |
| `plan-research-fanout.js` (132줄) | 35-36, 54 | `REQ-ATR-018/019/020`, `AC-ATR-023/024/025`, `design.md §D`, `acceptance.md` 내부 참조 |
| `sync-audit-4dim.js` (173줄) | 37-38 | `REQ-ATR-015/016/017`, `AC-ATR-020/021/022` |

따라서 REQ-APO-072(`.js` 확장자 추가)는 REQ-APO-071(중립화) 판정이 유효해지기 위한 **선행 조건**이다. 이것이 D1이 실제로 요구하는 유일한 Go 변경이며, 예상되었던 "가드 축소"와는 반대 방향이다.

#### F.8.3-a 정정 — CI 가드 정규식 클래스의 실제 범위 (v0.2.0 서술 오류)

v0.2.0의 §F.8.3은 `sync-audit-4dim.js:42`의 예시 문자열 `spec_id: "SPEC-FOO-001"`이 "C1 정규식 `SPEC-[A-Z0-9-]+-[0-9]{3}` 매칭 대상"이라고 기술했다. **이는 사실이 아니다.** `internal_content_leak_test.go`의 실제 클래스는 접두사 제한형이다:

| 클래스 | 실제 정규식 (실측) |
|---|---|
| C1 (`C1-spec-id-prefix`) | `\bSPEC-(V3R[2-6]\|AGENCY\|WORKTREE)-[A-Z0-9-]+\b` |
| C2 (`C2-req-ac-internal-prefix`) | `\b(REQ\|AC)-(ATR\|WO\|COORD\|UNP\|LNC\|TII)-[0-9]{3}\b` |
| S2 (short-sha) | `\b[0-9a-f]{7,8}([\s\.,;:!?]\|$)` |

토큰별 매칭 실측:

| 토큰 | C1 | C2 |
|---|---|---|
| `SPEC-FOO-001` | **미매치** | **미매치** |
| `REQ-ATR-018` / `AC-ATR-023` / `REQ-ATR-015` | 미매치 | **매치** |
| `SPEC-V3R6-FOO-001` | 매치 | 미매치 |

귀결 3가지:

1. `SPEC-FOO-001`은 CI 가드가 잡지 않는다. 이것은 도메인이 `FOO`인 **일반 플레이스홀더**이며 내부 개발 흔적이 아니므로 실질적으로도 무해하다. 중립화 대상은 `REQ-ATR-*` / `AC-ATR-*` / `design.md §C/§D` / `acceptance.md` 참조에 한정된다.
2. **RED/GREEN 왕복(AC-APO-072)은 달성 가능하다** — C2가 `plan-research-fanout.js:35-36,54`와 `sync-audit-4dim.js:37-38`의 실제 `REQ-ATR-*` / `AC-ATR-*` 토큰을 genuinely 매치하므로, `.js` 확장자 추가 전후로 FAIL→PASS 전이가 실제로 관측된다. 다만 이 왕복이 입증하는 것은 **"스캐너가 C2 클래스에 대해 `.js`를 읽는다"**이지 AC-APO-071이 나열한 정규식 전체가 CI로 강제된다는 뜻이 아니다.
3. AC-APO-071의 수동 정규식은 CI 클래스보다 **넓다**(SHA 범위 `{9,40}` vs CI `{7,8}`은 겹치지도 않는다). 비중첩 토큰 클래스는 **manual-only / CI-unenforced**로 명시 표기해야 하며, "CI green"을 그 클래스의 근거로 인용해서는 안 된다.

### F.8.4 정정 3 — `moai update` 보존 목록도 prefix-scoped이며 배포와 양립

`internal/cli/update/plan/plan.go` L135-145 / L189-196은 `.claude/workflows/hns-` 및 `.claude/workflows/harness-` 접두 경로만 user-owned로 분류한다. 3개 generic fan-out은 이 집합에 속하지 않으므로 자동으로 **template-managed**(덮어쓰기 가능)가 되며, 이는 배포 자산에 요구되는 정확한 의미다. 보존 계약 변경은 필요하지 않다(REQ-APO-073은 이 상태의 유지를 단언한다).

또한 `internal/template/catalog.yaml`에는 `.claude/workflows` 항목이 존재하지 않는다(유일한 "workflows" 매치는 스킬 설명 문자열). 배포는 embedded 트리의 generic FS walk로 이루어지므로 catalog 등록은 불필요하다.

### F.8.5 capability gate 존속 근거

배포는 **파일 존재**를 보장하지만 **런타임 지원**을 보장하지 않는다. dynamic workflow 실행은 Claude Code 최소 버전 요구가 있으며, 구버전 런타임의 사용자는 파일을 받고도 실행할 수 없다. 따라서 REQ-APO-011의 capability gate는 배포 이후에도 유지되며, gate 조건이 "파일 존재"에서 "파일 존재 AND 런타임 지원"으로 확장된다.

---

## §G Cross-References

- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution — REQ-APO-054가 각 에이전트 본문에 교차참조로 삽입할 SSOT. **본 SPEC은 이 파일을 편집하지 않는다**(§E.2 미편집 확인 대상). § Background Agent Execution의 write-concurrency 문장은 후속 SPEC 소관이다.
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — 병렬 배칭 근거 및 클래스 분류.
- `.claude/rules/moai/development/agent-authoring.md` § Prompt Craft — REQ-APO-042 금지 근거.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — REQ-APO-040/041 교차참조 대상 SSOT.
- `.claude/rules/moai/workflow/dynamic-workflows.md` — workflow 원시 및 `codemaps-extract.js` worked example.
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` §C.2 — Mode 4 동시 spawn 3-5 상한(REQ-APO-028 fan-out 폭 제약).
- `SPEC-DWF-CODEMAPS-PILOT-001` — `codemaps-extract.js` 스코핑 선례. 본 SPEC이 그 **비배포 acceptance criterion `AC-DCP-010`**(소유 요구사항 `REQ-DCP-009` / `REQ-DCP-010`; 정식 문구 `grep -r "codemaps-extract\|codemaps-pilot" internal/template/templates/` → nothing)을 **부분 supersede**한다(REQ-APO-070); high-count 증강 스코핑은 유지된다(REQ-APO-014).
- `internal/template/split_namespace_test.go` — `hns-*` / `harness-*` Runner 차단 가드(prefix-scoped, 개정 불요 — §F.8.2).
- `internal/template/internal_content_leak_test.go` — `leakTextExtensions`에 `.js` 추가 대상(REQ-APO-072, §F.8.3).
- `internal/cli/update/plan/plan.go` — user-owned 보존 분류(prefix-scoped, 변경 불요 — §F.8.4).
- `internal/template/internal_content_leak_test.go` — C1/C2/S2 클래스 정의(§F.8.3-a) + `leakTextExtensions` 확장 대상(REQ-APO-072).
- `.claude/rules/moai/workflow/dynamic-workflows.md` L80 / L131 — Group 6 개정 대상(REQ-APO-074).

### 후속 SPEC (본 SPEC에서 저작하지 않음)

- **`SPEC-WRITE-CONCURRENCY-SCOPE-001` (가칭)** — §C `Out of Scope — write-concurrency rule relaxation`이 이관한 축. 범위: `agent-common-protocol.md` / `CLAUDE.md`의 절대 금지형 → 스코프 한정형 개정, disjoint path manifest 포맷·바인딩·checker, 파일 이외 공유 상태(git index / 원격 브랜치 / 외부 API) 취급, 최소 1개 강제 지점. Tier L 예상(`design.md` 필수 — manifest 포맷이 설계 산출물이다). 입력 자료: 본 SPEC §F.4(3표면 실측 + L193/L198 이중 서술 관찰).
- 본 SPEC은 위 후속 SPEC을 저작하지 않으며, 그 진행 여부가 본 SPEC의 어떤 REQ/AC의 전제도 되지 않는다(REQ-APO-024b).
- `SPEC-WORKFLOW-CACHE-OPT-001` — `sync-audit-4dim` 병렬 심사 선례 인용.
- `CLAUDE.local.md` §2 / §25 — Template-First 및 템플릿 중립성.
