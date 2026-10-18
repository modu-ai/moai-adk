---
id: SPEC-CARD-ELIGIBLE-PULL-001
title: "카드 자격 임대 — bare `factory next` 제네릭 암 (b)/(b2)에 자격 게이트·self-dispatch 플래그·구조화 거부 도입"
version: "0.3.0"
status: in-progress
created: 2026-10-17
updated: 2026-10-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: M
amendment_of: SPEC-CARD-ELIGIBLE-PULL-001
tags: "factory-next, eligibility, self-dispatch, card-selection, structured-refusal, t1839"
---

# SPEC-CARD-ELIGIBLE-PULL-001

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-17 | manager-spec | 초기 플랜 페이스 초안(카드 t1839). 사고 t1780(2026-10-17): bare `moai factory next`를 실행한 레인이 다른 소비자를 기다리던 무관한 운영자 픽 카드를 임대 — 제네릭 암 (b)/(b2)에 관련성·자격 게이트가 없는 것이 원인. 운영자 위임 설계(카드 t1839 본문)를 GEARS REQ로 고정. 플래그 OFF 의미가 카드 본문에서 두 서술로 양립 불가하여 decision-index.md Q1으로 회부. |
| 0.2.0 | 2026-10-17 | manager-spec | Q1 해결 반영 — 레인 결정 사다리(디스크 증거 결정적; board/jev 불필요)가 해석 (ii)(플래그 = 정책 노브·킬스위치)로 판정: 카드 본문의 대우(자기 배정이 아닌 카드는 자가 배차 설정이 켜진 경우에만 임대 ⇒ OFF ⇒ 제네릭 암 무임대)와 리더 설계 문서 원칙 7 + §3 목표 표("generic arms only behind explicit eligibility filter (or removed)"). 해석 (i)의 닻("기존 테스트 무수정 GREEN")은 스폰 프롬프트 유래로 철회됨. decided_by: lane-17(tn1p4e 무인 위임, 리더 통보). REQ-CEP-001/002/004/006 재작성(OFF = 암 (b)/(b2) 무소유 카드 완전 휴면·조용한 스킵; ON = 자격 게이트 뒤 활성), AC-CEP-001 교체, 기존 무게이트 임대 고정 테스트의 의도적 갱신 정책(RED-first) plan.md §D에 명시. REQ/AC 총수 불변(7/7). |
| 0.2.1 | 2026-10-17 | manager-spec | plan-audit 1차 판정 FAIL(종합 0.83 — 스코어는 Tier M 통과선 0.80 초과, FAIL은 방화벽 동작: MP-9 순서 충돌 + codex 필수 백엔드 fail, receipt `rcpt-8b1c74f7dbede113c4974ecf`, 판정 파일 `.moai/reports/t1839/plan-audit.md`)의 결함 6건 수리. D1(치명): 게이트의 hold 마커 운반체를 상태 독립 `autoRankHoldMarked`(`todo_auto_rank.go:150`)로 교체 — `factoryQueuedHoldMarked`는 Queued 조건부라 picked 후보에 발화 불가; AC-CEP-002 (iii)에 picked+hold-marker 픽스처(no-lease + `refused hold-marker:` 라인) 확장. D2: 파일 범위에 `mcp_factory_card.go`(+`mcp_factory_msg_test.go`) 추가 + 양 표면 공유 거부 결과 경로 계획. D3(MP-9): AC-CEP-007의 RED+GREEN을 M1 RED-first 주기로 이동(M2 목록에서 제외), AC-CEP-003의 RED 셀을 M1 진입 관찰(기저 핀 회귀 수호 겸)·픽스처는 룰-페어링으로 재분류, E7 증거 사슬을 기준별 RED→GREEN 순서로 수정. D4: plan §C.1 상수 HEAD 핀을 진입 시 재계산 규칙으로 교체. D5: anchored `^…$` 셀렉터 전 명령 인라인 + 셀렉터 없는 internal/cli 전체 실행을 heavy run(슬롯 리스 `go-test-heavy` 규칙)으로 명시. D6: 롤아웃 노트(t1240 self-dispatch 배포는 `self_dispatch: true` 설정 필요 — 활성화 주체: 해당 배포 운영자/리더). Q1 (ii) 의미·설정 키·토큰 집합·Tier·감사 통과 항목은 불변. REQ/AC 총수 불변(7/7). |
| 0.2.2 | 2026-10-18 | manager-spec | Rule C 단일 수정 리비전(standing d-20261015T094733Z-9e76, t1793 선례 — 2차 감사 FAIL의 차단 결함 3건만 수리, receipt `rcpt-390fce8f7ecca3d52673bc8e`, 판정 파일 `.moai/reports/t1839/plan-audit.md` 2차 섹션). D7: AC-CEP-004를 M2 RED 목록·E7(iii)에서 제외하고 regression-guard(preserved-surface)로 재분류 — 증거는 M1 진입 현행 GREEN 핀 + E5 scoped-diff + M5(E6) preserved 재실행. D8: AC-CEP-003의 green path를 M4로 이동(룰-페어링 픽스처의 양 주장 — 적격 임대·실격 형제의 거부 라인 — 모두 M4에서 닫힘; 픽스처 분리 없음). D9: `held` 토큰을 폐쇄 집합에서 제거(도달 불가 — hold 상태는 조건 1의 입장 검사에서 항상 `not-admitted`로 선행 흡수) — 새 집합 `{not-admitted, spec-missing, hold-marker, predecessor-unmerged, serial-slot}`, 흡수 관계 명시. Q1 (ii) 의미와 1차 수리물(D1-D6) 불변. REQ/AC 총수 불변(7/7). |
| 0.3.0 | 2026-10-18 | manager-spec | In-place amendment(`## Amendments` §1 참조 — F1 판정: leader board d-20261018T021007Z-2620, 운영자 위임 + Jev 0.98; sync-audit F1의 codex 양방향 발산 증거 receipt `rcpt-b7f73391…`). REQ-CEP-003 조건 2를 fail-closed로 확장 — spec 식별자 없는 카드는 제네릭 암 (b)/(b2)에서 절대 임대 불가·항상 `spec-missing` 거부, priority==high 대리 제거. AC-CEP-003 재작성(무spec self-경로 거부 + 배정 정상 임대 + 룰-페어링 적격 형제의 spec 보유; M4 그대로). 클래스 축(additive dispatch-class axis — class×priority 회귀 테스트)은 rule B 후보로만 기록(frozen 9891, 미발급 — Out of Scope). 배정 경로(암 (a)·자기 배정 우선)·self-dispatch 기본 OFF·토큰 집합 불변. REQ/AC 총수 불변(7/7). |

## Amendments

### Amendment 1 — 2026-10-18, manager-spec (in-place, F1 판정)

- prior_completed_version: 0.2.2
- prior_completed_sha: 438a3c96e
- amendment_of: SPEC-CARD-ELIGIBLE-PULL-001 (self-referential — in-place)
- rationale: 본 amendment는 **첫 착지 전**의 SPEC에 대한 것이다 — 팩토리 카드 브랜치에서 run·sync가 완결되었으나 그 sync close(`438a3c96e`)는 브랜치-로컬이고 원격에는 아직 착지하지 않았다. sync-audit F1이 codex 양방향 발산 증거로 구현의 priority==high 대리가 요구문(무spec 카드 거부)과 어긋남을 발견했다(receipt `rcpt-b7f73391…`). F1 판정(leader board d-20261018T021007Z-2620, 운영자 위임 + Jev 0.98)이 방향을 확정: 요구문을 fail-closed로 — 무spec 카드는 제네릭 자가-인출 암 (b)/(b2)에서 절대 임대 불가·항상 `spec-missing` 거부, 대리 제거.
- scope: REQ-CEP-003 조건 2(fail-closed 전환)과 의존 문면(spec §A 요약 간구, plan.md §D 조건 문면, acceptance AC-CEP-002 (ii)·AC-CEP-003 재작성, spec §D 표 갱신). 배정 경로(암 (a)·자기 배정 우선)·self-dispatch 기본 OFF·거부 토큰 집합 불변. 클래스 축은 rule B 후보로만 기록(frozen 9891, 미발급). REQ/AC 총수 불변(7/7).

## §A Problem

bare `moai factory next`의 선택 패스(`internal/cli/factory_card.go:909-1103`)는 다섯 암으로 이루어진다.

| 암 | 대상 | 오늘의 검사 |
|----|------|-------------|
| (a) | 이 레인에 배정된 카드 | skip · selectionSkips · keep-set 큐 상태 · 자기 제외 직렬 슬롯 읽기 |
| (b-priority) | 이 레인 소유 번들의 다음 멤버 | 암 (b)의 조건 + 번들 소유 확인 |
| (b) | 무소유(ownerless) picked 레코드 row | skip · modeEligible · selectionSkips · 큐 항목 picked |
| (b2) | 레코드 row가 없는 큐 픽 카드 | modeEligible · hubWaitUnmerged |
| (c) | 큐 승격(queued → picked) | hold 마커 · blocked · selectionSkips/hub 대기 · 직렬 슬롯 |

암 (b)/(b2)는 위 표의 검사만 통과하면 **어떤** 무소유 픽 카드든 임대한다 — 그 카드가 어느 소비자를 위해 픽되었는지, Class C라면 설계 산출물(SPEC)을 갖췄는지, hold·선행 관계가 풀렸는지 묻는 자격 게이트가 없다. 2026-10-17 사고(카드 t1780 처리 중): bare next를 실행한 레인이 다른 소비자를 기다리던 무관한 운영자 픽 카드를 임대해 버렸다. 암 (a)의 배정 경로는 이 사고에 무관하다 — 문제는 "자기 소유가 아닌 카드를 줍는" 제네릭 암에만 있다.

운영자는 카드 t1839에서 수정 방향을 위임했다: (1) 배정 우선 경로는 무조건 보존, (2) 제네릭 암의 무소유 카드 임대 능력 자체가 기본 OFF인 self-dispatch 설정 키 뒤에 있다(킬스위치 — OFF면 암 (b)/(b2)는 무소유 카드를 전혀 임대하지 않는다), (3) ON에서만 암이 살아나며 그때 명시적 자격 술어(입장 확정 · spec 존재 — fail-closed, 무spec 카드는 항상 거부 · hold 아님 · 의존성 충족 · 직렬 슬롯 비어 있음) 전체를 통과해야 한다, (4) 거부는 산문이 아니라 안정 코드로 된 구조화 출력이며 카드에 흔적을 남기지 않는다. 이 해석(해석 (ii))은 Q1 해결로 확정되었다(§C D1).

## §B Requirements (GEARS)

- **REQ-CEP-001** (플래그 OFF = 제네릭 암 휴면) — **Where** `workflow.factory.self_dispatch`가 OFF(기본값)인 경우, the selection shall 제네릭 암 ((b) 무소유 picked row, (b2) 큐 픽 무레코드 카드)에서 자기 배정이 아닌 카드를 전혀 임대하지 않는다 — 두 암은 무소유 picked 카드를 완전히 건너뛴다(shall not lease). 이 스킵은 암 (c)의 hold-skip 패턴과 같은 **조용한 스킵**이다: 임대 없음, 기록 없음, 사유 라인 없음 — 자격 평가가 일어나지 않았으므로 실격 사유도 존재하지 않는다. 배정 경로(REQ-CEP-004)와 암 (c) 승격은 플래그와 무관하게 그대로 서비스한다.
- **REQ-CEP-002** (플래그 ON = 자격 게이트 뒤의 제네릭 암) — **Where** `workflow.factory.self_dispatch`가 ON인 경우, 제네릭 암 ((b)/(b2))은 활성화되며, 이 레인에 배정되지 않은 카드를 REQ-CEP-003의 자격 술어 **전체**를 통과할 때에만 임대한다(shall lease only through the full predicate). 실격 후보는 임대하지 않고 건너뛰며, 그 패스가 아무 카드도 임대하지 못하면 실격 사유는 REQ-CEP-005의 구조화 거부로 응답에 남는다.
- **REQ-CEP-003** (자격 술어 5조건) — **Where** 플래그가 ON이고 제네릭 암이 무소유 후보를 평가할 때, the selection shall 다음 다섯 조건이 모두 성립할 때만 후보를 통과시킨다:
  1. **admitted** — 큐 항목이 존재하고 picked 상태다(암 (b2)의 기존 큐 상태 검사의 일반화이며, 암 (b)의 row·큐 이중 검사가 같은 조건으로 명시적으로 적용된다);
  2. **spec 존재 — fail-closed** — spec 식별자를 레코드 row 또는 큐 힌트(`BacklogItem.SpecID`)에서 확인할 수 없는 카드는 유효 분류·우선순위와 무관하게 제네릭 암 (b)/(b2)를 통해 **절대 임대되지 않는다** — 항상 `spec-missing` 토큰으로 거부된다(shall not lease). priority==high 등 그 어떤 대리 조건도 이 거부를 면제할 수 없다(0.3.0 amendment — 종전 'Class C ⇒ spec 존재' 조건의 fail-closed 확장);
  3. **not held** — 기존 hold 술어(hold 상태·hold 마커)가 제네릭 암에도 명시적으로 적용된다;
  4. **dependencies met** — 선행(`after`) 미병합 대기와 hub 경로 공유 대기(`selectionSkips`/`hubWaitUnmerged`와 동일 술어)가 제네릭 암에 명시적으로 적용된다;
  5. **resources free** — serial 모드 후보는 직렬 슬롯이 비어 있을 때만 통과한다(기존 슬롯 상태 판정을 **읽기만** 한다 — 슬롯 의미는 바꾸지 않는다).
- **REQ-CEP-004** (배정·번들·승격 경로 보존) — **Where** 플래그가 ON이든 OFF든, 이 레인에 배정된 카드(암 (a)와 assigned rows)는 제네릭 암 자격 술어 및 플래그와 무관하게 기존 배정 경로의 조건 그대로 임대된다(shall lease through the assigned path unchanged). 이 레인 소유 번들의 다음 멤버(암 (b-priority)) 역시 기존 조건(번들 소유 확인 포함) 그대로 유지된다. 암 (c)의 큐 승격은 이 플래그의 적용 범위 밖이며 기존 조건 그대로다.
- **REQ-CEP-005** (구조화 거부) — **When** 플래그 ON에서 자격 미달 후보가 넘어가고 그 패스가 어떤 카드도 임대하지 못하면, the verb의 no-card 응답 shall 각 실격 사유를 안정 코드로 기계 판독 가능하게 명명한다 — 닫힌 토큰 집합 `not-admitted` · `spec-missing` · `hold-marker` · `predecessor-unmerged` · `serial-slot`(hold **상태**는 조건 1의 입장 검사에서 `not-admitted`로 흡수 발현된다 — 입장 = 큐 picked이며 hold 항목은 제네릭 인출에 입장하지 않는다; `hold-marker`는 picked+마커 조합에서 상태 독립 술어 경유로 발현한다), 기존 `refused <token>: <detail>` 거부 형태와 같은 문법(정확한 표면은 plan.md §D에 문서화). 토큰 집합은 닫혀 있다: 이 SPEC의 수정 없이는 새 사유 토큰이 추가되지 않는다.
- **REQ-CEP-006** (스킵·거부는 무흔적) — **When** 제네릭 암의 후보가 넘어가면 — 플래그 OFF의 조용한 스킵이든 ON의 게이트 거부든 — the selection shall 그 카드에 어떤 기록도 남기지 않는다: 레코드 row 생성 없음, claim 없음, 상태·소유자·버전 불변(wait-not-error 패턴, `factory_card.go:1006-1012`의 준수). 거부는 오류가 아니므로 exit 4(지명 거부)가 아니라 no-card 응답의 일부이며, 다음 패스가 같은 후보를 다시 평가할 수 있다.
- **REQ-CEP-007** (설정 키 단일·기본 OFF·템플릿 중립) — The config surface shall 게이트를 정확히 하나의 키 `workflow.factory.self_dispatch`(bool, 기본 false)로 노출한다 — repo `.moai/config/sections/workflow.yaml`과 배포 템플릿(`internal/template/templates`)의 workflow.yaml 양쪽에 같은 기본값으로 실리고 코드 기본값(`internal/config/defaults.go`)도 false다. 같은 의미의 두 번째 키를 만들지 않는다.

## §C Constraints and Decisions

- **D1 — 핵심 해석: 플래그의 의미(해결됨 — 해석 (ii), 킬스위치).** 카드 t1839 본문은 두 서술을 동시에 담아 양립 불가했다("제네릭 암은 플래그 ON일 때만 활성" 대 "flag OFF — 오늘과 동일"). decision-index.md Q1로 회부되었고, 레인 결정 사다리(디스크 증거 결정적 — board/jev 불필요)가 해석 (ii)로 판정했다: (1) 카드 본문의 대우(contrapositive) — "자기 배정이 아닌 카드는 … 통과할 때만, 자가 배차 설정이 켜진 경우에만 임대" ⇒ flag OFF ⇒ 제네릭 암 무임대; (2) 리더 설계 문서 원칙 7 + §3 목표 표 — "generic arms only behind explicit eligibility filter (or removed)". 해석 (i)의 닻("기존 테스트 무수정 GREEN")은 스폰 프롬프트 유래로 철회되었다. decided_by: lane-17(tn1p4e 무인 위임, 리더 통보). 따라서 OFF(기본)는 암 (b)/(b2)를 무소유 카드에 대해 완전히 휴면시키고(REQ-CEP-001), ON만이 자격 게이트 뒤에서 암을 살린다(REQ-CEP-002). Q1 항목은 운영자 반환 리뷰를 위해 decision-index.md에 그대로 보인다.
- **D2 — 키 선택.** `self_dispatch`/`selfDispatch`/`자가배차` 설정 키는 `internal/config`·workflow.yaml 전역에 존재하지 않는다(2026-10-17 grep 확인 — "self-dispatch"는 코드 주석·독트린 용어로만 있음). 따라서 카드 지시대로 `workflow.factory` 블록 아래 새 키 `self_dispatch`를 만든다(기존 workflow.* 블록 관례 — `branch_guard`·`factory_lease`·`slot_lease`와 같은 snake_case 블록 체계). sibling 키(`branch_guard`)의 반영 관례에 따라 repo workflow.yaml + 템플릿 workflow.yaml 양쪽에 동일 기본값 false로 실는다.
- **D3 — 토큰 집합.** 기존 닫힌 거부 토큰 집합(`factory_card.go:1213-1226`: unknown-card, dropped, held, owned, recorded, foreign-worktree, hold-marker, blocked, serial-slot, quota-hold, backend-skip, raced)에서 `hold-marker`/`serial-slot`을 재사용하고 `not-admitted`·`spec-missing`·`predecessor-unmerged`만 신규로 한다. 기존 `held` 토큰은 게이트의 집합에서 **제외**한다 — 사유 판정의 고정 순서(조건 1→5, 첫 사유만 출력)에서 hold 상태 후보는 조건 1(입장 = 큐 picked)에서 항상 선행 실격하므로 `held`는 도달 불가이고, hold 상태는 `not-admitted`로 흡수 발현된다. 카드 예시의 `serial-slot-busy`는 기존 `serial-slot` 토큰으로 통일한다(어휘 단일화 — 두 표면이 같은 사실을 다른 단어로 말하지 않게 한다).
- **D4 — wait-not-error 준수.** 게이트 실격도 OFF 스킵도 에러가 아니다. 암 (b2)의 기존 교훈(`factory_card.go:1006-1012`, 카드 t1533): 기록 먼저·거부 나중은 verb 전체를 실패시키고 무관한 준비 카드를 막는다. 게이트는 임대·기록 **이전**에 판정하고, 넘어간 후보는 기록도 claim도 없다(REQ-CEP-006).
- **D5 — 직렬 슬롯은 읽기 전용.** 자격 조건 5번("resources free")은 기존 슬롯 판정(`factorySerialSlotHeld`/`factorySerialInFlightExcluding`)을 게이트의 입력으로 **읽기만** 한다. 슬롯 의미·배제 규칙(t1407 운영자 판정, t1454/t1533 예외)은 한 건도 바꾸지 않는다.
- **D6 — 암 (b-priority) 무변경.** 이 레인 소유 번들의 다음 멤버는 "자기 소유 작업의 연장"이지 무소유 카드 줍기가 아니므로(REQ-TCI-018의 번들 소유 확인이 이미 붙어 있다) 플래그·자격 게이트 어느 쪽의 대상도 아니다. ON/OFF 모두에서 조건은 그대로다(REQ-CEP-004).
- **방법론.** `constitution.development_mode: tdd` — run phase는 RED 테스트 먼저. 기존 무게이트 임대 고정 테스트의 의도적 갱신 정책(RED-first 기대 뒤집기)은 plan.md §D에 명시한다.
- **Row-preserving.** 마이그레이션·행 삭제 없음. 이 SPEC은 큐 스키마와 레코드 row를 하나도 바꾸지 않는다(설정 키·선택 판정·응답 표면만 바꾼다).

## §D Acceptance Criteria

정규 AC 열거·증거 원장·시나리오는 `acceptance.md`(Tier M)에 있다. 요약:

| AC | REQ | 주장 | RED-now 셀 계획 | Green path |
|----|-----|------|------------------|------------|
| AC-CEP-001 | REQ-CEP-001 | 플래그 OFF(기본) — 암 (b)/(b2)는 무소유 picked 카드를 임대하지 않는다(조용한 스킵); 배정·자기 번들·암 (c)는 그대로 서비스 | RED: 현 트리는 게이트 없이 (b)/(b2) 임대 — 무임대 주장이 red(배정·번들·승격 대조 셀은 preserved) | M2→M3 |
| AC-CEP-002 | REQ-CEP-002, REQ-CEP-003 | 플래그 ON — 실격 무소유 picked 카드 미임대 + 정확한 토큰의 구조화 거부(실격 사유별 5셀) | RED: 현 트리는 게이트가 없어 임대 발생/사유 라인 부재 | M2→M3→M4 |
| AC-CEP-003 | REQ-CEP-002, REQ-CEP-003 | 플래그 ON — 무spec 카드는 self 경로에서 항상 `spec-missing` 거부(대리 불허); spec 보유 적격 카드는 (b)/(b2)로 임대(과차단 없음) | RED: M1 진입 관찰(기저 핀, 회귀 수호 겸) + 룰-페어링 픽스처(적격 형제는 spec 보유) | M4(양 주장 모두) |
| AC-CEP-004 | REQ-CEP-004 | 플래그 ON/OFF — 이 레인 배정 카드·자기 번들 멤버는 자격 술어·플래그와 무관하게 기존 경로로 임대; 암 (c) 승격도 플래그 범위 밖 | regression-guard: M1 진입 현행 GREEN 핀 + E5 무수정 범위 + M5 preserved 재실행 | M5(E5/E6) |
| AC-CEP-005 | REQ-CEP-006 | 스킵·거부는 row도 claim도 남기지 않음(상태·소유자·버전 불변; OFF 스킵과 ON 거부 모두) | RED: 무흔적 명세 | M3 |
| AC-CEP-006 | REQ-CEP-005 | 거부 표면이 기계 판독 가능(닫힌 토큰·안정 문법·exit 3·양면 동일) | RED: 문법 명세 | M4 |
| AC-CEP-007 | REQ-CEP-007 | 키 `workflow.factory.self_dispatch` 기본 false, repo·템플릿·코드 기본 일치 | RED: defaults 테스트 | M1 |

## Out of Scope

### Out of Scope — 스테이지 A: 소유의 생체성(liveness)·프로세스 정체성·epoch 펜싱

- 스테이지 A가 다루는 카드 소유의 생체성·프로세스 정체성·epoch 펜싱은 이 SPEC이 아니고 별도 스테이지가 소유한다. 이 SPEC의 게이트는 소유자가 "누구인가"가 아니라 후보가 "임대 가능한가"만 판정한다.

### Out of Scope — 스테이지 B: 내구 정체성·시도·런 멤버십 마이그레이션(retire-migrate)

- 내구 정체성(durable identity)·시도(attempts)·런 멤버십(run-membership) 마이그레이션과 retire-migrate는 이 SPEC의 범위가 아니다. 이 SPEC은 row-preserving이며 어떤 마이그레이션도 수반하지 않는다.

### Out of Scope — 스테이지 D(t1835): 자원 키 잠금으로 직렬 슬롯 대체

- 전역 직렬 슬롯을 자원 키 잠금(resource-keyed locks)으로 대체하는 작업은 카드 t1835가 소유한다. 이 SPEC의 "resources free" 조건은 기존 슬롯 상태를 자격 입력 **하나로 읽기만** 하며(REQ-CEP-003 조건 5, §C D5) 슬롯 의미·해제 규칙·예외(t1407/t1454/t1533)를 전혀 바꾸지 않는다.

### Out of Scope — t1829: 감사 게이트 백엔드 설정

- 감사 게이트 백엔드 설정(`audit.gates` 등)은 카드 t1829가 소유한다. 이 SPEC은 `workflow.factory` 키 하나만 추가하며 감사 설정을 건드리지 않는다.

### Out of Scope — 스테이지 E: 메타데이터 verb

- 카드 메타데이터 verb 군(스테이지 E)은 이 SPEC의 범위가 아니다. 이 SPEC은 `factory next`의 선택 판정과 no-card 응답 표면만 바꾼다.

### Out of Scope — 추가 dispatch-class 축(후보 기록, 미설계)

- 가법 dispatch-class 축(additive dispatch-class axis — class×priority 회귀 테스트)은 이 SPEC이 설계하지 않는다. 0.3.0 amendment(F1 판정)에서 **후보로만** 기록한다 — rule B(frozen 9891, 미발급)의 향후 SPEC 후보.
