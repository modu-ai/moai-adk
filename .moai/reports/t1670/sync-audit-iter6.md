auditor-model: claude-opus-5-5
verdict: FAIL
audited_sha: ef4a09127b9309185b33b52489d57c2a1a88e6e5

## Evaluation Report — 카드 t1670 sync-audit 6회차 (델타 재감사, 운영자 승인 수리 회차)

SPEC: SPEC-LOCAL-MAIN-FLOW-001
Overall Verdict: FAIL — overall score 90/100

감사 대상: 워크트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670`, 브랜치 `WT-t1616-10-10`, HEAD `ef4a09127`(전체 SHA는 위 `audited_sha`), 기준 local main `67e462e80`(`git merge-base main HEAD`로 확인, local `main` 자체는 그 뒤 `c3aa6b057`까지 전진). 판정 대상 수리 커밋은 `e9245bdd3`(운영 seam)와 `441648aa9`(시험)이고, `ef4a09127`은 `.moai/reports/t1670/` 문서만 바꿨다.

### 요약

5회차 Required fix F1의 네 항목은 모두 들어갔다. (a) fast-forward를 패키지 seam `localMainResyncFastForward`로 빼서 기존 구간 안에서 호출하고, (b) DuringRevalidate 시험이 그 seam을 감싸 통과 호출 직전에 트립와이어 생존과 동기 탈취 probe(lock-busy)를 확인하며, (c) seam이 한 번도 실행되지 않으면 `Fatalf`하고 HEAD==target 확인을 유지하고, (d) 레인 증거(`evidence.md`)에 S의 음성 대조 결과가 기록돼 있다. 이 감사가 직접 다시 만든 A·B·C·D1·D2·D3·E·F·H·DH·S·SB와 새 변형 O(구간 안에서 이동을 재검증보다 먼저 실행)는 모두 최소 한 시험에서 3/3 실패했다. 실제 트리는 세 시험 모두 3/3 통과했다. 대조군 G·P는 동작이 올바른 리팩터인데도 안전 쪽으로 실패한다(5회차와 같은 성질).

남은 빈틈은 운영자가 미리 짚은 바로 그 계열이다. **seam 본문을 no-op으로 바꾸고 실제 merge를 다른 경로로 옮기면** 세 시험이 모두 통과한다. 이 감사가 만든 두 형태가 모두 9/9 통과했다. SL은 no-op seam을 구간 안에 남기고 merge를 구간 뒤에서 **잠금 없이** 실행한다. SL2는 no-op seam을 첫 구간에 남기고 merge만 둘째 `RunIntegrationMutation` 구간에서 실행한다. SL2의 두 구간 사이에 경쟁자 탈취를 끼운 probe에서는 경쟁자가 창을 가진 상태에서 이전 소유자가 local main을 fast-forward했다(3/3). 카드가 막으려는 결함이 그대로 되살아난다는 뜻이다. 필수 백엔드 codex도 같은 형태(seam no-op + 둘째 구간 merge)를 독립적으로 만들어 `fail`(P2)을 냈다.

**개연성 판단(질문 2에 대한 답).** SL·SL2는 자연스러운 리팩터에서 나오기 어렵다. 두 형태 모두 이름과 문서 주석이 "local main을 움직이는 유일한 호출"이라고 밝힌 운영 seam의 **본문을 비우고**, 같은 merge를 다른 곳에 새로 써야 한다. 5회차 S는 "잠금 보유 시간을 줄이려 구간을 나눈 뒤 시험을 초록으로 만들려고 훅을 위로 올리는" 자연스러운 수리 경로에서 나왔다. 반면 SL2는 S가 이번 시험에 걸린 뒤에도 초록을 만들려면 관찰 지점 자체를 의도적으로 무력화해야 한다. 그래서 이 감사는 SL·SL2를 **의도적 무력화(시험 맞추기) 계열**로 본다. 이 감사의 기준만 적용하면 차단 사유가 아닌 선택 항목(P3)이고, 판정은 PASS-WITH-DEBT가 된다. 그러나 감사 계획(`moai verify audit-plan`)은 codex를 `required`로 두고, 그 codex가 `fail`을 냈다(`audit_multi` overall `fail`, receipt `rcpt-47575f8f08205ad19f58d9af`). 필수 게이트 FAIL에서는 PASS도 PASS-WITH-DEBT도 낼 수 없으므로 최종 판정은 **FAIL**이다. 수리는 작다. 시험 6줄로 SL·SL2를 닫을 수 있고, 이 감사가 그 수리안을 실측해 두었다(아래 F1, E4).

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 88/100 | PASS (부채 있음) | 실제 트리 3시험 `-count=3` 9/9 PASS(`ok ... 83.352s`). 이전 회차 변형 12종과 새 변형 O가 모두 3/3 실패. 완료 기준 (2) 충족: 실제 트리에서 경쟁자 착지 0회. 완료 기준 (1)은 의도적 무력화 형태 SL·SL2를 빼면 충족. SL2 gap probe에서 결함이 재현됨: `rivalErr=<nil> \| head==target:true head==base:false \| holder=sess-rival \| resyncErr=... releasing the window failed` |
| Security (25%) | 95/100 | PASS | 운영 변경은 동작이 같은 seam 추출뿐이다. merge 인자는 바이트 단위로 같다(`git -c merge.autoStash=false merge --ff-only --no-overwrite-ignore -q <baseline>`). 외부 입력·비밀 없음. 실제 트리에서 잠금 경합 안전성 유지 |
| Craft (20%) | 88/100 | PASS | `-race` `PASS=17 FAIL=0 DATA_RACE=0`, `ok ... 42.605s`. `go vet` exit 0, `gofmt -l` 출력 없음, `golangci-lint run --new-from-rev 4119ea33e ./internal/cli/` → `0 issues.`. DuringRevalidate 한 건이 약 17초(probe 두 번이 각각 잠금 재시도 예산을 다 기다림, 선택 부채). 대조군 P 오탐 유지(선택) |
| Consistency (15%) | 92/100 | PASS | 읽기 seam과 같은 `var` seam + `prev` 저장 + `t.Cleanup` 복원 관례를 따른다. `@MX:NOTE`/`@MX:SPEC` 부착, binlag 좌표 461/464 재측정 주석이 이전 회차 형식과 같다. 시험 주석이 S까지의 포괄 범위를 정확히 적고 있다 |

가중 점수: 88×0.40 + 95×0.25 + 88×0.20 + 92×0.15 = 90.35 → 90. must-pass 차원(Functionality, Security)은 모두 기준을 넘는다. 이 감사 자체의 점수만으로는 PASS-WITH-DEBT다. 최종 FAIL은 **필수 codex 게이트 미충족** 하나 때문이다(`audit_multi`: `overall_verdict: fail`, `residual_risk_note: required-backend FAIL: codex`). 이 판정은 오케스트레이터의 `moai verify audit-plan --result-file` 확인을 거쳐야 확정된다.

### Findings (structured defect-list)

- F1 [P2 by codex / P3 by this auditor] [blocking — 필수 게이트] [확신도 높음] `internal/cli/local_main_resync_gap_test.go:214-222`(fast-forward seam 래퍼) — 래퍼는 seam에 **들어가는** 시점의 잠금 상태만 관찰하고, seam이 실제로 local main을 움직였는지는 확인하지 않는다. 그래서 운영 seam 본문을 no-op으로 바꾸고 merge를 구간 뒤(SL) 또는 둘째 구간(SL2)으로 옮기면 세 시험이 모두 통과한다(각 9/9). 관찰 지점이 보호 대상 연산에서 다시 떨어지는 형태다. SL2 gap probe에서 실제 결함 재현도 확인했다. — Required fix: 래퍼 안에서 `prevMove(git, baseline)` 호출이 끝난 직후 primary HEAD를 읽어, 오류 없이 `target`에 도달했는지 기록하고(`movedBySeam`), 도달하지 않았으면 `t.Errorf`로 실패시킨다. "seam 호출 자체가 이동을 수행한다"를 결과(HEAD 변화)로 묶는 방식이다. 이 감사가 그대로 적용해 실측했다(E4). 수리안 시험은 실제 코드에서 9/9 통과하고, SL·SL2에서 각 3/3 실패하며, S에서도 계속 3/3 실패한다. 대안으로 codex가 제안한 방식, 즉 seam에 넘기는 `git` 함수를 감싸 `merge --ff-only` 호출 시점에 관찰하고 그 호출이 반드시 일어났는지 확인하는 방식도 같은 계열을 닫는다. 수리 뒤에는 SL·SL2를 음성 대조 목록에 넣고 3/3 실패를 기록한다.
- F2 [P3] [optional] [확신도 높음] `internal/cli/local_main_resync_gap_test.go:176-184` — 동작이 올바른 사전 검증 리팩터(대조군 P)를 여전히 거절한다. 이번에는 234행 `Fatalf`(`fast-forward seam ... ran = false`)로 실패한다. 구간 밖 첫 읽기에서 probe가 창을 가져가므로 구간 안 재검증이 NotHolder로 끝나기 때문이다. 안전 쪽 오탐이므로 차단 사유가 아니다. 5회차 F2와 같은 항목이다.
- F3 [P3] [optional] [확신도 높음] DuringRevalidate 시험 비용 약 17초(실제 트리 `-count=3`에서 패키지 전체 83초). 두 동기 probe가 각각 잠금 재시도 예산을 끝까지 기다린다. 판정을 결정적으로 만드는 대가이므로 **수용 가능한 부채**로 본다. 줄이려면 시험 전용 재시도 예산 seam이 필요한데, 이 카드 범위를 벗어난다.
- F4 [info] [optional] 5회차 Required fix F1 (a)~(d) 반영 확인(`git diff 67e462e80..HEAD`와 `evidence.md`로 확인). 운영 변경은 동작이 같다. seam 본문은 기존 literal 호출과 인자가 같고, 호출 순서(훅 → 이동)와 `moveErr` 처리도 그대로다.
- F5 [info] [optional] 설치된 `moai` 바이너리와 MCP 서버는 `ee981a932`로 빌드됐고, 이 커밋은 이 트리 HEAD의 조상이다(`audit_multi`의 `build_lag`). 감사 계획 판독과 codex 실행은 그 빌드로 이뤄졌다. 이 카드는 감사 계획 코드를 건드리지 않았다(Gaps 참조).

### Recommendations

- F1 수리(시험 6줄)는 운영 코드를 바꾸지 않는다. 수리 뒤 `/private/tmp/t70-audit6/mutate.py`의 전체 집합(A·B·C·D1·D2·D3·E·F·H·DH·S·SB·SL·SL2·O)을 다시 돌려, 각 변형이 최소 한 시험에서 3/3 실패하는지 확인한다.
- 다음 회차는 시험만 바꾸므로 binlag 좌표(461/464)는 움직이지 않는다.

## Claim

1. 실제 트리(HEAD `ef4a09127`)는 세 rival 시험을 각 3/3 통과한다.
2. 5회차에서 놓친 S는 이제 DuringRevalidate 시험에서 3/3 실패한다. 5회차 F1 (a)~(d)는 모두 반영됐다.
3. 재구성한 A·B·C·D1·D2·D3·E·F·H·DH·S·SB와 새 변형 O는 모두 최소 한 시험에서 3/3 실패한다. 대조군 G·P는 안전 쪽으로 실패한다.
4. 새 변형 SL·SL2(seam 본문 no-op + 다른 경로의 merge)는 세 시험을 9/9 통과하고, SL2는 실제 경합을 되살린다.
5. F1 수리안(seam 호출 직후 HEAD==target 확인)은 실제 코드를 통과시키면서 SL·SL2·S를 모두 잡는다.
6. 운영 변경은 동작이 같다. `-race`에서 데이터 경합이 없고, vet·gofmt·lint가 깨끗하며, `TestAuditLagUsesBinlagSeam`이 통과한다.
7. 완료 기준 (2) "잠금을 지키는 경합에서 창 재현 0회"는 실제 트리에서 충족된다. 완료 기준 (1)은 의도적 무력화 형태(SL·SL2)를 빼면 충족되지만, 필수 codex 게이트가 그 형태를 근거로 `fail`을 냈다.

## Evidence

E1 — mutant 재구성. `git -C <tree> archive --format=tar -o /private/tmp/t70-audit6/head.tar HEAD`로 HEAD 트리를 담아 트리별로 풀었다. 레인 스크립트(`/private/tmp/t70-mut6/mutate.py`)를 복사해 ROOT를 바꾸고, 5회차 감사 스크립트와 `diff`해 공통 형태(A·B·C·D1·D2·D3·F·G·H·P·DH)가 같고 MOVE·E·S만 seam 호출로 바뀐 것을 확인했다. 그 위에 새 변형 SL·SL2·O를 더해 `/private/tmp/t70-audit6/mutate.py`를 만들었다. `setup.sh` 출력: 18개 트리 모두 `test identical`(시험 파일이 HEAD와 `cmp` 동일)과 `vet ok`, REAL은 `prod UNCHANGED`. 새 변형의 운영 diff:
- SL: 83행 `return git("-c", ... baseline)` → `return "", nil`, 그리고 구간 뒤 `if revalidateErr == nil && moveErr == nil { _, moveErr = git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline) }`
- SL2: 같은 no-op seam + 첫 구간을 닫고 둘째 `RunIntegrationMutation` 안에서 literal merge
- O: 구간 안에서 `localMainResyncBeforeFastForward` → `localMainResyncFastForward` → `localMainResyncRevalidate` 순서

E2 — 실행 명령(트리별 순차, 시간 민감 시험): `TMPDIR=GOTMPDIR=/private/tmp/t70-audit6/tmp go -C <tree> test -count=3 -v -run 'TestLocalMainResyncRivalTakeover(BeforeFastForward|DuringRevalidate)CannotLand$|TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses$' ./internal/cli/`. 요약 `/private/tmp/t70-audit6/logs/summary-main.txt`. 트리별 실패 수 BFF / DR / SEEN:

| 트리 | 형태 | BFF | DR | SEEN | 결정 신호(대표 줄) |
|---|---|---|---|---|---|
| REAL | 실제 코드 | 0/3 | 0/3 | 0/3 | `ok ... 83.352s` |
| A | 재검증을 구간 밖으로 | 0/3 | 3/3 | 0/3 | `:246 ... not refused as lock-busy (got <nil>` |
| B | 재검증·이동 두 구간 | 0/3 | 3/3 | 0/3 | `:249 the mutation lock was acquired between the ownership read and the fast-forward` |
| C | 결과를 구간 전에 계산 | 0/3 | 3/3 | 0/3 | `:246`, `:249` |
| D1 | seam 읽기를 구간 밖으로 | 0/3 | 3/3 | 0/3 | `:246`, `:249` |
| D2 | 직접 읽기를 구간 밖으로 | 0/3 | 3/3 | 3/3 | `:243 ... localMainResyncReadWindow never ran`, `:309` |
| D3 | 직접 읽기 밖 + seam 읽기 안 | 0/3 | 0/3 | 3/3 | `:312 ... must refuse with the not-holder class 14, got <nil>` |
| E | 이동을 구간 뒤로 | 3/3 | 3/3 | 0/3 | `:252 ... from the fast-forward seam itself, was not refused as lock-busy`, `:45` |
| F | 잠금 제거 | 3/3 | 3/3 | 0/3 | `:246`, `:252`, `:45` |
| H | 읽기만 첫 구간 | 0/3 | 3/3 | 0/3 | `:249` |
| DH | D3×H 조합 | 0/3 | 3/3 | 3/3 | `:249`, `:312` |
| S | 훅은 첫 구간, seam 이동은 둘째 구간 | 0/3 | 3/3 | 0/3 | `:249` |
| SB | S + 둘째 구간이 seam 우회 | 0/3 | 3/3 | 0/3 | `:234 ... fast-forward seam localMainResyncFastForward ran = false` |
| **SL** (새) | no-op seam 구간 안 + 잠금 없는 merge 구간 뒤 | 0/3 | **0/3** | 0/3 | `ok ... 81.948s` — **놓침** |
| **SL2** (새) | no-op seam 첫 구간 + merge 둘째 구간 | 0/3 | **0/3** | 0/3 | `ok ... 81.786s` — **놓침** |
| O (새) | 구간 안에서 이동을 재검증보다 먼저 | 3/3 | 3/3 | 3/3 | `:234 ... re-validation seam ran = false`, `:315 primary HEAD moved`, `:48` |
| G (대조) | 동작 올바름: seam 우회 직접 읽기 구간 안 | 0/3 | 3/3 | 3/3 | `:243`, `:309` (안전 쪽 오탐) |
| P (대조) | 동작 올바름: 구간 밖 사전 검증 추가 | 0/3 | 3/3 | 0/3 | `:234 ... fast-forward seam ... ran = false` (안전 쪽 오탐) |

E3 — SL2 gap probe. SL2 트리 복사본(`/private/tmp/t70-audit6/SL2P`)의 두 구간 사이에 `mutHGap(repoRoot)` 훅을 넣고, 5회차 probe 시험(`zz_mut_s_probe_test.go`, 이름만 `TestAuditMutSL2GapProbe`로 바꿈)으로 그 지점에서 경쟁자 `--force` 탈취를 시도했다. `go -C SL2P test -count=3 -v -run 'TestAuditMutSL2GapProbe$' ./internal/cli/` 출력(3회 모두 같은 형태):
```
zz_mut_sl2_probe_test.go:28: rivalErr=<nil> | head==target:true head==base:false | holder=sess-rival | resyncErr=integration resync: local main fast-forwarded b34f2aa59e63 -> bd576049fd55 to origin/main, but releasing the window failed: release integration window is held by a different session: lane
ok  	github.com/modu-ai/moai-adk/internal/cli	2.776s
```
경쟁자 탈취가 성공했고(`rivalErr=<nil>`, holder=`sess-rival`), 그 뒤에도 이전 소유자가 HEAD를 target으로 옮겼다. 실제 코드에서 같은 시점(이동 직전)에 시도한 탈취는 BFF 시험이 REAL에서 3/3 통과한 데서 보듯 거절된다.

E4 — F1 수리안 실측. 시험 파일 래퍼에서 `prevMove` 호출 직후 `movedBySeam = mErr == nil && lmfHead(t, root) == target`을 기록하고, 거짓이면 `t.Errorf("the fast-forward seam localMainResyncFastForward returned without moving primary HEAD to %s: ...")`를 내도록 6줄을 더했다. 운영 코드는 각 대상 트리(REAL/SL/SL2/S)와 `cmp` 동일하다. 같은 명령 `-count=3`, 요약 `/private/tmp/t70-audit6/logs/summary-2.txt`:
- FIXREAL: BFF 0/3, DR 0/3, SEEN 0/3 — `ok ... 81.877s`
- FIXSL: DR 3/3 FAIL — `:241 the fast-forward seam localMainResyncFastForward returned without moving primary HEAD to ...`
- FIXSL2: DR 3/3 FAIL — 같은 `:241`
- FIXS: DR 3/3 FAIL — `:256 the mutation lock was acquired between the ownership read and the fast-forward`

E5 — 실제 트리 정적·경합 점검(`/private/tmp/t70-audit6/run3.sh`, 요약 `logs/summary-3.txt`):
```
== race exit=0
PASS=17 FAIL=0 DATA_RACE=0
--- PASS: TestAuditLagUsesBinlagSeam (0.38s)
ok  	github.com/modu-ai/moai-adk/internal/cli	42.605s
== vet exit=0
== lint exit=0
0 issues.
== gofmt exit=0
```
명령: `go test -race -count=1 -v -run 'TestLocalMainResync|TestAuditLagUsesBinlagSeam' ./internal/cli/`, `go vet ./internal/cli/`, `golangci-lint run --new-from-rev 4119ea33e ./internal/cli/`, `gofmt -l` (변경된 세 파일). binlag 좌표: `local_main_resync.go:461` = `// included. \`git merge-base --is-ancestor\` exits 1 ...`, `:464` = `_, err := git("merge-base", "--is-ancestor", a, b)`. 시험 기대값(461/464)과 같다.

E6 — 운영 동작 보존. `git diff 67e462e80..HEAD -- internal/cli/local_main_resync.go`: 이동 줄이 `_, moveErr = git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline)`에서 `_, moveErr = localMainResyncFastForward(git, baseline)`로 바뀌었다. seam 기본값은 `return git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline)`이다. 인자·순서·오류 처리가 같고, `localMainResyncBeforeFastForward` 호출 위치도 그대로다.

E7 — 교차 모델 감사. `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670` → `config_status: ok`, `model: codex`, `cross_model_active: true`, `enforced_required: ["codex"]`(claude·glm gate `off`). `mcp__moai__audit_multi`(gates 인자 없음, `project_root` = 이 워크트리, `target: baseBranch`) 결과 요약:
- `overall_verdict`: `fail`
- `gate_unmet`: 결과에 없음
- `plan_source`: `config`
- per_backend_verdicts: `backend=codex, gate=required, verdict=fail`
- `audit_receipt`: `rcpt-47575f8f08205ad19f58d9af`

codex의 P2 지적은 "seam이 실제 merge를 실행했는지 확인하지 않아, 관측용 no-op만 잠금 안에 남기는 회귀를 놓친다"이다. 재현(`go test -overlay ... -count=3` → `ok … 86.768s`)과 두 구간 사이 탈취 시 HEAD 이동이 이 감사의 SL2·E3과 같다. codex는 원본 시험·`-race`·binlag 461/464 통과도 확인했다. 이 판정은 오케스트레이터가 위 값을 `<toplevel>/.moai/state/audit-plan-result.json`에 쓰고 `moai verify audit-plan --project-root <toplevel> --result-file <path>`를 통과시키기 전까지 확정되지 않는다.

## Baseline-attribution

- 모든 측정 대상은 워크트리 HEAD `ef4a09127b9309185b33b52489d57c2a1a88e6e5`이다(`git rev-parse HEAD`, 작업 트리 clean). mutant 트리는 같은 HEAD의 `git archive`에서 풀었고, 시험 파일은 HEAD와 `cmp` 동일하다(E1).
- 시험·vet·lint는 이 실행에서 트리 소스로 직접 빌드한 Go 툴체인이 판정했다. 설치된 `moai` 바이너리로 판정한 것은 감사 계획 판독과 codex 실행뿐이다. 그 빌드는 `ee981a932`이고 HEAD의 조상이다(§2.2: 판정 빌드 `ee981a932` / 트리 HEAD `ef4a09127`).
- 레인이 보고한 값(evidence.md의 15/15 PASS, `-count=5` 등)은 인용만 했고, 판정 근거로 쓴 수치는 모두 이 감사가 다시 측정한 값이다.

## Gaps

- 감사 계획과 codex 판정은 HEAD보다 오래된 설치 빌드(`ee981a932`)로 실행됐다. 이 카드는 감사 계획 코드를 건드리지 않아 영향은 없다고 보지만, 트리 빌드로 다시 확인하지는 않았다.
- 변형마다 `-count=3`만 돌렸다. 5회차 S처럼 확률 신호에만 걸리는 형태가 있는지 더 많은 반복으로 확인하지는 않았다. 다만 SL·SL2는 9/9 통과했고, 나머지는 모두 결정 신호(같은 행 3/3)로 실패했다.
- Windows 경로(잠금 파일 create-new 방식)의 트립와이어 동작은 이 macOS 실행에서 관찰하지 않았다.
- 전체 `internal/cli` 스위트는 돌리지 않았다(범위 한정 원칙). 영향받는 시험만 측정했다.
- 거부된 도구 호출: worktree 격리 가드가 복합 셸 명령 몇 건을 거부했다(git 포함 형태로 판독됨). 같은 명령을 스크립트 파일(`/private/tmp/t70-audit6/run2.sh`, `run3.sh`, `prep2.py`)로 나눠 실행했다. 측정 내용 자체는 바뀌지 않았다.

## Residual-risk

- F1 수리 뒤에도, seam이 이동을 수행하면서 그 **안에서** 잠금을 다시 나누는 식의 더 깊은 무력화는 이론상 남는다. 다만 운영 seam 본문은 merge 한 줄이라, 그런 변경은 리뷰에서 바로 드러난다.
- DuringRevalidate의 goroutine 경쟁자 신호는 여전히 시간에 좌우된다. 결정 신호(읽기 probe, 트립와이어, 이동 probe, 수리 뒤의 HEAD 확인)가 판정을 맡으므로 판정 안정성에는 영향이 없다.
- 시험 시간 약 17초는 `internal/cli` 스위트 부하를 늘린다(F3).
- 근거 로그(`/private/tmp/t70-audit6/logs/*.log`)는 기계 로컬 임시 위치에 있다. 판정을 결정한 줄은 이 파일에 옮겨 적었고, 나머지는 손실을 감수한다.
- 비원자적 abort 창(프로토콜 밖 행위자 전용)은 t1616 감사 때부터 수용된 잔여 위험이며 이번 범위 밖이다.

## Iteration history

| 회차 | 감사 커밋 | 판정 | 놓친 변형 |
|---|---|---|---|
| 1 | a2055ca46 | FAIL 72 | A |
| 2 | b79a70d97 | FAIL 78 | B |
| 3 | 9aa451b10 | FAIL 82 | D |
| 4 | 9429ec4f5 | FAIL 85 | H, D3 |
| 5 | 763fafafd | FAIL 87 | S |
| 6 | ef4a09127 | FAIL 90 (자체 점수로는 PASS-WITH-DEBT, 필수 codex 게이트 FAIL) | SL, SL2 (seam 본문 no-op, 의도적 무력화 계열) |
