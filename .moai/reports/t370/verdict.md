# t370 — verdict (읽기 전용 조사, 수리 없음)

card: t370 (재정의: t354 수리 이후 잔여) · 조사 전용 · branch `WT-race-residual`
worktree `.claude/worktrees/t370` · 측정 트리 `origin/develop` = `1e5199b88`
수리 병합 `728f91006` 은 판독한 14개 head 전부의 조상.
측정 원문: `.moai/reports/t370/measurements.md`

**수리하지 않았다.** 코드 한 줄도 고치지 않았고 커밋도 없다.

---

## Claim

t354 의 수리는 **착지했으나 실패를 닫지 못했다.** 수리 이후 develop 의 CI `Race Test` 는
14회 중 12회에서 `TestConcurrencyStress` 로 붉다.

원인은 새로 생긴 결함이 아니라 **수리가 자기 진단을 절반만 따라간 것**이다. 두 축 모두 실측됐다:

1. **예산이 실행 머신에 맞춰 자라지 않는다.** `boardLockWaitBudget` 은 컴파일 시점 상수
   1.65s 인데, 요구 대기는 `48 × per-mutation cost` 로 머신 속도에 비례한다.
   버티는 조건은 `cost < 34.4ms`. sizing 입력값이 정확히 `33ms` — 마진이 사실상 0이다.
   CI `-race` 실측 cost 는 **42~105ms** 로 임계를 1.2~3.1배 넘는다.
2. **수리가 지목한 비율이 여전히 뒤집혀 있다.** t354 는 "재시도 지연 25ms < 변이 비용 33ms
   라 깨어난 경합자가 거의 항상 잠긴 락을 본다"고 적었다. 수리 후 per-attempt 대기
   기댓값은 **27.5ms**(상한 50ms)로, 실측 변이 비용 42~105ms 보다 여전히 작다.
   지터는 위상 고정을 깼지만 "깨어나면 잠겨 있다"는 성질 자체는 남았다.

**data race 가 아니다** — 판독한 전 run 에서 `DATA RACE` 0건. t354 서술 인용이 아니라
이번 조사의 측정이다.

**부하 의존이지만, 러너의 거친 부하가 아니다.** 갈리는 축은 `-race` 유무다.
같은 커밋·같은 러너 클래스에서 `TestConcurrencyStress` 는 비-race 잡 5/5 통과
(0.52~0.76s), race 잡 5/5 실패.

> 한정 주의: "`-race` 잡 전용"은 **이 테스트에 한해서만** 참이다. 같은 기간
> `TestBinaryLag_OneSeamServesBothSurfaces` 는 세 run 에서 비-race 잡에서도 붉다 —
> 별개 결함이고 이 축과 무관하다.

---

## Evidence

### 1) `-race` 가 판별 인자다 — 같은 커밋 안에서의 대조

| head | Race Test `TestConcurrencyStress` | 실패 adds | 비-race `Test (ubuntu)` 같은 테스트 |
|---|---|---|---|
| `d7010f86a` | FAIL | **8/48** | 잡 `success` |
| `52c3fe590` | FAIL 2.15s | 1/48 | **pass 0.76s** |
| `1728136c7` | FAIL 4.30s | 7/48 | **pass 0.52s** |
| `d8a1a8e4e` | FAIL 1.97s | 1/48 | **pass 0.53s** |
| `1e5199b88` | FAIL 2.88s | 3/48 | **pass 0.54s** |

로컬 darwin 대조군(`go test -json -race -count=1 ./internal/kanban/`): rc=0,
`TestConcurrencyStress pass 0.84`.

### 2) 수리 이후 발화율 — 14 run 중 12 붉음

`git rev-list 728f91006..origin/develop` (230 커밋) 대조로 자손을 가려낸
non-cancelled CI run 14개. `Race Test` 잡 결론과 실패 테스트를 전부 판독했다.

붉음 12: `48d8ef4be` `1a635aea8` `91ec14dbe` `a6bbbf82b` `15453140a` `ee50984ab`
`68ecbfe4a` `d7010f86a` `52c3fe590` `1728136c7` `d8a1a8e4e` `1e5199b88`
통과 2: `51daada00`(잡 전체 `success`) · `c6aa61346`(잡은 붉지만 원인은 다른 테스트)

`run_attempt` 는 판독한 다섯 head 전부 **1** — 숨은 attempt-1 실패 없음.

### 3) `DATA RACE` — 직접 관측, 전 run 0

스트림 4건 `grep -c "DATA RACE"` → 0·0·0·0.
로그 9건(`--log-failed`) 동일 grep → 전부 0.

### 4) 산술 — 계산과 관측 둘 다

계산: 테스트는 `writers=8 × addsPerWriter=6 = 48` 회 변이를 하나의 flock 으로 직렬화한다.
한 writer 의 최악 대기 상한은 `48 × cost`. 예산 1.65s 가 버티는 조건은 `cost < 34.4ms`.

`boardLockWaitBudget = boardLockSupportedWriters(10) × boardLockCIMutationCost(33ms) × boardLockHeadroom(5)`

`headroom = 5` 가 5배 여유로 읽히지만 아니다. 곱해지는 항이 **직렬화되는 변이 수(48)** 가
아니라 **지원 레인 수(10)** 다. `10 × 5 = 50 ≈ 48` 이라 우연히 "48회를 33ms 로 버티는 예산"과
같아졌을 뿐, 헤드룸은 이 환산에 다 쓰였고 남는 마진이 없다.

관측(역산, `성공 변이 수 ÷ 테스트 elapsed`):

| head | elapsed | 성공 변이 | cost |
|---|---|---|---|
| `d8a1a8e4e` | 1.97s | 47 | ≈ 42ms |
| `52c3fe590` | 2.15s | 47 | ≈ 46ms |
| `1e5199b88` | 2.88s | 45 | ≈ 64ms |
| `1728136c7` | 4.30s | 41 | ≈ 105ms |

대조: CI 비-race 11~16ms, 로컬 darwin -race 17.5ms — 둘 다 임계 34.4ms 아래이고, 둘 다 통과한다.

실패한 테스트의 elapsed 는 **12건 전부** 예산 1.65s 를 넘는다 —
1.94 · 1.97 · 2.15 · 2.16 · 2.16 · 2.18 · 2.32 · 2.37 · 2.41 · 2.88 · 4.03 · 4.30 (초).
최솟값 1.94s 조차 예산보다 크다. 실패 writer 가 예산을 **소진하고** 죽었다는 것과
12건 모두에서 일치하며, 예외가 하나도 없다.
(넷은 스트림 측정, 나머지 여덟은 병렬 판독 보고 인용 — 출처 구분은 measurements.md.)

### 5) 거친 러너 부하는 판별 인자가 아니다

`Race Test` 잡 총 소요 8m30s~9m23s, kanban 패키지 elapsed 15.0~19.6s — 넷이 비슷한데
실패 건수는 1·1·3·7 로 갈린다. 편차는 러너 속도가 아니라 불공정 락의 꼬리에서 온다.

---

## Baseline-attribution

- 트리: `.claude/worktrees/t370`, `git rev-parse --show-toplevel` 로 확인
- ref: `git fetch origin develop` 후 `origin/develop` = `1e5199b88` (이번 실행에서 fetch)
- 조상 판정: `git rev-list 728f91006..origin/develop` → 230 커밋, 후보 9개 전부 포함
- CI 측정원: `test-stream-ci-race-ubuntu-latest` / `test-stream-ci-test-ubuntu-latest`
  아티팩트(`go test -json` 스트림), 없는 run 은 `gh run view <id> --log-failed`
- 코드: `git show origin/develop:internal/kanban/board_store.go` 및
  `git show a680ea6e8` — 인용한 상수는 현재 `origin/develop` 의 값

---

## 원 카드 질문에 대한 답

- **질문 2 (실패 테스트가 하나인가)** — 아니다. `52c3fe590` 에는 두 건이 있다:
  `TestConcurrencyStress` + `TestGitDiffNameCount_Predicate`. 리드가 미판독이라 밝힌
  두 head 중 하나였다. 나머지 판독분에서도 `TestBinaryLag_OneSeamServesBothSurfaces` 와
  `TestGitDiffNameCount_Predicate` 가 같은 잡에서 함께 붉은 run 이 여럿이다.
  **별개 계열 — 이 카드에 귀속하지 않는다** (관계는 아래).
- **질문 3 (성격)** — data race 아님. 락 대기 소진. 전 run `DATA RACE` 0.
- **질문 4 (부하 의존 재현 조건)** — `-race` + ubuntu-latest 러너 클래스에서 재현되고,
  로컬 darwin `-race` 로는 재현되지 않는다. **로컬 부하는 만들지 않았다**(지시대로).
  재현 조건을 한 줄로: per-mutation cost 가 34.4ms 를 넘는 환경.

---

## 수리 범위 — 정하기만 하고, 고치지 않았다

세 갈래이고 성격이 다르다. 판정은 리드/운영자 몫.

**A. 예산을 머신에 맞춰 재도출** (가장 작음, 근본 아님)
`boardLockCIMutationCost` 를 42~105ms 실측 상단에 맞춰 올리거나, 예산을
`직렬화 변이 수 × 관측 cost` 로 재정의. 곱해지는 항을 지원 레인 수(10)가 아니라
가드가 실제로 직렬화하는 변이 수로 바꾸는 것이 요점. **한계**: 다시 상수이고,
더 느린 러너가 나오면 같은 방식으로 깨진다. 꼬리를 얇게 할 뿐 없애지 못한다.

**B. 락에 공정성을 준다** (근본, 가장 큼)
같은 프로세스 안의 writer 를 in-process mutex 로 먼저 직렬화하고 flock 은 교차 프로세스
경계에만 남기면, 프로세스 내 경합에서 기아가 사라진다(이 테스트의 8 writer 는 전부
같은 프로세스다). **한계**: 교차 프로세스 경합의 기아는 그대로. 범위가 `internal/kanban`
락 계층 전체로 번진다.

**C. 가드의 판정 기준을 분리한다** (테스트 쪽)
이 테스트가 지키는 것은 불변식(lost update·id 충돌·last_seq)이지 획득 지연이 아니다 —
그리고 불변식은 12번 붉은 run 전부에서 **한 번도 깨지지 않았다**. 락 획득 실패를
테스트 실패로 세지 말고, 지연은 별도의 명시적 예산 가드로 뽑는 안.
**한계**: 예산 회귀를 놓칠 수 있고, "규칙을 끈 것"과 구별하려면 예산 가드가 실제로
RED 를 낼 수 있음을 심어서 보여야 한다.

**어느 갈래든 판정에 반복 관측이 필요하다.** 수리 이후에도 초록 run 이 존재하므로
(`51daada00`), 한 번 초록은 닫힘의 근거가 못 된다. 이것이 t354 가 한 번 밟은 자리다.

---

## 이웃 카드와의 관계 (관계만 — 흡수·중복 판정 아님)

- **t354** — 1차 수리. 착지(`728f91006`)했고 산출물도 있으나, 스스로 선언한 Gap
  ("로컬 증거는 CI 실패가 닫혔음을 세우지 못한다") 이 그대로 열려 있다.
  이 카드의 측정이 그 Gap 을 메운 결과는 **닫히지 않았다** 이다.
- **t352** — `TestGitDiffNameCount_Predicate`. `c6aa61346`·`1a635aea8`·`52c3fe590`
  의 같은 잡에서 함께 붉다.
- **t326/t366 계열** — `TestBinaryLag_OneSeamServesBothSurfaces`.
  `48d8ef4be`·`91ec14dbe`·`a6bbbf82b` 에서 함께 붉은데, **비-race `Test (ubuntu)` 잡에서도**
  붉다(세 run 의 `Test (ubuntu)` conclusion 이 `failure` 인 이유). 이 카드의 `-race` 축과
  무관한 별개 결함이다.
- **t278** — `TestConfigChange_RT005ReloadIntegration` 은 이번 14 run 어디에서도
  실패로 나타나지 않았다. 다만 t278 이 적은 양상("로컬 darwin 초록 + CI ubuntu 붉음 +
  수 초 내 실패")은 이 카드가 실측한 축과 같은 모양이다 — t278 조치 (1)의 공통 인자
  조사가 이 측정을 재료로 쓸 수 있다.
- **t358** — 이 조사를 가능하게 한 배선. `test-stream.json.gz` 아티팩트가 있는 run 은
  `52c3fe590`(08-30 14:15) 이후뿐이고, 그 이전은 로그로만 판독됐다.

---

## Gaps — 관측하지 않은 것

- **로컬 재현을 하지 않았다.** 부하 생성 금지 지시에 따른 의도적 미측정.
  따라서 "cost 34.4ms 초과가 실패를 만든다"는 인과는 **CI 관측 + 산술**로 세운 것이고,
  통제된 실험으로 세운 것이 아니다
- per-mutation cost 는 **역산**이다. 스트림에 변이 단위 타임스탬프가 없다
- 초록 race run `51daada00` 의 `TestConcurrencyStress` elapsed 를 못 쟀다
  (아티팩트 부재). 초록일 때의 cost 를 모르므로 임계 34.4ms 를 **아래에서** 확인하지 못했다
- ubuntu-latest 의 vCPU 수 / `GOMAXPROCS` 를 확인하지 않았다.
  "`-race` 오버헤드가 코어 부족과 곱해진다"는 설명은 **가설**이다
- 스트림이 없는 9 run 은 `--log-failed` 로만 셌다. 그 출력은 실패 스텝만 담으므로
  다른 스텝의 실패는 안 보인다
- 수리 **이전** 발화율을 세지 않았다. 따라서 "수리가 발화율을 낮췄다/못 낮췄다"를
  정량으로 말할 수 없다 — 말할 수 있는 것은 "수리 이후에도 12/14 로 붉다" 까지다
- 제안한 수리 갈래 A/B/C 중 어느 것도 **시도하지 않았다**. 효과는 미검증이다
- 12건 중 8건의 elapsed·`N/48` 은 **병렬 판독 보고 인용**이고 제 스트림 측정이 아니다
  (스트림 아티팩트가 없는 run). 출처 구분은 measurements.md 마지막 절

## Residual-risk

- 역산 cost 는 실패 writer 의 대기와 성공 변이의 겹침 정도에 따라 ±가 있다.
  방향(임계 초과)은 네 head 모두 일치하지만 개별 수치는 근사다
- 이웃 카드 귀속(t352·t326/t366)은 테스트 이름과 종전 기록에 기댄 것이고,
  이번 실행에서 해당 카드를 열어 확인하지 않았다
- `-race` 축이 판별 인자라는 결론은 다섯 head 의 대조에 근거한다. 비-race 잡이
  더 긴 이력에서도 항상 초록인지는 세지 않았다
- 갈래 B(in-process mutex)는 교차 프로세스 경합을 남긴다. 실제 Factory 10 레인은
  **별개 프로세스**이므로, 이 테스트를 초록으로 만들어도 운영 경합은 안 고쳐질 수 있다 —
  갈래 선택 시 이 구분이 결정적이다
