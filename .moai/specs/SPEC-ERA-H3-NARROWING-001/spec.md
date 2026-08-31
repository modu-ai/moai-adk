---
id: SPEC-ERA-H3-NARROWING-001
title: "H-3 시대 분류 술어 축소 — 진행 중 SPEC의 V3R5 오분류 차단"
version: "0.3.0"
status: draft
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "era, classification, grandfather, drift, spec-audit"
tier: S
---

# SPEC-ERA-H3-NARROWING-001 — H-3 시대 분류 술어 축소

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-08-31 | manager-spec | 최초 작성 (Tier M). 카드 t382 |
| 0.2.0 | 2026-08-31 | manager-spec | **Tier M → S 정정** (LOC·파일 수 기준 미달, AC 8개가 S 상한에 부합 — acceptance.md 폐기하고 AC를 §3으로 인라인). 무게중심을 lint 심각도에서 **drift 면제**로 이동 |
| 0.3.0 | 2026-08-31 | manager-spec | RED 증거를 `verification-completeness.md` §2.1 4요소(명령·축자 stdout·exit code·트리 SHA)로 승격 — 원장 `.moai/reports/t382/red-evidence.md` R1~R3 신설, 기준선을 트리 `f72c0bf0f`에서 **재측정**(V3R5 23→24, grandfathered 285→286). AC-001의 단위 판정 명령이 오늘 초록이라는 사실을 명시하고 RED를 코퍼스로 이관. `matchesModernPhase`의 좁은 술어(`v3.0` 접두 / `v3r6`) 실측 반영. 가드를 2층으로 분리하고 코퍼스 층이 상시 가드가 아님을 명시. t371 무충돌 주장을 「절반만 검증됨」으로 정정 |

## 1. 배경

### 1.1 결함

`internal/spec/era.go`의 `ClassifyEra`는 SPEC을 다섯 시대로 분류하고, 그 분류가 **면제 여부를 가른다**. V2.x·V3R2-R4·V3R5는 `Era.EraFinal()`이 참이 되어 grandfather 보호를 받는다.

휴리스틱 체인은 첫 매치 승리다. H-3의 술어는 한 줄이다 (`era.go:150`, 트리 `9328a5242`):

```go
if hasSyncSection && syncSHA == "" {
    return EraV3R5, "H-3 (§E.2 present, sync_commit_sha missing)"
}
```

`hasSyncSection`은 progress.md에 리터럴 `§E.2`가 있는지만 본다(변수명은 오칭이며 코드 주석이 이를 명시한다 — sync 단계는 `§E.4`다). 그런데 `§E.2` 골격 제목은 **plan-phase에 manager-spec이 항상 찍는다** — `§E.1`~`§E.4`를 한꺼번에 만드는 것이 규약이다. 그리고 `sync_commit_sha`는 sync가 닫히기 전에는 비어 있다.

따라서 이 술어는 "이 SPEC이 언제 쓰였는가"가 아니라 **"이 SPEC이 아직 안 닫혔는가"** 를 재고 있다. 걸리는 순간 `return`이므로 `created` 날짜를 읽었을 H-5에 도달하지 못한다.

실측 (트리 `9328a5242`, `./bin/moai spec audit`): V3R5로 분류된 **23건 중 22건이 `created >= 2026-04-01`**, `created < 2026-04-01`인 것은 **0건**. 오늘의 V3R5 버킷에 진짜로 오래된 SPEC은 한 건도 없다.

### 1.2 무게중심 — 오늘 실제로 무는 곳은 drift 면제다

이 결함의 값을 정확히 매기기 위해 세 소비 축을 따로 잰다. 셋의 크기는 서로 크게 다르다.

**축 1 — `spec drift`의 status↔git 대조 면제. 오늘 22건에 걸려 있다. 여기가 무게중심이다.**

`internal/spec/drift.go`의 ④ 정렬 단계는 `EraFinal()`이 참이면 `GitImpliedStatus: "era-exempt"`, `Drifted: false`를 박고 그 SPEC을 git 분류 대상에서 뺀다. `./bin/moai spec drift` 실측(M13): 23건 중 **22행이 `era-exempt`**, 1행은 `terminal-exempt`. 즉 22건의 frontmatter status가 git 이력과 한 번도 대조되지 않는다.

면제가 걸리는 창이 정확히 **작업 중인 창**이라는 점이 이 축이 무게중심인 이유다. H-3은 `sync_commit_sha`가 비어 있는 동안만 발화하고, 그 구간이 곧 SPEC이 실제로 편집되고 status가 손으로 옮겨지는 구간 — status가 git과 어긋날 가능성이 가장 큰 구간이다. 검사는 정확히 그동안 꺼져 있다가, SPEC이 닫혀 status가 더 이상 움직이지 않게 된 뒤에야 켜진다.

**축 2 — `spec audit`의 MUST-FIX drift. 오늘도 앞으로도 0이며, 구조상 그렇다.**

유일한 V3R6 drift 차원인 `SyncStatusDrift`는 `sync_commit_sha`가 채워져 있을 것을 요구한다. H-3이 발화하는 조건은 그 값이 비어 있는 것이므로 둘은 상호 배타적이다. 게다가 sync가 닫혀 SHA가 채워지는 순간 H-3이 더 이상 발화하지 않아 H-4가 V3R6을 낸다 — **이 축은 자기치유한다.** 이 SPEC은 이 축에서 얻는 것이 없고, 그렇다고 주장하지 않는다.

**축 3 — lint 구조 게이트. 오늘 1건, `--strict` rc 델타 0. 잠재적 이득이다.**

`internal/spec/lint.go`의 `applyEraDemotion`은 grandfather SPEC에 대해 (a) `eraDemotableCodes`(`MissingExclusions`, `FrontmatterInvalid`) 두 코드의 ERROR를 warning으로 강등하고 `Advisory: true`를 붙이며, (b) 그 SPEC의 나머지 warning에도 `Advisory: true`를 붙여 `--strict` 승격을 막는다.

**오늘의 실측값을 여기 명시한다 — 나중 읽는 사람이 잠재를 실제로 오인하지 않도록.** 재분류 대상 중 non-terminal은 14건이고(M10), 그 14건에 대한 `./bin/moai spec lint --json` 결과는 findings 11건 전부 `advisory=true`이며 `--strict` **rc = 0**이다(M11·M12). 강등된 ERROR를 실제로 가진 SPEC은 **`SPEC-V3R5-INIT-WIZARD-EXPANSION-001` 단 1건**(`FrontmatterInvalid` 7건)이고, 그마저 §5에서 보듯 이 수정으로 **재분류되지 않는다.** 따라서 **이 수정의 오늘 lint rc 델타는 0이다.**

### 1.3 그렇다면 왜 고치는가

**이것은 분류의 정확성 결함이다.** 게이트가 지금 실패를 통과시키고 있다는 주장이 **아니다** — 축 3의 측정이 그 주장을 반증한다. 고치는 이유는 둘이다.

1. **22건이 "너는 옛 시대라 검사에서 빼 준다"는 통보를 받고 있는데, 그중 오래된 것은 0건이다**(M4). 면제의 근거가 사실과 다르다.
2. **면제 집합이 조용히 자란다.** plan-phase 골격이 `§E.2`를 찍는 한, 새로 만드는 모든 SPEC이 sync를 닫을 때까지 이 버킷에 들어간다. 이 SPEC 자신이 그 증거다 — `./bin/moai spec audit --filter-spec SPEC-ERA-H3-NARROWING-001`이 `Grandfathered: 1`, `[INFO] SPEC-ERA-H3-NARROWING-001 (V3R5)`를 낸다. 기준선은 이 산출물이 커밋되면서 V3R5 23 → 24로 이미 움직였다.

축 3의 이득은 **잠재적**이다: 지금 걸리는 게이트는 없고, 되살아나는 것은 앞으로 걸릴 능력이다. 이 문장을 완화하거나 반대로 부풀리지 않는다.

### 1.4 강등 범위의 정확한 주장

카드의 초기 보고가 과장했던 지점을 좁혀 적는다. **구조 게이트 두 개가 advisory로 내려가고, 그 SPEC에 한해 `--strict`가 무력화된다.** "모든 lint 오류가 강등된다"는 넓은 주장은 **거짓이다** — `applyEraDemotion`의 `switch` 첫 갈래가 `eraDemotableCodes[f.Code]`를 요구하므로 그 밖의 ERROR는 손대지 않고 통과한다. 다른 레인이 뮤테이션으로 반증했다: `status: bogus-status-value`를 심자 `StatusValueInvalid`가 advisory 표식 없이 **ERROR**로 나왔다. 그 반증을 물려받아 좁은 주장만 싣는다.

### 1.5 자기차폐 사례 — SPEC-V3R5-INIT-WIZARD-EXPANSION-001

V3R5 23건 중 유일하게 `created`를 읽을 수 없는 건이다. frontmatter가 거부된 snake_case 별칭(`created_at:` / `updated_at:` / `labels:`)을 써서 YAML 디코더가 통째로 버린다(`spec-frontmatter-schema.md` § Rejected Snake_Case Aliases). 실제 작성일은 본문 HISTORY의 **2026-05-22** — 임계값 이후다. 이름이 `V3R5`로 시작하지만 사실은 현대 SPEC이고, H-5에 안 보이는 이유는 오직 별칭 때문이다.

결함이 **자기차폐적**이다. H-3 오분류가 감싸 주는 그 잘못된 frontmatter가, 동시에 H-5의 구제를 막는 원인이다. 이 SPEC이 내는 `FrontmatterInvalid` 7건 — 스키마 위반을 신고할 수 있었던 유일한 신호 — 은 전부 `[grandfathered era — downgraded to warning]` 꼬리표를 달고 advisory로 잠겨 있다.

**단위를 붙여 적는 인구 수:** `created_at:`을 쓰는 spec.md는 글롭 `.moai/specs/*/spec.md` 기준 **46건**, 재귀 `grep -rl ... .moai/specs --include=spec.md` 기준 **47건**이다(초과 1건은 `.moai/specs/_archive/SPEC-DESIGN-CONST-AMEND-001/spec.md`, 감사 대상 밖). 그중 **V3R5 버킷 소속은 1건**(M7). 46이라는 인구가 수정 설계에 무엇을 뜻하는지는 §5 옵션 D에서 다룬다.

### 1.6 단위를 붙인 시대 분포 (트리 `9328a5242`, plan 산출물 커밋 이전)

| 값 | 수 | 어느 단위인가 |
|---|---|---|
| Total SPECs | 714 | `./bin/moai spec audit` 요약 |
| Grandfathered | 285 | 같은 요약 (V2.x + V3R2-R4 + V3R5) |
| V3R6 총계 | **429** | 714 − 285. 감사 요약이 직접 내지 않으므로 뺄셈으로 도출 |
| `EraAutoDetected` finding 중 V3R6 | 205 | **부분집합** — 이 finding은 `era:` frontmatter가 **없는** SPEC에만 발화한다. 시대 총계가 아니다 |
| `EraAutoDetected` 중 V2.x / V3R2-R4 / V3R5 | 144 / 118 / 23 | 이 셋은 285와 합이 같아 총계와 일치 |
| `EraUnclassified` finding | **0** | §5의 H-6 낙하 여부를 판정할 기준선 |

`EraAutoDetected` 집계를 시대 총계로 읽으면 V3R6이 205로 보인다. 그 혼동이 이 표가 존재하는 이유다.

## 2. 요구사항 (GEARS)

- **REQ-EH3-001** (unwanted): `ClassifyEra`는 modern-era 신호를 가진 SPEC을 H-3 경로로 `EraV3R5`로 분류해서는 안 된다(shall not). modern-era 신호는 H-5가 쓰는 것과 **동일한 술어**로 정의한다 — `matchesModernPhase(FrontmatterPhase)` 또는 `isAfterModernThreshold(FrontmatterCreated)`.
- **REQ-EH3-002** (capability-gate): Where a SPEC이 modern-era 신호를 하나도 갖지 않는 경우, `ClassifyEra`는 수정 이전과 동일한 시대와 rationale 문자열을 반환해야 한다(shall).
- **REQ-EH3-003** (event-driven): When H-3이 유예되면, `ClassifyEra`는 H-4 → H-4-legacy → H-5 → H-6 체인을 원래 순서 그대로 이어서 평가해야 한다(shall). 어떤 휴리스틱도 재배치하지 않는다.
- **REQ-EH3-004** (unwanted): 이 수정은 `eraDemotableCodes`, `applyEraDemotion`, `Era.EraFinal()`, `Era.IsModern()`을 변경해서는 안 된다(shall not). 변경 지점은 H-3 술어 한 곳이다.
- **REQ-EH3-005** (ubiquitous): `internal/spec/era_test.go`는 회귀 가드를 실어야 한다(shall) — 어떤 `EraSignals` 조합에서도 「H-5 술어가 참인데 결과가 `EraV3R5`」가 성립하지 않음을 단언하는 불변식 테스트.
- **REQ-EH3-006** (ubiquitous): `.claude/rules/local/lifecycle-sync-gate.md`의 휴리스틱 표 H-3 행은 새 술어를 반영해야 한다(shall).

## 3. 인수 기준

Tier S이므로 AC는 여기에 인라인한다. 측정은 모두 이 트리의 `./bin/moai`(`make build` 산출물)로 했고 PATH 바이너리는 쓰지 않았다.

**RED 증거의 소재.** `verification-completeness.md` §2.1은 release-blocking 기준의 RED 셀에 네 요소를 요구한다 — 명령 · 축자 stdout · exit code · 트리 SHA. 그 네 요소를 갖춘 항목은 `.moai/reports/t382/red-evidence.md`에 **R1 · R2 · R3**으로 있고, 트리 `f72c0bf0f`에서 측정했다. 아래 AC는 그 원장을 id로 인용한다.

`.moai/reports/t382/measurements-9328a5242.md`의 **M1~M13은 축자 stdout과 exit code를 갖지 않는다.** 그래서 배경 측정으로만 인용하고, 판정이 실제로 기대는 RED는 R1~R3뿐이다. M-번호가 아래에 나오면 그것은 배경이지 판정 근거가 아니다.

### 3.1 방향 선언

「오분류가 멈췄음만 증명하는 기준 집합은 판정의 절반」이라는 [HARD] 의무에 따라 각 AC가 재는 방향을 먼저 밝힌다.

| 방향 | AC |
|---|---|
| 오분류 중단 | AC-EH3-001, 002, 005, 006 |
| **grandfather 보호 유지** | AC-EH3-003, 004, 005 |
| **비용 측정** | AC-EH3-007 |
| 회귀 가드 | AC-EH3-008 |

AC-EH3-005는 양쪽에 동시에 선다 — 시대가 바뀐 SPEC과 **바뀌지 않은 SPEC**을 같은 표에서 한 건씩 귀속시키기 때문이다.

### 3.2 [HARD] 기준선이 이 SPEC 자신 때문에 움직였다 — 그래서 다시 쟀다

285 / 23 / 429는 이 SPEC 디렉터리가 생기기 **전**의 트리 `9328a5242`에서 잰 값이다. 이 SPEC 자신이 결함의 표본이므로(§1.3) 산출물이 커밋되면서 기준선이 움직였고, **plan-phase에서 트리 `f72c0bf0f`에 대고 다시 쟀다**(원장 R2 · R3):

| 값 | `9328a5242` (SPEC 생성 전) | **`f72c0bf0f` (현재 기준선)** |
|---|---|---|
| Total SPECs | 714 | **715** |
| Grandfathered | 285 | **286** |
| V3R5 | 23 | **24** |
| V3R6 (뺄셈 도출) | 429 | **429** |
| `EraUnclassified` | 0 | **0** |

아래 AC의 기대값은 전부 오른쪽 열 기준이다. **왼쪽 열의 수를 사후 근거로 재사용하지 않는다** — 그것이 baseline 귀속 위반이다.

이 표 자체도 만료된다. run-phase는 M1 착수 직전 자기 트리에서 R1·R2를 다시 돌려 오른쪽 열을 갱신하고, 갱신본에 대고 대조한다. 착수와 측정 사이에 다른 SPEC이 하나라도 추가되면 715와 286은 함께 움직인다.

### 3.3 단위 기준 — `internal/spec/era_test.go`

판정 명령(넷 공통): `go test -run TestClassifyEra ./internal/spec/`

**AC-EH3-001 — 날짜 신호가 있으면 H-3을 통과한다** (오분류 중단)
Given `ProgressMDExists: true` + `ProgressMDContent`에 `## §E.2 Run-phase Evidence`가 있고 `sync_commit_sha`는 빈 값이며 `FrontmatterCreated: "2026-08-25"`, When `ClassifyEra(signals)`를 호출하면, Then `era == EraV3R6` 이고 rationale이 `"H-5"`로 시작한다.
*오늘의 RED — 원장 R1.* 판정 명령 `go test -run TestClassifyEra ./internal/spec/`는 **오늘 초록이다**. 그 서브테스트가 아직 없기 때문이며, 셀렉터가 0건을 매치한 초록과 구별되지 않는다. 그래서 이 AC의 RED는 단위 명령이 아니라 **코퍼스 등가**로 세운다: 원장 R1이 exit **1** 과 `carrying a modern-era signal (misclassified): 23`을 낸다. 이 명제는 오늘 23번 거짓이다.
*green path:* M1의 유예절이 들어가면 R1이 exit 0 · `misclassified: 0`으로 뒤집히고, 새 서브테스트가 단위 층에서 같은 명제를 잡는다.

**AC-EH3-002 — phase 신호가 있으면 H-3을 통과한다** (오분류 중단)
Given 같은 progress 신호에 `FrontmatterCreated: ""`, `FrontmatterPhase: "v3.0.0"`, When 호출하면, Then `era == EraV3R6` 이고 rationale이 `"H-5"`로 시작한다.
*오늘의 RED — 원장 R1.* 현행 H-3은 `phase`를 읽지 않아 `EraV3R5`를 낸다. R1의 판정 술어는 `phase`와 `created`를 **둘 다** 포함하므로 이 AC도 R1의 exit 1에 실린다. 배경: 23건 중 `matchesModernPhase` 참이 5건(M8).
*주의 — `matchesModernPhase`는 좁다.* 코드는 `v3r6` 포함 또는 `v3.0` **접두**만 참으로 낸다(`era.go:265`). 따라서 `"v3.2.0 target"`는 거짓이며, 이 SPEC 자신도 `phase`가 아니라 `created`로만 재분류된다(원장 R3). AC 픽스처에 `"v3.0.0"`을 쓰는 것은 이 좁은 술어를 통과시키기 위해서다.

**AC-EH3-003 — 임계 이전 SPEC은 보호를 잃지 않는다** (보호 유지)
Given 같은 progress 신호에 `FrontmatterCreated: "2026-03-15"`, `FrontmatterPhase: ""`, When 호출하면, Then `era == EraV3R5` 이고 rationale이 `"H-3"`로 시작한다.
*이 기준은 오늘 통과한다. 그래서 뮤테이션으로 결정력을 부여한다:* 유예절(`&& !hasModernEraSignal(signals)`)을 무조건 스킵으로 바꾸면 — H-3을 지우면 — 이 서브테스트가 반드시 실패해야 한다. run-phase가 실제로 심어 실패를 관측하고 출력을 남긴 뒤 되돌린다. 관측 없이 "통과했다"만 보고하면 이 AC는 아무것도 결정하지 않는다.

**AC-EH3-004 — 신호가 전무하면 H-6으로 떨어지지 않는다** (보호 유지)
Given 같은 progress 신호에 `FrontmatterCreated: ""`, `FrontmatterPhase: ""`, `FrontmatterEra: ""` (INIT-WIZARD의 실제 모양 — 별칭 때문에 두 신호가 모두 빈 값으로 디코딩된다), When 호출하면, Then `era == EraV3R5`, rationale `"H-3"`. `EraUnclassified`가 **아니다**.
*뮤테이션 결정력:* AC-003과 같은 뮤테이션에서 이 서브테스트는 `EraUnclassified`를 받아 실패해야 한다.
*코퍼스 기준선:* `EraUnclassified` finding은 오늘 **0건**(M3). 수정 후에도 0이어야 하며, 1건이라도 생기면 그 SPEC을 이름으로 조사한다.

**`unclassified`가 실제로 무엇을 뜻하는가 — 코드를 읽어 확정한 답 (트리 `9328a5242`).** 채택 설계는 이 상태를 만들지 않지만(§5), 만들지 않는 것이 왜 옳은지는 값이 매겨져야 한다.

| 소비 지점 | 취급 | 노출인가 보호인가 |
|---|---|---|
| `lint.go` `isGrandfatheredSpecDir` | `EraFinal()` 거짓 → `applyEraDemotion` 미적용 | **노출** — 구조 게이트 ERROR 그대로, `--strict` 승격 가능 |
| `drift.go` ④ era 정렬 | `EraFinal()` 거짓 → `era-exempt` 미부여 | **노출** — git 대조 분류 대상 |
| `audit.go` `auditSpec` | `era == EraUnclassified` 갈래에서 INFO finding 하나 내고 **조기 return** — `checkV3R6Drift`가 아예 안 돈다 | **어느 쪽도 아니다** — `grandfathered`에도 `modern_era_clean`에도 세어지지 않고 MUST-FIX가 원리상 안 생긴다 |

요약: lint·drift 축에서는 노출, audit의 sync-drift 축에서는 침묵. INFO finding 하나로 표시되므로 완전한 은폐는 아니나 V3R6과는 다르다. 이것이 「신호 없음 → 재분류 없음」이라는 보수적 기본값의 근거다.

### 3.4 코퍼스 기준

**AC-EH3-005 — 시대 변화를 SPEC 단위로 귀속한다** (양방향)
Given 수정 전후 각 트리에서, When `./bin/moai spec audit --json`으로 SPEC별 `era`를 뽑아 대조하면, Then 넷이 모두 성립한다:
1. **총계 대조** — §3.2의 갱신된 기준선 기준으로 grandfathered 286 → **263**, V3R6 429 → **452**, V3R5 24 → **1**, `EraUnclassified` 0 → **0**. V2.x 144 · V3R2-R4 118은 **불변**(변하면 H-3 이외가 건드려졌다는 뜻이며 REQ-EH3-003 위반).
2. **한 건씩 귀속** — 시대가 바뀐 SPEC은 정확히 **23건**이며, `v3r5-population.txt`의 23건에서 `SPEC-V3R5-INIT-WIZARD-EXPANSION-001`을 뺀 22건 + 이 SPEC 자신 = 23건과 **원소 단위로 일치**한다. 총계만 맞추는 것으로는 이 기준을 만족하지 못한다.
3. **바뀐 건마다 근거 제시** — 23건 각각에 `created >= 2026-04-01` 또는 `matchesModernPhase(phase)` 참이 성립함을 SPEC별 표로 보인다. **근거 없이 시대가 바뀐 SPEC이 한 건이라도 있으면 실패** — 이것이 grandfather 손실 벡터를 매번 재는 장치다.
4. **표본 검증** — 23건 중 최소 5건을 뽑아 `created` 값과 본문 HISTORY 첫 행 날짜를 대조해 진짜 V3R6임을 확인하고, 표본 목록을 이름으로 남긴다.

*오늘의 RED — 원장 R2 · R1.* R2가 `Grandfathered: 286`을, R1이 재분류 대상 23건과 잔류 1건의 **이름 집합**을 낸다. 명제 1은 오늘 거짓이고, 명제 2의 원소 비교 대상은 R1의 `no-signal set` 출력이다(총계만으로는 만족되지 않는다).

*빈 스윕 방어:* R1의 첫 줄 `V3R5-classified SPECs: N`이 스윕 모집단이다. N = 0이면 조사 대상이 없었다는 뜻이므로 exit 0을 통과로 읽지 않는다 — 판정 전에 N ≥ 1을 확인한다.

**AC-EH3-006 — 게이트 복원을 주입 결함으로 증명한다** (오분류 중단)
「오늘 이미 통과하는 기준은 아무것도 결정하지 않는다」를 이 AC가 정면으로 다룬다. **코퍼스 lint rc 델타로는 쓸 수 없다** — §1.2 축 3에서 보였듯 수정 전후 모두 rc 0이다. 그러므로 결함을 심어 잰다.
Given 재분류 대상 중 하나(예: `SPEC-KANBAN-WORKTREE-001`)의 SPEC 디렉터리를 `t.TempDir()` 아래로 복사하고 `spec.md`의 「Out of Scope」 H3 소제목을 제거해 `MissingExclusions`를 인위적으로 성립시킨 뒤, When 그 사본에 `moai spec lint --strict --json`을 돌리면, Then 수정 **전** 바이너리는 rc **0** 이고 해당 finding이 `severity: warning` · `advisory: true` · 메시지 말미 `[grandfathered era — downgraded to warning]`를 달고 나오며, 수정 **후** 바이너리는 rc **1** 이고 같은 finding이 `severity: error` · advisory 없음으로 나온다.
*주의:* 결함 주입은 반드시 임시 사본에서. 실제 `.moai/specs/` 파일을 훼손하지 않는다. 수정 전 바이너리는 M1 착수 전 `bin/moai-pre-t382`로 보존한다.

**AC-EH3-007 — drift 축의 비용을 잰다** (비용 측정 — 무게중심 축)
Given `.moai/reports/t382/drift-before-9328a5242.txt`(23행: `era-exempt` 22 + `terminal-exempt` 1), When 수정 후 `./bin/moai spec drift`로 같은 SPEC들의 행을 뽑으면, Then:
1. `era-exempt`였던 22행 중 **21행**이 git 대조 결과(`completed` / `in-progress` / `implemented` 등)로 바뀐다. 남는 1행은 INIT-WIZARD(V3R5 유지). `SPEC-HOOK-PREEDIT-INVESTIGATE-001`은 이미 `terminal-exempt`라 변화 없다. 이 SPEC 자신의 행도 함께 확인한다.
2. 새로 `DRIFT`로 표시된 행이 있으면 **각각에 대해** frontmatter status와 git 이력을 직접 대조해 진짜 불일치인지 판정하고 건별 결과를 남긴다. 진짜 불일치는 이 카드의 결함이 아니라 **드러난 기존 상태**이며, 오탐이면 이 수정이 만든 비용이다.
3. **오탐이 1건이라도 확인되면 이 AC는 실패**이고 설계를 다시 연다.

*오늘의 RED:* 위 파일의 22행이 전부 `era-exempt`인 것이 기준선이다.

### 3.5 회귀 가드

**AC-EH3-008 — 불변식 가드가 known-bad 입력에서 실패한다**
Given `era_test.go`에 불변식 테스트 `TestClassifyEra_NoV3R5WhileModernSignal`을 심는다 — `{§E.2 유무} × {sync_commit_sha 유무} × {created ∈ (빈값, 임계 이전, 임계 이후)} × {phase ∈ (빈값, "v3.0.0", "v3.2.0 target")}` 곱집합을 순회하며 **「H-5 술어가 참인데 결과가 `EraV3R5`」인 조합이 하나도 없음**을 단언한다. When `go test -run TestClassifyEra_NoV3R5WhileModernSignal ./internal/spec/`를 돌리면, Then 수정 후 통과하고 **뮤테이션에서 반드시 실패한다** — H-3 술어에서 `&& !hasModernEraSignal(signals)`를 제거하면(=수정 이전 상태) 이 테스트가 실패해야 한다.
*뮤테이션이 이 AC의 본체다.* 통과 관측만으로는 가드가 공허한지 알 수 없다. run-phase는 절을 실제로 지워 실패를 관측하고, 되돌린 뒤 재통과까지 기록한다.
*가드 선택 근거:* 이 결함의 본질은 「첫 매치 승리 체인에서 앞선 절이 뒤 절을 굶긴다」이고 재발 경로는 유예 조건이 리팩터링에서 조용히 사라지는 것이다. 개별 케이스 테스트(AC-001~004)는 **테스트에 없는 조합**으로 재발하면 못 잡는다. 조합 순회 불변식이 그 구멍을 닫는다. `lint_movingref_test.go` · `lint_artifact_status_test.go`가 이미 「이 표식을 지우면 어떤 테스트가 실패해야 하는가」를 주석으로 명시하는 관행을 쓴다.

*가드는 두 층이며, 두 층인 이유가 있다.* 단위 불변식은 **합성 입력**을 돌고 코퍼스 프로브(원장 R1)는 **실제 카탈로그**를 돈다. 단위 층만 두면 술어는 맞는데 `LoadEraSignalsFromDir`가 frontmatter를 다르게 읽어 실트리에서 어긋나는 경우를 못 잡고, 코퍼스 층만 두면 오늘 카탈로그에 없는 조합의 재발을 못 잡는다. 두 층 모두 known-bad 입력에서 실패가 관측돼야 채택된다 — 단위는 뮤테이션으로, 코퍼스는 이미 관측된 exit 1(R1)로.

*코퍼스 층의 계속 발화 여부.* `red_probe.py`는 CI에 배선되지 않는다. 그래서 이 층은 **run-phase 1회 판정 도구**이지 상시 가드가 아니며, 그렇게만 주장한다. 상시로 재발을 잡는 것은 단위 불변식(`go test ./internal/spec/...`, CI가 매번 실행)이다. 코퍼스 층을 상시 가드로 소개하면 「멈춘 검사가 성공과 구별되지 않는」 결함을 새로 만드는 셈이다.

### 3.6 완료 정의

- [ ] AC-EH3-001 ~ 008 전부 통과, 각 AC마다 판정 명령과 축자 출력이 `.moai/reports/t382/` 아래에 있다
- [ ] 원장 R1이 exit **1 → 0** 으로 뒤집힌 것을 축자 출력으로 관측했고, 판정 전 스윕 모집단 N ≥ 1을 확인했다
- [ ] 기준선 표(§3.2 오른쪽 열)를 M1 착수 직전 트리에서 갱신했고, 대조는 갱신본에 대고 했다
- [ ] AC-003 · 004 · 008의 **뮤테이션 실패 관측**이 출력과 함께 기록되고 뮤테이션이 되돌려졌다
- [ ] `go test ./internal/spec/...` 통과 (로컬 전체 스위트 금지)
- [ ] `go vet ./internal/spec/...` · `golangci-lint run` 통과
- [ ] `.claude/rules/local/lifecycle-sync-gate.md` H-3 행 갱신 (REQ-EH3-006)
- [ ] 모든 수치가 측정 명령과 트리 SHA에 귀속돼 있다

## 4. 범위 밖

### Out of Scope — 형식이 깨진 frontmatter의 수리

- `SPEC-V3R5-INIT-WIZARD-EXPANSION-001`의 `created_at:` / `updated_at:` / `labels:` 별칭을 정본 필드로 고치는 일. 그 수리는 별도의 소유자를 갖는다(비-전이 frontmatter 정정 — `spec-frontmatter-schema.md` § Non-transition frontmatter corrections에 따라 manager-spec, 오케스트레이터 재위임 경유).
- 나머지 45건의 `created_at:` 보유 SPEC 일괄 수리.

### Out of Scope — era 엔진에 별칭 관용을 추가하는 일

- `EraSignals`가 `created_at:`을 `created:`의 대체 신호로 읽게 만드는 변경. §5 옵션 D에서 근거를 적어 기각했다.

### Out of Scope — 형제 카드가 소유한 영역

- **t371** — spec-lint CI 체크아웃 + 18건 finding 재분류. 의존은 §6에 적었으나 재분류 자체는 하지 않는다.
- **t376** — 새 전이 규칙 추가.
- **t380** — 경계 픽스처.
- 셋 모두 `internal/spec` 안에서 살아 있는 형제 카드다. 이 SPEC은 `era.go`의 H-3 술어 한 곳과 그 테스트, SSOT 문서 한 곳만 만진다.

### Out of Scope — 휴리스틱 재배치

- H-1~H-6의 **순서** 변경(§5 옵션 A에서 기각).
- `modernEraThreshold` 상수값(`2026-04-01`) 변경.

## 5. 설계 선택 — 네 후보와 기각 근거

### 옵션 A — H-5(날짜/phase 검사)를 H-3 앞으로 옮긴다

H-5가 H-3보다 앞서면 **H-4보다도 앞선다.** 그러면 임계 이후 SPEC은 정밀한 3-단계 술어(`§E.2 + §E.4 + sync_commit_sha`)로 잡혔을 경우에도 rationale이 `"H-5 (modern phase or created date)"`로 뭉개진다. 코드는 이 정밀도를 명시적 가치로 선언한다 — H-4-legacy 주석이 "explicit predicate is a stronger signal than the H-5 date heuristic"라는 이유만으로 그 갈래를 유지한다. 순서를 뒤집으면 그 가치를 카탈로그 전체에서 버린다. **기각.**

### 옵션 B — H-3 술어에 「현대 마커 부재」를 추가한다 (예: `§E.4` 없음을 요구)

측정으로 기각된다. V3R5 23건 중 **21건이 이미 `§E.4`를 갖고 있다**(M9) — plan-phase 골격이 `§E.1`~`§E.4`를 한꺼번에 찍기 때문이다. 판별력이 사실상 없어 23건 중 2건만 구제하며, 기존 테스트 `"H-3 §E.2 + §E.4 present but sync_commit_sha empty → V3R5"`를 정면으로 깬다 — 그 테스트가 표현하는 의도(sync 미완)는 옳고 이 옵션이 틀렸다. **기각.**

### 옵션 C — H-3을 modern-era 신호가 있을 때만 유예한다 ← **채택**

```go
if hasSyncSection && syncSHA == "" && !hasModernEraSignal(signals) {
    return EraV3R5, "H-3 (§E.2 present, sync_commit_sha missing)"
}
```

`hasModernEraSignal`은 H-5가 이미 쓰는 술어를 그대로 함수로 뽑은 것이다 — `matchesModernPhase(phase) || isAfterModernThreshold(created)`. H-5의 조건절도 이 헬퍼로 치환해 **두 지점이 같은 술어를 쓴다는 사실을 코드로 못박는다.** 그 동일성 위에 핵심 불변식이 선다: **수정 후에는 「H-3 적격이면서 동시에 H-5가 현대라 부를 SPEC」이 존재할 수 없다.** REQ-EH3-005의 가드가 검사하는 명제가 이것이다.

채택 근거 셋:

1. **유예가 양성 신호에 게이트된다.** 신호가 없으면 조건이 거짓이 되어 H-3이 그대로 발화한다. 따라서 **이 설계는 새로운 H-6 낙하를 원리상 만들 수 없다.** 판정 불가능한 SPEC은 보수적으로 보호를 유지하며, 이는 grandfather 절의 취지와 같은 방향이다. AC-EH3-004가 `EraUnclassified` 0 기준선(M3)에 대고 이를 검사한다.
2. **기존 테스트를 하나도 깨지 않는다.** `era_test.go`의 H-3 케이스들은 `FrontmatterCreated` / `FrontmatterPhase`를 비워 두고 있어 유예 조건이 거짓이 된다. 우연한 통과가 아니라, REQ-EH3-002가 요구하는 성질이 이미 그 테스트로 표현돼 있다는 뜻이다.
3. **rationale 정밀도와 휴리스틱 순서를 보존한다** — 옵션 A가 버리는 두 가지.

**날짜만 쓰지 않고 phase까지 포함한 이유.** 날짜만 쓰면 「임계 이전 `created` + `phase: "v3.0.0"`」인 SPEC이 여전히 grandfather로 남아 유예 조건과 H-5 술어가 어긋난다. 그러면 위 불변식이 성립하지 않고 가드도 쓸 수 없다. 둘을 붙이면 술어가 정확히 일치한다.

**이 선택이 지는 잔여 위험.** 임계 이전에 쓰였는데 `phase:`에 `v3.0` 계열 라벨을 단 SPEC은 grandfather 보호를 잃는다. 오늘 이 교집합은 **0건**이다 — V3R5 23건 중 `created < 2026-04-01`인 것이 0건이므로(M4) 구성상 공집합이다. 다만 "오늘 0건"이 "원리상 0건"은 아니므로 AC-EH3-005 명제 3이 시대가 바뀐 SPEC을 한 건씩 귀속시켜 이 벡터를 매번 재측정하게 한다. 원리상의 탈출구도 이미 있다 — frontmatter `era:` 필드(H-override)가 모든 휴리스틱보다 먼저 이기므로, 진짜로 오래된 SPEC은 `era: V3R5` 한 줄로 영구 고정할 수 있다.

### 옵션 D — era 엔진이 `created_at:`도 읽게 한다

「46건이 디코더가 못 읽는 형태로 쓰는 필드에 기대는 것이 옳으냐」는 질문에 대한 직답이다. **기각한다.**

두 가지가 이 기각을 지탱한다.

첫째, **범위가 1건이다.** 46건 중 V3R5 버킷에 있는 것은 INIT-WIZARD뿐이다(M7). 나머지 45건은 다른 경로로 이미 올바르게 분류돼 있다 — `created_at`이 era 엔진에 실제로 손해를 끼치는 사례는 단 하나이며, 그 하나를 위해 엔진의 입력 계약을 넓히는 것은 비례하지 않는다.

둘째, 그리고 더 중요하게 — **이 변경은 스키마 위반을 era 엔진에게 보이지 않게 만들어 `FrontmatterInvalid`가 존재해야 할 이유를 지운다.** §1.5의 자기차폐를 **영구화**하는 방향이다. 옳은 처리는 반대다: 옵션 C 아래에서 INIT-WIZARD는 신호가 없어 V3R5로 남고 7건의 `FrontmatterInvalid`가 계속 보인 채 남는다. 그 frontmatter가 별도 카드로 수리되는 순간 `created: 2026-05-22`가 읽히고, 그때 H-5가 자동으로 V3R6으로 올린다. **결함의 수리가 분류의 수리를 낳는다** — 엔진에 관용을 넣으면 이 연결이 끊긴다.

## 6. 형제 카드 의존 — t371

**t371**(spec-lint CI 체크아웃 + 18건 finding 재분류)은 이 카드의 영향을 받는다. SPEC의 grandfather 여부가 바뀌면 그 SPEC 위에서 `StatusGitConsistency`가 advisory인지도 함께 바뀌기 때문이다.

- **겹치는 파일: 없다 — 단, 절반만 검증된 주장이다.** 검증된 쪽: 이 SPEC의 수정 파일은 `era.go` · `era_test.go` · `lifecycle-sync-gate.md` 3개로 §5·§F가 고정하고 REQ-EH3-004가 못박는다. 검증 못 한 쪽: **t371의 산출물은 이 트리에 없다** — `find .moai/reports -maxdepth 1 -name 't371'`이 빈 결과를 내고 (`t370` · `t372` · `t375` · `t382`만 있다) SPEC 디렉터리도 없다. 따라서 「t371이 `lint.go`를 만진다」는 카드 설명에서 온 전언이지 이 트리에서 읽은 사실이 아니다. 무충돌은 **이쪽 3파일이 t371의 실제 수정 집합과 겹치지 않을 때만** 성립하며, 그 확인은 t371 산출물이 존재하는 트리에서 해야 한다.
- **겹치는 것은 행동이다.** 재분류 대상 중 현재 `StatusGitConsistency` finding 보유는 **1건** — `SPEC-KANBAN-TODO-CLI-001`(M11). 다만 이 rule은 발화 지점에서 `Advisory: true`를 세우므로(`lint.go:1335`) grandfather 상태가 바뀌어도 `--strict` 결과는 변하지 않는다. **측정상 실제 충돌 위험은 0이다.**
- **순서 권고:** 두 카드는 서로를 막지 않는다. 다만 t371이 finding 수 기준선을 갖는다면 그 기준선은 이 카드의 착지 전후로 달라질 수 있으므로 **어느 쪽이 나중이든 병합 트리에서 재측정**해야 한다. 그 조건이면 병렬 진행이 가능하며, 순서 지명은 리드의 몫이다.
