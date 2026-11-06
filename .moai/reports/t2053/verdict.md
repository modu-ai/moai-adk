---
card: t2053
audited_sha: 7d4c633f447a4442642e82ea244b6ec85e1fbc4b
base: 67b4734d6b5aa326cb77ccad5ba8d68452c72af4
verdict: PASS
judge: lane orchestrator (independent re-run; implementer self-report not relied on)
date: 2026-11-06
---

# t2053 판정

## 주장
1. 무시 파일은 경로·크기·종류만 기록하고 읽지 않으며, 추적 안 된 일반 파일은 내용 해시를 유지한다.
2. lsof·git 하위 프로세스와 init의 `git remote`에 시간 제한이 걸리고, 시간 초과는 안전 쪽으로 실패한다.
3. 마이그레이션·바이너리 확인·정리·백업 정리·사용자 잠금 대기에 진행 표시가 붙는다.
4. 마이그레이션 흐름의 부분 실패 상태가 거짓 성공 없이 수렴하거나 실패로 보고된다(리뷰 6회 + 감사 2건 반영).

## 증거 (이 트리, 이 실행에서 직접 실행)

| 명령 | 결과 |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./internal/cli/... ./internal/config/...` | exit 0 |
| `GOOS=windows go vet ./internal/cli/worktree/...` | exit 0 |
| `gofmt -l internal/` | 출력 없음 |
| `go test -count=1 -timeout 700s ./internal/cli/worktree/...` | `ok  github.com/modu-ai/moai-adk/internal/cli/worktree 356.940s` |
| `go test -count=1 -v -run 'NonTTYProgress\|AcquireUserLockNotified\|DetectGitConfig\|RunBinaryUpdateStep\|RunUpdateProgressStep\|PruneBackupsWithProgress\|UpdateConsolidation\|MovedBy\|Launcher' ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli 61.386s`, `=== RUN` 334개 |
| 직전 회차: `go test ./internal/worktreeroot/... ./internal/homestate/...`, `go test -run 'Worktree\|Rekey\|Registry' ./internal/hook/` | 모두 ok (이후 회차에서 이 패키지 코드는 바뀌지 않음) |
| `codex_review scope=card` (7회차) | `verdict: pass`, findings 0 |

## 기준 귀속
- 트리 HEAD `7d4c633f44`, 기준 `67b4734d6b`, 커밋 19개, 29파일 +5735/−242.
- 리뷰 결과: `card-review.md`.

## 미검증 (Gaps)
- `internal/cli` 전체 수트와 저장소 전체 수트는 레인 규칙에 따라 돌리지 않았다. 전체 판정은 CI 몫이다.
- C11(Windows 프로세스 간 잠금)은 macOS에서 컴파일만 확인했다.
- C2·C12의 RED 관찰 기록은 구현 에이전트의 컨텍스트 압축으로 보존되지 않았다.
- golangci-lint는 구현 에이전트 보고(`0 issues.`)이며 이 판정에서 다시 돌리지 않았다.
- 시간 제한 값(git 2분, move 5분, lsof 30초, init 5초, 잠금 안내 1초)은 600개 worktree 환경에서 실측하지 않았다.

## 잔여 위험
- 위치가 증명되지 않은 트리는 pending으로 남고 사람이 `--rollback`을 돌려야 수렴한다.
- `done.go`의 해석 실패를 "L1 아님"으로 보는 동작은 SPEC-WORKTREE-DONE-TIER-001 REQ-005가 명시한 것이라 바꾸지 않았다(후속 판단 필요).
- move-list의 잘린 마지막 줄은 건너뛰며, `worktreeroot.OwnerMarkerOf`는 읽기 오류를 삼킨다(감사 범위 밖).
- 구현 에이전트가 테스트 중 `pkill -x worktree.test`를 한 번 썼다. 같은 시각 다른 세션의 같은 이름 테스트 프로세스가 있었다면 영향을 받았을 수 있다.
- git이 띄운 자식 프로세스(textconv·fsmonitor)는 시간 초과 시 고아로 남을 수 있다(가설, 문서화만 함).
