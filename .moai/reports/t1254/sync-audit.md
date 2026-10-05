# t1254 독립 감사 — spec status 비-TTY 가드의 /dev/null 오판 수리

- 카드: t1254 (클래스 B · Tier S · SPEC 없음)
- 대상: 워크트리 `.claude/worktrees/t1254`, 브랜치 `WT-devnull-tty-guard`, HEAD `2fbe42c0a`
- 기준: develop `38148d891`
- 감사자: sync-auditor (평가 프로필 `default` — 필수 통과 차원: Functionality, Security)
- 판정: **FAIL** (차단 결함 1건, 수리는 한 줄 단위)
- 점수: Functionality 95 · Security 95 · Craft 60 · Consistency 75 → 조화평균 **78.4** (가중평균 85.0)

## 주장

1. 결함은 실재했고 수리로 사라졌다. 기준 커밋 바이너리는 `</dev/null` 입력에서 프롬프트를 띄우고 EOF 를 읽은 뒤 SPEC 을 "건너뜀(이미 완료)"으로 세며 exit 0 으로 끝난다. 수리 커밋 바이너리는 문서화된 비-TTY 오류로 exit 1 한다. 단위 수준과 CLI 끝단 양쪽에서 직접 재현했다.
2. 수리는 최소 범위다. `isTerminalFile(f)` 로 판정을 분리하고 `term.IsTerminal(int(f.Fd()))` 로 바꿨다. `golang.org/x/term` 은 기준 커밋 go.mod 의 직접 의존이므로 새 의존성이 아니다.
3. 기존 SpecStatus 15건, ExitCode 20건 테스트가 모두 통과한다. `stdinIsTerminalFn` 테스트 훅도 그대로 동작한다.
4. **그러나 새 테스트 파일이 프로젝트 린트를 깨뜨린다.** `golangci-lint run ./internal/cli/` 가 errcheck 3건을 보고한다. CI 의 lint 잡은 같은 설정(`.golangci.yml`, 예외 프리셋 없음)으로 전 패키지를 돌리므로 develop push 시 적색이 된다. 저자 판정서는 `go vet`·`gofmt` 만 근거로 들었고 golangci-lint 는 돌리지 않았다.

## 증거

### E1. 신규 테스트 (수리 후)

```
$ go test -count=1 -run TestIsTerminalFile -v ./internal/cli/
=== RUN   TestIsTerminalFileRejectsDevNull
--- PASS: TestIsTerminalFileRejectsDevNull (0.00s)
=== RUN   TestIsTerminalFileRejectsPipe
--- PASS: TestIsTerminalFileRejectsPipe (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.579s
exit=0
```

### E2. RED 독립 확인 — 기준 판정식 대 수리 판정식 (스크래치패드 프로브)

기준 커밋의 판정식(`git show 38148d891:internal/cli/spec_status.go` 295-301행, `fi.Mode() & os.ModeCharDevice`)을 그대로 옮긴 프로브를 저장소 밖에서 실행했다.

```
$ go run <scratchpad>/redprobe/main.go </dev/null
devnull old=true new=false
pipe    old=false new=false
stdin   old=true new=false
exit=0
```

`/dev/null` 과 `</dev/null` 로 연결된 stdin 에서 기준 판정식만 `true` 를 낸다. 파이프는 양쪽 모두 `false` 다. 저자가 주장한 원인(“문자 장치 = 터미널” 판정식)과 일치한다.

### E3. CLI 끝단 재현 — 기준 바이너리 대 수리 바이너리

픽스처: 스크래치패드의 임시 git 저장소, `main` 에 `feat: SPEC-FOO-001 land` 커밋 1개, `.moai/specs/SPEC-FOO-001/spec.md` 의 `status: in-progress`. 기준 바이너리는 `git archive 38148d891` 을 풀어 빌드했다.

기준 (`38148d891`):
```
$ ../moai-base spec status --sync-git </dev/null
  SPEC-FOO-001: in-progress → implemented
    Apply? [y/N]:     skipped SPEC-FOO-001

Summary: updated 0, skipped 1 (already done), 0 skipped (invalid status), 0 not found
exit=0
3:status: in-progress
```

수리 (`2fbe42c0a`):
```
$ ../moai-fix spec status --sync-git </dev/null
  SPEC-FOO-001: in-progress → implemented
   ERROR
  Spec status --sync-git: interactive confirmation unavailable in non-TTY context; pass --yes to auto-confirm.
exit=1
3:status: in-progress
```

`--yes` 대조 (수리 바이너리, 같은 `</dev/null`):
```
$ ../moai-fix spec status --sync-git --yes </dev/null
  SPEC-FOO-001: in-progress → implemented

Summary: updated 1, skipped 0 (already done), 0 skipped (invalid status), 0 not found
exit=0
3:status: implemented
```

기준 바이너리는 결함 그대로 조용히 건너뛰고, 그것을 "already done" 칸에 잘못 센다. 수리 바이너리는 오류로 중단하고 파일을 건드리지 않으며, `--yes` 경로는 여전히 기록한다.

### E4. 회귀 테스트

```
$ go test -count=1 -run SpecStatus ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	4.856s
exit=0
$ go test -list SpecStatus ./internal/cli/ | grep -c '^Test'
15

$ go test -count=1 -run ExitCode ./internal/cli/
ok  	github.com/modu-ai/moai-adk/internal/cli	29.316s
exit=0
$ go test -list ExitCode ./internal/cli/ | grep -c '^Test'
20
```

### E5. 정적 분석

```
$ go vet ./internal/cli/
vet=0
$ gofmt -l internal/cli/spec_status.go internal/cli/spec_status_tty_test.go
gofmt=0   (출력 없음)
$ GOOS=windows go vet ./internal/cli/
winvet=0
$ golangci-lint run ./internal/cli/        # v2.10.1
internal/cli/spec_status_tty_test.go:16:15: Error return value of `f.Close` is not checked (errcheck)
internal/cli/spec_status_tty_test.go:27:15: Error return value of `r.Close` is not checked (errcheck)
internal/cli/spec_status_tty_test.go:28:15: Error return value of `w.Close` is not checked (errcheck)
3 issues:
* errcheck: 3
lint=1
```

`.golangci.yml` 은 `errcheck` 를 켜고 예외 프리셋 없이 "Use the linter defaults; no custom exclusions." 로 둔다. CI 는 `.github/workflows/ci.yml:481` 에서 `golangci-lint run --timeout=5m` 을 돌리며, 444행 주석이 이 잡을 "always runs — fast, required check" 로 명시한다. `internal/cli` 전체에서 보고된 3건이 모두 신규 파일에 있으므로 기준 트리의 이 패키지는 린트가 깨끗했다.

저장소 관례:
```
$ grep -rn 'defer func() { _ = .*Close() }()' internal/cli/*_test.go | wc -l
      16
$ grep -rn '^\s*defer [a-z]*\.Close()$' internal/cli/*_test.go
internal/cli/spec_status_tty_test.go:16:	defer f.Close()
internal/cli/spec_status_tty_test.go:27:	defer r.Close()
internal/cli/spec_status_tty_test.go:28:	defer w.Close()
```

### E6. 변경 함수 커버리지

```
$ go test -count=1 -run TestIsTerminalFile -coverprofile=<scratchpad>/cov.out ./internal/cli/
$ go tool cover -func=<scratchpad>/cov.out | grep spec_status.go
spec_status.go:296:	stdinIsTerminal		0.0%
spec_status.go:302:	isTerminalFile		100.0%
```

### E7. 호출부·형제 판정식 전수

```
$ grep -rn 'stdinIsTerminal\|ModeCharDevice\|term\.IsTerminal\|isatty' internal pkg cmd | grep '\.go:'
(발췌)
internal/cli/spec_status.go:234:        if !stdinIsTerminalFn() {
internal/cli/spec_status.go:309:        var stdinIsTerminalFn = stdinIsTerminal
internal/cli/exitcode_streams_test.go:80: stdinIsTerminalFn = func() bool { return false }
internal/cli/contract.go:55-58:         contractStdinIsTerminalFn ... "instead of reusing stdinIsTerminalFn, whose os.ModeCharDevice test accepts the null device"
internal/cli/glamour_style.go:135,146:  writerIsTerminal "(mirrors spec_status.go stdinIsTerminal ...)" / ModeCharDevice
internal/cli/codex_init.go:118:         return err == nil && fi.Mode()&os.ModeCharDevice != 0
internal/cli/statusline.go:170:         if stdinFile.Mode()&os.ModeCharDevice == 0 {
internal/contract/sign/defaults.go:62:  return term.IsTerminal(int(os.Stdin.Fd()))
```

`stdinIsTerminal` 은 `stdinIsTerminalFn` 을 통해 234행 한 곳에서만 쓰인다. 다른 패키지나 호출부가 옛 `ModeCharDevice` 의미에 기대는 곳은 없다. 이번 변경으로 판정식이 `contract.go`·`internal/contract/sign` 의 기존 방식과 같아졌다.

### E8. Windows 동작 (추론)

`golang.org/x/term@v0.45.0/term_windows.go:17-21` 은 `windows.GetConsoleMode(windows.Handle(fd), &st)` 의 성공 여부로 판정한다. 콘솔 핸들이면 `true`, `NUL` 장치와 파이프는 `GetConsoleMode` 가 실패하므로 `false` 다. `os.DevNull` 은 Windows 에서 `"NUL"` 이므로 신규 테스트는 이식성이 있다. mintty/Git Bash 는 stdin 이 명명 파이프라 새 판정식이 `false` 를 내지만, 옛 판정식도 파이프에는 `ModeCharDevice` 가 서지 않아 `false` 였다. 따라서 그 환경의 동작은 달라지지 않는다. `GOOS=windows go vet` 통과(E5)로 컴파일 가능성까지만 확인했다.

## 기준 귀속

- 모든 측정은 이번 실행에서 HEAD `2fbe42c0a` 트리(워크트리 clean, 감사 중 HEAD 이동 없음)를 대상으로 했다. 기준 동작은 `git archive 38148d891` 로 추출한 트리를 스크래치패드에서 빌드한 바이너리와 기준 판정식 프로브로 쟀다.
- 도구: go (워크트리 go.mod), golangci-lint v2.10.1 (로컬). CI 는 v2.1.6 을 쓴다. `.golangci.yml` 주석이 두 버전의 린터 집합이 같다고 적어 두었지만, v2.1.6 으로 직접 돌리지는 않았다(미검증 참조).
- 저자 판정서(`.moai/reports/t1254/verdict.md`)의 수치는 근거로 쓰지 않고 위에서 다시 쟀다. 저자 수치(SpecStatus 15건, ExitCode 20건, RED 로그 내용)와 이번 측정 결과가 일치한다.

## 결함 목록

| id | 심각도 | 분류 | 위치 | 내용 | 필요한 수리 |
|---|---|---|---|---|---|
| F1 | High | **blocking** | `internal/cli/spec_status_tty_test.go:16,27,28` | errcheck 3건. CI 의 필수 lint 잡이 develop push 에서 적색이 된다(E5). 저자는 golangci-lint 를 돌리지 않았다. | 세 곳을 `defer func() { _ = f.Close() }()` 형태로 바꾼다(`r`, `w` 도 같게). 재측정: `golangci-lint run ./internal/cli/` → `0 issues` |
| F2 | Low | optional | `internal/cli/contract.go:55-57` | 주석이 "`stdinIsTerminalFn` 은 ModeCharDevice 로 판정해 널 장치를 통과시킨다"고 적는데, 이번 수리로 사실이 아니게 됐다. | 주석을 현재 상태에 맞게 고친다. 별도 시임(`contractStdinIsTerminalFn`)은 테스트 주입 용도라 유지해도 된다. |
| F3 | Low | optional (후속 후보) | `internal/cli/glamour_style.go:135,146` | "mirrors spec_status.go stdinIsTerminal" 주석이 더는 맞지 않는다. `writerIsTerminal` 은 stdout 이 `/dev/null` 이면 터미널로 보아 리치 렌더링을 켠다. 출력이 버려지는 스트림이라 실질 영향은 거의 없다고 본다(끝단 실행은 하지 않음). | 후속 카드로 넘길 만하다. 주석 정정과, 원하면 `isTerminalFile` 재사용. 우선순위 Low. |
| F4 | Medium | optional (범위 밖, 후속 후보) | `internal/cli/codex_init.go:116-118` | `defaultCodexPromptCapable` 이 같은 `ModeCharDevice` 판정식으로 **stdin 프롬프트 게이트**를 연다. 이 판정식이 `/dev/null` 에 `true` 를 내는 것은 E2 에서 쟀다. 따라서 `</dev/null` 에서 REQ-CI-009 의 비대화형 거절이 우회되고 프롬프트가 EOF 를 읽을 가능성이 높다. **가설 — 게이트를 끝단까지 돌려 보지는 않았다.** | 이번 카드와 같은 결함 계열이다. glamour 보다 먼저 볼 후속 카드로 권한다. |
| F5 | Info | optional | `internal/cli/statusline.go:170` | 같은 판정식. stdin 이 `/dev/null` 이면 `os.Stdin` 대신 빈 리더를 돌려주지만, 어느 쪽이든 읽을 데이터가 없어 동작 차이는 없다고 본다. | 조치 불필요. 기록만 한다. |
| F6 | Low | optional | `.moai/reports/t1254/verdict.md` | 판정서가 인용하는 `repro.log`·`after.log`·`specstatus.log`·`exitcode.log` 가 `.gitignore:235` (`.moai/reports/*`) 에 걸려 커밋되지 않았다. 브랜치에는 `verdict.md` 만 실린다. 판정서 본문에 핵심 줄이 인용돼 있어 피해는 작다. | 로그를 강제 추가하거나, 판정서에 원문을 그대로 옮긴다. |
| F7 | Info | optional | `internal/cli/spec_status.go:296` | 래퍼 `stdinIsTerminal` 의 커버리지가 0% 다(E6). 한 줄짜리 위임이라 테스트를 더할 가치는 낮다. | 조치 불필요. |

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | 카드 수용 기준 (1) RED→GREEN 을 단위(E1, E2)와 CLI 끝단(E3) 양쪽에서 확인. (2) `term.IsTerminal` 로 최소 수리. (3) SpecStatus 15건·ExitCode 20건 통과(E4). (4) 새 의존성 없음(go.mod 30행, 기준 커밋에도 같은 줄). |
| Security (25%) | 95 | PASS | 대화형 확인을 우회하던 입력 경로를 닫는 fail-closed 개선. 비밀·명령 실행·환경변수 추가 없음(diff grep 0건). Critical/High 없음. |
| Craft (20%) | 60 | FAIL | F1: 프로젝트 필수 린트 위반 3건. 변경 함수 커버리지 `isTerminalFile` 100%. 패키지 전체 커버리지는 재지 않았다. |
| Consistency (15%) | 75 | PASS (지적 있음) | 판정식이 저장소 안 기존 방식(`contract.go`, `internal/contract/sign`)과 같아졌다. 다만 `defer x.Close()` 가 저장소 관례(16곳)와 어긋나고, 두 주석(F2, F3)이 낡았다. |

필수 통과 차원(Functionality, Security)은 모두 통과했다. 전체 FAIL 은 차단 결함 F1 때문이다. F1 을 수리하면 재감사는 F1 한 건(`golangci-lint run ./internal/cli/` 0건 확인)으로 범위를 좁혀도 된다.

## 미검증

- CI 와 같은 golangci-lint v2.1.6 으로는 돌리지 않았다. 로컬 v2.10.1 결과와 `.golangci.yml` 의 버전 간 동일성 주장에 기대고 있다. v2 에서 errcheck 가 `defer f.Close()` 를 잡는 것은 로컬에서 관측했다.
- Windows 실행은 하지 않았다. E8 은 소스 판독과 `GOOS=windows go vet` 에 근거한 추론이다.
- 실제 터미널에서 대화형 프롬프트가 여전히 뜨는지는 관측하지 않았다. 이 세션의 stdin 은 TTY 가 아니다(E2 의 `stdin old=true new=false` 는 `</dev/null` 로 연결한 결과다).
- `internal/cli` 전체 스위트와 패키지 커버리지는 재지 않았다(CI 몫).
- F4 의 codex 게이트 우회는 판정식 수준에서만 쟀고 명령 끝단까지 돌리지 않았다.
- 교차 모델 감사(codex/GLM)는 돌리지 않았다. 변경이 이미 커밋돼 있어 `uncommittedChanges` 는 빈 diff 이고, `baseBranch` 는 원격 기본 헤드 기준이라 이 카드의 diff 가 아닌 것을 보게 된다.

## 잔여 위험

- F1 이 수리되지 않은 채 develop 에 병합되면 리드의 일괄 push 뒤 CI lint 잡이 적색이 되고, 같은 배치의 다른 카드 판정까지 가려진다.
- 같은 결함 계열(`ModeCharDevice` 로 TTY 를 판정)이 `codex_init.go` 에 남아 있다. 사용자 입장에서 같은 증상(비대화형 거절이 조용히 우회됨)이 다른 명령에서 재현될 수 있다.
- `term.IsTerminal` 은 mintty 같은 의사 터미널에서 `false` 를 내므로, 그 환경의 사용자는 `--yes` 없이는 이 명령을 쓸 수 없다. 옛 판정식도 같았다는 것이 추론이지만, 관측한 것은 아니다.
