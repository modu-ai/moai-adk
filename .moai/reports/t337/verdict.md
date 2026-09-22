# t337 verdict — stamped-but-dead window (Windows anchor PID probe)

카드: Windows 에서 스탬프가 있는데 죽은 세션이 살아있는 것으로 읽히지 않는다 — TREAT-AS-LIVE 역행 경로.
근거: t298 sync-audit F2 [MINOR][optional] (`.moai/reports/t298/sync-audit.md:178`), lane-17 2026-08-27.
트리: `WT-windows-anchor-probe` @ develop `b7462203a` (로컬 develop 병합 후), darwin.

## 판정 요약

카드 본체 결함(코드 축)은 **t426 이 선점 수리**했다 — 카드 발행(2026-08-27) 이후 `4b86425d3`
(t426 axis 2)이 무조건 true 스텁을 진짜 probe로 교체. 카드가 F2에서 물려받은 유일한 열린 잔여는
**(i) §G 문서 누락**이며, 본 커밋으로 수리(§G 불릿 1건 추가). 전체 DROPPED가 아닌 이유: F2의
성분 분해 ((i) 문서 + (iii) 코드 후속) 중 (iii)은 t426이 수행했으나 (i)은 어디서도 닫히지 않았다.

## Claim

1. `internal/session/anchor_pid_windows.go` 의 `isProcessAlive` 무조건 true 스텁은 현재 트리에 없다 — 진짜 OpenProcess + GetExitCodeProcess probe로 교체돼 있다.
2. `sessionPIDFromEnv` 는 죽은 PID를 기록하지 않고 거부한다 — "rejected rather than recorded" 주석 + 테스트.
3. 죽은 PID 거부 판정 경로는 darwin 테스트로 GREEN 이다.
4. t298 spec.md §G 에 스탬프-있는데-죽음(stamped-but-dead) 경로가 없었고, 본 카드가 불릿 1건으로 추가했다.

## Evidence

| # | 명령 | 관측 출력 |
|---|------|----------|
| E1 | `git show 4b86425d3 -- internal/session/anchor_pid_windows.go` | 이전: `func isProcessAlive(pid int) bool { _ = pid; return true }` → 이후: `probeProcessLiveness` (OpenProcess PROCESS_QUERY_LIMITED_INFORMATION + GetExitCodeProcess, stillActive=259, ERROR_ACCESS_DENIED → alive) |
| E2 | `go test ./internal/session/ -run "TestSessionPIDFromEnv\|TestRegister_RecordsLivePID\|TestResolveSessionPID" -count=1` | `ok github.com/modu-ai/moai-adk/internal/session 0.359s` — `session_pid_test.go:140` 의 `"-3", "6000" /* dead */` 거부 케이스 포함 |
| E3 | `GOOS=windows go vet ./internal/session/` | rc=0, 무출력 — **컴파일 근거일 뿐, 동작 근거 아님** (카드 검증 주의 그대로) |
| E4 | `git diff` (spec.md) | §G 첫 불릿 뒤 `A stamped-but-dead owner pid` 불릿 13줄 삽입, `grep -c stamped-but-dead` = 1, `git diff --stat` = 1 file 13 insertions |
| E5 | `git log --oneline --all -- internal/session/anchor_pid_windows.go` | 수리 커밋 = `4b86425d3` (t426) — 그 이전 `8ff3e0823` (t209)까진 스텁 |
| E6 | F2 원문 `.moai/reports/t298/sync-audit.md:178-196` | 분류: (i) §G 문서 누락 + (iii) 후속 카드 — (ii) 아님(이 SPEC 이전부터의 스텁) |

## Baseline-attribution

모든 측정은 본 커밋 직전 이 워크트리(`.claude/worktrees/t337`, 브랜치 `WT-windows-anchor-probe`,
HEAD = `b7462203a` = 로컬 develop)에서, 이 실행에서 수행. E2/E3는 darwin 개발기 — E3는
GOOS 크로스 컴파일이므로 실행 근거가 아니다. E1/E5는 git 히스토리 판독.

## Gaps

- **windows probe 의 동작 직접 테스트는 존재하지 않는다** — `probeProcessLiveness` 를 테스트하는
  windows 유닛 테스트 0건 (t426 이 만들지 않음). 첫 동작 관측은 develop push 이후 windows CI leg
  가 된다 (t426 커밋은 origin/develop 보다 32커밋 앞선 로컬 develop 에만 존재 — windows CI 는
  아직 이 probe 를 관측한 적 없다).
- 카드의 "RED 먼저 관측"은 **수행 불가** — t426 이 카드 수행보다 먼저 수리해, 본 카드가 시작한
  시점에는 결함이 이미 닫혀 RED 를 만들 트리가 없었다. 카드 TDD 지시의 선결 전제가 소멸한 형태.
- §H History 행은 추가하지 않았다 — 감사 후속 정정이며 frontmatter/version 을 건드리지 않는
  선택. 필요하면 sync 소관에서 재판단.

## Residual-risk

- spec.md §G 불릿의 서술이 t426 커밋 메시지·diff 에 근거하는 것은 사실이나, probe 의 windows
  런타임 동작 자체는 여전히 CI 관측 대기 — develop push 후 windows leg 가 붉으면 본 불릿의
  "reject path exercised" 서술은 재검토 대상이 된다.
- 카드 원문의 "TREAT-AS-LIVE 에 역행하는 유일한 경로"라는 정량 서술("유일")은 F2 원문의 서술을
  인용한 것으로, 본 카드는 그 범위(t298 SPEC 표면) 밖의 유사 경로를 전수하지 않았다.
