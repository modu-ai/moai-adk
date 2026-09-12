# t649 세션 간 탐색 오류 수정

## Claim
GPT 전용 CLAUDE_CONFIG_DIR 아래의 sessions 목록이 기존 mo.ai.kr 프로필과 분리되어 lead 탐색이 실패했다. 새 GPT 세션은 선택한 프로필의 sessions 디렉터리만 심볼릭 링크로 공유하도록 수정하고 로컬 rc.8 설치를 완료했다.

## Evidence
명령: `go test -p 1 ./internal/cli -run 'TestGateway|TestGPTBinding|TestDefaultLaunch' -count=1 -timeout=120s`

```text
ok  github.com/modu-ai/moai-adk/internal/cli  2.782s
```

실제 Claude 2.1.269 + 로컬 합성 응답으로 ListAgents를 호출한 대조 실험: isolated.json lead_seen=false; shared.json lead_seen=true. 두 실행 모두 종료 코드 0이다. success 필드는 실험 흐름 완료를 뜻하며 lead 탐색 성공과 구별한다.

설치된 moai로 GPT 구독 + ListAgents 실호출: live-discovery.json 참조. 종료 코드 0, lead_seen=True, 최종 결과 `FOUND`.

```text
This session is mo-ai-kr-2b [d0aee1] — the name other sessions use to message it (it is not listed below; a message to it would be a message to yourself).

Peer sessions (6):
  lane-5 [318726]  ·  interactive  ·  busy  ·  started 2d ago
  lane-1 [7160f3]  ·  interactive  ·  busy  ·  started 2d ago
  lead [a250a7]  ·  interactive  ·  busy  ·  started 2d ago
  lane-3 [abe6c3]  ·  interactive  ·  idle  ·  started 2d ago
  lane-4 [a9a13f]  ·  interactive  ·  busy  ·  started 2d ago
  lane-2 [e1c6c9]  ·  interactive  ·  busy  ·  started 27m ago
```

## Baseline-attribution
워크트리: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
HEAD: 81c1d58f9cf7045594ee61d5e4ff380948ce9eba (기존 미커밋 변경 포함)
설치: /Users/goos/go/bin/moai
BuildID: v3.2.0-rc.8-t649-peer-registry-711481c015fd
SHA256: 06dd7f8fdc3045f9586a2c9687a0e48034d19e6ba229b3592f92fe917180939b
빌드 명령과 소스 해시는 build.json, build.log 참조. 인증·대화 기록은 공유하지 않는다. cc/glm 경로는 이번 변경 대상이 아니다.

## Gaps
실제 lead에 메시지를 보내거나 배차/ACK를 검증하지 않았다. 네이티브 세션 탐색과 GPT 도구 호출 왕복까지 확인했다. 대화형 PTY 검사는 55초 안에 목표 문자열을 확인하지 못했다(interactive-result.json); 이후 print 모드의 실제 GPT 도구 호출에서 성공을 확인했다. Windows 런타임은 실행하지 않았다.

## Residual-risk
기존 실행 중인 GPT 세션은 분리된 등록 경로를 계속 사용하므로 새 세션으로 재시작해야 한다. 기존 분리 디렉터리를 가진 대화의 resume은 안내 오류를 내며 자동 이동하지 않는다. 다른 프로필끼리의 목록을 전역 통합하지 않는다. Windows에서 심볼릭 링크 생성 권한은 별도 CI 검증이 필요하다.

## 로컬 적용
`moai gpt -p mo.ai.kr -f lane-6`로 새로 시작한다. 기존 lead는 같은 프로필에서 실행 중이므로 재시작이 필요하지 않다. `ListAgents로 lead를 확인한 뒤 lead에게 카드 배차를 요청해 줘`라고 요청한다. 발송 대상은 현재 ListAgents 결과에서 얻고 UUID·소켓 경로를 도구의 이름 대신 넣지 않는다.
