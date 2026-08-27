# Progress — SPEC-CLI-TEST-CWD-ISOLATION-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-08-28
- Artifacts: spec.md / plan.md / acceptance.md (**Tier M, 3 artifacts** — rationale in
  spec.md §C: measured scope is S-shaped; M stands on the canonical 3-file artifact set,
  the iter-1-FAIL retry-slot reality, and the guard+per-test verification ceremony;
  downgrade path documented for independent audit judgment) + this progress skeleton.
- Budget: REQ 5 / AC 5 — Tier M ceilings (16 / 16) respected.
- Evidence base: **measured RED reproducer** (lead session, this worktree, base `d34a789a4`,
  2026-08-28; probe record `.moai/reports/t334/red-probe.md`): env-scrubbed
  `go test ./internal/cli -run 'Kanban|Factory' -count=1` → rc=0, 0.869 s, recreates
  `state/todo/leads.json` + `state/factory/workers.json`. Negative probes: internal/config,
  internal/kanban, `Config|Cache|Settings` — all clean (config-cache reclassified
  historical; internal/hook not probed, no claim). Base-tree fact: 4 committed `.moai` dirs
  on `d34a789a4` → tree judgments are baseline deltas, never emptiness. Plus t317 §G-1
  (`0ad4b52ba`) and primary-checkout residue re-verified 2026-08-28.
- **plan-audit iter1 (2026-08-28): FAIL 0.775 vs Tier M 0.80** — report
  `.moai/reports/t334/plan-audit-iter1.md`. v0.2.1 revision = D1 (baseline-delta scan) +
  D2 (tier M restored, acceptance.md retained) + D3 (attribution aligned, red-probe.md
  cited) + D4 (internal/hook struck) + D5/D7 (wording). iter2 delta re-audit pending.

## §E.2 Run-phase Evidence

### M1 — RED re-established (frozen reproducer, attribution triple)

- **Tree**: HEAD `9114ddbc3` (branch `WT-cli-test-cwd`, base `origin/develop` `d34a789a4`), worktree
  `.claude/worktrees/t334`, run-phase session, 2026-08-28.
- **Pre-run state**: `internal/cli/.moai` absent at pre-flight (plan-phase probe residue already
  gone); baseline scan `/tmp/t334-moai-pre.txt` = exactly the 4 committed dirs
  (`internal/harness/router/testdata/{keyword-force,normal,spec-overrides}/.moai` +
  `internal/template/templates/.moai`).
- **Command** (frozen reproducer, verbatim incl. env scrub):

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test ./internal/cli -run 'Kanban|Factory' -count=1
  ```

- **Verbatim output**: `ok  \tgithub.com/modu-ai/moai-adk/internal/cli\t0.913s` (rc=0)
- **RED judgment (post-run)**:
  - `find internal/cli/.moai -type f` → `internal/cli/.moai/state/todo/leads.json` +
    `internal/cli/.moai/state/factory/workers.json` (≥1 file; name-agnostic AC-001 satisfied)
  - baseline delta: `diff /tmp/t334-moai-pre.txt /tmp/t334-moai-post.txt` → `0a1 >
    ./internal/cli/.moai` (exit 1 = delta dirty)
- **RED for the right reason**: residue created by THIS run from the recorded clean baseline
  (4-line scan re-verified before the run); post-run `rm -rf internal/cli/.moai` restored the
  scan to baseline equality (diff exit 0).
- **Pre-flight companions**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...`
  exit 0; `golangci-lint run --timeout=2m` → `0 issues.` (lint baseline).

### M1 — per-test identification (REQ-4; selector bisect, each run env-scrubbed, `-count=1`)

`go test ./internal/cli -list 'Kanban|Factory'` → 55 tests. Every test/group was run with
`internal/cli/.moai` removed immediately before; producing status judged by
`find internal/cli/.moai -type f` after each run. Result — **5 producing tests**, all others
clean (union = exactly the reproducer's file set):

| # | Test | File:line | Residue produced | Root resolution hit |
|---|------|-----------|------------------|---------------------|
| 1 | `TestCC_FactoryEntryThroughRunCC` | `internal/cli/factory_test.go:747` | `state/todo/leads.json` + `state/factory/workers.json` | `launchProjectRoot()` → `resolveProjectDir()` → cwd (P2+P3) |
| 2 | `TestCC_KanbanEnvMutationIsRestored` | `internal/cli/cc_test.go:513` | `state/todo/leads.json` | same (P3) |
| 3 | `TestCC_KanbanFlagStrippedBeforeLaunch` | `internal/cli/cc_test.go:369` | `state/todo/leads.json` | same (P3) |
| 4 | `TestGLM_FactoryWorkerEntry` | `internal/cli/factory_test.go:829` | `state/factory/workers.json` | same (P2) |
| 5 | `TestGLM_KanbanFlagParity` | `internal/cli/glm_test.go:543` | `state/todo/leads.json` | same (P3) |

- **Mechanism note (why the existing seam does not catch them)**: all 5 already stub
  `findProjectRootFn` → `t.TempDir()` (`installFactoryLaunchSeam` / inline), but the
  name-claim call sites (`appendLeadName`, `resolveFactoryWorkerName`,
  `resolveCompanionName`) resolve their root via `launchProjectRoot()` → `resolveProjectDir()`
  (`internal/cli/session.go:264`) = `$CLAUDE_PROJECT_DIR` or `os.Getwd()` — a DIFFERENT
  resolver that bypasses the stubbed seam. Under the scrubbed env, root = the package cwd
  (`internal/cli`), so the registries land in-tree.
- **Verified clean** (no residue individually or in group): `TestCC_KanbanWritesNoStateRecord`,
  `TestLauncherWritesNoKanbanRecord`, `TestGLMTaskFactoryModeIgnoresModelOverride`, all 11
  `TestEnter*`, the 11-test cobra-family group (`TestCG_*`, `TestReject*`,
  `TestACFM022a/023c_*`, `TestNewMCPCmd_FactoryChildren`, `TestFactoryGenealogyInHelp`,
  `TestMoAIHuhTheme_*`, `TestRunLinkSpec_FactoryError`), and the 24-test pure-logic group
  (`TestParse*`, `TestResolve*`, `TestPrepareKanbanSettings*`, tier/blockcap/constant tests).
- **Frozen per-test selector list (D7 freeze)** — REQ-4 is judged by running each of these
  individually post-fix:
  `TestCC_FactoryEntryThroughRunCC`, `TestCC_KanbanEnvMutationIsRestored`,
  `TestCC_KanbanFlagStrippedBeforeLaunch`, `TestGLM_FactoryWorkerEntry`,
  `TestGLM_KanbanFlagParity`.

### M2 — residue guard implemented + guard RED observed (AC-002)

- **Change**: `internal/cli/main_test.go` TestMain extended with the residue guard
  (pre-`m.Run()` existence snapshot of the entry-cwd `.moai` path, post-`m.Run()` delta check,
  stderr failure naming the detected absolute path, exit-code override to 1). Guard judgment is
  directory-existence (never a file list — the `kanban/`→`todo/` drift); locus pinned from the
  entry cwd so a mid-run chdir cannot move it; delta semantics (pre-existing residue is not this
  run's delta). `go vet ./internal/cli/` exit 0; `gofmt -l internal/cli/main_test.go` empty.
- **Command** (guard RED observation; unfixed tree, same env scrub; the TestMain-anchored guard
  rides every selector, so the frozen selector IS the `<guard>`-augmented form):

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test ./internal/cli -run 'Kanban|Factory' -count=1
  ```

- **Verbatim output** (exit code 1):

  ```
  PASS
  RESIDUE GUARD FAIL: this test run created /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t334/internal/cli/.moai — internal/cli tests must not write .moai state into the package working directory (SPEC-CLI-TEST-CWD-ISOLATION-001 REQ-1/REQ-2/REQ-3). Isolate the producing test's project root (see the SPEC's mechanism ladder), then remove the directory so the guard re-arms for the next run.
  FAIL	github.com/modu-ai/moai-adk/internal/cli	0.895s
  FAIL
  ```

- **Mutant-probe counter-evidence (§D.2)**: the inner tests print `PASS` — the FAIL is the
  guard's alone, and the message names the actually-detected residue path (not a path that can
  never exist), so the guard is not vacuously passing.

### M3 — isolation fixes applied, targeted GREEN observed (AC-003 + REQ-4)

- **Change** (D4 ladder rung 1 — explicit root injection through the resolver's documented env
  parameter, the `TestCC_KanbanWritesNoStateRecord` / t161 medicine; test-only diff): each of
  the 5 producing tests gained

  ```go
  t.Setenv(config.EnvClaudeProjectDir, t.TempDir())
  ```

  at test entry (parent scope for the 4-subtest `TestCC_FactoryEntryThroughRunCC`). This
  redirects `launchProjectRoot() → resolveProjectDir()` — the resolver the name-claim call
  sites actually use — into the test-owned sandbox. None of the 5 tests call `t.Parallel()`
  (B2 audited). Files: `internal/cli/cc_test.go` (2 tests), `internal/cli/glm_test.go` (1),
  `internal/cli/factory_test.go` (2). `go vet ./internal/cli/` exit 0; `gofmt -l` empty on all
  four touched files.
- **AC-003 command** (frozen reproducer, verbatim, fixed tree, HEAD post-M2 `65e7e7672` +
  unstaged M3 edits):

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test ./internal/cli -run 'Kanban|Factory' -count=1
  ```

- **Verbatim output**: `ok  \tgithub.com/modu-ai/moai-adk/internal/cli\t1.004s` (rc=0; guard
  green — a residue would have failed the run per M2)
- **AC-003 judgment**: `test ! -e internal/cli/.moai` → succeeds; baseline delta
  (`diff /tmp/t334-moai-pre.txt <post-scan>`) → empty (exit 0). Pre-run state verified clean +
  baseline unchanged immediately before the run.
- **REQ-4 (frozen per-test selectors, each run individually, env-scrubbed, `-count=1`)**:
  `TestCC_FactoryEntryThroughRunCC` → ok 0.744s clean; `TestCC_KanbanEnvMutationIsRestored` →
  ok 0.725s clean; `TestCC_KanbanFlagStrippedBeforeLaunch` → ok 0.733s clean;
  `TestGLM_FactoryWorkerEntry` → ok 0.717s clean; `TestGLM_KanbanFlagParity` → ok 0.724s
  clean. Cumulative post-5-runs baseline delta → empty. Each run's guard rode the selector
  (TestMain) and stayed green.

### M4 — package-wide verification + wrap-up (AC-004 + AC-005)

- **AC-004 command** (full package, serial, single attempt, Bash timeout 600 s, env-scrubbed,
  fixed tree at M3 commit `f5ff4d74e`):

  ```
  unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED CLAUDE_PROJECT_DIR && go test ./internal/cli -count=1
  ```

- **Verbatim output**: `ok  \tgithub.com/modu-ai/moai-adk/internal/cli\t396.090s` (rc=0 —
  guard green: with the TestMain guard in place, any package-cwd `.moai` creation would have
  failed this run)
- **AC-004 judgment**: `test ! -e internal/cli/.moai` → succeeds; baseline delta → empty
  (exit 0). Pre-run state verified clean + baseline unchanged immediately before the run.
- **AC-005 enumeration** (`git diff --name-only d34a789a4..HEAD`):
  - Go files: `internal/cli/main_test.go`, `internal/cli/cc_test.go`,
    `internal/cli/factory_test.go`, `internal/cli/glm_test.go` — **all `*_test.go`; zero
    non-test Go changes; no production default-path behavior change; no env-guarded seam
    needed** (ladder rung 1 sufficed for all five tests).
  - Non-test files: the 4 SPEC artifacts (spec/plan/acceptance/progress) — documentation only.
  - Touched package: `internal/cli` only → full suite green per AC-004.
  - PRESERVE honored: no consumer-side file in the diff (`glm.go` walk, `state_dir.go`
    semantics, hook walkers untouched); `.gitignore` untouched; no template/CHANGELOG change.
- **Quality gates**: `golangci-lint run --timeout=2m` → `0 issues.` (baseline `0 issues.`,
  delta 0 new); `go vet ./internal/cli/` exit 0; `gofmt -l` empty on all touched files;
  `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0;
  `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` exit 0 (windows test-file compilation).
- **Evidence carry-over**: `red-probe.md`, `plan-audit-iter1.md`, `plan-audit-iter2.md`
  copied worktree → primary `.moai/reports/t334/` with `cmp` byte-verification
  (worktree-gitignored-artifact lesson).
- **Card-id convention**: `Card: t334` present in every commit body on the branch (3/3:
  `9114ddbc3` plan, `65e7e7672` M1-M2, `f5ff4d74e` M3; this M4 evidence commit is the 4th).

### Run-phase context note (user-directed, outside SPEC AC scope)

2026-08-28, mid-M1: the operator instructed immediate removal of ALL residue `.moai`
directories ("더이상 묻지말고 모두 .moai만 제거해라"). Removed: worktree `internal/cli/.moai`
(bisect residue) AND the primary checkout's stale `internal/cli/.moai` (the §D.6
forward-looking item — `state/config-cache.json`, `state/kanban/leads.json`,
`state/factory/workers.json`, Aug-13/20 file set), plus 40+ additional untracked stray
`.moai` directories discovered by a primary-wide scan (`internal/{permission,statusline,web}`,
`docs-site/**`, `.github*`, `.moai/specs/*` nested strays, `.omp`, etc. — all verified
untracked: `git ls-files | grep '/\.moai/'` returns content only under the 4 committed
roots). Post-cleanup scan: exactly the 4 tracked `.moai` roots remain outside backup
archives (`.moai-backups/`, `.moai/backups/`, `docs-site.hextra.backup/` deliberately
preserved — archived snapshots, not live residue). No tracked file touched; no source
change; does not affect the worktree baseline. Code-level isolation for the non-cli
packages (`internal/{permission,statusline,web}` residue regeneration) remains out of
scope per spec.md §B.
2026-08-28 (sync-phase): the primary-checkout residue removal was operator-confirmed through the
lane session (AskUserQuestion round) — the §D.6 post-merge cleanup item is discharged at the
source, not awaiting the lead.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-28
run_commit_sha: f5ff4d74e        # M3 implementation head; M4 evidence commit rides after this block
run_status: complete
ac_pass_count: 5                 # AC-001..AC-005 all PASS with recorded evidence (§E.2)
ac_fail_count: 0
preserve_list_post_run_count: 0 violations   # no consumer-side/gitignore/template file in diff
l44_pre_commit_fetch: pending-backfill        # recorded at the M4 evidence commit (below)
l44_post_push_fetch: not-applicable           # no push from this card worktree; integration into develop is lead-owned per gitflow lane protocol
new_warnings_or_lints_introduced: 0           # baseline "0 issues." → post "0 issues."
cross_platform_build:
  darwin: pass
  windows: pass
  windows_test_compile: pass
total_run_phase_files: 6                      # 4 internal/cli *_test.go + spec.md frontmatter + progress.md
m1_to_mN_commit_strategy: per-milestone commits — M1+M2 combined (RED evidence + residue guard, carries draft→in-progress), M3 isolation fixes, M4 evidence
```

**Pending integration (lead-owned, §D.5 closure gate)**: branch `WT-cli-test-cwd` (4 commits,
base `d34a789a4`) awaits merge into local `develop` via the integration worktree per the
gitflow lane protocol; CI on `origin/develop` is the full-suite verdict surface. The primary
checkout's stale `internal/cli/.moai` (§D.6 forward-looking item) was ALREADY removed during
run-phase by direct operator instruction — see the run-phase context note in §E.2.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-28
sync_commit_sha: pending-backfill-SPEC-CLI-TEST-CWD-ISOLATION-001
sync_status: complete
b12_self_test_a: pass    # grep -c 'SPEC-CLI-TEST-CWD-ISOLATION-001' CHANGELOG.md → 0 pre-emission (no prior entry)
b12_self_test_b: pass    # AC-ID census of acceptance.md → 5; §E.2 matrix / §E.3 ac_pass_count → 5/5
b12_self_test_c: pass    # every path named in the entry resolves via ls (spec.md + 4 test files, 5/5)
changelog_entry_position: "[Unreleased] / ### Fixed"
frontmatter_status_transitions:
  spec.md: in-progress → completed
  plan.md: n/a (markdown-header convention — no YAML frontmatter)
  acceptance.md: n/a (markdown-header convention — no YAML frontmatter)
  progress.md: n/a (markdown-header convention — no YAML frontmatter)
canary_compliance_check: not-applicable   # this SPEC declares no forward-looking policy its own sync tests
```

`sync_commit_sha` carries the `pending-backfill-*` placeholder per `spec-frontmatter-schema.md`
§ SHA placeholder backfill exemption (D3) — a commit cannot know its own hash until after it
lands — and is backfilled in the immediately following commit, the two-commit pattern every
sibling close on `origin/develop` uses (e.g. `4b36b00d6` → backfill `4fdbd55c1`, t340).

### 이 sync 커밋이 담은 것

마크다운 3종이다. `internal/` 아래는 한 파일도 건드리지 않았다.

| 산출물 | 내용 |
|---|---|
| `CHANGELOG.md` `[Unreleased] / Fixed` | 1 항목. 운영자 관점 서술 — 테스트 스위트가 리포지터리 트리 안에 `.moai/` 상태 디렉터리를 남기던 결함과 그 수정(5개 테스트 루트 고정 + TestMain 잔여 가드). REQ/AC 식별자는 본문에 AC 계수만 인용했다 |
| `spec.md` 프론트매터 | `status: in-progress → completed` (3-phase close — 종결 전이는 이 sync 커밋에 병합). `updated:` 는 이미 `2026-08-28`(= sync 커밋 날짜)이라 무변경 |
| `progress.md` | §E.2 운영자-확인 1행 추가 + 이 §E.4 블록 |

docs-site / README 는 건드리지 않았다. 이 변경은 테스트 전용 내부 수정이다 — 새 CLI 동사도,
새 설정 키도, 문서화된 플래그도, 사용자에게 보이는 동작 변경도 없다. docs-site 페이지가
필요 없다는 이 판단 자체가 이 블록에 기록된다 (배차 지시 4번).

### 관측된 검증

| 항목 | 명령 | 관측된 출력 |
|---|---|---|
| CHANGELOG 중복 없음 | `grep -c 'SPEC-CLI-TEST-CWD-ISOLATION-001' CHANGELOG.md` | `0` (사전 배출 시점) — 기존 항목 없음, 병렬 세션 중복 아님 |
| AC 계수 일치 | `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md \| sort -u \| wc -l` | `5` (AC-001..AC-005). §E.3 `ac_pass_count: 5` 와 일치 |
| 인용 경로 실재 | `ls` (spec.md + `internal/cli/{main,cc,factory,glm}_test.go`) | 전부 존재 (5/5) |
| run 커밋 조상 확인 | `git merge-base --is-ancestor f5ff4d74e HEAD` | rc `0` — §E.3 의 `run_commit_sha` 가 이 트리의 조상이다 |

`go build` / `go test` 는 돌리지 않았다. 이 커밋에 코드 변경이 없으므로 스위트를 태울 근거가
없고(배차 제약 — 문서 검증만 수행), 전 패키지 판정은 `origin/develop` CI의 몫이다.

### 미검증 / 잔여

- **§D.5 종결 게이트의 세 번째 항목은 리드 소관으로 열려 있다**: `WT-cli-test-cwd` 브랜치의
  `develop` 통합 + `origin/develop` CI 초록(전체 스위트 판정 면). 통합은 배차된 대로 리드가
  소유하며 이 세션은 push하지 않는다.
- **§D.6 CI 관측 사이클 1회** (darwin × windows 에서 guard 요주의 관찰) — 머지 후 실제 CI
  몫이며 sync 범위가 아니다.
- **§D.6 primary 체크아웃 잔여 제거는 run-phase 에 이미 시행됐다** (운영자 직접 지시, §E.2
  context note) — sync-phase 에 운영자 확인(2026-08-28, 리드 세션 AskUserQuestion)이 추가로
  기록됐다.
- `sync_commit_sha` 는 이 커밋 시점에 placeholder 상태다. 후속 백필 커밋이 실제 SHA를 채운다.
- **프론트매터 전이는 `spec.md` 한 곳뿐이다.** 배차 지시는 네 산출물 전부를 지목했으나, 이
  SPEC 은 `plan.md` / `acceptance.md` / `progress.md` 를 마크다운 헤더 규약으로 작성했다 — 세
  파일에 YAML 프론트매터 블록이 애초에 없다. 없는 블록을 새로 만드는 것은 manager-docs 에게
  금지된 본문 수정이므로 하지 않았다 (SPEC-CLI-STATE-DIR-BOUND-001 과 같은 마감 형태).
