# t1254 판정서 — spec status 비-TTY 가드의 /dev/null 오판

클래스 B · Tier S · plan 생략 · 기준 develop `38148d891`

## 주장
`moai spec status --sync-git` 의 비-TTY 가드(`internal/cli/spec_status.go` `stdinIsTerminal`)는 `/dev/null` 을 터미널로 판정했다. 원인은 `os.ModeCharDevice` 비트만 본 판정이다. `/dev/null` 은 문자 장치지만 터미널은 아니다. 판정을 `isTerminalFile(f)` 로 분리하고 `term.IsTerminal(int(f.Fd()))` 로 바꿔 수리했다.

## 원인 근거 (재현)
- 명령: `go test -count=1 -run TestIsTerminalFile -v ./internal/cli/` (판정 분리만 하고 ModeCharDevice 로직은 그대로 둔 상태)
- 출력:
  - `spec_status_tty_test.go:18: isTerminalFile(/dev/null) = true, want false`
  - `--- FAIL: TestIsTerminalFileRejectsDevNull`
  - `--- PASS: TestIsTerminalFileRejectsPipe` (대조군: 파이프는 원래도 비-TTY 로 판정됨)
  - exit=1. 전체 로그: `repro.log`
- 원인 검토: 대조군이 통과하고 `/dev/null` 만 실패했다. 문제는 가드 호출부가 아니라 "문자 장치 = 터미널" 이라는 판정식 자체다. t1234 N2 프로브(`devnull: ModeCharDevice=true term.IsTerminal=false`)와도 맞는다.

## 수리 후 증거
- 같은 명령: 두 테스트 모두 PASS, exit=0 (`after.log`)
- `go test -count=1 -run SpecStatus ./internal/cli/` exit=0 (`-list` 기준 15건, `specstatus.log`)
- `go test -count=1 -run ExitCode ./internal/cli/` exit=0 (`-list` 기준 20건, 비-TTY 중단 경로 포함, `exitcode.log`)
- `go vet ./internal/cli/` 통과, `gofmt -l` 출력 없음
- 새 의존성 없음: `golang.org/x/term v0.45.0` 은 이미 go.mod 의 직접 의존

## 미검증
- 실제 터미널에서 대화형 프롬프트가 계속 뜨는지는 사람이 직접 확인하지 않았다. 테스트 러너의 stdin 은 TTY 가 아니다.
- Windows 콘솔은 실행하지 않았다. `term.IsTerminal` 이 Windows 를 지원하지만 이 트리에서 관측하지는 않았다.
- `</dev/null` 로 CLI 를 끝까지 돌리는 확인은 하지 않았다. 판정 함수 단위로만 증명했다.
- `internal/cli` 전체 스위트는 CI 몫이다.

## 독립 sync 감사 (sync-auditor, opus)
- 1차 판정: **FAIL**, 조화평균 78.4 (Functionality 95 / Security 95 / Craft 60 / Consistency 75). 전문: `sync-audit.md`
- 감사가 직접 잰 것: 기준 `38148d891` 바이너리로 `moai spec status --sync-git </dev/null` 실행 → `Apply? [y/N]:` 후 `skipped 1 (already done)`, exit 0(결함 재현). 수리 바이너리 → `interactive confirmation unavailable in non-TTY context`, exit 1, 파일 무변경. `--yes` 경로는 그대로 `updated 1`.
- F1 [High, 차단] 새 테스트의 `defer X.Close()` 3곳이 errcheck 위반 → CI lint 잡 적색 예정이었음. **수리함**: 저장소 관례 `defer func() { _ = X.Close() }()` 로 교체. 재측정 `golangci-lint run ./internal/cli/` → `0 issues.` exit=0 (로컬 v2.10.1, `lint.log`). `TestIsTerminalFile` 재실행 ok.
- F2 [Low] `internal/cli/contract.go:55-57` 주석이 "stdinIsTerminalFn 은 널 장치를 통과시킨다"고 적어 이번 수리로 거짓이 됨. **수리함**(주석만).
- F6 [Low] 로그 파일들은 gitignore 대상이라 커밋되지 않음 — 결정 줄은 이 판정서에 인용해 둠.
- F7 [Info] 한 줄 래퍼 `stdinIsTerminal` 커버리지 0% — 판정 본체 `isTerminalFile` 은 100%.
- 재감사 생략(리드 판단): `sync-audit.md` 186행이 "F1 을 수리하면 재감사는 F1 한 건(`golangci-lint run ./internal/cli/` 0건 확인)으로 범위를 좁혀도 된다"고 정했고, 레인이 그 명령으로 `0 issues.` 를 관측했다. 로컬 lint 는 v2.10.1, CI 는 v2.1.6 이므로 최종 판정은 CI 다.

## 미확인 형제 후보 (범위 밖 — 후속 카드 여부는 리드 판단)
같은 `os.ModeCharDevice` 판정식이 남아 있는 곳. 판정식이 `/dev/null` 에 true 를 낸다는 것만 확인했고, 명령을 끝까지 돌려 결함을 확인하지는 않았다.
- `internal/cli/codex_init.go:116-118` `defaultCodexPromptCapable` — stdin 프롬프트 게이트. `</dev/null` 이면 비대화형 거절(REQ-CI-009)이 우회될 수 있다는 가설. 감사 판단상 우선순위가 가장 높다(Medium).
- `internal/cli/glamour_style.go:135,146` `writerIsTerminal` — stdout 이 `/dev/null` 이면 리치 렌더링이 켜짐. 출력이 버려지므로 실질 영향 작음(Low). "mirrors stdinIsTerminal" 주석은 이제 틀림.
- `internal/cli/statusline.go:170` — 어느 경로든 읽을 데이터가 없어 동작 차이 없음(Info).

## 잔여 위험
- CI 의 golangci-lint(v2.1.6)와 로컬(v2.10.1) 버전이 달라, CI 판정은 develop push 뒤 확인해야 한다.
- 위 형제 후보가 실제 결함이면 같은 계열 결함이 이 수리 뒤에도 남는다.
