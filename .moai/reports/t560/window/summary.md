# t560 통합 창 기록 — 병합하지 않음

창 보유: lane-8, 리드 지명 후 acquire(status free 확인 뒤). 워크트리 `.claude/worktrees/t560`, 브랜치 `WT-hook-env-scrub`.
결과: **재측정 비초록 → develop 에 병합하지 않음.** 창은 반납됨(아래 반납 경위).

## 흡수

- 흡수 대상: 로컬 `develop` `84e5666d9714571544bf05e88db9bdc034551ffb`(리드가 알린 tip 과 `git rev-parse develop` 일치).
- 사전 예측: `git merge-tree --write-tree --name-only 84e5666d9 HEAD` → `CONFLICT (content): Merge conflict in CHANGELOG.md` 1건(`t560-merge-tree-84e.txt`).
- 실행: `git merge --no-ff develop` → `t560-window-merge.log`, exit 1, `CONFLICT (content): Merge conflict in CHANGELOG.md` 만. index.lock·`Unable to write index` 없음.
- 병합되지 않은 경로: `CHANGELOG.md` 하나. 인덱스 stage 1 `ff68df7f`(기준점), stage 2 `127732e9`(= `2d38b094b:CHANGELOG.md`), stage 3 `1c2a1ea1`(= `develop:CHANGELOG.md`).

## CHANGELOG 충돌 해결 (리드 승인 방식: 합집합)

- 충돌 표지: 394 `<<<<<<< HEAD`, 401 `=======`, 421 `>>>>>>> develop`. HEAD 쪽 = t560 항목 6줄, develop 쪽 = develop 항목들. 같은 `### Fixed` 목록 머리에 양쪽이 끼워 넣은 인접 삽입.
- 충돌 자리 392행 `### Fixed` 는 `## [Unreleased]`(8행)와 `## [3.1.3] - 2026-08-24`(641행) 사이 — 출시된 절이 아니라 `[Unreleased]` 안.
- 해결: 세 줄이 여전히 표지인지 다시 읽고 확인한 뒤 그 세 줄만 삭제. 결과는 t560 항목 6줄 뒤에 develop 항목이 그대로 이어짐. 문구 무변경.
- 확인:
  - `git diff --check -- CHANGELOG.md` → exit 0
  - 충돌 표지 `^(<<<<<<<|=======|>>>>>>>)` → 0 (대조: 해결 전 사본 3, `changelog-conflicted-marker-count.txt`)
  - `git diff develop -- CHANGELOG.md`(`t560-changelog-vs-develop.diff`) → 추가 6줄·삭제 0줄, 추가 줄(`t560-added-now.txt`)과 t560 원 항목 추가 줄(`t560-added-orig.txt`)이 `cmp` exit 0
  - `^## \[Unreleased\]` → 1
- 흡수 병합 커밋: `2c07f89ff5c0627cf66eb5cfe2e81f889e476416`, 부모 `2d38b094b` · `84e5666d9`, 트리 `d0698df1c3c8e4fc1860ecd2bc8e81a252160e19`. 커밋 뒤 MERGE_HEAD 없음, 추적 수정 0.

## 흡수 트리 상태

- `go version` → `go1.26.8 darwin/arm64`, `go.mod:3` `go 1.26.8`(흡수 전 트리는 go1.26.4).
- `go list ./internal/gitenv ./internal/hook ./internal/hook/quality ./internal/hook/security` → 4패키지, exit 0.
- `git diff --name-only develop HEAD` → 18파일, 전부 t560 커밋의 파일(보고서 3, CHANGELOG, `internal/gitenv` 2, `internal/hook` 12).

## 재측정

사전 확인: `pgrep -lf 'go (test|build|vet)'` → 이 저장소가 아닌 `/Users/goos/MoAI/mo.ai.kr` 의 race·통합 테스트 3개. 이 저장소 internal/cli 컴파일 아님. 대조 zsh 44.

| 실행 | 파일 | exit | 판독 |
|---|---|---|---|
| `go test -count=1 ./internal/gitenv/ ./internal/hook/ ./internal/hook/quality/ ./internal/hook/security/` | `hook-packages.txt` / `.exit` | 1 | gitenv ok · hook/quality ok · hook/security ok · **internal/hook FAIL** |
| `go test -count=1 -v ./internal/template/ -run '^(TestHookWrapperPairParity\|TestHookWrapperConventions)$'` | `hook-copy-parity.txt` / `.exit` | 0 | RUN 2 / PASS 2 |

실패: `--- FAIL: TestSessionStart_DeferredScanDoesNotBlockReturn (0.56s)` — `session_start_parallel_test.go:97: Handle blocked 559.425833ms waiting for advisory scan; expected deferred (non-blocking) return`. 테스트는 drift 스캔을 2초 막히게 주입하고 `Handle()` 반환이 500ms 안이기를 요구한다(`:56-98`).

## 실패 귀속 — 확인된 것과 아닌 것

확인됨:
- t560 코드는 이 경로에 없다. `internal/hook/session_start*.go`(비테스트)가 t560 이 스크럽을 넣은 함수를 부르는 곳 0건. t560 이 바꾼 hook 파일은 `session_end.go` 등이며 `session_start.go` 는 아니다.
- 핸들러와 테스트는 develop tip 과 같다: `develop:internal/hook/session_start.go` = `HEAD:…` = `224968b6`, `develop:…/session_start_parallel_test.go` = `HEAD:…` = `1a448b79`. `git diff --name-only develop HEAD -- internal/hook/session_start*.go` 무출력.
- 흡수가 `session_start.go` 에 들인 +47줄(develop `0c86e61d0` t592, `1cfc6f544`)은 `Handle()` 첫머리 동기 경로다: `homestate.AcquireAdmissionLock`, `homestate.CheckRuntimeAdmission`, `registerProfileLease`.
- 실행 당시 load average 17.99 / 21.02 / 21.34. 같은 저장소 다른 세션이 `internal/hook`·`internal/template` 테스트를 동시에 돌리고 있었다(pid 32445).

확인되지 않음:
- 559ms 가 부하 때문인지, 새 동기 작업 때문인지. 단일 테스트 재실행으로 가를 수 있지만 아래 홈 쓰기 경로 때문에 돌리지 않았다.

## 홈 쓰기 경로 — 재실행을 멈춘 이유

- `registerProfileLease`(`session_start.go`)는 `CLAUDE_CONFIG_DIR` 가 있으면 `homestate.OpenProfileLeases()`(`internal/homestate/profile_lease.go:51-68`)를 연다. 위치는 `paths.MoaiHome()`(`internal/paths/paths.go:68-77`, `MOAI_HOME` 이 절대경로일 때만 그 값, 아니면 `~/.moai`) 아래 `run/profile-leases.db`, SQLite WAL·`busy_timeout(5000)`·immediate txlock. 토큰이 없으면 임시 임대를 만든다.
- 이 세션: `printenv CLAUDE_CONFIG_DIR` → `/Users/goos/.moai/claude-profiles/moai-adk`, `MOAI_PROFILE_LEASE_TOKEN` 길이 0, `MOAI_HOME` 미설정.
- `NewSessionStartHandler(` 를 쓰는 internal/hook 테스트 파일 18개(`t560-ss-handler-tests.txt`) 중 `MOAI_HOME`·`CLAUDE_CONFIG_DIR` 를 격리하는 파일은 0개. 격리 패턴은 핸들러를 직접 만들지 않는 `session_start_profile_lease_test.go:14-17`(`t.Setenv("MOAI_HOME", t.TempDir()…)`, `t.Setenv("CLAUDE_CONFIG_DIR", …)`)에만 있다(`t560-home-isolating-tests.txt`).
- 실제 파일 `/Users/goos/.moai/run/profile-leases.db`: mtime 04:17:14, 331776 B, `-wal`·`-shm` 없음. 내 `internal/hook` 실행 종료는 04:14:36 이므로 마지막 쓰기는 다른 세션이다. 내 실행이 그 전에 행을 썼는지는 mtime 으로 가를 수 없고, DB 는 열지 않았다.
- 창 전·재측정 뒤 동일: `~/.claude/settings.json` sha256 `86e2d9b6…`, `~/.zshrc` `~/.zprofile` `~/.zshenv` `~/.bashrc` `~/.bash_profile` `~/.profile` `~/.gitconfig` mtime. 이 감시 목록에 lease DB 는 없었다.

## 바이너리 사고와 반납 경위

- `~/go/bin/moai` mtime 04:14:35(1789067675), 70493122 B — 내 테스트 종료 1초 전 교체. 이 레인은 설치하지 않았다.
- 그 뒤 `moai integration release`, `moai integration status`, `moai version` 모두 exit 137, 출력 없음.
- 창 상태를 읽으려고 이전 카드(t587)에서 트리 `01059305e` 로 빌드해 둔 스크래치 바이너리로 `integration status` → `held`, holder lane-8.
- 같은 스크래치 바이너리로 `integration release` → `release-integration window released …`, exit 0. 뒤이은 status → `free`.
- 리드의 「release 가 137 이면 재시도하지 말고 출력만 보고, 리드가 복구 뒤 정리」 메시지는 이 반납 뒤에 도착했다.

## 미검증

- 실패의 원인(부하 대 동기 admission/lease 작업). 격리된 홈에서 단일 테스트 반복 실행으로만 가를 수 있다.
- develop tip 자체에서 같은 테스트가 실패하는지(같은 blob 이라 조건은 같지만 실행은 안 함).
- 이 레인의 테스트 실행이 실제 lease DB 에 행을 남겼는지.
- internal/hook 의 다른 테스트가 같은 조건에서 통과한 것이 격리 덕인지 우연인지.
