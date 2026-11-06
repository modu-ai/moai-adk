auditor-model: claude-opus-5-5

# t2085 독립 감사 판정 (sync-auditor, 2회차 — 마지막 허용 회차)

verdict: FAIL
audited_sha: 67895b0ef44bc8097268f8b1e0d951c7ffa49b45

- 카드: t2085 (Windows job 127703358711, internal/cli/worktree 최상위 실패 52건)
- 감사 대상 트리: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2085, 브랜치 WT-multi-os-test-3, HEAD 67895b0ef4(문서 전용) 아래 코드 커밋 81e10278ef
- 기준 커밋: 400cf31be0. 판정 대상 코드 커밋: fa6d5e3115, dfcc5b9203, 81e10278ef(1회차 수리)
- 점수: 0.70 (Functionality 55 / Security 95 / Craft 80 / Consistency 85; 가중 평균은 참고값이고, 판정은 차단 결함 G1과 Functionality 필수 통과 실패가 결정)
- 감사 방식: 유료·Codex·GLM 호출 없음(지시대로). 패키지 전체 시험과 go test ./... 는 실행하지 않음. 이 감사는 Windows에서 아무것도 실행하지 않았음.

## 반복 이력

- 1회차 (audited dfcc5b9203): FAIL 0.62. 차단 F1(직접 os.Chmod 를 쓰는 롤백 시험 2건이 Windows 건너뛰기를 받지 않음), F2(진행 기록 과대 주장), 선택 F3~F7.
- 1→2회차 사이 변경: 81e10278ef — 두 롤백 시험이 `t2053SkipIfRoot` 호출(F1), TestJournalWrite 는 os.SameFile 단언만 OS 조건화(F4). 67895b0ef4 — progress.md 정정(F2/F3/F5/F6 공개).
- 2회차 결과: 1회차 결함은 모두 고쳐졌음을 확인. 그러나 1회차가 "D군 분류 타당"으로 넘긴 19건 중 1건이 이번 변경 뒤에도 Windows 에서 계속 실패한다는 것을 CI 로그 자체로 확인(G1). 1회차 감사가 놓친 결함이며, 수리에서 태어난 결함(repair-born)이 아니다(81e10278ef 의 diff 밖, 처음부터 있던 오분류).

## 판정 요약

1회차 결함 F1~F5 는 수리되었다. 강제 Windows 분기 overlay 로 52건을 모두 다시 돌린 결과, A+B 32건은 CI 의 실패 지점 전부가 SKIP 되고 나머지 20건(C 1 + D 19)은 건너뛰지 않으며, 분할 18+14+1+19 가 기계적으로 재확인된다. 제품 수정(FromSlash)은 POSIX 에서 무변화이고 호출자 중 슬래시 표기에 기대는 곳도 없다.

그러나 D군의 `TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove` 는 표기 불일치로 실패한 것이 아니다. CI 로그에 찍힌 상태 파일 내용을 디코드하면 제품은 Windows 에서 이미 목적지로 올바르게 재키잉했고, 시험은 JSON 원문 바이트(`C:\\Users\\…`, 백슬래시가 이스케이프됨)에서 이스케이프되지 않은 경로(`C:\Users\…`)를 `strings.Contains` 로 찾다가 실패했다. 이번 변경은 이 단언을 건드리지 않으므로 이 시험은 Windows 에서 계속 실패한다. 진행 기록은 이 시험을 D군 "fix (below)"에 넣어 표기 수정으로 해결된다고 기록했다. 카드 목표(52건 처분) 미달 + 기록의 원인 오분류 → FAIL.

## 결함 목록 (structured defect-list)

- G1 [High] [blocking] [outside — 재현 증거 있음: CI 로그] internal/cli/worktree/migrate_hardening_c_t2053_test.go:305-307 — `raw, _ := os.ReadFile(stateFile); if !strings.Contains(string(raw), dest)`. 상태 파일은 JSON(MarshalIndent)이므로 Windows 경로의 백슬래시는 `\\` 로 저장된다. `dest` 는 단일 백슬래시이므로 Windows 에서 이 검사는 경로가 맞아도 항상 거짓이다. CI 출력의 파일 본문을 디코드하면 `path` 필드 == `dest` (아래 E4). 즉 제품 결함이 아니라 OS 를 가정한 시험 단언 결함이고, D군의 FromSlash 수정과 무관하며, 이번 카드 이후에도 Windows 에서 빨간색으로 남는다. — Required fix: 단언을 JSON 디코드 후 비교로 바꾼다(예: `var entries []struct{ Path string \`json:"path"\` }` 로 `json.Unmarshal(raw, &entries)` 후 어느 항목의 `Path` 가 `dest` 와 같은지(또는 `worktreeroot.Same`) 확인). 테스트 전용, 별도 커밋. 그리고 progress.md 원인 표에서 이 시험을 D군(표기)에서 빼서 "시험 단언 결함(JSON 이스케이프)" 행으로 옮기고 D 를 18건으로 정정한다. 수리 후 macOS 에서 이 시험을 실행해 PASS 를 확인(POSIX 에서는 이스케이프할 문자가 없으므로 동작 동일).
- G2 [Low] [optional — 기록 정직성] internal/cli/worktree/classify_windows_t2085_test.go:10-11 주석과 progress.md:15 D행 Evidence 칸 — "the card-binding re-key after a move addressed a key that did not exist" / "`rebindCardForPath(root, p.source, ...)` addressed a key in git's spelling". 첫 조회 키가 git 표기였던 것은 맞지만, `rebindCardForPath` 는 실패 시 `worktreeroot.CanonicalSpelling`(내부 `filepath.Clean`, Windows 에서 `/`→`\`)으로 대체 조회하고, CI 에서 `TestApply_CardBoundViaSymlinkSpelling_StillFollowsTheMove` 는 수정 전 Windows 에서 pass 였다(E5). 결함 근거로 제시된 문장이 CI 관측과 맞지 않는다. — Required fix(선택): 주석과 표 칸을 "첫 조회 키가 git 표기였다(대체 조회가 흡수함, CI pass)"로 사실대로 고치거나 해당 문구를 삭제.
- G3 [Low] [optional — 기록 표현] progress.md:18 — "every failing point of the 32 skips, nothing else in them runs". 같은 파일 29행과 이 감사의 실행 모두에서 `TestPlan_PendingIntent_NotProvablyUnmoved_IsInterrupted/the_source_carries_another_identity` 는 계속 실행된다. "nothing else in them runs" 는 문자 그대로는 거짓(29행과 자기모순). — Required fix(선택): "실패 지점 밖의 하위 시험 1개는 계속 실행된다"로 정정.
- (참고, 1회차 F6/F7 그대로) Windows 의 lsof 부재로 migrate 가 모든 트리를 건너뛰는 기존 동작, `unused: migrateOneTree` 기준선 lint — 이번 diff 와 무관, 기록에 공개됨.

## 1. Claim / Evidence

### C1. 1회차 F1 수리 — 확인

두 롤백 시험이 `t2053SkipIfRoot` 를 호출하고, 이 도우미가 `t2085SkipOnWindows(t, t2085ReasonPosixMode)` 를 먼저 부른다.
```
$ git diff dfcc5b9203 81e10278ef -- internal/   (발췌)
-	if os.Geteuid() == 0 {
-		t.Skip("root ignores directory modes")
-	}
+	t2053SkipIfRoot(t) // root and Windows both ignore the directory mode this test relies on
   (TestRollback_UnstatableDestination_IsNotTreatedAsAbsent, TestRollback_UnstatableSource_IsNotTreatedAsGone 두 곳)
$ grep -n -A6 "^func t2053SkipIfRoot" migrate_hardening_e_t2053_test.go
21:func t2053SkipIfRoot(t *testing.T) {
22-	t.Helper()
23-	t2085SkipOnWindows(t, t2085ReasonPosixMode)
24-	if os.Geteuid() == 0 {
```

남은 직접 `os.Chmod` / `os.Geteuid` 사용 전수(정리용 `t.Cleanup` 줄 제외):
```
$ grep -n "os\.Chmod\|os\.Geteuid\|#!/bin/sh\|#!/usr/bin/env" *_test.go | grep -v "t.Cleanup"
migrate_bounded_test.go:55:	t2053FakeBin(t, "git", fmt.Sprintf(`#!/bin/sh
migrate_bounded_test.go:101:	if os.Geteuid() == 0 {
migrate_bounded_test.go:105:	if err := os.Chmod(ignored, 0o000); err != nil {
migrate_bounded_test.go:116/122: (같은 시험)
migrate_bounded_test.go:178/202:	t2053FakeBin(t, "lsof", "#!/bin/sh\nexec sleep 30\n")
migrate_hardening_e_t2053_test.go:24/33: (도우미 본문)
migrate_hardening_e_t2053_test.go:352/433/554/583: `_ = os.Chmod(..., 0o644/0o755)` 복원
migrate_hardening_c_t2053_test.go:172/177/182: t2053ReadOnlyMoveList
migrate_hardening_c_t2053_test.go:222:	t2053FakeBin(t, "git", ...
movelist_test.go:333:	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
migrate_hardening_d_t2053_test.go:133/156: (위 두 시험, 이제 t2053SkipIfRoot 뒤)
```
도우미를 거치지 않는 chmod 사용 시험은 `TestCaptureManifest_IgnoredUnreadableFile_DeletionAndSizeChangeDetected`(bounded:100), `TestApply_ReportRowsPrecedeTheError`(t2053ReadOnlyMoveList 경유), `TestEnsureInstanceToken_FailsWhenTheAdminDirectoryCannotBeWritten`(이미 Windows skip). 셋 모두 52건에 없고 CI Windows 결과는:
```
TestApply_ReportRowsPrecedeTheError pass
TestCaptureManifest_IgnoredUnreadableFile_DeletionAndSizeChangeDetected pass
TestEnsureInstanceToken_FailsWhenTheAdminDirectoryCannotBeWritten skip
```
결론: 52건 중 chmod 또는 `#!/bin/sh` shim 에 기대면서 Windows 건너뛰기가 없는 시험은 남아 있지 않다.

### C2. "A(18)+B(14) 32건은 Windows 분기를 강제하면 실패 지점이 모두 건너뛰어진다" — 독립 재도출로 확인

방법: 레인 스크립트를 쓰지 않고 새로 작성. CI 로그에서 실패 목록을 직접 다시 뽑고(E1), `windows_skip_t2085_test.go` 의 `runtime.GOOS == "windows"` → `runtime.GOOS != ""`, `runtime.GOOS != "windows"` → `runtime.GOOS == ""` 로만 바꾼 사본을 `go test -overlay` 로 끼워(트리 무수정) 52건 전부 + `TestMovedBy_UnknownIsProtected` 를 두 묶음(26+26+1)으로 `-json` 실행. CI 에서 실패한 각 (시험/하위 시험)이 자신 또는 조상 단위에서 skip 되었는지 판정.

E1 (CI 로그 재파싱):
```
$ python3 -I iter2/parse_ci.py t2085-job.log ci-top.txt ci-sub.txt
top 52 sub 15 total 67
```
E2 (overlay 생성 + 실행):
```
$ python3 -I iter2/mk_overlay.py <tree> iter2
replaced ==windows: 1 !=windows: 1
halves 26 26
$ go test -count=1 -timeout 280s -overlay=iter2/overlay.json -json -run "<26개>" ./internal/cli/worktree/   (두 번)
exit=0   ... "Output":"ok  \tgithub.com/modu-ai/moai-adk/internal/cli/worktree\t93.124s\n"
exit=0   ... "Output":"ok  \tgithub.com/modu-ai/moai-adk/internal/cli/worktree\t88.015s\n"
```
E3 (판정):
```
$ python3 -I iter2/analyze.py iter2
forced-run FAIL events: 0 []
covered (all CI failing points skipped): 32
   TestMigrateVerb_WorktreeListTimeoutFails -> failing-subtests-skipped; still-run subtests=[]
   TestMigrationEngine_WorktreeListTimeoutIsAnError -> failing-subtests-skipped; still-run subtests=[]
   TestPlan_PendingIntent_NotProvablyUnmoved_IsInterrupted -> failing-subtests-skipped; still-run subtests=['TestPlan_PendingIntent_NotProvablyUnmoved_IsInterrupted/the_source_carries_another_identity']
   TestUpdateConsolidation_WorktreeListTimeoutFails -> failing-subtests-skipped; still-run subtests=[]
NOT covered: 20
   pass TestApply_PendingIntentButNeverMoved_IsSettledThenMoved
   pass TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove
   pass TestJournalWrite_ReplacesTheFileAtomically
   pass TestMigrateApply_ForwardMoveInterruptedAfterRename_ConvergesBack
   pass TestMigrateApply_ForwardMoveNonTimeoutErrorAfterRename_Reverses
   pass TestMigrateNamesIntegrationClass
   pass TestMigrateReverseMoveFailure_StaysAtDestination
   pass TestMigrate_ApplyMovesEligibleAndSkipsProtected
   pass TestMigrate_RollbackRestoresTheTrees
   pass TestMigrationSkipsOnUnreadableRegistry
   pass TestMovedBy_ProtectsAFailedReversePendingTree
   pass TestPlanMigration_SkipsATreeAnActiveProcessIsInside
   pass TestPlan_InterruptedMigration_NamesTheJournal
   pass TestReverseMove_ErrorAfterMoveCompleted_IsTreatedAsMovedBack
   pass TestReverseMove_StaleRegistration_RepairConverges
   pass TestRollback_EmptiedRegistrationFile_RepairsAndRestores
   pass TestRollback_ForwardMoveInterrupted_RepairsThenRestores
   pass TestRollback_RenamedBackStaleRegistration_RepairsAndRestores
   pass TestRollback_StepSettledAsNeverMoved_IsANoOp
   pass TestUpdateConsolidation_MovesToHomeRootAndSkipsAgent
TestMovedBy_UnknownIsProtected subtests: {... 'token_cannot_be_read': 'skip', 'move_list_cannot_be_read': 'skip', 'git_cannot_list_the_worktrees': 'skip', 'control:_a_moved,_registered_tree_is_protected': 'pass', 'control:_a_path_that_is_not_a_linked_worktree_is_not_protected': 'pass'}
```
해석: 덮인 32건 = 최상위 SKIP 28 + 실패 하위 시험만 SKIP 된 최상위 PASS 4. 진행 기록 29행의 "28 top-level SKIP, 4 top-level PASS"와 일치. 덮이지 않은 20건 = 기록의 D 19건 목록과 이름이 정확히 같고 + C 1건. 따라서 분할 18+14+1+19=52 가 강제 분기 실행으로 기계적으로 재확인된다. F3 의 하위 시험 3개 신규 skip 도 관측과 일치.

### C3. 1회차 F4 수리 — 확인

`TestJournalWrite_ReplacesTheFileAtomically` 에는 함수 첫 줄 skip 이 없고, 64행이 `if t2085SameFileIsMeaningful() && os.SameFile(before, after)`, 68행 JSON/`pending` 검사는 무조건 실행. 강제 Windows 분기(E3)에서 이 시험은 skip 이 아니라 `pass` — SameFile 단언만 빠지고 내용 검사는 돌았다. 전제 확인: Go 소스 `os/types_windows.go:287 func (fs *fileStat) loadFileId()` 가 `fs.path` 로 지연 오픈 → 같은 경로의 두 FileInfo 는 교체 뒤 항상 같은 파일로 읽힌다(사유 문구와 일치).

### C4. 1회차 F2/F3/F5 기록 정정 — 확인 (단, G3 신규)

- progress.md:6 "67 incl. 15 failing subtests" = E1 의 `top 52 sub 15 total 67`. (F5 해결)
- progress.md:12 A행 "16 call t2053Chmod / t2053SkipIfRoot; the other 2 … called os.Chmod directly and were NOT covered by the first commit (audit F1), now call t2053SkipIfRoot" — C1 증거와 일치. (F2a 해결)
- progress.md:28 표본 실행 항목에 "the last one was not touched by the first commit, so that PASS proved nothing for it" 명시. (F2c 해결) 해당 표본 실행 자체는 이 감사가 재현하지 않음(이력 기록).
- progress.md:37 TestMovedBy_UnknownIsProtected 하위 3개 신규 skip 공개, "probably vacuously" 로 추정 표지. (F3 해결)
- progress.md:38 lsof 메모, "inferred" 표지. (F6 공개)
- progress.md:18 "nothing else in them runs" — G3.

### C5. 제품 수정과 호출자 — 1회차 확인 사항 유지

```
$ go version
go version go1.26.8 darwin/arm64
$ grep -n -A4 "^func FromSlash\|^func ToSlash" $(go env GOROOT)/src/internal/filepathlite/path.go
176:func ToSlash(path string) string {
177-	if Separator == '/' {
178-		return path
184:func FromSlash(path string) string {
185-	if Separator == '/' {
186-		return path
```
호출자(`grep -rn "ParseWorktreeList(" internal cmd`, 제품 코드): launch_view.go:273(283행 이미 `filepath.FromSlash`, 멱등), launcher.go:519(526행 이미 FromSlash, 멱등), worktree/list.go:97, migrate.go:464(`inMigrationScope` 는 ToSlash 후 비교), migrate.go:1540, movelist.go:252. 1회차 목록과 동일, 새 호출자 없음. 제품 코드의 `rec.Path` 비교는 `worktreeroot.Same`/`underPath`(evalOrClean)/`CanonicalSpelling`/`os.Lstat` 등 경로 인식 비교뿐 — git 의 슬래시 표기에 기대는 곳 없음.
```
$ env GOOS=windows go vet ./internal/cli/worktree/ ./internal/cli/
winvet-exit=0
$ go vet ./internal/cli/worktree/
vet-exit=0
$ gofmt -l <변경된 8개 파일>
gofmt-exit=0   (출력 없음)
```

### C6. 지정 필터 시험 — 통과, 11건 훑음, 두 Unstatable 시험은 실행되어 PASS

```
$ go test -count=1 -timeout 280s -run 'ParseWorktreeList|Classify|TestJournalWrite_ReplacesTheFileAtomically|TestRollback_Unstatable' ./internal/cli/worktree/
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	9.286s
exit=0
$ (같은 명령에 -v, 결과 줄만)
--- PASS: TestClassify_MeasuredShapes (0.03s)
--- PASS: TestClassify_NeverMovedClassesAreNotMoaiMade (0.02s)
--- PASS: TestClassify_SpellingSafeRootsAndPrimary (0.00s)
--- SKIP: TestClassify_CaseDifferingDirectoriesStayDistinctOnACaseSensitiveVolume (0.00s)
--- PASS: TestParseWorktreeList_Porcelain (0.00s)
--- PASS: TestParseWorktreeList_RealGit (0.65s)
--- PASS: TestClassify_ListingEndToEnd (0.02s)
--- PASS: TestJournalWrite_ReplacesTheFileAtomically (0.03s)
--- PASS: TestRollback_UnstatableDestination_IsNotTreatedAsAbsent (2.16s)
--- PASS: TestRollback_UnstatableSource_IsNotTreatedAsGone (1.90s)
--- PASS: TestClassify_AgentNamedTreeFollowsItsRoot (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	5.487s
```
훑은 최상위 시험 11개(PASS 10, SKIP 1). SKIP 1건은 기준선 시험이며 사유는 이번 카드와 무관: `classify_test.go:231: volume is case-insensitive; the case-aliased test above covers it`. macOS 에서 t2085 건너뛰기는 하나도 발동하지 않음.

### C7. D군 추론 검토 — 18건은 타당, 1건은 반증 (G1)

타당한 쪽: 실패 단언이 `m6Registered`(git porcelain 원문에 OS 표기 경로를 substring 검색), 계획 행 `Source` 비교, 사람용 출력의 경로 검색, `p.source == w.movedTree` 비교 중 하나. 예: `TestUpdateConsolidation_MovesToHomeRootAndSkipsAgent` 의 "the update moved the agent-* tree" 는 migrate_sync_fix_test.go:66 `!m6Registered(t, w.repo, w.agentTree)` — 제품이 옮긴 것이 아니라 시험 도우미가 항상 거짓을 낸 것. `filepath.ToSlash(path)` 로 git 표기와 맞추면 해소되는 경로이고 CI 의 git 출력은 긴 이름 `C:/Users/runneradmin/…` (8.3 짧은 이름 아님). Windows 에서 녹색이 되는지는 미측정(Gaps).

반증된 쪽 (G1) — E4:
```
$ python3 -I iter2/symlink_check.py t2085-job.log
dest (as the test held it): 'C:\\Users\\runneradmin\\AppData\\Local\\Temp\\TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove3318261980\\001\\worktrees\\projects\\mig-repo\\legacy6'
state file body as printed: '[\n  {\n    "path": "C:\\\\Users\\\\runneradmin\\\\AppData\\\\Local\\\\Temp\\\\TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove3318261980\\\\001\\\\worktrees\\\\projects\\\\mig-repo\\\\legacy6",\n    "branch": "WT-legacy6",\n    "agent_name": "a",\n    "created_at": "2026-11-06T04:13:26.9742584Z"\n  }\n]'
test's check strings.Contains(raw, dest): False
decoded path field: 'C:\\Users\\runneradmin\\AppData\\Local\\Temp\\TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove3318261980\\001\\worktrees\\projects\\mig-repo\\legacy6'
decoded path == dest: True
```
(파이썬 repr 이라 `\\` 한 쌍 = 실제 백슬래시 1개.) 시드 값은 symlink 표기(`…\link\legacy6`)였는데 파일에는 목적지 `…\001\worktrees\projects\mig-repo\legacy6` 가 들어 있다 — 수정 전 Windows 에서도 재키잉은 성공했다. 실패는 JSON 원문(백슬래시 이스케이프)과 단일 백슬래시 경로의 substring 비교 때문이며, 이번 diff 는 이 단언(migrate_hardening_c_t2053_test.go:306)도 상태 파일 직렬화도 바꾸지 않는다. 시험 파일 전체에서 JSON 원문에 경로를 substring 검색하는 곳은 이 한 줄뿐(`grep -n "strings.Contains(string(raw\|..." *_test.go` 결과 중 경로를 찾는 줄은 c:306 하나, 나머지는 `"moved"`/`"result":"restored"` 같은 상수).

G2 관련 E5:
```
$ grep -E '"Action":"(pass|fail|skip)".*"Test":"TestApply_CardBound' t2085-job.log
{"Time":"2026-11-06T04:13:33.6655256Z","Action":"pass",...,"Test":"TestApply_CardBoundViaSymlinkSpelling_StillFollowsTheMove","Elapsed":4.49}
```
`rebindCardForPath`(migrate.go:1583) 은 정확 키 조회 실패 시 `worktreeroot.CanonicalSpelling`(worktreeroot.go:552, `filepath.Clean` 으로 시작) 대체 조회를 한다.

### C8. 숨겨진 제품 결함 탐색

A군 chmod 경로, B군 shim 경로, C 의 SameFile 은 OS 중립 제품 코드의 오류 처리 경로를 시험하는 것으로, 건너뛰기가 가리는 Windows 전용 제품 분기를 찾지 못했다. G1 도 제품 결함이 아니라 시험 단언 결함이다(제품은 올바르게 동작했음이 CI 로 관측됨). 이름을 댈 수 있는 숨겨진 제품 결함 없음.

## 2. Baseline-attribution

- 모든 명령은 2026-11-06 이 감사 실행 중, 트리 HEAD 67895b0ef44bc8097268f8b1e0d951c7ffa49b45(작업 트리 깨끗함 — 이 파일 작성 전)에서 실행. go1.26.8 darwin/arm64 로 트리에서 직접 컴파일(설치된 moai 바이너리 미사용, §2.2 해당 없음).
- CI 근거는 레인이 받아 둔 로그 사본 `t2085-job.log`(job 127703358711, Windows, 2026-11-06T04:0x–04:1xZ)를 이 감사가 새로 작성한 스크립트(iter2/parse_ci.py, dmsgs.py, symlink_check.py)로 다시 읽은 결과. 로그를 다시 내려받지 않았다. CI 가 돌린 트리는 기준 400cf31be0 과 시험 줄 번호가 일치(c:307 의 Errorf 는 HEAD 에서도 307행, 단언 306행).
- 강제 분기 overlay 는 레인의 t2085_overlay.py 를 재사용하지 않고 iter2/mk_overlay.py 로 다시 만들었다.

## 3. Gaps (관측하지 않은 것)

- Windows 실행 없음. 건너뛰기 분기, D군 수정, Windows 전용 시험(`TestParseWorktreeList_WindowsPathUsesTheOSSeparator`) 모두 Windows 에서 관측되지 않았다. G1 의 "계속 실패한다"는 CI 로그의 실제 파일 본문 + 바뀌지 않은 단언 코드로부터의 추론이며, 수정 후 Windows 실행으로 확인된 것은 아니다.
- D군 나머지 18건이 Windows 에서 녹색이 되는지는 미측정(진행 기록에도 공개됨). macOS 에서 D군 19건은 강제 분기 overlay 실행 안에서 pass(E3).
- 레인의 246s 실행(progress.md:30)은 재현하지 않았다. internal/cli 의 호출자 시험(launch_view, launcher)은 실행하지 않음 — 코드 읽기와 GOOS=windows vet 만.
- golangci-lint, /verify(moai gate), 커버리지, card-review, 교차 모델 감사(audit_multi/codex/glm)는 지시에 따라 이번 회차에서 실행하지 않음. Codex receipt 없음 — 판정이 FAIL 이므로 receipt guard 와의 충돌은 해당 없음.
- 도구 거부: 이 감사의 Bash 명령 1건(heredoc+여러 명령 복합)이 worktree 격리 가드에 의해 실행 전 거부되었다. 같은 내용을 스크립트 파일(iter2/parse_ci.py)로 옮겨 단순 명령으로 다시 실행했고 출력은 E1 에 그대로 인용. 대체로 "읽기만 한" 단계는 없다.

## 4. Residual-risk (관측했음에도 남는 위험)

- 강제 분기 overlay 는 `runtime.GOOS` 조건만 바꾼다. Windows 의 실제 파일시스템·PATH·git 동작은 재현하지 않으므로 A/B 건너뛰기가 Windows 에서 실제로 발동하는지는 Windows CI 가 첫 측정이다(코드상 조건은 `runtime.GOOS == "windows"` 하나뿐이라 위험은 낮음).
- D군 뒤에 다른 Windows 차이가 더 있을 수 있음(G1 이 그 실례). 같은 종류(JSON 원문에 Windows 경로 substring 검색)는 이 패키지 시험에서 c:306 하나뿐임을 grep 으로 확인했으나, 다른 형태(예: 사람용 출력 줄의 이스케이프, 대소문자)는 Windows 실행만 보여 준다.
- 스크래치 증거(iter2/ 의 ci-top.txt, ci-sub.txt, overlay.json, windows_skip_forced.go, forced1/2.json, analyze.py, symlink_check.py)는 세션 scratchpad 에만 있다. 결정적 출력 줄은 이 파일에 옮겼고, 나머지 원문은 세션과 함께 사라진다.

## Recommendations

1. G1 수리(테스트 전용): c:306 단언을 JSON 디코드 후 `path` 비교로 바꾸고, progress.md 원인 표를 D=18 + "시험 단언(JSON 이스케이프)" 1건으로 정정. 이 카드의 회차 상한(2회)에 도달했으므로 수리 후 확인 방식(델타 재감사 허용 여부 또는 후속 카드 분리)은 리더가 정한다. 반복 수리에서 태어난 결함이 아니므로 repair stop rule(연속 2회 repair-born)에는 해당하지 않는다.
2. G2/G3(선택): 주석·기록 문구를 CI 관측에 맞게 정정.
3. 최종 확인은 리더 소관의 Windows CI 실행.

## Operational Notes (unverified)

- G1 수정 후 macOS 확인 명령: `go test -count=1 -run 'TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove' ./internal/cli/worktree/`. 상태: assumption (수정 전이므로 미실행).
- 덮임 범위 재확인은 이 파일 C2 의 overlay 방식으로 macOS 에서 측정 가능. 상태: measured (이 감사에서 실행, E2/E3).
