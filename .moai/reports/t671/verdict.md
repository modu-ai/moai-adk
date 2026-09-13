# t671 — errcheck·staticcheck·unused 328건 수리 verdict

Baseline: HEAD `fac132d38` (= origin/develop), branch `WT-gateway-lint-sweep`, worktree `.claude/worktrees/t671`

---

## Claim

1. baseline에서 측정된 lint 이슈 328건(errcheck 296 · staticcheck 30 · unused 2, 76파일)을 전부 수리해 `golangci-lint run --timeout=5m`이 **exit 0 (0 issues)** 으로 끝난다.
2. 기계적 `_ =` 치환 금지 규칙을 준수한다: 쓰기·응답·영속성 경로의 오류는 실제로 전파하고, 실패가 결과에 영향을 줄 수 없는 경로만 사유 주석과 함께 `_ =`로 남긴다.
3. lint 수리로 인한 회귀가 없다 — 변경 대상 패키지 전부에서 테스트가 통과하고, GOOS=windows 빌드와 go vet이 통과한다.

## Evidence

### 최종 게이트 (이 트리, 이 러닝에서 직접 측정)

| 검사 | 명령 | 결과 |
|---|---|---|
| 전체 lint | `golangci-lint run --timeout=5m` | **exit 0, `0 issues`** — 전체 출력: `.moai/reports/t671/lint-green.log` |
| Windows 빌드 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (`WINDOWS-BUILD-OK`) |
| 정적 분석 | `go vet ./internal/...` | exit 0 |
| 포맷 | `gofmt -l internal/ cmd/ pkg/` | 빈 출력 (0개 파일) |
| 패키지 테스트 | `go test -count=1` 대상 패키지 일괄 (아래 Gaps 참조) | `.moai/reports/t671/test-green.log` |

### 커밋 (청크별)

| SHA | 범위 |
|---|---|
| `57b648c60` | internal/codexapp (ST1005 16건 소문자화 + errcheck) |
| `e6bdfa795` | internal/gateway 루트 18파일 (errcheck + QF1001) |
| `41f642dd1` | internal/gateway/auth 19파일 (errcheck + QF1003/QF1001/QF1008/SA1012 + unused) |
| `58118fccb` | internal/gateway/{conversation,receipt,translate,opaque} |
| `62fedfb82` | internal/cli 10파일 (errcheck + ST1013) |
| `265386815` | internal/{codexbridge,codextools,web,config} (ST1005 + errcheck + QF1001) |
| `0df3b4911` | 폐기 close 사유 주석 보강 (docs, 논리 변경 없음) |

### 수리 분류

- **실제 전파 (propagate)**: 저장소 잠금 해제·gateway 자식 핸들러 Close·auth store Close·fork receipt store Close 등을 named return에 결합. CLI stdout 출력 실패는 명령 오류로 반환. 테스트 셋업(os.WriteFile/os.Chmod/os.MkdirAll/json.Unmarshal)은 파일 스타일대로 `t.Fatal`/`t.Error`.
- **QF/ST quickfix**: ST1005 에러 문자열 소문자화(codexapp 16, codexbridge 4), ST1013 `http.StatusServiceUnavailable`, QF1001 De Morgan 5곳(양수 변수 추출 방식), QF1003 tagged switch, QF1007 조건부 병합.
- **unused**: `Store.platform` 필드와 POSIX `platformStore` 타입은 Windows 빌드(store_windows.go)가 참조하므로 제거 불가 — POSIX `openPlatformStore`가 `platform: &platformStore{}`를 설정해 참조로 만들어 해소(Windows 빌드로 검증).

### 의도적으로 `_ =`를 남긴 경우 (전부 사유 주석을 코드에 동반)

**원칙**: (a) Close가 항상 nil을 반환하도록 정의된 경우, (b) 오류 경로 폐기로 반환값이 이미 확정된 경우, (c) 잠금 해제가 Remove·process exit·unlock이 지배하는 경우, (d) 읽기 전용 소스(쓰기-back 없음), (e) 이미 사망한 프로세스/시작 실패 파이프, (f) once-guard로 의미 있는 호출이 별도로 오류를 단언하는 경우.

프로덕션 코드 (파일별):

| 위치 | 요약 사유 |
|---|---|
| `internal/cli/gateway_ui_state.go:44,125` | 읽기 전용 설정 소스, write-back 없음 |
| `internal/cli/migrate_cg.go:147,230,234,282` | 읽기 전용 소스 / 잠금 해제는 Remove+exit 지배 / 명시적 Close가 실오류 소유 |
| `internal/codexapp/client.go:134,151,156,157,192,269,270` | 생성 실패·폐기 파이프 / Wait 상태는 read-side fail(once) 뒤 불가 / 종료 경로 |
| `internal/codexapp/profile_unix.go` 2곳 | 임대 거절 경로의 fd 폐기 |
| `internal/gateway/anthropic.go` 8곳 | 업스트림 응답 본문 오류경로 폐기(클라이언트 오류는 확정) / nativeBody.Close는 항상 nil |
| `internal/gateway/openai.go` (JSON 경로 defer) | 본문 완전 읽음 |
| `internal/gateway/policy.go` (gzip defer) | 메모리 내 바이트 대상, flush 없음 |
| `internal/gateway/server.go` (프록시 defer) | 본문은 as-is 스트리밍 완료, `_, _ = io.Copy` 기존 관용과 일치 |
| `internal/gateway/supervisor_runner.go:67,74,166` | Serve가 리스너 소유 / cleanup은 once-guard + 명시 반환 경로 존재 / 성공경로 Close와 동일 관용 |
| `internal/gateway/auth/broker.go` 5곳 | 미기동·시작실패 파이프 폐기, teardown |
| `internal/gateway/auth/send.go:44,137,142,147,161,179` | GetBody 메모리 리더 / 접속 셋업 실패 폐기 / 미관측 응답 폐기 |
| `internal/gateway/auth/store.go:75,78,95,227,404` | 미획득 잠금 / 해제 경로 / 읽기 전용 / 스크래치 정리 |
| `internal/gateway/auth/store_posix.go` (defer Remove) | 원자적 replace가 live 파일 소유 |
| `internal/gateway/auth/lock_posix.go` (defer Close) | 읽기 전용 디렉터리 fd, Sync가 신호 소유 |
| `internal/gateway/receipt/store.go:136,194,224` | 잠금·읽기전용·임시 매니페스트 정리 |
| `internal/gateway/translate/stream.go` (defer Close) | context 훅이 이미 Close, once 성격 |

테스트 코드 (요약 — 코드에 주석 동반):

- `io.Pipe`·`os.Process.Release` Close — 항상 nil 반환 정의.
- `nativeBody.Close` 테스트 폐기 — 항상 nil.
- 서버 측 `r.Body.Close()` 폐기, 미독 프로브 응답 본문 — 단정은 클라이언트 측 관측이 소유.
- `process.Kill`·`cmd.Process.Kill` — 이미 종료된 프로세스가 기대 상황.
- exec-parent 헬퍼(`supervisor_exec_test.go`)의 `child.Stop` — 헬퍼는 stdout이 프로토콜 채널이라 오류 보고 수단이 exit code뿐.
- 브로커 자식 프로세스(`broker_regression_test.go`, `codexapp/client_test.go` 헬퍼)의 프로토콜 쓰기 실패는 `os.Exit(3)`으로 보고( nonzero exit = 부모의 ErrBroker ).

**nolint 2건** (빠른fix가 옳지 않은 경우, 사유 주석 동반):

- `send.go:199` `//nolint:staticcheck // QF1008: c.Write would recurse into headerConn.Write` — 임베딩 셀렉터 제거는 무한 재귀를 만들므로 명시적 `c.Conn.Write`가 의도.
- `credential_anthropic_oauth_test.go` `//nolint:staticcheck // SA1012` — nil Context 거부 계약 자체가 테스트 대상이므로 `context.TODO()` 치환은 커버리지 소실.

## Baseline-attribution

- 모든 수치는 이 워크트리(`.claude/worktrees/t671`)에서 `git rev-parse --short HEAD` = `fac132d38` 상태에서 측정한 `.moai/reports/t671/lint-baseline.log`(328행 이슈)와 대비, 동일 트리의 최종 커밋 `0df3b4911`에서 `golangci-lint run --timeout=5m`을 재실행하여 확인(exit 0, `0 issues` — `.moai/reports/t671/lint-green.log`). 커맨드+출력은 위 Evidence 표에 귀속.
- 청크별 중간 검증: 각 패키지 커밋 전 `go test -count=1 ./internal/<pkg>/` + `golangci-lint run ./internal/<pkg>/...` + `gofmt -l`을 해당 시점 트리에서 직접 실행(본문에 경과 수치 재인용 없음).

## Gaps

다음은 관측하지 못했거나, 이 환경에서 관측이 불가능한 항목:

1. **`go test ./...` 전체 로컬 미실행** — 레인 규율(전체 스위트는 CI 몫)에 따라 변경 영향 패키지(cli, codexapp, gateway/…, codexbridge, codextools, web, config)만 실행. 전 패키지 판정은 develop push 후 CI가 담당.
2. **python3 하위프로세스 픽스처 테스트 5개가 이 머신에서 실패 — baseline 재현 확인된 환경 의존 실패(내 수정 무관)**:
   - `internal/gateway` `TestAppServerSubprocessHTTPToolContinuation`, `internal/codexbridge` `TestAuditRealTransportEOFWakesBridge`, `TestAuditCancelBeforeRealStartReplyInterruptsTurn`, `TestCanceledStartBeyondReplyDeadlineStillInterrupts`, `TestAuditCanceledRPCDoesNotExhaustSharedTransport`
   - 재현·기제: 테스트가 `exec.LookPath("python3")` 결과(= `/Users/goos/.pyenv/shims/python3`, **bash 스크립트**)를 shebang 인터프리터로 쓰는 실행 스크립트를 만들고 `codexapp.Start`로 exec → macOS 커널은 스크립트를 인터프리터로 하는 shebang exec를 거부(`exec format error`) → `cmd.Start()` 실패 → "app server start failed". `/tmp` 실험 프로그램으로 동일 기제 독립 재현 완료. baseline `fac132d38`에서도 동일 환경이면 동일하게 실패하는 구조(실패 지점은 내 수정이 만지지 않는 `cmd.Start()` 경로).
   - 해당 5개를 `-skip`한 나머지 전부 통과(`.moai/reports/t671/test-green.log`).
3. **테스트가 수정한 파일의 통합 테스트 스위트 상호작용**(예: `-race` 전수)은 미측정 — CI 스위트 몫.
4. 내 논리 실수 1건을 중간에 저질렀고 수정했다: `gateway_ui_state_test.go`에서 "파일 부재 시 nil state도 유효"한 원래 테스트 의도를 `json.Unmarshal` 엄격화로 깼다가(첫 cli 전량 테스트에서 FAIL 1건), 파일 부재를 허용하고 malformed만 실패시키는 형태로 바로잡았다. 바로잡은 뒤 `TestGatewayUIState*` 전부 통과.

## Residual-risk

- **t695 병합 충돌 가능성**: `internal/gateway/translate/**`(7파일, 소규모 준수 수정)과 `native_policy.go` QF1001 2곳은 t695의 기능 수정과 다른 라인이지만, 같은 파일 병합 시 수동 개입이 필요할 수 있다. 병합 후 그 트리에서 `golangci-lint run` 재측정 권장.
- **`_ =` 잔여 목록의 판단 의존성**: (b)·(c)·(e)류는 "반환값이 이미 확정"이라는 현 구조에 대한 판단이므로, 향후 해당 함수의 오류 전파 구조가 바뀌면(예: fail-once 제거) 일부 주석이 낡을 수 있다.
- **테스트가 t.Error로 바뀐 cleanup Close들**(`t.Cleanup` 내 store/client Close): 머신 고부하 시 3초 타임아웃류 정리 실패가 새로 테스트 실패로 드러날 수 있다 — 이는 기존에 조용히 사라지던 실제 신호의 노출이며, flake가 아니라 지연의 가시화다.
- **ST1005 소문자화로 에러 문자열이 바뀜**(`codex app server ...`, `app server ...`): sentinel 비교만 쓰는 현재 코드베이스에는 문자열 매칭 의존이 없음을 grep으로 확인했으나, 외부 스크립트가 stderr 문자열을 파싱한다면 영향 가능(내부 CLI이므로 가능성 낮음).

🗿 MoAI
