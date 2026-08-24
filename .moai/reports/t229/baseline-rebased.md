# t229 — run-phase 착수 베이스라인 (rebase 후 재측정)

리드 지시(Kickoff §참고)에 따라 **바이너리 재설치 + origin/main rebase 후** 트리 기준으로 부분 RED 기대값을 재확인했다.

| 항목 | 값 |
|---|---|
| 측정 트리 | `6346e643b` (rebase 후 `WT-audit-verdict-converge` HEAD, base = `origin/main` `f7eec06c7`) |
| 이전 측정 트리 | `294b4b6ab` |
| 측정 방법 | `internal/cli` 에 임시 테스트를 넣어 `synthesizeReviewOutput` 직접 호출 후 삭제 |
| 측정 일자 | 2026-08-25 |

## rebase 가 필요했던 이유

`git rev-list --count --left-right origin/main...HEAD` 가 `10 6` 을 냈고, upstream 10건 중 **`baa100ce5`(t225, #1642)가 `mcp_codex.go`(+106/-16)와 `mcp_convergence.go` 를 건드렸다.** 스테일 base 위에 구현하면 같은 파일에서 충돌하므로 먼저 rebase 했다(6커밋 전부 문서, 충돌 0, 결과 `0 6`).

**합성 seam 자체는 t225 가 건드리지 않았다** — `git show baa100ce5 -- internal/cli/mcp_codex.go | grep -E '^[-+].*(synthesizeReviewOutput|codexFindingBullet|codexStatedVerdict|verdict :=)'` 무출력. 따라서 동작 베이스라인은 그대로이고 **줄 번호만 밀린다**(SPEC 인용은 M1 착수 시 재측정 대상).

부수 확인: t225 의 CR 대응 보고서가 **Major #8 을 t246 계열로 기록**했다 — 감사/핀 해소가 심사 대상과 다른 트리를 읽는 결함. 내가 F9 에서 좁혀 올린 것과 같은 계열이며, `projectRoot` threading 이 인용 가능한 선례로 남았다.

## 바이너리 상태

```
$ ~/go/bin/moai version
 v3.1.3-rc.1   30afb9a1d   built 2026-08-24T15:28:13Z
$ git merge-base --is-ancestor f505955a9 30afb9a1d ; echo $?   # t178 → 0 (포함)
```

리드가 알린 재설치가 확인됐다. 종전 `v3.1.2 @ a1b1ca696`(259커밋 스테일, t178·t186 미포함)이 라이브 프로브의 `pass` 를 만들어 낸 원인이었고, 지금 바이너리에는 `codexStatedVerdict` 파서가 들어 있다. **다만 이 SPEC 의 검증은 `go test` 로만 한다**(plan.md §A) — 바이너리 경로를 근거로 쓰지 않는다.

## 서식 corpus (AC-CVS-001 — 하나도 `pass` 가 아니어야 함)

| # | 합성 verdict | stated 매치 | bullet 매치 | 상태 |
|---|---|---|---|---|
| C1 | `pass` | false | false | **RED** |
| C2 | `pass` | false | false | **RED** |
| C3 | `pass` | false | false | **RED** |
| C4 | `pass` | false | false | **RED** |
| C5 | `pass` | false | false | **RED** |
| C6 | `pass` | false | false | **RED** |
| C7 | `pass` | false | false | **RED** |
| C8 | `pass` | false | false | **RED** |

**8/8 RED.** 감사 iter1 이 `294b4b6ab` 에서 측정한 것과 동일하다 — acceptance.md 의 "C1~C8 전부" RED 기대가 rebase 후에도 유효하다.

## 조합 corpus (AC-CVS-006 — 채택값 = 집합의 최댓값)

| # | 합성 verdict | 기대 | 상태 |
|---|---|---|---|
| K1 | `fail` | `fail` | GREEN |
| K2 | `fail` | `fail` | GREEN |
| **K3** | **`pass`** | **`fail`** | **RED** |
| K4 | `inconclusive` | `inconclusive` | GREEN |
| K5 | `fail` | `fail` | GREEN |
| K6 | `fail` | `fail` | GREEN |
| **K7** | **`pass`** | **`inconclusive`** | **RED** |
| K8 | `pass` | `pass` | GREEN |

**부분 RED 2행(K3·K7), 나머지 6행 GREEN** — 감사 iter2 실측과 **완전 일치**한다. rebase 도 바이너리 재설치도 이 기대를 바꾸지 않았다.

[HARD] **이 두 행을 초록으로 뒤집는 것은 M2**(점수 표기 신호 도입)다. M1(fall-through 교정)만 착지한 중간 상태에서 K3·K7 이 여전히 붉은 것은 **회귀가 아니라 설계대로**다 — 두 행 모두 `stated` 가 매치되므로 fall-through 경로를 타지 않는다. 그 시점에 기대값을 관측 동작에 맞춰 낮추면 두 행의 검출력이 사라지며, 그것이 이 카드가 다루는 실패 형태다(acceptance.md §C AC-CVS-006 `[HARD]`).

## 검증 명령

```
go test ./internal/cli/ -run TestT229BaselineRebased -v -count=1 -timeout 1200s   → ok 0.948s
go vet ./internal/cli/                                                            → rc=0
```

임시 테스트는 측정 후 삭제했다.
