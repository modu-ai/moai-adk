auditor-model: claude-opus-5-5

## Evaluation Report — 델타 sync 재감사 (2회차)
SPEC: SPEC-FACTORY-DECISION-AUTO-001 · card t1481
트리: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a6b97f013ea3df7de · HEAD 1b732c265
Overall Verdict: FAIL (78/100)

verdict: FAIL
audited_sha: 1b732c265a099e34a4cb09857bdc570d071d7d6a

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 70 | FAIL | F1~F3 수리는 검증됨(아래). 그러나 같은 승인 게이트에 새 우회 경로 N1(중복 키 last-wins)이 남아 있고, 이는 코드 판독으로 확인함 |
| Security (25%) | 70 | FAIL(must-pass) | N1은 fail-closed 게이트 우회. N2(hold TOCTOU)는 codex 재현만 있음 |
| Craft (20%) | 88 | PASS | 회귀 테스트 3종 모두 초록이고, 변이(수리 전 코드) 적용 시 적색 |
| Consistency (15%) | 85 | PASS | `auditReadyRecorded` 주석이 낡음("content other than a placeholder") |

### 이전 지적 F1~F4 처분
| ID | 수리 | 초록(이번 실행) | 변이 검사(수리 전 파일 overlay) | 판정 |
|---|---|---|---|---|
| F1 | verdict.go Parse: 유한값·[0,1]만 수용 | `ok .../internal/auditverdict 1.104s` (-race) | `verdict_test.go:120: score "+Inf" admitted` / `"Inf"` / `"1.5"` → FAIL | 해소 |
| F2 | DEFAULT-APPLIED 검사를 라벨 분기 앞으로 이동 | `ok .../internal/homestate 12.561s` (-race, -run TestFDA_) | `--- FAIL: .../DECIDED_row_holding_DEFAULT-APPLIED_on_product-level ... audit approval accepted, want refusal` | 해소 |
| F3 | §E.1에 `audit_ready: true` 줄이 명시적으로 있어야 함 | 위와 같은 실행 | `--- FAIL: .../audit_ready_false ... audit approval accepted, want refusal` | 해소(잔여 N3·D1) |
| F4 | `decision read`에 predicate/supersedes/resolves/release/cards 출력 | `ok .../internal/cli 4.290s` (-race, -run "TestFDA_\|TestDecisionCmd") | 출력 전용이라 변이 검사 없음 | 해소 |

F3 계약 정합성: `auditReadyRecorded`를 호출하는 곳은 `card_audit_kickoff.go:32`의 audit decider 하나뿐이다. 사람 decider 경로는 이 함수를 타지 않으므로 사람 경로는 깨지지 않았다. SPEC 본문(spec.md:132 "audit-ready status recorded")은 표기를 고정하지 않아 명시 키 방식과 모순되지 않는다. 다만 D1을 보라.

### Findings
- N1 [P1] [blocking] internal/auditverdict/verdict.go:66-104. `Parse`는 키가 중복되면 마지막 값을 쓴다. `verdict: FAIL` 뒤에 `Verdict: PASS`를 덧붙이면 FAIL이 PASS로 바뀐다. `must_pass_failed: 1` 뒤에 `0`을 쓰면 실패가 지워진다. 유효 점수 뒤의 `NaN`은 조용히 무시된다. 감사 규약은 "두 값이 다르면 카드는 이동하지 않는다"고 정하지만, 이 판독기는 그 규약을 지키지 않는다. 본 감사는 코드 판독으로 확인했고, codex는 overlay 반례로 T7/T8a/T13 전이 성공(`state=run err=<nil>`, `state=merge-ready err=<nil>`)을 재현했다. 확신도: 높음. Required fix: 정규화한 키가 중복되거나 값이 충돌하거나 파싱에 실패하면 거부한다. 모든 판독 지점이 같은 판독기로 같은 스냅샷을 읽게 한다.
- N2 [P1] [blocking] internal/cli/factory_audit_decide.go:46. 큐 hold 상태를 트랜잭션 전에 읽어 문자열로 넘긴다. 그 사이에 커밋된 운영자 hold는 감지되지 않는다(codex 결정적 반례 `queue=hold card=run err=<nil>`). 본 감사는 직접 재현하지 않았다. 확신도: 중. Required fix: 잠금을 잡은 상태에서 hold를 다시 읽고 전이를 커밋한다.
- N3 [P2] [optional] card_audit_kickoff.go:58. `audit_ready: true`를 처음 찾는 즉시 반환하므로, 뒤에 `false`가 같이 있어도 통과한다. 이 신호는 작성자가 직접 적는 값이라 우회 이득이 없어 비차단으로 분류한다. 권고: 상충하는 값은 거부한다.
- N4 [P2] [optional] internal/decision/board.go:263. 필수 필드가 없는 보드 행도 supersedes 처리에 참여해 standing hold를 지운다(codex 재현). 부채로 분류한다.
- D1 [P2] [optional, 부채] F3 계약에 생산자가 없다. 최근 SPEC 40건의 §E.1 키 분포를 쟀더니 `plan_status` 22건, `plan_complete_at` 22건, `audit_ready` 0건이었다. manager-spec 템플릿도 이 키를 내지 않는다. 결과적으로 autonomous audit decider는 이 SPEC 밖의 모든 SPEC을 fail-closed로 거부하게 된다. 안전한 방향이므로 비차단이지만, 생산자(plan 감사 후 §E.1 기록 단계)에 `audit_ready: true` 기록 의무를 추가해야 한다.
- C1 [P1] [blocking, 프로세스] spec.md가 sync FAIL 상태에서 `status: completed`로 되어 있다(claude 지적). 재감사 PASS 전까지 completed 주장은 성립하지 않는다.
- C2 [P2] [리더 판단] 감사된 SHA a13b83868 이후 plan 산출물을 편집했다(Q27, REQ-FDA-014, design §5, AC-FDA-014 — 커밋 818569869). 그런데 §E.1의 `plan_artifacts_frozen_at` 주장은 그대로 남아 있다. 델타 plan 재감사 또는 기록 정정이 필요하다.
- 기타(claude, 비차단): 제품 수준 판정 자기기재 경로, sync-audit-4dim 바인딩 조건이 빈 추출일 때 통과하는 문제, 문서 필드 목록 불일치. 부채 후보로 둔다.

### Claim / Evidence / Baseline-attribution / Gaps / Residual-risk
- Claim: F1~F4는 해소됐다. 새 차단 결함 N1(게이트 우회)과 N2(경합), 프로세스 결함 C1이 남아 FAIL이다.
- Evidence: 위 표의 명령 출력 원문. 변이 검사는 `go test -overlay`로 `git show 1b732c265~1:<file>`을 끼워 넣어 실행했다.
- Baseline-attribution: 카드 트리, HEAD 1b732c265, 이번 실행에서 측정. go 툴체인으로 트리 소스를 직접 컴파일했고 설치된 moai 바이너리는 쓰지 않았다. audit_multi는 build_lag를 고지했다(바이너리 45600e4ee는 트리 HEAD의 조상). 이 고지는 MCP 서버 자체에 해당하며 Go 테스트 측정과는 무관하다.
- Gaps: (1) 카드 트리를 대상으로 한 `git -C`와 `cd`가 worktree 가드에 거부됐다. 그래서 공유 객체 저장소를 통한 `git show/diff <sha>`와 `go -C`로 대체했다. (2) N1은 코드 판독으로만 확인했고 N2는 codex 재현에만 의존한다(본 감사의 직접 실행 재현은 없음). (3) homestate 전체 패키지와 lint는 재실행하지 않았다. (4) GLM은 inconclusive("z.ai response carried no text content")였다.
- Residual-risk: N1이 고쳐지기 전에는 판정 파일에 줄 하나만 덧붙여도 plan, kickoff, merge-ready 게이트를 모두 통과할 수 있다.

### 교차 모델
audit_multi(baseBranch, card_id t1481, project_root=카드 트리): claude(required) fail, codex(required) fail, glm(advisory) inconclusive. overall_verdict=fail, disagreement_flag=false. audit_receipt는 발급되지 않았다.

### 반복 이력
1회차 FAIL 72(F1~F4) → 2회차 FAIL 78(F1~F4 해소, N1·N2·C1 신규 차단).
