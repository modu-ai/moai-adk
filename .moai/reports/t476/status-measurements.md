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

---

# 2026-09-06 재측정 (lane-15 · 흡수 트리 `647cfeae7` 기준)

> 인계: lane-1 소관이 lane-15로 이관됐다(리드 배차 2026-09-06). 위 §1-§5는 lane-1의 기록으로
> 그대로 보존한다 — 이 절은 새 회차이며, 수치는 전부 이 회차에서 이 트리에서 잰 것이다.
> 기준 트리: local develop `3084f1071`을 흡수한 merge commit `647cfeae7` (미푸시 185커밋 포함, 충돌 0).

## 6.0 인계 시점 작업 트리 사고 기록

2026-09-04 03:43(KST) — lane-1의 마지막 커밋(03:20) **이후**에, 5개 codemaps 파일이
`/moai codemaps --force` CLI 템플릿 출력으로 덮였고 판정서 2건이 디스크에서 삭제돼 있었다
(모두 미커밋). lane-1의 측정 기반판은 HEAD에 커밋돼 있었으므로 `git restore`로 복원했다.
덮어쓴 템플릿 출력은 `moai codemaps --force`로 언제든 재생성 가능한 도구 출력이라 손실은 없다.
누가 돌렸는지(본인 실험인지 타 세션인지)는 판정하지 않았다 — 상태 복원만이 이 카드의 소관이다.

## 6.1 게이트 재측정 — t478 수리 게이트로

```
$ go build -o ./bin/moai ./cmd/moai && ./bin/moai graph check    # rc=1
codemaps  metric=described-source-diff     value=10  threshold=40  verdict=fresh
mx-index  metric=inventory-content-diff    value=18  threshold=1   verdict=stale
edges     metric=source-fingerprint-mismatch value=2  threshold=0   verdict=stale
citations metric=positive-cited-path-absence value=0 threshold=0   verdict=fresh
```

- **codemaps 행이 판별식이다**(리드 지시대로 rc는 아님 — mx-index/edges가 rc=1을 상시 만든다).
  lane-1의 실측 재생성이 185커밋 흡수 후에도 **fresh**(described drift 10, 임계 40 미만)로 유지됐다.
  스탬프만 찍었더라면 수리된 게이트가 value>0 + stale로 잡았을 것이다 — 즉 재생성이 진짜였음이
  수리 게이트로 확인됐다.
- mx-index(18)·edges(2)는 이 카드 소관 밖의 계층이다 — 값의 수리는 별도 카드 소관이고, 여기서는
  존재만 기록한다.

## 6.2 CI 적색 6원인 — 흡수 트리 재검증

origin/develop은 `25a3212a9`에 머물러 있어(185커밋 전부 미푸시) CI 실행은 Sep 3 것이 최신이다.
그래서 6원인을 흡수 트리에서 직접 재검증했다:

| # | 원인 (25a3212a9 CI) | 흡수 트리 `647cfeae7` |
|---|---|---|
| 1 | Graph Freshness — codemaps 144 | **수리됨** — t476(이 카드), fresh 10/40 (§6.1) |
| 2 | SPEC Lint — status 필드 ×2 | **미수리** — 동일 2줄 존재 (아래 근거) |
| 3 | CI/Lint — errcheck 1건 | **미수리** — `internal/template/catalog_tree_hash.go:60` 그대로 |
| 4 | Race — TestAlwaysLoadedTokenBudget +123 | **미수리** — overflow 123 동일 (아래 근거) |
| 5 | Race — TestStatusAheadBehindFromHeader | **수리됨** — t474, 통과 (아래 근거) |
| 6 | Race — TestStatusBranchHeaderShapes | **수리됨** — t474, 통과 (아래 근거) |

```
$ grep -n '^status:' .moai/specs/SPEC-CODEX-E2E-MEASURE-001/plan.md \
                     .moai/specs/SPEC-CODEX-E2E-MEASURE-001/acceptance.md
plan.md:5:status: completed
acceptance.md:5:status: completed          ← #2 미수리 재확인

$ go test ./internal/core/git/ -run 'TestStatusAheadBehindFromHeader|TestStatusBranchHeaderShapes' -count=1
ok  github.com/modu-ai/moai-adk/internal/core/git  4.282s   ← #5·#6 수리 확인

$ go test ./internal/config/ -run TestAlwaysLoadedTokenBudget -count=1 -v
token_budget_guard_test.go:69: always-loaded surface = 77723 tokens (budget 77600, headroom -123, 17 entries)
--- FAIL                                    ← #4 미수리 재확인 (overflow 123, CI 수치와 동일)
```

> 측정 대상 이동 주의: t473·t474의 테스트는 185커밋 안에서 **패키지가 이동했다** —
> status 테스트는 `internal/statusline` → `internal/core/git`, 예산 가드는 → `internal/config`.
> Sep 3 CI는 이동 전 위치에서 잰 값이다. 이동 후 위치에서 재측정했기 때문에 위 표가 유효하다.

## 6.3 큐 · 코드 규모 · SPEC 분포

```
$ ./bin/moai todo 2>/dev/null | awk -F'\t' '$1 ~ /^t[0-9]+$/ {print $2}' | sort | uniq -c
   7 picked
  13 queued                                   ← 활성 20장 (lane-1 회차 19장에서 +1)

$ find internal cmd pkg -name '*.go' ! -name '*_test.go' | wc -l        → 1097
$ find internal cmd pkg -name '*_test.go' | wc -l                       → 1775
$ find internal cmd pkg -name '*.go' ! -name '*_test.go' -print0 | xargs -0 wc -l | tail -1
  → 235201 total
$ go list ./... | wc -l                                                  → 137

$ grep -h '^status:' .moai/specs/*/spec.md | sort | uniq -c
  574 completed · 142 implemented · 31 archived · 13 draft · 10 in-progress
  (+ 인용형 "in-progress" 1 · superseded 9 · planned 1 · rejected 1 · retired 1 · 템플릿 1)
```

## 6.4 통합 · 잔량 — 기준을 명시한다

```
$ ./bin/moai integration status   →  release-integration window: free
$ git rev-parse --short origin/develop  →  25a3212a9 (lane-1 회차 이후 무변동)
```

잔량은 **로컬 develop(`3084f1071`) 기준**으로 셈한다. origin/develop 기준으로 셈하면
미푸시 185커밋이 브랜치마다 중복 계수돼(48개 · 합산 3,296커밋) 잔량 판독이 무의미해진다:

| 세는 대상 (기준: local develop) | 값 |
|---|---|
| 앞선 브랜치 | 24개 · 107커밋 |
| 이 중 **워크트리 보유** (= 살아 있는 카드) | **9개 · 62커밋** |
| 등록 워크트리 | 80 |

lane-1 회차(origin==local 시점, 17개·147커밋) 대비 감소분은 그 사이 착지한 카드들
(t474·t478·t479·t480·t482·t483·t484·t485 등)이다.

## Gaps (이 회차)

- 워크트리 80개 각각의 세션 생사는 재지 않았다 — 「브랜치가 앞서 있고 워크트리가 보유돼 있다」는
  기계적 교집합이지 세션이 살아 있다는 뜻이 아니다.
- SPEC 분포는 `spec.md`의 `status:` 필드 존재만 셌다 — 각 서술의 정확성은 검증하지 않았다.
- #2·#3·#4의 수리 작업은 이 카드 소관이 아니며, 착지 여부 판정까지만 했다.
- t475 흡수 종결의 **최종 재측정은 병합 트리에서** 이뤄져야 한다 — §6.1은 흡수 트리 기준이며,
  창에서 병합 뒤 develop 트리에서 같은 판별식(codemaps 행)을 다시 잰다.
