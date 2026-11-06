auditor-model: claude-opus-5-5

# t2116 sync-phase 감사 판정

verdict: PASS
audited_sha: 5c9e2cb7b780c64075c057525b73edacf074b53f

- 카드: t2116 (P0, Class B, SPEC 없음), 레인 lane-21, 트리 `.moai/worktrees/projects/moai-adk-go/t2116`, 브랜치 `WT-plan-run-sync`
- 판정: **PASS**, 종합 점수 **0.92** (가중 92.25/100)
- 감사 시작 시 `git rev-parse HEAD` = `5c9e2cb7b780c64075c057525b73edacf074b53f`, `git status --short` = 빈 출력. 감사 종료 직전 다시 읽어도 같은 값이었다(이 판정 파일을 쓰기 전).
- 카드 고유 범위: `git log 27adc925ef..HEAD --no-merges` 중 t2116 커밋 2개 — RED `1cacc0634b`, GREEN `66ec369381`.

## 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 95 | PASS | 집중 시험 homestate 22/22, cli 32/32 통과. 스테이지 6개 전부 재임대·스테이지 보존·재개 확인. 아래 증거 E1·E2 |
| Security (25%) | 95 | PASS | 소유자=보유자 동시 일치, T3가 등록 작업자·소유자를 다시 강제, 실패 시 전체 롤백. 변이 m2(소유자 검사 제거)는 기존 시험이 잡음(E4) |
| Craft (20%) | 88 | PASS | 변이 4종 전부 사멸, gofmt·go vet 깨끗. 다만 MCP 경로·`factory stage` 동사 재개는 스테이지 상태로 시험하지 않음(F4) |
| Consistency (15%) | 88 | PASS | 기존 술어·전이 재사용, 코드 변경 2줄. 규칙 문서 한 문장이 낡음(F1, t2082 소관) |

필수 통과 차원(Functionality, Security) 모두 통과. 차단(blocking) 결함 없음.

## 결함 목록

- **F1** [MINOR] [optional] [touched-file 아님 — outside] `.claude/rules/moai/workflow/factory-dispatch-mechanics.md:30` 및 템플릿 사본 `internal/template/templates/.claude/rules/moai/workflow/factory-dispatch-mechanics.md:30` — "For a card that is leased with an expired lease, held by its recorded owner, that one call applies the expiry and leases the card to the owner again…"는 이제 `leased`만이 아니라 plan … merge-ready 상태까지 적용되므로 범위가 좁게 쓰여 있다. 같은 파일 44행("A lease on a merge-ready … the lane re-leases it with `factory next --card`, the stage preserved")은 이 변경으로 비로소 참이 된다(수정 전에는 merge-ready 소유자 재임대가 거부되었음). 필요한 조치: t2082의 재작성에서 30행을 "a card leased or parked in a staged state (plan … merge-ready) with an expired lease"로 넓히고 `merging` 제외를 한 구절로 명시. 이 카드에서 고칠 필요 없음.
- **F2** [MINOR] [optional] [outside] `internal/cli/factory_card.go:2112` → `:2199` — Codex 레인(`MOAI_FACTORY_BACKEND=gpt`)이 만료된 merge-ready 카드를 `next --card`로 지명하면, 수정 전에는 keep-set 단계에서 `refused owned: … in flight` 로 거부되었고, 수정 후에는 재임대 판정으로 `assignedHere`가 켜져 keep-set을 통과한 뒤 `backend-skip`("a Codex lane cannot advance a card recorded at merge-ready")으로 거부된다. 두 경우 모두 클레임 전에 거부되어 기록을 쓰지 않으므로 동작상 안전하며, 거부 사유 토큰만 바뀐다. 이 경로를 시험하는 사례는 없다(코드 읽기로 확인, 실행 증거 없음 — 신뢰도 중간). 필요한 조치: 원하면 시험 한 개 추가. 판정에 영향 없음.
- **F3** [MINOR] [optional] [in-diff] `internal/homestate/card_lease_expired_owner.go:19` — `IsLeaseHoldingState(card.State)` 절은 `Card.LeaseExpired`(`internal/homestate/card_record.go:148`)가 이미 같은 조건을 걸기 때문에 중복이다. 그 결과 "모든 상태로 넓히는" 변이(m1)는 "merging 재허용" 변이(m3)와 동치가 된다(E4). 명시성 측면에서 남겨도 무방. 필요한 조치: 없음(가독성 선택).
- **F4** [MINOR] [optional] [in-diff, 시험 공백] 새 cli 시험은 CLI 동사 경로만 다룬다. MCP `factory_next` 경로는 `leased` 상태로만 시험되고(`TestT2013OwnerNextReleasesExpiredLeaseViaMCP`), 스테이지 상태로는 시험되지 않는다. CLI 재임대 뒤 `moai factory stage` 동사로 스테이지를 재개하는 흐름도 시험이 없다(homestate 시험이 전이 간선 T4a~T4f를 직접 구동). progress.md 잔여 위험 항목이 이 공백을 정직하게 적고 있다. 구현이 같은 함수를 공유하므로 위험은 낮다. 필요한 조치: 후속 카드에서 시험 추가 가능.
- **F5** [INFO] [outside] 흡수된 t2091 팁과 함께 `.moai/reports/t2071/sync-audit.md`, `.moai/reports/t2091/sync-audit.md`가 커밋되어 들어와 있다(`git diff --stat b00c35d103...HEAD`). 감사 산출물 규약상 판정 파일은 gitignore 대상이며 커밋하지 않는다. t2071·t2091 소관이므로 재감사하지 않음. 리더 확인용으로만 기록.

## 점검 항목별 결과

### A. 술어의 정확성
- `leaseHoldingStates` = `CardLeased, CardPlan, CardPlanAudit, CardRun, CardSync, CardSyncAudit, CardMergeReady, CardMerging` (`internal/homestate/card_record.go:66-68`). 새 술어는 이 중 `merging`만 뺀다 → 재임대 가능 = leased + 스테이지 6개. 카드 요구사항과 정확히 일치.
- `kickoff`(`card_record.go:23`)는 리스 보유 상태가 아니므로 `LeaseExpired`가 false를 돌려주고 술어도 false. 영향 없음.
- `applyLeaseExpiry`(`card_transition.go:422-439`)는 `next := cur`로 복사한 뒤 state(→assigned, merging만 blocked), LeaseHolder, LeaseExpiresAt, Version, UpdatedAt만 바꾼다. Stage·WorktreePath·EvidenceSHA·EvidencePath·MergeSHA는 보존된다. merging은 술어에서 제외되므로 재임대 트랜잭션은 항상 assigned 경로를 탄다.
- T3(`card_transition.go:136`, `assigned → leased`, `guardLeaseAcquire`, 582-603행): 라벨이 비지 않고 workers 테이블에 등록되어 있으며 `label == cur.OwnerLabel`이어야 한다. 스테이지 6개 모두에서 assigned → leased는 유효.
- 재개: T4a~T4f(`card_transition.go:137-142`) + `guardStageResume`(608-611행, `stageAdmits(cur.Stage, req.To)`). Stage가 보존되므로 재개 가능. `TestReacquireExpiredLeaseStagedStates_T2116`이 6개 상태 모두 재개 전이를 실제 수행해 확인.

### B. 호출자
- 비시험 호출자는 두 곳뿐: `internal/cli/factory_card.go:1717`(`factoryNextReacquire` → `ReacquireExpiredLease`)과 `:2112`(지명 판정). MCP `factory_next`는 같은 구현을 쓴다.
- 2112행 분기는 재임대 대상 행을 `assigned`로 바꾼 사본(`refusalRow`)으로 keep-set·직렬 슬롯 판정을 한다. 원래 상태가 `leased`였다고 가정하는 코드는 없다. `factorySerialInFlightExcluding`은 지명 카드 자신을 제외하므로 스테이지 상태였던 카드가 자기 슬롯을 막지 않는다.
- `RefreshDispatchCurrentLockHeld`는 재임대된 카드(`leased`)에 대해 T2/T3와 똑같이 호출된다.
- 여전히 거부되어야 하는 경우: 리더·다른 레인(보유자/소유자 불일치), 미만료 리스, `needs-decision`·`blocked`(리스 보유 상태가 아님 → `LeaseExpired` false), `merging`(명시 제외). 모두 술어 false → 기존 거부 유지. 거부 토큰 변화는 F2 한 건.

### C. 보안·소유권
- 술어가 소유자=보유자=라벨을 요구하고, T3가 등록 작업자+소유자를 트랜잭션 안에서 다시 강제한다. 술어 단계의 소유자 검사를 빼도(m2) T3에서 `ErrLeaseHolder`로 실패하고, 실패 시 `withCardTx`가 만료 반환까지 전부 롤백한다(`TestReacquireExpiredLeaseRollsBackWholly` 통과).
- 리더나 다른 레인이 이 경로로 카드를 가져갈 길은 없다.

### D. 시험 — 변이 탐침
카드 파일을 건드리지 않고 `go test -overlay`로 술어 파일만 바꿔 넣었다(변이 원본은 세션 스크래치패드).

| 변이 | 내용 | 결과 |
|---|---|---|
| m1 | 상태 절 전부 제거(모든 상태 허용) | 사멸 — `LeaseExpired`의 상태 게이트 때문에 실질적으로 m3와 동치. 3개 시험 FAIL |
| m2 | 소유자 검사 제거 | 사멸 — `holder_but_not_owner` FAIL (기존 시험) |
| m3 | merging 재허용 | 사멸 — 3개 시험 FAIL |
| m4 | `leased`만 허용(수정 전 동작) | 사멸 — 술어 시험 + 6개 스테이지 트랜잭션 시험 FAIL |

### E. 문서 정합성
F1 참조. 카드 고유 파일과 모순되는 문서는 없다.

### F. progress.md의 정직성
- RED 기록(`red-homestate.txt`, `red-cli.txt`)은 의도된 단언에서 실패했고 실패 문구가 사고 문구("refused owned: the card is in flight at sync (holder lane-1)")와 같다 — 확인함.
- GREEN cli 기록 `green-cli-narrow.txt` 끝줄 `ok … 126.257s` — 확인함. 병합 HEAD에서 cli 집중 시험을 다시 돌리지 않았다는 공백을 스스로 적어 둠 → 이번 감사가 병합 HEAD에서 cli 집중 시험 32/32를 다시 실행해 메움.
- 전체 패키지 시간초과를 통과로 주장하지 않음. 과장 없음.

## 증거 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim
1. 소유자가 스테이지 상태(plan, plan-audit, run, sync, sync-audit, merge-ready)의 만료 리스를 `next --card`로 되찾고, 스테이지·워크트리·증거가 보존된다.
2. 다른 레인, 미만료 리스, merging은 여전히 거부되고 아무것도 쓰지 않는다.
3. 새 시험은 수정 전 동작과 주요 변이를 잡아낸다.

### Evidence
- **E1** `go test ./internal/homestate/ -v -run 'T2116|Reacquire|IsReacquirable' -count=1` → 출력 `.moai/reports/t2116/audit/homestate.txt`, exit 0. `=== RUN` 22개, `--- PASS` 22개, `--- FAIL` 0개. 최상위:
  ```
  --- PASS: TestIsReacquirableExpiredLeaseCoversStagedStates_T2116 (0.00s)
  --- PASS: TestReacquireExpiredLeaseStagedStates_T2116 (41.89s)
  --- PASS: TestReacquireExpiredLeaseStagedIgnoresOtherShapes_T2116 (35.42s)
  --- PASS: TestReacquireExpiredLeaseIsOneTransaction (4.29s)
  --- PASS: TestReacquireExpiredLeaseRollsBackWholly (5.41s)
  --- PASS: TestReacquireExpiredLeaseIgnoresEveryOtherShape (26.98s)
  ok  	github.com/modu-ai/moai-adk/internal/homestate	114.725s
  ```
- **E2** `go test ./internal/cli/ -v -run 'TestT2116|TestT2013Owner|TestT2013Other|TestT2013Second|TestT2013Renew|TestT2013Adopt' -count=1 -timeout 270s` → 출력 `.moai/reports/t2116/audit/cli.txt`, exit 0. `=== RUN` 32개(최상위 14 + 하위 18), `--- PASS` 32개, FAIL·SKIP 0개. 최상위:
  ```
  --- PASS: TestT2013OtherLaneStillRefusedAfterRejoin (3.97s)
  --- PASS: TestT2013OwnerNextReleasesExpiredLeaseViaVerb (3.05s)
  --- PASS: TestT2013OwnerNextReleasesExpiredLeaseViaMCP (10.17s)
  --- PASS: TestT2013RenewStillRefusesExpiredWithoutWrite (3.94s)
  --- PASS: TestT2013OtherLaneNextRefusedOnExpiredLease (3.17s)
  --- PASS: TestT2013OwnerNextRefusedBeforeExpiry (3.17s)
  --- PASS: TestT2013OwnerNextLeavesWorktreeUntouched (5.13s)
  --- PASS: TestT2013AdoptTakesNoPartInReacquire (1.84s)
  --- PASS: TestT2013SecondOwnerNextAfterReacquireIsRefused (1.99s)
  --- PASS: TestT2116OwnerNextReleasesExpiredStagedLease (22.20s)
  --- PASS: TestT2116OtherLaneNextRefusedOnExpiredStagedLease (23.62s)
  --- PASS: TestT2116OwnerNextRefusedBeforeExpiryStaged (24.99s)
  --- PASS: TestT2116OwnerNextRefusedOnExpiredMerging (3.38s)
  --- PASS: TestT2116OwnerNextLeavesStagedWorktreeUntouched (3.40s)
  ok  	github.com/modu-ai/moai-adk/internal/cli	117.306s
  ```
- **E3** `gofmt -l` (카드 파일 4개) → 빈 출력, exit 0. `go vet ./internal/homestate/` → 빈 출력, exit 0.
- **E4** 변이 탐침, `go test ./internal/homestate/ -overlay <scratchpad>/mut/mN.json -run 'T2116|IgnoresEveryOtherShape' -count=1` (출력 `.moai/reports/t2116/audit/mutant-m{1,2,3,4}.txt`), 모두 exit 1:
  ```
  m1: --- FAIL: TestIsReacquirableExpiredLeaseCoversStagedStates_T2116
      card_lease_expired_owner_staged_t2116_test.go:66: state merging is reacquirable, want not (...)
      --- FAIL: TestReacquireExpiredLeaseStagedIgnoresOtherShapes_T2116/interrupted_merge
      card_lease_expired_owner_staged_t2116_test.go:144: reacquire = (false, illegal card transition: blocked → leased), want (false, nil)
      --- FAIL: TestReacquireExpiredLeaseIgnoresEveryOtherShape/interrupted_merge
      FAIL	github.com/modu-ai/moai-adk/internal/homestate	207.244s
  m2: --- FAIL: TestReacquireExpiredLeaseIgnoresEveryOtherShape/holder_but_not_owner
      card_lease_expired_owner_test.go:143: reacquire = (false, card lease holder refused: "worker-1" is not the registered owner "worker-2" of card stale), want (false, nil)
      FAIL	github.com/modu-ai/moai-adk/internal/homestate	141.963s
  m3: --- FAIL: TestIsReacquirableExpiredLeaseCoversStagedStates_T2116
      --- FAIL: TestReacquireExpiredLeaseStagedIgnoresOtherShapes_T2116/interrupted_merge
      --- FAIL: TestReacquireExpiredLeaseIgnoresEveryOtherShape/interrupted_merge
      FAIL	github.com/modu-ai/moai-adk/internal/homestate	137.800s
  m4: --- FAIL: TestReacquireExpiredLeaseStagedStates_T2116/{plan,plan-audit,run,sync,sync-audit,merge-ready}
      card_lease_expired_owner_staged_t2116_test.go:83: reacquire of an expired sync card = (false, <nil>), want a re-lease
      FAIL	github.com/modu-ai/moai-adk/internal/homestate	327.126s
  ```
- **E5** 코드 읽기: `card_lease_expired_owner.go:17-21, 36-81`, `card_record.go:66-72, 84, 147-155`, `card_transition.go:136-148, 300-318, 422-439, 582-611`, `factory_card.go:1704-1724, 2085-2200, 2336-2357, 2410-2430`.

### Codex 감사 영수증
- 호출 1회(재호출 없음): `mcp__moai__codex_audit`, `project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2116`, `target=baseBranch`, `mode=native`, `delta_base` 없음.
- 결과: `verdict: pass`, findings 0건. 요약 "지정한 기준 커밋 대비 변경에서 지적할 결함은 찾지 못했어. 정적 검토만 했고 … 테스트·빌드·재현 명령은 실행하지 않았어."
- `audit_receipt: rcpt-49ed7c1e5fb157b9fbd18305`, backend `codex`, resolved_model `gpt-6.1-sol`, effort high, server build `646955327e`, review_base `main (merge base b00c35d103fa175efe3b68340d9b176cba1b067b)` → 검토 대상은 `b00c35d103...5c9e2cb7b7` 차분(카드 고유 4개 파일 + 흡수된 t2091/t2071 내용), HEAD = `5c9e2cb7b780c64075c057525b73edacf074b53f`.

### Baseline-attribution
- 모든 시험은 HEAD `5c9e2cb7b780c64075c057525b73edacf074b53f`, 깨끗한 작업 트리에서 이 감사 중 실행. `go test`는 트리 소스를 직접 컴파일하므로 설치된 `moai` 바이너리와 무관(§2.2 해당 없음). 변이는 `-overlay`로만 주입, 트리는 바뀌지 않음(감사 종료 시 `git status --short` 빈 출력 재확인).
- `moai verify audit-plan --project-root <toplevel>` → `config_status: ok`, `model: codex`, gates claude=off / codex=required / glm=off, `cross_model_active: true`, `enforced_required: ["codex"]`.

### Gaps
- **audit_multi 미호출**: 감사 계획은 `cross_model_active: true`라 정의상 `audit_multi` 경로를 가리키지만, 운영자 상시 결정(이 감사 지시문)에 따라 `codex_audit` 단일 호출만 했다. 필수 게이트는 codex 하나이고 그 영수증이 있으므로 게이트는 충족된 것으로 보지만, `audit_multi` 수렴 결과 다이제스트는 없다.
- `codex_audit` 결과의 provenance transport가 `codex app-server`로 표기됨 — 이는 MCP 서버 내부 전송이며 이 감사는 app-server를 직접 쓰지 않았다(MCP 도구만 사용).
- **도구 거부 2건(§3.1)**: (1) 변수로 경로를 넘긴 `sed` 변이 생성 명령, (2) `for` 루프로 변이 4개를 돌리는 명령이 워크트리 격리 가드에 거부됨. 대신 변이 파일을 스크래치패드에 복사·편집하고 명령을 리터럴 경로로 하나씩 실행했다. 증거 성격은 바뀌지 않았다(실행 결과 그대로).
- 전체 패키지(`internal/homestate`, `internal/cli`)와 `golangci-lint`는 실행하지 않음(기계 부하 15분 평균 120~155, 지시상 금지). 린트는 gofmt + `go vet ./internal/homestate/`만. CI 판정이 정본.
- MCP 경로의 스테이지 상태 시험, CLI 재임대 뒤 `factory stage` 동사 재개 시험 없음(F4).
- F2의 Codex 레인 거부 토큰 변화는 코드 읽기로만 추론, 실행 관측 없음.
- 실제 만료된 스테이지 카드를 새 바이너리로 재임대해 본 현장 실행 없음. Windows·다중 OS 미검증.

### Residual-risk
- 만료값이 파싱 불가한 스테이지 카드도 `LeaseExpired`가 true를 돌려 소유자가 재임대할 수 있다(`card_record.go:152-154`). 수정 전 `leased`에도 같았던 기존 동작이며 소유자 한정이라 위험 낮음.
- 재임대된 카드는 `leased` 상태로 돌아오므로, 레인이 스테이지를 재개하지 않고 방치하면 다른 판정 경로(예: 스테이지 기반 화면)가 `leased`로 읽는다. 기존 `leased` 재임대와 동일한 성격.
- 변이 m1이 m3와 동치라서 "상태 게이트를 `LeaseExpired`에서까지 빼는" 변이는 이 카드 시험으로 따로 측정되지 않았다(그 함수는 카드 범위 밖).
- 집중 시험 원문은 `.moai/reports/t2116/audit/`(gitignore 대상)에만 있다. 판정을 가른 줄은 이 파일에 옮겨 적었다.

## 반복 이력
- 1회차(이 감사): PASS 0.92 @5c9e2cb7b7, 영수증 rcpt-49ed7c1e5fb157b9fbd18305.

AUDIT-VERDICT: PASS spec=t2116 receipts=rcpt-49ed7c1e5fb157b9fbd18305
