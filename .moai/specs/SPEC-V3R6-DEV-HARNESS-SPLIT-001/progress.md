# Progress — SPEC-V3R6-DEV-HARNESS-SPLIT-001

> §E 섹션 스켈레톤 (plan-phase: §E.1만 채움; §E.2~§E.4는 placeholder).
> §E.2/§E.3 = run-phase (manager-develop). §E.4 = sync-phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

- **SPEC ID**: SPEC-V3R6-DEV-HARNESS-SPLIT-001 — decomposition: SPEC ✓ | V3R6 ✓ | DEV ✓ | HARNESS ✓ | SPLIT ✓ | 001 ✓ → PASS
- **Tier**: S (minimal)
- **Status**: in-progress
- **Artifacts**: spec.md + plan.md + acceptance.md + progress.md (4 plan-phase 파일)
- **Frontmatter**: 12 required fields + `tier: S` + `depends_on: [SPEC-V3R6-DEV-HARNESS-CONSOLIDATION-001]`
- **REQ count**: 7 (REQ-DHS-001..007)
- **AC count**: 10 (AC-DHS-001a/b, 002a/b, 003a/b, 004, 005, 006, 007)
- **Exclusions**: §E에 5개 `### Out of Scope — <topic>` H3 + `-` bullet (OutOfScopeRule 충족)
- **Supersession-of-decision**: CONSOLIDATION-001의 통합-진입 *결정만* 번복; CONSOLIDATION-001은 completed 유지 (superseded 표시 금지)
- **Runner asymmetry 설계**: release-update만 Runner+manifest; github/release는 specialist 직접 라우팅 (no Runner/manifest)

_<plan-phase 완료. 다음: plan-auditor 감사 → 구현 착수 승인 → run-phase>_

## §E.2 Run-phase Evidence

10개 AC binary PASS/FAIL matrix (각 acceptance.md §D 검증 명령 실행 결과):

| AC | REQ | Status | Verification Command | Actual Output |
|----|-----|--------|----------------------|---------------|
| AC-DHS-001a | REQ-DHS-001 | PASS | `test -f .claude/commands/harness/{release-update,github,release}.md` | `PASS` (3 thin commands 존재) |
| AC-DHS-001b | REQ-DHS-001 | PASS | `grep -E '^argument-hint:.*since/issues\|pr/hotfix\|VERSION'` | `[--since vX.Y.Z \| --dry]` / `issues\|pr [...]` / `[VERSION] [--hotfix]` |
| AC-DHS-002a | REQ-DHS-002 | PASS | `test -f harness-{release-update,github,release}-specialist.md` + `find harness-devkit-*` | 새 이름 3개 존재; `harness-devkit-*` find 결과 EMPTY |
| AC-DHS-002b | REQ-DHS-002 | PASS | `grep '^name: harness-X-specialist'` + active `/harness:devkit` grep (formerly 제외) | 3개 NAME_OK; 활성 `/harness:devkit` 라우팅 EMPTY (Migration Provenance historical refs는 `formerly` 토큰으로 제외) |
| AC-DHS-003a | REQ-DHS-003 | PASS | `test ! -f devkit.md && test ! -f manifest.json` | `PASS` (통합 devkit.md + manifest.json 부재) |
| AC-DHS-003b | REQ-DHS-003 | PASS | release-update-run.js 존재 + devkit-run.js 부재 + release-update/manifest.json 존재 + github/release Runner 부재 | `PASS` (Runner 비대칭 — release-update만 Runner+manifest) |
| AC-DHS-004 | REQ-DHS-004 | PASS | `go test -run SplitHarnessNamespaceNoLeak` + `grep devkit *namespace_test.go` + RENAME check | RED 입증(planted `harness-release-update` leak → FAIL, sentinel `SPLIT_HARNESS_NAMESPACE_LEAK`) → 제거 → GREEN(`ok`); grep devkit EMPTY; RENAME_OK |
| AC-DHS-005 | REQ-DHS-005 | PASS-WITH-DEBT | `find internal/template/templates` 3 patterns | commands/harness EMPTY; 3 split agents EMPTY; `.claude/workflows/harness-*` Runner dir EMPTY. (find3 broad pattern `*workflows/harness-*`는 pre-existing `skills/moai/workflows/harness-builder.md`를 false-match — SPLIT-001 leak 아님, origin/main tracked) |
| AC-DHS-006 | REQ-DHS-006 | PASS | 5 surface active devkit grep (migration/formerly 제외) | 5개 surface 전부 EMPTY; 3개 split 이름 언급 존재 (각 surface 1+) |
| AC-DHS-007 | REQ-DHS-007 | PASS | `grep -rl 'DEV-HARNESS-SPLIT-001' memory/` | EMPTY (run-phase 메모리 미작성 — sync-phase manager-docs 책임) |

**Invariant 검증**:
- 기존 template test suite 무회귀: `go test ./internal/template/... -count=1` → `ok` (전체 PASS)
- Layer B harness specialist 4개 (cli-template/hook-ci/quality/workflow) 무변경 (touch 금지 준수)
- CONSOLIDATION-001 closed artifacts 무변경 (completed 유지, superseded 표시 안 함)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-22
run_commit_sha: de13ecc4c
run_status: implemented
ac_pass_count: 9
ac_pass_with_debt_count: 1   # AC-DHS-005 (broad find3 pattern pre-existing false-match — not a SPLIT-001 leak)
ac_fail_count: 0
preserve_list_post_run_count: 0   # Layer B 4 specialists + CONSOLIDATION-001 artifacts 무변경
new_warnings_or_lints_introduced: 0   # golangci-lint ./internal/template/... → 0 issues
cross_platform_build:
  host: pass     # go build ./... → exit 0
  windows: pass  # GOOS=windows GOARCH=amd64 go build ./... → exit 0
total_run_phase_files: 17   # 3 specialist rename + 1 Runner rename + 1 manifest move + 3 thin commands + 1 devkit rm + 1 test rename/rewrite + 5 doctrine + spec.md + progress.md
m1_to_mN_commit_strategy: single-combined   # Tier S — M1-M6 + status transition in one commit
red_green_evidence: "RED: planted harness-release-update leak → FAIL (sentinel SPLIT_HARNESS_NAMESPACE_LEAK); GREEN: leak removed + rebuild → ok"
```

**Residual-risk (잔여 위험)**:
- AC-DHS-005 find3 broad pattern (`*workflows/harness-*`)는 의도상 `.claude/workflows/harness-*` Runner 디렉터리를 겨냥하나 `.claude/skills/moai/workflows/harness-builder.md`도 매칭. 이는 origin/main에 이미 tracked된 v4 Builder 스킬로 SPLIT-001 leak이 아님. 정확한 Runner-leak 검증(`.claude/workflows/harness-{release-update,github,release}-*`)은 EMPTY.
- run_commit_sha placeholder는 commit 직후 backfill 필요.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs>_
