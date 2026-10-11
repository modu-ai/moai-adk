# 진행 기록 — SPEC-REVIEW-GATE-REPAIR-001

상태: completed (3-phase close, amendment 뒤 재종결). 이 파일의 §E.1 ~ §E.4 는 각 단계 담당자가 채운 기록이다.

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

**M3 — 턴 종료 상태표: 차단 횟수, 메시지** (REQ-RG-003, REQ-RG-004, REQ-RG-007; AC-RG-004, AC-RG-005, AC-RG-017, AC-RG-018, AC-RG-019, AC-RG-025).

- 시작 기준: HEAD `988230ea3` (WT-review-gate-fix), `git status --short` 출력 없음. 사전 빌드 `go build ./internal/cli/... ./internal/verify/...` 와 `GOOS=windows GOARCH=amd64 go build ./internal/cli/... ./internal/verify/...` 모두 종료 코드 0.
- 명령 형식: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/cli/ -run '<패턴>' -count=1 > <scratchpad 파일> 2>&1; echo "EXIT=$?"`. 파이프를 쓰지 않았으므로 종료 코드는 `go test` 의 것이다. PASS 줄 수를 세려고 같은 패턴에 `-v` 만 더한 재실행을 따로 했다.

### 1. RED — GREEN 이전 실행

- 패턴(8개 이름): `^(TestReviewGate_BlockCountIsOneAfterFirstBlock|TestReviewGate_BlockCapAllowsNextTurnWithMessage|TestReviewGate_AllowedAfterCapStaysAllowedAndCountKept|TestReviewGate_FreshPassClearsBlockCount|TestReviewGate_BlockCountIsPerReviewKey|TestReviewGate_NonFreshStateNamedInMessage|TestReviewGate_NoReviewAllowsCarryStateMessage|TestReviewGate_FreshInconclusiveAllowsWithMessage)$`
- 종료 코드(파이프 없음): **1**. 최상위 8개 전부 FAIL, 하위 테스트 FAIL 7개. 컴파일 실패 없음(TOOL_FAILURE 0).
- 원문 (`go test` 출력 그대로):

```text
codex review: fail: Verdict: fail

blocking finding on shared file.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewGate_BlockCountIsOneAfterFirstBlock (1.06s)
    codex_review_gate_test.go:387: block count for the key after one block = 0, want 1
codex review: fail: Verdict: fail

blocking finding on shared file.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewGate_BlockCapAllowsNextTurnWithMessage (1.22s)
    codex_review_gate_test.go:401: want allow naming state "차단 횟수 소진", got block: &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate (cached verdict): fail
        
        # codex review fail — fail
        
        Verdict: fail
        
        blocking finding on shared file.
        
        ## Findings HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
codex review: fail: Verdict: fail

blocking finding on shared file.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewGate_AllowedAfterCapStaysAllowedAndCountKept (1.57s)
    codex_review_gate_test.go:414: want allow naming state "차단 횟수 소진", got block: &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate (cached verdict): fail
        
        # codex review fail — fail
        
        Verdict: fail
        
        blocking finding on shared file.
        
        ## Findings HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
codex review: fail: Verdict: fail

blocking finding on shared file.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewGate_FreshPassClearsBlockCount (1.48s)
    codex_review_gate_test.go:432: want allow naming state "차단 횟수 소진", got block: &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate (cached verdict): fail
        
        # codex review fail — fail
        
        Verdict: fail
        
        blocking finding on shared file.
        
        ## Findings HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
codex review: fail: Verdict: fail

blocking finding on shared file.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewGate_BlockCountIsPerReviewKey (1.72s)
    codex_review_gate_test.go:458: want allow naming state "차단 횟수 소진", got block: &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision:block Reason:codex review gate (cached verdict): fail
        
        # codex review fail — fail
        
        Verdict: fail
        
        blocking finding on shared file.
        
        ## Findings HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
codex review gate: background review started for this tree state; the verdict is enforced at the next turn entry (codex-review-entry)
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
codex review gate: background review started for this tree state; the verdict is enforced at the next turn entry (codex-review-entry)
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
codex review gate: background review started for this tree state; the verdict is enforced at the next turn entry (codex-review-entry)
--- FAIL: TestReviewGate_NonFreshStateNamedInMessage (2.22s)
    --- FAIL: TestReviewGate_NonFreshStateNamedInMessage/absent (0.75s)
        codex_review_gate_test.go:501: systemMessage must name state "absent", got ""
    --- FAIL: TestReviewGate_NonFreshStateNamedInMessage/stale (0.74s)
        codex_review_gate_test.go:501: systemMessage must name state "stale", got ""
    --- FAIL: TestReviewGate_NonFreshStateNamedInMessage/key-mismatch (0.73s)
        codex_review_gate_test.go:501: systemMessage must name state "key-mismatch", got ""
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
--- FAIL: TestReviewGate_NoReviewAllowsCarryStateMessage (0.74s)
    --- FAIL: TestReviewGate_NoReviewAllowsCarryStateMessage/switch_off (0.09s)
        codex_review_gate_test.go:521: systemMessage must name state "스위치 꺼짐", got ""
    --- FAIL: TestReviewGate_NoReviewAllowsCarryStateMessage/no_reviewable_change (0.25s)
        codex_review_gate_test.go:526: systemMessage must name state "검토할 변경 없음", got ""
    --- FAIL: TestReviewGate_NoReviewAllowsCarryStateMessage/stop-hook_release (0.12s)
        codex_review_gate_test.go:534: systemMessage must name state "stop-hook 해제", got ""
    --- FAIL: TestReviewGate_NoReviewAllowsCarryStateMessage/reviewer_absent (0.28s)
        codex_review_gate_test.go:539: systemMessage must name state "reviewer 없음", got ""
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
codex review gate: reusing the cached inconclusive verdict for the unchanged tree (skip #10)
--- FAIL: TestReviewGate_FreshInconclusiveAllowsWithMessage (0.73s)
    codex_review_gate_test.go:550: systemMessage must name state "inconclusive", got ""
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	11.646s
FAIL
```

- 분류:
  - EXPECTED_RED 8개(최상위). 실패 단언은 13개다(`codex_review_gate_test.go` 의 `t.Errorf`/`t.Fatalf` 줄). 각 실패는 의도한 단언에서 났다. 첫 차단의 횟수가 기록되지 않음(`:387`), 두 번째 턴 종료가 차단됨(`:401`, `:414`), 상태 이름 누락(`:501` 3개, `:521`·`:526`·`:534`·`:539`, `:550`).
  - `FreshPassClearsBlockCount`(`:432`)와 `BlockCountIsPerReviewKey`(`:458`)는 그 앞의 cap 전제 단언에서 실패했다. 청산 단언과 키별 단언은 RED 에서 도달하지 못했고 GREEN 에서 도달했다.
  - TOOL_FAILURE 0. 새 테스트 함수이므로 REGRESSION_FAILURE 는 해당하지 않는다.

### 2. GREEN — 편집 이후

- 구현 (`internal/cli/codex_review_gate.go`):
  - 턴 종료 차단 횟수, C = 1 (`codexReviewBlockLimit`). 리뷰 키 단위 기록은 `.moai/state/verify/codex-review/blocked-<hex>.json` 이다. 이름은 같은 키의 전달 기록(`delivered-<hex>.json`, codex_review_delay.go)에서 접두만 바꿔 얻으므로 두 기록이 키를 다르게 셀 수 없다. 차단할 때는 기록을 먼저 쓰고 차단을 돌려준다. 허용은 기록을 바꾸지 않는다. fresh pass 는 기록을 지운다.
  - 상태 메시지는 설계 §7 의 어휘 아홉 상태를 쓴다. 결정은 바뀌지 않는다(REQ-RG-006). 키 상태는 absent, key-mismatch, stale 로 나눈다.
  - 검토 시작(kick) 경로와 차단 사유의 형식은 바꾸지 않았다.
- GREEN 명령: 위 RED 패턴과 같다. 종료 코드 **0**. 원문: `ok  github.com/modu-ai/moai-adk/internal/cli  13.310s` ([no tests to run] 없음).
- `-v` 재실행(같은 패턴): PASS 줄 15개(최상위 8, 하위 테스트 7).
- 기존 네 테스트(`TestReviewGate_DisabledAllows`, `TestReviewGate_LoopPreventionAllows`, `TestReviewGate_FailOpenOnMissingCodex`, `TestReviewGate_CacheHitReusesFailVerdict`): 이름으로 실행, 종료 코드 0, PASS 4.
- 회귀 비교: Stop 게이트 핸들러를 호출하는 기존 테스트 42개(이름은 `internal/cli/*_test.go` 의 호출 지점을 감싸는 최상위 함수에서 추출). 편집 이전(RED 시점, 구현 전)과 이후 모두 FAIL 1개이며 같은 테스트 `TestCodexReviewGatePrimaryCheckoutSkip`, 같은 메시지다. 나머지 41개는 PASS. 이 실패는 M3 와 무관하다. 임시 디렉터리 경로의 `/var` 와 `/private/var` 심볼릭 링크 불일치(`codex_review_gate_primary_scope_red_test.go:84`)이며 편집 이전 HEAD 코드에서 이미 실패했다.

### 3. AC 대응

| AC | 테스트 | 결과 |
|---|---|---|
| AC-RG-004 | `TestReviewGate_BlockCountIsOneAfterFirstBlock` | PASS |
| AC-RG-005 | `TestReviewGate_BlockCapAllowsNextTurnWithMessage` | PASS |
| AC-RG-017 | `TestReviewGate_AllowedAfterCapStaysAllowedAndCountKept` | PASS |
| AC-RG-018 | `TestReviewGate_NonFreshStateNamedInMessage` (absent, stale, key-mismatch) | PASS 3/3 |
| AC-RG-019 | `TestReviewGate_NoReviewAllowsCarryStateMessage` (스위치 꺼짐, 검토할 변경 없음, stop-hook 해제, reviewer 없음) | PASS 4/4 |
| AC-RG-025 | `TestReviewGate_FreshInconclusiveAllowsWithMessage` | PASS |
| AC 없음 (REQ-RG-004 청산 규칙) | `TestReviewGate_FreshPassClearsBlockCount` | PASS |
| AC 없음 (REQ-RG-004 키별 규칙) | `TestReviewGate_BlockCountIsPerReviewKey` | PASS |

AC 번호가 없는 두 테스트는 플랜 M3 의 RED 목록 밖이다. 청산 규칙과 키별 규칙이 테스트 없이 남는 것을 막으려고 추가했다. 리더 확인 대상이다.

### 4. 기대값 변경

- 없음. `codex_review_gate_test.go` 의 기존 테스트 함수는 한 줄도 바꾸지 않았다. 추가한 것은 import 세 개(`encoding/json`, `time`, `internal/verify`)와 새 테스트 함수 및 그 헬퍼다. M2 파일(`codex_review_delay.go`, `codex_review_delay_test.go`, `codex_review_entry_delivery_test.go`, `codex_review_cache_test.go`, `codex_review_gate_live_test.go`)은 건드리지 않았다.

### 5. 검증 수치

- 빌드: host exit 0, windows/amd64 exit 0 (`./internal/cli/... ./internal/verify/...`).
- `gofmt -l` (변경 파일 2개): 출력 없음.
- `golangci-lint run ./internal/cli/... --timeout=2m` → `0 issues.`, 종료 코드 0, 시간 초과 없음.
- 함수 단위 커버리지 (부분집합: M3 테스트 8개와 기존 네 개의 `-coverprofile`): HandleCodexReviewGate 82.2%, codexReviewGateFail 80.0%, codexReviewKeyState 100.0%, codexReviewBlockPath 100.0%, codexReviewBlockCount 77.8%, codexReviewRecordBlock 80.0%, codexReviewClearBlocks 100.0%, isPassVerdict 100.0%, codexGateAllowed 100.0%. 비어 있는 분기는 기록 쓰기 실패와 판독 실패의 fail-open 경로다.
- internal/cli 패키지 전체 커버리지: 측정하지 않았다 (Gap).

### 6. Gap 과 잔여 위험

- Gap:
  - (a) 정책 건너뜀 두 경로, 즉 오래된 바이너리(`staleBinarySkipApplies`)와 tree_scope 건너뜀은 메시지를 내지 않는다. 설계 §7 의 아홉 상태 어휘에 두 경로의 이름이 없다. REQ-RG-007 의 문언("검토 없이 허용되는 턴 종료")은 두 경로를 포함하므로 어휘를 정하는 판정이 필요하다. 리더 판정 대상이다.
  - (b) 차단 횟수 기록을 쓸 수 없거나 읽을 수 없으면 허용하고 `차단 횟수 소진` 을 낸다(fail-open). 어휘에 맞는 이름이 없어 가장 가까운 이름을 골랐다. 리더 판정 대상이다.
    - → 해결 (65c24753a): (a) 정책 건너뜀 두 경로(stale 바이너리 건너뜀, tree_scope 건너뜀)가 이제 각자의 상태 이름을 systemMessage 에 낸다. 허용 결정은 그대로다.
    - → 해결 (65c24753a): (b) 차단 횟수를 읽을 수 없거나 쓸 수 없으면 이제 `차단 횟수 기록 실패`를 낸다. `차단 횟수 소진`은 C 도달 상한 경로에만 남는다.
  - (c) `/verify` 게이트를 실행하지 않았다. 이 실행의 범위 밖이다.
  - (d) 저장소의 `core.hooksPath` 가 `/dev/null` 이므로 커밋 시 git 훅이 실행되지 않았다(관측).
  - (e) internal/cli 패키지 전체 커버리지는 측정하지 않았다. 위 부분집합 수치는 패키지 수치가 아니다.
  - (f) 쓰기와 판독 실패의 fail-open 분기는 M3 테스트가 실행하지 않았다(위 커버리지 표의 빈 부분).
  - (g) 기존 테스트 `TestCodexReviewGatePrimaryCheckoutSkip` 은 편집 이전부터 실패하며(macOS 임시 디렉터리 심볼릭 링크), M3 와 무관하다. 리더 확인 대상으로 남긴다.
- 잔여 위험:
  - (a) 두 Stop 이 같은 키에서 같은 순간에 차단하면 횟수가 한 번만 올라갈 수 있다. 읽기와 쓰기 사이에 잠금이 없다.
  - (b) 차단 기록 파일은 키마다 하나씩 쌓인다(GC 없음. 전달 기록과 같은 방식).
  - (c) `codexReviewBlockLimit` 은 `codex_review_gate.go` 에 있다. 프로젝트의 임계값 규칙(`internal/config/defaults.go`)과 다르다. 옮기려면 `internal/config` 를 바꿔야 해서 이 마일스톤의 파일 목록 밖이다.
- SPEC 본문과 frontmatter 는 편집하지 않았다(M3 범위 밖).

### 7. 커밋

이 파일은 자기 커밋의 SHA 를 적지 않는다. 커밋 메시지와 `git log` 가 기록이다.

### M3 후속 — RED→GREEN, 기대값 변경, 리더 결정

- RED 커밋: `1fc6ef92f` (`test(SPEC-REVIEW-GATE-REPAIR-001): add RED tests for named skip and record-failure states (card t1620)`). 테스트 파일만 바뀌었다. 바뀐 파일은 `internal/cli/codex_review_gate_test.go` 하나이고 numstat 은 106 insertions · 0 deletions 다. 기존 줄은 한 줄도 바뀌지 않았고, 프로덕션 코드는 이 커밋에 없다. 정책 건너뜀 두 경로와 기록 실패 두 경로의 상태 이름을 확인하는 테스트 네 개다.
- RED 명령: `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -v -skip LiveCodex -run '^(TestReviewGate_StaleBinarySkipNamesState|TestReviewGate_TreeScopeSkipNamesState|TestReviewGate_BlockCountReadFailureNamesRecordFailure|TestReviewGate_BlockCountWriteFailureNamesRecordFailure)$' ./internal/cli/`. 원문 (`.moai/state/verify/t1620-m3-red.txt` 28줄을 그대로 옮김):

```text
=== RUN   TestReviewGate_StaleBinarySkipNamesState
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
    codex_review_gate_test.go:571: systemMessage must name state "stale 바이너리 건너뜀", got ""
--- FAIL: TestReviewGate_StaleBinarySkipNamesState (0.24s)
=== RUN   TestReviewGate_TreeScopeSkipNamesState
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
    codex_review_gate_test.go:590: systemMessage must name state "tree_scope 건너뜀", got ""
--- FAIL: TestReviewGate_TreeScopeSkipNamesState (0.24s)
=== RUN   TestReviewGate_BlockCountReadFailureNamesRecordFailure
codex review: fail: Verdict: fail

blocking finding on shared file.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
    codex_review_gate_test.go:616: systemMessage must name state "차단 횟수 기록 실패", got "codex review gate: turn allowed without a fresh pass (state: 차단 횟수 소진)"
    codex_review_gate_test.go:618: an unreadable count must not be named as exhausted, got "codex review gate: turn allowed without a fresh pass (state: 차단 횟수 소진)"
--- FAIL: TestReviewGate_BlockCountReadFailureNamesRecordFailure (1.12s)
=== RUN   TestReviewGate_BlockCountWriteFailureNamesRecordFailure
codex review: fail: Verdict: fail

blocking finding on shared file.
{"basis":"no card branch: main","branch":"main","gate":"codex-review-gate","scope":"tree"}
codex review gate: block count not recorded (create temp file: open /private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestReviewGate_BlockCountWriteFailureNamesRecordFailure2539332822/001/.moai/state/verify/codex-review/.settings-local-3335700200.tmp: permission denied) — allowing (fail-open)
    codex_review_gate_test.go:653: systemMessage must name state "차단 횟수 기록 실패", got "codex review gate: turn allowed without a fresh pass (state: 차단 횟수 소진)"
    codex_review_gate_test.go:655: an unrecordable count must not be named as exhausted, got "codex review gate: turn allowed without a fresh pass (state: 차단 횟수 소진)"
--- FAIL: TestReviewGate_BlockCountWriteFailureNamesRecordFailure (1.53s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	4.087s
FAIL
```

- RED 분류 (tdd-result-contract):
  - 네 개 모두 EXPECTED_RED 다. 네 개의 전제가 모두 성립했다.
  - 네 경우 모두 허용(allow)이고 빠진 것은 상태 메시지다.
  - TOOL_FAILURE 와 REGRESSION_FAILURE 는 없다. 쓰기 실패 경로의 stderr 줄(`permission denied`)은 실제로 쓰기가 실패했음을 보여 준다.
- GREEN 커밋: `65c24753a` (`fix(SPEC-REVIEW-GATE-REPAIR-001): name the skip and record-failure states at turn end (card t1620)`). 프로덕션 파일 `internal/cli/codex_review_gate.go` 하나이고 29 insertions · 18 deletions 다. 주석 한 줄(C = 1, operator ruling Q2)도 이 커밋에 들어 있다. 커밋 순서(RED 가 GREEN 보다 먼저)는 리더가 확인했다.
- GREEN 결과: RED 네 테스트를 GREEN 에서 실행한 원문 (`.moai/state/verify/t1620-m3-green.txt`):

```text
--- PASS: TestReviewGate_StaleBinarySkipNamesState (0.25s)
--- PASS: TestReviewGate_TreeScopeSkipNamesState (0.24s)
--- PASS: TestReviewGate_BlockCountReadFailureNamesRecordFailure (1.13s)
--- PASS: TestReviewGate_BlockCountWriteFailureNamesRecordFailure (1.14s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.683s
```

- GREEN 게이트 계열 (기준선과 같은 `-run` 패턴, `.moai/state/verify/t1620-m3-green-family.txt`): 최상위 `--- PASS` 79, `--- FAIL` 3. 전체는 82개다(기준선 65 + M2 5 + M3 8 + RED 4). FAIL 셋은 기준선에서 이미 있던 심볼릭 링크 실패다(Run-phase Gap 1). 패키지 판정 줄: `FAIL	github.com/modu-ai/moai-adk/internal/cli	68.462s`.
- GREEN 린트: `golangci-lint run ./internal/cli/... --timeout=2m` 최종 줄 `0 issues.` (`.moai/state/verify/t1620-m3-green-lint.txt`).
- 게이트 계열 밖 (`.moai/state/verify/t1620-m3-green-extra.txt`): 최상위 PASS 18개다(`TestStaleBinarySkip_*` 3, `TestTreeScopeReader_*` 2, `TestPrimaryScopeReader_*` 2, `TestCodexStopChain_TreeScopeSkip`, `TestTreeScope_NonSkipValuesKeepTreeScope`, `TestTreeScopeSkip_WTSessionsStillReviewed`, `TestTreeScopePolicy_*` 6, `TestTreeScopeSkipRow`, `TestStopChainEffectParityGolden`). 이 실행의 FAIL 은 `TestWSR006_ReviewGateRootMatrix` 하나다. M3 에서 생긴 회귀이며, 아래 WSR006 소절에 적었다.

### WSR006 기대값 변경 — 다른 SPEC 테스트, 리더 결정 1

- 대상: `TestWSR006_ReviewGateRootMatrix` (`internal/cli/wsr_state_root_test.go`). 다른 SPEC 의 테스트다. 바뀐 파일은 이 하나이고 numstat 은 5/2, gofmt 출력은 없다.
- 이전 값 (원문): `if len(*calls) != 0 || codexOut != "{}" || multiOut != "{}"`
- 새 값: `codexOut != wantCodexSwitchOff`. multi 절은 그대로 `"{}"` 다. 여기서 `wantCodexSwitchOff` 는 `{"systemMessage":"codex review gate: turn allowed without a fresh pass (state: 스위치 꺼짐)"}` 다.
- 근거: REQ-RG-007 과 SPEC D4 (허용 시 항상 메시지). 되돌리면 SPEC 과 충돌한다.
- 원인 커밋: `1002a14c1` (M3). 스위치 꺼짐 분기가 bare allow 에서 이름 붙은 메시지로 바뀌었다.
- 기대값 변경 커밋: `4f94abf73` (`test(SPEC-REVIEW-GATE-REPAIR-001): expect the named switch-off state on the codex stop path (card t1620)`). 커밋 메시지에 같은 근거를 원문으로 적었다.
- 리더 결정 1: 기대값을 이름 붙은 메시지로 바꾼다. 제시된 두 안 가운데 (b) 스위치 꺼짐 경로를 다시 조용한 allow 로 되돌리는 안은 D4 위반으로 기록되어 있어 택하지 않았다.
- 회귀 증거 (원문):
  - 988230ea3 (M3 의 부모) 추출본: `--- PASS: TestWSR006_ReviewGateRootMatrix/F-noid_row_4 (0.25s)`, `ok  	github.com/modu-ai/moai-adk/internal/cli	14.599s` (scratch `wsr-988230ea3.txt`).
  - GREEN 직전 트리: `--- FAIL: TestWSR006_ReviewGateRootMatrix/F-noid_row_4 (0.19s)` (`.moai/state/verify/t1620-m3-baseline-red-and-wsr006.txt`).
  - GREEN 트리: `wsr_state_root_test.go:425: admin HEAD deleted: want N′ (both {} and no detector call), got codex={"systemMessage":"codex review gate: turn allowed without a fresh pass (state: 스위치 꺼짐)"} multi={} calls=[]` (`.moai/state/verify/t1620-m3-green-wsr.txt`). `no git` 변형도 같은 출력이다.
  - 원인: 테스트는 codex 경로에서 정확히 `{}` 를 요구한다(`wsr_state_root_test.go:424`). GREEN 은 이 출력을 바꾸지 않았다.
- 검증 산출 (디스크 확인): `--- PASS: TestWSR006_ReviewGateRootMatrix (9.96s)` (`.moai/state/verify/t1620-wsr-after.txt`). 같은 계열 전체는 최상위 `--- PASS` 13, `--- FAIL` 0 이다 (같은 디렉터리의 `t1620-wsr-all.txt`). `^func TestWSR` 정의는 13개다 (`git grep` 확인).

### M7 — 규칙 문장 정정, 전후 비교표, 후속 카드 매핑

- 커밋 `34e8a6973` (`docs(SPEC-REVIEW-GATE-REPAIR-001): correct the card-review double-review sentence (card t1620)`). 바뀐 파일은 규칙 파일 두 개, 라이브 `.claude/rules/moai/workflow/factory-dispatch-detail.md` 와 템플릿 미러 `internal/template/templates/…` 이다. numstat 은 2 insertions · 2 deletions 다.
- 커밋 전 점검 (디스크):
  - `git diff --stat`: 3 files, 3 insertions, 2 deletions. 세 번째 파일은 `internal/cli/codex_review_gate.go` 의 주석 한 줄이며, 이 커밋에는 넣지 않았다.
  - `gofmt -l internal/cli/codex_review_gate.go`: 출력 없음.
  - `cmp` (두 규칙 파일): 동일, 출력 없음.
- 템플릿 키 확인: 라이브 `.moai/config/sections/workflow.yaml:300–302` 에 `codex:` → `review_gate:` → `enabled: true`. 배포 템플릿 `internal/template/templates/.moai/config/sections/workflow.yaml:152–154` 에 같은 경로가 `enabled: false`(opt-in)로 있다. 규칙 문장은 "켜 둔 저장소"를 가리키므로 맞다.
- 가드 테스트 (라이브 트리, 커밋 전): 명령 `go test -count=1 -v -run '^(TestCardIDBaselineHasNoStaleEntries|TestCardIDBaselineIsPerFileAndPerLiteral|TestCardIDLeakClassDetectsShapes|TestCardIDLeakClassIgnoresLookalikes|TestDeclaredRuleMirrorForks|TestRuleTemplateMirrorDrift|TestReviewOwnership_CardReviewDoctrine|TestReviewOwnership_DoctrineMutants)$' ./internal/template/`. 판정 줄: `--- PASS` 8개(위 여덟 이름 전부), `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/template	0.451s`. 실패 0.
- 전후 비교표 (기록 2026-10-11T01:14Z, 리더 결정 전 상태, 원문):

| 계열 | 기준선 `f4b38e2a3` (M2 이전) | 현재 `65c24753a` | 근거 파일 · 비고 |
|---|---|---|---|
| 게이트 family (`-run` 패턴 동일, 최상위) | 65 개: PASS 62, FAIL 3 | 82 개: PASS 79, FAIL 3 | `before-run.txt`, `t1620-m3-green-family.txt`. 증가 17 = M2 신규 5 + M3 신규 8 + RED 4. FAIL 3 은 기준선과 같은 테스트(symlink)다. |
| 기대값 변경 (family 안) | 해당 없음 | M2 에서 기록한 변경(진행 기록 §2)뿐. M3 와 GREEN 은 0 | GREEN 커밋은 프로덕션 파일 하나만 바꿨다(`git show --stat 65c24753a`). |
| family 밖 계열 (stop-chain, tree-scope 등) | 측정하지 않았다 (Gap) | PASS 18, FAIL 0 (WSR 제외) | `t1620-m3-green-extra.txt`. 기준선에서 같은 계열을 돌리지 않았다. |
| `TestWSR006_ReviewGateRootMatrix` | `988230ea3` (M3 부모): PASS | `65c24753a`: FAIL (`F-noid_row_4`) | 회귀, M3 에서 생김. GREEN 은 출력을 바꾸지 않았다. 결정 대기(위 결정 1). |
| 린트 `internal/cli` | M3 시점 0 issues (진행 기록 §5) | 0 issues | `t1620-m3-green-lint.txt`. |

위 표는 기록 원문 그대로다. 그 뒤 `TestWSR006_ReviewGateRootMatrix` 행의 "결정 대기"는 WSR006 소절의 기대값 변경(`4f94abf73`)으로 끝났다.

- 후속 카드 매핑 (리더 지정, 이 SPEC 의 Follow-up 항목 그대로): F1 → t1662, F2 → t1637, F3 → t1639, F4 → t1683, F5 → t1674, F6 → t1635. F4 는 t1683 이 doctor 경고 추가만 맡고, 규칙 문장 정정은 이 M7 에서 수행했다. 큐 등록은 리더(운영자) 몫이라 레인은 등록하지 않았다. 이 매핑은 진행 기록에만 두고, 배포되는 규칙 파일에는 넣지 않았다.

### 리더 판정과 해결 기록

- 정책 건너뜀 이름 (6. Gap 과 잔여 위험 (a)): 리더는 (b) 안을 채택했다. 정책 건너뜀 두 경로(`codex_review_gate.go:161–169`, stale 바이너리와 tree_scope)에 systemMessage 만 더한다. 이름은 `stale 바이너리 건너뜀`, `tree_scope 건너뜀`. 허용 결정은 그대로다. 순서는 RED → GREEN 이다. 해결: `65c24753a`.
- 기록 실패 이름 (6. Gap 과 잔여 위험 (b)): 같은 M3 후속에서 `codex_review_gate.go:279–284` 의 읽기·쓰기 실패 경로를 `차단 횟수 기록 실패` 로 바꾼다. C 도달 상한 경로는 `차단 횟수 소진` 을 유지한다. 현재 코드는 두 경로 모두 `차단 횟수 소진` 을 냈다(`:279–280`, `:284`). 해결: `65c24753a`.
- REQ-RG-006 본문: 고치지 않았다. "검토 건너뜀의 결정은 바뀌지 않는다"는 모든 건너뜀 경로에 걸리는 일반 규칙이라, 두 경로를 열거하지 않아도 HISTORY 인용은 맞다. 판단 주체는 리더로 기록했다.
- 결정 1 (WSR006 기대값, 다른 SPEC 의 테스트): 이름 붙은 메시지로 바꾼다. 근거와 검증은 WSR006 소절, 커밋은 `4f94abf73` 이다.
- 결정 2 (tree_scope 이름): `tree_scope 건너뜀` 이름은 바꾸지 않고 유지한다. 잔여 위험으로 기록한다. 기본 설정(primary_scope=skip)의 primary checkout 건너뜀도 같은 이름으로 표시된다. `treeScopeSkipApplies`(`codex_review_tree_scope.go:132–153`)가 primary 갈래와 tree_scope 갈래 모두에서 true 를 돌려주기 때문이다. tree_scope 를 설정하지 않은 운영자가 혼동할 수 있다. 리더가 t1674 본문에 "두 갈래 분리 또는 사유 표시"를 추가했다.

### Run-phase Gap

1. **기준선 실패 3개 (이 카드의 변경과 무관).** 기준선은 `f4b38e2a3` (M2 이전, 편집 전 추출본) 이다. 명령: `MOAI_SKIP_LIVE_CODEX=1 go -C <scratch>/base-f4b38e2a3 test -count=1 -v -skip LiveCodex -run '^(TestReviewGate_|TestReviewEntry_|TestProducerFailReceipt|TestProduceCodexReviewReceipt_|TestKickInFlight_|TestRunCodexReviewRPC_|TestHasReviewableChanges_|TestCodexReviewGate|TestCodexReviewEntry_|TestHandleCodexReviewGate_|TestCodexReviewScope_|TestCardMergeBaseConfiguredMain|TestReviewableFromPorcelain)' ./internal/cli/`. 결과: 최상위 65개 중 62 PASS, 3 FAIL, `FAIL	github.com/modu-ai/moai-adk/internal/cli	46.848s`.
   - `--- FAIL: TestCodexReviewGatePrimaryCheckoutSkip (0.72s)` — `codex_review_gate_primary_scope_red_test.go:84: the primary-skip row must carry the resolved tree, got dir "/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCodexReviewGatePrimaryCheckoutSkip118067050/001" want "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCodexReviewGatePrimaryCheckoutSkip118067050/001"`
   - `--- FAIL: TestCodexReviewGate_TreeScopeRequestShapeUnchanged (0.74s)` — `codex_review_scope_test.go:87: thread/start cwd = "/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCodexReviewGate_TreeScopeRequestShapeUnchanged2101977025/001", want the resolved tree "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCodexReviewGate_TreeScopeRequestShapeUnchanged2101977025/001"`
   - `--- FAIL: TestCodexReviewGate_StaleEnvKeepsTreeScope (1.14s)` — `codex_review_scope_test.go:280: tree-scope cwd = "/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCodexReviewGate_StaleEnvKeepsTreeScope403953538/001", want the resolved tree "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestCodexReviewGate_StaleEnvKeepsTreeScope403953538/001"`
   - 원인은 macOS 임시 디렉터리 심볼릭 링크(`/var` → `/private/var`)다. 같은 세 테스트는 M3 이전 커밋 `988230ea3` 의 추출본(`FAIL	github.com/modu-ai/moai-adk/internal/cli	3.482s`)과 M3 직후 실행(`FAIL	github.com/modu-ai/moai-adk/internal/cli	54.829s`)에서도 FAIL 이었다. 이 카드는 고치지 않는다.
2. 카드 워크트리에서 `internal/cli` 패키지 전체는 실행하지 않았다. 구조적 적색이며, 착지 전 CI 가 게이트다(리더 확인 대상).
3. 전체 `/verify` 게이트는 실행하지 않았다(리더 수락 Gap).
4. 라이브 codex 테스트는 실행하지 않았다. 컴파일 확인만 했다(리더 지시). 게이트 계열 실행의 명령에는 `-skip 'LiveCodex'` 가 붙어 있다.
5. 과정 결함: M3 검증은 `TestWSR006_ReviewGateRootMatrix` 를 돌리지 않았다. M3 검증의 `-run` 패턴(`^(TestReviewGate_|TestReviewEntry_|TestProduceCodexReviewReceipt_|TestProducerFailReceipt|TestHandleCodexReviewGate|TestCodexReviewGate)`)이 이 테스트를 포함하지 않았다. 회귀는 M3 커밋 `1002a14c1` 에서 생겼고, M3 후속 GREEN 점검에서 발견됐다. 기대값을 바꾼 커밋 `4f94abf73` 이 처리했다.

### SF-1 수리 (sync-audit FAIL 뒤)

sync 감사의 SF-1(REQ-RG-004, 차단 횟수)을 고쳤다. 키당 차단 횟수의 확인(읽기와 상한 비교)과 기록(+1)을 키별 잠금 안의 한 단계로 묶었다. 잠금 대기는 1초로 상한을 두며, 상한 안에 잡지 못하면 `차단 횟수 기록 실패`로 허용한다. 허용 결정은 바꾸지 않았고(REQ-RG-006), 새 상태 이름은 없다.

- 시작 상태: `git rev-parse --short HEAD` 는 `3c2f5dfa2`, `git branch --show-current` 는 `WT-review-gate-fix`. 일치했다.
- RED 커밋 `4bd5d433a`: `internal/cli/codex_review_gate_concurrency_test.go` 하나(테스트 파일만).
- GREEN 커밋 `a13e02705`: `internal/cli/codex_review_gate.go` 하나(프로덕션 파일만).
- 잠금 헬퍼: 같은 패키지의 기존 원시 잠금 `acquireGateLockImpl`(`gate_lock_unix.go` 는 flock, `gate_lock_windows.go` 는 O_EXCL)을 재사용했다. 대기 상한과 죽은 소유자 파기는 `codex_review_gate.go` 에 새로 두었다. 재사용하지 않은 후보와 이유는 다음과 같다. `verify.acquireKeyLock` 은 다른 패키지의 비공개 함수이고 상한 뒤 잠금 없이 진행하는 fail-open 이라 요구와 반대다. `waitForGateLock` 과 `AcquireGateLock` 은 게이트 실행 산출물에 고정되어 있고 상한 뒤 무잠금으로 실행한다. `ClearStaleGateLock` 은 프로젝트 경로에 고정되어 있다. `lockfile.Lock` 은 상한이 없는 블로킹 잠금이다.

**RED** — 명령 `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -skip LiveCodex -run '^TestReviewGate_ConcurrentBlockCap$' ./internal/cli/`, 종료 코드 **1**. 수정 전 트리에서 관측한 원문:

```
--- FAIL: TestReviewGate_ConcurrentBlockCap (1.56s)
    codex_review_gate_concurrency_test.go:73: concurrent turn ends for one fail key blocked more than the cap (C = 1) in 73 of 100 rounds; 8 simultaneous turn ends per round, max blocks in one round = 8
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.066s
FAIL
```

**GREEN** — 명령 `MOAI_SKIP_LIVE_CODEX=1 go test -count=3 -skip LiveCodex -run '^TestReviewGate_ConcurrentBlockCap$' ./internal/cli/`, 종료 코드 **0**, 원문 `ok  	github.com/modu-ai/moai-adk/internal/cli	14.749s`. 커밋 `a13e02705` 과 같은 내용에서 실행했다. 주석 한 곳을 고치기 전에 같은 명령을 돌린 실행도 `ok  	github.com/modu-ai/moai-adk/internal/cli	14.426s`, 종료 코드 0 이었다.

**회귀 family** — HEAD `a13e02705`, 명령 형식 `MOAI_SKIP_LIVE_CODEX=1 go test -count=1 -v -skip LiveCodex -run '<패턴>' ./internal/cli/`. 패턴은 `.moai/reports/t1620/progress.md` 의 01:20Z 블록 "회귀 확인 목록" (a)–(d) 를 그대로 옮겼다. (d) 는 목록의 이름 나열에서 정규식을 만들었다.

- (a) 게이트 family: 최상위 PASS 80, FAIL 3, 종료 코드 1, 패키지 줄 `FAIL	github.com/modu-ai/moai-adk/internal/cli	103.727s`. FAIL 3개는 기준선과 같은 macOS 심볼릭 링크 실패(위 Run-phase Gap 1)다: `TestCodexReviewGatePrimaryCheckoutSkip`, `TestCodexReviewGate_TreeScopeRequestShapeUnchanged`, `TestCodexReviewGate_StaleEnvKeepsTreeScope`. 새 실패는 없다. 신규 `TestReviewGate_ConcurrentBlockCap` 은 PASS (4.68s)다. M3 기록의 집계 82(PASS 79, FAIL 3)보다 1 늘었고, 신규 테스트 한 개다.
- (b) WSR `^TestWSR`: 최상위 PASS 13, FAIL 0, 종료 코드 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	26.212s`. `TestWSR006_ReviewGateRootMatrix` 는 PASS 다.
- (c) stop 예외 경로 `^(TestStdinFailClosed_StopIgnoresStopHookActive|TestStdinFailClosed_CodexStopExempt)$`: PASS 2, FAIL 0, 종료 코드 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	1.357s`.
- (d) stop-chain·tree-scope `^(TestStaleBinarySkip_|TestTreeScope|TestPrimaryScopeReader_|TestCodexStopChain_TreeScopeSkip$|TestStopChainEffectParityGolden$)`: 최상위 PASS 19, FAIL 0, 종료 코드 0, `ok  	github.com/modu-ai/moai-adk/internal/cli	104.115s`. 기록된 이름 18개에 `TestTreeScopedSelfGateSkipsConfigOnlyTree` 가 더해졌다(`TestTreeScope*` 와일드카드에 걸린다). 기록의 `TestTreeScope_NonSkipValuesKeepTreeScope` 는 현재 소스에서 `TestTreeScope_NonSkipValuesKeepTreeRequest` 로 실행되었다. 이름 기록만 낡았고 결과에는 영향이 없다.

**정적 검사** — `gofmt -l internal/cli/codex_review_gate.go internal/cli/codex_review_gate_concurrency_test.go`: 출력 없음, 종료 0. `go build ./internal/cli/`: 종료 0. `GOOS=windows GOARCH=amd64 go build ./internal/cli/`: 종료 0(Windows 잠금 구현의 컴파일 확인). `golangci-lint run ./internal/cli/`(v2.1.6): `0 issues.`, 종료 0.

**커버리지(함수 단위)** — `-coverprofile` 로 차단 횟수 테스트 부분집합(`TestReviewGate_BlockCount*`, `TestReviewGate_ConcurrentBlockCap`, `TestReviewGate_FreshPassClearsBlockCount`, `TestReviewGate_BlockCapAllowsNextTurnWithMessage`, `TestReviewGate_AllowedAfterCapStaysAllowedAndCountKept`, `TestReviewGate_NonFreshStateNamedInMessage`)을 돌렸다. `codexReviewGateFail` 100.0%, `codexReviewClearBlocks` 100.0%, `codexReviewBlockLockPath` 100.0%, `codexReviewReserveBlock` 84.6%, `acquireCodexBlockLock` 86.7%, `codexReviewBreakStaleBlockLock` 12.5%. 패키지 전체 비율은 측정하지 않았다. 부분집합의 패키지 비율 6.0% 는 의미가 없다.

**Gaps (미검증)**

1. `internal/cli` 패키지 전체 스위트는 실행하지 않았다(레인 지시의 금지 명령).
2. 라이브 Codex 테스트는 실행하지 않았다. 모든 실행에 `-skip LiveCodex` 를 붙였다.
3. 전체 `/verify` 게이트는 실행하지 않았다.
4. 잠금 대기 상한 분기(`acquireCodexBlockLock` 의 시한 초과 반환)와 죽은 소유자 파기 본체(`codexReviewBreakStaleBlockLock`, 커버리지 12.5%)를 실행하는 테스트가 없다. 파기 본체는 `codexReviewBlockLockBreakable` 이 참일 때(Windows)만 동작하며, 이 머신에서 Windows 실행은 하지 않았고 크로스 컴파일만 확인했다. 후속 RED 후보는 구현 전에 쓴다: (i) 살아 있는 소유자가 쥔 잠금은 상한 안에 반환하며 `차단 횟수 기록 실패`로 허용하고, 차단 횟수 기록을 남기지 않는다. (ii) 죽은 PID 를 적은 산출물은 파기된 뒤 차단으로 이어진다. (iii) 살아 있는 PID 를 적은 산출물은 유지된다. (iv) 소유자 없이 오래된 빈 산출물은 파기된다. 이 카드는 엄격한 RED→GREEN 순서를 지키려고 구현 뒤에 테스트를 넣지 않았다.

**잔여 위험 (Residual-risk)**

- Windows 죽은 소유자 파기: 재읽기는 파기 창을 좁힐 뿐 닫지 못한다(`ClearStaleGateLock` 과 같은 잔여). 점검과 파기 사이에 잠금이 다른 프로세스로 넘어가면 드물게 살아 있는 잠금을 지울 수 있다.
- PID 재사용: 죽은 소유자의 PID 가 살아 있는 다른 프로세스에 쓰이면 산출물은 살아 있는 것으로 읽힌다. 그 키는 상한 뒤에도 `차단 횟수 기록 실패` 허용으로 남는다. 차단 횟수는 손상되지 않고 fail-open 이다. "결국 파기된다"는 PID 가 풀릴 때까지 성립하지 않는다.
- 시한 초과: 소유자가 살아 있으나 1초를 넘겨 멈추면, 경합하는 턴 종료는 차단 대신 허용한다(fail-open).
- 차단 횟수 정리: 잠금을 시한 안에 잡지 못하면 기록을 잠금 없이 지운다. 같은 키에서 차단과 정리가 겹치면 횟수가 되살아날 수 있다. 같은 키의 FAIL→PASS 전이가 전제라서 드물다.
- Unix 잠금 산출물(`blocked-<digest>.json.lock`)은 파기되지 않고 남는다. flock 산출물을 지우면 대기자와 경쟁이 생기기 때문이다. 키당 한 개의 무해한 빈 파일이다.

**커밋** — 이 항목의 커밋 제목: `docs(SPEC-REVIEW-GATE-REPAIR-001): record the SF-1 repair evidence in §E.2 (card t1620)`.

## §E.3 Run-phase Audit-Ready Signal

audit_ready: true
run_status: audit-ready
run_complete_at: 2026-10-11T01:21:03Z

run-phase 의 M1, M2, M3(후속 포함), M7 은 로컬 커밋으로 끝났고, sync 는 시작하지 않았다.

## §E.4 Sync-phase Audit-Ready Signal

audit_ready: true
sync_status: audit-ready
sync_complete_at: 2026-10-11T03:07:23Z
sync_commit_sha: c861336ee64bbb81f304eec02bea7a9f97dd5af2

SF-1 수리(§E.2)와 SF-2 수정 증서(커밋 `582bed1f8`, spec.md `## Amendments`) 뒤의 재종결이다. 수정 증서가 SPEC 을 in-progress 로 되돌렸고, 이 커밋이 completed 로 다시 닫는다. 이전 종결의 sync_commit_sha `7716c99ff2347dc004ee0e209791fc648772e1ef` 는 spec.md 의 prior_completed_sha 와 같은 값이다. 커밋은 자신의 해시를 기록할 수 없으므로 이 커밋의 sync_commit_sha 는 pending-backfill 자리표시다. 실제 SHA 는 다음 커밋에서 backfill 한다(spec-workflow.md § Sync (close), spec-frontmatter-schema.md § SHA placeholder backfill exemption).

**§E.4 Gaps (델타 감사로 넘김)**

- (i) 잠금 시한 초과 분기와 Windows 죽은 소유자 파기에는 테스트가 없다. 파기 함수 `codexReviewBreakStaleBlockLock` 의 함수 커버리지는 12.5%다(§E.2 SF-1 Gap 4 와 같은 수치).
- (ii) 래퍼의 35·36·79–80행 종료는 SPEC 에 명명되어 있고 그 변경은 t1674 몫이므로, 이 SPEC 은 래퍼를 바꾸지 않는다.
- (iii) SF-3 은 영구 편차다. §G 에 기록했다.

## §G Override and Refusal Record

- 2026-10-10T22:44:14Z SPEC-REVIEW-GATE-REPAIR-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 11 >= tier ceiling 3); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter10.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter11.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter8.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter9.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit.md
- 2026-10-10T23:24:45Z SPEC-REVIEW-GATE-REPAIR-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 11 >= tier ceiling 3); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter10.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter11.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter8.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit-iter9.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1620/.moai/reports/t1620/plan-audit.md

- 2026-10-10T23:27:18Z SPEC-REVIEW-GATE-REPAIR-001 operator-decision kickoff=approve decider=human actor=operator engine-event=seq-1250 (card.transition kickoff->assigned v57) source=AskUserQuestion(영실이 세션; 리더 전달, 레인은 AskUserQuestion 자체를 보지 않음) board=d-20261010T232653Z-a596 codex-leg=advisory(2413·a280) lane=transcribed-from-engine-event-not-authored
- 귀속 표 (영실이 판단, 리더 전달, 2026-10-10): F1 → t1662 (relate 142) · F2 → t1637 (relate 143) · F3 → t1639 (relate 145; 본문 추가 todo_runtime.go:364, todo_dispatch_current.go:53 은 기존) · F4 → t1683 (doctor 경고만 맡고, 규칙 문장 정정은 범위 밖이다. 정정은 M7 에서 수행) · F5 → t1674 (relate 146; 본문 "plan 시작 시 반영") · F6 → t1635 (relate 144).
- 2026-10-11 · **SF-3 영구 편차 (permanent deviation).** sync 커밋 `7716c99ff` 의 제목(`docs(SPEC-REVIEW-GATE-REPAIR-001): close the sync phase and record the audit-ready signal (card t1620)`)에는 닫기 중위(close infix)가 없다. sync-audit 은 이 제목을 `run-partial` / `in-progress` 로 분류했다(`.moai/reports/t1620/sync-audit.md` E-7 관측). 이력은 다시 쓸 수 없으므로 이 편차는 영구다. 이 기록은 보드 `d-20261011T023045Z-326e` 의 리더 판정 (d) 를 따르며, 랜딩 시점의 드리프트 처리는 리더의 판정이다.
- 2026-10-11 · **sync 감사 FAIL 기록.** 파일 `.moai/reports/t1620/sync-audit.md` (audited_sha 3c2f5dfa2, 판정 FAIL) 는 이 커밋에서 편집하지 않았다. SF-1 은 커밋 `4bd5d433a`(RED), `a13e02705`(GREEN), `e62a8bdd5`(§E.2 증거)로 수리했다. SF-2 는 커밋 `582bed1f8` 이 SPEC 본문(REQ-RG-007 범위 한계, Amendments 절)에 래퍼 네 종료 경로를 명명했고, 래퍼 동작은 바뀌지 않았다(t1674 몫). G1 감사 레그(필수 교차 모델, codex)는 델타 감사에서 실행하며, 그때 2413·a280 advisory 를 적용한다.
