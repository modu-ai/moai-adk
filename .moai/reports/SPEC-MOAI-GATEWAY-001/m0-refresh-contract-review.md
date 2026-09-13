# M0 OAuth 갱신 관측 검토

## Claim

현재 두 관측으로 T09 갱신 게이트를 통과시킬 수 없다. 첫 관측은 Opus 5·Sonnet 5 정상 응답과 별도 세션 헤더의 공존을, 둘째 관측은 한 차례 로컬 401 이후 같은 본문의 Bearer 교체와 upstream 200 복구를 입증한다. 해당 클라이언트가 refresh token POST를 직접 수행했는지는 관측하지 않았다. 제품 코드는 수정하지 않았고 추가 실제 클라이언트 호출도 하지 않았다.

`plan.md:70-94`의 M0는 정상 응답뿐 아니라 OAuth refresh 확인을 요구한다. `design.md:103-113`의 T09 양성 이전 구체 OAuth credential 등록 금지도 그대로 적용된다.

## Evidence

Python pathlib/hashlib/json으로 설치 바이너리와 두 결과 JSON을 읽고, `git -C <WT> rev-parse --short HEAD`를 실행한 출력:

```text
binary_sha256 06a96d5423f83770f120859f1c58e60d7252cc4c122aa13043b7e7cd716bc76a
HEAD 81c1d58f9
m0-after19-20260911T100000Z-905e6b.json 5afa8a254e13c49201f87fdf82bea60f7b9f245bcbb3c6d3d179ea87d6228263 False False
m0-after19-20260911T100109Z-d817f1.json da6b0796e27a7ede2425735510cf01040ee8c4d8efe39196cf71a5820d6d94ef True False
```

마지막 두 값은 각각 `oauth_recovery_observed`, `refresh_observed`이다. 결과 경로는 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/` 아래다.

설치 파일 `/Users/goos/.local/share/claude/versions/2.1.268`을 Python bytes 검색으로 읽었다. 다음 위치는 해당 SHA의 0 기반 바이트 오프셋이며 줄 번호가 아니다.

| 위치 | 실제 소스에서 확인한 경로 | 의미 |
|---|---|---|
| 162515964 | `P2`: `grant_type:"refresh_token"` → `st.post(Xt().TOKEN_URL,...)` → status 200 검사 → `tengu_oauth_token_refresh_success` | 이 이벤트는 해당 프로세스의 성공한 토큰 POST 뒤에 발생한다. |
| 162582117 | `tengu_oauth_401_recovered_from_disk` | 저장된 토큰이 실패한 토큰과 다르고 만료되지 않았으면 POST 없이 복구 가능하다. |
| 162583125 | `tengu_oauth_401_recovered_from_keychain` | keychain 값이 실패한 토큰과 다르면 POST 없이 복구 가능하다. |
| 162589804 | `tengu_oauth_token_refresh_starting` 뒤 P2 호출 | 시작 이벤트 자체는 완료 증거가 아니다. |
| 160582215 앞 모듈 | `i(t,e)` → `sink.logEvent(t,e)` | refresh 이벤트는 debug logger가 아니라 telemetry sink로 간다. |
| 162617093 앞 모듈 | `wU()` → `N2n({logEvent:m,logEventAsync:c})` → `zWe` | 실제 sink가 first-party event logger에 연결된다. |
| 162423892 | `zWe` → `lH` → logger emit | 일반 debug 파일 문자열 검색으로 이 이벤트를 관측한다는 근거가 없다. |
| 162407411 | first-party exporter endpoint 기본 `https://api.anthropic.com/api/event_logging/v2/batch` | 로컬 ANTHROPIC_BASE_URL만 바꿔서는 이 이벤트가 기존 forwarder에 들어오지 않는다. |
| 160588294 앞 모듈 | `S(e,r)` → `i("tengu_feature_ok",...)` | `S("oauth_token_refresh")`도 별도 diagnostics 파일 출력이 아니다. |
| 160044419 | TOKEN_URL = `https://platform.claude.com/v1/oauth/token` | 추가 관측 시 정확히 한정할 실제 endpoint다. |

현재 probe는 `--debug-file`을 읽어 telemetry 이름의 존재만 검사한다. 정상적인 refresh 성공 경로의 telemetry가 해당 파일에 실린다는 연결을 검증하지 않았으므로, 빈 marker 배열을 갱신 실패의 근거로 삼아서는 안 된다.

## Baseline-attribution

WT `moai-proxy-unified`, HEAD `81c1d58f9`에서 읽기 전용 검토했다. 설치 바이너리 2.1.268의 위 SHA와 두 결과 JSON에만 결론을 귀속한다. 다른 버전이나 플랫폼의 동작으로 일반화하지 않는다. 두 실행의 실제 결과는 부모가 만든 산출물을 읽어 확인한 것이며 이 검토자가 재실행한 결과가 아니다.

## Gaps

- 실제 토큰 POST 및 응답 200을 해당 child에 귀속한 관측이 없다.
- debug marker 수집에는 성공 경로의 양성 대조군이 없다. 읽은 파일의 전체 크기·잘림 여부도 현재 결과에 없다.
- telemetry sampling/비활성화/flush 실패 가능성 때문에 이벤트 부재만으로 POST 부재를 결론낼 수 없다.
- 별도 diagnostics 파일이나 표준 OTEL 설정이 이 first-party 이벤트를 수집한다는 경로는 확인하지 못했다.

## Residual-risk 및 제한된 추가 관측안

가장 직접적인 후보는 해당 child에만 적용하는 HTTPS proxy와 임시 CA로 정확한 token endpoint를 원본 인증서 검증을 유지해 중계하는 방식이다. 다른 세션·system trust store·keychain·토큰 만료값은 변경하지 않는다. 해당 child의 로컬 messages forwarder는 그대로 둔다.

실행 전에는 모의 token endpoint를 이용해 allowlist, 본문 크기 상한, 원본 TLS 검증 실패 시 차단, 시간 제한, 프로세스 회수, 비밀 미기록을 먼저 확인해야 한다. 설치 클라이언트의 child 전용 proxy/CA 수용 여부 자체도 아직 미검증이다. 구현이나 실행을 수행한 상태가 아니다.

관측기는 메모리에서만 요청을 전달하고 `grant_type=refresh_token` 여부, 응답 상태, 반환 access token의 SHA-256만 남긴다. refresh token·access token 원문·Authorization·전체 요청/응답·계정 metadata를 파일이나 출력에 남기지 않는다. 같은 한 차례 synthetic 401 시험에서 token POST 200 → 반환 access-token hash와 같은 Bearer → 같은 본문 upstream 200의 연결이 확인되어야 직접 갱신 양성이다. 저장소 토큰 채택만 일어나거나 timeout이면 여전히 갱신 GAP이다. 확인을 만들기 위해 기존 credential을 조작하거나 다른 세션을 중단해서는 안 된다.

CI와 저장소 전체 시험 판정은 이 읽기 전용 검토의 대상이 아니며 integration branch CI 판정은 PENDING이다.
