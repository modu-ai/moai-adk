# t476 진행 상황 실측 기록

> 측정 트리: worktree `.claude/worktrees/t476`, 브랜치 `WT-codemaps-progress`, HEAD `25a3212a9`
> 측정 시각: 2026-09-03T18:02Z (로컬 2026-09-04)
> 모든 수치는 이 트리에서 이 시각에 잰 것입니다. 리드가 전달한 `17:44Z` 회차 수치는 근거로 쓰지 않았습니다.

## 1. 백로그 큐

```
$ moai todo
queued 12 · picked 7 · dropped 26 (hidden)
```

카드 총 19장이 활성(`queued` + `picked`), `dropped` 26장은 숨김 처리됩니다.

## 2. develop

```
$ git fetch origin develop
$ git rev-parse --short origin/develop
25a3212a9
$ git rev-list --count origin/develop..HEAD
0
```

이 워크트리는 `origin/develop` tip에서 갈라져 나왔고 아직 자체 커밋이 없습니다.

## 3. CI — 적색 워크플로 3건, 그러나 **원인은 6건**

`origin/develop` `25a3212a9`에 대해 워크플로 7건이 돌았고 3건이 적색입니다.

```
$ gh run list --branch develop --limit 12 --json workflowName,conclusion,headSha
lsel-leak-guard            success   25a3212a9
Template Neutrality Check  success   25a3212a9
docs i18n parity check     success   25a3212a9
CodeQL                     success   25a3212a9
Graph Freshness            failure   25a3212a9   (run 33746474487)
SPEC Lint                  failure   25a3212a9   (run 33746474481)
CI                         failure   25a3212a9   (run 33746474518)
```

**워크플로 3건이 적색이지만, 그 안의 실패 원인은 6건**입니다. `CI` 워크플로 하나가 job 두 개에서
네 가지 사유로 깨지기 때문입니다.

| # | 워크플로 / job | 사유 | 카드 |
|---|---|---|---|
| 1 | Graph Freshness | codemaps `described-source-diff` 144 (임계 40) | **t475** (이 카드가 흡수) |
| 2 | SPEC Lint | `ArtifactStatusFieldForbidden` × 2 | **없음** |
| 3 | CI / Lint | `golangci-lint` errcheck 1건 | **없음** |
| 4 | CI / Race Test | `TestAlwaysLoadedTokenBudget` 초과 123 | **t473** |
| 5 | CI / Race Test | `TestStatusAheadBehindFromHeader` | **t474** |
| 6 | CI / Race Test | `TestStatusBranchHeaderShapes` | **t474** |

### 소관자 없는 적색 2건 — 이 트리에서 재현

CI 로그를 그대로 옮기지 않고 직접 재현했습니다.

**#2 — SPEC Lint (error 2건, warning 4311건)**

warning은 종료코드에 영향이 없습니다. exit 1을 만든 것은 error 2건입니다.

```
$ grep -n '^status:' .moai/specs/SPEC-CODEX-E2E-MEASURE-001/plan.md \
                     .moai/specs/SPEC-CODEX-E2E-MEASURE-001/acceptance.md
.moai/specs/SPEC-CODEX-E2E-MEASURE-001/plan.md:5:status: completed
.moai/specs/SPEC-CODEX-E2E-MEASURE-001/acceptance.md:5:status: completed
```

`spec.md`가 아닌 SPEC 아티팩트는 상태 축에서 무상태여야 하는데(`ArtifactStatusFieldForbidden`)
두 파일이 `status:`를 들고 있습니다. **각각 한 줄 삭제**입니다.

**#3 — CI / Lint (errcheck)**

```
$ sed -n '58,62p' internal/template/catalog_tree_hash.go
	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
```

`internal/template/catalog_tree_hash.go:60:14` — `fmt.Fprintf` 반환값 미검사.
쓰기 대상이 `sha256.New()`라 실제로 실패할 수 없는 자리이므로 `_, _ =` 한 줄이면 닫힙니다.

## 4. 통합 창

```
$ moai integration status
release-integration window: held
  holder:   00784125-9826-40a6-9605-d9edd3bbe501 (pid 28524)
  branch:   WT-update-hook-delivery
  worktree: .claude/worktrees/t466
  since:    2026-09-03T10:52:57Z
```

측정 시각 기준 **429분** 점유(약 7시간 9분). 그 사이 `origin/develop`은 `25a3212a9`에서
움직이지 않았습니다.

## 5. 미병합 잔량 — 단위를 밝혀 둡니다

세 가지 수가 서로 다른 모집단을 셉니다. 섞어 읽으면 안 됩니다.

```
$ git for-each-ref --format='%(refname:short) %(ahead-behind:origin/develop)' refs/heads/
```

| 세는 대상 | 값 |
|---|---|
| 로컬 브랜치 전체 | 117 |
| `origin/develop`보다 앞선 브랜치 | 32개 · 합산 192커밋 |
| 그중 **워크트리를 보유한 것** (= 살아 있는 카드 작업) | **17장 · 합산 147커밋** |
| 등록된 워크트리 전체 | 69 |

앞선 브랜치 32개와 살아 있는 카드 17장의 차이는 워크트리 없이 남은 잔재 브랜치입니다 —
squash 병합으로 착지했으나 로컬 ref가 남았거나 폐기된 것들입니다.

> 리드가 `17:44Z` 회차에 보고한 「미병합 12장 95커밋」과 다릅니다. **회차가 다릅니다** —
> 그 사이 작업이 더 쌓였고, 리드의 12는 카드 단위, 여기 17은 워크트리 보유 브랜치 단위입니다.
> 어느 쪽이 틀린 것이 아니라 세는 대상이 다릅니다.

## Gaps — 관측하지 않은 것

- `dropped` 26장의 내용은 열어보지 않았습니다.
- 적색 6건 중 소관자 있는 4건(t473 · t474 × 2 · t475)의 **수리 가능성**은 판정하지 않았습니다 —
  재현과 귀속까지만 했습니다.
- 워크트리 69개 각각의 생사·최종 갱신 시각은 재지 않았습니다. 위 17장은 「브랜치가 앞서 있고
  워크트리가 등록돼 있다」는 기계적 교집합이지, 그 세션이 살아 있다는 뜻이 아닙니다.
- `origin/main` 대비 수치는 재지 않았습니다. 이 리포의 통합 지점은 `develop`입니다.
