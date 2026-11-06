auditor-model: claude-opus-5-5

# t2085 독립 감사 판정 (sync-auditor, 1회차)

verdict: FAIL
audited_sha: dfcc5b920391759ab18e00f147a15a7d3653fe98

- 카드: t2085 (Windows job 127703358711, internal/cli/worktree 실패 52건)
- 감사 대상 트리: /Users/goos/.moai/worktrees/projects/moai-adk-go/t2085, 브랜치 WT-multi-os-test-3
- 기준 커밋: 400cf31be0 (t2079 749bdc2640 병합). 판정 대상 커밋: fa6d5e3115 (테스트 전용 Windows 건너뛰기), dfcc5b9203 (제품 수정 + 테스트 도우미 + Windows 전용 테스트)
- 점수: 0.62 (Functionality 45 / Security 95 / Craft 70 / Consistency 85, 가중 평균은 참고값이며 판정은 차단 결함 F1이 결정)
- 감사 방식: 유료·Codex·GLM 호출 없음(지시대로). 패키지 전체 시험과 go test ./... 은 실행하지 않음. 이 감사는 Windows 에서 아무것도 실행하지 않았음.
- 반복 이력: 1회차 (이전 감사 없음)

## 판정 요약

원인 표의 분할(18+14+1+19=52)은 CI 로그와 정확히 일치하고, 제품 수정(ParseWorktreeList의 filepath.FromSlash)은 macOS/Linux 에서 동작을 바꾸지 않으며 호출자 중 git의 슬래시 표기에 기대는 곳도 없다. 그러나 A군(posix-mode) 18건 중 2건 — `TestRollback_UnstatableDestination_IsNotTreatedAsAbsent`, `TestRollback_UnstatableSource_IsNotTreatedAsGone` — 은 공유 도우미(`t2053Chmod`/`t2053SkipIfRoot`)를 거치지 않고 `os.Chmod`를 직접 호출하므로 이번 건너뛰기가 닿지 않는다. 두 시험은 Windows 에서 여전히 같은 이유로 실패할 것이고, 진행 기록은 "A군은 모두 도우미를 호출한다", "도우미 수준 건너뛰기가 A·B·C 를 덮는다"고 적어 이를 과대 주장한다. 카드 목표(52건 처분)가 미달이고 기록에 과대 주장이 있으므로 FAIL.

## 결함 목록 (structured defect-list)

- F1 [High] [blocking] [outside — 재현 증거 있음] internal/cli/worktree/migrate_hardening_d_t2053_test.go:127, :152 — A군으로 분류된 두 시험이 `os.Chmod(parent, 0o000)`를 직접 호출하고 `os.Geteuid()==0` 만 확인한다. `t2085SkipOnWindows` 가 호출되지 않으므로 Windows 에서 건너뛰지 않는다. Windows 분기를 강제한 overlay 실행에서 두 시험만 `--- PASS`(실행됨)로 남고 나머지 31건은 SKIP 되었다(아래 증거 E3). CI 로그상 두 시험의 실패 문구("rollback treated an uninspectable destination as absent", "rollback skipped a step whose source it could not inspect")는 chmod 전제가 성립하지 않아 생긴 것이며, 제품 수정(D군)으로 바뀌지 않는다. — Required fix: 두 시험 첫 줄에 `t2085SkipOnWindows(t, t2085ReasonPosixMode)`를 넣거나(또는 `t2053SkipIfRoot`/`t2053Chmod` 로 바꾸어) 건너뛰기를 적용하고, progress.md 의 원인 표 A행 "all call t2053Chmod / t2053SkipIfRoot"와 18행 "the helper-level skip covers them"을 사실대로 고친다. 수리 후 같은 overlay 방식(E3)으로 33건 전부 SKIP 되는지 재확인한다.
- F2 [Medium] [blocking — 기록 정직성] .moai/reports/t2085/progress.md:12, :18, :27 — (a) 12행 "all call t2053Chmod / t2053SkipIfRoot"는 18건 중 2건에서 거짓. (b) 18행 "Group A, B and C lists: the helper-level skip covers them"은 33건 중 31건만 참. (c) 27행은 `TestRollback_UnstatableDestination_IsNotTreatedAsAbsent`를 "touched tests (helper skips must not skip on macOS)" 표본으로 들었으나 이 시험은 이번 diff 가 건드리지 않았고 건너뛰기 경로도 없다 — macOS PASS 는 아무것도 증명하지 않는다. — Required fix: F1 수리와 함께 세 줄을 관측 사실로 정정한다.
- F3 [Low] [optional] [in-diff 효과] migrate_hardening_e_t2053_test.go:379, :389, :400 — 52건에 속하지 않는 `TestMovedBy_UnknownIsProtected`의 하위 시험 3개(token_cannot_be_read, move_list_cannot_be_read, git_cannot_list_the_worktrees)가 Windows 에서 새로 건너뛰게 된다. CI 로그상 이 시험은 Windows 에서 5개 하위 시험 모두 PASS 였다(전제가 성립하지 않았는데도 통과 = Windows 에서는 공허한 통과로 추정). 숨겨지는 제품 결함은 없다고 판단하지만 기록에 공개되지 않았다. — Required fix(선택): progress.md Gaps 에 "52건 밖 TestMovedBy_UnknownIsProtected 하위 3개가 Windows 에서 새로 SKIP" 한 줄 추가.
- F4 [Low] [optional] [in-diff] migrate_hardening_a_t2053_test.go:49 — 건너뛰기가 함수 첫 줄에 있어 Windows 에서 유효한 내용 검사(67-70행: 두 번째 쓰기 후 JSON 유효성과 "pending" 포함)까지 함께 건너뛴다. os.SameFile 단언만 Windows 에서 건너뛰어도 된다. 제품 결함을 숨기지는 않음(writeFileAtomic 은 OS 중립, journalReplace=atomicfile.Replace). — Required fix(선택): SameFile 단언만 OS 조건으로 감싸기.
- F5 [Low] [optional] progress.md:6 — "58 incl. subtests"는 재도출 값 67(최상위 52 + 하위 15)과 다르다. — Required fix(선택): 숫자 정정.
- F6 [Info] [optional] [outside] internal/cli/worktree/migrate.go:286-299, :592-598 — Windows 에는 lsof 가 없으므로 실제 사용 환경에서 `defaultActiveProcessCWDs`가 실패하고 계획기는 모든 트리를 "cannot inspect active process working directories"로 건너뛴다(안전 측 실패). 즉 Windows 에서 migrate 는 사실상 아무것도 옮기지 않는다. 이번 diff 와 무관하며 B군 건너뛰기가 숨기는 결함도 아니다(시험은 probe 를 stub 함). 운영자 참고용으로만 기록. 상태: inferred (코드 읽기 + CI 메시지 `exec: "lsof": executable file not found in %PATH%`).
- F7 [Info] [optional] [outside, baseline] migrate.go:1051 — golangci-lint `unused: migrateOneTree`. 판정 대상 커밋이 건드리지 않은 줄이므로 기준선 결함.

## 1. Claim / Evidence / Baseline-attribution

### C1. 실패 52건과 분할 18+14+1+19 는 CI 로그와 일치한다 — 확인

Evidence (CI 로그 go test -json 이벤트를 독립 스크립트로 재도출):
```
$ python3 -I <scratch>/(인라인 스크립트) t2085-job.log
52 67 {'github.com/modu-ai/moai-adk/internal/cli/worktree'}
$ diff audit-top.txt claimed.txt && echo SAME52
SAME52
$ awk ... t2085-class3.tsv   # A=Chmod, B=FakeBin|BlockingGit, C=JournalWrite, D=나머지
18 14 1 19
```
하위 시험 실패 15건(52+15=67). 기록의 "58"은 불일치(F5).

### C2. 표본 검토: 주장한 전제가 실제 실패 원인인가 — A·B·D 에서 9건 확인, A 에서 반례 2건

- A `TestTreeIdent_UnreadableToken_IsAnError` (e:328-334): `t2053SkipIfRoot` → `t2053Chmod(token, 0o000)`. CI: "TreeIdent read an unreadable token as an empty one: {... Token:9807ae…}" — 토큰이 읽힘 = chmod 무효. 전제 일치.
- A `TestAppendMoveListLine_UnreadableList_IsAnError` (e:423-433): `t2053Chmod(mlPath, 0o200)`. CI: "appended to a list it could not read". 전제 일치.
- A `TestRollback_UnstatableDestination_IsNotTreatedAsAbsent` (d:127-147), `TestRollback_UnstatableSource_IsNotTreatedAsGone` (d:152-169): 전제는 chmod 맞음, 그러나 도우미를 거치지 않음 → F1.
- B `TestMigrationEngine_WorktreeListTimeoutIsAnError` (review:30-52): `t2053ListTimeout` → `t2053BlockingGit("list")` → `t2053FakeBin("git", "#!/bin/sh…")`. CI: "a failed worktree listing returned success". 전제 일치.
- B `TestApply_ManifestGitFailure_MovesNothing` (c:243): `t2053FailingGit` → `t2053FakeBin`. CI: "moved = 2, want 0". 전제 일치.
- B `TestDefaultActiveProcessCWDs_TimesOutOnBlockedLsof` (bounded:178): CI "exec: \"lsof\": executable file not found in %PATH% does not wrap ErrCommandTimeout". 전제 일치.
- D `TestMigrate_ApplyMovesEligibleAndSkipsProtected` (command:225-246): CI 출력에서 제품은 `legacy6 … [moved]`라고 보고했는데 시험은 "legacy6 was not moved"와 "protected tree … was moved or removed"를 냈다. 두 단언 모두 `m6Registered`(=`strings.Contains(git porcelain, OS 표기 경로)`)에 기대므로 표기 불일치가 원인. 로그 행: `C:/Users/runneradmin/…/.moai/worktrees/legacy6 → C:\Users\runneradmin\…\legacy6 [moved]`. 전제 일치.
- D `TestPlanMigration_SkipsATreeAnActiveProcessIsInside` (bounded:404-431): `p.source == w.movedTree` 비교. p.source 는 git 표기, w.movedTree 는 OS 표기 → "occupied=false free=false". 제품 수정으로 해결되는 경로. 전제 일치.
- D `TestMigrationSkipsOnUnreadableRegistry` (sync_fix:342-365): 이름의 "unreadable"은 chmod 가 아니라 깨진 JSON 기록 — chmod 전제 아님. `strings.Contains(out, w.movedTree)` 가 git 표기 출력과 OS 표기 경로를 비교 → D 분류 타당.
- D `TestMovedBy_ProtectsAFailedReversePendingTree`, `TestReverseMove_ErrorAfterMoveCompleted_IsTreatedAsMovedBack`: 실패 단언이 `m6Registered(...)` — D 분류 타당.

건너뛰기가 제품 결함을 숨기는 사례 탐색: A군의 chmod 전제는 OS 중립 오류 처리 경로를 시험하는 것이고(제품 쪽에 Windows 전용 분기 없음), B군은 fault 주입 수단 자체가 없으며, C 는 os.SameFile 의 Windows 지연 비교 탓이다. 이름을 댈 수 있는 숨겨진 제품 결함은 찾지 못했다. 다만 B군 일부(`TestPlanMigration_SkipsEveryTreeWhenLsofTimesOut`, `TestApply_ManifestGitFailure_MovesNothing` 등)의 CI 메시지에도 D 표기 증상(`Source:C:/…`)이 섞여 있었고, 이는 D 의 제품 수정이 덮는다. 별도로 Windows 의 lsof 부재(F6)는 건너뛰기와 무관한 기존 동작.

### C3. 건너뛰기는 runtime.GOOS == "windows" 에서만 발동하고, macOS 에서는 아무것도 건너뛰지 않는다 — 확인

Evidence:
```
$ go test -count=1 -v -run 'TestJournalWrite_ReplacesTheFileAtomically|TestDefaultActiveProcessCWDs_TimesOutOnBlockedLsof|TestAppendMoveListLine_UnreadableList_IsAnError|TestTreeIdent_UnreadableToken_IsAnError' ./internal/cli/worktree/
=== RUN   TestDefaultActiveProcessCWDs_TimesOutOnBlockedLsof
--- PASS: TestDefaultActiveProcessCWDs_TimesOutOnBlockedLsof (0.30s)
=== RUN   TestJournalWrite_ReplacesTheFileAtomically
--- PASS: TestJournalWrite_ReplacesTheFileAtomically (0.02s)
=== RUN   TestTreeIdent_UnreadableToken_IsAnError
--- PASS: TestTreeIdent_UnreadableToken_IsAnError (1.99s)
=== RUN   TestAppendMoveListLine_UnreadableList_IsAnError
--- PASS: TestAppendMoveListLine_UnreadableList_IsAnError (0.04s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	3.240s
```
SKIP 줄 0개. 도우미 본문: `if runtime.GOOS == "windows" { t.Skipf(...) }` (windows_skip_t2085_test.go:28).

### C4. 도우미 호출자 수: 기록은 33건을 주장 — 실제로 건너뛰기에 닿는 52건 소속 시험은 31건 (+52건 밖 1건)

Evidence (호출 지점을 감싸는 함수로 매핑, 간접 도우미 `t2053ListTimeout`/`t2053FailingGit` 포함):
- A: `t2053Chmod`/`t2053SkipIfRoot` 호출 시험 17개, 그중 52건 소속 16개 + 52건 밖 `TestMovedBy_UnknownIsProtected` 1개. 직접 `os.Chmod`만 쓰는 52건 소속 2개(F1).
- B: FakeBin 직접 2, BlockingGit 직접 7, ListTimeout 경유 3, FailingGit 경유 2 = 14 (전부 52건 소속).
- C: 1.
- 합계: 52건 소속 31개가 건너뛰기에 닿음, 2개는 닿지 않음.

Windows 분기를 강제한 overlay 실행(트리 수정 없음, `go test -overlay` 로 windows_skip_t2085_test.go 의 조건만 `runtime.GOOS != ""`로 바꾼 사본 사용) — E3:
```
$ go test -count=1 -v -timeout 280s -overlay <scratch>/ovl/overlay.json -run '^(…33개…|TestMovedBy_UnknownIsProtected)$' ./internal/cli/worktree/
--- SKIP: TestDefaultActiveProcessCWDs_TimesOutOnBlockedLsof (0.00s)
…
--- PASS: TestRollback_UnstatableDestination_IsNotTreatedAsAbsent (2.64s)
--- PASS: TestRollback_UnstatableSource_IsNotTreatedAsGone (3.18s)
…
    --- SKIP: TestMovedBy_UnknownIsProtected/token_cannot_be_read (0.00s)
    --- SKIP: TestMovedBy_UnknownIsProtected/move_list_cannot_be_read (0.00s)
    --- SKIP: TestMovedBy_UnknownIsProtected/git_cannot_list_the_worktrees (3.04s)
…
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	72.799s
```
33개 중 31개는 최상위 SKIP 또는 실패했던 하위 시험이 SKIP, 2개(F1)는 그대로 실행됨. `TestPlan_PendingIntent_NotProvablyUnmoved_IsInterrupted`는 실패했던 하위 시험 `the_destination_cannot_be_inspected`만 SKIP, 다른 하위 시험은 PASS — 의도대로.

### C5. 제품 수정은 macOS/Linux 에서 동작을 바꾸지 않는다 — 확인

Evidence (설치된 툴체인 소스):
```
go version go1.26.8 darwin/arm64
internal/filepathlite/path.go:184: func FromSlash(path string) string {
  if Separator == '/' { return path }
  return replaceStringByte(path, '/', Separator)
internal/filepathlite/path.go:176: func ToSlash(path string) string {
  if Separator == '/' { return path }
```
POSIX 에서 FromSlash/ToSlash 는 입력을 그대로 반환하므로 classify.go:112 와 migrate_command_test.go:163 은 바이트 단위로 이전과 같다(C6 의 m6Registered 도 동일).

호출자 검토(Windows 에서 슬래시 표기에 기대는가):
- internal/cli/launch_view.go:273 → 283행에서 이미 `filepath.FromSlash(rec.Path)` 후 `resolveSymlinks(filepath.Clean(...))` 비교. 이중 변환은 멱등.
- internal/cli/launcher.go:519 → 526행에서 이미 `filepath.FromSlash(rec.Path)`, 이후 `filepath.Base`/`filepath.Rel`. 멱등.
- worktree/list.go:97 → `filepath.Clean(rec.Path)`. Windows 에서 Clean 은 원래 `/`를 `\`로 바꾸므로 결과 동일.
- worktree/movelist.go:252 → `worktreeroot.Same(reg, p)` 비교.
- worktree/migrate.go:464 → 계획 행 source; `inMigrationScope`(453행)는 `filepath.ToSlash` 후 `"/.moai/worktrees/"` 검사 — 표기 무관.
- worktree/migrate.go:1540 → `pathListed`가 `canonicalTreePath`(EvalSymlinks 또는 Clean) 비교.
- classify.go 내부: `underPath`/`filepath.Rel`/`filepath.Base`, 229행 `filepath.ToSlash(rel)` — 표기 무관.
- 패키지 제품 코드 전체에서 슬래시 리터럴 검색 결과는 위 453행과 브랜치 이름 처리(clean.go:393, sweep.go:110/287/300)뿐. 정규식·URL 구성에 경로를 쓰는 곳 없음.
결론: 슬래시 표기에 기대는 호출자 없음.

```
$ env GOOS=windows go vet ./internal/cli/worktree/ ./internal/cli/
exit=0
$ go vet ./internal/cli/worktree/
vet-exit=0
$ gofmt -l <변경된 7개 파일>
gofmt-exit=0   (출력 없음)
```

### C6. m6Registered 는 POSIX 에서 이전과 정확히 같다 — 확인

ToSlash 가 POSIX 에서 항등(C5 증거)이므로 `strings.Contains(porcelain, filepath.ToSlash(path))` ≡ `strings.Contains(porcelain, path)`.

### C7. Windows 전용 시험은 Windows 용으로 컴파일되고, 변경 전 파서에서는 실패한다 — 컴파일은 확인, 실행은 안 됨

Evidence:
```
$ env GOOS=windows go test -c -o <scratch>/wt-windows.test.exe ./internal/cli/worktree/
(출력 없음, 성공)
$ go tool nm <scratch>/wt-windows.test.exe | grep -E 'TestParseWorktreeList_WindowsPathUsesTheOSSeparator|t2085SkipOnWindows'
1406168b0 T github.com/modu-ai/moai-adk/internal/cli/worktree.TestParseWorktreeList_WindowsPathUsesTheOSSeparator
14067f090 T github.com/modu-ai/moai-adk/internal/cli/worktree.t2085SkipOnWindows
$ go test -list 'ParseWorktreeList|Classify' ./internal/cli/worktree/    # darwin: Windows 전용 시험은 목록에 없음
```
추론: 변경 전 파서는 `Path: rest` = `"C:/Users/x/repo"`를 저장하므로 기대값 `` `C:\Users\x\repo` ``와 달라 실패한다. 변경 후에는 FromSlash 로 일치. 이 시험은 어디에서도 실행된 적이 없다.

### C8. 지정 필터 시험 — 통과

```
$ go test -count=1 -timeout 280s -run 'ParseWorktreeList|Classify' ./internal/cli/worktree/
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	1.411s
$ go test -list 'ParseWorktreeList|Classify' ./internal/cli/worktree/
TestClassify_MeasuredShapes
TestClassify_NeverMovedClassesAreNotMoaiMade
TestClassify_SpellingSafeRootsAndPrimary
TestClassify_CaseDifferingDirectoriesStayDistinctOnACaseSensitiveVolume
TestParseWorktreeList_Porcelain
TestParseWorktreeList_RealGit
TestClassify_ListingEndToEnd
TestClassify_AgentNamedTreeFollowsItsRoot
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	0.760s
```
훑은 시험 수: 최상위 8개.

### C9. diff 범위

```
$ git -C <tree> diff --stat 400cf31be0 dfcc5b9203
 7 files changed, 78 insertions(+), 2 deletions(-)
```
제품 파일은 classify.go 1개(+4/-1), 나머지는 _test.go. 보안 면: 새 입력 경로·비밀·명령 실행 표면 없음(Security 95).

## 2. Baseline-attribution

- 모든 명령은 2026-11-06 이 감사 실행 중, 트리 HEAD dfcc5b920391759ab18e00f147a15a7d3653fe98(작업 트리 깨끗함) 에서 실행. go1.26.8 darwin/arm64 툴체인으로 트리에서 직접 빌드(설치된 moai 바이너리 미사용 — §2.2 해당 없음).
- CI 근거는 레인이 받아 둔 로그 사본(scratchpad t2085-job.log, 러너 버전 2.338.2, Windows, 2026-11-06T04:02Z)을 이 감사가 독립 스크립트로 다시 읽은 결과. 로그 자체를 다시 내려받지는 않았다.
- CI 가 돌린 트리는 PR #1969 head 이며 시험 줄 번호가 기준 400cf31be0 과 일치함(예: command:228 Logf → HEAD 에서는 m6Registered 2줄 추가로 230).

## 3. Gaps (관측하지 않은 것)

- Windows 실행 없음. 건너뛰기 분기, D군 수정, Windows 전용 시험 모두 Windows 에서 관측되지 않았다. F1 의 "두 시험이 Windows 에서 계속 실패한다"는 CI 로그 + 코드 + 강제 분기 overlay 실행으로부터의 추론이다.
- D군 19건이 수정 후 Windows 에서 녹색이 되는지는 미측정. macOS 에서 D군 시험을 이 감사가 다시 돌리지 않았다(레인의 246s 실행 기록만 있음, 이 감사는 그 출력을 직접 보지 못함).
- internal/cli 의 ParseWorktreeList 호출자 시험(launch_view, launcher)은 실행하지 않음. 코드 읽기와 GOOS=windows vet 만 수행.
- golangci-lint 는 실행함(기준선 결함 1건, F7). /verify(moai gate), 커버리지 측정, card-review, 교차 모델 감사(audit_multi/codex/glm)는 지시에 따라 실행하지 않음.
- 도구 거부: 이 감사의 Bash 명령 4건이 worktree 격리 가드에 의해 실행 전 거부되었다(복합 명령·파이프·리다이렉트). 같은 내용을 단순 명령으로 나누어 다시 실행했으며, 대체로 읽기만 한 것은 없다. 첫 번째 지정 시험 명령도 거부 후 리다이렉트 없이 다시 실행해 출력을 그대로 인용했다.

## 4. Residual-risk (관측했음에도 남는 위험)

- D군 뒤에 두 번째 Windows 차이가 숨어 있을 수 있음(8.3 짧은 이름, 대소문자, `git worktree move` 의미, Windows 심볼릭 링크 권한 — `TestApply_StateEntrySpelledViaSymlink_StillFollowsTheMove`). 레인 기록에도 공개됨.
- Windows 기존 사용자 상태(journal·move list·card binding)가 git 표기로 저장되어 있다면, 새 코드는 OS 표기로 찾는다. `pathListed`/`canonicalTreePath`, `rebindCardForPath`의 `CanonicalSpelling` 대체 경로가 흡수할 것으로 보이나 Windows 에서 측정되지 않음. 레인 커밋 메시지의 "card-binding re-key addressed a key that did not exist"는 이 대체 경로를 감안하면 과장일 수 있다(inferred).
- F6: Windows 에서 lsof 부재로 migrate 가 모든 트리를 건너뛰는 기존 동작.
- 강제 분기 overlay 는 `runtime.GOOS` 조건만 바꾼다. Windows 의 실제 파일시스템·PATH 동작까지 재현하지 않는다.
- 스크래치 증거(overlay.json, windows_skip_forced.go, audit-fails.txt, wt-windows.test.exe)는 세션 scratchpad 에만 있으며 이 파일에 결정적 출력 줄을 옮겨 두었다. 나머지 원문은 세션 종료와 함께 사라진다.

## Recommendations

1. F1 수리: migrate_hardening_d_t2053_test.go 의 두 시험에 `t2085SkipOnWindows(t, t2085ReasonPosixMode)` 추가(테스트 전용, fa6d5e3115 와 같은 성격의 별도 커밋).
2. F2: progress.md 원인 표 A행·18행·27행 정정, F3·F5 공개.
3. 재감사 범위: F1/F2 델타만. 확인 방법은 E3 과 같은 overlay 강제 실행에서 33건 모두 SKIP(또는 실패 하위 시험 SKIP)인지 보는 것.
4. 최종 확인은 리더 소관의 Windows CI 실행.

## Operational Notes (unverified)

- 도우미 수준 건너뛰기의 적용 범위는 `go test -overlay`로 조건만 뒤집은 사본을 끼워 macOS 에서 측정할 수 있다 — 명령: E3. 상태: measured (이 감사에서 실행).
