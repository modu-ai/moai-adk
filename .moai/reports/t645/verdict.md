# t645 — 웹 콘솔 보안 모델의 '다른 계정에서는 닿지 않는다' 주장 실측과 4 로케일 정정

- card: t645 (Class B) · worktree `.claude/worktrees/t645` · branch `WT-web-console-reach`
- base: 로컬 develop `dae1b070b` (fast-forward)
- 판정: **주장은 틀렸다.** 루프백 TCP 포트는 계정별로 나뉘지 않는다. 4 로케일 문장을 정정했다. 인증 필요 여부는 운영자 결정으로 올린다.

## 1. 주장과 코드

`docs-site/content/{ko,en,ja,zh}/advanced/moai-web-console.md:173` 은 "콘솔은 `127.0.0.1` 에만 바인딩되므로 같은 머신의 다른 계정이나 원격 호스트에서는 닿지 않는다"고 했고, `:177` 의 "인증 없음" 문단은 이것을 전제로 삼았다.

코드는 평범한 TCP 리스너다 — `internal/web/server.go:180-183`:

```go
addr := fmt.Sprintf("%s:%d", loopbackHost, s.cfg.Port)   // loopbackHost = "127.0.0.1" (:44)
ln, err := net.Listen("tcp", addr)
```

TCP 소켓에는 파일 권한이 없고, 루프백 바인딩은 원격 호스트만 막는다. 인증 계층도 없다(`internal/web` 비테스트 소스에 로그인·토큰 경로 없음). `hostCheckMiddleware`(`internal/web/app.go:215-236`)는 Host 헤더가 루프백인지(DNS 리바인딩 방어)와 상태 변경 요청의 `Sec-Fetch-Site: same-origin`(브라우저발 CSRF 방어)만 본다. 브라우저가 아닌 로컬 클라이언트는 두 헤더를 직접 넣을 수 있으므로, 이 검사는 같은 머신의 다른 계정을 가려내는 장치가 아니다 — **코드 판독**이며 실행해 보지 않았다.

## 2. 실측 — 다른 계정의 루프백 전용 리스너에 접속

다른 로컬 계정을 만들려면 sudo·시스템 변경이 필요해 하지 않았다. 대신 이 머신에 **이미 떠 있는, 다른 계정이 연 `127.0.0.1` 전용 리스너**를 찾아 내 계정으로 접속했다. 콘솔과 같은 형태(루프백 전용 TCP)를 반대 방향에서 잰 것이다.

- 루프백·와일드카드 리스너 34개의 소유자를 `ps` 로 확인 → uid 501 이 아닌 것은 `launchd`(pid 1, root)와 `kdc`(pid 377, root) 둘. 그중 루프백 전용은 `127.0.0.1:8021`(launchd) 하나.
- 전문: `cross-account-loopback-probe.txt`

```
$ id -un; id -u                         → goos / 501
$ ps -o user=,uid=,pid=,comm= -p 1      → root 0 1 /sbin/launchd
$ nc -z -v -G 3 127.0.0.1 8021          → Connection to 127.0.0.1 port 8021 [tcp/intu-ec-client] succeeded!   exit=0
$ nc -z -v -G 3 127.0.0.1 1  (대조군)   → connectx ... failed: Connection refused                             exit=1
```

uid 501 프로세스가 uid 0 소유의 루프백 전용 리스너에 TCP 연결을 맺었다. 대조군(리스너 없는 포트)은 거절됐으므로 접속 성공은 리스너에 닿은 결과다.

## 3. 정정 (4 로케일, 파일당 2줄)

| 로케일 | `:173` 루프백 전용 | `:177` 인증 없음 |
|---|---|---|
| ko | 원격 호스트에서는 닿지 않지만, 루프백 포트는 계정별로 나뉘지 않아 같은 머신의 다른 계정은 접속할 수 있다 | 로그인·토큰 계층 없음. 루프백 바인딩은 원격은 막고 다른 계정은 못 막는다. 공유 머신에서는 콘솔이 켜진 동안 그 계정들도 접속할 수 있다 |
| en / ja / zh | 같은 내용을 각 로케일의 자연스러운 문장으로 | 같음 |

문서에는 **접속할 수 있다**까지만 적었다. "설정을 읽고 바꿀 수 있다"는 §1 의 코드 판독에 기댄 추론이라 문서 주장으로 올리지 않았다.

## 4. 검증

```
grep -rn "다른 계정|Another account|another account|別アカウント|其他账户" docs-site/content README*.md internal/cli/web.go .claude/skills
  → 정정 전 적중은 네 로케일 :173 뿐(다른 사본 없음)
grep -rln "moai-web-console" internal cmd pkg --include='*_test.go'   → 없음(이 페이지를 읽는 Go 테스트 없음)
로케일 구조: ko/en/ja/zh 모두 제목 18개, 208줄 (정정 전후 동일)
git diff --stat → 4 files changed, 8 insertions(+), 8 deletions(-)
hugo --source docs-site --destination <scratchpad> --logLevel warn → hugo-exit=0, warn/error 0줄 (hugo-build.txt)
빌드 결과 ko·en 페이지에 새 문장이 렌더됨(grep 적중)
```

## 5. 운영자 결정 필요 — 인증

다른 로컬 계정이 콘솔 포트에 닿을 수 있다는 것이 실측으로 확인됐다. 콘솔에 인증(예: 프로세스별 토큰)을 넣을지는 이 카드의 범위가 아니며 운영자 결정으로 올린다. 이 카드는 문서를 사실에 맞췄을 뿐 동작은 바꾸지 않았다.

## Gaps

- **정방향은 재지 않았다.** 다른 계정으로 `moai web` 콘솔 자체에 접속해 보지 않았다(계정 생성에 sudo 필요). 역방향(내 계정 → root 의 루프백 전용 리스너)으로 커널 루프백에 계정 구분이 없음을 보였고, 콘솔은 같은 `net.Listen("tcp", "127.0.0.1:…")` 이라 같은 규칙을 따른다는 것은 코드 판독이다.
- `moai web` 을 띄워 HTTP 요청이 처리되는지까지는 재지 않았다(실행 중인 설치 바이너리가 소스와 다를 수 있고, 카드의 질문은 도달 가능성이다).
- macOS 에서만 쟀다. Linux 도 루프백 TCP 에 계정 필터가 없는 것이 일반적이지만 이 머신에서 관측하지 않았다. 방화벽(pf 등)으로 사용자별 루프백 규칙을 두는 환경은 예외일 수 있다.
- zh·ja 렌더 확인은 빌드 성공과 소스 diff 로만 했다(렌더 결과 grep 은 ko·en 만).

## Residual-risk

- 문서의 "공유 머신에서 다른 계정도 접속할 수 있다"는 위험을 알리지만 막지는 않는다. 인증 결정 전까지 공유 머신 사용자는 이 문장에 기대야 한다.
