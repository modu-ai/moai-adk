# t451 — 창 진입 전 흡수·재측정 기록

## 흡수

| 항목 | 값 |
|---|---|
| 병합 전 HEAD | `c2679d712` |
| 흡수 대상 | 로컬 `develop` = `e9c6a8564` (behind 96) |
| merge base | `e79c010b8` |
| 병합 커밋 | `f8180acc3` |
| 충돌 | 0 (clean) |
| 병합 후 divergence (`develop...HEAD`) | `0 3` |

보고 시점에 인용했던 `27 2`는 **원격** `origin/develop` 기준이었다. 병합 대상은 로컬
`develop`이고, 그 기준으로는 behind 96이었다 — 리드 직독이 옳았다.

## 재실행 범위 산출

규율: 재실행 범위 = 파일 델타 패키지 ∪ 그것을 의존하는 패키지.

카드 파일 델타(`e79c010b8..c2679d712`):

```
internal/cli/doctor_codex.go
internal/cli/doctor_codex_test.go
internal/cli/doctor_golden_test.go
internal/codexwiring/skills.go
internal/codexwiring/skills_test.go
```

델타 패키지 = `internal/cli`, `internal/codexwiring`.

의존 패키지 산출:

```
go list -f '{{.ImportPath}}|{{join .TestImports " "}} {{join .XTestImports " "}} {{join .Deps " "}}' ./... \
  | awk -F'|' '$2 ~ /moai-adk\/internal\/(cli|codexwiring)( |$)/ {print $1}' | sort -u
→ github.com/modu-ai/moai-adk/cmd/moai
  github.com/modu-ai/moai-adk/internal/cli
```

확정 범위: `internal/codexwiring` ∪ `internal/cli` ∪ `cmd/moai`.

산출 중 실수 1건 — 첫 시도의 grep 패턴이 모듈 경로를 `moai-adk-go`로 썼다(실제는
`github.com/modu-ai/moai-adk`). 빈 출력이 나왔고, 그대로 두면 "의존 패키지 없음"이라는
거짓 판정이 됐을 것이다. 대조군(`go list ./... | wc -l` = 137, `len .Deps` = 543)으로
필터 침묵임을 확인해 정정했다. 두 번째 실수도 같은 계열 — grep이 ImportPath 필드까지
매칭해 `internal/cli/**` 하위 19개를 전부 의존 패키지로 잘못 보고했다. 구분자(`|`)를
넣어 임포트 필드만 보도록 고쳤다.

## 교차 점검 — 흡수분이 카드 표면을 건드렸는가

```
git log --oneline e79c010b8..develop -- internal/cli/doctor_codex.go \
  internal/cli/doctor_codex_test.go internal/cli/doctor_golden_test.go internal/codexwiring
→ (빈 출력)
```

대조군으로 범위 자체는 96커밋임을 확인했으므로 진짜 부재다. 다만 같은 범위가
`internal/cli`의 **다른** 파일 29개를 바꿨으므로, 파일 무교차가 패키지 무교차는 아니다.
그래서 패키지 전량을 다시 돌렸다.

## 재측정 (병합 트리 `f8180acc3`)

| 대상 | 명령 | 결과 |
|---|---|---|
| 전체 빌드 | `go build ./...` | rc=0 |
| 경량 | `go test ./internal/codexwiring/... ./cmd/moai/... -timeout 300s` | rc=0 · codexwiring ok 0.812s · cmd/moai no test files |
| 중량 | `go test ./internal/cli/ -timeout 1800s` (단독) | rc=0 · ok 646.636s |

원본 출력: `remeasure-light.log`, `remeasure-cli.log`.

`internal/cli` 646.636s는 리드가 인용한 실측 대역(626~788s) 안이다.

## 미검증 (Gaps)

- darwin/windows 매트릭스, 깨끗한 환경 — CI 몫. 로컬 초록은 조기 신호일 뿐이다.
- `internal/cli` 외 흡수분이 건드린 패키지들 — 각 레인이 자기 창에서 이미 측정했고,
  이 카드의 델타가 그 패키지들을 의존하지 않으므로 범위 밖으로 뒀다.
- 이 재측정은 로컬 develop `e9c6a8564` 기준이다. 창 호명 시점에 develop이 더
  움직였다면 무효이며, 그때 다시 잰다.

---

## 2차 흡수·재측정 (base 스테일화 대응)

1차 재측정이 646초 도는 사이 develop이 5커밋 전진했다(t458 착지, `3bdd5a803`).
스테일 기준으로 창에 들어가지 않고 흡수 후 전 범위를 다시 돌렸다.

| 항목 | 값 |
|---|---|
| 흡수 후 HEAD | `801318559` |
| 흡수분 | t458 — `.claude/rules/moai/workflow/main-checkout-branch-guard.md` + 템플릿 미러 + 리포트 |
| 재측정 | `go test ./internal/cli/ ./internal/codexwiring/... ./cmd/moai/... -timeout 1800s` |
| 결과 | rc=0 · `internal/cli` ok **501.041s** · `codexwiring` ok 1.042s · `cmd/moai` no test files |

원본 출력: `remeasure-cli-2.log`.

### 부수 발견 — develop 상속 레드 (소관 t453, 이 카드 아님)

병합 트리 `801318559`에서 `internal/config`가 레드다.

```
go test ./internal/config/ -run 'TestAlwaysLoadedTokenBudget$' -v
token_budget_guard_test.go:69: always-loaded surface = 77104 tokens
                               (budget 76400, headroom -704, 17 entries)
--- FAIL: TestAlwaysLoadedTokenBudget (0.12s)
```

귀속: 이 카드 것이 아니다.

```
git diff --name-only develop HEAD -- .claude/rules → 0 files
```

t451의 델타는 Go 5파일 + `.moai/reports/t451/**`뿐이고 always-loaded 표면을 한 건도
건드리지 않는다. 이 트리의 `.claude/rules`가 develop과 바이트 동일하므로 77,104는
develop 자신의 수치이며, develop에 합류하는 모든 레인이 이 레드를 상속한다.

원인은 리드가 develop 워크트리에서 직접 재어 닫았다 — t458이 그 구간의 유일한
always-loaded 표면 변경이고, 6395 → 7055바이트(Δ660)가 76,939 → 77,104(Δ165)에
대응한다(660/165 = 4.0 바이트/토큰). 내가 "가설"로 표시했던 것이 리드의 측정으로
확정됐다. 수리 소관은 t453(lane-9, 상한 77,200 → 적용 후 여유 96).

리드 판정: 이름 밝혀 제외하고 진입 — `TestAlwaysLoadedTokenBudget`, 소관 t453, 상속.

## 3차 흡수 (재측정 생략 — 근거 첨부)

2차 재측정이 501초 도는 사이 develop이 다시 5커밋 전진했다(t440 착지, `4c3b1653c`).
흡수했고(HEAD `07f0d39b1`), **이번에는 재측정을 생략했다.** 근거는 인상이 아니라 측정이다.

```
git diff --name-only 801318559 07f0d39b1 -- internal cmd pkg   → 0 files
git diff --name-only 801318559 07f0d39b1 | wc -l               → 9   (대조군)
```

전체 델타는 9건인데 `internal/`·`cmd/`·`pkg/` 하위는 0건 — 즉 필터가 조용한 것이
아니라 컴파일 대상이 실제로 불변이다. 흡수분 9건은 전부 `docs-site/content/**`와
`.moai/reports/t440/**`다. `internal/template/templates`(임베드 대상)도 `internal/`
하위이므로 같은 검사에 포함돼 0으로 확인됐고, `.claude/rules`도 델타에 없어 토큰 예산
표면 역시 불변이다.

따라서 `801318559`에서 얻은 501.041s 초록이 `07f0d39b1`에 그대로 이월된다.

## 이 카드가 겪은 구조적 성질

`internal/cli`는 626~788s 대역이고, 이 리포의 develop은 그보다 짧은 간격으로 움직인다.
**오래 걸리는 재측정은 그 자체가 자기 base를 낡게 만든다.** 무한 추격을 피하는 방법은
매번 다시 도는 것이 아니라, 새 델타가 컴파일 대상을 건드렸는지 먼저 재고 그 측정으로
재실행 여부를 판정하는 것이다 — 위 3차가 그 적용례다.
