# AS2: Claude 지연 도구 스키마 계약 실증

## 판정

**PASS: 초기 요청에 없는 도구 스키마가 ToolSearch 후속 요청에 추가됨을 실제 Claude Code에서 확인했다.** 따라서 초기 HTTP 요청에서 모든 스키마를 미리 확보하여 Codex thread/start에 한 번만 등록하는 방안 A는 이 조건에서 성립하지 않는다.

| 여정 | 결과 | 시간 | 산출물 수 |
|---|---|---:|---:|
| Claude → 로컬 모의 Anthropic → ToolSearch → tool_reference와 새 스키마 → 최종 응답 | PASS | 0.33초 | 4 |

## 실행 근거

```sh
/Users/goos/.local/bin/claude --version
# 2.1.269 (Claude Code)
python3 .moai/reports/t649/appserver-redesign/probe/claude-tool-surface.py
```

```text
exit_code=0
final_marker_seen=true
success=true
requests=2
```

| 비교 | 첫 요청 | ToolSearch 이후 두 번째 요청 |
|---|---|---|
| 전체 도구 수 | 12 | 13 |
| mcp__fixture__echo 스키마 | 없음 | 있음, defer_loading=true |
| DeferredToolPlaceholder | 있음, defer_loading=true | 있음, defer_loading=true |
| tool_reference | 없음 | mcp__fixture__echo |

기존 도구의 구조 필드(name/type/defer_loading/input_schema)는 변경되지 않았고 신규 도구 하나가 추가됐다. 즉 defer_loading 플래그만 바뀐 것이 아니다. 활성화 이후에도 새 도구의 defer_loading 값은 true였다.

## 기준 환경

Claude Code 2.1.269를 직접 실행했다. ENABLE_TOOL_SEARCH=true, 임시 CLAUDE_CONFIG_DIR와 임시 cwd, 합성 API 키, 127.0.0.1 임시 HTTP 서버, 고정 로컬 stdio MCP fixture를 사용했다. hooks 비활성화, strict MCP 설정, 기본 권한, 3턴 상한, 외부 실행 상한 45초였다. 실제 모델 서비스에 요청하지 않았다. MCP fixture는 echo 스키마를 제공했지만 echo 실행은 요청하지 않았다.

모의 서버가 첫 응답으로 ToolSearch(query=select:mcp__fixture__echo,max_results=1)를 요청했고, 실제 Claude가 검색 결과를 만든 뒤 두 번째 요청을 보냈다. 두 번째 응답은 고정 표식이었다. 프로세스 종료와 그룹 정리를 확인했고 오류는 없었다.

## 미검증 및 잔여 위험

이 결과는 CLI의 실제 도구 스키마 공개 계약 증거다. 실제 모델이 올바르게 ToolSearch를 선택하는지나 MoAI gateway의 동작을 증명하지 않는다. WebFetch와 추가 런타임 MCP 서버 변형은 시험하지 않았다. 원시 프롬프트·인증값은 저장하지 않고 요청의 경로, 구조화된 도구 스키마, tool_reference 이름만 저장했다.

## 산출물

- claude-tool-surface.py
- claude-tool-surface-result.json
- claude-tool-surface-summary.json
- claude-tool-surface-verdict.md


## ToolSearch 비활성화 대조 실험

```sh
python3 .moai/reports/t649/appserver-redesign/probe/claude-tool-surface-disabled.py
```

exit 0, success=true, 0.35초, 요청 1회. 초기 요청부터 fixture의 전체 스키마가 있었으며 ToolSearch는 없었다.

| 항목 | ENABLE_TOOL_SEARCH=true | ENABLE_TOOL_SEARCH=false |
|---|---:|---:|
| 초기 도구 수 | 12 | 22 |
| 초기 fixture 스키마 | 없음 | 있음 |
| ToolSearch 제공 | 있음 | 없음 |
| 구조 필드 도구 배열의 compact JSON UTF-8 바이트 | 10,998 | 17,295 |
| input_schema만 compact JSON UTF-8 바이트 합계 | 10,534 | 16,476 |

바이트 수는 저장한 name/type/defer_loading/input_schema 구조를 동일한 compact JSON 방식으로 다시 직렬화하여 측정했다. 원래 요청 전체의 wire byte나 도구 설명 길이를 측정한 값이 아니다. 모의 count_tokens는 고정값 1000이므로 토큰 증가나 비용을 주장하지 않는다.

### 설계 선택에 주는 근거

- A: ToolSearch를 유지하면서 첫 HTTP 요청의 정의만 선등록하는 방식은 이 픽스처에서 불가능하다.
- C: ToolSearch를 끄면 이 픽스처에서는 초기 전체 스키마를 확보할 수 있다. 그러나 ToolSearch 자체가 사라지고 초기 구조 바이트가 6,297 늘어난다. 다른 동적 연결까지 전체 선등록이 보장된다는 증거는 아니다.
- B: 고정 dispatcher는 도구 이름과 인수, 최신 실제 스키마를 별도로 전달·검증하여 ToolSearch 흐름을 보존하는 설계 후보다. App Server 왕복은 별도 실증에서 확인했지만, N개 도구 스키마가 하나의 공통 스키마로 바뀌었을 때 선택·인수 정확도와 추가 왕복의 영향은 검증하지 않았다.

추가 산출물: claude-tool-surface-disabled.py, claude-tool-surface-disabled-result.json, claude-tool-surface-comparison.json.
