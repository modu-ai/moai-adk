# 카드 t1281 sync 감사 — SPEC-STATUSLINE-LANDED-LABEL-001

- 감사 대상 트리: `WT-statusline-landed-label` HEAD `f4ca82858`, 카드 기준점 `CARD_BASE=$(git merge-base develop HEAD)` = `b59a5d69c`
- 감사자: sync-auditor (독립 재측정. `progress.md` 수치는 재사용하지 않음)
- 적용 규칙: `verification-claim-integrity.md` §1·§2·§3, `verification-completeness.md` §2(변이 탐침), `gitflow-lane-protocol.md` §8(범위는 merge-base 기준)

## 판정

**PASS-WITH-DEBT** — 차단 결함 0건, 선택 결함 4건.

| 차원 | 점수 | 판정 | 근거 요약 |
|---|---|---|---|
| Functionality (40%) | 96 | PASS | 패키지 테스트 통과, AC 18행 대응 테스트 확인, 기준 변이 4종 모두 검출됨 |
| Security (25%) | 95 | PASS | 새 입력면·서브프로세스·파일 쓰기 없음. diff 에 `exec.Command` 추가·삭제 0행 |
| Craft (20%) | 86 | PASS | 커버리지 90.4%, lint 0, vet 0. 다만 변이 D(스탬피드 가드 연결선) 생존 |
| Consistency (15%) | 92 | PASS | auto-done 과 동일한 호출 형태, 4개 로케일 문서, CHANGELOG 위치 정상 |

가중 합 93.2, 조화 평균 92.1. 필수 통과 차원(Functionality, Security) 모두 임계값 이상.

## 1. Claim (주장)

1. `go test ./internal/statusline/...` 가 이 트리에서 통과하고 커버리지는 90.4% 다.
2. golangci-lint v2.1.6 결과 0 issues, `go vet` 무출력으로 종료 0 이다.
3. 카드가 바꾼 Go 파일은 `internal/statusline` 의 3개뿐이며 `internal/kanban`, `internal/hook`, doctor 점검은 건드리지 않았다. 따라서 `TestBinaryLag` 와 `./internal/hook` 재측정은 **필요 없다**.
4. 집계 판정식은 `moai todo auto-done` 과 글자 그대로 같은 두 호출(`LandedAttributions(commits, LandedBranchFromRef(ref))`, `AutoDoneSubjectFresh(hit, AddedAt)`)을 쓴다. 두 번째 판정기는 없다.
5. 기호는 U+2691(바이트 `e2 9a 91`)이고, 렌더러 비테스트 코드와 4개 로케일 문서에 `✓` 가 남아 있지 않다.
6. 옛 기준 캐시는 표기를 내지 않고, 그 위에서 갱신이 실패해도 숫자 9 가 현재 기준 측정으로 읽히지 않는다.
7. picked 수는 착지 수로 줄지 않는다.
8. 판정서(`verdict.md`)에 기호 후보 3개와 기각 이유, 제외 부류, 「picked 에서 빼지 않는다」 유지 판단이 모두 있다.

## 2. Evidence (증거)

### 2.1 테스트·lint·vet

```
$ go test ./internal/statusline/... -count=1 -cover
ok  	github.com/modu-ai/moai-adk/internal/statusline	21.642s	coverage: 90.4% of statements
exit=0

$ golangci-lint version
golangci-lint has version v2.1.6 built with go1.26.8 from (unknown, modified: ?, mod sum: "h1:LXqShFfAGM5BDzEOWD2SL1IzJAgUOqES/HRBsfKjI+w=") on (unknown)
$ golangci-lint run ./internal/statusline/...
0 issues.

$ go vet ./internal/statusline/...
vet_exit=0
```

### 2.2 범위 (merge-base 기준)

```
$ git diff --name-only b59a5d69c..HEAD -- '*.go'
internal/statusline/landed.go
internal/statusline/landed_test.go
internal/statusline/renderer.go

$ git diff --name-only b59a5d69c..HEAD -- internal/kanban internal/hook internal/cli/doctor* | wc -l
       0
```

양성 대조: 같은 범위의 `'*.go'` 질의가 3행을 냈으므로 0행은 측정된 부재다. 변경된 Go 파일 diff 에서 `doctor|Doctor|NewHandler|hook\.` 를 grep 한 결과 0행이다.

### 2.3 문서·잔재 grep

```
$ grep -c '✓' docs-site/content/{ko,en,ja,zh}/advanced/statusline.md
en:0  zh:0  ko:0  ja:0
$ grep -c '⚑N' docs-site/content/{ko,en,ja,zh}/advanced/statusline.md
en:2  ja:3  ko:3  zh:3
$ grep -c -- '--format=%B' internal/statusline/landed.go
0
$ grep -n '✓' internal/statusline/*.go | grep -v _test.go
(0행)
```

양성 대조: `grep -c '✓' internal/statusline/landed_test.go` → 3(부정 단언 3곳). grep 자체는 `✓` 를 찾는다.

```
$ grep -o 'landedGlyph = "[^"]*"' internal/statusline/landed.go | xxd
00000000: 6c61 6e64 6564 476c 7970 6820 3d20 22e2  landedGlyph = ".
00000010: 9a91 220a                                ..".
```

### 2.4 변이 탐침 (임시 수정 후 `git checkout --` 로 복원)

| 변이 | 내용 | 결과 |
|---|---|---|
| A | 세대 경계(`AutoDoneSubjectFresh`) 제거 | `--- FAIL: TestRefreshLandedCounts_GenerationBoundary` — 검출 |
| B | 귀속 실패 시 제목의 `\bID\b` 언급도 세기 | `--- FAIL: TestRefreshLandedCounts_GenerationBoundary`, `--- FAIL: TestRefreshLandedCounts_NonAttributingMentionDoesNotCount` (`landed_test.go:745: ... Landed:1 ...`) — 검출 |
| C | `Known()` 에서 기준 비교 제거 | `--- FAIL: TestResolveLandedCounts_OldCriterionIsUnknown` — 검출 |
| D | `RefreshLandedCounts` 의 옛 기준 prev 초기화(3행) 제거 | `ok github.com/modu-ai/moai-adk/internal/statusline 1.703s` — **생존** |
| E | 기호를 `✓` 로 되돌림 | `--- FAIL: TestLandedGlyph_SingleCellLocaleNeutral`, `--- FAIL: TestRenderer_LandedAnnotation` — 검출 |

변이 D 의 실제 영향은 임시 테스트 한 개로 확인했다(측정 후 삭제):

```
== original
--- PASS: TestAuditScratch_NoSpawnAfterFailedRefreshOverOldCache (0.07s)
== mutant D
    zz_audit_scratch_test.go:33: spawn attempts after a failed refresh inside TTL = 2, want 0 (stampede guard)
--- FAIL: TestAuditScratch_NoSpawnAfterFailedRefreshOverOldCache (0.08s)
```

감사 종료 시 `git status --short` 는 비어 있다.

### 2.5 판정식 재사용 확인

```
$ grep -rn -e 'LandedAttributions(' -e 'AutoDoneSubjectFresh(' internal | grep -v _test.go
internal/statusline/landed.go:253:	attributed := kanban.LandedAttributions(commits, kanban.LandedBranchFromRef(ref))
internal/statusline/landed.go:256:		if hit, ok := attributed[c.ID]; ok && kanban.AutoDoneSubjectFresh(hit, c.AddedAt) {
internal/cli/todo_autodone.go:217:		attributions = kanban.LandedAttributions(commits, kanban.LandedBranchFromRef(ref))
internal/cli/todo_autodone.go:315:			if hit, ok := attributions[it.ID]; ok && kanban.AutoDoneSubjectFresh(hit, it.AddedAt) {
```

`kanban.ScanLandedSubjects` 는 구분자 없는 줄을 오류로 돌려주므로(`autodone_scan.go:251-258`), 형식이 깨진 출력은 `RefreshLandedCounts` 의 실패 경로로 가고 0 을 쓰지 않는다. `TestRefreshLandedCounts_FailedQueryKeepsThePriorMeasurement/malformed_stream` 이 이를 고정한다.

### 2.6 SPEC 린트

```
$ moai spec lint .moai/specs/SPEC-STATUSLINE-LANDED-LABEL-001
✓ No findings — all SPEC documents are valid
exit=0
```

## 3. Baseline-attribution (기준선 귀속)

- 위 모든 출력은 이 감사가 이 실행에서 `f4ca82858` 트리에 대해 직접 낸 것이다.
- 범위 판정의 왼쪽 끝은 읽는 시점에 다시 구한 `git merge-base develop HEAD` = `b59a5d69c` 다.
- `progress.md` §E.2 의 커버리지 90.7% 는 run 단계가 `7803fa4c0` 에서 잰 값이다. 이 감사의 측정은 90.4% 다(F3).

## 4. Findings (구조화된 결함 목록)

차단 결함은 없다.

- **F1** [Medium] [optional] `internal/statusline/landed.go:223-225` — 옛 기준 캐시를 현재 기준 자리표시로 바꾸는 3행을 지워도 커밋된 테스트가 모두 통과한다(변이 D 생존). 코드는 옳다. 다만 이 3행이 빠지면 옛 캐시 위에서 조회가 계속 실패할 때 렌더마다 갱신 자식이 생성되고(임시 테스트에서 TTL 안 2회 렌더 → 2회 생성), 첫 측정 중에도 같은 일이 생긴다. §C 가 보존을 요구한 스탬피드 가드의 새 연결점인데 고정한 테스트가 없다. 확신도 높음. — 필요한 수정: 옛 기준 캐시 + 실패하는 조회 뒤 `maybeRefreshLandedCounts` 를 두 번 불러 `landedSpawnProbe` 호출이 0 인지 단언하는 테스트 한 개 추가(2.4 의 임시 테스트 본문을 그대로 쓰면 된다).
- **F2** [Low] [optional] `internal/statusline/landed.go:232-240` — 빈 큐 경로에서도 `kanban.LandedRefFor` 가 먼저 호출되므로, 기준 브랜치가 설정되지 않은 프로젝트에서는 읽기 전용 `symbolic-ref` 한 번이 돌 수 있다. REQ-SLL-004 가 ref 해석을 명시적으로 제외하고 기준점 코드도 같은 순서였으므로 요구사항 위반은 아니다. REQ-SLL-005 의 「git 을 조회하지 않고」는 `landedGitRunner` 기준으로만 성립한다. 확신도 높음. — 필요한 수정: 없음. 원하면 `len(picked) == 0` 검사를 ref 해석 앞으로 옮기고 빈 큐 캐시의 `Ref` 를 비워 두면 된다.
- **F3** [Low] [optional] `.moai/reports/t1281/verdict.md:74`, `progress.md:55` — 커버리지 90.7% 는 이 트리의 재측정치 90.4% 와 다르다. 둘 다 85% 이상이라 판정에는 영향이 없지만 수치의 출처(`7803fa4c0`, run 단계)를 밝혀 두는 편이 맞다. 확신도 높음. — 필요한 수정: 다음에 판정서를 고칠 일이 있으면 측정 커밋을 병기.
- **F4** [Low] [optional] `CHANGELOG.md:80` — 「Measured on this tree before the change」 는 레인이 이전 세션에서 잰 값(landed=9, closed=4)이며 이 트리에서 다시 잰 것이 아니다. 판정서 §3 은 출처를 올바르게 레인 측정으로 적었다. 확신도 중간. — 필요한 수정: 「lane measurement before the change」 정도로 표현 조정(선택).

참고(이 카드 범위 밖, 기존 동작): 설정에서 온 `ref` 가 `git log` 인자로 그대로 들어간다. `-` 로 시작하는 값이면 옵션으로 해석될 수 있지만 auto-done 도 같은 경로를 쓰고 기준점 코드도 같았으며 설정 파일은 로컬 신뢰 입력이다. 이 카드의 결함으로 보지 않는다.

## 5. 요구 확인 목록

| 확인 항목 | 결과 | 근거 |
|---|---|---|
| 갱신당 `git log` 주제줄 조회 1회 | 충족 | `TestRefreshLandedCounts_OneInvocationRegardlessOfCardCount` (n=1, 50 모두 1회, argv 가 `LandedScanArgs` 와 같음) |
| 옛 기준 캐시는 아무것도 렌더하지 않음 | 충족 | 변이 C 검출, `TestResolveLandedCounts_OldCriterionIsUnknown` |
| 옛 기준 캐시 위 실패는 모름 유지 | 충족 | `TestRefreshLandedCounts_OldCriterionNeverSurvivesAFailure` (단, F1 의 스탬피드 측면은 미고정) |
| ⚑ 는 U+2691, ✓ 잔재 없음 | 충족 | 2.3 바이트 덤프·grep |
| picked 는 빼지 않음 | 충족 | `TestRenderer_LandedNeverSubtracts`, `renderer.go` 는 `%d/%d` 뒤에 붙이기만 함 |
| 판정서에 기호 후보·이유·유지 결정 | 충족 | `verdict.md` §1.2, §1.3 |
| 4개 로케일 문서가 원어로 자연스러움 | 충족 | ko 「깃발은 완료 표시가 아니라 … 확인하라는 표시」, ja 「旗は完了の印ではなく」, zh 「这面旗不代表完成」 — 직역투 없음. 네 로케일이 같은 세 가지(주제줄만 셈, 확인 후 done, 빼지 않음)를 전함 |
| doctor 점검·훅 생성자 변경 | **없음** | 2.2 — 변경 Go 파일 3개 모두 `internal/statusline` |

## 6. Gaps (미검증)

- 실제 저장소 캐시의 새 기준 수치(`⚑N` 의 N)는 재지 않았다.
- `maybeRefreshLandedCounts` 의 실제 프로세스 생성 경로는 `go test` 에서 막혀 있어 관측하지 않았다(탐침 지점까지만 확인).
- linux/windows 빌드는 재지 않았다. CI 매트릭스가 판정한다.
- `./internal/kanban/...` 는 변경이 없어(2.2) 다시 돌리지 않았다.
- docs-site Hugo 빌드는 돌리지 않았다. 변경은 문단 문구와 코드 주석 한 줄뿐이다.

## 7. Residual-risk (잔여 위험)

- F1 의 3행이 이후 리팩터링에서 사라지면 어떤 테스트도 잡지 못하고, 조회가 계속 실패하는 환경에서 렌더마다 자식 프로세스가 생긴다.
- 재발급 id 의 이전 세대가 새 카드 생성 뒤에 착지하면 한 건 과대 집계된다(SPEC §D 가 수용한 위험).
- 운영자가 `⚑` 를 여전히 「닫아도 됨」으로 읽을 가능성은 문서로만 막는다.
