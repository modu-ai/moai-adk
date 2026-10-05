# t606 — pre-commit 형식 검사가 스테이징 내용 대신 작업 파일을 읽음

- card: t606 (hooks 감사 2026-09-11 · H11 · P2)
- worktree: `.claude/worktrees/t606`, branch `WT-precommit-staged-read`
- base: `eabce74448e094dd1a4393044138a04b2016af99`
- 대상: `internal/cli/hook_install_precommit.go` (`preCommitHookContent` 상수) + `internal/template/templates/.git_hooks/pre-commit`
- 상태: run 완료 — 재현 → 수리 → GREEN → 뮤턴트 2종 → 회귀. 미커밋·미푸시. develop 흡수와 병합은 리드 창 지명 대기.

> **경로 주의.** 이 카드의 증거는 `.moai/reports/t606/` 가 아니라 이 디렉터리에 있다. 그 경로는 같은 id 를 쓰던 **다른 착지 카드**("커버리지 검사 실패 메시지의 지목 파일이 실행마다 바뀌는 결함", `WT-fail-name-order`, 커밋 `e526c8d5b`)가 이미 점유하고 있다. 작업 초기에 그 자리에 초안을 써서 착지본을 덮었고, `git restore` 로 원복한 뒤 리드 승인을 받아 이 경로로 옮겼다. 원복 검증: `git status --short -- .moai/reports/t606/` 무출력, 같은 디렉터리의 다른 5개 파일 무접촉, 손실 0.

## 1. 원인 — 코드 판독

pre-commit 훅의 gofmt 단계가 `gofmt -l "$f"` 로 **작업 트리의 파일**을 읽는다. 커밋에 실리는 것은 인덱스의 내용이므로, 두 사본이 갈리면 판정 대상과 커밋 대상이 서로 다른 파일이 된다. 결함은 대칭이다.

| 인덱스 | 작업 파일 | 훅 판정 | 옳은 판정 |
|---|---|---|---|
| 미포맷 | 포맷됨 | 통과 | 차단 |
| 포맷됨 | 미포맷 | 차단 | 통과 |

한 방향만 고치는 수리는 다른 방향을 깨뜨리므로 두 방향을 모두 잰다.

두 번째 결함은 수리 과정에서 드러났다. gofmt 의 **stdout 이 비었다는 것만으로 통과를 결정**하는 구조라, gofmt 가 아예 실행되지 못한 경우(입력을 읽지 못함, 스테이징 바이트가 파싱되지 않음)에도 게이트가 조용히 초록을 낸다. 통과시키는 쪽으로 실패하는 게이트는 없느니만 못하다 — 없으면 아무도 믿지 않지만, 있으면서 통과하면 근거로 읽힌다. 리드 지시로 이 카드 범위에 포함했다.

## 2. 재현

격리 저장소(`git init`, `t.TempDir`)에 develop 의 훅 원본을 `.git/hooks/pre-commit` 으로 설치하고 `git commit` 으로 구동했다. `go.mod` 를 두지 않아 vet 단계는 모듈 탐색 실패로 skip 되고, `.moai/config` 가 없어 heavy gate 도 skip 된다 — gofmt 단계만 남는다.

CASE A (결함 방향) — 인덱스 미포맷, 작업 파일 `gofmt -w` 적용:

```
staged_needs_formatting=true
worktree_needs_formatting=false
hook_exit=0
```

감사 H11 의 관측값(`staged_needs_formatting=true; hook_exit=0`)과 일치한다.

CASE B (반대 방향) — 인덱스 포맷됨, 작업 파일만 미포맷:

```
staged_needs_formatting=false
worktree_needs_formatting=true
[pre-commit] FAILED: the following staged files need formatting:
main.go
hook_exit=1
```

`gofmt -l /dev/stdin` 의 세 갈래도 직접 쟀다. 미포맷 → stdout `/dev/stdin`, exit 0. 포맷됨 → stdout 없음, exit 0. 문법 오류 → stdout 없음, stderr 진단, **exit 2**. 세 번째가 §1 두 번째 결함의 근거다.

## 3. 수리

두 사본에 동일한 편집을 적용했다(`TestPreCommitTemplateMatchesConstant` 가 바이트 동일성을 강제한다).

**(a) 인덱스 바이트 판정.** `gofmt -l "$f"` → `git show ":$f" | gofmt -l /dev/stdin`. 출력하는 이름은 인덱스 경로 `$f` 이므로 `/dev/stdin` 이 사용자 메시지로 새지 않는다. 실패 메시지에 판정 대상이 스테이징 바이트임을 명시했다.

**(b) 종료 코드 분기.** 한 번의 순회로 파일마다 두 결과(`fmt` / `err`)를 접두사로 구분해 모으고, `err` 가 하나라도 있으면 별도 메시지로 차단한다. 루프가 서브셸이라 변수를 밖으로 내보낼 수 없어 접두사 방식을 썼다.

```sh
if _out="$(git show ":$f" 2>/dev/null | gofmt -l /dev/stdin 2>/dev/null)"; then
    if [ -n "$_out" ]; then printf 'fmt %s\n' "$f"; fi
else
    printf 'err %s\n' "$f"
fi
```

**(c) vet 메시지.** 실행 범위와 리다이렉션은 그대로 두고 문구에만 판정 기준을 밝혔다("vetted in the working tree, not the index"). vet 은 디스크의 파일을 컴파일하므로 인덱스 스냅샷을 별도 체크아웃하지 않는 한 인덱스를 대상으로 삼을 수 없다. 형제 카드(pre-commit 이 go vet 출력을 삼킴)와 겹치지 않게 그 축은 건드리지 않았다.

### 의도적 동작 변경 — 문법 오류 스테이징 바이트

(b) 는 기존 동작을 **하나 바꾼다**. 종전에는 파싱되지 않는 스테이징 바이트가 gofmt 단계를 무플래그 통과하고 `go vet` 이 잡았다. 이제 gofmt 단계가 차단한다. Go 툴체인이 없거나 모듈 밖이라 vet 이 skip 되는 환경에서는 종전에 아무도 잡지 못하던 것을 잡게 되는 것이므로 개선으로 판단했으나, 무변경이 아니라 **의도된 변경**이다. 조사 초기 보고에서 "기존 동작 불변"이라고 적었던 부분은 이 결정으로 무효가 되었다.

**후속 수정 (리드 승인 후, 2026-09-12).** 첫 커밋(`62e138ba5`)의 두 사본에 이 결정과 **모순되는 주석**이 남아 있었다 — "파싱되지 않는 스테이징 바이트는 종전대로 표시되지 않는다(go vet 몫)". (a) 를 위해 쓴 문장이 (b) 를 뒤에 얹으면서 거짓이 된 것으로, 코드는 옳고 주석만 틀렸다. 주석이 지금 동작을 서술하도록 두 사본을 같은 편집으로 고쳤고(`TestPreCommitTemplateMatchesConstant` PASS), 리드 지시대로 `CHANGELOG.md` `[Unreleased] → Fixed` 에 동작 변경을 한 줄로 남겼다 — 조용히 통과하던 저장소가 갑자기 막히는 것은 결함이 아니라 게이트가 제 일을 하는 것이라는 점을 함께 적었다.

재검증(주석·CHANGELOG 편집 후, 부하 16.34): `go test ./internal/cli/ -run 'PreCommit|Precommit|StagedRead' -count=1` → `ok ... 22.783s`, `-v` 로 다시 재어 `--- PASS` 52 건 / `--- FAIL` 0 건. 선택자가 이름을 조용히 버리지 않았음을 PASS 수로 대조했고, 바이트 동일성 시험은 단독으로도 확인했다(`--- PASS: TestPreCommitTemplateMatchesConstant`). `internal/template/catalog.yaml` 에 `.git_hooks` 항목이 없어(grep 무출력) 카탈로그 해시 재생성은 불필요하다.

## 4. 검증

각 측정 직전 `uptime` 으로 1분 부하를 확인했다(리드 부하 게이트, < 30): 8.71 → 6.90 → 6.32 → 8.66 → 17.19.

### 4.1 빌드

```
make build            → 완료 (catalog.yaml 12899 bytes, bin/moai 재생성)
go build -o /dev/null ./cmd/moai   compile_exit=0
ls -la bin/moai       70592338 bytes, Sep 12 17:40
```

`make build` 는 파이프라인 뒤 종료 코드를 잡지 못해(zsh `PIPESTATUS` 미지원) 별도 `go build` 로 종료 코드를 확인했다.

### 4.2 GREEN — 새 시험 5본

```
go test ./internal/cli/ -run 'TestPrecommitStaged' -count=1 -v      test_exit=0
--- PASS: TestPrecommitStagedUnformattedWorktreeFormattedBlocks (0.62s)
--- PASS: TestPrecommitStagedFormattedWorktreeUnformattedAllows (0.49s)
--- PASS: TestPrecommitStagedFormattedNoDivergenceAllows (0.52s)
--- PASS: TestPrecommitStagedGofmtToolFailureBlocks (0.59s)
--- PASS: TestPrecommitStagedUnformattedNoDivergenceBlocks (0.65s)
ok  	github.com/modu-ai/moai-adk/internal/cli	3.820s
```

파일: `run2-green.txt`. `=== RUN` 5줄을 세어 선택자가 실제로 5본을 골랐음을 확인했다(선택자는 없는 이름을 조용히 버린다).

시험 구성 — 불일치 2본은 서로 배타적이라 한쪽만 맞추는 훅이 존재할 수 없고, 대조군 2본(불일치 없음·포맷됨 → 통과 / 불일치 없음·미포맷 → 차단)은 GREEN 이 "무엇이든 통과시키는 훅" 때문이 아님을 배제한다. 5본째가 도구 실패 방향이다.

### 4.3 뮤턴트 1 — 작업 트리 읽기로 되돌림

`git show ":$f" | gofmt -l /dev/stdin` → `gofmt -l "$f"`.

```
mutant_exit=1
--- FAIL: TestPrecommitStagedUnformattedWorktreeFormattedBlocks (0.62s)
--- FAIL: TestPrecommitStagedFormattedWorktreeUnformattedAllows (0.41s)
--- PASS: TestPrecommitStagedFormattedNoDivergenceAllows (0.44s)
--- PASS: TestPrecommitStagedGofmtToolFailureBlocks (0.55s)
--- PASS: TestPrecommitStagedUnformattedNoDivergenceBlocks (0.42s)
```

파일: `run1-mutant-worktreeread.txt`. 불일치 두 방향만 정확히 실패하고 대조군 셋은 통과한다.

### 4.4 뮤턴트 2 — 종료 코드 분기 무력화

`if [ -n "$FMT_ERR" ]` → `if [ -n "" ]`.

```
mutant_test_exit=1
--- FAIL: TestPrecommitStagedGofmtToolFailureBlocks (0.63s)
precommit_staged_read_e2e_test.go:160: commit succeeded although gofmt could not run (a silent pass):
    [main (root-commit) 67e7756] t606 fixture
     1 file changed, 6 insertions(+)
```

파일: `run3-mutant-noexitbranch.txt`. 나머지 4본은 통과했다. 커밋이 실제로 만들어졌다는 것이 "조용한 통과"의 관측이다 — 코드 판독이 아니라 실측이다.

두 뮤턴트 모두 Edit 로 넣고 같은 Edit 를 역으로 적용해 되돌렸으며, 원복은 `grep -n 'FMT_ERR" \]; then'` → 95행 1건, `grep -c 'if \[ -n "" \]'` → 0 으로 확인했다.

### 4.5 회귀 — pre-commit 계열 전량

```
go test ./internal/cli/ -run '(?i)precommit|(?i)pre_commit' -count=1 -v    test_exit=0
=== RUN  62 / --- PASS 62 / --- FAIL 0 / --- SKIP 0
ok  	github.com/modu-ai/moai-adk/internal/cli	20.485s
--- PASS: TestPreCommitTemplateMatchesConstant (0.00s)
```

파일: `run4-regression.txt`. 선택자가 고른 최상위 시험은 51본이고(`-list` 로 확인), 서브테스트를 포함해 62 RUN 이다. 두 사본의 바이트 동일성은 `TestPreCommitTemplateMatchesConstant` 가 낸 판정이다.

### 4.6 최종 확인 (뮤턴트 원복 후)

```
go test ./internal/cli/ -run 'TestPrecommitStaged|TestPreCommitTemplateMatchesConstant' -count=1 -v   final_exit=0
--- PASS 6 / --- FAIL 0
```

파일: `run5-final.txt`.

### 4.7 gofmt / vet

```
gofmt -l internal/cli/precommit_staged_read_e2e_test.go   출력 없음, exit 0
go vet ./internal/cli/                                     vet_exit=0
```

### 4.8 병합 트리 재측정 (창 1번차, 2026-09-12)

로컬 develop `e16b0e9a3`(gateway t649 포함) 을 흡수해 병합 커밋 `070d8f2dc` 를 만들고, 그 트리에서 다시 쟀다. 부하 7.83.

**CHANGELOG 충돌 1건.** 흡수 중 `### Fixed` 의 같은 자리에 t603 항목(develop)과 t606 항목(이 카드)이 각각 삽입돼 충돌했다. 내용이 겹치지 않는 순수 삽입이므로 **양쪽을 모두 보존**하고 마커만 제거했다(t603 먼저, t606 다음). 해소 후 마커 0건, 두 항목 모두 존재함을 grep 으로 확인했다.

```
make build                                            → 완료 (catalog.yaml 12899 bytes)
go test ./internal/cli/... -count=1 -timeout 1800s    → exit 1
  internal/cli                     FAIL  1228.721s
  하위 16개 패키지                  전부 ok
```

**FAIL 3건은 이 카드와 무관하다 — gateway 계열이다.**

```
--- FAIL: TestCodexCommand_RegisteredInLaunchGroup
    codex_launcher_test.go:236: launcher "cg" missing from the launchers section block
--- FAIL: TestCharacterize_GLM_WarningPrintedToStderr
--- FAIL: TestNoBareGLMEnvVarLiteralsInCLIProduction
    glm_env_parity_test.go:115: internal/cli/gateway_prepare.go:26:254 / :383 / :427
```

t606 계열 FAIL 은 **0건**이다(`FAIL: TestPreCommit` / `TestPrecommit` / `TestStagedRead` grep 무적중).

**귀속 — 잰 것.** 병합 트리가 develop 대비 바꾼 파일을 전수했더니(`diff --name-only e16b0e9a3 070d8f2dc`) 이 카드의 10개뿐이었다: 훅 두 사본 · e2e 시험 · CHANGELOG · verdict · run 로그 5. 실패가 지목한 `gateway_prepare.go` · `codex_launcher_test.go` · `glm_env_parity_test.go` 는 그 목록에 **없다**. 그 파일의 마지막 커밋은 `5575ba649 Merge App Server turn bridge and preserve gateway fixes (t649, t652)` 다.

**"develop 단독도 red" — 리드가 정적으로 닫은 근거 (2026-09-12, 인용).** 이 카드가 직접 재지 못한 부분을 리드가 코드 판독으로 대신 세웠다:

1. `launcher.go:135` 의 `cg` 는 `mode == cg` 비교이지 섹션 등록이 아니다 — 테스트가 요구하는 등록이 실제로 없다.
2. `gateway_prepare.go:26` 에 베어 리터럴이 관측된다.
3. 패리티 테스트는 `1444583bf`(t457) 이전부터 존재했고, 위반 코드는 `5575ba649`(t649/t652)에서 유입됐다 — **기존 테스트를 새 코드가 위반한 구조**다.

수리는 이 카드 범위 밖이며 별도 카드 **t669**(lane-5 배차)가 가져갔다. 판정 A(병합 진행)는 리드의 것이다.

## 5. Gaps — 관측하지 않은 것

- ~~**`internal/cli` 패키지 전량**은 돌리지 않았다.~~ → §4.8 에서 병합 트리로 쟀다(1228.721s).
- **develop `e16b0e9a3` 단독 트리**에서 위 red 3건을 직접 재지는 못했다. 워크트리 가드가 교차 트리 접근(`-C <develop트리>`)과 복합 명령을 모두 거부해 측정 경로가 없었다. §4.8 의 귀속은 diff 전수로 세운 것이고, "develop 단독도 red" 는 리드의 정적 판독이다 — 측정이 아니다.
- ~~**develop 흡수 후 재측정**은 아직이다.~~ → §4.8 에서 `e16b0e9a3` 을 흡수해 병합 트리(`070d8f2dc`)에서 다시 쟀다. §1~§4.7 의 측정은 base `eabce7444` 트리의 것이며, 그 값을 병합 후 근거로 재사용하지 않는다.
- **darwin 에서만** 실행했다. linux / windows 판정은 CI 몫이다.
- **heavy gate(`moai gate`) 와 go vet 단계**는 재현·시험 모두에서 성공 대역에 두고 분리하지 않았다. 전체 게이트가 어떤 커밋을 허용한다는 주장은 하지 않는다.
- **이 저장소에서 손으로 커밋해 본 종단 확인**은 하지 않았다. 시험이 진짜 git 저장소와 진짜 훅 본문으로 `git commit` 을 구동하므로 같은 경로를 지나지만, 이 저장소 자체에서의 확인은 아니다.
- **golangci-lint** 는 돌리지 않았다. `go vet` 만 확인했다.
- 4.2 의 CASE A / CASE B 최초 재현은 세션 안에서만 관측했고 파일로 내보내지 않았다. 같은 판정을 파일로 가진 것은 4.3 뮤턴트(`run1-mutant-worktreeread.txt`)이며, 위 §2 의 수치는 그 최초 실행의 인용이다.

## 6. Residual-risk — 관측했음에도 남는 것

- **`/dev/stdin` 이식성.** 이 경로가 없는 환경에서 gofmt 는 실패하는데, 이제 종료 코드 분기가 그것을 차단으로 바꾼다(4.4 로 실측). 즉 통과시키는 쪽이 아니라 막는 쪽으로 실패한다 — 안전한 방향이지만, 그런 환경에서는 모든 Go 커밋이 막힌다. 실측은 macOS 뿐이고, 종료 코드 분기가 플랫폼 무관이라는 것은 판독이다.
- **경로 형태.** `git show ":$f"` 는 저장소 루트 기준 경로를 요구한다. git 이 훅을 최상위에서 실행하므로 종전 `gofmt -l "$f"` 와 같은 전제이며, 이 전제가 깨지는 환경이라면 두 형태 모두 이미 깨져 있다. `core.hooksPath` 를 비표준으로 둔 환경은 확인하지 않았다.
- **개행이 든 파일명.** `fmt` / `err` 접두사 분리는 줄 단위이므로 파일명에 개행이 있으면 오분류된다. 종전 구현도 줄 단위 목록이라 같은 한계를 이미 갖고 있었고, 새로 좁아지지도 넓어지지도 않았다.
- **서브모듈·심볼릭 링크.** 스테이징 항목이 일반 파일이 아닐 때 `git show ":$f"` 가 무엇을 내는지 확인하지 않았다. 대상이 `.go` 확장자로 걸러진 목록이라 실무상 노출은 좁다.
- **성능.** 스테이징된 `.go` 파일 하나당 `git show` 프로세스가 하나 늘고, 스캔 뒤 `sed` 두 번이 붙는다. 훅의 sub-second 특성을 깨뜨릴 규모인지는 재지 않았다.
