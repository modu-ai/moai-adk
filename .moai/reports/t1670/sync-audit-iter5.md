auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: 763fafafdb944e3e825753b2227913e32adbff95

## Evaluation Report — 카드 t1670 sync-audit 5회차 (델타 재감사, 운영자 승인 수리 회차)

SPEC: SPEC-LOCAL-MAIN-FLOW-001
Overall Verdict: FAIL — overall score 87/100

감사 대상: 워크트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670`, 브랜치 `WT-t1616-10-10`, HEAD `763fafafd`(전체 SHA는 위 `audited_sha`), 기준 local main `67e462e80`. 판정 대상 수리 커밋은 `a8f346716`(시험 전용, `internal/cli/local_main_resync_gap_test.go` +63/−13)이고, `763fafafd`는 `.moai/reports/t1670/` 문서만 바꿨다. 4회차 감사 커밋 `9429ec4f5` 이후 운영 코드는 바뀌지 않았다(`git diff --stat 9429ec4f5..HEAD` → 운영 `.go` 파일 0개).

### 요약

4회차에서 빠졌던 두 변형은 이번 수리로 확실히 잡힌다. 트립와이어를 읽기 seam 안(동기 탈취 probe 직후)으로 옮겨서 mutant H(읽기만 첫 구간, 캐시 검증과 이동은 둘째 구간)가 3/3 실패하고, 새 시험 `TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses`가 mutant D3(직접 읽기를 밖으로 끌어내고 seam 읽기는 안에 남김)를 3/3 잡는다. 4회차 Required fix 두 항목(F1 트립와이어 이동, F2 경쟁자 기록 반환 시험)이 모두 반영됐다. 이 감사가 다시 만든 A·B·C·D1·D2·D3·E·F·H와 새로 설계한 조합 DH(D3×H)는 모두 최소 한 시험에서 3/3 실패했고, 실제 트리는 세 시험 모두 3/3 통과했다.

그러나 **재검증과 `localMainResyncBeforeFastForward` 호출은 첫 구간에 남기고, 실제 `git merge --ff-only`만 둘째 `RunIntegrationMutation` 구간으로 옮긴 변형(mutant S)** 은 세 시험을 13회 중 12회 통과한다(`-count=3` 3/3 통과, `-count=10` 9/10 통과). 트립와이어 확인과 두 동기 탈취 시도가 모두 첫 구간 안에서 끝나므로, 두 구간 사이 공백은 goroutine 경쟁자의 타이밍 신호로만 보인다. 이 감사가 그 공백에 경쟁자 탈취를 끼워 넣은 probe를 돌린 결과, 경쟁자가 창을 가져간 뒤에도 이전 소유자가 local main을 fast-forward했다(3/3). 실제 코드의 같은 지점(이동 직전 seam)에서 시도한 탈취는 `BeforeFastForward` 시험이 보여 주듯 3/3 거절된다. 즉 카드가 막으려는 결함이 되살아나는데 시험은 대부분 통과한다. 필수 백엔드 codex도 같은 형태를 독립적으로 만들어 `fail`(P2)을 냈다. 필수 게이트가 FAIL이므로 이 감사는 PASS나 PASS-WITH-DEBT를 낼 수 없다.

**S의 개연성(운영자 질문에 대한 답).** S는 실수 한 번으로 생기는 변형이 아니다. 주석이 금지한 구간 분할을 의도적으로 하고, 동시에 "pre-fast-forward" 훅을 merge에서 떼어 검사 쪽에 남겨야 한다. 그래서 H보다 개연성이 낮다. 하지만 악의적 편집에서만 나오는 형태도 아니다. merge 동안 잠금을 오래 쥐지 않으려고 구간을 나눈 개발자(또는 에이전트)가 B 모양에서 트립와이어 실패를 보고 훅을 위로 올려 초록을 만드는 경로가 바로 S다. 시험을 통과시키려는 반복 수정이 자연스럽게 도달하는 지점이라는 점에서 "그럴듯한 회귀"로 판단한다. 다만 이 계열(관찰 훅이 보호 대상 연산에서 떨어지는 형태)은 seam 한 개로 통째로 닫힌다(F1 Required fix). 읽기 쪽 D 계열을 닫은 방식과 같다.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 80/100 | FAIL | 실제 트리 3시험 `-count=3` 9/9 PASS. 완료 기준 (2) 충족: 실제 트리에서 경쟁자 착지 0회. 완료 기준 (1)은 미충족: mutant S가 12/13 통과하고 gap probe가 실제 경합 재현(`rivalErr=<nil> \| head==target:true ... holder=sess-rival \| resyncErr=... releasing the window failed`). 필수 codex `fail` P2 |
| Security (25%) | 95/100 | PASS | 시험 전용 변경. 외부 입력·비밀 없음. 운영 동작 무변경(4회차 이후 운영 `.go` 0개). 잠금 경합 안전성은 실제 트리에서 유지 |
| Craft (20%) | 88/100 | PASS | `-race` 17 PASS / 0 FAIL / DATA RACE 0. `gofmt -l` 출력 없음, `go vet` exit 0, `golangci-lint --new-from-rev 2d6dbc9fd` `0 issues.`. goroutine 정리(30s 기한 + stop 채널 + `wg.Wait`) 유지. 동작이 올바른 사전 검증 리팩터(mutant P)를 거절하는 오탐 1건(선택) |
| Consistency (15%) | 90/100 | PASS | 읽기 seam 패턴과 같은 `prev`/`t.Cleanup` 복원 관례, 기존 픽스처(`lmfResyncFixture`, `lmfRival`) 재사용, 문서 주석이 H·D3 포괄 범위를 정확히 기술 |

가중 점수: 80×0.40 + 95×0.25 + 88×0.20 + 90×0.15 = 86.85 → 87. Functionality는 must-pass 차원이므로 이 차원 FAIL이 전체 FAIL을 결정한다. 필수 codex 게이트 FAIL도 독립적으로 같은 결론을 강제한다.

### Findings (structured defect-list)

- F1 [P2] [blocking] [확신도 높음] `internal/cli/local_main_resync_gap_test.go:224-228`(관찰 지점은 `localMainResyncBeforeFastForward` 훅) — 세 시험의 결정적 신호(읽기 직후 동기 탈취, 읽기에서 심은 트립와이어의 이동 직전 확인, 이동 직전 동기 탈취)가 모두 운영 코드의 **no-op 훅** 위치에서 관찰된다. 훅은 merge와 같은 줄에 붙어 있다는 소스상 인접성만으로 이동을 대표하므로, 훅을 첫 구간에 남기고 merge만 둘째 구간으로 옮기면(mutant S) 결정적 신호가 모두 빗나가고 goroutine 경쟁자의 확률 신호만 남는다(13회 중 1회 검출). gap probe로 실제 결함 재현을 확인했다. — Required fix: 관찰 지점을 **이동을 실제로 수행하는 호출**로 바꾼다. (a) `localMainResyncReadWindow`와 같은 방식으로 fast-forward 자체를 패키지 seam(예: `var localMainResyncFastForward = func(git func(...string) (string, error), baseline string) (string, error) { return git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline) }`)으로 빼고 구간 안에서 호출한다. (b) DuringRevalidate 시험은 그 seam을 감싸, 통과 호출(call-through) **직전에** 트립와이어 생존을 확인하고 동기 탈취 probe가 lock-busy인지 확인한다. (c) 감싼 seam이 한 번도 실행되지 않았으면 `Fatalf`하고, 기존 HEAD==target 확인을 유지한다(seam을 우회해 이동하면 seam-미실행으로, 이동하지 않으면 HEAD 확인으로 걸린다). (d) S를 레인의 음성 대조 목록에 추가해 3/3 실패를 기록한다. 대안으로 PATH 앞에 둔 `git` 셸 래퍼가 `merge --ff-only` 호출에서 트립와이어를 확인하는 시험 전용 방식도 가능하지만, Windows 경로와 셸 의존이 생기므로 (a)가 더 단순하다.
- F2 [P3] [optional] [확신도 높음] `internal/cli/local_main_resync_gap_test.go:163-170` — `readProbed`가 첫 읽기에만 probe와 트립와이어를 걸기 때문에, 구간 밖에서 미리 실패를 빨리 내는 재검증을 하나 더 두고 구간 안 재검증도 유지한 **동작상 올바른** 리팩터(mutant P)를 거절한다(3/3, 첫 probe가 구간 밖에서 착지해 창을 가져가므로 재동기화가 NotHolder로 끝나고 212행 `Fatalf`). 안전 쪽 오탐이라 차단 사유는 아니다. — Fix(선택): F1의 seam을 도입하면 "이동을 승인한 마지막 읽기부터 이동까지"를 확인하는 형태로 바꿀 수 있다. 굳이 고치지 않는다면 시험 주석에 "구간 밖 사전 검증은 이 시험이 거절한다"고 적어 두면 충분하다.
- F3 [info] [optional] 4회차 Required fix 반영 확인: F1(트립와이어를 읽기 seam 안, probe 직후로 이동) 반영, F2(seam이 경쟁자 기록을 돌려주면 `MergeExitNotHolder` + HEAD 불변) 반영. 두 항목 모두 커밋 `a8f346716` diff에서 확인했다.
- F4 [info] [optional] 설치된 `moai` 바이너리와 MCP 서버가 `ee981a932`로 빌드돼 이 트리 HEAD의 조상이다(`audit_multi` 결과의 `build_lag`). 계획 판독과 codex 실행이 그 빌드로 이뤄졌다. 이 카드는 감사 계획 코드에 손대지 않았으므로 영향은 없다고 보지만 트리 빌드로 재확인하지 않았다(Gaps 참조).

### Recommendations

- F1 수리는 시험 파일 + 운영 seam 한 줄(동작 불변 리팩터)이면 충분하다. 수리 뒤 mutant 집합(A·B·C·D1·D2·D3·E·F·H·DH·S)을 `/private/tmp/t70-audit5/mutate.py`로 다시 돌려 각각 최소 한 시험에서 3/3 실패를 확인한다.
- `localMainResyncBeforeFastForward` 훅은 F1 seam이 생기면 역할이 겹친다. 정리 여부는 범위 밖이므로 남겨 두되, 남긴다면 F1 seam 안에서 호출해 인접성이 구조로 보장되게 하는 편이 낫다.

## Claim

1. 실제 트리(HEAD `763fafafd`)는 세 rival 시험을 각 3/3 통과한다.
2. 4회차에서 빠진 H와 D3는 이제 각각 DuringRevalidate와 SeenByOwnershipRead 시험에서 3/3 실패한다.
3. 재구성한 A·B·C·D1·D2·D3·E·F·H와 새 조합 DH는 모두 최소 한 시험에서 3/3 실패한다. G(동작 올바름, seam 우회)와 P(동작 올바름, 사전 검증)는 안전 쪽 오탐으로 실패한다.
4. 새 변형 S는 세 시험을 12/13 통과하며, 실제 경합(경쟁자가 창을 가진 상태에서 이전 소유자가 fast-forward)을 되살린다.
5. `-race`에서 데이터 경합이 없고, vet·gofmt·lint가 깨끗하며, `TestAuditLagUsesBinlagSeam`이 통과한다.
6. 완료 기준 (2) "잠금을 지키는 경합에서 창 재현 0회"는 실제 트리에서 충족되고, 완료 기준 (1) "재검증과 이동이 한 잠금 구간을 공유함을 시험이 확인"은 S 때문에 충족되지 않는다.

## Evidence

E1 — mutant 재구성. HEAD 트리를 `git archive`로 `/private/tmp/t70-audit5/head.tar`에 담아 트리별로 풀고, 4회차 스크립트(`/private/tmp/t70-mut5/mutate.py`)에 새 변형 P·S·DH를 덧붙인 `/private/tmp/t70-audit5/mutate.py`로 `internal/cli/local_main_resync.go`만 바꿨다. 시험 파일은 14개 트리 모두 HEAD와 `cmp` 동일(14줄 출력). REAL은 `head.tar` 원본과 `diff` 무출력(`REAL-unchanged`). 모든 트리 `go vet ./internal/cli/` exit 0.

E2 — 실행 명령(트리별 순차, 시간 민감 시험): `TMPDIR=GOTMPDIR=/private/tmp/t70-audit5/tmp go -C <tree> test -count=3 -v -run 'TestLocalMainResyncRivalTakeover(BeforeFastForward|DuringRevalidate)CannotLand$|TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses$' ./internal/cli/`. 요약(`/private/tmp/t70-audit5/logs/summary-main.txt`), 트리별 실패 수 BFF / DR / SEEN:

| 트리 | 형태 | BFF | DR | SEEN | 판정 |
|---|---|---|---|---|---|
| REAL | 실제 코드 | 0/3 | 0/3 | 0/3 | 통과 (`ok ... 56.208s`) |
| A | 재검증을 구간 밖으로 | 0/3 | 3/3 | 0/3 | 잡힘 (224·227행) |
| B | 재검증과 이동을 두 구간으로 | 0/3 | 3/3 | 0/3 | 잡힘 (227행) |
| C | 결과를 구간 앞에서 미리 계산 | 0/3 | 3/3 | 0/3 | 잡힘 (224·227행) |
| D1 | seam 읽기를 구간 밖으로 | 0/3 | 3/3 | 0/3 | 잡힘 (224·227행) |
| D2 | 직접 읽기를 구간 밖으로 | 0/3 | 3/3 | 3/3 | 잡힘 (221·284행) |
| D3 | 직접 읽기 밖, seam 읽기 안(lease만) | 0/3 | 0/3 | 3/3 | 잡힘 (287·290행) — 4회차 미검출 |
| E | 이동을 구간 뒤로 | 3/3 | 3/3 | 0/3 | 잡힘 (45·238·246행) |
| F | 뮤테이션 잠금 제거 | 3/3 | 3/3 | 0/3 | 잡힘 (45·224·238·246행) |
| G | 대조: seam 우회, 직접 읽기는 구간 안 | 0/3 | 3/3 | 3/3 | 안전 쪽 오탐 (221·284행) |
| H | 읽기만 첫 구간, 캐시 검증+이동 둘째 구간 | 0/3 | 3/3 | 0/3 | 잡힘 (227행) — 4회차 미검출 |
| DH (신규) | D3×H 조합 | 0/3 | 3/3 | 3/3 | 잡힘 (227·287·290행) |
| P (신규) | 동작 올바름: 구간 밖 사전 재검증 + 구간 안 재검증 | 0/3 | 3/3 | 0/3 | 안전 쪽 오탐 (212행: `pre-fast-forward seam ran = false`, 재동기화가 `refused — sess...`로 끝남) |
| S (신규) | 훅은 첫 구간, merge만 둘째 구간 | 0/3 | 0/3 | 0/3 | **놓침** (`ok ... 56.064s`) |

REAL 원문(발췌):

```
REAL-main: --- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (8.80s)
REAL-main: --- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (8.76s)
REAL-main: --- PASS: TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses (0.42s)
... (×3, DR 마지막 회차 9.34s, SEEN 1.06s)
```

S 원문(발췌, `-count=3`):

```
S-main: --- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (8.93s)
S-main: --- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (9.00s)
S-main: --- PASS: TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses (0.42s)
... (×3 모두 PASS)
```

E3 — S 추가 `-count=10`(`/private/tmp/t70-audit5/logs/summary-s10.txt`):

```
   1 --- FAIL: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand
  10 --- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand
   9 --- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand
  10 --- PASS: TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses
   1 local_main_resync_gap_test.go:238: a rival --force takeover landed while primary HEAD was still 96e45f21..., before the fast-forwar
```

유일한 검출은 goroutine 경쟁자 착지 신호(238행)였다. 합계 S 검출률 1/13.

E4 — S가 실제 결함인지 확인한 gap probe. S 트리를 복사한 `/private/tmp/t70-audit5/S2`에서 두 구간 사이에 `mutHGap(repoRoot)` 훅을 넣고(`s2patch.py`), 그 훅에서 경쟁자 `--force` 탈취를 시도하는 시험 `zz_mut_s_probe_test.go`(4회차 H probe와 같은 본문)를 `-count=3`으로 실행:

```
zz_mut_s_probe_test.go:28: rivalErr=<nil> | head==target:true head==base:false | holder=sess-rival | resyncErr=integration resync: local main fast-forwarded b93ec56fa619 -> 7c807161672d to origin/main, but releasing the window failed: release integration window is held by a different session: lane-rival (pid 88585) holds it (moai integration status reads it)
--- PASS: TestAuditMutSGapProbe (0.54s)
... (3회 모두 같은 결과)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.463s
```

경쟁자 탈취가 성공(`rivalErr=<nil>`)한 뒤에도 이전 소유자가 fast-forward했다. 실제 코드의 대응 지점(이동 직전, 구간 안) 대조는 E2의 REAL `BeforeFastForward` 3/3 PASS다 — 그 시험은 바로 그 지점에서 탈취를 시도하고 거절(`rivalErr != nil`)을 요구한다.

E5 — `-race`(실제 트리): `go -C <worktree> test -race -count=1 -v -run 'TestLocalMainResync|TestAuditLagUsesBinlagSeam' ./internal/cli/`

```
RACE_EXIT=0
PASS lines: 17  FAIL lines: 0  DATA RACE: 0
--- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (9.01s)
--- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (9.62s)
--- PASS: TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses (0.56s)
--- PASS: TestAuditLagUsesBinlagSeam (0.36s)
ok  	github.com/modu-ai/moai-adk/internal/cli	31.610s
```

E6 — 정적 검사(실제 트리): `gofmt -l internal/cli/local_main_resync_gap_test.go internal/cli/local_main_resync.go` → 출력 없음, `gofmt_exit=0`. `go vet ./internal/cli/` → `vet_exit=0`. `golangci-lint run --new-from-rev 2d6dbc9fd ./internal/cli/` → `0 issues.`, `lint_exit=0`.

E7 — 감사 계획: `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670` → `config_status: ok`, `model: codex`, codex `gate: required`, claude·glm `off`, `cross_model_active: true`, `enforced_required: ["codex"]`.

E8 — `mcp__moai__audit_multi`(project_root=워크트리, target=baseBranch, gates 인자 없음, 계약 모드가 `guided`라 card_id 없음). 오케스트레이터 확인용 digest:

```
overall_verdict: fail
gate_unmet: (결과에 없음)
plan_source: config
per_backend_verdicts:
  - backend: codex, gate: required, verdict: fail
audit_receipt: rcpt-796a3ab9955032683df32987
```

codex 요지: P2 — 재검증과 `localMainResyncBeforeFastForward`를 첫 구역에 남기고 merge만 둘째 구역으로 옮긴 overlay가 세 시험 모두 통과한 회차가 3회 중 2회, 별도 probe로 공백에서의 경합 재현(이 감사의 S·E4와 독립 일치). P3 — 첫 읽기에 고정된 트립와이어가 안전한 사전 검증 추가를 거절(이 감사의 P와 독립 일치). 이 판정은 오케스트레이터가 `audit-plan-result.json`으로 `moai verify audit-plan --result-file` 확인을 통과시키기 전까지 최종이 아니다.

## Baseline-attribution

- 모든 시험·정적 검사는 이번 실행에서 HEAD `763fafafd`(=`audited_sha`) 트리에 대해 수행했다. mutant 트리는 같은 HEAD의 `git archive` 산출물에서 운영 파일 한 개만 바꿨고 시험 파일은 `cmp`로 HEAD와 동일함을 확인했다.
- 환경: `TMPDIR=GOTMPDIR=/private/tmp/t70-audit5/tmp`($HOME 밖). 시간 민감 시험이라 트리별 순차 실행, 동시에 무거운 작업을 돌리지 않았다(lint·vet는 시험 배치가 끝난 뒤 실행).
- 4회차 수치(85/100, H·D3 미검출)는 비교 기준으로만 인용했고, 이번 판정의 근거는 모두 이번 실행 출력이다.
- 도구 출처(§2.2): `go`·`gofmt`·`golangci-lint`는 저장소 빌드가 아닌 외부 도구라 해당 없음. `moai verify audit-plan`과 MCP 서버는 `ee981a932` 빌드(트리 HEAD의 조상)로 판정했다 — Gaps 참조.

## Gaps

- 감사 계획 판독(E7)과 codex 실행(E8)은 설치 빌드 `ee981a932`가 수행했다. 이 트리에서 빌드한 바이너리로 재확인하지 않았다. 이 카드는 감사 계획 코드를 바꾸지 않았다는 판단은 diff 범위(`internal/cli/local_main_resync*.go`, `mcp_build_identity_test.go`, 보고서)에 근거한 추론이다.
- Windows 경로(잠금 아티팩트의 create-new 방식)는 macOS에서만 돌렸으므로 관찰하지 않았다. 트립와이어 메커니즘이 Windows에서도 "획득마다 아티팩트를 다시 쓴다"는 전제는 시험 주석과 양성 대조(종료 시 트립와이어 소멸) 확인에 기대며, 이번 실행에서는 darwin만 관찰했다.
- 전체 `internal/cli` 스위트는 돌리지 않았다(범위 한정 지침). 영향 범위 시험(`TestLocalMainResync*`, `TestAuditLagUsesBinlagSeam`)만 관찰했다.
- 커버리지 수치는 측정하지 않았다. 변경이 시험 전용이라 운영 커버리지에 영향이 없다고 보지만 측정값은 없다.
- 거절된 도구 호출: worktree 격리 가드가 복합 셸 명령 몇 개를 거절했다(`git`/`go` 문자열이 섞인 heredoc·루프). 같은 작업을 스크립트 파일(`/private/tmp/t70-audit5/*.sh`, `*.py`)로 나눠 실행했고, 측정 자체를 대체한 것은 없다.

## Residual-risk

- S 계열은 F1 seam으로 닫히지만, 그 seam 역시 "fast-forward를 seam 밖에서 다른 방식(예: `update-ref`)으로 수행"하는 변형에는 seam-미실행 `Fatalf`와 HEAD 확인에 기댄다. 그 경우는 이동 수단 자체를 바꾸는 편집이라 개연성이 낮다.
- goroutine 경쟁자 신호는 확률적이라, 기계 부하에 따라 S류 공백의 검출률이 바뀐다(이번 1/13). 결정적 신호가 이동 호출에 붙기 전까지 이 신호를 결정적 증거로 간주하면 안 된다.
- `/private/tmp/t70-audit5/logs/` 원문 로그는 머신 로컬 스크래치이며 이 파일에 옮긴 줄만 인용 대상이다. 나머지 로그(트리별 전체 `-v` 출력)는 알려진 손실로 남는다.
- 운영 동작 자체는 다섯 차례 감사 내내 올바르다고 판정됐고 이번에도 실제 트리 경합 0회였다. 남은 위험은 회귀 방지 시험의 강도이지 현재 코드의 결함이 아니다.

## Iteration history

| 회차 | 감사 SHA | 판정 | 점수 | 놓친 변형 |
|---|---|---|---|---|
| 1 | a2055ca46 | FAIL | 72 | 재검증 위치를 시험이 고정하지 못함 |
| 2 | b79a70d97 | FAIL | 78 | B(두 구간) |
| 3 | 9aa451b10 | FAIL | 82 | D(읽기만 구간 밖) |
| 4 | 9429ec4f5 | FAIL | 85 | H(읽기 자체 구간 + 캐시), D3(직접 읽기 밖, seam 안) |
| 5 | 763fafafd | FAIL | 87 | S(훅은 첫 구간, merge만 둘째 구간) — 12/13 통과, 실제 경합 재현 |

이번 회차 판정 영수증: rcpt-796a3ab9955032683df32987.
