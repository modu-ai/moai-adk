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
