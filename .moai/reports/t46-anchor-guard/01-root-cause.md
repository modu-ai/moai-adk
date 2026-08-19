# t46 — 워크트리 폐기가 앵컈 세션을 죽이는 원인 규명

카드: t46 "워크트리 폐기와 세션 종료가 묶여 있지 않아 레인이 죽는다" (등급 B)
브랜치: `worktree-t46-anchor-guard` @ `cf749fafe` (base `origin/main` = `d6b80a01c`)
작성: 2026-08-17 (run 레인, 칸반 tjv7iy)

## 1. 원인 체인 (Rule 4 — 원인/증상 구분)

| 단계 | 사실 | 증거 |
|---|---|---|
| ① 직접 원인 | `moai worktree done`이 트리 제거 전에 세션 정보를 전혀 확인하지 않음. `runDone`/`runDoneWithAutoMode` 모두 `WorktreeProvider.Remove(targetPath, force)` 직행 | `internal/cli/worktree/done.go` (수정 전 전문 열람) |
| ② 피해 메커니즘 | 트리 소멸 후, Claude Code 런타임의 **네이티브 워크트리 격리 가드**가 그 세션의 모든 Bash를 차단. 순수 조회(`git worktree list`) 포함 | 아래 §2 |
| ③ 감지 데이터 | 레인 세션(`moai cc -w`)은 **트리 안 레지스트리** `<tree>/.moai/state/active-sessions.json`에 `cwd=<tree>`로 등록. `PID`는 유효한 생사 신호 | `internal/hook/session_start.go:1268` (`filepath.Join(input.ProjectDir, session.DefaultRegistryPath)`), `internal/session/registry.go` Entry 스키마 |
| ④ 함정 | `LastHeartbeat`는 **갱신 주체가 없음** — 갱신하는 코드 경로가 `moai session heartbeat` CLI 동사뿐이고 호출하는 곳이 0곳. 실측: primary 레지스트리 엔트리 6개 전부 `last_heartbeat == started_at` | 전수 grep + 실측 레지스트리 열람 (2026-08-17) |

④ 때문에 "heartbeat 신선도"로 생사를 판정하면 장시간 작동한 **살아있는 레인**(사고 당시 형태)을 죽은 것으로 오판한다. 생사 판정은 PID 프로브가 담당하고 heartbeat는 보수적 하한(30분)으로만 쓴다.

## 2. (c) 후보 귀속 — CC 네이티브, MoAI 코드로 수정 불가

사고 메시지 "This session is isolated in the worktree X, but this command working directory resolved to the shared checkout — Refusing to run it there"의 주체:

1. MoAI Go 소스 전수 grep(`isolated in`, `Refusing to run`, `shared checkout`) — 해당 문구 0건
2. **본 레인 세션에서 2026-08-17 직접 목격** (가드가 다른 변형으로 발화, 원문):
   > This session is isolated in the worktree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t46-anchor-guard, but this command is too complex to verify that it stays inside the worktree; break it into plain, separate commands. Refusing to run it — a worktree-isolated session's git operations must target its own worktree.
3. CC changelog 스냅샷(`.moai/research/cc-changelog-snapshot-2.1.233.md`) 225/391/758줄 — 워크트리 격리 세션의 shared checkout 접근 차단은 CC 런타임 시행 사항

따라서 후보 (c) "가드가 앵커 부재 시 primary 강등"은 **상류(CC) 소관**이며 MoAI 코드로 구현할 수 없다. MoAI가 통제 가능한 유일한 개입점은 제거 **이전**이다 — 그래서 (a)(예방)을 구현했다.

## 3. 수정 내용 (최소, 카드 범위 = `done`)

- `internal/session/anchor.go` (신규): `LiveAnchoredSessions(treePath, now)` — 트리 안 레지스트리에서 (host 일치 ∧ cwd 가 트리 하위 ∧ PID 생존 또는 heartbeat 30분 이내) 엔트리 반환. 레지스트리 부재/손상은 fail-open(nil)
- `internal/session/anchor_pid_unix.go` / `anchor_pid_windows.go` (신규): `isProcessAlive` — `internal/cli/update_cleanup_*.go`와 동일 시맨틱(시그널 0 / Windows 보수적 항상-생존)
- `internal/cli/worktree/done.go`:
  - 대화형: 앵커 세션 존재 시 `ANCHORED_SESSIONS_PRESENT:` 센티넬로 **거부(exit 1)**. `--force`면 제거하되 stderr 경고
  - auto: 앵커 세션 존재 시 제거 **건너뛰고** `(false, nil)` 복귀(기존 graceful-degradation 시맨틱 유지)
  - `--help` Long에 가드 문서화

## 4. 검증 (증거 파일: 같은 디렉터리)

- `02-red.txt` — 수정 전 실패 재현: `go test ./internal/cli/worktree/ -run 'AnchoredSessionLive' -v` → 2 FAIL (auto가 Remove 호출 + 성공 보고, 대화형 거부 없음) = 사고 재현
- `03-green.txt` — 수정 후 동일 테스트 PASS + auto 건너뛰기 stderr 출력 확인
- 패키지 전체: `go test ./internal/cli/worktree/ ./internal/session/ -count=1` → ok 6.741s / ok 7.608s
- `go vet` 유닉스+`GOOS=windows` 교차컴파일 통과, `golangci-lint run` 0 issues, gofmt 변경 파일 무지적
- 커밋: `cf749fafe` (6 files, +472/−2), push 안 함(디스패치 지시)

## 4.1 증거 사본 위치

primary 체크아웃(`.moai/reports/t46-anchor-guard/`)으로의 복사는 워크트리 격리 가드가 Write 도구까지 차단해 실패 — 증거는 이 워크트리(`worktree-t46-anchor-guard`)의 같은 디렉터리에만 있다. 리드가 merge 전 primary로 복사하거나, 트리 폐기 후에는 커밋된 재현 테스트(`done_anchor_test.go`, `anchor_test.go`)와 커밋 메시지가 영구 증거가 된다.

## 5. 한계 및 후보 카드

- **EnterWorktree 중간 진입 세션은 비표시**: 세션이 다른 곳에서 시작해 중간에 트리에 진입한 경우(레지스트리 CWD는 시작 시점 값 그대로) 이 감지로 보이지 않는다. CwdChanged에서 레지스트리 CWD를 갱신하는 후속 카드 후보
- **같은 위험의 다른 제거 표면**: `moai worktree remove <path>`, `moai worktree clean`, PR-merge 자동 정리(`session_worktree_prmerge.go:165`가 `git worktree remove` 직접 호출)에는 미적용 — 별도 카드 후보
- **Windows**: PID 프로브가 보수적(항상 생존)이라 죽은 엔트리가 최대 30분(heartbeat 하한) 또는 `--force`까지 폐기를 막을 수 있음. 의도된 보호 우선 방향
