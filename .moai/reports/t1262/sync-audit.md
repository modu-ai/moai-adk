# t1262 독립 감사 판정서 (sync-auditor)

- 카드: t1262 (Tier S, Class B, SPEC 없음)
- 감사 대상: `WT-codex-init-tty` @ `f5a9f536d` (기준 develop `422aa1e20`)
- 감사 방식: 소스 읽기 전용. 기준 판 재현은 스크래치 사본(`git archive HEAD` + 기준 판 두 파일 덮어쓰기)에서 수행했고, 워크트리는 건드리지 않음
- 평가 프로필: 기본 프로필(Functionality·Security 필수 통과)

## 판정

**PASS-WITH-DEBT** — 조화평균 **90.5**

| 차원 | 점수 | 판정 | 근거 요약 |
|---|---|---|---|
| Functionality (필수) | 95 | PASS | 기준 판에서 새 테스트 2건 FAIL, HEAD 에서 PASS — 실제로 결함을 잡는 테스트임을 독립 재현 |
| Security (필수) | 95 | PASS | 보안 표면 없음. 프롬프트 게이트가 비대화형 입력을 비대화형으로 판정하도록 바뀐 쪽이라 안전 방향 |
| Craft | 85 | PASS | 수정은 최소(함수 본문 2개 → 기존 헬퍼 호출). 형제 조사에서 테스트 헬퍼 1곳 누락(F1) |
| Consistency | 88 | PASS | t1254 의 `isTerminalFile` 재사용, 주석 갱신. 패키지 안 TTY 판별 관용구는 여전히 셋(F5) |

조화평균: 4 / (1/95 + 1/95 + 1/85 + 1/88) = 90.5

부채(debt)는 이 변경이 새로 만든 것이 아니라, 같은 결함 계열이 테스트 헬퍼에 남아 있고 그 때문에 인수 기준 테스트 하나가 늘 건너뛰어진다는 기존 사실이다(F1). 차단 사유는 아니며 별도 카드로 다룰 사안이다.

## 확인 항목별 결과

### (1) 테스트가 기준 판에서 실패하고 HEAD 에서 통과하는가 — 확인

HEAD:

```
$ go test ./internal/cli/ -run DevNull -count=1 -v
=== RUN   TestIsTerminalFileRejectsDevNull
--- PASS: TestIsTerminalFileRejectsDevNull (0.00s)
=== RUN   TestDefaultCodexPromptCapableRejectsDevNull
--- PASS: TestDefaultCodexPromptCapableRejectsDevNull (0.00s)
=== RUN   TestWriterIsTerminalRejectsDevNull
--- PASS: TestWriterIsTerminalRejectsDevNull (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.885s
```

스크래치 사본(HEAD 트리 + `git show 422aa1e20:` 로 되돌린 `codex_init.go`·`glamour_style.go`):

```
internal/cli/codex_init.go:118:	return err == nil && fi.Mode()&os.ModeCharDevice != 0
internal/cli/glamour_style.go:146:	return (fi.Mode() & os.ModeCharDevice) != 0
=== RUN   TestIsTerminalFileRejectsDevNull
--- PASS: TestIsTerminalFileRejectsDevNull (0.00s)
=== RUN   TestDefaultCodexPromptCapableRejectsDevNull
    tty_devnull_test.go:20: defaultCodexPromptCapable() with stdin=/dev/null = true, want false
--- FAIL: TestDefaultCodexPromptCapableRejectsDevNull (0.00s)
=== RUN   TestWriterIsTerminalRejectsDevNull
    tty_devnull_test.go:31: writerIsTerminal(/dev/null) = true, want false
--- FAIL: TestWriterIsTerminalRejectsDevNull (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.921s
```

경합·커버리지:

```
$ go test ./internal/cli/ -run DevNull -count=1 -race -coverprofile=…/devnull.cov
ok  	github.com/modu-ai/moai-adk/internal/cli	2.314s	coverage: 5.2% of statements
codex_init.go:116:	defaultCodexPromptCapable	100.0%
glamour_style.go:137:	writerIsTerminal		75.0%
spec_status.go:302:	isTerminalFile			100.0%
```

`writerIsTerminal` 의 나머지 25%(비 `*os.File` 분기)는 `glamour_style_test.go:95` 의 `bytes.Buffer` 경로가 덮는다.

정적 검사:

```
$ go vet ./internal/cli/ && echo VET_OK
VET_OK
$ golangci-lint run ./internal/cli/
0 issues.
```

### (2) 옛 동작에 기대는 호출자·테스트 — 없음

- `codexPromptCapableFn` 을 쓰는 기존 테스트는 `codex_init_test.go:318-323` 에서 함수를 스텁으로 바꿔 끼우므로 실제 판별과 무관하다. `codex_init_test.go:961` 은 패닉 여부만 본다.
- `writerIsTerminal` 의 다른 호출자는 `glamour_style.go:165,183`, `doctor_render.go:115` 이며, 테스트는 `bytes.Buffer` 를 넘긴다.

```
$ go test ./internal/cli/ -run 'TestRunVersionBranch_NonTTYProceeds|Glamour|WriterIsTerminal|MarkdownRich|DefaultCodexInitGenerator|CodexInit.*Prompt|CodexOffer' -count=1 -v
--- PASS: TestCodexInitPromptIssuance (0.06s)
--- PASS: TestDefaultCodexInitGenerator (0.00s)
--- PASS: TestDoctor_GlamourCachePlaceholder (0.00s)
--- PASS: TestGlamourCacheIsWarn (0.00s)
--- PASS: TestGlamourStyleFromTheme_TokenMapping (0.00s)
--- PASS: TestGlamourStyleFromTheme_LightDarkDiffer (0.00s)
--- PASS: TestMarkdownRichEnabled_Matrix (0.00s)
--- PASS: TestGlamourRender_RichOutputStyled (0.00s)
--- PASS: TestSpecViewPlain_CommandUsesGlamourGateway (0.00s)
--- PASS: TestWriterIsTerminalRejectsDevNull (0.00s)
--- SKIP: TestRunVersionBranch_NonTTYProceeds (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.875s
```

마지막 SKIP 이 F1 이다.

### (3) statusline 「행동 결함 없음」 주장 — 성립

스크래치 사본에 임시 탐침 테스트를 두고, stdin 을 `/dev/null` 로 바꾼 상태에서 현재 `readStdinWithTimeout` 결과와 터미널 판별 방식이었다면 넘겼을 `*os.File` 을 각각 `statusline.ResolveConfigRoot` 에 넣었다:

```
zz_statusline_probe_test.go:26: current: returnsOsFile=false root="FALLBACK" replayLen=0 | term-style: root="FALLBACK" replayLen=0
ok  	github.com/modu-ai/moai-adk/internal/cli	0.848s
```

설정 루트와 재생 바이트가 동일하다. 판별은 틀리지만 결과는 같다는 판정서의 주장이 맞다. 실제 TTY 에서도 두 방식 모두 빈 reader 를 돌려주므로 차이가 없다.

### (4) Windows 에서 NUL 장치에 대한 `term.IsTerminal` — 원리상 정확, 실행은 미관측

`golang.org/x/term v0.45.0`(go.mod:30) 의 Windows 구현:

```go
func isTerminal(fd int) bool {
	var st uint32
	err := windows.GetConsoleMode(windows.Handle(fd), &st)
	return err == nil
}
```

NUL 은 콘솔 핸들이 아니므로 `GetConsoleMode` 가 실패해 `false` 가 된다. 교차 컴파일은 통과했다:

```
$ GOOS=windows GOARCH=amd64 go vet ./internal/cli/ && echo WIN_VET_OK
WIN_VET_OK
$ GOOS=windows GOARCH=amd64 go test -c -o /dev/null ./internal/cli/ && echo WIN_TESTCOMPILE_OK
WIN_TESTCOMPILE_OK
```

Windows 에서의 실제 실행 결과는 보지 못했다(미검증 항목 참조). 새 테스트는 `os.DevNull`(Windows 에서는 `NUL`)을 쓰므로 CI 의 Windows 매트릭스가 그대로 판정한다.

### (5) `internal/cli` 에 남은 `ModeCharDevice` 판별 — 형제 1곳 누락

```
$ grep -rn "ModeCharDevice" internal pkg cmd --include='*.go'
internal/contract/sign/defaults.go:58:  (주석 — 이미 term 방식)
internal/cli/statusline.go:170:         (항목 3 — 결과 동일, 수정 불요)
internal/cli/spec_status.go:301:        (주석)
internal/cli/codex_contract_test.go:93: (진단 이름 테스트 — TTY 판별 아님)
internal/cli/tty_devnull_test.go:9:     (주석)
internal/cli/update_version_test.go:679: ← 판정서 표에 없는 형제
internal/cli/codex_contract.go:155:     (파일 모드를 이름으로 바꾸는 진단 — TTY 판별 아님)
```

`update_version_test.go:679` 가 F1 이다.

## 발견 사항

- **F1** [Medium · 확신 높음] [optional — 별도 카드] `internal/cli/update_version_test.go:676-680` — 테스트 헬퍼 `isTerminalStdin` 이 `os.ModeCharDevice` 로 판별한다. `go test` 는 테스트 바이너리의 stdin 을 `/dev/null` 로 연결하므로(표준 입력을 파이프로 넘겨도 마찬가지) 헬퍼가 늘 `true` 를 내고, `TestRunVersionBranch_NonTTYProceeds`(AC-UVF-013)는 **항상 건너뛰어진다.** 스크래치 사본에서 헬퍼를 `isTerminalFile(os.Stdin)` 으로 바꿔 실제로 돌리면 테스트가 패닉한다 — 가려져 있던 두 번째 결함이다.

  ```
  $ echo x | go test ./internal/cli/ -run 'TestRunVersionBranch_NonTTYProceeds' -count=1 -v
      update_version_test.go:645: test requires non-TTY stdin; run via `go test` (pipe), not an interactive shell
  --- SKIP: TestRunVersionBranch_NonTTYProceeds (0.00s)

  (스크래치 사본, 헬퍼 교체 후)
  --- FAIL: TestRunVersionBranch_NonTTYProceeds (1.71s)
  panic: cannot create context from nil parent [recovered, repanicked]
  	…/internal/cli/update_version.go:343 +0x538
  	…/internal/cli/update_version_test.go:670 +0x340
  ```

  `update_version.go:343` 은 `context.WithTimeout(cmd.Context(), 5*time.Minute)` 이고, 테스트가 넘기는 `updateCmd` 에 컨텍스트가 없다. 운영 경로는 cobra 가 컨텍스트를 채우므로 사용자 영향은 없다고 추정하지만 관측하지는 않았다. 판정서의 형제 표가 이 위치를 빠뜨렸다.
  **필요한 수리(별도 카드):** 헬퍼를 `isTerminalFile(os.Stdin)` 으로 바꾸고, 테스트에서 `updateCmd.SetContext(context.Background())`(또는 동등한 방법)로 컨텍스트를 준 뒤 건너뛰기 없이 통과하는지 확인한다. 둘을 함께 고쳐야 한다 — 헬퍼만 고치면 적색이 된다.

- **F2** [Low · 확신 높음] [optional] `internal/cli/tty_devnull_test.go:14-18` — `defer f.Close()` 가 테스트 함수 반환 시 먼저 실행되고 `t.Cleanup` 의 `os.Stdin` 복원은 그 뒤에 실행된다. 그 사이 `os.Stdin` 이 닫힌 파일을 가리킨다. 테스트가 병렬이 아니므로 실해는 없다. **수리(선택):** 닫기도 `t.Cleanup` 으로 먼저 등록해 LIFO 로 복원이 먼저 일어나게 한다.

- **F3** [Low · 확신 중간] [optional] 패키지 안에 TTY 판별 관용구가 셋 공존한다 — `mattn/go-isatty`(`update_version.go:325`, `update_template_sync.go:850`, `init_update_notice.go:70`, `printer/printer.go:273`), `golang.org/x/term`(`spec_status.go`, `contract.go`, 이번 두 함수), `ModeCharDevice`(`statusline.go`). 앞의 둘은 모두 드라이버에 묻는 방식이라 `/dev/null` 판정은 같다. 이번 변경은 가장 가까운 기존 헬퍼를 재사용했으므로 올바른 선택이고, 통일은 범위 밖이다. **수리(선택):** 후속 정리 카드에서 한 관용구로 모을지 판단.

- **F4** [Info] statusline 판정서 주장은 항목 (3)의 탐침으로 확인됐다. 수정하지 않은 결정이 맞다.

## 증거 기준

- 이 감사에서 실행한 명령과 출력만 인용했다. 대상 트리는 `f5a9f536d`(감사 전후 `git rev-parse --short HEAD` = `f5a9f536d`, `git status --short` 출력 없음).
- 기준 판 재현은 `…/scratchpad/t1262-base`(HEAD 트리 + `422aa1e20` 판 두 파일)에서 측정.

## 미검증

- Windows 에서 새 테스트의 실제 실행 결과 — 교차 컴파일과 라이브러리 소스로만 확인. CI 매트릭스 판정 몫.
- 실제 TTY 에서 두 함수가 `true` 를 내는 경로 — 이 환경에 TTY 가 없어 확인 불가.
- `internal/cli` 전체 스위트 — 로컬 규율상 돌리지 않음. CI 판정 몫.
- F1 의 SKIP 이 CI 에서도 일어나는지 — 같은 `go test` 메커니즘이라 그럴 것으로 추정하지만 CI 로그는 보지 않음.
- 교차 모델 감사(codex·GLM) — 변경이 이미 커밋돼 미커밋 대상이 비어 있고, `baseBranch` 는 원격 기본 브랜치 기준이라 무관한 큰 diff 를 검토하게 되므로 실행하지 않음.

## 잔여 위험

- `statusline.go` 의 판별은 여전히 문자 장치를 터미널로 본다. 지금은 결과가 같지만, 이후 그 분기의 동작이 바뀌면 이번과 같은 결함이 드러날 수 있다.
- F1 이 고쳐지기 전까지 AC-UVF-013(비 TTY 다운그레이드 진행)은 자동 테스트로 보호받지 못한다.
