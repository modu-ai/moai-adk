# Native fork 부모 이력 결속 실증

## Claim

이 환경에서 native fork 자식이 시작되지 않아 부모 이력 보존 여부는 검증하지 못했다. 스키마 언급과 런타임 가용성을 구분한다.

| 여정 | 상태 | 시간 | 산출물 |
|---|---|---:|---:|
| Agent(subagent_type=fork) → 자식 부모 이력 비교 | FAIL / 가용성 차단 | 4.18초 | 3 |

## Evidence

```sh
python3 .moai/reports/t649/appserver-redesign/probe/fork-history-shape.py
```

설치본이 첫 실제 요청에 보낸 Agent input_schema의 model 설명에는 subagent_type: "fork"가 언급되어 있었다. 이를 확인한 후 모의 응답에서 Agent를 한 번 요청했다. 실제 런타임 결과:

```json
{"type":"tool_result","content":"Agent type 'fork' not found. Available agents: claude, Explore, general-purpose, Plan, statusline-setup","is_error":true,"tool_use_id":"toolu_fixture_fork_RANDOM_UNIQUE"}
```

HTTP 요청은 2개 모두 부모 요청이었다. 자식 agent-id 헤더가 없었고 CLI의 child parent_tool_use_id도 없었다. CLI exit 0은 모의 후속 텍스트 응답의 완료이며 fork 성공이 아니다. harness exit 1, success=false를 유지했다.

## Baseline-attribution

실제 Claude Code 2.1.269. 임시 HOME/CLAUDE_CONFIG_DIR/cwd, 합성 API 키, 로컬 모의 API, 고정 MCP fixture, hooks 비활성화. 외부 모델 호출은 0회. 각 실행 45초 제한, 프로세스 그룹 정리 오류 없음. 이전 일반 Agent 증거는 보존했다.

## Gaps

native fork 자식이 없어 self Agent tool_use ID 보존, canonical parent prefix, assistant 블록 변형, 자식 분기 ID 결속은 모두 미검증이다. 이 버전·빈 프로필에서 해당 subtype이 인식되지 않은 사실만 확인했다. feature gate나 지원 조건은 이번 probe에서 조사하지 않았으므로 전체 Claude 제품이 fork를 지원하지 않는다고 일반화하면 안 된다.

## Residual-risk

일반 Explore Agent의 식별 헤더 결과를 native fork 부모 이력 계약으로 대체할 수 없다. 스키마에 문자열이 있다는 사실만으로 실제 가용성을 선언하면 안 된다. 요청에 따라 추가 재시도 없이 종료했다.

산출물: fork-history-shape.py, fork-history-shape-result.json, fork-history-shape-verdict.md.
