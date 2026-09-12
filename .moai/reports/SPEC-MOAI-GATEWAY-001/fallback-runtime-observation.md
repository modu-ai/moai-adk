# Claude fallback 실제 클라이언트·로컬 mock 관측

## Claim

2026-09-11 19:00 Asia/Seoul 이후 기존 fallback_after19.py를 수정 없이 한 번 실행했다. 설치 Claude print mode에서 configured fallback과 명시 --fallback-model이 실제 요청 모델을 Astra에서 Sonnet 5로 바꾸는 것을 관측했다. 빈 fallbackModel overlay 경우에는 관측한 70초 동안 Astra만 요청했지만 timeout으로 종료되었으므로 성공으로 판정하지 않는다. availableModels에서 초기 Astra를 제외한 별도 경우는 최초 실제 요청이 Opus 5였다.

실제 공급자 응답·과금·gateway 제품·TUI의 PASS가 아니다. 모든 모델 요청은 격리된 local mock에 도착했다. mock은 PRIMARY gpt-6-astra에 **529**를 반환했다. 사전 메모의 503 시나리오를 실제로 실행한 것으로 기록하지 않는다.

## Evidence

실행 전 clock 도구:

```text
2026-09-11 09:53:45 UTC
2026-09-11 09:55:15 UTC
2026-09-11 09:57:37 UTC
2026-09-11 09:59:51 UTC
2026-09-11 10:00:07 UTC
```

각 대기는 60초 이하였고 마지막 clock 확인 뒤에만 실행했다. script 자체 NOT_BEFORE=10:00 UTC도 그대로 유지했다.

실행 명령(로컬 listener·격리된 설치 클라이언트 실행 승인 후):

```sh
/opt/homebrew/bin/timeout 440 python3 .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/fallback_after19.py
```

한 실행의 핸들 96199만 polling했다. 최종 exit 0과 출력:

```json
{"evidence": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/fallback-20260911T100024Z-8f92618d", "cases_observed": 5, "status": "OBSERVED_ONLY"}
```

종료 후 clock은 `2026-09-11 10:02:04 UTC`였다. outer script exit0은 모든 개별 case 성공을 뜻하지 않는다.

| case | 설정/인수 | 실제 순서·수신 횟수 | child exit / timeout | 해석 |
|---|---|---|---|---|
| configured-fallback | user fallbackModel=[claude-sonnet-5], overlay={} | gpt-6-astra ×3 → claude-sonnet-5 ×1 | 0 / false | 529 뒤 모델 전환 관측 |
| empty-overlay | 같은 user 설정 + overlay fallbackModel=[] | gpt-6-astra ×7 | -9 / true | 70초 내 Sonnet 요청0. timeout이므로 완전 차단/정상 종료 PASS 아님 |
| explicit-flag | empty overlay + --fallback-model claude-sonnet-5 | gpt-6-astra ×3 → claude-sonnet-5 ×1 | 0 / false | 명시 flag가 빈 fallback 배열보다 우선하는 동작 관측 |
| unknown-start-model | 초기 gateway-intentionally-unknown, empty overlay | 같은 unknown ×2 | 1 / false | stream:true 뒤 stream:false; mock404 뒤 명시 오류. 다른 모델 수신0 |
| policy-excluded-start-model | 초기 Astra, availableModels=[claude-sonnet-5], empty fallback | claude-opus-5 ×1 | 0 / false | 초기 explicit 모델도 allowlist Sonnet도 아닌 Opus 요청. 초기 모델·정책 enforcement 추가 필요 |

모든 path는 `/v1/messages?beta=true`였다. Astra 요청 max_tokens는 32000, Sonnet/Opus는 64000이었다. unknown의 두 요청 모두 32000이고 stream만 true→false였다. 요청별 raw SHA256은 results.json에 있다. raw 본문 자체를 이 관측기가 저장하지는 않았다.

정상 local 응답 3case의 stdout은 각각 다음 20바이트다.

```text
LOCAL_RESPONSE_ONLY
```

empty-overlay stdout은 0바이트다. unknown-start-model stdout은 다음과 같다.

```text
There's an issue with the selected model (gateway-intentionally-unknown). It may not exist or you may not have access to it. Run --model to pick a different model.
```

Astra를 실제로 시작한 세 case stderr는 604바이트의 unrecognized_model 안내였다. 설치 catalog가 이 ID를 모른다는 점, behavesAs/modelPicker 또는 modelOverrides 후보, unknown model의 context를 200k로 가정한다는 안내가 들어 있다. 임의 context override를 이번에 적용하지 않았다. policy-excluded-start-model stderr는 빈 문자열이다.

results.json의 `fallback_observed`는 **Sonnet 수신만** 센다. policy-excluded case의 false를 모델 전환 없음으로 해석하면 안 된다. 실제 요청 ID는 Opus 5였다.

각 case user settings.json의 전후 SHA256은 모두 다음과 같았다.

```text
0c0304ef3ce2af4cdb99ddf34f9809c838f857d5bbdc1d7dae51a20d7a1842b0
```

## Baseline-attribution

WT moai-proxy-unified / 기준 HEAD81c1d58f9. 실행 시 script가 기록한 시작 시각은 10:00:24 UTC다. 실행 전후 script SHA256은 동일하다.

```text
428271ff687b1c54a94e4a4e08dd1aa7f57aa3efa2b842acf7170896860ba7a3
```

실행 전 AST parse는 SYNTAX_OK였다. 실행 뒤 `/Users/goos/.local/bin/claude` symlink를 읽은 대상은 `/Users/goos/.local/share/claude/versions/2.1.268`이다. script는 PATH로 claude를 호출하며 실행 binary SHA를 별도로 캡처하지 않았다는 한계가 있다.

결과 디렉터리:

`.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/fallback-20260911T100024Z-8f92618d/`

results.json SHA256:

```text
afbf7d83b0c2b525b770850315acfee60e26929afa45ee0813d9ecf23e88ac05
```

이 파일과 다섯 case의 stdout/stderr 파일을 직접 읽었다. 각 case는 새 TemporaryDirectory HOME/CLAUDE_CONFIG_DIR/project, strict MCP, no-session-persistence, 합성 토큰과 로컬 서버를 사용했다. subprocess는 새 session으로 생성되고 timeout 시 자신의 process group만 Kill한 뒤 communicate/Wait한다. finally에서 process 종료 확인·server shutdown/server_close·thread join을 수행하는 기존 코드를 읽었으며 실행은 exception 없이 끝났다. global tmux나 다른 세션을 조작하지 않았다.

## Gaps

- empty overlay는 timeout이다. 70초 뒤에도 영구히 fallback이 없다는 주장, 사용자에게 정상 오류가 표시됐다는 주장, 전체 no-fallback PASS는 하지 않는다.
- 실제 TUI /model·s·Default, continue/resume, managed/project 설정 우선순위, 실제 gateway 설정 조립은 관측하지 않았다.
- 503·401·429·연결거부는 이번 script가 시험하지 않았다. 같은 provider 자동 재시도와 provider fallback을 구별해야 한다.
- 실제 provider inference/계정/과금은 호출하지 않았다. local mock의 응답 모델 이름은 공급자의 실제 모델 수용 증거가 아니다. 전체 네트워크 packet capture를 한 것은 아니므로 클라이언트의 모든 보조 통신0을 실측했다고 주장하지 않는다.
- 원본 사용자 설정이 아닌 임시 settings hash만 비교했다. 파일 권한/mtime의 전후 비교, 최종 port/PID 독립 census는 별도 기록하지 않았다.
- 제품 코드·script 수정이나 재실행은 하지 않았다. 전체 CI·SPEC 완료 판정은 PENDING이다.

## Residual-risk / 후속 결정 입력

빈 fallback overlay만으로 명시 flag를 무력화할 수 없다는 관측은 제품 argv 정책과 연결해야 한다. availableModels 기반의 초기 모델 검사는 별도 실제 요청 확인 없이 신뢰할 수 없다. observed Opus 대체 원인의 정확한 내부 분기는 이 실험만으로 확정하지 않는다. configured fallback과 explicitflag의 실측은 정책 결정 입력이며 사용자 설정을 조용히 제거하거나 명시 모델을 다른 provider로 대체할 권한을 새로 부여하지 않는다.
