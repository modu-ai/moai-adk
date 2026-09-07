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

### M2 — 재생성 + `docs-truth.md` 손 갱신 + 편입 (REQ-CM2-003/004/005/014)

| 명령 | 출력 | HEAD |
|---|---|---|
| `/bin/ls .moai/reports/t475/pre-regen/ \| wc -l` (M2.1 선행 조건) | `6` | `52f863f36` |
| `/bin/ls .moai/project/codemaps/` | 7항목 (문서 6 + `provenance.json`) | `295fd8f11` |
| `/usr/bin/grep -n "docs-truth" .claude/skills/moai/workflows/codemaps.md` | 무출력, exit 1 → 생성기 산출물 아님 | `295fd8f11` |
| `find .claude/agents/moai -maxdepth 1 -name '*.md' \| wc -l` | `11` — §1 표 행 1-11 과 전수 일대일 대조, 양방향 잉여 0 | `295fd8f11` |
| omission 15개 `\/usr/bin/grep -rl -F "<단위>" .moai/project/codemaps/ \| wc -l` | 15/15 모두 ≥1 (첫 통과 3개가 0 → 전체 경로 표기로 직접 보정) | `295fd8f11` |
| §A.3(b) 컷오프 명령 | **7구간** (`internal/cli` 59 / `web` 13 / `statusline` 12 / `kanban` 11 / `template` 10 / `settings` 6 / `codexwiring` 6) — 저작 시점과 동일 | `295fd8f11` |
| 구간별 `diff -u` | 6구간 변경 행 있음, `internal/codexwiring` 은 **빈 출력**을 증거로 첨부 | `295fd8f11` |

재측정으로 직전 판의 두 진술이 거짓임이 드러나 재생성 본문에 정정으로 실었다 — 패키지 단위 엣지 `1638` 미재현(이 판 `345`), `session_start.go` 가 트리 최대 비테스트 파일이라는 주장(실제 상위 둘은 `templ` 생성 산물 168KB / 121KB).

### M3 — 정확성 검증 3항목 (REQ-CM2-006/007/008)

| 명령 | 출력 | HEAD |
|---|---|---|
| (a) 정본 규약 추출 → 존재 검사 | unique normalized **175**, absent **0**. 표 수출: `cited-paths-table.txt` | `e397ec00d` |
| (a) 교차 대조 `./bin/moai graph check --json` `citations` | `value: 0` `verdict: "fresh"` — 표의 absent 0 과 일치 | `e397ec00d` |
| (b) 히트-0 패키지 재실행 | **48 → 39**. omission 판정 단위 잔여 0, fold 판정 `internal/core/git` 만 잔여 | `e397ec00d` |
| (c) 명명된 추출 명령 | **10행**(0행 아님 — 공허한 통과 아님), **10 hit / 0 miss**, 각각 file:line 해석 | `e397ec00d` |

(a)가 실제 회귀 1건을 잡았다 — 내가 `dependencies.md` 에 templ `tool` 지시어의 전체 모듈 경로를 적어 팬텀 인용 `cmd/templ` 이 생겼고, 그대로 두었으면 citations 계층이 fresh → stale 로 뒤집혔을 것이다. blockquote 면제로 숨기지 않고 문장을 고쳐 수리했다.

**자기 신고 — fold 산문 되돌림.** 첫 통과의 재생성이 fold 5개 전부에 대해서도 산문을 새로 썼다(§A.3(a1) 위반). (b)에서 **후보 중 히트-0 잔여가 0개**로 나온 것이 신호였고, fold 단위에 **대한** 서술만 제거해 되돌렸다. omission 15개는 영향 없음. 증거 파일 §④-b 와 `verdict.md` 관측 ②에 기록했다.

### M4 — 재스탬프 (REQ-CM2-009)

| 명령 | 출력 | HEAD |
|---|---|---|
| `git merge-base HEAD origin/develop` | `52f863f3666c9ec754253a06b96ed1fe844f1590` | `cd03be0d3` |
| `./bin/moai graph stamp codemaps --commit 52f863f3666c9ec754253a06b96ed1fe844f1590` | `OK: stamped …/provenance.json` · `EXIT=0` | `cd03be0d3` |
| `git merge-base HEAD origin/develop` (스탬프 직후 재해석) | 동일 값 — 스탬프 전/후 `diff` 무출력 | `cd03be0d3` |
| `git merge-base --is-ancestor <provenance commit_sha> origin/develop` | `ANCESTOR=0` | `cd03be0d3` |

워크트리 격리 가드가 런타임 계산 인자와 변수 인자를 둘 다 거절하므로, 값을 **바로 앞 호출에서 명령으로 해석**해 리터럴로 넘겼다(SPEC 본문 수의 전사가 아니다). 판정 근거는 `provenance.json` 이 기록한 값이다.

### M5 — 게이트 종결 (REQ-CM2-010)

```
$ ./bin/moai graph check ; echo EXIT=$?
codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
EXIT=1
```

`--json` content anchor: `52f863f3666c9ec754253a06b96ed1fe844f1590`.

**stale 계층 0개.** 종료 코드 1 의 원인은 mx-index / edges 의 `absent` 이며, `CheckResult.Failed()`(`internal/graph/check.go:142-149`)가 `VerdictFresh` 가 아닌 **모든** verdict 를 실패로 세기 때문이다. AC-CM2-010 은 신규 워크트리의 `absent` 를 합격 저해 요인이 아니라고 명시하므로 판정면은 계층 verdict 이고 종료 코드가 아니다.

관측 리포트 `verdict.md` 3항목 수출 완료(설정 무변경 동반).

### 범위 위생 (AC-CM2-011)

```
$ git diff --name-only HEAD~3 HEAD | sed 's|/[^/]*$||' | sort -u
.moai/project/codemaps
.moai/reports/t475
.moai/reports/t475/pre-regen
.moai/specs/SPEC-CODEMAPS-REFRESH-002
$ git diff --stat HEAD~3 HEAD -- '*.go' internal pkg cmd
(무출력 — Go 프로덕션 코드 0줄)
```

## §E.3 Run-phase Audit-Ready Signal

### AC PASS / FAIL 매트릭스

| AC | 상태 | 검증 명령 | 실제 출력 |
|---|---|---|---|
| AC-CM2-001 | PASS | `./bin/moai graph check` · `cat provenance.json` · `go list … \| wc -l` | `value=64 threshold=40 stale` / 앵커 `25a3212a9`, `2026-09-03T18:18:34Z` / `136` — 저작 시점과 차이 없음, §E.2 에 기록 |
| AC-CM2-002 | PASS | `wc -l < candidates.txt` · `/usr/bin/grep -cE '^\| \`[^\`]+\` \| (fold\|omission) \|' <증거파일>` | `20` / `20` — 동일. `comm` 양방향 무출력(집합 동일). 조건 3-8 전수 충족: 근거마다 책임 명명 + 부모 산문 위치 인용, 인용 좌표 15개 `sed\|grep -qF` 로 exit 0 재확인, `commandemit` omission |
| AC-CM2-003 | PASS | `/bin/ls .moai/project/codemaps/` | 7항목(문서 6 + provenance.json). 생성기 5문서 전부 현재 트리에서 재작성, `generated_at` 갱신 |
| AC-CM2-003a | PASS | `/usr/bin/grep -n "docs-truth" .claude/skills/moai/workflows/codemaps.md` · `find .claude/agents/moai -maxdepth 1 -name '*.md' \| wc -l` | 무출력 exit 1 / `11` — §1 전수 대조 11/11, 양방향 잉여 0. 증거 §③ 이 §② 와 **분리** |
| AC-CM2-004 | PASS | `/usr/bin/grep -rl -F "<단위>" .moai/project/codemaps/` | omission 15/15 모두 ≥1. 직접 보정 3건 기록 |
| AC-CM2-005 | PASS | `diff -u pre-regen/<doc>.md <doc>.md` | 7구간 전부 행 존재. `internal/codexwiring` 은 **빈 diff 출력**을 증거로 첨부 |
| AC-CM2-006 | PASS | 정본 3요소 + `normalizeCitedPath` 나머지 규칙으로 추출 → 존재 검사 · `graph check --json` | unique 175 / absent **0**; 계층 `value: 0` `verdict: fresh` — **일치**. 회귀 1건(`cmd/templ` 팬텀) 검출·수리 |
| AC-CM2-007 | PASS | 히트-0 패키지 명령 재실행 | `48 → 39`, 잔여 39 전수에 분류 부여. omission 잔여 **0**, fold `internal/core/git` 만 기록 전용 |
| AC-CM2-008 | PASS | REQ-CM2-008 이 명명한 추출 명령 | **10행**(0행 아님), 10 hit / 0 miss, 각각 file:line |
| AC-CM2-009 | regression-guard | `git merge-base --is-ancestor <provenance commit_sha> origin/develop` | `ANCESTOR=0`. §D.1 처분에 따라 **통과로 기록하지 않는다** — 이 구간에서 merge-base 와 작업 시작 HEAD 가 같아 두 형식을 구분하지 못한다 |
| AC-CM2-010 | PASS | `./bin/moai graph check` | codemaps `value=0 verdict=fresh`, citations `fresh`, **stale 계층 0개**. `EXIT=1` 은 mx-index/edges `absent` 탓이며 AC 가 합격 저해 요인이 아니라고 명시 |
| AC-CM2-011 | regression-guard | `git status --porcelain` · `git diff --name-only HEAD~3 HEAD` | 변경 경로가 허용 3접두사에 한정. Go 프로덕션 코드 0줄. §D.1 처분에 따라 통과로 기록하지 않는다 |
| AC-CM2-012 | PASS | `.moai/reports/t475/verdict.md` 판독 | 관측 3항목 전부 존재 — ① 앵커 귀속(t476 / `2026-09-03T18:18:34Z`) + 5일 누적 포함 ② 20개 전수 fold 5 / omission 15 분류 요약 ③ 잔여 42 전수 목록(harness 12 포함). wrong-tree 항목 **없음**. 설정 무변경 동반 |

**요약: PASS 11 · regression-guard 2 · FAIL 0.**

### 불변 조건

| 불변 조건 | 상태 | 근거 |
|---|---|---|
| Go 프로덕션 코드 0줄 (REQ-CM2-012) | 성립 | `git diff --stat HEAD~3 HEAD -- '*.go' internal pkg cmd` 무출력 |
| `gate.yaml` · 임계값 설정 무변경 | 성립 | 변경 경로 목록에 부재 |
| 접힘 정책 무변경 | 성립 | M1 은 기존 정책 **아래에서 단위를 분류**만 했고 정책 문구를 건드리지 않았다 |
| SPEC-CODEMAPS-REFRESH-001 아티팩트 무변경 | 성립 | 변경 경로 목록에 부재 |
| fold 판정 단위 산문 무변경 (§A.3(a1)) | **성립(사후 복구)** | 첫 통과에서 5건 위반 → 되돌림. fold 5개 전부 `grep -rc` 히트 0 재확인 |

### 신호

```yaml
run_complete_at: 2026-09-08
run_commit_sha: c126a85a5   # M4+M5 커밋. 자기 해시를 스스로 실을 수 없어 후속 커밋에서 backfill
run_status: complete
ac_pass_count: 11
ac_regression_guard_count: 2
ac_fail_count: 0
preserve_list_post_run_count: 0
cross_platform_build:
  applicable: false
  reason: "Go production code diff is 0 lines (REQ-CM2-012); no compiled artifact changed"
new_warnings_or_lints_introduced: 0
total_run_phase_files: 24
m1_to_mN_commit_strategy: "one commit per milestone group — M1 / M2 / M3 / M4+M5 (4 commits, no force-push, no --amend)"
evidence_paths:
  - .moai/reports/t475/codemaps-accuracy-verification.md
  - .moai/reports/t475/candidates.txt
  - .moai/reports/t475/cited-paths-table.txt
  - .moai/reports/t475/pre-regen/
  - .moai/reports/t475/verdict.md
open_items:
  - "fold 단위가 여전히 히트-0 인지를 직접 묻는 AC 가 없다 — AC-CM2-007 의 전제로만 간접적으로 걸린다. 후속 카드 재료(verdict.md 관측 ②)"
  - "graph check 종료 코드는 신규 워크트리에서 absent 계층 탓에 1 로 남는다 — 판정면은 계층 verdict (verdict.md 관측 ① 부수)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
