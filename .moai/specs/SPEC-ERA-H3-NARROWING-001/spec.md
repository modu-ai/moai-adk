---
id: SPEC-ERA-H3-NARROWING-001
title: "H-3 시대 분류 술어 축소 — 진행 중 SPEC의 V3R5 오분류 차단"
version: "0.1.0"
status: draft
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "era, classification, grandfather, lint, spec-audit"
tier: M
---

# SPEC-ERA-H3-NARROWING-001 — H-3 시대 분류 술어 축소

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-31 | manager-spec | 최초 작성. 카드 t382. 모든 수치는 트리 `9328a5242`에서 `./bin/moai`로 실측 |

## 1. 배경

`internal/spec/era.go`의 `ClassifyEra`는 SPEC을 다섯 시대(V2.x / V3R2-R4 / V3R5 / V3R6 / unclassified) 중 하나로 분류한다. 이 분류는 **처벌 여부를 가른다**: V2.x·V3R2-R4·V3R5는 `Era.EraFinal()`이 참이 되어 grandfather 보호를 받고, V3R6만이 lifecycle drift 검출과 구조 게이트의 적용을 받는다.

휴리스틱 체인은 **첫 매치 승리**다. H-3의 술어는 다음 한 줄이다 (`era.go:150`, 트리 `9328a5242`):

```go
if hasSyncSection && syncSHA == "" {
    return EraV3R5, "H-3 (§E.2 present, sync_commit_sha missing)"
}
```

여기서 `hasSyncSection`은 progress.md에 리터럴 `§E.2`가 있는지를 볼 뿐이다(변수명은 역사적 오칭이며 코드 주석이 이를 명시한다). 그런데 `§E.2` 골격 제목은 **plan-phase에 manager-spec이 항상 찍는다** — `§E.1`~`§E.4` 네 개 자리표시자 제목을 한꺼번에 만드는 것이 규약이다. 그리고 `sync_commit_sha`는 sync가 닫히기 전에는 비어 있다.

따라서 **모든 plan/run 단계의 현대 SPEC이 H-3에 걸린다.** 걸리는 순간 `return`이므로, `created` 날짜를 읽어 V3R6이라 판정했을 H-5까지 도달하지 못한다. 시대 분류가 "이 SPEC이 언제 쓰였는가"가 아니라 "이 SPEC이 아직 안 닫혔는가"를 재고 있는 셈이다.

실측 (트리 `9328a5242`, `./bin/moai spec audit`): 전체 714건 중 V3R5로 분류된 23건 가운데 **22건은 `created >= 2026-04-01`** 이고, **`created < 2026-04-01`인 것은 0건**이다. 즉 오늘의 V3R5 버킷에는 진짜로 오래된 SPEC이 한 건도 없다. 나머지 1건은 아래 §1.2에서 따로 다룬다.

### 1.1 결과 — 무엇이 실제로 무력화되는가

`internal/spec/lint.go`의 `applyEraDemotion`은 grandfather SPEC에 대해 두 가지를 한다.

1. `eraDemotableCodes`(`MissingExclusions`, `FrontmatterInvalid`) 두 코드에 한해 **ERROR를 warning으로 강등**하고 `Advisory: true`를 붙인다.
2. 그 SPEC의 **나머지 모든 warning에도 `Advisory: true`를 붙인다.** `Report.HasErrors()`는 `Strict && warning && !Advisory`일 때만 승격하므로, `--strict`가 그 SPEC 위에서는 아무것도 승격시키지 못한다.

**축소된 정확한 주장은 이것이다: 구조 게이트 두 개가 advisory로 내려가고, 그 SPEC에 한해 `--strict`가 무력화된다.** "모든 lint 오류가 강등된다"는 더 넓은 주장은 **거짓이다.** `applyEraDemotion`의 `switch`는 첫 갈래에서 `eraDemotableCodes[f.Code]`를 요구하므로 그 밖의 ERROR는 손대지 않고 통과시킨다. 이 사실은 다른 레인이 뮤테이션으로 반증했다 — `status: bogus-status-value`를 심자 `StatusValueInvalid`가 advisory 표식 없이 **ERROR** 심각도로 나왔다. 우리는 그 반증을 물려받아 좁은 주장만 싣는다.

### 1.2 자기차폐 사례 — SPEC-V3R5-INIT-WIZARD-EXPANSION-001

V3R5 23건 중 유일하게 `created`를 읽을 수 없는 건이다. frontmatter가 거부된 snake_case 별칭(`created_at:` / `updated_at:` / `labels:`)을 쓰고 있어 YAML 디코더가 통째로 버린다(`spec-frontmatter-schema.md` § Rejected Snake_Case Aliases). 실제 작성일은 본문 HISTORY에 적힌 **2026-05-22** — 임계값 이후다.

이 한 건이 중요한 이유는 결함이 **자기차폐적**이기 때문이다. H-3 오분류가 감싸 주는 그 잘못된 frontmatter가, 동시에 H-5가 이 SPEC을 구제하지 못하게 만드는 원인이다. 실측(M11)에서 이 SPEC은 `FrontmatterInvalid` **7건**을 내는데 전부 `[grandfathered era — downgraded to warning]` 꼬리표를 달고 advisory로 앉아 있다. 즉 스키마 위반을 신고할 수 있었던 유일한 신호가, 그 스키마 위반 때문에 얻은 보호에 의해 잠긴다.

`created_at:`을 쓰는 SPEC은 카탈로그에 46건 있다(글롭 `*/spec.md` 기준; 재귀 grep은 `_archive/` 1건을 더해 47을 낸다). 다만 그중 **V3R5 버킷에 있는 것은 이 1건뿐**이므로, 이 결함의 수정 범위에서 `created_at` 인구가 갖는 영향은 딱 이 한 건에 국한된다.

### 1.3 오늘 카탈로그에서 관측되는 델타 — 정직한 서술

이 수정이 오늘 무엇을 바꾸는지는 축마다 다르다. 세 축을 나눠 적는다.

- **lint 축 — 오늘의 델타는 0이다.** 재분류 대상 22건 중 non-terminal은 13건인데(INIT-WIZARD는 §3에서 보듯 재분류되지 않는다), 이 13건은 강등 대상 ERROR를 **애초에 하나도 갖고 있지 않다**(M11: `MovingRefUnpinned` 3건과 `StatusGitConsistency` 1건뿐이며 둘 다 발화 지점에서 이미 `Advisory: true`다). `--strict` rc는 수정 전후 모두 0이다. 이 축의 이득은 **잠재적**이다 — 앞으로 이 13건에 구조 결함이 생기면 그때는 게이트가 걸린다.
- **audit MUST-FIX 축 — 오늘의 델타는 0이다.** 유일한 V3R6 drift 차원인 `SyncStatusDrift`는 `sync_commit_sha`가 채워져 있을 것을 요구하는데, 이 23건은 정의상 그 값이 비어 있다(그게 H-3의 조건이다). 따라서 재분류돼도 MUST-FIX는 생기지 않는다.
- **drift 축 — 오늘 관측되는 유일한 델타다.** `./bin/moai spec drift`에서 23건 중 22건이 현재 `era-exempt`로 빠진다(1건은 이미 `terminal-exempt`). 수정 후 이들은 git 대조 분류에 들어간다. 이것은 이득인 동시에 **위험**이다 — 새 DRIFT 행이 생길 수 있고, 그것이 진짜인지 오탐인지는 표본으로 확인해야 한다(§3 AC-EH3-007).

이 절을 명시하는 이유는, 이 수정을 "게이트를 되살린다"고만 소개하면 오늘의 초록을 이득의 증거로 읽게 되기 때문이다. 오늘 되살아나는 게이트는 없다. 되살아나는 것은 **앞으로 걸릴 능력**이다.

## 2. 요구사항 (GEARS)

- **REQ-EH3-001** (unwanted): `ClassifyEra`는 modern-era 신호를 가진 SPEC을 H-3 경로로 `EraV3R5`로 분류해서는 안 된다(shall not). modern-era 신호는 H-5가 쓰는 것과 **동일한 술어**로 정의한다 — `matchesModernPhase(FrontmatterPhase)` 또는 `isAfterModernThreshold(FrontmatterCreated)`.
- **REQ-EH3-002** (capability-gate): Where a SPEC이 modern-era 신호를 **하나도** 갖지 않는 경우, `ClassifyEra`는 그 SPEC에 대해 수정 이전과 바이트 동일한 결과 — 시대와 rationale 문자열 모두 — 를 반환해야 한다(shall).
- **REQ-EH3-003** (event-driven): When H-3이 유예되면, `ClassifyEra`는 H-4 → H-4-legacy → H-5 → H-6 체인을 **원래 순서 그대로** 이어서 평가해야 한다(shall). 어떤 휴리스틱도 재배치하지 않는다.
- **REQ-EH3-004** (unwanted): 이 수정은 `eraDemotableCodes` 집합, `applyEraDemotion`의 동작, `Era.EraFinal()` / `Era.IsModern()`의 정의를 변경해서는 안 된다(shall not). 변경 지점은 H-3 술어 한 곳이다.
- **REQ-EH3-005** (ubiquitous): `internal/spec/era_test.go`는 회귀 가드를 실어야 한다(shall) — 어떤 `EraSignals` 조합에 대해서도 「H-5 술어가 참인데 결과가 `EraV3R5`」인 상태가 성립하지 않음을 단언하는 불변식 테스트.
- **REQ-EH3-006** (ubiquitous): `.claude/rules/local/lifecycle-sync-gate.md`의 휴리스틱 표 H-3 행은 새 술어를 반영해야 한다(shall). 이 파일은 코드와 함께 SSOT로 인용되므로 갱신하지 않으면 두 문서가 서로 반대를 지시한다.

## 3. 인수 기준

Tier M이므로 인수 기준 본문은 `acceptance.md`가 정본이다. 여기서는 방향만 요약한다 — 판정의 절반만 재는 기준 집합이 되지 않도록, **양방향**을 명시한다.

| AC | 재는 방향 | 요지 |
|----|----------|------|
| AC-EH3-001 | 오분류 중단 | 날짜 신호가 있는 H-3 모양 → V3R6 |
| AC-EH3-002 | 오분류 중단 | phase 신호가 있는 H-3 모양 → V3R6 |
| AC-EH3-003 | **보호 유지** | 임계 이전 `created` → 여전히 V3R5/H-3 |
| AC-EH3-004 | **보호 유지** | 신호 전무(INIT-WIZARD 모양) → 여전히 V3R5/H-3, H-6 낙하 없음 |
| AC-EH3-005 | 코퍼스 귀속 | 시대가 바뀐 SPEC을 **한 건씩 이름과 근거로** 귀속, grandfathered 총계 전후 대조 |
| AC-EH3-006 | 게이트 복원(주입 결함) | 재분류된 SPEC 사본에 구조 결함을 심었을 때 `--strict` rc 0→1 |
| AC-EH3-007 | **비용 측정** | drift 축에서 새로 생긴 행이 진짜인지 표본 확인 |
| AC-EH3-008 | 회귀 가드 | 가드가 known-bad 입력(뮤테이션)에서 실패함을 시연 |

## 4. 범위 밖

### Out of Scope — 형식이 깨진 frontmatter의 수리

- `SPEC-V3R5-INIT-WIZARD-EXPANSION-001`의 `created_at:` / `updated_at:` / `labels:` 별칭을 정본 필드로 고치는 일. 이 SPEC은 그 수리를 **하지 않는다.** §1.2에서 보였듯 그 수리는 별도의 소유자(비-전이 frontmatter 정정 — `spec-frontmatter-schema.md` § Non-transition frontmatter corrections에 따라 manager-spec, 오케스트레이터 재위임 경유)를 갖는다.
- 나머지 45건의 `created_at:` 보유 SPEC 일괄 수리.

### Out of Scope — era 엔진에 별칭 관용을 추가하는 일

- `EraSignals`가 `created_at:`을 `created:`의 대체 신호로 읽게 만드는 변경. §5에서 옵션 D로 검토하고 근거를 적어 기각했다.

### Out of Scope — 형제 카드가 소유한 영역

- **t371** — spec-lint CI 체크아웃 깊이와 18건 finding 재분류. 의존 관계는 §6에 적었으나 그 재분류 자체는 이 SPEC이 하지 않는다.
- **t376** — 새 전이 규칙 추가.
- **t380** — 경계 픽스처.
- 이 셋 모두 `internal/spec` 안에서 살아 있는 형제 카드다. 이 SPEC은 `internal/spec/era.go`의 H-3 술어 한 곳과 그 테스트, 그리고 SSOT 문서 한 곳만 만진다.

### Out of Scope — 휴리스틱 재배치

- H-1~H-6의 **순서**를 바꾸는 일. §5 옵션 A에서 근거를 적어 기각했다.
- `modernEraThreshold` 상수값(`2026-04-01`) 변경.

## 5. 설계 선택 — 네 후보와 기각 근거

### 옵션 A — H-5(날짜/phase 검사)를 H-3 앞으로 옮긴다

H-5는 H-3보다 앞서면 **H-4보다도 앞선다.** 그러면 임계 이후 SPEC은 정밀한 3-단계 술어(`§E.2 + §E.4 + sync_commit_sha`)로 잡혔을 경우에도 rationale이 `"H-5 (modern phase or created date)"`로 나온다. 코드는 이 정밀도를 명시적으로 가치로 선언한다 — H-4-legacy 주석은 "explicit predicate is a stronger signal than the H-5 date heuristic"라는 이유만으로 그 갈래를 유지한다. 순서를 뒤집으면 그 가치를 카탈로그 전체에서 버린다. **기각.**

### 옵션 B — H-3 술어에 「현대 마커 부재」를 추가한다 (예: `§E.4` 없음을 요구)

측정으로 기각된다. V3R5 23건 중 **21건이 이미 `§E.4`를 갖고 있다**(M9). plan-phase 골격이 `§E.1`~`§E.4`를 한꺼번에 찍기 때문이다. 즉 `§E.4` 유무는 판별력이 사실상 없고, 23건 중 2건만 구제한다. 게다가 기존 테스트 `"H-3 §E.2 + §E.4 present but sync_commit_sha empty → V3R5"`를 정면으로 깨뜨린다 — 그 테스트가 표현하는 의도(sync 미완)는 옳고 이 옵션이 틀렸다. **기각.**

### 옵션 C — H-3을 modern-era 신호가 있을 때만 유예한다 ← **채택**

```go
if hasSyncSection && syncSHA == "" && !hasModernEraSignal(signals) {
    return EraV3R5, "H-3 (§E.2 present, sync_commit_sha missing)"
}
```

`hasModernEraSignal`은 H-5가 이미 쓰는 술어를 그대로 함수로 뽑은 것이다 — `matchesModernPhase(phase) || isAfterModernThreshold(created)`. 두 곳이 같은 술어를 쓰는 것이 이 설계의 핵심 불변식이다: **수정 후에는 「H-3 적격이면서 동시에 H-5가 현대라 부를 SPEC」이 존재할 수 없다.** 이것이 REQ-EH3-005의 가드가 기계적으로 검사하는 명제다.

채택 근거 셋:

1. **유예가 양성 신호에 게이트된다.** 신호가 없으면 조건이 거짓이 되어 H-3이 그대로 발화한다. 따라서 **이 설계는 새로운 H-6 낙하를 원리상 만들 수 없다.** 판정 불가능한 SPEC은 보수적으로 보호를 유지한다 — 이것이 grandfather 절의 취지와 같은 방향이다.
2. **기존 테스트를 하나도 깨지 않는다.** `era_test.go`의 H-3 케이스들은 `FrontmatterCreated` / `FrontmatterPhase`를 비워 두고 있어 유예 조건이 거짓이 된다. 이는 우연한 통과가 아니라 REQ-EH3-002가 요구하는 성질이 이미 그 테스트로 표현돼 있다는 뜻이다.
3. **rationale 정밀도와 휴리스틱 순서를 보존한다.** 옵션 A가 버리는 두 가지를 그대로 지킨다.

**날짜만 쓰지 않고 phase까지 포함한 이유.** 날짜만 쓰면 「임계 이전 `created` + `phase: "v3.0.0"`」인 SPEC이 여전히 grandfather로 남아, 유예 조건과 H-5 술어가 어긋난다. 그러면 위 불변식이 성립하지 않고 가드도 쓸 수 없다. 둘을 붙이면 술어가 정확히 일치한다.

**이 선택이 지는 잔여 위험.** 임계 이전에 쓰였는데 `phase:`에 `v3.0` 계열 라벨을 단 SPEC은 grandfather 보호를 잃는다. 오늘 카탈로그에서 이 교집합은 **0건**이다 — V3R5 23건 중 `created < 2026-04-01`인 것이 0건이므로(M4) 구성상 공집합이다. 다만 "오늘 0건"이 "원리상 0건"은 아니므로, AC-EH3-005가 시대가 바뀐 SPEC을 **한 건씩** 귀속시켜 이 벡터를 매번 재측정하게 한다. 원리상의 탈출구도 이미 존재한다 — frontmatter `era:` 필드(H-override)가 어떤 휴리스틱보다 먼저 이기므로, 진짜로 오래된 SPEC은 `era: V3R5` 한 줄로 영구히 고정할 수 있다.

### 옵션 D — era 엔진이 `created_at:`도 읽게 한다

`created` 대신 날짜에 기대는 것이 옳으냐는 질문(카탈로그에 46건이 디코더가 못 읽는 형태를 쓴다)에 대한 직답이다. **기각한다.** 이 변경은 스키마 위반을 era 엔진에게 보이지 않게 만들어, `FrontmatterInvalid`가 존재해야 할 이유 자체를 지운다. §1.2에서 본 자기차폐를 **영구화**하는 방향이다. 옳은 처리는 반대다 — 옵션 C 아래에서 INIT-WIZARD는 신호가 없어 V3R5로 남고, 그 7건의 `FrontmatterInvalid`는 계속 advisory로 보인 채 남는다. 그 SPEC의 frontmatter가 별도 카드로 수리되는 순간 `created: 2026-05-22`가 읽히고, 그때 H-5가 자동으로 V3R6으로 올린다. **결함의 수리가 분류의 수리를 낳는다** — 엔진에 관용을 넣으면 이 연결이 끊긴다.

## 6. 형제 카드 의존 — t371

**t371**(spec-lint CI 체크아웃 + 18건 finding 재분류)은 이 카드의 영향을 받는다. SPEC의 grandfather 여부가 바뀌면 그 SPEC 위에서 `StatusGitConsistency`가 advisory인지도 함께 바뀌기 때문이다.

- **겹치는 파일:** 없다. 이 SPEC은 `internal/spec/era.go`만 수정하고 `internal/spec/lint.go`는 손대지 않는다(REQ-EH3-004). t371이 `lint.go`를 만진다면 텍스트 충돌은 발생하지 않는다.
- **겹치는 것은 행동이다.** 재분류 대상 SPEC 중 현재 `StatusGitConsistency` finding을 가진 것은 **1건** — `SPEC-KANBAN-TODO-CLI-001` (M11). 다만 이 rule은 발화 지점에서 이미 `Advisory: true`를 세우므로(`lint.go:1335`) grandfather 상태가 바뀌어도 `--strict` 결과는 변하지 않는다. 즉 **실제 충돌 위험은 측정상 0**이다.
- **순서 권고:** 두 카드는 서로를 막지 않는다. 다만 t371이 finding 수를 세는 기준선을 갖는다면, 그 기준선은 이 카드의 착지 전후로 달라질 수 있으므로 **어느 쪽이 나중이든 착지 후 재측정**이 필요하다. 리드가 순서를 정하되, 나중 카드가 자기 수치를 병합 트리에서 다시 잰다는 조건이면 병렬로 진행해도 무방하다.
