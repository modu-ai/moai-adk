# t606 — 커버리지 검사 실패 메시지의 지목 파일이 실행마다 바뀌는 결함

- card: t606 (Class B, Tier S) · t600 형제
- worktree: `.claude/worktrees/t606`, branch `WT-fail-name-order`
- base: 로컬 develop `b155c95f9` (t600 병합 포함, fast-forward)
- 상태: run 완료 — RED(흔들림 실측) → 수리 → GREEN → 뮤턴트 → vet·lint 통과. 미푸시. develop 병합은 리드 창 지명 대기.

## 1. 원인 — 코드 판독

`internal/cli/home_state_coverage.go` 의 `resolveHomeStateCoverageChangeSet` 에서 두 검사가 map 을 순회하다 **처음 걸린 파일 하나**만 오류 메시지에 넣고 반환한다. Go 의 map 순회 순서는 실행마다 무작위이므로, 걸리는 파일이 둘 이상이면 같은 입력에서도 메시지가 가리키는 파일이 바뀐다.

| 줄 | 순회 대상 | 메시지 |
|---|---|---|
| `:295-299` | `expectedBlobs` (map) | `audited production file changed after coverage tip: %s` — 카드가 지목한 검사 |
| `:344-347` | `all` (map) | `changed production file missing from diff: %s` — 같은 함수의 형제 검사, 같은 결함 |

### 이미 관측된 흔들림

t600 의 재현 실행(`.moai/reports/t600/run1-head-81c1d58f9.txt`, HEAD `81c1d58f9`)에서 **한 번의 `go test` 실행 안에** 같은 검사가 서로 다른 파일을 지목했다.

```
home_state_coverage_test.go:420: audited production file changed after coverage tip: internal/cli/launcher.go
        home_state_coverage_test.go:420: audited production file changed after coverage tip: internal/hook/session_end.go
```

같은 트리에서 불일치 후보는 4개였다(t600 verdict §3 모델: `launcher.go`, `mcp_server.go`, `hook/session_end.go`, `kanban/todo_root.go`).

## 2. 재현 설계

t600 수리 뒤 라이브 저장소를 읽는 테스트는 fixture 로 옮겨졌으므로, 흔들림은 불일치 후보가 둘 이상인 fixture 로 잰다. 한 프로세스에서 resolver 를 20회 호출해 서로 다른 오류 메시지 개수를 센다(`resolveRepeatedly`).

- `TestCommittedCoverageChangeSetNamesEveryPostTipChangeDeterministically` — 감사 파일 `internal/x/a.go`(rollout marker), `internal/y/b.go`(remediation marker)를 둘 다 인증 뒤 커밋으로 바꾼다. 기대: `audited production file changed after coverage tip: internal/x/a.go, internal/y/b.go` 1종 ×20.
- `TestCommittedCoverageChangeSetNamesEveryFileMissingFromDiffDeterministically` — 감사 대상 아닌 파일 둘을 remediation marker 에서 모드만 바꾼다(`git update-index --chmod=+x`). name-status 에는 나오지만 hunk 가 없어 changed-line range 에 들어가지 않는다. 기대: `changed production file missing from diff: internal/x/b.go, internal/x/c.go` 1종 ×20.

## 3. 수리

`internal/cli/home_state_coverage.go` 에서 두 검사 모두 불일치를 전부 모아 `sort.Strings` 로 정렬한 뒤 `", "` 로 나열한다. 불일치가 하나면 메시지는 이전과 바이트 단위로 같다(t600 의 `…ConsumesFreshProfile` 단정 유지). 형제 검사(`missing from diff`)는 분류 루프 앞의 별도 사전 순회로 옮겼다. 두 수정 모두 253행 아래라 `TestAuditLagUsesBinlagSeam` 의 좌표(245/253)는 그대로다. 형제 결함을 범위에 넣는 것은 리드가 승인했다.

## 4. 검증 (리드 슬롯 지명, 시작 시 load 30.02)

### 4.1 RED — 수리 전 트리

작업 사본 상태: `M internal/cli/home_state_coverage_test.go` 만 수정(production 파일은 수리 전).

```
go test ./internal/cli/ -count=1 -run '^TestCommittedCoverageChangeSetNamesEvery' -v
exit=1   (run1-red-prefix.txt)
home_state_coverage_test.go:731: messages across 20 runs = map[audited production file changed after coverage tip: internal/x/a.go:17 audited production file changed after coverage tip: internal/y/b.go:3], want only "... internal/x/a.go, internal/y/b.go"
--- FAIL: TestCommittedCoverageChangeSetNamesEveryPostTipChangeDeterministically (13.96s)
home_state_coverage_test.go:750: messages across 20 runs = map[changed production file missing from diff: internal/x/b.go:16 changed production file missing from diff: internal/x/c.go:4], want only "... internal/x/b.go, internal/x/c.go"
--- FAIL: TestCommittedCoverageChangeSetNamesEveryFileMissingFromDiffDeterministically (27.11s)
```

흔들림 실측: 같은 입력으로 20회 호출했을 때 두 검사 모두 지목 파일이 두 가지로 갈렸다(17/3, 16/4).

### 4.2 GREEN — `fix.patch` 적용 후

```
go test ./internal/cli/ -count=1 -run '^(TestCommittedCoverage.*|TestHomeStateChangedSurfaceCoverageConsumesFreshProfile|TestChangedProductionFiles.*|TestAuditLagUsesBinlagSeam)$' -v
exit=0   (run2-green.txt)
top-level: --- PASS 14 / --- FAIL 0 / --- SKIP 0
--- PASS: TestCommittedCoverageChangeSetNamesEveryPostTipChangeDeterministically (17.98s)
--- PASS: TestCommittedCoverageChangeSetNamesEveryFileMissingFromDiffDeterministically (38.11s)
--- PASS: TestAuditLagUsesBinlagSeam (1.22s)
ok  	github.com/modu-ai/moai-adk/internal/cli	99.224s
```

### 4.3 뮤턴트 — 두 `sort.Strings` 제거

```
go test ./internal/cli/ -count=1 -run '^TestCommittedCoverageChangeSetNamesEvery' -v
exit=1   (run3-mutant-nosort.txt)
...: internal/x/a.go, internal/y/b.go:16  /  ...: internal/y/b.go, internal/x/a.go:4
...: internal/x/b.go, internal/x/c.go:16  /  ...: internal/x/c.go, internal/x/b.go:4
--- FAIL 2
```

원복: 변이 전 사본(`.moai/state/verify/t606/home_state_coverage.fixed.go`)을 되돌려 놓고 `cmp` 로 바이트 동일 확인, `sort.Strings(changed|missing)` 2건 복귀. 미커밋 수리가 같은 파일에 있어 `git restore` 는 쓰지 않았다.

### 4.4 vet / lint

```
go vet ./internal/cli/                vet-exit=0   (run4-vet.txt)
golangci-lint run ./internal/cli/     lint-exit=0   0 issues.   (run4-lint.txt)
```

## Gaps

- `internal/cli` 전체 대조군은 리드 지시대로 병합 창에서 병합 트리로 한 번 돌린다(아직 안 돌림).
- 20회 반복의 흔들림 검출은 확률적이다. 후보 2개면 20회가 모두 같은 순서로 나올 확률은 약 2×(1/2)^20 이므로, 결정성이 깨지면 사실상 항상 잡힌다. 다만 원리상 0 은 아니다.
- darwin 에서만 실행했다. 모드 변경 fixture 는 `git update-index --chmod` 를 써서 파일시스템 권한에 기대지 않지만, windows 판정은 CI 몫이다.
- 부하가 높은 상태(load 30)에서 쟀다. 통과·실패 판정에는 영향이 없지만 실행 시간 수치는 그 영향을 받았다.

## Residual-risk

- 오류 메시지를 파싱하는 소비자가 있다면, 불일치가 여럿일 때 쉼표로 나열된 형식을 새로 만나게 된다. `grep -rln "changed after coverage tip\|missing from diff" internal cmd pkg` 의 출력은 `internal/cli/home_state_coverage.go` 와 `internal/cli/home_state_coverage_test.go` 둘뿐이라, 저장소 안에 이 메시지를 파싱하는 소비자는 없다. 저장소 밖 도구가 이 문자열을 읽는지는 확인하지 않았다.
