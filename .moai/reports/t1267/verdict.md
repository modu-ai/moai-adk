# t1267 판정서 — signtest 스냅숏과 git 백그라운드 maintenance 경합

- 카드: t1267 (Class B, Tier S, SPEC 없음) · 브랜치 `WT-signtest-snapshot-race` · 기준 로컬 develop `adf909ca9`
- 증상: develop `6b1e9bbd4` Race Test (run 36231073168, job 108374290407) — `TestAC_CONTRACT_014/human_path,_stdin_is_an_empty_pipe`: `signtest: snapshot: open …/.git/objects/maintenance.lock: no such file or directory` (`internal/cli/contract_ac_test.go:318`)
- 수리 커밋: `601e1beed`
- sync-audit: PASS 94.5 (조화 93.6) — `.moai/reports/t1267/sync-audit.md`
- PR 교차확인: no-link

## 원인 (관측으로 확인)
픽스처(`signtest.New`)의 `git commit` 이 매번 `git maintenance run --auto --quiet --detach` 를 띄운다. 분리된 프로세스라 커밋이 돌아온 뒤에도 `.git` 에 계속 쓰고, `.git` 까지 걷는 `Project.Snapshot` 이 `objects/maintenance.lock` 이 지워지는 순간에 걸렸다.

증거(`GIT_TRACE`, git 2.54, 이번 실행):
```
trace: run_command: git maintenance run --auto --quiet --detach
trace: built-in: git maintenance run --auto --quiet --detach
trace: built-in: git config gc.auto 1     ← 30ms 뒤, 앞 maintenance 가 아직 도는 중
```
감사가 따로 확인: 이 잠금은 `git maintenance run` 만 소유하고 `gc --auto` 는 보지 않음, `gc.auto=0` 만으로는 분리 실행이 사라지지 않음, CI 러너 git 2.55.0.

## 재현
- 회귀 테스트 `internal/contract/sign/signtest/signtest_maintenance_test.go` (`GIT_TRACE` 로 maintenance 실행 여부 단언, 양성 대조 `built-in: git commit`): 수정 전 RED `run_command: git maintenance run --auto --quiet --detach`, 수정 후 `ok`.
- 한계: 잠금 경합 자체(파일이 사라지는 순간)는 로컬에서 재현하지 못함 — 기본 설정 20회, `gc.auto=1` 강제 5회 모두 관측 0.

## 수리
`signtest.New` 가 첫 커밋 전에 `git config maintenance.auto false`. 모든 `signtest.New` 소비자(7개 파일)가 그대로 물려받음(감사 확인: 소비자 중 자체 저장소·자체 커밋 없음).

## 검증 (이번 실행, 워크트리 t1267)
- `go test -race -count=1 ./internal/contract/... ./internal/config/` → 4 패키지 ok
- `go test -race -count=3 -run TestAC_CONTRACT ./internal/cli/` → ok 46.0s
- `golangci-lint run ./internal/contract/...` → `0 issues.` · `GOOS=windows go build ./...` → exit 0 · gofmt 0건

## 미검증
- 수리 후 develop CI Race Test 초록 (push 후 리드 판독 — 이것이 경합 소멸의 최종 근거)
- macOS·Windows CI 러너의 git 버전
- 다른 `.git` 을 걷는 테스트 5곳(`internal/spec/ac_baseline_commit_guard_test.go`, `internal/cli/factory_lane_handoff_recover_test.go`, `internal/cli/factory_handoff_abandon_test.go`, `internal/kanban/slot_lease_test.go`, `internal/hook/agentmemory_test.go`) — 다른 픽스처라 범위 밖

## 잔여 위험 (감사 선택 항목)
- F1: 다른 픽스처들은 `gc.auto 0` 도 함께 거는데 여기선 생략(정확성 무관, 관례 차이)
- F2: 회귀 테스트가 `GIT_TRACE` 줄 텍스트에 의존
- F3: `Snapshot` 이 `.git` 전체를 걷는 설계라 새 백그라운드 writer 가 생기면 같은 부류 재발 가능
