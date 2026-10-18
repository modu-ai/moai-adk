---
id: SPEC-CARD-ELIGIBLE-PULL-001
title: "구현 계획 — 설정 키 배관, RED 우선 휴면·게이트 테스트, 자격 술어 적용, 구조화 거부 표면, 회귀 수호 배치"
version: "0.3.0"
created: 2026-10-17
updated: 2026-10-18
author: manager-spec
tier: M
---

# SPEC-CARD-ELIGIBLE-PULL-001 — plan

## §A Context

카드 t1839(사고 t1780 파생, Class C — 본 카드가 spec을 요구하는 설계 변경). bare `factory next`의 제네릭 암 (b)/(b2)(`internal/cli/factory_card.go:973-1017`)이 무소유 picked 카드를 관련성·자격 없이 임대한다. 본 플랜 페이스는 문서만 작성했다(트리 `93f4fc75c`, 브랜치 `WT-10-17-c` — run phase 진입 시 재확인). 개발 방식은 TDD(`quality.yaml` `constitution.development_mode: tdd`). 플래그의 의미는 Q1 해결로 **해석 (ii)(킬스위치)** 로 확정되었다(spec §C D1 — 레인 결정 사다리, decided_by: lane-17, 리더 통보): OFF(기본) = 암 (b)/(b2)가 무소유 카드를 전혀 임대하지 않음, ON = 자격 게이트 뒤에서 활성. "기존 테스트 무수정 GREEN" 닻은 스폰 프롬프트 유래로 철회되어 회귀 정책은 §D의 갱신 정책으로 대체되었다. 본 개정판(0.2.1)은 plan-audit 1차 판정 FAIL(종합 0.83 — 스코어는 통과선 0.80을 넘었으나 방화벽 동작: MP-9 순서 충돌 + codex 필수 백엔드 fail, receipt `rcpt-8b1c74f7dbede113c4974ecf`)의 결함 D1-D6을 수리한다(판정 파일 `.moai/reports/t1839/plan-audit.md`).

## §B Known Issues

| # | 위치 | 결함 | 근거 |
|---|------|------|------|
| K1 | `factory_card.go:973-994`(암 (b))·`:998-1018`(암 (b2)) | 무소유 picked 카드를 임대하는 관련성·자격 게이트가 없다 | 사고 t1780(2026-10-17): bare next가 다른 소비자 대기 카드를 임대 |
| K2 | 암 (b)/(b2)의 조건 블록 | hold 술어가 명시적으로 적용되지 않는다 — hold **상태**는 큐 picked 검사에 암묵적으로 걸리지만 hold **마커**(`factoryQueuedHoldMarked`)는 암 (c)·keep-set에만 있다 | `factory_card.go:1039` vs (b)/(b2) 부재 |
| K3 | 암 (b)/(b2) | Class C 카드의 spec 존재 검사가 없다 | 카드 t1839 요구 — `EffectiveCardClassification`(`internal/factory/classification.go:148`)·`BacklogItem.SpecID`는 이미 존재하나 이 암들은 읽지 않는다 |
| K4 | `factory_card.go:1775-1776` | no-card 응답이 사유를 전혀 담지 않는다(`no card is available` 한 줄) | 구조화 거부 부재 — 기존 `refused <token>: <detail>` 형태(`factoryNominateRefusal.Error()`, `:1237-1239`)는 지명 경로에만 있다 |
| K5 | `internal/config/types.go`(WorkflowConfig)·workflow.yaml | `workflow.factory` 블록과 self-dispatch 키가 없다(2026-10-17 grep — 기존 self-dispatch 관련 키 부재 확인) | 설정 키 신설 필요 |
| K6 | 기존 테스트 관점 | (b)/(b2)의 무게이트 임대를 고정하는 기존 테스트가 휴면 의미화와 충돌한다 | Q1 판정(해석 ii) — 갱신 정책(§D)으로 의도적 수정 |

## §C Pre-flight

1. 기저 핀 파생(재계산 규칙 — 이 문서는 상수 핀을 실지 않는다): run phase 진입 pre-flight에서 `git rev-parse --short HEAD`와 `git branch --show-current`를 실행해 **그 측정값을 이 카드의 기저 핀으로 기록**한다(§E.2 첫 행). 판정 기준은 진입 시 측정값이지 문서에 박제된 SHA가 아니다 — 박제 핀은 수리 커밋마다 낡는다(D4). 브랜치가 `WT-10-17-c`가 아니거나 측정 HEAD가 감사 시점 트리(`173871e8f`, 베이스 main@`93f4fc75c`) 이후의 이 문서의 수리 커밋 외의 예상 못한 이동을 보이면 정지·보고한다.
2. 기저 관찰(구현 커밋 이전): 무게이트 임대를 고정하는 테스트의 현행 GREEN이 갱신 정책의 RED 대조 기준이 된다. 일상 검증은 (a) 형태의 **anchored 셀렉터 인라인 명령**만 쓴다; (b)의 셀렉터 없는 `internal/cli` 패키지 전체 실행은 **heavy run**이다 — 리더의 부하 규칙하에서만 허용된다(`moai slot acquire --resource go-test-heavy --max-duration <bound>` → 실행 → `moai slot release`; gitflow-lane-protocol §8; 2026-10-17 기계 부하 50-60).
   - (a) scoped(기본): `go test ./internal/cli/ -run '^TestFactoryEligiblePullOffDormancy$' -count=1 -v` 형태 — M2가 짓는 신규 테스트는 파일 `factory_card_eligible_pull_test.go`, 최상위 가족 `TestFactoryEligiblePull*`로 명명하고, 인용되는 모든 명령 문자열의 `-run` 셀렉터는 모든 분기·서브테스트에서 앵커된 `^…$` 형태여야 한다(예: `-run '^TestFactoryEligiblePullOffDormancy$|^TestFactoryEligiblePullOnGate$'`). 기존 테스트의 기저 관찰도 §C.3에서 확정한 이름 그대로 개별 anchored 셀렉터로 지정한다.
   - (b) package-wide(예외 — 슬롯 리스 하에서만): `go test ./internal/cli/ -count=1` — 셀렉터가 없으므로 internal/cli 전체 스위트를 돈다.
   `go vet ./internal/cli/ ./internal/config/`; `golangci-lint run ./internal/cli/... ./internal/config/...` 기저 이슈 수 기록.
3. 테스트 파일 목록 확보 — (a) 갱신 대상: (b)/(b2) 무게이트 임대를 고정하는 기존 테스트, (b) 무수정 대상: 암 (a)/(b-priority)/(c) 경로 테스트. 커밋 전 `git diff`로 갱신 범위가 (a)에만 들어있는지 검증할 수 있게 파일·테스트 이름부터 고정한다.

## §D Constraints

- **파일 범위.** `internal/cli/factory_card.go`(+ 게이트 테스트 파일 `factory_card_eligible_pull_test.go`), `internal/cli/mcp_factory_card.go`(MCP no-card 표면 — `:187`의 고정 텍스트)와 MCP 표면 테스트가 사는 `internal/cli/mcp_factory_msg_test.go`, `internal/config/types.go`, `internal/config/defaults.go`, `internal/config/defaults_test.go`, `.moai/config/sections/workflow.yaml`, `internal/template/templates/.moai/config/sections/workflow.yaml`. 이 외 경로는 건드리지 않는다(스코프 규율). 기존 테스트 파일의 수정는 아래 갱신 정책이 허용하는 범위로 한정된다.
- **설정 키.** `workflow.factory.self_dispatch`(bool). `WorkflowConfig`에 `Factory` 블록 추가(기존 블록 관례 — `branch_guard`/`factory_lease` 준용), 코드 기본 false. repo·템플릿 yaml 양쪽에 주석과 함께 명시(false) — sibling 키(`branch_guard`)의 반영 관례.
- **롤아웃 주의(D6).** t1240 self-dispatch 런 중 레인의 bare-next 무소유 임대에 의존하는 기존 배포는 이 변경이 반영되는 시점에 `workflow.factory.self_dispatch: true`로 설정해야 그 면이 유지된다(기본 false는 제네릭 암 휴면). 활성화 주체·시점: 해당 배포의 운영자(팩토리 리더)가 변경 반영 시 설정한다 — 기본값 전환은 이 SPEC이 수행하지 않는다.
- **플래그 OFF = 제네릭 암 휴면(킬스위치, Q1 판정).** OFF(기본)에서 암 (b)/(b2)는 무소유 카드에 대해 어떤 임대도 하지 않는다. 스킵은 **조용한 스킵**이다 — 암 (c)의 hold-skip 패턴과 동일하게: 임대 없음, 기록 없음, 사유 라인 없음. 사유 라인은 플래그 ON에서 게이트가 후보를 **평가하고 실격시켰을 때만** 존재한다(평가가 없으면 사유도 없다). 배정 경로(암 (a))·자기 번들(암 (b-priority))·암 (c) 승격은 플래그와 무관하게 기존 조건 그대로다.
- **테스트 갱신 정책(RED-first).** 기존 (b)/(b2)의 무게이트 임대를 고정하던 테스트는 **의도적으로 갱신**한다: 기대를 뒤집은 직후 현 코드에서 RED를 관찰하고(이 관찰이 TDD의 RED 증거가 된다), 구현이 GREEN으로 만든다. 관측 없는 뒤집기(수정 후 바로 맞아들어가는 test-after)는 무효다. 무관 경로(암 (a)/(b-priority)/(c))의 기존 테스트는 무수정이며, 신규 시나리오는 신규 테스트 파일에 든다.
- **플래그 ON = 자격 게이트.** 5조건 술어(REQ-CEP-003)를 (b)/(b2) 평가에 적용하고 실격 사유(토큰+카드 id)를 수집해 넘긴다. 조건 2(spec 존재)는 **fail-closed**다 — spec 식별자가 없는 카드는 유효 분류·우선순위와 무관하게 제네릭 암에서 절대 임대되지 않고 항상 `spec-missing`으로 거부된다; priority==high 등 어떤 대리 조건도 이 거부를 면제할 수 없다(0.3.0 amendment, F1 — 대리 구현은 발산 결함이었다). hold **마커** 검사의 운반체는 상태 독립 함수 `autoRankHoldMarked`(`internal/cli/todo_auto_rank.go:150`)다 — `factoryQueuedHoldMarked`(`factory_card.go:1259-1266`)는 `it.State == Queued &&` 조건부라 picked 후보에는 절대 발화하지 않으므로 게이트의 마커 검사로 쓸 수 없다(D1). hold **상태** 검사는 큐 picked 검사(조건 1)가 담당한다 — hold 상태 후보는 입장에서 이미 실격하므로 항상 `not-admitted`로 발현한다(`held` 토큰은 없다, D9). 선행/hub 대기 술어는 기존 함수(`selectionSkips`/`hubWaitUnmerged`)를 재사용해 적용 지점만 명시적으로 바꾼다. 직렬 슬롯은 기존 판정을 읽기만 한다(D5).
- **공유 거부 결과 경로(D2).** 수집된 사유 라인은 선택 패스가 반환하는 결과에 실려 양 표면으로 흐른다 — cobra verb(`factory_card.go`)와 MCP 핸들러(`mcp_factory_card.go:187`가 자체 고정 텍스트를 갖는 지점)가 **하나의 공유 소스에서 같은 텍스트를 렌더**한다. 두 표면이 각자 문자열을 조립하면 AC-CEP-006(5)의 동일 텍스트 요구가 지켜질 수 없다.
- **Row-preserving·무마이그레이션.** 큐 스키마·레코드 row 구조 변경 없음, 행 삭제 없음.
- **그림자 측정 불필요(스테이지 C) 선언.** 이 SPEC(스테이지 C — 선택 거부·휴면)은 소유권을 바꾸지 않는다: OFF 스킵과 ON 게이트 실격 모두 row 생성·claim·상태 전이를 전혀 일으키지 않으므로(REQ-CEP-006) 임대·기록 경로의 기존 판정이 그대로 증언하고, 신규 동작을 그림자 실행으로 비교 측정할 측도가 없다. 그림자 측정은 상태를 바꾸는 스테이지(A/B/D)의 도구다.
- **구조화 거부의 정확한 표면(AC-CEP-006의 계약).** 플래그 ON에서 게이트가 후보를 넘겼고 그 패스가 아무 카드도 임대하지 못했을 때:

  ```
  no card is available
  refused <token>: <detail>
  ```

  - 첫 줄: 기존 no-card 줄(또는 기존 hold 안내 줄)을 그대로 유지 — 기존 소비자의 파서를 깨지 않는다.
  - 둘째 줄부터: 게이트 실격 후보 **마다 최대 한 줄**, 후보당 **첫 실격 사유 하나**(사유 판정 순서는 조건 1→5 순: not-admitted > spec-missing > hold-marker > predecessor-unmerged > serial-slot — 결정론적 출력을 위한 고정 순서. hold **상태** 후보는 조건 1(입장 = 큐 picked)에서 이미 실격하므로 항상 `not-admitted`로 발현한다 — `held` 토큰은 존재하지 않는다(D9); `hold-marker`는 picked+마커 조합에서 `autoRankHoldMarked` 경유로 발현한다).
  - `<token>` ∈ 닫힌 집합 `{not-admitted, spec-missing, hold-marker, predecessor-unmerged, serial-slot}`(기존 토큰 재사용 — spec §C D3; hold 상태는 `not-admitted`에 흡수 발현된다 — D9). `<detail>`는 카드 id를 포함한 한 줄(개행 없음).
  - 기계 판독 문법: `^refused (not-admitted|spec-missing|hold-marker|predecessor-unmerged|serial-slot): ` — 이 정규식이 표면의 계약이고, AC-CEP-006의 파서 셀은 이 문법으로 0줄 아님을 증명한다.
  - exit code는 **3**(no-card)을 유지한다 — 지명 거부의 4와 오류의 1이 아니다(REQ-CEP-006). cobra verb 출력과 MCP 도구 출력은 같은 텍스트를 낸다(`factoryNominateRefusal`의 양면 계약 준용).
  - 플래그 OFF의 스킵은 이 표면에 아무 라인도 추가하지 않는다(조용한 스킵).
- **주석 언어.** 영어(`language.yaml` 기본 — 주석 밀도는 주변 코드에 맞춘다).

## §E Self-Verification

run phase 완료 보고는 다음을 verbatim으로 싣는다(귀속 3요소 — 명령·출력·트리 SHA):

- **E1** AC 매트릭스(AC-CEP-001..007) 각 행 GREEN — acceptance.md §B의 셀과 1:1.
- **E2** 범위 스위트(anchored 셀렉터 인라인): `go test ./internal/cli/ -run '^TestFactoryEligiblePull…$' -count=1 -v` — §C.2 (a) 형태, M2에서 확정한 개별 `^…$` 셀렉터를 명령 문자열 안에 그대로 싣는다. `go test ./internal/config/ -run '^TestWorkflowFactorySelfDispatchDefault$' -count=1 -v`(M1 defaults 테스트명). 셀렉터 없는 `internal/cli` 전체 실행은 heavy run — 슬롯 리스 하에서만(§C.2 (b)).
- **E3** `go vet ./internal/cli/ ./internal/config/` exit 0.
- **E4** `golangci-lint run ./internal/cli/... ./internal/config/...` — 신규 이슈 0(기저 대비).
- **E5** 갱신 범위 검증: `git diff <pre-implementation SHA>..HEAD --stat -- <테스트 파일>` — 수정은 §C.3의 갱신 대상(a)에만, 무관 경로(b)는 변경 0. 갱신된 각 테스트는 RED 관찰 기록(§E.2)과 쌍이다.
- **E6** OFF 휴면 확인: AC-CEP-001 재실행 GREEN(무소유 무임대 + 배정·번들·승격 대조 셀 GREEN).
- **E7** RED 증거 사슬(기준별 RED→GREEN 순서 — D3/D7/D8): (i) AC-CEP-007 — M1 안의 RED-first 주기(defaults 테스트 RED를 키 부재 트리에서 관찰 → 구현 → GREEN); (ii) AC-CEP-003 — M1 **진입 시점**(구현 전 트리) 관찰로 기저 커밋에 핀 — 무게이트 암이 적격 후보를 임대하는 현행 사실(회귀 수호 핀을 겸함); (iii) 나머지 RED 전부(AC-CEP-001/002/005/006 + 갱신 테스트의 뒤집힌 기대) — M2에서 관찰; (iv) AC-CEP-004 — regression-guard(모든 셀이 현행 GREEN preserved, 도달 시 green-at-arrival): RED 관찰을 요구하지 않는다 — 증거는 M1 진입 현행 GREEN 핀 + E5 갱신 범위 scoped-diff(무관 경로 무수정) + M5의 E6 preserved 재실행(D7). 각 관찰의 명령·출력·종료 코드·트리 SHA를 §E.2에 verbatim으로 기록한다.

## §F Milestones

결정 되돌림 가능성 순서 — 인터페이스 결정(키·표면 문법)이 앞, 기계적 단계가 뒤.

### M1 (High) — 설정 키·플래그 배관(RED-first 주기) + AC-CEP-003 기저 관찰

진입 시: (1) §C.1 기저 핀 파생·기록; (2) **AC-CEP-003의 RED-now 셀 관찰** — 구현 전 트리에서 무게이트 암 (b)/(b2)이 적격 후보를 임대하는 현행 사실을 관찰해 기저 커밋 핀과 함께 §E.2에 기록한다(이 관찰이 이 기준의 회귀 수호 핀을 겸한다 — 게이트 도입 뒤에는 이 전제를 RED로 관찰할 수 없으므로 지금이 유일한 관찰 시점이다). 본체(RED-first): `defaults_test.go`에 기대를 먼저 작성(`Workflow.Factory.SelfDispatch` 기본 false — BranchGuard 패턴 준용, 테스트명 `TestWorkflowFactorySelfDispatchDefault`) → 키 부재 트리에서 **RED 관찰** → `internal/config/types.go`에 `Factory` 블록(`SelfDispatch bool`, yaml `factory.self_dispatch`) 추가, `defaults.go` 기본 false, repo·템플릿 workflow.yaml에 주석 포함 명시 false → GREEN. AC-CEP-007의 RED와 GREEN이 **모두 이 마일스톤 안에서** 닫힌다(D3 — M2의 RED 목록에는 없다). — 키 이름·블록 위치가 리뷰에서 가장 바뀔 수 있는 결정이므로 첫 마일스톤.

### M2 (High) — RED 테스트 우선 랜딩(신규 + 기존 기대 뒤집기)

신규 게이트 시나리오 RED 테스트(신규 파일 `factory_card_eligible_pull_test.go`): AC-CEP-001의 OFF 휴면(무임대 + 배정·번들·승격 대조 셀), AC-CEP-002의 실격 사유별 5셀((iii)은 hold 상태(`not-admitted` 흡수 발현)와 picked+hold-marker(`hold-marker` — 상태 독립 `autoRankHoldMarked` 경유, no-lease + `refused hold-marker:` 라인) 픽스처 포함), AC-CEP-005 무흔적·exit 3, AC-CEP-006 문법 파서 셀(AC-CEP-004는 regression-guard이므로 이 RED 목록에 없다 — D7). AC-CEP-003(ON 적격 임대)의 픽스처도 여기서 짓되 **룰-페어링**(verification-completeness §2)으로 — 같은 픽스처 안에 적격 후보(임대 주장)와 실격 형제(거부 라인 주장)를 함께 두어 과차단·과통과 양방향 뮤턴트에 저항하게 한다; 이 기준의 RED-now 관찰은 M1 진입에서 이미 끝났으므로 M2가 관찰을 반복하지 않는다(AC-CEP-007도 M1에서 닫혀 이 목록에 없다 — D3). 동시에 기존 (b)/(b2) 무게이트 임대 고정 테스트를 갱신 정책(§D)대로 기대 뒤집기 — 뒤집은 직후 현 코드에서 RED 관찰. **이 시점에 관찰 가능한 RED**(위 전부)를 verbatim으로 §E.2에 기록한 뒤에만 M3로 진입한다.

### M3 (High) — 플래그 판정·휴면 분기·자격 술어 적용(암 (b)/(b2))

선택 패스에 플래그 판정 추가: OFF면 암 (b)/(b2)가 무소유 후보를 조용히 건너뛴다(기록·claim·사유 없음), ON이면 5조건 술어를 적용하고 실격 사유(토큰+카드 id)를 수집해 넘긴다. hold 술어·선행/hub 대기 술어는 기존 함수 재사용, 직렬 슬롯은 읽기 전용(D5). 암 (a)/(b-priority)/(c)는 건드리지 않는다(D6).

### M4 (Medium) — no-card 응답의 구조화 거부 라인

수집된 사유를 §D 계약 문법으로 no-card 응답에 렌더(첫 줄 유지, 후보당 한 줄, exit 3, cobra·MCP 양면 동일, OFF는 라인 없음). AC-CEP-006의 green path이자 AC-CEP-003의 green path — 룰-페어링 픽스처의 **양 주장**(적격 임대와 실격 형제의 거부 라인)이 모두 이 마일스톤에서 닫힌다(D8; 픽스처를 분리하지 않는다).

### M5 (Medium) — 회귀 수호 배치

무관 경로 기존 테스트 무수정 GREEN, 갱신된 (b)/(b2) 테스트 GREEN, E3/E4(vet·lint), 커버리지 측정(휴면·게이트 신규 경로 ≥85%), E5 갱신 범위 `git diff` 확인, §E.2 증거 완결.

## §G Anti-Patterns

- **OFF에서 게이트 없이 그대로 임대하는 구현** — D1 위반(해석 (i)의 부활). OFF는 킬스위치다: 무소유 카드에 암 (b)/(b2)는 어떤 임대도 하지 않는다.
- **RED 관찰 없는 기대 뒤집기** — 갱신 정책 위반. 기존 테스트의 기대를 뒤집을 때는 뒤집은 직후 현 코드에서 RED를 관찰해야 한다(test-after·무관측 뒤집기는 E7 RED 증거를 만들 수 없다).
- **무관 경로 테스트 수정** — 갱신 대상은 (b)/(b2) 무게이트 임대 고정 테스트만이다(암 (a)/(b-priority)/(c) 테스트는 무수정).
- **거부 시 row 먼저 기록** — D4 위반(t1533의 실패 재연). 게이트·휴면 판정은 기록·임대 이전에 한다.
- **exit 4 남용** — 게이트 실격도 휴면 스킵도 지명 거부가 아니다. no-card(3) 안의 사유 라인이다(REQ-CEP-006).
- **직렬 슬롯 의미 변경** — 슬롯 판정·예외(t1407/t1454/t1533)는 읽기만 한다(D5). 슬롯 대체는 t1835.
- **두 번째 설정 키** — 유사 키(예: `self_dispatch_pull`)를 추가하지 않는다(REQ-CEP-007).
- **OFF 스킵에 사유 라인 출력** — 조용한 스킵이다(암 (c) hold-skip 패턴). 사유 라인은 ON의 게이트 평가 실격에만 존재한다.
- **사유 라인에 개행 포함 detail** — 문법 계약 위반(한 줄).

## §H Cross-References

- SPEC-TODO-AUTO-PICK-001 — no-card/hold 응답 형태와 닫힌 토큰 집합의 원류(REQ-TAU-005/007).
- SPEC-TODO-CLASSIFY-DISPATCH-001 — 직렬 슬롯 배타·승격 암의 원류(REQ-TCD-007/-008). 본 SPEC은 이 판정을 읽기만 한다.
- SPEC-TODO-HOLD-STATE-001 — 양성 열거 원칙(REQ-THS-012)과 조용한 hold-skip 패턴(REQ-THS-007) — OFF 휴면 스킵이 따르는 형태.
- t1240 self-dispatch 흡수면(REQ-SD-*) — 레인의 bare next 임대 허용면. 본 SPEC은 그 임대를 플래그 뒤로 옮기고 자격 조건을 얹는다(t1780 사고의 재발 방지).
- 사고 t1780(2026-10-17)·카드 t1839(본 카드)·t1835(스테이지 D)·t1829(감사 게이트 백엔드) — 범위 분리는 spec.md Out of Scope.
- decision-index.md Q1 — 플래그 의미 판정 기록(해석 ii, lane-resolved).
- 롤아웃 — 기존 t1240 self-dispatch 배포의 활성화(`self_dispatch: true`, 설정 주체·시점)는 §D 롤아웃 주의(D6) 참조.
- `.claude/rules/local/gitflow-lane-protocol.md` §6 — 레인은 `factory next`로만 카드를 임대한다는 운영 규칙(본 SPEC의 게이트가 그 경로에 붙는다).
