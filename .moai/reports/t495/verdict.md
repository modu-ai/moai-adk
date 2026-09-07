# t495 — codex 라이브 테스트 CI 상시 skip 해소

- card: t495 (class B — plan 생략, 원인 확정됨)
- lane: lane-2
- worktree: `.claude/worktrees/t495`, branch `WT-codex-ci-skip`
- base: `origin/develop` = `ace1c5440` (카드가 실측한 것과 동일 tip)

---

## Claim

1. 카드가 주장한 사실 4건은 전부 재현된다 — CI 는 codex 를 설치하지도, 라이브 스위치를
   켜지도 않으며, 그 결과 라이브 테스트 9본이 CI 형태 환경에서 전부 skip 되고 패키지는
   `ok` 를 보고한다.
2. `TestCodexSpecFiles_ExecPrimitivesCodexOnly` 는 정규식이 낡으면 **아무것도 검사하지 않고
   PASS** 한다(공허한 초록). 하한 단언을 추가해 같은 조건이 RED 가 되게 했다.
3. "이 축은 CI 에서 관측되지 않는다"를 산문이 아니라 **기계 검사되는 선언**으로 기록했다.
   선언이 코드·CI 현실과 어긋나면 세 가드 중 하나가 RED 가 된다.
4. **CI 에 codex 설치 레그를 넣지 않았다.** 이것은 누락이 아니라 판단이며, 근거는 아래
   Residual-risk 에 적었다 — 운영자 결정 사항으로 상신한다.

## Evidence

### E1 — 카드 전제 재현 (base `ace1c5440`)

```
$ grep -rn 'codex' .github/workflows/          → (무출력)
$ grep -rnE 'MOAI_CODEX_LIVE_PROBE|MOAI_AUDIT_PIN_LIVE' .github/workflows/  → (무출력)
```

CI 형태 환경(PATH 에서 codex 제거) 재현:

```
$ go test -c -o /tmp/t495-cli.test ./internal/cli
$ PATH=/tmp/t495-nopath:/usr/bin:/bin /tmp/t495-cli.test \
    -test.run 'TestCodexLive|TestHandleCodexReviewGate_Live|TestAuditPinLive' -test.v
--- SKIP: TestAuditPinLive_CodexPinConfirmation (0.00s)
--- SKIP: TestAuditPinLive_GLMDifferential (0.00s)
--- SKIP: TestCodexLive_ExplicitReadOnlyApprovalStall (0.00s)
--- SKIP: TestCodexLive_OmittedSandboxPolicyBaseline (0.00s)
--- SKIP: TestCodexLive_ReviewStartBaseBranchIsNotRejected (0.00s)
--- SKIP: TestCodexLive_ReviewStartEmitsTurnStarted (0.00s)
--- SKIP: TestCodexLive_SandboxPolicyStickiness (0.00s)
--- SKIP: TestCodexLive_ThreadReuseAndTurnInterrupt (0.00s)
--- SKIP: TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey (0.00s)
```

**9본 전부 SKIP.** 카드는 "12본"이라 적었으나 실측은 9본이다 — 이 판정서는 실측 9를 쓴다.
(`codex_protocol_liveness_test.go` 6본은 fake 기반이라 라이브 축이 아니다.)

이 머신에서는 codex 가 PATH 에 있어 3본이 실제로 돌아 PASS 한다
(`.moai/reports/t495/baseline-run.txt`). **로컬 초록은 CI 초록의 증거가 아니다** —
이 카드의 본체가 정확히 그 격차다.

### E2 — 뮤턴트 M1: 공허한 초록 실증 → 수리 → RED 성립

뮤턴트: `codex_launcher_guards_test.go:151` 의 스캔 정규식을 존재하지 않는 심볼로 교체
(`exec.Command|exec.CommandContext|os.StartProcess` → `exec.CommandZZ|os.StartProcessZZ`).
= "정규식이 낡았다"의 시뮬레이션.

| 상태 | 결과 | 증거 |
|---|---|---|
| 수리 **전** + 뮤턴트 | `--- PASS` ← **공허한 초록** | `.moai/reports/t495/m1-before-fix.txt` |
| 수리 **후** + 정상 정규식 | `ok` | 본문 아래 |
| 수리 **후** + 같은 뮤턴트 | `--- FAIL` | `.moai/reports/t495/m1-after-fix.txt` |

수리 후 뮤턴트 실패 메시지 (verbatim):

```
codex_launcher_guards_test.go:173: codex_launcher.go: zero process-start primitives found —
the scan regex "\\b(?:exec\\.CommandZZ|os\\.StartProcessZZ)\\s*\\(([^,)]*)" matched nothing,
so this guard asserted nothing. Either the primitive moved out of this file
(update codexSpecFiles) or the regex is stale (update it).
--- FAIL: TestCodexSpecFiles_ExecPrimitivesCodexOnly (0.00s)
```

하한값의 근거(실측): `codex_launcher.go` 의 process-start 프리미티브 = **1건**,
`codex_readiness.go` = **0건**(기존 단언이 이미 0을 요구). 따라서 하한은
"`codex_launcher.go` 는 ≥1" 이며, 관측된 형태 그대로다.

### E3 — 선언 가드 3본, 뮤턴트 4종으로 비공허성 확인

정상 상태:

```
--- PASS: TestCodexLiveAxis_DeclarationCoversEveryLiveGatedFile (0.09s)
--- PASS: TestCodexLiveAxis_DeclaredSwitchesExistInSource (0.00s)
    codex_live_axis_declaration_test.go:249: codex live axis unobserved in CI:
    18 workflow files scanned, none wires any of
    [MOAI_CODEX_LIVE_PROBE MOAI_CODEX_LIVE_BIN MOAI_SKIP_LIVE_CODEX MOAI_SKIP_LIVE_CODEX MOAI_AUDIT_PIN_LIVE]
--- PASS: TestCodexLiveAxis_NotObservedInCI (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.942s
```

| 뮤턴트 | 조작 | 관측 |
|---|---|---|
| M2 | 레지스트리 파일명 `audit_pin_live_test.go` → `..._ZZ_test.go` | RED, **양방향** 검출: "live-gated but undeclared" + "declares … but it is not live-gated" |
| M3-a | 선언 env `MOAI_CODEX_LIVE_PROBE` → `..._RENAMED` | RED "declared switch … does not occur in the file — the declaration drifted from the code" |
| M3-b | 선언 테스트 수 5 → 4 | RED "declares 4 live tests, source has 5" |
| M4 | `.github/workflows/ci.yml` 에 `MOAI_AUDIT_PIN_LIVE: "1"` 추가 | RED ".github/workflows/ci.yml wires MOAI_AUDIT_PIN_LIVE — the codex live axis IS observed in CI now, so the declaration … is stale" |

가드 자체에도 하한이 있다: 레지스트리 항목 수 `< 4` → `t.Fatalf`, 탐지된 라이브 파일 0건 →
`t.Fatalf`, 스캔한 워크플로 0건 → `t.Fatalf`, 선언 env 0개 → `t.Fatal`.
빈 집합끼리의 비교가 통과로 위장하는 경로를 전부 막았다.

**M4 실행 중 자기 결함 1건 발견·수리**: `t.Logf("… none wires …")` 가 무조건 실행돼
**실패한 실행에서도 "아무것도 배선 안 됐다"고 출력**했다. `wired == 0` 일 때만 찍도록 고쳤고,
M4 를 재실행해 모순 로그가 사라졌음을 확인했다.

### E4 — 정적 검사

```
$ gofmt -l internal/cli/    → (무출력)
$ go vet ./internal/cli/    → rc=0
$ make help | grep codex-live
test-codex-live      Observe the codex live axis (opt-in; needs a codex binary and
                     spends real codex/z.ai quota — CI never runs this; …)
```

### E5 — 패키지 회귀

```
$ go test ./internal/cli/... -count=1 -timeout 900s   → rc=0
ok  	github.com/modu-ai/moai-adk/internal/cli	387.341s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	3.849s
ok  	github.com/modu-ai/moai-adk/internal/cli/harness	9.084s
… (18개 하위 패키지 전부 ok) …
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	8.974s

$ grep -c '^--- FAIL' .moai/reports/t495/pkg-full.txt   → 0
```

전문: `.moai/reports/t495/pkg-full.txt`. 상위 패키지가 387초인 것은 이 머신에
codex 가 설치돼 있어 라이브 테스트 3본이 실제로 돌기 때문이다 — CI 에서는 그 3본이
skip 되므로 이 소요 시간은 CI 를 대표하지 않는다.

### E6 — 병합 트리 재측정 (통합 창, lane-2)

창 시점 재판독: `origin/develop` = `ace1c5440` (변동 없음), **로컬** `develop` = `a4aecc4a2`
(미푸시 14). 리드가 관측한 tip 은 후자이므로 흡수 대상은 로컬 develop 이다.

```
$ git merge --no-ff develop     → 충돌 0, merge commit c2b092001 (tree 1e69c3951)
   18 files changed, 2284 insertions(+), 14 deletions(-)
   — 전부 문서·SPEC·리포트(t491 / t494 / SPEC-CC-GD124-001 / 룰 2본).
     이 카드의 반경(internal/cli Go + Makefile)과 겹치는 파일 0건.
```

병합 트리에서:

```
$ gofmt -l internal/cli/                                → (무출력)
$ go vet ./internal/cli/                                → rc=0
$ go test ./internal/cli/... -count=1 -timeout 900s     → rc=0
ok  	github.com/modu-ai/moai-adk/internal/cli	446.156s
… 17 packages ok …
$ grep -c '^--- FAIL' .moai/reports/t495/pkg-full-merged.txt   → 0
```

전문: `.moai/reports/t495/pkg-full-merged.txt`.

**베이스 실행과의 대조** — 패키지 집합이 동일함을 값이 아니라 집합 비교로 확인했다:

```
$ diff <(grep '^ok' pkg-full.txt | awk '{print $2}') \
       <(grep '^ok' pkg-full-merged.txt | awk '{print $2}')   → (차이 없음)
$ grep -c '^ok' 양쪽                                            → 17 / 17
```

**정정**: 이 판정서의 초기 E5 절과 리드 보고에 "18 하위 패키지"라 적었으나 실측은
**17**이다. 두 실행 모두 17이며 집합도 동일하다. 원 서술을 지우지 않고 여기 정정을
붙인다 — 결론(rc=0, FAIL 0)은 영향받지 않지만 인용한 수치는 틀렸었다.

## Baseline-attribution

모든 수치는 이 실행에서, 이 트리(`.claude/worktrees/t495`, base `ace1c5440`)에 대해
직접 측정했다. 카드가 제시한 리드 실측치는 재현 대상이었지 근거로 재사용하지 않았다.
"9본"은 카드의 "12본"을 그대로 옮긴 값이 아니라 위 E1 명령의 출력을 센 값이다.

## Gaps — 관측하지 않은 것

- **CI 에서의 실제 동작은 관측하지 않았다.** E1 은 `PATH` 에서 codex 를 제거한 로컬
  재현이지 GitHub Actions 러너에서의 실행이 아니다. 판정은 develop push 가 일으키는
  CI 실행에 맡긴다(레인은 CI 를 직접 요청하지 않는다).
- **라이브 축 자체는 여전히 관측되지 않는다.** 이 카드는 skip 을 없애지 않았다 —
  침묵을 선언으로 바꿨을 뿐이다. 9본이 검사하는 동작(스레드 재사용, 샌드박스 정책 고착,
  AC-CRT-010, AC-AMP-006/007)은 지금도 CI 근거가 없다.
- **`make test-codex-live` 를 실제로 끝까지 돌리지 않았다.** 실 quota 를 쓰는 명령이라
  운영자 승인 없이 실행하지 않았다. 타깃이 `make help` 에 노출되는 것까지만 확인했다.
- `internal/cli` 외 패키지는 돌리지 않았다(변경 반경이 그 패키지 + Makefile 뿐).
- 워크트리 안의 LSP 진단은 신뢰하지 않았다(알려진 거짓 신호).

## Residual-risk

- **CI 배선을 하지 않은 것이 이 카드의 가장 큰 잔여 위험이다.** 근거: 라이브 테스트는
  실제 codex / z.ai quota 를 쓰고(로컬에서 `TestHandleCodexReviewGate_Live` 1본이 59초 소요),
  2본은 자격증명을 요구한다. 레인이 CI 시크릿을 프로비저닝할 수 없으므로, 지금 CI 잡을
  추가하면 **자격증명이 없어 또 skip 되는 잡**이 생긴다 — 침묵을 없애는 게 아니라 새 자리로
  옮기는 것이다. **운영자 결정 상신**: (a) codex + 시크릿을 CI 에 프로비저닝해 nightly /
  workflow_dispatch 로 이 축을 관측할지, (b) 현행 선언 상태를 유지할지. (a) 를 택하면
  `TestCodexLiveAxis_NotObservedInCI` 가 RED 로 알려주며 선언 갱신을 강제한다 — 설계된 동작이다.
- 탐지 관용구(`exec.LookPath(codexBinaryName)` / `codexLookPath(...)` / 라이브 env 리터럴)를
  쓰지 않는 새 라이브 테스트는 가드에 잡히지 않는다. `t.Fatal("detected zero live-gated files")`
  가 관용구 전면 교체는 잡지만, 부분적으로 새로운 관용구는 못 잡는다.
- `MOAI_SKIP_LIVE_CODEX` 는 opt-**out** 스위치라 `NotObservedInCI` 의 env 목록에 섞여 있다.
  워크플로가 이걸 배선하면 "CI 가 축을 관측한다"가 아니라 "명시적으로 끈다"는 뜻인데도 RED 가
  난다. 그 경우에도 선언을 갱신해야 하는 건 맞으므로 오탐이 아니라 보수적 동작으로 남겼다.

---

## Card Cross-Check

| 카드 조치 항목 | 처리 | card |
|---|---|---|
| CI codex 설치 + 라이브 env 배선 **또는** 미관측 명시 기록 | 후자 선택, 기계 검사 선언으로 구현 | t495 |
| `codex_launcher_guards_test.go:150` 하한 단언 추가 | 완료 (E2) | t495 |
| 뮤턴트로 RED 선(先)확립 | 완료 (M1 공허 초록 실증 → 수리 → RED) | t495 |
| skip 유지 시 "침묵이 아닌 선언" | 완료 (선언 + 3 가드 + Makefile 타깃) | t495 |
