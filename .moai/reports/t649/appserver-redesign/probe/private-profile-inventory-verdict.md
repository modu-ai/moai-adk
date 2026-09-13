# AS1: 독립 프로필 MCP 런타임 목록 확인

## 판정

| 여정 | 상태 | 시간 | 산출물 |
|---|---|---:|---:|
| 새 비공개 HOME → 공식 App Server → thread/start → MCP 목록 | PASS | 0.08초 | 3 |

```sh
python3 .moai/reports/t649/appserver-redesign/probe/private-profile-inventory.py
```

```text
account_type=null
thread_started=true
servers=[]
cursor_remaining=false
inference_started=false
success=true
process_exit=-15
profile_removed=true
```

## 기준

Codex CLI 0.154.0. t650 `internal/codexapp/client.go`의 Start 함수에서 인자를 직접 추출했다. 임시 디렉터리 0700을 CODEX_HOME/HOME/USERPROFILE에 지정하고 같은 소스의 환경 allowlist를 사용했다. 실제 t650 Go Client를 실행한 것은 아니며 동일 설정의 공식 바이너리를 Python 프로토콜 드라이버로 실행했다.

설치본에서 생성한 ClientRequest와 ListMcpServerStatusParams 스키마에 따라 `mcpServerStatus/list`를 threadId, detail=toolsAndAuthOnly, limit=100으로 요청했다. 서버 목록은 비어 있었고 후속 커서도 없었다. thread/start는 ephemeral=true, environments=[], read-only, approvalPolicy=never로 실행했다. 모델 턴은 시작하지 않았다.

정확한 override:

```text
approval_policy="never"
sandbox_mode="read-only"
web_search="disabled"
features.shell_tool=false
features.codex_hooks=false
features.hooks=false
features.plugin_hooks=false
features.plugins=false
features.apps=false
features.view_image=false
features.multi_agent=false
features.multi_agent_v2=false
tools.update_plan.enabled=false
tools.experimental_request_user_input.enabled=false
```

## 범위와 미검증

이 결과는 새 독립 프로필이 기존 사용자 MCP 목록을 상속하지 않았다는 런타임 목록 증거다. 모델 호출이 없으므로 모든 네이티브 도구의 등록·실행 차단, 인증 후 동일 상태, turn 시작 시 추가 서버 생성은 검증하지 않았다. 시작 알림은 목록 응답까지 관측된 것이 없었으며 이후 장기 관측을 한 것은 아니다.

토큰과 기존 인증 파일을 읽거나 복사하지 않았다. 개인 설정을 변경하지 않았다. 종료 후 임시 프로필이 제거된 것을 확인했다.

산출물: private-profile-inventory.py, private-profile-inventory-result.json, private-profile-inventory-verdict.md.
