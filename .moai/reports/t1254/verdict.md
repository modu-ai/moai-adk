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

## 잔여 위험
- 같은 모양의 판정식이 `internal/cli/glamour_style.go` `writerIsTerminal` 에도 있다(stdout 대상). 주석에 "mirrors spec_status.go stdinIsTerminal" 이라고 적혀 있다. 이 카드 범위 밖이라 고치지 않았고, 결함으로 확인하지도 않았다. 후속 판단은 리드 몫이다.
