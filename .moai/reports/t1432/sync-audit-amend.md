auditor-model: claude-opus-5-5
verdict: PASS-WITH-DEBT
audited_sha: 89b4e75ea83069ee60350e2a47612db5e3f7ee5e

# 델타 sync-audit — SPEC-HARNESS-RETENTION-HARDEN-001 amendment 0.4.1 (card t1432, M7–M10)

> 반출 주의: 이 파일은 지정 경로 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a3712f21dbd667f41/.moai/reports/t1432/sync-audit-amend.md` 에 쓰려 했으나, 감사 세션이 `.claude/worktrees/develop` 에 격리돼 있어 Write 가 거부됐다("Edit the worktree copy of this file instead of the shared-checkout path"). 대상 트리 소유자(리더)가 이 파일을 그 경로로 그대로 복사해야 반출이 완결된다.

## Evaluation Report
SPEC: SPEC-HARNESS-RETENTION-HARDEN-001 (amendment 0.4.1)
Overall Verdict: PASS-WITH-DEBT — 가중 점수 91.4/100, 차단 결함 0건

감사 범위: `f040ffcd3..89b4e75ea` 의 4개 커밋(46485ec48 M7 RED, 84ea58d48 M8 heal lock, aa42a4398 M9 테스트 부채, 89b4e75ea M10 판정서·진행 기록·CHANGELOG). 대상 트리: `.claude/worktrees/agent-a3712f21dbd667f41`(브랜치 `worktree-agent-a3712f21dbd667f41`). 리더 결정(부하 측정 5 s 초과, prune 단독이 이미 7.7–14 s)은 재론하지 않았다.

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 93/100 | PASS | `go test -count=1 -race -run Heal -v ./internal/harness/` → 11개 테스트 `--- PASS`, `ok ... 3.865s`; `-count=5 -race -run Heal` → `ok ... 13.528s`; 전체 패키지 `go test -count=1 -race -cover ./internal/harness/` → `ok ... 46.623s coverage: 87.0% of statements` |
| Security (25%) | 90/100 | PASS | 적대 엔트리 4종(symlink/directory/fifo/not-owned) 전부 `--- PASS`; 독립 변이 `noowner` 사살(`retention_heallock_test.go:308: want an error naming .../usage-log.jsonl.prune-heal, got <nil>`); `follow` 변이(O_NOFOLLOW 제거)는 생존 — F2 |
| Craft (20%) | 89/100 | PASS | 독립 변이 7종 중 5종 사살, 생존 2종은 SPEC 이 인정한 한계(relbefore — 프로브가 사살) + F2; 커버리지 87.0%; `gofmt -l internal/harness/` 빈 출력; `go vet` darwin·windows exit 0 |
| Consistency (15%) | 91/100 | PASS | `//go:build !windows` + windows 쌍둥이가 같은 심볼 정의; 오류 래핑 `%w`·`[WARN] harness/retention:` 접두 규약 유지; `git diff --stat 1e2151a38 89b4e75ea -- internal/lockfile` 빈 출력 |

가중: 0.40×93 + 0.25×90 + 0.20×89 + 0.15×91 = 91.4. must-pass(Functionality, Security) 모두 통과.

### Findings (structured defect-list)
- F1 [Low, 신뢰도 중] [optional] `internal/harness/retention_heal_unix.go:71` — heal-lock 엔트리 소유자 확인이 연 디스크립터(fstat)가 아니라 경로(`ownerCheck(path)` → `os.Lstat`)를 다시 읽는다. 열기와 확인 사이에 경로가 바뀌면 확인 대상과 잠근 inode 가 달라질 수 있다(로그 디렉터리 쓰기 권한자가 전제). Required fix(선택): `f.Stat()` 의 `Stat_t.Uid` 로 판정하거나, 확인 뒤 `os.Lstat(path)` 와 `os.SameFile` 재대조.
- F2 [Low, 신뢰도 높음] [optional] `internal/harness/retention_heal_unix.go:55` — 기존 정규 파일 open 의 `O_NOFOLLOW` 를 제거한 변이가 heal-lock 테스트 전체를 통과(`== follow: exit=0 ... ok`). 앞선 `Lstat` 정규 파일 판정과 뒤의 `os.SameFile` 이 막아 주므로 실해는 "링크 대상을 한 번 여는 것"(쓰기·절단 없음, O_NONBLOCK 이라 FIFO 무정지)에 그친다. REQ-HRH-016 "never follow a link" 는 정적 읽기로만 확인됨. Required fix(선택): Lstat 과 open 사이에 symlink 를 끼우는 결정적 seam 테스트, 또는 판정서 Gaps 에 "TOCTOU 링크 추종은 테스트 미고정" 명기.
- F3 [Low] [optional, 기존 부채 D2] `internal/harness/retention.go:281` — 생성·열기·잠금 실패 경로의 경고 줄은 코드상 출력되지만(읽기 확인) 테스트로 고정되지 않음. 판정서 A8 이 D2 미이행을 공시.
- F4 [Low] [optional] `CHANGELOG.md:14` — "closes the window" 문장에 판정서 A9 의 잔여(락을 지키지 않는 구 설치 바이너리가 업그레이드 창 동안 여전히 경합)가 빠져 있다. 0/20000 근거는 정당하나 범위는 "이 빌드끼리"다. Required fix(선택): "among processes running this build" 수준의 한정구 추가.
- F5 [Info] [optional] SPEC 명시 한계 재확인 — "제거 전에 락 해제" 변이(`relbefore`)는 결정적 테스트를 통과하고 레이스 프로브만 잡는다. 이번 감사에서 프로브가 실제로 사살함을 관측(C5). 조치 불요.

- F6 [Medium, 신뢰도 높음 — codex 런타임 재현, 감사자는 코드 읽기로 확인] [optional, 리더 판단 요청] `internal/harness/retention_heal_unix.go:77-83` — 기한 검사가 `flock` 실패 뒤에만 있어, 대기 중 프로세스가 멈췄다가 기한이 지난 뒤 재개돼 잠금을 얻으면 timeout 없이 heal 을 진행한다. codex adversarial 재현: `SIGSTOP`→2 s 후 해제→`SIGCONT`, `accepted expired acquisition: elapsed=2.319004916s err=nil`. REQ-HRH-016 "not acquired within the bounded wait → 제거하지 않음" 의 문자상 이탈이다. 상호 배제(정확성)는 깨지지 않고 결과는 heal 이 상한보다 늦게 진행되는 것뿐이라 optional 로 분류했으나, SPEC 문구 이탈이므로 리더가 blocking 으로 올릴 수 있다. Required fix: 획득 성공 직후 `time.Now().After(deadline)` 이면 LOCK_UN·close 후 timeout 오류 반환 + 재현 회귀 테스트.

### 교차 감사 (codex adversarial, baseBranch, project_root=대상 트리)
- verdict `fail`, 결함 1건(P2 = F6). `audit_receipt` 미발급(이 트리의 codex 게이트가 required 가 아님) → receipts=none.
- codex 가 관측한 브랜치 팁은 `8425b59a7` 이다. `git rev-parse worktree-agent-a3712f21dbd667f41` → `8425b59a773e9a5f4994c9e0c013417838e11384`; `git log 89b4e75ea..8425b59a7` → `8425b59a7 test(...): bound the blocked-pruner FIFO release (card t1432)`, `retention_tail_test.go` 60+/10-. **감사 중 감사 대상 트리에 외부 커밋이 들어왔다 — 프로세스 결함으로 리더에게 보고.** 이 감사의 판정은 `89b4e75ea` 기준이며, 뒤쪽 Go 실행 일부는 디스크의 새 테스트 파일을 읽었을 수 있다(운영 코드는 동일).

### Recommendations
- F1·F2 는 같은 함수의 TOCTOU 다듬기라 한 번에 처리 가능. 둘 다 이 카드를 막지 않는다.
- F4 는 CHANGELOG 한 구절 수정으로 끝난다.

---

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. heal lock 은 heal 을 직렬화한다: 배타 flock(`LOCK_EX|LOCK_NB`), 모든 open 에 `O_NOFOLLOW|O_NONBLOCK`, 10 ms 폴링·2 s 상한, 재검사+제거 구간에만 보유, 모든 경로에서 해제.
2. 적대 heal-lock 엔트리는 상태 파일·stamp·로그를 건드리지 않고 건너뛴다.
3. Windows 쌍은 도달 불가.
4. 테스트가 주장한 변이를 죽인다.
5. CHANGELOG "closed" 는 0/20000 프로브로 뒷받침된다.
6. RED 커밋이 수정 커밋보다 그래프상 앞선다.

### Evidence

**C1 — 코드 읽기** (`retention_heal_unix.go:34-94`, `retention.go:273-289`): `syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)`(:77), EWOULDBLOCK/EINTR 외 오류는 close 후 반환(:84-87), `time.Now().After(deadline)` 이면 close 후 타임아웃 오류(:88-91), 그 밖엔 `time.Sleep(pruneHealPoll)`; `pruneHealWait = 2 * time.Second`, `pruneHealPoll = 10 * time.Millisecond`. `healStateEntry` 는 `defer release()` 뒤 `removeStateEntryIfUnchanged` 만 수행하고 반환 — 교체 파일 생성(`openStateFile` 루프)은 락 밖. 상태 경로를 제거하는 주체는 락 보유자뿐이고 비-heal 경로는 `O_CREATE|O_EXCL` 생성만 하므로, 제거끼리의 상호 배제로 창이 닫힌다. inode 재사용은 mode·mtime 대조가 추가로 막는다(symlink↔regular, 0400↔0644).

```
$ go -C <tree> test -count=1 -race -run Heal -v ./internal/harness/
--- PASS: TestPruneHealLockHeldPastTheBoundFailsClosed (2.00s)
--- PASS: TestPruneHealLockHostileEntryFailsClosed (0.01s)
--- PASS: TestPruneCommonPathCreatesNoHealLock (0.00s)
--- PASS: TestHealRemovalIsNotDecidedByModTimeAlone (0.00s)
--- PASS: TestHealDoesNotRemoveAFreshStateFile (0.00s)
--- PASS: TestPruneHealFileArmKeepsAFreshStateFileSwappedInDuringTheHeal (0.01s)
--- PASS: TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal (0.01s)
--- PASS: TestPruneCommonPathIgnoresAHeldHealLock (0.01s)
--- PASS: TestPruneHealLockIsFreeWhilePrunerIsInItsArchiveStep (0.31s)
--- PASS: TestPruneHealSerializesOnTheHealLock (0.31s)
--- PASS: TestPruneHealWaitsForASharedHolder (0.31s)
ok  	github.com/modu-ai/moai-adk/internal/harness	3.865s
$ go -C <tree> test -count=5 -race -run Heal ./internal/harness/
ok  	github.com/modu-ai/moai-adk/internal/harness	13.528s
$ go -C <tree> test -count=1 -race -cover ./internal/harness/
ok  	github.com/modu-ai/moai-adk/internal/harness	46.623s	coverage: 87.0% of statements
```
위임 프롬프트의 `-run 'Heal|HRH'` 중 `HRH` 는 일치 테스트가 없다(`testing: warning: no tests to run`) — 테스트 이름에 HRH 가 없어서이며 결함 아님.

**C2** — `TestPruneHealLockHostileEntryFailsClosed` 네 하위 테스트가 상태 경로 `SameFile`, 링크 피해자 바이트, 로그 바이트, 경고 정확히 1줄, heal-lock 엔트리 무변경을 단언하고 PASS(위 출력). symlink 하위 테스트는 heal-lock 링크 대상(victim2) 무변경도 단언. 상태 경로가 곧 stamp 파일이므로 stamp 무변경도 포함.

**C3** — `retention_owner_windows.go:8` `func entryOwnedByCurrentUser(string) bool { return false }`; `retention.go:90` `ownerCheck: entryOwnedByCurrentUser` 가 유일한 운영 대입(필드 비공개); `healStateEntry` 는 :274 에서 소유 검사 실패 시 `acquireHealLock`(:279) 이전에 반환. `env GOOS=windows go -C <tree> vet ./internal/harness/ ./internal/lockfile/` → 출력 없음, exit 0. (정적 읽기 근거; Windows 런타임 미관측.)

**C4 — 독립 변이 실행** (스크래치 overlay, `-run "TestPruneHeal|TestPruneCommonPath|TestPruneStateRemovalFailureInReadOnlyDirSkips|TestPruneStateUnreplaceableInReadOnlyDirSkips|TestHeal"`):
```
== sh (LOCK_EX→LOCK_SH): exit=1
    retention_heallock_test.go:432: the pruner returned (<nil>) while another descriptor held a shared heal lock: it did not request an exclusive lock
== noowner (heal-lock 소유 검사 무력화): exit=1
    retention_heallock_test.go:308: want an error naming .../usage-log.jsonl.prune-heal, got <nil>
== follow (기존 파일 open 의 O_NOFOLLOW 제거): exit=0
    last: ok  	github.com/modu-ai/moai-adk/internal/harness	3.155s
== norelease (defer release() 제거): exit=1
    retention_heallock_test.go:532: the heal lock is still held after the prune returned: resource temporarily unavailable
    retention_heallock_test.go:346: the heal lock is held while the pruner is in its archive step: resource temporarily unavailable
== blindpoll (pruneHealPoll = 2s): exit=1
    retention_heallock_test.go:452: the pruner returned 1.706621583s after the shared heal lock was released, want under 1s: it did not poll for the lock
    retention_heallock_test.go:157: the pruner returned 1.704919625s after the heal lock was released, want under 1s: it did not poll for the lock
== relbefore (제거 전에 release): exit=0
    last: ok  	github.com/modu-ai/moai-adk/internal/harness	3.351s
== nolock (flock 호출을 nil 오류로 대체): exit=1
    retention_heallock_test.go:207: want an error naming .../usage-log.jsonl.prune-heal, got <nil>
    retention_heallock_test.go:116: the pruner returned (<nil>) while another descriptor held the heal lock: it did not wait
```
판정서 A4 의 주장(LOCK_SH 는 b2 만 사살, never-released 는 AC-HRH-003 c 가 사살, blind sleep 은 M9 후 사살)과 일치. 이번 실행에서는 (d) 도 norelease 를 잡았다(판정서는 finalizer 의존이라 미의존으로 기록 — 정직한 기록).

**C5 — 프로브 재실행** (커밋된 `amend-drafts/zz_heal_race_probe_test.go`, 이 트리 경로로 overlay 재생성):
```
$ go -C <tree> test -count=1 -v -timeout 300s -overlay .../probe.json -run TestZZProbeHealWindow ./internal/harness/
    zz_heal_race_probe_test.go:97: PROBE heal-window: trials=20000 heal_removed_the_swapped_in_fresh_entry=0 path_holds_F=20000 path_holds_other=0 heal_errors=0
--- PASS: TestZZProbeHealWindow (53.92s)
$ (같은 프로브 + relbefore 변이 overlay)
    zz_heal_race_probe_test.go:97: PROBE heal-window: trials=894 heal_removed_the_swapped_in_fresh_entry=5 path_holds_F=889 path_holds_other=0 heal_errors=0
```
0/20000 재현, 같은 프로브가 "제거 전 해제" 변이에서 5/894 를 세므로 실패 가능한 프로브임(양성 대조). 판정서 A5 의 0/20000 기록과 CHANGELOG 문장이 이와 일치.

**C6 — 커밋 그래프**:
```
$ git log --format='%H %P %s' --stat f040ffcd3..89b4e75ea
46485ec48 (parent f040ffcd3) test: observed RED baseline — retention_heallock_test.go(new), retention_owner_test.go, red-baseline-amend.md (운영 코드 변경 없음)
84ea58d48 (parent 46485ec48) fix(harness): heal lock — retention.go, retention_heal_unix.go, retention_heal_windows.go, .gitignore
aa42a4398 (parent 84ea58d48) test: N1, N2, N4 and the poll bound
89b4e75ea (parent aa42a4398) docs: verification and close-out
```
RED 독립 재현(46485ec48 의 `retention.go`·테스트 3종 overlay, heal 파일 2종 삭제):
```
--- FAIL: TestPruneHealLockHeldPastTheBoundFailsClosed (0.04s)
--- FAIL: TestPruneHealLockHostileEntryFailsClosed (symlink/directory/fifo/not-owned 전부)
--- FAIL: TestPruneHealSerializesOnTheHealLock (0.01s)
    retention_heallock_test.go:116: the pruner returned (<nil>) while another descriptor held the heal lock: it did not wait
--- FAIL: TestPruneHealWaitsForASharedHolder (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/harness	1.240s
```

**기타**: `gofmt -l internal/harness/` 빈 출력; `go -C <tree> vet ./internal/harness/ ./internal/lockfile/` exit 0; `git diff --stat 1e2151a38 89b4e75ea -- internal/lockfile` 빈 출력(REQ-HRH-012).

### Baseline-attribution
Go 측정은 모두 이 감사 세션에서 `go -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a3712f21dbd667f41` 로 그 트리(HEAD `89b4e75ea83069ee60350e2a47612db5e3f7ee5e`)의 디스크 내용을 대상으로 실행. 도구는 저장소 자체 빌드가 아니라 go 툴체인(go1.26.8)이라 §2.2 빌드 귀속 대상 아님. git 조회는 공유 객체 DB 를 이 세션의 develop 워크트리에서 SHA 로 읽었다. darwin arm64.

### Gaps
- 보고서를 지정 경로에 쓰지 못함(워크트리 격리 거부) — 스크래치에 저장, 리더 복사 필요.
- `golangci-lint run` 미측정: 대상 트리를 cwd 로 둘 수 없어(가드가 `cd`·`git -C`·복합 명령 거부, golangci-lint 는 `-C` 없음) `typechecking error: ... outside main module` 로 실패. 판정서 A6 의 `0 issues.` 를 대체 근거로 쓰지 않는다.
- 대상 트리의 `git status` 를 직접 관측하지 못함(가드 거부). 커밋 내용은 SHA 로 읽었고 Go 측정은 디스크 워킹 트리를 읽으므로, 미커밋 변경이 있었다면 측정 대상이 HEAD 와 다를 수 있다(리더 보고상 HEAD 89b4e75ea).
- MOAI_KANBAN 등 환경 변수 세척 없이 실행(복합 명령 거부).
- Linux·Windows 런타임 미관측. 부하 측정(DoD 15)은 재측정 안 함(리더 결정 사안).
- M9 의 N1/N2/N4 변이는 재실행하지 않음 — 해당 테스트는 전체 패키지 실행에서 PASS 만 관측.
- 교차 모델 감사(`audit_multi`) 미실행 — 위임 프롬프트가 요구하지 않았고 card_id 바인딩 지시도 없었다.

### Residual-risk
- 락을 지키지 않는 구 설치 바이너리가 업그레이드 창 동안 제거 경합을 낼 수 있다(판정서 A9; CHANGELOG 미기재 — F4).
- F1/F2 의 TOCTOU 는 로그 디렉터리 쓰기 권한자 전제이며 테스트로 고정되지 않았다.
- 타이밍 단언(300 ms, 1 s, 4.5 s)은 고부하 호스트에서 흔들릴 수 있다 — 이번 실행은 모두 통과.
- Leader process defect: an implementer commit (8425b59a7, test-only) landed while this audit ran; verdict is anchored to 89b4e75ea; production code identical.

## Iteration history
- 1회차(이 문서): 델타 M7–M10 최초 감사, PASS-WITH-DEBT 91.4.
