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

---

# 추가 — v0.5.0 조항 개정 (운영자 판정 B)

측정 트리: 같은 워크트리, HEAD `dc03e0496` 시점. 아래 명령은 모두 이 트리에서 이 실행으로 돌렸다.

## Claim

4. run-phase 이후 develop 흡수가 초록을 빨강으로 바꿨고, 이는 카드 결함이 아니라 AC-ACD-006 설계 비용의 발현이다.
5. 운영자 판정 B(실패 표면 축소)를 문면과 검증자에 반영했고, 좁힘이 검사 종료와 구별된다는 것을 관측했다.

## Evidence

### 주장 4 — 병합이 만든 레드와 그 비용
```
$ git merge --no-edit origin/develop     # rc=0, 충돌 0건
$ go test ./internal/spec/... -count=1
--- FAIL: TestACCounterFullCorpusMatchesBaseline
    ac_count_clause_test.go:430: .moai/specs/SPEC-TODO-DESTRUCTIVE-GUARD-001/acceptance.md:
        counts 16 but is absent from the snapshot
```
비용은 얼리지 않고 재유도한다:
```
$ for d in 7 3 1; do git log --since=$d.days --diff-filter=A --name-only --format= \
    origin/develop -- '.moai/specs/*/acceptance.md' | grep -c acceptance.md; done
```
`origin/develop` `947f5cffb`(2026-08-28) 관측: 7일 60 / 3일 29 / 24시간 6. 한 시간 전 같은 명령은 59 / 28 / 5 였다 — **값은 브랜치와 함께 움직이며, 판정이 딛고 선 것은 자릿수(하루 여러 건)이지 특정 수가 아니다.**

### 주장 5 — 좁힘의 관측 (오케스트레이터 독립 뮤테이션 2건)

에이전트가 3방향을 관측했고, 그와 별도로 내가 두 방향을 직접 심어 재현했다.

**이음매(새 파일이 정지) — GREEN + 보고**
```
$ mkdir -p .moai/specs/SPEC-ORCH-SEAM-PROBE-001    # AC-PRB-001 을 표시/미표시로 함께 담음
$ go test ./internal/spec/ -run TestACCounterFullCorpusMatchesBaseline -count=1 -v
    ac_count_clause_test.go:483:   absent-from-snapshot
        .moai/specs/SPEC-ORCH-SEAM-PROBE-001/acceptance.md: HALT AC-PRB-001
--- PASS
```
**사라짐(스냅샷에만 있는 파일) — RED** — 에이전트가 gap 으로 남긴 분기다:
```
$ printf '%s\n' '.moai/specs/SPEC-ORCH-GONE-PROBE-001/acceptance.md  COUNT 5 …' \
    >> .moai/reports/t338/ac-count-baseline.txt
$ go test ./internal/spec/ -run TestACCounterFullCorpusMatchesBaseline -count=1
    ac_count_clause_test.go:471: .moai/specs/SPEC-ORCH-GONE-PROBE-001/acceptance.md:
        present in the snapshot but no longer matched by the corpus glob
FAIL
```
둘 다 원복했다(프로브 디렉터리 삭제 확인 / 스냅샷 `cmp` 일치).

**통과가 공허하지 않은 근거**: 스냅샷 부재 파일이 0건이 아니라 **2건**이다(`SPEC-BACKLOG-LOCK-BUDGET-001` COUNT 6, `SPEC-TODO-DESTRUCTIVE-GUARD-001` COUNT 16). 좁혀진 분기가 실제로 두 번 실행되고 두 번 보고했다. **스냅샷을 재생성하지 않은 것은 의도적이다** — 재생성하면 그 2건이 흡수돼 이 검증이 공허해진다.

**전이 테이블**: 10칸, 신규 3칸 포함, 칸마다 `wantErr` 와 `wantReport` 를 함께 단언한다. 선택자 없는 전수 루프라 0개 선택으로 초록을 내는 형태가 아니다. 보고 문자열을 바이트 단위로 비교하므로 **보고 없이 좁히기만 한 구현은 여기서 죽는다.**

```
$ go test ./internal/spec/... -count=1
ok  github.com/modu-ai/moai-adk/internal/spec  29.633s
$ for f in spec plan acceptance design research progress; do \
    python3 .moai/reports/t338/iter2-scratch/counter.py <artifact> adj; done
41 / 11 / 24 / 2 / 10 / 48 — 전부 rc=0, ambiguous=0
```
`progress.md` 만 47 → 48 로 움직였다. §E.2.9 가 뮤테이션 출력을 그대로 인용하며 `AC-MUT-002` 를 처음 담기 때문이다 — **이 작업이 만든 이동이지 회귀가 아니다.**

## 종결된 카드의 산출물을 편집했다 (명시)

M5 정규화가 `.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/acceptance.md` 를 4행 편집했다 — `AC-GFC-011`·`AC-GFC-012` 뒤에 `[RETIRED]` 토큰 부착뿐이고 산문은 바뀌지 않았다. 이 파일은 **카드 t322 의 산출물이고 t322 는 3단계 착지 후 종결됐다.** M5 의 종단 이전 정규화 범위 안이지만, 종결된 카드의 파일을 나중 카드가 편집한 사실 자체를 여기 남긴다 — 그 SPEC 을 다시 여는 사람이 편집의 출처를 `git blame` 없이 알 수 있어야 한다.

## 이 개정의 Gaps

- **v0.5.0 문면을 본 감사관이 없다.** 기록된 `PASS 0.91`(iter-3)은 v0.4.1 에 대한 판정이고, 재감사는 운영자가 불필요로 판정했다. `progress.md` §E.1 과 HISTORY 에도 같은 말이 적혀 있다.
- **CI 판정이 없다.** 브랜치 미푸시. 전체 스위트는 지시대로 로컬에서 돌리지 않았다.
- **cross-platform 미측정.** 변경은 테스트 전용 Go 이고 플랫폼 분기가 없지만 재지 않았다.
- **부재 2건의 계수(6·16)가 옳은지는 손으로 유도하지 않았다.** 개정이 주장하는 것은 그 값의 정확성이 아니라 그것이 회귀가 아닌 새 관측이라는 것뿐이다.
- **`counter.py` 내부는 신뢰에 의존한다.** 계수기 자체에 결함이 있으면 여섯 산출물의 `ambiguous=0` 이 모두 공허해진다.

---

# 최종 — sync 착지 + 병합 트리 재측정

측정 트리: 같은 워크트리, HEAD `7ac5ce369`(`origin/develop` `48d8ef4be` 흡수 완료, `0 15`).

## Claim

6. sync 단계가 닫혔고, 규약의 첫 실사용에서 판별자가 실제로 수를 바꿨다.
7. 개정 전이라면 레드를 냈을 병합이 지금은 통과하며, 새 파일들이 보고된다 — 판정 B 가 실제로 작동한다.

## Evidence

### 주장 6 — 첫 실사용
CHANGELOG 항목을 쓰며 manager-docs 가 자기 B12 자가검사를 새 규약으로 돌렸다. 같은 파일, 같은 트리, 두 형태:
```
$ grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' <this SPEC's acceptance.md> | sort -u | wc -l
27
$ python3 .moai/reports/t338/iter2-scratch/counter.py <same file> adj
COUNT 24   (live=24 excluded=3 ambiguous=0)
```
**27 → 24.** 차이 3은 구 스윕이 볼 수 없던 폐기 기준이다. CHANGELOG 에는 24 가 적혔다. 이것이 이 카드가 존재하는 이유의 첫 실증이다 — 카드 원문의 사례(스윕 8 / 실제 7)와 같은 형태이고, 이번엔 규약이 그것을 잡았다.

상태 전이(3단계 close): `spec.md` · `plan.md` · `acceptance.md` · `progress.md` → `completed`. `design.md` · `research.md` 는 `draft` 로 남겼다 — 전이 규약이 4종만 명시하고 Tier L 6종을 상정하지 않는다. **임의로 넓히지 않았고 카드 t357 소관으로 넘겼다.**

### 주장 7 — 병합이 판정을 시험한 자리
`origin/develop` 26커밋을 흡수했다(CHANGELOG 충돌 1건 — 다른 카드가 같은 자리에 항목을 넣었다. 양쪽 항목을 모두 보존해 해소했고 마커 잔여 0).

흡수는 새 `acceptance.md` 2건을 들여왔다. **개정 전 검증자였다면 이것이 정확히 레드를 냈을 형태다.**
```
$ go test ./internal/spec/ -run TestACCounterFullCorpusMatchesBaseline -count=1 -v
    ac_count_clause_test.go:481: AC corpus: 4 file(s) matched by the glob but absent
        from the snapshot - reported, not failed (spec.md 3.5 rule 4)
    ac_count_clause_test.go:483:   absent-from-snapshot …/SPEC-BACKLOG-LOCK-BUDGET-001/…: COUNT 6
    ac_count_clause_test.go:483:   absent-from-snapshot …/SPEC-GUARD-LIVENESS-001/…: COUNT 13
    ac_count_clause_test.go:483:   absent-from-snapshot …/SPEC-GUARD-STATE-MODEL-001/…: COUNT 17
    ac_count_clause_test.go:483:   absent-from-snapshot …/SPEC-TODO-DESTRUCTIVE-GUARD-001/…: COUNT 16
--- PASS
$ go test ./internal/spec/... -count=1
ok  github.com/modu-ai/moai-adk/internal/spec  28.337s
```
부재 2건 → 4건으로 늘고 넷 다 이름과 관측으로 보고됐다. 통과했으되 조용하지 않다.

**계수기의 독립 교차 확인 1건**: `SPEC-GUARD-LIVENESS-001` 을 계수기가 **13** 으로 셌고, 그 카드가 자기 CHANGELOG 에 적은 값도 `13 acceptance criteria (AC-GDL-001..013)` 다. 다른 사람이 손으로 센 값과 기계 값이 일치한다 — 계수기 정확성의 외부 근거이며, `counter.py` 를 신뢰에만 의존한다는 잔여 위험을 부분적으로 줄인다(한 표본이므로 없애지는 못한다).

병합 트리 최종:
```
6종 계수기 41 / 11 / 24 / 2 / 10 / 48 — 전부 rc=0, ambiguous=0
corpus 전수 files=610  halting=0  files-with-excluded=5
manager-docs.md 로컬/미러 diff 무출력
추적 파일 미커밋 변경 0
```

## 이 단계의 Gaps

- **CI 판정이 여전히 없다.** 브랜치 미푸시. 병합 트리의 초록은 로컬 `internal/spec` 한정이고, 전체 스위트는 지시대로 로컬에서 돌리지 않았다.
- **부재 4건의 계수(6·13·16·17)가 옳은지 전수 손유도하지 않았다.** 13 한 건만 외부 값과 대조했다. 개정이 주장하는 것은 그 값들의 정확성이 아니라 그것들이 회귀가 아닌 새 관측이라는 것뿐이다.
- **CHANGELOG 충돌 해소를 렌더로 확인하지 않았다.** 마커 잔여 0과 두 항목 존재는 grep 으로 확인했으나 마크다운 렌더는 보지 않았다.
- **cross-platform · golangci-lint 미측정.** awk 계수기는 BSD awk(macOS)에서만 돌렸다.
- **v0.5.0 문면을 본 감사관이 없다.** `PASS 0.91`(iter-3)은 v0.4.1 판정이다.
