# t516 — moai gate: 자식 프로세스가 호스트 저장소에 쓰는 결함 (GH #1691)

- 카드: t516 (Tier M, Class B — plan 생략, 원인 규명은 run 소관)
- 워크트리: `.claude/worktrees/t516` · 브랜치 `WT-gate-host-writes`
- 기준: `origin/develop` tip `bf779ecf2`
- 측정 환경: darwin 25.6.0 · `git version 2.50.1 (Apple Git-155)` · `go test` 로컬

---

## Claim

1. **결함 2 (호스트 저장소 오염) 는 실재하며, 근본 원인은 `runStep` 이 자식 프로세스 환경을
   격리하지 않는 것이다.** `internal/hook/quality/gate.go` 의 `runStep` 은 `cmd.Dir` 만 설정하고
   `cmd.Env` 를 설정하지 않아, Go `os/exec` 규칙에 따라 부모 환경을 통째로 물려준다. `moai gate`
   를 호출하는 주체가 git pre-commit 훅이고, git 은 훅에 `GIT_DIR` / `GIT_INDEX_FILE` 을 export
   하며, 이 변수들은 작업 디렉터리보다 **우선**한다. 따라서 `cmd.Dir` 은 자식을 가두지 못한다.
2. **누출은 두 개의 독립된 쓰기 경로로 호스트에 도달한다** — 커밋 경로와 `git init` 경로.
3. **결함 1 (`core.bare` 뒤집힘) 의 「소거법 귀속」은 미검증이다** — 반증된 것이 아니라, 이 환경에서
   재현되지 않았다. MoAI 코드 중 `core.bare` 를 쓰는 곳은 **없다**.
4. 수리(저장소-위치 git 환경변수 스크럽) 후 세 재현이 모두 통과하며, 뮤턴트로 가드의 비공허성이
   증명된다.

---

## Evidence

### E1 — 근본 원인 지점 (수리 전 `gate.go`)

`runStep` 은 `exec.CommandContext` 로 자식을 만들고 `cmd.Dir` 만 설정한다. `cmd.Env` 대입은
파일 전체에 존재하지 않았다.

```
$ grep -n "exec.Command\|cmd.Env\|cmd.Dir" internal/hook/quality/gate.go
1065:	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--name-only")
1066:	cmd.Dir = dir
1139:	cmd := exec.CommandContext(stepCtx, name, args...)
1159:		cmd.Dir = dir
```

방어 코드 부재:

```
$ grep -rn "GIT_DIR\|GIT_WORK_TREE\|GIT_INDEX_FILE" --include='*.go' internal/
(출력 없음)
```

호출 주체가 pre-commit 훅이라는 근거 — `internal/cli/hook_install_precommit.go`
`preCommitHookContent` 말미:

```sh
if command -v moai >/dev/null 2>&1; then
    if ! MOAI_PRECOMMIT=1 moai gate; then
```

### E2 — 결함 2 재현 (RED, 수리 전)

```
$ go test ./internal/hook/quality/ -run 'TestRunStep_ChildFixtureCommitDoesNotLandInOuterRepo' -count=1 -v
=== RUN   TestRunStep_ChildFixtureCommitDoesNotLandInOuterRepo
    gate_step_git_env_test.go:274: the gate step's child wrote a commit into the OUTER repository: 1 commits before, 2 after.
        latest outer commit: 0870840 t516 throwaway fixture commit
    gate_step_git_env_test.go:278: the fixture commit did not land in the step's own repository either (1 commits before and after) — the child wrote somewhere else entirely
--- FAIL: TestRunStep_ChildFixtureCommitDoesNotLandInOuterRepo (1.51s)
```

이슈가 보고한 사고와 동일한 모양: 자식이 자기 디렉터리에서 만든 throwaway fixture 커밋이
**바깥 저장소**에 착지했다.

### E3 — `git init` 경로도 호스트에 도달 (RED, 수리 전)

```
$ go test ./internal/hook/quality/ -run 'TestRunStep_ChildGitInitDoesNotFlipOuterRepoCoreBare' -count=1 -v
    gate_step_git_env_test.go:334: child `git init` under a leaked GIT_DIR reported:
        err=<nil>
        Reinitialized existing Git repository in /private/var/folders/.../001/host/.git/
--- PASS: TestRunStep_ChildGitInitDoesNotFlipOuterRepoCoreBare (1.00s)
```

주의 — 이 시점의 PASS 는 **공허한 초록이었다.** 자식은 실제로 바깥 저장소를 재초기화했는데
(`Reinitialized existing`), 그때 단언은 `core.bare` 만 보고 있었다. 단언을 "재초기화 자체" 로
강화한 뒤에야 이 사실이 판정에 반영된다 (E6 참조).

### E4 — `core.bare` 를 쓰는 MoAI 코드 부재 (대조군 포함)

```
$ grep -rn "core\.bare" . --exclude-dir=.git
(출력 없음)

$ grep -rn "core\.hooksPath" . --exclude-dir=.git | head -5
CHANGELOG.md:30:...
.moai/specs/SPEC-FMT-GATE-001/plan.md:9:...
.moai/specs/SPEC-FMT-GATE-001/spec.md:28:...
.moai/specs/SPEC-FMT-GATE-001/spec.md:47:...
.moai/specs/SPEC-FMT-GATE-001/spec.md:52:...
```

확장자 필터 없이 트리 전체를 훑었고, 같은 형태의 대조 질의는 매치를 낸다. 즉 무출력은 도구의
침묵이 아니라 **부재의 보고**다.

### E5 — 수리 (GREEN)

`internal/hook/quality/step_git_env.go` (신규) — 저장소 **위치** 변수 11개만 제거하고 신원
(`GIT_AUTHOR_*`) · 동작 (`GIT_EDITOR`, `GIT_CONFIG_GLOBAL` …) 변수는 통과시킨다.
`internal/hook/quality/gate.go` `runStep` 에 `cmd.Env = stepEnv()` 한 줄.

```
$ go vet ./internal/hook/quality/ && echo VET-OK
VET-OK

$ go test ./internal/hook/quality/ -run 'TestRunStep_(DoesNotLeakOuterGitEnvToChild|ChildFixtureCommitDoesNotLandInOuterRepo|ChildGitInitDoesNotFlipOuterRepoCoreBare)' -count=1 -v
--- PASS: TestRunStep_DoesNotLeakOuterGitEnvToChild (0.43s)
--- PASS: TestRunStep_ChildFixtureCommitDoesNotLandInOuterRepo (1.54s)
    gate_step_git_env_test.go:348: child `git init` reported:
        err=<nil>
        Initialized empty Git repository in /private/var/folders/.../001/step/.git/
--- PASS: TestRunStep_ChildGitInitDoesNotFlipOuterRepoCoreBare (0.93s)
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	3.513s
```

`Reinitialized existing … host/.git` → `Initialized empty … step/.git` 로 바뀐 것이 수리의
직접 증거다.

### E6 — 뮤턴트 (가드 비공허성)

`gitRepoScopingEnvVars` 에서 `"GIT_DIR",` 한 줄만 제거:

```
--- FAIL: TestRunStep_DoesNotLeakOuterGitEnvToChild (0.61s)
    gate step child inherited GIT_DIR="…/001/host/.git" from the caller
--- FAIL: TestRunStep_ChildFixtureCommitDoesNotLandInOuterRepo (1.33s)
    the gate step's child wrote a commit into the OUTER repository: 1 commits before, 2 after.
    latest outer commit: dd1730e t516 throwaway fixture commit
--- FAIL: TestRunStep_ChildGitInitDoesNotFlipOuterRepoCoreBare (0.52s)
    the gate step's child re-initialised the OUTER repository (…/001/host)
FAIL
```

세 테스트 모두 뮤턴트를 잡았다. 목록의 그 한 줄이 실제로 하중을 받는다.

**부수 관측**: 뮤턴트 상태에서도 `core.bare` 단언은 여전히 침묵했다. 테스트 C 를 잡아낸 것은
E3 이후 추가한 "재초기화" 단언뿐이다 — 원래 단언만 두었다면 이 가드는 공허했다.

### E7 — 회귀 (영향 패키지 전수)

```
$ go test ./internal/hook/quality/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	18.551s

$ go test ./internal/hook/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/hook	61.731s

$ go test ./internal/cli/ -count=1 -timeout 900s
ok  	github.com/modu-ai/moai-adk/internal/cli	559.825s

$ gofmt -l internal/hook/quality/
(출력 없음)
```

### E8 — 이 저장소 무접촉 (리드 [HARD] #2)

재현은 전부 `t.TempDir()` 안의 격리 저장소에서 돌았다. 실행 후 이 워크트리:

```
$ git status --short
 M internal/hook/quality/gate.go
?? internal/hook/quality/gate_step_git_env_test.go
?? internal/hook/quality/step_git_env.go

$ git rev-list --count HEAD
6440

$ git log --oneline -3
bf779ecf2 Merge branch 'WT-codex-model-config' into develop
784912910 Merge branch 'develop' into WT-codex-model-config
1aaf4951f docs(t509): closed — landed on develop, and the lead's two self-corrections

$ git config --get core.bare
false
```

떠돌이 커밋 0건, HEAD 는 기준 `bf779ecf2` 그대로, `core.bare` 는 `false` 유지, 변경 파일은
의도한 3개뿐이다.

---

## Baseline-attribution

모든 수치는 이 실행에서, 이 트리(`.claude/worktrees/t516` @ `bf779ecf2` + 미커밋 3파일)에
대해 직접 측정했다. 다른 패키지·트리·시점의 수치를 옮겨 온 것은 없다.

- RED 기준선: 수리 커밋 **이전**의 작업 트리에서 실행 (E2/E3)
- GREEN 기준선: `cmd.Env = stepEnv()` 적용 후 동일 트리 (E5)
- 뮤턴트 기준선: GREEN 트리에서 `step_git_env.go` 한 줄만 되돌린 상태 (E6)
- 회귀 기준선: 뮤턴트 복원 후 최종 트리 (E7)

---

## Gaps

관측하지 **않은** 것:

1. **결함 1 (`core.bare`) 은 재현되지 않았다.** git 2.50.1 (Apple Git-155) / darwin 에서, 누출된
   `GIT_DIR` 아래 `git init` 은 호스트를 재초기화하되 `core.bare` 를 **바꾸지 않았다**. 제보자 환경
   (Linux 6.18 WSL2 · 네이티브 Linux · moai-adk 3.1.2) 은 재현하지 않았다. 따라서
   「소거법 귀속」은 **미검증**이지 반증이 아니다. 다만 두 사실이 함께 성립한다 — MoAI 코드는
   그 키를 쓰지 않고(E4), 누출 경로는 호스트 저장소에 실제로 도달한다(E3). 이 카드는 도달 경로를
   막았으므로, 뒤집힘의 저자가 무엇이든 그것이 **게이트 자식을 통해** 오는 길은 닫혔다.
2. **실제 pre-commit 훅 end-to-end 실행은 관측하지 않았다.** git 이 `GIT_DIR` 을 훅에 export
   한다는 것은 문서·동작 지식이고, 이 카드의 재현은 그 환경을 `t.Setenv` 로 **모사**했다. 훅을
   실제로 걸어 `moai gate` 를 통과시키는 실측은 하지 않았다.
3. **`gate.go:1065` 의 `git diff --cached` 경로는 건드리지 않았다.** 그 호출은 게이트 자신이
   staged 파일을 읽는 경로라 pre-commit 의 `GIT_DIR` 상속이 정상 동작이다. 이 판단은 코드 독해에
   근거하며, 스크럽을 그쪽에 적용했을 때의 동작은 측정하지 않았다.
4. **pytest 로 직접 재현하지 않았다.** 재현 자식은 Go 테스트 바이너리가 fixture 커밋을 흉내 낸
   것이다. 누출은 언어 무관(환경변수 상속)이지만, 실제 pytest 스위트로는 재측정하지 않았다.
5. **CI 판정은 없다.** 위 초록은 전부 로컬 darwin 단일 플랫폼이다. windows / linux 매트릭스와
   전체 스위트 판정은 develop push 가 일으키는 CI 실행 몫이다.
6. **커버리지 수치는 측정하지 않았다.**

---

## Residual-risk

- **스크럽 범위가 좁아서 남는 위험**: `GIT_CONFIG_COUNT` / `GIT_CONFIG_KEY_n` / `GIT_CONFIG_VALUE_n`
  으로도 자식 git 의 설정을 바깥에서 주입할 수 있다. 의도적으로 남겼다 — 이 변수들은 저장소
  **위치**가 아니라 **동작**을 지정하고, 지우면 그것을 의도적으로 쓰는 CI 의 게이트 동작이 바뀐다.
  다만 이 경로로 `core.bare` 를 세우는 것은 원리상 가능하다.
- **신원 변수를 남긴 대가**: `GIT_AUTHOR_*` 가 살아 있으므로, 자식 fixture 커밋은 여전히 바깥
  세션의 신원으로 만들어진다. 이제 그 커밋이 **자기 저장소**에 떨어지므로 무해하지만, 커밋 저자가
  게이트를 부른 사람으로 찍히는 것은 그대로다.
- **`cmd.Env` 명시가 유발할 수 있는 회귀**: 이제 `runStep` 자식은 `os.Environ()` 스냅샷을 받는다.
  이전에는 `nil` 상속이었으므로 실질 동작은 같지만, 프로세스가 실행 중 `os.Setenv` 를 호출하는
  경우 스냅샷 시점이 달라질 수 있다. 영향 패키지 전수 통과(E7)로 관측된 회귀는 없다.
- **다른 자식 실행 지점**: 이 카드는 `runStep` 한 곳만 고쳤다. 저장소 안의 다른
  `exec.Command` 호출자들이 같은 누출을 갖는지는 훑지 않았다 — 범위 규율(카드가 지목한 게이트
  자식)을 지킨 결과이며, 별도 스윕 카드가 필요하다.
