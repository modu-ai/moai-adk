# t698 auto-done 제목 귀속 세대 경계 — 판정 (lane, 2026-09-13)

## Claim (주장)

`moai todo auto-done`의 subject-attribution 귀속은 세대교차 재발행 카드(전임이 스토어에 없는 재발행)의 판정에서 `committer date < 카드 added_at`인 구세대 커밋을 더 이상 착지 근거로 쓰지 않는다. 시간 경계는 `internal/kanban`의 순수 술어 `AutoDoneSubjectFresh`로 구현되고, CLI 결합부(`planAutoDone`)가 이를 통과한 히트만 `AutoDoneFacts.SubjectHit`로 넣는다.

## Evidence (증거)

1. **RED (수리 전, 이 트리 커밋 이전 워킹 상태)**:

   ```
   $ go test ./internal/cli/ -run TestTodoAutoDone_ReissuedIDOlderCommitSkips -count=1
   --- FAIL: TestTodoAutoDone_ReissuedIDOlderCommitSkips (1.04s)
       stdout "done t9988 landing=landed source=auto-land ref=origin/develop
       form=subject-attribution ... closed=1" — closes t9988 on the old
       generation's commit
   FAIL
   ```

   t684 필드 결함과 동일 양상 재현: 전임 부재 → `AutoDoneDistinctTexts`=1 → 가드 M1 침묵 → 24시간 이전 커밋이 새 카드를 종결.

2. **GREEN (수리 후)**:

   ```
   $ go test ./internal/cli/ -run TestTodoAutoDone_ReissuedIDOlderCommitSkips -count=1
   ok  github.com/modu-ai/moai-adk/internal/cli  2.934s

   $ go test ./internal/kanban/ -count=1 -timeout 10m
   ok  github.com/modu-ai/moai-adk/internal/kanban  175.733s
   ```

3. **라이브 큐 행동 dry-run (이 트리 빌드 바이너리 — §2.2 도구 귀속)**:

   ```
   $ go build -o /tmp/t698-moai ./cmd/moai && /tmp/t698-moai todo auto-done --dry-run
   auto-done scanned=32 closed=0 skipped=32 ref=origin/develop
   note: dry-run — the queue record was not modified
   EXIT=0
   ```

## Baseline-attribution (baseline 귀속)

- 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t698`, 브랜치 `WT-auto-done-time-bound`, 기점 로컬 develop `44e56d017`.
- 측정 명령과 출력은 위 Evidence 그대로, 이 실행에서 관측.
- 변경 파일: `internal/kanban/autodone_scan.go`(`LandedCommit.CommitTime` 추가, 스캔 포맷 `%H%x00%ct%x00%s`, `AutoDoneSubjectFresh`), `internal/cli/todo_autodone.go`(귀속 히트에 신선도 게이트), 테스트 2파일.

## Gaps (미검증)

- `internal/cli` 패키지 전량은 재측정이 2회 모두 `go test` 기본 600s 타임아웃 panic으로 종료(`--- FAIL` 0건, 알려진 1582s급 패키지) — 대신 Todo 선택자 전체로 스코프 재측정: `go test ./internal/cli/ -run 'Todo' -count=1 -timeout 30m` → `ok github.com/modu-ai/moai-adk/internal/cli 374.920s`. 전량 판정은 CI 몫.
- golangci-lint 미실행(CI 몫으로 예상). `go vet`·`gofmt`는 통과.
- 스캔 포맷 3필드화가 외부 소비자(스크립트가 `--format` 출력을 파싱하는 형태)에 미치는 영향은 검증 밖 — `LandedScanArgs`는 내부 전용(grep 확인, 이 구조체 생성자는 이 파일 하나).

## Residual-risk (잔여 위험)

- 동일 초(committer time == added_at) 커밋은 fresh로 판정 — 세대 분리 목적상 무해하나 초단위 위변조 시나리오는 커버하지 않는다(술어 주석에 명시).
- `added_at` 파싱 불가 카드는 fail-closed(not-landed) — 스토어 손상 시 auto-done이 그 카드를 닫지 못하게 되는 방향.
- merge 커밋의 committer date는 머지 시점이라 정상 착지 흐름(카드 생성 → 구현 → 병합)에서는 항상 added_at 이후 — 카드가 만들어지기 전에 이미 병합된 커밋이 "이 카드의 착지"인 특이 흐름은 이제 의도적으로 닫히지 않는다(t684 판정과 일치).

## 관련

- 원인 판정: `.moai/reports/t684/field-defect-20260913.md` (t657-queue-merge 트리 보관)
- 수리 후보 1(시간 경계, 구조적) 채택 — 후보 2(재발행 원장)는 미채택, 별도 소관
