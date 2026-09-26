# t1263 판정서 — multi-review-gate 차단 사유 문구

- 카드: t1263 (Class B, Tier S, SPEC 없음) · 브랜치 `WT-review-gate-wording` · 기준 로컬 develop `6b1e9bbd4`
- 근거: t1213 sync-audit F1 (`internal/cli/mcp_convergence.go` `describeDisagreement`)
- PR 교차확인: no-link (리드 지정)
- sync-audit: PASS-WITH-DEBT 91.8 (조화 91.4) — `.moai/reports/t1263/sync-audit.md`

## 원인 (재현으로 확인)
`describeDisagreement` 는 pass/fail 분할이면 필수 FAIL 포함 여부와 무관하게 `cross-model disagreement (advisory, NOT a block)` 를 냈다. 필수 FAIL 이 섞이면 `converge()` 가 overall=fail 로 판정하고 `multi_review_gate.go` 가 그 값으로 차단하므로, 문구가 동작과 반대였다. 게이트는 `OverallVerdict` 만 보고 결정하며 노트는 사유로 되울릴 뿐이다.

재현 테스트 `internal/cli/convergence_note_wording_test.go` 를 수정 전 코드에서 실행:
```
residual_risk_note = "cross-model disagreement (advisory, NOT a block): pass=[codex(required)] fail=[claude(required)]", want it to contain "required-backend FAIL: claude"
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.791s
```

## 수리
분할에 필수 FAIL 이 있으면 `required-backend FAIL: <이름>; cross-model disagreement: pass=[…] fail=[…]`, 없으면(권고만 엇갈림) 기존 advisory 문구 유지. 스킬 `moai-ref-cross-model-audit` 출력 예시(템플릿·로컬, 바이트 동일)와 `catalog.yaml` 해시 갱신. sync-audit F3(필수 FAIL + 권고 PASS 케이스 누락)·F4(`describeRequiredFails` 주석) 반영.

## 증거 (이번 실행, 워크트리 t1263)
- `go test -count=1 -run TestConvergenceNoteMatchesBlockDecision -v ./internal/cli/` → `--- PASS` (3 케이스), `ok`
- `go test -count=1 -run 'TestConvergence|Convergence|Disagree|RequiredGate|MultiReview|AuditMulti|Participant|Divergence' ./internal/cli/` → `ok … 3.791s`
- `go test -count=1 ./internal/template/` → 해시 재생성 후 `ok … 66.052s` (재생성 전 catalog hash 3건 FAIL)
- `golangci-lint run ./internal/cli/...` → `0 issues.` · `GOOS=windows go build ./...` → exit 0 · gofmt 0건

## 미검증
- `internal/cli` 전체 스위트(CI 몫)

## 잔여 위험 / 후속 (기존 결함, 이 카드 범위 밖)
- sync-audit F1: 분할 없는 필수 FAIL 에 synthesis note 가 붙으면 노트가 실패 백엔드 이름 없이 `cross-model disagreement detected…` 로 나온다.
- sync-audit F2: `enforceRequiredGateUnmet` 가 overall 을 fail 로 뒤집어도 뒤에 `(advisory, NOT a block)` 이 남는다.
