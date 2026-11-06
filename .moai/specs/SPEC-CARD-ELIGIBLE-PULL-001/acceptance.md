---
id: SPEC-CARD-ELIGIBLE-PULL-001
title: "인수 기준 — 플래그 OFF 휴면, 자격 게이트 5조건, 구조화 거부"
version: "0.3.0"
created: 2026-10-17
updated: 2026-10-18
author: manager-spec
tier: M
---

# SPEC-CARD-ELIGIBLE-PULL-001 — acceptance

## §A Method

모든 기준은 두 셀(verification-completeness §2)으로 채택한다. 이 SPEC은 플랜 페이스 문서만 작성하므로(카드 t1839: 코드·테스트·설정 변경 금지) 이 문서는 **관찰이 아니라 셀 계획**을 기록한다 — 관찰되지 않은 출력을 적지 않는다.

- **신규 계층**(AC-CEP-001..007): RED-now 셀은 run phase가 **먼저** 작성하는 RED 테스트가 된다(카드 배달 제약: RED first). AC-CEP-001(플래그 OFF 휴면)도 이제 정당한 RED 기준이다 — 현 트리는 게이트 없이 (b)/(b2)가 무소유 카드를 임대하므로 무임대 주장이 현 코드에서 빨갛다. 단 AC-CEP-001의 배정·자기 번들·암 (c) 대조 셀과 AC-CEP-004의 대조 셀은 preserved(무변경 경로 — 현행 GREEN 유지)로 취급한다. 각 RED 셀은 "현 트리에서 왜 빨간가"를 서술하며, RED 관찰(명령·출력·종료 코드·트리 SHA)은 run phase가 verbatim으로 `progress.md` §E.2에 기록한다. 기준별 RED→GREEN 순서(D3/D7/D8): AC-CEP-007은 M1 안의 RED-first 주기(키 부재 트리에서 defaults RED 관찰 → 구현 → GREEN); AC-CEP-003의 RED-now 관찰은 M1 진입 시점(구현 전, 기저 커밋 핀) — 게이트 도입 뒤에는 그 전제를 RED로 관찰할 수 없다(회귀 수호 핀을 겸함); 나머지 RED(001/002/005/006)는 M2에서 관찰한다. AC-CEP-004는 regression-guard/preserved-surface 기준이다(모든 셀이 현행 GREEN — 도달 시 green-at-arrival): RED 관찰을 요구하지 않으며, 증거는 M1 진입 현행 GREEN 핀 + E5 갱신 범위 scoped-diff + M5의 E6 preserved 재실행이다(D7).
- **기존 테스트 갱신 정책(plan.md §D와 동일 문면).** 기존 (b)/(b2)의 무게이트 임대를 고정하던 테스트는 **의도적으로 갱신**한다 — RED-first: 기대를 뒤집어 현 코드에서 RED를 관찰한 뒤 구현이 GREEN으로 만든다. 무관 경로(암 (a)/(b-priority)/(c))의 기존 테스트는 무수정이다.
- 기준 이행의 검증 주체는 다음 명령군이다 — 모든 `-run` 셀렉터는 명령 문자열 안에 모든 분기에서 앵커된 `^…$` 형태로 인라인된다(plan §C.2 (a)): `go test ./internal/cli/ -run '^TestFactoryEligiblePull…$' -count=1 -v`, `go test ./internal/config/ -run '^TestWorkflowFactorySelfDispatchDefault$' -count=1 -v`, `go vet ./internal/cli/ ./internal/config/`, `golangci-lint run ./internal/cli/... ./internal/config/...`. 셀렉터 없는 `internal/cli` 패키지 전체 실행은 heavy run이므로 리더의 슬롯 리스 규칙(`moai slot acquire --resource go-test-heavy`; gitflow-lane-protocol §8, 2026-10-17 기계 부하 50-60)하에서만 허용된다.

## §B 인수 기준(Given-When-Then)

### AC-CEP-001 — 플래그 OFF(기본): 제네릭 암 휴면, 나머지 경로 지속

maps REQ-CEP-001

- **Given** `workflow.factory.self_dispatch`가 설정에 없다(기본 false). 큐·레코드에 (i) 무소유 picked 카드, 그리고 대조적으로 (ii) 이 레인에 배정된 카드, (iii) 이 레인 소유 번들의 다음 멤버, (iv) queued 카드가 각각 존재한다. **When** bare `moai factory next`를 실행한다. **Then** (i) 무소유 picked 카드를 임대하지 **않는다** — 암 (b)/(b2)는 무소유 카드에 완전히 무반응이며 기록·claim·사유 라인이 없다(조용한 스킵); (ii) 배정 카드는 암 (a) 경로로 임대된다; (iii) 번들 멤버는 암 (b-priority)로 임대된다; (iv) queued 카드는 암 (c) 승격으로 임대된다(플래그와 무관).
- RED-now 셀 계획: (i)이 RED — 현 트리는 게이트 없이 (b)/(b2)가 무소유 picked 카드를 임대하므로 무임대 주장이 현 코드에서 실패한다. (ii)~(iv)는 preserved 대조 셀(무변경 경로의 현행 GREEN 유지).
- Green path: M2(RED 확정) → M3(휴면 분기).

### AC-CEP-002 — 플래그 ON, 실격 미임대 + 정확한 사유 코드(실격 사유별 5셀)

maps REQ-CEP-002, REQ-CEP-003

- **Given** 플래그 ON. 무소유 picked 후보가 다음 중 **하나만** 결려 있다: (i) 큐 항목이 없거나 picked가 아니다(not-admitted) — hold **상태** 후보가 여기 속한다: 입장 = 큐 picked이므로 hold 상태는 조건 1에서 실격해 항상 `not-admitted`로 발현한다(`held` 토큰은 폐쇄 집합에 없다 — 흡수 관계의 명시, D9); (ii) spec 식별자가 row·큐 힌트 어디에도 없다(유효 분류·우선순위 무관 — Class C 포함, priority==high도 면제 없음; spec-missing — 0.3.0 amendment의 fail-closed); (iii) **picked 상태 + hold 마커** — 카드 텍스트가 hold 마커로 시작, 상태 독립 술어 `autoRankHoldMarked`(`internal/cli/todo_auto_rank.go:150`) 경유(`hold-marker`; `factoryQueuedHoldMarked`는 Queued 조건부라 이 picked+마커 조합을 잡지 못한다 — 픽스처는 반드시 picked+마커 조합으로 짠다); (iv) `after` 선행이 로컬 미병합이거나 hub 경로 공유자가 미병합이다(predecessor-unmerged); (v) serial 모드인데 직렬 슬롯이 다른 카드에 붙잡혀 있다(serial-slot). **When** bare `moai factory next`를 실행한다. **Then** 그 후보를 임대하지 않고, 응답에 해당 토큰의 `refused` 라인이 남는다((i)의 hold-상태 픽스처는 `refused not-admitted:` 라인을, (iii)의 picked+마커 픽스처는 `refused hold-marker:` 라인을 주장한다).
- RED-now 셀 계획(각 셀): RED 테스트가 먼저 작성된다 — 현 트리는 게이트가 없어 (i)~(v) 셀에서 임대가 일어나거나(예: hold 마커·Class C 무spec 후보가 (b)/(b2)로 임대), 임대가 안 일어나도 사유 라인이 없어 라인 주장이 실패한다.
- Green path: M2(RED 확정) → M3(게이트) → M4(표면).

### AC-CEP-003 — 플래그 ON, 무spec self-경로 거부 + 적격 임대 유지(0.3.0 amendment)

maps REQ-CEP-002, REQ-CEP-003

- **Given** 플래그 ON. (a) 무소유 picked 후보가 spec 식별자를 row·큐 힌트 어디에도 갖지 않는다(유효 분류·우선순위 무관 — priority==high도 포함). (b) 대조: 카드가 이 레인에 배정되어 있다(assigned). (c) 룰-페어링 픽스처의 적격 형제는 **spec 식별자를 보유한다**(spec 있는 Class C 또는 Class C 아님 + 나머지 4조건 충족). **When** bare `factory next`를 실행한다. **Then** (a) 무spec 후보는 제네릭 암 (b)/(b2)를 통해 **절대 임대되지 않고** 항상 `refused spec-missing:` 라인으로 거부된다(fail-closed); (b) 배정 카드는 기존 배정 경로 그대로 정상 임대된다(leader-assignment·자기 배정 우선 무영향); (c) spec 보유 적격 형제는 (b)/(b2) 경유로 정상 임대된다 — 게이트는 적격 카드를 막지 않는다.
- 셀 계획(0.3.0): (a)는 fail-closed의 직접 셀이자 F1 수정 검증 셀 — priority==high 대리가 남아 있으면 (a)가 RED로 붉어진다(대리 제거 확인). (a)의 RED는 amendment 구현 전 트리에서 관찰한다; 종전 RED 관찰(M1 진입의 무게이트 임대 사실, 기저 커밋 핀 — D3의 regression-guard)은 그대로 유효하다. (c)의 픽스처는 이제 적격 형제가 spec을 반드시 보유한다. 룰-페어링 유지 — 적격 임대 주장과 실격 형제의 거부 라인 주장이 같은 픽스처에 함께 산다.
- Green path: M4 — 픽스처의 **양 주장**(적격 임대와 실격 형제의 거부 라인)이 모두 표면이 존재하는 M4에서 닫힌다(D8; 룰-페어링 유지 — 픽스처를 분리하지 않는다; 마일스톤 불변).

### AC-CEP-004 — 배정·번들·승격 경로 면역(플래그 양상 모두)

maps REQ-CEP-004

- **Given** 플래그 ON과 OFF 각각에서: 카드가 **이 레인에 배정**되어 있다(assigned row, owner = 이 레인) — 자격 술어에 실패하는 상태까지 포함(예: spec 없는 Class C 배정 카드, hold 마커가 없고 keep-set이 허용하는 범위). 자기 소유 번들의 다음 멤버와 queued 카드도 각각 대조 셀로 존재한다. **When** bare `factory next`를 실행한다. **Then** 배정 카드는 기존 암 (a)의 조건 그대로, 번들 멤버는 (b-priority)의 조건 그대로, queued 카드는 암 (c) 승격으로 임대된다 — 어느 것도 제네릭 암 자격 술어·플래그의 영향을 받지 않는다.
- 셀 분류(D7 — regression-guard/preserved-surface): 이 기준의 모든 셀은 현행 GREEN preserved(도달 시 green-at-arrival)이므로 RED 관찰을 요구하지 않는다(two-cell discipline: green-at-arrival ⇒ regression-guard, release-blocking RED가 아님). 증거 = M1 진입 시점 현행 GREEN 핀(§E.2) + E5 갱신 범위 scoped-diff(무관 경로 무수정) + M5의 E6 preserved 재실행 GREEN.
- Green path: M5(E5/E6).

### AC-CEP-005 — 스킵·거부는 무흔적

maps REQ-CEP-006

- **Given** 플래그 OFF(조용한 스킵 경로)와 ON(게이트 거부 경로) 각각에서 실격 후보가 존재한다. **When** bare `factory next`가 그 후보를 넘긴다. **Then** 레코드 row 신규 생성 없음, claim 없음, 후보의 상태·소유자·버전·리스 불변 — 통과 전후 스냅샷 비교로 증명한다. exit는 no-card(3)이지 오류(1)·지명 거부(4)가 아니다.
- RED-now 셀 계획: 무흔적·exit 3 명세 셀 — 현 트리의 (b2)는 이미 wait-not-error를 지키지만, 플래그 휴면·게이트 실격 경로가 없어 이 기준의 대상 상황이 성립하지 않는다.
- Green path: M3.

### AC-CEP-006 — 거부 표면의 기계 판독

maps REQ-CEP-005

- **Given** 플래그 ON, 게이트가 후보를 넘겼고 그 패스가 아무 카드도 임대하지 못했다. **When** 응답 텍스트를 plan.md §D의 문법으로 파싱한다. **Then** (1) 첫 줄은 기존 `no card is available`(또는 기존 hold 안내 줄)을 유지하고, (2) 게이트 실격 후보마다 정확히 한 줄의 `refused <token>: <detail>`이 따르고(detail은 카드 id 포함, 한 줄), (3) 모든 토큰이 닫힌 집합 `not-admitted|spec-missing|hold-marker|predecessor-unmerged|serial-slot`에 속하며(hold 상태는 `not-admitted`로 흡수 발현 — D9), (4) exit code는 3을 유지하고, (5) cobra verb 출력과 MCP 도구 출력이 **하나의 공유 거부 결과 소스에서 렌더된 같은 텍스트**를 낸다(plan §D 공유 거부 결과 경로 — `mcp_factory_card.go:187`의 자체 고정 텍스트 지점 포함). 플래그 OFF의 스킵은 사유 라인을 내지 않는다(REQ-CEP-001의 조용한 스킵 — 평가가 없으므로 사유도 없다).
- RED-now 셀 계획: 문법 파싱 셀 — 현 트리의 no-card 응답(`factory_card.go:1775-1776`)에는 사유 라인이 없어 파서가 0줄을 고르고 기준이 실패한다.
- Green path: M4.

### AC-CEP-007 — 설정 키 단일·기본 false·템플릿 중립

maps REQ-CEP-007

- **Given** 아무것도 설정하지 않은 배포. **When** 설정을 로드한다. **Then** `workflow.factory.self_dispatch`는 false로 읽히고, repo `.moai/config/sections/workflow.yaml`과 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml` 양쪽에 키가 같은 기본값 false로 존재하며, `internal/config/defaults.go`의 코드 기본값도 false다. 같은 의미의 다른 키는 존재하지 않는다.
- RED-now 셀 계획(D3 — M1 안의 RED-first 주기): defaults 테스트(`TestWorkflowFactorySelfDispatchDefault`, `internal/config/defaults_test.go`의 `Workflow.BranchGuard.Enabled` 패턴 준용)를 **먼저** 작성해 키 부재 트리에서 RED를 관찰한 뒤(컴파일/조회 실패), 구현이 GREEN으로 만든다 — 이 기준의 RED와 GREEN이 모두 M1 안에서 닫힌다(M2의 RED 목록에 없다).
- Green path: M1(RED-first 주기 내).

## §C 품질 게이트·완료 정의(DoD)

- 위 7개 AC의 이원 셀(RED-now + green path; 대조 경로는 preserved)이 모두 채워지고 `progress.md` §E.2에 verbatim 증거(명령·출력·종료 코드·트리 SHA)로 기록된다.
- 테스트 갱신 범위: 무관 경로(암 (a)/(b-priority)/(c))의 기존 테스트는 무수정 — 커밋 전 `git diff`로 확인. 갱신된 테스트는 (b)/(b2) 무게이트 임대 고정 테스트뿐이며 각각 RED 관찰 기록을 수반한다(plan.md §D 갱신 정책).
- `go vet ./internal/cli/ ./internal/config/` exit 0; `golangci-lint run ./internal/cli/... ./internal/config/...` 신규 이슈 0(기저와 대비).
- 게이트·휴면 신규 경로(플래그 판정·자격 판정·사유 수집·표면 렌더)의 커버리지가 프로젝트 기준(85%)을 충족한다.
- 회귀 수호: 무관 경로 테스트 무수정 GREEN + 갱신된 (b)/(b2) 테스트 GREEN(AC-CEP-001/002).
