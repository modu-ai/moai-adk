# Agent 요청 식별 필드 실증

## 판정

| 여정 | 결과 | 시간 | 산출물 |
|---|---|---:|---:|
| 실제 Claude Agent → 자식 응답 → 부모 응답, HTTP 식별 비교 | PASS | 0.92초 | 3 |

```sh
python3 .moai/reports/t649/appserver-redesign/probe/agent-identity.py
```

Claude Code 2.1.269. CLI exit 0, HTTP 요청 4개, 최종 부모 표식 관측, 오류 없음.

## 직접 관측

모의 서버가 첫 응답에서 실제 Agent 도구를 요청했다. 자식 작업은 Explore 유형이며 다른 도구 없이 표식만 답하도록 했다. 다음 HTTP 응답으로 자식 표식을, 이후 부모 표식을 반환했다. CLI 자식 assistant 이벤트의 parent_tool_use_id는 합성 도구 ID toolu_fixture_agent였다.

| HTTP 요청 | X-Claude-Code-Session-Id | x-claude-code-agent-id |
|---|---|---|
| 1 부모 | 동일 합성 UUID | 없음 |
| 2 자식 | 동일 합성 UUID | af11571aa120fb7db |
| 3 부모 | 동일 합성 UUID | 없음 |
| 4 부모 | 동일 합성 UUID | 없음 |

metadata.user_id는 JSON 문자열이며 account_uuid/device_id/session_id 키를 가진다. 값 중 session_id만 저장했다. 그 값은 부모·자식에서 모두 같았다. 따라서 session_id만으로 부모·자식을 분리할 수 없다. 이 샘플에서는 x-claude-code-agent-id라는 명시적 헤더가 자식을 구분했다. 프롬프트 내용 추측을 사용하지 않았다.

## 기준과 제한

임시 CLAUDE_CONFIG_DIR/cwd와 합성 인증, 로컬 모의 Anthropic 서버, 고정 로컬 MCP fixture, hooks 비활성화, 3턴 제한, 외부 45초 제한을 사용했다. 외부 모델 추론이나 실제 웹 도구 실행은 없다. 그룹 정리 오류도 없다. 계정 UUID·device_id 값·인증 헤더·원시 프롬프트를 저장하지 않았다.

자식 한 개의 요청 한 번만 관측했다. 같은 자식의 여러 턴·resume·병렬 형제·중첩 Agent·auto-compaction에서 식별자가 안정적인지 미검증이다. 부모 요청은 초기+후속에서 같은 세션 ID를 유지했다. 네 번째 요청의 내부 생성 이유는 이 구조 캡처만으로 확정하지 않았다.

헤더는 클라이언트가 전송한 값이며 그 자체가 인증 증거가 아니다. 제품은 인증된 family/session 범위에 agent-id를 결속하고 잘못된 형식·충돌·교차 세션 사용을 거절해야 한다. 이 결과는 식별 필드 존재의 근거이지 구현 완료나 모든 버전의 보장이 아니다.

산출물: agent-identity.py, agent-identity-result.json, agent-identity-verdict.md.
