# t613 run — 전 라우트 Host 게이트

- 카드: t613 · SPEC: SPEC-WEB-CONSOLE-001 (REQ-WC-009 / AC-WC-009 개정본, 2026-09-10)
- 단계: run (cycle_type=tdd) · 브랜치: `WT-web-host-read`
- 기준 HEAD: `9db13c070` → RED 커밋 `4a81fb5dc` → GREEN 커밋(이 문서를 싣는 커밋)
- 운영자 결정: 판정서 §5 A안. Host 검사는 메서드·라우트와 무관하게 적용하고, Sec-Fetch-Site 검사는 POST/PUT/PATCH 에만 남긴다. `isLoopbackHost` 와 `glmkey.go` 는 건드리지 않는다.
- 적용 규칙: `verification-claim-integrity.md` §1·§2·§3(주장과 증거 분리, 측정 트리 귀속), §2.3(RED 를 별도 커밋으로 먼저 남겨 순서를 커밋 그래프로 증명), `verification-completeness.md` §1.1(변이본으로 실패를 관측해야 검사가 완성된다)

## 1. Claim

1. `hostCheckMiddleware` 가 이제 모든 메서드, 모든 라우트(`/static/` 포함)에서 loopback 이 아닌 Host 와 Host 부재를 403 으로 막는다. Sec-Fetch-Site 검사는 상태 변경 메서드에만 걸린다.
2. AC-WC-009 개정본의 대조군 행렬(Host 8행 × 정적 GET·동적 GET·POST 3열)이 테스트로 고정됐고, 허용 행은 200/2xx 에 저장까지, 거부 행은 403 에 내용·저장 없음으로 판정된다.
3. 등록된 읽기 라우트 8개(`/`, `/kanban`, `/monitor`, `/todo`, `/specs`, `/settings`, `/events`, `/static/app.js`)가 외부·부재 Host 의 GET·HEAD·OPTIONS·DELETE 에 403 을 돌려준다.
4. 실제 `net/http` 서버 경로에서도 외부·부재 Host GET 은 403 이고 프로필 값이 실리지 않으며, 외부·부재 Host POST 뒤에 프로필 설정과 `user.yaml`·`language.yaml`·`statusline.yaml` 은 바이트 단위로 그대로다.
5. `ssh -L` 을 본뜬 프로세스 내부 TCP 중계기를 거치면 `localhost:<중계 포트>` 요청은 200, 같은 중계기로 보낸 외부 Host 요청은 403 이다.
6. 새 테스트와 뒤집은 테스트는 공허하지 않다. 게이트를 원래 모양으로 되돌리거나(m1), `/static/` 을 빼거나(m2), HEAD·OPTIONS 를 빼면(m3) 해당 칸이 실패한다.

## 2. Evidence

### 2.1 RED — 바뀌지 않은 `app.go` 에서 (`4a81fb5dc` 트리, `app.go` 는 `9db13c070` 과 동일)

```
$ GOMAXPROCS=2 go test -p 1 -count=1 -timeout 600s ./internal/web/... > .moai/reports/t613/run/red.txt 2>&1
exit=1
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/web	23.572s
```

- 최상위 실패 11개: 새 테스트 5개(`TestHostGateControlMatrix`, `TestHostGateCoversEveryRouteAndMethod`, `TestHostCheckGatesGetFromForeignHost`, `TestGoldenPath_ReadWriteRoundTrip` 확장분, `TestSSHTunnelModel_LoopbackHostThroughForwarder`)와 뒤집은 정적 자산 테스트 6개.
- 실패 사유는 모두 403 기대에 200 이 온 것이거나, 403 이어야 할 응답에 내용이 실린 것이다. 대표 줄: `integration_test.go:142: foreign-Host GET status = 200, want 403`, `integration_test.go:334: foreign-Host GET through the forwarder: status = 200, want 403`.
- 행렬에서 실패한 칸은 `foreign`·`absent` × `static`·`settings` 네 칸뿐이다. POST 칸은 기존 게이트가 이미 막고 있어서 통과했다.
- 전 라우트 테스트의 거부 칸 64개(8 라우트 × 4 메서드 × 2 Host)가 모두 실패했다.
- 판정 가능성 확인: loopback 대조 실패 0, panic/race 0, 빌드 실패 0. 중계기 경유 (a) 요청과 연결 수 단언은 RED 에서도 통과했다. 실패는 외부 Host 칸에서만 났다.

### 2.2 GREEN

```
$ GOMAXPROCS=2 go test -p 1 -count=1 -timeout 600s ./internal/web/... > .moai/reports/t613/run/green.txt 2>&1
exit=0
ok  	github.com/modu-ai/moai-adk/internal/web	17.202s
```

### 2.3 Lint (`.moai/reports/t613/run/lint.txt`)

```
$ go vet ./internal/web/...           → exit=0, 출력 없음
$ golangci-lint run ./internal/web/... → 0 issues. exit=0
$ gofmt -l internal/web               → 출력 없음
```

변경 전 기준선도 같은 트리(`9db13c070`)에서 쟀다. `golangci-lint` 0 issues, `gofmt -l` 출력 없음. 새로 생긴 경고는 없다.

### 2.4 변이본 (`.moai/reports/t613/run/mutants.txt`)

추적 파일은 고치지 않고 `go test -overlay` 로만 주입했다. 변이 소스는 세션 스크래치 영역에 두었다.

| 변이본 | 게이트 모양 | exit | 실패한 최상위 테스트 | 실패한 거부 칸 |
|---|---|---|---|---|
| m1 | 원래대로 POST/PUT/PATCH 안에서만 Host 검사 | 1 | 11 (RED 와 같은 집합) | 64 (전 라우트·전 메서드) |
| m2 | `/static/` 경로를 게이트에서 제외 | 1 | 8 (행렬, 전 라우트, 정적 자산 6개) | 전 라우트 테스트의 `/static/app.js` 8칸 + 행렬의 `foreign`·`absent` static 2칸 |
| m3 | HEAD·OPTIONS 를 게이트에서 제외 | 1 | 1 (`TestHostGateCoversEveryRouteAndMethod`) | HEAD·OPTIONS 칸만 32, DELETE 는 여전히 403 |

세 변이본 모두 빌드 실패 0, loopback 대조 실패 0 이다. 변이본이 통과한 경우는 없다.

### 2.5 커밋 전 투명 문자 검사

RED 커밋 직전의 스테이징 diff(199,371자)에서 유니코드 Cf 범주 문자 0개. 같은 스크립트에 `chr(0x200B)` 를 넣은 대조 문자열은 1개로 셌다. GREEN 커밋 직전 값은 커밋 보고에 따로 적는다.

### 2.6 `isLoopbackHost` 허용 범위 프로브 (`.moai/reports/t613/run/loopback-probe.txt`)

추적 파일은 고치지 않고, 입력 13개를 `isLoopbackHost` 에 넣어 결과만 기록하는 테스트를 오버레이로 주입했다.

```
$ GOMAXPROCS=2 go test -count=1 -timeout 120s -overlay <scratch>/probe_overlay.json -run TestT613LoopbackProbe -v ./internal/web/
exit=0
--- PASS: TestT613LoopbackProbe (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/web	0.833s
```

`PROBE` 줄 13개가 찍혔다(빈 스윕이 아님). 값은 §6 첫 항목에 옮겼다.

## 3. Baseline-attribution

| 측정 | 트리 | 코드 상태 |
|---|---|---|
| lint 기준선 | `9db13c070` | 변경 전 |
| RED | 작업 트리 = `4a81fb5dc` 내용(테스트만 변경, `git diff -- internal/web/app.go` 빈 출력) | 원래 게이트 |
| GREEN, lint, 변이본 | `4a81fb5dc` + `app.go` 수정분(GREEN 커밋 내용) | 전 라우트 게이트 |

판정서 §4 의 "79개 실패"는 `d3b7d438d` 에서 잰 값이다. 이번 회차에서는 다시 세지 않고, 그 목록을 호출 지점 분류의 출발점으로만 썼다(아래 §4).

## 4. 테스트 분류

| 분류 | 개수 | 처리 | 내역 |
|---|---|---|---|
| 1. 의도적 계약 — 외부 Host GET 성공을 요구하던 테스트 | 7개 테스트 | 외부 Host 403 을 요구하도록 뒤집고, loopback 에서 200·본문·Content-Type 검사는 그대로 유지 | `TestStaticAssetsServedFromEmbed`, `TestHtmxEmbeddedAndServed`, `TestFontServedFromStatic`, `TestMascotPosesEmbeddedAndServed`, `TestCJKFontServedFromStatic`, `TestI18nDictionaryEmbedded`, `TestHostCheckDoesNotGateGet` → `TestHostCheckGatesGetFromForeignHost` (옛 이름은 `internal/web` 에 남아 있지 않음) |
| 2. 외부 Host POST 거부 — 회귀 대조군 | 4개 단언 | 변경 없음, GREEN 에서 통과 | `TestHostCheckRejectsForeignHostOnPost`, `TestShutdown_ForeignHostRejected403`, `glm_tier_test.go` "non-loopback host refused", 골든 패스의 외부 Host POST |
| 3. 부수적 — Host 없이 만든 요청(`example.com` 기본값) | 13개 파일, 22개 호출 지점 | `serveGet` 헬퍼로 loopback Host 이전 | 헬퍼 6개(`todoBodyFor`, `getBoard`, `renderAppBody`, `getTodoPage`, `kanbanBodyFor`, `renderConsolePage`)를 고쳐 판정서 목록의 나머지 72개 테스트가 한꺼번에 해결됨 |
| 3a. 그중 게이트 아래서 조용히 공허해질 뻔한 테스트 | 5개 | loopback 으로 옮겨 원래 판정을 되살림 | 판정서 79개 목록에 없던 것들. 아래 §4.1 |
| 4. 라우팅되지 않는 POST 파싱 테스트 | 8개 호출 지점 | 해당 없음 | `parseSchemaForm`·`bindForm`·`ParseForm` 만 호출하고 라우터를 타지 않는다(`crosssession`, `widget_policy` ×2, `write_safety` ×3, `statusline`, `i18n`) |
| 5. 신규 | 테스트 4개 + 골든 패스 확장 | — | `TestHostGateControlMatrix`(서브테스트 24), `TestHostGateCoversEveryRouteAndMethod`(거부 64 + loopback 대조 40), `TestHostGateDoesNotWidenFetchSiteCheck`, `TestSSHTunnelModel_LoopbackHostThroughForwarder` |

### 4.1 판정서 목록 밖에서 새로 찾은 것

게이트를 모든 메서드로 넓히면 실패하지 않고 오히려 **통과해 버리는** 테스트가 있었다. 403 본문이 단언을 공허하게 만족시키기 때문이다.

- `TestMascotRetiredAssetsRemoved` — "200 이 아니면 통과"라서 403 도 통과한다. loopback 으로 옮겨 실제 404 를 받게 했다.
- `TestTodoNavRowNotCurrentElsewhere` — 문자열 부재만 본다. loopback 으로 옮기고 200 확인을 한 줄 더했다.
- `TestAutonomyStubResolved` 의 두 번째 요청 — 조각 마커의 부재만 본다. loopback 으로 옮겼다.
- `TestConsoleRoutesLeaveBacklogUntouched` — 상태 코드를 보지 않아 403 이면 라우트를 하나도 실행하지 않은 채 통과한다. loopback 으로 옮겼다.
- `TestTodoRouteRejectsNonGET` 첫 번째 루프 — 403 이 CSRF 게이트에서 왔다고 주장하는데, Host 가 기본값이면 Host 게이트가 먼저 막는다. loopback Host 를 넣어 귀속을 바로잡았다.

재검색(`evil.example`, `attacker.example`, `not Host-gated`, `foreign`, `.Host = `)으로 찾은 외부 Host 단언은 카드가 적어 준 목록과 같았다. 목록 밖에 하나 더 있다. `coverage_test.go` 의 `TestIsLoopbackHost` 는 `isLoopbackHost` 단위 표로, 맨 `::1` 을 허용으로 기대한다. 이번 변경과 무관해서 건드리지 않았다.

## 5. Gaps

- **실제 `ssh -L` 은 재지 않았다.** sshd 를 거친 측정은 없고, 바이트를 그대로 옮기는 프로세스 내부 중계기 모델만 재었다. 실제 sshd 가 Host 를 바꾸지 않는다는 점은 모델의 전제일 뿐이다.
- **실제 브라우저 DNS rebinding 공격은 수행하지 않았다.** 서버가 외부·부재 Host 를 막는다는 것까지만 확인했다.
- 404 경로(등록되지 않은 경로)와 `/profile/*`·`/glm-key/reveal`·`/__shutdown__` 의 외부 Host GET 은 따로 단언하지 않았다. 게이트가 라우팅보다 앞에 있어 같은 분기를 타지만, 칸 단위 관측은 없다.
- 커버리지 수치는 재지 않았다.
- 변이 소스와 오버레이 JSON 은 스크래치 영역에 두었고 보고서 디렉터리로 옮기지 않았다. 게이트 줄과 실행 요약은 `mutants.txt` 에 있다.
- 이 문서는 `moai-domain-humanize` 최종 교정을 거치지 않았다.

## 6. Residual-risk

- **허용 범위와 SPEC 문구가 어긋난다.** 개정된 REQ-WC-009 는 "정확히 `localhost`, `127.0.0.1`, `[::1]`"만 받는다고 적는데, 실제 `isLoopbackHost` 는 더 넓다. 오버레이 프로브로 잰 값(`.moai/reports/t613/run/loopback-probe.txt`, GREEN 트리):
  - 받음: `127.1.2.3:3041`, `127.255.255.254`(= `127.0.0.0/8` 전체), `[::ffff:127.0.0.1]:3041`·`::ffff:127.0.0.1`(IPv4-mapped loopback), 대괄호 없는 `::1`
  - 거부: `LOCALHOST`, `Localhost:3041`, `localhost.`(대소문자·끝 점 구분), `0.0.0.0`, `attacker.example.com`, 빈 값

  넓은 쪽은 모두 loopback 주소라 rebinding 방어의 구멍은 아니다. 다만 문구와 코드 중 하나는 맞춰야 한다. 운영자 지시대로 코드는 바꾸지 않았다. `isLoopbackHost` 의 godoc("Accepts 127.0.0.1, localhost, and ::1")도 같은 이유로 실제보다 좁게 적혀 있다.

  > **후속 회차 반영(§8):** `LOCALHOST`·`Localhost:3041` 거부는 이 카드가 만든 회귀로 판정돼 고쳤다(이제 받음). 넓은 허용 범위는 그대로 두되 단위 표에 행으로 고정했고, godoc 은 실제 허용 집합으로 다시 썼다. SPEC 문구와의 어긋남은 여전히 남아 있다.
- **이름 비교가 대소문자를 구분한다.** 이전에는 이 거부가 POST 에만 걸렸지만 이제 GET 에도 걸린다. 브라우저는 호스트 이름을 소문자로 보내므로 일반 사용에는 영향이 없을 것으로 보지만, `Host: LOCALHOST` 를 그대로 보내는 도구가 있다면 읽기까지 403 이다(추론, 실제 클라이언트로 재지는 않음).

  > **해소(§8):** 리드 검토에서 회귀로 판정돼 후속 회차에서 `strings.EqualFold` 비교로 고쳤다. 원문은 기록으로 남긴다.
- 역방향 프록시가 Host 를 바꿔 보내는 구성은 이제 모든 요청이 403 이다. 공식 문서가 지원하지 않는 구성이지만, 쓰던 사용자가 있다면 읽기까지 막힌다.
- 문서(docs-site 보안 모델 절)는 아직 "읽기는 게이트하지 않는다"는 전제를 반영하지 않았을 수 있다. 이번 카드 범위 밖이다.
- `httptest` 는 Host 를 정규화하지 않는다. 실제 서버 경로는 골든 패스와 중계기 테스트로 따로 쟀지만, HTTP/2 나 비표준 Host 형식은 재지 않았다.

## 7. 산출물

- `.moai/reports/t613/run/red.txt`, `green.txt`, `lint.txt`, `mutants.txt`, `loopback-probe.txt`
- 코드: `internal/web/app.go` (게이트 이동 + 거짓이 된 주석 5곳 정정, `@MX:NOTE` 갱신)
- 테스트: `internal/web/host_gate_test.go` (신규), `integration_test.go` 확장, 17개 테스트 파일 수정

## 8. Follow-up increment (lead review)

리드 검토에서 이 카드가 만든 회귀가 하나 나왔다. `isLoopbackHost` 가 `hostname == "localhost"` 로 대소문자를 구분해 비교하는데, 카드 이전에는 GET 이 Host 게이트를 타지 않아 `Host: LOCALHOST` 로도 페이지를 읽을 수 있었다. 게이트를 전 메서드로 넓힌 뒤로는 같은 요청이 403 이다. 호스트 이름은 대소문자를 구분하지 않으므로(DNS, RFC 3986 host 구성요소) 정책 변경이 아니라 고칠 회귀로 다뤘다.

### 8.1 Claim

1. `isLoopbackHost` 가 `localhost` 를 대소문자와 무관하게 받는다(`strings.EqualFold`). 포트가 붙어도 같다. 이 비교는 Unicode 단순 대소문자 접기라서 ASCII 밖의 접기 동치 표기(U+017F 가 섞인 `localhoſt`)도 받는다(§8.5, 미결).
2. 나머지 허용 범위는 넓히지도 좁히지도 않았다. `127.0.0.0/8` 전체, 대괄호 유무와 무관한 `::1`, IPv4-mapped loopback(`::ffff:127.x.y.z`)은 그대로 받고, 빈 Host 를 포함한 그 밖의 값은 그대로 거부한다.
3. 이 허용 범위가 이제 오버레이 프로브가 아니라 커밋된 단위 표(`TestIsLoopbackHost`)의 행으로 고정됐다. 대조군 행렬(`TestHostGateControlMatrix`)에도 `LOCALHOST`·`LocalHost:3041` 행이 정적 GET·`/settings` GET·POST `/save` 세 열로 들어갔다.
4. 새 행은 공허하지 않다. 비교를 대소문자 구분으로 되돌린 변이본(m4)에서 대문자 칸 9개가 정확히 실패한다.
5. `isLoopbackHost` 의 두 번째 호출자인 GLM 키 공개 핸들러(`glmkey.go`)의 loopback 재검사도 이제 대문자 `localhost` 를 받는다. 기존 `TestGLMKeyRevealLoopbackOnly` 는 GREEN 에서 통과한다.

### 8.2 Evidence

RED — 바뀌지 않은 `app.go` 에서 (`.moai/reports/t613/run/red-2.txt`)

```
$ GOMAXPROCS=2 go test -p 1 -count=1 -timeout 600s ./internal/web/... > .moai/reports/t613/run/red-2.txt 2>&1
exit=1
--- FAIL: TestIsLoopbackHost (0.00s)
    coverage_test.go:186: isLoopbackHost("LOCALHOST") = false, want true
    coverage_test.go:186: isLoopbackHost("LocalHost:3041") = false, want true
    coverage_test.go:186: isLoopbackHost("Localhost:8080") = false, want true
--- FAIL: TestHostGateControlMatrix (0.21s)
    --- FAIL: TestHostGateControlMatrix/localhost-upper/{static,settings,save}
    --- FAIL: TestHostGateControlMatrix/localhost-mixed-port/{static,settings,save}
FAIL	github.com/modu-ai/moai-adk/internal/web	22.058s
```

(행렬 실패 6줄은 위에서 경로만 묶어 적었다. 원문은 `red-2.txt` 에 있다.) 최상위 실패 2개, 서브테스트 실패 6개, 빌드 실패·panic·race 0. 고정한 넓은 범위 행(`127.1.2.3:3041`, `127.255.255.254`, `[::ffff:127.0.0.1]:3041`, `::ffff:127.0.0.1`, `::1`)과 거부 행은 RED 에서도 통과했다. 즉 앞 회차 프로브 값이 맞았다.

GREEN (`.moai/reports/t613/run/green-2.txt`)

```
$ GOMAXPROCS=2 go test -p 1 -count=1 -timeout 600s ./internal/web/... > .moai/reports/t613/run/green-2.txt 2>&1
exit=0
ok  	github.com/modu-ai/moai-adk/internal/web	22.383s
```

GREEN 은 두 번 쟀다. 첫 실행(`23.834s`, exit=0)은 godoc 을 고치기 전 트리였고, Unicode 접기 사실을 godoc 에 적은 뒤(주석만 변경) 같은 명령을 다시 돌려 `green-2.txt` 를 덮어썼다. 위 값이 최종 트리의 값이다.

Lint (`.moai/reports/t613/run/lint-2.txt`)

```
$ go vet ./internal/web/...           → exit=0, 출력 없음
$ golangci-lint run ./internal/web/... → 0 issues. exit=0
$ gofmt -l internal/web               → 출력 없음, exit=0
```

변이본 (`.moai/reports/t613/run/mutants-2.txt`) — 추적 파일은 고치지 않고 `go test -overlay` 로만 주입했다.

```
m4: if hostname == "localhost" {   (GREEN: strings.EqualFold(hostname, "localhost"))
$ GOMAXPROCS=2 go test -p 1 -count=1 -timeout 600s -overlay <scratch>/m4_overlay.json ./internal/web/...
exit=1
build_failed_lines=0  panic_or_race=0  top_level_FAIL=2  subtest_FAIL=6
FAIL	github.com/modu-ai/moai-adk/internal/web	24.587s
```

실패 집합은 RED 와 같다. 대문자 단위 행 3개와 대문자 행렬 칸 6개만 실패하고, 넓은 범위 행과 거부 행은 통과한다.

허용 범위 재측정 (`.moai/reports/t613/run/loopback-probe-2.txt`) — GREEN 코드에 입력 23개를 넣는 프로브를 오버레이로 주입했고(`exit=0`, `PROBE` 줄 23개), 커밋된 표와 공개 핸들러 테스트를 `-v` 로 따로 돌렸다(`exit=0`, `--- PASS: TestIsLoopbackHost`, `--- PASS: TestGLMKeyRevealLoopbackOnly` 와 서브테스트 4개 PASS).

| 판정 | 입력 |
|---|---|
| 받음 | `localhost`, `localhost:3041`, `LOCALHOST`, `LocalHost:3041`, `Localhost:8080`, `127.0.0.1`, `127.0.0.1:3041`, `127.1.2.3:3041`, `127.255.255.254`, `::1`, `[::1]`, `[::1]:3041`, `::ffff:127.0.0.1`, `[::ffff:127.0.0.1]:3041` |
| 거부 | `localhost.`, `LOCALHOST.`, `0.0.0.0`, `0.0.0.0:8080`, `10.0.0.5:8080`, `attacker.example.com`, `localhost.attacker.example.com`, `LOCALHOST.attacker.example.com`, 빈 값 |

주석 정정: `isLoopbackHost` godoc 을 실제 허용 집합으로 다시 썼고, `hostCheckMiddleware` godoc 의 "(127.0.0.1 / localhost / ::1)" 을 `isLoopbackHost` 참조로 바꿨다. `internal/web` 에서 좁은 집합을 주장하는 주석을 다시 찾아 보니 `handlers_test.go:469` 가 남는데, 이것은 그 테스트가 넣는 입력 세 개를 설명하는 문장이지 허용 집합이 그것뿐이라는 주장이 아니어서 두었다.

투명 문자 검사: RED 커밋 직전 스테이징 diff(141,810자)에서 유니코드 Cf 문자 0개, `chr(0x200B)` 를 넣은 대조 문자열은 1개로 셌다. GREEN 커밋 직전 값은 커밋 보고에 따로 적는다.

### 8.3 Baseline-attribution

| 측정 | 트리 | 코드 상태 |
|---|---|---|
| RED (`red-2.txt`) | 작업 트리 = RED 커밋 `32ea783c4` 내용(기준 HEAD `4b275def4` + 테스트만 변경, 측정 시 `git diff -- internal/web/app.go` 빈 출력) | 대소문자 구분 비교 |
| GREEN, lint (`green-2.txt`, `lint-2.txt`) | `32ea783c4` + `app.go` 최종 수정분(이 문서를 싣는 GREEN 커밋 내용) | `strings.EqualFold` 비교 + 최종 godoc |
| m4, 프로브 Part A·B·C (`mutants-2.txt`, `loopback-probe-2.txt`) | 위 트리에서 godoc 한 항목(Unicode 접기 설명)만 빠진 상태 | 코드 줄은 최종 트리와 같다. 차이는 주석뿐이다 |

앞 회차 프로브(`loopback-probe.txt`)는 GREEN 커밋 `8239e8dd2` 트리 기준이다. 이번 회차의 넓은 범위 판정은 그 값을 옮겨 쓰지 않고 위 트리에서 다시 쟀다.

### 8.4 Gaps

- **`strings.EqualFold` 의 Unicode 접기 범위는 결정하지 않았다.** U+017F 가 섞인 `localhoſt` 도 받아진다는 것은 쟀지만(§8.5), ASCII 대소문자만 받도록 좁힐지는 정하지 않았고 단위 표에도 행이 없다. 다른 접기 쌍은 따로 재지 않았다.
- **끝 점 형태 `localhost.` 는 결정하지 않았다.** GREEN 에서도 거부된다(`LOCALHOST.` 도 거부). 절대 도메인 표기라 의미상 `localhost` 와 같다고 볼 여지가 있지만, 이번 회차에서는 받을지 말지를 정하지 않았고 단위 표에도 어느 쪽으로든 행을 넣지 않았다. 미결 항목이다.
- 공개 핸들러(`/glm-key/reveal`)에 대문자 `localhost` 로 요청하는 칸은 따로 단언하지 않았다. 같은 함수를 호출하므로 같은 판정을 받는다는 것은 코드 구조에서 나온 추론이다.
- 실제 클라이언트가 `Host: LOCALHOST` 를 보내는 경우는 재지 않았다. `httptest` 요청과 `isLoopbackHost` 직접 호출로만 쟀다.
- 커버리지 수치는 재지 않았다.
- 변이 소스, 프로브 테스트 파일, 오버레이 JSON 은 스크래치 영역에 두었고 보고서 디렉터리로 옮기지 않았다. 변이 줄과 실행 요약은 `mutants-2.txt`, 프로브 출력은 `loopback-probe-2.txt` 에 있다.
- 이 절도 `moai-domain-humanize` 최종 교정을 거치지 않았다.

### 8.5 Residual-risk

- **SPEC 문구와의 어긋남은 그대로다.** 개정 REQ-WC-009 는 "정확히 `localhost`, `127.0.0.1`, `[::1]`"라고 적지만 코드는 `127.0.0.0/8`, IPv4-mapped loopback, 대소문자 무관 `localhost` 까지 받는다. 이번 회차는 코드 쪽 문서(godoc)만 실제에 맞췄고, SPEC 문구 정리는 이 카드 범위 밖이다.
- 공개 핸들러의 loopback 재검사가 대문자 `localhost` 를 받게 된 것은 허용 방향의 변화다. 대문자 이름도 같은 loopback 이라 방어 경계가 바뀌지는 않는다고 보지만, 그 핸들러만의 별도 판단은 거치지 않았다.
- **`strings.EqualFold` 는 ASCII 대소문자만 접지 않는다.** 유니코드 단순 대소문자 접기라서 U+017F(LATIN SMALL LETTER LONG S)가 `s` 로 접힌다. 오버레이 프로브로 재 보니 `localhoſt`, `localhoſt:3041`, `LOCALHOſT` 가 모두 받아졌다(`loopback-probe-2.txt` Part C, `exit=0`). 즉 실제 허용 집합은 "대소문자 무관 `localhost`"보다 조금 넓다. godoc 에는 이 사실을 그대로 적었다.
  - rebinding 방어 관점의 판단(추론, 실제 브라우저로 재지는 않음): 브라우저는 URL 호스트를 IDNA(UTS #46) 매핑으로 정규화하는데, 그 매핑이 U+017F 를 `s` 로 바꾸므로 `http://localhoſt:<port>` 는 `Host: localhost` 로 나간다. 공격자 도메인이 이 경로로 `localhost` 로 위장하는 수단은 보이지 않는다. 브라우저가 아닌 클라이언트는 원래 loopback 에 직접 접속할 수 있으므로 새로 열리는 경로도 아니다.
  - 그래도 리드 지시(`strings.EqualFold`)의 문자를 따랐기 때문에 남는 선택이다. ASCII 대소문자만 받도록 좁힐지는 이번 회차에서 정하지 않았고 단위 표에도 행을 넣지 않았다. 미결 항목이다.
