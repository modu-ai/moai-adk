# t637 — internal/cli 컴파일 슬롯 실행 기록

- 카드: t637 · SPEC: SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001 v0.3.0 (in-progress)
- 트리: `.claude/worktrees/t637`, 브랜치 `WT-acquire-branch-record`, HEAD `f680dab46` (구현 `ba0725be8` 포함, 로컬 develop `00ae57ad7` 흡수 후)
- 슬롯: 리드 지명 lane-7 (2026-09-12). 시작 전 `ps` 로 go test/build 프로세스 0건 확인.
- 증거 디렉터리: `.moai/reports/t637/ac-evidence/`

## (b) AC 테스트 — 전부 rc=0, FAIL 0

| 증거 | 대상 | 결과 |
|---|---|---|
| slot-001 | status card 줄 | PASS 1 |
| slot-002 | card 없음 · 이름 없음 형태 | PASS 2 |
| slot-003 | branch_source flag · blank_flag (하위 테스트 선택자가 두 개만 고름 — 감사 때 미확인 항목 해소) | PASS 1+2 |
| slot-004 | branch_source config | PASS 1+1 |
| slot-005 | git-flow 폴백 경고 | PASS 1 |
| slot-006 | 폴백 불일치 가시화 | PASS 1 |
| slot-007 | 경고 없음 음성 4종 | PASS 4 |
| slot-008 | stderr 전용 · 거절 시 무경고 | PASS 2 (+하위 2) |
| slot-009 | 구 레코드 호환 | PASS 1 |
| slot-010 | --branch 도움말 | PASS 1 |
| slot-016a | t449 회귀 세트 | PASS 10 |
| slot-016b | config seam (슬롯 밖 허용) | PASS 13 |
| slot-015 · slot-016c | template 누출 · hook 통합락 (슬롯 밖, 이전 커밋) | PASS 1 · 14 |
| slot-lint | `golangci-lint run ./internal/cli/... ./internal/kanban/... ./internal/config/...` | rc=0, `0 issues.` |

## (c) 뮤테이션 13행 — 전부 kill, 원복은 백업 cp → cmp

행마다: `cp <file> /tmp/t637-mut/<file>.bak` (슬롯 시작 시 1회) → `sed` 적용 → 백업과 `diff` 로 바뀐 줄 1개 확인 → 지명 테스트 실행 → `cp` 원복 → `cmp` rc=0.

| 행 | 파일 | 지명 테스트 | 결과 |
|---|---|---|---|
| a | integration.go | GitFlowEmptyDevelopWarnsOnCallerFallback | FAIL (kill) |
| b | loader_integration_branch.go | GitHubFlowEmptyDevelopDoesNotWarn | FAIL (kill) |
| c | integration.go | Status_ShowsCardLine | FAIL (kill) |
| d | integration.go | RecordsBranchSource/{flag,config} | FAIL ×2 (kill) |
| e | integration.go | WarningIsOnStderrOnly/{text,json} | FAIL ×2 (kill) |
| f | integration.go | Status_OldRecordKeepsTodaysBranchLine | FAIL (kill) |
| g | loader_integration_branch.go | NonManualModeGitFlowDoesNotWarn | FAIL (kill) |
| h | integration.go | RecordsBranchSource/blank_flag | FAIL (kill, 빌드 실패 아님 확인) |
| i1 | loader_integration_branch.go | NoConfigCallerFallbackDoesNotWarn | FAIL (kill) |
| i2 | integration.go | NoConfigCallerFallbackDoesNotWarn | FAIL (kill) |
| j | integration.go | BranchFlagHelpNamesTheTarget | FAIL (kill) |
| k | integration.go | RefusedAcquireDoesNotWarn | FAIL (kill) |
| l | kanban/integration_lock.go | Status_OldRecordKeepsTodaysBranchLine | FAIL (kill) |

원복 확인: 세 파일 모두 백업과 `cmp` rc=0, `git status --short` 에 소스 변경 없음(증거 파일만). 원복 뒤 AC 테스트 14개를 한 번에 재실행(`restored-pass.txt`): rc=0, 최상위 PASS 14, FAIL 0.

## (d) 픽스처 바이너리

`go build -o /tmp/t637-fx-bin/moai ./cmd/moai` → build-rc=0. `--branch` 도움말(`cmd-010b.txt`):

```
    --branch                The integration target branch the merge lands on, not the card branch being merged (default: the configured git-flow develop branch, else the current branch)
```

## (e) 픽스처 셀 — 실제 창 무접촉

| 셀 | 설정 develop_branch | rc | 경고 수 | status |
|---|---|---|---|---|
| C1′ | `""` | 0 | 1 (`"WT-a"` 명시) | `card: tA` · `branch: WT-a (source: caller)` |
| C2′ | `""` | 0 | 1 | `card: tB` · `branch: WT-a (source: caller)` — 불일치가 한눈에 보임 |
| C3′ | `develop` | 0 | 0 | `card: tB` · `branch: develop (source: config)`, JSON `"branch_source":"config"` |

C1′ 경고 원문:

```
[moai:integration-lock] warning: git-flow project with no develop branch configured; the window was recorded against the caller's branch "WT-a". Set git_strategy.manual.develop_branch, or pass --branch <integration-target>.
```

실제 창 가드: 셀마다 앞뒤 `shasum /Users/goos/MoAI/moai-adk-go/.moai/state/integration-lock.json` → 여섯 번 모두 `063ae13b46b700f78ae34e9418bac488933033bf` 로 동일(`guard-before-c*.txt` · `guard-after-c*.txt`).

- 관찰: 슬롯 시작 시점에는 실제 락 파일이 없었다(`moai integration status --json` → `"held":false`). 픽스처 셀 직전에 재보니 파일이 생겨 있었다 — 보유자는 `lane-8` (`session_id 7e33c4bd-…`, branch `develop`, `acquired_at 2026-09-11T19:13:59Z`). 이 레인은 실제 root 로 acquire 를 호출한 적이 없다. 리드 공지의 "lane-4 보유"와 다르다.
- 픽스처 셀은 모두 서브셸 안에서 cd 했고, 셀 뒤 `pwd` 가 워크트리 루트 그대로였다.

## (f) make build

`make build` → rc=0 (`slot-make-build.txt`). Makefile 이 `gen-catalog-hashes --all` 을 스스로 돌려 "catalog.yaml updated" 를 출력했지만, `git diff --stat -- internal/template/catalog.yaml` 은 비어 있다 — 커밋된 해시와 같다. `bin/` 은 gitignore 대상.

## Gaps

- 전체 스위트(`go test ./...`)는 돌리지 않았다 — 범위 밖, CI 몫.
- `make build` 산출 `bin/moai` 로는 픽스처를 돌리지 않았다(픽스처는 `/tmp/t637-fx-bin/moai`, 같은 트리에서 LDFLAGS 없이 빌드).
- 크로스 플랫폼(windows/linux) 빌드는 확인하지 않았다 — CI 몫.

## Residual-risk

- `CLAUDE.local.md` 세 줄 편집은 primary 의 미커밋 사본과 줄 위치가 다르다(이 트리 370/389/403, primary 338/357/371). 이 카드의 develop 병합에서는 충돌하지 않지만, primary 작업이 main 에 커밋되거나 release PR 이 develop 사본을 main 으로 옮길 때 충돌할 수 있다.
