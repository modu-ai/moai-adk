# plan.md — SPEC-FACTORY-COMPLETION-RECOVERY-001

## §A — Context

- 트리: `.moai/worktrees/t1538`, 브랜치 `WT-t1538-factory-recovery`, base = local develop `a158b4b5f`(plan-scope 노트 `.moai/reports/t1538/plan-scope.md`).
- 근거: `.moai/reports/dispatch-audit-2026-10-06-codex.md`(코드 기준 main@ec13872f3 — 본 트리 base보다 앞서며 t1513 수리 `de388878e`·t1528 `ec13872f3` 미포함).
- Tier M — 산출물 3종(spec/plan/acceptance) + progress.md. 요구 16건·기준 16건(천장 준수).
- 소관 패키지: `internal/cli`(완료 게이트·auto-done·발급 표시), `internal/factory`(facts 구조체·GTD 스키마·relation), `internal/homestate`(reaper·전이 edge·CI reader·watchdog 기록).
- 앵커는 spec.md §B.2에서 본 트리 기준으로 전수 재확인됐다(14개 중 13개 일치, 1개 부분 이동 — todo_analysis.go:74는 기록 분기이고 임계 상수는 `internal/factory/backlog_analysis.go:33`).

## §B — Known issues in the current implementation

1. `todo done`·auto-done 완료 경로에 리더 증거검토 입력 없음 — `internal/cli/todo.go:1104`(`rec.ArchiveCard(id)`), `internal/cli/todo_autodone.go:311`(`AutoDoneFacts` 조립, receipt 필드 없음 — 구조체 `internal/factory/autodone_scan.go:108`).
2. auto-done 잠금 재검증이 ID 동등만 — `internal/cli/todo_autodone.go:385`.
3. 만료 임대 reconcile이 active run만 — `internal/homestate/factory_run_retire.go:333`. 단 전체 분류 원시 `ClassifyRuns`(status="")가 이미 `:161-163`에 존재. `retireRun`은 runs만 UPDATE(`:393`).
4. 사라진 owner의 재배정 edge 부재 — 만료 시 owner 유지(`card_transition.go:351-353`), 임대는 동일 owner만(`:451`), homestate 전체에 reassign 개념 0건(grep exit 1 관측).
5. 구동 중 행의 신호 대기 기록 부재 — watchdog/waiting_since/review_deadline 개념 0건(비테스트 homestate grep 0 관측). 무소유자 행은 t1513 소관(재범위 완료).
6. T19/T20 예약 edge 미구현 — 거부 `card_transition.go:267`, 술어 `:173`(`pushed→ci-green`, `ci-green→done`). 판정문 admission 선례는 `card_evidence_readers.go:93-207`(`audited_sha` 바인딩).
7. near-dup 발급 무표시 — 기록만(`internal/cli/todo_analysis.go:74`, 임계 `backlog_analysis.go:33`=0.80, 적용 `:188`), stdout은 id+위치뿐(`internal/cli/todo.go:851`).
8. 관계 처분 ledger 부재 — `gtd_relations` 스키마에 disposition 없음(`internal/factory/backlog_gtd_schema.go:47`), pair upsert만(`internal/factory/gtd_relation.go:154`). item disposition 쓰기(`gtd_clarify.go:86`)는 존재 — 건드리지 않는다.

## §C — Pre-flight

```bash
# 1. 브랜치·기준
git branch --show-current && git rev-parse --short HEAD   # WT-t1538-factory-recovery / a158b4b5f

# 2. 빌드·교차 플랫폼
go build ./... && GOOS=windows GOARCH=amd64 go build ./...

# 3. 소관 패키지 기준선 (신규 결함과 기존 baseline 구분)
# pipefail 필수: `| tail -5` 단독은 테스트 FAIL을 tail의 exit 0으로 가린다(게이트 실측 — FAIL 출력+pipeline_exit=0).
set -o pipefail; go test ./internal/cli -run '^(TestAutoDone.*|TestTodo.*)$' 2>&1 | tail -5
set -o pipefail; go test ./internal/homestate -run '^(TestFR_.*|TestLease.*|TestTransition.*|TestReconcile.*)$' 2>&1 | tail -5

# 4. t1513 충돌 사전 확인 — 슬롯 술어 관련 기존 테스트 목록
grep -rln 'factorySerialSlotHeld\|SerialSlot' internal/cli internal/homestate --include='*_test.go'

# 5. 개명된 receipt 명칭 충돌 확인 — 스토어 receipt(todo.go:818)과 구별
grep -rn 'LeaderApproval\|ApprovalReceipt' internal/cli internal/factory internal/homestate || echo "no conflicts"
```

## §D — Constraints (PRESERVE)

- **PRESERVE `internal/cli/factory_card.go` 슬롯 술어** — t1513 소관. 본 SPEC은 항목 (5)에 필요한 범위 밖에서 그 의미론을 다시 쓰지 않는다. 접촉은 읽기와 테스트 참조까지.
- **PRESERVE `internal/factory/gtd_clarify.go`** — item disposition 쓰기 경로(:86) 불변.
- **PRESERVE stdout 규율** — `todo add` stdout은 bare "id position" machine line(todo.go:816-824 주석, :851) 불변. near-dup 표시는 stderr로만.
- **PRESERVE retireRun 보존 의미론** — 회수 경로는 행 삭제 금지(factory_run_retire.go:393 UPDATE-only, :382-385 주석).
- **PRESERVE `applyLeaseExpiry` 의미론** — reaper는 기존 회수 의미론(assigned 복귀·mid-merge blocked·worktree 미접촉)을 재사용하고 새 의미론을 만들지 않는다.
- **PRESERVE 레인 경계** — receipt 발급·reaper 호출·reassign은 리더/운영자 경로. lane이 호출 가능한 CLI 표면을 만들지 않는다.
- t1533 파일 접촉 최소화 — hub 대기 결함 4건의 파일은 만지지 않는다.
- 금지: `--no-verify`, `git add -A`, 병합 창 밖 develop 조작, 강제 push.

## §E — Self-verification

manager-develop-prompt-template.md §E 형식 — E1 AC 매트릭스(acceptance.md 16건), E2 `go build ./...`+GOOS=windows, E3 `go test -cover ./internal/cli/... ./internal/homestate/... ./internal/factory/...`(85% 패키지 기준), E4 서브에이전트 경계 grep(해당 없음 — CLI 패키지이나 AskUserQuestion 신규 호출 금지), E5 lint(신규 vs baseline 구분), E6 커밋 SHA·push 상태, E8 RED 실패 출력(TDD — 각 AC의 pre-GREEN 실패 출력 인용).

## §F — Milestones

### M1 — 완료 게이트: receipt 검증 + auto-done 재검증 (항목 1·2, REQ-FCR-001..005)

파일별 변경:

| 파일 | 변경 |
|---|---|
| `internal/homestate/leader_approval.go` (신규) | 리더 승인 receipt 타입 + 바인딩 검증 함수(카드 UUID·**run id**·factory version·증거 해시 네 결합, REQ-FCR-001/005). 발급자 표식으로 리더 경로 구별. **homestate 배치 이유**: `factory → homestate` 의존이 기존 방향(`go list -deps ./internal/factory` 실측, `TestHomestateDoesNotImportFactory` 통과)이라 factory 배치 시 `homestate → factory → homestate` 순환 — cli와 homestate 양쪽이 이미 import하는 homestate에 둔다 |
| `internal/factory/autodone_scan.go` | `AutoDoneFacts`(:108)에 receipt 상태 필드 추가 — 기존 필드 의미론 불변 |
| `internal/cli/todo.go` | done 경로 :1104 `rec.ArchiveCard(id)` 앞에 receipt 검증 삽입(REQ-FCR-002). 거부는 stderr. **수동 done에도 auto-done과 동일한 직전 재검증·직렬화** — 네 바인딩을 archive 직전에 다시 검증하고 동시 factory 전이와 직렬화한다(스캔/발급 시점 검증만으로는 사이 전이가 version을 올려 stale receipt가 통과한다) |
| `internal/cli/todo_auto.go` | --auto 사이클의 archive 지점(:331 `r.ArchiveCard(card.ID)`)에 동일 receipt 게이트 — **세 번째 완료 표면**(REQ-FCR-002a). REQ-THS-012의 picked 긍정 나열 가드(:329-333)는 불변 |
| `internal/cli/todo_autodone.go` | :311 facts 조립에 receipt 반영(REQ-FCR-003); :385 재검증을 스냅샷 UUID·본문·state·SPEC·landing 전면 비교로 확장(REQ-FCR-004) — 그리고 **archive 직전(같은 lock 안)에 receipt 네 바인딩을 최종 재검증하고 factory 전이와 직렬화**한다: 스캔 승인 뒤 factory 전이로 version만 증가하면 5항목 비교는 통과하면서 stale receipt로 닫히는 경합을 차단한다 |
| `internal/homestate/card_transition.go` | 완료 전이 receipt 게이트 — T20(`ci-green→done`)의 예약 edge admission을 receipt 검증으로 구현(REQ-FCR-010), T18(`merged-local→done`, `guardNoRemote` 포함)에 동일 게이트 적용(REQ-FCR-002b). reserved edge 현행 거부 증거: `TestFR_AC019_ReservedCIEdgesRefused`. **기존 시험 수정 대상 — AC018 계열 전체(3건, 패키지 2곳)**: `TestFR_AC018_PushGateStore`(internal/homestate/fr_evidence_test.go:191, no-remote 절 :225-242 — "no remote — no CI verdict" 노트로 receipt 없는 T18 성공 단언), `TestFR_AC018_DecidePushGate`(internal/cli/factory_card_test.go:456, :475-481 — decide push gate로 receipt 없는 no-remote T18→done 단언), `TestFR_AC018_PushGateNeverFetches`(internal/cli/factory_push_nofetch_test.go:20 — 같은 push gate 경로, no-fetch 단언과 receipt 요구의 상호작용을 구현자가 판정). REQ-FCR-002(b) 하에서 세 시험의 기대는 반전/갱신된다(receipt 없으면 거부, 올바른 receipt면 성공). overlay 실측: as-is PASS / receipt 게이트 하 FAIL — 두 방향 기록 |
| 테스트 | `internal/cli/factory_card_test.go`의 `fcFixture`(:31)·`fcPlace`(:89) 스타일 + `todo_autodone_test.go` 표준 — `TestLeaderReceiptGate*`, `TestAutoDoneRecheckStaleRow` |

순서: homestate receipt 타입 → cli 게이트 → homestate 전이 게이트(T20·T18) → auto-done 직전 재검증 → recheck. M1 단독 커밋.

### M2 — 회수: reaper + reassign + watchdog + CI reader (항목 3·4·5·6, REQ-FCR-006..011)

파일별 변경:

| 파일 | 변경 |
|---|---|
| `internal/homestate/factory_run_retire.go` | 전체 run 만료 임대 reaper — `ClassifyRuns`(:161-163) 재사용, 하나의 transaction 안에서 만료 재검증 후 `applyLeaseExpiry` 의미론으로 회수(REQ-FCR-006/007). 행 삭제·생존자 선택 없음 |
| `internal/homestate/card_transition.go` | 운영자 경로 reassign edge — 기존 lease 없음 + 기존 owner 종료 증거 확인, version-checked(`:364` 패턴)(REQ-FCR-008). `:353`·`:451` 가드는 불변 — 새 edge가 그 옆에 추가된다 |
| `internal/homestate/card_record.go` (+스키마 마이그레이션) | 구동 중 행의 대기 기록 — waiting signal + review deadline 필드(REQ-FCR-009). `LeaseExpired`(:137-138) 의미론 불변. 기존 행은 "미판정" 기본값. **기존 정체 행 도달**: watchdog이 행을 처음 관측하는 시점(리더 유지관리 경로)에 `waiting_since`를 소급 기록해 이미 정체 중인 행도 검토 기한이 성숙한다 — 진입 백필 없이는 기존 행이 영원히 대상이 되지 못한다 |
| `internal/homestate/card_evidence_readers.go` (또는 인접 신규 파일) | CI 완료 reader — `audited_sha` 바인딩 선례(:93-207)를 따라 정확한 push SHA의 CI 증거만 인정, **T19만 연다**(REQ-FCR-010/011). T20은 reader 대상 밖 — receipt 게이트(M1)가 연다 |
| `internal/homestate/card_transition.go` (T17) + 스키마 | **pushed tip SHA 영속화** — T17(`merged-local→pushed`)은 현행 MergeSHA·remote ref만 기록해 일괄 push에서 카드 MergeSHA≠push tip이 되면 reader가 요구하는 "push 시점의 정확한 SHA"를 복원할 수 없다(overlay 실측: T17 성공 뒤 tip이 카드 행·이벤트 어디에도 없음). T17에 pushed tip SHA 기록을 추가하고, reader는 **저장된 tip SHA만** 인정 — push 뒤 remote ref가 이동해도 저장값이 변하지 않는다(REQ-FCR-011) |
| 리더 유지관리 경로 | reaper 호출 지점 — 리더/운영자 표면에만(REQ-FCR-014) |
| 테스트 | `fr_fixture_test.go` 패밀리(`frPlace` :136, `frLeaseUntil` :197) + `fr_transition_test.go`의 `frFixtureCard`(:58) + `factory_lease_reconcile_test.go`·`factory_run_retire_test.go` 확장 — `TestExpiredLeaseReaperAllRuns`, `TestOperatorReassignEdge`, `TestStalledWaitWatchdog`, `TestCICompletionReader` |

순서: reaper(기존 의미론 재사용, 최저 리스크) → reassign edge → watchdog 기록 → CI reader. 마이그레이션은 M2 첫 커밋에 포함.

### M3 — 표시·처분: near-dup 발급 표시 + 관계 disposition ledger (항목 7·8, REQ-FCR-012..013)

파일별 변경:

| 파일 | 변경 |
|---|---|
| `internal/cli/todo.go` | 발급 경로 :851 앞뒤에서 stderr near-dup 표시(관련 카드 id·점수·판정 안내) — stdout machine line 불변(REQ-FCR-012) |
| `internal/cli/todo_analysis.go` | :74 `BacklogMatchNear` 분기의 기록과 표시 연결 — 기록 의미론 불변 |
| `internal/factory/backlog_gtd_schema.go` | `gtd_relations` 스키마(:47)에 disposition·rationale 열 추가(마이그레이션) — gtd_items(:36) 불변 |
| `internal/factory/gtd_relation.go` | upsert(:154)가 disposition을 함께 영속화(REQ-FCR-013) — pair 키(subject, object, kind) 불변 |
| CLI 표면 | 관계 처분 기록의 리더/운영자 입력 경로 — 레인 쓰기 없음 |
| 테스트 | `todo_analysis_test.go`·`backlog_analysis_test.go` 스타일 — `TestTodoAddNearDupDisclosure`, `TestGTDRelationDisposition` |

## §G — Anti-patterns to avoid

- `go test -run '^<새 테스트 이름>$'`가 `[no tests to run]`으로 exit 0 — 미측정을 통과로 읽는 것(verification-completeness §1.1). 기준 판정은 스윕 카운트 확인을 포함한다.
- 만료 재검증을 transaction 밖에서 하고 lock 안에서 결과만 적용 — 재검증은 transaction 안에서(REQ-FCR-015).
- reaper를 run 선택용 reconciler(`ReconcileActiveRuns`)에 붙이는 것 — 감사 리포트가 명시한 금지다. 리더 유지관리 경로에 둔다.
- watchdog을 자동 완료·자동 해제로 구현 — 리더 재판정 요구까지다(REQ-FCR-009/016).
- CI reader를 "브랜치 최신 CI"로 구현 — 정확한 push SHA만(REQ-FCR-011).
- 슬롯 술어(factory_card.go) 우회 재정의 — t1513 소관.
- stdout에 near-dup 표시 — stderr 전용(REQ-FCR-012).

## §H — Integration notes

- **통합 경로(현행 계약 준수)**: 본 카드에는 통합 경로 변경의 운영자 승인 지시가 없다 — 레인은 카드 브랜치를 push하지 않고 PR을 만들지 않는다(AGENTS.local.md:223의 레인 WT 브랜치 push 금지·AGENTS.md:154-159의 develop 직렬 창 요구·`git-strategy.yaml` `manual.workflow: git-flow`가 현행 계약). 완료 시 레인은 브랜치 head·증거 경로를 리더 보고로 전달하고, 통합은 **리더가 당시 현행 절차**(설정상 develop 직렬 창, 또는 전환 승인이 선행된 경우 그에 맞는 경로)로 수행한다. 참고(정보, 승인 아님): t1453 PR #1751·t1513 PR #1758은 각 카드의 리더 배차가 명시적으로 main PR을 지시했던 개별 사례다 — 본 카드에는 그런 지시가 없으므로 선례만으로 계약을 대체하지 않는다. 설정 키(git-flow→github-flow) 전환은 운영자 후속 결정으로 리더 보고에 등록한다.
- **merge-base 측정 의무**: "이 카드가 무엇을 바꿨는가"는 `CARD_BASE=$(git merge-base develop HEAD)`로 재구한 뒤 판정 — 리터럴 base SHA 핀 금지(흡수가 리터럴 핀 범위를 오염시킨다). **ref는 실제 흡수한 ref다**: §H의 통합 경로가 develop 직렬 창(현행 계약)이므로 흡수 대상은 로컬 develop — origin/main 기준은 develop이 흡수한 다른 카드 변경까지 본 카드 기여로 계산하는 오류를 낸다(게이트 격리 재현: origin/main 기준 `foreign.go, own.md` vs develop 기준 `own.md`).
- **t1513 술어 재조정 확인(병합 창 필수 항목)**: main 흡수로 `de388878e`(t1513 슬롯 술어 수리)가 본 트리에 들어오면 `factorySerialSlotHeld` 주변 테스트가 충돌할 수 있다. 병합 직후 §C.4의 grep 목록 테스트를 돌리고, 술어 본체와 본 카드 변경의 간섭 여부를 판독한다. 본 카드는 술어 의미론을 쓰지 않으므로 충돌 시 원칙은 **t1513 착지분이 이긴다** — 본 카드의 watchdog(REQ-FCR-009)이 무소유자 행을 다시 만지는 형태로 해결하지 않는다.
- **앵커 재확인**: 흡수 후 run-phase 첫 커밋 전에 spec.md §B.2 표의 앵커를 흡수 트리에서 재확인 — 행 번호 이동 시 본 SPEC을 갱신 없이는 진행하지 않는다(manager-spec 재위임).

## §I — Cross-references

- `.moai/reports/t1538/plan-scope.md` — 스코프 노트(항목 (5) 재범위·t1513 경계·SPEC-ID).
- `.moai/reports/dispatch-audit-2026-10-06-codex.md` — 근거 감사(main@ec13872f3).
- 카드 t1533 — 제외된 후속 결함(hub 대기 4건). 카드 t1513(`de388878e`) — 슬롯 술어 수리, 무소유자 행 소멸. 카드 t1522 — 영수증 경로 선례.
- `internal/homestate/factory_run_retire.go:150-158` — `retirable` 긍정형 게이트(행 보존 원칙의 출처로만 인용; **OwnerDead 생존 조건은 run retirement 전용 — reaper는 상속하지 않는다**, REQ-FCR-007).
