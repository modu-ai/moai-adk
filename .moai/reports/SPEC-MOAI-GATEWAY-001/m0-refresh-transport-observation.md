# M0 OAuth 갱신 전송 관측

## Claim

2026-09-11 19:19 KST, 설치 Claude Code 2.1.268의 Opus 5 한 세션에서 실제 OAuth refresh token POST의 응답 200과 그 응답 토큰의 후속 사용을 직접 연결했다. 로컬 401을 한 차례 받은 본문이 반환 토큰과 같은 Bearer hash로 재전송되어 upstream 200을 받았다. 별도 X-MoAI-Session-Token과 OAuth beta 헤더가 유지되었다. 이 증거는 이전 changed-Bearer-only 관측의 갱신 출처 GAP을 보완한다.

제품 코드·SPEC 상태·credential 저장소를 수정하지 않았다. Claude가 기존 로그인과 갱신을 직접 처리하도록 했으며, 관측기가 credential 파일이나 keychain을 읽거나 복제하지 않았다. 제품 gateway 활성화 및 전체 AC 완료를 주장하지 않는다.

## Evidence

소스: `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/m0_refresh_transport.py`
시험: 같은 디렉터리 `test_m0_refresh_transport.py`

RED: 구현 파일을 만들기 전 `python3 -m unittest discover -s <gateway-entry> -p test_m0_refresh_transport.py -v`:

```text
ModuleNotFoundError: No module named 'm0_refresh_transport'
Ran 1 test in 0.000s
FAILED (errors=1)
```

첫 구현 실행은 sandbox의 localhost bind 거부로 시험을 수행하지 못했다. 승인된 모의 실행에서 초과 길이 입력의 조기 거부가 client BrokenPipe를 일으켰다. 시험을 Content-Length 헤더만 먼저 보내 거부를 판정하도록 고쳤다. 이후 외부 제한을 둔 명령:

```text
timeout 45 python3 -m unittest discover -s .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry -p test_m0_refresh_transport.py -v

test_verified_token_exchange_and_secret_absence (test_m0_refresh_transport.TransportTests) ... ok

----------------------------------------------------------------------
Ran 1 test in 1.153s

OK
```

한 시험 안에서 신뢰한 TLS 서버의 응답 토큰 hash, 허용하지 않은 경로의 404와 upstream 호출 수 불변, 64KiB 초과 선언의 413과 upstream 호출 수 불변, 기록에서 세 가지 모의 비밀 원문 부재, 종료 뒤 listener 연결 실패, 신뢰하지 않은 upstream 인증서의 502와 요청 미전송을 판정했다. 인증서 검증을 끄지 않았다.

실제 실행 직전 clock 도구 출력:

```json
{"current_time":"2026-09-11 10:19:07 UTC"}
```

실제 실행 명령과 출력:

```text
timeout 180 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/m0_refresh_transport.py
{"output_file": "m0-refresh-transport-20260911T101909Z.json", "status": "OBSERVED", "refresh_observed": true}
exit_code: 0
```

결과 JSON 원문:

```json
{
  "started_utc": "2026-09-11T10:19:09.599810+00:00",
  "refresh_observed": true,
  "status": "OBSERVED",
  "m0_output": "m0-after19-20260911T101909Z-2ba632.json",
  "proxy_connections": 9,
  "token_exchanges": [
    {
      "refresh_grant": true,
      "status": 200,
      "bearer_sha256": "a1abf1f2c46a4bd37f54cb767abcc47499842cb5ca016501e0cb181ba6a82fdb"
    }
  ],
  "proxy_closed": true,
  "temporary_ca_removed": true
}
```

연결된 M0 결과에는 Opus 5 `returncode: 0`, `stdout_exact_OK: true`, `stop_reason: null`이 기록되어 있다. 첫 요청은 본문 hash `48c3bdfb3a5ba329b8cb35109d796e62ed9ad540b1e32c2490d7d8ca5bcfdbed`로 synthetic 401을 받았다. 세 번째 요청은 같은 본문이고 Bearer hash가 위 token exchange의 값과 같으며 `same_challenged_body`, `bearer_changed_after_challenge`, `custom_token_matches`, `oauth_beta_present`가 모두 true, upstream_status가 200이다. 둘째 요청은 다른 본문의 정상 요청이므로 갱신 상관관계의 근거로 사용하지 않았다.

기존 M0 helper의 refresh_observed=false는 변경하지 않았다. 그 필드는 원래 telemetry 계측 미지원으로 고정되어 있고, 이번 wrapper 결과가 token POST와 이후 Bearer의 실제 상관관계를 별도로 판정한다.

## Baseline-attribution

WT `moai-proxy-unified`, HEAD `81c1d58f9`. 설치 Claude 2.1.268 SHA-256 `06a96d5423f83770f120859f1c58e60d7252cc4c122aa13043b7e7cd716bc76a`는 선행 source 검토의 측정값이다. 이 실행에서 설치 업데이트나 별도 모델 버전 변경을 하지 않았다.

실행 후 직접 읽은 파일 모드와 SHA-256:

```text
m0_refresh_transport.py 0o600 53b29fc8e7b6a637aa6a1f018016ff41709db15ea88bca41b455c6a41f14d021
test_m0_refresh_transport.py 0o600 32acde2afe744857dd27b1901acf5e450c45a72d8f144008414fd9528dcc02a3
m0-refresh-transport-20260911T101909Z.json 0o600 1855cb6f9941f79adf92c9cb504555e8ebeb64890c5aefdfe836066a7c812d13
m0-after19-20260911T101909Z-2ba632.json 0o600 8b215c147196ba8ccf5baf1971149a916931ff1bdb6c136973087e9667cd7575
81c1d58f9
```

HTTPS_PROXY, NO_PROXY와 NODE_EXTRA_CA_CERTS의 시작 시 설정은 공식 [Enterprise network configuration](https://code.claude.com/docs/en/network-config)의 proxy 및 custom CA 절에서 확인했다. source의 고정 TOKEN_URL과 refresh POST 구현은 `m0-refresh-contract-review.md`에 바이트 오프셋으로 기록했다.

## Gaps

- 이번 직접 refresh 관측은 Opus 5 한 세션이다. Sonnet 5 정상 응답은 선행 M0 증거이며 Sonnet의 별도 직접 refresh는 재실행하지 않았다.
- 제품 gateway factory/credential resolver에 이 경로를 연결한 실행은 아니다. 해당 구현과 검증은 별도다.
- Windows, 지속 실행, 자연 만료, 프로세스 여러 개의 동시 갱신은 이 관측 범위가 아니다.
- Claude의 갱신 후 credential 영속화는 저장소를 직접 읽어 검증하지 않았다. 관측한 것은 실제 token POST 성공과 같은 세션의 후속 사용이다.
- 저장소 전체 시험 및 integration branch CI 판정은 PENDING이다.

## Residual-risk

관측 프록시는 운영 구성품이 아니다. 임시 self-signed CA를 해당 child의 환경에만 추가하고 정확한 platform.claude.com:443의 `/v1/oauth/token` POST만 중계했다. refresh grant만 허용하고 body/response를 64KiB로 제한하며 upstream SSLContext는 CERT_REQUIRED와 hostname verification을 요구한다. 다른 HTTPS는 공인 IP 목적지에 한해 내용을 해독하지 않는 tunnel로 전달했다. 시스템 trust store에는 CA를 등록하지 않았다.

실제 실행의 listener 연결은 9회였고 token POST는 한 번이었다. observer 종료와 임시 CA 디렉터리 삭제가 결과에 기록되어 있다. 외부 timeout 180, 기존 helper의 내부 alarm/child kill·wait, observer socket 종료를 함께 사용했다. 실제 실행 marker를 exclusive create하여 같은 시도를 자동 반복하지 못하게 했다.

토큰 원문과 요청/응답 원문은 observer 기록에 남기지 않는다. 정확한 grant 여부·상태·Bearer 접두어를 포함한 hash만 기록한다. 이는 해당 관측 프로세스의 제한된 진단이며, 전체 시스템의 비밀 유출 부재를 감사했다는 뜻은 아니다.
