---
id: SPEC-GTD-AUTONOMY-001
created: 2026-09-15
updated: 2026-09-15
---

# SPEC-GTD-AUTONOMY-001 실행 작업

manager-develop 한 명이 아래 순서로 TDD를 수행한다. 각 task는 acceptance.md의 정확한 test selector가 RED인지 확인하고, 최소 구현으로 GREEN을 만든 뒤 영향 package를 재검증한다.

## M1 — 명칭·호환

- [x] `internal/cli/gtd.go`에 `NewGTDCommand`를 만들고 todo와 동일 constructor/handler를 공유한다.
- [x] `internal/cli/gtd_compat_test.go`에 19 verb×flag×exit/stdout/stderr/state matrix를 구현한다.
- [x] command/workflow/agent skill source와 template mirror에서 GTD canonical, todo thin compatibility를 구성한다.
- [x] `TestGTDAllTodoVerbsParity`, `TestGTDCanonicalSurfaceGolden`의 nonempty PASS를 확보한다.

## M2 — 저장·왕복

- [x] `internal/kanban/backlog_gtd_schema.go`에 기존 schema version과 분리된 additive migration을 작성한다.
- [x] old/new, backup/restore, export/import, WAL/vacuum, archive/reopen matrix를 격리 `MOAI_HOME` fixture로 작성한다.
- [x] `TestMigrateGTDSchemaIdempotent`, `TestGTDLegacyRoundTrip`을 PASS시킨다.

## M3 — GTD 절차

- [x] `gtd_capture.go`, `gtd_clarify.go`, `gtd_organize.go`, `gtd_reflect.go`, `gtd_engage.go`를 절차별로 작성한다.
- [x] allowed/denied/duplicate/stale/cancelled/dry-run fixture를 각 exact selector에 연결한다.

## M4 — 관계·private graph

- [x] `gtd_relation.go`에서 신규 관계와 기존 네 관계의 방향·bytes·archive/restore를 함께 검증한다.
- [x] `internal/graph/gtd_private.go`에 DB sibling projection, 0700/0600, logical revision, atomic publish를 구현한다.
- [x] repo/template/Git/log/telemetry/export/default-backup 비유출, opt-in backup, revoke/delete negative fixture를 구현한다.

## M5 — 위임·정책

- [x] `internal/mission/contract.go`, `policy.go`, `receipt.go`에 sealed contract, deterministic validation, stable operation ID를 구현한다.
- [x] scope/stale/evidence/input escalation mutant를 exact selector로 검출한다.

## M6 — `--auto`·governor

- [x] 기존 constructor 경계를 보존한 `internal/cli/goal.go`에서 mission natural language를 shell/condition parser와 분리한다.
- [x] `mission_mode=auto`와 기존 progression mode를 별도 저장한다.
- [x] `.claude/agents/moai/mission-governor.md`와 template mirror를 비쓰기 permission으로 추가한다.
- [x] approval 뒤 scope 내 무질문, scope 밖 blocked, auditor/policy 우선 테스트를 구현한다.

## M7 — MissionRuntime·복구

- [x] `internal/mission/runtime.go`에 full/partial/unsupported capability probe와 active-session-only 전이를 구현한다.
- [x] reconnect 불가, credential expiry, replace, lease loss/takeover, same snapshot resume를 테스트한다.
- [x] `internal/mission/operation.go`에서 crash cut readback/reconciliation을 구현한다.

## M8 — E2E·repo-local delivery

- [x] Capture→Clarify→Organize→seal→publish→pick→dispatch 허용/거절/crash/concurrent E2E를 작성한다.
- [x] local develop 기준 launcher-entered `WT-gtd-autonomy`에서 explicit-path commit만 만들고 lane push/card PR을 금지한다.
- [x] manager-git 전용 local develop integration lease와 `--no-ff` merge를 구현·검증한다.
- [x] lead batch push→`origin_develop_sha` CI→release 분기→main release PR→`main_landed_sha` ancestry를 검증한다.
- [x] owner/lease/SHA drift 및 각 side effect 직후 crash mutant를 검출한다.

## M9 — 생성물·최종 회귀

- [ ] 4개 언어 docs/navigation/help/completion과 template emitted output을 동기화한다.
- [x] 23개 release-blocking selector의 `=== RUN` count가 각각 1 이상이고 empty-sweep token이 없음을 확인한다.
- [x] AC-GTD-001·004 regression guards와 일반 goal/Kanban/Factory 동작을 확인한다.
- [ ] sync audit와 `origin_develop_sha`/release PR 현재 SHA CI evidence를 progress.md에 기록한다.

## Sync audit iteration 2 — production trust boundary 보정

- [x] governor decision과 독립 audit의 `0600` contained receipt를 mission/contract/snapshot/action/targets/expiry/issuer/HEAD/status에 결속하고 모든 unsafe·stale·FAIL·tamper 변이를 차단한다.
- [x] `--recommend`를 비권한 호환 flag로 제한하고 production boolean governance 우회를 제거한다.
- [x] bounded supervisor를 snapshot→governance→validate→owner→readback 순서로 영속 실행하고 blocked·completed replay에서 추가 effect와 질문을 금지한다.
- [x] publish→pick→leased disk dispatch를 실제 SQLite operation receipt와 runtime assignment에 연결하고 두 store·crash/restart exactly-once를 검증한다.
- [x] batch push→release branch→release PR→main merge capability owner의 SHA/gate/readback을 검증하고 미구성 production provider를 stable `provider_unsupported`로 차단한다.
- [x] archive/reopen/cancel/stale relation 및 v2 governance export/import·v1 compatibility·tamper 회귀를 검증한다.
- [x] source/template goal workflow 계약, catalog whole-tree hash, CGO0·Windows build, scoped vet, changed-scope lint, gofmt, diff-check를 재검증한다.

## Sync audit iteration 3 — semantic evidence·topology·completion 보정

- [x] action별 authoritative evidence value predicate를 추가하고 false/0/whitespace/wrong SHA/stale mutant를 effect 전에 차단한다.
- [x] supervised Git plan을 `WT-*` card worktree commit과 별도 develop worktree leased no-ff merge로 분리하고 두 경로를 snapshot lineage에 결속한다.
- [x] action 소진을 completion으로 간주하지 않고 sealed `0600` completion receipt와 landed ancestry를 필수화한다.
- [x] 실제 임시 linked worktree 두 개에서 commit→blocked/restart→merge→blocked/restart→completion과 foreign dirty 보존을 검증한다.
- [x] source/template workflow, catalog hash, target selector, CGO0·Windows build, scoped vet, gofmt/diff-check를 재검증한다.
- [ ] repository-wide CI와 개별 CLI fault-path 85% coverage 판정은 integration branch CI가 소유한다.

## Sync audit iteration 4 — 함수별 coverage·private account 검증

- [x] `ValidateAutoMissionIntegrity`, `OrganizeGTDItemWithRelations`, supervisor/dispatch/operation/NewCommand와 private projection의 함수별 coverage를 85% 이상으로 측정한다.
- [x] 실제 kernel owner/mode/current UID readback과 injected mismatched requester/owner/mode denial을 검증한다.
- [x] repo/template/Git/log/telemetry/export/default-backup collector 7개 surface에서 private projection bytes/path/content hit 0건을 확인한다.
- [x] explicit opt-in backup만 data/meta를 포함하고 publication 단계별 crash fault가 partial pair를 남기지 않음을 검증한다.
- [ ] repository-wide CI 최종 판정은 integration branch CI가 소유한다.
