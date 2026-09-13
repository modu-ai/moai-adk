# t650 AS-1 구현 인계

## Claim

새 `internal/codexapp` 패키지에 Codex App Server stdio 연결과 managed 인증·계정·모델 조회 메서드를 구현했다. 기존 review 소비자 `internal/cli/mcp_codex.go`는 변경하지 않았다. 해당 코드의 exec/stdin/scanner/send-lock 방식을 참고했으나, review 전용 완료 판독과 느슨한 파싱을 이 패키지로 옮기지 않았다.

이 결과는 AS-1의 로컬 기반 구현 증거다. AS 전체 완료나 설치 완료를 뜻하지 않는다.

## Evidence

수정 전 실패 로그: `red.log`, `auth-red.log`, `bounds-red.log`, `lease-red.log`.

마지막 실행:

```text
go test -race ./internal/codexapp -count=1 -coverprofile=.moai/reports/t650/coverage.out
ok  github.com/modu-ai/moai-adk/internal/codexapp 2.469s coverage: 85.9% of statements

go tool cover -func=.moai/reports/t650/coverage.out
total: (statements) 85.9%
```

`go vet ./internal/codexapp`: exit 0, 출력 없음.

`GOOS=windows GOARCH=amd64 go test ./internal/codexapp -c -o /tmp/moai-t650-codexapp.test.exe`: exit 0, 출력 없음. 이는 Windows 런타임 검증이 아니다.

실제 설치 바이너리 `/Users/goos/.local/bin/codex`의 `codex --version`:

```text
codex-cli 0.154.0
```

실제 무인증 smoke 명령:

```text
CODEXAPP_LIVE_BINARY=/Users/goos/.local/bin/codex go test ./internal/codexapp -run TestInstalledAppServerNoAuthSmoke -count=1 -v
=== RUN   TestInstalledAppServerNoAuthSmoke
initialize_user_agent="moai_as1_probe/0.154.0 (Mac OS 26.6.2; arm64) unknown (moai_as1_probe; 0.1.0)" account=null requiresOpenaiAuth=true
--- PASS: TestInstalledAppServerNoAuthSmoke (0.06s)
PASS
ok  github.com/modu-ai/moai-adk/internal/codexapp 0.458s
```

새 private profile에서 Start → Initialize → Account → Close를 실행했다. 토큰 파일 읽기·가져오기, 로그인 요청, 모델 생성 요청을 하지 않았다.

## Baseline-attribution

트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t650`, 브랜치 `WT-gateway-appserver-core`, HEAD `9935e4e3e`. 편집 전 fetch 및 `origin/main...HEAD` 비교 결과 `0 3302`. 기존 tracked 파일 변경 없이 아래 새 파일만 작성했다.

- internal/codexapp/client.go
- internal/codexapp/client_test.go
- internal/codexapp/methods.go
- internal/codexapp/profile_unix.go
- internal/codexapp/profile_unix_test.go
- internal/codexapp/profile_windows.go
- .moai/reports/t650/ 아래 실행 증거와 이 보고서

## 구현 계약

- `Start(ctx, Config{Binary, Home, MaxMessageBytes, QueueSize, ExpectedConfigSHA256})`: 절대 경로, canonical private profile, 현재 사용자 소유권, 프로세스 수명 lock을 검증한다. cwd와 HOME/USERPROFILE/CODEX_HOME은 private profile을 가리킨다. 환경변수는 PATH·locale·Windows runtime·temp 목록만 허용하며 API key·토큰·프록시·기존 Codex 환경변수는 상속하지 않는다.
- 기존 config.toml은 기본 거절한다. 호출자가 생성한 설정을 사용할 때에는 private regular file과 정확한 `ExpectedConfigSHA256` 일치를 요구한다.
- Unix는 nonblocking flock, Windows는 LockFileEx를 사용한다. 자식 프로세스의 동일 profile lock 획득 실패를 실제 subprocess 시험으로 확인했다. Windows DACL은 현재 사용자 소유 및 현재 사용자/SYSTEM/Administrators 허용만 인정한다.
- `Call`, `Notify`, `Events`, `Respond`, `Err`, `Close`: 숫자·문자열 RPC ID를 검증하고 응답을 분배한다. 알림과 서버 요청은 한 ordered channel에서 전달한다. queue·pending 요청·메시지 크기를 제한하고, overflow 시 연결을 종료한다.
- 서버 요청 ID에 응답은 한 번만 가능하다. 개별 Call 취소는 정상 송신 완료 뒤 process를 종료하지 않는다. 쓰기 도중 취소는 송신 여부가 불명확하므로 연결을 종료한다. 서버 응답 완료와 turn 완료는 구별한다.
- EOF·잘못된 JSON·읽기 오류·크기/queue 초과를 구별한다. 서버 error의 raw message/data와 stderr를 오류 문자열에 노출하지 않는다. 사용자 인증 결과의 일회용 URL·code는 호출자에게만 반환하며 호출자가 로그에 남기면 안 된다.
- `Initialize`는 experimentalApi를 opt-in한다. `Login`은 browser/chatgptDeviceCode/apiKey 세 모드만 허용한다. 외부 구독 토큰 주입은 제공하지 않는다. `CancelLogin`, `Logout`, `Account`, `Models`를 제공한다. Account 결과는 이메일을 제외한다.

## Gaps

- 실제 브라우저/device/API key 로그인·로그아웃·갱신·모델 응답은 아직 실행하지 않았다. managed 인증 메서드는 fake subprocess 계약 시험이며, 실제 smoke는 무인증 조회만 수행했다.
- Start의 native 도구 비활성 설정은 설정 인수로 전달했다. 도구 inventory 및 shell/file/MCP/agent 실행 0은 아직 실증하지 않았다. 해당 판정은 독립적인 AS-1 실행 격리 게이트다.
- Windows ACL·LockFileEx·종료 동작은 GitHub CI 런타임 검증 대기다. macOS 실측을 Windows 성공으로 전용하지 않는다.
- LSP 진단을 실행하지 않았다. Go test와 vet의 컴파일·정적 검사만 실행했다. 85.9%는 현재 플랫폼 패키지 전체 수치이며 Windows 파일의 실행 커버리지가 아니다.
- dynamic tool HTTP 연계, thread mapping, compaction, Agent/fork, UI/picker, 설치는 AS-2 이후 범위다.
- 저장소 전체 테스트 verdict는 integration branch CI가 소유하며 현재 PENDING이다. 커밋·push·PR·설치를 하지 않았다.

## Residual-risk

App Server 실험 API는 설치 버전별로 달라질 수 있다. initialize 성공은 dynamicTools 전체 capability 검증을 대신하지 않는다. profile 파일 소유권과 digest 확인은 호출자가 만든 설정을 승인하는 경계이며 설정 내용의 기능 안전성을 자동 증명하지 않는다. 프로세스 PID의 종료·reap을 보장하는 연결 구현으로, 외부 자식 작업까지 시작하는 기능을 이 패키지에서 허용해서는 안 된다. 기존 review 경로는 손대지 않아 공통화 회귀가 없으나, 새 패키지의 독립 코드 검토는 오케스트레이터가 이어서 수행한다.


추가 독립 검토 반영: `TestProfileLeaseIsCloseOnExec`가 수정 전 `profile lease descriptor leaks across exec`로 FAIL한 것을 `cloexec-red.log`에 보존했다. Unix open에 O_CLOEXEC를 추가한 뒤 race 포함 패키지 테스트가 PASS했다. 이 시험은 F_GETFD/FD_CLOEXEC 비트의 실제 커널 값을 판정한다. 부모 crash 후 전체 process-tree 복구 E2E는 수행하지 않았다.

설정 파일의 hash 확인은 lifetime lease 획득 전 수행한다. 현재 패키지는 config를 수정하지 않는다. 이후 config를 쓰는 소유자가 추가되면 같은 profile lock 획득 뒤 작성·검증하는 프로토콜 또는 시작 전 재검증을 구현해야 한다. 현재 상태를 경쟁 config 쓰기 안전성으로 주장하지 않는다.
