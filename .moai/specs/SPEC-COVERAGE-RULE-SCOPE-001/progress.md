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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
