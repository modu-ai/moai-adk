# plan.md — SPEC-DOCTOR-PLUGIN-DIGEST-PREFILTER-001

## §A Context

카드 t963. 진단은 t962 가 끝냈다(`.moai/reports/t962/verdict.md` — 5절 형식). 이 카드는 **진단하지
않는다**: 처방이 배차 시점에 확정돼 있고, 계획은 그것을 최소 변경으로 착지시키는 것이다.

대상: `internal/cli/doctor_disk.go` 의 `findPluginHashClusters`(`:298`) 와 그것이 부르는
`treeDigest`(`:326`).

## §B Known Issues

- **픽스처 헬퍼가 희소 파일을 만든다** (`doctor_disk_test.go:37-40`). 큰 바이트 픽스처는 읽기 비용을
  만들지 않으므로, 시간 기반 검증은 이 저장소에서 구조적으로 공허하다. 이것이 수락을 후보 집합으로
  진술하는 이유다(REQ-DPP-005).
- **기존 회귀 가드가 사전필터를 통과하도록 설계돼 있다** — 우연이 아니다.
  `TestPluginHashDoesNotEquateSameSizeDifferentContent` 는 두 프로파일에 모두 `Size: 4, Files: 1`
  를 준다. 사전필터를 넣어도 둘 다 후보가 되어 여전히 해시되고, 그 테스트의 양방향 단언이 그대로
  성립한다. **그 파일은 수정 대상이 아니다.**
- 이 기계의 다섯 프로파일은 `(Size, Files)` 가 전부 다르다 — 즉 착지 후 이 기계에서 해싱 후보는 0건이
  된다. 그것이 곧 「필터가 항상 빈다」는 뜻은 아니므로, 양성 대조(AC-DPP-002)가 없으면 수락이 공허해진다.

## §C Pre-flight

1. `sed -n '290,330p' internal/cli/doctor_disk.go` — 현재 `findPluginHashClusters` 본문 확인.
2. `grep -n 'profileCategoryStat' internal/cli/doctor_disk.go` — 구조체가 비교 가능한지(비교 가능한
   필드만 갖는지) 확인. 맵 키로 쓸 수 있어야 한다.
3. `go test ./internal/cli/ -run 'TestPluginHash' -count=1 -v` — 착지 전 기준선. `=== RUN` / `--- PASS`
   행을 그대로 인용한다(개수가 아니라 행으로).

## §D Constraints

- 프로덕션 변경 약 15줄. `treeDigest` 불변. 기존 테스트 파일의 기존 테스트 함수 불변.
- 전량 스위트를 로컬에서 돌리지 않는다 — `go test ./internal/cli/ -run '<지목>'` 로 재고, 전 패키지
  판정은 CI 에 맡긴다(CLAUDE.local.md §4, §6).

## §E Self-Verification

- 사전필터가 실제로 걸리는지를 **후보 함수 단위 테스트**로 직접 재고, 해싱 경로를 계측하지 않는다
  (REQ-DPP-003 의 이유).
- 공허성 배제: 음성 대조(AC-DPP-001)만 있으면 「항상 빈 슬라이스를 반환」하는 구현도 통과한다.
  양성 대조(AC-DPP-002)와 혼합(AC-DPP-003)이 그것을 막는다. **셋을 한 묶음으로 본다.**
- 변이 검사 권장: 사전필터를 `return nil` 로 바꿨을 때 AC-DPP-002/003/004 가 빨개지는지 1회 확인한 뒤
  되돌린다. 빨개지지 않으면 그 테스트는 계측기가 아니다.

## §F Milestones

우선 결정 가능성이 큰 것부터 — 아래 M1 의 시그니처와 반환 계약이 이 카드에서 가장 바뀌기 쉬운 결정이다.

### M1 (Priority High) — 후보 함수 신설

`internal/cli/doctor_disk.go` 에 순수 함수를 추가한다. 제안 시그니처(구현자가 다듬어도 된다):

```go
// pluginDigestCandidates returns the profiles whose plugins tree still needs a
// content digest: byte-identical trees necessarily share a (Size, Files) pair,
// so a profile alone in its group can never be part of a cluster.
func pluginDigestCandidates(perProfile map[string]map[string]profileCategoryStat) []string
```

계약: `plugins` 가 없거나 `Files == 0` 인 프로파일은 제외(현행 `findPluginHashClusters` 의 skip 조건과
동일), 같은 `(Size, Files)` 그룹의 크기가 2 이상인 프로파일만 반환, 결과는 정렬.

### M2 (Priority High) — 배선

`findPluginHashClusters` 의 순회를 후보 집합 위에서 돌도록 바꾼다. 해시·클러스터 확인·출력 정렬은
그대로 둔다. 진입 skip 조건은 후보 함수가 이미 적용하므로 중복되지 않게 정리한다.

### M3 (Priority High) — 테스트

`doctor_disk_test.go` 에 **추가만** 한다: AC-DPP-001(음성) · AC-DPP-002(양성) · AC-DPP-003(혼합+정렬).
AC-DPP-004 는 기존 테스트를 수정 없이 통과시키는 것으로 충족되므로 새 테스트를 쓰지 않는다.
벽시계 단언 금지 — `time.Since` / 경과시간 비교를 쓰지 않는다.

### M4 (Priority Medium) — 보고

완료 보고에 REQ-DPP-005 의 두 문장을 넣는다: 바이트 비례 항 제거 · 파일 개수 비례 훑기 항 잔존
(이 기계 1회 1.37초, t962 실측 귀속). **상수 시간이 됐다고 쓰지 않는다.** 범위 밖 4건을 후속 카드
후보로 명시한다.

## §G Anti-Patterns

- **시간으로 증명하려는 유혹** — 희소 픽스처 때문에 구조적으로 공허하다(§B).
- **기존 회귀 가드를 「이제 필터에 걸리니 고쳐야 한다」고 읽는 것** — 걸리지 않는다. 두 프로파일이 같은
  `(Size, Files)` 라서 둘 다 후보다. 그 파일을 열어 고치면 그 자체가 결함이다.
- **범위 확장** — 중복 훑기 통합·`--check` 경로·go/ast 가드·init 실패는 전부 별건이다(spec.md §5).
- **후보 집합이 비는 것을 성공 신호로 읽는 것** — 이 기계에서 0건이 나오는 것은 데이터 상태이지 구현
  정확성의 근거가 아니다.

## §H Cross-References

- `.moai/reports/t962/verdict.md` — 진단 원본(계열 A, E3/E4, 처방 후보 4건, 마찰 주의).
- `internal/cli/doctor_disk.go:298,326,356` — 대상 코드.
- `internal/cli/doctor_disk_test.go:37-40,138-169` — 희소 픽스처 헬퍼 · 회귀 가드.
- SPEC-CI-DOCTOR-BIN-001 — 같은 doctor 표면의 선행 SPEC(범위 겹침 없음).
