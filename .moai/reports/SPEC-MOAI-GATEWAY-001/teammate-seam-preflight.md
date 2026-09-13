# TEAMMATE 실행 연결점 사전 조사

## Claim

설치된 Claude 바이너리에 teammate 명령 환경 변수를 읽는 코드가 존재한다. 실제로 override가 실행되거나 pane별 안전한 bootstrap을 전달한다는 판정은 아직 아니다. 제품 구현·활성화 전에 실제 TUI와 로컬 mock으로 관측할 후보로 기록한다.

## Evidence

2026-09-11 부모가 실행한 읽기 전용 명령:

```text
$ command -v claude
/Users/goos/.local/bin/claude
$ ls -l /Users/goos/.local/bin/claude
lrwxr-xr-x@ 1 goos  staff  48 Sep 11 06:04 /Users/goos/.local/bin/claude -> /Users/goos/.local/share/claude/versions/2.1.268
$ rg -a -o 'CLAUDE_CODE_TEAMMATE_COMMAND|CLAUDE_CODE_SUBAGENT_MODEL_FORCE|CLAUDE_CODE_SUBAGENT_MODEL' /Users/goos/.local/share/claude/versions/2.1.268 | sort -u
CLAUDE_CODE_SUBAGENT_MODEL
CLAUDE_CODE_SUBAGENT_MODEL_FORCE
CLAUDE_CODE_TEAMMATE_COMMAND
```

각 exit 0. Python bytes 검색에서 TEAMMATE_COMMAND는 3회 나타났고, 두 번째 위치 161608464 근처에서 상수 WSr로 export된다. WSr 이름을 따라 찾은 180196106 근처의 함수는 process.env[WSr]를 읽어 명시 값이 있으면 command/prefixArgs 경로로, 없으면 현재 바이너리 경로로 분기한다. 같은 이름의 다른 모듈 지역 변수는 관련 근거에서 제외했다. 클라이언트 프로세스를 실행하지 않았다.

공식 [agent teams 문서](https://code.claude.com/docs/en/agent-teams)를 현재 web 도구로 열었다. 문서는 interactive 세션에서 named Agent 호출이 teammate를 만들고, print 모드는 일반 subagent로 동작한다고 설명한다. 모델 선택과 allowlist에 따른 대체 동작도 설명하므로 pane argv만으로 실제 모델을 판정할 수 없다. 해당 페이지를 확인한 범위에서 override 변수의 지원 보장을 얻은 것은 아니다.

## Baseline-attribution

로컬 읽기 대상은 위 2.1.268 경로다. 이 파일의 코드 존재 관측을 이전 버전 또는 다른 플랫폼 동작으로 확대하지 않는다. 제품 WT HEAD는 81c1d58f9다.

## Gaps

실제 override 실행, shell quoting, argv·환경 수신, 실제 pane ID·수명·재개, profile·MCP·permission 보존, 모델·provider 요청은 미검증이다. 팀 기능이 실험 상태라는 공식 문서 경계도 유지한다.

## 후속 관측

19:00 Asia/Seoul 이후 독립 tmux server·임시 HOME/config·로컬 mock·정리 보장 아래 실행한다. override helper는 우선 전달 인수와 합성 환경의 허용 목록만 기록하고 종료하도록 하여 실제 자식의 외부 송신을 만들지 않는다. native Agent schema를 실제 요청에서 확인하고 named 호출의 helper 실행·pane ID를 관측한다. 관측 후에야 single-use bootstrap·A/B/C 소유권 설계를 구체화한다. PATH의 tmux/claude를 가로채는 제품 wrapper로 대체하지 않는다.

## Residual-risk

바이너리 내부 연결점은 공식 안정 API라는 보장이 없다. 실제 호출 성공도 여러 lead 사이의 권한 분리나 부모 종료 처리를 증명하지 않는다. 지원 seam이 없거나 안전한 전달을 구성하지 못하면 현재 hybrid 활성화 거절을 유지하고 설계 대안을 검토한다.
