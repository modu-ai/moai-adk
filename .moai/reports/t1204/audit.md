# t1204 독립 감사 — sync-auditor

카드: t1204 (SPEC 없음) · 브랜치 `WT-test-gitconfig-leak` · 기준 `df526c9a9` · 감사 대상 HEAD `1abdc8205`
감사 방식: 추적 트리는 읽기 전용. 변이는 `go test -overlay`(빌드 시점 파일 치환)로 수행했고, 피해 저장소는 모두 스크래치패드의 임시 저장소다. 평가 프로필: 내장 기본값(Functionality·Security 필수 통과).

## 판정: FAIL (차단 결함 1건, 한 줄 수리)

카드가 주장한 수리 자체는 실측으로 입증됐다. 다만 같은 파일 `init_gitdetect_test.go:274` 에 같은 결함 계열의 호출이 스크럽 없이 남아 있고, 같은 계기(상속된 `GIT_DIR`)에서 이 호출이 공유 `.git/config` 를 `core.bare=true` 로 바꿔 저장소 전체를 깨뜨리는 것을 재현했다. 카드가 막으려던 「테스트 실행이 바깥 저장소 설정을 오염시키는 문제」가 이 파일 안에서 아직 닫히지 않았다.

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 70 | FAIL | 주장한 네 곳은 RED→GREEN·변이 4/4 입증. 같은 파일 :274 형제가 같은 계기에서 공유 설정을 파괴(F1) |
| Security (25%) | 85 | PASS | 새 입력 표면 없음, 비밀 없음, 의존성 변경 없음. 저장소 무결성 위험은 F1 로 Functionality 에 계상 |
| Craft (20%) | 82 | PASS | 읽기·쓰기 두 모양을 모두 잡는 회귀 테스트, 변경 함수 커버리지 100%. 경미: F2·F3 |
| Consistency (15%) | 92 | PASS | 기존 `internal/gitenv` 재사용, gofmt·vet·golangci-lint 무결 |

조화평균 81.4. 필수 통과 차원 Functionality 가 차단 결함으로 FAIL 이므로 전체 FAIL.

## 결함 목록

- **F1** [High] [blocking] `internal/cli/init_gitdetect_test.go:274` — `TestInitGitDetectionFillsConfig` 의 `exec.Command("git", "-C", projectDir, "init")` 가 스크럽되지 않았다. 상속된 `GIT_DIR` 가 링크된 워크트리의 gitdir(`<repo>/.git/worktrees/<name>`)을 가리키면 공유 설정에 `core.bare=true` 가 기록되고 저장소 전체가 `fatal: this operation must be run in a work tree` 로 멈춘다(R4). 일반 gitdir 에서도 테스트가 거짓 실패한다(R3). 수리 지시: 같은 파일의 `gitDetectInitRepo` 와 같은 방식으로 `cmd.Env = gitenv.Env()` 를 설정한다. 가능하면 회귀 테스트에 `core.bare` 단언을 한 줄 추가한다. 재감사 범위: R3·R4 재실행.
- **F2** [Low] [optional] `internal/cli/init_gitdetect_env_test.go:26` — `victimRemotes` 가 `git config` 의 오류를 버린다(`out, _ :=`). 종료 코드 1(일치 없음) 외의 실패(128 등)도 「원격 없음」으로 읽혀 쓰기 모양 단언이 공허하게 통과할 수 있는 경로다. M0·M4 에서는 정상 발화했으므로 현재 결함은 아니다. 수리 제안: 종료 코드 1 과 그 밖의 실패를 구분한다.
- **F3** [Low] [optional] `.moai/reports/t1204/verdict.md` — Claim 1·2 가 네 곳을 전부인 것처럼 열거하지만 같은 파일에 다섯 번째 호출(:274)이 있다. Residual-risk 의 「~32 other test files」는 명령이 적혀 있지 않은 수치다. 이 감사에서 잰 텍스트 패턴 수는 `internal/cli` 53 파일, 저장소 전체 136 파일이다(아래 E8). 기법이 달라 수치가 다를 수 있으므로 어느 쪽도 결함 개수의 증거는 아니다.
- **F4** [Info] [optional] 실제 사고의 계기(어느 프로세스가 `GIT_DIR` 를 `go test` 에 넘겼는가)는 관측되지 않았다. 현재 `core.hooksPath=/dev/null` 이다. 기제는 R1 로 입증됐고 verdict 도 「e.g.」로 한정했으므로 결함이 아닌 공백으로 기록한다.
- **F5** [Info] [optional] 공유 `.git/config` 에 잘못 들어간 `remote.upstream` 이 아직 남아 있다(E7). 제거는 리드 몫이다.
- **F6** [Low] [optional] `GIT_WORK_TREE` 단독 상속 경우는 시험하지 않았다. verdict 의 Gaps 에 이미 적혀 있다.

## 증거

### E1 — 수리된 트리: 요청 명령 4

```
$ go test ./internal/cli/ -run 'TestDetect.*Config' -count=1 -v   (RUN 줄 제외 발췌)
--- PASS: TestDetectGitConfig_IgnoresInheritedGitDir (0.75s)
--- PASS: TestDetectGitConfig (1.92s)
    --- PASS: TestDetectGitConfig/non-git_directory_falls_back_to_manual (0.02s)
    --- PASS: TestDetectGitConfig/git_repo_with_no_remote_is_manual (0.24s)
    --- PASS: TestDetectGitConfig/github_https_origin_is_personal_+_github (0.32s)
    --- PASS: TestDetectGitConfig/github_ssh_origin_is_personal_+_github (0.33s)
    --- PASS: TestDetectGitConfig/gitlab.com_origin_is_personal_+_gitlab (0.31s)
    --- PASS: TestDetectGitConfig/self-hosted_non-github_origin_is_personal_+_gitlab (0.34s)
    --- PASS: TestDetectGitConfig/remote_present_but_not_named_origin_keeps_github_default (0.36s)
--- PASS: TestDetectGitConfig_RemoteListError (0.00s)
--- PASS: TestDetectGitConfig_OriginURLError (0.00s)
--- PASS: TestDetectUserModifiedConfigs_HashDiff (0.00s)
--- PASS: TestDetectUserModifiedConfigs_NilBaseline (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.269s
exit=0

$ go test ./internal/gitenv/ -count=1; go vet ./internal/cli/; gofmt -l <3 files>
ok  	github.com/modu-ai/moai-adk/internal/gitenv	0.085s
gitenv_exit=0
vet_exit=0
gofmt_exit=0

$ golangci-lint run ./internal/cli/
0 issues.
lint_exit=0
```

### E2 — 회귀 테스트가 기준 코드에서 실패하는가 (검증 1)

방법: `git show df526c9a9:<file>` 로 기준본을 뽑아, 수리본에서 `cmd.Env = gitenv.Env()` 네 줄만 제거한 변이(M0)와 비교했다. 차이는 `cmd :=` 두 줄 분리와 미사용 import 를 막는 `var _ = gitenv.Env` 뿐으로, 의미상 기준 코드와 같다. 이 변이를 `-overlay` 로 주입해 실행했다.

```
##### M0_all4
exit=1
    init_gitdetect_env_test.go:49: read shape: detectGitConfig(prebuilt) = ("manual", "github"), want (personal, gitlab) — answered about the GIT_DIR repo
    init_gitdetect_env_test.go:57: write shape: victim repo gained remotes:
        remote.upstream.url https://gitlab.com/group/proj.git
        remote.upstream.fetch +refs/heads/*:refs/remotes/upstream/*
--- FAIL: TestDetectGitConfig_IgnoresInheritedGitDir (0.69s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.493s
```

verdict.md 의 RED 출력과 문구까지 일치한다.

### E3 — 네 곳 각각의 필요성 (검증 2, 한 곳씩만 되돌린 변이)

```
##### M1_prod_remoteList
exit=1
    init_gitdetect_env_test.go:49: read shape: detectGitConfig(prebuilt) = ("manual", "github"), want (personal, gitlab) — answered about the GIT_DIR repo
    init_gitdetect_env_test.go:60: write shape: detectGitConfig(dir) = ("manual", "github"), want (personal, github)
--- FAIL: TestDetectGitConfig_IgnoresInheritedGitDir (0.62s)
##### M2_prod_originURL
exit=1
    init_gitdetect_env_test.go:49: read shape: detectGitConfig(prebuilt) = ("personal", "github"), want (personal, gitlab) — answered about the GIT_DIR repo
--- FAIL: TestDetectGitConfig_IgnoresInheritedGitDir (0.73s)
##### M3_test_initRepo
exit=1
    init_gitdetect_env_test.go:54: git remote add upstream https://gitlab.com/group/proj.git: exit status 128
--- FAIL: TestDetectGitConfig_IgnoresInheritedGitDir (0.55s)
##### M4_test_addRemote
exit=1
    init_gitdetect_env_test.go:57: write shape: victim repo gained remotes:
        remote.upstream.url https://gitlab.com/group/proj.git
        remote.upstream.fetch +refs/heads/*:refs/remotes/upstream/*
    init_gitdetect_env_test.go:60: write shape: detectGitConfig(dir) = ("manual", "github"), want (personal, github)
--- FAIL: TestDetectGitConfig_IgnoresInheritedGitDir (0.65s)
```

네 곳 모두 단독 제거로 테스트가 빨개진다. 필요 없는 스크럽은 없다. M3 은 단언이 아니라 `t.Fatalf`(exit 128)로 잡히는데, 이는 `$TMPDIR` 가 git 저장소 안에 있지 않다는 환경 전제에 기댄다(경미, 결함으로 올리지 않음).

### E4 — 속성: 그 URL 을 쓸 수 있는 테스트 (검증 3)

```
$ git grep -n 'gitlab.com/group/proj.git' -- '*_test.go'
internal/cli/init_gitdetect_env_test.go:43:	scrubbedGit(t, prebuilt, "remote", "add", "origin", "https://gitlab.com/group/proj.git")
internal/cli/init_gitdetect_env_test.go:54:	gitAddRemote(t, dir, "upstream", "https://gitlab.com/group/proj.git")
internal/cli/init_gitdetect_test.go:81:			remotes:      map[string]string{"origin": "https://gitlab.com/group/proj.git"},
internal/cli/init_gitdetect_test.go:95:			remotes:      map[string]string{"upstream": "https://gitlab.com/group/proj.git"},

$ git grep -n 'gitlab.com/group/proj'      (저장소 전체, 확장자 무관)
  위 4행 + verdict.md:25 + internal/statusline/forge_test.go:24 ("https://gitlab.com/group/proj", 문자열 파싱 표 — git 쓰기 없음)

$ git log --oneline -S 'gitlab.com/group/proj.git' --all -- .
1abdc8205 docs(t1204): record RED/GREEN evidence for GIT_DIR scrub fix
01ba63152 fix(t1204): scrub inherited GIT_DIR from init git-detection
d74a2aaff refactor(cli): drop Git questions from init wizard, auto-detect git mode (#1113)
```

`upstream` + 그 URL 조합은 사고 시점 기준으로 `TestDetectGitConfig` 의 `remote present but not named origin` 부분 테스트(:95) 하나뿐이다. `env_test.go:54` 는 이 카드가 새로 넣은 것이고 수리 뒤에는 스크럽된다. 모든 ref 이력에서 그 URL 을 들여온 커밋은 `d74a2aaff` 하나다. 속성 판단은 성립한다.

### E5 — 사고 재현의 양성·음성 대조 (verdict Gaps 1 해소)

```
##### R1 base (M0 overlay) TestDetectGitConfig$ under GIT_DIR=victim_base/.git
exit=1
--- FAIL: TestDetectGitConfig (1.36s)
    --- FAIL: TestDetectGitConfig/github_ssh_origin_is_personal_+_github (0.16s)
    --- FAIL: TestDetectGitConfig/gitlab.com_origin_is_personal_+_gitlab (0.16s)
    --- FAIL: TestDetectGitConfig/self-hosted_non-github_origin_is_personal_+_gitlab (0.17s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.961s
FAIL
remote.origin.url https://github.com/modu-ai/moai-adk.git
remote.origin.fetch +refs/heads/*:refs/remotes/origin/*
remote.upstream.url https://gitlab.com/group/proj.git
remote.upstream.fetch +refs/heads/*:refs/remotes/upstream/*
victim_remote_exit=0 (1 = no match)
##### R2 fixed TestDetectGitConfig$ under GIT_DIR=victim_fixed/.git
exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli	3.094s
victim_remote_exit=1 (1 = no match)
```

기준 코드는 레인이 보고한 모양(부분 테스트 3건 FAIL, 피해 저장소에 origin + upstream)을 그대로 재현한다. 수리본은 같은 명령에서 피해 저장소를 건드리지 않는다.

### E6 — 같은 파일의 남은 형제 (F1 근거)

```
##### R3 fixed sibling TestInitGitDetectionFillsConfig under GIT_DIR=victim_sib/.git
exit=1
    init_gitdetect_test.go:277: git remote add origin https://github.com/modu-ai/moai-adk.git: exit status 128
--- FAIL: TestInitGitDetectionFillsConfig (0.15s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.743s
FAIL
victim_sib config sha before=9521128740784e9183224f2279fd3ff585dbb5e36c3a25bd4c6484faaf8138ca after=9521128740784e9183224f2279fd3ff585dbb5e36c3a25bd4c6484faaf8138ca
victim_remote_exit=1 (1 = no match)

# 링크된 워크트리 gitdir 모양 (수리된 트리)
per-worktree gitdir=.../scratchpad/t1204-audit/linked/main/.git/worktrees/wt
core.bare before: false
exit=1
    init_gitdetect_test.go:277: git remote add origin https://github.com/modu-ai/moai-adk.git: exit status 128
--- FAIL: TestInitGitDetectionFillsConfig (0.14s)
core.bare after: true
main status exit=128
fatal: this operation must be run in a work tree
```

일반 gitdir 에서는 설정 바이트가 바뀌지 않고 테스트만 거짓 실패한다. 워크트리별 gitdir 에서는 공유 설정이 `core.bare=true` 로 바뀌어 주 저장소가 쓸 수 없게 된다. `internal/gitenv` 패키지 문서가 적어 둔 세 번째 모양이며, 레인은 모두 링크된 워크트리에서 작업하므로 해당 가능성이 가장 큰 모양이다.

### E7 — 공유 설정의 현재 상태 (읽기 전용)

```
$ git config --get-regexp '^remote\.'
remote.origin.url https://github.com/modu-ai/moai-adk.git
remote.origin.fetch +refs/heads/*:refs/remotes/origin/*
remote.upstream.url https://gitlab.com/group/proj.git
remote.upstream.fetch +refs/heads/*:refs/remotes/upstream/*
```

### E8 — 형제 규모 (텍스트 패턴, 결함 수의 증거 아님)

```
$ grep -rln 'exec.Command("git"' internal/cli --include='*_test.go' | wc -l
      53
$ grep -rln 'exec.Command("git"' --include='*_test.go' internal pkg cmd | wc -l
     136
```

이 수는 이미 다른 방식으로 격리한 파일도 포함한다. 재현으로 확인한 형제는 E6 한 건뿐이다.

### E9 — 커버리지와 보안 탐침

```
$ go tool cover -func=<profile> | grep init_gitdetect.go
.../init_gitdetect.go:58:	detectGitConfig		100.0%
.../init_gitdetect.go:81:	hostFromRemoteURL	100.0%

$ git diff df526c9a9..HEAD -- internal/ | grep -nEi '^\+.*(token|secret|password|api[_-]?key|os\.Setenv|exec\.Command\(.*\+)'
sec_probe_exit=1 (1 = no match)
$ git diff df526c9a9..HEAD -- go.mod go.sum | wc -l
       0
```

## Baseline-attribution

모든 측정은 이 세션에서 워크트리 `.claude/worktrees/t1204`, HEAD `1abdc8205` 기준으로 수행했다. 기준 코드는 `git show df526c9a9:<path>` 로 뽑았고 `go test -overlay` 로만 주입했으며, 추적 트리는 수정하지 않았다(이 파일 추가 제외). 피해 저장소 `victim_base`·`victim_fixed`·`victim_sib`·`linked/` 는 모두 스크래치패드 아래 새로 만든 저장소다. 실행 스크립트는 `mk.py`, `run_mutants.sh`, `run_repro.sh`, `run_linked.sh` 이며 원본 로그와 함께 세션 스크래치패드에 있다. 커밋되지 않으므로 인용 수명은 세션에 한정된다.

## verdict.md 평가 (검증 5)

다섯 절이 모두 있고 순서도 맞다. Evidence 는 명령과 원문 출력을 담고, Baseline-attribution 은 RED 와 GREEN 의 측정 시점을 구분했으며, Gaps 는 기준 코드에서 레인 재현 명령을 다시 돌리지 않았다는 사실을 스스로 적었다. 그 공백은 이 감사의 E5 로 메워졌다. 부족한 점은 두 가지다. Claim 이 네 곳을 전부인 것처럼 열거해 같은 파일의 :274 를 놓쳤고, Residual-risk 에도 그 호출이 없다. 「~32」라는 수치에는 그것을 만든 명령이 적혀 있지 않다(F3).

## Gaps

- Windows·Linux 는 측정하지 않았다. CI 판정에 맡긴다.
- `go test ./...` 는 저장소 규칙에 따라 돌리지 않았다.
- 사고 당시 `GIT_DIR` 를 넘긴 실제 프로세스는 관측하지 않았다(F4).
- E8 의 형제 136 파일은 재현하지 않았다. E6 의 한 건만 확인했다.
- 교차 모델 감사(codex·GLM)는 돌리지 않았다.

## Residual-risk

- F1 을 고친 뒤에도 `internal/cli` 에는 스크럽 없는 운영 코드 git 호출이 여럿 있다(예: `session_worktree.go`, `session_worktree_automerge.go`, `spec_lint.go`). 훅 아래에서 `moai` 가 실행되면 같은 기제로 다른 저장소를 읽거나 쓸 수 있다. 이 카드의 범위 밖이다.
- 회귀 테스트는 `core.bare` 모양을 단언하지 않으므로, 앞으로 이 파일에 새로 들어오는 `git init` 호출은 잡지 못한다.
