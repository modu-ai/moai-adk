# progress.md — SPEC-CI-PR-TRIGGER-FILTER-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-29
card: t294
worktree: .claude/worktrees/t294 (branch WT-freshness-trigger @ b2370b2bf, origin/develop ee50984ab absorbed)
tier: S (spec.md + plan.md; AC inline in spec.md §3)
kickoff_open_decisions: "(none — 2026-08-30 운영자가 두 축 모두 닫음)"
kickoff_resolved_decisions: "축 A=좁은 범위(graph-freshness.yml 단독) · 축 B=별도 카드로 연기(concurrency 미변경 → 이중 발화 비배달, spec.md §0)"

## §E.2 Run-phase Evidence

측정 트리: `.claude/worktrees/t294`, 브랜치 `WT-freshness-trigger`, 편집 전 HEAD `b2370b2bf`.
비교 기준점 `origin/develop` = `ab897947a` (커밋 직전 `git fetch origin develop` 이후 재측정).
`origin/develop`은 분기 이후 10커밋 전진했으나 `git log HEAD..origin/develop -- .github/workflows/`
출력이 비어 있어 워크플로 디렉터리는 미접촉 — AC-PTF-003b의 이동 ref 비교가 이 카드의 변경만 잰다.

### AC 판정 매트릭스

| AC | 판정 | 검증 명령 | 관측 출력 (편집 후) | RED baseline (편집 전, 같은 트리) |
|---|---|---|---|---|
| AC-PTF-001 | PASS | `yq -o=json '.on.pull_request.branches' .github/workflows/graph-freshness.yml` | `[\n  "main",\n  "develop"\n]` (rc=0) | `null` (rc=0) — 필터 부재 |
| AC-PTF-002 | PASS | `grep -n -B6 '^  pull_request:' .github/workflows/graph-freshness.yml` | L4-7 주석 4줄 노출. 작성 시점 등가(`main`이 유일한 base, `6786c3fa4`)와 git-flow 전환(`11216d13f`)이 그 문장을 넓혔다는 사실을 모두 명명 | 주석 0줄 (출력은 L1-4뿐) |
| AC-PTF-003 (a) | PASS | `yq -o=json '.on.push' .github/workflows/graph-freshness.yml` | `{"branches":["main","develop"]}` (yq pretty 출력, rc=0) — 편집 전후 바이트 동일 | 동일 값 |
| AC-PTF-003 (b) | PASS | `git diff --stat origin/develop -- .github/workflows/` | ` .github/workflows/graph-freshness.yml \| 5 +++++` / `1 file changed, 5 insertions(+)` — 1파일. 전체 diff의 유일한 hunk가 `@@ -1,7 +1,12 @@`(`on:` 블록). `jobs:`/`permissions:`/`concurrency:` hunk 0건 | 무출력 (변경 0) |
| AC-PTF-004 | PASS | `yq '.concurrency.group' …` · `yq '.concurrency.cancel-in-progress' …` | `${{ github.workflow }}-${{ github.ref }}` · `true` — 편집 전과 동일 | 동일 값 |
| AC-PTF-005 | PASS | `actionlint .github/workflows/graph-freshness.yml` | rc=0, stdout 무출력 | rc=0, stdout 무출력 |
| AC-PTF-006 | PASS | AC-PTF-006이 명명한 `grep -c` 검사 (세기는 표식을 파일 안에서 1회로 요구하므로 이 셀은 표식을 직접 적지 않는다 — 유일한 등장은 아래 유예 관측 표) | `1` — 유예 관측 행 1건. 실행 기반 관측을 수행했다고 주장하는 §E 행은 0건 | `0` (rc=1) |

### 유예 관측 (실행 기반 — 미수행)

| 표식 | 언제 | 실행할 명령 | 그 판독이 확인할 것 |
|---|---|---|---|
| DEFERRED-OBS | `develop` 동결 해제 후, head가 `develop`인 PR이 실제로 열린 뒤 | `gh run list --workflow=graph-freshness.yml --limit 30 --json event,headSha,headBranch` | 같은 `headSha`에 `event`가 `push`와 `pull_request` 둘로 남는지 — 남는다면 이중 발화는 이 카드가 닫지 않았다는 것(spec.md §0·§1.3의 비배달 선언)이 실행으로도 확인된다 |

이 관측은 **수행되지 않았다.** `develop`이 동결돼 head가 `develop`인 PR을 열 수 없고, 카드 브랜치
push는 CI-inert다(조사 §5). 위 AC 판정은 전부 파서·린터 단언이며, 실행 발화 집합에 대한 주장이 아니다.

### 미검증 / 잔여 위험

- **실행 기반 회귀 관측 없음** — 위 유예 항목. 이 변경의 판정 주체는 `develop` 통합 후의 원격 CI다.
- **도구 의존** — `yq`(v4, `/opt/homebrew/bin/yq`)와 `actionlint`(`/Users/goos/go/bin/actionlint`)는 레인
  로컬 PATH에만 존재한다. CI 게이트로 승격하지 않았다.
- **이동 ref 비교** — AC-PTF-003b는 `origin/develop`(비고정 ref)에 대고 잰다. 위에 그 시점 SHA
  `ab897947a`를 못박았고, 미접촉 사실을 `git log HEAD..origin/develop -- .github/workflows/`로 확인했다.
- **템플릿 미러 없음(재확인 완료)** — `ls internal/template/templates/.github/workflows/` → `label-sync.yml`
  1개뿐. `graph-freshness.yml`의 템플릿 미러는 존재하지 않으므로 Template-First 후속이 없다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-30
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 6          # AC-PTF-001~006 (003은 a·b 두 셀 모두 PASS)
ac_fail_count: 0
deferred_observation_count: 1   # 실행 기반 발화 관측 — develop 동결로 수행 불가
preserve_list_post_run_count: 0 # plan.md §A.5 PRESERVE 위반 0건
l44_pre_commit_fetch: "git fetch origin develop → origin/develop=ab897947a; git rev-list --count --left-right origin/develop...HEAD → 10 2 (카드 브랜치가 분기 이후 전진한 정상 상태, 통합은 별도 창)"
l44_post_push_fetch: "N/A — 이 카드는 push하지 않는다(레인 지시)"
new_warnings_or_lints_introduced: 0   # actionlint rc=0 무출력, baseline과 동일
cross_platform_build: "N/A — Go 코드 미변경(워크플로 YAML 1파일). 이 diff가 영향 줄 수 있는 Go 패키지 없음"
total_run_phase_files: 1              # .github/workflows/graph-freshness.yml
spec_artifact_files_touched: 2        # spec.md(frontmatter status/updated만) + progress.md(§E.2·§E.3)
m1_to_mN_commit_strategy: "단일 커밋 — M1(필터+주석)·M2(보존 확인)·M3(유예 등록)이 1파일 5줄 변경과 그 증거이므로 분할하지 않음"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
