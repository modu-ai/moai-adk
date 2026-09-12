# ToolSearch 후속 턴 E2E 기준선

## 판정

| 여정 | 상태 | 실행 시간 | 산출물 |
|---|---|---:|---:|
| Sol: ToolSearch → WebFetch tool_reference → 후속 답변 | FAIL (필수 도구 호출 미관측) | 4.22초 | 3 |

실행 프로그램이 정상 종료하고 표식을 출력했지만 ToolSearch를 호출하지 않았다. 따라서 보고된 HTTP 400의 재현 성공도, 문제 해결도 입증하지 못했다. 구현 변경 우선순위가 App Server 재설계로 바뀌어 추가 유료 호출은 중단했다.

## 근거

실행 명령:

```sh
python3 .moai/reports/t649/subagent400/e2e/toolsearch_probe.py --binary /Users/goos/go/bin/moai --model gpt-5.6-sol --output .moai/reports/t649/subagent400/e2e/baseline-sol.json
```

실제 결과의 검증 필드:

```json
{
  "model": "gpt-5.6-sol",
  "binary_sha256": "25b68beaadc919b91adab97213b76daa694b284fe3e297958e627c3842bc08c2",
  "tool_uses": [],
  "tool_references": [],
  "marker_seen": true,
  "errors": [],
  "success": false
}
```

테스트 종료 코드 1, 자식 프로그램 종료 코드 0. 정상 답변만으로 합격하지 않는 검증이 동작했다.

## 기준 환경

기존 설치 rc.8, SHA256는 위 JSON 참조. 합성 작업 디렉터리 `/tmp/moai-t649-e2e-project`; 기존 격리 Claude fixture 사용. hooks 비활성화, MCP 빈 설정, 기본 권한, 최대 4턴, 실행 상한 90초. 실제 사용자의 실패 세션과 인증은 수정하지 않았다. 원시 요청·토큰·사용자 대화는 저장하지 않았다.

## 미검증

- ToolSearch와 tool_reference가 생성되지 않아 문제의 후속 요청 미검증.
- Astra와 실제 서브에이전트 경로 미실행.
- 수정 후보 바이너리 미실행.
- 프로세스 그룹 정리는 시도했지만 종료 후 signal_permission_denied가 기록됨. 완전한 프로세스 정리를 입증하지 못함.

## 잔여 위험 및 재사용 조건

제품이 도구 탐색을 비활성화하는 조건에서는 프롬프트만으로 이 여정을 실행할 수 없다. App Server 통합 인수검증에서는 실제 탐색 도구가 노출되는 고정 fixture를 구성하고, 도구 참조 발생과 다음 턴 성공을 모두 판정해야 한다. 현재 harness는 이 두 조건을 필수로 판정하므로 단순 표식 답변을 성공으로 잘못 보고하지 않는다.

## 산출물

- `toolsearch_probe.py`: 제한된 네이티브 실행과 구조화된 판정
- `baseline-sol.json`: 안전한 실행 결과
- `verdict.md`: 기준선 판정과 한계
