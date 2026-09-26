# t1262 판정서 — `/dev/null` 을 터미널로 오판하는 TTY 판별 (t1254 형제)

- 카드: t1262 (Tier S, Class B)
- 브랜치: `WT-codex-init-tty` (기준 develop `422aa1e20`)

## 주장

가설은 참이다. `os.ModeCharDevice` 로 터미널을 판별하던 두 곳이 `/dev/null`(문자 장치이지만 터미널이 아님)을 터미널로 판정했다.

| 위치 | 결함 | 영향 |
|---|---|---|
| `internal/cli/codex_init.go` `defaultCodexPromptCapable` | 결함 | `</dev/null` 에서 질문 가능으로 판정 → 프롬프트 출력 후 EOF 를 읽어 조용히 거절 |
| `internal/cli/glamour_style.go` `writerIsTerminal` | 결함 | `>/dev/null` 에서 glamour 리치 렌더링 경로로 진입(출력은 버려지지만 판정이 틀림) |
| `internal/cli/statusline.go` `readStdinWithTimeout` | 판정은 같은 방식으로 틀리지만 **행동 결함 없음** | `/dev/null` 을 터미널로 보고 빈 reader 를 돌려주는데, `/dev/null` 을 읽어도 즉시 EOF 이므로 결과가 동일. 범위 밖으로 두고 수정하지 않음 |

## 증거

수리 전 재현(`internal/cli/tty_devnull_test.go`):

```
$ go test ./internal/cli/ -run DevNull -count=1
--- FAIL: TestDefaultCodexPromptCapableRejectsDevNull (0.00s)
    tty_devnull_test.go:20: defaultCodexPromptCapable() with stdin=/dev/null = true, want false
--- FAIL: TestWriterIsTerminalRejectsDevNull (0.00s)
    tty_devnull_test.go:31: writerIsTerminal(/dev/null) = true, want false
FAIL
```

수리: 두 함수 모두 t1254 가 도입한 `isTerminalFile`(`term.IsTerminal`)로 교체.

수리 후:

```
$ go test ./internal/cli/ -run 'DevNull|IsTerminalFile|Codex|Glamour|Markdown|WriterIsTerminal' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	127.081s
$ go vet ./internal/cli/ && golangci-lint run ./internal/cli/
0 issues.
```

## 기준선 귀속

이 워크트리, 기준 develop `422aa1e20` 위에서 이번 실행으로 측정.

## 미검증

- 실제 터미널(TTY)에서 두 함수가 `true` 를 내는지는 자동 테스트로 확인하지 않음(테스트 환경에 TTY 없음). `term.IsTerminal` 은 t1254 에서 같은 방식으로 이미 쓰이고 있음.
- Windows 동작은 CI 매트릭스에 맡김.
- `internal/cli` 전체 스위트는 로컬에서 돌리지 않음(§4 규율) — CI 판정.

## 잔여 위험

- `readStdinWithTimeout` 의 판별도 같은 모양이다. 현재는 무해하지만 나중에 분기 행동이 바뀌면 드러날 수 있다.
