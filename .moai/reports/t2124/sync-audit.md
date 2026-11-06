auditor-model: claude-opus-5-5

## Evaluation Report
SPEC: 없음 (factory 카드 t2124, Class B)
종합 판정: PASS (점수 0.93)

verdict: PASS
audited_sha: d46e8464e0bba4ec3592fdb71cf9060ea5e2c517

감사 범위: 04d8c27a7e..d46e8464e0 (코드 커밋 0fd92f9b5e, d46e8464e0은 보고서만 추가). 작업 트리: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2124, 브랜치 WT-pr-codex-repair. 반복 이력: 이 카드의 첫 sync 감사(1회차).

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | 기준 코드에서 카드 시험 FAIL, HEAD에서 PASS (아래 E1, E2). 집중 시험 묶음 `ok ... 15.488s` (E3). 상대 경로 gitdir 변형도 갱신됨 (E5-A) |
| Security (25%) | 95/100 | PASS | 변경은 `os.Lstat` 한 번 추가(크기·수정 시각만 읽음, 내용·실행 없음). diff 위험 호출 grep 0건, `grep_exit=1` (E8) |
| Craft (20%) | 90/100 | PASS | `gofmt -l` 빈 출력, `go vet` 출력 없음 (E3). `stampGitdirTarget 84.6%` (E9). 변이체 2개 모두 카드 시험이 잡음 (E6, E7) |
| Consistency (15%) | 95/100 | PASS | 같은 파일의 `.git` 디렉터리 분기(`stampFile(b, filepath.Join(path, "config"))`)와 bare 분기(`stampFile(b, filepath.Join(p, "config"))`)와 같은 방식. 기존 `stampFile` 재사용, 새 추상화 없음 |

필수 통과 차원(Functionality, Security) 모두 기준 충족. blocking 결함 없음.

### Findings (structured defect-list)
- F1 [Low, 확신 높음] [optional] [touched-file] internal/homestate/location_memo.go:250 `stampFile` — 크기+수정 시각만 지문에 남기므로, 같은 길이로 내용이 바뀌고 수정 시각이 원래대로 되돌려지면 지문이 그대로라 오래된 답이 남는다. 재현: E5-E (`fpChanged=false stale=true`). 이번 카드가 만든 한계가 아니라 파일 안 모든 `stampFile` 사용처에 공통인 기존 설계다. 일반 `git config` 쓰기는 수정 시각을 바꾸므로 Codex P2가 지적한 경로는 해결됨. - Required fix: 없음(선택). 필요하면 후속 카드에서 내용 해시 여부를 비용과 함께 판단.
- F2 [Low, 확신 중간] [optional] [touched-file] internal/homestate/location_memo.go:227 `stampGitdirTarget` — 대상 디렉터리의 `config`만 찍고 `config.worktree`(extensions.worktreeConfig), `include.path`로 끌어오는 파일, commondir 쪽 `config`는 찍지 않는다. 관측: 링크된 작업 트리에서 `config --worktree core.worktree <other>`를 바꿔도 `CanonicalProjectRoot`의 새 답 자체가 바뀌지 않아(정규 루트=주 저장소) 오래된 답은 관측되지 않음 (E5-D `fpChanged=false stale=false`). `absgitdir-context`, `isgitdir` 같은 다른 memo 연산에서는 시험하지 않았다. - Required fix: 없음(선택). 다른 memo 연산에서 재현되면 그때 카드로.
- F3 [Info, 확신 낮음] [optional] [touched-file] `stampFile`은 `os.Lstat`이라 `config`가 심볼릭 링크면 링크 대상 변경을 따라가지 않는다. 시험하지 않음. - Required fix: 없음.
- F4 [Info] [optional] [outside] MCP 서버와 설치된 moai 바이너리가 04d8c27a7e(HEAD의 조상) 빌드다. 변경은 internal/homestate 위치 memo뿐이고 audit-plan/codex_audit 경로와 무관하지만 도구 출처 좌표로 기록한다 (E10). - Required fix: 없음.

### Recommendations
- 별도 조치 없이 병합 가능. F1·F2는 선택 사항이며 자동 수리 대상이 아니다.

## Evidence-bearing report

### Claim
1. 새 시험은 기준 코드(04d8c27a7e의 location_memo.go)에서 실패하고 HEAD에서 통과한다.
2. 집중 시험 묶음, gofmt, vet이 HEAD에서 깨끗하다.
3. 상대 경로 gitdir, config 없는 대상, 링크된 작업 트리 안정성 경계에서 수정이 올바르게 동작한다.
4. 카드 시험은 잘못된 경로를 찍는 변이체를 잡는다.
5. progress.md의 Gaps는 정직하다(전체 패키지 실행 시간 초과를 PASS로 적지 않음).
6. 감사 계획이 요구하는 codex 필수 관문을 충족했다.

### Evidence
E1 RED 재현 — 작업 트리는 건드리지 않고 `go test -overlay`로 location_memo.go만 기준 버전으로 바꿔 실행. 기준과 HEAD 파일 차이:
```
$ diff location_memo_base.go internal/homestate/location_memo.go
243a244,245
> 	// The directory's config can move the answer too (`core.worktree`), with HEAD and commondir untouched (card t2124).
> 	stampFile(b, filepath.Join(target, "config"))
$ go test -count=1 -timeout 150s -overlay overlay_base.json -v -run 'CoreWorktreeChangedInAnExternal' ./internal/homestate/
exit=1
=== RUN   TestLocationMemo_CoreWorktreeChangedInAnExternalGitDir
    location_memo_gitdir_config_t2124_test.go:42: CanonicalProjectRoot = "/private/var/folders/.../001/wt" after core.worktree changed, a fresh lookup answers "/private/var/folders/.../001/other" (a stale remembered answer; before = "/private/var/folders/.../001/wt")
    location_memo_gitdir_config_t2124_test.go:45: fingerprint unchanged after the external git directory's config changed
--- FAIL: TestLocationMemo_CoreWorktreeChangedInAnExternalGitDir (0.33s)
FAIL	github.com/modu-ai/moai-adk/internal/homestate	0.894s
```
(임시 디렉터리 경로 중간 부분만 `...`로 줄였고 나머지는 그대로.)

E2 HEAD GREEN:
```
$ go test -count=1 -timeout 150s -v -run 'CoreWorktreeChangedInAnExternal' ./internal/homestate/
=== RUN   TestLocationMemo_CoreWorktreeChangedInAnExternalGitDir
--- PASS: TestLocationMemo_CoreWorktreeChangedInAnExternalGitDir (0.44s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/homestate	0.810s
```

E3 집중 시험·서식·vet:
```
$ go test -count=1 -timeout 150s -run 'LocationMemo|CanonicalProjectRoot|ProjectKey|OutsideRepository' ./internal/homestate/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/homestate	15.488s
$ gofmt -l internal/homestate/
gofmt_exit=0        (출력 없음)
$ go vet ./internal/homestate/
vet_exit=0          (출력 없음)
```

E5 경계 사례 — 감사자 임시 시험 파일(스크래치에만 있고 overlay로 패키지에 주입, 트리에 없음) `go test -overlay overlay_edge.json -v -run 'TestAudit_'`:
```
A (상대 경로 "gitdir: ../meta.git"):
  control rev-parse with relative gitdir: ".../001/wt" err=<nil>
  A relative: before=".../001/wt" got=".../001/other" fresh=".../001/other" fpChanged=true
  --- PASS: TestAudit_A_RelativeGitdir (0.53s)
B (config 없는 대상 디렉터리):
  B stamp after .git content: "{H-[-]}"
  --- PASS: TestAudit_B_TargetWithoutConfig (0.01s)
C (링크된 작업 트리 관리 디렉터리, config 없음, 두 번 호출):
  C stamp after .git content: "{Hc[-]}"
  C admin config Lstat err=lstat .../repo/.git/worktrees/lw/config: no such file or directory; ... stable=true; a=".../001/repo" b=".../001/repo"
  --- PASS: TestAudit_C_LinkedWorktreeStable (0.39s)
D (링크된 작업 트리, extensions.worktreeConfig + config --worktree core.worktree):
  D config.worktree: before=".../001/repo" got=".../001/repo" fresh=".../001/repo" fpChanged=false stale=false
E (같은 길이 내용 변경 + 수정 시각 복원):
  E same-size+mtime-restored: before=".../001/oth1" got=".../001/oth1" fresh=".../001/oth2" fpChanged=false stale=true
```
B와 C의 첫 실행은 감사자 시험의 줄 추출 오류(.git 내용의 줄바꿈 때문에 `{...}` 표식이 다음 줄에 있음)로 B가 FAIL로 찍혔다. 추출을 고친 재실행 결과가 위 값이며, 지문 문자열 자체는 두 번 다 같았다(`stable=true`).

E6 변이체 1 (`"config"` → `"configx"`):
```
$ go test -count=1 -timeout 150s -overlay overlay_mut1.json -run 'CoreWorktreeChangedInAnExternal' ./internal/homestate/
--- FAIL: TestLocationMemo_CoreWorktreeChangedInAnExternalGitDir (0.23s)
    location_memo_gitdir_config_t2124_test.go:45: fingerprint unchanged after the external git directory's config changed
FAIL	github.com/modu-ai/moai-adk/internal/homestate	0.702s
```
E7 변이체 2 (대상의 부모 디렉터리 config를 찍음):
```
$ go test -count=1 -timeout 150s -overlay overlay_mut2.json -run 'CoreWorktreeChangedInAnExternal' ./internal/homestate/
--- FAIL: TestLocationMemo_CoreWorktreeChangedInAnExternalGitDir (0.66s)
    location_memo_gitdir_config_t2124_test.go:45: fingerprint unchanged after the external git directory's config changed
FAIL	github.com/modu-ai/moai-adk/internal/homestate	1.382s
```
두 변이체 모두 카드 시험이 죽였다(42행 오래된 답 단언도 함께 실패).

E8 보안 탐침:
```
$ git diff 04d8c27a7e d46e8464e0 -- internal/ | grep -nE '^\+.*(exec\.|os\.(Open|Remove|Write|Chmod)|ReadFile|Command\(|password|token|secret)'
grep_exit=1   (일치 없음)
```

E9 함수별 커버리지(LocationMemo 시험만):
```
$ go test -count=1 -timeout 150s -run 'LocationMemo' -coverprofile=t2124_lm.cov ./internal/homestate/
ok  	github.com/modu-ai/moai-adk/internal/homestate	6.405s	coverage: 5.4% of statements
location_memo.go:227:	stampGitdirTarget	84.6%
location_memo.go:250:	stampFile		100.0%
location_memo.go:193:	stampDotGit		77.8%
```
(5.4%는 패키지 전체 기준이고 이 시험 부분집합만 돌린 값이라 패키지 커버리지 주장이 아니다.)

E10 감사 계획과 codex 관문:
```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2124
"config_status": "ok", "model": "codex",
backends: claude gate "off", codex gate "required", glm gate "off"
"cross_model_active": true, "cross_model_required": true, "enforced_required": ["codex"]
exit=0
```
`mcp__moai__codex_audit` 1회 호출(project_root=작업 트리, target=baseBranch, delta_base=04d8c27a7ed77afc09217f8623519806d48df3c4, mode=native). 결과 필드 그대로: 판정 `pass`, `findings: []`, `next_steps: []`, `resolved_model: gpt-6.1-sol`, `transport: codex app-server`(MCP 서버 내부 전송), `review_base: delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)`, `audit_receipt: rcpt-5cfdeae2381bccf6460e80c0`, `build_commit: 04d8c27a7e`, `build_lag: moai binary lag: the installed binary was built from commit 04d8c27a7, an ancestor of this tree's HEAD d46e8464e.` 요약 문구는 정적 검토만 했고 시험·빌드·재현은 실행하지 않았다고 밝힘.

E11 작업 트리 상태(감사 중 쓰기 없음 확인):
```
$ git status --short
status_exit=0   (출력 없음)
$ git rev-parse HEAD
d46e8464e0bba4ec3592fdb71cf9060ea5e2c517
```

### Baseline-attribution
- 모든 시험·gofmt·vet은 이번 실행에서 이 작업 트리(HEAD d46e8464e0)에 대해 직접 돌렸다. RED 기준은 `git show 04d8c27a7e:internal/homestate/location_memo.go`로 뽑은 파일을 overlay로 주입한 것이며, 위 diff로 두 파일 차이가 수정 두 줄뿐임을 확인했다.
- 시험은 Go 툴체인이 트리 소스로 직접 빌드하므로 도구 출처 지연이 없다. moai CLI/MCP 측 측정(audit-plan, codex_audit)은 04d8c27a7e 빌드로 수행됐고, 이는 HEAD의 조상이다. 04d8c27a7e..HEAD의 변경은 internal/homestate 위치 memo뿐이라 감사 계획·codex 경로 동작과 무관하다고 판단하지만, 이 판단은 코드 읽기에 근거한 것이다.

### Gaps
- 감사 계획이 `cross_model_active: true`라 원칙상 `audit_multi` 경로지만, 리더 지시(audit_multi 금지, codex_audit 최대 1회)에 따라 codex 단일 호출로 대신했다. 필수 관문은 codex 하나뿐이고 claude·glm은 `off`라 관문 충족에는 영향이 없다고 보지만, audit_multi 수렴 결과는 없다.
- 전체 `go test ./internal/homestate/`는 지시대로 돌리지 않았다. progress.md도 전체 패키지 실행이 600.5초 시간 초과(TestRebindRollsBackWhenAWriteIsRefused 실행 중)로 끝났다고 Gap에 적었고, 이를 PASS로 주장하지 않았으며 기존 문제라고도 단정하지 않았다 — 기록은 정직하다. 패키지 전체 통과 여부는 여전히 미관측이다.
- 패키지 커버리지 85% 기준은 측정하지 않았다(부분집합 커버리지만 E9).
- race 검출기, golangci-lint, 다른 OS(Linux/Windows) 미실행.
- F2의 `absgitdir-context`, `isgitdir` memo 연산, F3의 심볼릭 링크 config는 시험하지 않았다.
- 거부된 도구 호출: 셸 명령 일부가 작업 트리 격리 가드에 거부됐다(문구 그대로: "this command names git in a form too complex to verify that it stays inside the worktree. Refusing to run it"; 다른 한 건은 "runs sed with a value computed at runtime (the variable F) ..."). 같은 작업을 리터럴 경로의 단순 명령으로 나눠 다시 실행했고, 측정 결과는 그 재실행에서 나왔다. 대체가 아니라 같은 명령의 분할 실행이다.

### Residual-risk
- 크기+수정 시각 지문 특성상 같은 길이·같은 수정 시각의 config 변경은 놓친다(F1, 재현됨). 실제 `git config` 경로에서는 일어나기 어렵다.
- 수정 시각 해상도가 낮은 파일 시스템(일부 네트워크 FS)에서는 같은 크기의 빠른 연속 변경이 같은 지문을 낼 수 있다. macOS APFS(나노초)에서만 확인했다.
- 링크된 작업 트리에서 `config.worktree`·include 파일·commondir config 변경이 다른 memo 연산에서 오래된 답을 낼 가능성은 남아 있다(F2, 미관측).
- 기록 파일(progress.md, card-review.md)이 커밋 d46e8464e0에 들어갔으나 이 감사의 판정에는 코드 커밋만 영향을 준다.
