# t613 경위 규명 — moai web 의 외부 Host GET 허용

- 카드: t613 (Go 전수 점검 F05, security-findings.md 에서는 SEC-02)
- 단계: plan — 수리 전 판정 요청 (class C, Tier M)
- 측정 트리: `d3b7d438d2c9bc041cb3b63ea41f9f1a03e867b1` (로컬 develop, 워크트리 `WT-web-host-read` 를 fast-forward)
- 적용 규칙: `verification-claim-integrity.md` §1·§3 (주장·증거 분리), `verification-completeness.md` §1.1 (대조군과 변이본으로 판정 가능성 확인)

## 1. 결론 요약

1. **develop 에서 재현된다.** 외부 Host 와 Host 부재 모두 동적 GET 7경로에서 200을 받고, `/settings` 는 프로필 표식을 그대로 돌려준다. 같은 라우터의 POST 게이트는 외부·부재 Host 를 403으로 막는다.
2. **허용은 설계상 의도였지만, 근거가 기록된 적은 없다.** 최초 커밋(`b1ab60454`, 2026-06-03, SPEC-WEB-CONSOLE-001 M1)의 위협 분석은 "DNS rebinding/CSRF → 무단 변경"만 다뤘고, 읽기 노출은 분석 대상이 아니었다. "읽기는 안전하다"는 전제가 주석과 AC 에 결론으로만 들어갔다.
3. **그 전제는 이후 SPEC 들이 그대로 물려받으며 노출 면이 커졌다.** `/specs`, `/todo`, `/kanban`, `/monitor` 가 차례로 GET 라우트로 추가됐고, 일부는 "GET 은 게이트되지 않으므로 통과한다"를 주석에 명시한다.
4. **정책을 강화해도 문서화된 사용 방식은 깨지지 않는다.** 공식 문서의 원격 접근 방법은 SSH 터널뿐이고, 터널은 Host 를 `localhost:<port>` 로 유지한다. 역방향 프록시는 "지원하지 않음"으로 명시돼 있다.
5. **대신 테스트 파급이 크다.** 모든 메서드에 Host 게이트를 거는 변이본에서 `internal/web` 최상위 테스트 79개가 실패한다(원본은 통과). 기존 테스트 대부분이 loopback 이 아닌 기본 Host 로 요청하기 때문이다.

## 2. 경위 — 왜 테스트가 허용을 보장하는가

| 시점 | 위치 | 내용 |
|---|---|---|
| 2026-06-03 `b1ab60454` | `.moai/specs/SPEC-WEB-CONSOLE-001/spec.md` REQ-WC-009 | Host 검사 대상을 `POST`/`PUT`/`PATCH` 로 한정. 목적은 "다른 로컬 origin 의 DNS-rebinding/CSRF 방지" |
| 같은 커밋 | `plan.md` R3 | 위험 "DNS-rebinding/CSRF", 영향 칸은 **"Unauthorized mutation"** 뿐. 완화책 "mutating method 에 Host 검사, loopback bind 가 바깥 경계" |
| 같은 커밋 | `acceptance.md` AC-WC-009 | "**And** a `GET` request is not Host-gated (read remains accessible)" |
| 같은 커밋 | `internal/web/handlers_test.go:492` `TestHostCheckDoesNotGateGet` | 외부 Host(`attacker.example.com`) GET → 200 을 **요구**. 범위 경계(쓰기만 검사한다)가 적극적 계약(외부 Host 읽기는 성공해야 한다)으로 굳어짐 |
| 같은 커밋 | `internal/web/app.go:202` 주석 | "GET(읽기)은 게이트하지 않는다 — 읽기는 안전하므로 foreign Host여도 통과시킨다" |
| 이후 | SPEC-INTERNAL-SECURITY-001 REQ-SEC-002 | 미들웨어를 다시 열어 Sec-Fetch-Site CSRF 게이트를 추가. 범위는 상태 변경 라우트뿐이라 읽기 전제는 재검토되지 않음 |
| 이후 | SPEC-WEB-CONSOLE-011 (`/specs`), SPEC-WEB-TODO-QUEUE-001 (`/todo`) 등 | GET 라우트 추가. `app.go:170` 주석이 "hostCheckMiddleware 는 GET 을 게이트하지 않으므로 보드 읽기는 … 통과"라고 전제를 재확인 |

판독: "loopback bind 가 바깥 경계"라는 완화책은 DNS rebinding 에는 성립하지 않는다. rebinding 공격에서 브라우저는 실제로 127.0.0.1 에 접속하면서 `Host: <공격자 도메인>` 을 보내고, 브라우저 입장에서는 같은 origin 이라 응답을 읽을 수 있다. 이 경로를 막는 서버 측 수단은 Host 검사뿐인데, 그 검사가 쓰기에만 걸려 있다. 설계 문서 어디에도 읽기 노출을 따져 보고 수용했다는 기록은 없다 — 분석이 빠진 것이지 위험을 수용한 결정은 아니다.

반면 GLM 키 원문은 따로 막혀 있다. `/glm-key/reveal` 은 POST 전용이고 핸들러가 loopback 을 다시 확인한다(`internal/web/glmkey.go:115-123`). 기본 렌더에는 설정 여부와 끝 4자리만 나간다(`glmkey.go:40-49`). 따라서 원문 키 탈취는 이 카드의 범위가 아니다.

## 3. 재현 — develop 에서의 대조군 행렬

### Claim
외부 Host·Host 부재의 동적 GET 은 develop 에서 200 과 설정 내용을 받고, POST 는 403 으로 막힌다.

### Evidence
명령(프로브는 `go test -overlay` 로만 주입, 패키지 트리는 변경하지 않음):

```
GOMAXPROCS=2 go test -p 1 -count=1 -timeout 180s -overlay .moai/reports/t613/probe/overlay.json -run 'TestT613HostMatrix|TestHostCheckDoesNotGateGet|TestHostCheckRejectsForeignHostOnPost' -v ./internal/web > .moai/reports/t613/probe/run-develop-d3b7d438d.txt 2>&1
```

종료 코드 0. `RESULT` 줄 40개(GET 35 + POST 5), 원문은 `.moai/reports/t613/probe/run-develop-d3b7d438d.txt`. 발췌:

```
RESULT method=GET host=localhost path=/settings status=200 bytes=107706 marker=true
RESULT method=GET host=foreign path=/settings status=200 bytes=107704 marker=true
RESULT method=GET host=absent path=/settings status=200 bytes=107703 marker=true
RESULT method=GET host=foreign path=/todo status=200 bytes=8607 marker=false
RESULT method=GET host=foreign path=/static/app.js status=200 bytes=34665 marker=false
RESULT method=POST host=localhost path=/save status=200 wrote=true
RESULT method=POST host=foreign path=/save status=403 wrote=false
RESULT method=POST host=absent path=/save status=403 wrote=false
--- PASS: TestT613HostMatrix (1.84s)
ok  	github.com/modu-ai/moai-adk/internal/web	2.463s
```

행렬 요약:

| Host | 동적 GET 6경로 (`/settings` `/` `/kanban` `/monitor` `/todo` `/specs`) | 정적 `/static/app.js` | POST `/save` |
|---|---|---|---|
| `localhost:3041` | 모두 200 | 200 | 200, 기록됨 |
| `127.0.0.1:3041` | 모두 200 | 200 | 200, 기록됨 |
| `[::1]:3041` | 모두 200 | 200 | 200, 기록됨 |
| `attacker.example.com:3041` | **모두 200**, `/settings` 표식 포함 | 200 | 403, 기록 안 됨 |
| 부재(`""`) | **모두 200**, `/settings` 표식 포함 | 200 | 403, 기록 안 됨 |

표식(`marker=true`)은 `readPreferences` 를 가짜 UserName 으로 바꿔 넣은 `/settings` 에서만 확인 대상이다. 다른 경로의 `marker=false` 는 노출이 없다는 뜻이 아니라 표식을 심지 않았다는 뜻이다.

### Baseline-attribution
이번 실행, 트리 `d3b7d438d`. 보고서 기준 트리 `main 2213871af` 와 `internal/web/app.go` 는 동일(`git diff --stat 2213871af d3b7d438d -- internal/web` 목록에 `app.go` 없음). 단 `internal/web` 전체로는 32개 파일이 달라 main 판독을 그대로 옮기지 않고 develop 에서 다시 쟀다.

## 4. 파급 측정 — 게이트를 모든 메서드로 넓히면

### Claim
Host 게이트를 모든 메서드에 걸면 `internal/web` 의 기존 테스트 79개가 실패하고, 실패는 403 으로 나타난다.

### Evidence
변이본: `app.go` 사본에서 `isLoopbackHost` 검사를 `switch` 앞으로 옮김(`.moai/reports/t613/probe/app_mutant_gate_all.go.txt`). 대조군은 원본 트리의 같은 패키지 실행.

```
GOMAXPROCS=2 go test -p 1 -count=1 -timeout 300s -overlay .moai/reports/t613/probe/overlay-mutant-gate-all.json ./internal/web > .moai/reports/t613/probe/run-mutant-gate-all.txt 2>&1     → exit 1
GOMAXPROCS=2 go test -p 1 -count=1 -timeout 300s ./internal/web > .moai/reports/t613/probe/run-baseline-package.txt 2>&1                                                                      → exit 0
```

- 변이본: `grep -c '^--- FAIL'` → `79`, `grep -c '403'` → `91`, 마지막 줄 `FAIL	github.com/modu-ai/moai-adk/internal/web	11.144s`
- 대조군: `ok  	github.com/modu-ai/moai-adk/internal/web	14.582s`
- 실패 표본: `todo_route_test.go:39: GET /todo status = 403, want 200 … forbidden: non-loopback Host header`, `tab_layout_test.go:147: GET / status = 403, want 200`, `handlers_test.go:56: /static/app.js: status = 403, want 200`

실패에는 `TestHostCheckDoesNotGateGet` 가 포함된다 — 계약을 뒤집으면 그 테스트는 반대 방향으로 고쳐야 한다.

### Baseline-attribution
이번 실행, 트리 `d3b7d438d`, 두 실행 모두 같은 트리.

## 5. 운영자 판정이 필요한 것

선택지(수리 방향). 어느 쪽이든 기존 CSRF(Sec-Fetch-Site)·키 공개 방어는 유지한다.

| 안 | 내용 | 막히는 것 | 비용 |
|---|---|---|---|
| A. 전 라우트 Host 게이트 | 검사를 메서드와 무관하게 적용(정적 자산 포함) | 외부·부재 Host 의 모든 요청 | 코드 변경은 몇 줄. 테스트 79개를 loopback Host 로 옮기는 기계적 수정. `TestHostCheckDoesNotGateGet` 반전 |
| B. 동적 라우트만 게이트 | `/static/` 은 열어 두고 나머지 GET 을 막음 | 외부·부재 Host 의 동적 화면 | A 보다 분기가 하나 늘고, 정적 자산 테스트 일부는 그대로 통과. 파급 수는 따로 재지 않음 |
| C. 현 정책 유지, 위험 수용을 기록 | 코드 그대로. `app.go` 주석의 "읽기는 안전하므로"를 근거 있는 수용 문구로 바꾸고 문서 보안 모델에 명시 | 없음 | 문서·주석만. DNS rebinding 조건에서 설정·SPEC·큐 화면 읽기 가능성이 남음 |

참고 사실:
- 공식 문서(`docs-site/content/ko/advanced/moai-web-console.md` 보안 모델)는 역방향 프록시·`0.0.0.0` 노출을 지원하지 않고, 원격 접근은 SSH 터널로 안내한다. 이 경우 Host 는 loopback 이라 A·B 어느 쪽에서도 막히지 않는다(추론, 실측 아님).
- 같은 문서는 "인증 없음은 루프백 전용이 전제"라고 적는데, 그 전제가 rebinding 읽기에서는 성립하지 않는다. A·B 를 택하면 문서 보안 모델 절에 Host 검사 설명을 더하는 것이 자연스럽다(4개 로케일 동기화 대상).

## 6. Gaps (측정하지 않은 것)

- 실제 브라우저 + DNS rebinding 공격은 수행하지 않았다. 최신 브라우저의 사설망 접근 제한이 이 경로를 얼마나 막는지는 확인하지 않았다.
- `/`, `/kanban`, `/monitor`, `/todo`, `/specs`, `/events` 가 구체적으로 어떤 값을 내보내는지는 내용 단위로 재지 않았다(상태 코드와 크기만).
- 안 B 의 테스트 파급 수는 재지 않았다.
- 79개 실패 전부가 Host 기본값 때문인지는 표본 3건과 `403` 출현 수로만 판단했다. 개별 확인은 하지 않았다.
- SSH 터널이 Host 를 loopback 으로 유지한다는 점은 실측이 아니라 일반적 동작에 기댄 추론이다.
- 이 보고서는 `moai-domain-humanize` 최종 교정을 거치지 않았다.

## 7. Residual-risk

- 프로브는 `httptest` 로 라우터를 직접 호출했다. 실제 `net/http` 서버 경로(예: Host 헤더 정규화)는 거치지 않았으므로, 서버 수준에서 결과가 달라질 여지는 남는다. 다만 게이트는 핸들러 체인 안에 있어 영향은 작다고 본다.
- `/events`(SSE)는 행렬에 넣지 않았다. 코드 주석상 값이 아니라 변경 신호만 흘리지만, 실측은 아니다.

## 8. 산출물

- `.moai/reports/t613/verdict.md` — 이 문서
- `.moai/reports/t613/probe/t613_probe_test.go.txt`, `overlay.json` — 대조군 행렬 프로브
- `.moai/reports/t613/probe/app_mutant_gate_all.go.txt`, `overlay-mutant-gate-all.json` — 파급 측정 변이본
- `.moai/reports/t613/probe/run-develop-d3b7d438d.txt`, `run-mutant-gate-all.txt`, `run-baseline-package.txt` — 실행 원문
