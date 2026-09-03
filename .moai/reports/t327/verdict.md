# t327 verdict — treeDirty가 described-worthy 판별식 없이 앵커 모드를 고른다

카드: testdata 만 더러운 트리가 `--commit` 병합기준 앵커를 거부당한다 — t322(SPEC-GRAPH-FRESHNESS-CADENCE-001)
v0.2.1 이 명시 유예한 잔여. 트리: `WT-treedirty-predicate` @ develop `2660bcd09`, darwin.

## 판정 요약

`treeDirty`(internal/mx/provenance.go)가 REQ-GFC-002 판별식(`IsDescribedWorthy` — `.go` 끝,
`_test.go` 아님, testdata 세그먼트 없음) 없이 raw `git status --porcelain`로 앵커 분기를 정해
testdata-only 더러움까지 `--commit` 거부 사유로 섰다. 수리: porcelain 출력 경로에 판별식 적용 +
`-z`(NUL 구분 — rename/quote 견고) + `--untracked-files=all`(디렉터리 축약의 거짓 음성 방지).
기존 `IsDescribedWorthy` 단일 구현 재사용 — 새 판별식을 만들지 않았다.

## Claim

1. 수리 전 트리에서 testdata-only 더러움이 `StampCodemaps(root, commit)` 거부와 flagless 경로의
   dirty-fingerprint 스탬프를 재현한다 (RED 관측됨).
2. 수리 후 동일 시나리오가 commit 앵커를 기록하고 production `.go` 더러움은 여전히 거부된다.
3. mx-scan/edges 레이어의 행동 변화는 안전하다 — 이들의 ContentFingerprint 비교자가 없다
   (t322 §O2: `checkCodemaps`가 유일 비교자, codemaps 전용).

## Evidence

| # | 명령 | 관측 출력 |
|---|------|----------|
| E1 (RED) | `go test ./internal/mx/ -run TestStampCodemaps -count=1` (수리 전) | `TestStampCodemaps_ExplicitCommitAllowsTestdataOnlyDirty` FAIL — "…described sources carry uncommitted changes" 거부 재현; `TestStampCodemaps_DefaultPathTestdataOnlyDirtyRecordsCommit` FAIL — dirty=true + fingerprint 기록 |
| E2 (GREEN) | 동일 명령 + `go test ./internal/mx/ -count=1` (수리 후) | `ok github.com/modu-ai/moai-adk/internal/mx 7.263s` — 신규 3테스트 + 기존 전부 |
| E3 | `go vet ./internal/mx/ && golangci-lint run internal/mx/` | vet ok, `0 issues.` |
| E4 | untracked 가드 | `TestStampCodemaps_ExplicitCommitRejectsUntrackedDescribedSource` — 신규 디렉터리 안 untracked `.go`가 여전히 앵커를 거부함을 단언 (수리 전후 GREEN — `--untracked-files=all` 부재 시 깨질 가드) |
| E5 | 형제 스윕 `grep -rn "status.*--porcelain" internal/ --include="*.go"` | 비테스트 6처 — `core/git/manager.go`(2), `verify/key.go`, `worktree/state_guard.go`, `cli/session_worktree.go` — 전부 **트리 전체** 더러움 질의로 판별식 표면 아님. `gitOut` 남은 호출자는 `GitHead`뿐 — treeDirty 의존 제거로 스테일 주석 갱신 |

## Baseline-attribution

모든 측정은 본 커밋 직전 이 워크트리(브랜치 `WT-treedirty-predicate`, HEAD = `2660bcd09` =
로컬 develop — t364 병합으로 b7462203a보다 2커밋 전진한 현재 head)에서 이 실행으로 수행.
E1의 RED는 수리 전 동일 트리에서 관측. TDD 순서 RED→수리→GREEN 준수.

## Gaps

- rename/copy porcelain 레코드(`R  to\0from\0`)의 소스-path 스킵 로직은 구현됐으나 전용 테스트가
  없다 — fixture git 리포에서 rename 상태를 만들어 단언하는 테스트는 미작성 (판별식 접점은
  destination path 기준이라 기존 테스트가 간접 커버).
- spec.md 갱신(§D.1·§E 유예 서술 반전 정리 + History)은 manager-spec 위임 진행 중 — 본 verdict
  작성 시점엔 아직 미커밋 상태.
- `gitOut`의 fail-open 방향(git 실패 → not dirty)은 현행 동작 유지 — 이 방향 자체의 적정성은
  카드 범위 밖.

## Residual-risk

- `--untracked-files=all`이 untracked 나열 비용을 키울 수 있다 — described roots 범위 한정이라
  실측상 무시 가능(E2 7.3s 전 패키지)이나 거대 리포에서의 성능은 미측정.
- `_test.go` 편집도 이제 앵커 거부를 하지 않는다 — REQ-GFC-002 명제상 의도(described 소스 제외)지만,
  codemaps가 _test.go를 서술하지 않는다는 전제가 깨지면 재검토 대상.
