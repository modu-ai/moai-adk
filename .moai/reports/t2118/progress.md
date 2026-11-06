# t2118 — lane record (lane-28, run to2bbk)

카드: [P0·Class B·#1969 CI] Windows race 행렬의 SQLite `invalid uri authority: C:` 474건. 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2118`, 브랜치 `WT-windows-race-sqlite`, 기준 HEAD `04d8c27a7e`(= local main). Class B라 SPEC 없음 — 원인 확정 증거는 이 파일에 남김.

## 1. 출발점 (2026-11-06)

- 근거: lane-28 #1969 CI 분석(검사 head `9b632c1eb3`), 보고서 `.moai/reports/pr-sweep-20261106/pr1969-ci-analysis.md`(t2077 트리에 복사본). 분류표에서 `invalid uri authority: C:` 474줄을 레인이 직접 셈(`grep -c` → 474). 패키지는 agenticfactory/ledger, mergeq, cli 등(분석 보고 기준).
- 레인 사전 조사(읽기만): DSN을 만드는 함수가 이미 4개 — `internal/factory/backlog_sqlite.go:270 backlogDSN`, `internal/homestate/profile_lease.go:51 profileLeaseDSN`, `internal/agenticfactory/ledger/importer/importer.go:165 readOnlyDSN`, `internal/factorymsg/store.go:163 brokerDSN`. `backlogDSN`에는 Windows 드라이브 경로를 `file:///C:/…`로 만들고 `#`·`%`·공백을 인코딩하는 시험이 있음(`internal/factory/backlog_sqlite_test.go:292-298`, `backlog_dsn_windows_test.go`). 반면 `"file:"+filepath.ToSlash(path)` 직접 연결이 운영 코드에 있음: `internal/factorymsg/run_state.go:89/128/172`, `internal/factorymsg/store.go:206/916`, `internal/cli/factory_card.go:3977`, `internal/cli/migrate_home_state.go:409/500`. 시험 코드에도 다수. 어느 지점이 474건을 냈는지는 아직 확정 안 함 — manager-develop가 확정.
- codex 규칙: 상시 결정 `d-20261106T080039Z-305c`(감사 단계마다 codex_audit 최대 2회, 감사관 자신 호출, 트리 불변, app-server·audit_multi 금지).

## 2. 진행 기록

(이하 단계별로 추가)

## 3. 원인 확정 (Class B cause-establishing evidence)

근거: CI 분류표 `categorized.tsv`(scratchpad, 읽기 전용)에서 `invalid uri authority` 줄 474건을 패키지·시험별로 집계. 명령: `grep 'invalid uri authority' categorized.tsv | awk -F'\t' '{print $3}' | sort | uniq -c`.

| 패키지 | 건수 |
|---|---|
| internal/agenticfactory/ledger | 416 |
| internal/agenticfactory/mergeq | 34 |
| internal/cli | 22 |
| internal/homestate | 1 |
| internal/web | 1 |
| 합계 | 474 |

오류 문자열은 두 모양이다.
1. `invalid uri authority: C:%5CUsers%5CRUNNER~1%5C…%5Cledger.db?_pr…` (473건 — ledger 416 + mergeq 34 + cli 22 + web 1): 백슬래시가 그대로 인코딩된 경로. 생산자는 `internal/agenticfactory/ledger/schema.go` `Open`(변경 전 줄 85-92)의 `url.URL{Scheme:"file", Path: path}` — 원시 Windows 경로를 넘겨 `file://C:%5CUsers…`가 되고 드라이버가 `C:%5C…`를 authority로 읽음. 모든 시험이 `ledger.Open`을 거쳐(af_flow_test.go, af_land_*_test.go, mergeq 시험, web board_store_test.go) 같은 한 지점에서 실패. 분류표 5열 파일명은 474건 전부 `_test.go`(시험이 연 저장소가 ledger).
2. `invalid uri authority: C: (1)` (1건 — homestate `TestRuntimeCensusCountsLiveAndIgnoresProvablyDead`): 생산자는 `internal/homestate/runtime_census.go`의 `filepath.ToSlash(factoryPath)`만 쓰고 앞 `/`를 붙이지 않는 DSN — `file://C:/…`.

재현(이 OS에서 문자열 수준, RED 출력 `.moai/reports/t2118/red-ledger.txt`): `ledgerDSN("C:\Users\dev\x.db")` = `file://C:%5CUsers%5Cdev%5Cx.db?…` — CI 오류 문자열과 같은 모양. `runtimeCensusDSN` = `file://C:%5C…` 파싱 불가.

잠복 지점(CI 증거에 안 나왔으나 같은 계열): `"file:"+filepath.ToSlash(path)` 직접 연결 — factorymsg/run_state.go ×3, factorymsg/store.go ×2, cli/factory_card.go, cli/migrate_home_state.go ×2 (`file:C:/…`은 authority가 없어 열릴 수는 있으나 `#`·`%`·`?`·공백이 든 경로를 인코딩하지 못함), discovery/factory_discovery.go(앞 `/` 없음). 이미 올바른 빌더(backlogDSN, profileLeaseDSN, importer readOnlyDSN, brokerDSN, ledger OpenReadOnly, homestate/factory.go 2곳, factoryQuotaReadOnlyDSN)는 건드리지 않음.

수리: 공유 빌더 `internal/sqliteuri.File(path, query)`(표준 라이브러리만, 순환 없음) — 드라이브 경로의 `\`를 `/`로 바꾸고 `/` 접두 → `file:///C:/…`, 특수문자 인코딩, 상대 경로는 절대화.

## 4. 구현 기록

- RED `e267f5f4a6` — sqliteuri.File 스텁 + ledgerDSN/runtimeCensusDSN 동작 보존 추출 + 문자열 수준 시험 + 운영 코드 손수 연결 감시 시험. 실패 출력: red-sqliteuri.txt / red-ledger.txt / red-homestate.txt.
- GREEN `9deb412a48` — `internal/sqliteuri.File` 도입, ledger·homestate census·discovery·factorymsg(5곳)·cli(3곳) 전환. 운영 코드 `"file:"+` 손수 연결 9→0.
- 검증(모두 이 트리, 로컬 OS=darwin): sqliteuri 11 pass, ledger 패키지 전체 486 pass, mergeq 91 pass, factorymsg 235 pass, homestate 763 pass(첫 실행은 기본 10분 timeout에 걸려 -timeout 25m로 재실행), discovery 15 pass, cli 집중(TestHomeState 61, TestFactoryStaleRunRefused 1). 0 fail. Windows build/vet 통과. lint 새 지적 0(기존 staticcheck 2건: discovery/leader_readers_darwin.go:93 SA1019, factorymsg/store.go:1333 SA4023 — 내 diff 밖).
- 미검증: Windows 실제 실행(이 OS에서 불가), internal/cli 전체·internal/web 시험(집중 규칙), 시험 코드의 손수 연결 32곳(CI 증거에 없음, 범위 밖).

## 5. 레인 재확인 (lane-28, 2026-11-06 13:3xZ, HEAD 9deb412a48)

- 커밋 순서: RED `e267f5f4a6` → GREEN `9deb412a48` (git log 직독). RED가 수정보다 먼저 별도 커밋 — verification-claim-integrity §2.3 충족.
- 빌드(레인 직접, 전용 GOCACHE): `GOOS=windows GOARCH=amd64 go build ./...` → `winbuild exit=0`, `go build ./...` → `build exit=0`. `GOOS=windows GOARCH=amd64 go test -c ./internal/sqliteuri/` → exit 0.
- 집중 시험(레인 직접): `go test -count=1 ./internal/sqliteuri/ ./internal/agenticfactory/ledger/` → `ok internal/sqliteuri 3.617s`, `ok internal/agenticfactory/ledger 110.466s`.
- Security Guardian "path-traversal (medium) line 85": 해당 줄은 `internal/agenticfactory/ledger/schema.go:85` `return sqliteuri.File(path, q)`. path는 호출자가 정한 원장 파일 위치(설정·시험 임시 디렉터리)이며 사용자 입력 경계가 새로 생긴 것이 아님 — 수정 전에도 같은 path가 `url.URL{Path: path}`로 드라이버에 갔음. 레인 판단: 오탐(새 경로 처리 위험 없음). sync 감사관 재확인 대상.
- Gaps: Windows 실제 실행 없음(CI Windows race 행렬이 판정). internal/cli 전체·internal/web 시험 미실행(레인 집중 규칙). 위임의 그 밖의 패키지 통과 수(mergeq 91, factorymsg 235, homestate 763, discovery 15, cli 집중 62)는 위임 보고·증거 파일 기준이며 레인이 재실행하지 않음. 이전 턴 위임 기록 조회 명령 1건이 워크트리 가드에 거절됨(위임 백그라운드 명령 확인용, 판정 근거 아님).
