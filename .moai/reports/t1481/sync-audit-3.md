auditor-model: claude-opus-5-5

## Evaluation Report (sync 재감사 3회차, 수렴 규칙 적용)
SPEC: SPEC-FACTORY-DECISION-AUTO-001 (card t1481)
Overall Verdict: FAIL — 점수 80

verdict: FAIL
audited_sha: 7b5a85141

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 70 | FAIL | N1 수리가 정상 plan-auditor 보고서까지 거부한다(B1, 아래 실측). 결과적으로 T8a 자동 승인 AC를 정상 산출물로는 충족할 수 없다 |
| Security (25%) | 85 | PASS | N1 last-wins 우회는 막혔다. N2는 범위가 좁아졌고 잔여 창은 부채로 둔다. codex가 찾은 FOUNDER 행 last-wins 문제는 해시 결속 때문에 부채로 분류했다 |
| Craft (20%) | 85 | PASS | 대상 테스트 3건 모두 green(-race). 회귀 테스트도 추가됐다 |
| Consistency (15%) | 80 | PASS | 파서 정규화가 기존 출력 양식과 맞지 않는다(B1과 같은 원인) |

### 수리 검증 (N1/N2/N3/C1)
| 항목 | 판정 | 근거 |
|---|---|---|
| N1 중복 키 last-wins | 우회는 해소. 대신 과잉 거부 회귀가 생김(B1) | verdict.go:88-103 중복 수집, Admit:170 거부. 변이 추론: Admit의 거부 분기를 지우면 `TestAdmit_RefusesDuplicatedDecisionKeys`의 "verdict PASS appended after FAIL" 케이스가 실패한다 |
| N2 hold 경합 | 범위 축소(수용된 잔여) | card_transition.go:494-497. hold 읽기가 tx 안의 커밋 직전으로 옮겨졌다. 변이 추론: QueueHoldRead를 무시하면 신규 테스트가 reads==0으로 실패한다. 호출자는 하나뿐이다(factory_audit_decide.go:48, grep으로 확인) |
| N3 audit_ready 상충 값 | 해소(같은 섹션 안 기준) | card_audit_kickoff.go:57-69, 신규 테스트의 상충 케이스가 거부된다 |
| C1 completed 주장 | 해소 | spec.md `status: in-progress` (diff 1b732c265..7b5a85141) |

### Findings
- B1 [P1] [blocking — AC 구현 불가] internal/auditverdict/verdict.go:86-103. 키를 소문자로 접은 뒤 중복을 판정한다. 그런데 plan-auditor.md가 정한 보고서 양식(648-649행 `Verdict:`/`Overall Score:` 헤더 + 203·212행 `verdict:`/`overall_score:` 기계 줄)에는 같은 키가 두 번, 같은 값으로 들어간다. 그래서 정상 보고서도 모두 "duplicated decision key(s)"로 거부된다. 실측 결과 이 카드의 plan-audit-iter4.md(PASS)와 템플릿 모양 그대로의 최소 보고서가 둘 다 거부됐다. 따라서 T8a 자동 승인은 규약을 지킨 auditor 산출물로는 영영 성립하지 않는다. 확신도: 높음(실행 관측). Required fix: 값이 서로 다른 중복과 파싱에 실패한 중복만 거부하고, 같은 값의 반복은 허용한다. 실제 보고서 양식(헤더와 기계 줄이 같은 값)을 쓰는 회귀 테스트를 추가한다. 참고: Claude 백엔드도 같은 위험을 P2로 제기했다(코드 미열람 추정).
- D2 [P1] [debt — 해시 결속] card_audit_kickoff.go:84-113. FOUNDER 행의 `Label`·`Class`·`Operator verdict`도 last-wins로 읽는다. 예를 들어 `Class: product-level` 뒤에 `Class: implementation-level`을 쓰면 product-level에 적용된 DEFAULT-APPLIED가 통과한다. 이 문제는 codex가 overlay 재현으로 확인했고 본 감사가 코드 판독으로 같은 결론을 얻었다. decision-index.md는 plan 해시에 묶여 있으므로 감사 후에 덧붙이면 해시 불일치로 거부된다. 즉 감사 전에 저작된 파일을 plan-auditor가 놓쳐야만 성립하므로 부채로 둔다. 다만 B1 수리와 같은 방식(중복·상충 필드 거부)으로 함께 고칠 것을 강하게 권고한다.
- D3 [P1] [debt — 수용된 N2 잔여] backlog.db와 factory.db는 서로 다른 저장소라 tx 안에서 읽어도 hold 쓰기를 잠그지 못한다. codex는 주입한 overlay로 `queue=hold card=run`을 재현했다. 실제 창은 같은 프로세스 안의 읽기와 SQLite 커밋 사이의 마이크로초 구간이다. 이 창을 악용해 얻는 사람도 없다(hold를 거는 주체가 운영자 본인이다). 그 순간 직후에 걸린 hold와 결과가 같으므로 실무상 악용할 수 없다고 판단한다. 구조적 수리는 t1458 원자 임대 몫이다.
- D4 [P2] [debt] verdict.go:146 tier를 frontmatter 밖 본문 예시에서도 읽는다(codex 재현: 0.76 점수가 S 기준 0.75로 통과). spec.md가 해시에 묶여 있어 감사 전에 저작돼야 성립한다.
- D5 [P2] [optional] card_audit_kickoff.go:45 첫 번째 `## §E.1`만 검사한다. 저자 자신의 신호이므로 우회로 얻는 이득이 없다(N3와 같은 분류).
- D6 [P2] [debt] progress.md §E.1 `plan_artifacts_frozen_at: a13b83868` 주장은 818569869(Q27 편집)와 모순된다(Claude 지적, 2회차 C2와 같음). §E.4에도 `in-progress -> completed` 전이 줄이 남아 있다.
- D1 [debt, 기존 수용] audit_ready를 내보내는 writer가 없어 자동 승인이 작동하지 않는다(fail-closed).

### Claim / Evidence / Baseline-attribution / Gaps / Residual-risk
- Claim: N1·N2(축소)·N3·C1 수리는 확인됐다. 다만 N1 수리가 새 차단 결함 B1을 만들었다.
- Evidence:
  - `go -C <T> test -count=1 -race ./internal/auditverdict/...` → `ok github.com/modu-ai/moai-adk/internal/auditverdict 1.104s`
  - `go -C <T> test -count=1 -race -run TestFDA_ ./internal/homestate/...` → `ok .../internal/homestate 17.383s`
  - `go -C <T> test -count=1 -race -run "TestFDA_|TestDecisionCmd" ./internal/cli/` → `ok .../internal/cli 7.436s`
  - B1 탐침(overlay probe_test.go, 트리 미변경): `dups=[verdict] label="PASS" ok=false reason="duplicated decision key(s): verdict"`(plan-audit-iter4.md) / `template-shaped: dups=[verdict overall_score] ok=false reason="duplicated decision key(s): verdict, overall_score"`
  - audit_multi: codex(required) fail(P1×2, P2×2), claude(required) inconclusive(diff 잘림), glm inconclusive. overall fail, gate_unmet claude, receipt id 미발급.
- Baseline-attribution: 트리 /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a6b97f013ea3df7de, HEAD 7b5a85141(`git log` 공유 객체 저장소로 확인). 테스트는 go 툴체인으로 그 트리 소스를 직접 컴파일했다. moai 바이너리 판정은 쓰지 않았다(audit_multi build_lag: 45600e4ee는 HEAD의 조상).
- Gaps: (1) 카드 트리를 대상으로 한 `git -C`와 `cd`, 복합 명령이 worktree 가드에 거부됐다. 그래서 자기 트리의 plain git(공유 객체 저장소)과 `go -C`로 대체했다. (2) codex의 overlay 재현(D2~D5)은 직접 재실행하지 않았다. D2만 코드 판독으로 교차 확인했다. (3) lint와 패키지 전체 테스트는 재실행하지 않았다. (4) Claude 백엔드는 diff가 잘려 inconclusive였고 GLM은 응답이 없었다.
- Residual-risk: B1이 그대로면 T8a 자동 승인은 정상 auditor 보고서로 영영 성립하지 않는다. 지금은 fail-closed 방향이라 안전하지만 기능은 사실상 죽어 있다. D2는 auditor가 놓칠 경우에만 성립하는 우회다.

### Iteration history
1회차 FAIL 72(F1~F4) → 2회차 FAIL 78(N1·N2·C1) → 3회차 FAIL 80(N1~N3·C1 해소/축소, N1 수리가 낳은 B1 단일 차단). 다음 재감사는 B1(같은 김에 D2) 델타로 범위를 좁히면 된다.
