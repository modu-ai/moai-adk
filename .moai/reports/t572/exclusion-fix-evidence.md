# t572 — MissingExclusions 본문 수리 검증 (2026-09-09)

## Claim

SPEC-OWNERSHIP-SILENCE-001 spec.md §6 의 `MissingExclusions` ERROR("'Out of Scope' section has
no items")가 본문 수리(다섯 `### Out of Scope —` 하위 절의 선두 산문을 `- ` 불릿으로 전환,
version 0.1.1 + HISTORY 행)로 제거됐다. 런타임 결함 아닌 plan-phase 콘텐츠 결함이며, body-owner
(manager-spec) 되위임 소관으로 수리됐다.

## Evidence

명령(워크트리 `.claude/worktrees/t572`, HEAD `1649bff43` + 본 spec.md 수리 워킹 사본 위에서 실행):

```
$ go run ./cmd/moai spec lint --strict     # rc=1
$ /usr/bin/grep -c 'OWNERSHIP-SILENCE' <output>
20
$ /usr/bin/grep 'MissingExclusions' <output> | /usr/bin/grep -c 'OWNERSHIP-SILENCE'
0
$ /usr/bin/grep 'OWNERSHIP-SILENCE' <output> | awk '{print $1, $2}' | sort | uniq -c
  10 WARNING CoverageIncomplete
  10 WARNING ModalityUnjudged
$ tail -3 <output>
0 error(s), 4718 warning(s)
exit status 1
rc=1
```

## Baseline-attribution

- 수리 전 실측(2026-09-08, manager-develop run 보고 + 본 레인 baseline): 본 SPEC의
  `MissingExclusions` ERROR 1건, 코퍼스 error 0건/warning 4,698건(rc=1).
- 수리 후 실측(본 파일, 2026-09-09, 이번 실행): 본 SPEC의 `MissingExclusions` **0건**, 코퍼스
  error 0건/warning 4,718건(rc=1 불변 — 증가분 +20은 본 SPEC의 Coverage 10 + Modality 10
  WARNING).

## Gaps

- 잔존 20 WARNING(CoverageIncomplete 10 + ModalityUnjudged 10)은 plan-phase 콘텐츠 부채로
  의도 보류했다 — 코퍼스 전체가 동일 부류(3,599+492건)의 소음 축이며 REQ 서술 재구성은 본 수리
  스코프 밖이다(관례 강제 후속 카드에서 코퍼스 스윕과 함께 처리 후보).

## Residual-risk

- 코퍼스 rc=1 은 상속 상태다(카드 착지 전부터 동일) — 본 수리는 error 축을 0으로 유지한다.
