auditor-model: claude-opus-5-5

## Evaluation Report (4회차 · 최종 델타 — 범위 B1·D2)
SPEC: SPEC-FACTORY-DECISION-AUTO-001 (card t1481)
Overall Verdict: PASS-WITH-DEBT — 점수 88

verdict: PASS-WITH-DEBT
audited_sha: b8cd707d0e57dfb439fdf76d12a3a828d6b3a7e2

### Dimension Scores (델타 범위)
| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 90 | PASS | B1 해소: 같은 값 반복은 허용하고, 정규화 후 상충하거나 파싱되지 않는 반복은 거부한다(verdict.go:97-105). 실제 보고서 fixture가 승인된다. D2 해소: 같은 행에서 Label/Class/Operator verdict가 상충하면 거부하고 같은 값 반복은 허용한다(card_audit_kickoff.go:86-111) |
| Security (25%) | 88 | PASS | 남은 두 지적은 모두 fail-closed 방향이다. FAIL/보류 상태가 게이트를 통과하는 경로는 재현되지 않았다 |
| Craft (20%) | 85 | PASS | 회귀 테스트 추가(TestParse_RealReportShapeEqualDuplicatesAdmitted, conflicting/equal Class 하위 케이스). 순서 의존 정규화 비대칭이 남아 있다(F1) |
| Consistency (15%) | 88 | PASS | 정규화 규칙이 plan-auditor 보고서 양식(헤더와 기계 줄)과 맞는다 |

### Findings
- F1 [P2] [optional] internal/auditverdict/verdict.go:100,109 — 중복 비교는 대문자로 정규화하지만, f.Label에는 마지막 원문이 남는다. 그래서 `verdict: PASS` 뒤에 `Verdict: pass`가 오면 거부되고, 역순이면 승인된다(codex overlay 재현). 거부 쪽은 fail-closed이고 승인 쪽은 두 줄 모두 PASS이므로 정당한 승인이다. 차단 사유가 아니다. 권장 수리: 첫 값(또는 정규화값)으로 Label을 고정하고, 양쪽 순서를 검사하는 테스트를 추가한다.
- F2 [P2] [optional/debt] internal/homestate/card_audit_kickoff.go:84-111 — 첫 `### ` 행보다 앞의 머리말(코드 블록 예시 포함)도 하나의 블록으로 파싱한다. 그래서 예시 속 `Label:` 두 줄이 서로 다르면 정상 Kickoff가 거부된다(codex 재현). 방향은 fail-closed이다. 이 SPEC의 실제 decision-index.md는 첫 Q 행 앞에 Label/Class/Operator verdict 줄이 0개다(awk 실측, 출력 없음). 따라서 AC는 구현 가능하고 부채로 남긴다. 권장 수리: `### Q<N>:` 블록만 파싱하고, 머리말과 코드 펜스는 제외한다.
- 수용된 부채 D3~D6은 그대로 유지한다(D6: 델타 plan 재감사 미실시).

### Evidence (Claim / Evidence / Baseline / Gaps / Residual-risk)
- Claim: B1과 D2가 해소됐다. CRITICAL 결함은 남아 있지 않다.
- Evidence:
  - `go -C <tree> test -count=1 -race -v ./internal/auditverdict/...` → `--- PASS: TestAdmit_RefusesDuplicatedDecisionKeys`, `--- PASS: TestParse_RealReportShapeEqualDuplicatesAdmitted` 외 6건 PASS, `ok github.com/modu-ai/moai-adk/internal/auditverdict 1.114s`
  - `go -C <tree> test -count=1 -race -v -run 'TestFDA_AuditDeciderFounderRows' ./internal/homestate/` → 하위 케이스 8건 PASS(`conflicting_Class_lines`, `equal_duplicate_Class_lines` 포함), `ok .../internal/homestate 6.349s`
  - 테스트 판독 결과: FAIL 뒤에 `Verdict: PASS`를 덧붙인 경우 거부, 서로 다른 must_pass_failed/blocking_count/score/hash 반복 거부, `overall_score: banana` 반복 거부, testdata/plan-audit-report.md(5행 `Verdict: PASS`, 6행 `Overall Score: 0.885`, 10행 `verdict: PASS`)는 DuplicateKeys 0건에 Label PASS.
  - codex_audit adversarial(baseBranch, project_root=card tree) → verdict fail, P2 2건(F1·F2). 둘 다 fail-closed 과잉거부이거나 정당한 승인이다. 분류는 optional/debt이다.
- Baseline-attribution: 카드 트리 HEAD b8cd707d0(worktree gitdir HEAD → WT-decision-automation, `git rev-parse b8cd707d0` = b8cd707d0e57…). 테스트는 이번 실행에서 이 트리의 소스로 컴파일했다.
- Gaps: (1) 카드 트리를 대상으로 한 `git -C`는 worktree 가드에 거부됐다. 그래서 파일 직접 판독과 `go -C`로 대체했고, 이번 회차 diff는 git으로 열람하지 않았다. (2) codex 재현은 직접 재실행하지 않았다. F2 전제는 awk로 실제 decision-index를 판독해 교차 확인했다. (3) codex 결과에는 설치 moai 바이너리 45600e4ee가 HEAD의 조상이라는 build_lag 고지가 있다. 다만 codex 리뷰 자체는 트리 소스를 읽었다. (4) lint와 패키지 전체 테스트는 범위 밖이라 실행하지 않았다. (5) Claude/GLM 백엔드는 실행하지 않았다(codex 필수만 수행).
- Residual-risk: 템플릿이 decision-index 머리말에 Label 예시를 넣기 시작하면 F2가 Kickoff를 막는다. 다만 우회가 아니라 차단이다.

### Receipt 및 audit_multi 재실행 (Stop 훅 AUDIT_RECEIPT_VIOLATION 이후)
- `codex_audit`과 `audit_multi`(둘 다 project_root=카드 트리, baseBranch)는 모두 `audit_receipt` 필드 없이 반환됐다. 그래서 인용할 receipt id가 없다. id를 지어내지 않았다.
- audit_multi 결과: overall fail, gate_unmet=claude.
  - Claude 백엔드는 diff가 잘려 inconclusive였다. GLM은 응답이 없었다.
  - codex는 F1만 다시 제기했다(P2, kickoff overlay에서 순서 의존 재현).
- Claude 백엔드의 추가 지적은 다음과 같이 분류했다.
  - (a) 머리말 `Plan Artifact Hash:` 산문과 기계 줄 hex가 상충할 수 있다. optional이다. 템플릿 650행은 머리말에도 실제 해시를 요구하므로 규약대로 쓰면 값이 같다. 산문은 해시를 계산하지 않은 경우에만 나오는데, 그 경우는 어차피 hash 결속 실패로 거부돼야 한다.
  - (b) 문서의 "exactly once"가 같은 값 반복 허용 규칙과 어긋난다. optional, 문서 부채이다.
  - (c) 어휘 밖 Label을 쓴 행의 빈 verdict가 통과할 수 있다. 확신도 0.45의 미검증 가설이고, D2와 같은 부류(plan 해시 결속)의 부채이다.
  - (d) D6과 같은 지적이다(기존 부채).
- 셋 다 CRITICAL 조건(FAIL/보류 통과, 데이터 손실, AC 구현 불가)에 해당하지 않는다. 판정은 유지한다. 다만 receipt로 뒷받침할 수 없으므로 런타임 receipt 가드가 이 PASS-WITH-DEBT를 거부할 수 있다. 그 처분은 리더 몫이다.

### Iteration history
1회차 FAIL 72 → 2회차 FAIL 78 → 3회차 FAIL 80(B1) → 4회차 PASS-WITH-DEBT 88(B1·D2 해소, F1·F2는 optional, 수렴 규칙상 최종)
