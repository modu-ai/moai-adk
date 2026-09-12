# t649 수리 후 실제 E2E 검증

## Claim

실제 GPT-6 Astra 답변, Read 도구 왕복, 도구 대화의 새 프로세스 재개와 세 제공자 선택기가 통과했다. rc.8 후보에서도 대화형 `/model`로 Astra를 선택한 뒤 답변 표시와 native assistant 기록을 함께 확인했다.

| 여정 | 상태 | 실측 초 | 증거 수 | 증거 |
|---|---|---:|---:|---|
| GPT6 첫 답변 | PASS | 5.73 | 1 | [JSON](answer-repaired.json) |
| Read 도구 왕복 | PASS | 9.07 | 1 | [JSON](tool-repaired-corrected.json) |
| 도구 대화 새 프로세스 재개 | PASS | 2.79 | 1 | [JSON](resume-tool-repaired.json) |
| GPT 선택기 | PASS | 21.65 | 1 | [JSON](picker-gpt-repaired.json) |
| Claude 선택기 | PASS | 21.85 | 1 | [JSON](picker-cc-repaired.json) |
| GLM 선택기 | PASS | 21.84 | 1 | [JSON](picker-glm-repaired.json) |
| rc.8 대화형 선택·답변 | PASS | 26.07 | 1 | [JSON](interactive-rc8.json) |

## Evidence

각 JSON은 실제 실행 출력에서 추출한 판정이다. 원시 터미널·API 본문·인증 값은 저장하지 않았다. print 여정은 exit 0, result 이벤트의 합성 표식, 오류 없음이 모두 필요하다. 도구 여정은 Read 이벤트도 필요하다.

최종 대화형 여정은 화면의 `⏺T649_INTERACTIVE_OK`와 별도로 테스트 전용 native JSONL에서 assistant 역할·gpt-6-astra 모델·정확한 답변을 확인했다. 프롬프트의 재표시를 답변으로 판정하지 않았다.

```json
{"success":true,"visible_assistant_marker":true,"native_assistant_match":true,"native_models":["gpt-6-astra"],"family_id":"2707a84c-37c0-4ded-9adf-59e726948e9f"}
```

재현 명령은 [gateway_probe.py](gateway_probe.py)와 [interactive_probe.py](interactive_probe.py)에 있다. 최종 대화형 실행 명령은 `python3 interactive_probe.py`이다.

## Baseline-attribution

- 첫 여섯 여정: `/tmp/moai-t649-repaired`, SHA256 `5b259dbbaee35c325d3ff8dd959d7192ff9a6ba91edb2dabab5a18bbc8a792d7`.
- 최종 대화형: `/tmp/moai-v3.2.0-rc.8-t649`, SHA256 `7767eed67b43981891f2cd3cf0954d76daf480a5109c8cba14f6efde176d4092`.
- 실제 Claude Code를 Python PTY에서 실행했다. 테스트 cwd는 `/tmp/moai-t649-e2e-project`이며 permission mode default, hooks 비활성화, MCP 빈 목록이다.
- 대화형 선택기는 테스트 전용 새 family에만 onboarding·trust를 준비하는 [claude_fixture.py](claude_fixture.py)를 사용했다. 사용자 전역 설정을 변경하지 않았다.

## Gaps

- macOS 한 환경·한 계정의 합성 대화이다. Windows/GitHub CI, 전체 MCP, 실제 프로젝트, Claude·GLM 유료 추론은 이 실행에서 검사하지 않았다.
- print 여정의 자연 종료 뒤 프로세스 그룹 신호에서 `signal_permission_denied`가 기록되었다. exit 0과 결과는 관찰했으나 모든 하위 프로세스 정리를 보장하지 않는다. 선택기와 최종 대화형 여정은 제한된 실행 후 종료했고 cleanup 오류가 없다.
- 설치는 이 에이전트가 수행하지 않았다. 후보 바이너리 경로를 검증한 결과이다.

## Residual-risk

- 첫 도구 실패는 `--allowedTools`가 프롬프트를 소비한 하네스 오류였다. 정확한 오류와 `--` 수정은 [진단](tool-harness-diagnosis.json), 원래 실패는 [기록](tool-repaired.json)에 보존했다.
- 첫 대화형 실행은 답변 표시 직후 종료하여 native flush를 확인하지 못했다. [실패 기록](interactive-repaired.json)을 보존했고, 6초 대기를 추가한 별도 rc.8 후보 실행에서 화면과 native 교차 검증이 통과했다.
- 선택기 성공은 허용 모델 집합과 세션 한정 변경의 UI 증거이다. 서버의 다른 제공자 요청 거절은 별도 단위·통합 검증 소관이다.
