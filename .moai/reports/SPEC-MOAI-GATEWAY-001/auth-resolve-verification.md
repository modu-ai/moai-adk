# AUTH 만료 시 해석 연결 검증

## Claim

`Store.ResolveFresh(ctx, Broker, verifier) (CredentialRef,error)`를 새 `auth/resolve.go`에 추가했다. 기존 명시 로그인 상태가 유효하면 broker/verifier를 호출하지 않는다. missing/tombstone이면 자동 로그인하지 않는다. 만료 상태만 기존 Refresh의 operation lock/세대 CAS로 처리한 뒤 caller context로 다시 읽고, 새로운 유효 세대의 Store 소유 ref만 반환한다. 실패한 시도 뒤 다른 작업이 유효한 새 세대를 게시했으면 그 세대를 재사용할 수 있다. 재시도 루프·선제 갱신 시각·401 fallback·다른 모델/API 키 fallback은 없다.

공통 Store/Broker/Owns/F1/F2/Windows gate와 CLI/root는 변경하지 않았다. factory에서 호출할 수 있는 구성 요소이며 아직 실제 production factory에 연결하지 않았다. 반환 뒤 logout 경합은 기존 SendAuthorized의 송신 직전 세대 장벽이 계속 담당한다.

## Evidence

모든 시험 명령은 같은 invocation에서 provider 환경을 unset하고 `GOCACHE=/tmp/gateway-foundation-cache`로 실행했다.

시험 먼저 작성한 RED:

`go test ./internal/gateway/auth -run TestResolveFresh -count=1 -timeout 20s`

```text
# github.com/modu-ai/moai-adk/internal/gateway/auth [github.com/modu-ai/moai-adk/internal/gateway/auth.test]
internal/gateway/auth/resolve_test.go:41:14: s.ResolveFresh undefined (type *Store has no field or method ResolveFresh)
internal/gateway/auth/resolve_test.go:48:13: s.ResolveFresh undefined (type *Store has no field or method ResolveFresh)
internal/gateway/auth/resolve_test.go:57:20: empty.ResolveFresh undefined (type *Store has no field or method ResolveFresh)
internal/gateway/auth/resolve_test.go:78:16: s.ResolveFresh undefined (type *Store has no field or method ResolveFresh)
internal/gateway/auth/resolve_test.go:114:16: s.ResolveFresh undefined (type *Store has no field or method ResolveFresh)
internal/gateway/auth/resolve_test.go:141:14: s.ResolveFresh undefined (type *Store has no field or method ResolveFresh)
FAIL github.com/modu-ai/moai-adk/internal/gateway/auth [build failed]
FAIL
```

중간 fixture는 Login 없이 state.json만 심어 sidecar lock이 없는 상태였고 동시 해석 시험에서 Status ErrAuthState가 발생했다. 별도 Status 호출로 sidecar를 먼저 만든 진단 대조군은 통과했다. 최종 fixture는 실제 명시 로그인 계약대로 loginFixture를 먼저 실행한 뒤 내부 시험 도우미로 expiry만 과거로 바꾼다. sidecar 없는 상태의 동시 생성 문제를 해결한 것으로 기록하지 않는다.

최종 AUTH 범위 regression/race:

`go test -race ./internal/gateway/auth -count=1 -timeout 40s -coverprofile=/tmp/gateway-resolve-cover.out`

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway/auth	5.086s	coverage: 88.4% of statements
```

`go vet ./internal/gateway/auth`: exit 0, 빈 출력.

`GOCACHE=/tmp/gateway-foundation-cache go tool cover -func=/tmp/gateway-resolve-cover.out`의 새 함수 행:

```text
github.com/modu-ai/moai-adk/internal/gateway/auth/resolve.go:11: ResolveFresh 88.9%
```

첫 cover 도구 호출에는 cache 변수가 이어지지 않아 기본 cache permission 오류가 났다. 위 명시 cache 재실행만 coverage 근거다. 함수 행은 공백 열만 축약했다.

새 시험의 실제 단언:

- 유효 ref는 owned이며 broker/verifier 호출 수 0; missing/logout도 추가 호출 0.
- 동시 8호출에서 broker 1회·verifier 1회이며 모두 owned ref를 받는다.
- broker 실패·verifier 401·여전히 만료된 후보는 ref nil, 기존 generation 1 유지.
- refresh 중 logout은 미로그인 유지·ref nil.
- 지연 시도 중 새 유효 generation 2를 게시한 fixture는 실패한 후보 대신 새 owned ref를 받는다.
- caller context가 state.lock 대기에 적용되어 deadline 오류와 ref nil을 반환한다.
- broker/verifier가 없을 때 만료 ref를 반환하지 않는다.

## Baseline-attribution

2026-09-11 WT moai-proxy-unified, 기준 HEAD 81c1d58f9. 현재 Store/Refresh를 먼저 읽고 새 resolve.go/resolve_test.go만 작성했다. 가짜 broker와 합성 token fixture만 사용했으며 실제 계정·설치 클라이언트·외부 network는 호출하지 않았다. 이전 테스트 수치를 이번 새 함수의 근거로 전용하지 않았다.

## Gaps

공식 broker refresh와 실제 verifier, root/child factory 연결, Windows native 실행은 남아 있다. Windows OpenStore gate는 유지한다. 임의 Broker/verifier가 caller context를 무시하면 ResolveFresh가 그 외부 구현을 강제로 종료할 수는 없다. 실제 CodexBroker의 timeout/watchdog는 기존 별도 계약이다. 저장소 전체 판정은 통합 브랜치 CI 소관이며 PENDING이다. commit/push/PR/merge 없음.

## Residual-risk

반환 직후 credential이 만료·폐기될 수 있으므로 ResolveFresh 성공을 송신 허가로 사용하면 안 된다. SendAuthorized를 반드시 사용한다. 상태가 계속 바뀌거나 최신 세대도 만료면 무한 갱신하지 않고 오류를 낸다. 명시 로그인 없이 state 파일만 주입하거나 sidecar 파일을 외부에서 삭제한 상태의 동시 잠금 생성은 이번 범위에서 해결하지 않았다. 이는 위 관측에 한정한 경계이며 모든 Store 동시성의 결함 판정은 아니다.
