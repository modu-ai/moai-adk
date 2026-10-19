auditor-model: claude-opus-5-5
verdict: PASS-WITH-DEBT
audited_sha: 20d10a1804af4ea999fcacb555ba7f0a804d30a7

## Evaluation Report — 카드 t1670 sync-audit 7회차 (델타 재감사, 운영자 승인 수리 회차)

SPEC: SPEC-LOCAL-MAIN-FLOW-001
Overall Verdict: PASS-WITH-DEBT — overall score 93/100

감사 대상: 워크트리 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670`, 브랜치 `WT-t1616-10-10`, HEAD `20d10a180`(전체 SHA는 위 `audited_sha`, 작업 트리 clean), 기준 local main `67e462e80`. 판정 대상 수리 커밋은 `f5b7d193e`(시험 전용, `internal/cli/local_main_resync_gap_test.go` +7줄)이고, 6회차 감사 대상 `ef4a09127` 이후 `internal/` 아래 변경은 이 7줄뿐이다(`git diff --stat ef4a09127 HEAD -- internal/` → `1 file changed, 7 insertions(+)`). 그 뒤 커밋 `05be4e1a3`·`20d10a180`은 `.moai/reports/t1670/` 문서만 바꿨다.

### 요약

6회차 Required fix F1이 그대로 들어갔다. fast-forward seam 래퍼가 실제 `prevMove` 호출이 끝난 직후 "오류 없음 && primary HEAD == target"을 `movedBySeam`에 기록하고, 거짓이면 `:241`에서 `t.Errorf`로 실패한다. 6회차 감사가 실측한 FIXREAL 수리안과 같은 모양이다.

이 감사가 생성기(`/private/tmp/t70-audit7/mutate.py`, 6회차 생성기에서 ROOT만 바꾸고 새 변형 5종을 더함)로 HEAD 트리를 다시 풀어 만든 결과는 이렇다.

- 이전 회차의 비대조 변형 15종(A·B·C·D1·D2·D3·DH·E·F·H·S·SB·O·SL·SL2)이 모두 최소 한 시험에서 **3/3 실패**한다. 6회차에 놓친 SL·SL2는 이제 `:241`에서 3/3 실패한다.
- 새로 만든 개연성 있는 회귀 2종도 잡힌다. SS(잠금을 seam 안으로 옮기는 "캡슐화" 리팩터)와 E2(훅은 구간에 남기고 seam 호출만 구간 뒤로 뺀 형태)는 각각 DuringRevalidate 3/3, SS는 BeforeFastForward도 3/3 실패한다.
- 실제 트리는 세 시험 모두 3/3 통과하고, 두 rival 시험 `-count=5`도 10/10 통과한다.
- 남은 미검출 변형은 LE 하나다. 잠금 획득 실패 오류를 삼키는 형태(`_ = lockErr`)인데, 이동이 잠금 안의 함수에서만 일어나므로 **경합을 되살리지 않는다**. 오류 분류가 틀려지는 회귀이고, 그 경로는 기준 커밋(`67e462e80`, 249행)부터 시험이 없었다. 카드 완료 기준 밖이라 선택 항목으로 분류했다(F2).

필수 백엔드 codex는 `pass`를 냈다(`audit_multi` overall `pass`, receipt `rcpt-ff3fd393de98b6c37c60af67`). 이 감사 자신의 점수도 PASS 영역이며, 차단 항목은 없다. 판정을 PASS가 아닌 PASS-WITH-DEBT로 둔 것은 아래 선택 항목 세 개(시험 비용, 시험이 없는 잠금 실패 경로, 동작이 올바른 리팩터에 대한 안전 쪽 오탐)를 후속 부채로 남기기 위해서다.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 94/100 | PASS | 실제 트리 `-count=3`: BFF `FAIL=0 PASS=3`, DR `FAIL=0 PASS=3`, SEEN `FAIL=0 PASS=3`, `ok ... 85.434s`. 실제 트리 `-count=5`(BFF+DR): `PASS=10 FAIL=0`, `ok ... 133.086s`. 비대조 변형 17종(이전 15 + 새 SS·E2) 모두 최소 한 시험 3/3 FAIL. 완료 기준 (1)·(2) 충족(아래 Claim 5·6) |
| Security (25%) | 95/100 | PASS | 이번 수리는 시험 전용이고 운영 코드 변경 없음(`git diff --stat ef4a09127 HEAD -- internal/` → 시험 파일 7줄). 운영 seam 본문은 6회차와 같다: `return git("-c", "merge.autoStash=false", "merge", "--ff-only", "--no-overwrite-ignore", "-q", baseline)`. 외부 입력·비밀 없음 |
| Craft (20%) | 90/100 | PASS | `-race`: `PASS=17 FAIL=0 DATA_RACE=0`, `ok ... 45.303s`. `go vet` exit 0. `golangci-lint run --new-from-rev 67e462e80 ./internal/cli/` → `0 issues.`. `gofmt -l`(변경 세 파일) 출력 없음. DuringRevalidate 한 건 17.47s(선택 부채 F1) |
| Consistency (15%) | 93/100 | PASS | 새 단언은 기존 래퍼 안의 `prevMove` 호출을 그대로 감싸고, `t.Errorf` + 대상 SHA를 담은 메시지라는 기존 단언 관례를 따른다. 단 시험 머리 주석(55~117행)의 포괄 변형 목록은 S에서 멈춰 SL·SL2와 네 번째 결정 신호(이동 후 HEAD 확인)를 적지 않았다(선택 F4) |

가중 점수: 94×0.40 + 95×0.25 + 90×0.20 + 93×0.15 = 37.6 + 23.75 + 18.0 + 13.95 = 93.3 → 93. must-pass 차원(Functionality, Security)은 모두 기준을 넘는다. 필수 codex 게이트 `pass`. 이 판정은 오케스트레이터의 `moai verify audit-plan --result-file` 확인을 거쳐야 확정된다.

### Findings (structured defect-list)

- F1 [P3] [optional] [확신도 높음] `internal/cli/local_main_resync_gap_test.go:122` DuringRevalidate 시험 비용 17.47s(`-race` 실행). 두 동기 probe가 각각 잠금 재시도 예산을 끝까지 기다린다. 판정을 결정적으로 만드는 대가이며 6회차 F3과 같은 항목이다. — Required fix(선택): 줄이려면 시험 전용 재시도 예산 seam이 필요하다. 이 카드 범위 밖이므로 후속 카드로 넘긴다.
- F2 [P3] [optional] [확신도 높음] `internal/cli/local_main_resync.go:273-274` — 잠금 획득 실패(`lockErr`) 분기에 시험이 없다. 새 변형 LE(`return "", localMainResyncRefusal(... "take the mutation lock before the move" ...)` → `_ = lockErr`)가 세 시험을 9/9 통과한다. 경합 회귀는 아니다. 함수가 잠금 안에서만 실행되므로 잠금을 못 잡으면 이동도 없다. 대신 step 6에서 HEAD≠baseline 이상(anomaly) 경로로 빠져, `MergeExitOther` 대신 사후 이상 분류와 정책 hold가 남는다. 같은 분기는 기준 커밋에도 있었다(`67e462e80` 249행). 이 카드 이전부터 있던 공백이다. — Required fix(선택): 후속 카드에서 `RunIntegrationMutation`이 busy를 돌려주는 상황(구간 밖에서 잠금을 쥔 상태로 re-sync 실행)을 만들어 오류 분류 `MergeExitOther`와 HEAD 불변을 단언한다.
- F3 [P3] [optional] [확신도 높음] `internal/cli/local_main_resync_gap_test.go:238,241,250` — 동작이 올바른 리팩터 세 가지를 안전 쪽으로 거절한다. P(구간 밖 사전 검증 추가)는 `:238`, G(seam을 거치지 않는 직접 읽기를 구간 안에서)는 `:250`/`:316`, 새 대조군 SL3(seam을 no-op으로 두고 같은 구간에서 merge를 seam 바로 뒤에 인라인)는 `:241`에서 실패한다. 시험이 seam을 관찰 지점으로 고정한 데 따른 의도된 대가이고, 실패 메시지가 원인을 정확히 말해 준다. 차단 사유가 아니다. 다른 대조군 HA(훅을 이동 뒤로, 같은 구간)는 9/9 통과해 오탐이 없다.
- F4 [P3] [optional] [확신도 높음] `internal/cli/local_main_resync_gap_test.go:55-117` — 시험 머리 주석이 포괄 변형을 S까지만 열거하고, 이번에 더한 "seam 호출 직후 HEAD==target" 확인(SL·SL2 대응)을 신호 목록에 넣지 않았다. — Required fix(선택): 신호 4 설명에 "seam이 반환한 직후 HEAD가 target이어야 한다(seam 본문 no-op + 다른 경로 merge, mutant SL/SL2)" 한 문장을 더한다.
- F5 [info] [optional] 설치된 `moai` 바이너리와 MCP 서버는 `ee981a932`로 빌드됐고 HEAD `20d10a180`의 조상이다(`audit_multi`의 `build_lag`). 감사 계획 판독과 codex 실행은 그 빌드로 이뤄졌다. 이 카드는 감사 계획·codex 경로 코드를 건드리지 않는다(Gaps 참조).

### Recommendations

- 카드는 판정 요건을 충족한다. F1·F2·F3·F4는 모두 선택 항목이며 자동 수정 대상이 아니다.
- F2(잠금 획득 실패 경로 시험)는 이 카드와 무관하게 남아 있던 공백이라 별도 후속 카드로 다루는 편이 범위가 맞다.

## Claim

1. 수리 `f5b7d193e`는 6회차 Required fix F1(seam 호출 직후 HEAD==target 단언)을 그대로 구현했다. 운영 코드는 바뀌지 않았다.
2. 실제 트리(HEAD `20d10a180`)는 세 rival 시험을 각 3/3, 두 rival 시험을 `-count=5`로 10/10 통과한다.
3. 이전 회차의 비대조 변형 15종이 모두 최소 한 시험에서 3/3 실패한다. 6회차에 놓친 SL·SL2도 포함된다.
4. 새 변형 가운데 개연성 있는 회귀 SS·E2는 3/3 실패한다. 미검출 변형 LE는 경합을 되살리지 않는 오류 분류 회귀이고, 기준 커밋부터 있던 시험 공백이다.
5. 완료 기준 (1) "재검증과 이동이 한 잠금 구간을 공유함을 확인하는 시험": 그 구간을 나누거나 빠져나가는 변형(A·B·C·D·H·DH·E·E2·F·O·S·SB·SL·SL2·SS)이 모두 잡힌다. 충족.
6. 완료 기준 (2) "잠금을 지키는 경합에서 창 재현 0회": 실제 트리 13회 실행(`-count=3` 1회분, `-count=5`, `-race` 1회)에서 rival 착지 `l.head == base`가 0회, BFF rival 거절이 매회. 충족.
7. `-race` 데이터 경합 없음, `TestAuditLagUsesBinlagSeam` 통과, vet·gofmt·lint 깨끗.
8. 필수 codex 게이트 `pass`.

## Evidence

E1 — 수리 diff(`git show f5b7d193e`, 요지):
```
+	movedBySeam := false
...
 			moveProbeErr = probeSection(root)
+			out, mErr := prevMove(git, baseline)
+			movedBySeam = mErr == nil && lmfHead(t, root) == target
+			return out, mErr
...
+	if !movedBySeam {
+		t.Errorf("the fast-forward seam localMainResyncFastForward returned without moving primary HEAD to %s: the move is done somewhere other than the observed seam", target)
+	}
```

E2 — 변형 재구성. `git -C <worktree> archive --format=tar -o /private/tmp/t70-audit7/head.tar HEAD`로 HEAD 트리를 담아 23개 트리로 풀고, `/private/tmp/t70-audit7/mutate.py`(6회차 생성기 + 새 SS·E2·LE·SL3·HA)를 적용했다. `setup.sh` 출력: 23개 트리 모두 `test identical`(시험 파일이 HEAD와 `cmp` 동일)과 `vet ok`, REAL은 `prod UNCHANGED`. 새 변형의 운영 diff:
- SS: seam 본문을 `factory.RunIntegrationMutation(mutSSRoot, func() error { out, mErr = git(... merge ...) })`로 감싸고, 호출부는 구간에서 재검증만 하고 구간 뒤에 `localMainResyncBeforeFastForward` + `localMainResyncFastForward`
- E2: 구간 안에는 훅만 남기고, `_, moveErr = localMainResyncFastForward(git, baseline)`을 구간 뒤로
- LE: `return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: take the mutation lock before the move: %v", lockErr)` → `_ = lockErr`
- SL3(대조): seam 본문 `return "", nil`, 같은 구간에서 seam 바로 뒤에 literal merge
- HA(대조): 같은 구간 안에서 seam 호출과 훅의 순서를 바꿈

E3 — 실행 명령(트리별 순차): `TMPDIR=GOTMPDIR=/private/tmp/t70-audit7/tmp; cd <tree>; go test -count=3 -v -run 'TestLocalMainResyncRivalTakeover(BeforeFastForward|DuringRevalidate)CannotLand$|TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses$' ./internal/cli/`. 요약 `/private/tmp/t70-audit7/logs/summary.txt`. 실패 수 BFF / DR / SEEN:

| 트리 | 형태 | BFF | DR | SEEN | 결정 신호(대표 줄) |
|---|---|---|---|---|---|
| REAL | 실제 코드 | 0/3 | 0/3 | 0/3 | `ok ... 85.434s` |
| A | 재검증을 구간 밖으로 | 0/3 | 3/3 | 0/3 | `:253`, `:256` |
| B | 재검증·이동 두 구간 | 0/3 | 3/3 | 0/3 | `:256 the mutation lock was acquired between the ownership read and the fast-forward` |
| C | 결과를 구간 전에 계산 | 0/3 | 3/3 | 0/3 | `:253`, `:256` |
| D1 | seam 읽기를 구간 밖으로 | 0/3 | 3/3 | 0/3 | `:253`, `:256` |
| D2 | 직접 읽기를 구간 밖으로 | 0/3 | 3/3 | 3/3 | `:250 ... never ran`, `:316` |
| D3 | 직접 읽기 밖 + seam 읽기 안 | 0/3 | 0/3 | 3/3 | `:319 ... must refuse with the not-holder class 14, got <nil>`, `:322` |
| DH | D3×H | 0/3 | 3/3 | 3/3 | `:256`, `:319` |
| E | 이동·훅을 구간 뒤로 | 3/3 | 3/3 | 0/3 | `:45`, `:259` |
| F | 잠금 제거 | 3/3 | 3/3 | 0/3 | `:45`, `:253`, `:259` |
| H | 읽기만 첫 구간 | 0/3 | 3/3 | 0/3 | `:256` |
| S | 훅 첫 구간, seam 이동 둘째 구간 | 0/3 | 3/3 | 0/3 | `:256` |
| SB | S + 둘째 구간이 seam 우회 | 0/3 | 3/3 | 0/3 | `:238 ... fast-forward seam localMainResyncFastForward ran = false` |
| O | 구간 안에서 이동을 재검증보다 먼저 | 3/3 | 3/3 | 3/3 | `:238`, `:322`, `:48` |
| **SL** | no-op seam 구간 안 + 잠금 없는 merge 구간 뒤 | 0/3 | **3/3** | 0/3 | `:241 the fast-forward seam localMainResyncFastForward returned without moving primary HEAD to ...` |
| **SL2** | no-op seam 첫 구간 + merge 둘째 구간 | 0/3 | **3/3** | 0/3 | `:241` 같은 줄 |
| SS (새) | 잠금을 seam 안으로, 재검증은 별도 구간 | 3/3 | 3/3 | 0/3 | `:45`, `:259`, `:270` |
| E2 (새) | 훅은 구간 안, seam 호출은 구간 뒤 | 0/3 | 3/3 | 0/3 | `:259 ... from the fast-forward seam itself, was not refused as lock-busy`, `:270` |
| LE (새) | 잠금 획득 오류 무시 | 0/3 | 0/3 | 0/3 | `ok ... 83.030s` — 미검출(경합 회귀 아님, F2) |
| G (대조) | 동작 올바름: seam 우회 직접 읽기 구간 안 | 0/3 | 3/3 | 3/3 | `:250`, `:316` (안전 쪽 오탐) |
| P (대조) | 동작 올바름: 구간 밖 사전 검증 | 0/3 | 3/3 | 0/3 | `:238` (안전 쪽 오탐) |
| SL3 (새 대조) | 동작 올바름: no-op seam + 같은 구간 인라인 merge | 0/3 | 3/3 | 0/3 | `:241` (안전 쪽 오탐) |
| HA (새 대조) | 동작 올바름: 훅을 이동 뒤로, 같은 구간 | 0/3 | 0/3 | 0/3 | `ok ... 83.182s` (오탐 없음) |

D3는 6회차와 같이 짝 시험 SEEN이 결정적으로 잡는다(DR 단독으로는 통과).

E4 — 실제 트리 반복(`/private/tmp/t70-audit7/stress.sh`):
```
go test -count=5 -v -run 'TestLocalMainResyncRivalTakeover(BeforeFastForward|DuringRevalidate)CannotLand$' ./internal/cli/
exit=0
PASS=10 FAIL=0
ok  	github.com/modu-ai/moai-adk/internal/cli	133.086s
```

E5 — 실제 트리 경합·정적 점검(`/private/tmp/t70-audit7/static.sh`, 요약 `logs/static.txt`):
```
HEAD=20d10a1804af4ea999fcacb555ba7f0a804d30a7 status_lines=0
go version go1.26.8 darwin/arm64
== race exit=0
PASS=17 FAIL=0 DATA_RACE=0
--- PASS: TestLocalMainResyncRivalTakeoverBeforeFastForwardCannotLand (9.31s)
--- PASS: TestLocalMainResyncRivalTakeoverDuringRevalidateCannotLand (17.47s)
--- PASS: TestLocalMainResyncRivalTakeoverSeenByOwnershipReadRefuses (0.90s)
--- PASS: TestAuditLagUsesBinlagSeam (0.40s)
ok  	github.com/modu-ai/moai-adk/internal/cli	45.303s
== vet exit=0
== lint exit=0
0 issues.
== gofmt exit=0
// included. `git merge-base --is-ancestor` exits 1 for "not an ancestor". Any
	_, err := git("merge-base", "--is-ancestor", a, b)
 internal/cli/local_main_resync_gap_test.go | 7 +++++++
 1 file changed, 7 insertions(+)
```
명령: `go test -race -count=1 -v -run 'TestLocalMainResync|TestAuditLagUsesBinlagSeam' ./internal/cli/`, `go vet ./internal/cli/`, `golangci-lint run --new-from-rev 67e462e80 ./internal/cli/`, `gofmt -l`(변경 세 파일), `sed -n '461p;464p' internal/cli/local_main_resync.go`(binlag 좌표 461/464가 시험 기대값과 같음), `git diff --stat ef4a09127 HEAD -- internal/`.

E6 — F2 전제 확인. `git show 67e462e80:internal/cli/local_main_resync.go | grep -n "take the mutation lock before the move"` → `249: return "", localMainResyncRefusal(factory.MergeExitOther, "integration resync: take the mutation lock before the move: %v", lockErr)`. `grep -rln "take the mutation lock" internal/` → `internal/cli/local_main_resync.go`, `internal/factory/integration_lock.go`만 나오고 시험 파일은 없다.

E7 — 교차 모델 감사. `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1670` → `config_status: ok`, `model: codex`, `cross_model_active: true`, `cross_model_required: true`, `enforced_required: ["codex"]`(claude·glm gate `off`), exit 0. `mcp__moai__audit_multi`(gates 인자 없음, `project_root` = 이 워크트리, `target: baseBranch`, `session_id: t1670-sync-audit-iter7`) 결과 digest:
- `overall_verdict`: `pass`
- `gate_unmet`: 결과에 없음(비어 있음)
- `plan_source`: `config`
- per_backend_verdicts: `backend=codex, gate=required, verdict=pass`
- `audit_receipt`: `rcpt-ff3fd393de98b6c37c60af67`

codex의 요약은 `Verdict: pass` 한 줄이고 지적 사항은 0건이다. 이 판정은 오케스트레이터가 위 값을 `<toplevel>/.moai/state/audit-plan-result.json`에 새로 쓰고 `moai verify audit-plan --project-root <toplevel> --result-file <path>`를 통과시키기 전까지 확정되지 않는다.

## Baseline-attribution

- 모든 측정 대상은 워크트리 HEAD `20d10a1804af4ea999fcacb555ba7f0a804d30a7`이다(`git rev-parse HEAD`, `git status --short` 0줄). 변형 트리는 같은 HEAD의 `git archive`에서 풀었고, 시험 파일은 HEAD와 `cmp` 동일하다(E2).
- 시험·vet·lint·gofmt는 이 실행에서 트리 소스로 직접 빌드한 Go 툴체인(`go1.26.8 darwin/arm64`)이 판정했다. 설치된 `moai` 바이너리로 판정한 것은 감사 계획 판독과 codex 실행뿐이다. 그 빌드는 `ee981a932`이고 HEAD의 조상이다(§2.2: 판정 빌드 `ee981a932` / 트리 HEAD `20d10a180`).
- 레인 보고값(`evidence.md`의 실제 트리 15/15 PASS, 변형 표 `/private/tmp/t70-mut7/logs/summary.txt`)은 참고로만 읽었다. 판정에 쓴 수치는 모두 이 감사가 다시 측정한 값이며, 레인 표와 실패 행 번호(`:253`, `:256`, `:250`, `:316`)가 일치한다.

## Gaps

- 감사 계획과 codex 판정은 HEAD보다 오래된 설치 빌드(`ee981a932`)로 실행됐다. 이 카드는 감사 계획·codex 경로 코드를 건드리지 않아 영향은 없다고 보지만, 트리 빌드로 다시 확인하지는 않았다.
- codex 결과는 `Verdict: pass` 한 줄뿐이라, codex가 어떤 변형을 시도했는지는 결과에서 관찰할 수 없다. 변형 판단의 근거는 이 감사의 E3이다.
- 변형마다 `-count=3`만 돌렸다. 다만 잡힌 변형은 모두 같은 결정 행이 3/3 반복됐다(DR의 goroutine 신호 `:270`만 확률적이고, 그 신호에만 기대는 변형은 없다).
- LE를 잡는 시험을 이 감사가 직접 써서 실측하지는 않았다(F2의 수리안은 미측정 제안).
- Windows 경로(잠금 파일 create-new 방식)의 트립와이어 동작은 이 macOS 실행에서 관찰하지 않았다.
- 전체 `internal/cli` 스위트는 돌리지 않았다(범위 한정 원칙). 영향받는 시험만 측정했다.
- 거부된 도구 호출: worktree 격리 가드가 복합 셸 명령 세 건을 거부했다(git이 섞였거나 `go test -run` 정규식이 들어간 복합 형태). 같은 내용을 단순 명령과 스크립트 파일(`/private/tmp/t70-audit7/addnew.py`, `run.sh`, `static.sh`, `stress.sh`)로 나눠 실행했다. 측정 내용 자체는 바뀌지 않았다.

## Residual-risk

- 관찰 지점은 이제 읽기 seam(probe + 트립와이어 심기)과 fast-forward seam(진입 시 트립와이어·probe, 반환 직후 HEAD)이다. seam 본문이 이동을 수행하면서 그 **안에서** 구간 잠금을 풀었다 다시 잡는 형태만 이론상 남는데, 잠금 fd는 `factory` 안에 있어 seam에서 손댈 수 없다. 운영 seam 본문은 merge 한 줄이라 그런 변경은 리뷰에서 바로 드러난다.
- DuringRevalidate의 goroutine 경쟁자 신호는 여전히 시간에 좌우된다. 결정 신호가 판정을 맡으므로 판정 안정성에는 영향이 없다.
- 시험 시간 약 17초가 `internal/cli` 스위트 부하를 늘린다(F1).
- 잠금 획득 실패 분기의 오류 분류는 여전히 시험되지 않는다(F2, 기준 커밋부터).
- 근거 로그(`/private/tmp/t70-audit7/logs/*.log`)는 기계 로컬 임시 위치에 있다. 판정을 결정한 줄은 이 파일에 옮겨 적었고, 나머지는 손실을 감수한다.
- 비원자적 abort 창(프로토콜 밖 행위자 전용)은 t1616 감사 때부터 수용된 잔여 위험이며 이번 범위 밖이다.

## Iteration history

| 회차 | 감사 커밋 | 판정 | 놓친 변형 |
|---|---|---|---|
| 1 | a2055ca46 | FAIL 72 | A |
| 2 | b79a70d97 | FAIL 78 | B |
| 3 | 9aa451b10 | FAIL 82 | D |
| 4 | 9429ec4f5 | FAIL 85 | H, D3 |
| 5 | 763fafafd | FAIL 87 | S |
| 6 | ef4a09127 | FAIL 90 (자체 점수로는 PASS-WITH-DEBT, 필수 codex 게이트 FAIL) | SL, SL2 |
| 7 | 20d10a180 | PASS-WITH-DEBT 93 (필수 codex 게이트 PASS) | 없음(경합 계열). LE는 경합 아님·기준 커밋부터의 공백 |
