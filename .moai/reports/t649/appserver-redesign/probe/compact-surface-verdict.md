# AS4: Claude /compact 구조 실증

## 판정

| 여정 | 결과 | 시간 | 산출물 |
|---|---|---:|---:|
| 실제 Claude seed → 같은 합성 세션 resume /compact → 완료 이벤트 | PASS | 합계 0.87초 | 3 |
| HTTP 요청만으로 압축 시작을 안정적으로 구별 | GAP | — | — |

실제 Claude Code 2.1.269, 비공개 임시 HOME/cwd, 로컬 모의 Anthropic 서버와 합성 인증을 사용했다. 유료 API 요청과 사용자 세션 변경은 없다. CLI 실행은 두 번뿐이다.

## 실행과 근거

```sh
python3 .moai/reports/t649/appserver-redesign/probe/compact-surface.py
```

첫 실행은 짧은 합성 메시지에 고정 답변을 받았다. 두 번째는 첫 실행의 session_id를 메모리에서 받아 --resume <id> /compact를 실행했다. 두 실행 모두 exit 0이고 서버 요청은 각 1회였다.

압축 실행의 stream-json에는 다음 구조 이벤트가 실제 발생했다:

```json
{"type":"system","subtype":"compact_boundary"}
```

그 후 result subtype=success, is_error=false를 관측했다. 이는 CLI 출력에서 완료된 압축 경계를 기계적으로 관측할 수 있다는 증거다. 압축 요청 전에 HTTP gateway가 같은 사건을 안다는 증거는 아니다.

## HTTP 구조 비교

두 요청 모두 `/v1/messages?beta=true`였고 최상위 키 집합이 같았다:

```text
context_management,max_tokens,messages,metadata,model,output_config,stream,system,thinking,tools
```

model=claude-opus-5, max_tokens=64000, stream=true, metadata 키=user_id, system text 블록 길이도 동일했다. context_management는 두 요청 모두 clear_thinking_20251015/keep=all이며 압축을 명시하는 설정은 아니었다.

허용 목록(Content-Type, anthropic-beta, anthropic-version, x-app)에서 압축 전용 헤더는 확인되지 않았다. 두 번째 요청에는 context-1m beta 항목이 빠졌으나 이 차이를 압축의 고유 식별자로 사용할 근거는 없다. 메시지는 seed의 2개에서 압축의 4개로 늘었으며 일반 대화에서도 가능한 구조다. 프롬프트 문구를 이용한 분류는 만들지 않았다.

## 미검증 및 설계 의미

output_config와 thinking의 값, 허용 목록 밖 헤더 값은 캡처하지 않았다. 따라서 가능한 모든 HTTP 신호가 없다고 단정할 수 없다. 기록된 구조만으로는 안정적인 압축 시작 판별을 입증하지 못했다. 사용자 자동 압축이나 도구 실행 중 압축은 검증하지 않았다.

CLI의 compact_boundary 이벤트를 제어 채널에서 받아 후속 세대 전환을 처리하는 후보는 있지만, interactive 환경에서 gateway까지 전달하는 제품 연결과 이벤트 순서는 별도 검증이 필요하다. HTTP 본문 문자열만 보고 Codex thread/compact/start를 호출하는 인터셉터는 구현하지 않았다.

프로세스 종료와 그룹 정리 오류가 없었다. 원시 프롬프트·인증값·session_id를 저장하지 않았다. 요청 키, 제한된 비밀 아닌 헤더, 메시지 역할/블록 종류, CLI 이벤트 종류만 기록했다.

산출물: compact-surface.py, compact-surface-result.json, compact-surface-verdict.md.
