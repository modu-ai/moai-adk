# 개정 기록 — AC-SLB-003 계기 수리 (카드 t518, lane-4)

## 승인 주체

**이 개정은 리드가 승인했다.** 저자(lane-4)의 자기 승인이 아니다.

옛 판독이 **실패**했으므로(아래) 「양쪽 판독에서 충족」 안전 조건이 성립하지 않고,
그래서 저자가 아니라 리드가 근거를 적어 승인하는 경로를 탔다. t538 AC-011 과 같은
판별식 — **계기 결함이지 작업 결함이 아니다.**

## 옛 판독은 실패했다 (숨기지 않는다)

병합 트리 `27317a6dd` 에서 `AC-SLB-003` 의 기구가 **FAIL** 했다.

```
--- FAIL: TestTableCollection_CorpusListFindingsUnchanged (0.27s)
  code CoverageIncomplete: narrow-path findings = 16,
  live non-advisory findings = 0 (scanned 812 docs)
      — the table branch moved a list-form finding
```

## 실패 원인 — 두 피연산자가 같은 양을 재지 않는다

`27317a6dd:internal/spec/lint.go` 의 **`CoverageRule.Check`** 는 그 함수 본문 안에서
`Advisory: true` 를 **무조건** 한 번 설정한다(주석: `reports, never gates`). 조건부가
아니다. (줄번호는 낡으므로 함수명으로 지목한다.)

`27317a6dd:internal/spec/lint_req_table_test.go` 의 두 계수는 비대칭이었다:

| 피연산자 | 집계 방식 | `CoverageIncomplete` 에 대해 |
|---|---|---|
| narrow | `narrowCounts[f.Code]++` — **필터 없음** | findings 가 생기는 즉시 > 0 |
| live | `if !f.Advisory { … }` — 자문 제외 | **구조적으로 항상 0** |

따라서 이 코드에서 등식은 원리상 성립할 수 없다.

## 병합 전 통과는 우연한 0 == 0 이었다

병합 전까지 등식이 성립한 것은 narrow 측이 코퍼스 전체에서 `LegacyEARSKeyword` 7건만
냈기 때문이다. 그 제한은 은폐된 적이 없다 — `progress.md` M-A1 절에 이미
「narrow 정규식이 거의 매치하지 않는다」로 적혀 있었다. 즉 **단언이 성립한 것이 아니라
피연산자 한쪽이 0이었을 뿐이다.**

## t528 코드와의 의미적 충돌이 아니다

프로브(측정 후 삭제)로 narrow 측 `CoverageIncomplete` 16건의 출처를 전수 열거했다:

```
PROBE 16 .moai/specs/SPEC-AC-COLLECTOR-ANCHOR-001/spec.md
(그 외 문서 0건)

git cat-file -e cc63fa54f:.moai/specs/SPEC-AC-COLLECTOR-ANCHOR-001/spec.md
  → fatal: exists on disk, but not in 'cc63fa54f'
```

t528 이 develop 으로 들여온 **자기 SPEC 문서 1개**가 코퍼스에 들어와 narrow 계수를
0 → 16 으로 움직였다. `parseSingleACLine` 확장(t528 의 코드 변경)과는 무관하다.

## 수리

`lint_req_table_test.go` 의 narrow 집계에도 `!f.Advisory` 를 적용해 두 피연산자를
like-for-like 로 맞췄다. 프로덕션 코드는 건드리지 않았다(테스트 1파일).

## 비공허성 — 뮤턴트로 RED 를 관측했다

수리 후 세 뮤턴트를 `internal/spec/lint_req_widen.go`
`parseREQsWithProvenance` 에 순차 주입하고 원본 복원을 해시로 확인했다
(baseline `747106f5fedb8fa28a057a23827c55e7c05a20576e74f6eb9ff3d015c947e0a8`,
복원 후 동일).

| 뮤턴트 | 주입 내용 | 도달 | 결과 |
|---|---|---|---|
| A | `wide = wide[1:]` — 리스트 항목 첫 개를 표 분기가 밀어냄 | 282 문서 | **미검출 (PASS)** |
| B | `if len(tbl) > 0 { wide = nil }` — 표가 있는 문서에서 리스트를 삼킴 | **3 문서 (실행됨)** | **미검출 (PASS)** |
| C | `wide = nil` 무조건 — 라이브 경로가 표 항목만 유지 | 812 문서 | **검출 (FAIL)** ✅ |

C 의 실패 메시지는 이 테스트가 스스로 적어둔 목적 문구와 일치한다:

```
code LegacyEARSKeyword: narrow-path findings = 7, live non-advisory findings = 0
  (scanned 812 docs) — the table branch moved a list-form finding
```

### 못 잡은 뮤턴트도 남긴다 — 가드의 경계

A·B 는 **미도달이 아니라 실제로 실행되고도 살아남았다**(도달성 별도 측정:
표 항목을 가진 문서 3개/20항목, 리스트 항목을 가진 문서 282개, 스캔 812).
따라서 이 가드는:

- **잡는다**: 라이브 경로가 리스트 유래 비자문 finding 을 **전면적으로** 잃는 경우
- **못 잡는다**: 항목 단위 밀어내기(A), 표 보유 문서에 국한된 손실(B)

B 가 3개 문서에서 실행되고도 통과한 이유는 그 3개 문서가 `LegacyEARSKeyword` /
`DuplicateREQID` 비자문 finding 을 내지 않기 때문이다 — 변별력의 원천이 그 문서들에
없다.

## 비공허 가드가 발동하지 않음을 관측

`lint_req_table_test.go` 의 `if len(codes) == 0` 가드는 **발동하지 않는다**.
수리 후 테스트 자신의 로그가 비어 있지 않은 맵을 찍는다:

```
scanned 812 documents; per-code narrow == live-non-advisory: map[LegacyEARSKeyword:7]
```

가드가 있다는 사실과 그것이 지금 무엇을 허용하는지는 다른 이야기이므로,
가드의 존재가 아니라 **비발동을 관측**해 적는다.

## 잔존 변별력 — 세었다

```
SCANNED 812
NARROW_ALL(unfiltered, 옛 피연산자) map[CoverageIncomplete:16 LegacyEARSKeyword:7]
NARROW_NONADVISORY(수리된 피연산자)  map[LegacyEARSKeyword:7]
LIVE_NONADVISORY                     map[LegacyEARSKeyword:7]
DUPLICATE_REQID  narrow=0 live=0
LEGACY_EARS      narrow=7 live=7
```

`DuplicateREQID` 0/0 은 **주장이 아니라 측정**이다(위 프로브 출력). 이 코드의 변별력은
순서 의존 사례가 코퍼스에 등장할 때 발현되는 잠재 변별력이며 오늘은 0건이다.

## 수리 후 재측정 (병합 트리 `27317a6dd` + 수리)

```
go vet ./internal/spec/... ./internal/cli/... ./internal/kanban/...
  → exit 0, 출력 0줄
gofmt -l internal/spec/lint_req_table_test.go
  → 무출력

go test ./internal/spec/... ./internal/kanban/... -count=1   → exit 0
  ok internal/spec    72.688s
  ok internal/kanban 142.460s

go test ./internal/cli/... -count=1 -timeout 900s            → exit 0
  ok internal/cli    461.510s   (+ 하위 16개 패키지 전부 ok)
```

## 잔여 미검증

- CI 판정 없음 (미푸시). 로컬 초록은 조기 신호이며 darwin/windows 매트릭스가 아니다
- A·B 뮤턴트가 그린 사각지대(항목 단위 밀어내기·표 문서 국한 손실)는 이 카드에서
  닫지 않았다 — 가드를 넓히는 것은 별개 판단이다
- `CoverageIncomplete` 는 이제 이 테스트의 비교 대상에서 빠졌다. 그 코드 자체의
  회귀는 이 테스트가 아니라 `CoverageRule` 자신의 테스트가 지킨다
