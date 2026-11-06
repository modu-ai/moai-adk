# t2098 progress — af-flow complete가 release/main-batch로 착지하는 라우트 기본 꺼짐 복원

Tree: worktree `t2098` @ base `9cf8a1fc14`, branch `WT-af-flow-moai`, lane-33, run to2bbk. Class B (run→sync).

## 원인 (확정)

`moai factory complete`가 af-flow에서 local main 대신 `release/main-batch`로 착지. 사슬: `internal/cli/factory_card.go:3142` → `factoryReleaseRoute`(`internal/cli/factory_release_route.go:48-62`) → `config.LoadAgenticTrunkReleaseBranch`(`internal/config/loader_agentic_release_branch.go:33-49`)가 af-flow+`release_branch_prefix`만으로 브랜치를 돌려줌(운영자 의도 미문의). t2086이 config를 af-flow로 전환하며 무장; `release/main-batch`가 `.moai/worktrees/main-batch`에 체크아웃돼 있어 조건 성립. 근거 문서: `.moai/reports/af-flow-v2-20261106/design-diff.md` §2 (주 체크아웃 직독).

## 수정 (dev-t2098 위임, lane 스폰)

- `internal/config/types.go` — `ModeProfile.ReleaseRoute bool` (`yaml:"release_route"`), 기본 꺼짐/부재=꺼짐, 주석에 카드·운영자 규칙·REQ-FCR-026 축소 근거·typed save round-trip 유의 기록.
- `internal/config/loader_agentic_release_branch.go` — `prefix == "" || !profile.ReleaseRoute` → `""` 게이트 (문서 주석 갱신).
- `internal/cli/factory_release_route.go` — 문서 주석만 (옵트인·기본 꺼짐·게이트 위치).
- 시험 4파일: `agentic_trunk_test.go`(fixture에 `release_route: true`), `factory_aff_s2_t2013_test.go`(신규 `TestAFFS2ReleaseRouteRequiresExplicitOptIn`), `af_flow_test.go`, `loader_factory_default_workflow_test.go`(신계약 재작성).

## Evidence (TDD result contract)

- RED (EXPECTED_RED, exit=1):
  - `go test ./internal/config -run 'TestAFFlowLoaderNormalizesLegacy|TestT2056_ReleaseBranchReaderSharesTheDefault' -count=1` → `af_flow_test.go:65: LoadAgenticTrunkReleaseBranch = "release/main-batch", want "" (release_route absent)` ×2 + `loader_factory_default_workflow_test.go:172` 등 4행.
  - `go test ./internal/cli -run 'TestAFFS2ReleaseRouteRequiresExplicitOptIn' -count=1` → `factory_aff_s2_t2013_test.go:579: factoryReleaseRoute ok = true, want false`.
  - 전제 고백(감사관 참고): route-true 케이스는 수정 전 우연 통과(비-stRICT yaml.Unmarshal이 키 무시) — 수정 후에는 옳은 이유로 통과. 초기 2회 실패는 시험 작성 버그(작성자 자가 수정 후 재관측)로 결함 증거 아님.
- GREEN (PASS):
  - `go build ./...` exit 0; gofmt -l 7파일 클린.
  - `go test ./internal/cli -run 'TestAFFS2|TestAgenticTrunk' -count=1 -v` exit 0, 25 --- PASS (296s; AC-FCR-026 착지 전 suite를 옵트인 장착 상태로 통과).
  - `go test ./internal/config -count=1` exit 0 (21.8s, 패키지 전체).

## Gaps

- `internal/cli` 전체 suite 미실행(레인 로컬 규율; `go test ./...` 금지). main-batch/release_branch_prefix bare-prefix af-flow fixture는 grep으로 부재 확인(나머지 적중은 release-flow — 다른 워크플로, 라우트 미무장).
- 착지(complete)는 전 레인 complete 금지 홀드(d-20261106)로 merge-ready에서 정지 — 리더·운영자 FF 착지.

## 결정 기록

- decision record: decided_by=lane-33 orchestrator evidence_refs=.moai/reports/af-flow-v2-20261106/design-diff.md §2(주 체크아웃 직독), 배차문, 운영자 상시 규칙 d-20261106T055816Z-449d ladder_path=disk evidence (S0 설계 확정: 명시 옵트인 키, 로더 게이트 단일 수정점)
- decision record: decided_by=lane-33 orchestrator evidence_refs=dev-t2098 완료 보고 + git diff 스폿 검증(로더 게이트·types 필드 확인) ladder_path=delegate deliverable read (커밋 승인)

## §E.4 Sync-phase Audit-Ready Signal

sync_status: complete
card_review: pass, no findings — `.moai/reports/t2098/card-review.md` (codex mcp codex_review scope=card, base `9cf8a1fc14`, head `f4d697c561`)
run_commit: `f4d697c561` ([t2098] fix(factory): gate the af-flow release route behind release_route opt-in)
sync_commit: pending-backfill — 파일은 자기 커밋 sha를 인용할 수 없으므로 본 sync 커밋의 짧은 sha는 커밋 메시지 본문과 레인 완료 보고에 기록, 후속 커밋에서 backfill (하우스 패턴)
sync_commit_files: CHANGELOG.md, .moai/reports/t2098/progress.md
spec_md_status: n/a — Class B 카드, SPEC 산출물 없음
