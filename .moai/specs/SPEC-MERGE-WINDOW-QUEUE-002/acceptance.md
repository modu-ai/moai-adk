# SPEC-MERGE-WINDOW-QUEUE-002 — 인수 기준

> 각 AC는 판정 명령의 관측 출력으로 PASS를 결정한다. RED-now 셀은 plan 단계 트리 `f7606c7bc`에서 실측한 실패 관측이고, green-path 셀은 어느 마일스톤이 무엇으로 뒤집는지를 명명한다. §D AC Matrix — 8 AC / 10 REQ.
>
> **스크럽 변수 전체 목록(모든 RED/green 판정 명령 공통 — 3종)**: `MOAI_KANBAN_ID` · `MOAI_KANBAN_LEAD_ADDR` · `MOAI_KANBAN_SETTINGS_INJECTED` — 레인 세션 env가 테스트의 env 민감 경로를 오염하지 않게 AGENTS.md §4의 env-scrub **복합 형태**(`unset <3변수> && <명령>` — 한 호출 안에서 결합, zsh/bash 공통 유효)로 실행한다. 문서 셀의 `unset ...` 축약 표기는 그대로 실행하면 `zsh: invalid parameter name`이 난다(D15).

## §D AC Matrix

### AC-MWQ2-001 (maps REQ-MWQ2-001) — 검증기가 형식 불량 식별자를 거부 (R7-3 단위)

- **Given** tree/Base 중 하나가 40-hex SHA가 아닌 레코드 (예: `Base: "bad"`)
- **When** `ValidateRemeasureRecord`가 그 레코드를 검사하면
- **Then** record-invalid 오류를 반환한다 (메시지가 식별자 형식을 이름으로 밝힌다)
- **RED-now**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582VerifierAdmitsAShortBaseSHA$' -v ./internal/factory/` → exit 1, stdout 원문 — §D.3 장부 **E-LEDGER-001** 인용 (트리 `f7606c7bc` 실측)
- **green-path**: M1이 (i) 검증기에 40-hex 형식 검증을 추가해 뒤집는다 → 위 명령 exit 0.
- **판정**: 위 명령의 관측 종료 코드 0 + `ok` 행.

### AC-MWQ2-002 (maps REQ-MWQ2-002, REQ-MWQ2-008) — 짧은 base가 panic으로 창을 점유하지 않는다 (R7-3 e2e)

- **Given** `Base: "bad"` 레코드가 쓰여 있고 sess-b가 창을 쥐며 sess-c가 대기열에 있는 fixture
- **When** `RunMergeStep`이 실행되면
- **Then** panic 없이 cause 1(`MergeExitRecordInvalid`)로 거부하고, 창을 release하며, sess-c를 승격한다
- **RED-now**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease$' -v ./internal/factory/` → exit 1, stdout 원문 — §D.3 장부 **E-LEDGER-002** 인용 (트리 `f7606c7bc` 실측)
- **green-path**: M1이 (ii) 모든 SHA 접두 렌더를 길이 안전 접두로 바꿔 뒤집는다 → 위 명령 exit 0.
- **판정**: 위 명령의 관측 종료 코드 0 + `ok` 행; 회귀 방어로 `grep -n '\[:12\]' internal/factory/integration_merge_step.go`가 `minStrLen` 경유(또는 동등 방어) 외의 낫 슬라이스를 남기지 않음을 M5 스윕이 기록.

### AC-MWQ2-003 (maps REQ-MWQ2-003, REQ-MWQ2-004, REQ-MWQ2-009) — no-test 마커가 혼합 스윕을 거짓 거부하지 않고 fail 이벤트 거부와 빈 스윕 거부는 유지된다 (R7-2 + ① 계약 승격)

- **Given** 실제 패키지의 per-test pass 1건 + 무테스트 패키지의 output 이벤트(`[no test files]`) + 패키지 수준 skip을 담은 `go test -json` 스트림
- **When** `countGoTestJSONTests`가 그 스트림을 세면
- **Then** 오류 없이 count=1, structured=true를 반환한다 — 빈 스윕 거부는 총 per-test pass가 0일 때만 발동한다
- **RED-now**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582MixedSweepWithNoTestPackageCounts$' -v ./internal/factory/` → exit 1, stdout 원문 — §D.3 장부 **E-LEDGER-003** 인용 (트리 `f7606c7bc` 실측)
- **green-path**: M2가 마커 선두 거부를 제거하고 총 패스 기준 판정으로 바꿔 뒤집는다 → 위 명령 exit 0.
- **계약 보존(부정 절차 — REQ-MWQ2-004)**: 기존 거부 테스트 가족 `TestClassify`(빈 스윕·잘림·fail 이벤트·`go test` 무 `-json`) 전체가 E6 계열에서 GREEN을 유지해야 한다 — 총 패스 0 스트림과 fail 이벤트는 계속 거부된다.
- **판정**: 위 명령의 관측 종료 코드 0 + E6에서 `TestClassify` 계열 `ok`.

### AC-MWQ2-004 (maps REQ-MWQ2-005) — join-분류 경계 정합 봉인 (②, 미재현)

- **Given** `shellJoinArgs`가 인자 경계를 인용으로 보존해 만든 명령 문자열
- **When** 분류기(`shellSegments`/`shellFields`)가 그 문자열을 다시 파싱하면
- **Then** join이 보존한 경계가 필드로 읽혀 돌려지고, `go test` 인식과 `-json` 탐지가 유지된다; apostrophe 임베드 인자의 escape 바이트 잔류(현재 동작, 경계 유지·판정 무영향)는 봉인된다
- **RED-now**: 없음 — plan 단계 round-trip 관측에서 경계 붕괴 **미재현** (4/5 완전 일치; apostrophe 케이스는 `it\'s fine` 바이트 잔류, 하나의 필드 유지 — §Research §R2)
- **green-path**: M5 — 수리 없음, 봉인 테스트가 회귀 방어로 남는다.
- **판정**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^(TestRedT1582ClassifiersReadBackTheJoinedBoundaries|TestRedT1582JoinPreservesArgBoundariesInQuoting|TestRedT1582SingleArgScrubCompoundStaysVerbatim)$' ./internal/factory/ ./internal/cli/` → 종료 코드 0 + `ok` 두 패키지.

### AC-MWQ2-005 (maps REQ-MWQ2-006, REQ-MWQ2-010) — 측정 실패 REFUSED는 비제오 exit, 경합 verdict는 0 유지 (③)

- **Given** 재측정 레코드가 부재(또는 무효)인 카드의 병합 준비 점검
- **When** 본 트리 빌드 바이너리로 `./bin/moai factory merge ready --card <id> --spec <SPEC-ID> --develop develop`을 실행하면
- **Then** 출력에 `merge-readiness: REFUSED` + 실패 조건(재측정 레코드 조건)이 렌더되고 **exit 코드는 비제오**다; 창 경합 시나리오(waiting)의 exit은 여전히 0이다
- **RED-now (v0.3.0 plan 단계 실측 — 오버레이 테스트 `internal/cli/factory_merge_ready_red_t1582_test.go`)**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582MergeReadyMeasurementFailureStillExitsZero$' -v ./internal/cli/` → exit 1, stdout 원문 — §D.3 장부 **E-LEDGER-004** 인용 (트리 `f7606c7bc` 실측 — D4 fixture: (a)-(c) 통과, (d)만 파손)
- **green-path**: M3가 측정 실패 조건에 비제오 exit 코드를 부여해 뒤집는다 → 위 명령 exit 0(테스트 GREEN), REFUSED 출력과 조건 이름은 유지.
- **판정**: 위 명령 종료 코드 0 + `ok`(테스트 관측) + waiting 시나리오 관측 exit 0 유지(REQ-MWQ2-010) + 수리 후 본 트리 빌드 전면 관측(`make build` 후 `./bin/moai factory merge ready ...`, E4)에서 비제오.

### AC-MWQ2-006 (maps REQ-MWQ2-007) — cause-7의 hold+release 원자성 (R7-1)

- **Given** 병합 실패 + abort 후 dirty가 남은 cause-7 상태가 시딩된 merge step (`AfterPrecheck` seam dirty 파일 + `Git` seam merge 실패 — §Research §R5 설계)
- **When** hold 쓰기와 release 사이에 외부 mutation(status 갱신 또는 acquire)을 주입하는 오버레이 테스트가 인터리빙을 시도하면
- **Then** hold→release가 하나의 직렬화 구역 안에서 실행되어 주입이 끼어들 수 없다 — 원 호출자의 release는 hold 착지 뒤에 실패하지 않는다
- **RED-now / 시딩 실측 (v0.3.0 plan 단계)**: `TestRedT1582Cause7SeedingWritesHoldThenReleases` (`internal/factory/mergestep_red_t1582_test.go`) — `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582Cause7SeedingWritesHoldThenReleases$' -v ./internal/factory/` → exit 0, stdout: `--- PASS: TestRedT1582Cause7SeedingWritesHoldThenReleases (2.13s)` — 관측 내용: cause 7 도달(code 7) + hold policy 기록 + release로 창 비움·**승격 없음**(REQ-MWQ-018: hold 상태로 승격 금지 — 대기 티켓은 큐 유지) + 원인 에러 래핑 없음 (트리 `f7606c7bc` 실측). **원자화의 직접 RED는 M4 관측** — pre-repair 트리에 `writeMergeHold`와 `releaseHeldWindow` 사이 결정적 주입 지점(seam)이 없어 단일 프로세스 관측이 불가하며, 이 한계를 plan 단계에서 숨기지 않는다(plan.md §F M4).
- **green-path**: M4가 원자화 관측 테스트를 이 가족에 추가하고 hold 기록 + release를 하나의 `withIntegrationLockMutation`으로 옮겨 뒤집는다 → 인터리빙 주입이 있어도 hold→release 순서 유지 + release 실패 소멸.
- **판정 (실행 계수 가드)**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582Cause7' -v ./internal/factory/` (접두 앵커 — cause-7 가족 전체 선택) → **verbose 출력의 `--- PASS` 행 수가 2 이상**(시딩 설비 1 + M4 원자화 관측 테스트 1)이고 종료 코드 0 + `ok`. **`[no tests to run]`(0테스트 매치) 또는 PASS 1행은 통과가 아니다** — 가족이 M4 없이 시딩 설비만 남은 상태를 구별한다. E6의 `TestIntegrationMerge` 계열 GREEN 유지.

### AC-MWQ2-007 (maps 전체 REQ) — 영향 식별자의 전체 테스트 가족 재실행

- **Given** M1-M4가 수리한 식별자들 (`ValidateRemeasureRecord`, `countGoTestJSONTests`, `RunMergeStep`, `newFactoryMergeReadyCommand`, cause-7 구역)
- **When** 그 식별자들의 전체 테스트 가족을 한 번의 관측으로 돌리면
- **Then** `./internal/factory/`와 `./internal/cli/`의 영향 계열이 GREEN이다
- **판정**: 슬롯 임대 선행(`moai slot acquire --resource go-test-heavy --max-duration 45m` → 실행 → `moai slot release`) 후
  `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^(TestIntegrationMerge|TestIntegrationRem|TestRemeasure|TestClassify|TestShellJoin|TestMWQ19|TestRedT1582|TestMergeStep|TestFactoryMergeReady)' ./internal/factory/ ./internal/cli/` (시작 그룹 앵커 — 가족 접두 선택 의도: 각 접두의 더 긴 테스트 이름까지 포함; `TestMergeStep`·`TestFactoryMergeReady`·`TestRemeasure`는 M1-M4가 건드린 식별자의 나머지 가족)
  → 종료 코드 0 + 두 패키지 `ok` 행 (파이프 마스킹 없이 종료 코드 직접 관측 — t1576 착지 런 교훈).

### AC-MWQ2-008 (maps TRUST 5 Unified) — 서식·정적 품질

- **판정**: `gofmt -l internal/factory/ internal/cli/` → 빈 출력 (exit 0) + `go vet ./internal/factory/ ./internal/cli/` → exit 0. 두 명령의 관측 출력을 보고에 인용.

## §D.1 시나리오 보강 (edge cases)

- **AC-MWQ2-001 edge**: `Tree`가 형식 불량인 경우도 같은 거부 사유로 거부된다 — run 단계 테스트가 표 형태로 둘 다 커버(`Base`만 아니라 `Tree`).
- **AC-MWQ2-002 edge**: 정확히 12자 base도 렌더가 panicking하지 않아야 한다(경계 길이) — 길이 안전 접두는 12자 미만 전체를 커버.
- **AC-MWQ2-003 edge**: 마커가 **실제로** 빈 스윕을 나타내는 경우(전 패키지 no-test, 총 패스 0)는 계속 거부된다 — 부정 절차 테스트가 존재해야 한다.
- **AC-MWQ2-005 edge**: `--json` 경로의 REFUSED 판정도 동일 exit 코드를 따른다 — JSON 소비자(스크립트)가 판독하는 주 고객이다.

## §D.2 품질 게이트 기준

- TRUST 5: Tested(위 8 AC) · Readable(수리 주석은 기존 밀도·언어 영어 준수) · Unified(gofmt) · Secured(해당 없음 — 입력은 git·레코드 파일) · Trackable(Conventional Commits + 카드 id).
- 검증 근원: 모든 go test는 본 트리 컴파일, 모든 moai CLI 관측은 `make build` 산물 `./bin/moai` 경로 호출 — 설치본 판정 금지.

## §D.3 증거 장부 — RED 원문 stdout (verification-completeness §2.1 캐리어)

> 각 항목은 트리 `f7606c7bc`에서 실행된 원문 stdout(raw bytes)이다. 감사자가 iteration 1-3에서 4/4 바이트 일치 재실행 확인(plan-audit.md §Evidence E-G — "all four claimed REDs ... reproduced byte-consistently", receipt rcpt-7b69c79f30c6981fbbcc0aa1) — 이 장부는 그 확인분을 인용 형태로 옮긴 것이다. 괄호 안 지속 시간은 실행마다 변동하는 값이다.

### E-LEDGER-001 — AC-MWQ2-001 (검증기 거부, R7-3 단위)

명령: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582VerifierAdmitsAShortBaseSHA$' -v ./internal/factory/` — exit 1

```
=== RUN   TestRedT1582VerifierAdmitsAShortBaseSHA
    remeasure_red_t1582_test.go:58: RED t1582-R7-3: the verifier admits a record whose Base is 3 bytes — it reaches the [:12] message renders downstream
--- FAIL: TestRedT1582VerifierAdmitsAShortBaseSHA (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/factory
```

### E-LEDGER-002 — AC-MWQ2-002 (창 점유 panic, R7-3 e2e)

명령: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease$' -v ./internal/factory/` — exit 1

```
=== RUN   TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease
    remeasure_red_t1582_test.go:89: RED t1582-R7-3: the merge step panicked on the short base "bad" (runtime error: slice bounds out of range [:12] with length 3) and died before releasing the window — the window stays held
--- FAIL: TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease (12.33s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/factory
```

### E-LEDGER-003 — AC-MWQ2-003 (혼합 스윕 거짓 거부, R7-2)

명령: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582MixedSweepWithNoTestPackageCounts$' -v ./internal/factory/` — exit 1

```
=== RUN   TestRedT1582MixedSweepWithNoTestPackageCounts
    remeasure_red_t1582_test.go:35: RED t1582-R7-2: the [no test files] marker anywhere in the stream refuses a mixed sweep that measured 1 passing test: runner reported [no test files] — an empty sweep cannot stand for a re-measure
--- FAIL: TestRedT1582MixedSweepWithNoTestPackageCounts (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/factory
```

### E-LEDGER-004 — AC-MWQ2-005 (측정 실패 REFUSED exit 0, ③)

명령: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582MergeReadyMeasurementFailureStillExitsZero$' -v ./internal/cli/` — exit 1

```
=== RUN   TestRedT1582MergeReadyMeasurementFailureStillExitsZero
    factory_merge_ready_red_t1582_test.go:41: RED t1582-AC005: the measurement-failure REFUSED verdict still exits 0 (err nil) — out: merge-readiness: REFUSED — failing condition: re-measure-record
        no window was taken
          sync-audit     PASS  §E.4 sync_status: complete — the sync phase record reads closed
          conflict-free  PASS  merge-tree develop + WT-card clean; result tree cca1377e26e2ce841ee10fd78f5f61e6dda22f10
          tree-identity  PASS  merge result tree cca1377e26e2ce841ee10fd78f5f61e6dda22f10 == card branch tree cca1377e26e2ce841ee10fd78f5f61e6dda22f10 (HEAD^{{tree}} == HEAD^2^{{tree}} holds)
          re-measure-record FAIL  no valid re-measure record for the candidate tree (no re-measure record for tree cca1377e26e2ce841ee10fd78f5f61e6dda22f10)
        {"verdict":"refused","lane":"lane-9","card":"t9001","failed_condition":"re-measure-record","detail":"the condition triple failed — no window was taken","checks":[{"name":"sync-audit","passed":true,"detail":"§E.4 sync_status: complete — the sync phase record reads closed"},{"name":"conflict-free","passed":true,"detail":"merge-tree develop + WT-card clean; result tree cca1377e26e2ce841ee10fd78f5f61e6dda22f10"},{"name":"tree-identity","passed":true,"detail":"merge result tree cca1377e26e2ce841ee10fd78f5f61e6dda22f10 == card branch tree cca1377e26e2ce841ee10fd78f5f61e6dda22f10 (HEAD^{{tree}} == HEAD^2^{{tree}} holds)"},{"name":"re-measure-record","passed":false,"detail":"no valid re-measure record for the candidate tree (no re-measure record for tree cca1377e26e2ce841ee10fd78f5f61e6dda22f10)"}]}
--- FAIL: TestRedT1582MergeReadyMeasurementFailureStillExitsZero (10.88s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli
```

## Definition of Done

- [ ] 8 AC 전부 판정 명령 관측으로 PASS (5-섹션 보고 형식)
- [ ] RED 3건(R7-2, R7-3 단위/e2e)의 RED-before/GREEN-after 쌍이 progress.md §E.2에 기록
- [ ] M5 동일 클래스 스윕 3종 grep 기록
- [ ] decision-index Q1(FOUNDER)이 운영자/리더 처분됨 — Q1이 미처분 상태로 run 진입 시 plan-audit의 clarification gate 대상
- [ ] Out of Scope 경계(spec.md §3) 밖 파일 무변경 — **tracked 수정과 untracked 신규를 합쳐 수집해 대조**: `git status --porcelain` + `git ls-files --others --exclude-standard`의 합집합이 Out of Scope 허용 목록 밖을 가리키면 실패. **허용 목록**(D21 — 수리 대상과 수리가 갱신하는 파일까지 포함, 미포함 시 올바로 완료된 작업을 실패로 판정하는 내부 모순): SPEC 디렉터리 `.moai/specs/SPEC-MERGE-WINDOW-QUEUE-002/`; RED 오버레이 4파일 `internal/factory/remeasure_red_t1582_test.go`·`mergestep_red_t1582_test.go`, `internal/cli/shelljoin_red_t1582_test.go`·`factory_merge_ready_red_t1582_test.go`; **수리 대상 소스 4파일** `internal/factory/integration_remeasure.go`·`internal/factory/integration_merge_step.go`, `internal/cli/integration_remeasure.go`·`internal/cli/factory_merge.go`; **D3 재작성·계약 갱신 테스트** `internal/factory/integration_remeasure_run_test.go` 및 E6 가족이 커버하는 회귀 테스트(`internal/factory/integration_merge_step_test.go`·`integration_remeasure_test.go`, `internal/cli/factory_merge_test.go`·`integration_remeasure_cmd_test.go`·`integration_remeasure_test.go`). (`git diff --name-only` 단독은 untracked를 못 봐서 신규 파일을 전멸시킨다 — D17 실측: 본 트리에서 diff 0행 vs porcelain 3항목 7파일.)
