# 일반 변환 부모 추가 시험

## Claim

최초 일반 변환 구현의 escaped lone UTF-16 surrogate 입력에서 오류 대신 대체 문자로 바뀐 출력이 생성되었다.
이는 일반 변환 보고서의 invalid UTF-8 바이트 거절 시험과 다른 입력 경계다. 아래 최초 증거는 수정 전 관측이다.
후속 TDD 수리로 단독 surrogate를 거절했고 동일 부모 probe에서 거절을 재확인했다. 전체 변환 감사 판정은 아니다.

## Evidence

원본 시험은 ignored 경로의 `gateway-entry/unicode-probe/main.go`에 두었다. six ASCII bytes
`[92,117,100,56,48,48]`로 lone surrogate escape를 구성하여 user.content에 넣고 공개 Request API를 호출한다.

```text
$ GOCACHE=/tmp/gateway-parent-probe-cache go run ./.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/unicode-probe
rejected=false
output={"input":[{"content":[{"text":"�","type":"input_text"}],"role":"user"}],"max_output_tokens":1,"model":"gpt-5.6-sol","store":false}
```

Exit 0. Request가 error를 반환하지 않았다는 관측이며 이 probe의 exit 0을 통과 판정으로 쓰지 않는다.

## Baseline-attribution

작업 디렉터리는 `moai-proxy-unified` WT다. 같은 실행에서 `git rev-parse --short HEAD`를 다시 읽은 값은
`81c1d58f9`다. request.go의 SHA-256은 일반 변환 완료 보고서와 같은
`1fbcb3b65a2087a4fa0c96fe1402151bab97bd733271990f8b28ce1635e3f5be`다.

## Gaps

후속 수리의 valid surrogate pair, low surrogate 단독, literal U+FFFD, escaped U+FFFD, escaped backslash 대조군을 실행했다.
이 범위의 미수정 상태는 해소했다. Claude나 실제 공급자를 호출하지 않았으며 모든 Unicode/프로토콜 조합을 전수 검사한 것은 아니다.

## 후속 수리 증거

Request 공개 API의 `TestUnicodeEscapesAreScalars`를 구현 전에 추가했다. ASCII byte 92·117과 hex 문자열을 조합하여 escape를 만들었다.
단독 high/low, high 뒤 일반 scalar, 연속 high 네 음성군에서 `escaped surrogate replaced`와 대체 문자 출력이 관측되었다.
당시 scoped 명령 `GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/translate -run TestUnicodeEscapesAreScalars -count=1`은 exit 1,
`FAIL github.com/modu-ai/moai-adk/internal/gateway/translate 0.395s`였다.

기존 JSON decoder 앞에 작은 Unicode escape 검사만 추가했다. 기존 UTF-8·duplicate·depth·size 검사와 decoder는 유지했다.

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/translate -run TestUnicodeEscapesAreScalars -count=1
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.326s
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/translate -count=1 -coverprofile=/tmp/gateway-translation-unicode-cover.out
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.344s	coverage: 92.6% of statements
$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway/translate -count=1
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.537s
$ GOCACHE=/tmp/gateway-translation-cache go run ./.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/unicode-probe
rejected=true
output=
```

각 exit 0. 같은 WT의 HEAD는 재확인한 `81c1d58f9`다. 수리 뒤 request.go SHA-256은
`15fd3fa483a1a8a9ab7d01aa235ed23ad81b6b63a3a4bd48ac8af2e2c14f735d`다.

## Residual-risk

JSON decoder 앞의 UTF-8 바이트 검사만으로 JSON escape의 Unicode scalar 값까지 보장되지는 않는다. 고칠 때에는
기존 duplicate/depth/size 검사를 유지하고 별도 거대한 JSON parser를 만들지 않는 최소 검증을 우선한다.
