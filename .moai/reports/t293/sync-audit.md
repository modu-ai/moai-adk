# sync-audit — 카드 t293 / SPEC-STATUSLINE-PROFILE-RESPECT-001

> 독립 sync-phase 품질 감사. 감사자: sync-auditor (cold start, run 수행자와 다른 세션).
> 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t293` · 브랜치 `WT-statusline-profile` · HEAD `1fb1ab78a` · run 기점 `3abde7053` · origin/develop 병합 커밋 `8ef14f5ae`.
> 작성일 2026-08-27. 모든 수치는 **이번 실행, 이 트리, 이 HEAD**에서 직접 관측한 것이다.

## 판정 요약

| 항목 | 값 |
|---|---|
| **Overall Verdict** | **PASS-WITH-DEBT** |
| 가중 조화평균 | **86.7 / 100** |
| Tier M 임계 | 80 |
| must-pass 방화벽 | Functionality 88 ✓ · Security 92 ✓ — 통과 |
| 결함 수 | HIGH 1 · MEDIUM 2 · LOW 4 · INFO 1 (blocking 0) |
| #1675 결함 폐쇄 여부 | **코드 층위는 폐쇄됨(변이 시험으로 확인). 다만 운영 효과는 아직 발효 전** — F2 참조 |

---

## 1. Claim (주장)

1. 두 옵트아웃 레버(`segments.github: false`, `statusline.forge: none|off|미인식값`)가 렌더뿐 아니라 **분리형 refresh 자식 프로세스의 생성 자체**를 막는다 (REQ-001 / REQ-002).
2. 옵트아웃이 없을 때의 전개(all-enabled fallback)는 회귀 없이 보존된다 (REQ-003).
3. `none → github` 되돌림이 캐시 삭제 없이 한 refresh 주기 안에 복구된다 (REQ-004).
4. forge pair의 4상태 표시 계약(`7/3` · `0/0` · `-/-` · 무출력)이 테스트로 고정됐다 (REQ-005).
5. 등록된 프로젝트 하위 디렉터리(미등록 워크트리 포함)가 읽기 경로에서 그 프로젝트의 프로필로 해석된다 (REQ-006/007/008).
6. AC-009 / M5(원장 쓰기 정규화)는 킥오프 결정 D1에 따른 의도적 이연이다.
7. `internal/profile` 커버리지 84.3%(<85%)는 SPEC 범위 밖 기존 미커버 함수 소관이다.

각 주장에 대한 반증 시도 결과는 아래와 같다.

---

## 2. Evidence (증거 — 명령과 실측 출력)

### 2.1 대상 패키지 테스트 (기준선)

```
$ go test ./internal/statusline/... ./internal/profile/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/statusline	14.924s
ok  	github.com/modu-ai/moai-adk/internal/profile	0.467s
rc=0
```

전체 스위트는 의도적으로 돌리지 않았다(레인 동시 실행 부하 규율). 변경면은 두 패키지 + `internal/cli` 호출부에 한정된다.

### 2.2 변이 시험 — 게이트 3개 전부 살아 있음

세 개의 프로덕션 게이트를 각각 제거하고, 해당 테스트가 실제로 무너지는지 확인했다. 세 변이 모두 사본 백업 후 적용하고, 시험 직후 원복해 `git diff --stat`이 빈 출력임을 확인했다.

**변이 A — `builder.go`의 세그먼트 spawn 게이트 제거** (`if b.renderer.isSegmentEnabled(...)` 조건절을 걷어내고 무조건 호출로 되돌림):

```
$ go test ./internal/statusline/ -count=1 -run 'SegmentGateSuppressesPairAndSpawn|NilSegmentsPreservesSpawn'
--- FAIL: TestBuilder_SegmentGateSuppressesPairAndSpawn (0.00s)
    forge_spawn_gate_test.go:181: gated render spawn attempts = 1, want 0 (REQ-001: the segment gate must reach the spawn, not just the render)
    forge_spawn_gate_test.go:189: enabled render spawn attempts = 2, want 1 — the gate must be the only delta
FAIL	github.com/modu-ai/moai-adk/internal/statusline	0.393s
rc=1
```

**변이 B — `github.go`의 명시 override 조기 반환 제거**:

```
$ go test ./internal/statusline/ -count=1 -run 'ExplicitNoForgeSpawnsNothing|TwoWayRevert|UnsetOverrideStillSpawns'
--- FAIL: TestMaybeRefreshGitHubCounts_ExplicitNoForgeSpawnsNothing (0.00s)
    --- FAIL: .../forge_none,_absent_cache (0.00s)
        forge_spawn_gate_test.go:69: spawn attempts = 1, want 0 — an explicit no-forge override must gate the refresh child
    --- FAIL: .../forge_off,_absent_cache          spawn attempts = 2, want 0
    --- FAIL: .../forge_none,_stale_cache…         spawn attempts = 3, want 0
    --- FAIL: .../a_typo_names_no_forge_either     spawn attempts = 4, want 0
--- FAIL: TestForgeOptOut_TwoWayRevert (0.00s)
    forge_spawn_gate_test.go:228: pre-revert spawn attempts = 1, want 0
FAIL	github.com/modu-ai/moai-adk/internal/statusline	0.415s
rc=1
```

**변이 C — `profile.go`의 `lookupSubtreeProjectKey` 호출 무력화**:

```
$ go test ./internal/profile/ -count=1 -run 'Subtree|Deepest|Lexical|Stale|ExactMatchBeatsSubtree'
--- FAIL: TestResolveLaunchProfileForProject_SubtreeWorktreeResolves (0.00s)
    subtree_resolve_test.go:64: subtree session resolved to "", want "alpha"
--- FAIL: TestResolveLaunchProfileForProject_DeepestRegisteredAncestorWins (0.00s)
    subtree_resolve_test.go:83: nested subtree resolved to "", want "beta" (deepest registered ancestor)
--- FAIL: TestResolveLaunchProfileForProject_StaleAncestorSkippedNotDeadEnd (0.00s)
    subtree_resolve_test.go:149: session under a stale registration resolved to "", want "alpha" (walk continues past the unusable entry)
FAIL	github.com/modu-ai/moai-adk/internal/profile	0.376s
rc=1
```

원복 확인:

```
$ git status --short
?? .moai/reports/t293/evidence/
$ git diff --stat
(빈 출력)
```

즉 세 게이트 전부 **제거하면 테스트가 무너진다**. 공허한 초록이 아니다.

### 2.3 셀렉터 정직성 (매치 수 실측)

```
$ go test ./internal/statusline/ -count=1 -v -run 'ExplicitNoForgeSpawnsNothing|SegmentGateSuppressesPairAndSpawn|TwoWayRevert'
--- PASS: TestMaybeRefreshGitHubCounts_ExplicitNoForgeSpawnsNothing (0.01s)
--- PASS: TestBuilder_SegmentGateSuppressesPairAndSpawn (0.00s)
--- PASS: TestForgeOptOut_TwoWayRevert (0.34s)
→ 최상위 3건 + 하위 4건, rc=0

$ go test ./internal/profile/ -count=1 -v -run 'Subtree|Deepest|Lexical|Stale|ExactMatchBeatsSubtree'
--- PASS: TestResolveLaunchProfile_StaleRecordIgnored
--- PASS: TestResolveForProject_StaleProjectEntrySkipped
--- PASS: TestResolveLaunchProfileForProject_SubtreeWorktreeResolves
--- PASS: TestResolveLaunchProfileForProject_DeepestRegisteredAncestorWins
--- PASS: TestResolveLaunchProfileForProject_LexicalPrefixIsNotAnAncestor
--- PASS: TestResolveLaunchProfileForProject_SubtreeMissFallsThrough
--- PASS: TestResolveLaunchProfileForProject_StaleAncestorSkippedNotDeadEnd
--- PASS: TestResolveLaunchProfileForProject_ExactMatchBeatsSubtree
→ 최상위 8건, rc=0
```

두 셀렉터 모두 0-매치가 아니다. §E.2가 인용한 RED 출력(`red-statusline.txt`, `red-profile.txt`)의 실패 테스트 이름은 위 GREEN 목록과 정확히 대응하며, 실패 메시지도 "spawn attempts = N, want 0" / "resolved to \"\", want alpha" 즉 **게이트·walk 부재**라는 명시된 사유이지 픽스처 파손이 아니다.

### 2.4 `__no_such_profile__` 도달 경로 (과제 4번)

```
$ grep -rn "__no_such_profile__" --include='*.go' --include='*.sh' --include='*.tmpl' --include='*.yaml' --include='*.json' . | grep -v '.moai/reports\|.moai/specs'
(출력 없음)

$ grep -rn "ResolveLaunchProfileForProject" --include='*.go' .
internal/cli/launcher.go:150:	resolved := profile.ResolveLaunchProfileForProject(root, profileName)
(나머지는 전부 _test.go)
```

확인된 사실:

- 리터럴 `__no_such_profile__`은 이 저장소가 생산하지 않는다 — SPEC §F의 "발원지 미상" Gap은 사실이다.
- 프로덕션 호출자는 `unifiedLaunchDefault`(`launcher.go:150`) **단 한 곳**이다. 즉 수정 경로는 `moai cc` / `moai glm` / `moai cg`의 **bare launch**(`-p` 미지정)에서만 작동하며, 이는 팩토리 레인이 실제로 타는 경로가 맞다. 워크트리가 익명 프로필로 떨어지던 증상은 이 경로에서 해소된다.
- **다만 프로필 해석은 statusline 설정 읽기와 무관하다.** `runStatusline`은 `findProjectRootFn()`으로 CWD를 거슬러 올라가 프로젝트 루트를 잡고 `<root>/.moai/config/sections/statusline.yaml`을 읽는다(`internal/cli/statusline.go:82,85,183`). 프로필 디렉터리는 이 경로에 개입하지 않는다. 따라서 이슈 표제 "폴백 프로필이 설정을 무시"는 **오귀속**이며, SPEC §A.1 F1이 이를 정확히 반증해 두었다. 감사자로서 이 반증을 독립 확인했다.

### 2.5 두 레버가 서로 다른 파일을 읽는다 (F3)

```
$ grep -n boardRoot internal/statusline/builder.go
255:		boardRoot := resolveBoardRoot(input)
265:		data.GitHub = resolveGitHubCounts(boardRoot)
267:			maybeRefreshGitHubCounts(boardRoot)

$ sed -n '34,39p' internal/statusline/backlog.go
func resolveBoardRoot(input *StdinData) string {
	if input != nil && input.Worktree != nil && input.Worktree.OriginalCwd != "" {
		return input.Worktree.OriginalCwd
	}
	return resolveProjectDir(input)
}
```

- **세그먼트 레버(REQ-001)** — `b.renderer.isSegmentEnabled`은 `Options.SegmentConfig`를 읽고, 그 값은 `findProjectRootFn()`이 잡은 **워크트리 자신의** `statusline.yaml`에서 온다.
- **forge 레버(REQ-002)** — `forgeOverride(boardRoot)`는 워크트리 세션에서 `worktree.original_cwd`, 즉 **primary 체크아웃의** `statusline.yaml`을 읽는다.

두 레버가 서로 다른 루트를 본다. `resolveGitHubCounts`가 이미 `boardRoot` 기준이었으므로 렌더와 spawn 사이의 일관성은 유지되지만, SPEC은 두 레버를 모두 "the project's `statusline.yaml`"이라고 한 문장으로 서술한다.

### 2.6 M7 운영 편집의 실제 발효 상태 (F2)

```
$ grep -n 'forge' /Users/goos/MoAI/moai-adk-go/.moai/config/sections/statusline.yaml
grep_rc=1
$ wc -l /Users/goos/MoAI/moai-adk-go/.moai/config/sections/statusline.yaml
      33 ...
```

primary 체크아웃의 `statusline.yaml`에는 `forge` 키가 **없다**(33줄, 병합 전 원본). 반면 이 워크트리의 같은 파일 12행은 `forge: "none"`이다. 워크트리 세션의 `boardRoot`는 primary이므로, **현재 시점에 팩토리 레인 세션의 gh 폴링은 멈추지 않는다.** develop→main 병합분이 primary 워킹트리에 도달한 뒤에 비로소 발효된다.

### 2.7 커버리지 귀속 검증

```
$ go tool cover -func=… | grep -E 'lookupSubtreeProjectKey|ResolveLaunchProfileForProject|GetCurrentName|Delete'
profile.go:77:	GetCurrentName			100.0%
profile.go:156:	Delete				76.9%
profile.go:387:	lookupSubtreeProjectKey		88.9%
profile.go:450:	ResolveLaunchProfileForProject	100.0%
total:				(statements)	84.3%
```

- 신규 함수 수치(88.9% / 100.0%)는 §E.2 주장 그대로 **참**이다.
- 85% 미만 함수는 13개이며 전부 기존 코드다: `IsValidPermissionMode` 0.0%, `IsValidProfileName` 0.0%, `EnsureDir` 70.0%, `GetBaseDir` 71.4%, `WritePreferences` 72.7%, `launchCandidateIsUsable` 75.0%, `Delete` 76.9%, `RecordLastUsedProfileForProject` 77.5%, `syncStatusline` 78.3% 등.
- **다만 §E.2가 지목한 `GetCurrentName`은 100.0%다** — 귀속 문구가 틀렸다. 결론("범위 밖 기존 미커버")은 유지되지만 근거 이름 하나는 반증됐다(F4).

### 2.8 나머지 폐쇄 게이트 재실행

```
$ go vet ./internal/statusline/... ./internal/profile/... ./internal/cli/     → rc=0 (무출력)
$ GOOS=windows GOARCH=amd64 go build ./...                                   → rc=0 (무출력)
$ GOOS=windows GOARCH=amd64 go vet ./internal/statusline/... ./internal/profile/... → rc=0
```

윈도 쪽은 `build`만이 아니라 `vet`까지 돌려 **테스트 코드의 크로스 컴파일**도 확인했다(빌드만으로는 `_test.go`가 컴파일되지 않는다).

### 2.9 AC-010 / AC-011

```
$ grep -n 'exec\.Command\|"gh"\|git remote' internal/statusline/forge_spawn_gate_test.go internal/statusline/forge_pair_states_test.go internal/profile/subtree_resolve_test.go
forge_spawn_gate_test.go:77:  // ... the `git remote` cost ...          ← 주석
forge_spawn_gate_test.go:243: stubDir := writeForgeStub(t, "gh", 42, 17, false)   ← PATH 앞단 스텁, 실 바이너리 아님
```

실제 `gh` 호출도, 실 `git remote` 네트워크 호출도 없다. `TestForgeOptOut_TwoWayRevert`는 `t.Setenv("PATH", stubDir+…)`로 스텁을 앞세우며 windows에서는 `t.Skip`한다. AC-010 충족.
신규 코드는 환경변수를 새로 읽지 않고, 기존 참조는 `config.EnvClaudeConfigDir` / `config.EnvNoProfileFallback` 상수를 경유하며 경로는 전부 `filepath` API다. AC-011 충족.

### 2.10 증거 파일의 도달 가능성 (F1)

```
$ git ls-files .moai/reports/t293/
.moai/reports/t293/progress.md
.moai/reports/t293/verdict.md

$ git status --short
?? .moai/reports/t293/evidence/

$ git ls-tree --full-tree -r --name-only origin/develop -- .moai/reports/t293/
.moai/reports/t293/progress.md
.moai/reports/t293/verdict.md
```

`evidence/` 11개 파일은 **추적되지 않으며 origin/develop에도 없다**. §E.2가 인용한 원 경로 `.moai/state/verify/t293-dev/` 역시 gitignore 대상의 머신 로컬 경로다. 즉 병합된 트리를 읽는 제3자에게는 인용된 증거 경로가 해소되지 않는다.

---

## 3. Baseline-attribution (baseline 귀속)

- 위 모든 명령은 **이번 감사 실행에서**, **이 워크트리**(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t293`)에 대해, **HEAD `1fb1ab78a`** 상태로 수행했다. `git diff --stat`이 빈 출력임을 변이 시험 전후로 확인했으므로 측정 대상 트리는 커밋 상태와 동일하다.
- run-phase 산출 수치(`ac-matrix.txt`, `red-*.txt`, `coverage.txt`)는 **인용이 아니라 대조 대상**으로만 사용했다. 판정 근거가 되는 수치는 전부 이 감사가 직접 재실행해 관측한 것이다.
- 커버리지 84.3%는 이 실행의 `go test -coverprofile` + `go tool cover -func` 출력이며, run-phase 파일에서 옮겨온 값이 아니다(공교롭게 일치).
- primary 체크아웃 `statusline.yaml`의 33줄·forge 키 부재는 이 실행의 `grep`/`wc` 관측이다(파일 읽기만 수행, git 조작 없음).
- 감사 시작·종료 시점 모두 HEAD는 `1fb1ab78a`로 불변이었고, 이 감사 창 안에서 다른 세션의 커밋은 관측되지 않았다.

---

## 4. Gaps (미검증 — 명시적으로 관측하지 않은 것)

1. **전체 스위트 미실행.** `go test ./...`는 부하 규율에 따라 돌리지 않았다. 두 대상 패키지 + `internal/cli` vet만 확인했으므로, 변경면 밖 패키지의 회귀 여부는 CI 몫이다.
2. **실 바이너리 통합 렌더 미수행.** `moai statusline`을 실제로 실행해 상태줄을 눈으로 확인하지 않았다. F2의 "폴링이 실제로 발생/정지하는가"는 파일 읽기 경로 추적으로 판단했지, 자식 프로세스 계수를 실측한 것이 아니다.
3. **`__no_such_profile__` 발원지 미규명.** 저장소에 없음만 확인했고, 상위 프로세스 체인은 추적하지 않았다. SPEC §F의 accepted Gap과 동일한 위치에 머문다.
4. **`launch.yaml` 실물 원장 형태 미확인.** subtree walk가 실제 사용자 원장의 모든 항목 형태(대소문자 변형, SameFile 별칭, 죽은 워크트리 경로)에 대해 어떻게 반응하는지는 픽스처로만 검증했다.
5. **CodeRabbit 리뷰 상태 미판독.** 이 감사는 코드·테스트·증거만 봤다. 병합된 PR의 리뷰 게이트 상태는 리드 소관.
6. **`resolveBoardRoot` 비대칭의 사용자 영향 범위 미측정.** 워크트리 설정과 primary 설정이 갈리는 케이스가 실제로 얼마나 발생하는지는 세지 않았다.

---

## 5. Residual-risk (잔여 위험 — 관측했음에도 여전히 틀릴 수 있는 것)

- **레버 비대칭이 조용히 오작동할 수 있다(F3).** 워크트리의 `statusline.yaml`에 `forge: none`을 적어 넣은 운영자는 그것이 먹힐 것이라 기대하지만, 실제로 읽히는 것은 primary의 파일이다. 실패가 조용하다 — 오류도, 경고도 없이 폴링이 계속된다. 지금은 두 파일이 결국 같은 커밋 내용으로 수렴하므로 드러나지 않지만, 워크트리에서 설정을 손보는 순간 재현된다.
- **M7 키 휘발(운영자·lane-2 모두 기록한 위험).** `.moai/config`는 `moai update`의 `CleanMoaiManagedPaths`가 통째로 지우고 재배포하는 뿌리다. 파일 안 경고 주석은 사람에게만 보이고 삭제 코드는 그 주석을 읽지 않는다. 옵트아웃은 다음 update에서 사라질 수 있다.
- **원장 성장(REQ-009 이연분).** 읽기 경로만 고쳤으므로 워크트리마다 `projects[]` 행이 계속 쌓인다. 읽기 쪽이 이제 관대해졌으므로 쌓인 행이 틀린 프로필을 가리키는 상황도 원리상 가능하다 — 다만 exact-match가 subtree보다 우선하므로(`TestResolveLaunchProfileForProject_ExactMatchBeatsSubtree` 통과) 현 시점 위험은 낮다.
- **프로필 상속 확대의 부작용(F7).** 등록된 프로젝트 안에 무관한 별도 체크아웃이 중첩돼 있으면, 그 세션도 이제 상위 프로젝트의 프로필(즉 자격증명이 담긴 `CLAUDE_CONFIG_DIR`)로 해석된다. 의도된 동작의 이면이며, 벤더링된 하위 저장소가 있는 프로젝트에서 예상 밖으로 작동할 수 있다.
- **`0/0` 증상은 고쳐진 것이 아니라 우회됐다.** 이슈가 보고한 "실제 11 issues / 8 PRs인데 0/0 표시"는 낡은 캐시가 정직하게 유지된 결과이며, 이 SPEC은 그 계약을 **고정**했지 바꾸지 않았다. 이 저장소에서는 `forge: none`으로 pair 자체가 사라져 증상이 관측되지 않을 뿐, 다른 저장소에서는 동일하게 재현된다.
- **AC-005의 테스트는 RED를 갖지 않는다.** M4는 설계상 test-only characterization이므로 변이 시험 대상이 아니었다. 즉 이 테스트가 실제 계약 위반을 잡아낼지는 미래의 회귀로만 검증된다.

---

## 6. Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| **Functionality (40%)** | **88** / 100 | PASS | §2.2 변이 3/3 사살 — 세 게이트 전부 실효. §2.3 셀렉터 비-0매치(3+4 / 8). §2.4 프로필 수정이 실제 런처 경로(`launcher.go:150`)에 놓임 확인. 감점: §2.6 M7 운영 효과가 현재 미발효, §2.5 두 레버의 루트 불일치, 통합 렌더 미검증 |
| **Security (25%)** | **92** / 100 | PASS | 신규 exec·신규 쓰기 없음; subtree walk는 `os.Stat` 읽기 전용. REQ-008 경계는 lexical이 아니라 `filepath.Dir` 연쇄에 의한 **구조적** 판정이라 `/proj-other`가 `/proj`에 매치될 수 없음(테스트로 고정). 변경 방향이 자식 프로세스·API 폴링을 **줄이는** 쪽이라 429 회귀면을 좁힌다. 감점: §5 F7 프로필 상속 확대 |
| **Craft (20%)** | **78** / 100 | PASS | 테스트 품질 우수 — spawn 계수 seam 배치가 "게이트됨"과 "guard에 막힘"을 구분하고, 짝 양성 케이스(`enabled` 렌더 = 1회)와 음성 공간 테스트(`UnsetOverrideStillSpawns`)를 함께 둠. 감점: 패키지 커버리지 84.3% < 85 바(§2.7), 귀속 문구 오류(F4), 증거 디렉터리 미추적(F1) |
| **Consistency (15%)** | **88** / 100 | PASS | `filepath` API·`envkeys` 상수·MX 주석·Conventional Commit `(t293)` 각인·템플릿 미러 없음(중립성 의도적 보존, 파일 내 명시). 기존 `resolveGitHubCounts`의 `boardRoot` 기준과 spawn 게이트가 동일 기준을 씀. 감점: §2.5 비대칭이 SPEC 서술에 반영되지 않음 |

**가중 조화평균** = 1 / (0.40/88 + 0.25/92 + 0.20/78 + 0.15/88) = 1 / 0.0115315 = **86.7**

**must-pass 방화벽**: Functionality 88 ≥ 80 ✓ · Security 92 ≥ 80 ✓ → 통과.

---

## 7. Findings (구조화 결함 목록)

- **F1** [HIGH] [optional] `.moai/reports/t293/evidence/` (11개 파일) — 미추적이며 `origin/develop`에도 없음. §E.2가 인용한 원 경로 `.moai/state/verify/t293-dev/`는 gitignore 대상 머신 로컬 경로. 병합 트리를 읽는 제3자에게 인용 경로가 해소되지 않아 VCI §2 귀속 요건을 만족하지 못한다. **필요한 수정**: `evidence/` 11개 파일을 추적 커밋으로 올리거나(권장), 올릴 수 없다면 §E.2의 증거 경로 문장을 "머신 로컬, 재현 명령은 §E.2 표에 인라인" 으로 정정할 것.
- **F2** [MEDIUM] [optional] `.moai/reports/t293/verdict.md` Claim 문단 — "코드 병합 전이라도 이 레포 세션의 gh 폴링이 멈춘다"는 실측으로 반증됨. 워크트리 세션의 `boardRoot`는 primary 체크아웃이고, primary의 `statusline.yaml`에는 `forge` 키가 없다(§2.6). **필요한 수정**: 해당 문장을 "primary 체크아웃에 병합분이 도달한 뒤 발효된다"로 정정.
- **F3** [MEDIUM] [optional] `internal/statusline/builder.go:255-267` + `internal/cli/statusline.go:82-85` — 세그먼트 레버는 `projectRoot`(워크트리), forge 레버는 `boardRoot`(primary)를 읽어 두 옵트아웃이 서로 다른 파일에서 온다. **필요한 수정**: 후속 카드로 (a) 두 레버의 루트를 일치시키거나 (b) SPEC/코드 주석에 이 비대칭을 명시할 것. 코드 변경 없이 문서화만으로도 조용한 실패는 크게 줄어든다.
- **F4** [LOW] [optional] `.moai/specs/SPEC-STATUSLINE-PROFILE-RESPECT-001/progress.md` §E.2 Gaps — 커버리지 부족을 "Delete, GetCurrentName paths"에 귀속했으나 `GetCurrentName`은 100.0%(§2.7). **필요한 수정**: 실측 목록(`IsValidProfileName` 0.0%, `EnsureDir` 70.0%, `GetBaseDir` 71.4%, `WritePreferences` 72.7% 등)으로 교체.
- **F5** [LOW] [optional] `internal/profile` 패키지 커버리지 84.3% < 85% 바. 기존 미커버 함수 소관임은 §2.7로 **확인됨**(신규 함수는 88.9%/100%). **필요한 수정**: 별도 카드로 `IsValidProfileName` / `IsValidPermissionMode`(둘 다 0.0%)에 테스트를 붙이면 바를 넘긴다 — 이 SPEC의 몫은 아님.
- **F6** [LOW] [optional] `internal/statusline/forge_spawn_gate_test.go:151-166` — AC-001/AC-003은 "config fixture"를 전제하지만 테스트는 `Options.SegmentConfig` 맵을 직접 주입한다. yaml→맵 경로는 `internal/cli/statusline_test.go`가 기존에 덮고 있어 실질 공백은 없다. **필요한 수정**: 없음(문서상 표현만 정확화하면 충분).
- **F7** [LOW] [optional] `internal/profile/profile.go:387` `lookupSubtreeProjectKey` — 등록 프로젝트 안에 중첩된 무관한 체크아웃도 상위 프로필(자격증명 디렉터리)을 상속한다. 설계된 동작의 이면. **필요한 수정**: 없음. 벤더링 하위 저장소가 생길 때 재평가할 것.
- **F8** [INFO] AC-009 / M5 이연은 전달 범위를 훼손하지 않는다. 읽기 경로만으로 자체 완결적이고(exact-match가 subtree보다 우선함이 테스트로 고정), 원장 성장이라는 대가는 §5에 명시돼 있으며 후속 카드 t297이 큐에 있다. 이연 자체는 결함이 아니다.

**blocking 결함 0건.** F1~F7은 전부 optional 등급이다 — 어느 것도 SPEC이 실제로 진술한 요구를 깨뜨리지 않으며, 정확성 결함도 아니다. F1과 F2는 **보고 정확성**의 문제이므로 코드 수정이 아니라 문장 정정으로 닫힌다.

---

## 8. #1675 폐쇄 판정

**코드 층위: 폐쇄.** 리드가 지목한 판정 기준 — "설정값이 맞느냐가 아니라 gh 폴링 자식이 생성되지 않느냐" — 는 충족됐다. 두 게이트 모두 `maybeRefreshGitHubCounts` **안**(override)과 **호출부**(segment)에서 spawn 지점 이전에 반환하며, 그 사실을 변이 시험으로 확인했다(§2.2). 렌더만 억제하고 spawn은 남는 형태가 아니다.

**운영 층위: 아직 미발효.** 팩토리 레인(워크트리) 세션이 읽는 파일은 primary 체크아웃의 `statusline.yaml`이고, 그 파일에는 아직 `forge` 키가 없다(§2.6 실측). 병합분이 primary 워킹트리에 도달하는 시점에 발효된다. 이는 코드 결함이 아니라 배포 상태이며, verdict.md의 문장만 정정하면 된다(F2).

**이슈 표제의 인과는 오귀속이었다.** "폴백 프로필이 설정을 무시"가 아니라 "옵트아웃 키가 렌더러가 읽는 파일에 애초에 없었고, 있었더라도 spawn은 막지 못했을 것"이 실체다. SPEC §A.1이 이를 정확히 반증해 두었고, 이 감사가 독립적으로 재확인했다.

---

**보고서 경로**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t293/.moai/reports/t293/sync-audit.md`
**변이 시험 잔여물**: 없음 — 세 파일 모두 원복 확인(`git diff --stat` 빈 출력).
**커밋·푸시**: 수행하지 않음.
