# t247 판정 — DROPPED (전제 전면 반증)

- 카드: t247 — PR #1600 분할(리뷰 불가 크기)
- 레인: lane-4
- 판정 트리: `.claude/worktrees/t247`, 브랜치 `WT-pr1600-split`, HEAD `b7462203a`
- 측정 시각: 2026-09-02
- 판정: **DROPPED** — 카드가 세운 전제 4건이 모두 현재 상태에서 거짓이고, 그 결과 범위 (1)~(4)가 전부 성립하지 않는다.

---

## Claim

카드 본문의 전제는 다음 네 가지였다(리드 측정 2026-08-24 기준).

| # | 카드가 세운 전제 | 현재 실측 | 판정 |
|---|---|---|---|
| P1 | PR #1600이 열려 있고 리뷰를 못 받는 상태 | `state = MERGED`, 2026-08-25T03:23:18Z 머지 | **거짓** |
| P2 | 497파일 — CodeRabbit 150 상한 초과 | `changedFiles = 10` | **거짓** |
| P3 | CodeRabbit이 "Review skipped: 497 files exceed the limit of 150"으로 리뷰를 건너뜀 | 결합 상태 `state=success` / `Review completed`, `Merge Risk:` 줄이 head와 접두사 일치 | **거짓** |
| P4 | `mergeable = CONFLICTING` — 충돌이 누적 중 | 머지 완료. 머지 커밋 `07a4ea0ed`가 `origin/main`의 조상(rc=0). 브랜치 `WT-server-version`은 원격에서 삭제됨 | **거짓** |

전제가 무너지면서 카드가 지시한 작업 네 가지가 모두 대상을 잃는다.

- (1) 497파일의 구성을 갈라 생성물과 손으로 쓴 변경을 분리 → 실제 10파일. 가를 대량 생성물이 없다.
- (2) 충돌 해소 → 이미 머지됨. 해소할 충돌이 없다.
- (3) 리뷰 가능한 단위로 분할 → 10파일에 리뷰가 이미 완료됐다. 분할할 PR이 없다.
- (4) 분할 후 각 PR이 150파일 상한 아래인지 확인 → 10 < 150. 확인할 분할본이 없다.

카드가 경계한 대표 mutant("파일 수만 줄이고 같은 변경을 한 PR에 몰아넣는 분할")도 자연히 소멸한다 — 분할 자체가 일어날 수 없다.

## Evidence

명령과 그 출력 원문은 같은 디렉터리의 로그에 있다.

**E1 — PR 상태 전체** (`pr1600-state.json`)

```
$ gh pr view 1600 --json number,title,state,mergeable,headRefName,headRefOid,\
    baseRefName,changedFiles,additions,deletions,mergedAt,mergedBy,url
{"additions":597,"baseRefName":"main","changedFiles":10,"deletions":7,
 "headRefName":"WT-server-version",
 "headRefOid":"33bdaa05dfc3298fb50ac3186eb0651a58730a73",
 "mergeable":"UNKNOWN","mergedAt":"2026-08-25T03:23:18Z",
 "mergedBy":{"login":"GoosLab",...},"number":1600,"state":"MERGED",
 "title":"feat(mcp): make a running MCP server's build version visible",
 "url":"https://github.com/modu-ai/moai-adk/pull/1600"}
```

제목(`feat(mcp): make a running MCP server's build version visible`)과 브랜치
(`WT-server-version`) 모두 카드가 지목한 PR과 일치한다 — 동일 PR을 재고 있다.
`mergeable`이 `UNKNOWN`인 것은 머지된 PR에서 GitHub이 병합 가능성을 더 계산하지 않기
때문이며, 카드가 인용한 `CONFLICTING`이 아니다.

**E2 — 변경 파일 10개 전량** (`A-filelist.log`)

```
CHANGELOG.md                              +2/-0
internal/cli/doctor.go                    +1/-0
internal/cli/doctor_mcp_version.go        +104/-0
internal/cli/doctor_mcp_version_test.go   +149/-0
internal/cli/mcp_server.go                +28/-1
internal/cli/mcp_server_runtime.go        +150/-0
internal/cli/mcp_server_runtime_test.go   +154/-0
internal/cli/testdata/doctor-dark.golden  +3/-2
internal/cli/testdata/doctor-light.golden +3/-2
internal/cli/testdata/doctor-nocolor.golden +3/-2
```

전량 열거다(표본이 아니다). 소스 4 + 테스트 2 + 골든 3 + CHANGELOG 1 — 대량 생성물도,
재임베드 산출물도 없다. 카드가 가르라고 한 "생성물 대 손으로 쓴 변경"의 축 자체가 없다.

**E3 — 머지 커밋 조상 판정** (`B-ancestry.log`)

```
$ git merge-base --is-ancestor 07a4ea0ed73f79aa45f3011cb53d278f2ee7ae8e origin/main
rc=0
```

`origin/main`은 이 트리에서 `git fetch origin main` 직후 판독했다.

**E4 — CodeRabbit 2조건 판독** (`C-coderabbit-status.log`, `D-mergerisk.log`)

조건 ①: 결합 상태 엔드포인트

```
$ gh api "repos/modu-ai/moai-adk/commits/33bdaa05dfc3298fb50ac3186eb0651a58730a73/status" \
    --jq '.statuses[] | select(.context=="CodeRabbit") | ...'
state=success description=Review completed
```

조건 ②: `Merge Risk:` 줄의 커밋 접두사가 현재 `headRefOid`와 일치

```
Merge Risk:** _🟡 Moderate_ · up to `33bda`
```

`33bda` = `headRefOid` `33bdaa05dfc3298fb50ac3186eb0651a58730a73`의 접두사. 두 조건이
함께 성립하므로 이 PR은 실제로 리뷰를 받았다 — 카드가 인용한 "Review skipped"는 현재
head에 해당하지 않는다. 보강: 리뷰 제출 기록 2건(`coderabbitai[bot]`, 2026-08-23T10:57:43Z /
2026-08-24T21:31:04Z)이 존재한다.

**E5 — 원격 브랜치 부재** (`E-remote-branch.log`)

```
$ git ls-remote --heads origin WT-server-version
(출력 없음)
exit=0
```

출력이 비었고 종료 코드는 0 — 조회는 성공했고 그 브랜치가 원격에 없다. 명령 실패로 인한
빈 출력이 아니다.

## Baseline-attribution

모든 측정은 이 판정 트리(`.claude/worktrees/t247`, HEAD `b7462203a`, 브랜치
`WT-pr1600-split`)에서 2026-09-02에 실행했다. 리드가 2026-08-24에 잰 값(497파일 ·
CONFLICTING · Review skipped)은 이 판정의 baseline이 아니라 **반증 대상**이며, 여기에 그대로
옮겨 적지 않았다. `origin/main`은 판정 직전 `git fetch origin main`으로 갱신한 뒤 읽었다.

전제가 반전된 시점은 특정 가능하다. 리드 측정(2026-08-24) 이후 PR이 갱신됐고
2026-08-25T03:23:18Z에 머지됐다 — 즉 카드가 발행된 뒤 카드 없이 해소된 경우이지, 리드의
당시 측정이 틀렸다는 뜻이 아니다.

## Gaps

관측하지 않은 것을 명시한다.

- **G1** 2026-08-24의 497파일이 어떤 경로로 10파일이 됐는지 — 브랜치 리셋인지, 리베이스인지,
  누군가 손으로 정리한 것인지 — 를 추적하지 않았다. 원격 브랜치가 삭제돼 reflog가 없고,
  카드 판정(작업이 남아 있는가)에 이 경로는 영향을 주지 않는다.
- **G2** 머지된 10파일의 내용 자체는 검토하지 않았다. 이 카드의 범위는 "리뷰 가능한 크기로
  분할"이지 그 변경의 품질 심사가 아니다. 품질은 이미 CodeRabbit 리뷰와 머지 게이트가 봤다.
- **G3** 497파일 상태에서 함께 딸려 있던 파일들이 다른 PR로 옮겨져 아직 열려 있는지 확인하지
  않았다. 그런 PR이 있다면 그것은 t247이 아니라 별도 카드의 대상이다.
- **G4** 로컬·CI 어느 쪽 테스트도 돌리지 않았다. 이 판정은 코드를 바꾸지 않으므로 회귀 표면이
  없다.

## Residual-risk

- **R1** 관측한 것은 GitHub API가 보고하는 현재 상태다. PR이 머지 후 되돌려지는(revert) 일은
  가능하지만, 그 경우 새 카드가 필요한 새 사실이지 이 판정을 소급해 무효로 만들지 않는다.
- **R2** G3이 열려 있다 — 원래 497파일에 섞여 있던 변경 중 일부가 아직 리뷰를 못 받은 다른
  PR에 남아 있을 수 있다. 그렇다면 "리뷰 불가 크기"라는 **문제 자체**는 다른 좌표에 살아 있고,
  이 카드가 지목한 좌표(PR #1600)에서만 사라진 것이다. 리드가 이 축을 새 카드로 세울지 판단할
  수 있도록 여기에 남긴다.
- **R3** 큐 변경은 하지 않았다. `moai todo drop t247`은 운영자·리드의 행위이며, 이 판정서는
  그 판단의 근거일 뿐이다.

---

## 리드 조치 요청

1. t247을 **DROPPED**로 닫을 것 — 남은 작업이 없다.
2. R2(G3)를 새 카드로 세울지 판정할 것 — 세운다면 범위는 "497파일에 섞여 있던 변경 중
   아직 리뷰를 못 받은 잔여분 식별"이며, PR #1600은 그 대상이 아니다.
