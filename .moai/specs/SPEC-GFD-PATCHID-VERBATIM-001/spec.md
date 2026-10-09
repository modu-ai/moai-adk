---
id: SPEC-GFD-PATCHID-VERBATIM-001
title: "착지 판정 patch-id 비교의 공백 충실화 — --verbatim 전환과 미지원 git fail-closed"
version: "0.1.2"
status: completed
created: 2026-10-07
updated: 2026-10-07
author: GOOS
priority: P1
phase: "v3.2.0"
module: "internal/cli/worktree"
lifecycle: spec-anchored
tags: "landing-predicate, patch-id, verbatim, whitespace-divergence, fail-closed, worktree-sweep, data-destruction"
tier: M
related_specs: [SPEC-GITHUB-FLOW-DEFAULT-001]
---

# SPEC-GFD-PATCHID-VERBATIM-001 — 착지 판정 patch-id 비교의 공백 충실화

## HISTORY

| 버전 | 날짜 | 작성 | 변경 |
|---|---|---|---|
| 0.1.0 | 2026-10-07 | manager-spec (카드 t1561) | plan-phase 산출물 최초 작성 (Tier M 3종 + progress.md + decision-index.md). REQ 5개·AC 6개 |
| 0.1.1 | 2026-10-07 | manager-spec (카드 t1561) | plan-audit 1회차(0.83, FAIL) 수리. D1: 세션 종료 제2 결함 지점(`internal/cli/session_worktree.go` ②팔 `git cherry` 공백 정규화 조기 착지)을 범위로 편입 — §A.1·§A.2·§B.1·§B.2 정정, REQ-GPV-001~003 문언 확장, AC-GPV-007 신설, 대상 파일 4개. D2: AC-GPV-003 을 실측 기준선 셀과 함께 regression-guard 로 재분류. D3: §B 1·3행 종료코드 기재. D4: AC-GPV-005/006/007 관문 몸통 측정(식별자 plan 시점 확정·케이스 수 고정·공집행 실패 문구). D5-D7 선택 수리 동반. REQ 5개·AC 7개 |
| 0.1.2 | 2026-10-07 | manager-spec (카드 t1561) | plan-audit 2회차(0.88, FAIL — D1-D7 전부 RESOLVED 확인) 수리. D8: AC-GPV-007 §D 본문(Given-When-Then + 공집행 실패 조항 + M1.4·M2.3 green path) 신설 — 0.1.1 치환에서 본문이 누락된 미완결 수리, §F 게이트·plan §E 행렬을 ..007 로 확장. D9: plan §F-M3.1 검증 배치를 두 패키지 테스트·두 경로 vet/lint·4파일 diff-stat 으로 갱신. D10: acceptance §B 헤더 날짜 문언 정정. D11: plan §B8 pathspec 4파일·§C 현 상태 주석·M1 판정 5+RED 분리. REQ 5개·AC 7개 불변 |

## §A 배경

### A.1 결함 요지

`internal/cli/worktree/landing_predicate.go` 의 계층 2(누적 patch-id)는 `git patch-id --stable` 로 패치를 재는데, 이 모드는 **공백·들여쓰기·행 번호 차이를 정규화해 없던 것으로 만든다**. 그래서 원격에 착지한 squash 커밋이 카드 tip 의 누적 변경과 공백만 다르면 계층 2 가 "같은 변경이 이미 착지했다"고 잘못 답하고, `moai worktree sweep` 이 트리를 폐기(DISPOSE)한다 — 원격이 운반하지 않는 tip 내용물의 **유일 사본이 삭제되는** 데이터 파괴다.

이 결함의 생존 지점은 둘이다. **제2 지점은 세션 종료 착지 술어다.** `internal/cli/session_worktree.go` 의 `gitBranchLandedReal` 은 ②팔에서 `git cherry`(890행)의 커밋별 patch-id 동치로 "+"가 없으면 조기 "착지"를 답하고(902-904행) 공유 술어에 도달하지 않는다 — cherry 는 같은 공백 정규화를 쓰므로, 공백 갈린 단일 커밋 카드는 수리 뒤에도 세션 출구에서 유일 사본 파괴로 이어진다(감사 1회차 프로브: `git cherry master card` → `- a82b4346…`, "-" = 동치 = 착지 판정, git 2.54.0). 이 SPEC 은 두 지점을 함께 막는다.

### A.2 현황 (실측 — 좌표는 `§F 측정 핀`)

- **해시 수준 전제가 실측됐다.** git 2.54.0 (Apple Git-157) 에서 공백만 다른 두 패치("line 7 card" vs "line 7     card")가 동일 patch-id `4087e767dea83d923c696a63bf0ad97f5780403d` 로 붕괴된다(RED 기록 프로브 출력).
- **술어 수준 RED 가 트리에 있다.** `TestLandingPredicateWhitespaceDivergenceKeepsTree` (`internal/cli/worktree/landing_predicate_test.go:554`)가 `sweep: landed="yes" verdict="DISPOSE" reason="", want preserve` 로 실패한다. 이 테스트는 회귀 관문이며 run 단계 GREEN 의 판정식이다.
- **결함 코드 위치.** `landingPatchIDs()` 가 `git patch-id --stable` 을 고정 인자로 실행한다(`landing_predicate.go:93`). 계층 2 는 세 계층 술어(조상 관계 → 누적 patch-id → PR 병합 상태)의 두 번째다.
- **호출자는 셋이다.** `landedBeyondAncestry`(`moai worktree done` + `moai worktree sweep`)과 세션 종료 정리다. 세션 종료 정리는 공유 술어를 통째로 쓰지 않고 **자체 3팔 술어**(`gitBranchLandedReal`, `internal/cli/session_worktree.go:857`)로 판정한다 — ①조상 관계, ②커밋별 patch-id 동치(`git cherry`, 공유 술어에 도달하기 전에 답한다), ③공유 누적 patch-id(`worktree.LandedByPatchID`). REQ-WSS-304(소유 SPEC: SPEC-WEB-SETTINGS-SAVE-001)가 이 출구 경로를 네트워크 호출 없이 묶는다. 공유 쪽 파일 헤더의 계약: **답할 수 없는 계층은 결코 "착지"가 아니며, 확정 불가 착지는 트리를 보존한다.**

### A.3 이 SPEC 의 성격

계층 2 의 **비교 모드 하나**를 공백 충실(`--verbatim`, git 2.42+)로 바꾸고, 미지원 git 에서는 기존 계약대로 fail-closed 로 흡수하는 수리다. 세 계층 술어의 재설계도, 호출자 계약의 변경도 아니다. `quality.yaml` `development_mode: tdd` — RED 가 이미 트리에 있으므로 run 단계는 GREEN + 회귀 보강의 모양을 가진다.

## §B 범위

### B.1 In Scope

| 항목 | 내용 | 마일스톤 |
|---|---|---|
| 계층 2 비교 모드 전환 | 카드 누적 diff 와 ref log 스트림 양쪽을 공백 충실 모드(`git patch-id --verbatim`)로 계산. 누적-vs-커밋별 비교 형상은 유지 | M1 |
| 기능 감지 | 공백 충실 모드 지원 여부를 패키지 시임(seam)으로 감지. 테스트가 로컬 git 버전과 무관하게 양쪽을 재현 | M1 |
| 미지원 git fail-closed | 미지원이 감지되면 계층 2 는 "답할 수 없음" 오류 → 호출자 보존, 계층 3 은 계속 판정 | M1 |
| 세션 종료 ②팔 수리 | `internal/cli/session_worktree.go` 의 `git cherry` 팔을 공백 충실 동치로 복속 — 공백 갈림 패치로 조기 "착지"를 답하는 경로 차단. ①·③팔과 호출 계약은 유지 | M1 |
| 회귀 보강 테스트 | 미지원 시임 케이스 + 다중 커밋 누적 공백 갈림 케이스 + 세션 종료 경로(단일 커밋 공백 갈림 카드 → 세션 출구 preserve) 신설(기존 테스트 단언 무변경) | M2 |

### B.2 Out of Scope

### Out of Scope — 계층 1·계층 3 과 세션 종료 ①·③팔의 재설계

- 조상 관계(`merge-base --is-ancestor`)와 PR 병합 상태(`gh`) 계층의 동작·계약·호출 형상, 그리고 세션 종료 술어의 ①팔(조상)·③팔(공유 누적 patch-id 호출부)의 재설계는 이 SPEC 이 바꾸지 않는다. 공백 갈림 상황에서 계층 3 이 계속 판정하는 것은 이 SPEC 이 지키는 전제다. 단, 세션 종료 ②팔(`git cherry` 동치)의 공백 충실화는 §B.1 의 In Scope 다 — 그 팔은 이 SPEC 이 막는 결함의 제2 생존 지점이기 때문이다.

### Out of Scope — 공백 정규화 비교의 보존

- 미지원 git 에서 `--stable` 로 조용히 하락하는 경로는 **의도적으로 만들지 않는다** — 그것은 결함의 재현 경로다(REQ-GPV-003). 가용성 회복을 위한 별도 전략이 필요해지면 후속 SPEC 의 몫이다.

### Out of Scope — 커밋 상한·성능 재조정

- `landingPatchIDCommitCap`(500, design D-5)과 gh 타임아웃(10s) 값은 이 SPEC 에서 다시 잴 대상이 아니다. `--verbatim` 전환은 비교 모드만 바꾸고 상한 의미(초과 시 cannot answer)는 그대로다.

### Out of Scope — RED 관문 테스트의 단언 변경

- `TestLandingPredicateWhitespaceDivergenceKeepsTree` 의 단언문은 회귀 관문으로 고정한다. run 단계에서 이 테스트의 주석 조정은 허용하되 단언 변경은 금지다.

### Out of Scope — git 버전 바닥값 강제

- 사용자 환경에 git 업그레이드를 요구하거나 minimum-git 경고 표면을 만들지 않는다. 미지원 환경은 fail-closed 로 흡수되며, 계층 3 이 판정을 이어받는다.

## §C 요구사항

> 요구사항은 GEARS 형식이며 각 항은 `acceptance.md` 의 `AC-GPV-00n` 이 `maps REQ-…` 로 덮는다. Tier M 상한은 REQ 16개·AC 16개이고 이 SPEC 은 REQ 5개·AC 7개를 둔다.

### REQ-GPV-001 — 공백 갈림 패치의 오확인 금지

**When** 누적 patch-id 계층이 카드 tip 의 누적 변경과 통합 ref 위 커밋의 변경을 비교할 때, 두 패치가 공백·들여쓰기·행 번호만 다르고 내용이 다르면 계층은 그 커밋을 카드 변경의 착지로 읽지 않아야 한다(shall not). **When** 세션 종료 착지 술어가 커밋별 패치 동치로 착지를 판정할 때(`git cherry` 팔), 같은 공백 갈림을 동치로 읽어 조기에 "착지"를 답해서는 안 된다(shall not). 공백 차이를 없던 것으로 만드는 patch-id 정규화가 어느 착지 경로의 판정 근거로도 남아 있어서는 안 된다(shall not).

### REQ-GPV-002 — 공백 충실 비교의 양면 적용

**While** 로컬 git 이 공백 충실 patch-id 모드(`--verbatim`)를 지원하는 동안, 계층 2 는 카드 누적 diff 와 ref log 스트림 양쪽을 같은 모드로 계산해야 하고(shall), 세션 종료 술어의 커밋별 동치 팔도 같은 모드의 patch-id 로 동치를 판정해야 한다(shall). 누적-vs-커밋별 비교 형상과 커밋 상한 초과 시의 cannot answer 동작은 바뀌지 않아야 하고(shall not), 커밋별 동치로 이미 확인해 온 착지 케이스(개별 커밋이 upstream 에 각자 존재하는 카드)는 계속 확인돼야 한다(shall). 바이트가 같은 패치는 전환 뒤에도 계속 일치해야 한다(shall).

### REQ-GPV-003 — 미지원 git 의 fail-closed

**When** 로컬 git 이 공백 충실 모드를 지원하지 않는다고 감지되면, 계층 2 와 세션 종료 술어의 커밋별 동치 팔은 답할 수 없음을 오류로 보고해야 하고(shall), 호출자(`done`·`sweep`·세션 종료 정리)는 트리를 보존해야 하며(shall), `done`·`sweep` 의 계층 3(PR 병합 상태)은 계속 판정해야 한다(shall). 미지원 환경에서 cherry 팔이 공백 정규화 동치만으로 독립적인 "착지" 답변을 내는 경로도 없어야 한다(shall not). 이때 공백 정규화 비교로 조용히 하락하는 경로는 없어야 한다(shall not).

### REQ-GPV-004 — 기존 착지 시맨틱의 보존

**While** 계층 2 의 비교 모드가 바뀌는 동안, 바이트 동일 누적 패치의 squash 착지 확인(F1·F2), 이후 변경 관측(커밋 패치 비교로 계층 2 가 답하는 형상), 문맥 갈림 행의 계층 3 위임(F5), 커밋 상한 행, gh 실패 fail-closed 행의 시맨틱은 보존돼야 하고(shall), 기존 착지 판정 테스트는 단언 변경 없이 통과해야 한다(shall).

### REQ-GPV-005 — 지원 감지의 시임

**Where** 공백 충실 모드 지원 여부가 패키지 수준에서 감지될 때, 감지 결과는 테스트가 대체할 수 있는 패키지 시임으로 돼야 하고(shall), 테스트는 로컬 git 버전에 의존하지 않고 지원·미지원 양쪽을 재현해야 한다(shall).

## §D 결정 요약

각 결정의 미해결 근거와 기각된 대안은 `decision-index.md` Q1·Q2 가, 실행 절차는 `plan.md` 가 운반한다.

| ID | 입장 |
|---|---|
| P1 | 계층 2 비교를 `git patch-id --verbatim` 로 전환하고, 양쪽 스트림에 같은 모드를 쓴다. 지원 감지는 패키지 시임 하나로 둔다 (decision-index Q1, DEFAULT-APPLIED) |
| P2 | 미지원 git 에서 계층 2 는 cannot answer 오류 → 호출자 preserve, 계층 3 이 판정을 이어받는다. `--stable` 하락 경로는 두지 않는다 (decision-index Q2, DEFAULT-APPLIED) |
| P3 | RED 관문 테스트의 단언은 그대로 두고, 보강 테스트(미지원 시임·다중 커밋 누적 공백 갈림)를 별도로 더한다 — RED 기록의 잔여위험(단일 시나리오 RED)을 M2 에서 소화 |
| P4 | 이 SPEC 은 Tier M — 착지 판정은 `done`·`sweep`·세션 종료 세 호출 지점이 같이 읽는 공유 계약이므로 검증 계층(`acceptance.md`)을 별도로 둔다 |

## §E 인접 SPEC 과의 관계

- **SPEC-GITHUB-FLOW-DEFAULT-001** (`status: in-progress`) — REQ-GFD-002 가 창시한 착지 판정의 후속 수리다. `related_specs` 로 참조만 하고 `depends_on` 이 아니다: 수리 대상 코드는 이미 이 트리에 착지돼 있고(`cad44a751`), 전임 SPEC 의 진행 상태와 무관하게 본 수리는 독립 착지 가능하다. 전임 SPEC 이 아직 completed 가 아니므로 그 HISTORY·Amendments 행은 본 SPEC 의 권한 앵커가 아니다 — decision-index Q1·Q2 가 FOUNDER 로 남는 이유다.

## §F 측정 핀

RED 측정은 카드 트리 `cad44a75163b6f0f056551354ef8008fe799090f`(브랜치 `WT-landing-patchid`, `origin/main` `cad44a751` 기저)와 git 2.54.0 (Apple Git-157)에 핀한다. patch-id 공백 붕괴 실측값(`4087e767dea83d923c696a63bf0ad97f5780403d`)과 RED 테스트 실패 출력의 원문은 `.moai/reports/t1561/red-reproduction.md` 가 운반한다. 이 SPEC 의 어떤 판정도 핀이 움직인 뒤 재인용하지 않는다 — 인용하려면 다시 잰다. plan 단계 관측(동일 트리, 관측 시각 병기): 2026-10-06T21:36Z — `landing_predicate.go` 내 `stable` 2히트(90·93행), `verbatim` 0히트, 신규 보강 테스트 식별자 0히트. 2026-10-07 — RED 테스트 재실행 종료코드 1, `grep -c '\-\-stable'` → `2` 종료코드 0, `grep -rc 'verbatim'` → `…landing_predicate.go:0` 종료코드 1, 세션 신설 관문 식별자 0히트 종료코드 1, `go vet ./internal/cli ./internal/cli/worktree/` 종료코드 0(빈 출력), `golangci-lint run ./internal/cli ./internal/cli/worktree/` → `0 issues.` 종료코드 0. 세션 종료 ②팔 조기 착지 프로브(`git cherry master card` → `- a82b4346…`, git 2.54.0 한정)와 `--stable`/`--verbatim` 독립 검증 프로브는 plan-audit 1회차 감사자의 관측이다(`.moai/reports/t1561/plan-audit-iter1.md` 증거 표 8·9행) — 저작자 재관측이 아니므로 그 등급으로만 인용한다.

---

🗿 MoAI
