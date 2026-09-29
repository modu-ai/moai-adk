# 재배포 후 실사용 오류 수집

## 관측 대상

- family: `f0579c71-d3e9-4f9b-a83c-8a5f6ddaef77`
- 로그: `/Users/goos/.moai/state/gateway-conversations/families/f0579c71-d3e9-4f9b-a83c-8a5f6ddaef77/native/debug/f0579c71-d3e9-4f9b-a83c-8a5f6ddaef77.txt`
- 설치 Build ID: `v3.2.0-rc.11-b45c81349-summary-isolation-dirty`
- 실제 gateway PID 35071: `/Users/goos/go/bin/moai internal-gateway`, 2026-09-15 01:13:26 시작.
- `lsof -a -p 35071 -d txt -F nsi`와 `stat`에서 실행 이미지와 설치 파일 inode가 모두 `3233053054`임을 확인. 구버전 실행 파일 사용으로 설명할 근거 없음.

## 오류 목록 — 수집 중

### E03: 메인 대화 요청 scope mismatch — 우선 처리

2026-09-15 01:28:41 KST, 두 번째 프로젝트 family `b702a146-4196-4e47-b15c-34ac6169c6a6`에서 메인 요청 자체가 실패했다. 두 debug 파일을 Ruby로 읽고 마지막 API 오류 주변 행을 추출하여 확인했다. 자동 요약에 한정된 장애가 아니다.

```text
2026-09-14T16:28:40.936Z [DEBUG] [API REQUEST] /v1/messages source=repl_main_thread:outputStyle:custom
2026-09-14T16:28:41.008Z [ERROR] API error (attempt 1/11): 400 400 {"error":{"message":"appserver_scope_mismatch","type":"invalid_request_error"},"type":"error"}
2026-09-14T16:28:41.010Z [ERROR] [engine] turn ended in error: API Error: 400 appserver_scope_mismatch
```

해당 요청 기록에서 오류까지 72ms이다. 이는 느린 토큰 생성 측정값이 아니라 빠르게 실패한 요청이다. 정확한 내부 거부 분기 및 앞선 요약 오류와의 인과관계는 아직 미확정이다. 사용자 프로세스 변경·재시작·수정 배포는 수행하지 않았다.

### E01: 자동 요약 scope mismatch 재발 — 미해결

한국 시간 01:14:50.311, 01:15:36.077에 `agent_summary` 요청이 HTTP 400 `appserver_scope_mismatch`로 실패했다. 오류 뒤 작업 에이전트의 도구 실행은 계속 성공했다. 직전 수정이 실제 대화형 요약 오류까지 해결했다는 판정은 철회한다. 실제 실패 요청의 원문·세부 내부 거부 지점은 아직 미확보다.

```text
2026-09-14T16:14:50.311Z [ERROR] API error (attempt 1/11): 400 400 {"error":{"message":"appserver_scope_mismatch","type":"invalid_request_error"},"type":"error"}
2026-09-14T16:15:36.077Z [ERROR] API error (attempt 1/11): 400 400 {"error":{"message":"appserver_scope_mismatch","type":"invalid_request_error"},"type":"error"}
```

### N01: 로컬 취소 — gateway 결함으로 집계하지 않음

```text
2026-09-14T16:13:52.593Z [DEBUG] [onCancel] source=local streamMode=responding
2026-09-14T16:13:52.593Z [ERROR] Error in API request: Request was aborted.
```

취소 이후 새 요청의 첫 바이트는 5.717초에 도착했다. 취소 동작의 구체적인 입력 주체까지 로그로 확정한 것은 아니다.

### E02: 생성된 jq 표현식 오류 — 실행 명령 확인 필요

01:16:47.470 Bash가 exit 1로 실패했다. 후속 훅에 `failed to parse jq expression`, `unexpected token "main"`이 기록됐다. 해당 시각의 native 자식 transcript에서 생성된 `for branch in main develop; do gh api ... --jq ...; done` 명령 자체에 과도한 따옴표가 들어간 것을 확인했다. gateway가 내용을 변조했다는 근거는 확보되지 않았다. 모델 생성 명령 오류로 별도 분류한다.

### P01: 첫 바이트 대기 경고 — 원인 미확정

01:15:26.883 및 01:16:09.917에 `no stream chunk 30.0s after request sent`가 기록됐다. 모델 생성 시간과 같은 owner의 작업·요약 대기 경합을 분리해 확인할 필요가 있다. 01:17:00 관측 시점의 도구 종료는 성공 33회, 실패 1회, 도구 실행 최댓값 2.390초다. 01:16:30.915에 E01도 추가 재발해 누적 3회다.

### D01: 세부 거부 사유 누락 — 일괄 수정 시 계측 보완 후보

공개 오류와 비공개 rejection 로그는 `appserver_scope_mismatch`로 집계되며 실제 실패 요청의 원문 및 내부 세부 분기까지는 알 수 없다. 원문·자격 증명을 기록하지 않고 분류 결과, 거부 지점의 고정 코드, 건수·길이 등 최소 메타데이터로 원인을 확인할 방법이 필요하다. 이 기록은 새 계측 구현 완료 주장이 아니다.

## 최신 집계

01:19:38 KST 관측: HTTP 오류 7회, 자동 요약 성공 0회, 도구 성공 74회·실패 1회. 첫 바이트 30초 초과 경고는 01:15:26.883, 01:16:09.917, 01:19:10.139, 01:19:10.926에 관측했다. 아직 작업 완료 판정은 하지 않았다.

01:21:49 관측: HTTP 오류 10회, 자동 요약 성공 1회, 도구 성공 91회·실패 2회. 성공한 요약은 01:20:30.624에 `Testing web and session packages`를 반환했으며 별도 `:summary:` owner의 idle barrier도 확인했다. 따라서 모든 요약 요청이 분리되지 않는다고 일반화하면 안 된다.

### A01: MoAI 웹 스크립트 검사 실패 — 브랜치별 차이 확인 필요

01:21:01.879에 자식이 `/tmp/moai-app-main-f0579c71.js`를 Node로 평가해 `ReferenceError: stampRefreshed is not defined`를 얻었다. 동일 자식의 다음 검사 `/tmp/moai-app-develop-f0579c71.js`는 `evaluation=ok`였다. 원본 스냅샷 검사를 수행한 자식의 증거이며 이 수집자가 Node 검사를 재실행한 것은 아니다. 주 checkout의 `internal/web/assets/app.js:554`와 수리 worktree의 같은 파일 727행에는 해당 listener 위치 차이가 관측된다. 기존 수정의 통합 여부를 확인하기 전에 중복 패치하지 않는다.

## 추가 프로젝트

- family: `b702a146-4196-4e47-b15c-34ac6169c6a6`
- 프로젝트: `/Users/goos/MoAI/mo.ai.kr`
- 01:21:48 관측: API 오류 0회, 도구 성공 13회. 제목 생성 및 메인 요청의 첫 바이트 로그는 4.024초·5.690초다(동시 요청이므로 각 모델과 임의로 일대일 대응하지 않음).
- 이 프로젝트의 업무 코드 문제는 수정 범위에서 제외하고 GPT/gateway 및 MoAI-ADK 원인의 오류만 수집한다.

01:23:07 추가 관측: mo.ai.kr에서도 자동 요약의 HTTP 400 `appserver_scope_mismatch`가 2회 발생했다. 01:23:03.130 오류는 `agent_summary` 뒤이며 이후 `agent:builtin:Explore` 요청이 계속 진행됐다. 도구 성공 38회, 도구 오류 0회. 같은 시점 첫 family는 API 오류 12회, 요약 성공 1회, 도구 성공 102회·실패 2회다. E01은 서로 다른 프로젝트에서 재현된 공통 gateway 이슈로 묶는다.

## 수집 원칙

## 성능 관측 — 2026-09-15 01:27 KST

### Claim

긴 첫 바이트 대기가 관측되어 응답 지연을 정상으로 일괄 판정할 수 없다. 순수 디코딩 TPS는 미측정이다. ADK 자식의 원본 App Server 요청 5건에서는 요청 구간 전체를 분모로 한 출력 처리량이 39.50~53.43 tokens/s였다. 화면에 표시되는 텍스트만의 속도와 다르다.

### Evidence

두 debug 파일에 Ruby `scan(/first byte after (\d+)ms/)`를 적용하고 오름차순 정렬, `(n-1)*p`의 반올림 인덱스로 분위수를 계산했다. 관측 출력:

```text
f0579c71 samples=113 p50_ms=80 p95_ms=15982 max_ms=46993 over_30s=4 subset_ge_1s=45 subset_p50_ms=8224
b702a146 samples=90 p50_ms=86 p95_ms=16066 max_ms=99232 over_30s=2 subset_ge_1s=33 subset_p50_ms=9654
```

`sqlite3 -readonly .../logs_2.sqlite`에서 thread `01a0a0b2-94af-7793-9a70-1baf5499ee82`의 `responses_websocket.stream_request` 시각을 읽고, 해당 rollout의 `event_msg/token_count` 직전 요청과 연결했다. `last_token_usage.output_tokens / (token_count 시각 - 요청 시각)` 계산 출력:

```text
end_UTC       request_seconds output reasoning output_per_request_second
16:22:16.477  16.960          670    182       39.50
16:22:33.429  16.943          708    615       41.79
16:23:03.770  30.337          1271   842       41.90
16:23:50.875  47.098          2232   2140      47.39
16:24:44.652  53.773          2873   516       53.43
```

### Baseline-attribution

실제 운영 중 두 family의 로그 스냅샷이며 합성 부하 테스트가 아니다. 토큰 처리량 표는 첫 family의 `gpt-5.6-sol`, `reasoning_effort=high` 자식 thread에만 해당한다. 마지막 표본 input=159080, cached_input=156544였다. 원본 전송 로그에는 `auth_connection_reused=true`가 기록되어 있다.

### Gaps

첫 바이트 표본은 메인·자식·요약 및 버퍼 전달 구간이 혼재한다. 1초 이상 부분집합은 단순 시간 필터이며 실제 생성 요청으로 분류한 결과가 아니다. 토큰별 도착 시각, 순수 디코딩 TPS, 동일 프롬프트의 직접 Codex 대조군, 두 번째 프로젝트의 원본 TPS는 미측정이다. 모델·gateway·대기열별 지연 기여도는 확정하지 않았다.

### Residual-risk

요약 400과 긴 대기가 함께 관측됐지만 인과관계는 입증되지 않았다. 짧은 첫 바이트 중앙값으로 대화 응답이 빠르다고 판정하거나 요청 구간 TPS를 화면 텍스트 TPS로 표시하면 잘못된 결론이 된다. 이번 관측에서는 제품 코드와 사용자 프로세스를 변경하지 않았다.

동일 오류는 횟수와 시각을 묶는다. 모델 대기, 도구 실행 실패, 정상적인 보호 검사, HTTP/스트리밍 실패를 구분한다. 사용자 작업·프로세스·pending 도구를 변경하지 않고 관측한다. 원인을 확정하지 않은 증상은 가설로 남기며, 일괄 수정 전 실제 요청 구조를 재현해야 한다.
