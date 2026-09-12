# t662 verdict — TestSessionStart_DeferredScanDoesNotBlockReturn: 결함인가 예산이 얇은가

- 카드: t662 (Tier S, Class B)
- 워크트리: `.claude/worktrees/t662`, 브랜치 `WT-deferred-scan-budget`
- 트리: `fb8bfff95` (배차 지정 base = 로컬 develop, 일치 확인)
- 대상: `internal/hook/session_start.go` / `internal/hook/session_start_parallel_test.go:41`

---

## Claim

1. **(A) "t603 흡수가 원인을 제거했다"는 배제된다.** 제거할 코드 변화가 존재하지 않는다.
2. **(B) 예산이 얇다 — lane-4 가설 확정.** 500ms 예산의 실질 여유는 250ms이고, 동기 작업 중앙값이 이미 그 여유의 82~87%를 소모한다.
3. **(추가, 카드 범위 밖에서 발견) 이 단정은 자기가 이름 붙인 실패를 판별할 수 없다.** 160회 관측 전부에서 지연(deferral)은 정상 동작했다. 이 테스트는 지연 실패로 붉어진 적이 한 번도 없다.

---

## Evidence

### (A) 코드 동일성 — 기계적 확정

```
$ git rev-parse eabce7444:internal/hook
20b5a3dede8cf7d22dd5ddc94b8e63b5f18ecde2
$ git rev-parse 63fa51cf3:internal/hook
20b5a3dede8cf7d22dd5ddc94b8e63b5f18ecde2
```

`internal/hook` 트리 해시가 base와 병합본에서 **동일**하다. 전체 diff도 확인했다:

```
$ git diff --stat eabce7444 63fa51cf3
 .claude/agents/harness/hns-release-update-specialist.md  |  37 +++-
 .claude/hooks/moai/sync-phase-quality-gate.sh            |  43 ++++-
 .claude/hooks/tests/test-language-routing-contract.sh    |  52 ++++++
 .moai/reports/t594/verdict.md                            | 182 ++++++++++
 .moai/reports/t603/repro-cc-output.txt                   |   4 +
 .moai/reports/t603/verdict.md                            | 121 ++++++++++
 .moai/reports/t604/hook-root-test.log                    | 165 ++++++++++
 .moai/reports/t604/verdict.md                            | 157 +++++++++
 CHANGELOG.md                                             |   1 +
 internal/template/hook_cpp_gate_behavior_test.go         | 186 +++++++++++
 internal/template/templates/.claude/hooks/moai/sync-phase-quality-gate.sh | 43 ++++-
 11 files changed, 977 insertions(+), 14 deletions(-)
```

유일한 Go 파일 추가분 `internal/template/hook_cpp_gate_behavior_test.go`는 **다른 패키지**(`internal/template`)의 테스트 파일이라 `internal/hook` 테스트 바이너리에 들어가지 않는다. 나머지는 셸 스크립트·마크다운·CHANGELOG다.

따라서 두 커밋의 `internal/hook` 테스트 바이너리는 동일하다. **동일 부하 N회 재측정은 불필요하다 — 잴 코드 차이가 없다.** 두 커밋 사이 적녹 차이가 관측된다면 그것은 코드가 아니라 환경에 귀속된다.

### (B) 예산 분해 측정

프로브: `internal/hook/zz_t662_probe_test.go` (임시, 측정 후 삭제 — 이 verdict가 산출물).
`Handle`의 경과시간을 두 성분으로 가른다.

- **FAST**: drift fn 즉시 반환 → join 기여 ≈ 0 → **순수 동기 작업**
- **BLOCK**: 실제 테스트와 동일한 2초 주입 블록 → join bound 250ms **항상 만료** → 동기작업 + 250ms

| run | 국면 | n | 부하(1분) 전→후 | min | p50 | p90 | p99 | max |
|---|---|---|---|---|---|---|---|---|
| run1 | FAST | 100 | 9.12 → (연속) | 193.19ms | 203.68ms | 235.18ms | 279.04ms | 311.57ms |
| run1 | BLOCK | 40 | (연속) → 12.32 | 438.41ms | 453.33ms | 466.12ms | 472.27ms | 499.20ms |
| run2 | FAST | 300 | 7.95 → 25.14 | 181.61ms | 218.39ms | 293.99ms | 444.40ms | 539.75ms |
| run3 | BLOCK | 120 | 23.23 → 14.64 | 434.93ms | 458.31ms | 485.75ms | 546.35ms | 646.52ms |

**분해 관계 성립 (독립 2회 재현):**

```
run1: BLOCK p50 453.33 − FAST p50 203.68 = 249.65ms  ≈ joinBound 250ms
run2/3: BLOCK p50 458.31 − FAST p50 218.39 = 239.92ms ≈ joinBound 250ms
```

즉 `BLOCK ≈ 동기작업 + 250ms`가 결정론적으로 성립하고, FAST는 동기작업의 대리값이다.

**따라서 예산 산술:**

```
테스트 예산                       500ms
이 테스트에서 join bound 항상 만료  −250ms
─────────────────────────────────────────
동기 작업에 남는 실질 여유          250ms
관측된 동기 작업 중앙값             203.68 ~ 218.39ms  (여유의 82~87%)
중앙값에서 실패까지 남는 마진        약 32 ~ 46ms
```

**적색 재현 (세션 내 최초):**

```
T662 BLOCK over-500ms-budget: 4/120 = 3.3%   (부하 23.23 → 14.64)
T662 BLOCK over-500ms-budget: 0/40  = 0.0%   (부하 9.12 → 12.32, max 499.20ms — 예산까지 0.8ms)
```

n=40 국면은 통과했으나 **최댓값이 예산에서 0.8ms 떨어진 지점**에 있었다. 통과가 여유를 뜻하지 않는다.

### (추가 소견) 단정이 자기가 이름 붙인 실패를 판별하지 못한다

실패 메시지는 `expected deferred (non-blocking) return`이다. 그러나 지연이 실제로 실패했다면 `Handle`은 주입 블록 전체를 기다려 **약 2,000ms**가 되어야 한다.

BLOCK 관측 **160회**(40+120) 전량의 범위:

```
min 434.93ms   max 646.52ms
```

**2,000ms 근처가 단 한 번도 없다.** 즉 지연은 160/160 정상 동작했고, 붉어진 4회는 전부 `동기작업 + 250ms`가 500ms를 넘긴 것이다. 이 단정은 지연 실패로 붉어진 적이 없으며, 구조상 붉어질 수 없는 것이 아니라 **붉어져도 구분되지 않는다** — 임계 500ms는 정상 중앙값(453~458ms)의 1.1배 지점에 있고, 판별해야 할 동기 실행(2,000ms)에서는 4배나 떨어져 있다.

`session_start.go:1531-1533`의 근거 주석은 이렇게 적혀 있다:

> `The existing TestSessionStart_DeferredScanDoesNotBlockReturn asserts Handle returns in <500ms; 250ms preserves that margin even when the injected slow scan forces the bound to elapse.`

이 추론이 성립하려면 나머지 동기 작업이 250ms 안에 들어야 한다. 측정값(p50 218ms, p90 294ms, max 540ms)은 그 전제가 중앙값에서 겨우 성립하고 상위 십분위에서 깨짐을 보인다. **주석의 "preserves that margin"은 측정되지 않은 전제였다.**

---

## Baseline-attribution

- 트리: `fb8bfff95` (`git rev-parse --short HEAD`, 워크트리 `.claude/worktrees/t662`)
- (A) 명령: `git rev-parse <commit>:internal/hook`, `git diff --stat eabce7444 63fa51cf3` — 이 트리에서, 이번 실행에서 수행
- (B) 명령: `T662_N=<n> go test ./internal/hook/ -run 'TestT662Probe(Fast|Block)' -count=1 -v -timeout 900s`
- 원본 로그: `.moai/reports/t662/probe-run1.log`, `probe-run2.log`, `probe-run3.log` (부하값 각 로그 선두/말미에 기록)
- 부하 게이트: 착수 시각 7.00, 측정 각 국면 전후 위 표에 기재

### N 선정 근거

- **(A) N=0.** 트리 해시 동일성이 코드 차이 부재를 기계적으로 확정하므로 재실행 표본은 정보를 0만큼 준다. 돌렸다면 머신을 쟀을 것이다.
- **(B) FAST N=300 / BLOCK N=120.** 관측하려는 것은 중앙값이 아니라 상위 꼬리다. 기존 표본의 적색률이 대략 1/8~1/10이므로 꼬리를 최소 십여 회 잡으려면 100회대가 필요하다. FAST는 1회 약 0.2초라 300회가 약 65초, BLOCK은 1회 최소 0.25초라 120회가 약 55초 — 둘 다 단일 슬롯 안에서 반복 가능한 비용이다.

---

## Gaps — 관측하지 않은 것

- **FAST의 예측 적색률(26%)이 BLOCK의 관측 적색률(3.3%)과 8배 어긋난다. 이 불일치는 해소하지 못했다.** 원인 가설(미검증): FAST는 1회 0.2초로 300회를 돌려 임시 디렉터리 생성·GC 회전율이 BLOCK(1회 0.25초 이상)보다 훨씬 높아, **프로브 자신이 만든 부하 프로파일이 두 국면에서 달랐다.** 실제로 run2에서 부하가 7.95 → 25.14로 올랐고 그 상승분에 프로브 자신의 기여가 얼마인지 분리하지 않았다. 따라서 **FAST 꼬리를 BLOCK 꼬리의 예측치로 쓰지 말 것.** p50 수준의 분해 관계(249ms/240ms)만 두 국면에서 재현됐다.
- **동기 작업 200ms의 내역을 분해하지 않았다.** `Handle`의 어느 단계가 그 200ms를 쓰는지 계측하지 않았다 — 프로덕션 코드 계측이 필요해 이번 범위 밖에 두었다. 예산을 넓히는 것과 동기 작업을 줄이는 것 중 무엇이 옳은지는 이 분해 없이 정할 수 없다.
- **CI(리눅스 러너)에서 재측정하지 않았다.** 위 수치는 전부 이 darwin 개발 머신 것이다. 중앙값이 예산의 91%(458/500)인 이상 CI 머신 성능에 따라 적색률이 크게 달라질 수 있으나 확인하지 않았다.
- **다른 8개 초록 관측(부하 6.42~18.13)을 재현하지 않았다.** 기존 표본을 인용만 했고 이 트리에서 다시 재지 않았다.
- **수리하지 않았다.** 프로덕션 코드/테스트 변경 0건. 아래 제안은 제안일 뿐 검증되지 않았다.

## Residual-risk

- 위 측정은 전부 `t.TempDir()` 기반 빈 프로젝트에서 나왔다. 실제 저장소에서 `Handle`의 동기 작업은 더 무거울 수 있고, 그러면 마진은 여기 적은 32~46ms보다 얇다.
- 적색률은 부하 단조 함수가 아님이 기존 표본(14.30 빨강 / 14.08 초록)과 이번 run1(부하 9~12에서 max 499.20ms)에서 모두 확인된다. 어떤 임계를 고르든 "부하가 낮으면 안전"이라는 보증은 없다.
- 프로브 파일은 삭제했다. 재현하려면 이 verdict의 설계(FAST/BLOCK 분해)를 다시 구현해야 한다.

---

## 제안 (미실행 — 리드 판정 대기)

이 카드는 "결함인지 예산이 얇은 것인지 가른다"였고 답은 **예산**이다. 다만 측정 과정에서 더 근본적인 것이 나왔다: **임계값이 잘못된 양을 재고 있다.**

단정의 의도는 "동기 실행이 아니라 지연 실행임"의 판별이다. 판별해야 할 두 상태의 실측 거리는 이렇다:

```
지연 동작 (정상)   434.93 ~ 646.52ms   ← 160/160 관측
동기 실행 (결함)   약 2,000ms          ← 주입 블록 크기, 0회 관측
현재 임계          500ms               ← 정상 분포 안에 박혀 있다
```

임계를 두 상태 사이(예: **1,000ms**)로 옮기면 판별력은 보존되고(동기 실행 2,000ms는 여전히 실패) 거짓 적색 띠는 사라진다. 정상 최댓값 646ms 대비 1.55배, 결함값 2,000ms 대비 0.5배로 양쪽 마진이 대칭에 가깝다.

입력 지연(input lag) 예산 자체가 지켜야 할 계약이라면, 그것은 **이름과 실패 메시지가 다른 별개의 단정**으로 세워야 하며 임계는 측정에서 유도해야 한다 — 지금처럼 지연 판별 단정에 얹어 두면 둘 다 제대로 못 한다.

어느 쪽으로 갈지, 혹은 동기 작업 200ms 내역 분해를 먼저 할지 지시 주시면 이어가겠습니다.
