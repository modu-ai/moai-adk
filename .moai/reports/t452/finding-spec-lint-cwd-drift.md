# 발견 — cwd 드리프트 하에서 `moai spec lint` 가 공허한 초록을 낸다

t452 sync 단계에서 부수적으로 관측했다. 이 카드의 판정에는 영향이 없다(유효한 측정으로 대체했다).
카드 발행 여부는 운영자 소관이며, 리드가 카드 후보로 올렸다.

## 관측

워크트리 `.claude/worktrees/t452` 에서 인자 없이 돌린 전역 `moai spec lint` 가
`0 error(s), 4306 warning(s)` / exit 0 을 냈는데, 출력에 `CODEX-SKILL-LOADER` 가 **0 줄** 등장한다.

```
$ git rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t452

$ ls /Users/goos/MoAI/moai-adk-go/.moai/specs/SPEC-CODEX-SKILL-LOADER-001/
ls: … No such file or directory          ← 이 SPEC 은 카드 브랜치에만 있다

$ ls .moai/specs/SPEC-CODEX-SKILL-LOADER-001/
acceptance.md  plan.md  progress.md  spec.md

$ timeout 90 moai spec lint .moai/specs/SPEC-CODEX-SKILL-LOADER-001/spec.md
0 error(s), 13 warning(s)                ← 파일 경로로 겨누면 13 건이 나온다
```

경고 13 건을 내는 SPEC 이 전역 실행에서 0 줄이면, 그 실행은 이 트리를 읽지 않았다.

## 기제 — 첫 귀속이 틀렸다. 정정해 기록한다

**처음에 나는 「CLI 에 MCP 의 `project_root` 같은 방어가 없다」고 귀속했다. 그것은 틀렸다.**
증상만 보고 기제를 추론했고 소스를 읽지 않았다. 리드가 소스로 반증했고, 나도 직접 확인했다:

```go
// internal/cli/spec_lint.go
cwd, err := os.Getwd()
baseDir := detectBaseDir(cwd)

// :150
func detectBaseDir(cwd string) string {
	specsDir := filepath.Join(cwd, ".moai", "specs")
	if _, err := os.Stat(specsDir); err == nil {
		return specsDir
	}
	return cwd
}
```

`detectBaseDir` 는 cwd 의 `.moai/specs` 가 있으면 그것을 쓴다. 이 워크트리에는 그 디렉터리가
실재하므로(위 `ls` 가 4 파일을 냈다), **cwd 가 워크트리였다면 워크트리를 읽었을 것이다.**
이 함수는 옳게 동작한다 — 여기를 고치러 가면 안 된다.

**실제 기제는 실행 컨텍스트다.** 백그라운드로 넘어간 그 실행의 cwd 가 워크트리가 아니었다.

## 그래서 남는 결함의 형태

**cwd 가 어긋난 실행이 오류 없이 exit 0 을 내고, 대상 SPEC 이 그 트리에 없으면
「없어서 안 나온 것」과 「봤는데 깨끗한 것」이 출력상 구분되지 않는다.**
어느 트리를 읽었는지가 출력 어디에도 없다.

MCP 쪽이 `project_root` 로 막아 둔 것과 **같은 위험**이되(그 문서도 "없다고 보고되지 않고
그냥 부재한다"고 적는다), 원인은 인자 부재가 아니라 실행 컨텍스트다.

## 이 카드에 미친 영향

없다. 그 `0 errors` 를 근거로 쓰지 않았고, 트리 안에서 파일 경로로 겨눈 측정
(`0 error(s), 13 warning(s)`, exit 0)으로 대체해 판정했다. 그것을 근거로 썼다면
부재를 통과로 읽는 공허한 초록이 됐을 것이다.

**여전히 미관측**: 이 워크트리 전체 범위의 `moai spec lint`(120 초 초과로 판독 못 함).
