auditor-model: claude-opus-5-5

# t2081 sync 감사 보고서 (독립 감사, sync-auditor)

verdict: FAIL
audited_sha: 3be3e51f36402c6ab8e01f7958b21f6f9cd2ab77

- 카드: t2081 (병렬 `git worktree move`, cap 4)
- 감사 대상 커밋: `3be3e51f36402c6ab8e01f7958b21f6f9cd2ab77` (`git rev-parse HEAD`로 확인)
- 카드 자체 변경 범위: `23f0568bc6..HEAD`의 `internal/cli/worktree/migrate.go`, `internal/config/defaults.go`, `internal/cli/worktree/migrate_parallel_t2081_test.go`
- 종합 판정: FAIL (BLOCKING 2건, 둘 다 재현 증거 있음. 하나는 Codex도 독립적으로 지적)
- 점수(가중): Functionality 60 (FAIL) · Security 90 (PASS) · Craft 62 · Consistency 85 → must-pass 방화벽(Functionality)에 걸려 전체 FAIL
- Codex 영수증: `rcpt-8b29f47630e6e68e12831a06` (호출 1회, 재시도 없음)

## 1. Claim (주장)

1. 설계 자체(배치 준비 → `git worktree move`만 동시 실행 → 계획 순서대로 한 트리씩 마무리)는 의도대로 이동 호출만 겹친다. 저널 flush, `appendMoveListLine`, `hook.RekeyWorktreeEntry`, 카드 재바인딩, 이동 후 매니페스트 비교는 모두 고루틴 밖, 메인 루프에서 계획 순서로 실행된다. 고루틴이 쓰는 공유 상태는 각자의 `movedAhead.moveErr` 하나뿐이고 `wg.Wait()` 뒤에 읽히므로 엔진 코드 자체에는 경합이 없다.
2. 배치 폭 1은 예전의 엄격한 직렬 순서와 같다(준비 1 → 이동 1 → 마무리 1). `TestMigrationEngine_ConcurrencyOneIsTheOldSerialRun`이 이를 고정하고 통과한다.
3. 그러나 이동 seam(`forwardGitMove`)이 이제 동시에 호출된다는 계약 변화 때문에 기존 테스트 하나가 `-race`에서 결정적으로 데이터 경합을 낸다(F1). CI의 race remainder 잡은 이 패키지를 포함하므로 CI가 빨갛게 된다.
4. 카드가 새로 넣은 복구 증명 테스트 하나가 `-race`에서 간헐적으로 실패한다(5회 중 2회). 재실행이 마이그레이션 잠금에 막혀 아무것도 옮기지 않는다(F2). "재실행이 수렴한다"는 증명이 신뢰할 수 있게 서 있지 않다.
5. progress.md의 수치(23%/37% 단축)는 내부 산술이 맞고 "짝끼리만 비교 가능"이라는 단서가 정직하게 달려 있다. 다만 "증명 테스트가 CI의 OS 매트릭스에서 돈다"는 Gap 문장은 사실과 다르다(F3).

## 2. Evidence (명령 + 출력 원문)

### 2.1 HEAD 확인

```
$ git rev-parse HEAD
3be3e51f36402c6ab8e01f7958b21f6f9cd2ab77
$ git rev-parse --show-toplevel
/Users/goos/.moai/worktrees/projects/moai-adk-go/t2081
```

### 2.2 카드 테스트, `-race` (Functionality)

```
$ go test ./internal/cli/worktree -run 'TestGitWorktreeMove_|TestMigrationEngine_' -count=1 -race -v
--- PASS: TestGitWorktreeMove_TheConsistencyCheckSeesADamagedRegistration (1.28s)
    migrate_parallel_t2081_test.go:222: 12 trees x 8 rounds of concurrent moves with concurrent readers: git's view stayed consistent every round
--- PASS: TestGitWorktreeMove_ConcurrentMovesOfDifferentTreesAreConsistent (13.94s)
--- PASS: TestGitWorktreeMove_TheSameTreeMovedTwiceAtOnceFailsForOne (0.89s)
--- PASS: TestMigrationEngine_MovesABatchTogetherAndFinishesInPlanOrder (4.89s)
--- PASS: TestMigrationEngine_ConcurrencyOneIsTheOldSerialRun (5.28s)
    migrate_parallel_t2081_test.go:430: engine: migration lock: /private/var/folders/kt/.../mig-repo/.moai/state/worktree-migrate/lock is held by another mover; this run moves nothing
--- FAIL: TestMigrationEngine_StoppedAfterAWholeBatchIsPreparedThenRerunMovesEverything (3.50s)
--- PASS: TestMigrationEngine_StoppedAfterABatchMovedThenRollbackRestoresEveryTree (4.35s)
--- PASS: TestMigrationEngine_WorktreeListTimeoutIsAnError (5.22s)
--- PASS: TestMigrationEngine_AnnouncesBeforeThePlanRuns (3.10s)
--- PASS: TestMigrationEngine_ReadsManifestsAheadButBoundedAndMovesInPlanOrder (4.81s)
--- PASS: TestMigrationEngine_AManifestReadAheadStillMustMatchAfterTheMove (2.85s)
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	51.874s
```

(임시 디렉터리 경로 중간만 `...`로 줄였다.)

### 2.3 F2 재현 빈도

```
$ go test ./internal/cli/worktree -run 'TestMigrationEngine_StoppedAfterAWholeBatchIsPreparedThenRerunMovesEverything' -count=3 -v
--- PASS: ... (5.23s)
--- PASS: ... (4.79s)
--- PASS: ... (5.66s)
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	16.368s

$ go test ./internal/cli/worktree -run 'TestMigrationEngine_Stopped' -count=4 -race
--- FAIL: TestMigrationEngine_StoppedAfterAWholeBatchIsPreparedThenRerunMovesEverything (4.93s)
    migrate_parallel_t2081_test.go:430: engine: migration lock: /private/var/folders/kt/.../lock is held by another mover; this run moves nothing
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	43.496s
```

집계: `-race` 없이 3/3 통과, `-race`에서 5회 중 2회 실패(2.2의 1회 + 2.3의 4회 중 1회). `TestMigrationEngine_StoppedAfterABatchMovedThenRollbackRestoresEveryTree`는 `-race` 5회 모두 통과.

### 2.4 F1 재현 (Codex 지적 검증)

```
$ go test ./internal/cli/worktree -run 'TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending' -count=3 -race
WARNING: DATA RACE
Read at 0x00c0000124f8 by goroutine 124:
  ...TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending.func1()
      .../internal/cli/worktree/migrate_hardening_d_t2053_test.go:100 +0x1dc
  ...runMigrationEngineWith.func1.2()
      .../internal/cli/worktree/migrate.go:1018 +0xe0
Previous write at 0x00c0000124f8 by goroutine 123:
  ...migrate_hardening_d_t2053_test.go:100 +0x268
  ...migrate.go:1018 +0xe0
Goroutine 124 (running) created at:
  ...runMigrationEngineWith.func1()   .../migrate.go:1016 +0x57c
--- FAIL: TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending (2.33s)
    testing.go:1712: race detected during execution of test
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	8.394s

$ go test ./internal/cli/worktree -run 'TestMigrateApply_Forward|TestApply_ReportRowsPrecedeTheError' -count=1 -race
... (같은 DATA RACE, 같은 두 줄 위치)
--- FAIL: TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending (2.12s)
    testing.go:1712: race detected during execution of test
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	16.155s
```

원인 코드(`migrate_hardening_d_t2053_test.go:92-100`): seam 대체 함수가 잠금 없이 공유 슬라이스에 `restore = append(restore, ...)`를 한다. 두 트리 세계(`t2053TwoTreeWorld`)의 두 이동이 이제 한 배치에서 동시에 호출된다. 같은 묶음의 나머지 forward 실패 테스트는 경합 없이 통과했다.

CI 포함 여부(`.github/workflows/ci.yml:687-689`): race remainder 잡은 `internal/cli`와 rosterguard를 뺀 모든 패키지를 `go test -json -race -count=1`로 돈다. `internal/cli/worktree`가 포함된다.

### 2.5 Craft / Consistency

```
$ gofmt -l internal/cli/worktree internal/config
(출력 없음)
$ go vet ./internal/cli/worktree ./internal/config
(출력 없음)
```

### 2.6 Security (grep 기반 점검)

변경은 내부 CLI 엔진의 순서 재배치와 상수 하나이며, 새 외부 입력·비밀·셸 문자열 조립이 없다. 이동 인자는 기존과 같이 `gitOutputWithin(... "worktree", "move", from, to)`로 argv 배열로 넘어간다(`migrate.go:1357-1360`, `2204-2208`). 새 테스트의 환경 변수 `MOAI_T2081_STRESS_TREES`/`ROUNDS`는 `strconv.Atoi` 후 양수일 때만 쓰인다. High/Critical 소견 없음.

### 2.7 중첩 트리 도달 가능성 확인 (F7 전제)

```
$ grep -rlE "/\.moai/worktrees/[^/]+/\.claude/worktrees/" /Users/goos/MoAI/moai-adk-go/.git/worktrees --include=gitdir
(출력 없음)
```

현재 이 저장소에는 중첩 등록이 없다. F7은 원리상 도달 가능하다는 소견이며, 이 저장소에서 지금 발생한다는 주장이 아니다.

### 2.8 CI OS 범위 (F3 전제)

```
$ grep -n "os: \[\|runs-on:\|matrix:\|go test" .github/workflows/ci.yml
110:    runs-on: ${{ matrix.os }}
117:        os: [ubuntu-latest]
...
788:        os: [ubuntu-latest, macos-latest, windows-latest]
812:          go test -json -tags=integration -race -timeout 180s ./test/integration/harness/... > test-stream.json || rc=$?
$ grep -n "os: \[\|go test\|on:\|branches" .github/workflows/release-pr-multi-os.yml
14:    branches: [main]
125:        os: [ubuntu-latest, macos-latest, windows-latest]
407:            echo "::notice::Intentional exclusion: non-release PR (detect-release skipped by design) — matrix not required"
```

카드 PR의 단위 테스트는 ubuntu에서만 돈다. 3-OS 매트릭스는 `test/integration/harness`(ci.yml)와 릴리스 PR(release-pr-multi-os.yml)에만 있다.

### 2.9 Codex (1회 호출)

호출: `mcp__moai__codex_audit`, `project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2081`, `target=baseBranch`, `delta_base=23f0568bc6f77d442008a4fac824f8dd2d4f5ed9`, `mode=native`.

반환 요약: 판정 `fail`, 소견 1건 [P2] — `TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending`의 `forwardGitMove` 대체 함수가 공유 `restore` 슬라이스에 잠금 없이 append하므로, 이제 두 트리의 함수가 동시에 호출되어 데이터 경합과 복구 콜백 유실이 생길 수 있다; mutex를 넣거나 그 테스트의 이동 동시성을 1로 제한하라(위치 `internal/cli/worktree/migrate.go:1016`). Codex 스스로 정적 검토이며 테스트를 실행하지 않았다고 밝혔다. 이 감사가 2.4에서 실행으로 확인했다.

`audit_receipt`: `rcpt-8b29f47630e6e68e12831a06`; `review_base`: `delta_base (merge base 23f0568bc6f77d442008a4fac824f8dd2d4f5ed9)`; resolved_model `gpt-6.1-sol`; `gate_unmet` 없음.

## 3. Baseline-attribution (무엇을 기준으로 쟀나)

- 모든 테스트·vet·gofmt는 이 실행에서, 이 워크트리(`3be3e51f36`)의 소스 트리에 대해 Go 툴체인(go1.26.8 darwin-arm64, race 출력의 경로로 확인)으로 직접 실행했다. 빌드 캐시는 트리 밖이다.
- 코드 판단의 범위는 `git diff 23f0568bc6 HEAD -- internal/cli/worktree/migrate.go internal/config/defaults.go`와 새 테스트 파일이다. 같은 범위의 diffstat에는 main에서 흡수된 다른 카드 파일(t2063~t2069의 hook·factorymsg 테스트 등)도 섞여 있는데, 이는 23f0568bc6가 그 main 커밋들을 품고 있지 않기 때문이며 이 감사의 대상이 아니다. Codex의 delta 리뷰(`23f0568bc6...HEAD`)에도 같은 이유로 그 파일들이 포함됐다.
- MCP 서버 빌드는 `646955327e`로 트리 HEAD의 조상이다(결과의 `build_lag` 안내). 이 감사는 moai CLI 결과를 증거로 쓰지 않았고, Codex 리뷰는 트리의 git diff를 읽으므로 판정에 영향이 없다고 본다.

## 4. 소견 (structured defect list)

| ID | 심각도 | 위치 태그 | 요약 | 위치 |
|---|---|---|---|---|
| F1 | BLOCKING | in-diff + 재현 | 이동 seam이 동시에 호출되면서 기존 테스트 `TestMigrateApply_ForwardFailure_UnreadableDestinationIdentity_StaysPending`가 `-race`에서 결정적으로 데이터 경합 실패. CI race remainder 잡이 이 패키지를 돈다. (Codex 동일 지적) | `internal/cli/worktree/migrate.go:1013-1021`, `internal/cli/worktree/migrate_hardening_d_t2053_test.go:92-100` |
| F2 | BLOCKING | in-diff + 재현 | 새 증명 테스트 `TestMigrationEngine_StoppedAfterAWholeBatchIsPreparedThenRerunMovesEverything`가 `-race`에서 5회 중 2회 실패: 재실행이 "lock is held by another mover"로 아무것도 옮기지 않음. 재실행 수렴 증명이 불안정하고 CI에서 깜빡일 수 있다. | `internal/cli/worktree/migrate_parallel_t2081_test.go:398-439`, `migrate.go:1010-1012`(정지 경로), `migrate.go:1150-1176`(read-ahead) |
| F3 | SHOULD-FIX | outside(증거 문서) | progress.md Gap의 "증명 테스트가 CI의 OS 매트릭스에서 돈다"는 부정확: 카드 PR 단위 테스트는 ubuntu만, 3-OS는 릴리스 PR에서만. cap 4가 모든 플랫폼에 무조건 적용되는데 관측은 macOS뿐이고 Windows는 릴리스 PR 전까지 관측되지 않는다. | `.moai/reports/t2081/progress.md:55`, `internal/config/defaults.go:289-297` |
| F4 | SHOULD-FIX (가설, 신뢰도 중하) | in-diff | `worktree.useRelativePaths=true`(git 2.48+)에서는 첫 상대경로 move가 이름 변경 뒤에 공유 저장소 config(`extensions.relativeWorktrees`)를 쓸 수 있다. 동시 첫 이동이 `config.lock`을 다투면 이름 변경 후 실패(반쯤 옮겨진 트리)가 날 수 있다. 엔진은 이 상태를 역방향/pending으로 다루지만, 스트레스 테스트는 기본 설정만 돈다. git 소스로 확인하지 않은 가설이다. | `migrate_parallel_t2081_test.go:159-223` |
| F5 | MINOR | in-diff | 마무리 루프 안의 조기 반환(`journal.settle` 오류)이 이제 같은 배치의 형제 트리 최대 3개를 "옮겨졌지만 re-key·rebind·`moved` 줄 없음" 상태로 남긴다. `--rollback`으로 수렴하는 상태("moved" 정지 테스트와 같음)지만 이 경로 자체는 테스트되지 않았고, progress.md의 crash model은 kill만 다룬다. | `migrate.go:1089-1091` |
| F6 | MINOR | in-diff | "moved" 정지 후 rollback 테스트가 디스크와 git 목록만 확인한다. move-list의 `restored` 해제 줄, 레지스트리 re-key 복원, 저널 step 결과는 확인하지 않는다. | `migrate_parallel_t2081_test.go:468-483` |
| F7 | MINOR | touched-file | 한 배치 안 순서 비결정성: (a) basename이 같은 두 moai-made 트리(`destinationFor`가 basename만 씀, 계획 시 디스크만 검사)는 직렬일 때 계획 순서가, 배치에서는 경합이 승자를 정한다. (b) 중첩된 moai-made 트리(Witness (a)가 `.moai/worktrees/**` 전체를 덮음)가 같은 배치에 들면 직렬일 때 "identity unavailable"로 건너뛰던 자식이 이제 pending을 만든 뒤 ENOENT → location unknown → pending으로 끝날 수 있다. 결과는 fail-safe 방향이다. 이 저장소에는 현재 중첩 등록이 없다(2.7). | `migrate.go:976-1009`, `migrate.go:540-617`, `classify.go:158-165` |
| F8 | MINOR | in-diff | 스트레스 테스트의 `readerFailures`는 세기만 하고 어디서도 단언하지 않는 죽은 변수. | `migrate_parallel_t2081_test.go:173,187-189` |
| F9 | MINOR | in-diff | 진행 표시(`progress.tree`)는 마무리 중인 트리 하나만 알리는데, 실제로는 최대 3개가 이미 이동 중/이동 완료. 운영자 화면이 진행 상황을 실제보다 적게 보여 준다. | `migrate.go:1054-1058` |
| F10 | MINOR | outside(증거 문서) | 레인의 검증 기록에 `-race` 실행이 없다. F1·F2는 `-race`에서만 드러난다. | `.moai/reports/t2081/progress.md:50-52` |
| (참고) | — | outside | basename 충돌 시 git이 목적지가 디렉터리면 basename을 덧붙여 다른 트리 안으로 옮기는 기존 위험(직렬에서도 동일). 이 카드가 만든 것이 아니며 판정에 영향 없음. 가설 수준(git 동작 기억에 근거, 실행 확인 없음). | `migrate.go:540-617` |

### 필요한 수정 (Required fix)

- **F1**: `migrate_hardening_d_t2053_test.go`의 seam 대체 함수에 `sync.Mutex`를 넣는다(또는 그 테스트만 `moveConcurrency = 1`). 이어서 `forwardGitMove`/`reverseGitMove`를 바꾸는 모든 테스트(`migrate_review_t2053_test.go:405`, `migrate_hardening_b_t2053_test.go:63,170`, `migrate_hardening_c_t2053_test.go:136,176`, 카드 테스트 자신)를 동시 호출 기준으로 다시 점검하고, `migrate.go:1354-1360`의 seam 주석에 "이제 동시에 호출된다"는 계약을 적는다. 수정 뒤 패키지의 seam 사용 테스트를 `-race`로 돌린 출력을 첨부한다.
- **F2**: 조기 반환 시 엔진이 진행 중인 read-ahead 캡처를 기다리게 하거나(`manifestPrefetcher.stop`이 진행 중인 capture 고루틴의 종료를 대기), 테스트가 첫 실행의 백그라운드 작업이 끝난 뒤 재실행하게 한다. 원인 가설은 "정지된 실행의 read-ahead 캡처가 엔진 반환 뒤에도 git 자식을 계속 띄우고, lock fd를 닫기 전에 fork된 자식이 exec 전까지 flock을 함께 쥔다"이며 **검증되지 않았다**. 수정 후 `-race -count=10` 이상의 통과 출력을 첨부한다.
- **F3**: progress.md의 Gap 문장을 사실대로 고친다(카드 PR CI=ubuntu, 3-OS=릴리스 PR). cap을 플랫폼 조건으로 둘지는 선택이나, Windows 관측 전까지는 Gap으로 명시한다.
- **F4**: `git -c worktree.useRelativePaths=true` 변형의 스트레스 라운드를 추가하거나, git 소스로 이름 변경과 config 쓰기의 순서를 확인해 기록한다.
- F5~F10: 선택 사항(optional). 판정에 영향 없음.

## 5. 점검 항목별 판단 (지시된 a~i)

- (a) 이동 호출만 겹치는가: 그렇다. 고루틴 본문은 `a.moveErr = forwardGitMove(...)` 한 줄이고, 나머지 공유 파일 쓰기는 모두 메인 루프에서 계획 순서로 실행된다. 엔진 쪽 경합은 없다. 다만 seam을 바꾸는 테스트 코드의 경합은 F1.
- (b) 폭 1 = 예전 직렬: 코드상 그렇고 테스트로 고정되어 통과.
- (c) crash model: "준비 후 이동 전" 정지와 "이동 후 마무리 전" 정지 두 끝은 테스트됐다. 전자의 재실행 수렴 테스트는 F2로 불안정하고, 후자는 rollback만 다루며 move-list·레지스트리 복원은 확인하지 않는다(F6). 이동 도중 kill은 주입되지 않았다. 각 트리 상태가 기존 단일 트리 테스트(`ForwardMoveInterrupted*`)가 다루는 상태 중 하나라는 주장은 rollback이 step별로 독립 처리하는 코드(`migrate.go:2030-2119`)로 보아 타당하지만, 여러 트리가 서로 다른 반쯤 상태에 있는 조합은 관측되지 않았다.
- (d) 증명 테스트: 감지기 대조군(back-pointer 손상)은 실제로 실패할 수 있음을 보인다. 다만 목록·prune 분기는 대조되지 않는다. "같은 트리 두 번 동시 이동" 대조군은 git의 거부 동작을 보여 줄 뿐 감지기 대조군은 아니다. 12트리 × 8라운드(한 라운드에 12개 동시)는 cap 4보다 넓어 의미 있는 범위지만 표본(96회)은 작다. 타이밍 임계값(2초 대기, 20ms sleep)은 넉넉해서 깜빡임 위험은 낮다고 본다.
- (e) 플랫폼: git의 move 코드 경로는 플랫폼 공통이고 실패는 fail-safe 방향으로 처리되므로 cap 4를 무조건 적용하는 것 자체가 위험하다고 보지는 않는다. 단, 관측 근거에 대한 서술이 부정확하다(F3).
- (f) 수치: 표의 산술(23.3%/37.4% 단축, 255/194/457/289 ms)은 모두 맞다. 짝 두 개뿐이고 짝 사이 편차가 두 배 가까이라 효과 크기 추정은 거칠지만, 문서가 그 한계를 밝히고 있다. 이 감사가 재측정하지는 않았다.
- (g) 누수·교착: 엔진이 시작한 이동 고루틴은 모든 반환 경로 전에 `wg.Wait()`로 합류한다. 교착은 없다. read-ahead 캡처 고루틴은 조기 반환 뒤에도 끝날 때까지 돈다(t2070 설계: "captures already running finish on their own"). 이 카드의 새 조기 반환 경로가 그 성질을 테스트에서 드러냈다고 보며, F2의 가설 원인이다.
- (h) read-ahead(폭 3)와 배치(폭 4): `ensureBatch`가 `get(j..j+3)`을 차례로 부르고, `get(j)`가 슬롯을 풀어야 `j+3` 캡처가 시작되므로 교착은 없다. 배치의 준비가 잠깐 직렬화될 뿐이다. 이동 전 캡처와 이동 후 비교 사이 간격은 넓어졌지만 비교는 그대로 남아 있다.
- (i) 사후 작성 테스트: progress.md는 폭 1 고정 변이에서 배치 테스트 3개가 실패했다고 적었다. 이 감사는 소스 수정이 금지되어 변이를 재현하지 않았다(Gap).

## 6. Gaps (관측하지 않은 것)

- `moai verify audit-plan`을 돌리지 않았다. 운영자 승인 결정(d-20261106T055051Z-9b01, 리더 전달)이 Codex 1회 호출만 허용하고 다른 감사 경로를 금지했기 때문이며, 계획이 고른 경로 대신 그 결정을 따랐다. `audit_multi`를 부르지 않았으므로 orchestrator용 digest(`overall_verdict`/`gate_unmet`/`plan_source`)는 없다. 결정 기록 자체도 `moai decision read`로 직접 읽지 않았다.
- 변이 시험(폭 1 고정)을 재현하지 않았다(소스 수정 금지).
- 전체 패키지·전체 스위트를 돌리지 않았다(지시: 다른 회귀 실행 중). seam을 바꾸는 테스트 중 `TestMigrateApply_Forward*`와 `TestApply_ReportRowsPrecedeTheError`만 `-race`로 확인했고, `migrate_review_t2053_test.go`·`hardening_b/c`의 나머지 seam 테스트 전부를 `-race`로 돌리지는 않았다.
- F2의 원인(fork 시점 fd 상속)은 가설이며 검증하지 않았다. F4는 git 내부 동작에 대한 가설이다.
- Linux·Windows git에서는 아무것도 돌리지 않았다. 이 감사도 macOS에서만 실행했다.
- progress.md의 성능 수치를 재측정하지 않았다.
- golangci-lint를 돌리지 않았다.
- 거부된 도구 호출: 없음.

## 7. Residual-risk (관측했는데도 남는 위험)

- F1·F2를 고친 뒤에도, 동시 이동의 안전성 근거는 macOS·Apple git 2.54.0·기본 config에서의 96회 이동이다. 다른 git 설정(F4)이나 Windows의 파일 핸들 잠금에서 이동이 더 자주 실패할 수 있다. 실패는 fail-safe 경로로 가지만 운영자에게 pending/rollback 작업을 남긴다.
- kill 한 번이 남길 수 있는 미정리 트리 수가 1에서 최대 4로 늘었다(문서에 명시됨). 복구 경로는 같지만 운영자가 처리할 양이 커진다.
- F2의 원인이 가설과 다르면(예: 엔진 반환 뒤 남는 다른 백그라운드 작업), 같은 프로세스 안에서 엔진을 다시 부르는 경로(update consolidation 등)에서도 잠금 충돌이 날 수 있다. 실 운영에서는 kill 후 새 프로세스로 재실행하므로 영향은 작다고 본다.
- Codex 리뷰는 정적 검토였고(그 스스로 밝힘) 한 번만 호출했다. 이 감사와 독립적인 두 번째 의견은 그 1건뿐이다.

## 8. Iteration history

- 1회차(이번): FAIL. 이전 sync 감사 없음. 다음 재감사는 F1·F2의 수정 delta(`3be3e51f36..<수리 커밋>`)로 범위를 좁히면 된다.
