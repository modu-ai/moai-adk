# progress.md — SPEC-CODEMAPS-REFRESH-002

카드: t475 · 워크트리: `.claude/worktrees/t475` · 브랜치: `WT-codemaps-stale`

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-08
tier: M
artifacts: [spec.md, plan.md, acceptance.md]
baseline_tree_sha: 52f863f36
baseline_note: "카드 텍스트의 described-source-diff 144는 낡음 — 본 워크트리 실측 64 (spec.md §A.1)"
```

## §E.2 Run-phase Evidence

측정 트리: 워크트리 `.claude/worktrees/t475`, 브랜치 `WT-codemaps-stale`, HEAD `52f863f36`.
증거 파일: `.moai/reports/t475/codemaps-accuracy-verification.md`(7섹션) · `candidates.txt` · `pre-regen/` · `verdict.md`.

### M0 — 기준선 재측정 (REQ-CM2-001 / AC-CM2-001)

| 명령 | 출력 | HEAD |
|---|---|---|
| `./bin/moai graph check ; echo EXIT=$?` | `codemaps metric=described-source-diff value=64 threshold=40 verdict=stale` / `citations … value=0 verdict=fresh` / mx-index·edges `verdict=absent` / `measured from: 25a3212a9` / `EXIT=1` | `52f863f36` |
| `cat .moai/project/codemaps/provenance.json` | `commit_sha 25a3212a93b4c811cbb22e3c0b34d43571fa65b4` · `tree_root …/worktrees/t476` · `generated_at 2026-09-03T18:18:34Z` · `described_roots [internal cmd pkg]` | `52f863f36` |
| `go list ./internal/... ./cmd/... ./pkg/... \| wc -l` | `136` | `52f863f36` |
| `git merge-base HEAD origin/develop` | `52f863f3666c9ec754253a06b96ed1fe844f1590` | `52f863f36` |
| `git rev-parse origin/develop` | `9dddac8828dac771d6780a9acde85844c91a5f42` | `52f863f36` |

**spec.md §A 저작 시점 값과의 차이: 없다** (64 / 40 / stale, 136, 앵커 `25a3212a9` 전부 동일). 재측정이 기준선을 대체하며 값이 같으므로 §A 를 그대로 쓴다.

### M1 — 후보 산출 + 전수 판별 (REQ-CM2-002 / AC-CM2-002)

| 명령 | 출력 | HEAD |
|---|---|---|
| `wc -l /tmp/Zero.txt /tmp/A.txt /tmp/B.txt` | `48 / 5 / 14` (+ C층 `internal/chain` 1 → 후보 **20**) | `52f863f36` |
| `wc -l < .moai/reports/t475/candidates.txt` | `20` | `52f863f36` |
| `/usr/bin/grep -cE '^\| \`[^\`]+\` \| (fold\|omission) \|' .moai/reports/t475/codemaps-accuracy-verification.md` | `20` | `52f863f36` |
| `comm -23` / `comm -13` (candidates vs 판정 행 단위 집합) | 양방향 무출력 — 집합 동일 | `52f863f36` |

분류: **fold 5 / omission 15.** 인용 좌표 15개 전부 `sed -n '<n>p' … \| grep -qF '<인용문>'` 로 exit 0 재확인(증거 파일 §①-d).

**M2.0 사본은 M1 단계에서 선행 수행했다** — 판정 인용 좌표를 재생성 후에도 해석 가능한 파일에 고정하기 위해서다. `/bin/ls .moai/reports/t475/pre-regen/ | wc -l` → `6`. (셸 프로필이 `ls` 를 `ls -la` 로 alias 하므로 aliased `ls | wc -l` 은 `9` 를 낸다 — 실측은 unaliased `/bin/ls` 로 했다.)

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
