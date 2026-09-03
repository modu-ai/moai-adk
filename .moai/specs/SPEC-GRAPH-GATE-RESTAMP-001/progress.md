# SPEC-GRAPH-GATE-RESTAMP-001 — 진행 기록 (카드 t478)

## §E.1 Plan-phase Audit-Ready Signal

- 저작 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t478`, 브랜치 `WT-graph-gate-restamp`,
  HEAD `456665e8d` (plan-phase 저작 시점).
- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).
- SPEC ID 정규식 자체검사: `[[ "SPEC-GRAPH-GATE-RESTAMP-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]`
  → `PASS`.
- plan-phase에서 이 트리에서 직접 측정한 값(`spec.md` §B.2 / §B.3):
  - `git log -1 --format=%H ad272be20 -- .moai/project/codemaps/ ':(exclude).moai/project/codemaps/provenance.json'`
    → `2f28bc394718c2cb49a6dc96577d12c6da35b05d`
  - `git merge-base --is-ancestor 2f28bc394 ad272be20` → rc 0
  - `git diff --name-only ad272be20 -- internal cmd pkg | wc -l` → `405`
  - `git diff --name-only 2f28bc394 -- internal cmd pkg | wc -l` → `570`
  - `git diff --name-only ad272be20 -- .moai/project/codemaps/ ':(exclude)…/provenance.json'`
    → 본문 6개 파일 출력 (규칙 A 프로브 발화 — `spec.md` §D.2)
- 레인 결정 2건 반영 완료 (2026-09-04). **미해소 항목 없음** — `[NEEDS CLARIFICATION]` 표식 0개.
  1. 규칙 A `content_anchor_source` 토큰 → `working-tree-differs-from-stamp` (로직 불변; `spec.md` §D.2).
  2. 규칙 A 프로브 → `git diff` ∪ `git ls-files --others --exclude-standard` (`spec.md` §B.4, §D.1).
     근거 실측: `internal/graph/check_test.go` `writeCodemapsProvenance`가 base 커밋 이후
     `.moai/project/codemaps/modules.md`를 `os.WriteFile`로 쓰고 커밋하지 않음 → 기존 픽스처 본문은 미추적.
     정정 전이라면 규칙 A·B 모두 발화하지 않아 규칙 C(absent)로 떨어져 AC-5가 구조적으로 실패.
- 인수 기준 신설: **AC-7** (미추적 본문 픽스처 → 규칙 A로 해석, verdict는 `absent`가 아님). 기존 AC 번호
  재부여 없음 (AC-1~AC-6 그대로).
- 정정 라운드 3 (2026-09-04, manager-develop 하류 발견 → 레인 결정). `spec.md` §B.4 / §D.1의 **거짓 문장
  2건을 본문에서 정정**했다(주석 덧붙이기 아님).
  1. §B.4 거짓: "모든 기존 픽스처가 본문을 미추적으로 남긴다" → 실제로는 헬퍼가 두 겹이며 안쪽
     `writeCodemapsProvenanceBlock`은 본문을 쓰지 않는다(본문-부재 픽스처 존재). 확인함.
  2. §D.1 거짓: "본문-부재는 `codemaps directory missing` 분기가 이미 처분한다" → 그 분기는
     `os.Stat(dir).IsDir()`(`check.go:285`)이라 `provenance.json`만 든 디렉터리는 통과한다. 확인함.
  - 결정: 규칙 C를 **C1(본문 부재 → absent, 오류 nil, exit 1) / C2(앵커 미해석 → absent + 오류, exit 2)**로
    분리. C1의 선례는 `internal/graph/check_citations.go` `docs == 0` 분기(absent, 오류 없음) — 확인함.
  - 레인 실측 인용(HEAD `2649fe296`): `TestCheckFreshness_DescribedRootsScopeFidelity`와
    `TestGraphCheckCmd_AbsentExitsOne` 2건이 규칙 C의 **시스템 오류** 때문에 실패. 두 실패 모두 absent
    판정이 아니라 오류 반환을 문제 삼는다. `spec.md` §B.5에 기록.
  - **AC-4 재작성**(C1 형태 + "어떤 앵커 해석 결과도 fresh가 아니다" 불변식 + C2 도달 불가 선언·픽스처
    금지), **AC-5 개정**(허가된 픽스처 편집 정확히 1건:
    `TestCheckFreshness_DescribedRootsScopeFidelity`에 미추적 `modules.md` 부여, 단언·이름·판별식 불변).
    AC 번호 재부여 없음.
- 코드는 작성하지 않았다. `provenance.json`은 만지지 않았다.

## §E.2 Run-phase Evidence

측정 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t478`, 브랜치 `WT-graph-gate-restamp`,
run-phase 착수 시점 HEAD `cd28923f2`. 아래 모든 명령은 이 트리에서 이번 실행에 직접 돌렸다.

### E.2.1 변경 파일

| 파일 | 성격 |
|---|---|
| `internal/graph/check.go` | `LayerReport`에 보고 전용 필드 2개, 앵커 토큰 상수 2개, `resolveContentAnchor` + `hasNonBlankLine` + `codemapsBodyPathspec` 신규, `checkCodemaps` clean 경로 배선 |
| `internal/graph/check_restamp_anchor_test.go` | 신규 — AC-1/2a/2b/3/4/7 + dirty 경로 무앵커 잠금 |
| `internal/cli/graph_check.go` | `writeLayerAttribution`에 `measured from:` 1줄 추가, `--help` 본문의 codemaps 지표 설명 정정(구 설명은 이 변경으로 거짓이 됨) |

`internal/cli/graph_check.go`는 `plan.md` §D / `acceptance.md` §D의 "변경 파일 = check.go + 신규 테스트"
제약을 벗어난다. 배차문이 사람 가독 출력면 반영을 명시 지시했고(`writeLayerAttribution`의 기존 모양을
따를 것), help 본문의 구 설명이 이 변경으로 **거짓 서술**이 되기 때문이다. 은폐하지 않고 여기 기록한다.

### E.2.2 RED — 수리 전 관측 (AC-1의 필수 전제)

수리 전 코드(필드 2개만 추가하고 앵커 로직 없음)에 신규 테스트를 걸어 관측했다. 컴파일 오류가 아니라
**동작 불일치**로 떨어지게 하려고 필드 추가와 로직 배선을 분리했다.

```
$ go test ./internal/graph/ -run 'TestCheckCodemaps_' -v     # RC=1
--- FAIL: TestCheckCodemaps_BareRestampStaysStale (0.88s)
    verdict = "fresh", want "stale" — a bare re-stamp over an untouched body must not buy a green gate
    value = 0, want >= threshold 40 — the window must start at the body's last change, not at the stamp
--- FAIL: TestCheckCodemaps_UnresolvableAnchorIsAbsentPlusError (0.74s)
    an unresolvable anchor must surface a system error, got verdict "fresh" value 0
--- FAIL: TestCheckCodemaps_UncommittedRegenerationIsFresh / _CommittedRegenerationIsFresh /
          _RuleBAnchorIsAncestorOfStamp / _UntrackedBodyResolvesViaRuleA
```

전문: `.moai/reports/t478/red-before-repair.txt`. AC-1의 핵심 관측 — **맨손 재스탬프가 수리 전에는
`fresh` / `value=0`으로 통과했다.** 위조-초록이 실재함을 이 출력이 증명한다.

### E.2.3 GREEN — 수리 후

```
$ go test ./internal/graph/ -run 'TestCheckCodemaps_' -v     # RC=0
--- PASS: TestCheckCodemaps_BareRestampStaysStale (1.32s)
--- PASS: TestCheckCodemaps_UncommittedRegenerationIsFresh (1.09s)
--- PASS: TestCheckCodemaps_CommittedRegenerationIsFresh (1.29s)
--- PASS: TestCheckCodemaps_RuleBAnchorIsAncestorOfStamp (1.35s)
--- PASS: TestCheckCodemaps_UnresolvableAnchorIsAbsentPlusError (0.78s)
--- PASS: TestCheckCodemaps_UntrackedBodyResolvesViaRuleA (0.67s)
--- PASS: TestCheckCodemaps_DirtyPathCarriesNoAnchor (0.47s)
ok  	github.com/modu-ai/moai-adk/internal/graph	14.361s
```

전문: `.moai/reports/t478/green-after-repair.txt`.

### E.2.4 뮤턴트 2건 — 주장이 아니라 실행

| 뮤턴트 | 적용 | 관측 | 증거 |
|---|---|---|---|
| 규칙 A 프로브에서 `ls-files --others` 항 제거 (`acceptance.md` AC-7 지정 방향) | 실제 적용 후 실행, 되돌림 | `TestCheckCodemaps_UntrackedBodyResolvesViaRuleA` RED — `no commit in the stamped history touches the codemaps body`. **`spec.md` §B.4의 예측(규칙 C로 추락)과 정확히 일치** | `.moai/reports/t478/mutant-1-drop-lsfiles-others.txt` |
| dirty 경로에 앵커를 실어 보냄 | 실제 적용 후 실행, 되돌림 | `TestCheckCodemaps_DirtyPathCarriesNoAnchor` RED — `dirty path reported anchor ""/"working-tree-differs-from-stamp", want both empty` | `.moai/reports/t478/mutant-2-anchor-leaks-into-dirty-path.txt` |

두 번째 뮤턴트를 돌린 이유: 그 테스트는 수리 전에도 통과한다. 뮤턴트를 실행해 RED를 보이지 않으면
공허한 초록이며, 회귀 잠금이라고 부를 근거가 없다.

### E.2.5 검증 배치 (모두 이 실행, 이 트리)

| 명령 | 결과 |
|---|---|
| `go build ./...` | RC=0, 무출력 |
| `go vet ./internal/graph/... ./internal/cli/...` | RC=0, 무출력 |
| `gofmt -l internal/graph/check.go internal/graph/check_restamp_anchor_test.go internal/cli/graph_check.go` | 무출력 |
| `go test ./internal/graph/...` | **RC=1** — 신규 테스트는 전부 PASS, 기존 `TestCheckFreshness_DescribedRootsScopeFidelity` 1건 FAIL (§E.2.7 블로커) |
| `go test ./internal/cli/...` | **RC=1** — FAIL 2건: `TestGraphCheckCmd_AbsentExitsOne`(§E.2.7 블로커), `TestBinaryLag_DoctorCheckNameSetIsUnchanged`(**HEAD 기준 이미 RED — 이 카드와 무관**) |

`TestBinaryLag_…`의 사전-적색 귀속은 추정이 아니라 측정이다: `internal/graph/check.go`와
`internal/cli/graph_check.go`를 `git show HEAD:`로 되돌리고 신규 테스트 파일을 치운 트리에서 같은
테스트를 돌려 동일하게 FAIL하는 것을 확인했다
(`.moai/reports/t478/baseline-cli-binarylag-red-at-head.txt`). 같은 실행에서
`TestGraphCheckCmd_AbsentExitsOne`은 PASS했다 — 즉 그쪽은 이 변경이 깬 것이다.

### E.2.6 실트리 종단 검증 (AC-6 — 파이프 없는 종료코드 판독)

```
$ ./bin/moai graph check --json > /tmp/t478-e2e-json.txt 2>/tmp/t478-e2e-err.txt
$ echo $?
1
```

`./bin/moai`는 이 트리에서 `go build -o ./bin/moai ./cmd/moai`로 만든 것이다(`~/go/bin/moai` 아님).
종료코드는 파이프를 거치지 않고 직접 읽었다.

codemaps 층 판정: `value=146` `threshold=40` `verdict=stale`,
`content_anchor=ad272be20abff9e4f3b1b363fce3e48dac4c5132`,
`content_anchor_source=working-tree-differs-from-stamp`. 이 트리는 실제로 stale이 맞고, 앵커는
`spec.md` §B.3이 기록한 대로 규칙 A로 해석됐다. 사람 가독 출력에도
`  measured from: ad272be20 (working-tree-differs-from-stamp)` 한 줄이 실린다.
전문: `.moai/reports/t478/e2e-graph-check.json`, `…/e2e-graph-check-text.txt`.

`git status --porcelain .moai/project/codemaps/` → 무출력. `provenance.json`은 바이트 불변이며
`moai graph stamp codemaps`는 이 트리에서 한 번도 실행하지 않았다.

### E.2.7 블로커와 그 해소 — `spec.md` §B.4 / §D.1의 도달면 오측정

**블로커(레인 제기, 2026-09-04).** AC-4와 AC-5가 당시 SPEC 문면 그대로는 동시에 만족될 수 없었다.
배차문의 STOP 지시에 따라 임의 이탈 없이 판단을 리드에게 되돌렸다.

- `spec.md` §B.4: *"모든 기존 codemaps 층 픽스처에서 본문은 미추적 상태다"* — 거짓. 이 측정은
  `writeCodemapsProvenance`(본문 `modules.md`를 씀)만 보고, 본문을 전혀 쓰지 않는 안쪽 헬퍼
  `writeCodemapsProvenanceBlock` 직접 호출 픽스처를 세지 않았다.
- `spec.md` §D.1: *"[규칙 C에 닿는] 그 경우는 앞선 'codemaps directory missing' 분기가 이미 absent로
  처분한다"* — 거짓. `os.Stat(dir).IsDir()` 분기는 `provenance.json` 하나만 든 디렉터리를 통과시킨다.

영향 범위(측정): 기존 테스트 정확히 2건. 둘 다 본문 없는 픽스처이며 수리 전에는 `fresh`/`value=0`이었다.

| 테스트 | 관측(HEAD `2649fe296`) |
|---|---|
| `internal/graph/check_regression_lock_test.go` `TestCheckFreshness_DescribedRootsScopeFidelity` | `CheckFreshness: codemaps stamp … not comparable in this checkout: no commit in the stamped history touches the codemaps body` |
| `internal/cli/graph_check_test.go` `TestGraphCheckCmd_AbsentExitsOne` | `absent layers must exit 1, got graph check: codemaps stamp … not comparable …` |

**해소(리드 판정, spec.md v0.1.3).** 두 실패 중 어느 쪽도 **absent 판정이 틀렸다**고 말하지 않았다.
둘 다 **시스템 오류가 틀렸다**고 말했다. 그래서 규칙 C를 확정된 관측과 실패한 측정으로 가른다 —
1/2 종료코드 계약이 이미 싣고 있는 구분이다.

| 갈래 | 조건 | 처분 |
|---|---|---|
| C1 | 본문 집합이 비어 있음(작업 트리·S 양쪽) | `verdict=absent` + 본문 부재 reason, **오류 nil** → exit 1 |
| C2 | 본문은 존재하나 앵커 미해석(이력 절단) | `verdict=absent` + 시스템 오류 → exit 2 (불변) |

C1은 발명이 아니라 형제 층과의 정합이다 — `internal/graph/check_citations.go` `checkCitations`의
`docs == 0` 분기가 이미 absent + 오류 없음으로 처분한다. C1이 없는 상태가 오히려 두 층의 분기였다.
이 트리에서 실제로 확인됐다: 수리 후 그 픽스처의 citations 행은 `no codemaps documents to check`,
codemaps 행은 `no codemaps documents to anchor on` — 같은 사실을 같은 처분으로 말한다.

### E.2.8 C1/C2 분리의 RED → GREEN

**RED (분리 전, 규칙 C가 한 갈래이던 코드).** AC-4를 C1 형태로 재작성한 직후 관측했다.

```
$ go test ./internal/graph/ -run TestCheckCodemaps_BodyAbsentIsAbsentWithoutError -count=1 -v   # RC=1
    check_restamp_anchor_test.go:242: an absent body is a determinate observation, not a failed
    measurement — want a nil error, got: codemaps stamp 613498e70e36 not comparable in this
    checkout: no commit in the stamped history touches the codemaps body
--- FAIL: TestCheckCodemaps_BodyAbsentIsAbsentWithoutError (0.66s)
```

전문: `.moai/reports/t478/red-ac4-c1-before-split.txt`. 재작성한 AC-4가 분리 전 코드에서 통과하지
않는다는 것 — 즉 공허하지 않다는 것 — 이 이 관측의 요지다.

**GREEN (분리 후).**

```
$ go test ./internal/graph/ -run 'TestCheckCodemaps_' -count=1 -v   # RC=0
--- PASS: TestCheckCodemaps_BareRestampStaysStale
--- PASS: TestCheckCodemaps_UncommittedRegenerationIsFresh
--- PASS: TestCheckCodemaps_CommittedRegenerationIsFresh
--- PASS: TestCheckCodemaps_RuleBAnchorIsAncestorOfStamp
--- PASS: TestCheckCodemaps_BodyAbsentIsAbsentWithoutError
--- PASS: TestCheckCodemaps_UntrackedBodyResolvesViaRuleA
--- PASS: TestCheckCodemaps_DirtyPathCarriesNoAnchor
ok  	github.com/modu-ai/moai-adk/internal/graph	11.985s
```

전문: `.moai/reports/t478/green-after-c1c2-split.txt`.

**뮤턴트 3 — C1을 시스템 오류로 되돌리기(분리 자체의 판별식).** 실제로 적용해 실행하고 되돌렸다.

```
$ go test ./internal/cli/ -run TestGraphCheckCmd_AbsentExitsOne -count=1 -v   # RC=1
    graph_check_test.go:271: absent layers must exit 1, got graph check: codemaps stamp
    115259995ad1 not comparable in this checkout: no codemaps documents to anchor on
--- FAIL: TestGraphCheckCmd_AbsentExitsOne
$ go test ./internal/graph/ -run TestCheckCodemaps_BodyAbsentIsAbsentWithoutError -count=1   # RC=1
    check_restamp_anchor_test.go:242: … want a nil error, got: … no codemaps documents to anchor on
```

전문: `.moai/reports/t478/mutant-3-c1-as-system-error.txt`. 되돌린 뒤 `go build ./...` RC=0.

**C2는 픽스처를 만들지 않는다 — 선언된 미검증이다.** 얕은 경계 커밋과 루트 커밋 모두 모든 파일을
ADDED로 보고하므로, 본문이 존재하는 한 규칙 B의 `git log -1 <S> -- <본문>`은 비지 않는다. 즉 C2에
도달하는 git 상태가 없다. 도달하지 못하는 픽스처의 초록은 공허한 초록이므로(`plan.md` §G 안티패턴),
C2는 fail-closed 분기로 코드에 남기고 그 근거를 `check.go`의 `errNoBodyCommit` 주석에 적어 둔다.

### E.2.9 허가된 픽스처 편집 — 정확히 1건

- **`TestGraphCheckCmd_AbsentExitsOne` — 편집하지 않았다.** 리드의 예측대로 C1 하에서 **무수정으로
  통과**하는 것을 먼저 확인했다(`.moai/reports/t478/absent-exits-one-unmodified-pass.txt`, RC=0).
  가정하지 않고 측정한 뒤 손대지 않았다.
- **`TestCheckFreshness_DescribedRootsScopeFidelity` — 픽스처에 미추적 `modules.md` 1개 추가.**
  `writeCodemapsProvenance`가 이미 쓰는 것과 같은 모양이다. 규칙 A를 타고 S에 앵커되어 기존 단언이
  그대로 성립한다. **단언·이름·뮤턴트 판별식(하드코딩된 `DefaultDescribedRoots`)은 바꾸지 않았다.**
  편집 전 `RC=1`(`.moai/reports/t478/scope-fidelity-before-fixture-edit.txt`) → 편집 후 `RC=0`
  (`…/scope-fidelity-after-fixture-edit.txt`).

다른 어떤 기존 테스트도 수정하지 않았다.

### E.2.10 검증 배치 (재실행, 분리 후)

| 명령 | 결과 |
|---|---|
| `go build ./...` | RC=0, 무출력 |
| `go vet ./internal/graph/... ./internal/cli/...` | RC=0, 무출력 |
| `gofmt -l <변경 4파일>` | 무출력 |
| `go test ./internal/graph/... -count=1` | **RC=0** — `ok internal/graph 24.676s`, `ok internal/graph/symbol 0.580s` |
| `go test ./internal/cli/... -count=1` | RC=1 — 실패 **1건뿐**이며 `TestBinaryLag_DoctorCheckNameSetIsUnchanged`(이 카드와 무관, 아래) |

블로커였던 2건은 이제 통과한다. `TestGraphCheckCmd_AbsentExitsOne`은 `internal/cli` 실패 목록에서
사라졌고, `TestCheckFreshness_DescribedRootsScopeFidelity`는 `internal/graph` RC=0에 포함된다.

### E.2.11 종단 검증 재실행 (AC-6 — 파이프 없는 종료코드 판독)

```
$ go build -o ./bin/moai ./cmd/moai
$ ./bin/moai graph check --json > /tmp/t478-e2e2.json 2>/tmp/t478-e2e2.err
$ echo $?
1
```

codemaps 층: `value=146` `threshold=40` `verdict=stale`,
`content_anchor=ad272be20abff9e4f3b1b363fce3e48dac4c5132`,
`content_anchor_source=working-tree-differs-from-stamp`. 사람 가독 출력에도
`  measured from: ad272be20 (working-tree-differs-from-stamp)` 한 줄이 실린다.
전문: `.moai/reports/t478/e2e-graph-check.json`, `…/e2e-graph-check-text.txt`.

`git status --porcelain .moai/project/codemaps/` → 무출력. `provenance.json`은 바이트 불변이며
`moai graph stamp codemaps`는 이 트리에서 한 번도 실행하지 않았다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete
run_commit_sha: pending-backfill
ac_pass_count: 7      # AC-1, AC-2a, AC-2b, AC-3, AC-4(C1), AC-5, AC-6, AC-7 — AC-2는 a/b 두 하위로 셈
ac_fail_count: 0
ac_deliberately_unfixtured: 1   # C2 — 도달 가능한 git 상태가 없다(progress.md §E.2.8). fail-closed로 코드에 남김
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0   # go vet RC=0, gofmt clean
cross_platform_build:
  darwin_native: "go build ./... RC=0"
  windows_cross: not-run
total_run_phase_files: 4   # internal/graph/check.go, internal/graph/check_restamp_anchor_test.go,
                           # internal/graph/check_regression_lock_test.go(허가된 픽스처 편집 1건),
                           # internal/cli/graph_check.go
m1_to_mN_commit_strategy: "M1 착지 후 블로커 제기 → 리드가 spec.md v0.1.3로 §B.4/§D.1 정정 → C1/C2 분리를 후속 커밋으로 착지. 두 커밋 모두 t478을 명시"
blocker_raised_and_resolved:
  raised_at_commit: 2649fe296
  clause: "spec.md §B.4 (fixture survey) + §D.1 (rule-C reachability) — 둘 다 거짓"
  resolution: "spec.md v0.1.3 — 규칙 C를 C1(absent, 오류 nil, exit 1) / C2(absent + 오류, exit 2)로 분리"
  detail: "progress.md §E.2.7"
sanctioned_fixture_edits: 1   # TestCheckFreshness_DescribedRootsScopeFidelity — 미추적 modules.md 추가.
                              # 단언·이름·뮤턴트 판별식 불변. AbsentExitsOne은 무수정 통과(측정 확인)
pre_existing_red_not_attributable_to_this_card:
  - test: "internal/cli TestBinaryLag_DoctorCheckNameSetIsUnchanged"
    status: "RED at HEAD cd28923f2 — 이 카드 착수 전부터 붉다"
    evidence: ".moai/reports/t478/baseline-cli-binarylag-red-at-head.txt"
    method: "check.go/graph_check.go를 git show HEAD:로 되돌리고 신규 테스트 파일을 치운 트리에서 동일 실패 확인. 같은 실행에서 TestGraphCheckCmd_AbsentExitsOne은 PASS"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
