# t338 — 레인 판정 기록 (중간, 통합 판정 대기)

카드: t338 · SPEC-AC-COUNT-DISCRIMINATOR-001 · 레인 lane-7
측정 트리: `.claude/worktrees/t338`, 브랜치 `WT-ac-count-sweep`, HEAD `e5b509f3a`
상태: **run 완료 · 통합 전 blocker 1건으로 리드 판정 대기**

## Claim

1. 배차 전제("자체 커밋 0 = 미착수")는 거짓이다 — 완결된 plan-phase 산출물이 미추적으로 존재했다.
2. run-phase 는 AC-ACD-001~008 전부 PASS 로 완료했다.
3. develop 을 흡수하면 baseline 테스트가 실패한다. 이는 카드 결함이 아니라 AC-ACD-006 설계 비용의 발현이다.

## Evidence

### 주장 1 — 배차 전제 반증
```
$ git status --short          # 워크트리, 착수 전
?? .moai/reports/t338/
?? .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/
$ kill -0 67129               # .git/worktrees/t338/locked 가 지목한 소유자
kill 67129 failed: no such process       # rc=1
$ ls -la .git/worktrees/t338/index.lock
-rw-r--r-- 0 bytes  Aug 28 08:44          # 스테일, git 쓰기를 전부 차단
```
회수 커밋 `0be2e8062` — 43파일. plan-audit 3회 실재(0.67 FAIL → 0.72 FAIL → 0.91 PASS, Tier L 임계 0.85).

### 주장 2 — run-phase (오케스트레이터 독립 재현)
```
$ for f in acceptance research spec plan design progress; do python3 \
    .moai/reports/t338/iter2-scratch/counter.py \
    .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001/$f.md adj; done
COUNT 24   (live=24 excluded=3 ambiguous=0)
COUNT 10   (live=10 excluded=0 ambiguous=0)
COUNT 41   (live=41 excluded=1 ambiguous=0)
COUNT 11   (live=11 excluded=1 ambiguous=0)
COUNT 2    (live=2  excluded=0 ambiguous=0)
COUNT 47   (live=47 excluded=0 ambiguous=0)

$ bash .moai/reports/t338/iter2-scratch/corpus.sh | tail -1
files=606  halting=0  files-with-excluded=5

$ go test ./internal/spec/... -count=1
ok  github.com/modu-ai/moai-adk/internal/spec  31.730s

$ diff .claude/agents/moai/manager-docs.md \
       internal/template/templates/.claude/agents/moai/manager-docs.md
                                          # 무출력, rc=0
$ grep -c 'MOAI-AC-COUNTER-BEGIN' .claude/agents/moai/manager-docs.md   # 1
$ grep -c 'MOAI-AC-COUNTER-END'   .claude/agents/moai/manager-docs.md   # 1
$ grep -coE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' \
    internal/template/templates/.claude/agents/moai/manager-docs.md     # 0
```
`progress.md` 만 15 → 47 로 이동했다. §E.2 가 판정 기록으로 식별자를 이름으로 담기 때문이며 이 실행이 만든 이동이다. 여섯 건 모두 `ambiguous=0` — 정지 회귀 없음.

### 주장 3 — 병합 후 레드
```
$ git merge --no-edit origin/develop      # rc=0, 충돌 0건
$ go test ./internal/spec/... -count=1
--- FAIL: TestACCounterFullCorpusMatchesBaseline (3.31s)
    ac_count_clause_test.go:430: .moai/specs/SPEC-TODO-DESTRUCTIVE-GUARD-001/acceptance.md:
        counts 16 but is absent from the snapshot
FAIL  github.com/modu-ai/moai-adk/internal/spec  29.868s
```
실패 총계 1건, 원인 1개. 비용 실측 — develop 에 새로 추가된 `acceptance.md`:
```
$ git log --since=7.days --diff-filter=A --name-only --format= origin/develop \
    -- '.moai/specs/*/acceptance.md' | grep -c acceptance.md      # 59
$ ... --since=3.days ...                                          # 28
$ ... --since=1.days ...                                          #  5
```

## Baseline-attribution

모든 수치는 이 트리에서 이 실행으로 측정했다. plan-phase 가 인용한 corpus 602 는 채택하지 않고 재유도했다 — develop 흡수 전 606, 흡수 후 607. `git rev-list --count --left-right origin/develop...HEAD` 는 판정마다 fetch 와 같은 호출에서 재측정했다(origin/develop 이 이 세션 동안 `d566ecc75 → 77b2bcae6` 로 이동했다).

## Gaps — 관측하지 않은 것

- 전체 스위트 미실행. 영향 패키지만 돌렸다(`internal/spec`, `internal/template`). 브랜치가 미푸시라 **CI 판정이 존재하지 않는다**.
- awk 계수기의 GNU awk / busybox awk / Windows 동작 미측정. BSD awk(macOS)에서만 확인했고 `GOOS` 별 vet 은 컴파일만 증명한다.
- 실제 HALT 파일이 corpus 에 0건이라 스냅샷 파서의 HALT 분기는 합성 입력으로만 실행됐다.
- `golangci-lint` 미실행.
- 인용 축 124건 · 별칭 축 85건은 전수 판정하지 않았다. 상한임을 명시했고 결함으로 단언하지 않았다.

## Residual-risk

- **스냅샷은 고무도장이 될 수 있다.** M5 가 재생성을 정상 절차로 두므로 나쁜 이유로 움직인 수도 재생성 한 번으로 통과한다. 상태별 집계는 가시화 장치이지 방지 장치가 아니다.
- **다른 카드 소유 파일 편집.** M5 정규화가 `SPEC-GRAPH-FRESHNESS-CADENCE-001/acceptance.md`(t322 소유, in-progress)에 `[RETIRED]` 토큰을 4행 부착했다. 산문 무변경이나 통합 순서에 걸린다.
- **`spec.md`·`research.md` 의 청결을 지키는 AC 가 없다.** AC-ACD-006 의 corpus 는 `acceptance.md` 한정이다. 이번 run 에서 `§E.2` 초고가 실제로 한 번 정지했고 §3.4 적용으로 해소했다 — 재발 가능한 자리다.
- **`design.md`·`research.md` 가 `draft` 로 남았다.** 상태 전이 규약이 4종만 명시한다. `moai spec audit` 은 경고하지 않는다. Tier L 6종 전이 규약 공백이며 이 카드 범위 밖이다.
