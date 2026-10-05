# t407 판정 — codemaps 출처 스탬프 고아 판정 (Graph Freshness)

**브랜치**: `WT-codemaps-orphan-stamp` @ `84b7b5ce6` (base `origin/develop` `9145806d8`)
**변경**: `.github/workflows/graph-freshness.yml` 1파일 (+27/−6)
**일시**: 2026-09-02 · 측정 트리: 이 워크트리 HEAD + `origin/{main,develop,release/v3.1.4}` fetch 시점 값

## 판정 요약

**가드가 결함이 아니라 가드의 판정식이 결함이다.** 스탬프-대-base(base-ancestry) 검사는 squash 머지 시대의 생존 술어다. git-flow 전환 후 release PR은 머지 커밋으로 들어가며, develop 앵커 스탬프는 머지가 착지하기 전까지 base 조상일 수 없다 — 즉 **모든 release PR이 구조적으로 적색**이다(CI run 33382844542 실측). 스탬프 방식(develop에서 재생성)은 정상 운용이므로 손대지 않았다.

## Claim / Evidence / Baseline

### C1 — 7fc0af324 출처 규명 (카드 범위 1)

- **Claim**: 스탬프 커밋은 develop 배치 정리 머지이며, main 조상이 아닌 것은 git-flow 구조의 필연이다.
- **Evidence**:
  - `git log -1 7fc0af324` → `merge(develop): absorb fef7a4b9b for batch cleanup (codemaps + memory index)` (2026-08-31 15:45 +0900)
  - `provenance.json`(develop 헤드) → `tree_root: …/worktrees/t287`, `generated_at: 2026-08-31T06:57:41Z` — t287 워크트리에서 codemaps 재생성 후 develop 흡수
  - 조상 검사: develop=ANCESTOR, release/v3.1.4=ANCESTOR, **main=NOT-ANCESTOR**
  - main이 develop 커밋을 객체로 받는 유일 경로는 release 머지 커밋(직행 PR #1670·#1676·#1688은 전부 squash — `(PR번호)` 접미사 실측). squash는 새 커밋을 만들므로 develop 쪽 커밋은 main 역사에 절대 도달하지 않는다.
- **Baseline**: 본 세션 `git fetch origin develop/main` 후 로컬 산출.

### C2 — 적색 분포 실측 (카드 범위 2: #1684·#1681)

- **Claim**: orphan-bound 적색은 현재 #1685에만 나오지만, 구조상 모든 release PR에 재발한다. #1684·#1681은 별개 축(staleness)으로 적색이다.
- **Evidence** (잡 로그 직접 판독):
  - **#1685** (release→main, run 33382844542): 가드 적색 — `stamp 7fc0af324 … is NOT an ancestor of PR base origin/main`
  - **#1684** (dependabot→main, run 33357932822): `graph check … verdict=stale value=45 threshold=40` — **가드 단계가 아예 없음**(잡 로그에 Guard 스텝 부재 실측). 이유: main의 workflow는 `6786c3fa4`(t250 원본, 가드 없음)이고 가드는 develop에서 `e4eb15ea4`(t291)→`df4466d12`(t294)로 착륙해 main 미도달.
  - **#1681** (fix→main, run 33090462635, 08-27): `verdict=stale value=51 threshold=40` — 동일한 구형 workflow.
  - **main 푸시 자체** (run 33489007959, 09-01, 헤드 `7ad9f8534`): `stale value=45` — **main은 t289 squash(08-27, #1668) 이후 계속 적색**. main의 `provenance.json` 스탬프 `a995e58fa`는 main 조상이 아님(고아+내용 45개 뒤처짐 — 이중 사망).
  - 측정법 검증: `git diff --name-only a995e58fa 7ad9f8534 -- internal cmd pkg` = 45파일 — CI 값 45와 정확히 일치.
- **Baseline**: `gh run view` 로그 판독(2026-09-02) + 로컬 git 조상 검사.

### C3 — 처방: 재정의 (카드 범위 3) — 재스탬프 계열은 측정으로 전부 사멸

- **Claim**: 일회성 재스탬프로는 이 계열을 닫을 수 없고, 가드 판정식의 레인 분리가 유일한 녹색 경로다.
- **Evidence** (두 재스탬프 변형 모두 실측 탈락):
  1. **generation-HEAD 스탬프**(현행 7fc0af324 유지): 가드 적색(C1).
  2. **merge-base 스탬프**(`moai graph stamp codemaps --commit "$(git merge-base HEAD origin/main)"` — REQ-SR-010 레시피): merge-base=`48239c7dc`이므로 가드는 통과하지만, 그 트리→release 헤드(`26898312e`) 사이 described-worthy(.go 비테스트) 변경 **91파일 ≥ 임계 40** → freshness 적색. 측정: `git diff --name-only 48239c7dc 26898312e -- internal cmd pkg` 필터 계수.
  - 반면 **현행 스탬프(7fc0af324) 기준 release 헤드 drift는 10 raw → described-worthy 필터 후 1**(`pkg/version/version.go` 뿐) — freshness 통과.
- **적용한 수정**: `release/*` 헤드 레인은 도달성 판정 대상을 `HEAD`(pull_request 체크아웃 = 머지 프리뷰 `refs/pull/N/merge`, 헤드 브랜치 전역 역사 포함)로, 그 외 레인은 기존 `origin/${GITHUB_BASE_REF}` 유지. 머지 커밋 생존 술어 = HEAD 도달성이고 squash 생존 술어 = base 조상이므로, 레인별로 정확한 생존 조건을 검사하게 된다.
- **수정 후 #1685 예상**: 가드 통과 + freshness 1/40 통과 = **완전 녹색**. 머지(머지 커밋) 후 main은 스탬프를 자기 조상으로 흡수 + drift 1 → **main의 6일째 적색(stale 45)도 자가 치유**.

### C4 — 회귀 단언 (카드 기준: "초록 한 번은 근거가 아니다")

3-시나리오 실측 (`git merge-base --is-ancestor`, rc로 단언):

| 시나리오 | 스탬프 | 판정 대상 | 기대 | 실측 |
|---|---|---|---|---|
| S1 release 레인 | 7fc0af324 (develop) | release/v3.1.4 헤드 | 통과 | rc=0 ✓ |
| S2 squash 레인 보존 | 7fc0af324 (develop) | origin/main (base) | 거부 | rc=1 ✓ |
| S3 진짜 고아 (양 레인) | def79d6fa (t409 미병합 브랜치 헤드) | release 헤드 / main | 양쪽 거부 | rc=1, rc=1 ✓ |

- `actionlint` 통과. 머지 커밋 착지 시 조상 이전(base-tip, release-head 부모)은 git 구조 불변식이며, 최종 증거는 #1685 머지 후 main 푸시 런의 녹색(리드 판독 몫).

## Gaps (관측하지 않은 것)

- #1685에 수정이 실린 **후의 실제 CI 녹색** — 수정은 release 브랜치에 아직 미적용(아래 착지 경로). 로컬 시뮬레이션이지 원격 런 관측이 아니다.
- GitHub 머지 프리뷰(`refs/pull/N/merge`)의 HEAD 도달성은 GitHub 구조 불변식에 의존(체크아웃 문서 근거), 로컬 재현 불가.
- release PR이 doctrine(머지 커밋)을 어기고 squash로 눌리는 경우의 원격 거동 — develop이 상설 ref라 객체 보존은 되지만 실측 안 함.

## Residual-risk

- `GITHUB_HEAD_REF` 접두 판정(`release/`)은 레인 식별을 브랜치명 관례에 의존한다. release/*가 아닌 이름으로 main에 머지커밋-랜딩되는 PR이 생기면 그 PR은 squash 술어(base)를 적용받는다 — 안전 쪽 오류(거부)이며 침묵 녹색이 아니다.
- main 푸시 런의 객체 존재 검사는 fetch-depth 0이 모든 브랜치를 당려와 완전한 고아 탐지가 아니다 — 이는 수정 전부터의 한계로 본 카드 범위 밖.

## 착지 경로 (리드 소관)

1. **develop 통합**: `WT-codemaps-orphan-stamp` @ `84b7b5ce6` → 통합 창 경유 develop 머지(레인 표준 요청).
2. **#1685 자체 해제(선택)**: 가드 수정이 develop에 착지하면 release/v3.1.4에 `git merge <84b7b5ce6을 포함한 develop>` 또는 cherry-pick 후 재push → CI 재실행에서 가드+freshness 동시 통과 예상. 미적용 시 #1685는 다음 release(v3.1.5+)부터 혜택.
3. 배포 금지 조항(t406 동일) 준수 — 태그/릴리스 작업 없음.
