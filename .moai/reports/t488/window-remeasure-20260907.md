# t488 — 통합 창 병합-트리 재측정 (2026-09-07)

## Claim

t488(`WT-premerge-drift-assert`, SPEC-PREMERGE-SETTINGS-DRIFT-001)는 로컬 `develop`
흡수 후의 병합 트리에서 카드 반경 전 패키지가 통과하며, always-loaded 토큰 예산
가드가 PASS이고, sync 종결 이후 `.go` 착지가 없다.

## Evidence

흡수: `git merge develop --no-edit` → `Merge made by the 'ort' strategy.` 충돌 0
(65 files changed, 6551 insertions(+), 281 deletions(-)).

```
$ go test ./internal/config/... ./internal/kanban/... ./internal/template/... -count=1
ok  github.com/modu-ai/moai-adk/internal/config              4.088s
ok  github.com/modu-ai/moai-adk/internal/config/atomicfile   0.869s
ok  github.com/modu-ai/moai-adk/internal/config/toolpolicy   0.522s
ok  github.com/modu-ai/moai-adk/internal/kanban            142.230s
ok  github.com/modu-ai/moai-adk/internal/template           26.303s
ok  github.com/modu-ai/moai-adk/internal/template/agentemit   2.016s
?   github.com/modu-ai/moai-adk/internal/template/scripts    [no test files]

$ go test ./internal/cli/... -count=1
ok  github.com/modu-ai/moai-adk/internal/cli               365.424s
  (+ agentlint, harness, pr, preference, printer, specid, taskledger, uikit,
     update, update/backup, update/deploy, update/merge, update/plan,
     update/report, wizard, worktree — 전부 ok)

$ go vet ./internal/cli/... ./internal/config/... ./internal/kanban/... ./internal/template/...
(무출력, rc=0)

$ go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v
    token_budget_guard_test.go:69: always-loaded surface = 76634 tokens
        (budget 77600, headroom 966, 17 entries)
--- PASS: TestAlwaysLoadedTokenBudget (0.00s)

$ git log --oneline 955e5190c..dc35fd02b --name-only -- '*.go'
(무출력)
$ git log --oneline 955e5190c..dc35fd02b --name-only     # catch-all 대조군
dc35fd02b … progress.md, CHANGELOG.md
78b3211bb … sync-audit.md
b3d0091f3 … progress.md
```

## Baseline-attribution

- 흡수 tip: 로컬 `develop` = `df74b3c9d` (창을 잡은 직후 `git rev-parse --short develop`
  로 직접 재판독). `origin/develop` = `615d18c1f` 로 뒤처져 있어 흡수 대상이 아니다.
- 카드 tip(흡수 전): `dc35fd02b`, `develop..HEAD` = 17 커밋.
- 위 수치는 전부 이 창에서, 이 병합 트리에서 실행한 명령의 출력이다.

**headroom 966 은 이 트리의 실측이다.** 리드가 `develop`(`df74b3c9d`)에서 잰 값은
1,044 이며 그것은 내 측정이 아니다. 두 값의 차 78 토큰은 이 카드가
`kanban-dispatch.md` 에 더한 2줄에 귀속된다 — 예측치와 일치하지만, 일치했다고
적을 뿐 예측치를 내 값으로 인용하지 않는다.

## Gaps

- 전체 스위트(`go test ./...`)는 돌리지 않았다. 판정은 CI 몫(§4.1 규율).
- darwin/windows 매트릭스, 크로스 플랫폼 빌드 미측정.
- `golangci-lint` 미실행.
- `make build` 미실행 — 이 창에서 바이너리를 재빌드하지 않았으므로 임베드 축은
  미측정이다.
- `moai integration acquire` 는 이 창에서 settings-drift 프리플라이트 행을
  출력하지 않았다. 설치된 `moai` 바이너리가 이 카드의 코드를 아직 싣지 않은
  것으로 보이나, 바이너리 커밋을 확인하지 않았으므로 원인은 미검증이다.

## Residual-risk

- 병합은 자동으로 성공했으나 자동 병합 성공이 의미적 정합을 보장하지 않는다.
  develop 이 데려온 t472/t473/t489/t490 변경과 이 카드의 `internal/cli`·
  `internal/kanban` 변경이 겹치는 지점은 테스트 통과로만 확인됐다.
- 증거 파일(이 문서)은 위 측정 **뒤에** 커밋되므로, develop 병합 트리와 재측정
  트리의 차이는 이 파일 1개로 귀속된다.
