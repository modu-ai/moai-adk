# t649 실제 E2E 1차 실행 보고

## Claim

세 launcher의 실제 Claude Code 모델 선택기에서 제공자별 Custom model 목록과 s 세션 한정 선택 확인을 관측했다. GPT-6 Astra 실제 답변은 HTTP 400으로 실패했다. 따라서 GPT 사용 가능 완료 판정은 내리지 않는다.

| 사용자 여정 | 판정 | 관측 소요(초) | 결과 파일 수 | 증거 |
|---|---|---:|---:|---|
| moai cc /model | PASS | 25.66 | 1 | [picker-cc-bootstrap.json](picker-cc-bootstrap.json) |
| moai glm /model | PASS | 25.68 | 1 | [picker-glm-switch.json](picker-glm-switch.json) |
| moai gpt /model | PASS | 25.69 | 1 | [picker-gpt-bootstrap.json](picker-gpt-bootstrap.json) |
| GPT-6 Astra 실제 답변 | FAIL | 0.87 | 1 | [answer-first.json](answer-first.json) |
| GPT Read 도구 왕복 | BLOCKED | 미실행 | 0 | 선행 답변 실패 |
| GPT 새 프로세스 resume | BLOCKED | 미실행 | 0 | 선행 답변 실패 |

## Evidence

cc 목록: claude-opus-5, claude-sonnet-5. glm 목록: glm-5.3, glm-5.3-flash. gpt 목록: gpt-6-astra, gpt-5.6-sol, gpt-5.6-terra, gpt-5.6-luna. 각 번호 행의 모델 집합을 정확히 대조했다. GLM은 flash에서 glm-5.3으로 실제 선택된 ID까지 추출했다. GPT 선택기 결과는 라벨과 Custommodel 문자열이 붙는 파서 오류를 수정하여 저장된 행을 재분석했으며 JSON에 이를 명시했다.

실행 명령은 [reproduce.txt](reproduce.txt), 세부 인수는 [gateway_probe.py](gateway_probe.py), 초기화 범위는 [claude_fixture.py](claude_fixture.py)에 있다.

```text
/tmp/moai-t649-implementation gpt status
GPT: logged in
/Users/goos/.local/bin/claude --version
2.1.268 (Claude Code)
```

실제 답변 결과 파일의 발췌:

```json
{"exit_code":1,"elapsed_seconds":0.87,"error_categories":["400 Bad Request","API Error"],"assistant_models":["<synthetic>"],"marker_in_final_result":false,"success":false}
```

최종 선택기 세 실행은 각각 `picker_exact_allowlist=true`, `session_only_selection_confirmed=true`, `cleanup_errors=[]`였다. 선택기 프로세스는 검사 완료 뒤 제한 시간에 그룹 종료했으므로 exit=-9는 답변 프로세스 종료 코드와 구별한다.

## Baseline-attribution

2026-09-12 macOS, WT-unified-gateway HEAD 81c1d58f9. /tmp/moai-t649-implementation SHA-256 9f2a7b217d731b0c6b8900391c94270ccade5b937c1a815daaa084053789a479. Claude Code 2.1.268. GPT status 관측값: GPT: logged in. 이 바이너리는 이후 수정 소스와 같다고 간주하지 않는다. [baseline.json](baseline.json)

## Gaps

실제 GPT 답변은 1회만 요청했다. 도구 호출과 새 프로세스 resume는 첫 답변 성공에 의존하므로 실행하지 않았다. HTTP 400의 upstream/변환기 발생 위치는 판정하지 못했다. 로그인 완료는 모델 사용 권한이나 정상 응답의 증거가 아니다. 기본 hook·MCP 설정, 관리 설정, 서브에이전트 요청, 입력한 외부 제공자 모델의 서버 거절, cc/GLM 실제 추론, Windows는 이번 E2E에서 검증하지 않았다.

## Residual-risk

선택기 PASS는 화면 목록과 s 선택의 증거다. 실제 요청 라우팅·GPT reasoning 왕복의 증거가 아니다. Default 행은 남아 있으며 해당 행의 실제 요청 모델은 측정하지 않았다. 최초 시작은 테마·API key 확인·bypass 경고에 막혔다. 최종 검증은 새 합성 작업 디렉터리와 신규 private family만 초기화한 wrapper를 사용했다. 사용자 전역 설정이나 인증 저장소를 수정하지 않았다. --permission-mode default, disableAllHooks, 빈 MCP 목록으로 조건을 제한했고 bypass 승인은 하지 않았다.

독립 코드 검토 결과는 [independent-review.md](../independent-review.md)를 별도로 참조한다. 그 결함과 이번 실제 HTTP 400의 인과관계는 입증되지 않았다.

## 추가 진단 증거

추가 진단: /tmp/moai-t649-safe-diagnostic (SHA-256 4e7bb30b2fef34fadb0345e8272dd1b16dc356097235f06e7c4bff65610bfb5c)으로 조율된 실제 답변 1회를 추가 실행했다. 1.75초 후 같은 HTTP 400을 관측했고 T649_DIAG 표시는 PTY로 전달되지 않았다. answer-safe-diagnostic.json에 별도 기록했다. 초기 바이너리 1회와 합쳐 실제 답변 시도는 총 2회다.
