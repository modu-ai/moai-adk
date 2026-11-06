auditor-model: claude-opus-5-5

# t2106 sync 감사 판정

verdict: PASS-WITH-DEBT
audited_sha: 901273f1558dcadfb090f564d14e90270ef148d8

- 카드: t2106 (Class B, `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved` 불안정)
- 감사 대상: `git diff 9cf8a1fc14 HEAD -- internal` (movelist.go, movelist_dedupe_t2106_test.go)
- 점수: 94/100 (Functionality 97 / Security 100 / Craft 90 / Consistency 92)
- Codex 영수증: rcpt-b7cae4d26e4f1d2872464a77 (codex_audit 1회, native, baseBranch, 결과 pass, 지적 0건)
- BLOCKING 결함: 없음. 남은 것은 optional MINOR 3건(카드 diff 안)과 범위 밖 참고 1건.

## Claim

1. 주장된 원인이 실제 원인이다. 옛 `revLines`는 "last first"라고 적혀 있으면서 파일 순서를 돌려줬고, 그래서 `appendMoveListLine`은 새 줄을 파일의 **첫 줄**과 비교했다. `ts`가 초 단위라서, 테스트가 심어 둔 첫 `pending`과 엔진이 새로 쓰는 `pending`이 같은 초에 만들어질 때만 텍스트가 같아지고, 그때 새 줄이 "중복"으로 버려져 결과가 `pending,reversed,moved`가 된다.
2. 수리 뒤에는 해당 테스트가 시계와 무관하게 통과한다.
3. 새 단위 테스트 가운데 3개는 옛 코드에서 의도한 이유로 실패하고(RED), 2개는 수리 전후 모두 통과하는 대조군이다.
4. 새 중복 규칙은 함수 주석대로 "새 줄의 두 경로 중 하나라도 이름에 담은 가장 최근 레코드"와 비교한다. old_path와 new_path를 위치에 상관없이 대조하므로, `pending old->new` 뒤의 `reversed new->old`도 같은 경로를 가리키는 레코드로 본다. 레코드가 아니거나 경로가 빈 줄은 파일의 마지막 줄하고만 비교한다.
5. 이동 목록에 쓰는 코드는 모두 `appendMoveListLine`을 거치고, 수리 때문에 기록돼야 할 줄이 빠지는 경우는 없다. 버려지는 줄은 같은 경로를 가리키는 최신 레코드와 바이트 단위로 같은 줄뿐이다.
6. 형식 검사와 vet가 깨끗하다.

## Evidence

**HEAD 확인**
```
$ git rev-parse HEAD
901273f1558dcadfb090f564d14e90270ef148d8
$ git rev-parse --show-toplevel
/Users/goos/.moai/worktrees/projects/moai-adk-go/t2106
```

**원인 코드 대조 (읽기)**
- `migrate.go:1905-1913` `m6MoveLineTriggerReason`: `"ts": time.Now().UTC().Format(time.RFC3339)`, 즉 초 단위 타임스탬프가 줄 텍스트에 들어간다.
- `migrate_hardening_e_t2053_test.go:54`: 픽스처가 `m6MoveLine("pending", w.movedTree, dest, ident)`를 빈 이동 목록의 첫 줄로 심는다.
- `migrate.go:605`: 중단된 단계를 정리하면서 `reversed`(new->old, reason 포함)를 쓴다.
- `migrate.go:915`: 엔진이 `m6MoveLineTrigger(trigger, "pending", p.source, p.destination, ...)`를 쓴다. trigger가 `migrate`이면 픽스처 줄과는 `ts`만 다르다.
- `migrate.go:967`: `moved`를 쓴다. 옛 코드에서도 첫 줄(`pending`)과 다르므로 항상 기록된다.
- 결론: 옛 코드에서 버려질 수 있는 줄은 엔진의 `pending` 하나뿐이고, 버려지면 결과가 정확히 `pending,reversed,moved`가 된다. 같은 초일 때만 버려지므로 실패가 들쭉날쭉한 것도 설명된다. 이 테스트에서 시계에 따라 결과가 달라지는 다른 경로는 찾지 못했다(롤백 단계는 `failed`로 정리된 단계를 건너뛰고, 나머지 검사는 파일 상태만 본다).

**옛 코드 재현: 트리를 바꾸지 않고 `go test -overlay`로 옛 movelist.go만 바꿔 끼움**

옛 파일은 `git archive --output=<scratchpad>/old-movelist.tar 9cf8a1fc14 internal/cli/worktree/movelist.go`로 scratchpad에 풀었고, overlay JSON도 scratchpad에 두었다. 트리 안에는 아무것도 쓰지 않았다.
```
$ go test -overlay=<scratchpad>/overlay.json ./internal/cli/worktree -run "TestApply_PendingIntentButNeverMoved_IsSettledThenMoved" -count=10
--- FAIL: TestApply_PendingIntentButNeverMoved_IsSettledThenMoved (4.60s)
    migrate_hardening_e_t2053_test.go:118: move-list results = pending,reversed,moved, want pending,reversed,pending,moved
--- FAIL: TestApply_PendingIntentButNeverMoved_IsSettledThenMoved (2.90s)
    migrate_hardening_e_t2053_test.go:118: move-list results = pending,reversed,moved, want pending,reversed,pending,moved
--- FAIL: TestApply_PendingIntentButNeverMoved_IsSettledThenMoved (4.45s)
    migrate_hardening_e_t2053_test.go:118: move-list results = pending,reversed,moved, want pending,reversed,pending,moved
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	48.979s
FAIL
```
→ 옛 코드에서 10회 중 3회 실패했고, 세 번 모두 주장된 증상과 같다.

```
$ go test -overlay=<scratchpad>/overlay.json ./internal/cli/worktree -run "TestAppendMoveListLine_" -count=1 -v
--- PASS: TestAppendMoveListLine_UnreadableList_IsAnError (0.00s)
--- PASS: TestAppendMoveListLine_IdenticalRerunRightAfterAddsNothing (0.00s)
    movelist_dedupe_t2106_test.go:73: pending, reversed, pending must stay three records; got 2 lines
--- FAIL: TestAppendMoveListLine_SameTextAfterASettlementOfTheSamePathsIsAppended (0.00s)
    movelist_dedupe_t2106_test.go:88: a duplicate of the latest record for these paths was appended: 3 lines, want 2
--- FAIL: TestAppendMoveListLine_DuplicateOfTheLatestRecordIsDroppedWhateverComesFirst (0.00s)
--- PASS: TestAppendMoveListLine_AnotherTreesRecordInBetweenDoesNotChangeTheComparison (0.00s)
    movelist_dedupe_t2106_test.go:123: an immediate repeat of a non-record line was appended: 3 lines, want 2
    movelist_dedupe_t2106_test.go:130: a non-record line equal only to the FIRST line must be appended: 2 lines, want 3
--- FAIL: TestAppendMoveListLine_ANonRecordLineIsComparedWithTheLatestLineOnly (0.01s)
FAIL
```
→ RED 3개는 각각 의도한 이유로 실패한다(첫 줄 비교 때문에 버려짐 / 첫 줄이 다른 트리라서 중복이 추가됨 / 비레코드 줄이 양방향으로 틀림). 대조군 2개는 옛 코드에서도 통과한다. 단위 테스트는 같은 문자열 변수를 재사용하므로 시계에 의존하지 않는다.

**수리된 코드 (HEAD 트리)**
```
$ go test ./internal/cli/worktree -run "TestAppendMoveListLine_" -count=1 -v
--- PASS: TestAppendMoveListLine_UnreadableList_IsAnError (0.00s)
--- PASS: TestAppendMoveListLine_IdenticalRerunRightAfterAddsNothing (0.00s)
--- PASS: TestAppendMoveListLine_SameTextAfterASettlementOfTheSamePathsIsAppended (0.00s)
--- PASS: TestAppendMoveListLine_DuplicateOfTheLatestRecordIsDroppedWhateverComesFirst (0.00s)
--- PASS: TestAppendMoveListLine_AnotherTreesRecordInBetweenDoesNotChangeTheComparison (0.00s)
--- PASS: TestAppendMoveListLine_ANonRecordLineIsComparedWithTheLatestLineOnly (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	0.744s

$ go test ./internal/cli/worktree -run "TestApply_PendingIntentButNeverMoved_IsSettledThenMoved" -count=10
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	40.226s

$ go test ./internal/cli/worktree -run "MovedBy|MovedTo|MoveList|Interrupted|Reverse|Rollback" -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	214.235s
```

**정적 검사**
```
$ gofmt -l internal/cli/worktree
(출력 없음)
$ go vet ./internal/cli/worktree
(출력 없음)
$ env GOOS=windows go vet ./internal/cli/worktree
(출력 없음)
```

**이동 목록에 쓰는 코드 전수 확인**
```
$ grep -rn -e "moveListPath(" -e "move-list.jsonl" -e "appendMoveListLine" internal --exclude="*_test.go"
```
테스트가 아닌 코드에서 쓰기 경로는 `appendMoveListLine` 하나뿐이다(migrate.go:605, 915, 933, 941, 967, 1185, 1189, 1194, 1204, 1212, 1215, 1673, 1682, 1766, 1834). 다른 곳(`movelist.go:224`의 `loadMoveList`, migrate.go:533/625/654/805/810)은 읽기이거나 디렉터리 경로 계산이다. 각 호출 지점을 차례로 확인했다.
- 913-943 정방향 이동: `pending`(src->dst) → 실패 시 `reversed`(dst->src, reason) 또는 `pending`(dst->src, msg) → 성공 시 `moved`. 연이은 줄은 result, 방향, reason 가운데 하나가 다르므로 버려지지 않는다.
- 1185-1215 역방향 이동: `pending`/`refused`/`reversed`에 reason이 각기 달라서 같은 텍스트가 연달아 나오지 않는다.
- 1673/1682 롤백 refuse/hold, 1766/1834 `restored`: 단계마다 경로가 다르다. 같은 단계가 같은 초에 같은 `restored`를 두 번 쓰면 둘째 줄은 버려지는데, 내용이 같은 진짜 중복이므로 의도한 동작이다.
- 옛 우연 동작(첫 줄 비교)에 기대는 테스트가 있는지 찾았다: `grep -rn -e "pending,reversed" -e "t2053ResultsFor(" -e "IdenticalRerun" -e "identical re-run" internal/cli/worktree`로 순서를 단정하는 테스트는 대상 테스트 하나뿐임을 확인했다. e_t2053:362-410의 `moved` 줄 추가는 반환값을 버리고 MovedBy/MovedTo만 확인하므로 줄 수에 의존하지 않는다.

**경계 조건 (코드 읽기)**
- 빈 파일, 또는 없는 파일(`fs.ErrNotExist` → data 빈 문자열): `revLines`가 빈 결과를 돌려주고 ok=false가 되어 추가한다. 정상이다.
- 마지막 줄 끝에 개행이 없는 파일: `strings.Split`가 마지막 줄을 그대로 돌려주므로 비교가 맞다. 개행 보정(502-506)은 바뀌지 않았다.
- 파일 안의 파싱 불가 줄: 경로 탐색에서 건너뛴다(545-547). 새 줄이 레코드가 아니면 마지막 줄(rev[0])과만 비교한다.
- 읽기 실패: 바뀌지 않았다(오류 반환). `TestAppendMoveListLine_UnreadableList_IsAnError` 통과.

**Codex (요청받은 1회 호출)**
```
mcp__moai__codex_audit(project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2106, target=baseBranch, mode=native)
→ verdict: pass, findings: [], audit_receipt: rcpt-b7cae4d26e4f1d2872464a77,
  build_commit: 646955327e, review_base: main (merge base 9cf8a1fc147605a8221206616d04fb43569961d8),
  resolved_model: gpt-6.1-sol
  summary: "지정한 기준 커밋 대비 변경분을 정적으로 검토했으나 지적할 결함은 찾지 못했어. 지침대로 테스트·빌드·재현 명령은 실행하지 않았어."
```
Codex가 찾은 것: 없음(지적 0건, 정적 검토만 수행).

호출 뒤 트리 상태:
```
$ git status --short
(출력 없음)
```

## Findings

| id | 심각도 | 분류 | 위치 | 요약 | 필요한 조치 |
|---|---|---|---|---|---|
| F1 | MINOR | optional, in-diff, 확신도 높음 | internal/cli/worktree/movelist_dedupe_t2106_test.go:113 | `ANonRecordLineIsComparedWithTheLatestLineOnly` 주석이 "Control"이라고 적혀 있는데, 실제로는 옛 코드에서 양방향 모두 실패하는 RED 테스트다(progress.md도 RED로 분류함). | 주석의 "Control:"을 "RED before the repair:"로 고친다. |
| F2 | MINOR | optional, in-diff, 확신도 중간 | internal/cli/worktree/movelist.go:557-563 | `namesAny`는 경로를 정확한 문자열로 비교하는데, 같은 파일의 다른 판단은 `worktreeroot.Same`을 쓴다. 같은 트리를 다른 표기(심볼릭 링크 해석 차이 등)로 적은 최신 레코드를 놓치고, 더 오래된 같은 표기 레코드와 비교할 수 있다. 실제로 잘못 버려지려면 그 오래된 레코드가 `ts`까지 바이트 단위로 같아야 하므로(같은 초) 사실상 일어나기 어렵다. | 지금은 그대로 두어도 된다. 경로 표기가 섞이는 쓰기 경로가 생기면 `worktreeroot.Same`으로 바꾼다. |
| F3 | MINOR | optional, in-diff, 확신도 중간 | internal/cli/worktree/movelist.go:543-551 | 경로를 담은 레코드가 하나도 없으면(트리의 첫 intent 줄) 줄을 추가할 때마다 파일 전체를 JSON 파싱한다. 실행 한 번의 비용은 O(트리 수 × 줄 수)다. 수천 줄이면 작은 구조체 수천 번 언마샬이라 수용할 만하지만 측정하지는 않았다. 옛 코드도 매번 파일 전체를 읽고 분할했다. | 지금은 조치하지 않는다. 이동 목록이 수만 줄을 넘으면 측정한다. |
| F4 | MINOR | optional, outside(범위 밖, 판정에 영향 없음) | internal/cli/worktree/movelist.go:479-481, migrate.go:1907 | 줄 텍스트에 초 단위 `ts`가 들어 있어서, 주석의 "an identical re-run adds nothing"은 같은 초 안에서만 성립한다. 수리 전부터 있던 성질이고 이 카드가 바꾼 것이 아니다. | 운영자에게 보여 주기만 한다. 큐에 넣지 않는다. |

수리가 낳은 결함(repair-born): 해당 없음(첫 감사).

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 97 | PASS | 옛 코드 overlay에서 3/10 실패(`pending,reversed,moved`), 수리 뒤 10/10 통과, 단위 테스트 6/6 통과, 이동 목록 관련 선별 테스트 `ok ... 214.235s` |
| Security (25%) | 100 | PASS | 신뢰 경계 변화 없음. 파일 읽기 오류 처리는 그대로이고, 입력은 같은 프로세스가 쓴 이동 목록이며, 새 외부 입력 표면이 없다. JSON 파싱 실패는 건너뛰기(fail-safe) |
| Craft (20%) | 90 | PASS | gofmt/vet(darwin, windows) 깨끗, RED→GREEN이 결정적. F1 주석 오표기, F3 비용 미측정. 커버리지는 측정하지 않음(Gaps) |
| Consistency (15%) | 92 | PASS | 파일의 기존 스타일과 주석 어조가 일치한다. F2: 경로 비교 방식이 파일 안의 `worktreeroot.Same` 관례와 다르다 |

must-pass(Functionality, Security) 모두 PASS.

## Baseline-attribution

- 측정 트리: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2106`, HEAD `901273f1558dcadfb090f564d14e90270ef148d8`, `git status --short` 깨끗(Codex 호출 전후 모두).
- 옛 코드 기준선: `9cf8a1fc14`의 `internal/cli/worktree/movelist.go`. `git archive`로 scratchpad에 풀어 `go test -overlay`로 바꿔 끼웠고, 같은 트리의 다른 파일(새 테스트 포함)과 함께 이번 실행에서 측정했다.
- 판정한 빌드: `go test`가 이 트리 소스를 직접 컴파일하므로 판정 빌드와 측정 트리가 같다. Codex는 MCP 서버 빌드 `646955327e`(설치본)를 통해 호출됐고, 리뷰 대상 diff는 서버가 이 트리에서 직접 읽었다(review_base merge base 9cf8a1fc14).
- progress.md의 수치 대조: "이 트리, 수리 전 1/10"과 "t2070 트리 4/10"은 작성자의 측정이라 다시 관측하지 않았다. 내가 이번에 잰 옛 코드 수치는 3/10으로, 같은 증상이 같은 테스트에서 나왔다. "수리 뒤 25/25"는 다시 돌리지 않았고 내 측정은 10/10이다. "504.6 s 스윕 ok"도 다시 돌리지 않았고 대신 선별 스윕(`MovedBy|MovedTo|MoveList|Interrupted|Reverse|Rollback`)을 돌려 ok를 확인했다. progress.md는 25/25가 증명이 아니라는 점(실패율 1/10에서 약 7% 확률로 우연히 나옴)을 스스로 밝혀 두었고, 수치의 출처도 정직하게 적었다. Gaps 항목도 실제 미실행 사항과 일치한다.

## Gaps

- `moai verify audit-plan`을 실행하지 않았다. 스폰 지시가 허용 명령을 read-only git, vet, gofmt, 선별 go test로 한정했고, Codex 호출도 정확히 1회로 지정했기 때문이다. 감사 계획 대조(`--result-file` 검사)는 하지 않았다. `audit_multi`를 쓰지 않았으므로 오케스트레이터가 넘길 digest도 없다.
- 패키지 전체(`go test ./internal/cli/worktree`)와 `go test ./...`는 돌리지 않았다(리더 제약, 머신에서 회귀 실행 중).
- `-race`, golangci-lint, 커버리지 측정은 하지 않았다.
- Windows는 vet만 했고 테스트는 돌리지 않았다.
- F3의 비용은 측정하지 않은 추정이다.
- Codex는 정적 검토만 했다고 스스로 밝혔다(테스트·재현 미실행).
- 거부된 도구 호출: 셸 glob 확장 때문에 `--include=*_test.go` / `--exclude=*_test.go`를 따옴표 없이 쓴 grep 두 번이 zsh `no matches found`로 실행되지 않았다. 따옴표를 붙여 다시 실행했고 결과는 위 Evidence에 있다. 정책 거부는 없었다.

## Residual-risk

- 관측한 범위 안에서는 시계에 따라 결과가 달라지는 경로가 남아 있지 않다. 다만 패키지 전체를 돌리지 않았으므로, 같은 초에 같은 텍스트가 연달아 쓰이는 다른 시나리오(이제는 최신 레코드 기준으로 버려짐)를 단정하는 테스트가 선별 정규식 밖에 있다면 아직 드러나지 않았을 수 있다. 작성자의 504.6 s 마이그레이션 스윕은 ok였다고 기록돼 있다(재관측 안 함).
- 원격 CI(darwin/windows 매트릭스)가 최종 판정 표면이다.
- Codex 결과는 지적 0건이지만 정적 검토라서 실행 증거로 대신할 수 없다.

## Iteration history

- iter 1 (이번): PASS-WITH-DEBT 94 @901273f155, Codex 영수증 rcpt-b7cae4d26e4f1d2872464a77.
