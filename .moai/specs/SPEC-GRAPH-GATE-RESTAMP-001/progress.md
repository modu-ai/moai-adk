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

### E.2.7 블로커 — `spec.md` §B.4 / §D.1이 규칙 C의 도달면을 잘못 측정했다

**AC-4와 AC-5가 현재 SPEC 문면 그대로는 동시에 만족될 수 없다.** 배차문의 STOP 지시에 따라
임의 이탈 없이 여기서 판단을 리드에게 되돌린다.

- `spec.md` §B.4: *"모든 기존 codemaps 층 픽스처에서 본문은 미추적 상태다"* — **거짓**. 이 측정은
  `writeCodemapsProvenance`(본문 `modules.md`를 씀)만 보고, 본문을 전혀 쓰지 않는
  `writeCodemapsProvenanceBlock` 직접 호출 픽스처를 세지 않았다.
- `spec.md` §D.1: *"[규칙 C에 닿는] 그 경우는 앞선 'codemaps directory missing' 분기가 이미
  absent로 처분한다"* — **거짓**. 디렉터리는 존재하고 그 안에 `provenance.json` 하나만 있는 상태가
  가능하며, `os.Stat(dir).IsDir()` 분기는 이를 통과시킨다. 그 상태가 곧 규칙 C다.

영향 범위(측정): 정확히 기존 테스트 2건. 둘 다 codemaps 디렉터리에 `provenance.json`만 있고 본문
`*.md`가 없는 픽스처이며, 수리 전에는 `fresh`/`value=0`이었다.

| 테스트 | 관측 |
|---|---|
| `internal/graph/check_regression_lock_test.go` `TestCheckFreshness_DescribedRootsScopeFidelity` | `CheckFreshness: codemaps stamp 80bf120351cd not comparable in this checkout: no commit in the stamped history touches the codemaps body` |
| `internal/cli/graph_check_test.go` `TestGraphCheckCmd_AbsentExitsOne` | `absent layers must exit 1, got graph check: codemaps stamp 61195101480e not comparable in this checkout: …` |

두 테스트 모두 codemaps 본문을 주제로 삼지 않는다(각각 described_roots 범위 충실성, CLI absent
종료코드). 본문 없는 모양은 우발적이다.

규칙 C가 실제로 도달 가능한 유일한 상태가 **"본문이 양쪽 트리 어디에도 없음"**임을 확인했다. AC-4가
예시로 든 두 형태는 모두 규칙 C에 닿지 않는다 — 얕은 이력의 경계 커밋은 root로 취급돼
`git log -1 <S> -- <본문>`이 그 커밋 자신을 반환하고, "초기 커밋 이후 본문 무변경" 역시 root 커밋이
모든 파일을 추가로 보고하므로 같은 이유로 비지 않는다.

선택지(리드 판정 사항, 레인이 고르지 않는다):

1. **기존 테스트 2건의 픽스처에 커밋된 codemaps 본문 1개를 추가한다.** 두 테스트의 단언과 주제는
   불변이고 우발적 모양만 바뀐다. 비용: "기존 테스트를 SPEC 요구 없이 수정하지 않는다"는 계약과
   AC-5 문면에 대한 명시적 예외 승인 필요. 변경량은 두 파일 합쳐 수 줄.
2. **본문 부재를 제3의 앵커 소스 토큰으로 분기한다**(예: `no-body-present`, 앵커 = S). 기존 동작이
   보존되나 REQ-GGR-009의 토큰 집합이 넓어지므로 `manager-spec` 경유 SPEC 개정이 필요하고, 그러면
   규칙 C는 도달 불가가 되어 AC-4가 픽스처를 갖지 못한다.
3. **`spec.md` §B.4 / §D.1 / AC-4 / AC-5를 `manager-spec`이 정정**한 뒤 이 카드로 재위임.

어느 선택지든 `spec.md` §B.4의 측정 오류와 §D.1의 도달면 오류는 기록 정정이 필요하다 — 판정이
어느 쪽으로 나든 그 두 문장은 거짓인 채로 남기 때문이다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: blocked
run_blocker: "spec.md §B.4 / §D.1 measurement error — AC-4 and AC-5 cannot both hold as written (progress.md §E.2.7)"
run_commit_sha: pending-backfill
ac_pass_count: 6      # AC-1, AC-2a, AC-2b, AC-3, AC-6, AC-7
ac_fail_count: 0
ac_blocked_count: 2   # AC-4 (fixture collides with AC-5), AC-5 (2 pre-existing tests red)
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0   # go vet RC=0, gofmt clean
cross_platform_build:
  darwin_native: "go build ./... RC=0"
  windows_cross: not-run
total_run_phase_files: 3
m1_to_mN_commit_strategy: "single M1 commit; M2/M3 folded in — the work is one coherent unit and is BLOCKED pending the §E.2.7 decision"
pre_existing_red_not_attributable_to_this_card:
  - "internal/cli TestBinaryLag_DoctorCheckNameSetIsUnchanged (RED at HEAD cd28923f2, measured)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
