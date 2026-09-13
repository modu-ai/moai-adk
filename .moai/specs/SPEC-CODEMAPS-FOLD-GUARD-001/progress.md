# progress.md — SPEC-CODEMAPS-FOLD-GUARD-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-14
- plan-audit iter1 PASS 0.86 (Tier S threshold 0.75, must-pass 7/7) — findings D1-D6 applied; report at `.moai/reports/t748/plan-audit.md`
- tier: S
- artifacts: spec.md, plan.md, progress.md (Tier S 2 + progress)
- baseline_tree: 146faed9d (`.claude/worktrees/t748`, plan-phase 저작 시점)

## §E.2 Run-phase Evidence

### Pre-flight (plan §C, run start, tree f600a1c5d)

- `git branch --show-current` → `WT-codemaps-fold-ac`; `git rev-parse --short HEAD` → `f600a1c5d` (plan 저작 시점 146faed9d 대비 plan 커밋 1개 흡수 상태 — SPEC artifacts 자체)
- `ls internal/graph/codemaps_fold_guard_test.go` → `No such file or directory`, EXIT=1 — AC-CFG-001 RED-now 재현
- fold 토큰 grep(5문서 × 5토큰, `-F`) → 무출력, EXIT=1 — 히트 0 (§A.2 일치)
- `grep -c '^fold ' fold-judgments.txt` → `5`, EXIT=0 — floor 5행 (§A.2 일치)
- depends_on: REFRESH-002 `status: completed` / STAMP-ANCESTRY-001 `status: completed`
- `go test ./internal/graph/` → `ok ... 77.345s` (기존 baseline 녹색)

### AC Matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CFG-001 가드 존재 + 현재 트리 녹색 | PASS | `go test ./internal/graph/ -run 'TestCodemapsFold' -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph` (m2_restored_pass.log, exit 0) — RED-now는 pre-flight `ls` EXIT=1로 관측 |
| AC-CFG-002 변조 → FAIL, 복원 → PASS | PASS | `go test ./internal/graph/ -run 'TestZZRedProbeAInjectedToken' -v` (transient probe) + `TestCodemapsFoldGuardFixtures/injected_fold_token_fails_naming_unit_and_document` | RED: `--- FAIL ... fold unit "internal/kanban/prlink_landedref.go" appears in modules.md:329`; restore: `--- PASS: TestCodemapsFoldGuardFixtures` |
| AC-CFG-003 기록 floor | PASS | `TestZZRedProbeBLostFloorLine` (RED) + `TestCodemapsFoldGuardFixtures/lost_floor_line_fails_naming_the_unit` | RED: `--- FAIL ... record lost floor unit line(s): internal/hook/quality/step_git_env.go`; restore: PASS |
| AC-CFG-004 스탬프 독립 | PASS | /tmp fixture 스캔 전체 (provenance.json 부재 상태에서 PASS) + `git diff --stat HEAD~1..HEAD` | fixture PASS (스탬프 파일 없이 판정 성립 — 미독의 구조적 증명); diff 범위 = spec.md + guard 테스트 2파일뿐, checker/워크플로 변경 0 |
| AC-CFG-005 비침묵 도달성 | PASS | `go test ./internal/graph/ -count=1` (기본 스위트 경로) | `ok ... 46.335s`, exit 0 — 가드가 별도 워크플로 편집 없이 기본 스위트에 편입; 모든 실패 경로가 `t.Fatal`/`t.Errorf`로 표면화 |

### Evidence files (this tree, `.moai/state/verify/t748/`)

- `m1_green_guard.log` — M1 가드 테스트 최초 GREEN (`-run 'TestCodemapsFold' -v`, exit 0)
- `m2_red_probes.log` — RED 3건 verbatim (`--- FAIL` × 3 + restore `--- PASS`) — probe 파일은 미커밋 transient, 증거 수출 후 삭제
- `m2_restored_pass.log` — 복원 후 커밋된 fixture 테스트 전수 PASS
- `m3_full_pkg_test.log` — 패키지 전체 스위트 `ok` 46.335s
- `m3_lint.log` — `golangci-lint run ./internal/graph/...` → `0 issues.`

### E-item summary

- **E2 빌드**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- **E3 테스트**: `go test ./internal/graph/ -count=1` → `ok ... 46.335s` (변경 영향 패키지 한정 — 전체 스위트는 CI 몫, lead 조건)
- **E5 린트**: `golangci-lint run ./internal/graph/...` → `0 issues.` (신규 이슈 0)
- **E7 blocker**: 없음
- **E8 RED verbatim**: `m2_red_probes.log` — 3개 fixture 변조 각각 `--- FAIL` + 가드 메시지(단위·문서 명명), 무변조 restore `--- PASS`. RED 없이 GREEN만 있는 상태 없음.

### 동결 표면 (REQ-CFG-005)

`git status --porcelain -- .moai/project/codemaps/ .moai/reports/t475/ .moai/reports/t747/` → 무출력 (변조 0). RED 입증은 전부 `/tmp` 사본(t.TempDir) 위에서 수행.

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-09-14
- run_commit_sha: "ffc872ae5" (backfilled — M2 증거 커밋; 본 필드를 운반하는 M3 커밋은 자기 SHA를 알 수 없어 후속 커밋에서 backfill)
- run_status: complete
- ac_pass_count: 5
- ac_fail_count: 0
- preserve_list_post_run_count: 3 (.moai/project/codemaps/**, .moai/reports/t475/**, .moai/reports/t747/** — 전부 변조 0)
- l44_pre_commit_fetch: n/a (카드 워크트리 — 통합 병합과 push는 리드 소관, gitflow-lane-protocol §4)
- l44_post_push_fetch: n/a (동일 — 레인 push 금지)
- new_warnings_or_lints_introduced: 0
- cross_platform_build.linux: pass (go build ./... exit 0)
- cross_platform_build.windows: pass (GOOS=windows GOARCH=amd64 go build ./... exit 0)
- total_run_phase_files: 2 (internal/graph/codemaps_fold_guard_test.go 신설 + spec.md frontmatter status 전환)
- m1_to_mN_commit_strategy: M1 가드 신설(128474b19) → M2 RED 입증 + progress 갱신 → M3 verdict/backfill

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
