# 진행 기록 — SPEC-REVIEW-GATE-REPAIR-001

상태: draft (plan 단계). 이 파일의 §E.2 ~ §E.4 는 자리표시다. 해당 단계의 담당자가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-10T16:27:35Z

- 이력 (당시 상태 draft, 재감사 대기): plan-audit 1회차 FAIL(score 0.625, MP-7). 수리 1회차(spec v0.1.1)를 적용했다. plan-audit 2회차는 FAIL(codex, 영수증 rcpt-0efbcef22b780f388f658933, 발견 F-1·F-2)이었고, 수리 2회차(spec v0.1.2)를 적용했다. 다음 plan-audit 이 통과하기 전까지 plan 은 audit-ready 가 아니다. 재감사 통과 전에는 audit-ready 로 표시하지 않는다.
- v0.2.0 적용 (분할): plan-audit 3회차 FAIL(codex, 영수증 rcpt-3e341c03b359a26ff7d01963, 발견 F-3: 활성 모드를 무시하고 manual.workflow 를 읽음, P2)에 대해 리더 판정 d-20261010T070728Z-6cdf(영실이 판단, 운영자 위임)는 F-3 을 결함으로 보아 부채로 받지 않고 REQ-ACE-005 를 분할하라는 것이었다. 기준 브랜치 선택(REQ-RG-005–008 과 그 인수 기준, AC-RG-006 포함)은 후속 카드 t1662 로 이관했고, 범위는 카드 항목 (1)·(3)·(5) 로 좁혔고, (6) 중 REQ-RG-016 은 이 SPEC 에 남는다 (121e, 6d67) (spec.md v0.2.0). 계획은 다음 계획 감사가 통과하기 전까지 audit-ready 가 아니다. [ID mapping v0.2.2: REQ-RG-016→REQ-RG-012. 이 항목의 REQ-RG-005–008 은 v0.2.0 에서 폐기된 식별자다.]
- plan_complete_at: 재감사 통과 시 기록한다. 현재 없음.
- tier: L (spec.md · plan.md · acceptance.md · design.md · research.md, 그리고 progress.md 와 decision-index.md)
- 운영자 판정: Q1–Q6 과 보드 판정 D3·D4·D11 기록 완료 (decision-index.md, 보드 기록 d-20261010T053034Z-c50f, d-20261010T053834Z-422b, d-20261010T054047Z-3fc2). 2회차: Q1 개정 (가) 및 F-2 기록 (보드 기록 d-20261010T063222Z-43e0, 보류 d-20261010T063038Z-f1a1 해소)
- 인용 재검증: research.md §3 (연구 입력 3건의 인용 전수), §9 (수리 1회차 재검증)
- 기준선 셀: acceptance.md §C (B-1 회귀 가드, B-2 결함 고정)
- 릴리스 게이트: 0건 (acceptance.md §A)
- 감사 입력 기록: `.moai/reports/t1620/plan-audit.md` §1.4 의 D1–D12 와 O1–O11 을 수리 대상으로 삼았다. 감사 보고서 파일은 수정하지 않았다.
- audit basis: plan-audit-iter11.md (PASS 0.875, audited_sha e34207f32…, plan_artifact_hash bb2aa916…, audit_receipt rcpt-8ad9caf0755e01030d38a921; ruling d-20261010T223148Z-5710)

## §E.2 Run-phase Evidence

**M2 — 진입 훅 비차단과 전달 기록** (REQ-RG-001, REQ-RG-002; AC-RG-001, AC-RG-002, AC-RG-003).

- 시작 기준: HEAD `f4b38e2a3` (WT-review-gate-fix), `git status --short` 출력 없음. 사전 빌드 `go build ./internal/cli/... ./internal/verify/...` 와 `GOOS=windows GOARCH=amd64 go build ./internal/cli/... ./internal/verify/...` 모두 종료 코드 0.
- 명령 형식: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -run '<패턴>' -count=1 2>&1 | tail -n 120`. 이 형식의 종료 코드는 `tail` 의 것이므로, 같은 패턴을 파이프 없이 한 번 더 실행해 `go test` 의 종료 코드를 따로 적었다.

### 1. 편집 이전 기준선 (확장 목록, acceptance.md §C)

- 패턴 `^(TestReviewEntry_SubdirSessionReadsRootReceipt|TestProduceCodexReviewReceipt_SubdirRootStoresAtGitRoot|TestReviewGate_CacheHitReusesFailVerdict|TestReviewEntry_FreshFailBlocks)$` → `ok  github.com/modu-ai/moai-adk/internal/cli  4.760s`. `-list` 매치 4개. 네 테스트 모두 편집 이전에 통과했다(결함 고정).
- `TestProducerFailReceiptPreservesDetailForDelayedBlock`: 편집 이전 측정을 하지 않았다. 사후에 HEAD 의 `internal/cli/codex_review_cache_test.go` 를 임시로 넣어 같은 형식으로 측정했다 → `ok  github.com/modu-ai/moai-adk/internal/cli  2.345s`. 편집본은 바이트 단위로 복원했고 `cmp` 로 확인했다.
- `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` (codex_review_gate_live_test.go): 실행하지 않았다. 이 머신에는 codex 가 PATH 에 있어 실행하면 실제 모델 리뷰가 돈다. 컴파일만 확인했다 (Gap 참조).

### 2. 기대값 변경 — 이전 기대값 → 새 기대값 → 근거 결정

| 테스트 (파일:행) | 이전 기대값 | 새 기대값 | 근거 결정 |
|---|---|---|---|
| `TestReviewEntry_FreshFailBlocks` (codex_review_delay_test.go:395–402) | 진입 결정 `block`, Reason 에 "delayed verdict" 와 상세 | 진입 결정 `block` 아님, AdditionalContext 에 "delayed verdict" 와 상세 | REQ-RG-001 (차단 금지), REQ-RG-002 (추가 맥락 전달). 결함 고정 셀 B-2 (acceptance.md §C) |
| `TestReviewEntry_SubdirSessionReadsRootReceipt` (codex_review_delay_test.go:199–203) | 진입 결정 `block` | 진입 결정 `block` 아님, AdditionalContext 에 "delayed verdict" (루트 고정 영수증을 읽었다는 증거) | REQ-RG-001 |
| `TestProduceCodexReviewReceipt_SubdirRootStoresAtGitRoot` (codex_review_delay_test.go:276–280) | 진입 결정 `block` | 진입 결정 `block` 아님, AdditionalContext 에 상세 | REQ-RG-001, REQ-RG-002 |
| `TestProducerFailReceiptPreservesDetailForDelayedBlock` (codex_review_cache_test.go:224–228) | 진입 결정 `block`, Reason 에 상세 | 진입 결정 `block` 아님, AdditionalContext 에 상세 | REQ-RG-001, REQ-RG-002 |
| `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` (codex_review_gate_live_test.go:136–140, 실행 안 함) | 진입 결정 `block` | 진입 결정 `block` 아님, AdditionalContext 비어 있지 않음 | REQ-RG-001, REQ-RG-002 |

함수 이름은 바꾸지 않았다. 결함을 고정한 테스트는 지우지 않고 기대값만 바꾸었다(plan.md §G). 바뀐 이유는 각 함수의 주석에 적었다.

**정정 기록.** 첫 RED 실행은 `TestReviewGate_CacheHitReusesFailVerdict` 를 패턴에 넣었다. 그러나 216–227행은 `TestProducerFailReceiptPreservesDetailForDelayedBlock` (codex_review_cache_test.go:174–236) 에 속하고, 이 함수는 빠져 있었다. 그 상태로 7개 FAIL 을 얻었다. 이름을 바로잡아 10개를 다시 실행했다. 아래 RED 원문은 정정된 실행이다. 첫 실행의 7개 판정은 같은 원인이므로 여전히 유효하다.

### 3. RED — GREEN 이전 실행

- 컴파일 확인: `go test ./internal/cli/ -run '^$' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/cli  0.916s [no tests to run]`. TOOL_FAILURE 없음.
- RED 패턴(10개 이름): `^(TestReviewEntry_FreshFailNeverBlocks|TestReviewEntry_FirstPromptDeliversDetailOnce|TestReviewEntry_RepeatPromptOmitsDetail|TestReviewEntry_NewKeyDeliversAgain|TestReviewEntry_InconclusiveAllowsWithoutBlock|TestReviewEntry_FreshFailBlocks|TestReviewEntry_SubdirSessionReadsRootReceipt|TestProduceCodexReviewReceipt_SubdirRootStoresAtGitRoot|TestProducerFailReceiptPreservesDetailForDelayedBlock|TestReviewGate_CacheHitReusesFailVerdict)$`
- 종료 코드 (파이프 없는 같은 패턴): **1**. `--- FAIL` 줄 8개.
- 원문 (`go test` 출력 그대로; 행 번호는 RED 실행 시점의 값):

```text
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
codex review: fail: Verdict: fail

blocking finding on shared file.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestProducerFailReceiptPreservesDetailForDelayedBlock (1.45s)
    codex_review_cache_test.go:224: a fresh FAIL receipt must not block at turn entry (REQ-RG-001), got &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate (delayed verdict — the background review of the previous turn failed): fail
        
        # codex review fail — fail
        
        Verdict: fail
        
        blocking finding on shared file.
        
        ## Findings HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewEntry_SubdirSessionReadsRootReceipt (0.73s)
    codex_review_delay_test.go:199: a subdir session must read the root-anchored FAIL receipt and not block, got &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate (delayed verdict — the background review of the previous turn failed): fail HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
codex review: fail: Verdict: fail

blocking finding on shared file.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestProduceCodexReviewReceipt_SubdirRootStoresAtGitRoot (1.10s)
    codex_review_delay_test.go:276: the subdir session's entry must not block on the root-stored fail, got &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate (delayed verdict — the background review of the previous turn failed): fail
        
        # codex review fail — fail
        
        Verdict: fail
        
        blocking finding on shared file.
        
        ## Findings HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewEntry_FreshFailBlocks (1.10s)
    codex_review_delay_test.go:395: a fresh FAIL receipt must not block at turn entry (REQ-RG-001), got &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate (delayed verdict — the background review of the previous turn failed): fail
        
        # codex review fail — fail
        
        reviewer failed the turn
        
        ## Findings
        
        - [P1] delayed finding — blocking finding on shared file. (main.go:3) HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewEntry_FreshFailNeverBlocks (1.07s)
    codex_review_entry_delivery_test.go:121: AC-RG-001: a fresh FAIL receipt must not block the prompt; got decision=block with reason="codex review gate (delayed verdict — the background review of the previous turn failed): fail\n\n# codex review fail — fail\n\nreviewer failed the turn\n\n## Findings\n\n- [P1] delayed finding — blocking finding on shared file. (main.go:3)"
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewEntry_FirstPromptDeliversDetailOnce (1.05s)
    codex_review_entry_delivery_test.go:138: AC-RG-002: the first prompt must carry the finding detail exactly once as additional context; got 0 occurrence(s) in additionalContext=""
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewEntry_RepeatPromptOmitsDetail (1.55s)
    codex_review_entry_delivery_test.go:159: AC-RG-003: the second prompt of a delivered key must not repeat the detail; reason="codex review gate (delayed verdict — the background review of the previous turn failed): fail\n\n# codex review fail — fail\n\nreviewer failed the turn\n\n## Findings\n\n- [P1] delayed finding — blocking finding on shared file. (main.go:3)" additionalContext=""
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewEntry_NewKeyDeliversAgain (1.93s)
    codex_review_entry_delivery_test.go:183: REQ-RG-002: a changed review key must deliver its own detail exactly once as additional context; got 0 occurrence(s) in additionalContext=""
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
codex review entry: background verdict inconclusive for the unchanged tree — allowing
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	12.167s
FAIL
```

- 분류:
  - EXPECTED_RED 8개. 각 실패는 의도한 단언에서 났다. 현재 코드의 실제 출력은 `Decision: block` 이고 상세는 Reason 에 있다. AdditionalContext 는 비어 있다.
    - `TestReviewEntry_FreshFailNeverBlocks` (AC-RG-001, codex_review_entry_delivery_test.go:121): 차단 결정.
    - `TestReviewEntry_FirstPromptDeliversDetailOnce` (AC-RG-002, :138): 추가 맥락에 상세 0회.
    - `TestReviewEntry_RepeatPromptOmitsDetail` (AC-RG-003, :159): 둘째 프롬프트의 Reason 에 상세가 다시 실림.
    - `TestReviewEntry_NewKeyDeliversAgain` (REQ-RG-002, :183): 새 키의 추가 맥락에 상세 0회.
    - 기대값 변경 4개 (위 표 중 앞의 넷): 차단 결정.
  - GUARD 2개 (편집 이전에도 통과해야 하는 가드, RED 아님): `TestReviewEntry_InconclusiveAllowsWithoutBlock`, `TestReviewGate_CacheHitReusesFailVerdict`. 둘 다 단독 실행 `ok`.
  - TOOL_FAILURE 0, REGRESSION_FAILURE 0.

### 4. GREEN — 편집 이후

- 구현 (codex_review_delay.go): 진입 훅 (6) 분기는 차단 대신 `hookSpecificOutput.additionalContext` 로 상세를 한 번 전달한다. 전달 기록은 영수증 결합 필드 다섯 개(head, tree_digest, config_digest, command, tool_version)의 길이 접두 SHA-256 을 키로 하는 파일이다: `.moai/state/verify/codex-review/delivered-<hex>.json`. 배타 생성(O_EXCL)이 선점이다. 기록을 쓸 수 없으면 전달하는 쪽으로 fail-open 한다.
- 트리 키 불변: 기록은 `.moai/state/` 아래에 있다. `verify.Key` 는 `git ls-files --others --exclude-standard` 가 나열한 미무시 파일만 해시하므로 기록이 트리 키를 바꾸지 않는다. 픽스처와 실제 저장소 모두 `.moai/state/` 를 무시한다.
- AC-RG-003 비공허 전제: `TestReviewEntry_RepeatPromptOmitsDetail` 에 "첫 프롬프트 뒤에도 같은 키의 fail 영수증이 fresh 임" 전제 단언을 GREEN 이후에 추가했다. 행위 단언이 아니다. 이 전제는 편집 이전 코드에서도 성립한다. RED 원문의 둘째 프롬프트 Reason 에 fresh fail 상세가 실려 있기 때문이다. 전제가 없으면 키가 움직였을 때 침묵이 빈 통과가 된다.
- GREEN 명령: 위 RED 패턴과 같다. 종료 코드 **0**, `--- FAIL` 0. 원문:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	13.666s
```

- 이웃 패턴 (정규식 `TestReviewEntry_.*` 과 명시한 13개 이름, 생성·수거 마커 테스트와 턴 종료 가드 포함): `ok  github.com/modu-ai/moai-adk/internal/cli  20.592s`.

### 5. 검증 수치

- 빌드: host exit 0, windows/amd64 exit 0 (`./internal/cli/... ./internal/verify/...`).
- `gofmt -l` (변경 파일): 출력 없음.
- `golangci-lint run ./internal/cli/... ./internal/verify/... --timeout=2m` → `0 issues.` 첫 실행에서 errcheck 1건(전달 기록 해시의 `fmt.Fprintf` 반환값)이 나와 `_, _ =` 로 고쳤다.
- internal/verify 패키지 커버리지: `coverage: 83.6% of statements`. 이 마일스톤에서 internal/verify 파일은 수정하지 않았다. 패키지의 기존 값이다.
- internal/cli 패키지 전체 커버리지: 측정하지 않았다 (Gap). 변경 함수의 함수 단위 커버리지는 이웃 패턴 부분집합에서 측정했다. 측정 시점의 코드이며, 그 뒤의 변경은 반환값 무시 표기와 주석뿐이다: `HandleCodexReviewEntry` 86.7%, `codexReviewClaimDelivery` 90.9%, `codexReviewDeliveryPath` 100.0%.

### 6. Gap 과 잔여 위험

- Gap:
  - (a) 라이브 테스트 `TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` 는 컴파일만 했다. codex 가 있어 실행하면 실제 모델 리뷰가 돈다.
  - (b) internal/cli 패키지 전체 커버리지 미측정(전체 스위트 금지).
  - (c) `/verify` 게이트를 실행하지 않았다. 이 게이트는 테스트 단계를 전체 범위로 돌리고 `.moai/state` 에 스냅샷을 기록하므로, 이 실행의 범위 밖이다.
  - (d) 저장소의 `core.hooksPath` 가 `/dev/null` 이므로 커밋 시 git 훅이 실행되지 않았다(관측).
  - (e) AskUserQuestion 검사(지정 명령)는 빈 출력이 아니었다. 매치는 도움말 문구, 린트 규칙 본문, `internal/hook/pre_tool.go:856` 의 도구 이름 검사다. 호출 지점은 없고, 이 마일스톤의 diff 에는 AskUserQuestion 이 0건이다.
- 잔여 위험:
  - (a) 같은 키가 24시간 TTL 을 넘긴 뒤 다시 fail 로 나오면, 키당 한 번 규칙 때문에 상세가 다시 실리지 않는다.
  - (b) 전달 기록 파일은 키마다 하나씩 쌓인다(GC 없음. 기존 상세 파일과 같은 방식).
  - (c) 기록을 쓸 수 없으면 매 프롬프트 전달된다(fail-open).
  - (d) 기록 생성 뒤 출력 전에 프로세스가 끊기면 그 키의 상세는 전달되지 않는다.
- B2 교차 SPEC 점검: `grep -n "Retired\|superseded" internal/cli/codex_review_*.go` 의 적중은 `codex_review_target_test.go:202` 하나다. 기준 브랜치 해석(t1426)에 관한 주석이며 진입 훅의 차단·전달과 무관하다. 충돌 없음.
- SPEC 상태: spec.md 의 `status: draft` (version 0.2.8) 는 그대로다. spec-frontmatter-schema.md 에 따르면 `draft → in-progress` 전이는 첫 run 단계 커밋의 manager-develop 소관이다. 이 마일스톤 범위는 SPEC 본문 편집을 금지하므로 수행하지 않았다. 리더 판단 대상이다.

### 7. 커밋

이 파일은 자기 커밋의 SHA 를 적지 않는다. 커밋 메시지와 `git log` 가 기록이다.

## §E.3 Run-phase Audit-Ready Signal

_run-phase 진입 전. 자리표시._

## §E.4 Sync-phase Audit-Ready Signal

_sync-phase 진입 전. 자리표시._

## §G Override and Refusal Record

- 2026-10-10T22:44:14Z SPEC-REVIEW-GATE-REPAIR-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 11 >= tier ceiling 3); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter10.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter11.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter8.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter9.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit.md
- 2026-10-10T23:24:45Z SPEC-REVIEW-GATE-REPAIR-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 11 >= tier ceiling 3); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter10.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter11.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter8.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter9.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit.md

- 2026-10-10T23:27:18Z SPEC-REVIEW-GATE-REPAIR-001 operator-decision kickoff=approve decider=human actor=operator engine-event=seq-1250 (card.transition kickoff->assigned v57) source=AskUserQuestion(영실이 세션; 리더 전달, 레인은 AskUserQuestion 자체를 보지 않음) board=d-20261010T232653Z-a596 codex-leg=advisory(2413·a280) lane=transcribed-from-engine-event-not-authored
- 귀속 표 (영실이 판단, 리더 전달, 2026-10-10): F1 → t1662 (relate 142) · F2 → t1637 (relate 143) · F3 → t1639 (relate 145; 본문 추가 todo_runtime.go:364, todo_dispatch_current.go:53 은 기존) · F4 → t1683 (doctor 경고만 맡고, 규칙 문장 정정은 범위 밖이다. 정정은 M7 에서 수행) · F5 → t1674 (relate 146; 본문 "plan 시작 시 반영") · F6 → t1635 (relate 144).
