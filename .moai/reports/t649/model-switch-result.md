# GPT 모델 변경 400 수정 및 로컬 설치 결과

## Claim

Sol 답변 이후 GPT-6 Astra로 전환할 때 발생한 모델 계열별 receipt 검증 실패를 합성 대화에서 재현하고 수정했다. 기존 receipt와 암호문은 원래 귀속 그대로 검증하며, 실제 구독 서버에서 GPT-5.6 세 모델과 Astra의 여섯 전환 방향을 확인했다. 마지막 성공 답변 이후 API 오류와 로컬 명령이 남은 대화도 재개할 수 있도록 수정했다. 다른 계정·세션·대화, 변조·누락된 reasoning, 미완료 도구 출력은 계속 거절한다.

독립 검토는 지정 수정 범위 PASS다. 최종 rc.8 실행 파일에서 실제 Claude Code 화면의 Sol 답변 → /model → Astra 답변을 확인하고, 같은 해시의 파일을 /Users/goos/go/bin/moai에 설치했다.

## Evidence

```text
$ python3 .moai/reports/t649/e2e/model-switch-interactive.py
{
  "success": true,
  "visible_sol_answer": true,
  "visible_astra_answer": true,
  "selected_model_ids": [
    "gpt-6-astra"
  ],
  "error_categories": [],
  "timeout": false,
  "cleanup_errors": [],
  "native_matches": [
    {
      "model": "gpt-5.6-sol",
      "marker": "T649_SWITCH_SOL_OK",
      "trailing_period": true
    },
    {
      "model": "gpt-6-astra",
      "marker": "T649_SWITCH_ASTRA_OK",
      "trailing_period": true
    }
  ],
  "binary_sha256": "25b68beaadc919b91adab97213b76daa694b284fe3e297958e627c3842bc08c2"
}
exit 0

$ /Users/goos/go/bin/moai version
[v3.2.0-rc.8] [v3.2.0-rc.8-t649-model-switch-d40ebe382fb9] [built 2026-09-12T02:00:50Z]
$ /Users/goos/go/bin/moai gpt status
GPT: logged in
```

원본 사용자 대화의 읽기 전용 검사:
```text
$ go test -count=1 -overlay=/tmp/t649-native-resume-readonly-overlay.json -run '^TestReadOnlyActualFailedTranscript$' -v ./internal/gateway/conversation
=== RUN   TestReadOnlyActualFailedTranscript
    actual_readonly_test.go:6: model=gpt-5.6-sol error=<nil>
--- PASS: TestReadOnlyActualFailedTranscript (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/gateway/conversation 0.452s
```

패키지·race·변조 거부 시험의 명령과 원문 출력은 [독립 검토](model-switch-review.md), 서버 왕복의 실행별 결과는 [원인 및 수정 증거](model-switch-evidence/diagnosis-and-fix.md)에 기록했다. 화면 시험의 두 수집 실패도 attempt1/attempt2로 보존했다.

## Baseline-attribution

- 워크트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- 브랜치: `WT-unified-gateway`, HEAD: `81c1d58f9` + 기존 미커밋 구현
- Claude Code: `2.1.269`
- 설치 바이너리 SHA256: `25b68beaadc919b91adab97213b76daa694b284fe3e297958e627c3842bc08c2`
- 이전 rc.8 백업: `/Users/goos/.moai/releases/moai-v3.2.0-rc.8-before-model-switch-7767eed67b43`
- 설치 메타데이터: [model-switch-build.json](model-switch-build.json)

## Gaps

원본 대화 파일은 읽기 전용으로만 검사했다. 현재 실행 중인 사용자 세션에 새 바이너리를 주입하거나 동일 대화를 동시에 실행하지 않았다. 실제 사용자의 긴 대화와 hooks/MCP 활성 조건에서 수정 후 서버 왕복은 아직 관측하지 않았다. 화면 E2E는 합성 프로젝트에서 hooks/MCP를 끈 조건이다. 1M 컨텍스트 연결과 전체 SPEC/Windows GitHub CI는 이번 수정으로 완료되지 않는다.

## Residual-risk

암호문 호환성은 현재 구독 endpoint와 계정의 관측에 근거하며 서버 정책이 바뀌면 재검증이 필요하다. 화면 시험은 답변과 native 기록 확인 후 제한된 프로세스 그룹 정리로 종료했고, 자식 종료 -9와 시험 명령 exit 0을 구분했다. 원래 실행 중이던 gateway에는 이전 바이너리 코드가 남으므로 종료 후 재개해야 수정이 적용된다.

기존 대화 재개 명령(원래 프로젝트 디렉터리에서):
```bash
moai gpt --model gpt-6-astra --resume eed72b11-4441-4467-b80c-1e90eece1e3b
```
