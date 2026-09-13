# M0 — OAuth와 로컬 인증 공존 측정

Verdict: INCONCLUSIVE

## Claim

Claude Code 2.1.268에서 별도 `X-MoAI-Session-Token` 헤더와 로그인 상태의 Bearer 헤더가 같은 요청에 실렸다. 그러나 원래 Anthropic endpoint가 모든 해당 요청에 HTTP 429를 반환하여 성공한 inference와 OAuth refresh는 관측하지 못했다. T09 양성 게이트는 충족하지 않았다.

`ANTHROPIC_AUTH_TOKEN`에 로컬 시험 토큰을 넣는 대조군에서는 Bearer 값이 그 시험 토큰으로 대체되고 upstream은 HTTP 401을 반환했다. 이 대조군은 별도 credential 재주입 없이 같은 Authorization 헤더를 그대로 전달하는 방식에 한정한다.

## Evidence

실행 명령(현재 작업 트리, 제한된 포트 바인딩 권한을 허용한 뒤):

```sh
timeout 150 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/m0_probe.py
```

프로세스 exit 0. 이것은 하네스 종료 상태이며 게이트 PASS가 아니다. 두 Claude 자식은 각각 내부 제한에서 종료되었다.

`m0-result.json`을 Python `json`과 `collections.Counter`로 집계한 실제 출력:

```json
{
  "results": [
    {"mode": "custom", "timed_out": true},
    {"mode": "auth-token", "timed_out": true}
  ],
  "request_counts": {"custom": 14, "auth-token": 14},
  "statuses": {"custom": {"429": 14}, "auth-token": {"401": 14}},
  "custom_bearer_and_session_header": true,
  "auth_token_replaces_bearer": true
}
```

원본 관측 파일은 같은 디렉터리의 `m0-requests.json`과 `m0-result.json`이다. 요청 본문·토큰 원문은 기록하지 않았다. 인증 종류, 해시, 시험 토큰 일치 여부, beta 헤더, upstream 상태만 기록했다. `custom` 요청에는 `oauth-2025-04-20` beta가 있었고, API key 헤더는 없었다.

최초 sandbox 시도는 `ThreadingHTTPServer`의 bind에서 다음 오류로 실패했다. 그 시도는 upstream에 도달하지 않았다.

```text
PermissionError: [Errno 1] Operation not permitted
```

## Baseline-attribution

- `WT-unified-gateway`, HEAD `81c1d58f9`, SPEC 0.7.0.
- `claude --version` → `2.1.268 (Claude Code)`.
- 로그인 상태 조회는 `loggedIn: true`, `authMethod: claude.ai`, `subscriptionType: max`였다. 개인정보는 본 보고서에 복제하지 않았다.
- 임시 프로젝트, `--setting-sources ''`, `--strict-mcp-config`, 도구 없음, 고정된 짧은 시험 prompt를 사용했다.
- `ANTHROPIC_*` 상속 키를 제거하고 각 후보만 설정했다. upstream 목적지는 `api.anthropic.com`으로 고정했다. 별도 로컬 인증 헤더는 upstream 전달 전에 제거했다.
- 두 prompt 실행에서 client가 재시도하여 실제 POST 계수는 각각 14건이었다.

## Gaps

- 성공한 OAuth inference, refresh, 만료 후 재인증은 관측하지 못했다.
- 429 응답 본문은 저장하지 않아 사용량 제한의 구체 사유와 해제 시각은 판정하지 않는다.
- 응답을 전부 읽어 전달하는 최소 forwarder이므로 streaming 지연과 취소 보존을 검증하지 않았다.
- 설정 파일 env 우선순위·MCP 인증·구독 과금은 이 시험의 판정 대상이 아니다.

## Residual-risk

별도 헤더 후보가 유망하다는 관측만으로 OAuth passthrough를 구현하거나 catalog에 등록하지 않는다. `plan.md` M0와 `design.md` §2.2의 양성 게이트를 유지한다. 429를 인증 방식의 음성 판정으로 바꾸지 않는다. 성공 응답을 얻을 수 있는 상태에서 같은 게이트를 재측정해야 한다.
