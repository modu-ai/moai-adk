auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: f99b0ebe0bf107733923adf68a7bbac534a90926

## Evaluation Report
SPEC: SPEC-FACTORY-DECISION-AUTO-001 (card t1481)
Overall Verdict: FAIL — 점수 72/100 (필수 게이트 codex FAIL, Functionality 필수 차원 미달)

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 60 | FAIL | 선언된 테스트는 전부 초록(아래 Evidence). 그러나 승인 게이트 `auditverdict.Admit`가 비정상 점수를 통과시킴(F1) — verdict.go:159 `case f.Score < threshold:`는 NaN에 대해 거짓이고 상한 검사가 없다 |
| Security (25%) | 75 | PASS(경계) | 비밀·주입 표면 없음. 다만 F1·F2는 fail-closed 게이트의 우회 경로 |
| Craft (20%) | 80 | PASS | lint 0 issues(progress 기재, 본 감사 미재측정), decision 커버리지 85.6%(기재) |
| Consistency (15%) | 85 | PASS | 템플릿 중립성 스캔 0건; F2는 함수 주석과 코드가 서로 어긋남 |

### Findings
- F1 [P1] [blocking] internal/auditverdict/verdict.go:157-160 — `NaN`·`+Inf`·`2.0` 점수가 plan 승인 게이트를 통과. codex overlay 재현에서 NaN 판정으로 kickoff→run 전이 성공(`state=run err=<nil>`); 본 감사는 코드 판독으로 비교 의미론(NaN < x == false, 상한 부재)을 확인. Required fix: `math.IsNaN/IsInf` 거부 + `0 <= score <= 1` 범위 제한, 회귀 테스트 추가.
- F2 [P1] [blocking] internal/homestate/card_audit_kickoff.go:88 — `label != "FOUNDER"`에서 먼저 continue 하므로 `Label: DECIDED` + `Class: product-level` + `DEFAULT-APPLIED` 행이 제한을 우회. 같은 함수 주석(:62 "any DEFAULT-APPLIED verdict outside an implementation-level row")과 불일치. Required fix: DEFAULT-APPLIED 검사를 라벨 분기 앞으로 이동, 픽스처 추가.
- F3 [P2] [blocking 후보 — 리더 판단] card_audit_kickoff.go:57 — §E.1 본문이 비어있지 않으면 audit-ready로 인정(`audit_ready: false`도 통과, codex 재현). Required fix: 명시 상태 파싱, 부정/미확인 거부.
- F4 [P2] [optional] internal/cli/decision.go:135 — `decision read` 출력에 `resolves`·`predicate`·`release`·`cards` 누락(codex 재현). watchdog의 대기 해제 판단 근거 부족. Required fix: 전 필드 출력 또는 `--json`.

### 확인된 사항 (지시 항목)
- PARTIAL 부채(AC-008, 011~013, 019)와 M0 미실행은 progress.md 매트릭스·§E.3 `ac_partial`/`gaps`·Recorded debts에 정직하게 기록됨 — PASS로 주장되지 않음.
- 통합 위험: progress.md Residual-risk에 "this card lands LAST" + 필드 부재 판정 Kickoff/T7 거부 명시. CHANGELOG에도 "이 필드가 없는 기존 plan 판정은 Kickoff/T7에서 거부된다 — 진행 중 카드는 재감사 필요" 명시. 착지 순서는 progress에만 있고 CHANGELOG엔 없음(정상 범위).
- 템플릿 중립성: agents/skills 템플릿 추가 줄에 SPEC/REQ/카드/날짜/절대경로/CLAUDE.local 토큰 0건.

## Claim / Evidence
- `go test -count=1 ./internal/decision/... ./internal/auditverdict/... ./internal/template/agentemit/...` → exit=0
  `ok internal/decision 0.738s` / `ok internal/auditverdict 0.260s` / `ok internal/template/agentemit 0.461s`
- `go test -count=1 -run 'TestFDA_' ./internal/homestate/ ./internal/hook/ ./internal/cli/` → exit=0
  `ok internal/homestate 19.635s` / `ok internal/hook 12.910s` / `ok internal/cli 2.691s`
- `git diff develop...WT-decision-automation -- internal/template/templates/.claude/{agents,skills}/ | grep '^+' | grep -E 'SPEC-[A-Z]|REQ-[A-Z]|t1[0-9]{3}|date|/Users/|CLAUDE\.local'` → 0 hits
- `git rev-parse WT-decision-automation` → f99b0ebe0bf107733923adf68a7bbac534a90926
- audit_multi(baseBranch, card_id t1481): claude(required) pass, codex(required) **fail** (F1~F4), glm(advisory) inconclusive ("z.ai response carried no text content"). overall_verdict=fail, disagreement_flag=true. audit_receipt 미발급.

## Baseline-attribution
트리 /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a6b97f013ea3df7de, HEAD f99b0ebe0, 이번 실행에서 측정. codex 비교 기준 d7112d005.

## Gaps
- 워크트리 가드가 대상 트리에 대한 `git -C`/`cd && git`/변수 경로 `sed`를 거부 — 브랜치 diff는 develop 트리에서 브랜치명으로, 코드는 Read로 대체.
- F1~F4의 실행 재현은 codex overlay 출력에 의존; 본 감사는 F1·F2를 코드 판독으로만 확인(직접 실행 재현 없음).
- `-race`, 커버리지, lint, `internal/template` 전체, homestate 전체(AC-015) 미재측정.
- moai MCP 서버 바이너리 45600e4ee가 HEAD의 조상(build lag) — CLI 기반 출력은 신뢰하지 않음; 테스트는 소스에서 직접 실행.
- glm 백엔드 무응답.

## Residual-risk
- F1·F2 수리 후에도 착지 순서(이 카드 마지막) 미준수 시 진행 중 카드 정지.
- `decision-index.md` 해시 입력 확대로 1회 캐시 미스.

## Iteration history
- iter 1 (본 감사): FAIL 72.
