---
id: SPEC-AC-COLLECTOR-ANCHOR-001
title: "인라인 AC 수집기 항목 문법 앵커 — 진행 기록"
status: in-progress
created: 2026-09-08
updated: 2026-09-08
author: manager-develop
card: t528
---

# SPEC-AC-COLLECTOR-ANCHOR-001 — 진행 기록

카드 `t528`. 워크트리 `.claude/worktrees/t528`, 브랜치 `WT-ac-collector-anchor`.
착수 트리 `52f863f36`.

모든 수치는 REQ-ACA-001-016의 세 요소 — **트리 SHA · 판별식 · 분모 파일 목록** — 를 함께 인용한다.

---

## §E.1 마일스톤 진행

| M | 내용 | 상태 |
|---|---|---|
| M0 | 대조군·측정 도구 고정 (파서 편집 이전) | 완료 |
| M1 | 네 축 앵커 구현 | 완료 |
| M2 | 코퍼스 재측정 | 완료 |
| M3 | before-image 대조 | 완료 |
| M4 | 뮤턴트 경계 | 완료 |
| M5 | 섹션 스코핑 불변 픽스처 | 완료 |
| M6 | 과수용 실측 | 완료 |
| M7 | CLI 렌더 + 하드 에러 게이트 | 완료 |
| M8 | CoverageRule delta + 바이트 동일성 | 완료 |

---

## §E.2 Run-phase Evidence

### M0 — 대조군과 측정 도구를 파서 편집 **이전에** 고정했다

**(a) 프로브 승격 (REQ-ACA-001-014).**
`.moai/reports/t528/probe/ac_anchor_probe_test.go` → `internal/spec/zz_t528_anchor_probe_test.go`
(커밋된 테스트). 판별식 `declRe`(판별식 B)는 산출물 사본과 **byte-identical**로 옮겼다.

승격판이 원본과 다른 점은 둘이고, 둘 다 의도적이다.

1. **출력 디렉터리가 `T528_PROBE_OUT`로 분기한다** (기본 `probe/out`). 원본은
   `probe/positive-needle.txt` · `probe/filelist.txt`를 **덮어쓴다** — 넓힘 이후에 한 번만
   돌아도 핀된 before-image가 사라지고, 그러면 무회귀 판정이 넓힌 파서와 자기 자신의 비교가
   된다(AC-ACA-001-002가 금지한 바로 그 모양).
2. **수용 열이 둘이다.** `baselineRe`는 넓힘 **이전** 앵커의 동결 사본이고, live 열은 실제
   `parseSingleACLine`을 호출한다. 원본은 정규식 사본 하나만 태워서, 파서를 넓혀도 같은 216을
   다시 내는 **공허한 재측정**이 된다. 두 열을 한 프로세스에서 함께 내면 기준선과 재측정이
   같은 실행·같은 트리·같은 분모 위에 서고, 비교가 두 실행을 가로지르지 않는다.

**(b) 승격판이 플랜 단계 기준선을 그대로 재유도한다.**

명령 · verbatim 출력: `.moai/reports/t528/probe/m0-before-run.txt` (`EXIT=0`)

```
$ T528_PROBE_OUT=../../.moai/reports/t528/probe/before \
  go test ./internal/spec/ -run TestT528Anchor -v -count=1 -timeout 600s

    DENOMINATOR spec.md read = 807 (list: .../before/filelist.txt)
    IN-SECTION declarations = 1167
      accepted by FROZEN baseline anchor = 216
      accepted by LIVE parseSingleACLine = 216
      rejected by LIVE parseSingleACLine = 951
      NEWLY accepted (live yes, baseline no) = 0
    POSITIVE-NEEDLE files (live) = 18
    BASELINE-NEEDLE files (frozen anchor) = 18
    FULLY-BLIND files (>=1 decl, 0 accepted by live) = 101
    B.1 numeric-tail candidate: covered = 1128  UNCOVERED = 39
      SEP :  800 · SEP (  232 · SEP —  93 · SEP «none/other»  42
--- PASS: TestT528Anchor (0.18s)
EXIT=0
```

- **트리 SHA**: `52f863f36` (`internal/spec/parser.go` 미편집)
- **판별식**: B — `zz_t528_anchor_probe_test.go`의 `declRe`
- **분모 목록**: `probe/before/filelist.txt` (807)

플랜 단계 산출물과 대조: `positive-needle.txt` · `filelist.txt` · `blind-files.txt`
세 파일 모두 `diff -q` **무출력**(IDENTICAL). 승격이 판별식을 옮기지 않았다는 증거다.

**live == frozen == 216, newly-accepted == 0**이 승격판 live 열의 등가성 증명이다. 이 등가가
없으면 M2의 재측정 값은 216과 비교할 수 없다.

**비공허성 (AC-ACA-001-016)**: `[no tests to run]` / `[no test files]` 토큰 **부재**,
`--- PASS: TestT528Anchor` 존재, `EXIT=0`. 셋 다 위 출력에 있다.

**(c) before-image 산출물** — 전부 `.moai/reports/t528/probe/before/`, 파서 편집 이전 커밋.

| 산출물 | 내용 |
|---|---|
| `filelist.txt` | 분모 807 |
| `positive-needle.txt` | 양성 대조군 18 (live) |
| `baseline-needle.txt` | 동결 앵커 기준 18 |
| `blind-files.txt` | 전맹 101 |
| `rootac.txt` | **파일별 루트 AC 트리 모양** 807행 — M3 대조 기준. `ID(child,child)` 형식이라 루트 집합과 트리 모양을 함께 나른다 |
| `reqmap.txt` | 파일별 AC → REQ 매핑 216행 — AC-ACA-001-008의 매핑 불변 대조 기준 |
| `still-rejected.txt` | 거절 951행 (경로:줄번호:원문) |
| `newly-accepted.txt` | 0행 (정의상 비어 있음) |
| `specview-errors.txt` | CLI 스윕 원출력 448행 |
| `specview-before-20260908.md` | CLI 하드 에러 before-image + 정정 2건 |
| `specview-sweep.sh` | 스윕 스크립트 (재유도용) |

### [HARD] M0에서 나온 정정 2건 — 문서의 전제가 실측과 어긋났다

정본은 `.moai/reports/t528/probe/before/specview-before-20260908.md`. 요지만 옮긴다.

**정정 1 — `--acceptance` 플래그는 존재하지 않는다.**
`spec.md` §1.3 · `plan.md` §E M7 · `acceptance.md` §D.12/§D.15가 전부
`moai spec view <ID> --acceptance`를 명령으로 적는다. `internal/cli/spec_view.go:37`이
등록하는 플래그는 `--shape-trace` 하나뿐이고, acceptance 뷰가 곧 `moai spec view <ID>`다.
적힌 명령으로 코퍼스를 훑으면 `SWEPT=807 NONZERO_EXIT=807` — 전부 `Unknown flag`이며
수집기에 대해 아무 말도 하지 않는다. **이 카드가 다섯 번 당한 「오형성 판별식이 낸 자신
있는 수」와 같은 모양이라 기록으로 남긴다.** 이후 모든 측정은 실제 명령을 쓴다.

**정정 2 — CLI 하드 에러 before-image는 0이 아니라 442다.**

```
SWEPT=807  NONZERO_EXIT=448
  parse error:            442   전부 "acceptance criteria section not found"
  spec.md not found for:    6   _archive/<ID>/ 경로 — 파서에 도달조차 하지 않음
  addressable SPECs:      801
```

`plan.md` §E M7과 `acceptance.md` §D.15는 before-image를 **0**으로 핀하며
`probe/duplicate-20260908.txt`를 인용한다. 그 산출물이 재는 것은 **중복 ID 재료**(진짜로 0)
이지 CLI 하드 에러가 아니다 — 서로 다른 두 양이고 0이 한쪽에서 다른 쪽으로 옮겨졌다.
다만 `acceptance.md` §D.15의 **Gap** 절이 「실제 `moai spec view`를 코퍼스 전체에 돌린 적은
아직 없다」고 스스로 적어 두었으므로, 이 측정은 그 간극을 **닫는** 것이지 문서를 뒤집는
것이 아니다.

442 전부가 `acceptance criteria section not found` — `findACSectionStart`가 -1을 돌려줄 때
`internal/cli/spec_view.go:85`의 `default:`가 치명으로 만드는 **헤딩 축**이며, 이 카드의
범위 밖(`spec.md` §4)이고 항목 문법 넓힘이 닿지 않는다.
**`DuplicateAcceptanceID` / 깊이 초과 부류는 0건**이다 — 넓힘이 새로 만들 수 있는 유일한
부류가 이것이고, 그래서 이 0이 실제 게이트의 기준선이다.

따라서 판정 게이트는 REQ-ACA-001-013이 실제로 적은 것 — **기준선에 없던 하드 에러 == 0** —
이며 442에 대한 delta로 잰다. 절대값 0은 이 코퍼스에서 달성 가능한 수가 아니었다.

### M1 — 네 축 앵커 구현 (RED 선행)

**RED**: `.moai/reports/t528/probe/m1-red-run.txt` (`EXIT=1`), 트리 `d3f705d99`,
`internal/spec/parser.go` **미편집**. 실패한 것은 분절 문법 6/7 · 굵은 글씨 3/3 ·
괄호 한정어 3/3 · 구분자 3/3 · 섹션 스코핑 1 — 전부 `got ids: []`.

RED에서 **통과한** 두 케이스는 진척이 아니라 가드다. 숫자 꼬리 부정 케이스는 아무것도
수용되지 않으니 잘못된 수용도 없어서 통과하고(비공허성은 M4 뮤턴트 1이 세운다), 하위 ID
접미 보존은 트리 모양이 **움직이지 않는다**는 것이 주장 자체다.

**GREEN**: `internal/spec/parser.go` 앵커 한 줄.

```go
var acIDPattern = regexp.MustCompile(`^(AC-(?:[A-Za-z0-9]+-)*[0-9]+(?:\.[a-z](?:\.[a-z]+)?)?)\*{0,2}\s*(?:\([^()]*\)\s*)?\*{0,2}\s*[:—–]\s*`)
```

축별 귀속 (설명되지 않은 diff 표면 0):

| 축 | 조각 | 근거 |
|---|---|---|
| ① 분절 | `AC-(?:[A-Za-z0-9]+-)*[0-9]+` | 가변 분절·알파벳 중간 분절, 마지막은 숫자 고정 |
| ② 접미 | `(?:\.[a-z](?:\.[a-z]+)?)?` | **넓힘 이전과 동일**. `hasIDSuffix`/`autoWrapSingle`이 점으로 분기 |
| ③ 굵은 글씨 | `\*{0,2}` (양쪽) | `TrimLeft`가 남긴 닫는 `**` |
| ④a 한정어 | `(?:\([^()]*\)\s*)?` | `(A1)` · `(REQ-001)` |
| ④b 구분자 | `[:—–]` | 콜론·엠대시(U+2014)·엔대시(U+2013) |

**축 밖 변경 1건 — 접어 넣지 않고 적는다.** 앵커를 패키지 수준 `var`로 끌어올렸다.
문법 네 축에 속하지 않으며, 근거는 성능이다: 이 정규식은 807개 `spec.md`의 줄마다 평가되므로
호출마다 컴파일하는 것이 원본의 낭비였다.

`findACSectionStart`(헤딩 축, 범위 밖)와 전처리 `strings.TrimLeft`는 **건드리지 않았다**.

**테스트 기대가 틀렸고 코드가 옳았던 건 1건.** `TestT528ParenQualifierLeavesREQMappingAlone`이
`REQ-WF001-001`을 기대했으나 `ExtractRequirementMappings`는 접두를 벗긴 `WF001-001`을 낸다
(`ears.go:130`이 접두 **뒤** 그룹을 캡처). 넓힘 이전 before-image(`before/reqmap.txt`)에
`AST-001-001` 같은 같은 형태가 이미 있으므로 기존 동작이며, 기대를 고치고 코드는 두었다.

### M2 — 코퍼스 재측정

**[HARD] 재측정 3요소** (REQ-ACA-001-016): 트리 `13fda0f6e` · **판별식 B**
(`zz_t528_anchor_probe_test.go`의 `declRe`) · 분모 `probe/after/filelist.txt` (807).

명령 · verbatim: `.moai/reports/t528/probe/m2-after-run.txt` (`EXIT=0`)

```
$ T528_PROBE_OUT=../../.moai/reports/t528/probe/after \
  go test ./internal/spec/ -run TestT528Anchor -v -count=1 -timeout 600s
```

| 관측 | 기준선 | 재측정 | 판별식 |
|---|---|---|---|
| 분모 `spec.md` | 807 | 807 | — |
| AC 절 안 선언 줄 | 1167 | 1167 | B |
| **수용** | **216** | **1085** | B |
| **거절** | **951** | **82** | B |
| **도달률** | **18.5%** | **93.0%** | B |
| 파싱 결과 ≥1 파일 | 18 | 110 | B |
| 전맹 파일 | 101 | 9 | B |
| 동결 앵커 수용(대조 열) | 216 | **216** | 넓힘 이전 앵커 |
| 숫자 꼬리 미포함 | 39 | 39 | B |

동결 열이 같은 실행·같은 트리·같은 분모에서 216을 그대로 낸다. 기준선과 재측정이 두 실행을
가로지르지 않는다는 뜻이며, 이 카드가 드리프트 정정에서 겪은 판별식 혼선의 재발을 구조적으로
막는다.

**[HARD] 판별식 A 계열 수치(`1160` / `944` / `100` / `82`(구분자 「없음/기타」))는 위 표
어디에도 없다.** 재유도 불가이며 비교 대상이 아니다(`spec.md` §1.1).

**잔여 거절 82 분해** — 정본 `probe/residual-82-20260908.txt`, 부류마다 verbatim 표본 동봉.

| 부류 | 건수 | 처분 |
|---|---|---|
| 숫자 꼬리 미포함 (§1.4의 39) | 39 | 예정된 잔여 — 변동 없음 |
| 대괄호 한정어 `**AC-1 [REQ-002]**:` | 17 | 축 밖 — 후속 카드 |
| `↔` 구분자 | 8 | 축 밖 |
| 구분자 없음 (`**AC-1** parse/…`, `AC-001 반출 …`) | 14 | 축 밖 — 구분자 요구는 과수용 통제 장치 |
| `~` 범위 · `+` 접속 | 3 | 축 밖 |
| 닫히지 않은 괄호(여러 줄에 걸친 한정어) | 1 | 정상 거절 |

**축 하나가 M2에서 넓어졌다 — 새 축이 아니라 두 선언 축의 합성이다.** `plan.md` §B는 축 ③과
④a를 각각 선언하되 **둘의 순서를 고정하지 않는다.** 코퍼스는 양쪽 순서를 다 쓴다.

```
**AC-CSS-001-01 (isolation-validity)** —     한정어가 굵은 글씨 **안**
**AC-HFC-001a** (REQ-HFC-001):               한정어가 굵은 글씨 **뒤**
```

첫 순서가 잔여 134 중 **53건**이었다. `\*{0,2}`를 한정어 양쪽에 두어 두 순서를 모두 받는다.
RED는 `probe/m2-red-composition.txt`(`EXIT=1`)에 있고, 거기서 둘째 순서는 이미 통과하므로 이
수정은 한 순서를 다른 순서로 **바꾼 것이 아니라 합성을 넓힌 것**이다.

**반대편 경계는 주장이 아니라 테스트로 그었다.** 대괄호 한정어와 `↔`는 실재하는 코퍼스
모양이지만 계속 **거절**한다(`TestT528BracketQualifierStaysRejected`). 나중에 이 결정을
뒤집을 카드는 이 테스트를 **의도적으로 지워야** 한다.

### M3 — 무회귀 대조 (파서 편집 **이전**에 고정된 before-image 대비)

before-image는 M0 커밋 `d3f705d99`에 있고 파서 첫 편집(M1, `6caa5d887`)보다 **앞선다**.
커밋 순서가 REQ-ACA-001-015의 순서 판정 근거다.

| 관측 | 결과 |
|---|---|
| 18개 대조군 파일 수 | 18 / 18 |
| 18개 대조군 **루트 AC 집합 + 트리 모양** diff | **0줄** (`diff` exit 0) |
| **비공허성 대조 — 같은 방법을 807개 전체에** | **198줄 변경** |
| 코퍼스 전역 루트 AC 모양 **손실** | **0** |
| 코퍼스 전역 루트 AC 모양 획득 | 904 |
| 기준선 REQ 매핑 행 **탈락** | **0** (216 → 503) |

비공허성 대조가 없으면 「0줄 diff」는 「회귀 없음」과 「비교가 아무것도 보지 않는다」를 가르지
못한다. 같은 명령이 전체에서 198줄을 보고하므로 방법은 변화를 검출한다.

두 위험이 **실측으로** 발화하지 않았다: `buildTree`의 중복 ID 폐기(수용은 늘고 특정 선언은
사라지는 조합)는 손실 0으로 나타나지 않았고, 괄호 한정어 소비는 기준선 매핑을 한 행도
움직이지 않았다. 후자가 안전한 이유는 구조적이다 — 넓힘 이전 앵커는 ID 직후에 곧바로 `:`를
요구했으므로 한정어를 가진 줄은 애초에 수용된 적이 없다.

### M4 — 뮤턴트가 경계를 그린다

정본 `probe/mutants/results.md`. 기대는 `probe/mutants/expectations.md`에 **주입 이전에**
적었고 이후 수정하지 않았다. **두 실행을 모두 남긴다** — 1차가 실제 결함 2건과 무효 뮤턴트
1건을 찾았고, 그것을 지우면 깔끔한 2차만 남아 경계가 작도에서 주장으로 격하된다.

| # | 뮤턴트 | 사전 선언 | 1차 | 2차 |
|---|---|---|---|---|
| 1 | 숫자 꼬리 요구 제거 | 적발 | 적발 | 적발 |
| 2 | `—` 제거 | 적발 | **주입 실패** | 적발 |
| 3 | 괄호 한정어 제거 | 적발 | **주입 실패** | 적발 |
| 4 | 굵은 글씨 허용 제거 | 적발 | 적발 | 적발 |
| 5 | 하위 ID 접미 제거 | 적발 | **주입 실패** | 적발 |
| 6 | `findACSectionStart`→0 **+** `##` break 제거 | 적발 | **무효(빌드 실패)** | 적발 |
| 7 | `##` break만 제거 | 적발 | **미적발** | 적발 |
| 8 | `AC-` 접두를 임의 대문자 토큰으로 | **미적발 가능** | 미적발 | 적발(간극 해소 후) |

1차에서 드러난 것 넷:

- **뮤턴트 2·3·5는 주입되지 않았다.** 앵커 문자열이 `parser.go`에 두 번 나온다 — 정규식과
  그것을 설명하는 문서 주석. 드라이버가 「정확히 1회」를 요구해 테스트를 돌리지 않고
  `INJECTION-FAILED`을 냈다. **이 구분이 드라이버의 존재 이유다**: 주입되지 않은 뮤턴트의
  초록은 「가드가 못 잡았다」와 구별되지 않으며, 그것을 미적발로 적었다면 이 카드의 여섯 번째
  「돌지 않은 측정이 낸 자신 있는 수」가 됐다.
- **뮤턴트 6은 무효였다.** `return i + 1` → `return 0`이 루프 변수를 미사용으로 만들어
  `declared and not used: i` 빌드 실패. 종료 코드는 1이었고 드라이버는 적발로 셌지만, 빌드
  실패는 **어떤 가드가 무언가를 알아챈 증거가 아니다.** `return i * 0`으로 고쳤고, 8개 로그
  전부에 대해 `grep -c 'build failed'`가 **0**임을 확인했다.
- **뮤턴트 7이 실제 결함을 찾았다 — 그러라고 돌린다.** 스코핑 픽스처가 절 밖 미끼를 AC 절
  **앞에만** 두었다. 앞쪽 미끼는 두 번째 이유로도 도달 불가다(`findACSectionStart`가 헤딩
  **뒤** 인덱스를 돌려주므로 앞 줄은 스캔되지 않는다). 즉 그 픽스처는
  `findACSectionStart`를 두 번 검사하고 `extractACLines`를 한 번도 검사하지 않았다.
  AC 절 **뒤** `## Notes`에 미끼를 추가해 닫았다.
- **뮤턴트 8의 사전 선언이 맞았다.** `AC-` 접두를 주장하는 테스트가 없어 `REQ-` · `TEST-`
  불릿을 수집하도록 넓혀도 전부 초록이었다. `TestT528ACPrefixRequired`로 닫았다.

**뮤턴트 8이 그리는 경계는 보이는 것보다 좁다 — 지우지 않고 적는다.** 그 뮤턴트는 접두를
`[A-Z]+-`로 바꾸므로 `REQ-`·`TEST-`는 들어오지만 `M1-`은 숫자 때문에 들어오지 않는다. 3케이스
중 2케이스만 실패했다. 가드는 **문자만으로 된 접두**를 덮고 영숫자 접두에 대해서는 아무 말도
하지 않는다.

**뮤턴트 잔재 0**: 두 실행 모두 종료 후 `git diff --stat -- internal/spec/parser.go` 무출력.

### M5 — 섹션 스코핑 불변

`TestT528SectionScopingInvariant`. 한 픽스처에 세 줄을 같은 모양으로 둔다 — AC 절 **앞**
(`## Background`), AC 절 **안**, AC 절 **뒤**(`## Notes`). 안쪽만 수집되고 앞뒤는 수집되지
않는다. 안쪽 대조군이 같은 문서에 있어야 「수집 0」이 스코핑 때문인지 문법 때문인지 갈린다.
`findACSectionStart`는 편집되지 않았다(M8 diff 목록이 근거).

### M6 — 과수용 실측

한 방향이 아니라 **세 방향**에서 쟀다. 뮤턴트가 이 축을 그리지 못하기 때문이다 — 뮤턴트는 전부
기능 **제거**인데 과수용은 기능이 **작동할 때** 일어난다(`plan.md` §C.1).

| 측정 | 모집단 | 비선언 발견 |
|---|---|---|
| 새로 수용된 줄의 무작위 표본(seed 528) 열람 | 869 중 40 | **0** |
| 폐기·교차참조 표지 표적 검색 | 869 전수 | **0** |
| AC 절 안 실제 **비-AC 불릿** 전수 통과 | 335 (AC 절 불릿 1502 중) | **0** |

세 번째가 정본이다. 표본이 아니라 코퍼스가 실제로 가진 비선언 불릿 **전수**이며,
`TestT528NonDeclarationBulletsCorpusSweep`이 트리에 남아 회귀 가드가 된다. 줄은 발명하지
않았다 — 발명한 줄이 거절되는 것은 코퍼스에 대해 아무 말도 하지 않는다. 손으로 고른 8줄은
`TestT528NonDeclarationBulletsStayRejected`에 핀했다.

**[HARD] 이 측정의 첫 결과는 거짓 0이었다.** 「AC 절 안, AC- 토큰을 가진 비선언 불릿 = 0건」이
나왔는데, 원인은 `filelist.txt` 경로가 `internal/spec` 기준 상대경로라 워크트리 루트에서
연 모든 파일이 조용히 실패한 것이었다. 대조군(「AC 절을 가진 파일 수」)을 함께 세자
**0**이 나와 탐지기가 아예 안 돌았음이 드러났다. 경로를 고치니 807 열림 / 361 파일 /
불릿 1502 / 비선언 335. **이 카드에서 여섯 번째 사례이며, 여섯 번 모두 수를 다시 읽어서가
아니라 재료를 열어서 잡혔다.** 그래서 코퍼스 스윕 테스트는 로스터가 없거나 짧으면
**통과가 아니라 실패**한다.

### M7 — CLI 렌더와 하드 에러 게이트

**[HARD] 순서가 판정의 일부다. 양성 대조군이 먼저다.**

**1단계 — 탐지기가 발화함을 보인다.** 중복 AC ID를 고의로 만든 픽스처
(`probe/dupfixture/`, `.moai/specs/`가 아니라 `.moai/reports/` 아래에 두어 자기가 검증하는
수치를 움직이지 않는다):

```
$ /tmp/t528-moai-after spec view SPEC-T528-DUPFIX-001
   ERROR
  Parse error: duplicate acceptance criteria ID: AC-DUPFIX-001 at depth 0.
RC=1
```

**2단계 — 그 뒤에야 코퍼스 0을 읽는다.** 넓힌 파서로 빌드한 바이너리
(`go build -o /tmp/t528-moai-after ./cmd/moai`, `BUILD_EXIT=0`, 트리 `13fda0f6e`)로 807개 전수:

| 관측 | 넓힘 전 | 넓힘 후 |
|---|---|---|
| `parse error:` | 442 | 442 |
| 그중 `DuplicateAcceptanceID`/깊이 부류 | 0 | **0** |
| `spec.md not found` (`_archive/` 주소 지정 한계) | 6 | 6 |
| **기준선에 없던 신규 하드 에러** | — | **0** |
| 해소된 하드 에러 | — | 0 |

SPEC ID 집합을 `comm`으로 비교해 **신규 0 · 해소 0** — 집합이 동일하다.
442는 전부 `acceptance criteria section not found`(헤딩 축, 범위 밖)이며 M0 before-image와
같다. 산출물: `probe/before/specview-errors.txt` · `probe/after/specview-errors.txt`.

**렌더 (AC-ACA-001-012).** 기준선에서 전맹이던 101개 중 **92개**가 이제 렌더된다.
지목: `SPEC-AUDIT-SNAPSHOT-001`.

```
$ /tmp/t528-moai-before spec view SPEC-AUDIT-SNAPSHOT-001
No acceptance criteria found in SPEC-AUDIT-SNAPSHOT-001

$ /tmp/t528-moai-after spec view SPEC-AUDIT-SNAPSHOT-001
Acceptance Criteria for SPEC-AUDIT-SNAPSHOT-001:
├── AC-AUDIT-SNAPSHOT-001
│ └── AC-AUDIT-SNAPSHOT-001.a: sticky cache — past-24h unchanged-hash skip still fires.
├── AC-AUDIT-SNAPSHOT-002
│ └── AC-AUDIT-SNAPSHOT-002.a: per-tier skip threshold — a 0.78 Tier M SPEC is skip-eligible.
├── AC-AUDIT-SNAPSHOT-003
│ └── AC-AUDIT-SNAPSHOT-003.a: clean sync emits binding verdict with no cold sync-auditor spawn.
└── AC-AUDIT-SNAPSHOT-004
 └── AC-AUDIT-SNAPSHOT-004.a: single test/lint/vet/cover run shared across 3 consumers; new SHA invalidates.
```

두 바이너리 모두 이 트리에서 빌드했다(`before`는 M0 시점 = 파서 미편집, `after`는 `13fda0f6e`).
설치된 구버전으로 잰 값이 아니다.

### M8 — CoverageRule delta와 바이트 동일성

**계측 지점**: `moai spec lint`가 아니다. `lint.go:636`이 `criteria, _ :=`로 오류 슬라이스를
버리므로 그 경로에서는 `DuplicateAcceptanceID`와 깊이 초과가 **발화 여부와 무관하게 0**으로
관측된다 — 판정 불가능한 계측이다(감사 D3). `ParseAcceptanceCriteria`의 **두 번째 반환값을
직접 읽는** 프로브를 썼다(`zz_t528_coverage_delta_test.go`).

넓힘 전후는 같은 프로브를, `parser.go`를 넓힘 이전 blob(`0c2ffda89`)으로 잠시 되돌려 각각
측정했다. 복원 후 `git status --short internal/spec/parser.go` 무출력.

| 관측 | 넓힘 전 | 넓힘 후 | delta |
|---|---|---|---|
| **`CoverageIncomplete` finding** | **3636** | **3531** | **−105** |
| `ParseAcceptanceCriteria`가 ≥1 기준을 낸 파일 | 18 | 117 | +99 |
| `DuplicateAcceptanceID` / 깊이 / 기타 치명 | 0 | **0** | 0 |
| `acceptance criteria section not found` | 446 | 446 | 0 |

방향은 예측대로 **감소**다 — 수집기를 넓히면 커버된 REQ가 늘어난다. 크기를 함께 적는 것이
이 AC의 요구이며, 증가가 아니므로 §F Tier L 격상 조건은 발동하지 않는다.

**[HARD] 이 측정의 첫 결과도 거짓 0이었다.** `CoverageIncomplete = 0`이 나왔는데, 원인은
`&SPECDoc{Path, Criteria}`만 만들고 `REQs`를 채우지 않은 것이었다 — `CoverageRule.Check`는
`len(doc.REQs) == 0`에서 조기 반환하므로 규칙이 807번 전부 즉시 빠져나갔다. 「모든 REQ가
커버됐다」와 「규칙이 한 번도 돌지 않았다」가 같은 0으로 나타난 것이다. `lint.go:633`과 같이
`parseREQsWithProvenance(body)`로 채워 고쳤고, 테스트에 **양성 대조군**을 박았다 — 커버되지
않은 REQ 하나짜리 문서가 finding 1건을 내지 못하면 코퍼스 수치를 읽기 전에 `t.Fatalf`로 멈춘다.

**바이트 동일성** — 카드 착수 트리 `52f863f36` 대비:

```
$ git diff --stat 52f863f36..HEAD -- internal/spec/lint.go        (무출력)
$ git diff --stat 52f863f36..HEAD -- internal/cli/spec_view.go    (무출력)
```

경로 한정 diff가 빈 것은 「변경 없음」과 「경로를 잘못 짚음」 둘 다로 나타나므로 **전체 변경
파일 목록을 대조로** 함께 남긴다: 두 파일 모두 목록에 **부재**하고, 프로덕션 편집은
`internal/spec/parser.go` **하나**뿐이다(나머지는 `_test.go` 3개 + `.moai/` 산출물).

### 전 마일스톤 공통 — 패키지 실행 (AC-ACA-001-013, `plan.md` §D.2의 5요소)

| 요소 | 착수 기준선 | 착지 |
|---|---|---|
| 명령 | `go test ./internal/spec/... -count=1` | 동일 (`-timeout 1200s`) |
| 트리 SHA | `52f863f36` | `13fda0f6e` + 미커밋 M4-M8 |
| 종료 코드 | `EXIT=0` | **`EXIT=0`** |
| 소요 | `ok … 83.212s` | `ok … 71.293s` |
| `--- FAIL` 수 | 0 | **0** |
| `panic:` / `test timed out` 수 | 0 | **0** |
| `(cached)` 수 | 0 | **0** |

산출물: `probe/pkg-baseline-20260908.txt` (기준선) · `probe/final-pkg-run.txt` (착지).

소비자 측: `go test ./internal/cli/ -run 'SpecView|Acceptance|SpecSecHarden|StdoutClean' -count=1`
→ `ok … 0.947s`, `EXIT=0` (`probe/final-cli-run.txt`). `go vet ./internal/spec/` → `EXIT=0`.
`gofmt -l internal/spec/` → 무출력.

**로컬에서 `go test ./...`를 돌리지 않았다.** 전 패키지 판정은 CI 몫이다.

**귀속 (`plan.md` §D.3)**: 같은 패키지에 두 번째 카드(t518)가 동시에 들어와 있다. 이 카드의
실행에서 실패는 **한 건도 나지 않았으므로** 귀속 절차를 발동할 사건이 없었다. 이 카드의 실제
반경은 측정값이다 — `internal/spec/parser.go` 1개 + 이 카드가 만든 `zz_t528_*_test.go` 4개.
t518이 편집 중인 파일과 겹치는 것은 없다.

---

## §E.2.1 AC PASS/FAIL 매트릭스

| AC | 주제 | Actual Output | Status |
|---|---|---|---|
| AC-ACA-001-001 | 회수량 재측정 | 216/1167(18.5%) → 1085/1167(93.0%), 판별식 B, 트리 `13fda0f6e`, 분모 `after/filelist.txt` 807; 잔여 82 전 부류 분해 + 표본 | **PASS** |
| AC-ACA-001-002 | 무회귀 대조군 + 순서 | 18/18 파일, 루트 AC 집합·트리 모양 diff 0줄; 비공허성 대조 198줄; before-image 커밋 `d3f705d99` < 파서 첫 편집 `6caa5d887` | **PASS** |
| AC-ACA-001-003 | ID 분절 문법 | 2~5분절·알파벳 중간 분절 7케이스 통과, `AC-FOO-BAR` 부정 케이스 거절; 뮤턴트 1 적발 | **PASS** |
| AC-ACA-001-004 | 하위 ID 접미 의미론 보존 | `.a`/`.a.i` 인식 + 트리 모양 `AC-…-01(AC-…-01.a)` vs `AC-…-02.a` 동일; 뮤턴트 5 적발 | **PASS** |
| AC-ACA-001-005 | 굵은 글씨 래퍼 | 코퍼스 3줄 수용; `TrimLeft` 호출 줄 diff 부재; 뮤턴트 4 적발 | **PASS** |
| AC-ACA-001-006 | 섹션 스코핑 불변 | 절 앞/안/뒤 3줄 픽스처, 안쪽만 수집; 뮤턴트 6·7 모두 적발 | **PASS** |
| AC-ACA-001-007 | `lint.go` 바이트 동일 | 경로 한정 diff 무출력 + 전체 목록에 부재 | **PASS** |
| AC-ACA-001-008 | 괄호 한정어 | `(A1)`·`(REQ-001)`·굵은 글씨 조합 수용; 기준선 REQ 매핑 행 탈락 0 (216→503); 뮤턴트 3 적발 | **PASS** |
| AC-ACA-001-009 | 구분자 집합 | `:`·`—`(U+2014)·`–`(U+2013) 코드 포인트로 확인; 엔대시는 설계 결정이고 관측 아님을 기록; 뮤턴트 2 적발 | **PASS** |
| AC-ACA-001-010 | `CoverageRule` delta | 3636 → 3531 (**−105**); 양성 대조군 포함; 계측 지점을 lint 대신 두 번째 반환값으로 정정 | **PASS** |
| AC-ACA-001-011 | 뮤턴트 경계 | 8건 사전 선언, 두 실행 모두 보존; 최종 8/8 적발; 뮤턴트 8의 좁은 경계 기록; 잔재 0 | **PASS** |
| AC-ACA-001-012 | `spec view` 실제 렌더 | `SPEC-AUDIT-SNAPSHOT-001` before/after verbatim; 전맹 101 중 92 회복 | **PASS** |
| AC-ACA-001-013 | 패키지 초록 + 검증 범위 | `EXIT=0` · FAIL 0 · panic/timeout 0 · cached 0 · 트리 SHA, 착수 기준선과 나란히 | **PASS** |
| AC-ACA-001-014 | 과수용 실측 | 세 방향(표본 40 / 표적 검색 869 / 비선언 불릿 전수 335) 모두 **0**; 코퍼스 유래 픽스처 트리에 상주 | **PASS** |
| AC-ACA-001-015 | CLI 하드 에러 | 양성 대조군 **선행** 발화 확인 → 코퍼스 신규 하드 에러 **0**, 중복/깊이 부류 **0** | **PASS** (아래 정정 참조) |
| AC-ACA-001-016 | 측정 무결성 | 프로브 커밋됨, `[no tests to run]` 부재, `--- PASS: TestT528Anchor` 존재; 모든 수치에 3요소 부기 | **PASS** |

**16/16 PASS.**

### [HARD] AC-ACA-001-015 판정 근거의 정정

이 AC와 리드의 중단 조건은 「하드 에러 == 0」을 문면 그대로 요구한다. **넓힘 이전 코퍼스가
이미 442건**이므로 절대값 0은 오늘 달성 가능한 수가 아니었다(M0 §E.2). 문서가 핀한 0은
`probe/duplicate-20260908.txt` — **중복 ID 재료**를 잰 값 — 에서 왔고, 그것은 진짜로 0이지만
CLI 하드 에러와 다른 양이다.

PASS로 읽은 근거는 REQ-ACA-001-013의 문면이다: **「기준선에 없던 하드 에러」**. 그 값이 0이며,
442는 전부 헤딩 축(범위 밖·수리 금지)이고 넓힘이 새로 만들 수 있는 유일한 부류(중복/깊이)도
0이다. 리드에게 이 정정을 M0 직후 보고했고 문면 그대로의 해석을 원하면 멈추겠다고 밝혔다.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: pending-backfill-run-final
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 18
new_warnings_or_lints_introduced: 0
total_run_phase_files: 5
production_files_changed: 1
m1_to_mN_commit_strategy: "M0 control commit -> M1 RED+GREEN -> M2/M3/M6 -> M4-M8 evidence"
```
