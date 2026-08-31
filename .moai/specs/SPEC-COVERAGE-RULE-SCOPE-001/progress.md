# SPEC-COVERAGE-RULE-SCOPE-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-30
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
baseline_tree: ee50984ab
open_clarifications: 2   # plan.md §C.1 방향, §D 심각도 안
```

## §E.2 Run-phase Evidence

### 측정 기준 트리 — 이 절의 모든 수치는 여기 적힌 트리에서 잰 것이다

이 절의 코퍼스 수치는 **`origin/develop`을 흡수하기 전** 트리에서 쟀다. 흡수는 이후 실제로
일어났고(병합 커밋 `d5ca9582d`), 재측정 결과는 **§병합 트리 재측정 — 상속 대 귀속 분리**에
있다. 예측이 아니라 관측이다: 병합 트리에서 `--strict`는 **rc=1 / error 2**로 끝나며, 그 2건은
t336이 착지시킨 SPEC의 `ArtifactStatusFieldForbidden`으로 **t362 귀속이 아니다**.

따라서 아래 표의 `error 0` · `rc 0` 값들은 **명시된 트리에 대한 참**이고 **현재 상태의 서술이
아니다.** 병합 후 값과 **대조**하되 이 값을 **대체**해서는 안 되며, 반대로 병합 후의 적색을
이 카드에 귀속시켜서도 안 된다. 그 둘을 가르는 검사는 `doc.REQs` 소비 4개 코드의 병합 가로지르기
불변성이며, 아래 절에 기록돼 있다.

| 수치 | 측정 트리 | 근거 |
|---|---|---|
| M1 전 코퍼스 실측 (704 / 63 / 1,085 / 846 / 825) | `68ecbfe4a` | `.moai/reports/t362/m1-corpus-measurement.txt` |
| M2 단독 코퍼스 무변화 (177건, plain rc 0, strict rc 0) | `130846ab2`의 트리 (빌드 시점 HEAD `0d102d7c7` + 미커밋 M2 변경 = 그 커밋의 트리; `git diff --stat 130846ab2 -- internal/ cmd/ pkg/` 공집합으로 확인) | `.moai/reports/t362/m2-lint-corpus{,-strict}.json` |
| **M3 전 코퍼스 (1,090건 / warning 1,090 / error 0 / 비자문 warning 0 / plain rc 0 / strict rc 0)** | **`130846ab2`의 트리** (위와 동일 근거) | `.moai/reports/t362/m3-lint-corpus{,-strict}.json` |
| M3 코드별 내역 (846 / 113 / 43 / 25 / 24 / 18 / 14 / 6 / 1) | `130846ab2`의 트리 | 위와 동일 |
| 6건 `InvalidREQID` 원문 (REQ-256K-001..006) | `130846ab2`의 트리 | `m3-lint-corpus.json` |
| 코퍼스 뮤테이션 탐침 (shipped 6 / mutant 0 / delta 6) | `2f3bd1a31`의 트리 (모집단 동결 후 재생성) | `.moai/reports/t362/m2-gate0-decomposition.txt` `[F]` |
| Gate 0 분해 (825 / 519 / 200 / 106 / 22 / 21) | `2f3bd1a31`의 트리 (모집단은 동결 리터럴이므로 트리와 무관하게 재현) | 같은 보고서 `[A]`~`[E]` |
| 숫자-시작 도메인 0/706 | `7ba784171`의 트리 | `ls -d .moai/specs/SPEC-* \| sed 's\|.*/SPEC-\|\|' \| grep -cE '^[0-9]'` |
| C2 미러 유/무 대조 (0건 vs 8 advisory warning, 양쪽 rc 0) | `130846ab2`의 트리 | `.moai/reports/t362/c2-nomirror-strict.txt` |
| 패키지 게이트 (`go test` ok / `golangci-lint` 0 issues / `go vet` rc 0 / 커버리지 89.4%) | `130846ab2` 이후 매 커밋에서 재실행, 최종 `2f3bd1a31` | 본문 각 항목 |


### M1 — 넓힌 REQ 파서 구현 + 전 코퍼스 실측 (2026-08-30)

측정 트리: 워크트리 `.claude/worktrees/t362`, 브랜치 `WT-coverage-rule-scope`, HEAD `68ecbfe4a`.
(plan/acceptance가 기준으로 삼은 `ee50984ab`는 그 뒤 develop 병합으로 밀렸다. M1의 모든 수치는
`68ecbfe4a` 기준이며, 코퍼스가 702 → 704로 늘어난 것이 그 차이의 관측 가능한 형태다.)

**M1 범위 한정**: 넓힌 패턴은 `parseSPECDoc`에 **배선하지 않았다.** `parseREQs` /
`reqLinePattern`은 그대로다. 배선은 §D 심각도 결정(M3) 이후의 일이다.

| AC | 상태 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-CRS-001-001 | PASS | `go test ./internal/spec/... -run TestParseREQsWide_SixObservedShapes -v` | 10/10 서브테스트 PASS (여섯 형태 + 좁은 형태 + bold/asterisk/indent 변형) |
| AC-CRS-001-002 | PASS | `MOAI_T362_CORPUS_SCAN=1 go test ./internal/spec/... -run TestCorpusREQWideningMeasurement` | `.moai/reports/t362/m1-corpus-measurement.txt` — 아래 실측표 |
| AC-CRS-001-003 | 해당 없음 (M2 소관) | — | M1은 넓힌 패턴을 배선하지 않으므로 `InvalidREQID` 실발화가 없다. 다만 배선 시 발생할 값을 **825건**으로 미리 쟀다(측정 [6]) |
| AC-CRS-001-004 | PASS | `go test ./internal/spec/... -run TestParseREQsWide` (뮤테이션 전/후) | RED 증거 `.moai/reports/t362/m1-red-evidence.txt` |

**실측표 (`68ecbfe4a`, glob `.moai/specs/SPEC-*/spec.md`)**

| 측정 | 좁은 현행 | 넓힌 뒤 (실측) | 계획 추정 |
|---|---|---|---|
| `spec.md` 총수 | 704 | 704 | 702 |
| REQ 정의 줄을 갖는 파일 | 16 | **63** | 62 |
| REQ 정의 줄 총수 | 239 | **1,085** | 1,077 |
| 예상 `CoverageIncomplete` | 0 | **846** | 741 |
| 미커버 REQ를 가진 SPEC | — | **47** | 47 |
| 배선 시 `InvalidREQID` | 0 | **825** | 미측정 |
| NARROW ⊄ WIDE (미포함 항목) | — | **0** | 미측정 |

**741 방향 단언의 판정: 예측대로다.** 실측 846 ≥ 추정 741 — 741은 **하한으로 행동했다**
(초과분 +105, +14.2%). acceptance.md AC-CRS-001-002가 요구한 방향 검증을 이로써 만족한다.

**포함관계**: 넓은 집합은 좁은 집합의 **진부분집합 관계를 만족한다** — 코퍼스 704개 전체에서
NARROW에는 있고 WIDE에는 없는 `(파일, reqID)` 쌍이 **0건**이다. 넓힌 패턴은 리스트 항목에
앵커돼 있고(`^\s*[-*]\s+`) 현행 `reqLinePattern`은 앵커가 없어 문장 중간 하이픈에도 걸리므로
이론상 포함관계가 깨질 수 있으나, 실제 코퍼스에서는 그런 줄이 하나도 없다.

**Tier 분포 (새로 수집되는 47개 파일)**: `(absent)` 17 / `M` 8 / `L` 7 / `2` 7 / `3` 5 / `S` 3.
`tier: 2` / `tier: 3`은 S/M/L 열거형 밖의 값으로, 코퍼스에 존재하는 별건 결함이다(이 SPEC 범위 밖).

**실행 동작 불변성**: 이 트리에서 빌드한 바이너리의 전 코퍼스 lint 출력이 병합 전 baseline과
**바이트 동일**하다(`cmp` exit 0, 140,223 bytes). finding 히스토그램: MovingRefUnpinned 113 /
MissingExclusions 24 / StatusGitConsistency 18 / FrontmatterInvalid 14 / LegacyEARSKeyword 7 /
OwnershipTransitionInvalid 1.

**품질 게이트**: `go vet ./internal/spec/...` exit 0 · `go test ./internal/spec/... -count=1` ok 32.759s ·
`golangci-lint run ./internal/spec/...` 0 issues · `gofmt -l` 신규 3파일 모두 clean.

**신규 파일**: `internal/spec/lint_req_widen.go`, `internal/spec/lint_req_widen_test.go`,
`internal/spec/lint_req_widen_corpus_test.go`.

**M1 Gaps**: ① 넓힌 패턴은 배선되지 않았으므로 846은 **시뮬레이션 값**이다 — 실제 규칙 발화는
M3 배선 후에만 관측된다. ② 825건 `InvalidREQID`는 M2가 닫아야 할 크기이며, M1은 재지 않은
정합 방안을 제시하지 않는다. ③ 넓힌 패턴은 조사자가 고른 하나이며, 다른 넓힘은 다른 수를 낸다.

### M2 — reqIDPattern / 추출 패턴 정합 (2026-08-30)

측정 트리: 워크트리 `.claude/worktrees/t362`, 브랜치 `WT-coverage-rule-scope`, HEAD `0d102d7c7`.

**Gate 0 — 825건 분해.** M2 착수 전에 넓힌 추출이 모으지만 현행 `reqIDPattern`이 거부하는
825건을 전수 분류했다(`.moai/reports/t362/m2-gate0-decomposition.txt`,
`MOAI_T362_CORPUS_SCAN=1 go test ./internal/spec/... -run TestCorpusRejectedREQIDDecomposition`).

| 형태 | 예 | 건수 |
|---|---|---|
| 3-분절, 알파 도메인 | `REQ-AME-001` | 519 |
| 5-분절, 두 분절 alnum 도메인 | `REQ-V3R2-RT-001-001` | 200 |
| 3-분절, alnum 도메인 | `REQ-GD1-001` | 106 |

관측된 형태는 이 셋뿐이다. 도메인 분절은 1개 또는 2개, 숫자 꼬리는 3자리 그룹 1개 또는 2개
— 코퍼스 전체가 이 범위 안에 있다.

**오독(misread) 건수 = 0.** 보고서 `[D]` 절에 **두 판독을 모두** 방출한다:

```
misread_p1_raw_extra_req_tokens=22      ← P1-raw (과대 술어, 폐기됨)
misread_p1_corrected_tokens_only=0      ← P1-corrected (교정 술어)
misread_p2_empty_text=0
misread_p3_non_req_heading=0
misread_verdict_union_corrected=0       ← 판정 = P1-corrected ∪ P2 ∪ P3
misread_union_raw_upper_bound=22        ← 교정 전 상한 (판정 아님)
```

P1-raw(캡처 본문에 다른 REQ 토큰이 등장)는 매핑 행을 잡으려던 술어인데, REQ 정의가 본문에서
다른 REQ를 인용하는 것은 정상이므로 잡히지 않는다 — 22건 전부가 그런 정의 줄이다.
**교정 술어 P1-corrected는 사람이 읽어 버린 것이 아니라 기계적으로 좁힌 것이다**: 캡처 본문에서
REQ 토큰을 전부 제거하고 이어서 구두점·공백을 제거했을 때 잔여가 **빈 문자열**이면 매핑 행,
산문이 남으면 정의. 진짜 매핑 행은 아무것도 남기지 않는다. 이 술어로 0건이다.

**22건은 `[D]` 절에 `p1_raw=<file>:<line> <ID>` + 원문 줄로 전수 열거된다**(`[E]` 절이 21건
bold-marker 항목을 열거하는 것과 같은 모양). 교정 술어만 방출하면 그 0이 판단이 옳아서 0인지
틀려서 0인지 읽는 쪽에서 구분할 수 없다 — 폐기한 술어의 입력이 남아 있어야 폐기가 검증 가능하다.

**모집단 정의가 얼어 있다는 점이 중요하다.** `[A]`~`[E]`의 모집단은 **M2 이전 검증 패턴**
(`^REQ-[A-Z]{2,5}-\d{3}-\d{3}$`, 보고서에 리터럴로 고정)이 거부한 집합이다. 초기 판본은 이
모집단을 **살아 있는 `reqIDPattern`** 으로 정의했고, 그래서 M2가 착지하는 순간 `[A]`~`[D]`의
모든 수치가 825/519/200/106/22에서 6/6/0으로 **표제를 그대로 둔 채 조용히 재기준화**됐다.
움직이는 정의에 대고 잰 측정이며, 이 SPEC이 문서화하려는 결함을 증거 도구 안에서 재생산한
것이다. 이제 리터럴로 얼어 있어 움직이지 않는다. M2 이후 잔여(6건)는
`rejected_by_shippedPattern_postM2`로 `[D]` 끝에 **따로** 보고한다.

**21행 차이의 원인 = 굵게 표시 마커.** 넓힘 1,085 − 좁힘 239 = 846인데 `InvalidREQID`는
825다. 차이 21은 "형태상 좁은 패턴에 맞지만 좁은 **줄** 패턴이 수집하지 못한" ID이며,
**21건 전부 `SPEC-UTIL-002/spec.md`의 `- **REQ-UTIL-002-0NN**: …` 줄**이다. 현행
`reqLinePattern`은 ID 직후에 `\s*:`를 요구하는데 그 사이에 `**`가 끼어 매치가 끊긴다.
들여쓰기도 빈 본문도 아니고 굵게 마커 하나가 원인이다. (보고서 `[E]` 절에 21건 전수.)

**채택안: (i) 검증을 추출보다 좁게 유지.** `reqIDPattern`을
`^REQ-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)?-\d{3}(?:-\d{3})?$`로 넓혔다. 추출과 **정확히**
맞추는 (ii)안은 거부했다 — `doc.REQs`를 채우는 것이 그 추출이므로, 정합시키는 순간
`InvalidREQIDRule`은 판정 대상 전부가 구성상 통과하는 공허한 규칙이 되고, 그 규칙의 미실행은
성공과 구분되지 않는다(`verification-completeness.md` §1.1). 이 SPEC이 문서화하려는 결함
바로 그 형태다.

남는 거부 클래스(추출은 모으지만 검증은 거부): 도메인 분절이 글자로 시작하지 않는 경우,
도메인 분절 3개 이상, 숫자 꼬리가 3자리 그룹 1~2개가 아닌 경우. 이 클래스가 실제 추출로
도달 가능함을 `TestReqIDPattern_RejectsShapesTheExtractionAccepts`(뮤테이션 탐침)와
`TestReqIDPattern_ExtractionRejectionClassIsReachable`(실규칙 경유)이 함께 강제한다.
추출과 검증을 정합시킨 뮤턴트는 전자에서 RED가 된다.

**공허성은 추론이 아니라 코퍼스 뮤테이션으로 관측했다.** 채택한 패턴의 모양에서 "거부 클래스가
남아 있으니 공허하지 않다"고 추론하는 것으로는 부족하다 — 그 클래스에 실제 문서가 도달하는지는
별개의 사실이다. Gate 0 `[F]`에 option (ii) 뮤턴트(검증을 추출과 **정확히** 정합시킨
`^REQ-[A-Z0-9]+(?:-[A-Z0-9]+)*-\d+$`)를 같은 스캔 안에서 함께 재도록 넣었다:

```
blast_InvalidREQID_proposedPattern  narrow=0  wide=6   ← 채택 패턴 (현재 shipped)
blast_InvalidREQID_vacuousMutant    narrow=0  wide=0   ← option (ii) 뮤턴트
mutant_probe_delta_wide=6
```

델타 0이었다면 채택 패턴의 거부 클래스는 정규식이 뭐라고 적혀 있든 **실제 문서로는 도달 불가**,
즉 공허한 규칙이었을 것이다. 델타는 6이다.

**발화하는 6건과, 그것이 진짜 위반이라는 근거.** M3 전 코퍼스 실행에서 실제로 나온 finding
(`jq -r '.[] | select(.code=="InvalidREQID")' .moai/reports/t362/m3-lint-corpus.json`):

```
warning  advisory=true  .moai/specs/SPEC-HANDOFF-CTXGUIDE-001/spec.md:40  REQ ID "REQ-256K-001" does not match pattern …
…                                                                  :41  "REQ-256K-002"
…                                                                  :42  "REQ-256K-003"
…                                                                  :43  "REQ-256K-004"
…                                                                  :44  "REQ-256K-005"
…                                                                  :45  "REQ-256K-006"
```

전부 한 SPEC의 도메인 분절 `256K` — 숫자로 시작한다. 이것이 규약 위반이라는 근거도 측정이다:
`ls -d .moai/specs/SPEC-* | sed 's|.*/SPEC-||' | grep -cE '^[0-9]'` → **0** (706개 중 0개).
숫자로 시작하는 도메인 분절은 SPEC ID 수준에서도 코퍼스 전체에 존재하지 않으며,
`specIDPattern`(`^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`)이 같은 글자-시작 규칙을 기계적으로 못박고
있다. REQ 정의 1,085건 중 숫자-시작 도메인은 이 6건뿐 — 표현되지 않은 규약이 아니라 이상치다.
AC-CRS-001-003의 "0이 아닌 경우 각 건이 실제 규약 위반임을 개별 근거와 함께 제시한다"를
이로써 만족한다.

**삭제 후보 판정: 해당 없음.** 리드의 addendum은 option (ii)를 택해 잔여 가치가 0일 때
`InvalidREQIDRule`을 **삭제 후보로 명시**하라고 요구한다. 이 카드는 option (i)를 택했고 잔여
발화가 6건이므로 그 조항은 발동하지 않는다. 근거를 남기는 이유는 판정이 보이게 하기 위해서다:
만약 델타가 0이었다면 `moai spec lint --help`가 광고하는 "REQ ID uniqueness"가 아무것도 검사하지
않는 규칙 위에서 계속 돌게 되고, **공허한 파서를 고치면서 공허한 규칙을 남기는 것은 수리 안에서
같은 결함을 재생산하는 일**이다. 삭제 자체는 이 카드의 결정이 아니며 리드 소관이다. 같은 취지를
`internal/spec/lint.go`의 `reqIDPattern` 헤더에도 기록했다.

`DuplicateREQIDRule`은 공허하지 않으며 계속 동작한다 —
`TestDuplicateREQIDRule_StillFiresOnWidenedShapes`가 넓어진 형태(`REQ-HOOK-001`)에 대해
중복 탐지가 실제로 발화함을 강제한다(검증을 넓히기 전에는 `continue`에 걸려 0건이었다).

| AC | 상태 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-CRS-001-003 | PASS **(`130846ab2` 트리 기준, 병합 전)** | `moai spec lint --json` (M2 바이너리, 전 코퍼스) | `InvalidREQID` 0건 / 총 177건 / rc=0 · `--strict` rc=0 — 그 트리의 baseline과 코드별 건수 동일. 현재 상태 서술 아님 |
| AC-CRS-001-004 | PASS | `go test ./internal/spec/ -run 'TestReqIDPattern\|TestDuplicateREQIDRule_StillFires' -v` | RED 증거 `.moai/reports/t362/m2-red-evidence.txt`, 이후 4/4 PASS |

**M2 단독 착지 시 코퍼스 무변화**: `moai spec lint --json` 177건, 코드별 내역
MovingRefUnpinned 113 / MissingExclusions 24 / StatusGitConsistency 18 / FrontmatterInvalid 14 /
LegacyEARSKeyword 7 / OwnershipTransitionInvalid 1 — baseline과 동일. `doc.REQs`가 아직 좁은
집합이므로 검증을 넓힌 것만으로는 관측 가능한 변화가 없다(설계대로).

### M3 — 넓힌 파서 배선 + 심각도 처리 (2026-08-30)

측정 트리: **커밋 `130846ab2`의 트리** (빌드 시점 HEAD는 `0d102d7c7`였고 미커밋 M2+M3 변경이
얹혀 있었으며, 그 워킹트리가 `130846ab2`가 되었다 — `git diff --stat 130846ab2 -- internal/ cmd/ pkg/`
공집합으로 확인). 바이너리는 그 트리에서 빌드(`go build -o <scratch>/moai-m3 ./cmd/moai`, rc=0).
**이 절의 코퍼스 수치는 전부 병합 전 값이며 현재 상태의 서술이 아니다** — §병합 트리 재측정 참조.

**배선 전에 드러난 사실: `CoverageRule`은 `doc.REQs`의 유일한 소비자가 아니다.**
Gate 0의 `[F]` 절에서 전 소비자를 좁힘/넓힘으로 시뮬레이션했다:

| 코드 | 심각도 | 좁힘 | 넓힘 |
|---|---|---|---|
| `CoverageIncomplete` | error | 0 | **846** |
| `ModalityMalformed` | error | 0 | **25** |
| `InvalidREQID` (M2 후 shipped 패턴) | error | 0 | **6** |
| `DuplicateREQID` | error | 0 | 0 |
| `LegacyEARSKeyword` | warning | 7 | **43** |

네 error 코드 중 어느 것도 `eraDemotableCodes`에 없다. 따라서 **`CoverageRule`만 자문 처리하는
option A로는 31건의 error가 남아 `develop`을 붉힌다** — REQ-CRS-001-003과 AC-CRS-001-005가
금지하는 결과다. 리드의 구속력 있는 판정(option A)은 `CoverageRule` 범위로 서술돼 있었고, 이
파급은 그 시점에 측정돼 있지 않았다.

**적용한 처리 — option A + 넓힘 전용 자문(REQ-CRS-001-003이 열거한 "자문 처리" 수단).**

1. `CoverageRule`: 발화 지점에서 무조건 `warning` + `Advisory: true` (리드 판정 그대로).
2. `REQEntry.Widened` 출처 플래그: 좁은 패턴이 **같은 줄에서 같은 ID로** 수집하지 못한 항목만 참.
   `parseREQsWithProvenance`가 채운다.
3. `ModalityMalformed` / `InvalidREQID` / `DuplicateREQID` / `LegacyEARSKeyword`:
   `Widened` 항목에 대한 발화만 `warning` + `Advisory: true`로 내린다(`reqFindingSeverity`).
   **좁은 항목의 동작은 바이트 동일하게 보존된다** —
   `TestWidenedOnlyFindingsAreAdvisory`의 "narrow ... stays an error" 서브테스트 2건이 강제.

`eraDemotableCodes` 경유가 아니라 **발화 지점**을 쓴 이유는 t342 헤더가 적은 것과 같다: 그 맵은
`SeverityError`에 대해서만 참조되므로 warning은 애초에 닿지 못하고, 대상 finding들이 어느 era
경로로도 강등되지 않는 modern-era SPEC 위에 앉아 있다.

| AC | 상태 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-CRS-001-001 | PASS | `go test ./internal/spec/ -run TestParseSPECDoc_CollectsWidenedShapes -v` | PASS — 여섯 형태가 **실경로** `doc.REQs`에 도달 |
| AC-CRS-001-002 | PASS | `moai spec lint --json` (M3 바이너리, 전 코퍼스) | 총 **1,090**건. `CoverageIncomplete` 846은 M1 시뮬레이션값과 **일치**(846) |
| AC-CRS-001-005 | PASS **(`130846ab2` 트리 기준, 병합 전)** | `moai spec lint --json` / `moai spec lint --strict --json` | **그 트리에서** plain rc=0, strict rc=0. 심각도 분포: warning 1,090 / error **0**. 비자문 warning **0**건. **이 값은 현재 상태의 서술이 아니다** — `origin/develop` 흡수 후 값은 아래 §병합 트리 재측정 참조 (그 트리에서 rc=1, error 2, 단 **둘 다 t362 귀속 아님**) |

**M3 전 코퍼스 finding 내역** (`.moai/reports/t362/m3-lint-corpus.json`):

```
846 CoverageIncomplete   113 MovingRefUnpinned   43 LegacyEARSKeyword   25 ModalityMalformed
 24 MissingExclusions     18 StatusGitConsistency 14 FrontmatterInvalid   6 InvalidREQID
  1 OwnershipTransitionInvalid                                        총 1090, 전부 warning/advisory
```

baseline 177 → 1,090 (+913 = 846 + 25 + 6 + 36). **기존 6개 코드의 건수는 `LegacyEARSKeyword`
7→43 외에 하나도 변하지 않았다** — 넓힘이 기존 finding을 건드리지 않았다는 관측 가능한 형태다.

**시뮬레이션 대 실측**: Gate 0 `[F]`가 예측한 846 / 25 / 6 / 43이 실측과 **전부 일치**한다.

**품질 게이트**: `go vet ./internal/spec/...` rc=0 · `go test ./internal/spec/... -count=1`
ok 36.755s · `golangci-lint run ./internal/spec/...` `0 issues.`

**부채 기록 ① — `REQEntry.Widened`가 심각도를 결정한다.**

비준된 기제는 **출처 플래그가 심각도를 결정하게** 만든다. 그것이 부채다.

- **부채**: 심각도가 `REQEntry.Widened`에 걸려 있으므로, **자문 범위는 그 플래그를 세울 수 있는
  경로의 집합만큼만 좁다.** 다른 이유로 `Widened = true`를 세우는 코드 경로가 새로 생기면
  자문 범위가 조용히 넓어지고, 그 사실을 알리는 것은 아무것도 없다. 그 경로가 만드는 finding은
  error였어야 할 것이 warning으로 나가고, 리포트에는 그대로 보이므로 억제된 티도 나지 않는다.
- **현재 이를 묶고 있는 불변식 (측정함)**: 비테스트 코드에서 이 플래그를 세우는 경로는
  **정확히 하나**다 — `parseREQsWithProvenance`(`internal/spec/lint_req_widen.go:79`), 그리고
  **좁은 패턴이 같은 줄에서 같은 ID를 수집하지 못했을 때에만** 세운다. 읽어서 심각도를 정하는
  곳도 하나다 — `reqFindingSeverity`(`internal/spec/lint.go:626`). 확인 명령과 출력:

  ```
  $ grep -rn "Widened" internal/ | grep -v "_test.go"
  internal/spec/lint.go:444      # REQEntry 필드 주석
  internal/spec/lint.go:452      #   Widened bool          ← 선언
  internal/spec/lint.go:621      # reqFindingSeverity 주석
  internal/spec/lint.go:626      #   if req.Widened        ← 유일한 판독
  internal/spec/lint_req_widen.go:35, :79                  ← :79 이 유일한 기록
  internal/spec/lint_artifact_status.go:156                # 무관 (영어 단어 "Widened")
  internal/cli/mode_migrate.go:240                         # 무관 (출력 문자열)
  ```

  `_test.go`의 `REQEntry{… Widened: true}` 리터럴은 픽스처이며 실행 경로가 아니다. 위 grep이
  비테스트에서 두 번째 기록 지점을 보이는 순간 이 불변식은 깨진 것이다.
- **왜 불변식을 적어 두는가**: 나중에 두 번째 경로를 추가하는 사람이 **자기가 자문 범위를 넓히고
  있다는 사실을 그 시점에 보게 하려고** 적는다. 적어 두지 않으면 그 사실은 넓어진 뒤에야, 그것도
  누군가 error를 기대했던 finding이 warning으로 나온 것을 눈치챘을 때에만 드러난다.

**부채 기록 ② (REQ-CRS-001-003의 자문 처리가 남기는 것).** 이 처리로 두 개의 가드가 **선언만 하고
집행하지 않는 상태**가 되었다.

- **왜 지금 자문인가**: 넓힌 파서가 켜지는 순간 이 코퍼스는 한 번도 검사받은 적 없는 규칙
  846+25+6건에 노출된다. 이를 error로 착지시키면 합리적 대응이 일괄 억제(bulk suppression)가
  되고, 그것이 이 SPEC이 막으려는 결과다.
- **승격 조건**: `CoverageIncomplete` 846건이 정리되거나 예외 처리되면 `CoverageRule`의 심각도가
  `error`로 돌아간다. 넓힘 전용 finding(현재 25+6+36건)이 정리되면 `reqFindingSeverity`의
  `Widened` 분기를 삭제한다.
- **잊었을 때**: 승격 조건은 산문이고 산문은 발화하지 않는다. 승격 시점을 넘겨 자문으로 남은
  가드는 **미실행이 성공과 구분되지 않는 검사**이며, 이는 이 SPEC이 적발하려고 열린 결함과
  같은 형태다. 이 패키지에서 산문 승격 조건 위에 잠든 규칙은 t342의 `MovingRefUnpinned`가 첫
  번째이고, 이것이 **두 번째**다. 두 번째가 생겼다는 사실 자체가, 이 형태가 일회성 사고가 아니라
  재발 중인 패턴이라는 신호다.

**M3 Gaps**: ① `LegacyEARSKeyword` 7건(좁은 항목)은 종전대로 era 강등에 의존해 자문이다 — 만약
modern-era SPEC에 좁은 항목이 새로 생기면 `--strict`가 붉어질 수 있으나, 이는 M3 이전에도
참이었으므로 회귀가 아니다. ② M4(형제 `acceptance.md` 판독)는 미착수다. ③ CI 판정은 아직 없다 —
위 rc는 이 트리의 로컬 실행이며, 조용한 head에서의 CI 완주는 병합 후에 판정한다.

### 병합 트리 재측정 — 상속 대 귀속 분리 (2026-08-31)

리드가 `origin/develop`를 이 브랜치에 흡수하고 병합 트리에서 재측정했다. 아래 수치는 **리드가
이번 실행에서 직접 잰 값**이며, 이 카드가 재도출한 것이 아니다.

**병합**

| 항목 | 값 |
|---|---|
| 병합 커밋 | `d5ca9582d` (`git merge --no-edit origin/develop`, 충돌 0, 이후 `git rev-parse -q --verify MERGE_HEAD` 빈 출력) |
| 흡수된 head | `d8a1a8e4e` (t318) — t332 `1728136c7`, t336 `34cc70a90`을 포함 |
| 병합 전 브랜치 head | `523ce9d1b` |

**병합 트리 측정** (병합 트리에서 빌드, `go build -o <scratch>/moai-t362-merged ./cmd/moai` rc=0)

| 명령 | 결과 |
|---|---|
| `moai spec lint --strict --json` | **rc=1 · error 2 · warning 1,091 · 총 1,093** · 비자문 warning **0** |
| `moai spec lint` (plain) | **rc=1** · `2 error(s), 1091 warning(s)` |
| `go vet ./internal/spec/...` | rc=0 |
| `go test ./internal/spec/... -count=1` | ok 38.4s |

증거: `.moai/state/verify/t362/merge-tree-strict.json`, `.moai/state/verify/t362/merge-tree-plain.txt`.

**귀속 분리 — 이 문단의 요점**

error 2건은 **상속분이며 t336 귀속**이다. t362 귀속이 아니다.

```
ArtifactStatusFieldForbidden  .moai/specs/SPEC-INTEGRATION-LOCK-ATOMIC-001/plan.md:5
ArtifactStatusFieldForbidden  .moai/specs/SPEC-INTEGRATION-LOCK-ATOMIC-001/acceptance.md:5
```

두 건은 리드가 병합 **전** `d8a1a8e4e`에서 읽은 두 건과 **파일·행 단위로 일치**한다. 그리고
`ArtifactStatusFieldForbiddenRule`은 `doc.REQs`를 읽지 않으므로, **이 카드의 어떤 변경으로도
이 finding을 만들거나 억누를 수 없다.**

**두 방향을 가르는 검사는 이것이다 — `doc.REQs` 소비 4개 코드가 병합을 가로질러 불변인가.**

| 코드 | 병합 전 (`523ce9d1b`) | 병합 후 (`d5ca9582d`) |
|---|---|---|
| `CoverageIncomplete` | 846 | **846** |
| `ModalityMalformed` | 25 | **25** |
| `InvalidREQID` | 6 | **6** |
| `DuplicateREQID` | 0 | **0** |

네 코드가 전부 불변이므로 **흡수는 이 카드의 파급에 아무것도 더하지 않았다.** 총 +3의 델타는
전부 다른 곳이다 — `MovingRefUnpinned` 113→114, `ArtifactStatusFieldForbidden` 0→2, 둘 다
develop이 가져온 새 SPEC 디렉터리와 함께 들어왔다.

**따라서 t362 귀속 error = 0.** 이 카드 자신의 기여는 여전히 error 0 · 비자문 warning 0이고,
`--strict`가 1로 끝나는 것은 **오직 상속된 두 건 때문**이다.

**분리는 양방향으로 명시한다.** 상속된 적색을 이 카드에 귀속시키는 것과, 진짜 이 카드의 적색을
"상속분"이라며 미루는 것은 **같은 결함**이다. 위 4-코드 불변 검사가 그 둘을 가르는 근거이며,
"관련 없어 보인다"는 인상이 아니다. 만약 네 코드 중 하나라도 병합을 가로질러 움직였다면 그
증분은 이 카드가 해명해야 할 몫이다.

**격리 측정은 존재한다 — 리포지터리의 `SPEC Lint` CI 워크플로가 그것이다.**

앞선 판본은 이 귀속이 "리드의 판독에 근거하며 격리 측정은 없다"고 적었다. **그 서술은 실재하는
증거를 과소평가한 것이었다.** `SPEC Lint`는 `ci.yml`과 **별개의 워크플로**이며, 체크아웃한
소스에서 빌드하므로 로컬 `moai spec lint`를 괴롭히는 **설치본 지연(binary lag)에 면역**이다.

`gh run list --branch develop --workflow "SPEC Lint"`로 직접 읽은 판정:

| 커밋 | 카드 | SPEC Lint |
|---|---|---|
| `1728136c7` | t332 | **success** |
| `34cc70a90` | t336 | **failure** |
| `d8a1a8e4e` | t318 | **failure** |

**회귀 지점은 `34cc70a90`이고, 세 트리 중 어느 것도 이 카드의 브랜치를 담고 있지 않다.**
실패 로그도 직접 읽었다(`gh run view 33318836580 --log-failed`): `2 error(s), 159 warning(s)`,
`ArtifactStatusFieldForbidden`이 `SPEC-INTEGRATION-LOCK-ATOMIC-001/plan.md:5`와
`acceptance.md:5` — **우리 병합 트리가 보고하는 바로 그 두 줄이다.**

즉 격리 측정은 **소스 빌드 위에서, 이 카드의 코드가 없는 트리에서** 이미 이루어졌고, 이제
전언이 아니라 직접 판독이다.

**남는 gap은 이것뿐이며 이보다 넓지 않다**: `d8a1a8e4e`만으로 **로컬에서** develop 전용
바이너리를 빌드하지는 않았다. 대신 CI의 소스 빌드를 소비했다 — 그것은 로컬 빌드의 **대체재가
아니라 더 강한 증거**다(깨끗한 환경 + 소스 빌드 + 이 카드 코드 부재).

**되풀이하지 않기 위해 기록하는 판독 오류**: 처음에는 `ci.yml` 실행 **안에서** "SPEC Lint" job을
찾았고, 없다는 것을 근거로 리드의 주장을 잠시 약한 것으로 취급했다. 그것은 **별개의 워크플로**다.
**엉뚱한 곳에서의 부재는 부재가 아니다.**

**부채 기록 ③ — §2.3 미러는 정당화가 죽었는데도 남아 있다.**

리드는 미러 삭제를 **M4에** 비준했다. 지금이 아니다. 그래서 부채로 기록한다.

- **원래 정당화는 더 이상 성립하지 않는다.** 미러의 존재 이유는 "삭제하면 develop이 붉어진다"
  였고, 그 이유는 **option A 채택과 함께 죽었다.** 측정으로 확인했다 —
  `.moai/reports/t362/c2-nomirror-strict.txt`: 미러를 뺀 사본은 **8건, 전부 WARNING, error 0,
  rc=0**(plain·`--strict` 양쪽).
- **그런데도 남기는 이유는 하나뿐이다**: 지금 지우면 **영구 자문 warning 8건**이 남고, M4 이후에
  지우면 **0건**이 남는다. 그것이 유일한 근거이며, 원래 정당화와는 다른 근거다.
- **따라서 읽는 사람은 미러가 계속 있다는 사실을 "아직 정당하다"는 증거로 읽어서는 안 된다.**
  정당화는 이미 죽었고, 남아 있는 것은 삭제 시점 선택의 결과일 뿐이다.

**M4 재측정 의무 — 8→0은 관측이 아니라 예측이다.**

`c2-nomirror-strict.txt`는 **M4 이전 트리**에서 쟀다. 따라서 "미러를 지우면 8건이 0건이 된다"는
것은 **아직 예측**이다. M4에 다음을 의무로 건다:

- 미러를 삭제한 뒤 **재측정하고, 8건이 실제로 0건이 되었는지 기록한다.**
- 그 수치는 **측정한 트리 SHA에 핀**한다.
- **0이 되지 않으면 그것은 장부 정리 문제가 아니라 M4에 관한 발견이다** — 왜 남았는지가 M4가
  답해야 할 질문이 된다.

**범위 밖**: `SPEC-INTEGRATION-LOCK-ATOMIC-001/`은 다른 카드 소관이며 이 카드는 건드리지
않았다. 리드가 운영자에게 별도로 라우팅한다.

### M4 — `CoverageRule`가 형제 `acceptance.md`를 읽는다

**측정 기준 트리**: 워크트리 `.claude/worktrees/t362`, 브랜치 `WT-coverage-rule-scope`,
기준 HEAD `706e2ae4e`(이 절의 모든 수치는 그 HEAD 위의 작업 트리에서 잰 것이고, 바이너리도
같은 트리에서 빌드했다 — `go build -o /tmp/moai-t362-m4-{base,after} ./cmd/moai`, 양쪽 rc=0).

**수리 모양**: plan.md §C.2 (ii)안. `CoverageRule.Check`가 `filepath.Dir(doc.Path)`의 형제
`acceptance.md`를 직접 읽어 커버 집합을 합집합한다. `parseSPECDoc`은 건드리지 않았으므로
`doc.Criteria` 소비자는 영향을 받지 않는다. 신규 파일 `internal/spec/lint_coverage_sibling.go`,
`lint.go` 변경은 3줄 + 주석 4줄.

**커버 집합 술어**: 형제 파일 전문에 `ExtractRequirementMappings`를 적용한다. 인라인 경로
(`ParseAcceptanceCriteria`)의 두 좁힘 — `##`+"acceptance" 표제 요구와 `AC-…:` 콜론 줄 요구 —
는 **spec.md가 혼합 문서라서** 존재하는 것이고 `acceptance.md`에는 근거가 넘어오지 않는다.
실측이 이를 강제했다: `acceptance.md` 622개 중 표제 조건 충족 169개, 인라인 파싱 가능한
`- AC-…:` 줄 보유 **3개**. 이 SPEC 자신의 `acceptance.md`도 3개에 들지 않는다 — 인라인 술어를
그대로 가져왔다면 **자기를 낳은 문서를 커버하지 못하는 수리**가 됐다.

**AC 판정**

| AC | 결과 | 근거 |
|---|---|---|
| AC-CRS-001-006a | PASS | `TestCoverageSibling_CoveredByAcceptanceMD` — 형제에만 AC가 있는 픽스처에서 `CoverageIncomplete` 0건 |
| AC-CRS-001-006b | PASS | `TestCoverageSibling_UncoveredStillFires` — 양쪽 어디에도 AC가 없는 REQ에 대해 여전히 1건 발화. **rc가 아니라 발화 여부로 판정**(개정 문구) |
| AC-CRS-001-007 | PASS | `TestCoverageSibling_NoAcceptanceArtifact` — `acceptance.md` 부재 Tier S 픽스처, 오류·패닉 없이 인라인 AC만으로 판정 |
| AC-CRS-001-008 | PASS | 인라인 집합과 형제 집합의 **합집합**이므로 spec.md 중복 기재를 새로 요구하지 않는다. `spec.md` 코퍼스 무변경 |

**RED 선행 확립** (`.moai/reports/t362/m4-red-before.txt`, 구현 전):

```
--- FAIL: TestCoverageSibling_CoveredByAcceptanceMD
    CoverageIncomplete findings = 1, want 0
--- FAIL: TestCoverageSibling_UncoveredStillFires
    CoverageIncomplete findings = 2, want 1
--- PASS: TestCoverageSibling_NoAcceptanceArtifact
```

007이 구현 전에도 GREEN인 것은 **회귀 방지 기준**이기 때문이다(부재 경로가 깨지지 않았는지를
본다). 공허하지 않다는 것은 아래 뮤테이션 M2·M3가 보인다.

**뮤테이션 3종 — 쌍이 실제로 갈라 세는지**

| 뮤턴트 | 006a | 006b | 007 | 증거 |
|---|---|---|---|---|
| M1 형제 무시(합집합 제거) | **RED** (0→1) | **RED** (1→2) | PASS | `m4-mutant-ignore-sibling.txt` |
| M2 규칙 끄기(`return nil`) | **PASS(공허)** | **RED** (1→0) | **RED** (1→0) | `m4-mutant-rule-off.txt` |
| M3 부재를 하드 실패로 | — | — | **RED**(panic) | `m4-mutant-absent-hard-fail.txt` |

**M2가 쌍의 존재 이유다.** 규칙을 꺼도 006a는 통과한다 — 앞쪽만 보면 "올바로 읽는다"와
"규칙을 껐다"가 구분되지 않는다. 그 구분을 만드는 것은 006b뿐이다. M3은 007의 픽스처가 실제로
부재 분기를 지난다는 것을 보인다(부재 파일 경로가 panic 메시지에 그대로 찍혔다).

**코퍼스 실측 — 델타 0이며, 그것은 코퍼스가 정한 값이다**

| 측정 | 값 |
|---|---|
| `CoverageIncomplete` (M4 전 / 후) | **846 / 846 — 델타 0** |
| finding 총수 (전 / 후) | 1,093 / 1,093 |
| 비자문 warning | 0 / 0 |
| error | 2 / 2 (상속분, 아래) |
| plain rc / `--strict` rc | 1 / 1 (양쪽 모두 상속 error 2건 때문) |

`jq -S` 정규화 후 before/after finding 배열이 **바이트 동일**(diff exit 0).

**델타 0의 원인을 재봤다 — 구현이 아니라 코퍼스다.** 846건은 47개 SPEC에 얹혀 있다. 그중 43개는
`acceptance.md`를 **가지고 있고**, 그 43개 중 `maps REQ-` 매핑을 선언한 것은 **0개**다(23개는
AC id만 적고 매핑을 적지 않으며, 나머지는 둘 다 없다). 반대로 `maps REQ-`를 가진
`acceptance.md` 14개는 전부 finding 0건인 SPEC의 것이다. **두 모집단이 서로소이므로 코퍼스
수치는 구현이 무엇을 하든 움직일 수 없다.** 근거:
`.moai/reports/t362/m4-population-measurement.txt`, `m4-residual-measurement.txt`.

**여기서 더 넓히지 않은 이유**: `acceptance.md`의 맨 `REQ-…` 토큰을 커버로 세면 846건 대부분이
조용히 사라진다. 그것은 M3이 자문 부채로 **일부러 드러낸** 수치를 지우는 일이고, "이 REQ는 범위
밖"이라고 적힌 산문까지 커버로 세게 된다. 매핑 선언이 곧 커버 선언이며, 선언이 없는 문서가
미커버로 읽히는 것은 규칙이 **작동하는** 모습이다. 잔여는 코퍼스 작성 방식의 문제이지 이 술어의
좁음이 아니다.

**모든 코퍼스 수치는 로컬 측정이며, CI보다 체계적으로 19 높다.**

이 §E.2에 적힌 warning 수치는 **전부 local**이다. CI의 `SPEC Lint` 워크플로는
`actions/checkout@v7`를 `fetch-depth` 없이 쓰므로 히스토리 없는 depth-1 셸로 클론을 받고,
`StatusGitConsistencyRule`과 `OwnershipTransitionRule`은 둘 다 `git log --follow`로 판정하므로
CI에서는 **볼 것이 없어 아무것도 내지 않는다**. 그래서 같은 트리에서도

- local = CI + `StatusGitConsistency` 18 + `OwnershipTransitionInvalid` 1 = **CI + 19**

이 성립한다. 리드가 CI 로그에서 잰 값(트리 `1e5199b88`): warning 1,072 + error 2.
이 카드가 로컬에서 잰 값(트리 `9610e013e`): warning 1,091 + error 2. 차이는 정확히 19다.

**따라서 CI에서 1,091을 찾다가 1,072를 보고 회귀로 읽어서는 안 된다.** 둘은 같은 측정이 아니며
출처 없이 비교해서는 안 된다. CI 체크아웃 자체는 **다른 카드 소관**이고 이 카드는 건드리지
않았다.

**846 델타는 이 오프셋에 오염되지 않는다.** `CoverageIncomplete`는 git 히스토리에 의존하지
않으므로 CI와 로컬이 **동일하게 846**을 읽는다(리드의 CI 로그 판독: CI `1e5199b88`에서 846).
오프셋을 만드는 두 규칙은 `StatusGitConsistency`와 `OwnershipTransition` 둘뿐이다.

**flag 축은 무해함이 이 트리에서도 재확인됐다.** 같은 바이너리·같은 트리에서 plain과
`--strict`의 JSON을 각각 뽑아 `cmp` 했다 — 세 쌍 모두 **바이트 동일**:

| 비교 | 명령 | 결과 |
|---|---|---|
| M4 이전 plain vs `--strict` | `/tmp/moai-t362-m4-base spec lint [--strict] --json` | `cmp` rc=0 |
| M4 이후 plain vs `--strict` | `/tmp/moai-t362-m4-committed spec lint [--strict] --json` | `cmp` rc=0 |
| M4 이전 `--strict` vs 이후 `--strict` | 위 두 바이너리 | `cmp` rc=0 → **델타 0** |

코드가 예측하는 그대로다: `Strict`는 `HasErrors()` 안에서만 소비되며 finding 방출 경로에
닿지 않는다(`lint.go:56-66`). 다른 레인이 보고한 178 vs 159 불일치는 **이 트리에서 재현되지
않는다.**

**라벨을 붙인 최종 수치** (전부 `local`, 트리 `9610e013e`, `spec.md` 분모 710 —
`glob .moai/specs/SPEC-*/spec.md`로 이 트리에서 잰 값이며, 계획 단계의 704는 더 이른 트리
`68ecbfe4a`의 값이다):

| 수치 | 값 |
|---|---|
| `CoverageIncomplete` (local, `spec lint --strict --json`, `9610e013e`, 분모 710) | **846** |
| 같은 값, M4 **이전** (local, `spec lint --strict --json`, `706e2ae4e`, 분모 710) | **846 → 델타 0** |
| finding 총수 (local, `--strict --json`, `9610e013e`) | 1,093 |
| warning / error 분해 (local, plain 사람용 출력, `9610e013e`) | `2 error(s), 1091 warning(s)` |
| 비자문 warning (local, `--strict --json`, `9610e013e`) | 0 |
| plain rc / `--strict` rc (local, `9610e013e`) | 1 / 1 — 양쪽 모두 상속 error 2건 때문 |
| CI 환산값 (위 오프셋 적용, **미관측 — 산출값**) | warning 1,072 + error 2 |

마지막 행은 **관측이 아니라 산출**이다. 이 커밋은 아직 푸시되지 않았으므로 CI가 이 트리를
본 적이 없다.

**살아있는 문서에서의 증거 — §2.3 미러 제거 시뮬레이션 (8 → 0)**

`.moai/reports/t362/m4-mirror-removal-sim-{before,after}.json`. 코퍼스는 건드리지 않았다:
이 SPEC의 `spec.md`·`acceptance.md`를 `/tmp/m4sim/`으로 복사한 뒤 §2.3 미러 표 13줄만 지우고
단일 경로로 lint 했다.

- M4 **이전** 바이너리: `CoverageIncomplete` **8건**(REQ-CRS-001-001..008)
- M4 **이후** 바이너리: **0건**

§E.2가 M4에 걸어둔 재측정 의무는 이로써 이행됐다. 다만 **이것은 스크래치 사본 측정**이며,
실제 트리의 미러 삭제는 manager-spec 소관으로 이 커밋 **뒤에** 순서 지어져 있다 — 실트리
8→0 확인은 그 삭제 뒤 오케스트레이터가 재측정한다.

**상속 error 2건은 그대로 2건**: `SPEC-INTEGRATION-LOCK-ATOMIC-001`의 `plan.md`·`acceptance.md`
`ArtifactStatusFieldForbidden`. 다른 카드 소관이고 이 카드는 건드리지 않았다.

**품질 게이트**: `go vet ./internal/spec/...` rc=0 · `go test ./internal/spec/...` rc=0(38.2s) ·
`golangci-lint run ./internal/spec/...` **0 issues**. 전체 스위트는 로컬에서 돌리지 않았다(CI 몫).

**미검증으로 남는 것(Gap)**
- windows/linux 빌드 — 미측정. darwin `go build ./cmd/moai` rc=0만 관측.
- 실트리 미러 삭제 후의 8→0 — 위 시뮬레이션은 스크래치 사본이다. 실트리 확인은 미이행.
- 매핑을 선언하지 않는 43개 `acceptance.md`의 커버리지 — 이 카드가 닫지 않는다. 코퍼스 작성
  문제이며 별도 카드 소관이다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready           # M1-M4 완료
run_complete_at: 2026-08-31
run_commit_sha: pending-backfill-m4
ac_pass_count: 9                  # AC-CRS-001-001..005 + 006a + 006b + 007 + 008
ac_fail_count: 0
ac_deferred_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: pending
l44_post_push_fetch: pending
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues
cross_platform_build:
  darwin: pass                        # go build ./cmd/moai rc=0
  windows: not-measured
total_run_phase_files: 10             # lint.go, lint_req_widen.go, lint_coverage_sibling.go
                                      # + 4 test files + 3 M4 fixture dirs
m1_to_mN_commit_strategy: milestone-per-commit
open_amendments: 1                    # C2 (§2.3 미러 삭제) — manager-spec 소관, 이 커밋 뒤 순서
corpus_coverage_incomplete_delta: 0   # 846 -> 846 (local, spec lint --strict --json,
                                      # 706e2ae4e -> 9610e013e, spec.md 분모 710)
                                      # 모집단 서로소 — m4-population-measurement.txt
corpus_figures_source: local          # CI는 shallow checkout이라 warning이 19 낮다(§E.2)
plain_vs_strict_byte_identical: true  # cmp rc=0, 3쌍 모두
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-31
sync_commit_sha: pending-backfill-sync   # 커밋은 자기 해시를 알 수 없다 — 직후 커밋에서 백필
sync_status: complete
b12_self_test_a: pass          # grep -c 'SPEC-COVERAGE-RULE-SCOPE-001' CHANGELOG.md -> 0 (추가 전)
b12_self_test_b: pass          # acceptance.md 고유 AC id 9건 = CHANGELOG 인용 9건
b12_self_test_c: pass          # 인용한 구현 경로 9개 전부 ls 확인
changelog_entry_position: "[Unreleased] ### Changed 최상단 (line 168)"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed"   # 3-phase close, 이 sync 커밋 1개에 병합
  plan_md: none          # status 필드 없음 — ArtifactStatusFieldForbidden 규칙이 금지
  acceptance_md: none    # 위와 같음
  progress_md: none      # 위와 같음
canary_compliance_check:
  applicable: false      # 이 SPEC은 전방위 정책을 정의하지 않는다
docs_surface_review:
  readme: no-change      # 'spec lint' 미언급
  docs_site: no-change   # content/*/cli-reference/spec.md 는 플래그만 문서화, 룰/코드/심각도 미열거
mx_tag_validation: no-change   # 이 SPEC은 @MX 주석을 추가·변경하지 않았다
```

**실트리 8 → 0 — §E.2가 미이행으로 남긴 gap을 sync에서 닫았다**

§E.2의 미러 제거 시뮬레이션은 `/tmp/m4sim/` 스크래치 사본 측정이었고, "실트리 확인은
미이행"으로 기록돼 있었다. §2.3 미러 표가 `c4a0c967d`에서 실제로 삭제된 뒤, **같은 실트리
파일**에 두 바이너리를 걸어 다시 쟀다.

| 바이너리 | 트리 | `CoverageIncomplete` | rc |
|---|---|---|---|
| `/tmp/moai-t362-m4-base` (M4 이전, run-phase 산물 — 출처는 §E.2에서 승계) | `c4a0c967d` 작업트리 | **8** (REQ-CRS-001-001..008) | 0 |
| `/tmp/moai-t362-sync` (`go build ./cmd/moai` rc=0, 이 트리에서 새로 빌드) | `c4a0c967d` 작업트리 | **0** | 0 |

명령: `<binary> spec lint --json .moai/specs/SPEC-COVERAGE-RULE-SCOPE-001/spec.md`.
양쪽 rc가 모두 0인 것은 A안(자문 등급)의 예상된 결과이며, 판정은 발화 여부로 한다.
`after` 쪽 바이너리만 이 트리에서 빌드했고, `before` 쪽은 run-phase가 남긴 산물이라
**그 출처는 승계값**이다 — 이 절이 새로 관측한 것은 두 값이 갈린다는 사실이다.

**sync 단계에서 새로 재지 않은 것(Gap)**
- 전 코퍼스 수치(846 / 1,093 / 25 / 6) — §E.2의 run-phase 측정을 승계했고 sync에서 재측정하지
  않았다. 승계값임을 CHANGELOG와 여기 양쪽에 명시했다.
- windows/linux 빌드 — 여전히 미측정.
- CI 판정 — 이 브랜치는 미푸시이므로 CI가 이 트리를 본 적이 없다. 통합은 리드 소관.
- 독립 sync-audit — 이 카드는 수행하지 않았다.
