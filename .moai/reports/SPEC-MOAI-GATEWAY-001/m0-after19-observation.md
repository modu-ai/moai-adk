# M0 — 19시 이후 실제 Claude 인증 관측

## Claim

Opus 5와 Sonnet 5의 실제 Anthropic 요청은 HTTP 200을 받았고 각 Claude 프로세스는 exit 0과 `OK`를 반환했다. 별도 `X-MoAI-Session-Token`과 기존 OAuth Bearer가 함께 실렸으며 로컬 세션 토큰은 upstream 전달 전에 제거했다.

별도 한 번의 합성 401 관측에서는 같은 요청 본문이 바뀐 Bearer로 재전송되어 HTTP 200을 받았다. 이는 인증 복구 관측이다. 이 클라이언트가 새 토큰을 직접 발급받았는지 다른 프로세스의 갱신 결과를 채택했는지는 아직 분리하지 못했으므로, T09 전체 PASS나 refresh 발급 성공으로 표시하지 않는다.

## Evidence

2026-09-11 10:00:03 UTC에 부모가 시각을 확인했다. 앞서 예약한 대기 프로세스는 관측기의 자체 시각 제한과 함께 10:00:00.028718 UTC에 M0를 시작했다.

기본 관측 명령:

```sh
timeout 160 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/m0_after19.py
```

출력과 종료:

```text
{"status": "INCONCLUSIVE", "output_file": "m0-after19-20260911T100000Z-905e6b.json", "requests": 4, "stop_reason": null, "refresh_observed": false}
exit 0
```

부모가 위 JSON을 직접 읽었다. Opus 5·Sonnet 5 각각 `returncode:0`, `stdout_exact_OK:true`; 요청 4개 모두 `authorization_kind:bearer`, `api_key_present:false`, `oauth_beta_present:true`, `custom_token_matches:true`, `upstream_status:200`이었다. refresh 시도는 이 기본 관측에 포함되지 않는다.

후속의 명시적 단일 401 관측:

```sh
timeout 160 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/m0_after19.py --refresh-probe
```

```text
{"status": "INCONCLUSIVE", "output_file": "m0-after19-20260911T100109Z-d817f1.json", "requests": 5, "stop_reason": null, "refresh_observed": false}
exit 0
```

부모가 원문 JSON을 읽었다. 첫 Opus 제목용 요청에만 로컬 합성 401을 반환했고 upstream으로 보내지 않았다. 이후 동일 본문 hash의 Opus 요청은 `same_challenged_body:true`, `bearer_changed_after_challenge:true`, `custom_token_matches:true`, `upstream_status:200`이었다. 그 사이 본 대화 요청도 HTTP 200을 받았다. 후속 Sonnet 요청 둘도 HTTP 200이며 두 프로세스 모두 exit 0·정확한 `OK`였다.

관측기는 `oauth_recovery_observed:true`를 기록했다. 임시 debug 파일에서 미리 정한 telemetry marker만 검색했으나 `markers_present:[]`였다. 이 부재를 갱신 실패나 미실행 증거로 해석하지 않는다. raw debug 파일은 임시 디렉터리와 함께 제거했다.

실제 정책 필드도 관측했다. 두 모델의 제목용 요청은 max_tokens 64000, stream true, thinking disabled, effort high, JSON schema 출력이었다. 본 대화 요청은 같은 출력 상한·stream, adaptive thinking, effort high, context_management 포함이었다. 도구는 의도적으로 비활성화한 시험이므로 tool_count 0이다.

## Baseline-attribution

- WT: `moai-proxy-unified`, branch `WT-unified-gateway`, HEAD `81c1d58f9`.
- 실행 시 source_session_id: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`.
- 실제 요청한 Claude ID: `claude-opus-5`, `claude-sonnet-5`.
- 증거 JSON은 같은 ignored `gateway-entry` 디렉터리 안의 위 파일명이다. 토큰 원문·메시지 본문은 저장하지 않고 종류·hash·허용된 정책 메타데이터만 기록했다.
- 세션별 임시 프로젝트, 고정 Anthropic endpoint, 요청 수 제한, 실제 client 종료·listener 정리, 외부 timeout을 사용했다. 별도의 GPT 계정 시험과 로컬 mock TUI 시험은 이 보고서의 요청 수에 포함하지 않는다.

## Gaps

토큰 자체 발급과 디스크/keychain의 새 토큰 채택을 구별하는 관측이 남아 있다. 현재 제품 gateway를 통한 passthrough·stream 지연·실제 도구·재개·설정 우선순위 검증은 아니다. 관측기 결과의 INCONCLUSIVE를 전체 인증 실패로도 바꾸지 않는다.

## Residual-risk

이 forwarder는 응답을 모두 읽고 전달하므로 스트리밍 보존의 근거가 아니다. 성공한 짧은 요청만으로 모든 Claude 요청 정책이 gateway에서 지원된다고 볼 수 없다. 토큰 갱신 출처 확인 전에는 M0의 남은 게이트와 생산 활성화를 별도로 다룬다.
