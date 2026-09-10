# t456 — 착지 판정 렌더 비용 실측 (kickoff 전 사전 측정)

카드 [HARD]: "비용을 먼저 재십시오. 캐시·지연 계산·기존 판정 재사용 중 무엇이 맞는지 측정 후 정할 것."
본 문서는 그 측정이며, 구현 이전 단계다. Implementation Kickoff Approval 미개방 상태.

- 트리: `.claude/worktrees/t456` · 브랜치 `WT-statusline-landed` · HEAD `5107bbfff` (로컬 develop 팁)
- 측정 대상 ref: `origin/develop` = `400f37eb9`
- 측정 시각: 2026-09-03 · 로컬 darwin (CI 아님)

## Claim 1 — 착지 판정은 카드 1장당 git 서브프로세스 1회다

`GitLandedQuerier.Landed(cardID)` 는 카드 하나마다
`git log <ref> --perl-regexp --grep=\b<id>\b --oneline` 을 1회 실행한다
(`internal/kanban/prlink_landed.go`, `Landed` 본문). 카드 수에 선형이다.

Evidence — 단일 질의 1회 실측:

```
$ time git log origin/develop --perl-regexp '--grep=\bt336\b' --oneline
0.04s user 0.04s system 42% cpu 0.174 total
```

Baseline-attribution: 이 트리(`5107bbfff`), `origin/develop`=`400f37eb9`, 이번 실행에서 측정.

## Claim 2 — 현행 큐 규모에서 렌더당 약 13.9초가 붙는다

Evidence — 큐 실측:

```
$ moai todo list --limit 0 | awk '{print $2}' | sort | uniq -c
  76 picked
   4 queued
```

80장 × 0.174s ≈ **13.9s / 렌더**. picked 76장만 쳐도 13.2s.

대조: t305 는 렌더당 git spawn 을 7회에서 2회로 줄인 카드다. 렌더 경로에 80회를 얹는 것은
그 방향의 정반대이며, `internal/statusline/backlog.go` 가 명문화한 계약
("Constant-cost per render ... must never grow with the number of cards")을 정면으로 깬다.

**판정: 렌더마다 라이브 계산하는 안은 측정으로 기각된다.**

## Claim 3 — N회 질의는 1회 질의로 접힌다 (40배)

착지 판정에 필요한 정보는 "ref 이력이 이 카드 id 를 부르는가" 하나다. 카드마다 묻는 대신
이력을 한 번 훑고 토큰을 교집합하면 서브프로세스는 1회다.

Evidence:

```
$ time git log origin/develop --format='%B' > /tmp/full.txt
0.09s user 0.08s system 49% cpu 0.347 total
```

1회 0.347s vs 80회 13.9s — **40배**. 다만 0.347s 도 렌더 경로에 상주시킬 수는 없다.

## Claim 4 — 표시 오차 규모: picked 76 중 48이 이미 이력에 있다

Evidence — 셈한 기준을 `Landed` 와 일치시켰다. `--grep` 은 제목이 아니라 **전체 메시지**를
훑으므로 `%B` 로 재측정했다(제목만인 `%s` 기준으로는 38 — 기준 불일치라 폐기).

```
$ grep -oE '\bt[0-9]{1,5}\b' /tmp/full.txt | sort -u > /tmp/incommits_full.txt
$ comm -12 /tmp/picked.txt /tmp/incommits_full.txt | wc -l
48
```

즉 statusline 이 `🔄 TODO: 76/4` 를 찍는 동안 76 중 **48**은 이미 `origin/develop` 이력에
이름이 있다. 실제 in-flight 는 약 28이다.

## 권고 기전 — 새로 만들지 않는다, `github.go` 패턴을 재사용한다

`internal/statusline/github.go` 에 이미 같은 모양의 문제가 풀려 있다:

- 렌더 경로는 **작은 캐시 파일 1회 읽기**만 한다 (`resolveGitHubCounts`) — 네트워크도 git 도 안 탄다
- 신선도가 지나면 **분리된 자식 프로세스**를 띄우고 즉시 반환한다 (`maybeRefreshGitHubCounts`) — stale-while-revalidate
- stampede 가드는 캐시 파일 자신의 타임스탬프
- `isSelfInvocable` 가드가 `go test` 하에서 fork bomb 을 막는다

이 카드에 그대로 대응한다: 렌더는 landed 캐시 파일 1회 읽기(git spawn 0회), 갱신은 자식이
`git log <ref> --format=%B` **1회**로 전량 계산. 캐시 키는 ref SHA — ref 가 안 움직이면 재계산도 없다.

세 후보 중 판정:

| 후보 | 판정 |
|---|---|
| 렌더마다 라이브 계산 | **기각** — Claim 2, 13.9s |
| 기존 판정 재사용 | **불가** — `moai todo pr` 은 저장 없이 매번 라이브 계산한다(카드 본문 ②). 재사용할 저장물이 없다 |
| 캐시 + 지연 계산 | **채택** — 위 패턴, 렌더 경로 git spawn 0회 |

## Gaps — 관측하지 않은 것

- `moai statusline` 전체 렌더 지연의 현행 baseline 을 재지 않았다. 상대비교(0회 vs 80회)로
  판정이 서므로 구현 전 필수는 아니나, 회귀 근거로는 구현 시점에 재야 한다.
- 캐시 무효화 키를 ref SHA 로 잡을 때, ref 를 서브프로세스 없이 읽는 경로
  (`.git/refs/...` vs `packed-refs`)를 아직 실측하지 않았다.
- 48이라는 수는 "이력이 id 를 부른다"는 `--grep` 기준의 수다. 다른 카드의 보고 커밋이
  이 id 를 언급한 경우도 포함된다 — 이는 `Landed` 자신의 기준이므로 판정은 일치하나,
  "착지"의 의미가 그 기준보다 좁아야 하는지는 이 카드의 범위 밖이다.
- Class B 이므로 원인은 코드 판독 + 위 실측으로 세웠다. 런타임 재현(실제 렌더에서
  80회 spawn 관측)은 하지 않았다 — 렌더 경로에 아직 그 코드가 없기 때문이다.

## Residual-risk

- `%B` 전량 훑기는 이력 길이에 선형이다(현재 5,674 커밋 / 0.347s). 자식 쪽이라 렌더는 안 물지만,
  이력이 크게 자라면 자식의 예산을 재야 한다.
- 캐시가 stale 인 동안 statusline 은 옛 수를 보인다 — github 카운트와 같은 성질의 트레이드오프다.
