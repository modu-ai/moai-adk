# t476 ↔ t475 겹침 판정

> 카드: t476 (codemaps 갱신 + 진행 상황 HTML 리포트)
> 대상: t475 (codemaps stale — described-source-diff 144, 임계 40)
> 측정 트리: worktree `.claude/worktrees/t476`, 브랜치 `WT-codemaps-progress`, HEAD `25a3212a9`
> 측정: 2026-09-04

## 판정

**t476이 t475를 흡수합니다.** t475는 별도 카드로 돌 필요가 없고, t476이 착지한 뒤
`moai graph check`가 초록이면 흡수로 종결하면 됩니다.

## 증거

### 1. t475의 적색을 이 트리에서 재현

CI와 같은 방식으로 head에서 빌드해 판정기를 돌렸습니다.

```
$ go build -o ./bin/moai ./cmd/moai      # rc=0
$ ./bin/moai graph check                  # rc=1
codemaps  metric=described-source-diff value=144 threshold=40 verdict=stale
mx-index  metric=inventory-content-diff  value=0   threshold=1  verdict=absent
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
  contribution: 0 described-worthy file(s) vs first parent db177683d
                (inherited — this change contributed none of it)
```

리드가 보고한 144/40이 이 트리에서 그대로 재현됩니다. `contribution: 0`이므로 이 브랜치가
만든 드리프트가 아니라 전부 상속분입니다.

> 주의: 위 rc는 파이프 없이 따로 잰 값입니다. `./bin/moai graph check | tail`로 재면
> `tail`의 rc(0)가 잡혀 초록으로 오독됩니다.

### 2. 낡음이 스탬프가 아니라 **서술 내용**에 있다

t475 카드가 "스탬프만 갱신하면 낡은 서술에 새 앵커를 붙이는 셈"이라고 경고합니다.
그 경고가 실제로 유효한지 확인했습니다 — `modules.md`에 박힌 실측 수치 4건을 표본으로
현재 트리와 대조했습니다.

| 패키지 | modules.md 서술 | 실제 (HEAD) | 판정 |
|---|---|---|---|
| `internal/cli` | 264 non-test (2026-09-02 실측) | 273 | **드리프트 +9** |
| `internal/statusline` | 19 non-test | 20 | **드리프트 +1** |
| `internal/tui` | 19 non-test | 19 | 일치 |
| `internal/web` | 29 non-test | 29 | 일치 |

표본 4건 중 2건이 어긋납니다. 스탬프 커밋 `ad272be20` 이후 described roots
(`internal` · `cmd` · `pkg`) 아래에서 변경된 파일은 **396건**입니다
(`git diff --name-only ad272be20 HEAD -- internal cmd pkg | wc -l`).
판정기가 세는 144는 그중 described-worthy로 걸러진 부분집합입니다.

### 3. 지름길이 실재하고, 게이트가 그것을 구별하지 못한다

```
$ ./bin/moai graph --help
  stamp codemaps [--flags]  Stamp codemaps provenance (run after regenerating codemaps)
```

`moai graph stamp codemaps`는 provenance만 다시 찍습니다. 그런데 판정 지표는
`described-source-diff` — **스탬프 커밋 이후 변경분**이라, 스탬프만 찍어도 값이 0으로
떨어져 CI가 초록이 됩니다.

**즉 이 게이트는 진짜 재생성과 맨 재스탬프를 구별하지 못합니다.** t475를 단독으로 돌리면
가장 값싼 통과 경로가 곧 거짓 초록이고, 그 경로를 막는 것은 카드 본문의 주의 한 줄뿐입니다.

서술 내용의 재생성은 CLI가 아니라 `/moai:codemaps` 스킬이 수행합니다 —
`provenance.json`의 `generated_by`가 `codemaps-gen`이고, 이것이 t476 절차의 1단계입니다.

## 결론

- t476을 절차대로 수행하면(서술 재생성 → 스탬프) t475의 적색이 해소됩니다 → **t475는 no-op**.
- t475를 단독으로 먼저 돌리면 같은 표면(`.moai/project/codemaps/`)을 건드려 t476과 충돌하고,
  제대로 하면 그 작업이 곧 t476의 1단계라 중복입니다.
- 따라서 **t475 배차는 계속 보류**하고, t476 착지 후 병합 트리에서 `moai graph check`를
  다시 돌려 초록이면 흡수로 종결하는 것이 맞습니다.

## Gaps — 관측하지 않은 것

- **재생성 후 값이 실제로 임계 아래로 떨어지는 것을 아직 보지 않았습니다.** 지표가 스탬프
  기준이라 구조상 0이 되지만, 관측은 t476 수행 후에야 가능합니다.
- 144는 `25a3212a9`(이 워크트리 생성 시점의 `origin/develop` tip) 기준입니다. develop이
  전진하면 값이 움직입니다.
- `modules.md` 외 5개 문서(`overview` · `data-flow` · `dependencies` · `docs-truth` ·
  `entry-points`)의 서술 정확도는 표본 검사하지 않았습니다.
- CI 적색 3건 중 나머지 2건(토큰 예산 초과 123 · `internal/core/git` 픽스처)은 이 카드의
  표면이 아니며 건드리지 않았습니다.

## 정정 기록

착수 초기에 `internal/navigator`를 "서술됐으나 실재하지 않는 패키지"로 읽었습니다. **틀렸습니다** —
`go list`에 안 뜬 이유는 그 층에 빌드 대상 파일이 없고 하위 패키지(`astx` · `detect` · `fix` ·
`route` · `sync` · `tiers`)로만 구성돼서입니다. 부모 층위 서술은 정당합니다.
같은 이유로 `internal/skills`가 서술되지 않은 것도 결함이 아닙니다 — 그 아래에는
`workflow_split_test.go` 하나뿐이라 서술 대상 패키지가 아닙니다.
