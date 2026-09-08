# t297 — launch-ledger 쓰기 정규화 (REQ-009): 재현 증거

카드 t297 (Class B, Tier S) · lane-15 · branch `WT-launch-ledger-write` · base `304bc8158`

## 1. 전제 측정 (구현 전, 실기계)

측정 대상: `~/.moai/claude-profiles/launch.yaml` — 이 머신의 실제 원장.

```
total project rows: 8
alive dirs: 8
dead dirs: 0
--- worktree-shaped keys:
count: 3
  /Users/goos/MoAI/moai-adk-go/.claude/worktrees/release-v313: moai-adk
  /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t267: moai-adk
  /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t289: moai-adk
```

8행 중 3행이 워크트리 모양 키 — "워크트리마다 행이 쌓인다"는 카드 전제가 실측으로 확인됨.
쓰기 주체 추적: `internal/cli/launcher.go:39` (`recordLastProfileFn` → `profile.RecordLastUsedProfileForProject`)
+ `internal/web/app.go:127`. 런처의 `findProjectRoot`(`internal/cli/glm.go:1064`)은 CWD에서 위로 `.moai`를
찾는데 워크트리는 자체 `.moai/`를 가지므로 워크트리 경로 자체가 기록 키가 된다 — 증가 메커니즘 확정.

## 2. 수정 전 RED 재현 (`red-prefix-run.log` 전문)

명령: `go test ./internal/profile -run 'TestReproWorktreeLaunchesGrowLedgerMonotonically|TestRecordFromSubtree|TestRecordFromRegisteredNestedProject' -count=1 -v`

```
=== RUN   TestReproWorktreeLaunchesGrowLedgerMonotonically
    growth_repro_test.go:37: REPRO: 5 worktree launches grew the ledger to 6 rows (root registered all along)
--- PASS: TestReproWorktreeLaunchesGrowLedgerMonotonically (0.00s)
=== RUN   TestRecordFromSubtreeFoldsIntoRegisteredRoot
    write_normalization_test.go:83: projects rows = 4 (map[.../proj:alpha .../wt1:beta .../wt2:beta .../wt3:beta]), want 1 — worktree launches must fold into the registered root, not add rows
--- FAIL: TestRecordFromSubtreeFoldsIntoRegisteredRoot (0.00s)
=== RUN   TestRecordFromSubtreeThenResolveFromFreshSiblingWorktree
    write_normalization_test.go:109: fresh sibling worktree resolved to "alpha", want beta — the folded write is invisible to the ancestor walk
--- FAIL: TestRecordFromSubtreeThenResolveFromFreshSiblingWorktree (0.00s)
=== RUN   TestRecordFromSubtreeWithNoRegisteredAncestorKeepsOwnRow
--- PASS: TestRecordFromSubtreeWithNoRegisteredAncestorKeepsOwnRow (0.00s)
=== RUN   TestRecordFromRegisteredNestedProjectUpdatesOwnRowNotParent
--- PASS: TestRecordFromRegisteredNestedProjectUpdatesOwnRowNotParent (0.00s)
=== RUN   TestRecordFromSubtreeWithLegacyOwnRowUpdatesOwnRow
--- PASS: TestRecordFromSubtreeWithLegacyOwnRowUpdatesOwnRow (0.00s)
FAIL
```

- 성장 수치: 등록된 루트 위에서 워크트리 5회 기록 → 6행 (실행당 +1행, 단조 증가).
- fold/정합성 2건 RED — 결함 재현. 보존 대상 동작 3건(cold-start·중첩 프로젝트·레거시 중복) 수정 전부터 GREEN.

## 3. 재현 관찰 과정에서 잡힌 픽스처 결함 (기록용)

초판 성장 재현 테스트는 루트 디렉터리 생성 전에 `subtreeLedger`로 등록해
`normalizeProjectKey`가 EvalSymlinks 실패 → 미해결 `/var/...` 철자로 등록(L-002 철자 분열)했다.
그 결과 수정 후에도 fold가 우회되어 6행으로 통과하는 위장 관측이 나왔다. 픽스처를
"실존 디렉터리만 등록"(실제 원장의 불변식)으로 고친 뒤 같은 테스트는 수정 후
`rows = 1, want 2`로 RED — 성장이 사라진 반전 증거. 해당 스크래치 테스트는
fold 테스트가 계승하고 삭제함(`growth_repro_test.go`). 실제 원장은 실존 경로만 기록하므로
이 픽스처 오류는 제품 코드 결함이 아니다.

## 4. 수정 후 (GREEN)

- `green-profile-final.log` — normalization 5 + prune 6 + 회귀 3 = 14건 전부 PASS.
- `green-worktree-final.log` — 배선 13건 전부 PASS.
- `teeth-mutant{1,2,3,4}-*.log` — 변이 4종 전부 RED 관측 후 복원.
