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

**오독(misread) 건수 = 0.** 세 가지 기계적 술어로 쟀다. P2(캡처된 본문이 빈 줄) 0건,
P3(가장 가까운 상위 표제가 교차참조·Gaps·이력 등 요구사항 절이 아님) 0건.
P1(캡처 본문에 다른 REQ 토큰이 등장 — 매핑 행 후보) 22건이 걸렸으나, **22건을 전수 열람한
결과 모두 다른 REQ를 본문에서 인용하는 진짜 정의 줄이었다.** 따라서 P1은 위양성률 100%인
과대 술어이고, 교정된 판정으로는 오독 0건이다. 넓힌 추출이 정의 아닌 줄을 정의로 읽는 사례는
코퍼스에 없다. (술어별 전수 목록은 보고서 `[D]` 절.)

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

`DuplicateREQIDRule`은 공허하지 않으며 계속 동작한다 —
`TestDuplicateREQIDRule_StillFiresOnWidenedShapes`가 넓어진 형태(`REQ-HOOK-001`)에 대해
중복 탐지가 실제로 발화함을 강제한다(검증을 넓히기 전에는 `continue`에 걸려 0건이었다).

| AC | 상태 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-CRS-001-003 | PASS | `moai spec lint --json` (M2 바이너리, 전 코퍼스) | `InvalidREQID` 0건 / 총 177건 / rc=0 · `--strict` rc=0 — 병합 전 baseline과 코드별 건수 동일 |
| AC-CRS-001-004 | PASS | `go test ./internal/spec/ -run 'TestReqIDPattern\|TestDuplicateREQIDRule_StillFires' -v` | RED 증거 `.moai/reports/t362/m2-red-evidence.txt`, 이후 4/4 PASS |

**M2 단독 착지 시 코퍼스 무변화**: `moai spec lint --json` 177건, 코드별 내역
MovingRefUnpinned 113 / MissingExclusions 24 / StatusGitConsistency 18 / FrontmatterInvalid 14 /
LegacyEARSKeyword 7 / OwnershipTransitionInvalid 1 — baseline과 동일. `doc.REQs`가 아직 좁은
집합이므로 검증을 넓힌 것만으로는 관측 가능한 변화가 없다(설계대로).

### M3 — 넓힌 파서 배선 + 심각도 처리 (2026-08-30)

측정 트리: HEAD `0d102d7c7`. 바이너리는 이 트리에서 빌드(`go build -o <scratch>/moai-m3 ./cmd/moai`, rc=0).

**배선 전에 드러난 사실: `CoverageRule`은 `doc.REQs`의 유일한 소비자가 아니다.**
Gate 0의 `[F]` 절에서 전 소비자를 좁힘/넓힘으로 시뮬레이션했다:

| 코드 | 심각도 | 좁힘 | 넓힘 |
|---|---|---|---|
| `CoverageIncomplete` | error | 0 | **846** |
| `ModalityMalformed` | error | 0 | **25** |
| `InvalidREQID` (M2 후 패턴) | error | 0 | **6** |
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
| AC-CRS-001-005 | PASS | `moai spec lint --json` / `moai spec lint --strict --json` | **plain rc=0, strict rc=0.** 심각도 분포: warning 1,090 / error **0**. 비자문 warning **0**건 |

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

**부채 기록 (REQ-CRS-001-003의 자문 처리가 남기는 것).** 이 처리로 두 개의 가드가 **선언만 하고
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

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready-partial   # M2 + M3 완료, M4 미착수
run_complete_at: 2026-08-30
run_commit_sha: pending-backfill-m2-m3
ac_pass_count: 5                  # AC-CRS-001-001..005
ac_fail_count: 0
ac_deferred_count: 3              # AC-CRS-001-006a/006b/007 (M4 소관)
preserve_list_post_run_count: 0
l44_pre_commit_fetch: pending
l44_post_push_fetch: pending
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues
cross_platform_build:
  darwin: pass                        # go build ./cmd/moai rc=0
  windows: not-measured
total_run_phase_files: 5              # lint.go, lint_req_widen.go + 3 test files
m1_to_mN_commit_strategy: milestone-per-commit
open_amendments: 2                    # C1 (AC-CRS-001-006b 기계 서술), C2 (§2.3 미러) — manager-spec 소관
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
