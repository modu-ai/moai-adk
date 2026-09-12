# t525 F1 수리 — internal/cli 슬롯 실행 기록

트리: `.claude/worktrees/t525`, 브랜치 `WT-speclint-red`, HEAD `25f689832`(F1 수리 `85cf76b2d` + `--json --sarif` 확장 `25f689832`). 리드가 슬롯을 지명했다. 시작 직전 확인에서 슬롯 밖 실행(다른 세션의 `go test -race ./internal/cli ./internal/gateway …`, `cli.test`)이 보여 끝날 때까지 기다린 뒤 시작했다. 매 실행 직전에 `pgrep -f 'go (test|build|vet)'`·`pgrep -x cli.test` 가 비어 있음을 다시 확인했다.

## RED — 수리 전 코드로 되돌린 상태

- 뮤턴트: `git show cb1b9a16d:internal/cli/spec_lint.go > internal/cli/spec_lint.go`(파일 전체를 수리 전 판으로). 적용 뒤 `git hash-object` = `f088f2b91…` = `git rev-parse cb1b9a16d:internal/cli/spec_lint.go`(`mutant-sha.txt`).
- 명령: `go test -count=1 -timeout 1200s -run 'TestSpecLintBaseline_FlagContractRejections' -v ./internal/cli/` → `red.txt`, `red.exit`.
- 결과: exit 1. `=== RUN` 9, `--- PASS` 0, `--- FAIL` 9(부모 1 + 서브테스트 8). 실패 메시지 `stderr is empty` 8건. 서브테스트 8개에 새 `--json_with_--sarif` 가 들어 있다.

판독: 테스트는 수리 전 코드에서 거절 경로 8개 모두를 "사용자가 아무것도 못 본다"는 이유로 붉게 만든다. 붉은 이유가 결함 그 자체다.

## 복원

- 백업(뮤턴트 전 사본)으로 되돌림 → `cmp` exit 0, sha256 `6f2e27f5533a46c8e10df57b2bd8f4d3b7dc58d32db2f67a8770aca0f9ef9d85`(뮤턴트 전 값과 같음), 추적 수정 0.

## GREEN

| 명령 | 파일 | exit | RUN | PASS | FAIL | SKIP | `no tests to run` |
|---|---|---|---|---|---|---|---|
| `-run 'TestSpecLintBaseline_FlagContractRejections'` | `green.txt` | 0 | 9 | 9 | 0 | 0 | 0 |
| `-run 'TestSpecLint'` (이 파일 전체 계열) | `green-file.txt` | 0 | 33 | 33 | 0 | 0 | 0 |

공통 인자: `-count=1 -timeout 1200s -v ./internal/cli/`, 환경 스크럽 `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED`.

## 시각·부하·홈

- RED 20:14:17–20:14:30, GREEN 20:14:49–20:15:16, 끝 시점 load 22.61(`times.txt`).
- 실행 전후 동일: `~/.claude/settings.json` sha256 `86e2d9b6…`, `~/.zshrc`·`~/.zprofile`·`~/.zshenv`·`~/.gitconfig` mtime.

## 미검증

- `internal/cli` 패키지의 다른 테스트는 돌리지 않았다(슬롯 범위는 이 파일 계열).
- 기준선 로드 실패(exit 3)·쓰기 실패(exit 2) 경로는 코드 판독뿐이다(교차 프로세스·테스트 단언 없음).
- windows·darwin 매트릭스와 CI.
