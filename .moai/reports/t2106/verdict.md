auditor-model: claude-opus-5-5

# t2106 sync 감사 판정 (2회차, main 흡수 뒤 재수리)

verdict: PASS-WITH-DEBT
audited_sha: 56017148b4ba7c06abde8fdcc913cd647f0d572a

- 카드: t2106 (Class B, `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved` 불안정)
- 감사 대상: `56017148b4` (브랜치 WT-testapply-pendingintentb), 흡수 기준 main `b25adf4fca`
- 판정 범위: `git diff b25adf4fca...HEAD -- internal/cli/worktree`와 t2079 수리와 합쳐진 코드의 의미
- 점수: 95/100 (Functionality 97 / Security 100 / Craft 90 / Consistency 93, 가중 합 95.75를 내림)
- BLOCKING 결함: 없음. optional 결함 2건(카드 diff 안, 이번에 새로 1건)과 1회차에서 넘어온 optional 3건.
- 감사 경로: Claude 단독(운영자 결정 d-20261106T051044Z-7a4c). Codex·GLM 도구는 부르지 않았다. 영수증 없음.

## Claim

1. `revLines`는 이제 정확히 한 번만 뒤집는다. t2079의 끝에서부터 도는 루프만 남았고, t2106이 덧붙였던 교환 블록은 `56017148b4`에서 빠졌다. 두 수리는 서로를 상쇄하지 않고 합쳐진다.
2. main 대비 `appendMoveListLine`의 동작 변화는 한 가지뿐이다. 같은 경로를 가진 최신 레코드가 새 줄과 똑같고, 그 뒤에 다른 트리의 레코드가 끼어 있을 때 새 줄을 더 이상 추가하지 않는다. main이 건너뛰던 경우는 모두 그대로 건너뛴다(건너뛰는 집합이 main의 상위집합). 이 변화는 함수 주석이 이미 적어 둔 규칙("the latest entry naming either of its two paths")을 구현한 것이고, 테스트가 이를 지킨다(돌연변이 M2가 죽는다).
3. 이동 목록 관련 선별 테스트 11개가 모두 통과한다. 이 중 t2079 테스트 2개와 t2106 테스트 5개가 들어 있다. 예전에 불안정하던 테스트는 10회 반복에서 모두 통과했다.
4. 이중 반전 돌연변이(M1)는 테스트 5개를 빨갛게 만든다. 즉 병합 때 생긴 결함이 다시 들어오면 테스트가 잡는다.
5. 범위: main 대비 HEAD의 차이는 `movelist.go`, `movelist_dedupe_t2106_test.go`, 그리고 `.moai/reports/t2106/` 아래 보고서 3개뿐이다. 병합이 main 내용 말고 다른 파일을 건드리지 않았다.

## Evidence

모든 명령은 이번 실행에서 워크트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2106`, HEAD `56017148b4ba7c06abde8fdcc913cd647f0d572a`에서 돌렸다. 테스트는 모두 트리에서 바로 빌드한 것(`go test`)이다.

### E1. 트리와 HEAD

```
$ git rev-parse --show-toplevel HEAD
/Users/goos/.moai/worktrees/projects/moai-adk-go/t2106
56017148b4ba7c06abde8fdcc913cd647f0d572a
$ git status --short
(빈 출력)
$ git rev-parse main
b25adf4fca05e6dba07abbdf307ccaf2b80b9033
$ git merge-base --is-ancestor b25adf4fca HEAD && echo ancestor_ok
ancestor_ok
```

### E2. 판정 대상 diff (main 대비, movelist.go)

```
$ git diff b25adf4fca...HEAD -- internal/cli/worktree/movelist.go
@@ -491,8 +491,7 @@ func appendMoveListLine(mlPath, line string) error {
 	// The dedupe: compare against the latest entry naming either path.
-	rev := revLines(string(data))
-	if len(rev) > 0 && rev[0] == line {
+	if latest, ok := latestLineNamingPathsOf(string(data), line); ok && latest == line {
 		return nil // identical re-run: nothing to append
 	}
@@ -522,3 +521,40 @@ func revLines(data string) []string {
+// latestLineNamingPathsOf ... (신규 함수)
+// namesAny ... (신규 함수)
```

`revLines` 본문은 main과 diff가 없다. 현재 본문(movelist.go:513-523):

```go
// revLines returns the non-empty lines of the file, last first.
func revLines(data string) []string {
	lines := strings.Split(data, "\n")
	var out []string
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			out = append(out, lines[i])
		}
	}
	return out
}
```

`56017148b4`가 지운 것(병합 커밋 대비):

```
$ git diff 8bac970c9f HEAD --stat
 internal/cli/worktree/movelist.go | 7 +------
$ git show 56017148b4 -- internal/cli/worktree/movelist.go
-	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
-		out[i], out[j] = out[j], out[i]
-	}
```

### E3. 호출자 열거

```
$ grep -rn -e 'revLines' -e 'appendMoveListLine' -e 'latestLineNamingPathsOf' -e 'namesAny' internal/ --include='*.go' | grep -v '_test.go'
internal/cli/worktree/migrate.go:746:	return appendMoveListLine(moveListPath(root), m6MoveLineReason("reversed", ...))
internal/cli/worktree/migrate.go:1253:	if ierr := appendMoveListLine(mlPath, m6MoveLineTrigger(trigger, "pending", ...))
internal/cli/worktree/migrate.go:1280:			mlErr := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "reversed", ...
internal/cli/worktree/migrate.go:1288:			mlErr := appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", ...))
internal/cli/worktree/migrate.go:1314:	if merr := appendMoveListLine(mlPath, m6MoveLineTrigger(trigger, "moved", ...))
internal/cli/worktree/migrate.go:1554:			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", ...))
internal/cli/worktree/migrate.go:1558:			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "refused", ...))
internal/cli/worktree/migrate.go:1563:			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", ...))
internal/cli/worktree/migrate.go:1573:			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", ...
internal/cli/worktree/migrate.go:1581:			appendMoveListLine(mlPath, m6MoveLineTriggerReason(trigger, "pending", ...
internal/cli/worktree/migrate.go:1584:	return true, "", appendMoveListLine(mlPath, m6MoveLineTrigger(trigger, "reversed", ...))
internal/cli/worktree/migrate.go:2042:			if lerr := appendMoveListLine(mlPath, m6MoveLineReason("refused", ...))
internal/cli/worktree/migrate.go:2051:			if lerr := appendMoveListLine(mlPath, m6MoveLineReason("pending", ...))
internal/cli/worktree/migrate.go:2135:				if lerr := appendMoveListLine(mlPath, m6MoveLine("restored", ...))
internal/cli/worktree/migrate.go:2203:			if lerr := appendMoveListLine(mlPath, m6MoveLine("restored", ...))
internal/cli/worktree/movelist.go:486:func appendMoveListLine(mlPath, line string) error {
internal/cli/worktree/movelist.go:494:	if latest, ok := latestLineNamingPathsOf(string(data), line); ok && latest == line {
internal/cli/worktree/movelist.go:514:func revLines(data string) []string {
internal/cli/worktree/movelist.go:531:	rev := revLines(data)
...
```

(출력은 줄 끝을 `...`로 줄였다. 위치와 결과 토큰은 그대로다.)

- `revLines`의 운영 코드 호출자는 `latestLineNamingPathsOf` 하나뿐이다(movelist.go:531). main에서는 `appendMoveListLine`이 직접 불렀고, 지금은 이 새 함수를 거쳐서만 불린다. 다른 소비자가 없으니 순서 의미가 바뀌어 영향받는 곳이 없다.
- `appendMoveListLine`의 쓰기 호출자는 migrate.go의 15곳이다. 모두 `m6MoveLineTriggerReason`이 만든 JSON 레코드를 넘기고, 이 레코드는 `old_path`와 `new_path`가 비어 있지 않다. 그래서 운영 경로에서는 언제나 "경로 일치" 분기를 탄다. 비레코드 분기(`rev[0]` 비교)는 테스트에서만 탄다.
- 테스트에서 이 함수들을 부르는 파일: `migrate_hardening_e_t2053_test.go`, `movelist_dedupe_t2079_test.go`, `migrate_hardening_c_t2053_test.go`, `movelist_dedupe_t2106_test.go`.

### E4. 선별 테스트 (-v)

```
$ timeout 300 go test ./internal/cli/worktree/ -run 'AppendMoveListLine|MoveList|Dedupe|PendingIntentButNeverMoved' -count=1 -v > base.txt 2>&1; echo "exit=$?"
exit=0
$ grep -c '^=== RUN' base.txt
11
$ grep -c 'no tests to run' base.txt
0
--- PASS: TestApply_PendingIntentButNeverMoved_IsSettledThenMoved (2.95s)
--- PASS: TestMovedTo_UnreadableMoveList_GivesNoAnswer (3.21s)
--- PASS: TestAppendMoveListLine_UnreadableList_IsAnError (0.00s)
--- PASS: TestPlan_UnreadableMoveList_IsARunError (2.92s)
--- PASS: TestAppendMoveListLine_EqualToEarlierEntryButNotLatest_IsAppended (0.01s)
--- PASS: TestAppendMoveListLine_EqualToLatestEntry_IsNotAppended (0.00s)
--- PASS: TestAppendMoveListLine_IdenticalRerunRightAfterAddsNothing (0.00s)
--- PASS: TestAppendMoveListLine_SameTextAfterASettlementOfTheSamePathsIsAppended (0.00s)
--- PASS: TestAppendMoveListLine_DuplicateOfTheLatestRecordIsDroppedWhateverComesFirst (0.00s)
--- PASS: TestAppendMoveListLine_AnotherTreesRecordInBetweenDoesNotChangeTheComparison (0.11s)
--- PASS: TestAppendMoveListLine_ANonRecordLineIsComparedWithTheLatestLineOnly (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	10.006s
```

11개 실행, 11개 통과. t2079 테스트 2개(`EqualToEarlierEntryButNotLatest_IsAppended`, `EqualToLatestEntry_IsNotAppended`)와 t2106 테스트 5개가 모두 들어 있다. 병합 직후 실패했다고 보고된 4개 테스트가 이제 통과한다.

### E5. 예전 불안정 테스트 10회 반복

```
$ timeout 300 go test ./internal/cli/worktree/ -run 'TestApply_PendingIntentButNeverMoved_IsSettledThenMoved$' -count=10 > flaky10.txt 2>&1; echo "exit=$?"
exit=0
$ cat flaky10.txt
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	51.340s
```

### E6. 돌연변이 탐침 (`go test -overlay`, 트리는 바꾸지 않음)

돌연변이 소스는 스크래치 디렉터리에 만들었고, `-overlay` JSON으로 `movelist.go`만 바꿔 끼웠다. 추적 파일은 건드리지 않았다(E9).

- M1: `revLines` 끝에 교환 블록을 다시 넣는다(병합 때 생긴 이중 반전 재현).
- M2: 중복 검사를 main의 `rev[0] == line` 비교로 되돌린다(t2106의 의미 변화 제거).
- M3: `namesAny`가 `old_path` 위치만 본다(`new_path` 위치 일치를 지운다).

```
$ timeout 300 go test -overlay .../m1.json ./internal/cli/worktree/ -run '<같은 패턴>' -count=1 -v > m1.txt 2>&1; echo "exit=$?"
exit=1
== m1
--- FAIL: TestApply_PendingIntentButNeverMoved_IsSettledThenMoved (5.01s)
--- FAIL: TestAppendMoveListLine_EqualToEarlierEntryButNotLatest_IsAppended (0.01s)
--- FAIL: TestAppendMoveListLine_EqualToLatestEntry_IsNotAppended (0.01s)
--- FAIL: TestAppendMoveListLine_SameTextAfterASettlementOfTheSamePathsIsAppended (0.01s)
--- FAIL: TestAppendMoveListLine_ANonRecordLineIsComparedWithTheLatestLineOnly (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	15.296s
11

$ ... -overlay .../m2.json ... ; echo "exit=$?"
exit=1
== m2
--- FAIL: TestAppendMoveListLine_AnotherTreesRecordInBetweenDoesNotChangeTheComparison (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	10.140s
11

$ ... -overlay .../m3.json ... ; echo "exit=$?"
exit=0
== m3
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	10.596s
11
```

M1 죽음(테스트 5개 빨강), M2 죽음(1개 빨강), M3 생존(F5 참고).

### E7. 커버리지 (선별 테스트 기준)

```
$ timeout 300 go test ./internal/cli/worktree/ -run '<같은 패턴>' -count=1 -coverprofile=cover.out > cover.txt 2>&1; echo "exit=$?"
exit=0
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	23.621s	coverage: 27.6% of statements
$ go tool cover -func=cover.out | grep -E 'appendMoveListLine|revLines|latestLineNamingPathsOf|namesAny'
movelist.go:486:		appendMoveListLine			75.0%
movelist.go:514:		revLines				100.0%
movelist.go:530:		latestLineNamingPathsOf			92.3%
movelist.go:553:		namesAny				100.0%
```

패키지 전체 27.6%는 선별 테스트만 돌린 값이라 패키지 커버리지가 아니다. 판단에 쓴 값은 바뀐 함수별 수치다.

### E8. 형식·정적 검사, 보안 탐침

```
$ gofmt -l internal/cli/worktree/; echo "gofmt_exit=$?"
gofmt_exit=0
$ go vet ./internal/cli/worktree/; echo "vet_exit=$?"
vet_exit=0
$ grep -n -e 'json.Unmarshal' -e 'os.ReadFile' -e 'exec.Command' internal/cli/worktree/movelist.go
...
487:	data, rerr := os.ReadFile(mlPath)
536:	if err := json.Unmarshal([]byte(line), &want); err != nil || (want.OldPath == "" && want.NewPath == "") {
541:		if err := json.Unmarshal([]byte(l), &have); err != nil {
```

카드가 더한 코드는 이미 읽은 로컬 파일 내용을 고정 구조체로 언마샬하고 문자열을 비교할 뿐이다. 새 셸 실행, 새 파일 경로 조립, 외부 입력 경계가 없다. 파싱에 실패한 줄은 건너뛴다(541행).

### E9. 범위

```
$ git diff --name-only b25adf4fca HEAD
.moai/reports/t2106/progress.md
.moai/reports/t2106/sync-audit.md
.moai/reports/t2106/verdict.md
internal/cli/worktree/movelist.go
internal/cli/worktree/movelist_dedupe_t2106_test.go
$ git log --format='%h %s' b25adf4fca..HEAD
56017148b4 [t2106] fix(worktree): drop the second reversal of revLines ...
8bac970c9f [t2106] merge: main b25adf4fca into the card branch ...
2e2eb1a2b8 [t2106] docs: verdict.md ...
cd86c39b4c [t2106] merge: local main 04d8c27a7e into the card branch before landing ...
1cb6d67a2a [t2106] docs(audit): sync-audit verdict PASS-WITH-DEBT 94 for 901273f155 ...
901273f155 [t2106] docs: progress record ...
af81b0db5a [t2106] fix(worktree): move-list duplicate check ...
```

두 점(`b25adf4fca HEAD`) 비교로 트리 대 트리를 봤다. 위 5개 파일 말고는 HEAD 트리가 main 트리와 같다. 즉 병합이 main 내용 이외의 파일을 만들거나 바꾸지 않았다. 브랜치 커밋은 모두 카드 id를 달고 있다.

### E10. 감사 계획

```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2106 > plan.json 2>&1; echo "exit=$?"
exit=0
"config_status": "ok", "model": "codex",
backends: claude gate off / codex gate advisory / glm gate off
"cross_model_active": true, "cross_model_required": false, "enforced_required": []
```

required 백엔드가 없다. advisory인 Codex 다리는 운영자 결정으로 돌리지 않았다(Gaps G1).

## 판단 근거 (질문별)

**1. 한 번만 뒤집는가, 두 수리가 합쳐지는가.** 그렇다. E2의 `revLines`는 끝에서부터 도는 루프 하나뿐이고 교환 블록이 없다. 함수 본문은 main과 바이트 단위로 같다(movelist.go diff에 `revLines` 본문 줄이 없다). 이중 반전을 다시 넣으면(M1) 테스트 5개가 실패하므로, 상쇄 상태는 지금 테스트가 잡아낸다.

**main 대비 동작 회귀 여부.** `revLines`의 다른 호출자가 없으므로 순서 의미 변화는 퍼지지 않는다(E3). `appendMoveListLine`에 대해서는 아래처럼 추론했다.
- main은 `rev[0] == line`일 때 건너뛴다. 이때 `line`이 경로를 가진 레코드라면 `rev[0]`도 같은 텍스트라서 같은 경로를 담고 있고, 새 코드의 루프는 첫 원소에서 바로 그것을 돌려준다. 그러니 새 코드도 건너뛴다. `line`이 비레코드라면 새 코드도 `rev[0]`과 비교하므로 같다.
- 따라서 main이 건너뛰는 경우는 새 코드도 모두 건너뛴다. 새로 건너뛰는 경우는 "같은 경로의 최신 레코드가 새 줄과 같고, 그 뒤에 다른 경로의 줄이 있는 경우"뿐이다. 반대로 같은 경로의 레코드가 없으면(`ok=false`) 새 코드는 추가하고, main도 `rev[0] != line`이라 추가한다.
- 이 차이를 지키는 테스트는 `AnotherTreesRecordInBetweenDoesNotChangeTheComparison`이다(M2에서 실패).

**4. 어느 쓰기 경로의 의미가 바뀌는가.** E3의 migrate.go 15곳 모두 같은 함수로 레코드를 쓰므로 같은 방식으로 바뀐다. 실제로 차이가 나려면 같은 트리에 대해 trigger·result·경로·ident·reason이 같고 `ts`(초 단위)까지 같은 줄을, 다른 트리의 줄을 사이에 두고 다시 쓰려 해야 한다. 한 실행에서 여러 트리를 같은 초 안에 처리하는 루프(1253/1314행의 intent→moved 흐름, 2042~2203행의 복원 흐름)에서 이론상 가능하다. 이 경우 새 동작(추가 안 함)이 함수 주석의 규칙과 맞는다. 쓰기 경로별로 이 끼어들기를 직접 재현하는 통합 테스트는 없다. 일반 단위 테스트(`AnotherTreesRecordInBetween...`)만 있다. 이것은 F6으로 기록했다(optional).

## Findings

| ID | 심각도 | 분류 | 위치 | 내용 | 필요한 수정 |
|----|--------|------|------|------|-------------|
| F5 | MINOR | optional, in-diff, 확신도 높음(측정) | internal/cli/worktree/movelist.go:553-560, movelist_dedupe_t2106_test.go | `namesAny`가 `new_path` 위치로 맞추는 동작을 구별하는 테스트가 없다. M3(`old_path`만 비교)가 11개 테스트를 모두 통과했다(E6). 기존 테스트에서는 공유 경로가 늘 `old_path` 쪽으로도 일치하기 때문이다. | 원하면 테스트 한 줄을 더한다. 파일 `[pending A(/r/a→/h/a), moved X(/x→/h/a)]`에 pending A를 추가하면 3줄이 되어야 한다(M3라면 2줄). 필수는 아니다. |
| F6 | MINOR | optional, touched-file, 확신도 중간(추론) | internal/cli/worktree/migrate.go 15개 쓰기 지점 | 다른 트리의 줄이 끼어든 뒤 같은 초에 같은 줄을 다시 쓰는 경우의 의미 변화(추가 → 건너뜀)를 쓰기 경로 단위로 재현하는 테스트가 없다. 단위 수준에서는 덮여 있다. | 조치하지 않아도 된다. 쓰기 경로가 같은 초 재시도를 일부러 하게 되면 그때 통합 테스트를 더한다. |
| F1(이월) | MINOR | optional, in-diff | movelist_dedupe_t2106_test.go:113 | 1회차 F1 그대로. 주석은 "Control:"인데 실제로는 옛 코드에서 실패하던 RED 테스트다. 이번 M1에서도 이 테스트가 실패했다. | "Control:"을 "RED before the repair:"로 고친다. |
| F2(이월) | MINOR | optional, in-diff | movelist.go:553-560 | 1회차 F2 그대로. 경로를 정확한 문자열로 비교한다(같은 파일의 다른 판단은 `worktreeroot.Same`). | 표기가 섞이는 쓰기 경로가 생기면 바꾼다. |
| F3(이월) | MINOR | optional, in-diff | movelist.go:539-548 | 1회차 F3 그대로. 경로가 일치하는 레코드가 없으면 파일 전체를 파싱한다. 측정하지 않았다. | 이동 목록이 수만 줄을 넘으면 측정한다. |
| F4(이월) | MINOR | optional, outside(판정에 영향 없음) | movelist.go:479-481 | 1회차 F4 그대로. 초 단위 `ts` 때문에 "identical re-run adds nothing"은 같은 초 안에서만 성립한다. 수리 전부터 있던 성질이다. | 운영자에게 보여 주기만 한다. |

BLOCKING 결함: 0건. repair-born(이전 감사 SHA 이후 diff에서 생긴) blocking 결함도 없다. `8bac970c9f` 병합이 만든 이중 반전은 이번 감사 전에 `56017148b4`가 이미 닫았고, M1이 그 재발을 막는 테스트가 있음을 보여 준다.

## Dimension scores

| 차원 | 점수 | 판정 | 근거 |
|------|------|------|------|
| Functionality (40%) | 97 | PASS | E4 `ok ... 10.006s`, 11/11 통과(t2079 2개, t2106 5개 포함). E5 10회 반복 `ok ... 51.340s`. M1이 5개를 빨갛게 만든다(E6). |
| Security (25%) | 100 | PASS | E8: 새 코드는 로컬 파일 내용 언마샬과 문자열 비교뿐이다. 새 외부 입력·셸·경로 조립 없음. `go vet` exit 0. |
| Craft (20%) | 90 | PASS | E7: `revLines` 100%, `namesAny` 100%, `latestLineNamingPathsOf` 92.3%, `appendMoveListLine` 75.0%. M3 생존(F5), 주석 오표기 이월(F1). |
| Consistency (15%) | 93 | PASS | E8 `gofmt_exit=0`, `vet_exit=0`. 브랜치 커밋 모두 카드 id 표기(E9). 경로 비교 방식이 파일의 다른 판단과 다르다(F2 이월). |

가중 합: 0.40×97 + 0.25×100 + 0.20×90 + 0.15×93 = 95.75 → 95. must-pass 차원(Functionality, Security) 모두 통과.

## Baseline-attribution

- 측정 대상 트리: HEAD `56017148b4ba7c06abde8fdcc913cd647f0d572a`, `git status --short` 빈 출력(E1). 흡수 기준 main `b25adf4fca05e6dba07abbdf307ccaf2b80b9033`, HEAD의 조상임을 확인했다.
- 테스트, 커버리지, vet, gofmt는 모두 이 트리에서 `go` 툴체인이 소스를 바로 빌드해 측정했다. 설치된 `moai` 바이너리(`v3.2.1-rc.3`, `v3.2.0-1369-gf9d8866951`)는 감사 계획 읽기(E10)에만 썼다. 그 결과는 이 트리의 설정 파일을 읽은 값이고 코드 판단에는 쓰지 않았다.
- 돌연변이는 `go test -overlay`로 같은 트리 위에 `movelist.go` 한 파일만 바꿔 끼워 측정했다.
- 1회차(901273f155) 측정값은 다시 쓰지 않았다. 이월 결함 F1~F4는 이번 트리에서 해당 줄을 다시 읽어 여전히 있음을 확인했다(테스트 파일 113행 주석, movelist.go 553-560행).

## Gaps

- G1. Codex(advisory)와 GLM(off) 다리는 돌리지 않았다. 운영자 결정 d-20261106T051044Z-7a4c(Claude 단독)에 따른 것이다. 감사 계획상 required 백엔드가 없어(`enforced_required: []`) 충족되지 않은 게이트는 없다. 이번 판정에는 교차 모델 검토가 없다.
- G2. 패키지 전체(`go test ./internal/cli/worktree/`)와 저장소 전체 테스트는 돌리지 않았다. 지시대로 선별 패턴만 실행했다. 원격 CI와 다른 OS(Windows, Linux)는 확인하지 않았다.
- G3. 돌연변이 명령을 처음에 `| grep` 파이프 형태로 냈을 때 워크트리 격리 가드가 3건 모두 거부했다(실행되지 않음). 같은 명령을 파일 리디렉션 형태로 다시 내서 실제로 실행했고, 위 E6 결과는 그 재실행 결과다. 원본 소스를 읽는 것으로 대체하지 않았다. 거부 기록은 `.moai/lessons-inbox.jsonl`에 `tool_failure` 행으로 남았을 수 있다.
- G4. F6(쓰기 경로별 의미 변화)은 코드를 읽고 추론한 것이고, 쓰기 경로 단위 재현은 하지 않았다.
- G5. `latestLineNamingPathsOf`의 덮이지 않은 7.7%가 어느 분기인지 줄 단위로 확인하지 않았다(빈 파일 분기로 추정).
- G6. 불안정 테스트 반복은 10회만 했다. 조건이 같은 초 경계에 걸려 있어 10회 통과로 0%를 증명하지는 못한다. 다만 원인(첫 줄 비교)이 코드에서 사라졌고 M1이 그 재발을 잡는다.

## Residual-risk

- 같은 초 안에서 다른 트리 줄이 끼어든 뒤 같은 줄을 다시 쓰는 경로가 있으면, 이제 그 줄이 기록되지 않는다. 함수 주석의 규칙과 맞는 동작이지만, 그 줄을 별도 사건으로 세던 소비자가 있다면 결과가 달라질 수 있다. 이동 목록 소비자(`movedTo` 계열)는 최신 위치만 읽으므로 영향이 없다고 보지만, 소비자 전체를 열거해 확인하지는 않았다.
- `namesAny`의 `new_path` 일치 분기는 테스트가 구별하지 못하므로(F5), 나중에 그 분기를 실수로 지워도 이 테스트 묶음은 초록으로 남는다.
- 이 판정은 로컬 선별 측정이다. 전체 회귀와 다른 OS는 CI에 맡겨져 있다.

## Iteration history

- iter 1: PASS-WITH-DEBT 94 @901273f155 (`.moai/reports/t2106/sync-audit.md`), Codex 영수증 rcpt-b7cae4d26e4f1d2872464a77. optional F1~F4.
- 사이에 일어난 일: t2079가 main에 먼저 들어가 같은 `revLines` 순서 버그를 루프 방향으로 고쳤다. `8bac970c9f`에서 main을 카드 브랜치에 병합할 때 충돌 없이 자동 병합됐지만, `revLines`가 두 번 뒤집혀 t2079 테스트 2개와 t2106 테스트 2개가 실패했다(카드 커밋 메시지 기록. 이번 감사는 병합 트리를 직접 측정하지 않았고, M1로 같은 상태를 재현해 실패를 관측했다). `56017148b4`가 t2106의 교환 블록을 지웠다.
- iter 2 (이번): PASS-WITH-DEBT 95 @56017148b4, Claude 단독, 영수증 없음. 새 optional F5(M3 생존), F6(쓰기 경로 단위 테스트 없음). F1~F4 이월.

## Operational Notes (unverified)

- inferred: F5를 닫는 테스트는 E6의 M3 overlay로 바로 확인할 수 있다. 측정 방법은 `go test -overlay <m3.json> ./internal/cli/worktree/ -run AppendMoveListLine -count=1`이 실패하는지 보는 것이다.
