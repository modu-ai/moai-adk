# t1204 재감사 2차 — sync-auditor

카드: t1204 (SPEC 없음) · 브랜치 `WT-test-gitconfig-leak` · 기준 `df526c9a9` · 감사 대상 HEAD `343dbdc57` (수리 `85a90cda1` + 문서 `343dbdc57`)
범위: 1차 감사(`audit.md`)의 결함 F1·F2·F3 에 대한 델타 재감사, 새로 보고된 `.moai/db` 누수의 범위 판단, 요청된 검사 재실행.
방식: 추적 트리는 읽기 전용. 변이는 `go test -overlay` 로만 주입했고, 피해 저장소는 모두 세션 스크래치패드(`.../scratchpad/t1204-audit2/`)에 새로 만들었다. 평가 프로필: 내장 기본값(Functionality·Security 필수 통과).

## 판정: PASS-WITH-DEBT

차단 결함이던 F1 은 닫혔다. 두 테스트 파일의 git 호출 네 곳이 모두 스크럽되고, 새 회귀 테스트는 스크럽 한 줄을 빼면 빨개지며, 링크된 워크트리 모양에서도 공유 설정이 바뀌지 않는다. 남은 것은 부채 두 건이다. 운영 코드 쪽 `.moai/db` 누수(N2)는 이 카드의 범위 밖이라 후속 카드로 넘기는 것이 맞고, `victimConfig` 주석이 코드보다 강한 보장을 약속한다(N1).

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 90 | PASS | F1 닫힘(E1·E3), 변이 MG 적중(E2), 요청 테스트 12건 PASS(E4) |
| Security (25%) | 88 | PASS | 새 입력 표면·비밀·의존성 변경 없음(E6). 저장소 무결성 위험은 재현에서 사라짐 |
| Craft (20%) | 85 | PASS | vet·gofmt·golangci-lint 무결, 조회 헬퍼가 128 을 치명 처리. 감점: N1 |
| Consistency (15%) | 93 | PASS | `gitInitAt` 한 곳으로 모아 기존 `internal/gitenv` 재사용, 파일 관례 준수 |

가중 평균 88.9, 조화 평균 88.9. 필수 통과 차원 둘 다 PASS.

## 결함 목록

- **F1** [High → 닫힘] `init_gitdetect_test.go:274` 는 이제 `gitInitAt` 을 거치며, 그 헬퍼가 `cmd.Env = gitenv.Env()` 를 단다(`:33-34`). 두 파일의 `exec.Command("git"…)` 은 모두 네 곳이고 네 곳 모두 스크럽된다(E1). R3·R4 재실행에서 공유 설정 해시가 그대로이고 `core.bare=false`, 원격 없음(E3).
- **F2** [Low → 처방대로 닫힘] `victimConfig` 가 종료 코드 1 만 「일치 없음」으로 받고 나머지는 `t.Fatalf` 로 처리한다. 1차 감사가 짚은 128 모양(깨진 설정)은 이제 치명 처리된다(E5).
- **F3** [Low → 닫힘] verdict.md 는 다섯 절을 순서대로 갖췄고, 「~32」가 사라졌으며, 형제 수치 53·136·134 에 명령이 붙었다. 세 수치를 이 감사에서 다시 재서 일치를 확인했다(E7). 경미한 점: 피해 저장소 결과 블록 일부(`(each)`, `false   (victim1)`)는 원문 출력이 아니라 정리된 형태다. 결과는 이 감사의 E3 가 독립적으로 재현했으므로 결함으로 올리지 않는다.
- **N1** [Low] [optional] `internal/cli/init_gitdetect_env_test.go:24-26` — 주석은 「읽을 수 없는 설정이 손대지 않은 설정으로 통과할 수 없다」고 약속하지만, git 은 파일이 없거나 권한이 없을 때도 종료 코드 1 을 낸다(E5: missing 1, unreadable 1). 그 두 경우는 여전히 「일치 없음」으로 읽힌다. `TestGitInitAt_IgnoresInheritedWorktreeGitDir` 는 `core.bare` 전제 단언이 이를 잡지만, `TestDetectGitConfig_IgnoresInheritedGitDir` 에는 전제 단언이 없다. 다만 그 피해 저장소는 테스트가 직접 만들므로 파일 부재 경로는 사실상 도달하지 않는다. 수리 제안: 주석을 「종료 코드 1 이외의 실패」로 좁히거나, 첫 테스트에 `core.bare` 전제 읽기 한 줄을 넣는다.
- **N2** [Medium] [optional — 후속 카드] 운영 코드 누수. 기제: `internal/core/git/checkout.go:115-117` `runGitRevParse` 가 환경을 스크럽하지 않으므로 `ResolveGitDirs` 가 상속된 `GIT_DIR` 를 따른다 → `internal/homestate/paths.go:48,72-79` `CanonicalProjectRoot` 가 워크트리별 gitdir 을 주 체크아웃 루트로 옮긴다 → 그 루트가 임시 경로 안이면 `ProjectDir` 가 `<루트>/.moai/db/<key>` 를 고르고 `paths.go:280` 이 `project.json` 을 쓴다. 재현은 E3(R4'). 일반 gitdir 모양(R3')에서는 누수가 없다. 범위 판단은 아래 별도 절에.
- **N3** [Info] 이번 실행에서 `hostFromRemoteURL` 커버리지가 80.0% 로 1차의 100% 와 다르다. 1차는 다른 테스트 집합의 프로필이었고 이 함수는 카드가 바꾸지 않았다(E4). 결함 아님.

## N2 범위 판단 — 후속 카드로 넘긴다

1. 생산자가 카드 밖에 있다. 카드가 바꾼 운영 파일은 `init_gitdetect.go` 하나뿐이고, `internal/homestate`·`internal/core` 에는 변경이 없다(E6 `diffstat` 출력 없음). 카드가 누수를 만든 것이 아니라, F1 수리로 테스트가 `runInit` 까지 도달하게 되면서 이미 있던 동작이 드러난 것이다.
2. 카드의 주장과 축이 다르다. 카드는 「테스트 실행이 바깥 저장소의 git 설정을 오염시킨다」를 막는 것이고, 그 주장은 E3 에서 입증됐다. N2 는 추적되지 않는 파일 하나를 쓰는 것이며 설정 파괴가 아니다.
3. 수리는 설계 결정이다. `ResolveGitDirs`/`CanonicalProjectRoot` 를 스크럽하면 호출자 전원의 경로 해석이 바뀐다. git 훅 아래에서 도는 `moai` 가 `GIT_DIR` 를 따라야 하는지는 운영 의미의 문제이지 테스트 격리의 문제가 아니다. `runInit` 을 부르는 테스트 파일도 17개다(E7).
4. 테스트 쪽에서 `GIT_DIR` 를 지우는 국소 처방은 가능하지만, 운영 결함을 가리기만 하므로 권하지 않는다.

임시 경로 밖의 실제 레인 워크트리에서 같은 일이 일어나면 `~/.moai/db/<주 체크아웃 key>/project.json` 에 이미 있는 것과 같은 내용을 쓰게 되어 실질 피해가 없을 것으로 **추론**한다. 측정하지 않았다(Gaps).

## 증거

### E1 — 두 테스트 파일의 git 호출 전수

```
$ grep -n -A2 'exec.Command("git"' internal/cli/init_gitdetect_test.go internal/cli/init_gitdetect_env_test.go internal/cli/init_gitdetect.go
internal/cli/init_gitdetect_env_test.go:17:	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
internal/cli/init_gitdetect_env_test.go-18-	cmd.Env = gitenv.Env()
--
internal/cli/init_gitdetect_env_test.go:29:	cmd := exec.Command("git", append([]string{"config", "--file", configPath}, args...)...)
internal/cli/init_gitdetect_env_test.go-30-	cmd.Env = gitenv.Env() // keep an inherited GIT_DIR out of repository setup
--
internal/cli/init_gitdetect_test.go:33:	cmd := exec.Command("git", "-C", dir, "init")
internal/cli/init_gitdetect_test.go-34-	cmd.Env = gitenv.Env() // an inherited GIT_DIR would outrank -C
--
internal/cli/init_gitdetect_test.go:43:	cmd := exec.Command("git", "-C", dir, "remote", "add", name, url)
internal/cli/init_gitdetect_test.go-44-	cmd.Env = gitenv.Env() // an inherited GIT_DIR would outrank -C
--
internal/cli/init_gitdetect.go:26:	cmd := exec.Command("git", "-C", dir, "remote")
internal/cli/init_gitdetect.go-27-	cmd.Env = gitenv.Env() // an inherited GIT_DIR would outrank -C
--
internal/cli/init_gitdetect.go:38:	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
internal/cli/init_gitdetect.go-39-	cmd.Env = gitenv.Env() // an inherited GIT_DIR would outrank -C

$ grep -nE 'exec\.|os/exec' internal/cli/init_gitdetect_test.go internal/cli/init_gitdetect_env_test.go
internal/cli/init_gitdetect_test.go:11:	"os/exec"
internal/cli/init_gitdetect_test.go:33:	cmd := exec.Command("git", "-C", dir, "init")
internal/cli/init_gitdetect_test.go:43:	cmd := exec.Command("git", "-C", dir, "remote", "add", name, url)
internal/cli/init_gitdetect_env_test.go:5:	"os/exec"
internal/cli/init_gitdetect_env_test.go:17:	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
internal/cli/init_gitdetect_env_test.go:29:	cmd := exec.Command("git", append([]string{"config", "--file", configPath}, args...)...)
internal/cli/init_gitdetect_env_test.go:33:		var exitErr *exec.ExitError
```

`CommandContext` 등 다른 형태의 호출은 없다. 네 곳 모두 스크럽된다.

### E2 — 새 회귀 테스트의 변이 적중 (검증 2)

변이 MG: `gitInitAt` 의 스크럽 한 줄만 제거하고 `-overlay` 로 주입했다. `gitenv` 는 `gitAddRemote` 가 계속 쓰므로 미사용 import 우회가 필요 없다.

```
$ diff internal/cli/init_gitdetect_test.go MG_init_gitdetect_test.go
34d33
< 	cmd.Env = gitenv.Env() // an inherited GIT_DIR would outrank -C
diff_exit=1

##### MG mutant (gitInitAt scrub removed) TestGitInitAt_IgnoresInheritedWorktreeGitDir
exit=1
    init_gitdetect_env_test.go:104: shared config gained core.bare=true — gitInitAt re-initialized the inherited worktree gitdir
--- FAIL: TestGitInitAt_IgnoresInheritedWorktreeGitDir (0.52s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.110s
FAIL
##### MG control (HEAD, no overlay) same test
exit=0
--- PASS: TestGitInitAt_IgnoresInheritedWorktreeGitDir (0.52s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.093s
```

verdict.md 의 RED 출력(`:104`, 같은 문구)과 일치한다.

### E3 — R3'·R4' 재실행 (HEAD `343dbdc57`, 새 피해 저장소)

```
HEAD=343dbdc57 status=[]
##### R3' HEAD TestInitGitDetectionFillsConfig under GIT_DIR=victim_sib/.git
exit=0
--- PASS: TestInitGitDetectionFillsConfig (0.47s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.052s
victim_sib config sha before=9521128740784e9183224f2279fd3ff585dbb5e36c3a25bd4c6484faaf8138ca after=9521128740784e9183224f2279fd3ff585dbb5e36c3a25bd4c6484faaf8138ca
core.bare=false
victim_remote_exit=1 (1 = no match)
victim_sib worktree listing:
.git
##### R4' HEAD linked-worktree gitdir
per-worktree gitdir=/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/aa4a8059-2c32-4b20-bdde-000da988d1ca/scratchpad/t1204-audit2/linked/main/.git/worktrees/wt
core.bare before: false
exit=0
--- PASS: TestInitGitDetectionFillsConfig (0.35s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.933s
core.bare after: false
shared config sha before=9521128740784e9183224f2279fd3ff585dbb5e36c3a25bd4c6484faaf8138ca after=9521128740784e9183224f2279fd3ff585dbb5e36c3a25bd4c6484faaf8138ca
remote_exit=1 (1 = no match)
main status exit=0
?? .moai/
wt status exit=0
/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/aa4a8059-2c32-4b20-bdde-000da988d1ca/scratchpad/t1204-audit2/linked/main/.moai/db/main-3b0e7bbf/project.json
{
  "schema_version": 1,
  "project_key": "main-3b0e7bbf",
  "project_root": "/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/aa4a8059-2c32-4b20-bdde-000da988d1ca/scratchpad/t1204-audit2/linked/main"
}
```

1차 감사의 R4(`core.bare after: true`, `fatal: this operation must be run in a work tree`)가 사라졌다. `?? .moai/` 는 N2 이며, 일반 gitdir 모양(R3')의 피해 저장소에는 `.git` 외에 아무것도 생기지 않았다.

### E4 — 요청 테스트·커버리지

```
$ go test ./internal/cli/ -run 'TestDetectGitConfig|TestInitGitDetectionFillsConfig|TestGitInitAt_IgnoresInheritedWorktreeGitDir' -count=1 -v -coverprofile=<scratch>/c.out   (--- / ok lines)
test_exit=0
--- PASS: TestDetectGitConfig_IgnoresInheritedGitDir (0.85s)
--- PASS: TestGitInitAt_IgnoresInheritedWorktreeGitDir (0.54s)
--- PASS: TestDetectGitConfig (1.66s)
    --- PASS: TestDetectGitConfig/non-git_directory_falls_back_to_manual (0.02s)
    --- PASS: TestDetectGitConfig/git_repo_with_no_remote_is_manual (0.18s)
    --- PASS: TestDetectGitConfig/github_https_origin_is_personal_+_github (0.29s)
    --- PASS: TestDetectGitConfig/github_ssh_origin_is_personal_+_github (0.30s)
    --- PASS: TestDetectGitConfig/gitlab.com_origin_is_personal_+_gitlab (0.28s)
    --- PASS: TestDetectGitConfig/self-hosted_non-github_origin_is_personal_+_gitlab (0.31s)
    --- PASS: TestDetectGitConfig/remote_present_but_not_named_origin_keeps_github_default (0.29s)
--- PASS: TestDetectGitConfig_RemoteListError (0.00s)
--- PASS: TestDetectGitConfig_OriginURLError (0.00s)
--- PASS: TestInitGitDetectionFillsConfig (0.48s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	4.361s	coverage: 7.0% of statements
github.com/modu-ai/moai-adk/internal/cli/init_gitdetect.go:58:			detectGitConfig				100.0%
github.com/modu-ai/moai-adk/internal/cli/init_gitdetect.go:81:			hostFromRemoteURL			80.0%
$ go test ./internal/gitenv/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/gitenv	0.244s
gitenv_exit=0
```

패키지 전체 7.0% 는 `-run` 으로 좁힌 실행의 수치라 판정 근거로 쓰지 않는다.

### E5 — `git config --file` 종료 코드 탐침 (F2·N1 근거)

```
git version 2.54.0 (Apple Git-157)
existing, no match: exit=1
missing file: exit=1
warning: unable to access '.../t1204-audit2/f2/unreadable': Permission denied
unreadable file: exit=1
fatal: bad config line 1 in file .../t1204-audit2/f2/malformed
malformed file: exit=128
error: invalid key pattern: [
bad regexp: exit=6
```

128·6 은 이제 치명 처리된다(F2 닫힘). 부재·권한 거부는 1 이라 「일치 없음」과 구별되지 않는다(N1).

### E6 — 정적 검사·보안 탐침·범위

```
$ go vet ./internal/cli/
vet_exit=0
$ gofmt -l <3 files>
gofmt_exit=0
$ golangci-lint run ./internal/cli/
lint_exit=0
0 issues.
$ git diff --stat df526c9a9..HEAD -- internal/homestate internal/core
diffstat_exit=0
$ git diff df526c9a9..HEAD -- internal/ | grep -nEi '^\+.*(token|secret|password|api[_-]?key|os\.Setenv|exec\.Command\(.*\+)'
sec_probe_exit=1 (1 = no match)
$ git diff df526c9a9..HEAD -- go.mod go.sum | wc -l
       0
```

N2 생산자 경로 확인(읽기 전용):

```
$ grep -n -A12 'func runGitRevParse' internal/core/git/checkout.go
115:func runGitRevParse(dir string, args ...string) (string, error) {
116-	full := append([]string{"-C", dir, "rev-parse"}, args...)
117-	cmd := ExecCommand("git", full...)
118-	var stdout, stderr bytes.Buffer
119-	cmd.Stdout = &stdout
120-	cmd.Stderr = &stderr
121-	if err := cmd.Run(); err != nil {
```

`cmd.Env` 설정이 없다.

### E7 — verdict.md 구조와 수치 재측정 (F3)

```
$ grep -rln 'exec.Command("git"' internal/cli --include='*_test.go' | wc -l
      53
$ grep -rln 'exec.Command("git"' --include='*_test.go' internal pkg cmd | wc -l
     136
$ grep -rln 'exec.Command("git"' --include='*_test.go' internal pkg cmd | xargs grep -L 'gitenv' | wc -l
     134
--- verdict.md section headers
5:## Claim
28:## Evidence
177:## Baseline-attribution
190:## Gaps
203:## Residual-risk
--- "~32" remnant?
exit=1 (1 = none)
--- other runInit callers in tests
      17
```

## Baseline-attribution

모든 측정은 이 세션에서 워크트리 `.claude/worktrees/t1204`, HEAD `343dbdc57`, 깨끗한 작업 트리(`status=[]`, E3 첫 줄) 기준이다. 변이는 `-overlay` 로만 주입했고 추적 파일은 이 보고서 추가 외에 바꾸지 않았다. 피해 저장소 `victim_sib`·`linked/`·`f2/` 는 이 실행에서 새로 만들었다. 실행 스크립트 `mk.py`·`run.sh`·`static.sh`·`f2.sh`·`counts.sh` 와 원본 로그(`MG.log`·`MG0.log`·`R3.log`·`R4.log`·`tests.log`·`lint.log`)는 세션 스크래치패드에 있으며 커밋되지 않으므로 인용 수명은 이 세션에 한정된다.

## Gaps

- 임시 경로 밖 실제 저장소를 `GIT_DIR` 로 둔 N2 모양은 측정하지 않았다(실제 `~/.moai` 를 건드리게 되므로). 「피해 없음」은 추론이다.
- N2 가 기준 코드에서도 나는지는 이 테스트로 잴 수 없다. 기준본은 `:274` 에서 먼저 실패해 `runInit` 에 닿지 않는다(1차 E6). 다른 `runInit` 호출 테스트 17개로는 재지 않았다.
- `GIT_WORK_TREE` 단독 상속은 여전히 시험하지 않았다. verdict 의 기록대로 `GIT_WORK_TREE` 를 함께 두면 변이가 `core.bare` 를 뒤집지 않는다는 점은 이 감사에서 재현하지 않았다.
- Linux·Windows, `go test ./...` 는 저장소 규칙에 따라 CI 에 맡긴다.
- 교차 모델 감사(codex·GLM)는 돌리지 않았다.

## Residual-risk

- N2 가 남아 있는 한, 링크된 워크트리 안에서 git 훅이 `go test ./internal/cli/...` 를 부르면 주 체크아웃에 `.moai/db/` 가 생길 수 있다(외부 트리가 임시 경로 안일 때 재현됨).
- 저장소 전체에 스크럽 없는 git 을 쓰는 테스트 파일 후보가 134개 남아 있다. 결함 수가 아니라 텍스트 패턴 수다.
- 새 테스트는 `GIT_DIR` 단독 모양만 단언한다. `GIT_WORK_TREE` 가 함께 있는 모양의 다른 파괴 형태는 덮지 않는다.

## 권고

- N2 를 후속 카드로 올린다. 생산자는 `internal/core/git/checkout.go:115` `runGitRevParse` 이고, 쓰기 지점은 `internal/homestate/paths.go:280` 이다. 결정할 것은 `ResolveGitDirs` 가 상속된 `GIT_DIR` 를 따를지 여부다.
- N1 은 병합 전에 고쳐도 되고 그대로 두어도 된다. 고친다면 주석 한 줄을 좁히는 것으로 충분하다.
