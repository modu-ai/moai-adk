# t637 — acquire 창 기록의 브랜치 출처 판정 (방향 결정 전 보고)

- 카드: t637 · 레인 워크트리 `.claude/worktrees/t637` · 브랜치 `WT-acquire-branch-record`
- 판독 좌표: develop `4c99d973e` (= origin/develop, 이 트리 HEAD)
- 작성: 2026-09-11 · 단계: 판독 + 재현 (SPEC·run 이전)

## 1. 결론 요약

카드의 전제("acquire 가 호출 트리의 브랜치를 적는다")는 **절반만 맞다.** 브랜치 판독은 t449(`4d58f407b`, 2026-09-03)에서 이미 3단 순서로 바뀌었다 — `--branch` 인자 → 설정의 git-flow `develop_branch` → 호출 트리. 호출 트리가 기록되는 것은 앞의 둘이 모두 비었을 때뿐인 폴백이다.

lane-4 가 본 `on WT-zerowidth-blank-review` 는 이 폴백을 탄 결과로 읽힌다. primary 체크아웃의 작업 사본 `git-strategy.yaml` 이 `workflow: git-flow` 인데 `develop_branch: ""` 이기 때문이다. 게다가 문서화된 모든 acquire 호출이 `--branch` 도 `--card` 도 넘기지 않는다.

그래서 카드가 제시한 두 방향은 둘 다 그대로는 맞지 않는다.

- **`--branch` 인자 추가**: 이미 있다(`internal/cli/integration.go:292`). 추가할 것이 없다.
- **필드 이름을 '호출 트리'로 바꾸기**: 설정이 정상인 경로에서는 필드가 통합 대상(develop)을 정확히 담는다(재현 C3). 이름을 바꾸면 정상 경로에서 오히려 거짓이 된다.

리드의 실제 문제("status 가 엉뚱한 카드를 가리킨다")는 두 층으로 갈린다.

1. **설정 층**: `develop_branch` 가 비면 조용히 호출 트리로 떨어진다. 경고가 없다.
2. **표면 층**: `branch` 필드는 설계상 통합 대상이라, 정상 동작해도 `develop` 만 보여 준다. 카드를 식별하는 필드는 `card`(`--card`)인데, `status` 텍스트 출력이 이 필드를 찍지 않는다(`integration.go:214`). JSON 에만 나온다.

## 2. 코드 판독 (좌표 `4c99d973e`)

| 위치 | 내용 |
|---|---|
| `internal/cli/integration.go:246` | `branch, wt := resolveIntegrationTarget(branchFlag, config.LoadGitFlowDevelopBranch(root))` |
| `internal/cli/integration.go:117-131` | 해석 순서: 명시 인자 → 설정값 → `currentBranch()` + `os.Getwd()`(129행) |
| `internal/cli/integration.go:97` | `currentBranch()` = 호출 프로세스 cwd 에서 `git rev-parse --abbrev-ref HEAD` |
| `internal/config/loader_integration_branch.go:34-48` | `mode == manual` 이고 `workflow == git-flow` 일 때만 `develop_branch` 를 돌려주고, 나머지는 전부 `""` |
| `internal/cli/integration.go:292` | `--branch` 도움말 "Branch being integrated" — 원천 브랜치로 읽히지만, 구현은 통합 **대상**으로 취급한다(`worktreeForBranch` 로 그 브랜치를 체크아웃한 트리를 찾음) |
| `internal/cli/integration.go:214` | status 텍스트는 holder/branch/worktree/since 만 찍고 `card` 는 찍지 않는다 |
| `internal/hook/integration_lock_guard.go:101` | 가드는 `lock.Branch` 를 거부 **메시지**에만 쓴다. 판정 입력이 아니다 |
| 바이너리 드리프트 | `git diff --stat ed71054d3 4c99d973e -- internal/cli/integration.go internal/config/loader_integration_branch.go internal/kanban/integration_lock.go internal/config/types.go` → 출력 없음 (설치된 rc.7 커밋과 develop 사이에 해당 코드 변경 없음) |

## 3. 설정 판독 (primary 체크아웃, lock root)

`integrationLockRoot()`(`integration.go:46`)는 `CLAUDE_PROJECT_DIR`, 없으면 `git rev-parse --git-common-dir` 의 부모를 쓴다. 레인 워크트리에서 호출해도 **primary 체크아웃의 작업 사본 설정**을 읽는다.

```
$ grep -n -E '^\s*(mode|workflow|develop_branch):' /Users/goos/MoAI/moai-adk-go/.moai/config/sections/git-strategy.yaml
4:  mode: "manual" # manual, personal, team
16:    workflow: git-flow
43:    develop_branch: ""
...
$ stat -f '%Sm %N' .../git-strategy.yaml
Sep 10 12:34:33 2026 /Users/goos/MoAI/moai-adk-go/.moai/config/sections/git-strategy.yaml
$ git log -1 --format='%h %cI %s' 030987513
030987513 2026-09-10T16:22:28+09:00 merge: t612 MX scanner pending-WARN defects F03+F04 (Go review report)
```

- 비교: `origin/develop` 블롭은 `develop_branch: develop`(16행), `main` 블롭은 `workflow: github-flow`.
- 작업 사본의 마지막 수정(12:34)이 t612 병합 커밋(16:22)보다 이르다. 따라서 그 창 시점에도 `develop_branch` 는 비어 있었다고 판단한다(§5 Gaps 참조).
- CLAUDE.local.md §2.3 이 경고하는 `moai update` 이후 git-flow 키 재적용이 반쯤만 된 모양이다(`workflow` 는 복구, `develop_branch`·`release_branch_prefix`·`rc_version_format` 은 빈 값).

## 4. 재현 — /tmp 픽스처 (실제 창·develop 워크트리 미접촉)

픽스처: `/tmp/t637-fx`(root), 워크트리 `/tmp/t637-fx-wt/{develop,cardA(WT-a),cardB(WT-b)}`. 바이너리는 설치본 `moai-adk v3.2.0-rc.7`. 셀마다 `CLAUDE_PROJECT_DIR=/tmp/t637-fx` 를 같은 호출 안에서 지정했다.

| 셀 | 설정 `develop_branch` | 호출 위치 | 인자 | 기록 branch / worktree | 판정 |
|---|---|---|---|---|---|
| C1 대조군: 같은 트리에서 자기 브랜치 병합 | `""` | cardA | 없음 | `WT-a` / `/tmp/t637-fx-wt/cardA` | 병합 대상 WT-a 와 **일치** |
| C2 대조군: 다른 트리에서 다른 브랜치 병합 | `""` | cardA | 없음 (병합 의도 WT-b) | `WT-a` / `/tmp/t637-fx-wt/cardA` | **불일치** — lane-4 관측 재현 |
| C3 설정 정상 | `develop` | cardA | `--card tB` | `develop` / `/private/tmp/t637-fx-wt/develop` | 통합 대상 정확. 카드는 text status 에 안 보임 |
| C4 명시 인자 | `""` | cardA | `--branch WT-b --card tB` | `WT-b` / `/private/tmp/t637-fx-wt/cardB` | 인자를 원천 브랜치로 주면 창 트리가 **카드 트리**로 기록됨(t449 결함 모양의 재진입) |

원문 출력:

```
== C1/C2: cwd=/tmp/t637-fx-wt/cardA config develop_branch empty, no --branch
release-integration window acquired by fx-lane4 on WT-a
release-integration window: held
  holder:   lane-4 (fx-lane4, pid 14968)
  branch:   WT-a
  worktree: /tmp/t637-fx-wt/cardA
  since:    2026-09-11T02:59:16Z
release-integration window released (was fx-lane4 on WT-a)
== C4: cwd=/tmp/t637-fx-wt/cardA --branch WT-b --card tB
release-integration window acquired by fx-lane4 on WT-b
release-integration window: held
  holder:   lane-4 (fx-lane4, pid 14968)
  branch:   WT-b
  worktree: /private/tmp/t637-fx-wt/cardB
  since:    2026-09-11T02:59:17Z
{"held":true,"lock":{"session_id":"fx-lane4","session_name":"lane-4","pid":14968,"pid_source":"session-owner","branch":"WT-b","worktree":"/private/tmp/t637-fx-wt/cardB","acquired_at":"2026-09-11T02:59:17Z","card":"tB"},"root":"/tmp/t637-fx","stale":false}
release-integration window released (was fx-lane4 on WT-b)
== C3: cwd=/tmp/t637-fx-wt/cardA config develop_branch=develop, no --branch, --card tB
release-integration window acquired by fx-lane4 on develop
release-integration window: held
  holder:   lane-4 (fx-lane4, pid 14968)
  branch:   develop
  worktree: /private/tmp/t637-fx-wt/develop
  since:    2026-09-11T02:59:57Z
release-integration window released (was fx-lane4 on develop)
```

실제 창 미접촉 확인: `ls -la /Users/goos/MoAI/moai-adk-go/.moai/state/ | grep -i integration-lock || echo NO_REAL_LOCK_FILE` → `NO_REAL_LOCK_FILE`.

문서화된 호출 전수(`grep -rn -o 'moai integration acquire[^`"|)]*'`): CLAUDE.local.md 338·357·371, `kanban-dispatch.md:230`, `gitflow-lane-protocol.md:47·51`, `hns-release-specialist.md:345` — 7건 모두 `--branch` 없음, `--card` 없음.

## 5. 두 방향의 장단 (+ 제3안)

| 안 | 장점 | 단점 |
|---|---|---|
| A. `--branch` 인자 추가 | — | 이미 존재(t449). 게다가 이름·도움말이 "원천 브랜치"로 읽혀, 레인이 카드 브랜치를 넘기면 C4 처럼 창 트리가 카드 트리로 기록된다. 인자를 늘리는 쪽은 오기록 경로를 넓힌다 |
| B. 필드 이름을 '호출 트리'로 정직화 | 폴백 경로의 기록과는 이름이 맞는다 | 정상 경로(C3·명시 인자)에서는 호출 트리가 아니므로 이름이 거짓이 된다. t449 가 고친 의미를 되돌린다. "어느 카드인가"도 여전히 답하지 못한다 |
| **C. (권고) 출처를 드러내고 카드를 보이게** | 리드의 실제 질문(누가 **어느 카드**를 병합 중인가)에 직접 답한다. 필드 의미(통합 대상)는 유지 | 코드 3곳 + 문서 7곳 수정. internal/cli 컴파일 필요 |

C안 구성(SPEC 후보 범위):

1. `status` 텍스트 출력에 `card:` 줄 추가(기록돼 있으면).
2. 브랜치 출처를 기록·표시: `branch_source: flag|config|caller`. 그리고 git-flow 워크플로인데 `develop_branch` 가 비어 호출 트리로 떨어질 때 acquire 가 경고 1줄을 낸다(거부는 하지 않음 — 폴백은 비-git-flow 프로젝트에선 정당하다).
3. `--branch` 도움말을 "integration target branch (the branch the merge lands on)"로 고친다.
4. 문서화된 호출 7곳에 `--card <id>` 를 싣는다.

별도(코드 밖, 리드·운영자 소관): primary 체크아웃 `git-strategy.yaml` 의 `develop_branch: develop` 재적용. 이 레인은 설정을 편집하지 않았다.

## 6. Gaps (관측하지 않은 것)

- lane-4 가 2026-09-10 t612 창에서 실제로 실행한 바이너리 버전. 설치본 mtime 은 `Sep 11 04:20` 로 그 창보다 뒤다. t449 이전 바이너리였어도 관측 모양은 같다.
- 그 창 시점의 `git-strategy.yaml` 내용은 mtime(12:34 < 16:22)으로 추정했을 뿐, 그 시점 스냅숏을 본 것은 아니다.
- lane-4 가 `--branch`·`--card` 를 넘기지 않았다는 것은 문서 전수로 추정한 것이다. 그 창의 명령 원문은 보지 못했다.
- 픽스처 재현은 설치 바이너리로 했다. develop 트리를 컴파일한 바이너리는 쓰지 않았다(컴파일 슬롯 미승인). 대신 해당 파일들의 `ed71054d3..4c99d973e` diff 가 비었음을 확인했다.

## 7. Residual-risk

- C안의 경고는 새 표면이라 kanban-dispatch 문서와 테스트(`integration_lock_cli_test.go`)의 출력 단정에 걸릴 수 있다.
- 설정만 고치고 C안을 하지 않으면 status 는 `branch: develop` 을 정확히 보여 주지만, 카드 두 장을 병행하는 레인이 어느 카드를 병합 중인지는 여전히 텍스트로 보이지 않는다.
- 픽스처 `/tmp/t637-fx`, `/tmp/t637-fx-wt/` 는 SPEC 단계 재사용을 위해 남겨 두었다.
