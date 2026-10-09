# SPEC-MERGE-WINDOW-QUEUE-002 — 구현 계획

> 카드 t1582 · plan 단계 산출물 · Tier M (spec.md + plan.md + acceptance.md + progress.md) · decision-index.md 동반 (`interview.decision_gate: on`)

## §A Context

- 트리: 카드 워크트리(브랜치 `WT-p1-p2-t1576`), base = develop tip `f7606c7bc` (이동 좌표 아님 — 본 SPEC의 RED 측정 트리 SHA이며, §Research 좌표는 이 커밋 기준).
- 상위 SPEC: `.moai/specs/SPEC-MERGE-WINDOW-QUEUE-001/` (status: completed) — REQ-MWQ-014~021 계약을 계승·갱신.
- 증거 원천: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1576/.moai/reports/t1576/progress.md` (다른 트리, 읽기 전용 — 라운드 1-7 처분표), 본 트리의 `.moai/reports/t1576/card-review.md` 사본, lane-7 t1571 게이트 관측 3건.
- 개발 모드: `tdd` (`.moai/config/sections/quality.yaml`) — RED-GREEN-REFACTOR. RED-first는 카드의 강제 절차이기도 하다.
- plan 단계에서 이미 작성한 RED 재현 오버레이 테스트 파일 (run 단계가 GREEN으로 뒤집는 대상):
  - `internal/factory/remeasure_red_t1582_test.go` — R7-2·R7-3 RED 3건 + item ② round-trip 봉인 + item ① probe 관측
  - `internal/cli/shelljoin_red_t1582_test.go` — item ② join 경계 보존 봉인 + env-scrub 단일 인자 봉인

### RED 재현 상태 (plan 단계 실측, 트리 `f7606c7bc`)

| 항목 | 상태 | 관측 명령 (빌드 없이 go test — 도구는 본 트리 컴파일) | 관측 출력 (발췌) |
|---|---|---|---|
| ④-R7-2 | **재현됨 (RED)** | `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582MixedSweepWithNoTestPackageCounts$' ./internal/factory/` | `RED t1582-R7-2: the [no test files] marker anywhere in the stream refuses a mixed sweep that measured 1 passing test: runner reported [no test files] — an empty sweep cannot stand for a re-measure` |
| ④-R7-3 (단위) | **재현됨 (RED)** | `... go test -count=1 -run '^TestRedT1582VerifierAdmitsAShortBaseSHA$' ./internal/factory/` | `RED t1582-R7-3: the verifier admits a record whose Base is 3 bytes — it reaches the [:12] message renders downstream` |
| ④-R7-3 (e2e) | **재현됨 (RED)** | `... go test -count=1 -run '^TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease$' ./internal/factory/` | `RED t1582-R7-3: the merge step panicked on the short base "bad" (runtime error: slice bounds out of range [:12] with length 3) and died before releasing the window — the window stays held` |
| ① probe | **관측 기록 (문서화된 잔여 — 수리 대상 아님)** | `... -run 'TestRedT1582ProbeNonTestCommandFailureSemanticsRideExitZero' ./internal/factory/` | `observed: failure semantics in a non-test command's output ride exit 0 to a valid record — spec.md §D residual, not repaired here` |
| ② round-trip | **미재현 (봉인 GREEN)** | `... -run '^TestRedT1582ClassifiersReadBackTheJoinedBoundaries$' ./internal/factory/` + `-run '^TestRedT1582JoinPreservesArgBoundariesInQuoting$' ./internal/cli/` | 4/5 완전 일치; apostrophe 케이스: `got ... "it\\'s fine" ...` vs 원본 `"it's fine"` — 경계는 하나의 필드로 유지(붕괴 없음), escape 바이트 잔류는 도구·`-json` 판정에 무해 → 현재 동작 봉인 |
| ③ (merge-ready) | **재현됨 (RED — v0.3.0 plan 단계 실측)** | `... go test -count=1 -run '^TestRedT1582MergeReadyMeasurementFailureStillExitsZero$' -v ./internal/cli/` | `RED t1582-AC005: the measurement-failure REFUSED verdict still exits 0 (err nil)` — 조건 표: `sync-audit PASS`·`conflict-free PASS`·`tree-identity PASS`·`re-measure-record FAIL no valid re-measure record for the candidate tree` + `merge-readiness: REFUSED — failing condition: re-measure-record` 출력 |
| ④-R7-1 (시딩 설비) | **관측됨 (positive control GREEN — v0.3.0 plan 단계)** | `... go test -count=1 -run '^TestRedT1582Cause7SeedingWritesHoldThenReleases$' -v ./internal/factory/` | `--- PASS: TestRedT1582Cause7SeedingWritesHoldThenReleases (2.13s)` exit 0 — cause 7 도달(code 7) + hold policy 기록 + release로 창 비움·승격 없음(REQ-MWQ-018: hold 상태로 승격 금지) + 원인 에러 래핑 없음. 원자화 직접 RED는 M4 관측(§F M4) |

## §B Known Issues

- B1 (동일 클래스 재민팅): t1576의 경험 — 수리는 지명 인스턴스를 고치지만 새 텍스트에서 같은 클래스 새 인스턴스가 나온다(`feedback_repairs_fix_named_instances_but_new_same_class_instances`). 각 수리 뒤 동일 클래스 스윕을 의무화(M5).
- B2 (R7-3 e2e 테스트 비용): plan 단계 실측에서 `TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease` 단독 12-47초 — fixture의 실제 git 작업(워크트리 생성·병합) 비용. run 단계에서 이 계열을 돌릴 때 슬롯 임대 고려.
- B3 (이동 좌표): t1576 인계 기록의 좌표(`integration_merge_step.go:287/509`, `integration_remeasure.go:270`)는 t1576 트리 기준 — 본 트리에서 일부 이동(§Research). 좌표는 증거가 아니라 탐색 시작점이고, 판정 근거는 함수 이름 + 관측 출력이다.
- B4 (카드 텍스트의 경로 보정): 카드가 R7-3 지점을 `internal/cli/integration_merge_step.go`로 적었으나 실제는 `internal/factory/integration_merge_step.go`. 분류기들(`shellSegments` 등)도 카드는 cli 표면으로 서술했으나 실제로는 `internal/factory/integration_remeasure.go` 소속(`shellJoinArgs`만 cli) — item ②의 round-trip 검증은 두 패키지에 걸친다.
- B5 (`strings.Fields` 서술은 수리 전 것): 카드 텍스트의 "classifiers parse the quoted-joined string with `strings.Fields`"는 t1576 수리 이전 서술 — 현재 코드는 quote 인식 `shellFields`를 쓴다(285행). 미재현 결론의 근거 일부.

## §C Pre-flight (run 단계 시작 전)

```bash
git rev-parse --short HEAD          # 기대: f7606c7bc (흡수로 이동했으면 §Research 좌표 재확인)
git branch --show-current           # 기대: WT-p1-p2-t1576
unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^TestRedT1582' ./internal/factory/ ./internal/cli/   # RED 3건 FAIL + 봉인 GREEN 재확인 (baseline-first; 접두 앵커 — RED 가족 전체 선택 의도)
make build                          # ③ 관측용 본 트리 바이너리 (./bin/moai 경로 호출 — 설치본 금지)
```

## §D Constraints

- **PRESERVE** (건드리지 않는 것): `internal/hook/protected_zone_path.go`, `todo_issuance.go`, `todo_auto_lane.go`, `factory_card.go`의 nominate 표면 — Out of Scope 명시 항목. `internal/factory/integration_remeasure.go`의 fail 이벤트 거부 로직(t1576 r1 수리)의 계약 — REQ-MWQ2-004로 승격될 뿐, 동작을 바꾸지 않는다.
- RED-first: 모든 수리는 대응 RED 테스트가 실패하는 것이 관측된 뒤에만 설계·적용한다. plan 단계 관측 출력(§A 표)이 그 증거다.
- RED-only baseline 커밋(D6): run 단계의 **첫 커밋**은 RED 오버레이 테스트 2파일(`internal/factory/remeasure_red_t1582_test.go`, `internal/cli/shelljoin_red_t1582_test.go`)을 첫 수리 커밋 **앞에서 단독**으로 커밋한다 (`test(SPEC-MERGE-WINDOW-QUEUE-002): RED baseline — plan-phase reproduction overlays`). 커밋 그래프가 RED-before 관측 순서를 증언한다 — 같은 커밋에 수리를 섞으면 그래프는 관측 순서를 증언할 수 없다 (verification-claim-integrity §2.3 ordering attribution).
- 전 계열 재실행: 수리가 건드린 식별자는 새 테스트만이 아니라 그 식별자의 전체 테스트 가족을 다시 돌린다.
- 검증은 lane-local: `./internal/factory/`·`./internal/cli/` 영향 계열만. 전체 스위트 금지(타 레인 경합 hang — t1542 교훈). 장시간 실행 전 `moai slot acquire --resource go-test-heavy --max-duration 45m` → 실행 → `moai slot release`.
- 커밋 규율: Conventional Commits (`fix(SPEC-MERGE-WINDOW-QUEUE-002): M1 ...`), `🗿 MoAI` 트레일러, 카드 id 병기.
- 금지: `--no-verify`, `git add -A`류 스윕 스테이징, 설치본 moai 바이너리로 측정.

## §Research — 결함별 현재 코드 위치 (트리 `f7606c7bc` 기준, 이동 좌표)

> 좌표는 이동하는 값이다(moving coordinates — baseline doctrine): 판정 근거는 함수 이름과 관측 출력이고, 행번호는 탐색 시작점일 뿐이다.

### §R1 — item ① (ValidateRemeasureRecord 무시 클래스)

- `internal/factory/integration_remeasure.go` — `ValidateRemeasureRecord`(128행): exit 비제오·구조 없음·빈 스윕·식별자 누락만 검사.
- `ClassifyStructuredOutput`(163행) → `countGoTestJSONTests`(423행): fail 이벤트(per-test·패키지) 거부는 455-461행 — **t1576 라운드 1 수리, 유지됨**. 잘림 감지(started/finished) 429-481행, 세그먼트 재실행 재-arming 463-469행(라운드 8).
- **프루빙 결과**: (a) non-`go test` 명령의 실패 의미 출력 → `ClassifyStructuredOutput`이 structured=false를 반환 → exit 0이면 valid — **spec.md §D 잔여, candidate-CI 소관** (probe 테스트로 기록). (b) `sh -c` 세그먼트 exit 마스킹 — 실패 세그먼트의 출력이 합쳐진 stdout에 남아 있으면 fail 이벤트로 이미 잡힘; 리다이렉션으로 지운 것은 호출자 측정 포기. (c) 마커 상호작용 — **실질 재현 = R7-2와 동일 지점**. 결론: ①의 수리 실체는 REQ-MWQ2-003(R7-2) + REQ-MWQ2-004(계약 승격).

### §R2 — item ② (sh -c 인자 경계)

- 실행 쪽: `internal/cli/integration_remeasure.go` — `shellJoinArgs`(61행)·`shellQuoteArg`(103행)·단일 인자 verbatim 계약(62-64행, t1576 r8).
- 분류 쪽: `internal/factory/integration_remeasure.go` — `shellSegments`(205행)·`shellFields`(285행)·`isGoTestCommand`(187행)·`stripLeadingEnv`(371행)·`requestsGoTestJSON`(395행).
- **round-trip 관측**: join이 보존한 경계(공백+`|`, `;`, `&`, 환경할당 값의 공백)를 분류기가 모두 읽어 돌림(4/5 완전 일치). apostrophe 임베드 인자(`it's fine` → `'it'\''s fine'`)는 분류 필드에 escape 바이트가 `it\'s fine`로 잔류하지만 **하나의 필드로 유지** — 경계 붕괴 아니며 `go test` 인식·`-json` 탐지 무영향. **미재현 — 봉인 테스트로 고정, 수리 없음**(REQ-MWQ2-005는 이 정합성을 계약으로 고정).

### §R3 — item ③ (INVALID 판정 + nil)

- `internal/cli/factory_merge.go` — `newFactoryMergeReadyCommand`: 조건 실패 → REFUSED 출력(142행) + `emitFactoryMergeVerdict` 반환 — `asJSON=false`면 nil(330-333행) = **exit 0**. 네 번째 조건(재측정 레코드)은 `VerifyRemeasure`(112-121행)에서 `ReadRemeasureRecord`/`ValidateRemeasureRecord` 실패로 REFUSED를 만든다 — **측정 실패 클래스가 exit 0으로 나가는 지점**.
- `newFactoryMergeGateCommand`(262-315행): REFUSED/PROCEED verdict + nil — FACTORY-LANE-AUTONOMY-001의 문서화된 의도(파일 상단 12-14행 계약 주석), Out of Scope.
- `internal/cli/factory_card.go` — `factoryCompleteCard`: release 실패를 출력하고 nil로 계속(1939-1941, 1958-1960, 2010-2012행) — 병합·전이 성공 뒤의 정리 실패라 클래스 밖(§R5).
- `internal/cli/integration_remeasure.go` — `remeasureVerdictError`(126-131행): t1576 r2 수리 유지 — INVALID 판정이 오류로 반환됨(이미 방어됨).
- **수리 방향**: `merge ready`가 refused 판정을 낼 때 실패 조건이 재측정 레코드 부재/무효(측정 실패)면 비제오 exit, 창 경합(waiting·holder)은 verdict 유지 — decision-index Q1이 범위를 묻는다.

### §R4 — ④-R7-3 (짧은 base panic)

- `internal/factory/integration_merge_step.go` — `RunMergeStep`: cause-2 분기 287행 `record.Base[:12]`, cause-3 분기 297행 `record.Base[:12]`, 섹션 내 tip 재확인 397행 `tip[:12]`, `postMergeHold` 729행 `mergeSHA[:12]`, 병합 트리 불일치 529-531행 `mergedTree[:12]`/`record.Tree[:12]`.
- `ValidateRemeasureRecord`(`internal/factory/integration_remeasure.go` 128행): Base 검사는 **비어있음만**(135행) — "bad" 통과 → cause 1을 타지 않고 287행 panic.
- 길이 안전 접두의 기존 패턴: 같은 파일의 `minStrLen`(586행)이 `CompletePostMergeHold`(578행)에서 이미 사용 — 수리는 (i) 검증기에 SHA 형식 검증 추가(근본) + (ii) 렌더 경로의 길이 안전 접두(방어).
- **plan 단계 관측**: panic 재현 + 창 held 유지 관측 완료(§A 표). t1576 트리 좌표 287행과 동일행 — 병합 흡수 시 이동 가능.

### §R5 — ④-R7-1 (cause-7 hold/release 분리)

- `internal/factory/integration_merge_step.go` — cause-7 경로 509-518행: `writeMergeHold`(714행, policy 쓰기)와 `releaseWindow`(695행 → `ReleaseIntegrationLock`)가 **서로 다른 mutation**. t1576 트리 좌표 509행.
- 창의 구조적 사실: hold 쓰기 완료와 release 실행 사이에 다른 프로세스의 `status` 갱신(stale 정리·승격) 또는 `acquire --force`가 끼어들 수 있는 시간창이 존재 — 인터리빙 시 원 호출자의 release는 foreign-holder로 실패하고 hold와 승격의 순서가 보장되지 않는다.
- **수리 방향**(카드 지시): 실패 판정 + hold 기록 + release를 하나의 `withIntegrationLockMutation` 임계 구역으로 — decision-index Q2가 잠금 기계를 묻는다.
- **run 단계 RED 설계**(M4): cause-7 시딩 = `AfterPrecheck` seam에서 untracked dirty 파일 생성 + `Git` seam이 `merge` 호출에 실패를 반환 → abort(no-op) 후 dirty 유지 → cause 7 진입. 인터리빙 관측 = hold 쓰기와 release 사이에 두 번째 mutation(외부 acquire) 주입 → 현재: release 실패 + 승격 우회 순서 관측; 수리 후: 단일 구역에서 hold+release 원자 실행.

### §R6 — ④-R7-2 (마커 거짓 거부)

- `internal/factory/integration_remeasure.go` — `emptySweepMarker` 상수(416행), `countGoTestJSONTests`의 선두 마커 거부(424-426행): `strings.Contains(text, "[no test files]")` — output 이벤트 안의 마커도 잡는다. t1576 트리 좌표 270행 → 본 트리 424행(라운드 8/9 수리로 파일 성장, 이동 확인).
- 실제 `go test -json`에서 무테스트 패키지는 `{"Action":"output",...,"Output":"?   \tpkg\t[no test files]\n"}` + 패키지 수준 `{"Action":"skip","Package":"..."}` 종료 이벤트 — 현재 코드는 패키지 수준 skip을 종료 이벤트로 이미 인정(475행 주석), 마커 선두 거부만이 정상 혼합 스윕을 막는다.
- **수리 방향**(카드 지시): 총 per-test pass 개수 기준 판정 + 패키지 skip 종료 인정(기존) — REQ-MWQ-015의 "runner-reported empty sweep" 정의를 "총 per-test pass 0"으로 갱신. 완전 잘림 방어(started/finished)는 유지.

## §E Self-Verification (run 단계 산출물 — acceptance.md의 AC와 1:1)

| # | 증거 | 명령 (worktree 안, env-scrub 복합) |
|---|---|---|
| E1 | AC-MWQ2-001/002 (R7-3) | `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^(TestRedT1582VerifierAdmitsAShortBaseSHA|TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease)$' ./internal/factory/` |
| E2 | AC-MWQ2-003 (R7-2) | `... && go test -count=1 -run '^TestRedT1582MixedSweepWithNoTestPackageCounts$' ./internal/factory/` |
| E3 | AC-MWQ2-004 (② 봉인) | `... && go test -count=1 -run '^(TestRedT1582ClassifiersReadBackTheJoinedBoundaries|TestRedT1582JoinPreservesArgBoundariesInQuoting|TestRedT1582SingleArgScrubCompoundStaysVerbatim)$' ./internal/factory/ ./internal/cli/` |
| E4 | AC-MWQ2-005 (③) | 본 트리 빌드로 관측: `make build && ./bin/moai factory merge ready ...` (M3 절차) — 수리 전 exit 0+REFUSED 출력, 수리 후 비제오 |
| E5 | AC-MWQ2-006 (R7-1) | `... && go test -count=1 -run '^TestRedT1582Cause7' -v ./internal/factory/` — verbose 출력의 `--- PASS` 행 **2 이상**(시딩 설비 + M4 원자화 관측; 0테스트 매치나 1행은 통과 아님 — 실행 계수 가드, acceptance.md AC-006) |
| E6 | AC-MWQ2-007 (전 계열) | 슬롯 임대 후 `... && go test -count=1 -run '^(TestIntegrationMerge|TestIntegrationRem|TestRemeasure|TestClassify|TestShellJoin|TestMWQ19|TestRedT1582|TestMergeStep|TestFactoryMergeReady)' ./internal/factory/ ./internal/cli/` (시작 그룹 앵커 — 가족 접두 선택; `TestMergeStep`·`TestFactoryMergeReady`·`TestRemeasure`는 M1-M4 영향 식별자의 나머지 가족) |
| E7 | AC-MWQ2-008 (품질) | `gofmt -l internal/factory/ internal/cli/` 빈 출력 + `go vet ./internal/factory/ ./internal/cli/` exit 0 |

각 E-항목은 verification-claim-integrity §3의 5-섹션 형식(Claim/Evidence/Baseline-attribution/Gaps/Residual-risk)으로 보고한다.

## §F Milestones (Priority 순 — 카드 지정 우선순위 따름)

### M1 — [Priority High] ④-R7-3: SHA 형식 검증 + 길이 안전 접두 (창 점유 panic — 최우선)

- **RED (이미 관측됨)**: `TestRedT1582VerifierAdmitsAShortBaseSHA`, `TestRedT1582ShortBaseSHAPanicsBeforeWindowRelease` — §A 표의 관측 출력.
- 수리: (i) `ValidateRemeasureRecord`에 tree/Base의 40-hex 형식 검증 추가(레코드 무효 사유로 조기 거부); (ii) **총괄 규칙 — 길이 불문 모든 SHA 접두 렌더는 `minStrLen` 길이 안전 경유**이며, 현재 실측 대상은 **두 패키지 12곳**(트리 `f7606c7bc` 실측, `grep -n '\[:12\]'`; 미래 변형도 총괄 규칙이 흡수):
  - `internal/factory/integration_merge_step.go` 10곳 — 280(cause 1 거부 메시지)·287(cause 2)·292(cause 10)·297(cause 3)·300(cause 4)·307(cause 5)·397(섹션 내 tip 재확인)·529·531(cause 8 병합 트리 불일치)·729(postMergeHold merge SHA);
  - `internal/cli/integration_remeasure.go` 2곳 — 45(remeasure verb 기록 출력 `rec.Tree[:12]`·`rec.Base[:12]`)·128(`remeasureVerdictError`가 M1(i)이 거부한 바로 그 레코드의 `rec.Tree[:12]`를 렌더 — **cli 파일도 M1 스코프에 포함**, §G 안티패턴("검증기 거부만 추가하고 렌더 경로를 방어하지 않는 것")의 생표본: 거부 경로 자체가 panic하면 M1(i)이 의미 없음).
  - (i)이 근본, (ii)는 방어 — (i)만으로 12곳의 모든 입력을 막지 못하는 형태(수기 레코드의 다른 필드 형태)에 대비.
- GREEN 판정: E1 두 테스트 GREEN + cause 1 코드 + 창 release + C 승격 관측.
- 전 계열: E6의 `TestIntegrationMerge`·`TestMergeStep` 가족 전체 재실행(+ cli의 `TestIntegrationRem` 계열 — 45·128행이 건드리는 verb 출력 계약).

### M2 — [Priority Medium] ④-R7-2: 마커 → 총 per-test 패스 기준 판정

- **RED (이미 관측됨)**: `TestRedT1582MixedSweepWithNoTestPackageCounts` — §A 표.
- 수리: `countGoTestJSONTests`의 선두 마커 거부 제거, 패키지 skip을 종료 이벤트로 유지, 총 per-test pass가 0일 때만 빈 스윕 거부. REQ-MWQ-015 정의 갱신을 spec/spec.md §C.2에 명시(REQ-MWQ2-003 스윕 판정 + REQ-MWQ2-009 빈 스윕 무효 — D5 분할 후).
- **기존 핀 테스트 처분 (D3)**: `TestRemeasureMixedTestAndEmptyPackageRemainsInvalid` (`internal/factory/integration_remeasure_run_test.go:185` — 실제 `go test -json ./...`의 혼합 보고(실제 테스트 통과 + 마커)에서 레코드가 invalid로 남는 것을 핀, "Keep that conservative contract" 주석)는 이 계약 갱신으로 **M2에서 재작성한다**: 본문을 수리 후 기대로 반전(혼합 스윕은 레코드 valid — 총 per-test pass 카운트 기록)하고, 총 패스 0 스트림은 여전히 invalid인 부정 절차를 같은 테스트 가족에 유지한다. 재작성 없이 두면 이 테스트가 M2 수리 뒤에도 구 계약을 핀 채 적색으로 남아 E6를 깬다. 재작성 후 이 테스트는 M2의 GREEN 감시 가족에 편입되고 E6의 `TestRemeasure` 접두로 재실행된다.
- GREEN 판정: E2 GREEN + 기존 빈 스윕·잘림·fail 이벤트 거부 테스트(`TestClassify`) 전체 GREEN 유지 + D3 재작성 테스트 GREEN.

### M3 — [Priority Medium] ③: 측정 실패 판정의 비제오 exit

- **RED (이미 관측됨 — v0.3.0 plan 단계)**: `TestRedT1582MergeReadyMeasurementFailureStillExitsZero` (`internal/cli/factory_merge_ready_red_t1582_test.go`) — §A 표의 verbatim 관측. fixture는 D4 설계를 그대로 구현: `mergeReadyFixture(t, "complete", ...)`로 (a) sync-audit·(b) conflict-free·(c) tree-identity를 통과시키고 시딩된 재측정 레코드를 제거해 (d)만 파손 — fixture 자신의 주석이 명명하는 "(d)-failure route"(clears the store). `mergeReadyFixture` 재사용으로 CLAUDE_PROJECT_DIR lock root·cwd가 모두 fixture 안에 고정된다.
- **관측된 RED 출력(요지)**: 조건 표 `sync-audit PASS`·`conflict-free PASS`·`tree-identity PASS`·`re-measure-record FAIL` + `merge-readiness: REFUSED — failing condition: re-measure-record` + **err nil(exit 0)** — `CheckRemeasureRecord = "re-measure-record"`(`internal/factorylane/merge.go:50`).
- 수리: REFUSED 원인이 측정 실패 조건(재측정 레코드 부재/무효)일 때 `exitCodeError` 계열의 비제오 반환 — 창 경합 verdict(waiting)는 nil 유지(decision-index Q1의 Default). `factory_merge.go` REFUSED 분기에 조건별 exit 코드 부여.
- GREEN 판정: 위 RED 테스트가 GREEN으로 뒤집힘(비제오 exit + REFUSED 출력 유지), waiting 시나리오는 exit 0 유지(REQ-MWQ2-006 + REQ-MWQ2-010).
- CLI 전면 관측은 수리 후 본 트리 빌드로 재확인: `make build && ./bin/moai factory merge ready ...`(E4).
- clarification 상태: 마커 없음 — Q1이 FOUNDER 행으로 decision-index에 기록, Default는 측정 실패 클래스만 비제오; Q1의 운영자 처분이 run 진입 게이트.

### M4 — [Priority Medium] ④-R7-1: cause-7 hold+release 원자화

- **시딩 설비 (이미 관측됨 — v0.3.0 plan 단계)**: `TestRedT1582Cause7SeedingWritesHoldThenReleases` (`internal/factory/mergestep_red_t1582_test.go`) — Git seam이 merge 호출에 실패와 untracked residue를 주입하고 abort 응답만 스크립트(나머지 git 호출은 실제 러너 위임)해 cause 7에 도달시킨다. 관측(§A 표): 코드 7 + hold policy 기록 + release로 창 비움·승격 없음(REQ-MWQ-018: hold 상태로 승격 금지) + 원인 에러 래핑 없음.
- **원자화 직접 RED (run 단계 M4 관측)**: pre-repair 트리에는 `writeMergeHold`와 `releaseHeldWindow` 사이에 결정적 주입 지점(seam)이 없어 단일 프로세스 관측이 불가 — M4가 수리 설계와 함께 인터리빙 주입 관측 테스트를 이 가족에 작성한다(외부 mutation 주입 시 hold→release 순서). plan 단계에서 이 사실을 숨기지 않고 AC-006의 **실행 계수 가드**(verbose `--- PASS` ≥ 2)로 vacuous window를 닫는다.
- 수리: hold 기록 + release를 하나의 `withIntegrationLockMutation` 안으로 — `ReleaseIntegrationLock`의 승격 로직과 policy 쓰기가 같은 직렬화 구역을 공유하도록(decision-index Q2 Default: 기존 mutation 기계 재사용).
- GREEN 판정: 인터리빙 주입이 있어도 hold→release 순서가 구역 안에서 유지 + release 실패가 사라짐 + positive control 유지 GREEN.

### M5 — [Priority Low] ② 봉인 완료 + 동일 클래스 스윕

- ②는 미재현 — 봉인 테스트 GREEN 유지(E3). 수리 없음.
- 동일 클래스 스윕 기록: M1-M4가 건드린 표면(검증기·countGoTestJSONTests·merge step 렌더·merge ready 분기·cause-7 구역)에서 같은 클래스 새 인스턴스 스캔 — `grep -rn '\[:12\]' internal/factory internal/cli`(길이 안전 접두 미적용 잔존 스캔; **`-r` 필수 — GNU grep(리눅스 CI)은 디렉터리 피연산자를 하강하지 않아 `-r` 없으면 `Is a directory` 오류, macOS BSD grep과 가용성 분기 — D20**), `grep -rn 'Contains.*no test\|noTestFiles' internal/`(마커 거부 잔존 스캔), `grep -n 'strings.Fields' internal/factory/integration_remeasure.go`(무식한 분할 재유입 스캔) — 각 스캔은 **no-match(exit 1)와 탐색 오류(exit 2)를 구분 기록**하고 결과를 progress.md §E.2에 남긴다.

### M6 — [Priority Medium] 전 계열 재실행 + 품질 체인

- 슬롯 임대 후 E6(영향 계열 전체) + E7(gofmt·vet) — 파이프 마스킹 없이 종료 코드 관측(t1576 착지 런의 교훈).
- RED 3건이 GREEN으로 뒤집혔음을 §E.2에 RED-before/GREEN-after 쌍으로 기록.

## §G Anti-Patterns

- 검증기 거부만 추가하고 렌더 경로를 방어하지 않는 것(방어 없는 근본 수리) — 수기 레코드의 다른 형태가 다른 렌더를 때린다.
- 마커 거부 제거를 "빈 스윕 허용"으로 잘못 일반화하는 것 — 총 per-test pass 0은 여전히 무효다(REQ-MWQ-015의 본 취지 유지).
- `merge gate`의 verdict 설계까지 뒤집는 것 — FACTORY-LANE-AUTONOMY-001 계약 파괴.
- 인터리빙 관측 없이 원자화만 적용하는 것(RED 없는 GREEN) — t1576의 F4/F5 선례처럼 구역 재배치가 새 창을 만들 수 있다.

## §H Cross-References

- 상위: `.moai/specs/SPEC-MERGE-WINDOW-QUEUE-001/` (REQ-MWQ-014/015/017/018/019/021 계승·갱신)
- 인계 원천: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1576/.moai/reports/t1576/progress.md` 리뷰 라운드 6-7 표 (읽기 전용)
- 계약 표면: `internal/cli/factory_merge.go` 상단 verdict 계약 주석 (SPEC-FACTORY-LANE-AUTONOMY-001)
- RED 오버레이: `internal/factory/remeasure_red_t1582_test.go`, `internal/cli/shelljoin_red_t1582_test.go` (본 트리, plan 단계 작성 — run 단계 M1-M4에서 GREEN으로 뒤집는 대상)
