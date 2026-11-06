# t2094 진행 기록 — t2075 감사 후속 (Windows 셸 cwd 보호, 바쁜 트리 안내 문구)

Card t2094 · run to2bbk · lane-31 · Class B (run → sync, SPEC 없음, 사원조사는 레인 소유) · 2026-11-06

## 배차·임대

- 배차: 리더 (leader, run to2bbk generation 2) — "t2075 감사 후속 — Windows PowerShell·MSYS/Cygwin 셸 cwd 보호, 바쁜 트리 안내 문구"
- 규칙 파일: `.moai/reports/pr-sweep-20261106/af-flow-only.md` 전부 준수 (착지=local main만, Claude sync-auditor 1회, Codex 도구 금지, t2086 main 착지 전 complete 금지)
- 임대: `moai factory next --card t2094 --run to2bbk` → worktree=t2094, stage=-, picked
- 워크트리: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2094` (branch `WT-2a04549ec77bbb0f395c6487`, factory 생성)

## 기준 병합 (t2075)

- `git cat-file -e 2a04549ec77bbb0f395c6487e52d28a61c09ba7f^{commit}` → EXISTS
- `git log -1`: `2a04549ec77bbb0f395c6487e52d28a61c09ba7f [t2075] fix(worktree): Windows migration no longer skips every tree for want of lsof (card t2075)`
- 관계: `git merge-base --is-ancestor 9b632c1eb3 2a04549ec7...` → exit 0 (t2075는 베이스의 후손) → fast-forward 병합
- `git merge 2a04549ec7...` → HEAD = 2a04549ec7, 트리 깨끗함

## 카드 범위 (factory next 출력 원문)

> t2075(2a04549ec77bbb0f395c6487e52d28a61c09ba7f)는 Windows에서 'OS가 사용 중 폴더의 이름 변경을 거부한다'는 전제로 lsof 스캔을 대신함. 감사 F2: PowerShell Set-Location·MSYS/Cygwin 셸 cwd는 이 보호를 못 받을 수 있음(미검증 가설), F3: 바쁜 트리가 failed+무관한 rollback 안내로 보임. Windows 프로세스 cwd 확인 수단(PEB 읽기 등) 조사와 F3 문구 수리. 기준 t2075 위, Windows 판정은 CI.

## 레인 코드 조사 (run 단계 사원조사)

F3 경로 (직독 확인):
- `internal/cli/worktree/migrate.go:1136-1160` — `forwardGitMove` 실패 → `forwardFailureLocation`==atSource → outcome "failed", reason은 git 에러문 원문 (`git worktree move: %v`)
- `migrate.go:199-200` — 요약에서 `report.failed+report.reversed > 0`이면 journal+rollback 힌트 출력. 바쁜 트리는 한 발도 안 움직였으므로 rollback은 그 트리에 대해 무관한 안내 (F3 본체)
- `migrate.go:1361-1376` — failed 행 렌더링 (`Worktree migration failed %s: %s`)

F2 시임:
- `migrate.go:288` `var activeProcessCWDProbe = defaultActiveProcessCWDs` (주입 시임)
- `migrate_cwd_windows.go:14-16` — Windows 스텁: 항상 `errProcessCWDOSGuarded` 반환 (OS 가드 전제에 전부 의존)
- plan 스킵 체인 `migrate.go:583-590`: OS-guarded는 스킵 없음, 그 외 오류는 "cannot inspect active process working directories" 스킵
- 테스트 계약 (`migrate_oscwd_t2075_test.go`): OS-guarded → 스킵 없음 / 그 외 오류 → 전체 스킵 유지
- 의존성: `golang.org/x/sys v0.49.0` 이미 go.mod 존재 (신규 의존성 불필요)

## 레인 사슬 판단 1 — F2 스캔 계약

스페셜리스트에 지시한 계약 (레인 판단, 근거: t2075의 "스캔 실패로 전체 스킵" 회귀 금지 의도 유지):

1. PEB 스캔(NtQueryInformationProcess + ReadProcessMemory, toolhelp32 열거, x/sys/windows)으로 실제 Win32 cwd 목록 생산
2. **총체적 실패**(단 하나의 프로세스 cwd도 관측 불가) → `errProcessCWDOSGuarded` 반환 — t2075가 고친 "전체 스킵" 회귀를 재도입하지 않고 오늘과 동일하게 OS 가드+move 실패 처리에 위임
3. **부분 성공**(동일 사용자 프로세스는 읽히고 타 사용자는 거부) → 부분 목록+nil 오류 — 못 읽은 프로세스의 잠금은 move 시점 OS 가드+F3 문구가 받는다
4. 빈 목록+nil 오류는 금지 (t2075 스텁 주석의 불변식 — "빈 목록은 안에 프로세스 없음으로 읽힌다")

F2 전제 분석(PowerShell Set-Location의 Win32 cwd 불동기, MSYS 셸 cwd 처리)은 이 호스트에서 측정 불가 → t2075 선례대로 windows 빌드 태그 전제 측정 시험으로 CI에 맡김.

## 위임

- `manager-develop` 스폰 (depth-1 리프, effort=medium — 카드 클래스 B 매핑)
- 범위: F2(PEB 스캔 구현+전제 시험) + F3(바쁜 트리 reason 문구+rollback 힌트 게이팅), POSIX 파일 동결, outcome 분류 불변
- 검증 지시: GOOS=windows `go vet ./internal/cli/worktree/`, GOOS=windows `go test -c` (컴파일만), 공통 경로 표적 `-run` 시험 (전체 수트 금지 — 레인 로컬 규율)
- 상태: 백그라운드 실행 중 — 완료 통지 대기 (결과를 예측·보고하지 않음)

## 대기

- (아래에 스페셜리스트 반환·검증 결과가 기록됨)

### 감시 기록 (05:43:20Z, watchdog pass 1)

- renew: renewed (만료 05:58:27Z) · HEAD 2a04549ec7 (불변) · evidence mtime 05:38:56Z(레인 자작) · 통합 창 free
- 분류: awaited-delegate — manager-develop 생존(완료 통지 미도착), 산출물(커밋) 아직 없음 → 명시 대기, 재확인점=다음 깨어남(+≤280s)
- 결정 기록: decided_by=lane-31 watchdog evidence_refs=.moai/state/watchdog/t2094.json ladder_path=§3.5-explicit-wait

### 감시 기록 (05:49:30Z, watchdog pass 2)

- renew: renewed (만료 06:04:17Z) · HEAD 2a04549ec7 (불변) · 통합 창 free
- 분류: awaited-delegate(대기 유지) — 위임 증거 이동 중: 트리에 M migrate.go / M migrate_cwd_windows.go / M migrate_oscwd_windows_test.go / ?? migrate_busy_tree_t2094_test.go 관측(스페셜리스트 작업 흔적). 컴파일 진단(UnusedVar w)은 편집 중 상태 — 레인 개입 없음
- 결정 기록: decided_by=lane-31 watchdog evidence_refs=.moai/state/watchdog/t2094.json ladder_path=§3.5-explicit-wait

### 감시 기록 (05:54:20Z, watchdog pass 3)

- renew: renewed (만료 06:09:10Z) · HEAD 2a04549ec7 (불변) · 통합 창 free
- 분류: awaited-delegate(대기 유지) — 위임 증거 계속 이동: pass 2 대비 migrate_hardening_c_t2053_test.go 신규 수정 관측(더티 5파일). 블로커 조건(두 번 연속 무변화) 미충족
- 결정 기록: decided_by=lane-31 watchdog evidence_refs=.moai/state/watchdog/t2094.json ladder_path=§3.5-explicit-wait

## run 단계 결과 (스페셜리스트 반환 + 레인 재검증)

### 커밋

- `1852d1a34deec532f4e715fff2053920141e3cf0` — `[t2094] fix(worktree): real Windows process-cwd scan, busy-tree failure guidance, gated rollback hint (card t2094)` (base 2a04549ec7, 5파일 +520/−23, 🗿 MoAI 트레일러)
- 직독: `git show --stat` → migrate.go(28)·migrate_cwd_windows.go(174)·migrate_oscwd_windows_test.go(205)·migrate_busy_tree_t2094_test.go(신설 123)·migrate_hardening_c_t2053_test.go(13); `git status --short` → 깨끗함; migrate_cwd_posix.go 미변경(동결 준수)

### F2 — Windows PEB 프로세스 cwd 스캔 (스페셜리스트 구현, 요지)

- CreateToolhelp32Snapshot 열거 → PROCESS_QUERY_LIMITED_INFORMATION|PROCESS_VM_READ → NtQueryInformationProcess(ProcessBasicInformation)로 PEB → ReadProcessMemory로 ProcessParameters(0x20/0x10)→CurrentDirectory(0x38/0x24 UNICODE_STRING). 동일 사용자 프로세스는 무승격 관측, 타 사용자는 개별 스킵. `readProcessValue`에 @MX:WARN+@MX:REASON. 신규 의존성 없음(x/sys v0.49.0), cgo 없음
- 계약(레인 판단 이행): 총체적 실패(스냅샷 실패·ntdll 해석 실패·관측 0건) → `errProcessCWDOSGuarded` — t2075의 "전체 스킵" 회귀 없음. 부분 성공 → 부분 목록+nil. 빈 목록+nil은 구조적으로 불가(len==0 → 오류)

### F2 전제 분석 (CI 측정 시험으로 인코딩 — 이 호스트 미측정, gaps 아닌 설계상 한계)

1. PowerShell 5.1/7: Set-Location은 프로바이더 위치만 갱신, Win32 SetCurrentDirectory 미호출 → Win32 cwd는 실행 디렉터리에 잔류. **기계적 확인: PowerShell Set-Location 안의 트리는 rename 성공 — OS 가드도 PEB 스캔도 못 봄. 문서화된 잔여 갭(스캔으로 닫을 수 없는 기제 한계)**
2. cmd.exe: cd(/d 포함)는 SetCurrentDirectory 호출 → rename 거부
3. MSYS2/Git-for-Windows bash: msys-2.0.dll chdir이 Win32 cwd 동기화 → rename 거부(빌트인만 사용해 측정 오염 방지, bash.exe 부재 시 skip)

### F3 — 바쁜 트리 실패 렌더링

- atSource 실패 시 `failedAtSourceReason`이 `activeProcessCWDProbe()` 재탐사 → 소스가 관측 목록에 있으면 reason을 "a process is inside this worktree — close it or move its working directory elsewhere, then run the migration again (<git 오류 원문>)"로 교체. 프로브 오류(OS-guarded 포함) 시 원래 git 오류문 유지 — 하위 오류 상실 없음
- 요약 journal/rollback 힌트 게이트: `report.moved > 0 || report.interrupted > 0`일 때만 출력. 패키지 내 의존 시험 1건(TestApply_FailedTrees_PrintTheJournalPath) 의도 변경(2트리 중 1트리 실패로 저널에 실제 이동 존재), 전부-실패-무-힌트 케이스 신규 시험 고정. outcome 분류·exit 코드 불변

### 레인 재검증 (이번 런 직접 실행 — 귀속: 이 트리 HEAD 1852d1a34d)

- `GOOS=windows GOARCH=amd64 go vet ./internal/cli/worktree/` → windows-vet-exit=0
- `GOOS=windows GOARCH=amd64 go test -c -o /dev/null ./internal/cli/worktree/` → windows-compile-exit=0
- `go test ./internal/cli/worktree/ -run 'TestApply_BusyTree|TestApply_AllFailures|TestMigrateSummary' -count=1` → ok 10.433s (PASS)

### 재사용 귀속 (verification-plan-contract)

- 스페셜리스트 실행분(같은 트리 키 — 커밋 전 워킹 트리 내용 = 커밋 1852d1a34d 내용 동일): darwin vet exit 0, windows-386 vet exit 0, `-run 'TestPlanMigration|TestMigrate|TestApply'` ok 154.880s, gofmt -l 무출력 → owner=manager-develop, reused_from=t2094 run 스페셜리스트 보고, rerun_reason=없음(키 동일)

### 스페셜리스트 명시 미측정 (Gaps)

- Windows 런타임 동작 전부(PEB 스캔 엔드투엔드, 전제 시험 3종) → CI Windows 러너 몫. 이 호스트는 컴파일만(amd64·386)
- 비트수 교차 읽기(64비트 moai가 32비트 프로세스 PEB) → 실패 읽음으로 해당 프로세스 스킵(코드 주석 문서화, 어디서도 미측정)
- PowerShell 잔여 갭(Set-Location이 OS 가드·스캔 양쪽에 비가시) → 기제 한계, 코드 경로로 닫지 않음

### 스페셜리스트 잔여 참고 (리더 전달 예정)

- `printUpdateReport`(migrate.go ~1375, 갱신-통합 경로)의 구형 게이트(`failed+reversed > 0`)는 카드 범위 밖으로 의도 미수정 — 동일 게이트 적용 여부는 후속 결정
- 이동 0건+destination 좌초 실패만 있는 런은 요약 힌트 미출력(지시 게이트 결과) — JSON 행의 per-row journal 필드는 유지

## 카드-리뷰 정산 기록

- af-flow(운영자 지시 d-20261106T051044Z-7a4c)에 따라 codex_review 기반 card-review는 수행하지 않음 — Claude sync-auditor 1회가 그 자리를 대신함(이 기록이 card-review.md 부재의 이유). 3.2 카드당 codex_audit 1회 허용(d-20261106T055816Z-449d)은 감사관의 선택권으로 감사관 프롬프트에 명시

## sync 감사 1라운드 (2026-11-06 ~06:37Z)

- 감사관: sync-auditor (Claude, auditor-model glm-5.3), 1회 소환 (af-flow 체인)
- 판정: **FAIL** — harmonic 76/100, 기능성 방화벽(40% 축 <통과문턱) 발동. 판정 파일 `.moai/reports/t2094/verdict.md` (audited_sha 1852d1a34deec532f4e715fff2053920141e3cf0, 직독 확인)
- 점수: 기능성 62 FAIL / 보안 93 PASS / 장인정신 70 PASS / 일관성 88 PASS
- codex_audit 1회 소진(운영자 결정 d-20261106T055816Z-449d): 영수증 rcpt-78c868a9f2988feca76a0b71, 이 head 결합, codex 판정 fail 4건 → F1~F4로 독립 재판정 일치. **재호출 시 리더 경유 운영자 새 승인 필요 — t2094 잔여 허용량 없음**
- 차단 결함 (전부 in-diff):
  - F1 [High] migrate_oscwd_windows_test.go:142 — PowerShell 확장 문자열 안 `[Environment]::CurrentDirectory` 미평가(리터럴 출력) → Windows 러너 확정 적색
  - F2 [High] migrate_oscwd_windows_test.go:224 — bash `read -t`에 nil Stdin(null 장치) → 즉시 EOF, 가짜 실패
  - F3 [Medium] migrate.go:203 — 힌트 게이트 `moved>0||interrupted>0`가 좌초 사례(전진 절방+역이동 실패 → destination 잔류+pending 저널)에서 복구 힌트 상실 — 레인 판단: 카드 내 수리(수용 아님)
  - F4 [High] migrate_oscwd_windows_test.go:192 — cmd.exe만 kill, ping 손자 생존 → TempDir 정리 실패
- 선택 결함 (수리 범위 밖, 기록만): F5(printUpdateReport 구형 게이트, 범위 밖) / F6(재탐사 비용, 임의) / F7(WOW64 주석 과장, 임의)
- 준수 확인: PEB 오프셋 해석적 검증(CURDIR→CommandLine 연쇄 교차), F2 계약 구조 보장, 오류 보존 경로, 분류·exit 불변, POSIX 파일 미변경, 범위 5파일
- 결정 기록: decided_by=lane-31 sync-audit-r1 evidence_refs=.moai/reports/t2094/verdict.md ladder_path=Claude sync-auditor(af-flow) — FAIL→수리 사이클 진입(천장: 수리 후 재감사 최대 2회, 동일 감사관, codex 재호출 없음)

## 수리 사이클 1 (진행 중)

- 같은 manager-develop에 재위임: F1·F2·F4(테스트 확정 적색 수리) + F3(좌초 실패 시 힌트 유지 — 이동>0‖중단>0‖좌초>0 게이트로 확장, 분류·exit 불변), 결함 원문 소스 대조 뒤 구현, 전방 커밋

## 수리 사이클 1 결과 (2026-11-06 ~07:05Z)

- 커밋: `3db7aea4f5a9465c1f6771cb69a86fe3d5746e86` — `[t2094] fix(worktree): repair sync-audit round-1 findings F1-F4 (card t2094)` (parent 1852d1a34d, 3파일 +76/−6, 브랜치 위 전방 커밋, 트리 깨끗함 — 직독 확인)
- F1: PowerShell 확장 문자열 안 표현식 → 서브식 보간 `win32=$([Environment]::CurrentDirectory)`로 수리 (구현원 소스 대조 뒤)
- F2: bash 자식 Stdin nil(null 장치→즉시 EOF) → os.Pipe로 쓰기 끝을 t.Cleanup까지 홀드
- F4: cmd.exe 단일 kill → `taskkill /T /F /PID` 트리 kill(Kill 폴백)
- F3(레인 판단 이행): `treeOutcome.Stranded`(reverseAfterFailure !movedBack·위치 불명 분기에서 설정) → `migrationReport.stranded` 집계 → 게이트 `moved>0 || interrupted>0 || stranded>0`. 분류·exit 불변. 게이트 단위 표에 stranded 행 신설, 행동 시험 TestApply_StrandedTreeKeepsTheRollbackHint 신설, 바쁜-트리 무힌트 케이스는 불변 PASS

### reflog 해석 정정 (레인 오독 기록)

- 레인은 reflog의 `reset: moving to HEAD` 2건을 `git reset --hard`로 해석해 "미커밋 수리분 폐기"로 기록했으나, 구현자 처리 기록(정정): `git stash push -u -m t2094-repair-wip-verify2` → detached base 비교 → 복귀 → `git stash apply` → diff 검증 뒤 커밋. 수정분 유실 없음, 태그 스태시-apply는 하네스가 허용하는 절차. `git reset`(mixed)도 동일 reflog 문구를 남기므로 --hard 단정은 과독이었음 — 구현 기록에 정정 반영

### 광역 런 실패 1건 — 선행 결함(수리 비유발) 판정

- `TestApply_PendingIntentButNeverMoved_IsSettledThenMoved`: base 2a04549ec7에서 -count=8 → 6 FAIL/2 PASS, 1852d1a34d에서도 FAIL, 수리 트리에서 -count=3 → 1 PASS/2 FAIL — 초 단위 벽시계 의존 flake, 카드 이전부터 존재
- 원인(측정, 1852d1a34d 계측): `appendMoveListLine` dedupe(movelist.go:494-497)가 `revLines(...)[0]`과 비교하는데 revLines(movelist.go:514-523)는 "last first" 문서와 달리 **정방향** 반환 → dedupe가 첫 행과 비교. pending 선과 재실행 intent 선이 같은 초에 기록되면 바이트 동일 → intent append가 무시됨
- 수리 범위 밖(이열거 수정 시 무브리스트 동사 전반 동작 변경) — **리더 전달: 후속 카드 후보**(revLines 순서 vs 문서, 또는 dedupe 비교 기준). 레인은 큐 변경 불가로 기록만

### 레인 재검증 (HEAD 3db7aea4f5 직접 실행)

- `GOOS=windows GOARCH=amd64 go vet ./internal/cli/worktree/` → win-vet-exit=0
- `GOOS=windows GOARCH=amd64 go test -c -o /dev/null ./internal/cli/worktree/` → win-compile-exit=0
- `go test ... -run 'TestApply_BusyTree|TestApply_AllFailures|TestApply_StrandedTree|TestMigrateSummary|TestApply_FailedTrees' -count=1` → ok 11.477s
- 재사용 귀속: 구현원 실행분(같은 트리 키) — darwin vet 0, win-386 vet 0, gofmt 무출력, 6시험 -v 전부 PASS 15.148s

## 영수증 보강 + 사슬 재개 (2026-11-06 ~07:26Z)

- 리더 판정: 정리안 **(a)** 채택 — 감사 대상 cd7b5803bd(영수증 rcpt-affc80b4b23c072c0447fef3 결합), deb65b740f는 판정 파일만 추가한 커밋, 신규 Codex 호출 없음. complete는 v3.2.1-rc.1 설치 확인 뒤(t2098은 이미 main 착지)
- 2차 codex_audit(de2d 재감사 조항): rcpt-affc80b4b23c072c0447fef3, 수리 델타 codex 판정 pass·결함 0. 최종 verdict.md는 audited_sha를 cd7b5803bd로 재고정(정리안 (a) 이행), `verdict: PASS`·`audited_sha:` 각 1줄 (직독 확인)
- 8859 금지 절차(d-20261103T000836Z-af10): 이번 세션 미사용 확인 — 전 갱신 renewed, 만료 재임대 호출 없음
- merge-ready 1차 거절 원문: `Factory stage: card evidence refused: no sync-audit verdict file under /Users/goos/.moai/worktrees/projects/moai-adk-go/t2094/.moai/reports/t2094.`
- 원인(소스 직독, internal/homestate/card_evidence_readers.go:100): 판정 파일 글롭이 `<phase>*.md`=`sync-audit*.md` — `verdict.md`는 미매칭. 또한 audited_sha가 기록 증거 커밋의 접두사여야 함(T5/T11) — 기록 deb65b740f vs 판정 audited cd7b5803bd 불일치 → 리더 예고대로 rebind-evidence 경로
- 조치: 디스크 개명 verdict.md → sync-audit-iter2.md(-iter 최댓값 규칙) + 트리 청결을 위해 개명 커밋 → `moai factory rebind-evidence t2094 --reason ...` (기록 증거 deb65b740f→cd7b5803bd, 감사 행에 사유 원문 기록) → merge-ready 재시도

### Gaps

- 감시 측정 명령의 `find … -newer /tmp` 세그먼트가 워크트리 가드에 거부됨(외부 경로 참조). 같은 채널을 `stat`/`ls -lT` 직독으로 확보했으므로 결정에 미반영 — 거부 자체는 이 절에 기록(§3.1 refused-tool degradation).
