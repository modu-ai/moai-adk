# t199 — internal/web 자기-SIGTERM TOCTOU

**카드**: t199 (Class B, Tier S) — #1602 Release Verify(ubuntu) FAILURE 차단 해소
**베이스**: `origin/release/v3.1.3` @ `63d51b453`
**브랜치**: `WT-web-sigterm-toctou` (worktree `.claude/worktrees/t199`)

---

## 1. Claim

`internal/web`의 `ListenAndServe`가 리스너를 먼저 바인드하고 시그널 등록을 나중에 해서, 바인드 관찰을 신호로 삼는 호출자가 등록 전 창에 SIGTERM을 보내면 프로세스가 기본 처분으로 즉사한다. 등록을 바인드 앞으로 옮기면 창이 사라진다.

## 2. Evidence

### 2.1 원인 재현 (RED)

리드 진단은 창의 존재를 지목했고, darwin에서는 창이 너무 좁아 발현되지 않았다. 창을 300ms로 넓혀 결정론적으로 재현했다 — 프로브를 옛 등록 지점 바로 앞에 삽입:

```
225:	time.Sleep(300 * time.Millisecond) // TOCTOU-WINDOW-PROBE
```

```
$ go test ./internal/web/ -run TestServer_GracefulShutdownOnSIGTERM -count=1 -v
=== RUN   TestServer_GracefulShutdownOnSIGTERM
signal: terminated
FAIL	github.com/modu-ai/moai-adk/internal/web	0.603s
FAIL
```

CI 증상과 형태가 일치한다: `signal: terminated`, **`--- FAIL` 라인 없음** — 테스트가 실패한 게 아니라 테스트 바이너리가 죽는다. 이것이 run 32643066279 / job 97203064509 의 `FAIL internal/web 30.294s` 와 같은 실패 모양이다.

darwin 미발현은 결함 부정이 아니라 창이 좁다는 뜻이었고, 창을 넓히자 darwin에서도 그대로 재현됐다.

### 2.2 조치

`internal/web/server.go` — 13줄 추가 / 4줄 삭제:

```diff
 func (s *Server) ListenAndServe(ctx context.Context) error {
+	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
+	defer stop() // also releases the registration on the bind-failure return below
+
 	if err := s.bind(); err != nil {
 		return err
 	}
@@
-	// Signal-aware context: cancel on SIGINT/SIGTERM (REQ-WC-003).
-	sigCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
-	defer stop()
-
```

`defer stop()`이 등록 직후에 서므로 바인드 실패 조기 return 경로에서도 등록이 새지 않는다. `s.app.triggerShutdown = func() { stop() }` 배선, drain 로직, `sigCtx` 소비 지점은 모두 무변경 — REQ-WC-003 5초 drain invariant 보존.

시그널 테스트 코드는 카드 지시대로 유지했다.

### 2.3 AC 판정

| AC | 결과 | 근거 |
|---|---|---|
| (1) go vet + build 통과 | PASS | `go build ./internal/web/` `go vet ./internal/web/` 무출력. 교차: `GOOS=linux go vet` ok, `GOOS=windows go vet` ok. `golangci-lint run ./internal/web/` → `0 issues.` |
| (2) darwin `go test -race -count=10 ./internal/web/` 초록 | **부분** — 대상 테스트는 PASS, 패키지 전체는 무관한 선행 결함으로 FAIL (§3) | `-run TestServer_GracefulShutdownOnSIGTERM -count=10 -race` → 10/10 PASS. 패키지 `-count=1 -race` → `ok 15.075s` |
| (3) NotifyContext의 net.Listen 선행을 잠그는 회귀 테스트 (문자열 grep 금지) | PASS | `internal/web/server_signal_order_test.go` — go/ast 파싱 구조 검증 |
| (4) release/v3.1.3 병합·push 후 #1602 Release Verify(ubuntu) 초록 | 통합 후 리드 판정 | 최종 판정은 CI |

### 2.4 회귀 테스트 (AC-3)

`server_signal_order_test.go`, 문자열 grep 아님 — `go/parser`로 `server.go`를 파싱해 AST 노드 위치를 비교한다:

- `TestListenAndServeRegistersSignalsBeforeBinding` — `ListenAndServe` 본문에서 `signal.NotifyContext` 호출 노드가 `s.bind()` 호출 노드보다 앞서는지 위치로 단언. 이름 변경·리포맷은 오탐을 내지 않고, 순서를 되돌리면 잡힌다.
- `TestBindIsTheOnlyListenSite` — 순서 테스트가 딛고 선 전제를 고정: `net.Listen`이 `bind` 한 곳에서만 호출된다. 다른 경로가 생기면 순서 가드가 그 경로를 못 덮으므로 이 전제가 먼저 깨져야 한다.

행위 테스트로는 이 창을 덮을 수 없다 — 창 안에서는 테스트가 실패하는 게 아니라 프로세스가 죽어 판정 자체가 사라진다.

**RED baseline**: 수정 전 `server.go`(= `63d51b453` 판본)로 되돌려 실행하면 실패한다 —

```
--- FAIL: TestListenAndServeRegistersSignalsBeforeBinding
    server_signal_order_test.go:49: signal.NotifyContext at server.go:226:18 runs AFTER
    s.bind() at server.go:207:12; a SIGTERM arriving between the bind and the
    registration kills the process
```

수정 후:

```
$ go test ./internal/web/ -run 'TestListenAndServeRegistersSignalsBeforeBinding|TestBindIsTheOnlyListenSite' -count=1 -v
--- PASS: TestListenAndServeRegistersSignalsBeforeBinding (0.00s)
--- PASS: TestBindIsTheOnlyListenSite (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.534s
```

### 2.5 GREEN

```
$ go test ./internal/web/ -run TestServer_GracefulShutdownOnSIGTERM -count=10 -race
--- PASS: TestServer_GracefulShutdownOnSIGTERM  (×10)
ok  	github.com/modu-ai/moai-adk/internal/web	1.862s

$ go test ./internal/web/ -count=1 -race
ok  	github.com/modu-ai/moai-adk/internal/web	15.075s
```

## 3. Gaps — AC-2 패키지 전체 `-count=10`은 초록이 아니다 (선행 결함, t199 무관)

`go test -race -count=10 ./internal/web/` 는 실패한다. 원인은 t199가 아니다:

```
--- FAIL: TestProfileRename/ok
    profile_bar_test.go:197: rename status = 400, want 200
```

10회 중 **9회** 실패(1회차만 통과). 귀속 측정 — 같은 트리에서 내 변경만 걷어내고(신규 테스트 파일 이동 + `git checkout -- internal/web/server.go`) 동일 명령 실행:

| 트리 | `--- FAIL: TestProfileRename/ok` 건수 |
|---|---|
| t199 수정 적용 | 9 / 10 |
| 베이스 `63d51b453` (수정 없음) | **9 / 10** |

동일 — 선행 결함이며 t199 변경과 무관하다. 시그널 경로와 접점도 없다(`httptest` 기반 프로파일 CRUD 테스트, `ListenAndServe` 미경유).

**추정 원인(미검증 가설)**: `TestProfileRename/ok`이 `scratch` → `scratch2` 리네임 후 `profile.GetBaseDir()`를 그대로 두는데, 이 베이스 디렉터리가 프로세스 전역이라 한 바이너리 안의 2회차부터 `scratch2`가 이미 존재해 400이 된다. CI는 `-count=1`이라 이 실패를 본 적이 없다. → **후속 카드 후보** (t199 범위 밖 — 카드 지시가 명시한 조치는 시그널 등록 순서 하나뿐이고, 프로파일 테스트 격리는 별건).

## 4. Baseline-attribution

- 트리: `63d51b453` (`origin/release/v3.1.3`) + 본 카드 변경 2파일
- 모든 측정은 이 워크트리(`.claude/worktrees/t199`)에서 이번 실행 중 관측한 출력
- 플랫폼: darwin/arm64. ubuntu 실행 결과는 관측하지 않았다 — AC-4는 통합 후 CI 몫

## 5. Residual-risk

- **ubuntu 미관측**: 재현·검증 모두 darwin. 창을 없앤 변경이므로 플랫폼 의존이 없어야 하지만, ubuntu 초록은 #1602 CI가 최종 판정한다.
- **다른 자기-시그널 경로**: 이번 가드는 `ListenAndServe` 한 함수의 순서만 잠근다. 다른 곳에서 리스너를 여는 경로가 생기면 `TestBindIsTheOnlyListenSite`가 먼저 깨지도록 해뒀지만, 시그널을 쓰는 다른 서버가 생기면 그 경로는 별도 가드가 필요하다.
- **바인드 실패 창**: 등록 이후 바인드 실패로 조기 return 하는 사이에 도착한 SIGTERM은 `stop()`으로 등록이 풀리며 흡수된다(수 마이크로초). 그 경로는 어차피 오류를 들고 종료하므로 실질 영향 없음.
- **선행 결함 잔존**: §3의 `TestProfileRename` 격리 결함은 고치지 않았다 — 범위 밖. 남겨두면 누군가 `-count>1`로 이 패키지를 돌릴 때마다 다시 걸린다.
