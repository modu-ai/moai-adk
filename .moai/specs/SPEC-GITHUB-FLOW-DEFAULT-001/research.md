# SPEC-GITHUB-FLOW-DEFAULT-001 — 조사 기록

> 이 문서는 plan-phase 에서 이 SPEC 의 입장이 의지하는 사실을 한곳에 모은다. 각 사실은 **명령, 관측 출력, 측정한 트리**로 귀속된다. 입력 보고서 두 건(`.moai/reports/t1453/m0-measurement.md`, `m0-research-synthesis.md`)에서 가져온 수치는 "인용"으로 표시하고 이 회차에 다시 측정하지 못한 것은 §9 Gaps 에 올린다.
>
> 측정 트리: 카드 트리 `4bf547bcad7c155b1e91485921569db709ec3ac2`. 원격 팁은 2026-10-02 에 `git fetch origin develop main` 으로 가져왔고(종료 코드 0) 해석값은 develop `284e09c44023598affe486f17701717ca173e6ca`, main `4755c5e506225ba90b7a303c5763fa303c699492` 다. 원격 팁은 움직이는 참조이므로 아래 수치를 팁이 움직인 뒤 다시 인용하지 않는다. 개정 0.1.1(2026-10-02)에서 다시 잰 값은 계획 커밋 `855563dba79da74528e0f01560efe80f1016cc11` 에서 측정했고 GitHub 쪽 값은 같은 날 `gh` 로 읽은 것이다(원장 M-n).

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

아래 표는 **2026-10-02 의 이 개정 회차에서 다시 측정해 재현된 것**만 담는다. `gh` 목록 응답은 움직이는 데이터이고 창을 고정하지 않은 `--limit` 표본은 같은 명령이 다른 구성을 돌려주는 것이 관측됐다(아래 마지막 항목). 그래서 가능한 곳은 `--created` 로 창을 고정했다. 평균은 모두 평균이며 중앙값은 이 문서·`design.md` 어디에도 쓰지 않는다.

| 관측 | 명령 | 결과 |
|---|---|---|
| develop push CI 최근 12회 (원장 M-1) | `gh run list --workflow=CI --branch develop --limit 12 --json conclusion,createdAt,startedAt,updatedAt --jq …` | failure 7·cancelled 5·success 0, `createdAt` 2026-10-01T13:38:32Z ~ 2026-10-02T09:27:17Z. failure 7회의 `startedAt`→`updatedAt` 은 1046·1223·1413·1484·1517·1551·1443 초(평균 1382 초). 같은 구성을 이 회차에 세 번 관측 |
| main push CI, 창 고정 (원장 M-2) | `gh run list --workflow=CI --event push --branch main --created '2026-07-01..2026-10-02' --limit 200 --json conclusion,createdAt,startedAt,updatedAt --jq …` | n=200(한도에 닿았으므로 창 안의 최근 200회), success 146·cancelled 43·failure 10·startup_failure 1, `createdAt` 2026-07-16T21:58:31Z ~ 2026-09-10T01:46:49Z, success 평균 438 초. 같은 결과를 두 번 관측 |
| 최근 병합 PR 20건의 대상 (원장 M-3) | `gh pr list --state merged --limit 20 --json baseRefName,mergedAt --jq …` | develop 17건·main 3건, `mergedAt` 2026-09-09T20:10:46Z ~ 2026-10-01T16:45:43Z. 두 번 관측 |
| CI 트리거 | `git grep -n -E "branches: \[main" -- .github/workflows/ci.yml` | `:18 branches: [main, develop]`(push), `:20 branches: [main]`(pull_request) (원장 E-21) |

해석(관측에서 나온 것만):

- `ci.yml` 의 PR 트리거가 main 대상 PR 만 포함하고 최근 병합 PR 이 대부분 develop 대상이므로 카드 PR 이 main 을 향하면 별도 트리거 변경 없이 PR 마다 CI 가 돈다.
- develop push CI 의 최근 12회에 success 가 없다는 사실은 절체 경계의 전제(녹색 팁)가 된다 — 운영자가 origin/develop 팁 CI 녹색을 선행 조건으로 정했다(design D-17).
- **1회차 기록에서 뺀 항목**: (1) "PR 이벤트 CI 60회 표본"(success 29·failure 28·cancelled 3, 평균 190 초)은 감사가 같은 명령을 다시 돌렸을 때 다른 값(46/8/6, 평균 470 초)이 나와 재현되지 않으므로 표와 설계의 증거 표에서 뺐다. (2) "main push CI 30회"(success 21·평균 594 초)는 창을 고정하지 않은 `--limit 30` 표본이었고, 이 회차에 같은 명령이 success 7·failure 8·cancelled 15, `createdAt` 2026-07-09 ~ 2026-08-08 이라는 다른 구성을 돌려주었다 — 재현되지 않아 위의 창 고정 표본으로 대체했다. (3) 입력 보고서가 인용한 "CI 1런 평균 1786 초(n=9)"는 어느 표본에서도 같은 선택 기준으로 재현하지 못했다. 이 SPEC 은 그 수치들을 입력으로 쓰지 않는다.
- 비용 사실(질적): 지금은 약 `lead_push_threshold`(20)장의 카드가 develop push CI 한 번으로 검증된다. 카드 PR 체제는 카드마다 PR CI 를 한 번 돌리고 병합 뒤 main push CI 가 또 돈다. 청구 분은 측정하지 못했다. 저장소가 공개(`private: false`)이므로 호스티드 러너 분이 청구 대상인지는 GitHub 정책 확인이 필요하며 이 회차에는 확인하지 않았다.
- `ci.yml` 의 `concurrency` 는 ref 단위로 `cancel-in-progress: true` 다. cancelled 런의 개별 원인은 런별로 확인하지 않았다.

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
  - 창을 고정하지 않은 `gh run list --limit N` 표본은 같은 명령이 다른 구성을 돌려주는 것이 관측됐다(§4). 1회차의 PR 이벤트 표본과 main push 30회 표본은 재현되지 않아 표에서 뺐다. 이 SPEC 의 CI 사용량 근거는 §4 의 재현된 세 행뿐이다.
  - **병합 큐의 후속 측정 항목**(카드 발행은 리더 소관, 이 SPEC 은 카드 id 를 적지 않는다): 호스티드 러너 청구 분, CodeRabbit 한도 이력(3회 표본 "Review completed" 는 인용이며 재측정하지 않았다), 녹색 PR 뒤 적색 main 의 빈도.
  - 이 회차의 도구 거부(verification-claim-integrity §3.1) 둘: (1) 워크트리 가드가 스크래치 프로브를 만드는 복합 셸 한 번(`mkdir` 과 `cat` 과 `go run` 을 이은 호출)을 "too complex" 로 거부했다. 같은 일을 `Write` 도구로 파일을 만든 뒤 `go -C <dir> run .` 한 번으로 다시 했고 프로브 결과는 그 두 번째 경로에서만 왔다. (2) `for` 반복문으로 네 파일의 `status:` 를 훑는 복합 호출을 거부했고 파일마다 단순한 `grep` 으로 다시 했다. 두 거부 모두 인용한 수치는 거부되지 않은 재호출에서 왔다.
  - 스윕 가드 프로브(원장 M-5)의 소스는 저장소 밖 스크래치에 있고 커밋하지 않았다. 반복 가능한 형태는 AC-GFD-021 의 가드 자체 시험이다.
  - `moai constitution amend` 는 실행하지 않았고 `--dry-run` 도 실행하지 않았다(`--dry-run` 이 대화형 층을 건너뛰는지 확인하지 못했다).
  - 허용 목록 상한(40)의 근거가 되는 class N 줄 수는 이 회차에 다시 세지 않았다.
  - 판정 도구 빌드: 설치된 `moai`(v3.2.0-rc.25, 커밋 802a72235)는 이 트리 HEAD 의 엄격한 조상이며 178 커밋 뒤처져 있다(`git merge-base --is-ancestor` 종료 코드 0). `moai constitution validate`·`moai spec lint` 의 인용은 이 트리 HEAD 에서 빌드한 `moai` 를 경로로 호출한 결과이고 설치본은 쓰지 않았다(verification-claim-integrity §2.2).
- 이전 회차의 열린 질문은 모두 결정이 되었다 — §11.

## §10 상시로드 기준선 (REQ-GFD-014)

위 §9 마지막에서 두 번째 항목의 수치를 그대로 쓴다. M4 착수 시 다시 측정해 "전" 값으로 삼고 M4 종료 때 "후" 값을 잰다. 호스트가 `InstructionsLoaded` 관측을 노출하면 함께 기록하고, 노출하지 않으면 그 사실을 보고한다(`.claude/rules/moai/workflow/rule-loading-budget.md`).

## §11 이전 회차 질문 7건의 처분

1회차에 열려 있던 일곱 질문은 모두 번호 붙은 결정이 되었다. 표는 `plan.md` §I 에 있고 결정의 본문·출처·뒤집기 여부는 `design.md` 의 D-16 ~ D-22 에 있다. 이 문서의 사실이 그 결정을 지지하는 곳은 다음과 같다.

- D-16(rc 규칙 R-a): §8 — `scripts/release.sh` 와 `release.yml` 의 검사 구조.
- D-17(절체 선행 조건): §4 — develop push CI 12회에 success 가 없다는 관측, 그리고 `.moai/reports/t1453/m0-develop-ci-red.md` 의 적색 원인.
- D-18(수렴 PR 폴백): §3 — 대형 diff 의 PR 가능 여부는 측정하지 못했다(Gap).
- D-19(t810 흡수): §6.
- D-20(병합 큐 보류): §2 의 `strict: false` 와 §4.
- D-21(와이어 식별자): §7.1.
- D-22(필수 체크 `Release PR Multi-OS Gate`): §8 — 비릴리스 PR 에서 무동작 성공이라는 구조.

---

🗿 MoAI
