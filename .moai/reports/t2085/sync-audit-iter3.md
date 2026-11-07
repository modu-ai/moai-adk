auditor-model: claude-opus-5-5

# t2085 독립 감사 판정 (sync-auditor, 3회차 — 운영자 승인 마지막 회차)

verdict: PASS
audited_sha: 2612df0a015d70b6928ba32492a185fddde167ce

- 카드: t2085 (Windows job 127703358711, internal/cli/worktree 최상위 실패 52건)
- 감사 대상 트리: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2085, 브랜치 WT-multi-os-test-3, HEAD 2612df0a01(문서 전용), 그 아래 코드 커밋 24f7b7b019
- 기준 커밋: 400cf31be0. 판정 대상 코드 커밋: fa6d5e3115, dfcc5b9203, 81e10278ef, 24f7b7b019
- 점수: 0.90 (Functionality 90 / Security 95 / Craft 88 / Consistency 88). 필수 통과 차원(Functionality, Security) 모두 통과, 차단 결함 0건.
- 감사 방식: 유료·Codex·GLM 호출 없음(지시대로). 패키지 전체 시험과 `go test ./...` 는 실행하지 않았고, 필터를 건 시험과 overlay 실행만 했다. Windows 에서는 아무것도 실행하지 않았다.

## 반복 이력

- 1회차 (audited dfcc5b9203): FAIL 0.62. 차단 F1(직접 `os.Chmod` 롤백 시험 2건에 Windows 건너뛰기 없음), F2(진행 기록 과대 주장).
- 2회차 (audited 67895b0ef4): FAIL 0.70. 차단 G1(`TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove` 는 표기 문제가 아니라 JSON 원문 substring 단언 문제라서 수정 뒤에도 Windows 에서 계속 실패), 선택 G2(주석·기록의 "없는 키" 문구), G3("nothing else in them runs" 자기모순).
- 2→3회차 사이 변경: 24f7b7b019 — 상태 파일을 JSON 으로 디코드한 뒤 `path == dest` 로 비교(G1), Windows 전용 시험 주석 정정(G2). 2612df0a01 — progress.md 에 E 행 신설, D 를 18건으로 정정, 하위 시험 1개가 계속 실행된다고 명시(G3), 감사 이력 추가.
- 3회차 결과: G1~G3 수리 확인. 새 단언은 macOS 변이 시험 5종에서 모두 실패해 변이를 잡았다. 분할 18+14+1+18+1=52 를 정적 도달 분석과 강제 Windows 분기 실행 양쪽으로 재확인했다. 수리에서 태어난 차단 결함(repair-born)은 없다.

## 판정 요약

2회차의 차단 결함 G1 이 고쳐졌다. CI 로그가 찍은 상태 파일 본문에 새 단언 논리를 그대로 적용하면 참이 나오고(옛 단언은 거짓), 제품이 재키잉에 실패하는 변이 4종과 옛 단언 대조군 1종은 macOS 에서 모두 FAIL 이 났다. 강제 Windows 분기 overlay 로 52건과 `TestMovedBy_UnknownIsProtected` 를 다시 돌린 결과, A+B 32건은 CI 의 실패 지점이 전부 SKIP 되었고 나머지 20건(C 1 + D 18 + E 1)은 건너뛰지 않았다. 건너뛰기 도우미에 닿는 시험 가운데 52건 밖의 것은 `TestMovedBy_UnknownIsProtected` 하나이고, 이미 기록에 공개되어 있다. 제품 수정(FromSlash)은 POSIX 에서 무변화이고, 호출자 중 git 의 슬래시 표기에 기대는 곳은 없다. D 18건의 CI 메시지는 모두 표기 불일치로 설명된다. Windows 실측이 없다는 점은 공개된 Gap 이며, 그 자체로 FAIL 사유는 아니다.

## 결함 목록 (structured defect-list)

차단 결함 없음. 아래는 모두 optional 이다.

- H1 [Low] [optional] [touched-file] .moai/reports/t2085/progress.md:29 — "no iter3 was run, none is claimed" 는 작성 시점에는 참이었지만, 운영자 승인으로 이번 3회차가 실행되었으므로 이제 낡은 문장이다. 과대 주장은 아니다. Required fix(선택): 이 판정 파일을 가리키는 3회차 줄을 감사 이력에 덧붙인다.
- H2 [Low] [optional] [touched-file] .moai/reports/t2085/progress.md:37 — "ok 246.0s" 실행은 1·2회차 수리 이전 트리에서 측정된 것인데 트리 SHA 가 적혀 있지 않다. 이번 감사에서 재현하지 않았다. Required fix(선택): 측정 SHA(dfcc5b9203 시점으로 추정)를 적거나 "수리 이전 트리" 라고 표지.
- H3 [Low] [optional] [touched-file] .moai/reports/t2085/progress.md:16 — D 행 "a real defect pair" 는 제품 쪽 영향을 다소 크게 말한다. 제품 내부 비교는 `worktreeroot.Same`·`underPath`·정규 표기 대체 조회라서 표기 차이를 흡수하고, 실제 제품 쪽 영향은 사용자 출력에서 원본은 `C:/…`, 목적지는 `C:\…` 로 표기가 섞이는 것과 첫 정확 조회가 빗나가는 정도다(E8). 수정 자체는 무해하고 타당하다. Required fix(선택): "출력 표기 불일치 + 시험 도우미 결함" 으로 범위를 좁혀 적는다.
- H4 [Info] [optional] [outside] internal/cli/worktree/windows_skip_t2085_test.go:41 `t2085ReasonSameFile` 상수는 주석에서만 참조된다. golangci-lint(darwin, GOOS=windows 둘 다)는 이를 보고하지 않았다(E9). 조치 불필요.
- (이월, 기록에 공개됨) 1회차 F3: `TestMovedBy_UnknownIsProtected` 의 하위 시험 3개가 Windows 에서 새로 건너뛰어진다. F6: Windows 에는 lsof 가 없어 실제 migrate 가 모든 트리를 건너뛴다. F7: 기준선 lint `unused: migrateOneTree`. 모두 이번 판정 밖이다.

## 1. Claim / Evidence

### C1. G1 수리 — CI 로그 본문에 새 단언을 적용하면 참

이 감사가 새로 작성한 스크립트(iter3/ci.py)로 레인이 받아 둔 CI 로그를 다시 읽었다.
```
$ python3 -I iter3/ci.py t2085-job.log iter3
top-level statuses in the Windows job: {'pass': 301, 'skip': 2, 'fail': 52}
failing top 52 failing sub 15 total 67
G1 failing output lines: ['migrate_hardening_c_t2053_test.go:307: the state entry was not re-keyed to C:\\Users\\runner']
G1 dest: C:\Users\runneradmin\AppData\Local\Temp\TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove3318261980\001\worktrees\projects\mig-repo\legacy6
G1 old assertion strings.Contains(raw,dest): False
G1 new assertion (decode; any path == dest): True | entries: 1
```
CI 에서 이 시험의 실패는 307행 한 줄뿐이었고, 그 앞의 `moved != 2` Fatalf 는 통과했다. 수정 뒤 Windows 에서도 `p.destination` 은 엔진이 만든 OS 표기 경로이고 `destinationFor(filepath.Base(rec.Path))` 의 Base 는 두 구분자를 모두 처리하므로, 재키잉 목적지는 바뀌지 않는다고 추론한다(inferred — 코드 읽기, Windows 미측정).

수정 diff (24f7b7b019, 발췌):
```
-	if !strings.Contains(string(raw), dest) {
+	var entries []struct {
+		Path string `json:"path"`
+	}
+	if jerr := json.Unmarshal(raw, &entries); jerr != nil {
+		t.Fatalf("the state file is not valid JSON: %v\n%s", jerr, raw)
+	}
+	var rekeyed bool
+	for _, e := range entries {
+		rekeyed = rekeyed || e.Path == dest
+	}
+	if !rekeyed {
```

### C2. G1 새 단언의 변이 시험 — 5종 모두 macOS 에서 FAIL

scratchpad 사본을 `go test -overlay` 로 끼워 넣었다(트리 무수정). 제품의 재키잉을 각각 다른 방식으로 망가뜨렸다.
```
$ python3 -I iter3/mut.py <tree> iter3/mut
M0-none: exit=0
    --- PASS: TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove (1.97s)
M1-rekey-never-finds-the-entry: exit=1          (migrate.go: RekeyWorktreeEntry(root, p.source+"-nope", ...) → "no entry" 오류가 삼켜져 재키잉 안 됨)
    migrate_hardening_c_t2053_test.go:320: the state entry was not re-keyed to /private/var/folders/...
    --- FAIL: TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove (1.66s)
M2-rekey-to-a-wrong-destination: exit=1         (목적지 + "-wrong"; 옛 substring 단언이라면 통과했을 변이)
    --- FAIL: TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove (1.93s)
M3-no-canonical-fallback: exit=1                (hook/worktree_registry.go: 정규 표기 대체 조회를 끔 — 이 시험이 지키는 바로 그 동작)
    --- FAIL: TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove (1.77s)
M4-rekey-call-removed: exit=1                   (재키잉 호출을 error(nil) 로 대체)
    --- FAIL: TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove (4.01s)
M5-old-assertion-on-M3: exit=1                  (대조군: 옛 단언 + M3)
    --- FAIL: TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove (2.75s)
```
결론: 새 단언은 공허하지 않다. 옛 단언보다 엄격하고(M2), 재키잉 실패 변이를 모두 잡는다.

### C3. G2·G3 기록 정정 — 확인

- classify_windows_t2085_test.go:10-11 주석은 이제 "the first state-store lookup after a move used the other spelling (a canonical fallback absorbed that one)" 이다. 코드와 대조했다. `hook.rekeyWorktreeEntry`(internal/hook/worktree_registry.go)는 정확 일치에 실패하면 `worktreeroot.CanonicalSpelling` 으로 다시 조회하고, `rebindCardForPath`(migrate.go:1583)도 같은 대체 조회를 한다. CI 관측과도 맞는다. C1 의 디코드 결과 `path == dest`, `TestApply_CardBoundViaSymlinkSpelling_StillFollowsTheMove pass`(E7).
- progress.md:16 D 행: "An earlier draft also claimed the card-binding re-key addressed a missing key; the CI log contradicts it … Only the first lookup used git's spelling." — 사실과 일치.
- progress.md:19: "One subtest outside the failing points (…/the_source_carries_another_identity) still runs" — C5 관측과 일치.

### C4. 분할 18+14+1+18+1 = 52 — 정적 도달 분석과 기록 대조

```
$ python3 -I iter3/reach.py <pkg> iter3/ci-top.txt iter3/ci-status.json
tests reaching a t2085 skip helper: 34 | inside the 52: 33 | outside: 1
  OUTSIDE TestMovedBy_UnknownIsProtected migrate_hardening_e_t2053_test.go CI-Windows: pass
of the 52, NOT reaching a helper: 19
$ python3 -I iter3/split.py <pkg> iter3/ci-top.txt iter3/uncovered.txt progress.md
A only: 18 | B only: 14 | both A and B: [] | C: ['TestJournalWrite_ReplacesTheFileAtomically']
progress D list: 18 | uncovered minus C,E: 18 | equal: True | diff: []
```
A(`t2053Chmod`/`t2053SkipIfRoot` 도달) 18, B(`t2053FakeBin`/`t2053BlockingGit` 도달) 14, 겹침 없음, C 1(`t2085SameFileIsMeaningful`), 도우미에 닿지 않는 19 = D 18 + E 1. progress.md 의 D 목록 18개는 이름까지 정확히 일치하고, 표의 N 열은 A 18 / B 14 / C 1 / E 1 / D 18 이다. 도우미 이름이 호출 외 형태(함숫값 전달 등)로 쓰인 곳은 없다(grep 결과 주석 5줄뿐).

### C5. 강제 Windows 분기 overlay — A+B 32건의 CI 실패 지점 전부 SKIP

방법: 레인 스크립트를 쓰지 않고 iter3/forced.py 로 다시 작성했다. `windows_skip_t2085_test.go` 의 `runtime.GOOS == "windows"` → `!= ""`, `!= "windows"` → `== ""` 로만 바꾼 사본을 overlay 로 끼웠다. 52건 + `TestMovedBy_UnknownIsProtected` 를 두 묶음으로 `-json` 실행하고, CI 에서 실패한 각 말단 단위(시험 또는 하위 시험)가 자신이나 조상 단위에서 skip 되었는지 판정했다.
```
$ python3 -I iter3/forced.py <tree> iter3
replaced ==windows: 1 !=windows: 1
half 1: 26 tests, exit=0, tail=...,"Output":"ok  \tgithub.com/modu-ai/moai-adk/internal/cli/worktree\t38.689s\n"}
half 2: 27 tests, exit=0, tail=...,"Output":"ok  \tgithub.com/modu-ai/moai-adk/internal/cli/worktree\t51.585s\n"}
$ python3 -I iter3/analyze.py iter3
forced-run FAIL events: []
covered (every CI failing point skipped): 32 | top-level SKIP: 28 | top-level PASS with failing subtests skipped: 4
    TestMigrateVerb_WorktreeListTimeoutFails pass still-run subtests: []
    TestMigrationEngine_WorktreeListTimeoutIsAnError pass still-run subtests: []
    TestPlan_PendingIntent_NotProvablyUnmoved_IsInterrupted pass still-run subtests: ['TestPlan_PendingIntent_NotProvablyUnmoved_IsInterrupted/the_source_carries_another_identity']
    TestUpdateConsolidation_WorktreeListTimeoutFails pass still-run subtests: []
not covered: 20
    pass TestApply_PendingIntentButNeverMoved_IsSettledThenMoved
    pass TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove
    pass TestJournalWrite_ReplacesTheFileAtomically
    (… D 17건 더, 모두 pass …)
TestMovedBy_UnknownIsProtected (outside the 52): {control 2개: pass, token_cannot_be_read/move_list_cannot_be_read/git_cannot_list_the_worktrees: skip, 최상위: pass} | CI: {6개 모두 pass}
```
52건 밖에서 바뀐 것은 이미 공개된 `TestMovedBy_UnknownIsProtected` 하위 시험 3개뿐이다(정적 분석 C4 의 OUTSIDE 1건과 일치).

### C6. 직접 `os.Chmod` / `os.Geteuid` 사용 전수 — Windows 실패 시험 중 건너뛰기 없는 것 0건

```
$ grep -n "os\.Chmod\|os\.Geteuid\|syscall\.Chmod\|Fchmod" *_test.go | grep -v "t.Cleanup\|_ = os.Chmod"   (함수별로 묶음)
migrate_bounded_test.go:101,105,116,122  in TestCaptureManifest_IgnoredUnreadableFile_DeletionAndSizeChangeDetected   CI-Windows: pass (52건 아님)
migrate_hardening_c_t2053_test.go:172    in t2053ReadOnlyMoveList (사용처 TestApply_/TestUpdateConsolidation_ReportRowsPrecedeTheError)  CI-Windows: pass, pass
migrate_hardening_d_t2053_test.go:133    in TestRollback_UnstatableDestination_IsNotTreatedAsAbsent   CI: fail → 128행 t2053SkipIfRoot(t) 가 먼저 실행
migrate_hardening_d_t2053_test.go:156    in TestRollback_UnstatableSource_IsNotTreatedAsGone          CI: fail → 151행 t2053SkipIfRoot(t) 가 먼저 실행
migrate_hardening_e_t2053_test.go:24,33  in t2053SkipIfRoot / t2053Chmod (도우미 본문, 23·32행에서 t2085SkipOnWindows 선행)
movelist_test.go:333,339                 in TestEnsureInstanceToken_FailsWhenTheAdminDirectoryCannotBeWritten   CI: skip (기존 runtime.GOOS 가드)
```
CI 의 Windows job 에서 실패했으면서 Windows 건너뛰기 없이 직접 chmod/geteuid 를 쓰는 시험은 없다.

### C7. FromSlash/ToSlash 무변화, 호출자, Windows vet

```
$ go version
go version go1.26.8 darwin/arm64
$ grep -n -A3 "^func FromSlash\|^func ToSlash" $(go env GOROOT)/src/internal/filepathlite/path.go
176:func ToSlash(path string) string {
177-	if Separator == '/' {
178-		return path
184:func FromSlash(path string) string {
185-	if Separator == '/' {
186-		return path
$ grep -rn 'ParseWorktreeList(' internal cmd | grep -v '_test.go'
internal/cli/launch_view.go:273
internal/cli/launcher.go:519
internal/cli/worktree/movelist.go:252
internal/cli/worktree/list.go:97
internal/cli/worktree/migrate.go:464
internal/cli/worktree/migrate.go:1540
```
호출자 검토(코드 읽기): launch_view.go:282 와 launcher.go:526 은 이미 `filepath.FromSlash(rec.Path)` 를 하므로 멱등이다. movelist.go 는 `isRegistered` → `worktreeroot.Same`. list.go:129 는 `filepath.Clean` 후 `worktreeroot.Same`. classify.go 146-174 는 `worktreeroot.Same`/`underPath`. migrate.go:452 는 `ToSlash` 후 범위 판정, 575 는 `Same`, 553 의 `findInterrupted` 는 `worktreeroot.Same(rec.OldPath, p)`, 580/589/599 는 `resolveTreePath`/`underPath`/정규 표기 비교. git 의 슬래시 표기에 기대는 문자열 비교는 없다.
```
$ env GOOS=windows go vet ./internal/cli/worktree/ ./internal/cli/; echo "winvet-exit=$?"
winvet-exit=0
$ go vet ./internal/cli/worktree/; echo "vet-exit=$?"
vet-exit=0
$ gofmt -l $(git diff --name-only 400cf31be0 HEAD -- '*.go'); echo "gofmt-exit=$?"
gofmt-exit=0      (출력 없음)
```

### C8. 지정 필터 시험 — 통과, 최상위 12건 훑음

```
$ go test -count=1 -timeout 280s -run 'ParseWorktreeList|Classify|TestJournalWrite_ReplacesTheFileAtomically|TestRollback_Unstatable|TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove' ./internal/cli/worktree/
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	26.874s
exit=0
$ (같은 명령에 -v, 결과 줄만)
--- PASS: TestClassify_MeasuredShapes (0.53s)
--- PASS: TestClassify_NeverMovedClassesAreNotMoaiMade (0.03s)
--- PASS: TestClassify_SpellingSafeRootsAndPrimary (0.02s)
    classify_test.go:231: volume is case-insensitive; the case-aliased test above covers it
--- SKIP: TestClassify_CaseDifferingDirectoriesStayDistinctOnACaseSensitiveVolume (0.00s)
--- PASS: TestParseWorktreeList_Porcelain (0.00s)
--- PASS: TestParseWorktreeList_RealGit (2.46s)
--- PASS: TestClassify_ListingEndToEnd (0.21s)
--- PASS: TestJournalWrite_ReplacesTheFileAtomically (0.19s)
--- PASS: TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove (6.29s)
--- PASS: TestRollback_UnstatableDestination_IsNotTreatedAsAbsent (6.74s)
--- PASS: TestRollback_UnstatableSource_IsNotTreatedAsGone (5.31s)
--- PASS: TestClassify_AgentNamedTreeFollowsItsRoot (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	22.465s
v-exit=0
```
훑은 최상위 시험은 12건(PASS 11, SKIP 1)이다. SKIP 1건은 기준선 시험이고 사유는 이번 카드와 무관하다(대소문자 무시 볼륨). macOS 에서 t2085 건너뛰기는 하나도 발동하지 않았다. Windows 전용 `TestParseWorktreeList_WindowsPathUsesTheOSSeparator` 는 `//go:build windows` 라서 여기서는 컴파일되지 않는다(GOOS=windows vet 은 통과).

### C9. D 18건 추론 검토 — CI 메시지 전수가 표기 불일치로 설명됨

iter3/dmsgs.py, iter3/full.py 로 20건의 CI 실패 메시지를 다시 뽑았다. D 18건의 실패 단언은 네 종류다. `m6Registered` 계열("not registered at its source", "registration does not name the source only", "was not moved to the home root", "protected tree … was moved or removed"), 계획 행 `p.source` 문자열 비교(`switch p.source { case w.movedTree: … }` → "occupied=false free=false"), 사람용 출력에서 OS 표기 경로를 substring 검색하는 경우("the plan does not list the integration tree", "not skipped with a registry reason"), 그리고 "no row for C:\…" 이다. 의심이 갔던 세 건은 출력 원문으로 반증했다.
```
TestMigrate_ApplyMovesEligibleAndSkipsPr |         moai-made	C:/Users/runneradmin/.../002/mig-repo/.moai/worktrees/legacy6 → C:\Users\runneradmin\...\001\worktrees\projects\mig-repo\legacy6	[moved]
TestMigrate_ApplyMovesEligibleAndSkipsPr |         moved 4, skipped 2, failed 0, reversed 0
TestMigrate_ApplyMovesEligibleAndSkipsPr |     migrate_command_test.go:234: legacy6 was not moved to the home root
TestMigrateNamesIntegrationClass |         integration	C:/Users/runneradmin/.../002/mig-repo/.claude/worktrees/main → ...	[skipped] (class integration is never m
TestMigrateNamesIntegrationClass |     migrate_review_fix_test.go:404: the plan does not list the integration tree:
TestMigrationSkipsOnUnreadableRegistry |         moai-made	C:/Users/runneradmin/.../002/mig-repo/.moai/worktrees/legacy6 → ...	[skipped] (session regist
```
제품은 legacy6 를 실제로 옮겼고(`[moved]`), 통합 트리를 `integration` 으로 분류했으며, 레지스트리 사유로 건너뛰었다. 실패한 것은 출력의 원본 경로가 `C:/…` 인데 시험이 `C:\…` 로 찾았기 때문이다. lsof 부재(F6)는 원인이 아니다. `m6NewWorld` 가 `activeProcessCWDProbe` 를 스텁한다(migrate_command_test.go:87). 수정 뒤에는 원본이 FromSlash 로 `C:\…` 가 되고, `m6Registered` 는 ToSlash 로 git 표기와 맞춘다. 이 추론은 타당하다고 판단하지만, Windows 에서 녹색이 되는지는 미측정이다(Gaps).

### C10. lint

```
$ golangci-lint run ./internal/cli/worktree/; echo lint-exit=$?
internal/cli/worktree/migrate.go:1051:6: func migrateOneTree is unused (unused)
1 issues:
* unused: 1
lint-exit=1
$ GOOS=windows golangci-lint run ./internal/cli/worktree/   (같은 1건, winlint-exit=1)
```
유일한 보고는 기준선 결함(1회차 F7)이다. migrate.go 는 이번 diff 에 없다(`git diff --stat 400cf31be0 HEAD` 에 미포함). 새 lint 보고는 없다.

### C11. 숨겨진 제품 결함 탐색

A군 chmod 경로, B군 shim 경로, C 의 SameFile 은 OS 중립 제품 코드의 오류 처리 경로를 시험한다. 건너뛰기가 가리는 Windows 전용 제품 분기는 찾지 못했다. FromSlash 이후 바뀌는 제품 동작은 Windows 의 사용자 출력 표기(`moai worktree list`, migrate 계획 행의 원본 경로가 백슬래시로 바뀜)와, 상태 파일·카드 바인딩 첫 정확 조회의 적중이다. 둘 다 개선 방향이다. 이름을 댈 수 있는 숨겨진 제품 결함은 없다.

## 2. Baseline-attribution

- 모든 명령은 2026-11-06 이 감사 실행 중에, 트리 HEAD 2612df0a015d70b6928ba32492a185fddde167ce 에서 실행했다(시작 시와 이 파일 작성 직전 모두 `git status --short` 비어 있음, HEAD 불변). go1.26.8 darwin/arm64 로 트리에서 직접 컴파일했고, 설치된 moai 바이너리는 쓰지 않았다(§2.2 해당 없음). golangci-lint v2.1.6(go1.26.8 빌드).
- CI 근거는 레인이 받아 둔 로그 사본 `t2085-job.log`(job 127703358711, Windows)다. 이 감사가 새로 작성한 스크립트 iter3/ci.py, dmsgs.py, full.py 로 다시 읽었고, 로그를 다시 내려받지는 않았다. CI 가 돌린 트리의 G1 단언 위치(c:306-307)는 기준 400cf31be0 과 일치한다.
- 강제 분기 overlay 와 변이 시험은 레인의 t2085_overlay.py 를 재사용하지 않고 iter3/forced.py, iter3/mut.py 로 다시 만들었다.

## 3. Gaps (관측하지 않은 것)

- Windows 실행 없음. 건너뛰기 분기, D 18건 수정, E 1건 수정, Windows 전용 시험 모두 Windows 에서 관측되지 않았다. C1 의 "새 단언은 Windows 에서 참"은 CI 가 찍은 파일 본문에 같은 논리를 적용한 결과이고, 수정된 바이너리를 Windows 에서 돌린 결과가 아니다. 첫 측정은 리더 소관의 Windows CI 다.
- D 18건이 Windows 에서 녹색이 되는지는 미측정이다(기록에도 공개됨). macOS 에서는 강제 분기 overlay 안에서 pass 였다(C5).
- 패키지 전체 시험, `go test ./...`, internal/cli 의 호출자 시험(launch_view, launcher)은 지시에 따라 실행하지 않았다. 호출자는 코드 읽기와 GOOS=windows vet 으로만 확인했다.
- 교차 모델 감사(audit_multi/codex/glm), /verify(moai gate), 커버리지, card-review 는 지시에 따라 실행하지 않았다. 이 트리의 `workflow.audit.gates.codex` 는 `required` 인데 이번 감사에는 Codex receipt 가 없다. 그래서 receipt guard 가 이 PASS 를 거부할 수 있다. 우회하지 않았고, 최종 메시지는 `receipts=none` 으로 남긴다. 필요한 Codex 다리를 채울지는 리더가 정한다.
- 레인의 246.0s 실행(progress.md:37)과 8.8s 실행(progress.md:33)은 재현하지 않았다. 대신 같은 시험을 C8 에서 직접 실행했다.
- 도구 거부: 이번 회차에는 거부된 명령이 없다.
- receipt guard 거부(실측): 첫 최종 메시지(`AUDIT-VERDICT: PASS spec=t2085 receipts=none`) 뒤에 Stop hook 이 `AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS but the audit could not be corroborated — no receipt cited.` 를 반환하며 codex_audit 또는 audit_multi 호출을 요구했다. 리더 지시(Codex·유료 호출 금지, guard 거부 시 우회 없이 Gap 으로 기록)에 따라 호출하지 않았다. 판정 문구도 바꾸지 않았다. 이 PASS 는 receipt 로 뒷받침되지 않은 Claude 단독 판정이고, Codex 다리를 채울지는 리더가 정한다.

## 4. Residual-risk (관측했음에도 남는 위험)

- 강제 분기 overlay 는 `runtime.GOOS` 조건만 바꾼다. Windows 의 실제 파일시스템·PATH·git 동작은 재현하지 않는다. 다만 건너뛰기 조건은 `runtime.GOOS == "windows"` 하나뿐이라 위험은 낮다.
- D 군 뒤에 Windows 차이가 더 숨어 있을 수 있다(예: 8.3 짧은 이름, 대소문자, 사람용 출력의 다른 이스케이프). G1 이 그런 실례였다. JSON 원문에서 경로를 substring 으로 찾던 곳은 c:306 하나였고(2회차 grep), 이제 사라졌다.
- Windows 사용자 출력의 경로 표기가 `C:/…` 에서 `C:\…` 로 바뀐다. 이 출력을 문자열로 파싱하는 외부 소비자가 있다면 영향을 받는다. 저장소 안의 제품 호출자는 C7 에서 모두 확인했다.
- 스크래치 증거(iter3/ 아래 ci-status.json, ci-top.txt, uncovered.txt, forced1/2.json, mut/, lint*.txt 와 스크립트)는 세션 scratchpad 에만 있다. 판정을 가른 출력 줄은 이 파일로 옮겼고, 나머지 원문은 세션과 함께 사라진다.

## Recommendations

1. 착지 전 확인: 리더 소관의 Windows CI 실행으로 52건(특히 D 18 + E 1)의 실제 결과를 측정한다.
2. H1~H3(선택): progress.md 의 감사 이력·측정 SHA·D 행 문구를 다듬는다. 판정에는 영향이 없다.

## Operational Notes (unverified)

- Windows CI 에서 확인할 명령: `go test -count=1 -run '<D 18건 + TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove + TestParseWorktreeList_WindowsPathUsesTheOSSeparator>' ./internal/cli/worktree/`. 상태: assumption(미실행).
- 건너뛰기 범위 재확인은 이 파일 C5 의 overlay 방식으로 macOS 에서 측정할 수 있다. 상태: measured(이번 감사에서 실행, C5).
