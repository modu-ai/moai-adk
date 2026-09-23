# t1110 판정서 — 런처의 `--no-chrome` 기본 주입 제거

- 카드: t1110 (Tier S · 클래스 B, SPEC 없음)
- 브랜치: `WT-launcher-chrome-on` (기준 로컬 develop `176d8b658`)
- 워크트리: `.claude/worktrees/t1110`
- 커밋: `21d2afc09` 수리 본체, 그 뒤 후속 커밋 1건(회귀 테스트 보강·CHANGELOG·이 판정서)

## 주장

1. `moai cc`·`glm`·factory/kanban 리드·레인이 공유하는 기동 경로(`internal/cli/launcher.go`)가 더 이상 `--no-chrome`을 스스로 붙이지 않는다.
2. 사용자가 명시한 `--no-chrome`·`--chrome`은 claude 에 한 번씩 그대로 전달된다. `--` 뒤에 넘긴 경우도 같다.
3. `DO_CLAUDE_CHROME` 은 더 이상 읽지 않으며, 옛 값이 남아 있어도 오류가 나지 않는다.
4. 범위 밖으로 둔 `internal/cli/mcp_claude.go:214`(감사용 헤드리스 `-p` 하위 프로세스)와 템플릿 스킬 참고 문서의 `--no-chrome` 예시는 바뀌지 않았다.

## 증거 (이 트리에서 실행)

| 단계 | 명령 | 관측 |
|---|---|---|
| RED (수리 전, `176d8b658` + 테스트) | `go test ./internal/cli/ -run 'Chrome' -count=1` | exit 1 · `argv = [claude --no-chrome] carries --no-chrome 1 time(s)` · `--chrome` 하위 테스트 `argv = [claude] carries --chrome 0 time(s)` (`red.log`) |
| GREEN | `go test ./internal/cli/ -run 'Chrome' -count=1 -v` | exit 0 · 테스트 3개, 하위 테스트 3개(`--no-chrome`·`--chrome`·`after separator`) 모두 PASS (`green2.log`) |
| 변이 ① 기본 주입 복원 | `buildArgs` 의 `a := []string{"claude"}` 바로 뒤에 `a = append(a, "--no-chrome")` 삽입 | exit 1 · 3개 테스트 FAIL (`mutant-inject.log`) |
| 변이 ② 명시 플래그 삼킴 | 인자 파서에 `case "--no-chrome":`(본문 없음) 추가 | exit 1 · `TestLaunchForwardsExplicitChromeFlags/--no-chrome` FAIL (`mutant-drop.log`) |
| 인접 테스트 | `go test ./internal/cli/ -run 'Launch'` / `'Unified'` / `'ReadSettingsLocal'` (각각 `-count=1`) | 셋 다 exit 0 · `ok` (`targeted-{launch,unified,settings}.log`) |
| lint | `golangci-lint run ./internal/cli/` | exit 0 · `0 issues.` (`lint.log`) |
| 빌드 | `go build ./...` 와 `GOOS=windows GOARCH=amd64 go build ./internal/cli/` | 둘 다 exit 0 |
| docs | `hugo --source docs-site --quiet --destination <scratchpad>` | exit 0 · 출력 없음(`hugo.log` 0바이트는 `--quiet` 때문) |

두 변이 모두 판정 후 원본으로 되돌렸다. 되돌린 뒤 `launcher.go` 에 남은 `no-chrome` 문자열은 주석 2곳뿐이다.

## 독립 감사

`sync-auditor` 판정은 **PASS** 이며 가중 조화평균은 약 92.7 이다. 전문은 `sync-audit.md` 에 있다.
- claude argv 를 만드는 곳은 `launcher.go` 의 `buildArgs` 한 곳이고, cc·glm·kanban/factory·`--spawn`·`-w` 는 모두 `unifiedLaunch → launchClaude → runLaunchClaude` 로 모인다(감사자 확인).
- 차단 결함은 0건이다. 참고 결함 처리 결과는 다음과 같다.
  - F1(로그에 명령이 없음): 위 표로 해소.
  - F2(`hugo.log` 가 0바이트): `--quiet` 때문이며 exit 0 을 이 판정서에 기록해 해소.
  - F3(`-- --no-chrome` 회귀 테스트 없음): `after separator` 하위 테스트를 추가해 해소.
  - F4(CHANGELOG 없음): `CHANGELOG.md` Unreleased › Fixed 에 항목을 추가해 해소.
  - F5(`moai glm --help` 에 chrome 행 없음): 이번 커밋 이전부터 있던 상태이고 범위 밖이라 **미처리**로 남긴다.

## 미검증

- 전체 스위트(`go test ./...`)와 CI 매트릭스(darwin/windows)는 돌리지 않았다. push 뒤 develop CI 가 판정한다.
- 실제 claude 세션에서 `/chrome` 이 연결되는지는 수동으로 확인하지 않았다. argv 에 `--no-chrome` 이 없다는 것까지만 측정했다.
- 변이 두 방향은 이 레인이 실행했고, 감사자는 로그로만 확인했다.

## 남은 위험

- 전에는 `DO_CLAUDE_CHROME=true` 로 Chrome 을 켜 두던 사용자도 이제 Claude Code 기본값을 따른다. 그 기본값이 꺼짐이면 `--chrome` 을 직접 넘겨야 한다. 이 내용은 CHANGELOG 와 docs 에 적었다.
