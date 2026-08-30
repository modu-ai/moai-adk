---
id: SPEC-COVERAGE-RULE-SCOPE-001
title: "CoverageRule 파싱 경로 결함 두 건 — REQ 파서 협소성과 acceptance.md 미판독"
version: "0.1.0"
status: in-progress
created: 2026-08-30
updated: 2026-08-30
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "lint, coverage-rule, req-parser, spec-lint, internal/spec"
tier: M
---

# SPEC-COVERAGE-RULE-SCOPE-001 — CoverageRule 파싱 경로 결함 두 건

## HISTORY

### 2026-08-30 — 최초 작성 (t362 조사 결과 반영)

카드 t362는 결함 ① 하나만 지목한 채 열렸다. 조사 과정에서 결함 ②가 드러났고, 리드가 두 건을 한 SPEC으로 묶기로 판정했다. 근거는 둘이 같은 규칙의 같은 파싱 경로에 있고, 수리가 같은 파서를 건드린다는 점이다. **다만 두 건의 무게는 같지 않다 — 결함 ②가 이 SPEC의 본체이고, 결함 ①은 현재 잠복 상태인 작은 쪽이다.**

측정 기준: 워크트리 `.claude/worktrees/t362`, 브랜치 `WT-coverage-rule-scope`, HEAD `ee50984ab`(측정 시점 `origin/develop` head와 동일). 이 트리에서 빌드한 바이너리(`go build -o /tmp/moai-t362 ./cmd/moai`, rc=0)로만 쟀다. 설치본 `~/go/bin/moai`(v3.1.2, `343399d2f`)는 다른 트리이므로 근거로 쓰지 않았다.

### 2026-08-30 — 조사 중 사고 기록: 무효 픽스처가 결론을 뒤집을 뻔했다

결함 ① 재현에 쓴 첫 픽스처는 3-분절 AC ID(`AC-FIX-001`)를 썼다. 파서가 요구하는 형태는 4-분절(`AC-[A-Z0-9]+-[0-9]+-[0-9]+`, `internal/spec/parser.go:218`)이라 **대조군까지 함께 실패했다.** 그 출력을 액면대로 읽었다면 결론은 "이 규칙은 아예 돌지 않는다"로 뒤집혔을 것이다. 이 SPEC 자체가 "돌지 않는 규칙"을 다루는 문서이므로, 조사가 바로 그 거짓 형태를 만들 뻔했다는 사실을 기록으로 남긴다. 픽스처를 4-분절로 고친 뒤에야 실험군/대조군이 갈렸다.

### 2026-08-30 — 이 SPEC이 자기 결함을 스스로 재현했다

작성 직후 이 트리의 바이너리로 lint를 돌린 결과 `CoverageIncomplete` 8건, rc=1이 나왔다. 원인은 결함 ① 그 자체다 — 이 SPEC은 Tier M이라 AC를 `acceptance.md`에 두었고, 그것이 코퍼스에서 **관행을 실제로 따른 첫 SPEC**이었다. §1.2가 "코퍼스 위반 0건"이라 적은 이유가 여기서 확인된다: 0이었던 것은 아무도 규약을 따르지 않았기 때문이고, 따르는 순간 발화한다.

착지 시 develop을 붉히지 않기 위해 §2.3에 최소 중복 표를 넣었다. 그 표는 관행 위반이며, 결함 ①이 수리되면 삭제 대상이다.

우회 과정에서 같은 파싱 경로의 **세 번째 협소성**도 드러났다(§4 참조). 자세한 경위: `### 1.2 … acceptance.md …`라는 산문 표제가 `findACSectionStart`의 첫 매치가 되어 AC 절을 가로챘고, 실제 AC 표는 영영 읽히지 않았다. 표제 문구를 바꾼 뒤에야 rc=0이 되었다.

---

## 1. 배경과 문제

`moai spec lint`는 `--help`에서 "AC→REQ coverage (100% required)"와 "REQ ID uniqueness"를 광고한다. `ee50984ab` 트리의 전 코퍼스 측정 결과, 두 비교 모두 대부분의 SPEC에서 **일어나지 않는다.**

### 1.1 결함 ② — REQ 파서가 너무 좁아 규칙이 볼 것이 없다 (본체)

`parseREQs`가 쓰는 `reqLinePattern`은 `-\s+(REQ-[A-Z]{2,5}-\d{3}-\d{3})\s*:\s*(.+)`이다(`internal/spec/lint.go:452`). 실제 코퍼스에 대고 잰 값:

| 측정 | 값 |
|---|---|
| `spec.md` 총수 | 702 |
| 위 패턴에 걸리는 파일 | **15 (2.1%)** |
| 걸리지 않는 687 중 REQ 토큰을 **실제로 담고 있는** 파일 | **618 (90%)** |

걸리지 않는 ID 형태로 관측된 것들: `REQ-HOOK-001`(3-분절), `REQ-WF001-001`(도메인 안에 숫자 — `[A-Z]{2,5}`가 금지), `REQ-VNRN-RT-001-001`(5-분절), `REQ-HRN-FND-001`(두 분절 도메인), `REQ-TUX1-001`, `REQ-WC01-001`.

결과는 한 규칙에 그치지 않고 겹쳐진다. `parseREQs`가 아무것도 모으지 못하면 `CoverageRule`은 `len(doc.REQs) == 0`에서 즉시 반환하고, `InvalidREQIDRule`과 `DuplicateREQIDRule`도 검사할 대상이 없다. 전 코퍼스 실행이 이를 뒷받침한다 — `InvalidREQID` 0건, `DuplicateREQID` 0건.

**파서가 좁은 것이 검사기들이 볼 것을 잃은 원인이다.** 검사기 자체의 논리 결함이 아니다.

### 1.2 결함 ① — `CoverageRule`이 형제 AC 파일을 읽지 않는다 (잠복)

`discoverSPECs`(`lint.go:313`)는 `SPEC-*/spec.md`만 모은다. `parseSPECDoc`(`lint.go:456`)은 그 한 파일의 본문만 파싱해 `ParseAcceptanceCriteria`에 넘긴다. `CoverageRule.Check`(`lint.go:682`)는 `doc.REQs`와 `doc.Criteria`를 비교하는데, 둘 다 spec.md 하나에서 나온다. **`acceptance.md`는 어떤 코드 경로에서도 열리지 않는다.**

재현(HEAD `ee50984ab`). AC 한 줄의 위치만 다른 픽스처 두 벌:

```
/tmp/moai-t362 spec lint .moai/state/verify/t362/fixture/SPEC-FIXM-001/spec.md
  → ERROR CoverageIncomplete .../SPEC-FIXM-001/spec.md 11 REQ REQ-FIX-001-001 is not referenced by any AC   (rc=1)

/tmp/moai-t362 spec lint .moai/state/verify/t362/fixture-inline/SPEC-FIXI-001/spec.md
  → 해당 코드 없음
```

A(관행대로 `acceptance.md`에만 AC를 둠)는 걸리고, B(같은 AC 줄을 spec.md `## 3. Acceptance Criteria`에 둔 대조군)는 걸리지 않는다. 둘 다 무관한 `MissingExclusions`는 함께 낸다.

**다만 코퍼스 위반 건수는 0이다.** 전 코퍼스 lint(`/tmp/moai-t362 spec lint --json`, 총 177건) 안에 `CoverageIncomplete`는 한 건도 없다. 내역: MovingRefUnpinned 113 / MissingExclusions 24 / StatusGitConsistency 18 / FrontmatterInvalid 14 / LegacyEARSKeyword 7 / OwnershipTransitionInvalid 1. develop 병합 전(`15453140a`)과 후(`ee50984ab`)에 같은 값이 나왔다.

이 0이 공허한 0이 아니라는 근거를 함께 남긴다. `CoverageRule.Check`는 `if len(doc.REQs) == 0 { return nil }`로 시작하므로, 0은 "규칙이 아무것도 못 봤다"는 뜻일 수도 있다. 그렇지 않다 — 702개 중 15개가 `parseREQs` 형태의 REQ 줄을 갖고, 규칙은 그 15개 위에서 실제로 돌았으며, 15개 모두 spec.md 안에 AC를 중복해 두었기 때문에 통과한다. 즉 **관행(AC를 `acceptance.md`에 둔다)을 따르는 Tier M SPEC은 실제로 0개이고, 현장의 관행은 AC를 spec.md에 복제하는 쪽이며, 그 복제가 lint를 만족시키고 있다.** 결함 ①은 활성이 아니라 잠복이다.

### 1.3 관행과 린터 중 어느 쪽이 옳은가

증거는 관행 쪽을 가리킨다.

- `.claude/skills/moai/workflows/plan/spec-assembly.md:55`는 "AC is inline in `spec.md §3`"을 **Tier S의 성질로** 적는다. 56-57행(Tier M / Tier L)은 `acceptance.md`를 요구하며 spec.md AC 요건을 두지 않는다.
- `.claude/rules/moai/development/manager-develop-prompt-template.md:131`은 AC 계수의 **SSOT를 `acceptance.md`로** 못박는다(명시적으로 `progress.md`가 아니라고 적는다).

린터는 Tier S의 전제를 전 Tier에 적용하고 있고, SSOT가 아니라 관습적 복사본을 읽고 있다. 다만 반대 방향(Tier M/L도 spec.md에 AC 블록을 두도록 관행을 고친다)도 성립 가능한 선택지이며, 그 경우 위 SSOT 조항도 함께 고쳐야 한다. 방향 결정은 plan.md §C에서 다룬다.

---

## 2. 요구사항 (GEARS)

> **자기참조 주석**: 이 SPEC은 REQ 줄 파싱을 고치는 문서이므로, 아래 REQ 줄은 **`parseREQs`가 현재 받아들이는 형태**(`- REQ-CRS-001-NNN: ...`)로 적었다. 넓힌 파서가 아직 없는 상태에서 이 SPEC이 자기가 고치려는 규칙에 보이지 않는 문서가 되면 곤란하기 때문이다. 이 자기참조가 바로 요점이다.

### 2.1 결함 ② — REQ 파서 (본체)

- REQ-CRS-001-001: `parseREQs` 함수는 코퍼스에서 실제로 쓰이는 REQ ID 형태 — 두 분절 이상의 도메인, 도메인 안의 숫자, 3-분절 및 5-분절 — 를 REQ 정의 줄로 인식해야 한다(shall).
- REQ-CRS-001-002: **Where** 추출 패턴이 넓어진 상태에서, `reqIDPattern`(검증)과 `reqLinePattern`(추출)은 같은 ID 집합을 지시해야 한다(shall) — 넓힌 추출이 `InvalidREQIDRule`을 코퍼스 전역에서 발화시켜서는 안 된다.
- REQ-CRS-001-003: **When** 넓힌 파서를 적용한 lint를 코퍼스 전체에 대해 실행할 때, 그 실행은 `develop` 브랜치를 적색으로 만드는 미해소 `error` 등급 finding을 남겨서는 안 된다(shall not) — 심각도 하향, 자문(advisory) 처리, 동일 SPEC 내 코퍼스 정리, 범위 축소 중 하나의 수단으로 이를 보장한다.
- REQ-CRS-001-004: 넓힘의 파급 규모는 실제 Go 파서로 재측정해야 한다(shall). 계획 단계의 추정치(741건)는 근사이며 baseline으로 쓰지 않는다.
- REQ-CRS-001-005: `parseREQs` 함수는 인식 대상 형태의 REQ 정의 줄을 조용히 버려서는 안 된다(shall not) — 인식 실패는 관측 가능한 신호로 드러나야 하며, 빈 결과가 통과로 읽혀서는 안 된다.

### 2.2 결함 ① — 형제 AC 파일 판독 (잠복)

- REQ-CRS-001-006: **When** `CoverageRule`이 AC 집합을 수집할 때, 규칙은 해당 SPEC 디렉터리의 형제 파일 `acceptance.md`에 선언된 AC를 포함해야 한다(shall).
- REQ-CRS-001-007: **When** 어떤 REQ를 참조하는 AC가 spec.md와 `acceptance.md` 어디에도 존재하지 않을 때, `CoverageRule`은 여전히 `CoverageIncomplete`를 내고 종료코드 1을 반환해야 한다(shall).
- REQ-CRS-001-008: 이 SPEC의 수리는 Tier M/L SPEC이 spec.md에 AC를 중복 기재하도록 요구해서는 안 된다(shall not) — `acceptance.md`를 AC의 SSOT로 두는 기존 규약을 보존한다.

### 2.3 Acceptance Criteria 대응표 (결함 ① 때문에 강제된 중복)

> **이 표는 관행 위반이며, 그 사실 자체가 증거다.** 위 REQ들의 AC 정본은 `acceptance.md`이고 이 SPEC은 Tier M이므로, 규약대로라면 여기에 AC를 다시 적을 이유가 없다. 그러나 작성 직후 이 트리의 바이너리로 lint를 돌린 결과 `CoverageIncomplete` 8건이 나왔다 — `acceptance.md`를 읽지 않기 때문이다. 즉 **이 SPEC은 자기가 옹호하는 규약을 따르면 코퍼스를 붉힌다.** 아래 표는 착지 시 develop을 붉히지 않기 위한 최소 중복이며, 결함 ①이 수리되면 삭제 대상이다(REQ-CRS-001-006 참조).

- AC-CRS-001-001: 여섯 가지 ID 형태 인식 (maps REQ-CRS-001-001)
- AC-CRS-001-002: 전 코퍼스 수집량 실측 (maps REQ-CRS-001-004)
- AC-CRS-001-003: `InvalidREQID` 오탐 부재 (maps REQ-CRS-001-002)
- AC-CRS-001-004: 조용한 누락 없음, 뮤테이션 RED 확립 (maps REQ-CRS-001-005)
- AC-CRS-001-005: 착지 시 미해소 error 0건 (maps REQ-CRS-001-003)
- AC-CRS-001-006: 회귀 쌍 — 관행 픽스처 PASS + 진짜 부재 픽스처 rc=1 (maps REQ-CRS-001-006, REQ-CRS-001-007)
- AC-CRS-001-007: `acceptance.md` 부재 Tier S 픽스처 무사 통과 (maps REQ-CRS-001-006)
- AC-CRS-001-008: AC 중복 기재를 새로 요구하지 않음 (maps REQ-CRS-001-008)

---

## 3. 제약

- 대상 패키지는 `internal/spec` 하나다. 다른 패키지의 공개 API는 건드리지 않는다.
- `SPECDoc`은 spec.md 하나만 담는다. 형제 산출물을 읽는 규칙은 이 패키지에 이미 세 건의 선례(`MovingRefUnpinnedRule`, `HaikuResidualRule`, `ArtifactStatusFieldForbiddenRule`)가 있으므로 새로운 발명이 아니다.
- `lint.skip`과 era 강등(`eraDemotableCodes`)의 기존 의미론을 깨지 않는다.
- 검증 범위는 건드린 패키지로 한정한다. 전 패키지 판정은 CI 몫이다.

---

## 4. Exclusions — 이 SPEC이 만들지 않는 것

이 절은 out of scope 항목을 명시한다.

### Out of Scope — REQ 줄 규약 자체의 코퍼스 전면 정리

- 넓힌 패턴으로도 702개 중 62개(8.8%)만 REQ 정의 줄을 갖는다. 나머지는 REQ 토큰을 산문과 표에서 참조할 뿐 정의 줄로 선언하지 않는다. 이 간극의 상당 부분은 파서 버그가 아니라 규약과 관행의 어긋남이며, 전 코퍼스를 규약에 맞춰 다시 쓰는 일은 이 SPEC에 담지 않는다.
- 단, REQ-CRS-001-003을 만족시키기 위해 필요한 **범위 한정 정리**는 예외로 이 SPEC에 포함될 수 있다. 그 판단은 plan.md §D의 심각도 결정에 종속된다.

### Out of Scope — `MissingExclusions` (가설은 반증됨, 수리는 범위 밖)

- 조사 초기 가설은 "`## 2. Out of Scope`의 숫자 접두를 파서가 잘못 다룬다"였다. **이 가설은 반증되었다.** `OutOfScopeRule.Check`(`lint.go:896`)는 `###` 접두 + "out of scope" 포함 줄에서만 `inOutOfScope`를 켠다. `## 2. Out of Scope`는 H2라서 절에 진입하지 못하고 `hasContent`가 false로 남는다. 숫자 접두는 무관하다.
- 근거: 이 SPEC의 §4는 `## 4. Exclusions`(숫자 접두 H2) 아래 `### Out of Scope — …` H3를 두었고, lint가 `MissingExclusions` 없이 rc=0으로 통과했다.
- 코퍼스 형태(같은 트리 측정): **H3 형태 684개 / H2뿐 2개.** 실제 원인은 규약(H3 하위 표제 필수)과 일부 SPEC의 H2 관행이 어긋난 것이다. 코퍼스 정리는 이 SPEC의 소관이 아니다.

### Out of Scope — AC 절 탐지의 세 번째 협소성

- 조사 중 같은 파싱 경로에서 세 번째 협소성이 관측되었다. `findACSectionStart`(`parser.go:64`)는 **문서에서 처음 나오는** `##` 접두 + "acceptance" 포함 줄을 AC 절 시작으로 잡고, `extractACLines`는 다음 `##`에서 멈춘다. 따라서 본문에 `acceptance.md`를 언급하는 표제가 먼저 나오면 그 표제가 AC 절을 가로채고 즉시 종료되어 **AC가 0개로 파싱된다.**
- 이 SPEC 작성 중 실제로 발생했다(§HISTORY 참조). 표제 문구를 바꿔 우회했다.
- **파급은 측정했고, 최초 추정은 틀렸다.** 같은 트리에서 잰 값(`.moai/reports/t362/ac_section_reach.py`, `cross_check.py`): `spec.md` 703개 중 파서가 AC 절에 도달하는 것 326개, **도달 못했는데 AC 토큰은 보유한 것 0개**, 한국어 AC 표제 보유 47개. 넓은 REQ 패턴에 걸리는 63개 중 미도달 16개는 **전부 AC 토큰 0개**다.
- 따라서 파서가 진입하지 못하는 SPEC은 있는 AC를 잃는 것이 아니라 spec.md에 AC를 두지 않을 뿐이다. 741 시뮬레이션은 `maps REQ-…` 출현으로 커버리지를 셌으므로 이들의 REQ는 이미 미커버로 계산돼 있다 — **추정치 밖의 증분이 아니라 안이다.**
- 한국어 표제 47개도 같다. "숨은 증분"이 아니다.
- 수리는 이 SPEC 범위에 넣지 않는다. 협소성 자체는 실재하므로 기록만 남긴다.

### Out of Scope — 다른 lint 규칙군

- `MovingRefUnpinned`(113건), `StatusGitConsistency`(18건), `FrontmatterInvalid`(14건) 등 코퍼스의 다른 finding 계열은 이 SPEC의 소관이 아니다. 각각 별도 카드가 있거나 필요하다.

---

## 5. Gaps — 관측하지 않은 것

- **"13건"이라는 전달값의 출처.** 리드가 lane-1을 통해 manager-spec으로부터 전달받은 값이며, 누구도 그 출력을 관측하지 않았다. 이 트리의 측정값은 0이다. 미검증 전달값으로만 취급한다.
- **`MissingExclusions` 가설.** 이제 미검증이 아니다 — 코드로 반증했고 코퍼스 형태(H3 684 / H2뿐 2)도 쟀다(§4). 다만 H2뿐인 2건이 실제로 finding 24건 전부의 원인인지는 개별 대조하지 않았다.
- **741의 실제 오차 크기.** 방향은 하한으로 예측되지만(plan.md §B 관측 3) 크기는 재지 않았다. M1이 실제 Go 파서로 잰다.
- **Tier별 발현 차이.** Tier L(89개) / Tier S(113개) SPEC이 결함 ①을 다르게 드러내는지 확인하지 않았다. 15/702 분할은 Tier가 아니라 코드 경로(REQ 줄 존재 여부)를 기준으로 갈렸으므로, Tier 축의 분석은 별도 측정이 필요하다.
- **741이라는 추정치의 실측 대응값.** Python 근사(`.moai/reports/t362/widen_sim.py`)이며 실제 Go 파서로 재지 않았다. REQ-CRS-001-004가 이 간극을 닫는다.
