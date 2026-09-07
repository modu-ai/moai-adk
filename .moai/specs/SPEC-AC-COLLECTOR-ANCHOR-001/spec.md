---
id: SPEC-AC-COLLECTOR-ANCHOR-001
title: "인라인 AC 수집기의 항목 문법 앵커가 코퍼스 선언 줄의 81.5%를 거절한다"
version: "0.1.0"
status: completed
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/spec"
lifecycle: spec-anchored
tags: "parser, acceptance-criteria, spec-view, internal/spec, t528"
tier: M
---

# SPEC-AC-COLLECTOR-ANCHOR-001 — 인라인 AC 수집기 항목 문법 앵커

## HISTORY

### 2026-09-08 — 최초 작성 (카드 t528 실측 기준)

카드 `t528`에서 열렸다. 기준선은 워크트리 `.claude/worktrees/t528`, 브랜치 `WT-ac-collector-anchor`, HEAD `52f863f36`(측정 시점 `origin/develop`과 동일)에서 임시 프로브(`internal/spec/zz_t528_probe_test.go`, 커밋하지 않음)로 직접 쟀다. 그 1차 측정의 기록은 `.moai/reports/t528/measurement-20260908.md`다. **이후 D2' 정정으로 증거 정본은 `.moai/reports/t528/probe/run-20260908.txt`(판별식 B, `EXIT=0`)로 이동했고**, `measurement-20260908.md`는 판별식 A의 폐기된 1차 기준선 기록으로만 남는다(§1.1). 어느 쪽이든 수치는 산출물에서 인용한 것이며 재유도하거나 반올림하지 않았다.

측정에서 먼저 세운 것은 **수집기가 실제로 돈다**는 사실이다. 코퍼스 806개 `spec.md` 중 18개는 지금도 AC를 정상 파싱한다(`SPEC-ARTIFACT-STATELESS-001` → `AC-AST-001-01`, `SPEC-CLIFIX-CONCURRENCY-001` → `AC-CONC-001-001`). 이 양성 대조군이 있기 때문에 나머지의 0은 **미실행이 아니라 거절**로 읽을 수 있다.

### 2026-09-08 — 1차 분류의 정정을 승계한다

1차 측정은 「헤딩 없음 + AC 토큰 보유 = 37 파일」을 blind로 셌으나, 표본을 열어보니 대부분 본문 산문 언급이었다(`AC-BH-006's.` 같은 줄). needle이 불릿을 요구하지 않아 산문을 선언으로 센 것이다. 불릿 필수로 좁혀 재측정한 값이 이 SPEC이 인용하는 수치다. 이 정정은 간극을 **넓히는** 방향이므로 기록으로 남긴다.

### 2026-09-08 — 범위를 항목 문법 한 축으로 한정한다 (운영자 판정)

같은 파싱 경로에는 축이 둘 있다. **섹션 스코핑**(`findACSectionStart`가 `##…acceptance` 헤딩을 요구)과 **항목 문법**(`parseSingleACLine`의 ID·구분자 앵커). 운영자는 2026-09-08에 이 카드를 **항목 문법 축 하나**로 한정했다. 섹션 스코핑은 그대로 둔다.

### 2026-09-08 — 감사 FAIL(iteration 1)의 지적을 반영한다

`.moai/reports/t528/plan-audit.md`가 MP-7 + blocking 결함 6건으로 FAIL을 냈다. 이 개정이 반영한 것은 다음과 같으며, 새 근거는 전부 `.moai/reports/t528/probe/` 아래 산출물이다.

- **프로브를 트리에 남겼다.** 기준선 판별식이 삭제된 코드 안에만 있어 제3자가 216/1160을 재유도할 수 없었다(D1). 이제 프로브 소스(`probe/ac_anchor_probe_test.go`)와 종료 코드가 붙은 실행 출력(`probe/run-20260908.txt`, `EXIT=0`)이 기록에 있다.
- **분모를 산출물로 고정했다.** `find`를 다시 돌려 세는 대신 프로브가 읽은 파일 목록 807줄을 `probe/filelist.txt`로 쓴다(D6).
- **18개 양성 대조군의 구성원 목록을 남겼다.** `probe/positive-needle.txt`. 종전에는 수만 있고 목록이 없어 대조군이 재현 불가였다(D2).
- **§B.1의 열린 결정을 실측으로 닫았다.** 아래 §1.4.
- **§4의 「가드를 약화시키지 않는다」 주장을 철회했다.** 아래 §4. 인용한 출처가 콜론 형식을 산문 가드의 절반으로 지목하고 있었으므로, 그 인용문 옆에 반대 결론을 둔 것이 잘못이다(D4).
- **계측 지점 두 개를 옮겼다.** `ValidateDepth` / `DuplicateAcceptanceID`는 lint 경로에서 관측할 수 없고(§1.3), 진짜 붉어지는 자리는 CLI다(D3).
- **과수용을 재는 요구를 추가했다**(REQ-ACA-001-012). 넓힘의 성공 지표와 실패 지표가 같은 방향이라 통상 GREEN이 판별하지 못한다(D5).

### 2026-09-08 — 판별식 드리프트 정정 (D1 재발 사례를 기록으로 남긴다)

재측정이 1167을 내놨고 기준선은 1160이었다. 이것을 **코퍼스가 커졌다**로 읽을 뻔했다. 두 판별식을 한 프로세스에서 같은 트리에 태워 분리했다(`probe/drift_probe_test.go`, 출력 `probe/drift-20260908.txt`, `EXIT=0`):

```
discriminator A (digit-final), own card INCLUDED : files=807 decls=1160
discriminator B (letter/dot ok), own card INCLUDED: files=807 decls=1167
discriminator A, own card EXCLUDED               : files=806 decls=1160
=> corpus effect of own card (A): files +1 decls +0
=> discriminator effect A->B (own excluded): decls +7
```

여기서 나오는 결론 셋을 못박는다.

1. **own-card 오염은 없다.** 이 카드 자신의 `spec.md`가 코퍼스에 들어온 효과는 파일 +1, 선언 **+0**이다. (종전 판본은 이 자리에 「기준선 216/1160은 유효하다」라고 적었다 — **그 문장은 아래 2026-09-08 D2' 정정으로 대체됐다.** 정본 기준선은 판별식 B의 216/1167이며, `1160`은 이 항목 안에서만 판별식 A의 관측으로 남는다. own-card 무오염이라는 사실 자체는 두 판별식 모두에서 성립한다 — 위 출력의 A·B 두 행이 그것을 보인다.) 감사가 지적한 806→807 드리프트(D6)는 실재하지만 **분모 축에만** 닿고 선언 수를 움직이지 않는다.
2. **1160과 1167은 비교 가능한 수가 아니다.** 두 수의 차 7은 코퍼스 변화가 아니라 판별식 변화(A→B)다. 나란히 적으면서 「늘었다」고 읽는 순간 그것은 측정이 아니라 착시다.
3. **[HARD] 이후 모든 재측정은 트리 SHA와 함께 판별식을 핀한다.** 판별식을 적지 않은 비교는 무의미하다(REQ-ACA-001-016).

이 사건은 감사가 D1로 이름 붙인 위험이 실제로 한 번 더 재현된 것이다. 정정 방향이 「기준선이 흔들린다」에서 「기준선이 굳는다」로 갔으므로 좋은 소식이지만, **좋은 방향이라서 기록을 생략하지 않는다.**

### 2026-09-08 — 중복 ID 노출 실측: 현재 코퍼스에 재료가 **없다**

D3이 지목한 위험(넓힘이 새 `DuplicateAcceptanceID`를 만들어 `moai spec view`를 하드 에러로 바꾼다)의 **재료를 실측했다.** 명령 `go test ./internal/spec/ -run TestT528DuplicateExposure -v -count=1 -timeout 600s`, `EXIT=0`, 출력 `probe/duplicate-20260908.txt`, 프로브 `probe/duplicate_probe_test.go`:

```
TODAY   (current parser grammar) : files with duplicate ids = 0, dropped lines = 0
WIDENED (B.1 candidate grammar)  : files with duplicate ids = 0, dropped lines = 0
```

**측정된 사실**: §B.1 후보 문법은 현재 코퍼스에서 중복 ID를 **하나도 만들지 않는다.** 따라서 이 변경만으로는 `internal/cli/spec_view.go:80-86`의 `default: return fmt.Errorf` 경로가 발화하지 않는다 — 운영자가 정한 D3 통과 조건(하드 에러 SPEC 수 0)은 **오늘 달성 가능한 상태로 측정됐다.** 희망이 아니라 실측이다.

AC와 halt-and-report 규칙은 결정된 그대로 둔다. 재료가 0이라는 사실이 가드를 불필요하게 만들지 않는다 — run 단계 구현이 §B.1 후보와 다르게 넓히면 이 수는 달라지고, 그때 가드가 살아 있어야 한다. **다만 0은 그 자체로 새 위험을 만든다** — 아래 §1.3의 양성 대조군 요구가 그 답이다.

### 2026-09-08 — [HARD] 정정: 「4 파일 / 24 줄」은 거짓 수치였다 (판별식 오형성 3번째 사례)

중복 노출의 1차 측정은 **「4 파일 / 24 dropped lines」**를 보고했다. **그 수는 틀렸다.** 스캔 정규식의 ID 문자류가 `.`을 빠뜨리고 마지막 문자를 숫자로 요구해, 서로 **다른** 하위 ID를 같은 ID로 접었다 — `AC-LCLN-001.1` / `.2` / `.3`에서 `AC-LCLN-001`을, `AC-HFC-001a` / `001b`에서 `AC-HFC-001`을 잘라낸 것이다. 코퍼스 원문:

```
112:- **AC-LCLN-001** (top-level): `moai agent lint --strict` exit 0 on `main` HEAD after all phases…
113:  - AC-LCLN-001.1: Baseline captured before each LCLN-Phase.
114:  - AC-LCLN-001.2: Each LCLN-Phase reduces a contiguous subset of findings…
90:- **AC-HFC-001a** (REQ-HFC-001): **Given** a `PostToolUseFailure` `HookInput`…
91:- **AC-HFC-001b** (REQ-HFC-001): **Given** a `PostToolUseFailure` payload…
```

이 ID들은 코퍼스에서 **서로 다른 항목**이고, 게다가 §1.4가 분해한 미포함 39건에 속해 §B.1 후보가 애초에 받아들이지도 않는다. 중복은 코퍼스에 있던 것이 아니라 **판별식이 만들어낸 것**이다.

#### 이것이 이 카드에서 다섯 번째다 — 그리고 run 단계에서 여섯 번째가 나왔다

같은 실패 부류 — **코퍼스에 맞지 않게 형성된 판별식이 거짓 수치·거짓 분류를 낸다** — 가 이 카드를 재는 동안 여섯 번 재현됐다. 4·5행은 감사 iteration 2(`.moai/reports/t528/plan-audit-2.md`, D1'·D2')가 잡았고, 6행은 run 단계 M0이 잡았다(아래 2026-09-08 항목).

| # | 거짓 수치 / 거짓 분류 | 판별식의 결함 | 정정 |
|---|---|---|---|
| 1 | 선언 1160 → 1167 「증가」 | 두 판별식(A/B)의 값을 비교 가능한 것으로 취급 | 한 프로세스에서 둘 다 태워 분리 |
| 2 | 이 카드 자신의 SPEC이 선언 +7 기여 | 같은 원인 — 판별식 차이를 코퍼스 효과로 귀속 | own-card 포함/제외 실측: 파일 +1, 선언 **+0** |
| 3 | 중복 「4 파일 / 24 줄」 | ID 문자류에서 `.` 누락 + 숫자-종료 요구 → 하위 ID 접힘 | §B.1 후보 문법 자체를 admission 판정으로 사용 |
| 4 | 「단어 꼬리 7건은 선언이 아닌 개념 라벨」 | 판별식이 아니라 **분류**의 오형성 — 코퍼스 줄을 열지 않고 형태 이름만 보고 처분을 결정 | 7개 형태를 `/usr/bin/grep`으로 전수 열람 → 전부 진짜 선언. 처분을 「진짜 선언, 다른 축」으로 정정(§1.4) |
| 5 | 헤드라인 `944` / `100` / 「기준선 판별식 = A」 | 보존된 프로브는 B인데 문서가 A라고 적음 — 판별식 라벨과 산출물의 소유자 불일치 | 정본 기준선을 **B**(216 / 1167 / 951 / 101)로 재선언, A 계열은 「삭제된 프로브의 기록, 재유도 불가」로 강등(§1.1) |
| 6 | `SWEPT=807 NONZERO_EXIT=807` — 코퍼스 전량이 하드 에러라는 판정 | 판별식이 아니라 **호출 형태**의 오형성. 어느 빌드도 받지 않는 플래그(`--acceptance`)를 문서가 명령으로 지목했고, 스윕이 그 형태를 그대로 실행했다. 807은 수집기가 아니라 **인자 파싱**을 잰 수다 | run 단계 M0이 오류 본문을 열어 `Unknown flag: --acceptance`를 읽고 형태를 정정. 실제 명령(`moai spec view <ID>`)으로 재측정해 442를 세웠다. 플래그를 halt 조건에 써넣은 것은 레인이므로 **문서만의 잘못이 아니라 레인의 잘못**이다 |

**4행은 특히 아프게 남긴다.** 이 분류는 「코퍼스 수치는 그 뒤의 표본 줄을 읽기 전까지 채택하지 않는다」를 [HARD] 규율로 승격한 **바로 그 개정에서** 표본 줄을 읽지 않고 채택됐다. 규율을 만든 문서가 같은 revision 안에서 그 규율을 어겼다. 규율의 존재는 준수의 증거가 아니다.

**이 카드가 고치려는 결함이 정확히 이것이다** — 코퍼스가 실제로 쓰는 모양에 맞지 않는 앵커. 그것을 측정하는 동안 같은 부류를 여섯 번 재생산했다. 이 기록은 자기비판이 아니라 **증거다**: `acceptance.md`의 뮤턴트 규율과 대조군 규율이 의례가 아니라는 가장 강한 근거가 여기 있다. 판별식은 자연히 이렇게 어긋나며, 어긋난 판별식은 **자신 있게 통과한다.**

#### 여섯 번 모두 같은 방법으로 잡혔다

여섯 중 어느 것도 **수를 다시 읽어서** 잡히지 않았다. 여섯 다 **코퍼스의 실제 줄 또는 산출물 원문을 열어서** 잡혔다(4행은 `/usr/bin/grep` 전수 열람, 5행은 프로브 소스와 출력 파일 직독, 6행은 스윕이 뱉은 오류 본문 직독 — 807이라는 수는 아무리 다시 읽어도 오형성을 드러내지 않았다). 그래서 이것을 규율로 승격한다 — `plan.md` §D: **코퍼스 수치는 그 뒤의 표본 줄을 읽기 전까지 채택하지 않는다.**

### 2026-09-08 — 감사 iteration 2(PASS-WITH-DEBT)의 debt gate 2건을 해소한다

`.moai/reports/t528/plan-audit-2.md`가 **PASS-WITH-DEBT 0.84**를 냈고, critical blocking 2건(D1'·D2')을 `internal/spec/parser.go` 첫 편집 **이전에** 해소할 조건으로 걸었다. 이 개정이 그 둘을 갚는다. 새 측정은 하지 않았다 — 필요한 실측은 전부 감사 보고서의 Evidence 절과 `.moai/reports/t528/probe/` 산출물에 이미 있다.

- **D1' 해소** — §1.4의 단어 꼬리 7건 처분을 「선언이 아닌 개념 라벨 / 영구 배제」에서 **「진짜 선언, 다른 축, 이 카드 밖」**으로 정정하고 7줄을 verbatim 인용했다. 결정(숫자 꼬리 고정)은 유지하되 근거를 실측 가능한 것으로 교체했다. §4의 「영구 배제」는 철회했다.
- **D2' 해소** — 정본 기준선 판별식을 **B**로 재선언했다. 보존된 프로브(`probe/ac_anchor_probe_test.go:30`)가 B이고, D1'가 B의 충실도를 세웠기 때문이다. 헤드라인 수치는 **216 / 1167 / 951 / 101**(분모 807)이 되고, A 계열 `944` · `100`은 재유도 불가로 강등됐다.
- **판정 핀 주석**: 감사 verdict는 **정정 이전** 해시에 대해 내려졌다 — `spec.md` `86f81c26…`, `plan.md` `328c7081…`, `acceptance.md` `40681504…`. 이 개정이 세 해시를 전부 움직인다. 감사자는 해소 판정이 이진이므로 재감사를 요구하지 않았다(`plan-audit-2.md` § Recommendation 2).
- 함께 처리한 optional 2건: `plan.md` §I·`spec.md` §7의 `parser.go` 인용 줄번호 ±2 정정(D4'), Tier M 상한 동시 도달의 「알려진 제약」 기록(D5').

### 2026-09-08 — [HARD] 정정: 존재하지 않는 플래그를 명령으로 지목했다 (오형성 6번째 사례)

`spec.md` §1.3, `plan.md` §E M7, `acceptance.md` §D.12·§D.15가 모두 `moai spec view <ID> --acceptance`를 명령으로 적었다. **그 플래그는 없다.** `internal/cli/spec_view.go:37`이 등록하는 플래그는 `--shape-trace` 하나뿐이며, acceptance 뷰는 `moai spec view <ID>`의 기본 동작이다. 세 파일에서 명령으로 쓰인 자리 13곳(`spec.md` 4 · `plan.md` 4 · `acceptance.md` 5)을 실제 형태로 고쳤다.

**어떻게 드러났나.** 문서가 적은 형태로 코퍼스를 훑으면 `SWEPT=807 NONZERO_EXIT=807`이 나온다 — 807개 전량 실패. 자신 있고 균일한 수이지만, 그것이 잰 것은 수집기가 아니라 **인자 파싱**이다. 실행 하나하나가 `Unknown flag: --acceptance` / `RC=1`이었다.

**어떻게 잡혔나.** run 단계 M0이 807이라는 수를 채택하지 않고 **오류 본문을 열어 읽었다.** 수를 다시 읽는 것으로는 영영 드러나지 않았을 결함이다 — 807/807은 「전량 결함」으로도 완벽하게 읽힌다. 증거: `.moai/reports/t528/probe/before/specview-before-20260908.md`.

**소관.** 이 플래그는 run 단계 halt 조건에 레인이 써넣었다. 따라서 문서만의 잘못이 아니라 **레인의 잘못**이며, 위 정정 표 6행에 그렇게 기록한다.

### 2026-09-08 — [HARD] 정정: `spec view` 하드 에러 before-image는 0이 아니라 442다

같은 M0 측정이 두 번째 carried-over 0을 드러냈다. `plan.md` §E M7과 `acceptance.md` §D.15가 하드 에러 before-image를 **0**으로 못박고 `probe/duplicate-20260908.txt`를 근거로 댔으나, 그 산출물이 잰 것은 **중복 ID 재료**(진짜로 0)이지 CLI 하드 에러가 아니다. 서로 다른 두 양이고, 한쪽의 0이 다른 쪽으로 옮겨 적혔다.

실제 명령으로 코퍼스 전체(분모 807, 트리 `52f863f36`)를 돌린 값:

```
SWEPT=807  NONZERO_EXIT=448
  parse error:            442   전부 "acceptance criteria section not found"
  spec.md not found for:    6   _archive/ 경로 — 파서에 도달하지 않음
```

442건 전부가 **헤딩 축**의 기존 조건이며, 이 카드가 §4에서 건드리지 않기로 한 축이다. 항목 문법 넓힘은 이 수를 움직이지 않는다. `_archive/` 6건은 CLI가 bare SPEC-ID로 주소를 잡지 못해 생기는 것으로, 파서에 닿지 않으므로 하드 에러 집계에서 제외한다.

**따라서 REQ-ACA-001-013의 게이트는 총량이 아니라 442 대비 델타다** — 기준선에 **없던** 하드 에러가 0건이어야 하고, 넓힘이 새로 만들 수 있는 중복/깊이 부류는 그 안에서 따로 감시한다. 이 읽기는 측정 뒤 레인이 승인했다. 문자 그대로의 「0이어야 한다」를 고수하면, 이 카드가 손대지 않기로 한 축의 기존 조건 때문에 카드가 멈춘다 — 카드가 만들지도 고칠 수도 없는 사유로.

이 정정의 방향은 **간극을 넓힌다**(0으로 알던 자리가 442였다). 좋은 방향이 아니라서 더더욱 기록한다.

---

## 1. 배경과 문제

`internal/spec/parser.go`의 `ParseAcceptanceCriteria`는 `spec.md`의 인라인 Acceptance Criteria를 읽는다. 두 개의 앵커가 순서대로 걸린다.

1. `findACSectionStart` — `##` 접두 + 소문자화 시 `acceptance`를 포함하는 헤딩을 요구한다. **이 축은 이 SPEC의 범위 밖이다**(§4).
2. `parseSingleACLine`(`internal/spec/parser.go:218`) — `strings.TrimLeft(trimmed, "- *")` 이후 다음 정규식과 일치할 것을 요구한다.

```go
acIDPattern := regexp.MustCompile(`^(AC-[A-Z0-9]+-[0-9]+-[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?)\s*:\s*`)
```

즉 **정확히 3분절**이고 2·3번째 분절이 숫자인 ID, 그리고 그 직후에 곧바로 붙는 콜론. 코퍼스가 실제로 쓰는 모양은 이것이 아니다.

### 1.1 실측 기준선

[HARD] 정본 기준선은 **판별식 B**다. 판별식을 적지 않은 수는 이 표에 없다.

- **판별식 B** (정본 기준선 판별식) — `^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9.-]*[A-Za-z0-9])\*{0,2}\s*(.*)$`. 불릿 필수, 점(`.`) 허용, ID가 문자로 끝나도 받는다. 정본 소스는 `probe/ac_anchor_probe_test.go:30`이고 트리에 커밋돼 있다.
- **판별식 A** (폐기된 1차 판별식) — `^\s*[-*+]\s+\*{0,2}(AC-[A-Za-z0-9-]*[0-9])\*{0,2}\s*(.*)$`. 점 불가, ID 마지막 문자가 숫자. 이 판별식을 태운 프로브(`zz_t528_probe_test.go`)는 트리에서 삭제됐다.

**왜 B가 정본인가 — 편의가 아니라 충실도다.** §1.4가 실측으로 세운 사실은, A가 거절하고 B가 받는 단어 꼬리 7건이 **전부 진짜 AC 선언**이라는 것이다. 즉 두 판별식 중 코퍼스가 실제로 쓰는 모양에 더 가까운 쪽은 B다. 기준선을 A에 두면 이 카드는 자기가 고치려는 결함 — 코퍼스에 맞지 않게 형성된 앵커 — 을 자기 계측 도구 안에 그대로 두고 시작하게 된다.

| 관측 | 값 | 판별식 | 산출물 |
|---|---|---|---|
| 프로브가 읽은 `spec.md` (분모) | 807 | — | `probe/filelist.txt` |
| AC 절 **안**에 있는 AC 선언 줄(불릿) | 1167 | B | `probe/run-20260908.txt` |
| 현행 파서가 수용 | 216 | A·B 공통 | `probe/run-20260908.txt` |
| 현행 파서가 거절 | 951 | B | `probe/run-20260908.txt` |
| 도달률 | 18.5% | B | `probe/run-20260908.txt` |
| in-section 선언 ≥1 인데 파싱 결과 0인 `spec.md` | 101 | B | `probe/run-20260908.txt` |
| 파싱 결과 ≥1 (양성 needle 대조군) | 18 | A·B 공통 | `probe/positive-needle.txt` |
| 숫자 꼬리 후보가 덮는 선언 / 남기는 선언 | 1128 / 39 | B | `probe/run-20260908.txt` |

명령: `go test ./internal/spec/ -run TestT528Anchor -v -count=1 -timeout 600s`. verbatim 출력은 `.moai/reports/t528/probe/run-20260908.txt`이며 말미에 `EXIT=0`이 붙어 있다.

**`216`과 양성 needle 18은 두 판별식에서 동일하다.** A가 잡는 줄은 B도 잡고, 수용 판정은 현행 파서 정규식이 줄 전체에 대해 내리므로 캡처 폭과 무관하다. 즉 이 정정은 **분모 축(1160↔1167)과 그로부터 유도되는 거절·blind 수**에만 닿는다 — 대조군과 수용량은 그대로다. 정정을 실제보다 넓게 읽지 않기 위해 적는다.

**판별식 A 계열 수치의 처분 — 강등.** A로만 잰 수치 중 트리의 보존 산출물로 재유도되는 것은 `1160`(`probe/drift-20260908.txt`) 하나뿐이다. `944`(거절) · `100`(blind) · `180`(절 밖 선언) · `360`(`##…acceptance` 헤딩 보유) · `806`(own-card 제외 분모) · 구분자 `82`는 **삭제된 프로브의 기록이며 재유도 불가**다. 이들은 `.moai/reports/t528/measurement-20260908.md`에 남아 있으나 **기준선으로 쓰지 않는다**. 이 문서에서 A 계열 수치가 살아 있는 자리는 HISTORY의 드리프트 정정 항목과, 아래 §4·§5에서 「A 계열·재유도 불가」라고 명시적으로 라벨된 관측뿐이다.

**분모는 이제 명령이 아니라 산출물이 고정한다.** 코퍼스는 자기 수정적이다 — 이 카드의 `spec.md`를 쓰는 행위 자체가 `spec.md` 하나를 더했다. 그래서 프로브가 **자기가 읽은 파일 목록**을 `probe/filelist.txt`(807줄)로 쓴다. 나중에 `find`를 다시 돌려 806↔807을 만나면 그것이 「넓힘의 효과」인지 「분모의 이동」인지 갈리지 않는다. 목록이 있으면 갈린다.

### 1.4 §B.1 결정의 실측 근거 — 숫자 꼬리 규칙이 무엇을 덮고 무엇을 남기는가

`plan.md` §B.1이 열어 두었던 결정(마지막 분절을 숫자로 고정하는 것이 관측된 줄 전부를 덮는가)을 실측으로 닫았다. 명령·출력은 `probe/run-20260908.txt`(`EXIT=0`), 판별식 B, 트리 `52f863f36`, 분모 `probe/filelist.txt` 807.

```
B.1 numeric-tail candidate: covered = 1128  UNCOVERED = 39
```

미포함 39건의 내역(verbatim, 숫자를 `N`으로 정규화):

| 부류 | 형태와 건수 | 합 | 처분 |
|---|---|---|---|
| 단어 꼬리 | `AC-ORDERING` 1 · `AC-MRR-GREEN` 1 · `AC-MUTANT` 1 · `AC-GREEN` 1 · `AC-EVIDENCE` 1 · `AC-SCOPE` 1 · `AC-HFC-GATE` 1 | 7 | **진짜 선언 — 다른 축, 이 카드 밖** |
| 알파벳 접미 | `AC-UTIL-N-Nb` 1 · `AC-HFC-Na` 1 · `AC-HFC-Nb` 1 | 3 | 진짜 선언 — 다른 축, 이 카드 밖 |
| 점 하위번호 | `AC-LCLN-N.N` 10 | 10 | 진짜 선언 — 다른 축, 이 카드 밖 |
| 범위 표기 | `AC-SVG-N..N` 5 · `AC-DCC-N..N` 4 · `AC-HML-N..N` 4 · `AC-PLW-N..N` 2 | 15 | 다른 축 — 문법이 아니라 의미론 문제 |
| 문자 접두 분절 | `AC-MRR-GN` 4 | 4 | 진짜 선언 — 다른 축, 이 카드 밖 |

**결정: 마지막 분절을 숫자로 고정한다.** 결정은 그대로이고, **처분은 정정한다.**

#### [HARD] 정정 — 단어 꼬리 7건은 「선언이 아닌 개념 라벨」이 아니다

종전 판본은 이 7건을 **「다른 SPEC 본문이 AC 절 안에서 개념 라벨로 쓴 토큰이지 선언 ID가 아니다」**, **「가드가 일한 증거」**, **「영구 배제」**라고 적었다. **그 주장은 거짓이다.** 코퍼스 줄을 열어 보면 7건 전부가 Given-When-Then을 갖춘 실제 AC 선언 불릿이다. 실행한 명령과 verbatim 출력(트리 `52f863f36`, `/usr/bin/grep` — 셸 `grep`은 ugrep 래퍼라 조용히 건너뛴다):

```
$ /usr/bin/grep -rn -E '^\s*[-*+]\s+\*{0,2}AC-(GREEN|MUTANT|EVIDENCE|SCOPE|ORDERING|MRR-GREEN|HFC-GATE)\b' .moai/specs --include=spec.md
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:48:- **AC-GREEN**: Given the repaired tree, When `go run ./cmd/moai spec lint --strict` runs …
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:49:- **AC-MUTANT**: Given the two `status:` lines restored via `git checkout 615d18c1f -- …`, When …
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:50:- **AC-ORDERING**: Given the branch history, When `git log --oneline` is read, Then …
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:51:- **AC-SCOPE**: Given the branch diff for the two files, When … Then it shows exactly two deleted lines …
SPEC-SPECLINT-ARTIFACT-STATUS-001/spec.md:52:- **AC-EVIDENCE** (parent: REQ-007): Given `.moai/reports/t490/`, When the evidence files are read, Then …
SPEC-V3R6-MAIN-RED-REMEDIATION-001/spec.md:124:- AC-MRR-GREEN: `go test ./internal/template/...` 0 fail + cross-platform build exit 0 + lint baseline 유지.
SPEC-HOOK-FAILURE-CLASSIFY-001/spec.md:97:- **AC-HFC-GATE** (quality gate): **Given** the full package, **When** `go test ./internal/hook/… ` runs, **Then** …
```

7개 형태 전부에 대해 각각 1건 이상 실재하며, 프로브가 센 건수(각 1)와 일치한다.

**결정은 살아남고, 근거가 바뀐다.** 마지막 분절을 숫자로 고정하는 것이 옳은 이유는 「이 7건이 선언이 아니라서」가 **아니다**. 실측으로 세울 수 있는 유일한 이유는 이것이다 — **숫자 꼬리 말고는 선언과 임의 토큰을 가르는 측정된 성질이 없다.** 꼬리 제약을 놓으면 문법은 `AC-` 접두를 가진 임의 토큰을 전부 받게 되고, 그 폭을 좁힐 다른 축은 이 카드에 없다.

**처분이 바뀐다.** 7건은 나머지 32건과 **같은 부류**다 — 진짜 선언인데 이 카드의 축이 닿지 못하는 것. 따라서 미포함 39건은 「올바른 배제 7 + 유예 32」가 아니라 **문법이 받아들이지 못하는 진짜 선언 39건**이며, 그 전부가 후속 카드 후보다. 「영구 배제」는 철회한다(§4).

**재검토 트리거**: 후속 카드가 `AC-` 접두 토큰 중 선언과 비선언을 가르는 축을 실측으로 세우면, 이 결정을 다시 연다. 그 축이 서기 전에는 열지 않는다.

#### [HARD] 미측정 가설 — 「임의 토큰이 전부 들어온다」

「숫자 꼬리를 놓으면 `AC-`로 시작하는 임의 대문자 토큰이 전부 들어온다」는 **실측 근거가 0인 가설이다.** 관측된 미포함 16개 형태 중 산문 라벨은 **하나도 없었다** — 전부 실제 선언이었다. 위험은 개연적이지만 이 코퍼스에서 확인된 바 없으며, 그 사실을 여기 적는다. `AC-FOO-BAR` 부정 케이스(AC-ACA-001-003)는 이 가설을 **문법 수준에서** 검사할 뿐 코퍼스에서 그 재료를 찾은 것이 아니다.

#### 나머지 부류의 축 배정 (변경 없음)

- 알파벳 접미(`-001a`/`-001b`)와 점 하위번호(`004.1`)는 `hasIDSuffix`가 `strings.Contains(id, ".")`인 이상 `autoWrapSingle` 분기를 뒤집어 **트리 모양을 바꾼다**(`parser.go:177,185`) — REQ-ACA-001-002가 보존을 요구하는 바로 그 의미론이다.
- 범위 표기(`001..005`)는 한 줄이 여러 AC를 선언하는 형태라 문법이 아니라 **한 줄 대 여러 항목**의 의미론을 정해야 한다.
- 어느 것도 이 카드의 네 축 안에서 결정할 수 없다.
- **비율**: 39 / 1167 = 3.3%(판별식 B). 이 카드가 회수하지 않는 몫이며, `plan.md` §F의 Tier L 격상 조건도 아니다 — 후속 카드는 32가 아니라 **39**에서 크기를 잡아야 한다.

이 결정은 이제 열린 상태가 아니다. `plan.md` §B.1이 달고 있던 clarification 마커는 삭제했다. 재검토 트리거는 둘이며 그 밖에는 열지 않는다: (1) **run 단계 M2 재측정에서 미포함 건수가 39를 유의하게 넘으면** 축 구성을 다시 본다(`plan.md` §F), (2) 후속 카드가 `AC-` 접두 토큰의 선언/비선언을 가르는 축을 **실측으로** 세우면 꼬리 제약을 다시 본다(위 정정 절).

### 1.2 거절 원인 — 코퍼스에서 유도했다

[HARD] **이 절의 집계는 전부 판별식 A 계열이며 재유도 불가다** — 보존 프로브(판별식 B)는 형태별 집계를 `UNCOVERED-SHAPE`(미포함 39건)에 대해서만 내고, 아래 상위 10행과 구분자 집계는 삭제된 프로브의 출력(`.moai/reports/t528/measurement-20260908.md`)에만 있다. 이 수들은 **§B 네 축을 유도한 근거**로 쓰였고 그 역할은 유효하다(형태의 존재와 상대적 크기를 보이는 것이 목적이었다). 그러나 **어떤 것도 §1.1의 정본 기준선이 아니며**, 재측정 값과 나란히 놓지 않는다(REQ-ACA-001-016). 넓힘 이후 형태별 분포를 다시 재려면 판별식 B로 새로 재고 그 사실을 적는다.

ID 분절 형태(판별식 A · 재유도 불가, 숫자를 `N`으로 정규화, 상위 10행):

| 형태 | 건수 |
|---|---|
| `AC-N` | 150 |
| `AC-VNRN-RT-N-N` | 109 |
| `AC-WFN-N` | 85 |
| `AC-SPC-N-N` | 63 |
| `AC-ORC-N-N` | 56 |
| `AC-EXTN-N` | 53 |
| `AC-MIGN-N` | 42 |
| `AC-CON-N-N` | 40 |
| `AC-HRN-N-N` | 33 |
| `AC-UTIL-N-N` | 30 |

ID 직후 구분자: `:` 759 · `(` 228 · `—` 91 · 없음/기타 82.

코퍼스에서 그대로 가져온 실패 줄:

```
- **AC-ASE-001** — **Given** the run-phase commit on `WT-achwd-strip-exempt`,
- AC-AUDIT-SNAPSHOT-001 (A1): sticky cache — past-24h unchanged-hash skip still fires.
- AC-AUTONOMY-TIERS-001 (REQ-001): `moai init` wizard offers 3-tier selection; …
```

세 줄이 세 가지 거절 사유를 각각 대표한다: 굵은 글씨 래퍼 + 엠대시 구분자, 괄호 한정어, 그리고 (앞의 둘과 겹쳐서) 분절 수 불일치.

### 1.3 사용자에게 보이는 결함

지도를 **호출 지점**과 **결과 소비 지점**으로 나눠 적는다. 종전 판본은 소비 지점만 적어 호출 지점 하나를 빠뜨렸고, 그 누락이 무해하지 않았다(감사 D8).

**호출 지점 2곳** (`grep -rn --include='*.go' 'ParseAcceptanceCriteria' . | grep -v _test.go`):

- `internal/spec/lint.go:636` — `criteria, _ := ParseAcceptanceCriteria(body, false)`. **두 번째 반환값을 버린다.**
- `internal/cli/spec_view.go:72` — `criteria, parseErrors := spec.ParseAcceptanceCriteria(...)`. 오류 슬라이스를 실제로 읽는 유일한 프로덕션 호출자.

**결과 소비 지점 2곳**:

- `internal/cli/spec_view.go` — `moai spec view <ID>`. blind인 100개 SPEC에 대해 지금 `No acceptance criteria found in <ID>`를 출력한다. **이것이 이 카드의 일차 정당화다.**
- `internal/spec/lint.go:915` — `collectAllREQIDs(doc.Criteria)`, `CoverageRule`. 자문(advisory) 등급 `warning`을 낸다. 넓히면 커버된 REQ가 **늘기만** 하므로 finding은 줄어드는 방향이다. **이 카드에서는 읽기 전용이다**(§3).

#### 오류가 어디서 버려지고 어디서 치명이 되는가 (계측 지점의 정정)

`ValidateDepth` 초과와 `DuplicateAcceptanceID`는 `parseAcceptanceCriteriaInternal`이 오류 슬라이스에 담아 돌려준다. 그 슬라이스의 운명이 호출 지점마다 다르다.

```go
// internal/spec/lint.go:635-636 — 오류를 버린다
criteria, _ := ParseAcceptanceCriteria(body, false)
doc.Criteria = criteria
```

```go
// internal/cli/spec_view.go:80-86 — 분류되지 않은 오류는 명령을 실패시킨다
case *spec.DanglingRequirementReference: … // warning
case *spec.MissingRequirementMapping:     … // warning
default:
    return fmt.Errorf("parse error: %w", err)
```

결과는 이렇다. **`moai spec lint`를 아무리 돌려도 이 두 오류는 0으로 관측된다** — 실제로 발화하든 안 하든. 그러므로 lint 실행으로 이 위험을 재려는 계측은 판정 불가능하며, 종전 판본의 AC-ACA-001-010 셋째 절이 정확히 그 자리에 있었다(감사 D3). 반대로 **CLI에서는 두 오류가 `default`로 떨어져 명령 전체를 실패시킨다.** 넓힘이 어떤 SPEC에서 중복 ID를 새로 만들면, 그 SPEC의 `moai spec view`는 오늘의 `No acceptance criteria found`(무해)에서 **하드 에러**로 바뀐다 — 이 카드가 고치겠다고 지목한 바로 그 명령이 기준선보다 나빠지는 방향이다.

**이 카드는 그 방향을 계측만 하고 수리하지 않는다**(운영자 판정 2026-09-08). 판정하는 수는 총량이 아니라 **기준선 대비 델타**다 — 넓힘 전 하드 에러는 **442**건이고 전부 `acceptance criteria section not found`, 즉 이 카드가 건드리지 않기로 한 헤딩 축의 기존 조건이다(`.moai/reports/t528/probe/before/specview-before-20260908.md`). 기준선에 **없던** 건이 나오면 카드를 멈추고 보고하며, `internal/cli/spec_view.go`는 편집하지 않는다(REQ-ACA-001-013, §4).

#### 재료는 실측으로 0이다 — 그리고 그 0이 새 위험을 만든다

§B.1 후보 문법이 현재 코퍼스에서 만드는 중복 ID는 **0건**이다(HISTORY의 중복 노출 실측, `probe/duplicate-20260908.txt`, `EXIT=0`). 넓힘 전후 모두 `files with duplicate ids = 0, dropped lines = 0`. 즉 이 변경만으로는 CLI 치명 경로가 발화하지 않는다.

[HARD] **그러나 0은 「하드 에러 수 == 0」 AC를 공허하게 만든다.** 재료가 없으면 그 AC는 **구현이 존재하지 않아도 통과하고, 측정이 아예 돌지 않아도 똑같이 통과한다** — 빈 집합 위의 초록이며, 이 카드가 D1에서 이미 한 번 당한 모양이다.

그래서 AC-ACA-001-015는 **양성 대조군을 요구한다**: 중복 AC ID를 **고의로 만든** 픽스처에서 계측이 하드 에러를 **실제로 잡아내는 것**을 먼저 보인다. 탐지기가 발화함을 보인 **뒤에야** 코퍼스 전역의 0을 「회귀 없음」으로 읽는다. 그 순서를 지키지 않으면 코퍼스 0은 아무것도 주장하지 않는다.

#### 도달성 잔여

- `buildTree`(`parser.go:137-143`)는 **중복 ID를 만나면 `continue`로 그 줄을 버린다.** 넓힘으로 산문 불릿이 먼저 들어오고 진짜 선언이 뒤에 오면, 진짜 선언이 조용히 사라진다 — 수용이 늘었는데 특정 선언은 사라지는 조합이 성립한다. 이 방향은 AC-ACA-001-002(무회귀 대조군, 파일별 AC 집합 비교)와 AC-ACA-001-014(과수용 표본)가 함께 덮는다.
- `CheckDanglingReferences`는 **프로덕션 호출자가 0건**이다(정의 `parser.go:297`, 호출은 `parser_test.go:313`뿐). 넓힘이 그 함수에 닿지 않는다.
- `ACIDInvalid`는 프로덕션 규칙으로 **존재하지 않는다** — 유일한 등장이 `lint_coverage_sibling.go:18`의 주석이다. 소비자로 세지 않는다.

---

## 2. 요구사항 (GEARS)

> **자기참조 주석**: 아래 REQ 줄과 형제 `acceptance.md`의 AC ID는 **현행 파서가 이미 받아들이는 4분절 형태**로 적었다. 자기가 고치려는 파서에 스스로 보이지 않는 문서가 되지 않기 위해서다.

- REQ-ACA-001-001: `parseSingleACLine` 함수는 코퍼스에서 실제로 쓰이는 AC ID 분절 형태 — 가변 분절 수, 알파벳 중간 분절 — 를 AC 선언 ID로 인식해야 한다(shall).
- REQ-ACA-001-002: `parseSingleACLine` 함수는 기존 하위 ID 접미(`.a`, `.a.i`)의 인식 의미론을 보존해야 한다(shall).
- REQ-ACA-001-003: **Where** AC ID가 굵은 글씨 래퍼(`**AC-…**`)로 감싸여 있을 때, `parseSingleACLine` 함수는 그 래퍼를 벗기고 ID를 인식해야 한다(shall).
- REQ-ACA-001-004: **Where** AC ID와 구분자 사이에 괄호 한정어(`(A1)`, `(REQ-001)`)가 있을 때, `parseSingleACLine` 함수는 그 한정어를 건너뛰고 선언을 인식해야 한다(shall).
- REQ-ACA-001-005: `parseSingleACLine` 함수는 ID 뒤 구분자로 콜론(`:`)과 함께 엠대시(`—`), 엔대시(`–`)를 인식해야 한다(shall).
- REQ-ACA-001-006: `findACSectionStart`의 섹션 스코핑은 변경되어서는 안 된다(shall not) — AC 절 밖의 선언 줄은 넓힘 이후에도 수집되지 않는다.
- REQ-ACA-001-007: 넓힌 앵커는 현재 파싱에 성공하는 18개 `spec.md`의 결과를 회귀시켜서는 안 된다(shall not).
- REQ-ACA-001-008: 넓힘의 회수량은 코퍼스에서 재측정해야 한다(shall) — 수용/도달 비가 기준선 216/1167(18.5%, 판별식 B)에서 얼마로 움직이는지 명령 출력 그대로 기록한다.
- REQ-ACA-001-009: **When** 넓힌 앵커를 적용한 lint를 코퍼스 전체에 실행할 때, `CoverageRule` finding 수의 전후 변화를 계측해야 한다(shall) — 증가분이 있으면 억제하지 않고 설명한다.
- REQ-ACA-001-010: 넓힌 앵커의 경계는 뮤턴트 탐침으로 그려야 한다(shall) — 각 뮤턴트는 적발 기대 여부를 **사전에** 선언하고, 적발되지 않은 뮤턴트는 가드의 경계 기록으로 보존한다.
- REQ-ACA-001-011: 이 SPEC의 수리는 `internal/spec/lint.go`를 편집해서는 안 된다(shall not).
- REQ-ACA-001-012: 넓힘이 **잘못 수용하는 줄**의 규모를 실측해야 한다(shall) — 새로 수용된 줄의 표본을 열람하고, 선언이 아닌 것(산문 인용·폐기 표시·교차참조)의 건수와 줄을 기록한다. 회수량 관측이 이 의무를 대신하지 않는다.
- REQ-ACA-001-013: **Where** 넓힘 이후 어떤 `spec.md`의 `moai spec view <ID>`가 기준선에 없던 하드 에러(`parse error:`)를 내면, 그 건수를 실측해 보고해야 하며(shall) 이 카드 안에서 수리해서는 안 된다(shall not) — `internal/cli/spec_view.go`는 편집하지 않는다.
- REQ-ACA-001-014: 회수량 측정에 쓰는 프로브는 트리에 **커밋된 테스트로 존재해야 한다**(shall) — 판정 명령은 최소 1개 테스트를 실제로 선택해 실행했음을 함께 보여야 하며, 아무것도 고르지 않은 실행(`[no tests to run]`)은 통과가 아니라 실패다.
- REQ-ACA-001-015: 무회귀 대조군은 파서를 편집하기 **전에** 산출물로 고정되어야 한다(shall) — 넓힘 이후에 만든 「이전 상태」는 대조군이 아니라 사후 기록이다.
- REQ-ACA-001-016: 모든 재측정은 **트리 SHA · 판별식 · 분모 파일 목록 산출물** 셋을 함께 인용해야 한다(shall) — 셋 중 하나라도 빠진 수는 기준선과 비교할 수 없다.

---

## 3. 제약

- 대상 프로덕션 파일은 `internal/spec/parser.go` 하나다. 그 밖의 프로덕션 편집은 이 SPEC의 범위가 아니다(테스트 파일과 픽스처는 예외).
- [HARD] `internal/spec/lint.go`는 **바이트 동일**해야 한다. `collectAllREQIDs`(lint.go:915)가 수집기를 소비하지만, 소비자 수리는 이 카드의 일이 아니다 — 리드가 정한 경계다.
- [HARD] `internal/cli/spec_view.go`도 **바이트 동일**해야 한다. §1.3의 치명 경로는 이 카드가 계측만 하고 수리하지 않는다(운영자 판정: 측정만, 수리는 별도 후속 카드).
- `findACSectionStart`의 섹션 스코핑은 손대지 않는다.
- 측정 프로브는 커밋된 테스트로 트리에 있어야 한다(REQ-ACA-001-014). 「임시 프로브를 만들고 지운다」는 기준선에서 이미 한 번 재유도 불가를 만들었다.
- 형제 `acceptance.md` 처리는 이미 `lint_coverage_sibling.go`가 `ExtractRequirementMappings`로 해결했다. 다시 다루지 않는다.
- 검증 범위는 건드린 패키지로 한정한다(`go test ./internal/spec/...`). 전 패키지 판정은 CI 몫이며, 로컬 전체 스위트는 돌리지 않는다.

---

## 4. Exclusions — 이 SPEC이 만들지 않는 것

이 절은 out of scope 항목을 명시한다.

### Out of Scope — 섹션 스코핑(헤딩 축)

- `findACSectionStart`가 `##…acceptance` 헤딩을 요구하는 스코핑은 **변경하지 않는다.**

#### 종전 주장의 철회 — 「가드를 약화시키지 않는다」는 틀렸다

이 절의 이전 판본은 「항목 문법을 넓히는 일은 이 가드를 약화시키지 않는다」고 단언하면서 그 근거로 `internal/spec/lint_coverage_sibling.go:29-32`의 머리주석을 인용했다. **그 인용문은 정확히 반대를 말한다**(감사 D4). 출처 verbatim:

```
// ParseAcceptanceCriteria, which is scoped twice over: findACSectionStart needs
// an `##` heading containing "acceptance", and parseSingleACLine needs the
// `AC-…:` colon form. BOTH scopings exist because spec.md is a mixed document
// in which prose must not be read as AC.
```

산문 가드의 주체로 명시된 둘 중 **하나가 바로 이 카드가 넓히려는 콜론 형식**이다. 인용문이 부정하는 결론을 인용문 옆에 둔 것이 잘못이므로, 주장을 철회하고 아래로 바꾼다.

**정정된 서술**: 항목 문법을 넓히는 일은 **콜론 형식이 담당하던 산문 가드의 일부를 의도적으로 포기한다.** 남는 것은 섹션 스코핑 한 겹뿐이며, 그것은 위험을 AC 절 **안**으로 한정할 뿐 0으로 만들지 않는다. 이 카드는 그 대가를 **측정으로 상환한다**(REQ-ACA-001-012) — 「약화되지 않는다」는 논증으로 상환하지 않는다.

구체적 반례(감사가 작성, verbatim). AC 절 **안**의 산문 불릿:

```
- AC-OGR-003 (RETIRED) — 이 항목은 이관됐다.
```

현행 앵커는 거절한다(콜론 없음). §B.3+§B.4를 반영한 넓힌 앵커는 id `AC-OGR-003` → 괄호 한정어 `(RETIRED)` 건너뜀 → 구분자 `—` → content=산문으로 **수용한다.** 섹션 스코핑은 이 줄을 막지 못한다 — 절 **안**에 있기 때문이다.

감사의 무효 결과도 함께 남긴다. 코퍼스의 엠대시 AC 불릿 19파일을 전수 열람했으나 **실제 오탐 줄은 찾지 못했다.** 이것을 안전의 근거로 읽지 않는다 — **반례의 부재는 가드의 존재가 아니고**, 넓힘 이후 새로 들어올 951줄(판별식 B 기준 거절분)은 아직 아무도 열어보지 않았다. 그 열람이 REQ-ACA-001-012이다.

#### 관측 — 헤딩 앵커는 좁기만 한 것이 아니라 **느슨하기도** 하다 (흡수하지 않음)

`findACSectionStart`는 `##` 이상 깊이의 헤딩 중 소문자화 시 `acceptance`를 **포함하기만 하면** 받고, **첫 일치**를 돌려준다. 실측(`probe/drift-20260908.txt`, `EXIT=0`):

- AC 절을 **파일 이름 `acceptance.md`를 언급하는 헤딩**에 앵커하는 `spec.md` = **41**
- AC 절을 `###` 이상 깊이 헤딩에 앵커하는 `spec.md` = **9**

대부분은 무해하다 — `## §H. Acceptance Criteria (summary — full GWT in acceptance.md)`는 정말로 AC 절이다. 그러나 **이 카드 자신의 `spec.md`가 진짜 오탐이다**: AC 절이 `### Out of Scope — 형제 acceptance.md 인라인 파싱`에 앵커된다.

- 이것은 **관측으로만 기록한다.** 헤딩 축은 범위 밖이며(운영자 판정), 이 카드는 흡수하지 않는다.
- **다만 이것은 간극을 만든다**: 일부 in-section 선언 수가 **잘못 앵커된 절**에서 뽑혔을 수 있다. 방향과 크기를 재지 않았다(§5 Gaps). 후속 카드 후보.

### Out of Scope — AC 절 밖의 선언 줄 180건 (판별식 A 관측, 재유도 불가 · 흡수하지 않음)

- AC 절 밖에 있는 AC 선언 줄 **180건**이 관측됐다. **이 수는 판별식 A 계열이며 재유도 불가다** — 보존 프로브(판별식 B)는 절 밖 선언을 세지 않는다. 출처는 `.moai/reports/t528/measurement-20260908.md`의 `outside-section=180` 한 줄이고, 정본 기준선(§1.1)의 어떤 수와도 비교 대상이 아니다. 범위 밖 관측이므로 이 카드는 재측정하지 않는다. 리드는 이를 **관측으로 기록하되 이 SPEC에 흡수하지 말 것**을 명시적으로 지시했다.
- 이 180건이 진짜 AC인지 산문인지 표본 확인하지 않았다(§5 Gaps). 그 판별부터가 별도 작업이다.
- 후속 카드 후보로만 남긴다.

### Out of Scope — `internal/spec/lint.go` 편집

- `collectAllREQIDs`(lint.go:915)는 이 수집기의 소비자다. 넓힘이 그 규칙의 출력을 움직일 수 있지만, **소비자 수리는 이 카드가 아니다.**
- 이 카드는 그 파일을 읽기만 하고, 바이트 동일성을 인수 기준으로 검증한다(AC-ACA-001-007).

### Out of Scope — `internal/cli/spec_view.go` 편집 (계측만, 수리하지 않음)

- §1.3이 보인 CLI 치명 경로(`spec_view.go:80-86`의 `default → parse error:`)는 이 카드가 **재기만 한다.** 운영자 판정 2026-09-08: **측정만, 수리는 별도 후속 카드.**
- 판정은 REQ-ACA-001-013 / AC-ACA-001-015: 넓힘 이후 하드 에러를 내는 SPEC 수가 **0**이면 통과, 0이 아니면 카드를 **멈추고 보고한다.** 여기서 고치지 않는다.
- 이 파일은 읽기만 한다. `git diff --stat -- internal/cli/spec_view.go`가 빈 출력임을 인수 기준으로 확인한다.

### Out of Scope — 헤딩 앵커의 느슨함 (오탐 41 + 9)

- §4의 관측 절이 기록한 41 / 9건은 헤딩 축의 결함이며, 이 카드는 흡수하지 않는다.
- 이 카드가 만드는 것도 아니다 — 기준선에도 넓힘 이후에도 같은 상태로 남는다.

### Out of Scope — 미포함 39건의 회수

- §1.4가 분해한 39건(단어 꼬리 7 · 알파벳 접미 3 · 점 하위번호 10 · 범위 표기 15 · 문자 접두 분절 4)은 이 카드가 회수하지 않는다.
- **[HARD] 「단어 꼬리 7건은 영구 배제다」는 철회한다.** 종전 판본은 7건을 「규칙이 일한 결과」로 적었으나, 코퍼스 전수 열람 결과 7건 전부가 진짜 AC 선언이었다(§1.4의 verbatim 인용). 이것은 결정의 **비용**이지 규칙이 작동한 증거가 아니다.
- **39건 전부가 「문법이 받아들이지 못하는 진짜 선언」이며, 전부 후속 카드 후보다.** 「올바른 배제 7 + 유예 32」라는 분해는 더 이상 유효하지 않다. 후속 카드는 32가 아니라 **39**에서 크기를 잡는다.
- 재개 트리거는 §1.4가 정본이다 — M2 재측정에서 미포함이 39를 유의하게 넘거나, `AC-` 접두 토큰의 선언/비선언을 가르는 축이 실측으로 서면 다시 연다.

### Out of Scope — 형제 `acceptance.md` 인라인 파싱

- `acceptance.md` 쪽은 `lint_coverage_sibling.go`가 `ExtractRequirementMappings`로 이미 우회 처리했다. 그 파일의 머리주석이 왜 인라인 스코핑 두 개를 형제 파일에 적용하지 않는지 설명한다.

### Out of Scope — 코퍼스의 AC 표기 전면 정리

- 넓힌 문법으로도 여전히 남는 거절 줄이 있을 수 있다(구분자 「없음/기타」 — 판별식 A 82건 / 판별식 B 42건 중 일부. **두 수는 비교 대상이 아니고, A쪽은 재유도 불가다**). 코퍼스를 규약에 맞춰 다시 쓰는 일은 이 SPEC에 담지 않는다.

---

## 5. Gaps — 관측하지 않은 것

- **넓힘 이후의 회수량.** 넓힌 문법이 951건(판별식 B 기준 거절분) 중 실제로 몇 건을 회수하는지는 아직 재지 않았다. run 단계 측정 대상이다(REQ-ACA-001-008).
- **새로 수용될 951줄의 성질.** 그 중 몇 줄이 선언이 아닌지 아무도 열어보지 않았다. 감사의 엠대시 19파일 전수 열람은 **현재 수용되는 범위**에 대한 것이며, 넓힘 이후 들어올 줄에 대해서는 아무 말도 하지 않는다. run 단계가 표본으로 잰다(REQ-ACA-001-012).
- **`CoverageRule` finding 전후 delta.** 방향은 「줄어드는 쪽」으로 예측되지만 크기를 재지 않았다. run 단계가 잰다(REQ-ACA-001-009).
- **CLI 하드 에러의 현재 건수 — M0이 닫았다.** 이 항목은 플랜 단계에서 「재지 않았다」로 열려 있었다. M0이 코퍼스 전체에 CLI를 돌려 **442**를 세웠고 전부 `acceptance criteria section not found` 부류다(`.moai/reports/t528/probe/before/specview-before-20260908.md`, 트리 `52f863f36`, 분모 807). 남은 것은 넓힘 **이후**의 같은 측정이며 M7이 잰다(REQ-ACA-001-013).
- **AC 절 밖 180건의 성질.** 세기만 했고 진짜 AC인지 산문인지 표본 확인하지 않았다.
- **잘못 앵커된 AC 절이 in-section 수에 준 영향.** §4가 41 + 9건을 세었으나, 그 중 몇 건이 실제로 엉뚱한 절을 AC 절로 읽어 선언 수를 부풀렸는지(또는 깎았는지) 재지 않았다. 정본 기준선 1167(판별식 B)은 그 영향을 포함한 값이다.
- **`ValidateDepth`의 신규 발화 규모.** `DuplicateAcceptanceID` 쪽은 실측했다(재료 0). 깊이 초과 쪽은 재지 않았다.
- **`moai spec view`의 넓힘 이후 코퍼스 전역 실행.** 넓힘 **전**의 실행은 M0이 마쳤다(위 항목, 442). 넓힌 파서로 빌드한 바이너리를 코퍼스 전체에 돌려 델타를 내는 것은 아직 남았고, M7이 잰다.
- **`buildTree` 계층화의 실제 영향.** 들여쓰기 기반 부모-자식 구성이 새로 수용된 줄에서 어떤 트리를 만드는지 재지 않았다. 중복 ID `continue` 폐기(`parser.go:137-143`)가 실제로 몇 줄을 삼키는지도 재지 않았다.
- **구분자 「없음/기타」의 내역.** 집계만 있고 분해하지 않았다. 판별식 A에서 82, 판별식 B에서 42로 관측됐다 — **두 수는 비교 대상이 아니다**(판별식이 다르다).
- **기준선 실행의 종료 코드.** `measurement-20260908.md`의 실행에는 종료 코드가 기록되지 않았다. 재실행분(`probe/run-20260908.txt`, `probe/drift-20260908.txt`)에는 `EXIT=0`이 있다.

---

## 6. 잔여 위험

- **넓힌 문법이 AC가 아닌 불릿을 AC로 읽을 수 있다** (가장 큼). 섹션 스코핑이 위험을 AC 절 안으로 한정하지만 0으로 만들지 않으며, §4의 반례가 보이듯 절 안의 산문 불릿은 스코핑이 막지 못한다. 증거는 둘로 나뉜다 — **뮤턴트**는 구현이 자기 테스트에 대해 갖는 경계를 그리고(REQ-ACA-001-010), **표본 열람**은 코퍼스에 대한 과수용을 잰다(REQ-ACA-001-012). 뮤턴트만으로는 이 축을 그릴 수 없다: 뮤턴트는 기능을 **제거**하는 방향이고, 과수용은 기능이 **작동할 때** 일어난다.
- **`doc.Criteria`가 커지면 `ValidateDepth` / `DuplicateAcceptanceID`가 새로 발화할 수 있다.** 발화 자체는 코퍼스가 원래 갖고 있던 사실이 드러나는 것이지만, **CLI에서는 하드 에러가 된다**(§1.3). 이 카드가 고치겠다고 지목한 명령이 기준선보다 나빠지는 유일한 경로다. **현재 코퍼스에서 그 재료는 실측 0이다**(`probe/duplicate-20260908.txt`) — 위험이 사라진 것이 아니라 **오늘 재료가 없는 것**이며, run 단계 구현이 §B.1 후보와 다르게 넓히면 달라진다. 계측하되 수리하지 않는다(REQ-ACA-001-013).
- **판별식이 코퍼스에 맞지 않게 형성돼 거짓 수치·거짓 분류를 낳는다 — 이 카드에서 이미 다섯 번 일어났다**(HISTORY의 정정 표 5행. 4·5행은 감사 iteration 2가 잡았고, 4행은 그 규율을 [HARD]로 승격한 개정 자신이 어긴 것이다). 이것은 이 카드가 고치려는 결함과 **같은 부류**다. 완화는 둘이며 어느 쪽도 수를 다시 읽는 것이 아니다: 표본 줄 열람(`plan.md` §D.1)과 양성 대조군(AC-ACA-001-015).
- **같은 패키지를 두 번째 카드가 동시에 편집한다.** 패키지 테스트 실패의 귀속이 자동으로 이 카드로 오지 않는다 — 되돌림 재현과 `git diff` 반경 확인이 선행한다(`plan.md` §D.3).
- **`buildTree`의 중복 ID 폐기가 진짜 선언을 삼킬 수 있다.** `parser.go:137-143`이 중복을 만나면 `continue`로 그 줄을 버리므로, 산문 불릿이 먼저 들어오고 진짜 선언이 뒤에 오면 **수용은 늘었는데 그 선언은 사라진다.** 회수량만 재는 관측은 이 조합을 초록으로 읽는다.
- **`buildTree`의 들여쓰기 기반 계층화가 새로 수용된 줄들에서 의도치 않은 부모-자식 관계를 만들 수 있다.**

---

## 7. 교차 참조

- 카드: `t528` · 워크트리 `.claude/worktrees/t528` · 브랜치 `WT-ac-collector-anchor`
- 실측 정본: `.moai/reports/t528/measurement-20260908.md`
- 감사 판정: `.moai/reports/t528/plan-audit.md` (iteration 1, FAIL — 이 개정이 응답한다)
- 재유도 가능한 프로브와 실행 기록:
  - `.moai/reports/t528/probe/ac_anchor_probe_test.go` — 판별식 그 자체(소스)
  - `.moai/reports/t528/probe/run-20260908.txt` — 실행 출력 + `EXIT=0`
  - `.moai/reports/t528/probe/drift_probe_test.go` · `drift-20260908.txt` — 판별식 드리프트 분리 + `EXIT=0`
  - `.moai/reports/t528/probe/filelist.txt` — 분모 807줄
  - `.moai/reports/t528/probe/positive-needle.txt` — 18개 대조군 구성원 전체
  - `.moai/reports/t528/probe/duplicate_probe_test.go` · `duplicate-20260908.txt` — 중복 ID 노출 실측(넓힘 전후 0/0) + `EXIT=0`. 소스 머리주석이 「4 파일 / 24 줄」 거짓 수치의 정정 기록을 담는다
  - `.moai/reports/t528/probe/pkg-baseline-20260908.txt` — 패키지 테스트 착수 기준선 + `EXIT=0`
- `internal/spec/lint.go:636` — 오류 슬라이스를 버리는 호출 지점 (읽기 전용)
- `internal/cli/spec_view.go:80-86` — 미분류 오류가 치명이 되는 경로 (읽기 전용, 편집 금지)
- `internal/spec/parser.go:137-143` — `buildTree`의 중복 ID `continue` 폐기
- `internal/spec/parser.go:218` — 현행 AC ID 앵커
- `internal/spec/parser.go` — `findACSectionStart`(섹션 스코핑, 범위 밖), `buildTree`, `autoWrapSingle`
- `internal/spec/lint.go:915` — `collectAllREQIDs` 소비자 (읽기 전용)
- `internal/spec/lint_coverage_sibling.go` — 두 스코핑의 정당화가 기록된 머리주석
- `internal/cli/spec_view.go:72` — `moai spec view` (사용자에게 보이는 결함 지점)
- SPEC-COVERAGE-RULE-SCOPE-001 — 같은 파싱 경로의 REQ 쪽 협소성을 다룬 선행 SPEC
