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

### 커밋 df4466d12의 파일별 분리 (sync-phase 정정, 2026-08-30)

리드는 이 커밋을 쪼개지 않기로 정했지만(단일 커밋 유지), 두 축이 섞여 있다는 사실은 파일
목록으로 분리해 적어 둔다 — `run_commit_sha`가 이 한 커밋을 가리키는 것과, 그 커밋이 실제로
두 종류의 변경을 담고 있다는 것은 별개다.

`git show --name-status df4466d12`:

```
M	.github/workflows/graph-freshness.yml
M	.moai/specs/SPEC-CI-PR-TRIGGER-FILTER-001/plan.md
M	.moai/specs/SPEC-CI-PR-TRIGGER-FILTER-001/progress.md
M	.moai/specs/SPEC-CI-PR-TRIGGER-FILTER-001/spec.md
```

- **plan-phase 축** — `spec.md`(§0 비배달 선언·§1.6 범위주장 정정·§1.7 미판정 관측 추가, 버전
  0.1.0 → 0.2.0) + `plan.md`(M0 결정 축 종결). 두 파일 모두 kickoff 세션에서 작성됐으나 **그
  세션이 커밋하지 않고 넘겼다** — 이 run-phase 세션이 그 미커밋 편집을 물려받은 상태로 시작했다.
- **run-phase 축** — `.github/workflows/graph-freshness.yml` (5줄 삽입, hunk 1개, `on:` 블록
  안). 이것이 이 SPEC의 유일한 실제 구현 변경이다.
- `progress.md`는 두 축을 다 반영한다(§E.1이 kickoff 결정을, §E.2/§E.3가 run-phase 증거를 기록).

**원인을 그대로 적는다**: kickoff 세션이 plan-phase 산출물(spec.md 0.2.0 개정 + plan.md M0
종결)을 로컬에 편집만 하고 커밋하지 않은 채 run-phase로 넘어갔다. 이 run-phase 에이전트는
그 미커밋 상태를 물려받았고, SPEC 상태를 `draft → in-progress`로 전이시키는 이 커밋에서 그
미커밋 편집을 함께 실었다 — SPEC 본문을 상태 전이 도중 미커밋 상태로 남겨 두는 대신, 이미
지어진 이 커밋에 얹는 쪽을 택했다. 리드가 이후 이 커밋을 파일 목록별로 쪼개지 않기로
확인했으므로(커밋 이력을 재작성하지 않는다는 sync-phase 지시와 일치), 여기 적는 분리는
독자가 두 축을 구분해 읽을 수 있게 하는 기록이지 커밋을 나누자는 제안이 아니다.

### moving-ref 핀 재검증 (sync-phase, 2026-08-30)

§E.2 상단에 못박은 `origin/develop` 값(`ab897947a`, run-phase 측정 시점)이 sync-phase 시점에도
여전히 이 카드의 변경만 잰다는 것을 재확인한다 — `verification-completeness.md` §4가 요구하는
재측정이다.

```console
$ git fetch origin develop
# From https://github.com/modu-ai/moai-adk
#  * branch                develop    -> FETCH_HEAD
$ git rev-parse origin/develop
68ecbfe4aa542edc1be09137e151251cebc2c58c
$ git log HEAD..origin/develop -- .github/workflows/
# (무출력, rc=0)
```

`origin/develop`은 run-phase 측정 이후 `ab897947a` → `68ecbfe4aa5`로 다시 전진했다(사이 커밋
있음). 그럼에도 `.github/workflows/` 디렉터리를 건드린 커밋은 여전히 0건 — 이 핀은 sync-phase
시점에도 유효하다. 값이 바뀌었다는 사실 자체를 적어 두는 것은 `verification-completeness.md`
§4의 "재인용 금지, 재측정" 규율을 따른 것이다 — 이전 SHA를 그대로 베껴 쓰지 않았다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-30
run_commit_sha: df4466d12   # M1 커밋 — 자기 SHA는 그 커밋 안에서 쓸 수 없어 후속 커밋에서 backfill
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

```yaml
sync_complete_at: 2026-08-30
sync_commit_sha: f36f65f44   # backfilled — sync 커밋 자체 SHA
sync_status: complete
b12_self_test_a: "grep -c 'SPEC-CI-PR-TRIGGER-FILTER-001' CHANGELOG.md → 0 (사전 중복 없음, emission 진행)"
b12_self_test_b: "grep -oE 'AC-PTF-[0-9]+' spec.md | sort -u | wc -l → 6 (spec.md §3 인라인 AC, acceptance.md 없음 — Tier S). CHANGELOG 엔트리는 6개 AC(AC-PTF-001~006) 전부를 이름으로 인용한다"
b12_self_test_c: "ls .github/workflows/graph-freshness.yml → 존재 확인 완료(CHANGELOG가 인용하는 유일한 구현 파일)"
changelog_entry_position: "### Fixed 섹션, [Unreleased] 하위"
frontmatter_status_transitions.spec_md: "in-progress → implemented → completed (이 sync 커밋 하나로 병합 — 3-phase close)"
frontmatter_status_transitions.updated_field: "2026-08-30로 갱신 (spec.md만 — plan.md/progress.md는 status 필드를 갖지 않음, spec-frontmatter-schema.md § Artifact Statelessness)"
canary_compliance_check: "해당 없음 — 이 SPEC은 forward-looking 정책을 스스로 시험하지 않는다(트리거 필터 명시화이지 카나리 정책이 아님)"
```

### 4-locale 문서 동기화 판단 (README/docs-site)

**판단: 동기화 불필요.** 이유를 스스로 적는다 — 이 변경은 `.github/workflows/graph-freshness.yml`의
`on.pull_request.branches` 필터 1개와 그 옆 주석뿐이다. README·docs-site(`adk.mo.ai.kr`)는
`moai` CLI/에이전트를 사용하는 최종 사용자를 대상으로 하는 문서이고, 이 diff는 저장소 자체의
CI 파이프라인 내부 배선이다 — 사용자가 `moai init`으로 받는 어떤 템플릿에도 속하지 않는다
(§E.2 "템플릿 미러 없음" 확인 — `internal/template/templates/.github/`에는 `label-sync.yml`
1개뿐, `graph-freshness.yml` 자체가 미러되지 않는다). 사용자 관측 가능한 동작, CLI 플래그,
설정 스키마, 문서화된 워크플로 중 어느 것도 바뀌지 않으므로 4-locale 동기화 대상이 아니다.

### 커밋

이 sync 커밋(§E.4 포함)이 이 SPEC의 유일한 sync-phase 커밋이며, `implemented → completed`
전이를 `spec.md` 프런트매터에 병합해 싣는다. `sync_commit_sha`는 이 커밋 안에서 자신을 가리킬
수 없으므로 `pending-backfill-t294` placeholder로 기록하고, 바로 다음 커밋에서 실제 SHA로
backfill한다(spec-frontmatter-schema.md § SHA placeholder backfill exemption).

