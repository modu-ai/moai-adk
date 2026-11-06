# t2137 progress (Class A, main red: home-guard)

## Claim
Two NEW home accesses cleared; TestHomeGuard_RepositoryHasNoNewHomeAccess passes.

## Evidence
Before (lane-reproduced, as dispatched; not re-run by me): exit 1, 2 new:
- internal/agenticfactory/core/sanitize_test.go:95 `"~/work/notes.md"`
- internal/cli/factory_fixture_home_test.go:36 message "...real ~/.moai"

After (this run, base 04d8c27a7e, fix commit 629bf3ea8b):
```
=== RUN   TestHomeGuard_RepositoryHasNoNewHomeAccess
    repo_guard_test.go:96: scanned 3927 test files: 160 violations, seed 184 entries, 0 new, 24 stale seed entries (stale entries do not fail; regenerate with TESTGUARD_WRITE_SEED=1)
--- PASS: TestHomeGuard_RepositoryHasNoNewHomeAccess (1.74s)
ok  	github.com/modu-ai/moai-adk/internal/testguard	1.842s
```
- `go test -count=1 -run TestSanitizeLeavesOrdinaryTextAlone ./internal/agenticfactory/core/` -> ok 0.262s
- `go test -count=1 -run TestFcFixtureKeepsMoaiHomeInsideTheTestSandbox ./internal/cli/` -> ok 1.479s
- `gofmt -l` on both files: no output.

Changes: marker line `// home-guard:allow deliberate input: ...` directly above the tilde entry; message reworded to "the real user home's .moai". Seed untouched.

## Gaps
- Pre-fix guard output taken from the dispatch, not re-run here.
- Only the three named tests were run, not the full packages.
- The 24 stale seed entries are untouched by design.

## Residual-risk
Merge with main may add other new violations.

## 대기 기록 (lane-28, 2026-11-07)

- sync-audit iter-1 FAIL @76b6dad266 (커밋 6ef059250b): 코드 0.93 전 차원 PASS, 차단 결함 없음. 사유는 audit-plan의 codex required(workflow.yaml:38)와 배차문의 Codex 미승인 충돌 — 영수증 없음.
- 기다리는 것: 리더/운영자 판정 (a) Codex 리뷰 승인 → 재감사 / (b) Claude 단독 결과 수용. 상대: 리더(uds:/tmp/cc-socks/82930.sock), 보고 전송함.
- 재확인 시점: 리더 메시지가 새 턴으로 들어올 때. 카드는 sync 단계 유지, 트리 보존.
