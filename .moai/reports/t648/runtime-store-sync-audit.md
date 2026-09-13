# Todo 실행 저장 기반 — 독립 구현 감사

SPEC: **SPEC-TODO-RUNTIME-STORE-001**
Overall Verdict: **PASS — 최초 내부 저장 기반 4 AC에 한정**
가중 점수: **100/100**
날짜: 2026-09-12

## Claim

독립 감사자가 실제 SQLite·Git·일반 폴더 fixture를 실행하여 4 AC를 확인했다. 기존 실행 시작·카드 할당·상태 보고 API의 값이 Todo DB에 저장되고, 공개 `LoadPure`에서 카드와 함께 읽힌다. 런타임 상태 보고는 카드 완료 권위로 승격되지 않는다. 실패한 assignment의 암묵 run 및 assignment, 실패한 확장 설치는 부분 기록을 남기지 않았다. 기존 카드 수정·보관·공식 relocation·JSON 재이전에서 데이터가 보존됐다.

**품질 판정은 서로 다른 두 실측 기준을 합친 제한된 재감사다.** B1에서 `kanban` 패키지 전체 테스트 380개가 통과하고 커버리지 86.2%를 측정했다. 그 뒤 B2에서는 남은 lint 16건의 cleanup 명시·테스트 Scan 오류 처리만 바뀌었으며, 정확한 변경 확인과 관련 회귀 119개·race·독립 SQL fixture·린트·vet를 다시 실행했다. B1의 86.2%를 B2에서 새로 측정한 전체 패키지 수치라고 주장하지 않는다.

이 PASS는 순서 예외를 품질 면제로 바꾸지 않는다. 사용자 승인 기록 [staged-verification-approval.md](staged-verification-approval.md)의 대상은 검증 순서뿐이며, 안전성·4 AC·85% 기준·독립 감사는 유지했다. 전체 Todo 통일 23 AC, t648 완료, 자동 `picked → done`, ownership/identity/receipt, Graph, 운영 이전·설치·원격 반영은 이번 판정에 포함하지 않는다. 이전 Factory guard 감사의 64.2%는 다른 패키지의 별도 측정이다.

### 평가 프로파일과 검사 범위

- `.moai/config/sections/harness.yaml:7`: `default_profile: "default"`.
- `.moai/config/evaluator-profiles/default.md:11,22`: Craft 85% 기준. `.moai/config/sections/quality.yaml:5`의 목표 85, `coverage_exemptions.enabled: false`도 확인했다.
- `.claude/rules/moai/workflow/sync-coverage-scope-contract.md:16`: auto 모드의 테스트·커버리지 범위는 변경된 패키지와 필요한 의존 범위다. 따라서 새 파일만 85%라는 수치로 패키지 기준을 대체하지 않았다.
- 같은 계약 `:21–25`: 선택 검사를 전체 저장소 검사로 조용히 확대하지 않는다.
- `.moai/specs/SPEC-TODO-RUNTIME-STORE-001/acceptance.md:20`: 변경 로직 coverage·race·공개 API 오류 전파를 측정한다.
- `.claude/agents/moai/sync-auditor.md:112`: 확인 재감사는 열거한 결함 변경분에 한정하며 처음부터 전체 감사를 반복하지 않는다. B2의 제한된 재검증 근거다. 생산 변화는 cleanup 호출 순서를 유지한 명시적 반환값 무시이고, 테스트 변화는 누락했던 Scan 오류를 실패로 처리하는 강화다. 새로운 기능·저장 transaction·schema 변경은 B1 이후 없다.
- `sync_scope`: 최초 child의 runtime 저장·읽기 연결·이전 보존 및 이번 감사에서 발견한 storage 관련 lint 28건.
- `coverage_scope`: B1 변경 패키지 `internal/kanban` 전체; B2 관련 회귀와 신규·변경 Go coverage block. 전 저장소 검사는 아니다.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS | `ok  \tgithub.com/modu-ai/moai-adk/internal/kanban\t273.761s\tcoverage: 86.2% of statements` (B1); `ok  \tgithub.com/modu-ai/moai-adk/internal/kanban\t46.439s\tcoverage: 47.6% of statements` (B2의 선택 검사) |
| Security (25%) | 100/100 | PASS, 변경 범위 한정 | 미래 버전·퇴역 원본·오류 rollback fixture PASS; `todo_runtime.go:191`의 `WHERE id=? ... id=?`, `:207`의 `VALUES(?,?,?,?,?,?)` 확인; 아래 독립 SQL 결과 원문 |
| Craft (20%) | 100/100 | PASS | B1 `coverage: 86.2% of statements`; B2 `changed-blocks aggregate=121/138=87.68%`; lint `0 issues.` |
| Consistency (15%) | 100/100 | PASS | 최종 `golangci-lint`: `0 issues.`; vet·diffcheck·gofmt stdout 없음, exit 0 |

보안 점수는 로컬 저장 변경의 거절·SQL 입력·transaction 보존 경계를 대상으로 한다. SQL 값은 매개변수로 바인딩하고, 결합되는 conflict SQL은 두 상수 중 하나다. 새 의존성 선언 변경은 없었다. HTTP 인증·XSS·SSRF 등 이번 변경에 없는 표면을 검사했다고 주장하지 않으며, 전 제품 OWASP 인증이나 취약점 DB 대조는 수행하지 않았다. 보안 checklist와 테스트 피라미드 reference는 검사 경계를 정하는 데 사용했으며 수락을 대신하지 않았다.

### AC 결과

| AC | 판정 | 실제 검증 내용 |
|---|---|---|
| AC-TRS-001 | PASS | Git/SPEC 있는 프로젝트의 실제 manifest·provenance·owner·reported_state·event_kind·Git SHA, 일반 폴더/SPEC 없음, implicit run, 중복 없는 최신 upsert, legacy runs/cards 미증가, live/archive membership 보존 |
| AC-TRS-002 | PASS | checkpointed/legacy DB의 반복 공개 JSON 읽기, 빈 runtime 배열, 실제 활성 WAL 읽기 전후 bytes 및 동시 commit 중 card/runtime snapshot 일치 |
| AC-TRS-003 | PASS | 실제 SQLite abort trigger 오류가 호출자에게 반환되고 implicit run/assignment 모두 rollback, BoardLock을 통한 카드 mutation 경쟁, 후속 whole-record write 및 stale DTO가 최신 runtime을 지우지 않음 |
| AC-TRS-004 | PASS | core1 last_seq/card/findings/archive 보존, extension 설치 실패 rollback, 미래 core/extension 및 partial schema·퇴역 writer 거절/bytes 보존, relocation 및 JSON migration runtime 보존, Todo에 Factory 테이블·이전 marker 부재 |

### Findings 및 해소 이력

- **F1 [Medium] [blocking → resolved] [confidence: high]** `todo_runtime.go`와 `todo_runtime_safety_test.go`의 신규 errcheck 12건. 최초 독립 lint가 28건을 보고했고 새 파일 12건은 Close/Rollback 반환 처리였다. 담당자가 cleanup을 명시하되 Scan/Rows.Err/Commit/BoardLock Release 오류 반환은 유지했다. B1 재검사에서 신규 12건이 사라졌다.
- **F2 [Medium] [blocking → resolved] [confidence: high]** `backlog_integrity_audit_test.go`, `backlog_pure_reader_test.go`, `backlog_migrate.go`의 나머지 errcheck 16건. 최초 diff에서 해당 지적 줄은 새 runtime 연결 변경 밖에 있었으나, 기준 HEAD의 lint를 따로 실행하지 않았으므로 오래된 결함이라는 인과는 단정하지 않았다. root가 저장소 안전 회귀에 관련된 정확한 16건으로 수정 범위를 확정했고, B2에서 검사 오류 처리·cleanup만 교정한 뒤 lint 0건을 확인했다.
- 검증 공백 해소: 작성자의 전체 패키지 240초 timeout과 독립 선택 검사 47.0%는 전체 패키지 품질 PASS가 아니었다. B1 전체 패키지를 `-parallel=4 -timeout=600s`로 1회 실행하여 실제 86.2%를 측정했다. 앞선 timeout 원인을 코드·장비·deadlock으로 단정하지 않는다.
- partial stamped schema 및 relocation의 수정 전 RED는 작성자 [runtime-store-implementation.md](runtime-store-implementation.md)에 귀속된다. 독립 감사는 해당 수정 후 fixture를 재실행하고 별도 JSON 이전·copy rollback fixture를 추가 실행했다. 작성자 RED를 감사자 직접 RED로 표시하지 않는다.
- **남은 blocking/optional 제품 결함: 이번 검증에서 발견하지 않음.** 관측하지 않은 위험은 아래 Gaps/Residual-risk로 보존한다.

## Evidence

모든 명령의 cwd는 아래 worktree다. Go 검사는 한 호출에서 다음 환경을 제거했다.

```bash
unset MOAI_HOME MOAI_CONFIG_DIR MOAI_PROJECT_ROOT CLAUDE_PROJECT_DIR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_COMMON_DIR && <command>
```

### B1 — 변경 패키지 전체 테스트와 커버리지

다른 검사 프로세스가 끝난 뒤 실행했다. 앞선 240초 검사에서 결과를 얻지 못한 패키지 분모를 측정하려고 동시 테스트 상한을 4로 정하고 제한을 600초로 변경한 1회 검사다. 전 저장소 검사가 아니고, 변경한 옵션이 앞선 실패 원인을 고쳤다는 주장도 아니다.

```bash
go test ./internal/kanban -count=1 -parallel=4 -v -coverprofile=/tmp/todo-runtime-audit.PNUqKd/package-coverage.out -timeout=600s > /tmp/todo-runtime-audit.PNUqKd/package-tests.log 2>&1
```

관측 원문 마지막 부분:

```text
--- PASS: TestFactoryFreeSlots (0.00s)
    --- PASS: TestFactoryFreeSlots/missing_registry_means_all_free (0.61s)
    --- PASS: TestFactoryFreeSlots/dead_claim_is_pruned_and_reads_free (1.45s)
    --- PASS: TestFactoryFreeSlots/claims_beyond_workers_do_not_widen_the_result (1.44s)
    --- PASS: TestFactoryFreeSlots/live_claim_makes_its_slot_busy (1.30s)
PASS
coverage: 86.2% of statements
ok  	github.com/modu-ai/moai-adk/internal/kanban	273.761s	coverage: 86.2% of statements
```

Exit 0. `rg -c '^--- PASS:' /tmp/todo-runtime-audit.PNUqKd/package-tests.log` 원문: `380`. FAIL/SKIP 검색 결과 행 없음. 전체 stdout은 [package-tests.log](/tmp/todo-runtime-audit.PNUqKd/package-tests.log), profile은 [package-coverage.out](/tmp/todo-runtime-audit.PNUqKd/package-coverage.out)에 보존했다. 이 문서의 출력은 원문 발췌이며 전체 stdout을 모두 싣지는 않았다.

해당 실행 중 runtime/adapter 결과 원문:

```text
--- PASS: TestRecordFactoryRunStartRecordsMetadataWithoutClaimingSpecProvenance (1.25s)
--- PASS: TestTodoRuntimeSafetyAssignmentAbortRollsBackSeed (0.76s)
--- PASS: TestTodoRuntimeSafetyErrorsPropagate (1.06s)
--- PASS: TestTodoRuntimeSafetyCorruptRowsDoNotReadAsEmpty (1.09s)
--- PASS: TestTodoRuntimeSafetyExtensionInstallRollsBack (0.47s)
--- PASS: TestTodoRuntimeSafetyStampedPartialSchemaRefused (0.72s)
--- PASS: TestTodoRuntimeSafetyRetiredAndMissingCardRefuse (0.47s)
--- PASS: TestTodoRuntimeSafetyArchiveAndStaleDTOArePreserved (0.91s)
--- PASS: TestTodoRuntimeSafetyExtensionPreservesLegacyContentAndSorts (1.76s)
--- PASS: TestTodoRuntimeSafetyRelocationPreservesRuntime (0.44s)
--- PASS: TestTodoRuntimeSafetyParityRejectsRuntimeLoss (0.00s)
--- PASS: TestTodoRuntimeSafetyCardMutationAndAssignmentSerialize (0.36s)
--- PASS: TestTodoRuntimeSafetyPureSnapshotReadsActiveWAL (0.71s)
--- PASS: TestTodoRuntimeStorePublicReadbackSurvivesCardEdit (4.02s)
--- PASS: TestTodoRuntimeStoreFutureSchemaPreservesBytes (0.01s)
--- PASS: TestTodoRuntimeStoreByteGuardRejectsSameLengthChange (0.00s)
--- PASS: TestTodoRuntimeStoreNonGitWithoutSpecSeedsRun (0.59s)
--- PASS: TestTodoRuntimeStoreLegacyPureReadReturnsEmptyRuntimeWithoutWriting (0.35s)
--- PASS: TestTodoRuntimeStoreWriterRejectsFutureTodoVersions (1.84s)
--- PASS: TestTodoRuntimeStorePartialSchemaRefusedWithoutMutation (0.71s)
```

### B2 — 마지막 16건 교정 후 관련 회귀와 race

```bash
go test ./internal/kanban -run '(Backlog|Migrate|Quarantine|Parity|TodoRuntime|RecordFactoryRunStart|Pure|Audit)' -count=1 -v -coverprofile=/tmp/todo-runtime-audit.PNUqKd/final-delta-coverage.out -timeout=180s > /tmp/todo-runtime-audit.PNUqKd/final-delta-tests.log 2>&1
```

```text
--- PASS: TestBacklogLock_TimeoutNamesLockPath (1.67s)
--- PASS: TestBacklogLockStuckHolderSurfacesBoundedNamedError (1.66s)
PASS
coverage: 47.6% of statements
ok  	github.com/modu-ai/moai-adk/internal/kanban	46.439s	coverage: 47.6% of statements
```

Exit 0. top-level PASS `119`. 47.6%는 선택 검사에서 패키지 전체 statement를 분모로 집계한 수치다. 전체 패키지 커버리지가 86.2%에서 47.6%로 회귀했다는 뜻이 아니다. 전체 stdout은 [final-delta-tests.log](/tmp/todo-runtime-audit.PNUqKd/final-delta-tests.log)에 보존했다.

```bash
go test -race ./internal/kanban -run '^(TestTodoRuntime|TestRecordFactoryRunStart|TestAuditReadSnapshot)' -count=1 -timeout=180s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/kanban	59.257s
```

Exit 0. 별도 race 검사에서 새 runtime 및 최종 수정과 관련된 snapshot 회귀를 실행했다. 모든 가능한 interleaving의 증명은 아니다.

### B2 — 독립 작성 SQLite fixture

저장소 파일을 수정하지 않고 `/tmp`의 Go overlay로 독립 테스트 두 개를 작성했다. public JSON을 새 큐 파일로 전달하여 실제 `Load()` migration을 수행하고, `copyRuntime`에 중복 run 입력을 주어 실패 후 runtime 테이블·stamp가 남지 않는지 확인했다. 운영 DB는 사용하지 않았다.

```bash
go test -overlay=/tmp/todo-runtime-audit.PNUqKd/overlay.json ./internal/kanban -run '^TestIndependentRuntime' -count=1 -v -timeout=60s
```

```text
=== RUN   TestIndependentRuntimeLegacyJSONRoundTrip
    runtime_audit_overlay_test.go:68: public JSON migration preserved card and runtime values
    runtime_audit_overlay_test.go:69: Todo contains no Factory tables or Factory migration markers
--- PASS: TestIndependentRuntimeLegacyJSONRoundTrip (1.56s)
=== RUN   TestIndependentRuntimeCopyAbortIsAtomic
    runtime_audit_overlay_test.go:98: copy failure rolled back runtime rows, tables, and version stamp
--- PASS: TestIndependentRuntimeCopyAbortIsAtomic (0.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	2.243s
```

Exit 0. [fixture 원문](/tmp/todo-runtime-audit.PNUqKd/runtime_audit_test.go), [overlay](/tmp/todo-runtime-audit.PNUqKd/overlay.json).

### B2 — lint, vet, 형식

```bash
golangci-lint run ./internal/kanban --timeout=2m && go vet ./internal/kanban && git diff --check && gofmt -l internal/kanban/todo_runtime.go internal/kanban/factory_runtime.go internal/kanban/backlog_migrate.go internal/kanban/backlog_integrity_audit_test.go internal/kanban/backlog_pure_reader_test.go
```

```text
0 issues.
```

Exit 0. vet/diffcheck/gofmt는 추가 stdout 없음. 최초 독립 lint 출력은 `28 issues: * errcheck: 28`, B1 중간 출력은 `16 issues: * errcheck: 16`이었다. 중간 16건 원문은 [lint-final.log](/tmp/todo-runtime-audit.PNUqKd/lint-final.log)에 보존했다. 파일명과 달리 이 파일은 교정 전 B1의 로그이며 최종 B2 PASS와 혼동하지 않는다.

### 커버리지 보조 지표 — 변경 연결부를 포함

Go profile과 `git diff --unified=0`를 대조한 [집계 스크립트](/tmp/todo-runtime-audit.PNUqKd/coverage_summary.rb)를 실행했다. 새 runtime 전체와 adapter, backlog 읽기·publication·parity 변경을 모두 포함한다. 이 집계는 **추가·변경 줄과 교차하는 Go coverage block 전체 statement 수**로서, 정확한 변경 줄 커버리지와는 다르다. 패키지 기준을 대체하지 않는다.

```bash
ruby /tmp/todo-runtime-audit.PNUqKd/coverage_summary.rb /tmp/todo-runtime-audit.PNUqKd/final-delta-coverage.out
```

```text
backlog_migrate.go file=281/357=78.71% changed-blocks=16/19=84.21%
backlog_sqlite.go file=118/141=83.69% changed-blocks=2/2=100.00%
backlog_store.go file=249/263=94.68% changed-blocks=0/0=N/A
factory_runtime.go file=23/25=92.00% changed-blocks=5/5=100.00%
todo_runtime.go file=98/112=87.50% changed-blocks=98/112=87.50%
changed-blocks aggregate=121/138=87.68%
Changed-blocks count whole Go coverage blocks intersecting added/modified lines; not exact changed-line coverage.
```

`backlog_store.go`의 변경은 DTO 필드 선언이라 실행 statement 분모가 없다. 신규 runtime만 골라 통과 수치를 만든 것은 아니다. 변경 중 예외 처리의 모든 branch가 실행된 것은 아니다.

## Baseline-attribution

- cwd: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/todo-unified`
- `git rev-parse --short HEAD`: `a315dad9a`
- `git branch --show-current`: `WT-todo-unified`
- `moai session current`: `01a09337-4567-7690-b6a5-dd0c421f3179`
- 감사자는 생산·저장소 테스트·SPEC를 수정하지 않았다. 소유 저장소 산출물은 이 보고서 하나다. 운영 DB·카드 상태·Git branch/commit/push는 변경하지 않았다.

### B1 전체 패키지 검사 기준

HEAD 위 미커밋 소스였다. runtime `3e4d7711c29dbcf4189f2991a94e4132d035c925d80a7fbbe4b77317b35fd8c8`, safety test `2ac89c95f80c45a1c5cdb73b878b8518a598e57e88b2466df8d38a5861fca93e`, backlog_migrate `b988013611a398c5b8d075ed2f3b6802528db22e216ac6e202df97d373482b31`.

원문·profile SHA256:

```text
2e6f8f85c20533b2e7e45597e20265bbe47169f5575343a392af84e7bc152695  /tmp/todo-runtime-audit.PNUqKd/package-tests.log
b5b6114e430bd3b8e66a3c1c7d82259440797d9d47f30f0df5410b98b5560d06  /tmp/todo-runtime-audit.PNUqKd/package-coverage.out
```

### B2 최종 재감사 기준

`shasum -a 256` 실제 출력:

```text
3e4d7711c29dbcf4189f2991a94e4132d035c925d80a7fbbe4b77317b35fd8c8  internal/kanban/todo_runtime.go
2d9e915f70fa170a4d7c97c67dd5d02783ff6ae40befe5ab4f0fe0275d901f85  internal/kanban/factory_runtime.go
2ca44fcdfc5336dbacfdc25dc25bb35eeea4c8ca1d844a73122842f0344b6398  internal/kanban/backlog_store.go
22ca752be7482daa01b06db22da150b16ae6503e963a2423a1c8cd3dcb30e70a  internal/kanban/backlog_sqlite.go
f24692a0309fb13a0fc49320d2802afdb74faaef2e75f30a949e97ce22bb6fdb  internal/kanban/backlog_migrate.go
091feb3e2568abdda5266c5936fab02d26a51c0415bcde76872bb4f107f3dd82  internal/kanban/todo_runtime_store_test.go
2ac89c95f80c45a1c5cdb73b878b8518a598e57e88b2466df8d38a5861fca93e  internal/kanban/todo_runtime_safety_test.go
b4f6c2e39cd1f9cc5c4197d7084fb9d4331ee79826bc13ea1d8c7aa148f35fc6  internal/kanban/factory_runtime_test.go
b7d866eb036ffe459b55556c0cca7359f34f70aed5e35753108b54b3c94bbe82  internal/kanban/backlog_integrity_audit_test.go
744b01b1662722704683046d53fbf891bcc8a1d57de87f279ce1f5e976ca0be0  internal/kanban/backlog_pure_reader_test.go
```

반복 이력: 최초 기능 범위 회귀 110개/47.0%와 race 23.224s 통과 후 lint 28건 발견 → 신규 12건 교정 후 race 25.959s·독립 fixture 통과, lint16 유지 → B1 전체 패키지 380개/86.2% 관측 → 나머지16건 한정 교정 → B2 회귀119개·race59.257s·독립fixture2개·lint/vet/형식 통과. B1→B2의 생산 변경은 backlog_migrate의 세 cleanup 반환 처리이며 transaction 순서·저장 구조는 그대로임을 diff로 확인했다.

## Gaps

- B2에서 전체 패키지를 다시 실행하지 않았다. 결함 변경분 재감사 계약에 따라 B1 전체 측정과 B2의 영향 범위 검사 결과를 함께 사용했다. 별도 CI·전 저장소 테스트·다른 OS·설치 바이너리·운영 DB 검증은 하지 않았다.
- 취약점 데이터베이스에 대한 의존성 대조, 외부 모델 교차 감사는 수행하지 않았다. 선언 의존성 변경 여부와 이번 저장 경계의 코드·실행 증거에 한정된 보안 판정이다.
- migration admission, identity/ownership/fencing, receipt 및 자동 완료의 미구현을 이 최초 child의 결함으로 만들지 않았다. 전체 umbrella 종료 전에 별도 구현·검증해야 하는 범위다.
- `/tmp` 전체 로그와 독립 fixture는 현재 로컬 보존 자료로서 장기 보존을 보장하지 않는다. 중요한 명령·판정·원문 발췌·hash는 이 추적 가능한 보고서에 포함했다.

## Residual-risk

여러 실제 fixture가 통과해도 모든 파일시스템·전원 장애·프로세스 간 interleaving을 증명하지는 않는다. 활성 WAL snapshot과 BoardLock 경쟁은 이번 검사한 스케줄의 증거다. runtime reported_state는 실행자의 보고 값이며 소유권 검증 또는 카드 완료 확정으로 사용하면 안 된다. 새 run·assignment 테이블이 있다는 사실만으로 Factory 전체 저장소 통일이나 기존 과거 기록의 운영 이전이 끝난 것은 아니다. 이 PASS로 운영 데이터 변경·카드 done 조작·설치·push·PR·병합·배포 권한이 추가되지 않는다.
