auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: e60f0e02b0884699179c388cc8c2483b1d1d3801
blocking_count: 1
blocking_defect: G1

# t2001 sync 감사 5회차 (main 흡수 + base_count 수리 확인) — WT-cc-bash-monitor

- 카드: t2001 (Class B, SPEC 없음) · 리더 결정: d-20261106T185643Z-4e7a (먼저 재현, 재현되면 수리) · Codex gate advisory 전환: d-20261106T230753Z-d852 (이번 감사는 Claude 단독, 유료 호출 없음)
- 고정 대상: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001`, `git branch --show-current` → `WT-cc-bash-monitor`, `git rev-parse HEAD` → `e60f0e02b0884699179c388cc8c2483b1d1d3801`. 감사 시작과 이 파일을 쓰기 직전에 두 번 확인했고 같았다. 작업 트리는 두 번 모두 깨끗했다(`git status --short` 빈 출력). 재현은 `go test -overlay`로 scratchpad 사본만 바꿔 끼워 했고, 트리는 한 번도 고치지 않았다.
- 종합: **FAIL, 점수 76/100**. 이번 수리의 판정 로직 자체는 맞다. RED를 직접 재현했고(수리 전 `crossed-by-change`/`block`), 수리 후에는 `inherited`/`warn`이다. `BaseCount`가 비면 막고, push 이벤트와 다른 층은 그대로 막으며, `CodemapsInputsStable`이 true일 때 base count는 누적 count와 같은 anchor를 쓴다. 병합 해소 세 건도 의미상 맞다. 그런데 **같은 뺄셈이 CLI stderr 문장에 남아 있다(G1, blocking)**. 이번 수리가 겨냥한 겹침 사례에서 `moai graph check`는 "inherited — the base alone carries 4 >= threshold 5"라고 찍는다. 판정은 inherited인데 근거로 내놓는 숫자는 거짓이고, 부등식 자체도 성립하지 않는다. 문서는 이 stderr 한 줄이 상속인지 원인인지 알려 준다고 약속하므로 기능 결함으로 본다.

## 이전 항목 상태

| 항목 | 상태 |
|---|---|
| D2 (stamp 재독 TOCTOU) | OPEN (Low, optional) — 이번 범위 밖 |
| D3 (장치 하나만 지운 변이체 통과) | OPEN (Low, optional) — 범위 밖 |
| D4 (문서의 "바이트 단위" 과장) | OPEN (Low, optional) — 범위 밖 |
| D5 (CHANGELOG 이중 "or") | OPEN (Info, optional) — 해당 문장 그대로 남아 있음 |
| D6 (Windows lockfile 테스트를 CI가 안 돌림) | OPEN (Info, optional) — 범위 밖 |
| N4 (임계값을 PR이 바꾸는 경우) | OPEN (Info, optional) — 범위 밖 |
| F7 (inbox 이중 rotation 경합) | OPEN (Low, optional) — 범위 밖 |
| F8 (base=`HEAD^1` 전제) | OPEN (Info, optional) — 이번 수리도 `HEAD^1`을 base로 쓴다(`attachBaseCount`가 `rep.ContributionBase` 재사용) |
| Codex 지연 P2 (`driftOrigin`의 Value−Contribution) | **판정 경로 CLOSED** — 재현 후 수리 확인. 표시 경로는 G1로 남음 |

## 결함 목록 (structured defect-list)

- **G1** [Medium] [blocking] [in-diff, 재현 증거 있음] `internal/cli/graph_check.go:306` — `driftOriginSentence`가 아직 `base := l.Value - *l.Contribution`으로 base를 계산한다. e60f0e02b0은 `driftOrigin`만 `BaseCount`로 바꿨고 이 함수는 건드리지 않았다. 그래서 겹침 사례(Value=6, Contribution=2, BaseCount=6, threshold=5)에서 stderr가 "inherited — the base alone carries 4 >= threshold 5, this change added 2"를 찍는다. `crossed-by-change` 문장도 같은 뺄셈을 써서 겹치면 base를 적게 적는다. 함수 주석(:300 "Base alone is value - contribution")도 옛 설명 그대로다. `internal/cli/graph_check_policy_test.go`에는 문장 속 숫자를 확인하는 단언이 없어서(`grep "base alone carries\|base carried"` 0건) 테스트가 잡지 못했다. 수리 전에는 판정과 문장이 같은 뺄셈을 써서 서로 맞았지만, 이번 수리 뒤로는 둘이 서로 어긋난다. 엄밀히는 이전 감사 SHA 이후 diff 밖이라 repair-born은 아니고, 수리가 덜 된 것이다.
  - Required fix: `driftOriginSentence`에서 `l.BaseCount`가 nil이 아니면 `*l.BaseCount`를 base로 쓴다. nil이면 inherited·crossed 문장에 base 숫자를 내지 않는다(그 경우 판정은 이미 anchor-moved라 이 분기에 오지 않지만, 숫자를 지어내지 않게 한다). 주석 :300도 고친다. 겹침 사례(Value=6, Contribution=2, BaseCount=6, threshold=5)에서 문장에 "carries 6"이 나오고 "carries 4"는 안 나온다는 CLI 테스트를 하나 더한다.
- **G2** [Low] [optional, 같은 수리 커밋에 넣기를 권장] [in-diff, 병합에서 생김] `CHANGELOG.md:33` — 카드 lockfile 항목이 "the worktree migration releases its lock explicitly"라고 적는다. 병합 bf6d3c4d83이 `withMigrationLock`의 명시적 `defer lockfile.Unlock(f)`를 뺐으므로 이제 사실이 아니다. b7c4e23b8a 시점에는 맞는 문장이었다(`git show b7c4e23b8a:CHANGELOG.md | grep -c "releases its lock explicitly"` → `1`, 당시 Unlock 존재). Required fix: 해당 절을 "the worktree migration takes main's platform-split lock (released when the handle closes)"처럼 실제 동작으로 고치거나 지운다.
- **G3** [Low] [optional] [in-diff, 이전부터 있었음] `CHANGELOG.md:33` — "a `LockFileEx` byte-range lock on one byte at offset 0"이라고 적었지만 코드는 `OffsetHigh: 0xFFFFFFFF`, 즉 offset `0xFFFFFFFF00000000`을 잠근다(`internal/lockfile/lockfile_windows.go:41-47`). b7c4e23b8a 시점에도 같은 문장이었다(`grep -c "one byte at offset 0"` → `1`). 4회차가 놓친 항목이다. Required fix: offset 설명을 코드에 맞춘다.
- **G4** [Info] [optional] [outside] `internal/cli/worktree/migrate_lock_windows.go:12-14` (main 소유) — 주석이 "internal/lockfile's Windows arm is a process-local mutex by contract"라고 한다. 이 카드가 lockfile Windows 쪽을 `LockFileEx`로 바꿨으므로 병합 뒤에는 낡은 설명이다. 동작에는 영향 없다(이 함수는 lockfile을 거치지 않고 직접 `LockFileEx`를 부른다). 판정에 넣지 않고 운영자 표시용으로만 둔다.
- **G5** [Info] [optional] [in-diff] `internal/graph/check.go:452-479` (dirty-generation 경로) — 이 경로는 `attachBaseCount`를 부르지 않으므로 `BaseCount`가 늘 nil이고, stale이면 origin이 이제 `anchor-moved`다. 수리 전에는 Value=1, 기본 threshold=5라 `crossed-by-change`였다. 두 경우 모두 block이라 fail-closed는 지켜지지만, 이 경로에서 `anchor-moved` 이름("this change touched the codemaps body or stamp")은 정확한 설명이 아니다. 수리는 요구하지 않는다.

## 차원별 점수

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 68/100 | FAIL | RED 재현: 수리 전 `origin="crossed-by-change" class="block"`, 수리 후 `PASS` (`Value=6 Contribution=2 threshold=5 stable=true`). `go test ./internal/graph/` → `ok ... 79.976s coverage: 88.2% of statements`. 그러나 G1 probe: `sentence: inherited — the base alone carries 4 >= threshold 5, this change added 2` |
| Security (25%) | 88/100 | PASS | 새 입력 경계 없음. `attachBaseCount`는 기존 `gitOutput`과 이미 검증된 `roots`(`validateDescribedRoot`)·anchor·`HEAD^1` SHA만 인자로 쓴다. 측정 실패는 nil로 남고 nil은 `anchor-moved`(block) |
| Craft (20%) | 82/100 | PASS | `golangci-lint run ./internal/graph/` → `0 issues.`, `gofmt -l` 빈 출력, `GOOS=windows go vet` 세 패키지 exit 0, graph 커버리지 88.2%. 감점: G1을 막을 문장 단언 테스트 없음 |
| Consistency (15%) | 70/100 | FAIL | 문서 네 언어와 CHANGELOG의 `base_count` 설명은 코드와 맞다. 그러나 CHANGELOG에 병합 뒤 거짓이 된 절(G2)과 원래부터 틀린 offset(G3)이 있다. 커밋 메시지는 카드 id·Conventional Commit 형식을 지킨다 |

가중 합 0.4×68 + 0.25×88 + 0.2×82 + 0.15×70 = 76.1 → **76/100**. Must-pass인 Functionality가 blocking G1 때문에 기준 아래이므로 → **FAIL**.

## Claim / Evidence

**Claim A — 이전 감사 이후 범위**

```
$ git log --oneline --first-parent b7c4e23b8a..HEAD
e60f0e02b0 [t2001] fix(graph): measure the base's own codemaps count directly instead of subtracting the contribution (card t2001)
bf6d3c4d83 [t2001] merge: absorb main b25adf4fca into the card branch before the landing re-audit (card t2001)
63e525cbf3 [t2001] docs(audit): record the round-4 fix-confirm sync audit verdict PASS-WITH-DEBT at b7c4e23b8a (card t2001)
$ git diff --stat b7c4e23b8a 63e525cbf3 | tail -5
 .moai/reports/t2001/sync-audit-4.md | 229 ++++++++++++++++++++++++++++++++++++
 1 file changed, 229 insertions(+)
$ git rev-parse bf6d3c4d83^1 bf6d3c4d83^2
63e525cbf38047088c68fcf735ce702133443cc8
b25adf4fca05e6dba07abbdf307ccaf2b80b9033
```

**Claim B — RED를 직접 관측했다 (a)**

scratchpad에 현재 `check_policy.go` 사본을 만들고 `driftOrigin`의 `BaseCount` 두 분기를 수리 전 한 줄(`case l.Value-*l.Contribution >= threshold:`, `git show e60f0e02b0^:internal/graph/check_policy.go`와 같은 줄)로 되돌린 뒤 `-overlay`로 끼웠다. 트리 파일은 그대로다.

```
$ GOCACHE=<scratch>/gocache go test -overlay <scratch>/overlay.json -count=1 -run TestClassifyVerdicts_BaseAndChangeOverlapIsInherited -v ./internal/graph/
exit=1
=== RUN   TestClassifyVerdicts_BaseAndChangeOverlapIsInherited
    check_overlap_test.go:41: Value=6 Contribution=2 threshold=5 stable=true
    check_overlap_test.go:50: origin="crossed-by-change" class="block", want "inherited"/"warn" (Value=6 Contribution=2 threshold=5)
--- FAIL: TestClassifyVerdicts_BaseAndChangeOverlapIsInherited (1.19s)
FAIL
$ GOCACHE=<scratch>/gocache go test -count=1 -run TestClassifyVerdicts_BaseAndChangeOverlapIsInherited -v ./internal/graph/
exit=0
    check_overlap_test.go:41: Value=6 Contribution=2 threshold=5 stable=true
--- PASS: TestClassifyVerdicts_BaseAndChangeOverlapIsInherited (1.10s)
ok  	github.com/modu-ai/moai-adk/internal/graph	1.832s
```

픽스처는 실제 git 저장소다(stamp 뒤 base 커밋이 g1..g6 여섯 파일 추가, PR 커밋이 g1·g2를 다시 고침). 따라서 `attachBaseCount`가 git에서 6을 쟀다는 것도 이 PASS로 확인된다.

**Claim C — fail-closed 경로 (b), 코드 읽기**

- `driftOrigin` 순서(`check_policy.go:117-134`): Contribution nil → `""`(warn 불가) → `Contribution >= threshold` → originated → `CodemapsInputsStable` nil/false → anchor-moved → `BaseCount == nil` → anchor-moved → `*BaseCount >= threshold` → inherited → 그 밖 crossed-by-change. inherited만 warn 후보다.
- `ClassifyVerdicts`(`:81-101`): `event != EventPullRequest || inheritedPolicy == InheritedDriftBlock || nonFresh != 1`이면 바로 반환하므로 push·빈 이벤트·다른 층이 하나라도 non-fresh면 모두 block. 이번 커밋은 이 함수를 바꾸지 않았다.
- `attachBaseCount`(`check.go:854-869`): Contribution nil, `ContributionBase` 빈 값, anchor 빈 값, git 오류면 BaseCount를 nil로 남긴다. nil은 위 순서대로 anchor-moved(block). 신규 테스트 `TestClassifyVerdicts_UnmeasuredBaseCountBlocks`가 이를 고정한다.
- **같은 anchor 전제**: `checkCodemaps`가 `attachBaseCount(projectRoot, roots, anchor, &rep)`에 누적 count와 같은 `anchor` 변수를 넘긴다(`check.go:557-568`). anchor는 `resolveContentAnchor(stamp S)`로 정해지며, 규칙 A는 작업 트리 body와 S의 차이, 규칙 B는 `git log -1 S -- body`(HEAD와 무관)다. `CodemapsInputsStable == true`는 작업 트리의 codemaps 디렉터리(본문+stamp, untracked·ignored·링크 없음)가 `HEAD^1`과 같음을 뜻하므로 `HEAD^1`에서 계산해도 S·규칙 A 판단·anchor가 모두 같다. stable이 아니면 BaseCount 값과 상관없이 anchor-moved다.
- **BaseCount가 실제 base보다 커지는 경우(false warn)**: BaseCount는 `git diff --name-only anchor HEAD^1 -- roots`에 같은 술어(`mx.IsDescribedWorthy`)를 씌운 값이다. `HEAD^1` 시점에 같은 검사를 돌렸을 때의 Value와 같다(커밋에는 untracked가 없다). 유일하게 "Value 안의 base 몫"보다 커지는 경우는 PR이 base가 바꾼 파일을 anchor 내용으로 되돌릴 때다. 그때도 base 단독은 이미 임계값 이상이었으므로 "base가 이미 넘겼다"는 inherited 정의에는 맞는다. 거짓 warn을 만드는 경로는 찾지 못했다.

**Claim D — G1 관측**

scratchpad 테스트 파일 하나를 `-overlay`로 `internal/cli`에 끼워 `driftOriginSentence`를 겹침 사례로 직접 불렀다(트리에 파일 없음).

```
$ GOCACHE=<scratch>/gocache go test -overlay <scratch>/overlay_cli.json -count=1 -run TestAudit5ProbeDriftOriginSentenceOverlap -v ./internal/cli/
exit=0
=== RUN   TestAudit5ProbeDriftOriginSentenceOverlap
    zz_audit5_probe_test.go:16: sentence: inherited — the base alone carries 4 >= threshold 5, this change added 2
--- PASS: TestAudit5ProbeDriftOriginSentenceOverlap (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.759s
$ grep -rn "Contribution\|BaseCount" internal/cli/*.go | grep -v "_test.go"
internal/cli/graph_check.go:222:	case l.Contribution != nil:
internal/cli/graph_check.go:223:		base := l.ContributionBase
internal/cli/graph_check.go:228:			*l.Contribution, base, driftOriginSentence(l))
internal/cli/graph_check.go:229:	case l.ContributionAbsentReason != "":
internal/cli/graph_check.go:232:		_, _ = fmt.Fprintf(errs, "  contribution: unmeasured — %s\n", l.ContributionAbsentReason)
internal/cli/graph_check.go:303:	if l.Contribution == nil {
internal/cli/graph_check.go:306:	base := l.Value - *l.Contribution
internal/cli/graph_check.go:309:		return fmt.Sprintf("inherited — the base alone carries %d >= threshold %d, this change added %d", base, l.Threshold, *l.Contribution)
internal/cli/graph_check.go:311:		return fmt.Sprintf("crossed by this change — the base carried %d < threshold %d, this change added %d and took the count over the line", base, l.Threshold, *l.Contribution)
internal/cli/graph_check.go:313:		return fmt.Sprintf("originated by this change — it alone is %d >= threshold %d", *l.Contribution, l.Threshold)
$ grep -rn "base alone carries\|base carried" internal/cli/graph_check_policy_test.go
(출력 없음)
```

**Claim E — 병합 해소 (c)**

```
$ git diff bf6d3c4d83^2 bf6d3c4d83 -- internal/cli/worktree/migrate.go
@@ -1725,6 +1725,14 @@ func rekeyAndRebindBack(root string, s journalStep) error {
 func rebindCardForPath(root, oldPath, newPath string) error {
+	// The plan's source path comes from `git worktree list --porcelain`, ...
+	oldPath = filepath.FromSlash(oldPath)
+	newPath = filepath.FromSlash(newPath)
$ git diff --stat bf6d3c4d83^2 bf6d3c4d83 -- internal/cli/worktree/migrate_command_test.go CHANGELOG.md
 CHANGELOG.md | 2 ++
 1 file changed, 2 insertions(+)
$ git diff --stat bf6d3c4d83^1 bf6d3c4d83 -- internal/cli/worktree/migrate_command_test.go
 internal/cli/worktree/migrate_command_test.go | 26 ++++++++++++--------------
$ grep -c "activeProcessCWDProbe = func" internal/cli/worktree/migrate_command_test.go
1
$ git diff --stat bf6d3c4d83^1 bf6d3c4d83 -- CHANGELOG.md
 CHANGELOG.md | 9 ++++++++-
$ grep -c "card t2001" CHANGELOG.md
2
$ grep -n "^<<<<<<<\|^=======$\|^>>>>>>>" CHANGELOG.md internal/cli/worktree/migrate.go internal/cli/worktree/migrate_command_test.go
(출력 없음)
```

- `migrate.go`: 병합 결과는 main 버전과 카드의 `FromSlash` 한 덩어리만 다르다. `withMigrationLock`은 main 그대로다: `defer f.Close()` → `tryMigrationLock(f)` → `fn()`.
- 명시적 Unlock을 빼도 되는 이유: `migrate_lock_other.go`(`!windows`)는 `lockfile.TryLock`(flock, fd 닫히면 해제), `migrate_lock_windows.go`는 `windows.LockFileEx(..., LOCKFILE_EXCLUSIVE_LOCK|LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)`로 핸들에 묶인 잠금이라 핸들을 닫으면 풀린다. 카드가 Unlock을 넣은 원래 이유는 lockfile Windows 쪽이 프로세스 내부 mutex라 Close로 풀리지 않던 시절의 동작이었고, main의 `tryMigrationLock`은 Windows에서 lockfile을 아예 거치지 않는다. 그래서 t2001 Windows 잠금 회귀가 다시 생기지 않는다. 같은 범위의 Windows 컴파일: `GOOS=windows go vet ./internal/graph/ ./internal/lockfile/ ./internal/cli/worktree/` → exit 0.
- `migrate_command_test.go`: 병합 결과가 main과 바이트 단위로 같다(^2 대비 diff 없음). probe 주입은 한 번만 있고, `m6Registered`는 git 출력(이미 slash)과 `filepath.ToSlash(path)`를 비교하므로 카드 쪽 양변 ToSlash와 의미가 같다.
- `CHANGELOG.md`: main 대비 카드 항목 두 줄만 더해졌고, 카드 대비 main 항목이 들어왔다. 양쪽 다 보존. 단 G2·G3.

```
$ GOCACHE=<scratch>/gocache go test -count=1 -run Migrate ./internal/cli/worktree/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	73.156s
$ GOCACHE=<scratch>/gocache go test -count=1 -run GraphCheck ./internal/cli/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli	32.351s
```

**Claim E2 — Craft 측정**

```
$ GOCACHE=<scratch>/gocache go test -count=1 -cover ./internal/graph/
exit=0
ok  	github.com/modu-ai/moai-adk/internal/graph	79.976s	coverage: 88.2% of statements
$ gofmt -l internal/graph/ internal/cli/graph_check.go internal/cli/worktree/migrate.go internal/cli/worktree/migrate_command_test.go
gofmt_exit=0   (목록 출력 없음)
$ GOCACHE=<scratch>/gocache golangci-lint run ./internal/graph/
lint_exit=0
0 issues.
$ GOCACHE=<scratch>/gocache GOOS=windows go vet ./internal/graph/ ./internal/lockfile/ ./internal/cli/worktree/
exit=0
```

**Claim F — 부채 항목 (d)**: D2~D6·N4·F7·F8은 이번 범위 밖이고 열린 채 둔다. 이번 diff가 그 항목들의 코드를 고치지 않았다(e60f0e02b0 stat은 graph 4파일·문서 5파일만). F8 전제(`HEAD^1`)는 BaseCount에도 이어진다.

**Claim G — 문서 (e)**

```
$ grep -c "base_count" docs-site/content/en/cli-reference/graph.md docs-site/content/ko/cli-reference/graph.md docs-site/content/ja/cli-reference/graph.md docs-site/content/zh/cli-reference/graph.md CHANGELOG.md
docs-site/content/en/cli-reference/graph.md:1
docs-site/content/ko/cli-reference/graph.md:1
docs-site/content/ja/cli-reference/graph.md:1
docs-site/content/zh/cli-reference/graph.md:1
CHANGELOG.md:1
$ git show --stat e60f0e02b0
 CHANGELOG.md                                |  2 +-
 docs-site/content/en/cli-reference/graph.md |  2 +-
 docs-site/content/ja/cli-reference/graph.md |  2 +-
 docs-site/content/ko/cli-reference/graph.md |  2 +-
 docs-site/content/zh/cli-reference/graph.md |  2 +-
 internal/graph/check.go                     | 45 ++++++++++++++++++++----
 internal/graph/check_overlap_test.go        | 53 +++++++++++++++++++++++++++++
 internal/graph/check_policy.go              | 39 ++++++++++++---------
 internal/graph/check_policy_test.go         | 17 ++++++++-
 9 files changed, 136 insertions(+), 28 deletions(-)
```

en·ja 문장은 직접 읽었다: "measured directly between the anchor and the first parent and reported as `base_count`, not derived by subtraction" / "アンカーから最初の親までを直接数えた数で `base_count` として報告され、引き算では求めません". ko·zh는 바뀐 줄 수와 `base_count` 존재만 확인했다.

**Claim H — 감사 계획·Binding run conditions**

```
$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001
exit=0
"config_status": "ok", "model": "codex", claude gate "off", codex gate "advisory", glm gate "off",
"cross_model_active": true, "cross_model_required": false, "enforced_required": []
$ grep -n -i "binding run conditions" .moai/reports/t2001/progress.md
grep_exit=1
```

필수 백엔드가 없으므로(`enforced_required: []`) 지시대로 Codex/GLM/audit_multi는 부르지 않았다. 묶인 실행 조건은 없다.

## Baseline-attribution

모든 명령은 이번 실행에서 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001`, HEAD `e60f0e02b0884699179c388cc8c2483b1d1d3801`에 대해 돌렸다. `go test`·`go vet`·`gofmt`·`golangci-lint`는 트리 소스를 직접 컴파일하므로 판정 빌드가 곧 트리 HEAD다(Go 1.27.3 darwin/arm64). 공유 빌드 캐시가 실행 중 깨져 있어서(`could not import internal/unsafeheader (open .../go-build/df/...-d: no such file or directory)`) 모든 Go 명령은 scratchpad의 전용 `GOCACHE`로 다시 돌렸다. RED·G1 관측은 `-overlay`로 scratchpad 파일을 끼운 빌드다. `moai verify audit-plan`은 설치된 바이너리가 판정했고 그 커밋은 이번에 확인하지 않았다(설정만 읽는 동사라 판정 대상 Go 코드를 실행하지 않는다).

## Gaps

- 거부된 도구 호출: worktree 격리 가드가 복합 셸 명령 4건(파이프·변수·`|` 정규식이 섞인 `git`/`go`/`python3` 호출)을 거부했다. 모두 단순 명령으로 나눠 다시 실행했고, 소스 읽기로 대체한 측정은 없다. python 패치 대신 `sed`로 scratchpad 사본을 만들었다.
- 구현 에이전트가 보고한 RED는 이번에 overlay로 다시 관측했다(Claim B). 단 수리 전 원본 파일 전체가 아니라 현재 파일에서 두 분기만 되돌린 사본이다. 되돌린 줄이 `git show e60f0e02b0^:internal/graph/check_policy.go`의 줄과 같다는 것은 diff로 확인했지만, 그 원본 파일 전체로 돌리지는 않았다(원본에는 `BaseCount` 필드가 없어 새 테스트가 컴파일되지 않는다).
- `go build ./...`, 호스트 전체 `go vet`, `internal/cli` 전체 스위트, `internal/lockfile` 테스트는 돌리지 않았다(부하 지시). Windows는 vet(컴파일)만 했고 실제 Windows 실행은 없다(D6 그대로).
- G1은 함수를 직접 부른 probe로 관측했다. 빌드한 `moai graph check`를 겹침 픽스처에 돌려 stderr를 받은 종단 관측은 하지 않았다. `writeLayerAttribution`이 이 함수를 그대로 찍는 것은 코드 읽기(`graph_check.go:227-228`)로만 이었다.
- ko·zh 문서는 바뀐 줄 수와 `base_count` 존재만 확인했고 문장을 읽지 않았다.
- `golangci-lint`는 `./internal/graph/`만 돌렸다. `internal/cli`·`internal/cli/worktree`는 이번 실행에서 lint하지 않았다.
- 원격 CI·PR head는 이번에 읽지 않았다(로컬 AF-Flow 착지 전 재감사).

## Residual-risk

- G1을 고쳐도, 겹침 사례를 실제 CLI 종단으로 돌리는 테스트가 없으면 stderr와 JSON이 다시 어긋날 수 있다.
- Windows에서 `LockFileEx` 잠금은 핸들 닫힘으로 풀리지만, Microsoft 문서는 닫힘 뒤 해제 시점이 시스템 자원에 달렸다고 적는다. 같은 프로세스에서 마이그레이션을 바로 두 번 돌리면 두 번째가 잠김을 볼 여지가 이론상 있다. main의 설계이고(G4와 같은 파일) 이 카드가 만든 위험은 아니다.
- BaseCount는 `HEAD^1`이 base라는 전제(F8)와 anchor가 `HEAD^1`에서도 같다는 추론(Claim C)에 기댄다. stamp S가 `HEAD^1`의 조상이 아닌 병합 모양(복사된 stamp)은 테스트로 덮지 않았다.
- 공유 빌드 캐시 손상은 다른 세션의 동시 작업 때문으로 보인다. 같은 시각 다른 레인의 Go 실행 결과도 오염됐을 수 있다.

## 반복 기록

- 1회차 (`sync-audit-1.md`): FAIL.
- 2회차 (`sync-audit-2.md`): FAIL @ 7a946cfca2.
- 3회차 (`sync-audit-3.md`): FAIL @ 2fad4a8de6 — D1 blocking.
- 4회차 (`sync-audit-4.md`): PASS-WITH-DEBT @ b7c4e23b8a — D1 CLOSED, D6 NEW.
- 5회차 (이 파일): FAIL @ e60f0e02b0 — main 흡수 해소 OK, Codex 지연 P2 판정 경로 CLOSED(RED 재현), NEW G1(blocking)·G2·G3·G4·G5. repair-born blocking 아님(G1은 b7c4e23b8a..HEAD diff 밖) → 수리 중단 규칙 카운트 0.
