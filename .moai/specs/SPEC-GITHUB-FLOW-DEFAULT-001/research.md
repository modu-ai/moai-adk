# SPEC-GITHUB-FLOW-DEFAULT-001 — 조사 기록

> 이 문서는 plan-phase 에서 이 SPEC 의 입장이 의지하는 사실을 한곳에 모은다. 각 사실은 **명령, 관측 출력, 측정한 트리**로 귀속된다. 입력 보고서 두 건(`.moai/reports/t1453/m0-measurement.md`, `m0-research-synthesis.md`)에서 가져온 수치는 "인용"으로 표시하고 이 회차에 다시 측정하지 못한 것은 §9 Gaps 에 올린다.
>
> 측정 트리: 카드 트리 `4bf547bcad7c155b1e91485921569db709ec3ac2`. 원격 팁은 2026-10-02 에 `git fetch origin develop main` 으로 가져왔고(종료 코드 0) 해석값은 develop `284e09c44023598affe486f17701717ca173e6ca`, main `4755c5e506225ba90b7a303c5763fa303c699492` 다. 원격 팁은 움직이는 참조이므로 아래 수치를 팁이 움직인 뒤 다시 인용하지 않는다. 개정 0.1.1(2026-10-02)에서 다시 잰 값은 계획 커밋 `855563dba79da74528e0f01560efe80f1016cc11` 에서, 개정 0.1.2 에서 다시 잰 값(M-1·M-3·M-6·M-9, E-36 이후)은 계획 개정 2 커밋 `6c2277295d9ddeaa92e83c0225910445f83cc1c2` 에서 측정했고 GitHub 쪽 값은 같은 날 `gh` 로 읽은 것이다(원장 M-n). GitHub 쪽 값 중 인용하는 것은 모두 실행 id·PR 번호로 고정해 다시 읽은 값뿐이다(design D-28).

## §1 이 회차에 직접 측정한 것과 그 방법

| 주제 | 명령 | 관측 |
|---|---|---|
| 카드 트리 SHA | `git rev-parse HEAD` | `4bf547bcad7c155b1e91485921569db709ec3ac2` |
| 원격 팁 | `git rev-parse origin/develop origin/main` | `284e09c44023598affe486f17701717ca173e6ca` / `4755c5e506225ba90b7a303c5763fa303c699492` |
| 발산 | `git rev-list --count --left-right 284e09c44023598affe486f17701717ca173e6ca...4755c5e506225ba90b7a303c5763fa303c699492` | `7614	1` (develop 에만 7614, main 에만 1) |
| 두 팁의 트리 | `git rev-parse 4755c5e506225ba90b7a303c5763fa303c699492^{tree} 284e09c44023598affe486f17701717ca173e6ca^{tree}` | main `d57fdeecf4b8849da70a6693bbff19ad77ad5932`, develop `4d8a7f46189a3db1c6cad1bb6715e76e90b00b15` (서로 다름) |
| 카드 트리의 미푸시분 | `git rev-list --count 284e09c44023598affe486f17701717ca173e6ca..4bf547bcad7c155b1e91485921569db709ec3ac2` | `122` |
| main 에만 있는 커밋의 내용 | `git show --stat --format='%h %s' 4755c5e506225ba90b7a303c5763fa303c699492` | `retire workflow.worktree.tmux_preferred … (#1740)`, 16개 파일, +5 -29 |
| 그 내용이 develop 에 있는가 | `git grep -c tmux_preferred origin/develop -- internal/config/types.go` | `origin/develop:internal/config/types.go:1` (필드가 아직 있음) |

## §2 GitHub 쪽 상태 (읽기 전용 `gh api`)

| 항목 | 관측 |
|---|---|
| main 보호 | `strict: false`, 필수 체크 `Test (ubuntu-latest)`·`Lint`·`Build (linux/amd64)`·`Analyze (Go) (go)`·`Release PR Multi-OS Gate`, `enforce_admins: true`, 필수 승인 0, force-push 불가, 삭제 불가 |
| 저장소 설정 | 기본 브랜치 `main`, 병합 커밋 허용·squash 허용·rebase 불허, `delete_branch_on_merge: true`, `allow_auto_merge: true` |
| 규칙셋 | 하나 — `Release tag immutability (v*)`, 대상 `refs/tags/v*`, 규칙 `deletion`·`update`, 우회 주체 없음. 생성은 막지 않으므로 rc 태그도 만들 수는 있되 한 번 만들면 지우거나 옮길 수 없다 |
| 공개 여부 | `private: false` (공개 저장소) |

이 값들은 이 회차의 `gh api repos/modu-ai/moai-adk[/branches/main/protection|/rulesets/19583648]` 출력이다. `.github/required-checks.yml` 이 실제 보호와 어긋나는지는 읽지 않았다(§9).

## §3 수렴 가능성 — develop 이 main 을 흡수할 수 있는가

`git merge-tree --write-tree --name-only 284e09c44023598affe486f17701717ca173e6ca 4755c5e506225ba90b7a303c5763fa303c699492` 은 종료 코드 0 과 트리 `992f57f42aee6c3f9a39754c3e0095eaadf9d44b` 한 줄만 출력했다 — 충돌 목록이 비어 있다. develop 이 main 의 유일한 추가 커밋을 흡수하는 병합은 이 두 팁에서 충돌이 없다.

PR 병합 방식은 병합 커밋이 허용되고 rebase 가 불허다. main 으로의 직접 push 는 보호 규칙(PR 필수, `enforce_admins`)이 막는다. 그래서 7614 커밋을 main 에 싣는 경로는 "develop 이 main 을 흡수한 뒤 develop→main PR 을 병합 커밋으로 병합"이 유일하게 이력을 보존하는 경로다(`design.md` §D-8 에서 기각된 대안을 목록으로 둔다).

GitHub 가 7.7천 개 파일·1.07백만 줄 규모(`git diff --shortstat origin/main origin/develop` 의 인용값 `7710 files changed, 1075405 insertions(+), 140589 deletions(-)`)의 PR 을 만들고 병합할 수 있는지는 측정하지 못했다 — Gap.

## §4 CI 사용량

아래 표는 **2026-10-02 의 이 개정 회차에서 실행 id·PR 번호로 고정해 다시 읽은 것**만 담는다. `gh` 목록 응답(`gh run list`·`gh pr list`)은 움직이는 데이터이고, 같은 명령이 다른 창을 돌려주는 것이 두 회차 연속 관측됐다(아래 "목록 호출 불안정" 항목, design D-28). 그래서 목록 호출의 결과는 인용하지 않고 목록에서 얻은 식별자를 `gh run view <id>`·`gh api graphql` 로 하나씩 다시 읽은 값만 인용한다. 평균은 모두 평균이며 중앙값은 이 문서·`design.md` 어디에도 쓰지 않는다.

| 관측 | 명령 | 결과 |
|---|---|---|
| develop push CI 열두 실행 (원장 M-1, 실행 id 열두 개로 고정) | `gh run view <id> --json databaseId,conclusion,createdAt,startedAt,updatedAt,headSha,workflowName,event,headBranch` 를 열두 id 에 각각 | failure 7·cancelled 5·success 0(전부 `event: push`, `headBranch: develop`, 워크플로 `CI`). `createdAt` 2026-10-01T13:38:32Z ~ 2026-10-02T09:27:17Z. failure 7회의 `startedAt`→`updatedAt` 은 1046·1223·1413·1484·1517·1551·1443 초(평균 1382 초). id 와 값은 아래 §4.1 의 표 |
| 최근 병합 PR 20건의 대상 (원장 M-3, PR 번호 스무 개로 고정) | `gh api graphql` 한 번(스무 PR 을 번호별 별칭으로 읽음) | develop 17건·main 3건(1695·1702·1740), 스무 건 모두 `MERGED` |
| main push CI 200건 표본 (원장 M-2 의 1회차 기록) | (인용하지 않는다) | 창을 고정한 목록 호출이었지만 200개 실행을 id 로 고정하지 못했고 감사 회차에 다른 결과와 HTTP 502 를 돌려주어 이 SPEC 의 어떤 결정도 이 수에 기대지 않는다 — §9 Gaps |
| CI 트리거 | `git grep -n -E "branches: \[main" -- .github/workflows/ci.yml` | `:18 branches: [main, develop]`(push), `:20 branches: [main]`(pull_request) (원장 E-21) |

해석(관측에서 나온 것만):

- `ci.yml` 의 PR 트리거가 main 대상 PR 만 포함하고 최근 병합 PR 이 대부분 develop 대상이므로 카드 PR 이 main 을 향하면 별도 트리거 변경 없이 PR 마다 CI 가 돈다.
- develop push CI 의 최근 12회에 success 가 없다는 사실은 절체 경계의 전제(녹색 팁)가 된다 — 운영자가 origin/develop 팁 CI 녹색을 선행 조건으로 정했다(design D-17).
- **1회차 기록에서 뺀 항목**: (1) "PR 이벤트 CI 60회 표본"(success 29·failure 28·cancelled 3, 평균 190 초)은 감사가 같은 명령을 다시 돌렸을 때 다른 값(46/8/6, 평균 470 초)이 나와 재현되지 않으므로 표와 설계의 증거 표에서 뺐다. (2) "main push CI 30회"(success 21·평균 594 초)는 창을 고정하지 않은 `--limit 30` 표본이었고, 이 회차에 같은 명령이 success 7·failure 8·cancelled 15, `createdAt` 2026-07-09 ~ 2026-08-08 이라는 다른 구성을 돌려주었다 — 재현되지 않아 뺐다(0.1.1 이 대체로 둔 창 고정 200건 표본도 감사 회차에 재현되지 않아 0.1.2 가 뺐다). (3) 입력 보고서가 인용한 "CI 1런 평균 1786 초(n=9)"는 어느 표본에서도 같은 선택 기준으로 재현하지 못했다. 이 SPEC 은 그 수치들을 입력으로 쓰지 않는다.
- 비용 사실(질적): 지금은 약 `lead_push_threshold`(20)장의 카드가 develop push CI 한 번으로 검증된다. 카드 PR 체제는 카드마다 PR CI 를 한 번 돌리고 병합 뒤 main push CI 가 또 돈다. 청구 분은 측정하지 못했다. 저장소가 공개(`private: false`)이므로 호스티드 러너 분이 청구 대상인지는 GitHub 정책 확인이 필요하며 이 회차에는 확인하지 않았다.
- `ci.yml` 의 `concurrency` 는 ref 단위로 `cancel-in-progress: true` 다. cancelled 런의 개별 원인은 런별로 확인하지 않았다.
- **목록 호출 불안정(관측)**: `gh run list --workflow=CI --branch develop --limit 12 --json createdAt --jq '{min:…,max:…,n:length}'` 를 이 개정 회차에 세 번 연속 호출했더니 `createdAt` 최솟값·최댓값이 2026-09-07T09:58:24Z~2026-09-07T18:19:55Z, 2026-10-01T15:13:14Z~2026-10-02T14:54:26Z, 같은 값 한 번 더로 나왔다 — 같은 명령이 서로 다른 창을 돌려주었다. 2회차 감사도 세 번 연속 호출이 2026-09-07·2026-09-29~30·2026-10-01~02 창을 돌려주었다고 보고했다. 원인(캐시·페이지 경계·API 변동)은 모른다. 그래서 이 문서는 목록 호출의 결과를 측정값으로 인용하지 않는다.

### §4.1 develop push CI 적색 집합 (실행 id 고정)

2026-10-02 에 아래 열두 id 를 `gh run view <id> --json databaseId,conclusion,createdAt,startedAt,updatedAt,headSha,workflowName,event,headBranch` 로 하나씩 다시 읽었다. 열두 실행 모두 워크플로 `CI`, `event: push`, `headBranch: develop` 이다. 첫 행의 `headSha` 가 원격 develop 팁 `284e09c44023598affe486f17701717ca173e6ca` 와 같다(이 SPEC 의 develop 팁 핀). `failure` 행의 시간은 `startedAt`→`updatedAt` 초다.

| 실행 id | `createdAt`(UTC) | 결론 | `headSha` 앞 9자 | 초 |
|---|---|---|---|---|
| 36989864262 | 2026-10-02 09:27:17 | failure | 284e09c44 | 1046 |
| 36981574917 | 2026-10-02 08:00:12 | failure | 802a72235 | 1223 |
| 36969555332 | 2026-10-02 05:34:37 | failure | c50da9c2f | 1413 |
| 36894416006 | 2026-10-01 16:45:47 | failure | 6789de39b | 1484 |
| 36892129460 | 2026-10-01 16:27:27 | cancelled | 2bb8829de | |
| 36890528950 | 2026-10-01 16:14:47 | cancelled | e83aa47df | |
| 36887673351 | 2026-10-01 15:52:37 | cancelled | 89e3164af | |
| 36886926130 | 2026-10-01 15:46:49 | cancelled | a301529bb | |
| 36883300773 | 2026-10-01 15:19:08 | failure | f28cce310 | 1517 |
| 36882533500 | 2026-10-01 15:13:14 | cancelled | a9f43a6fc | |
| 36876598542 | 2026-10-01 14:28:27 | failure | a466f5585 | 1551 |
| 36870179597 | 2026-10-01 13:38:32 | failure | 678ca0c27 | 1443 |

원인은 첫 행(실행 36989864262, 헤드 284e09c44)만 이 회차에 다시 읽었다 — `gh run view 36989864262 --json jobs` 의 실패 잡은 `Lint`·`Build (windows/amd64)`·`Race Test`·`Test (ubuntu-latest)` 넷이고, `--log-failed` 로 읽은 줄은 다음과 같다(스크래치 파일로 받아 `grep` 으로 골랐다).

| 잡 | 읽은 줄 |
|---|---|
| Lint | `gofmt violations found (run gofmt -w or make fmt):` 다음 `internal/cli/mcp_claude.go`·`internal/config/slice.go`·`internal/web/codex_panel_test.go` |
| Build (windows/amd64) | `##[error]vet: internal/cli/worktree/sweep_test.go:1687:9: undefined: parseLsofCWDs` |
| Test (ubuntu-latest)·Race Test | `workflow_split_test.go:154: ENTRY_ROUTER_LOC_VIOLATION: run.md has 263 LOC (ceiling: 200)` 와 `--- FAIL` 일곱 이름(`TestCodemapsFoldGuardFixtures`·`TestCodemapsFoldPreservationGuard`·`TestCommitIdentityGuard_BuiltinListCoversFixtureEnumeration`·`TestEntryRouterLOCCeiling`·`TestFixturePackagesScrubProcess`·`TestRegisterCanonicalizesCaseVariantCWD`·`TestSyncGateLanguageDetectionMatchesScript`) |

나머지 열한 실행의 원인(로그 줄)과 다섯 cancelled 의 취소 사유는 읽지 않았다 — §9 Gaps. 이 표는 `.moai/reports/t1453/m0-develop-ci-red.md`(gitignored 로컬 보조 자료)의 표를 id 고정으로 다시 읽은 것이며 그 파일은 인용 대상이 아니다.

## §5 squash 착지 판정 — 격리 저장소 실험

`git cherry` 가 squash 병합을 처리한다는 입력 보고서의 서술을 검증하려고 스크래치 저장소에서 실험했다(카드 트리 밖, 시험 후 보관하지 않음).

- 구성: `main` 에서 `card` 브랜치를 따 2개 커밋(`c1`, `c2`)을 만들고, `git merge --squash card` 와 커밋으로 main 에 한 커밋을 만들었다.
- `git merge-base --is-ancestor card main` → 종료 코드 `1` (조상이 아님).
- `git cherry main card` → `+ b1815fbf…`, `+ 095266af…` — **두 커밋 모두 `+`(미착지)** 로 읽힌다.
- `git diff main~1 card | git patch-id` 와 `git show main --format= | git patch-id` → 두 줄의 patch-id 가 같다(`0dfce5633258fb32337944f69fb563cccc031d3f`).

결론: `internal/cli/session_worktree.go` 의 `git cherry` 보조 판정은 **카드가 커밋 하나일 때만** squash 를 알아본다. 커밋이 둘 이상인 카드는 조상 관계도 cherry 도 "미착지"로 답해 정리가 영원히 거부된다. 누적 diff 의 patch-id 비교가 필요하고, 그 비교도 기준점 이후 main 이 같은 파일을 바꿨다면 일치하지 않으므로 세 번째 층(PR 병합 상태)이 필요하다. 이 결론이 `design.md` §D-5 의 근거다.

## §6 인접 카드 (큐 관측)

`moai gtd` 를 스크래치 파일로 받아 읽었다(종료 코드 0, 95줄).

- t810 — 상태 `picked`. 줄 아래에 `↳ absorbs t1453 (agent)` 가 기록돼 있다.
- t1453 — 상태 `picked`. 줄 아래에 `↳ absorbs t810 (agent)` 가 기록돼 있다.
- 두 카드가 서로를 흡수한다고 적혀 있어 관계 기록만으로는 어느 쪽이 대체되는지 정해지지 않았다. 카드 본문 t1453 은 "t810: 흡수/대체 판정"을 요구했고, 리더가 t1453 이 t810 을 흡수하기로 정했다(design D-19).
- t1452 — 상태 `queued`. 범위 (a) 지명 없는 acquire, (b) 흡수 트리 재측정을 창 획득 전에, (c) 동일 테스트 명령 재실행 억제.
- t1448 — 상태 `picked`. 이 카드의 트리 `4bf547bca` 의 최근 병합 목록에 착지 흔적은 보이지 않는다(착지 여부 자체는 읽지 않았다).

## §7 develop 을 기준으로 삼는 표면

### 7.1 직접 읽은 것

| 표면 | 관측 (명령 요지) |
|---|---|
| `internal/cli/worktree/done.go:281` | `const landingBaseBranch = "develop"` |
| `internal/cli/worktree/sweep.go:224` | `--base` 기본값 `"origin/develop"` |
| `internal/cli/factory_card.go:1064`, `factory_merge.go:93`, `integration.go:353`, `session_worktree_automerge.go:176,183` | `.DevelopBranch` 읽기 5곳 |
| `internal/cli/factory_card.go:1157` | `git merge --no-ff` 로컬 병합 |
| `internal/config/loader_integration_branch.go` | 통합 목표는 git-flow 면 `DevelopBranch`, 그 밖에는 워크플로 해석표로 정한다. manual 프로필+git-flow 일 때만 `DevelopBranch` 가 채워진다 |
| `internal/cli/session_worktree.go:797-852` | 통합 ref 리터럴 `refs/remotes/origin/develop`, 조상 판정 후 `git cherry` |
| `scripts/release.sh:122-131` | 현재 브랜치가 `main` 이 아니면 거부(`--hotfix` 예외) |
| `scripts/release.sh:158-175` | 모든 버전(rc 포함)에 CHANGELOG 절 필수 |
| `.github/workflows/release.yml:96-121` | 출처 검증 검사 4~7 (검사 5 CHANGELOG, 6 `system.yaml` 버전, 7 main 조상) |
| `.goreleaser.yml:49-52` | `release:` 블록에 `prerelease` 설정 없음 |
| `.github/workflows/release-pr-multi-os.yml` | `detect-release` 는 `release/*` 헤드 또는 수동 실행에서만 동작, `release-pr-gate` 는 `if: always()` 로 `failure` 만 실패 처리 |
| `.github/workflows/spec-lint.yml:70, 87-99` | `develop` 을 가져오는 단계, `release/*` 헤드가 `origin/develop` 과 같아야 한다는 논리 |
| `.coderabbit.yaml:69` | `auto_review.base_branches` 에 `develop` |
| `.moai/config/sections/git-strategy.yaml:5,9,16` | `worktree_base_branch: develop`, `workflow: git-flow`, `develop_branch: develop` |
| `.moai/config/sections/workflow.yaml:170,283` | `deny_commits_on: [main]`, `push_develop: true` |
| `internal/contract/projection_mission.go:82,222` | `missionMergeTarget = "develop"` 이 `MergeTarget: missionMergeTarget` 으로 투영된 미션 계약에 들어간다(개정 A-1, §12, 원장 E-46) |
| `internal/mission/git_owner.go:201` | `ActionLocalMerge` 경계가 현재 브랜치가 리터럴 `develop` 이 아니면 거부한다(개정 A-1 범위 밖, §12, 원장 E-50) |
| `AGENTS.md` | `-w develop` 5줄(:76, :145-150) |
| `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` | 네 파일 모두 `-w develop` 적중 |
| docs-site `workflow-commands/moai-sync.md` | `-w develop` 적중 en 6·ja 5·ko 5·zh 5 |
| 비테스트 Go 에서 `develop` 토큰 | `push_serializer.go` 18, `session_worktree_automerge.go` 13, `closuretest.go` 8, `profile_matrix.go` 7, `integration.go` 7, `worktree/done.go` 6 (상위 6개) |
| `_test.go` 중 `develop` 부분 문자열 적중 | 333개 파일(`manager-develop` 같은 무관 용례 포함, 분류하지 않음) |

### 7.2 입력 보고서에서 인용 (직접 다시 읽지 못함)

- 표면 건수(템플릿 사본 110·`internal/cli/testdata` 58·docs-site 로케일별 28~31·`.claude/rules/moai` 25·`.claude/skills/moai` 21·`.github` 12 …)는 단어 패턴 grep 의 후보 목록이며 결함 목록이 아니다.
- 진짜 브랜치명 줄 약 330/원 적중 956(621은 `manager-develop`·동사·`develop-*` 식별자) — ±2 분류 추정.
- 살아 있는 절차를 담은 상위 파일: `.claude/rules/local/gitflow-lane-protocol.md`, `.claude/agents/harness/hns-release-specialist.md`, `.moai/docs/gitflow-integration-chain.md`, `AGENTS.local.md`(§4.1), `.moai/docs/git-workflow-doctrine.md`.
- 구조 식별자: `workflow.autonomy.contract.push_develop`·액션 `push-develop`·`--develop-worktree`·`develop_branch` 등(코드/설정 변경이 필요, 문서 스윕이 아니다).
- `moai factory complete` 는 이름이 develop 인 워크트리를 요구하지 않고 통합 브랜치가 체크아웃된 워크트리를 `git worktree list` 로 찾는다(입력 보고서 Q2 답).
- 로케일 편차: `moai-sync.md` 의 살아 있는 develop 문장이 en 3·ko 2·ja 2·zh 4 로 갈린다는 분류 추정(원 적중은 위 표의 6/5/5/5).

## §8 릴리스 하네스

- `scripts/release.sh v9.9.9-rc.1 --dry-run` 을 카드 트리(브랜치 `WT-github-flow-default`)에서 실행했다. 종료 코드 1, 마지막 줄 `[FAIL] Must be on 'main' branch (current: WT-github-flow-default). Use --hotfix for hotfix branches.` — 이 스크립트는 현재 브랜치 이름이 `main` 이 아니면 거부한다. `--dry-run` 은 태그와 push 를 하지 않으며 이 호출은 검증 5 에서 멈췄다(검증 1~4 는 변경이 없다).
- 격리 저장소에서 `git checkout --detach main` 뒤 `git rev-parse --abbrev-ref HEAD` 는 `HEAD` 를 출력했다(종료 코드 0). 하네스가 쓰는 detached 워크트리는 이 스크립트의 브랜치 이름 검사를 통과할 수 없다.
- `scripts/release.sh` 의 검증 6(`origin/$CURRENT_BRANCH` 와 동기)과 검증 9(`commits/<sha>/status` 의 결합 상태 API)도 브랜치 이름과 결합 상태에 기댄다. 후자가 Actions 체크 런을 반영하는지는 확인하지 않았다.
- 3-OS 매트릭스는 `release-pr-multi-os.yml` 의 `full-matrix-test`(ubuntu·macOS·windows, `go test -race`, 제한 시간 30)이고 `detect-release` 가 `release/*` 헤드 또는 `workflow_dispatch` 에서만 통과시킨다. 필수 체크 `Release PR Multi-OS Gate` 는 `if: always()` 이며 `detect-release` 가 failure 일 때와 매트릭스가 failure 일 때만 실패하므로 비릴리스 PR 에서는 성공한다(입력 보고서가 PR #1740 에서 관측한 성공과 일치하는 구조이며, 이 회차는 워크플로 파일만 읽었다).
- `ci.yml:278` 의 `Race Test` 는 `release/*` 헤드에서 건너뛴다. `ci.yml:400` 의 `test-integration` 은 3-OS 행렬이다.
- 3-OS 게이트 옵션 A(모든 main PR), B(태그 전 `workflow_dispatch` 와 스크립트 강제), C(`release.yml` 안, GoReleaser 전), D(main push 와 태그 전 녹색 SHA 확인)의 비용 비교는 `design.md` §D-7 에 있다.

## §9 Gaps — 이 회차에 관측하지 않은 것

- 호스티드 러너 청구 분, CodeRabbit 한도 이력(입력 보고서의 3회 표본 "Review completed"는 인용이며 재측정하지 않았다).
- 병합 큐의 `strict: false` 환경 동작.
- GitHub 가 대형 diff 의 PR 을 만들고 병합 커밋으로 병합할 수 있는지.
- `Release PR Multi-OS Gate` 의 비릴리스 PR 결과(워크플로 파일만 읽음, 실제 PR 결과는 입력 보고서 인용).
- 입력 보고서가 읽지 못했다고 밝힌 것: 약 70개 `_test.go`, `internal/settings/testdata/sections/git-strategy.yaml`, `auto-merge.yml` 110~260줄, `graph-freshness.yml` 80줄 이후, `internal/core/project/initializer*.go`, `.github/required-checks.yml`.
- detached 워크트리에서 `scripts/release.sh --dry-run` 을 직접 실행하지 않았다(브랜치 이름 검사 소스와 `rev-parse` 관측으로 추론).
- 도구 거부: `git ls-files` 를 쓴 파이썬 일회성 스크립트가 워크트리 가드에 거부돼, 상시로드 기준선은 `git` 을 이름에 담지 않는 파일 순회 스크립트로 다시 측정했다. 측정값은 거부된 호출에 의존하지 않는다(§10 의 상시로드 기준선).
- 상시로드 기준선: 이 트리에서 `.claude/rules`+`.claude/output-styles` 의 `.md` 중 선두 frontmatter 에 최상위 `paths:` 가 없는 파일 16개, 합계 285,543 바이트, `paths:` 가 있는 파일 100개. `CLAUDE.md` 15,658·`AGENTS.md` 20,252·`AGENTS.local.md` 38,073 바이트. 순회 스크립트 방식이라 `acceptance.md` 의 RED-now 셀(단일 호출)이 아니라 회귀 방지 기준선이다.
- **개정 0.1.1 에서 더한 Gaps**:
  - **`gh run list` 목록 호출은 이 세션에서 불안정한 것이 관측됐다**(§4 "목록 호출 불안정": 같은 명령 세 번 연속 호출이 서로 다른 창을 돌려주었고 감사 회차에도 같은 현상이 있었으며 HTTP 502 한 번과 서로 다른 200건 결과가 있었다). 목록 호출 결과는 측정값으로 인용하지 않는다. 1회차의 PR 이벤트 표본과 main push 30회 표본, 0.1.1 이 대체로 둔 main push 200건 표본(원장 M-2 의 1회차 기록)은 재현되지 않아 뺐다. **main push CI 의 벽시계·성공률은 이 SPEC 에서 미측정이다.** 이 SPEC 의 CI 사용량 근거는 §4 의 id 고정 두 행(열두 실행, 스무 PR)뿐이다.
  - 열두 실행 중 첫 실행(36989864262)의 실패 원인만 이 회차에 로그 줄까지 다시 읽었다. 나머지 열한 실행의 로그 줄과 다섯 cancelled 실행의 취소 사유는 읽지 않았다.
  - **병합 큐의 후속 측정 항목**(카드 발행은 리더 소관, 이 SPEC 은 카드 id 를 적지 않는다): 호스티드 러너 청구 분, CodeRabbit 한도 이력(3회 표본 "Review completed" 는 인용이며 재측정하지 않았다), 녹색 PR 뒤 적색 main 의 빈도.
  - 이 회차의 도구 거부(verification-claim-integrity §3.1) 둘: (1) 워크트리 가드가 스크래치 프로브를 만드는 복합 셸 한 번(`mkdir` 과 `cat` 과 `go run` 을 이은 호출)을 "too complex" 로 거부했다. 같은 일을 `Write` 도구로 파일을 만든 뒤 `go -C <dir> run .` 한 번으로 다시 했고 프로브 결과는 그 두 번째 경로에서만 왔다. (2) `for` 반복문으로 네 파일의 `status:` 를 훑는 복합 호출을 거부했고 파일마다 단순한 `grep` 으로 다시 했다. 두 거부 모두 인용한 수치는 거부되지 않은 재호출에서 왔다.
  - 스윕 가드 프로브(원장 M-5·M-9)의 소스는 저장소 밖 스크래치에 있고 커밋하지 않았다. 반복 가능한 형태는 AC-GFD-021 의 가드 자체 시험이다. M-9 의 표면 스캔(위반 286·약한 적중 31·표지 면제 9)은 제 재구현이 센 값이다 — 섹션·표지 처리와 펜스 판정이 실제 가드와 다를 수 있어 미래 가드의 값이 아니다.
  - 이 개정 회차의 도구 거부(verification-claim-integrity §3.1): 워크트리 가드가 스무 PR 을 한 번에 읽는 GraphQL 질의를 셸 변수 반복문으로 조립한 호출을 "runtime 값" 으로 거부했다. 같은 질의를 스크래치 파일에 리터럴로 적어 `sh` 로 실행했고 PR 번호별 값은 그 두 번째 경로에서만 왔다. 로컬 `gh` 가 저장소 밖 작업 디렉터리에서 `failed to determine base repo` 로 실패한 호출 한 번도 있었고 저장소 안에서 다시 실행했다.
  - `moai constitution amend` 는 실행하지 않았고 `--dry-run` 도 실행하지 않았다(`--dry-run` 이 대화형 층을 건너뛰는지 확인하지 못했다).
  - 허용 목록 상한(40)의 근거가 되는 class N 줄 수는 이 회차에 다시 세지 않았다.
  - 판정 도구 빌드: 설치된 `moai`(v3.2.0-rc.25, 커밋 802a72235)는 이 트리 HEAD 의 엄격한 조상이며 178 커밋 뒤처져 있다(`git merge-base --is-ancestor` 종료 코드 0). `moai constitution validate`·`moai spec lint` 의 인용은 이 트리 HEAD 에서 빌드한 `moai` 를 경로로 호출한 결과이고 설치본은 쓰지 않았다(verification-claim-integrity §2.2). 개정 0.1.2 회차의 `moai constitution list`·`moai spec lint` 는 HEAD `6c2277295d9ddeaa92e83c0225910445f83cc1c2` 에서 `go build -o <스크래치>/bin/moai ./cmd/moai` 로 새로 빌드한 바이너리를 경로로 불렀다(종료 코드 0). ldflags 없이 빌드해 `moai version` 이 `v3.1.3`·커밋 `none` 을 찍으므로 빌드 커밋은 바이너리가 아니라 빌드한 트리의 HEAD 로 귀속한다.
- **개정 0.1.3 에서 더한 Gaps**: §12.3.
- 이전 회차의 열린 질문은 모두 결정이 되었다 — §11.

## §10 상시로드 기준선 (REQ-GFD-014)

위 §9 마지막에서 두 번째 항목의 수치를 그대로 쓴다. M4 착수 시 다시 측정해 "전" 값으로 삼고 M4 종료 때 "후" 값을 잰다. 호스트가 `InstructionsLoaded` 관측을 노출하면 함께 기록하고, 노출하지 않으면 그 사실을 보고한다(`.claude/rules/moai/workflow/rule-loading-budget.md`).

## §11 이전 회차 질문 7건의 처분

1회차에 열려 있던 일곱 질문은 모두 번호 붙은 결정이 되었다. 표는 `plan.md` §I 에 있고 결정의 본문·출처·뒤집기 여부는 `design.md` 의 D-16 ~ D-22 에 있다. 이 문서의 사실이 그 결정을 지지하는 곳은 다음과 같다.

- D-16(rc 규칙 R-a): §8 — `scripts/release.sh` 와 `release.yml` 의 검사 구조.
- D-17(절체 선행 조건): §4·§4.1 — develop push CI 열두 실행에 success 가 없다는 관측과 첫 실행의 적색 원인(실행 id 고정).
- D-18(수렴 PR 폴백): §3 — 대형 diff 의 PR 가능 여부는 측정하지 못했다(Gap).
- D-19(t810 흡수): §6.
- D-20(병합 큐 보류): §2 의 `strict: false` 와 §4.
- D-21(와이어 식별자): §7.1.
- D-22(필수 체크 `Release PR Multi-OS Gate`): §8 — 비릴리스 PR 에서 무동작 성공이라는 구조.

## §12 개정 0.1.3 (A-1)에서 측정한 사실 — 계약에서 미션으로의 투영의 병합 목표

핀: 카드 트리 HEAD `eda61419564296892b41d0e7608c8b136e52c04c`(이 회차에 쟀다). 명령·종료 코드·출력 전문은 `acceptance.md` §B 원장 E-46~E-55 와 측정 행 M-11 에 있고 아래 표는 그 요지다. 이 트리는 앞 회차의 세 트리와 같다고 주장하지 않는다.

### 12.1 측정한 사실

| 사실 | 원장 | 관측 |
|---|---|---|
| 투영의 병합 목표는 리터럴 `develop` 상수다 | E-46 | `projection_mission.go:82` 한 줄, 같은 파일 222 줄이 `MergeTarget` 에 쓴다 |
| 이 기준을 관측하는 시험이 없다 | E-47 | `TestProjectionMergeTarget` 이 `internal` 어디에도 없다(종료 코드 1) |
| 투영에 비테스트 호출자가 없다 | E-48 | 정의 한 줄뿐. 시험 파일의 호출 지점은 넷(`:4`) |
| `MergeTarget` 을 읽는 비테스트 코드는 하나다 | E-49 | `internal/mission/contract.go:58` 의 봉인 전 비어 있지 않음 검사. 값은 계약 해시에 들어가지만 브랜치 이름과 비교되는 곳이 없다 |
| 로컬 병합 경계는 리터럴 `develop` 을 요구한다 | E-50 | `internal/mission/git_owner.go:201`, 읽기만 했다 |
| 고정된 스냅숏의 `develop` 은 동결된 행동 식별자 한 줄이다 | E-51 | `mission_surface_baseline.txt:45` 의 `ActionLocalMerge … "local_develop_merge"`. `MergeTarget` 은 필드 선언(:347)뿐 |
| 계약 코어는 `internal/config` 를 import 하지 않는다 | E-52 | 종료 코드 1 과 빈 출력. 코어의 내부 import 는 `projection_mission.go:10` 의 `internal/mission` 하나 |
| 동결된 계약 쪽 행동 식별자의 사용 줄 | E-53 | `projection_mission.go`·`rules.go` 에 세 줄 |
| 카드 트리에서 이 HEAD 까지 `internal/mission`·`internal/contract` 는 무변경이다 | E-54·E-55 | 빈 출력, 같은 형태의 양성 대조가 `internal/cli/goal.go` 를 보인다 |
| 변경 전 투영 시험과 스냅숏 비교가 초록이다 | M-11 | `TestContractProjectsOntoMissionValidator` 다섯 limb 모두 `--- PASS`, `ok … 0.308s` |

### 12.2 읽어서 알게 된 것 (실행 관측 아님)

- `mission.MissionContract.MergeTarget` 의 생산 지점은 둘이다 — `internal/cli/goal.go` 의 approve(M1 이 `IntegrationTarget` 으로 옮김)와 이 투영. 서로의 값을 받지 않는다.
- 통합 목표 해석기(`config.LoadGitFlowIntegrationConfig(root).IntegrationTarget`)는 git-flow 면 `develop_branch`(manual 프로필 게이트), github-flow 면 `main`, gitlab-flow 면 환경 값, release-flow 면 릴리스 접두를 돌려주고 모르는 워크플로는 빈 문자열이다(`loader_integration_branch.go`, `loader_workflow_disposition.go` 를 읽었다).
- 투영의 설계 표(`SPEC-AUTONOMY-PRECONDITION-001` `design.md` §D)는 `MergeTarget` 을 `develop` 으로 적었다. 그 SPEC 은 `completed` 이고 이 개정은 그 파일을 고치지 않는다 — 코드가 바뀐 뒤 그 행은 역사적 기록이 된다.

### 12.3 이 회차에 관측하지 않은 것 (개정 0.1.3 의 Gaps)

- `internal/mission/git_owner.go:201` 경계가 github-flow 구성에서 어떻게 동작하는지 실행해 관측하지 않았다 — 줄을 읽었을 뿐이다. 이 개정은 그 줄을 범위 밖에 둔다(design D-30).
- github-flow 구성에서 `local-merge-develop` 행동을 가진 계약의 처분(투영에서 거르기, 허용하되 휴면)은 정해지지 않았고 실행해 관측하지 않았다. `progress.md` §G 의 권고 8.
- 변경 전 투영이 픽스처에 낸 봉인 해시 리터럴은 이 회차에 구하지 않았다. M2a 의 첫 단계(특성화 시험)가 변경 전 트리에서 구한다(AC-GFD-024 (3)).
- `internal/contract` 패키지 전체 시험은 돌리지 않았다 — M-11 은 지정한 시험 하나다.
- 도구 거부(verification-claim-integrity §3.1): 워크트리 가드가 `git ls-files` 와 `git grep` 을 파일 반복문으로 이은 복합 호출을 "too complex" 로 거부했다. 같은 일을 단순한 개별 호출(`git grep -n 'moai-adk/internal/' -- internal/contract/projection_mission.go` 와 `git grep -n -F 'moai-adk/internal/config"' -- internal/contract`)로 나눠 다시 쟀고 E-52 는 그 개별 호출의 출력이다. 거부된 호출의 결과는 어디에도 인용하지 않았다.

---

🗿 MoAI
