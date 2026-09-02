# t439 — internal/hook -race 레이스 수리 판정서

card: t439
branch: WT-hook-race
base: origin/develop `f7cabfc29`
측정 트리: `f7cabfc29` (RED) / 수리 커밋 (GREEN)
플랫폼: darwin/arm64, go1.26.4

---

## Claim

`internal/hook` 의 `-race` 실패는 **데이터 레이스 1건**이며, 그 1건이 CI 의 `--- FAIL`
대량 발생을 만든다. 원인은 패키지 변수 `agentMemoryPrimaryRootFn`(`internal/hook/agentmemory.go:46`)
을 `t.Parallel()` 테스트들이 교대로 쓰는 것이다. 수리는 병렬성 제거이며, 근거는
같은 패키지에 이미 확립된 선례다.

## Evidence

### RED — 수리 전 (`f7cabfc29`)

```
go test -race -count=1 ./internal/hook/
```
- exit code: `1`
- `WARNING: DATA RACE` 건수: **1**
- `--- FAIL` 건수: **62**
- 로그: `.moai/reports/t439/race-red-baseline.log`

레이스 스택 (같은 로그에서 발췌):

```
WARNING: DATA RACE
Read at 0x000103d0a5e0 by goroutine 582:
  internal/hook.swapMirrorPrimaryRoot()
      internal/hook/memory_mirror_test.go:19
  internal/hook.TestMirrorNoOpInPrimarySession()
      internal/hook/memory_mirror_test.go:92

Previous write at 0x000103d0a5e0 by goroutine 580:
  internal/hook.swapMirrorPrimaryRoot.func1()
      internal/hook/memory_mirror_test.go:22
  testing.(*common).Cleanup.func1()
```

한 테스트가 헬퍼 진입에서 seam 을 **읽는** 동안, 다른 테스트의 `t.Cleanup` 이 같은
주소를 **복원 기록**한다.

### GREEN — 수리 후

| 명령 | exit | DATA RACE | FAIL | 로그 |
|---|---|---|---|---|
| `go test -race -count=1 ./internal/hook/` | 0 | 0 | 0 | `race-green.log` |
| `go test -race -count=2 ./internal/hook/` | 0 | 0 | 0 | `race-green-count2.log` |
| `go test -race -count=1 ./internal/hook/...` | 0 | 0 | 0 | `race-green-subtree.log` |
| `go vet ./internal/hook/...` | 0 | — | — | — |

### 공허하지 않음 (swept set 비어 있지 않음)

```
go test -race -count=1 -v -run 'TestMirror' ./internal/hook/
```
`--- PASS: TestMirror*` **8건** — 대상 테스트가 실제로 선택·실행됐다.
로그: `.moai/reports/t439/mirror-swept.log`

## Baseline-attribution

- RED 는 `f7cabfc29` 트리(= 배차문이 지목한 develop 팁)에서 이 실행으로 관측.
- GREEN 은 같은 트리에 이 카드의 수리만 얹은 상태에서 같은 명령으로 관측.
- 두 측정 사이의 유일한 차이는 아래 수리 diff 다. 다른 커밋·다른 트리에서 옮겨온 수치 없음.

## 판독 — 왜 병렬성 제거인가

배차문의 [HARD] 지시대로 수리 전에 t223 맥락을 먼저 읽었다.

1. **도입 커밋에 병렬 근거가 없다.** `47986a7af` (card: t223) 이 이 파일을 처음
   만들었고, 커밋 메시지에 `t.Parallel()` 을 요구하는 서술은 없다. 병렬은 파일의
   기본 스타일로 붙은 것이지 요구사항이 아니다.
2. **같은 패키지에 반대 방향 선례가 이미 있다.** 자매 seam `handoffRenameFunc`
   (`internal/hook/handoff_inject.go:36`) 을 교체하는 테스트 5개
   (`handoff_inject_test.go:363, 396, 451, 482, 513`) 는 **전부 `t.Parallel()` 이 없다.**
   세 번째 seam `newFanInEvidenceSourceFn` 을 교체하는
   `TestSessionEndSelectsEdgeSource` 도 직렬이다. 즉 이 패키지의 확립된 규약은
   "패키지 seam 을 바꾸는 테스트는 직렬"이고, `memory_mirror_test.go` 가 규약 밖으로
   나간 쪽이다.
3. **명문 계약도 이미 있다.** `captureStderr` (`handoff_inject_test.go:80`) 는 주석에
   *"NOT parallel-safe (mutates global os.Stderr) — callers must not t.Parallel()"* 라고
   적어 두었다. `TestMirrorFailsOpenOnUnresolvablePrimary` 는 그 계약도 함께 어기고
   있었다 — 같은 파일의 두 번째 위반이며, 레이스 스택에는 나타나지 않았다.
4. **대안(헬퍼를 테스트 로컬 상태로 전환)은 기각.** seam 을 지역화하려면
   `MirrorAgentMemoryFile` 에 주입 파라미터를 추가해야 하고, 이는 t223 범위 밖의
   프로덕션 시그니처 변경이다. 얻는 병렬성은 각 0.00s 짜리 테스트 5개분으로 사실상 0.

따라서 병렬성 제거는 "기본값이라서"가 아니라 **패키지 규약으로의 복귀**다.

## 수리 diff 요약

| 파일 | 변경 |
|---|---|
| `internal/hook/memory_mirror_test.go` | `swapMirrorPrimaryRoot` 에 직렬 계약 주석 추가; seam 을 바꾸는 테스트 4개에서 `t.Parallel()` 제거 |
| `internal/hook/agentmemory_coverage_test.go` | `TestMirrorAgentMemoryNoOps` 에서 `t.Parallel()` 제거 (형제 스윕에서 발견) |

`TestMirrorAgentMemoryDispatchedInWriteEditBranch` 는 seam 을 바꾸지 않으므로
`t.Parallel()` 을 유지했다.

## 형제 표면 스윕

seam 을 **경유해서** 바꾸는 호출부까지 훑었다 (`grep 'swapMirrorPrimaryRoot('`) —
변수명만 grep 했을 때 놓쳤던 파일이 하나 더 나왔다.

| 표면 | 결과 |
|---|---|
| `agentMemoryPrimaryRootFn` 직접 참조 | `memory_mirror_test.go` 헬퍼뿐 |
| `swapMirrorPrimaryRoot(` 호출부 6곳 | 5곳 위반(수리), 1곳(`TestMirrorAgentMemoryRelativePath`) 이미 직렬 |
| `handoffRenameFunc` 교체 5곳 | 전부 이미 직렬 — 위반 0 |
| `newFanInEvidenceSourceFn` 교체 1곳 | 이미 직렬 — 위반 0 |
| `captureStderr` 호출부 3곳 | 2곳(`handoff_inject_test.go:217, 743`) 이미 직렬, 1곳 수리 |

## Gaps — 관측하지 않은 것

- **CI 판정을 재현하지 않았다.** 배차문이 인용한 Actions run `33603867343` 의
  `--- FAIL` 50건은 리드가 확증한 CI 측정이고, 이 판정서의 62건은 **로컬 darwin
  측정**이다. 두 수치는 같은 원인의 서로 다른 관측이며 서로를 대체하지 않는다.
  최종 판정은 develop push 가 일으키는 CI 이고, push 도 판독도 리드 소관이다.
- **windows / linux 매트릭스 미측정.** 로컬은 darwin/arm64 하나뿐이다.
- **레이스 재현의 통계적 성질.** `-count=2` 까지만 돌렸다. 레이스는 스케줄 의존이라
  GREEN 이 "영원히 없음"을 증명하지는 않는다. 다만 RED 는 `-count=1` 에서도
  결정적으로 재현됐다.
- **`gofmt -l` 이 나열한 기존 미포맷 파일 16개는 손대지 않았다** — 이 카드 범위 밖.
  수리한 두 파일은 목록에 없다(포맷 클린).

## Residual-risk

- 남은 `t.Parallel()` 테스트가 나중에 이 seam 을 다시 건드리면 레이스가 되살아난다.
  방어는 헬퍼 주석(계약 명문화) 하나뿐이며, **기계적 가드는 없다.** 계약을 어겨도
  컴파일·린트는 침묵한다.
- 같은 결함 계열(패키지 전역 상태 + `t.Parallel()`)이 `internal/hook` 밖의 다른
  패키지에 있는지는 측정하지 않았다 — 범위는 이 카드가 건드린 패키지로 한정했다.
